# 《星杯传说》核心引擎静态类配置字典 (Static Classes Config)

**文档说明**：本文档定义了游戏引擎中所有的“静态数据结构（Static Classes）”。这些结构体在游戏运行期间是**全局只读（Read-Only）**的，通常由 JSON/CSV 配置表在服务器启动时一次性加载入内存。它们构成了游戏规则的物理边界。

---

## 1. 全局系统常量配置 (Global Game Constants)
定义了游戏最基础的物理上限。这些常量会在引擎校验动作合法性时被频繁调用。

```go
type GlobalConstants struct {
    DefaultMaxHandSize int // 默认手牌上限 (6)
    DefaultMaxEnergy   int // 默认能量上限 (3)
    DefaultMaxHeal     int // 默认治疗上限 (2)
    
    MaxStarStones      int // 战绩区星石上限 (5)
    TargetStarCups     int // 宣告胜利所需星杯数 (5)
    
    StandardMaxMorale  int // 标准模式士气上限/初始值 (15)
    EpicMaxMorale      int // 8人模式士气上限/初始值 (18)
    
    DeckTotalCards     int // 基础牌库总牌数 (150)
}
```

---

## 2. 卡牌模板配置 (Card Template)
定义了基础牌库（150张）中每一种卡牌的静态属性。游戏中玩家抽到的每一张实体牌，都持有一个指向此模板的引用（TemplateID）。

```go
type CardTemplate struct {
    TemplateID  string      // 卡牌唯一模板ID（如 "card_attack_fire_1"）
    Name        string      // 卡牌名称（如 "火焰斩"）
    
    CardType    CardType    // 卡牌大类（如 Attack, Spell, Exclusive）
    Element     ElementType // 元素属性（如 Fire, Water）
    Destiny     DestinyType // 命格属性（基础牌库通常为 None，角色专属卡有命格）
    
    BaseDamage  int         // 基础伤害值（所有基础攻击牌均为 2，魔弹为 2）
    IsCounter   bool        // 是否可用于应战（例如暗灭可应战其他系，但自身不可被应战）
    
    Description string      // 卡牌描述文本（用于前端展示）
	
    CharacterSkillMap map[string]string // 角色-技能映射表 Key: 角色ID (CharacterID), Value: 技能ID (SkillID)
}
```

---

## 3. 角色基础配置 (Character Data Template)
定义了所有登场角色的静态白板面板。对应角色Excel配置表中的基础数据。

```go
type CharacterTemplate struct {
    CharacterID string      // 角色唯一ID（如 "role_saintess"）
    Name        string      // 角色名称（如 "圣女"）
    Title       string      // 称号/前缀（可选）
    
    Expansion   string      // 所属扩展包（如 "基础 3.0", "一扩", "特典"）
    Destiny     DestinyType // 角色的固有命格（如 Sheng/圣, Xue/血）
    
    // 面板重载 (如果角色拥有与全局默认不同的上限，在此配置)
    OverrideMaxHandSize *int // 初始手牌上限重载 (如无则使用默认 6)
    OverrideMaxEnergy   *int // 初始能量上限重载 (如贤者为 4)
    OverrideMaxHeal     *int // 初始治疗上限重载 (如圣使为 6，染污者为 0)
    
    // 初始专属指示物 (游戏开始 PhaseGameInit 时发放)
    InitialTokens map[TokenType]int // 如英灵人形配置：{ WarRune: 3 }
    
    // 初始场上标记/专属卡 (游戏开始时即拥有，如异端裁决所、圣煌辉光炮)
    InitialFieldMarks map[CardFieldMark]int // 如圣庭检察士：{ InquisitionCourt: 1 }，圣弓：{ HolyGloryCannon: 1 }
    
    // 技能树挂载 (指向 Skill Definition ID)
    SkillIDs    []string    // 该角色拥有的所有技能ID列表
}
```

---

## 4. 技能定义配置 (Skill Definition Config)
核心解构配置。将复杂的技能文本解构为引擎可读的校验逻辑和执行约束。

### 4.1 技能主干配置 (Main Struct)
```go
type SkillDefinition struct {
    SkillID     string        // 技能唯一ID（如 "skill_saintess_heal"）
    Name        string        // 技能名称（如 "治愈之光"）
    
    Category    SkillCategory // 技能来源类别（普通/独有/专属/大招）
    Type        SkillType     // 发动机制类别（被动/启动/法术/响应）
    Timing      TriggerTiming // 事件总线监听的触发钩子（如 TimingActive）
    
    TargetRule  *TargetRuleConfig // 目标选择规则（控制前端选取头像）
    Condition   *ConditionConfig  // 状态前置条件（控制按钮是否点亮）
    Mandatory   *SkillMandatoryConfig // 可选：强制发动配置（命中后在当前行动阶段仅允许执行本技能）
    ActionTransform *ActionTransformConfig // 可选：行动改写（在执行前把当前行动重解释到另一条流水线）
    ResponseGroup *ResponseGroupConfig // 可选：响应分组（同一触发窗内二选一/互斥）
    ReplacesSkillIDs []string // 可选：当本技能被选中时，替换并取消这些候选技能
    Cost        *CostConfig       // 费用消耗配置（控制系统扣除资源）
    
    Description string        // 技能文本描述（用于前端展示）
    Effects []EffectNode
}
```

### 4.1.1 技能强制发动配置 (Skill Mandatory Config, 可选)
用于统一表达“命中条件后，本行动阶段只能执行该技能”的规则（避免拆分成两个技能ID）。
```go
type SkillMandatoryConfig struct {
    MatchTiming TriggerTiming // 在哪个钩子上判定是否进入“强制发动”锁（通常为 TimingStartup）
    ConditionExpression string // 命中强制锁的附加表达式（基于 Event/State/Player）
    LockMode    SkillMandatoryLockMode // 锁定模式
}

type SkillMandatoryLockMode int
const (
    SkillMandatoryLockNone                   SkillMandatoryLockMode = 0 // 不锁定（默认）
    SkillMandatoryLockActionPhaseToSelfSkill SkillMandatoryLockMode = 1 // 本行动阶段仅允许执行该 SkillID
)
```

> 运行时约定：命中 `Mandatory` 后，当前行动阶段只允许提交 `ActionType=Skill` 且 `SkillID=锁定技能ID`；其余行动（含 `Pass`）一律非法。

### 4.1.2 响应分组配置 (Response Group Config, 可选)
用于引擎层硬约束“同一触发窗口的响应技能二选一”。
```go
type ResponseGroupConfig struct {
    GroupID string // 响应分组ID（同一触发窗口内按 GroupID 聚合候选技能）
    Mode model.ResponseGroupMode // 分组仲裁模式
    OptionOrder int // 前端展示排序（小者靠前）
}

type ResponseGroupMode int
const (
    ResponseGroupChooseOne ResponseGroupMode = 0 // 组内最多选择一个响应技能
)
```

> 运行时约定：同组 `Mode=ResponseGroupChooseOne` 时，客户端最多提交一个技能；若被选技能配置了 `ReplacesSkillIDs`，则被替换技能在同窗口直接取消执行。
>
> 事件处理伪代码参考：`docs/data_model.md` 的 **4.5.7 事件处理伪代码（候选收集 -> 分组 -> 选择 -> 替换 -> 入执行队列）**。

