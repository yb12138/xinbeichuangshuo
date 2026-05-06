package model

import "fmt"

// Card 卡牌
type Card struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Type        CardType `json:"type"`    // Attack, Magic
	Element     Element  `json:"element"` // Earth, Water, Fire, Wind, Thunder, Light, Dark
	Damage      int      `json:"damage"`  // 基础伤害值
	Description string   `json:"description"`

	// [新增] 命格和独有技相关字段
	Faction         string `json:"faction,omitempty"`          // 命格 (e.g., "圣", "血", "幻", "技")
	ExclusiveChar1  string `json:"exclusive_char1,omitempty"`  // 独有技角色 ID
	ExclusiveChar2  string `json:"exclusive_char2,omitempty"`  // 独有技角色 ID
	ExclusiveSkill1 string `json:"exclusive_skill1,omitempty"` // 独有技1
	ExclusiveSkill2 string `json:"exclusive_skill2,omitempty"` // 独有技2
}

type BuffType int

const (
	BuffTypeBasic   BuffType = iota // 基础效果 (圣盾, 虚弱, 中毒)
	BuffTypeSpecial                 // 特殊效果 (五系束缚, 挑衅)
	BuffTypeMorph                   // 形态 (英灵形态, 审判形态)
)

type CharacterOrientation string

const (
	OrientationNormal CharacterOrientation = "Normal"
	OrientationTapped CharacterOrientation = "Tapped"
)

// Buff 状态效果
type Buff struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Duration int      `json:"duration"` // 剩余回合数
	Value    int      `json:"value"`    // 数值(如中毒层数)
	LogicID  string   `json:"logic_id"` // 关联的逻辑处理ID
	SourceID string   `json:"source_id"`
	Type     BuffType `json:"type"`
}

// Player 玩家
type Player struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"` // 角色 (前端展示)
	Camp Camp   `json:"camp"`
	Hand []Card `json:"hand"`
	// 仅用于协议视图的“精灵祝福”派生展示字段（真源在 Field 中的 EffectElfBlessing 盖牌）。
	Blessings []Card `json:"blessings,omitempty"`
	// 角色专属技能卡区：不计入手牌，不参与爆牌；用于五系束缚/血蔷薇庭院等专属卡。
	ExclusiveCards []Card       `json:"exclusive_cards,omitempty"`
	Field          []*FieldCard `json:"field"`    // 场上放置的牌
	MaxHand        int          `json:"max_hand"` // 手牌上限

	Heal    int `json:"heal"`
	MaxHeal int `json:"max_heal"`

	Gem     int `json:"gem"`
	Crystal int `json:"crystal"`

	Buffs    []Buff `json:"buffs"`
	IsActive bool   `json:"is_active"` // 是否为当前回合行动者

	Tokens    map[string]int `json:"tokens"`
	Status    map[string]int `json:"status"`
	CharaZone []string       `json:"chara_zone"`

	ActiveRuleModifiers map[string]*RuleModifierInstance `json:"active_rule_modifiers,omitempty"`

	Orientation CharacterOrientation `json:"orientation,omitempty"`
	Form        string               `json:"form,omitempty"`

	Character       *Character      `json:"character,omitempty"`
	TurnState       PlayerTurnState `json:"turn_state"`
	RoomSelectState string          `json:"room_select_state,omitempty"`
}

// GetPlayerDisplayName 获取角色显示名（优先角色名，否则玩家名）
func GetPlayerDisplayName(p *Player) string {
	if p == nil {
		return "?"
	}
	if p.Character != nil && p.Character.Name != "" {
		return p.Character.Name
	}
	return p.Name
}

// Action 玩家行动请求
type Action struct {
	Type      ActionType `json:"type"`
	SourceID  string     `json:"source_id"`
	TargetID  string     `json:"target_id"`
	Card      *Card      `json:"card"` // 使用的卡牌
	CardIdx   int        `json:"card_idx"`
	ExtraArgs []string   `json:"extra_args"`

	CounterInitiator string `json:"counter_initiator,omitempty"` // 原始应战发起者
}

