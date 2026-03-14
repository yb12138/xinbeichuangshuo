package engine

import (
	"testing"

	"starcup-engine/internal/engine/skills"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func bloodPriestessTestCard(id string, ele model.Element) model.Card {
	return model.Card{
		ID:          id,
		Name:        "测试牌",
		Type:        model.CardTypeAttack,
		Element:     ele,
		Faction:     "血",
		Damage:      2,
		Description: "test",
	}
}

func TestBloodPriestessSharedLife_DrawBeforePlaceOverflowThenApply(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Witch", "blood_priestess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["bp_bleed_form"] = 1
	p1.Hand = []model.Card{
		bloodPriestessTestCard("h1", model.ElementFire),
		bloodPriestessTestCard("h2", model.ElementWater),
		bloodPriestessTestCard("h3", model.ElementWind),
		bloodPriestessTestCard("h4", model.ElementThunder),
		bloodPriestessTestCard("h5", model.ElementEarth),
		bloodPriestessTestCard("h6", model.ElementDark),
	}
	p1.ExclusiveCards = append(p1.ExclusiveCards, makeStarterBloodSharedLifeCard(p1))
	game.State.Deck = rules.InitDeck()

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdSkill,
		SkillID:  "bp_shared_life",
	})
	requireChoicePrompt(t, game, "p1", "bp_shared_life_target")
	if err := game.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("choose shared-life target failed: %v", err)
	}

	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptDiscard {
		t.Fatalf("expected overflow discard interrupt before placing shared life")
	}
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{6, 7},
	})

	holder, fc := game.findBloodPriestessSharedLife(p1)
	if holder == nil || fc == nil {
		t.Fatalf("expected shared life effect placed after overflow resolution")
	}
	if holder.ID != "p1" {
		t.Fatalf("expected shared life holder p1, got %s", holder.ID)
	}
	if got := game.State.RedMorale; got != 13 {
		t.Fatalf("expected morale down by 2 before shared-life placement, got %d", got)
	}
	if got := len(p1.Hand); got != 6 {
		t.Fatalf("expected final hand count 6, got %d", got)
	}
	if got := game.GetMaxHand(p1); got != 7 {
		t.Fatalf("expected bleed-form shared-life max hand 7, got %d", got)
	}
}

func TestBloodPriestessBleeding_EnterOnMoraleLossAndAutoReleaseOnLowHand(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Witch", "blood_priestess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	p1.MaxHand = 6
	p1.Hand = []model.Card{
		bloodPriestessTestCard("a1", model.ElementFire),
		bloodPriestessTestCard("a2", model.ElementWater),
		bloodPriestessTestCard("a3", model.ElementWind),
		bloodPriestessTestCard("a4", model.ElementThunder),
		bloodPriestessTestCard("a5", model.ElementEarth),
		bloodPriestessTestCard("a6", model.ElementDark),
		bloodPriestessTestCard("a7", model.ElementLight),
		bloodPriestessTestCard("a8", model.ElementFire),
	}
	damageOverflowCtx := game.buildContext(p1, nil, model.TriggerNone, nil)
	damageOverflowCtx.Flags["FromDamageDraw"] = true
	game.checkHandLimit(p1, damageOverflowCtx)
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptDiscard {
		t.Fatalf("expected discard interrupt")
	}
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{6, 7},
	})

	if got := p1.Tokens["bp_bleed_form"]; got != 1 {
		t.Fatalf("expected enter bleed form, got %d", got)
	}
	if got := p1.Heal; got != 1 {
		t.Fatalf("expected heal +1 on entering bleed form, got %d", got)
	}

	// 手动降到2张，验证“手牌<3立即脱离流血形态”。
	p1.Hand = p1.Hand[:2]
	_ = game.GetMaxHand(p1) // GetMaxHand 内会触发强制重置逻辑
	if got := p1.Tokens["bp_bleed_form"]; got != 0 {
		t.Fatalf("expected auto release from bleed form at hand<3, got %d", got)
	}
}

func TestBloodPriestessBleeding_TurnStartSelfDamageBeforeBuff(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Witch", "blood_priestess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Heal = 0
	p1.Tokens["bp_bleed_form"] = 1
	p1.Tokens["bp_bleed_tick_done_turn"] = 0
	game.State.Deck = rules.InitDeck()
	p1.Hand = []model.Card{
		bloodPriestessTestCard("s1", model.ElementFire),
		bloodPriestessTestCard("s2", model.ElementWater),
		bloodPriestessTestCard("s3", model.ElementWind),
	}

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseBuffResolve
	game.Drive()

	if got := p1.Tokens["bp_bleed_tick_done_turn"]; got != 1 {
		t.Fatalf("expected bleed tick consumed at turn start, got %d", got)
	}
	// 承伤摸1：3 -> 4
	if got := len(p1.Hand); got != 4 {
		t.Fatalf("expected turn-start self-damage draw 1 card, hand=4 got %d", got)
	}
}

