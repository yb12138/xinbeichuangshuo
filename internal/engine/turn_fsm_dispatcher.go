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
		return e.driveTurnBeforeStartStage(currentPid, player)
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
	if stage == "" {
		stage = model.TurnStageTurnBeforeStart
	}
	if e.State.Subflow != model.SubflowNone || len(e.State.CombatStack) > 0 {
		e.setTurnStage(stage)
		return stage
	}
	switch stage {
	case model.TurnStageActionExecution:
		return stage
	case model.TurnStageActionEnd, model.TurnStageExtraAction, model.TurnStageTurnEnd,
		model.TurnStageTurnBeforeStart, model.TurnStageBeforeAction, model.TurnStageTurnStart, model.TurnStageActionStart:
		e.setTurnStage(stage)
		return stage
	default:
		if player != nil && player.TurnState.LastActionType != "" {
			stage = model.TurnStageActionEnd
		} else if player != nil && player.Tokens != nil && player.Tokens["post_action_end_effect_pending"] > 0 {
			stage = model.TurnStageActionEnd
		} else if player != nil && player.TurnState.HasProcessedTurnStart {
			stage = model.TurnStageActionStart
		} else {
			stage = model.TurnStageTurnBeforeStart
		}
	}
	e.setTurnStage(stage)
	return stage
}

func (e *GameEngine) driveTurnBeforeStartStage(currentPid string, player *model.Player) driveOutcome {
	e.setTurnStage(model.TurnStageBeforeAction)
	return driveContinueLoop
}

