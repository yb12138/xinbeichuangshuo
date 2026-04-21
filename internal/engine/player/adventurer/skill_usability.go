// gameflow: 冒险者技能可用性检查器。

package adventurer

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// CheckFraudUsability 检查欺诈技能是否可用。
// 需要：手牌中有至少2张相同元素的牌，或至少3张任意牌（可包含空元素）。
func CheckFraudUsability(engine player.SkillUsabilityCheckerEngine, p *model.Player, skillDef model.SkillDefinition) bool {
	elementCount := map[model.Element]int{}
	for _, card := range p.Hand {
		elementCount[card.Element]++
	}
	for element, count := range elementCount {
		if element != "" && count >= 2 {
			return true
		}
		if count >= 3 {
			return true
		}
	}
	return false
}
