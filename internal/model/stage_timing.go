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

type TriggerTiming string

const (
	TimingUnknown                TriggerTiming = ""
	TimingOnGameStart            TriggerTiming = "TimingOnGameStart"
	TimingOnCampChanged          TriggerTiming = "TimingOnCampChanged"
	TimingActive                 TriggerTiming = "TimingActive"
	TimingStartup                TriggerTiming = "TimingStartup"
	TimingOnTurnStart            TriggerTiming = "TimingOnTurnStart"
	TimingOnBeforeAction         TriggerTiming = "TimingOnBeforeAction"
	TimingBeforeActionExecute    TriggerTiming = "TimingBeforeActionExecute"
	TimingOnActionEnd            TriggerTiming = "TimingOnActionEnd"
	TimingOnSkillExecuted        TriggerTiming = "TimingOnSkillExecuted"
	TimingOnAttackDeclared       TriggerTiming = "TimingOnAttackDeclared"
	TimingOnMagicDeclared        TriggerTiming = "TimingOnMagicDeclared"
	TimingOnHitCheck             TriggerTiming = "TimingOnHitCheck"
	TimingOnDamageCalculated     TriggerTiming = "TimingOnDamageCalculated"
	TimingOnDamageApplied        TriggerTiming = "TimingOnDamageApplied"
	TimingOnDamageTaken          TriggerTiming = "TimingOnDamageTaken"
	TimingBeforeMoraleLoss       TriggerTiming = "TimingBeforeMoraleLoss"
	TimingBeforeCardDrawn        TriggerTiming = "TimingBeforeCardDrawn"
	TimingOnCardDrawn            TriggerTiming = "TimingOnCardDrawn"
	TimingOnCardDiscarded        TriggerTiming = "TimingOnCardDiscarded"
	TimingOnCardPlayedOrRevealed TriggerTiming = "TimingOnCardPlayedOrRevealed"
	TimingOnHealOverflow         TriggerTiming = "TimingOnHealOverflow"
	TimingOnFieldMarkChanged     TriggerTiming = "TimingOnFieldMarkChanged"
	TimingOnOrientationChanged   TriggerTiming = "TimingOnOrientationChanged"
	TimingOnTurnEnd              TriggerTiming = "TimingOnTurnEnd"
)