func (e *GameEngine) driveBeforeActionStage(currentPid string, player *model.Player) driveOutcome {
	if e.State.Subflow != model.SubflowNone || len(e.State.CombatStack) > 0 || e.State.TurnStage != model.TurnStageBeforeAction {
		return driveUnhandled
	}

	e.setTurnStage(model.TurnStageBeforeAction)
	// 蝶舞者：凋零的“对方士气最低为1”持续到其下个回合开始前。
	if e.isButterflyDancer(player) {
		if player.Tokens == nil {
			player.Tokens = map[string]int{}
		}
		if player.Tokens["bt_wither_active"] > 0 {
			player.Tokens["bt_wither_active"] = 0
			e.Log(fmt.Sprintf("%s 的 [凋零] 效果到期：对方士气下限保护已解除", player.Name))
		}
	}
	// 血之巫女：流血形态回合开始先自损1点法术伤害（先于中毒/虚弱）。
	if e.isBloodPriestess(player) {
		if player.Tokens == nil {
			player.Tokens = map[string]int{}
		}
		if hasBloodPriestessBleedingForm(player) && player.Tokens["bp_bleed_tick_done_turn"] <= 0 {
			player.Tokens["bp_bleed_tick_done_turn"] = 1
			e.Log(fmt.Sprintf("%s 的 [流血] 生效：回合开始对自己造成1点法术伤害", player.Name))
			e.AddPendingDamage(model.PendingDamage{
				SourceID:   player.ID,
				TargetID:   player.ID,
				Damage:     1,
				DamageType: "magic",
			})
			e.enterDamageResolution(model.TurnStageBeforeAction)
			return driveContinueLoop
		}
	}
	// 行动开始前的基础效果先结算：
	// - 中毒先入延迟伤害管线，严格先于后续选择型状态；
	// - 虚弱保留在场上，交由后续选择中断处理。
	fieldCtx := e.buildContext(player, nil, model.TriggerOnBuffPhase, nil)
	e.triggerFieldEffects(player, model.EffectTriggerOnBeforeAction, fieldCtx)
	if e.State.PendingInterrupt != nil {
		return driveStop
	}
	if len(e.State.PendingDamageQueue) > 0 {
		e.enterDamageResolution(model.TurnStageBeforeAction)
		return driveContinueLoop
	}
	// 构建上下文
	skillCtx := e.buildContext(player, nil, model.TriggerOnBuffPhase, nil)

	// 选择型/事件型状态（如虚弱、五系束缚、元素封印）由 dispatcher 的通用 resolver 处理。
	e.dispatcher.OnTrigger(model.TriggerOnBuffPhase, skillCtx)
	// 处理可能产生的中断（例如虚弱需要玩家选择弃牌还是跳过）
	if e.State.PendingInterrupt != nil {
		return driveStop
	}

	if e.State.TurnStage == model.TurnStageTurnEnd {
		return driveStop
	}

	// 处理完 BuffResolve 后，检查是否有延迟伤害需要结算
	if len(e.State.PendingDamageQueue) > 0 {
		e.enterDamageResolution(model.TurnStageTurnStart)
	} else {
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

	e.setTurnStage(model.TurnStageTurnStart)
	player.TurnState.HasProcessedTurnStart = true
	if e.runPlayerPhaseHooks(player, turnStartPhaseHooks) {
		return driveStop
	}
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

	e.setTurnStage(model.TurnStageActionStart)
	if e.runPlayerPhaseHooks(player, actionStartPhaseHooks) {
		return driveStop
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
	// 3. 行动选择阶段
	if player.Tokens == nil {
		player.Tokens = map[string]int{}
	}
	// 勇者：精疲力竭在“下个行动阶段开始”时结束，立即转正并对自己造成3点法术伤害。
	if e.isHero(player) &&
		hasHeroExhaustionForm(player) &&
		player.Tokens["hero_exhaustion_release_pending"] > 0 {
		beforePoses := e.snapshotPlayerPoses()
		leaveHeroExhaustionForm(player)
		player.Tokens["hero_exhaustion_release_pending"] = 0
		e.Log(fmt.Sprintf("%s 的 [精疲力竭] 结束：转正，并对自己造成3点法术伤害", player.Name))
		e.dispatchOrientationChanges(beforePoses)
		e.InflictDamage(player.ID, player.ID, 3, "magic")
		if e.State.PendingInterrupt != nil || len(e.State.PendingDamageQueue) > 0 {
			return driveStop
		}
	}
	// 魔剑士暗影形态持续到“下个行动阶段开始前”，在这里统一退场。
	e.maybeReleaseMagicSwordsmanShadowAtActionStart(player)
	if player.Tokens["judgment"] >= 4 &&
		player.Tokens["arbiter_skip_forced_doomsday"] == 0 &&
		player.Tokens["arbiter_forced_doomsday_done_turn"] == 0 {
		targetIDs := e.campEnemyIDs(player.Camp)
		if len(targetIDs) > 0 {
			e.PushInterrupt(&model.Interrupt{
				Type:     model.InterruptChoice,
				PlayerID: currentPid,
				Context: map[string]interface{}{
					"choice_type":   "arbiter_forced_doomsday_target",
					"user_id":       currentPid,
					"target_ids":    targetIDs,
					"waiting_phase": model.NormalizeResumePoint(model.TurnStageActionExecution),
				},
			})
			return driveStop
		}
	}

	tauntSourceID := ""
	tauntSrcName := ""
	if tauntCard := getHeroTauntCard(player); tauntCard != nil {
		src := e.State.Players[tauntCard.SourceID]
		// 仅对“敌方来源的挑衅”生效；非法残留直接移除。
		if src == nil || src.Camp == player.Camp {
			e.RemoveFieldCard(player.ID, model.EffectHeroTaunt)
		} else {
			tauntSourceID = src.ID
			tauntSrcName = model.GetPlayerDisplayName(src)
			hasAttackCard := false
			for idx := 0; idx < playableCardCount(player); idx++ {
				c, _, _, ok := getPlayableCardByIndex(player, idx)
				if ok && c.Type == model.CardTypeAttack {
					hasAttackCard = true
					break
				}
			}
			if !hasAttackCard {
				e.Log(fmt.Sprintf("[Taunt] %s 受到【挑衅】约束但无攻击牌，跳过本次行动阶段", player.Name))
				e.RemoveFieldCard(player.ID, model.EffectHeroTaunt)
				e.enterTurnEndStage()
				return driveContinueLoop
			}
		}
	}
	hasHeroTaunt := tauntSourceID != ""

	var validOptions []model.PromptOption
	var specialOptions []model.PromptOption
	currentExtraAction := player.TurnState.CurrentExtraAction
	isRestrictedExtraAction := currentExtraAction == "Attack" || currentExtraAction == "Magic"
	canMagicAction := e.canCastMagicInAction(player)
	canMagicSkillAction := e.hasUsableActionSkillForExtraMagic(player)
	hasRestrictedExtraActionCard := true
	if isRestrictedExtraAction {
		hasRestrictedExtraActionCard = e.checkExtraActionCards(player, currentExtraAction, player.TurnState.CurrentExtraElement)
	}
	hasFighterHundredDragon := e.isFighter(player) && hasFighterHundredDragonForm(player)

	// 行动类型选项：
	// - 额外攻击行动：只能选攻击
	// - 额外法术行动：只能选法术
	// - 常规行动：攻击/法术都可选
	switch currentExtraAction {
	case "Attack":
		if hasRestrictedExtraActionCard {
			validOptions = append(validOptions, model.PromptOption{ID: "attack", Label: "攻击"})
		}
	case "Magic":
		// 额外法术行动：允许“法术牌”或“主动技能(视为法术行动)”。
		if hasRestrictedExtraActionCard {
			validOptions = append(validOptions, model.PromptOption{ID: "magic", Label: "法术"})
		}
	default:
		if hasFighterHundredDragon {
			validOptions = append(validOptions, model.PromptOption{ID: "attack", Label: "攻击（百式幻龙拳）"})
		} else if hasHeroTaunt {
			validOptions = append(validOptions, model.PromptOption{ID: "attack", Label: "攻击（受挑衅约束）"})
		} else {
			validOptions = append(validOptions, model.PromptOption{ID: "attack", Label: "攻击"})
			// 常规阶段：即便当前形态不能“打出法术牌”，只要存在可用主动技能，
			// 仍应保留“法术行动”入口（例如魔剑士【暗影流星】）。
			if canMagicAction || canMagicSkillAction {
				validOptions = append(validOptions, model.PromptOption{ID: "magic", Label: "法术"})
			}
		}
	}

	if !hasHeroTaunt && !hasFighterHundredDragon && !isRestrictedExtraAction && !e.State.HasPerformedStartup {
		// 未执行启动技能时，按条件过滤特殊行动
		maxHand := e.GetMaxHand(player)
		canBuyOrSynth := len(player.Hand)+3 <= maxHand

		if canBuyOrSynth {
			specialOptions = append(specialOptions, model.PromptOption{ID: "buy", Label: "购买"})
		}

		var totalStones int
		if player.Camp == model.RedCamp {
			totalStones = e.State.RedGems + e.State.RedCrystals
		} else {
			totalStones = e.State.BlueGems + e.State.BlueCrystals
		}
		if canBuyOrSynth && totalStones >= 3 {
			specialOptions = append(specialOptions, model.PromptOption{ID: "synthesize", Label: "合成"})
		}

		currentEnergy := player.Gem + player.Crystal
		if totalStones > 0 && currentEnergy < 3 {
			specialOptions = append(specialOptions, model.PromptOption{ID: "extract", Label: "提炼"})
		}

		if len(specialOptions) > 0 {
			validOptions = append(validOptions, model.PromptOption{ID: "special", Label: "特殊"})
		}
	}

	// 常规行动下："无法行动"表示展示手牌并重摸；
	// 额外行动受限下：当无合法动作时，允许主动宣告跳过本次额外行动。
	if !isRestrictedExtraAction {
		hasAttackCard := false
		hasMagicCard := false
		for idx := 0; idx < playableCardCount(player); idx++ {
			c, _, _, ok := getPlayableCardByIndex(player, idx)
			if !ok {
				continue
			}
			if c.Type == model.CardTypeAttack {
				hasAttackCard = true
			}
			if c.Type == model.CardTypeMagic && canMagicAction {
				hasMagicCard = true
			}
		}
		canNormalAction := hasAttackCard || hasMagicCard || canMagicSkillAction
		if hasFighterHundredDragon {
			canNormalAction = hasAttackCard
		}
		// 仅当无法执行一般行动（无攻击牌也无法术牌）时提供"无法行动"
		if !canNormalAction {
			validOptions = append(validOptions, model.PromptOption{ID: "cannot_act", Label: "无法行动（展示手牌）"})
		}
	} else if !hasRestrictedExtraActionCard {
		validOptions = append(validOptions, model.PromptOption{ID: "cannot_act", Label: "跳过额外行动"})
	}

	promptMessage := fmt.Sprint("请选择行动类型")
	if currentExtraAction == "Attack" {
		promptMessage = fmt.Sprint("当前为额外攻击行动，仅可执行攻击。请选择行动类型")
	} else if currentExtraAction == "Magic" {
		promptMessage = fmt.Sprint("当前为额外法术行动，仅可执行法术。请选择行动类型")
	} else if hasFighterHundredDragon {
		if locked := e.fighterLockedTarget(player); locked != nil {
			promptMessage = fmt.Sprintf("你处于【百式幻龙拳】状态：本行动阶段只能主动攻击 %s；若本行动阶段结束仍处于该形态，则自动转正。", model.GetPlayerDisplayName(locked))
		} else {
			promptMessage = fmt.Sprint("你处于【百式幻龙拳】状态：本行动阶段只能主动攻击已锁定目标；若本行动阶段结束仍处于该形态，则自动转正。")
		}
	} else if hasHeroTaunt {
		promptMessage = fmt.Sprintf("你受到【挑衅】影响：本次行动阶段必须且只能主动攻击 %s。", tauntSrcName)
	}
	if isRestrictedExtraAction && !hasRestrictedExtraActionCard {
		promptMessage = fmt.Sprint("当前为额外行动阶段，但你没有满足约束的可执行动作。可选择跳过本次额外行动。")
	}

	prompt := &model.Prompt{
		Type:           model.PromptConfirm,
		PlayerID:       currentPid,
		Message:        promptMessage,
		Options:        validOptions,
		SpecialOptions: specialOptions,
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
		if len(e.State.PendingDamageQueue) > 0 {
			e.enterDamageResolution(model.TurnStageExtraAction)
		} else if len(e.State.ActionQueue) > 0 {
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
		if len(e.State.PendingDamageQueue) > 0 {
			e.enterDamageResolution(model.TurnStageExtraAction)
		} else {
			e.enterExtraActionStage()
		}
		return driveContinueLoop
	}

	currentAction := e.State.ActionQueue[0] // 只读取，不弹出
	if !queuedActionUsesVirtualCard(currentAction.SourceSkill) {
		if !e.repairQueuedActionCard(player, &e.State.ActionQueue[0]) {
			e.Log("[Warn] PhaseBeforeAction: 无法修复队列中的卡牌索引，丢弃该行动")
			e.State.ActionQueue = e.State.ActionQueue[1:]
			if len(e.State.PendingDamageQueue) > 0 {
				e.enterDamageResolution(model.TurnStageExtraAction)
			} else if len(e.State.ActionQueue) > 0 {
				e.enterActionExecutionStage()
			} else {
				e.enterExtraActionStage()
			}
			return driveContinueLoop
		}
		currentAction = e.State.ActionQueue[0]
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

		// ==========================================
		// 五系封印触发点 1/3：攻击前打出卡牌时触发
		// ==========================================
		// [新增] 先触发 TriggerOnCardUsed (封印等通用卡牌触发)
		if !e.State.ActionQueue[0].HasTriggeredCardUsed {
			// 技能转化攻击（如欺诈/多重射击）不消耗攻击牌，不触发 CardUsed。
			if queuedActionUsesVirtualCard(currentAction.SourceSkill) {
				e.State.ActionQueue[0].HasTriggeredCardUsed = true
			} else {
				// 1. 获取使用的卡牌 (用于事件触发)
				cardIdx := currentAction.CardIndex
				cardUsed, _, _, ok := getPlayableCardByIndex(player, cardIdx)
				if !ok {
					e.Log("[Warn] PhaseBeforeAction: 卡牌索引失效，丢弃该行动")
					e.State.ActionQueue = e.State.ActionQueue[1:]
					e.enterExtraActionStage()
					return driveContinueLoop
				}
				// 此时还未消耗，获取副本
				cardUsed = e.applyBlazeWitchAttackCardTransform(player, cardUsed)

				// 2. 触发 TriggerOnCardUsed 【五系封印在此处触发】
				// 如果玩家面前有对应元素的封印，这里会触发并添加PendingDamage
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

		e.recordAttackTargetContext(player, targetID)

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
			},
		}

		// 仅在本条攻击尚未触发过 AttackStart 时触发（确认响应技能后会再次进入此处，不再重复触发）
		var attackStartCtx *model.Context
		if !e.State.ActionQueue[0].HasTriggeredAttackStart {
			e.runAttackStartStateResets(player)
			e.State.ActionQueue[0].HasTriggeredAttackStart = true
			attackStartCtx = e.buildContext(player, target, model.TriggerOnAttackStart, eventCtx)
			player.TurnState.LastActionType = string(model.ActionAttack)
			e.dispatcher.OnTrigger(model.TriggerOnAttackStart, attackStartCtx)
			if e.State.PendingInterrupt != nil {
				return driveStop
			}
			if e.maybeTriggerMoonGoddessMedusa(player, target, currentAction.SourceSkill, currentAction.Card, attackStartCtx) {
				return driveStop
			}
		}

		// 无中断或已确认响应后：初始化战斗
		e.applyAttackPreCombatRoleRules(player, target, &currentAction, eventCtx)
		isForcedHit := eventCtx.AttackInfo != nil && eventCtx.AttackInfo.IsHitForced
		ignoreShield := eventCtx.AttackInfo != nil && eventCtx.AttackInfo.IgnoreShield

		// 消耗卡牌（从手牌中移除）
		card := *currentAction.Card
		if !queuedActionUsesVirtualCard(currentAction.SourceSkill) {
			cardIdx := currentAction.CardIndex
			usedCard, err := consumePlayableCardByIndex(player, cardIdx)
			if err != nil {
				e.Log("[Warn] PhaseBeforeAction: 卡牌索引失效，丢弃该行动")
				e.enterExtraActionStage()
				return driveContinueLoop
			}
			card = usedCard
			card = e.applyBlazeWitchAttackCardTransform(player, card)
			e.NotifyCardRevealed(currentPid, []model.Card{card}, "attack")
			e.State.DiscardPile = append(e.State.DiscardPile, card)
		}

		// 记录攻击行动次数
		player.TurnState.AttackCount += 1

		// 从队列中弹出行动（因为即将执行）
		e.State.ActionQueue = e.State.ActionQueue[1:]

		// 初始化战斗（使用实际卡牌，而不是队列中的指针）
		e.initCombat(currentPid, targetID, &card, isForcedHit, eventCtx.AttackInfo.CanBeResponded, ignoreShield)
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
		if len(e.State.PendingDamageQueue) > 0 {
			e.enterDamageResolution(model.TurnStageTurnEnd)
		} else {
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

	// 阴阳师式神咒束：在响应阶段开始前，先检查是否触发“代应战”。
	if e.tryStartOnmyojiBindingInterrupt(combatReq) {
		return driveStop
	}
	// 若已完成式神咒束选择，自动执行“视为应战”流程。
	if e.executeOnmyojiBindingCounter(combatReq) {
		return driveStop
	}
	// 阴阳师阴阳转换：若目标阴阳师存在“同命格应战”机会，先询问是否发动。
	if e.tryStartOnmyojiYinYangInterrupt(combatReq) {
		return driveStop
	}

	// 如果强制命中，直接结算伤害
	if combatReq.IsForcedHit {
		e.Log(fmt.Sprintf("[Combat] 攻击强制命中！跳过响应阶段，直接结算..."))
		if atk := e.State.Players[combatReq.AttackerID]; atk != nil && atk.Tokens != nil {
			atk.Tokens["elf_elemental_shot_thunder_pending"] = 0
		}
		e.clearCombatStack()
		e.AddPendingDamageFront(model.PendingDamage{
			SourceID:     combatReq.AttackerID,
			TargetID:     combatReq.TargetID,
			Damage:       combatReq.Card.Damage,
			DamageType:   "Attack",
			Card:         combatReq.Card,
			IsCounter:    combatReq.IsCounter,
			IgnoreShield: combatReq.IgnoreShield,
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
	// 暗灭规则兜底：默认暗灭攻击不可应战。
	// 例外：魔剑士【暗影抗拒】在“非自己行动阶段”可用【魔弹】响应，应保留应战入口。
	if combatReq.Card != nil && combatReq.Card.Element == model.ElementDark && !e.canUseShadowRejectResponseMagic(target) {
		combatReq.CanBeResponded = false
	}
	noHolyDefend := attacker != nil && attacker.Tokens != nil && attacker.Tokens["bs_no_holy_defend_current_attack"] > 0
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
	// 若上一次行动在 OnPhaseEnd 触发中断后返回，这里补做“行动结束追加效果”
	// （如迅捷赐福），避免被前置中断吞掉。
	if player.TurnState.LastActionType == "" && player.Tokens != nil && player.Tokens["post_action_end_effect_pending"] > 0 {
		actionType := model.ActionAttack
		if player.Tokens["post_action_end_effect_magic"] > 0 {
			actionType = model.ActionMagic
		}
		player.Tokens["post_action_end_effect_pending"] = 0
		player.Tokens["post_action_end_effect_magic"] = 0
		if e.handlePostActionEndEffects(player, actionType) {
			return driveStop
		}
	}

	if player.TurnState.LastActionType != "" {
		lastActionType := model.ActionType(player.TurnState.LastActionType)
		skipHolySwordPhaseEnd := false
		if player.Tokens != nil && player.Tokens["holy_sword_phase_end_pending"] > 0 {
			skipHolySwordPhaseEnd = true
			player.Tokens["holy_sword_phase_end_pending"] = 0
		}
		specialPhaseEndDispatched := false
		if player.Tokens != nil &&
			player.Tokens["special_phase_end_dispatched"] > 0 &&
			(lastActionType == model.ActionBuy || lastActionType == model.ActionSynthesize || lastActionType == model.ActionExtract) {
			specialPhaseEndDispatched = true
			player.Tokens["special_phase_end_dispatched"] = 0
		}
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

		if !skipHolySwordPhaseEnd && e.maybeTriggerHolySwordDrawFromPhaseEndCtx(skillCtx) {
			return driveStop
		}

		// 清除记录，防止死循环触发（非常重要！）
		player.TurnState.LastActionType = ""

		// 广播事件！
		// 此时 WindFuryHandler.CanUse 会被调用
		// 如果 CanUse 返回 true，Dispatcher 会根据 ResponseOptional 推送中断给用户
		// 特殊行动(Buy/Synthesize/Extract)在 ActionSelection 已完成过一次 OnPhaseEnd，
		// 这里跳过重复触发，避免被动结算两次。
		if !specialPhaseEndDispatched {
			e.dispatcher.OnTrigger(model.TriggerOnPhaseEnd, skillCtx)
		}

		// 如果触发了技能（产生了中断，比如用户需要确认是否发动风怒），直接 return 等待用户
		if e.State.PendingInterrupt != nil {
			if player.Tokens == nil {
				player.Tokens = map[string]int{}
			}
			player.Tokens["post_action_end_effect_pending"] = 1
			if lastActionType == model.ActionMagic {
				player.Tokens["post_action_end_effect_magic"] = 1
			} else {
				player.Tokens["post_action_end_effect_magic"] = 0
			}
			// 【重要】恢复 LastActionType，因为中断回来后还要处理 PhaseExtraAction
			// 但为了避免重复触发 EventPhaseEnd，我们需要一个标志位，或者让中断处理完直接进队列检查
			// 简单做法：中断回来后，Phase 依然是 ExtraAction，但 LastActionType 已被清空，所以不会二次触发
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

	e.setTurnStage(model.TurnStageTurnEnd)
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
