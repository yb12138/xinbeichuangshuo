// gameflow: 行动前阶段：可选行动列表、虚拟攻击牌等。

// before_action_phase_runtime 实现行动阶段（TurnStageActionExecution）里「执行队首待办行动之前」的规则链。
// 玩家已在行动枢纽选定攻击或法术后，ActionQueue 队首会在此依次经过：使用牌时点（含封印等延迟伤害）、
// 攻击宣言与应战入口（initCombat），或法术结算（PerformMagic）；对应设计文档中 ActionExecution 且队列非空时的子流程。
package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

// beforeActionPeekReadyHead 确保队首 QueuedAction 可执行：对齐手牌索引、非虚拟牌须有 Card。
// 若队列异常（空、牌找不到、数据坏档）则丢弃本条并回到额外行动或下一条队列；此时 stop=true，调用方直接 return Drive。
func (e *GameEngine) beforeActionPeekReadyHead(player *model.Player) (head *model.QueuedAction, immediate driveOutcome, stop bool) {
	if len(e.State.ActionQueue) == 0 {
		// 窗口上应为「有待执行行动」；空队列时直接回到额外行动链或先清延迟伤害，避免卡在行动执行态。
		e.Log("[Warn] PhaseBeforeAction: 行动队列为空，执行阶段修复")
		if e.routePendingDamageWithDefaultReturn(model.TurnStageExtraAction) {
			return nil, driveContinueLoop, true
		}
		e.enterExtraActionStage()
		return nil, driveContinueLoop, true
	}

	head = &e.State.ActionQueue[0]
	if !head.UsesVirtualCard {
		if !e.repairQueuedActionCard(player, head) {
			e.Log("[Warn] PhaseBeforeAction: 无法修复队列中的卡牌索引，丢弃该行动")
			e.State.ActionQueue = e.State.ActionQueue[1:]
			return nil, e.beforeActionRecoverAfterDroppedHead(), true
		}
		head = &e.State.ActionQueue[0]
	}
	if head.Card == nil {
		e.Log("[Warn] PhaseBeforeAction: 队列中的卡牌数据缺失，丢弃该行动")
		e.State.ActionQueue = e.State.ActionQueue[1:]
		return nil, e.beforeActionRecoverAfterDroppedHead(), true
	}
	return head, 0, false
}

