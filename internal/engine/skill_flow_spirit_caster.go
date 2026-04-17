// gameflow: 通灵师：念咒、百鬼夜行等。

package engine

import (
	"fmt"
	"starcup-engine/internal/engine/core/runtimeutil"
	"starcup-engine/internal/model"
)

func (e *GameEngine) continueSpiritCasterTalisman(user *model.Player, skillID string, targetIDs []string) error {
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	switch skillID {
	case "sc_talisman_thunder":
		if e.CanPayCrystalCost(user.ID, 1) {
			e.PushInterrupt(&model.Interrupt{
				Type:     model.InterruptChoice,
				PlayerID: user.ID,
				Context: map[string]interface{}{
					"choice_type": "sc_spiritual_collapse_confirm",
					"user_id":     user.ID,
					"mode":        "sc_talisman_thunder",
					"target_ids":  append([]string{}, targetIDs...),
				},
			})
			return nil
		}
		e.resolveSpiritCasterThunderDamage(user, targetIDs, 0)
	case "sc_talisman_wind":
		return e.startSpiritCasterWindDiscardFlow(user, targetIDs)
	default:
		return fmt.Errorf("未知灵符技能: %s", skillID)
	}
	return nil
}

func (e *GameEngine) resolveSpiritCasterThunderDamage(user *model.Player, targetIDs []string, bonus int) {
	if user == nil {
		return
	}
	damage := 1 + bonus
	if damage < 0 {
		damage = 0
	}
	targetSet := runtimeutil.IDsToSet(runtimeutil.DedupeIDs(targetIDs))
	ordered := e.reverseOrderTargetIDsFrom(user.ID, true)
	hitCount := 0
	for _, targetID := range ordered {
		if !targetSet[targetID] {
			continue
		}
		e.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   targetID,
			Damage:     damage,
			DamageType: model.MagicAttack,
		})
		hitCount++
	}
	e.Log(fmt.Sprintf("%s 发动 [灵符-雷鸣]：对%d名角色各造成%d点法术伤害", user.Name, hitCount, damage))
	e.routePendingDamageWithDefaultReturn(model.TurnStageExtraAction)
}

func (e *GameEngine) startSpiritCasterWindDiscardFlow(user *model.Player, targetIDs []string) error {
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	targetSet := runtimeutil.IDsToSet(runtimeutil.DedupeIDs(targetIDs))
	orderedAll := e.reverseOrderTargetIDsFrom(user.ID, true)
	ordered := make([]string, 0, len(targetIDs))
	for _, playerID := range orderedAll {
		if !targetSet[playerID] {
			continue
		}
		ordered = append(ordered, playerID)
	}
	if len(ordered) == 0 {
		e.Log(fmt.Sprintf("%s 的 [灵符-风行]：无有效目标", user.Name))
		return nil
	}

	cursor := 0
	for cursor < len(ordered) {
		target := e.State.Players[ordered[cursor]]
		if target == nil || len(target.Hand) == 0 {
			if target != nil {
				e.Log(fmt.Sprintf("%s 的 [灵符-风行]：%s 无手牌可弃置", user.Name, target.Name))
			}
			cursor++
			continue
		}
		break
	}
	if cursor >= len(ordered) {
		e.Log(fmt.Sprintf("%s 的 [灵符-风行]：所有目标均无手牌可弃置", user.Name))
		return nil
	}

	currentTargetID := ordered[cursor]
	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: currentTargetID,
		Context: map[string]interface{}{
			"choice_type":        "sc_talisman_wind_discard",
			"user_id":            user.ID,
			"ordered_target_ids": ordered,
			"cursor":             cursor,
			"current_target_id":  currentTargetID,
		},
	})
	return nil
}

func (e *GameEngine) resolveSpiritCasterHundredNightSingle(user *model.Player, targetID string, bonus int) error {
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	target := e.State.Players[targetID]
	if target == nil {
		return fmt.Errorf("目标不存在")
	}
	damage := 1 + bonus
	e.AddPendingDamage(model.PendingDamage{
		SourceID:   user.ID,
		TargetID:   target.ID,
		Damage:     damage,
		DamageType: model.MagicAttack,
	})
	e.Log(fmt.Sprintf("%s 发动 [百鬼夜行]：对 %s 造成%d点法术伤害", user.Name, target.Name, damage))
	return nil
}

