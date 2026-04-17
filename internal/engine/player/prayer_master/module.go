// gameflow: 祈祷师模块入口声明。

package prayer_master

import (
	"starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/types"
)

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "prayer_enter_form", Handler: &skills.PrayerEnterFormHandler{}},
		{ID: "prayer_rune_gain", Handler: &skills.PrayerRuneGainHandler{}},
		{ID: "prayer_radiant_faith", Handler: &skills.PrayerRadiantFaithHandler{}},
		{ID: "prayer_dark_curse", Handler: &skills.PrayerDarkCurseHandler{}},
		{ID: "prayer_power_blessing", Handler: &skills.PrayerPowerBlessingHandler{}},
		{ID: "prayer_swift_blessing", Handler: &skills.PrayerSwiftBlessingHandler{}},
		{ID: "prayer_mana_tide", Handler: &skills.PrayerManaTideHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"prayer_master_extra_action_pick": types.ChoiceRouteRole("prayer_master"),
		"prayer_master_rune_pick":         types.ChoiceRouteRole("prayer_master"),
		"prayer_power_blessing_followup":  types.ChoiceRouteRole("prayer_master"),
		"prayer_swift_blessing_followup":  types.ChoiceRouteRole("prayer_master"),
	}
}
