package engine

import "starcup-engine/internal/model"

type deferredFollowupResolver func(*GameEngine, model.DeferredFollowup) error

type deferredFollowupHandler struct {
	label   string
	resolve deferredFollowupResolver
}

var deferredFollowupHandlers = buildDeferredFollowupHandlers()

func buildDeferredFollowupHandlers() map[string]deferredFollowupHandler {
	handlers := map[string]deferredFollowupHandler{}
	registerDeferredFollowupHandlers(handlers, buildBloodPriestessDeferredFollowupHandlers())
	registerDeferredFollowupHandlers(handlers, buildSpiritCasterDeferredFollowupHandlers())
	registerDeferredFollowupHandlers(handlers, buildAssassinDeferredFollowupHandlers())
	return handlers
}

func registerDeferredFollowupHandlers(dst map[string]deferredFollowupHandler, src map[string]deferredFollowupHandler) {
	for followupType, handler := range src {
		if followupType == "" || handler.resolve == nil {
			continue
		}
		dst[followupType] = handler
	}
}

func (e *GameEngine) resolveDeferredFollowup(f model.DeferredFollowup) (bool, string, error) {
	handler, ok := deferredFollowupHandlers[f.Type]
	if !ok || handler.resolve == nil {
		return false, "", nil
	}
	return true, handler.label, handler.resolve(e, f)
}
