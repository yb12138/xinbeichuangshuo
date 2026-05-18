// gameflow: 祈祷师角色选择流。

package prayer_master

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

type choiceHandler struct{}

func NewChoiceHandler() engineplayer.ChoiceHandler {
	return choiceHandler{}
}

func (choiceHandler) BuildPrompt(_ engineplayer.ChoiceRuntime, choiceType, playerID string, _ *model.Player, _ map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "prayer_power_blessing_response":
		return &model.Prompt{
			Type:         model.PromptConfirm,
			PlayerID:     playerID,
			ChoiceType:   choiceType,
			Message:      "【威力赐福】是否移除该赐福，使本次攻击伤害+2？",
			Options:      []model.PromptOption{{ID: "0", Label: "是"}, {ID: "1", Label: "否"}},
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationBranchSelect, Layout: "overlay"},
		}
	case "prayer_swift_blessing_response":
		return &model.Prompt{
			Type:         model.PromptConfirm,
			PlayerID:     playerID,
			ChoiceType:   choiceType,
			Message:      "【迅捷赐福】是否移除该赐福，获得额外1次攻击行动？",
			Options:      []model.PromptOption{{ID: "0", Label: "是"}, {ID: "1", Label: "否"}},
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationBranchSelect, Layout: "overlay"},
		}
	}
	return nil
}

func (choiceHandler) HandleChoice(rt engineplayer.ChoiceRuntime, _ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "prayer_power_blessing_response":
		return true, handlePrayerPowerBlessingResponseChoice(rt, selectionIndex, ctxData)
	case "prayer_swift_blessing_response":
		return true, handlePrayerSwiftBlessingResponseChoice(rt, selectionIndex, ctxData)
	default:
		return false, nil
	}
}

func handlePrayerPowerBlessingResponseChoice(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	if selectionIndex == 0 {
		rt.RemoveFieldCard(user.ID, model.EffectPowerBlessing)
		rt.Log(fmt.Sprintf("%s 的 [威力赐福] 生效，本次攻击伤害+2", user.Name))
	}
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil && len(rt.GetPendingDamageQueue()) > 0 {
		rt.EnterDamageResolution(nil)
	}
	return nil
}

func handlePrayerSwiftBlessingResponseChoice(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	if selectionIndex == 0 {
		rt.RemoveFieldCard(user.ID, model.EffectSwiftBlessing)
		model.AppendAttackAction(user, "迅捷赐福")
		rt.Log(fmt.Sprintf("%s 的 [迅捷赐福] 生效，获得额外攻击行动", user.Name))
	}
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		rt.EnterExtraActionStage()
	}
	return nil
}
