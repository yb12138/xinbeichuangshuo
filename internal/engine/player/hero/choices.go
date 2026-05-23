// gameflow: 英雄角色选择流。

package hero

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
	case "hero_roar_draw":
		return &model.Prompt{
			Type:       model.PromptConfirm,
			PlayerID:   playerID,
			Message:    "【怒吼】请选择摸牌数量：",
			ChoiceType: choiceType,
			Options: []model.PromptOption{
				{ID: "0", Label: "摸0张"},
				{ID: "1", Label: "摸1张"},
			},
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationNumeric, NumericBase: 0},
		}
	default:
		return nil
	}
}

func (choiceHandler) HandleChoice(rt engineplayer.ChoiceRuntime, _ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "hero_roar_draw":
		return true, handleHeroRoarDrawChoice(rt, selectionIndex, ctxData)
	default:
		return false, nil
	}
}

func handleHeroRoarDrawChoice(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	drawCount := 0
	switch selectionIndex {
	case 0:
		drawCount = 0
	case 1:
		drawCount = 1
	default:
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	if drawCount > 0 {
		rt.DrawCards(user.ID, drawCount)
	}
	rt.Log(fmt.Sprintf("%s 的 [怒吼] 结算：摸%d张牌", user.Name, drawCount))
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		rt.EnterActionExecutionStage()
	}
	return nil
}
