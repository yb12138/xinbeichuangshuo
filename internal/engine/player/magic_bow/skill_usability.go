// gameflow: 魔弓技能可用性检查器。

package magic_bow

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// CheckThunderScatterUsability 检查雷霆散射技能是否可用。
// 需要：本回合未锁定充能，且至少有1张雷系充能牌。
func CheckThunderScatterUsability(engine player.SkillUsabilityCheckerEngine, p *model.Player, skillDef model.SkillDefinition) bool {
	if p.TurnState.UsedSkillCounts["mb_charge_lock_turn"] > 0 {
		return false
	}
	return engine.CountCoverCardsByEffectAndElement(p, model.EffectMagicBowCharge, model.ElementThunder) > 0
}
