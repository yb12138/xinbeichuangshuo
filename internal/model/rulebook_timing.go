package model

const (
	TimingGameInitial     Timing = "game.initial"
	TimingTurnBeforeStart Timing = "turn.before_start"
	TimingTurnStart       Timing = "turn.start"
	TimingActionBefore    Timing = "turn.action_before"
	TimingActionStart     Timing = "turn.action_start"
	TimingActionDuring    Timing = "turn.action_during"
	TimingActionEnd       Timing = "turn.action_end"
	TimingActionPost      Timing = "turn.action_post"
	TimingTurnEnd         Timing = "turn.end"

	TimingAttackDeclare         Timing = "attack.declare"
	TimingAttackSelectTarget    Timing = "attack.select_target"
	TimingAttackPlayCard        Timing = "attack.play_card"
	TimingAttackModifyCard      Timing = "attack.modify_card"
	TimingAttackCommitted       Timing = "attack.committed"
	TimingAttackForceHitCheck   Timing = "attack.force_hit_check"
	TimingAttackNoResponseCheck Timing = "attack.no_response_check"
	TimingAttackResponse        Timing = "attack.response"
	TimingAttackHit             Timing = "attack.hit"
	TimingAttackMiss            Timing = "attack.miss"

	TimingMagicDeclare      Timing = "magic.declare"
	TimingMagicSelectTarget Timing = "magic.select_target"
	TimingMagicValidate     Timing = "magic.validate"
	TimingMagicResolve      Timing = "magic.resolve"
	TimingMagicHealOverflow Timing = "magic.heal_overflow"

	TimingMagicMissileResponse      Timing = "magic.missile.response"
	TimingMagicMissileDefend        Timing = "magic.missile.defend"
	TimingMagicMissileCounter       Timing = "magic.missile.counter"
	TimingMagicMissileResponseSkill Timing = "magic.missile.response_skill"

	TimingDamageSourceDeal   Timing = "damage.source_deal"
	TimingDamageTargetBefore Timing = "damage.target_before"
	TimingHealBefore         Timing = "heal.before"
	TimingHealUse            Timing = "heal.use"
	TimingHealCap            Timing = "heal.cap"
	TimingDamageApplied      Timing = "damage.applied"
	TimingDamageTaken        Timing = "damage.taken"
	TimingSettleDraw         Timing = "settle.draw"
	TimingSettleDiscard      Timing = "settle.discard"
	TimingSettleHandLimit    Timing = "settle.hand_limit_check"
	TimingMoraleLossCheck    Timing = "settle.morale_loss_check"
	TimingMoraleLossApplied  Timing = "settle.morale_loss_applied"
	TimingGameEndCheck       Timing = "settle.game_end_check"
	TimingDamageResolved     Timing = "damage.resolved"
)

// NormalizeTiming maps legacy engine timings to their rulebook timeline.
func NormalizeTiming(t Timing) Timing {
	switch t {
	case TimingOnGameStart, Timing("on_game_start"):
		return TimingGameInitial
	case Timing("on_turn_before_start"):
		return TimingTurnBeforeStart
	case TimingOnTurnStart, Timing("on_turn_start"):
		return TimingTurnStart
	case TimingOnBeforeAction, Timing("before_action"):
		return TimingActionBefore
	case TimingBeforeActionExecute, TimingStartup:
		return TimingActionStart
	case TimingOnActionEnd, Timing("on_action_end"):
		return TimingActionEnd
	case Timing("post_action_end"):
		return TimingActionPost
	case TimingOnTurnEnd, Timing("on_turn_end"), Timing("on_turn_end_final"):
		return TimingTurnEnd
	case TimingOnAttackDeclared, Timing("on_attack_declared"):
		return TimingAttackDeclare
	case Timing("on_attack_target_ctx"):
		return TimingAttackSelectTarget
	case Timing("on_attack_card_hook"), Timing("on_attack_card_transform"):
		return TimingAttackModifyCard
	case Timing("on_attack_gating"), Timing("on_hit_check"), TimingOnHitCheck:
		return TimingAttackResponse
	case Timing("post_attack_hit"):
		return TimingAttackHit
	case Timing("on_attack_miss"):
		return TimingAttackMiss
	case TimingOnMagicDeclared:
		return TimingMagicDeclare
	case TimingOnHealOverflow:
		return TimingMagicHealOverflow
	case Timing("on_magic_missile_defend"):
		return TimingMagicMissileDefend
	case Timing("on_magic_missile_counter"):
		return TimingMagicMissileCounter
	case Timing("on_magic_missile_response_skill_aug"):
		return TimingMagicMissileResponseSkill
	case TimingOnDamageCalculated, Timing("on_damage_calculate"):
		return TimingDamageSourceDeal
	case Timing("on_damage_before_taken"):
		return TimingDamageTargetBefore
	case Timing("on_heal_resist"):
		return TimingHealBefore
	case Timing("on_heal_cap_calculate"):
		return TimingHealCap
	case TimingOnDamageApplied, Timing("on_damage_applied"):
		return TimingDamageApplied
	case TimingOnDamageTaken, Timing("on_damage_taken"):
		return TimingDamageTaken
	case Timing("post_damage_resolved"), Timing("on_damage_after_apply"):
		return TimingDamageResolved
	case TimingBeforeCardDrawn, TimingOnCardDrawn:
		return TimingSettleDraw
	case TimingOnCardDiscarded, TimingOnCardPlayedOrRevealed:
		return TimingSettleDiscard
	case TimingBeforeMoraleLoss:
		return TimingMoraleLossCheck
	case Timing("on_morale_loss_applied"):
		return TimingMoraleLossApplied
	default:
		return t
	}
}

// LegacyTimingName maps split rulebook timings back to the closest legacy
// timing used by existing skill dispatch, debug flows, and UI compatibility.
func LegacyTimingName(t Timing) Timing {
	switch t {
	case TimingGameInitial:
		return TimingOnGameStart
	case TimingTurnBeforeStart, TimingTurnStart:
		return TimingOnTurnStart
	case TimingActionBefore:
		return TimingOnBeforeAction
	case TimingActionStart:
		return TimingStartup
	case TimingActionEnd, TimingActionPost:
		return TimingOnActionEnd
	case TimingTurnEnd:
		return TimingOnTurnEnd
	case TimingAttackDeclare, TimingAttackSelectTarget, TimingAttackPlayCard, TimingAttackModifyCard, TimingAttackCommitted:
		return TimingOnAttackDeclared
	case TimingAttackForceHitCheck, TimingAttackNoResponseCheck, TimingAttackResponse, TimingAttackHit, TimingAttackMiss:
		return TimingOnHitCheck
	case TimingMagicDeclare, TimingMagicSelectTarget, TimingMagicValidate, TimingMagicResolve, TimingMagicHealOverflow:
		return TimingOnMagicDeclared
	case TimingMagicMissileDefend:
		return Timing("on_magic_missile_defend")
	case TimingMagicMissileCounter:
		return Timing("on_magic_missile_counter")
	case TimingMagicMissileResponseSkill:
		return Timing("on_magic_missile_response_skill_aug")
	case TimingDamageSourceDeal:
		return TimingOnDamageCalculated
	case TimingDamageApplied:
		return TimingOnDamageApplied
	case TimingDamageTaken:
		return TimingOnDamageTaken
	case TimingMoraleLossCheck:
		return TimingBeforeMoraleLoss
	case TimingSettleDraw:
		return TimingOnCardDrawn
	case TimingSettleDiscard:
		return TimingOnCardDiscarded
	default:
		return t
	}
}
