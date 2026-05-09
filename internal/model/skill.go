package model

// Enums for Skill System

// SkillType 技能类型
type SkillType int

const (
	SkillTypePassive  SkillType = iota // 被动 (无需操作，自动触发)
	SkillTypeStartup                   // 启动 (回合开始前主动使用)
	SkillTypeAction                    // 法术/攻击 (行动阶段主动使用)
	SkillTypeResponse                  // 响应 (在特定事件发生时插入使用)

)

// SkillTag 技能标签
type SkillTag string

const (
	TagGem       SkillTag = "Gem"       // [宝石] 消耗
	TagCrystal   SkillTag = "Crystal"   // [水晶] 消耗
	TagTurnLimit SkillTag = "TurnLimit" // [回合限定]
	TagForce     SkillTag = "Force"     // [强制]
	TagUnique    SkillTag = "Unique"    // [独有]
	TagUltimate  SkillTag = "Ultimate"  // [大招]
	TagExclusive SkillTag = "Exclusive" // [专属]
	TagOptional  SkillTag = "Optional"  // [可选] 响应技能需要玩家确认
	TagDamage    SkillTag = "Damage"    // 造成伤害
	TagHeal      SkillTag = "Heal"      // 治疗
	TagControl   SkillTag = "Control"   // 控制
	TagDiscard   SkillTag = "Discard"   // 弃牌
)

// 1. 新增：目标类型枚举 (前端根据这个决定让玩家点谁)
type TargetType int

const (
	TargetNone     TargetType = iota // 无需目标 (如: 购买、自我Buff)
	TargetSelf                       // 必须选自己
	TargetEnemy                      // 任意敌人
	TargetAlly                       // 任意队友 (不含自己)
	TargetAllySelf                   // 任意队友 (含自己) - 即"我方角色"
	TargetAny                        // 全场任意角色
	TargetSpecific                   // 特殊逻辑 (由后端检查，前端可能需要根据上下文高亮)
)

// ResponseType 响应类型
type ResponseType int

const (
	ResponseMandatory ResponseType = iota // 强制响应
	ResponseOptional                      // 可选响应
	ResponseSilent                        // 静默执行
)

// InteractionType 技能交互类型
type InteractionType string

const (
	InteractionNone    InteractionType = ""        // 无交互，直接执行
	InteractionDiscard InteractionType = "Discard" // 需要弃牌选择
)

// InteractionConfig 交互配置
type InteractionConfig struct {
	MinSelect int    `json:"min_select"` // 最少选择数量
	MaxSelect int    `json:"max_select"` // 最多选择数量
	Prompt    string `json:"prompt"`     // 交互提示信息
}

// SkillRole 定义技能响应时的身份限制 [新增]
type SkillRole string

const (
	RoleAny      SkillRole = ""         // 不限身份 (默认)
	RoleAttacker SkillRole = "Attacker" // 仅作为攻击发起者时可触发 (如: 追击, 吸血)
	RoleDefender SkillRole = "Defender" // 仅作为受击目标时可触发 (如: 水影, 反伤)
)

// SkillDefinition 技能静态配置模型 (增强版)
type SkillDefinition struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	CharacterID string     `json:"character_id"`
	Type        SkillType  `json:"type"`
	Tags        []SkillTag `json:"tags"`
	Description string     `json:"description"`
	// 同一触发窗口内的执行优先级（大者先执行）。
	Priority int `json:"priority"`

	// --- 自动化校验字段 ---
	CostGem      int `json:"cost_gem"`      // 基础宝石消耗 (若为-1代表需要动态计算)
	CostCrystal  int `json:"cost_crystal"`  // 基础水晶消耗
	CostDiscards int `json:"cost_discards"` // 基础弃牌数

	DiscardElement   Element  `json:"discard_element"`   // 要求的元素（可选）
	DiscardType      CardType `json:"discard_type"`      // 要求的卡牌类型（可选）
	DiscardFate      string   `json:"discard_fate"`      // 命格要求（可选）
	RequireExclusive bool     `json:"require_exclusive"` // 是否必须使用独有牌

	// --- 新增：场上牌放置 ---
	PlaceCard   bool          `json:"place_card"`   // 是否放置牌到场上
	PlaceMode   FieldCardMode `json:"place_mode"`   // 放置模式 (Effect/Cover)
	PlaceEffect EffectType    `json:"place_effect"` // 效果类型 (仅Effect模式)
	PlaceHook   FieldHook     `json:"place_hook"`   // 触发时机 (仅Effect模式)

	// --- 新增：盖牌消耗 ---
	CostCoverCards int `json:"cost_cover_cards"` // 消耗盖牌数量

	// --- 新增：交互配置 ---
	InteractionType   InteractionType   `json:"interaction_type"`   // 交互类型
	InteractionConfig InteractionConfig `json:"interaction_config"` // 交互配置

	// --- 新增：前端/交互引导字段 ---
	TargetType   TargetType `json:"target_type"` // 目标选择限制
	MinTargets   int        `json:"min_targets"` // 最小目标数
	MaxTargets   int        `json:"max_targets"` // 最大目标数 (通常为1，部分技能如 AOE 可能更多)
	RequiredRole SkillRole  `json:"required_role"`

	// Timings 是统一触发窗口声明（与引擎 OnTiming / collectSkillsForTiming 对齐）。
	Timings []FlowTiming `json:"timings"`

	ResponseType ResponseType `json:"response_type"`
	LogicHandler string       `json:"logic_handler"`
}

