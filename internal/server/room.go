package server

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"starcup-engine/internal/data"
	"starcup-engine/internal/engine"
	"starcup-engine/internal/engine/skills"
	"starcup-engine/internal/model"
)

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type    string          `json:"type"`    // "action", "room", "chat"
	Payload json.RawMessage `json:"payload"` // Raw JSON for flexible parsing
}

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

// RoomEvent represents room-related events
type RoomEvent struct {
	Action     string          `json:"action"` // "joined", "left", "started", "player_list", "error", "assigned"
	RoomCode   string          `json:"room_code"`
	PlayerID   string          `json:"player_id,omitempty"`
	PlayerName string          `json:"player_name,omitempty"`
	Players    []PlayerInfo    `json:"players,omitempty"`
	Characters []CharacterView `json:"characters,omitempty"` // 角色与技能数据，从后端获取
	Message    string          `json:"message,omitempty"`
	Camp       string          `json:"camp,omitempty"`
	CharRole   string          `json:"char_role,omitempty"`
	// 断线重连令牌（仅发送给本人）
	ReconnectToken string `json:"reconnect_token,omitempty"`
}

// PlayerInfo represents basic player information for room events
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

type lineupPlayer struct {
	id   string
	name string
	role string
	camp model.Camp
}

// AvailableSkill 当前可发动的主动技能摘要（供前端展示与选目标）
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

// GameStateUpdate represents a filtered game state for a specific player
type GameStateUpdate struct {
	Phase               string                `json:"phase"`
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

// PlayerView represents a player's view (hiding other players' hands)
type PlayerView struct {
	ID                 string             `json:"id"`
	Name               string             `json:"name"`
	Camp               string             `json:"camp"`
	Role               string             `json:"role"`
	HandCount          int                `json:"hand_count"`
	MaxHand            int                `json:"max_hand"`
	ExclusiveCardCount int                `json:"exclusive_card_count"`
	Hand               []model.Card       `json:"hand,omitempty"`            // Only for self
	Blessings          []model.Card       `json:"blessings,omitempty"`       // Only for self (精灵射手祝福)
	ExclusiveCards     []model.Card       `json:"exclusive_cards,omitempty"` // Only for self (专属技能卡区)
	Field              []*model.FieldCard `json:"field"`
	Heal               int                `json:"heal"`
	MaxHeal            int                `json:"max_heal"`
	Gem                int                `json:"gem"`     // 个人能量：宝石
	Crystal            int                `json:"crystal"` // 个人能量：水晶
	IsActive           bool               `json:"is_active"`
	Buffs              []model.Buff       `json:"buffs"`
	Tokens             map[string]int     `json:"tokens,omitempty"` // 指示物（审判/元素/形态等）
}

// Room represents a game room
type Room struct {
	Code    string
	Clients map[string]*Client
	Engine  *engine.GameEngine
	Started bool
	HostID  string
	// SeatOrder 为座次（用于前端固定展示顺序）。开局后按该顺序广播 player_list。
	SeatOrder []string

	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan []byte

	mu       sync.RWMutex
	engineMu sync.Mutex

	// 机器人全局观察信息（用于手牌类型推断）
	botIntel *botIntel
	// 机器人最近一次收到的Prompt缓存（用于非中断提示，如CombatInteraction/ActionSelection）
	botPromptCache map[string]*model.Prompt
	// AskInput 全局版本号：每次新提示+1，用于丢弃旧定时器动作。
	botPromptEpoch uint64
}

// Available character roles
var availableRoles = []string{
	"berserker", "blade_master", "sealer",
	"archer", "assassin", "angel",
	"saintess", "magical_girl",
	"valkyrie", "elementalist", "arbiter",
	"adventurer", "holy_lancer",
	"elf_archer", "plague_mage", "magic_swordsman", "crimson_sword_spirit",
	"prayer_master", "crimson_knight", "war_homunculus", "priest", "onmyoji",
	"blaze_witch",
	"sage", "magic_bow", "magic_lancer", "spirit_caster", "bard", "hero", "fighter", "holy_bow", "soul_sorcerer", "moon_goddess", "blood_priestess", "butterfly_dancer",
}

// NewRoom creates a new game room
func NewRoom(code string) *Room {
	room := &Room{
		Code:           code,
		Clients:        make(map[string]*Client),
		Register:       make(chan *Client),
		Unregister:     make(chan *Client),
		Broadcast:      make(chan []byte, 256),
		Started:        false,
		botIntel:       newBotIntel(),
		botPromptCache: make(map[string]*model.Prompt),
		botPromptEpoch: 0,
	}
	return room
}

// Run starts the room's main loop
func (r *Room) Run() {
	for {
		select {
		case client := <-r.Register:
			r.handleRegister(client)
		case client := <-r.Unregister:
			r.handleUnregister(client)
		case message := <-r.Broadcast:
			r.broadcastToAll(message)
		}
	}
}

func generateReconnectToken() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Int63())
}

func (r *Room) handleRegister(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 只要带了重连参数，优先尝试重连（无论是否已开局）。
	// 这样可避免旧连接尚未及时清理时，被“房间已满/游戏已开始”误拦截。
	hasReconnectParams := client.ReconnectPlayerID != "" || client.ReconnectToken != ""
	if hasReconnectParams {
		if r.tryReconnectLocked(client) {
			return
		}
		// token 校验失败时，允许按“房间码+player_id”兜底认领离线席位。
		if r.tryReconnectByPlayerIDLocked(client) {
			return
		}
		// token 失效时，允许按“房间码+名字”兜底认领离线席位。
		if r.tryReconnectByNameLocked(client) {
			return
		}
		log.Printf("Reconnect rejected in room %s: incoming_name=%s player_id=%s token_present=%v started=%v",
			r.Code, client.Name, client.ReconnectPlayerID, client.ReconnectToken != "", r.Started)
		if r.Started {
			client.SendMessage(WSMessage{
				Type:    "room",
				Payload: mustMarshal(RoomEvent{Action: "error", Message: "游戏已开始，且重连校验失败，无法加入"}),
			})
			return
		}
	} else if r.tryReconnectByNameLocked(client) {
		return
	}

	if r.Started {
		client.SendMessage(WSMessage{
			Type:    "room",
			Payload: mustMarshal(RoomEvent{Action: "error", Message: "游戏已开始，无法加入"}),
		})
		return
	}

	if len(r.Clients) >= 6 {
		client.SendMessage(WSMessage{
			Type:    "room",
			Payload: mustMarshal(RoomEvent{Action: "error", Message: "房间已满"}),
		})
		return
	}

	// Assign player ID (camp 由玩家自己选择)
	playerID, err := r.nextAvailablePlayerIDLocked()
	if err != nil {
		client.SendMessage(WSMessage{
			Type:    "room",
			Payload: mustMarshal(RoomEvent{Action: "error", Message: err.Error()}),
		})
		return
	}
	client.PlayerID = playerID
	client.Camp = model.Camp("") // 空表示未选择阵营

	// 创建房间后默认不分配角色，等待玩家主动选择
	client.CharRole = ""
	client.IsBot = false
	client.BotMode = ""
	client.Disconnected = false
	client.ReconnectToken = generateReconnectToken()

	client.Room = r
	r.Clients[client.PlayerID] = client
	if r.HostID == "" {
		r.HostID = client.PlayerID
	}

	// Send assignment to joining player（含角色数据供前端展示）
	client.SendMessage(WSMessage{
		Type: "room",
		Payload: mustMarshal(RoomEvent{
			Action:         "assigned",
			RoomCode:       r.Code,
			PlayerID:       client.PlayerID,
			Camp:           string(client.Camp),
			CharRole:       client.CharRole,
			Characters:     buildCharacterViews(),
			Message:        map[bool]string{true: "你是房主", false: ""}[r.isHost(client)],
			ReconnectToken: client.ReconnectToken,
		}),
	})

	// Broadcast updated player list to all
	r.broadcastPlayerList()

	log.Printf("Player %s (%s) joined room %s", client.Name, client.PlayerID, r.Code)
}

