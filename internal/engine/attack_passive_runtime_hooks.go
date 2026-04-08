package engine

import (
	"fmt"
	"strings"

	"starcup-engine/internal/model"
)

type attackPassiveDamageHook func(e *GameEngine, attacker *model.Player, target *model.Player, action model.Action, damage int) int

// applyTimingOnDamageCalculatedAttackPassiveModifiers 在伤害计算时按固定顺序应用攻击方被动修正。
func (e *GameEngine) applyTimingOnDamageCalculatedAttackPassiveModifiers(attacker *model.Player, target *model.Player, action model.Action, baseDamage int) int {
	damage := baseDamage
	for _, hook := range e.damageCalculatedAttackPassiveHooks {
		damage = hook(e, attacker, target, action, damage)
		if damage < 0 {
			damage = 0
		}
	}
	return damage
}

func attackPassiveElfFireShotHook(_ *GameEngine, attacker *model.Player, _ *model.Player, action model.Action, damage int) int {
	if attacker == nil || consumeAttackDamageRuleBonus(attacker, "elf_elemental_shot_fire_attack_bonus", action) <= 0 {
		return damage
	}
	return damage + 1
}

func attackPassiveMagicSwordsmanShadowHook(e *GameEngine, attacker *model.Player, _ *model.Player, _ model.Action, damage int) int {
	if e == nil || attacker == nil || !isCharacter(attacker, "magic_swordsman") || !hasMagicSwordsmanShadowForm(attacker) {
		return damage
	}
	e.Log(fmt.Sprintf("[Passive] %s 的 [暗影之力] 生效，伤害 +1", attacker.Name))
	return damage + 1
}

func attackPassiveMagicLancerBonusHook(e *GameEngine, attacker *model.Player, _ *model.Player, action model.Action, damage int) int {
	if e == nil || attacker == nil || !isCharacter(attacker, "magic_lancer") || action.Type != model.ActionAttack || action.CounterInitiator != "" {
		return damage
	}

	if darkBonus := consumeAttackDamageRuleBonus(attacker, "ml_dark_release_next_attack_bonus", action); darkBonus > 0 {
		damage += darkBonus
		e.Log(fmt.Sprintf("[Passive] %s 的 [暗之解放] 生效，本次主动攻击伤害 +1", attacker.Name))
	}

	if additional := consumeAttackDamageRuleBonus(attacker, "ml_fullness_next_attack_bonus", action); additional > 0 {
		damage += additional
		e.Log(fmt.Sprintf("[Passive] %s 的 [充盈] 生效，本次主动攻击伤害 +%d", attacker.Name, additional))
	}

	return damage
}

func attackPassiveFighterBonusHook(e *GameEngine, attacker *model.Player, _ *model.Player, action model.Action, damage int) int {
	if e == nil || attacker == nil || !isCharacter(attacker, "fighter") {
		return damage
	}

	if action.Type == model.ActionAttack &&
		action.CounterInitiator == "" &&
		consumeAttackDamageRuleBonus(attacker, "fighter_charge_attack_bonus", action) > 0 {
		damage += 1
		attacker.TurnState.SkillFlowState["fighter_charge_pending"] = 0
		e.Log(fmt.Sprintf("[Passive] %s 的 [蓄力一击] 生效，本次主动攻击伤害 +1", attacker.Name))
	}

	if hasFighterHundredDragonForm(attacker) {
		if action.Type == model.ActionAttack && action.CounterInitiator == "" {
			damage += 2
			e.Log(fmt.Sprintf("[Passive] %s 的 [百式幻龙拳] 生效，本次主动攻击伤害 +2", attacker.Name))
		} else if action.Type == model.ActionAttack && action.CounterInitiator != "" {
			damage += 1
			e.Log(fmt.Sprintf("[Passive] %s 的 [百式幻龙拳] 生效，本次应战攻击伤害 +1", attacker.Name))
		}
	}

	return damage
}

func attackPassiveHeroRoarBonusHook(e *GameEngine, attacker *model.Player, _ *model.Player, action model.Action, damage int) int {
	if e == nil || attacker == nil || !isCharacter(attacker, "hero") || action.Type != model.ActionAttack || action.CounterInitiator != "" {
		return damage
	}

	if consumeAttackDamageRuleBonus(attacker, "hero_roar_attack_bonus", action) <= 0 {
		return damage
	}

	attacker.TurnState.UsedSkillCounts["hero_roar_active"] = 0
	e.Log(fmt.Sprintf("[Passive] %s 的 [怒吼] 生效，本次主动攻击伤害 +2", attacker.Name))
	return damage + 2
}

func attackPassiveAssassinStealthBonusHook(e *GameEngine, attacker *model.Player, _ *model.Player, action model.Action, damage int) int {
	if e == nil || attacker == nil || !isCharacter(attacker, "assassin") || action.Type != model.ActionAttack || action.CounterInitiator != "" {
		return damage
	}
	extra := 0
	// 兜底：若攻击前钩子未写入一次性加成，但攻击者仍处于潜行形态，则按实时能量计算。
	if hasAssassinStealthForm(attacker) {
		extra = attacker.Gem + attacker.Crystal
	}
	if extra <= 0 {
		return damage
	}
	e.Log(fmt.Sprintf("[Passive] %s 处于[潜行]，本次主动攻击伤害额外+%d（剩余能量）", attacker.Name, extra))
	return damage + extra
}

func attackPassiveHolyBowPenaltyHook(e *GameEngine, attacker *model.Player, _ *model.Player, action model.Action, damage int) int {
	if e == nil || attacker == nil || !isCharacter(attacker, "holy_bow") || action.Type != model.ActionAttack || action.CounterInitiator != "" || action.Card == nil {
		return damage
	}
	if strings.TrimSpace(action.Card.Faction) == "圣" {
		return damage
	}
	e.Log(fmt.Sprintf("[Passive] %s 的 [天之弓] 生效：非圣命格主动攻击伤害 -1", attacker.Name))
	return damage - 1
}

func attackPassiveSwordEmperorBonusHook(e *GameEngine, attacker *model.Player, _ *model.Player, action model.Action, damage int) int {
	if e == nil || attacker == nil || !e.isSwordEmperor(attacker) || action.Type != model.ActionAttack || action.CounterInitiator != "" {
		return damage
	}

	if consumeAttackDamageRuleBonus(attacker, "se_demon_soul_attack_bonus", action) <= 0 {
		return damage
	}

	e.Log(fmt.Sprintf("[Passive] %s 的 [恶魔之魂] 生效：本次主动攻击伤害 +1", attacker.Name))
	return damage + 1
}

func attackPassiveBeastSamuraiBonusHook(e *GameEngine, attacker *model.Player, target *model.Player, action model.Action, damage int) int {
	if e == nil || attacker == nil || target == nil || !e.isBeastSamurai(attacker) || action.Type != model.ActionAttack || action.CounterInitiator != "" {
		return damage
	}
	if !e.beastSamuraiInIaijutsuForm(attacker) || effectivePlayerOrientation(target) != model.OrientationTapped {
		return damage
	}
	e.Log(fmt.Sprintf("[Passive] %s 的 [御魂流居合形态·横置目标增伤] 生效，本次主动攻击伤害 +1", attacker.Name))
	return damage + 1
}