func (e *GameEngine) resolveSpiritCasterHundredNightFireAOE(user *model.Player, excludeIDs []string, bonus int) error {
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	exclude := runtimeutil.IDsToSet(runtimeutil.DedupeIDs(excludeIDs))
	damage := 1 + bonus
	ordered := e.reverseOrderTargetIDsFrom(user.ID, true)
	hitCount := 0
	for _, playerID := range ordered {
		if exclude[playerID] {
			continue
		}
		target := e.State.Players[playerID]
		if target == nil {
			continue
		}
		e.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   target.ID,
			Damage:     damage,
			DamageType: model.MagicAttack,
		})
		hitCount++
	}
	e.Log(fmt.Sprintf("%s 发动 [百鬼夜行·火]：对除2名指定角色外的其他角色各造成%d点法术伤害（命中%d名）", user.Name, damage, hitCount))
	return nil
}

func (e *GameEngine) buildSpiritCasterChoicePrompt(choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "sc_incant_confirm":
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【念咒】是否将1张手牌面朝下放置为妖力？",
			Options:  []model.PromptOption{{ID: "0", Label: "是"}, {ID: "1", Label: "否"}},
			Min:      1,
			Max:      1,
		}
	case "sc_incant_card":
		options := make([]model.PromptOption, 0, len(player.Hand))
		for idx, c := range player.Hand {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", idx), Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(c))})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【念咒】请选择要作为妖力盖放的手牌：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	case "sc_hundred_night_power":
		powers := spiritCasterPowerCovers(player)
		options := make([]model.PromptOption, 0, len(powers))
		for i, fc := range powers {
			if fc == nil {
				continue
			}
			eleZh := elementNameForPrompt(string(fc.Card.Element))
			if eleZh == "" {
				eleZh = string(fc.Card.Element)
			}
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", i), Label: fmt.Sprintf("%s（%s系）", fc.Card.Name, eleZh)})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【百鬼夜行】请选择要移除的1个妖力：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	case "sc_hundred_night_fire_reveal":
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【百鬼夜行】移除的是火系妖力，是否展示并改为范围伤害？",
			Options:  []model.PromptOption{{ID: "0", Label: "展示并改为范围伤害"}, {ID: "1", Label: "不展示，改为单体伤害"}},
			Min:      1,
			Max:      1,
		}
	case "sc_hundred_night_exclude_pick":
		targetIDs := runtimeutil.ParseStringSliceContextValue(data["target_ids"])
		selectedSet := runtimeutil.IDsToSet(runtimeutil.ParseStringSliceContextValue(data["selected_exclude_ids"]))
		options := make([]model.PromptOption, 0, len(targetIDs))
		for _, targetID := range targetIDs {
			if selectedSet[targetID] {
				continue
			}
			if target := e.State.Players[targetID]; target != nil {
				options = append(options, model.PromptOption{ID: targetID, Label: target.Name})
			}
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【百鬼夜行】请选择第 %d/2 名排除目标：", len(selectedSet)+1),
			Options:  options,
			Min:      1,
			Max:      1,
		}
	case "sc_spiritual_collapse_confirm":
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【灵力崩解】是否消耗1点水晶（红宝石可替代），使本次每段伤害额外+1？",
			Options:  []model.PromptOption{{ID: "0", Label: "是"}, {ID: "1", Label: "否"}},
			Min:      1,
			Max:      1,
		}
	case "sc_talisman_wind_discard":
		currentTargetID, _ := data["current_target_id"].(string)
		target := e.State.Players[currentTargetID]
		if target == nil {
			return nil
		}
		options := make([]model.PromptOption, 0, len(target.Hand))
		for idx, c := range target.Hand {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", idx), Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(c))})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【灵符-风行】请 %s 选择1张手牌弃置：", target.Name),
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}
	return nil
}