func (r *Room) handleUnregister(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.Clients[client.PlayerID]
	if !ok {
		return
	}
	if existing != client {
		return
	}

	// 开局后真人离线：保留席位等待重连，不再自动切换为机器人。
	if r.Started && existing != nil && !existing.IsBot {
		oldSend := existing.Send
		existing.Disconnected = true
		// 更换发送队列，避免旧 WritePump 退出时影响离线席位后续状态广播。
		existing.Send = make(chan []byte, 256)
		safeCloseBytesChan(oldSend)

		if r.HostID == client.PlayerID {
			r.HostID = ""
			r.ensureHostLocked()
		}

		r.broadcastRoomEvent(RoomEvent{
			Action:     "left",
			RoomCode:   r.Code,
			PlayerID:   client.PlayerID,
			PlayerName: client.Name,
			Message:    fmt.Sprintf("%s 离线，可通过房间号+玩家名重连；房主可选择是否启用机器人托管", client.Name),
		})
		r.broadcastPlayerList()
		log.Printf("Player %s disconnected and kept reconnectable seat in room %s", client.Name, r.Code)
	} else {
		delete(r.Clients, client.PlayerID)
		delete(r.botPromptCache, client.PlayerID)
		safeCloseBytesChan(existing.Send)

		if r.HostID == client.PlayerID {
			r.HostID = ""
			r.ensureHostLocked()
		}

		// Broadcast player left
		r.broadcastRoomEvent(RoomEvent{
			Action:     "left",
			RoomCode:   r.Code,
			PlayerID:   client.PlayerID,
			PlayerName: client.Name,
		})

		r.broadcastPlayerList()
		log.Printf("Player %s left room %s", client.Name, r.Code)
	}
}

func (r *Room) tryReconnectLocked(client *Client) bool {
	if client == nil {
		log.Printf("Reconnect failed in room %s: client=nil", r.Code)
		return false
	}
	if client.ReconnectPlayerID == "" || client.ReconnectToken == "" {
		log.Printf("Reconnect failed in room %s: missing params player_id=%q token_present=%v",
			r.Code, client.ReconnectPlayerID, client.ReconnectToken != "")
		return false
	}
	existing, ok := r.Clients[client.ReconnectPlayerID]
	if !ok || existing == nil {
		log.Printf("Reconnect failed in room %s: target player %s not found", r.Code, client.ReconnectPlayerID)
		return false
	}
	if existing.ReconnectToken == "" || existing.ReconnectToken != client.ReconnectToken {
		log.Printf("Reconnect failed in room %s: token mismatch for player %s", r.Code, client.ReconnectPlayerID)
		return false
	}
	return r.reconnectIntoSeatLocked(client, existing, "重连成功")
}

func (r *Room) tryReconnectByNameLocked(client *Client) bool {
	if client == nil {
		return false
	}
	name := strings.TrimSpace(client.Name)
	if name == "" {
		return false
	}

	var matched *Client
	for _, c := range r.Clients {
		if c == nil {
			continue
		}
		if !strings.EqualFold(strings.TrimSpace(c.Name), name) {
			continue
		}
		// 允许认领：
		// 1) 离线的人类席位；2) 房主手动切换出来的托管席位。
		if c.IsBot {
			if c.BotMode != "takeover" {
				continue
			}
		} else if !c.Disconnected {
			continue
		}
		if matched != nil {
			log.Printf("Reconnect by name rejected in room %s: ambiguous name=%s", r.Code, name)
			return false
		}
		matched = c
	}
	if matched == nil {
		return false
	}
	return r.reconnectIntoSeatLocked(client, matched, "通过房间码+玩家名重连成功")
}

func (r *Room) tryReconnectByPlayerIDLocked(client *Client) bool {
	if client == nil {
		return false
	}
	playerID := strings.TrimSpace(client.ReconnectPlayerID)
	if playerID == "" {
		return false
	}
	existing, ok := r.Clients[playerID]
	if !ok || existing == nil {
		log.Printf("Reconnect by player_id failed in room %s: target player %s not found", r.Code, playerID)
		return false
	}

	// 允许认领：
	// 1) 离线的人类席位；2) 房主手动切换出来的托管席位。
	if existing.IsBot {
		if existing.BotMode != "takeover" {
			log.Printf("Reconnect by player_id failed in room %s: target player %s is non-takeover bot", r.Code, playerID)
			return false
		}
	} else if !existing.Disconnected {
		log.Printf("Reconnect by player_id failed in room %s: target player %s still online", r.Code, playerID)
		return false
	}

	if client.Name != "" && !strings.EqualFold(strings.TrimSpace(client.Name), strings.TrimSpace(existing.Name)) {
		log.Printf("Reconnect by player_id accepted in room %s: incoming_name=%s, seat_name=%s, player_id=%s",
			r.Code, client.Name, existing.Name, playerID)
	}
	return r.reconnectIntoSeatLocked(client, existing, "通过房间码+玩家ID重连成功")
}

