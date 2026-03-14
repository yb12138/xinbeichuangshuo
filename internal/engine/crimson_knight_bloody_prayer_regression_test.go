package engine

import (
	"testing"

	"starcup-engine/internal/engine/skills"
	"starcup-engine/internal/model"
)

func TestCrimsonKnightBloodyPrayer_CanSplitHealToTwoAllies(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Crimson", "crimson_knight", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "AllyA", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "AllyB", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p4", "Enemy", "onmyoji", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p3 := game.State.Players["p3"]
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p3.TurnState = model.NewPlayerTurnState()
	p1.Heal = 3
	p2.Heal = 0
	p3.Heal = 0

	handler := skills.GetHandler("crk_bloody_prayer")
	if handler == nil {
		t.Fatalf("crk_bloody_prayer handler not found")
	}
	ctx := game.buildContext(p1, nil, model.TriggerOnTurnStart, nil)
	if !handler.CanUse(ctx) {
		t.Fatalf("expected bloody prayer can use when heal>0 and has allies")
	}
	if err := handler.Execute(ctx); err != nil {
		t.Fatalf("execute bloody prayer failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptChoice {
		t.Fatalf("expected choice interrupt after bloody prayer, got %+v", game.State.PendingInterrupt)
	}

	// X = 3
	if err := game.handleWeakChoiceInput("p1", 2); err != nil {
		t.Fatalf("choose x failed: %v", err)
	}
	// 选择 2 名队友
	if err := game.handleWeakChoiceInput("p1", 1); err != nil {
		t.Fatalf("choose ally count failed: %v", err)
	}
	// 第一个队友：p2（当前列表第一项）
	if err := game.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("choose first ally failed: %v", err)
	}
	// 第二个队友：只剩 p3，索引仍为 0
	if err := game.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("choose second ally failed: %v", err)
	}
	// 分配：p2 +2，p3 +1（X=3 时索引1）
	if err := game.handleWeakChoiceInput("p1", 1); err != nil {
		t.Fatalf("choose split failed: %v", err)
	}

	if got := p1.Heal; got != 0 {
		t.Fatalf("expected p1 heal=0 after remove 3, got %d", got)
	}
	if got := p2.Heal; got != 2 {
		t.Fatalf("expected p2 heal=2, got %d", got)
	}
	if got := p3.Heal; got != 1 {
		t.Fatalf("expected p3 heal=1, got %d", got)
	}
	if got := p1.Tokens["crk_blood_mark"]; got != 1 {
		t.Fatalf("expected p1 blood mark=1, got %d", got)
	}
	if len(game.State.PendingDamageQueue) != 1 {
		t.Fatalf("expected 1 pending self-damage, got %d", len(game.State.PendingDamageQueue))
	}
	pd := game.State.PendingDamageQueue[0]
	if pd.SourceID != "p1" || pd.TargetID != "p1" || pd.Damage != 3 || pd.DamageType != "magic" {
		t.Fatalf("unexpected pending damage: %+v", pd)
	}
}
