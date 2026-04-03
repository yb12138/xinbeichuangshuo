package engine

import (
	"fmt"
	"starcup-engine/internal/engine/runtimeutil"
	"strconv"

	"starcup-engine/internal/model"
)

func (e *GameEngine) buildValkyrieChoicePrompt(choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "valkyrie_military_glory_mode":
		maxX := runtimeutil.ToIntContextValue(data["max_x"])
		options := []model.PromptOption{
			{ID: "0", Label: "你+1治疗并脱离英灵形态"},
		}
		if maxX > 0 {
			options = append(options, model.PromptOption{
				ID:    "1",
				Label: fmt.Sprintf("移除我方战绩区星石（1~%d）并指定角色+X治疗", maxX),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【军威神光】请选择效果：",
			Options:  options,
			Min:      1,
			Max:      1,
		}

	case "valkyrie_military_glory_x":
		maxX := runtimeutil.ToIntContextValue(data["max_x"])
		options := make([]model.PromptOption, 0, maxX)
		for x := 1; x <= maxX; x++ {
			options = append(options, model.PromptOption{
				ID:    strconv.Itoa(x),
				Label: fmt.Sprintf("X=%d", x),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【军威神光】请选择X：",
			Options:  options,
			Min:      1,
			Max:      1,
		}

	case "valkyrie_military_glory_target":
		targetIDs := runtimeutil.ParseStringSliceContextValue(data["target_ids"])
		options := make([]model.PromptOption, 0, len(targetIDs))
		for _, targetID := range targetIDs {
			if target := e.State.Players[targetID]; target != nil {
				options = append(options, model.PromptOption{
					ID:    targetID,
					Label: target.Name,
				})
			}
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【军威神光】请选择目标角色：",
			Options:  options,
			Min:      1,
			Max:      1,
		}

	case "valkyrie_heroic_extra_confirm":
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【英灵召唤】是否额外弃1张法术牌并令当前战斗目标+1治疗？",
			Options: []model.PromptOption{
				{ID: "0", Label: "是"},
				{ID: "1", Label: "否"},
			},
			Min: 1,
			Max: 1,
		}

	case "valkyrie_heroic_discard_card":
		if player == nil {
			return nil
		}
		magicIndices := parseIntSliceContextValue(data["magic_indices"])
		options := make([]model.PromptOption, 0, len(magicIndices))
		for _, idx := range magicIndices {
			if idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    strconv.Itoa(idx),
				Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(player.Hand[idx])),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【英灵召唤】请选择要额外弃置的1张法术牌：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	}

	return nil
}

func (e *GameEngine) handleValkyrieChoiceInput(_ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)

	switch choiceType {
	case "valkyrie_military_glory_mode":
		userID, _ := ctxData["user_id"].(string)
		camp, _ := ctxData["camp"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		if selectionIndex == 0 {
			e.Heal(userID, 1)
			if user.Tokens == nil {
				user.Tokens = map[string]int{}
			}
			user.Tokens["valkyrie_spirit"] = 0
			e.Log(fmt.Sprintf("%s 选择军威神光选项1：+1治疗并脱离英灵形态", user.Name))
			e.PopInterrupt()
			return true, nil
		}
		if selectionIndex == 1 {
			maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
			if maxX <= 0 {
				return true, fmt.Errorf("当前阵营无可用能量")
			}
			e.State.PendingInterrupt.Context = map[string]interface{}{
				"choice_type": "valkyrie_military_glory_x",
				"user_id":     userID,
				"camp":        camp,
				"max_x":       maxX,
			}
			e.notifyInterruptPrompt()
			return true, nil
		}
		return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)

	case "valkyrie_military_glory_x":
		userID, _ := ctxData["user_id"].(string)
		camp, _ := ctxData["camp"].(string)
		maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
		if maxX <= 0 {
			return true, fmt.Errorf("当前阵营无可用能量")
		}
		x := selectionIndex + 1
		if x <= 0 || x > maxX || x >= 3 {
			return true, fmt.Errorf("无效的X值")
		}
		targetIDs := make([]string, 0, len(e.State.PlayerOrder))
		for _, pid := range e.State.PlayerOrder {
			if e.State.Players[pid] != nil {
				targetIDs = append(targetIDs, pid)
			}
		}
		e.State.PendingInterrupt.Context = map[string]interface{}{
			"choice_type": "valkyrie_military_glory_target",
			"user_id":     userID,
			"camp":        camp,
			"x":           x,
			"target_ids":  targetIDs,
		}
		e.notifyInterruptPrompt()
		return true, nil

	case "valkyrie_military_glory_target":
		userID, _ := ctxData["user_id"].(string)
		camp, _ := ctxData["camp"].(string)
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
		x := runtimeutil.ToIntContextValue(ctxData["x"])
		if x <= 0 || x >= 3 {
			return true, fmt.Errorf("无效的X值")
		}
		total := e.GetCampCrystals(camp) + e.GetCampGems(camp)
		if x > total {
			return true, fmt.Errorf("阵营能量不足")
		}
		useCrystal := x
		if crystals := e.GetCampCrystals(camp); useCrystal > crystals {
			useCrystal = crystals
		}
		if useCrystal > 0 {
			e.ModifyCrystal(camp, -useCrystal)
		}
		if remain := x - useCrystal; remain > 0 {
			e.ModifyGem(camp, -remain)
		}
		e.Heal(targetID, x)
		e.Log(fmt.Sprintf("%s 选择军威神光选项2：移除%d星石并使 %s +%d治疗", user.Name, x, target.Name, x))
		e.PopInterrupt()
		return true, nil

	case "valkyrie_heroic_extra_confirm":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		if selectionIndex == 1 {
			e.PopInterrupt()
			e.resumePendingAttackHit(ctxData)
			return true, nil
		}
		if selectionIndex == 0 {
			magicIndices := make([]int, 0)
			for i, card := range user.Hand {
				if card.Type == model.CardTypeMagic {
					magicIndices = append(magicIndices, i)
				}
			}
			if len(magicIndices) == 0 {
				e.PopInterrupt()
				e.resumePendingAttackHit(ctxData)
				return true, nil
			}
			e.State.PendingInterrupt.Context = map[string]interface{}{
				"choice_type":   "valkyrie_heroic_discard_card",
				"user_id":       userID,
				"magic_indices": magicIndices,
				"user_ctx":      ctxData["user_ctx"],
			}
			e.notifyInterruptPrompt()
			return true, nil
		}
		return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)

	case "valkyrie_heroic_discard_card":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		magicIndices := parseIntSliceContextValue(ctxData["magic_indices"])
		cardIdx, ok := runtimeutil.ResolveSelectionToCandidate(selectionIndex, magicIndices)
		if !ok {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		if cardIdx < 0 || cardIdx >= len(user.Hand) || user.Hand[cardIdx].Type != model.CardTypeMagic {
			return true, fmt.Errorf("请选择法术牌")
		}
		card := user.Hand[cardIdx]
		e.NotifyCardRevealed(userID, []model.Card{card}, "discard")
		user.Hand = append(user.Hand[:cardIdx], user.Hand[cardIdx+1:]...)
		e.State.DiscardPile = append(e.State.DiscardPile, card)

		rawCtx, _ := ctxData["user_ctx"].(*model.Context)
		targetID := ""
		if rawCtx != nil && rawCtx.TriggerCtx != nil {
			targetID = rawCtx.TriggerCtx.TargetID
		}
		if targetID == "" && rawCtx != nil && rawCtx.Target != nil {
			targetID = rawCtx.Target.ID
		}
		if targetID != "" {
			e.Heal(targetID, 1)
			if target := e.State.Players[targetID]; target != nil {
				e.Log(fmt.Sprintf("%s 因英灵召唤额外效果，获得1点治疗", target.Name))
			}
		}
		e.PopInterrupt()
		e.resumePendingAttackHit(ctxData)
		return true, nil
	}

	return false, nil
}
