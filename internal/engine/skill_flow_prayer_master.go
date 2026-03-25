package engine

import (
	"fmt"
	"strings"

	"starcup-engine/internal/model"
)

func (e *GameEngine) buildPrayerMasterChoicePrompt(choiceType, playerID string, _ *model.Player, _ map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "prayer_power_blessing_trigger":
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【威力赐福】是否移除该赐福，使本次攻击伤害+2？",
			Options:  []model.PromptOption{{ID: "0", Label: "是"}, {ID: "1", Label: "否"}},
			Min:      1,
			Max:      1,
		}
	case "prayer_swift_blessing_trigger":
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【迅捷赐福】是否移除该赐福，获得额外1次攻击行动？",
			Options:  []model.PromptOption{{ID: "0", Label: "是"}, {ID: "1", Label: "否"}},
			Min:      1,
			Max:      1,
		}
	}

	return nil
}

func (e *GameEngine) handlePrayerMasterChoiceInput(_ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)

	switch choiceType {
	case "prayer_power_blessing_trigger":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		if selectionIndex == 0 {
			e.RemoveFieldCard(user.ID, model.EffectPowerBlessing)
			sourceID, _ := ctxData["source_id"].(string)
			targetID, _ := ctxData["target_id"].(string)
			for i := range e.State.PendingDamageQueue {
				pd := &e.State.PendingDamageQueue[i]
				if pd.SourceID != sourceID || pd.TargetID != targetID {
					continue
				}
				if !strings.EqualFold(pd.DamageType, "Attack") {
					continue
				}
				pd.Damage += 2
				e.Log(fmt.Sprintf("%s 的 [威力赐福] 生效，本次攻击伤害+2", user.Name))
				break
			}
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil && len(e.State.PendingDamageQueue) > 0 {
			e.enterDamageResolution(nil)
		}
		return true, nil

	case "prayer_swift_blessing_trigger":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		if selectionIndex == 0 {
			e.RemoveFieldCard(user.ID, model.EffectSwiftBlessing)
			user.TurnState.PendingActions = append(user.TurnState.PendingActions, model.ActionContext{Source: "迅捷赐福", MustType: "Attack"})
			e.Log(fmt.Sprintf("%s 的 [迅捷赐福] 生效，获得额外攻击行动", user.Name))
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil && e.State.TurnStage != model.TurnStageExtraAction {
			e.enterExtraActionStage()
		}
		return true, nil
	}

	return false, nil
}
