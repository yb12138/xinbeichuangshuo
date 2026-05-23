// gameflow: 回合 TurnStage 状态机：各 Stage 的 drive* 实现。

package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

type driveOutcome int

const (
	driveUnhandled driveOutcome = iota
	driveContinueLoop
	driveStop
)

func (e *GameEngine) driveNonTurnPhase(currentPid string, player *model.Player) driveOutcome {
	switch {
	case e.isDamageResolutionActive():
		return e.drivePendingDamageResolutionPhase()
	case e.IsDiscardSelectionActive():
		return e.DriveDiscardSelectionPhase()
	case e.isResponseWindowActive():
		return e.driveResponseRecoveryPhase()
	case e.IsCombatInteractionWindow():
		return e.driveCombatInteractionPhase(currentPid, player)
	default:
		return driveUnhandled
	}
}

func (e *GameEngine) driveTurnFSM(currentPid string, player *model.Player) driveOutcome {
	stage := e.syncTurnStageForDispatch(player)
	switch stage {
	case model.TurnStageTurnBeforeStart:
		return e.driveTurnBeforeStartStage(player)
	case model.TurnStageBeforeAction:
		return e.driveBeforeActionStage(currentPid, player)
	case model.TurnStageTurnStart:
		return e.driveTurnStartStage(currentPid, player)
	case model.TurnStageActionStart:
		return e.driveActionStartStage(currentPid, player)
	case model.TurnStageActionExecution:
		return e.driveActionExecutionStage(currentPid, player)
	case model.TurnStageActionEnd:
		return e.driveActionEndStage(currentPid, player)
	case model.TurnStageExtraAction:
		return e.driveExtraActionStage(currentPid, player)
	case model.TurnStageTurnEnd:
		return e.driveTurnEndStage(currentPid, player)
	default:
		return driveUnhandled
	}
}

func (e *GameEngine) syncTurnStageForDispatch(player *model.Player) model.TurnStage {
	stage := e.State.TurnStage
	setAndReturn := func(next model.TurnStage) model.TurnStage {
		e.setTurnStage(next)
		return next
	}
	if stage == "" {
		stage = model.TurnStageTurnBeforeStart
	}
	if e.State.Subflow != model.SubflowNone || len(e.State.CombatStack) > 0 {
		return setAndReturn(stage)
	}
	if player != nil {
		if stage == model.TurnStageActionStart && !player.TurnState.HasProcessedTurnStart {
			return setAndReturn(model.TurnStageTurnStart)
		}
		needsActionEndCatchup := player.TurnState.LastActionType != ""
		if (stage == model.TurnStageExtraAction || stage == model.TurnStageTurnEnd) && needsActionEndCatchup {
			return setAndReturn(model.TurnStageActionEnd)
		}
	}
	switch stage {
	case model.TurnStageActionExecution:
		return stage
	case model.TurnStageActionEnd, model.TurnStageExtraAction, model.TurnStageTurnEnd,
		model.TurnStageTurnBeforeStart, model.TurnStageBeforeAction, model.TurnStageTurnStart, model.TurnStageActionStart:
		return setAndReturn(stage)
	default:
		if player != nil && player.TurnState.LastActionType != "" {
			stage = model.TurnStageActionEnd
		} else if player != nil && player.TurnState.HasProcessedTurnStart {
			stage = model.TurnStageActionStart
		} else {
			stage = model.TurnStageTurnBeforeStart
		}
	}
	return setAndReturn(stage)
}

func (e *GameEngine) driveTurnBeforeStartStage(player *model.Player) driveOutcome {
	if e.State.Subflow != model.SubflowNone || len(e.State.CombatStack) > 0 || e.State.TurnStage != model.TurnStageTurnBeforeStart {
		return driveUnhandled
	}

	if e.runTimingTurnStartBeforeStartHooks(player) {
		if e.State.PendingInterrupt != nil {
			return driveStop
		}
		return driveContinueLoop
	}

	e.setTurnStage(model.TurnStageBeforeAction)
	return driveContinueLoop
}

