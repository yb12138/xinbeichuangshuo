// gameflow: 按角色身份挂载的承伤/转伤/命中钩子。
// 合并自 damage_attack_hit_hooks.go。

package engine

import (
	"fmt"
	"strings"

	engineplayer "starcup-engine/internal/engine/player"
	butterflydancer "starcup-engine/internal/engine/player/butterfly_dancer"
	soulsorcererpkg "starcup-engine/internal/engine/player/soul_sorcerer"
	swordemperor "starcup-engine/internal/engine/player/sword_emperor"
	"starcup-engine/internal/model"
)

// ---------- 命中判定规则（原 damage_attack_hit_hooks.go） ----------

// applyTimingOnHitCheckPendingDamageAttackHitRules 在命中判定时处理攻击伤害命中规则。
func (e *GameEngine) applyTimingOnHitCheckPendingDamageAttackHitRules(pd *model.PendingDamage, attacker *model.Player, victim *model.Player) {
	pendingDamageBerserkerBloodRoarHook(e, pd, attacker, victim)
}

func pendingDamageBerserkerBloodRoarHook(e *GameEngine, pd *model.PendingDamage, attacker *model.Player, victim *model.Player) {
	if e == nil || pd == nil || attacker == nil || victim == nil || pd.IsCounter || !isCharacter(attacker, "berserker") || victim.Heal != 2 {
		return
	}
	pd.SetInterceptTag(model.CombatInterceptForceHit)
	pd.SetInterceptTag(model.CombatInterceptIgnoreHolyShield)
	pd.IgnoreHeal = true
	e.Log(fmt.Sprintf("%s 发动 [血腥咆哮]！目标治疗剂为2，强制命中且无视圣盾", attacker.Name))
}

// ---------- 承伤/转伤钩子类型 ----------
type pendingDamageAttackInitHook func(e *GameEngine, pd *model.PendingDamage, attacker *model.Player, victim *model.Player)
type pendingDamageBeforeTakenHook func(e *GameEngine, pd *model.PendingDamage) bool
type pendingDamageAfterTakenHook func(e *GameEngine, pd *model.PendingDamage) bool
type pendingDamageHealResistRule func(e *GameEngine, pd *model.PendingDamage, target *model.Player) bool
type pendingDamageBeforeApplyHook func(e *GameEngine, pd *model.PendingDamage) bool
type pendingDamageHealCapHook func(e *GameEngine, pd *model.PendingDamage, target *model.Player, maxHeal int) int
type pendingDamageAfterApplyHook func(e *GameEngine, pd *model.PendingDamage, target *model.Player)
type pendingDamageAfterResolvedHook func(e *GameEngine, pd *model.PendingDamage) bool

// applyTimingOnAttackDeclaredPendingDamageInitRules 在攻击宣言后初始化伤害运行态。
func (e *GameEngine) applyTimingOnAttackDeclaredPendingDamageInitRules(pd *model.PendingDamage, attacker *model.Player, victim *model.Player) {
	for _, hook := range e.attackDeclaredPendingDamageInitOps {
		hook(e, pd, attacker, victim)
	}
}

// applyTimingOnDamageCalculatedBeforeTakenRules 在承伤触发前处理伤害计算阶段规则。
func (e *GameEngine) applyTimingOnDamageCalculatedBeforeTakenRules(pd *model.PendingDamage) bool {
	for _, hook := range e.damageCalculatedBeforeTakenHooks {
		if hook(e, pd) {
			return true
		}
	}
	return false
}

// applyTimingOnDamageTakenAfterTakenRules 在承伤触发后处理后续规则。
func (e *GameEngine) applyTimingOnDamageTakenAfterTakenRules(pd *model.PendingDamage) bool {
	for _, hook := range e.damageTakenAfterTakenHooks {
		if hook(e, pd) {
			return true
		}
	}
	return false
}

// applyTimingOnDamageAppliedBeforeApplyRules 在真正扣血前处理应用前规则。
func (e *GameEngine) applyTimingOnDamageAppliedBeforeApplyRules(pd *model.PendingDamage) bool {
	for _, hook := range e.damageAppliedBeforeApplyHooks {
		if hook(e, pd) {
			return true
		}
	}
	return false
}

// applyTimingOnDamageCalculatedHealCapRules 在治疗抵伤额度计算时应用上限规则。
func (e *GameEngine) applyTimingOnDamageCalculatedHealCapRules(pd *model.PendingDamage, target *model.Player, maxHeal int) int {
	limited := maxHeal
	for _, hook := range e.damageCalculatedHealCapHooks {
		limited = hook(e, pd, target, limited)
		if limited < 0 {
			limited = 0
		}
	}
	return limited
}

