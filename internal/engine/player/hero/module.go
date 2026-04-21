// gameflow: 勇者模块入口声明。

package hero

import (
	"fmt"

	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:               "hero",
		Defaults:         ApplyDefaults,
		StarterCards:     StarterCards,
		Choices:          NewChoiceHandler(),
		Skills:           SkillEntries(),
		ChoiceRouteSpecs: ChoiceRouteSpecs(),
		TimingHookSpecs: []player.TimingHookSpec{
			{Timing: player.TimingOnDamageCalculate, Priority: 500, Hook: damageCalculateHook},
			{Timing: player.TimingOnAttackDeclared, Priority: 100, Hook: pendingDamageInitHook},
			{Timing: player.TimingOnAttackGating, Priority: 200, Hook: attackGatingHook},
			{Timing: player.TimingOnAttackMiss, Priority: 100, Hook: attackMissHook},
			{Timing: player.TimingPostActionEnd, Priority: 200, Hook: postActionEndHook},
		},
	}
}

// ApplyDefaults 初始化角色默认属性。
func ApplyDefaults(p *model.Player) {
	if p == nil {
		return
	}
	p.Crystal += 2
	if p.Tokens == nil {
		p.Tokens = map[string]int{}
	}
	p.Tokens["hero_anger"] = 0
	p.Tokens["hero_wisdom"] = 0
}

// StarterCards 返回开局专属牌列表。
func StarterCards(p *model.Player) []model.Card {
	if p == nil || p.Character == nil {
		return nil
	}
	return []model.Card{
		{
			ID:              fmt.Sprintf("starter-%s-hero_taunt", p.ID),
			Name:            "挑衅",
			Type:            model.CardTypeMagic,
			Element:         model.ElementFire,
			Faction:         p.Character.Faction,
			Description:     "勇者开局自带专属技能卡",
			ExclusiveChar1:  p.Character.ID,
			ExclusiveSkill1: "挑衅",
		},
	}
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "hero_heart", Handler: &HeroHeartHandler{}},
		{ID: "hero_roar", Handler: &HeroRoarHandler{}},
		{ID: "hero_forbidden_power", Handler: &HeroForbiddenPowerHandler{}},
		{ID: "hero_exhaustion", Handler: &HeroExhaustionHandler{}},
		{ID: "hero_calm_mind", Handler: &HeroCalmMindHandler{}},
		{ID: "hero_taunt", Handler: &HeroTauntHandler{}},
		{ID: "hero_dead_duel", Handler: &HeroDeadDuelHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"hero_roar_draw": types.ChoiceRouteRole("hero"),
	}
}
