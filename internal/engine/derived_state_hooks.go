package engine

import "starcup-engine/internal/model"

type playerDerivedStateHook func(e *GameEngine, player *model.Player)
type campCupChangedHook func(e *GameEngine, changedCamp model.Camp)

var playerDerivedStateHooks = []playerDerivedStateHook{
	syncHolyLancerDerivedStateOnPlayerSetup,
}

var campCupChangedHooks = []campCupChangedHook{
	syncHolyLancerDerivedStateOnCampCupChanged,
}

func (e *GameEngine) refreshPlayerDerivedState(player *model.Player) {
	for _, hook := range playerDerivedStateHooks {
		hook(e, player)
	}
}

func (e *GameEngine) refreshAllPlayerDerivedStates() {
	if e == nil || e.State == nil {
		return
	}
	seen := map[string]struct{}{}
	for _, pid := range e.State.PlayerOrder {
		player := e.State.Players[pid]
		if player == nil {
			continue
		}
		seen[pid] = struct{}{}
		e.refreshPlayerDerivedState(player)
	}
	for pid, player := range e.State.Players {
		if player == nil {
			continue
		}
		if _, ok := seen[pid]; ok {
			continue
		}
		e.refreshPlayerDerivedState(player)
	}
}

func (e *GameEngine) runCampCupChangedHooks(changedCamp model.Camp) {
	for _, hook := range campCupChangedHooks {
		hook(e, changedCamp)
	}
}

func (e *GameEngine) addCampCup(camp model.Camp) bool {
	if e == nil || e.State == nil {
		return false
	}
	changed := false
	switch camp {
	case model.RedCamp:
		if e.State.RedCups < 5 {
			e.State.RedCups++
			changed = true
		}
	case model.BlueCamp:
		if e.State.BlueCups < 5 {
			e.State.BlueCups++
			changed = true
		}
	}
	if changed {
		e.runCampCupChangedHooks(camp)
	}
	return changed
}
