package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

type responseCompletionKind string

const (
	responseCompletionSkip    responseCompletionKind = "skip"
	responseCompletionConfirm responseCompletionKind = "confirm"
)

type responseResumeState struct {
	kind          responseCompletionKind
	interruptType model.InterruptType
	playerID      string
	skillID       string
	offeredSkills []string

	resumeDrawCtx        *model.Context
	resumeAttackHitCtx   *model.Context
	resumeDamageTakenCtx *model.Context
	resumeAttackMissCtx  *model.Context
	resumeMoraleCtx      *model.Context
	resumePhaseEndCtx    *model.Context

	responseResumePoint string
}

type responseSkipHook func(e *GameEngine, state *responseResumeState)

var responseSkipHooks = []responseSkipHook{
	holyLancerEarthSkippedResponseHook,
}

func (e *GameEngine) captureResponseResumeStateFromInterrupt(kind responseCompletionKind, skillID string, intr *model.Interrupt) responseResumeState {
	state := responseResumeState{
		kind:    kind,
		skillID: skillID,
	}
	if intr == nil {
		return state
	}
	state.interruptType = intr.Type
	state.playerID = intr.PlayerID
	state.offeredSkills = append(state.offeredSkills, intr.SkillIDs...)
	state.captureContextValue(intr.Context)
	return state
}

func (e *GameEngine) captureResponseResumeStateFromContext(kind responseCompletionKind, skillID string, ctx *model.Context) responseResumeState {
	state := responseResumeState{
		kind:          kind,
		interruptType: model.InterruptResponseSkill,
		skillID:       skillID,
		offeredSkills: []string{skillID},
	}
	if ctx == nil {
		return state
	}
	if ctx.User != nil {
		state.playerID = ctx.User.ID
	}
	state.captureContext(ctx)
	return state
}

func (s *responseResumeState) captureContextValue(raw interface{}) {
	switch data := raw.(type) {
	case *model.Context:
		s.captureContext(data)
	case map[string]interface{}:
		if ctx, ok := data["user_ctx"].(*model.Context); ok {
			s.captureContext(ctx)
		}
	}
}

func (s *responseResumeState) captureContext(ctx *model.Context) {
	if ctx == nil {
		return
	}
	switch ctx.Trigger {
	case model.TriggerBeforeDraw:
		if s.resumeDrawCtx == nil {
			s.resumeDrawCtx = ctx
		}
	case model.TriggerOnAttackHit:
		if s.resumeAttackHitCtx == nil {
			s.resumeAttackHitCtx = ctx
		}
	case model.TriggerOnDamageTaken:
		if s.resumeDamageTakenCtx == nil {
			s.resumeDamageTakenCtx = ctx
		}
	case model.TriggerOnAttackMiss:
		if s.resumeAttackMissCtx == nil {
			s.resumeAttackMissCtx = ctx
		}
	case model.TriggerBeforeMoraleLoss:
		if s.resumeMoraleCtx == nil {
			s.resumeMoraleCtx = ctx
		}
	case model.TriggerOnPhaseEnd:
		if s.resumePhaseEndCtx == nil {
			s.resumePhaseEndCtx = ctx
		}
	}
	if point := normalizeChoiceResumePoint(ctx.Selections["response_resume_phase"]); point != "" {
		s.responseResumePoint = point
	}
}

func (s responseResumeState) offeredSkill(skillID string) bool {
	for _, sid := range s.offeredSkills {
		if sid == skillID {
			return true
		}
	}
	return false
}

func (e *GameEngine) prepareConfirmedResponseResume(state responseResumeState) {
	if state.resumeDrawCtx != nil {
		e.resumePendingDraw(state.resumeDrawCtx)
	}
	if state.resumeAttackHitCtx != nil {
		e.markPendingAttackDamageHitProcessed(state.resumeAttackHitCtx)
	}
}

func (e *GameEngine) runResponseSkipHooks(state *responseResumeState) {
	if state == nil {
		return
	}
	for _, hook := range responseSkipHooks {
		hook(e, state)
		if e.State.PendingInterrupt != nil {
			return
		}
	}
}

