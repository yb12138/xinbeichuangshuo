// gameflow: 牧师技能可用性检查器。

package priest

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// CheckWaterPowerDiscardUsability 检查水之力量技能的弃牌要求是否可达成。
// 需要：手牌 >= 2，且至少有1张水系牌。
func CheckWaterPowerDiscardUsability(engine player.SkillUsabilityCheckerEngine, p *model.Player, skillDef model.SkillDefinition) bool {
	if len(p.Hand) < 2 {
		return false
	}
	for _, card := range p.Hand {
		if card.Element == model.ElementWater {
			return true
		}
	}
	return false
}
