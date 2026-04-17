// gameflow: runtime/interrupt 编排器安装与 Prompt 构建。

package engine

import (
	intr "starcup-engine/internal/engine/runtime/interrupt"
	"starcup-engine/internal/model"
)

var _ intr.EngineInterface = (*GameEngine)(nil)

func (e *GameEngine) GetState() *model.GameState {
	if e == nil {
		return nil
	}
	return e.State
}

func (e *GameEngine) GetPlayerByID(playerID string) *model.Player {
	if e == nil || e.State == nil {
		return nil
	}
	return e.State.Players[playerID]
}

func (e *GameEngine) installInterruptOrchestrator() {
	ar := intr.NewActionRules()
	pr := intr.NewPromptRules()
	registerInterruptActionRules(ar)
	registerInterruptPromptRules(pr)
	e.interruptOrchestrator = intr.NewOrchestrator(e, ar, pr)
}

func allowedInterruptActionTypes(types ...model.PlayerActionType) map[model.PlayerActionType]bool {
	set := make(map[model.PlayerActionType]bool, len(types))
	for _, t := range types {
		set[t] = true
	}
	return set
}

func registerInterruptActionRules(r *intr.ActionRules) {
	wrap := func(h func(*GameEngine, model.PlayerAction) error) intr.ActionHandler {
		return func(en intr.EngineInterface, act model.PlayerAction) error {
			return h(en.(*GameEngine), act)
		}
	}
	r.Register(model.InterruptResponseSkill, &intr.ActionRule{Handler: wrap((*GameEngine).handleInterruptResponseSkillAction)})
	r.Register(model.InterruptStartupSkill, &intr.ActionRule{Handler: wrap((*GameEngine).handleInterruptStartupSkillAction)})
	r.Register(model.InterruptGiveCards, &intr.ActionRule{Handler: wrap((*GameEngine).handleInterruptGiveCardsAction)})
	r.Register(model.InterruptChoice, &intr.ActionRule{Handler: wrap((*GameEngine).handleInterruptChoiceAction)})
	r.Register(model.InterruptMagicMissile, &intr.ActionRule{
		Allowed:              allowedInterruptActionTypes(model.CmdRespond),
		InvalidActionMessage: "当前为【魔弹】响应阶段，请使用响应指令",
		Handler:              wrap((*GameEngine).handleMagicMissileResponse),
	})
	r.Register(model.InterruptMagicBulletFusion, &intr.ActionRule{
		Allowed:              allowedInterruptActionTypes(model.CmdSelect),
		InvalidActionMessage: "当前为【魔弹融合】确认阶段，请选择是否发动",
		Handler:              wrap((*GameEngine).handleMagicBulletFusionResponse),
	})
	r.Register(model.InterruptMagicBulletDirection, &intr.ActionRule{
		Allowed:              allowedInterruptActionTypes(model.CmdSelect),
		InvalidActionMessage: "当前为【魔弹掌控】方向选择阶段，请提交选择",
		Handler:              wrap((*GameEngine).handleMagicBulletDirectionResponse),
	})
	r.Register(model.InterruptHolySwordDraw, &intr.ActionRule{
		Allowed:              allowedInterruptActionTypes(model.CmdSelect),
		InvalidActionMessage: "当前为【圣剑】后续选择阶段，请提交选择",
		Handler:              wrap((*GameEngine).handleHolySwordDrawResponse),
	})
	r.Register(model.InterruptSaintHeal, &intr.ActionRule{
		Allowed:              allowedInterruptActionTypes(model.CmdSelect),
		InvalidActionMessage: "当前为【圣疗】选择阶段，请提交选择",
		Handler:              wrap((*GameEngine).handleSaintHealResponse),
	})
	r.Register(model.InterruptMagicBlast, &intr.ActionRule{
		Allowed:              allowedInterruptActionTypes(model.CmdSelect, model.CmdCancel),
		InvalidActionMessage: "当前为【魔爆冲击】响应阶段，请选择弃牌或取消",
		Handler:              wrap((*GameEngine).handleMagicBlastResponse),
	})
}

