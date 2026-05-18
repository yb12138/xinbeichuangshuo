package plague_mage_test

import (
	"starcup-engine/internal/engine"
	"starcup-engine/internal/testutils"
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func driveSeveralSteps(game *engine.GameEngine, steps int) {
	for i := 0; i < steps; i++ {
		if game.State.PendingInterrupt != nil {
			return
		}
		game.Drive()
	}
}

func TestPlagueOutbreak_UsesTurnEndRewardInsteadOfImmediateHeal(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Plague", "plague_mage", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p1.Heal = 0
	p1.Hand = []model.Card{
		{ID: "earth-card", Name: "地牌", Type: model.CardTypeMagic, Element: model.ElementEarth},
	}
	p2.Hand = make([]model.Card, game.GetMaxHand(p2))
	for i := range p2.Hand {
		p2.Hand[i] = model.Card{
			ID:      string(rune('a' + i)),
			Name:    "手牌",
			Type:    model.CardTypeAttack,
			Element: model.ElementFire,
			Damage:  1,
		}
	}

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSkill,
		SkillID:    "plague_outbreak",
		Selections: []int{0},
	})

	if p1.Heal != 1 {
		t.Fatalf("expected outbreak to gain only immortal heal during action resolution, got %d", p1.Heal)
	}
	if game.State.PendingInterrupt == nil || !engine.IsDiscardSelectionInterrupt(game.State.PendingInterrupt) || game.State.PendingInterrupt.PlayerID != "p2" {
		t.Fatalf("expected target overflow discard after outbreak damage, got %+v", game.State.PendingInterrupt)
	}

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p2",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})

	driveSeveralSteps(game, 8)

	if p1.Heal != 2 {
		t.Fatalf("expected outbreak to gain +1 from immortal and +1 from turn-end reward, got %d", p1.Heal)
	}
	if got := p1.TurnState.UsedSkillCounts["plague_outbreak_morale_drop"]; got != 0 {
		t.Fatalf("expected outbreak turn reward flag cleared after turn end, got %d", got)
	}
}

func TestPlagueDeathTouch_TargetsEnemyOnlyAndSuppressesImmortal(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Plague", "plague_mage", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p3 := game.State.Players["p3"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p3.TurnState = model.NewPlayerTurnState()
	p1.Heal = 3
	p1.Hand = []model.Card{
		{ID: "fire-a", Name: "火牌A", Type: model.CardTypeMagic, Element: model.ElementFire},
		{ID: "fire-b", Name: "火牌B", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdSkill,
		SkillID:  "plague_death_touch",
		TargetID: "p2",
	}); err == nil {
		t.Fatalf("expected death touch to reject allied target")
	}

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdSkill,
		SkillID:  "plague_death_touch",
		TargetID: "p3",
	})
	ctx := testutils.RequireChoiceContext(t, game, "p1", "plague_death_touch_element")
	flow, ok := ctx[model.PromptFlowContextKey].(*model.PromptFlowState)
	if !ok || flow.FlowID != "plague_death_touch" || flow.StepID != "element" {
		t.Fatalf("expected death touch flow at element step, got %+v", ctx[model.PromptFlowContextKey])
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0}})
	ctx = testutils.RequireChoiceContext(t, game, "p1", "plague_death_touch_x")
	flow, ok = ctx[model.PromptFlowContextKey].(*model.PromptFlowState)
	if !ok || flow.StepID != "x" || flow.Selection("element").Element != string(model.ElementFire) {
		t.Fatalf("expected death touch flow to accumulate element selection, got %+v", flow)
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0}})
	// After X selection, directly multi-select cards (no Y selection anymore)
	ctx = testutils.RequireChoiceContext(t, game, "p1", "plague_death_touch_cards")
	flow, ok = ctx[model.PromptFlowContextKey].(*model.PromptFlowState)
	if !ok || flow.StepID != "cards" || flow.Selection("x").Count != 2 {
		t.Fatalf("expected death touch flow to accumulate X selection, got %+v", flow)
	}
	// Multi-select both fire cards at once
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0, 1}})
	if p1.Heal != 1 {
		t.Fatalf("expected death touch to remove 2 heal immediately after final choice, got %d", p1.Heal)
	}
	if len(p1.Hand) != 0 {
		t.Fatalf("expected death touch to discard exactly 2 same-element cards, remaining hand=%+v", p1.Hand)
	}

	driveSeveralSteps(game, 8)

	if p1.Heal != 1 {
		t.Fatalf("expected death touch to remove 2 heal and suppress immortal, got %d", p1.Heal)
	}
	if got := p1.TurnState.UsedSkillCounts["plague_block_immortal"]; got != 0 {
		t.Fatalf("expected immortal suppression flag cleared after action end, got %d", got)
	}
}

