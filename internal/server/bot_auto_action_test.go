package server

import (
	"testing"
	"time"

	"starcup-engine/internal/model"
)

func TestBotAutoRespondsInCombat(t *testing.T) {
	room := NewRoom("TEST")
	human := &Client{
		Room:     room,
		Send:     make(chan []byte, 64),
		PlayerID: "p1",
		Name:     "human",
		Camp:     model.RedCamp,
		CharRole: "berserker",
	}
	bot := &Client{
		Room:     room,
		Send:     make(chan []byte, 64),
		PlayerID: "p2",
		Name:     "bot",
		Camp:     model.BlueCamp,
		CharRole: "adventurer",
		IsBot:    true,
		BotMode:  "added",
	}
	room.Clients[human.PlayerID] = human
	room.Clients[bot.PlayerID] = bot
	room.HostID = human.PlayerID

	if err := room.startGame(); err != nil {
		t.Fatalf("start game: %v", err)
	}

	// 强制设置为 p1 行动，避免随机先手导致用例不稳定。
	room.engineMu.Lock()
	state := room.Engine.State

	turnIdx := -1
	for i, pid := range state.PlayerOrder {
		if pid == human.PlayerID {
			turnIdx = i
			break
		}
	}
	if turnIdx < 0 {
		room.engineMu.Unlock()
		t.Fatalf("human player not found in order: %+v", state.PlayerOrder)
	}

	state.CurrentTurn = turnIdx
	state.Phase = model.PhaseActionSelection
	for pid, p := range state.Players {
		p.IsActive = pid == human.PlayerID
	}

	state.Players[human.PlayerID].Hand = []model.Card{
		{
			ID:      "test-atk",
			Name:    "测试攻击",
			Type:    model.CardTypeAttack,
			Element: model.ElementFire,
			Damage:  1,
		},
	}
	state.Players[bot.PlayerID].Hand = nil
	room.engineMu.Unlock()

	if err := room.submitAction(model.PlayerAction{
		PlayerID:  human.PlayerID,
		Type:      model.CmdAttack,
		TargetID:  bot.PlayerID,
		CardIndex: 0,
	}); err != nil {
		t.Fatalf("human attack failed: %v", err)
	}

	deadline := time.Now().Add(3 * time.Second)
	for {
		room.engineMu.Lock()
		phase := room.Engine.State.Phase
		combatDepth := len(room.Engine.State.CombatStack)
		room.engineMu.Unlock()

		if phase != model.PhaseCombatInteraction || combatDepth == 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("bot did not respond in time, phase=%s combatDepth=%d", phase, combatDepth)
		}
		time.Sleep(30 * time.Millisecond)
	}
}

func TestRunBotTurnIgnoresStaleCombatPrompt(t *testing.T) {
	room := NewRoom("TEST2")
	human := &Client{
		Room:     room,
		Send:     make(chan []byte, 64),
		PlayerID: "p1",
		Name:     "human",
		Camp:     model.RedCamp,
		CharRole: "berserker",
	}
	bot := &Client{
		Room:     room,
		Send:     make(chan []byte, 64),
		PlayerID: "p2",
		Name:     "bot",
		Camp:     model.BlueCamp,
		CharRole: "adventurer",
		IsBot:    true,
		BotMode:  "added",
	}
	room.Clients[human.PlayerID] = human
	room.Clients[bot.PlayerID] = bot
	room.HostID = human.PlayerID

	if err := room.startGame(); err != nil {
		t.Fatalf("start game: %v", err)
	}

	// 当前是行动选择，不是战斗响应阶段。
	room.engineMu.Lock()
	state := room.Engine.State
	state.Phase = model.PhaseActionSelection
	state.CombatStack = nil
	state.Players[bot.PlayerID].Hand = []model.Card{
		{
			ID:      "test-atk",
			Name:    "测试攻击",
			Type:    model.CardTypeAttack,
			Element: model.ElementFire,
			Damage:  1,
		},
	}
	beforeHand := len(state.Players[bot.PlayerID].Hand)
	room.engineMu.Unlock()

	// 传入过期的“战斗响应”提示，机器人应直接忽略，不应报错/不应出牌。
	staleCombatPrompt := &model.Prompt{
		Type:     model.PromptConfirm,
		PlayerID: bot.PlayerID,
		Options: []model.PromptOption{
			{ID: "take", Label: "承受伤害 (Take)"},
		},
	}
	if err := room.runBotTurn(bot.PlayerID, staleCombatPrompt, 0); err != nil {
		t.Fatalf("runBotTurn should ignore stale prompt, got err: %v", err)
	}

	room.engineMu.Lock()
	defer room.engineMu.Unlock()
	afterHand := len(room.Engine.State.Players[bot.PlayerID].Hand)
	if afterHand != beforeHand {
		t.Fatalf("bot should not consume cards on stale prompt, before=%d after=%d", beforeHand, afterHand)
	}
}

func TestPromptActionableForMagicMissileInterrupt(t *testing.T) {
	room := NewRoom("TEST3")
	human := &Client{
		Room:     room,
		Send:     make(chan []byte, 64),
		PlayerID: "p1",
		Name:     "human",
		Camp:     model.RedCamp,
		CharRole: "berserker",
	}
	bot := &Client{
		Room:     room,
		Send:     make(chan []byte, 64),
		PlayerID: "p2",
		Name:     "bot",
		Camp:     model.BlueCamp,
		CharRole: "adventurer",
		IsBot:    true,
		BotMode:  "added",
	}
	room.Clients[human.PlayerID] = human
	room.Clients[bot.PlayerID] = bot
	room.HostID = human.PlayerID

	if err := room.startGame(); err != nil {
		t.Fatalf("start game: %v", err)
	}

	prompt := &model.Prompt{
		Type:     model.PromptConfirm,
		PlayerID: bot.PlayerID,
		Options: []model.PromptOption{
			{ID: "take", Label: "承受伤害 (take)"},
			{ID: "defend", Label: "防御 (defend)"},
			{ID: "counter", Label: "传递 (counter)"},
		},
	}

	room.engineMu.Lock()
	room.Engine.State.Phase = model.PhaseResponse
	room.Engine.State.CombatStack = nil
	room.Engine.State.PendingInterrupt = &model.Interrupt{
		Type:     model.InterruptMagicMissile,
		PlayerID: bot.PlayerID,
	}
	ok := room.isPromptActionableLocked(bot.PlayerID, prompt)
	room.engineMu.Unlock()

	if !ok {
		t.Fatalf("magic missile interrupt prompt should be actionable for bot")
	}
}
