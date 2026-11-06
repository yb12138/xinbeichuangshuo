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

// 审计 魔法少女-1：方向只在魔法少女主动使用魔弹时选一次；传递过程中（包括魔法少女自己接到并传递魔弹）不再询问方向。
func TestMagicBulletControl_DirectionAskedOnlyOncePerBullet(t *testing.T) {
	game := newMagicBulletControlGame(t, testutils.NoopObserver{})
	// 座位 p1(红) p2(红) p3(蓝) p4(蓝)；逆向（索引递增找下一名对手）：p1 → p3 → p1 → p4。
	// 正向会是 p1 → p4 → p2，所以第二跳落在 p1 说明逆向被保留。
	game.State.Players["p3"].Hand = []model.Card{{ID: "mb3", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 2}}
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{1}})
	for _, step := range []struct{ responder, card, next string }{{"p3", "mb3", "p1"}} {
		intr := game.State.PendingInterrupt
		if intr == nil || intr.Type != model.InterruptMagicMissile || intr.PlayerID != step.responder {
			t.Fatalf("expected missile on %s, got %+v", step.responder, intr)
		}
		testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: step.responder, Type: model.CmdRespond, ExtraArgs: []string{"counter"}, CardID: step.card})
		intr = game.State.PendingInterrupt
		if intr != nil && intr.Type == model.InterruptMagicBulletDirection {
			t.Fatalf("direction must not be asked again while passing (after %s)", step.responder)
		}
		if intr == nil || intr.PlayerID != step.next {
			t.Fatalf("expected reverse direction kept, next=%s, got %+v", step.next, intr)
		}
	}
}

// 魔法少女不是施法者、只是接到魔弹并传递时，也不询问方向。
func TestMagicBulletControl_PassingMagicalGirlNotAskedDirection(t *testing.T) {
	game := buildMagicBulletFusionChainGame(t, "magical_girl")
	game.State.Players["p1"].Hand = []model.Card{{ID: "mb", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 2}}
	game.State.Players["p2"].Hand = []model.Card{{ID: "mb2", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 2}}
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdMagic, CardID: testutils.PlayableCardID(t, game, "p1", 0)})
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"counter"}, CardID: "mb2"})
	if intr := game.State.PendingInterrupt; intr == nil || intr.Type != model.InterruptMagicMissile || intr.PlayerID != "p3" {
		t.Fatalf("expected missile passed to p3 without direction prompt, got %+v", intr)
	}
}
