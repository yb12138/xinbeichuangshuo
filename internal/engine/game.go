package engine

import (
	"errors"
	"fmt"
	"math/rand"
	"starcup-engine/internal/data"
	"starcup-engine/internal/engine/skills"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
	"strings"
	"time"
)

type GameEngine struct {
	State      *model.GameState
	dispatcher *SkillDispatcher
	observer   model.GameObserver // [新增] 持有观察者
	lifecycle  AttackLifecycle
	// 记录“当前回合内各角色已对哪些敌方角色造成过法术伤害”。
	turnMagicDamageTargets map[string]map[string]bool
	actionSummary          *actionSummary
	actionSummaryTurn      int
	suppressSealOnDiscard  bool
}

func NewGameEngine(observer model.GameObserver) *GameEngine {
	skills.InitHandlers()
	engine := &GameEngine{
		State:                  model.NewGameState(),
		observer:               observer,
		turnMagicDamageTargets: map[string]map[string]bool{},
		actionSummaryTurn:      0,
	}
	engine.lifecycle = NewAttackLifecycle(engine)
	engine.dispatcher = NewSkillDispatcher(engine)
	return engine
}

// AddPlayer 加入一名玩家。role 为角色配置 ID（与房间 CharRole、前端 char_role 一致），仅按 data.Character.ID 绑定，不按展示名匹配。
func (e *GameEngine) AddPlayer(id, name, role string, camp model.Camp) error {
	if len(e.State.Players) >= 6 {
		return errors.New("游戏人数已满 (6人)")
	}
	if _, exists := e.State.Players[id]; exists {
		return errors.New("玩家ID已存在")
	}

	player := &model.Player{
		ID:             id,
		Name:           name,
		Camp:           camp,
		Role:           role,
		Hand:           make([]model.Card, 0),
		Blessings:      make([]model.Card, 0),
		ExclusiveCards: make([]model.Card, 0),
		MaxHand:        6, // 初始手牌上限
		Heal:           0,
		MaxHeal:        2,
		IsActive:       false,
		Tokens:         map[string]int{},
		Orientation:    model.OrientationNormal,
		TurnState:      model.NewPlayerTurnState(),
	}

	// 查找并绑定角色数据
	characters := data.GetCharacters()
	for _, c := range characters {
		if c.ID == role {
			charCopy := c // Copy struct
			player.Character = &charCopy
			player.MaxHand = c.MaxHand
			break
		}
	}
	if player.Character == nil {
		e.Log(fmt.Sprintf("Warning: Character not found for character id %s", role))
	}
	e.runPlayerBootstrapHooks(player, addPlayerBootstrapHooks)

	e.State.Players[id] = player
	e.State.PlayerOrder = append(e.State.PlayerOrder, id)
	e.refreshPlayerDerivedState(player)
	return nil
}

// buildMagicMissilePrompt 构建魔弹响应提示
// StartGame 开始游戏
func (e *GameEngine) StartGame() error {
	if len(e.State.Players) < 2 {
		return errors.New("玩家人数不足")
	}
	const initialHandSize = 4

	// 1. 初始化牌库
	e.State.Deck = rules.InitDeck()
	e.State.Deck = rules.Shuffle(e.State.Deck)

	// 2. 发初始手牌
	for _, pid := range e.State.PlayerOrder {
		player := e.State.Players[pid]
		cards, newDeck, _ := rules.DrawCards(e.State.Deck, e.State.DiscardPile, initialHandSize)
		player.Hand = append(player.Hand, cards...)
		e.State.Deck = newDeck
		e.runPlayerBootstrapHooks(player, gameStartBootstrapHooks)
	}

	// 3. 随机决定先手
	rand.Seed(time.Now().UnixNano())
	startIndex := rand.Intn(len(e.State.PlayerOrder))
	firstPlayerID := e.State.PlayerOrder[startIndex]
	e.Log(fmt.Sprintf("[Game] 游戏开始! 首发玩家: %s (%s)",
		e.State.Players[firstPlayerID].Name,
		e.State.Players[firstPlayerID].Camp))

	e.State.CurrentTurn = startIndex

	player := e.State.Players[firstPlayerID]
	player.IsActive = true
	player.TurnState = model.NewPlayerTurnState()
	e.actionSummaryTurn = 1

	e.State.TurnStage = model.TurnStageTurnBeforeStart
	e.State.CombatStage = model.CombatStageNone
	e.State.Subflow = model.SubflowNone
	e.resetTurnMagicDamageTracker()

	// 进入第一回合
	e.Drive()

	return nil
}

