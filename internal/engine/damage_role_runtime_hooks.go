package engine

import "starcup-engine/internal/model"

type pendingDamageAttackInitHook func(e *GameEngine, pd *model.PendingDamage, attacker *model.Player, victim *model.Player)
type pendingDamageBeforeTakenHook func(e *GameEngine, pd *model.PendingDamage) bool
type pendingDamageHealResistRule func(e *GameEngine, pd *model.PendingDamage, target *model.Player) bool
type pendingDamageHealCapHook func(e *GameEngine, pd *model.PendingDamage, target *model.Player, maxHeal int) int
type pendingDamageAfterApplyHook func(e *GameEngine, pd *model.PendingDamage, target *model.Player)

var pendingDamageAttackInitHooks = []pendingDamageAttackInitHook{
	pendingDamageHeroRoarMissArmHook,
	pendingDamageFighterChargeMissArmHook,
}

var pendingDamageBeforeTakenHooks = []pendingDamageBeforeTakenHook{
	pendingDamageHealResistGateHook,
	pendingDamageSoulLinkTransferHook,
}

var pendingDamageHealResistRules = []pendingDamageHealResistRule{
	pendingDamageRoseCourtyardHealResistRule,
	pendingDamageCrimsonKnightHealResistRule,
	pendingDamagePlagueMageHealResistRule,
}

var pendingDamageHealCapHooks = []pendingDamageHealCapHook{
	pendingDamagePriestHealCapHook,
}

var pendingDamageAfterApplyHooks = []pendingDamageAfterApplyHook{
	pendingDamageResetCrimsonSwordSpiritLocksHook,
	pendingDamageResetBlazeWitchLocksHook,
}

func (e *GameEngine) runPendingDamageAttackInitHooks(pd *model.PendingDamage, attacker *model.Player, victim *model.Player) {
	for _, hook := range pendingDamageAttackInitHooks {
		if hook != nil {
			hook(e, pd, attacker, victim)
		}
	}
}

func (e *GameEngine) runPendingDamageBeforeTakenHooks(pd *model.PendingDamage) bool {
	for _, hook := range pendingDamageBeforeTakenHooks {
		if hook == nil {
			continue
		}
		if hook(e, pd) {
			return true
		}
	}
	return false
}

func (e *GameEngine) applyPendingDamageHealCapHooks(pd *model.PendingDamage, target *model.Player, maxHeal int) int {
	limited := maxHeal
	for _, hook := range pendingDamageHealCapHooks {
		if hook == nil {
			continue
		}
		limited = hook(e, pd, target, limited)
		if limited < 0 {
			limited = 0
		}
	}
	return limited
}

func (e *GameEngine) runPendingDamageAfterApplyHooks(pd *model.PendingDamage, target *model.Player) {
	for _, hook := range pendingDamageAfterApplyHooks {
		if hook != nil {
			hook(e, pd, target)
		}
	}
}

func pendingDamageHeroRoarMissArmHook(e *GameEngine, pd *model.PendingDamage, attacker *model.Player, _ *model.Player) {
	if e == nil || pd == nil || attacker == nil || !e.isHero(attacker) {
		return
	}
	if attacker.TurnState.UsedSkillCounts["hero_roar_active"] > 0 {
		pd.HeroRoarMissArmed = true
	}
}

func pendingDamageFighterChargeMissArmHook(e *GameEngine, pd *model.PendingDamage, attacker *model.Player, _ *model.Player) {
	if e == nil || pd == nil || attacker == nil || !e.isFighter(attacker) || attacker.Tokens == nil {
		return
	}
	if attacker.TurnState.SkillFlowState["fighter_charge_pending"] > 0 {
		pd.FighterChargeMissArmed = true
	}
}

func pendingDamageSoulLinkTransferHook(e *GameEngine, pd *model.PendingDamage) bool {
	if e == nil || pd == nil {
		return false
	}
	return e.maybeTriggerSoulLinkTransfer(pd)
}

// pendingDamageHealResistGateHook 统一承载“治疗抵伤门禁”的角色技能规则。
// 这类规则统一通过修改 PendingDamage 标记（IgnoreHeal）生效。
// 具体规则由 pendingDamageHealResistRules 注册，避免把角色分支硬编码在入口函数。
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
	for _, rule := range pendingDamageHealResistRules {
		if rule == nil {
			continue
		}
		if rule(e, pd, target) {
			pd.IgnoreHeal = true
			break
		}
	}
	return false
}

// 血蔷薇庭院：场上效果在场期间，所有伤害均不可被治疗抵伤。
func pendingDamageRoseCourtyardHealResistRule(e *GameEngine, _ *model.PendingDamage, _ *model.Player) bool {
	return e != nil && e.isRoseCourtyardActive()
}

// 红莲骑士：仅允许“腥红信仰白名单”中的自伤使用治疗抵御。
func pendingDamageCrimsonKnightHealResistRule(e *GameEngine, pd *model.PendingDamage, target *model.Player) bool {
	if e == nil || pd == nil || target == nil || !e.isCrimsonKnight(target) {
		return false
	}
	return target.ID != pd.SourceID || !pd.AllowCrimsonFaithHeal
}

// 瘟疫法师圣渎：攻击伤害不可用治疗抵挡，法术伤害可以。
func pendingDamagePlagueMageHealResistRule(e *GameEngine, pd *model.PendingDamage, target *model.Player) bool {
	if e == nil || pd == nil || target == nil || !e.isPlagueMage(target) {
		return false
	}
	return pd.DamageType == "Attack"
}

func pendingDamagePriestHealCapHook(e *GameEngine, _ *model.PendingDamage, target *model.Player, maxHeal int) int {
	if e == nil || target == nil || !e.isPriest(target) || maxHeal <= 1 {
		return maxHeal
	}
	return 1
}

func pendingDamageResetCrimsonSwordSpiritLocksHook(_ *GameEngine, _ *model.PendingDamage, target *model.Player) {
	if target == nil || target.Tokens == nil {
		return
	}
	target.Tokens["css_blood_barrier_lock"] = 0
}

func pendingDamageResetBlazeWitchLocksHook(_ *GameEngine, _ *model.PendingDamage, target *model.Player) {
	if target == nil || target.Tokens == nil {
		return
	}
	target.Tokens["bw_substitute_lock"] = 0
	target.Tokens["bw_mana_inversion_lock"] = 0
}
