package server

import (
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
)

func hasAvailableSkill(skills []AvailableSkill, skillID string) bool {
	for _, s := range skills {
		if s.ID == skillID {
			return true
		}
	}
	return false
}

func TestBuildAvailableActionSkills_ElementalistIgniteAndMoonlightGating(t *testing.T) {
	room := NewRoom("ELEM")
	room.Engine = engine.NewGameEngine(room)

	if err := room.Engine.AddPlayer("p1", "Elem", "elementalist", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := room.Engine.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	room.Engine.State.CurrentTurn = 0
	room.Engine.State.Phase = model.PhaseActionSelection
	p1 := room.Engine.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()

	p1.Tokens["element"] = 2
	p1.Gem = 0
	p1.Crystal = 3
	skills := room.buildAvailableActionSkills("p1")
	if hasAvailableSkill(skills, "elementalist_ignite") {
		t.Fatalf("expected ignite hidden when element<3")
	}
	if hasAvailableSkill(skills, "elementalist_moonlight") {
		t.Fatalf("expected moonlight hidden when gem=0")
	}

	p1.Tokens["element"] = 3
	p1.Gem = 1
	skills = room.buildAvailableActionSkills("p1")
	if !hasAvailableSkill(skills, "elementalist_ignite") {
		t.Fatalf("expected ignite available when element>=3")
	}
	if !hasAvailableSkill(skills, "elementalist_moonlight") {
		t.Fatalf("expected moonlight available when gem>=1")
	}
}
