package viewmodel

import (
	"strconv"
	"strings"

	"starcup-engine/internal/model"
)

func presentationForPrompt(p *model.Prompt) *model.PromptPresentation {
	if p == nil {
		return nil
	}
	presentation := model.PromptPresentation{}
	if p.Presentation != nil {
		presentation = *p.Presentation
	}
	if presentation.Kind == "" {
		presentation.Kind = inferPresentationKind(p)
	}
	if presentation.Layout == "" {
		presentation.Layout = inferPresentationLayout(p, presentation.Kind)
	}
	if presentation.Kind == model.PresentationCardPicker && presentation.CardSource == "" {
		presentation.CardSource = inferCardSource(p)
	}
	if presentation.Kind == model.PresentationTargetPicker && presentation.TargetFilter == "" {
		presentation.TargetFilter = "custom"
	}
	if presentation.CancelPolicy == "" {
		presentation.CancelPolicy = inferCancelPolicy(p)
	}
	if declineIndex := promptDeclineIndex(p); declineIndex >= 0 {
		presentation.HasDecline = true
		presentation.DeclineIndex = declineIndex
	}
	return &presentation
}

func inferPresentationKind(p *model.Prompt) model.PresentationKind {
	if p.UIMode == model.PromptUIModeActionHub {
		return model.PresentationActionHub
	}
	switch p.Type {
	case model.PromptChooseSkill:
		return model.PresentationSkillChoice
	case model.PromptChooseExtract:
		return model.PresentationBranchSelect
	case model.PromptChooseCards:
		return model.PresentationCardPicker
	case model.PromptConfirm:
		if promptHasResponseOptions(p) {
			return model.PresentationResponse
		}
		return model.PresentationBranchSelect
	default:
		return model.PresentationBranchSelect
	}
}

func inferPresentationLayout(p *model.Prompt, kind model.PresentationKind) string {
	switch kind {
	case model.PresentationResponse:
		return "inline"
	case model.PresentationSkillChoice, model.PresentationBranchSelect, model.PresentationNumeric:
		return "overlay"
	case model.PresentationCardPicker:
		if inferCardSource(p) == "field" {
			return "field_cover"
		}
		return "inline"
	default:
		return ""
	}
}

func inferCardSource(p *model.Prompt) string {
	for _, option := range p.Options {
		if option.FieldIndex != nil {
			return "field"
		}
	}
	return "hand"
}

func inferCancelPolicy(p *model.Prompt) string {
	if p == nil {
		return "deny"
	}
	if promptDeclineIndex(p) >= 0 {
		return "decline"
	}
	if p.Cancelable {
		return "abort"
	}
	return "deny"
}

func promptHasResponseOptions(p *model.Prompt) bool {
	if p == nil {
		return false
	}
	for _, option := range p.Options {
		switch strings.ToLower(strings.TrimSpace(option.ID)) {
		case "take", "take_damage", "defend", "counter":
			return true
		}
	}
	return false
}

func promptDeclineIndex(p *model.Prompt) int {
	if p == nil {
		return -1
	}
	for index, option := range p.Options {
		switch strings.ToLower(strings.TrimSpace(option.ID)) {
		case "-1", "skip", "cancel", "refuse", "decline", "pass", "cannot_act":
			return index
		}
	}
	if p.Presentation != nil && p.Presentation.HasDecline {
		if p.Presentation.DeclineIndex >= 0 && p.Presentation.DeclineIndex < len(p.Options) {
			return p.Presentation.DeclineIndex
		}
	}
	return -1
}

func promptOptionButtonLabel(p *model.Prompt, presentation *model.PromptPresentation, option model.PromptOption, optionIndex int) string {
	if label := strings.TrimSpace(option.ButtonLabel); label != "" {
		return label
	}
	id := strings.ToLower(strings.TrimSpace(option.ID))
	switch id {
	case "take", "take_damage":
		return "命中"
	case "defend":
		return "防御"
	case "counter":
		return "应战"
	case "skip":
		return "跳过"
	case "cancel", "refuse", "decline", "pass":
		return "取消"
	case "cannot_act":
		return "无法行动"
	case "-1":
		if presentation != nil && presentation.CancelPolicy == "decline" {
			return "不发动"
		}
		return "取消"
	}
	if presentation != nil {
		switch presentation.Kind {
		case model.PresentationNumeric:
			if n, ok := parseNonNegativePromptOptionID(option.ID); ok {
				if presentation.NumericBase == 0 {
					return strconv.Itoa(n)
				}
				return strconv.Itoa(n + presentation.NumericBase)
			}
		case model.PresentationSkillChoice:
			return "发动"
		case model.PresentationBranchSelect:
			if presentation.HasDecline && optionIndex == presentation.DeclineIndex {
				if presentation.CancelPolicy == "decline" {
					return "不发动"
				}
				return "取消"
			}
			if label := strings.TrimSpace(option.Label); label != "" {
				return label
			}
		case model.PresentationCardPicker:
			if presentation.CardSource == "field" {
				if label := strings.TrimSpace(option.Label); label != "" {
					return label
				}
				return "选择"
			}
			return "选择"
		}
	}
	if label := strings.TrimSpace(option.Label); label != "" {
		return label
	}
	return strings.TrimSpace(option.ID)
}

func parseNonNegativePromptOptionID(id string) (int, bool) {
	normalized := strings.TrimSpace(id)
	if normalized == "" {
		return 0, false
	}
	for _, r := range normalized {
		if r < '0' || r > '9' {
			return 0, false
		}
	}
	n, err := strconv.Atoi(normalized)
	if err != nil || n < 0 {
		return 0, false
	}
	return n, true
}
