package engine

import (
	"fmt"
	"starcup-engine/internal/engine/runtimeutil"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func (e *GameEngine) buildMagicBowChoicePrompt(choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "mb_charge_draw_x":
		maxDraw := runtimeutil.ToIntContextValue(data["max_draw"])
		if maxDraw <= 0 {
			maxDraw = 4
		}
		options := make([]model.PromptOption, 0, maxDraw+1)
		for x := 0; x <= maxDraw; x++ {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", x), Label: fmt.Sprintf("X=%d（摸%d张）", x, x)})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【充能】请选择摸牌数量X（0~4）：", Options: options, Min: 1, Max: 1}

	case "mb_charge_place_count":
		maxPlace := runtimeutil.ToIntContextValue(data["max_place"])
		if maxPlace < 0 {
			maxPlace = 0
		}
		options := make([]model.PromptOption, 0, maxPlace+1)
		for count := 0; count <= maxPlace; count++ {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", count), Label: fmt.Sprintf("放置%d张充能", count)})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【充能】请选择要放置为充能的手牌数量：", Options: options, Min: 1, Max: 1}

	case "mb_charge_place_cards", "mb_demon_eye_charge_card":
		if player == nil {
			return nil
		}
		remaining := parseIntSliceContextValue(data["remaining_indices"])
		if len(remaining) == 0 && choiceType == "mb_demon_eye_charge_card" {
			remaining = allHandIndices(player)
		}
		selectedCount := len(parseIntSliceContextValue(data["selected_indices"]))
		needCount := runtimeutil.ToIntContextValue(data["need_count"])
		if choiceType == "mb_demon_eye_charge_card" && needCount <= 0 {
			needCount = 1
		}
		if needCount <= 0 {
			needCount = 1
		}
		options := make([]model.PromptOption, 0, len(remaining))
		for _, idx := range remaining {
			if idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", idx), Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(player.Hand[idx]))})
		}
		remainingPick := needCount - selectedCount
		if remainingPick < 1 {
			remainingPick = 1
		}
		if len(options) > 0 && remainingPick > len(options) {
			remainingPick = len(options)
		}
		message := fmt.Sprintf("【充能】请选择%d张作为充能的手牌：", remainingPick)
		if choiceType == "mb_demon_eye_charge_card" {
			message = "【魔眼】请选择1张手牌作为充能："
		}
		return &model.Prompt{Type: model.PromptChooseCards, PlayerID: playerID, Message: message, Options: options, Min: remainingPick, Max: remainingPick}

	case "mb_thunder_scatter_extra":
		maxExtra := runtimeutil.ToIntContextValue(data["max_extra"])
		if maxExtra < 0 {
			maxExtra = 0
		}
		options := make([]model.PromptOption, 0, maxExtra+1)
		for x := 0; x <= maxExtra; x++ {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", x), Label: fmt.Sprintf("额外移除%d个雷系充能", x)})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【雷光散射】请选择额外移除雷系充能数量X：", Options: options, Min: 1, Max: 1}
	}

	return nil
}