func (e *GameEngine) driveBeforeActionStage(currentPid string, player *model.Player) driveOutcome {
	if e.State.Subflow != model.SubflowNone || len(e.State.CombatStack) > 0 || e.State.TurnStage != model.TurnStageBeforeAction {
		return driveUnhandled
	}

	// 回合开始前先按固定顺序结算基础效果 hook（如中毒、五系束缚、虚弱）。
	if e.runActionBoundaryTimingStageHooks(player, actionBoundaryResolveField) {
		if e.State.PendingInterrupt != nil {
			return driveStop
		}
		return driveContinueLoop
	}
	// 角色 TimingActionStart hooks（如精疲力竭结束结算）。
	if e.runActionBoundaryTimingStageHooks(player, actionBoundaryResolveActionStart) {
		if e.State.PendingInterrupt != nil {
			return driveStop
		}
		return driveContinueLoop
	}
	// 其余 TimingActionBefore 的通用技能/状态仍走 dispatcher 主流程。
	skillCtx := e.BuildContext(player, nil, model.TimingActionBefore, nil)
	e.dispatcher.OnTiming(skillCtx.Timing, skillCtx)
	if e.State.PendingInterrupt != nil {
		return driveStop
	}

	if e.State.TurnStage == model.TurnStageTurnEnd {
		return driveStop
	}

	// 处理完 BuffResolve 后，检查是否有延迟伤害需要结算
	if !e.routePendingDamageWithReturn(model.TurnStageTurnStart) {
		e.clearSubflow()
		e.clearCombatStage()
	}
	e.setTurnStage(model.TurnStageTurnStart)
	return driveContinueLoop
}

func (e *GameEngine) drivePendingDamageResolutionPhase() driveOutcome {
	// 延迟伤害结算阶段
	if e.ProcessPendingDamages() {
		return driveStop // 有中断，暂停
	}

	// 队列处理完毕，进入下一阶段
	if e.RestoreReturnPoint() {
	} else {
		e.clearSubflow()
		e.clearCombatStage()
		e.setTurnStage(model.TurnStageActionStart)
	}

	return driveContinueLoop
}

func (e *GameEngine) driveTurnStartStage(currentPid string, player *model.Player) driveOutcome {
	if e.State.Subflow != model.SubflowNone || len(e.State.CombatStack) > 0 || e.State.TurnStage != model.TurnStageTurnStart {
		return driveUnhandled
	}

	if player.Tokens == nil {
		player.Tokens = map[string]int{}
	}
	eventCtx := &model.EventContext{
		Type:     model.EventTurnStart,
		SourceID: currentPid,
	}
	if player.TurnState.HasProcessedTurnStart {
		e.setTurnStage(model.TurnStageActionStart)
		return driveContinueLoop
	}

	if e.runTimingTurnStartHooks(player) {
		if e.State.PendingInterrupt != nil {
			return driveStop
		}
		return driveContinueLoop
	}
	player.TurnState.HasProcessedTurnStart = true
	turnStartCtx := e.BuildTimedContext(player, nil, model.TimingTurnStart, eventCtx)
	e.dispatcher.OnTiming(turnStartCtx.Timing, turnStartCtx)
	if e.State.PendingInterrupt != nil {
		return driveStop
	}

	e.setTurnStage(model.TurnStageActionStart)
	return driveContinueLoop
}

func (e *GameEngine) driveActionStartStage(currentPid string, player *model.Player) driveOutcome {
	if e.State.Subflow != model.SubflowNone || len(e.State.CombatStack) > 0 || e.State.TurnStage != model.TurnStageActionStart {
		return driveUnhandled
	}

	if e.runActionBoundaryTimingStageHooks(player, actionBoundaryResolveActionStart) {
		if e.State.PendingInterrupt != nil {
			return driveStop
		}
		return driveContinueLoop
	}
	startupCtx := e.BuildTimedContext(player, nil, model.TimingActionStart, &model.EventContext{
		Type:     model.EventTurnStart,
		SourceID: currentPid,
	})
	e.dispatcher.OnTiming(startupCtx.Timing, startupCtx)
	if e.State.PendingInterrupt != nil {
		if e.State.PendingInterrupt.Type == model.InterruptStartupSkill {
			// Startup 中断由 dispatcher 直接写入 PendingInterrupt，这里补发提示。
			prompt := e.buildStartupSkillPrompt()
			e.NotifyPrompt(prompt)
		}
		return driveStop
	}

	// 没有启动技能，继续到 ActionSelection
	e.enterActionExecutionStage()
	return driveContinueLoop
}

