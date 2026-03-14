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
	ExclusiveChar1  string `json:"exclusive_char1,omitempty"`  // 独有技角色1
	ExclusiveChar2  string `json:"exclusive_char2,omitempty"`  // 独有技角色2
	ExclusiveSkill1 string `json:"exclusive_skill1,omitempty"` // 独有技1
	ExclusiveSkill2 string `json:"exclusive_skill2,omitempty"` // 独有技2
}

type BuffType int

const (
	BuffTypeBasic   BuffType = iota // 基础效果 (圣盾, 虚弱, 中毒)
	BuffTypeSpecial                 // 特殊效果 (五系束缚, 挑衅)
	BuffTypeMorph                   // 形态 (英灵形态, 审判形态)
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
	// 精灵射手“祝福”独立牌区：不计入手牌上限，但可按手牌方式打出。
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
	CharaZone []string       `json:"chara_zone"`

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

	CounterInitiator string   `json:"counter_initiator,omitempty"` // 原始应战发起者
	IsDarkCounter    bool     `json:"is_dark_counter,omitempty"`   // 是否为暗灭应战反弹
	ExcludedPlayers  []string `json:"excluded_players,omitempty"`  // 反弹时排除的玩家
	CanBeResponded   bool     `json:"can_be_responded"`            // 是否可被应战
	IsHitForced      bool     `json:"is_hit_forced"`               // 是否强制命中

	Extra   string `json:"extra,omitempty"`
	SkillID string `json:"skill_id,omitempty"`
}

// QueuedAction 队列中的行动（用于额外行动处理）
type QueuedAction struct {
	SourceID                string     `json:"source_id"`                  // 发起者ID
	TargetID                string     `json:"target_id"`                  // 目标ID（攻击/法术的目标）
	TargetIDs               []string   `json:"target_ids,omitempty"`       // 多目标ID (新增支持)
	Type                    ActionType `json:"type"`                       // Attack 或 Magic
	Element                 Element    `json:"element"`                    // 可选：元素限制（如疾风技要求风系）
	Card                    *Card      `json:"card"`                       // 可选：预定义的卡牌（如果已选择）
	CardIndex               int        `json:"card_index"`                 // 卡牌在手牌中的索引
	SourceSkill             string     `json:"source_skill"`               // 来源技能ID（如疾风技、烈风技）
	HasTriggeredCardUsed    bool       `json:"has_triggered_card_used"`    // 是否已触发卡牌使用事件
	HasTriggeredAttackStart bool       `json:"has_triggered_attack_start"` // 是否已触发攻击开始（避免确认响应技能后再次触发）
}

// CombatRequest 战斗请求（用于战斗交互阶段）
type CombatRequest struct {
	AttackerID     string `json:"attacker_id"`      // 攻击者ID
	TargetID       string `json:"target_id"`        // 目标ID
	Card           *Card  `json:"card"`             // 使用的攻击卡牌
	IsForcedHit    bool   `json:"is_forced_hit"`    // 是否强制命中
	CanBeResponded bool   `json:"can_be_responded"` // 是否可被应战
	IsCounter      bool   `json:"is_counter"`       // 是否为应战反弹攻击（命中加水晶）

	// 阴阳师“式神咒束”链路专用上下文
	OnmyojiBindingChecked    bool   `json:"onmyoji_binding_checked,omitempty"`    // 本次战斗是否已检查过代应战
	OnmyojiBindingActorID    string `json:"onmyoji_binding_actor_id,omitempty"`   // 代应战阴阳师ID
	OnmyojiBindingCounterID  string `json:"onmyoji_binding_counter_id,omitempty"` // 预选应战牌ID
	OnmyojiBindingTargetID   string `json:"onmyoji_binding_target_id,omitempty"`  // 预选反弹目标ID
	OnmyojiBindingUseFaction bool   `json:"onmyoji_binding_use_faction,omitempty"`
	// 阴阳师“阴阳转换”交互标记：仅用于控制“先询问是否发动”流程不重复弹出
	OnmyojiYinYangChecked bool `json:"onmyoji_yinyang_checked,omitempty"`
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
	InterruptDiscard              InterruptType = "Discard"
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
	Phase       GamePhase          `json:"phase"`
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
	ActionQueue         []QueuedAction  `json:"action_queue,omitempty"` // 额外行动队列
	CombatStack         []CombatRequest `json:"combat_stack,omitempty"` // 战斗请求栈
	HasPerformedStartup bool            `json:"has_performed_startup"`  // 是否已执行启动技能（限制特殊行动）

	MagicBulletChain *MagicBulletChain `json:"magic_bullet_chain,omitempty"` // 魔弹链条

	// 延迟伤害队列（用于避免嵌套的伤害结算中断）
	PendingDamageQueue []PendingDamage `json:"pending_damage_queue,omitempty"`
	// 延迟后续队列（用于“先结算伤害/中断，再继续技能后续”）。
	DeferredFollowups []DeferredFollowup `json:"deferred_followups,omitempty"`

	// 状态机返回阶段 (用于 PendingDamageResolution 等临时阶段结束后恢复)
	ReturnPhase GamePhase `json:"return_phase,omitempty"`
}

