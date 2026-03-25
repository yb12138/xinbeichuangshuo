package engine

import (
	"fmt"
	"strings"

	"starcup-engine/internal/model"
)

type attackTargetContextHook func(e *GameEngine, player *model.Player, targetID string)
type attackStartStateResetHook func(e *GameEngine, player *model.Player)
type attackPreCombatHook func(e *GameEngine, player *model.Player, target *model.Player, currentAction *model.QueuedAction, eventCtx *model.EventContext)

var attackTargetContextHooks = []attackTargetContextHook{
	recordMagicBowAttackTargetOrder,
}

var attackStartStateResetHooks = []attackStartStateResetHook{
	resetHolyLancerAttackFlags,
	resetBerserkerAttackFlags,
	resetSwordEmperorAttackFlags,
	resetBeastSamuraiAttackFlags,
	resetMagicSwordsmanAttackFlags,
	resetFighterAttackFlags,
}

var attackPreCombatHooks = []attackPreCombatHook{
	applyHeroAttackGating,
	applyFighterAttackGating,
	applyMoonGoddessAttackGating,
	applyAssassinAttackGating,
	applyHolyLancerAttackGating,
	applyMagicSwordsmanAttackGating,
	applyElfArcherAttackGating,
	applyDarkElementNoCounterRule,
	applyBeastSamuraiAttackGating,
}

func (e *GameEngine) recordAttackTargetContext(player *model.Player, targetID string) {
	for _, hook := range attackTargetContextHooks {
		hook(e, player, targetID)
	}
}

func (e *GameEngine) runAttackStartStateResets(player *model.Player) {
	for _, hook := range attackStartStateResetHooks {
		hook(e, player)
	}
}

func (e *GameEngine) applyAttackPreCombatRoleRules(player *model.Player, target *model.Player, currentAction *model.QueuedAction, eventCtx *model.EventContext) {
	for _, hook := range attackPreCombatHooks {
		hook(e, player, target, currentAction, eventCtx)
	}
}

func recordMagicBowAttackTargetOrder(e *GameEngine, player *model.Player, targetID string) {
	if !e.isMagicBow(player) || player.TurnState.UsedSkillCounts == nil {
		return
	}
	for i, pid := range e.State.PlayerOrder {
		if pid == targetID {
			player.TurnState.UsedSkillCounts["mb_last_attack_target_order"] = i + 1
			return
		}
	}
}

func resetHolyLancerAttackFlags(_ *GameEngine, player *model.Player) {
	ensurePlayerTokensMap(player)
	player.Tokens["holy_lancer_block_sacred_strike"] = 0
	player.Tokens["holy_lancer_sky_spear_no_counter"] = 0
}

func resetBerserkerAttackFlags(_ *GameEngine, player *model.Player) {
	ensurePlayerTokensMap(player)
	player.Tokens["berserker_blood_roar_ignore_shield"] = 0
}

func resetSwordEmperorAttackFlags(_ *GameEngine, player *model.Player) {
	ensurePlayerTokensMap(player)
	player.Tokens["se_guard_disabled_current_attack"] = 0
	player.Tokens["se_angel_soul_armed"] = 0
	player.Tokens["se_demon_soul_armed"] = 0
	player.Tokens["se_demon_damage_bonus_pending"] = 0
}

func resetBeastSamuraiAttackFlags(_ *GameEngine, player *model.Player) {
	ensurePlayerTokensMap(player)
	clearBeastSamuraiAttackTokens(player)
}

func resetMagicSwordsmanAttackFlags(_ *GameEngine, player *model.Player) {
	ensurePlayerTokensMap(player)
	player.Tokens["ms_yellow_spring_pending"] = 0
}

func resetFighterAttackFlags(_ *GameEngine, player *model.Player) {
	ensurePlayerTokensMap(player)
	player.Tokens["fighter_attack_start_skill_lock"] = 0
}

func applyHeroAttackGating(_ *GameEngine, player *model.Player, _ *model.Player, _ *model.QueuedAction, eventCtx *model.EventContext) {
	if player == nil || player.Tokens == nil || player.Tokens["hero_calm_force_no_counter"] <= 0 || eventCtx == nil || eventCtx.AttackInfo == nil {
		return
	}
	eventCtx.AttackInfo.CanBeResponded = false
	player.Tokens["hero_calm_force_no_counter"] = 0
}

