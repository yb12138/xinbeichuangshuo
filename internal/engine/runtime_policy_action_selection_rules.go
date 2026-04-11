// gameflow: 行动选择阶段可选项过滤规则。

package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

func actionSelectionArbiterForcedDoomsdayOptionsPolicy(e *GameEngine, player *model.Player, state *actionSelectionState) {
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
	state.setActionRule(actionSelectionRuleForceSkillAsMagic, "arbiter_forced_doomsday", 30)
	state.canMagicAction = false
	state.canMagicSkillAction = true
	state.promptChoiceType = "arbiter_forced_doomsday"
	state.promptSkillID = "arbiter_doomsday"
	state.actionRulePromptMessage = "你的审判已达上限：本行动阶段必须发动【末日审判】。"
}

func actionSelectionHeroTauntOptionsPolicy(e *GameEngine, player *model.Player, state *actionSelectionState) {
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
	state.setActionRule(actionSelectionRuleForceAttackOrSkip, "hero_taunt", 10)
	state.constrainedTargetID = src.ID
	state.constrainedTargetName = model.GetPlayerDisplayName(src)
	state.ruleRequiresSkipOnly = !hasPlayableAttackCard(player)
}

func actionSelectionFighterHundredDragonOptionsPolicy(e *GameEngine, player *model.Player, state *actionSelectionState) {
	if e == nil || player == nil || state == nil {
		return
	}
	if !e.isFighter(player) || !hasFighterHundredDragonForm(player) {
		return
	}
	state.setActionRule(actionSelectionRuleForceAttack, "fighter_hundred_dragon", 20)
	if locked := e.fighterLockedTarget(player); locked != nil {
		state.actionRulePromptMessage = fmt.Sprintf("你处于【百式幻龙拳】状态：本行动阶段只能主动攻击 %s；若本行动阶段结束仍处于该形态，则自动转正。", model.GetPlayerDisplayName(locked))
	} else {
		state.actionRulePromptMessage = "你处于【百式幻龙拳】状态：本行动阶段只能主动攻击已锁定目标；若本行动阶段结束仍处于该形态，则自动转正。"
	}
}

func actionSelectionArbiterForcedDoomsdayValidationPolicy(e *GameEngine, player *model.Player, state *actionSelectionState) {
	if e == nil || player == nil || player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] <= 0 {
		return
	}
	state.setActionRule(actionSelectionRuleForceSkillAsMagic, "arbiter_forced_doomsday", 30)
	state.requiredSkillID = "arbiter_doomsday"
	state.forceSkillMustUseMessage = "审判已达上限：本行动阶段必须发动 [末日审判]"
	state.forceSkillOnlyMessage = "审判已达上限：本行动阶段只能发动 [末日审判]"
}

func actionSelectionHeroTauntValidationPolicy(e *GameEngine, player *model.Player, state *actionSelectionState) {
	if e == nil || player == nil {
		return
	}
	if player.TurnState.UsedSkillCounts["hero_taunt_active_turn"] <= 0 || player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] > 0 {
		return
	}
	src := activeHeroTauntSource(e, player)
	if src == nil {
		return
	}
	sourceName := model.GetPlayerDisplayName(src)
	state.setActionRule(actionSelectionRuleForceAttackOrSkip, "hero_taunt", 10)
	state.constrainedTargetID = src.ID
	state.constrainedTargetName = sourceName
	state.ruleRequiresSkipOnly = !hasPlayableAttackCard(player)
	state.forceAttackOnlyMessage = fmt.Sprintf("你受到【挑衅】影响：本次行动阶段必须且只能主动攻击 %s，或选择跳过行动", sourceName)
	state.onSkipChosen = func(e *GameEngine, player *model.Player, result *actionSelectionValidationResult) (bool, error) {
		e.Log(fmt.Sprintf("[Taunt] %s 选择跳过本次行动阶段，并移除来自 %s 的【挑衅】", player.Name, sourceName))
		clearHeroTauntRestriction(e, player)
		e.enterTurnEndStage()
		if result != nil {
			result.handled = true
		}
		return true, nil
	}
	state.onAttackAccepted = func(_ *GameEngine, _ *model.Player, _ model.PlayerAction, result *actionSelectionValidationResult) error {
		if result != nil {
			result.consumeHeroTauntOnAttack = true
		}
		return nil
	}
}

func actionSelectionFighterHundredDragonValidationPolicy(e *GameEngine, player *model.Player, state *actionSelectionState) {
	if e == nil || player == nil || !e.isFighter(player) || !hasFighterHundredDragonForm(player) {
		return
	}
	state.setActionRule(actionSelectionRuleForceAttack, "fighter_hundred_dragon", 20)
	state.forceAttackOnlyMessage = "百式幻龙拳状态下只能主动攻击"
	state.onNonAttackChosen = func(e *GameEngine, player *model.Player, act model.PlayerAction, _ *actionSelectionValidationResult) error {
		switch act.Type {
		case model.CmdMagic, model.CmdSkill:
			e.clearFighterHundredDragon(player, fmt.Sprintf("%s 尝试执行法术行动，取消 [百式幻龙拳] 并转正", player.Name))
			return fmt.Errorf("百式幻龙拳状态下不能执行法术行动；状态已取消，请重新选择行动")
		case model.CmdBuy, model.CmdSynthesize, model.CmdExtract:
			e.clearFighterHundredDragon(player, fmt.Sprintf("%s 尝试执行特殊行动，取消 [百式幻龙拳] 并转正", player.Name))
			return fmt.Errorf("百式幻龙拳状态下不能执行特殊行动；状态已取消，请重新选择行动")
		default:
			return nil
		}
	}
	state.onAttackAccepted = func(e *GameEngine, player *model.Player, act model.PlayerAction, _ *actionSelectionValidationResult) error {
		targetID := act.TargetID
		if targetID == "" && len(act.TargetIDs) > 0 {
			targetID = act.TargetIDs[0]
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
		return nil
	}
}
