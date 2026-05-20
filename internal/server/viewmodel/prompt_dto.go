package viewmodel

import (
	"fmt"
	"strings"

	"starcup-engine/internal/model"
)

// PromptDTO 前端 Prompt 数据传输对象（wire DTO）
// 与 model.Prompt 解耦，仅保留前端渲染需要的字段
type PromptDTO struct {
	Type             string                    `json:"type"`
	PlayerID         string                    `json:"player_id"`
	Message          string                    `json:"message"`
	ChoiceType       string                    `json:"choice_type,omitempty"`
	SkillID          string                    `json:"skill_id,omitempty"`
	Options          []PromptOptionDTO         `json:"options"`
	SpecialOptions   []PromptOptionDTO         `json:"special_options,omitempty"`
	UIMode           string                    `json:"ui_mode,omitempty"`
	Presentation     *model.PromptPresentation `json:"presentation"`
	EffectHints      []string                  `json:"effect_hints,omitempty"`
	Min              int                       `json:"min"`
	Max              int                       `json:"max"`
	AttackerID       string                    `json:"attacker_id,omitempty"`
	CounterTargetIDs []string                  `json:"counter_target_ids,omitempty"`
	AttackElement    string                    `json:"attack_element,omitempty"`
}

// PromptOptionDTO 选项 DTO
type PromptOptionDTO struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	ButtonLabel string `json:"button_label"`
	Hint        string `json:"hint,omitempty"`
	CardID      string `json:"card_id,omitempty"`
	TargetID    string `json:"target_id,omitempty"`
	FieldIndex  *int   `json:"field_index,omitempty"`
	Element     string `json:"element,omitempty"`
}

// ToPromptDTO 将 model.Prompt 转换为 PromptDTO
func ToPromptDTO(p *model.Prompt) *PromptDTO {
	if p == nil {
		return nil
	}
	presentation := presentationForPrompt(p)
	validateTargetPickerOptions(p, presentation)
	dto := &PromptDTO{
		Type:             string(p.Type),
		PlayerID:         p.PlayerID,
		Message:          p.Message,
		ChoiceType:       p.ChoiceType,
		SkillID:          p.SkillID,
		UIMode:           p.UIMode,
		Presentation:     presentation,
		EffectHints:      p.EffectHints,
		Min:              p.Min,
		Max:              p.Max,
		AttackerID:       p.AttackerID,
		CounterTargetIDs: p.CounterTargetIDs,
		AttackElement:    p.AttackElement,
	}
	// 转换 Options
	for _, o := range p.Options {
		optionIndex := len(dto.Options)
		dto.Options = append(dto.Options, PromptOptionDTO{
			ID:          o.ID,
			Label:       o.Label,
			ButtonLabel: promptOptionButtonLabel(p, presentation, o, optionIndex),
			Hint:        o.Hint,
			CardID:      o.CardID,
			TargetID:    o.TargetID,
			FieldIndex:  o.FieldIndex,
			Element:     o.Element,
		})
	}
	// 转换 SpecialOptions
	for _, o := range p.SpecialOptions {
		optionIndex := len(dto.SpecialOptions)
		dto.SpecialOptions = append(dto.SpecialOptions, PromptOptionDTO{
			ID:          o.ID,
			Label:       o.Label,
			ButtonLabel: promptOptionButtonLabel(p, presentation, o, optionIndex),
			Hint:        o.Hint,
			CardID:      o.CardID,
			TargetID:    o.TargetID,
			FieldIndex:  o.FieldIndex,
			Element:     o.Element,
		})
	}
	return dto
}

func validateTargetPickerOptions(p *model.Prompt, presentation *model.PromptPresentation) {
	if p == nil || presentation == nil || presentation.Kind != model.PresentationTargetPicker {
		return
	}
	validateTargetPickerOptionList(p, p.Options, "options")
	validateTargetPickerOptionList(p, p.SpecialOptions, "special_options")
}

func validateTargetPickerOptionList(p *model.Prompt, options []model.PromptOption, fieldName string) {
	for i, o := range options {
		if strings.TrimSpace(o.TargetID) != "" {
			continue
		}
		if isTargetPickerControlOption(o.ID) {
			continue
		}
		panic(fmt.Sprintf("target_picker prompt %q for player %q %s[%d] (%q) is missing target_id", p.ChoiceType, p.PlayerID, fieldName, i, o.ID))
	}
}

func isTargetPickerControlOption(id string) bool {
	switch strings.ToLower(strings.TrimSpace(id)) {
	case "-1", "cancel", "decline", "refuse", "skip", "pass", "back", "done", "finish":
		return true
	default:
		return false
	}
}
