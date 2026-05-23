package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"starcup-engine/internal/model"
)

// TestScenarioRequest is the request body for POST /api/test/setup-scenario.
type TestScenarioRequest struct {
	HumanPlayer TestPlayerConfig   `json:"human_player"`
	BotPlayers  []TestPlayerConfig `json:"bot_players"`
	Setup       TestSetup          `json:"setup"`
}

// TestPlayerConfig describes a player for scenario setup.
type TestPlayerConfig struct {
	Name     string `json:"name"`
	Camp     string `json:"camp"`
	CharRole string `json:"char_role"`
}

// TestSetup controls the post-start configuration.
type TestSetup struct {
	FirstTurnPlayer string      `json:"first_turn_player"`
	BotsPaused      bool        `json:"bots_paused"`
	Cheats          []TestCheat `json:"cheats"`
}

// TestCheat is a single cheat command.
type TestCheat struct {
	Target  string   `json:"target"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

// TestScenarioResponse is returned after successful setup.
type TestScenarioResponse struct {
	RoomCode      string   `json:"room_code"`
	HumanPlayerID string   `json:"human_player_id"`
	BotPlayerIDs  []string `json:"bot_player_ids"`
}

// HandleTestSetupScenario creates a room with pre-configured players and state for E2E testing.
// Only available when STARCUP_TEST_MODE=1.
func (s *Server) HandleTestSetupScenario(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("STARCUP_TEST_MODE") != "1" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TestScenarioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	room := s.createRoom()

	// Add players directly under lock, bypassing WebSocket registration.
	room.mu.Lock()

	humanPID, err := room.nextAvailablePlayerIDLocked()
	if err != nil {
		room.mu.Unlock()
		http.Error(w, "Failed to assign human player ID: "+err.Error(), http.StatusInternalServerError)
		return
	}
	human := &Client{
		Room:           room,
		Send:           make(chan []byte, 256),
		PlayerID:       humanPID,
		Name:           req.HumanPlayer.Name,
		Camp:           model.Camp(req.HumanPlayer.Camp),
		CharRole:       req.HumanPlayer.CharRole,
		IsBot:          false,
		Disconnected:   true,
		ReconnectToken: generateReconnectToken(),
	}
	room.Clients[humanPID] = human
	room.HostID = humanPID

	botPIDs := make([]string, 0, len(req.BotPlayers))
	for _, bp := range req.BotPlayers {
		pid, err := room.nextAvailablePlayerIDLocked()
		if err != nil {
			room.mu.Unlock()
			http.Error(w, "Failed to assign bot player ID: "+err.Error(), http.StatusInternalServerError)
			return
		}
		bot := &Client{
			Room:     room,
			Send:     make(chan []byte, 256),
			PlayerID: pid,
			Name:     bp.Name,
			Camp:     model.Camp(bp.Camp),
			CharRole: bp.CharRole,
			IsBot:    true,
			BotMode:  "added",
		}
		room.Clients[pid] = bot
		botPIDs = append(botPIDs, pid)
	}

	room.BotsPaused = req.Setup.BotsPaused
	room.mu.Unlock()

	if err := room.startGame(); err != nil {
		http.Error(w, "Failed to start game: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Execute cheats sequentially under engineMu.
	room.engineMu.Lock()
	for _, cheat := range req.Setup.Cheats {
		targetPID := resolveCheatTarget(cheat.Target, humanPID, botPIDs)
		act := model.PlayerAction{
			PlayerID:  targetPID,
			Type:      model.CmdCheat,
			TargetID:  cheat.Command,
			ExtraArgs: append([]string{targetPID}, cheat.Args...),
		}
		if err := room.Engine.HandleCheat(act); err != nil {
			room.engineMu.Unlock()
			http.Error(w, fmt.Sprintf("Cheat %s failed: %v", cheat.Command, err), http.StatusInternalServerError)
			return
		}
		room.Engine.Drive()
	}

	// Force turn to the desired player, clearing any pending interrupts first.
	if req.Setup.FirstTurnPlayer != "" {
		turnPID := resolveCheatTarget(req.Setup.FirstTurnPlayer, humanPID, botPIDs)
		room.Engine.State.PendingInterrupt = nil
		room.Engine.State.InterruptQueue = nil
		turnAct := model.PlayerAction{
			PlayerID:  turnPID,
			Type:      model.CmdCheat,
			TargetID:  "turn",
			ExtraArgs: []string{turnPID},
		}
		if err := room.Engine.HandleCheat(turnAct); err != nil {
			room.engineMu.Unlock()
			http.Error(w, fmt.Sprintf("Force turn failed: %v", err), http.StatusInternalServerError)
			return
		}
		room.Engine.Drive()
	}
	room.engineMu.Unlock()

	log.Printf("[TestAPI] Scenario setup complete: room=%s human=%s bots=%v", room.Code, humanPID, botPIDs)

	resp := TestScenarioResponse{
		RoomCode:      room.Code,
		HumanPlayerID: humanPID,
		BotPlayerIDs:  botPIDs,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func resolveCheatTarget(target string, humanPID string, botPIDs []string) string {
	switch target {
	case "human":
		return humanPID
	default:
		for i, pid := range botPIDs {
			if target == fmt.Sprintf("%d", i) || target == pid {
				return pid
			}
		}
		return target
	}
}
