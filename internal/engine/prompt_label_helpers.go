// gameflow: Prompt 选项展示文案。

package engine

import "starcup-engine/internal/engine/hook/promptfmt"

func elementNameForPrompt(raw string) string {
	return promptfmt.ElementName(raw)
}
