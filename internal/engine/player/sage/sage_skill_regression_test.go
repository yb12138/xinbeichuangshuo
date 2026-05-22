package sage_test

import (
	"strings"
	"testing"

	"starcup-engine/internal/data"
	"starcup-engine/internal/engine"
	"starcup-engine/internal/engine/core/runtimeutil"
	"starcup-engine/internal/model"
	"starcup-engine/internal/testutils"
)

func sageTestCard(id, name string, cardType model.CardType, ele model.Element) model.Card {
	return model.Card{
		ID:      id,
		Name:    name,
		Type:    cardType,
		Element: ele,
		Faction: "咏",
		Damage:  2,
	}
}

func runUntilChoiceInterrupt(g *engine.GameEngine, maxStep int) {
	for i := 0; i < maxStep; i++ {
		if g.State.PendingInterrupt != nil {
			return
		}
		if len(g.State.PendingDamageQueue) == 0 {
			return
		}
		g.ProcessPendingDamages()
	}
}

func TestSageMagicRebound_SameElementDiscardChain(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
	if err := g.AddPlayer("p1", "Sage", "sage", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p1.Hand = []model.Card{
		sageTestCard("f1", "火焰斩", model.CardTypeAttack, model.ElementFire),
		sageTestCard("f2", "烈焰击", model.CardTypeAttack, model.ElementFire),
		sageTestCard("f3", "炎刃", model.CardTypeMagic, model.ElementFire),
	}
	// 伤害摸牌固定为非火系，确保"同系弃牌"候选稳定为上述3张火系牌。
	g.State.Deck = []model.Card{
		sageTestCard("d1", "水涟斩", model.CardTypeAttack, model.ElementWater),
	}

	g.AddPendingDamage(model.PendingDamage{
		SourceID:   "p2",
		TargetID:   "p1",
		Damage:     1,
		DamageType: model.MagicAttack,
	})
	g.State.CombatStage = model.CombatStageCalcDamage

	runUntilChoiceInterrupt(g, 12)
	if g.State.PendingInterrupt == nil {
		t.Fatalf("expected sage rebound confirm interrupt, got nil")
	}
	if got := testutils.ChoiceTypeOfInterrupt(g.State.PendingInterrupt); got != "sage_magic_rebound_confirm" {
		t.Fatalf("expected choice_type sage_magic_rebound_confirm, got %q", got)
	}
	ctxData := testutils.RequireChoiceContext(t, g, "p1", "sage_magic_rebound_confirm")
	flow := testutils.RequirePromptFlow(t, ctxData, "sage_magic_rebound", "confirm")

	if err := g.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("confirm rebound failed: %v", err)
	}
	// 新流程：确认后直接进入元素选择（不再有X选择）
	if got := testutils.ChoiceTypeOfInterrupt(g.State.PendingInterrupt); got != "sage_magic_rebound_element" {
		t.Fatalf("expected choice_type sage_magic_rebound_element, got %q", got)
	}
	ctxData = testutils.RequireChoiceContext(t, g, "p1", "sage_magic_rebound_element")
	flow = testutils.RequirePromptFlow(t, ctxData, "sage_magic_rebound", "element")
	if got := flow.Selection("confirm").OptionIndexes; len(got) != 1 || got[0] != 0 {
		t.Fatalf("expected rebound flow to accumulate confirm yes, got %+v in %+v", got, flow)
	}

	// 仅有火系满足至少2张。
	if err := g.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("choose rebound element failed: %v", err)
	}
	if got := testutils.ChoiceTypeOfInterrupt(g.State.PendingInterrupt); got != "sage_magic_rebound_cards" {
		t.Fatalf("expected choice_type sage_magic_rebound_cards, got %q", got)
	}
	ctxData = testutils.RequireChoiceContext(t, g, "p1", "sage_magic_rebound_cards")
	flow = testutils.RequirePromptFlow(t, ctxData, "sage_magic_rebound", "cards")
	if got := flow.Selection("element").Element; got != string(model.ElementFire) {
		t.Fatalf("expected rebound flow to accumulate fire element, got %s in %+v", got, flow)
	}

	// 多选3张同系牌（火系索引0,1,2）
	if err := g.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0, 1, 2}}); err != nil {
		t.Fatalf("choose rebound cards (multi-select 3 fire) failed: %v", err)
	}
	if got := testutils.ChoiceTypeOfInterrupt(g.State.PendingInterrupt); got != "sage_magic_rebound_target" {
		t.Fatalf("expected choice_type sage_magic_rebound_target, got %q", got)
	}
	ctxData = testutils.RequireChoiceContext(t, g, "p1", "sage_magic_rebound_target")
	flow = testutils.RequirePromptFlow(t, ctxData, "sage_magic_rebound", "target")
	if got := flow.Selection("cards").Count; got != 3 {
		t.Fatalf("expected rebound flow to accumulate x=3 cards, got %d in %+v", got, flow)
	}
	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
	if len(targetIDs) != 2 {
		t.Fatalf("expected rebound target pool include self (2 players), got %v", targetIDs)
	}

	// 目标池包含自己，选 p2 作为目标（索引1）。
	if err := g.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}}); err != nil {
		t.Fatalf("choose rebound target failed: %v", err)
	}
	if g.State.PendingInterrupt != nil {
		t.Fatalf("expected rebound flow finished, got pending interrupt %+v", g.State.PendingInterrupt)
	}

	// 反弹后应前插两段伤害：自己 X=3 先结算，目标 X-1=2 后结算。
	if got := len(g.State.PendingDamageQueue); got < 2 {
		t.Fatalf("expected at least 2 queued damages after rebound, got %d", got)
	}
	first := g.State.PendingDamageQueue[0]
	second := g.State.PendingDamageQueue[1]
	if first.TargetID != "p1" || first.Damage != 3 || !strings.EqualFold(string(first.DamageType), string(model.MagicAttack)) {
		t.Fatalf("unexpected first rebound damage: %+v", first)
	}
	if second.TargetID != "p2" || second.Damage != 2 || !strings.EqualFold(string(second.DamageType), string(model.MagicAttack)) {
		t.Fatalf("unexpected second rebound damage: %+v", second)
	}
	if got := len(p1.Hand); got != 1 {
		t.Fatalf("expected 1 card left after discarding 3 same-element cards, got %d", got)
	}
}

