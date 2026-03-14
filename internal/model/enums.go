package model

// Camp 阵营
type Camp string

const (
	RedCamp  Camp = "Red"
	BlueCamp Camp = "Blue"
)

// Element 元素
type Element string

const (
	ElementEarth   Element = "Earth"   // 地
	ElementWater   Element = "Water"   // 水
	ElementFire    Element = "Fire"    // 火
	ElementWind    Element = "Wind"    // 风
	ElementThunder Element = "Thunder" // 雷
	ElementLight   Element = "Light"   // 光 (仅法术)
	ElementDark    Element = "Dark"    // 暗 (仅攻击-暗灭)
)

// CardType 卡牌类型
type CardType string

const (
	CardTypeAttack CardType = "Attack"
	CardTypeMagic  CardType = "Magic"
)

// GamePhase 游戏阶段
type GamePhase string

const (
	// 旧阶段（保留以保持兼容性）
	PhaseResponse         GamePhase = "Response"         // 响应阶段
	PhaseDiscardSelection GamePhase = "DiscardSelection" // 爆牌选择阶段
	PhaseEnd              GamePhase = "End"              // 游戏结束

	// 新的11步回合结构阶段
	PhaseBuffResolve       GamePhase = "BuffResolve"       // 1. Buff结算阶段
	PhaseStartup           GamePhase = "Startup"           // 2. 启动技能阶段
	PhaseActionSelection   GamePhase = "ActionSelection"   // 3. 行动选择阶段
	PhaseBeforeAction      GamePhase = "BeforeAction"      // 4. 行动前阶段
	PhaseActionExecution   GamePhase = "ActionExecution"   // 5. 行动执行阶段
	PhaseCombatInteraction GamePhase = "CombatInteraction" // 6. 战斗交互阶段（等待响应）
	PhaseDamageResolution  GamePhase = "DamageResolution"  // 7. 伤害结算阶段
	PhasePendingDamageResolution GamePhase = "PendingDamageResolution" // 7.5 延迟伤害结算
	PhaseExtraAction       GamePhase = "ExtraAction"       // 8. 额外行动阶段（处理队列）
	PhaseTurnEnd           GamePhase = "TurnEnd"           // 9. 回合结束阶段
)

// ActionType 行动类型
type ActionType string

const (
	ActionAttack     ActionType = "Attack"
	ActionMagic      ActionType = "Magic"
	ActionBuy        ActionType = "Buy"
	ActionSynthesize ActionType = "Synthesize"
	ActionExtract    ActionType = "Extract"
)
