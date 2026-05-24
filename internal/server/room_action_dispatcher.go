package server

import (
	"encoding/json"
	"fmt"
	"time"

	"starcup-engine/internal/model"
	"starcup-engine/internal/server/bot"
)

func (r *Room) handleRoomAction(client *Client, payload json.RawMessage) {
	var roomAction RoomActionRequest
	if err := json.Unmarshal(payload, &roomAction); err != nil {
		r.sendProtocolErrorToClient(client, protocolErrorCodeInvalidJSON, "RoomAction 负载不是合法 JSON", CmdRoomAction, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	switch roomAction.Action {
	case RoomActionDissolveRoom:
		r.handleDissolveRoomAction(client)
	case RoomActionAddBot:
		r.handleAddBotRoomAction(client, roomAction)
	case RoomActionRemoveBot:
		r.handleRemoveBotRoomAction(client, roomAction)
	case RoomActionTakeoverPlayer:
		r.handleTakeoverPlayerRoomAction(client, roomAction)
	case RoomActionChangeCamp:
		r.handleChangeCampRoomAction(client, roomAction)
	case RoomActionChangeRole:
		r.handleChangeRoleRoomAction(client, roomAction)
	case RoomActionStart:
		r.handleStartRoomAction(client)
	default:
		r.sendProtocolErrorToClient(client, protocolErrorCodeUnknownRoomAction, "未知房间动作", CmdRoomAction, map[string]interface{}{
			"action":    roomAction.Action,
			"room_code": r.Code,
			"player_id": client.PlayerID,
		})
	}
}

func (r *Room) sendRoomErrorToClient(client *Client, message string) {
	r.sendRoomEventToClient(client, RoomEvent{Action: "error", Message: message})
}

func (r *Room) handleDissolveRoomAction(client *Client) {
	var toClose []*Client

	r.mu.Lock()
	if !r.isHost(client) {
		r.mu.Unlock()
		r.sendRoomErrorToClient(client, "仅房主可解散房间")
		return
	}

	dissolveMsg := fmt.Sprintf("房主 %s 已解散房间", client.Name)
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
	r.botIntel = bot.NewMemory()
	r.publicTimelineSnapshot = nil
	r.mu.Unlock()

	r.engineMu.Lock()
	r.Engine = nil
	r.engineMu.Unlock()

	for _, c := range toClose {
		safeCloseBytesChan(c.Send)
		safeCloseConn(c.connSnapshot())
	}
}

func (r *Room) handleAddBotRoomAction(client *Client, roomAction RoomActionRequest) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Started {
		r.sendRoomErrorToClient(client, "游戏已开始，无法添加机器人")
		return
	}
	if !r.isHost(client) {
		r.sendRoomErrorToClient(client, "仅房主可添加机器人")
		return
	}
	if len(r.Clients) >= 6 {
		r.sendRoomErrorToClient(client, "房间已满")
		return
	}

	pid, err := r.nextAvailablePlayerIDLocked()
	if err != nil {
		r.sendRoomErrorToClient(client, err.Error())
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
}

func (r *Room) handleRemoveBotRoomAction(client *Client, roomAction RoomActionRequest) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Started {
		r.sendRoomErrorToClient(client, "游戏开始后不能移除机器人席位")
		return
	}
	if !r.isHost(client) {
		r.sendRoomErrorToClient(client, "仅房主可移除机器人")
		return
	}
	targetID := roomAction.TargetID
	if targetID == "" {
		r.sendRoomErrorToClient(client, "缺少机器人ID")
		return
	}
	target, ok := r.Clients[targetID]
	if !ok || target == nil || !target.IsBot {
		r.sendRoomErrorToClient(client, "目标不是机器人")
		return
	}
	delete(r.Clients, targetID)
	delete(r.botPromptCache, targetID)
	safeCloseBytesChan(target.Send)
	r.broadcastPlayerList()
}

func (r *Room) handleTakeoverPlayerRoomAction(client *Client, roomAction RoomActionRequest) {
	var takeoverBotID string

	r.mu.Lock()
	if !r.Started {
		r.mu.Unlock()
		r.sendRoomErrorToClient(client, "游戏未开始，无需托管")
		return
	}
	if !r.isHost(client) {
		r.mu.Unlock()
		r.sendRoomErrorToClient(client, "仅房主可启用托管")
		return
	}
	targetID := roomAction.TargetID
	if targetID == "" {
		r.mu.Unlock()
		r.sendRoomErrorToClient(client, "缺少目标玩家ID")
		return
	}
	target, ok := r.Clients[targetID]
	if !ok || target == nil {
		r.mu.Unlock()
		r.sendRoomErrorToClient(client, "目标玩家不存在")
		return
	}
	if target.IsBot {
		r.mu.Unlock()
		r.sendRoomErrorToClient(client, "目标已是机器人托管")
		return
	}
	if !target.Disconnected {
		r.mu.Unlock()
		r.sendRoomErrorToClient(client, "目标玩家当前在线，无需托管")
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
}

func (r *Room) handleChangeCampRoomAction(client *Client, roomAction RoomActionRequest) {
	if r.Started {
		r.sendRoomErrorToClient(client, "游戏已开始，无法调整阵营")
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	target := client
	if roomAction.TargetID != "" && roomAction.TargetID != client.PlayerID {
		if !r.isHost(client) {
			r.sendRoomErrorToClient(client, "仅房主可调整机器人阵营")
			return
		}
		t, ok := r.Clients[roomAction.TargetID]
		if !ok || t == nil || !t.IsBot {
			r.sendRoomErrorToClient(client, "仅可调整机器人阵营")
			return
		}
		target = t
	}

	camp := model.Camp(roomAction.Camp)
	if camp != model.RedCamp && camp != model.BlueCamp {
		r.sendRoomErrorToClient(client, "无效阵营")
		return
	}
	if target.Camp == camp {
		return
	}
	if r.campCount(camp) >= 3 {
		r.sendRoomErrorToClient(client, "该阵营人数已满")
		return
	}
	target.Camp = camp
	r.broadcastPlayerList()
	go r.autoStartIfReady()
}

func (r *Room) handleChangeRoleRoomAction(client *Client, roomAction RoomActionRequest) {
	if r.Started {
		r.sendRoomErrorToClient(client, "游戏已开始，无法调整角色")
		return
	}
	if roomAction.CharRole == "" {
		r.sendRoomErrorToClient(client, "缺少角色")
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	target := client
	if roomAction.TargetID != "" && roomAction.TargetID != client.PlayerID {
		if !r.isHost(client) {
			r.sendRoomErrorToClient(client, "仅房主可调整机器人角色")
			return
		}
		t, ok := r.Clients[roomAction.TargetID]
		if !ok || t == nil || !t.IsBot {
			r.sendRoomErrorToClient(client, "仅可调整机器人角色")
			return
		}
		target = t
	}

	if !isValidRole(roomAction.CharRole) {
		r.sendRoomErrorToClient(client, "无效角色")
		return
	}
	for pid, c := range r.Clients {
		if pid != target.PlayerID && c.CharRole == roomAction.CharRole {
			r.sendRoomErrorToClient(client, "该角色已被其他玩家选择")
			return
		}
	}
	target.CharRole = roomAction.CharRole
	r.broadcastPlayerList()
	go r.autoStartIfReady()
}

func (r *Room) handleStartRoomAction(client *Client) {
	r.mu.Lock()
	if r.Started {
		r.mu.Unlock()
		r.sendRoomErrorToClient(client, "游戏已开始")
		return
	}
	if len(r.Clients) < 2 {
		r.mu.Unlock()
		r.sendRoomErrorToClient(client, "至少需要2名玩家才能开始")
		return
	}
	r.mu.Unlock()

	if err := r.startGame(); err != nil {
		r.sendRoomErrorToClient(client, err.Error())
	}
}
