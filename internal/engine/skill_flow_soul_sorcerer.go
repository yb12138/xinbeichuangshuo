// gameflow: 灵魂术士：灵魂链接、吞噬、魔魂等同调多步中断。

package engine

import (
	"fmt"
	"starcup-engine/internal/engine/core/runtimeutil"

	"starcup-engine/internal/model"
)

func (e *GameEngine) buildSoulSorcererChoicePrompt(choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "ss_convert_color":
		var modeOrder []string
		if arr, ok := data["mode_order"].([]string); ok {
			modeOrder = append(modeOrder, arr...)
		} else if arr, ok := data["mode_order"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					modeOrder = append(modeOrder, s)
				}
			}
		}
		var options []model.PromptOption
		for _, mode := range modeOrder {
			switch mode {
			case "y2b":
				options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "黄魂 -> 蓝魂（转换1点）"})
			case "b2y":
				options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "蓝魂 -> 黄魂（转换1点）"})
			}
		}
		return &model.Prompt{
			Type:       model.PromptConfirm,
			PlayerID:   playerID,
			ChoiceType: choiceType,
			Message:    "【灵魂转换】请选择转换方向：",
			Options:    options,
			Min:        1,
			Max:        1,
		}

	case "ss_link_target":
		var allyIDs []string
		if arr, ok := data["ally_ids"].([]string); ok {
			allyIDs = append(allyIDs, arr...)
		} else if arr, ok := data["ally_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					allyIDs = append(allyIDs, s)
				}
			}
		}
		var options []model.PromptOption
		for _, aid := range allyIDs {
			if p := e.State.Players[aid]; p != nil {
				options = append(options, model.PromptOption{ID: aid, Label: p.Name})
			}
		}
		return &model.Prompt{
			Type:       model.PromptConfirm,
			PlayerID:   playerID,
			ChoiceType: choiceType,
			Message:    "【灵魂链接】请选择要放置灵魂链接的队友：",
			Options:    options,
			Min:        1,
			Max:        1,
		}

	case "ss_link_transfer_x":
		maxX := runtimeutil.ToIntContextValue(data["max_x"])
		if maxX < 0 {
			maxX = 0
		}
		options := make([]model.PromptOption, 0, maxX+1)
		for x := 0; x <= maxX; x++ {
			label := fmt.Sprintf("移除%d点蓝魂并转移%d点伤害", x, x)
			if x == 0 {
				label = "不转移伤害"
			}
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", x), Label: label})
		}
		return &model.Prompt{
			Type:       model.PromptConfirm,
			PlayerID:   playerID,
			ChoiceType: choiceType,
			Message:    "【灵魂链接】请选择要转移的伤害点数X：",
			Options:    options,
			Min:        1,
			Max:        1,
		}

	case "ss_recall_pick":
		var magicIndices []int
		if arr, ok := data["magic_indices"].([]int); ok {
			magicIndices = append(magicIndices, arr...)
		} else if arr, ok := data["magic_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					magicIndices = append(magicIndices, int(f))
				}
			}
		}
		if len(magicIndices) == 0 {
			if arr, ok := data["remaining_indices"].([]int); ok {
				magicIndices = append(magicIndices, arr...)
			} else if arr, ok := data["remaining_indices"].([]interface{}); ok {
				for _, v := range arr {
					if f, ok := v.(float64); ok {
						magicIndices = append(magicIndices, int(f))
					}
				}
			}
		}
		var options []model.PromptOption
		for _, idx := range magicIndices {
			if player == nil || idx < 0 || idx >= len(player.Hand) {
				continue
			}
			if player.Hand[idx].Type != model.CardTypeMagic {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(player.Hand[idx])),
			})
		}
		maxSelect := len(options)
		if maxSelect < 1 {
			maxSelect = 1
		}
		return &model.Prompt{
			Type:       model.PromptChooseCards,
			PlayerID:   playerID,
			ChoiceType: choiceType,
			Message:    "【灵魂召还】请选择要弃置的法术牌（至少1张）：",
			Options:    options,
			Min:        1,
			Max:        maxSelect,
		}
	}

	return nil
}

