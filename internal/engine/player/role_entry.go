// gameflow: 玩家角色入口定义。

package player

import (
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// ChoiceRuntime 抽象角色选择流运行时能力。
type ChoiceRuntime interface {
	model.IGameEngine
	LookupPlayer(playerID string) *model.Player
	AllPlayers() map[string]*model.Player
	PlayerOrder() []string
	PopInterrupt()
	HasPendingInterrupt() bool
	NotifyInterruptPrompt()
	ReplacePendingInterruptContext(data map[string]interface{}) error
	ResumePendingAttackMiss(ctx *model.Context) bool
	ResumePendingAttackHit(ctxData map[string]interface{})
	ApplyChoiceResumePoint(raw interface{})
	RoutePendingDamageOr(defaultReturn interface{}, onNoPending func()) bool
	RoutePendingDamageWithReturn(returnTo interface{}) bool
	EnterExtraActionStage()
	EnterTurnEndStage()
	EnterDamageResolution(returnTo interface{})
	EnterActionExecutionStage()
	AllOtherPlayerIDs(userID string) []string
	PlayerOrderPosition(playerID string) int
	StartDraw(ctx *model.Context)
	NewDrawContext(player *model.Player, amount int, reason string) *model.Context
	RestorePhaseAfterInterruptedDraw(ctx *model.Context) bool
	PushInterrupt(intr *model.Interrupt)
	PendingDamageQueueLen() int
	ActionQueueLen() int
	AttachExclusiveEffectCard(sourceID, targetID string, effect model.EffectType, card model.Card) error
	ResumePendingMoraleLoss(ctx *model.Context) bool
	EnterResponseWindow()
	ApplyStealthEffect(player *model.Player)
}

// ChoiceHandler 抽象角色选择流入口。
type ChoiceHandler interface {
	BuildPrompt(rt ChoiceRuntime, choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt
	HandleChoice(rt ChoiceRuntime, playerID string, selectionIndex int, ctxData map[string]interface{}) (bool, error)
}

// CancelChoiceHandler 为支持取消的角色选择流提供扩展入口。
type CancelChoiceHandler interface {
	HandleCancel(rt ChoiceRuntime, playerID string, ctxData map[string]interface{}) (bool, error)
}

// SkillEntry 定义角色技能与策略绑定入口。
type SkillEntry struct {
	ID      string
	Handler model.SkillHandler
	Policy  types.SkillPolicy
}

// FollowupHost 抽象延迟后续回调可用的最小引擎能力。
type FollowupHost interface {
	Log(message string)
	LookupPlayer(playerID string) *model.Player
	ResolveSkillFollowup(req ResolveSkillFollowupReq) error
}

// FollowupSpec 定义角色模块贡献给延迟后续执行表的条目。
type FollowupSpec struct {
	Label   string
	Resolve func(host FollowupHost, f model.DeferredFollowup) error
}

// ResolveSkillFollowupReq 是角色 followup resolver 发起引擎后续结算的通用请求。
type ResolveSkillFollowupReq struct {
	Kind     string
	Followup model.DeferredFollowup
}

// RoleEntry 表示单个角色在 player 子目录的统一入口定义。
type RoleEntry struct {
	ID string
	// Defaults 用于初始化角色默认属性。
	Defaults func(player *model.Player)
	// HandLimit 角色手牌上限规则。
	HandLimit HandLimitRule
	// MaxHeal 角色治疗上限规则。
	MaxHeal func(player *model.Player, current int) int
	// Choices 角色专属选择流程。
	Choices ChoiceHandler
	// Skills 角色技能与策略绑定入口（统一机制）。
	Skills []SkillEntry
	// ChoiceRouteSpecs 角色贡献到全局 choice 路由映射的条目。
	ChoiceRouteSpecs map[string]types.ChoiceRouteSpec
	// FollowupSpecs 角色贡献到全局 DeferredFollowups 执行映射的条目。
	FollowupSpecs map[string]FollowupSpec
}

// ApplyDefaults 应用默认角色属性。
func (e RoleEntry) ApplyDefaults(player *model.Player) {
	if e.Defaults == nil {
		return
	}
	e.Defaults(player)
}

// HandLimitRule 返回角色手牌规则。
func (e RoleEntry) HandLimitRule() HandLimitRule {
	if e.HandLimit == nil {
		return DefaultHandLimitRule
	}
	return e.HandLimit
}

// ApplyMaxHeal 应用治疗上限规则。
func (e RoleEntry) ApplyMaxHeal(player *model.Player, current int) int {
	if e.MaxHeal == nil {
		return current
	}
	return e.MaxHeal(player, current)
}

// BuildChoicePrompt 构建角色选择提示。
func (e RoleEntry) BuildChoicePrompt(rt ChoiceRuntime, choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	if e.Choices == nil {
		return nil
	}
	return e.Choices.BuildPrompt(rt, choiceType, playerID, player, data)
}

// HandleChoice 处理角色选择输入。
func (e RoleEntry) HandleChoice(rt ChoiceRuntime, playerID string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	if e.Choices == nil {
		return false, nil
	}
	return e.Choices.HandleChoice(rt, playerID, selectionIndex, ctxData)
}

// HandleChoiceCancel 处理角色选择流的取消事件。
func (e RoleEntry) HandleChoiceCancel(rt ChoiceRuntime, playerID string, ctxData map[string]interface{}) (bool, error) {
	if e.Choices == nil {
		return false, nil
	}
	handler, ok := e.Choices.(CancelChoiceHandler)
	if !ok {
		return false, nil
	}
	return handler.HandleCancel(rt, playerID, ctxData)
}

// ChoiceRoutes 返回角色 choice 路由声明。
func (e RoleEntry) ChoiceRoutes() map[string]types.ChoiceRouteSpec {
	return e.ChoiceRouteSpecs
}

// SkillPolicies 返回角色技能策略声明。
func (e RoleEntry) SkillPolicies() map[string]types.SkillPolicy {
	if len(e.Skills) == 0 {
		return nil
	}
	out := make(map[string]types.SkillPolicy, len(e.Skills))
	for _, skill := range e.Skills {
		if skill.ID != "" {
			out[skill.ID] = skill.Policy
		}
	}
	return out
}

// Followups 返回角色延迟后续声明。
func (e RoleEntry) Followups() map[string]FollowupSpec {
	return e.FollowupSpecs
}
