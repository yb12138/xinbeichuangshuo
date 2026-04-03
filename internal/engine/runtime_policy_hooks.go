package engine

import (
	"fmt"

	"starcup-engine/internal/engine/skills"
	"starcup-engine/internal/model"
)

type actionSelectionOptionPolicy func(e *GameEngine, player *model.Player, state *actionSelectionOptionState)
type actionSelectionValidationPolicy func(e *GameEngine, player *model.Player, act model.PlayerAction, result *actionSelectionValidationResult) error

type combatInteractionHook func(e *GameEngine, req *model.CombatRequest) bool

type responseSkillIDAugmenter func(sd *SkillDispatcher, skillIDs []string, ctx *model.Context) []string
type responseSkillIDNormalizer func(sd *SkillDispatcher, skillIDs []string, ctx *model.Context) []string

var beforeActionPhaseHooks = []playerPhaseHook{
	beforeActionPoisonHook,
	beforeActionFiveElementsBindHook,
	beforeActionWeakHook,
}

var actionSelectionOptionPolicies = []actionSelectionOptionPolicy{
	actionSelectionArbiterForcedDoomsdayOptionsPolicy,
	actionSelectionHeroTauntOptionsPolicy,
	actionSelectionFighterHundredDragonOptionsPolicy,
}

var actionSelectionValidationPolicies = []actionSelectionValidationPolicy{
	actionSelectionArbiterForcedDoomsdayValidationPolicy,
	actionSelectionHeroTauntValidationPolicy,
	actionSelectionFighterHundredDragonValidationPolicy,
}

var combatInteractionHooks = []combatInteractionHook{
	combatInteractionOnmyojiBindingInterruptHook,
	combatInteractionOnmyojiBindingCounterHook,
	combatInteractionOnmyojiYinYangInterruptHook,
}

var responseSkillIDAugmenters = []responseSkillIDAugmenter{
	augmentBeastSamuraiResponseSkillIDs,
}

var responseSkillIDNormalizers = []responseSkillIDNormalizer{
	normalizeFighterResponseSkillIDs,
}

func turnBeforeStartButterflyDancerWitherExpiryHook(e *GameEngine, player *model.Player) bool {
	if !e.isButterflyDancer(player) {
		return false
	}
	ensurePlayerTokensMap(player)
	if player.Tokens["bt_wither_active"] <= 0 {
		return false
	}
	player.Tokens["bt_wither_active"] = 0
	e.Log(fmt.Sprintf("%s 的 [凋零] 效果到期：对方士气下限保护已解除", player.Name))
	return false
}

func turnStartBloodPriestessBleedHook(e *GameEngine, player *model.Player) bool {
	if !e.isBloodPriestess(player) {
		return false
	}
	if !hasBloodPriestessBleedingForm(player) || player.TurnState.UsedSkillCounts["bp_bleed_tick"] > 0 {
		return false
	}
	player.TurnState.UsedSkillCounts["bp_bleed_tick"] = 1
	e.Log(fmt.Sprintf("%s 的 [流血] 生效：回合开始对自己造成1点法术伤害", player.Name))
	e.AddPendingDamage(model.PendingDamage{
		SourceID:   player.ID,
		TargetID:   player.ID,
		Damage:     1,
		DamageType: "magic",
	})
	e.enterDamageResolution(model.TurnStageTurnStart)
	return true
}

func beforeActionPoisonHook(e *GameEngine, player *model.Player) bool {
	if player == nil {
		return false
	}
	for _, fc := range player.Field {
		if fc == nil || fc.Mode != model.FieldEffect || fc.Trigger != model.EffectTriggerOnBeforeAction || fc.Effect != model.EffectPoison {
			continue
		}
		allowCrimsonFaithHeal := fc.SourceID != "" && fc.SourceID == player.ID
		e.AddPendingDamage(model.PendingDamage{
			SourceID:              fc.SourceID,
			TargetID:              player.ID,
			Damage:                1,
			DamageType:            "poison",
			AllowCrimsonFaithHeal: allowCrimsonFaithHeal,
		})
		player.RemoveFieldCard(fc)
		e.State.DiscardPile = append(e.State.DiscardPile, fc.Card)
		e.Log(fmt.Sprintf("[Effect] %s 受到中毒伤害", player.Name))
		e.Log(fmt.Sprintf("[Field] %s 面前的【%s】触发效果并被弃置", player.Name, fc.Card.Name))
		e.enterDamageResolution(model.TurnStageBeforeAction)
		return true
	}
	return false
}

