package engine

import "starcup-engine/internal/types"

// ChoiceRouteKind 表示 choice 路由的类型。
type ChoiceRouteKind = types.ChoiceRouteKind

// ChoiceRouteSpec 表示 choice 路由的声明。
type ChoiceRouteSpec = types.ChoiceRouteSpec

// 路由类型常量。
const (
	ChoiceRouteKindRole   = types.ChoiceRouteKindRole
	ChoiceRouteKindSystem = types.ChoiceRouteKindSystem
)

// ChoiceRouteRole 创建角色 choice 路由规格。
func ChoiceRouteRole(role string) ChoiceRouteSpec {
	return types.ChoiceRouteRole(role)
}

// ChoiceRouteSystem 创建系统 choice 路由规格。
func ChoiceRouteSystem() ChoiceRouteSpec {
	return types.ChoiceRouteSystem()
}