// HasTiming 报告 Timings 是否包含给定触发窗口。
func (s SkillDefinition) HasTiming(t FlowTiming) bool {
	for _, x := range s.Timings {
		if x == t {
			return true
		}
	}
	return false
}

// PrimaryTimingOrLegacy 返回 Timings 的首项；空则视为主动/即时窗口。
func (s SkillDefinition) PrimaryTimingOrLegacy() FlowTiming {
	if len(s.Timings) > 0 {
		return s.Timings[0]
	}
	return TimingActive
}

// 3. 增强：事件上下文 (支持数据修改)
// 比如【精准射击】：强制命中，但伤害-1。这需要修改 AttackEventInfo
type EventContext struct {
	Type     EventType
	SourceID string
	TargetID string
	Card     *Card

	ActionType ActionType

	// 指针类型，允许 Handler 修改正在发生的数值
	DamageVal *int

	BuffID string
	// 姿态/形态相关
	OperatorID          string
	PrevOrientation     CharacterOrientation
	NewOrientation      CharacterOrientation
	PrevForm            string
	NewForm             string
	DiscardedMagicCount int
	// 攻击详情
	AttackInfo *AttackEventInfo

	// 摸牌相关
	DrawCount *int // 摸牌数量，可以被修改
}

type AttackEventInfo struct {
	IsHit            bool   // 是否命中
	IsHitForced      bool   // 是否强制命中 (如: 圣剑)
	IgnoreShield     bool   // 是否无视圣盾 (如: 列风技)
	Element          string // 攻击属性
	CanBeResponded   bool   // 是否可被应战 (如: 暗灭=false)
	ActionType       string // 行动类型 (Attack)
	CounterInitiator string // 原始应战发起者 (空表示主动攻击)
	InterceptTags    map[CombatInterceptTag]bool
	ElementOverride  string // 非-empty 时覆盖卡牌元素的显示（如天枪暗属性）
}

// PromptType 定义用户交互类型
type PromptType string

const (
	PromptChooseCards   PromptType = "choose_cards"   // 选择卡牌
	PromptChooseSkill   PromptType = "choose_skill"   // 选择技能
	PromptConfirm       PromptType = "confirm"        // 确认操作
	PromptChooseExtract PromptType = "choose_extract" // 提炼：多选星石
)

// PresentationKind 定义弹框展示类型（用于前端渲染决策）
type PresentationKind string

const (
	PresentationResponse      PresentationKind = "response"      // 命中/应战/防御（图标按钮）
	PresentationBranchSelect  PresentationKind = "branch_select" // 分支选择（长文案按钮）
	PresentationNumeric       PresentationKind = "numeric"       // 数字选择（X值/治疗点数）
	PresentationCardPicker    PresentationKind = "card_picker"   // 卡牌选择
	PresentationTargetPicker  PresentationKind = "target_picker" // 目标选择
	PresentationSkillChoice   PresentationKind = "skill_choice"  // 技能选择
	PresentationActionHub     PresentationKind = "action_hub"    // 行动枢纽
)

// PromptPresentation 定义弹框展示细节（后端显式声明，前端只渲染）
type PromptPresentation struct {
	Kind        PresentationKind `json:"kind"`                    // 必填：展示类型
	Layout      string           `json:"layout,omitempty"`        // 可选："inline"/"overlay"/"grid"/"heal_allocate"
	NumericBase int              `json:"numeric_base,omitempty"`  // numeric 类型的数字起始值（0 或 1）
	CancelPolicy string          `json:"cancel_policy,omitempty"` // 可选："allow"/"deny"/"implicit"
}

