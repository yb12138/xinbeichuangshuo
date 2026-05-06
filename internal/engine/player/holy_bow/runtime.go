package holy_bow

import (
	"starcup-engine/internal/model"
)

// ShardMissEligibleAllies returns ally IDs whose hand size is at least x.
func ShardMissEligibleAllies(allPlayers map[string]*model.Player, playerOrder []string, user *model.Player, x int) []string {
	if user == nil || x <= 0 {
		return nil
	}
	allyIDs := make([]string, 0)
	for _, pid := range playerOrder {
		p := allPlayers[pid]
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

// ShardMissValidXValues returns all x values in [1..maxX] that have at least one eligible ally.
func ShardMissValidXValues(allPlayers map[string]*model.Player, playerOrder []string, user *model.Player, maxX int) []int {
	if user == nil || maxX <= 0 {
		return nil
	}
	valid := make([]int, 0, maxX)
	for x := 1; x <= maxX; x++ {
		if len(ShardMissEligibleAllies(allPlayers, playerOrder, user, x)) > 0 {
			valid = append(valid, x)
		}
	}
	return valid
}
