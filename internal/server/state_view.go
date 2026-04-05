package server

import (
	"starcup-engine/internal/model"
	"starcup-engine/internal/server/stateview"
)

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
	return stateview.BuildMaskedFieldForViewer(owner, viewerID)
}

func hasRuleModifierWithModifierID(p *model.Player, modifierID string) bool {
	return stateview.HasRuleModifierWithModifierID(p, modifierID)
}

func combatPolicyAttackBonusByModifierID(p *model.Player, modifierID string) int {
	return stateview.CombatPolicyAttackBonusByModifierID(p, modifierID)
}

func countMagicBowChargesByElement(p *model.Player, element model.Element) int {
	return stateview.CountMagicBowChargesByElement(p, element)
}

func countMagicBowCharges(p *model.Player) int {
	return stateview.CountMagicBowCharges(p)
}

func countSpiritCasterPowers(p *model.Player) int {
	return stateview.CountSpiritCasterPowers(p)
}

func countMoonDarkMoons(p *model.Player) int {
	return stateview.CountMoonDarkMoons(p)
}

func countButterflyCocoons(p *model.Player) int {
	return stateview.CountButterflyCocoons(p)
}

func countBloodSharedLifeAsSource(state *model.GameState, sourceID string) int {
	return stateview.CountBloodSharedLifeAsSource(state, sourceID)
}

func countBloodSharedLifeAsHolder(player *model.Player) int {
	return stateview.CountBloodSharedLifeAsHolder(player)
}

func (r *Room) buildStateForPlayer(playerID string) GameStateUpdate {
	state := r.Engine.State
	hasPerformedStartup := false
	if state != nil && len(state.PlayerOrder) > 0 && state.CurrentTurn >= 0 && state.CurrentTurn < len(state.PlayerOrder) {
		currentPlayer := state.Players[state.PlayerOrder[state.CurrentTurn]]
		if currentPlayer != nil {
			hasPerformedStartup = currentPlayer.TurnState.HasStartupSkillOrSpecialActionsLocked()
		}
	}

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
		// UI 派生计数 → PlayerView 显式字段（真源在 Blessings/Field/RuleModifiers 上）
		view.ElfBlessingCount = len(p.Blessings)
		view.MagicBowChargeCount = countMagicBowCharges(p)
		view.SpiritCasterPowerCount = countSpiritCasterPowers(p)
		view.MoonDarkMoonCount = countMoonDarkMoons(p)
		view.ButterflyCocoonCount = countButterflyCocoons(p)
		view.BloodSharedLifeActive = countBloodSharedLifeAsSource(state, p.ID)
		view.BloodSharedLifeBound = countBloodSharedLifeAsHolder(p)
		view.MagicLancerDarkReleaseBonus = combatPolicyAttackBonusByModifierID(p, "ml_dark_release_next_attack_bonus")
		view.MagicLancerFullnessBonus = combatPolicyAttackBonusByModifierID(p, "ml_fullness_next_attack_bonus")
		if hasRuleModifierWithModifierID(p, "ml_dark_release_lock_turn") {
			view.MagicLancerDarkReleaseLockTurn = 1
		} else {
			view.MagicLancerDarkReleaseLockTurn = 0
		}
		// 清理不应暴露给前端的 Tokens 镜像
		delete(view.Tokens, "mb_charge_count")
		delete(view.Tokens, "sc_power_count")
		delete(view.Tokens, "mg_dark_moon_count")
		delete(view.Tokens, "bt_cocoon_count")
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
		HasPerformedStartup: hasPerformedStartup,
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
