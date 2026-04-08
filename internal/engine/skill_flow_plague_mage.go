package engine

import (
	"fmt"
	"starcup-engine/internal/engine/runtimeutil"

	"starcup-engine/internal/model"
)

func (e *GameEngine) buildPlagueMageChoicePrompt(choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "plague_death_touch_element":
		elements := runtimeutil.ParseStringSliceContextValue(data["elements"])
		options := make([]model.PromptOption, 0, len(elements))
		for i, ele := range elements {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", i),
				Label: fmt.Sprintf("%s系", elementNameForPrompt(ele)),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【死亡之触】请选择弃置同系牌的元素：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	case "plague_death_touch_x":
		maxHeal := runtimeutil.ToIntContextValue(data["max_heal"])
		options := make([]model.PromptOption, 0, maxHeal-1)
		for x := 2; x <= maxHeal; x++ {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", x-2),
				Label: fmt.Sprintf("X=%d（移除%d点治疗）", x, x),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【死亡之触】请选择X值：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	case "plague_death_touch_y":
		maxCards := runtimeutil.ToIntContextValue(data["max_cards"])
		options := make([]model.PromptOption, 0, maxCards-1)
		for y := 2; y <= maxCards; y++ {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", y-2),
				Label: fmt.Sprintf("Y=%d（弃%d张同系牌）", y, y),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【死亡之触】请选择Y值：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	case "plague_death_touch_cards":
		remaining := runtimeutil.ParseChoiceIntSlice(data["remaining_indices"])
		yNeed := runtimeutil.ToIntContextValue(data["y_value"])
		selectedCount := len(runtimeutil.ParseChoiceIntSlice(data["selected_indices"]))
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
		remainingPick := yNeed - selectedCount
		if remainingPick < 1 {
			remainingPick = 1
		}
		if len(options) > 0 && remainingPick > len(options) {
			remainingPick = len(options)
		}
		return &model.Prompt{
			Type:     model.PromptChooseCards,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【死亡之触】请选择要弃置的%d张牌：", remainingPick),
			Options:  options,
			Min:      remainingPick,
			Max:      remainingPick,
		}
	case "plague_death_touch_target":
		targetIDs := runtimeutil.ParseStringSliceContextValue(data["target_ids"])
		options := make([]model.PromptOption, 0, len(targetIDs))
		for i, targetID := range targetIDs {
			if target := e.State.Players[targetID]; target != nil {
				options = append(options, model.PromptOption{
					ID:    fmt.Sprintf("%d", i),
					Label: target.Name,
				})
			}
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【死亡之触】请选择1名敌方角色承受法术伤害：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	default:
		return nil
	}
}

func (e *GameEngine) handlePlagueMageChoiceInput(_ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	return dispatchChoiceInputByType(choiceType, selectionIndex, ctxData, map[string]skillChoiceInputHandler{
		"plague_death_touch_element": func(idx int, data map[string]interface{}) error {
			return e.handlePlagueDeathTouchElementChoice(data, idx)
		},
		"plague_death_touch_x": func(idx int, data map[string]interface{}) error {
			return e.handlePlagueDeathTouchXChoice(data, idx)
		},
		"plague_death_touch_y": func(idx int, data map[string]interface{}) error {
			return e.handlePlagueDeathTouchYChoice(data, idx)
		},
		"plague_death_touch_cards": func(idx int, data map[string]interface{}) error {
			return e.handlePlagueDeathTouchCardsChoice(data, idx)
		},
		"plague_death_touch_target": func(idx int, data map[string]interface{}) error {
			return e.handlePlagueDeathTouchTargetChoice(data, idx)
		},
	})
}

func (e *GameEngine) handlePlagueDeathTouchElementChoice(ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}

	elements := runtimeutil.ParseStringSliceContextValue(ctxData["elements"])
	if selectionIndex < 0 || selectionIndex >= len(elements) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}

	chosenElement := elements[selectionIndex]
	e.State.PendingInterrupt.Context = map[string]interface{}{
		"choice_type":      "plague_death_touch_x",
		"user_id":          userID,
		"target_id":        ctxData["target_id"],
		"chosen_element":   chosenElement,
		"max_heal":         user.Heal,
		"max_cards":        len(getCardIndicesByElement(user, model.Element(chosenElement))),
		"selected_indices": []int{},
	}
	e.notifyInterruptPrompt()
	return nil
}

func (e *GameEngine) handlePlagueDeathTouchXChoice(ctxData map[string]interface{}, selectionIndex int) error {
	xValue := selectionIndex + 2
	if maxHeal := runtimeutil.ToIntContextValue(ctxData["max_heal"]); xValue < 2 || xValue > maxHeal {
		return fmt.Errorf("无效的X值")
	}
	ctxData["choice_type"] = "plague_death_touch_y"
	ctxData["x_value"] = xValue
	e.State.PendingInterrupt.Context = ctxData
	e.notifyInterruptPrompt()
	return nil
}

func (e *GameEngine) handlePlagueDeathTouchYChoice(ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}

	yValue := selectionIndex + 2
	if maxCards := runtimeutil.ToIntContextValue(ctxData["max_cards"]); yValue < 2 || yValue > maxCards {
		return fmt.Errorf("无效的Y值")
	}

	chosenElement, _ := ctxData["chosen_element"].(string)
	ctxData["choice_type"] = "plague_death_touch_cards"
	ctxData["y_value"] = yValue
	ctxData["selected_indices"] = []int{}
	ctxData["remaining_indices"] = getCardIndicesByElement(user, model.Element(chosenElement))
	e.State.PendingInterrupt.Context = ctxData
	e.notifyInterruptPrompt()
	return nil
}

