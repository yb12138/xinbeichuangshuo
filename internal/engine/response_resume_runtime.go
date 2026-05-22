// gameflow: 从暂停点恢复攻击命中/承伤等 FlowTiming 上下文。

package engine

import (
	engineplayer "starcup-engine/internal/engine/player"
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

	responseResumePoint interface{}
}

func (e *GameEngine) clearActionEndCatchupMarkerAfterResponse(state responseResumeState) {
	if state.resumePhaseEndCtx == nil || state.playerID == "" {
		return
	}
	player := e.State.Players[state.playerID]
	if player == nil {
		return
	}
	// ActionEnd 响应窗口已经处理完成时，清理遗留标记，避免状态机在 ExtraAction/TurnEnd 再次补跑同一轮 ActionEnd。
	player.TurnState.LastActionType = ""
	player.TurnState.LastActionCard = nil
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
	if s.resumeDrawCtx == nil && ctx.EventCtx != nil && ctx.EventCtx.DrawCount != nil &&
		(ctx.EventCtx.Type == model.EventBeforeDraw || ctx.EventCtx.Type == model.EventAfterDraw) {
		s.resumeDrawCtx = ctx
	}
	switch {
	case ctx.BeforeDrawPhase():
		if s.resumeDrawCtx == nil {
			s.resumeDrawCtx = ctx
		}
	case ctx.ResumeAttackHitPhase():
		if s.resumeAttackHitCtx == nil {
			s.resumeAttackHitCtx = ctx
		}
	case ctx.ResumeDamageTakenPhase():
		if s.resumeDamageTakenCtx == nil {
			s.resumeDamageTakenCtx = ctx
		}
	case ctx.ResumeAttackMissPhase():
		if s.resumeAttackMissCtx == nil {
			s.resumeAttackMissCtx = ctx
		}
	case ctx.ResumeBeforeMoraleLossPhase():
		if s.resumeMoraleCtx == nil {
			s.resumeMoraleCtx = ctx
		}
	case ctx.ResumeActionEndPhase():
		if s.resumePhaseEndCtx == nil {
			s.resumePhaseEndCtx = ctx
		}
	}
	if point, ok := choiceResumePointValue(ctx.Selections["response_resume_phase"]); ok {
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

// runResponseSkillSkipEffects 在"跳过响应"时执行后效。
func (e *GameEngine) runResponseSkillSkipEffects(state *responseResumeState) {
	if state == nil {
		return
	}
	e.dispatchRoleTimingHook(engineplayer.TimingResponseSkillSkip, engineplayer.TimingHookContext{
		TargetID:       state.playerID,
		OfferedSkillID: state.skillID,
		OfferedSkills:  state.offeredSkills,
		ResumePhase:    resumePhaseFromState(state),
		InterruptType:  string(state.interruptType),
	})
}

func resumePhaseFromState(state *responseResumeState) string {
	if state == nil {
		return ""
	}
	if state.resumeAttackHitCtx != nil {
		return "attack_hit"
	}
	if state.resumeDamageTakenCtx != nil {
		return "damage_taken"
	}
	if state.resumeDrawCtx != nil {
		return "draw"
	}
	if state.resumeAttackMissCtx != nil {
		return "attack_miss"
	}
	if state.resumeMoraleCtx != nil {
		return "morale_loss"
	}
	if state.resumePhaseEndCtx != nil {
		return "action_end"
	}
	return ""
}

func (e *GameEngine) restoreSkippedResponseAfterPop(state responseResumeState) bool {
	e.clearActionEndCatchupMarkerAfterResponse(state)

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
	if hasChoiceResumePoint(state.responseResumePoint) {
		if e.routePendingDamageWithReturn(state.responseResumePoint) {
			return true
		}
		e.applyChoiceResumePoint(state.responseResumePoint)
		return true
	}
	if state.resumePhaseEndCtx != nil {
		e.enterExtraActionStage()
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
	e.clearActionEndCatchupMarkerAfterResponse(state)

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
	if hasChoiceResumePoint(state.responseResumePoint) {
		if e.routePendingDamageWithReturn(state.responseResumePoint) {
			return true
		}
		e.applyChoiceResumePoint(state.responseResumePoint)
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
		if !e.routePendingDamageWithReturn(defaultReturn) {
			e.enterTurnEndStage()
		}
	}
	return true
}
