package engine

import (
	"testing"

	"starcup-engine/internal/model"
)

func requireSaintHealStage(t *testing.T, game *GameEngine, playerID, stage string) map[string]interface{} {
	t.Helper()
	if game.State.PendingInterrupt == nil {
		t.Fatalf("expected saint heal interrupt, got nil")
	}
	if game.State.PendingInterrupt.Type != model.InterruptSaintHeal {
		t.Fatalf("expected saint heal interrupt, got %s", game.State.PendingInterrupt.Type)
	}
	if game.State.PendingInterrupt.PlayerID != playerID {
		t.Fatalf("expected saint heal interrupt for %s, got %s", playerID, game.State.PendingInterrupt.PlayerID)
	}
	ctx, ok := game.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		t.Fatalf("saint heal context type mismatch")
	}
	got, _ := ctx["stage"].(string)
	if got != stage {
		t.Fatalf("expected saint heal stage %q, got %q (ctx=%+v)", stage, got, ctx)
	}
	return ctx
}

func TestSaintess_SaintHeal_TwoTargetSplitCanChooseMagicExtraAction(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Saintess", "saintess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdSkill,
		SkillID:   "saint_heal",
		TargetIDs: []string{"p1", "p2"},
	})
	requireSaintHealStage(t, game, "p1", "allocate_heal")

	// 选择第二名目标获得2点治疗，自己获得1点治疗。
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{1},
	})
	requireSaintHealStage(t, game, "p1", "choose_extra_action")

	// 选择额外法术行动。
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{1},
	})

	if p1.Heal != 1 || p2.Heal != 2 {
		t.Fatalf("expected saint heal split p1=1 p2=2, got p1=%d p2=%d", p1.Heal, p2.Heal)
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected saint heal flow to resolve cleanly, got %+v", game.State.PendingInterrupt)
	}
	if !game.isActionSelectionWindow() {
		t.Fatalf("expected extra magic action to enter action selection window, got %s", game.runtimeStateLabel())
	}
	if p1.TurnState.CurrentExtraAction != "Magic" {
		t.Fatalf("expected saint heal to grant extra magic action, got %q", p1.TurnState.CurrentExtraAction)
	}
}

func TestSaintess_Mercy_BecomesPersistentFixedHandCapState(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Saintess", "saintess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 1

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionStart
	game.Drive()

	mercyIdx := startupSkillIndexByID(t, game, "p1", "mercy")
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{mercyIdx},
	})

	if p1.Gem != 0 {
		t.Fatalf("expected mercy to consume 1 gem, got %d", p1.Gem)
	}
	if p1.Crystal != 1 {
		t.Fatalf("expected mercy to grant 1 crystal to self, got %d", p1.Crystal)
	}
	if !p1.HasFieldEffect(model.EffectMercy) {
		t.Fatalf("expected mercy field effect to persist on self")
	}
	if got := game.GetMaxHand(p1); got != 7 {
		t.Fatalf("expected mercy fixed max hand 7, got %d", got)
	}

	// 下一次自己的启动阶段，不应再次弹出怜悯。
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	game.State.PendingInterrupt = nil
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionStart
	game.Drive()

	if game.State.PendingInterrupt != nil && game.State.PendingInterrupt.Type == model.InterruptStartupSkill {
		t.Fatalf("expected mercy not to reappear once active, got %+v", game.State.PendingInterrupt)
	}
}
