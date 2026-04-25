// gameflow: 封印师角色选择流。

package sealer

import (
	"fmt"

	"starcup-engine/internal/engine/core/runtimeutil"
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

type choiceHandler struct{}

func NewChoiceHandler() engineplayer.ChoiceHandler {
	return choiceHandler{}
}

func (choiceHandler) BuildPrompt(rt engineplayer.ChoiceRuntime, choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	if choiceType != "five_elements_bind" && choiceType != "sealer_five_elements_bind_pick" {
		return nil
	}
	drawCount := runtimeutil.ToIntContextValue(data["draw_count"])
	return &model.Prompt{
		Type:     model.PromptConfirm,
		PlayerID: playerID,
		Message:  fmt.Sprintf("【五系束缚】%s，你需要做出选择：", player.Name),
		Options: []model.PromptOption{
			{ID: "0", Label: fmt.Sprintf("摸 %d 张牌取消束缚", drawCount)},
			{ID: "1", Label: "跳过行动阶段"},
		},
		Min: 1,
		Max: 1,
	}
}

func (choiceHandler) HandleChoice(rt engineplayer.ChoiceRuntime, _ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "five_elements_bind", "sealer_five_elements_bind_pick":
		return true, handleFiveElementsBindChoice(rt, selectionIndex, ctxData)
	default:
		return false, nil
	}
}

func handleFiveElementsBindChoice(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	drawCount := runtimeutil.ToIntContextValue(ctxData["draw_count"])
	targetPlayerID, _ := ctxData["player_id"].(string)
	player := rt.GetPlayers()[targetPlayerID]
	if player == nil {
		rt.PopInterrupt()
		return fmt.Errorf("五系束缚目标玩家不存在")
	}

	rt.RemoveFieldCard(player.ID, model.EffectFiveElementsBind)

	switch selectionIndex {
	case 0:
		rt.Log(fmt.Sprintf("[FiveElementsBind] %s 选择摸 %d 张牌", player.Name, drawCount))
		rt.DrawCards(player.ID, drawCount)

		rt.PopInterrupt()
		if rt.GetPendingInterrupt() == nil {
			rt.ApplyChoiceResumePoint(model.TurnStageActionStart)
		}
		return nil
	case 1:
		rt.Log(fmt.Sprintf("[FiveElementsBind] %s 选择放弃行动", player.Name))
		if player.TurnState.UsedSkillCounts == nil {
			player.TurnState.UsedSkillCounts = map[string]int{}
		}
		player.TurnState.UsedSkillCounts["arbiter_skip_forced_doomsday"] = 1

		rt.PopInterrupt()
		if rt.GetPendingInterrupt() == nil {
			rt.EnterTurnEndStage()
		}
		return nil
	default:
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
}
