package engine

import (
	"fmt"
	"strings"

	"starcup-engine/internal/model"
)

type attackPassiveDamageHook func(e *GameEngine, attacker *model.Player, target *model.Player, action model.Action, damage int) int

var attackPassiveDamageHooks = []attackPassiveDamageHook{
	attackPassiveElfFireShotHook,
	attackPassiveMagicSwordsmanShadowHook,
	attackPassiveMagicLancerBonusHook,
	attackPassiveFighterBonusHook,
	attackPassiveHeroRoarBonusHook,
	attackPassiveAssassinStealthBonusHook,
	attackPassiveHolyBowPenaltyHook,
	attackPassiveSwordEmperorBonusHook,
	attackPassiveBeastSamuraiBonusHook,
}

func (e *GameEngine) applyAttackPassiveRuntimeHooks(attacker *model.Player, target *model.Player, action model.Action, baseDamage int) int {
	damage := baseDamage
	for _, hook := range attackPassiveDamageHooks {
		if hook == nil {
			continue
		}
		damage = hook(e, attacker, target, action, damage)
		if damage < 0 {
			damage = 0
		}
	}
	return damage
}

func attackPassiveElfFireShotHook(e *GameEngine, attacker *model.Player, _ *model.Player, _ model.Action, damage int) int {
	if e == nil || attacker == nil {
		return damage
	}
	ensurePlayerTokensMap(attacker)
	if attacker.Tokens["elf_elemental_shot_fire_pending"] <= 0 {
		return damage
	}
	attacker.Tokens["elf_elemental_shot_fire_pending"] = 0
	e.Log(fmt.Sprintf("[Passive] %s 的 [火之矢] 生效，伤害 +1", attacker.Name))
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
	if attacker.TurnState.UsedSkillCounts == nil {
		attacker.TurnState.UsedSkillCounts = map[string]int{}
	}
	if attacker.TurnState.UsedSkillCounts["ml_dark_release_next_attack_bonus"] > 0 {
		damage += 1
		attacker.TurnState.UsedSkillCounts["ml_dark_release_next_attack_bonus"] = 0
		e.Log(fmt.Sprintf("[Passive] %s 的 [暗之解放] 生效，本次主动攻击伤害 +1", attacker.Name))
	}
	if bonus := attacker.TurnState.UsedSkillCounts["ml_fullness_next_attack_bonus"]; bonus > 0 {
		damage += bonus
		attacker.TurnState.UsedSkillCounts["ml_fullness_next_attack_bonus"] = 0
		e.Log(fmt.Sprintf("[Passive] %s 的 [充盈] 生效，本次主动攻击伤害 +%d", attacker.Name, bonus))
	}
	return damage
}

func attackPassiveFighterBonusHook(e *GameEngine, attacker *model.Player, _ *model.Player, action model.Action, damage int) int {
	if e == nil || attacker == nil || !isCharacter(attacker, "fighter") {
		return damage
	}
	ensurePlayerTokensMap(attacker)
	if action.Type == model.ActionAttack &&
		action.CounterInitiator == "" &&
		attacker.Tokens["fighter_charge_damage_pending"] > 0 {
		damage += 1
		attacker.Tokens["fighter_charge_damage_pending"] = 0
		attacker.Tokens["fighter_charge_pending"] = 0
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
	ensurePlayerTokensMap(attacker)
	if attacker.Tokens["hero_roar_damage_pending"] <= 0 {
		return damage
	}
	attacker.Tokens["hero_roar_damage_pending"] = 0
	attacker.Tokens["hero_roar_active"] = 0
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
	ensurePlayerTokensMap(attacker)
	if attacker.Tokens["se_demon_damage_bonus_pending"] <= 0 {
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
