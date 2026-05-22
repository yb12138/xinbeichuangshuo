package model

type TurnStage string

const (
	TurnStageTurnBeforeStart TurnStage = "TurnBeforeStart"
	TurnStageTurnStart       TurnStage = "TurnStart"
	TurnStageBeforeAction    TurnStage = "BeforeAction"
	TurnStageActionStart     TurnStage = "ActionStart"
	TurnStageActionExecution TurnStage = "ActionExecution"
	TurnStageActionEnd       TurnStage = "ActionEnd"
	TurnStageExtraAction     TurnStage = "ExtraAction"
	TurnStageTurnEnd         TurnStage = "TurnEnd"
)

type CombatStage string

const (
	CombatStageNone       CombatStage = ""
	CombatStageDeclare    CombatStage = "CombatDeclare"
	CombatStageHitCheck   CombatStage = "CombatHitCheck"
	CombatStageCalcDamage CombatStage = "CombatCalcDamage"
	CombatStageHeal       CombatStage = "CombatHeal"
	CombatStageApply      CombatStage = "CombatApply"
	CombatStageDraw       CombatStage = "CombatDraw"
)

type Subflow string

const (
	SubflowNone             Subflow = ""
	SubflowResponse         Subflow = "Response"
	SubflowDiscardSelection Subflow = "DiscardSelection"
)

// Timing is the canonical rule timing identifier used by both skill dispatch
// and role timing hooks.
type Timing string

// FlowTiming is kept as a compatibility alias for existing skill flow code.
type FlowTiming = Timing

const (
	TimingUnknown FlowTiming = ""
	// Deprecated: use TimingGameInitial.
	TimingOnGameStart FlowTiming = "TimingOnGameStart"
	// Deprecated: lifecycle timing; keep runtime-only compatibility.
	TimingOnCampChanged FlowTiming = "TimingOnCampChanged"
	// Deprecated: use TimingActionDuring for rulebook timeline work.
	TimingActive FlowTiming = "TimingActive"
	// Deprecated: use TimingActionStart.
	TimingStartup FlowTiming = "TimingStartup"
	// Deprecated: use TimingTurnStart.
	TimingOnTurnStart FlowTiming = "TimingOnTurnStart"
	// Deprecated: use TimingActionBefore.
	TimingOnBeforeAction FlowTiming = "TimingOnBeforeAction"
	// Deprecated: use TimingActionStart.
	TimingBeforeActionExecute FlowTiming = "TimingBeforeActionExecute"
	// Deprecated: use TimingActionEnd or TimingActionPost according to the rulebook phase.
	TimingOnActionEnd FlowTiming = "TimingOnActionEnd"
	// Deprecated: lifecycle timing; keep compatibility until skill-post hooks are fully migrated.
	TimingOnSkillExecuted FlowTiming = "TimingOnSkillExecuted"
	// Deprecated: use TimingMagicDeclare, TimingMagicSelectTarget, TimingMagicValidate, TimingMagicResolve, or TimingMagicHealOverflow.
	TimingOnMagicDeclared FlowTiming = "TimingOnMagicDeclared"
	// Deprecated: use TimingDamageApplied.
	TimingOnDamageApplied FlowTiming = "TimingOnDamageApplied"
	// Deprecated: use TimingSettleDraw.
	TimingBeforeCardDrawn FlowTiming = "TimingBeforeCardDrawn"
	// Deprecated: use TimingSettleDraw.
	TimingOnCardDrawn FlowTiming = "TimingOnCardDrawn"
	// Deprecated: use TimingSettleDiscard.
	TimingOnCardDiscarded FlowTiming = "TimingOnCardDiscarded"
	// Deprecated: use TimingSettleDiscard for settlement, or keep this legacy card-use timing until card play phases are fully split.
	TimingOnCardPlayedOrRevealed FlowTiming = "TimingOnCardPlayedOrRevealed"
	// Deprecated: use TimingMagicHealOverflow or TimingHealCap according to context.
	TimingOnHealOverflow FlowTiming = "TimingOnHealOverflow"
	// Deprecated: lifecycle timing; keep compatibility until field mark hooks are split.
	TimingOnFieldMarkChanged FlowTiming = "TimingOnFieldMarkChanged"
	// Deprecated: lifecycle timing; keep compatibility until orientation hooks are split.
	TimingOnOrientationChanged FlowTiming = "TimingOnOrientationChanged"
	// Deprecated: use TimingTurnEnd.
	TimingOnTurnEnd FlowTiming = "TimingOnTurnEnd"
)
