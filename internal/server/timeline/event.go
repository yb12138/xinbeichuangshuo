package timeline

import (
	"starcup-engine/internal/model"
	"starcup-engine/internal/server/protocol"
)

type EventMeta struct {
	EventID     int64
	TurnID      int
	TurnStage   string
	CombatStage string
	Subflow     string
	ChainID     string
}

func BuildEvent(meta EventMeta, eventType string, data map[string]interface{}, message string) protocol.TimelineEvent {
	event := protocol.TimelineEvent{
		EventID:      meta.EventID,
		TurnID:       meta.TurnID,
		TurnStage:    meta.TurnStage,
		CombatStage:  meta.CombatStage,
		Subflow:      meta.Subflow,
		ChainID:      meta.ChainID,
		Type:         mapGameplayTimelineType(eventType, data),
		Outcome:      mapGameplayTimelineOutcome(eventType),
		Visibility:   "TimelineVisibilityPublic",
		Message:      message,
		GameplayType: eventType,
	}

	if actor := firstNonEmptyString(
		StringValue(data["player_id"]),
		StringValue(data["source_id"]),
		StringValue(data["attacker_id"]),
	); actor != "" {
		event.ActorUserID = actor
	}
	if actorName := firstNonEmptyString(
		StringValue(data["player_name"]),
		StringValue(data["source_name"]),
	); actorName != "" {
		event.ActorName = actorName
	}
	if target := firstNonEmptyString(
		StringValue(data["target_id"]),
		StringValue(data["player_id"]),
	); target != "" && target != event.ActorUserID {
		event.TargetUserIDs = []string{target}
	}
	if targetName := StringValue(data["target_name"]); targetName != "" {
		event.TargetName = targetName
	}
	if actionType := StringValue(data["action_type"]); actionType != "" {
		event.ActionType = actionType
	}
	if skillID := StringValue(data["skill_id"]); skillID != "" {
		event.SkillID = skillID
	}
	event.CardIDs = extractCardIDs(data["cards"])
	event.Cards = extractCards(data["cards"])
	event.Hidden = boolValue(data["hidden"])
	event.Damage = intValue(data["damage"])
	event.DamageType = StringValue(data["damage_type"])
	event.DetailKind = StringValue(data["kind"])
	event.CuePhase = StringValue(data["phase"])
	event.DrawCount = intValue(data["draw_count"])
	event.Reason = StringValue(data["reason"])
	event.Deltas = buildTimelineDeltas(eventType, data)

	return event
}

func StringValue(v interface{}) string {
	s, _ := v.(string)
	return s
}

func mapGameplayTimelineType(eventType string, data map[string]interface{}) string {
	switch eventType {
	case "prompt":
		return "TimelineInterruptRaised"
	case "card_revealed":
		actionType := StringValue(data["action_type"])
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

func buildTimelineDeltas(eventType string, data map[string]interface{}) []protocol.TimelineDelta {
	switch eventType {
	case "damage_dealt":
		targetID := StringValue(data["target_id"])
		damage := intValue(data["damage"])
		if targetID != "" && damage > 0 {
			return []protocol.TimelineDelta{{
				Type:         "TimelineDeltaDamage",
				TargetUserID: targetID,
				Value:        damage,
			}}
		}
	case "draw_cards":
		playerID := StringValue(data["player_id"])
		count := intValue(data["draw_count"])
		if playerID != "" && count > 0 {
			return []protocol.TimelineDelta{{
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
			if id := StringValue(m["id"]); id != "" {
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
					ID:          StringValue(card["id"]),
					Name:        StringValue(card["name"]),
					Type:        model.CardType(StringValue(card["type"])),
					Element:     model.Element(StringValue(card["element"])),
					Faction:     StringValue(card["faction"]),
					Damage:      intValue(card["damage"]),
					Description: StringValue(card["description"]),
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
