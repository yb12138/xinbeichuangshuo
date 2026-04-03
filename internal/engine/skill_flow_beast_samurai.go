package engine

import (
	"fmt"
	"starcup-engine/internal/engine/runtimeutil"
	"strings"

	"starcup-engine/internal/model"
)

func beastSamuraiDiscardedMagicCount(cards []model.Card) int {
	count := 0
	for _, card := range cards {
		if card.Type == model.CardTypeMagic {
			count++
		}
	}
	return count
}

func (e *GameEngine) beastSamuraiResumePoint(ctxData map[string]interface{}, fallback interface{}) interface{} {
	if resumePoint := normalizeChoiceResumePoint(ctxData["resume_phase"]); resumePoint != "" {
		return resumePoint
	}
	return fallback
}

func (e *GameEngine) beastSamuraiReplacePendingInterruptWithDiscard(playerID string, ctxData map[string]interface{}) {
	if e.State.PendingInterrupt == nil {
		return
	}
	e.State.PendingInterrupt.Type = model.InterruptDiscard
	e.State.PendingInterrupt.PlayerID = playerID
	e.State.PendingInterrupt.Context = ctxData
	e.notifyInterruptPrompt()
}

func (e *GameEngine) beastSamuraiFinishResume(resumePoint interface{}) {
	e.PopInterrupt()
	if e.State.PendingInterrupt != nil {
		return
	}
	if normalizeChoiceResumePoint(resumePoint) != "" {
		e.applyChoiceResumePoint(resumePoint)
		return
	}
	if len(e.State.PendingDamageQueue) > 0 {
		e.enterDamageResolution(nil)
		return
	}
	if len(e.State.ActionQueue) > 0 {
		e.enterActionExecutionStage()
		return
	}
	e.applyChoiceResumePoint(model.TurnStageExtraAction)
}

func (e *GameEngine) beastSamuraiFindPendingAttackDamage(rawCtx *model.Context) *model.PendingDamage {
	if rawCtx == nil || rawCtx.TriggerCtx == nil {
		return nil
	}
	for i := range e.State.PendingDamageQueue {
		pd := &e.State.PendingDamageQueue[i]
		if !strings.EqualFold(pd.DamageType, "Attack") {
			continue
		}
		if pd.SourceID != rawCtx.TriggerCtx.SourceID || pd.TargetID != rawCtx.TriggerCtx.TargetID {
			continue
		}
		return pd
	}
	return nil
}

func (e *GameEngine) beastSamuraiFinishReversal(rawCtx *model.Context, target *model.Player, need, actualDiscarded int, resumePoint interface{}) {
	if target != nil && actualDiscarded < need {
		userName := "兽灵武士"
		if rawCtx != nil && rawCtx.User != nil {
			userName = rawCtx.User.Name
		}
		loss := e.applyCampMoraleLoss(target.Camp, 1)
		if loss > 0 {
			e.Log(fmt.Sprintf("%s 的 [逆反居合斩] 生效：%s 实际弃牌%d/%d，%s方士气-%d", userName, target.Name, actualDiscarded, need, target.Camp, loss))
		} else {
			e.Log(fmt.Sprintf("%s 的 [逆反居合斩] 生效：%s 实际弃牌%d/%d，但%s方士气已触及下限", userName, target.Name, actualDiscarded, need, target.Camp))
		}
	}
	if rawCtx != nil {
		e.markPendingAttackDamageHitProcessed(rawCtx)
	}
	e.beastSamuraiFinishResume(resumePoint)
}

