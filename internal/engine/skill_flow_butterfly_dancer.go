package engine

import (
	"fmt"
	"starcup-engine/internal/engine/runtimeutil"
	"strconv"
	"strings"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func (e *GameEngine) butterflyActionTargetIDs() []string {
	if e == nil || e.State == nil {
		return nil
	}
	targetIDs := make([]string, 0, len(e.State.PlayerOrder))
	for _, pid := range e.State.PlayerOrder {
		if e.State.Players[pid] != nil {
			targetIDs = append(targetIDs, pid)
		}
	}
	return targetIDs
}

func (e *GameEngine) buildButterflyChoicePrompt(choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "bt_dance_mode":
		canDiscard := runtimeutil.ToBoolContextValue(data["can_discard"])
		options := []model.PromptOption{{ID: "0", Label: "摸1张牌"}}
		if canDiscard {
			options = append(options, model.PromptOption{ID: "1", Label: "弃1张牌"})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, ChoiceType: choiceType, Message: "【舞动】请选择先执行的动作：", Options: options, Min: 1, Max: 1}

	case "bt_dance_discard":
		var options []model.PromptOption
		for idx, c := range player.Hand {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", idx), Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(c))})
		}
		return &model.Prompt{Type: model.PromptChooseCards, PlayerID: playerID, ChoiceType: choiceType, Message: "【舞动】请选择要弃置的1张手牌：", Options: options, Min: 1, Max: 1}

	case "bt_cocoon_overflow_discard":
		discardCount := runtimeutil.ToIntContextValue(data["discard_count"])
		if discardCount < 0 {
			discardCount = 0
		}
		cocoonIndices := butterflyCocoonFieldIndices(player)
		if discardCount > len(cocoonIndices) {
			discardCount = len(cocoonIndices)
		}
		var options []model.PromptOption
		for _, idx := range cocoonIndices {
			if idx < 0 || idx >= len(player.Field) || player.Field[idx] == nil {
				continue
			}
			fc := player.Field[idx]
			if fc.Mode != model.FieldCover || fc.Effect != model.EffectButterflyCocoon {
				continue
			}
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", idx), Label: fmt.Sprintf("茧[%d]: %s", idx, formatCardInfo(fc.Card))})
		}
		return &model.Prompt{Type: model.PromptChooseCards, PlayerID: playerID, ChoiceType: choiceType, Message: fmt.Sprintf("【茧上限】请选择要舍弃的%d个茧：", discardCount), Options: options, Min: discardCount, Max: discardCount}

	case "bt_reverse_mode":
		canBranch2 := runtimeutil.ToBoolContextValue(data["can_branch2"])
		options := []model.PromptOption{{ID: "0", Label: "分支①：对目标造成1点不可治疗抵御的法术伤害"}}
		if canBranch2 {
			options = append(options, model.PromptOption{ID: "1", Label: "分支②：移除2个茧或自伤4，然后移除1个蛹"})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, ChoiceType: choiceType, Message: "【倒逆之蝶】请选择发动分支：", Options: options, Min: 1, Max: 1}

	case "bt_reverse_target":
		var targetIDs []string
		if arr, ok := data["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := data["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		var options []model.PromptOption
		for _, tid := range targetIDs {
			if p := e.State.Players[tid]; p != nil {
				options = append(options, model.PromptOption{ID: tid, Label: p.Name})
			}
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, ChoiceType: choiceType, Message: "【倒逆之蝶】请选择分支①伤害目标：", Options: options, Min: 1, Max: 1}

	case "bt_reverse_branch2_cost":
		canRemove := runtimeutil.ToBoolContextValue(data["can_remove_cocoon"])
		options := []model.PromptOption{}
		if canRemove {
			options = append(options, model.PromptOption{ID: "0", Label: "移除2个茧"})
		}
		options = append(options, model.PromptOption{ID: "1", Label: "对自己造成4点法术伤害"})
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, ChoiceType: choiceType, Message: "【倒逆之蝶】请选择分支②代价：", Options: options, Min: 1, Max: 1}

	case "bt_reverse_branch2_pick":
		cocoonIndices := butterflyCocoonFieldIndices(player)
		pickCount := 2
		if pickCount > len(cocoonIndices) {
			pickCount = len(cocoonIndices)
		}
		var options []model.PromptOption
		for _, idx := range cocoonIndices {
			if idx < 0 || idx >= len(player.Field) || player.Field[idx] == nil {
				continue
			}
			fc := player.Field[idx]
			if fc.Mode != model.FieldCover || fc.Effect != model.EffectButterflyCocoon {
				continue
			}
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", idx), Label: fmt.Sprintf("茧[%d]: %s", idx, formatCardInfo(fc.Card))})
		}
		return &model.Prompt{Type: model.PromptChooseCards, PlayerID: playerID, ChoiceType: choiceType, Message: fmt.Sprintf("【倒逆之蝶】分支②请选择要移除的%d个茧：", pickCount), Options: options, Min: pickCount, Max: pickCount}

	case "bt_pilgrimage_pick", "bt_poison_pick":
		var cocoonIndices []int
		if arr, ok := data["cocoon_indices"].([]int); ok {
			cocoonIndices = append(cocoonIndices, arr...)
		} else if arr, ok := data["cocoon_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					cocoonIndices = append(cocoonIndices, int(f))
				}
			}
		}
		options := []model.PromptOption{{ID: "-1", Label: "不发动"}}
		for _, idx := range cocoonIndices {
			if idx < 0 || idx >= len(player.Field) || player.Field[idx] == nil {
				continue
			}
			fc := player.Field[idx]
			if fc.Mode != model.FieldCover || fc.Effect != model.EffectButterflyCocoon {
				continue
			}
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: fmt.Sprintf("移除茧[%d]: %s", idx, formatCardInfo(fc.Card))})
		}
		msg := "【朝圣】是否移除1个茧抵御1点伤害？"
		if choiceType == "bt_poison_pick" {
			msg = "【毒粉】是否移除1个茧使该次法术伤害+1？"
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, ChoiceType: choiceType, Message: msg, Options: options, Min: 1, Max: 1}

	case "bt_mirror_pair":
		var labels []string
		if arr, ok := data["pair_labels"].([]string); ok {
			labels = append(labels, arr...)
		} else if arr, ok := data["pair_labels"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					labels = append(labels, s)
				}
			}
		}
		options := []model.PromptOption{{ID: "-1", Label: "不发动"}}
		for i, label := range labels {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", i+1), Label: fmt.Sprintf("移除并展示：%s", label)})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, ChoiceType: choiceType, Message: "【镜花水月】是否发动并改写该次2点法术伤害？", Options: options, Min: 1, Max: 1}

	case "bt_wither_confirm":
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, ChoiceType: choiceType, Message: "【凋零】可发动：是否对目标造成1点法术伤害并对自己造成2点法术伤害？", Options: []model.PromptOption{{ID: "0", Label: "发动凋零"}, {ID: "1", Label: "不发动"}}, Min: 1, Max: 1}

	case "bt_wither_target":
		var targetIDs []string
		if arr, ok := data["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := data["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		var options []model.PromptOption
		for _, tid := range targetIDs {
			if p := e.State.Players[tid]; p != nil {
				options = append(options, model.PromptOption{ID: tid, Label: p.Name})
			}
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, ChoiceType: choiceType, Message: "【凋零】请选择1名目标角色：", Options: options, Min: 1, Max: 1}
	}
	return nil
}

