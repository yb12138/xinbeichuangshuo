package magical_girl_test

import (
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/testutils"
)

func newMagicBulletControlGame(t *testing.T, obs model.GameObserver) *engine.GameEngine {
	t.Helper()
	game := engine.NewGameEngine(obs)
	for _, p := range []struct {
		id, role string
		camp     model.Camp
	}{
		{"p1", "magical_girl", model.RedCamp},
		{"p2", "saintess", model.RedCamp},
		{"p3", "berserker", model.BlueCamp},
		{"p4", "angel", model.BlueCamp},
	} {
		if err := game.AddPlayer(p.id, p.id, p.role, p.camp); err != nil {
			t.Fatalf("add %s failed: %v", p.id, err)
		}
	}
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	game.State.PlayerOrder = []string{"p1", "p2", "p3", "p4"}
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{{ID: "mb", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 2}}
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdMagic,
		CardID:   testutils.PlayableCardID(t, game, "p1", 0),
	}); err != nil {
		t.Fatalf("play magic bullet failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptMagicBulletDirection {
		t.Fatalf("expected direction interrupt, got %+v", game.State.PendingInterrupt)
	}
	return game
}

func TestMagicBulletControl_ReverseEmitsSkillActivated(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := newMagicBulletControlGame(t, obs)
	if n := len(obs.SkillActivated("magic_bullet_control")); n != 0 {
		t.Fatalf("direction prompt should not emit skill_activated, got %d", n)
	}
	if err := game.HandleAction(model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{1}}); err != nil {
		t.Fatalf("choose reverse failed: %v", err)
	}
	testutils.RequireSingleSkillActivation(t, obs, "magic_bullet_control", "p1", "p3")
}

func TestMagicBulletControl_ForwardDoesNotEmitSkillActivated(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := newMagicBulletControlGame(t, obs)
	if err := game.HandleAction(model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0}}); err != nil {
		t.Fatalf("choose forward failed: %v", err)
	}
	if n := len(obs.SkillActivated("magic_bullet_control")); n != 0 {
		t.Fatalf("forward direction is not a control activation, got %d", n)
	}
}
