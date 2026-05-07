package adventurer_test

import (
	"starcup-engine/internal/engine"
	"starcup-engine/internal/testutils"
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func requireChoiceType(t *testing.T, game *engine.GameEngine, playerID, ct string) map[string]interface{} {
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

// 回归测试：欺诈应先在手牌区选择同系牌；选择2张后再选择五系攻击元素。
func TestAdventurerFraud_PickTwoThenChooseAttackElement(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Adventurer", "adventurer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

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
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdSkill,
		SkillID:   "adventurer_fraud",
		TargetIDs: []string{"p2"},
	})
	requireChoiceType(t, game, "p1", "adventurer_fraud_pick")

	// 先在手牌区选择2张同系牌（火）
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0, 1}})
	requireChoiceType(t, game, "p1", "adventurer_fraud_attack_element")

	prompt := game.BuildPendingInterruptPrompt()
	if prompt == nil {
		t.Fatalf("expected fraud attack element prompt")
	}
	if prompt.Type != model.PromptConfirm {
		t.Fatalf("expected confirm prompt for fraud attack element, got %s", prompt.Type)
	}
	if got := len(prompt.Options); got != 5 {
		t.Fatalf("expected 5 attack element options, got %d", got)
	}
	wantIDs := []string{
		string(model.ElementWater),
		string(model.ElementFire),
		string(model.ElementEarth),
		string(model.ElementWind),
		string(model.ElementThunder),
	}
	for idx, want := range wantIDs {
		if idx >= len(prompt.Options) || prompt.Options[idx].ID != want {
			t.Fatalf("expected option[%d]=%s, got %+v", idx, want, prompt.Options)
		}
	}

	// 选择攻击系别=雷（索引4）
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{4}})

	// 欺诈攻击会被自动推进到战斗交互阶段，检查战斗栈中的攻击元素
	if len(game.State.CombatStack) == 0 || (game.State.CombatStage != model.CombatStageDeclare && game.State.CombatStage != model.CombatStageHitCheck) {
		t.Fatalf("expected combat interaction state after fraud resolve, got combat=%s stack=%d", game.State.CombatStage, len(game.State.CombatStack))
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

func TestAdventurerFraud_PickThreeAutoConvertsToDark(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Adventurer", "adventurer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{ID: "f1", Name: "火刃A", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
		{ID: "f2", Name: "火刃B", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
		{ID: "f3", Name: "火刃C", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
		{ID: "w1", Name: "水盾A", Type: model.CardTypeMagic, Element: model.ElementWater},
	}

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdSkill,
		SkillID:   "adventurer_fraud",
		TargetIDs: []string{"p2"},
	})
	requireChoiceType(t, game, "p1", "adventurer_fraud_pick")

	// 直接选择3张同系牌，应自动转暗灭攻击，不再弹攻击系别选择框。
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0, 1, 2}})

	if got := len(game.State.CombatStack); got == 0 {
		t.Fatalf("expected combat stack entry after fraud dark convert")
	}
	last := game.State.CombatStack[len(game.State.CombatStack)-1]
	if last.Card == nil || last.Card.Element != model.ElementDark {
		t.Fatalf("expected fraud auto convert to Dark, got %+v", last.Card)
	}
	if last.CanBeResponded {
		t.Fatalf("expected dark fraud attack cannot be responded")
	}
	if len(p1.Hand) != 1 {
		t.Fatalf("expected hand count reduced by 3, got %d", len(p1.Hand))
	}
	if p1.Crystal != 1 {
		t.Fatalf("expected lucky fortune crystal gain after dark fraud attack, got %d", p1.Crystal)
	}
}
