package engine

// catalogChoiceRouteSpecTable 与 choice_type_catalog.txt 一一对应（显式 type→路由规格，不做前缀推断）。
// 已迁移角色的路由已移到 player/<role>/module.go 的 ChoiceRouteSpecs()。
var catalogChoiceRouteSpecTable = map[string]ChoiceRouteSpec{
	// 系统级路由
	"basic_effect_pick":    ChoiceRouteSystem(),
	"system_discard_cards": ChoiceRouteSystem(),
	"buy_resource":         ChoiceRouteSystem(),
	"heal":                 ChoiceRouteSystem(),
	"weak":                 ChoiceRouteSystem(),
	"five_elements_bind":   ChoiceRouteSpecial("five_elements_bind"),
}

func init() {
	mountPlayerChoiceRouteSpecs(catalogChoiceRouteSpecTable)
}
