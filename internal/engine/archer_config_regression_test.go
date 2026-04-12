package engine

import (
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

type archerCaptureObserver struct {
	events []model.GameEvent
}

func (o *archerCaptureObserver) OnGameEvent(event model.GameEvent) {
	o.events = append(o.events, event)
}

func (o *archerCaptureObserver) findPublicDiscardReveal(playerID string) map[string]interface{} {
	for i := len(o.events) - 1; i >= 0; i-- {
		event := o.events[i]
		if event.Type != model.EventCardRevealed {
			continue
		}
		data, ok := event.Data.(map[string]interface{})
		if !ok {
			continue
		}
		actionType, _ := data["action_type"].(string)
		eventPlayerID, _ := data["player_id"].(string)
		hidden, _ := data["hidden"].(bool)
		if eventPlayerID == playerID && actionType == "discard" && !hidden {
			return data
		}
	}
	return nil
}

func TestArcher_PiercingShotDiscard_IsPublicReveal(t *testing.T) {
	obs := &archerCaptureObserver{}
	game := NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Archer", "archer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Ally", "angel", model.RedCamp); err != nil {
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
	p2.Heal = 0

	p1.Hand = []model.Card{
		{ID: "atk", Name: "箭", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 2},
		{ID: "spell", Name: "水镜", Type: model.CardTypeMagic, Element: model.ElementWater},
	}
	p2.Hand = []model.Card{
		{ID: "counter", Name: "风斩", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 1},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0,
	}); err != nil {
		t.Fatalf("archer attack failed: %v", err)
	}
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"counter"}, CardIndex: 0, TargetID: "p3",
	}); err != nil {
		t.Fatalf("counter response failed: %v", err)
	}
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0},
	}); err != nil {
		t.Fatalf("confirm piercing shot failed: %v", err)
	}
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0},
	}); err != nil {
		t.Fatalf("discard for piercing shot failed: %v", err)
	}

	reveal := obs.findPublicDiscardReveal("p1")
	if reveal == nil {
		t.Fatalf("expected piercing shot discard to emit a public discard reveal event")
	}
	if len(p1.Hand) != 0 {
		t.Fatalf("expected piercing shot discard to consume the spell card, got %d cards", len(p1.Hand))
	}
}

func TestArcher_PiercingShot_NotPromptedOnHit(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Archer", "archer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Ally", "angel", model.RedCamp); err != nil {
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
	p2.Heal = 0

	p1.Hand = []model.Card{
		{ID: "atk", Name: "箭", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 2},
		{ID: "spell", Name: "水镜", Type: model.CardTypeMagic, Element: model.ElementWater},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0,
	}); err != nil {
		t.Fatalf("archer attack failed: %v", err)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"take"},
	}); err != nil {
		t.Fatalf("take response failed: %v", err)
	}

	if game.State.PendingInterrupt != nil &&
		game.State.PendingInterrupt.Type == model.InterruptResponseSkill &&
		game.State.PendingInterrupt.PlayerID == "p1" &&
		hasSkillID(game.State.PendingInterrupt.SkillIDs, "piercing_shot") {
		t.Fatalf("piercing_shot should not dispatch on hit branch")
	}
}

func TestArcher_LightningArrow_DisablesCounterButAllowsDefend(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Archer", "archer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Ally", "angel", model.RedCamp); err != nil {
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
	p2.Heal = 0

	p1.Hand = []model.Card{
		{ID: "thunder-atk", Name: "雷矢", Type: model.CardTypeAttack, Element: model.ElementThunder, Damage: 2},
	}
	p2.Hand = []model.Card{
		{ID: "holy-light", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight},
		{ID: "counter", Name: "雷刃", Type: model.CardTypeAttack, Element: model.ElementThunder, Damage: 1},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0,
	}); err != nil {
		t.Fatalf("lightning arrow attack failed: %v", err)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"counter"}, CardIndex: 1, TargetID: "p3",
	}); err == nil {
		t.Fatalf("expected lightning arrow to forbid counter response")
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"defend"}, CardIndex: 0,
	}); err != nil {
		t.Fatalf("expected lightning arrow attack to still allow defend: %v", err)
	}

	if len(p2.Hand) != 1 || p2.Hand[0].Name != "雷刃" {
		t.Fatalf("expected only holy light to be consumed on defend, remaining hand=%+v", p2.Hand)
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected defend path to resolve cleanly, got pending interrupt %+v", game.State.PendingInterrupt)
	}
}

func TestArcher_PreciseShot_AutoAppliesForceHit(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Archer", "archer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Ally", "angel", model.RedCamp); err != nil {
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
	p2.Heal = 0

	p1.Hand = []model.Card{{
		ID:              "precise-shot-card",
		Name:            "精准射击",
		Type:            model.CardTypeAttack,
		Element:         model.ElementWind,
		Damage:          2,
		ExclusiveChar1:  "archer",
		ExclusiveSkill1: "精准射击",
	}}
	p2.Hand = []model.Card{
		{ID: "holy-light", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight},
		{ID: "counter", Name: "风斩", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 1},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0,
	}); err != nil {
		t.Fatalf("precise shot attack failed: %v", err)
	}

	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected precise shot to auto-resolve without response prompt, got %+v", game.State.PendingInterrupt)
	}
	if len(p2.Hand) != 3 {
		t.Fatalf("expected precise shot to force-hit for 1 damage draw, got hand=%d", len(p2.Hand))
	}
	if p2.Hand[0].Name != "圣光" || p2.Hand[1].Name != "风斩" {
		t.Fatalf("expected defender response cards to remain unused, hand=%+v", p2.Hand)
	}
}
