package server

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
)

func (r *Room) campCount(camp model.Camp) int {
	count := 0
	for _, client := range r.Clients {
		if client.Camp == camp {
			count++
		}
	}
	return count
}

func (r *Room) validateLineupLocked() error {
	if len(r.Clients) < 2 {
		return fmt.Errorf("至少需要2名玩家才能开始")
	}
	redCount := 0
	blueCount := 0
	roleOwners := make(map[string]string, len(r.Clients))
	for _, client := range r.Clients {
		if client.Camp == "" {
			return fmt.Errorf("有人未选择阵营")
		}
		if client.CharRole == "" {
			return fmt.Errorf("有人未选择角色")
		}
		if owner, ok := roleOwners[client.CharRole]; ok {
			return fmt.Errorf("角色不可重复：%s 与 %s 都选择了 %s", owner, client.Name, client.CharRole)
		}
		roleOwners[client.CharRole] = client.Name
		switch client.Camp {
		case model.RedCamp:
			redCount++
		case model.BlueCamp:
			blueCount++
		default:
			return fmt.Errorf("存在无效阵营配置")
		}
	}
	if redCount < 1 || redCount > 3 {
		return fmt.Errorf("红队需1-3人，当前%d人", redCount)
	}
	if blueCount < 1 || blueCount > 3 {
		return fmt.Errorf("蓝队需1-3人，当前%d人", blueCount)
	}
	return nil
}

func (r *Room) canAutoStartLocked() bool {
	return r.validateLineupLocked() == nil
}

func buildInterleavedLineup(players []lineupPlayer, randomizer *rand.Rand) []lineupPlayer {
	if len(players) <= 2 || randomizer == nil {
		out := make([]lineupPlayer, len(players))
		copy(out, players)
		return out
	}

	redPlayers := make([]lineupPlayer, 0, len(players))
	bluePlayers := make([]lineupPlayer, 0, len(players))
	otherPlayers := make([]lineupPlayer, 0, len(players))
	for _, player := range players {
		switch player.camp {
		case model.RedCamp:
			redPlayers = append(redPlayers, player)
		case model.BlueCamp:
			bluePlayers = append(bluePlayers, player)
		default:
			otherPlayers = append(otherPlayers, player)
		}
	}

	randomizer.Shuffle(len(redPlayers), func(i, j int) { redPlayers[i], redPlayers[j] = redPlayers[j], redPlayers[i] })
	randomizer.Shuffle(len(bluePlayers), func(i, j int) { bluePlayers[i], bluePlayers[j] = bluePlayers[j], bluePlayers[i] })
	randomizer.Shuffle(len(otherPlayers), func(i, j int) { otherPlayers[i], otherPlayers[j] = otherPlayers[j], otherPlayers[i] })

	firstGroup := redPlayers
	secondGroup := bluePlayers
	if len(bluePlayers) > len(redPlayers) || (len(bluePlayers) == len(redPlayers) && len(bluePlayers) > 0 && randomizer.Intn(2) == 1) {
		firstGroup, secondGroup = bluePlayers, redPlayers
	}

	result := make([]lineupPlayer, 0, len(players))
	for len(firstGroup) > 0 || len(secondGroup) > 0 {
		if len(firstGroup) > 0 {
			result = append(result, firstGroup[0])
			firstGroup = firstGroup[1:]
		}
		if len(secondGroup) > 0 {
			result = append(result, secondGroup[0])
			secondGroup = secondGroup[1:]
		}
	}
	result = append(result, otherPlayers...)
	return result
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
	for _, player := range lineup {
		r.SeatOrder = append(r.SeatOrder, player.id)
	}
	r.Started = true
	// 同步广播固定座次，避免前端仍沿用 lobby 随机顺序导致同阵营扎堆显示。
	r.broadcastPlayerList()
	r.mu.Unlock()

	// Create game engine with this room as observer。
	r.Engine = engine.NewGameEngine(r)

	// Add all players to the engine。
	for _, player := range lineup {
		err := r.Engine.AddPlayer(player.id, player.name, player.role, player.camp)
		if err != nil {
			log.Printf("Error adding player: %v", err)
		}
	}
	r.resetPublicTimelineSnapshot()

	// Broadcast game started（含角色数据供前端技能 fallback）。
	r.broadcastRoomEvent(RoomEvent{
		Action:     "started",
		RoomCode:   r.Code,
		Message:    "游戏开始！",
		Characters: buildCharacterViews(),
	})

	// Start the game。
	r.engineMu.Lock()
	if err := r.Engine.StartGame(); err != nil {
		r.engineMu.Unlock()
		log.Printf("Error starting game: %v", err)
		r.mu.Lock()
		r.Started = false
		r.mu.Unlock()
		r.Engine = nil
		r.publicTimelineSnapshot = nil
		r.broadcastRoomEvent(RoomEvent{
			Action:  "error",
			Message: fmt.Sprintf("游戏启动失败: %v", err),
		})
		return nil
	}
	// Drive game loop。
	r.Engine.Drive()
	r.engineMu.Unlock()

	// 若首个操作者是机器人，自动驱动其行动。
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
