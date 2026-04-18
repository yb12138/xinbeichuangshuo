// gameflow: 神圣弓手：射击与圣箭相关流程。

package engine

import (
	"starcup-engine/internal/model"
)

func holyBowChoicePlayerOptions(e *GameEngine, playerIDs []string) []model.PromptOption {
	options := make([]model.PromptOption, 0, len(playerIDs))
	for _, playerID := range playerIDs {
		if player := e.State.Players[playerID]; player != nil {
			options = append(options, model.PromptOption{ID: playerID, Label: player.Name})
		}
	}
	return options
}