func (e *GameEngine) ResolveButterflyChrysalis(userID string) error {
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	now := addButterflyPupa(user, 1)
	cards, newDeck, newDiscard := rules.DrawCards(e.State.Deck, e.State.DiscardPile, 4)
	e.State.Deck = newDeck
	e.State.DiscardPile = newDiscard
	added := addButterflyCocoonCards(user, cards)
	e.Log(fmt.Sprintf("%s 发动 [蛹化]：蛹+1（当前%d），获得%d个茧", user.Name, now, added))
	e.checkHandLimit(user, nil)
	overflow := butterflyCocoonCount(user) - butterflyCocoonCapEngine
	if overflow > 0 {
		e.PushInterrupt(&model.Interrupt{Type: model.InterruptChoice, PlayerID: user.ID, Context: map[string]interface{}{"choice_type": "bt_cocoon_overflow_discard", "user_id": user.ID, "discard_count": overflow}})
	}
	return nil
}

func (e *GameEngine) StartButterflyReverse(userID string) error {
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: user.ID,
		Context: map[string]interface{}{
			"choice_type": "bt_reverse_mode",
			"user_id":     user.ID,
			"can_branch2": butterflyPupa(user) > 0,
			"target_ids":  e.butterflyActionTargetIDs(),
		},
	})
	e.Log(fmt.Sprintf("%s 发动 [倒逆之蝶]：已弃2张牌，请选择发动分支", user.Name))
	return nil
}