func beforeActionFiveElementsBindHook(e *GameEngine, player *model.Player) bool {
	if player == nil {
		return false
	}
	for _, fc := range player.Field {
		if fc == nil || fc.Mode != model.FieldEffect || fc.Trigger != model.EffectTriggerOnBeforeAction || fc.Effect != model.EffectFiveElementsBind {
			continue
		}
		sealCount := 0
		for _, fieldPlayer := range e.GetAllPlayers() {
			if fieldPlayer == nil {
				continue
			}
			for _, fieldCard := range fieldPlayer.Field {
				if fieldCard == nil || fieldCard.Mode != model.FieldEffect || !model.IsElementalSealEffect(fieldCard.Effect) {
					continue
				}
				sealCount++
				if sealCount >= 2 {
					break
				}
			}
			if sealCount >= 2 {
				break
			}
		}
		e.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: player.ID,
			Context: map[string]interface{}{
				"choice_type": "five_elements_bind",
				"player_id":   player.ID,
				"draw_count":  2 + sealCount,
			},
		})
		e.Log(fmt.Sprintf("[Buff] %s 触发五系束缚判定，等待玩家选择...", player.Name))
		return true
	}
	return false
}

func beforeActionWeakHook(e *GameEngine, player *model.Player) bool {
	if player == nil || !player.HasFieldEffect(model.EffectWeak) {
		return false
	}
	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: player.ID,
		Context: map[string]interface{}{
			"choice_type": "weak",
		},
	})
	e.Log(fmt.Sprintf("[Buff] %s 触发虚弱判定，等待玩家选择...", player.Name))
	return true
}

func turnStartValkyrieMilitaryGloryHook(e *GameEngine, player *model.Player) bool {
	if e == nil || player == nil || !isCharacter(player, "valkyrie") {
		return false
	}
	ensurePlayerTokensMap(player)
	if player.Tokens["valkyrie_spirit"] <= 0 || player.TurnState.UsedSkillCounts["valkyrie_military_glory"] > 0 {
		return false
	}
	ctx := e.buildTimedContext(player, nil, model.TriggerOnTurnStart, model.TimingOnTurnStart, &model.EventContext{
		Type:     model.EventTurnStart,
		SourceID: player.ID,
	})
	handler := skills.GetHandler("valkyrie_military_glory")
	if handler == nil || !handler.CanUse(ctx) {
		return false
	}
	player.TurnState.UsedSkillCounts["valkyrie_military_glory"] = 1
	if err := handler.Execute(ctx); err != nil {
		e.Log(fmt.Sprintf("[Skill Error] 军威神光执行失败: %v", err))
		return false
	}
	e.recordSkillUsage(player.ID, "军威神光", model.SkillTypeStartup)
	return e.State.PendingInterrupt != nil
}

func startupHeroExhaustionReleaseHook(e *GameEngine, player *model.Player) bool {
	if e == nil || player == nil || !e.isHero(player) || player.TurnState.HasUsedTriggerSkill || !hasHeroExhaustionForm(player) {
		return false
	}
	ensurePlayerTokensMap(player)
	if player.Tokens["hero_exhaustion_release_pending"] <= 0 {
		return false
	}
	beforePoses := e.snapshotPlayerPoses()
	leaveHeroExhaustionForm(player)
	player.Tokens["hero_exhaustion_release_pending"] = 0
	e.Log(fmt.Sprintf("%s 的 [精疲力竭] 结束：转正，手牌上限恢复，并对自己造成3点法术伤害", player.Name))
	e.dispatchOrientationChanges(beforePoses)
	e.AddPendingDamage(model.PendingDamage{
		SourceID:   player.ID,
		TargetID:   player.ID,
		Damage:     3,
		DamageType: "magic",
	})
	return true
}

