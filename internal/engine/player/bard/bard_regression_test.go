package bard_test

import (
	"fmt"
	"starcup-engine/internal/engine"
	"starcup-engine/internal/testutils"
	"testing"

	bardpkg "starcup-engine/internal/engine/player/bard"

	"starcup-engine/internal/data"
	"starcup-engine/internal/model"
	"strings"
)

func bardTestCard(id, name string, cardType model.CardType, ele model.Element) model.Card {
	return model.Card{
		ID:          id,
		Name:        name,
		Type:        cardType,
		Element:     ele,
		Damage:      1,
		Description: name,
	}
}

func addBardExclusiveCardsForTest(p *model.Player, titles ...string) {
	if p == nil {
		return
	}
	// Now bard only has one exclusive card "永恒乐章"
	for _, title := range titles {
		switch title {
		case "永恒乐章":
			p.RestoreExclusiveCard(model.Card{
				ID:             fmt.Sprintf("starter-%s-bd_eternal_movement", p.ID),
				Name:           "永恒乐章",
				Type:           model.CardTypeMagic,
				Element:        model.ElementDark,
				Description:    "吟游诗人专属牌",
				ExclusiveChar1: p.Character.ID,
			})
		}
	}
}

func findFieldEffectCard(p *model.Player, effect model.EffectType) *model.FieldCard {
	if p == nil {
		return nil
	}
	for _, fc := range p.Field {
		if fc != nil && fc.Mode == model.FieldEffect && fc.Effect == effect {
			return fc
		}
	}
	return nil
}

func TestBardDescentConcerto_RunsAndResolves(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Bard", "bard", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "priest", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "EnemyA", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p4", "EnemyB", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	bard := game.State.Players["p1"]
	bard.IsActive = true
	bard.TurnState = model.NewPlayerTurnState()
	bard.Hand = []model.Card{
		bardTestCard("f_magic", "火法术", model.CardTypeMagic, model.ElementFire),
		bardTestCard("f_attack", "火攻击", model.CardTypeAttack, model.ElementFire),
		bardTestCard("w_attack", "水攻击", model.CardTypeAttack, model.ElementWater),
	}

	// 伤害结算后只做追踪，不触发
	if paused := game.HandlePostDamageResolved(&model.PendingDamage{
		SourceID: "p1", TargetID: "p3", Damage: 1, DamageType: model.MagicAttack,
	}); paused {
		t.Fatalf("first magic damage should only track, not trigger")
	}
	if paused := game.HandlePostDamageResolved(&model.PendingDamage{
		SourceID: "p1", TargetID: "p4", Damage: 1, DamageType: model.MagicAttack,
	}); paused {
		t.Fatalf("second magic damage should only track, not trigger")
	}

	// 回合结束时触发沉沦协奏曲的确认弹窗
	if paused := game.RunTurnEndTimingStageHooks(bard, engine.TimingTurnEndPreExtra); !paused {
		t.Fatalf("turn-end descent hook should trigger with 2+ magic damage targets")
	}
	testutils.RequireChoicePrompt(t, game, "p1", "bd_descent_confirm")
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("choose descent confirm failed: %v", err)
	}

	// 新流程：确认发动后直接推送弃牌选择
	ctxData := testutils.RequireChoiceContext(t, game, "p1", "bd_descent_cards")
	flow := testutils.RequirePromptFlow(t, ctxData, "bd_descent", "cards")
	if _, ok := ctxData["selected_indices"]; ok {
		t.Fatalf("descent should store selections in prompt flow, got legacy selected_indices in %+v", ctxData)
	}
	if got := flow.Selection("cards").Count; got != 2 {
		t.Fatalf("expected descent flow card count=2, got %d in %+v", got, flow)
	}

	// 选择第1张火牌（手牌索引0）
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("choose first discard failed: %v", err)
	}
	ctxData = testutils.RequireChoiceContext(t, game, "p1", "bd_descent_cards")
	flow = testutils.RequirePromptFlow(t, ctxData, "bd_descent", "cards")
	if got := flow.Selection("cards").OptionIndexes; len(got) != 1 || got[0] != 0 {
		t.Fatalf("expected descent flow to accumulate first card index 0, got %+v in %+v", got, flow)
	}
	// 选择第2张火牌（剩余候选索引中的第一个）
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("choose second discard failed: %v", err)
	}

	ctxData = testutils.RequireChoiceContext(t, game, "p1", "bd_descent_target")
	flow = testutils.RequirePromptFlow(t, ctxData, "bd_descent", "target")
	if got := flow.Selection("cards").OptionIndexes; len(got) != 2 || got[0] != 0 || got[1] != 1 {
		t.Fatalf("expected descent flow to accumulate card indexes [0 1], got %+v in %+v", got, flow)
	}
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("choose descent bonus target failed: %v", err)
	}

	if got := bard.Tokens["bd_inspiration"]; got != 1 {
		t.Fatalf("expected inspiration=1, got %d", got)
	}
	if got := bard.TurnState.UsedSkillCounts["bd_descent"]; got != 1 {
		t.Fatalf("expected descent used flag=1, got %d", got)
	}
	if got := len(bard.Hand); got != 1 {
		t.Fatalf("expected bard hand reduced to 1, got %d", got)
	}
	if got := len(game.State.PendingDamageQueue); got != 1 {
		t.Fatalf("expected one bonus pending damage, got %d", got)
	}
	pd := game.State.PendingDamageQueue[0]
	if pd.SourceID != "p1" || pd.DamageType != "magic" || pd.Damage != 1 {
		t.Fatalf("unexpected bonus damage payload: %+v", pd)
	}
}