func (e *GameEngine) handleButterflyChoiceInput(playerID string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	return dispatchChoiceRouteByType(choiceType, selectionIndex, ctxData, map[string]skillChoiceRouteHandler{
		"bt_dance_mode": func(idx int, data map[string]interface{}) (bool, error) {
			return e.handleButterflyChoiceInputByType(playerID, idx, data)
		},
		"bt_dance_discard": func(idx int, data map[string]interface{}) (bool, error) {
			return e.handleButterflyChoiceInputByType(playerID, idx, data)
		},
		"bt_cocoon_overflow_discard": func(idx int, data map[string]interface{}) (bool, error) {
			return e.handleButterflyChoiceInputByType(playerID, idx, data)
		},
		"bt_reverse_mode": func(idx int, data map[string]interface{}) (bool, error) {
			return e.handleButterflyChoiceInputByType(playerID, idx, data)
		},
		"bt_reverse_target": func(idx int, data map[string]interface{}) (bool, error) {
			return e.handleButterflyChoiceInputByType(playerID, idx, data)
		},
		"bt_reverse_branch2_cost": func(idx int, data map[string]interface{}) (bool, error) {
			return e.handleButterflyChoiceInputByType(playerID, idx, data)
		},
		"bt_reverse_branch2_pick": func(idx int, data map[string]interface{}) (bool, error) {
			return e.handleButterflyChoiceInputByType(playerID, idx, data)
		},
		"bt_pilgrimage_pick": func(idx int, data map[string]interface{}) (bool, error) {
			return e.handleButterflyChoiceInputByType(playerID, idx, data)
		},
		"bt_poison_pick": func(idx int, data map[string]interface{}) (bool, error) {
			return e.handleButterflyChoiceInputByType(playerID, idx, data)
		},
		"bt_mirror_pair": func(idx int, data map[string]interface{}) (bool, error) {
			return e.handleButterflyChoiceInputByType(playerID, idx, data)
		},
		"bt_wither_confirm": func(idx int, data map[string]interface{}) (bool, error) {
			return e.handleButterflyChoiceInputByType(playerID, idx, data)
		},
		"bt_wither_target": func(idx int, data map[string]interface{}) (bool, error) {
			return e.handleButterflyChoiceInputByType(playerID, idx, data)
		},
	})
}

func (e *GameEngine) handleButterflyChoiceInputByType(playerID string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	return dispatchChoiceRouteByType(butterflyChoiceFlow(choiceType), selectionIndex, ctxData, map[string]skillChoiceRouteHandler{
		"dance": func(idx int, data map[string]interface{}) (bool, error) {
			return e.handleButterflyDanceFlow(playerID, idx, data)
		},
		"reverse": func(idx int, data map[string]interface{}) (bool, error) {
			return e.handleButterflyReverseFlow(playerID, idx, data)
		},
		"damage_response": func(idx int, data map[string]interface{}) (bool, error) {
			return e.handleButterflyDamageResponseFlow(playerID, idx, data)
		},
		"wither": func(idx int, data map[string]interface{}) (bool, error) {
			return e.handleButterflyWitherFlow(playerID, idx, data)
		},
	})
}

func butterflyChoiceFlow(choiceType string) string {
	switch choiceType {
	case "bt_dance_mode", "bt_dance_discard", "bt_cocoon_overflow_discard":
		return "dance"
	case "bt_reverse_mode", "bt_reverse_target", "bt_reverse_branch2_cost", "bt_reverse_branch2_pick":
		return "reverse"
	case "bt_pilgrimage_pick", "bt_poison_pick", "bt_mirror_pair":
		return "damage_response"
	case "bt_wither_confirm", "bt_wither_target":
		return "wither"
	default:
		return ""
	}
}

