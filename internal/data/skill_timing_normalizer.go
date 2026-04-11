package data

import "starcup-engine/internal/model"

func normalizeSkillTimings(characters []model.Character) []model.Character {
	for ci := range characters {
		for si := range characters[ci].Skills {
			skill := &characters[ci].Skills[si]
			if len(skill.Timings) == 0 {
				skill.Timings = []model.FlowTiming{model.TimingActive}
			}
		}
	}
	return characters
}