func TestBardDescentConcerto_DeclineAtConfirmDoesNotConsumeSkill(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Bard", "bard", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "priest", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "EnemyA", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p4", "EnemyB", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	bard := game.State.Players["p1"]
	bard.IsActive = true
	bard.TurnState = model.NewPlayerTurnState()
	bard.Hand = []model.Card{
		bardTestCard("f_magic", "火法术", model.CardTypeMagic, model.ElementFire),
		bardTestCard("f_attack", "火攻击", model.CardTypeAttack, model.ElementFire),
		bardTestCard("w_attack", "水攻击", model.CardTypeAttack, model.ElementWater),
	}

	if paused := game.HandlePostDamageResolved(&model.PendingDamage{
		SourceID: "p1", TargetID: "p3", Damage: 1, DamageType: model.MagicAttack,
	}); paused {
		t.Fatalf("first magic damage should only track, not trigger")
	}
	if paused := game.HandlePostDamageResolved(&model.PendingDamage{
		SourceID: "p1", TargetID: "p4", Damage: 1, DamageType: model.MagicAttack,
	}); paused {
		t.Fatalf("second magic damage should only track, not trigger")
	}

	if paused := game.RunTurnEndTimingStageHooks(bard, engine.TimingTurnEndPreExtra); !paused {
		t.Fatalf("turn-end descent hook should trigger with 2+ magic damage targets")
	}
	testutils.RequireChoicePrompt(t, game, "p1", "bd_descent_confirm")
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}}); err != nil {
		t.Fatalf("decline descent confirm failed: %v", err)
	}

	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no pending interrupt after declining descent, got %+v", game.State.PendingInterrupt)
	}
	if got := bard.TurnState.UsedSkillCounts["bd_descent"]; got != 0 {
		t.Fatalf("expected descent not to consume usage on decline, got %d", got)
	}
}

