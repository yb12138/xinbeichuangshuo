package engine

import "starcup-engine/internal/model"

type gameStartPlayerHook func(e *GameEngine, player *model.Player)
type gameStartFinalizeHook func(e *GameEngine) bool

type timingOnGameStartStage int

const (
	timingOnGameStartAddPlayer timingOnGameStartStage = iota
	timingOnGameStartInitialDeal
	timingOnGameStartFinalizeIdle
)

// runTimingOnGameStartHooks 统一处理 TimingOnGameStart 阶段规则。
func (e *GameEngine) runTimingOnGameStartHooks(player *model.Player, stage timingOnGameStartStage) bool {
	switch stage {
	case timingOnGameStartAddPlayer:
		for _, hook := range e.gameStartAddPlayerHooks {
			hook(e, player)
		}
	case timingOnGameStartInitialDeal:
		for _, hook := range e.gameStartInitialDealHooks {
			hook(e, player)
		}
	case timingOnGameStartFinalizeIdle:
		for _, hook := range e.gameStartFinalizeHooks {
			if hook(e) {
				return true
			}
		}
	default:
		panic("unregistered TimingOnGameStart stage")
	}
	return false
}

// runPlayerAddBootstrapTiming 在玩家入场时执行初始化规则。
func (e *GameEngine) runPlayerAddBootstrapTiming(player *model.Player) {
	e.runTimingOnGameStartHooks(player, timingOnGameStartAddPlayer)
}

// runPlayerGameStartBootstrapTiming 在游戏开局发牌后执行开局规则。
func (e *GameEngine) runPlayerGameStartBootstrapTiming(player *model.Player) {
	e.runTimingOnGameStartHooks(player, timingOnGameStartInitialDeal)
}

func bootstrapApplyRoleDefaults(e *GameEngine, player *model.Player) {
	if e == nil {
		return
	}
	e.applyRoleDefaults(player)
}

func bootstrapEnsureStarterRoleCards(e *GameEngine, player *model.Player) {
	if e == nil {
		return
	}
	e.ensureStarterRoleCards(player)
}