func (e *GameEngine) buildBeastSamuraiChoicePrompt(choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "bs_beast_return_x":
		maxX := runtimeutil.ToIntContextValue(data["max_x"])
		options := make([]model.PromptOption, 0, maxX+1)
		for x := 0; x <= maxX; x++ {
			label := fmt.Sprintf("X=%d", x)
			if x == 0 {
				label = "X=0（不移除兽魂）"
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", x),
				Label: label,
			})
		}
		return &model.Prompt{
			Type:       model.PromptConfirm,
			PlayerID:   playerID,
			ChoiceType: choiceType,
			Message:    fmt.Sprintf("【兽返】请选择要移除的兽魂数量（0-%d）：", maxX),
			Options:    options,
			Min:        1,
			Max:        1,
		}
	case "bs_reversal_x":
		maxX := runtimeutil.ToIntContextValue(data["max_x"])
		options := make([]model.PromptOption, 0, maxX+1)
		for x := 0; x <= maxX; x++ {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", x),
				Label: fmt.Sprintf("X=%d（目标将弃置%d张手牌）", x, x+2),
			})
		}
		return &model.Prompt{
			Type:       model.PromptConfirm,
			PlayerID:   playerID,
			ChoiceType: choiceType,
			Message:    fmt.Sprintf("【逆反居合斩】请选择要移除的兽魂数量（0-%d）：", maxX),
			Options:    options,
			Min:        1,
			Max:        1,
		}
	case "bs_iaijutsu_style_mode":
		modes := runtimeutil.ParseChoiceIntSlice(data["modes"])
		if len(modes) == 0 {
			modes = []int{0, 1}
		}
		options := make([]model.PromptOption, 0, len(modes))
		for _, mode := range modes {
			switch mode {
			case 0:
				options = append(options, model.PromptOption{ID: "0", Label: "摸1张牌"})
			case 1:
				options = append(options, model.PromptOption{ID: "1", Label: "弃1张牌"})
			}
		}
		if len(options) == 0 {
			return nil
		}
		return &model.Prompt{
			Type:       model.PromptConfirm,
			PlayerID:   playerID,
			ChoiceType: choiceType,
			Message:    "【御魂流居合式】请选择“摸1张牌”或“弃1张牌”：",
			Options:    options,
			Min:        1,
			Max:        1,
		}
	default:
		return nil
	}
}