### 4.1.3 规则修饰器模板配置 (Rule Modifier Template Config, 可选)
用于统一承载“属性修饰/治疗策略/技能门禁/卡牌来源投影/指示物策略”等规则能力。技能通过 `EffectApplyRuleModifier` 引用模板并施加实例。
```go
type RuleModifierTemplate struct {
    ModifierID string               // 模板ID（全局唯一）
    Domain model.RuleModifierDomain // 规则域（Attribute / HealPolicy / SkillGate / CardSource / TokenPolicy / HealResistPolicy / MoralePolicy）
    Priority int                    // 规则优先级（大者先应用）
    ConditionExpression string       // 命中条件；为空表示常驻生效
    StackPolicy model.RuleModifierStackPolicy // 叠加策略

    AttrPayload *RuleAttrPayload
    HealPolicyPayload *RuleHealPolicyPayload
    SkillGatePayload *RuleSkillGatePayload
    CardSourcePayload *RuleCardSourcePayload
    TokenPolicyPayload *RuleTokenPolicyPayload
    HealResistPolicyPayload *RuleHealResistPolicyPayload
    MoralePolicyPayload *RuleMoralePolicyPayload
}

type PlayerAttributeType int
const (
    PlayerAttributeMaxHand   PlayerAttributeType = 0 // 手牌上限
    PlayerAttributeMaxEnergy PlayerAttributeType = 1 // 能量上限
    PlayerAttributeMaxHeal   PlayerAttributeType = 2 // 治疗上限
)

type AttributeModifyOpType int
const (
    AttributeModifyAdd AttributeModifyOpType = 0 // 在当前值基础上 +Value
    AttributeModifySet AttributeModifyOpType = 1 // 直接设为 Value
)

type RuleAttrValueSourceMode int
const (
    RuleAttrValueFromFixed RuleAttrValueSourceMode = 0 // 使用 RuleAttrPayload.Value 固定值（默认）
    RuleAttrValueFromExpression RuleAttrValueSourceMode = 1 // 使用 RuleAttrPayload.ValueExpression 动态表达式
    RuleAttrValueFromTokenLinear RuleAttrValueSourceMode = 2 // 使用 RuleAttrPayload.TokenLink（按 Token 数量线性计算）
)

type RuleAttrTokenOwnerScope int
const (
    RuleAttrTokenOwnerTarget RuleAttrTokenOwnerScope = 0 // 按规则作用目标（Target）的 Token 计算
    RuleAttrTokenOwnerSource RuleAttrTokenOwnerScope = 1 // 按规则施加来源（Source）的 Token 计算
)

type HealApplyMode int
const (
    HealApplyRespectMax HealApplyMode = 0 // 按目标当前有效治疗上限结算（默认）
    HealApplyIgnoreMax  HealApplyMode = 1 // 忽略目标当前有效治疗上限
)

type RuleModifierDomain int
const (
    RuleModifierDomainAttribute  RuleModifierDomain = 0 // 属性域（如 MaxHeal+1）
    RuleModifierDomainHealPolicy RuleModifierDomain = 1 // 治疗策略域（如 IgnoreMax+AbsoluteMax）
    RuleModifierDomainSkillGate  RuleModifierDomain = 2 // 技能门禁域（如禁用指定技能）
    RuleModifierDomainCardSource RuleModifierDomain = 3 // 卡牌来源域（如指定场标可视为手牌）
    RuleModifierDomainTokenPolicy RuleModifierDomain = 4 // 指示物策略域（如 IgnoreMax+AbsoluteMax）
    RuleModifierDomainHealResistPolicy RuleModifierDomain = 5 // 治疗抵伤策略域（如每次受伤窗口最多用1点治疗抵伤）
    RuleModifierDomainMoralePolicy RuleModifierDomain = 6 // 士气策略域（如阵营士气下限/上限）
)

type MoralePolicyApplyScope int
const (
    MoralePolicyApplySourceTeam MoralePolicyApplyScope = 0 // 作用于规则施加者所属阵营
    MoralePolicyApplyEnemyTeam  MoralePolicyApplyScope = 1 // 作用于规则施加者敌方阵营
    MoralePolicyApplyTargetTeam MoralePolicyApplyScope = 2 // 作用于规则目标所属阵营
    MoralePolicyApplyAllTeams   MoralePolicyApplyScope = 3 // 同时作用于双方阵营
)

type RuleModifierStackPolicy int
const (
    RuleModifierStackAppend RuleModifierStackPolicy = 0 // 直接叠加
    RuleModifierStackRefreshByModifierID RuleModifierStackPolicy = 1 // 相同 ModifierID 刷新时长
    RuleModifierStackReplaceByDomainPriority RuleModifierStackPolicy = 2 // 同域按优先级替换
)

type RuleModifierLifetimeType int
const (
    RuleLifeThisEffectChain RuleModifierLifetimeType = 0 // 当前技能执行链结束失效
    RuleLifeUntilTurnEnd RuleModifierLifetimeType = 1 // 到当前回合结束
    RuleLifeUntilSourceNextTurnStart RuleModifierLifetimeType = 2 // 到施加者下回合开始
    RuleLifeUntilSourceNextTurnEnd RuleModifierLifetimeType = 3 // 到施加者下回合结束
    RuleLifePermanent RuleModifierLifetimeType = 4 // 常驻，需显式移除
    RuleLifeUntilCombatEnd RuleModifierLifetimeType = 5 // 到当前战斗结算链结束（CombatDraw 收尾后）
)

type SkillGateMode int
const (
    SkillGateDisallowList SkillGateMode = 0 // 禁用 SkillIDs 列表中的技能
)

type CardSourceProjectionMode int
const (
    CardSourceProjectionAsHand CardSourceProjectionMode = 0 // 将目标场标牌投影为手牌候选来源
)

type TokenApplyMode int
const (
    TokenApplyRespectMax TokenApplyMode = 0 // 按目标当前有效指示物上限结算（默认）
    TokenApplyIgnoreMax  TokenApplyMode = 1 // 忽略目标当前有效指示物上限
)

type RuleAttrPayload struct {
    AttrType model.PlayerAttributeType
    Operation model.AttributeModifyOpType
    ValueSourceMode model.RuleAttrValueSourceMode // 取值来源模式；缺省按 RuleAttrValueFromFixed 处理
    Value int                                     // ValueSourceMode=RuleAttrValueFromFixed 时生效
    ValueExpression string                        // ValueSourceMode=RuleAttrValueFromExpression 时生效
    TokenLink *RuleAttrTokenLinkPayload           // ValueSourceMode=RuleAttrValueFromTokenLinear 时生效
}

type RuleAttrTokenLinkPayload struct {
    OwnerScope model.RuleAttrTokenOwnerScope // 按谁的 Token 计算（Target / Source）
    TokenType model.TokenType
    Coefficient int // 线性系数（可正可负）
    Offset int      // 线性偏移（可正可负）
    MinValue *int   // 可选：下限钳制
    MaxValue *int   // 可选：上限钳制
}
// TokenLink 解析公式：resolved = Coefficient * tokenCount + Offset；若配置 MinValue/MaxValue，再做上下限钳制。

type RuleHealPolicyPayload struct {
    ApplyMode model.HealApplyMode
    AbsoluteMax *int
}

type RuleSkillGatePayload struct {
    Mode model.SkillGateMode
    SkillIDs []string
}

type RuleCardSourcePayload struct {
    ProjectionMode model.CardSourceProjectionMode
    FieldMarks []model.CardFieldMark
}

type RuleTokenPolicyPayload struct {
    TokenType model.TokenType
    ApplyMode model.TokenApplyMode
    AbsoluteMax *int
}

type RuleHealResistPolicyPayload struct {
    PerDamageWindowHealCap *int // 每次受伤窗口可使用的治疗抵伤上限（nil 表示不限制）
}

type RuleMoralePolicyPayload struct {
    ApplyScope model.MoralePolicyApplyScope
    MinMorale *int // 可选：士气下限（nil 表示不限制下限）
    MaxMorale *int // 可选：士气上限（nil 表示不限制上限）
}
```

### 4.2 目标规则配置 (Target Rule Config)
用于指导前端UI渲染目标选择光圈，并为后端提供合法性校验参数。
```go
type TargetRuleConfig struct {
    SelectType  TargetSelectType // 目标身份限制（如 Enemy, Teammate, Any）
    MinCount    int              // 最少必须选择的数量
    MaxCount    int              // 最多允许选择的数量
    
    // 复杂目标条件过滤（可选）
    RequireStatus   *StatusEffect   // 目标身上必须存在某状态
    RequireFieldMark *CardFieldMark // 目标面前必须存在某场上标记（如：必须拥有"风穴"）
    RequireHeal     *int            // 目标拥有的治疗量必须 >= 该值
    RequireHand     *int            // 目标手牌数要求（如：手牌<4的对手）
    RequireHandMax  *int            // 目标手牌数必须 <= 该值

    // 可选：多目标点数分配（如“任意分配3点治疗给1~3名角色”）
    RequireTargetAllocations bool    // 是否要求客户端提交 TargetAllocations
    AllocationTotal          *int    // 分配总值（如 3）
    MinAllocationPerTarget   *int    // 单目标最小分配值（如 1）
    MaxAllocationPerTarget   *int    // 单目标最大分配值（如 3）
    AllowedActionRefs []ActionType   // 可选：若技能需要玩家在 Attack/Magic 中二选一，则在这里声明允许值

    // 可选：命名数值输入约束（如同一技能同时输入 X/Y）
    NamedValueConstraints []NamedValueConstraint
}

type NamedValueConstraint struct {
    Key string            // 变量名（如 "X", "Y"）
    Required bool         // 是否必须提交该变量
    MinExpression string  // 下界表达式（空表示不限制）
    MaxExpression string  // 上界表达式（空表示不限制）
}
```

### 4.3 状态前置条件 (Condition Config)
用于判定该技能当前是否允许发动（不涉及资源扣除，纯状态校验）。
```go
type ConditionConfig struct {
    PhaseLimit         []GamePhase           // 仅在特定阶段合法（如 启动技仅在 ActionStart）
    RequireOrientation *CharacterOrientation // 姿态要求（如 仅限"横置"状态下发动）
    RequireForm        *string               // 形态要求（如 "shadow" 仅暗影形态下，由技能配置定义）
    
    MaxHandLimit       *int                  // 手牌数必须 <= 该值（如 苍炎魔女【魔女之怒】< 4）
    MinHandLimit       *int                  // 手牌数必须 >= 该值
    
    RequireFieldMark   *CardFieldMark        // 场上必须拥有特定盖牌（如 必须拥有"剑魂"）
    RequireToken       *TokenType            // 必须拥有特定指示物（如 必须拥有永恒乐章可解析为 FieldMark）
    IsTurnLimited      bool                  // 是否为【回合限定】（单回合内仅能触发一次）
    
    CustomExpression   string                // 自定义表达式（基于 Player/CombatContext 等的复杂条件）
    RequireSourceType *EventSourceType       // 来源限制：若配置为 SourcePlayer，则系统自然结算不触发
    RequireOperator   *OperatorIdentity      // 主体限制：触发事件的那个玩家，必须和我是什么关系？
    FieldMarkChangeFilter *FieldMarkChangeFilterConfig // 当 Timing=TimingOnFieldMarkChanged 时，筛选哪些变更可触发
    CustomEventFilter string                 // 复杂事件表达式留空备用（FieldMarkChangeFilter 无法表达时）
}

// 场标变更过滤器（配合 TimingOnFieldMarkChanged）
type FieldMarkChangeFilterConfig struct {
    AcceptBehaviors        []string        // 接受的行为：["Placed", "Removed"]
    AcceptTypesWhenPlaced  []StatusEffect  // Placed 时接受的类型，空=nil 表示任意
    AcceptTypesWhenRemoved []StatusEffect  // Removed 时接受的类型，空=nil 表示任意
}
```