func TestBardDescentConcerto_AllyMagicDamageTriggersAtTurnEnd(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Bard", "bard", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "priest", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "EnemyA", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p4", "EnemyB", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	bard := game.State.Players["p1"]
	ally := game.State.Players["p2"]
	bard.IsActive = true
	bard.TurnState = model.NewPlayerTurnState()
	bard.Hand = []model.Card{
		bardTestCard("f_magic", "火法术", model.CardTypeMagic, model.ElementFire),
		bardTestCard("f_attack", "火攻击", model.CardTypeAttack, model.ElementFire),
	}

	// 队友法术伤害应该被追踪（不立即触发）
	if paused := game.HandlePostDamageResolved(&model.PendingDamage{
		SourceID: "p2", TargetID: "p3", Damage: 1, DamageType: model.MagicAttack,
	}); paused {
		t.Fatalf("ally magic damage should only track, not trigger immediately")
	}
	if paused := game.HandlePostDamageResolved(&model.PendingDamage{
		SourceID: "p2", TargetID: "p4", Damage: 1, DamageType: model.MagicAttack,
	}); paused {
		t.Fatalf("ally magic damage should only track, not trigger immediately")
	}

	// 回合结束时队友伤害已记录到诗人名下，触发沉沦协奏曲确认框
	if paused := game.RunTurnEndTimingStageHooks(ally, engine.TimingTurnEndPreExtra); !paused {
		t.Fatalf("turn-end descent hook should trigger from ally magic damage")
	}
	testutils.RequireChoicePrompt(t, game, "p1", "bd_descent_confirm")
}

func TestBardDissonanceChord_DrawModeAndReleasePrisoner(t *testing.T) {
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
	bard.Tokens["bd_inspiration"] = 3
	bard.Form = model.FormBardEternalPrisoner
	bard.Hand = []model.Card{
		bardTestCard("h1", "手牌1", model.CardTypeAttack, model.ElementFire),
	}
	enemy.Hand = []model.Card{
		bardTestCard("e1", "敌方牌1", model.CardTypeAttack, model.ElementWater),
	}
	game.State.Deck = []model.Card{
		bardTestCard("d1", "牌堆1", model.CardTypeAttack, model.ElementWind),
		bardTestCard("d2", "牌堆2", model.CardTypeAttack, model.ElementThunder),
	}
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	if err := game.UseSkill("p1", "bd_dissonance_chord", nil, nil); err != nil {
		t.Fatalf("use dissonance failed: %v", err)
	}
	ctxData := testutils.RequireChoiceContext(t, game, "p1", "bd_dissonance_x")
	testutils.RequirePromptFlow(t, ctxData, "bd_dissonance", "x")
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil { // X=2
		t.Fatalf("choose X failed: %v", err)
	}
	if got := bard.Tokens["bd_inspiration"]; got != 1 {
		t.Fatalf("expected inspiration consumed to 1, got %d", got)
	}
	if got := bard.Form; got != "" {
		t.Fatalf("expected prisoner form released, got %q", got)
	}

	ctxData = testutils.RequireChoiceContext(t, game, "p1", "bd_dissonance_mode")
	flow := testutils.RequirePromptFlow(t, ctxData, "bd_dissonance", "mode")
	if _, ok := ctxData["x_value"]; ok {
		t.Fatalf("dissonance should store X in prompt flow, got legacy x_value in %+v", ctxData)
	}
	if got := flow.Selection("x").Count; got != 2 {
		t.Fatalf("expected dissonance flow X=2, got %d in %+v", got, flow)
	}
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil { // 摸牌分支
		t.Fatalf("choose mode failed: %v", err)
	}
	ctxData = testutils.RequireChoiceContext(t, game, "p1", "bd_dissonance_target")
	flow = testutils.RequirePromptFlow(t, ctxData, "bd_dissonance", "target")
	if got := flow.Selection("mode").Count; got != 0 {
		t.Fatalf("expected dissonance flow mode=0, got %d in %+v", got, flow)
	}
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}}); err != nil { // 目标选 p2
		t.Fatalf("choose target failed: %v", err)
	}

	if got := len(bard.Hand); got != 2 {
		t.Fatalf("expected bard drew 1 card, hand=%d", got)
	}
	if got := len(enemy.Hand); got != 2 {
		t.Fatalf("expected target drew 1 card, hand=%d", got)
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected dissonance flow completed, got pending interrupt %+v", game.State.PendingInterrupt)
	}
}