func (r *Room) reconnectIntoSeatLocked(client *Client, existing *Client, successMessage string) bool {
	if client == nil || existing == nil {
		return false
	}

	// 关闭旧连接/旧发送队列（离线占位或托管席位都在这里统一替换）。
	if existing.Conn != nil {
		_ = existing.Conn.Close()
	}
	safeCloseBytesChan(existing.Send)

	client.PlayerID = existing.PlayerID
	client.Name = existing.Name
	client.Camp = existing.Camp
	client.CharRole = existing.CharRole
	client.IsBot = false
	client.BotMode = ""
	client.Disconnected = false
	client.Room = r
	client.ReconnectToken = generateReconnectToken()

	r.Clients[client.PlayerID] = client
	cachedPrompt := r.botPromptCache[client.PlayerID]
	delete(r.botPromptCache, client.PlayerID)

	// 发送重连分配信息
	client.SendMessage(WSMessage{
		Type: "room",
		Payload: mustMarshal(RoomEvent{
			Action:         "assigned",
			RoomCode:       r.Code,
			PlayerID:       client.PlayerID,
			Camp:           string(client.Camp),
			CharRole:       client.CharRole,
			Characters:     buildCharacterViews(),
			Message:        successMessage,
			ReconnectToken: client.ReconnectToken,
		}),
	})

	// 立即补发当前状态与提示
	if r.Engine != nil {
		r.engineMu.Lock()
		stateView := r.buildStateForPlayer(client.PlayerID)
		var prompt *model.Prompt
		if p := r.Engine.GetCurrentPrompt(); p != nil && p.PlayerID == client.PlayerID {
			prompt = clonePrompt(p)
		} else if cachedPrompt != nil {
			prompt = clonePrompt(cachedPrompt)
		}
		r.engineMu.Unlock()

		client.SendMessage(WSMessage{
			Type: "event",
			Payload: mustMarshal(map[string]interface{}{
				"event_type": "state_update",
				"state":      stateView,
			}),
		})
		if prompt != nil {
			client.SendMessage(WSMessage{
				Type: "event",
				Payload: mustMarshal(map[string]interface{}{
					"event_type": "prompt",
					"prompt":     prompt,
				}),
			})
		}
	}

	// 通知其他玩家
	r.broadcastRoomEvent(RoomEvent{
		Action:     "joined",
		RoomCode:   r.Code,
		PlayerID:   client.PlayerID,
		PlayerName: client.Name,
		Message:    fmt.Sprintf("%s 重新连接", client.Name),
	})
	r.broadcastPlayerList()

	log.Printf("Player %s (%s) reconnected to room %s", client.Name, client.PlayerID, r.Code)
	return true
}

func isValidRole(role string) bool {
	for _, r := range availableRoles {
		if r == role {
			return true
		}
	}
	return false
}

func (r *Room) campCount(camp model.Camp) int {
	n := 0
	for _, c := range r.Clients {
		if c.Camp == camp {
			n++
		}
	}
	return n
}

func (r *Room) validateLineupLocked() error {
	if len(r.Clients) < 2 {
		return fmt.Errorf("至少需要2名玩家才能开始")
	}
	redN := 0
	blueN := 0
	roleOwners := make(map[string]string, len(r.Clients))
	for _, c := range r.Clients {
		if c.Camp == "" {
			return fmt.Errorf("有人未选择阵营")
		}
		if c.CharRole == "" {
			return fmt.Errorf("有人未选择角色")
		}
		if owner, ok := roleOwners[c.CharRole]; ok {
			return fmt.Errorf("角色不可重复：%s 与 %s 都选择了 %s", owner, c.Name, c.CharRole)
		}
		roleOwners[c.CharRole] = c.Name
		switch c.Camp {
		case model.RedCamp:
			redN++
		case model.BlueCamp:
			blueN++
		default:
			return fmt.Errorf("存在无效阵营配置")
		}
	}
	if redN < 1 || redN > 3 {
		return fmt.Errorf("红队需1-3人，当前%d人", redN)
	}
	if blueN < 1 || blueN > 3 {
		return fmt.Errorf("蓝队需1-3人，当前%d人", blueN)
	}
	return nil
}

func (r *Room) canAutoStartLocked() bool {
	return r.validateLineupLocked() == nil
}

func (r *Room) isHost(client *Client) bool {
	return client != nil && client.PlayerID != "" && client.PlayerID == r.HostID
}

func (r *Room) nextAvailablePlayerIDLocked() (string, error) {
	for i := 1; i <= 6; i++ {
		pid := fmt.Sprintf("p%d", i)
		if _, exists := r.Clients[pid]; !exists {
			return pid, nil
		}
	}
	return "", fmt.Errorf("房间已满")
}

func (r *Room) ensureHostLocked() {
	// 当前host仍在线且非机器人，则保持不变
	if host, ok := r.Clients[r.HostID]; ok && host != nil && !host.IsBot && !host.Disconnected {
		return
	}
	// 优先选在线真人玩家作为新host
	for _, c := range r.Clients {
		if c != nil && !c.IsBot && !c.Disconnected {
			r.HostID = c.PlayerID
			return
		}
	}
	// 兜底：如果当前都离线，保留任意真人席位作为 host（便于其重连后恢复管理权）。
	for _, c := range r.Clients {
		if c != nil && !c.IsBot {
			r.HostID = c.PlayerID
			return
		}
	}
	// 最终兜底：没有真人时清空 host。
	r.HostID = ""
}

func safeCloseBytesChan(ch chan []byte) {
	if ch == nil {
		return
	}
	defer func() {
		_ = recover()
	}()
	close(ch)
}

func (r *Room) broadcastPlayerList() {
	r.ensureHostLocked()

	var players []PlayerInfo
	for _, pid := range r.orderedClientIDsLocked() {
		c := r.Clients[pid]
		if c == nil {
			continue
		}
		players = append(players, PlayerInfo{
			ID:       c.PlayerID,
			Name:     c.Name,
			Camp:     string(c.Camp),
			CharRole: c.CharRole,
			Ready:    c.Camp != "" && c.CharRole != "",
			IsOnline: c.IsBot || !c.Disconnected,
			IsBot:    c.IsBot,
			IsHost:   c.PlayerID == r.HostID,
			BotMode:  c.BotMode,
		})
	}

	r.broadcastRoomEvent(RoomEvent{
		Action:     "player_list",
		RoomCode:   r.Code,
		Players:    players,
		Characters: buildCharacterViews(),
	})
}

func (r *Room) orderedClientIDsLocked() []string {
	if len(r.Clients) == 0 {
		return nil
	}

	ids := make([]string, 0, len(r.Clients))
	seen := make(map[string]struct{}, len(r.Clients))

	// 开局后优先按固定座次顺序。
	for _, pid := range r.SeatOrder {
		if _, ok := r.Clients[pid]; !ok {
			continue
		}
		ids = append(ids, pid)
		seen[pid] = struct{}{}
	}

	// 未进入 SeatOrder 的玩家（未开局或中途补位）按 pid 稳定排序，避免 map 遍历抖动。
	rest := make([]string, 0, len(r.Clients)-len(ids))
	for pid := range r.Clients {
		if _, ok := seen[pid]; ok {
			continue
		}
		rest = append(rest, pid)
	}
	sort.Strings(rest)
	ids = append(ids, rest...)
	return ids
}

func buildInterleavedLineup(players []lineupPlayer, r *rand.Rand) []lineupPlayer {
	if len(players) <= 2 || r == nil {
		out := make([]lineupPlayer, len(players))
		copy(out, players)
		return out
	}

	red := make([]lineupPlayer, 0, len(players))
	blue := make([]lineupPlayer, 0, len(players))
	other := make([]lineupPlayer, 0, len(players))
	for _, p := range players {
		switch p.camp {
		case model.RedCamp:
			red = append(red, p)
		case model.BlueCamp:
			blue = append(blue, p)
		default:
			other = append(other, p)
		}
	}

	r.Shuffle(len(red), func(i, j int) { red[i], red[j] = red[j], red[i] })
	r.Shuffle(len(blue), func(i, j int) { blue[i], blue[j] = blue[j], blue[i] })
	r.Shuffle(len(other), func(i, j int) { other[i], other[j] = other[j], other[i] })

	first := red
	second := blue
	if len(blue) > len(red) || (len(blue) == len(red) && len(blue) > 0 && r.Intn(2) == 1) {
		first, second = blue, red
	}

	result := make([]lineupPlayer, 0, len(players))
	for len(first) > 0 || len(second) > 0 {
		if len(first) > 0 {
			result = append(result, first[0])
			first = first[1:]
		}
		if len(second) > 0 {
			result = append(result, second[0])
			second = second[1:]
		}
	}
	result = append(result, other...)
	return result
}

