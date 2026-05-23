package sealer_test

import (
	"starcup-engine/internal/engine"
	"starcup-engine/internal/testutils"
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func TestFiveElementsBind_BuffPhaseChoiceUsesSealCountCapAtTwo(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Sealer", "sealer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p2.AddFieldCard(&model.FieldCard{
		Card:     model.Card{ID: "bind", Name: "五系束缚", Type: model.CardTypeMagic, Element: model.ElementLight},
		OwnerID:  p2.ID,
		SourceID: p1.ID,
		Mode:     model.FieldEffect,
		Effect:   model.EffectFiveElementsBind,
		Hook:     model.FieldHookOnBeforeAction,
	})
	p1.AddFieldCard(&model.FieldCard{
		Card:     model.Card{ID: "seal-fire", Name: "火之封印", Type: model.CardTypeMagic, Element: model.ElementFire},
		OwnerID:  p1.ID,
		SourceID: p1.ID,
		Mode:     model.FieldEffect,
		Effect:   model.EffectSealFire,
		Hook:     model.FieldHookOnCardPlayedOrRevealed,
	})
	p1.AddFieldCard(&model.FieldCard{
		Card:     model.Card{ID: "seal-water", Name: "水之封印", Type: model.CardTypeMagic, Element: model.ElementWater},
		OwnerID:  p1.ID,
		SourceID: p1.ID,
		Mode:     model.FieldEffect,
		Effect:   model.EffectSealWater,
		Hook:     model.FieldHookOnCardPlayedOrRevealed,
	})
	p2.AddFieldCard(&model.FieldCard{
		Card:     model.Card{ID: "seal-wind", Name: "风之封印", Type: model.CardTypeMagic, Element: model.ElementWind},
		OwnerID:  p2.ID,
		SourceID: p1.ID,
		Mode:     model.FieldEffect,
		Effect:   model.EffectSealWind,
		Hook:     model.FieldHookOnCardPlayedOrRevealed,
	})

	game.State.CurrentTurn = 1
	game.State.TurnStage = model.TurnStageBeforeAction
	game.Drive()

	testutils.RequireChoicePrompt(t, game, "p2", "five_elements_bind")
	ctxData, _ := game.State.PendingInterrupt.Context.(map[string]interface{})
	drawCount, _ := ctxData["draw_count"].(int)
	if drawCount != 4 {
		t.Fatalf("expected five elements bind draw count to cap at 4, got %d", drawCount)
	}
	if got := testutils.CountFieldEffect(p2, model.EffectFiveElementsBind); got != 1 {
		t.Fatalf("five elements bind should remain until player resolves choice, got %d", got)
	}
}

func TestFiveElementsBind_DrawCancelRemovesStatusAndResumesStartup(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Sealer", "sealer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p2.Hand = nil
	p2.AddFieldCard(&model.FieldCard{
		Card:     model.Card{ID: "bind", Name: "五系束缚", Type: model.CardTypeMagic, Element: model.ElementLight},
		OwnerID:  p2.ID,
		SourceID: p1.ID,
		Mode:     model.FieldEffect,
		Effect:   model.EffectFiveElementsBind,
		Hook:     model.FieldHookOnBeforeAction,
	})
	p1.AddFieldCard(&model.FieldCard{
		Card:     model.Card{ID: "seal-fire", Name: "火之封印", Type: model.CardTypeMagic, Element: model.ElementFire},
		OwnerID:  p1.ID,
		SourceID: p1.ID,
		Mode:     model.FieldEffect,
		Effect:   model.EffectSealFire,
		Hook:     model.FieldHookOnCardPlayedOrRevealed,
	})

	game.State.Deck = rules.InitDeck()
	game.State.CurrentTurn = 1
	game.State.TurnStage = model.TurnStageBeforeAction
	game.Drive()
	testutils.RequireChoicePrompt(t, game, "p2", "five_elements_bind")

	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p2", Selections: []int{0}}); err != nil {
		t.Fatalf("resolve five elements bind draw cancel failed: %v", err)
	}
	if got := testutils.CountFieldEffect(p2, model.EffectFiveElementsBind); got != 0 {
		t.Fatalf("five elements bind should be removed after resolving choice, got %d", got)
	}
	if len(p2.Hand) != 3 {
		t.Fatalf("expected target to draw 3 cards after cancel, got %d", len(p2.Hand))
	}
	if game.State.TurnStage != model.TurnStageActionStart {
		t.Fatalf("expected turn stage to resume at action start after cancel, got %s", game.State.TurnStage)
	}
}

