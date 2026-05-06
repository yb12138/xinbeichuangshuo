// gameflow: GameEngine 构造与核心字段；钩子表占位。

package engine

import (
	engineplayer "starcup-engine/internal/engine/player"

	choicert "starcup-engine/internal/engine/runtime/choice"
	intr "starcup-engine/internal/engine/runtime/interrupt"
	skillrt "starcup-engine/internal/engine/runtime/skill"
	"starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
)

// GameEngine 对局运行时核心：持有全局 GameState，并在 Drive 中推进回合/战斗/中断。
// 各切片与 map 为「按当前上场角色动态装配」的策略钩子，用于把通用流程与角色特例解耦。
type GameEngine struct {
	// State 完整对局快照（回合、玩家、战斗栈、中断、牌库等）；所有流程读写的中枢。
	State *model.GameState
	// dispatcher 技能按 FlowTiming 分发与执行（含可选响应中断）。
	dispatcher *SkillDispatcher
	// observer 对局事件观察者（日志、协议推送）；不参与规则判定。
	observer model.GameObserver
	// lifecycle 攻击生命周期辅助（与 CombatStack 协同的细粒度步骤）。
	lifecycle AttackLifecycle
	// skillRuntime 新架构的技能运行时。
	skillRuntime *skillrt.Runtime
	// choiceEngine 统一 InterruptChoice（仅 ChoiceSpec，无隐式回退）。
	choiceEngine *choicert.Engine
	// interruptOrchestrator 中断动作与 Prompt 编排。
	interruptOrchestrator *intr.Orchestrator
	// Turn 主流程阶段：动态装配阶段钩子与中断策略。
	beforeActionFieldHooks []turnTimingHook

	// TimingOnGameStart：入场初始化 / 开局发牌后。
	gameStartAddPlayerHooks   []gameStartPlayerHook
	gameStartInitialDealHooks []gameStartPlayerHook
	// turnMagicDamageTargets 本回合「施法者 → 曾对其造成法术伤害的敌方」；用于吟游诗人【沉沦协奏曲】等统计。
	turnMagicDamageTargets map[string]map[string]bool
	// actionSummary 当前行动周期内的资源/治疗等汇总，供日志或后续效果读取。
	actionSummary *actionSummary
	// actionSummaryTurn 汇总所绑定的回合索引，用于跨回合清空。
	actionSummaryTurn int
	// suppressSealOnDiscard 为 true 时跳过某次弃牌链路上的封印检测（避免死循环或特殊剧情）。
	suppressSealOnDiscard bool
	// roleTimingHooks 声明式 Timing Hook 注册表（按 timing 分组，已排序）。
	roleTimingHooks map[engineplayer.TimingPoint][]roleTimingHookEntry
	// skillResume 当前待恢复的技能执行上下文。
	skillResume *skillResumeState
	// postActionEndResume 当前待补执行的行动后效果。
	postActionEndResume *postActionEndResumeState
}

type skillResumeState struct {
	playerID       string
	skillID        string
	targetIDs      []string
	discardedCards []model.Card
}

type postActionEndResumeState struct {
	playerID   string
	actionType model.ActionType
}

func (e *GameEngine) resetTurnMagicDamageTracker() {
	e.turnMagicDamageTargets = map[string]map[string]bool{}
}

// buildContext：组装 User/Target/Timing/EventCtx。
func (e *GameEngine) buildContext(user *model.Player, target *model.Player, timing model.FlowTiming, eventCtx *model.EventContext) *model.Context {
	ctx := &model.Context{
		Game:             e,
		User:             user,
		Target:           target,
		Timing:           timing,
		EventCtx:         eventCtx,
		Selections:       make(map[string]any),
		Flags:            make(map[string]bool),
		PendingInterrupt: e.State.PendingInterrupt,
		Targets:          []*model.Player{},
	}
	ctx.Selections["current_resume_point"] = e.currentChoiceResumePoint()
	ctx.Selections["current_turn_stage"] = e.State.TurnStage
	ctx.Selections["current_combat_stage"] = e.State.CombatStage
	ctx.Selections["current_subflow"] = e.State.Subflow
	if target != nil {
		ctx.Targets = append(ctx.Targets, target)
	}
	return ctx
}

// NewGameEngine 构造引擎：初始化状态、注册技能 handler、装配 TimingOnAttackDeclared 等钩子表。
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
	engine.skillRuntime = skillrt.NewRuntime()
	engine.dispatcher.SetRuntime(engine.skillRuntime)
	engine.choiceEngine = choicert.NewEngine(choicert.NewSpecRegistry())
	engine.choiceEngine.SetHost(&choiceHostBridge{e: engine})
	bootstrapChoiceSpecs(engine)
	engine.installInterruptOrchestrator()
	engine.rebuildTimingOnAttackDeclaredRegistry()
	engine.roleTimingHooks = mountRoleTimingHooks()
	return engine
}
