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
)