// 回归：法术反弹的触发时点必须在"承伤摸牌完成之后"。
// 若触发早于摸牌，本用例中将无法凑出2张同系牌，不会出现反弹询问。
func TestSageMagicRebound_DispatchAfterDamageDraw(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
	if err := g.AddPlayer("p1", "Sage", "sage", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p1.Hand = []model.Card{
		sageTestCard("f1", "火焰斩", model.CardTypeAttack, model.ElementFire),
	}
	// 受1点法伤后会摸1张；这张牌补成"第2张同系牌"，使法术反弹满足 X>1。
	g.State.Deck = []model.Card{
		sageTestCard("f2", "炎流", model.CardTypeMagic, model.ElementFire),
	}

	g.AddPendingDamage(model.PendingDamage{
		SourceID:   "p2",
		TargetID:   "p1",
		Damage:     1,
		DamageType: model.MagicAttack,
	})
	g.State.CombatStage = model.CombatStageCalcDamage

	runUntilChoiceInterrupt(g, 12)
	if g.State.PendingInterrupt == nil {
		t.Fatalf("expected rebound confirm after damage draw, got nil")
	}
	if got := testutils.ChoiceTypeOfInterrupt(g.State.PendingInterrupt); got != "sage_magic_rebound_confirm" {
		t.Fatalf("expected choice_type sage_magic_rebound_confirm, got %q", got)
	}
	if got := len(p1.Hand); got != 2 {
		t.Fatalf("expected draw finished before rebound confirm, hand should be 2, got %d", got)
	}
}

// 回归：同一次结算链里若连续承受两次1点法术伤害，应逐条触发法术反弹询问。
func TestSageMagicRebound_TwoOneMagicDamagesPromptTwice(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
	if err := g.AddPlayer("p1", "Sage", "sage", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p1.Hand = []model.Card{
		sageTestCard("f1", "火焰斩", model.CardTypeAttack, model.ElementFire),
		sageTestCard("f2", "烈焰击", model.CardTypeMagic, model.ElementFire),
	}
	g.State.Deck = []model.Card{
		sageTestCard("d1", "补牌1", model.CardTypeAttack, model.ElementWater),
		sageTestCard("d2", "补牌2", model.CardTypeAttack, model.ElementEarth),
	}

	g.AddPendingDamage(model.PendingDamage{
		SourceID:   "p2",
		TargetID:   "p1",
		Damage:     1,
		DamageType: model.MagicAttack,
	})
	g.AddPendingDamage(model.PendingDamage{
		SourceID:   "p2",
		TargetID:   "p1",
		Damage:     1,
		DamageType: model.MagicAttack,
	})
	g.State.CombatStage = model.CombatStageCalcDamage

	runUntilChoiceInterrupt(g, 16)
	if got := testutils.ChoiceTypeOfInterrupt(g.State.PendingInterrupt); got != "sage_magic_rebound_confirm" {
		t.Fatalf("expected first rebound confirm, got %q", got)
	}
	// 第一次选择不发动，流程应继续到下一条1点法伤并再次询问。
	if err := g.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}}); err != nil {
		t.Fatalf("skip first rebound confirm failed: %v", err)
	}
	runUntilChoiceInterrupt(g, 16)
	if got := testutils.ChoiceTypeOfInterrupt(g.State.PendingInterrupt); got != "sage_magic_rebound_confirm" {
		t.Fatalf("expected second rebound confirm after next 1-damage, got %q", got)
	}
}

