package viewmodel

import (
	"encoding/json"
	"testing"

	"starcup-engine/internal/model"
)

func TestToPromptDTOAlwaysCarriesPresentationAndButtonLabels(t *testing.T) {
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

func TestToPromptDTOInfersCardPickerFieldPresentation(t *testing.T) {
	fieldIndex := 2
	prompt := &model.Prompt{
		Type:     model.PromptChooseCards,
		PlayerID: "p1",
		Message:  "请选择盖牌",
		Options: []model.PromptOption{
			{ID: "0", Label: "移除茧[2]", FieldIndex: &fieldIndex},
		},
		Min: 1,
		Max: 1,
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
}