const (
	PromptUIModeActionHub = "action_hub"
)

// PromptOption 定义可选项
type PromptOption struct {
	ID          string `json:"id"`                     // 选项ID (card index / skill id)
	Label       string `json:"label"`                  // 原始显示标签（兼容老客户端）
	ButtonLabel string `json:"button_label,omitempty"` // 按钮短文案（如：发动/放弃/取消/1）
	Hint        string `json:"hint,omitempty"`         // 选项说明（展示在按钮上方）
}

// Prompt 定义用户交互提示
type Prompt struct {
	Type     PromptType `json:"type"`      // 交互类型
	PlayerID string     `json:"player_id"` // 目标玩家
	Message  string     `json:"message"`   // 提示消息
	// 结构化提示语义：用于前端/机器人避免依赖 message 文案做判断。
	ChoiceType string         `json:"choice_type,omitempty"`
	SkillID    string         `json:"skill_id,omitempty"`
	Options    []PromptOption `json:"options"` // 可选项
	// 行动选择场景下"特殊"按钮对应的子选项（如：购买/合成/提炼）
	SpecialOptions []PromptOption `json:"special_options,omitempty"`
	// 可选 UI 渲染模式；action_hub 表示由底部行动半球承载
	UIMode string `json:"ui_mode,omitempty"`
	// 展示协议：后端显式声明展示类型，前端按此渲染（不再猜测）
	Presentation *PromptPresentation `json:"presentation,omitempty"`
	// 额外效果提示（用于前端在响应弹窗中解释"为何可/不可应战、命中后附加效果"等）
	EffectHints []string `json:"effect_hints,omitempty"`

	// 可取消标记：前端据此决定是否显示取消/跳过按钮，无需依赖 phantom option。
	Cancelable bool `json:"cancelable,omitempty"`

	// 选择约束 (CLI只展示，不理解语义)
	Min int `json:"min"` // 最少选择数
	Max int `json:"max"` // 最多选择数

	// 应战专用：发起攻击方ID、可选反弹目标、攻击牌元素（应战须同系或暗灭）
	AttackerID       string   `json:"attacker_id,omitempty"`
	CounterTargetIDs []string `json:"counter_target_ids,omitempty"`
	AttackElement    string   `json:"attack_element,omitempty"` // Earth/Water/Fire/Wind/Thunder/Dark
}

// PlayerInput 定义玩家输入
type PlayerInput struct {
	Type   string   `json:"type"`   // "choose", "confirm", "cancel"
	Values []string `json:"values"` // 选项ID列表
}

// EventType 事件类型
type EventType int

const (
	EventNone EventType = iota
	EventAttack
	EventMagic
	EventDamage
	EventHeal
	EventBuff
	EventCardUsed
	EventBuffRemoved
	EventTurnStart
	EventBeforeDraw
	EventAfterDraw
	EventPhaseEnd
)

// Character 角色模型
type Character struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"` // 称号 (如: 剑圣)
	Title          string            `json:"title"`
	Faction        string            `json:"faction"` // 势力 (如: 圣, 血)
	Description    string            `json:"description"`
	MaxHand        int               `json:"max_hand"` // 基础手牌上限
	Skills         []SkillDefinition `json:"skills"`   // 角色拥有的技能列表
	ExclusiveCards []string          `json:"exclusive_cards"`
}

type DrawOptions struct {
	PreventOverflow bool
	Reason          string
}

