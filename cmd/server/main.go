package main

import (
	"log"
	"net/http"
	"os"

	"starcup-engine/internal/server"
)

func main() {
	s := server.NewServer()

	// WebSocket endpoint
	http.HandleFunc("/ws", s.HandleWebSocket)

	// REST endpoints
	http.HandleFunc("/api/health", s.HandleHealth)
	http.HandleFunc("/api/room/create", s.HandleCreateRoom)
	http.HandleFunc("/api/room/info", s.HandleRoomInfo)

	// Test API (only accessible when STARCUP_TEST_MODE=1; handler double-checks internally)
	if os.Getenv("STARCUP_TEST_MODE") == "1" {
		http.HandleFunc("/api/test/setup-scenario", s.HandleTestSetupScenario)
	}

	// Serve static files for frontend (if exists)
	fs := http.FileServer(http.Dir("./web/dist"))
	http.Handle("/", fs)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if port[0] != ':' {
		port = ":" + port
	}
	log.Printf("星杯传说 WebSocket 服务器启动在 %s", port)
	log.Printf("WebSocket: ws://localhost%s/ws?room=XXXX&name=PlayerName", port)
	log.Printf("创建房间: GET /api/room/create")
	log.Printf("房间信息: GET /api/room/info?room=XXXX")

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal("Server error:", err)
	}
}
