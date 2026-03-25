package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

func (e *GameEngine) buildCrimsonKnightChoicePrompt(choiceType, playerID string, _ *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "crk_bloody_prayer_x":
		maxX := toIntContextValue(data["max_x"])
		options := make([]model.PromptOption, 0, maxX)
		for x := 1; x <= maxX; x++ {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", x), Label: fmt.Sprintf("X=%d（移除%d治疗并对自己造成%d法伤）", x, x, x)})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【血腥祷言】请选择X值：", Options: options, Min: 1, Max: 1}

	case "crk_bloody_prayer_ally_count":
		allyIDs := parseStringSliceContextValue(data["ally_ids"])
		xValue := toIntContextValue(data["x_value"])
		if xValue <= 0 {
			xValue = 1
		}
		options := []model.PromptOption{{ID: "0", Label: "选择1名队友"}}
		if len(allyIDs) >= 2 && xValue >= 2 {
			options = append(options, model.PromptOption{ID: "1", Label: "选择2名队友（治疗将分配）"})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【血腥祷言】请选择要分配治疗的队友数量：", Options: options, Min: 1, Max: 1}

	case "crk_bloody_prayer_target":
		allyIDs := parseStringSliceContextValue(data["ally_ids"])
		selectedSet := idsToSet(parseStringSliceContextValue(data["selected_ally_ids"]))
		allyCount := toIntContextValue(data["ally_count"])
		if allyCount <= 0 {
			allyCount = 1
		}
		options := make([]model.PromptOption, 0, len(allyIDs))
		for _, allyID := range allyIDs {
			if selectedSet[allyID] {
				continue
			}
			if target := e.State.Players[allyID]; target != nil {
				options = append(options, model.PromptOption{ID: allyID, Label: target.Name})
			}
		}
		pickIndex := len(selectedSet) + 1
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: fmt.Sprintf("【血腥祷言】请选择第 %d/%d 名队友：", pickIndex, allyCount), Options: options, Min: 1, Max: 1}

	case "crk_bloody_prayer_split":
		selected := parseStringSliceContextValue(data["selected_ally_ids"])
		if len(selected) != 2 {
			return nil
		}
		xValue := toIntContextValue(data["x_value"])
		if xValue < 2 {
			return nil
		}
		first := e.State.Players[selected[0]]
		second := e.State.Players[selected[1]]
		if first == nil || second == nil {
			return nil
		}
		options := make([]model.PromptOption, 0, xValue-1)
		for firstHeal := 1; firstHeal < xValue; firstHeal++ {
			secondHeal := xValue - firstHeal
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", firstHeal-1), Label: fmt.Sprintf("%s +%d，%s +%d", first.Name, firstHeal, second.Name, secondHeal)})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【血腥祷言】请选择治疗分配：", Options: options, Min: 1, Max: 1}
	}

	return nil
}

func (e *GameEngine) handleCrimsonKnightChoiceInput(_ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)

	switch choiceType {
	case "crk_bloody_prayer_x":
		maxX := toIntContextValue(ctxData["max_x"])
		xValue := selectionIndex + 1
		if xValue < 1 || xValue > maxX {
			return true, fmt.Errorf("无效的X值")
		}
		ctxData["x_value"] = xValue
		ctxData["selected_ally_ids"] = []string{}
		allyIDs := parseStringSliceContextValue(ctxData["ally_ids"])
		if len(allyIDs) == 0 {
			return true, fmt.Errorf("没有可分配治疗的队友")
		}
		if len(allyIDs) >= 2 && xValue >= 2 {
			ctxData["choice_type"] = "crk_bloody_prayer_ally_count"
		} else {
			ctxData["ally_count"] = 1
			ctxData["choice_type"] = "crk_bloody_prayer_target"
		}
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil

	case "crk_bloody_prayer_ally_count":
		xValue := toIntContextValue(ctxData["x_value"])
		allyIDs := parseStringSliceContextValue(ctxData["ally_ids"])
		maxCount := 1
		if len(allyIDs) >= 2 && xValue >= 2 {
			maxCount = 2
		}
		allyCount := selectionIndex + 1
		if allyCount < 1 || allyCount > maxCount {
			return true, fmt.Errorf("无效的队友数量选择")
		}
		ctxData["ally_count"] = allyCount
		ctxData["selected_ally_ids"] = []string{}
		ctxData["choice_type"] = "crk_bloody_prayer_target"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil

	case "crk_bloody_prayer_target":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		allyIDs := parseStringSliceContextValue(ctxData["ally_ids"])
		selected := dedupeNonEmptyIDs(parseStringSliceContextValue(ctxData["selected_ally_ids"]))
		selectedSet := idsToSet(selected)
		remaining := make([]string, 0, len(allyIDs))
		for _, allyID := range allyIDs {
			if !selectedSet[allyID] {
				remaining = append(remaining, allyID)
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(remaining) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		allyCount := toIntContextValue(ctxData["ally_count"])
		if allyCount <= 0 {
			allyCount = 1
		}
		chosenID := remaining[selectionIndex]
		selected = append(selected, chosenID)
		ctxData["selected_ally_ids"] = selected

		xValue := toIntContextValue(ctxData["x_value"])
		if xValue <= 0 || user.Heal < xValue {
			return true, fmt.Errorf("治疗不足，无法结算血腥祷言")
		}
		if len(selected) < allyCount {
			ctxData["choice_type"] = "crk_bloody_prayer_target"
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return true, nil
		}

		if allyCount <= 1 {
			if err := e.resolveCrimsonKnightBloodyPrayer(user, xValue, map[string]int{selected[0]: xValue}); err != nil {
				return true, err
			}
		} else {
			if xValue < 2 {
				return true, fmt.Errorf("X不足以分配给2名队友")
			}
			ctxData["choice_type"] = "crk_bloody_prayer_split"
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return true, nil
		}

		e.PopInterrupt()
		if e.State.PendingInterrupt == nil && len(e.State.PendingDamageQueue) > 0 {
			e.enterDamageResolution(nil)
		}
		return true, nil

	case "crk_bloody_prayer_split":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		xValue := toIntContextValue(ctxData["x_value"])
		if xValue < 2 || user.Heal < xValue {
			return true, fmt.Errorf("治疗不足，无法结算血腥祷言")
		}
		selected := parseStringSliceContextValue(ctxData["selected_ally_ids"])
		if len(selected) != 2 {
			return true, fmt.Errorf("血腥祷言分配目标数量异常")
		}
		if selectionIndex < 0 || selectionIndex >= xValue-1 {
			return true, fmt.Errorf("无效的分配选项")
		}
		firstHeal := selectionIndex + 1
		secondHeal := xValue - firstHeal
		alloc := map[string]int{selected[0]: firstHeal, selected[1]: secondHeal}
		if err := e.resolveCrimsonKnightBloodyPrayer(user, xValue, alloc); err != nil {
			return true, err
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil && len(e.State.PendingDamageQueue) > 0 {
			e.enterDamageResolution(nil)
		}
		return true, nil
	}

	return false, nil
}

func dedupeNonEmptyIDs(ids []string) []string {
	result := make([]string, 0, len(ids))
	seen := map[string]bool{}
	for _, id := range ids {
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		result = append(result, id)
	}
	return result
}