func TestFiveElementsBind_DrawCancelOverflowQueuesDiscardWithoutMismatchLog(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Sealer", "sealer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p2.Hand = []model.Card{
		{ID: "h1", Name: "手牌1", Type: model.CardTypeAttack, Element: model.ElementFire},
		{ID: "h2", Name: "手牌2", Type: model.CardTypeAttack, Element: model.ElementWater},
		{ID: "h3", Name: "手牌3", Type: model.CardTypeAttack, Element: model.ElementWind},
		{ID: "h4", Name: "手牌4", Type: model.CardTypeAttack, Element: model.ElementEarth},
		{ID: "h5", Name: "手牌5", Type: model.CardTypeAttack, Element: model.ElementThunder},
	}
	p2.AddFieldCard(&model.FieldCard{
		Card:     model.Card{ID: "bind", Name: "五系束缚", Type: model.CardTypeMagic, Element: model.ElementLight},
		OwnerID:  p2.ID,
		SourceID: p1.ID,
		Mode:     model.FieldEffect,
		Effect:   model.EffectFiveElementsBind,
		Hook:     model.FieldHookOnBeforeAction,
	})
	p1.AddFieldCard(&model.FieldCard{
		Card:     model.Card{ID: "seal-fire", Name: "火之封印", Type: model.CardTypeMagic, Element: model.ElementFire},
		OwnerID:  p1.ID,
		SourceID: p1.ID,
		Mode:     model.FieldEffect,
		Effect:   model.EffectSealFire,
		Hook:     model.FieldHookOnCardPlayedOrRevealed,
	})

	game.State.Deck = rules.InitDeck()
	game.State.CurrentTurn = 1
	game.State.TurnStage = model.TurnStageBeforeAction
	game.Drive()
	testutils.RequireChoicePrompt(t, game, "p2", "five_elements_bind")

	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p2", Selections: []int{0}}); err != nil {
		t.Fatalf("resolve five elements bind draw cancel failed: %v", err)
	}

	if obs.CountLogContains("EnterDiscardSelection: 缺少与弃牌子流程匹配的 PendingInterrupt") != 0 {
		t.Fatalf("unexpected discard subflow mismatch log while overflow discard was queued")
	}
	if game.State.PendingInterrupt == nil || !engine.IsDiscardSelectionInterrupt(game.State.PendingInterrupt) {
		t.Fatalf("expected queued overflow discard interrupt to become pending, got %+v", game.State.PendingInterrupt)
	}
	if game.State.Subflow != model.SubflowDiscardSelection {
		t.Fatalf("expected discard subflow after queued overflow interrupt activates, got %s", game.State.Subflow)
	}
	ctxData, _ := game.State.PendingInterrupt.Context.(map[string]interface{})
	if got := ctxData["discard_count"]; got != 2 {
		t.Fatalf("expected overflow discard_count=2, got %v", got)
	}
}

func TestElementalSeal_RevealedDiscardRunsButHiddenDiscardDoesNot(t *testing.T) {
	newGame := func() *engine.GameEngine {
		game := engine.NewGameEngine(testutils.NoopObserver{})
		if err := game.AddPlayer("p1", "Target", "berserker", model.RedCamp); err != nil {
			t.Fatal(err)
		}
		if err := game.AddPlayer("p2", "Sealer", "sealer", model.BlueCamp); err != nil {
			t.Fatal(err)
		}
		target := game.State.Players["p1"]
		target.AddFieldCard(&model.FieldCard{
			Card:     model.Card{ID: "seal-fire", Name: "火之封印", Type: model.CardTypeMagic, Element: model.ElementFire},
			OwnerID:  target.ID,
			SourceID: "p2",
			Mode:     model.FieldEffect,
			Effect:   model.EffectSealFire,
			Hook:     model.FieldHookOnCardPlayedOrRevealed,
		})
		return game
	}

	game := newGame()
	game.NotifyCardRevealed("p1", []model.Card{{
		ID:      "fire-card",
		Name:    "火牌",
		Type:    model.CardTypeMagic,
		Element: model.ElementFire,
	}}, "discard")
	if len(game.State.PendingDamageQueue) != 1 {
		t.Fatalf("revealed discard should dispatch elemental seal, got %d pending damages", len(game.State.PendingDamageQueue))
	}

	game = newGame()
	game.NotifyCardHidden("p1", []model.Card{{
		ID:      "fire-card",
		Name:    "火牌",
		Type:    model.CardTypeMagic,
		Element: model.ElementFire,
	}}, "discard")
	if len(game.State.PendingDamageQueue) != 0 {
		t.Fatalf("hidden discard should not dispatch elemental seal, got %d pending damages", len(game.State.PendingDamageQueue))
	}
}

