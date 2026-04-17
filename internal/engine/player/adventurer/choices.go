// gameflow: 冒险家角色选择流。

package adventurer

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
	case "adventurer_steal_sky_mode":
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【偷天换日】请选择效果：",
			Options: []model.PromptOption{
				{ID: "0", Label: "转移对方战绩区1红宝石到我方"},
				{ID: "1", Label: "将我方战绩区全部蓝水晶转换成红宝石"},
			},
			Min: 1,
			Max: 1,
		}
	case "adventurer_steal_sky_extra_action":
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【偷天换日】请选择额外行动类型：",
			Options: []model.PromptOption{
				{ID: "0", Label: "额外+1攻击行动"},
				{ID: "1", Label: "额外+1法术行动"},
			},
			Min: 1,
			Max: 1,
		}
	}
	return nil
}

func (choiceHandler) HandleChoice(rt engineplayer.ChoiceRuntime, _ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "adventurer_steal_sky_mode":
		return true, handleAdventurerStealSkyModeChoice(rt, selectionIndex, ctxData)
	case "adventurer_steal_sky_extra_action":
		return true, handleAdventurerStealSkyExtraActionChoice(rt, selectionIndex, ctxData)
	default:
		return false, nil
	}
}

func handleAdventurerStealSkyModeChoice(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	return fmt.Errorf("adventurer choice handler requires full engine access - temporarily disabled")
}

func handleAdventurerStealSkyExtraActionChoice(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.LookupPlayer(userID)
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	switch selectionIndex {
	case 0:
		model.AppendAttackAction(user, "偷天换日")
	case 1:
		model.AppendMagicAction(user, "偷天换日")
	default:
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	rt.PopInterrupt()
	return nil
}
