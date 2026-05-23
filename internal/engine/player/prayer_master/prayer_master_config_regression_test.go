package prayer_master_test

import (
	"starcup-engine/internal/engine"
	skillhandlers "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/testutils"
	"testing"

	"starcup-engine/internal/model"
)

func TestPrayerRuneGain_IsHookNotRegisteredCharacterSkill(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Prayer", "prayer_master", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	for _, skill := range p1.Character.Skills {
		if skill.ID == "prayer_rune_gain" {
			t.Fatalf("prayer_rune_gain should be a passive hook, not a registered character skill")
		}
	}
	if handler := skillhandlers.GetHandler("prayer_rune_gain"); handler != nil {
		t.Fatalf("prayer_rune_gain should not register a skill handler")
	}
}

func TestPrayerRuneGainHook_ActiveAttackAddsRuneOnlyInPrayerForm(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Prayer", "prayer_master", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{{
		ID:      "prayer-attack-1",
		Name:    "光刃",
		Type:    model.CardTypeAttack,
		Element: model.ElementLight,
		Damage:  2,
	}}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdAttack,
		TargetID: "p2",
		CardID:   testutils.PlayableCardID(t, game, "p1", 0),
	}); err != nil {
		t.Fatalf("attack outside prayer form failed: %v", err)
	}
	if got := p1.Tokens["prayer_rune"]; got != 0 {
		t.Fatalf("expected no rune outside prayer form, got %d", got)
	}

	game = engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Prayer", "prayer_master", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	p1 = game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Form = model.FormPrayerMasterPrayer
	p1.Tokens["prayer_rune"] = 2
	p1.Hand = []model.Card{{
		ID:      "prayer-attack-2",
		Name:    "光刃",
		Type:    model.CardTypeAttack,
		Element: model.ElementLight,
		Damage:  2,
	}}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdAttack,
		TargetID: "p2",
		CardID:   testutils.PlayableCardID(t, game, "p1", 0),
	}); err != nil {
		t.Fatalf("attack in prayer form failed: %v", err)
	}
	if got := p1.Tokens["prayer_rune"]; got != 3 {
		t.Fatalf("expected active attack in prayer form to cap rune at 3, got %d", got)
	}
}

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

func TestPrayerPowerBlessing_ConsumesSelectedHandCard(t *testing.T) {
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
	p1.Hand = []model.Card{
		{
			ID:              "hand-p1-prayer_power_blessing",
			Name:            "威力赐福",
			Type:            model.CardTypeMagic,
			Element:         model.ElementLight,
			ExclusiveChar1:  "prayer_master",
			ExclusiveSkill1: "威力赐福",
		},
	}

	if err := game.UseSkill("p1", "prayer_power_blessing", []string{"p2"}, []int{0}); err != nil {
		t.Fatalf("use prayer power blessing failed: %v", err)
	}
	if len(p1.Hand) != 0 {
		t.Fatalf("expected selected power blessing hand card consumed, got %d cards", len(p1.Hand))
	}
	if !testutils.HasFieldEffect(p2, model.EffectPowerBlessing) {
		t.Fatalf("expected power blessing field effect on ally")
	}
}

func TestPrayerPowerBlessing_MissingTargetDoesNotConsumeExclusiveCard(t *testing.T) {
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
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{
			ID:              "hand-p1-prayer_power_blessing",
			Name:            "威力赐福",
			Type:            model.CardTypeMagic,
			Element:         model.ElementLight,
			ExclusiveChar1:  "prayer_master",
			ExclusiveSkill1: "威力赐福",
		},
	}

	if err := game.UseSkill("p1", "prayer_power_blessing", nil, []int{0}); err == nil {
		t.Fatalf("expected missing target to fail")
	}
	if len(p1.Hand) != 1 {
		t.Fatalf("expected failed power blessing to keep selected hand card, got %d cards", len(p1.Hand))
	}
}

func TestPrayerSwiftBlessing_ConsumesSelectedHandCard(t *testing.T) {
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
	p1.Hand = []model.Card{
		{
			ID:              "hand-p1-prayer_swift_blessing",
			Name:            "迅捷赐福",
			Type:            model.CardTypeMagic,
			Element:         model.ElementWind,
			ExclusiveChar1:  "prayer_master",
			ExclusiveSkill1: "迅捷赐福",
		},
	}

	if err := game.UseSkill("p1", "prayer_swift_blessing", []string{"p2"}, []int{0}); err != nil {
		t.Fatalf("use prayer swift blessing failed: %v", err)
	}
	if len(p1.Hand) != 0 {
		t.Fatalf("expected selected swift blessing hand card consumed, got %d cards", len(p1.Hand))
	}
	if !testutils.HasFieldEffect(p2, model.EffectSwiftBlessing) {
		t.Fatalf("expected swift blessing field effect on ally")
	}
}
