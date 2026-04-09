package data

import "starcup-engine/internal/model"

func normalizeSkillTimings(characters []model.Character) []model.Character {
	for ci := range characters {
		for si := range characters[ci].Skills {
			skill := &characters[ci].Skills[si]
			if len(skill.Timings) > 0 {
				continue
			}

			timings := make([]model.TriggerTiming, 0, 1+len(skill.ExtraTriggers))
			appendUnique := func(t model.TriggerTiming) {
				if t == model.TimingUnknown {
					return
				}
				for _, existing := range timings {
					if existing == t {
						return
					}
				}
				timings = append(timings, t)
			}

			appendUnique(model.LegacyTriggerToTiming(skill.Trigger))
			for _, legacy := range skill.ExtraTriggers {
				appendUnique(model.LegacyTriggerToTiming(legacy))
			}
			skill.Timings = timings
		}
	}
	return characters
}