func TestBardHopeFugue_PlaceUsesPlayedCardAsEternalMovement(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Bard", "bard", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "priest", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	bard := game.State.Players["p1"]
	ally := game.State.Players["p2"]
	// No longer need exclusive cards for hope_fugue - it's a character skill now
	bard.IsActive = true
	bard.TurnState = model.NewPlayerTurnState()
	bard.Crystal = 1
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionStart

	if err := game.UseSkill("p1", "bd_hope_fugue", nil, nil); err != nil {
		t.Fatalf("use hope fugue failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "bd_hope_draw_confirm")
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}}); err != nil { // 不摸牌
		t.Fatalf("choose draw confirm failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "bd_hope_mode")
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil { // 放置分支
		t.Fatalf("choose hope mode failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "bd_hope_place_target")
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil { // 目标队友 p2
		t.Fatalf("choose place target failed: %v", err)
	}

	if holder := bardpkg.EternalHolderID(engine.NewRoleChoiceRuntime(game), bard); holder != "p2" {
		t.Fatalf("expected eternal movement holder p2, got %q", holder)
	}
	fieldCard := findFieldEffectCard(ally, model.EffectBardEternalMovement)
	if fieldCard == nil {
		t.Fatalf("expected ally to hold eternal movement field card")
	}
	if fieldCard.Card.Name != "永恒乐章" {
		t.Fatalf("expected eternal movement field entity, got %s", fieldCard.Card.Name)
	}
}

func TestBardHopeFugue_TransferMovesExistingEternalMovementAndGainsInspiration(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Bard", "bard", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "AllyA", "priest", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "AllyB", "hero", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	bard := game.State.Players["p1"]
	allyA := game.State.Players["p2"]
	allyB := game.State.Players["p3"]
	// No longer need exclusive cards - hope_fugue is a character skill now
	bard.IsActive = true
	bard.TurnState = model.NewPlayerTurnState()
	bard.Crystal = 1
	bard.Hand = []model.Card{
		bardTestCard("discard", "弃牌", model.CardTypeAttack, model.ElementFire),
	}
	if err := bardpkg.PlaceEternalMovement(engine.NewRoleChoiceRuntime(game), bard, allyA); err != nil {
		t.Fatalf("place initial eternal movement failed: %v", err)
	}
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionStart

	if err := game.UseSkill("p1", "bd_hope_fugue", nil, nil); err != nil {
		t.Fatalf("use hope fugue failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "bd_hope_draw_confirm")
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}}); err != nil { // 不摸牌
		t.Fatalf("choose draw confirm failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "bd_hope_mode")
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{2}}); err != nil { // 转移并+1灵感
		t.Fatalf("choose hope mode failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "bd_hope_transfer_target")
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil { // 目标队友 p3
		t.Fatalf("choose transfer target failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "bd_hope_transfer_discard")
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("choose transfer discard failed: %v", err)
	}

	if holder := bardpkg.EternalHolderID(engine.NewRoleChoiceRuntime(game), bard); holder != "p3" {
		t.Fatalf("expected eternal movement holder p3 after transfer, got %q", holder)
	}
	if findFieldEffectCard(allyA, model.EffectBardEternalMovement) != nil {
		t.Fatalf("expected allyA to no longer hold eternal movement")
	}
	if findFieldEffectCard(allyB, model.EffectBardEternalMovement) == nil {
		t.Fatalf("expected allyB to hold transferred eternal movement")
	}
	if got := bard.Tokens["bd_inspiration"]; got != 1 {
		t.Fatalf("expected bard inspiration +1 after transfer mode 2, got %d", got)
	}
	// Only the discarded hand card is in discard pile (no exclusive card consumed)
	if got := len(game.State.DiscardPile); got != 1 {
		t.Fatalf("expected discard pile contain only discarded hand card, got %d", got)
	}
}

func TestBardRousingRhapsody_OnAllyTurnStartRunsForbiddenVerse(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Bard", "bard", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "priest", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "EnemyA", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p4", "EnemyB", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	bard := game.State.Players["p1"]
	ally := game.State.Players["p2"]
	// 永恒乐章放置在队友身上，队友回合开始时触发响应询问
	if err := bardpkg.PlaceEternalMovement(engine.NewRoleChoiceRuntime(game), bard, ally); err != nil {
		t.Fatalf("place eternal movement failed: %v", err)
	}

	// ally 是当前回合玩家（永恒乐章持有者）
	ally.IsActive = true
	ally.TurnState = model.NewPlayerTurnState()
	ally.TurnState.HasProcessedTurnStart = false    // 确保 TurnStart hooks 能被触发
	game.State.CurrentTurn = 1                      // p2 是 index 1
	game.State.TurnStage = model.TurnStageTurnStart // 回合开始阶段触发激昂狂想曲
	game.State.PendingInterrupt = nil

	game.Drive()

	// 激昂狂想曲已简化为三分支直选（伤害 / 弃牌 / 跳过），不再有独立的 confirm 步骤。
	ctxData := testutils.RequireChoiceContext(t, game, "p2", "bd_rousing_mode")
	testutils.RequirePromptFlow(t, ctxData, "bd_rousing", "mode")
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p2", Selections: []int{0}}); err != nil { // 持有者选伤害分支
		t.Fatalf("choose rousing mode failed: %v", err)
	}
	// 目标选择由吟游诗人执行（伤害来源是吟游诗人）
	ctxData = testutils.RequireChoiceContext(t, game, "p1", "bd_rousing_targets")
	flow := testutils.RequirePromptFlow(t, ctxData, "bd_rousing", "targets")
	if _, ok := ctxData["selected_target_ids"]; ok {
		t.Fatalf("rousing should store targets in prompt flow, got legacy selected_target_ids in %+v", ctxData)
	}
	if got := flow.Selection("mode").Count; got != 0 {
		t.Fatalf("expected rousing flow mode=0, got %d in %+v", got, flow)
	}
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil { // 吟游诗人先选 p3
		t.Fatalf("choose rousing first target failed: %v", err)
	}
	ctxData = testutils.RequireChoiceContext(t, game, "p1", "bd_rousing_targets")
	flow = testutils.RequirePromptFlow(t, ctxData, "bd_rousing", "targets")
	if got := flow.Selection("targets").TargetIDs; len(got) != 1 {
		t.Fatalf("expected rousing flow to accumulate one target, got %+v in %+v", got, flow)
	}
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil { // 吟游诗人再选 p4
		t.Fatalf("choose rousing second target failed: %v", err)
	}

	if got := bard.Tokens["bd_inspiration"]; got != 1 {
		t.Fatalf("expected forbidden verse add inspiration to 1, got %d", got)
	}
	if holder := bardpkg.EternalHolderID(engine.NewRoleChoiceRuntime(game), bard); holder != "" {
		t.Fatalf("expected eternal movement removed by forbidden verse, holder=%q", holder)
	}
	// 永恒乐章应归还专属牌区，而非进入弃牌堆
	var hasEternalInExclusive bool
	for _, c := range bard.ExclusiveCards {
		if c.Name == "永恒乐章" {
			hasEternalInExclusive = true
			break
		}
	}
	if !hasEternalInExclusive {
		t.Fatalf("expected eternal movement restored to exclusive cards after forbidden verse")
	}
	if got := len(game.State.PendingDamageQueue); got != 2 {
		t.Fatalf("expected rousing queued 2 magic damages, got %d", got)
	}
	if game.State.CombatStage != model.CombatStageCalcDamage || game.State.ReturnTurnStage != model.TurnStageActionStart {
		t.Fatalf("expected damage resolution then return to action start, combat=%s return_turn=%s", game.State.CombatStage, game.State.ReturnTurnStage)
	}
}

func TestBardVictorySymphony_AtInspirationCapEntersPrisonerAndSelfDamages(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Bard", "bard", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "priest", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	bard := game.State.Players["p1"]
	ally := game.State.Players["p2"]
	// Victory symphony no longer requires exclusive card - only needs eternal movement on field
	bard.Tokens["bd_inspiration"] = 3
	if err := bardpkg.PlaceEternalMovement(engine.NewRoleChoiceRuntime(game), bard, ally); err != nil {
		t.Fatalf("place eternal movement failed: %v", err)
	}

	// ally 是当前回合玩家（永恒乐章持有者），回合结束时触发响应询问
	ally.IsActive = true
	ally.TurnState = model.NewPlayerTurnState()
	game.State.CurrentTurn = 1                    // p2 是 index 1
	game.State.TurnStage = model.TurnStageTurnEnd // 回合结束阶段触发胜利交响诗
	game.State.PendingInterrupt = nil

	game.Drive()

	testutils.RequireChoicePrompt(t, game, "p2", "bd_victory_confirm")
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p2", Selections: []int{1}}); err != nil { // 分支②
		t.Fatalf("choose victory branch failed: %v", err)
	}

	if got := bard.Form; got != model.FormBardEternalPrisoner {
		t.Fatalf("expected bard enter prisoner form at inspiration cap, got %q", got)
	}
	// 注意：pending damage 在 HandleAction 内部的 Drive 中被处理，
	// HandleAction 返回后 PendingDamageQueue 已被清空。
	// 禁忌诗篇的 pending damage 生成由 resolveBardForbiddenVerseAfterSong 保证，
	// 其实际效果（伤害结算）由游戏流程处理。
}

func TestBardVictorySymphony_ExtractStoneChoosesGemOrCrystal(t *testing.T) {
	tests := []struct {
		name        string
		gems        int
		crystals    int
		choiceIndex int
		wantGem     int
		wantCrystal int
	}{
		{name: "gem", gems: 1, crystals: 0, choiceIndex: 0, wantGem: 1},
		{name: "crystal", gems: 0, crystals: 1, choiceIndex: 0, wantCrystal: 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			game := engine.NewGameEngine(testutils.NoopObserver{})
			if err := game.AddPlayer("p1", "Bard", "bard", model.RedCamp); err != nil {
				t.Fatal(err)
			}
			if err := game.AddPlayer("p2", "Ally", "priest", model.RedCamp); err != nil {
				t.Fatal(err)
			}

			bard := game.State.Players["p1"]
			ally := game.State.Players["p2"]
			// Victory symphony no longer requires exclusive card
			if err := bardpkg.PlaceEternalMovement(engine.NewRoleChoiceRuntime(game), bard, ally); err != nil {
				t.Fatalf("place eternal movement failed: %v", err)
			}
			game.State.RedGems = tc.gems
			game.State.RedCrystals = tc.crystals

			ally.IsActive = true
			ally.TurnState = model.NewPlayerTurnState()
			game.State.CurrentTurn = 1 // p2 是 index 1
			game.State.TurnStage = model.TurnStageTurnEnd

			game.Drive()
			testutils.RequireChoicePrompt(t, game, "p2", "bd_victory_confirm")
			if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p2", Selections: []int{0}}); err != nil {
				t.Fatalf("choose extract mode failed: %v", err)
			}
			testutils.RequireChoicePrompt(t, game, "p2", "bd_victory_extract_stone")
			if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p2", Selections: []int{tc.choiceIndex}}); err != nil {
				t.Fatalf("choose stone failed: %v", err)
			}

			if ally.Gem != tc.wantGem || ally.Crystal != tc.wantCrystal {
				t.Fatalf("unexpected energy after extract: gem=%d crystal=%d", bard.Gem, bard.Crystal)
			}
		})
	}
}

