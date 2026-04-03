package engine

import (
	"fmt"
	"sort"
	"starcup-engine/internal/engine/runtimeutil"

	"starcup-engine/internal/model"
)

func buildBloodPriestessDeferredFollowupHandlers() map[string]deferredFollowupHandler {
	return map[string]deferredFollowupHandler{
		"blood_priestess_shared_life_place": {
			label:   "BloodPriestess",
			resolve: (*GameEngine).resolveBloodPriestessSharedLifePlaceFollowup,
		},
		"blood_priestess_blood_sorrow_apply": {
			label:   "BloodPriestess",
			resolve: (*GameEngine).resolveBloodPriestessBloodSorrowFollowup,
		},
		"blood_priestess_curse_discard": {
			label:   "BloodPriestess",
			resolve: (*GameEngine).resolveBloodPriestessCurseDiscardFollowup,
		},
	}
}

func (e *GameEngine) buildBloodPriestessChoicePrompt(choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "bp_shared_life_target":
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
		return &model.Prompt{
			Type:       model.PromptConfirm,
			PlayerID:   playerID,
			ChoiceType: choiceType,
			Message:    "【同生共死】请选择放置目标：",
			Options:    options,
			Min:        1,
			Max:        1,
		}

	case "bp_blood_sorrow_mode":
		return &model.Prompt{
			Type:       model.PromptConfirm,
			PlayerID:   playerID,
			ChoiceType: choiceType,
			Message:    "【血之哀伤】请选择后续效果（结算时会先对自己造成2点法术伤害）：",
			Options: []model.PromptOption{
				{ID: "0", Label: "移除同生共死"},
				{ID: "1", Label: "转移同生共死目标"},
			},
			Min: 1,
			Max: 1,
		}

	case "bp_blood_sorrow_target":
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
		return &model.Prompt{
			Type:       model.PromptConfirm,
			PlayerID:   playerID,
			ChoiceType: choiceType,
			Message:    "【血之哀伤】请选择新的同生共死目标：",
			Options:    options,
			Min:        1,
			Max:        1,
		}

	case "bp_blood_wail_x":
		return &model.Prompt{
			Type:       model.PromptConfirm,
			PlayerID:   playerID,
			ChoiceType: choiceType,
			Message:    "【血之悲鸣】请选择X值：",
			Options: []model.PromptOption{
				{ID: "0", Label: "X=0（伤害=1）"},
				{ID: "1", Label: "X=1（伤害=2）"},
				{ID: "2", Label: "X=2（伤害=3）"},
			},
			Min: 1,
			Max: 1,
		}

	case "bp_curse_discard":
		discardCount := runtimeutil.ToIntContextValue(data["discard_count"])
		if discardCount < 0 {
			discardCount = 0
		}
		if discardCount > len(player.Hand) {
			discardCount = len(player.Hand)
		}
		options := make([]model.PromptOption, 0, len(player.Hand))
		for idx, card := range player.Hand {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(card)),
			})
		}
		return &model.Prompt{
			Type:       model.PromptChooseCards,
			PlayerID:   playerID,
			ChoiceType: choiceType,
			Message:    fmt.Sprintf("【血之诅咒】请弃置%d张手牌：", discardCount),
			Options:    options,
			Min:        discardCount,
			Max:        discardCount,
		}
	}

	return nil
}

