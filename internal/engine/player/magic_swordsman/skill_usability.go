// gameflow: 魔剑士技能可用性检查器。

package magic_swordsman

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// CheckShadowMeteorUsability 检查暗影流星技能是否可用。
// 需要：处于暗影形态，且至少可弃2张法术牌。
func CheckShadowMeteorUsability(engine player.SkillUsabilityCheckerEngine, p *model.Player, skillDef model.SkillDefinition) bool {
	if !player.HasForm(p, model.FormMagicSwordsmanShadow) {
		return false
	}
	magicCount := 0
	for _, card := range p.Hand {
		if card.Type == model.CardTypeMagic {
			magicCount++
		}
	}
	return magicCount >= 2
}
