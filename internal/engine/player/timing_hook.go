// gameflow: Timing Hook 声明式注册类型。

package player

import "starcup-engine/internal/model"

// TimingPoint 标识 Hook 触发时机。
type TimingPoint string

const (
	TimingPostActionEnd      TimingPoint = "post_action_end"
	TimingPostAttackHit      TimingPoint = "post_attack_hit"
	TimingPostDamageResolved TimingPoint = "post_damage_resolved"
)

// TimingHookSpec 角色贡献到全局 timing hook 链的条目。
type TimingHookSpec struct {
	Timing   TimingPoint
	Priority int // 数值越小越先执行
	Hook     TimingHookFunc
}

// TimingHookContext 传递给 Hook 的上下文。
type TimingHookContext struct {
	SourceID      string
	TargetID      string
	ActionType    model.ActionType     // post_action_end
	DamageType    model.DamageType     // post_damage_resolved / post_attack_hit
	Damage        int                  // post_damage_resolved
	IsCounter     bool                 // post_attack_hit
	Card          *model.Card          // post_attack_hit
	PendingDamage *model.PendingDamage // 原始 PD（可选）
}

// TimingHookResult Hook 执行结果。
type TimingHookResult struct {
	Interrupted bool // true = 产生了中断，状态机应暂停
}

// TimingHookFunc 统一 Hook 签名。
type TimingHookFunc func(rt HookRuntime, ctx TimingHookContext) TimingHookResult

// HookRuntime 抽象 Timing Hook 运行时能力（窄接口）。
type HookRuntime interface {
	Log(message string)
	GetPlayer(playerID string) *model.Player
	PushInterrupt(intr *model.Interrupt)
	PushDiscardChoiceInterrupt(playerID string, data map[string]interface{})
	Heal(targetID string, amount int)
	AddPendingDamage(pd model.PendingDamage)
	GetMaxHand(player *model.Player) int
	GetPlayerEnergyCap(player *model.Player) int
	DrawCards(playerID string, amount int)
	GetPendingDamageQueue() []model.PendingDamage
	SetPendingDamageQueue(queue []model.PendingDamage)
	SnapshotPlayerPoses() map[string]PoseSnapshot
	DispatchOrientationChanges(before map[string]PoseSnapshot)
	// Phase 4: handlePostDamageResolved 专用委托方法
	CampMorale(camp model.Camp) int
	HasPendingDiscardFor(playerID string) bool
	PlayerOrder() []string
}

// PoseSnapshot 记录玩家姿态快照（用于 orientation 变更前后对比）。
type PoseSnapshot struct {
	Orientation model.CharacterOrientation
	Form        string
}
