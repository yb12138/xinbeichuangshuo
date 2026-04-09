package engine

import (
	"fmt"
	"starcup-engine/internal/engine/core/runtimeutil"

	"starcup-engine/internal/model"
)

func (e *GameEngine) blazeWitchAttackElement(player *model.Player, card model.Card) model.Element {
	if player == nil || player.Tokens == nil {
		return card.Element
	}
	if !e.isBlazeWitch(player) || !hasBlazeWitchFlameForm(player) {
		return card.Element
	}
	if card.Type != model.CardTypeAttack {
		return card.Element
	}
	if card.Element == model.ElementWater || card.Element == model.ElementDark {
		return card.Element
	}
	return model.ElementFire
}

func (e *GameEngine) applyBlazeWitchAttackCardTransform(player *model.Player, card model.Card) model.Card {
	card.Element = e.blazeWitchAttackElement(player, card)
	return card
}

func (e *GameEngine) buildBlazeWitchChoicePrompt(choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "bw_witch_wrath_draw":
		return &model.Prompt{
			Type:       model.PromptConfirm,
			PlayerID:   playerID,
			Message:    "【魔女之怒】请选择摸牌数量：",
			ChoiceType: choiceType,
			Options: []model.PromptOption{
				{ID: "0", Label: "摸0张"},
				{ID: "1", Label: "摸1张"},
				{ID: "2", Label: "摸2张"},
			},
			Min: 1,
			Max: 1,
		}
	case "bw_substitute_doll_card":
		magicIndices := runtimeutil.ParseChoiceIntSlice(data["magic_indices"])
		options := make([]model.PromptOption, 0, len(magicIndices))
		for _, idx := range magicIndices {
			if player == nil || idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(player.Hand[idx])),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【替身玩偶】请选择弃置1张法术牌：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	case "bw_mana_inversion_x":
		maxX := runtimeutil.ToIntContextValue(data["max_x"])
		options := make([]model.PromptOption, 0, maxX-1)
		for x := 2; x <= maxX; x++ {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", x-2),
				Label: fmt.Sprintf("X=%d（弃%d张法术牌，造成%d点法术伤害）", x, x, x-1),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【魔能反转】请选择X值：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	case "bw_mana_inversion_cards":
		remaining := runtimeutil.ParseChoiceIntSlice(data["remaining_indices"])
		selectedCount := len(runtimeutil.ParseChoiceIntSlice(data["selected_indices"]))
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
		return &model.Prompt{
			Type:     model.PromptChooseCards,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【魔能反转】请选择要弃置的%d张法术牌：", remainingPick),
			Options:  options,
			Min:      remainingPick,
			Max:      remainingPick,
		}
	case "bw_substitute_doll_target":
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【替身玩偶】请选择摸1张牌的队友：",
			Options:  buildPromptOptionsForPlayerIDs(e.State.Players, runtimeutil.ParseStringSliceContextValue(data["target_ids"])),
			Min:      1,
			Max:      1,
		}
	case "bw_mana_inversion_target":
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【魔能反转】请选择法术伤害目标：",
			Options:  buildPromptOptionsForPlayerIDs(e.State.Players, runtimeutil.ParseStringSliceContextValue(data["target_ids"])),
			Min:      1,
			Max:      1,
		}
	default:
		return nil
	}
}

func (e *GameEngine) handleBlazeWitchChoiceInput(_ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	return dispatchChoiceInputByType(choiceType, selectionIndex, ctxData, map[string]skillChoiceInputHandler{
		"bw_witch_wrath_draw": func(idx int, data map[string]interface{}) error {
			return e.handleBlazeWitchWrathDrawChoice(data, idx)
		},
		"bw_substitute_doll_card": func(idx int, data map[string]interface{}) error {
			return e.handleBlazeWitchSubstituteCardChoice(data, idx)
		},
		"bw_mana_inversion_x": func(idx int, data map[string]interface{}) error {
			return e.handleBlazeWitchManaInversionXChoice(data, idx)
		},
		"bw_mana_inversion_cards": func(idx int, data map[string]interface{}) error {
			return e.handleBlazeWitchManaInversionCardsChoice(data, idx)
		},
		"bw_substitute_doll_target": func(idx int, data map[string]interface{}) error {
			return e.handleBlazeWitchTargetChoice(data, idx)
		},
		"bw_mana_inversion_target": func(idx int, data map[string]interface{}) error {
			return e.handleBlazeWitchTargetChoice(data, idx)
		},
	})
}

