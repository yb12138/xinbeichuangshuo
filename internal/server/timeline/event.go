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

type Payload struct {
	Type       string
	Message    string
	PlayerID   string
	PlayerName string
	SourceID   string
	SourceName string
	TargetID   string
	TargetName string
	AttackerID string
	ActionType string
	SkillID    string
	SkillName  string
	EffectText string
	Summary    string
	TargetIDs  []string
	Cards      []model.Card
	Hidden     bool
	Damage     int
	DamageType string
	Kind       string
	Phase      string
	DrawCount  int
	Reason     string
	Deltas     []protocol.TimelineDelta
}

func BuildEvent(meta EventMeta, payload Payload) protocol.TimelineEvent {
	event := protocol.TimelineEvent{
		EventID:      meta.EventID,
		TurnID:       meta.TurnID,
		TurnStage:    meta.TurnStage,
		CombatStage:  meta.CombatStage,
		Subflow:      meta.Subflow,
		ChainID:      meta.ChainID,
		Type:         mapGameplayTimelineType(payload),
		Outcome:      mapGameplayTimelineOutcome(payload.Type),
		Visibility:   "TimelineVisibilityPublic",
		Message:      payload.Message,
		GameplayType: payload.Type,
		ActionType:   payload.ActionType,
		SkillID:      payload.SkillID,
		SkillName:    payload.SkillName,
		EffectText:   payload.EffectText,
		Summary:      payload.Summary,
		Cards:        cloneCards(payload.Cards),
		CardIDs:      cardIDs(payload.Cards),
		Hidden:       payload.Hidden,
		Damage:       payload.Damage,
		DamageType:   payload.DamageType,
		DetailKind:   payload.Kind,
		CuePhase:     payload.Phase,
		DrawCount:    payload.DrawCount,
		Reason:       payload.Reason,
	}

	if actor := firstNonEmptyString(payload.PlayerID, payload.SourceID, payload.AttackerID); actor != "" {
		event.ActorUserID = actor
	}
	if actorName := firstNonEmptyString(payload.PlayerName, payload.SourceName); actorName != "" {
		event.ActorName = actorName
	}
	if len(payload.TargetIDs) > 0 {
		event.TargetUserIDs = append([]string{}, payload.TargetIDs...)
	} else if target := firstNonEmptyString(payload.TargetID, payload.PlayerID); target != "" && (payload.Type == "damage_dealt" || target != event.ActorUserID) {
		event.TargetUserIDs = []string{target}
	}
	if payload.TargetName != "" {
		event.TargetName = payload.TargetName
	}
	event.Deltas = buildTimelineDeltas(payload)

	return event
}

func mapGameplayTimelineType(payload Payload) string {
	switch payload.Type {
	case "prompt":
		return "TimelineInterruptRaised"
	case "card_revealed":
		if payload.ActionType == "defend" || payload.ActionType == "counter" {
			return "TimelineResponseSelected"
		}
		return "TimelineActionDeclared"
	case "damage_dealt":
		return "TimelineCombatResolved"
	case "combat_cue":
		return "TimelineActionDeclared"
	case "skill_activated", "special_action":
		return "TimelineActionDeclared"
	case "state_delta":
		return "TimelineEffectResolved"
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

func buildTimelineDeltas(payload Payload) []protocol.TimelineDelta {
	if len(payload.Deltas) > 0 {
		out := make([]protocol.TimelineDelta, len(payload.Deltas))
		copy(out, payload.Deltas)
		return out
	}
	switch payload.Type {
	case "damage_dealt":
		if payload.TargetID != "" && payload.Damage > 0 {
			return []protocol.TimelineDelta{{
				Type:         "TimelineDeltaDamage",
				TargetUserID: payload.TargetID,
				Value:        payload.Damage,
			}}
		}
	case "draw_cards":
		if payload.PlayerID != "" && payload.DrawCount > 0 {
			return []protocol.TimelineDelta{{
				Type:         "TimelineDeltaHandCount",
				TargetUserID: payload.PlayerID,
				Value:        payload.DrawCount,
			}}
		}
	}
	return nil
}

func cardIDs(cards []model.Card) []string {
	out := make([]string, 0, len(cards))
	for _, card := range cards {
		if card.ID != "" {
			out = append(out, card.ID)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func cloneCards(cards []model.Card) []model.Card {
	if len(cards) == 0 {
		return nil
	}
	out := make([]model.Card, len(cards))
	copy(out, cards)
	return out
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
