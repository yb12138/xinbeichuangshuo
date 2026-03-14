package engine

import (
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

// 回归：暗灭攻击不可应战，战斗请求应直接标记 CanBeResponded=false，
// 前端响应弹框不应出现“应战”按钮。
func TestDarkAttack_CombatRequestNotRespondable(t *testing.T) {
	g := NewGameEngine(noopObserver{})
	if err := g.AddPlayer("p1", "Attacker", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Defender", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	g.State.CurrentTurn = 0
	g.State.Deck = rules.InitDeck()
	g.State.Phase = model.PhaseActionSelection

	p1 := g.State.Players["p1"]
	p2 := g.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()

	p1.Hand = []model.Card{
		{ID: "atk-dark-1", Name: "暗灭", Type: model.CardTypeAttack, Element: model.ElementDark, Damage: 1},
	}

	if err := g.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	}); err != nil {
		t.Fatalf("dark attack failed: %v", err)
	}

	if len(g.State.CombatStack) != 1 {
		t.Fatalf("expected combat stack size 1, got %d", len(g.State.CombatStack))
	}
	req := g.State.CombatStack[0]
	if req.Card == nil || req.Card.Element != model.ElementDark {
		t.Fatalf("expected dark combat request, got %+v", req.Card)
	}
	if req.CanBeResponded {
		t.Fatalf("dark attack should be non-counterable")
	}
}
