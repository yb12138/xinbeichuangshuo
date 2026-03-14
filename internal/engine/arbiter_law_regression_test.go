package engine

import (
	"strings"
	"testing"

	"starcup-engine/internal/model"
)

type captureObserver struct {
	events []model.GameEvent
}

func (o *captureObserver) OnGameEvent(event model.GameEvent) {
	o.events = append(o.events, event)
}

func (o *captureObserver) countLogContains(substr string) int {
	n := 0
	for _, e := range o.events {
		if e.Type != model.EventLog {
			continue
		}
		if strings.Contains(e.Message, substr) {
			n++
		}
	}
	return n
}

// 回归测试：仲裁法则仅在首次回合开始时生效，后续回合开始不应重复触发
func TestArbiterLaw_OnlyTriggersOnceOnTurnStart(t *testing.T) {
	obs := &captureObserver{}
	game := NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Arbiter", "arbiter", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 0
	p1.Crystal = 0

	ctx1 := game.buildContext(p1, nil, model.TriggerOnTurnStart, &model.EventContext{
		Type:     model.EventTurnStart,
		SourceID: "p1",
	})
	game.dispatcher.OnTrigger(model.TriggerOnTurnStart, ctx1)

	if p1.Crystal != 2 {
		t.Fatalf("expected crystal=2 after first turn-start trigger, got %d", p1.Crystal)
	}

	ctx2 := game.buildContext(p1, nil, model.TriggerOnTurnStart, &model.EventContext{
		Type:     model.EventTurnStart,
		SourceID: "p1",
	})
	game.dispatcher.OnTrigger(model.TriggerOnTurnStart, ctx2)

	if p1.Crystal != 2 {
		t.Fatalf("expected crystal to stay 2 after second trigger, got %d", p1.Crystal)
	}
	if got := obs.countLogContains("[仲裁法则]"); got != 1 {
		t.Fatalf("expected [仲裁法则] log once, got %d", got)
	}
}

// 回归测试：审判形态的“回合开始审判+1”应独立保留，不依赖仲裁法则重复触发
func TestArbiterForm_JudgmentAutoGainAtStartup(t *testing.T) {
	obs := &captureObserver{}
	game := NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Arbiter", "arbiter", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseStartup

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 0
	p1.Tokens = map[string]int{
		"arbiter_law_inited": 1,
		"arbiter_form":       1,
		"judgment":           3,
	}

	game.Drive()

	if p1.Tokens["judgment"] != 4 {
		t.Fatalf("expected judgment to increase to 4 in startup, got %d", p1.Tokens["judgment"])
	}
	if got := obs.countLogContains("[仲裁法则]"); got != 0 {
		t.Fatalf("expected no [仲裁法则] log in form upkeep, got %d", got)
	}
	if got := obs.countLogContains("处于审判形态，回合开始审判+1"); got != 1 {
		t.Fatalf("expected exactly one form upkeep log, got %d", got)
	}
}
