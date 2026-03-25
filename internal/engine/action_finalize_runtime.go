package engine

type actionFinalizeHook func(e *GameEngine) bool

var actionFinalizeIdleHooks = []actionFinalizeHook{
	actionFinalizeBloodPriestessBleedHook,
}

func (e *GameEngine) runActionFinalizeHooksIfIdle() bool {
	if e == nil || e.actionSummary == nil || !e.actionSummary.active {
		return false
	}
	if !e.isActionFinalizeIdle() {
		return false
	}
	triggered := false
	for _, hook := range actionFinalizeIdleHooks {
		if hook != nil && hook(e) {
			triggered = true
		}
	}
	return triggered
}

func actionFinalizeBloodPriestessBleedHook(e *GameEngine) bool {
	if e == nil {
		return false
	}
	return e.resolveBloodPriestessBleedExitOnActionEnd()
}