// applyTimingOnDamageTakenAfterApplyRules 在伤害应用后执行后置清理。
func (e *GameEngine) applyTimingOnDamageTakenAfterApplyRules(pd *model.PendingDamage, target *model.Player) {
	for _, hook := range e.damageTakenAfterApplyHooks {
		hook(e, pd, target)
	}
	// 伤害应用后清理已迁移到 TimingOnDamageAfterApply TimingHookSpec。
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
	for _, hook := range e.damageTakenAfterResolvedHooks {
		if hook(e, pd) {
			return true
		}
	}
	return false
}

func pendingDamageSoulLinkTransferHook(e *GameEngine, pd *model.PendingDamage) bool {
	if e == nil || pd == nil {
		return false
	}
	return soulsorcererpkg.MaybeSoulLinkTransfer(newRoleChoiceRuntime(e), pd)
}

// 剑帝命中后置：承伤触发后、治疗抵伤前执行命中分支。
// 使用 AttackPostHitEffectsDone 标记确保同一次伤害只处理一次。
func pendingDamageSwordEmperorAfterTakenHook(e *GameEngine, pd *model.PendingDamage) bool {
	if e == nil || pd == nil || pd.AttackPostHitEffectsDone || pd.HasCheck(model.PendingDamageCheckAttackMissResolved) || !strings.EqualFold(string(pd.DamageType), string(model.AttackDamage)) {
		return false
	}
	swordemperor.ResolveAttackHitAftermath(newRoleChoiceRuntime(e), pd)
	pd.AttackPostHitEffectsDone = true
	return false
}

// 蝶舞者承伤前响应：治疗抵伤处理后、正式扣血前的统一插入点。
func pendingDamageButterflyBeforeApplyHook(e *GameEngine, pd *model.PendingDamage) bool {
	if e == nil || pd == nil {
		return false
	}
	return butterflydancer.MaybeDamageResponses(newRoleChoiceRuntime(e), pd)
}

// pendingDamageHealResistGateHook 统一承载“治疗抵伤门禁”的角色技能规则。
// 这类规则统一通过修改 PendingDamage 标记（IgnoreHeal）生效。
func pendingDamageHealResistGateHook(e *GameEngine, pd *model.PendingDamage) bool {
	if e == nil || pd == nil {
		return false
	}
	if pd.IgnoreHeal {
		return false
	}
	target := e.State.Players[pd.TargetID]
	if target == nil {
		return false
	}
	if e.dispatchTimingOnDamageCalculated(timingOnDamageCalculatedContext{
		Op:            timingOnDamageCalculatedHealResist,
		PendingDamage: pd,
		Target:        target,
	}).IgnoreHeal {
		pd.IgnoreHeal = true
	}
	return false
}

// applyTimingOnDamageCalculatedHealResistRules 在伤害计算阶段判定治疗抵伤门禁。
func (e *GameEngine) applyTimingOnDamageCalculatedHealResistRules(pd *model.PendingDamage, target *model.Player) bool {
	for _, rule := range e.damageCalculatedHealResistRules {
		if rule(e, pd, target) {
			return true
		}
	}
	return false
}

func pendingDamagePriestHealCapHook(e *GameEngine, _ *model.PendingDamage, target *model.Player, maxHeal int) int {
	if e == nil || target == nil || !isCharacter(target, "priest") || maxHeal <= 1 {
		return maxHeal
	}
	return 1
}

// 封印师五系封印：伤害应用后结算封印牌移除。
// 该逻辑由封印技能在 PendingDamage 中写入 EffectTypeToRemove 标记触发，
// 不在伤害主流程硬编码处理。
func pendingDamageElementalSealCleanupHook(e *GameEngine, pd *model.PendingDamage, target *model.Player) {
	if e == nil || pd == nil || target == nil || !model.IsElementalSealEffect(pd.EffectTypeToRemove) {
		return
	}
	if !e.RemoveFieldCard(target.ID, pd.EffectTypeToRemove) {
		return
	}
	e.Log(fmt.Sprintf("[Seal] %s 的 %s 伤害结算后被移除", target.Name, pd.EffectTypeToRemove))
}

func pendingDamageRolePostResolvedHook(e *GameEngine, pd *model.PendingDamage) bool {
	if e == nil || pd == nil {
		return false
	}
	return e.handlePostDamageResolved(pd)
}