// buildCharacterViews 从 data 包构建角色视图
func buildCharacterViews() []CharacterView {
	chars := data.GetCharacters()
	views := make([]CharacterView, 0, len(chars))
	for _, c := range chars {
		skills := make([]SkillView, 0, len(c.Skills))
		for _, s := range c.Skills {
			skills = append(skills, SkillView{
				ID:               s.ID,
				Title:            s.Title,
				Description:      s.Description,
				Type:             int(s.Type),
				MinTargets:       s.MinTargets,
				MaxTargets:       s.MaxTargets,
				TargetType:       int(s.TargetType),
				CostGem:          s.CostGem,
				CostCrystal:      s.CostCrystal,
				CostDiscards:     s.CostDiscards,
				DiscardElement:   string(s.DiscardElement),
				RequireExclusive: s.RequireExclusive,
			})
		}
		views = append(views, CharacterView{
			ID:      c.ID,
			Name:    c.Name,
			Title:   c.Title,
			Faction: c.Faction,
			Skills:  skills,
		})
	}
	return views
}

func (r *Room) broadcastRoomEvent(event RoomEvent) {
	msg := WSMessage{
		Type:    "room",
		Payload: mustMarshal(event),
	}
	data, _ := json.Marshal(msg)
	for _, c := range r.Clients {
		if c.IsBot || c.Disconnected {
			continue
		}
		select {
		case c.Send <- data:
		default:
		}
	}
}

func (r *Room) broadcastToAll(message []byte) {
	for _, c := range r.Clients {
		if c.IsBot || c.Disconnected {
			continue
		}
		select {
		case c.Send <- message:
		default:
		}
	}
}

func (r *Room) startGame() error {
	lineup := make([]lineupPlayer, 0, 6)

	r.mu.Lock()
	if r.Started {
		r.mu.Unlock()
		return fmt.Errorf("游戏已开始")
	}
	if err := r.validateLineupLocked(); err != nil {
		r.mu.Unlock()
		return err
	}
	for _, client := range r.Clients {
		lineup = append(lineup, lineupPlayer{
			id:   client.PlayerID,
			name: client.Name,
			role: client.CharRole,
			camp: client.Camp,
		})
	}
	seed := time.Now().UnixNano()
	lineup = buildInterleavedLineup(lineup, rand.New(rand.NewSource(seed)))
	r.SeatOrder = r.SeatOrder[:0]
	for _, p := range lineup {
		r.SeatOrder = append(r.SeatOrder, p.id)
	}
	r.Started = true
	// 同步广播固定座次，避免前端仍沿用 lobby 随机顺序导致同阵营扎堆显示。
	r.broadcastPlayerList()
	r.mu.Unlock()

	// Create game engine with this room as observer
	r.Engine = engine.NewGameEngine(r)

	// Add all players to the engine
	for _, player := range lineup {
		err := r.Engine.AddPlayer(player.id, player.name, player.role, player.camp)
		if err != nil {
			log.Printf("Error adding player: %v", err)
		}
	}

	// Broadcast game started（含角色数据供前端技能 fallback）
	r.broadcastRoomEvent(RoomEvent{
		Action:     "started",
		RoomCode:   r.Code,
		Message:    "游戏开始！",
		Characters: buildCharacterViews(),
	})

	// Start the game
	r.engineMu.Lock()
	if err := r.Engine.StartGame(); err != nil {
		r.engineMu.Unlock()
		log.Printf("Error starting game: %v", err)
		r.mu.Lock()
		r.Started = false
		r.mu.Unlock()
		r.Engine = nil
		r.broadcastRoomEvent(RoomEvent{
			Action:  "error",
			Message: fmt.Sprintf("游戏启动失败: %v", err),
		})
		return nil
	}
	// Drive game loop
	r.Engine.Drive()
	r.engineMu.Unlock()

	// 若首个操作者是机器人，自动驱动其行动
	go r.scheduleAnyBotIfPrompt()
	return nil
}

func (r *Room) autoStartIfReady() {
	r.mu.RLock()
	if r.Started || !r.canAutoStartLocked() {
		r.mu.RUnlock()
		return
	}
	r.mu.RUnlock()

	if err := r.startGame(); err != nil {
		if err.Error() == "游戏已开始" {
			return
		}
		r.broadcastRoomEvent(RoomEvent{
			Action:  "error",
			Message: fmt.Sprintf("自动开始失败: %v", err),
		})
	}
}