// QueuedAction 队列中的行动（用于额外行动处理）
type QueuedAction struct {
	SourceID                    string     `json:"source_id"`                      // 发起者ID
	TargetID                    string     `json:"target_id"`                      // 目标ID（攻击/法术的目标）
	TargetIDs                   []string   `json:"target_ids,omitempty"`           // 多目标ID (新增支持)
	Type                        ActionType `json:"type"`                           // Attack 或 Magic
	Element                     Element    `json:"element"`                        // 可选：元素限制（如疾风技要求风系）
	Card                        *Card      `json:"card"`                           // 可选：预定义的卡牌（如果已选择）
	CardIndex                   int        `json:"card_index"`                     // 卡牌在手牌中的索引
	SourceSkill                 string     `json:"source_skill"`                   // 来源技能ID（如疾风技、烈风技）
	UsesVirtualCard             bool       `json:"uses_virtual_card,omitempty"`    // 是否为非手牌实体驱动的虚拟牌行动
	HasDispatchedCardUsed       bool       `json:"has_dispatched_card_used"`       // 是否已触发卡牌使用事件
	HasDispatchedAttackDeclared bool       `json:"has_dispatched_attack_declared"` // 是否已触发攻击开始（避免确认响应技能后再次触发）
}

// CombatRequest 战斗请求（用于战斗交互阶段）
type CombatRequest struct {
	AttackerID      string                      `json:"attacker_id"`      // 攻击者ID
	TargetID        string                      `json:"target_id"`        // 目标ID
	Card            *Card                       `json:"card"`             // 使用的攻击卡牌
	IsForcedHit     bool                        `json:"is_forced_hit"`    // 是否强制命中
	IgnoreShield    bool                        `json:"ignore_shield"`    // 是否无视圣盾
	CanBeResponded  bool                        `json:"can_be_responded"` // 是否可被应战
	IsCounter       bool                        `json:"is_counter"`       // 是否为应战反弹攻击（命中加水晶）
	InterceptTags   map[CombatInterceptTag]bool `json:”intercept_tags,omitempty”`
	ElementOverride string                      `json:"element_override,omitempty"` // 非-empty 时覆盖卡牌元素的显示

	// 阴阳师”阴阳转换”交互标记：仅用于控制”先询问是否发动”流程不重复弹出
	OnmyojiYinYangChecked bool `json:”onmyoji_yinyang_checked,omitempty”`
}

// GameState 游戏状态
// Interrupt represents a blocking game state that requires player input
type Interrupt struct {
	Type     InterruptType // Type of interrupt
	PlayerID string        // Player who needs to respond
	SkillIDs []string      // Available skill IDs (for response skills)
	Context  interface{}   // Additional context data
}

// InterruptType defines the type of game interruption
type InterruptType string

const (
	InterruptResponseSkill        InterruptType = "ResponseSkill"
	InterruptStartupSkill         InterruptType = "StartupSkill"
	InterruptChoice               InterruptType = "Choice"
	InterruptMagicMissile         InterruptType = "MagicMissile"
	InterruptGiveCards            InterruptType = "GiveCards"            // 天使祝福等：选牌交给他人
	InterruptMagicBulletFusion    InterruptType = "MagicBulletFusion"    // 魔弹融合：地系/火系牌当魔弹
	InterruptMagicBulletDirection InterruptType = "MagicBulletDirection" // 魔弹掌控：选择传递方向
	InterruptHolySwordDraw        InterruptType = "HolySwordDraw"        // 圣剑：选择摸X弃X
	InterruptSaintHeal            InterruptType = "SaintHeal"            // 圣疗：分配治疗
	InterruptMagicBlast           InterruptType = "MagicBlast"           // 魔爆冲击：选择目标弃牌
)

