package spirit_caster_test

import (
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/testutils"
)

func TestSpiritCasterIncantation_EmitsSkillActivatedOnce(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "SpiritCaster", "spirit_caster", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "EnemyA", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "EnemyB", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		spiritCasterTestCard("w1", "风符", model.CardTypeMagic, model.ElementWind),
		spiritCasterTestCard("h1", "补牌", model.CardTypeAttack, model.ElementFire),
	}
	game.State.Players["p2"].Hand = []model.Card{spiritCasterTestCard("a1", "攻击A1", model.CardTypeAttack, model.ElementFire)}
	game.State.Players["p3"].Hand = []model.Card{spiritCasterTestCard("b1", "攻击B1", model.CardTypeAttack, model.ElementWater)}
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	if err := game.UseSkill("p1", "sc_talisman_wind", []string{"p2", "p3"}, []int{0}); err != nil {
		t.Fatalf("use talisman wind failed: %v", err)
	}
	game.ProcessPendingSkillResume()
	testutils.RequireChoicePrompt(t, game, "p1", "sc_incant_confirm")
	testutils.MustHandleAction(t, game, model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}})
	testutils.RequireChoicePrompt(t, game, "p1", "sc_incant_card")
	if n := len(obs.SkillActivated("sc_incantation")); n != 0 {
		t.Fatalf("expected no incantation event before card placed, got %d", n)
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}})

	testutils.RequireSingleSkillActivation(t, obs, "sc_incantation", "p1", "p1")
}
