package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

type timingOnBeforeActionStage int

const (
	timingOnBeforeActionResolveField timingOnBeforeActionStage = iota
	timingOnBeforeActionResolveActionStart
)

type combatInteractionPolicyHook func(e *GameEngine, req *model.CombatRequest) bool
type combatDefendValidationPolicy func(e *GameEngine, player *model.Player, req *model.CombatRequest) error
type combatCounterCardPolicy func(e *GameEngine, player *model.Player, req *model.CombatRequest, card model.Card) (bool, model.Card, error)
type combatCounterElementPolicy func(e *GameEngine, player *model.Player, req *model.CombatRequest, counterCard model.Card) (bool, bool)
type combatCounterResolvePolicy func(e *GameEngine, player *model.Player, req *model.CombatRequest, counterCard *model.Card, useFaction bool)
type magicMissileDefendValidationPolicy func(e *GameEngine, player *model.Player, chain *model.MagicBulletChain) error
type magicMissileCounterValidationPolicy func(e *GameEngine, player *model.Player, chain *model.MagicBulletChain, card model.Card) error
type responseSkillIDAugmenter func(sd *SkillDispatcher, skillIDs []string, ctx *model.Context) []string
type responseSkillIDNormalizer func(sd *SkillDispatcher, skillIDs []string, ctx *model.Context) []string
type attackDeclaredInterruptHook func(e *GameEngine, attacker *model.Player, target *model.Player, currentAction *model.QueuedAction, userCtx *model.Context) bool
type actionEndInterruptHook func(e *GameEngine, ctx *model.Context) bool
type actionSelectionOptionPolicy func(e *GameEngine, player *model.Player, state *actionSelectionState)
type actionSelectionValidationPolicy func(e *GameEngine, player *model.Player, state *actionSelectionState)

// runTimingOnBeforeActionStageHooks 统一处理 TimingOnBeforeAction 阶段规则。
func (e *GameEngine) runTimingOnBeforeActionStageHooks(player *model.Player, stage timingOnBeforeActionStage) bool {
	switch stage {
	case timingOnBeforeActionResolveField:
		return e.runTimingOnBeforeActionHooks(player)
	case timingOnBeforeActionResolveActionStart:
		return e.runTimingBeforeActionExecuteHooks(player)
	default:
		panic(fmt.Sprintf("unregistered TimingOnBeforeAction stage: %d", stage))
	}
}

// runTimingOnBeforeActionHooks 在回合 before-action 固定阶段按顺序处理场上效果。
func (e *GameEngine) runTimingOnBeforeActionHooks(player *model.Player) bool {
	for _, hook := range e.beforeActionFieldHooks {
		if hook(e, player) {
			return true
		}
	}
	return false
}

// applyTimingBeforeActionExecuteOptionPolicies 在行动入口生成选项前应用规则约束。
func (e *GameEngine) applyTimingBeforeActionExecuteOptionPolicies(player *model.Player, state *actionSelectionState) {
	for _, policy := range e.beforeActionOptionPolicies {
		policy(e, player, state)
	}
}

// applyTimingBeforeActionExecuteValidationPolicies 在行动输入校验前应用规则约束。
func (e *GameEngine) applyTimingBeforeActionExecuteValidationPolicies(player *model.Player, state *actionSelectionState) {
	for _, policy := range e.beforeActionValidationPolicies {
		policy(e, player, state)
	}
}

// runTimingOnHitCheckCombatInteractionPolicies 在战斗交互阶段执行命中判定策略链。
func (e *GameEngine) runTimingOnHitCheckCombatInteractionPolicies(req *model.CombatRequest) bool {
	for _, hook := range e.hitCheckCombatInteractionHooks {
		if hook(e, req) {
			return true
		}
	}
	return false
}

// runTimingOnAttackDeclaredInterruptPolicies 在攻击宣言后执行中断策略。
func (e *GameEngine) runTimingOnAttackDeclaredInterruptPolicies(attacker *model.Player, target *model.Player, currentAction *model.QueuedAction, userCtx *model.Context) bool {
	for _, hook := range e.attackDeclaredInterrupts {
		if hook(e, attacker, target, currentAction, userCtx) {
			return true
		}
	}
	return false
}

