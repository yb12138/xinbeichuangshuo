// gameflow: 运行时策略：注册各类 policy/hook 的入口。

package engine

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

type timingOnBeforeActionStage int

const (
	timingOnBeforeActionResolveField timingOnBeforeActionStage = iota
	timingOnBeforeActionResolveActionStart
)

// runTimingOnBeforeActionStageHooks 统一处理 TimingOnBeforeAction 阶段规则。
func (e *GameEngine) runTimingOnBeforeActionStageHooks(player *model.Player, stage timingOnBeforeActionStage) bool {
	switch stage {
	case timingOnBeforeActionResolveField:
		return e.RunTimingOnBeforeActionHooks(player)
	case timingOnBeforeActionResolveActionStart:
		return e.runTimingBeforeActionExecuteHooks(player)
	default:
		panic(fmt.Sprintf("unregistered TimingOnBeforeAction stage: %d", stage))
	}
}

// RunTimingOnBeforeActionHooks 在回合 before-action 固定阶段按顺序处理场上效果。
func (e *GameEngine) RunTimingOnBeforeActionHooks(player *model.Player) bool {
	for _, hook := range e.beforeActionFieldHooks {
		if hook(e, player) {
			return true
		}
	}
	return false
}

// applyTimingBeforeActionExecuteOptionPolicies 在行动入口生成选项前应用规则约束。
func (e *GameEngine) applyTimingBeforeActionExecuteOptionPolicies(player *model.Player, state *ActionSelectionState) {
	ctx := engineplayer.TimingHookContext{
		Player:         player,
		ChoiceRuntime:  NewRoleChoiceRuntime(e),
		OptionModifier: actionSelectionModifierAdapter{state: state},
	}
	e.dispatchAllRoleTimingHooks(engineplayer.TimingBeforeActionOption, ctx)
}

// applyTimingBeforeActionExecuteValidationPolicies 在行动输入校验前应用规则约束。
func (e *GameEngine) applyTimingBeforeActionExecuteValidationPolicies(player *model.Player, state *ActionSelectionState) {
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
	e.dispatchAllRoleTimingHooks(engineplayer.TimingBeforeActionValidation, ctx)
}

// RunTimingOnHitCheckCombatInteractionPolicies 在战斗交互阶段执行命中判定策略链。
func (e *GameEngine) RunTimingOnHitCheckCombatInteractionPolicies(req *model.CombatRequest) bool {
	result := e.dispatchRoleTimingHook(engineplayer.TimingOnCombatInteraction, engineplayer.TimingHookContext{
		CombatRequest: req,
	})
	return result.Interrupted
}

// runTimingOnAttackDeclaredInterruptPolicies 在攻击宣言后执行中断策略。
func (e *GameEngine) runTimingOnAttackDeclaredInterruptPolicies(attacker *model.Player, target *model.Player, currentAction *model.QueuedAction, userCtx *model.Context) bool {
	ctx := engineplayer.TimingHookContext{
		Attacker: attacker,
		Target:   target,
		Action:   currentAction,
		UserCtx:  userCtx,
	}
	result := e.dispatchRoleTimingHook(engineplayer.TimingOnAttackDeclaredInterrupt, ctx)
	return result.Interrupted
}

// applyTimingOnHitCheckCombatDefendValidation 在防御判定时执行校验策略。
func (e *GameEngine) applyTimingOnHitCheckCombatDefendValidation(player *model.Player, req *model.CombatRequest) error {
	result := e.dispatchRoleTimingHook(engineplayer.TimingOnDefendValidation, engineplayer.TimingHookContext{
		Player:        player,
		CombatRequest: req,
	})
	return result.ValidationError
}

