package elf_archer_test

import (
	"fmt"
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/testutils"
)

func setupElfDamageGame(t *testing.T, elfActive bool) *engine.GameEngine {
	t.Helper()
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Elf", "elf_archer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	deck := make([]model.Card, 0, 30)
	for i := 0; i < 30; i++ {
		deck = append(deck, model.Card{ID: fmt.Sprintf("d%d", i), Name: "补牌", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2})
	}
	game.State.Deck = deck
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = elfActive
	p2.IsActive = !elfActive
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p1.Heal = 0
	p2.Heal = 0
	p1.Hand = nil
	p2.Hand = nil
	return game
}

func fillHand(p *model.Player, n int, prefix string) {
	p.Hand = make([]model.Card, n)
	for i := range p.Hand {
		p.Hand[i] = model.Card{ID: fmt.Sprintf("%s%d", prefix, i), Name: "手牌", Type: model.CardTypeAttack, Element: model.ElementWater, Damage: 2}
	}
}

func dealElfMagicDamage(t *testing.T, game *engine.GameEngine) {
	t.Helper()
	game.AddPendingDamage(model.PendingDamage{SourceID: "p1", TargetID: "p2", Damage: 1, DamageType: model.MagicAttack})
	game.Drive()
}

func discardCtx(t *testing.T, game *engine.GameEngine, playerID string) map[string]interface{} {
	t.Helper()
	intr := game.State.PendingInterrupt
	if intr == nil || !engine.IsDiscardSelectionInterrupt(intr) || intr.PlayerID != playerID {
		t.Fatalf("expected discard interrupt for %s, got %+v", playerID, intr)
	}
	data, _ := intr.Context.(map[string]interface{})
	return data
}

func discardFirstN(t *testing.T, game *engine.GameEngine, playerID string, n int) {
	t.Helper()
	sel := make([]int, n)
	for i := range sel {
		sel[i] = i
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: playerID, Type: model.CmdSelect, Selections: sel})
}

func TestElfAnimalCompanion_TriggersOnMagicDamage(t *testing.T) {
	game := setupElfDamageGame(t, true)
	p1 := game.State.Players["p1"]
	fillHand(p1, 1, "own")
	dealElfMagicDamage(t, game)

	testutils.RequireResponseSkillPrompt(t, game, "p1")
	if skillIndex(game.State.PendingInterrupt.SkillIDs, "elf_animal_companion") < 0 {
		t.Fatalf("expected animal companion on magic damage, got %+v", game.State.PendingInterrupt.SkillIDs)
	}
	testutils.ChooseResponseSkillByID(t, game, "p1", "elf_animal_companion")
	if len(p1.Hand) != 2 {
		t.Fatalf("expected elf drew 1 (hand=2), got %d", len(p1.Hand))
	}
	data := discardCtx(t, game, "p1")
	if c, _ := data["discard_count"].(int); c != 1 {
		t.Fatalf("expected skill discard 1, got %+v", data)
	}
	discardFirstN(t, game, "p1", 1)
	if len(p1.Hand) != 1 {
		t.Fatalf("expected hand=1 after skill discard, got %d", len(p1.Hand))
	}
}

func TestElfAnimalCompanion_NotTriggeredOutsideOwnTurn(t *testing.T) {
	game := setupElfDamageGame(t, false)
	dealElfMagicDamage(t, game)
	if intr := game.State.PendingInterrupt; intr != nil && intr.Type == model.InterruptResponseSkill &&
		(skillIndex(intr.SkillIDs, "elf_animal_companion") >= 0 || skillIndex(intr.SkillIDs, "elf_pet_empower") >= 0) {
		t.Fatalf("animal companion/pet empower must not trigger outside elf's turn, got %+v", intr.SkillIDs)
	}
}

func TestElfAnimalCompanion_OverflowDiscardsToLimitThenOneMore(t *testing.T) {
	game := setupElfDamageGame(t, true)
	p1 := game.State.Players["p1"]
	maxHand := game.GetMaxHand(p1)
	fillHand(p1, maxHand, "own")
	moraleBefore := game.State.RedMorale
	dealElfMagicDamage(t, game)

	testutils.ChooseResponseSkillByID(t, game, "p1", "elf_animal_companion")
	if len(p1.Hand) != maxHand+1 {
		t.Fatalf("expected hand=%d after draw, got %d", maxHand+1, len(p1.Hand))
	}
	// 第一步：爆牌弃牌，弃到上限。
	data := discardCtx(t, game, "p1")
	if reason, _ := data["discard_reason"].(string); reason != "hand_overflow" {
		t.Fatalf("expected overflow discard first, got %+v", data)
	}
	if len(game.State.InterruptQueue) != 0 {
		t.Fatalf("skill discard must not be queued before overflow resolves, got %d queued", len(game.State.InterruptQueue))
	}
	discardFirstN(t, game, "p1", 1)
	if len(p1.Hand) != maxHand {
		t.Fatalf("expected hand at limit %d after overflow discard, got %d", maxHand, len(p1.Hand))
	}
	if game.State.RedMorale != moraleBefore-1 {
		t.Fatalf("expected overflow discard to cost 1 morale (%d -> %d), got %d", moraleBefore, moraleBefore-1, game.State.RedMorale)
	}
	// 第二步：技能本身的弃1张。
	data = discardCtx(t, game, "p1")
	if reason, _ := data["discard_reason"].(string); reason == "hand_overflow" {
		t.Fatalf("expected skill discard after overflow, got another overflow discard %+v", data)
	}
	discardFirstN(t, game, "p1", 1)
	if len(p1.Hand) != maxHand-1 {
		t.Fatalf("expected hand=%d after extra skill discard, got %d", maxHand-1, len(p1.Hand))
	}
	if game.State.RedMorale != moraleBefore-1 {
		t.Fatalf("skill discard must not cost morale, got %d", game.State.RedMorale)
	}
	game.Drive()
	if intr := game.State.PendingInterrupt; intr != nil && engine.IsDiscardSelectionInterrupt(intr) {
		t.Fatalf("expected no further discard, got %+v", intr)
	}
}

func TestElfPetEmpower_TriggersOnMagicDamage(t *testing.T) {
	game := setupElfDamageGame(t, true)
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.Crystal = 1
	dealElfMagicDamage(t, game)

	testutils.RequireResponseSkillPrompt(t, game, "p1")
	if skillIndex(game.State.PendingInterrupt.SkillIDs, "elf_pet_empower") < 0 {
		t.Fatalf("expected pet empower on magic damage, got %+v", game.State.PendingInterrupt.SkillIDs)
	}
	handBefore := len(p2.Hand)
	testutils.ChooseResponseSkillByID(t, game, "p1", "elf_pet_empower")
	if p1.Crystal != 0 {
		t.Fatalf("expected pet empower to cost 1 crystal, got %d", p1.Crystal)
	}
	if len(p2.Hand) != handBefore+1 {
		t.Fatalf("expected target drew 1, got %d -> %d", handBefore, len(p2.Hand))
	}
	discardCtx(t, game, "p2")
	discardFirstN(t, game, "p2", 1)
	if len(p2.Hand) != handBefore {
		t.Fatalf("expected target discarded 1, got %d", len(p2.Hand))
	}
}
