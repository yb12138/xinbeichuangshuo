package engine

import "starcup-engine/internal/engine/promptfmt"

func elementNameForPrompt(raw string) string {
	return promptfmt.ElementName(raw)
}