type GameState struct {
	TurnStage   TurnStage          `json:"turn_stage,omitempty"`
	CombatStage CombatStage        `json:"combat_stage,omitempty"`
	Subflow     Subflow            `json:"subflow,omitempty"`
	Players     map[string]*Player `json:"players"`
	PlayerOrder []string           `json:"player_order"` // Add this if missing
	TurnOrder   []string           `json:"turn_order"`   // Maybe same as PlayerOrder?
	CurrentTurn int                `json:"current_turn"` // Index in TurnOrder

	CurrentPlayer string `json:"current_player"` // ID

	Deck        []Card `json:"deck"`
	DiscardPile []Card `json:"discard_pile"`
	DeckCount   int    `json:"deck_count"` // Derived or actual

	// Global resources
	RedMorale    int `json:"red_morale"`
	BlueMorale   int `json:"blue_morale"`
	RedCups      int `json:"red_cups"` // 圣杯
	BlueCups     int `json:"blue_cups"`
	RedGems      int `json:"red_gems"`
	BlueGems     int `json:"blue_gems"`
	RedCrystals  int `json:"red_crystals"`
	BlueCrystals int `json:"blue_crystals"`

	ActionStack []Action `json:"action_stack"` // 响应栈

	PendingOptionalSkills []PendingSkill `json:"pending_optional_skills"` // 等待确认的可选技能

	// Interrupt system - unified blocking game states
	PendingInterrupt *Interrupt   `json:"pending_interrupt,omitempty"` // Current interrupt (nil if no interrupt)
	InterruptQueue   []*Interrupt `json:"interrupt_queue,omitempty"`   // Wait list for interrupts

	// 11步回合结构新增字段
	ActionQueue []QueuedAction  `json:"action_queue,omitempty"` // 额外行动队列
	CombatStack []CombatRequest `json:"combat_stack,omitempty"` // 战斗请求栈

	MagicBulletChain *MagicBulletChain `json:"magic_bullet_chain,omitempty"` // 魔弹链条

	// 延迟伤害队列（用于避免嵌套的伤害结算中断）
	PendingDamageQueue []PendingDamage `json:"pending_damage_queue,omitempty"`
	// 流程边界恢复点队列（角色级，用于 after_draw/after_damage 等边界触发）。
	FlowContinuations []FlowContinuation `json:”flow_continuations,omitempty”`

	// 状态机返回阶段
	ReturnTurnStage   TurnStage   `json:"return_turn_stage,omitempty"`
	ReturnCombatStage CombatStage `json:"return_combat_stage,omitempty"`
	ReturnSubflow     Subflow     `json:"return_subflow,omitempty"`
	GameOver          bool        `json:"game_over,omitempty"`

	// 回合控制
	NextTurnPlayerOverride string `json:"next_turn_player_override,omitempty"` // 下回合玩家覆盖（用于额外回合等）
}

// DamageType 定义伤害/行动类型枚举文本。
type DamageType string

const (
	AttackDamage DamageType = "Attack"
	MagicAttack  DamageType = "magic"
	// MagicDamage 保留兼容别名，逐步迁移到 MagicAttack。
	MagicDamage DamageType = MagicAttack
)

