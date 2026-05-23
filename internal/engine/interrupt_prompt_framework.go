// gameflow: Prompt 构建公共框架与选项生成。

package engine

import (
	"fmt"
	"strconv"
	"strings"

	"starcup-engine/internal/engine/core/runtimeutil"
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
			{ID: "take", Label: "承受 (take) - 结算伤害/效果", ButtonLabel: "命中"},
			{ID: "counter", Label: "应战 (counter card_id) - 尝试反击", ButtonLabel: "应战"},
			{ID: "defend", Label: "防御 (defend) - 使用圣光（圣盾需提前放置）", ButtonLabel: "防御"},
		},
		Presentation: &model.PromptPresentation{
			Kind:   model.PresentationResponse,
			Layout: "inline",
		},
	}
}

func (e *GameEngine) buildResponseSkillPrompt() *model.Prompt {
	playerID := e.State.PendingInterrupt.PlayerID
	player := e.State.Players[playerID]
	skillIDs := e.State.PendingInterrupt.SkillIDs

	skillByID := buildSkillLookupMap(player)
	n := len(skillIDs)
	message := "你触发了响应技能，请选择要发动的技能。"
	if n > 1 {
		message = fmt.Sprintf("你触发了 %d 个响应技能，请选择 1 个发动，或跳过。", n)
	} else if n == 1 {
		if skill, ok := skillByID[skillIDs[0]]; ok && strings.TrimSpace(skill.Title) != "" {
			message = fmt.Sprintf("你触发了响应技能【%s】，请选择是否发动。", strings.TrimSpace(skill.Title))
		}
	}

	options := buildSkillSelectionOptions(skillIDs, skillByID)
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
		Presentation: &model.PromptPresentation{
			Kind:   model.PresentationSkillChoice,
			Layout: "overlay",
		},
	}
}

func (e *GameEngine) buildStartupSkillPrompt() *model.Prompt {
	playerID := e.State.PendingInterrupt.PlayerID
	player := e.State.Players[playerID]
	skillIDs := e.State.PendingInterrupt.SkillIDs

	options := buildSkillSelectionOptions(skillIDs, buildSkillLookupMap(player))
	options = append(options, model.PromptOption{
		ID:    "skip",
		Label: "跳过",
		Hint:  "本回合不发动启动技能",
	})

	return &model.Prompt{
		Type:     model.PromptChooseSkill,
		PlayerID: playerID,
		Message:  "你可以发动启动技能，请选择 1 个发动，或跳过。",
		Options:  options,
		Min:      1,
		Max:      1,
		Presentation: &model.PromptPresentation{
			Kind:   model.PresentationSkillChoice,
			Layout: "overlay",
		},
	}
}

// buildSkillLookupMap 从玩家角色技能列表构建 ID→SkillDefinition 映射。
func buildSkillLookupMap(player *model.Player) map[string]model.SkillDefinition {
	skillByID := make(map[string]model.SkillDefinition)
	if player != nil && player.Character != nil {
		for _, skill := range player.Character.Skills {
			skillByID[skill.ID] = skill
		}
	}
	return skillByID
}

// buildSkillSelectionOptions 从技能 ID 列表构建选项（不含跳过项，由调用方追加）。
func buildSkillSelectionOptions(skillIDs []string, skillByID map[string]model.SkillDefinition) []model.PromptOption {
	options := make([]model.PromptOption, 0, len(skillIDs)+1)
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
	return options
}

func formatCardInfo(card model.Card) string {
	return promptfmt.FormatCardInfo(card)
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

	giveCount := runtimeutil.ToIntContextValue(data["give_count"])
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
			ID:     strconv.Itoa(i),
			Label:  fmt.Sprintf("%d: %s", i+1, formatCardInfo(card)),
			CardID: card.ID,
		})
	}

	return &model.Prompt{
		Type:     model.PromptChooseCards,
		PlayerID: playerID,
		Message:  message,
		Options:  options,
		Min:      giveCount,
		Max:      giveCount,
		Presentation: &model.PromptPresentation{
			Kind:       model.PresentationCardPicker,
			CardSource: "hand",
		},
	}
}
