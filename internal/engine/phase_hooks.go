// gameflow: 阶段边界上的通用钩子入口。

package engine

import (
	engineplayer "starcup-engine/internal/engine/player"
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

type TurnEndTimingStage int

const (
	TimingTurnEndPreExtra TurnEndTimingStage = iota
	TimingTurnEndFinal
)

// runTimingTurnStartStageHooks 统一处理回合开始时（TimingTurnStart）阶段规则。
func (e *GameEngine) runTimingTurnStartStageHooks(player *model.Player, stage timingOnTurnStartStage) bool {
	switch stage {
	case timingOnTurnStartMain:
		if player != nil {
			result := e.dispatchAllRoleTimingHooks(engineplayer.TimingTurnStart, engineplayer.TimingHookContext{
				SourceID:     player.ID,
				TurnPlayerID: player.ID,
			})
			if result.Interrupted {
				return true
			}
		}
		return false
	default:
		panic("unregistered TimingTurnStart stage")
	}
}

// RunTurnEndTimingStageHooks 统一处理回合结束时（TimingTurnEnd）的运行时子阶段规则。
func (e *GameEngine) RunTurnEndTimingStageHooks(player *model.Player, stage TurnEndTimingStage) bool {
	switch stage {
	case TimingTurnEndPreExtra:
		// TimingHookSpec dispatch for turn end (pre-extra)
		if player != nil {
			result := e.dispatchAllRoleTimingHooks(engineplayer.TimingTurnEndPreExtra, engineplayer.TimingHookContext{
				SourceID:     player.ID,
				TurnPlayerID: player.ID,
			})
			if result.Interrupted {
				return true
			}
		}
		return false
	case TimingTurnEndFinal:
		// TimingHookSpec dispatch for turn end final
		if player != nil {
			result := e.dispatchAllRoleTimingHooks(engineplayer.TimingTurnEndFinal, engineplayer.TimingHookContext{
				SourceID:     player.ID,
				TurnPlayerID: player.ID,
			})
			if result.Interrupted {
				return true
			}
		}
		return false
	default:
		panic("unregistered TimingTurnEndPreExtra stage")
	}
}

// runTimingTurnStartBeforeStartHooks 回合开始前（TimingTurnBeforeStart）固定结算点。
func (e *GameEngine) runTimingTurnStartBeforeStartHooks(player *model.Player) bool {
	// 回合开始前已迁移到 TimingHookSpec。
	if player != nil {
		result := e.dispatchRoleTimingHook(engineplayer.TimingTurnBeforeStart, engineplayer.TimingHookContext{
			TargetID:     player.ID,
			TurnPlayerID: player.ID,
		})
		return result.Interrupted
	}
	return false
}

// runTimingTurnStartHooks 回合开始时（TimingTurnStart）固定结算点。
func (e *GameEngine) runTimingTurnStartHooks(player *model.Player) bool {
	return e.runTimingTurnStartStageHooks(player, timingOnTurnStartMain)
}

// runTimingActionStartExecuteHooks 行动阶段开始时（TimingActionStart）固定结算点。
func (e *GameEngine) runTimingActionStartExecuteHooks(player *model.Player) bool {
	// TimingHookSpec dispatch for action start
	if player != nil {
		result := e.dispatchAllRoleTimingHooks(engineplayer.TimingActionStart, engineplayer.TimingHookContext{
			SourceID:     player.ID,
			TurnPlayerID: player.ID,
		})
		if result.Interrupted {
			return true
		}
	}
	return false
}

// runTimingTurnEndPreExtraHooks 回合结束时前置结算（额外行动判定前）。
func (e *GameEngine) runTimingTurnEndPreExtraHooks(player *model.Player) bool {
	return e.RunTurnEndTimingStageHooks(player, TimingTurnEndPreExtra)
}

// runTimingTurnEndFinalHooks 回合结束时最终结算点。
func (e *GameEngine) runTimingTurnEndFinalHooks(player *model.Player) bool {
	return e.RunTurnEndTimingStageHooks(player, TimingTurnEndFinal)
}

// blaze_witch/assassin hooks 已迁移到 TimingHookSpec
// arbiter/holy_bow turnStart hooks 已迁移到 TimingHookSpec
// crimson_knight turnEnd hooks 已迁移到 TimingHookSpec
// magic_swordsman/bard/moon/onmyoji hooks 已迁移到 TimingHookSpec
