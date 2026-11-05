// gameflow: 格斗家行动选择策略。

package fighter

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

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
		mod.SetActionRulePromptMessage(fmt.Sprintf("你处于【百式幻龙拳】状态：主动攻击 %s 可保持该状态；攻击其他目标会取消状态。", model.GetPlayerDisplayName(locked)))
	} else {
		mod.SetActionRulePromptMessage("你处于【百式幻龙拳】状态：主动攻击已锁定目标可保持该状态；攻击其他目标会取消状态。")
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
		// 只报错、不取消状态：这些入口本就不显示，误点（或过期的界面）不应丢掉形态。
		switch act.Type {
		case model.CmdMagic, model.CmdSkill:
			return fmt.Errorf("百式幻龙拳状态下不能执行法术行动，请选择主动攻击")
		case model.CmdBuy, model.CmdSynthesize, model.CmdExtract:
			return fmt.Errorf("百式幻龙拳状态下不能执行特殊行动，请选择主动攻击")
		default:
			return nil
		}
	})
	mod.SetOnAttackAccepted(func(rt engineplayer.ChoiceRuntime, player *model.Player, act model.PlayerAction) error {
		targetID := act.TargetID
		if targetID == "" && len(act.TargetIDs) > 0 {
			targetID = act.TargetIDs[0]
		}
		if rt.GetPlayers()[targetID] == nil {
			return fmt.Errorf("目标不存在")
		}
		locked := LockedTarget(rt, player)
		if locked == nil {
			ClearHundredDragon(rt, player, fmt.Sprintf("%s 的 [百式幻龙拳] 状态异常：未锁定目标，立即转正", player.Name))
			return fmt.Errorf("百式幻龙拳未锁定目标，状态已取消，请重新选择行动")
		}
		if locked.ID != targetID {
			ClearHundredDragon(rt, player, fmt.Sprintf("%s 攻击目标变化，取消 [百式幻龙拳] 并继续本次攻击", player.Name))
		}
		return nil
	})
}