func (e *GameEngine) handleSpiritCasterChoiceInput(playerID string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "sc_incant_confirm":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		skillID, _ := ctxData["skill_id"].(string)
		targetIDs := runtimeutil.DedupeIDs(runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"]))
		switch selectionIndex {
		case 0:
			if len(user.Hand) == 0 {
				e.Log(fmt.Sprintf("%s 的 [念咒] 未触发：无手牌可放置为妖力", user.Name))
				e.PopInterrupt()
				return true, e.continueSpiritCasterTalisman(user, skillID, targetIDs)
			}
			ctxData["choice_type"] = "sc_incant_card"
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return true, nil
		case 1:
			e.PopInterrupt()
			return true, e.continueSpiritCasterTalisman(user, skillID, targetIDs)
		default:
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
	case "sc_incant_card":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		skillID, _ := ctxData["skill_id"].(string)
		targetIDs := runtimeutil.DedupeIDs(runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"]))
		if selectionIndex < 0 || selectionIndex >= len(user.Hand) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		card := user.Hand[selectionIndex]
		user.Hand = append(user.Hand[:selectionIndex], user.Hand[selectionIndex+1:]...)
		if !addSpiritCasterPowerCard(user, card) {
			user.Hand = append(user.Hand, card)
			return true, fmt.Errorf("放置妖力失败")
		}
		e.Log(fmt.Sprintf("%s 发动 [念咒]：将1张手牌盖放为妖力（当前妖力%d）", user.Name, spiritCasterPowerCount(user, "")))
		e.PopInterrupt()
		return true, e.continueSpiritCasterTalisman(user, skillID, targetIDs)
	case "sc_hundred_night_power":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		powers := spiritCasterPowerCovers(user)
		if len(powers) == 0 {
			return true, fmt.Errorf("没有可移除的妖力")
		}
		powerIdx, ok := runtimeutil.ResolveSelectionToCandidate(selectionIndex, func() []int {
			idxs := make([]int, 0, len(powers))
			for i := range powers {
				idxs = append(idxs, i)
			}
			return idxs
		}())
		if !ok || powerIdx < 0 || powerIdx >= len(powers) || powers[powerIdx] == nil {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		selectedPower := powers[powerIdx]
		card := selectedPower.Card
		user.RemoveFieldCard(selectedPower)
		syncSpiritCasterPowerToken(user)
		e.State.DiscardPile = append(e.State.DiscardPile, card)
		e.Log(fmt.Sprintf("%s 发动 [百鬼夜行]：移除1个妖力", user.Name))

		ctxData["removed_card"] = card
		if card.Element == model.ElementFire {
			ctxData["removed_element"] = string(card.Element)
			ctxData["removed_name"] = card.Name
			ctxData["choice_type"] = "sc_hundred_night_fire_reveal"
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return true, nil
		}
		ctxData["choice_type"] = "sc_hundred_night_target"
		ctxData["target_ids"] = append([]string{}, e.State.PlayerOrder...)
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil
	case "sc_hundred_night_fire_reveal":
		switch selectionIndex {
		case 0:
			userID, _ := ctxData["user_id"].(string)
			user := e.State.Players[userID]
			if user != nil {
				if removedCard, ok := ctxData["removed_card"].(model.Card); ok {
					e.NotifyCardRevealed(user.ID, []model.Card{removedCard}, "discard")
				}
				e.Log(fmt.Sprintf("%s 展示了火系妖力，触发 [百鬼夜行] 范围分支", user.Name))
			}
			ctxData["choice_type"] = "sc_hundred_night_exclude_pick"
			ctxData["target_ids"] = append([]string{}, e.State.PlayerOrder...)
			ctxData["selected_exclude_ids"] = []string{}
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return true, nil
		case 1:
			ctxData["choice_type"] = "sc_hundred_night_target"
			ctxData["target_ids"] = append([]string{}, e.State.PlayerOrder...)
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return true, nil
		default:
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
	case "sc_hundred_night_exclude_pick":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		allTargetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
		if len(allTargetIDs) < 2 {
			return true, fmt.Errorf("可选目标不足2名")
		}
		selected := runtimeutil.DedupeIDs(runtimeutil.ParseStringSliceContextValue(ctxData["selected_exclude_ids"]))
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
		if len(selected) < 2 {
			ctxData["selected_exclude_ids"] = selected
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return true, nil
		}
		if e.CanPayCrystalCost(user.ID, 1) {
			ctxData["choice_type"] = "sc_spiritual_collapse_confirm"
			ctxData["mode"] = "sc_hundred_night_fire_aoe"
			ctxData["exclude_ids"] = selected
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return true, nil
		}
		if err := e.resolveSpiritCasterHundredNightFireAOE(user, selected, 0); err != nil {
			return true, err
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.enterDamageResolution(nil)
		}
		return true, nil
	case "sc_hundred_night_target":
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
		if e.CanPayCrystalCost(user.ID, 1) {
			ctxData["choice_type"] = "sc_spiritual_collapse_confirm"
			ctxData["mode"] = "sc_hundred_night_single"
			ctxData["target_id"] = targetID
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return true, nil
		}
		if err := e.resolveSpiritCasterHundredNightSingle(user, targetID, 0); err != nil {
			return true, err
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.enterDamageResolution(nil)
		}
		return true, nil
	case "sc_spiritual_collapse_confirm":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		bonus := 0
		if selectionIndex == 0 {
			if !e.ConsumeCrystalCost(user.ID, 1) {
				return true, fmt.Errorf("灵力崩解需要1点水晶（红宝石可替代）")
			}
			bonus = 1
			e.Log(fmt.Sprintf("%s 发动 [灵力崩解]：本次每段伤害额外+1", user.Name))
		} else if selectionIndex != 1 {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		mode, _ := ctxData["mode"].(string)
		switch mode {
		case "sc_talisman_thunder":
			targetIDs := runtimeutil.DedupeIDs(runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"]))
			e.PopInterrupt()
			e.resolveSpiritCasterThunderDamage(user, targetIDs, bonus)
		case "sc_hundred_night_single":
			targetID, _ := ctxData["target_id"].(string)
			if targetID == "" {
				return true, fmt.Errorf("百鬼夜行目标缺失")
			}
			if err := e.resolveSpiritCasterHundredNightSingle(user, targetID, bonus); err != nil {
				return true, err
			}
			e.PopInterrupt()
		case "sc_hundred_night_fire_aoe":
			excludeIDs := runtimeutil.DedupeIDs(runtimeutil.ParseStringSliceContextValue(ctxData["exclude_ids"]))
			if len(excludeIDs) != 2 {
				return true, fmt.Errorf("百鬼夜行火系分支需要2名排除目标")
			}
			if err := e.resolveSpiritCasterHundredNightFireAOE(user, excludeIDs, bonus); err != nil {
				return true, err
			}
			e.PopInterrupt()
		default:
			return true, fmt.Errorf("灵力崩解上下文无效")
		}
		if e.State.PendingInterrupt == nil && len(e.State.PendingDamageQueue) > 0 {
			e.enterDamageResolution(nil)
		}
		return true, nil
	case "sc_talisman_wind_discard":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("灵符师不存在")
		}
		ordered := runtimeutil.ParseStringSliceContextValue(ctxData["ordered_target_ids"])
		if len(ordered) == 0 {
			return true, fmt.Errorf("灵符-风行上下文无效")
		}
		cursor := runtimeutil.ToIntContextValue(ctxData["cursor"])
		if cursor < 0 || cursor >= len(ordered) {
			return true, fmt.Errorf("灵符-风行游标无效")
		}
		currentTargetID, _ := ctxData["current_target_id"].(string)
		if currentTargetID == "" {
			currentTargetID = ordered[cursor]
		}
		target := e.State.Players[currentTargetID]
		if target == nil {
			return true, fmt.Errorf("弃牌目标不存在")
		}
		if len(target.Hand) == 0 {
			e.Log(fmt.Sprintf("%s 的 [灵符-风行]：%s 已无手牌，跳过", user.Name, target.Name))
		} else {
			candidates := allHandIndices(target)
			cardIdx, ok := runtimeutil.ResolveSelectionToCandidate(selectionIndex, candidates)
			if !ok || cardIdx < 0 || cardIdx >= len(target.Hand) {
				return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
			}
			card := target.Hand[cardIdx]
			target.Hand = append(target.Hand[:cardIdx], target.Hand[cardIdx+1:]...)
			e.NotifyCardHidden(target.ID, []model.Card{card}, "discard")
			e.State.DiscardPile = append(e.State.DiscardPile, card)
			e.Log(fmt.Sprintf("%s 的 [灵符-风行]：%s 选择弃置了1张手牌", user.Name, target.Name))
		}

		nextCursor := cursor + 1
		for nextCursor < len(ordered) {
			nextTarget := e.State.Players[ordered[nextCursor]]
			if nextTarget == nil {
				nextCursor++
				continue
			}
			if len(nextTarget.Hand) <= 0 {
				e.Log(fmt.Sprintf("%s 的 [灵符-风行]：%s 无手牌可弃置", user.Name, nextTarget.Name))
				nextCursor++
				continue
			}
			ctxData["cursor"] = nextCursor
			ctxData["current_target_id"] = nextTarget.ID
			e.State.PendingInterrupt.PlayerID = nextTarget.ID
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return true, nil
		}

		e.Log(fmt.Sprintf("%s 的 [灵符-风行] 结算完成", user.Name))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil && len(e.State.PendingDamageQueue) > 0 {
			e.enterDamageResolution(nil)
		}
		return true, nil
	}
	return false, nil
}
