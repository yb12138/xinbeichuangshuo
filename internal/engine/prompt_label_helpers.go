package engine

import "starcup-engine/internal/engine/hook/promptfmt"

func elementNameForPrompt(raw string) string {
	return promptfmt.ElementName(raw)
}
