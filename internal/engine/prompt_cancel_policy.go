package engine

import "starcup-engine/internal/model"

func promptCancelPolicyFromContext(ctxData map[string]interface{}) string {
	if ctxData == nil {
		return ""
	}
	prompt, ok := ctxData["prompt"].(*model.Prompt)
	if !ok || prompt == nil || prompt.Presentation == nil {
		return ""
	}
	return prompt.Presentation.CancelPolicy
}

func isPromptCancelAllowed(ctxData map[string]interface{}) bool {
	switch promptCancelPolicyFromContext(ctxData) {
	case model.CancelPolicyAbort, model.CancelPolicyBack, model.CancelPolicyDecline:
		return true
	default:
		return false
	}
}
