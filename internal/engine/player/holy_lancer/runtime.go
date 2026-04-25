package holy_lancer

import (
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// SyncRevelationMaxHeal synchronizes the holy lancer revelation max-heal derived state.
func SyncRevelationMaxHeal(rt engineplayer.ChoiceRuntime, player *model.Player) {
	if player == nil || !engineplayer.IsCharacter(player, "holy_lancer") {
		return
	}
	enemyCamp := model.BlueCamp
	if player.Camp == model.BlueCamp {
		enemyCamp = model.RedCamp
	}
	maxHeal := 2
	if rt.GetCampCups(string(player.Camp)) >= rt.GetCampCups(string(enemyCamp)) {
		maxHeal = 3
	}
	player.MaxHeal = maxHeal
}

// SyncDerivedStateOnCampCupChanged refreshes all player derived states when a camp cup changes.
func SyncDerivedStateOnCampCupChanged(rt engineplayer.ChoiceRuntime) {
	rt.RefreshAllPlayerDerivedStates()
}
