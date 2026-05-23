// gameflow: Choice 路由类型定义（共享包，避免 engine/player 循环依赖）。

package types

// ChoiceRouteKind 定义 choice 路由类型枚举。
type ChoiceRouteKind string

const (
	ChoiceRouteKindRole   ChoiceRouteKind = "role"
	ChoiceRouteKindSystem ChoiceRouteKind = "system"
)

// ChoiceRouteSpec 定义 choice_type 的路由目的地（角色路由/系统路由）。
type ChoiceRouteSpec struct {
	Kind      ChoiceRouteKind
	Role      string
	PhaseSync string
}

// ChoiceRouteRole 返回角色路由规格。
func ChoiceRouteRole(role string) ChoiceRouteSpec {
	return ChoiceRouteSpec{
		Kind: ChoiceRouteKindRole,
		Role: role,
	}
}

// ChoiceRouteRoleWithPhaseSync 返回带阶段同步声明的角色路由规格。
func ChoiceRouteRoleWithPhaseSync(role, phaseSync string) ChoiceRouteSpec {
	return ChoiceRouteSpec{
		Kind:      ChoiceRouteKindRole,
		Role:      role,
		PhaseSync: phaseSync,
	}
}

// ChoiceRouteSystem 返回系统路由规格。
func ChoiceRouteSystem() ChoiceRouteSpec {
	return ChoiceRouteSpec{
		Kind: ChoiceRouteKindSystem,
	}
}

// Valid 检查路由规格是否有效。
func (s ChoiceRouteSpec) Valid() bool {
	switch s.Kind {
	case ChoiceRouteKindRole:
		return s.Role != ""
	case ChoiceRouteKindSystem:
		return true
	default:
		return false
	}
}
