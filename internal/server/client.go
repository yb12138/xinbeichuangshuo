package server

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"starcup-engine/internal/model"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 8192
)

// Client represents a connected player
type Client struct {
	Room     *Room
	Conn     *websocket.Conn
	Send     chan []byte
	PlayerID string
	Name     string
	Camp     model.Camp
	CharRole string // 角色ID
	IsBot    bool   // 是否为机器人席位
	BotMode  string // bot来源: "added" | "takeover"
	// Disconnected 表示真人席位当前掉线（无活跃 websocket 连接）。
	// 机器人席位固定视为在线参与对局，不使用该标记。
	Disconnected bool

	// Reconnect support
	ReconnectToken    string
	ReconnectPlayerID string

	mu sync.Mutex
}

// NewClient creates a new client
func NewClient(conn *websocket.Conn, name string) *Client {
	return &Client{
		Conn: conn,
		Send: make(chan []byte, 256),
		Name: name,
	}
}

func (c *Client) connSnapshot() *websocket.Conn {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Conn
}

func safeCloseConn(conn *websocket.Conn) {
	if conn == nil {
		return
	}
	_ = conn.Close()
}

// SendMessage sends a message to the client
func (c *Client) SendMessage(msg interface{}) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}
	select {
	case c.Send <- data:
	default:
		log.Printf("Client %s send buffer full, dropping message", c.PlayerID)
	}
}

// ReadPump pumps messages from the websocket connection to the room
func (c *Client) ReadPump() {
	conn := c.connSnapshot()
	if conn == nil {
		return
	}

	defer func() {
		if c.Room != nil {
			c.Room.Unregister <- c
		}
		safeCloseConn(conn)
	}()

	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Parse incoming message
		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Printf("Error parsing message: %v", err)
			continue
		}

		// Handle different message types
		if c.Room != nil {
			c.Room.HandleMessage(c, &wsMsg)
		}
	}
}

// WritePump pumps messages from the room to the websocket connection
func (c *Client) WritePump() {
	conn := c.connSnapshot()
	if conn == nil {
		return
	}

	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		safeCloseConn(conn)
	}()

	for {
		select {
		case message, ok := <-c.Send:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The room closed the channel
				_ = conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
