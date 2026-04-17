// gameflow: Choice 路由类型定义（共享包，避免 engine/player 循环依赖）。

package types

// ChoiceRouteKind 定义 choice 路由类型枚举。
type ChoiceRouteKind string

const (
	ChoiceRouteKindRole         ChoiceRouteKind = "role"
	ChoiceRouteKindSystem       ChoiceRouteKind = "system"
	ChoiceRouteKindTargetPrompt ChoiceRouteKind = "target_prompt"
	ChoiceRouteKindSpecial      ChoiceRouteKind = "special"
)

// ChoiceRouteSpec 定义 choice_type 的路由目的地（角色路由/系统路由/目标路由/特殊路由）。
type ChoiceRouteSpec struct {
	Kind         ChoiceRouteKind
	Role         string
	TargetPrompt string
	Special      string
}

// ChoiceRouteRole 返回角色路由规格。
func ChoiceRouteRole(role string) ChoiceRouteSpec {
	return ChoiceRouteSpec{
		Kind: ChoiceRouteKindRole,
		Role: role,
	}
}

// ChoiceRouteSystem 返回系统路由规格。
func ChoiceRouteSystem() ChoiceRouteSpec {
	return ChoiceRouteSpec{
		Kind: ChoiceRouteKindSystem,
	}
}

// ChoiceRouteTargetPrompt 返回目标模板路由规格。
func ChoiceRouteTargetPrompt(route string) ChoiceRouteSpec {
	return ChoiceRouteSpec{
		Kind:         ChoiceRouteKindTargetPrompt,
		TargetPrompt: route,
	}
}

// ChoiceRouteSpecial 返回特殊路由规格。
func ChoiceRouteSpecial(special string) ChoiceRouteSpec {
	return ChoiceRouteSpec{
		Kind:    ChoiceRouteKindSpecial,
		Special: special,
	}
}

// Valid 检查路由规格是否有效。
func (s ChoiceRouteSpec) Valid() bool {
	switch s.Kind {
	case ChoiceRouteKindRole:
		return s.Role != ""
	case ChoiceRouteKindSystem:
		return true
	case ChoiceRouteKindTargetPrompt:
		return s.TargetPrompt != ""
	case ChoiceRouteKindSpecial:
		return s.Special != ""
	default:
		return false
	}
}
