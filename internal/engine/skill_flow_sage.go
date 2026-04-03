package engine

import (
	"fmt"
	"starcup-engine/internal/engine/runtimeutil"

	"starcup-engine/internal/model"
)

func (e *GameEngine) buildSageChoicePrompt(choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "sage_magic_rebound_confirm":
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【法术反弹】是否发动？",
			Options: []model.PromptOption{
				{ID: "0", Label: "是"},
				{ID: "1", Label: "否"},
			},
			Min: 1,
			Max: 1,
		}
	case "sage_magic_rebound_x":
		maxX := runtimeutil.ToIntContextValue(data["max_x"])
		options := make([]model.PromptOption, 0, max(0, maxX-1))
		for x := 2; x <= maxX; x++ {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", x-2),
				Label: fmt.Sprintf("X=%d（弃%d张同系牌）", x, x),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【法术反弹】请选择X值：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	case "sage_magic_rebound_element":
		xValue := runtimeutil.ToIntContextValue(data["x_value"])
		elements := availableElementsByMinCount(player, xValue)
		options := make([]model.PromptOption, 0, len(elements))
		for idx, ele := range elements {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%s系", elementNameForPrompt(ele)),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【法术反弹】请选择弃置同系牌的元素：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	case "sage_magic_rebound_cards", "sage_arcane_cards", "sage_holy_cards":
		remaining := parseIntSliceContextValue(data["remaining_indices"])
		selectedCount := len(parseIntSliceContextValue(data["selected_indices"]))
		targetCount := runtimeutil.ToIntContextValue(data["x_value"])
		options := make([]model.PromptOption, 0, len(remaining))
		for _, idx := range remaining {
			if player == nil || idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(player.Hand[idx])),
			})
		}
		remainingPick := targetCount - selectedCount
		if remainingPick < 1 {
			remainingPick = 1
		}
		if len(options) > 0 && remainingPick > len(options) {
			remainingPick = len(options)
		}
		msg := fmt.Sprintf("请选择%d张牌：", remainingPick)
		switch choiceType {
		case "sage_magic_rebound_cards":
			msg = fmt.Sprintf("【法术反弹】请选择%d张同系牌：", remainingPick)
		case "sage_arcane_cards":
			msg = fmt.Sprintf("【魔道法典】请选择%d张异系牌：", remainingPick)
		case "sage_holy_cards":
			msg = fmt.Sprintf("【圣洁法典】请选择%d张异系牌：", remainingPick)
		}
		return &model.Prompt{
			Type:     model.PromptChooseCards,
			PlayerID: playerID,
			Message:  msg,
			Options:  options,
			Min:      remainingPick,
			Max:      remainingPick,
		}
	case "sage_arcane_x", "sage_holy_x":
		maxX := runtimeutil.ToIntContextValue(data["max_x"])
		minX := 2
		if choiceType == "sage_holy_x" {
			minX = 3
		}
		options := make([]model.PromptOption, 0, max(0, maxX-minX+1))
		for x := minX; x <= maxX; x++ {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", x-minX),
				Label: fmt.Sprintf("X=%d（弃%d张异系牌）", x, x),
			})
		}
		msg := "【魔道法典】请选择X值："
		if choiceType == "sage_holy_x" {
			msg = "【圣洁法典】请选择X值："
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  msg,
			Options:  options,
			Min:      1,
			Max:      1,
		}
	case "sage_holy_target_count":
		maxCount := runtimeutil.ToIntContextValue(data["max_target_count"])
		if maxCount < 1 {
			maxCount = 1
		}
		options := make([]model.PromptOption, 0, maxCount)
		for count := 1; count <= maxCount; count++ {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", count-1),
				Label: fmt.Sprintf("选择%d名角色", count),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【圣洁法典】请选择要获得治疗的角色数量：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	case "sage_holy_targets":
		allTargetIDs := runtimeutil.ParseStringSliceContextValue(data["target_ids"])
		selectedSet := runtimeutil.IDsToSet(runtimeutil.ParseStringSliceContextValue(data["selected_target_ids"]))
		targetCount := runtimeutil.ToIntContextValue(data["target_count"])
		options := make([]model.PromptOption, 0, len(allTargetIDs))
		for _, targetID := range allTargetIDs {
			if selectedSet[targetID] {
				continue
			}
			if target := e.State.Players[targetID]; target != nil {
				options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: target.Name})
			}
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【圣洁法典】请选择第 %d/%d 名治疗目标：", len(selectedSet)+1, targetCount),
			Options:  options,
			Min:      1,
			Max:      1,
		}
	case "sage_magic_rebound_target", "sage_arcane_target":
		targetIDs := runtimeutil.ParseStringSliceContextValue(data["target_ids"])
		options := make([]model.PromptOption, 0, len(targetIDs))
		for _, targetID := range targetIDs {
			if target := e.State.Players[targetID]; target != nil {
				options = append(options, model.PromptOption{ID: targetID, Label: target.Name})
			}
		}
		msg := "请选择目标角色："
		if choiceType == "sage_magic_rebound_target" {
			msg = "【法术反弹】请选择目标角色："
		} else {
			msg = "【魔道法典】请选择目标角色："
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  msg,
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}

	return nil
}

