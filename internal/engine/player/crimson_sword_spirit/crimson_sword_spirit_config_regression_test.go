package crimson_sword_spirit_test

import (
	"starcup-engine/internal/engine"
	"testing"

	"starcup-engine/internal/model"
)

func TestCrimsonBloodRose_RequiresAllyAsSecondTarget(t *testing.T) {
	g := engine.NewGameEngine(nil)
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
	p3 := g.State.Players["p3"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["css_blood"] = 2
	p1.Heal = 0
	p3.Heal = 2
	g.State.RedCrystals = 1
	g.State.RedGems = 0

	g.State.CurrentTurn = 0
	g.State.TurnStage = model.TurnStageActionExecution

	err := g.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdSkill,
		SkillID:   "css_blood_rose",
		TargetIDs: []string{"p2", "p2"},
	})
	if err == nil {
		t.Fatalf("expected blood rose reject when 2nd target is not ally")
	}

	err = g.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdSkill,
		SkillID:   "css_blood_rose",
		TargetIDs: []string{"p3", "p1"},
	})
	if err != nil {
		t.Fatalf("expected blood rose allow ally as first target, got err=%v", err)
	}
	if p1.Tokens["css_blood"] != 0 {
		t.Fatalf("expected blood rose spend 2 blood, got %d", p1.Tokens["css_blood"])
	}
	if p3.Heal != 0 {
		t.Fatalf("expected first target lose up to 2 heal, got %d", p3.Heal)
	}
	if p1.Heal != 1 {
		t.Fatalf("expected second target gain 1 heal, got %d", p1.Heal)
	}
	if g.State.RedCrystals != 0 || g.State.RedGems != 1 {
		t.Fatalf("expected camp crystal->gem conversion, got red crystals=%d gems=%d", g.State.RedCrystals, g.State.RedGems)
	}
}

func TestCrimsonBloodRose_ConvertsCampCrystalAndHealsSecondTarget(t *testing.T) {
	g := engine.NewGameEngine(nil)
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
	p3.Heal = 0
	g.State.RedCrystals = 1
	g.State.RedGems = 0

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
		t.Fatalf("expected first target lose 2 heal, got %d", p2.Heal)
	}
	if p3.Heal != 1 {
		t.Fatalf("expected second target gain 1 heal, got %d", p3.Heal)
	}
	if g.State.RedCrystals != 0 || g.State.RedGems != 1 {
		t.Fatalf("expected camp crystal convert to gem, got red crystals=%d gems=%d", g.State.RedCrystals, g.State.RedGems)
	}
}
