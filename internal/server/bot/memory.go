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

func (m *Memory) ObserveReveal(data map[string]interface{}) {
	if m == nil {
		return
	}
	hidden, _ := data["hidden"].(bool)
	if hidden {
		// 保持“人类可见信息”原则：暗弃不纳入推断。
		return
	}
	playerID, _ := data["player_id"].(string)
	if playerID == "" {
		return
	}
	actionType, _ := data["action_type"].(string)
	cards := extractCardsFromEvent(data["cards"])

	m.mu.Lock()
	defer m.mu.Unlock()
	ps := m.ensurePlayer(playerID)
	for _, c := range cards {
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
	if actionType == "defend" {
		ps.DefendShown += len(cards)
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

func extractCardsFromEvent(raw interface{}) []model.Card {
	switch cards := raw.(type) {
	case []model.Card:
		return cards
	case []interface{}:
		out := make([]model.Card, 0, len(cards))
		for _, item := range cards {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			out = append(out, model.Card{
				ID:      toString(m["id"]),
				Name:    toString(m["name"]),
				Type:    model.CardType(toString(m["type"])),
				Element: model.Element(toString(m["element"])),
				Damage:  toInt(m["damage"]),
			})
		}
		return out
	default:
		return nil
	}
}

func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func toInt(v interface{}) int {
	switch t := v.(type) {
	case int:
		return t
	case int64:
		return int(t)
	case float64:
		return int(t)
	case float32:
		return int(t)
	default:
		return 0
	}
}
