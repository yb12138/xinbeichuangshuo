// gameflow: 魔枪士角色选择流。

package magic_lancer

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
	case "ml_black_spear_x":
		maxX := runtimeutil.ToIntContextValue(data["max_x"])
		if maxX < 1 {
			maxX = 1
		}
		options := make([]model.PromptOption, 0, maxX)
		for x := 1; x <= maxX; x++ {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", x), Label: fmt.Sprintf("X=%d（消耗%d蓝水晶，伤害额外+%d）", x, x, x+2)})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【漆黑之枪】请选择X值：", Options: options, Min: 1, Max: 1}
	}
	return nil
}

func (choiceHandler) HandleChoice(rt engineplayer.ChoiceRuntime, _ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "ml_black_spear_x":
		return true, handleMagicLancerBlackSpearXChoice(rt, selectionIndex, ctxData)
	default:
		return false, nil
	}
}

func handleMagicLancerBlackSpearXChoice(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	return fmt.Errorf("magic_lancer choice handler requires full engine access - temporarily disabled")
}
