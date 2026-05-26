package model

import "fmt"

// NarrativeTracePayload carries stable event-chain metadata for timeline
// projection. It is runtime-only metadata and is not part of saved game state.
type NarrativeTracePayload struct {
	NarrativeWindowID  string     `json:"narrative_window_id,omitempty"`
	ActionID           string     `json:"action_id,omitempty"`
	CombatID           string     `json:"combat_id,omitempty"`
	SourceEventID      string     `json:"source_event_id,omitempty"`
	ParentEventID      *int64     `json:"parent_event_id,omitempty"`
	NarrativeKind      string     `json:"narrative_kind,omitempty"`
	VisualKind         string     `json:"visual_kind,omitempty"`
	CardRole           string     `json:"card_role,omitempty"`
	SkillPhase         string     `json:"skill_phase,omitempty"`
	Timing             string     `json:"timing,omitempty"`
	EffectType         string     `json:"effect_type,omitempty"`
	ExtraActionType    string     `json:"extra_action_type,omitempty"`
	ExtraActionElement string     `json:"extra_action_element,omitempty"`
	FieldCard          *FieldCard `json:"field_card,omitempty"`
}

// TimelineMarkerPayload is used for structured narrative events that do not
// naturally map to a legacy gameplay payload, such as action_started or
// extra_action_granted.
type TimelineMarkerPayload struct {
	PlayerID           string     `json:"player_id,omitempty"`
	PlayerName         string     `json:"player_name,omitempty"`
	ActionType         string     `json:"action_type,omitempty"`
	SkillID            string     `json:"skill_id,omitempty"`
	SkillName          string     `json:"skill_name,omitempty"`
	EffectText         string     `json:"effect_text,omitempty"`
	Summary            string     `json:"summary,omitempty"`
	TargetIDs          []string   `json:"target_ids,omitempty"`
	NarrativeKind      string     `json:"narrative_kind,omitempty"`
	VisualKind         string     `json:"visual_kind,omitempty"`
	CardRole           string     `json:"card_role,omitempty"`
	SkillPhase         string     `json:"skill_phase,omitempty"`
	Timing             string     `json:"timing,omitempty"`
	EffectType         string     `json:"effect_type,omitempty"`
	ExtraActionType    string     `json:"extra_action_type,omitempty"`
	ExtraActionElement string     `json:"extra_action_element,omitempty"`
	FieldCard          *FieldCard `json:"field_card,omitempty"`
}

// CardRevealedPayload is the typed payload for EventCardRevealed events.
type CardRevealedPayload struct {
	PlayerID   string `json:"player_id"`
	PlayerName string `json:"player_name"`
	Cards      []Card `json:"cards"`
	ActionType string `json:"action_type"`
	Hidden     bool   `json:"hidden"`
}

// DamageDealtPayload is the typed payload for EventDamageDealt events.
type DamageDealtPayload struct {
	SourceID   string `json:"source_id"`
	SourceName string `json:"source_name"`
	TargetID   string `json:"target_id"`
	TargetName string `json:"target_name"`
	Damage     int    `json:"damage"`
	DamageType string `json:"damage_type"`
}

// ActionStepPayload is the typed payload for EventActionStep events.
type ActionStepPayload struct {
	Line string `json:"line"`
	Kind string `json:"kind"`
}

// CombatCuePayload is the typed payload for EventCombatCue events.
type CombatCuePayload struct {
	AttackerID string `json:"attacker_id"`
	TargetID   string `json:"target_id"`
	Phase      string `json:"phase"`
}

// DrawCardsPayload is the typed payload for EventDrawCards events.
type DrawCardsPayload struct {
	PlayerID   string `json:"player_id"`
	PlayerName string `json:"player_name"`
	DrawCount  int    `json:"draw_count"`
	Reason     string `json:"reason"`
}

// SkillActivatedPayload is the typed payload for EventSkillActivated events.
type SkillActivatedPayload struct {
	PlayerID   string   `json:"player_id"`
	PlayerName string   `json:"player_name"`
	SkillID    string   `json:"skill_id"`
	SkillName  string   `json:"skill_name"`
	EffectText string   `json:"effect_text"`
	TargetIDs  []string `json:"target_ids,omitempty"`
}

// SpecialActionPayload is the typed payload for EventSpecialAction events.
type SpecialActionPayload struct {
	PlayerID   string   `json:"player_id"`
	PlayerName string   `json:"player_name"`
	ActionType string   `json:"action_type"`
	TargetIDs  []string `json:"target_ids,omitempty"`
	Summary    string   `json:"summary"`
}

// StateDeltaItem describes a single public-visible state change.
type StateDeltaItem struct {
	Type          string     `json:"type"`
	Scope         string     `json:"scope,omitempty"`
	TargetUserID  string     `json:"target_user_id,omitempty"`
	Camp          string     `json:"camp,omitempty"`
	Field         string     `json:"field,omitempty"`
	Before        int        `json:"before,omitempty"`
	After         int        `json:"after,omitempty"`
	Value         int        `json:"value,omitempty"`
	Reason        string     `json:"reason,omitempty"`
	SourceEventID string     `json:"source_event_id,omitempty"`
	BeforeText    string     `json:"before_text,omitempty"`
	AfterText     string     `json:"after_text,omitempty"`
	FieldCard     *FieldCard `json:"field_card,omitempty"`
}

