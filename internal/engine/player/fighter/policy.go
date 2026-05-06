// gameflow: 格斗家行动选择策略。

package fighter

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// playerOrderPosition returns the 1-based position of playerID in the player order.
func playerOrderPosition(rt engineplayer.ChoiceRuntime, playerID string) int {
	order := rt.GetPlayerOrder()
	for i, pid := range order {
		if pid == playerID {
			return i + 1
		}
	}
	return 0
}

// HundredDragonOptionPolicy 百式幻龙拳行动选项策略。
func HundredDragonOptionPolicy(rt engineplayer.ChoiceRuntime, player *model.Player, mod engineplayer.ActionSelectionModifier) {
	if rt == nil || player == nil {
		return
	}
	if !engineplayer.IsCharacter(player, "fighter") || !InHundredDragonForm(player) {
		return
	}
	mod.SetActionRule("force_attack", "fighter_hundred_dragon", 20)
	if locked := LockedTarget(rt, player); locked != nil {
		mod.SetActionRulePromptMessage(fmt.Sprintf("你处于【百式幻龙拳】状态：本行动阶段只能主动攻击 %s；若本行动阶段结束仍处于该形态，则自动转正。", model.GetPlayerDisplayName(locked)))
	} else {
		mod.SetActionRulePromptMessage("你处于【百式幻龙拳】状态：本行动阶段只能主动攻击已锁定目标；若本行动阶段结束仍处于该形态，则自动转正。")
	}
}

// HundredDragonValidationPolicy 百式幻龙拳行动校验策略。
func HundredDragonValidationPolicy(rt engineplayer.ChoiceRuntime, player *model.Player, mod engineplayer.ActionSelectionValidationModifier) {
	if rt == nil || player == nil || !engineplayer.IsCharacter(player, "fighter") || !InHundredDragonForm(player) {
		return
	}
	mod.SetActionRule("force_attack", "fighter_hundred_dragon", 20)
	mod.SetForceAttackOnlyMessage("百式幻龙拳状态下只能主动攻击")
	mod.SetOnNonAttackChosen(func(rt engineplayer.ChoiceRuntime, player *model.Player, act model.PlayerAction) error {
		switch act.Type {
		case model.CmdMagic, model.CmdSkill:
			ClearHundredDragon(rt, player, fmt.Sprintf("%s 尝试执行法术行动，取消 [百式幻龙拳] 并转正", player.Name))
			return fmt.Errorf("百式幻龙拳状态下不能执行法术行动；状态已取消，请重新选择行动")
		case model.CmdBuy, model.CmdSynthesize, model.CmdExtract:
			ClearHundredDragon(rt, player, fmt.Sprintf("%s 尝试执行特殊行动，取消 [百式幻龙拳] 并转正", player.Name))
			return fmt.Errorf("百式幻龙拳状态下不能执行特殊行动；状态已取消，请重新选择行动")
		default:
			return nil
		}
	})
	mod.SetOnAttackAccepted(func(rt engineplayer.ChoiceRuntime, player *model.Player, act model.PlayerAction) error {
		targetID := act.TargetID
		if targetID == "" && len(act.TargetIDs) > 0 {
			targetID = act.TargetIDs[0]
		}
		targetOrder := playerOrderPosition(rt, targetID)
		if targetOrder == 0 {
			return fmt.Errorf("目标不存在")
		}
		lockedOrder := engineplayer.GetSkillFlowState(player, "fighter_hundred_dragon_target_order")
		if lockedOrder == 0 {
			ClearHundredDragon(rt, player, fmt.Sprintf("%s 的 [百式幻龙拳] 状态异常：未锁定目标，立即转正", player.Name))
			return fmt.Errorf("百式幻龙拳未锁定目标，状态已取消，请重新选择行动")
		}
		if lockedOrder != targetOrder {
			ClearHundredDragon(rt, player, fmt.Sprintf("%s 攻击目标变化，取消 [百式幻龙拳] 并继续本次攻击", player.Name))
		}
		return nil
	})
}
