// gameflow: RuleModifier 实例的生效与查询。

package engine

import "starcup-engine/internal/model"

func ensurePlayerRuleModifiersMap(player *model.Player) {
	if player != nil && player.ActiveRuleModifiers == nil {
		player.ActiveRuleModifiers = map[string]*model.RuleModifierInstance{}
	}
}

func buildRuleModifierInstanceID(modifierID, sourceSkillID string) string {
	if sourceSkillID == "" {
		return modifierID
	}
	return sourceSkillID + ":" + modifierID
}

func ensureCombatPolicyPayload(modifier *model.RuleModifierInstance) *model.RuleCombatPolicyPayload {
	if modifier == nil {
		return nil
	}
	if modifier.CombatPolicyPayload == nil {
		modifier.CombatPolicyPayload = &model.RuleCombatPolicyPayload{}
	}
	return modifier.CombatPolicyPayload
}

func (e *GameEngine) ApplySkillGateRule(playerID string, modifierID string, sourceSkillID string, skillIDs []string, lifetime model.RuleModifierLifetimeType) {
	if e == nil || e.State == nil || playerID == "" || modifierID == "" {
		return
	}
	player := e.State.Players[playerID]
	if player == nil {
		return
	}
	ensurePlayerRuleModifiersMap(player)
	instanceID := buildRuleModifierInstanceID(modifierID, sourceSkillID)
	blockedSkills := append([]string(nil), skillIDs...)
	player.ActiveRuleModifiers[instanceID] = &model.RuleModifierInstance{
		RuleInstanceID: instanceID,
		ModifierID:     modifierID,
		SourceUserID:   playerID,
		SourceSkillID:  sourceSkillID,
		TargetUserID:   playerID,
		Domain:         model.RuleModifierDomainSkillGate,
		Lifetime:       lifetime,
		SkillGatePayload: &model.RuleSkillGatePayload{
			Mode:     model.SkillGateDisallowList,
			SkillIDs: blockedSkills,
		},
	}
}

func (e *GameEngine) ApplyNextAttackDamageRule(playerID string, modifierID string, sourceSkillID string, bonus int, lifetime model.RuleModifierLifetimeType) {
	if e == nil || e.State == nil || playerID == "" || modifierID == "" || bonus == 0 {
		return
	}
	player := e.State.Players[playerID]
	if player == nil {
		return
	}
	ensurePlayerRuleModifiersMap(player)
	instanceID := buildRuleModifierInstanceID(modifierID, sourceSkillID)
	if existing := player.ActiveRuleModifiers[instanceID]; existing != nil &&
		existing.Domain == model.RuleModifierDomainCombatPolicy {
		payload := ensureCombatPolicyPayload(existing)
		payload.AttackDamageBonus += bonus
		payload.OnlyActionType = model.ActionAttack
		payload.RequireActiveOnly = true
		payload.ConsumeOnMatch = true
		existing.Lifetime = lifetime
		return
	}
	player.ActiveRuleModifiers[instanceID] = &model.RuleModifierInstance{
		RuleInstanceID: instanceID,
		ModifierID:     modifierID,
		SourceUserID:   playerID,
		SourceSkillID:  sourceSkillID,
		TargetUserID:   playerID,
		Domain:         model.RuleModifierDomainCombatPolicy,
		Lifetime:       lifetime,
		CombatPolicyPayload: &model.RuleCombatPolicyPayload{
			AttackDamageBonus: bonus,
			OnlyActionType:    model.ActionAttack,
			RequireActiveOnly: true,
			ConsumeOnMatch:    true,
		},
	}
}

func (e *GameEngine) ApplyNextAttackInterceptTagRule(playerID string, modifierID string, sourceSkillID string, tag model.CombatInterceptTag, lifetime model.RuleModifierLifetimeType) {
	if e == nil || e.State == nil || playerID == "" || modifierID == "" || tag == model.CombatInterceptNone {
		return
	}
	player := e.State.Players[playerID]
	if player == nil {
		return
	}
	ensurePlayerRuleModifiersMap(player)
	instanceID := buildRuleModifierInstanceID(modifierID, sourceSkillID)
	if existing := player.ActiveRuleModifiers[instanceID]; existing != nil &&
		existing.Domain == model.RuleModifierDomainCombatPolicy {
		payload := ensureCombatPolicyPayload(existing)
		if payload.InterceptTags == nil {
			payload.InterceptTags = map[model.CombatInterceptTag]bool{}
		}
		payload.InterceptTags[tag] = true
		payload.OnlyActionType = model.ActionAttack
		payload.RequireActiveOnly = true
		payload.ConsumeOnMatch = true
		existing.Lifetime = lifetime
		return
	}
	player.ActiveRuleModifiers[instanceID] = &model.RuleModifierInstance{
		RuleInstanceID: instanceID,
		ModifierID:     modifierID,
		SourceUserID:   playerID,
		SourceSkillID:  sourceSkillID,
		TargetUserID:   playerID,
		Domain:         model.RuleModifierDomainCombatPolicy,
		Lifetime:       lifetime,
		CombatPolicyPayload: &model.RuleCombatPolicyPayload{
			OnlyActionType:    model.ActionAttack,
			RequireActiveOnly: true,
			ConsumeOnMatch:    true,
			InterceptTags: map[model.CombatInterceptTag]bool{
				tag: true,
			},
		},
	}
}