// PendingDamage 代表一个待处理的伤害事件
type PendingDamage struct {
	SourceID                  string                         `json:"source_id"`
	TargetID                  string                         `json:"target_id"`
	Damage                    int                            `json:"damage"`
	DamageType                DamageType                     `json:"damage_type"`
	OverflowMoraleLossFixed   int                            `json:"overflow_morale_loss_fixed,omitempty"` // 本次伤害摸牌若导致士气下降，则固定为该值
	IgnoreHeal                bool                           `json:"ignore_heal,omitempty"`                // 本次伤害是否不可被治疗抵御
	CapDrawToHandLimit        bool                           `json:"cap_draw_to_hand_limit,omitempty"`     // 本次伤害摸牌是否“最多摸到手牌上限”
	AllowCrimsonFaithHeal     bool                           `json:"allow_crimson_faith_heal,omitempty"`   // 红莲骑士[腥红信仰]是否可用治疗抵御本次自伤
	EffectTypeToRemove        EffectType                     `json:"effect_type_to_remove,omitempty"`      // 伤害结算后需要移除的场上效果 (例如封印)
	SourceSkillID             string                         `json:"source_skill_id,omitempty"`            // 伤害来源技能ID（用于回合内追踪）
	Card                      *Card                          `json:"card,omitempty"`                       // 关联的卡牌 (用于攻击伤害判定)
	IgnoreShield              bool                           `json:"ignore_shield,omitempty"`              // 本次攻击伤害是否无视圣盾
	InterceptTags             map[CombatInterceptTag]bool    `json:"intercept_tags,omitempty"`
	AttackHitFlowDispatched   bool                           `json:"attack_hit_flow_dispatched,omitempty"`   // 本次攻击伤害是否已完成 OnAttackHit 分发
	HealResolved              bool                           `json:"heal_resolved"`                          // 是否已处理治疗选择
	DamageTakenFlowDispatched bool                           `json:"damage_taken_flow_dispatched,omitempty"` // 本次伤害是否已完成 OnDamageTaken 响应分发
	IsCounter                 bool                           `json:"is_counter"`                             // 是否为应战攻击（命中加水晶而非宝石）
	AttackHitResourceType     string                         `json:"attack_hit_resource_type,omitempty"`     // 攻击命中后发放的战绩资源类型(gem/crystal)
	AttackHitResourceGranted  bool                           `json:"attack_hit_resource_granted,omitempty"`  // 是否已成功发放命中战绩资源
	AttackPostHitEffectsDone  bool                           `json:"attack_post_hit_effects_done,omitempty"` // 命中后、承伤前的一次性后效是否已处理
	Checks                    map[PendingDamageCheckKey]bool `json:"checks,omitempty"`                       // 伤害实例运行时检查位（按 key 标记一次性检查/来源属性）
}

type PendingDamageCheckKey string

const (
	PendingDamageCheckHeroRoarMissArmed      PendingDamageCheckKey = "hero_roar_miss_armed"
	PendingDamageCheckFighterChargeMissArmed PendingDamageCheckKey = "fighter_charge_miss_armed"
	PendingDamageCheckAttackMissResolved     PendingDamageCheckKey = "attack_miss_resolved"
	PendingDamageCheckSoulLinkChecked        PendingDamageCheckKey = "soul_link_checked"
	PendingDamageCheckFromSoulLink           PendingDamageCheckKey = "from_soul_link"
	PendingDamageCheckBeforeApplyDefend      PendingDamageCheckKey = "before_apply_defend_checked"
	PendingDamageCheckBeforeApplyResponse    PendingDamageCheckKey = "before_apply_response_checked"
)

func (pd *PendingDamage) HasCheck(key PendingDamageCheckKey) bool {
	return pd != nil && pd.Checks != nil && pd.Checks[key]
}

func (pd *PendingDamage) SetCheck(key PendingDamageCheckKey, enabled bool) {
	if pd == nil || key == "" {
		return
	}
	if enabled {
		if pd.Checks == nil {
			pd.Checks = map[PendingDamageCheckKey]bool{}
		}
		pd.Checks[key] = true
		return
	}
	if pd.Checks == nil {
		return
	}
	delete(pd.Checks, key)
	if len(pd.Checks) == 0 {
		pd.Checks = nil
	}
}

// FlowContinuationKind 流程边界类型（枚举，替代字符串 type）。
type FlowContinuationKind string

const (
	FlowContinuationAfterDraw    FlowContinuationKind = "after_draw"
	FlowContinuationAfterDamage  FlowContinuationKind = "after_damage"
	FlowContinuationAfterDiscard FlowContinuationKind = "after_discard"
)

