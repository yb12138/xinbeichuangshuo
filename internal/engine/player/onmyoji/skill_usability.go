// gameflow: 阴阳师技能可用性检查器。

package onmyoji

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// CheckShikigamiDescendUsability 检查式神降临技能是否可用。
// 需要：手牌中有至少2张相同阵营的牌。
func CheckShikigamiDescendUsability(engine player.SkillUsabilityCheckerEngine, p *model.Player, skillDef model.SkillDefinition) bool {
	factionCount := map[string]int{}
	for _, card := range p.Hand {
		if card.Faction == "" {
			continue
		}
		factionCount[card.Faction]++
		if factionCount[card.Faction] >= 2 {
			return true
		}
	}
	return false
}
