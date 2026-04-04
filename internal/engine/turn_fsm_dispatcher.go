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
	// 弃牌阶段应当伴随 PendingInterrupt(Discard)。
	// 若中断已被消费但阶段未恢复，修复到可继续推进的阶段，避免空转。
	if e.State.PendingInterrupt == nil {
		e.Log("[Warn] PhaseDiscardSelection: 无待处理中断，执行阶段修复")
		if e.routePendingDamageWithReturn(model.TurnStageExtraAction) {
			return driveContinueLoop
		}
		if len(e.State.ActionQueue) > 0 {
			e.enterActionExecutionStage()
		} else if len(e.State.CombatStack) > 0 {
			e.clearSubflow()
			if e.State.CombatStage == model.CombatStageNone {
				e.setCombatStage(model.CombatStageHitCheck)
			}
		} else {
			e.enterTurnEndStage()
		}
		return driveContinueLoop
	}
	return driveStop
}

func (e *GameEngine) driveBeforeActionPhase(currentPid string, player *model.Player) driveOutcome {
	e.setTurnStage(model.TurnStageActionExecution)
	// 4. 行动前阶段
	// 从队列中获取当前行动（不弹出，因为后续阶段可能需要使用）
	if len(e.State.ActionQueue) == 0 {
		e.Log("[Warn] PhaseBeforeAction: 行动队列为空，执行阶段修复")
		if e.routePendingDamageWithDefaultReturn(model.TurnStageExtraAction) {
			return driveContinueLoop
		}
		e.enterExtraActionStage()
		return driveContinueLoop
	}

	currentAction := e.State.ActionQueue[0] // 只读取，不弹出
	if !currentAction.UsesVirtualCard {
		if !e.repairQueuedActionCard(player, &e.State.ActionQueue[0]) {
			e.Log("[Warn] PhaseBeforeAction: 无法修复队列中的卡牌索引，丢弃该行动")
			e.State.ActionQueue = e.State.ActionQueue[1:]
			if e.routePendingDamageWithReturn(model.TurnStageExtraAction) {
				return driveContinueLoop
			}
			if len(e.State.ActionQueue) > 0 {
				e.enterActionExecutionStage()
			} else {
				e.enterExtraActionStage()
			}
			return driveContinueLoop
		}
		currentAction = e.State.ActionQueue[0]
	}
	if currentAction.Card == nil {
		e.Log("[Warn] PhaseBeforeAction: 队列中的卡牌数据缺失，丢弃该行动")
		e.State.ActionQueue = e.State.ActionQueue[1:]
		if e.routePendingDamageWithReturn(model.TurnStageExtraAction) {
			return driveContinueLoop
		}
		if len(e.State.ActionQueue) > 0 {
			e.enterActionExecutionStage()
		} else {
			e.enterExtraActionStage()
		}
		return driveContinueLoop
	}

	// 获取目标（从 HandleAction 传入的 TargetID，需要存储）
	// 注意：这里我们需要从某个地方获取目标ID，可能需要修改 QueuedAction 结构
	// 暂时假设目标已经在某个地方存储了，或者从 ActionStack 中获取

	// 根据行动类型触发相应事件
	if currentAction.Type == model.ActionAttack {
		// 触发攻击开始事件
		targetID := currentAction.TargetID

		if targetID == "" {
			e.Log("[Error] 攻击行动缺少目标")
			return driveStop
		}

		target := e.State.Players[targetID]
		if target == nil {
			e.Log("[Error] 目标玩家不存在")
			return driveStop
		}

		// [新增] 先触发 TriggerOnCardUsed (封印等通用卡牌触发)
		if !e.State.ActionQueue[0].HasTriggeredCardUsed {
			// 技能转化攻击（如欺诈/多重射击）不消耗攻击牌，不触发 CardUsed。
			if currentAction.UsesVirtualCard {
				e.State.ActionQueue[0].HasTriggeredCardUsed = true
			} else {
				// 1. 使用队列中已准备好的卡牌副本（元素转化等已在排队/修复阶段处理）
				cardUsed := *currentAction.Card

				// 2. 触发 TriggerOnCardUsed
				cardCtx := &model.EventContext{
					Type:     model.EventCardUsed,
					Card:     &cardUsed,
					SourceID: currentPid,
					TargetID: targetID,
				}
				skillCtxUsed := e.buildContext(player, nil, model.TriggerOnCardUsed, cardCtx)
				e.dispatcher.OnTrigger(model.TriggerOnCardUsed, skillCtxUsed)

				// 标记已触发
				e.State.ActionQueue[0].HasTriggeredCardUsed = true

				// 3. 处理可能产生的延迟伤害 (即封印伤害)
				// 【五系封印伤害在此处结算】
				if e.processPendingDamages() {
					return driveStop // 有中断 (如伤害导致爆牌)，暂停 Drive
				}

				// 4. 处理可能产生的其他中断
				if e.State.PendingInterrupt != nil {
					return driveStop
				}
			}
		}

		e.recordAttackTargetLifecycle(player, targetID)

		eventCtx := &model.EventContext{
			Type:     model.EventAttack,
			SourceID: currentPid,
			TargetID: targetID,
			Card:     currentAction.Card,
			AttackInfo: &model.AttackEventInfo{
				IsHit:            false,
				CanBeResponded:   true,
				ActionType:       string(model.ActionAttack),
				CounterInitiator: "",
				InterceptTags:    map[model.CombatInterceptTag]bool{},
			},
		}

		// 仅在本条攻击尚未触发过 AttackStart 时触发（确认响应技能后会再次进入此处，不再重复触发）
		var attackStartCtx *model.Context
		if !e.State.ActionQueue[0].HasTriggeredAttackStart {
			e.resetAttackStartLifecycle(player)
			e.State.ActionQueue[0].HasTriggeredAttackStart = true
			attackStartCtx = e.buildContext(player, target, model.TriggerOnAttackStart, eventCtx)
			player.TurnState.LastActionType = string(model.ActionAttack)
			e.dispatcher.OnTrigger(model.TriggerOnAttackStart, attackStartCtx)
			if e.State.PendingInterrupt != nil {
				return driveStop
			}
			// 角色“攻击开始”阶段的中断由统一策略钩子处理，主流程不再直连角色技能实现。
			if e.runAttackStartInterruptHooks(player, target, &currentAction, attackStartCtx) {
				return driveStop
			}
		}

		// 无中断或已确认响应后：初始化战斗
		e.applyAttackPreCombatLifecycle(player, target, &currentAction, eventCtx)
		isForcedHit := eventCtx.AttackInfo != nil && eventCtx.AttackInfo.IsHitForced
		ignoreShield := eventCtx.AttackInfo != nil && eventCtx.AttackInfo.IgnoreShield

		// 消耗卡牌（从手牌中移除）
		card := *currentAction.Card
		if !currentAction.UsesVirtualCard {
			cardIdx := currentAction.CardIndex
			if _, err := consumePlayableCardByIndex(player, cardIdx); err != nil {
				e.Log("[Warn] PhaseBeforeAction: 卡牌索引失效，丢弃该行动")
				e.enterExtraActionStage()
				return driveContinueLoop
			}
			e.NotifyCardRevealed(currentPid, []model.Card{card}, "attack")
			e.State.DiscardPile = append(e.State.DiscardPile, card)
		}

		// 记录攻击行动次数
		player.TurnState.AttackCount += 1

		// 从队列中弹出行动（因为即将执行）
		e.State.ActionQueue = e.State.ActionQueue[1:]

		// 初始化战斗（使用实际卡牌，而不是队列中的指针）
		e.initCombat(currentPid, targetID, &card, isForcedHit, eventCtx.AttackInfo.CanBeResponded, ignoreShield, eventCtx.AttackInfo.InterceptTags)
		return driveContinueLoop
	}

	if currentAction.Type == model.ActionMagic {
		// 触发卡牌使用事件
		targetID := currentAction.TargetID
		if targetID == "" && len(currentAction.TargetIDs) > 0 {
			targetID = currentAction.TargetIDs[0]
		}

		if !e.State.ActionQueue[0].HasTriggeredCardUsed {
			cardCtx := &model.EventContext{
				Type:     model.EventCardUsed,
				Card:     currentAction.Card,
				SourceID: currentPid,
				TargetID: targetID,
			}

			skillCtx := e.buildContext(player, nil, model.TriggerOnCardUsed, cardCtx)

			// 触发卡牌使用事件
			e.dispatcher.OnTrigger(model.TriggerOnCardUsed, skillCtx)
			e.State.ActionQueue[0].HasTriggeredCardUsed = true

			// 如果触发了中断，等待用户输入
			if e.State.PendingInterrupt != nil {
				return driveStop
			}

			// 处理可能产生的延迟伤害（如封印），确保优先结算
			if e.processPendingDamages() {
				return driveStop
			}
			if e.State.PendingInterrupt != nil {
				return driveStop
			}
		}

		// 从队列中弹出行动
		e.State.ActionQueue = e.State.ActionQueue[1:]

		player.TurnState.LastActionType = string(model.ActionMagic)

		// 没有中断，执行法术逻辑
		// targetID 已经在上面计算过，包含了 TargetIDs[0] 的回退逻辑
		if err := e.PerformMagic(currentPid, targetID, currentAction.CardIndex); err != nil {
			e.Log(fmt.Sprintf("[Error] 法术执行失败: %v", err))
		}

		// 【新增检查】
		// 如果 PerformMagic 导致了中断 (比如触发了减伤技能)，
		// Phase 会被 ResolveDamage 改为 PhaseDamageResolution 或其他响应阶段。
		// 此时我们应该 break，让 Drive 处理中断，而不是强制跳到 ExtraAction
		if e.State.PendingInterrupt != nil {
			return driveContinueLoop
		}

		// 法术执行完毕，进入回合结束阶段
		if !e.routePendingDamageWithReturn(model.TurnStageTurnEnd) {
			e.enterTurnEndStage()
		}
		return driveContinueLoop
	}

	return driveContinueLoop
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
		// 通过玩家回合态消费“行动结束中断恢复标记”，避免依赖全局临时字段。
		skipActionEndInterruptHooks := e.consumeActionEndHookResuming(player)
		eventCtx := &model.EventContext{
			Type:       model.EventPhaseEnd,
			SourceID:   currentPid,
			ActionType: lastActionType, // 告诉技能，刚才结束的是 Attack
		}
		if eventCtx.ActionType == model.ActionAttack {
			eventCtx.AttackInfo = &model.AttackEventInfo{
				ActionType:       string(model.ActionAttack),
				CounterInitiator: "",
			}
		}

		skillCtx := e.buildContext(player, nil, model.TriggerOnPhaseEnd, eventCtx)

		// 行动结束时机的角色中断统一经由钩子注入（如圣剑第三击中断）。
		if !skipActionEndInterruptHooks && e.runActionEndInterruptHooks(skillCtx) {
			return driveStop
		}

		// 清除记录，防止死循环触发（非常重要！）
		player.TurnState.LastActionType = ""

		// 广播事件！
		// 此时 WindFuryHandler.CanUse 会被调用
		// 如果 CanUse 返回 true，Dispatcher 会根据 ResponseOptional 推送中断给用户
		e.dispatcher.OnTrigger(model.TriggerOnPhaseEnd, skillCtx)

		// 如果触发了技能（产生了中断，比如用户需要确认是否发动风怒），直接 return 等待用户
		if e.State.PendingInterrupt != nil {
			// OnPhaseEnd 已派发完，后续恢复时只补执行行动后续效果，不重复派发 OnPhaseEnd。
			e.markActionEndHookResuming(player)
			e.enqueuePostActionEndFollowup(player.ID, lastActionType)
			return driveStop
		}
		// 行动结束后场上赐福结算（如迅捷赐福）
		if e.handlePostActionEndEffects(player, lastActionType) {
			return driveStop
		}
	}

	e.enterExtraActionStage()
	return driveContinueLoop
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