// runTimingOnActionEndInterruptPolicies 在行动结束时执行中断策略。
func (e *GameEngine) runTimingOnActionEndInterruptPolicies(ctx *model.Context) bool {
	for _, hook := range e.actionEndInterrupts {
		if hook(e, ctx) {
			return true
		}
	}
	return false
}

// applyTimingOnHitCheckCombatDefendValidation 在防御判定时执行校验策略。
func (e *GameEngine) applyTimingOnHitCheckCombatDefendValidation(player *model.Player, req *model.CombatRequest) error {
	steps := make([]func() error, 0, len(e.hitCheckCombatDefendValidationPolicies))
	for _, rule := range e.hitCheckCombatDefendValidationPolicies {
		r := rule
		steps = append(steps, func() error { return r(e, player, req) })
	}
	return runTimingErrorChain(steps...)
}

// applyTimingOnHitCheckCombatCounterCardPolicy 在应战出牌时执行卡牌校验策略。
func (e *GameEngine) applyTimingOnHitCheckCombatCounterCardPolicy(player *model.Player, req *model.CombatRequest, card model.Card) (bool, model.Card, error) {
	for _, policy := range e.hitCheckCombatCounterCardPolicies {
		handled, transformed, err := policy(e, player, req, card)
		if err != nil {
			return false, model.Card{}, err
		}
		if handled {
			return true, transformed, nil
		}
	}
	return false, card, nil
}

// applyTimingOnHitCheckCombatCounterElementPolicy 在应战元素判定时执行校验策略。
func (e *GameEngine) applyTimingOnHitCheckCombatCounterElementPolicy(player *model.Player, req *model.CombatRequest, counterCard model.Card) (bool, bool) {
	for _, policy := range e.hitCheckCombatCounterElementPolicies {
		allowed, useFaction := policy(e, player, req, counterCard)
		if allowed {
			return true, useFaction
		}
	}
	return false, false
}

// applyTimingOnHitCheckCombatCounterResolvePolicy 在应战成立后执行结算策略。
func (e *GameEngine) applyTimingOnHitCheckCombatCounterResolvePolicy(player *model.Player, req *model.CombatRequest, counterCard *model.Card, useFaction bool) {
	for _, policy := range e.hitCheckCombatCounterResolvePolicies {
		policy(e, player, req, counterCard, useFaction)
	}
}

// applyTimingOnHitCheckMagicMissileDefendValidation 在魔弹防御判定时执行校验策略。
func (e *GameEngine) applyTimingOnHitCheckMagicMissileDefendValidation(player *model.Player, chain *model.MagicBulletChain) error {
	steps := make([]func() error, 0, len(e.hitCheckMagicMissileDefendPolicies))
	for _, rule := range e.hitCheckMagicMissileDefendPolicies {
		r := rule
		steps = append(steps, func() error { return r(e, player, chain) })
	}
	return runTimingErrorChain(steps...)
}

// applyTimingOnHitCheckMagicMissileCounterValidation 在魔弹传递判定时执行校验策略。
func (e *GameEngine) applyTimingOnHitCheckMagicMissileCounterValidation(player *model.Player, chain *model.MagicBulletChain, card model.Card) error {
	steps := make([]func() error, 0, len(e.hitCheckMagicMissileCounterPolicies))
	for _, rule := range e.hitCheckMagicMissileCounterPolicies {
		r := rule
		steps = append(steps, func() error { return r(e, player, chain, card) })
	}
	return runTimingErrorChain(steps...)
}

// applyTimingOnHitCheckResponseSkillAugment 在响应技能列表构建时追加技能。
func (sd *SkillDispatcher) applyTimingOnHitCheckResponseSkillAugment(skillIDs []string, ctx *model.Context) []string {
	if sd == nil || sd.engine == nil {
		return skillIDs
	}
	current := skillIDs
	for _, augmenter := range sd.engine.hitCheckResponseSkillIDAugmenters {
		current = augmenter(sd, current, ctx)
	}
	return current
}

// applyTimingOnHitCheckResponseSkillNormalize 在响应技能列表展示前规范化顺序/互斥项。
func (sd *SkillDispatcher) applyTimingOnHitCheckResponseSkillNormalize(skillIDs []string, ctx *model.Context) []string {
	if sd == nil || sd.engine == nil {
		return skillIDs
	}
	current := skillIDs
	for _, normalizer := range sd.engine.hitCheckResponseSkillIDNormalizers {
		current = normalizer(sd, current, ctx)
	}
	return current
}
