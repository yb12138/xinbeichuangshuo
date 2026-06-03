package engine

import (
	"testing"

	"starcup-engine/internal/model"
)

func TestDecoratePromptForClient_ClonesPromptWithoutGeneratingLabels(t *testing.T) {
	game := NewGameEngine(nil)
	prompt := &model.Prompt{
		Type:     model.PromptConfirm,
		PlayerID: "p1",
		Message:  "【散华轮舞】请选择发动分支：",
		Options: []model.PromptOption{
			{ID: "0", Label: "消耗1蓝水晶（可用红宝石替代）：放置庭院并+2鲜血"},
			{ID: "1", Label: "消耗1红宝石：放置庭院并+2鲜血（上限4）且弃牌至4"},
		},
		Min: 1,
		Max: 1,
		Presentation: &model.PromptPresentation{
			Kind:   model.PresentationBranchSelect,
			Layout: "overlay",
		},
	}

	got := game.decoratePromptForClient(prompt)
	if len(got.Options) != 2 {
		t.Fatalf("expected 2 options, got %+v", got.Options)
	}
	for i, option := range got.Options {
		if option.ButtonLabel != "" {
			t.Fatalf("option %d should not synthesize button label in engine prompt decorator, got %q", i, option.ButtonLabel)
		}
		if option.Hint != "" {
			t.Fatalf("option %d should not synthesize hint in engine prompt decorator, got %q", i, option.Hint)
		}
	}
	got.Options[0].Label = "mutated"
	if prompt.Options[0].Label == "mutated" {
		t.Fatal("decoratePromptForClient should clone option slices")
	}
}

func TestGetCurrentPrompt_RebuildsCombatInteractionPrompt(t *testing.T) {
	game := NewGameEngine(nil)
	if err := game.AddPlayer("p1", "Alice", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Bob", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Cara", "hero", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	card := model.Card{ID: "attack-1", Name: "水涟斩", Type: model.CardTypeAttack, Element: model.ElementWater, Damage: 2}
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	game.State.CombatStage = model.CombatStageHitCheck
	game.State.Subflow = model.SubflowNone
	game.State.PendingInterrupt = nil
	game.State.CombatStack = []model.CombatRequest{{
		AttackerID:     "p1",
		TargetID:       "p2",
		Card:           &card,
		CanBeResponded: true,
	}}

	prompt := game.GetCurrentPrompt()
	if prompt == nil {
		t.Fatal("expected combat interaction prompt")
	}
	if prompt.PlayerID != "p2" {
		t.Fatalf("expected prompt player p2, got %q", prompt.PlayerID)
	}
	if prompt.AttackerID != "p1" {
		t.Fatalf("expected attacker p1, got %q", prompt.AttackerID)
	}
	if prompt.Presentation == nil || prompt.Presentation.Kind != model.PresentationResponse {
		t.Fatalf("expected response presentation, got %+v", prompt.Presentation)
	}
	if prompt.AttackElement != string(model.ElementWater) {
		t.Fatalf("expected attack element Water, got %q", prompt.AttackElement)
	}
	optionIDs := map[string]bool{}
	for _, option := range prompt.Options {
		optionIDs[option.ID] = true
	}
	for _, want := range []string{"take", "defend", "counter"} {
		if !optionIDs[want] {
			t.Fatalf("expected option %s in %+v", want, prompt.Options)
		}
	}
}