func holyLancerEarthSkippedResponseHook(e *GameEngine, state *responseResumeState) {
	if state == nil || state.kind != responseCompletionSkip || state.resumeAttackHitCtx == nil {
		return
	}
	if !state.offeredSkill("holy_lancer_earth_spear") || state.playerID == "" {
		return
	}
	user := e.State.Players[state.playerID]
	if user == nil || !e.isHolyLancer(user) {
		return
	}
	if user.Tokens == nil {
		user.Tokens = map[string]int{}
	}
	// 未发动地枪时，补触发圣击（若本次攻击未被天枪/地枪阻断）。
	if user.Tokens["holy_lancer_block_sacred_strike"] != 0 {
		return
	}
	e.Heal(user.ID, 1)
	e.Log(fmt.Sprintf("%s 未发动 [地枪]，触发 [圣击]：+1治疗", user.Name))
}

func (e *GameEngine) restoreSkippedResponseAfterPop(state responseResumeState) bool {
	if state.resumeDrawCtx != nil {
		e.resumePendingDraw(state.resumeDrawCtx)
		if e.State.PendingInterrupt == nil {
			e.restorePhaseAfterInterruptedDraw(state.resumeDrawCtx)
		}
	}
	if state.resumeAttackHitCtx != nil && e.markPendingAttackDamageHitProcessed(state.resumeAttackHitCtx) {
		if e.State.PendingInterrupt == nil {
			e.enterDamageResolution(nil)
		}
		return true
	}
	if state.resumeDamageTakenCtx != nil && len(e.State.PendingDamageQueue) > 0 {
		if e.State.PendingInterrupt == nil {
			e.enterDamageResolution(nil)
		}
		return true
	}
	if state.resumeAttackMissCtx != nil && e.resumePendingAttackMiss(state.resumeAttackMissCtx) {
		return true
	}
	if state.resumeMoraleCtx != nil && e.State.PendingInterrupt == nil && e.resumePendingMoraleLoss(state.resumeMoraleCtx) {
		return true
	}
	if e.State.PendingInterrupt != nil {
		return true
	}
	if state.resumeDrawCtx != nil {
		return true
	}
	if state.resumePhaseEndCtx != nil {
		e.enterExtraActionStage()
		return true
	}
	if state.responseResumePoint != "" {
		if len(e.State.PendingDamageQueue) > 0 {
			e.setReturnPoint(state.responseResumePoint)
			e.enterDamageResolution(nil)
		} else {
			e.applyChoiceResumePoint(state.responseResumePoint)
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

func (e *GameEngine) restoreConfirmedResponseAfterPop(state responseResumeState) bool {
	if e.State.PendingInterrupt != nil {
		return true
	}
	if state.resumeDrawCtx != nil {
		e.restorePhaseAfterInterruptedDraw(state.resumeDrawCtx)
		return true
	}
	if state.resumeAttackMissCtx != nil && e.resumePendingAttackMiss(state.resumeAttackMissCtx) {
		return true
	}
	if state.resumeMoraleCtx != nil && e.resumePendingMoraleLoss(state.resumeMoraleCtx) {
		return true
	}
	if state.responseResumePoint != "" {
		if len(e.State.PendingDamageQueue) > 0 {
			e.setReturnPoint(state.responseResumePoint)
			e.enterDamageResolution(nil)
		} else {
			e.applyChoiceResumePoint(state.responseResumePoint)
		}
		return true
	}
	if len(e.State.ActionStack) > 0 {
		e.enterResponseWindow()
	} else if len(e.State.ActionQueue) > 0 {
		e.enterActionExecutionStage()
	} else {
		defaultReturn := interface{}(model.TurnStageTurnEnd)
		if state.resumePhaseEndCtx != nil {
			defaultReturn = model.TurnStageExtraAction
		}
		if !e.routePendingDamageWithDefaultReturn(defaultReturn) {
			e.enterTurnEndStage()
		}
	}
	return true
}