// OnGameEvent implements model.GameObserver
func (r *Room) OnGameEvent(event model.GameEvent) {
	var botPromptPlayerID string
	var botPrompt *model.Prompt
	var botPromptEpoch uint64

	switch event.Type {
	case model.EventLog:
		// Broadcast log to all players
		r.broadcastGameEvent("log", map[string]interface{}{
			"message": event.Message,
		})

	case model.EventStateUpdate:
		// Send filtered state to each player
		r.mu.RLock()
		for _, client := range r.Clients {
			if client.IsBot || client.Disconnected {
				continue
			}
			stateView := r.buildStateForPlayer(client.PlayerID)
			client.SendMessage(WSMessage{
				Type: "event",
				Payload: mustMarshal(map[string]interface{}{
					"event_type": "state_update",
					"state":      stateView,
				}),
			})
		}
		r.mu.RUnlock()

	case model.EventAskInput:
		var prompt *model.Prompt
		switch p := event.Data.(type) {
		case *model.Prompt:
			prompt = p
		case model.Prompt:
			cp := p
			prompt = &cp
		}
		if prompt != nil {
			// 先推送状态给所有人（含手牌），确保客户端有最新数据
			r.mu.Lock()
			r.botPromptEpoch++
			botPromptEpoch = r.botPromptEpoch
			// 一次 AskInput 仅有一个有效提示，清空旧缓存避免旧定时器误动作。
			r.botPromptCache = map[string]*model.Prompt{
				prompt.PlayerID: clonePrompt(prompt),
			}
			for _, c := range r.Clients {
				if c.IsBot || c.Disconnected {
					continue
				}
				stateView := r.buildStateForPlayer(c.PlayerID)
				c.SendMessage(WSMessage{
					Type: "event",
					Payload: mustMarshal(map[string]interface{}{
						"event_type": "state_update",
						"state":      stateView,
					}),
				})
			}
			// Send prompt only to the target player
			if client, exists := r.Clients[prompt.PlayerID]; exists {
				if client.IsBot {
					botPromptPlayerID = client.PlayerID
					botPrompt = clonePrompt(prompt)
				} else if client.Disconnected {
					// 真人离线且未托管：暂停等待重连或房主手动托管。
				} else {
					client.SendMessage(WSMessage{
						Type: "event",
						Payload: mustMarshal(map[string]interface{}{
							"event_type": "prompt",
							"prompt":     prompt,
						}),
					})
				}
			}
			// Notify other players that someone is being prompted
			for pid, client := range r.Clients {
				if client.IsBot || client.Disconnected {
					continue
				}
				if pid != prompt.PlayerID {
					client.SendMessage(WSMessage{
						Type: "event",
						Payload: mustMarshal(map[string]interface{}{
							"event_type": "waiting",
							"player_id":  prompt.PlayerID,
							"message":    fmt.Sprintf("等待 %s 操作...", prompt.PlayerID),
						}),
					})
				}
			}
			r.mu.Unlock()
		}

	case model.EventError:
		r.broadcastGameEvent("error", map[string]interface{}{
			"message": event.Message,
		})

	case model.EventGameEnd:
		r.broadcastGameEvent("game_end", map[string]interface{}{
			"message": event.Message,
		})

	case model.EventCardRevealed:
		if data, ok := event.Data.(map[string]interface{}); ok {
			r.botIntel.observeReveal(data)
			r.broadcastGameEvent("card_revealed", data)
		}

	case model.EventDamageDealt:
		if data, ok := event.Data.(map[string]interface{}); ok {
			r.broadcastGameEvent("damage_dealt", data)
		}

	case model.EventActionStep:
		if data, ok := event.Data.(map[string]interface{}); ok {
			r.broadcastGameEvent("action_step", data)
		}

	case model.EventCombatCue:
		if data, ok := event.Data.(map[string]interface{}); ok {
			r.broadcastGameEvent("combat_cue", data)
		}

	case model.EventDrawCards:
		if data, ok := event.Data.(map[string]interface{}); ok {
			r.broadcastGameEvent("draw_cards", data)
		}
	}

	if botPromptPlayerID != "" {
		go r.scheduleBotIfNeeded(botPromptPlayerID, botPrompt, botPromptEpoch)
	}
}

func (r *Room) broadcastGameEvent(eventType string, data map[string]interface{}) {
	data["event_type"] = eventType
	msg := WSMessage{
		Type:    "event",
		Payload: mustMarshal(data),
	}
	msgData, _ := json.Marshal(msg)
	for _, c := range r.Clients {
		if c.IsBot || c.Disconnected {
			continue
		}
		select {
		case c.Send <- msgData:
		default:
		}
	}
}

func countMagicBowChargesByElement(p *model.Player, element model.Element) int {
	if p == nil {
		return 0
	}
	count := 0
	for _, fc := range p.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectMagicBowCharge {
			continue
		}
		if element != "" && fc.Card.Element != element {
			continue
		}
		count++
	}
	return count
}

func countMagicBowCharges(p *model.Player) int {
	return countMagicBowChargesByElement(p, "")
}

func countSpiritCasterPowers(p *model.Player) int {
	if p == nil {
		return 0
	}
	count := 0
	for _, fc := range p.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectSpiritCasterPower {
			continue
		}
		count++
	}
	return count
}

func countMoonDarkMoons(p *model.Player) int {
	if p == nil {
		return 0
	}
	count := 0
	for _, fc := range p.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectMoonDarkMoon {
			continue
		}
		count++
	}
	return count
}

func countButterflyCocoons(p *model.Player) int {
	if p == nil {
		return 0
	}
	count := 0
	for _, fc := range p.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectButterflyCocoon {
			continue
		}
		count++
	}
	return count
}

func countBloodSharedLifeAsSource(state *model.GameState, sourceID string) int {
	if state == nil || sourceID == "" {
		return 0
	}
	count := 0
	for _, p := range state.Players {
		if p == nil {
			continue
		}
		for _, fc := range p.Field {
			if fc == nil || fc.Mode != model.FieldEffect || fc.Effect != model.EffectBloodSharedLife {
				continue
			}
			if fc.SourceID == sourceID {
				count++
			}
		}
	}
	return count
}

func countBloodSharedLifeAsHolder(player *model.Player) int {
	if player == nil {
		return 0
	}
	count := 0
	for _, fc := range player.Field {
		if fc == nil || fc.Mode != model.FieldEffect || fc.Effect != model.EffectBloodSharedLife {
			continue
		}
		count++
	}
	return count
}

// 当前手牌上限仅用于展示，避免直接对引擎内玩家对象调用 GetMaxHand 带来状态副作用。
func (r *Room) previewMaxHand(player *model.Player) int {
	if player == nil {
		return 0
	}
	playerCopy := *player
	if player.Tokens != nil {
		playerCopy.Tokens = make(map[string]int, len(player.Tokens))
		for k, v := range player.Tokens {
			playerCopy.Tokens[k] = v
		}
	}
	return r.Engine.GetMaxHand(&playerCopy)
}

func buildMaskedFieldForViewer(owner *model.Player, viewerID string) []*model.FieldCard {
	if owner == nil || len(owner.Field) == 0 {
		return nil
	}
	out := make([]*model.FieldCard, 0, len(owner.Field))
	for _, fc := range owner.Field {
		if fc == nil {
			continue
		}
		clone := *fc
		// 魔弓“充能”、灵符师“妖力”、月之女神“暗月”、蝶舞者“茧”对非持有者隐藏具体牌面信息，仅保留数量与盖牌属性。
		if owner.ID != viewerID && clone.Mode == model.FieldCover &&
			(clone.Effect == model.EffectMagicBowCharge || clone.Effect == model.EffectSpiritCasterPower || clone.Effect == model.EffectMoonDarkMoon || clone.Effect == model.EffectButterflyCocoon) {
			maskedName := "盖牌"
			if clone.Effect == model.EffectMagicBowCharge {
				maskedName = "充能"
			} else if clone.Effect == model.EffectSpiritCasterPower {
				maskedName = "妖力"
			} else if clone.Effect == model.EffectMoonDarkMoon {
				maskedName = "暗月"
			} else if clone.Effect == model.EffectButterflyCocoon {
				maskedName = "茧"
			}
			clone.Card = model.Card{
				ID:          clone.Card.ID,
				Name:        maskedName,
				Type:        clone.Card.Type,
				Description: "盖牌（内容对他人不可见）",
			}
		}
		out = append(out, &clone)
	}
	return out
}