func TestSageWisdomCodex_ForceDiscardAfterHeavyMagicDamage(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
	if err := g.AddPlayer("p1", "Sage", "sage", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p1.Hand = []model.Card{
		sageTestCard("f1", "火焰斩", model.CardTypeAttack, model.ElementFire),
	}
	g.State.Deck = []model.Card{
		sageTestCard("d1", "补牌1", model.CardTypeAttack, model.ElementWater),
		sageTestCard("d2", "补牌2", model.CardTypeAttack, model.ElementEarth),
		sageTestCard("d3", "补牌3", model.CardTypeAttack, model.ElementWind),
		sageTestCard("d4", "补牌4", model.CardTypeAttack, model.ElementThunder),
	}

	g.AddPendingDamage(model.PendingDamage{
		SourceID:   "p2",
		TargetID:   "p1",
		Damage:     4,
		DamageType: model.MagicAttack,
	})
	g.State.CombatStage = model.CombatStageCalcDamage

	runUntilChoiceInterrupt(g, 16)
	if g.State.PendingInterrupt == nil {
		t.Fatalf("expected wisdom codex forced discard interrupt, got nil")
	}
	if !engine.IsDiscardSelectionInterrupt(g.State.PendingInterrupt) {
		t.Fatalf("expected forced discard interrupt, got %+v", g.State.PendingInterrupt)
	}
	data, _ := g.State.PendingInterrupt.Context.(map[string]interface{})
	if discardCount := runtimeutil.ToIntContextValue(data["discard_count"]); discardCount != 1 {
		t.Fatalf("expected wisdom codex discard_count=1, got %d", discardCount)
	}
	if !runtimeutil.ToBoolContextValue(data["is_damage_resolution"]) {
		t.Fatalf("expected wisdom codex discard stay in damage resolution")
	}
	if got := p1.Gem; got != 2 {
		t.Fatalf("expected wisdom codex gain 2 gems after heavy magic damage, got %d", got)
	}
}

func TestSageArcaneCodex_TargetPoolIncludesSelfAndSelfDamageStillRunsRebound(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
	if err := g.AddPlayer("p1", "Sage", "sage", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 1
	// 手牌含 2 张火系 + 水/地各1张：
	// 魔道法典弃2张异系（水、地）后，仍保留2张同系（火）可触发法术反弹。
	p1.Hand = []model.Card{
		sageTestCard("f1", "火焰斩", model.CardTypeAttack, model.ElementFire),
		sageTestCard("f2", "烈焰击", model.CardTypeMagic, model.ElementFire),
		sageTestCard("w1", "水涟斩", model.CardTypeAttack, model.ElementWater),
		sageTestCard("e1", "地裂斩", model.CardTypeAttack, model.ElementEarth),
	}
	g.State.Deck = []model.Card{
		sageTestCard("d1", "补牌1", model.CardTypeAttack, model.ElementWind),
		sageTestCard("d2", "补牌2", model.CardTypeMagic, model.ElementThunder),
	}
	g.State.CurrentTurn = 0
	g.State.TurnStage = model.TurnStageActionExecution

	if err := g.UseSkill("p1", "sage_arcane_codex", nil, nil); err != nil {
		t.Fatalf("use arcane codex failed: %v", err)
	}
	if got := p1.Gem; got != 0 {
		t.Fatalf("expected arcane codex consume 1 gem, got %d", got)
	}
	if got := testutils.ChoiceTypeOfInterrupt(g.State.PendingInterrupt); got != "sage_arcane_cards" {
		t.Fatalf("expected choice_type sage_arcane_cards, got %q", got)
	}
	ctxData := testutils.RequireChoiceContext(t, g, "p1", "sage_arcane_cards")
	testutils.RequirePromptFlow(t, ctxData, "sage_arcane_codex", "cards")

	// 多选2张异系（水索引2、地索引3）。
	if err := g.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{2, 3}}); err != nil {
		t.Fatalf("choose arcane cards (water+earth) failed: %v", err)
	}
	if got := testutils.ChoiceTypeOfInterrupt(g.State.PendingInterrupt); got != "sage_arcane_target" {
		t.Fatalf("expected choice_type sage_arcane_target, got %q", got)
	}
	ctxData = testutils.RequireChoiceContext(t, g, "p1", "sage_arcane_target")
	flow := testutils.RequirePromptFlow(t, ctxData, "sage_arcane_codex", "target")
	if got := flow.Selection("cards").Count; got != 2 {
		t.Fatalf("expected arcane flow to accumulate x=2 cards, got %d in %+v", got, flow)
	}
	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
	if len(targetIDs) != 2 {
		t.Fatalf("expected arcane target pool include self (2 players), got %v", targetIDs)
	}

	// 目标池包含自己，选 p2 作为目标（索引1）。
	if err := g.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}}); err != nil {
		t.Fatalf("choose arcane target failed: %v", err)
	}
	if got := len(g.State.PendingDamageQueue); got < 2 {
		t.Fatalf("expected 2 pending magic damages from arcane codex, got %d", got)
	}
	if g.State.PendingDamageQueue[0].TargetID != "p2" || g.State.PendingDamageQueue[0].Damage != 1 {
		t.Fatalf("expected first pending target magic damage=1, got %+v", g.State.PendingDamageQueue[0])
	}
	if g.State.PendingDamageQueue[1].TargetID != "p1" || g.State.PendingDamageQueue[1].Damage != 1 {
		t.Fatalf("expected second pending self magic damage=1, got %+v", g.State.PendingDamageQueue[1])
	}

	// 目标伤害结算后继续处理自己的1点法伤，应出现一次法术反弹询问。
	runUntilChoiceInterrupt(g, 16)
	if g.State.PendingInterrupt == nil {
		t.Fatalf("expected rebound confirm after self damage from arcane codex")
	}
	if got := testutils.ChoiceTypeOfInterrupt(g.State.PendingInterrupt); got != "sage_magic_rebound_confirm" {
		t.Fatalf("expected rebound confirm after self 1-damage, got %q", got)
	}
	if err := g.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}}); err != nil {
		t.Fatalf("skip rebound confirm failed: %v", err)
	}
	runUntilChoiceInterrupt(g, 16)
	if g.State.PendingInterrupt != nil {
		t.Fatalf("expected only one rebound prompt after single self damage, got %+v", g.State.PendingInterrupt)
	}
}

