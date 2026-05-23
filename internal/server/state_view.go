package server

import (
	"starcup-engine/internal/model"
	"starcup-engine/internal/server/bot"
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

func countElfBlessings(p *model.Player) int {
	return stateview.CountElfBlessings(p)
}

func countSwordEmperorSwordSouls(p *model.Player) int {
	return stateview.CountSwordEmperorSwordSouls(p)
}

func countBloodSharedLifeAsSource(state *model.GameState, sourceID string) int {
	return stateview.CountBloodSharedLifeAsSource(state, sourceID)
}

func countBloodSharedLifeAsHolder(player *model.Player) int {
	return stateview.CountBloodSharedLifeAsHolder(player)
}

func currentTurnPlayerID(state *model.GameState) string {
	if state == nil {
		return ""
	}
	if len(state.PlayerOrder) > 0 && state.CurrentTurn >= 0 && state.CurrentTurn < len(state.PlayerOrder) {
		return state.PlayerOrder[state.CurrentTurn]
	}
	if state.CurrentPlayer != "" {
		return state.CurrentPlayer
	}
	for _, p := range state.Players {
		if p != nil && p.IsActive {
			return p.ID
		}
	}
	return ""
}

func (r *Room) buildStateForPlayer(playerID string) GameStateUpdate {
	state := r.Engine.State
	currentPlayerID := currentTurnPlayerID(state)
	hasPerformedStartup := false
	if state != nil && currentPlayerID != "" {
		currentPlayer := state.Players[currentPlayerID]
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
			Indicators:         map[string]int{},
		}
		for k, v := range p.Tokens {
			view.Tokens[k] = v
		}
		setIndicator := func(key string, value int) {
			if value > 0 {
				view.Indicators[key] = value
			}
		}
		setIndicator("elf_blessing_count", countElfBlessings(p))
		setIndicator("mb_charge_count", countMagicBowCharges(p))
		setIndicator("sc_power_count", countSpiritCasterPowers(p))
		setIndicator("mg_dark_moon_count", countMoonDarkMoons(p))
		setIndicator("bt_cocoon_count", countButterflyCocoons(p))
		setIndicator("bp_shared_life_active", countBloodSharedLifeAsSource(state, p.ID))
		setIndicator("bp_shared_life_bound", countBloodSharedLifeAsHolder(p))
		setIndicator("se_sword_soul_count", countSwordEmperorSwordSouls(p))
		setIndicator("ml_dark_release_next_attack_bonus", combatPolicyAttackBonusByModifierID(p, "ml_dark_release_next_attack_bonus"))
		if hasRuleModifierWithModifierID(p, "ml_dark_release_lock_turn") {
			view.Indicators["ml_dark_release_lock_turn"] = 1
		}
		// 清理不应暴露给前端的 Tokens 镜像
		delete(view.Tokens, "elf_blessing_count")
		delete(view.Tokens, "mb_charge_count")
		delete(view.Tokens, "sc_power_count")
		delete(view.Tokens, "mg_dark_moon_count")
		delete(view.Tokens, "bt_cocoon_count")
		delete(view.Tokens, "bp_shared_life_active")
		delete(view.Tokens, "bp_shared_life_bound")
		delete(view.Tokens, "css_blood_cap")
		delete(view.Tokens, "ml_dark_release_next_attack_bonus")
		delete(view.Tokens, "ml_dark_release_lock_turn")
		delete(view.Tokens, "se_sword_soul_count")
		// 仅自己可见手牌具体内容，他人只能看到数量
		if pid == playerID {
			view.Hand = p.Hand
			view.ExclusiveCards = p.ExclusiveCards
			view.CurrentExtraAction = p.TurnState.CurrentExtraAction
			if len(p.TurnState.CurrentExtraElement) > 0 {
				view.CurrentExtraElement = make([]string, len(p.TurnState.CurrentExtraElement))
				for i, e := range p.TurnState.CurrentExtraElement {
					view.CurrentExtraElement[i] = string(e)
				}
			}
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
		CurrentPlayer:       currentPlayerID,
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

// buildBotStateSnapshot 构建 bot 决策所需的状态快照（精简版，不依赖 viewmodel）
func (r *Room) buildBotStateSnapshot(playerID string) bot.StateSnapshot {
	state := r.Engine.State
	currentPlayerID := currentTurnPlayerID(state)
	hasPerformedStartup := false
	if state != nil && currentPlayerID != "" {
		currentPlayer := state.Players[currentPlayerID]
		if currentPlayer != nil {
			hasPerformedStartup = currentPlayer.TurnState.HasStartupSkillOrSpecialActionsLocked()
		}
	}

	players := make(map[string]bot.PlayerSnapshot)
	for pid, p := range state.Players {
		snapshot := bot.PlayerSnapshot{
			ID:                 p.ID,
			Name:               p.Name,
			Camp:               string(p.Camp),
			Role:               p.Role,
			Form:               r.Engine.GetPlayerForm(pid),
			Orientation:        string(r.Engine.GetPlayerOrientation(pid)),
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
			snapshot.Tokens[k] = v
		}
		// 仅自己可见手牌具体内容
		if pid == playerID {
			snapshot.Hand = p.Hand
			snapshot.ExclusiveCards = p.ExclusiveCards
		}
		players[pid] = snapshot
	}

	return bot.StateSnapshot{
		TurnStage:           string(state.TurnStage),
		CombatStage:         string(state.CombatStage),
		Subflow:             string(state.Subflow),
		CurrentPlayer:       currentPlayerID,
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
	}
}
