package engine_test

import (
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
	"starcup-engine/internal/testutils"
)

func newSageCancelGame(t *testing.T) (*engine.GameEngine, *model.Player) {
	t.Helper()
	g := engine.NewGameEngine(testutils.NoopObserver{})
	if err := g.AddPlayer("p1", "Sage", "sage", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	g.State.Deck = rules.InitDeck()
	g.State.CurrentTurn = 0
	g.State.TurnStage = model.TurnStageActionExecution
	p1 := g.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 1
	p1.Hand = []model.Card{
		{ID: "a", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
		{ID: "b", Name: "水涟", Type: model.CardTypeAttack, Element: model.ElementWater, Damage: 2},
	}
	return g, p1
}

// 通用-2：首步取消（无上一步可退）应放弃流程并退费，回到行动选择。
func TestSkillCancel_FirstStepRefundsCostAndRestoresActionWindow(t *testing.T) {
	g, p1 := newSageCancelGame(t)
	testutils.MustHandleAction(t, g, model.PlayerAction{PlayerID: "p1", Type: model.CmdSkill, SkillID: "sage_arcane_codex"})
	if p1.Gem != 0 || g.State.PendingInterrupt == nil {
		t.Fatalf("expected gem paid and card prompt, gem=%d intr=%+v", p1.Gem, g.State.PendingInterrupt)
	}
	testutils.MustHandleAction(t, g, model.PlayerAction{PlayerID: "p1", Type: model.CmdCancel})
	if g.State.PendingInterrupt != nil {
		t.Fatalf("expected flow aborted, got %+v", g.State.PendingInterrupt)
	}
	if p1.Gem != 1 {
		t.Fatalf("expected gem refunded, got %d", p1.Gem)
	}
	if p1.TurnState.HasActed || p1.TurnState.LastActionType != "" || p1.TurnState.UsedSkillCounts["sage_arcane_codex"] != 0 {
		t.Fatalf("expected action marks rolled back, got %+v", p1.TurnState)
	}
	if !g.IsActionSelectionWindow() {
		t.Fatalf("expected action selection window, stage=%s", g.State.TurnStage)
	}
}

// 通用-2：后续步骤 Abort 取消也退费。
func TestSkillCancel_LaterAbortStepRefundsCost(t *testing.T) {
	g, p1 := newSageCancelGame(t)
	testutils.MustHandleAction(t, g, model.PlayerAction{PlayerID: "p1", Type: model.CmdSkill, SkillID: "sage_arcane_codex"})
	testutils.MustHandleAction(t, g, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0, 1}})
	testutils.RequireChoiceContext(t, g, "p1", "sage_arcane_target")
	testutils.MustHandleAction(t, g, model.PlayerAction{PlayerID: "p1", Type: model.CmdCancel})
	if g.State.PendingInterrupt != nil || p1.Gem != 1 || len(p1.Hand) != 2 {
		t.Fatalf("expected abort + refund with hand intact, intr=%v gem=%d hand=%d", g.State.PendingInterrupt, p1.Gem, len(p1.Hand))
	}
}

// 通用-2：正常结算后费用不退。
func TestSkillCancel_ResolvedSkillKeepsCost(t *testing.T) {
	g, p1 := newSageCancelGame(t)
	testutils.MustHandleAction(t, g, model.PlayerAction{PlayerID: "p1", Type: model.CmdSkill, SkillID: "sage_arcane_codex"})
	testutils.MustHandleAction(t, g, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0, 1}})
	testutils.MustHandleAction(t, g, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{1}})
	if p1.Gem != 0 {
		t.Fatalf("expected gem kept spent after resolution, got %d", p1.Gem)
	}
}

// 通用-3：回合结束钩子在额外行动之后才执行，且整回合只执行一次。
func TestTurnEndHooks_RunAfterExtraActions(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
	if err := g.AddPlayer("p1", "BS", "beast_samurai", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	g.State.Deck = rules.InitDeck()
	g.State.CurrentTurn = 0
	p1 := g.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = nil
	p1.Form = string(model.FormBeastSamuraiIaijutsu)
	p1.Tokens["bs_beast_soul"] = 2
	model.AppendAttackAction(p1, "测试额外攻击")
	g.State.TurnStage = model.TurnStageTurnEnd

	g.Drive()
	if p1.Tokens["bs_beast_soul"] != 2 {
		t.Fatalf("turn-end hook must not run before extra action, beast soul=%d", p1.Tokens["bs_beast_soul"])
	}
	if !g.IsActionSelectionWindow() {
		t.Fatalf("expected extra action selection, stage=%s", g.State.TurnStage)
	}
}

// 通用-4：回合限定按任意一个回合各限一次：任何玩家回合开始时都重置。
func TestTurnLimit_ResetsAtEveryPlayersTurnStart(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
	if err := g.AddPlayer("p1", "MG", "moon_goddess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	g.State.Deck = rules.InitDeck()
	g.State.CurrentTurn = 0
	p1 := g.State.Players["p1"]
	p2 := g.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState.UsedSkillCounts["mg_blasphemy"] = 1
	p1.TurnState.UsedSkillCounts["non_turn_limit_marker"] = 1

	g.NextTurn()
	if g.State.PlayerOrder[g.State.CurrentTurn] != p2.ID {
		t.Fatalf("expected p2 turn")
	}
	if p1.TurnState.UsedSkillCounts["mg_blasphemy"] != 0 {
		t.Fatalf("turn-limited count should reset at opponent's turn start")
	}
	if p1.TurnState.UsedSkillCounts["non_turn_limit_marker"] != 1 {
		t.Fatalf("non turn-limit markers must be preserved")
	}
}
