package spirit_caster_test

import (
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/testutils"
)

func newSpiritCasterNoHandGame(t *testing.T, obs model.GameObserver, costCard model.Card) *engine.GameEngine {
	t.Helper()
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
	p1.Hand = []model.Card{costCard} // 仅有发动成本，弃掉后无手牌
	p1.Gem = 0
	p1.Crystal = 0
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	return game
}

// 灵符师-1：风行弃牌后无手牌，不再弹"无手牌可放置为妖力，是否跳过？"，直接进入风行弃牌。
func TestSpiritCasterTalismanWind_NoHandSkipsIncantConfirm(t *testing.T) {
	game := newSpiritCasterNoHandGame(t, testutils.NoopObserver{}, spiritCasterTestCard("w1", "风符", model.CardTypeMagic, model.ElementWind))
	game.State.Players["p2"].Hand = []model.Card{spiritCasterTestCard("a1", "攻击A1", model.CardTypeAttack, model.ElementFire)}
	game.State.Players["p3"].Hand = []model.Card{spiritCasterTestCard("b1", "攻击B1", model.CardTypeAttack, model.ElementWater)}

	if err := game.UseSkill("p1", "sc_talisman_wind", []string{"p2", "p3"}, []int{0}); err != nil {
		t.Fatalf("use talisman wind failed: %v", err)
	}
	game.Drive()
	testutils.RequireChoicePrompt(t, game, "p3", "sc_talisman_wind_discard")
	testutils.MustHandleAction(t, game, model.PlayerAction{Type: model.CmdSelect, PlayerID: "p3", Selections: []int{0}})
	testutils.RequireChoicePrompt(t, game, "p2", "sc_talisman_wind_discard")
	testutils.MustHandleAction(t, game, model.PlayerAction{Type: model.CmdSelect, PlayerID: "p2", Selections: []int{0}})
	if got := len(game.State.Players["p2"].Hand) + len(game.State.Players["p3"].Hand); got != 0 {
		t.Fatalf("expected both targets discarded by wind, remaining hand=%d", got)
	}
}

// 灵符师-2：雷鸣无手牌且付不起水晶时，不询问灵力崩解，直接造成伤害。
func TestSpiritCasterTalismanThunder_NoHandNoCrystalSkipsCollapsePrompt(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := newSpiritCasterNoHandGame(t, obs, spiritCasterTestCard("t1", "雷符", model.CardTypeMagic, model.ElementThunder))

	if err := game.UseSkill("p1", "sc_talisman_thunder", []string{"p2", "p3"}, []int{0}); err != nil {
		t.Fatalf("use talisman thunder failed: %v", err)
	}
	if intr := game.State.PendingInterrupt; intr != nil {
		if ct := testutils.ChoiceTypeOfInterrupt(intr); ct == "sc_spiritual_collapse_confirm" {
			t.Fatalf("should not ask spiritual collapse without crystal")
		}
	}
	if got := len(game.State.PendingDamageQueue); got != 2 {
		t.Fatalf("expected 2 thunder damages queued, got %d", got)
	}
	if q := game.State.PendingDamageQueue; q[0].TargetID != "p3" || q[1].TargetID != "p2" || q[0].Damage != 1 || q[1].Damage != 1 {
		t.Fatalf("expected reverse-order 1-dmg thunder p3->p2, got %+v", q)
	}
	game.Drive()
	for _, id := range []string{"p2", "p3"} {
		if got := countSpiritCasterDamageEvents(obs, "p1", id, 1, model.MagicAttack); got != 1 {
			t.Fatalf("expected thunder damage resolved once on %s, got %d", id, got)
		}
	}
}

// 灵符师-2：有水晶时仍询问灵力崩解。
func TestSpiritCasterTalismanThunder_NoHandWithCrystalStillPromptsCollapse(t *testing.T) {
	game := newSpiritCasterNoHandGame(t, testutils.NoopObserver{}, spiritCasterTestCard("t1", "雷符", model.CardTypeMagic, model.ElementThunder))
	game.State.Players["p1"].Crystal = 1
	if err := game.UseSkill("p1", "sc_talisman_thunder", []string{"p2", "p3"}, []int{0}); err != nil {
		t.Fatalf("use talisman thunder failed: %v", err)
	}
	game.Drive()
	testutils.RequireChoicePrompt(t, game, "p1", "sc_spiritual_collapse_confirm")
}
