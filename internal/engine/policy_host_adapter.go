// gameflow: types.PolicyHost 的 engine 适配实现（技能使用策略）。

package engine

import "starcup-engine/internal/types"

type enginePolicyHost struct {
	e *GameEngine
}

func (h enginePolicyHost) Log(message string) {
	if h.e == nil {
		return
	}
	h.e.Log(message)
}

func (h enginePolicyHost) DropQueuedOverflowDiscardForPlayer(playerID string) {
	if h.e == nil {
		return
	}
	dropQueuedOverflowDiscardForPlayer(h.e, playerID)
}

var _ types.PolicyHost = enginePolicyHost{}
