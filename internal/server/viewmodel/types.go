package viewmodel

import "starcup-engine/internal/model"

// CharacterView 角色摘要（供前端展示，与后端 data.GetCharacters 一致）
type CharacterView struct {
	ID      string      `json:"id"`
	Name    string      `json:"name"`
	Title   string      `json:"title"`
	Faction string      `json:"faction"`
	Skills  []SkillView `json:"skills"`
}

// SkillView 技能摘要（含主动技元数据，供前端 fallback）
type SkillView struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	Type             int    `json:"type"`        // SkillType: 0=Passive, 1=Startup, 2=Action, 3=Response
	MinTargets       int    `json:"min_targets"` // 仅主动技有效
	MaxTargets       int    `json:"max_targets"` // 仅主动技有效
	TargetType       int    `json:"target_type"` // 仅主动技有效
	CostGem          int    `json:"cost_gem"`
	CostCrystal      int    `json:"cost_crystal"`
	CostDiscards     int    `json:"cost_discards"`
	DiscardElement   string `json:"discard_element,omitempty"`
	RequireExclusive bool   `json:"require_exclusive,omitempty"` // 是否必须使用独有牌
}

// RoomEvent represents room-related events.
type RoomEvent struct {
	Action         string          `json:"action"` // "joined", "left", "started", "player_list", "error", "assigned"
	RoomCode       string          `json:"room_code"`
	PlayerID       string          `json:"player_id,omitempty"`
	PlayerName     string          `json:"player_name,omitempty"`
	Players        []PlayerInfo    `json:"players,omitempty"`
	Characters     []CharacterView `json:"characters,omitempty"` // 角色与技能数据，从后端获取
	Message        string          `json:"message,omitempty"`
	Camp           string          `json:"camp,omitempty"`
	CharRole       string          `json:"char_role,omitempty"`
	ReconnectToken string          `json:"reconnect_token,omitempty"` // 断线重连令牌（仅发送给本人）
}

// PlayerInfo represents basic player information for room events.
type PlayerInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Camp     string `json:"camp"`
	CharRole string `json:"char_role"`
	Ready    bool   `json:"ready"`
	IsOnline bool   `json:"is_online"`
	IsBot    bool   `json:"is_bot,omitempty"`
	IsHost   bool   `json:"is_host,omitempty"`
	BotMode  string `json:"bot_mode,omitempty"`
}

// AvailableSkill 当前可发动的主动技能摘要（供前端展示与选目标）。
type AvailableSkill struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	MinTargets       int    `json:"min_targets"`
	MaxTargets       int    `json:"max_targets"`
	TargetType       int    `json:"target_type"` // model.TargetNone=0, TargetSelf=1, TargetEnemy=2, ...
	CostGem          int    `json:"cost_gem"`
	CostCrystal      int    `json:"cost_crystal"`
	CostDiscards     int    `json:"cost_discards"`
	DiscardType      string `json:"discard_type,omitempty"`      // 弃牌类型要求（Attack/Magic）
	DiscardElement   string `json:"discard_element,omitempty"`   // 弃牌元素要求（如 "Water"）
	RequireExclusive bool   `json:"require_exclusive,omitempty"` // 是否必须使用独有牌（卡牌下标了技能名）
	PlaceCard        bool   `json:"place_card,omitempty"`        // 是否放置场上牌
	PlaceEffect      string `json:"place_effect,omitempty"`      // 放置的效果类型（如 Shield/Poison/Weak）
}

// GameStateUpdate represents a filtered game state for a specific player.
type GameStateUpdate struct {
	TurnStage           string                `json:"turn_stage,omitempty"`
	CombatStage         string                `json:"combat_stage,omitempty"`
	Subflow             string                `json:"subflow,omitempty"`
	CurrentPlayer       string                `json:"current_player"`
	HasPerformedStartup bool                  `json:"has_performed_startup"`
	Players             map[string]PlayerView `json:"players"`
	RedMorale           int                   `json:"red_morale"`
	BlueMorale          int                   `json:"blue_morale"`
	RedCups             int                   `json:"red_cups"`
	BlueCups            int                   `json:"blue_cups"`
	RedGems             int                   `json:"red_gems"`
	BlueGems            int                   `json:"blue_gems"`
	RedCrystals         int                   `json:"red_crystals"`
	BlueCrystals        int                   `json:"blue_crystals"`
	DeckCount           int                   `json:"deck_count"`
	DiscardCount        int                   `json:"discard_count"`
	AvailableSkills     []AvailableSkill      `json:"available_skills"`
	Characters          []CharacterView       `json:"characters,omitempty"` // 角色与技能数据
}

// PlayerView represents a player's view (hiding other players' hands).
type PlayerView struct {
	ID                 string             `json:"id"`
	Name               string             `json:"name"`
	Camp               string             `json:"camp"`
	Role               string             `json:"role"`
	Orientation        string             `json:"orientation,omitempty"`
	Form               string             `json:"form,omitempty"`
	HandCount          int                `json:"hand_count"`
	MaxHand            int                `json:"max_hand"`
	ExclusiveCardCount int                `json:"exclusive_card_count"`
	Hand               []model.Card       `json:"hand,omitempty"`            // Only for self
	ExclusiveCards     []model.Card       `json:"exclusive_cards,omitempty"` // Only for self (专属技能卡区)
	Field              []*model.FieldCard `json:"field"`
	Heal               int                `json:"heal"`
	MaxHeal            int                `json:"max_heal"`
	Gem                int                `json:"gem"`     // 个人能量：宝石
	Crystal            int                `json:"crystal"` // 个人能量：水晶
	IsActive           bool               `json:"is_active"`
	Buffs              []model.Buff       `json:"buffs"`
	Tokens             map[string]int     `json:"tokens,omitempty"` // 纯指示物；不含 *_count / 派生镜像（见下方顶层字段）

	// 派生计数（JSON 可与历史 key 同名，但不放在 tokens map 内）
	ElfBlessingCount               int `json:"elf_blessing_count,omitempty"`
	MagicBowChargeCount            int `json:"mb_charge_count,omitempty"`
	SpiritCasterPowerCount         int `json:"sc_power_count,omitempty"`
	MoonDarkMoonCount              int `json:"mg_dark_moon_count,omitempty"`
	ButterflyCocoonCount           int `json:"bt_cocoon_count,omitempty"`
	BloodSharedLifeActive          int `json:"bp_shared_life_active,omitempty"`
	BloodSharedLifeBound           int `json:"bp_shared_life_bound,omitempty"`
	MagicLancerDarkReleaseBonus    int `json:"ml_dark_release_next_attack_bonus,omitempty"`
	MagicLancerDarkReleaseLockTurn int `json:"ml_dark_release_lock_turn,omitempty"` // 0/1，与 tokens 分离
	SwordEmperorSwordSoulCount     int `json:"se_sword_soul_count,omitempty"`

	// 额外行动约束（仅自身可见）
	CurrentExtraAction  string   `json:"current_extra_action,omitempty"`
	CurrentExtraElement []string `json:"current_extra_element,omitempty"`
}
