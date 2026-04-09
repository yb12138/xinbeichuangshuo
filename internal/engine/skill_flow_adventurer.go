package engine

import (
	"fmt"
	"sort"
	"starcup-engine/internal/engine/core/runtimeutil"
	"strconv"
	"strings"

	"starcup-engine/internal/model"
)

func (e *GameEngine) isForcedAdventurerParadiseResponse(playerID string) bool {
	intr := e.State.PendingInterrupt
	if intr == nil || intr.Type != model.InterruptResponseSkill || intr.PlayerID != playerID {
		return false
	}
	player := e.State.Players[playerID]
	if player == nil || player.TurnState.SkillFlowState == nil || player.TurnState.SkillFlowState["adventurer_extract_requires_paradise"] <= 0 {
		return false
	}
	for _, sid := range intr.SkillIDs {
		if sid == "adventurer_paradise" {
			return true
		}
	}
	return false
}

func (e *GameEngine) resolveAdventurerLuckyFortuneFromFraud(user *model.Player) {
	if user == nil {
		return
	}
	user.Crystal++
	e.Log(fmt.Sprintf("%s 的 [强运] 触发，获得1蓝水晶", user.Name))
	e.Log(fmt.Sprintf("[Skill] %s 使用了技能: 强运", user.Name))
}

func (e *GameEngine) resolveAdventurerUndergroundLaw(user *model.Player) {
	if user == nil {
		return
	}
	e.ModifyGem(string(user.Camp), 2)
	e.Log(fmt.Sprintf("%s 的 [地下法则] 生效，本次购买改为战绩区+2宝石", user.Name))
	e.Log(fmt.Sprintf("[Skill] %s 使用了技能: 地下法则", user.Name))
}

func (e *GameEngine) buildFraudCombos(user *model.Player, element model.Element, need int, allowAnyElementForDark bool) []string {
	if user == nil || need <= 0 {
		return nil
	}
	elemToIdx := map[model.Element][]int{}
	for i, c := range user.Hand {
		elemToIdx[c.Element] = append(elemToIdx[c.Element], i)
	}

	var targets []model.Element
	if allowAnyElementForDark {
		for ele, idxs := range elemToIdx {
			if len(idxs) >= need {
				targets = append(targets, ele)
			}
		}
	} else {
		if len(elemToIdx[element]) >= need {
			targets = append(targets, element)
		}
	}

	var combos []string
	for _, ele := range targets {
		idxs := elemToIdx[ele]
		for _, picked := range runtimeutil.PickKIndices(idxs, need) {
			parts := make([]string, 0, len(picked))
			for _, v := range picked {
				parts = append(parts, fmt.Sprintf("%d", v))
			}
			combos = append(combos, fmt.Sprintf("%s:%s", ele, strings.Join(parts, ",")))
		}
	}
	return combos
}

