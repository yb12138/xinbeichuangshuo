package angel_test

import (
	"testing"

	"starcup-engine/internal/engine"
	skillhandlers "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
	"starcup-engine/internal/testutils"
)

func newAngelGame(t *testing.T) (*engine.GameEngine, *model.Player, *model.Player) {
	t.Helper()
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Angel", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	return game, p1, p2
}

func pendingChoiceType(game *engine.GameEngine) string {
	if game.State.PendingInterrupt == nil {
		return ""
	}
	return testutils.ChoiceTypeOfInterrupt(game.State.PendingInterrupt)
}

// 天使羁绊：放置中毒不算"移除基础效果"或"使用圣盾"，不触发。
func TestAngelBond_PlacingPoisonDoesNotTrigger(t *testing.T) {
	game, p1, p2 := newAngelGame(t)
	p1.Hand = []model.Card{
		{ID: "poison-1", Name: "中毒", Type: model.CardTypeMagic, Element: model.ElementEarth},
		{ID: "x", Name: "水涟斩", Type: model.CardTypeAttack, Element: model.ElementWater, Damage: 2},
	}
	if err := game.PerformMagic("p1", "p2", 0); err != nil {
		t.Fatalf("cast poison failed: %v", err)
	}
	if !testutils.HasFieldEffect(p2, model.EffectPoison) {
		t.Fatalf("expected poison placed on enemy")
	}
	if got := pendingChoiceType(game); got == "angel_bond_heal_target" {
		t.Fatalf("placing poison must not trigger angel bond")
	}
}

// 天使羁绊：自己身上的封印被动移除（伤害结算后自动移除）不触发。
func TestAngelBond_PassiveSealRemovalDoesNotTrigger(t *testing.T) {
	game, p1, _ := newAngelGame(t)
	p1.AddFieldCard(&model.FieldCard{
		Card:     model.Card{ID: "seal", Name: "火之封印", Type: model.CardTypeMagic},
		OwnerID:  p1.ID,
		SourceID: "p2",
		Mode:     model.FieldEffect,
		Effect:   model.EffectSealFire,
	})
	if !game.RemoveFieldCard(p1.ID, model.EffectSealFire) {
		t.Fatalf("expected seal removed")
	}
	if got := pendingChoiceType(game); got == "angel_bond_heal_target" {
		t.Fatalf("passive seal removal must not trigger angel bond")
	}
}

// 风之洁净：场上没有基础效果时不能发动（CanUse 决定服务端按钮置灰）。
func TestAngelCleanse_RequiresBasicEffectOnField(t *testing.T) {
	game, p1, p2 := newAngelGame(t)
	p1.Hand = []model.Card{{ID: "w1", Name: "风神斩", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 2}}
	handler := skillhandlers.GetHandler("angel_cleanse")
	if handler == nil {
		t.Fatalf("angel_cleanse handler not registered")
	}
	ctx := &model.Context{Game: game, User: p1, Timing: model.TimingActionDuring, Selections: map[string]any{}, Flags: map[string]bool{}}
	if handler.CanUse(ctx) {
		t.Fatalf("angel_cleanse must not be usable with no basic effect on field")
	}

	p2.AddFieldCard(&model.FieldCard{
		Card:     model.Card{ID: "weak", Name: "虚弱", Type: model.CardTypeMagic, Element: model.ElementWind},
		OwnerID:  p2.ID,
		SourceID: p1.ID,
		Mode:     model.FieldEffect,
		Effect:   model.EffectWeak,
		Hook:     model.FieldHookOnBeforeAction,
	})
	if !handler.CanUse(ctx) {
		t.Fatalf("angel_cleanse should be usable with a basic effect on field")
	}
	if err := game.UseSkill("p1", "angel_cleanse", nil, []int{0}); err != nil {
		t.Fatalf("angel_cleanse failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "basic_effect_pick")
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0}})
	if testutils.HasFieldEffect(p2, model.EffectWeak) {
		t.Fatalf("expected weak removed")
	}
	// 主动移除仍然触发天使羁绊。
	testutils.RequireChoicePrompt(t, game, "p1", "angel_bond_heal_target")
}