func TestSageHolyCodex_MultiSelectCardsAndTargetCountBoundaries(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
	if err := g.AddPlayer("p1", "Sage", "sage", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p3", "EnemyA", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p4", "EnemyB", "assassin", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p2 := g.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 1
	p1.Hand = []model.Card{
		sageTestCard("h1", "火焰斩", model.CardTypeAttack, model.ElementFire),
		sageTestCard("h2", "水涟斩", model.CardTypeAttack, model.ElementWater),
		sageTestCard("h3", "风神斩", model.CardTypeAttack, model.ElementWind),
		sageTestCard("h4", "雷光斩", model.CardTypeAttack, model.ElementThunder),
	}
	g.State.CurrentTurn = 0
	g.State.TurnStage = model.TurnStageActionExecution

	if err := g.UseSkill("p1", "sage_holy_codex", nil, nil); err != nil {
		t.Fatalf("use holy codex failed: %v", err)
	}
	if got := p1.Gem; got != 0 {
		t.Fatalf("expected holy codex consume exactly 1 gem, got gem=%d", got)
	}
	// 新流程：直接进入卡牌多选（不再有X选择）
	if got := testutils.ChoiceTypeOfInterrupt(g.State.PendingInterrupt); got != "sage_holy_cards" {
		t.Fatalf("expected choice_type sage_holy_cards, got %q", got)
	}
	ctxData := testutils.RequireChoiceContext(t, g, "p1", "sage_holy_cards")
	testutils.RequirePromptFlow(t, ctxData, "sage_holy_codex", "cards")

	// 越界：只选2张牌（少于最小值3）
	if err := g.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0, 1}}); err == nil || !strings.Contains(err.Error(), "至少需要选择3张") {
		t.Fatalf("expected too few cards error, got %v", err)
	}

	// 选择4张异系牌（多选）
	if err := g.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0, 1, 2, 3}}); err != nil {
		t.Fatalf("choose holy cards (multi-select 4) failed: %v", err)
	}
	if got := testutils.ChoiceTypeOfInterrupt(g.State.PendingInterrupt); got != "sage_holy_targets" {
		t.Fatalf("expected choice_type sage_holy_targets, got %q", got)
	}
	ctxData = testutils.RequireChoiceContext(t, g, "p1", "sage_holy_targets")
	flow := testutils.RequirePromptFlow(t, ctxData, "sage_holy_codex", "targets")
	if got := flow.Selection("cards").Count; got != 4 {
		t.Fatalf("expected holy flow to accumulate x=4 cards, got %d in %+v", got, flow)
	}
	prompt := g.BuildChoicePrompt()
	if prompt == nil {
		t.Fatalf("expected holy targets prompt")
	}
	if prompt.Presentation == nil || prompt.Presentation.Kind != model.PresentationTargetPicker {
		t.Fatalf("expected holy targets prompt to use target_picker presentation, got %+v", prompt.Presentation)
	}
	// X=4时，最多治疗目标 = X-2 = 2；前端应在同一个目标选择阶段内点选最多2名角色后确认。
	if prompt.Min != 1 || prompt.Max != 2 {
		t.Fatalf("expected holy target picker min=1 max=2, got min=%d max=%d", prompt.Min, prompt.Max)
	}
	if len(prompt.Options) != 4 {
		t.Fatalf("expected all players as holy target options, got %d", len(prompt.Options))
	}
	if prompt.Options[0].ID != "p1" || prompt.Options[1].ID != "p2" {
		t.Fatalf("expected holy target option IDs to be player IDs, got %+v", prompt.Options[:2])
	}

	// 越界：X=4 时最多只能选2名角色治疗，一次提交3名目标应报错。
	if err := g.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0, 1, 2}}); err == nil || !strings.Contains(err.Error(), "治疗目标数量需为1-2名") {
		t.Fatalf("expected invalid target count boundary error, got %v", err)
	}
	if got := testutils.ChoiceTypeOfInterrupt(g.State.PendingInterrupt); got != "sage_holy_targets" {
		t.Fatalf("expected still stay at sage_holy_targets after invalid input, got %q", got)
	}

	// 一次性确认2名治疗目标。
	if err := g.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0, 1}}); err != nil {
		t.Fatalf("choose holy targets failed: %v", err)
	}

	if got := p1.Heal; got != 2 {
		t.Fatalf("expected p1 heal +2, got %d", got)
	}
	if got := p2.Heal; got != 2 {
		t.Fatalf("expected p2 heal +2, got %d", got)
	}
	if got := len(p1.Hand); got != 0 {
		t.Fatalf("expected 4 cards discarded by holy codex, got hand=%d", got)
	}
	if got := len(g.State.PendingDamageQueue); got == 0 {
		t.Fatalf("expected self magic damage queued after holy codex")
	}
	pd := g.State.PendingDamageQueue[0]
	if pd.SourceID != "p1" || pd.TargetID != "p1" || pd.Damage != 3 || !strings.EqualFold(string(pd.DamageType), string(model.MagicAttack)) {
		t.Fatalf("unexpected holy codex self damage: %+v", pd)
	}
}