// 2. 增强：IGameEngine 接口定义
// 这里必须定义 `Execute` 方法能调用的实际动作
type IGameEngine interface {
	// 资源操作
	ModifyGem(camp string, amount int)
	ModifyCrystal(camp string, amount int)
	// 红宝石可作为蓝水晶替代（仅"水晶消耗"方向）
	GetUsableCrystal(playerID string) int
	CanPayCrystalCost(playerID string, amount int) bool
	ConsumeCrystalCost(playerID string, amount int) bool

	// 玩家操作
	GetPlayers() map[string]*Player
	PlayerOrder() []string
	DrawCards(playerID string, amount int)
	DrawCardsWithOptions(playerID string, amount int, opts DrawOptions)
	DrawRawCards(amount int) ([]Card, bool)
	NotifyCardRevealed(playerID string, cards []Card, actionType DamageType)
	DiscardCard(card *FieldCard) error //丢弃指定牌
	AppendToDiscard(cards []Card)
	Heal(playerID string, amount int)
	CheckHandLimit(playerID string, stayInTurn bool)
	GetAllPlayers() []*Player // 获取所有玩家
	FindFieldEffectBySource(effect EffectType, sourceID string) (*Player, *FieldCard)
	GetMaxHand(player *Player) int
	GetPlayerEnergyCap(player *Player) int
	GetPlayerOrientation(playerID string) CharacterOrientation
	GetPlayerForm(playerID string) string
	RefreshPlayerDerivedState(playerID string)
	ApplySkillGateRule(playerID string, modifierID string, sourceSkillID string, skillIDs []string, lifetime RuleModifierLifetimeType)
	ApplyNextAttackDamageRule(playerID string, modifierID string, sourceSkillID string, bonus int, lifetime RuleModifierLifetimeType)
	ApplyNextAttackInterceptTagRule(playerID string, modifierID string, sourceSkillID string, tag CombatInterceptTag, lifetime RuleModifierLifetimeType)
	IsSkillBlocked(playerID string, skillID string) bool

	InflictDamage(sourceID, targetID string, amount int, damageType DamageType)
	RemoveFieldCard(targetID string, effect EffectType) bool
	RemoveFieldCardBy(targetID string, effect EffectType, sourceID string) bool
	TakeFieldCard(targetID string, fieldIndex int, sourceID string) (Card, error)
	GetCampCups(camp string) int
	AddCampCup(camp Camp) bool
	GetCampMorale(camp string) int
	SetCampMorale(camp string, value int)
	GetCampGems(camp string) int
	GetCampCrystals(camp string) int

	// 日志
	// 提炼流程（由角色包调用）
	StartExtractForPlayer(playerID string) error

	Log(msg string)

	// 行动步骤（桌面展示）
	NotifyActionStep(line string)
	NotifyDamageDealt(sourceID, targetID string, damage int, damageType DamageType)

	// 启动技能确认
	ConfirmStartupSkill(playerID string, skillID string) error
	SkipStartupSkill(playerID string) error

	// 响应技能确认
	ConfirmResponseSkill(playerID string, skillID string) error

	// 弃牌确认
	ConfirmDiscard(playerID string, indices []int) error

	// 响应跳过
	SkipResponse() error

	// UI交互接口
	GetCurrentPrompt() *Prompt

	PushInterrupt(intr *Interrupt)

	AddPendingDamage(pd PendingDamage)
	AddPendingDamageFront(pd PendingDamage)

	// FlowContinuation 接口（角色级流程边界恢复）
	AppendFlowContinuation(cont FlowContinuation)
}

// Context 技能执行上下文
type Context struct {
	Game   IGameEngine
	User   *Player
	Target *Player
	// 【修改点 2】新增多目标支持
	Targets []*Player

	Timing FlowTiming

	// 触发上下文（可选，用于复杂事件）
	EventCtx *EventContext

	// 命令行参数（向后兼容）
	Args []string

	// 技能私有输入（UI / AI 注入）
	Selections map[string]any

	// 系统行为开关（仅 bool）
	Flags map[string]bool

	// 当前PendingInterrupt （仅供Handler读取，不要修改）
	PendingInterrupt *Interrupt
}

// SkillHandler 技能逻辑接口
type SkillHandler interface {
	// 检查额外条件
	CanUse(ctx *Context) bool

	// 执行效果
	Execute(ctx *Context) error
}

// RoomCharacterSelection 房间角色选择模型
type RoomCharacterSelection struct {
	RoomID        string            `json:"room_id"`
	PlayerChoices map[string]string `json:"player_choices"` // PlayerID -> CharacterID
	ReadyStatus   map[string]bool   `json:"ready_status"`   // PlayerID -> IsReady
	Available     []string          `json:"available"`      // 可选角色ID列表
}

// ActionContext 定义行动的约束条件
type ActionContext struct {
	Source      string    `json:"source"`       // 来源，例如 "WindFury", "SwordShadow", "BaseTurn"
	MustElement []Element `json:"must_element"` // 强制要求的属性，空字符串表示无限制
	MustType    string    `json:"must_type"`    // 强制要求的类型，例如 "Attack" 或 "Magic"
}

