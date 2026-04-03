package model

type RuleModifierDomain int

const (
	RuleModifierDomainAttribute RuleModifierDomain = iota
	RuleModifierDomainHealPolicy
	RuleModifierDomainSkillGate
	RuleModifierDomainCombatPolicy
	RuleModifierDomainCardSource
	RuleModifierDomainTokenPolicy
	RuleModifierDomainHealResistPolicy
	RuleModifierDomainMoralePolicy
)

type RuleModifierLifetimeType int

const (
	RuleLifeThisEffectChain RuleModifierLifetimeType = iota
	RuleLifeUntilTurnEnd
	RuleLifeUntilSourceNextTurnStart
	RuleLifeUntilSourceNextTurnEnd
	RuleLifePermanent
	RuleLifeUntilCombatEnd
)

type SkillGateMode int

const (
	SkillGateDisallowList SkillGateMode = iota
)

type RuleSkillGatePayload struct {
	Mode     SkillGateMode `json:"mode"`
	SkillIDs []string      `json:"skill_ids,omitempty"`
}

type RuleCombatPolicyPayload struct {
	AttackDamageBonus int                         `json:"attack_damage_bonus,omitempty"`
	OnlyActionType    ActionType                  `json:"only_action_type,omitempty"`
	RequireActiveOnly bool                        `json:"require_active_only,omitempty"`
	ConsumeOnMatch    bool                        `json:"consume_on_match,omitempty"`
	InterceptTags     map[CombatInterceptTag]bool `json:"intercept_tags,omitempty"`
}

type RuleModifierInstance struct {
	RuleInstanceID string                   `json:"rule_instance_id"`
	ModifierID     string                   `json:"modifier_id"`
	SourceUserID   string                   `json:"source_user_id"`
	SourceSkillID  string                   `json:"source_skill_id"`
	TargetUserID   string                   `json:"target_user_id"`
	Domain         RuleModifierDomain       `json:"domain"`
	Lifetime       RuleModifierLifetimeType `json:"lifetime"`

	SkillGatePayload    *RuleSkillGatePayload    `json:"skill_gate_payload,omitempty"`
	CombatPolicyPayload *RuleCombatPolicyPayload `json:"combat_policy_payload,omitempty"`
}
