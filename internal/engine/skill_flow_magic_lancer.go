package engine

import (
	"fmt"
	"starcup-engine/internal/engine/core/runtimeutil"
	"strings"

	"starcup-engine/internal/model"
)

// prepareMagicLancerFullnessStep 在“充盈”结算过程中推进到下一个需要选择的角色。
// 返回 true 表示所有角色都已处理完。
func (e *GameEngine) prepareMagicLancerFullnessStep(ctxData map[string]interface{}, user *model.Player) (bool, error) {
	if ctxData == nil || user == nil {
		return true, fmt.Errorf("充盈上下文无效")
	}
	orderIDs := runtimeutil.ParseStringSliceContextValue(ctxData["order_ids"])
	if len(orderIDs) == 0 {
		return true, nil
	}
	idx := runtimeutil.ToIntContextValue(ctxData["order_index"])
	if idx < 0 {
		idx = 0
	}
	for idx < len(orderIDs) {
		pid := orderIDs[idx]
		target := e.State.Players[pid]
		if target == nil {
			idx++
			continue
		}
		allowSkip := target.Camp == user.Camp
		candidates := allHandIndices(target)
		if len(candidates) == 0 {
			e.Log(fmt.Sprintf("%s 的 [充盈] 结算：%s 无手牌可弃，跳过", user.Name, target.Name))
			idx++
			continue
		}
		ctxData["order_index"] = idx
		ctxData["current_player_id"] = pid
		ctxData["allow_skip"] = allowSkip
		ctxData["candidates"] = candidates
		return false, nil
	}
	return true, nil
}