### 4.4 行动改写配置 (Action Transform Config)
用于表达“视为某行动/改写到另一条结算流水线”的能力，避免把具体技能逻辑耦合到 `EffectType` 枚举中。
```go
type ActionTransformConfig struct {
    Hook     TriggerTiming       // 建议固定为 TimingBeforeActionExecute（执行前改写）
    Optional bool                // true=可选发动；false=满足条件后强制改写
    Priority int                 // 多个改写同时命中时的优先级（大者优先）
    CancelCurrentAction bool     // 是否取消当前行动默认结算（如替代 Buy 的原生流程）

    Match   ActionTransformMatch // 命中条件
    Rewrite *ActionRewriteConfig // 改写结果；nil 表示仅取消当前行动并继续执行 Skill.Effects
}

type ActionTransformMatch struct {
    RequireActionType         *ActionType   // 当前待执行行动类型要求（如 Magic）
    RequirePlayedCardTypes    []CardType    // 打出的牌类型要求（可空）
    RequirePlayedCardElements []ElementType // 打出的牌系别要求（可空）
    ExcludeTemplateIDs        []string      // 排除模板（如“除暗灭外”）
}

type ActionRewriteConfig struct {
    FlowRef             ActionFlowType         // 目标流水线（如 ActionFlowMagicBulletChain）
    ActionTypeRef       *ActionType            // 可选：同时改写 ActionType（不需要时 nil）
    ExecuteImmediately  bool                   // true=立即按改写后的行动进入结算（不是“+1次行动”）
    TreatAsActiveAttack bool                   // 当改写为 Attack 时，是否标记为主动攻击
    ElementPickMode     RewriteElementPickMode // 改写后行动的元素来源策略
    FixedElementRef     *ElementType           // 当 ElementPickMode=RewriteElementFixed 时使用
}

type RewriteElementPickMode int
const (
    RewriteElementNone          RewriteElementPickMode = 0 // 不改写元素，沿用原行动/原牌元素
    RewriteElementFixed         RewriteElementPickMode = 1 // 使用 FixedElementRef
    RewriteElementFromActionRef RewriteElementPickMode = 2 // 使用 ClientActionRequest.ElementRef
)

type ActionFlowType int
const (
    ActionFlowNormalCombat    ActionFlowType = 0 // 常规攻击/法术战斗流水线
    ActionFlowMagicBulletChain ActionFlowType = 1 // 魔弹传递链路流水线
)
```