// FlowContinuation 流程边界恢复点（角色级）。
type FlowContinuation struct {
	Kind      FlowContinuationKind `json:"kind"`
	RoleID    string               `json:"role_id"`
	PlayerID  string               `json:"player_id"`
	SkillID   string               `json:"skill_id,omitempty"`
	TargetIDs []string             `json:"target_ids,omitempty"`
	Data      map[string]any       `json:"data,omitempty"`
}

// PendingSkill 等待确认的可选技能
type PendingSkill struct {
	SkillID  string     `json:"skill_id"`
	UserID   string     `json:"user_id"`
	TargetID string     `json:"target_id"`
	Timing   FlowTiming `json:"timing,omitempty"`
}

// NewGameState creates a new game state
func NewGameState() *GameState {
	return &GameState{
		TurnStage:   "",
		CombatStage: CombatStageNone,
		Subflow:     SubflowNone,
		Players:     make(map[string]*Player),
		PlayerOrder: []string{}, // Initialize
		TurnOrder:   []string{},
		CurrentTurn: 0,

		Deck:        make([]Card, 0),
		DiscardPile: make([]Card, 0),

		RedMorale:          15,
		BlueMorale:         15,
		RedCups:            0,
		BlueCups:           0,
		RedGems:            0,
		BlueGems:           0,
		RedCrystals:        0,
		BlueCrystals:       0,
		ActionStack:        []Action{},
		PendingInterrupt:   nil, // No interrupt initially
		ActionQueue:        []QueuedAction{},
		CombatStack:        []CombatRequest{},
		MagicBulletChain:   nil,
		PendingDamageQueue: []PendingDamage{}, // 初始化延迟伤害队列
	}
}

// MatchExclusive 检查卡牌是否匹配指定角色 ID 与独有技名
func (c Card) MatchExclusive(characterID, skillTitle string) bool {
	if c.ExclusiveChar1 == characterID && c.ExclusiveSkill1 == skillTitle {
		return true
	}
	if c.ExclusiveChar2 == characterID && c.ExclusiveSkill2 == skillTitle {
		return true
	}
	return false
}

// HasExclusiveCard 检查玩家是否持有指定技能对应的独有牌（优先专属卡区，其次手牌兼容旧逻辑）
func (p *Player) HasExclusiveCard(characterID, skillTitle string) bool {
	if p == nil || characterID == "" || skillTitle == "" {
		return false
	}
	for _, c := range p.ExclusiveCards {
		if c.MatchExclusive(characterID, skillTitle) {
			return true
		}
	}
	for _, c := range p.Hand {
		if c.MatchExclusive(characterID, skillTitle) {
			return true
		}
	}
	return false
}

// ConsumeExclusiveCard 消耗指定技能对应的独有牌。
// 优先从专属卡区消耗；若不存在则回退到手牌（兼容旧测试与历史流程）。
func (p *Player) ConsumeExclusiveCard(characterID, skillTitle string) (Card, bool) {
	if p == nil || characterID == "" || skillTitle == "" {
		return Card{}, false
	}
	for i, c := range p.ExclusiveCards {
		if !c.MatchExclusive(characterID, skillTitle) {
			continue
		}
		p.ExclusiveCards = append(p.ExclusiveCards[:i], p.ExclusiveCards[i+1:]...)
		return c, true
	}
	for i, c := range p.Hand {
		if !c.MatchExclusive(characterID, skillTitle) {
			continue
		}
		p.Hand = append(p.Hand[:i], p.Hand[i+1:]...)
		return c, true
	}
	return Card{}, false
}

// RestoreExclusiveCard 将专属卡放回专属卡区（按卡牌ID去重）。
func (p *Player) RestoreExclusiveCard(card Card) {
	if p == nil || card.ID == "" {
		return
	}
	for _, c := range p.ExclusiveCards {
		if c.ID == card.ID {
			return
		}
	}
	p.ExclusiveCards = append(p.ExclusiveCards, card)
}

