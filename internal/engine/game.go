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
	// TimingOnAttackDeclared: 按已上场角色动态装配的执行表。
	attackDeclaredCardTransformHooks   []attackCardRuntimeTransformHook
	attackDeclaredTargetContextHooks   []attackTargetContextHook
	attackDeclaredStateResetHooks      []attackStartStateResetHook
	attackDeclaredPreCombatHooks       []attackPreCombatHook
	attackDeclaredPendingDamageInitOps []pendingDamageAttackInitHook
	attackDeclaredOpHandlers           map[timingOnAttackDeclaredOp]timingOnAttackDeclaredHandler
	// TimingOnHitCheck: 命中判定阶段动态装配执行表。
	hitCheckCombatInteractionHooks         []combatInteractionPolicyHook
	hitCheckCombatDefendValidationPolicies []combatDefendValidationPolicy
	hitCheckCombatCounterCardPolicies      []combatCounterCardPolicy
	hitCheckCombatCounterElementPolicies   []combatCounterElementPolicy
	hitCheckCombatCounterResolvePolicies   []combatCounterResolvePolicy
	hitCheckMagicMissileDefendPolicies     []magicMissileDefendValidationPolicy
	hitCheckMagicMissileCounterPolicies    []magicMissileCounterValidationPolicy
	hitCheckResponseSkipHooks              []responseSkipHook
	hitCheckResponseSkillIDAugmenters      []responseSkillIDAugmenter
	hitCheckResponseSkillIDNormalizers     []responseSkillIDNormalizer
	hitCheckOpHandlers                     map[timingOnHitCheckOp]timingOnHitCheckHandler
	hitCheckSkillOpHandlers                map[timingOnHitCheckSkillOp]timingOnHitCheckSkillHandler
	// TimingOnDamageCalculated: 伤害计算阶段动态装配执行表。
	damageCalculatedAttackPassiveHooks []attackPassiveDamageHook
	damageCalculatedBeforeTakenHooks   []pendingDamageBeforeTakenHook
	damageCalculatedHealCapHooks       []pendingDamageHealCapHook
	damageCalculatedHealResistRules    []pendingDamageHealResistRule
	damageTakenAfterTakenHooks         []pendingDamageAfterTakenHook
	damageAppliedBeforeApplyHooks      []pendingDamageBeforeApplyHook
	damageTakenAfterApplyHooks         []pendingDamageAfterApplyHook
	damageTakenAfterResolvedHooks      []pendingDamageAfterResolvedHook
	damageCalculatedOpHandlers         map[timingOnDamageCalculatedOp]timingOnDamageCalculatedHandler
	// Turn 主流程阶段：动态装配阶段钩子与中断策略。
	turnStartBeforeStartHooks      []turnTimingHook
	turnStartMainHooks             []turnTimingHook
	beforeActionFieldHooks         []turnTimingHook
	beforeActionExecuteHooks       []turnTimingHook
	beforeActionOptionPolicies     []actionSelectionOptionPolicy
	beforeActionValidationPolicies []actionSelectionValidationPolicy
	turnEndPreExtraHooks           []turnTimingHook
	turnEndFinalHooks              []turnTimingHook
	attackDeclaredInterrupts       []attackDeclaredInterruptHook
	actionEndInterrupts            []actionEndInterruptHook
	// TimingBeforeActionExecute / TimingOnActionEnd：特殊行动策略装配表。
	specialActionOverridePolicies []specialActionOverridePolicy
	specialActionPostHooks        []specialActionPostHook
	// Skill 主动发动成功后的后置策略装配表。
	skillPostHooks []skillPostHook
	// TimingOnGameStart：入场初始化 / 开局发牌后 / 行动收尾。
	gameStartAddPlayerHooks   []gameStartPlayerHook
	gameStartInitialDealHooks []gameStartPlayerHook
	gameStartFinalizeHooks    []gameStartFinalizeHook
	// TimingOnCampChanged：玩家入场派生状态 / 星杯变更派生状态。
	campChangedPlayerSetupHooks []campChangedPlayerHook
	campChangedCampCupHooks     []campChangedCupHook
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
