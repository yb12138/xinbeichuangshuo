package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

func (e *GameEngine) buildWarHomunculusChoicePrompt(choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "hom_rune_reforge_distribution":
		total := toIntContextValue(data["total_runes"])
		if total <= 0 {
			total = 3
		}
		options := make([]model.PromptOption, 0, total+1)
		for warRunes := 0; warRunes <= total; warRunes++ {
			magicRunes := total - warRunes
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", warRunes), Label: fmt.Sprintf("战纹 %d / 魔纹 %d", warRunes, magicRunes)})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: fmt.Sprintf("【符文改造】请选择战纹/魔纹分配（总计%d）：", total), Options: options, Min: 1, Max: 1}

	case "hom_rune_smash_x", "hom_glyph_fusion_x":
		maxX := toIntContextValue(data["max_x"])
		minX := 1
		message := "【战纹碎击】请选择X（弃置同系牌数量）："
		if choiceType == "hom_glyph_fusion_x" {
			minX = 2
			message = "【魔纹融合】请选择X（弃置异系且元素互不相同的牌数量）："
		}
		options := make([]model.PromptOption, 0, maxX-minX+1)
		for xValue := minX; xValue <= maxX; xValue++ {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", xValue), Label: fmt.Sprintf("X=%d", xValue)})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: message, Options: options, Min: 1, Max: 1}

	case "hom_rune_smash_cards", "hom_glyph_fusion_cards":
		remaining := parseIntSliceContextValue(data["remaining_indices"])
		xValue := toIntContextValue(data["x_value"])
		selectedCount := len(parseIntSliceContextValue(data["selected_indices"]))
		options := make([]model.PromptOption, 0, len(remaining))
		for _, idx := range remaining {
			if player == nil || idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", idx), Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(player.Hand[idx]))})
		}
		remainingPick := xValue - selectedCount
		if remainingPick < 1 {
			remainingPick = 1
		}
		if len(options) > 0 && remainingPick > len(options) {
			remainingPick = len(options)
		}
		message := fmt.Sprintf("【战纹碎击】请选择要弃置的%d张牌：", remainingPick)
		if choiceType == "hom_glyph_fusion_cards" {
			message = fmt.Sprintf("【魔纹融合】请选择要弃置的%d张牌（元素不可重复）：", remainingPick)
		}
		return &model.Prompt{Type: model.PromptChooseCards, PlayerID: playerID, Message: message, Options: options, Min: remainingPick, Max: remainingPick}

	case "hom_rune_smash_y", "hom_glyph_fusion_y":
		maxY := toIntContextValue(data["max_y"])
		options := make([]model.PromptOption, 0, maxY+1)
		for yValue := 0; yValue <= maxY; yValue++ {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", yValue), Label: fmt.Sprintf("Y=%d", yValue)})
		}
		message := "【战纹碎击】请选择Y（额外翻转战纹数）："
		if choiceType == "hom_glyph_fusion_y" {
			message = "【魔纹融合】请选择Y（额外翻转魔纹数）："
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: message, Options: options, Min: 1, Max: 1}

	case "hom_dual_echo_target":
		targetIDs := parseStringSliceContextValue(data["target_ids"])
		options := make([]model.PromptOption, 0, len(targetIDs)+1)
		for _, targetID := range targetIDs {
			if target := e.State.Players[targetID]; target != nil {
				options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: target.Name})
			}
		}
		options = append(options, model.PromptOption{ID: "cancel", Label: "取消"})
		damage := toIntContextValue(data["damage"])
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: fmt.Sprintf("【双重回响】请选择额外造成%d点法术伤害的目标：", damage), Options: options, Min: 1, Max: 1}
	}

	return nil
}

