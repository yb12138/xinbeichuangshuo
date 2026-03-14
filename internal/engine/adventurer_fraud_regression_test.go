package engine

import (
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func requireChoiceType(t *testing.T, game *GameEngine, playerID, ct string) map[string]interface{} {
	t.Helper()
	if game.State.PendingInterrupt == nil {
		t.Fatalf("expected pending interrupt, got nil")
	}
	if game.State.PendingInterrupt.Type != model.InterruptChoice {
		t.Fatalf("expected choice interrupt, got %s", game.State.PendingInterrupt.Type)
	}
	if game.State.PendingInterrupt.PlayerID != playerID {
		t.Fatalf("expected pending interrupt player=%s, got %s", playerID, game.State.PendingInterrupt.PlayerID)
	}
	ctx, ok := game.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		t.Fatalf("choice context type mismatch")
	}
	got, _ := ctx["choice_type"].(string)
	if got != ct {
		t.Fatalf("expected choice_type=%s, got %s", ct, got)
	}
	return ctx
}

// 回归测试：欺诈(弃2)应支持“先选攻击系别(不含光/暗)”且可选“具体两张同系牌”
func TestAdventurerFraud_Mode2_ElementAndDiscardComboSelectable(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Adventurer", "adventurer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.Phase = model.PhaseActionSelection

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{ID: "f1", Name: "火刃A", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
		{ID: "f2", Name: "火刃B", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
		{ID: "f3", Name: "火刃C", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
		{ID: "w1", Name: "水盾A", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 0},
		{ID: "w2", Name: "水盾B", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 0},
	}

	// 发动欺诈（技能入口）
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdSkill,
		SkillID:   "adventurer_fraud",
		TargetIDs: []string{"p2"},
	})
	requireChoiceType(t, game, "p1", "adventurer_fraud_mode")

	// 选择“弃2同系”
	mustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0}})
	requireChoiceType(t, game, "p1", "adventurer_fraud_attack_element")

	// 选择攻击系别=雷（索引4）
	mustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{4}})
	requireChoiceType(t, game, "p1", "adventurer_fraud_discard_element")

	// 选择弃牌同系=火（在可弃元素 [Water, Fire] 中索引1）
	mustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{1}})
	ctx := requireChoiceType(t, game, "p1", "adventurer_fraud_discard_combo")

	// 火系3选2，应有多个组合供玩家选择
	var combosLen int
	if arr, ok := ctx["combos"].([]string); ok {
		combosLen = len(arr)
	}
	if combosLen < 2 {
		t.Fatalf("expected multiple discard combos for fire 3-choose-2, got %d", combosLen)
	}

	// 选择其中一个组合
	mustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{1}})

	// 欺诈攻击会被自动推进到战斗交互阶段，检查战斗栈中的攻击元素
	if game.State.Phase != model.PhaseCombatInteraction {
		t.Fatalf("expected phase CombatInteraction after fraud resolve, got %s", game.State.Phase)
	}
	if len(game.State.CombatStack) == 0 {
		t.Fatalf("expected combat stack entry after fraud resolve")
	}
	last := game.State.CombatStack[len(game.State.CombatStack)-1]
	if last.Card == nil {
		t.Fatalf("expected combat card not nil")
	}
	if last.Card.Element != model.ElementThunder {
		t.Fatalf("expected fraud attack element Thunder, got %s", last.Card.Element)
	}
	if last.Card.Damage != 2 {
		t.Fatalf("expected fraud attack damage=2, got %d", last.Card.Damage)
	}
	if len(p1.Hand) != 3 {
		t.Fatalf("expected hand count reduced by 2, got %d", len(p1.Hand))
	}
	if p1.Crystal != 1 {
		t.Fatalf("expected lucky fortune crystal gain after fraud attack start, got %d", p1.Crystal)
	}
}
