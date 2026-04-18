// gameflow: 兽魂武士：兽魂警觉等。

package engine

import (
	"fmt"
	"starcup-engine/internal/engine/core/runtimeutil"
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
	if resumePoint, ok := choiceResumePointValue(ctxData["resume_phase"]); ok {
		return resumePoint
	}
	return fallback
}

func (e *GameEngine) beastSamuraiReplacePendingInterruptWithDiscard(playerID string, ctxData map[string]interface{}) {
	if e.State.PendingInterrupt == nil {
		return
	}
	ctxData = normalizeDiscardChoiceContext(ctxData)
	e.State.PendingInterrupt.Type = model.InterruptChoice
	e.State.PendingInterrupt.PlayerID = playerID
	e.State.PendingInterrupt.Context = ctxData
	e.syncGamePhaseWithInterrupt(e.State.PendingInterrupt)
	e.notifyInterruptPrompt()
}

func (e *GameEngine) beastSamuraiFinishResume(resumePoint interface{}) {
	e.PopInterrupt()
	if e.State.PendingInterrupt != nil {
		return
	}
	if hasChoiceResumePoint(resumePoint) {
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
	if rawCtx == nil || rawCtx.EventCtx == nil {
		return nil
	}
	for i := range e.State.PendingDamageQueue {
		pd := &e.State.PendingDamageQueue[i]
		if !strings.EqualFold(string(pd.DamageType), string(model.AttackDamage)) {
			continue
		}
		if pd.SourceID != rawCtx.EventCtx.SourceID || pd.TargetID != rawCtx.EventCtx.TargetID {
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

func (e *GameEngine) handleBeastSamuraiDiscardSelections(playerID string, selections []int, providedCtx map[string]interface{}) error {
	if e == nil || e.State == nil || e.State.PendingInterrupt == nil {
		return fmt.Errorf("当前没有待处理的弃牌操作")
	}
	ctxData := providedCtx
	if ctxData == nil {
		var ok bool
		ctxData, ok = e.State.PendingInterrupt.Context.(map[string]interface{})
		if !ok || ctxData == nil {
			return fmt.Errorf("兽魂弃牌上下文错误")
		}
	}
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "bs_alert_source_discard":
		userID, _ := ctxData["user_id"].(string)
		actorID, _ := ctxData["actor_id"].(string)
		user := e.State.Players[userID]
		actor := e.State.Players[actorID]
		if user == nil || actor == nil {
			return fmt.Errorf("兽魂警戒弃牌上下文不存在")
		}
		removed, err := removeCardsByIndicesFromHand(actor, append([]int{}, selections...))
		if err != nil {
			return err
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
		return nil

	case "bs_beast_return_self_discard":
		userID, _ := ctxData["user_id"].(string)
		sourceID, _ := ctxData["source_id"].(string)
		user := e.State.Players[userID]
		source := e.State.Players[sourceID]
		if user == nil {
			return fmt.Errorf("兽返弃牌执行者不存在")
		}
		removed, err := removeCardsByIndicesFromHand(user, append([]int{}, selections...))
		if err != nil {
			return err
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
				"resume_phase":  resumePoint,
			})
			return nil
		}
		e.beastSamuraiFinishResume(resumePoint)
		return nil

	case "bs_beast_return_source_discard":
		userID, _ := ctxData["user_id"].(string)
		sourceID, _ := ctxData["source_id"].(string)
		user := e.State.Players[userID]
		source := e.State.Players[sourceID]
		if user == nil || source == nil {
			return fmt.Errorf("兽返来源弃牌上下文不存在")
		}
		removed, err := removeCardsByIndicesFromHand(source, append([]int{}, selections...))
		if err != nil {
			return err
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
		return nil

	case "bs_iaijutsu_style_discard":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return fmt.Errorf("御魂流居合式弃牌执行者不存在")
		}
		removed, err := removeCardsByIndicesFromHand(user, append([]int{}, selections...))
		if err != nil {
			return err
		}
		if len(removed) > 0 {
			e.NotifyCardHidden(user.ID, removed, "discard")
			e.State.DiscardPile = append(e.State.DiscardPile, removed...)
		}
		e.beastSamuraiFinishResume(e.beastSamuraiResumePoint(ctxData, model.TurnStageActionStart))
		return nil

	case "bs_reversal_target_discard":
		targetID, _ := ctxData["target_id"].(string)
		target := e.State.Players[targetID]
		if target == nil {
			return fmt.Errorf("逆反居合斩目标不存在")
		}
		rawCtx, _ := ctxData["user_ctx"].(*model.Context)
		need := runtimeutil.ToIntContextValue(ctxData["need_count"])
		removed, err := removeCardsByIndicesFromHand(target, append([]int{}, selections...))
		if err != nil {
			return err
		}
		if len(removed) > 0 {
			e.NotifyCardHidden(target.ID, removed, "discard")
			e.State.DiscardPile = append(e.State.DiscardPile, removed...)
		}
		e.beastSamuraiFinishReversal(rawCtx, target, need, len(removed), e.beastSamuraiResumePoint(ctxData, model.CombatStageCalcDamage))
		return nil

	default:
		return fmt.Errorf("非兽魂弃牌选择类型: %s", choiceType)
	}
}
