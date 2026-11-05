package prayer_master_test

import (
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/testutils"
)

// 威力赐福：命中后选"是"，本次攻击伤害 +2（2 点攻击 → 目标摸 4 张）。
func TestPrayerPowerBlessing_AddsTwoDamageToCurrentAttack(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Attacker", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{{ID: "a1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2}}
	p1.AddFieldCard(&model.FieldCard{
		Card:     model.Card{ID: "pb", Name: "威力赐福", Type: model.CardTypeMagic},
		OwnerID:  p1.ID,
		SourceID: "prayer",
		Mode:     model.FieldEffect,
		Effect:   model.EffectPowerBlessing,
	})
	p2.Hand = nil
	for i := 0; i < 8; i++ {
		game.State.Deck = append(game.State.Deck, model.Card{ID: "d" + string(rune('a'+i)), Name: "补牌", Type: model.CardTypeAttack, Element: model.ElementWater, Damage: 2})
	}
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardID: testutils.PlayableCardID(t, game, "p1", 0)})
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"take"}})
	testutils.RequireChoicePrompt(t, game, "p1", "prayer_power_blessing_response")
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0}})

	if testutils.HasFieldEffect(p1, model.EffectPowerBlessing) {
		t.Fatalf("power blessing should be removed after use")
	}
	if got := len(p2.Hand); got != 4 {
		t.Fatalf("expected 2+2=4 damage (target draws 4), got hand=%d", got)
	}
}

func setupRadiantFaith(t *testing.T, hand []model.Card) (*engine.GameEngine, *model.Player, *model.Player) {
	t.Helper()
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Prayer", "prayer_master", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Form = model.FormPrayerMasterPrayer
	p1.Tokens["prayer_rune"] = 1
	p1.Hand = hand
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	return game, p1, game.State.Players["p2"]
}

// 光辉信仰：弃 2 张是效果，手牌不足时部分发动（1 张弃 1 张，0 张不弃）。
func TestPrayerRadiantFaith_PartialDiscardWhenHandShort(t *testing.T) {
	game, p1, ally := setupRadiantFaith(t, []model.Card{{ID: "h1", Name: "水涟斩", Type: model.CardTypeAttack, Element: model.ElementWater, Damage: 2}})
	if err := game.UseSkill("p1", "prayer_radiant_faith", []string{"p2"}, []int{0}); err != nil {
		t.Fatalf("radiant faith with 1 card should be usable, got %v", err)
	}
	if len(p1.Hand) != 0 || ally.Heal != 1 || p1.Tokens["prayer_rune"] != 0 {
		t.Fatalf("unexpected result hand=%d allyHeal=%d rune=%d", len(p1.Hand), ally.Heal, p1.Tokens["prayer_rune"])
	}

	game, _, ally = setupRadiantFaith(t, nil)
	if err := game.UseSkill("p1", "prayer_radiant_faith", []string{"p2"}, nil); err != nil {
		t.Fatalf("radiant faith with empty hand should be usable, got %v", err)
	}
	if ally.Heal != 1 {
		t.Fatalf("expected ally +1 heal, got %d", ally.Heal)
	}
}
