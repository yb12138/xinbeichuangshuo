package bard_test

import (
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/testutils"
)

// 不谐和弦：选完 X 后在选分支时取消，退回选 X 并返还灵感；再次确认只扣一次。
func TestBardDissonanceChord_CancelAtModeRefundsInspiration(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Bard", "bard", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	bard := game.State.Players["p1"]
	bard.IsActive = true
	bard.TurnState = model.NewPlayerTurnState()
	bard.Tokens["bd_inspiration"] = 3
	bard.Form = model.FormBardEternalPrisoner
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	if err := game.UseSkill("p1", "bd_dissonance_chord", nil, nil); err != nil {
		t.Fatalf("use dissonance failed: %v", err)
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}) // X=2
	if got := bard.Tokens["bd_inspiration"]; got != 1 {
		t.Fatalf("expected inspiration 1 after X=2, got %d", got)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "bd_dissonance_mode")

	testutils.MustHandleAction(t, game, model.PlayerAction{Type: model.CmdCancel, PlayerID: "p1"})
	testutils.RequireChoicePrompt(t, game, "p1", "bd_dissonance_x")
	if got := bard.Tokens["bd_inspiration"]; got != 3 {
		t.Fatalf("expected inspiration refunded to 3 after back, got %d", got)
	}
	if bard.Form != model.FormBardEternalPrisoner {
		t.Fatalf("expected prisoner form restored after back, got %q", bard.Form)
	}

	testutils.MustHandleAction(t, game, model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}) // X=2 again
	if got := bard.Tokens["bd_inspiration"]; got != 1 {
		t.Fatalf("expected inspiration 1 after re-confirm (no double charge), got %d", got)
	}
}

// 沉沦协奏曲：选完第 1 张后，候选只剩与它同系的牌。
func TestBardDescentConcerto_SecondCardFilteredByFirstElement(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Bard", "bard", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "EnemyA", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "EnemyB", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	bard := game.State.Players["p1"]
	bard.IsActive = true
	bard.TurnState = model.NewPlayerTurnState()
	bard.Hand = []model.Card{
		bardTestCard("f1", "火1", model.CardTypeAttack, model.ElementFire),
		bardTestCard("w1", "水1", model.CardTypeAttack, model.ElementWater),
		bardTestCard("f2", "火2", model.CardTypeAttack, model.ElementFire),
		bardTestCard("w2", "水2", model.CardTypeAttack, model.ElementWater),
	}
	game.HandlePostDamageResolved(&model.PendingDamage{SourceID: "p1", TargetID: "p2", Damage: 1, DamageType: model.MagicAttack})
	game.HandlePostDamageResolved(&model.PendingDamage{SourceID: "p1", TargetID: "p3", Damage: 1, DamageType: model.MagicAttack})
	if !game.RunTurnEndTimingStageHooks(bard, engine.TimingTurnEndPreExtra) {
		t.Fatalf("descent should trigger")
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}})
	// 选第 1 张：火1（手牌下标 0）
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("pick first card failed: %v", err)
	}
	prompt := game.BuildChoicePrompt()
	if prompt == nil {
		t.Fatalf("expected second-card prompt")
	}
	if len(prompt.Options) != 1 || prompt.Options[0].CardID != "f2" {
		t.Fatalf("expected only same-element card f2 as candidate, got %+v", prompt.Options)
	}
}

// 沉沦协奏曲：回合限定按每个回合各限一次；别的玩家回合开始前要重置标记。
func TestBardDescentConcerto_TurnLimitResetsOnEveryTurn(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Bard", "bard", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	bard := game.State.Players["p1"]
	enemy := game.State.Players["p2"]
	bard.IsActive = true
	bard.TurnState = model.NewPlayerTurnState()
	bard.TurnState.UsedSkillCounts["bd_descent_prompted"] = 1
	bard.TurnState.UsedSkillCounts["bd_descent"] = 1
	game.State.CurrentTurn = 0

	// 真实地推进到敌方（非吟游诗人）的回合：重置发生在回合切换时。
	game.NextTurn()
	if !enemy.IsActive {
		t.Fatalf("expected enemy turn after NextTurn")
	}

	if bard.TurnState.UsedSkillCounts["bd_descent_prompted"] != 0 || bard.TurnState.UsedSkillCounts["bd_descent"] != 0 {
		t.Fatalf("descent markers must reset at another player's turn start, got %+v", bard.TurnState.UsedSkillCounts)
	}
}