func (e *GameEngine) handleSageChoiceInput(_ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)

	switch choiceType {
	case "sage_magic_rebound_confirm":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		if selectionIndex == 1 {
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil && len(e.State.PendingDamageQueue) > 0 {
				e.enterDamageResolution(nil)
			}
			return true, nil
		}
		if selectionIndex != 0 {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		maxX := maxSameElementCount(user)
		if maxX < 2 {
			return true, fmt.Errorf("同系手牌不足2张，无法发动法术反弹")
		}
		ctxData["choice_type"] = "sage_magic_rebound_x"
		ctxData["max_x"] = maxX
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil

	case "sage_magic_rebound_x":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		xValue := selectionIndex + 2
		maxX := maxSameElementCount(user)
		if xValue < 2 || xValue > maxX {
			return true, fmt.Errorf("无效的X值")
		}
		ctxData["x_value"] = xValue
		ctxData["choice_type"] = "sage_magic_rebound_element"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil

	case "sage_magic_rebound_element":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		xValue := runtimeutil.ToIntContextValue(ctxData["x_value"])
		elements := availableElementsByMinCount(user, xValue)
		if selectionIndex < 0 || selectionIndex >= len(elements) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		chosenElement := model.Element(elements[selectionIndex])
		ctxData["chosen_element"] = string(chosenElement)
		ctxData["choice_type"] = "sage_magic_rebound_cards"
		ctxData["selected_indices"] = []int{}
		ctxData["remaining_indices"] = getCardIndicesByElement(user, chosenElement)
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil

	case "sage_magic_rebound_cards", "sage_arcane_cards", "sage_holy_cards":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		xValue := runtimeutil.ToIntContextValue(ctxData["x_value"])
		if xValue <= 0 {
			return true, fmt.Errorf("X值无效")
		}
		remaining := parseIntSliceContextValue(ctxData["remaining_indices"])
		selected := parseIntSliceContextValue(ctxData["selected_indices"])
		cardIdx, ok := runtimeutil.ResolveSelectionToCandidate(selectionIndex, remaining)
		if !ok || cardIdx < 0 || cardIdx >= len(user.Hand) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		chosenCard := user.Hand[cardIdx]
		for _, idx := range selected {
			if choiceType != "sage_magic_rebound_cards" &&
				idx >= 0 && idx < len(user.Hand) &&
				user.Hand[idx].Element == chosenCard.Element {
				return true, fmt.Errorf("需弃置异系牌，不能重复选择同系")
			}
		}
		if choiceType == "sage_magic_rebound_cards" {
			chosenElement, _ := ctxData["chosen_element"].(string)
			if string(chosenCard.Element) != chosenElement {
				return true, fmt.Errorf("法术反弹需弃置同系牌")
			}
		}
		selected = append(selected, cardIdx)
		nextRemaining := make([]int, 0, len(remaining))
		if choiceType == "sage_magic_rebound_cards" {
			for _, idx := range remaining {
				if idx != cardIdx {
					nextRemaining = append(nextRemaining, idx)
				}
			}
		} else {
			nextRemaining = removeElementIndices(remaining, user, chosenCard.Element, cardIdx)
		}
		if len(selected) < xValue {
			ctxData["selected_indices"] = selected
			ctxData["remaining_indices"] = nextRemaining
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return true, nil
		}

		ctxData["selected_indices"] = selected
		switch choiceType {
		case "sage_magic_rebound_cards":
			ctxData["choice_type"] = "sage_magic_rebound_target"
			ctxData["target_ids"] = e.allOtherPlayerIDs(user.ID)
		case "sage_arcane_cards":
			ctxData["choice_type"] = "sage_arcane_target"
			ctxData["target_ids"] = e.allOtherPlayerIDs(user.ID)
		case "sage_holy_cards":
			maxTargetCount := xValue - 2
			if maxTargetCount < 1 {
				return true, fmt.Errorf("圣洁法典治疗目标数量无效")
			}
			ctxData["choice_type"] = "sage_holy_target_count"
			ctxData["max_target_count"] = maxTargetCount
			ctxData["target_ids"] = append([]string{}, e.State.PlayerOrder...)
		}
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil

	case "sage_arcane_x", "sage_holy_x":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
		minX := 2
		if choiceType == "sage_holy_x" {
			minX = 3
		}
		xValue := selectionIndex + minX
		if xValue < minX || xValue > maxX {
			return true, fmt.Errorf("无效的X值")
		}
		ctxData["x_value"] = xValue
		ctxData["selected_indices"] = []int{}
		ctxData["remaining_indices"] = allHandIndices(user)
		if choiceType == "sage_arcane_x" {
			ctxData["choice_type"] = "sage_arcane_cards"
		} else {
			ctxData["choice_type"] = "sage_holy_cards"
		}
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil

	case "sage_holy_target_count":
		targetCount := selectionIndex + 1
		maxCount := runtimeutil.ToIntContextValue(ctxData["max_target_count"])
		if targetCount < 1 || targetCount > maxCount {
			return true, fmt.Errorf("无效的治疗目标数量")
		}
		ctxData["target_count"] = targetCount
		ctxData["selected_target_ids"] = []string{}
		ctxData["choice_type"] = "sage_holy_targets"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil

	case "sage_holy_targets":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		targetCount := runtimeutil.ToIntContextValue(ctxData["target_count"])
		if targetCount <= 0 {
			return true, fmt.Errorf("治疗目标数量无效")
		}
		allTargetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
		selected := runtimeutil.ParseStringSliceContextValue(ctxData["selected_target_ids"])
		selectedSet := runtimeutil.IDsToSet(selected)
		remaining := make([]string, 0, len(allTargetIDs))
		for _, targetID := range allTargetIDs {
			if !selectedSet[targetID] {
				remaining = append(remaining, targetID)
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(remaining) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		selected = append(selected, remaining[selectionIndex])
		ctxData["selected_target_ids"] = selected
		if len(selected) < targetCount {
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return true, nil
		}

		selectedCards := parseIntSliceContextValue(ctxData["selected_indices"])
		xValue := runtimeutil.ToIntContextValue(ctxData["x_value"])
		if xValue <= 2 || len(selectedCards) != xValue {
			return true, fmt.Errorf("圣洁法典弃牌参数无效")
		}
		removed, err := removeCardsByIndicesFromHand(user, append([]int{}, selectedCards...))
		if err != nil {
			return true, err
		}
		e.NotifyCardRevealed(user.ID, removed, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, removed...)
		for _, targetID := range selected {
			e.Heal(targetID, 2)
		}
		damage := xValue - 1
		if damage > 0 {
			e.AddPendingDamage(model.PendingDamage{
				SourceID:   user.ID,
				TargetID:   user.ID,
				Damage:     damage,
				DamageType: "magic",
			})
		}
		e.Log(fmt.Sprintf("%s 发动 [圣洁法典]：为%d名角色各+2治疗，并对自己造成%d点法术伤害", user.Name, len(selected), damage))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if !e.routePendingDamageWithReturn(model.TurnStageExtraAction) {
				e.enterExtraActionStage()
			}
		}
		return true, nil

	case "sage_magic_rebound_target", "sage_arcane_target":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		selected := parseIntSliceContextValue(ctxData["selected_indices"])
		xValue := runtimeutil.ToIntContextValue(ctxData["x_value"])
		if xValue <= 1 || len(selected) != xValue {
			return true, fmt.Errorf("弃牌参数无效")
		}
		removed, err := removeCardsByIndicesFromHand(user, append([]int{}, selected...))
		if err != nil {
			return true, err
		}
		e.NotifyCardRevealed(user.ID, removed, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, removed...)
		targetID := targetIDs[selectionIndex]
		if targetID == user.ID {
			return true, fmt.Errorf("该技能不能以自己为目标")
		}
		target := e.State.Players[targetID]
		if target == nil {
			return true, fmt.Errorf("目标不存在")
		}

		switch choiceType {
		case "sage_magic_rebound_target":
			damageToTarget := xValue - 1
			damageToSelf := xValue
			pending := make([]model.PendingDamage, 0, 2)
			if damageToTarget > 0 {
				pending = append(pending, model.PendingDamage{
					SourceID:   user.ID,
					TargetID:   targetID,
					Damage:     damageToTarget,
					DamageType: "magic",
				})
			}
			if damageToSelf > 0 {
				pending = append(pending, model.PendingDamage{
					SourceID:   user.ID,
					TargetID:   user.ID,
					Damage:     damageToSelf,
					DamageType: "magic",
				})
			}
			e.prependPendingDamages(pending)
			e.Log(fmt.Sprintf("%s 发动 [法术反弹]：弃%d张同系牌，对 %s 造成%d点法术伤害，并对自己造成%d点法术伤害", user.Name, xValue, target.Name, damageToTarget, damageToSelf))
		case "sage_arcane_target":
			damage := xValue - 1
			if damage > 0 {
				e.AddPendingDamage(model.PendingDamage{
					SourceID:   user.ID,
					TargetID:   targetID,
					Damage:     damage,
					DamageType: "magic",
				})
				e.AddPendingDamage(model.PendingDamage{
					SourceID:   user.ID,
					TargetID:   user.ID,
					Damage:     damage,
					DamageType: "magic",
				})
			}
			e.Log(fmt.Sprintf("%s 发动 [魔道法典]：弃%d张异系牌，对 %s 与自己各造成%d点法术伤害", user.Name, xValue, target.Name, damage))
		}

		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if !e.routePendingDamageWithReturn(model.TurnStageExtraAction) {
				e.enterExtraActionStage()
			}
		}
		return true, nil
	}

	return false, nil
}