func TestSageArcaneCodexSelfTargetQueuesTwoSelfDamages(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
	if err := g.AddPlayer("p1", "Sage", "sage", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := g.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 1
	p1.Hand = []model.Card{
		sageTestCard("h1", "火焰斩", model.CardTypeAttack, model.ElementFire),
		sageTestCard("h2", "水涟斩", model.CardTypeAttack, model.ElementWater),
		sageTestCard("h3", "风神斩", model.CardTypeAttack, model.ElementWind),
		sageTestCard("h4", "雷光斩", model.CardTypeAttack, model.ElementThunder),
	}
	g.State.CurrentTurn = 0
	g.State.TurnStage = model.TurnStageActionExecution

	if err := g.UseSkill("p1", "sage_arcane_codex", nil, nil); err != nil {
		t.Fatalf("use arcane codex failed: %v", err)
	}
	if err := g.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0, 1, 2, 3}}); err != nil {
		t.Fatalf("choose arcane cards failed: %v", err)
	}
	ctxData := testutils.RequireChoiceContext(t, g, "p1", "sage_arcane_target")
	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
	selfSelection := -1
	for idx, targetID := range targetIDs {
		if targetID == "p1" {
			selfSelection = idx
			break
		}
	}
	if selfSelection < 0 {
		t.Fatalf("expected target pool to include self, got %v", targetIDs)
	}
	if err := g.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{selfSelection}}); err != nil {
		t.Fatalf("choose self target failed: %v", err)
	}

	if got := len(g.State.PendingDamageQueue); got < 2 {
		t.Fatalf("expected two pending self damages, got %d", got)
	}
	for i := 0; i < 2; i++ {
		pd := g.State.PendingDamageQueue[i]
		if pd.SourceID != "p1" || pd.TargetID != "p1" || pd.Damage != 3 || !strings.EqualFold(string(pd.DamageType), string(model.MagicAttack)) {
			t.Fatalf("unexpected self damage %d: %+v", i, pd)
		}
	}
}

func TestSageArcaneCodexSelfTargetDrawsBothDamageInstances(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
	if err := g.AddPlayer("p1", "Sage", "sage", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := g.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 1
	p1.Hand = []model.Card{
		sageTestCard("h1", "火焰斩", model.CardTypeAttack, model.ElementFire),
		sageTestCard("h2", "水涟斩", model.CardTypeAttack, model.ElementWater),
		sageTestCard("h3", "风神斩", model.CardTypeAttack, model.ElementWind),
		sageTestCard("h4", "雷光斩", model.CardTypeAttack, model.ElementThunder),
	}
	g.State.Deck = []model.Card{
		sageTestCard("d1", "伤害摸牌1", model.CardTypeAttack, model.ElementFire),
		sageTestCard("d2", "伤害摸牌2", model.CardTypeAttack, model.ElementWater),
		sageTestCard("d3", "伤害摸牌3", model.CardTypeAttack, model.ElementWind),
		sageTestCard("d4", "伤害摸牌4", model.CardTypeAttack, model.ElementEarth),
		sageTestCard("d5", "伤害摸牌5", model.CardTypeAttack, model.ElementThunder),
		sageTestCard("d6", "伤害摸牌6", model.CardTypeAttack, model.ElementLight),
	}
	g.State.CurrentTurn = 0
	g.State.TurnStage = model.TurnStageActionExecution

	if err := g.UseSkill("p1", "sage_arcane_codex", nil, nil); err != nil {
		t.Fatalf("use arcane codex failed: %v", err)
	}
	if err := g.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0, 1, 2, 3}}); err != nil {
		t.Fatalf("choose arcane cards failed: %v", err)
	}
	ctxData := testutils.RequireChoiceContext(t, g, "p1", "sage_arcane_target")
	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
	selfSelection := -1
	for idx, targetID := range targetIDs {
		if targetID == "p1" {
			selfSelection = idx
			break
		}
	}
	if selfSelection < 0 {
		t.Fatalf("expected target pool to include self, got %v", targetIDs)
	}
	if err := g.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{selfSelection}}); err != nil {
		t.Fatalf("choose self target failed: %v", err)
	}

	for len(g.State.PendingDamageQueue) > 0 && g.State.PendingInterrupt == nil {
		g.ProcessPendingDamages()
	}
	if got := len(p1.Hand); got != 6 {
		t.Fatalf("expected two 3-damage self hits to draw 6 cards, got hand=%d cards=%+v", got, p1.Hand)
	}
}

