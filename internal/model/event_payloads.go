package model

// CardRevealedPayload is the Data payload for EventCardRevealed events.
type CardRevealedPayload struct {
	PlayerID   string `json:"player_id"`
	PlayerName string `json:"player_name"`
	Cards      []Card `json:"cards"`
	ActionType string `json:"action_type"`
	Hidden     bool   `json:"hidden"`
}

// DamageDealtPayload is the Data payload for EventDamageDealt events.
type DamageDealtPayload struct {
	SourceID   string `json:"source_id"`
	SourceName string `json:"source_name"`
	TargetID   string `json:"target_id"`
	TargetName string `json:"target_name"`
	Damage     int    `json:"damage"`
	DamageType string `json:"damage_type"`
}

// ActionStepPayload is the Data payload for EventActionStep events.
type ActionStepPayload struct {
	Line string `json:"line"`
	Kind string `json:"kind"`
}

// CombatCuePayload is the Data payload for EventCombatCue events.
type CombatCuePayload struct {
	AttackerID string `json:"attacker_id"`
	TargetID   string `json:"target_id"`
	Phase      string `json:"phase"`
}

// DrawCardsPayload is the Data payload for EventDrawCards events.
type DrawCardsPayload struct {
	PlayerID   string `json:"player_id"`
	PlayerName string `json:"player_name"`
	DrawCount  int    `json:"draw_count"`
	Reason     string `json:"reason"`
}