func (e *GameEngine) handleBeastSamuraiChoiceInput(selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "bs_beast_return_x":
		userID, _ := ctxData["user_id"].(string)
		sourceID, _ := ctxData["source_id"].(string)
		user := e.State.Players[userID]
		source := e.State.Players[sourceID]
		if user == nil {
			return true, fmt.Errorf("兽返执行者不存在")
		}
		maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
		if current := e.beastSamuraiBeastSoul(user); maxX > current {
			maxX = current
		}
		if selectionIndex < 0 || selectionIndex > maxX {
			return true, fmt.Errorf("无效的X值: %d", selectionIndex)
		}
		x := selectionIndex
		consumed := e.consumeBeastSamuraiBeastSoul(user, x)
		resumePoint := e.beastSamuraiResumePoint(ctxData, model.CombatStageCalcDamage)
		e.Log(fmt.Sprintf("%s 的 [兽返] 结算：移除%d点兽魂，残心同步+%d", user.Name, consumed, consumed))

		selfDiscardCount := x
		if selfDiscardCount > len(user.Hand) {
			selfDiscardCount = len(user.Hand)
		}
		if selfDiscardCount > 0 {
			e.beastSamuraiReplacePendingInterruptWithDiscard(user.ID, map[string]interface{}{
				"choice_type":   "bs_beast_return_self_discard",
				"user_id":       user.ID,
				"source_id":     sourceID,
				"x_value":       x,
				"discard_count": selfDiscardCount,
				"prompt":        fmt.Sprintf("【兽返】请选择弃置%d张手牌：", selfDiscardCount),
				"resume_phase":  normalizeChoiceResumePoint(resumePoint),
			})
			return true, nil
		}
		if source != nil && len(source.Hand) > 0 {
			e.beastSamuraiReplacePendingInterruptWithDiscard(source.ID, map[string]interface{}{
				"choice_type":   "bs_beast_return_source_discard",
				"user_id":       user.ID,
				"source_id":     source.ID,
				"discard_count": 1,
				"prompt":        fmt.Sprintf("【兽返】请选择弃置1张手牌："),
				"resume_phase":  normalizeChoiceResumePoint(resumePoint),
			})
			return true, nil
		}
		e.beastSamuraiFinishResume(resumePoint)
		return true, nil

	case "bs_reversal_x":
		userID, _ := ctxData["user_id"].(string)
		targetID, _ := ctxData["target_id"].(string)
		user := e.State.Players[userID]
		target := e.State.Players[targetID]
		if user == nil || target == nil {
			return true, fmt.Errorf("逆反居合斩上下文目标不存在")
		}
		maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
		if current := e.beastSamuraiBeastSoul(user); maxX > current {
			maxX = current
		}
		if selectionIndex < 0 || selectionIndex > maxX {
			return true, fmt.Errorf("无效的X值: %d", selectionIndex)
		}
		x := selectionIndex
		consumed := e.consumeBeastSamuraiBeastSoul(user, x)
		if user.Tokens == nil {
			user.Tokens = map[string]int{}
		}
		user.Tokens["bs_reversal_pending_x"] = x
		rawCtx, _ := ctxData["user_ctx"].(*model.Context)
		if pd := e.beastSamuraiFindPendingAttackDamage(rawCtx); pd != nil {
			pd.Damage = 0
		}
		need := x + 2
		resumePoint := e.beastSamuraiResumePoint(ctxData, model.CombatStageCalcDamage)
		e.Log(fmt.Sprintf("%s 的 [逆反居合斩] 结算：移除%d点兽魂，残心同步+%d，改写本次攻击为弃牌效果", user.Name, consumed, consumed))

		discardCount := need
		if discardCount > len(target.Hand) {
			discardCount = len(target.Hand)
		}
		if discardCount > 0 {
			e.beastSamuraiReplacePendingInterruptWithDiscard(target.ID, map[string]interface{}{
				"choice_type":   "bs_reversal_target_discard",
				"user_id":       user.ID,
				"target_id":     target.ID,
				"x_value":       x,
				"need_count":    need,
				"discard_count": discardCount,
				"prompt":        fmt.Sprintf("【逆反居合斩】请选择弃置%d张手牌：", discardCount),
				"resume_phase":  normalizeChoiceResumePoint(resumePoint),
				"user_ctx":      rawCtx,
			})
			return true, nil
		}
		e.beastSamuraiFinishReversal(rawCtx, target, need, 0, resumePoint)
		return true, nil

	case "bs_iaijutsu_style_mode":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("御魂流居合式执行者不存在")
		}
		modes := runtimeutil.ParseChoiceIntSlice(ctxData["modes"])
		if len(modes) == 0 {
			modes = []int{0, 1}
		}
		if selectionIndex < 0 || selectionIndex >= len(modes) {
			return true, fmt.Errorf("无效的模式选择: %d", selectionIndex)
		}
		mode := modes[selectionIndex]
		resumePoint := e.beastSamuraiResumePoint(ctxData, model.TurnStageActionStart)

		beforePoses := e.snapshotPlayerPoses()
		soulAfter := e.addBeastSamuraiBeastSoul(user, 1, true)
		if e.beastSamuraiInIaijutsuForm(user) {
			zanshinAfter := e.addBeastSamuraiZanshin(user, 1)
			e.Log(fmt.Sprintf("%s 的 [御魂流居合式] 生效：兽魂+1（当前%d），因已处于御魂流居合形态，残心+1（当前%d）", user.Name, soulAfter, zanshinAfter))
		} else {
			e.enterBeastSamuraiIaijutsuForm(user)
			e.Log(fmt.Sprintf("%s 的 [御魂流居合式] 生效：兽魂+1（当前%d），进入御魂流居合形态", user.Name, soulAfter))
		}
		e.dispatchOrientationChanges(beforePoses)

		switch mode {
		case 0:
			e.DrawCards(user.ID, 1)
			e.beastSamuraiFinishResume(resumePoint)
			return true, nil
		case 1:
			discardCount := 1
			if len(user.Hand) < discardCount {
				discardCount = len(user.Hand)
			}
			if discardCount > 0 {
				e.beastSamuraiReplacePendingInterruptWithDiscard(user.ID, map[string]interface{}{
					"choice_type":   "bs_iaijutsu_style_discard",
					"user_id":       user.ID,
					"discard_count": discardCount,
					"prompt":        "【御魂流居合式】请选择弃置1张手牌：",
					"resume_phase":  normalizeChoiceResumePoint(resumePoint),
				})
				return true, nil
			}
			e.beastSamuraiFinishResume(resumePoint)
			return true, nil
		default:
			return true, fmt.Errorf("未知的御魂流居合式模式: %d", mode)
		}
	default:
		return false, nil
	}
}