func (e *GameEngine) driveActionExecutionStage(currentPid string, player *model.Player) driveOutcome {
	switch {
	case e.needsActionExecutionActionEndCatchup(player):
		lastActionType := "<nil>"
		if player != nil {
			lastActionType = player.TurnState.LastActionType
		}
		e.Log(fmt.Sprintf("[Debug] ActionExecution 命中 ActionEnd 补结算: player=%s last_action_type=%s action_queue=%d", currentPid, lastActionType, len(e.State.ActionQueue)))
		return e.driveActionExecutionRecoveryPhase(currentPid, player)
	case e.IsActionSelectionWindow() && (player == nil || player.TurnState.LastActionType == ""):
		// 启动技结算后会回到 ActionExecution，但此时没有“刚结束的行动”可补收尾，
		// 需要直接回到行动选择，而不是继续推入 ActionEnd/ExtraAction。
		return e.driveActionSelectionPhase(currentPid, player)
	case e.IsActionSelectionWindow():
		return e.driveActionSelectionPhase(currentPid, player)
	case e.IsBeforeActionWindow():
		return e.driveBeforeActionPhase(currentPid, player)
	case e.State.Subflow == model.SubflowNone && len(e.State.CombatStack) == 0 && e.State.TurnStage == model.TurnStageActionExecution:
		return e.driveActionExecutionRecoveryPhase(currentPid, player)
	default:
		return driveUnhandled
	}
}

func (e *GameEngine) driveActionSelectionPhase(currentPid string, player *model.Player) driveOutcome {
	e.setTurnStage(model.TurnStageActionExecution)

	state := e.BuildActionSelectionOptions(currentPid, player)
	prompt := &model.Prompt{
		Type:           model.PromptConfirm,
		PlayerID:       currentPid,
		Message:        state.promptMessage,
		ChoiceType:     state.promptChoiceType,
		SkillID:        state.promptSkillID,
		Options:        state.ValidOptions,
		SpecialOptions: state.specialOptions,
		UIMode:         model.PromptUIModeActionHub,
		Presentation:   &model.PromptPresentation{Kind: model.PresentationActionHub},
	}
	e.NotifyPrompt(prompt)
	return driveStop
}

func (e *GameEngine) DriveDiscardSelectionPhase() driveOutcome {
	if e.State.PendingInterrupt == nil || !IsDiscardSelectionInterrupt(e.State.PendingInterrupt) {
		return driveUnhandled
	}
	return driveStop
}

// driveBeforeActionPhase 驱动「行动阶段中、队首行动尚未真正结算」的一段：从队列取出攻击/法术并交给规则链。
// 与回合阶段 TurnStageBeforeAction（回合开始前的中毒/束缚等）不同，此处始终是 ActionExecution 内的执行前窗口。
func (e *GameEngine) driveBeforeActionPhase(currentPid string, player *model.Player) driveOutcome {
	e.setTurnStage(model.TurnStageActionExecution)

	head, immediate, stop := e.beforeActionPeekReadyHead(player)
	if stop {
		return immediate
	}

	switch head.Type {
	case model.ActionAttack:
		return e.driveBeforeActionAttack(currentPid, player, head)
	case model.ActionMagic:
		return e.driveBeforeActionMagic(currentPid, player, head)
	default:
		// 正常对局只应入队 Attack/Magic；其它类型视为状态异常，丢弃以免卡在行动阶段。
		e.Log(fmt.Sprintf("[Error] PhaseBeforeAction: 不支持的队列行动类型 %s", head.Type))
		e.State.ActionQueue = e.State.ActionQueue[1:]
		return e.beforeActionRecoverAfterDroppedHead()
	}
}

