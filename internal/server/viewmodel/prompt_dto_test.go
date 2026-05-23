package viewmodel

import (
	"encoding/json"
	"testing"

	"starcup-engine/internal/model"
)

func TestToPromptDTORequiresExplicitPresentationAndCarriesButtonLabels(t *testing.T) {
	prompt := &model.Prompt{
		Type:     model.PromptConfirm,
		PlayerID: "p1",
		Message:  "请选择分支",
		Options: []model.PromptOption{
			{ID: "0", Label: "黄色灵魂转蓝色灵魂"},
			{ID: "-1", Label: "不发动"},
		},
		Min: 1,
		Max: 1,
		Presentation: &model.PromptPresentation{
			Kind:         model.PresentationBranchSelect,
			Layout:       "overlay",
			CancelPolicy: "decline",
			CancelLabel:  "不发动",
			HasDecline:   true,
			DeclineIndex: 1,
		},
	}

	dto := ToPromptDTO(prompt)
	if dto.Presentation == nil {
		t.Fatal("expected presentation to be generated")
	}
	if dto.Presentation.Kind != model.PresentationBranchSelect {
		t.Fatalf("expected branch_select presentation, got %+v", dto.Presentation)
	}
	if dto.Presentation.CancelPolicy != "decline" || !dto.Presentation.HasDecline || dto.Presentation.DeclineIndex != 1 {
		t.Fatalf("expected decline cancel policy, got %+v", dto.Presentation)
	}
	if got := dto.Options[0].ButtonLabel; got != "黄色灵魂转蓝色灵魂" {
		t.Fatalf("expected branch button label from backend, got %q", got)
	}
	if got := dto.Options[1].ButtonLabel; got != "不发动" {
		t.Fatalf("expected decline button label from backend, got %q", got)
	}
	assertPromptInteraction(t, dto.Interaction, "option", "option_index", "immediate", "select")

	raw, err := json.Marshal(dto)
	if err != nil {
		t.Fatalf("marshal prompt dto: %v", err)
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("decode prompt dto: %v", err)
	}
	if _, ok := decoded["presentation"]; !ok {
		t.Fatalf("expected presentation in JSON: %s", raw)
	}
}

func TestToPromptDTORequiresCardPickerFieldPresentation(t *testing.T) {
	fieldIndex := 2
	prompt := &model.Prompt{
		Type:     model.PromptChooseCards,
		PlayerID: "p1",
		Message:  "请选择盖牌",
		Options: []model.PromptOption{
			{ID: "0", Label: "移除茧[2]", FieldIndex: &fieldIndex},
		},
		Min:          1,
		Max:          1,
		Presentation: &model.PromptPresentation{Kind: model.PresentationCardPicker, Layout: "field_cover", CardSource: "field"},
	}

	dto := ToPromptDTO(prompt)
	if dto.Presentation == nil {
		t.Fatal("expected presentation to be generated")
	}
	if dto.Presentation.Kind != model.PresentationCardPicker || dto.Presentation.CardSource != "field" {
		t.Fatalf("expected field card_picker presentation, got %+v", dto.Presentation)
	}
	if got := dto.Options[0].ButtonLabel; got != "移除茧[2]" {
		t.Fatalf("expected field option button label, got %q", got)
	}
	assertPromptInteraction(t, dto.Interaction, "field", "option_index", "manual", "select")
}

func TestToPromptDTOPanicsWhenPresentationMissing(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected missing presentation to panic")
		}
	}()
	ToPromptDTO(&model.Prompt{Type: model.PromptConfirm, PlayerID: "p1"})
}

func TestToPromptDTOPreservesCardIDOnCardOptions(t *testing.T) {
	prompt := &model.Prompt{
		Type:     model.PromptChooseCards,
		PlayerID: "p1",
		Message:  "请选择手牌",
		Options: []model.PromptOption{
			{ID: "0", Label: "1: 测试牌", CardID: "card-001"},
		},
		Min:          1,
		Max:          1,
		Presentation: &model.PromptPresentation{Kind: model.PresentationCardPicker, CardSource: "hand"},
	}

	dto := ToPromptDTO(prompt)
	if got := dto.Options[0].CardID; got != "card-001" {
		t.Fatalf("expected card_id to be preserved, got %q", got)
	}
	assertPromptInteraction(t, dto.Interaction, "hand", "card_id", "manual", "select")
}

func TestToPromptDTOPanicsWhenHandCardPickerOptionMissingCardID(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected missing card_id to panic")
		}
	}()
	ToPromptDTO(&model.Prompt{
		Type:     model.PromptChooseCards,
		PlayerID: "p1",
		Message:  "请选择手牌",
		Options: []model.PromptOption{
			{ID: "0", Label: "1: 测试牌"},
		},
		Min:          1,
		Max:          1,
		Presentation: &model.PromptPresentation{Kind: model.PresentationCardPicker, CardSource: "hand"},
	})
}