### 4.5 费用消耗配置 (Cost Config)
定义了成功宣告技能后，系统需要强制扣除的资源与卡牌。包含对“同系/同命格”等复合弃牌条件的支持。
```go
type CostConfig struct {
    // 0. 发动载体（区分“仅按钮发动”与“必须打出映射卡牌”）
    CardPlayCostType model.CardPlayCostType

    // 1. 星石消耗
    Stones   []StoneCostConfig    // 能量/星石消耗列表
    
    // 2. 专属指示物消耗
    Tokens   []TokenCostConfig    // 专属指示物消耗列表
    
    // 3. 弃牌要求 (依赖卡牌过滤器)
    Discards []DiscardRequirement // 弃牌要求组合（如 弃2张同系牌）
    
    // 4. 其他属性扣除
    HealCost int                  // 需要移除的治疗点数
    HPCost   int                  // 需要对自己造成的法术伤害值（卖血技）
}

// —— 子结构体定义 ——

// 发动载体类型
type CardPlayCostType int
const (
    CardPlayNotRequired CardPlayCostType = 0 // 发动不需要打出卡牌（如被动触发、纯按钮技能）
    CardPlayRequired    CardPlayCostType = 1 // 发动必须打出与技能映射的卡牌（如基础法术牌、独有牌）
)

// 星石消耗条目
type StoneCostConfig struct {
    Type   StarStoneType // Gem, Crystal, 或 Any(通配符能量)
    Amount int           // 数量
}

// 指示物消耗条目
type TokenCostConfig struct {
    Type   TokenType     // 消耗哪种指示物
    Amount int           // 数量
}

// 弃牌要求条目
type DiscardRequirement struct {
    Count  int              // 必须弃除的数量
    Filter CardFilterConfig // 这些牌必须满足的条件
}

// —— 卡牌过滤器与聚合维度 ——

// 卡牌聚合匹配维度 (用于定义多张卡牌间的关系)
type MatchAttribute int

const (
    MatchNone          MatchAttribute = iota // 无需一致（随便弃几张）
    MatchElement                             // 必须同系（如：弃2张同系牌）
    MatchDestiny                             // 必须同命格（如：弃2张同命格牌）
    MatchType                                // 必须同类型（如：弃2张攻击牌）
    MatchOppositeElement                     // 必须异系（如：贤者弃X张异系牌）
)

// 卡牌过滤器 (支持复合判断)
type CardFilterConfig struct {
    // 1. 个体校验限制（单张牌必须满足的条件）
    ReqCardType *CardType    // 要求卡牌类型（如 必须是法术牌）
    ReqElement  *ElementType // 要求元素属性（如 必须是火系牌）
    ReqDestiny  *DestinyType // 要求命格属性
    
    // 2. 聚合校验限制（多张牌之间必须满足的条件）
    SameAttribute MatchAttribute // 必须在哪个维度上保持一致（如 MatchElement）
}

// ==========================================
// 4.6 效果执行配置 (Effect Nodes)
// ==========================================

// 单个执行节点 (将挂载到 SkillDefinition 中的 Effects 数组里)
type EffectNode struct {
EffectType model.EffectType       // 要执行的具体微观逻辑（如：造成伤害、加血、摸牌）
Target     model.EffectTargetType // 这个动作作用于谁（如：自己、选中的人、全场）
Value      int                    // 动作的数值（如：伤害值、摸牌数）

// 可选：关联的具体实体
StatusRef  *model.StatusEffect    // 用于 EffectPlaceStatus 时，放置什么状态？
TokenRef   *model.TokenType       // 用于 EffectAddToken 时，增加什么指示物？
ActionRef  *model.ActionType      // 用于 EffectAddAction 时，限定追加行动类型（Attack/Magic）
ElementRef *model.ElementType     // 可选：系别限制（如风怒追击的风系攻击行动）；当 EffectType=EffectSetCurrentCombatElement 时表示改写后的战斗系别；当 EffectType=EffectRemoveFieldMark 时可按系别过滤移除
StoneRef   *model.StarStoneType   // 用于 EffectAddTeamStone / EffectAddEnergyStone / EffectConvertTeamStone / EffectConvertEnergyStone 时，指定源颜色或星石类型（Gem/Crystal/Any）
StoneToRef *model.StarStoneType   // 用于 EffectConvertTeamStone / EffectConvertEnergyStone 时，指定转换目标星石类型
FromTargetRef *model.EffectTargetType // 用于 EffectTransferTeamStone / EffectTransferCard / EffectTransferFieldMark 时，指定来源目标（队伍/玩家）
FieldMarkRef *model.CardFieldMark // 用于 EffectRemoveFieldMark / EffectPlaceDeckTopAsFieldMark / EffectPlacePlayedCardAsFieldMark / EffectPlaceHandCardAsFieldMark / EffectTransferFieldMark 时，指定场标类型（如 Blessing）
VisibilityRef *model.CardVisibilityType // 用于 EffectPlacePlayedCardAsFieldMark / EffectPlaceHandCardAsFieldMark 时，指定放置后明暗（VisibilityPublic / VisibilityHidden）
OrientationRef *model.CharacterOrientation // 用于 EffectSetOrientation 时，设置姿态（Normal/Tapped）
FormRef    *string                // 用于 EffectSetForm 时，设置命名形态（nil 表示清空形态）
BranchRef  *PerTargetBranchConfig // 用于 EffectPerTargetBranch 时，描述逐目标的响应分支规则
RuleModifierRef *string           // 用于 EffectApplyRuleModifier：引用 RuleModifierTemplate.ModifierID
RuleLifetimeRef *model.RuleModifierLifetimeType // 用于 EffectApplyRuleModifier：实例持续时长
RuleRemoveRef *model.RuleModifierRemoveQuery // 用于 EffectRemoveRuleModifier：移除筛选条件
}

// —— 微观效果类型枚举 (Effect Type) ——
type EffectType int
const (
    EffectNone                 EffectType = 0
    EffectDamage               EffectType = 1  // 造成法术伤害
    EffectAttackDamage         EffectType = 2  // 造成攻击伤害 (附加物理命中判定)
    EffectHeal                 EffectType = 3  // 调整治疗（正数=增加；负数=移除；最小不低于0）
    EffectDrawCard             EffectType = 4  // 摸牌
    EffectDiscard              EffectType = 5  // 强制目标弃牌
    EffectAddToken             EffectType = 6  // 调整专属指示物（Value>0 增加；Value<0 移除；移除不足按可移除量结算且不阻断）
    EffectAddAction            EffectType = 7  // 增加额外行动次数 (如: +1 攻击行动)
    EffectPlaceStatus          EffectType = 8  // 放置基础效果 (如: 挑衅、五系封印)
    EffectRemoveStatus         EffectType = 9  // 移除基础效果 (如: 天使解除状态)
    EffectRemoveStatusToHand   EffectType = 10 // 将场上基础效果牌收入手牌 (如封印破碎)
    EffectAttackDamageModifier EffectType = 11 // 攻击伤害增减修饰 (如狂化、撕裂)
    EffectChangeMorale         EffectType = 12 // 直接扣除/增加已生效士气（立即生效）
    EffectTransferCard         EffectType = 13 // 转移卡牌（FromTargetRef -> Target；FromTargetRef 为空时兼容旧语义：Target -> Self）
    EffectAdjustHand           EffectType = 14 // 将手牌调整为X张 (如圣煌辉光炮)
    EffectSwapMorale           EffectType = 15 // 将一方士气调整与另一方相同
    EffectApplyCombatTag       EffectType = 16 // 当执行这个 Effect 时，引擎会把 Value 字段转换为 CombatInterceptTag，并塞入当前 CombatContext 的 InterceptTags 池子里。
    EffectReducePendingMoraleLoss EffectType = 17 // 将当前事件窗口中的 PendingMoraleLoss 减少 Value（最小为 0）
    EffectCancelHit            EffectType = 18 // 取消命中/抵挡：替代 SystemOp_BlockCurrentAttackOrMagicBullet
    EffectSkipPhase            EffectType = 19 // 跳过阶段：替代 SystemOp_SkipActionPhase
    EffectAddTeamStone         EffectType = 20 // 调整目标阵营战绩区星石（StoneRef 指定 Gem/Crystal/Any；Value>0 增加，Value<0 移除；移除不足时按可移除量结算且不阻断）
    EffectPerTargetBranch      EffectType = 21 // 对选中目标逐个发起响应判定，并按成功/失败分支结算
    EffectAddEnergyStone       EffectType = 22 // 调整目标玩家能量区星石（StoneRef 指定 Gem/Crystal/Any；Value>0 增加，Value<0 移除；移除不足按可移除量结算且不阻断）
    EffectSetOrientation       EffectType = 23 // 设置目标姿态（OrientationRef 指定 Normal/Tapped）
    EffectSetForm              EffectType = 24 // 设置/清空目标命名形态（FormRef 为空表示清空）
    EffectSetHandLimitFixed    EffectType = 25 // 设置目标手牌上限固定值（Value；>0 生效）
    EffectTransferTeamStone    EffectType = 26 // 在队伍战绩区间转移星石（FromTargetRef=源队伍，Target=目标队伍，StoneRef=类型，Value=数量）
    EffectConvertTeamStone     EffectType = 27 // 转换队伍战绩区星石颜色（Target=队伍，StoneRef=源颜色，StoneToRef=目标颜色，Value<=0 表示全部）
    EffectRedirectCurrentExtractOutput EffectType = 28 // 将当前 Extract 行动产出的星石重定向给 TargetSelected（可结合 Action.TargetAllocations 分配）
    EffectApplyRuleModifier    EffectType = 29 // 向目标施加规则修饰实例（RuleModifierRef + RuleLifetimeRef）
    EffectRemoveRuleModifier   EffectType = 30 // 按 RuleRemoveRef 条件移除目标规则修饰实例
    EffectPlaceDeckTopAsFieldMark EffectType = 31 // 将目标牌库顶 Value 张牌放置到目标角色面前并赋予 FieldMarkRef（默认面朝下）
    EffectRemoveFieldMark      EffectType = 32 // 移除目标角色面前指定 FieldMarkRef 的场标牌（Value 为数量；不足按可移除量结算；若 ElementRef 非空则仅移除该系别）；并刷新 Event.RemovedFieldCard* 为最近被移除牌快照
    EffectConvertEnergyStone   EffectType = 33 // 转换目标玩家能量区星石颜色（StoneRef=源颜色，StoneToRef=目标颜色，Value<=0 表示全部）
    EffectModifyPendingDamage  EffectType = 34 // 修改当前战斗结算链的待生效伤害值（TargetCurrentCombat；Value 可正可负；最终不低于 0）
    EffectPlacePlayedCardAsFieldMark EffectType = 35 // 将“本次打出的牌实体”放置到目标角色面前并赋予 FieldMarkRef（可结合 VisibilityRef 指定明暗）
    EffectSetCurrentCombatElement EffectType = 36 // 改写当前战斗上下文系别（通常 Target=TargetCurrentCombat，ElementRef 必填）
    EffectRedirectCurrentCombatTarget EffectType = 37 // 改写当前战斗承受者（Target=新承受者）
    EffectSetCurrentCounterExecutor EffectType = 38 // 改写当前战斗“应战执行者”（Target=新执行者）
    EffectPlaceHandCardAsFieldMark EffectType = 39 // 将目标手牌中的 Value 张牌放置到目标角色面前并赋予 FieldMarkRef（通常由 Action.UsedCardUUIDs 指定，支持 VisibilityRef）
    EffectRemoveSelectedFieldCard EffectType = 40 // 移除 Action.Targets[].SelectedFieldCards 指定的场上牌（最多 Value 张）；并刷新 Event.RemovedFieldCard* 为最近被移除牌快照
    EffectRevealRemovedFieldCard EffectType = 41 // 将 Event.RemovedFieldCard 公开展示（若该牌原为暗置场标）
    EffectAddTeamCup EffectType = 42 // 调整目标阵营星杯数（Target=TargetSelfTeam/TargetEnemyTeam；Value>0 增加，Value<0 移除；结果钳制在 [0, TargetStarCups]）
    EffectPlaceOverflowDiscardAsFieldMark EffectType = 43 // 将 Event.OverflowDiscardCardIDs 中的卡牌实体放置到目标角色面前并赋予 FieldMarkRef（通常用于“爆牌弃牌留场”）
    EffectGrantExtraTurn EffectType = 44 // 为目标角色追加额外回合（Value=追加回合数；Value<=0 按 1 处理）
    EffectTransferFieldMark EffectType = 45 // 将在场 FieldMark 实体从 FromTargetRef 指向范围迁移到 Target 指向角色（保持原实体与明暗）
)

// —— 当前战斗改写执行约束（Combat Rewrite）——
// 1) EffectSetCurrentCombatElement：ElementRef 必填；Target 建议固定为 TargetCurrentCombat。
// 2) EffectRedirectCurrentCombatTarget：Target 必须可解析为单一玩家；写入 CombatContext.TargetID。
// 3) EffectSetCurrentCounterExecutor：Target 必须可解析为单一玩家；写入 CombatContext.CounterExecutorID。
// 4) 若同链路同时改写目标与应战执行者，建议先执行 EffectRedirectCurrentCombatTarget，再执行 EffectSetCurrentCounterExecutor。

// —— 效果作用目标枚举 (Effect Target Type) ——
type EffectTargetType int
const (
    TargetNone            EffectTargetType = 0
    TargetSelf            EffectTargetType = 1 // 技能发动者自己
    TargetSelected        EffectTargetType = 2 // 玩家通过 UI 选中的目标 (Enemy/Teammate由TargetRule限制)
    TargetAllEnemies      EffectTargetType = 3 // 所有敌人 (AOE)
    TargetAllTeammates    EffectTargetType = 4 // 所有队友 (含自己)
    TargetAllPlayers      EffectTargetType = 5 // 全场所有人
    TargetTriggerSource   EffectTargetType = 6 // 触发该被动的源头 (如: 暗杀者反伤攻击她的人)
    TargetSelfTeam        EffectTargetType = 7 // 己方阵营 (如神之庇护抵御士气下降)
    TargetCurrentCombat   EffectTargetType = 8 // 当前战斗上下文 (用于伤害修饰等)
    TargetCurrentEvent    EffectTargetType = 9 // 当前事件上下文 (用于修改 PendingMoraleLoss 等待结算值)
    TargetEnemyTeam       EffectTargetType = 10 // 敌方阵营（用于跨队星石转移）
    TargetAllOthers       EffectTargetType = 11 // 全体其他角色（排除技能发动者自己）
    TargetAllExceptSelected EffectTargetType = 12 // 全体角色中排除 SubmitAction 已选目标（用于“除所选目标外其余角色”）
)

type RuleModifierRemoveMode int
const (
    RuleRemoveByModifierID RuleModifierRemoveMode = 0
    RuleRemoveByDomain RuleModifierRemoveMode = 1
    RuleRemoveBySourceSkill RuleModifierRemoveMode = 2
    RuleRemoveAll RuleModifierRemoveMode = 3
)

type RuleModifierRemoveQuery struct {
    Mode model.RuleModifierRemoveMode
    ModifierID *string
    Domain *model.RuleModifierDomain
    SourceSkillID *string
    Limit int // <=0 表示不限制数量
}

// —— 逐目标响应分支配置 (Per Target Branch Config) ——
type PerTargetSourceType int
const (
    PerTargetSelectedTargets PerTargetSourceType = 0 // 遍历本次技能已选中的目标列表
)

type PerTargetBranchConfig struct {
    TargetSource        PerTargetSourceType       // 从哪里取“逐个处理”的目标集合
    InterruptType       model.InterruptType       // 对每个目标发起的中断类型（如 WaitDiscard / WaitChoice）
    TimeoutAsDeclined   bool                      // 超时是否按“失败分支”处理
    DiscardRequirement  *DiscardRequirement       // 当 InterruptType=WaitDiscard 时，目标需满足的弃牌要求
    DiscardVisibility   *model.CardVisibilityType // 目标弃牌展示可见性（如 VisibilityPublic）
    OnSuccess           []EffectNode              // 目标响应成功时执行
    OnDeclined          []EffectNode              // 目标响应失败/拒绝时执行
}
```

### 4.7 状态延后结算配置 (Status Resolve Config)
用于描述“状态牌被放置后，在未来某个时点才触发结算”的规则。技能本身只负责 `EffectPlaceStatus`，真正的延后结算逻辑统一由该配置驱动。

