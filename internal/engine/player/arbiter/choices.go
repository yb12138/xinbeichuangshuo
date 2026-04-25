// gameflow: 仲裁者角色选择流。

package arbiter

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

type choiceHandler struct{}

func NewChoiceHandler() engineplayer.ChoiceHandler {
	return choiceHandler{}
}

func (choiceHandler) BuildPrompt(rt engineplayer.ChoiceRuntime, choiceType, playerID string, _ *model.Player, _ map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "arbiter_balance_mode":
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【判决天平】请选择一个分支：",
			Options: []model.PromptOption{
				{ID: "0", Label: "弃掉当前手上的所有手牌"},
				{ID: "1", Label: "将手牌补到上限，并我方战绩区+1红宝石"},
			},
			Min: 1,
			Max: 1,
		}
	default:
		return nil
	}
}

func (choiceHandler) HandleChoice(rt engineplayer.ChoiceRuntime, playerID string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "arbiter_balance_mode":
		return true, handleArbiterBalanceChoice(rt, playerID, selectionIndex, ctxData)
	default:
		return false, nil
	}
}

func handleArbiterBalanceChoice(rt engineplayer.ChoiceRuntime, playerID string, selectionIndex int, ctxData map[string]interface{}) error {
	user := rt.GetPlayers()[playerID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	switch selectionIndex {
	case 0:
		// 分支1：弃掉所有手牌
		rt.AppendToDiscard(user.Hand)
		user.Hand = nil
		rt.Log(fmt.Sprintf("%s 选择判决天平分支1：弃掉所有手牌", user.Name))
	case 1:
		// 分支2：补牌到上限并我方战绩区+1红宝石
		maxHand := user.MaxHand
		if len(user.Hand) < maxHand {
			rt.DrawCards(playerID, maxHand-len(user.Hand))
		}
		rt.ModifyGem(string(user.Camp), 1)
		rt.Log(fmt.Sprintf("%s 选择判决天平分支2：补牌到上限并我方战绩区+1红宝石", user.Name))
	default:
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	rt.PopInterrupt()
	return nil
}
