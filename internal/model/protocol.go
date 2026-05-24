package model

// --- 1. 上行指令 (Action) ---

// PlayerActionType 玩家操作类型（指令类型）
type PlayerActionType string

const (
	// 系统指令
	CmdStart PlayerActionType = "Start"
	CmdQuit  PlayerActionType = "Quit"
	CmdPass  PlayerActionType = "Pass"
	CmdHelp  PlayerActionType = "Help"

	// 战斗/经济指令
	CmdAttack     PlayerActionType = "Attack"     // atk <target> <card_id>
	CmdMagic      PlayerActionType = "Magic"      // magic <target> <card_id>
	CmdBuy        PlayerActionType = "Buy"        // buy
	CmdSynthesize PlayerActionType = "Synthesize" // syb
	CmdExtract    PlayerActionType = "Extract"    // ext
	CmdSkill      PlayerActionType = "Skill"

	// 特殊状态指令
	CmdCannotAct PlayerActionType = "CannotAct" // 无法行动宣告

	// 交互响应指令
	CmdConfirm PlayerActionType = "Confirm" // confirm
	CmdCancel  PlayerActionType = "Cancel"  // cancel / skip
	CmdSelect  PlayerActionType = "Select"  // choose <idx...> / discard <idx...>
	CmdRespond PlayerActionType = "Respond" // take / counter / defend

	// 调试
	CmdCheat PlayerActionType = "Cheat"
)

func IsKnownPlayerActionType(actionType PlayerActionType) bool {
	switch actionType {
	case CmdStart,
		CmdQuit,
		CmdPass,
		CmdHelp,
		CmdAttack,
		CmdMagic,
		CmdBuy,
		CmdSynthesize,
		CmdExtract,
		CmdSkill,
		CmdCannotAct,
		CmdConfirm,
		CmdCancel,
		CmdSelect,
		CmdRespond,
		CmdCheat:
		return true
	default:
		return false
	}
}

// PlayerAction 玩家发送给引擎的唯一数据包
type PlayerAction struct {
	PlayerID string           `json:"player_id"`
	Type     PlayerActionType `json:"type"`

	TargetID string `json:"target_id,omitempty"` // single-target shorthand

	// 【新增】多目标支持
	TargetIDs []string `json:"target_ids,omitempty"`

	CardID     string   `json:"card_id,omitempty"`
	CardIDs    []string `json:"card_ids,omitempty"`
	SkillID    string   `json:"skill_id,omitempty"`
	Selections []int    `json:"selections,omitempty"`
	ExtraArgs  []string `json:"extra_args,omitempty"`
}

// --- 2. 下行事件 (Event) ---

// GameEventType 游戏事件类型
type GameEventType string

const (
	EventLog            GameEventType = "Log"            // 普通日志
	EventStateUpdate    GameEventType = "StateUpdate"    // 状态变更 (UI刷新)
	EventAskInput       GameEventType = "AskInput"       // 请求输入 (Prompt)
	EventError          GameEventType = "Error"          // 操作错误
	EventGameEnd        GameEventType = "GameEnd"        // 游戏结束
	EventCardRevealed   GameEventType = "CardRevealed"   // 明牌展示：出牌/弃牌等，供前端动画
	EventDamageDealt    GameEventType = "DamageDealt"    // 伤害结算：攻击/法术命中，供前端暴血特效
	EventActionStep     GameEventType = "ActionStep"     // 行动步骤：供桌面区域展示行动流程
	EventCombatCue      GameEventType = "CombatCue"      // 对战提示：攻击/防御/承受/应战，供前端对战动画
	EventDrawCards      GameEventType = "DrawCards"      // 摸牌事件：供前端公共牌堆->角色区动画
	EventSkillActivated GameEventType = "SkillActivated" // 技能成功发动，供前端技能牌匾/历史播报
	EventSpecialAction  GameEventType = "SpecialAction"  // 购买/提炼/合成等特殊行动
	EventStateDelta     GameEventType = "StateDelta"     // 公共可见状态变化
)

// GameEvent 引擎发送给 UI 的唯一数据包
type GameEvent struct {
	Type    GameEventType `json:"type"`
	Message string        `json:"message"` // 用于 Log/Error

	Prompt         *Prompt                `json:"prompt,omitempty"`
	CardRevealed   *CardRevealedPayload   `json:"card_revealed,omitempty"`
	DamageDealt    *DamageDealtPayload    `json:"damage_dealt,omitempty"`
	ActionStep     *ActionStepPayload     `json:"action_step,omitempty"`
	CombatCue      *CombatCuePayload      `json:"combat_cue,omitempty"`
	DrawCards      *DrawCardsPayload      `json:"draw_cards,omitempty"`
	SkillActivated *SkillActivatedPayload `json:"skill_activated,omitempty"`
	SpecialAction  *SpecialActionPayload  `json:"special_action,omitempty"`
	StateDelta     *StateDeltaPayload     `json:"state_delta,omitempty"`
}

// GameObserver 观察者接口
type GameObserver interface {
	OnGameEvent(event GameEvent)
}

// --- 3. WebSocket 消息类型 ---

// WSMessageType WebSocket消息类型
type WSMessageType string

const (
	WSTypeAction WSMessageType = "action" // 玩家指令
	WSTypeEvent  WSMessageType = "event"  // 游戏事件
	WSTypeRoom   WSMessageType = "room"   // 房间事件
	WSTypeChat   WSMessageType = "chat"   // 聊天消息
)

// WSEventType WebSocket事件子类型
type WSEventType string

const (
	WSEventLog            WSEventType = "log"             // 日志
	WSEventStateUpdate    WSEventType = "state_update"    // 状态更新
	WSEventPrompt         WSEventType = "prompt"          // 请求输入
	WSEventWaiting        WSEventType = "waiting"         // 等待其他玩家
	WSEventError          WSEventType = "error"           // 错误
	WSEventGameEnd        WSEventType = "game_end"        // 游戏结束
	WSEventChat           WSEventType = "chat"            // 聊天
	WSEventCardRevealed   WSEventType = "card_revealed"   // 明牌展示（出牌/弃牌动画）
	WSEventDamageDealt    WSEventType = "damage_dealt"    // 伤害结算（暴血特效）
	WSEventActionStep     WSEventType = "action_step"     // 行动步骤（桌面展示）
	WSEventCombatCue      WSEventType = "combat_cue"      // 对战提示（攻击/防御/承受/应战）
	WSEventDrawCards      WSEventType = "draw_cards"      // 摸牌动画触发（事件驱动）
	WSEventSkillActivated WSEventType = "skill_activated" // 技能成功发动
	WSEventSpecialAction  WSEventType = "special_action"  // 购买/提炼/合成
	WSEventStateDelta     WSEventType = "state_delta"     // 公共可见状态变化
)

// RoomActionType 房间事件类型
type RoomActionType string

const (
	RoomActionJoined     RoomActionType = "joined"      // 玩家加入
	RoomActionLeft       RoomActionType = "left"        // 玩家离开
	RoomActionStarted    RoomActionType = "started"     // 游戏开始
	RoomActionPlayerList RoomActionType = "player_list" // 玩家列表更新
	RoomActionAssigned   RoomActionType = "assigned"    // 分配玩家ID和阵营
	RoomActionError      RoomActionType = "error"       // 房间错误
)