func (e *GameEngine) handleWarHomunculusChoiceInput(_ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)

	switch choiceType {
	case "hom_rune_reforge_distribution":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		total := toIntContextValue(ctxData["total_runes"])
		if total <= 0 {
			total = 3
		}
		warRunes := selectionIndex
		if warRunes < 0 || warRunes > total {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		if user.Tokens == nil {
			user.Tokens = map[string]int{}
		}
		user.Tokens["hom_war_rune"] = warRunes
		user.Tokens["hom_magic_rune"] = total - warRunes
		e.Log(fmt.Sprintf("%s 的 [符文改造]：战纹=%d，魔纹=%d", user.Name, user.Tokens["hom_war_rune"], user.Tokens["hom_magic_rune"]))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.setTurnStage(model.TurnStageActionStart)
			e.clearCombatStage()
			e.clearSubflow()
		}
		return true, nil

	case "hom_rune_smash_x", "hom_glyph_fusion_x":
		maxX := toIntContextValue(ctxData["max_x"])
		minX := 1
		nextChoice := "hom_rune_smash_cards"
		if choiceType == "hom_glyph_fusion_x" {
			minX = 2
			nextChoice = "hom_glyph_fusion_cards"
		}
		xValue := selectionIndex
		if xValue < minX || xValue > maxX {
			xValue = selectionIndex + minX
		}
		if xValue < minX || xValue > maxX {
			return true, fmt.Errorf("无效的X值")
		}
		candidates := parseIntSliceContextValue(ctxData["candidate_indices"])
		if xValue > len(candidates) {
			return true, fmt.Errorf("可选牌数量不足")
		}
		ctxData["choice_type"] = nextChoice
		ctxData["x_value"] = xValue
		ctxData["selected_indices"] = []int{}
		ctxData["remaining_indices"] = append([]int{}, candidates...)
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil

	case "hom_rune_smash_cards", "hom_glyph_fusion_cards":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		remaining := parseIntSliceContextValue(ctxData["remaining_indices"])
		selected := parseIntSliceContextValue(ctxData["selected_indices"])
		xValue := toIntContextValue(ctxData["x_value"])
		cardIdx, ok := resolveSelectionToCandidate(selectionIndex, remaining)
		if !ok {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		if cardIdx < 0 || cardIdx >= len(user.Hand) {
			return true, fmt.Errorf("无效的手牌索引: %d", cardIdx)
		}
		attackElement, _ := ctxData["attack_element"].(string)
		if choiceType == "hom_rune_smash_cards" && attackElement != "" && string(user.Hand[cardIdx].Element) != attackElement {
			return true, fmt.Errorf("战纹碎击需选择与攻击同系的牌")
		}
		if choiceType == "hom_glyph_fusion_cards" && attackElement != "" && string(user.Hand[cardIdx].Element) == attackElement {
			return true, fmt.Errorf("魔纹融合需选择与攻击异系的牌")
		}
		if choiceType == "hom_glyph_fusion_cards" {
			for _, idx := range selected {
				if idx >= 0 && idx < len(user.Hand) && user.Hand[idx].Element == user.Hand[cardIdx].Element {
					return true, fmt.Errorf("魔纹融合需选择元素互不相同的异系牌")
				}
			}
		}
		selected = append(selected, cardIdx)
		nextRemaining := make([]int, 0, len(remaining))
		for _, idx := range remaining {
			if idx == cardIdx {
				continue
			}
			if choiceType == "hom_glyph_fusion_cards" {
				if idx >= 0 && idx < len(user.Hand) && user.Hand[idx].Element == user.Hand[cardIdx].Element {
					continue
				}
			}
			nextRemaining = append(nextRemaining, idx)
		}
		if len(selected) < xValue {
			ctxData["selected_indices"] = selected
			ctxData["remaining_indices"] = nextRemaining
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return true, nil
		}
		ctxData["selected_indices"] = selected
		maxY := toIntContextValue(ctxData["max_y"])
		if maxY > 0 {
			if choiceType == "hom_rune_smash_cards" {
				ctxData["choice_type"] = "hom_rune_smash_y"
			} else {
				ctxData["choice_type"] = "hom_glyph_fusion_y"
			}
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return true, nil
		}
		ctxData["y_value"] = 0
		return true, e.resolveHomunculusRuneChoice(ctxData, choiceType == "hom_glyph_fusion_cards")

	case "hom_rune_smash_y", "hom_glyph_fusion_y":
		maxY := toIntContextValue(ctxData["max_y"])
		yValue := selectionIndex
		if yValue < 0 || yValue > maxY {
			return true, fmt.Errorf("无效的Y值")
		}
		ctxData["y_value"] = yValue
		return true, e.resolveHomunculusRuneChoice(ctxData, choiceType == "hom_glyph_fusion_y")

	case "hom_dual_echo_target":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		targetIDs := parseStringSliceContextValue(ctxData["target_ids"])
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		costPending := toIntContextValue(ctxData["cost_pending"])
		if costPending > 0 {
			if !e.ConsumeCrystalCost(user.ID, costPending) {
				return true, fmt.Errorf("双重回响需要1蓝水晶（红宝石可替代）")
			}
			ctxData["cost_pending"] = 0
		}
		damage := toIntContextValue(ctxData["damage"])
		if damage < 0 {
			damage = 0
		}
		targetID := targetIDs[selectionIndex]
		e.AddPendingDamage(model.PendingDamage{SourceID: user.ID, TargetID: targetID, Damage: damage, DamageType: "magic_no_morale"})
		if target := e.State.Players[targetID]; target != nil {
			e.Log(fmt.Sprintf("%s 的 [双重回响] 对 %s 造成%d点法术伤害", user.Name, target.Name, damage))
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil && len(e.State.PendingDamageQueue) > 0 {
			e.enterDamageResolution(nil)
		}
		return true, nil
	}

	return false, nil
}
