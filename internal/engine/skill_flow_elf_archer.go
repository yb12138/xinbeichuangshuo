package engine

import (
	"fmt"
	"starcup-engine/internal/engine/runtimeutil"

	"starcup-engine/internal/model"
)

func (e *GameEngine) buildElfArcherChoicePrompt(choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "elf_elemental_shot_cost":
		canMagic, _ := data["can_discard_magic"].(bool)
		canBless, _ := data["can_remove_bless"].(bool)
		options := make([]model.PromptOption, 0, 2)
		if canMagic {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "弃1张法术牌发动"})
		}
		if canBless {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "移除1个祝福发动"})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【元素射击】请选择发动消耗：", Options: options, Min: 1, Max: 1}

	case "elf_elemental_shot_discard_magic":
		if player == nil {
			return nil
		}
		idxs := parseIntSliceContextValue(data["magic_indices"])
		options := make([]model.PromptOption, 0, len(idxs))
		for _, idx := range idxs {
			if idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", idx), Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(player.Hand[idx]))})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【元素射击】请选择弃置的法术牌：", Options: options, Min: 1, Max: 1}

	case "elf_elemental_shot_remove_blessing":
		if player == nil {
			return nil
		}
		idxs := parseIntSliceContextValue(data["blessing_indices"])
		options := make([]model.PromptOption, 0, len(idxs))
		for _, idx := range idxs {
			if idx < 0 || idx >= len(player.Blessings) {
				continue
			}
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", idx), Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(player.Blessings[idx]))})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【元素射击】请选择要移除的祝福：", Options: options, Min: 1, Max: 1}

	case "elf_animal_companion_confirm":
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【动物伙伴】是否发动（摸1弃1）？", Options: []model.PromptOption{{ID: "0", Label: "是"}, {ID: "1", Label: "否"}}, Min: 1, Max: 1}

	case "elf_pet_empower_confirm":
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【宠物强化】是否消耗1蓝水晶，将效果改为任意角色摸1弃1？", Options: []model.PromptOption{{ID: "0", Label: "是"}, {ID: "1", Label: "否"}}, Min: 1, Max: 1}
	}

	return nil
}