func TestBloodPriestessBloodSorrow_TransferThenRemove(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Witch", "blood_priestess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		bloodPriestessTestCard("c1", model.ElementFire),
		bloodPriestessTestCard("c2", model.ElementWater),
	}
	p1.ExclusiveCards = append(p1.ExclusiveCards, makeStarterBloodSharedLifeCard(p1))
	game.State.Deck = rules.InitDeck()
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	// 1) 先放置同生共死到 p2。
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdSkill,
		SkillID:  "bp_shared_life",
	})
	requireChoicePrompt(t, game, "p1", "bp_shared_life_target")
	// 目标列表顺序按 PlayerOrder: p1,p2,p3；这里选 p2(index=1)。
	if err := game.handleWeakChoiceInput("p1", 1); err != nil {
		t.Fatalf("choose shared-life target p2 failed: %v", err)
	}
	game.Drive() // 触发 deferred 放置同生共死
	holder, _ := game.findBloodPriestessSharedLife(p1)
	if holder == nil || holder.ID != "p2" {
		t.Fatalf("expected shared life holder p2 before blood sorrow, got %+v", holder)
	}
	// 让上限足够高，避免血之哀伤自伤摸牌触发爆牌弃牌中断，聚焦转移/移除逻辑本身。
	p1.Tokens["bp_bleed_form"] = 1

	// 2) 启动血之哀伤，选择“转移”到 p3。
	game.State.CurrentTurn = 0
	p1.IsActive = true
	ctx := game.buildContext(p1, nil, model.TriggerOnTurnStart, nil)
	h := &skills.BloodPriestessBloodSorrowHandler{}
	if !h.CanUse(ctx) {
		t.Fatalf("expected blood sorrow can use when shared life exists")
	}
	if err := h.Execute(ctx); err != nil {
		t.Fatalf("execute blood sorrow failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "bp_blood_sorrow_mode")
	if err := game.handleWeakChoiceInput("p1", 0); err != nil { // 转移分支
		t.Fatalf("choose blood sorrow transfer mode failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "bp_blood_sorrow_target")
	if err := game.handleWeakChoiceInput("p1", 2); err != nil { // 选 p3
		t.Fatalf("choose blood sorrow transfer target p3 failed: %v", err)
	}
	holder, _ = game.findBloodPriestessSharedLife(p1)
	if holder == nil || holder.ID != "p3" {
		t.Fatalf("expected shared life holder p3 after transfer, got %+v", holder)
	}

	// 3) 再次发动血之哀伤，选择“移除”。
	game.State.CurrentTurn = 0
	p1.IsActive = true
	ctx = game.buildContext(p1, nil, model.TriggerOnTurnStart, nil)
	if !h.CanUse(ctx) {
		t.Fatalf("expected blood sorrow can use before remove branch")
	}
	if err := h.Execute(ctx); err != nil {
		t.Fatalf("execute blood sorrow(remove) failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "bp_blood_sorrow_mode")
	if err := game.handleWeakChoiceInput("p1", 1); err != nil { // 移除分支
		t.Fatalf("choose blood sorrow remove mode failed: %v", err)
	}
	holder, fc := game.findBloodPriestessSharedLife(p1)
	if holder != nil || fc != nil {
		t.Fatalf("expected shared life removed, holder=%+v card=%+v", holder, fc)
	}
	if !p1.HasExclusiveCard(p1.Character.Name, "同生共死") {
		t.Fatalf("expected shared-life card restored to exclusive zone after remove branch")
	}
}

func TestBloodPriestessSharedLife_FixedHandCapTargetExempt(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Witch", "blood_priestess", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Lancer", "magic_lancer", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p2.Tokens["ml_phantom_form"] = 1 // 恒定手牌上限=5
	p1.Hand = []model.Card{
		bloodPriestessTestCard("f1", model.ElementFire),
		bloodPriestessTestCard("f2", model.ElementWater),
		bloodPriestessTestCard("f3", model.ElementWind),
	}

	if err := game.placeBloodPriestessSharedLife(p1, p2, bloodPriestessSharedLifeCard(p1)); err != nil {
		t.Fatalf("place shared life failed: %v", err)
	}

	// 普通形态：同生共死对固定上限目标不生效，对血之巫女自身照常生效。
	p1.Tokens["bp_bleed_form"] = 0
	if got := game.GetMaxHand(p1); got != 4 {
		t.Fatalf("expected priestess max hand 4 in normal form with shared life, got %d", got)
	}
	if got := game.GetMaxHand(p2); got != 5 {
		t.Fatalf("expected fixed-cap target keep max hand 5, got %d", got)
	}

	// 流血形态：自身改为+1；目标仍应保持固定上限不变。
	p1.Tokens["bp_bleed_form"] = 1
	if got := game.GetMaxHand(p1); got != 7 {
		t.Fatalf("expected priestess max hand 7 in bleed form with shared life, got %d", got)
	}
	if got := game.GetMaxHand(p2); got != 5 {
		t.Fatalf("expected fixed-cap target still 5 in bleed form, got %d", got)
	}
}