func (e *GameEngine) handleButterflyDanceFlow(playerID string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	return dispatchChoiceRouteByType(choiceType, selectionIndex, ctxData, map[string]skillChoiceRouteHandler{
		"bt_dance_mode": func(idx int, data map[string]interface{}) (bool, error) {
			return e.handleButterflyChoiceInputByTypeLegacy(playerID, idx, data)
		},
		"bt_dance_discard": func(idx int, data map[string]interface{}) (bool, error) {
			return e.handleButterflyChoiceInputByTypeLegacy(playerID, idx, data)
		},
		"bt_cocoon_overflow_discard": func(idx int, data map[string]interface{}) (bool, error) {
			return e.handleButterflyChoiceInputByTypeLegacy(playerID, idx, data)
		},
	})
}

func (e *GameEngine) handleButterflyReverseFlow(playerID string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	return dispatchChoiceRouteByType(choiceType, selectionIndex, ctxData, map[string]skillChoiceRouteHandler{
		"bt_reverse_mode": func(idx int, data map[string]interface{}) (bool, error) {
			return e.handleButterflyChoiceInputByTypeLegacy(playerID, idx, data)
		},
		"bt_reverse_target": func(idx int, data map[string]interface{}) (bool, error) {
			return e.handleButterflyChoiceInputByTypeLegacy(playerID, idx, data)
		},
		"bt_reverse_branch2_cost": func(idx int, data map[string]interface{}) (bool, error) {
			return e.handleButterflyChoiceInputByTypeLegacy(playerID, idx, data)
		},
		"bt_reverse_branch2_pick": func(idx int, data map[string]interface{}) (bool, error) {
			return e.handleButterflyChoiceInputByTypeLegacy(playerID, idx, data)
		},
	})
}

func (e *GameEngine) handleButterflyDamageResponseFlow(playerID string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	return dispatchChoiceRouteByType(choiceType, selectionIndex, ctxData, map[string]skillChoiceRouteHandler{
		"bt_pilgrimage_pick": func(idx int, data map[string]interface{}) (bool, error) {
			return e.handleButterflyChoiceInputByTypeLegacy(playerID, idx, data)
		},
		"bt_poison_pick": func(idx int, data map[string]interface{}) (bool, error) {
			return e.handleButterflyChoiceInputByTypeLegacy(playerID, idx, data)
		},
		"bt_mirror_pair": func(idx int, data map[string]interface{}) (bool, error) {
			return e.handleButterflyChoiceInputByTypeLegacy(playerID, idx, data)
		},
	})
}

func (e *GameEngine) handleButterflyWitherFlow(playerID string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	return dispatchChoiceRouteByType(choiceType, selectionIndex, ctxData, map[string]skillChoiceRouteHandler{
		"bt_wither_confirm": func(idx int, data map[string]interface{}) (bool, error) {
			return e.handleButterflyChoiceInputByTypeLegacy(playerID, idx, data)
		},
		"bt_wither_target": func(idx int, data map[string]interface{}) (bool, error) {
			return e.handleButterflyChoiceInputByTypeLegacy(playerID, idx, data)
		},
	})
}

