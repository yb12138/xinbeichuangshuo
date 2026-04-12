// gameflow: Skill Runtime 与 GameEngine 之间的宿主接口（避免 runtime 依赖 engine 包）。

package skill

import "starcup-engine/internal/model"

// Host 由 engine 侧实现，提供技能执行与中断所需的引擎能力。
type Host interface {
	Log(msg string)
	GameState() *model.GameState

	SnapshotPlayerPoses() any
	DispatchOrientationChanges(before any)
	SyncPendingDamageFromContext(ctx *model.Context)
	RecordSkillUsage(playerID, title string, skillType model.SkillType)

	ApplyHitCheckAugment(skillIDs []string, ctx *model.Context) []string
	ApplyHitCheckNormalize(skillIDs []string, ctx *model.Context) []string

	PublishStartupInterrupt(playerID string, skillIDs []string, sharedCtx *model.Context)
	PublishResponseInterrupt(player *model.Player, skillIDs []string, sharedCtx *model.Context)
	OnStartupInterruptPublished()

	GetMaxHand(player *model.Player) int
	DropQueuedOverflowDiscardForPlayer(playerID string)

	PopInterrupt()
	SetPendingInterrupt(intr *model.Interrupt)
	PendingInterrupt() *model.Interrupt
	EnterDiscardSelection()
	NotifyInterruptPrompt()

	CaptureResponseResumeStateOnConfirm(skillID string, ctx *model.Context) any
	PrepareConfirmedResponseResume(state any)
	RestoreConfirmedResponseAfterPop(state any)
}
