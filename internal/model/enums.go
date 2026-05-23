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

// ActionType 行动类型
type ActionType string

const (
	ActionAttack     ActionType = "Attack"
	ActionMagic      ActionType = "Magic"
	ActionBuy        ActionType = "Buy"
	ActionSynthesize ActionType = "Synthesize"
	ActionExtract    ActionType = "Extract"
)

const ExtraActionAny = "Any"