// PendingDamage 代表一个待处理的伤害事件
type PendingDamage struct {
	SourceID                   string     `json:"source_id"`
	TargetID                   string     `json:"target_id"`
	Damage                     int        `json:"damage"`
	DamageType                 string     `json:"damage_type"`
	IgnoreHeal                 bool       `json:"ignore_heal,omitempty"`                  // 本次伤害是否不可被治疗抵御
	CapDrawToHandLimit         bool       `json:"cap_draw_to_hand_limit,omitempty"`       // 本次伤害摸牌是否“最多摸到手牌上限”
	AllowCrimsonFaithHeal      bool       `json:"allow_crimson_faith_heal,omitempty"`     // 红莲骑士[腥红信仰]是否可用治疗抵御本次自伤
	EffectTypeToRemove         EffectType `json:"effect_type_to_remove,omitempty"`        // 伤害结算后需要移除的场上效果 (例如封印)
	Card                       *Card      `json:"card,omitempty"`                         // 关联的卡牌 (用于攻击伤害判定)
	Stage                      int        `json:"stage"`                                  // 处理阶段: 0=Init, 1=HitProcessed, 2=DamageProcessed
	HealResolved               bool       `json:"heal_resolved"`                          // 是否已处理治疗选择
	IsCounter                  bool       `json:"is_counter"`                             // 是否为应战攻击（命中加水晶而非宝石）
	SoulLinkChecked            bool       `json:"soul_link_checked,omitempty"`            // 灵魂链接“承伤前”是否已检查
	FromSoulLink               bool       `json:"from_soul_link,omitempty"`               // 是否为灵魂链接转移产生的法术伤害
	ButterflyPilgrimageChecked bool       `json:"butterfly_pilgrimage_checked,omitempty"` // 本次伤害是否已检查过朝圣
	ButterflyStage5Checked     bool       `json:"butterfly_stage5_checked,omitempty"`     // 本次伤害是否已检查过毒粉/镜花水月
}

// DeferredFollowup 代表一个待执行的技能后续结算。
type DeferredFollowup struct {
	Type      string                 `json:"type"`                 // 后续类型
	UserID    string                 `json:"user_id"`              // 执行者ID
	SkillID   string                 `json:"skill_id,omitempty"`   // 关联技能ID
	TargetIDs []string               `json:"target_ids,omitempty"` // 关联目标ID列表
	Data      map[string]interface{} `json:"data,omitempty"`       // 附加上下文
}

// PendingSkill 等待确认的可选技能
type PendingSkill struct {
	SkillID     string      `json:"skill_id"`
	UserID      string      `json:"user_id"`
	TargetID    string      `json:"target_id"`
	TriggerType TriggerType `json:"trigger_type"`
}

// NewGameState creates a new game state
func NewGameState() *GameState {
	return &GameState{
		Phase:       "",
		Players:     make(map[string]*Player),
		PlayerOrder: []string{}, // Initialize
		TurnOrder:   []string{},
		CurrentTurn: 0,

		Deck:        make([]Card, 0),
		DiscardPile: make([]Card, 0),

		RedMorale:           15,
		BlueMorale:          15,
		RedCups:             0,
		BlueCups:            0,
		RedGems:             0,
		BlueGems:            0,
		RedCrystals:         0,
		BlueCrystals:        0,
		ActionStack:         []Action{},
		PendingInterrupt:    nil, // No interrupt initially
		ActionQueue:         []QueuedAction{},
		CombatStack:         []CombatRequest{},
		HasPerformedStartup: false,
		MagicBulletChain:    nil,
		PendingDamageQueue:  []PendingDamage{}, // 初始化延迟伤害队列
		DeferredFollowups:   []DeferredFollowup{},
	}
}

