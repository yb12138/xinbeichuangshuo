package model

// TimingCategory groups canonical timing identifiers by their rule timeline.
type TimingCategory string

const (
	TimingCategoryUnknown TimingCategory = ""
	TimingCategorySystem  TimingCategory = "system"
	TimingCategoryTurn    TimingCategory = "turn"
	TimingCategoryAttack  TimingCategory = "attack"
	TimingCategoryMagic   TimingCategory = "magic"
	TimingCategoryDamage  TimingCategory = "damage"
	TimingCategorySettle  TimingCategory = "settle"
)

// TimingDescriptor records how a timing is consumed before it is dispatched.
type TimingDescriptor struct {
	ID              Timing
	Category        TimingCategory
	RuleName        string
	SkillVisible    bool
	RoleHookVisible bool
	RuntimeOnly     bool
}

var timingDescriptors = map[Timing]TimingDescriptor{
	// Skill flow timings.
	TimingUnknown:                timingDescriptor(TimingUnknown, TimingCategorySystem, "Unknown timing", false, false, true),
	TimingOnGameStart:            timingDescriptor(TimingOnGameStart, TimingCategorySystem, "Game start", true, false, false, Timing("on_game_start")),
	TimingOnCampChanged:          timingDescriptor(TimingOnCampChanged, TimingCategorySystem, "Camp changed", true, false, false, Timing("on_camp_changed")),
	TimingActive:                 timingDescriptor(TimingActive, TimingCategoryTurn, "Active action window", true, false, false),
	TimingStartup:                timingDescriptor(TimingStartup, TimingCategoryTurn, "Action startup", true, false, false),
	TimingOnTurnStart:            timingDescriptor(TimingOnTurnStart, TimingCategoryTurn, "Turn start", true, false, false, Timing("on_turn_start")),
	TimingOnBeforeAction:         timingDescriptor(TimingOnBeforeAction, TimingCategoryTurn, "Before action", true, false, false, Timing("before_action")),
	TimingBeforeActionExecute:    timingDescriptor(TimingBeforeActionExecute, TimingCategoryTurn, "Before action execute", true, false, false),
	TimingOnActionEnd:            timingDescriptor(TimingOnActionEnd, TimingCategoryTurn, "Action end", true, false, false, Timing("on_action_end"), Timing("post_action_end")),
	TimingOnSkillExecuted:        timingDescriptor(TimingOnSkillExecuted, TimingCategoryTurn, "Skill executed", true, false, false, Timing("on_skill_post")),
	TimingOnMagicDeclared:        timingDescriptor(TimingOnMagicDeclared, TimingCategoryMagic, "Magic declared", true, false, false),
	TimingOnDamageApplied:        timingDescriptor(TimingOnDamageApplied, TimingCategoryDamage, "Damage applied", true, false, false, Timing("on_damage_applied")),
	TimingBeforeMoraleLoss:       timingDescriptor(TimingBeforeMoraleLoss, TimingCategorySettle, "Before morale loss", true, false, false, Timing("on_morale_loss_applied")),
	TimingBeforeCardDrawn:        timingDescriptor(TimingBeforeCardDrawn, TimingCategorySettle, "Before card drawn", true, false, false),
	TimingOnCardDrawn:            timingDescriptor(TimingOnCardDrawn, TimingCategorySettle, "Card drawn", true, false, false),
	TimingOnCardDiscarded:        timingDescriptor(TimingOnCardDiscarded, TimingCategorySettle, "Card discarded", true, false, false),
	TimingOnCardPlayedOrRevealed: timingDescriptor(TimingOnCardPlayedOrRevealed, TimingCategorySettle, "Card played or revealed", true, false, false),
	TimingOnHealOverflow:         timingDescriptor(TimingOnHealOverflow, TimingCategoryMagic, "Heal overflow", true, false, false),
	TimingOnFieldMarkChanged:     timingDescriptor(TimingOnFieldMarkChanged, TimingCategorySystem, "Field mark changed", true, false, false),
	TimingOnOrientationChanged:   timingDescriptor(TimingOnOrientationChanged, TimingCategorySystem, "Orientation changed", true, false, false),
	TimingOnTurnEnd:              timingDescriptor(TimingOnTurnEnd, TimingCategoryTurn, "Turn end", true, false, false, Timing("on_turn_end"), Timing("on_turn_end_final")),

	// Rulebook timings.
	TimingGameInitial:     timingDescriptor(TimingGameInitial, TimingCategorySystem, "Game initial", true, true, false, TimingOnGameStart, Timing("on_game_start")),
	TimingTurnBeforeStart: timingDescriptor(TimingTurnBeforeStart, TimingCategoryTurn, "Turn before start", true, true, false, Timing("on_turn_before_start")),
	TimingTurnStart:       timingDescriptor(TimingTurnStart, TimingCategoryTurn, "Turn start", true, true, false, TimingOnTurnStart, Timing("on_turn_start")),
	TimingActionBefore:    timingDescriptor(TimingActionBefore, TimingCategoryTurn, "Action before", true, true, false, TimingOnBeforeAction, Timing("before_action")),
	TimingActionStart:     timingDescriptor(TimingActionStart, TimingCategoryTurn, "Action start", true, true, false, TimingStartup, TimingBeforeActionExecute),
	TimingActionDuring:    timingDescriptor(TimingActionDuring, TimingCategoryTurn, "Action during", true, true, false, TimingActive),
	TimingActionEnd:       timingDescriptor(TimingActionEnd, TimingCategoryTurn, "Action end", true, true, false, TimingOnActionEnd, Timing("on_action_end")),
	TimingActionPost:      timingDescriptor(TimingActionPost, TimingCategoryTurn, "Action post", true, true, false, Timing("post_action_end")),
	TimingTurnEnd:         timingDescriptor(TimingTurnEnd, TimingCategoryTurn, "Turn end", true, true, false, TimingOnTurnEnd, Timing("on_turn_end"), Timing("on_turn_end_final")),

	TimingAttackDeclare:         timingDescriptor(TimingAttackDeclare, TimingCategoryAttack, "Attack declare", true, true, false),
	TimingAttackSelectTarget:    timingDescriptor(TimingAttackSelectTarget, TimingCategoryAttack, "Attack select target", true, true, false, Timing("on_attack_target_ctx")),
	TimingAttackPlayCard:        timingDescriptor(TimingAttackPlayCard, TimingCategoryAttack, "Attack play card", true, true, false),
	TimingAttackModifyCard:      timingDescriptor(TimingAttackModifyCard, TimingCategoryAttack, "Attack modify card", true, true, false, Timing("on_attack_card_hook"), Timing("on_attack_card_transform")),
	TimingAttackCommitted:       timingDescriptor(TimingAttackCommitted, TimingCategoryAttack, "Attack committed", true, true, false),
	TimingAttackForceHitCheck:   timingDescriptor(TimingAttackForceHitCheck, TimingCategoryAttack, "Attack force hit check", true, true, false),
	TimingAttackNoResponseCheck: timingDescriptor(TimingAttackNoResponseCheck, TimingCategoryAttack, "Attack no response check", true, true, false, Timing("on_attack_gating")),
	TimingAttackResponse:        timingDescriptor(TimingAttackResponse, TimingCategoryAttack, "Attack response", true, true, false),
	TimingAttackHit:             timingDescriptor(TimingAttackHit, TimingCategoryAttack, "Attack hit", true, true, false, Timing("post_attack_hit")),
	TimingAttackMiss:            timingDescriptor(TimingAttackMiss, TimingCategoryAttack, "Attack miss", true, true, false, Timing("on_attack_miss")),

	TimingMagicDeclare:      timingDescriptor(TimingMagicDeclare, TimingCategoryMagic, "Magic declare", true, true, false, TimingOnMagicDeclared),
	TimingMagicSelectTarget: timingDescriptor(TimingMagicSelectTarget, TimingCategoryMagic, "Magic select target", true, true, false),
	TimingMagicValidate:     timingDescriptor(TimingMagicValidate, TimingCategoryMagic, "Magic validate", true, true, false),
	TimingMagicResolve:      timingDescriptor(TimingMagicResolve, TimingCategoryMagic, "Magic resolve", true, true, false),
	TimingMagicHealOverflow: timingDescriptor(TimingMagicHealOverflow, TimingCategoryMagic, "Magic heal overflow", true, true, false, TimingOnHealOverflow),

	TimingMagicMissileResponse:      timingDescriptor(TimingMagicMissileResponse, TimingCategoryMagic, "Magic missile response", true, true, false),
	TimingMagicMissileDefend:        timingDescriptor(TimingMagicMissileDefend, TimingCategoryMagic, "Magic missile defend", true, true, false, Timing("on_magic_missile_defend")),
	TimingMagicMissileCounter:       timingDescriptor(TimingMagicMissileCounter, TimingCategoryMagic, "Magic missile counter", true, true, false, Timing("on_magic_missile_counter")),
	TimingMagicMissileResponseSkill: timingDescriptor(TimingMagicMissileResponseSkill, TimingCategoryMagic, "Magic missile response skill", true, true, false, Timing("on_magic_missile_response_skill_aug")),

	TimingDamageSourceDeal:   timingDescriptor(TimingDamageSourceDeal, TimingCategoryDamage, "Damage source deal", true, true, false),
	TimingDamageTargetBefore: timingDescriptor(TimingDamageTargetBefore, TimingCategoryDamage, "Damage target before", true, true, false, Timing("on_damage_before_taken")),
	TimingHealBefore:         timingDescriptor(TimingHealBefore, TimingCategoryDamage, "Heal before", true, true, false, Timing("on_heal_resist")),
	TimingHealUse:            timingDescriptor(TimingHealUse, TimingCategoryDamage, "Heal use", true, true, false),
	TimingHealCap:            timingDescriptor(TimingHealCap, TimingCategoryDamage, "Heal cap", true, true, false, Timing("on_heal_cap_calculate")),
	TimingDamageApplied:      timingDescriptor(TimingDamageApplied, TimingCategoryDamage, "Damage applied", true, true, false, TimingOnDamageApplied, Timing("on_damage_applied")),
	TimingDamageTaken:        timingDescriptor(TimingDamageTaken, TimingCategoryDamage, "Damage taken", true, true, false),
	TimingSettleDraw:         timingDescriptor(TimingSettleDraw, TimingCategorySettle, "Settle draw", true, true, false, TimingBeforeCardDrawn, TimingOnCardDrawn),
	TimingSettleDiscard:      timingDescriptor(TimingSettleDiscard, TimingCategorySettle, "Settle discard", true, true, false, TimingOnCardDiscarded, TimingOnCardPlayedOrRevealed),
	TimingSettleHandLimit:    timingDescriptor(TimingSettleHandLimit, TimingCategorySettle, "Settle hand limit", true, true, false),
	TimingMoraleLossCheck:    timingDescriptor(TimingMoraleLossCheck, TimingCategorySettle, "Morale loss check", true, true, false, TimingBeforeMoraleLoss),
	TimingMoraleLossApplied:  timingDescriptor(TimingMoraleLossApplied, TimingCategorySettle, "Morale loss applied", true, true, false, Timing("on_morale_loss_applied")),
	TimingGameEndCheck:       timingDescriptor(TimingGameEndCheck, TimingCategorySettle, "Game end check", true, true, false),
	TimingDamageResolved:     timingDescriptor(TimingDamageResolved, TimingCategoryDamage, "Damage resolved", true, true, false, Timing("post_damage_resolved"), Timing("on_damage_after_apply")),

	// Role hook timings. These string identifiers currently live in
	// internal/engine/player and are mirrored here to keep model independent.
	Timing("post_action_end"):      roleTimingDescriptor("post_action_end", TimingCategoryTurn, "Post action end", TimingActionPost),
	Timing("post_attack_hit"):      roleTimingDescriptor("post_attack_hit", TimingCategoryAttack, "Post attack hit"),
	Timing("post_damage_resolved"): roleTimingDescriptor("post_damage_resolved", TimingCategoryDamage, "Post damage resolved"),

	Timing("on_turn_before_start"): roleTimingDescriptor("on_turn_before_start", TimingCategoryTurn, "Turn before start", TimingTurnBeforeStart),
	Timing("on_turn_start"):        roleTimingDescriptor("on_turn_start", TimingCategoryTurn, "Turn start", TimingOnTurnStart, TimingTurnStart),
	Timing("on_turn_end"):          roleTimingDescriptor("on_turn_end", TimingCategoryTurn, "Turn end pre extra", TimingOnTurnEnd),
	Timing("on_turn_end_final"):    roleTimingDescriptor("on_turn_end_final", TimingCategoryTurn, "Turn end final", TimingOnTurnEnd),
	Timing("before_action"):        roleTimingDescriptor("before_action", TimingCategoryTurn, "Before action", TimingOnBeforeAction, TimingActionStart),
	Timing("on_action_end"):        roleTimingDescriptor("on_action_end", TimingCategoryTurn, "Action end role hook", TimingOnActionEnd, TimingActionEnd),

	Timing("on_attack_gating"):      roleTimingDescriptor("on_attack_gating", TimingCategoryAttack, "Attack gating"),
	Timing("on_attack_card_hook"):   roleTimingDescriptor("on_attack_card_hook", TimingCategoryAttack, "Attack card hook", Timing("on_attack_card_transform")),
	Timing("on_attack_state_reset"): roleTimingDescriptor("on_attack_state_reset", TimingCategoryAttack, "Attack state reset"),
	Timing("on_attack_target_ctx"):  roleTimingDescriptor("on_attack_target_ctx", TimingCategoryAttack, "Attack target context"),
	Timing("on_attack_miss"):        roleTimingDescriptor("on_attack_miss", TimingCategoryAttack, "Attack miss"),
	Timing("on_counter_policy"):     roleTimingDescriptor("on_counter_policy", TimingCategoryAttack, "Counter policy"),
	Timing("on_defend_validation"):  roleTimingDescriptor("on_defend_validation", TimingCategoryAttack, "Defend validation"),
	Timing("on_response_skill_aug"): roleTimingDescriptor("on_response_skill_aug", TimingCategoryAttack, "Response skill augment"),
	Timing("on_response_skill_normalize"): roleTimingDescriptor(
		"on_response_skill_normalize", TimingCategoryAttack, "Response skill normalize",
	),
	Timing("on_response_skill_advance"): roleTimingDescriptor("on_response_skill_advance", TimingCategoryAttack, "Response skill advance"),
	Timing("on_response_skill_skip"):    roleTimingDescriptor("on_response_skill_skip", TimingCategoryAttack, "Response skill skip"),
	Timing("on_combat_interaction"):     roleTimingDescriptor("on_combat_interaction", TimingCategoryAttack, "Combat interaction"),
	Timing("on_counter_card_policy"):    roleTimingDescriptor("on_counter_card_policy", TimingCategoryAttack, "Counter card policy"),
	Timing("on_counter_element_check"):  roleTimingDescriptor("on_counter_element_check", TimingCategoryAttack, "Counter element check"),
	Timing("on_counter_resolve"):        roleTimingDescriptor("on_counter_resolve", TimingCategoryAttack, "Counter resolve"),
	Timing("on_magic_missile_defend"):   roleTimingDescriptor("on_magic_missile_defend", TimingCategoryMagic, "Magic missile defend", TimingMagicMissileDefend),
	Timing("on_magic_missile_counter"):  roleTimingDescriptor("on_magic_missile_counter", TimingCategoryMagic, "Magic missile counter", TimingMagicMissileCounter),
	Timing("on_magic_missile_response_skill_aug"): roleTimingDescriptor(
		"on_magic_missile_response_skill_aug", TimingCategoryMagic, "Magic missile response skill augment", TimingMagicMissileResponseSkill,
	),

	Timing("on_damage_before_taken"): roleTimingDescriptor("on_damage_before_taken", TimingCategoryDamage, "Damage before taken"),
	Timing("on_damage_after_taken"):  roleTimingDescriptor("on_damage_after_taken", TimingCategoryDamage, "Damage after taken"),
	Timing("on_damage_applied"):      roleTimingDescriptor("on_damage_applied", TimingCategoryDamage, "Damage applied", TimingOnDamageApplied),
	Timing("on_damage_after_apply"):  roleTimingDescriptor("on_damage_after_apply", TimingCategoryDamage, "Damage after apply"),
	Timing("on_heal_resist"):         roleTimingDescriptor("on_heal_resist", TimingCategoryDamage, "Heal resist"),
	Timing("on_heal_cap_calculate"):  roleTimingDescriptor("on_heal_cap_calculate", TimingCategoryDamage, "Heal cap calculate"),

	Timing("on_game_start"):       roleTimingDescriptor("on_game_start", TimingCategorySystem, "Game start", TimingOnGameStart),
	Timing("on_player_added"):     roleTimingDescriptor("on_player_added", TimingCategorySystem, "Player added"),
	Timing("on_camp_changed"):     roleTimingDescriptor("on_camp_changed", TimingCategorySystem, "Camp changed", TimingOnCampChanged),
	Timing("on_player_setup"):     roleTimingDescriptor("on_player_setup", TimingCategorySystem, "Player setup"),
	Timing("on_camp_cup_changed"): roleTimingDescriptor("on_camp_cup_changed", TimingCategorySystem, "Camp cup changed"),

	Timing("on_morale_loss_applied"): roleTimingDescriptor("on_morale_loss_applied", TimingCategorySettle, "Morale loss applied", TimingBeforeMoraleLoss),

	Timing("before_action_option"):         roleTimingDescriptor("before_action_option", TimingCategoryTurn, "Before action option"),
	Timing("before_action_validation"):     roleTimingDescriptor("before_action_validation", TimingCategoryTurn, "Before action validation"),
	Timing("on_attack_declared_interrupt"): roleTimingDescriptor("on_attack_declared_interrupt", TimingCategoryAttack, "Attack declared interrupt"),
	Timing("on_combat_counter_card"):       roleTimingDescriptor("on_combat_counter_card", TimingCategoryAttack, "Combat counter card"),
	Timing("on_after_cannot_act"):          roleTimingDescriptor("on_after_cannot_act", TimingCategoryTurn, "After cannot act"),
	Timing("on_special_action_override"):   roleTimingDescriptor("on_special_action_override", TimingCategoryTurn, "Special action override"),
	Timing("on_special_action_post"):       roleTimingDescriptor("on_special_action_post", TimingCategoryTurn, "Special action post"),
	Timing("on_skill_post"):                roleTimingDescriptor("on_skill_post", TimingCategoryTurn, "Skill post", TimingOnSkillExecuted),
	Timing("on_attack_card_transform"):     roleTimingDescriptor("on_attack_card_transform", TimingCategoryAttack, "Attack card transform", Timing("on_attack_card_hook")),
}

func timingDescriptor(id Timing, category TimingCategory, ruleName string, skillVisible, roleHookVisible, runtimeOnly bool, _ ...Timing) TimingDescriptor {
	return TimingDescriptor{
		ID:              id,
		Category:        category,
		RuleName:        ruleName,
		SkillVisible:    skillVisible,
		RoleHookVisible: roleHookVisible,
		RuntimeOnly:     runtimeOnly,
	}
}

func roleTimingDescriptor(id string, category TimingCategory, ruleName string, aliases ...Timing) TimingDescriptor {
	return timingDescriptor(Timing(id), category, ruleName, false, true, false, aliases...)
}

// TimingDescriptorOf returns registry metadata for a known timing.
func TimingDescriptorOf(t Timing) (TimingDescriptor, bool) {
	desc, ok := timingDescriptors[t]
	return desc, ok
}

// TimingCategoryOf returns the rule timeline category for a known timing.
func TimingCategoryOf(t Timing) TimingCategory {
	desc, ok := TimingDescriptorOf(t)
	if !ok {
		return TimingCategoryUnknown
	}
	return desc.Category
}