func (e *GameEngine) handleBeastSamuraiDiscardInput(playerID string, selections []int) (bool, error) {
	if e.State.PendingInterrupt == nil || e.State.PendingInterrupt.Type != model.InterruptDiscard {
		return false, nil
	}
	ctxData, ok := e.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		return false, nil
	}
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "bs_alert_source_discard":
		userID, _ := ctxData["user_id"].(string)
		actorID, _ := ctxData["actor_id"].(string)
		user := e.State.Players[userID]
		actor := e.State.Players[actorID]
		if user == nil || actor == nil {
			return true, fmt.Errorf("兽魂警戒弃牌上下文不存在")
		}
		removed, err := removeCardsByIndicesFromHand(actor, append([]int{}, selections...))
		if err != nil {
			return true, err
		}
		if len(removed) > 0 {
			e.NotifyCardRevealed(actor.ID, removed, "discard")
			e.State.DiscardPile = append(e.State.DiscardPile, removed...)
		}
		if beastSamuraiDiscardedMagicCount(removed) > 0 {
			after := e.addBeastSamuraiBeastSoul(user, 1, false)
			e.Log(fmt.Sprintf("%s 的 [兽魂警戒] 生效：%s 展示弃牌中含法术牌，兽魂+1（当前%d）", user.Name, actor.Name, after))
		}
		e.beastSamuraiFinishResume(e.beastSamuraiResumePoint(ctxData, model.TurnStageActionExecution))
		return true, nil

	case "bs_beast_return_self_discard":
		userID, _ := ctxData["user_id"].(string)
		sourceID, _ := ctxData["source_id"].(string)
		user := e.State.Players[userID]
		source := e.State.Players[sourceID]
		if user == nil {
			return true, fmt.Errorf("兽返弃牌执行者不存在")
		}
		removed, err := removeCardsByIndicesFromHand(user, append([]int{}, selections...))
		if err != nil {
			return true, err
		}
		if len(removed) > 0 {
			e.NotifyCardHidden(user.ID, removed, "discard")
			e.State.DiscardPile = append(e.State.DiscardPile, removed...)
		}
		resumePoint := e.beastSamuraiResumePoint(ctxData, model.CombatStageCalcDamage)
		if source != nil && len(source.Hand) > 0 {
			e.beastSamuraiReplacePendingInterruptWithDiscard(source.ID, map[string]interface{}{
				"choice_type":   "bs_beast_return_source_discard",
				"user_id":       user.ID,
				"source_id":     source.ID,
				"discard_count": 1,
				"prompt":        "【兽返】请选择弃置1张手牌：",
				"resume_phase":  normalizeChoiceResumePoint(resumePoint),
			})
			return true, nil
		}
		e.beastSamuraiFinishResume(resumePoint)
		return true, nil

	case "bs_beast_return_source_discard":
		userID, _ := ctxData["user_id"].(string)
		sourceID, _ := ctxData["source_id"].(string)
		user := e.State.Players[userID]
		source := e.State.Players[sourceID]
		if user == nil || source == nil {
			return true, fmt.Errorf("兽返来源弃牌上下文不存在")
		}
		removed, err := removeCardsByIndicesFromHand(source, append([]int{}, selections...))
		if err != nil {
			return true, err
		}
		if len(removed) > 0 {
			e.NotifyCardHidden(source.ID, removed, "discard")
			e.State.DiscardPile = append(e.State.DiscardPile, removed...)
		}
		if beastSamuraiDiscardedMagicCount(removed) > 0 {
			after := e.addBeastSamuraiBeastSoul(user, 1, false)
			e.Log(fmt.Sprintf("%s 的 [兽返] 生效：%s 弃牌中含法术牌，兽魂+1（当前%d）", user.Name, source.Name, after))
		}
		e.beastSamuraiFinishResume(e.beastSamuraiResumePoint(ctxData, model.CombatStageCalcDamage))
		return true, nil

	case "bs_iaijutsu_style_discard":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("御魂流居合式弃牌执行者不存在")
		}
		removed, err := removeCardsByIndicesFromHand(user, append([]int{}, selections...))
		if err != nil {
			return true, err
		}
		if len(removed) > 0 {
			e.NotifyCardHidden(user.ID, removed, "discard")
			e.State.DiscardPile = append(e.State.DiscardPile, removed...)
		}
		e.beastSamuraiFinishResume(e.beastSamuraiResumePoint(ctxData, model.TurnStageActionStart))
		return true, nil

	case "bs_reversal_target_discard":
		targetID, _ := ctxData["target_id"].(string)
		target := e.State.Players[targetID]
		if target == nil {
			return true, fmt.Errorf("逆反居合斩目标不存在")
		}
		rawCtx, _ := ctxData["user_ctx"].(*model.Context)
		need := runtimeutil.ToIntContextValue(ctxData["need_count"])
		removed, err := removeCardsByIndicesFromHand(target, append([]int{}, selections...))
		if err != nil {
			return true, err
		}
		if len(removed) > 0 {
			e.NotifyCardHidden(target.ID, removed, "discard")
			e.State.DiscardPile = append(e.State.DiscardPile, removed...)
		}
		e.beastSamuraiFinishReversal(rawCtx, target, need, len(removed), e.beastSamuraiResumePoint(ctxData, model.CombatStageCalcDamage))
		return true, nil

	default:
		return false, nil
	}
}
