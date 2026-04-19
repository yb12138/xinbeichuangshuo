// gameflow: 圣枪骑士：地枪、圣击等互斥与补结算。

package engine

import (
	"starcup-engine/internal/model"
)

func (e *GameEngine) syncHolyLancerRevelationMaxHeal(player *model.Player) {
	if player == nil || !e.isHolyLancer(player) {
		return
	}
	enemyCamp := model.BlueCamp
	if player.Camp == model.BlueCamp {
		enemyCamp = model.RedCamp
	}
	maxHeal := 2
	if e.GetCampCups(string(player.Camp)) >= e.GetCampCups(string(enemyCamp)) {
		maxHeal = 3
	}
	player.MaxHeal = maxHeal
}

func syncHolyLancerDerivedStateOnPlayerSetup(e *GameEngine, player *model.Player) {
	e.syncHolyLancerRevelationMaxHeal(player)
}

func syncHolyLancerDerivedStateOnCampCupChanged(e *GameEngine, _ model.Camp) {
	e.refreshAllPlayerDerivedStates()
}