func TestPlagueDeathTouch_CancelChoiceRestoresActionWindow(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Plague", "plague_mage", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p1.Heal = 3
	p1.Hand = []model.Card{
		{ID: "fire-a", Name: "火牌A", Type: model.CardTypeMagic, Element: model.ElementFire},
		{ID: "fire-b", Name: "火牌B", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdSkill,
		SkillID:  "plague_death_touch",
		TargetID: "p2",
	})
	testutils.RequireChoicePrompt(t, game, "p1", "plague_death_touch_element")

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdCancel,
	})

	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected death touch prompt cleared after cancel, got %+v", game.State.PendingInterrupt)
	}
	if game.State.TurnStage != model.TurnStageActionExecution {
		t.Fatalf("expected to return to action execution after cancel, got %s", game.State.TurnStage)
	}
	if p1.Heal != 3 {
		t.Fatalf("expected heal unchanged after cancel, got %d", p1.Heal)
	}
	if len(p1.Hand) != 2 {
		t.Fatalf("expected hand unchanged after cancel, got %+v", p1.Hand)
	}
	if got := p1.TurnState.UsedSkillCounts["plague_block_immortal"]; got != 0 {
		t.Fatalf("expected immortal suppression flag reset on cancel, got %d", got)
	}
	if p1.TurnState.HasActed {
		t.Fatalf("expected cancel not to consume action")
	}
}

func TestPlagueDeathTouch_CancelAtCardPickRestoresActionWindow(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Plague", "plague_mage", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p1.Heal = 4
	p1.Hand = []model.Card{
		{ID: "fire-a", Name: "火牌A", Type: model.CardTypeMagic, Element: model.ElementFire},
		{ID: "fire-b", Name: "火牌B", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
		{ID: "fire-c", Name: "火牌C", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdSkill,
		SkillID:  "plague_death_touch",
		TargetID: "p2",
	})
	testutils.RequireChoicePrompt(t, game, "p1", "plague_death_touch_element")
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0}})
	testutils.RequireChoicePrompt(t, game, "p1", "plague_death_touch_x")
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0}})
	// After X selection, directly go to multi-select cards (no Y selection)
	testutils.RequireChoicePrompt(t, game, "p1", "plague_death_touch_cards")

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdCancel,
	})

	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected death touch prompt cleared after cancel at card pick, got %+v", game.State.PendingInterrupt)
	}
	if game.State.TurnStage != model.TurnStageActionExecution {
		t.Fatalf("expected to return to action execution after cancel at card pick, got %s", game.State.TurnStage)
	}
	if p1.Heal != 4 {
		t.Fatalf("expected heal unchanged after cancel at card pick, got %d", p1.Heal)
	}
	if len(p1.Hand) != 3 {
		t.Fatalf("expected hand unchanged after cancel at card pick, got %+v", p1.Hand)
	}
	if got := p1.TurnState.UsedSkillCounts["plague_block_immortal"]; got != 0 {
		t.Fatalf("expected immortal suppression flag reset on cancel at card pick, got %d", got)
	}
	if p1.TurnState.HasActed {
		t.Fatalf("expected cancel at card pick not to consume action")
	}
}

func TestPlagueToxicNova_ConsumesExactlyOneGem(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Plague", "plague_mage", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p1.Gem = 1

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdSkill,
		SkillID:  "plague_toxic_nova",
	})

	if p1.Gem != 0 {
		t.Fatalf("expected toxic nova to consume exactly 1 gem, got %d", p1.Gem)
	}
	if p1.Heal != 2 {
		t.Fatalf("expected toxic nova to grant +1治疗 and dispatch immortal once, got %d", p1.Heal)
	}
}
