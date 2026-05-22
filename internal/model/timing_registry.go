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
	Legacy          []Timing
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
	TimingOnAttackDeclared:       timingDescriptor(TimingOnAttackDeclared, TimingCategoryAttack, "Attack declared", true, false, false, Timing("on_attack_declared")),
	TimingOnMagicDeclared:        timingDescriptor(TimingOnMagicDeclared, TimingCategoryMagic, "Magic declared", true, false, false),
	TimingOnHitCheck:             timingDescriptor(TimingOnHitCheck, TimingCategoryAttack, "Attack hit check", true, false, false, Timing("on_hit_check")),
	TimingOnDamageCalculated:     timingDescriptor(TimingOnDamageCalculated, TimingCategoryDamage, "Damage calculated", true, false, false, Timing("on_damage_calculate")),
	TimingOnDamageApplied:        timingDescriptor(TimingOnDamageApplied, TimingCategoryDamage, "Damage applied", true, false, false, Timing("on_damage_applied")),
	TimingOnDamageTaken:          timingDescriptor(TimingOnDamageTaken, TimingCategoryDamage, "Damage taken", true, false, false, Timing("on_damage_taken")),
	TimingBeforeMoraleLoss:       timingDescriptor(TimingBeforeMoraleLoss, TimingCategorySettle, "Before morale loss", true, false, false, Timing("on_morale_loss_applied")),
	TimingBeforeCardDrawn:        timingDescriptor(TimingBeforeCardDrawn, TimingCategorySettle, "Before card drawn", true, false, false),
	TimingOnCardDrawn:            timingDescriptor(TimingOnCardDrawn, TimingCategorySettle, "Card drawn", true, false, false),
	TimingOnCardDiscarded:        timingDescriptor(TimingOnCardDiscarded, TimingCategorySettle, "Card discarded", true, false, false),
	TimingOnCardPlayedOrRevealed: timingDescriptor(TimingOnCardPlayedOrRevealed, TimingCategorySettle, "Card played or revealed", true, false, false),
	TimingOnHealOverflow:         timingDescriptor(TimingOnHealOverflow, TimingCategoryMagic, "Heal overflow", true, false, false),
	TimingOnFieldMarkChanged:     timingDescriptor(TimingOnFieldMarkChanged, TimingCategorySystem, "Field mark changed", true, false, false),
	TimingOnOrientationChanged:   timingDescriptor(TimingOnOrientationChanged, TimingCategorySystem, "Orientation changed", true, false, false),
	TimingOnTurnEnd:              timingDescriptor(TimingOnTurnEnd, TimingCategoryTurn, "Turn end", true, false, false, Timing("on_turn_end"), Timing("on_turn_end_final")),

	// Role hook timings. These string identifiers currently live in
	// internal/engine/player and are mirrored here to keep model independent.
	Timing("post_action_end"):      roleTimingDescriptor("post_action_end", TimingCategoryTurn, "Post action end"),
	Timing("post_attack_hit"):      roleTimingDescriptor("post_attack_hit", TimingCategoryAttack, "Post attack hit"),
	Timing("post_damage_resolved"): roleTimingDescriptor("post_damage_resolved", TimingCategoryDamage, "Post damage resolved"),

	Timing("on_turn_before_start"): roleTimingDescriptor("on_turn_before_start", TimingCategoryTurn, "Turn before start"),
	Timing("on_turn_start"):        roleTimingDescriptor("on_turn_start", TimingCategoryTurn, "Turn start", TimingOnTurnStart),
	Timing("on_turn_end"):          roleTimingDescriptor("on_turn_end", TimingCategoryTurn, "Turn end pre extra", TimingOnTurnEnd),
	Timing("on_turn_end_final"):    roleTimingDescriptor("on_turn_end_final", TimingCategoryTurn, "Turn end final", TimingOnTurnEnd),
	Timing("before_action"):        roleTimingDescriptor("before_action", TimingCategoryTurn, "Before action", TimingOnBeforeAction),
	Timing("on_action_end"):        roleTimingDescriptor("on_action_end", TimingCategoryTurn, "Action end role hook", TimingOnActionEnd),

	Timing("on_attack_declared"):    roleTimingDescriptor("on_attack_declared", TimingCategoryAttack, "Attack declared", TimingOnAttackDeclared),
	Timing("on_attack_gating"):      roleTimingDescriptor("on_attack_gating", TimingCategoryAttack, "Attack gating"),
	Timing("on_attack_card_hook"):   roleTimingDescriptor("on_attack_card_hook", TimingCategoryAttack, "Legacy attack card hook", Timing("on_attack_card_transform")),
	Timing("on_attack_state_reset"): roleTimingDescriptor("on_attack_state_reset", TimingCategoryAttack, "Attack state reset"),
	Timing("on_attack_target_ctx"):  roleTimingDescriptor("on_attack_target_ctx", TimingCategoryAttack, "Attack target context"),
	Timing("on_attack_miss"):        roleTimingDescriptor("on_attack_miss", TimingCategoryAttack, "Attack miss"),
	Timing("on_hit_check"):          roleTimingDescriptor("on_hit_check", TimingCategoryAttack, "Attack hit check", TimingOnHitCheck),
	Timing("on_counter_policy"):     roleTimingDescriptor("on_counter_policy", TimingCategoryAttack, "Legacy counter policy"),
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
	Timing("on_magic_missile_defend"):   roleTimingDescriptor("on_magic_missile_defend", TimingCategoryMagic, "Magic missile defend"),
	Timing("on_magic_missile_counter"):  roleTimingDescriptor("on_magic_missile_counter", TimingCategoryMagic, "Magic missile counter"),
	Timing("on_magic_missile_response_skill_aug"): roleTimingDescriptor(
		"on_magic_missile_response_skill_aug", TimingCategoryMagic, "Magic missile response skill augment",
	),

	Timing("on_damage_calculate"):    roleTimingDescriptor("on_damage_calculate", TimingCategoryDamage, "Damage calculate", TimingOnDamageCalculated),
	Timing("on_damage_before_taken"): roleTimingDescriptor("on_damage_before_taken", TimingCategoryDamage, "Damage before taken"),
	Timing("on_damage_after_taken"):  roleTimingDescriptor("on_damage_after_taken", TimingCategoryDamage, "Damage after taken"),
	Timing("on_damage_applied"):      roleTimingDescriptor("on_damage_applied", TimingCategoryDamage, "Damage applied", TimingOnDamageApplied),
	Timing("on_damage_taken"):        roleTimingDescriptor("on_damage_taken", TimingCategoryDamage, "Damage taken", TimingOnDamageTaken),
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

func timingDescriptor(id Timing, category TimingCategory, ruleName string, skillVisible, roleHookVisible, runtimeOnly bool, legacy ...Timing) TimingDescriptor {
	return TimingDescriptor{
		ID:              id,
		Category:        category,
		RuleName:        ruleName,
		SkillVisible:    skillVisible,
		RoleHookVisible: roleHookVisible,
		RuntimeOnly:     runtimeOnly,
		Legacy:          append([]Timing(nil), legacy...),
	}
}

func roleTimingDescriptor(id string, category TimingCategory, ruleName string, legacy ...Timing) TimingDescriptor {
	return timingDescriptor(Timing(id), category, ruleName, false, true, false, legacy...)
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
