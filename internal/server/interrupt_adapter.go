package server

import "starcup-engine/internal/model"

func buildRequireActionPayload(prompt *model.Prompt) RequireActionPayload {
	payload := RequireActionPayload{
		InterruptType: inferInterruptType(prompt),
		TargetUserID:  prompt.PlayerID,
		Timeout:       0,
		Msg:           prompt.Message,
		ValidActions:  []string{CmdSubmitAction},
		RequireCount:  prompt.Max,
		PromptType:    string(prompt.Type),
		Prompt:        clonePrompt(prompt),
	}
	if payload.RequireCount <= 0 {
		payload.RequireCount = prompt.Min
	}
	return payload
}

func inferInterruptType(prompt *model.Prompt) string {
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
