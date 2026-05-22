package onmyoji_test

import (
	"starcup-engine/internal/engine"
	"starcup-engine/internal/engine/core/runtimeutil"
	"starcup-engine/internal/testutils"
	"testing"

	"starcup-engine/internal/model"
)

func choiceTypeOf(intr *model.Interrupt) string {
	if intr == nil {
		return ""
	}
	data, ok := intr.Context.(map[string]interface{})
	if !ok {
		return ""
	}
	v, _ := data["choice_type"].(string)
	return v
}

func TestOnmyojiDarkRitual_ChoosesTargetAtTurnEnd(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Onmyoji", "onmyoji", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["onmyoji_ghost_fire"] = 3

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageTurnEnd

	game.Drive()

	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptChoice {
		t.Fatalf("expected choice interrupt for dark ritual, got %+v", game.State.PendingInterrupt)
	}
	if got := choiceTypeOf(game.State.PendingInterrupt); got != "onmyoji_dark_ritual_target" {
		t.Fatalf("expected onmyoji_dark_ritual_target prompt, got %s", got)
	}
	ctxData, _ := game.State.PendingInterrupt.Context.(map[string]interface{})
	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
	if len(targetIDs) != 1 || targetIDs[0] != "p2" {
		t.Fatalf("expected dark ritual target pool only include enemy p2, got %+v", targetIDs)
	}
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("choose dark ritual target failed: %v", err)
	}
	if got := p1.Tokens["onmyoji_ghost_fire"]; got != 0 {
		t.Fatalf("expected ghost fire reset to 0, got %d", got)
	}
	if len(game.State.PendingDamageQueue) == 0 {
		t.Fatalf("expected pending damage from dark ritual")
	}
	pd := game.State.PendingDamageQueue[0]
	if pd.SourceID != "p1" || pd.TargetID != "p2" || pd.Damage != 2 || pd.DamageType != "magic" {
		t.Fatalf("unexpected dark ritual pending damage: %+v", pd)
	}
}

