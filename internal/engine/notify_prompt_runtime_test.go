package engine

import (
	"testing"

	"starcup-engine/internal/model"
)

func TestDecoratePromptForClient_BranchSelectUsesFullLabels(t *testing.T) {
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
		if option.ButtonLabel != option.Label {
			t.Fatalf("option %d should use full branch label as button label, got button=%q label=%q", i, option.ButtonLabel, option.Label)
		}
		if option.Hint != "" {
			t.Fatalf("option %d should not duplicate branch label into hint, got %q", i, option.Hint)
		}
	}
}
