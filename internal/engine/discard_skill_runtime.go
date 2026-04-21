// gameflow: 弃牌技能与弃牌子流程（如展示/封印联动）。

package engine

import (
	"fmt"

	"starcup-engine/internal/engine/core/runtimeutil"
	beastsamurai "starcup-engine/internal/engine/player/beast_samurai"
	"starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
)

// ConfirmDiscard 确认执行弃牌。
func (e *GameEngine) ConfirmDiscard(playerID string, indices []int) error {
	data, err := e.pendingDiscardContext()
	if err != nil {
		return err
	}
	skillID, hasSkillID := data["skill_id"].(string)

	choiceType, _ := data["choice_type"].(string)
	if isBeastSamuraiDiscardChoiceType(choiceType) {
		return e.handleBeastSamuraiDiscardSelections(playerID, indices, data)
	}

	if e.State.PendingInterrupt == nil {
		return fmt.Errorf("当前没有待处理的弃牌操作")
	}
	if e.State.PendingInterrupt.PlayerID != "" && e.State.PendingInterrupt.PlayerID != playerID {
		return fmt.Errorf("当前不是你的弃牌回合")
	}

	if hasSkillID && skillID != "" {
		return e.handleSkillDiscardSelection(playerID, indices, data)
	}

	return e.handleDiscardSelection(playerID, indices, data)
}

func (e *GameEngine) confirmDiscardChoiceSelections(playerID string, indices []int, data map[string]interface{}) error {
	if data == nil {
		var err error
		data, err = e.pendingDiscardContext()
		if err != nil {
			return err
		}
	}
	skillID, hasSkillID := data["skill_id"].(string)
	choiceType, _ := data["choice_type"].(string)
	if isBeastSamuraiDiscardChoiceType(choiceType) {
		return e.handleBeastSamuraiDiscardSelections(playerID, indices, data)
	}
	if hasSkillID && skillID != "" {
		return e.handleSkillDiscardSelection(playerID, indices, data)
	}
	return e.handleDiscardSelection(playerID, indices, data)
}

func (e *GameEngine) handleSkillDiscardSelection(playerID string, indices []int, data map[string]interface{}) error {
	skillID, _ := data["skill_id"].(string)
	if skillID == "" {
		return fmt.Errorf("技能弃牌上下文缺少 skill_id")
	}
	if _, hasCtx := data["user_ctx"]; !hasCtx {
		return e.handleDeferredSkillDiscardSelection(playerID, skillID, indices, data)
	}
	return e.handleContextSkillDiscardSelection(skillID, indices, data)
}

func (e *GameEngine) handleDeferredSkillDiscardSelection(playerID, skillID string, indices []int, data map[string]interface{}) error {
	targetIDs := runtimeutil.ParseStringSliceContextValue(data["target_ids"])
	resumePoint := data["resume_phase"]

	e.PopInterrupt()
	if e.State.PendingInterrupt != nil {
		return fmt.Errorf("当前仍有其他待处理的中断")
	}
	// 规则：为发动技能而产生的弃牌中断，处理完必须回到技能声明的恢复点后再继续施放技能。
	e.applyChoiceResumePoint(mustChoiceResumePoint(resumePoint, "resume_phase"))
	return e.UseSkill(playerID, skillID, targetIDs, indices)
}