func (e *GameEngine) driveCombatInteractionPhase(currentPid string, player *model.Player) driveOutcome {
	e.setCombatStage(model.CombatStageHitCheck)
	combatReq, target, attacker, ok := e.peekCombatInteractionRequest()
	if !ok {
		return driveStop
	}
	if e.RunAttackResponseCombatInteractionPolicies(combatReq) {
		return driveStop
	}
	attackKind := attackKindFromCounter(combatReq.IsCounter)
	if e.dispatchAttackRulebookTiming(model.TimingAttackForceHitCheck, attacker, target, combatReq.Card, attackInfoFromCombatRequest(combatReq, false), attackKind) {
		return driveStop
	}
	if e.resolveForcedHitCombat(combatReq) {
		return driveContinueLoop
	}
	if e.dispatchAttackRulebookTiming(model.TimingAttackNoResponseCheck, attacker, target, combatReq.Card, attackInfoFromCombatRequest(combatReq, false), attackKind) {
		return driveStop
	}

	shieldFallbackReady := e.HasUsableShieldForCombat(target, *combatReq)
	counterTargets := e.buildCombatCounterTargets(combatReq.AttackerID)
	options := e.buildCombatResponseOptions(combatReq, shieldFallbackReady, counterTargets)
	hints := e.buildCombatInteractionHints(*combatReq, shieldFallbackReady)
	if e.dispatchAttackRulebookTiming(model.TimingAttackResponse, attacker, target, combatReq.Card, attackInfoFromCombatRequest(combatReq, false), attackKind) {
		return driveStop
	}

	attackerRole := combatReq.AttackerID
	if attacker != nil {
		attackerRole = attacker.Name
	}
	prompt := &model.Prompt{
		Type:             model.PromptConfirm,
		PlayerID:         combatReq.TargetID,
		AttackerID:       combatReq.AttackerID,
		CounterTargetIDs: counterTargets,
		AttackElement:    e.promptAttackElementForCombatResponse(combatReq, attacker), // 应战须同系或暗灭
		EffectHints:      hints,
		Message: fmt.Sprintf("%s 需要响应来自 %s 的攻击 (%s)",
			target.Name,
			attackerRole,
			combatReq.Card.Name),
		Options:      options,
		Presentation: &model.PromptPresentation{Kind: model.PresentationResponse, Layout: "inline"},
	}
	e.NotifyPrompt(prompt)
	return driveStop
}

func (e *GameEngine) peekCombatInteractionRequest() (*model.CombatRequest, *model.Player, *model.Player, bool) {
	if len(e.State.CombatStack) == 0 {
		e.Log("[Error] PhaseCombatInteraction: 战斗栈为空")
		return nil, nil, nil, false
	}
	combatReq := &e.State.CombatStack[len(e.State.CombatStack)-1]
	target := e.State.Players[combatReq.TargetID]
	if target == nil {
		e.Log("[Error] PhaseCombatInteraction: 目标玩家不存在")
		return nil, nil, nil, false
	}
	attacker := e.State.Players[combatReq.AttackerID]
	return combatReq, target, attacker, true
}

func (e *GameEngine) resolveForcedHitCombat(combatReq *model.CombatRequest) bool {
	if !combatReq.IsForcedHit {
		return false
	}
	e.Log("[Combat] 攻击强制命中！跳过响应阶段，直接结算...")
	e.clearCombatStack()
	e.AddPendingDamageFront(model.PendingDamage{
		SourceID:      combatReq.AttackerID,
		TargetID:      combatReq.TargetID,
		Damage:        combatReq.Card.Damage,
		DamageType:    model.AttackDamage,
		Card:          combatReq.Card,
		IsCounter:     combatReq.IsCounter,
		IgnoreShield:  combatReq.IgnoreShield,
		InterceptTags: model.CloneCombatInterceptTags(combatReq.InterceptTags),
	})
	e.enterDamageResolution(model.TurnStageActionEnd)
	return true
}

func (e *GameEngine) canUseHolyDefend(combatReq *model.CombatRequest) bool {
	return combatReq != nil && !combatReq.HasInterceptTag(model.CombatInterceptIgnoreTargetHoly)
}

func (e *GameEngine) canUseCounter(combatReq *model.CombatRequest) bool {
	return combatReq != nil && combatReq.CanBeResponded
}

func (e *GameEngine) promptAttackElementForCombatResponse(combatReq *model.CombatRequest, attacker *model.Player) string {
	if combatReq == nil || combatReq.Card == nil {
		return ""
	}
	if combatReq.Card.Element == model.ElementDark {
		return string(model.ElementDark)
	}
	// 圣枪骑士【天枪】文案口径：本次攻击视为暗灭且无法应战。
	if combatReq != nil && combatReq.ElementOverride != "" {
		return combatReq.ElementOverride
	}
	return string(combatReq.Card.Element)
}

func (e *GameEngine) buildCombatCounterTargets(attackerID string) []string {
	attacker := e.State.Players[attackerID]
	if attacker == nil {
		return nil
	}
	counterTargets := make([]string, 0, len(e.State.PlayerOrder))
	for _, pid := range e.State.PlayerOrder {
		if pid == attackerID {
			continue
		}
		p := e.State.Players[pid]
		if p != nil && p.Camp == attacker.Camp {
			counterTargets = append(counterTargets, pid)
		}
	}
	return counterTargets
}

