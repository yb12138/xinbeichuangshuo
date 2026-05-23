package assassin_test

import (
	"starcup-engine/internal/engine"
	"starcup-engine/internal/testutils"
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

// 回归：暗杀者在受伤摸牌前出现【水影】响应时，选择“跳过”后必须回到伤害结算流程，
// 不能停留在 Response 子流程导致 Drive 空转。
func TestAssassinWaterShadowSkip_ResumesPendingDamageResolution(t *testing.T) {
	game := engine.NewGameEngine(&testutils.CaptureObserver{})
	if err := game.AddPlayer("p1", "Attacker", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Assassin", "assassin", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	game.State.Deck = rules.InitDeck()

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p2.Heal = 0

	// p1 打出主动攻击牌；p2 预留一张水系牌，确保可触发水影可选响应。
	p1.Hand = []model.Card{
		{ID: "atk-fire-1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
	}
	p2.Hand = []model.Card{
		{ID: "water-card-1", Name: "水弹", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 1},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdAttack,
		TargetID: "p2",
		CardID:   testutils.PlayableCardID(t, game, "p1", 0),
	}); err != nil {
		t.Fatalf("attack failed: %v", err)
	}
	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	}); err != nil {
		t.Fatalf("take failed: %v", err)
	}

	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response-skill interrupt after damage, got %+v", game.State.PendingInterrupt)
	}
	if game.State.PendingInterrupt.PlayerID != "p2" {
		t.Fatalf("expected response interrupt for p2, got %s", game.State.PendingInterrupt.PlayerID)
	}
	hasWaterShadow := false
	for _, sid := range game.State.PendingInterrupt.SkillIDs {
		if sid == "water_shadow" {
			hasWaterShadow = true
			break
		}
	}
	if !hasWaterShadow {
		t.Fatalf("expected water_shadow in response skills, got %+v", game.State.PendingInterrupt.SkillIDs)
	}

	// 机器人路径：Select 选择“跳过”（索引等于技能数量）
	skipIdx := len(game.State.PendingInterrupt.SkillIDs)
	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p2",
		Type:       model.CmdSelect,
		Selections: []int{skipIdx},
	}); err != nil {
		t.Fatalf("skip response failed: %v", err)
	}

	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no pending interrupt after skip, got %+v", game.State.PendingInterrupt)
	}
	if game.State.Subflow == model.SubflowResponse {
		t.Fatalf("subflow should not stay in response after skip (would stall drive)")
	}
}

// 回归：暗杀者发动【水影】后，仅弃置水系牌，不改变摸牌数；
// 弃牌后摸牌数应保持不变。
func TestAssassinWaterShadowConfirm_PreservesRemainingDamageDraw(t *testing.T) {
	game := engine.NewGameEngine(&testutils.CaptureObserver{})
	if err := game.AddPlayer("p1", "Attacker", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Assassin", "assassin", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	game.State.Deck = rules.InitDeck()

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p2.Heal = 0

	// p2 使用水影只弃1张水系牌时，仍应保留“原伤害摸牌 - 1”的剩余摸牌。
	p1.Hand = []model.Card{
		{ID: "atk-fire-1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
	}
	p2.Hand = []model.Card{
		{ID: "water-card-1", Name: "水弹", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 1},
		{ID: "fire-card-1", Name: "炎爆", Type: model.CardTypeMagic, Element: model.ElementFire, Damage: 1},
	}
	initialHand := len(p2.Hand)

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardID: testutils.PlayableCardID(t, game, "p1", 0),
	}); err != nil {
		t.Fatalf("attack failed: %v", err)
	}
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"take"},
	}); err != nil {
		t.Fatalf("take failed: %v", err)
	}

	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response-skill interrupt after damage, got %+v", game.State.PendingInterrupt)
	}
	rawCtx, ok := game.State.PendingInterrupt.Context.(*model.Context)
	if !ok || rawCtx == nil || rawCtx.EventCtx == nil || rawCtx.EventCtx.DrawCount == nil {
		t.Fatalf("expected before-draw context on water_shadow interrupt, got %+v", game.State.PendingInterrupt.Context)
	}
	pendingDrawCount := *rawCtx.EventCtx.DrawCount
	if pendingDrawCount <= 0 {
		t.Fatalf("expected positive pending draw count before water_shadow, got %d", pendingDrawCount)
	}
	skillIdx := -1
	for i, sid := range game.State.PendingInterrupt.SkillIDs {
		if sid == "water_shadow" {
			skillIdx = i
			break
		}
	}
	if skillIdx < 0 {
		t.Fatalf("expected water_shadow in response skills, got %+v", game.State.PendingInterrupt.SkillIDs)
	}
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p2", Type: model.CmdSelect, Selections: []int{skillIdx},
	}); err != nil {
		t.Fatalf("choose water_shadow failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || !engine.IsDiscardSelectionInterrupt(game.State.PendingInterrupt) {
		t.Fatalf("expected discard interrupt after choosing water_shadow, got %+v", game.State.PendingInterrupt)
	}

	// 弃置第 1 张（水系）发动水影。
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p2", Type: model.CmdSelect, Selections: []int{0},
	}); err != nil {
		t.Fatalf("confirm water_shadow discard failed: %v", err)
	}

	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no pending interrupt after water_shadow resolves, got %+v", game.State.PendingInterrupt)
	}
	// 水影不改变摸牌数，仅弃置水系牌
	expectedHand := initialHand - 1 + pendingDrawCount
	if got := len(p2.Hand); got != expectedHand {
		t.Fatalf("expected hand size %d (initial=%d discard=1 draw=%d), got %d", expectedHand, initialHand, pendingDrawCount, got)
	}
}
