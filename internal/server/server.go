package server

import (
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Server represents the WebSocket server
type Server struct {
	rooms    map[string]*Room
	upgrader websocket.Upgrader
	mu       sync.RWMutex
}

// NewServer creates a new WebSocket server
func NewServer() *Server {
	return &Server{
		rooms: make(map[string]*Room),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
	}
}

// HandleWebSocket handles WebSocket connection requests
func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	roomCode := r.URL.Query().Get("room")
	playerName := r.URL.Query().Get("name")
	createRoom := r.URL.Query().Get("create") == "true"
	reconnectPlayerID := r.URL.Query().Get("player_id")
	reconnectToken := r.URL.Query().Get("reconnect_token")

	if playerName == "" {
		http.Error(w, "Missing player name", http.StatusBadRequest)
		return
	}

	// Create or join room
	var room *Room
	if createRoom {
		room = s.createRoom()
		roomCode = room.Code
		log.Printf("Created room: %s", roomCode)
	} else {
		if roomCode == "" {
			http.Error(w, "Missing room code", http.StatusBadRequest)
			return
		}
		room = s.getRoom(roomCode)
		if room == nil {
			http.Error(w, "Room not found", http.StatusNotFound)
			return
		}
	}

	// Upgrade to WebSocket
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Create client
	client := NewClient(conn, playerName)
	client.ReconnectPlayerID = reconnectPlayerID
	client.ReconnectToken = reconnectToken

	// Register client to room
	room.Register <- client

	// Start client goroutines
	go client.WritePump()
	go client.ReadPump()
}

// HandleCreateRoom handles room creation via HTTP
func (s *Server) HandleCreateRoom(w http.ResponseWriter, r *http.Request) {
	room := s.createRoom()
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"room_code":"` + room.Code + `"}`))
}

// HandleRoomInfo returns room information
func (s *Server) HandleRoomInfo(w http.ResponseWriter, r *http.Request) {
	roomCode := r.URL.Query().Get("room")
	room := s.getRoom(roomCode)

	w.Header().Set("Content-Type", "application/json")
	if room == nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Room not found"}`))
		return
	}

	room.mu.RLock()
	defer room.mu.RUnlock()

	var players []PlayerInfo
	for _, c := range room.Clients {
		players = append(players, PlayerInfo{
			ID:       c.PlayerID,
			Name:     c.Name,
			Camp:     string(c.Camp),
			CharRole: c.CharRole,
			Ready:    c.Camp != "" && c.CharRole != "",
			IsOnline: c.IsBot || !c.Disconnected,
			IsBot:    c.IsBot,
			IsHost:   c.PlayerID == room.HostID,
			BotMode:  c.BotMode,
		})
	}

	response := map[string]interface{}{
		"room_code":    room.Code,
		"player_count": len(room.Clients),
		"started":      room.Started,
		"players":      players,
	}

	data, _ := mustMarshal(response).MarshalJSON()
	w.Write(data)
}

func (s *Server) createRoom() *Room {
	s.mu.Lock()
	defer s.mu.Unlock()

	code := generateRoomCode()
	for s.rooms[code] != nil {
		code = generateRoomCode()
	}

	room := NewRoom(code)
	s.rooms[code] = room
	go room.Run()

	return room
}

func (s *Server) getRoom(code string) *Room {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.rooms[code]
}

func generateRoomCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, 4)
	for i := range code {
		code[i] = charset[rand.Intn(len(charset))]
	}
	return string(code)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
