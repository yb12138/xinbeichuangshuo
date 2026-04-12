// gameflow: Orchestrator 对 GameEngine 的最小依赖面。

package interrupt

import "starcup-engine/internal/model"

// EngineInterface 由 *engine.GameEngine 实现。
type EngineInterface interface {
	GetState() *model.GameState
	GetPlayerByID(playerID string) *model.Player
	Log(message string)
	NotifyInterruptPrompt()
	// ApplyInterruptPhase 在挂起中断成为当前 Pending 时同步子流程/回合/战斗阶段。
	ApplyInterruptPhase(intr *model.Interrupt)
	// ReconcileSubflowAfterInterruptPop 在弹出中断后收敛子流程（如弃牌选择）。
	ReconcileSubflowAfterInterruptPop(popped *model.Interrupt)
}
