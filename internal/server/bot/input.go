package bot

import "starcup-engine/internal/model"

// DecisionInput 机器人决策输入（不依赖 viewmodel）
type DecisionInput struct {
	PlayerID        string
	State           StateSnapshot
	Prompt          *model.Prompt
	AvailableSkills []AvailableSkill
	Memory          *Memory
}

// StateSnapshot 游戏状态快照（精简版，仅机器人需要）
type StateSnapshot struct {
	TurnStage           string
	CombatStage         string
	Subflow             string
	CurrentPlayer       string
	HasPerformedStartup bool
	Players             map[string]PlayerSnapshot
	// 阵营资源
	RedMorale   int
	BlueMorale  int
	RedCups     int
	BlueCups    int
	RedGems     int
	BlueGems    int
	RedCrystals int
	BlueCrystals int
	// 牌库
	DeckCount    int
	DiscardCount int
}

// PlayerSnapshot 玩家视角（精简版，仅机器人决策需要）
type PlayerSnapshot struct {
	ID               string
	Name             string
	Camp             string
	Role             string
	Form             string
	Orientation      string
	HandCount        int
	MaxHand          int
	ExclusiveCardCount int
	Hand             []model.Card    // 仅自己有内容
	ExclusiveCards   []model.Card
	Field            []*model.FieldCard
	Heal             int
	MaxHeal          int
	Gem              int
	Crystal          int
	IsActive         bool
	Buffs            []model.Buff
	Tokens           map[string]int
}

// AvailableSkill 可发动技能摘要（精简版）
type AvailableSkill struct {
	ID               string
	Title            string
	Description      string
	MinTargets       int
	MaxTargets       int
	TargetType       int
	CostGem          int
	CostCrystal      int
	CostDiscards     int
	DiscardType      string
	DiscardElement   string
	RequireExclusive bool
	PlaceCard        bool
	PlaceEffect      string
}