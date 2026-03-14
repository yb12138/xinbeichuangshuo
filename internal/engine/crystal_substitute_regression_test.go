package engine

import (
	"strings"
	"testing"

	"starcup-engine/internal/model"
)

// 回归：需要蓝水晶的响应技能，在无蓝水晶但有红宝石时也应可发动（红宝石替代蓝水晶）。
func TestCrystalSubstitute_SwordShadow_ResponseSkill(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "p1", "blade_master", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "p2", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 0
	p1.Gem = 1
	p1.Hand = []model.Card{
		{ID: "atk-1", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}
	p2.Heal = 0

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0,
	}); err != nil {
		t.Fatalf("attack failed: %v", err)
	}
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"take"},
	}); err != nil {
		t.Fatalf("take failed: %v", err)
	}

	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response-skill interrupt, got %+v", game.State.PendingInterrupt)
	}
	hasSwordShadow := false
	for _, sid := range game.State.PendingInterrupt.SkillIDs {
		if sid == "sword_shadow" {
			hasSwordShadow = true
			break
		}
	}
	if !hasSwordShadow {
		t.Fatalf("expected sword_shadow in skill list, got %+v", game.State.PendingInterrupt.SkillIDs)
	}

	if err := game.ConfirmResponseSkill("p1", "sword_shadow"); err != nil {
		t.Fatalf("confirm sword_shadow failed: %v", err)
	}
	if p1.Gem != 0 || p1.Crystal != 0 {
		t.Fatalf("expected crystal-like cost consume gem fallback, got gem=%d crystal=%d", p1.Gem, p1.Crystal)
	}
}

// 回归：红宝石可替代蓝水晶发动主动技能。
func TestCrystalSubstitute_ArbiterBalance_ActionSkill(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "p1", "arbiter", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "p2", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 0
	p1.Gem = 1

	if err := game.UseSkill("p1", "arbiter_balance", nil, nil); err != nil {
		t.Fatalf("use arbiter_balance with gem-as-crystal failed: %v", err)
	}
	if p1.Gem != 0 || p1.Crystal != 0 {
		t.Fatalf("expected cost consumed from gem fallback, got gem=%d crystal=%d", p1.Gem, p1.Crystal)
	}
}

// 回归：宝石消耗不能用蓝水晶反向替代。
func TestCrystalSubstitute_GemCostCannotUseCrystal(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "p1", "holy_lancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "p2", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1
	p1.Gem = 0

	err := game.UseSkill("p1", "holy_lancer_prayer", nil, nil)
	if err == nil {
		t.Fatalf("expected gem-only skill to fail when only crystal is available")
	}
	if !strings.Contains(err.Error(), "资源不足") {
		t.Fatalf("expected resource error, got: %v", err)
	}
}
