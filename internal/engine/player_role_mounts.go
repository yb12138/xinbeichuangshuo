// gameflow: 将 player 目录角色声明挂载到 engine 全局表。

package engine

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