func (e *GameEngine) buildAdventurerChoicePrompt(choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "adventurer_fraud_mode":
		can2, _ := data["can2"].(bool)
		can3, _ := data["can3"].(bool)
		options := make([]model.PromptOption, 0, 2)
		if can2 {
			options = append(options, model.PromptOption{ID: "0", Label: "弃2张同系牌，视为非暗灭任意系主动攻击"})
		}
		if can3 {
			options = append(options, model.PromptOption{ID: "1", Label: "弃3张同系牌，视为暗灭主动攻击"})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【欺诈】请选择发动方式：", Options: options, Min: 1, Max: 1}

	case "adventurer_fraud_attack_element":
		options := make([]model.PromptOption, 0, 6)
		for _, ele := range []string{
			string(model.ElementWater), string(model.ElementFire), string(model.ElementEarth),
			string(model.ElementWind), string(model.ElementThunder), string(model.ElementLight),
		} {
			options = append(options, model.PromptOption{ID: ele, Label: fmt.Sprintf("%s系", elementNameForPrompt(ele))})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【欺诈】请选择本次攻击系别（不可选暗）：", Options: options, Min: 1, Max: 1}

	case "adventurer_fraud_discard_element":
		elems := runtimeutil.ParseStringSliceContextValue(data["discard_elements"])
		options := make([]model.PromptOption, 0, len(elems))
		for _, ele := range elems {
			options = append(options, model.PromptOption{ID: ele, Label: fmt.Sprintf("弃%s系同系2张", elementNameForPrompt(ele))})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【欺诈】请选择用于弃置的同系牌：", Options: options, Min: 1, Max: 1}

	case "adventurer_fraud_discard_combo":
		combos := runtimeutil.ParseStringSliceContextValue(data["combos"])
		options := make([]model.PromptOption, 0, len(combos))
		for _, combo := range combos {
			parts := strings.Split(combo, ":")
			if len(parts) < 2 {
				continue
			}
			ele := parts[0]
			label := fmt.Sprintf("%s系 组合", elementNameForPrompt(ele))
			if len(parts) == 2 && player != nil {
				idxStrs := strings.Split(parts[1], ",")
				cardLabels := make([]string, 0, len(idxStrs))
				for _, idxStr := range idxStrs {
					idx, err := strconv.Atoi(idxStr)
					if err != nil || idx < 0 || idx >= len(player.Hand) {
						continue
					}
					cardLabels = append(cardLabels, fmt.Sprintf("%d:%s", idx+1, player.Hand[idx].Name))
				}
				if len(cardLabels) > 0 {
					label = fmt.Sprintf("%s系 [%s]", elementNameForPrompt(ele), strings.Join(cardLabels, " + "))
				}
			}
			options = append(options, model.PromptOption{ID: combo, Label: label})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【欺诈】请选择要弃置的同系牌组合：", Options: options, Min: 1, Max: 1}

	case "adventurer_paradise_target":
		allyIDs := runtimeutil.ParseStringSliceContextValue(data["ally_ids"])
		options := make([]model.PromptOption, 0, len(allyIDs))
		for _, allyID := range allyIDs {
			if target := e.State.Players[allyID]; target != nil {
				options = append(options, model.PromptOption{ID: allyID, Label: target.Name})
			}
		}
		transferGem := runtimeutil.ToIntContextValue(data["transfer_gem"])
		transferCrystal := runtimeutil.ToIntContextValue(data["transfer_crystal"])
		transferTotal := runtimeutil.ToIntContextValue(data["transfer_total"])
		if transferTotal <= 0 {
			transferTotal = transferGem + transferCrystal
		}
		message := "【冒险者天堂】请选择接收能量的队友："
		if transferTotal > 0 {
			message = fmt.Sprintf("【冒险者天堂】请选择接收提炼结果的队友（共%d点：宝石%d / 水晶%d）：", transferTotal, transferGem, transferCrystal)
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: message, Options: options, Min: 1, Max: 1}

	case "adventurer_steal_sky_mode":
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【偷天换日】请选择效果：",
			Options: []model.PromptOption{
				{ID: "0", Label: "转移对方战绩区1红宝石到我方"},
				{ID: "1", Label: "将我方战绩区全部蓝水晶转换成红宝石"},
			},
			Min: 1,
			Max: 1,
		}

	case "adventurer_steal_sky_extra_action":
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【偷天换日】请选择额外行动类型：",
			Options: []model.PromptOption{
				{ID: "0", Label: "额外+1攻击行动"},
				{ID: "1", Label: "额外+1法术行动"},
			},
			Min: 1,
			Max: 1,
		}
	}

	return nil
}

func (e *GameEngine) handleAdventurerChoiceInput(_ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)

	switch choiceType {
	case "adventurer_fraud_mode":
		can2, _ := ctxData["can2"].(bool)
		can3, _ := ctxData["can3"].(bool)
		modeList := make([]string, 0, 2)
		if can2 {
			modeList = append(modeList, "2")
		}
		if can3 {
			modeList = append(modeList, "3")
		}
		if selectionIndex < 0 || selectionIndex >= len(modeList) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		mode := modeList[selectionIndex]
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}

		if mode == "2" {
			counts := map[string]int{}
			for _, card := range user.Hand {
				counts[string(card.Element)]++
			}
			discardElems := make([]string, 0)
			for _, ele := range []string{
				string(model.ElementWater), string(model.ElementFire), string(model.ElementEarth),
				string(model.ElementWind), string(model.ElementThunder), string(model.ElementLight), string(model.ElementDark),
			} {
				if counts[ele] >= 2 {
					discardElems = append(discardElems, ele)
				}
			}
			if len(discardElems) == 0 {
				return true, fmt.Errorf("没有可用于弃2同系的元素")
			}
			ctxData["choice_type"] = "adventurer_fraud_attack_element"
			ctxData["fraud_mode"] = "2"
			ctxData["discard_elements"] = discardElems
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return true, nil
		}

		combos := e.buildFraudCombos(user, model.ElementDark, 3, true)
		if len(combos) == 0 {
			return true, fmt.Errorf("没有可用于弃3同系的组合")
		}
		ctxData["choice_type"] = "adventurer_fraud_discard_combo"
		ctxData["fraud_mode"] = "3"
		ctxData["chosen_element"] = string(model.ElementDark)
		ctxData["combos"] = combos
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil

	case "adventurer_fraud_attack_element":
		attackElems := []string{
			string(model.ElementWater), string(model.ElementFire), string(model.ElementEarth),
			string(model.ElementWind), string(model.ElementThunder), string(model.ElementLight),
		}
		if selectionIndex < 0 || selectionIndex >= len(attackElems) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		ctxData["chosen_element"] = attackElems[selectionIndex]
		ctxData["choice_type"] = "adventurer_fraud_discard_element"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil

	case "adventurer_fraud_discard_element":
		elems := runtimeutil.ParseStringSliceContextValue(ctxData["discard_elements"])
		if selectionIndex < 0 || selectionIndex >= len(elems) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		ele := elems[selectionIndex]
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		combos := e.buildFraudCombos(user, model.Element(ele), 2, false)
		if len(combos) == 0 {
			return true, fmt.Errorf("该元素无可弃组合")
		}
		ctxData["choice_type"] = "adventurer_fraud_discard_combo"
		ctxData["discard_element"] = ele
		ctxData["combos"] = combos
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil

	case "adventurer_fraud_discard_combo":
		combos := runtimeutil.ParseStringSliceContextValue(ctxData["combos"])
		if selectionIndex < 0 || selectionIndex >= len(combos) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		parts := strings.Split(combos[selectionIndex], ":")
		if len(parts) != 2 {
			return true, fmt.Errorf("组合格式错误")
		}
		ele := parts[0]
		idxStrs := strings.Split(parts[1], ",")
		idxs := make([]int, 0, len(idxStrs))
		for _, idxStr := range idxStrs {
			idx, err := strconv.Atoi(idxStr)
			if err != nil {
				return true, fmt.Errorf("组合索引错误")
			}
			idxs = append(idxs, idx)
		}
		sort.Sort(sort.Reverse(sort.IntSlice(idxs)))
		for _, idx := range idxs {
			if idx < 0 || idx >= len(user.Hand) {
				return true, fmt.Errorf("弃牌索引越界")
			}
			card := user.Hand[idx]
			e.NotifyCardRevealed(userID, []model.Card{card}, "discard")
			user.Hand = append(user.Hand[:idx], user.Hand[idx+1:]...)
			e.State.DiscardPile = append(e.State.DiscardPile, card)
		}

		mode, _ := ctxData["fraud_mode"].(string)
		attackElement := model.Element(ele)
		canBeResponded := true
		if mode == "3" {
			attackElement = model.ElementDark
			canBeResponded = false
		} else if chosen, ok := ctxData["chosen_element"].(string); ok && chosen != "" {
			attackElement = model.Element(chosen)
		}

		rawCtx, ok := ctxData["user_ctx"].(*model.Context)
		if ok && rawCtx != nil && rawCtx.TriggerCtx != nil && rawCtx.TriggerCtx.Card != nil && rawCtx.TriggerCtx.AttackInfo != nil {
			rawCtx.TriggerCtx.Card.Faction = ""
			rawCtx.TriggerCtx.Card.Element = attackElement
			rawCtx.TriggerCtx.Card.Damage = 2
			rawCtx.TriggerCtx.AttackInfo.CanBeResponded = canBeResponded
			e.Log(fmt.Sprintf("%s 发动[欺诈]完成，弃同系牌并将本次攻击改为 %s", user.Name, attackElement))
			e.resolveAdventurerLuckyFortuneFromFraud(user)
			e.PopInterrupt()
			return true, nil
		}

		targetID, _ := ctxData["fraud_target_id"].(string)
		if targetID == "" || e.State.Players[targetID] == nil {
			return true, fmt.Errorf("欺诈目标无效")
		}
		virtualCard := model.Card{
			ID:      "fraud_virtual_attack",
			Name:    "欺诈",
			Type:    model.CardTypeAttack,
			Element: attackElement,
			Faction: "",
			Damage:  2,
		}
		e.State.ActionQueue = append(e.State.ActionQueue, model.QueuedAction{
			SourceID:        userID,
			TargetID:        targetID,
			Type:            model.ActionAttack,
			Element:         attackElement,
			Card:            &virtualCard,
			CardIndex:       -1,
			SourceSkill:     "adventurer_fraud",
			UsesVirtualCard: true,
		})
		e.resolveAdventurerLuckyFortuneFromFraud(user)
		e.enterActionExecutionStage()
		e.Log(fmt.Sprintf("%s 发动[欺诈]完成，弃同系牌并对 %s 发起%s系主动攻击", user.Name, e.State.Players[targetID].Name, attackElement))
		e.PopInterrupt()
		return true, nil

	case "adventurer_paradise_target":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		allyIDs := runtimeutil.ParseStringSliceContextValue(ctxData["ally_ids"])
		if selectionIndex < 0 || selectionIndex >= len(allyIDs) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		ally := e.State.Players[allyIDs[selectionIndex]]
		if ally == nil {
			return true, fmt.Errorf("队友不存在")
		}

		transferGem := runtimeutil.ToIntContextValue(ctxData["transfer_gem"])
		transferCrystal := runtimeutil.ToIntContextValue(ctxData["transfer_crystal"])
		transferTotal := runtimeutil.ToIntContextValue(ctxData["transfer_total"])
		if transferTotal <= 0 {
			transferTotal = transferGem + transferCrystal
		}
		fromPending, _ := ctxData["from_pending"].(bool)
		if transferTotal <= 0 {
			e.clearAdventurerExtractState(user)
			e.PopInterrupt()
			return true, nil
		}

		capLeft := e.getPlayerEnergyCap(ally) - (ally.Gem + ally.Crystal)
		if capLeft < transferTotal {
			return true, fmt.Errorf("%s 能量空间不足，无法接收全部提炼结果", ally.Name)
		}
		if !fromPending {
			if user.Gem < transferGem || user.Crystal < transferCrystal {
				return true, fmt.Errorf("自身提炼结果异常，无法转移")
			}
			user.Gem -= transferGem
			user.Crystal -= transferCrystal
		}
		ally.Gem += transferGem
		ally.Crystal += transferCrystal

		removedEnergy := false
		if user.Crystal > 0 {
			user.Crystal--
			removedEnergy = true
		} else if user.Gem > 0 {
			user.Gem--
			removedEnergy = true
		}
		e.clearAdventurerExtractState(user)
		if removedEnergy {
			e.Log(fmt.Sprintf("%s 发动[冒险者天堂]，将提炼结果交给 %s（宝石%d/水晶%d），并移除自身1点能量", user.Name, ally.Name, transferGem, transferCrystal))
		} else {
			e.Log(fmt.Sprintf("%s 发动[冒险者天堂]，将提炼结果交给 %s（宝石%d/水晶%d）", user.Name, ally.Name, transferGem, transferCrystal))
		}
		e.PopInterrupt()
		return true, nil

	case "adventurer_steal_sky_mode":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		enemyCamp, _ := ctxData["enemy_camp"].(string)
		selfCamp, _ := ctxData["self_camp"].(string)
		switch selectionIndex {
		case 0:
			if e.GetCampGems(enemyCamp) > 0 {
				e.ModifyGem(enemyCamp, -1)
				e.ModifyGem(selfCamp, 1)
			}
		case 1:
			crystals := e.GetCampCrystals(selfCamp)
			if crystals > 0 {
				e.ModifyCrystal(selfCamp, -crystals)
				e.ModifyGem(selfCamp, crystals)
			}
		default:
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		e.State.PendingInterrupt.Context = map[string]interface{}{
			"choice_type": "adventurer_steal_sky_extra_action",
			"user_id":     userID,
		}
		e.notifyInterruptPrompt()
		e.Log(fmt.Sprintf("%s 完成[偷天换日]主效果，等待选择额外行动", user.Name))
		return true, nil

	case "adventurer_steal_sky_extra_action":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		switch selectionIndex {
		case 0:
			model.AppendAttackAction(user, "偷天换日")
		case 1:
			model.AppendMagicAction(user, "偷天换日")
		default:
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		e.PopInterrupt()
		return true, nil
	}

	return false, nil
}