func startupArbiterForcedDoomsdayHook(e *GameEngine, player *model.Player) bool {
	if e == nil || player == nil || !isCharacter(player, "arbiter") {
		return false
	}
	ensurePlayerTokensMap(player)
	if player.Tokens["judgment"] < 4 || player.TurnState.UsedSkillCounts["arbiter_skip_forced_doomsday"] != 0 || player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_done_turn"] != 0 {
		player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] = 0
		return false
	}
	if len(e.campEnemyIDs(player.Camp)) == 0 {
		player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] = 0
		return false
	}
	if player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] == 0 {
		e.Log(fmt.Sprintf("%s 的审判已达上限：本行动阶段必须发动 [末日审判]", player.Name))
	}
	player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] = 1
	return false
}

func startupHeroTauntHook(e *GameEngine, player *model.Player) bool {
	if e == nil || player == nil {
		return false
	}
	player.TurnState.UsedSkillCounts["hero_taunt_active_turn"] = 0
	if player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] > 0 {
		return false
	}
	src := activeHeroTauntSource(e, player)
	if src == nil {
		return false
	}
	player.TurnState.UsedSkillCounts["hero_taunt_active_turn"] = 1
	e.Log(fmt.Sprintf("[Taunt] %s 在本行动阶段受到 %s 的【挑衅】影响：必须且只能主动攻击该勇者，或选择跳过行动并移除此牌", player.Name, model.GetPlayerDisplayName(src)))
	return false
}

func activeHeroTauntSource(e *GameEngine, player *model.Player) *model.Player {
	if e == nil || player == nil {
		return nil
	}
	tauntCard := getHeroTauntCard(player)
	if tauntCard == nil {
		clearHeroTauntRestriction(e, player)
		return nil
	}
	src := e.State.Players[tauntCard.SourceID]
	if src == nil || src.Camp == player.Camp {
		clearHeroTauntRestriction(e, player)
		return nil
	}
	return src
}

func clearHeroTauntRestriction(e *GameEngine, player *model.Player) {
	consumeHeroTauntRestriction(e, player)
}

func consumeHeroTauntRestriction(e *GameEngine, player *model.Player) {
	if e == nil || player == nil {
		return
	}
	player.TurnState.UsedSkillCounts["hero_taunt_active_turn"] = 0
	e.RemoveFieldCard(player.ID, model.EffectHeroTaunt)
}

func hasPlayableAttackCard(player *model.Player) bool {
	if player == nil {
		return false
	}
	for idx := 0; idx < playableCardCount(player); idx++ {
		card, _, _, ok := getPlayableCardByIndex(player, idx)
		if ok && card.Type == model.CardTypeAttack {
			return true
		}
	}
	return false
}

func actionSelectionArbiterForcedDoomsdayOptionsPolicy(e *GameEngine, player *model.Player, state *actionSelectionOptionState) {
	if e == nil || player == nil || state == nil {
		return
	}
	if player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] <= 0 {
		return
	}
	skillDef := findCharacterSkill(player.Character, "arbiter_doomsday")
	if skillDef == nil || !e.isActionSkillUsableForExtraMagic(player, *skillDef) {
		return
	}
	state.hasArbiterForcedDoomsday = true
	state.canMagicAction = false
	state.canMagicSkillAction = true
	state.promptChoiceType = "arbiter_forced_doomsday"
	state.promptSkillID = "arbiter_doomsday"
}

