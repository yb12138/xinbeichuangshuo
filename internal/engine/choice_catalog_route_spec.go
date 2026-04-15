package engine

type choiceRouteKind string

const (
	choiceRouteKindRole         choiceRouteKind = "role"
	choiceRouteKindSystem       choiceRouteKind = "system"
	choiceRouteKindTargetPrompt choiceRouteKind = "target_prompt"
	choiceRouteKindSpecial      choiceRouteKind = "special"
)

// choiceCatalogRouteSpec 定义 choice_type 的路由目的地（角色路由/系统路由/目标路由/特殊路由）。
type choiceCatalogRouteSpec struct {
	Kind         choiceRouteKind
	Role         string
	TargetPrompt string
	Special      string
}

func choiceRouteRole(role string) choiceCatalogRouteSpec {
	return choiceCatalogRouteSpec{
		Kind: choiceRouteKindRole,
		Role: role,
	}
}

func choiceRouteSystem() choiceCatalogRouteSpec {
	return choiceCatalogRouteSpec{
		Kind: choiceRouteKindSystem,
	}
}

func choiceRouteTargetPrompt(route string) choiceCatalogRouteSpec {
	return choiceCatalogRouteSpec{
		Kind:         choiceRouteKindTargetPrompt,
		TargetPrompt: route,
	}
}

func choiceRouteSpecial(special string) choiceCatalogRouteSpec {
	return choiceCatalogRouteSpec{
		Kind:    choiceRouteKindSpecial,
		Special: special,
	}
}

func (s choiceCatalogRouteSpec) valid() bool {
	switch s.Kind {
	case choiceRouteKindRole:
		return s.Role != ""
	case choiceRouteKindSystem:
		return true
	case choiceRouteKindTargetPrompt:
		return s.TargetPrompt != ""
	case choiceRouteKindSpecial:
		return s.Special != ""
	default:
		return false
	}
}