// Drive 状态机驱动函数，自动在阶段间转换或等待用户输入
func (e *GameEngine) Drive() {
	iterations := 0
	resolveDriveOutcome := func(outcome driveOutcome) (handled bool, shouldStop bool) {
		if outcome == driveUnhandled {
			return false, false
		}
		return true, outcome == driveStop
	}
	for {
		e.Log(fmt.Sprintf("[Debug] Drive Loop: %d, %s", iterations, e.runtimeStateLabel()))
		iterations++
		// 如果有待处理的中断，不自动推进
		if e.State.PendingInterrupt != nil {
			return
		}
		// 仅在没有待处理延迟伤害时推进“延迟后续”。
		// 这样可保证诸如“封印伤害先结算，再继续技能后续”的严格顺序。
		if !e.isDamageResolutionActive() &&
			len(e.State.PendingDamageQueue) == 0 &&
			len(e.State.DeferredFollowups) > 0 {
			e.processDeferredFollowups()
			if e.State.PendingInterrupt != nil {
				return
			}
			continue
		}

		// 行动收尾：先跑行动结束后的全局 hook，再输出汇总信息。
		if e.runActionFinalizeHooksIfIdle() {
			if e.State.PendingInterrupt != nil {
				return
			}
			if !e.isActionFinalizeIdle() {
				continue
			}
		}

		// 行动汇总：当系统回到可继续行动的空闲状态时输出汇总信息
		e.finalizeActionSummaryIfIdle()

		currentPid := e.State.PlayerOrder[e.State.CurrentTurn]
		player := e.State.Players[currentPid]
		if handled, shouldStop := resolveDriveOutcome(e.driveNonTurnPhase(currentPid, player)); handled {
			if shouldStop {
				return
			}
			continue
		}
		if handled, shouldStop := resolveDriveOutcome(e.driveTurnFSM(currentPid, player)); handled {
			if shouldStop {
				return
			}
			continue
		}
		return
	}
}

// checkGameEnd 检查游戏是否结束
func (e *GameEngine) checkGameEnd() {
	// 星杯胜利：任一方星杯达到 5
	if e.State.RedCups >= 5 {
		e.Notify(model.EventGameEnd, "红方胜利！星杯达到 5", nil)
		e.setGameOver(true)
		return
	}
	if e.State.BlueCups >= 5 {
		e.Notify(model.EventGameEnd, "蓝方胜利！星杯达到 5", nil)
		e.setGameOver(true)
		return
	}
	// 检查是否有玩家的士气归零
	for _, player := range e.State.Players {
		if player.Camp == model.RedCamp && e.State.RedMorale <= 0 {
			e.Notify(model.EventGameEnd, "蓝方胜利！红方士气归零", nil)
			e.setGameOver(true)
			return
		}
		if player.Camp == model.BlueCamp && e.State.BlueMorale <= 0 {
			e.Notify(model.EventGameEnd, "红方胜利！蓝方士气归零", nil)
			e.setGameOver(true)
			return
		}
	}
}

