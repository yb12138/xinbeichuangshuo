package engine

// catalogChoiceRouteSpecTable 与 choice_type_catalog.txt 一一对应（显式 type→路由规格，不做前缀推断）。
var catalogChoiceRouteSpecTable = map[string]ChoiceRouteSpec{
	// 系统级路由
	"basic_effect_pick":    ChoiceRouteSystem(),
	"system_discard_cards": ChoiceRouteSystem(),
	"buy_resource":         ChoiceRouteSystem(),
	"heal":                 ChoiceRouteSystem(),
	"weak":                 ChoiceRouteSystem(),
}
