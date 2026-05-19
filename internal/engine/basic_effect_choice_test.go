package engine

import (
	"testing"

	"starcup-engine/internal/model"
)

func TestBuildBasicEffectChoicePromptUsesCancelPolicy(t *testing.T) {
	prompt := buildBasicEffectChoicePrompt("p1", map[string]interface{}{
		"cancel_policy": "decline",
		"options": []basicEffectChoiceOption{
			{ID: "0", Label: "移除中毒"},
		},
	})

	if prompt == nil {
		t.Fatal("expected prompt")
	}
	if prompt.Presentation == nil || prompt.Presentation.CancelPolicy != "decline" {
		t.Fatalf("expected cancel_policy=decline, got %+v", prompt.Presentation)
	}
}

func TestBasicEffectChoiceCancelPolicyRecognition(t *testing.T) {
	tests := []struct {
		name        string
		policy      string
		expectError bool
	}{
		{name: "empty is denied", policy: "", expectError: true},
		{name: "deny is denied", policy: "deny", expectError: true},
		{name: "decline is accepted", policy: "decline", expectError: false},
		{name: "abort is accepted", policy: "abort", expectError: false},
		{name: "back is accepted", policy: "back", expectError: false},
	}

	cancel := systemChoiceCancel("basic_effect_pick")
	if cancel == nil {
		t.Fatal("expected basic_effect_pick cancel handler")
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := NewGameEngine(nil)
			game.State.Players["p1"] = &model.Player{ID: "p1", Name: "P1"}
			_, err := cancel(game, "p1", map[string]any{
				"choice_type":   "basic_effect_pick",
				"cancel_policy": tt.policy,
				"skill_name":    "测试技能",
				"options": []basicEffectChoiceOption{
					{ID: "0", Label: "移除中毒"},
				},
			})
			if tt.expectError && err == nil {
				t.Fatalf("expected error for policy %q", tt.policy)
			}
			if !tt.expectError && err != nil {
				t.Fatalf("expected policy %q to be accepted, got %v", tt.policy, err)
			}
		})
	}
}