// HandleAction 核心路由器：处理所有 Action
func (e *GameEngine) HandleAction(act model.PlayerAction) error {
	e.Log(fmt.Sprintf("[Debug] HandleAction 收到指令: %s", act.Type))
	// === 1. 第一优先级：系统指令 (随时可执行) ===
	// 允许玩家在任何时候退出或查看帮助，哪怕是在选择弃牌的时候
	if act.Type == model.CmdQuit {
		// e.Notify(model.EventGameEnd, "玩家强制退出", nil)
		return fmt.Errorf("EXIT_GAME") // 或者特定的退出逻辑
	}
	if act.Type == model.CmdHelp {
		// 帮助信息通常由 CLI 直接处理，Engine 也可以返回特定的提示
		return nil
	}

	// 作弊指令 (Debug用)
	if act.Type == model.CmdCheat {
		if err := e.handleCheat(act); err != nil {
			return err
		}
		// 作弊成功后也驱动一次状态机，让回合和提示立即更新
		if e.State.PendingInterrupt == nil {
			e.Drive()
		}
		return nil
	}

	// === 2. 第二优先级：中断处理 (Interrupt) ===
	// 如果当前有挂起的中断，**必须** 先处理中断，禁止执行其他普通指令
	if e.State.PendingInterrupt != nil {
		// 处理中断输入
		err := e.handleInterruptAction(act)
		if err != nil {
			return err // 处理失败（如输入非法），直接返回错误，不驱动引擎
		}

		// 【关键】中断处理成功后，驱动状态机继续运行
		// 因为 handleInterruptAction 内部调用了 PopInterrupt，现在的状态可能已经变了
		e.Drive()
		return nil // 中断处理完直接返回，不要往下执行普通逻辑
	}

	// === 3. 第三优先级：游戏结束拦截 ===
	if e.State.GameOver {
		return fmt.Errorf("游戏已结束")
	}

	// 3. 回合权校验
	currentPlayer := e.State.PlayerOrder[e.State.CurrentTurn]
	// 特殊情况：战斗响应阶段，允许目标玩家操作
	if e.isCombatInteractionWindow() {
		// 在战斗响应逻辑内部校验目标ID，这里先放行
	} else {
		// 其他阶段，必须是当前回合玩家
		if act.PlayerID != currentPlayer && act.Type != model.CmdStart {
			return fmt.Errorf("不是你的回合")
		}
	}

	// 这里只调用逻辑处理函数，不要在这里调用 Drive
	var err error

	switch {
	case e.isActionSelectionWindow():
		// 行动选择阶段：处理攻击、法术、特殊行动
		err = e.handleActionSelection(act)

	case e.isCombatInteractionWindow():
		// 战斗交互阶段：处理响应 (take/defend/counter)
		if act.Type == model.CmdRespond {
			err = e.handleCombatResponse(act)
		} else {
			err = fmt.Errorf("当前必须响应战斗 (使用 take/defend/counter)")
		}

	// 以前的 Start 逻辑、Confirm 逻辑等，可以根据 Phase 归类
	// 如果 Start 只能在游戏未开始时用，可以在这里加一个 case model.PhaseInit
	default:
		// 处理一些尚未归类的全局指令（如 Start）
		if act.Type == model.CmdStart {
			err = e.StartGame()
		} else {
			err = fmt.Errorf("当前状态 (%s) 不支持该指令", e.runtimeStateLabel())
		}
	}

	// === 6. 统一驱动 ===
	// 如果逻辑执行出错，直接返回错误，不驱动引擎
	if err != nil {
		return err
	}
	e.Log(fmt.Sprintf("[Debug] 指令执行成功，准备 Drive. %s, Interrupt: %v", e.runtimeStateLabel(), e.State.PendingInterrupt))

	// 如果逻辑执行成功（err == nil），说明状态已经改变（ActionQueue加了东西，或者Phase变了）
	// 这时候踩一脚油门，让自动流程跑起来
	if e.State.PendingInterrupt == nil {
		e.Drive()
	} else {
		e.Log("[Debug] 存在挂起中断，暂不 Drive")
	}

	return nil
}

