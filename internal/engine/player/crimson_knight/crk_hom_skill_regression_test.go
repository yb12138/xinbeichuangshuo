package crimson_knight_test

import (
	"starcup-engine/internal/engine"
	playerpkg "starcup-engine/internal/engine/player"
	"starcup-engine/internal/testutils"
	"strings"
	"testing"

	"starcup-engine/internal/engine/skill"
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

func TestCrimsonKnightCalmMind_AutoGrantsEndedActionType(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
	if err := g.AddPlayer("p1", "Crimson", "crimson_knight", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p1.TurnState = model.NewPlayerTurnState()
	p1.Form = model.FormCrimsonKnightHotBlooded
	p1.Crystal = 1

	h := skills.GetHandler("crk_calm_mind")
	if h == nil {
		t.Fatalf("crk_calm_mind handler not found")
	}
	ctx := g.BuildContext(p1, nil, model.TimingOnActionEnd, &model.EventContext{
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
	if got := p1.Form; got != "" {
		t.Fatalf("expected hot form cleared, got %q", got)
	}
	if g.State.PendingInterrupt != nil {
		t.Fatalf("expected calm mind to resolve without extra choice prompt, got %+v", g.State.PendingInterrupt)
	}
	if len(p1.TurnState.PendingActions) == 0 {
		t.Fatalf("expected one pending action from calm mind")
	}
	last := p1.TurnState.PendingActions[len(p1.TurnState.PendingActions)-1]
	if last.MustType != string(model.ActionMagic) {
		t.Fatalf("expected calm mind to grant extra magic action, got %+v", last)
	}
}

func TestCrimsonKnightHotBlood_AutoReleaseOnTurnEnd(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
	if err := g.AddPlayer("p1", "Crimson", "crimson_knight", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p1.IsActive = true
	p1.Heal = 0
	p1.Form = model.FormCrimsonKnightHotBlooded
	g.State.CurrentTurn = 0
	g.State.TurnStage = model.TurnStageTurnEnd

	g.Drive()

	if got := p1.Form; got != "" {
		t.Fatalf("expected hot form cleared at turn end, got %q", got)
	}
	if got := p1.Heal; got != 2 {
		t.Fatalf("expected heal +2 at turn end, got %d", got)
	}
}

func TestCrimsonKnightHotBlood_NextTurnNoFallbackRelease(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
	if err := g.AddPlayer("p1", "Crimson", "crimson_knight", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p1.IsActive = true
	p1.Heal = 0
	p1.Form = model.FormCrimsonKnightHotBlooded
	g.State.CurrentTurn = 0

	// 规则约束：回合结束退形态只在 TurnEnd 固定时序触发，NextTurn 不再做兜底处理。
	g.NextTurn()

	if got := p1.Form; got != model.FormCrimsonKnightHotBlooded {
		t.Fatalf("expected hot form unchanged when bypassing TurnEnd, got %q", got)
	}
	if got := p1.Heal; got != 0 {
		t.Fatalf("expected no heal when bypassing TurnEnd, got %d", got)
	}
}

func TestCrimsonKnightHotForm_DamageOverflowNoMoraleLoss(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
	if err := g.AddPlayer("p1", "Attacker", "priest", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Crimson", "crimson_knight", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	g.State.CurrentTurn = 0
	g.State.TurnStage = model.TurnStageActionExecution

	p1 := g.State.Players["p1"]
	p2 := g.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p2.Form = model.FormCrimsonKnightHotBlooded

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

	testutils.MustHandleAction(t, g, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdAttack,
		TargetID: "p2",
		CardID:   testutils.PlayableCardID(t, g, "p1", 0),
	})
	testutils.MustHandleAction(t, g, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	})
	if g.State.PendingInterrupt == nil || !engine.IsDiscardSelectionInterrupt(g.State.PendingInterrupt) {
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
	testutils.MustHandleAction(t, g, model.PlayerAction{
		PlayerID:   "p2",
		Type:       model.CmdSelect,
		Selections: selections,
	})

	if got := g.State.BlueMorale; got != blueMoraleBefore {
		t.Fatalf("expected no morale loss in hot form damage overflow, before=%d after=%d", blueMoraleBefore, got)
	}
}

func TestHomRuneReforge_ReallocateAndOverflowCheckOnTurnEnd(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
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
	playerpkg.ClearForm(p1, model.FormWarHomunculusBurst)
	// 进入形态前 6 张手牌，符文改造摸1后=7（形态内上限+1），回合结束转正后应触发弃1。
	p1.Hand = makeHandCards(6, model.ElementFire)

	h := skills.GetHandler("hom_rune_reforge")
	if h == nil {
		t.Fatalf("hom_rune_reforge handler not found")
	}
	ctx := g.BuildContext(p1, nil, model.TimingOnTurnStart, &model.EventContext{
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
	if got := p1.Form; got != model.FormWarHomunculusBurst {
		t.Fatalf("expected burst form entered, got %q", got)
	}
	if g.State.PendingInterrupt == nil || choiceTypeOfInterrupt(g.State.PendingInterrupt) != "hom_rune_reforge_distribution" {
		t.Fatalf("expected hom_rune_reforge_distribution choice, got %+v", g.State.PendingInterrupt)
	}

	// 选择 战纹2 / 魔纹1
	if err := g.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{2}}); err != nil {
		t.Fatalf("choose rune distribution failed: %v", err)
	}
	if p1.Tokens["hom_war_rune"] != 2 || p1.Tokens["hom_magic_rune"] != 1 {
		t.Fatalf("unexpected rune distribution: war=%d magic=%d", p1.Tokens["hom_war_rune"], p1.Tokens["hom_magic_rune"])
	}

	g.State.CurrentTurn = 0
	g.State.TurnStage = model.TurnStageTurnEnd
	g.Drive()

	if got := p1.Form; got != "" {
		t.Fatalf("expected burst form cleared at turn end, got %q", got)
	}
	if g.State.PendingInterrupt == nil || !engine.IsDiscardSelectionInterrupt(g.State.PendingInterrupt) {
		t.Fatalf("expected discard interrupt after form ends overflow, got %+v", g.State.PendingInterrupt)
	}
	data, _ := g.State.PendingInterrupt.Context.(map[string]interface{})
	if dc, _ := data["discard_count"].(int); dc != 1 {
		t.Fatalf("expected discard_count=1 after form ends, got %v", data["discard_count"])
	}
}

func TestHomGlyphFusion_MaxXUsesDistinctElements(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
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
	ctx := g.BuildContext(p1, g.State.Players["p2"], model.TimingOnHitCheck, &model.EventContext{
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
		t.Fatalf("expected glyph fusion can use with at least 2 hand cards")
	}
	if err := h.Execute(ctx); err != nil {
		t.Fatalf("execute glyph fusion failed: %v", err)
	}
	data, _ := g.State.PendingInterrupt.Context.(map[string]interface{})
	// 验证是直接选牌模式
	if choiceType, _ := data["choice_type"].(string); choiceType != "hom_glyph_fusion_cards" {
		t.Fatalf("expected choice_type=hom_glyph_fusion_cards, got %s", choiceType)
	}
	// 验证 min_pick 为 2
	if minPick, _ := data["min_pick"].(int); minPick != 2 {
		t.Fatalf("expected min_pick=2, got %v", data["min_pick"])
	}

	// 直接多选两张彼此异系的牌（水系索引0和风系索引2，元素互不相同）
	if err := g.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0, 2}}); err != nil {
		t.Fatalf("choose glyph cards failed: %v", err)
	}
}

// TestHomGlyphFusion_CanTriggerWithSameElementAsAttack 验证：
// 攻击牌元素与手牌元素相同时，魔纹融合仍然可以触发（新规则：选择的牌彼此异系，不要求与攻击牌异系）
func TestHomGlyphFusion_CanTriggerWithSameElementAsAttack(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
	if err := g.AddPlayer("p1", "Hom", "war_homunculus", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["hom_magic_rune"] = 1
	// 手牌有火系和水系牌，攻击牌也是火系
	// 新规则下魔纹融合仍可触发：选择火系和水系彼此异系的牌即可
	p1.Hand = []model.Card{
		{ID: "f1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
		{ID: "w1", Name: "水涟斩", Type: model.CardTypeAttack, Element: model.ElementWater, Damage: 2},
	}
	h := skills.GetHandler("hom_glyph_fusion")
	if h == nil {
		t.Fatalf("hom_glyph_fusion handler not found")
	}
	damageVal := 2
	ctx := g.BuildContext(p1, g.State.Players["p2"], model.TimingOnHitCheck, &model.EventContext{
		Type:      model.EventAttack,
		SourceID:  p1.ID,
		TargetID:  "p2",
		DamageVal: &damageVal,
		Card: &model.Card{
			ID:      "atk",
			Name:    "火焰斩",
			Type:    model.CardTypeAttack,
			Element: model.ElementFire, // 攻击牌是火系，与手牌火系牌元素相同
			Damage:  2,
		},
		AttackInfo: &model.AttackEventInfo{ActionType: "Attack", IsHit: false},
	})

	// 检查魔纹融合是否可以触发（即使攻击牌元素与部分手牌相同）
	if !h.CanUse(ctx) {
		t.Fatalf("expected glyph fusion can trigger even when attack element matches some hand cards")
	}
	t.Logf("SUCCESS: hom_glyph_fusion triggers even when attack element (Fire) matches some hand cards (Fire, Water)")
}

func TestHomAttackMissResponseGroup_ChooseOneOnly(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
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
	p1.Tokens["hom_war_rune"] = 1
	p1.Tokens["hom_magic_rune"] = 1
	p1.Hand = []model.Card{
		{ID: "w1", Name: "水涟斩", Type: model.CardTypeAttack, Element: model.ElementWater, Damage: 2},
		{ID: "g1", Name: "风神斩", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 2},
	}

	ctx := g.BuildContext(p1, p2, model.TimingOnHitCheck, &model.EventContext{
		Type:     model.EventAttack,
		SourceID: p1.ID,
		TargetID: p2.ID,
		Card: &model.Card{
			ID:      "atk",
			Name:    "火焰斩",
			Type:    model.CardTypeAttack,
			Element: model.ElementFire,
			Damage:  2,
		},
		AttackInfo: &model.AttackEventInfo{ActionType: "Attack", IsHit: false},
	})

	g.Dispatcher().OnTiming(ctx.Timing, ctx)
	if g.State.PendingInterrupt == nil || g.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected attack miss response interrupt, got %+v", g.State.PendingInterrupt)
	}
	if !testutils.InterruptHasSkillID(g.State.PendingInterrupt, "hom_rage_suppress") ||
		!testutils.InterruptHasSkillID(g.State.PendingInterrupt, "hom_glyph_fusion") {
		t.Fatalf("expected both miss-response skills before choosing, got %+v", g.State.PendingInterrupt.SkillIDs)
	}

	if err := g.ConfirmResponseSkill("p1", "hom_rage_suppress"); err != nil {
		t.Fatalf("confirm rage suppress failed: %v", err)
	}
	if g.State.PendingInterrupt != nil && testutils.InterruptHasSkillID(g.State.PendingInterrupt, "hom_glyph_fusion") {
		t.Fatalf("expected glyph fusion to be removed after choosing rage suppress, got %+v", g.State.PendingInterrupt)
	}
}

func TestHomDualEcho_TargetChoiceCanCancel(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
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
	ctx := g.BuildContext(p1, p1, model.TimingOnDamageTaken, &model.EventContext{
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

	prompt := g.BuildChoicePrompt()
	if !promptHasOptionID(prompt, "cancel") {
		t.Fatalf("expected cancel option in hom_dual_echo_target prompt, got %+v", prompt)
	}

	if err := g.HandleInterruptAction(model.PlayerAction{
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

func TestHomDualEcho_WhenDamagingEnemyInTwoPlayerGameCanTargetSelf(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
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
	p1.Crystal = 1

	h := skills.GetHandler("hom_dual_echo")
	if h == nil {
		t.Fatalf("hom_dual_echo handler not found")
	}

	damageVal := 2
	ctx := g.BuildContext(p1, p2, model.TimingOnDamageTaken, &model.EventContext{
		Type:      model.EventDamage,
		SourceID:  p1.ID,
		TargetID:  p2.ID,
		DamageVal: &damageVal,
	})
	if !h.CanUse(ctx) {
		t.Fatalf("expected dual echo can use when damaging enemy")
	}
	if err := h.Execute(ctx); err != nil {
		t.Fatalf("execute dual echo failed: %v", err)
	}
	if g.State.PendingInterrupt == nil || choiceTypeOfInterrupt(g.State.PendingInterrupt) != "hom_dual_echo_target" {
		t.Fatalf("expected hom_dual_echo_target interrupt, got %+v", g.State.PendingInterrupt)
	}
	prompt := g.GetCurrentPrompt()
	if prompt == nil {
		t.Fatalf("expected dual echo target prompt")
	}
	if len(prompt.Options) < 1 || prompt.Options[0].Label != p1.Name {
		t.Fatalf("expected self to be the valid alternate target in two-player game, got %+v", prompt.Options)
	}
}

func TestHomDualEcho_TargetConfirmConsumesCostAndQueuesDamage(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
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
	ctx := g.BuildContext(p1, p1, model.TimingOnDamageTaken, &model.EventContext{
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
	if err := g.HandleInterruptAction(model.PlayerAction{
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
	if pd.SourceID != "p1" || pd.TargetID != "p2" || pd.Damage != 2 || pd.DamageType != "magic_no_morale" {
		t.Fatalf("unexpected pending damage %+v", pd)
	}
	if pd.CapDrawToHandLimit {
		t.Fatalf("expected dual echo to keep full damage draw instead of capping at hand limit")
	}
}

func TestHomDualEcho_NoMoraleDamageStillOverflowsAndDoesNotDropMorale(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
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
	blueMoraleBefore := g.State.BlueMorale

	h := skills.GetHandler("hom_dual_echo")
	if h == nil {
		t.Fatalf("hom_dual_echo handler not found")
	}
	damageVal := 2
	ctx := g.BuildContext(p1, p1, model.TimingOnDamageTaken, &model.EventContext{
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
	if err := g.HandleInterruptAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("confirm dual echo target failed: %v", err)
	}
	g.State.CombatStage = model.CombatStageCalcDamage
	if paused := g.ProcessPendingDamages(); !paused {
		t.Fatalf("expected overflow discard interrupt from full no-morale damage")
	}
	if g.State.PendingInterrupt == nil || !engine.IsDiscardSelectionInterrupt(g.State.PendingInterrupt) {
		t.Fatalf("expected discard interrupt after dual echo overflow, got %+v", g.State.PendingInterrupt)
	}
	data, _ := g.State.PendingInterrupt.Context.(map[string]interface{})
	if dc, _ := data["discard_count"].(int); dc != 1 {
		t.Fatalf("expected discard_count=1 after overflow, got %v", data["discard_count"])
	}
	if noMorale, _ := data["no_morale_loss"].(bool); !noMorale {
		t.Fatalf("expected overflow discard to carry no_morale_loss flag")
	}
	if err := g.HandleInterruptAction(model.PlayerAction{
		PlayerID:   "p2",
		Type:       model.CmdSelect,
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("resolve dual echo overflow discard failed: %v", err)
	}
	if got := g.State.BlueMorale; got != blueMoraleBefore {
		t.Fatalf("expected no morale loss from dual echo overflow, before=%d after=%d", blueMoraleBefore, got)
	}
}

func TestCrimsonKnightFaith_OnlyWhitelistedSelfDamageCanUseHeal(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
	if err := g.AddPlayer("p1", "Crimson", "crimson_knight", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	g.State.Deck = rules.InitDeck()
	g.State.CombatStage = model.CombatStageCalcDamage

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
			DamageType: model.MagicAttack,
		},
	}
	for i := 0; i < 5 && len(g.State.PendingDamageQueue) > 0; i++ {
		if paused := g.ProcessPendingDamages(); paused {
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
			DamageType:            model.MagicAttack,
			AllowCrimsonFaithHeal: true,
		},
	}
	g.State.PendingInterrupt = nil
	if paused := g.ProcessPendingDamages(); !paused {
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
	g := engine.NewGameEngine(testutils.NoopObserver{})
	if err := g.AddPlayer("p1", "Crimson", "crimson_knight", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	g.State.CombatStage = model.CombatStageCalcDamage

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
		Hook:     model.FieldHookOnBeforeAction,
	})

	g.RunFieldCardsForHook(p1, model.FieldHookOnBeforeAction, nil)
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

	if paused := g.ProcessPendingDamages(); !paused {
		t.Fatalf("expected heal choice interrupt for self-poison")
	}
	if g.State.PendingInterrupt == nil || choiceTypeOfInterrupt(g.State.PendingInterrupt) != "heal" {
		t.Fatalf("expected heal choice interrupt after self-poison, got %+v", g.State.PendingInterrupt)
	}
}

func TestCrimsonKnightBloodyPrayerXPrompt_NoZeroOption(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
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
	ctx := g.BuildContext(p1, nil, model.TimingActive, nil)
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
	g := engine.NewGameEngine(testutils.NoopObserver{})
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
	playerpkg.SetForm(p1, model.FormWarHomunculusBurst)
	p1.Hand = []model.Card{
		{ID: "f1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
		{ID: "f2", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
	}

	damageVal := 2
	h := skills.GetHandler("hom_rune_smash")
	if h == nil {
		t.Fatalf("hom_rune_smash handler not found")
	}
	ctx := g.BuildContext(p1, p2, model.TimingOnHitCheck, &model.EventContext{
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
	if g.State.PendingInterrupt == nil || choiceTypeOfInterrupt(g.State.PendingInterrupt) != "hom_rune_smash_cards" {
		t.Fatalf("expected hom_rune_smash_cards choice for direct card selection, got %+v", g.State.PendingInterrupt)
	}

	// 直接多选两张同系牌（索引0和1），不再需要先选择X
	if err := g.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0, 1}}); err != nil {
		t.Fatalf("choose rune smash cards failed: %v", err)
	}
	// Y=1：额外翻转1战纹并造成1点法伤
	if err := g.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}}); err != nil {
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

func TestHomRuneSmash_ResponseSkillOnAttackHit(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
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
	p1.Tokens["hom_war_rune"] = 1
	p1.Tokens["hom_magic_rune"] = 0
	// 只要有手牌就可以触发战纹碎击，选择的牌彼此同系即可（不要求与攻击牌同系）
	p1.Hand = []model.Card{
		{ID: "f1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
	}

	ctx := g.BuildContext(p1, p2, model.TimingOnHitCheck, &model.EventContext{
		Type:     model.EventAttack,
		SourceID: p1.ID,
		TargetID: p2.ID,
		Card: &model.Card{
			ID:      "atk",
			Name:    "火焰斩",
			Type:    model.CardTypeAttack,
			Element: model.ElementFire,
			Damage:  2,
		},
		AttackInfo: &model.AttackEventInfo{ActionType: "Attack", IsHit: true},
	})

	g.Dispatcher().OnTiming(ctx.Timing, ctx)

	// 检查是否有响应技能中断
	if g.State.PendingInterrupt == nil {
		t.Fatalf("expected response skill interrupt on attack hit, got nil")
	}
	if g.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected InterruptResponseSkill, got %s", g.State.PendingInterrupt.Type)
	}
	if !testutils.InterruptHasSkillID(g.State.PendingInterrupt, "hom_rune_smash") {
		t.Fatalf("expected hom_rune_smash in response skills, got %+v", g.State.PendingInterrupt.SkillIDs)
	}
	t.Logf("SUCCESS: hom_rune_smash is in response skills: %v", g.State.PendingInterrupt.SkillIDs)
}

// TestHomRuneSmash_CanTriggerWithDifferentElementFromAttack 验证：
// 攻击牌元素与手牌元素不同时，战纹碎击仍然可以触发（新规则：选择的牌彼此同系，不要求与攻击牌同系）
func TestHomRuneSmash_CanTriggerWithDifferentElementFromAttack(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
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
	p1.Tokens["hom_war_rune"] = 1
	p1.Tokens["hom_magic_rune"] = 0
	// 手牌有火系和水系牌，攻击牌是风系
	// 新规则下战纹碎击仍可触发：选择火系牌彼此同系即可
	p1.Hand = []model.Card{
		{ID: "f1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
		{ID: "w1", Name: "水涟斩", Type: model.CardTypeAttack, Element: model.ElementWater, Damage: 2},
	}

	ctx := g.BuildContext(p1, p2, model.TimingOnHitCheck, &model.EventContext{
		Type:     model.EventAttack,
		SourceID: p1.ID,
		TargetID: p2.ID,
		Card: &model.Card{
			ID:      "atk",
			Name:    "风神斩",
			Type:    model.CardTypeAttack,
			Element: model.ElementWind, // 攻击牌是风系，与手牌元素不同
			Damage:  2,
		},
		AttackInfo: &model.AttackEventInfo{ActionType: "Attack", IsHit: true},
	})

	g.Dispatcher().OnTiming(ctx.Timing, ctx)

	// 检查是否有响应技能中断（即使攻击牌元素与手牌元素不同，也应能触发）
	if g.State.PendingInterrupt == nil {
		t.Fatalf("expected response skill interrupt on attack hit even with different element, got nil")
	}
	if !testutils.InterruptHasSkillID(g.State.PendingInterrupt, "hom_rune_smash") {
		t.Fatalf("expected hom_rune_smash to trigger even when attack element differs from hand, got %+v", g.State.PendingInterrupt.SkillIDs)
	}
	t.Logf("SUCCESS: hom_rune_smash triggers even when attack element (Wind) differs from hand elements (Fire, Water)")
}

func TestHomRuneSmash_FullAttackFlow(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
	if err := g.AddPlayer("p1", "Hom", "war_homunculus", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	g.State.CurrentTurn = 0
	g.State.Deck = rules.InitDeck()
	g.State.TurnStage = model.TurnStageActionExecution

	p1 := g.State.Players["p1"]
	p2 := g.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p1.Tokens["hom_war_rune"] = 1
	p1.Tokens["hom_magic_rune"] = 0
	// 只要有手牌就可以触发战纹碎击，选择的牌彼此同系即可（不要求与攻击牌同系）
	p1.Hand = []model.Card{
		{ID: "atk-fire", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
		{ID: "same-fire", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
	}
	p2.Hand = []model.Card{
		{ID: "def-light", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight},
	}

	// 发起攻击
	testutils.MustHandleAction(t, g, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdAttack,
		TargetID: "p2",
		CardID:   testutils.PlayableCardID(t, g, "p1", 0),
	})

	// Drive 进入战斗响应阶段
	g.Drive()
	t.Logf("After attack Drive: TurnStage=%s, PendingInterrupt=%+v, CombatStage=%s", g.State.TurnStage, g.State.PendingInterrupt, g.State.CombatStage)

	// 目标承受伤害（跳过防御/应战）
	testutils.MustHandleAction(t, g, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	})

	// Drive 处理伤害结算和后续
	g.Drive()
	t.Logf("After take Drive: TurnStage=%s, PendingInterrupt=%+v, CombatStage=%s", g.State.TurnStage, g.State.PendingInterrupt, g.State.CombatStage)

	// 伤害结算后应出现战纹碎击响应技能中断
	if g.State.PendingInterrupt == nil {
		t.Fatalf("expected response skill prompt after damage resolution, got nil")
	}
	if g.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected ResponseSkill interrupt, got %+v", g.State.PendingInterrupt)
	}

	testutils.RequireResponseSkillPrompt(t, g, "p1")
	if !testutils.InterruptHasSkillID(g.State.PendingInterrupt, "hom_rune_smash") {
		t.Fatalf("expected hom_rune_smash in response skills, got %+v", g.State.PendingInterrupt.SkillIDs)
	}
	t.Logf("SUCCESS: hom_rune_smash triggered after attack hit in full flow")
}
