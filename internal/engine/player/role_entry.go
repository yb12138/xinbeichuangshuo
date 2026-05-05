// gameflow: 玩家角色入口定义。

package player

import (
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// CombatPolicyType 战斗策略类型（用于 TimingHook 分派）。
type CombatPolicyType string

const (
	CombatPolicyDefendValidation    CombatPolicyType = "defend_validation"
	CombatPolicyMagicMissileDefend  CombatPolicyType = "magic_missile_defend"
	CombatPolicyMagicMissileCounter CombatPolicyType = "magic_missile_counter"
)

// CombatPolicyRuntime 战斗策略运行时接口（用于 TimingHook 中需要战斗上下文的场景）。
type CombatPolicyRuntime interface {
	HookRuntime
	GetCombatRequest() *model.CombatRequest
	GetMagicBulletChain() *model.MagicBulletChain
}

// StateReader 状态读取器，暴露 GameState 的只读访问。
type StateReader interface {
	// 玩家
	GetPlayers() map[string]*model.Player
	GetPlayerOrder() []string
	GetCurrentTurnIndex() int

	// 阵营士气
	GetRedMorale() int
	GetBlueMorale() int

	// 中断与队列
	GetPendingInterrupt() *model.Interrupt
	GetPendingDamageQueue() []model.PendingDamage
	GetCombatStack() []model.CombatRequest
	GetActionQueue() []model.QueuedAction

	// 牌堆
	GetDiscardPile() []model.Card
	GetDeck() []model.Card

	// 阶段状态
	GetTurnStage() model.TurnStage
	GetCombatStage() model.CombatStage
	GetSubflow() model.Subflow

	// 魔弹链
	GetMagicBulletChain() *model.MagicBulletChain
}

// EffectCardOps 效果牌增删改查。
type EffectCardOps interface {
	FindEffectCard(source *model.Player, effect model.EffectType) (*model.Player, *model.FieldCard)
	AttachEffectCard(source, target *model.Player, effect model.EffectType, card model.Card) error
	DetachEffectCard(source *model.Player, effect model.EffectType) (*model.Player, model.Card, bool)
	RemoveEffectCard(source *model.Player, effect model.EffectType, restoreCard bool) bool
	EmitBuffRemovedDispatch(sourceID, targetID string, effect model.EffectType)
}

// InterruptOps 中断流程管理（仅保留 orchestrator 调用）。
type InterruptOps interface {
	PopInterrupt()
	NotifyInterruptPrompt()
	PushInterrupt(intr *model.Interrupt)
	PushDiscardChoiceInterrupt(playerID string, data map[string]interface{})
}

// StageOps 阶段切换。
type StageOps interface {
	EnterExtraActionStage()
	EnterTurnEndStage()
	EnterDamageResolution(returnTo interface{})
	EnterActionExecutionStage()
	EnterActionEndStage()
	EnterResponseWindow()
	ApplyChoiceResumePoint(raw interface{})
}

// DamageOps 伤害路由。
type DamageOps interface {
	RoutePendingDamageOr(defaultReturn interface{}, onNoPending func()) bool
	RoutePendingDamageWithReturn(returnTo interface{}) bool
	ResumePendingMoraleLoss(ctx *model.Context) bool
}

// TargetFilterRule 目标过滤规则 — 角色模块声明哪些玩家不能成为主动攻击目标。
type TargetFilterRule interface {
	CannotBeActiveAttackTarget(player *model.Player) bool
}

// TargetFilterRuleFuncs 函数式适配器。
type TargetFilterRuleFuncs struct {
	CannotBeTarget func(p *model.Player) bool
}

func (r TargetFilterRuleFuncs) CannotBeActiveAttackTarget(player *model.Player) bool {
	if r.CannotBeTarget == nil {
		return false
	}
	return r.CannotBeTarget(player)
}

// CombatOps 战斗操作。
type CombatOps interface {
	EnsureCombatInteractionWindow()
	ResolveCounterAttack(counterPlayerID, counterTargetID string, counterCard model.Card)
	NotifyCombatCue(attackerID, targetID, cueType string)
	ConsumePlayableCardByCardID(playerID, cardID string) (model.Card, bool)
	EnqueueVirtualAttack(sourceID, targetID string, card model.Card, sourceSkill string)
	ResumePendingAttackMiss(ctx *model.Context) bool
	ResumePendingAttackHit(ctxData map[string]interface{})
}

// DrawOps 抽牌流程。
type DrawOps interface {
	DrawCardsDirect(playerID string, amount int, reason string)
	DrawRawCards(amount int) ([]model.Card, bool)
	StartDraw(ctx *model.Context)
	NewDrawContext(player *model.Player, amount int, reason string) *model.Context
	RestorePhaseAfterInterruptedDraw(ctx *model.Context) bool
}

// HandOps 手牌上限 + 弃牌堆特殊操作。
type HandOps interface {
	RoleFixedMaxHandCapValue(player *model.Player) (int, bool)
	TakeDiscardPileCardByID(cardID string) (model.Card, bool)
}

// PoseOps 姿态快照。
type PoseOps interface {
	PoseChangeGuard() func()
}

// MagicBulletOps 魔弹系统。
type MagicBulletOps interface {
	SetMagicBulletChain(chain *model.MagicBulletChain)
	GetPlayableCardByIndex(player *model.Player, idx int) (model.Card, bool)
	ConsumePlayableCardByIndex(player *model.Player, idx int) (model.Card, error)
	PerformMagic(playerID, targetID string, cardIdx int, isFusion bool) error
	ExecuteMagicBullet(player *model.Player, reverse, isFusion bool, fusionCard *model.Card) error
	FindNextMagicBulletTarget(playerID string) string
	DispatchHitCheckMagicMissileCounter(player *model.Player, chain *model.MagicBulletChain, card *model.Card) error
	DispatchHitCheckMagicMissileDefend(player *model.Player, chain *model.MagicBulletChain) error
}

// SkillOps 技能状态与记录。
type SkillOps interface {
	IsSkillStillUsable(skillID string, user *model.Player, ctx *model.Context) bool
	RecordSkillUsage(playerID, title string, skillType model.SkillType)
	IsActionSkillUsableForExtraMagic(player *model.Player, skillDef model.SkillDefinition) bool
	RecordMagicDamageTarget(sourceID, targetID string)
	MagicDamageTargetCount(sourceID string) int
}

// MoraleOps 士气操作（仅保留有逻辑的）。
type MoraleOps interface {
	AddCampMorale(camp model.Camp, amount int) int
	ApplyCampMoraleLoss(camp model.Camp, wantLoss int) int
}

// GameOps 游戏状态。
type GameOps interface {
	CheckGameEnd()
	RefreshAllPlayerDerivedStates()
	BuildContext(user, target *model.Player, timing model.FlowTiming, eventCtx *model.EventContext) *model.Context
}

// CombatPolicyContext 战斗策略上下文。
type CombatPolicyContext struct {
	Player        *model.Player
	CombatRequest *model.CombatRequest
	Chain         *model.MagicBulletChain
	Card          model.Card
	CounterCard   *model.Card
	UseFaction    bool
}

// InterruptSpec 定义角色包贡献的中断处理条目。
type InterruptSpec struct {
	Type                 model.InterruptType
	BuildPrompt          func(rt ChoiceRuntime) *model.Prompt
	HandleAction         func(rt ChoiceRuntime, act model.PlayerAction) error
	AllowedActionTypes   []model.PlayerActionType
	InvalidActionMessage string
}

// ChoiceSpec 定义声明式选择流程条目。
type ChoiceSpec struct {
	ChoiceType   string
	BuildPrompt  func(rt ChoiceRuntime, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt
	HandleChoice func(rt ChoiceRuntime, playerID string, selectionIndex int, data map[string]interface{}) (bool, error)
}

// ChoiceRuntime 抽象角色选择流运行时能力（嵌入子接口的组合接口）。
type ChoiceRuntime interface {
	model.IGameEngine
	StateReader    // 状态读取（通用字段访问）
	EffectCardOps  // 效果牌增删改查
	InterruptOps   // 中断流程管理
	StageOps       // 阶段切换
	DamageOps      // 伤害路由
	CombatOps      // 战斗操作
	DrawOps        // 抽牌流程
	HandOps        // 手牌上限 + 弃牌堆特殊操作
	PoseOps        // 姿态快照
	MagicBulletOps // 魔弹系统
	SkillOps       // 技能状态与记录
	MoraleOps      // 士气操作
	GameOps        // 游戏状态

	// PendingDamage direct access for choice handlers
	GetPendingDamage() *model.PendingDamage
	GetPendingDamageByIndex(index int) (*model.PendingDamage, bool)

	// PendingDiscard victim helper
	PendingDiscardVictimID() string

	// Convenience methods for backward compatibility
	LookupPlayer(playerID string) *model.Player
	HasPendingInterrupt() bool
	PendingDamageQueueLen() int
	ActionQueueLen() int
	AllPlayers() []*model.Player
	ReplacePendingInterruptContext(data map[string]interface{}) error
	ReplacePendingInterruptPlayerID(playerID string)
	PendingInterrupt() *model.Interrupt
	AddToDiscardPile(cards ...model.Card)
	SetReturnPoint(returnTo interface{})
	MagicBulletChain() *model.MagicBulletChain
	PlayerOrder() []string
	TopCombatRequest() *model.CombatRequest
	CampEnemyIDs(camp model.Camp) []string
	AllOtherPlayerIDs(userID string) []string
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

// SkillUsabilityChecker 技能可用性检查器。
type SkillUsabilityChecker func(engine SkillUsabilityCheckerEngine, player *model.Player, skillDef model.SkillDefinition) bool

// SkillUsabilityCheckerEngine 抽象技能可用性检查所需的引擎能力。
type SkillUsabilityCheckerEngine interface {
	LookupPlayer(playerID string) *model.Player
	PlayerOrder() []string
	GetAllPlayers() []*model.Player
	HasForm(player *model.Player, form string) bool
	IsCharacter(player *model.Player, roleID string) bool
	GetToken(player *model.Player, key string) int
	CountCoverCardsByEffectAndElement(player *model.Player, effect model.EffectType, element model.Element) int
}

// MoraleLossModifierEngine 士气损失修改器所需的引擎能力。
type MoraleLossModifierEngine interface {
	GetAllPlayers() []*model.Player
}

// MoraleLossModifierExtra 士气损失修改器的额外上下文（可选）。
type MoraleLossModifierExtra struct {
	Victim             *model.Player
	FromDamageDraw     bool
	IsDamageResolution bool
}

// MoraleLossModifier 调整指定阵营的士气损失量。
// 接收 (engine, camp, currentMorale, proposedLoss, extra)，返回修改后的损失量。
// 按 roleRegistry 顺序链式调用，前一个修改器的输出是后一个的输入。
type MoraleLossModifier func(engine MoraleLossModifierEngine, camp model.Camp, currentMorale int, proposedLoss int, extra MoraleLossModifierExtra) int

// CannotActChecker 无法行动判断函数。
// 返回 (canCannotAct, reason)：
//   - canCannotAct=true: 该角色认为当前可以宣告无法行动
//   - canCannotAct=false: 该角色认为当前不能宣告无法行动，继续走默认判断
//
// player 参数已包含手牌、形态等信息，大多数角色只需要检查 player 即可。
type CannotActChecker func(player *model.Player) (bool, string)

// BuyRewardResult 购买奖励改写结果（纯数据，由引擎应用）。
type BuyRewardResult struct {
	Handled     bool   // true 表示已处理（跳过默认奖励逻辑）
	AddGems     int    // 向阵营战绩区添加的宝石数
	AddCrystals int    // 向阵营战绩区添加的水晶数
	LogMessage  string // 日志消息
}

// SpecialActionHookSpec 特殊行动钩子：角色模块可改写特殊行动的执行内容。
type SpecialActionHookSpec struct {
	// BuyRewardOverride 改写购买奖励（如冒险者地下法则：+2宝石代替+1宝石+1水晶）。
	BuyRewardOverride func(p *model.Player, campStones int, maxStones int) BuyRewardResult
}

// MagicBulletAbilities 魔弹相关能力声明（数据驱动，引擎不再硬编码角色 ID）。
type MagicBulletAbilities struct {
	CanFuse   bool // 可将地系或火系法术牌当魔弹使用
	CanDirect bool // 可选择魔弹传递方向
}

// FlowContinuationHandler 角色流程边界处理函数（用函数类型，不用接口）。
type FlowContinuationHandler func(rt ChoiceRuntime, cont model.FlowContinuation) error

// RoleEntry 表示单个角色在 player 子目录的统一入口定义。
type RoleEntry struct {
	ID                         string
	Defaults                   func(player *model.Player)
	StarterCards               func(player *model.Player) []model.Card
	HandLimit                  HandLimitRule
	TargetFilter               TargetFilterRule
	EnergyCapRule              EnergyCapRule
	SpecialActionHook          SpecialActionHookSpec
	MagicBullet                MagicBulletAbilities
	MaxHeal                    func(player *model.Player, current int) int
	Choices                    ChoiceHandler
	ChoiceSpecs                []ChoiceSpec
	Skills                     []SkillEntry
	ChoiceRouteSpecs           map[string]types.ChoiceRouteSpec
	FlowContinuationHandlers   map[model.FlowContinuationKind]FlowContinuationHandler // 流程边界处理器（替代 FollowupSpecs）
	InterruptSpecs             []InterruptSpec
	TimingHookSpecs            []TimingHookSpec
	SkillUsabilityCheckers     map[string]SkillUsabilityChecker
	AttackCardElementTransform func(player *model.Player, card model.Card) model.Element
	AttackElementResolver      func(player *model.Player, card model.Card) model.Element                            // 攻击牌元素解析（可选，如烈焰魔女火焰形态）
	CannotActChecker           CannotActChecker                                                                     // 角色自定义无法行动判断hook（可选）
	HandLimitModifier          HandLimitModifier                                                                    // 全局手牌上限修改器（可选，如血之巫女同生共死）
	MoraleLossModifier         MoraleLossModifier                                                                   // 全局士气损失修改器（可选，如蝶舞者枯萎）
	BlocksActionType           func(player *model.Player, actionType model.ActionType) bool                         // 行动类型限制（可选）
	PlayableCoverEffects       []model.EffectType                                                                   // 可作为可打牌使用的盖牌效果类型（可选）
	ExcludeCardFromDiscard     func(player *model.Player, card model.Card) bool                                     // 弃牌时排除特定卡牌（可选，如精灵射手祝福牌）
	AfterMoraleLossHook        func(rt model.IGameEngine, victim *model.Player, finalLoss int, fromDamageDraw bool) // 士气损失后置钩子（可选，如灵魂巫师灵魂吞噬）
	MaybeDarkRitual            func(rt ChoiceRuntime, player *model.Player) bool                                    // 暗仪发动检查（可选，如阴阳师暗仪）
	AfterDiscardFollowup       func(rt ChoiceRuntime, player *model.Player) bool                                    // 弃牌后续处理（可选，如魔枪幻影星尘/魔弓魔眼）；返回 true 表示已接管后续流程
	ConsumeTauntRestriction    func(rt ChoiceRuntime, player *model.Player)                                         // 挑衅约束消耗（可选，如勇者挑衅）
	ResolveChrysalis           func(rt ChoiceRuntime, userID string) error                                          // 蛹化直接结算（可选，如蝶舞者蛹化）
	StartReverse               func(rt ChoiceRuntime, userID string) error                                          // 倒逆之蝶分支编排（可选，如蝶舞者倒逆之蝶）
}

// ApplyDefaults 应用默认角色属性。
func (e RoleEntry) ApplyDefaults(player *model.Player) {
	if e.Defaults == nil {
		return
	}
	e.Defaults(player)
}

// ApplyStarterCards 应用开局专属牌。
func (e RoleEntry) ApplyStarterCards(player *model.Player) []model.Card {
	if e.StarterCards == nil {
		return nil
	}
	return e.StarterCards(player)
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
