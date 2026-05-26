package server

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/server/bot"
)

// Room represents a game room.
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
	actorInbox chan roomActorCall

	mu       sync.RWMutex
	engineMu sync.Mutex
	timelineMu sync.Mutex

	// 机器人全局观察信息（用于手牌类型推断）。
	botIntel *bot.Memory
	// 机器人最近一次收到的 Prompt 缓存（用于非中断提示，如 CombatInteraction/ActionSelection）。
	botPromptCache map[string]*model.Prompt
	// AskInput 全局版本号：每次新提示+1，用于丢弃旧定时器动作。
	botPromptEpoch uint64
	// NotifyTimeline 事件序号，需在单房间内严格单调递增。
	timelineSeq int64
	// timelineHistory 保存最近公共时间线，用于刷新/重连后重建当前战斗叙事层。
	timelineHistory []TimelineEvent
	// publicTimelineSnapshot 是 timeline state_delta 的公共可见状态基线。
	publicTimelineSnapshot *publicTimelineSnapshot
	// BotsPaused E2E 测试模式：暂停 bot 自动行动
	BotsPaused bool
	// actorLoopStarted 标记房间 Run 循环已启动，可用于将 gameplay 输入统一串行化到 inbox。
	actorLoopStarted actorLoopFlag
}

// NewRoom creates a new game room.
func NewRoom(code string) *Room {
	return &Room{
		Code:           code,
		Clients:        make(map[string]*Client),
		Register:       make(chan *Client),
		Unregister:     make(chan *Client),
		Broadcast:      make(chan []byte, 256),
		actorInbox:     make(chan roomActorCall, 128),
		Started:        false,
		botIntel:       bot.NewMemory(),
		botPromptCache: make(map[string]*model.Prompt),
		botPromptEpoch: 0,
	}
}

// Run starts the room's main loop.
func (r *Room) Run() {
	r.actorLoopStarted.Store(true)
	for {
		select {
		case client := <-r.Register:
			r.handleRegister(client)
		case client := <-r.Unregister:
			r.handleUnregister(client)
		case message := <-r.Broadcast:
			r.broadcastToAll(message)
		case call := <-r.actorInbox:
			r.handleActorCall(call)
		}
	}
}

func generateReconnectToken() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Int63())
}
