package engine

import (
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

type noopObserver struct{}

func (noopObserver) OnGameEvent(event model.GameEvent) {}

// 回归测试：剑影在同回合后续攻击结束时应继续询问（直到本回合真正发动）
func TestBladeMaster_SwordShadow_ReAskOnEachAttackEnd(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "BladeMaster", "blade_master", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.Phase = model.PhaseActionSelection

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.Heal = 0

	// 两张非风系攻击牌（避免风怒追击干扰）；有蓝水晶满足剑影条件
	p1.Hand = []model.Card{
		{ID: "a1", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
		{ID: "a2", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}
	p1.Crystal = 1
	// 人为放一个额外攻击 token，确保同回合发生第二次攻击行动
	p1.TurnState.PendingActions = append(p1.TurnState.PendingActions, model.ActionContext{
		Source:      "test-token",
		MustType:    "Attack",
		MustElement: nil,
	})

	// 第一次攻击
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0,
	}); err != nil {
		t.Fatalf("first attack failed: %v", err)
	}
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"take"},
	}); err != nil {
		t.Fatalf("first take failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response prompt after first attack end")
	}
	if len(game.State.PendingInterrupt.SkillIDs) != 1 || game.State.PendingInterrupt.SkillIDs[0] != "sword_shadow" {
		t.Fatalf("expected only sword_shadow after first attack end, got %+v", game.State.PendingInterrupt.SkillIDs)
	}
	// 第一次不发动（跳过）
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdCancel,
	}); err != nil {
		t.Fatalf("skip first response failed: %v", err)
	}

	// 推进到第二次攻击行动
	game.Drive()

	// 第二次攻击
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0,
	}); err != nil {
		t.Fatalf("second attack failed: %v", err)
	}
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"take"},
	}); err != nil {
		t.Fatalf("second take failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response prompt again after second attack end")
	}
	if len(game.State.PendingInterrupt.SkillIDs) != 1 || game.State.PendingInterrupt.SkillIDs[0] != "sword_shadow" {
		t.Fatalf("expected only sword_shadow after second attack end, got %+v", game.State.PendingInterrupt.SkillIDs)
	}
}

// 回归测试：风怒追击在同回合后续攻击结束时应继续询问（直到本回合真正发动）
func TestBladeMaster_WindFury_ReAskOnEachAttackEnd(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "BladeMaster", "blade_master", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.Phase = model.PhaseActionSelection

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.Heal = 0

	// 三张风系攻击牌：前两张用于两次攻击，保留一张确保第二次攻击结束时仍满足风怒可用条件
	p1.Hand = []model.Card{
		{ID: "a1", Name: "风斩", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 1},
		{ID: "a2", Name: "风斩", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 1},
		{ID: "a3", Name: "风斩", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 1},
	}
	p1.Gem = 0
	p1.Crystal = 0
	p1.TurnState.PendingActions = append(p1.TurnState.PendingActions, model.ActionContext{
		Source:      "test-token",
		MustType:    "Attack",
		MustElement: nil,
	})

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0,
	}); err != nil {
		t.Fatalf("first attack failed: %v", err)
	}
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"take"},
	}); err != nil {
		t.Fatalf("first take failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response prompt after first attack end")
	}
	if len(game.State.PendingInterrupt.SkillIDs) != 1 || game.State.PendingInterrupt.SkillIDs[0] != "wind_fury" {
		t.Fatalf("expected only wind_fury after first attack end, got %+v", game.State.PendingInterrupt.SkillIDs)
	}
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdCancel,
	}); err != nil {
		t.Fatalf("skip first response failed: %v", err)
	}

	game.Drive()

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0,
	}); err != nil {
		t.Fatalf("second attack failed: %v", err)
	}
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"take"},
	}); err != nil {
		t.Fatalf("second take failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response prompt again after second attack end")
	}
	if len(game.State.PendingInterrupt.SkillIDs) != 1 || game.State.PendingInterrupt.SkillIDs[0] != "wind_fury" {
		t.Fatalf("expected only wind_fury after second attack end, got %+v", game.State.PendingInterrupt.SkillIDs)
	}
}
