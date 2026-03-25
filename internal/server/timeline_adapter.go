package server

import (
	"fmt"
	"sync/atomic"

	"starcup-engine/internal/model"
)

func (r *Room) nextTimelineEventMeta() (int64, string, int, string, string, string) {
	eventID := atomic.AddInt64(&r.timelineSeq, 1)

	turnID := 0
	turnStage := ""
	combatStage := ""
	subflow := ""
	if r.Engine != nil && r.Engine.State != nil {
		turnID = r.Engine.State.CurrentTurn + 1
		turnStage = string(r.Engine.State.TurnStage)
		combatStage = string(r.Engine.State.CombatStage)
		subflow = string(r.Engine.State.Subflow)
	}

	return eventID, fmt.Sprintf("chain_%d", eventID), turnID, turnStage, combatStage, subflow
}

func (r *Room) buildTimelineNotify(eventType string, data map[string]interface{}, message string) TimelineNotifyPayload {
	eventID, chainID, turnID, turnStage, combatStage, subflow := r.nextTimelineEventMeta()
	event := TimelineEvent{
		EventID:      eventID,
		TurnID:       turnID,
		TurnStage:    turnStage,
		CombatStage:  combatStage,
		Subflow:      subflow,
		ChainID:      chainID,
		Type:         mapGameplayTimelineType(eventType, data),
		Outcome:      mapGameplayTimelineOutcome(eventType),
		Visibility:   "TimelineVisibilityPublic",
		Message:      message,
		GameplayType: eventType,
	}

	if actor := firstNonEmptyString(
		stringValue(data["player_id"]),
		stringValue(data["source_id"]),
		stringValue(data["attacker_id"]),
	); actor != "" {
		event.ActorUserID = actor
	}
	if actorName := firstNonEmptyString(
		stringValue(data["player_name"]),
		stringValue(data["source_name"]),
	); actorName != "" {
		event.ActorName = actorName
	}
	if target := firstNonEmptyString(
		stringValue(data["target_id"]),
		stringValue(data["player_id"]),
	); target != "" && target != event.ActorUserID {
		event.TargetUserIDs = []string{target}
	}
	if targetName := stringValue(data["target_name"]); targetName != "" {
		event.TargetName = targetName
	}
	if actionType := stringValue(data["action_type"]); actionType != "" {
		event.ActionType = actionType
	}
	if skillID := stringValue(data["skill_id"]); skillID != "" {
		event.SkillID = skillID
	}
	event.CardIDs = extractCardIDs(data["cards"])
	event.Cards = extractCards(data["cards"])
	event.Hidden = boolValue(data["hidden"])
	event.Damage = intValue(data["damage"])
	event.DamageType = stringValue(data["damage_type"])
	event.DetailKind = stringValue(data["kind"])
	event.CuePhase = stringValue(data["phase"])
	event.DrawCount = intValue(data["draw_count"])
	event.Reason = stringValue(data["reason"])
	event.Deltas = buildTimelineDeltas(eventType, data)

	return TimelineNotifyPayload{
		RoomID:   r.Code,
		SeqStart: eventID,
		SeqEnd:   eventID,
		IsReplay: false,
		Events:   []TimelineEvent{event},
	}
}

func mapGameplayTimelineType(eventType string, data map[string]interface{}) string {
	switch eventType {
	case "prompt":
		return "TimelineInterruptRaised"
	case "card_revealed":
		actionType := stringValue(data["action_type"])
		if actionType == "defend" || actionType == "counter" {
			return "TimelineResponseSelected"
		}
		return "TimelineActionDeclared"
	case "damage_dealt":
		return "TimelineCombatResolved"
	case "combat_cue":
		return "TimelineActionDeclared"
	case "game_end":
		return "TimelineChainClosed"
	default:
		return "TimelineEffectResolved"
	}
}

func mapGameplayTimelineOutcome(eventType string) string {
	switch eventType {
	case "game_end":
		return "TimelineOutcomeSuccess"
	default:
		return "TimelineOutcomeSuccess"
	}
}

func buildTimelineDeltas(eventType string, data map[string]interface{}) []TimelineDelta {
	switch eventType {
	case "damage_dealt":
		targetID := stringValue(data["target_id"])
		damage := intValue(data["damage"])
		if targetID != "" && damage > 0 {
			return []TimelineDelta{{
				Type:         "TimelineDeltaDamage",
				TargetUserID: targetID,
				Value:        damage,
			}}
		}
	case "draw_cards":
		playerID := stringValue(data["player_id"])
		count := intValue(data["draw_count"])
		if playerID != "" && count > 0 {
			return []TimelineDelta{{
				Type:         "TimelineDeltaHandCount",
				TargetUserID: playerID,
				Value:        count,
			}}
		}
	}
	return nil
}

func extractCardIDs(raw interface{}) []string {
	switch cards := raw.(type) {
	case []model.Card:
		out := make([]string, 0, len(cards))
		for _, card := range cards {
			if card.ID != "" {
				out = append(out, card.ID)
			}
		}
		return out
	case []interface{}:
		out := make([]string, 0, len(cards))
		for _, item := range cards {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			if id := stringValue(m["id"]); id != "" {
				out = append(out, id)
			}
		}
		return out
	default:
		return nil
	}
}

func extractCards(raw interface{}) []model.Card {
	switch cards := raw.(type) {
	case []model.Card:
		if len(cards) == 0 {
			return nil
		}
		out := make([]model.Card, len(cards))
		copy(out, cards)
		return out
	case []interface{}:
		out := make([]model.Card, 0, len(cards))
		for _, item := range cards {
			switch card := item.(type) {
			case model.Card:
				out = append(out, card)
			case map[string]interface{}:
				out = append(out, model.Card{
					ID:          stringValue(card["id"]),
					Name:        stringValue(card["name"]),
					Type:        model.CardType(stringValue(card["type"])),
					Element:     model.Element(stringValue(card["element"])),
					Faction:     stringValue(card["faction"]),
					Damage:      intValue(card["damage"]),
					Description: stringValue(card["description"]),
				})
			}
		}
		if len(out) == 0 {
			return nil
		}
		return out
	default:
		return nil
	}
}

func stringValue(v interface{}) string {
	s, _ := v.(string)
	return s
}

func intValue(v interface{}) int {
	switch t := v.(type) {
	case int:
		return t
	case int64:
		return int(t)
	case float64:
		return int(t)
	case float32:
		return int(t)
	default:
		return 0
	}
}

func boolValue(v interface{}) bool {
	b, _ := v.(bool)
	return b
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
