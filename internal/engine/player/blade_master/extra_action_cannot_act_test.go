package blade_master_test

import (
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

// promptCaptureObserver captures the last AskInput prompt for inspection.
type promptCaptureObserver struct {
	lastPrompt *model.Prompt
}

func (o *promptCaptureObserver) OnGameEvent(event model.GameEvent) {
	if event.Type != model.EventAskInput {
		return
	}
	if p := event.Prompt; p != nil {
		cp := *p
		cp.Options = append([]model.PromptOption(nil), p.Options...)
		o.lastPrompt = &cp
	}
}

// TestSkipExtraAction_ViaDriveLoop 验证通过完整 Drive 循环，
// 额外行动阶段点"跳过额外行动"后：手牌不变、进入 TurnEnd、不再弹出行动选择。
func TestSkipExtraAction_ViaDriveLoop(t *testing.T) {
	obs := &promptCaptureObserver{}
	game := engine.NewGameEngine(obs)

	if err := game.AddPlayer("p1", "BladeMaster", "blade_master", model.RedCamp); err != nil {
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

	// 手牌只有火系攻击牌（无风系），设置额外风系攻击行动约束
	p1.Hand = []model.Card{
		{ID: "a1", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
		{ID: "a2", Name: "火斩2", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}
	p1.TurnState.CurrentExtraAction = "Attack"
	p1.TurnState.CurrentExtraElement = []model.Element{model.ElementWind}

	handBefore := len(p1.Hand)

	// Drive 生成行动选择提示
	game.Drive()

	if obs.lastPrompt == nil {
		t.Fatal("expected action selection prompt after Drive")
	}

	// 验证提示消息包含"额外行动"
	foundExtra := false
	for _, opt := range obs.lastPrompt.Options {
		if opt.ID == "cannot_act" {
			foundExtra = true
			if opt.Label != "跳过额外行动" {
				t.Fatalf("expected cannot_act label '跳过额外行动', got: %s", opt.Label)
			}
		}
	}
	if !foundExtra {
		t.Fatalf("expected cannot_act option in prompt, options: %+v", obs.lastPrompt.Options)
	}

	// 点击"跳过额外行动"
	obs.lastPrompt = nil
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdCannotAct,
	}); err != nil {
		t.Fatalf("CannotAct failed: %v", err)
	}
	game.Drive()

	handAfter := len(p1.Hand)
	if handAfter != handBefore {
		t.Fatalf("hand count should not change after skipExtraAction: before=%d after=%d", handBefore, handAfter)
	}

	// p1 不应再处于活跃状态（回合已结束，轮到 p2）
	if p1.IsActive {
		t.Fatalf("p1 should NOT be active after skipping extra action and Drive")
	}
	// CurrentExtraAction 应已清空
	if p1.TurnState.CurrentExtraAction != "" {
		t.Fatalf("CurrentExtraAction should be cleared, got: %s", p1.TurnState.CurrentExtraAction)
	}
}

// TestSkipExtraAction_MultiplePendingActions 验证跳过一个额外行动后，
// Drive 正确取出下一个额外行动并再次弹出行动选择。
func TestSkipExtraAction_MultiplePendingActions(t *testing.T) {
	obs := &promptCaptureObserver{}
	game := engine.NewGameEngine(obs)

	if err := game.AddPlayer("p1", "BladeMaster", "blade_master", model.RedCamp); err != nil {
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

	// 手牌只有火系攻击牌
	p1.Hand = []model.Card{
		{ID: "a1", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}

	// 设置两个额外行动：风系攻击 + 火系攻击
	p1.TurnState.CurrentExtraAction = "Attack"
	p1.TurnState.CurrentExtraElement = []model.Element{model.ElementWind}
	model.AppendAttackAction(p1, "测试额外火系攻击", model.ElementFire)

	handBefore := len(p1.Hand)

	// Drive 生成第一个额外行动提示
	game.Drive()

	// 跳过风系额外行动（手牌无风系牌）
	obs.lastPrompt = nil
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdCannotAct,
	}); err != nil {
		t.Fatalf("CannotAct failed: %v", err)
	}
	game.Drive()

	handAfter := len(p1.Hand)
	if handAfter != handBefore {
		t.Fatalf("hand count should not change: before=%d after=%d", handBefore, handAfter)
	}

	// 验证第二个额外行动（火系攻击）被取出
	// 手牌有火系攻击牌，所以应该能看到 attack 选项
	if obs.lastPrompt == nil {
		t.Fatal("expected new prompt after skipping first extra action")
	}
	foundAttack := false
	for _, opt := range obs.lastPrompt.Options {
		if opt.ID == "attack" {
			foundAttack = true
		}
	}
	if !foundAttack {
		t.Fatalf("expected attack option in second extra action prompt, options: %+v", obs.lastPrompt.Options)
	}

	// CurrentExtraAction 应为 "Attack"（火系）
	if p1.TurnState.CurrentExtraAction != "Attack" {
		t.Fatalf("expected CurrentExtraAction=Attack for second extra action, got: %s", p1.TurnState.CurrentExtraAction)
	}
}
