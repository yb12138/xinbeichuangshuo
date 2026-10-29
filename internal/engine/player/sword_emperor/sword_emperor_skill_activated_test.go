package sword_emperor_test

import (
	"reflect"
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/testutils"
)

// runSwordEmperorDefendedAttack 让 p1 主动攻击 p2，p2 用圣光抵挡，触发剑帝未命中结算。
func runSwordEmperorDefendedAttack(t *testing.T, game *engine.GameEngine) {
	t.Helper()
	prepareSwordEmperorActiveTurn(game, "p1", []model.Card{
		swordEmperorTestCard("atk1", "炎斩", model.CardTypeAttack, model.ElementFire, 2),
	})
	game.State.Players["p2"].Hand = []model.Card{
		swordEmperorTestCard("def1", "圣光", model.CardTypeMagic, model.ElementLight, 0),
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdAttack,
		TargetID: "p2",
		CardID:   testutils.PlayableCardID(t, game, "p1", 0),
	})
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		CardID:    testutils.PlayableCardID(t, game, "p2", 0),
		ExtraArgs: []string{"defend"},
	})
}

func newSwordEmperorObservedGame(t *testing.T) (*engine.GameEngine, *testutils.CaptureObserver) {
	t.Helper()
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "SwordEmperor", "sword_emperor", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	return game, obs
}

func TestSwordEmperor_MissEmitsFeintAndSwordSoulGuardSkillActivated(t *testing.T) {
	game, obs := newSwordEmperorObservedGame(t)
	runSwordEmperorDefendedAttack(t, game)

	for _, skillID := range []string{"se_feint", "se_sword_soul_guard"} {
		events := obs.SkillActivated(skillID)
		if len(events) != 1 {
			t.Fatalf("expected exactly one skill_activated for %s, got %d", skillID, len(events))
		}
		if events[0].PlayerID != "p1" || !reflect.DeepEqual(events[0].TargetIDs, []string{"p1"}) {
			t.Fatalf("unexpected %s payload: %+v", skillID, events[0])
		}
	}
}

func TestSwordEmperor_SwordSoulGuardAtCapDoesNotEmit(t *testing.T) {
	game, obs := newSwordEmperorObservedGame(t)
	for i := 0; i < 3; i++ {
		giveSwordEmperorSwordSoul(t, game, "p1", "soul_cap_"+string(rune('a'+i)))
	}
	runSwordEmperorDefendedAttack(t, game)

	if got := len(obs.SkillActivated("se_sword_soul_guard")); got != 0 {
		t.Fatalf("expected no sword soul guard event at cap, got %d", got)
	}
	if got := len(obs.SkillActivated("se_feint")); got != 1 {
		t.Fatalf("expected one feint event at sword soul cap, got %d", got)
	}
}