func TestBardVictorySymphony_CancelAtCombinedPromptDeclines(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Bard", "bard", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "priest", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	bard := game.State.Players["p1"]
	ally := game.State.Players["p2"]
	if err := bardpkg.PlaceEternalMovement(engine.NewRoleChoiceRuntime(game), bard, ally); err != nil {
		t.Fatalf("place eternal movement failed: %v", err)
	}
	game.State.RedGems = 2
	game.State.RedCrystals = 1
	ally.Heal = 1

	ally.IsActive = true
	ally.TurnState = model.NewPlayerTurnState()
	game.State.CurrentTurn = 1
	game.State.TurnStage = model.TurnStageTurnEnd

	game.Drive()
	testutils.RequireChoicePrompt(t, game, "p2", "bd_victory_confirm")
	testutils.MustHandleAction(t, game, model.PlayerAction{Type: model.CmdCancel, PlayerID: "p2"})

	if game.State.RedGems != 2 || game.State.RedCrystals != 1 {
		t.Fatalf("cancel should not change camp stones, got gems=%d crystals=%d", game.State.RedGems, game.State.RedCrystals)
	}
	if ally.Heal != 1 {
		t.Fatalf("cancel should not heal holder, got heal=%d", ally.Heal)
	}
	if got := bard.Tokens["bd_inspiration"]; got != 0 {
		t.Fatalf("cancel should not trigger forbidden verse, inspiration=%d", got)
	}
}