```go
// 0) 状态延后结算枚举
type StatusResolveMoment int
const (
    ResolveNone                               StatusResolveMoment = 0
    ResolveOnHolderNextBeforeAction           StatusResolveMoment = 1 // 持有者下个行动阶段开始前
    ResolveOnHolderCardPlayedOrRevealed       StatusResolveMoment = 2 // 持有者打出/展示牌时
    ResolveOnHolderIncomingAttackOrMagicBullet StatusResolveMoment = 3 // 持有者遭受攻击或魔弹时
)

type StatusResolveMode int
const (
    ResolveModeAuto   StatusResolveMode = 0 // 自动结算
    ResolveModeChoice StatusResolveMode = 1 // 结算时弹窗选择
)

type StatusSystemOpType int
const (
    StatusSystemOpNone          StatusSystemOpType = 0
    SkipCurrentActionPhase      StatusSystemOpType = 1 // 跳过当前行动阶段
    SystemOp_StartMagicBulletChain StatusSystemOpType = 2 // 启动魔弹传递链路结算器
)

// 1) 状态延后结算主配置（统一用 Timing、EffectType、Target、Value、Ref 表达）
type StatusResolveConfig struct {
    ConfigID        string                  // 配置唯一ID（如 "status_resolve_poison"）
    StatusType      model.StatusEffect      // 作用于哪种状态（如 Poison）
    ResolveTiming   model.TriggerTiming     // 触发时机（如 TimingOnBeforeAction、TimingOnHitCheck）
    RequireHolderIsTurnPlayer bool          // 仅当状态持有者=当前回合玩家时触发（中毒、虚弱、五系束缚用）
    RequireHolderIsCombatTarget bool        // 仅当状态持有者=当前战斗目标时触发（圣盾用）
    RequirePlayedCardElementMatchesSealElement bool // 仅当打出/展示的牌系别=状态绑定的系别时触发（元素封印用）
    EnforceNextActionMustActiveAttackSource bool // 轻量动作锁（挑衅专用）：持有者下个行动阶段必须主动攻击 StatusMeta.SourceUserID，否则跳过该行动阶段
    ResolveMode     model.StatusResolveMode // Auto / Choice
    CanDecline      bool                    // 触发条件满足后是否允许放弃结算（圣盾应为 false）

    TriggerLimit    int                     // 最大触发次数（通常为 1）
    RemoveOnResolved bool                   // 结算完成后是否移除该状态牌
    TimeoutFallbackChoiceIndex int          // Choice 模式下超时默认分支索引

    ResolveEffects  []EffectNode            // Auto 模式：结算效果序列
    Choices         []StatusResolveChoice   // Choice 模式：候选分支
}

// 2) Choice 分支定义
type StatusResolveChoice struct {
    ChoiceID   string          // 分支ID（如 "skip_action"）
    Label      string          // 前端按钮文案
    Effects    []EffectNode    // 本分支执行的效果（EffectType、Target、Value、Ref）
}
```



# 第三部分：核心引擎动态实体字典 (Dynamic Entities & Runtime State)

**文档说明**：本文档定义了游戏引擎运行期间的“动态实体（Dynamic Entities）”。它们代表了游戏当前的真实状态（如谁在哪个房间、手里拿着什么牌、有多少血量）。这些数据在游戏过程中被高频读写，也是持久化（存档/断线重连）的核心数据。

---

## 1. 卡牌实例实体 (Card Instance)
区别于静态的 `CardTemplate`（卡牌模板），这是玩家游戏中真正抓在手里或盖在场上的那张“物理卡牌”。

```go
type CardInstance struct {
    UUID       string             // 每一局游戏中，每一张牌的绝对唯一全局ID（防止作弊和状态混乱）
    TemplateID string             // 指向静态配置表中的 CardTemplate (获取它的基础伤害、属性等)
    
    FieldMark  model.CardFieldMark // 当前在场上的特殊标记（默认 None。如果是盖牌，这里就是 Cover）
}
```

---

## 2. 玩家状态实体 (Player Entity)
记录单个玩家在对局中的所有实时数据。这是事件总线和技能结算时修改最频繁的对象。

```go
type Player struct {
    UserID       string                 // 玩家账号唯一标识
    Nickname     string                 // 玩家昵称
    State        model.PlayerState      // 网络状态（如 Ready, Playing, Offline）
    Team         model.Faction          // 所属阵营（Red 或 Blue）
    SeatIndex    int                    // 座位号（决定行动顺序，如 0~5）
    
    // —— 角色与状态数据 ——
    CharacterID  string                 // 选定的角色ID（指向静态 CharacterTemplate）
    Orientation  model.CharacterOrientation // 当前姿态（正常 / 横置）
    Form         *string                // 当前形态（如 "shadow"、"holy_glory"，nil 表示无名横置）
    
    // —— 游戏内资产数据 ——
    HandCards    []*CardInstance        // 当前手牌列表
    FieldCards   []*CardInstance        // 放置在面前的场上牌（包含基础效果、盖牌、影、茧等）

    // 可选：性能索引（派生缓存，不是主数据）
    StatusIndex map[model.StatusEffect][]string // value: FieldCardID 列表
    
    Energy       []model.StarStoneType  // 能量区（存放 Gem 或 Crystal）
    Heal         int                    // 当前治疗点数
    
    Tokens       map[model.TokenType]int// 专属指示物数量缓存（如英灵人形：WarRune=2, MagicRune=1）
    
    // —— 运行时规则约束（由 EffectApplyRuleModifier / EffectRemoveRuleModifier 维护） ——
    ActiveRuleModifiers map[string]*RuleModifierInstance // Key: RuleInstanceID

    // —— 回合临时状态标记 (Turn End 时清空) ——
    HasStartup   bool                   // 本回合是否已发动过启动技
    ActionCount  int                    // 本回合已执行的普通行动次数
    AttackCount  int                    // 本回合已发动的攻击次数（用于剑魔、剑圣等判定）
}

type RuleModifierInstance struct {
    RuleInstanceID string                    // 实例ID（运行时唯一）
    ModifierID     string                    // 对应 RuleModifierTemplate.ModifierID
    SourceUserID   string                    // 施加来源用户
    SourceSkillID  string                    // 施加来源技能
    TargetUserID   string                    // 作用目标用户
    Lifetime       model.RuleModifierLifetimeType // 生命周期策略
}

type FieldCardInstance struct {
    ID         string
    CardUUID   string
    HolderID   string // 挂在谁面前
    FieldMark  model.CardFieldMark

    // 关键：状态语义挂在实体牌上
    StatusMeta *StatusMeta // nil 表示这张场上牌不是“基础效果”
}

type StatusMeta struct {
    EffectType model.StatusEffect
    Class      model.StatusClass // Basic / Exclusive
    BoundElement *model.ElementType // 绑定系别（元素封印用；其余状态为 nil）
    SourceUserID string            // 状态施加者（挑衅等“攻击来源约束”用）
}
```

---

## 3. 房间大厅实体 (Room Entity)
用于管理 WebSocket 连接、对局匹配、选角BP阶段的大厅对象。

```go
type Room struct {
    RoomID       string                 // 房间全局唯一短号（如 "89757"）
    RoomName     string                 // 房间名称
    State        model.RoomState        // 房间当前生命周期（Waiting, Drafting, Playing, Finished）
    
    OwnerID      string                 // 房主 UserID
    MaxPlayers   int                    // 最大玩家数（通常为 6）
    
    Players      map[string]*Player     // 房间内的所有玩家集合 (Key 为 UserID)
    
    // 如果房间状态进入 Playing，则实例化并挂载核心游戏引擎
    GameEngine   *GameContext           // 正在运行的对局上下文状态机
}
```

---

## 4. 核心对局上下文 (Game Context / Engine State)
这是整个对局的“上帝视角”，维护着当前的回合轮转、公共资源区（战绩区/士气）以及微型状态机的流转。

```go
type GameContext struct {
    // —— 全局对局资源 ——
    RedMorale    int                    // 红方当前士气
    BlueMorale   int                    // 蓝方当前士气
    RedCups      int                    // 红方星杯数
    BlueCups     int                    // 蓝方星杯数
    
    RedStones    []model.StarStoneType  // 红方战绩区星石
    BlueStones   []model.StarStoneType  // 蓝方战绩区星石
    
    Deck         []*CardInstance        // 当前牌库（待抽取的牌）
    DiscardPile  []*CardInstance        // 弃牌堆
    
    // —— 状态机轮转控制 ——
    CurrentPhase model.GamePhase        // 引擎当前处于哪个核心阶段（如 PhaseCombatHitCheck）
    TurnPlayerID string                 // 当前大回合所属的玩家 UserID
    
    // —— 战斗结算微型上下文 (仅在 Combat 的 6 个阶段中存在) ——
    CurrentCombat *CombatContext        // 记录当前正在发生的攻击/法术的结算状态

    // —— 本回合事件轨迹索引（用于“某技能是否造成过士气下降”类判定） ——
    SkillMoraleDropTraces []SkillMoraleDropTrace
    TurnDamageTraces      []TurnDamageTrace // 本回合伤害轨迹（用于“去重目标数”类判定）
    
    // —— 挂起与中断状态 ——
    PendingInterrupt *InterruptContext  // 如果引擎挂起，记录正在等待谁、做什么操作
}

type SkillMoraleDropTrace struct {
    TurnOwnerID   string        // 该轨迹所属回合的回合玩家
    SourceUserID  string        // 伤害来源玩家
    SourceSkillID string        // 伤害来源技能ID（普攻/无技能时可为空）
    TargetTeam    model.Faction // 实际士气下降的阵营
    Amount        int           // 本次实际下降值（>0）
}

type TurnDamageTrace struct {
    TurnOwnerID   string // 该轨迹所属回合的回合玩家
    SourceUserID  string // 伤害来源玩家
    SourceSkillID string // 伤害来源技能ID（普攻/无技能时可为空）
    TargetUserID  string // 受伤目标玩家（去重维度）
    IsMagic       bool   // 是否为法术伤害（true=法术；false=攻击）
    Amount        int    // 本次实际生效伤害（>0）
}

// 战斗过程上下文（记录一次攻击/法术从发起到结算的完整生命周期）
type CombatContext struct {
    SourceID     string                 // 发起者 UserID
    TargetID     string                 // 承受者 UserID (多目标可改用数组)
    SkillID      string                 // 触发的技能ID（如果是普攻则为空）
    AttackCard   *CardInstance          // 记录打出的攻击牌/法术牌（用于判定系别、命格）
    FlowType     model.ActionFlowType   // 当前结算流水线（NormalCombat / MagicBulletChain）
    OverrideElement *model.ElementType  // 当前战斗系别覆写（nil 表示沿用 AttackCard 系别）
    CounterExecutorID string            // 当前“应战执行者”用户ID（为空时按引擎默认）
    
    BaseDamage   int                    // 初始伤害计算值
    ExtraDamage  int                    // 额外增减伤（如狂战士撕裂+2）
    FinalDamage  int                    // 最终将造成的实际伤害
    
    IsHit        bool                   // 是否命中（如果被应战、圣盾、圣光抵挡，则设为 false）
}

// 中断上下文（告诉前端现在要弹出什么框）
type InterruptContext struct {
    Type         model.InterruptType    // 挂起类型（如 WaitResponse, WaitDiscard）
    TargetUserID string                 // 正在等待哪个玩家的操作
    Timeout      int                    // 超时时间（秒，用于AI托管逻辑）
    
    // 动态限制条件下发给前端
    DiscardFilter *model.CardFilterConfig // 如果是 WaitDiscard，告诉前端只能选哪些牌
}
```
# 第四部分：核心引擎前后端通信协议 (WebSocket DTOs)