func (e *GameEngine) handleButterflyChoiceInputByTypeLegacy(playerID string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "bt_dance_mode":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		modes := []string{"draw"}
		if runtimeutil.ToBoolContextValue(ctxData["can_discard"]) {
			modes = append(modes, "discard")
		}
		if selectionIndex < 0 || selectionIndex >= len(modes) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		mode := modes[selectionIndex]
		if mode == "discard" {
			if len(user.Hand) <= 0 {
				return true, fmt.Errorf("手牌不足，无法弃牌")
			}
			ctxData["choice_type"] = "bt_dance_discard"
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return true, nil
		}
		cards, newDeck, newDiscard := rules.DrawCards(e.State.Deck, e.State.DiscardPile, 1)
		e.State.Deck = newDeck
		e.State.DiscardPile = newDiscard
		user.Hand = append(user.Hand, cards...)
		e.NotifyDrawCards(user.ID, len(cards), "bt_dance_draw")
		cocoons, deckAfter, discardAfter := rules.DrawCards(e.State.Deck, e.State.DiscardPile, 1)
		e.State.Deck = deckAfter
		e.State.DiscardPile = discardAfter
		added := addButterflyCocoonCards(user, cocoons)
		e.Log(fmt.Sprintf("%s 发动 [舞动]：摸1张牌，并将牌库顶%d张牌放置为茧", user.Name, added))
		e.checkHandLimit(user, nil)
		overflow := butterflyCocoonCount(user) - butterflyCocoonCapEngine
		if overflow > 0 {
			e.PushInterrupt(&model.Interrupt{Type: model.InterruptChoice, PlayerID: user.ID, Context: map[string]interface{}{"choice_type": "bt_cocoon_overflow_discard", "user_id": user.ID, "discard_count": overflow}})
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if !e.routePendingDamageWithDefaultReturn(model.TurnStageExtraAction) {
				e.enterExtraActionStage()
			}
		}
		return true, nil

	case "bt_dance_discard":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		if selectionIndex < 0 || selectionIndex >= len(user.Hand) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		card := user.Hand[selectionIndex]
		user.Hand = append(user.Hand[:selectionIndex], user.Hand[selectionIndex+1:]...)
		e.NotifyCardRevealed(user.ID, []model.Card{card}, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, card)
		cocoons, deckAfter, discardAfter := rules.DrawCards(e.State.Deck, e.State.DiscardPile, 1)
		e.State.Deck = deckAfter
		e.State.DiscardPile = discardAfter
		added := addButterflyCocoonCards(user, cocoons)
		e.Log(fmt.Sprintf("%s 发动 [舞动]：弃1张牌，并将牌库顶%d张牌放置为茧", user.Name, added))
		overflow := butterflyCocoonCount(user) - butterflyCocoonCapEngine
		if overflow > 0 {
			e.PushInterrupt(&model.Interrupt{Type: model.InterruptChoice, PlayerID: user.ID, Context: map[string]interface{}{"choice_type": "bt_cocoon_overflow_discard", "user_id": user.ID, "discard_count": overflow}})
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.enterExtraActionStage()
		}
		return true, nil

	case "bt_cocoon_overflow_discard":
		if selectionIndex < 0 {
			return true, fmt.Errorf("请先选择要舍弃的茧后再确认")
		}
		return true, e.handleButterflyCocoonOverflowSelections(playerID, []int{selectionIndex})

	case "bt_reverse_mode":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		canBranch2 := butterflyPupa(user) > 0
		modes := []string{"branch1"}
		if canBranch2 {
			modes = append(modes, "branch2")
		}
		if selectionIndex < 0 || selectionIndex >= len(modes) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		if modes[selectionIndex] == "branch1" {
			ctxData["choice_type"] = "bt_reverse_target"
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return true, nil
		}
		if butterflyPupa(user) <= 0 {
			return true, fmt.Errorf("蛹不足，无法发动分支②")
		}
		ctxData["choice_type"] = "bt_reverse_branch2_cost"
		ctxData["can_remove_cocoon"] = butterflyCocoonCount(user) >= 2
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil

	case "bt_reverse_target":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		var targetIDs []string
		if arr, ok := ctxData["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := ctxData["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		targetID := targetIDs[selectionIndex]
		e.AddPendingDamage(model.PendingDamage{SourceID: user.ID, TargetID: targetID, Damage: 1, DamageType: "magic", IgnoreHeal: true})
		if target := e.State.Players[targetID]; target != nil {
			e.Log(fmt.Sprintf("%s 的 [倒逆之蝶] 分支①：对 %s 造成1点不可治疗抵御的法术伤害", user.Name, target.Name))
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if !e.routePendingDamageWithReturn(model.TurnStageExtraAction) {
				e.enterExtraActionStage()
			}
		}
		return true, nil

	case "bt_reverse_branch2_cost":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		canRemove := runtimeutil.ToBoolContextValue(ctxData["can_remove_cocoon"])
		modes := []string{}
		if canRemove {
			modes = append(modes, "remove_cocoon")
		}
		modes = append(modes, "self_damage")
		if selectionIndex < 0 || selectionIndex >= len(modes) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		if modes[selectionIndex] == "remove_cocoon" {
			ctxData["choice_type"] = "bt_reverse_branch2_pick"
			delete(ctxData, "remaining_indices")
			delete(ctxData, "selected_indices")
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return true, nil
		}
		e.AddPendingDamage(model.PendingDamage{SourceID: user.ID, TargetID: user.ID, Damage: 4, DamageType: "magic"})
		now := addButterflyPupa(user, -1)
		e.Log(fmt.Sprintf("%s 的 [倒逆之蝶] 分支②：对自己造成4点法术伤害并移除1个蛹（当前蛹=%d）", user.Name, now))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if !e.routePendingDamageWithReturn(model.TurnStageExtraAction) {
				e.enterExtraActionStage()
			}
		}
		return true, nil

	case "bt_reverse_branch2_pick":
		if selectionIndex < 0 {
			return true, fmt.Errorf("请先选择要移除的茧后再确认")
		}
		return true, e.handleButterflyReverseBranch2PickSelections(playerID, []int{selectionIndex})

	case "bt_pilgrimage_pick", "bt_poison_pick":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		var cocoonIndices []int
		if arr, ok := ctxData["cocoon_indices"].([]int); ok {
			cocoonIndices = append(cocoonIndices, arr...)
		} else if arr, ok := ctxData["cocoon_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					cocoonIndices = append(cocoonIndices, int(f))
				}
			}
		}
		if selectionIndex == -1 || selectionIndex == 0 {
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				if !e.routePendingDamageWithDefaultReturn(nil) {
					e.enterExtraActionStage()
				}
			}
			return true, nil
		}
		pickIdx := -1
		if selectionIndex >= 1 && selectionIndex <= len(cocoonIndices) {
			pickIdx = cocoonIndices[selectionIndex-1]
		} else if idx, ok := runtimeutil.ResolveSelectionToCandidate(selectionIndex, cocoonIndices); ok {
			pickIdx = idx
		}
		if pickIdx < 0 {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		removed, ok := removeButterflyCocoonByFieldIndex(user, pickIdx)
		if !ok {
			return true, fmt.Errorf("选择的茧无效")
		}
		e.NotifyCardRevealed(user.ID, []model.Card{removed}, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, removed)
		damageIdx := runtimeutil.ToIntContextValue(ctxData["damage_index"])
		if damageIdx < 0 || damageIdx >= len(e.State.PendingDamageQueue) {
			return true, fmt.Errorf("伤害上下文不存在")
		}
		pd := &e.State.PendingDamageQueue[damageIdx]
		if choiceType == "bt_pilgrimage_pick" {
			if pd.Damage > 0 {
				pd.Damage--
			}
			e.Log(fmt.Sprintf("%s 发动 [朝圣]：移除1个茧，抵御1点伤害（剩余伤害=%d）", user.Name, pd.Damage))
		} else {
			pd.Damage++
			e.Log(fmt.Sprintf("%s 发动 [毒粉]：移除1个茧，本次法术伤害+1（当前伤害=%d）", user.Name, pd.Damage))
		}
		if removed.Type == model.CardTypeMagic {
			e.queueButterflyWitherTrigger(user)
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.enterDamageResolution(nil)
		}
		return true, nil

	case "bt_mirror_pair":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		if selectionIndex == -1 || selectionIndex == 0 {
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				if !e.routePendingDamageWithDefaultReturn(nil) {
					e.enterExtraActionStage()
				}
			}
			return true, nil
		}
		var pairDefs []string
		if arr, ok := ctxData["pair_defs"].([]string); ok {
			pairDefs = append(pairDefs, arr...)
		} else if arr, ok := ctxData["pair_defs"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					pairDefs = append(pairDefs, s)
				}
			}
		}
		pairChoice := -1
		if selectionIndex >= 1 && selectionIndex <= len(pairDefs) {
			pairChoice = selectionIndex - 1
		} else if selectionIndex >= 0 && selectionIndex < len(pairDefs) {
			pairChoice = selectionIndex
		}
		if pairChoice < 0 || pairChoice >= len(pairDefs) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		parts := strings.Split(pairDefs[pairChoice], ",")
		if len(parts) != 2 {
			return true, fmt.Errorf("镜花水月配对参数无效")
		}
		left, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
		right, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err1 != nil || err2 != nil {
			return true, fmt.Errorf("镜花水月配对索引无效")
		}
		removed, err := removeButterflyCocoonByFieldIndices(user, []int{left, right})
		if err != nil {
			return true, err
		}
		if len(removed) != 2 || removed[0].Element != removed[1].Element {
			return true, fmt.Errorf("镜花水月需要移除2张同系茧")
		}
		e.NotifyCardRevealed(user.ID, removed, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, removed...)
		damageIdx := runtimeutil.ToIntContextValue(ctxData["damage_index"])
		if damageIdx < 0 || damageIdx >= len(e.State.PendingDamageQueue) {
			return true, fmt.Errorf("伤害上下文不存在")
		}
		pd := &e.State.PendingDamageQueue[damageIdx]
		originSourceID := pd.SourceID
		pd.Damage = 0
		e.AddPendingDamage(model.PendingDamage{SourceID: user.ID, TargetID: originSourceID, Damage: 1, DamageType: "magic"})
		e.AddPendingDamage(model.PendingDamage{SourceID: user.ID, TargetID: originSourceID, Damage: 1, DamageType: "magic"})
		for _, c := range removed {
			if c.Type == model.CardTypeMagic {
				e.queueButterflyWitherTrigger(user)
			}
		}
		if target := e.State.Players[originSourceID]; target != nil {
			e.Log(fmt.Sprintf("%s 发动 [镜花水月]：抵御原伤害，并改为对 %s 造成2次1点法术伤害", user.Name, target.Name))
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.enterDamageResolution(nil)
		}
		return true, nil

	case "bt_wither_confirm":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		if user.Tokens == nil {
			user.Tokens = map[string]int{}
		}
		if selectionIndex == 0 {
			ctxData["choice_type"] = "bt_wither_target"
			ctxData["target_ids"] = e.butterflyActionTargetIDs()
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return true, nil
		}
		if selectionIndex != 1 {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		if user.TurnState.SkillFlowState["bt_wither_pending"] > 0 {
			user.TurnState.SkillFlowState["bt_wither_pending"]--
		}
		if user.TurnState.SkillFlowState["bt_wither_pending"] > 0 {
			ctxData["choice_type"] = "bt_wither_confirm"
			ctxData["target_ids"] = e.butterflyActionTargetIDs()
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return true, nil
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if !e.routePendingDamageWithDefaultReturn(nil) {
				e.enterExtraActionStage()
			}
		}
		return true, nil

	case "bt_wither_target":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		var targetIDs []string
		if arr, ok := ctxData["target_ids"].([]string); ok {
			targetIDs = append(targetIDs, arr...)
		} else if arr, ok := ctxData["target_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					targetIDs = append(targetIDs, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		targetID := targetIDs[selectionIndex]
		e.AddPendingDamage(model.PendingDamage{SourceID: user.ID, TargetID: targetID, Damage: 1, DamageType: "magic"})
		e.AddPendingDamage(model.PendingDamage{SourceID: user.ID, TargetID: user.ID, Damage: 2, DamageType: "magic"})
		if user.Tokens == nil {
			user.Tokens = map[string]int{}
		}
		user.Tokens["bt_wither_active"] = 1
		if user.TurnState.SkillFlowState["bt_wither_pending"] > 0 {
			user.TurnState.SkillFlowState["bt_wither_pending"]--
		}
		if target := e.State.Players[targetID]; target != nil {
			e.Log(fmt.Sprintf("%s 发动 [凋零]：对 %s 造成1点法术伤害，并对自己造成2点法术伤害；对方士气最低为1直到其下回合开始前", user.Name, target.Name))
		}
		if user.TurnState.SkillFlowState["bt_wither_pending"] > 0 {
			ctxData["choice_type"] = "bt_wither_confirm"
			ctxData["target_ids"] = e.butterflyActionTargetIDs()
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return true, nil
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.enterDamageResolution(nil)
		}
		return true, nil
	}
	return false, nil
}

