// gameflow: 圣弓：信仰/圣炮资源与碎片偏射辅助。
package engine

import (
	"starcup-engine/internal/model"
)

func holyBowFaith(player *model.Player) int {
	return tokenValueBounded(player, "hb_faith", holyBowFaithCapEngine)
}

func addHolyBowFaith(player *model.Player, delta int) int {
	return addTokenValueBounded(player, "hb_faith", delta, holyBowFaithCapEngine)
}

func holyBowCannon(player *model.Player) int {
	return tokenValueBounded(player, "hb_cannon", holyBowCannonCapEngine)
}

func (e *GameEngine) holyBowShardMissEligibleAllies(user *model.Player, x int) []string {
	if e == nil || user == nil || x <= 0 {
		return nil
	}
	allyIDs := make([]string, 0)
	for _, pid := range e.State.PlayerOrder {
		p := e.State.Players[pid]
		if p == nil || p.Camp != user.Camp || p.ID == user.ID {
			continue
		}
		if len(p.Hand) < x {
			continue
		}
		allyIDs = append(allyIDs, p.ID)
	}
	return allyIDs
}

func (e *GameEngine) holyBowShardMissValidXValues(user *model.Player, maxX int) []int {
	if e == nil || user == nil || maxX <= 0 {
		return nil
	}
	valid := make([]int, 0, maxX)
	for x := 1; x <= maxX; x++ {
		if len(e.holyBowShardMissEligibleAllies(user, x)) > 0 {
			valid = append(valid, x)
		}
	}
	return valid
}