func (e *GameEngine) buildCombatResponseOptions(combatReq *model.CombatRequest, shieldFallbackReady bool, counterTargets []string) []model.PromptOption {
	takeLabel := "承受伤害"
	if shieldFallbackReady {
		takeLabel = "承受（将触发圣盾）"
	}
	options := []model.PromptOption{{ID: "take", Label: takeLabel}}
	if e.canUseHolyDefend(combatReq) {
		options = append(options, model.PromptOption{ID: "defend", Label: "防御"})
	}
	if e.canUseCounter(combatReq) && len(counterTargets) > 0 {
		options = append(options, model.PromptOption{ID: "counter", Label: "应战"})
	}
	return options
}

func (e *GameEngine) buildCombatInteractionHints(combatReq model.CombatRequest, shieldFallbackReady bool) []string {
	hints := make([]string, 0, 4)
	if combatReq.HasInterceptTag(model.CombatInterceptIgnoreHolyShield) || combatReq.IgnoreShield {
		hints = append(hints, "本次攻击无视【圣盾】。")
	}
	if combatReq.HasInterceptTag(model.CombatInterceptUnrespondable) || !e.canUseCounter(&combatReq) {
		hints = append(hints, "本次攻击无法应战。")
	}
	if shieldFallbackReady {
		hints = append(hints, "你身上有【圣盾】：若本次选择承受伤害，将优先消耗圣盾并抵挡本次攻击。")
	}
	if !e.canUseHolyDefend(&combatReq) {
		hints = append(hints, "本次攻击处于【一击无念】劫持中，不能使用【圣光】防御。")
	}
	return hints
}

func (e *GameEngine) driveActionEndStage(currentPid string, player *model.Player) driveOutcome {
	e.setTurnStage(model.TurnStageActionEnd)

	if player.TurnState.LastActionType != "" {
		lastActionType := model.ActionType(player.TurnState.LastActionType)
		if e.runActionEndSequence(currentPid, player, lastActionType) {
			return driveStop
		}
	}

	if !e.routePendingDamageWithReturn(model.TurnStageExtraAction) {
		e.enterExtraActionStage()
	}
	return driveContinueLoop
}

func (e *GameEngine) runActionEndSequence(currentPid string, player *model.Player, actionType model.ActionType) bool {
	if e == nil || player == nil || actionType == "" {
		return false
	}

	var phaseEndCard *model.Card
	if player.TurnState.LastActionCard != nil {
		cardCopy := *player.TurnState.LastActionCard
		phaseEndCard = &cardCopy
	}
	eventCtx := &model.EventContext{
		Type:       model.EventPhaseEnd,
		SourceID:   currentPid,
		Card:       phaseEndCard,
		ActionType: actionType,
	}
	if actionType == model.ActionAttack {
		eventCtx.AttackInfo = &model.AttackEventInfo{
			ActionType:       string(model.ActionAttack),
			CounterInitiator: "",
		}
	}
	skillCtx := e.BuildContext(player, nil, model.TimingActionEnd, eventCtx)
	if skillCtx.Selections == nil {
		skillCtx.Selections = map[string]interface{}{}
	}
	skillCtx.Selections["response_resume_phase"] = model.TurnStageActionEnd

	// 攻击结束钩子优先于 OnPhaseEnd（例如圣剑三连击中断）。
	// 先清理行动结束标记，再派发 OnPhaseEnd，避免恢复路径重复派发。
	player.TurnState.LastActionType = ""
	player.TurnState.LastActionCard = nil

	e.dispatcher.OnTiming(skillCtx.Timing, skillCtx)
	if e.State.PendingInterrupt != nil {
		// OnPhaseEnd 已派发，后续仅补执行行动后场上追加效果。
		e.queuePostActionEndResume(player.ID, actionType)
		return true
	}
	return e.HandlePostActionEndEffects(player, actionType)
}

