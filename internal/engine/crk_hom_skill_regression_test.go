package engine

import (
	"strings"
	"testing"

	"starcup-engine/internal/engine/skills"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func choiceTypeOfInterrupt(intr *model.Interrupt) string {
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

func promptHasOptionID(prompt *model.Prompt, optionID string) bool {
	if prompt == nil {
		return false
	}
	for _, opt := range prompt.Options {
		if opt.ID == optionID {
			return true
		}
	}
	return false
}

func makeHandCards(n int, element model.Element) []model.Card {
	out := make([]model.Card, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, model.Card{
			ID:      string(rune('a' + i)),
			Name:    "测试牌",
			Type:    model.CardTypeAttack,
			Element: element,
			Damage:  2,
			Faction: "幻",
		})
	}
	return out
}

func TestCrimsonKnightCalmMind_AllowsActionTypeChoice(t *testing.T) {
	g := NewGameEngine(noopObserver{})
	if err := g.AddPlayer("p1", "Crimson", "crimson_knight", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["crk_hot_form"] = 1
	p1.Crystal = 1

	h := skills.GetHandler("crk_calm_mind")
	if h == nil {
		t.Fatalf("crk_calm_mind handler not found")
	}
	ctx := g.buildContext(p1, nil, model.TriggerOnPhaseEnd, &model.EventContext{
		Type:       model.EventPhaseEnd,
		SourceID:   p1.ID,
		ActionType: model.ActionMagic, // 法术行动结束后，仍应允许选择“攻击行动”
	})
	if !h.CanUse(ctx) {
		t.Fatalf("expected calm mind can use in hot form after action end")
	}
	if err := h.Execute(ctx); err != nil {
		t.Fatalf("execute calm mind failed: %v", err)
	}
	if got := p1.Tokens["crk_hot_form"]; got != 0 {
		t.Fatalf("expected hot form reset to 0, got %d", got)
	}
	if g.State.PendingInterrupt == nil || choiceTypeOfInterrupt(g.State.PendingInterrupt) != "crk_calm_mind_action" {
		t.Fatalf("expected crk_calm_mind_action choice, got %+v", g.State.PendingInterrupt)
	}

	// 选择“额外攻击行动”
	if err := g.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("choose calm mind action failed: %v", err)
	}
	if len(p1.TurnState.PendingActions) == 0 {
		t.Fatalf("expected one pending action from calm mind")
	}
	last := p1.TurnState.PendingActions[len(p1.TurnState.PendingActions)-1]
	if last.MustType != "Attack" {
		t.Fatalf("expected calm mind chosen attack action, got %+v", last)
	}
}

func TestCrimsonKnightHotBlood_AutoReleaseOnTurnEnd(t *testing.T) {
	g := NewGameEngine(noopObserver{})
	if err := g.AddPlayer("p1", "Crimson", "crimson_knight", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p1.IsActive = true
	p1.Heal = 0
	p1.Tokens["crk_hot_form"] = 1
	g.State.CurrentTurn = 0
	g.State.Phase = model.PhaseTurnEnd

	g.Drive()

	if got := p1.Tokens["crk_hot_form"]; got != 0 {
		t.Fatalf("expected hot form reset to 0 at turn end, got %d", got)
	}
	if got := p1.Heal; got != 2 {
		t.Fatalf("expected heal +2 at turn end, got %d", got)
	}
}

func TestCrimsonKnightHotBlood_NextTurnFallbackStillReleases(t *testing.T) {
	g := NewGameEngine(noopObserver{})
	if err := g.AddPlayer("p1", "Crimson", "crimson_knight", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p1.IsActive = true
	p1.Heal = 0
	p1.Tokens["crk_hot_form"] = 1
	g.State.CurrentTurn = 0

	// 模拟“跳过 PhaseTurnEnd 直接调用 NextTurn”的路径，仍应触发回合结束退形态。
	g.NextTurn()

	if got := p1.Tokens["crk_hot_form"]; got != 0 {
		t.Fatalf("expected hot form reset to 0 in NextTurn fallback, got %d", got)
	}
	if got := p1.Heal; got != 2 {
		t.Fatalf("expected heal +2 in NextTurn fallback, got %d", got)
	}
}

func TestCrimsonKnightHotForm_DamageOverflowNoMoraleLoss(t *testing.T) {
	g := NewGameEngine(noopObserver{})
	if err := g.AddPlayer("p1", "Attacker", "priest", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Crimson", "crimson_knight", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	g.State.CurrentTurn = 0
	g.State.Phase = model.PhaseActionSelection

	p1 := g.State.Players["p1"]
	p2 := g.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p2.Tokens["crk_hot_form"] = 1

	p1.Hand = []model.Card{
		{ID: "atk1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2, Faction: "咏"},
	}
	// 受2点伤害后摸2张，超上限2张，进入爆牌弃牌流程。
	p2.Hand = []model.Card{
		{ID: "h1", Name: "牌1", Type: model.CardTypeAttack, Element: model.ElementFire},
		{ID: "h2", Name: "牌2", Type: model.CardTypeAttack, Element: model.ElementWater},
		{ID: "h3", Name: "牌3", Type: model.CardTypeAttack, Element: model.ElementWind},
		{ID: "h4", Name: "牌4", Type: model.CardTypeAttack, Element: model.ElementThunder},
		{ID: "h5", Name: "牌5", Type: model.CardTypeMagic, Element: model.ElementDark},
		{ID: "h6", Name: "牌6", Type: model.CardTypeMagic, Element: model.ElementLight},
	}
	blueMoraleBefore := g.State.BlueMorale

	mustHandleAction(t, g, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})
	mustHandleAction(t, g, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	})
	if g.State.PendingInterrupt == nil || g.State.PendingInterrupt.Type != model.InterruptDiscard {
		t.Fatalf("expected overflow discard interrupt, got %+v", g.State.PendingInterrupt)
	}
	data, _ := g.State.PendingInterrupt.Context.(map[string]interface{})
	discardCount, _ := data["discard_count"].(int)
	if discardCount <= 0 {
		t.Fatalf("expected discard_count > 0, got %v", data["discard_count"])
	}
	selections := make([]int, 0, discardCount)
	for i := 0; i < discardCount; i++ {
		selections = append(selections, i)
	}
	mustHandleAction(t, g, model.PlayerAction{
		PlayerID:   "p2",
		Type:       model.CmdSelect,
		Selections: selections,
	})

	if got := g.State.BlueMorale; got != blueMoraleBefore {
		t.Fatalf("expected no morale loss in hot form damage overflow, before=%d after=%d", blueMoraleBefore, got)
	}
}

func TestHomRuneReforge_ReallocateAndOverflowCheckOnTurnEnd(t *testing.T) {
	g := NewGameEngine(noopObserver{})
	if err := g.AddPlayer("p1", "Hom", "war_homunculus", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	g.State.Deck = rules.InitDeck()

	p1 := g.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 1
	p1.Tokens["hom_war_rune"] = 3
	p1.Tokens["hom_magic_rune"] = 0
	p1.Tokens["hom_burst_form"] = 0
	// 进入形态前 6 张手牌，符文改造摸1后=7（形态内上限+1），回合结束转正后应触发弃1。
	p1.Hand = makeHandCards(6, model.ElementFire)

	h := skills.GetHandler("hom_rune_reforge")
	if h == nil {
		t.Fatalf("hom_rune_reforge handler not found")
	}
	ctx := g.buildContext(p1, nil, model.TriggerOnTurnStart, &model.EventContext{
		SourceID: p1.ID,
	})
	if !h.CanUse(ctx) {
		t.Fatalf("expected rune reforge can use with 1 gem and non-burst form")
	}
	if err := h.Execute(ctx); err != nil {
		t.Fatalf("execute rune reforge failed: %v", err)
	}
	if got := p1.Gem; got != 0 {
		t.Fatalf("expected gem consumed to 0, got %d", got)
	}
	if got := p1.Tokens["hom_burst_form"]; got != 1 {
		t.Fatalf("expected burst form entered, got %d", got)
	}
	if g.State.PendingInterrupt == nil || choiceTypeOfInterrupt(g.State.PendingInterrupt) != "hom_rune_reforge_distribution" {
		t.Fatalf("expected hom_rune_reforge_distribution choice, got %+v", g.State.PendingInterrupt)
	}

	// 选择 战纹2 / 魔纹1
	if err := g.handleWeakChoiceInput("p1", 2); err != nil {
		t.Fatalf("choose rune distribution failed: %v", err)
	}
	if p1.Tokens["hom_war_rune"] != 2 || p1.Tokens["hom_magic_rune"] != 1 {
		t.Fatalf("unexpected rune distribution: war=%d magic=%d", p1.Tokens["hom_war_rune"], p1.Tokens["hom_magic_rune"])
	}

	g.State.CurrentTurn = 0
	g.State.Phase = model.PhaseTurnEnd
	g.Drive()

	if got := p1.Tokens["hom_burst_form"]; got != 0 {
		t.Fatalf("expected burst form cleared at turn end, got %d", got)
	}
	if g.State.PendingInterrupt == nil || g.State.PendingInterrupt.Type != model.InterruptDiscard {
		t.Fatalf("expected discard interrupt after form ends overflow, got %+v", g.State.PendingInterrupt)
	}
	data, _ := g.State.PendingInterrupt.Context.(map[string]interface{})
	if dc, _ := data["discard_count"].(int); dc != 1 {
		t.Fatalf("expected discard_count=1 after form ends, got %v", data["discard_count"])
	}
}

func TestHomGlyphFusion_MaxXUsesDistinctElements(t *testing.T) {
	g := NewGameEngine(noopObserver{})
	if err := g.AddPlayer("p1", "Hom", "war_homunculus", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["hom_magic_rune"] = 2
	p1.Hand = []model.Card{
		{ID: "h1", Name: "水涟斩", Type: model.CardTypeAttack, Element: model.ElementWater, Damage: 2},
		{ID: "h2", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 2},
		{ID: "h3", Name: "风神斩", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 2},
	}
	h := skills.GetHandler("hom_glyph_fusion")
	if h == nil {
		t.Fatalf("hom_glyph_fusion handler not found")
	}
	damageVal := 2
	ctx := g.buildContext(p1, g.State.Players["p2"], model.TriggerOnAttackMiss, &model.EventContext{
		Type:      model.EventAttack,
		SourceID:  p1.ID,
		TargetID:  "p2",
		DamageVal: &damageVal,
		Card: &model.Card{
			ID:      "atk",
			Name:    "火焰斩",
			Type:    model.CardTypeAttack,
			Element: model.ElementFire,
			Damage:  2,
		},
		AttackInfo: &model.AttackEventInfo{ActionType: "Attack"},
	})

	if !h.CanUse(ctx) {
		t.Fatalf("expected glyph fusion can use with 2 distinct off-elements")
	}
	if err := h.Execute(ctx); err != nil {
		t.Fatalf("execute glyph fusion failed: %v", err)
	}
	data, _ := g.State.PendingInterrupt.Context.(map[string]interface{})
	if maxX, _ := data["max_x"].(int); maxX != 2 {
		t.Fatalf("expected max_x=2 by distinct elements, got %v", data["max_x"])
	}

	// 选择 X=2（通过选项值）
	if err := g.handleWeakChoiceInput("p1", 2); err != nil {
		t.Fatalf("choose glyph x failed: %v", err)
	}
	// 先选一张水系牌（索引0）
	if err := g.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("choose first glyph card failed: %v", err)
	}
	// 剩余候选中不应再有另一张水系（索引1），只保留风系（索引2）
	data, _ = g.State.PendingInterrupt.Context.(map[string]interface{})
	remaining, _ := data["remaining_indices"].([]int)
	if len(remaining) != 1 || remaining[0] != 2 {
		t.Fatalf("expected remaining only index 2 after distinct filter, got %+v", remaining)
	}
}

func TestHomDualEcho_TargetChoiceCanCancel(t *testing.T) {
	g := NewGameEngine(noopObserver{})
	if err := g.AddPlayer("p1", "Hom", "war_homunculus", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1

	h := skills.GetHandler("hom_dual_echo")
	if h == nil {
		t.Fatalf("hom_dual_echo handler not found")
	}

	damageVal := 2
	ctx := g.buildContext(p1, p1, model.TriggerOnDamageTaken, &model.EventContext{
		Type:      model.EventDamage,
		SourceID:  p1.ID,
		TargetID:  p1.ID,
		DamageVal: &damageVal,
	})
	if !h.CanUse(ctx) {
		t.Fatalf("expected dual echo can use")
	}
	if err := h.Execute(ctx); err != nil {
		t.Fatalf("execute dual echo failed: %v", err)
	}
	if got := p1.Crystal; got != 1 {
		t.Fatalf("cost should not be consumed before target selection, got crystal=%d", got)
	}
	if g.State.PendingInterrupt == nil || choiceTypeOfInterrupt(g.State.PendingInterrupt) != "hom_dual_echo_target" {
		t.Fatalf("expected hom_dual_echo_target interrupt, got %+v", g.State.PendingInterrupt)
	}

	prompt := g.buildChoicePrompt()
	if !promptHasOptionID(prompt, "cancel") {
		t.Fatalf("expected cancel option in hom_dual_echo_target prompt, got %+v", prompt)
	}

	if err := g.handleInterruptAction(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdCancel,
	}); err != nil {
		t.Fatalf("cancel dual echo target choice failed: %v", err)
	}

	if g.State.PendingInterrupt != nil {
		t.Fatalf("expected interrupt cleared after cancel, got %+v", g.State.PendingInterrupt)
	}
	if got := p1.Crystal; got != 1 {
		t.Fatalf("cancel should not consume crystal, got %d", got)
	}
	if len(g.State.PendingDamageQueue) != 0 {
		t.Fatalf("cancel should not enqueue extra damage, got %+v", g.State.PendingDamageQueue)
	}
}

func TestHomDualEcho_TargetConfirmConsumesCostAndQueuesDamage(t *testing.T) {
	g := NewGameEngine(noopObserver{})
	if err := g.AddPlayer("p1", "Hom", "war_homunculus", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1

	h := skills.GetHandler("hom_dual_echo")
	if h == nil {
		t.Fatalf("hom_dual_echo handler not found")
	}
	damageVal := 2
	ctx := g.buildContext(p1, p1, model.TriggerOnDamageTaken, &model.EventContext{
		Type:      model.EventDamage,
		SourceID:  p1.ID,
		TargetID:  p1.ID,
		DamageVal: &damageVal,
	})
	if !h.CanUse(ctx) {
		t.Fatalf("expected dual echo can use")
	}
	if err := h.Execute(ctx); err != nil {
		t.Fatalf("execute dual echo failed: %v", err)
	}
	if err := g.handleInterruptAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("confirm dual echo target failed: %v", err)
	}

	if got := p1.Crystal; got != 0 {
		t.Fatalf("expected crystal consumed on confirm, got %d", got)
	}
	if len(g.State.PendingDamageQueue) != 1 {
		t.Fatalf("expected one pending damage after confirm, got %d", len(g.State.PendingDamageQueue))
	}
	pd := g.State.PendingDamageQueue[0]
	if pd.SourceID != "p1" || pd.TargetID != "p2" || pd.Damage != 2 || pd.DamageType != "magic" {
		t.Fatalf("unexpected pending damage %+v", pd)
	}
	if !pd.CapDrawToHandLimit {
		t.Fatalf("expected dual echo pending damage to cap draw to hand limit")
	}
}

func TestHomDualEcho_DamageDrawCapsAtHandLimit(t *testing.T) {
	g := NewGameEngine(noopObserver{})
	if err := g.AddPlayer("p1", "Hom", "war_homunculus", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	g.State.Deck = rules.InitDeck()

	p1 := g.State.Players["p1"]
	p2 := g.State.Players["p2"]
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1
	p2.Heal = 0
	p2.Hand = makeHandCards(5, model.ElementFire) // 默认上限6，仅剩1手牌空间

	h := skills.GetHandler("hom_dual_echo")
	if h == nil {
		t.Fatalf("hom_dual_echo handler not found")
	}
	damageVal := 2
	ctx := g.buildContext(p1, p1, model.TriggerOnDamageTaken, &model.EventContext{
		Type:      model.EventDamage,
		SourceID:  p1.ID,
		TargetID:  p1.ID,
		DamageVal: &damageVal,
	})
	if !h.CanUse(ctx) {
		t.Fatalf("expected dual echo can use")
	}
	if err := h.Execute(ctx); err != nil {
		t.Fatalf("execute dual echo failed: %v", err)
	}
	if err := g.handleInterruptAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("confirm dual echo target failed: %v", err)
	}
	g.State.Phase = model.PhasePendingDamageResolution
	for i := 0; i < 8 && len(g.State.PendingDamageQueue) > 0; i++ {
		if paused := g.processPendingDamages(); paused {
			t.Fatalf("unexpected interrupt while resolving dual echo damage: %+v", g.State.PendingInterrupt)
		}
	}
	if len(g.State.PendingDamageQueue) != 0 {
		t.Fatalf("pending damage queue not drained, len=%d", len(g.State.PendingDamageQueue))
	}
	if got := len(p2.Hand); got != g.GetMaxHand(p2) {
		t.Fatalf("expected dual echo draw capped at max hand=%d, got hand=%d", g.GetMaxHand(p2), got)
	}
}

func TestCrimsonKnightFaith_OnlyWhitelistedSelfDamageCanUseHeal(t *testing.T) {
	g := NewGameEngine(noopObserver{})
	if err := g.AddPlayer("p1", "Crimson", "crimson_knight", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	g.State.Deck = rules.InitDeck()
	g.State.Phase = model.PhasePendingDamageResolution

	p1 := g.State.Players["p1"]
	p1.TurnState = model.NewPlayerTurnState()
	p1.Heal = 2
	p1.Hand = nil

	// 非白名单自伤：不应弹治疗抵御。
	g.State.PendingDamageQueue = []model.PendingDamage{
		{
			SourceID:   p1.ID,
			TargetID:   p1.ID,
			Damage:     1,
			DamageType: "magic",
			Stage:      0,
		},
	}
	for i := 0; i < 5 && len(g.State.PendingDamageQueue) > 0; i++ {
		if paused := g.processPendingDamages(); paused {
			t.Fatalf("non-whitelisted self damage should not open heal choice, got %+v", g.State.PendingInterrupt)
		}
	}
	if got := len(p1.Hand); got != 1 {
		t.Fatalf("expected self damage draws 1 card, got %d", got)
	}
	if got := p1.Heal; got != 2 {
		t.Fatalf("expected heal unchanged for non-whitelisted self damage, got %d", got)
	}

	// 白名单自伤：应弹治疗抵御。
	g.State.PendingDamageQueue = []model.PendingDamage{
		{
			SourceID:              p1.ID,
			TargetID:              p1.ID,
			Damage:                1,
			DamageType:            "magic",
			AllowCrimsonFaithHeal: true,
			Stage:                 0,
		},
	}
	g.State.PendingInterrupt = nil
	if paused := g.processPendingDamages(); !paused {
		t.Fatalf("expected heal choice interrupt for whitelisted self damage")
	}
	if g.State.PendingInterrupt == nil || choiceTypeOfInterrupt(g.State.PendingInterrupt) != "heal" {
		t.Fatalf("expected heal choice interrupt, got %+v", g.State.PendingInterrupt)
	}
	if g.State.PendingInterrupt.PlayerID != p1.ID {
		t.Fatalf("expected heal choice for p1, got %s", g.State.PendingInterrupt.PlayerID)
	}
}

func TestCrimsonKnightFaith_SelfPoisonCanUseHeal(t *testing.T) {
	g := NewGameEngine(noopObserver{})
	if err := g.AddPlayer("p1", "Crimson", "crimson_knight", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	g.State.Phase = model.PhasePendingDamageResolution

	p1 := g.State.Players["p1"]
	p1.TurnState = model.NewPlayerTurnState()
	p1.Heal = 1
	p1.AddFieldCard(&model.FieldCard{
		Card: model.Card{
			ID:      "poison-self",
			Name:    "中毒",
			Type:    model.CardTypeMagic,
			Element: model.ElementEarth,
		},
		OwnerID:  p1.ID,
		SourceID: p1.ID,
		Mode:     model.FieldEffect,
		Effect:   model.EffectPoison,
		Trigger:  model.EffectTriggerOnTurnStart,
	})

	g.triggerFieldEffects(p1, model.EffectTriggerOnTurnStart, nil)
	if len(g.State.PendingDamageQueue) != 1 {
		t.Fatalf("expected one poison pending damage, got %d", len(g.State.PendingDamageQueue))
	}
	pd := g.State.PendingDamageQueue[0]
	if !pd.AllowCrimsonFaithHeal {
		t.Fatalf("expected self-poison damage to allow crimson faith heal")
	}
	if pd.SourceID != p1.ID || pd.TargetID != p1.ID {
		t.Fatalf("unexpected poison pending damage %+v", pd)
	}

	if paused := g.processPendingDamages(); !paused {
		t.Fatalf("expected heal choice interrupt for self-poison")
	}
	if g.State.PendingInterrupt == nil || choiceTypeOfInterrupt(g.State.PendingInterrupt) != "heal" {
		t.Fatalf("expected heal choice interrupt after self-poison, got %+v", g.State.PendingInterrupt)
	}
}

func TestCrimsonKnightBloodyPrayerXPrompt_NoZeroOption(t *testing.T) {
	g := NewGameEngine(noopObserver{})
	if err := g.AddPlayer("p1", "Crimson", "crimson_knight", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Ally", "priest", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p1.TurnState = model.NewPlayerTurnState()
	p1.Heal = 2

	h := skills.GetHandler("crk_bloody_prayer")
	if h == nil {
		t.Fatalf("crk_bloody_prayer handler not found")
	}
	ctx := g.buildContext(p1, nil, model.TriggerNone, nil)
	if !h.CanUse(ctx) {
		t.Fatalf("expected bloody prayer can use with heal>0 and ally exists")
	}
	if err := h.Execute(ctx); err != nil {
		t.Fatalf("execute bloody prayer failed: %v", err)
	}

	prompt := g.GetCurrentPrompt()
	if prompt == nil {
		t.Fatalf("expected bloody prayer x prompt")
	}
	for _, opt := range prompt.Options {
		if strings.Contains(opt.Label, "X=0") {
			t.Fatalf("bloody prayer prompt should not contain X=0 option, got %+v", prompt.Options)
		}
	}
	if len(prompt.Options) != 2 {
		t.Fatalf("expected options X=1..2, got %d", len(prompt.Options))
	}
}

func TestHomRuneSmash_BurstAddsAttackAndMagicDamage(t *testing.T) {
	g := NewGameEngine(noopObserver{})
	if err := g.AddPlayer("p1", "Hom", "war_homunculus", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p2 := g.State.Players["p2"]
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p1.Tokens["hom_war_rune"] = 3
	p1.Tokens["hom_magic_rune"] = 0
	p1.Tokens["hom_burst_form"] = 1
	p1.Hand = []model.Card{
		{ID: "f1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
		{ID: "f2", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
	}

	damageVal := 2
	h := skills.GetHandler("hom_rune_smash")
	if h == nil {
		t.Fatalf("hom_rune_smash handler not found")
	}
	ctx := g.buildContext(p1, p2, model.TriggerOnAttackHit, &model.EventContext{
		Type:      model.EventAttack,
		SourceID:  p1.ID,
		TargetID:  p2.ID,
		DamageVal: &damageVal,
		Card: &model.Card{
			ID:      "atk",
			Name:    "火焰斩",
			Type:    model.CardTypeAttack,
			Element: model.ElementFire,
			Damage:  2,
		},
		AttackInfo: &model.AttackEventInfo{ActionType: "Attack", IsHit: true},
	})
	if !h.CanUse(ctx) {
		t.Fatalf("expected rune smash can use")
	}
	if err := h.Execute(ctx); err != nil {
		t.Fatalf("execute rune smash failed: %v", err)
	}
	if g.State.PendingInterrupt == nil || choiceTypeOfInterrupt(g.State.PendingInterrupt) != "hom_rune_smash_x" {
		t.Fatalf("expected hom_rune_smash_x choice, got %+v", g.State.PendingInterrupt)
	}

	// X=2，弃2张同系牌
	if err := g.handleWeakChoiceInput("p1", 2); err != nil {
		t.Fatalf("choose rune smash x failed: %v", err)
	}
	if err := g.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("choose first card failed: %v", err)
	}
	if err := g.handleWeakChoiceInput("p1", 1); err != nil {
		t.Fatalf("choose second card failed: %v", err)
	}
	// Y=1：额外翻转1战纹并造成1点法伤
	if err := g.handleWeakChoiceInput("p1", 1); err != nil {
		t.Fatalf("choose rune smash y failed: %v", err)
	}

	if damageVal != 3 {
		t.Fatalf("expected attack damage +1 (X-1), got %d", damageVal)
	}
	if p1.Tokens["hom_war_rune"] != 1 || p1.Tokens["hom_magic_rune"] != 2 {
		t.Fatalf("unexpected rune flip result war=%d magic=%d", p1.Tokens["hom_war_rune"], p1.Tokens["hom_magic_rune"])
	}
	if len(g.State.PendingDamageQueue) == 0 {
		t.Fatalf("expected pending magic damage from Y")
	}
	pd := g.State.PendingDamageQueue[0]
	if pd.TargetID != p2.ID || pd.Damage != 1 || pd.DamageType != "magic" {
		t.Fatalf("unexpected rune smash pending damage: %+v", pd)
	}
}
