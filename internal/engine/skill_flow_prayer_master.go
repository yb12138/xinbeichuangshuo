// gameflow: 祈祷师：威力赐福/迅捷赐福等在 PendingDamage 前后的确认中断。

package engine

import (
	"fmt"
	"strings"

	"starcup-engine/internal/model"
)

func (e *GameEngine) buildPrayerMasterChoicePrompt(choiceType, playerID string, _ *model.Player, _ map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "prayer_power_blessing_followup":
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【威力赐福】是否移除该赐福，使本次攻击伤害+2？",
			Options:  []model.PromptOption{{ID: "0", Label: "是"}, {ID: "1", Label: "否"}},
			Min:      1,
			Max:      1,
		}
	case "prayer_swift_blessing_followup":
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
	return dispatchChoiceInputByType(choiceType, selectionIndex, ctxData, map[string]skillChoiceInputHandler{
		"prayer_power_blessing_followup": e.handlePrayerPowerBlessingFollowupChoice,
		"prayer_swift_blessing_followup": e.handlePrayerSwiftBlessingFollowupChoice,
	})
}

func (e *GameEngine) handlePrayerPowerBlessingFollowupChoice(selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
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
			if !strings.EqualFold(string(pd.DamageType), string(model.AttackDamage)) {
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
	return nil
}

func (e *GameEngine) handlePrayerSwiftBlessingFollowupChoice(selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	if selectionIndex == 0 {
		e.RemoveFieldCard(user.ID, model.EffectSwiftBlessing)
		model.AppendAttackAction(user, "迅捷赐福")
		e.Log(fmt.Sprintf("%s 的 [迅捷赐福] 生效，获得额外攻击行动", user.Name))
	}
	e.PopInterrupt()
	if e.State.PendingInterrupt == nil && e.State.TurnStage != model.TurnStageExtraAction {
		e.enterExtraActionStage()
	}
	return nil
}