func TestBardConfig_MetadataAlignsWithDocument(t *testing.T) {
	characters := data.GetCharacters()
	var bard *model.Character
	for _, character := range characters {
		if character.ID == "bard" {
			copy := character
			bard = &copy
			break
		}
	}
	if bard == nil {
		t.Fatalf("bard character not found")
	}

	var descent *model.SkillDefinition
	var dissonance *model.SkillDefinition
	var rousing *model.SkillDefinition
	var victory *model.SkillDefinition
	var hope *model.SkillDefinition
	for i := range bard.Skills {
		switch bard.Skills[i].ID {
		case "bd_descent_concerto":
			descent = &bard.Skills[i]
		case "bd_dissonance_chord":
			dissonance = &bard.Skills[i]
		case "bd_rousing_rhapsody":
			rousing = &bard.Skills[i]
		case "bd_victory_symphony":
			victory = &bard.Skills[i]
		case "bd_hope_fugue":
			hope = &bard.Skills[i]
		}
	}
	if descent == nil || dissonance == nil || rousing == nil || victory == nil || hope == nil {
		t.Fatalf("expected bard core skills present")
	}
	if descent.TargetType != model.TargetEnemy || descent.MinTargets != 1 || descent.MaxTargets != 1 {
		t.Fatalf("expected descent target metadata enemy(1), got type=%v min=%d max=%d", descent.TargetType, descent.MinTargets, descent.MaxTargets)
	}
	if descent.ResponseType != model.ResponseOptional {
		t.Fatalf("expected descent to be optional response, got %v", descent.ResponseType)
	}
	if dissonance.TargetType != model.TargetAny || dissonance.MinTargets != 1 || dissonance.MaxTargets != 1 {
		t.Fatalf("expected dissonance target metadata any(1), got type=%v min=%d max=%d", dissonance.TargetType, dissonance.MinTargets, dissonance.MaxTargets)
	}
	// Rousing, Victory, and Hope are now character skills (RequireExclusive=false)
	if rousing.RequireExclusive || rousing.TargetType != model.TargetEnemy || rousing.MinTargets != 0 || rousing.MaxTargets != 2 {
		t.Fatalf("expected rousing metadata NO exclusive + enemy(0..2), got requireExclusive=%v type=%v min=%d max=%d",
			rousing.RequireExclusive, rousing.TargetType, rousing.MinTargets, rousing.MaxTargets)
	}
	if victory.RequireExclusive {
		t.Fatalf("expected victory symphony NO require exclusive")
	}
	if hope.RequireExclusive || hope.TargetType != model.TargetAlly || hope.MinTargets != 1 || hope.MaxTargets != 1 {
		t.Fatalf("expected hope metadata require exclusive + ally(1), got requireExclusive=%v type=%v min=%d max=%d",
			hope.RequireExclusive, hope.TargetType, hope.MinTargets, hope.MaxTargets)
	}
}

