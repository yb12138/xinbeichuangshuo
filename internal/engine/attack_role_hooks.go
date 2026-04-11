// gameflow: 攻击链上按攻击者/防御者身份挂载的运行时钩子。

package engine

import (
	"fmt"
	"strings"

	"starcup-engine/internal/model"
)

type attackTargetContextHook func(e *GameEngine, player *model.Player, targetID string)
type attackStartStateResetHook func(e *GameEngine, player *model.Player)
type attackPreCombatHook func(e *GameEngine, player *model.Player, target *model.Player, currentAction *model.QueuedAction, eventCtx *model.EventContext)

// recordTimingOnAttackDeclaredTargetContext 在攻击宣言时写入目标上下文。
func (e *GameEngine) recordTimingOnAttackDeclaredTargetContext(player *model.Player, targetID string) {
	for _, hook := range e.attackDeclaredTargetContextHooks {
		hook(e, player, targetID)
	}
}

// resetTimingOnAttackDeclaredState 在攻击宣言时清理一次性状态。
func (e *GameEngine) resetTimingOnAttackDeclaredState(player *model.Player) {
	for _, hook := range e.attackDeclaredStateResetHooks {
		hook(e, player)
	}
}

// applyTimingOnAttackDeclaredPreCombatRules 在进入战斗交互前应用攻击劫持策略。
func (e *GameEngine) applyTimingOnAttackDeclaredPreCombatRules(player *model.Player, target *model.Player, currentAction *model.QueuedAction, eventCtx *model.EventContext) {
	for _, hook := range e.attackDeclaredPreCombatHooks {
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
	player.TurnState.UsedSkillCounts["holy_lancer_block_sacred_strike"] = 0
	player.TurnState.UsedSkillCounts["holy_lancer_sky_spear_no_counter"] = 0
}

func resetSwordEmperorAttackFlags(_ *GameEngine, player *model.Player) {
	player.TurnState.UsedSkillCounts["se_guard_disabled_current_attack"] = 0
	player.TurnState.UsedSkillCounts["se_angel_soul_armed"] = 0
	player.TurnState.UsedSkillCounts["se_demon_soul_armed"] = 0
}

func resetBeastSamuraiAttackFlags(_ *GameEngine, player *model.Player) {
	clearBeastSamuraiAttackTokens(player)
}

func resetMagicSwordsmanAttackFlags(_ *GameEngine, player *model.Player) {
	player.TurnState.UsedSkillCounts["ms_yellow_spring_pending"] = 0
}

func resetFighterAttackFlags(_ *GameEngine, player *model.Player) {
	player.TurnState.SkillFlowState["fighter_attack_start_skill_lock"] = 0
}

func applyHeroAttackGating(_ *GameEngine, player *model.Player, _ *model.Player, _ *model.QueuedAction, eventCtx *model.EventContext) {
	if player == nil || player.TurnState.UsedSkillCounts["hero_calm_force_no_counter"] <= 0 || eventCtx == nil || eventCtx.AttackInfo == nil {
		return
	}
	eventCtx.AttackInfo.SetInterceptTag(model.CombatInterceptUnrespondable)
	player.TurnState.UsedSkillCounts["hero_calm_force_no_counter"] = 0
}

func applyFighterAttackGating(_ *GameEngine, player *model.Player, _ *model.Player, _ *model.QueuedAction, eventCtx *model.EventContext) {
	if player == nil || player.TurnState.SkillFlowState["fighter_qiburst_force_no_counter"] <= 0 || eventCtx == nil || eventCtx.AttackInfo == nil {
		return
	}
	eventCtx.AttackInfo.SetInterceptTag(model.CombatInterceptUnrespondable)
	player.TurnState.SkillFlowState["fighter_qiburst_force_no_counter"] = 0
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

func applyMoonGoddessAttackGating(_ *GameEngine, player *model.Player, _ *model.Player, _ *model.QueuedAction, eventCtx *model.EventContext) {
	if player == nil || player.TurnState.UsedSkillCounts["mg_next_attack_no_counter"] <= 0 || eventCtx == nil || eventCtx.AttackInfo == nil {
		return
	}
	eventCtx.AttackInfo.SetInterceptTag(model.CombatInterceptUnrespondable)
	player.TurnState.UsedSkillCounts["mg_next_attack_no_counter"]--
	if player.TurnState.UsedSkillCounts["mg_next_attack_no_counter"] < 0 {
		player.TurnState.UsedSkillCounts["mg_next_attack_no_counter"] = 0
	}
}

func applyAssassinAttackGating(e *GameEngine, player *model.Player, _ *model.Player, _ *model.QueuedAction, eventCtx *model.EventContext) {
	if !isCharacter(player, "assassin") || !hasAssassinStealthForm(player) || eventCtx == nil || eventCtx.AttackInfo == nil {
		return
	}
	ensurePlayerTokensMap(player)
	eventCtx.AttackInfo.SetInterceptTag(model.CombatInterceptUnrespondable)
	e.Log(fmt.Sprintf("[Skill] %s 处于[潜行]：本次主动攻击无法应战", player.Name))
}

func applyHolyLancerAttackGating(e *GameEngine, player *model.Player, _ *model.Player, _ *model.QueuedAction, eventCtx *model.EventContext) {
	if !e.isHolyLancer(player) || player.TurnState.UsedSkillCounts["holy_lancer_sky_spear_no_counter"] <= 0 || eventCtx == nil || eventCtx.AttackInfo == nil {
		return
	}
	eventCtx.AttackInfo.SetInterceptTag(model.CombatInterceptUnrespondable)
	player.TurnState.UsedSkillCounts["holy_lancer_sky_spear_no_counter"] = 0
}

func applyMagicSwordsmanAttackGating(e *GameEngine, player *model.Player, _ *model.Player, _ *model.QueuedAction, eventCtx *model.EventContext) {
	if !e.isMagicSwordsman(player) || player.TurnState.UsedSkillCounts["ms_yellow_spring_pending"] <= 0 || eventCtx == nil || eventCtx.AttackInfo == nil {
		return
	}
	eventCtx.AttackInfo.SetInterceptTag(model.CombatInterceptUnrespondable)
}

func applyDarkElementNoCounterRule(_ *GameEngine, _ *model.Player, _ *model.Player, currentAction *model.QueuedAction, eventCtx *model.EventContext) {
	if currentAction == nil || currentAction.Card == nil || currentAction.Card.Element != model.ElementDark || eventCtx == nil || eventCtx.AttackInfo == nil {
		return
	}
	eventCtx.AttackInfo.SetInterceptTag(model.CombatInterceptUnrespondable)
}

func applyBeastSamuraiAttackGating(e *GameEngine, player *model.Player, _ *model.Player, currentAction *model.QueuedAction, eventCtx *model.EventContext) {
	if !e.isBeastSamurai(player) || player.TurnState.UsedSkillCounts["bs_one_strike_armed"] <= 0 || eventCtx == nil || eventCtx.AttackInfo == nil {
		return
	}
	player.TurnState.UsedSkillCounts["bs_one_strike_armed"] = 0
	eventCtx.AttackInfo.SetInterceptTag(model.CombatInterceptIgnoreHolyShield)
	eventCtx.AttackInfo.SetInterceptTag(model.CombatInterceptIgnoreTargetHoly)
	if currentAction != nil && currentAction.Card != nil && strings.TrimSpace(currentAction.Card.Faction) == "技" {
		eventCtx.AttackInfo.SetInterceptTag(model.CombatInterceptForceHit)
	}
	e.Log(fmt.Sprintf("%s 的 [一击无念·下次攻击劫持] 生效：本次主动攻击无视圣盾、不可用圣光抵挡%s", player.Name, func() string {
		if currentAction != nil && currentAction.Card != nil && strings.TrimSpace(currentAction.Card.Faction) == "技" {
			return "，且技命格攻击强制命中"
		}
		return ""
	}()))
}
