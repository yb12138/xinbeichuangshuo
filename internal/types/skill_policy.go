// gameflow: 技能策略类型定义（共享包，避免 engine/player 循环依赖）。

package types

import "starcup-engine/internal/model"

// SkillPolicy 定义技能使用策略钩子。
type SkillPolicy struct {
	// PrepareSkillUse 阶段覆盖默认 CostDiscards，决定后续弃牌交互/校验需要的张数。
	ResolveDiscardCount func(ctx PolicyContext) int
	// ValidateSkillDiscardSelection 阶段的二次校验，用于检查弃牌组合规则。
	ValidateDiscardedCards func(ctx PolicyContext) error
	// ResolveSkillTargets 阶段的声明式目标规则。
	TargetRules TargetRuleSet
	// ConsumeSkillInputs 阶段自定义弃牌入弃牌堆行为。
	ResolveDiscardPile func(ctx PolicyContext) []model.Card
	// ExecuteSkillFlow 阶段的后置钩子。
	AfterConsume func(host PolicyHost, ctx PolicyContext) (bool, error)
	// ExecuteSkillFlow 阶段 handler 成功执行后的收尾钩子。
	AfterExecute func(host PolicyHost, ctx PolicyContext) error
	// FinishSkillUse 阶段是否跳过默认收尾。
	SkipAutoPhaseEnd bool
	// ValidateSkillDiscardSelection 阶段对专属卡改为手动处理。
	ManualExclusiveCard bool
	// 响应链中同组技能互斥；发动其中一个后，本组其他响应技能不再继续提供。
	ExclusiveResponseGroup string
}

// PolicyContext 是技能策略回调使用的只读快照。
type PolicyContext struct {
	SkillID          string
	PlayerID         string
	SkillDef         model.SkillDefinition
	RequiredDiscards int
	DiscardedCards   []model.Card
	TargetIDs        []string
	ActualTargetIDs  []string
}

// PolicyHost 抽象角色策略可调用的引擎能力。
type PolicyHost interface {
	Log(message string)
	DropQueuedOverflowDiscardForPlayer(playerID string)
}
