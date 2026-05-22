// gameflow: 按角色身份挂载的承伤/转伤/命中钩子。
// 合并自 damage_attack_hit_hooks.go。

package engine

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// ---------- 命中判定规则（原 damage_attack_hit_hooks.go） ----------

// applyAttackHitPendingDamageRules 在命中判定时处理攻击伤害命中规则。
func (e *GameEngine) applyAttackHitPendingDamageRules(pd *model.PendingDamage, attacker *model.Player, victim *model.Player) {
	result := e.dispatchRoleTimingHook(engineplayer.TimingOnHitCheck, engineplayer.TimingHookContext{
		SourceID:      pd.SourceID,
		TargetID:      pd.TargetID,
		IsCounter:     pd.IsCounter,
		PendingDamage: pd,
	})
	// 如果有 Interrupted 或 Blocked，则后续逻辑已由角色包处理
	if result.Interrupted || result.Blocked {
		return
	}
}

// ---------- 承伤/转伤钩子 ----------

// applyDamageTargetBeforeRules 在承伤触发前处理伤害计算阶段规则。
func (e *GameEngine) applyDamageTargetBeforeRules(pd *model.PendingDamage) bool {
	result := e.dispatchRoleTimingHook(engineplayer.TimingOnDamageBeforeTaken, engineplayer.TimingHookContext{
		SourceID:      pd.SourceID,
		TargetID:      pd.TargetID,
		PendingDamage: pd,
		DamageType:    pd.DamageType,
		Damage:        pd.Damage,
	})
	if result.Interrupted {
		return true
	}
	// 治疗抵抗门禁：触发 TimingOnHealResist hooks
	if pd != nil && !pd.IgnoreHeal {
		target := e.State.Players[pd.TargetID]
		if target != nil {
			e.dispatchAllRoleTimingHooks(engineplayer.TimingOnHealResist, engineplayer.TimingHookContext{
				TargetID:      target.ID,
				PendingDamage: pd,
			})
		}
	}
	return false
}

// applyTimingOnDamageTakenAfterTakenRules 在承伤触发后处理后续规则。
func (e *GameEngine) applyTimingOnDamageTakenAfterTakenRules(pd *model.PendingDamage) bool {
	result := e.dispatchRoleTimingHook(engineplayer.TimingOnDamageAfterTaken, engineplayer.TimingHookContext{
		SourceID:      pd.SourceID,
		TargetID:      pd.TargetID,
		PendingDamage: pd,
		DamageType:    pd.DamageType,
		Damage:        pd.Damage,
	})
	return result.Interrupted
}

// applyTimingOnDamageAppliedRules 在⑤实际产生伤害时处理扣除治疗后的响应。
func (e *GameEngine) applyTimingOnDamageAppliedRules(pd *model.PendingDamage) bool {
	result := e.dispatchRoleTimingHook(engineplayer.TimingOnDamageApplied, engineplayer.TimingHookContext{
		SourceID:      pd.SourceID,
		TargetID:      pd.TargetID,
		PendingDamage: pd,
		DamageType:    pd.DamageType,
		Damage:        pd.Damage,
	})
	return result.Interrupted
}

// applyTimingOnDamageTakenRules 在⑥实际承受伤害、准备摸牌前处理响应。
func (e *GameEngine) applyTimingOnDamageTakenRules(pd *model.PendingDamage) bool {
	result := e.dispatchRoleTimingHook(engineplayer.TimingOnDamageTaken, engineplayer.TimingHookContext{
		SourceID:      pd.SourceID,
		TargetID:      pd.TargetID,
		PendingDamage: pd,
		DamageType:    pd.DamageType,
		Damage:        pd.Damage,
	})
	return result.Interrupted
}

// applyHealCapRules 在治疗抵伤额度计算时应用上限规则。
func (e *GameEngine) applyHealCapRules(pd *model.PendingDamage, target *model.Player, maxHeal int) int {
	result := e.dispatchAllRoleTimingHooks(engineplayer.TimingOnHealCapCalculate, engineplayer.TimingHookContext{
		TargetID:      target.ID,
		PendingDamage: pd,
		HealCap:       maxHeal,
	})
	limited := maxHeal + result.HealCapDelta
	if limited < 0 {
		limited = 0
	}
	return limited
}

// applyTimingOnDamageTakenAfterApplyRules 在伤害应用后执行后置清理。
func (e *GameEngine) applyTimingOnDamageTakenAfterApplyRules(pd *model.PendingDamage, target *model.Player) {
	// 封印师五系封印清理（系统级，不迁移到角色包）
	if pd != nil && target != nil && model.IsElementalSealEffect(pd.EffectTypeToRemove) {
		if e.RemoveFieldCard(target.ID, pd.EffectTypeToRemove) {
			e.Log(fmt.Sprintf("[Seal] %s 的 %s 伤害结算后被移除", target.Name, pd.EffectTypeToRemove))
		}
	}
	if pd != nil && target != nil {
		e.dispatchAllRoleTimingHooks(engineplayer.TimingOnDamageAfterApply, engineplayer.TimingHookContext{
			SourceID:      pd.SourceID,
			TargetID:      target.ID,
			PendingDamage: pd,
		})
	}
}

// applyTimingOnDamageTakenAfterResolvedRules 在整次伤害出队后处理结算后规则。
func (e *GameEngine) applyTimingOnDamageTakenAfterResolvedRules(pd *model.PendingDamage) bool {
	return e.HandlePostDamageResolved(pd)
}
