package engine

import (
	"strings"
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func TestBuffResolve_PoisonResolvesBeforeWeaknessChoice(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Target", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.Deck = rules.InitDeck()
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageBeforeAction

	p1 := game.State.Players["p1"]
	p1.TurnState = model.NewPlayerTurnState()
	p1.Heal = 0
	// 故意先放虚弱再放中毒，确认行动开始前仍是“中毒先、虚弱后”。
	p1.AddFieldCard(&model.FieldCard{
		Card:     model.Card{ID: "weak-1", Name: "虚弱", Type: model.CardTypeMagic, Element: model.ElementWind},
		OwnerID:  p1.ID,
		SourceID: "p2",
		Mode:     model.FieldEffect,
		Effect:   model.EffectWeak,
		Trigger:  model.EffectTriggerOnBeforeAction,
	})
	p1.AddFieldCard(&model.FieldCard{
		Card:     model.Card{ID: "poison-1", Name: "中毒", Type: model.CardTypeMagic, Element: model.ElementEarth},
		OwnerID:  p1.ID,
		SourceID: "p2",
		Mode:     model.FieldEffect,
		Effect:   model.EffectPoison,
		Trigger:  model.EffectTriggerOnBeforeAction,
	})

	game.Drive()

	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptChoice {
		t.Fatalf("expected weakness choice after poison resolution, got %+v", game.State.PendingInterrupt)
	}
	ctxData, _ := game.State.PendingInterrupt.Context.(map[string]interface{})
	if got, _ := ctxData["choice_type"].(string); got != "weak" {
		t.Fatalf("expected weakness choice_type, got %q", got)
	}
	if len(game.State.PendingDamageQueue) != 0 {
		t.Fatalf("poison damage should be fully resolved before weakness prompt, queue=%d", len(game.State.PendingDamageQueue))
	}
	if got := countFieldEffect(p1, model.EffectPoison); got != 0 {
		t.Fatalf("poison should be removed after resolving, got %d", got)
	}
	if got := countFieldEffect(p1, model.EffectWeak); got != 1 {
		t.Fatalf("weakness should remain until player chooses, got %d", got)
	}
	if len(game.State.DiscardPile) == 0 || game.State.DiscardPile[len(game.State.DiscardPile)-1].Name != "中毒" {
		t.Fatalf("expected poison card to be discarded after resolution, discard=%+v", game.State.DiscardPile)
	}
}

func TestBeforeActionHooks_PoisonEntersDamageResolutionBeforeWeakness(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Target", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.TurnState = model.NewPlayerTurnState()
	p1.AddFieldCard(&model.FieldCard{
		Card:     model.Card{ID: "weak-1", Name: "虚弱", Type: model.CardTypeMagic, Element: model.ElementWind},
		OwnerID:  p1.ID,
		SourceID: "p2",
		Mode:     model.FieldEffect,
		Effect:   model.EffectWeak,
		Trigger:  model.EffectTriggerOnBeforeAction,
	})
	p1.AddFieldCard(&model.FieldCard{
		Card:     model.Card{ID: "poison-1", Name: "中毒", Type: model.CardTypeMagic, Element: model.ElementEarth},
		OwnerID:  p1.ID,
		SourceID: "p2",
		Mode:     model.FieldEffect,
		Effect:   model.EffectPoison,
		Trigger:  model.EffectTriggerOnBeforeAction,
	})

	if interrupted := game.runTimingOnBeforeActionHooks(p1); !interrupted {
		t.Fatal("expected beforeAction hooks to stop on poison damage resolution")
	}
	if game.State.CombatStage != model.CombatStageCalcDamage {
		t.Fatalf("expected poison hook to enter damage resolution, got combat stage %s", game.State.CombatStage)
	}
	if game.State.ReturnTurnStage != model.TurnStageBeforeAction {
		t.Fatalf("expected poison hook to return to before action, got %s", game.State.ReturnTurnStage)
	}
	if len(game.State.PendingDamageQueue) != 1 {
		t.Fatalf("expected one pending poison damage, got %d", len(game.State.PendingDamageQueue))
	}
	if got := countFieldEffect(p1, model.EffectPoison); got != 0 {
		t.Fatalf("poison should be removed immediately after hook trigger, got %d", got)
	}
	if got := countFieldEffect(p1, model.EffectWeak); got != 1 {
		t.Fatalf("weakness should remain after poison hook, got %d", got)
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no weakness interrupt before poison resolves, got %+v", game.State.PendingInterrupt)
	}

	game.processPendingDamages()
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected poison damage to resolve without interrupt here, got %+v", game.State.PendingInterrupt)
	}
	if !game.restoreReturnPoint() {
		t.Fatal("expected to restore before-action return point after poison resolution")
	}
	if interrupted := game.runTimingOnBeforeActionHooks(p1); !interrupted {
		t.Fatal("expected weakness hook to stop with choice interrupt")
	}
	if got := choiceTypeOfInterrupt(game.State.PendingInterrupt); got != "weak" {
		t.Fatalf("expected weakness interrupt after poison resolution, got %q", got)
	}
}