func (e *GameEngine) handleContextSkillDiscardSelection(skillID string, indices []int, data map[string]interface{}) error {
	minSelect, _ := data["min"].(int)
	maxSelect, _ := data["max"].(int)
	if len(indices) < minSelect {
		return fmt.Errorf("至少需要选择 %d 张牌，你选择了 %d 张", minSelect, len(indices))
	}
	if len(indices) > maxSelect {
		return fmt.Errorf("最多只能选择 %d 张牌，你选择了 %d 张", maxSelect, len(indices))
	}

	userCtx, hasCtx := data["user_ctx"]
	if !hasCtx {
		return fmt.Errorf("技能上下文丢失")
	}
	ctx, ok := userCtx.(*model.Context)
	if !ok {
		return fmt.Errorf("技能上下文格式错误")
	}
	if ctx.Selections == nil {
		ctx.Selections = make(map[string]any)
	}
	ctx.Selections["discard_indices"] = indices

	handler := skills.GetHandler(skillID)
	if handler == nil {
		return fmt.Errorf("技能处理器不存在")
	}

	beforePoses := e.snapshotPlayerPoses()
	if err := handler.Execute(ctx); err != nil {
		return fmt.Errorf("技能执行失败: %v", err)
	}
	e.dispatchOrientationChanges(beforePoses)

	if discardedCards, ok := ctx.Selections["discardedCards"].([]model.Card); ok {
		e.State.DiscardPile = append(e.State.DiscardPile, discardedCards...)
	}
	if ctx.BeforeDrawPhase() {
		e.resumePendingDraw(ctx)
	}

	if nextSkillIDs, ok := data["remaining_skills"].([]string); ok && len(nextSkillIDs) > 0 {
		e.State.PendingInterrupt.Type = model.InterruptResponseSkill
		e.State.PendingInterrupt.SkillIDs = nextSkillIDs
		e.State.PendingInterrupt.Context = ctx
		e.Log("[System] 弃牌技能执行完毕，你还可以选择发动其他技能")
		e.enterResponseWindow()
		return nil
	}

	e.PopInterrupt()
	if e.State.PendingInterrupt == nil {
		e.resumePhaseAfterSkillDiscardContext(ctx)
	}
	return nil
}

func (e *GameEngine) resumePhaseAfterSkillDiscardContext(ctx *model.Context) bool {
	if ctx == nil || e.State.PendingInterrupt != nil {
		return false
	}
	if ctx.BeforeDrawPhase() {
		return e.restorePhaseAfterInterruptedDraw(ctx)
	}
	if ctx.ResumeActionEndPhase() {
		// ActionEnd 响应中的弃牌交互完成后，避免 LastActionType 残留触发同一轮 ActionEnd 重入。
		if ctx.User != nil {
			ctx.User.TurnState.LastActionType = ""
			ctx.User.TurnState.LastActionCard = nil
		}
		if point, ok := choiceResumePointValue(ctx.Selections["response_resume_phase"]); ok {
			if e.routePendingDamageWithReturn(point) {
				return true
			}
			e.applyChoiceResumePoint(point)
			return true
		}
		e.enterExtraActionStage()
		return true
	}
	if ctx.ResumeAttackMissPhase() && e.resumePendingAttackMiss(ctx) {
		return true
	}
	if ctx.TurnStartOrStartupWindow() {
		// 启动技能（回合开始触发）中的弃牌后续：应继续当前回合流程。
		e.clearSubflow()
		e.clearCombatStage()
		if !e.routePendingDamageWithReturn(model.TurnStageActionStart) {
			e.setTurnStage(model.TurnStageActionStart)
		}
		return true
	}

	if len(e.State.ActionStack) > 0 {
		e.enterResponseWindow()
	} else if len(e.State.ActionQueue) > 0 {
		e.enterActionExecutionStage()
	} else {
		e.enterTurnEndStage()
	}
	return true
}

// ---- 兽武者：残心/兽魂资源与居合形态 ----

func (e *GameEngine) consumeBeastSamuraiBeastSoul(player *model.Player, amount int) int {
	if player == nil || amount <= 0 {
		return 0
	}
	current := beastsamurai.BeastSoul(player)
	if amount > current {
		amount = current
	}
	if amount <= 0 {
		return 0
	}
	beastsamurai.AddBeastSoul(player, -amount, true)
	beastsamurai.AddZanshin(player, amount)
	return amount
}

func (e *GameEngine) enterBeastSamuraiIaijutsuForm(player *model.Player) bool {
	if player == nil {
		return false
	}
	changed := effectivePlayerOrientation(player) != model.OrientationTapped || effectivePlayerForm(player) != model.FormBeastSamuraiIaijutsu
	player.Orientation = model.OrientationTapped
	player.Form = model.FormBeastSamuraiIaijutsu
	return changed
}

func (e *GameEngine) leaveBeastSamuraiIaijutsuForm(player *model.Player) bool {
	if player == nil {
		return false
	}
	changed := effectivePlayerOrientation(player) != model.OrientationNormal || effectivePlayerForm(player) != ""
	player.Orientation = model.OrientationNormal
	player.Form = ""
	return changed
}

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
			after := beastsamurai.AddBeastSoul(user, 1, false)
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
			after := beastsamurai.AddBeastSoul(user, 1, false)
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