func TestSageArcaneCodexSelfTargetReportsActualDrawWhenStockRunsOut(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	g := engine.NewGameEngine(obs)
	if err := g.AddPlayer("p1", "Sage", "sage", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := g.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 1
	p1.Hand = []model.Card{
		sageTestCard("h1", "火焰斩", model.CardTypeAttack, model.ElementFire),
		sageTestCard("h2", "水涟斩", model.CardTypeAttack, model.ElementWater),
		sageTestCard("h3", "风神斩", model.CardTypeAttack, model.ElementWind),
		sageTestCard("h4", "雷光斩", model.CardTypeAttack, model.ElementThunder),
	}
	g.State.Deck = nil
	g.State.DiscardPile = nil
	g.State.CurrentTurn = 0
	g.State.TurnStage = model.TurnStageActionExecution

	if err := g.UseSkill("p1", "sage_arcane_codex", nil, nil); err != nil {
		t.Fatalf("use arcane codex failed: %v", err)
	}
	if err := g.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0, 1, 2, 3}}); err != nil {
		t.Fatalf("choose arcane cards failed: %v", err)
	}
	ctxData := testutils.RequireChoiceContext(t, g, "p1", "sage_arcane_target")
	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
	selfSelection := -1
	for idx, targetID := range targetIDs {
		if targetID == "p1" {
			selfSelection = idx
			break
		}
	}
	if selfSelection < 0 {
		t.Fatalf("expected target pool to include self, got %v", targetIDs)
	}
	if err := g.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{selfSelection}}); err != nil {
		t.Fatalf("choose self target failed: %v", err)
	}

	for len(g.State.PendingDamageQueue) > 0 && g.State.PendingInterrupt == nil {
		g.ProcessPendingDamages()
	}
	if got := len(p1.Hand); got != 4 {
		t.Fatalf("expected only discarded cards to be redrawn when stock is empty, got hand=%d", got)
	}
	draws := make([]int, 0, 2)
	for _, ev := range obs.Events {
		if ev.Type == model.EventDrawCards && ev.DrawCards != nil && ev.DrawCards.Reason == "damage_draw" {
			draws = append(draws, ev.DrawCards.DrawCount)
		}
	}
	if len(draws) != 2 || draws[0] != 3 || draws[1] != 1 {
		t.Fatalf("expected actual damage draw events [3 1], got %v", draws)
	}
}

func TestSageExtract_CanReachFourthEnergyAndStopsAtCap(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
	if err := g.AddPlayer("p1", "Sage", "sage", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	g.State.CurrentTurn = 0
	g.State.TurnStage = model.TurnStageActionExecution

	p1 := g.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 3
	p1.Crystal = 0
	g.State.RedGems = 1
	g.State.RedCrystals = 0

	testutils.MustHandleAction(t, g, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdExtract,
	})
	testutils.RequireChoicePrompt(t, g, "p1", "extract")

	testutils.MustHandleAction(t, g, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	})

	if got := p1.Gem + p1.Crystal; got != 4 {
		t.Fatalf("expected sage energy reach cap=4 after extract, got %d (gem=%d crystal=%d)", got, p1.Gem, p1.Crystal)
	}
	if g.State.RedGems != 0 {
		t.Fatalf("expected one red gem extracted from camp pool, got red_gems=%d", g.State.RedGems)
	}

	// 到达4后应不可再提炼（上限锁死）。
	g.State.RedGems = 1
	err := g.HandleExtract(p1)
	if err == nil || !strings.Contains(err.Error(), "能量已达上限") {
		t.Fatalf("expected extract blocked at cap=4, got err=%v", err)
	}
}