// beforeActionRecoverAfterDroppedHead 丢弃非法队首后的去向：优先结算可能挂起的延迟伤害，再决定继续本条行动阶段或进入额外行动。
func (e *GameEngine) beforeActionRecoverAfterDroppedHead() driveOutcome {
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

// beforeActionRunCardUsedIfNeeded 触发技能与状态上的「打出/展示卡牌」时点（TimingOnCardPlayedOrRevealed），例如五系封印等会在此后插入延迟伤害并优先结算。
// virtualSkipCardDispatch：技能视为的攻击（欺诈、多重射击等）不从手牌打出实体攻击牌，规则上不走「使用那张手牌」的触发链，只记标记以免重复。
func (e *GameEngine) beforeActionRunCardUsedIfNeeded(player *model.Player, currentPid, targetID string, head *model.QueuedAction, cardForEvent *model.Card, virtualSkipCardDispatch bool) (immediate driveOutcome, stop bool) {
	if head.HasDispatchedCardUsed {
		return 0, false
	}
	if virtualSkipCardDispatch {
		head.HasDispatchedCardUsed = true
		return 0, false
	}
	if cardForEvent == nil {
		return 0, false
	}
	cardCtx := &model.EventContext{
		Type:     model.EventCardUsed,
		Card:     cardForEvent,
		SourceID: currentPid,
		TargetID: targetID,
	}
	skillCtx := e.BuildContext(player, nil, model.TimingOnCardPlayedOrRevealed, cardCtx)
	e.dispatcher.OnTiming(skillCtx.Timing, skillCtx)
	head.HasDispatchedCardUsed = true
	if e.State.PendingInterrupt != nil {
		return driveStop, true
	}
	if e.ProcessPendingDamages() {
		return driveStop, true
	}
	if e.State.PendingInterrupt != nil {
		return driveStop, true
	}
	return 0, false
}

// driveBeforeActionAttack 主动攻击链：宣言攻击（AttackStart）→ 若有被动/确认则中断 → 从手牌移除攻击牌（非虚拟）→ 压入战斗栈，由后续阶段处理目标「承受/防御/应战」。
// HasDispatchedAttackDeclared 用于在「响应技能确认」后再次进入本段时不重复宣言。
func (e *GameEngine) driveBeforeActionAttack(currentPid string, player *model.Player, head *model.QueuedAction) driveOutcome {
	targetID := head.TargetID
	if targetID == "" {
		e.Log("[Error] 攻击行动缺少目标")
		return driveStop
	}
	target := e.State.Players[targetID]
	if target == nil {
		e.Log("[Error] 目标玩家不存在")
		return driveStop
	}

	var cardForUsed *model.Card
	if !head.UsesVirtualCard {
		c := *head.Card
		// 先变换卡牌（如烈焰魔女火焰形态），再触发 TimingOnCardPlayedOrRevealed
		c = e.transformAttackCard(player, c)
		cardForUsed = &c
	}
	if out, stop := e.beforeActionRunCardUsedIfNeeded(player, currentPid, targetID, head, cardForUsed, head.UsesVirtualCard); stop {
		return out
	}

	e.recordAttackTargetLifecycle(player, targetID)

	// 复用已保存的事件上下文（响应技能中断后恢复时，技能可能已修改 AttackInfo）
	var eventCtx *model.EventContext
	if head.HasDispatchedAttackDeclared && head.SavedAttackEventCtx != nil {
		eventCtx = head.SavedAttackEventCtx
	} else {
		eventCtx = &model.EventContext{
			Type:     model.EventAttack,
			SourceID: currentPid,
			TargetID: targetID,
			Card:     head.Card,
			AttackInfo: &model.AttackEventInfo{
				IsHit:            false,
				CanBeResponded:   true,
				ActionType:       string(model.ActionAttack),
				CounterInitiator: "",
				InterceptTags:    map[model.CombatInterceptTag]bool{},
			},
		}
	}

	if !head.HasDispatchedAttackDeclared {
		e.resetAttackStartLifecycle(player)
		head.HasDispatchedAttackDeclared = true
		attackStartCtx := e.BuildContext(player, target, model.TimingOnAttackDeclared, eventCtx)
		player.TurnState.LastActionType = string(model.ActionAttack)
		cardSnapshot := *head.Card
		player.TurnState.LastActionCard = &cardSnapshot
		e.dispatcher.OnTiming(attackStartCtx.Timing, attackStartCtx)
		// 保存事件上下文，响应技能中断后复用（技能可能修改了 AttackInfo）
		head.SavedAttackEventCtx = eventCtx
		if e.State.PendingInterrupt != nil {
			return driveStop
		}
		if e.runTimingOnAttackDeclaredInterruptPolicies(player, target, head, attackStartCtx) {
			return driveStop
		}
	}

	e.applyAttackPreCombatLifecycle(player, target, head, eventCtx)
	isForcedHit := eventCtx.AttackInfo != nil && eventCtx.AttackInfo.IsHitForced
	ignoreShield := eventCtx.AttackInfo != nil && eventCtx.AttackInfo.IgnoreShield

	card := *head.Card
	if !head.UsesVirtualCard {
		if _, err := e.consumeQueuedActionCard(player, head); err != nil {
			e.Log("[Warn] PhaseBeforeAction: 卡牌ID失效，丢弃该行动")
			e.enterExtraActionStage()
			return driveContinueLoop
		}
		e.NotifyCardRevealed(currentPid, []model.Card{card}, "attack")
		e.State.DiscardPile = append(e.State.DiscardPile, card)
	}

	player.TurnState.AttackCount++
	e.State.ActionQueue = e.State.ActionQueue[1:]
	// initCombat：进入战斗交互（应战窗口）或强制命中时直接转入伤害队列；与桌游中「攻击已打出、等待对方响应」一致。
	e.initCombat(currentPid, targetID, &card, isForcedHit, eventCtx.AttackInfo.CanBeResponded, ignoreShield, eventCtx.AttackInfo.InterceptTags, eventCtx.AttackInfo.ElementOverride)
	return driveContinueLoop
}

// driveBeforeActionMagic 法术牌：在使用牌时点与延迟伤害处理完后，执行 PerformMagic（效果/伤害可能在流程中再挂中断）。
// 本引擎规则：一次「队列中的法术」执行结束后倾向进入 TurnEnd（或先清延迟伤害再 TurnEnd），与主动攻击进入战斗栈的路径不同。
func (e *GameEngine) driveBeforeActionMagic(currentPid string, player *model.Player, head *model.QueuedAction) driveOutcome {
	targetID := head.TargetID
	if targetID == "" && len(head.TargetIDs) > 0 {
		targetID = head.TargetIDs[0]
	}

	if out, stop := e.beforeActionRunCardUsedIfNeeded(player, currentPid, targetID, head, head.Card, false); stop {
		return out
	}

	e.State.ActionQueue = e.State.ActionQueue[1:]
	player.TurnState.LastActionType = string(model.ActionMagic)
	if head.Card != nil {
		cardSnapshot := *head.Card
		player.TurnState.LastActionCard = &cardSnapshot
	} else {
		player.TurnState.LastActionCard = nil
	}

	cardID := queuedActionCardID(head)
	if err := e.PerformMagicByID(currentPid, targetID, cardID); err != nil {
		e.Log(fmt.Sprintf("[Error] 法术执行失败: %v", err))
	}
	if e.State.PendingInterrupt != nil {
		return driveContinueLoop
	}
	if !e.routePendingDamageWithReturn(model.TurnStageActionEnd) {
		e.enterActionEndStage()
	}
	return driveContinueLoop
}
