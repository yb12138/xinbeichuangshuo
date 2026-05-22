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
	TimingUnknown                FlowTiming = ""
	TimingOnGameStart            FlowTiming = "TimingOnGameStart"
	TimingOnCampChanged          FlowTiming = "TimingOnCampChanged"
	TimingActive                 FlowTiming = "TimingActive"
	TimingStartup                FlowTiming = "TimingStartup"
	TimingOnTurnStart            FlowTiming = "TimingOnTurnStart"
	TimingOnBeforeAction         FlowTiming = "TimingOnBeforeAction"
	TimingBeforeActionExecute    FlowTiming = "TimingBeforeActionExecute"
	TimingOnActionEnd            FlowTiming = "TimingOnActionEnd"
	TimingOnSkillExecuted        FlowTiming = "TimingOnSkillExecuted"
	TimingOnAttackDeclared       FlowTiming = "TimingOnAttackDeclared"
	TimingOnMagicDeclared        FlowTiming = "TimingOnMagicDeclared"
	TimingOnHitCheck             FlowTiming = "TimingOnHitCheck"
	TimingOnDamageCalculated     FlowTiming = "TimingOnDamageCalculated"
	TimingOnDamageApplied        FlowTiming = "TimingOnDamageApplied"
	TimingOnDamageTaken          FlowTiming = "TimingOnDamageTaken"
	TimingBeforeMoraleLoss       FlowTiming = "TimingBeforeMoraleLoss"
	TimingBeforeCardDrawn        FlowTiming = "TimingBeforeCardDrawn"
	TimingOnCardDrawn            FlowTiming = "TimingOnCardDrawn"
	TimingOnCardDiscarded        FlowTiming = "TimingOnCardDiscarded"
	TimingOnCardPlayedOrRevealed FlowTiming = "TimingOnCardPlayedOrRevealed"
	TimingOnHealOverflow         FlowTiming = "TimingOnHealOverflow"
	TimingOnFieldMarkChanged     FlowTiming = "TimingOnFieldMarkChanged"
	TimingOnOrientationChanged   FlowTiming = "TimingOnOrientationChanged"
	TimingOnTurnEnd              FlowTiming = "TimingOnTurnEnd"
)
