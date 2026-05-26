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
	Trace      *model.NarrativeTracePayload
}

func BuildEvent(meta EventMeta, payload Payload) protocol.TimelineEvent {
	event := protocol.TimelineEvent{
		EventID:            meta.EventID,
		TurnID:             meta.TurnID,
		TurnStage:          meta.TurnStage,
		CombatStage:        meta.CombatStage,
		Subflow:            meta.Subflow,
		ChainID:            meta.ChainID,
		SourceEventID:      traceString(payload.Trace, func(t *model.NarrativeTracePayload) string { return t.SourceEventID }),
		ParentEventID:      traceParentEventID(payload.Trace),
		NarrativeWindowID:  traceString(payload.Trace, func(t *model.NarrativeTracePayload) string { return t.NarrativeWindowID }),
		ActionID:           traceString(payload.Trace, func(t *model.NarrativeTracePayload) string { return t.ActionID }),
		CombatID:           traceString(payload.Trace, func(t *model.NarrativeTracePayload) string { return t.CombatID }),
		NarrativeKind:      traceString(payload.Trace, func(t *model.NarrativeTracePayload) string { return t.NarrativeKind }),
		VisualKind:         traceString(payload.Trace, func(t *model.NarrativeTracePayload) string { return t.VisualKind }),
		CardRole:           traceString(payload.Trace, func(t *model.NarrativeTracePayload) string { return t.CardRole }),
		SkillPhase:         traceString(payload.Trace, func(t *model.NarrativeTracePayload) string { return t.SkillPhase }),
		Timing:             traceString(payload.Trace, func(t *model.NarrativeTracePayload) string { return t.Timing }),
		EffectType:         traceString(payload.Trace, func(t *model.NarrativeTracePayload) string { return t.EffectType }),
		ExtraActionType:    traceString(payload.Trace, func(t *model.NarrativeTracePayload) string { return t.ExtraActionType }),
		ExtraActionElement: traceString(payload.Trace, func(t *model.NarrativeTracePayload) string { return t.ExtraActionElement }),
		FieldCard:          traceFieldCard(payload.Trace),
		Type:               mapGameplayTimelineType(payload),
		Outcome:            mapGameplayTimelineOutcome(payload.Type),
		Visibility:         "TimelineVisibilityPublic",
		Message:            payload.Message,
		GameplayType:       payload.Type,
		ActionType:         payload.ActionType,
		SkillID:            payload.SkillID,
		SkillName:          payload.SkillName,
		EffectText:         payload.EffectText,
		Summary:            payload.Summary,
		Cards:              cloneCards(payload.Cards),
		CardIDs:            cardIDs(payload.Cards),
		Hidden:             payload.Hidden,
		Damage:             payload.Damage,
		DamageType:         payload.DamageType,
		DetailKind:         payload.Kind,
		CuePhase:           payload.Phase,
		DrawCount:          payload.DrawCount,
		Reason:             payload.Reason,
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
	applyDefaultNarrativeFields(&event, payload)
	applyNarrativeDeltaFields(&event)

	return event
}

func traceString(trace *model.NarrativeTracePayload, pick func(*model.NarrativeTracePayload) string) string {
	if trace == nil {
		return ""
	}
	return pick(trace)
}

func traceParentEventID(trace *model.NarrativeTracePayload) *int64 {
	if trace == nil || trace.ParentEventID == nil {
		return nil
	}
	value := *trace.ParentEventID
	return &value
}

func traceFieldCard(trace *model.NarrativeTracePayload) *model.FieldCard {
	if trace == nil || trace.FieldCard == nil {
		return nil
	}
	cp := *trace.FieldCard
	return &cp
}

func applyDefaultNarrativeFields(event *protocol.TimelineEvent, payload Payload) {
	if event == nil {
		return
	}
	if event.NarrativeKind == "" {
		event.NarrativeKind = defaultNarrativeKind(payload)
	}
	if event.VisualKind == "" {
		event.VisualKind = defaultVisualKind(payload)
	}
	if event.CardRole == "" && payload.Type == "card_revealed" {
		event.CardRole = payload.ActionType
	}
	if event.SkillPhase == "" && payload.Type == "skill_activated" {
		event.SkillPhase = "triggered"
	}
	if event.NarrativeKind == "field_effect_applied" || event.NarrativeKind == "field_effect_removed" {
		event.VisualKind = firstNonEmptyString(event.VisualKind, "effect_token")
	}
}

func applyNarrativeDeltaFields(event *protocol.TimelineEvent) {
	if event == nil || len(event.Deltas) == 0 {
		return
	}
	for _, delta := range event.Deltas {
		if delta.FieldCard == nil {
			continue
		}
		if event.FieldCard == nil {
			fieldCardCopy := *delta.FieldCard
			event.FieldCard = &fieldCardCopy
		}
		if event.EffectType == "" {
			event.EffectType = string(delta.FieldCard.Effect)
		}
		if event.ActorUserID == "" && delta.FieldCard.SourceID != "" {
			event.ActorUserID = delta.FieldCard.SourceID
		}
		if len(event.TargetUserIDs) == 0 {
			if delta.TargetUserID != "" {
				event.TargetUserIDs = []string{delta.TargetUserID}
			} else if delta.FieldCard.OwnerID != "" {
				event.TargetUserIDs = []string{delta.FieldCard.OwnerID}
			}
		}
		return
	}
}

func defaultNarrativeKind(payload Payload) string {
	switch payload.Type {
	case "timeline_marker":
		if payload.Trace != nil && payload.Trace.NarrativeKind != "" {
			return payload.Trace.NarrativeKind
		}
	case "card_revealed":
		return "card_played"
	case "combat_cue":
		if payload.Phase == "attack" {
			return "combat_declared"
		}
		return "combat_response"
	case "damage_dealt":
		return "damage_dealt"
	case "skill_activated":
		return "skill_triggered"
	case "draw_cards":
		return "draw_cards"
	case "state_delta":
		return narrativeKindForStateDeltas(payload.Deltas)
	}
	return ""
}

func defaultVisualKind(payload Payload) string {
	switch payload.Type {
	case "card_revealed":
		if payload.Hidden || payload.ActionType == "discard" {
			return "none"
		}
		return "card"
	case "skill_activated":
		return "skill_token"
	case "damage_dealt":
		return "damage"
	case "state_delta":
		if narrativeKindForStateDeltas(payload.Deltas) != "" {
			return "effect_token"
		}
	}
	if payload.Trace != nil {
		return payload.Trace.VisualKind
	}
	return ""
}

func narrativeKindForStateDeltas(deltas []protocol.TimelineDelta) string {
	for _, delta := range deltas {
		switch delta.Type {
		case "field_card_added":
			return "field_effect_applied"
		case "field_card_removed":
			return "field_effect_removed"
		}
	}
	return "state_delta"
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
	case "timeline_marker":
		if payload.Trace != nil && payload.Trace.NarrativeKind == "action_closed" {
			return "TimelineChainClosed"
		}
		return "TimelineEffectResolved"
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
