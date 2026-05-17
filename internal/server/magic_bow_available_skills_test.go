package server

import (
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
)

func TestBuildAvailableActionSkills_MagicBowThunderScatterRequiresThunderCharge(t *testing.T) {
	room := NewRoom("MAGICBOW")
	room.Engine = engine.NewGameEngine(room)

	if err := room.Engine.AddPlayer("p1", "MagicBow", "magic_bow", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := room.Engine.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	room.Engine.State.CurrentTurn = 0
	room.Engine.State.TurnStage = model.TurnStageActionExecution
	p1 := room.Engine.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Field = append(p1.Field, &model.FieldCard{
		Card: model.Card{
			ID:      "thunder-charge",
			Name:    "雷光充能",
			Type:    model.CardTypeAttack,
			Element: model.ElementThunder,
		},
		Mode:     model.FieldCover,
		Effect:   model.EffectMagicBowCharge,
		Hook:     model.FieldHookManual,
		OwnerID:  p1.ID,
		SourceID: p1.ID,
	})

	skills := room.buildAvailableActionSkills("p1")
	if !hasAvailableSkill(skills, "mb_thunder_scatter") {
		t.Fatalf("expected thunder scatter available with an unlocked thunder charge")
	}

	p1.TurnState.UsedSkillCounts["mb_charge_lock_turn"] = 1
	skills = room.buildAvailableActionSkills("p1")
	if hasAvailableSkill(skills, "mb_thunder_scatter") {
		t.Fatalf("expected thunder scatter hidden during the charge lock turn")
	}

	p1.TurnState.UsedSkillCounts["mb_charge_lock_turn"] = 0
	p1.Field[0].Card.Element = model.ElementFire
	skills = room.buildAvailableActionSkills("p1")
	if hasAvailableSkill(skills, "mb_thunder_scatter") {
		t.Fatalf("expected thunder scatter hidden without a thunder charge")
	}
}
