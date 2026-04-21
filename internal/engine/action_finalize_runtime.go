// gameflow: 单次行动结束后的收尾钩子与状态收敛。

package engine

func (e *GameEngine) runActionFinalizeHooksIfIdle() bool {
	if e == nil || e.actionSummary == nil || !e.actionSummary.active {
		return false
	}
	if !e.isActionFinalizeIdle() {
		return false
	}
	return e.runTimingOnActionEndFinalizeEffects()
}

// runTimingOnActionEndFinalizeEffects 在行动彻底收尾后按固定顺序执行收尾规则。
func (e *GameEngine) runTimingOnActionEndFinalizeEffects() bool {
	return e.runTimingOnGameStartHooks(nil, timingOnGameStartFinalizeIdle)
}
