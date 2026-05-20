package server

import (
	"fmt"
	"sort"

	"starcup-engine/internal/model"
)

func (r *Room) buildSyncStatePayload(playerID string) SyncStatePayload {
	stateView := r.buildStateForPlayer(playerID)
	payload := SyncStatePayload{
		RoomState:           map[bool]string{true: "Playing", false: "Lobby"}[r.Started],
		TurnStage:           stateView.TurnStage,
		CombatStage:         stateView.CombatStage,
		Subflow:             stateView.Subflow,
		TurnPlayerID:        stateView.CurrentPlayer,
		HasPerformedStartup: stateView.HasPerformedStartup,
		MoraleRed:           stateView.RedMorale,
		MoraleBlue:          stateView.BlueMorale,
		CupsRed:             stateView.RedCups,
		CupsBlue:            stateView.BlueCups,
		StonesRed:           []int{stateView.RedGems, stateView.RedCrystals},
		StonesBlue:          []int{stateView.BlueGems, stateView.BlueCrystals},
		DeckCount:           stateView.DeckCount,
		DiscardCount:        stateView.DiscardCount,
		AvailableSkills:     append([]AvailableSkill{}, stateView.AvailableSkills...),
		Characters:          append([]CharacterView{}, stateView.Characters...),
	}

	orderedIDs := make([]string, 0, len(stateView.Players))
	seen := make(map[string]struct{}, len(stateView.Players))
	for _, pid := range r.Engine.State.PlayerOrder {
		if _, ok := stateView.Players[pid]; !ok {
			continue
		}
		orderedIDs = append(orderedIDs, pid)
		seen[pid] = struct{}{}
	}
	for pid := range stateView.Players {
		if _, ok := seen[pid]; ok {
			continue
		}
		orderedIDs = append(orderedIDs, pid)
	}
	sort.Strings(orderedIDs[len(seen):])

	for _, pid := range orderedIDs {
		view := stateView.Players[pid]
		payload.Players = append(payload.Players, view)
	}

	return payload
}

func (r *Room) translateClientAction(playerID string, req ClientActionRequest) (model.PlayerAction, error) {
	if !model.IsKnownPlayerActionType(req.ActionType) {
		return model.PlayerAction{}, newProtocolInputError(
			protocolErrorCodeUnknownActionType,
			fmt.Sprintf("未知 action_type: %s", req.ActionType),
			map[string]interface{}{"action_type": req.ActionType},
		)
	}

	action := model.PlayerAction{
		PlayerID: playerID,
		Type:     req.ActionType,
		SkillID:  req.SkillID,
	}

	if len(req.Targets) > 0 {
		action.TargetIDs = make([]string, 0, len(req.Targets))
		for _, target := range req.Targets {
			if target.TargetUserID == "" {
				continue
			}
			action.TargetIDs = append(action.TargetIDs, target.TargetUserID)
		}
	}
	if len(action.TargetIDs) == 1 {
		action.TargetID = action.TargetIDs[0]
	}

	if len(req.OptionIndexes) > 0 {
		action.Selections = append([]int{}, req.OptionIndexes...)
	}

	player := r.Engine.State.Players[playerID]
	if player == nil {
		return action, fmt.Errorf("玩家不存在")
	}
	cardIDs := append([]string{}, req.CardIDs...)
	if req.CardID != "" {
		cardIDs = append([]string{req.CardID}, cardIDs...)
	}
	if len(cardIDs) > 0 {
		action.CardIDs = append([]string{}, cardIDs...)
		if req.CardID != "" {
			action.CardID = req.CardID
		} else {
			action.CardID = cardIDs[0]
		}
		switch action.Type {
		case model.CmdAttack, model.CmdMagic, model.CmdRespond:
			action.CardID = cardIDs[0]
		}
	}

	if len(req.ExtraArgs) > 0 {
		action.ExtraArgs = append(action.ExtraArgs, req.ExtraArgs...)
	}

	return action, nil
}
