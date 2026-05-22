// gameflow: 精灵射手模块入口声明。

package elf_archer

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:               "elf_archer",
		Choices:          NewChoiceHandler(),
		Skills:           SkillEntries(),
		ChoiceRouteSpecs: ChoiceRouteSpecs(),
		TimingHookSpecs: []player.TimingHookSpec{
			{Timing: player.TimingDamageSourceDeal, Priority: 100, Hook: damageCalculateHook},
			{Timing: player.TimingPostActionEnd, Priority: 100, Hook: postActionEndHook},
			{Timing: player.TimingPostAttackHit, Priority: 400, Hook: postAttackHitHook},
			{Timing: player.TimingPostDamageResolved, Priority: 600, Hook: postDamageResolvedHook},
			{Timing: player.TimingOnTurnEnd, Priority: 300, Hook: turnEndHook},
		},
		PlayableCoverEffects:   []model.EffectType{model.EffectElfBlessing},
		ExcludeCardFromDiscard: IsBlessingCard,
	}
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "elf_elemental_shot", Handler: &ElfElementalShotHandler{}},
		{
			ID:      "elf_animal_companion",
			Handler: &ElfAnimalCompanionHandler{},
			Policy: types.SkillPolicy{
				ExclusiveResponseGroup: "elf_archer_pet_response",
			},
		},
		{
			ID:      "elf_ritual",
			Handler: &ElfRitualHandler{},
			Policy: types.SkillPolicy{
				AfterExecute: func(host types.PolicyHost, ctx types.PolicyContext) error {
					host.DropQueuedOverflowDiscardForPlayer(ctx.PlayerID)
					return nil
				},
			},
		},
		{
			ID:      "elf_pet_empower",
			Handler: &ElfPetEmpowerHandler{},
			Policy: types.SkillPolicy{
				ExclusiveResponseGroup: "elf_archer_pet_response",
			},
		},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"elf_archer_elemental_shot_pick":     types.ChoiceRouteRole("elf_archer"),
		"elf_archer_pet_pick":                types.ChoiceRouteRole("elf_archer"),
		"elf_pet_empower_confirm":            types.ChoiceRouteRole("elf_archer"),
		"elf_pet_empower_target":             types.ChoiceRouteRole("elf_archer"),
		"elf_ritual_release_target":          types.ChoiceRouteRole("elf_archer"),
	}
}
