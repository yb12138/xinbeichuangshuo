package server

import "starcup-engine/internal/model"

func countMagicBowChargesByElement(p *model.Player, element model.Element) int {
	if p == nil {
		return 0
	}
	count := 0
	for _, fc := range p.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectMagicBowCharge {
			continue
		}
		if element != "" && fc.Card.Element != element {
			continue
		}
		count++
	}
	return count
}

func countMagicBowCharges(p *model.Player) int {
	return countMagicBowChargesByElement(p, "")
}

func countSpiritCasterPowers(p *model.Player) int {
	if p == nil {
		return 0
	}
	count := 0
	for _, fc := range p.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectSpiritCasterPower {
			continue
		}
		count++
	}
	return count
}

func countMoonDarkMoons(p *model.Player) int {
	if p == nil {
		return 0
	}
	count := 0
	for _, fc := range p.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectMoonDarkMoon {
			continue
		}
		count++
	}
	return count
}

func countButterflyCocoons(p *model.Player) int {
	if p == nil {
		return 0
	}
	count := 0
	for _, fc := range p.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectButterflyCocoon {
			continue
		}
		count++
	}
	return count
}

func countBloodSharedLifeAsSource(state *model.GameState, sourceID string) int {
	if state == nil || sourceID == "" {
		return 0
	}
	count := 0
	for _, p := range state.Players {
		if p == nil {
			continue
		}
		for _, fc := range p.Field {
			if fc == nil || fc.Mode != model.FieldEffect || fc.Effect != model.EffectBloodSharedLife {
				continue
			}
			if fc.SourceID == sourceID {
				count++
			}
		}
	}
	return count
}

func countBloodSharedLifeAsHolder(player *model.Player) int {
	if player == nil {
		return 0
	}
	count := 0
	for _, fc := range player.Field {
		if fc == nil || fc.Mode != model.FieldEffect || fc.Effect != model.EffectBloodSharedLife {
			continue
		}
		count++
	}
	return count
}

// 当前手牌上限仅用于展示，避免直接对引擎内玩家对象调用 GetMaxHand 带来状态副作用。
func (r *Room) previewMaxHand(player *model.Player) int {
	if player == nil {
		return 0
	}
	playerCopy := *player
	if player.Tokens != nil {
		playerCopy.Tokens = make(map[string]int, len(player.Tokens))
		for k, v := range player.Tokens {
			playerCopy.Tokens[k] = v
		}
	}
	return r.Engine.GetMaxHand(&playerCopy)
}

func buildMaskedFieldForViewer(owner *model.Player, viewerID string) []*model.FieldCard {
	if owner == nil || len(owner.Field) == 0 {
		return nil
	}
	out := make([]*model.FieldCard, 0, len(owner.Field))
	for _, fc := range owner.Field {
		if fc == nil {
			continue
		}
		clone := *fc
		// 魔弓“充能”、灵符师“妖力”、月之女神“暗月”、蝶舞者“茧”对非持有者隐藏具体牌面信息，仅保留数量与盖牌属性。
		if owner.ID != viewerID && clone.Mode == model.FieldCover &&
			(clone.Effect == model.EffectMagicBowCharge || clone.Effect == model.EffectSpiritCasterPower || clone.Effect == model.EffectMoonDarkMoon || clone.Effect == model.EffectButterflyCocoon) {
			maskedName := "盖牌"
			if clone.Effect == model.EffectMagicBowCharge {
				maskedName = "充能"
			} else if clone.Effect == model.EffectSpiritCasterPower {
				maskedName = "妖力"
			} else if clone.Effect == model.EffectMoonDarkMoon {
				maskedName = "暗月"
			} else if clone.Effect == model.EffectButterflyCocoon {
				maskedName = "茧"
			}
			clone.Card = model.Card{
				ID:          clone.Card.ID,
				Name:        maskedName,
				Type:        clone.Card.Type,
				Description: "盖牌（内容对他人不可见）",
			}
		}
		out = append(out, &clone)
	}
	return out
}