func (e *GameEngine) buildMagicLancerChoicePrompt(choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "ml_black_spear_x":
		maxX := runtimeutil.ToIntContextValue(data["max_x"])
		if maxX < 1 {
			maxX = 1
		}
		options := make([]model.PromptOption, 0, maxX)
		for x := 1; x <= maxX; x++ {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", x), Label: fmt.Sprintf("X=%d（消耗%d蓝水晶，伤害额外+%d）", x, x, x+2)})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【漆黑之枪】请选择X值：", Options: options, Min: 1, Max: 1}

	case "ml_dark_barrier_mode":
		maxMagic := runtimeutil.ToIntContextValue(data["max_magic"])
		maxThunder := runtimeutil.ToIntContextValue(data["max_thunder"])
		options := make([]model.PromptOption, 0, 2)
		if maxMagic > 0 {
			options = append(options, model.PromptOption{ID: "0", Label: "弃法术牌"})
		}
		if maxThunder > 0 {
			options = append(options, model.PromptOption{ID: "1", Label: "弃雷系牌"})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【暗之障壁】请选择本次弃牌类型：", Options: options, Min: 1, Max: 1}

	case "ml_dark_barrier_x":
		maxX := runtimeutil.ToIntContextValue(data["max_x"])
		if maxX < 1 {
			maxX = 1
		}
		options := make([]model.PromptOption, 0, maxX)
		for x := 1; x <= maxX; x++ {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", x), Label: fmt.Sprintf("弃置%d张", x)})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【暗之障壁】请选择X值：", Options: options, Min: 1, Max: 1}

	case "ml_dark_barrier_cards":
		remaining := parseIntSliceContextValue(data["remaining_indices"])
		selectedCount := len(parseIntSliceContextValue(data["selected_indices"]))
		xValue := runtimeutil.ToIntContextValue(data["x_value"])
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
		return &model.Prompt{Type: model.PromptChooseCards, PlayerID: playerID, Message: fmt.Sprintf("【暗之障壁】请选择要弃置的%d张牌：", remainingPick), Options: options, Min: remainingPick, Max: remainingPick}

	case "ml_fullness_cost_card":
		options := make([]model.PromptOption, 0, len(player.Hand))
		for idx, card := range player.Hand {
			if card.Type != model.CardTypeMagic && card.Element != model.ElementThunder {
				continue
			}
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", idx), Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(card))})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【充盈】请选择要弃置的1张法术牌或雷系牌：", Options: options, Min: 1, Max: 1}

	case "ml_fullness_discard_step":
		currentID, _ := data["current_player_id"].(string)
		target := e.State.Players[currentID]
		if target == nil {
			return nil
		}
		allowSkip, _ := data["allow_skip"].(bool)
		options := make([]model.PromptOption, 0, len(target.Hand)+1)
		if allowSkip {
			options = append(options, model.PromptOption{ID: "skip", Label: "不弃置"})
		}
		candidates := parseIntSliceContextValue(data["candidates"])
		for _, idx := range candidates {
			if idx < 0 || idx >= len(target.Hand) {
				continue
			}
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", idx), Label: fmt.Sprintf("%s 的第%d张：%s", target.Name, idx+1, formatCardInfo(target.Hand[idx]))})
		}
		msg := fmt.Sprintf("【充盈】请选择 %s 的弃牌：", target.Name)
		if allowSkip {
			msg = fmt.Sprintf("【充盈】请选择 %s 是否弃牌：", target.Name)
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: msg, Options: options, Min: 1, Max: 1}

	case "ml_stardust_target":
		targetIDs := runtimeutil.ParseStringSliceContextValue(data["target_ids"])
		options := make([]model.PromptOption, 0, len(targetIDs))
		for _, targetID := range targetIDs {
			if target := e.State.Players[targetID]; target != nil {
				options = append(options, model.PromptOption{ID: targetID, Label: target.Name})
			}
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【幻影星尘】请选择2点法术伤害目标：", Options: options, Min: 1, Max: 1}
	}

	return nil
}

func (e *GameEngine) handleMagicLancerChoiceInput(_ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	return dispatchChoiceRouteByType(choiceType, selectionIndex, ctxData, map[string]skillChoiceRouteHandler{
		"ml_black_spear_x":         e.handleMagicLancerChoiceInputByType,
		"ml_dark_barrier_mode":     e.handleMagicLancerChoiceInputByType,
		"ml_dark_barrier_x":        e.handleMagicLancerChoiceInputByType,
		"ml_dark_barrier_cards":    e.handleMagicLancerChoiceInputByType,
		"ml_fullness_cost_card":    e.handleMagicLancerChoiceInputByType,
		"ml_fullness_discard_step": e.handleMagicLancerChoiceInputByType,
		"ml_stardust_target":       e.handleMagicLancerChoiceInputByType,
	})
}

func (e *GameEngine) handleMagicLancerChoiceInputByType(selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	return dispatchChoiceRouteByType(magicLancerChoiceFlow(choiceType), selectionIndex, ctxData, map[string]skillChoiceRouteHandler{
		"black_spear":  e.handleMagicLancerBlackSpearFlow,
		"dark_barrier": e.handleMagicLancerDarkBarrierFlow,
		"fullness":     e.handleMagicLancerFullnessFlow,
		"stardust":     e.handleMagicLancerStardustFlow,
	})
}

func magicLancerChoiceFlow(choiceType string) string {
	switch choiceType {
	case "ml_black_spear_x":
		return "black_spear"
	case "ml_dark_barrier_mode", "ml_dark_barrier_x", "ml_dark_barrier_cards":
		return "dark_barrier"
	case "ml_fullness_cost_card", "ml_fullness_discard_step":
		return "fullness"
	case "ml_stardust_target":
		return "stardust"
	default:
		return ""
	}
}

func (e *GameEngine) handleMagicLancerBlackSpearFlow(selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	return dispatchChoiceRouteByType(choiceType, selectionIndex, ctxData, map[string]skillChoiceRouteHandler{
		"ml_black_spear_x": e.handleMagicLancerChoiceInputByTypeLegacy,
	})
}

func (e *GameEngine) handleMagicLancerDarkBarrierFlow(selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	return dispatchChoiceRouteByType(choiceType, selectionIndex, ctxData, map[string]skillChoiceRouteHandler{
		"ml_dark_barrier_mode":  e.handleMagicLancerChoiceInputByTypeLegacy,
		"ml_dark_barrier_x":     e.handleMagicLancerChoiceInputByTypeLegacy,
		"ml_dark_barrier_cards": e.handleMagicLancerChoiceInputByTypeLegacy,
	})
}

func (e *GameEngine) handleMagicLancerFullnessFlow(selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	return dispatchChoiceRouteByType(choiceType, selectionIndex, ctxData, map[string]skillChoiceRouteHandler{
		"ml_fullness_cost_card":    e.handleMagicLancerChoiceInputByTypeLegacy,
		"ml_fullness_discard_step": e.handleMagicLancerChoiceInputByTypeLegacy,
	})
}

func (e *GameEngine) handleMagicLancerStardustFlow(selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	return dispatchChoiceRouteByType(choiceType, selectionIndex, ctxData, map[string]skillChoiceRouteHandler{
		"ml_stardust_target": e.handleMagicLancerChoiceInputByTypeLegacy,
	})
}

func (e *GameEngine) handleMagicLancerChoiceInputByTypeLegacy(selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "ml_black_spear_x":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
		xValue := selectionIndex + 1
		if xValue < 1 || xValue > maxX {
			return true, fmt.Errorf("无效的X值")
		}
		if !e.ConsumeCrystalCost(user.ID, xValue) {
			return true, fmt.Errorf("漆黑之枪需要%d点蓝水晶（红宝石可替代）", xValue)
		}
		targetID, _ := ctxData["target_id"].(string)
		bonus := xValue + 2
		applied := false
		for idx := range e.State.PendingDamageQueue {
			pd := &e.State.PendingDamageQueue[idx]
			if !strings.EqualFold(string(pd.DamageType), string(model.AttackDamage)) {
				continue
			}
			if pd.SourceID != user.ID {
				continue
			}
			if targetID != "" && pd.TargetID != targetID {
				continue
			}
			pd.Damage += bonus
			applied = true
			break
		}
		if !applied {
			e.Log("[Warn] 漆黑之枪未找到可叠加的攻击伤害条目")
		}
		e.Log(fmt.Sprintf("%s 的 [漆黑之枪] 生效：消耗%d点蓝水晶，本次攻击伤害额外+%d", user.Name, xValue, bonus))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.enterDamageResolution(nil)
		}
		return true, nil

	case "ml_dark_barrier_mode":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		maxMagic := runtimeutil.ToIntContextValue(ctxData["max_magic"])
		maxThunder := runtimeutil.ToIntContextValue(ctxData["max_thunder"])
		modes := make([]string, 0, 2)
		if maxMagic > 0 {
			modes = append(modes, "magic")
		}
		if maxThunder > 0 {
			modes = append(modes, "thunder")
		}
		if selectionIndex < 0 || selectionIndex >= len(modes) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		mode := modes[selectionIndex]
		maxX := maxMagic
		if mode == "thunder" {
			maxX = maxThunder
		}
		if maxX <= 0 {
			return true, fmt.Errorf("可弃牌数量不足")
		}
		ctxData["mode"] = mode
		ctxData["max_x"] = maxX
		ctxData["choice_type"] = "ml_dark_barrier_x"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil

	case "ml_dark_barrier_x":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		mode, _ := ctxData["mode"].(string)
		maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
		xValue := selectionIndex + 1
		if xValue < 1 || xValue > maxX {
			return true, fmt.Errorf("无效的X值")
		}
		remaining := make([]int, 0)
		for idx, card := range user.Hand {
			if mode == "magic" {
				if card.Type == model.CardTypeMagic {
					remaining = append(remaining, idx)
				}
			} else if mode == "thunder" && card.Element == model.ElementThunder {
				remaining = append(remaining, idx)
			}
		}
		if len(remaining) < xValue {
			return true, fmt.Errorf("可选弃牌不足，无法选择X=%d", xValue)
		}
		ctxData["x_value"] = xValue
		ctxData["selected_indices"] = []int{}
		ctxData["remaining_indices"] = remaining
		ctxData["choice_type"] = "ml_dark_barrier_cards"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil

	case "ml_dark_barrier_cards":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		mode, _ := ctxData["mode"].(string)
		xValue := runtimeutil.ToIntContextValue(ctxData["x_value"])
		remaining := parseIntSliceContextValue(ctxData["remaining_indices"])
		selected := parseIntSliceContextValue(ctxData["selected_indices"])
		cardIdx, ok := runtimeutil.ResolveSelectionToCandidate(selectionIndex, remaining)
		if !ok || cardIdx < 0 || cardIdx >= len(user.Hand) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		card := user.Hand[cardIdx]
		if mode == "magic" && card.Type != model.CardTypeMagic {
			return true, fmt.Errorf("暗之障壁当前需要弃法术牌")
		}
		if mode == "thunder" && card.Element != model.ElementThunder {
			return true, fmt.Errorf("暗之障壁当前需要弃雷系牌")
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
			return true, nil
		}
		removed, err := removeCardsByIndicesFromHand(user, append([]int{}, selected...))
		if err != nil {
			return true, err
		}
		e.NotifyCardRevealed(user.ID, removed, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, removed...)
		e.Log(fmt.Sprintf("%s 的 [暗之障壁] 生效：弃置%d张%s牌", user.Name, xValue, map[string]string{"magic": "法术", "thunder": "雷系"}[mode]))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if len(e.State.PendingDamageQueue) > 0 {
				e.enterDamageResolution(nil)
			} else if len(e.State.ActionStack) > 0 {
				e.enterResponseWindow()
			}
		}
		return true, nil

	case "ml_fullness_cost_card":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		candidates := make([]int, 0)
		for idx, card := range user.Hand {
			if card.Type == model.CardTypeMagic || card.Element == model.ElementThunder {
				candidates = append(candidates, idx)
			}
		}
		cardIdx, ok := runtimeutil.ResolveSelectionToCandidate(selectionIndex, candidates)
		if !ok || cardIdx < 0 || cardIdx >= len(user.Hand) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		costCard := user.Hand[cardIdx]
		user.Hand = append(user.Hand[:cardIdx], user.Hand[cardIdx+1:]...)
		e.NotifyCardRevealed(user.ID, []model.Card{costCard}, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, costCard)

		orderIDs := make([]string, 0)
		for _, playerID := range e.reverseOrderTargetIDsFrom(user.ID, false) {
			target := e.State.Players[playerID]
			if target == nil || target.Camp == user.Camp {
				continue
			}
			orderIDs = append(orderIDs, playerID)
		}
		lockedAllyID, _ := ctxData["locked_ally_id"].(string)
		if lockedAllyID != "" {
			target := e.State.Players[lockedAllyID]
			if target != nil && target.Camp == user.Camp && target.ID != user.ID {
				orderIDs = append(orderIDs, lockedAllyID)
			}
		}
		ctxData["order_ids"] = orderIDs
		ctxData["order_index"] = 0
		ctxData["bonus"] = 0
		ctxData["choice_type"] = "ml_fullness_discard_step"

		done, err := e.prepareMagicLancerFullnessStep(ctxData, user)
		if err != nil {
			return true, err
		}
		if done {
			if user.TurnState.UsedSkillCounts == nil {
				user.TurnState.UsedSkillCounts = map[string]int{}
			}
			model.AppendAttackAction(user, "充盈")
			e.Log(fmt.Sprintf("%s 的 [充盈] 生效：无可处理弃牌目标，获得额外1次攻击行动", user.Name))
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				e.enterExtraActionStage()
			}
			return true, nil
		}
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil

	case "ml_fullness_discard_step":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		currentID, _ := ctxData["current_player_id"].(string)
		target := e.State.Players[currentID]
		if target == nil {
			return true, fmt.Errorf("弃牌目标不存在")
		}
		allowSkip, _ := ctxData["allow_skip"].(bool)
		candidates := parseIntSliceContextValue(ctxData["candidates"])
		if len(candidates) == 0 {
			allowSkip = true
		}
		skipped := false
		chosenCard := model.Card{}
		if allowSkip && selectionIndex == 0 {
			skipped = true
		} else {
			optionIdx := selectionIndex
			if allowSkip {
				optionIdx--
			}
			cardIdx, ok := runtimeutil.ResolveSelectionToCandidate(optionIdx, candidates)
			if !ok || cardIdx < 0 || cardIdx >= len(target.Hand) {
				return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
			}
			chosenCard = target.Hand[cardIdx]
			target.Hand = append(target.Hand[:cardIdx], target.Hand[cardIdx+1:]...)
			e.NotifyCardRevealed(target.ID, []model.Card{chosenCard}, "discard")
			e.State.DiscardPile = append(e.State.DiscardPile, chosenCard)
		}
		if skipped {
			e.Log(fmt.Sprintf("%s 的 [充盈]：%s 选择不弃牌", user.Name, target.Name))
		} else {
			e.Log(fmt.Sprintf("%s 的 [充盈]：%s 弃置了 %s", user.Name, target.Name, chosenCard.Name))
			if target.ID != user.ID && (chosenCard.Type == model.CardTypeMagic || chosenCard.Element == model.ElementThunder) {
				ctxData["bonus"] = runtimeutil.ToIntContextValue(ctxData["bonus"]) + 1
			}
		}

		ctxData["order_index"] = runtimeutil.ToIntContextValue(ctxData["order_index"]) + 1
		done, err := e.prepareMagicLancerFullnessStep(ctxData, user)
		if err != nil {
			return true, err
		}
		if !done {
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return true, nil
		}

		bonus := runtimeutil.ToIntContextValue(ctxData["bonus"])
		if bonus > 0 {
			e.ApplyNextAttackDamageRule(user.ID, "ml_fullness_next_attack_bonus", "ml_fullness", bonus, model.RuleLifeUntilTurnEnd)
		}
		model.AppendAttackAction(user, "充盈")
		e.Log(fmt.Sprintf("%s 的 [充盈] 结算完成：本回合下次主动攻击伤害额外+%d，额外获得1次攻击行动", user.Name, bonus))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if !e.routePendingDamageWithReturn(model.TurnStageExtraAction) {
				e.enterExtraActionStage()
			}
		}
		return true, nil

	case "ml_stardust_target":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		targetID := targetIDs[selectionIndex]
		target := e.State.Players[targetID]
		if target == nil {
			return true, fmt.Errorf("目标不存在")
		}
		e.AddPendingDamage(model.PendingDamage{SourceID: user.ID, TargetID: targetID, Damage: 2, DamageType: model.MagicAttack})
		e.Log(fmt.Sprintf("%s 的 [幻影星尘] 生效：对 %s 造成2点法术伤害", user.Name, target.Name))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if !e.routePendingDamageWithDefaultReturn(nil) {
				e.setTurnStage(model.TurnStageActionStart)
				e.clearCombatStage()
				e.clearSubflow()
			}
		}
		return true, nil
	}

	return false, nil
}