func (e *GameEngine) handleBloodPriestessChoiceInput(playerID string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "bp_shared_life_target":
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
		target := e.State.Players[targetIDs[selectionIndex]]
		if target == nil {
			return true, fmt.Errorf("同生共死目标不存在")
		}
		if user.Character == nil {
			return true, fmt.Errorf("角色信息缺失")
		}
		linkCard, ok := user.ConsumeExclusiveCard(user.Character.ID, "同生共死")
		if !ok {
			return true, fmt.Errorf("未找到【同生共死】专属技能卡")
		}
		drawCtx := e.newDrawContext(user, 2, "bp_shared_life_draw")
		if drawCtx == nil {
			user.RestoreExclusiveCard(linkCard)
			return true, fmt.Errorf("同生共死摸牌上下文创建失败")
		}
		drawCtx.Selections["draw_resume_phase"] = normalizeChoiceResumePoint(model.TurnStageActionExecution)
		drawCtx.Selections["draw_followup"] = model.DeferredFollowup{
			Type:      "blood_priestess_shared_life_place",
			UserID:    user.ID,
			TargetIDs: []string{target.ID},
			Data: map[string]interface{}{
				"card": linkCard,
			},
		}
		e.startDraw(drawCtx)
		e.Log(fmt.Sprintf("%s 发动 [同生共死]：先摸2张牌，待爆牌结算后放置于 %s 面前", user.Name, target.Name))

		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if !e.routePendingDamageWithDefaultReturn(model.TurnStageActionExecution) {
				e.restorePhaseAfterInterruptedDraw(drawCtx)
			}
		}
		return true, nil

	case "bp_blood_sorrow_mode":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		if selectionIndex < 0 || selectionIndex > 1 {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		if !runtimeutil.ToBoolContextValue(ctxData["damage_queued"]) {
			e.AddPendingDamage(model.PendingDamage{
				SourceID:   user.ID,
				TargetID:   user.ID,
				Damage:     2,
				DamageType: "magic",
			})
			ctxData["damage_queued"] = true
		}
		if selectionIndex == 0 {
			e.enqueueDeferredFollowup(model.DeferredFollowup{
				Type:   "blood_priestess_blood_sorrow_apply",
				UserID: user.ID,
				Data: map[string]interface{}{
					"mode": "remove",
				},
			})
			e.Log(fmt.Sprintf("%s 发动 [血之哀伤]：先对自己造成2点法术伤害，伤害结算后移除【同生共死】", user.Name))
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				e.setReturnPoint(model.TurnStageActionStart)
				e.enterDamageResolution(nil)
			}
			return true, nil
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
		if len(targetIDs) == 0 {
			return true, fmt.Errorf("无可转移目标")
		}
		ctxData["choice_type"] = "bp_blood_sorrow_target"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil

	case "bp_blood_sorrow_target":
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
		if !runtimeutil.ToBoolContextValue(ctxData["damage_queued"]) {
			e.AddPendingDamage(model.PendingDamage{
				SourceID:   user.ID,
				TargetID:   user.ID,
				Damage:     2,
				DamageType: "magic",
			})
			ctxData["damage_queued"] = true
		}
		target := e.State.Players[targetIDs[selectionIndex]]
		if target == nil {
			return true, fmt.Errorf("转移目标不存在")
		}
		e.enqueueDeferredFollowup(model.DeferredFollowup{
			Type:   "blood_priestess_blood_sorrow_apply",
			UserID: user.ID,
			Data: map[string]interface{}{
				"mode":      "transfer",
				"target_id": target.ID,
			},
		})
		e.Log(fmt.Sprintf("%s 发动 [血之哀伤]：先对自己造成2点法术伤害，伤害结算后将【同生共死】转移至 %s", user.Name, target.Name))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.setReturnPoint(model.TurnStageActionStart)
			e.enterDamageResolution(nil)
		}
		return true, nil

	case "bp_blood_wail_x":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		if !hasBloodPriestessBleedingForm(user) {
			return true, fmt.Errorf("仅流血形态下可发动血之悲鸣")
		}
		targetID, _ := ctxData["target_id"].(string)
		target := e.State.Players[targetID]
		if target == nil {
			return true, fmt.Errorf("目标角色不存在")
		}
		if selectionIndex < 0 || selectionIndex > 2 {
			return true, fmt.Errorf("无效的X值")
		}
		damage := selectionIndex + 1
		e.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   target.ID,
			Damage:     damage,
			DamageType: "magic",
		})
		e.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   user.ID,
			Damage:     damage,
			DamageType: "magic",
		})
		e.Log(fmt.Sprintf("%s 发动 [血之悲鸣]：对 %s 和自己各造成%d点法术伤害", user.Name, target.Name, damage))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.setReturnPoint(model.TurnStageExtraAction)
			e.enterDamageResolution(nil)
		}
		return true, nil

	case "bp_curse_discard":
		if selectionIndex < 0 {
			return true, fmt.Errorf("请在手牌区选择需要弃置的手牌后确认")
		}
		return true, e.handleBloodCurseDiscardSelections(playerID, []int{selectionIndex})
	}

	return false, nil
}

func (e *GameEngine) resolveBloodPriestessSharedLifePlaceFollowup(f model.DeferredFollowup) error {
	user := e.State.Players[f.UserID]
	if user == nil {
		return fmt.Errorf("执行者不存在: %s", f.UserID)
	}
	if !e.isBloodPriestess(user) {
		return fmt.Errorf("仅血之巫女可执行同生共死后续")
	}
	if len(f.TargetIDs) != 1 {
		return fmt.Errorf("同生共死后续目标数量错误: %d", len(f.TargetIDs))
	}
	target := e.State.Players[f.TargetIDs[0]]
	if target == nil {
		return fmt.Errorf("同生共死目标不存在: %s", f.TargetIDs[0])
	}

	var card model.Card
	if f.Data != nil {
		if v, ok := f.Data["card"]; ok {
			switch c := v.(type) {
			case model.Card:
				card = c
			case *model.Card:
				if c != nil {
					card = *c
				}
			}
		}
	}
	if card.ID == "" || card.Name == "" {
		return fmt.Errorf("同生共死后续缺少原始专属卡")
	}

	if err := e.placeBloodPriestessSharedLife(user, target, card); err != nil {
		user.RestoreExclusiveCard(card)
		return err
	}
	e.Log(fmt.Sprintf("%s 的 [同生共死] 生效：放置于 %s 面前", user.Name, target.Name))

	e.checkHandLimit(user, nil)
	if target.ID != user.ID {
		e.checkHandLimit(target, nil)
	}
	return nil
}

