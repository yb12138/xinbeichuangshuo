package model

import "fmt"

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

// Validate checks that the typed payload attached to the event matches its type.
func (e GameEvent) Validate() error {
	switch e.Type {
	case EventAskInput:
		if e.Prompt == nil {
			return fmt.Errorf("game event %s requires prompt payload", e.Type)
		}
		if e.CardRevealed != nil || e.DamageDealt != nil || e.ActionStep != nil || e.CombatCue != nil || e.DrawCards != nil {
			return fmt.Errorf("game event %s cannot carry other typed payloads", e.Type)
		}
	case EventCardRevealed:
		if e.CardRevealed == nil {
			return fmt.Errorf("game event %s requires card_revealed payload", e.Type)
		}
		if e.Prompt != nil || e.DamageDealt != nil || e.ActionStep != nil || e.CombatCue != nil || e.DrawCards != nil {
			return fmt.Errorf("game event %s cannot carry other typed payloads", e.Type)
		}
	case EventDamageDealt:
		if e.DamageDealt == nil {
			return fmt.Errorf("game event %s requires damage_dealt payload", e.Type)
		}
		if e.Prompt != nil || e.CardRevealed != nil || e.ActionStep != nil || e.CombatCue != nil || e.DrawCards != nil {
			return fmt.Errorf("game event %s cannot carry other typed payloads", e.Type)
		}
	case EventActionStep:
		if e.ActionStep == nil {
			return fmt.Errorf("game event %s requires action_step payload", e.Type)
		}
		if e.Prompt != nil || e.CardRevealed != nil || e.DamageDealt != nil || e.CombatCue != nil || e.DrawCards != nil {
			return fmt.Errorf("game event %s cannot carry other typed payloads", e.Type)
		}
	case EventCombatCue:
		if e.CombatCue == nil {
			return fmt.Errorf("game event %s requires combat_cue payload", e.Type)
		}
		if e.Prompt != nil || e.CardRevealed != nil || e.DamageDealt != nil || e.ActionStep != nil || e.DrawCards != nil {
			return fmt.Errorf("game event %s cannot carry other typed payloads", e.Type)
		}
	case EventDrawCards:
		if e.DrawCards == nil {
			return fmt.Errorf("game event %s requires draw_cards payload", e.Type)
		}
		if e.Prompt != nil || e.CardRevealed != nil || e.DamageDealt != nil || e.ActionStep != nil || e.CombatCue != nil {
			return fmt.Errorf("game event %s cannot carry other typed payloads", e.Type)
		}
	case EventLog, EventStateUpdate, EventError, EventGameEnd:
		if e.Prompt != nil || e.CardRevealed != nil || e.DamageDealt != nil || e.ActionStep != nil || e.CombatCue != nil || e.DrawCards != nil {
			return fmt.Errorf("game event %s cannot carry typed payloads", e.Type)
		}
	default:
		return fmt.Errorf("unknown game event type %s", e.Type)
	}
	return nil
}
