package viewmodel

import (
	"fmt"
	"strconv"
	"strings"

	"starcup-engine/internal/model"
)

func presentationForPrompt(p *model.Prompt) *model.PromptPresentation {
	if p == nil {
		return nil
	}
	if p.Presentation == nil {
		panic(fmt.Sprintf("prompt %q for player %q is missing presentation", p.ChoiceType, p.PlayerID))
	}
	presentation := *p.Presentation
	if presentation.Kind == "" {
		panic(fmt.Sprintf("prompt %q for player %q has empty presentation.kind", p.ChoiceType, p.PlayerID))
	}
	if presentation.Kind == model.PresentationCardPicker && presentation.CardSource == "" {
		panic(fmt.Sprintf("card_picker prompt %q for player %q is missing presentation.card_source", p.ChoiceType, p.PlayerID))
	}
	if presentation.Kind == model.PresentationTargetPicker && presentation.TargetFilter == "" {
		panic(fmt.Sprintf("target_picker prompt %q for player %q is missing presentation.target_filter", p.ChoiceType, p.PlayerID))
	}
	if presentation.CancelPolicy != "" && presentation.CancelLabel == "" {
		presentation.CancelLabel = cancelLabelForPolicy(presentation.CancelPolicy)
	}
	return &presentation
}

func cancelLabelForPolicy(policy string) string {
	switch policy {
	case "back":
		return "返回"
	case "decline":
		return "不发动"
	case "abort":
		return "取消"
	default:
		return ""
	}
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
	case "cannot_act":
		return "无法行动"
	}
	if presentation != nil {
		if presentation.HasDecline && optionIndex == presentation.DeclineIndex {
			if label := strings.TrimSpace(presentation.CancelLabel); label != "" {
				return label
			}
		}
		switch presentation.Kind {
		case model.PresentationNumeric:
			if n, ok := parseNonNegativePromptOptionID(option.ID); ok {
				if presentation.NumericBase == 0 {
					return strconv.Itoa(n)
				}
				return strconv.Itoa(n + presentation.NumericBase)
			}
		case model.PresentationSkillChoice:
			if id == "skip" || id == "cancel" || id == "decline" || id == "pass" || id == "-1" {
				if label := strings.TrimSpace(presentation.CancelLabel); label != "" {
					return label
				}
			}
			return "发动"
		case model.PresentationCardPicker:
			if presentation.CardSource == "hand" || presentation.CardSource == "proxy" {
				return "选择"
			}
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