// MatchExclusive 检查卡牌是否匹配指定角色和独有技
func (c Card) MatchExclusive(characterName, skillTitle string) bool {
	if c.ExclusiveChar1 == characterName && c.ExclusiveSkill1 == skillTitle {
		return true
	}
	if c.ExclusiveChar2 == characterName && c.ExclusiveSkill2 == skillTitle {
		return true
	}
	return false
}

// HasExclusiveCard 检查玩家是否持有指定技能对应的独有牌（优先专属卡区，其次手牌兼容旧逻辑）
func (p *Player) HasExclusiveCard(characterName, skillTitle string) bool {
	if p == nil || characterName == "" || skillTitle == "" {
		return false
	}
	for _, c := range p.ExclusiveCards {
		if c.MatchExclusive(characterName, skillTitle) {
			return true
		}
	}
	for _, c := range p.Hand {
		if c.MatchExclusive(characterName, skillTitle) {
			return true
		}
	}
	return false
}

// ConsumeExclusiveCard 消耗指定技能对应的独有牌。
// 优先从专属卡区消耗；若不存在则回退到手牌（兼容旧测试与历史流程）。
func (p *Player) ConsumeExclusiveCard(characterName, skillTitle string) (Card, bool) {
	if p == nil || characterName == "" || skillTitle == "" {
		return Card{}, false
	}
	for i, c := range p.ExclusiveCards {
		if !c.MatchExclusive(characterName, skillTitle) {
			continue
		}
		p.ExclusiveCards = append(p.ExclusiveCards[:i], p.ExclusiveCards[i+1:]...)
		return c, true
	}
	for i, c := range p.Hand {
		if !c.MatchExclusive(characterName, skillTitle) {
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
func (p *Player) GetFieldEffects(trigger EffectTrigger) []*FieldCard {
	var effects []*FieldCard
	for _, fc := range p.Field {
		if fc.Mode == FieldEffect && fc.Trigger == trigger {
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

// EffectTrigger 定义效果触发时机
type EffectTrigger string

const (
	EffectTriggerOnAttack    EffectTrigger = "OnAttack"    // 攻击时触发
	EffectTriggerOnDamaged   EffectTrigger = "OnDamaged"   // 受到伤害时触发
	EffectTriggerOnTurnStart EffectTrigger = "OnTurnStart" // 回合开始时触发
	EffectTriggerManual      EffectTrigger = "Manual"      // 被技能点名时触发
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
	EffectStealth          EffectType = "Stealth"          // 潜行
	// 魔弓“充能”使用的盖牌效果标识（Mode=Cover）。
	EffectMagicBowCharge EffectType = "MagicBowCharge"
	// 灵符师“妖力”使用的盖牌效果标识（Mode=Cover）。
	EffectSpiritCasterPower EffectType = "SpiritCasterPower"
	// 吟游诗人“永恒乐章”场上效果标识（Mode=Effect）。
	EffectBardEternalMovement EffectType = "BardEternalMovement"
	// 勇者“挑衅”场上效果标识（Mode=Effect）。
	EffectHeroTaunt EffectType = "HeroTaunt"
	// 灵魂术士“灵魂链接”场上效果标识（Mode=Effect）。
	EffectSoulLink EffectType = "SoulLink"
	// 月之女神“暗月”盖牌效果标识（Mode=Cover）。
	EffectMoonDarkMoon EffectType = "MoonDarkMoon"
	// 血之巫女“同生共死”场上效果标识（Mode=Effect）。
	EffectBloodSharedLife EffectType = "BloodSharedLife"
	// 蝶舞者“茧”盖牌效果标识（Mode=Cover）。
	EffectButterflyCocoon EffectType = "ButterflyCocoon"
)

// FieldCard 表示场上放置的卡牌
type FieldCard struct {
	Card     Card          `json:"card"`      // 原始卡牌
	OwnerID  string        `json:"owner_id"`  // 牌当前在哪个玩家面前
	SourceID string        `json:"source_id"` // 谁放的牌
	Mode     FieldCardMode `json:"mode"`      // 效果牌还是盖牌
	Effect   EffectType    `json:"effect"`    // 仅Effect模式下有意义
	Trigger  EffectTrigger `json:"trigger"`   // 触发时机
	Locked   bool          `json:"locked"`    // 是否锁定
	Duration int           `json:"duration"`  // 持续回合数 (-1为永久)
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
