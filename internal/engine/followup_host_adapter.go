// gameflow: player.FollowupHost 的 engine 适配实现。

package engine

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

type engineFollowupHost struct {
	e *GameEngine
}

func (h engineFollowupHost) Log(message string) {
	if h.e == nil {
		return
	}
	h.e.Log(message)
}

func (h engineFollowupHost) LookupPlayer(playerID string) *model.Player {
	if h.e == nil || h.e.State == nil || playerID == "" {
		return nil
	}
	return h.e.State.Players[playerID]
}

func (h engineFollowupHost) ResolveSkillFollowup(req engineplayer.ResolveSkillFollowupReq) error {
	if h.e == nil || h.e.State == nil {
		return fmt.Errorf("引擎未初始化")
	}
	switch req.Kind {
	default:
		return fmt.Errorf("未知后续流程类型: %s", req.Kind)
	}
}

var _ engineplayer.FollowupHost = engineFollowupHost{}