func actionSelectionHeroTauntOptionsPolicy(e *GameEngine, player *model.Player, state *actionSelectionOptionState) {
	if player == nil || state == nil {
		return
	}
	if player.TurnState.UsedSkillCounts["hero_taunt_active_turn"] <= 0 || player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] > 0 {
		return
	}
	src := activeHeroTauntSource(e, player)
	if src == nil {
		return
	}
	state.tauntSourceID = src.ID
	state.tauntSourceName = model.GetPlayerDisplayName(src)
	state.hasHeroTaunt = true
	state.tauntRequiresSkip = !hasPlayableAttackCard(player)
}

func actionSelectionFighterHundredDragonOptionsPolicy(e *GameEngine, player *model.Player, state *actionSelectionOptionState) {
	if e == nil || player == nil || state == nil {
		return
	}
	state.hasFighterHundredDragon = e.isFighter(player) && hasFighterHundredDragonForm(player)
}

func actionSelectionArbiterForcedDoomsdayValidationPolicy(e *GameEngine, player *model.Player, act model.PlayerAction, _ *actionSelectionValidationResult) error {
	if e == nil || player == nil || player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] <= 0 {
		return nil
	}
	if act.Type != model.CmdSkill {
		return fmt.Errorf("审判已达上限：本行动阶段必须发动 [末日审判]")
	}
	if act.SkillID != "arbiter_doomsday" {
		return fmt.Errorf("审判已达上限：本行动阶段只能发动 [末日审判]")
	}
	return nil
}

func actionSelectionHeroTauntValidationPolicy(e *GameEngine, player *model.Player, act model.PlayerAction, result *actionSelectionValidationResult) error {
	if e == nil || player == nil {
		return nil
	}
	if player.TurnState.UsedSkillCounts["hero_taunt_active_turn"] <= 0 || player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] > 0 {
		return nil
	}
	src := activeHeroTauntSource(e, player)
	if src == nil {
		return nil
	}
	sourceName := model.GetPlayerDisplayName(src)
	if act.Type == model.CmdCannotAct {
		e.Log(fmt.Sprintf("[Taunt] %s 选择跳过本次行动阶段，并移除来自 %s 的【挑衅】", player.Name, sourceName))
		clearHeroTauntRestriction(e, player)
		e.enterTurnEndStage()
		if result != nil {
			result.handled = true
		}
		return nil
	}
	if act.Type != model.CmdAttack {
		return fmt.Errorf("你受到【挑衅】影响：本次行动阶段必须且只能主动攻击 %s，或选择跳过行动", sourceName)
	}
	targetID := act.TargetID
	if targetID == "" && len(act.TargetIDs) > 0 {
		targetID = act.TargetIDs[0]
	}
	if targetID == "" {
		return fmt.Errorf("你受到【挑衅】影响：攻击必须指定 %s", sourceName)
	}
	if e.State.Players[targetID] == nil {
		return fmt.Errorf("目标不存在")
	}
	if targetID != src.ID {
		return fmt.Errorf("你受到【挑衅】影响：本次行动阶段只能主动攻击 %s", sourceName)
	}
	if result != nil {
		result.consumeHeroTauntOnAttack = true
	}
	return nil
}

