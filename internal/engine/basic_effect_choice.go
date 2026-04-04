package engine

import (
	"fmt"
	"starcup-engine/internal/model"
)

type basicEffectChoiceOption struct {
	ID         string
	TargetID   string
	FieldIndex int
	Effect     string
	Label      string
}

func parseBasicEffectChoiceOptions(raw interface{}) []basicEffectChoiceOption {
	options := make([]basicEffectChoiceOption, 0)
	switch value := raw.(type) {
	case []basicEffectChoiceOption:
		options = append(options, value...)
	case []interface{}:
		for _, item := range value {
			m, ok := item.(map[string]interface{})
			if !ok || m == nil {
				continue
			}
			option, ok := parseBasicEffectChoiceOptionMap(m)
			if ok {
				options = append(options, option)
			}
		}
	case []map[string]interface{}:
		for _, item := range value {
			option, ok := parseBasicEffectChoiceOptionMap(item)
			if ok {
				options = append(options, option)
			}
		}
	}
	return options
}

func parseBasicEffectChoiceOptionMap(m map[string]interface{}) (basicEffectChoiceOption, bool) {
	if m == nil {
		return basicEffectChoiceOption{}, false
	}
	id, _ := m["id"].(string)
	targetID, _ := m["target_id"].(string)
	effect, _ := m["effect"].(string)
	label, _ := m["label"].(string)
	fieldIndex := -1
	if iv, ok := m["field_index"].(int); ok {
		fieldIndex = iv
	} else if fv, ok := m["field_index"].(float64); ok {
		fieldIndex = int(fv)
	}
	if targetID == "" || effect == "" || fieldIndex < 0 || label == "" {
		return basicEffectChoiceOption{}, false
	}
	if id == "" {
		id = fmt.Sprintf("%s|%d|%s", targetID, fieldIndex, effect)
	}
	return basicEffectChoiceOption{
		ID:         id,
		TargetID:   targetID,
		FieldIndex: fieldIndex,
		Effect:     effect,
		Label:      label,
	}, true
}

func buildBasicEffectChoicePrompt(playerID string, data map[string]interface{}) *model.Prompt {
	options := parseBasicEffectChoiceOptions(data["options"])
	if len(options) == 0 {
		return nil
	}
	promptMessage, _ := data["prompt"].(string)
	if promptMessage == "" {
		promptMessage = "请选择要处理的基础效果："
	}
	promptOptions := make([]model.PromptOption, 0, len(options))
	for _, option := range options {
		promptOptions = append(promptOptions, model.PromptOption{
			ID:    option.ID,
			Label: option.Label,
		})
	}
	return &model.Prompt{
		Type:     model.PromptConfirm,
		PlayerID: playerID,
		Message:  promptMessage,
		Options:  promptOptions,
		Min:      1,
		Max:      1,
	}
}

func (e *GameEngine) handleBasicEffectChoiceInput(playerID string, selectionIndex int, data map[string]interface{}) error {
	options := parseBasicEffectChoiceOptions(data["options"])
	if selectionIndex < 0 || selectionIndex >= len(options) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	selected := options[selectionIndex]
	operation, _ := data["operation"].(string)
	skillName, _ := data["skill_name"].(string)
	user := e.State.Players[playerID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}

	switch operation {
	case "remove":
		if _, err := e.RemoveFieldCardAt(selected.TargetID, selected.FieldIndex, playerID); err != nil {
			return err
		}
		e.Log(fmt.Sprintf("%s 的 [%s]：移除了 %s", user.Name, skillName, selected.Label))
		e.NotifyActionStep(fmt.Sprintf("%s 的【%s】移除了 %s", user.Name, skillName, selected.Label))
	case "take":
		takenCard, err := e.TakeFieldCard(selected.TargetID, selected.FieldIndex, playerID)
		if err != nil {
			return err
		}
		user.Hand = append(user.Hand, takenCard)
		e.Log(fmt.Sprintf("%s 的 [%s]：收回了 %s，并将该牌加入手牌", user.Name, skillName, selected.Label))
		e.NotifyActionStep(fmt.Sprintf("%s 的【%s】收回了 %s", user.Name, skillName, selected.Label))
	default:
		return fmt.Errorf("未知的基础效果选择操作: %s", operation)
	}

	e.PopInterrupt()
	if e.State.PendingInterrupt == nil {
		// 规则：这里是技能执行中的“目标选择子步骤”，不是系统自动阶段结算。
		// 选择完成后按技能声明的 resume_phase 继续流程，保证后续仍在该技能约束的阶段节点上。
		e.applyChoiceResumePoint(mustChoiceResumePointFromMap(data, "resume_phase"))
	}
	return nil
}

func (e *GameEngine) RemoveFieldCardAt(targetID string, fieldIndex int, sourceID string) (model.Card, error) {
	target := e.State.Players[targetID]
	if target == nil {
		return model.Card{}, fmt.Errorf("目标不存在")
	}
	if fieldIndex < 0 || fieldIndex >= len(target.Field) {
		return model.Card{}, fmt.Errorf("无效的场上牌索引")
	}
	fc := target.Field[fieldIndex]
	if fc == nil {
		return model.Card{}, fmt.Errorf("场上牌不存在")
	}

	target.Field = append(target.Field[:fieldIndex], target.Field[fieldIndex+1:]...)
	e.State.DiscardPile = append(e.State.DiscardPile, fc.Card)
	e.Log(fmt.Sprintf("%s 的场上牌被移除: %s", target.Name, fc.Effect))
	if fc.Mode == model.FieldEffect {
		e.emitBuffRemovedTrigger(sourceID, targetID, fc.Effect)
	}
	return fc.Card, nil
}
