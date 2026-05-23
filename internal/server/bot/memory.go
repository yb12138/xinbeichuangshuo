package bot

import (
	"math"
	"sync"

	"starcup-engine/internal/model"
)

// Memory records public reveal information that the bot can use for inference.
type Memory struct {
	mu      sync.RWMutex
	players map[string]*PlayerRevealStats
}

type PlayerRevealStats struct {
	AttackShown int
	MagicShown  int
	DefendShown int
	ElementSeen map[model.Element]int
}

func NewMemory() *Memory {
	return &Memory{
		players: make(map[string]*PlayerRevealStats),
	}
}

func (m *Memory) ensurePlayer(playerID string) *PlayerRevealStats {
	ps, ok := m.players[playerID]
	if ok {
		return ps
	}
	ps = &PlayerRevealStats{ElementSeen: map[model.Element]int{}}
	m.players[playerID] = ps
	return ps
}

func (m *Memory) ObserveCardRevealed(payload model.CardRevealedPayload) {
	if m == nil {
		return
	}
	if payload.Hidden {
		// 保持“人类可见信息”原则：暗弃不纳入推断。
		return
	}
	if payload.PlayerID == "" {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	ps := m.ensurePlayer(payload.PlayerID)
	for _, c := range payload.Cards {
		if c.Type == model.CardTypeAttack {
			ps.AttackShown++
		}
		if c.Type == model.CardTypeMagic {
			ps.MagicShown++
		}
		if c.Element != "" {
			ps.ElementSeen[c.Element]++
		}
	}
	if payload.ActionType == "defend" {
		ps.DefendShown += len(payload.Cards)
	}
}

func (m *Memory) DefendBias(playerID string) float64 {
	if m == nil {
		return 0
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	ps := m.players[playerID]
	if ps == nil {
		return 0
	}
	return math.Min(0.2, float64(ps.DefendShown)*0.04)
}

func (m *Memory) AttackBias(playerID string) float64 {
	if m == nil {
		return 0
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	ps := m.players[playerID]
	if ps == nil {
		return 0
	}
	total := ps.AttackShown + ps.MagicShown
	if total == 0 {
		return 0
	}
	attackRatio := float64(ps.AttackShown) / float64(total)
	return clamp((attackRatio-0.5)*0.2, -0.1, 0.1)
}
