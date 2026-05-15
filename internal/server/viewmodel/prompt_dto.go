package viewmodel

import "starcup-engine/internal/model"

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
	Presentation     *model.PromptPresentation `json:"presentation,omitempty"`
	EffectHints      []string                  `json:"effect_hints,omitempty"`
	Cancelable       bool                      `json:"cancelable,omitempty"`
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
	ButtonLabel string `json:"button_label,omitempty"`
	Hint        string `json:"hint,omitempty"`
}

// ToPromptDTO 将 model.Prompt 转换为 PromptDTO
func ToPromptDTO(p *model.Prompt) *PromptDTO {
	if p == nil {
		return nil
	}
	dto := &PromptDTO{
		Type:             string(p.Type),
		PlayerID:         p.PlayerID,
		Message:          p.Message,
		ChoiceType:       p.ChoiceType,
		SkillID:          p.SkillID,
		UIMode:           p.UIMode,
		Presentation:     p.Presentation,
		EffectHints:      p.EffectHints,
		Cancelable:       p.Cancelable,
		Min:              p.Min,
		Max:              p.Max,
		AttackerID:       p.AttackerID,
		CounterTargetIDs: p.CounterTargetIDs,
		AttackElement:    p.AttackElement,
	}
	// 转换 Options
	for _, o := range p.Options {
		dto.Options = append(dto.Options, PromptOptionDTO{
			ID:          o.ID,
			Label:       o.Label,
			ButtonLabel: o.ButtonLabel,
			Hint:        o.Hint,
		})
	}
	// 转换 SpecialOptions
	for _, o := range p.SpecialOptions {
		dto.SpecialOptions = append(dto.SpecialOptions, PromptOptionDTO{
			ID:          o.ID,
			Label:       o.Label,
			ButtonLabel: o.ButtonLabel,
			Hint:        o.Hint,
		})
	}
	return dto
}
