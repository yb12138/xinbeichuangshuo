package elf_archer_test

import (
	"starcup-engine/internal/engine"
	playerpkg "starcup-engine/internal/engine/player"
	elfarcher "starcup-engine/internal/engine/player/elf_archer"
	"starcup-engine/internal/testutils"
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func activateElfElementalShotByPickingCard(t *testing.T, game *engine.GameEngine, playerID string) {
	t.Helper()
	testutils.ChooseResponseSkillByID(t, game, playerID, "elf_elemental_shot")
	testutils.RequireChoicePrompt(t, game, playerID, "elf_archer_elemental_shot_pick")
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   playerID,
		Type:       model.CmdSelect,
		Selections: []int{0},
	})
}

func TestElfElementalShotPick_CancelSupported(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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
		{ID: "atk-fire", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
		{ID: "magic-cost", Name: "烈焰术", Type: model.CardTypeMagic, Element: model.ElementFire},
	}
	p2.Hand = []model.Card{
		{ID: "def-light", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight},
	}

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdAttack,
		TargetID: "p2",
		CardID:   testutils.PlayableCardID(t, game, "p1", 0),
	})
	testutils.ChooseResponseSkillByID(t, game, "p1", "elf_elemental_shot")
	testutils.RequireChoicePrompt(t, game, "p1", "elf_archer_elemental_shot_pick")

	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdCancel})

	if game.State.PendingInterrupt != nil {
		if game.State.PendingInterrupt.Type == model.InterruptChoice {
			ctxData, _ := game.State.PendingInterrupt.Context.(map[string]interface{})
			if got, _ := ctxData["choice_type"].(string); got == "elf_archer_elemental_shot_pick" {
				t.Fatalf("expected elemental-shot pick prompt closed after cancel")
			}
		}
	}
	if got := len(p1.Hand); got != 1 {
		t.Fatalf("expected no extra discard after cancel, remaining hand should only have cost card, got %d", got)
	}
	if p1.Hand[0].ID != "magic-cost" {
		t.Fatalf("expected remaining card to be magic-cost after cancel, got %+v", p1.Hand[0])
	}
	if p1.TurnState.SkillFlowState["elf_elemental_shot_wind_pending"] != 0 {
		t.Fatalf("expected no wind pending state on cancel, got %d", p1.TurnState.SkillFlowState["elf_elemental_shot_wind_pending"])
	}
}

func TestElfElementalShotWind_GrantsExtraAttackOnlyAfterActionEnd(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdAttack,
		TargetID: "p2",
		CardID:   testutils.PlayableCardID(t, game, "p1", 0),
	})
	activateElfElementalShotByPickingCard(t, game, "p1")

	if len(p1.TurnState.PendingActions) != 0 {
		t.Fatalf("wind shot should not grant extra action before attack ends, got %+v", p1.TurnState.PendingActions)
	}
	if p1.TurnState.SkillFlowState["elf_elemental_shot_wind_pending"] != 1 {
		t.Fatalf("expected wind pending state before combat resolves, got %d", p1.TurnState.SkillFlowState["elf_elemental_shot_wind_pending"])
	}

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"defend"},
		CardID:    testutils.PlayableCardID(t, game, "p2", 0),
	})

	game.Drive()
	if game.State.TurnStage != model.TurnStageActionExecution {
		t.Fatalf("expected extra action to enter action execution window, got %s", game.State.TurnStage)
	}
	if p1.TurnState.CurrentExtraAction != "Attack" {
		t.Fatalf("expected wind shot to grant extra attack after action end, got %q", p1.TurnState.CurrentExtraAction)
	}
	if p1.TurnState.SkillFlowState["elf_elemental_shot_wind_pending"] != 0 {
		t.Fatalf("expected wind pending state cleared after action end, got %d", p1.TurnState.SkillFlowState["elf_elemental_shot_wind_pending"])
	}
}

func TestElfElementalShotWater_AutoResolvesOnCurrentTarget(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdAttack,
		TargetID: "p2",
		CardID:   testutils.PlayableCardID(t, game, "p1", 0),
	})
	activateElfElementalShotByPickingCard(t, game, "p1")

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	})

	testutils.RequireResponseSkillPrompt(t, game, "p1")
	if p2.Heal != 1 {
		t.Fatalf("expected water shot to heal current combat target by 1, got %d", p2.Heal)
	}
	if p3.Heal != 0 {
		t.Fatalf("expected non-target player to stay unchanged, got heal=%d", p3.Heal)
	}
}

func TestElfElementalShotPickAllowsBlessingMagic(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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
		{ID: "hand-attack", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}
	playerpkg.SetForm(p1, model.FormElfArcherRitual)
	markElfBlessings(p1, []model.Card{
		{ID: "bless-magic", Name: "圣盾", Type: model.CardTypeMagic, Element: model.ElementLight, Damage: 0},
	})

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdAttack,
		TargetID: "p2",
		CardID:   testutils.PlayableCardID(t, game, "p1", 0),
	})
	testutils.ChooseResponseSkillByID(t, game, "p1", "elf_elemental_shot")
	testutils.RequireChoicePrompt(t, game, "p1", "elf_archer_elemental_shot_pick")
	ctxData, ok := game.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		t.Fatalf("expected elemental shot context map")
	}
	if got, _ := ctxData["choice_type"].(string); got != "elf_archer_elemental_shot_pick" {
		t.Fatalf("expected direct elemental shot pick prompt, got %q", got)
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})

	if got := elfarcher.CountBlessings(p1); got != 0 {
		t.Fatalf("expected blessing to be consumed after elemental shot, got %d", got)
	}
	if got := len(p1.Hand); got != 0 {
		t.Fatalf("expected attack card to be consumed and no hand card left, got %d", got)
	}
}

func TestElfRitualRelease_TargetsEnemyOnly(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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
	playerpkg.SetForm(p1, model.FormElfArcherRitual)

	game.Drive()

	testutils.RequireChoicePrompt(t, game, "p1", "elf_ritual_release_target")
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
