package sword_emperor

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

const SwordQiCap = 5

func SwordQi(p *model.Player) int {
	return player.TokenValue(p, "se_sword_qi", SwordQiCap)
}

func AddSwordQi(p *model.Player, delta int) int {
	return player.AddToken(p, "se_sword_qi", delta, SwordQiCap)
}

// ClearCombatTokens zeros turn-state combat markers for guard-disabled, angel-soul, and demon-soul.
func ClearCombatTokens(p *model.Player) {
	if p == nil {
		return
	}
	ts := &p.TurnState
	ts.UsedSkillCounts["se_guard_disabled_current_attack"] = 0
	ts.UsedSkillCounts["se_angel_soul_armed"] = 0
	ts.UsedSkillCounts["se_demon_soul_armed"] = 0
}