// StateDeltaPayload is the typed payload for EventStateDelta events.
type StateDeltaPayload struct {
	Deltas []StateDeltaItem `json:"deltas"`
	Reason string           `json:"reason,omitempty"`
}

// Validate checks that the typed payload attached to the event matches its type.
func (e GameEvent) Validate() error {
	hasAnyTypedPayload := func() bool {
		return e.Prompt != nil || e.CardRevealed != nil || e.DamageDealt != nil || e.ActionStep != nil ||
			e.CombatCue != nil || e.DrawCards != nil || e.SkillActivated != nil || e.SpecialAction != nil ||
			e.StateDelta != nil || e.TimelineMarker != nil
	}
	hasOtherTypedPayload := func(expected string) bool {
		if expected != "prompt" && e.Prompt != nil {
			return true
		}
		if expected != "card_revealed" && e.CardRevealed != nil {
			return true
		}
		if expected != "damage_dealt" && e.DamageDealt != nil {
			return true
		}
		if expected != "action_step" && e.ActionStep != nil {
			return true
		}
		if expected != "combat_cue" && e.CombatCue != nil {
			return true
		}
		if expected != "draw_cards" && e.DrawCards != nil {
			return true
		}
		if expected != "skill_activated" && e.SkillActivated != nil {
			return true
		}
		if expected != "special_action" && e.SpecialAction != nil {
			return true
		}
		if expected != "state_delta" && e.StateDelta != nil {
			return true
		}
		if expected != "timeline_marker" && e.TimelineMarker != nil {
			return true
		}
		return false
	}

	switch e.Type {
	case EventAskInput:
		if e.Prompt == nil {
			return fmt.Errorf("game event %s requires prompt payload", e.Type)
		}
		if hasOtherTypedPayload("prompt") {
			return fmt.Errorf("game event %s cannot carry other typed payloads", e.Type)
		}
	case EventCardRevealed:
		if e.CardRevealed == nil {
			return fmt.Errorf("game event %s requires card_revealed payload", e.Type)
		}
		if hasOtherTypedPayload("card_revealed") {
			return fmt.Errorf("game event %s cannot carry other typed payloads", e.Type)
		}
	case EventDamageDealt:
		if e.DamageDealt == nil {
			return fmt.Errorf("game event %s requires damage_dealt payload", e.Type)
		}
		if hasOtherTypedPayload("damage_dealt") {
			return fmt.Errorf("game event %s cannot carry other typed payloads", e.Type)
		}
	case EventActionStep:
		if e.ActionStep == nil {
			return fmt.Errorf("game event %s requires action_step payload", e.Type)
		}
		if hasOtherTypedPayload("action_step") {
			return fmt.Errorf("game event %s cannot carry other typed payloads", e.Type)
		}
	case EventCombatCue:
		if e.CombatCue == nil {
			return fmt.Errorf("game event %s requires combat_cue payload", e.Type)
		}
		if hasOtherTypedPayload("combat_cue") {
			return fmt.Errorf("game event %s cannot carry other typed payloads", e.Type)
		}
	case EventDrawCards:
		if e.DrawCards == nil {
			return fmt.Errorf("game event %s requires draw_cards payload", e.Type)
		}
		if hasOtherTypedPayload("draw_cards") {
			return fmt.Errorf("game event %s cannot carry other typed payloads", e.Type)
		}
	case EventSkillActivated:
		if e.SkillActivated == nil {
			return fmt.Errorf("game event %s requires skill_activated payload", e.Type)
		}
		if hasOtherTypedPayload("skill_activated") {
			return fmt.Errorf("game event %s cannot carry other typed payloads", e.Type)
		}
	case EventSpecialAction:
		if e.SpecialAction == nil {
			return fmt.Errorf("game event %s requires special_action payload", e.Type)
		}
		if hasOtherTypedPayload("special_action") {
			return fmt.Errorf("game event %s cannot carry other typed payloads", e.Type)
		}
	case EventStateDelta:
		if e.StateDelta == nil {
			return fmt.Errorf("game event %s requires state_delta payload", e.Type)
		}
		if hasOtherTypedPayload("state_delta") {
			return fmt.Errorf("game event %s cannot carry other typed payloads", e.Type)
		}
	case EventTimelineMarker:
		if e.TimelineMarker == nil {
			return fmt.Errorf("game event %s requires timeline_marker payload", e.Type)
		}
		if hasOtherTypedPayload("timeline_marker") {
			return fmt.Errorf("game event %s cannot carry other typed payloads", e.Type)
		}
	case EventLog, EventStateUpdate, EventError, EventGameEnd:
		if hasAnyTypedPayload() {
			return fmt.Errorf("game event %s cannot carry typed payloads", e.Type)
		}
	default:
		return fmt.Errorf("unknown game event type %s", e.Type)
	}
	return nil
}