func TestToPromptDTOAllowsCardPickerControlOptionWithoutCardID(t *testing.T) {
	prompt := &model.Prompt{
		Type:     model.PromptChooseCards,
		PlayerID: "p1",
		Message:  "请选择或取消",
		Options: []model.PromptOption{
			{ID: "-1", Label: "不弃置"},
			{ID: "0", Label: "1: 测试牌", CardID: "card-001"},
		},
		Min:          1,
		Max:          1,
		Presentation: &model.PromptPresentation{Kind: model.PresentationCardPicker, CardSource: "hand", HasDecline: true},
	}

	dto := ToPromptDTO(prompt)
	if got := dto.Options[0].CardID; got != "" {
		t.Fatalf("expected control option card_id to stay empty, got %q", got)
	}
	if got := dto.Options[1].CardID; got != "card-001" {
		t.Fatalf("expected card option card_id to be preserved, got %q", got)
	}
}

func TestToPromptDTOPreservesTargetIDOnTargetOptions(t *testing.T) {
	prompt := &model.Prompt{
		Type:     model.PromptConfirm,
		PlayerID: "p1",
		Message:  "请选择目标",
		Options: []model.PromptOption{
			{ID: "0", Label: "任意展示文案", TargetID: "p3"},
		},
		Min:          1,
		Max:          1,
		Presentation: &model.PromptPresentation{Kind: model.PresentationTargetPicker, TargetFilter: "custom"},
	}

	dto := ToPromptDTO(prompt)
	if dto.Presentation == nil || dto.Presentation.Kind != model.PresentationTargetPicker || dto.Presentation.TargetFilter != "custom" {
		t.Fatalf("expected custom target_picker presentation, got %+v", dto.Presentation)
	}
	if got := dto.Options[0].TargetID; got != "p3" {
		t.Fatalf("expected target_id to be preserved, got %q", got)
	}
	assertPromptInteraction(t, dto.Interaction, "target", "option_index", "immediate", "select")
}

func TestToPromptDTOInteractionForMultiTargetPickerIsManual(t *testing.T) {
	prompt := &model.Prompt{
		Type:     model.PromptConfirm,
		PlayerID: "p1",
		Message:  "请选择2名目标",
		Options: []model.PromptOption{
			{ID: "0", Label: "目标A", TargetID: "p2"},
			{ID: "1", Label: "目标B", TargetID: "p3"},
		},
		Min:          2,
		Max:          2,
		Presentation: &model.PromptPresentation{Kind: model.PresentationTargetPicker, TargetFilter: "custom", MultiTarget: true},
	}

	dto := ToPromptDTO(prompt)
	assertPromptInteraction(t, dto.Interaction, "target", "option_index", "manual", "select")
}

func assertPromptInteraction(t *testing.T, got *PromptInteractionDTO, source, value, confirmMode, submitAction string) {
	t.Helper()
	if got == nil {
		t.Fatal("expected prompt interaction")
	}
	if got.SelectionSource != source || got.SelectionValue != value || got.ConfirmMode != confirmMode || got.SubmitAction != submitAction {
		t.Fatalf("unexpected prompt interaction: got %+v, want source=%q value=%q confirm=%q submit=%q", got, source, value, confirmMode, submitAction)
	}
}

func TestToPromptDTOPanicsWhenTargetPickerOptionMissingTargetID(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected target_picker option without target_id to panic")
		}
	}()

	ToPromptDTO(&model.Prompt{
		Type:       model.PromptConfirm,
		PlayerID:   "p1",
		Message:    "请选择目标",
		ChoiceType: "broken_target_prompt",
		Options: []model.PromptOption{
			{ID: "0", Label: "目标"},
		},
		Min:          1,
		Max:          1,
		Presentation: &model.PromptPresentation{Kind: model.PresentationTargetPicker, TargetFilter: "custom"},
	})
}

func TestToPromptDTOAllowsTargetPickerControlOptionWithoutTargetID(t *testing.T) {
	prompt := &model.Prompt{
		Type:       model.PromptConfirm,
		PlayerID:   "p1",
		Message:    "请选择目标",
		ChoiceType: "finishable_target_prompt",
		Options: []model.PromptOption{
			{ID: "0", Label: "目标", TargetID: "p2"},
			{ID: "finish", Label: "完成目标选择", ButtonLabel: "完成"},
		},
		Min:          1,
		Max:          1,
		Presentation: &model.PromptPresentation{Kind: model.PresentationTargetPicker, TargetFilter: "custom"},
	}

	dto := ToPromptDTO(prompt)
	if len(dto.Options) != 2 {
		t.Fatalf("expected 2 options, got %+v", dto.Options)
	}
	if got := dto.Options[1].ID; got != "finish" {
		t.Fatalf("expected finish control option to be preserved, got %+v", dto.Options[1])
	}
}
