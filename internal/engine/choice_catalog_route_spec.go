package engine

import "starcup-engine/internal/types"

// ChoiceRouteKind 是 types.ChoiceRouteKind 的别名，保持 engine 包内兼容性。
type ChoiceRouteKind = types.ChoiceRouteKind

// ChoiceRouteSpec 是 types.ChoiceRouteSpec 的别名。
type ChoiceRouteSpec = types.ChoiceRouteSpec

// 导出常量（通过别名自动可用）。
const (
	ChoiceRouteKindRole   = types.ChoiceRouteKindRole
	ChoiceRouteKindSystem = types.ChoiceRouteKindSystem
)

// ChoiceRouteRole 调用 types 包的构造函数。
func ChoiceRouteRole(role string) ChoiceRouteSpec {
	return types.ChoiceRouteRole(role)
}

// ChoiceRouteSystem 调用 types 包的构造函数。
func ChoiceRouteSystem() ChoiceRouteSpec {
	return types.ChoiceRouteSystem()
}
