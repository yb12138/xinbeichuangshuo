// gameflow: 圣枪技能可用性检查器。

package holy_lancer

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// CheckPunishmentUsability 检查惩戒技能是否可用。
// 需要：有至少1个敌方目标且有治疗。
func CheckPunishmentUsability(engine player.SkillUsabilityCheckerEngine, p *model.Player, skillDef model.SkillDefinition) bool {
	for _, pid := range engine.PlayerOrder() {
		target := engine.LookupPlayer(pid)
		if target == nil || target.ID == p.ID || target.Heal <= 0 {
			continue
		}
		return true
	}
	return false
}
