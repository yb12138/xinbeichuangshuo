package elementalist_test

import (
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
	"starcup-engine/internal/testutils"
)

func newFreezeCancelGame(t *testing.T) (*engine.GameEngine, *model.Player) {
	t.Helper()
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Elem", "elementalist", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{elementalistExclusiveCard(p1, "冰冻", model.ElementFire)}
	game.State.Players["p2"].Hand = nil
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	game.State.Deck = rules.InitDeck()
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSkill, SkillID: "elementalist_freeze", Selections: []int{0}})
	testutils.RequireChoiceContext(t, game, "p1", "elementalist_freeze_damage_target")
	return game, p1
}

// 审计 元素师-3：冰冻第1步可取消（收回独有技手牌、回到行动选择），第2步可返回第1步。
func TestElementalistFreeze_FirstStepCancelRestoresCardAndAction(t *testing.T) {
	game, p1 := newFreezeCancelGame(t)
	if p := game.BuildPendingInterruptPrompt(); p == nil || p.Presentation == nil || p.Presentation.CancelPolicy != model.CancelPolicyAbort {
		t.Fatalf("expected abort cancel policy on step 1, got %+v", p)
	}
	discardBefore := len(game.State.DiscardPile)
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdCancel})
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected freeze flow aborted, got %+v", game.State.PendingInterrupt)
	}
	if len(p1.Hand) != 1 || p1.Hand[0].Name != "冰冻" || len(game.State.DiscardPile) != discardBefore-1 {
		t.Fatalf("expected exclusive card returned to hand, hand=%+v discard=%d->%d", p1.Hand, discardBefore, len(game.State.DiscardPile))
	}
	if p1.TurnState.HasActed || p1.TurnState.UsedSkillCounts["elementalist_freeze"] != 0 {
		t.Fatalf("expected action marks rolled back, got %+v", p1.TurnState)
	}
	if !game.IsActionSelectionWindow() {
		t.Fatalf("expected action selection window, stage=%s", game.State.TurnStage)
	}
	// 可以重新发动。
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSkill, SkillID: "elementalist_freeze", Selections: []int{0}})
	testutils.RequireChoiceContext(t, game, "p1", "elementalist_freeze_damage_target")
}

func TestElementalistFreeze_SecondStepGoesBack(t *testing.T) {
	game, p1 := newFreezeCancelGame(t)
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{1}})
	testutils.RequireChoiceContext(t, game, "p1", "elementalist_freeze_heal_target")
	if p := game.BuildPendingInterruptPrompt(); p == nil || p.Presentation == nil || p.Presentation.CancelPolicy != model.CancelPolicyBack {
		t.Fatalf("expected back cancel policy on step 2, got %+v", p)
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdCancel})
	testutils.RequireChoiceContext(t, game, "p1", "elementalist_freeze_damage_target")
	// 重新选：伤害给 p2，治疗给自己。
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{1}})
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0}})
	if p1.Heal != 1 {
		t.Fatalf("expected freeze to resolve after going back, heal=%d", p1.Heal)
	}
}