func TestElementalSeal_UsesBoundElementMetaForMatching(t *testing.T) {
	newGame := func() *engine.GameEngine {
		game := engine.NewGameEngine(testutils.NoopObserver{})
		if err := game.AddPlayer("p1", "Target", "berserker", model.RedCamp); err != nil {
			t.Fatal(err)
		}
		if err := game.AddPlayer("p2", "Sealer", "sealer", model.BlueCamp); err != nil {
			t.Fatal(err)
		}
		target := game.State.Players["p1"]
		target.AddFieldCard(&model.FieldCard{
			Card:     model.Card{ID: "seal-meta", Name: "元素封印", Type: model.CardTypeMagic, Element: model.ElementFire},
			OwnerID:  target.ID,
			SourceID: "p2",
			Mode:     model.FieldEffect,
			Effect:   model.EffectSealFire,
			Hook:     model.FieldHookOnCardPlayedOrRevealed,
			Meta: map[string]string{
				model.FieldMetaBoundElement: string(model.ElementWater),
			},
		})
		return game
	}

	game := newGame()
	game.NotifyCardRevealed("p1", []model.Card{{
		ID:      "water-card",
		Name:    "水牌",
		Type:    model.CardTypeMagic,
		Element: model.ElementWater,
	}}, "discard")
	if len(game.State.PendingDamageQueue) != 1 {
		t.Fatalf("bound element meta should allow water reveal to dispatch, got %d pending damages", len(game.State.PendingDamageQueue))
	}
	if game.State.PendingDamageQueue[0].EffectTypeToRemove != model.EffectSealFire {
		t.Fatalf("expected pending damage to remove original seal effect, got %s", game.State.PendingDamageQueue[0].EffectTypeToRemove)
	}

	game = newGame()
	game.NotifyCardRevealed("p1", []model.Card{{
		ID:      "fire-card",
		Name:    "火牌",
		Type:    model.CardTypeMagic,
		Element: model.ElementFire,
	}}, "discard")
	if len(game.State.PendingDamageQueue) != 0 {
		t.Fatalf("bound element meta should override effect enum match, got %d pending damages", len(game.State.PendingDamageQueue))
	}
}

func TestSealBreak_CanPickGlobalBasicEffectWithoutPreselectedTarget(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Sealer", "sealer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "EnemyA", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "EnemyB", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p3 := game.State.Players["p3"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1
	p2.AddFieldCard(&model.FieldCard{
		Card:     model.Card{ID: "shield-card", Name: "圣盾", Type: model.CardTypeMagic, Element: model.ElementLight},
		OwnerID:  p2.ID,
		SourceID: "p3",
		Mode:     model.FieldEffect,
		Effect:   model.EffectShield,
		Hook:     model.FieldHookOnDamaged,
	})
	p3.AddFieldCard(&model.FieldCard{
		Card:     model.Card{ID: "seal-fire-card", Name: "火之封印", Type: model.CardTypeMagic, Element: model.ElementFire},
		OwnerID:  p3.ID,
		SourceID: "p2",
		Mode:     model.FieldEffect,
		Effect:   model.EffectSealFire,
		Hook:     model.FieldHookOnCardPlayedOrRevealed,
	})

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdSkill,
		SkillID:  "seal_break",
	}); err != nil {
		t.Fatalf("seal_break without target should succeed, got %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "basic_effect_pick")

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{1},
	}); err != nil {
		t.Fatalf("seal_break global choice should succeed, got %v", err)
	}

	if len(p1.Hand) != 1 || p1.Hand[0].ID != "seal-fire-card" {
		t.Fatalf("expected seal_break to take selected global effect, got hand=%+v", p1.Hand)
	}
	if got := testutils.CountFieldEffect(p3, model.EffectSealFire); got != 0 {
		t.Fatalf("expected selected global effect to be removed from p3, got %d", got)
	}
	if got := testutils.CountFieldEffect(p2, model.EffectShield); got != 1 {
		t.Fatalf("expected unselected effect on p2 to remain, got %d", got)
	}
}
