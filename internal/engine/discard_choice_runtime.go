// gameflow: 弃牌交互统一走 InterruptChoice + choice_type 的桥接与运行时辅助。

package engine

import (
	"fmt"
	"strconv"
	"strings"

	playerpkg "starcup-engine/internal/engine/player"
	elfarcherpkg "starcup-engine/internal/engine/player/elf_archer"
	"starcup-engine/internal/model"
)

const choiceTypeSystemDiscardCards = "system_discard_cards"

func normalizeDiscardChoiceContext(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		data = map[string]interface{}{}
	}
	if choiceType, _ := data["choice_type"].(string); strings.TrimSpace(choiceType) == "" {
		data["choice_type"] = choiceTypeSystemDiscardCards
	}
	data["discard_subflow"] = true
	return data
}

func newDiscardChoiceInterrupt(playerID string, data map[string]interface{}) *model.Interrupt {
	return &model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: playerID,
		Context:  normalizeDiscardChoiceContext(data),
	}
}

func isBeastSamuraiDiscardChoiceType(choiceType string) bool {
	return strings.HasPrefix(choiceType, "bs_") && strings.HasSuffix(choiceType, "_discard")
}

func isDiscardChoiceType(choiceType string) bool {
	return choiceType == choiceTypeSystemDiscardCards || isBeastSamuraiDiscardChoiceType(choiceType)
}

func (e *GameEngine) pendingDiscardContext() (map[string]interface{}, error) {
	if e == nil || e.State == nil || e.State.PendingInterrupt == nil {
		return nil, fmt.Errorf("当前没有待处理的弃牌操作")
	}
	intr := e.State.PendingInterrupt
	if intr.Type == model.InterruptChoice {
		data, _ := intr.Context.(map[string]interface{})
		if data == nil {
			return nil, fmt.Errorf("弃牌中断上下文格式错误")
		}
		choiceType, _ := data["choice_type"].(string)
		if isDiscardChoiceType(choiceType) || isDiscardSelectionInterrupt(intr) {
			return data, nil
		}
	}
	return nil, fmt.Errorf("当前没有待处理的弃牌操作")
}

func (e *GameEngine) buildDiscardChoicePromptFromData(playerID string, data map[string]interface{}) *model.Prompt {
	if e == nil || e.State == nil {
		return nil
	}
	player := e.State.Players[playerID]
	if player == nil || data == nil {
		return nil
	}
	skillID, _ := data["skill_id"].(string)
	promptChoiceType, _ := data["choice_type"].(string)
	if strings.TrimSpace(promptChoiceType) == "" {
		promptChoiceType = choiceTypeSystemDiscardCards
	}
	var message string
	var min, max int

	if count, ok := data["discard_count"].(int); ok && count > 0 {
		min = count
		max = count
		message = fmt.Sprintf("手牌上限溢出！请弃置 %d 张牌：", count)
		if customMsg, ok := data["prompt"].(string); ok && customMsg != "" {
			message = customMsg
		}
	} else {
		if v, ok := data["min"].(int); ok {
			min = v
		} else {
			min = 1
		}
		if v, ok := data["max"].(int); ok && v > 0 {
			max = v
		} else {
			max = len(player.Hand)
		}
		if customMsg, ok := data["prompt"].(string); ok && customMsg != "" {
			message = customMsg
		} else {
			message = fmt.Sprintf("请选择 %d-%d 张牌弃置：", min, max)
		}
	}

	var options []model.PromptOption
	discardType, _ := data["discard_type"].(model.CardType)
	discardElement, _ := data["discard_element"].(model.Element)
	excludeBlessings, _ := data["exclude_blessings"].(bool)
	remainingIndices := playerpkg.ParseIntSliceContextValue(data["remaining_indices"])
	allowedIndices := map[int]struct{}{}
	if len(remainingIndices) > 0 {
		for _, idx := range remainingIndices {
			allowedIndices[idx] = struct{}{}
		}
	}
	for i, card := range player.Hand {
		if len(allowedIndices) > 0 {
			if _, ok := allowedIndices[i]; !ok {
				continue
			}
		}
		if discardType != "" && card.Type != discardType {
			continue
		}
		if discardElement != "" && card.Element != discardElement {
			continue
		}
		if excludeBlessings && elfarcherpkg.IsBlessingCard(player, card) {
			continue
		}
		options = append(options, model.PromptOption{
			ID:    strconv.Itoa(i),
			Label: fmt.Sprintf("%d: %s", i+1, formatCardInfo(card)),
		})
	}

	return &model.Prompt{
		Type:       model.PromptChooseCards,
		PlayerID:   playerID,
		ChoiceType: promptChoiceType,
		Message:    message,
		SkillID:    skillID,
		Options:    options,
		Min:        min,
		Max:        max,
	}
}