func TestWeaknessPrompt_OrderMatchesConfig(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Target", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	game.State.PendingInterrupt = &model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: "p1",
		Context: map[string]interface{}{
			"choice_type": "weak",
		},
	}

	prompt := game.buildChoicePrompt()
	if prompt == nil {
		t.Fatal("expected weakness prompt")
	}
	if len(prompt.Options) != 2 {
		t.Fatalf("expected 2 weakness options, got %d", len(prompt.Options))
	}
	if !strings.Contains(prompt.Options[0].Label, "跳过行动阶段") {
		t.Fatalf("expected option 0 to be skip action phase, got %q", prompt.Options[0].Label)
	}
	if !strings.Contains(prompt.Options[1].Label, "摸3张牌") {
		t.Fatalf("expected option 1 to be draw three cards, got %q", prompt.Options[1].Label)
	}
}

func TestWeaknessChoiceMappingMatchesConfig(t *testing.T) {
	newWeakGame := func() *GameEngine {
		game := NewGameEngine(noopObserver{})
		if err := game.AddPlayer("p1", "Target", "angel", model.RedCamp); err != nil {
			t.Fatal(err)
		}
		game.State.Deck = rules.InitDeck()
		game.State.TurnStage = model.TurnStageBeforeAction

		p1 := game.State.Players["p1"]
		p1.TurnState = model.NewPlayerTurnState()
		p1.AddFieldCard(&model.FieldCard{
			Card:     model.Card{ID: "weak-1", Name: "虚弱", Type: model.CardTypeMagic, Element: model.ElementWind},
			OwnerID:  p1.ID,
			SourceID: "p2",
			Mode:     model.FieldEffect,
			Effect:   model.EffectWeak,
			Trigger:  model.EffectTriggerOnBeforeAction,
		})
		game.State.PendingInterrupt = &model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: p1.ID,
			Context: map[string]interface{}{
				"choice_type": "weak",
			},
		}
		return game
	}

	t.Run("skip_action_phase", func(t *testing.T) {
		game := newWeakGame()
		p1 := game.State.Players["p1"]

		if err := game.handleWeakChoiceInput("p1", 0); err != nil {
			t.Fatalf("skip weakness choice failed: %v", err)
		}
		if got := countFieldEffect(p1, model.EffectWeak); got != 0 {
			t.Fatalf("weakness should be removed after skip choice, got %d", got)
		}
		if game.State.TurnStage != model.TurnStageTurnEnd {
			t.Fatalf("skip choice should end turn, got turn stage %s", game.State.TurnStage)
		}
		if len(p1.Hand) != 0 {
			t.Fatalf("skip choice should not draw cards, hand=%d", len(p1.Hand))
		}
	})

	t.Run("draw_three_then_continue", func(t *testing.T) {
		game := newWeakGame()
		p1 := game.State.Players["p1"]

		if err := game.handleWeakChoiceInput("p1", 1); err != nil {
			t.Fatalf("draw weakness choice failed: %v", err)
		}
		if got := countFieldEffect(p1, model.EffectWeak); got != 0 {
			t.Fatalf("weakness should be removed after draw choice, got %d", got)
		}
		if len(p1.Hand) != 3 {
			t.Fatalf("draw choice should add 3 cards, hand=%d", len(p1.Hand))
		}
		if game.State.TurnStage != model.TurnStageActionStart {
			t.Fatalf("draw choice should continue to action start, got turn stage %s", game.State.TurnStage)
		}
	})
}
