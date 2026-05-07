package prayer_master_test

import (
	"starcup-engine/internal/engine"
	"starcup-engine/internal/testutils"
	"testing"

	"starcup-engine/internal/model"
)

func TestPrayerEnterForm_ConsumesGemAndSetsForm(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Prayer", "prayer_master", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionStart

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 1

	game.Drive()
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptStartupSkill {
		t.Fatalf("expected startup interrupt for prayer enter form, got %+v", game.State.PendingInterrupt)
	}
	if err := game.ConfirmStartupSkill("p1", "prayer_enter_form"); err != nil {
		t.Fatalf("confirm prayer enter form failed: %v", err)
	}
	if p1.Gem != 0 {
		t.Fatalf("expected prayer consume 1 gem, got %d", p1.Gem)
	}
	if p1.Form != model.FormPrayerMasterPrayer {
		t.Fatalf("expected prayer form %q, got %q", model.FormPrayerMasterPrayer, p1.Form)
	}
}

func TestPrayerPowerBlessing_ConsumesExclusiveZoneCardDirectly(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Prayer", "prayer_master", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.ExclusiveCards = []model.Card{
		{
			ID:              "starter-p1-prayer_power_blessing",
			Name:            "威力赐福",
			Type:            model.CardTypeMagic,
			Element:         model.ElementLight,
			ExclusiveChar1:  "prayer_master",
			ExclusiveSkill1: "威力赐福",
		},
	}

	if err := game.UseSkill("p1", "prayer_power_blessing", []string{"p2"}, nil); err != nil {
		t.Fatalf("use prayer power blessing failed: %v", err)
	}
	if p1.HasExclusiveCard(p1.Character.ID, "威力赐福") {
		t.Fatalf("expected power blessing consumed from exclusive zone")
	}
	if !testutils.HasFieldEffect(p2, model.EffectPowerBlessing) {
		t.Fatalf("expected power blessing field effect on ally")
	}
}

func TestPrayerSwiftBlessing_ConsumesExclusiveZoneCardDirectly(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Prayer", "prayer_master", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.ExclusiveCards = []model.Card{
		{
			ID:              "starter-p1-prayer_swift_blessing",
			Name:            "迅捷赐福",
			Type:            model.CardTypeMagic,
			Element:         model.ElementWind,
			ExclusiveChar1:  "prayer_master",
			ExclusiveSkill1: "迅捷赐福",
		},
	}

	if err := game.UseSkill("p1", "prayer_swift_blessing", []string{"p2"}, nil); err != nil {
		t.Fatalf("use prayer swift blessing failed: %v", err)
	}
	if p1.HasExclusiveCard(p1.Character.ID, "迅捷赐福") {
		t.Fatalf("expected swift blessing consumed from exclusive zone")
	}
	if !testutils.HasFieldEffect(p2, model.EffectSwiftBlessing) {
		t.Fatalf("expected swift blessing field effect on ally")
	}
}