**文档说明**：本文档定义了前后端通过 WebSocket 进行交互的标准数据传输对象（DTO）。《星杯传说》采用**服务器权威（Server-Authoritative）**架构，前端（Vue）仅作为状态的展示层和操作的收集器，所有核心逻辑、状态变更及出牌校验均由后端（Go）完成并下发。

---

## 1. 基础通信信封 (The Envelope)
所有的 WebSocket 消息（无论上行还是下行）都必须包裹在此标准信封中，以便前端 Router 或后端 Handler 进行路由分发。

```go
// Go 结构体定义
type WsMessage struct {
    Cmd  string      // 核心指令：如 "SyncState", "RequireAction", "NotifyTimeline", "SubmitAction"
    Data interface{} // 具体的负载数据 (Payload)，根据 Cmd 的不同解析为不同的结构体
}
```

---

## 2. 下行协议：服务端 -> 客户端 (Server to Client)
后端主动推给前端的消息，分为“状态刷新”、“操作阻断”和“时间线演出事件”三大类。

### 2.1 全局状态同步 (`Cmd: "SyncState"`)
用于断线重连、或者每个阶段/动作结算完毕后，全量刷新前端的 Vuex/Pinia 状态库。
**⚠️ 安全原则（战争迷雾）**：下发的数据中会自动屏蔽其他玩家的手牌 UUID 和具体 TemplateID，仅下发数量。

```json
// Data 负载示例 (前端接收到的 JSON)：
{
  "room_state": 2,              // 房间状态 (Playing)
  "current_phase": 4,           // 当前阶段 (ActionStart)
  "turn_player_id": "user_123", // 当前回合归属玩家
  "morale_red": 15,
  "morale_blue": 15,
  "cups_red": 0,
  "cups_blue": 0,
  "stones_red": [1, 2],         // 红方战绩区星石 (如：1宝石, 1水晶)
  "players": [
    {
      "user_id": "user_123",
      "character_id": "role_sword_emperor",
      "hand_count": 5,          // 其他玩家只发数量
      "hand_cards": [           // 只有当前玩家自己的视角，才会包含具体手牌数据
        {"uuid": "c-991", "template_id": "card_attack_fire_1", "field_mark": 0},
        {"uuid": "c-992", "template_id": "card_spell_heal", "field_mark": 0}
      ],
      "field_cards": [],        // 场上卡牌（如基础效果、影、茧）
      "energy": [1, 1],         // 能量区
      "heal": 0,
      "tokens": {"SwordQi": 2}  // 专属指示物（剑气: 2）
    }
    // ... 其他玩家数据对象
  ]
}
```

### 2.2 动作请求与中断弹窗 (`Cmd: "RequireAction"`)
当引擎状态机挂起（遇到 `PendingInterrupt`）时下发。前端收到后，必须立刻弹出对应的 UI 操作面板或解锁出牌按钮。

```json
// Data 负载示例 (例：要求玩家弃2张同系牌)：
{
  "interrupt_type": "WaitDiscard",  // 挂起类型：要求弃牌
  "target_user_id": "user_123",     // 要求谁操作（非该玩家的前端仅显示等待提示）
  "timeout": 30,                    // 倒计时 (秒)
  "msg": "发动技能需要弃2张同系牌，请选择：",
  
  // 核心：指示前端 UI 如何限制玩家的点选
  "valid_actions": ["Submit"],      // 允许的操作类型
  "card_filter": {                  // 卡牌过滤器（由前端据此将不符合条件的牌置灰禁用）
    "req_card_type": null,
    "req_element": null,
    "same_attribute": 1             // MatchElement (必须同系)
  },
  "require_count": 2                // 必须选中2张
  // 【全新升级：二级索敌规则】告诉前端是否需要弹出子菜单
  "sub_select_rule": {
    "sub_type": "FieldMark",        // 如需要玩家选完人后，接着选此人身上的基础效果
    "sub_min_count": 1,
    "sub_max_count": 1
  }
}
```

### 2.3 时间线事件通知 (`Cmd: "NotifyTimeline"`)
用于向前端推送可动画化的事件流。此通道不承载“真值状态”，仅用于演出播放；真值仍以 `SyncState` 为准。

```go
type TimelineNotifyPayload struct {
    RoomID string
    SeqStart int64           // 本批次首事件 EventID（含）
    SeqEnd int64             // 本批次末事件 EventID（含）
    IsReplay bool            // true=历史回放补发；false=实时推送
    Events []TimelineEvent
}

type TimelineEventType int
const (
    TimelinePhaseEntered TimelineEventType = 0
    TimelineActionDeclared TimelineEventType = 1
    TimelineActionRejected TimelineEventType = 2
    TimelineResponseWindowOpened TimelineEventType = 3
    TimelineResponseSelected TimelineEventType = 4
    TimelineResponseDeclined TimelineEventType = 5
    TimelineSkillTriggered TimelineEventType = 6
    TimelineEffectResolved TimelineEventType = 7
    TimelineCombatResolved TimelineEventType = 8
    TimelineStatusResolved TimelineEventType = 9
    TimelineInterruptRaised TimelineEventType = 10
    TimelineInterruptCleared TimelineEventType = 11
    TimelineChainClosed TimelineEventType = 12
)

type TimelineEventOutcome int
const (
    TimelineOutcomeNone TimelineEventOutcome = 0
    TimelineOutcomeSuccess TimelineEventOutcome = 1
    TimelineOutcomeDeclined TimelineEventOutcome = 2
    TimelineOutcomeBlocked TimelineEventOutcome = 3
    TimelineOutcomeMiss TimelineEventOutcome = 4
    TimelineOutcomeTimeout TimelineEventOutcome = 5
    TimelineOutcomeRejected TimelineEventOutcome = 6
)

type TimelineEventVisibility int
const (
    TimelineVisibilityPublic TimelineEventVisibility = 0
    TimelineVisibilityActorOnly TimelineEventVisibility = 1
    TimelineVisibilityActorTeam TimelineEventVisibility = 2
    TimelineVisibilityTargetOnly TimelineEventVisibility = 3
)

type TimelineDeltaType int
const (
    TimelineDeltaDamage TimelineDeltaType = 0
    TimelineDeltaHeal TimelineDeltaType = 1
    TimelineDeltaMorale TimelineDeltaType = 2
    TimelineDeltaHandCount TimelineDeltaType = 3
    TimelineDeltaTeamStone TimelineDeltaType = 4
    TimelineDeltaTeamCup TimelineDeltaType = 5
    TimelineDeltaEnergyStone TimelineDeltaType = 6
    TimelineDeltaToken TimelineDeltaType = 7
    TimelineDeltaFieldMark TimelineDeltaType = 8
    TimelineDeltaStatus TimelineDeltaType = 9
    TimelineDeltaOrientation TimelineDeltaType = 10
    TimelineDeltaForm TimelineDeltaType = 11
)

type TimelineDelta struct {
    Type model.TimelineDeltaType
    TargetUserID string
    Value int
    StoneRef *model.StarStoneType
    TokenRef *model.TokenType
    FieldMarkRef *model.CardFieldMark
    StatusRef *model.StatusEffect
}

type TimelineEvent struct {
    EventID int64
    TurnID int
    Phase model.GamePhase
    Timing model.TriggerTiming
    ChainID string
    ParentEventID *int64

    Type model.TimelineEventType
    Outcome model.TimelineEventOutcome
    Visibility model.TimelineEventVisibility

    ActorUserID string
    TargetUserIDs []string
    ActionType *model.ActionType
    SkillID *string
    CardIDs []string
    Deltas []TimelineDelta
    Message string
}
```