func (e *GameEngine) handleMagicBowChoiceInput(_ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)

	switch choiceType {
	case "mb_charge_draw_x":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		maxDraw := runtimeutil.ToIntContextValue(ctxData["max_draw"])
		if maxDraw <= 0 {
			maxDraw = 4
		}
		xValue := selectionIndex
		if xValue < 0 || xValue > maxDraw {
			return true, fmt.Errorf("无效的X值")
		}
		if xValue > 0 {
			cards, newDeck, newDiscard := rules.DrawCards(e.State.Deck, e.State.DiscardPile, xValue)
			e.State.Deck = newDeck
			e.State.DiscardPile = newDiscard
			user.Hand = append(user.Hand, cards...)
			e.NotifyDrawCards(user.ID, len(cards), "mb_charge")
		}
		room := magicBowChargeCapEngine - magicBowChargeCount(user, "")
		maxPlace := xValue
		if maxPlace > len(user.Hand) {
			maxPlace = len(user.Hand)
		}
		if maxPlace > room {
			maxPlace = room
		}

		overflow := len(user.Hand) - e.GetMaxHand(user)
		if overflow > 0 {
			moraleLoss := overflow
			allowedByFloor := e.campMorale(user.Camp) - e.moraleFloorForCamp(user.Camp)
			if allowedByFloor < 0 {
				allowedByFloor = 0
			}
			if moraleLoss > allowedByFloor {
				moraleLoss = allowedByFloor
			}
			if moraleLoss > 0 {
				lossEventCtx := &model.EventContext{Type: model.EventDamage, DamageVal: &moraleLoss}
				lossCtx := e.buildContext(user, nil, model.TriggerBeforeMoraleLoss, lossEventCtx)
				lossCtx.Flags["IsMagicDamage"] = false
				if lossCtx.Selections == nil {
					lossCtx.Selections = map[string]any{}
				}
				lossCtx.Selections["discarded_cards"] = []model.Card{}
				lossCtx.Selections["from_damage_draw"] = false
				lossCtx.Selections["victim_id"] = user.ID
				lossCtx.Selections["discard_player_id"] = user.ID
				lossCtx.Selections["morale_loss_stay_in_turn"] = true
				lossCtx.Selections["morale_loss_is_damage_resolution"] = false
				lossCtx.Selections["mb_charge_resume"] = true
				lossCtx.Selections["mb_charge_user_id"] = user.ID
				lossCtx.Selections["mb_charge_max_place"] = maxPlace
				e.dispatcher.OnTrigger(model.TriggerBeforeMoraleLoss, lossCtx)

				pendingResponse := false
				for _, intr := range e.State.InterruptQueue {
					if intr != nil && intr.Type == model.InterruptResponseSkill {
						pendingResponse = true
						break
					}
				}
				if pendingResponse {
					lossCtx.Selections["morale_loss_pending"] = true
					lossCtx.Selections["morale_loss_value"] = moraleLoss
					lossCtx.Selections["is_magic"] = false
					lossCtx.Selections["overflow_morale_loss_fixed"] = 0
					e.PopInterrupt()
					return true, nil
				}

				finalLoss := e.applyMoraleLossAfterTrigger(user, moraleLoss, false, false, 0, []model.Card{}, lossCtx)
				e.Log(fmt.Sprintf("%s 的 [充能] 摸牌后超出手牌上限%d：士气-%d（本次不弃牌）", user.Name, overflow, finalLoss))
				e.checkGameEnd()
			}
		}

		if maxPlace <= 0 {
			e.Log(fmt.Sprintf("%s 的 [充能] 生效：摸%d张，不放置充能", user.Name, xValue))
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				e.setTurnStage(model.TurnStageActionStart)
				e.clearCombatStage()
				e.clearSubflow()
			}
			return true, nil
		}
		ctxData["choice_type"] = "mb_charge_place_count"
		ctxData["max_place"] = maxPlace
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		e.Log(fmt.Sprintf("%s 的 [充能] 生效：摸%d张，可放置最多%d张充能", user.Name, xValue, maxPlace))
		return true, nil

	case "mb_charge_place_count":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		maxPlace := runtimeutil.ToIntContextValue(ctxData["max_place"])
		if maxPlace < 0 {
			maxPlace = 0
		}
		needCount := selectionIndex
		if needCount < 0 || needCount > maxPlace {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		if needCount == 0 {
			e.Log(fmt.Sprintf("%s 选择不放置充能", user.Name))
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				e.setTurnStage(model.TurnStageActionStart)
				e.clearCombatStage()
				e.clearSubflow()
			}
			return true, nil
		}
		ctxData["choice_type"] = "mb_charge_place_cards"
		ctxData["need_count"] = needCount
		ctxData["selected_indices"] = []int{}
		ctxData["remaining_indices"] = allHandIndices(user)
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil

	case "mb_charge_place_cards", "mb_demon_eye_charge_card":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		remaining := parseIntSliceContextValue(ctxData["remaining_indices"])
		if len(remaining) == 0 {
			remaining = allHandIndices(user)
		}
		selected := parseIntSliceContextValue(ctxData["selected_indices"])
		needCount := runtimeutil.ToIntContextValue(ctxData["need_count"])
		if choiceType == "mb_demon_eye_charge_card" && needCount <= 0 {
			needCount = 1
		}
		if needCount <= 0 {
			needCount = 1
		}
		cardIdx, ok := runtimeutil.ResolveSelectionToCandidate(selectionIndex, remaining)
		if !ok || cardIdx < 0 || cardIdx >= len(user.Hand) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		selected = append(selected, cardIdx)
		nextRemaining := make([]int, 0, len(remaining))
		for _, idx := range remaining {
			if idx != cardIdx {
				nextRemaining = append(nextRemaining, idx)
			}
		}
		if len(selected) < needCount {
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
		added := addMagicBowChargeCards(user, removed)
		if added < len(removed) {
			e.State.DiscardPile = append(e.State.DiscardPile, removed[added:]...)
		}
		if choiceType == "mb_demon_eye_charge_card" {
			maxEnergy := e.getPlayerEnergyCap(user)
			if user.Gem+user.Crystal < maxEnergy {
				user.Crystal++
				if user.Gem+user.Crystal > maxEnergy {
					user.Crystal -= user.Gem + user.Crystal - maxEnergy
				}
			}
			e.Log(fmt.Sprintf("%s 的 [魔眼] 生效：放置1张充能并获得1点蓝水晶", user.Name))
		} else {
			e.Log(fmt.Sprintf("%s 的 [充能] 生效：放置%d张充能", user.Name, added))
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.setTurnStage(model.TurnStageActionStart)
			e.clearCombatStage()
			e.clearSubflow()
		}
		return true, nil

	case "mb_thunder_scatter_extra":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
		if len(targetIDs) == 0 {
			return true, fmt.Errorf("雷光散射没有可选目标")
		}
		maxExtra := runtimeutil.ToIntContextValue(ctxData["max_extra"])
		extraX := selectionIndex
		if extraX < 0 || extraX > maxExtra {
			return true, fmt.Errorf("无效的X值")
		}
		for _, targetID := range targetIDs {
			e.AddPendingDamage(model.PendingDamage{SourceID: user.ID, TargetID: targetID, Damage: 1, DamageType: "magic"})
		}
		actualExtra := 0
		for i := 0; i < extraX; i++ {
			if _, ok := removeMagicBowChargeByElement(user, model.ElementThunder); !ok {
				break
			}
			actualExtra++
		}
		if actualExtra <= 0 {
			e.Log(fmt.Sprintf("%s 的 [雷光散射] 生效：对所有对手各造成1点法术伤害", user.Name))
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil && len(e.State.PendingDamageQueue) > 0 {
				e.setReturnPoint(model.TurnStageExtraAction)
				e.enterDamageResolution(nil)
			}
			return true, nil
		}
		lockedTargetID, _ := ctxData["locked_target_id"].(string)
		if lockedTargetID != "" {
			lockedValid := false
			for _, targetID := range targetIDs {
				if targetID == lockedTargetID {
					lockedValid = true
					break
				}
			}
			if !lockedValid {
				return true, fmt.Errorf("雷光散射预选目标无效")
			}
			target := e.State.Players[lockedTargetID]
			if target == nil {
				return true, fmt.Errorf("目标不存在")
			}
			e.AddPendingDamage(model.PendingDamage{SourceID: user.ID, TargetID: lockedTargetID, Damage: actualExtra, DamageType: "magic"})
			e.Log(fmt.Sprintf("%s 的 [雷光散射] 生效：对所有对手各1点，并对 %s 额外造成%d点法术伤害", user.Name, target.Name, actualExtra))
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil && len(e.State.PendingDamageQueue) > 0 {
				e.setReturnPoint(model.TurnStageExtraAction)
				e.enterDamageResolution(nil)
			}
			return true, nil
		}
		ctxData["choice_type"] = "mb_thunder_scatter_target"
		ctxData["extra_x"] = actualExtra
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil

	case "mb_thunder_scatter_target", "mb_multi_shot_target", "mb_demon_eye_target":
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
		switch choiceType {
		case "mb_thunder_scatter_target":
			extraX := runtimeutil.ToIntContextValue(ctxData["extra_x"])
			if extraX > 0 {
				e.AddPendingDamage(model.PendingDamage{SourceID: user.ID, TargetID: targetID, Damage: extraX, DamageType: "magic"})
			}
			e.Log(fmt.Sprintf("%s 的 [雷光散射] 生效：对所有对手各1点，并对 %s 额外造成%d点法术伤害", user.Name, target.Name, extraX))
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil && len(e.State.PendingDamageQueue) > 0 {
				e.setReturnPoint(model.TurnStageExtraAction)
				e.enterDamageResolution(nil)
			}
			return true, nil
		case "mb_multi_shot_target":
			prevOrder := user.TurnState.UsedSkillCounts["mb_last_attack_target_order"]
			if prevOrder > 0 {
				for idx, playerID := range e.State.PlayerOrder {
					if idx+1 == prevOrder && playerID == targetID {
						return true, fmt.Errorf("多重射击不能选择上次攻击目标")
					}
				}
			}
			virtualCard := model.Card{ID: fmt.Sprintf("mb_multi_shot_%s_%d", user.ID, len(e.State.DiscardPile)+len(e.State.ActionQueue)+1), Name: "多重射击", Type: model.CardTypeAttack, Element: model.ElementDark, Damage: 1, Description: "由多重射击视为的暗系主动攻击（伤害-1）"}
			e.State.ActionQueue = append(e.State.ActionQueue, model.QueuedAction{SourceID: user.ID, TargetID: target.ID, Type: model.ActionAttack, Element: model.ElementDark, Card: &virtualCard, CardIndex: -1, SourceSkill: "mb_multi_shot", UsesVirtualCard: true})
			e.Log(fmt.Sprintf("%s 的 [多重射击] 生效：对 %s 发起1次暗系追加攻击（伤害-1）", user.Name, target.Name))
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				e.enterActionExecutionStage()
			}
			return true, nil
		case "mb_demon_eye_target":
			if targetID == user.ID {
				return true, fmt.Errorf("魔眼不能以自己为目标")
			}
			if len(target.Hand) > 0 {
				e.State.PendingInterrupt.Type = model.InterruptDiscard
				e.State.PendingInterrupt.PlayerID = targetID
				e.State.PendingInterrupt.Context = map[string]interface{}{"discard_count": 1, "prompt": "【魔眼】请选择弃置1张手牌：", "mb_demon_eye_user_id": user.ID, "mb_demon_eye_target_id": targetID}
				e.Log(fmt.Sprintf("%s 的 [魔眼] 生效：请选择 %s 弃置1张手牌", user.Name, target.Name))
				e.notifyInterruptPrompt()
				return true, nil
			}
			e.DrawCards(user.ID, 3)
			ctxData["choice_type"] = "mb_demon_eye_charge_card"
			ctxData["need_count"] = 1
			ctxData["selected_indices"] = []int{}
			ctxData["remaining_indices"] = allHandIndices(user)
			e.notifyInterruptPrompt()
			e.Log(fmt.Sprintf("%s 的 [魔眼] 生效：%s 无法弃牌，改为自己摸3张牌并选择1张作为充能", user.Name, target.Name))
			return true, nil
		}
	}

	return false, nil
}
