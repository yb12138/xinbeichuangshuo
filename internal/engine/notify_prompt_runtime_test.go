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