func TestSageConfig_MetadataAlignsWithDocument(t *testing.T) {
	characters := data.GetCharacters()
	var sage *model.Character
	for _, character := range characters {
		if character.ID == "sage" {
			copy := character
			sage = &copy
			break
		}
	}
	if sage == nil {
		t.Fatalf("sage character not found")
	}

	var wisdom *model.SkillDefinition
	var rebound *model.SkillDefinition
	var arcane *model.SkillDefinition
	var holy *model.SkillDefinition
	for i := range sage.Skills {
		switch sage.Skills[i].ID {
		case "sage_wisdom_codex":
			wisdom = &sage.Skills[i]
		case "sage_magic_rebound":
			rebound = &sage.Skills[i]
		case "sage_arcane_codex":
			arcane = &sage.Skills[i]
		case "sage_holy_codex":
			holy = &sage.Skills[i]
		}
	}
	if wisdom == nil || rebound == nil || arcane == nil || holy == nil {
		t.Fatalf("expected sage core skills all present")
	}

	if strings.Contains(wisdom.Description, "可弃") {
		t.Fatalf("expected wisdom codex description to reflect forced discard, got %q", wisdom.Description)
	}
	if rebound.TargetType != model.TargetAny || rebound.MinTargets != 1 || rebound.MaxTargets != 1 {
		t.Fatalf("expected spell rebound target metadata any(1), got type=%v min=%d max=%d", rebound.TargetType, rebound.MinTargets, rebound.MaxTargets)
	}
	if arcane.CostGem != 1 || arcane.TargetType != model.TargetAny || arcane.MinTargets != 1 || arcane.MaxTargets != 1 {
		t.Fatalf("expected arcane codex metadata gem=1 target any(1), got gem=%d type=%v min=%d max=%d", arcane.CostGem, arcane.TargetType, arcane.MinTargets, arcane.MaxTargets)
	}
	if holy.CostGem != 1 || holy.TargetType != model.TargetAny || holy.MinTargets != 1 || holy.MaxTargets != 6 {
		t.Fatalf("expected holy codex metadata gem=1 target any(1..6), got gem=%d type=%v min=%d max=%d", holy.CostGem, holy.TargetType, holy.MinTargets, holy.MaxTargets)
	}
}