// PlayerTurnState 玩家回合内状态
type PlayerTurnState struct {
	HasUsedActionSkill           bool            `json:"has_used_action_skill"`            // 本回合是否已在行动开始阶段使用过启动技
	SpecialActionsLockedThisTurn bool            `json:"special_actions_locked_this_turn"` // 本回合锁定购买/合成/提炼（读写见 player_turn_marks.go）
	HasProcessedTurnStart        bool            `json:"has_processed_turn_start"`         // 是否已完成本回合 TurnStart 钩子
	HasActed                     bool            `json:"has_acted"`                        // 是否已执行行动
	ActionPhaseSkippedThisTurn   bool            `json:"action_phase_skipped_this_turn"`   // 本回合行动阶段被跳过（虚弱/五行封印/无法行动等）
	UsedSkillCounts              map[string]int  `json:"used_skill_counts"`                // 技能ID -> 本回合使用次数
	PendingActions               []ActionContext `json:"pending_actions"`                  // 待执行的行动队列
	CurrentExtraAction           string          `json:"current_extra_action"`             // 当前额外行动类型: "Attack", "Magic", ""
	CurrentExtraElement          []Element       `json:"current_extra_element"`            // 当前额外行动元素限制: "Wind", "Fire", etc.
	AttackCount                  int             `json:"attack_count"`                     // 本回合攻击行动次数
	LastActionType               string          `json:"last_action_type"`                 // 记录刚刚结束的行动类型 (Attack/Magic)
	LastActionCard               *Card           `json:"last_action_card,omitempty"`       // 记录刚刚结束行动对应的卡牌快照（供 ActionEnd 统一构建 OnPhaseEnd 上下文）
	SkillFlowState               map[string]int  `json:"skill_flow_state"`                 // 多步技能中间态（如星尘/痛苦链接/冒险者提炼），回合重置时自动清空
}

// NewPlayerTurnState 初始化回合状态
func NewPlayerTurnState() PlayerTurnState {
	return PlayerTurnState{
		HasUsedActionSkill:    false,
		HasProcessedTurnStart: false,
		HasActed:              false,
		UsedSkillCounts:       make(map[string]int),
		SkillFlowState:        make(map[string]int),
		PendingActions:        []ActionContext{}, // 初始化为空队列
		CurrentExtraAction:    "",
		CurrentExtraElement:   nil,
		AttackCount:           0,
	}
}

func AppendExtraAction(p *Player, source string, mustType string, mustElement ...Element) {
	if p == nil {
		return
	}
	ctx := ActionContext{
		Source:   source,
		MustType: mustType,
	}
	if len(mustElement) > 0 {
		ctx.MustElement = append([]Element{}, mustElement...)
	}
	p.TurnState.PendingActions = append(p.TurnState.PendingActions, ctx)
}

func AppendAttackAction(p *Player, source string, mustElement ...Element) {
	AppendExtraAction(p, source, "Attack", mustElement...)
}

func AppendMagicAction(p *Player, source string, mustElement ...Element) {
	AppendExtraAction(p, source, "Magic", mustElement...)
}

// Contains 检查技能标签列表是否包含指定标签
func ContainsSkillTag(tags []SkillTag, tag SkillTag) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

// GetHandlerIDByEffect 根据场上效果枚举获取 Handler ID
func GetHandlerIDByEffect(effect EffectType) string {
	switch effect {
	case EffectShield:
		return "holy_shield"
	case EffectSealWater:
		return "water_seal"
	case EffectSealFire:
		return "fire_seal"
	case EffectSealEarth:
		return "earth_seal"
	case EffectSealWind:
		return "wind_seal"
	case EffectSealThunder:
		return "thunder_seal"

	default:
		return ""
	}
}

// MagicBulletChain 魔弹传递链条
type MagicBulletChain struct {
	CurrentDamage  int      `json:"current_damage"`
	InvolvedIDs    []string `json:"involved_ids"`     // 当前轮参与过的玩家ID
	SourcePlayerID string   `json:"source_player_id"` // 上一个传递者（伤害来源）
	TargetID       string   `json:"target_id"`        // 当前目标（需要响应的玩家）
	Reverse        bool     `json:"reverse"`          // 是否逆向传递（魔弹掌控）
	IsFusion       bool     `json:"is_fusion"`        // 是否由魔弹融合触发（地系/火系牌当魔弹）
	FusionCard     *Card    `json:"fusion_card"`      // 融合使用的原始卡牌
}
