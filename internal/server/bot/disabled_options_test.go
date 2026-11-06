package bot

import (
	"testing"

	"starcup-engine/internal/model"
)

func TestWithoutDisabledOptions_FiltersAndMapsBack(t *testing.T) {
	prompt := &model.Prompt{
		Type: model.PromptConfirm,
		Options: []model.PromptOption{
			{ID: "0", Label: "分支①"},
			{ID: "1", Label: "分支②", Disabled: true, DisabledReason: "需要至少1个蛹"},
			{ID: "2", Label: "分支③"},
		},
	}
	visible, indexMap := withoutDisabledOptions(prompt)
	if visible == prompt {
		t.Fatal("expected a filtered copy when an option is disabled")
	}
	if len(visible.Options) != 2 || visible.Options[1].ID != "2" {
		t.Fatalf("unexpected visible options: %+v", visible.Options)
	}
	if len(prompt.Options) != 3 {
		t.Fatal("original prompt must not be mutated")
	}
	if got := remapSelections([]int{1}, indexMap); got[0] != 2 {
		t.Fatalf("filtered index 1 should map back to original index 2, got %v", got)
	}
}

func TestWithoutDisabledOptions_NoDisabledReturnsSamePrompt(t *testing.T) {
	prompt := &model.Prompt{Options: []model.PromptOption{{ID: "0"}, {ID: "1"}}}
	visible, indexMap := withoutDisabledOptions(prompt)
	if visible != prompt || indexMap != nil {
		t.Fatal("prompts without disabled options should pass through untouched")
	}
	if got := remapSelections([]int{1}, nil); got[0] != 1 {
		t.Fatalf("nil map must keep selections, got %v", got)
	}
}