func applyFighterAttackGating(_ *GameEngine, player *model.Player, _ *model.Player, _ *model.QueuedAction, eventCtx *model.EventContext) {
	if player == nil || player.Tokens == nil || player.Tokens["fighter_qiburst_force_no_counter"] <= 0 || eventCtx == nil || eventCtx.AttackInfo == nil {
		return
	}
	eventCtx.AttackInfo.CanBeResponded = false
	player.Tokens["fighter_qiburst_force_no_counter"] = 0
}

func applyMoonGoddessAttackGating(_ *GameEngine, player *model.Player, _ *model.Player, _ *model.QueuedAction, eventCtx *model.EventContext) {
	if player == nil || player.Tokens == nil || player.Tokens["mg_next_attack_no_counter"] <= 0 || eventCtx == nil || eventCtx.AttackInfo == nil {
		return
	}
	eventCtx.AttackInfo.CanBeResponded = false
	player.Tokens["mg_next_attack_no_counter"]--
	if player.Tokens["mg_next_attack_no_counter"] < 0 {
		player.Tokens["mg_next_attack_no_counter"] = 0
	}
}

func applyAssassinAttackGating(e *GameEngine, player *model.Player, _ *model.Player, _ *model.QueuedAction, eventCtx *model.EventContext) {
	if !isCharacter(player, "assassin") || !hasAssassinStealthForm(player) || eventCtx == nil || eventCtx.AttackInfo == nil {
		return
	}
	ensurePlayerTokensMap(player)
	eventCtx.AttackInfo.CanBeResponded = false
	e.Log(fmt.Sprintf("[Skill] %s 处于[潜行]：本次主动攻击无法应战", player.Name))
}

func applyHolyLancerAttackGating(e *GameEngine, player *model.Player, _ *model.Player, _ *model.QueuedAction, eventCtx *model.EventContext) {
	if !e.isHolyLancer(player) || player.Tokens == nil || player.Tokens["holy_lancer_sky_spear_no_counter"] <= 0 || eventCtx == nil || eventCtx.AttackInfo == nil {
		return
	}
	eventCtx.AttackInfo.CanBeResponded = false
	player.Tokens["holy_lancer_sky_spear_no_counter"] = 0
}

func applyMagicSwordsmanAttackGating(e *GameEngine, player *model.Player, _ *model.Player, _ *model.QueuedAction, eventCtx *model.EventContext) {
	if !e.isMagicSwordsman(player) || player.Tokens == nil || player.Tokens["ms_yellow_spring_pending"] <= 0 || eventCtx == nil || eventCtx.AttackInfo == nil {
		return
	}
	eventCtx.AttackInfo.CanBeResponded = false
}

func applyElfArcherAttackGating(e *GameEngine, player *model.Player, _ *model.Player, _ *model.QueuedAction, eventCtx *model.EventContext) {
	if !e.isElfArcher(player) || player.Tokens == nil || player.Tokens["elf_elemental_shot_thunder_pending"] <= 0 || eventCtx == nil || eventCtx.AttackInfo == nil {
		return
	}
	eventCtx.AttackInfo.CanBeResponded = false
}

func applyDarkElementNoCounterRule(_ *GameEngine, _ *model.Player, _ *model.Player, currentAction *model.QueuedAction, eventCtx *model.EventContext) {
	if currentAction == nil || currentAction.Card == nil || currentAction.Card.Element != model.ElementDark || eventCtx == nil || eventCtx.AttackInfo == nil {
		return
	}
	eventCtx.AttackInfo.CanBeResponded = false
}

func applyBeastSamuraiAttackGating(e *GameEngine, player *model.Player, _ *model.Player, currentAction *model.QueuedAction, eventCtx *model.EventContext) {
	if !e.isBeastSamurai(player) || player.Tokens == nil || player.Tokens["bs_one_strike_armed"] <= 0 || eventCtx == nil || eventCtx.AttackInfo == nil {
		return
	}
	player.Tokens["bs_one_strike_armed"] = 0
	player.Tokens["bs_ignore_shield_current_attack"] = 1
	player.Tokens["bs_no_holy_defend_current_attack"] = 1
	if currentAction != nil && currentAction.Card != nil && strings.TrimSpace(currentAction.Card.Faction) == "技" {
		eventCtx.AttackInfo.IsHitForced = true
	}
	e.Log(fmt.Sprintf("%s 的 [一击无念·下次攻击劫持] 生效：本次主动攻击无视圣盾、不可用圣光抵挡%s", player.Name, func() string {
		if currentAction != nil && currentAction.Card != nil && strings.TrimSpace(currentAction.Card.Faction) == "技" {
			return "，且技命格攻击强制命中"
		}
		return ""
	}()))
}