func TestOnmyojiBinding_RequiresGemAndCrystal(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Attacker", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "TargetAlly", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Onmyoji", "onmyoji", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p4", "AttackerAlly", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	p3 := game.State.Players["p3"]
	p3.Form = model.FormOnmyojiShikigami
	p3.Hand = []model.Card{
		{ID: "c1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Faction: "咏", Damage: 2},
	}

	req := model.CombatRequest{
		AttackerID:     "p1",
		TargetID:       "p2",
		Card:           &model.Card{ID: "atk", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Faction: "咏", Damage: 2},
		CanBeResponded: true,
	}

	// 仅2宝石，0水晶：不满足式神咒束代应战成本
	game.State.BlueGems = 2
	game.State.BlueCrystals = 0
	if game.RunAttackResponseCombatInteractionPolicies(&req) {
		t.Fatalf("binding should not start without crystal")
	}

	// 1宝石+1水晶：可触发询问
	req2 := model.CombatRequest{
		AttackerID:     "p1",
		TargetID:       "p2",
		Card:           &model.Card{ID: "atk2", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Faction: "咏", Damage: 2},
		CanBeResponded: true,
	}
	game.State.BlueGems = 1
	game.State.BlueCrystals = 1
	if !game.RunAttackResponseCombatInteractionPolicies(&req2) {
		t.Fatalf("binding should start with 1 gem + 1 crystal")
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptChoice {
		t.Fatalf("expected binding confirm interrupt, got %+v", game.State.PendingInterrupt)
	}
	if got := choiceTypeOf(game.State.PendingInterrupt); got != "onmyoji_binding_confirm" {
		t.Fatalf("expected onmyoji_binding_confirm, got %s", got)
	}
}

func TestOnmyojiBinding_FullFlow(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Attacker", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "TargetAlly", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Onmyoji", "onmyoji", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p4", "AttackerAlly", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	p3 := game.State.Players["p3"]
	p3.Form = model.FormOnmyojiShikigami
	p3.Hand = []model.Card{
		{ID: "c1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Faction: "咏", Damage: 2},
	}

	game.State.BlueGems = 1
	game.State.BlueCrystals = 1

	req := model.CombatRequest{
		AttackerID:     "p1",
		TargetID:       "p2",
		Card:           &model.Card{ID: "atk", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Faction: "咏", Damage: 2},
		CanBeResponded: true,
	}
	if !game.RunAttackResponseCombatInteractionPolicies(&req) {
		t.Fatalf("binding should start")
	}

	// Step 1: confirm prompt
	testutils.RequireChoicePrompt(t, game, "p3", "onmyoji_binding_confirm")

	// Choose "yes" (index 0)
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p3",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})

	// Step 2: should have card selection prompt
	testutils.RequireChoicePrompt(t, game, "p3", "onmyoji_binding_card")

	// Choose card (index 0)
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p3",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})

	// Step 3: should have counter target prompt
	testutils.RequireChoicePrompt(t, game, "p3", "onmyoji_binding_counter_target")

	// Choose target (p4 is attacker's ally, index 0)
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p3",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})

	// Flow should complete
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no pending interrupt after full flow, got %+v", game.State.PendingInterrupt)
	}
}

func TestOnmyojiBinding_WithYinyangConversion(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Attacker", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "TargetAlly", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Onmyoji", "onmyoji", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p4", "AttackerAlly", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	p3 := game.State.Players["p3"]
	p3.Form = model.FormOnmyojiShikigami
	if p3.Tokens == nil {
		p3.Tokens = map[string]int{}
	}
	p3.Tokens["onmyoji_ghost_fire"] = 2 // 已有2鬼火
	// 使用同命格牌（咏命格）而非同系牌，触发阴阳转换
	p3.Hand = []model.Card{
		{ID: "c1", Name: "地裂斩", Type: model.CardTypeAttack, Element: model.ElementEarth, Faction: "咏", Damage: 2},
	}

	game.State.BlueGems = 1
	game.State.BlueCrystals = 1

	// Push initial combat request (attacker attacking target)
	game.State.CombatStack = []model.CombatRequest{
		{
			AttackerID:     "p1",
			TargetID:       "p2",
			Card:           &model.Card{ID: "atk", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Faction: "咏", Damage: 2},
			CanBeResponded: true,
		},
	}

	req := model.CombatRequest{
		AttackerID:     "p1",
		TargetID:       "p2",
		Card:           &model.Card{ID: "atk", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Faction: "咏", Damage: 2},
		CanBeResponded: true,
	}
	if !game.RunAttackResponseCombatInteractionPolicies(&req) {
		t.Fatalf("binding should start")
	}

	// Step 1: confirm prompt
	testutils.RequireChoicePrompt(t, game, "p3", "onmyoji_binding_confirm")
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p3",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})

	// Step 2: card selection - should have yinyang conversion option
	testutils.RequireChoicePrompt(t, game, "p3", "onmyoji_binding_card")
	// Choose card (index 0)
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p3",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})

	// Step 3: counter target
	testutils.RequireChoicePrompt(t, game, "p3", "onmyoji_binding_counter_target")
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p3",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})

	// Flow should complete
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no pending interrupt after full flow, got %+v", game.State.PendingInterrupt)
	}

	// Verify 阴阳转换 effects:
	// 1. 鬼火应该增加到3（从2增加到3）
	if p3.Tokens["onmyoji_ghost_fire"] != 3 {
		t.Fatalf("expected ghost fire to be 3 after yinyang conversion, got %d", p3.Tokens["onmyoji_ghost_fire"])
	}
	// 2. 应该脱离式神形态
	if p3.Form == model.FormOnmyojiShikigami {
		t.Fatalf("expected to leave shikigami form after yinyang conversion")
	}
	// 3. 应战战斗请求应该已推入战斗栈
	if len(game.State.CombatStack) == 0 {
		t.Fatalf("expected counter combat request in combat stack")
	}
	// 4. 应战战斗请求的卡牌伤害应该是鬼火数=3（阴阳转换效果）
	counterReq := game.State.CombatStack[len(game.State.CombatStack)-1]
	if counterReq.Card == nil {
		t.Fatalf("expected counter card to exist")
	}
	if counterReq.Card.Damage != 3 {
		t.Fatalf("expected counter card damage to be 3 (ghost fire count), got %d", counterReq.Card.Damage)
	}
	// 5. 应战者应该是阴阳师 p3
	if counterReq.AttackerID != "p3" {
		t.Fatalf("expected counter attacker to be p3, got %s", counterReq.AttackerID)
	}
	// 6. 应战目标应该是攻击者的队友 p4
	if counterReq.TargetID != "p4" {
		t.Fatalf("expected counter target to be p4, got %s", counterReq.TargetID)
	}
}
