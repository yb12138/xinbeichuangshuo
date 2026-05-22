// gameflow: 运行时策略：注册各类 policy/hook 的入口。

package engine

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

type actionBoundaryTimingStage int

const (
	actionBoundaryResolveField actionBoundaryTimingStage = iota
	actionBoundaryResolveActionStart
)

// runActionBoundaryTimingStageHooks 统一处理行动阶段开始前/开始时规则。
func (e *GameEngine) runActionBoundaryTimingStageHooks(player *model.Player, stage actionBoundaryTimingStage) bool {
	switch stage {
	case actionBoundaryResolveField:
		return e.RunActionBeforeTimingHooks(player)
	case actionBoundaryResolveActionStart:
		return e.runTimingActionStartExecuteHooks(player)
	default:
		panic(fmt.Sprintf("unregistered TimingActionBefore stage: %d", stage))
	}
}

// RunActionBeforeTimingHooks 在行动阶段开始前固定阶段按顺序处理场上效果。
func (e *GameEngine) RunActionBeforeTimingHooks(player *model.Player) bool {
	for _, hook := range e.beforeActionFieldHooks {
		if hook(e, player) {
			return true
		}
	}
	return false
}

// applyTimingActionStartExecuteOptionPolicies 在行动阶段中生成选项前应用规则约束。
func (e *GameEngine) applyTimingActionStartExecuteOptionPolicies(player *model.Player, state *ActionSelectionState) {
	ctx := engineplayer.TimingHookContext{
		Player:         player,
		ChoiceRuntime:  NewRoleChoiceRuntime(e),
		OptionModifier: actionSelectionModifierAdapter{state: state},
	}
	e.dispatchAllRoleTimingHooks(engineplayer.TimingActionStartOption, ctx)
}

// applyTimingActionStartExecuteValidationPolicies 在行动阶段中输入校验前应用规则约束。
func (e *GameEngine) applyTimingActionStartExecuteValidationPolicies(player *model.Player, state *ActionSelectionState) {
	ctx := engineplayer.TimingHookContext{
		Player:         player,
		ChoiceRuntime:  NewRoleChoiceRuntime(e),
		OptionModifier: actionSelectionModifierAdapter{state: state},
		ValidationModifier: actionSelectionValidationModifierAdapter{
			actionSelectionModifierAdapter: actionSelectionModifierAdapter{state: state},
			result:                         nil,
			engine:                         e,
		},
	}
	e.dispatchAllRoleTimingHooks(engineplayer.TimingActionStartValidation, ctx)
}

// RunAttackResponseCombatInteractionPolicies 在战斗交互阶段执行命中判定策略链。
func (e *GameEngine) RunAttackResponseCombatInteractionPolicies(req *model.CombatRequest) bool {
	result := e.dispatchRoleTimingHook(engineplayer.TimingCombatInteraction, engineplayer.TimingHookContext{
		CombatRequest: req,
	})
	return result.Interrupted
}

// runAttackDeclareInterruptPolicies 在攻击宣言后执行中断策略。
func (e *GameEngine) runAttackDeclareInterruptPolicies(attacker *model.Player, target *model.Player, currentAction *model.QueuedAction, userCtx *model.Context) bool {
	ctx := engineplayer.TimingHookContext{
		Attacker: attacker,
		Target:   target,
		Action:   currentAction,
		UserCtx:  userCtx,
	}
	result := e.dispatchRoleTimingHook(engineplayer.TimingAttackDeclareInterrupt, ctx)
	return result.Interrupted
}

// applyAttackResponseDefendValidation 在防御判定时执行校验策略。
func (e *GameEngine) applyAttackResponseDefendValidation(player *model.Player, req *model.CombatRequest) error {
	result := e.dispatchRoleTimingHook(engineplayer.TimingDefendValidation, engineplayer.TimingHookContext{
		Player:        player,
		CombatRequest: req,
	})
	return result.ValidationError
}

// applyAttackResponseCounterCardPolicy 在应战出牌时执行卡牌校验策略。
func (e *GameEngine) applyAttackResponseCounterCardPolicy(player *model.Player, req *model.CombatRequest, card model.Card) (bool, model.Card, error) {
	ctx := engineplayer.TimingHookContext{
		Player:        player,
		CombatRequest: req,
		CounterCard:   &card,
	}
	result := e.dispatchRoleTimingHook(engineplayer.TimingCombatCounterCard, ctx)
	if result.ValidationError != nil {
		return false, model.Card{}, result.ValidationError
	}
	return result.Handled, result.Card, nil
}

