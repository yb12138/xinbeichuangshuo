package engine

import (
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func findPublicDiscardReveal(obs *captureObserver, playerID string) map[string]interface{} {
	if obs == nil {
		return nil
	}
	for i := len(obs.events) - 1; i >= 0; i-- {
		event := obs.events[i]
		if event.Type != model.EventCardRevealed {
			continue
		}
		data, ok := event.Data.(map[string]interface{})
		if !ok {
			continue
		}
		eventPlayerID, _ := data["player_id"].(string)
		actionType, _ := data["action_type"].(string)
		hidden, _ := data["hidden"].(bool)
		if eventPlayerID == playerID && actionType == "discard" && !hidden {
			return data
		}
	}
	return nil
}

func startupSkillIndexByID(t *testing.T, game *GameEngine, playerID, skillID string) int {
	t.Helper()
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptStartupSkill {
		t.Fatalf("expected startup interrupt, got %+v", game.State.PendingInterrupt)
	}
	if game.State.PendingInterrupt.PlayerID != playerID {
		t.Fatalf("expected startup interrupt for %s, got %s", playerID, game.State.PendingInterrupt.PlayerID)
	}
	for i, sid := range game.State.PendingInterrupt.SkillIDs {
		if sid == skillID {
			return i
		}
	}
	t.Fatalf("skill %s not found in startup skills: %+v", skillID, game.State.PendingInterrupt.SkillIDs)
	return -1
}

func TestAssassinWaterShadow_InterruptsNormalDrawAndPublicReveal(t *testing.T) {
	obs := &captureObserver{}
	game := NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Assassin", "assassin", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{ID: "water-1", Name: "水弹", Type: model.CardTypeMagic, Element: model.ElementWater},
		{ID: "fire-1", Name: "火球", Type: model.CardTypeMagic, Element: model.ElementFire},
	}

	game.DrawCards("p1", 2)

	requireResponseSkillPrompt(t, game, "p1")
	chooseResponseSkillByID(t, game, "p1", "water_shadow")
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptDiscard {
		t.Fatalf("expected discard interrupt for water shadow, got %+v", game.State.PendingInterrupt)
	}

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})

	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected draw replacement to resolve cleanly, got %+v", game.State.PendingInterrupt)
	}
	if got := len(p1.Hand); got != 2 {
		t.Fatalf("expected hand size 2 after replacing 1 draw with 1 public discard, got %d", got)
	}
	if reveal := findPublicDiscardReveal(obs, "p1"); reveal == nil {
		t.Fatalf("expected water shadow discard to emit public reveal event")
	}
}

func TestAssassinStealth_DrawChoiceDelaysStealthUntilDrawResolves(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Assassin", "assassin", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionStart

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 1
	p1.Hand = []model.Card{
		{ID: "water-1", Name: "水弹", Type: model.CardTypeMagic, Element: model.ElementWater},
		{ID: "magic-1", Name: "炎爆", Type: model.CardTypeMagic, Element: model.ElementFire},
	}

	game.Drive()
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{startupSkillIndexByID(t, game, "p1", "stealth")},
	})

	requireChoicePrompt(t, game, "p1", "assassin_stealth_draw")
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})

	requireResponseSkillPrompt(t, game, "p1")
	if hasAssassinStealthForm(p1) {
		t.Fatalf("stealth should not apply before the optional draw path fully resolves")
	}

	chooseResponseSkillByID(t, game, "p1", "water_shadow")
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptDiscard {
		t.Fatalf("expected discard interrupt after choosing water shadow, got %+v", game.State.PendingInterrupt)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0, 1},
	}); err == nil {
		t.Fatalf("expected extra magic discard to be rejected before stealth actually applies")
	}

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})

	if !hasAssassinStealthForm(p1) {
		t.Fatalf("expected stealth followup to apply after the interrupted draw path resolves")
	}
	if got := len(p1.Hand); got != 1 {
		t.Fatalf("expected draw to be replaced by 1 discard, leaving 1 card in hand, got %d", got)
	}
	if got := game.GetMaxHand(p1); got != 5 {
		t.Fatalf("expected stealth hand limit to be 5, got %d", got)
	}
}

func TestAssassinStealth_HandLimitMinusOneAndReleasesNextStartup(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Assassin", "assassin", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionStart

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 1
	p1.Hand = []model.Card{
		{ID: "c1", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
		{ID: "c2", Name: "水斩", Type: model.CardTypeAttack, Element: model.ElementWater, Damage: 1},
		{ID: "c3", Name: "风斩", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 1},
		{ID: "c4", Name: "雷斩", Type: model.CardTypeAttack, Element: model.ElementThunder, Damage: 1},
		{ID: "c5", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight},
		{ID: "c6", Name: "暗影", Type: model.CardTypeMagic, Element: model.ElementDark},
	}

	game.Drive()
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{startupSkillIndexByID(t, game, "p1", "stealth")},
	})

	requireChoicePrompt(t, game, "p1", "assassin_stealth_draw")
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{1},
	})

	if !hasAssassinStealthForm(p1) {
		t.Fatalf("expected stealth to apply immediately on the no-draw branch")
	}
	if got := game.GetMaxHand(p1); got != 5 {
		t.Fatalf("expected stealth to reduce max hand to 5, got %d", got)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptDiscard {
		t.Fatalf("expected overflow discard after entering stealth with 6 cards, got %+v", game.State.PendingInterrupt)
	}

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})

	if got := len(p1.Hand); got != 5 {
		t.Fatalf("expected hand size 5 after stealth overflow discard, got %d", got)
	}

	p1.TurnState = model.NewPlayerTurnState()
	game.State.PendingInterrupt = nil
	game.State.InterruptQueue = nil
	game.State.TurnStage = model.TurnStageActionStart
	game.State.CurrentTurn = 0

	game.Drive()

	if hasAssassinStealthForm(p1) {
		t.Fatalf("expected stealth to release at next startup")
	}
	if got := game.GetMaxHand(p1); got != 6 {
		t.Fatalf("expected max hand to return to 6 after release, got %d", got)
	}
}

func TestAssassinStealth_DrawSkipResponse_ResumesStartupFlow(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Assassin", "assassin", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionStart

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 1
	p1.Hand = []model.Card{
		{ID: "water-1", Name: "水弹", Type: model.CardTypeMagic, Element: model.ElementWater},
	}

	game.Drive()
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{startupSkillIndexByID(t, game, "p1", "stealth")},
	})

	requireChoicePrompt(t, game, "p1", "assassin_stealth_draw")
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})

	requireResponseSkillPrompt(t, game, "p1")
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{len(game.State.PendingInterrupt.SkillIDs)},
	})

	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected skip response to finish the stealth draw path cleanly, got %+v", game.State.PendingInterrupt)
	}
	if !hasAssassinStealthForm(p1) {
		t.Fatalf("expected stealth to apply after skipping water shadow during stealth draw")
	}
	if got := len(p1.Hand); got != 2 {
		t.Fatalf("expected skipped water shadow to allow the draw (hand=2), got %d", got)
	}
	if game.State.CombatStage == model.CombatStageCalcDamage ||
		game.State.CombatStage == model.CombatStageHeal ||
		game.State.CombatStage == model.CombatStageApply ||
		game.State.CombatStage == model.CombatStageDraw {
		t.Fatalf("stealth draw should not incorrectly fall into damage resolution combat stages")
	}
}