```json
// Data 负载示例（某玩家在命中判定窗使用圣光）：
{
  "room_id": "room_1001",
  "seq_start": 10231,
  "seq_end": 10233,
  "is_replay": false,
  "events": [
    {
      "event_id": 10231,
      "turn_id": 15,
      "phase": "CombatHitCheck",
      "timing": "TimingOnHitCheck",
      "chain_id": "chain_889",
      "type": "TimelineResponseWindowOpened",
      "outcome": "TimelineOutcomeNone",
      "visibility": "TimelineVisibilityPublic",
      "actor_user_id": "user_b",
      "target_user_ids": ["user_b"],
      "message": "user_b 可选择应战/圣光/圣盾"
    },
    {
      "event_id": 10232,
      "turn_id": 15,
      "phase": "CombatHitCheck",
      "timing": "TimingOnHitCheck",
      "chain_id": "chain_889",
      "parent_event_id": 10231,
      "type": "TimelineResponseSelected",
      "outcome": "TimelineOutcomeBlocked",
      "visibility": "TimelineVisibilityPublic",
      "actor_user_id": "user_b",
      "target_user_ids": ["user_b"],
      "card_ids": ["card_inst_771"],
      "message": "user_b 打出圣光，抵挡本次攻击"
    },
    {
      "event_id": 10233,
      "turn_id": 15,
      "phase": "CombatHitCheck",
      "timing": "TimingOnHitCheck",
      "chain_id": "chain_889",
      "type": "TimelineChainClosed",
      "outcome": "TimelineOutcomeSuccess",
      "visibility": "TimelineVisibilityPublic",
      "actor_user_id": "user_a",
      "message": "本次命中判定链路结束"
    }
  ]
}
```

### 2.3.1 时间线发射规范 (Timeline Emission Spec)
1. **严格顺序**：同一房间 `EventID` 严格递增，服务端按增序广播，前端按增序播放。
2. **链路闭合**：每条结算链（`ChainID`）必须以 `TimelineChainClosed` 结束，避免前端动画队列悬挂。
3. **交互可追踪**：涉及玩家抉择时，必须包含“窗口打开 -> 选择/放弃/超时 -> 窗口关闭”事件对。
4. **效果有 Delta**：产生状态变化的事件必须写入 `Deltas`；无变化也要发事件并给出 `Outcome`。
5. **可见性约束**：按 `Visibility` 裁剪接收方；禁止向无权限客户端广播私密信息（暗置牌、手牌细节等）。
6. **真值分离**：Timeline 仅用于动画，不作为前端最终状态来源；关键结算段后必须跟随 `SyncState` 纠偏。
7. **断线续播**：重连后可按 `SeqStart/SeqEnd` 批量补发历史事件（`IsReplay=true`），并与最新 `SyncState` 对齐。

### 2.3.2 兼容说明 (`Cmd: "NotifyEvent"`)
历史文本播报通道 `NotifyEvent` 可继续保留，但建议逐步迁移到 `NotifyTimeline`。  
迁移期间可双发：`NotifyTimeline` 供动画与日志，`NotifyEvent` 供旧 UI 文本战报。

---

## 3. 上行协议：客户端 -> 服务端 (Client to Server)
前端收集玩家在屏幕上的点选结果（选了哪几张牌、点了谁的头像），打包提交给后端。

### 3.1 玩家提交行动 (`Cmd: "SubmitAction"`)
这是前端发给后端的**唯一**核心操作结构。无论是攻击、交治疗、弃牌还是放技能，全部复用此结构。前端不做规则校验，只负责传参。

```go
// Go 接收端结构体定义
type ClientActionRequest struct {
    ActionType string   // 动作指令类型 (如: "Attack", "Magic", "Skill", "Buy", "Pass")
    
    // 玩家在手牌区或场上区选中的卡牌 UUID 列表
    // (例如：打出水涟斩的 UUID，或者选中2张准备弃掉的牌的 UUID)
	UsedCardUUIDs []string
    
    // 复合目标树
	// 允许前端不仅传回选中的人，还能传回选中的附属物（状态、对方手牌等）
	Targets []TargetNode

    // 可选：当技能要求“对已选目标进行点数分配”时填写（Key: TargetUserID, Value: 分配值）
    TargetAllocations map[string]int

    // 可选：当技能要求在追加行动中二选一时填写（如 Attack / Magic）
    ActionRef *model.ActionType

    // 可选：当某效果需要玩家为“Any 星石”选择具体颜色时填写（Gem / Crystal）
    StoneRef *model.StarStoneType

    // 可选：当技能/行动改写需要玩家指定元素时填写（如“视为任意系主动攻击”）
    ElementRef *model.ElementType

    // 可选：命名数值输入（支持同一技能同时输入多个变量，如 X/Y）
    // 例：{"X": 3, "Y": 2}
    NamedValues map[string]int
    
    // 如果 ActionType 是 "Skill"，此处必填技能的唯一 ID
    SkillID string 
}

// 复合目标节点
type TargetNode struct {
    TargetUserID   string   // 第一层：选中的目标玩家 ID

    // 第二层：在这个玩家身上选了什么附属物？（前端按需填入，不需要时为空）
	SelectedFieldCards []string // 传 FieldCardID
    SelectedHandCards  []string // 选中的对方手牌 UUID (用于看牌、偷牌等技能)
    SelectedTokens []string // 选中的专属指示物 (如需要移除某人身上的剑气)
}
```

#### 上行通信示例：
**场景 A：玩家放弃应战/放弃交治疗（点击跳过）**
```json
{
  "Cmd": "SubmitAction",
  "Data": {
    "ActionType": "Pass"
  }
}
```

**场景 B：玩家对 User_B 发起攻击（打出1张火系攻击牌）**
```json
{
  "Cmd": "SubmitAction",
  "Data": {
    "ActionType": "Attack",
    "SelectedCardUUIDs": ["uuid_card_001"],
    "TargetUserIDs": ["user_B"]
  }
}
```

**场景 C：玩家发动技能【治愈之光】给 User_B 和 User_C 加血**
```json
{
  "Cmd": "SubmitAction",
  "Data": {
    "ActionType": "Skill",
    "SkillID": "skill_saintess_heal",
    "SelectedCardUUIDs": [],
    "TargetUserIDs": ["user_B", "user_C"]
  }
}
```

**场景 D：玩家发动技能【治愈之光】给 User_B 和 User_C 加血**
```json
{
  "Cmd": "SubmitAction",
  "Data": {
    "ActionType": "Skill",
    "SkillID": "skill_angel_song",
    "Targets": [
      {
        "TargetUserID": "user_B",
        "SelectedMarks": ["MarkPoison"]
      }
    ]
  }
}
```

```markdown
# 第五部分：事件总线上下文与规则修饰符 (Event Context & Rule Modifiers)

**文档说明**：在复杂的卡牌游戏中，技能往往不仅是简单的数值加减，还包含对底层规则的“临时篡改”（如：本次攻击不可应战），以及极其严苛的触发责任人判定（如：必须是“你主动”移除标记时）。本部分定义了用于支撑这些高级逻辑的上下文实体与过滤器。

---

## 1. 规则修饰符 (Rule Modifiers) 

为了支持诸如“无视圣盾”、“不可应战”、“不掉士气”等效果，我们需要在动态战斗实体中加入特权标记，并在微观效果枚举中增加对应的修改指令。

### 1.1 战斗上下文升级 (`CombatContext` 补充)
在引擎的战斗状态机中，增加用于旁路（Bypass）标准规则的布尔值开关。

```go
type CombatContext struct {
    SourceID     string                 // 发起者 UserID
    TargetID     string                 // 承受者 UserID
    AttackCard   *model.CardInstance    // 记录打出的攻击牌
    FlowType     model.ActionFlowType   // 当前结算流水线（NormalCombat / MagicBulletChain）
    CounterExecutorID string            // 当前“应战执行者”用户ID（为空时按引擎默认）
    BaseDamage   int                    // 初始伤害计算值
    FinalDamage  int                    // 最终实际伤害
    IsHit        bool                   // 是否命中
    
    InterceptTags map[model.CombatInterceptTag]bool  //劫持标记池
    OverrideElement *model.ElementType  // 强规则：强行改变本次攻击的系别（覆盖原卡牌属性）
}
```

## 2. 事件溯源机制 (Event Source Tracking)

当引擎通过 EventBus 广播事件时（如某个基础效果被移除），必须携带完整的“案卷信息”，以便具有严苛触发条件的被动技能（如天使的【天使羁绊】）进行合法性校验。

### 2.1 事件总线负载上下文 (`EventContext`)

```go
// 事件来源类型枚举：区分是系统自然流转，还是玩家主动操作
type EventSourceType int
const (
    SourceSystem EventSourceType = 0 // 系统自然结算（如：回合结束挑衅消失、中毒跳血后消失）
    SourcePlayer EventSourceType = 1 // 玩家主动行为（如：打出某张牌、发动某个技能）
)

