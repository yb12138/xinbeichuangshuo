// gameflow: 阶段边界上的通用钩子入口。

package engine

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	bardpkg "starcup-engine/internal/engine/player/bard"
	magicswordsman "starcup-engine/internal/engine/player/magic_swordsman"
	moonpkg "starcup-engine/internal/engine/player/moon"
	"starcup-engine/internal/model"
)

type turnTimingHook func(e *GameEngine, player *model.Player) bool

func ensurePlayerTokensMap(player *model.Player) {
	if player != nil && player.Tokens == nil {
		player.Tokens = map[string]int{}
	}
}

type timingOnTurnStartStage int

const (
	timingOnTurnStartMain timingOnTurnStartStage = iota
)

type timingOnTurnEndStage int

const (
	timingOnTurnEndPreExtra timingOnTurnEndStage = iota
	timingOnTurnEndFinal
)

// runTimingOnTurnStartStageHooks 统一处理 TimingOnTurnStart 阶段规则。
func (e *GameEngine) runTimingOnTurnStartStageHooks(player *model.Player, stage timingOnTurnStartStage) bool {
	switch stage {
	case timingOnTurnStartMain:
		for _, hook := range e.turnStartMainHooks {
			if hook(e, player) {
				return true
			}
		}
		if player != nil {
			result := e.dispatchAllRoleTimingHooks(engineplayer.TimingOnTurnStart, engineplayer.TimingHookContext{
				SourceID:     player.ID,
				TurnPlayerID: player.ID,
			})
			if result.Interrupted {
				return true
			}
		}
		return false
	default:
		panic(fmt.Sprintf("unregistered TimingOnTurnStart stage: %d", stage))
	}
}

// runTimingOnTurnEndStageHooks 统一处理 TimingOnTurnEnd 阶段规则。
func (e *GameEngine) runTimingOnTurnEndStageHooks(player *model.Player, stage timingOnTurnEndStage) bool {
	switch stage {
	case timingOnTurnEndPreExtra:
		for _, hook := range e.turnEndPreExtraHooks {
			if hook(e, player) {
				return true
			}
		}
		// TimingHookSpec dispatch for turn end (pre-extra)
		if player != nil {
			result := e.dispatchAllRoleTimingHooks(engineplayer.TimingOnTurnEnd, engineplayer.TimingHookContext{
				SourceID:     player.ID,
				TurnPlayerID: player.ID,
			})
			if result.Interrupted {
				return true
			}
		}
		return false
	case timingOnTurnEndFinal:
		for _, hook := range e.turnEndFinalHooks {
			if hook(e, player) {
				return true
			}
		}
		// TimingHookSpec dispatch for turn end final
		if player != nil {
			result := e.dispatchAllRoleTimingHooks(engineplayer.TimingOnTurnEndFinal, engineplayer.TimingHookContext{
				SourceID:     player.ID,
				TurnPlayerID: player.ID,
			})
			if result.Interrupted {
				return true
			}
		}
		return false
	default:
		panic(fmt.Sprintf("unregistered TimingOnTurnEnd stage: %d", stage))
	}
}

// runTimingOnTurnStartBeforeStartHooks 回合开始前（TurnBeforeStart）固定结算点。
func (e *GameEngine) runTimingOnTurnStartBeforeStartHooks(player *model.Player) bool {
	// TimingOnTurnBeforeStart 已迁移到 TimingHookSpec。
	if player != nil {
		result := e.dispatchRoleTimingHook(engineplayer.TimingOnTurnBeforeStart, engineplayer.TimingHookContext{
			TargetID:     player.ID,
			TurnPlayerID: player.ID,
		})
		return result.Interrupted
	}
	return false
}

// runTimingOnTurnStartHooks 回合开始（TurnStart）固定结算点。
func (e *GameEngine) runTimingOnTurnStartHooks(player *model.Player) bool {
	return e.runTimingOnTurnStartStageHooks(player, timingOnTurnStartMain)
}

// runTimingBeforeActionExecuteHooks 行动开始（ActionStart）固定结算点。
func (e *GameEngine) runTimingBeforeActionExecuteHooks(player *model.Player) bool {
	for _, hook := range e.beforeActionExecuteHooks {
		if hook(e, player) {
			return true
		}
	}
	// TimingHookSpec dispatch for before action
	if player != nil {
		result := e.dispatchAllRoleTimingHooks(engineplayer.TimingBeforeAction, engineplayer.TimingHookContext{
			SourceID:     player.ID,
			TurnPlayerID: player.ID,
		})
		if result.Interrupted {
			return true
		}
	}
	return false
}

// runTimingOnTurnEndPreExtraHooks 回合结束前置结算（额外行动判定前）。
func (e *GameEngine) runTimingOnTurnEndPreExtraHooks(player *model.Player) bool {
	return e.runTimingOnTurnEndStageHooks(player, timingOnTurnEndPreExtra)
}

// runTimingOnTurnEndFinalHooks 回合结束最终结算点。
func (e *GameEngine) runTimingOnTurnEndFinalHooks(player *model.Player) bool {
	return e.runTimingOnTurnEndStageHooks(player, timingOnTurnEndFinal)
}

// blaze_witch/assassin hooks 已迁移到 TimingHookSpec
// arbiter/holy_bow turnStart hooks 已迁移到 TimingHookSpec
// crimson_knight turnEnd hooks 已迁移到 TimingHookSpec

func startupMagicSwordsmanShadowReleaseHook(e *GameEngine, player *model.Player) bool {
	magicswordsman.MaybeReleaseShadowAtActionStart(e, player)
	return false
}

func startupBardRousingHook(e *GameEngine, player *model.Player) bool {
	return bardpkg.MaybeRousingAtTurnStart(newRoleChoiceRuntime(e), player)
}

func turnEndMoonGoddessHook(e *GameEngine, player *model.Player) bool {
	return moonpkg.MaybeMoonCycleAtTurnEnd(newRoleChoiceRuntime(e), player)
}

func turnEndBardHook(e *GameEngine, player *model.Player) bool {
	return bardpkg.MaybeVictoryAtTurnEnd(newRoleChoiceRuntime(e), player)
}

func turnEndOnmyojiHook(e *GameEngine, player *model.Player) bool {
	return e.maybeOnmyojiDarkRitual(player)
}
