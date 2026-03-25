package engine

import (
	"fmt"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func buildAssassinDeferredFollowupHandlers() map[string]deferredFollowupHandler {
	return map[string]deferredFollowupHandler{
		"assassin_stealth_apply": {
			label:   "Assassin",
			resolve: (*GameEngine).resolveAssassinStealthApplyFollowup,
		},
	}
}

func (e *GameEngine) newDrawContext(player *model.Player, amount int, reason string) *model.Context {
	return e.newDrawContextWithOptions(player, amount, reason, model.DrawOptions{})
}

func (e *GameEngine) newDrawContextWithOptions(player *model.Player, amount int, reason string, opts model.DrawOptions) *model.Context {
	if player == nil || amount < 0 {
		return nil
	}

	resumePoint := e.currentChoiceResumePoint()
	if intr := e.State.PendingInterrupt; intr != nil && intr.Type == model.InterruptChoice {
		if data, ok := intr.Context.(map[string]interface{}); ok {
			if waitingPoint := normalizeChoiceResumePoint(data["waiting_phase"]); waitingPoint != "" {
				resumePoint = waitingPoint
			}
		}
	}

	drawCount := amount
	eventCtx := &model.EventContext{
		Type:      model.EventBeforeDraw,
		SourceID:  player.ID,
		TargetID:  player.ID,
		DrawCount: &drawCount,
		ActionType: func() model.ActionType {
			if player.TurnState.LastActionType == "" {
				return ""
			}
			return model.ActionType(player.TurnState.LastActionType)
		}(),
	}
	ctx := e.buildContext(player, player, model.TriggerBeforeDraw, eventCtx)
	if opts.PreventOverflow {
		ctx.Flags["preventOverflow"] = true
	}
	if resumePoint != "" && resumePoint != normalizeChoiceResumePoint(model.TurnStageTurnEnd) {
		ctx.Flags["StayInTurn"] = true
	}
	if reason == "" {
		reason = opts.Reason
	}
	if reason == "" {
		reason = "draw"
	}
	ctx.Selections["draw_reason"] = reason
	ctx.Selections["draw_resume_phase"] = resumePoint
	return ctx
}

func (e *GameEngine) startDraw(ctx *model.Context) bool {
	if ctx == nil || ctx.User == nil || ctx.TriggerCtx == nil || ctx.TriggerCtx.DrawCount == nil {
		return false
	}

	prevPending := e.State.PendingInterrupt
	prevQueueLen := len(e.State.InterruptQueue)
	e.dispatcher.OnTrigger(model.TriggerBeforeDraw, ctx)

	if e.State.PendingInterrupt != prevPending || len(e.State.InterruptQueue) > prevQueueLen {
		e.Log("[System] 等待响应前暂停摸牌...")
		return false
	}

	e.resumePendingDraw(ctx)
	return e.State.PendingInterrupt == nil
}

func (e *GameEngine) enqueuePendingDrawFollowup(ctx *model.Context) {
	if ctx == nil || ctx.Selections == nil {
		return
	}
	if queued, _ := ctx.Selections["draw_followup_queued"].(bool); queued {
		return
	}

	raw := ctx.Selections["draw_followup"]
	followup, ok := raw.(model.DeferredFollowup)
	if !ok || followup.Type == "" {
		return
	}

	e.State.DeferredFollowups = append([]model.DeferredFollowup{followup}, e.State.DeferredFollowups...)
	ctx.Selections["draw_followup_queued"] = true
}

func (e *GameEngine) restorePhaseAfterInterruptedDraw(ctx *model.Context) bool {
	if ctx == nil {
		return false
	}
	if e.State.PendingInterrupt != nil {
		return true
	}

	if ctx.Flags["FromDamageDraw"] {
		e.enterDamageResolution(nil)
		return true
	}

	if resumePoint := normalizeChoiceResumePoint(ctx.Selections["draw_resume_phase"]); resumePoint != "" {
		e.applyChoiceResumePoint(ctx.Selections["draw_resume_phase"])
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

func (e *GameEngine) applyAssassinStealthEffect(player *model.Player) {
	if player == nil {
		return
	}
	beforePoses := e.snapshotPlayerPoses()
	enterAssassinStealthForm(player)
	e.Log(fmt.Sprintf("%s 进入潜行形态：转为横置，手牌上限-1，无法成为主动攻击目标", player.Name))
	e.dispatchOrientationChanges(beforePoses)
	checkCtx := e.buildContext(player, nil, model.TriggerNone, nil)
	checkCtx.Flags["StayInTurn"] = true
	e.checkHandLimit(player, checkCtx)
}

func (e *GameEngine) releaseAssassinStealthEffect(player *model.Player) {
	if player == nil {
		return
	}
	if !hasAssassinStealthForm(player) {
		return
	}

	beforePoses := e.snapshotPlayerPoses()
	leaveAssassinStealthForm(player)
	e.Log(fmt.Sprintf("%s 脱离潜行形态并转正", player.Name))
	e.dispatchOrientationChanges(beforePoses)
}

func (e *GameEngine) resolveAssassinStealthApplyFollowup(f model.DeferredFollowup) error {
	user := e.State.Players[f.UserID]
	if user == nil {
		return fmt.Errorf("暗杀者潜行后续执行者不存在: %s", f.UserID)
	}
	if !isCharacter(user, "assassin") {
		return fmt.Errorf("仅暗杀者可执行潜行后续")
	}

	e.applyAssassinStealthEffect(user)
	return nil
}

func (e *GameEngine) executeResolvedDraw(ctx *model.Context, drawCount int, reason string) {
	target := ctx.User
	cards, newDeck, newDiscard := rules.DrawCards(e.State.Deck, e.State.DiscardPile, drawCount)
	e.State.Deck = newDeck
	e.State.DiscardPile = newDiscard
	target.Hand = append(target.Hand, cards...)
	e.NotifyDrawCards(target.ID, drawCount, reason)

	ctx.Trigger = model.TriggerAfterDraw
	if ctx.TriggerCtx != nil {
		ctx.TriggerCtx.Type = model.EventAfterDraw
		ctx.TriggerCtx.DrawCount = &drawCount
	}
	e.dispatcher.OnTrigger(model.TriggerAfterDraw, ctx)

	e.checkHandLimit(target, ctx)
	e.Log(fmt.Sprintf("%s 摸了 %d 张牌", target.Name, drawCount))
}
