package engine

import "starcup-engine/internal/model"

// GetMaxHand 计算玩家的动态手牌上限
func (e *GameEngine) GetMaxHand(player *model.Player) int {
	if player == nil {
		return 0
	}
	if fixed, ok := e.fixedMaxHandCapValue(player); ok {
		return fixed
	}

	maxHand := e.applyRoleMaxHandModifiers(player, player.MaxHand)
	if maxHand < 0 {
		return 0
	}
	return maxHand
}

func (e *GameEngine) fixedMaxHandCapValue(player *model.Player) (int, bool) {
	if !e.hasFixedMaxHandCap(player) {
		return 0, false
	}
	if e.isPrayerMaster(player) && hasPrayerMasterPrayerForm(player) {
		return 5, true
	}
	if e.isMagicLancer(player) && hasMagicLancerPhantomForm(player) {
		return 5, true
	}
	if e.isHero(player) && hasHeroExhaustionForm(player) {
		return 4, true
	}
	return 7, true // 怜悯
}

func (e *GameEngine) applyRoleMaxHandModifiers(player *model.Player, maxHand int) int {
	if e.isWarHomunculus(player) && hasWarHomunculusBurstForm(player) {
		maxHand++
	}
	if e.isBlazeWitch(player) && hasBlazeWitchFlameForm(player) {
		maxHand += player.Tokens["bw_rebirth"] - 2
	}
	if isCharacter(player, "assassin") && hasAssassinStealthForm(player) {
		maxHand--
	}
	// 蝶舞者：生命之火，手牌上限 = 基础上限 - 蛹数，最低为3。
	if e.isButterflyDancer(player) {
		maxHand -= butterflyPupa(player)
		if maxHand < 3 {
			maxHand = 3
		}
	}
	// 血之巫女：同生共死根据形态动态修正手牌上限。
	maxHand += e.bloodPriestessSharedLifeDeltaFor(player)
	return maxHand
}