func (e *GameEngine) resolveBloodPriestessBloodSorrowFollowup(f model.DeferredFollowup) error {
	user := e.State.Players[f.UserID]
	if user == nil {
		return fmt.Errorf("执行者不存在: %s", f.UserID)
	}
	mode := ""
	if f.Data != nil {
		if raw, ok := f.Data["mode"].(string); ok {
			mode = raw
		}
	}
	if mode == "" {
		return fmt.Errorf("血之哀伤后续模式缺失")
	}
	checked := map[string]bool{}
	checkCap := func(player *model.Player) {
		if player == nil || checked[player.ID] {
			return
		}
		checked[player.ID] = true
		e.checkHandLimit(player, nil)
	}

	switch mode {
	case "remove":
		if !e.removeBloodPriestessSharedLife(user, true) {
			return fmt.Errorf("当前没有可移除的同生共死")
		}
		e.Log(fmt.Sprintf("%s 的 [血之哀伤] 后续结算：移除【同生共死】", user.Name))
		checkCap(user)
		return nil
	case "transfer":
		targetID := ""
		if f.Data != nil {
			if raw, ok := f.Data["target_id"].(string); ok {
				targetID = raw
			}
		}
		target := e.State.Players[targetID]
		if target == nil {
			return fmt.Errorf("转移目标不存在: %s", targetID)
		}
		holder, card, ok := e.detachBloodPriestessSharedLife(user)
		if !ok {
			return fmt.Errorf("当前没有可转移的同生共死")
		}
		if err := e.attachExclusiveEffectCard(user, target, model.EffectBloodSharedLife, card); err != nil {
			if holder != nil {
				_ = e.attachExclusiveEffectCard(user, holder, model.EffectBloodSharedLife, card)
			}
			return err
		}
		e.Log(fmt.Sprintf("%s 的 [血之哀伤] 后续结算：将【同生共死】转移至 %s", user.Name, target.Name))
		checkCap(user)
		checkCap(holder)
		checkCap(target)
		return nil
	default:
		return fmt.Errorf("未知的血之哀伤后续模式: %s", mode)
	}
}

func (e *GameEngine) resolveBloodPriestessCurseDiscardFollowup(f model.DeferredFollowup) error {
	user := e.State.Players[f.UserID]
	if user == nil {
		return fmt.Errorf("执行者不存在: %s", f.UserID)
	}
	discardNeed := 3
	if len(user.Hand) < discardNeed {
		discardNeed = len(user.Hand)
	}
	if discardNeed <= 0 {
		e.Log(fmt.Sprintf("%s 的 [血之诅咒] 后续：无需弃牌", user.Name))
		return nil
	}
	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: user.ID,
		Context: map[string]interface{}{
			"choice_type":   "bp_curse_discard",
			"user_id":       user.ID,
			"discard_count": discardNeed,
		},
	})
	e.Log(fmt.Sprintf("%s 的 [血之诅咒] 后续：伤害结算完成，请弃置%d张牌", user.Name, discardNeed))
	return nil
}

func (e *GameEngine) handleBloodCurseDiscardSelections(playerID string, selections []int) error {
	if e.State.PendingInterrupt == nil || e.State.PendingInterrupt.Type != model.InterruptChoice {
		return fmt.Errorf("当前不存在可处理的血之诅咒弃牌")
	}
	ctxData, ok := e.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		return fmt.Errorf("血之诅咒上下文错误")
	}
	choiceType, _ := ctxData["choice_type"].(string)
	if choiceType != "bp_curse_discard" {
		return fmt.Errorf("当前中断不是血之诅咒弃牌")
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
	if discardNeed > len(user.Hand) {
		discardNeed = len(user.Hand)
	}
	if discardNeed == 0 {
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if !e.routePendingDamageWithReturn(model.TurnStageExtraAction) {
				e.enterExtraActionStage()
			}
		}
		return nil
	}

	if len(selections) != discardNeed {
		return fmt.Errorf("需要选择 %d 张手牌弃置", discardNeed)
	}

	chosen := make([]int, 0, discardNeed)
	seen := make(map[int]struct{}, discardNeed)
	for _, idx := range selections {
		if idx < 0 || idx >= len(user.Hand) {
			return fmt.Errorf("无效的弃牌索引: %d", idx)
		}
		if _, ok := seen[idx]; ok {
			return fmt.Errorf("不能重复选择同一张牌")
		}
		seen[idx] = struct{}{}
		chosen = append(chosen, idx)
	}

	sort.Sort(sort.Reverse(sort.IntSlice(chosen)))
	discarded := make([]model.Card, 0, len(chosen))
	for _, idx := range chosen {
		discarded = append(discarded, user.Hand[idx])
		user.Hand = append(user.Hand[:idx], user.Hand[idx+1:]...)
	}

	if len(discarded) > 0 {
		e.NotifyCardRevealed(user.ID, discarded, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, discarded...)
	}
	e.Log(fmt.Sprintf("%s 的 [血之诅咒] 后续：弃置%d张牌", user.Name, len(discarded)))

	e.PopInterrupt()
	if e.State.PendingInterrupt == nil {
		if !e.routePendingDamageWithDefaultReturn(model.TurnStageExtraAction) {
			e.enterExtraActionStage()
		}
	}
	return nil
}