// TestSageArcaneCodex_SelfTargetReboundCombo 验证贤者自洗牌套路：
// 1. 魔道法典以自己为目标，弃2张异系牌，对自己造成2次各1点法术伤害
// 2. 第一次1点法伤：摸1牌（与3张同系一致），触发法术反弹(X=4)
// 3. 法术反弹：弃4张同系牌，对敌方造成3点伤害，自己受4点法伤摸4牌
// 4. 智慧法典触发（4点>3）：+2宝石弃1牌
// 5. 第二次1点法伤（魔道法典的自身伤害）：再触发法术反弹
func TestSageArcaneCodex_SelfTargetReboundCombo(t *testing.T) {
	g := engine.NewGameEngine(testutils.NoopObserver{})
	if err := g.AddPlayer("p1", "Sage", "sage", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 1
	// 手牌含 3 张火系 + 水/地各1张：
	// 魔道法典弃2张异系（水、地）后，保留3张同系（火）。
	// 第一次1点法伤摸1牌后，若抽到火系则变成4张同系可触发法术反弹。
	p1.Hand = []model.Card{
		sageTestCard("f1", "火焰斩", model.CardTypeAttack, model.ElementFire),
		sageTestCard("f2", "烈焰击", model.CardTypeMagic, model.ElementFire),
		sageTestCard("f3", "炎刃", model.CardTypeAttack, model.ElementFire),
		sageTestCard("w1", "水涟斩", model.CardTypeAttack, model.ElementWater),
		sageTestCard("e1", "地裂斩", model.CardTypeAttack, model.ElementEarth),
	}
	// 牌堆：确保每次伤害摸牌后都能凑出同系牌
	g.State.Deck = []model.Card{
		sageTestCard("d1", "火补1", model.CardTypeAttack, model.ElementFire), // 第一次1点法伤摸牌
		sageTestCard("d2", "火补2", model.CardTypeAttack, model.ElementFire), // 反弹4点自伤摸牌
		sageTestCard("d3", "火补3", model.CardTypeAttack, model.ElementFire),
		sageTestCard("d4", "火补4", model.CardTypeAttack, model.ElementFire),
		sageTestCard("d5", "火补5", model.CardTypeAttack, model.ElementFire), // 第二次1点法伤摸牌
		sageTestCard("d6", "火补6", model.CardTypeAttack, model.ElementFire), // 第二次反弹自伤摸牌
		sageTestCard("d7", "火补7", model.CardTypeAttack, model.ElementFire),
		sageTestCard("d8", "水补1", model.CardTypeAttack, model.ElementWater),
		sageTestCard("d9", "水补2", model.CardTypeAttack, model.ElementWater),
		sageTestCard("d10", "水补3", model.CardTypeAttack, model.ElementWater),
	}
	g.State.CurrentTurn = 0
	g.State.TurnStage = model.TurnStageActionExecution

	// --- Step 1: 使用魔道法典 ---
	if err := g.UseSkill("p1", "sage_arcane_codex", nil, nil); err != nil {
		t.Fatalf("use arcane codex failed: %v", err)
	}

	// --- Step 2: 多选2张异系牌（水索引3、地索引4）---
	if got := testutils.ChoiceTypeOfInterrupt(g.State.PendingInterrupt); got != "sage_arcane_cards" {
		t.Fatalf("expected choice_type sage_arcane_cards, got %q", got)
	}
	if err := g.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{3, 4}}); err != nil {
		t.Fatalf("choose arcane cards (water+earth) failed: %v", err)
	}

	// --- Step 3: 选择目标为自己 ---
	if got := testutils.ChoiceTypeOfInterrupt(g.State.PendingInterrupt); got != "sage_arcane_target" {
		t.Fatalf("expected choice_type sage_arcane_target, got %q", got)
	}
	// 选 p1（自己），索引0
	if err := g.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("choose self as arcane target failed: %v", err)
	}

	// 验证伤害队列：目标(p1)受1点 + 自身(p1)受1点，共2条
	if got := len(g.State.PendingDamageQueue); got < 2 {
		t.Fatalf("expected 2 pending magic damages (target=self, two hits), got %d", got)
	}

	// --- Step 4: 结算第一次1点法伤 → 摸1牌（火系）→ 触发法术反弹 ---
	runUntilChoiceInterrupt(g, 16)
	if got := testutils.ChoiceTypeOfInterrupt(g.State.PendingInterrupt); got != "sage_magic_rebound_confirm" {
		t.Fatalf("expected sage_magic_rebound_confirm, got %q", got)
	}

	// --- Step 5: 确认发动法术反弹 ---
	if err := g.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("confirm rebound failed: %v", err)
	}

	// --- Step 6: 选择元素（火系） ---
	// 新流程：确认后直接进入元素选择
	if got := testutils.ChoiceTypeOfInterrupt(g.State.PendingInterrupt); got != "sage_magic_rebound_element" {
		t.Fatalf("expected sage_magic_rebound_element, got %q", got)
	}
	if err := g.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("choose fire element failed: %v", err)
	}

	// --- Step 7: 多选4张同系牌 ---
	// 新流程：多选卡牌，X=选中数量
	if got := testutils.ChoiceTypeOfInterrupt(g.State.PendingInterrupt); got != "sage_magic_rebound_cards" {
		t.Fatalf("expected sage_magic_rebound_cards, got %q", got)
	}
	// 选择4张火系牌（索引0,1,2,3）
	if err := g.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0, 1, 2, 3}}); err != nil {
		t.Fatalf("choose rebound cards (multi-select 4 fire) failed: %v", err)
	}

	// --- Step 8: 选择法术反弹目标 → p2 ---
	if err := g.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}}); err != nil {
		t.Fatalf("choose rebound target p2 failed: %v", err)
	}

	// --- Step 9: 结算法术反弹的伤害 ---
	// 反弹伤害队列：4点法伤给p1 + 3点法伤给p2
	// 4点法伤 > 3 → 触发智慧法典（+2宝石，弃1牌）
	runUntilChoiceInterrupt(g, 32)
	t.Logf("After rebound resolution: interrupt=%v, damageQueue=%d, hand=%d",
		g.State.PendingInterrupt != nil, len(g.State.PendingDamageQueue), len(p1.Hand))

	// 处理所有中断直到遇到贤者技能中断或队列为空
	for g.State.PendingInterrupt != nil {
		ct := testutils.ChoiceTypeOfInterrupt(g.State.PendingInterrupt)
		t.Logf("Handling interrupt: type=%s, damageQueue=%d, hand=%d", ct, len(g.State.PendingDamageQueue), len(p1.Hand))
		if ct == "discard_card" || ct == "system_discard_cards" {
			if err := g.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
				t.Fatalf("discard failed (type=%s): %v", ct, err)
			}
			t.Logf("After discard: interrupt=%v, damageQueue=%d, hand=%d",
				g.State.PendingInterrupt != nil, len(g.State.PendingDamageQueue), len(p1.Hand))
		} else if strings.HasPrefix(ct, "sage_") {
			break
		} else {
			break
		}
	}
	t.Logf("After loop: interrupt=%v, damageQueue=%d, hand=%d",
		g.State.PendingInterrupt != nil, len(g.State.PendingDamageQueue), len(p1.Hand))

	// --- Step 10: 继续结算魔道法典的第二次1点法伤 → 应触发第二次法术反弹 ---
	for i := 0; i < 16; i++ {
		if g.State.PendingInterrupt != nil {
			break
		}
		if len(g.State.PendingDamageQueue) == 0 {
			break
		}
		t.Logf("Step %d: processing damage, queueLen=%d, hand=%d", i, len(g.State.PendingDamageQueue), len(p1.Hand))
		g.ProcessPendingDamages()
		t.Logf("Step %d: after damage, interrupt=%v, queueLen=%d, hand=%d", i, g.State.PendingInterrupt != nil, len(g.State.PendingDamageQueue), len(p1.Hand))
	}
	if g.State.PendingInterrupt == nil {
		fireCount2 := 0
		for _, c := range p1.Hand {
			if c.Element == model.ElementFire {
				fireCount2++
			}
		}
		t.Fatalf("expected rebound confirm after second 1-damage, but interrupt is nil. damageQueue=%d, hand=%d, fireCount=%d", len(g.State.PendingDamageQueue), len(p1.Hand), fireCount2)
	}
	if got := testutils.ChoiceTypeOfInterrupt(g.State.PendingInterrupt); got != "sage_magic_rebound_confirm" {
		ct := testutils.ChoiceTypeOfInterrupt(g.State.PendingInterrupt)
		t.Fatalf("expected second rebound confirm, got %q", ct)
	}
	// 第二次法术反弹成功触发，套路验证通过
}
