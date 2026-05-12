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

func (choiceHandler) BuildPrompt(_ engineplayer.ChoiceRuntime, choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "five_elements_bind":
		return buildFiveElementsBindPrompt(playerID, data)
	default:
		return nil
	}
}

func buildFiveElementsBindPrompt(playerID string, data map[string]interface{}) *model.Prompt {
	drawCount := 2
	if dc := runtimeutil.ToIntContextValue(data["draw_count"]); dc > 0 {
		drawCount = dc
	}
	return &model.Prompt{
		Type:     model.PromptConfirm,
		PlayerID: playerID,
		Message:  "【五系束缚】请选择：",
		Options: []model.PromptOption{
			{ID: "0", Label: fmt.Sprintf("摸%d张牌", drawCount)},
			{ID: "1", Label: "放弃行动"},
		},
		Min:          1,
		Max:          1,
		Presentation: &model.PromptPresentation{Kind: model.PresentationBranchSelect, Layout: "overlay"},
	}
}

func (choiceHandler) HandleChoice(rt engineplayer.ChoiceRuntime, _ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "five_elements_bind":
		return true, handleFiveElementsBindChoice(rt, ctxData, selectionIndex)
	default:
		return false, nil
	}
}

func handleFiveElementsBindChoice(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	playerID, _ := ctxData["player_id"].(string)
	if playerID == "" {
		playerID, _ = ctxData["user_id"].(string)
	}
	player := rt.GetPlayers()[playerID]
	if player == nil {
		return fmt.Errorf("玩家不存在")
	}
	switch selectionIndex {
	case 0:
		drawCount := 2
		if dc, ok := ctxData["draw_count"].(int); ok && dc > 0 {
			drawCount = dc
		}
		rt.RemoveFieldCard(player.ID, model.EffectFiveElementsBind)
		rt.Log(fmt.Sprintf("[FiveElementsBind] %s 选择摸 %d 张牌", player.Name, drawCount))
		rt.DrawCards(player.ID, drawCount)

		rt.PopInterrupt()
		if rt.GetPendingInterrupt() == nil {
			rt.ApplyChoiceResumePoint(model.TurnStageActionStart)
		}
		return nil
	case 1:
		rt.RemoveFieldCard(player.ID, model.EffectFiveElementsBind)
		rt.Log(fmt.Sprintf("[FiveElementsBind] %s 选择放弃行动", player.Name))
		player.TurnState.ActionPhaseSkippedThisTurn = true

		rt.PopInterrupt()
		if rt.GetPendingInterrupt() == nil {
			rt.EnterTurnEndStage()
		}
		return nil
	default:
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
}
