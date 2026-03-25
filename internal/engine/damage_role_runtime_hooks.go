package engine

import "starcup-engine/internal/model"

type pendingDamageAttackInitHook func(e *GameEngine, pd *model.PendingDamage, attacker *model.Player, victim *model.Player)
type pendingDamageHealCapHook func(e *GameEngine, pd *model.PendingDamage, target *model.Player, maxHeal int) int
type pendingDamageAfterApplyHook func(e *GameEngine, pd *model.PendingDamage, target *model.Player)

var pendingDamageAttackInitHooks = []pendingDamageAttackInitHook{
	pendingDamageHeroRoarMissArmHook,
	pendingDamageFighterChargeMissArmHook,
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
	if e == nil || pd == nil || attacker == nil || !e.isHero(attacker) || attacker.Tokens == nil {
		return
	}
	if attacker.Tokens["hero_roar_active"] > 0 {
		pd.HeroRoarMissArmed = true
	}
}

func pendingDamageFighterChargeMissArmHook(e *GameEngine, pd *model.PendingDamage, attacker *model.Player, _ *model.Player) {
	if e == nil || pd == nil || attacker == nil || !e.isFighter(attacker) || attacker.Tokens == nil {
		return
	}
	if attacker.Tokens["fighter_charge_pending"] > 0 {
		pd.FighterChargeMissArmed = true
	}
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