func (e *GameEngine) handleSoulSorcererChoiceInput(playerID string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "ss_convert_color":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		var modeOrder []string
		if arr, ok := ctxData["mode_order"].([]string); ok {
			modeOrder = append(modeOrder, arr...)
		} else if arr, ok := ctxData["mode_order"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					modeOrder = append(modeOrder, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(modeOrder) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		mode := modeOrder[selectionIndex]
		switch mode {
		case "y2b":
			if soulSorcererYellow(user) <= 0 {
				return true, fmt.Errorf("黄色灵魂不足")
			}
			if soulSorcererBlue(user) >= soulSorcererBlueCapEngine {
				return true, fmt.Errorf("蓝色灵魂已满")
			}
			addSoulSorcererYellow(user, -1)
			addSoulSorcererBlue(user, 1)
			e.Log(fmt.Sprintf("%s 的 [灵魂转换] 生效：黄魂-1，蓝魂+1（黄:%d 蓝:%d）", user.Name, soulSorcererYellow(user), soulSorcererBlue(user)))
		case "b2y":
			if soulSorcererBlue(user) <= 0 {
				return true, fmt.Errorf("蓝色灵魂不足")
			}
			if soulSorcererYellow(user) >= soulSorcererYellowCapEngine {
				return true, fmt.Errorf("黄色灵魂已满")
			}
			addSoulSorcererBlue(user, -1)
			addSoulSorcererYellow(user, 1)
			e.Log(fmt.Sprintf("%s 的 [灵魂转换] 生效：蓝魂-1，黄魂+1（黄:%d 蓝:%d）", user.Name, soulSorcererYellow(user), soulSorcererBlue(user)))
		default:
			return true, fmt.Errorf("无效的灵魂转换模式")
		}
		rawCtx, _ := ctxData["user_ctx"].(*model.Context)
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if len(e.State.ActionQueue) > 0 {
				e.enterActionExecutionStage()
			} else if len(e.State.CombatStack) > 0 {
				if e.State.CombatStage == model.CombatStageNone {
					e.setCombatStage(model.CombatStageHitCheck)
				}
				e.clearSubflow()
			} else if len(e.State.PendingDamageQueue) > 0 {
				e.enterDamageResolution(nil)
			} else if e.State.ReturnTurnStage != "" || e.State.ReturnCombatStage != model.CombatStageNone || e.State.ReturnSubflow != model.SubflowNone {
				e.restoreReturnPoint()
			} else if rawCtx != nil && rawCtx.AttackDeclaredPhase() {
				e.enterExtraActionStage()
			} else {
				e.enterExtraActionStage()
			}
		}
		return true, nil

	case "ss_link_target":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		var allyIDs []string
		if arr, ok := ctxData["ally_ids"].([]string); ok {
			allyIDs = append(allyIDs, arr...)
		} else if arr, ok := ctxData["ally_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					allyIDs = append(allyIDs, s)
				}
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(allyIDs) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		target := e.State.Players[allyIDs[selectionIndex]]
		if target == nil {
			return true, fmt.Errorf("目标队友不存在")
		}
		if target.Camp != user.Camp || target.ID == user.ID {
			return true, fmt.Errorf("灵魂链接只能指定队友")
		}
		if soulSorcererYellow(user) < 1 || soulSorcererBlue(user) < 1 {
			return true, fmt.Errorf("灵魂不足，无法放置灵魂链接")
		}
		if user.Character == nil {
			return true, fmt.Errorf("角色信息缺失")
		}
		linkCard, ok := user.ConsumeExclusiveCard(user.Character.ID, "灵魂链接")
		if !ok {
			return true, fmt.Errorf("未找到【灵魂链接】专属技能卡")
		}
		addSoulSorcererYellow(user, -1)
		addSoulSorcererBlue(user, -1)
		if err := e.placeSoulLink(user, target, linkCard); err != nil {
			user.RestoreExclusiveCard(linkCard)
			addSoulSorcererYellow(user, 1)
			addSoulSorcererBlue(user, 1)
			return true, err
		}
		e.Log(fmt.Sprintf("%s 发动 [灵魂链接]：移除1黄魂+1蓝魂，并将灵魂链接放置于 %s 面前", user.Name, target.Name))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.setTurnStage(model.TurnStageActionStart)
			e.clearCombatStage()
			e.clearSubflow()
		}
		return true, nil

	case "ss_link_transfer_x":
		sorcererID, _ := ctxData["sorcerer_id"].(string)
		sorcerer := e.State.Players[sorcererID]
		if sorcerer == nil {
			return true, fmt.Errorf("灵魂术士不存在")
		}
		damageIdx := runtimeutil.ToIntContextValue(ctxData["damage_index"])
		if damageIdx < 0 || damageIdx >= len(e.State.PendingDamageQueue) {
			return true, fmt.Errorf("伤害上下文不存在")
		}
		pd := &e.State.PendingDamageQueue[damageIdx]
		sourceID, _ := ctxData["source_id"].(string)
		targetID, _ := ctxData["target_id"].(string)
		if sourceID != "" && sourceID != pd.SourceID {
			return true, fmt.Errorf("伤害来源已变化")
		}
		if targetID != "" && targetID != pd.TargetID {
			return true, fmt.Errorf("伤害目标已变化")
		}
		maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
		if maxX < 0 {
			maxX = 0
		}
		x := selectionIndex
		if x < 0 || x > maxX {
			return true, fmt.Errorf("无效的X值")
		}
		if x > soulSorcererBlue(sorcerer) {
			x = soulSorcererBlue(sorcerer)
		}
		if x > pd.Damage {
			x = pd.Damage
		}
		counterpartID, _ := ctxData["counterpart_id"].(string)
		counterpart := e.State.Players[counterpartID]
		if x > 0 && counterpart != nil {
			addSoulSorcererBlue(sorcerer, -x)
			pd.Damage -= x
			if pd.Damage < 0 {
				pd.Damage = 0
			}
			e.AddPendingDamage(model.PendingDamage{
				SourceID:   pd.SourceID,
				TargetID:   counterpart.ID,
				Damage:     x,
				DamageType: model.MagicAttack,
				Checks: map[model.PendingDamageCheckKey]bool{
					model.PendingDamageCheckFromSoulLink: true,
				},
			})
			e.Log(fmt.Sprintf("%s 的 [灵魂链接] 生效：移除%d点蓝魂，将%d点伤害转移给 %s（法术伤害）", sorcerer.Name, x, x, counterpart.Name))
		} else {
			e.Log(fmt.Sprintf("%s 的 [灵魂链接] 选择不转移伤害", sorcerer.Name))
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.enterDamageResolution(nil)
		}
		return true, nil

	case "ss_recall_pick":
		if selectionIndex < 0 {
			return true, fmt.Errorf("请从可选法术牌中至少选择1张")
		}
		return true, e.handleSoulRecallSelections(playerID, []int{selectionIndex})
	}

	return false, nil
}

