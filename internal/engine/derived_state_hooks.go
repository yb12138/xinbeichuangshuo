package engine

import "starcup-engine/internal/model"

type campChangedPlayerHook func(e *GameEngine, player *model.Player)
type campChangedCupHook func(e *GameEngine, changedCamp model.Camp)

type timingOnCampChangedStage int

const (
	timingOnCampChangedPlayerSetup timingOnCampChangedStage = iota
	timingOnCampChangedCampCup
)

// runTimingOnCampChangedHooks 统一处理 TimingOnCampChanged 阶段规则。
func (e *GameEngine) runTimingOnCampChangedHooks(player *model.Player, changedCamp model.Camp, stage timingOnCampChangedStage) {
	switch stage {
	case timingOnCampChangedPlayerSetup:
		for _, hook := range e.campChangedPlayerSetupHooks {
			hook(e, player)
		}
	case timingOnCampChangedCampCup:
		for _, hook := range e.campChangedCampCupHooks {
			hook(e, changedCamp)
		}
	default:
		panic("unregistered TimingOnCampChanged stage")
	}
}

func (e *GameEngine) refreshPlayerDerivedState(player *model.Player) {
	e.refreshTimingDerivedStateOnPlayerSetup(player)
}

// refreshTimingDerivedStateOnPlayerSetup 在玩家初始化/刷新时同步派生状态。
func (e *GameEngine) refreshTimingDerivedStateOnPlayerSetup(player *model.Player) {
	e.runTimingOnCampChangedHooks(player, "", timingOnCampChangedPlayerSetup)
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

// refreshTimingDerivedStateOnCampCupChanged 在星杯变化时同步相关派生状态。
func (e *GameEngine) refreshTimingDerivedStateOnCampCupChanged(changedCamp model.Camp) {
	e.runTimingOnCampChangedHooks(nil, changedCamp, timingOnCampChangedCampCup)
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
		e.refreshTimingDerivedStateOnCampCupChanged(camp)
	}
	return changed
}