func (e *GameEngine) handleElfArcherChoiceInput(_ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)

	switch choiceType {
	case "elf_elemental_shot_cost":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		canMagic, _ := ctxData["can_discard_magic"].(bool)
		canBless, _ := ctxData["can_remove_bless"].(bool)
		modeList := make([]int, 0, 2)
		if canMagic {
			modeList = append(modeList, 0)
		}
		if canBless {
			modeList = append(modeList, 1)
		}
		if selectionIndex < 0 || selectionIndex >= len(modeList) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		if modeList[selectionIndex] == 0 {
			ctxData["choice_type"] = "elf_elemental_shot_discard_magic"
			ctxData["magic_indices"] = getCardIndicesByType(user, model.CardTypeMagic)
		} else {
			ctxData["choice_type"] = "elf_elemental_shot_remove_blessing"
			ctxData["blessing_indices"] = elfBlessingHandIndices(user)
		}
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil

	case "elf_elemental_shot_discard_magic", "elf_elemental_shot_remove_blessing":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		key := "magic_indices"
		if choiceType == "elf_elemental_shot_remove_blessing" {
			key = "blessing_indices"
		}
		candidates := parseIntSliceContextValue(ctxData[key])
		cardIdx, ok := runtimeutil.ResolveSelectionToCandidate(selectionIndex, candidates)
		if !ok {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}

		var card model.Card
		if choiceType == "elf_elemental_shot_remove_blessing" {
			if cardIdx < 0 || cardIdx >= len(user.Blessings) {
				return true, fmt.Errorf("无效的祝福索引: %d", selectionIndex)
			}
			card = user.Blessings[cardIdx]
			removeElfBlessingByCardID(user, card.ID)
		} else {
			if cardIdx < 0 || cardIdx >= len(user.Hand) {
				return true, fmt.Errorf("无效的手牌索引: %d", selectionIndex)
			}
			card = user.Hand[cardIdx]
			user.Hand = append(user.Hand[:cardIdx], user.Hand[cardIdx+1:]...)
		}
		e.NotifyCardRevealed(user.ID, []model.Card{card}, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, card)

		if user.Tokens == nil {
			user.Tokens = map[string]int{}
		}
		user.Tokens["elf_elemental_shot_water_pending"] = 0
		user.Tokens["elf_elemental_shot_earth_pending"] = 0
		user.Tokens["elf_elemental_shot_wind_pending"] = 0
		attackElement, _ := ctxData["attack_element"].(string)
		switch attackElement {
		case string(model.ElementFire):
			e.ApplyNextAttackDamageRule(user.ID, "elf_elemental_shot_fire_attack_bonus", "elf_elemental_shot", 1, model.RuleLifeThisEffectChain)
		case string(model.ElementWater):
			user.Tokens["elf_elemental_shot_water_pending"] = 1
		case string(model.ElementWind):
			user.Tokens["elf_elemental_shot_wind_pending"] = 1
		case string(model.ElementThunder):
			e.ApplyNextAttackInterceptTagRule(user.ID, "elf_elemental_shot_thunder_attack_tag", "elf_elemental_shot", model.CombatInterceptUnrespondable, model.RuleLifeThisEffectChain)
		case string(model.ElementEarth):
			user.Tokens["elf_elemental_shot_earth_pending"] = 1
		}
		e.Log(fmt.Sprintf("%s 发动 [元素射击]（%s）", user.Name, attackElement))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if len(e.State.ActionQueue) > 0 {
				e.enterActionExecutionStage()
			} else if len(e.State.PendingDamageQueue) > 0 {
				e.enterDamageResolution(nil)
			}
		}
		return true, nil

	case "elf_animal_companion_confirm":
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
		if e.CanPayCrystalCost(user.ID, 1) {
			e.State.PendingInterrupt.Context = map[string]interface{}{"choice_type": "elf_pet_empower_confirm", "user_id": userID}
			e.notifyInterruptPrompt()
			return true, nil
		}
		e.DrawCards(user.ID, 1)
		e.PushInterrupt(&model.Interrupt{Type: model.InterruptDiscard, PlayerID: user.ID, Context: map[string]interface{}{"discard_count": 1, "stay_in_turn": true, "prompt": "【动物伙伴】请选择弃置1张牌：", "exclude_blessings": true}})
		e.PopInterrupt()
		return true, nil

	case "elf_pet_empower_confirm":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		if selectionIndex == 0 && e.ConsumeCrystalCost(user.ID, 1) {
			e.State.PendingInterrupt.Context = map[string]interface{}{"choice_type": "elf_pet_empower_target", "user_id": userID, "target_ids": append([]string{}, e.State.PlayerOrder...)}
			e.notifyInterruptPrompt()
			return true, nil
		}
		e.DrawCards(user.ID, 1)
		e.PushInterrupt(&model.Interrupt{Type: model.InterruptDiscard, PlayerID: user.ID, Context: map[string]interface{}{"discard_count": 1, "stay_in_turn": true, "prompt": "【动物伙伴】请选择弃置1张牌：", "exclude_blessings": true}})
		e.PopInterrupt()
		return true, nil

	case "elf_pet_empower_target":
		targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		target := e.State.Players[targetIDs[selectionIndex]]
		if target == nil {
			return true, fmt.Errorf("目标不存在")
		}
		e.DrawCards(target.ID, 1)
		if len(target.Hand) > e.GetMaxHand(target) {
			e.Log(fmt.Sprintf("[宠物强化] %s 摸牌后触发爆牌，本次弃1由爆牌弃牌结算承担", target.Name))
			e.PopInterrupt()
			return true, nil
		}
		e.PushInterrupt(&model.Interrupt{Type: model.InterruptDiscard, PlayerID: target.ID, Context: map[string]interface{}{"discard_count": 1, "stay_in_turn": true, "prompt": fmt.Sprintf("【宠物强化】%s 请弃置1张牌：", target.Name), "exclude_blessings": e.isElfArcher(target)}})
		e.PopInterrupt()
		return true, nil

	case "elf_elemental_shot_water_target", "elf_elemental_shot_earth_target", "elf_ritual_release_target":
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
		switch choiceType {
		case "elf_elemental_shot_water_target":
			e.Heal(targetID, 1)
		case "elf_elemental_shot_earth_target":
			e.AddPendingDamage(model.PendingDamage{SourceID: user.ID, TargetID: targetID, Damage: 1, DamageType: "magic"})
		case "elf_ritual_release_target":
			leaveElfArcherRitualForm(user)
			user.Tokens["elf_ritual_release_waiting"] = 0
			e.AddPendingDamage(model.PendingDamage{SourceID: user.ID, TargetID: targetID, Damage: 2, DamageType: "magic"})
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil && len(e.State.PendingDamageQueue) > 0 {
			e.enterDamageResolution(nil)
		}
		return true, nil
	}

	return false, nil
}