// applyAttackResponseCounterElementPolicy 在应战元素判定时执行校验策略。
func (e *GameEngine) applyAttackResponseCounterElementPolicy(player *model.Player, req *model.CombatRequest, counterCard model.Card) (bool, bool) {
	ctx := engineplayer.TimingHookContext{
		Player:        player,
		CombatRequest: req,
		CounterCard:   &counterCard,
	}
	result := e.dispatchRoleTimingHook(engineplayer.TimingCounterElementCheck, ctx)
	return result.Handled, result.UseFaction
}

// applyAttackResponseCounterResolvePolicy 在应战成立后执行结算策略。
func (e *GameEngine) applyAttackResponseCounterResolvePolicy(player *model.Player, req *model.CombatRequest, counterCard *model.Card, useFaction bool) {
	ctx := engineplayer.TimingHookContext{
		Player:         player,
		CombatRequest:  req,
		CounterCardPtr: counterCard,
		UseFaction:     useFaction,
	}
	e.dispatchAllRoleTimingHooks(engineplayer.TimingCounterResolve, ctx)
}

// applyTimingMagicMissileDefendValidation 在魔弹防御判定时执行校验策略。
func (e *GameEngine) applyTimingMagicMissileDefendValidation(player *model.Player, chain *model.MagicBulletChain) error {
	result := e.dispatchRoleTimingHook(engineplayer.TimingMagicMissileDefend, engineplayer.TimingHookContext{
		Player:           player,
		MagicBulletChain: chain,
		UserCtx:          e.buildMagicMissileTimingContext(player, chain, model.TimingMagicMissileDefend),
	})
	return result.ValidationError
}

// applyTimingMagicMissileCounterValidation 在魔弹传递判定时执行校验策略。
func (e *GameEngine) applyTimingMagicMissileCounterValidation(player *model.Player, chain *model.MagicBulletChain, card model.Card) error {
	result := e.dispatchRoleTimingHook(engineplayer.TimingMagicMissileCounter, engineplayer.TimingHookContext{
		Player:           player,
		MagicBulletChain: chain,
		Card:             &card,
		UserCtx:          e.buildMagicMissileTimingContext(player, chain, model.TimingMagicMissileCounter),
	})
	return result.ValidationError
}

// applyTimingMagicMissileResponseSkillAugment 在魔弹响应窗口构建前追加可用响应技能。
func (e *GameEngine) applyTimingMagicMissileResponseSkillAugment(skillIDs []string, player *model.Player, chain *model.MagicBulletChain) []string {
	if e == nil || player == nil || chain == nil {
		return skillIDs
	}
	timingCtx := engineplayer.TimingHookContext{
		OfferedSkillIDs:  skillIDs,
		Player:           player,
		MagicBulletChain: chain,
		UserCtx:          e.buildMagicMissileTimingContext(player, chain, model.TimingMagicMissileResponseSkill),
	}
	result := e.dispatchAllRoleTimingHooks(engineplayer.TimingMagicMissileResponseSkillAug, timingCtx)
	if len(result.SkillIDs) > 0 {
		return result.SkillIDs
	}
	return skillIDs
}

func (e *GameEngine) buildMagicMissileTimingContext(player *model.Player, chain *model.MagicBulletChain, timing model.Timing) *model.Context {
	if e == nil || e.State == nil || player == nil || chain == nil {
		return nil
	}
	return e.BuildContext(player, player, timing, &model.EventContext{
		Type:     model.EventMagic,
		SourceID: chain.SourcePlayerID,
		TargetID: chain.TargetID,
	})
}

// applyAttackResponseSkillAugment 在响应技能列表构建时追加技能。
func (sd *SkillDispatcher) applyAttackResponseSkillAugment(skillIDs []string, ctx *model.Context) []string {
	if sd == nil || sd.engine == nil {
		return skillIDs
	}
	timingCtx := engineplayer.TimingHookContext{
		OfferedSkillIDs: skillIDs,
		UserCtx:         ctx,
	}
	result := sd.engine.dispatchAllRoleTimingHooks(engineplayer.TimingResponseSkillAug, timingCtx)
	if len(result.SkillIDs) > 0 {
		return result.SkillIDs
	}
	return skillIDs
}

// applyAttackResponseSkillNormalize 在响应技能列表展示前规范化顺序/互斥项。
func (sd *SkillDispatcher) applyAttackResponseSkillNormalize(skillIDs []string, ctx *model.Context) []string {
	if sd == nil || sd.engine == nil {
		return skillIDs
	}
	timingCtx := engineplayer.TimingHookContext{
		OfferedSkillIDs: skillIDs,
		UserCtx:         ctx,
	}
	result := sd.engine.dispatchAllRoleTimingHooks(engineplayer.TimingResponseSkillNormalize, timingCtx)
	if len(result.SkillIDs) > 0 {
		return result.SkillIDs
	}
	return skillIDs
}
