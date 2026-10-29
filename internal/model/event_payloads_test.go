package model

import "testing"

func TestGameEventValidateRequiresMatchingTypedPayload(t *testing.T) {
	tests := []struct {
		name    string
		event   GameEvent
		wantErr bool
	}{
		{
			name:  "prompt needs prompt payload",
			event: GameEvent{Type: EventAskInput, Prompt: &Prompt{}},
		},
		{
			name: "prompt rejects other payloads",
			event: GameEvent{
				Type:        EventAskInput,
				Prompt:      &Prompt{},
				DamageDealt: &DamageDealtPayload{},
			},
			wantErr: true,
		},
		{
			name:  "card revealed needs payload",
			event: GameEvent{Type: EventCardRevealed, CardRevealed: &CardRevealedPayload{}},
		},
		{
			name:  "damage dealt needs payload",
			event: GameEvent{Type: EventDamageDealt, DamageDealt: &DamageDealtPayload{}},
		},
		{
			name:  "action step needs payload",
			event: GameEvent{Type: EventActionStep, ActionStep: &ActionStepPayload{}},
		},
		{
			name:  "combat cue needs payload",
			event: GameEvent{Type: EventCombatCue, CombatCue: &CombatCuePayload{}},
		},
		{
			name:  "draw cards needs payload",
			event: GameEvent{Type: EventDrawCards, DrawCards: &DrawCardsPayload{}},
		},
		{
			name:  "skill activated needs payload",
			event: GameEvent{Type: EventSkillActivated, SkillActivated: &SkillActivatedPayload{PlayerID: "p1", SkillID: "s1", SkillName: "技能"}},
		},
		{
			name:  "special action needs payload",
			event: GameEvent{Type: EventSpecialAction, SpecialAction: &SpecialActionPayload{PlayerID: "p1", ActionType: "Buy"}},
		},
		{
			name:  "state delta needs payload",
			event: GameEvent{Type: EventStateDelta, StateDelta: &StateDeltaPayload{Deltas: []StateDeltaItem{{Type: "morale", Camp: "Red", Before: 15, After: 14, Value: -1}}}},
		},
		{
			name:    "log rejects typed payloads",
			event:   GameEvent{Type: EventLog, Message: "ok", Prompt: &Prompt{}},
			wantErr: true,
		},
		{
			name:    "unknown type errors",
			event:   GameEvent{Type: GameEventType("Unknown")},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.event.Validate()
			if tc.wantErr && err == nil {
				t.Fatal("expected validation error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected validation error: %v", err)
			}
		})
	}
}
