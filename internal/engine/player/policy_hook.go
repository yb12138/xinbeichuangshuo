// gameflow: 策略 Hook 统一类型定义。
// 采用与 TimingHookSpec 相同的声明式架构，让角色包自主声明策略。

package player

import (
	"starcup-engine/internal/model"
)

// ---------- 策略类型常量 ----------

type PolicyType string

const (
	// 战斗阶段策略
	PolicyCombatInteraction      PolicyType = "combat_interaction"       // 战斗交互中断
	PolicyCombatCounterCard      PolicyType = "combat_counter_card"      // 反击卡牌策略
	PolicyCombatCounterElement   PolicyType = "combat_counter_element"   // 反击元素策略
	PolicyCombatCounterResolve   PolicyType = "combat_counter_resolve"   // 反击结算策略
	PolicyResponseSkillAugment   PolicyType = "response_skill_augment"   // 响应技能增强
	PolicyResponseSkillNormalize PolicyType = "response_skill_normalize" // 响应技能规范化
	PolicyDamageAfterResolved    PolicyType = "damage_after_resolved"    // 伤害结算后钩子

	// 行动选择策略
	PolicyBeforeActionOption     PolicyType = "before_action_option"     // 行动选项策略
	PolicyBeforeActionValidation PolicyType = "before_action_validation" // 行动验证策略

	// 技能/特殊行动策略
	PolicySkillPost             PolicyType = "skill_post"              // 技能后置钩子
	PolicySpecialActionOverride PolicyType = "special_action_override" // 特殊行动覆盖
	PolicySpecialActionPost     PolicyType = "special_action_post"     // 特殊行动后置

	// 攻击阶段策略
	PolicyAttackDeclaredInterrupt PolicyType = "attack_declared_interrupt" // 攻击宣言中断
	PolicyAttackCardTransform     PolicyType = "attack_card_transform"     // 攻击牌变换
)

// ---------- 统一策略 Hook 签名 ----------

// PolicyHookFunc 统一策略 Hook 函数签名。
// 所有策略 Hook 采用相同签名，通过 PolicyHookContext 容器传递参数，PolicyHookResult 容器返回结果。
type PolicyHookFunc func(host PolicyHost, ctx PolicyHookContext) PolicyHookResult

// ---------- 策略声明结构 ----------

// PolicySpec 策略声明条目。
// 角色模块在 RoleEntry.PolicySpecs 中声明，引擎启动时收集装配到对应策略表。
type PolicySpec struct {
	Type     PolicyType
	Priority int // 数值越小越先执行
	Hook     PolicyHookFunc
}

// ---------- 输入容器 ----------

// PolicyHookContext 策略 Hook 输入参数容器。
// 包含所有策略类型可能需要的字段，具体策略按需使用。
type PolicyHookContext struct {
	// 战斗相关
	CombatRequest  *model.CombatRequest
	CounterCard    model.Card
	CounterCardPtr *model.Card
	UseFaction     bool
	MagicChain     *model.MagicBulletChain

	// 行动相关
	Player       *model.Player
	Attacker     *model.Player
	Target       *model.Player
	Action       *model.QueuedAction
	UserCtx      *model.Context
	PlayerAction model.PlayerAction
	ActionType   model.ActionType // 特殊行动类型

	// 响应技能
	SkillIDs []string

	// 伤害相关
	PendingDamage *model.PendingDamage
	MaxHeal       int

	// 技能后置相关
	SkillID string // 技能 ID

	// 行动选择策略相关
	ChoiceRuntime      ChoiceRuntime                     // 选择运行时（用于行动选择策略）
	OptionModifier     ActionSelectionModifier           // 行动选项修改器
	ValidationModifier ActionSelectionValidationModifier // 行动验证修改器

	// 其他
	ResponseState map[string]any
}

// ---------- 输出容器 ----------

// PolicyHookResult 策略 Hook 返回结果容器。
type PolicyHookResult struct {
	Handled      bool               // 是否处理了此策略
	Stop         bool               // 短路信号（阻止后续 Hook 或流程继续）
	Card         model.Card         // 返回卡牌
	UseFaction   bool               // 使用命格应战
	MaxHeal      int                // 治疗上限修正
	Err          error              // 错误（如验证失败）
	SkillIDs     []string           // 增强后的技能列表
	PlayerAction model.PlayerAction // 返回修改后的行动
}

