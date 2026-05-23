// gameflow: 统一技能能量支付规则。

package resource

import "starcup-engine/internal/model"

// UsableCrystal returns the amount available for a blue-crystal cost.
// Red gems may substitute only in the crystal-cost direction.
func UsableCrystal(p *model.Player) int {
	if p == nil {
		return 0
	}
	return p.Crystal + p.Gem
}

// CanPayCrystalCost checks a pure blue-crystal cost with red-gem substitution.
func CanPayCrystalCost(p *model.Player, amount int) bool {
	if amount <= 0 {
		return true
	}
	return UsableCrystal(p) >= amount
}

// SpendCrystalCost pays a pure blue-crystal cost, spending crystals first and
// then red gems for any remainder.
func SpendCrystalCost(p *model.Player, amount int) bool {
	if amount <= 0 {
		return true
	}
	if !CanPayCrystalCost(p, amount) {
		return false
	}
	useCrystal := amount
	if useCrystal > p.Crystal {
		useCrystal = p.Crystal
	}
	p.Crystal -= useCrystal
	p.Gem -= amount - useCrystal
	return true
}

// CanPaySkillEnergyCost checks a combined skill cost:
// 1) gem cost must be paid by gems;
// 2) crystal cost may be paid by crystals plus gems left after the gem cost.
func CanPaySkillEnergyCost(p *model.Player, gemCost, crystalCost int) bool {
	if p == nil {
		return false
	}
	gemCost = normalizeCost(gemCost)
	crystalCost = normalizeCost(crystalCost)
	if p.Gem < gemCost {
		return false
	}
	return p.Crystal+(p.Gem-gemCost) >= crystalCost
}

// SpendSkillEnergyCost pays a combined skill cost. Gems required as gems are
// spent first; the remaining crystal cost is paid by crystals first, then gems.
func SpendSkillEnergyCost(p *model.Player, gemCost, crystalCost int) bool {
	if !CanPaySkillEnergyCost(p, gemCost, crystalCost) {
		return false
	}
	gemCost = normalizeCost(gemCost)
	crystalCost = normalizeCost(crystalCost)

	p.Gem -= gemCost
	return SpendCrystalCost(p, crystalCost)
}

func normalizeCost(cost int) int {
	if cost < 0 {
		return 0
	}
	return cost
}