// handleInterruptAction 专门处理中断状态下的输入
func (e *GameEngine) handleInterruptAction(act model.PlayerAction) error {
	if act.PlayerID != e.State.PendingInterrupt.PlayerID {
		return fmt.Errorf("当前不是等待你的响应")
	}

	switch e.State.PendingInterrupt.Type {
	case model.InterruptResponseSkill:
		return e.handleInterruptResponseSkillAction(act)
	case model.InterruptStartupSkill:
		return e.handleInterruptStartupSkillAction(act)
	case model.InterruptDiscard:
		return e.handleInterruptDiscardAction(act)
	case model.InterruptGiveCards:
		return e.handleInterruptGiveCardsAction(act)
	case model.InterruptChoice:
		return e.handleInterruptChoiceAction(act)
	case model.InterruptMagicMissile:
		if act.Type == model.CmdRespond {
			return e.handleMagicMissileResponse(act)
		}
	case model.InterruptMagicBulletFusion:
		if act.Type == model.CmdSelect {
			return e.handleMagicBulletFusionResponse(act)
		}
	case model.InterruptMagicBulletDirection:
		if act.Type == model.CmdSelect {
			return e.handleMagicBulletDirectionResponse(act)
		}
	case model.InterruptHolySwordDraw:
		if act.Type == model.CmdSelect {
			return e.handleHolySwordDrawResponse(act)
		}
	case model.InterruptSaintHeal:
		if act.Type == model.CmdSelect {
			return e.handleSaintHealResponse(act)
		}
	case model.InterruptMagicBlast:
		if act.Type == model.CmdSelect || act.Type == model.CmdCancel {
			return e.handleMagicBlastResponse(act)
		}
	}

	return fmt.Errorf("当前中断类型不支持该指令")
}

func (e *GameEngine) buildContext(user *model.Player, target *model.Player, trigger model.TriggerType, eventCtx *model.EventContext) *model.Context {
	ctx := &model.Context{
		Game:       e,
		User:       user,
		Target:     target,
		Trigger:    trigger,
		Timing:     e.defaultTimingForTrigger(trigger),
		TriggerCtx: eventCtx,
		// 初始化 map 避免 handler 写入时 panic
		Selections: make(map[string]any),
		Flags:      make(map[string]bool),
		// 当前PendingInterrupt （仅供Handler读取，不要修改）
		PendingInterrupt: e.State.PendingInterrupt,
		// 自动将单个 Target 包装进 Targets 切片，方便多目标技能处理
		Targets: []*model.Player{},
	}
	ctx.Selections["current_resume_point"] = e.currentChoiceResumePoint()
	ctx.Selections["current_turn_stage"] = string(e.State.TurnStage)
	ctx.Selections["current_combat_stage"] = string(e.State.CombatStage)
	ctx.Selections["current_subflow"] = string(e.State.Subflow)

	if target != nil {
		ctx.Targets = append(ctx.Targets, target)
	}

	return ctx
}

// AddPendingDamage 将延迟伤害添加到队列
func (e *GameEngine) AddPendingDamage(pd model.PendingDamage) {
	e.State.PendingDamageQueue = append(e.State.PendingDamageQueue, pd)
	e.Log(fmt.Sprintf("[System] 延迟伤害已添加: Source: %s, Target: %s, Damage: %d, Type: %s",
		pd.SourceID, pd.TargetID, pd.Damage, pd.DamageType))

	if !e.isDamageResolutionActive() {
		if e.State.ReturnTurnStage == "" && e.State.ReturnCombatStage == model.CombatStageNone && e.State.ReturnSubflow == model.SubflowNone {
			e.setReturnPoint(e.currentChoiceResumePoint())
		}
		e.enterDamageResolution(nil)
	}
}

// AddPendingDamageFront 将延迟伤害插入队列头部（用于“必须先结算”的伤害）。
func (e *GameEngine) AddPendingDamageFront(pd model.PendingDamage) {
	e.State.PendingDamageQueue = append([]model.PendingDamage{pd}, e.State.PendingDamageQueue...)
	e.Log(fmt.Sprintf("[System] 延迟伤害已前插: Source: %s, Target: %s, Damage: %d, Type: %s",
		pd.SourceID, pd.TargetID, pd.Damage, pd.DamageType))

	if !e.isDamageResolutionActive() {
		if e.State.ReturnTurnStage == "" && e.State.ReturnCombatStage == model.CombatStageNone && e.State.ReturnSubflow == model.SubflowNone {
			e.setReturnPoint(e.currentChoiceResumePoint())
		}
		e.enterDamageResolution(nil)
	}
}