func registerInterruptPromptRules(r *intr.PromptRules) {
	wrap := func(b func(*GameEngine) *model.Prompt) intr.PromptBuilder {
		return func(en intr.EngineInterface) *model.Prompt {
			return b(en.(*GameEngine))
		}
	}
	r.Register(model.InterruptResponseSkill, wrap((*GameEngine).buildResponseSkillPrompt))
	r.Register(model.InterruptStartupSkill, wrap((*GameEngine).buildStartupSkillPrompt))
	r.Register(model.InterruptChoice, wrap((*GameEngine).buildChoicePrompt))
	r.Register(model.InterruptMagicMissile, wrap((*GameEngine).buildMagicMissilePrompt))
	r.Register(model.InterruptGiveCards, wrap((*GameEngine).buildGiveCardsPrompt))
	r.Register(model.InterruptMagicBulletFusion, wrap((*GameEngine).buildMagicBulletFusionPrompt))
	r.Register(model.InterruptMagicBulletDirection, wrap((*GameEngine).buildMagicBulletDirectionPrompt))
	r.Register(model.InterruptHolySwordDraw, wrap((*GameEngine).buildHolySwordDrawPrompt))
	r.Register(model.InterruptSaintHeal, wrap((*GameEngine).buildSaintHealPrompt))
	r.Register(model.InterruptMagicBlast, wrap((*GameEngine).buildMagicBlastPrompt))
}

func (e *GameEngine) buildPendingInterruptPrompt() *model.Prompt {
	if e == nil || e.interruptOrchestrator == nil {
		return nil
	}
	return e.interruptOrchestrator.BuildInterruptPrompt()
}

// NotifyInterruptPrompt 实现 interrupt.EngineInterface（编排器在 Push/Pop 后刷新提示）。
func (e *GameEngine) NotifyInterruptPrompt() {
	if e == nil {
		return
	}
	e.notifyInterruptPrompt()
}

// ApplyInterruptPhase 实现 interrupt.EngineInterface。
func (e *GameEngine) ApplyInterruptPhase(intr *model.Interrupt) {
	e.syncGamePhaseWithInterrupt(intr)
}

// ReconcileSubflowAfterInterruptPop 实现 interrupt.EngineInterface。
func (e *GameEngine) ReconcileSubflowAfterInterruptPop(popped *model.Interrupt) {
	if e == nil || e.State == nil {
		return
	}
	if e.State.Subflow == model.SubflowDiscardSelection && !e.hasDiscardSelectionInterrupt() && isDiscardSelectionInterrupt(popped) {
		e.clearSubflow()
	}
}

// syncGamePhaseWithInterrupt 将子流程/回合阶段与当前中断类型对齐（由 Orchestrator 在入栈/出栈时调用）。
func (e *GameEngine) syncGamePhaseWithInterrupt(interrupt *model.Interrupt) {
	if e == nil || interrupt == nil {
		return
	}
	switch interrupt.Type {
	case model.InterruptResponseSkill:
		e.enterResponseWindow()
	case model.InterruptStartupSkill:
		e.clearSubflow()
		e.clearCombatStage()
		if e.State.TurnStage != model.TurnStageTurnStart && e.State.TurnStage != model.TurnStageActionStart {
			e.setTurnStage(model.TurnStageActionStart)
		}
	case model.InterruptChoice:
		if isDiscardSelectionInterrupt(interrupt) {
			e.enterDiscardSelection()
		} else {
			e.clearSubflow()
			e.clearCombatStage()
			if e.State.TurnStage == "" {
				e.setTurnStage(model.TurnStageActionExecution)
			}
		}
	case model.InterruptMagicMissile:
		e.enterResponseWindow()
	case model.InterruptGiveCards:
		e.enterDiscardSelection()
	case model.InterruptMagicBulletFusion, model.InterruptMagicBulletDirection:
		e.clearSubflow()
		e.clearCombatStage()
		e.setTurnStage(model.TurnStageActionExecution)
	case model.InterruptHolySwordDraw:
		e.clearSubflow()
		e.setCombatStage(model.CombatStageDraw)
	case model.InterruptSaintHeal:
		e.clearSubflow()
		e.setCombatStage(model.CombatStageHeal)
	case model.InterruptMagicBlast:
		e.enterResponseWindow()
	}
}

// PushInterrupt 经 interrupt.Orchestrator 入栈并同步阶段。
func (e *GameEngine) PushInterrupt(interrupt *model.Interrupt) {
	if e == nil || interrupt == nil {
		return
	}
	if e.interruptOrchestrator == nil {
		panic("engine: interrupt orchestrator not installed")
	}
	e.interruptOrchestrator.PushInterrupt(interrupt)
}

// PopInterrupt 经 interrupt.Orchestrator 出栈并收敛子流程。
func (e *GameEngine) PopInterrupt() {
	if e == nil {
		return
	}
	if e.interruptOrchestrator == nil {
		panic("engine: interrupt orchestrator not installed")
	}
	e.interruptOrchestrator.PopInterrupt()
}
