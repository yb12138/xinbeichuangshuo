package engine

import (
	"starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
)

type GameEngine struct {
	State      *model.GameState
	dispatcher *SkillDispatcher
	observer   model.GameObserver // [新增] 持有观察者
	lifecycle  AttackLifecycle
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
	engine.rebuildTimingOnAttackDeclaredRegistry()
	return engine
}
