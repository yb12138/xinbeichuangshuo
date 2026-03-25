package engine

import (
	"testing"

	"starcup-engine/internal/model"
)

func TestCrimsonBloodRose_RequiresEnemyAndAllyTargets(t *testing.T) {
	g := NewGameEngine(nil)
	if err := g.AddPlayer("p1", "CSS", "crimson_sword_spirit", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p3", "Ally", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["css_blood"] = 2

	g.State.CurrentTurn = 0
	g.State.TurnStage = model.TurnStageActionExecution

	err := g.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdSkill,
		SkillID:   "css_blood_rose",
		TargetIDs: []string{"p2", "p2"},
	})
	if err == nil {
		t.Fatalf("expected duplicate enemy targets to be rejected")
	}

	err = g.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdSkill,
		SkillID:   "css_blood_rose",
		TargetIDs: []string{"p3", "p1"},
	})
	if err == nil {
		t.Fatalf("expected double ally targets to be rejected")
	}
}

func TestCrimsonBloodRose_UsesSelectedAllyCrystalAndEnemyHeal(t *testing.T) {
	g := NewGameEngine(nil)
	if err := g.AddPlayer("p1", "CSS", "crimson_sword_spirit", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p3", "Ally", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p2 := g.State.Players["p2"]
	p3 := g.State.Players["p3"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["css_blood"] = 2
	p2.Heal = 3
	p3.Crystal = 1
	p3.Gem = 0

	g.State.CurrentTurn = 0
	g.State.TurnStage = model.TurnStageActionExecution

	if err := g.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdSkill,
		SkillID:   "css_blood_rose",
		TargetIDs: []string{"p2", "p3"},
	}); err != nil {
		t.Fatalf("blood rose failed: %v", err)
	}

	if p1.Tokens["css_blood"] != 0 {
		t.Fatalf("expected blood rose spend 2 blood, got %d", p1.Tokens["css_blood"])
	}
	if p2.Heal != 1 {
		t.Fatalf("expected enemy lose 2 heal, got %d", p2.Heal)
	}
	if p3.Crystal != 0 || p3.Gem != 1 {
		t.Fatalf("expected ally crystal convert to gem, got crystal=%d gem=%d", p3.Crystal, p3.Gem)
	}
}
