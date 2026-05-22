// gameflow: 攻击链上按攻击者/防御者身份挂载的运行时钩子。
// 合并自 attack_card_runtime_hooks.go、attack_passive_runtime_hooks.go。

package engine

import (
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// ---------- 攻击目标上下文 / 状态重置 / 预战斗规则 ----------

// recordTimingOnAttackDeclaredTargetContext 在攻击宣言时写入目标上下文。
func (e *GameEngine) recordTimingOnAttackDeclaredTargetContext(player *model.Player, targetID string) {
	if player == nil {
		return
	}
	e.dispatchAllRoleTimingHooks(engineplayer.TimingOnAttackTargetCtx, engineplayer.TimingHookContext{
		SourceID: player.ID,
		TargetID: targetID,
	})
}

// resetTimingOnAttackDeclaredState 在攻击宣言时清理一次性状态。
func (e *GameEngine) resetTimingOnAttackDeclaredState(player *model.Player) {
	if player == nil {
		return
	}
	e.dispatchAllRoleTimingHooks(engineplayer.TimingOnAttackStateReset, engineplayer.TimingHookContext{
		SourceID: player.ID,
	})
}

// applyAttackDeclarePreCombatRules 在进入战斗交互前应用攻击劫持策略。
func (e *GameEngine) applyAttackDeclarePreCombatRules(player *model.Player, target *model.Player, currentAction *model.QueuedAction, eventCtx *model.EventContext) {
	applyCombatPolicyAttackGating(nil, player, nil, currentAction, eventCtx)
	applyDarkElementNoCounterRule(nil, nil, nil, currentAction, eventCtx)
	if player != nil && eventCtx != nil && eventCtx.AttackInfo != nil {
		var card *model.Card
		var tid string
		if currentAction != nil {
			card = currentAction.Card
		}
		if target != nil {
			tid = target.ID
		}
		e.dispatchAllRoleTimingHooks(engineplayer.TimingOnAttackGating, engineplayer.TimingHookContext{
			SourceID:   player.ID,
			TargetID:   tid,
			Card:       card,
			AttackInfo: eventCtx.AttackInfo,
		})
	}
}

func applyCombatPolicyAttackGating(_ *GameEngine, player *model.Player, _ *model.Player, currentAction *model.QueuedAction, eventCtx *model.EventContext) {
	if player == nil || eventCtx == nil || eventCtx.AttackInfo == nil {
		return
	}
	action := model.Action{
		SourceID: player.ID,
		Type:     model.ActionAttack,
	}
	if currentAction != nil {
		action.TargetID = currentAction.TargetID
		action.Card = currentAction.Card
	}
	consumeAttackCombatPolicyInterceptTags(player, action, eventCtx.AttackInfo)
}

func applyDarkElementNoCounterRule(_ *GameEngine, _ *model.Player, _ *model.Player, currentAction *model.QueuedAction, eventCtx *model.EventContext) {
	if currentAction == nil || currentAction.Card == nil || currentAction.Card.Element != model.ElementDark || eventCtx == nil || eventCtx.AttackInfo == nil {
		return
	}
	eventCtx.AttackInfo.SetInterceptTag(model.CombatInterceptUnrespondable)
}

// ---------- 攻击卡牌变换 ----------

// applyAttackModifyCardTransforms 在攻击宣言时按固定顺序应用卡面变换规则。
func (e *GameEngine) applyAttackModifyCardTransforms(player *model.Player, card model.Card) model.Card {
	ctx := engineplayer.TimingHookContext{
		Player:      player,
		CounterCard: &card,
	}
	result := e.dispatchRoleTimingHook(engineplayer.TimingOnAttackCardTransform, ctx)
	if result.Handled && result.Card.Name != "" {
		return result.Card
	}
	return card
}

// ---------- 攻击被动增伤（原 attack_passive_runtime_hooks.go） ----------

// applyDamageSourceDealAttackModifiers 在伤害计算时按固定顺序应用攻击方被动修正。
func (e *GameEngine) applyDamageSourceDealAttackModifiers(attacker *model.Player, target *model.Player, action model.Action, baseDamage int) int {
	damage := baseDamage
	if attacker != nil {
		var targetID string
		if target != nil {
			targetID = target.ID
		}
		result := e.dispatchAllRoleTimingHooks(engineplayer.TimingOnDamageCalculate, engineplayer.TimingHookContext{
			SourceID:         attacker.ID,
			TargetID:         targetID,
			ActionType:       action.Type,
			Card:             action.Card,
			Damage:           damage,
			CounterInitiator: action.CounterInitiator,
		})
		damage += result.DamageDelta
		if damage < 0 {
			damage = 0
		}
	}
	return damage
}
