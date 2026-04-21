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
	// 所有角色路由已迁移到 player/<role>/module.go:
	// - 吟游诗人（bard）- bd_*
	// - 血之巫女（blood_priestess）- bp_*
	// - 蝶舞者（butterfly_dancer）- bt_*
	// - 红莲骑士（crimson_knight）- crk_*
	// - 精灵射手（elf_archer）- elf_*
	// - 格斗家（fighter）- fighter_*
	// - 圣弓（holy_bow）- hb_*
	// - 英灵人形（war_homunculus）- hom_*
	// - 魔弓（magic_bow）- mb_*
	// - 魔枪（magic_lancer）- ml_*
	// - 神官（priest）- priest_*
	// - 祈祷师（prayer_master）- prayer_*
	// - 剑帝（sword_emperor）- se_*
	// - 灵魂术士（soul_sorcerer）- ss_*
	// - 兽灵武士（beast_samurai）- bs_*
	// - 灵符师（spirit_caster）- sc_*
}

func init() {
	mountPlayerChoiceRouteSpecs(catalogChoiceRouteSpecTable)
}