func (r *Room) buildStateForPlayer(playerID string) GameStateUpdate {
	state := r.Engine.State

	players := make(map[string]PlayerView)
	for pid, p := range state.Players {
		view := PlayerView{
			ID:                 p.ID,
			Name:               p.Name,
			Camp:               string(p.Camp),
			Role:               p.Role,
			HandCount:          len(p.Hand),
			MaxHand:            r.previewMaxHand(p),
			ExclusiveCardCount: len(p.ExclusiveCards),
			Field:              buildMaskedFieldForViewer(p, playerID),
			Heal:               p.Heal,
			MaxHeal:            p.MaxHeal,
			Gem:                p.Gem,
			Crystal:            p.Crystal,
			IsActive:           p.IsActive,
			Buffs:              p.Buffs,
			Tokens:             map[string]int{},
		}
		for k, v := range p.Tokens {
			view.Tokens[k] = v
		}
		delete(view.Tokens, "adventurer_extract_last_gem")
		delete(view.Tokens, "adventurer_extract_last_crystal")
		delete(view.Tokens, "mg_moon_cycle_used_turn")
		// 魔枪公开状态：下次主动攻击加成/当回合互斥锁，便于前端角色面板展示。
		if p.TurnState.UsedSkillCounts != nil {
			if v := p.TurnState.UsedSkillCounts["ml_dark_release_next_attack_bonus"]; v > 0 {
				view.Tokens["ml_dark_release_next_attack_bonus"] = v
			}
			if v := p.TurnState.UsedSkillCounts["ml_fullness_next_attack_bonus"]; v > 0 {
				view.Tokens["ml_fullness_next_attack_bonus"] = v
			}
			if v := p.TurnState.UsedSkillCounts["ml_dark_release_lock_turn"]; v > 0 {
				view.Tokens["ml_dark_release_lock_turn"] = v
			}
		}
		// 精灵射手祝福数量（独立牌区）透出给前端做指示物展示。
		blessings := len(p.Blessings)
		if blessings > 0 {
			view.Tokens["elf_blessing_count"] = blessings
		}
		// 魔弓充能数量（盖牌内容对他人隐藏，仅展示数量）。
		chargeCount := countMagicBowCharges(p)
		if chargeCount > 0 {
			view.Tokens["mb_charge_count"] = chargeCount
		} else {
			delete(view.Tokens, "mb_charge_count")
		}
		// 灵符师妖力数量（盖牌内容对他人隐藏，仅展示数量）。
		powerCount := countSpiritCasterPowers(p)
		if powerCount > 0 {
			view.Tokens["sc_power_count"] = powerCount
		} else {
			delete(view.Tokens, "sc_power_count")
		}
		// 月之女神暗月数量（盖牌内容对他人隐藏，仅展示数量）。
		darkMoonCount := countMoonDarkMoons(p)
		if darkMoonCount > 0 {
			view.Tokens["mg_dark_moon_count"] = darkMoonCount
		} else {
			delete(view.Tokens, "mg_dark_moon_count")
		}
		// 血之巫女同生共死：仅展示是否存在激活中的连结。
		sharedLifeCount := countBloodSharedLifeAsSource(state, p.ID)
		if sharedLifeCount > 0 {
			view.Tokens["bp_shared_life_active"] = sharedLifeCount
		} else {
			delete(view.Tokens, "bp_shared_life_active")
		}
		sharedLifeBoundCount := countBloodSharedLifeAsHolder(p)
		if sharedLifeBoundCount > 0 {
			view.Tokens["bp_shared_life_bound"] = sharedLifeBoundCount
		} else {
			delete(view.Tokens, "bp_shared_life_bound")
		}
		// 蝶舞者茧数量（盖牌内容对他人隐藏，仅展示数量）。
		cocoonCount := countButterflyCocoons(p)
		if cocoonCount > 0 {
			view.Tokens["bt_cocoon_count"] = cocoonCount
		} else {
			delete(view.Tokens, "bt_cocoon_count")
		}
		// 仅自己可见手牌具体内容，他人只能看到数量
		if pid == playerID {
			view.Hand = p.Hand
			view.Blessings = p.Blessings
			view.ExclusiveCards = p.ExclusiveCards
		}
		players[pid] = view
	}

	var availableSkills []AvailableSkill
	if state.Phase == model.PhaseActionSelection {
		if self := state.Players[playerID]; self != nil && self.IsActive {
			availableSkills = r.buildAvailableActionSkills(playerID)
		}
	}

	return GameStateUpdate{
		Phase:               string(state.Phase),
		CurrentPlayer:       state.CurrentPlayer,
		HasPerformedStartup: state.HasPerformedStartup,
		Players:             players,
		RedMorale:           state.RedMorale,
		BlueMorale:          state.BlueMorale,
		RedCups:             state.RedCups,
		BlueCups:            state.BlueCups,
		RedGems:             state.RedGems,
		BlueGems:            state.BlueGems,
		RedCrystals:         state.RedCrystals,
		BlueCrystals:        state.BlueCrystals,
		DeckCount:           len(state.Deck),
		DiscardCount:        len(state.DiscardPile),
		AvailableSkills:     availableSkills,
		Characters:          buildCharacterViews(),
	}
}

// buildAvailableActionSkills 返回当前玩家可发动的主动技能列表（用于前端按钮可用态）。
func (r *Room) buildAvailableActionSkills(playerID string) []AvailableSkill {
	p := r.Engine.State.Players[playerID]
	if p == nil || p.Character == nil {
		return nil
	}
	var list []AvailableSkill
	for _, sd := range p.Character.Skills {
		if sd.Type != model.SkillTypeAction {
			continue
		}
		if sd.ID == "adventurer_fraud" {
			// 欺诈：至少满足其一
			// 1) 有2张同系可用于弃牌
			// 2) 有3张同系可作为暗灭攻击
			elemCount := map[model.Element]int{}
			for _, c := range p.Hand {
				elemCount[c.Element]++
			}
			canUseFraud := false
			for ele, n := range elemCount {
				if ele != "" && n >= 2 {
					canUseFraud = true
					break
				}
				if n >= 3 {
					canUseFraud = true
					break
				}
			}
			if !canUseFraud {
				continue
			}
		}
		if sd.ID == "onmyoji_shikigami_descend" {
			factionCount := map[string]int{}
			hasSameFactionPair := false
			for _, c := range p.Hand {
				if c.Faction == "" {
					continue
				}
				factionCount[c.Faction]++
				if factionCount[c.Faction] >= 2 {
					hasSameFactionPair = true
					break
				}
			}
			if !hasSameFactionPair {
				continue
			}
		}
		if sd.ID == "mb_thunder_scatter" {
			if p.TurnState.UsedSkillCounts["mb_charge_lock_turn"] > 0 {
				continue
			}
			if countMagicBowChargesByElement(p, model.ElementThunder) <= 0 {
				continue
			}
		}
		if sd.ID == "bd_dissonance_chord" {
			inspiration := 0
			if p.Tokens != nil {
				inspiration = p.Tokens["bd_inspiration"]
			}
			if inspiration <= 1 {
				continue
			}
		}
		if sd.ID == "elementalist_ignite" {
			element := 0
			if p.Tokens != nil {
				element = p.Tokens["element"]
			}
			if element < 3 {
				continue
			}
		}
		// 回合限定：本回合已用过则不再展示
		if model.ContainsSkillTag(sd.Tags, model.TagTurnLimit) {
			if p.TurnState.UsedSkillCounts[sd.ID] > 0 {
				continue
			}
		}
		// 必杀技资源过滤规则：
		// - 宝石消耗必须由宝石支付
		// - 水晶消耗可由“剩余宝石”替代
		if sd.CostGem > 0 || sd.CostCrystal > 0 {
			if p.Gem < sd.CostGem {
				continue
			}
			usableCrystal := p.Crystal + (p.Gem - sd.CostGem)
			if usableCrystal < sd.CostCrystal {
				continue
			}
		}
		// 独有技：必须拥有对应独有牌（手牌或专属卡区）才能使用
		if sd.RequireExclusive {
			if !p.HasExclusiveCard(p.Character.Name, sd.Title) {
				continue
			}
		}
		// 通用可用性兜底：复用技能 Handler 的 CanUse，提前过滤“指示物不足/形态不符”等条件。
		// 这样前端会直接把技能置灰（或不展示），避免点击后才报“技能发动失败”。
		if !r.canUseActionSkillNow(p, sd) {
			continue
		}
		list = append(list, AvailableSkill{
			ID:               sd.ID,
			Title:            sd.Title,
			Description:      sd.Description,
			MinTargets:       sd.MinTargets,
			MaxTargets:       sd.MaxTargets,
			TargetType:       int(sd.TargetType),
			CostGem:          sd.CostGem,
			CostCrystal:      sd.CostCrystal,
			CostDiscards:     sd.CostDiscards,
			DiscardType:      string(sd.DiscardType),
			DiscardElement:   string(sd.DiscardElement),
			RequireExclusive: sd.RequireExclusive,
			PlaceCard:        sd.PlaceCard,
			PlaceEffect:      string(sd.PlaceEffect),
		})
	}
	return list
}

