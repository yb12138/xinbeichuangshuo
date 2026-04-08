package engine

import "fmt"

// Drive 状态机驱动函数，自动在阶段间转换或等待用户输入
func (e *GameEngine) Drive() {
	iterations := 0
	resolveDriveOutcome := func(outcome driveOutcome) (handled bool, shouldStop bool) {
		if outcome == driveUnhandled {
			return false, false
		}
		return true, outcome == driveStop
	}
	for {
		e.Log(fmt.Sprintf("[Debug] Drive Loop: %d, %s", iterations, e.runtimeStateLabel()))
		iterations++
		// 如果有待处理的中断，不自动推进
		if e.State.PendingInterrupt != nil {
			return
		}
		// 仅在没有待处理延迟伤害时推进“延迟后续”。
		// 这样可保证诸如“封印伤害先结算，再继续技能后续”的严格顺序。
		if !e.isDamageResolutionActive() &&
			len(e.State.PendingDamageQueue) == 0 &&
			len(e.State.DeferredFollowups) > 0 {
			e.processDeferredFollowups()
			if e.State.PendingInterrupt != nil {
				return
			}
			continue
		}

		// 行动收尾：先跑行动结束后的全局 hook，再输出汇总信息。
		if e.runActionFinalizeHooksIfIdle() {
			if e.State.PendingInterrupt != nil {
				return
			}
			if !e.isActionFinalizeIdle() {
				continue
			}
		}

		// 行动汇总：当系统回到可继续行动的空闲状态时输出汇总信息
		e.finalizeActionSummaryIfIdle()

		currentPid := e.State.PlayerOrder[e.State.CurrentTurn]
		player := e.State.Players[currentPid]
		if handled, shouldStop := resolveDriveOutcome(e.driveNonTurnPhase(currentPid, player)); handled {
			if shouldStop {
				return
			}
			continue
		}
		if handled, shouldStop := resolveDriveOutcome(e.driveTurnFSM(currentPid, player)); handled {
			if shouldStop {
				return
			}
			continue
		}
		return
	}
}
