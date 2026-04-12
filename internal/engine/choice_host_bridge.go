// gameflow: 将 *GameEngine 注入 runtime/choice（无 Legacy 回退）。

package engine

import choicert "starcup-engine/internal/engine/runtime/choice"

type choiceHostBridge struct {
	e *GameEngine
}

var _ choicert.Host = (*choiceHostBridge)(nil)

func (*choiceHostBridge) ChoiceEngineHost() {}