func (e *GameEngine) IsSkillBlocked(playerID string, skillID string) bool {
	if e == nil || e.State == nil || playerID == "" || skillID == "" {
		return false
	}
	player := e.State.Players[playerID]
	if player == nil || len(player.ActiveRuleModifiers) == 0 {
		return false
	}
	for _, modifier := range player.ActiveRuleModifiers {
		if modifier == nil || modifier.Domain != model.RuleModifierDomainSkillGate || modifier.SkillGatePayload == nil {
			continue
		}
		if modifier.SkillGatePayload.Mode != model.SkillGateDisallowList {
			continue
		}
		for _, blockedID := range modifier.SkillGatePayload.SkillIDs {
			if blockedID == skillID {
				return true
			}
		}
	}
	return false
}

func combatPolicyRuleMatchesAction(modifier *model.RuleModifierInstance, action model.Action) bool {
	if modifier == nil || modifier.Domain != model.RuleModifierDomainCombatPolicy || modifier.CombatPolicyPayload == nil {
		return false
	}
	payload := modifier.CombatPolicyPayload
	if payload.OnlyActionType != "" && action.Type != payload.OnlyActionType {
		return false
	}
	if payload.RequireActiveOnly && action.CounterInitiator != "" {
		return false
	}
	return true
}

func AttackDamageRuleBonusForModifier(player *model.Player, modifierID string) int {
	if player == nil || modifierID == "" || len(player.ActiveRuleModifiers) == 0 {
		return 0
	}
	total := 0
	for _, modifier := range player.ActiveRuleModifiers {
		if modifier == nil || modifier.ModifierID != modifierID || modifier.Domain != model.RuleModifierDomainCombatPolicy || modifier.CombatPolicyPayload == nil {
			continue
		}
		total += modifier.CombatPolicyPayload.AttackDamageBonus
	}
	return total
}

func consumeAttackDamageRuleBonus(player *model.Player, modifierID string, action model.Action) int {
	if player == nil || modifierID == "" || len(player.ActiveRuleModifiers) == 0 {
		return 0
	}
	total := 0
	for key, modifier := range player.ActiveRuleModifiers {
		if modifier == nil || modifier.ModifierID != modifierID || !combatPolicyRuleMatchesAction(modifier, action) {
			continue
		}
		total += modifier.CombatPolicyPayload.AttackDamageBonus
		if modifier.CombatPolicyPayload.ConsumeOnMatch {
			delete(player.ActiveRuleModifiers, key)
		}
	}
	if len(player.ActiveRuleModifiers) == 0 {
		player.ActiveRuleModifiers = nil
	}
	return total
}

func consumeAttackCombatPolicyInterceptTags(player *model.Player, action model.Action, attackInfo *model.AttackEventInfo) {
	if player == nil || attackInfo == nil || len(player.ActiveRuleModifiers) == 0 {
		return
	}
	for key, modifier := range player.ActiveRuleModifiers {
		if modifier == nil || !combatPolicyRuleMatchesAction(modifier, action) || modifier.CombatPolicyPayload == nil || len(modifier.CombatPolicyPayload.InterceptTags) == 0 {
			continue
		}
		for tag, enabled := range modifier.CombatPolicyPayload.InterceptTags {
			if enabled {
				attackInfo.SetInterceptTag(tag)
			}
		}
		if modifier.CombatPolicyPayload.ConsumeOnMatch {
			delete(player.ActiveRuleModifiers, key)
		}
	}
	if len(player.ActiveRuleModifiers) == 0 {
		player.ActiveRuleModifiers = nil
	}
}

func (e *GameEngine) expireRuleModifiersByLifetime(player *model.Player, lifetime model.RuleModifierLifetimeType) {
	if player == nil || len(player.ActiveRuleModifiers) == 0 {
		return
	}
	for key, modifier := range player.ActiveRuleModifiers {
		if modifier == nil || modifier.Lifetime == lifetime {
			delete(player.ActiveRuleModifiers, key)
		}
	}
	if len(player.ActiveRuleModifiers) == 0 {
		player.ActiveRuleModifiers = nil
	}
}
