// gameflow: 吟游诗人技能可用性检查器。

package bard

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// CheckDissonanceChordUsability 检查不协和音技能是否可用。
// 需要：灵感指示物 > 1。
func CheckDissonanceChordUsability(engine player.SkillUsabilityCheckerEngine, p *model.Player, skillDef model.SkillDefinition) bool {
	inspiration := engine.GetToken(p, "bd_inspiration")
	return inspiration > 1
}
