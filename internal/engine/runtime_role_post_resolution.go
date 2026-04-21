// gameflow: 多角色共用后置效果派发（攻击命中后、行动结束后、伤害结算完成后）。
// 所有角色逻辑通过 Timing Hook 分派，engine 不直接调用任何 player 包函数。
package engine

import (
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// handlePostAttackHitEffects 处理"攻击命中后"的角色附加效果。
func (e *GameEngine) handlePostAttackHitEffects(pd *model.PendingDamage) bool {
	if pd == nil {
		return false
	}
	attacker := e.State.Players[pd.SourceID]
	if attacker == nil {
		return false
	}
	if attacker.Tokens == nil {
		attacker.Tokens = map[string]int{}
	}
	result := e.dispatchRoleTimingHook(engineplayer.TimingPostAttackHit, engineplayer.TimingHookContext{
		SourceID:      pd.SourceID,
		TargetID:      pd.TargetID,
		IsCounter:     pd.IsCounter,
		Card:          pd.Card,
		PendingDamage: pd,
	})
	return result.Interrupted
}

// handlePostActionEndEffects 处理行动结束后的场上效果追加结算。
func (e *GameEngine) handlePostActionEndEffects(player *model.Player, actionType model.ActionType) bool {
	if player == nil {
		return false
	}
	if actionType != model.ActionAttack && actionType != model.ActionMagic {
		return false
	}
	result := e.dispatchRoleTimingHook(engineplayer.TimingPostActionEnd, engineplayer.TimingHookContext{
		SourceID:   player.ID,
		ActionType: actionType,
	})
	return result.Interrupted
}

// handlePostDamageResolved 处理"伤害结算完成后"的附加效果。
// 使用 dispatchAllRoleTimingHooks 确保所有角色的 hook 都能执行（不短路）。
func (e *GameEngine) handlePostDamageResolved(pd *model.PendingDamage) bool {
	if pd == nil {
		return false
	}
	source := e.State.Players[pd.SourceID]
	if source == nil {
		return false
	}
	result := e.dispatchAllRoleTimingHooks(engineplayer.TimingPostDamageResolved, engineplayer.TimingHookContext{
		SourceID:      pd.SourceID,
		TargetID:      pd.TargetID,
		DamageType:    pd.DamageType,
		Damage:        pd.Damage,
		PendingDamage: pd,
	})
	return result.Interrupted
}
