// gameflow: Prompt 构建公共框架与选项生成。

package engine

import (
	"fmt"
	"strconv"
	"strings"

	"starcup-engine/internal/engine/hook/promptfmt"
	"starcup-engine/internal/model"
)

func (e *GameEngine) buildStandardResponsePrompt() *model.Prompt {
	if !e.isResponseWindowActive() || len(e.State.ActionStack) == 0 {
		return nil
	}

	lastAction := e.State.ActionStack[len(e.State.ActionStack)-1]
	targetID := lastAction.TargetID

	return &model.Prompt{
		Type:     model.PromptConfirm,
		PlayerID: targetID,
		Message:  fmt.Sprintf("你成为了 %s 的目标，请做出响应 (take/counter/defend)", lastAction.Type),
		Options: []model.PromptOption{
			{ID: "take", Label: "承受 (take) - 结算伤害/效果"},
			{ID: "counter", Label: "应战 (counter <idx>) - 尝试反击"},
			{ID: "defend", Label: "防御 (defend) - 使用圣光（圣盾需提前放置）"},
		},
	}
}

func (e *GameEngine) buildResponseSkillPrompt() *model.Prompt {
	playerID := e.State.PendingInterrupt.PlayerID
	player := e.State.Players[playerID]

	skillIDs := e.State.PendingInterrupt.SkillIDs
	n := len(skillIDs)
	options := make([]model.PromptOption, 0, len(skillIDs)+1)

	skillByID := make(map[string]model.SkillDefinition)
	if player != nil && player.Character != nil {
		for _, skill := range player.Character.Skills {
			skillByID[skill.ID] = skill
		}
	}

	message := "你触发了响应技能，请选择要发动的技能。"
	if n > 1 {
		message = fmt.Sprintf("你触发了 %d 个响应技能，请选择 1 个发动，或跳过。", n)
	} else if n == 1 {
		if skill, ok := skillByID[skillIDs[0]]; ok && strings.TrimSpace(skill.Title) != "" {
			message = fmt.Sprintf("你触发了响应技能【%s】，请选择是否发动。", strings.TrimSpace(skill.Title))
		}
	}

	for i, skillID := range skillIDs {
		skill, ok := skillByID[skillID]
		if !ok {
			options = append(options, model.PromptOption{
				ID:    skillID,
				Label: fmt.Sprintf("技能 %d", i+1),
			})
			continue
		}

		label := strings.TrimSpace(skill.Title)
		if label == "" {
			label = fmt.Sprintf("技能 %d", i+1)
		}
		costStr := ""
		if skill.CostGem > 0 || skill.CostCrystal > 0 {
			costStr = fmt.Sprintf(" [💎%d 🔷%d]", skill.CostGem, skill.CostCrystal)
		}
		if costStr != "" {
			label = fmt.Sprintf("%s%s", label, costStr)
		}
		hint := strings.TrimSpace(skill.Description)
		options = append(options, model.PromptOption{
			ID:    skill.ID,
			Label: label,
			Hint:  hint,
		})
	}

	options = append(options, model.PromptOption{
		ID:    "skip",
		Label: "跳过",
		Hint:  "不发动响应技能",
	})

	return &model.Prompt{
		Type:     model.PromptChooseSkill,
		PlayerID: playerID,
		Message:  message,
		Options:  options,
		Min:      1,
		Max:      1,
	}
}

func (e *GameEngine) buildStartupSkillPrompt() *model.Prompt {
	playerID := e.State.PendingInterrupt.PlayerID
	player := e.State.Players[playerID]

	skillIDs := e.State.PendingInterrupt.SkillIDs
	message := "你可以发动启动技能，请选择 1 个发动，或跳过。"
	options := make([]model.PromptOption, 0, len(skillIDs)+1)

	skillByID := make(map[string]model.SkillDefinition)
	if player != nil && player.Character != nil {
		for _, skill := range player.Character.Skills {
			skillByID[skill.ID] = skill
		}
	}

	for i, skillID := range skillIDs {
		skill, ok := skillByID[skillID]
		if !ok {
			options = append(options, model.PromptOption{
				ID:    skillID,
				Label: fmt.Sprintf("技能 %d", i+1),
			})
			continue
		}

		label := strings.TrimSpace(skill.Title)
		if label == "" {
			label = fmt.Sprintf("技能 %d", i+1)
		}
		costStr := ""
		if skill.CostGem > 0 || skill.CostCrystal > 0 {
			costStr = fmt.Sprintf(" [💎%d 🔷%d]", skill.CostGem, skill.CostCrystal)
		}
		if costStr != "" {
			label = fmt.Sprintf("%s%s", label, costStr)
		}
		hint := strings.TrimSpace(skill.Description)
		options = append(options, model.PromptOption{
			ID:    skill.ID,
			Label: label,
			Hint:  hint,
		})
	}

	options = append(options, model.PromptOption{
		ID:    "skip",
		Label: "跳过",
		Hint:  "本回合不发动启动技能",
	})

	return &model.Prompt{
		Type:     model.PromptChooseSkill,
		PlayerID: playerID,
		Message:  message,
		Options:  options,
		Min:      1,
		Max:      1,
	}
}

func formatCardInfo(card model.Card) string {
	return promptfmt.FormatCardInfo(card)
}

func (e *GameEngine) buildDiscardPrompt() *model.Prompt {
	playerID := e.State.PendingInterrupt.PlayerID
	player := e.State.Players[playerID]

	data := e.State.PendingInterrupt.Context.(map[string]interface{})
	skillID, _ := data["skill_id"].(string)
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
	remainingIndices := parseIntSliceContextValue(data["remaining_indices"])
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
		if excludeBlessings && isElfBlessingCard(player, card.ID) {
			continue
		}
		options = append(options, model.PromptOption{
			ID:    strconv.Itoa(i),
			Label: fmt.Sprintf("%d: %s", i+1, formatCardInfo(card)),
		})
	}

	return &model.Prompt{
		Type:     model.PromptChooseCards,
		PlayerID: playerID,
		Message:  message,
		SkillID:  skillID,
		Options:  options,
		Min:      min,
		Max:      max,
	}
}

func (e *GameEngine) buildGiveCardsPrompt() *model.Prompt {
	playerID := e.State.PendingInterrupt.PlayerID
	player := e.State.Players[playerID]
	if player == nil {
		return nil
	}

	data, ok := e.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		return nil
	}

	var giveCount int
	if gc, ok := data["give_count"].(int); ok {
		giveCount = gc
	} else if gcf, ok := data["give_count"].(float64); ok {
		giveCount = int(gcf)
	}
	receiverID, _ := data["receiver_id"].(string)
	if giveCount <= 0 || receiverID == "" {
		return nil
	}

	receiver := e.State.Players[receiverID]
	receiverName := receiverID
	if receiver != nil {
		receiverName = receiver.Name
	}

	message := fmt.Sprintf("请选择 %d 张牌交给 %s：", giveCount, receiverName)

	var options []model.PromptOption
	for i, card := range player.Hand {
		options = append(options, model.PromptOption{
			ID:    strconv.Itoa(i),
			Label: fmt.Sprintf("%d: %s", i+1, formatCardInfo(card)),
		})
	}

	return &model.Prompt{
		Type:     model.PromptChooseCards,
		PlayerID: playerID,
		Message:  message,
		Options:  options,
		Min:      giveCount,
		Max:      giveCount,
	}
}
