package server

import "starcup-engine/internal/data"

// Available character roles.
var availableRoles = []string{
	"berserker", "blade_master", "sealer",
	"archer", "assassin", "angel",
	"saintess", "magical_girl",
	"valkyrie", "elementalist", "arbiter",
	"adventurer", "holy_lancer",
	"elf_archer", "plague_mage", "magic_swordsman", "crimson_sword_spirit",
	"prayer_master", "crimson_knight", "war_homunculus", "priest", "onmyoji",
	"blaze_witch",
	"sage", "magic_bow", "magic_lancer", "spirit_caster", "bard", "hero", "fighter", "holy_bow", "sword_emperor", "beast_samurai", "soul_sorcerer", "moon_goddess", "blood_priestess", "butterfly_dancer",
}

func isValidRole(role string) bool {
	for _, availableRole := range availableRoles {
		if availableRole == role {
			return true
		}
	}
	return false
}

// buildCharacterViews 从 data 包构建角色视图。
func buildCharacterViews() []CharacterView {
	chars := data.GetCharacters()
	views := make([]CharacterView, 0, len(chars))
	for _, character := range chars {
		skills := make([]SkillView, 0, len(character.Skills))
		for _, skill := range character.Skills {
			skills = append(skills, SkillView{
				ID:               skill.ID,
				Title:            skill.Title,
				Description:      skill.Description,
				Type:             int(skill.Type),
				MinTargets:       skill.MinTargets,
				MaxTargets:       skill.MaxTargets,
				TargetType:       int(skill.TargetType),
				CostGem:          skill.CostGem,
				CostCrystal:      skill.CostCrystal,
				CostDiscards:     skill.CostDiscards,
				DiscardElement:   string(skill.DiscardElement),
				RequireExclusive: skill.RequireExclusive,
			})
		}
		views = append(views, CharacterView{
			ID:      character.ID,
			Name:    character.Name,
			Title:   character.Title,
			Faction: character.Faction,
			Skills:  skills,
		})
	}
	return views
}
