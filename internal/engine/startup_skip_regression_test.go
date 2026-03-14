package engine

import (
	"strings"
	"testing"

	"starcup-engine/internal/model"
)

// 回归：启动阶段选择“跳过”后，不应在同一回合再次弹出启动技能中断。
func TestStartupSkillSkip_OnlyPromptsOncePerTurn(t *testing.T) {
	game := NewGameEngine(&captureObserver{})
	if err := game.AddPlayer("p1", "Assassin", "assassin", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 1 // 潜行启动技可用

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseStartup

	game.Drive()
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptStartupSkill {
		t.Fatalf("expected startup interrupt, got: %+v", game.State.PendingInterrupt)
	}
	if game.State.PendingInterrupt.PlayerID != "p1" {
		t.Fatalf("expected startup interrupt for p1, got: %s", game.State.PendingInterrupt.PlayerID)
	}

	// 选择“跳过”（下标等于技能数量）。
	skipIdx := len(game.State.PendingInterrupt.SkillIDs)
	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{skipIdx},
	}); err != nil {
		t.Fatalf("skip startup skill failed: %v", err)
	}

	if !p1.TurnState.HasUsedTriggerSkill {
		t.Fatalf("expected HasUsedTriggerSkill=true after skip")
	}
	if game.State.PendingInterrupt != nil && game.State.PendingInterrupt.Type == model.InterruptStartupSkill {
		t.Fatalf("startup interrupt should not reappear in same turn")
	}
	if game.State.Phase != model.PhaseActionSelection {
		t.Fatalf("expected phase to move to ActionSelection, got %s", game.State.Phase)
	}
}

// 回归：启动阶段确认发动一个启动技能后，本回合应立即结束启动阶段，不能继续选择其他启动技能。
func TestStartupSkillConfirm_EndsStartupPhaseAfterOneSkill(t *testing.T) {
	game := NewGameEngine(&captureObserver{})
	if err := game.AddPlayer("p1", "Arbiter", "arbiter", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 1 // 仲裁仪式可用

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseStartup

	game.Drive()
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptStartupSkill {
		t.Fatalf("expected startup interrupt, got: %+v", game.State.PendingInterrupt)
	}

	ritualIdx := -1
	for i, id := range game.State.PendingInterrupt.SkillIDs {
		if id == "arbiter_ritual" {
			ritualIdx = i
			break
		}
	}
	if ritualIdx < 0 {
		t.Fatalf("startup interrupt does not contain arbiter_ritual: %+v", game.State.PendingInterrupt.SkillIDs)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{ritualIdx},
	}); err != nil {
		t.Fatalf("confirm startup skill failed: %v", err)
	}

	if !p1.TurnState.HasUsedTriggerSkill {
		t.Fatalf("expected HasUsedTriggerSkill=true after confirming startup skill")
	}
	if !game.State.HasPerformedStartup {
		t.Fatalf("expected HasPerformedStartup=true after confirming startup skill")
	}
	if p1.Tokens["arbiter_form"] != 1 {
		t.Fatalf("expected arbiter_form=1 after ritual, got %d", p1.Tokens["arbiter_form"])
	}
	if game.State.PendingInterrupt != nil && game.State.PendingInterrupt.Type == model.InterruptStartupSkill {
		t.Fatalf("startup interrupt should not reappear after confirming one startup skill")
	}
	if game.State.Phase != model.PhaseActionSelection {
		t.Fatalf("expected phase to move to ActionSelection, got %s", game.State.Phase)
	}
}

// 回归：本回合执行过启动技能后，不允许执行特殊行动（购买/合成/提炼）。
func TestStartupSkillConfirm_DisablesSpecialActionsInSameTurn(t *testing.T) {
	game := NewGameEngine(&captureObserver{})
	if err := game.AddPlayer("p1", "Arbiter", "arbiter", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 1 // 仲裁仪式可用

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseStartup

	game.Drive()
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptStartupSkill {
		t.Fatalf("expected startup interrupt, got: %+v", game.State.PendingInterrupt)
	}

	ritualIdx := -1
	for i, id := range game.State.PendingInterrupt.SkillIDs {
		if id == "arbiter_ritual" {
			ritualIdx = i
			break
		}
	}
	if ritualIdx < 0 {
		t.Fatalf("startup interrupt does not contain arbiter_ritual: %+v", game.State.PendingInterrupt.SkillIDs)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{ritualIdx},
	}); err != nil {
		t.Fatalf("confirm startup skill failed: %v", err)
	}

	if game.State.Phase != model.PhaseActionSelection {
		t.Fatalf("expected phase to move to ActionSelection, got %s", game.State.Phase)
	}
	if !game.State.HasPerformedStartup {
		t.Fatalf("expected HasPerformedStartup=true after confirming startup skill")
	}

	beforeHand := len(p1.Hand)
	err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdBuy,
	})
	if err == nil {
		t.Fatalf("expected special action to be blocked after startup skill")
	}
	if !strings.Contains(err.Error(), "不能执行特殊行动") {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(p1.Hand) != beforeHand {
		t.Fatalf("buy should not be executed when blocked, before=%d after=%d", beforeHand, len(p1.Hand))
	}
}
