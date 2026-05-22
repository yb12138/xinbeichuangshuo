package engine

import (
	"testing"

	"starcup-engine/internal/model"
)

func TestHandOverflowDiscardPromptCarriesOverflowReason(t *testing.T) {
	game := NewGameEngine(nil)
	if err := game.AddPlayer("p1", "Overflow", "fighter", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	player := game.State.Players["p1"]
	player.Hand = []model.Card{
		{ID: "c1", Name: "A", Type: model.CardTypeAttack, Element: model.ElementFire},
		{ID: "c2", Name: "B", Type: model.CardTypeMagic, Element: model.ElementWater},
	}

	game.pushHandOverflowDiscardInterrupt(player, 1, handOverflowContext{})
	prompt := game.BuildPendingInterruptPrompt()
	if prompt == nil || prompt.Presentation == nil {
		t.Fatalf("expected overflow discard prompt, got %+v", prompt)
	}
	if prompt.Presentation.CardFilter != "overflow_discard" {
		t.Fatalf("expected overflow card filter, got %+v", prompt.Presentation)
	}
	if prompt.Presentation.DiscardReason != discardReasonHandOverflow {
		t.Fatalf("expected hand overflow discard reason, got %+v", prompt.Presentation)
	}
}

func TestFixedDiscardCountDefaultsToNonOverflowPrompt(t *testing.T) {
	game := NewGameEngine(nil)
	if err := game.AddPlayer("p1", "Skill", "fighter", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	player := game.State.Players["p1"]
	player.Hand = []model.Card{{ID: "c1", Name: "A", Type: model.CardTypeAttack, Element: model.ElementFire}}

	prompt := game.buildDiscardChoicePromptFromData(player.ID, map[string]interface{}{
		"choice_type":   choiceTypeSystemDiscardCards,
		"discard_count": 1,
		"skill_id":      "sample_skill",
	})
	if prompt == nil || prompt.Presentation == nil {
		t.Fatalf("expected skill discard prompt, got %+v", prompt)
	}
	if prompt.Presentation.CardFilter != "discard" {
		t.Fatalf("expected regular discard filter, got %+v", prompt.Presentation)
	}
	if prompt.Presentation.DiscardReason != discardReasonSkillEffect {
		t.Fatalf("expected skill effect discard reason, got %+v", prompt.Presentation)
	}
}
