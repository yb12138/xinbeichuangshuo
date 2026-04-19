// gameflow: 多角色共用后置效果派发（攻击命中后、行动结束后、伤害结算完成后）。
package engine

import (
	"starcup-engine/internal/engine/core/runtimeutil"
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// handlePostAttackHitEffects 处理"攻击命中后"的角色附加效果。
// 通过 Timing Hook 分派，具体逻辑在各角色 player/<role>/timing_hooks.go 中实现。
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
	return e.dispatchRoleTimingHook(engineplayer.TimingPostAttackHit, engineplayer.TimingHookContext{
		SourceID:      pd.SourceID,
		TargetID:      pd.TargetID,
		IsCounter:     pd.IsCounter,
		Card:          pd.Card,
		PendingDamage: pd,
	})
}

// handlePostActionEndEffects 处理行动结束后的场上效果追加结算。
// 通过 Timing Hook 分派，具体逻辑在各角色 player/<role>/timing_hooks.go 中实现。
func (e *GameEngine) handlePostActionEndEffects(player *model.Player, actionType model.ActionType) bool {
	if player == nil {
		return false
	}
	if actionType != model.ActionAttack && actionType != model.ActionMagic {
		return false
	}
	return e.dispatchRoleTimingHook(engineplayer.TimingPostActionEnd, engineplayer.TimingHookContext{
		SourceID:   player.ID,
		ActionType: actionType,
	})
}

// handlePostDamageResolved 处理"伤害结算完成后"的附加效果。
// 已迁移到 Timing Hook 的角色：BeastSamurai、BlazeWitch、MagicLancer、Sage。
// 仍保留在 engine 的角色：Bard（降灵）、ElfArcher（动物响应）、MoonGoddess（亵渎）。
func (e *GameEngine) handlePostDamageResolved(pd *model.PendingDamage) bool {
	if pd == nil {
		return false
	}
	source := e.State.Players[pd.SourceID]
	if source == nil {
		return false
	}
	// Phase 1: 声明式 Hook 分派（BeastSamurai、BlazeWitch、MagicLancer、Sage）
	if e.dispatchRoleTimingHook(engineplayer.TimingPostDamageResolved, engineplayer.TimingHookContext{
		SourceID:      pd.SourceID,
		TargetID:      pd.TargetID,
		DamageType:    pd.DamageType,
		Damage:        pd.Damage,
		PendingDamage: pd,
	}) {
		_ = e.tryQueueMoonGoddessBlasphemy(pd)
		return true
	}
	// Phase 2: 仍需 engine 状态的剩余逻辑
	if pd.Damage <= 0 {
		return false
	}
	target := e.State.Players[pd.TargetID]
	if pd.Damage > 0 && runtimeutil.IsMagicDamageType(pd.DamageType) {
		if e.tryBardDescentAfterMagicDamage(pd) {
			_ = e.tryQueueMoonGoddessBlasphemy(pd)
			return true
		}
	}
	if e.queueElfAnimalResponse(source, target, pd) {
		_ = e.tryQueueMoonGoddessBlasphemy(pd)
		return true
	}
	if e.tryQueueMoonGoddessBlasphemy(pd) {
		return true
	}
	return false
}