func (r *Room) buildStateForPlayer(playerID string) GameStateUpdate {
	state := r.Engine.State

	players := make(map[string]PlayerView)
	for pid, p := range state.Players {
		view := PlayerView{
			ID:                 p.ID,
			Name:               p.Name,
			Camp:               string(p.Camp),
			Role:               p.Role,
			Orientation:        string(r.Engine.GetPlayerOrientation(pid)),
			Form:               r.Engine.GetPlayerForm(pid),
			HandCount:          len(p.Hand),
			MaxHand:            r.previewMaxHand(p),
			ExclusiveCardCount: len(p.ExclusiveCards),
			Field:              buildMaskedFieldForViewer(p, playerID),
			Heal:               p.Heal,
			MaxHeal:            p.MaxHeal,
			Gem:                p.Gem,
			Crystal:            p.Crystal,
			IsActive:           p.IsActive,
			Buffs:              p.Buffs,
			Tokens:             map[string]int{},
		}
		for k, v := range p.Tokens {
			view.Tokens[k] = v
		}
		delete(view.Tokens, "adventurer_extract_last_gem")
		delete(view.Tokens, "adventurer_extract_last_crystal")
		delete(view.Tokens, "mg_moon_cycle_used_turn")
		delete(view.Tokens, "prayer_form")
		delete(view.Tokens, "crk_hot_form")
		delete(view.Tokens, "onmyoji_form")
		delete(view.Tokens, "bw_flame_form")
		delete(view.Tokens, "hb_form")
		delete(view.Tokens, "ml_phantom_form")
		delete(view.Tokens, "bd_prisoner_form")
		delete(view.Tokens, "hero_exhaustion_form")
		delete(view.Tokens, "fighter_hundred_dragon_form")
		delete(view.Tokens, "mg_dark_form")
		delete(view.Tokens, "bp_bleed_form")
		delete(view.Tokens, "arbiter_form")
		delete(view.Tokens, "elf_ritual_form")
		delete(view.Tokens, "ms_shadow_form")
		delete(view.Tokens, "hom_burst_form")
		// 魔枪公开状态：下次主动攻击加成/当回合互斥锁，便于前端角色面板展示。
		if p.TurnState.UsedSkillCounts != nil {
			if v := p.TurnState.UsedSkillCounts["ml_dark_release_next_attack_bonus"]; v > 0 {
				view.Tokens["ml_dark_release_next_attack_bonus"] = v
			}
			if v := p.TurnState.UsedSkillCounts["ml_fullness_next_attack_bonus"]; v > 0 {
				view.Tokens["ml_fullness_next_attack_bonus"] = v
			}
			if v := p.TurnState.UsedSkillCounts["ml_dark_release_lock_turn"]; v > 0 {
				view.Tokens["ml_dark_release_lock_turn"] = v
			}
		}
		// 精灵射手祝福数量（独立牌区）透出给前端做指示物展示。
		blessings := len(p.Blessings)
		if blessings > 0 {
			view.Tokens["elf_blessing_count"] = blessings
		}
		// 魔弓充能数量（盖牌内容对他人隐藏，仅展示数量）。
		chargeCount := countMagicBowCharges(p)
		if chargeCount > 0 {
			view.Tokens["mb_charge_count"] = chargeCount
		} else {
			delete(view.Tokens, "mb_charge_count")
		}
		// 灵符师妖力数量（盖牌内容对他人隐藏，仅展示数量）。
		powerCount := countSpiritCasterPowers(p)
		if powerCount > 0 {
			view.Tokens["sc_power_count"] = powerCount
		} else {
			delete(view.Tokens, "sc_power_count")
		}
		// 月之女神暗月数量（盖牌内容对他人隐藏，仅展示数量）。
		darkMoonCount := countMoonDarkMoons(p)
		if darkMoonCount > 0 {
			view.Tokens["mg_dark_moon_count"] = darkMoonCount
		} else {
			delete(view.Tokens, "mg_dark_moon_count")
		}
		// 血之巫女同生共死：仅展示是否存在激活中的连结。
		sharedLifeCount := countBloodSharedLifeAsSource(state, p.ID)
		if sharedLifeCount > 0 {
			view.Tokens["bp_shared_life_active"] = sharedLifeCount
		} else {
			delete(view.Tokens, "bp_shared_life_active")
		}
		sharedLifeBoundCount := countBloodSharedLifeAsHolder(p)
		if sharedLifeBoundCount > 0 {
			view.Tokens["bp_shared_life_bound"] = sharedLifeBoundCount
		} else {
			delete(view.Tokens, "bp_shared_life_bound")
		}
		// 蝶舞者茧数量（盖牌内容对他人隐藏，仅展示数量）。
		cocoonCount := countButterflyCocoons(p)
		if cocoonCount > 0 {
			view.Tokens["bt_cocoon_count"] = cocoonCount
		} else {
			delete(view.Tokens, "bt_cocoon_count")
		}
		// 仅自己可见手牌具体内容，他人只能看到数量
		if pid == playerID {
			view.Hand = p.Hand
			view.Blessings = p.Blessings
			view.ExclusiveCards = p.ExclusiveCards
		}
		players[pid] = view
	}

	var availableSkills []AvailableSkill
	if state.TurnStage == model.TurnStageActionExecution && len(state.ActionQueue) == 0 && state.Subflow == model.SubflowNone && len(state.CombatStack) == 0 {
		if self := state.Players[playerID]; self != nil && self.IsActive {
			availableSkills = r.buildAvailableActionSkills(playerID)
		}
	}

	return GameStateUpdate{
		TurnStage:           string(state.TurnStage),
		CombatStage:         string(state.CombatStage),
		Subflow:             string(state.Subflow),
		CurrentPlayer:       state.CurrentPlayer,
		HasPerformedStartup: state.HasPerformedStartup,
		Players:             players,
		RedMorale:           state.RedMorale,
		BlueMorale:          state.BlueMorale,
		RedCups:             state.RedCups,
		BlueCups:            state.BlueCups,
		RedGems:             state.RedGems,
		BlueGems:            state.BlueGems,
		RedCrystals:         state.RedCrystals,
		BlueCrystals:        state.BlueCrystals,
		DeckCount:           len(state.Deck),
		DiscardCount:        len(state.DiscardPile),
		AvailableSkills:     availableSkills,
		Characters:          buildCharacterViews(),
	}
}
