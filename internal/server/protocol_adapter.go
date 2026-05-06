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
	if !model.IsKnownPlayerActionType(model.PlayerActionType(req.ActionType)) {
		return model.PlayerAction{}, newProtocolInputError(
			protocolErrorCodeUnknownActionType,
			fmt.Sprintf("未知 action_type: %s", req.ActionType),
			map[string]interface{}{"action_type": req.ActionType},
		)
	}

	action := model.PlayerAction{
		PlayerID: playerID,
		Type:     model.PlayerActionType(req.ActionType),
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
	if action.TargetID == "" && req.TargetRef != "" {
		action.TargetID = req.TargetRef
	}

	if len(req.OptionIndexes) > 0 {
		action.Selections = append([]int{}, req.OptionIndexes...)
	}

	player := r.Engine.State.Players[playerID]
	if player == nil {
		return action, fmt.Errorf("玩家不存在")
	}
	if len(req.UsedCardUUIDs) > 0 {
		indexes, err := findPlayableCardIndexesByUUID(player, req.UsedCardUUIDs)
		if err != nil {
			return action, err
		}
		switch action.Type {
		case model.CmdAttack, model.CmdMagic, model.CmdRespond:
			action.CardIndex = indexes[0]
		case model.CmdSkill, model.CmdSelect:
			if len(action.Selections) == 0 {
				action.Selections = indexes
			}
		}
	}

	if req.ResponseMode != "" {
		action.ExtraArgs = append(action.ExtraArgs, req.ResponseMode)
	}
	if len(req.ExtraArgs) > 0 {
		action.ExtraArgs = append(action.ExtraArgs, req.ExtraArgs...)
	}

	return action, nil
}

func findPlayableCardIndexesByUUID(player *model.Player, ids []string) ([]int, error) {
	indexes := make([]int, 0, len(ids))
	for _, id := range ids {
		idx := findPlayableCardIndexByUUID(player, id)
		if idx < 0 {
			return nil, fmt.Errorf("未找到卡牌: %s", id)
		}
		indexes = append(indexes, idx)
	}
	return indexes, nil
}

func findPlayableCardIndexByUUID(player *model.Player, id string) int {
	if player == nil || id == "" {
		return -1
	}
	for i, card := range player.Hand {
		if card.ID == id {
			return i
		}
	}
	base := len(player.Hand)
	blessings := listElfBlessingsForPlayableIndex(player)
	for i, card := range blessings {
		if card.ID == id {
			return base + i
		}
	}
	base += len(blessings)
	for i, card := range player.ExclusiveCards {
		if card.ID == id {
			return base + i
		}
	}
	return -1
}

func listElfBlessingsForPlayableIndex(player *model.Player) []model.Card {
	if player == nil {
		return nil
	}
	out := make([]model.Card, 0)
	for _, fc := range player.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectElfBlessing {
			continue
		}
		out = append(out, fc.Card)
	}
	return out
}
