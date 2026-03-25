package engine

import "starcup-engine/internal/model"

func (e *GameEngine) hasUsableActionSkillForExtraMagic(player *model.Player) bool {
	if player == nil || player.Character == nil {
		return false
	}

	for _, skillDef := range player.Character.Skills {
		if skillDef.Type != model.SkillTypeAction {
			continue
		}
		if e.isActionSkillUsableForExtraMagic(player, skillDef) {
			return true
		}
	}
	return false
}

func (e *GameEngine) isActionSkillUsableForExtraMagic(player *model.Player, skillDef model.SkillDefinition) bool {
	// 回合限定：本回合已用过则不可再用。
	if model.ContainsSkillTag(skillDef.Tags, model.TagTurnLimit) && player.TurnState.UsedSkillCounts[skillDef.ID] > 0 {
		return false
	}
	// 资源校验（宝石/水晶）。
	if !canPaySkillEnergyCost(player, skillDef.CostGem, skillDef.CostCrystal) {
		return false
	}
	// 独有技：需拥有对应独有牌（手牌或专属卡区）。
	if skillDef.RequireExclusive && !player.HasExclusiveCard(player.Character.Name, skillDef.Title) {
		return false
	}
	// 弃牌成本可达成性。
	if !e.canSatisfyActionSkillDiscardRequirement(player, skillDef) {
		return false
	}
	// 目标可达成性（仅做最小目标数校验）。
	if !e.hasActionSkillValidTarget(player, skillDef) {
		return false
	}

	// 与前端/房间可用技能筛选保持一致的技能特例。
	switch skillDef.ID {
	case "ms_shadow_meteor":
		// 魔剑士【暗影流星】需要处于暗影形态，且至少可弃2张法术牌。
		if !hasMagicSwordsmanShadowForm(player) {
			return false
		}
		magicCount := 0
		for _, card := range player.Hand {
			if card.Type == model.CardTypeMagic {
				magicCount++
			}
		}
		if magicCount < 2 {
			return false
		}
	case "adventurer_fraud":
		elementCount := map[model.Element]int{}
		for _, card := range player.Hand {
			elementCount[card.Element]++
		}
		canUseFraud := false
		for element, count := range elementCount {
			if element != "" && count >= 2 {
				canUseFraud = true
				break
			}
			if count >= 3 {
				canUseFraud = true
				break
			}
		}
		if !canUseFraud {
			return false
		}
	case "onmyoji_shikigami_descend":
		factionCount := map[string]int{}
		hasSameFactionPair := false
		for _, card := range player.Hand {
			if card.Faction == "" {
				continue
			}
			factionCount[card.Faction]++
			if factionCount[card.Faction] >= 2 {
				hasSameFactionPair = true
				break
			}
		}
		if !hasSameFactionPair {
			return false
		}
	case "mb_thunder_scatter":
		if player.TurnState.UsedSkillCounts["mb_charge_lock_turn"] > 0 {
			return false
		}
		if e.countCoverCardsByEffectAndElement(player, model.EffectMagicBowCharge, model.ElementThunder) <= 0 {
			return false
		}
	case "bd_dissonance_chord":
		inspiration := 0
		if player.Tokens != nil {
			inspiration = player.Tokens["bd_inspiration"]
		}
		if inspiration <= 1 {
			return false
		}
	case "elementalist_ignite":
		element := 0
		if player.Tokens != nil {
			element = player.Tokens["element"]
		}
		if element < 3 {
			return false
		}
	case "angel_cleanse":
		if !e.hasAnyBasicFieldEffectTarget() {
			return false
		}
	case "holy_lancer_punishment":
		hasValidTarget := false
		for _, playerID := range e.State.PlayerOrder {
			target := e.State.Players[playerID]
			if target == nil || target.ID == player.ID || target.Heal <= 0 {
				continue
			}
			hasValidTarget = true
			break
		}
		if !hasValidTarget {
			return false
		}
	}

	return true
}

func (e *GameEngine) canSatisfyActionSkillDiscardRequirement(player *model.Player, skillDef model.SkillDefinition) bool {
	if skillDef.ID == "priest_water_power" {
		if len(player.Hand) < 2 {
			return false
		}
		for _, card := range player.Hand {
			if card.Element == model.ElementWater {
				return true
			}
		}
		return false
	}

	requiredDiscards := skillDef.CostDiscards
	if requiredDiscards <= 0 {
		return true
	}

	matched := 0
	for _, card := range player.Hand {
		effectiveElement := card.Element
		if skillDef.DiscardElement != "" {
			effectiveElement = e.blazeWitchAttackElement(player, card)
		}
		if skillDef.DiscardElement != "" && effectiveElement != skillDef.DiscardElement {
			continue
		}
		if skillDef.DiscardType != "" && card.Type != skillDef.DiscardType {
			continue
		}
		if skillDef.DiscardFate != "" && card.Faction != skillDef.DiscardFate {
			continue
		}
		if skillDef.RequireExclusive && !card.MatchExclusive(player.Character.Name, skillDef.Title) {
			continue
		}
		matched++
		if matched >= requiredDiscards {
			return true
		}
	}
	return false
}

func (e *GameEngine) hasActionSkillValidTarget(player *model.Player, skillDef model.SkillDefinition) bool {
	switch skillDef.TargetType {
	case model.TargetNone, model.TargetSelf:
		return true
	}

	minTargets := skillDef.MinTargets
	if minTargets <= 0 {
		if skillDef.TargetType >= model.TargetEnemy {
			minTargets = 1
		} else {
			minTargets = 0
		}
	}
	if minTargets <= 0 {
		return true
	}

	candidates := 0
	for _, playerID := range e.State.PlayerOrder {
		target := e.State.Players[playerID]
		if target == nil {
			continue
		}
		switch skillDef.TargetType {
		case model.TargetEnemy:
			if target.Camp != player.Camp {
				candidates++
			}
		case model.TargetAlly:
			if target.Camp == player.Camp && target.ID != player.ID {
				candidates++
			}
		case model.TargetAllySelf:
			if target.Camp == player.Camp {
				candidates++
			}
		case model.TargetAny, model.TargetSpecific:
			candidates++
		}
		if candidates >= minTargets {
			return true
		}
	}
	return false
}

func (e *GameEngine) hasAnyBasicFieldEffectTarget() bool {
	isBasicEffect := func(effect model.EffectType) bool {
		switch effect {
		case model.EffectShield,
			model.EffectWeak,
			model.EffectPoison,
			model.EffectSealFire,
			model.EffectSealWater,
			model.EffectSealEarth,
			model.EffectSealWind,
			model.EffectSealThunder,
			model.EffectPowerBlessing,
			model.EffectSwiftBlessing:
			return true
		default:
			return false
		}
	}

	for _, player := range e.State.Players {
		if player == nil {
			continue
		}
		for _, fieldCard := range player.Field {
			if fieldCard.Mode == model.FieldEffect && isBasicEffect(fieldCard.Effect) {
				return true
			}
		}
	}
	return false
}
