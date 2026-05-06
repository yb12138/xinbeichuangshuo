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
	wrapResult := func(h func(*GameEngine, model.PlayerAction) (intr.ActionResult, error)) intr.ActionResultHandler {
		return func(en intr.EngineInterface, act model.PlayerAction) (intr.ActionResult, error) {
			return h(en.(*GameEngine), act)
		}
	}
	// 通用中断类型（非角色专属）
	r.Register(model.InterruptResponseSkill, &intr.ActionRule{HandleResult: wrapResult((*GameEngine).handleInterruptResponseSkillAction)})
	r.Register(model.InterruptStartupSkill, &intr.ActionRule{HandleResult: wrapResult((*GameEngine).handleInterruptStartupSkillAction)})
	r.Register(model.InterruptGiveCards, &intr.ActionRule{HandleResult: wrapResult((*GameEngine).handleInterruptGiveCardsAction)})
	r.Register(model.InterruptChoice, &intr.ActionRule{HandleResult: wrapResult((*GameEngine).handleInterruptChoiceAction)})

	// 角色专属中断类型：通过 InterruptSpecs 动态注册
	mountRoleInterruptSpecs(r, nil)
}

func registerInterruptPromptRules(r *intr.PromptRules) {
	wrap := func(b func(*GameEngine) *model.Prompt) intr.PromptBuilder {
		return func(en intr.EngineInterface) *model.Prompt {
			return b(en.(*GameEngine))
		}
	}
	// 通用中断类型（非角色专属）
	r.Register(model.InterruptResponseSkill, wrap((*GameEngine).buildResponseSkillPrompt))
	r.Register(model.InterruptStartupSkill, wrap((*GameEngine).buildStartupSkillPrompt))
	r.Register(model.InterruptChoice, wrap((*GameEngine).buildChoicePrompt))
	r.Register(model.InterruptGiveCards, wrap((*GameEngine).buildGiveCardsPrompt))

	// 角色专属中断类型：通过 InterruptSpecs 动态注册
	mountRoleInterruptSpecs(nil, r)
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

// mountRoleInterruptSpecs 遍历所有角色的 InterruptSpecs，注册到 action / prompt rules。
func mountRoleInterruptSpecs(ar *intr.ActionRules, pr *intr.PromptRules) {
	for _, entry := range roleRegistry.Entries() {
		for _, spec := range entry.InterruptSpecs {
			s := spec
			if ar != nil && s.HandleActionResult != nil {
				ar.Register(s.Type, &intr.ActionRule{
					HandleResult: func(en intr.EngineInterface, act model.PlayerAction) (intr.ActionResult, error) {
						rt := newRoleChoiceRuntime(en.(*GameEngine))
						result, err := s.HandleActionResult(rt, act)
						return intr.ActionResult{
							Consumed: result.Consumed,
							AfterPop: func(_ intr.EngineInterface) {
								if result.AfterConsume != nil {
									result.AfterConsume(rt)
								}
							},
						}, err
					},
					Allowed:              allowedInterruptActionTypes(s.AllowedActionTypes...),
					InvalidActionMessage: s.InvalidActionMessage,
				})
			}
			if pr != nil && s.BuildPrompt != nil {
				pr.Register(s.Type, intr.PromptBuilder(func(en intr.EngineInterface) *model.Prompt {
					return s.BuildPrompt(newRoleChoiceRuntime(en.(*GameEngine)))
				}))
			}
		}
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
