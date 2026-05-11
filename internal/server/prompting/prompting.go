package prompting

import (
	"starcup-engine/internal/model"
	"starcup-engine/internal/server/protocol"
	"starcup-engine/internal/server/viewmodel"
)

func ClonePrompt(src *model.Prompt) *model.Prompt {
	if src == nil {
		return nil
	}
	cp := *src
	if src.Options != nil {
		cp.Options = append([]model.PromptOption{}, src.Options...)
	}
	if src.SpecialOptions != nil {
		cp.SpecialOptions = append([]model.PromptOption{}, src.SpecialOptions...)
	}
	if src.CounterTargetIDs != nil {
		cp.CounterTargetIDs = append([]string{}, src.CounterTargetIDs...)
	}
	return &cp
}

func BuildRequireActionPayload(prompt *model.Prompt) protocol.RequireActionPayload {
	payload := protocol.RequireActionPayload{
		InterruptType: InferInterruptType(prompt),
		TargetUserID:  prompt.PlayerID,
		Timeout:       0,
		Msg:           prompt.Message,
		ValidActions:  []string{protocol.CmdSubmitAction},
		RequireCount:  prompt.Max,
		PromptType:    string(prompt.Type),
		Prompt:        viewmodel.ToPromptDTO(prompt),
	}
	if payload.RequireCount <= 0 {
		payload.RequireCount = prompt.Min
	}
	return payload
}

func InferInterruptType(prompt *model.Prompt) string {
	if prompt == nil {
		return "WaitChoice"
	}
	switch prompt.Type {
	case model.PromptChooseCards:
		return "WaitDiscard"
	case model.PromptChooseSkill, model.PromptChooseExtract:
		return "WaitChoice"
	case model.PromptConfirm:
		for _, opt := range prompt.Options {
			switch opt.ID {
			case "take", "counter", "defend":
				return "WaitResponse"
			}
		}
		return "WaitChoice"
	default:
		return "WaitChoice"
	}
}
