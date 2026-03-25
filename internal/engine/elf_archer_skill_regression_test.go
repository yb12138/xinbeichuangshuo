package engine

import (
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func activateElfElementalShotByDiscardMagic(t *testing.T, game *GameEngine, playerID string) {
	t.Helper()
	chooseResponseSkillByID(t, game, playerID, "elf_elemental_shot")
	requireChoicePrompt(t, game, playerID, "elf_elemental_shot_cost")
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   playerID,
		Type:       model.CmdSelect,
		Selections: []int{0},
	})
	requireChoicePrompt(t, game, playerID, "elf_elemental_shot_discard_magic")
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   playerID,
		Type:       model.CmdSelect,
		Selections: []int{0},
	})
}

func TestElfElementalShotWind_GrantsExtraAttackOnlyAfterActionEnd(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Elf", "elf_archer", model.RedCamp); err != nil {
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
	p1.Hand = []model.Card{
		{ID: "atk-wind", Name: "风斩", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 1},
		{ID: "magic-cost", Name: "风语", Type: model.CardTypeMagic, Element: model.ElementWind},
	}
	p2.Hand = []model.Card{
		{ID: "def-magic", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight},
	}

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})
	activateElfElementalShotByDiscardMagic(t, game, "p1")

	if len(p1.TurnState.PendingActions) != 0 {
		t.Fatalf("wind shot should not grant extra action before attack ends, got %+v", p1.TurnState.PendingActions)
	}
	if p1.Tokens["elf_elemental_shot_wind_pending"] != 1 {
		t.Fatalf("expected wind pending token before combat resolves, got %d", p1.Tokens["elf_elemental_shot_wind_pending"])
	}

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"defend"},
		CardIndex: 0,
	})

	game.Drive()
	if game.State.TurnStage != model.TurnStageActionExecution {
		t.Fatalf("expected extra action to enter action execution window, got %s", game.State.TurnStage)
	}
	if p1.TurnState.CurrentExtraAction != "Attack" {
		t.Fatalf("expected wind shot to grant extra attack after action end, got %q", p1.TurnState.CurrentExtraAction)
	}
	if p1.Tokens["elf_elemental_shot_wind_pending"] != 0 {
		t.Fatalf("expected wind pending token cleared after action end, got %d", p1.Tokens["elf_elemental_shot_wind_pending"])
	}
}

func TestElfElementalShotWater_AutoResolvesOnCurrentTarget(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Elf", "elf_archer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Target", "plague_mage", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Other", "angel", model.RedCamp); err != nil {
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
	p1.Hand = []model.Card{
		{ID: "atk-water", Name: "水斩", Type: model.CardTypeAttack, Element: model.ElementWater, Damage: 1},
		{ID: "magic-cost", Name: "潮汐术", Type: model.CardTypeMagic, Element: model.ElementWater},
	}

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})
	activateElfElementalShotByDiscardMagic(t, game, "p1")

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	})

	requireResponseSkillPrompt(t, game, "p1")
	if p2.Heal != 1 {
		t.Fatalf("expected water shot to heal current combat target by 1, got %d", p2.Heal)
	}
	if p3.Heal != 0 {
		t.Fatalf("expected non-target player to stay unchanged, got heal=%d", p3.Heal)
	}
}

func TestElfRitualRelease_TargetsEnemyOnly(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Elf", "elf_archer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageTurnEnd

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p3 := game.State.Players["p3"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p3.TurnState = model.NewPlayerTurnState()
	p1.Tokens = map[string]int{
		"elf_ritual_release_waiting": 0,
	}
	enterElfArcherRitualForm(p1)
	p1.Blessings = nil
	syncElfBlessings(p1)

	game.Drive()

	requireChoicePrompt(t, game, "p1", "elf_ritual_release_target")
	ctxData, ok := game.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		t.Fatalf("expected ritual release context map")
	}
	targetIDs, ok := ctxData["target_ids"].([]string)
	if !ok {
		t.Fatalf("expected target_ids []string, got %+v", ctxData["target_ids"])
	}
	if len(targetIDs) != 1 || targetIDs[0] != "p3" {
		t.Fatalf("expected only enemy target p3, got %+v", targetIDs)
	}
}
