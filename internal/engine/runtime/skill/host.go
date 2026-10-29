// gameflow: Skill Runtime 与 GameEngine 之间的宿主接口（避免 runtime 依赖 engine 包）。

package skill

import "starcup-engine/internal/model"

// Host 由 engine 侧实现，提供技能执行与中断所需的引擎能力。
type Host interface {
	Log(msg string)
	GameState() *model.GameState

	SnapshotPlayerPoses() any
	DispatchOrientationChanges(before any)
	SnapshotPendingActions(player *model.Player) any
	PublishPendingActionDiff(player *model.Player, before any, source string)
	SyncPendingDamageFromContext(ctx *model.Context)
	RecordSkillUsage(playerID, title string, skillType model.SkillType)
	NotifySkillActivated(playerID, skillID, skillName, effectText string, targetIDs []string)

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

	ConsumeSkillEnergyCost(playerID string, gemCost, crystalCost int) bool

	// RunInSkillCause 在技能因果下执行 fn（场上牌技能以放置时的因果为 Origin）。
	RunInSkillCause(skill model.SkillDefinition, ctx *model.Context, fn func())
}
