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
	case e.isDiscardSelectionActive():
		return e.driveDiscardSelectionPhase()
	case e.isResponseWindowActive():
		return e.driveResponseRecoveryPhase()
	case e.isCombatInteractionWindow():
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

	if e.runPlayerPhaseHooks(player, turnBeforeStartPhaseHooks) {
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
	if e.runPlayerPhaseHooks(player, beforeActionPhaseHooks) {
		if e.State.PendingInterrupt != nil {
			return driveStop
		}
		return driveContinueLoop
	}
	// 其余 TriggerOnBuffPhase 的通用技能/状态仍走 dispatcher 主流程。
	skillCtx := e.buildContext(player, nil, model.TriggerOnBuffPhase, nil)
	e.dispatcher.OnTrigger(model.TriggerOnBuffPhase, skillCtx)
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
	if e.processPendingDamages() {
		return driveStop // 有中断，暂停
	}

	// 队列处理完毕，进入下一阶段
	if e.restoreReturnPoint() {
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

	if e.runPlayerPhaseHooks(player, turnStartPhaseHooks) {
		if e.State.PendingInterrupt != nil {
			return driveStop
		}
		return driveContinueLoop
	}
	player.TurnState.HasProcessedTurnStart = true
	turnStartCtx := e.buildTimedContext(player, nil, model.TriggerOnTurnStart, model.TimingOnTurnStart, eventCtx)
	e.dispatcher.OnTrigger(model.TriggerOnTurnStart, turnStartCtx)
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

	if e.runPlayerPhaseHooks(player, actionStartPhaseHooks) {
		if e.State.PendingInterrupt != nil {
			return driveStop
		}
		return driveContinueLoop
	}
	startupCtx := e.buildTimedContext(player, nil, model.TriggerOnTurnStart, model.TimingStartup, &model.EventContext{
		Type:     model.EventTurnStart,
		SourceID: currentPid,
	})
	e.dispatcher.OnTrigger(model.TriggerOnTurnStart, startupCtx)
	if e.State.PendingInterrupt != nil {
		if e.State.PendingInterrupt.Type == model.InterruptStartupSkill {
			// Startup 中断由 dispatcher 直接写入 PendingInterrupt，这里补发提示。
			prompt := e.buildStartupSkillPrompt()
			e.Notify(model.EventAskInput, "请选择是否发动启动技能", prompt)
		}
		return driveStop
	}

	// 没有启动技能，继续到 ActionSelection
	e.enterActionExecutionStage()
	return driveContinueLoop
}

func (e *GameEngine) driveActionExecutionStage(currentPid string, player *model.Player) driveOutcome {
	switch {
	case e.isActionSelectionWindow():
		return e.driveActionSelectionPhase(currentPid, player)
	case e.isBeforeActionWindow():
		return e.driveBeforeActionPhase(currentPid, player)
	case e.State.Subflow == model.SubflowNone && len(e.State.CombatStack) == 0 && e.State.TurnStage == model.TurnStageActionExecution:
		return e.driveActionExecutionRecoveryPhase(currentPid, player)
	default:
		return driveUnhandled
	}
}

func (e *GameEngine) driveActionSelectionPhase(currentPid string, player *model.Player) driveOutcome {
	e.setTurnStage(model.TurnStageActionExecution)

	state := e.buildActionSelectionOptions(currentPid, player)
	prompt := &model.Prompt{
		Type:           model.PromptConfirm,
		PlayerID:       currentPid,
		Message:        state.promptMessage,
		ChoiceType:     state.promptChoiceType,
		SkillID:        state.promptSkillID,
		Options:        state.validOptions,
		SpecialOptions: state.specialOptions,
		UIMode:         model.PromptUIModeActionHub,
	}
	e.Notify(model.EventAskInput, "请选择行动类型", prompt)
	return driveStop
}

func (e *GameEngine) driveDiscardSelectionPhase() driveOutcome {
	if e.State.PendingInterrupt == nil || !isDiscardSelectionInterruptType(e.State.PendingInterrupt.Type) {
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
	// 6. 战斗交互阶段（等待响应）
	if len(e.State.CombatStack) == 0 {
		e.Log("[Error] PhaseCombatInteraction: 战斗栈为空")
		return driveStop
	}

	// 查看栈顶战斗请求
	idx := len(e.State.CombatStack) - 1
	combatReq := &e.State.CombatStack[idx]
	target := e.State.Players[combatReq.TargetID]

	if target == nil {
		e.Log("[Error] PhaseCombatInteraction: 目标玩家不存在")
		return driveStop
	}

	for _, hook := range combatInteractionHooks {
		if hook != nil && hook(e, combatReq) {
			return driveStop
		}
	}

	// 如果强制命中，直接结算伤害
	if combatReq.IsForcedHit {
		e.Log("[Combat] 攻击强制命中！跳过响应阶段，直接结算...")
		e.clearCombatStack()
		e.AddPendingDamageFront(model.PendingDamage{
			SourceID:      combatReq.AttackerID,
			TargetID:      combatReq.TargetID,
			Damage:        combatReq.Card.Damage,
			DamageType:    "Attack",
			Card:          combatReq.Card,
			IsCounter:     combatReq.IsCounter,
			IgnoreShield:  combatReq.IgnoreShield,
			InterceptTags: model.CloneCombatInterceptTags(combatReq.InterceptTags),
		})
		e.setReturnPoint(model.TurnStageActionEnd)
		e.enterDamageResolution(nil)
		return driveContinueLoop
	}

	// 圣盾改为“承受伤害(take)时”再触发，先给玩家应战/防御的选择机会。
	shieldFallbackReady := e.hasUsableShieldForCombat(target, *combatReq)

	// 应战反弹目标：攻击方的队友（不含攻击者本人）
	var counterTargets []string
	attacker := e.State.Players[combatReq.AttackerID]
	if attacker != nil {
		for pid, p := range e.State.Players {
			if p.Camp == attacker.Camp && pid != combatReq.AttackerID {
				counterTargets = append(counterTargets, pid)
			}
		}
	}
	attackerRole := combatReq.AttackerID
	if attacker != nil {
		attackerRole = attacker.Name
	}

	// 通知目标玩家选择响应方式（无圣盾时正常选项）
	var options []model.PromptOption
	// 暗灭/暗影抗拒等交互策略在 combatInteractionHooks 中统一处理。
	noHolyDefend := combatReq.HasInterceptTag(model.CombatInterceptIgnoreTargetHoly)
	takeLabel := "承受伤害"
	if shieldFallbackReady {
		takeLabel = "承受（将触发圣盾）"
	}
	if combatReq.CanBeResponded {
		options = []model.PromptOption{{ID: "take", Label: takeLabel}}
		if !noHolyDefend {
			options = append(options, model.PromptOption{ID: "defend", Label: "防御"})
		}
		if len(counterTargets) > 0 {
			options = append(options, model.PromptOption{ID: "counter", Label: "应战"})
		}
	} else {
		options = []model.PromptOption{{ID: "take", Label: takeLabel}}
		if !noHolyDefend {
			options = append(options, model.PromptOption{ID: "defend", Label: "防御"})
		}
	}
	hints := e.buildCombatEffectHints(*combatReq, attacker)
	if shieldFallbackReady {
		hints = append(hints, "你身上有【圣盾】：若本次选择承受伤害，将优先消耗圣盾并抵挡本次攻击。")
	}
	if noHolyDefend {
		hints = append(hints, "本次攻击处于【一击无念】劫持中，不能使用【圣光】防御。")
	}

	prompt := &model.Prompt{
		Type:             model.PromptConfirm,
		PlayerID:         combatReq.TargetID,
		AttackerID:       combatReq.AttackerID,
		CounterTargetIDs: counterTargets,
		AttackElement:    string(combatReq.Card.Element), // 应战须同系或暗灭
		EffectHints:      hints,
		Message: fmt.Sprintf("%s 需要响应来自 %s 的攻击 (%s)",
			target.Name,
			attackerRole,
			combatReq.Card.Name),
		Options: options,
	}

	e.Notify(model.EventAskInput, "请选择响应方式", prompt)
	return driveStop // 等待用户输入
}

func (e *GameEngine) driveActionEndStage(currentPid string, player *model.Player) driveOutcome {
	e.setTurnStage(model.TurnStageActionEnd)

	if player.TurnState.LastActionType != "" {
		lastActionType := model.ActionType(player.TurnState.LastActionType)
		skipActionEndInterruptHooks := player.TurnState.ConsumeActionEndInterruptHookSkipOnce()
		if e.runActionEndSequence(currentPid, player, lastActionType, skipActionEndInterruptHooks) {
			return driveStop
		}
	}

	e.enterExtraActionStage()
	return driveContinueLoop
}

func (e *GameEngine) runActionEndSequence(currentPid string, player *model.Player, actionType model.ActionType, skipActionEndInterruptHooks bool) bool {
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
	skillCtx := e.buildContext(player, nil, model.TriggerOnPhaseEnd, eventCtx)
	if skillCtx.Selections == nil {
		skillCtx.Selections = map[string]interface{}{}
	}
	skillCtx.Selections["response_resume_phase"] = model.TurnStageActionEnd

	// 攻击结束钩子优先于 OnPhaseEnd（例如圣剑三连击中断）。
	if !skipActionEndInterruptHooks && e.runActionEndInterruptHooks(skillCtx) {
		return true
	}

	// 先清理行动结束标记，再派发 OnPhaseEnd，避免恢复路径重复派发。
	player.TurnState.LastActionType = ""
	player.TurnState.LastActionCard = nil

	e.dispatcher.OnTrigger(model.TriggerOnPhaseEnd, skillCtx)
	if e.State.PendingInterrupt != nil {
		// OnPhaseEnd 已派发，后续仅补执行行动后场上追加效果。
		e.enqueuePostActionEndFollowup(player.ID, actionType)
		return true
	}
	return e.handlePostActionEndEffects(player, actionType)
}

func (e *GameEngine) driveExtraActionStage(currentPid string, player *model.Player) driveOutcome {
	e.setTurnStage(model.TurnStageExtraAction)
	// 8. 额外行动阶段（处理队列）
	if len(e.State.ActionQueue) > 0 {
		// 弹出队列第一个行动
		queuedAction := e.State.ActionQueue[0]
		e.State.ActionQueue = e.State.ActionQueue[1:]

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
		e.enterTurnEndStage()
	}

	return driveContinueLoop
}

func (e *GameEngine) driveTurnEndStage(currentPid string, player *model.Player) driveOutcome {
	if e.State.Subflow != model.SubflowNone || len(e.State.CombatStack) > 0 || e.State.TurnStage != model.TurnStageTurnEnd {
		return driveUnhandled
	}

	// 9. 回合结束阶段
	if e.runPlayerPhaseHooks(player, turnEndPreExtraActionHooks) {
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
		player.TurnState.CurrentExtraAction = currentAction.MustType
		player.TurnState.CurrentExtraElement = currentAction.MustElement

		e.enterActionExecutionStage()

		// 显示行动约束信息
		constraintInfo := e.buildConstraintInfo(currentAction.MustType, currentAction.MustElement)
		e.Log(fmt.Sprintf("[Turn] %s %s 额外行动开始 (剩余 %d 次额外行动)%s",
			player.Name, currentAction.Source, len(player.TurnState.PendingActions)+1, constraintInfo))

		return driveContinueLoop
	}

	if e.runPlayerPhaseHooks(player, turnEndFinalHooks) {
		return driveStop
	}

	e.NextTurn()
	return driveContinueLoop
}

func (e *GameEngine) driveActionExecutionRecoveryPhase(currentPid string, player *model.Player) driveOutcome {
	e.setTurnStage(model.TurnStageActionExecution)
	// 行动执行阶段通常用于“行动中弹出的中断”（如魔弹融合/圣疗等）。
	// 当中断被消费后，如果没有显式阶段回切，这里负责把流程接回主状态机，
	// 避免停在 ActionExecution 导致 Drive 直接返回而卡局。
	if e.routePendingDamageWithDefaultReturn(model.TurnStageExtraAction) {
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
