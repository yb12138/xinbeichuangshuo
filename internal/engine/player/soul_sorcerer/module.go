// gameflow: 灵魂术士模块入口声明。

package soul_sorcerer

import (
	"starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/types"
)

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "ss_soul_devour", Handler: &skills.SoulSorcererSoulDevourHandler{}},
		{ID: "ss_soul_recall", Handler: &skills.SoulSorcererSoulRecallHandler{}},
		{ID: "ss_soul_convert", Handler: &skills.SoulSorcererSoulConvertHandler{}},
		{ID: "ss_soul_mirror", Handler: &skills.SoulSorcererSoulMirrorHandler{}},
		{ID: "ss_soul_blast", Handler: &skills.SoulSorcererSoulBlastHandler{}},
		{ID: "ss_soul_grant", Handler: &skills.SoulSorcererSoulGrantHandler{}},
		{
			ID:      "ss_soul_link",
			Handler: &skills.SoulSorcererSoulLinkHandler{},
			Policy: types.SkillPolicy{
				ManualExclusiveCard: true,
			},
		},
		{ID: "ss_soul_amp", Handler: &skills.SoulSorcererSoulAmpHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"ss_convert_color":    types.ChoiceRouteRole("soul_sorcerer"),
		"ss_link_target":      types.ChoiceRouteRole("soul_sorcerer"),
		"ss_link_transfer_x":  types.ChoiceRouteRole("soul_sorcerer"),
		"ss_soul_devour_pick": types.ChoiceRouteRole("soul_sorcerer"),
		"ss_soul_recall_pick": types.ChoiceRouteRole("soul_sorcerer"),
	}
}
