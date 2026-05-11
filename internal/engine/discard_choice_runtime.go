// gameflow: 弃牌交互统一走 InterruptChoice + choice_type 的桥接与运行时辅助。

package engine

import (
	"fmt"
	"strconv"
	"strings"

	"starcup-engine/internal/engine/core/runtimeutil"
	playerpkg "starcup-engine/internal/engine/player"
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

func IsDiscardChoiceType(choiceType string) bool {
	return choiceType == choiceTypeSystemDiscardCards
}

// isDiscardSubflow checks whether the interrupt context represents a discard
// sub-flow (regardless of which role owns the choice_type).
func isDiscardSubflow(data map[string]interface{}) bool {
	return runtimeutil.ToBoolContextValue(data["discard_subflow"])
}

// shouldExcludeCardFromDiscard 遍历所有角色条目，判断某张牌是否应从弃牌选项中排除。
func shouldExcludeCardFromDiscard(player *model.Player, card model.Card) bool {
	for _, entry := range roleRegistry.Entries() {
		if entry.ExcludeCardFromDiscard != nil && entry.ExcludeCardFromDiscard(player, card) {
			return true
		}
	}
	return false
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
		if IsDiscardChoiceType(choiceType) || isDiscardSubflow(data) || IsDiscardSelectionInterrupt(intr) {
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
	} else if downTo, ok := data["discard_down_to"].(int); ok && downTo >= 0 {
		count := len(player.Hand) - downTo
		if count <= 0 {
			return nil
		}
		min = count
		max = count
		message = fmt.Sprintf("请弃置 %d 张牌（弃至%d张）：", count, downTo)
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
		if excludeBlessings && shouldExcludeCardFromDiscard(player, card) {
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
