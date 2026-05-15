// gameflow: 技能能量可支付性（与 engine.interface 规则一致，避免 runtime 依赖 engine）。

package skill

import (
	"starcup-engine/internal/engine/resource"
	"starcup-engine/internal/model"
)

// CanPaySkillEnergyCost 规则：
// 1) 宝石消耗必须由宝石支付（不可用水晶替代）；
// 2) 水晶消耗可由「剩余宝石」替代。
func CanPaySkillEnergyCost(p *model.Player, gemCost, crystalCost int) bool {
	return resource.CanPaySkillEnergyCost(p, gemCost, crystalCost)
}
