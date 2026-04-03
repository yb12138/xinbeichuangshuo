package engine

import "starcup-engine/internal/model"

type playerBootstrapHook func(e *GameEngine, player *model.Player)

var addPlayerBootstrapHooks = []playerBootstrapHook{
	bootstrapApplyRoleDefaults,
}

var gameStartBootstrapHooks = []playerBootstrapHook{
	bootstrapEnsureStarterRoleCards,
}

func (e *GameEngine) runPlayerBootstrapHooks(player *model.Player, hooks []playerBootstrapHook) {
	for _, hook := range hooks {
		if hook != nil {
			hook(e, player)
		}
	}
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
