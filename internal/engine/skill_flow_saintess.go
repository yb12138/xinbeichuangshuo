// gameflow: 圣女技能流。

package engine

import (
	"fmt"

	"starcup-engine/internal/engine/core/runtimeutil"
	"starcup-engine/internal/model"
)

func (e *GameEngine) buildSaintessChoicePrompt(choiceType, playerID string, _ *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "frost_prayer_target":
		options := buildPromptOptionsForPlayerIDs(e.State.Players, runtimeutil.ParseStringSliceContextValue(data["target_ids"]))
		if len(options) == 0 {
			return nil
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【冰霜祷言】请选择1名角色获得+1治疗：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	default:
		return nil
	}
}

func (e *GameEngine) handleSaintessChoiceInput(_ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "frost_prayer_target":
		return true, e.handleFrostPrayerChoice(selectionIndex, ctxData)
	default:
		return false, nil
	}
}

func (e *GameEngine) handleFrostPrayerChoice(selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
	if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	targetID := targetIDs[selectionIndex]
	target := e.State.Players[targetID]
	if target == nil {
		return fmt.Errorf("目标不存在")
	}
	e.Heal(targetID, 1)
	e.Log(fmt.Sprintf("%s 的 [冰霜祷言] 生效：%s +1治疗", user.Name, target.Name))
	e.PopInterrupt()
	return nil
}