func (r *Room) probeTargetForActionSkill(user *model.Player, targetType model.TargetType) *model.Player {
	if r == nil || r.Engine == nil || user == nil {
		return nil
	}
	switch targetType {
	case model.TargetSelf:
		return user
	case model.TargetEnemy:
		for _, p := range r.Engine.State.Players {
			if p != nil && p.Camp != user.Camp {
				return p
			}
		}
	case model.TargetAlly:
		for _, p := range r.Engine.State.Players {
			if p != nil && p.Camp == user.Camp && p.ID != user.ID {
				return p
			}
		}
	case model.TargetAllySelf:
		return user
	case model.TargetAny, model.TargetSpecific:
		return user
	}
	return nil
}

func (r *Room) canUseActionSkillNow(user *model.Player, sd model.SkillDefinition) bool {
	if r == nil || r.Engine == nil || user == nil {
		return false
	}
	if sd.LogicHandler == "" {
		return true
	}
	handler := skills.GetHandler(sd.LogicHandler)
	if handler == nil {
		return true
	}
	probeTarget := r.probeTargetForActionSkill(user, sd.TargetType)
	targetID := user.ID
	if probeTarget != nil {
		targetID = probeTarget.ID
	}
	ctx := &model.Context{
		Game:    r.Engine,
		User:    user,
		Target:  probeTarget,
		Trigger: model.TriggerNone,
		TriggerCtx: &model.EventContext{
			Type:     model.EventNone,
			SourceID: user.ID,
			TargetID: targetID,
		},
		Selections: map[string]any{},
		Flags:      map[string]bool{},
	}
	if probeTarget != nil {
		ctx.Targets = []*model.Player{probeTarget}
	}
	return handler.CanUse(ctx)
}

// HandleMessage processes incoming WebSocket messages
func (r *Room) HandleMessage(client *Client, msg *WSMessage) {
	switch msg.Type {
	case "action":
		r.handleAction(client, msg.Payload)
	case "chat":
		r.handleChat(client, msg.Payload)
	case "room":
		r.handleRoomAction(client, msg.Payload)
	}
}

func (r *Room) handleAction(client *Client, payload json.RawMessage) {
	if !r.Started || r.Engine == nil {
		client.SendMessage(WSMessage{
			Type:    "event",
			Payload: mustMarshal(map[string]interface{}{"event_type": "error", "message": "游戏尚未开始"}),
		})
		return
	}

	var action model.PlayerAction
	if err := json.Unmarshal(payload, &action); err != nil {
		log.Printf("Error parsing action: %v", err)
		return
	}

	// Ensure the action is from the correct player
	action.PlayerID = client.PlayerID

	// Handle the action
	if err := r.submitAction(action); err != nil {
		// 同步错误直接返回给当前客户端，保证 UI 可见
		client.SendMessage(WSMessage{
			Type: "event",
			Payload: mustMarshal(map[string]interface{}{
				"event_type": "error",
				"message":    err.Error(),
			}),
		})
		return
	}
}

func (r *Room) submitAction(action model.PlayerAction) error {
	r.engineMu.Lock()
	defer r.engineMu.Unlock()

	if !r.Started || r.Engine == nil {
		return fmt.Errorf("游戏尚未开始")
	}

	// 记录本次 action 前的提示版本号，用于判断是否产生了新的 AskInput。
	r.mu.RLock()
	promptEpochBefore := r.botPromptEpoch
	r.mu.RUnlock()

	if err := r.Engine.HandleAction(action); err != nil {
		return err
	}
	// 与现有流程保持一致：额外触发一次 Drive，确保提示及时刷新
	r.Engine.Drive()

	// 兜底状态同步：
	// - 若本次 action 没有产生新的 AskInput（prompt 版本未变化）
	// - 且当前没有挂起中断
	// 则主动推送一次 state_update，避免前端保留过期弹框导致“已操作但仍卡在中断提示”。
	r.mu.RLock()
	promptEpochAfter := r.botPromptEpoch
	r.mu.RUnlock()
	if promptEpochAfter == promptEpochBefore && r.Engine.State.PendingInterrupt == nil {
		r.OnGameEvent(model.GameEvent{Type: model.EventStateUpdate})
	}
	return nil
}

func (r *Room) handleChat(client *Client, payload json.RawMessage) {
	var chatMsg struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(payload, &chatMsg); err != nil {
		return
	}

	r.broadcastGameEvent("chat", map[string]interface{}{
		"player_id":   client.PlayerID,
		"player_name": client.Name,
		"message":     chatMsg.Message,
	})
}

