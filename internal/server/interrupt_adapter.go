package server

import (
	"starcup-engine/internal/model"
	"starcup-engine/internal/server/prompting"
)

func clonePrompt(src *model.Prompt) *model.Prompt {
	return prompting.ClonePrompt(src)
}

func buildRequireActionPayload(prompt *model.Prompt) RequireActionPayload {
	return prompting.BuildRequireActionPayload(prompt)
}