// ---------- 策略 Host 接口 ----------

// PolicyHost 策略 Hook 可用的引擎能力接口。
// 角色包的策略 Hook 函数通过此接口调用引擎方法，避免直接导入 engine 包。
type PolicyHost interface {
	// 基础能力
	Log(message string)
	LookupPlayer(playerID string) *model.Player
	AllPlayers() map[string]*model.Player
	State() *model.GameState
	PlayerOrder() []string
	CurrentTurn() int

	// 中断/流程控制
	PushInterrupt(intr *model.Interrupt)
	PopInterrupt()
	NotifyCardRevealed(playerID string, cards []model.Card, actionType model.DamageType)
	NotifyCombatCue(attackerID, targetID, cueType string)
	ConsumeCardByIndex(player *model.Player, idx int) (model.Card, error)
	AddToDiscardPile(cards ...model.Card)

	// 战斗相关
	InitCombat(attackerID, targetID string, card *model.Card, isForcedHit, canBeResponded, ignoreShield bool, interceptTags map[model.CombatInterceptTag]bool, isCounter ...bool)
	ResolveMagicBowPierceMiss(attackerID, targetID string, attackCard *model.Card, isCounter bool)
	TopCombatRequest() *model.CombatRequest
	PopCombatRequest()
	BuildContext(user, target *model.Player, timing model.FlowTiming, eventCtx *model.EventContext) *model.Context
	DispatchOnTiming(ctx *model.Context)
	PendingInterrupt() *model.Interrupt

	// 角色判断
	IsCharacter(player *model.Player, roleID string) bool
	HasForm(player *model.Player, form string) bool
	GetToken(player *model.Player, key string) int

	// 战斗策略委托方法（核心逻辑在 engine 包，角色包 Hook 只做包装）
	ApplyFactionCounterBonuses(actor *model.Player, card *model.Card)
	CanUseFactionCounter(incoming *model.Card) bool

	// 魔剑士委托方法
	CanUseShadowRejectResponse(player *model.Player, currentTurnPlayerID string) bool

	// 其他角色策略委托方法
	ApplyMoonMedusaInterrupt(attacker, target *model.Player, action *model.QueuedAction, ctx *model.Context) bool
	ApplyBeastSamuraiResponseSkillAugment(skillIDs []string, ctx *model.Context) []string
	ApplyFighterResponseSkillNormalize(skillIDs []string, ctx *model.Context) []string
	ApplyArbiterSkillPostCleanup(ctx *model.Context)
	ApplyAdventurerUndergroundLawOverride(player *model.Player, action model.PlayerAction) (model.PlayerAction, bool)
	ApplyHolyBowHolyGloryExitHook(player *model.Player, actionType model.ActionType)
	ApplyBlazeWitchAttackCardTransform(player *model.Player, card model.Card) model.Card
	HandlePostDamageResolved(pd *model.PendingDamage) bool
}

// ---------- 行动选择策略接口（兼容旧架构） ----------

// ActionSelectionModifier 抽象行动选择状态的可变字段。
type ActionSelectionModifier interface {
	SetActionRule(mode string, source string, priority int)
	SetCanMagicAction(v bool)
	SetCanMagicSkillAction(v bool)
	SetPromptChoiceType(ct string)
	SetPromptSkillID(sid string)
	SetActionRulePromptMessage(msg string)
	SetConstrainedTarget(id, name string)
	SetRuleRequiresSkipOnly(v bool)
}

// ActionSelectionValidationModifier 扩展行动选择校验阶段可用的回调设置。
type ActionSelectionValidationModifier interface {
	ActionSelectionModifier
	SetRequiredSkillID(sid string)
	SetForceSkillMustUseMessage(msg string)
	SetForceSkillOnlyMessage(msg string)
	SetForceAttackOnlyMessage(msg string)
	SetOnSkipChosen(callback func(rt ChoiceRuntime, player *model.Player) (bool, error))
	SetOnNonAttackChosen(callback func(rt ChoiceRuntime, player *model.Player, act model.PlayerAction) error)
	SetOnAttackAccepted(callback func(rt ChoiceRuntime, player *model.Player, act model.PlayerAction) error)
	MarkConsumeHeroTauntOnAttack() // 标记攻击后消耗挑衅效果
}