func (e *GameEngine) handlePlagueDeathTouchCardsChoice(ctxData map[string]interface{}, selectionIndex int) error {
	remaining := runtimeutil.ParseChoiceIntSlice(ctxData["remaining_indices"])
	selected := append([]int{}, runtimeutil.ParseChoiceIntSlice(ctxData["selected_indices"])...)
	yValue := runtimeutil.ToIntContextValue(ctxData["y_value"])

	cardIdx, ok := runtimeutil.ResolveSelectionToCandidate(selectionIndex, remaining)
	if !ok {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	selected = append(selected, cardIdx)

	nextRemaining := make([]int, 0, len(remaining))
	for _, idx := range remaining {
		if idx != cardIdx {
			nextRemaining = append(nextRemaining, idx)
		}
	}

	if len(selected) < yValue {
		ctxData["selected_indices"] = selected
		ctxData["remaining_indices"] = nextRemaining
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}

	ctxData["selected_indices"] = selected
	if targetID, _ := ctxData["target_id"].(string); targetID != "" {
		return e.resolvePlagueDeathTouchFinal(ctxData, targetID)
	}

	userID, _ := ctxData["user_id"].(string)
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	ctxData["choice_type"] = "plague_death_touch_target"
	ctxData["target_ids"] = e.campEnemyIDs(user.Camp)
	e.State.PendingInterrupt.Context = ctxData
	e.notifyInterruptPrompt()
	return nil
}

func (e *GameEngine) handlePlagueDeathTouchTargetChoice(ctxData map[string]interface{}, selectionIndex int) error {
	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
	if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	return e.resolvePlagueDeathTouchFinal(ctxData, targetIDs[selectionIndex])
}

func (e *GameEngine) resolvePlagueDeathTouchFinal(ctxData map[string]interface{}, targetID string) error {
	userID, _ := ctxData["user_id"].(string)
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	if e.State.Players[targetID] == nil {
		return fmt.Errorf("目标不存在")
	}

	selected := append([]int{}, runtimeutil.ParseChoiceIntSlice(ctxData["selected_indices"])...)
	xValue := runtimeutil.ToIntContextValue(ctxData["x_value"])
	yValue := runtimeutil.ToIntContextValue(ctxData["y_value"])

	removed, err := removeCardsByIndicesFromHand(user, selected)
	if err != nil {
		return err
	}
	e.NotifyCardRevealed(user.ID, removed, "discard")
	e.State.DiscardPile = append(e.State.DiscardPile, removed...)

	if user.Heal < xValue {
		return fmt.Errorf("治疗不足，无法移除X=%d", xValue)
	}
	user.Heal -= xValue

	damage := xValue + yValue - 3
	if damage < 0 {
		damage = 0
	}
	user.TurnState.UsedSkillCounts["plague_block_immortal"] = 1
	user.TurnState.HasActed = true
	user.TurnState.LastActionType = string(model.ActionMagic)
	user.TurnState.LastActionCard = nil
	e.AddPendingDamage(model.PendingDamage{
		SourceID:           user.ID,
		TargetID:           targetID,
		Damage:             damage,
		DamageType:         "magic",
		CapDrawToHandLimit: true,
	})

	e.PopInterrupt()
	if e.State.PendingInterrupt == nil {
		if !e.routePendingDamageWithReturn(model.TurnStageExtraAction) {
			e.enterExtraActionStage()
		}
	}
	return nil
}