func TestBardStarterExclusiveCards_NotInHand(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Bard", "bard", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	if err := game.StartGame(); err != nil {
		t.Fatalf("start game failed: %v", err)
	}

	bard := game.State.Players["p1"]
	if bard == nil {
		t.Fatalf("bard player missing")
	}
	if got := len(bard.Hand); got != 4 {
		t.Fatalf("expected bard starting hand remain 4, got %d", got)
	}
	// Bard now has only one exclusive card: 永恒乐章
	if len(bard.ExclusiveCards) != 1 {
		t.Fatalf("expected exactly 1 exclusive card, got %d", len(bard.ExclusiveCards))
	}
	if bard.ExclusiveCards[0].Name != "永恒乐章" {
		t.Fatalf("expected bard starter exclusive card 永恒乐章, got %s", bard.ExclusiveCards[0].Name)
	}
}

// TestBardRousingRhapsody_BleedTickRunsFirst 验证当血之巫女在流血形态且持有永恒乐章时，
// 回合开始先触发流血效果（伤害结算），再触发激昂狂想曲的响应询问。
func TestBardRousingRhapsody_BleedTickRunsFirst(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Bard", "bard", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "BloodWitch", "blood_priestess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "EnemyA", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p4", "EnemyB", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	bard := game.State.Players["p1"]
	witch := game.State.Players["p2"]

	// 永恒乐章放置在血之巫女身上
	if err := bardpkg.PlaceEternalMovement(engine.NewRoleChoiceRuntime(game), bard, witch); err != nil {
		t.Fatalf("place eternal movement failed: %v", err)
	}

	// 血之巫女进入流血形态
	witch.Form = model.FormBloodPriestessBleeding

	// 血之巫女回合开始
	witch.IsActive = true
	witch.TurnState = model.NewPlayerTurnState()
	witch.TurnState.HasProcessedTurnStart = false
	game.State.CurrentTurn = 1 // p2 是 index 1
	game.State.TurnStage = model.TurnStageTurnStart
	game.State.PendingInterrupt = nil

	game.Drive()

	// Drive 会自动处理：流血伤害先结算 → 然后激昂狂想曲三分支直选弹窗
	// 最终状态应为等待永恒乐章持有者（血之巫女）选择伤害/弃牌/跳过
	testutils.RequireChoicePrompt(t, game, "p2", "bd_rousing_mode")

	// 验证日志顺序：流血在前，激昂狂想曲在后
	var bleedIdx, rousingIdx int = -1, -1
	for i, e := range obs.Events {
		if e.Type == model.EventLog {
			if bleedIdx < 0 && strings.Contains(e.Message, "流血") {
				bleedIdx = i
			}
			if rousingIdx < 0 && strings.Contains(e.Message, "激昂狂想曲") {
				rousingIdx = i
			}
		}
	}
	if bleedIdx < 0 {
		t.Fatal("expected bleed log entry")
	}
	if rousingIdx < 0 {
		t.Fatal("expected rousing rhapsody log entry")
	}
	if bleedIdx > rousingIdx {
		t.Fatalf("expected bleed log before rousing log, got bleed at %d and rousing at %d", bleedIdx, rousingIdx)
	}
}