func actionSelectionFighterHundredDragonValidationPolicy(e *GameEngine, player *model.Player, act model.PlayerAction, _ *actionSelectionValidationResult) error {
	if e == nil || player == nil || !e.isFighter(player) || !hasFighterHundredDragonForm(player) {
		return nil
	}
	switch act.Type {
	case model.CmdAttack:
		targetID := act.TargetID
		if targetID == "" && len(act.TargetIDs) > 0 {
			targetID = act.TargetIDs[0]
		}
		if targetID == "" {
			return fmt.Errorf("百式幻龙拳状态下攻击必须指定目标")
		}
		targetPlayer := e.State.Players[targetID]
		if targetPlayer == nil {
			return fmt.Errorf("目标不存在")
		}
		if targetPlayer.Camp == player.Camp {
			return fmt.Errorf("攻击目标必须是敌方角色")
		}
		targetOrder := e.playerOrderPosition(targetID)
		if targetOrder == 0 {
			return fmt.Errorf("目标不存在")
		}
		lockedOrder := player.Tokens["fighter_hundred_dragon_target_order"]
		if lockedOrder == 0 {
			e.clearFighterHundredDragon(player, fmt.Sprintf("%s 的 [百式幻龙拳] 状态异常：未锁定目标，立即转正", player.Name))
			return fmt.Errorf("百式幻龙拳未锁定目标，状态已取消，请重新选择行动")
		}
		if lockedOrder != targetOrder {
			e.clearFighterHundredDragon(player, fmt.Sprintf("%s 攻击目标变化，取消 [百式幻龙拳] 并继续本次攻击", player.Name))
		}
	case model.CmdMagic, model.CmdSkill:
		e.clearFighterHundredDragon(player, fmt.Sprintf("%s 尝试执行法术行动，取消 [百式幻龙拳] 并转正", player.Name))
		return fmt.Errorf("百式幻龙拳状态下不能执行法术行动；状态已取消，请重新选择行动")
	case model.CmdBuy, model.CmdSynthesize, model.CmdExtract:
		e.clearFighterHundredDragon(player, fmt.Sprintf("%s 尝试执行特殊行动，取消 [百式幻龙拳] 并转正", player.Name))
		return fmt.Errorf("百式幻龙拳状态下不能执行特殊行动；状态已取消，请重新选择行动")
	}
	return nil
}

func combatInteractionOnmyojiBindingInterruptHook(e *GameEngine, req *model.CombatRequest) bool {
	return e != nil && req != nil && e.tryStartOnmyojiBindingInterrupt(req)
}

func combatInteractionOnmyojiBindingCounterHook(e *GameEngine, req *model.CombatRequest) bool {
	return e != nil && req != nil && e.executeOnmyojiBindingCounter(req)
}

func combatInteractionOnmyojiYinYangInterruptHook(e *GameEngine, req *model.CombatRequest) bool {
	return e != nil && req != nil && e.tryStartOnmyojiYinYangInterrupt(req)
}

func augmentBeastSamuraiResponseSkillIDs(sd *SkillDispatcher, skillIDs []string, ctx *model.Context) []string {
	if sd == nil || sd.engine == nil || ctx == nil || ctx.Trigger != model.TriggerOnPhaseEnd || ctx.TriggerCtx == nil || ctx.TriggerCtx.ActionType != model.ActionAttack || ctx.User == nil {
		return skillIDs
	}
	if !sd.engine.isBeastSamurai(ctx.User) || containsSkillID(skillIDs, "bs_one_strike_no_thought") || sd.engine.beastSamuraiZanshin(ctx.User) < beastSamuraiZanshinCapEngine {
		return skillIDs
	}
	skillDef := findCharacterSkill(ctx.User.Character, "bs_one_strike_no_thought")
	if skillDef == nil {
		return skillIDs
	}
	handler := skills.GetHandler(skillDef.LogicHandler)
	if handler == nil || !handler.CanUse(ctx) {
		return skillIDs
	}
	return append(skillIDs, skillDef.ID)
}

func normalizeFighterResponseSkillIDs(sd *SkillDispatcher, skillIDs []string, ctx *model.Context) []string {
	if sd == nil || sd.engine == nil || len(skillIDs) <= 1 || ctx == nil || ctx.User == nil {
		return skillIDs
	}
	if ctx.Trigger != model.TriggerOnAttackStart || !sd.engine.isFighter(ctx.User) || ctx.TriggerCtx == nil || ctx.TriggerCtx.AttackInfo == nil || ctx.TriggerCtx.AttackInfo.CounterInitiator != "" {
		return skillIDs
	}
	hasCharge := false
	hasBurst := false
	for _, sid := range skillIDs {
		if sid == "fighter_charge_strike" {
			hasCharge = true
		} else if sid == "fighter_burst_crash" {
			hasBurst = true
		}
	}
	if hasCharge && hasBurst {
		return []string{"fighter_charge_strike"}
	}
	return skillIDs
}