// HasFieldEffect 检查是否已有指定基础效果
func (p *Player) HasFieldEffect(effect EffectType) bool {
	for _, fc := range p.Field {
		if fc.Mode == FieldEffect && fc.Effect == effect {
			return true
		}
	}
	return false
}

// AddFieldCard 在玩家面前添加场上牌
func (p *Player) AddFieldCard(fc *FieldCard) {
	p.Field = append(p.Field, fc)
}

// RemoveFieldCard 移除指定的场上牌
func (p *Player) RemoveFieldCard(fc *FieldCard) {
	for i, fieldCard := range p.Field {
		if fieldCard == fc {
			p.Field = append(p.Field[:i], p.Field[i+1:]...)
			break
		}
	}
}

// GetFieldEffects 获取指定触发时机的效果牌
func (p *Player) GetFieldEffects(hook FieldHook) []*FieldCard {
	var effects []*FieldCard
	for _, fc := range p.Field {
		if fc.Mode == FieldEffect && fc.Hook == hook {
			effects = append(effects, fc)
		}
	}
	return effects
}

// GetCoverCards 获取所有盖牌
func (p *Player) GetCoverCards() []*FieldCard {
	var covers []*FieldCard
	for _, fc := range p.Field {
		if fc.Mode == FieldCover {
			covers = append(covers, fc)
		}
	}
	return covers
}

// GetCoverCardsByEffect 获取指定效果类型的盖牌。
func (p *Player) GetCoverCardsByEffect(effect EffectType) []*FieldCard {
	if p == nil {
		return nil
	}
	var covers []*FieldCard
	for _, fc := range p.Field {
		if fc == nil || fc.Mode != FieldCover || fc.Effect != effect {
			continue
		}
		covers = append(covers, fc)
	}
	return covers
}

// ConsumeCoverCards 消耗指定数量的盖牌
func (p *Player) ConsumeCoverCards(n int) ([]Card, error) {
	covers := p.GetCoverCards()
	if len(covers) < n {
		return nil, fmt.Errorf("盖牌不足，需要 %d 张，当前只有 %d 张", n, len(covers))
	}

	var consumed []Card
	for i := 0; i < n; i++ {
		consumed = append(consumed, covers[i].Card)
		p.RemoveFieldCard(covers[i])
	}
	return consumed, nil
}

// HasElement 检查玩家是否有指定元素的牌
func (p *Player) HasElement(element Element) bool {
	for _, card := range p.Hand {
		if card.Element == element {
			return true
		}
	}
	return false
}

// FieldCardMode 定义场上牌的模式
type FieldCardMode string

const (
	FieldEffect FieldCardMode = "Effect" // 效果牌：圣盾/中毒/封印等
	FieldCover  FieldCardMode = "Cover"  // 盖牌：作为资源/条件
)

// FieldHook 场上效果牌在何时参与结算（与技能 FlowTiming 无关）
type FieldHook string

const (
	FieldHookOnAttack               FieldHook = "OnAttack"               // 攻击时结算
	FieldHookOnDamaged              FieldHook = "OnDamaged"              // 受到伤害时结算
	FieldHookOnTurnStart            FieldHook = "OnTurnStart"            // 回合开始时结算
	FieldHookOnBeforeAction         FieldHook = "OnBeforeAction"         // 行动阶段开始前结算
	FieldHookOnCardPlayedOrRevealed FieldHook = "OnCardPlayedOrRevealed" // 打出或展示卡牌时结算
	FieldHookManual                 FieldHook = "Manual"                 // 由技能逻辑显式调用
)

// EffectType 定义效果类型
type EffectType string

