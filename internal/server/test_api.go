package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"starcup-engine/internal/model"
)

// TestScenarioRequest is the request body for POST /api/test/setup-scenario.
type TestScenarioRequest struct {
	HumanPlayer  TestPlayerConfig   `json:"human_player"`
	HumanPlayers []TestPlayerConfig `json:"human_players"`
	BotPlayers   []TestPlayerConfig `json:"bot_players"`
	Setup        TestSetup          `json:"setup"`
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
	RoomCode       string   `json:"room_code"`
	HumanPlayerID  string   `json:"human_player_id"`
	HumanPlayerIDs []string `json:"human_player_ids"`
	BotPlayerIDs   []string `json:"bot_player_ids"`
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
	humanPlayers := normalizeTestHumanPlayers(req)
	if len(humanPlayers) == 0 {
		http.Error(w, "At least one human player is required", http.StatusBadRequest)
		return
	}

	room := s.createRoom()

	// Add players directly under lock, bypassing WebSocket registration.
	room.mu.Lock()

	humanPIDs := make([]string, 0, len(humanPlayers))
	for _, hp := range humanPlayers {
		pid, err := room.nextAvailablePlayerIDLocked()
		if err != nil {
			room.mu.Unlock()
			http.Error(w, "Failed to assign human player ID: "+err.Error(), http.StatusInternalServerError)
			return
		}
		human := &Client{
			Room:           room,
			Send:           make(chan []byte, 256),
			PlayerID:       pid,
			Name:           hp.Name,
			Camp:           model.Camp(hp.Camp),
			CharRole:       hp.CharRole,
			IsBot:          false,
			Disconnected:   true,
			ReconnectToken: generateReconnectToken(),
		}
		room.Clients[pid] = human
		if room.HostID == "" {
			room.HostID = pid
		}
		humanPIDs = append(humanPIDs, pid)
	}

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
		targetPID := resolveCheatTarget(cheat.Target, humanPIDs, botPIDs)
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
		turnPID := resolveCheatTarget(req.Setup.FirstTurnPlayer, humanPIDs, botPIDs)
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

	log.Printf("[TestAPI] Scenario setup complete: room=%s humans=%v bots=%v", room.Code, humanPIDs, botPIDs)

	resp := TestScenarioResponse{
		RoomCode:       room.Code,
		HumanPlayerID:  humanPIDs[0],
		HumanPlayerIDs: humanPIDs,
		BotPlayerIDs:   botPIDs,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func normalizeTestHumanPlayers(req TestScenarioRequest) []TestPlayerConfig {
	if len(req.HumanPlayers) > 0 {
		return req.HumanPlayers
	}
	if req.HumanPlayer.Name != "" || req.HumanPlayer.Camp != "" || req.HumanPlayer.CharRole != "" {
		return []TestPlayerConfig{req.HumanPlayer}
	}
	return nil
}

func resolveCheatTarget(target string, humanPIDs []string, botPIDs []string) string {
	switch target {
	case "human":
		if len(humanPIDs) > 0 {
			return humanPIDs[0]
		}
	case "bot":
		if len(botPIDs) > 0 {
			return botPIDs[0]
		}
	default:
		if index, ok := parseIndexedTestTarget(target, "human"); ok && index >= 0 && index < len(humanPIDs) {
			return humanPIDs[index]
		}
		if index, ok := parseIndexedTestTarget(target, "bot"); ok && index >= 0 && index < len(botPIDs) {
			return botPIDs[index]
		}
		for i, pid := range botPIDs {
			if target == fmt.Sprintf("%d", i) || target == pid {
				return pid
			}
		}
		for i, pid := range humanPIDs {
			if target == fmt.Sprintf("h%d", i) || target == pid {
				return pid
			}
		}
		return target
	}
	return target
}

func parseIndexedTestTarget(target string, prefix string) (int, bool) {
	if !strings.HasPrefix(target, prefix) {
		return 0, false
	}
	raw := strings.TrimPrefix(target, prefix)
	if raw == "" {
		return 0, false
	}
	var index int
	if _, err := fmt.Sscanf(raw, "%d", &index); err != nil {
		return 0, false
	}
	return index, true
}
