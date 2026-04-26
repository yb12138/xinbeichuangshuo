// gameflow: 灵魂术士模块入口声明。

package soul_sorcerer

import (
	"fmt"
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:                  "soul_sorcerer",
		Defaults:            ApplyDefaults,
		StarterCards:        StarterCards,
		Choices:             NewChoiceHandler(),
		Skills:              SkillEntries(),
		ChoiceRouteSpecs:    ChoiceRouteSpecs(),
		AfterMoraleLossHook: ApplySoulDevour,
		TimingHookSpecs: []player.TimingHookSpec{
			{Timing: player.TimingOnDamageBeforeTaken, Priority: 100, Hook: damageBeforeTakenHook},
		},
	}
}

// ApplyDefaults 初始化角色默认属性。
func ApplyDefaults(p *model.Player) {
	if p == nil {
		return
	}
	if p.Tokens == nil {
		p.Tokens = map[string]int{}
	}
	p.Tokens["ss_blue_soul"] = 0
	p.Tokens["ss_yellow_soul"] = 0
}

// StarterCards 返回开局专属牌列表。
func StarterCards(p *model.Player) []model.Card {
	if p == nil || p.Character == nil {
		return nil
	}
	return []model.Card{
		{
			ID:              fmt.Sprintf("starter-%s-soul_link", p.ID),
			Name:            "灵魂链接",
			Type:            model.CardTypeMagic,
			Element:         model.ElementDark,
			Faction:         p.Character.Faction,
			Description:     "灵魂术士开局自带专属技能卡",
			ExclusiveChar1:  p.Character.ID,
			ExclusiveSkill1: "灵魂链接",
		},
	}
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "ss_soul_devour", Handler: &SoulSorcererSoulDevourHandler{}},
		{ID: "ss_soul_recall", Handler: &SoulSorcererSoulRecallHandler{}},
		{ID: "ss_soul_convert", Handler: &SoulSorcererSoulConvertHandler{}},
		{ID: "ss_soul_mirror", Handler: &SoulSorcererSoulMirrorHandler{}},
		{ID: "ss_soul_blast", Handler: &SoulSorcererSoulBlastHandler{}},
		{ID: "ss_soul_grant", Handler: &SoulSorcererSoulGrantHandler{}},
		{
			ID:      "ss_soul_link",
			Handler: &SoulSorcererSoulLinkHandler{},
			Policy: types.SkillPolicy{
				ManualExclusiveCard: true,
			},
		},
		{ID: "ss_soul_amp", Handler: &SoulSorcererSoulAmpHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"ss_convert_color":    types.ChoiceRouteRole("soul_sorcerer"),
		"ss_link_target":      types.ChoiceRouteRole("soul_sorcerer"),
		"ss_link_transfer_x":  types.ChoiceRouteRole("soul_sorcerer"),
		"ss_recall_pick":      types.ChoiceRouteRole("soul_sorcerer"),
		"ss_soul_devour_pick": types.ChoiceRouteRole("soul_sorcerer"),
		"ss_soul_recall_pick": types.ChoiceRouteRole("soul_sorcerer"),
	}
}
