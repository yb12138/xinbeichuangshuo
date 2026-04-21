// gameflow: 元素师技能可用性检查器。

package elementalist

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// CheckIgniteUsability 检查点燃技能是否可用。
// 需要：元素指示物 >= 3。
func CheckIgniteUsability(engine player.SkillUsabilityCheckerEngine, p *model.Player, skillDef model.SkillDefinition) bool {
	element := engine.GetToken(p, "element")
	return element >= 3
}