func (r *Room) handleRoomAction(client *Client, payload json.RawMessage) {
	var roomAction struct {
		Action   string `json:"action"`
		Camp     string `json:"camp,omitempty"`
		CharRole string `json:"char_role,omitempty"`
		TargetID string `json:"target_id,omitempty"`
		BotName  string `json:"bot_name,omitempty"`
	}
	if err := json.Unmarshal(payload, &roomAction); err != nil {
		return
	}

	sendRoomError := func(msg string) {
		client.SendMessage(WSMessage{
			Type:    "room",
			Payload: mustMarshal(RoomEvent{Action: "error", Message: msg}),
		})
	}

	switch roomAction.Action {
	case "dissolve_room":
		var toClose []*Client
		var dissolveMsg string

		r.mu.Lock()
		if !r.isHost(client) {
			r.mu.Unlock()
			sendRoomError("仅房主可解散房间")
			return
		}
		dissolveMsg = fmt.Sprintf("房主 %s 已解散房间", client.Name)
		r.broadcastRoomEvent(RoomEvent{
			Action:     "dissolved",
			RoomCode:   r.Code,
			PlayerID:   client.PlayerID,
			PlayerName: client.Name,
			Message:    dissolveMsg,
		})
		for _, c := range r.Clients {
			if c != nil {
				toClose = append(toClose, c)
			}
		}
		r.Clients = make(map[string]*Client)
		r.SeatOrder = nil
		r.Started = false
		r.HostID = ""
		r.botPromptCache = make(map[string]*model.Prompt)
		r.botPromptEpoch++
		r.botIntel = newBotIntel()
		r.mu.Unlock()

		r.engineMu.Lock()
		r.Engine = nil
		r.engineMu.Unlock()

		for _, c := range toClose {
			safeCloseBytesChan(c.Send)
			safeCloseConn(c.connSnapshot())
		}
		log.Printf("Room %s dissolved by host %s", r.Code, client.Name)
		return

	case "add_bot":
		r.mu.Lock()
		defer r.mu.Unlock()

		if r.Started {
			sendRoomError("游戏已开始，无法添加机器人")
			return
		}
		if !r.isHost(client) {
			sendRoomError("仅房主可添加机器人")
			return
		}
		if len(r.Clients) >= 6 {
			sendRoomError("房间已满")
			return
		}
		pid, err := r.nextAvailablePlayerIDLocked()
		if err != nil {
			sendRoomError(err.Error())
			return
		}
		botName := roomAction.BotName
		if botName == "" {
			botName = fmt.Sprintf("机器人%s", pid)
		}
		bot := &Client{
			Room:     r,
			Send:     make(chan []byte, 256),
			PlayerID: pid,
			Name:     botName,
			Camp:     model.Camp(""),
			CharRole: "",
			IsBot:    true,
			BotMode:  "added",
		}
		r.Clients[pid] = bot
		r.broadcastPlayerList()
		return

	case "remove_bot":
		r.mu.Lock()
		defer r.mu.Unlock()

		if r.Started {
			sendRoomError("游戏开始后不能移除机器人席位")
			return
		}
		if !r.isHost(client) {
			sendRoomError("仅房主可移除机器人")
			return
		}
		targetID := roomAction.TargetID
		if targetID == "" {
			sendRoomError("缺少机器人ID")
			return
		}
		target, ok := r.Clients[targetID]
		if !ok || target == nil || !target.IsBot {
			sendRoomError("目标不是机器人")
			return
		}
		delete(r.Clients, targetID)
		delete(r.botPromptCache, targetID)
		safeCloseBytesChan(target.Send)
		r.broadcastPlayerList()
		return

	case "takeover_player":
		var takeoverBotID string
		r.mu.Lock()
		if !r.Started {
			r.mu.Unlock()
			sendRoomError("游戏未开始，无需托管")
			return
		}
		if !r.isHost(client) {
			r.mu.Unlock()
			sendRoomError("仅房主可启用托管")
			return
		}
		targetID := roomAction.TargetID
		if targetID == "" {
			r.mu.Unlock()
			sendRoomError("缺少目标玩家ID")
			return
		}
		target, ok := r.Clients[targetID]
		if !ok || target == nil {
			r.mu.Unlock()
			sendRoomError("目标玩家不存在")
			return
		}
		if target.IsBot {
			r.mu.Unlock()
			sendRoomError("目标已是机器人托管")
			return
		}
		if !target.Disconnected {
			r.mu.Unlock()
			sendRoomError("目标玩家当前在线，无需托管")
			return
		}

		oldSend := target.Send
		takeover := &Client{
			Room:           r,
			Send:           make(chan []byte, 256),
			PlayerID:       target.PlayerID,
			Name:           target.Name,
			Camp:           target.Camp,
			CharRole:       target.CharRole,
			IsBot:          true,
			BotMode:        "takeover",
			ReconnectToken: target.ReconnectToken,
		}
		r.Clients[targetID] = takeover
		takeoverBotID = targetID
		safeCloseBytesChan(oldSend)

		if r.HostID == targetID {
			r.HostID = ""
			r.ensureHostLocked()
		}

		r.broadcastRoomEvent(RoomEvent{
			Action:     "left",
			RoomCode:   r.Code,
			PlayerID:   targetID,
			PlayerName: target.Name,
			Message:    fmt.Sprintf("房主已将 %s 切换为机器人托管", target.Name),
		})
		r.broadcastPlayerList()
		r.mu.Unlock()

		if takeoverBotID != "" {
			go func(pid string) {
				time.Sleep(120 * time.Millisecond)
				r.scheduleBotIfNeeded(pid, nil, 0)
			}(takeoverBotID)
		}
		return

	case "change_camp":
		if r.Started {
			return
		}
		r.mu.Lock()
		defer r.mu.Unlock()

		target := client
		if roomAction.TargetID != "" && roomAction.TargetID != client.PlayerID {
			if !r.isHost(client) {
				sendRoomError("仅房主可调整机器人阵营")
				return
			}
			t, ok := r.Clients[roomAction.TargetID]
			if !ok || t == nil || !t.IsBot {
				sendRoomError("仅可调整机器人阵营")
				return
			}
			target = t
		}

		camp := model.Camp(roomAction.Camp)
		if camp != model.RedCamp && camp != model.BlueCamp {
			sendRoomError("无效阵营")
			return
		}
		if target.Camp == camp {
			return
		}
		if r.campCount(camp) >= 3 {
			sendRoomError("该阵营人数已满")
			return
		}
		target.Camp = camp
		r.broadcastPlayerList()
		go r.autoStartIfReady()
		return

	case "change_role":
		if r.Started || roomAction.CharRole == "" {
			return
		}
		r.mu.Lock()
		defer r.mu.Unlock()

		target := client
		if roomAction.TargetID != "" && roomAction.TargetID != client.PlayerID {
			if !r.isHost(client) {
				sendRoomError("仅房主可调整机器人角色")
				return
			}
			t, ok := r.Clients[roomAction.TargetID]
			if !ok || t == nil || !t.IsBot {
				sendRoomError("仅可调整机器人角色")
				return
			}
			target = t
		}

		if !isValidRole(roomAction.CharRole) {
			sendRoomError("无效角色")
			return
		}
		// Check if role is available
		for pid, c := range r.Clients {
			if pid != target.PlayerID && c.CharRole == roomAction.CharRole {
				sendRoomError("该角色已被其他玩家选择")
				return
			}
		}
		target.CharRole = roomAction.CharRole
		r.broadcastPlayerList()
		go r.autoStartIfReady()
		return

	case "start":
		r.mu.Lock()
		if r.Started {
			r.mu.Unlock()
			sendRoomError("游戏已开始")
		} else if len(r.Clients) < 2 {
			r.mu.Unlock()
			sendRoomError("至少需要2名玩家才能开始")
		} else {
			r.mu.Unlock()
			if err := r.startGame(); err != nil {
				sendRoomError(err.Error())
			}
		}
	}
}

func mustMarshal(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}
