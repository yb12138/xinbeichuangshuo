package blade_master_test

import (
	"fmt"
	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
	"starcup-engine/internal/testutils"
	"testing"
)

func TestDebugHolySwordInterruptFiring(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "BladeMaster", "blade_master", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.TurnState.UsedSkillCounts["wind_fury"] = 1
	p1.TurnState.UsedSkillCounts["sword_shadow"] = 1
	p1.Gem = 0
	p1.Crystal = 0

	p1.Hand = []model.Card{
		{ID: "a1", Name: "火斩1", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
		{ID: "a2", Name: "火斩2", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
		{ID: "a3", Name: "火斩3", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}
	model.AppendAttackAction(p1, "test-extra-1")
	model.AppendAttackAction(p1, "test-extra-2")
	p2.Heal = 0

	// First 2 attacks
	for i := 0; i < 2; i++ {
		if err := game.HandleAction(model.PlayerAction{
			PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0,
		}); err != nil {
			t.Fatalf("attack #%d failed: %v", i+1, err)
		}
		if err := game.HandleAction(model.PlayerAction{
			PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"take"},
		}); err != nil {
			t.Fatalf("respond take #%d failed: %v", i+1, err)
		}
	}
	fmt.Printf("After attack #2: TurnStage=%s, AttackCount=%d, PendingActions=%v, CurrentExtraAction=%q\n",
		game.State.TurnStage, p1.TurnState.AttackCount, p1.TurnState.PendingActions, p1.TurnState.CurrentExtraAction)

	// 3rd attack
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0,
	}); err != nil {
		t.Fatalf("third attack failed: %v", err)
	}

	fmt.Printf("After attack #3: TurnStage=%s, AttackCount=%d, PendingActions=%v, CurrentExtraAction=%q\n",
		game.State.TurnStage, p1.TurnState.AttackCount, p1.TurnState.PendingActions, p1.TurnState.CurrentExtraAction)
	if game.State.PendingInterrupt != nil {
		fmt.Printf("After attack #3: PendingInterrupt.Type=%s, PlayerID=%s\n", game.State.PendingInterrupt.Type, game.State.PendingInterrupt.PlayerID)
	} else {
		fmt.Println("After attack #3: NO PendingInterrupt")
	}
}
