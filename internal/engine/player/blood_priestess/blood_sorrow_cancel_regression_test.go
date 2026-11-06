package blood_priestess_test

import (
	"testing"

	"starcup-engine/internal/engine"
	bloodpriestesspkg "starcup-engine/internal/engine/player/blood_priestess"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
	"starcup-engine/internal/testutils"
)

func newBloodSorrowGame(t *testing.T) (*engine.GameEngine, *model.Player) {
	t.Helper()
	game := engine.NewGameEngine(testutils.NoopObserver{})
	for _, p := range []struct {
		id, role string
		camp     model.Camp
	}{
		{"p1", "blood_priestess", model.RedCamp}, {"p2", "angel", model.RedCamp}, {"p3", "berserker", model.BlueCamp},
	} {
		if err := game.AddPlayer(p.id, p.id, p.role, p.camp); err != nil {
			t.Fatal(err)
		}
	}
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{bloodPriestessTestCard("c1", model.ElementFire), bloodPriestessTestCard("c2", model.ElementWater)}
	p1.ExclusiveCards = append(p1.ExclusiveCards, testutils.MakeStarterBloodSharedLifeCard(p1))
	game.State.Deck = rules.InitDeck()
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSkill, SkillID: "bp_shared_life"})
	testutils.MustHandleAction(t, game, model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}})
	game.Drive()
	p1.Form = model.FormBloodPriestessBleeding
	ctx := game.BuildContext(p1, nil, model.TimingTurnStart, nil)
	if err := (&bloodpriestesspkg.BloodSorrowHandler{}).Execute(ctx); err != nil {
		t.Fatal(err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "bp_blood_sorrow_mode")
	return game, p1
}

// 审计 血之巫女-4：血之哀伤第一步可取消（不自伤、同生共死不动），第二步可返回第一步。
func TestBloodPriestessBloodSorrow_CancelAndBack(t *testing.T) {
	game, p1 := newBloodSorrowGame(t)
	if p := game.GetCurrentPrompt(); p == nil || p.Presentation == nil || p.Presentation.CancelPolicy != model.CancelPolicyAbort {
		t.Fatalf("expected abort on mode step, got %+v", p)
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}})
	testutils.RequireChoicePrompt(t, game, "p1", "bp_blood_sorrow_target")
	if p := game.GetCurrentPrompt(); p == nil || p.Presentation == nil || p.Presentation.CancelPolicy != model.CancelPolicyBack {
		t.Fatalf("expected back on target step, got %+v", p)
	}
	if n := len(game.State.PendingDamageQueue); n != 0 {
		t.Fatalf("self damage must not be queued before target chosen, got %d", n)
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{Type: model.CmdCancel, PlayerID: "p1"})
	testutils.RequireChoicePrompt(t, game, "p1", "bp_blood_sorrow_mode")
	testutils.MustHandleAction(t, game, model.PlayerAction{Type: model.CmdCancel, PlayerID: "p1"})
	if intr := game.State.PendingInterrupt; intr != nil && testutils.ChoiceTypeOfInterrupt(intr) == "bp_blood_sorrow_mode" {
		t.Fatalf("expected sorrow prompt closed")
	}
	if n := len(game.State.PendingDamageQueue); n != 0 {
		t.Fatalf("cancel must not queue self damage, got %d", n)
	}
	holder, _ := bloodpriestesspkg.FindSharedLife(engine.NewRoleChoiceRuntime(game), p1)
	if holder == nil || holder.ID != "p2" {
		t.Fatalf("shared life must stay on p2, got %+v", holder)
	}
}
