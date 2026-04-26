// gameflow: types.PolicyHost 的 engine 适配实现（技能使用策略）。

package engine

import (
	"fmt"

	"starcup-engine/internal/types"
)

type enginePolicyHost struct {
	e *GameEngine
}

func (h enginePolicyHost) Log(message string) {
	if h.e == nil {
		return
	}
	h.e.Log(message)
}

func (h enginePolicyHost) BeginSkillFollowup(req types.BeginSkillFollowupReq) error {
	if h.e == nil || h.e.State == nil {
		return fmt.Errorf("引擎未初始化")
	}
	switch req.Kind {
	default:
		return fmt.Errorf("未知后续流程类型: %s", req.Kind)
	}
}

var _ types.PolicyHost = enginePolicyHost{}