func (e *GameEngine) handleButterflyCocoonOverflowSelections(playerID string, selections []int) error {
	if e.State.PendingInterrupt == nil || e.State.PendingInterrupt.Type != model.InterruptChoice {
		return fmt.Errorf("当前不存在可处理的茧上限弃置")
	}
	ctxData, ok := e.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		return fmt.Errorf("茧上限上下文错误")
	}
	choiceType, _ := ctxData["choice_type"].(string)
	if choiceType != "bt_cocoon_overflow_discard" {
		return fmt.Errorf("当前中断不是茧上限弃置")
	}
	userID, _ := ctxData["user_id"].(string)
	if userID == "" {
		userID = playerID
	}
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	discardNeed := runtimeutil.ToIntContextValue(ctxData["discard_count"])
	if discardNeed < 0 {
		discardNeed = 0
	}
	cocoonIndices := butterflyCocoonFieldIndices(user)
	if discardNeed > len(cocoonIndices) {
		discardNeed = len(cocoonIndices)
	}
	if len(selections) != discardNeed {
		return fmt.Errorf("需要选择 %d 个茧舍弃", discardNeed)
	}
	picked := make([]int, 0, len(selections))
	seen := make(map[int]struct{}, len(selections))
	for _, selection := range selections {
		fieldIdx, ok := runtimeutil.ResolveSelectionToCandidate(selection, cocoonIndices)
		if !ok {
			return fmt.Errorf("无效的茧索引: %d", selection)
		}
		if _, dup := seen[fieldIdx]; dup {
			return fmt.Errorf("不能重复选择同一个茧")
		}
		seen[fieldIdx] = struct{}{}
		picked = append(picked, fieldIdx)
	}
	removed, err := removeButterflyCocoonByFieldIndices(user, picked)
	if err != nil {
		return err
	}
	if len(removed) > 0 {
		e.NotifyCardHidden(user.ID, removed, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, removed...)
	}
	e.Log(fmt.Sprintf("%s 的 [茧上限] 结算：舍弃%d个茧", user.Name, len(removed)))
	e.PopInterrupt()
	if e.State.PendingInterrupt == nil {
		e.enterExtraActionStage()
	}
	return nil
}