// applyTimingOnHitCheckCombatCounterCardPolicy 在应战出牌时执行卡牌校验策略。
func (e *GameEngine) applyTimingOnHitCheckCombatCounterCardPolicy(player *model.Player, req *model.CombatRequest, card model.Card) (bool, model.Card, error) {
	ctx := engineplayer.TimingHookContext{
		Player:        player,
		CombatRequest: req,
		CounterCard:   &card,
	}
	result := e.dispatchRoleTimingHook(engineplayer.TimingOnCombatCounterCard, ctx)
	if result.ValidationError != nil {
		return false, model.Card{}, result.ValidationError
	}
	return result.Handled, result.Card, nil
}

// applyTimingOnHitCheckCombatCounterElementPolicy 在应战元素判定时执行校验策略。
func (e *GameEngine) applyTimingOnHitCheckCombatCounterElementPolicy(player *model.Player, req *model.CombatRequest, counterCard model.Card) (bool, bool) {
	ctx := engineplayer.TimingHookContext{
		Player:        player,
		CombatRequest: req,
		CounterCard:   &counterCard,
	}
	result := e.dispatchRoleTimingHook(engineplayer.TimingOnCounterElementCheck, ctx)
	return result.Handled, result.UseFaction
}

// applyTimingOnHitCheckCombatCounterResolvePolicy 在应战成立后执行结算策略。
func (e *GameEngine) applyTimingOnHitCheckCombatCounterResolvePolicy(player *model.Player, req *model.CombatRequest, counterCard *model.Card, useFaction bool) {
	ctx := engineplayer.TimingHookContext{
		Player:         player,
		CombatRequest:  req,
		CounterCardPtr: counterCard,
		UseFaction:     useFaction,
	}
	e.dispatchAllRoleTimingHooks(engineplayer.TimingOnCounterResolve, ctx)
}

// applyTimingOnHitCheckMagicMissileDefendValidation 在魔弹防御判定时执行校验策略。
func (e *GameEngine) applyTimingOnHitCheckMagicMissileDefendValidation(player *model.Player, chain *model.MagicBulletChain) error {
	result := e.dispatchRoleTimingHook(engineplayer.TimingOnMagicMissileDefend, engineplayer.TimingHookContext{
		Player:           player,
		MagicBulletChain: chain,
	})
	return result.ValidationError
}

// applyTimingOnHitCheckMagicMissileCounterValidation 在魔弹传递判定时执行校验策略。
func (e *GameEngine) applyTimingOnHitCheckMagicMissileCounterValidation(player *model.Player, chain *model.MagicBulletChain, card model.Card) error {
	result := e.dispatchRoleTimingHook(engineplayer.TimingOnMagicMissileCounter, engineplayer.TimingHookContext{
		Player:           player,
		MagicBulletChain: chain,
		Card:             &card,
	})
	return result.ValidationError
}

// applyTimingOnHitCheckResponseSkillAugment 在响应技能列表构建时追加技能。
func (sd *SkillDispatcher) applyTimingOnHitCheckResponseSkillAugment(skillIDs []string, ctx *model.Context) []string {
	if sd == nil || sd.engine == nil {
		return skillIDs
	}
	timingCtx := engineplayer.TimingHookContext{
		OfferedSkillIDs: skillIDs,
		UserCtx:         ctx,
	}
	result := sd.engine.dispatchAllRoleTimingHooks(engineplayer.TimingOnResponseSkillAug, timingCtx)
	if len(result.SkillIDs) > 0 {
		return result.SkillIDs
	}
	return skillIDs
}

// applyTimingOnHitCheckResponseSkillNormalize 在响应技能列表展示前规范化顺序/互斥项。
func (sd *SkillDispatcher) applyTimingOnHitCheckResponseSkillNormalize(skillIDs []string, ctx *model.Context) []string {
	if sd == nil || sd.engine == nil {
		return skillIDs
	}
	timingCtx := engineplayer.TimingHookContext{
		OfferedSkillIDs: skillIDs,
		UserCtx:         ctx,
	}
	result := sd.engine.dispatchAllRoleTimingHooks(engineplayer.TimingOnResponseSkillNormalize, timingCtx)
	if len(result.SkillIDs) > 0 {
		return result.SkillIDs
	}
	return skillIDs
}
