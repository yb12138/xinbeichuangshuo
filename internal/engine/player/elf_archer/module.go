// gameflow: 精灵射手模块入口声明。

package elf_archer

import (
	"starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
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
			{Timing: player.TimingPostActionEnd, Priority: 100, Hook: postActionEndHook},
			{Timing: player.TimingPostAttackHit, Priority: 400, Hook: postAttackHitHook},
		},
	}
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "elf_elemental_shot", Handler: &skills.ElfElementalShotHandler{}},
		{ID: "elf_animal_companion", Handler: &skills.ElfAnimalCompanionHandler{}},
		{ID: "elf_ritual", Handler: &skills.ElfRitualHandler{}},
		{ID: "elf_pet_empower", Handler: &skills.ElfPetEmpowerHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"elf_archer_elemental_shot_pick":     types.ChoiceRouteRole("elf_archer"),
		"elf_archer_pet_pick":                types.ChoiceRouteRole("elf_archer"),
		"elf_elemental_shot_cost":            types.ChoiceRouteRole("elf_archer"),
		"elf_elemental_shot_discard_magic":   types.ChoiceRouteRole("elf_archer"),
		"elf_elemental_shot_remove_blessing": types.ChoiceRouteRole("elf_archer"),
		"elf_pet_empower_confirm":            types.ChoiceRouteRole("elf_archer"),
		"elf_pet_empower_target":             types.ChoiceRouteTargetPrompt("elf"),
		"elf_ritual_release_target":          types.ChoiceRouteTargetPrompt("elf"),
	}
}