func (e *GameEngine) handleButterflyReverseBranch2PickSelections(playerID string, selections []int) error {
	if e.State.PendingInterrupt == nil || e.State.PendingInterrupt.Type != model.InterruptChoice {
		return fmt.Errorf("当前不存在可处理的倒逆之蝶分支②选茧")
	}
	ctxData, ok := e.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		return fmt.Errorf("倒逆之蝶分支②上下文错误")
	}
	choiceType, _ := ctxData["choice_type"].(string)
	if choiceType != "bt_reverse_branch2_pick" {
		return fmt.Errorf("当前中断不是倒逆之蝶分支②选茧")
	}
	userID, _ := ctxData["user_id"].(string)
	if userID == "" {
		userID = playerID
	}
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	const pickNeed = 2
	if len(selections) != pickNeed {
		return fmt.Errorf("需要选择 %d 个茧", pickNeed)
	}
	cocoonIndices := butterflyCocoonFieldIndices(user)
	picked := make([]int, 0, len(selections))
	seen := make(map[int]struct{}, len(selections))
	for _, selection := range selections {
		fieldIdx, ok := runtimeutil.ResolveSelectionToCandidate(selection, cocoonIndices)
		if !ok {
			return fmt.Errorf("无效的茧索引: %d", selection)
		}
		if _, dup := seen[fieldIdx]; dup {
			return fmt.Errorf("不能重复选择同一个茧")
		}
		seen[fieldIdx] = struct{}{}
		picked = append(picked, fieldIdx)
	}
	removed, err := removeButterflyCocoonByFieldIndices(user, picked)
	if err != nil {
		return err
	}
	if len(removed) > 0 {
		e.NotifyCardRevealed(user.ID, removed, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, removed...)
	}
	for _, c := range removed {
		if c.Type == model.CardTypeMagic {
			e.queueButterflyWitherTrigger(user)
		}
	}
	now := addButterflyPupa(user, -1)
	e.Log(fmt.Sprintf("%s 的 [倒逆之蝶] 分支②：移除2个茧并移除1个蛹（当前蛹=%d）", user.Name, now))
	e.PopInterrupt()
	if e.State.PendingInterrupt == nil {
		if !e.routePendingDamageWithReturn(model.TurnStageExtraAction) {
			e.enterExtraActionStage()
		}
	}
	return nil
}