func (e *GameEngine) driveExtraActionStage(currentPid string, player *model.Player) driveOutcome {
	e.setTurnStage(model.TurnStageExtraAction)
	pendingTokenCount := 0
	if player != nil {
		pendingTokenCount = len(player.TurnState.PendingActions)
	}
	e.Log(fmt.Sprintf("[Debug] ExtraAction 阶段: player=%s action_queue=%d pending_action_tokens=%d", currentPid, len(e.State.ActionQueue), pendingTokenCount))
	// 8. 额外行动阶段（处理队列）
	if len(e.State.ActionQueue) > 0 {
		// 弹出队列第一个行动
		queuedAction := e.State.ActionQueue[0]
		e.State.ActionQueue = e.State.ActionQueue[1:]
		e.Log(fmt.Sprintf("[Debug] ExtraAction 消费队列行动: type=%s element=%s source=%s remaining_action_queue=%d", queuedAction.Type, queuedAction.Element, queuedAction.SourceID, len(e.State.ActionQueue)))

		// 设置当前额外行动约束
		player.TurnState.CurrentExtraAction = string(queuedAction.Type)
		if queuedAction.Element != "" {
			// 【修改点】将单个 Element 包装成切片
			player.TurnState.CurrentExtraElement = []model.Element{queuedAction.Element}
		} else {
			// 如果没有限制，置为 nil (或空切片)
			player.TurnState.CurrentExtraElement = nil
		}

		// 设置阶段为 BeforeAction
		e.enterActionExecutionStage()
	} else {
		// 队列为空，进入回合结束
		e.Log("[Debug] ExtraAction 队列为空，进入 TurnEnd")
		e.enterTurnEndStage()
	}

	return driveContinueLoop
}

func (e *GameEngine) driveTurnEndStage(currentPid string, player *model.Player) driveOutcome {
	if e.State.Subflow != model.SubflowNone || len(e.State.CombatStack) > 0 || e.State.TurnStage != model.TurnStageTurnEnd {
		return driveUnhandled
	}
	pendingTokenCount := 0
	if player != nil {
		pendingTokenCount = len(player.TurnState.PendingActions)
	}
	e.Log(fmt.Sprintf("[Debug] TurnEnd 阶段: player=%s pending_action_tokens=%d", currentPid, pendingTokenCount))

	// 9. 回合结束阶段
	if e.runTimingTurnEndPreExtraHooks(player) {
		return driveStop
	}
	// 检查是否有待执行的行动令牌 (处理额外行动)
	// 将PendingActions逻辑迁移至此
	if len(player.TurnState.PendingActions) > 0 {
		// 取出第一个行动令牌
		currentAction := player.TurnState.PendingActions[0]
		player.TurnState.PendingActions = player.TurnState.PendingActions[1:]

		// 重置行动状态，允许再次行动
		player.TurnState.HasActed = false

		// 设置 TurnState 中的约束，然后调用 Drive (进入 ActionSelection)
		if currentAction.MustType == "" {
			player.TurnState.CurrentExtraAction = model.ExtraActionAny
		} else {
			player.TurnState.CurrentExtraAction = currentAction.MustType
		}
		player.TurnState.CurrentExtraElement = currentAction.MustElement

		e.enterActionExecutionStage()

		// 显示行动约束信息
		constraintInfo := e.buildConstraintInfo(player.TurnState.CurrentExtraAction, currentAction.MustElement)
		e.Log(fmt.Sprintf("[Turn] %s %s 额外行动开始 (剩余 %d 次额外行动)%s",
			player.Name, currentAction.Source, len(player.TurnState.PendingActions)+1, constraintInfo))

		return driveContinueLoop
	}

	if e.runTimingTurnEndFinalHooks(player) {
		return driveStop
	}

	e.NextTurn()
	return driveContinueLoop
}

func (e *GameEngine) driveActionExecutionRecoveryPhase(currentPid string, player *model.Player) driveOutcome {
	e.setTurnStage(model.TurnStageActionExecution)
	// 行动执行阶段通常用于"行动中弹出的中断"（如魔弹融合/圣疗等）。
	// 当中断被消费后，如果没有显式阶段回切，这里负责把流程接回主状态机，
	// 避免停在 ActionExecution 导致 Drive 直接返回而卡局。
	//
	// 返回点判断：
	// - 启动技能场景（LastActionType == ""）：伤害结算后应回到 ActionExecution，进入行动选择
	// - 正常行动结束场景（LastActionType != ""）：伤害结算后应回到 ExtraAction，检查额外行动队列
	defaultReturn := model.TurnStageExtraAction
	if player != nil && player.TurnState.LastActionType == "" {
		defaultReturn = model.TurnStageActionExecution
	}
	if e.routePendingDamageWithDefaultReturn(defaultReturn) {
		return driveContinueLoop
	}
	if len(e.State.CombatStack) > 0 {
		e.clearSubflow()
		if e.State.CombatStage == model.CombatStageNone {
			e.setCombatStage(model.CombatStageHitCheck)
		}
		return driveContinueLoop
	}
	if len(e.State.ActionQueue) > 0 {
		e.enterActionExecutionStage()
		return driveContinueLoop
	}
	e.enterActionEndStage()
	return driveContinueLoop
}
