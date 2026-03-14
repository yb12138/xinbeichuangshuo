package engine

import (
	"testing"

	"starcup-engine/internal/engine/skills"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

type noopElfBlessingObserver struct{}

func (noopElfBlessingObserver) OnGameEvent(event model.GameEvent) {}

func buildElfBlessingGame(t *testing.T) *GameEngine {
	t.Helper()

	game := NewGameEngine(noopElfBlessingObserver{})
	game.State.Deck = rules.InitDeck()

	if err := game.AddPlayer("p1", "Elf", "elf_archer", model.RedCamp); err != nil {
		t.Fatalf("add p1 failed: %v", err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatalf("add p2 failed: %v", err)
	}

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection
	game.State.PlayerOrder = []string{"p1", "p2"}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()

	return game
}

func TestElfRitualStoresBlessingsOutsideHand(t *testing.T) {
	game := buildElfBlessingGame(t)
	p1 := game.State.Players["p1"]

	p1.Gem = 1
	p1.Hand = []model.Card{
		{ID: "hand-1", Name: "普通手牌1", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
		{ID: "hand-2", Name: "普通手牌2", Type: model.CardTypeAttack, Element: model.ElementWater, Damage: 1},
		{ID: "hand-3", Name: "普通手牌3", Type: model.CardTypeMagic, Element: model.ElementWind, Damage: 0},
		{ID: "hand-4", Name: "普通手牌4", Type: model.CardTypeMagic, Element: model.ElementEarth, Damage: 0},
		{ID: "hand-5", Name: "普通手牌5", Type: model.CardTypeAttack, Element: model.ElementThunder, Damage: 1},
		{ID: "hand-6", Name: "普通手牌6", Type: model.CardTypeMagic, Element: model.ElementLight, Damage: 0},
	}
	beforeHand := len(p1.Hand)

	handler := skills.GetHandler("elf_ritual")
	if handler == nil {
		t.Fatalf("elf_ritual handler not found")
	}
	ctx := &model.Context{
		Game:  game,
		User:  p1,
		Flags: map[string]bool{},
	}
	if !handler.CanUse(ctx) {
		t.Fatalf("elf_ritual should be usable")
	}
	if err := handler.Execute(ctx); err != nil {
		t.Fatalf("elf_ritual execute failed: %v", err)
	}

	if got := len(p1.Hand); got != beforeHand {
		t.Fatalf("ritual should not change normal hand size, got=%d want=%d", got, beforeHand)
	}
	if got := len(p1.Blessings); got != 3 {
		t.Fatalf("ritual should create 3 blessings, got=%d", got)
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("ritual draw should not trigger overflow discard interrupt")
	}
	if p1.Gem != 0 {
		t.Fatalf("ritual should consume 1 gem, got=%d", p1.Gem)
	}
	if p1.Tokens["elf_ritual_form"] != 1 {
		t.Fatalf("elf_ritual_form token should be 1, got=%d", p1.Tokens["elf_ritual_form"])
	}
}

func TestElfBlessingCanBePlayedAsMagic(t *testing.T) {
	game := buildElfBlessingGame(t)
	p1 := game.State.Players["p1"]

	p1.Hand = nil
	p1.Blessings = []model.Card{
		{ID: "bless-magic", Name: "圣盾", Type: model.CardTypeMagic, Element: model.ElementLight, Damage: 0},
	}
	syncElfBlessings(p1)

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdMagic,
		TargetID:  "p1",
		CardIndex: 0, // 手牌为空时，索引0指向第一张祝福
	}); err != nil {
		t.Fatalf("magic with blessing should succeed: %v", err)
	}

	if got := len(p1.Blessings); got != 0 {
		t.Fatalf("blessing should be consumed after play, got=%d", got)
	}
	if p1.HasFieldEffect(model.EffectShield) == false {
		t.Fatalf("blessing magic should resolve to shield field effect")
	}
	if got := len(game.State.DiscardPile); got != 0 {
		t.Fatalf("shield should stay on field instead of discard, got discard=%d", got)
	}
}

func TestElfBlessingCanBePlayedAsAttack(t *testing.T) {
	game := buildElfBlessingGame(t)
	p1 := game.State.Players["p1"]

	p1.Hand = nil
	p1.Blessings = []model.Card{
		// 使用暗系避免触发「元素射击」中断，聚焦验证“祝福可作为攻击牌打出”。
		{ID: "bless-attack", Name: "祝福之刃", Type: model.CardTypeAttack, Element: model.ElementDark, Damage: 1},
	}
	syncElfBlessings(p1)

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0, // 手牌为空时，索引0指向第一张祝福
	}); err != nil {
		t.Fatalf("attack with blessing should enqueue action: %v", err)
	}

	game.Drive()

	if got := len(p1.Blessings); got != 0 {
		t.Fatalf("blessing should be consumed after attack, got=%d", got)
	}
	if got := len(game.State.CombatStack); got != 1 {
		t.Fatalf("combat stack should have 1 request, got=%d", got)
	}
	if game.State.CombatStack[0].Card == nil || game.State.CombatStack[0].Card.ID != "bless-attack" {
		t.Fatalf("combat card should be the blessing attack card")
	}
	if got := len(game.State.DiscardPile); got != 1 {
		t.Fatalf("discard pile should include used blessing attack, got=%d", got)
	}
}

func TestElfRitualStartupConfirmShouldNotLeaveOverflowDiscard(t *testing.T) {
	game := buildElfBlessingGame(t)
	p1 := game.State.Players["p1"]

	p1.Gem = 1
	p1.Hand = []model.Card{
		{ID: "h1", Name: "手牌1", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
		{ID: "h2", Name: "手牌2", Type: model.CardTypeAttack, Element: model.ElementWater, Damage: 1},
		{ID: "h3", Name: "手牌3", Type: model.CardTypeMagic, Element: model.ElementWind, Damage: 0},
		{ID: "h4", Name: "手牌4", Type: model.CardTypeMagic, Element: model.ElementEarth, Damage: 0},
		{ID: "h5", Name: "手牌5", Type: model.CardTypeAttack, Element: model.ElementThunder, Damage: 1},
		{ID: "h6", Name: "手牌6", Type: model.CardTypeMagic, Element: model.ElementLight, Damage: 0},
	}

	startupCtx := game.buildContext(p1, nil, model.TriggerOnTurnStart, &model.EventContext{
		Type:     model.EventTurnStart,
		SourceID: p1.ID,
	})
	game.State.PendingInterrupt = &model.Interrupt{
		Type:     model.InterruptStartupSkill,
		PlayerID: p1.ID,
		SkillIDs: []string{"elf_ritual"},
		Context:  startupCtx,
	}
	game.State.Phase = model.PhaseStartup

	if err := game.ConfirmStartupSkill(p1.ID, "elf_ritual"); err != nil {
		t.Fatalf("confirm startup ritual failed: %v", err)
	}

	if got := len(p1.Hand); got != 6 {
		t.Fatalf("ritual startup confirm should keep normal hand size 6, got=%d", got)
	}
	if got := len(p1.Blessings); got != 3 {
		t.Fatalf("ritual startup confirm should create 3 blessings, got=%d", got)
	}
	if game.State.PendingInterrupt != nil && game.State.PendingInterrupt.Type == model.InterruptDiscard {
		t.Fatalf("should not leave pending discard interrupt after ritual")
	}
	for _, intr := range game.State.InterruptQueue {
		if intr != nil && intr.Type == model.InterruptDiscard && intr.PlayerID == p1.ID {
			t.Fatalf("should not keep queued discard interrupt for ritual player")
		}
	}
}