// 事件总线派发时的核心参数包
type EventContext struct {
    CombatCtx   *CombatContext      // 关联的战斗状态（若在战斗阶段外触发，则为 nil）
    IsCancelled bool                // 事件是否已被阻断
    
    // ======== 【事件溯源字段】 ========
    SourceType  EventSourceType     // 触发来源类型（System / Player）
    OperatorID  string              // 如果是 Player 造成的，记录该玩家的 UserID
    CauseAction string              // 导致该事件的具体原因（如：打出的卡牌UUID，或技能的 SkillID）
    
    // ======== 【标记改变专用字段 (TimingOnFieldMarkChanged)】 ========
    MarkType    model.CardFieldMark // 发生变动的标记（如：HolyShield, Poison）
    MarkAction  string              // 变动动作："Placed"(放置) 或 "Removed"(移除)
    
    // ======== 【伤害摸牌/士气结算字段 (TimingOnDamageTaken + CombatDraw)】 ========
    PendingMoraleLoss int           // 当前窗口待扣士气值（>0 时，允许神之庇护等“X点抵御”技能响应；CombatDraw 收尾时一次性落地）
    MoraleDropSourceUserID string   // 本次待扣士气对应的伤害来源玩家
    MoraleDropSourceSkillID string  // 本次待扣士气对应的伤害来源技能ID（普攻/无技能时可为空）
    MoraleDropApplied int           // 当前窗口最终落地的士气下降值（收尾后写入；未落地前为0）

    // ======== 【提炼产出重定向专用字段 (Extract Redirect)】 ========
    ExtractOutputStones []model.StarStoneType // 本次 Extract 产出的星石明细（用于重定向/分配）

    // ======== 【批量弃牌结果统计字段】 ========
    DiscardedMagicCount int                    // 本链路最近一次批量弃牌中，弃掉的法术牌数量
    DiscardedElementCount map[model.ElementType]int // 本链路最近一次批量弃牌中，按系别统计的弃牌数量

    // ======== 【爆牌弃牌快照字段】 ========
    OverflowDiscardOwnerID string              // 本次因超手牌上限发生爆牌弃置的角色ID
    OverflowDiscardCardIDs []string            // 本次爆牌弃置产生的卡牌实体列表（按弃置顺序）
    OverflowDiscardCount int                   // OverflowDiscardCardIDs 的快照数量
    
    // ======== 【打出/展示牌专用字段 (TimingOnCardPlayedOrRevealed)】 ========
    PlayedCardElement model.ElementType // 打出或展示的卡牌系别（元素封印用此与 StatusMeta.BoundElement 比较）

    // ======== 【场标移除快照字段】 ========
    RemovedFieldCardID string           // 本链路最近一次被移除的场上牌ID（由 EffectRemoveFieldMark / EffectRemoveSelectedFieldCard 写入）
    RemovedFieldCardType model.CardType // 被移除场上牌对应原卡牌类型（Attack/Magic/Base）
    RemovedFieldCardElement model.ElementType // 被移除场上牌对应原卡牌系别（用于后续分支判定）
    RemovedFieldCardMark model.CardFieldMark  // 被移除场上牌的场标类型（如 DemonForce / Cover）
    RemovedFieldCardWasHidden bool      // 被移除时是否为暗置（用于展示分支）

}

// ======== 表达式辅助函数约定 ========
// State.GetSelectedFieldCardElement(selfUserID, actionTargets):
// 读取本次提交中“由自身提交的 SelectedFieldCards 首张牌”对应原卡牌系别；未命中返回 None。
```

> 运行时约定：当 `EffectDiscard` 完成一次批量弃牌（包括“弃掉全部手牌”）后，引擎应在当前结算链的 `EventContext` 刷新 `DiscardedMagicCount` 与 `DiscardedElementCount`，供同链后续 `Condition/Effect` 表达式读取。
>
> 运行时约定：当角色因超手牌上限发生爆牌弃置（尤其是由伤害导致摸牌后爆牌）时，引擎应写入 `OverflowDiscardOwnerID/OverflowDiscardCardIDs/OverflowDiscardCount`；当 `EffectPlaceOverflowDiscardAsFieldMark` 执行后，已搬运实体应从 `OverflowDiscardCardIDs` 中消费移除，避免同链路重复搬运。


# 第六部分：基础效果与系统级指令模型 (Status Effects & System Operations)

**文档说明**：为了避免在技能配置和状态机运转中使用模糊的“纯文本”指令（如直接写“抵挡攻击”、“跳过行动”），本部分将其全部抽象为严谨的数据结构、枚举和系统上下文API。

---

## 1. 基础效果配置模型 (Status Effect Config)
定义诸如中毒、虚弱、圣盾等基础效果（Status）的静态属性和生命周期。

```go
type StatusEffectConfig struct {
    StatusID     string               // 唯一标识 (如 "Status_Poison", "Status_HolyShield", "Status_Weakness")
    Name         string               // 名称 (如 "中毒", "圣盾")
    
    // --- 属性规则 ---
    IsUnique     bool                 // 是否唯一 (若为 true 则同一角色不可叠加)
    MaxStacks    int                  // 最大层数 (若不唯一时的堆叠上限)
    
    // --- 统一为“临时技能”模型 ---
    // 基础效果本质上是挂载在玩家身上的临时被动技/响应技
    Timing       model.TriggerTiming  // 监听的事件钩子 (如 TimingOnHitCheck, TimingBeforeActionExecute)
    Condition    *ConditionConfig     // 触发前置条件表达式
    Effects      []EffectNode         // 触发后执行的标准动作序列 (复用 EffectDamage, EffectSystemOp 等)
}
```

## 2. 系统级指令枚举 (System Operation Type)
用于引擎底层直接阻断流水线、改变阶段走向的最高优先级指令（SystemOp）。当效果触发或玩家做出抉择时，由 Handler 向状态机抛出此类指令。

```go
type SystemOperationType string

const (
    // 战斗/结算相关阻断
    SysOp_BlockAttackOrMagicBullet SystemOperationType = "BlockCurrentAttackOrMagicBullet" // 抵挡当前的攻击或魔弹判定 (用于圣盾/圣光)
    
    // 回合与阶段相关跳转
    SysOp_SkipActionPhase          SystemOperationType = "SkipCurrentActionPhase"          // 强制跳过 ActionStart、ActionExecution 阶段，直接进入 ActionEnd
    SysOp_TerminateTurn            SystemOperationType = "TerminateTurn"                   // 强制结束回合
)
```

## 3. 玩家交互抉择枚举 (Resolution Choice Type)
处理诸如【虚弱】、【五系束缚】等需要玩家在阶段初进行“二选一”操作的标准选项实体。

```go
type ResolutionChoiceType string

const (
    Choice_SkipAction      ResolutionChoiceType = "SkipAction"       // 抉择：跳过接下来的行动阶段
    Choice_DrawAndContinue ResolutionChoiceType = "DrawAndContinue"  // 抉择：摸牌并继续行动 (如虚弱摸3张，五系摸2+X张)
)
```

## 4. 引擎上下文 API 扩展 (Game Context API)
用于在技能的前置条件 (`CustomExpression`) 中，提供严谨的函数调用，代替纯文本的口语化描述。这些方法实现在 `GameContext` / `State` 单例中。

```go
// GameContext / State 暴露给表达式引擎的查询接口
type GameStateAPI interface {
    
    // --- 场上状态查询 ---
    HasAnyStatusOnField() bool                     // 全场是否存在任何基础效果 (用于天使/封印师大招前置校验)
    GetStatusCount(statusID string) int            // 获取全场某类基础效果的总数 (如用于封印师五系束缚算X值)
    
    // --- 历史轨迹查询 ---
    GetAttackCountThisTurn() int                   // 获取本回合当前玩家已发动的攻击次数 (用于剑圣/剑女)
    HasSkillUsedThisTurn(skillID string) bool      // 本回合是否发动过某技能 (用于回合限定)
    GetSkillMoraleDropThisTurn(sourceUserID string, sourceSkillID string) int // 本回合由指定技能造成的士气下降总量
    HasSkillMoraleDropThisTurn(sourceUserID string, sourceSkillID string) bool // 便捷查询：是否造成过士气下降（总量>0）
    CountTurnDistinctDamageTargets(sourceUserID string, onlyMagic bool, onlyEnemy bool) int // 本回合由指定来源造成过伤害的去重目标数
    HasTurnDistinctDamageTargetsAtLeast(sourceUserID string, onlyMagic bool, onlyEnemy bool, threshold int) bool // 便捷查询：去重目标数是否达到阈值
    IsTokenAtCap(userID string, tokenType model.TokenType) bool // 指定指示物是否达到当前有效上限
    CountTeamFieldMark(sourceUserID string, mark model.CardFieldMark) int // 查询 source 所在队伍当前持有的指定场标总数
    CountSelectedTargetsWithHealAtLeast(actionTargets []TargetNode, minHeal int) int // 统计本次已选目标中治疗>=minHeal的人数（如 Y=目标中治疗>0的人数）
    HasExecutedSpecialActionThisTurn(userID string) bool // 本回合是否执行过任意特殊行动（Buy/Synthesize/Extract/Deadlock）
    GetPlayerOrientation(userID string) model.CharacterOrientation // 查询指定角色当前姿态（Normal/Tapped）
    IsSameTeam(userA string, userB string) bool // 判断两名角色是否同阵营
    GetAliveTeammateCount(userID string) int // 获取当前存活队友数（不含自身）

    // --- 规则/面板查询 ---
    GetEffectivePlayerAttribute(userID string, attr model.PlayerAttributeType) int // 查询 RuleModifier 归并后的最终属性值
    IsSkillDisabled(userID string, skillID string) bool // 查询 SkillGate 规则是否禁用该技能
    GetActiveRuleModifiers(userID string) []RuleModifierInstance // 获取目标当前生效的规则实例
}
```

> 运行时约定（TurnDamageTraces）：
> 1. 仅在 `TimingOnDamageApplied` 且 `Amount > 0` 时写入；
> 2. 去重统计按 `TargetUserID` 做 distinct；
> 3. 查询默认作用于当前回合窗口（`TurnOwnerID == TurnPlayerID`）。
> 4. `CountSelectedTargetsWithHealAtLeast` 基于本次提交目标快照计算，同一结算链内多次读取应保持一致。
> 5. `GetPlayerOrientation` 在同一战斗结算链内应返回一致姿态快照，避免同链读值抖动。
> 6. `IsSameTeam/GetAliveTeammateCount` 返回当前时点阵营/存活关系快照，供目标合法性与“队友数量限制”判定使用。
