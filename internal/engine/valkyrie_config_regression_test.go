package engine

import (
	"testing"

	playerpkg "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func TestValkyrie_HeroicSummon_ExtraDiscardHealsCurrentCombatTarget(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Valkyrie", "valkyrie", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Target", "berserker", model.BlueCamp); err != nil {
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
	p1.Crystal = 1
	p1.Hand = []model.Card{
		{ID: "atk1", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
		{ID: "magic1", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight, Damage: 0},
	}

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	})

	requireResponseSkillPrompt(t, game, "p1")
	chooseResponseSkillByID(t, game, "p1", "valkyrie_heroic_summon")
	requireChoicePrompt(t, game, "p1", "valkyrie_heroic_discard_card")
	mustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0}})

	if p1.Heal != 0 {
		t.Fatalf("expected heroic summon extra heal not to target self, got self heal=%d", p1.Heal)
	}
	if p2.Heal != 1 {
		t.Fatalf("expected heroic summon extra heal to target current combat target, got target heal=%d", p2.Heal)
	}
}

func TestValkyrie_HeroicSummon_ExtraDiscardCanCancel(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Valkyrie", "valkyrie", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Target", "berserker", model.BlueCamp); err != nil {
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
	p1.Crystal = 1
	p1.Hand = []model.Card{
		{ID: "atk1", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
		{ID: "magic1", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight, Damage: 0},
	}

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	})

	requireResponseSkillPrompt(t, game, "p1")
	chooseResponseSkillByID(t, game, "p1", "valkyrie_heroic_summon")
	requireChoicePrompt(t, game, "p1", "valkyrie_heroic_discard_card")

	mustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdCancel})

	if got := len(p1.Hand); got != 1 {
		t.Fatalf("expected cancel keep magic card in hand, got hand=%d", got)
	}
	if p2.Heal != 0 {
		t.Fatalf("expected cancel not to apply extra heal, got target heal=%d", p2.Heal)
	}
	if game.State.PendingInterrupt != nil && game.State.PendingInterrupt.Type == model.InterruptChoice {
		ctxData, _ := game.State.PendingInterrupt.Context.(map[string]interface{})
		if got, _ := ctxData["choice_type"].(string); got == "valkyrie_heroic_discard_card" {
			t.Fatalf("expected heroic discard choice cleared after cancel")
		}
	}
}

func TestValkyrie_HeroicSummon_DoesNotEnterSpiritOnCounterHit(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Attacker", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "CounterValkyrie", "valkyrie", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "CounterTarget", "angel", model.RedCamp); err != nil {
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
	p2.Crystal = 1
	p1.Hand = []model.Card{
		{ID: "atk-p1-fire", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}
	p2.Hand = []model.Card{
		{ID: "atk-p2-fire", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"counter"},
		CardIndex: 0,
		TargetID:  "p3",
	})
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p3",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	})

	requireResponseSkillPrompt(t, game, "p2")
	chooseResponseSkillByID(t, game, "p2", "valkyrie_heroic_summon")

	if playerpkg.HasForm(p2, model.FormValkyrieHeroic) {
		t.Fatalf("expected counter-hit heroic summon not to enter spirit form, got form=%q", p2.Form)
	}
}

func TestValkyrie_MilitaryGlory_BranchTwoDoesNotExitSpirit(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Valkyrie", "valkyrie", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageTurnStart

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	playerpkg.SetForm(p1, model.FormValkyrieHeroic)
	game.State.RedCrystals = 2

	game.Drive()

	requireChoicePrompt(t, game, "p1", "valkyrie_military_glory_mode")
	mustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{1}})
	requireChoicePrompt(t, game, "p1", "valkyrie_military_glory_x")
	mustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{1}})
	requireChoicePrompt(t, game, "p1", "valkyrie_military_glory_target")

	targetSelection := -1
	ctxData, _ := game.State.PendingInterrupt.Context.(map[string]interface{})
	if targetIDs, ok := ctxData["target_ids"].([]string); ok {
		for i, tid := range targetIDs {
			if tid == "p2" {
				targetSelection = i
				break
			}
		}
	}
	if targetSelection < 0 {
		t.Fatalf("ally target p2 not found in military glory target list")
	}
	mustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{targetSelection}})

	if !playerpkg.HasForm(p1, model.FormValkyrieHeroic) {
		t.Fatalf("expected branch two to keep spirit form, got form=%q", p1.Form)
	}
	if got := p2.Heal; got != 2 {
		t.Fatalf("expected branch two to heal target by X=2, got %d", got)
	}
}

func TestValkyrie_PeaceWalker_ReleasesSpiritOnActiveAttack(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Valkyrie", "valkyrie", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Target", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	playerpkg.SetForm(p1, model.FormValkyrieHeroic)
	p1.Hand = []model.Card{
		{ID: "atk1", Name: "风斩", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 1},
	}

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})

	if playerpkg.HasForm(p1, model.FormValkyrieHeroic) {
		t.Fatalf("expected peace walker to release spirit form on active attack, got form=%q", p1.Form)
	}
}