const (
	EffectShield           EffectType = "Shield"           // 圣盾
	EffectPoison           EffectType = "Poison"           // 中毒
	EffectWeak             EffectType = "Weak"             // 虚弱
	EffectSealFire         EffectType = "SealFire"         // 火之封印
	EffectSealWater        EffectType = "SealWater"        // 水之封印
	EffectSealEarth        EffectType = "SealEarth"        // 地之封印
	EffectSealWind         EffectType = "SealWind"         // 风之封印
	EffectSealThunder      EffectType = "SealThunder"      // 雷之封印
	EffectFiveElementsBind EffectType = "FiveElementsBind" // 五系束缚
	EffectRoseCourtyard    EffectType = "RoseCourtyard"    // 血蔷薇庭院
	EffectPowerBlessing    EffectType = "PowerBlessing"    // 威力赐福
	EffectSwiftBlessing    EffectType = "SwiftBlessing"    // 迅捷赐福
	EffectMercy            EffectType = "Mercy"            // 怜悯
	// 魔弓“充能”使用的盖牌效果标识（Mode=Cover）。
	EffectMagicBowCharge EffectType = "MagicBowCharge"
	// 灵符师“妖力”使用的盖牌效果标识（Mode=Cover）。
	EffectSpiritCasterPower EffectType = "SpiritCasterPower"
	// 吟游诗人“永恒乐章”场上效果标识（Mode=Effect）。
	EffectBardEternalMovement EffectType = "BardEternalMovement"
	// 勇者“挑衅”场上效果标识（Mode=Effect）。
	EffectHeroTaunt EffectType = "HeroTaunt"
	// 剑帝“剑魂”盖牌效果标识（Mode=Cover）。
	EffectSwordSoul EffectType = "SwordSoul"
	// 灵魂术士“灵魂链接”场上效果标识（Mode=Effect）。
	EffectSoulLink EffectType = "SoulLink"
	// 月之女神“暗月”盖牌效果标识（Mode=Cover）。
	EffectMoonDarkMoon EffectType = "MoonDarkMoon"
	// 血之巫女“同生共死”场上效果标识（Mode=Effect）。
	EffectBloodSharedLife EffectType = "BloodSharedLife"
	// 蝶舞者“茧”盖牌效果标识（Mode=Cover）。
	EffectButterflyCocoon EffectType = "ButterflyCocoon"
	// 精灵射手“祝福区”盖牌效果标识（Mode=Cover，可按手牌方式打出）。
	EffectElfBlessing EffectType = "ElfBlessing"
)

// FieldCard 表示场上放置的卡牌
type FieldCard struct {
	Card     Card              `json:"card"`           // 原始卡牌
	OwnerID  string            `json:"owner_id"`       // 牌当前在哪个玩家面前
	SourceID string            `json:"source_id"`      // 谁放的牌
	Mode     FieldCardMode     `json:"mode"`           // 效果牌还是盖牌
	Effect   EffectType        `json:"effect"`         // 仅Effect模式下有意义
	Hook     FieldHook         `json:"field_hook"`     // 结算钩子
	Locked   bool              `json:"locked"`         // 是否锁定
	Duration int               `json:"duration"`       // 持续回合数 (-1为永久)
	Meta     map[string]string `json:"meta,omitempty"` // 状态运行时元数据（如绑定元素）
}

func IsBasicEffect(name string) bool {
	// 确保这里的字符串常量与 RemoveFieldCard 中使用的完全一致
	switch name {
	case "Shield", "圣盾":
		return true
	case "Weak", "虚弱":
		return true
	case "Poison", "中毒":
		return true
	case "SealFire", "火之封印":
		return true
	case "SealWater", "水之封印":
		return true
	case "SealEarth", "地之封印":
		return true
	case "SealWind", "风之封印":
		return true
	case "SealThunder", "雷之封印":
		return true
	case "PowerBlessing", "威力赐福":
		return true
	case "SwiftBlessing", "迅捷赐福":
		return true
	}
	return false
}
