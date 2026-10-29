package blood_priestess_test

import (
	"reflect"
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
	"starcup-engine/internal/testutils"
)

func requireBleedingActivations(t *testing.T, obs *testutils.CaptureObserver, want int) {
	t.Helper()
	got := obs.SkillActivated("bp_bleeding")
	if len(got) != want {
		t.Fatalf("expected %d bp_bleeding skill_activated, got %d: %+v", want, len(got), got)
	}
	for _, p := range got {
		if p.PlayerID != "p1" || p.SkillName != "流血" || !reflect.DeepEqual(p.TargetIDs, []string{"p1"}) {
			t.Fatalf("unexpected bp_bleeding payload: %+v", p)
		}
	}
}

func TestBloodPriestessSkillActivated_BleedingEnterAndRelease(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Witch", "blood_priestess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	p1.MaxHand = 6
	for i, ele := range []model.Element{model.ElementFire, model.ElementWater, model.ElementWind, model.ElementThunder, model.ElementEarth, model.ElementDark, model.ElementLight, model.ElementFire} {
		p1.Hand = append(p1.Hand, bloodPriestessTestCard(string(rune('a'+i)), ele))
	}
	overflowCtx := game.BuildContext(p1, nil, model.TimingActionDuring, nil)
	overflowCtx.Flags["FromDamageDraw"] = true
	game.CheckHandLimitCtx(p1, overflowCtx)
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{6, 7}})
	if p1.Form != model.FormBloodPriestessBleeding {
		t.Fatalf("expected bleed form, got %q", p1.Form)
	}
	requireBleedingActivations(t, obs, 1)

	p1.Hand = p1.Hand[:2]
	game.State.TurnStage = model.TurnStageActionEnd
	game.HandlePostActionEndEffects(game.State.Players["p2"], model.ActionMagic)
	if p1.Form != "" {
		t.Fatalf("expected bleed form released, got %q", p1.Form)
	}
	requireBleedingActivations(t, obs, 2)
}

func TestBloodPriestessSkillActivated_BleedingTurnStartTick(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Witch", "blood_priestess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Form = model.FormBloodPriestessBleeding
	game.State.Deck = rules.InitDeck()
	p1.Hand = []model.Card{
		bloodPriestessTestCard("s1", model.ElementFire),
		bloodPriestessTestCard("s2", model.ElementWater),
		bloodPriestessTestCard("s3", model.ElementWind),
	}
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageTurnStart
	game.Drive()

	requireBleedingActivations(t, obs, 1)
}

func TestBloodPriestessSkillActivated_NoEmitWhenNotBleeding(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Witch", "blood_priestess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	game.State.Players["p1"].Hand = nil
	game.State.TurnStage = model.TurnStageActionEnd
	game.HandlePostActionEndEffects(game.State.Players["p2"], model.ActionMagic)
	requireBleedingActivations(t, obs, 0)
}