func (e *GameEngine) handleSoulRecallSelections(playerID string, selections []int) error {
	if e.State.PendingInterrupt == nil || e.State.PendingInterrupt.Type != model.InterruptChoice {
		return fmt.Errorf("当前不存在可处理的灵魂召还弃牌")
	}
	ctxData, ok := e.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		return fmt.Errorf("灵魂召还上下文错误")
	}
	choiceType, _ := ctxData["choice_type"].(string)
	if choiceType != "ss_recall_pick" {
		return fmt.Errorf("当前中断不是灵魂召还选牌")
	}

	userID, _ := ctxData["user_id"].(string)
	if userID == "" {
		userID = playerID
	}
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}

	magicIndices := runtimeutil.ParseChoiceIntSlice(ctxData["magic_indices"])
	if len(magicIndices) == 0 {
		magicIndices = runtimeutil.ParseChoiceIntSlice(ctxData["remaining_indices"])
	}
	allowed := make(map[int]struct{}, len(magicIndices))
	orderedCandidates := make([]int, 0, len(magicIndices))
	for _, idx := range magicIndices {
		if idx < 0 || idx >= len(user.Hand) {
			continue
		}
		if user.Hand[idx].Type != model.CardTypeMagic {
			continue
		}
		allowed[idx] = struct{}{}
		orderedCandidates = append(orderedCandidates, idx)
	}
	if len(allowed) == 0 {
		return fmt.Errorf("灵魂召还没有可弃置的法术牌")
	}
	if len(selections) == 0 {
		return fmt.Errorf("灵魂召还至少选择1张法术牌")
	}
	if len(selections) > len(allowed) {
		return fmt.Errorf("选择数量超过可选法术牌数量")
	}

	picked := make([]int, 0, len(selections))
	seen := make(map[int]struct{}, len(selections))
	for _, idx := range selections {
		resolvedIdx, ok := runtimeutil.ResolveSelectionToAllowedIndex(idx, orderedCandidates, allowed)
		if !ok {
			return fmt.Errorf("灵魂召还只能选择法术牌")
		}
		if _, dup := seen[resolvedIdx]; dup {
			return fmt.Errorf("不能重复选择同一张牌")
		}
		seen[resolvedIdx] = struct{}{}
		picked = append(picked, resolvedIdx)
	}

	removed, err := removeCardsByIndicesFromHand(user, picked)
	if err != nil {
		return err
	}
	e.NotifyCardRevealed(user.ID, removed, "discard")
	e.State.DiscardPile = append(e.State.DiscardPile, removed...)
	gain := len(removed)
	before := soulSorcererBlue(user)
	after := addSoulSorcererBlue(user, gain)
	e.Log(fmt.Sprintf("%s 发动 [灵魂召还]：弃置%d张法术牌，蓝色灵魂 +%d（%d→%d）", user.Name, gain, gain, before, after))

	e.PopInterrupt()
	if e.State.PendingInterrupt == nil {
		if !e.routePendingDamageWithReturn(model.TurnStageExtraAction) {
			e.enterExtraActionStage()
		}
	}
	return nil
}