func (e *GameEngine) handleBlazeWitchWrathDrawChoice(ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	if selectionIndex < 0 || selectionIndex > 2 {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	if selectionIndex > 0 {
		e.DrawCards(user.ID, selectionIndex)
	}
	e.Log(fmt.Sprintf("%s 的 [魔女之怒]：选择摸%d张牌", user.Name, selectionIndex))
	e.PopInterrupt()
	if e.State.PendingInterrupt == nil {
		e.setTurnStage(model.TurnStageActionStart)
		e.clearCombatStage()
		e.clearSubflow()
	}
	return nil
}

func (e *GameEngine) handleBlazeWitchSubstituteCardChoice(ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	magicIndices := runtimeutil.ParseChoiceIntSlice(ctxData["magic_indices"])
	cardIdx, ok := runtimeutil.ResolveSelectionToCandidate(selectionIndex, magicIndices)
	if !ok || cardIdx < 0 || cardIdx >= len(user.Hand) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	if user.Hand[cardIdx].Type != model.CardTypeMagic {
		return fmt.Errorf("替身玩偶需弃置法术牌")
	}
	ctxData["selected_card_index"] = cardIdx
	ctxData["choice_type"] = "bw_substitute_doll_target"
	ctxData["target_ids"] = runtimeutil.ParseStringSliceContextValue(ctxData["ally_ids"])
	e.State.PendingInterrupt.Context = ctxData
	e.notifyInterruptPrompt()
	return nil
}

func (e *GameEngine) handleBlazeWitchManaInversionXChoice(ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
	xValue := selectionIndex + 2
	if xValue < 2 || xValue > maxX {
		return fmt.Errorf("无效的X值")
	}
	magicIndices := make([]int, 0, len(user.Hand))
	for i, card := range user.Hand {
		if card.Type == model.CardTypeMagic {
			magicIndices = append(magicIndices, i)
		}
	}
	if len(magicIndices) < xValue {
		return fmt.Errorf("法术牌不足，无法弃置X=%d张", xValue)
	}
	ctxData["choice_type"] = "bw_mana_inversion_cards"
	ctxData["x_value"] = xValue
	ctxData["selected_indices"] = []int{}
	ctxData["remaining_indices"] = magicIndices
	e.State.PendingInterrupt.Context = ctxData
	e.notifyInterruptPrompt()
	return nil
}

func (e *GameEngine) handleBlazeWitchManaInversionCardsChoice(ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	remaining := runtimeutil.ParseChoiceIntSlice(ctxData["remaining_indices"])
	selected := append([]int{}, runtimeutil.ParseChoiceIntSlice(ctxData["selected_indices"])...)
	xValue := runtimeutil.ToIntContextValue(ctxData["x_value"])

	cardIdx, ok := runtimeutil.ResolveSelectionToCandidate(selectionIndex, remaining)
	if !ok || cardIdx < 0 || cardIdx >= len(user.Hand) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	if user.Hand[cardIdx].Type != model.CardTypeMagic {
		return fmt.Errorf("魔能反转需弃置法术牌")
	}
	selected = append(selected, cardIdx)
	nextRemaining := make([]int, 0, len(remaining))
	for _, idx := range remaining {
		if idx != cardIdx {
			nextRemaining = append(nextRemaining, idx)
		}
	}
	if len(selected) < xValue {
		ctxData["selected_indices"] = selected
		ctxData["remaining_indices"] = nextRemaining
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return nil
	}

	enemyIDs := make([]string, 0)
	for _, playerID := range e.State.PlayerOrder {
		target := e.State.Players[playerID]
		if target == nil || target.Camp == user.Camp {
			continue
		}
		enemyIDs = append(enemyIDs, target.ID)
	}
	if len(enemyIDs) == 0 {
		return fmt.Errorf("无可选敌方目标")
	}
	ctxData["selected_indices"] = selected
	ctxData["choice_type"] = "bw_mana_inversion_target"
	ctxData["target_ids"] = enemyIDs
	e.State.PendingInterrupt.Context = ctxData
	e.notifyInterruptPrompt()
	return nil
}

func (e *GameEngine) handleBlazeWitchTargetChoice(ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
	if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	targetID := targetIDs[selectionIndex]
	target := e.State.Players[targetID]
	if target == nil {
		return fmt.Errorf("目标不存在")
	}

	switch choiceType, _ := ctxData["choice_type"].(string); choiceType {
	case "bw_substitute_doll_target":
		cardIdx := runtimeutil.ToIntContextValue(ctxData["selected_card_index"])
		if cardIdx < 0 || cardIdx >= len(user.Hand) {
			return fmt.Errorf("无效的弃牌索引")
		}
		if user.Hand[cardIdx].Type != model.CardTypeMagic {
			return fmt.Errorf("替身玩偶需弃置法术牌")
		}
		card := user.Hand[cardIdx]
		user.Hand = append(user.Hand[:cardIdx], user.Hand[cardIdx+1:]...)
		e.NotifyCardRevealed(user.ID, []model.Card{card}, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, card)
		e.DrawCards(targetID, 1)
		e.Log(fmt.Sprintf("%s 的 [替身玩偶] 生效：%s 摸1张牌", user.Name, target.Name))
	case "bw_mana_inversion_target":
		selected := append([]int{}, runtimeutil.ParseChoiceIntSlice(ctxData["selected_indices"])...)
		xValue := runtimeutil.ToIntContextValue(ctxData["x_value"])
		if xValue < 2 || len(selected) != xValue {
			return fmt.Errorf("魔能反转弃牌参数错误")
		}
		for _, idx := range selected {
			if idx < 0 || idx >= len(user.Hand) || user.Hand[idx].Type != model.CardTypeMagic {
				return fmt.Errorf("魔能反转弃牌必须为法术牌")
			}
		}
		removed, err := removeCardsByIndicesFromHand(user, selected)
		if err != nil {
			return err
		}
		e.NotifyCardRevealed(user.ID, removed, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, removed...)
		damage := xValue - 1
		if damage > 0 {
			e.AddPendingDamage(model.PendingDamage{
				SourceID:   user.ID,
				TargetID:   targetID,
				Damage:     damage,
				DamageType: model.MagicAttack,
			})
		}
		e.Log(fmt.Sprintf("%s 的 [魔能反转] 生效：弃%d张法术牌，对 %s 造成%d点法术伤害", user.Name, xValue, target.Name, damage))
	}

	e.PopInterrupt()
	if e.State.PendingInterrupt == nil && len(e.State.PendingDamageQueue) > 0 {
		e.enterDamageResolution(nil)
	}
	return nil
}
