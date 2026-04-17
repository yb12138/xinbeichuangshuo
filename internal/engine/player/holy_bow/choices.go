// gameflow: 圣弓手角色选择流。

package holy_bow

import (
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

type choiceHandler struct{}

func NewChoiceHandler() engineplayer.ChoiceHandler {
	return choiceHandler{}
}

func (choiceHandler) BuildPrompt(_ engineplayer.ChoiceRuntime, choiceType, playerID string, _ *model.Player, _ map[string]interface{}) *model.Prompt {
	return nil
}

func (choiceHandler) HandleChoice(_ engineplayer.ChoiceRuntime, _ string, _ int, _ map[string]interface{}) (bool, error) {
	return false, nil
}
