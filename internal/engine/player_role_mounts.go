// gameflow: 将 player 目录角色声明挂载到 engine 全局表。

package engine

import (
	"strings"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

func mountPlayerChoiceRouteSpecs(table map[string]ChoiceRouteSpec) {
	if table == nil || roleRegistry == nil {
		return
	}
	for _, entry := range roleRegistry.Entries() {
		for choiceType, route := range entry.ChoiceRoutes() {
			// 直接赋值，无需转换（类型已统一）
			table[choiceType] = route
		}
	}
}

func mountPlayerSkillPolicySpecs(table map[string]SkillPolicy) {
	if table == nil || roleRegistry == nil {
		return
	}
	for _, entry := range roleRegistry.Entries() {
		for skillID, policy := range entry.SkillPolicies() {
			// 直接赋值，无需转换（类型已统一）
			table[skillID] = policy
		}
	}
}

func mountPlayerDeferredFollowupSpecs(table map[string]deferredFollowupHandler) {
	if table == nil || roleRegistry == nil {
		return
	}
	for _, entry := range roleRegistry.Entries() {
		for followupType, spec := range entry.Followups() {
			handler, ok := toDeferredFollowupHandler(spec)
			if !ok {
				continue
			}
			table[followupType] = handler
		}
	}
}

func toDeferredFollowupHandler(spec engineplayer.FollowupSpec) (deferredFollowupHandler, bool) {
	if spec.Resolve == nil {
		return deferredFollowupHandler{}, false
	}
	label := strings.TrimSpace(spec.Label)
	if label == "" {
		label = "RoleFollowup"
	}
	return deferredFollowupHandler{
		label: label,
		resolve: func(e *GameEngine, f model.DeferredFollowup) error {
			return spec.Resolve(engineFollowupHost{e: e}, f)
		},
	}, true
}
