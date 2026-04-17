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
	// 灵媒师（spirit_caster）路由尚未迁移
	"sc_hundred_night_exclude_pick": ChoiceRouteRole("spirit_caster"),
	"sc_hundred_night_fire_reveal":  ChoiceRouteRole("spirit_caster"),
	"sc_hundred_night_power":        ChoiceRouteRole("spirit_caster"),
	"sc_hundred_night_target":       ChoiceRouteTargetPrompt("sc"),
	"sc_incant_card":                ChoiceRouteRole("spirit_caster"),
	"sc_incant_confirm":             ChoiceRouteRole("spirit_caster"),
	"sc_spiritual_collapse_confirm": ChoiceRouteRole("spirit_caster"),
	"sc_talisman_wind_discard":      ChoiceRouteRole("spirit_caster"),
	"sc_talisman_pick":              ChoiceRouteRole("spirit_caster"),
	// 以下角色路由已迁移到 player/<role>/module.go:
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
}

func init() {
	mountPlayerChoiceRouteSpecs(catalogChoiceRouteSpecTable)
}
