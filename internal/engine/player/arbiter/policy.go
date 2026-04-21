// gameflow: 仲裁者行动选择策略。

package arbiter

import (
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// ForcedDoomsdayOptionPolicy 末日审判强制发动选项策略。
func ForcedDoomsdayOptionPolicy(rt engineplayer.ChoiceRuntime, player *model.Player, mod engineplayer.ActionSelectionModifier) {
	if rt == nil || player == nil {
		return
	}
	if player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] <= 0 {
		return
	}
	skillDef := FindCharacterSkill(player.Character, "arbiter_doomsday")
	if skillDef == nil || !rt.IsActionSkillUsableForExtraMagic(player, *skillDef) {
		return
	}
	mod.SetActionRule("force_skill_magic", "arbiter_forced_doomsday", 30)
	mod.SetCanMagicAction(false)
	mod.SetCanMagicSkillAction(true)
	mod.SetPromptChoiceType("arbiter_forced_doomsday")
	mod.SetPromptSkillID("arbiter_doomsday")
	mod.SetActionRulePromptMessage("你的审判已达上限：本行动阶段必须发动【末日审判】。")
}

// ForcedDoomsdayValidationPolicy 末日审判强制发动校验策略。
func ForcedDoomsdayValidationPolicy(rt engineplayer.ChoiceRuntime, player *model.Player, mod engineplayer.ActionSelectionValidationModifier) {
	if rt == nil || player == nil || player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] <= 0 {
		return
	}
	mod.SetActionRule("force_skill_magic", "arbiter_forced_doomsday", 30)
	mod.SetRequiredSkillID("arbiter_doomsday")
	mod.SetForceSkillMustUseMessage("审判已达上限：本行动阶段必须发动 [末日审判]")
	mod.SetForceSkillOnlyMessage("审判已达上限：本行动阶段只能发动 [末日审判]")
}

// FindCharacterSkill 在角色技能列表中查找指定 ID 的技能定义。
func FindCharacterSkill(character *model.Character, skillID string) *model.SkillDefinition {
	if character == nil {
		return nil
	}
	for i := range character.Skills {
		if character.Skills[i].ID == skillID {
			return &character.Skills[i]
		}
	}
	return nil
}