// processPendingDamages 处理伤害队列中的所有伤害
// 返回 true 如果产生了中断需要暂停 Drive
func (e *GameEngine) processPendingDamages() bool {
	for len(e.State.PendingDamageQueue) > 0 {
		// Peek: 取出队列中第一个延迟伤害（暂不弹出，等待所有步骤完成）
		pd := &e.State.PendingDamageQueue[0]

		// 先单独处理攻击伤害命中链（OnAttackHit）。
		if e.processPendingAttackHit(pd) {
			return true
		}

		// [重构] 如果攻击已经被判定为未命中（如圣盾格挡），直接移除并继续处理下一个
		if pd.AttackMissResolved {
			e.State.PendingDamageQueue = e.State.PendingDamageQueue[1:]
			continue
		}

		// 通用承伤前流程（所有伤害都走这里）。
		// 角色机制（如灵魂链接）通过 pending damage runtime hooks 插入。
		if e.runPendingDamageBeforeTakenHooks(pd) {
			return true
		}

		if e.dispatchPendingDamageTaken(pd) {
			return true
		}
		if strings.EqualFold(pd.DamageType, "Attack") && !pd.AttackMissResolved && !pd.AttackPostHitEffectsDone {
			e.resolveSwordEmperorAttackHitAftermath(pd)
			pd.AttackPostHitEffectsDone = true
		}
		if e.resolvePendingDamageHealChoice(pd) {
			return true
		}
		// 蝶舞者：伤害应用前的时点响应（朝圣/毒粉/镜花水月）。
		if e.maybeTriggerButterflyDamageResponses(pd) {
			return true
		}

		// 应用伤害 & 移除效果
		if pd.Damage < 0 {
			pd.Damage = 0
		}

		target := e.State.Players[pd.TargetID]
		source := e.State.Players[pd.SourceID]
		if target != nil && pd.Damage > 0 {
			if pd.DamageType == "Attack" && source != nil {
				e.NotifyActionStep(fmt.Sprintf("总共对%s造成%d点伤害", model.GetPlayerDisplayName(target), pd.Damage))
			}
			e.NotifyDamageDealt(pd.SourceID, pd.TargetID, pd.Damage, pd.DamageType)
		}
		if target != nil {
			// 执行实际扣血/摸牌逻辑
			e.applyDamageWithOptions(target, pd.Damage, pd.DamageType, pd.CapDrawToHandLimit, pd.SourceID, pd.SourceSkillID, pd.OverflowMoraleLossFixed)
			// 角色落伤后清理逻辑统一交由 hook 扩展。
			e.runPendingDamageAfterApplyHooks(pd, target)

			// ==========================================
			// 五系封印移除点：伤害结算后移除封印
			// ==========================================
			// 如果指定了 EffectTypeToRemove，在伤害结算后移除场上效果
			// 【五系封印在此处移除】
			// 封印触发时会在 PendingDamage 中设置 EffectTypeToRemove
			if pd.EffectTypeToRemove != "" {
				e.RemoveFieldCard(target.ID, pd.EffectTypeToRemove)
				e.Log(fmt.Sprintf("[System] 移除了 %s 的场上效果: %s", target.Name, pd.EffectTypeToRemove))
			}
		}
		resolved := *pd

		// 处理完毕，从队列中弹出
		e.State.PendingDamageQueue = e.State.PendingDamageQueue[1:]
		// 伤害结算后触发额外技能（例如动物伙伴）。
		if e.handlePostDamageResolved(&resolved) {
			return true
		}

		// 伤害结算可能产生新的中断 (例如爆牌弃牌)，如果有中断，暂停
		if e.State.PendingInterrupt != nil {
			return true
		}
	}
	return false
}
