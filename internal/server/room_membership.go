package server

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"starcup-engine/internal/model"
	"starcup-engine/internal/server/prompting"
)

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
			r.sendRoomEventToClient(client, RoomEvent{Action: "error", Message: "游戏已开始，且重连校验失败，无法加入"})
			return
		}
	} else if r.tryReconnectByNameLocked(client) {
		return
	}

	if r.Started {
		r.sendRoomEventToClient(client, RoomEvent{Action: "error", Message: "游戏已开始，无法加入"})
		return
	}

	if len(r.Clients) >= 6 {
		r.sendRoomEventToClient(client, RoomEvent{Action: "error", Message: "房间已满"})
		return
	}

	// Assign player ID (camp 由玩家自己选择)。
	playerID, err := r.nextAvailablePlayerIDLocked()
	if err != nil {
		r.sendRoomEventToClient(client, RoomEvent{Action: "error", Message: err.Error()})
		return
	}
	client.PlayerID = playerID
	client.Camp = model.Camp("") // 空表示未选择阵营

	// 创建房间后默认不分配角色，等待玩家主动选择。
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

	// Send assignment to joining player（含角色数据供前端展示）。
	r.sendRoomEventToClient(client, RoomEvent{
		Action:         "assigned",
		RoomCode:       r.Code,
		PlayerID:       client.PlayerID,
		Camp:           string(client.Camp),
		CharRole:       client.CharRole,
		Characters:     buildCharacterViews(),
		Message:        map[bool]string{true: "你是房主", false: ""}[r.isHost(client)],
		ReconnectToken: client.ReconnectToken,
	})

	// Broadcast updated player list to all。
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

		// Broadcast player left。
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

	// 发送重连分配信息。
	r.sendRoomEventToClient(client, RoomEvent{
		Action:         "assigned",
		RoomCode:       r.Code,
		PlayerID:       client.PlayerID,
		Camp:           string(client.Camp),
		CharRole:       client.CharRole,
		Characters:     buildCharacterViews(),
		Message:        successMessage,
		ReconnectToken: client.ReconnectToken,
	})

	// 立即补发当前状态与提示。
	if r.Engine != nil {
		r.engineMu.Lock()
		var prompt *model.Prompt
		if p := r.Engine.GetCurrentPrompt(); p != nil && p.PlayerID == client.PlayerID {
			prompt = prompting.ClonePrompt(p)
		} else if cachedPrompt != nil {
			prompt = prompting.ClonePrompt(cachedPrompt)
		}
		r.engineMu.Unlock()

		r.sendSyncStateToClient(client)
		if prompt != nil {
			r.sendRequireActionToClient(client, prompt)
		}
	}

	// 通知其他玩家。
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

func (r *Room) isHost(client *Client) bool {
	return client != nil && client.PlayerID != "" && client.PlayerID == r.HostID
}

func (r *Room) nextAvailablePlayerIDLocked() (string, error) {
	for i := 1; i <= 6; i++ {
		playerID := fmt.Sprintf("p%d", i)
		if _, exists := r.Clients[playerID]; !exists {
			return playerID, nil
		}
	}
	return "", fmt.Errorf("房间已满")
}

func (r *Room) ensureHostLocked() {
	// 当前 host 仍在线且非机器人，则保持不变。
	if host, ok := r.Clients[r.HostID]; ok && host != nil && !host.IsBot && !host.Disconnected {
		return
	}
	// 优先选在线真人玩家作为新 host。
	for _, client := range r.Clients {
		if client != nil && !client.IsBot && !client.Disconnected {
			r.HostID = client.PlayerID
			return
		}
	}
	// 兜底：如果当前都离线，保留任意真人席位作为 host（便于其重连后恢复管理权）。
	for _, client := range r.Clients {
		if client != nil && !client.IsBot {
			r.HostID = client.PlayerID
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
	for _, playerID := range r.orderedClientIDsLocked() {
		client := r.Clients[playerID]
		if client == nil {
			continue
		}
		players = append(players, PlayerInfo{
			ID:       client.PlayerID,
			Name:     client.Name,
			Camp:     string(client.Camp),
			CharRole: client.CharRole,
			Ready:    client.Camp != "" && client.CharRole != "",
			IsOnline: client.IsBot || !client.Disconnected,
			IsBot:    client.IsBot,
			IsHost:   client.PlayerID == r.HostID,
			BotMode:  client.BotMode,
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
	for _, playerID := range r.SeatOrder {
		if _, ok := r.Clients[playerID]; !ok {
			continue
		}
		ids = append(ids, playerID)
		seen[playerID] = struct{}{}
	}

	// 未进入 SeatOrder 的玩家（未开局或中途补位）按 pid 稳定排序，避免 map 遍历抖动。
	rest := make([]string, 0, len(r.Clients)-len(ids))
	for playerID := range r.Clients {
		if _, ok := seen[playerID]; ok {
			continue
		}
		rest = append(rest, playerID)
	}
	sort.Strings(rest)
	ids = append(ids, rest...)
	return ids
}
