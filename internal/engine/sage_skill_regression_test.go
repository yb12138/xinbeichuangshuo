package engine

import (
	"strings"
	"testing"

	"starcup-engine/internal/model"
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

func runUntilChoiceInterrupt(g *GameEngine, maxStep int) {
	for i := 0; i < maxStep; i++ {
		if g.State.PendingInterrupt != nil {
			return
		}
		if len(g.State.PendingDamageQueue) == 0 {
			return
		}
		g.processPendingDamages()
	}
}

func TestSageMagicRebound_SameElementDiscardChain(t *testing.T) {
	g := NewGameEngine(noopObserver{})
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
	// 伤害摸牌固定为非火系，确保“同系弃牌”候选稳定为上述3张火系牌。
	g.State.Deck = []model.Card{
		sageTestCard("d1", "水涟斩", model.CardTypeAttack, model.ElementWater),
	}

	g.AddPendingDamage(model.PendingDamage{
		SourceID:   "p2",
		TargetID:   "p1",
		Damage:     1,
		DamageType: "magic",
		Stage:      0,
	})
	g.State.Phase = model.PhasePendingDamageResolution

	runUntilChoiceInterrupt(g, 12)
	if g.State.PendingInterrupt == nil {
		t.Fatalf("expected sage rebound confirm interrupt, got nil")
	}
	if got := choiceTypeOf(g.State.PendingInterrupt); got != "sage_magic_rebound_confirm" {
		t.Fatalf("expected choice_type sage_magic_rebound_confirm, got %q", got)
	}

	if err := g.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("confirm rebound failed: %v", err)
	}
	if got := choiceTypeOf(g.State.PendingInterrupt); got != "sage_magic_rebound_x" {
		t.Fatalf("expected choice_type sage_magic_rebound_x, got %q", got)
	}

	// 选择 X=3（选项从 X=2 开始，索引 1 -> X=3）
	if err := g.handleWeakChoiceInput("p1", 1); err != nil {
		t.Fatalf("choose rebound x failed: %v", err)
	}
	if got := choiceTypeOf(g.State.PendingInterrupt); got != "sage_magic_rebound_element" {
		t.Fatalf("expected choice_type sage_magic_rebound_element, got %q", got)
	}

	// 仅有火系满足 X=3。
	if err := g.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("choose rebound element failed: %v", err)
	}
	if got := choiceTypeOf(g.State.PendingInterrupt); got != "sage_magic_rebound_cards" {
		t.Fatalf("expected choice_type sage_magic_rebound_cards, got %q", got)
	}

	// 连续3次选择同系牌（同系允许，不能被“异系去重”误伤）。
	for i := 0; i < 3; i++ {
		if err := g.handleWeakChoiceInput("p1", 0); err != nil {
			t.Fatalf("choose rebound cards step=%d failed: %v", i+1, err)
		}
	}
	if got := choiceTypeOf(g.State.PendingInterrupt); got != "sage_magic_rebound_target" {
		t.Fatalf("expected choice_type sage_magic_rebound_target, got %q", got)
	}

	// 目标选 p2（玩家顺序 p1,p2 -> 索引1）。
	if err := g.handleWeakChoiceInput("p1", 1); err != nil {
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
	if first.TargetID != "p1" || first.Damage != 3 || !strings.EqualFold(first.DamageType, "magic") {
		t.Fatalf("unexpected first rebound damage: %+v", first)
	}
	if second.TargetID != "p2" || second.Damage != 2 || !strings.EqualFold(second.DamageType, "magic") {
		t.Fatalf("unexpected second rebound damage: %+v", second)
	}
	if got := len(p1.Hand); got != 1 {
		t.Fatalf("expected 1 card left after discarding 3 same-element cards, got %d", got)
	}
}

// 回归：法术反弹的触发时点必须在“承伤摸牌完成之后”。
// 若触发早于摸牌，本用例中将无法凑出2张同系牌，不会出现反弹询问。
func TestSageMagicRebound_TriggerAfterDamageDraw(t *testing.T) {
	g := NewGameEngine(noopObserver{})
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
	// 受1点法伤后会摸1张；这张牌补成“第2张同系牌”，使法术反弹满足 X>1。
	g.State.Deck = []model.Card{
		sageTestCard("f2", "炎流", model.CardTypeMagic, model.ElementFire),
	}

	g.AddPendingDamage(model.PendingDamage{
		SourceID:   "p2",
		TargetID:   "p1",
		Damage:     1,
		DamageType: "magic",
		Stage:      0,
	})
	g.State.Phase = model.PhasePendingDamageResolution

	runUntilChoiceInterrupt(g, 12)
	if g.State.PendingInterrupt == nil {
		t.Fatalf("expected rebound confirm after damage draw, got nil")
	}
	if got := choiceTypeOf(g.State.PendingInterrupt); got != "sage_magic_rebound_confirm" {
		t.Fatalf("expected choice_type sage_magic_rebound_confirm, got %q", got)
	}
	if got := len(p1.Hand); got != 2 {
		t.Fatalf("expected draw finished before rebound confirm, hand should be 2, got %d", got)
	}
}

// 回归：同一次结算链里若连续承受两次1点法术伤害，应逐条触发法术反弹询问。
func TestSageMagicRebound_TwoOneMagicDamagesPromptTwice(t *testing.T) {
	g := NewGameEngine(noopObserver{})
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
		DamageType: "magic",
		Stage:      0,
	})
	g.AddPendingDamage(model.PendingDamage{
		SourceID:   "p2",
		TargetID:   "p1",
		Damage:     1,
		DamageType: "magic",
		Stage:      0,
	})
	g.State.Phase = model.PhasePendingDamageResolution

	runUntilChoiceInterrupt(g, 16)
	if got := choiceTypeOf(g.State.PendingInterrupt); got != "sage_magic_rebound_confirm" {
		t.Fatalf("expected first rebound confirm, got %q", got)
	}
	// 第一次选择不发动，流程应继续到下一条1点法伤并再次询问。
	if err := g.handleWeakChoiceInput("p1", 1); err != nil {
		t.Fatalf("skip first rebound confirm failed: %v", err)
	}
	runUntilChoiceInterrupt(g, 16)
	if got := choiceTypeOf(g.State.PendingInterrupt); got != "sage_magic_rebound_confirm" {
		t.Fatalf("expected second rebound confirm after next 1-damage, got %q", got)
	}
}

// 回归：对自己发动法术反弹时会形成嵌套结算；
// 新产生伤害遵循“后产生先结算”（LIFO）顺序。
func TestSageMagicRebound_SelfTargetNestedLIFO(t *testing.T) {
	g := NewGameEngine(noopObserver{})
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
	// 自己作为目标时，首轮反弹会生成 2 与 1 两段自伤；
	// 其中 1 点结算后再触发新一轮反弹。这里准备同系补牌，确保能进入嵌套。
	g.State.Deck = []model.Card{
		sageTestCard("w1", "浪涌1", model.CardTypeAttack, model.ElementWater),
		sageTestCard("w2", "浪涌2", model.CardTypeMagic, model.ElementWater),
		sageTestCard("w3", "浪涌3", model.CardTypeAttack, model.ElementWater),
	}

	g.AddPendingDamage(model.PendingDamage{
		SourceID:   "p2",
		TargetID:   "p1",
		Damage:     1,
		DamageType: "magic",
		Stage:      0,
	})
	g.State.Phase = model.PhasePendingDamageResolution

	runUntilChoiceInterrupt(g, 16)
	if got := choiceTypeOf(g.State.PendingInterrupt); got != "sage_magic_rebound_confirm" {
		t.Fatalf("expected rebound confirm, got %q", got)
	}
	if err := g.handleWeakChoiceInput("p1", 0); err != nil { // 发动
		t.Fatalf("confirm rebound failed: %v", err)
	}
	if err := g.handleWeakChoiceInput("p1", 0); err != nil { // X=2
		t.Fatalf("choose rebound x=2 failed: %v", err)
	}
	if err := g.handleWeakChoiceInput("p1", 0); err != nil { // 选择火系
		t.Fatalf("choose rebound element failed: %v", err)
	}
	if err := g.handleWeakChoiceInput("p1", 0); err != nil { // 选第1张火牌
		t.Fatalf("choose rebound card#1 failed: %v", err)
	}
	if err := g.handleWeakChoiceInput("p1", 0); err != nil { // 选第2张火牌
		t.Fatalf("choose rebound card#2 failed: %v", err)
	}
	if got := choiceTypeOf(g.State.PendingInterrupt); got != "sage_magic_rebound_target" {
		t.Fatalf("expected rebound target choice, got %q", got)
	}
	if err := g.handleWeakChoiceInput("p1", 0); err != nil { // 目标选自己（playerOrder: p1,p2）
		t.Fatalf("choose rebound self target failed: %v", err)
	}

	if got := len(g.State.PendingDamageQueue); got < 2 {
		t.Fatalf("expected at least 2 rebound damages queued, got %d", got)
	}
	first := g.State.PendingDamageQueue[0]
	second := g.State.PendingDamageQueue[1]
	if first.TargetID != "p1" || first.Damage != 2 || !strings.EqualFold(first.DamageType, "magic") {
		t.Fatalf("expected first nested damage self=2, got %+v", first)
	}
	if second.TargetID != "p1" || second.Damage != 1 || !strings.EqualFold(second.DamageType, "magic") {
		t.Fatalf("expected second nested damage self=1, got %+v", second)
	}

	// 继续推进：2点自伤与1点自伤结算后，应因后者再次进入法术反弹询问（嵌套触发）。
	runUntilChoiceInterrupt(g, 24)
	if got := choiceTypeOf(g.State.PendingInterrupt); got != "sage_magic_rebound_confirm" {
		t.Fatalf("expected nested rebound confirm after self-target chain, got %q", got)
	}
}

// 回归：魔道法典 X=2 且目标为自己时，会产生两次1点法术伤害；
// 若同系手牌条件满足，应逐次出现法术反弹询问。
func TestSageArcaneCodex_SelfTargetTriggersReboundPerOneDamage(t *testing.T) {
	g := NewGameEngine(noopObserver{})
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
	g.State.Phase = model.PhaseActionSelection

	if err := g.UseSkill("p1", "sage_arcane_codex", nil, nil); err != nil {
		t.Fatalf("use arcane codex failed: %v", err)
	}
	if got := p1.Gem; got != 0 {
		t.Fatalf("expected arcane codex consume 1 gem, got %d", got)
	}
	if got := choiceTypeOf(g.State.PendingInterrupt); got != "sage_arcane_x" {
		t.Fatalf("expected choice_type sage_arcane_x, got %q", got)
	}

	// 选 X=2（minX=2，索引0）。
	if err := g.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("choose arcane x=2 failed: %v", err)
	}
	if got := choiceTypeOf(g.State.PendingInterrupt); got != "sage_arcane_cards" {
		t.Fatalf("expected choice_type sage_arcane_cards, got %q", got)
	}

	// 弃2张异系（水、地）。
	if err := g.handleWeakChoiceInput("p1", 2); err != nil {
		t.Fatalf("choose arcane card#1(water) failed: %v", err)
	}
	if err := g.handleWeakChoiceInput("p1", 2); err != nil {
		t.Fatalf("choose arcane card#2(earth) failed: %v", err)
	}
	if got := choiceTypeOf(g.State.PendingInterrupt); got != "sage_arcane_target" {
		t.Fatalf("expected choice_type sage_arcane_target, got %q", got)
	}

	// 目标选自己（player order: p1,p2）。
	if err := g.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("choose arcane self target failed: %v", err)
	}
	if got := len(g.State.PendingDamageQueue); got < 2 {
		t.Fatalf("expected 2 pending magic damages from arcane self-target, got %d", got)
	}
	if g.State.PendingDamageQueue[0].TargetID != "p1" || g.State.PendingDamageQueue[0].Damage != 1 {
		t.Fatalf("expected first pending self magic damage=1, got %+v", g.State.PendingDamageQueue[0])
	}
	if g.State.PendingDamageQueue[1].TargetID != "p1" || g.State.PendingDamageQueue[1].Damage != 1 {
		t.Fatalf("expected second pending self magic damage=1, got %+v", g.State.PendingDamageQueue[1])
	}

	// 第一段1点法伤结算后应出现法术反弹询问。
	runUntilChoiceInterrupt(g, 16)
	if got := choiceTypeOf(g.State.PendingInterrupt); got != "sage_magic_rebound_confirm" {
		t.Fatalf("expected first rebound confirm after first 1-damage, got %q", got)
	}
	// 跳过第一段反弹，继续处理第二段1点法伤。
	if err := g.handleWeakChoiceInput("p1", 1); err != nil {
		t.Fatalf("skip first rebound confirm failed: %v", err)
	}
	runUntilChoiceInterrupt(g, 16)
	if got := choiceTypeOf(g.State.PendingInterrupt); got != "sage_magic_rebound_confirm" {
		t.Fatalf("expected second rebound confirm after second 1-damage, got %q", got)
	}
}

func TestSageHolyCodex_XAndTargetCountBoundaries(t *testing.T) {
	g := NewGameEngine(noopObserver{})
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
	g.State.Phase = model.PhaseActionSelection

	if err := g.UseSkill("p1", "sage_holy_codex", nil, nil); err != nil {
		t.Fatalf("use holy codex failed: %v", err)
	}
	if got := p1.Gem; got != 0 {
		t.Fatalf("expected holy codex consume exactly 1 gem, got gem=%d", got)
	}
	if got := choiceTypeOf(g.State.PendingInterrupt); got != "sage_holy_x" {
		t.Fatalf("expected choice_type sage_holy_x, got %q", got)
	}

	// 越界：maxX=4 时，索引2 -> X=5，应报错。
	if err := g.handleWeakChoiceInput("p1", 2); err == nil || !strings.Contains(err.Error(), "无效的X值") {
		t.Fatalf("expected invalid X boundary error, got %v", err)
	}
	if got := choiceTypeOf(g.State.PendingInterrupt); got != "sage_holy_x" {
		t.Fatalf("expected still stay at sage_holy_x after invalid input, got %q", got)
	}

	// 选择最大 X=4（索引1）。
	if err := g.handleWeakChoiceInput("p1", 1); err != nil {
		t.Fatalf("choose holy x=4 failed: %v", err)
	}
	if got := choiceTypeOf(g.State.PendingInterrupt); got != "sage_holy_cards" {
		t.Fatalf("expected choice_type sage_holy_cards, got %q", got)
	}

	// 依次选择4张异系牌（每次选当前候选第一张）。
	for i := 0; i < 4; i++ {
		if err := g.handleWeakChoiceInput("p1", 0); err != nil {
			t.Fatalf("choose holy cards step=%d failed: %v", i+1, err)
		}
	}
	if got := choiceTypeOf(g.State.PendingInterrupt); got != "sage_holy_target_count" {
		t.Fatalf("expected choice_type sage_holy_target_count, got %q", got)
	}

	// 越界：X=4 时最多只能选2名角色治疗，索引3应报错。
	if err := g.handleWeakChoiceInput("p1", 3); err == nil || !strings.Contains(err.Error(), "无效的治疗目标数量") {
		t.Fatalf("expected invalid target count boundary error, got %v", err)
	}
	if got := choiceTypeOf(g.State.PendingInterrupt); got != "sage_holy_target_count" {
		t.Fatalf("expected still stay at sage_holy_target_count after invalid input, got %q", got)
	}

	// 选择边界上限：2名角色。
	if err := g.handleWeakChoiceInput("p1", 2); err != nil {
		t.Fatalf("choose holy target count failed: %v", err)
	}
	if got := choiceTypeOf(g.State.PendingInterrupt); got != "sage_holy_targets" {
		t.Fatalf("expected choice_type sage_holy_targets, got %q", got)
	}

	// 依次选择 p1 与 p2 为治疗目标（每次选当前候选第一项）。
	if err := g.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("choose holy target#1 failed: %v", err)
	}
	if err := g.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("choose holy target#2 failed: %v", err)
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
	if pd.SourceID != "p1" || pd.TargetID != "p1" || pd.Damage != 3 || !strings.EqualFold(pd.DamageType, "magic") {
		t.Fatalf("unexpected holy codex self damage: %+v", pd)
	}
}

func TestSageExtract_CanReachFourthEnergyAndStopsAtCap(t *testing.T) {
	g := NewGameEngine(noopObserver{})
	if err := g.AddPlayer("p1", "Sage", "sage", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	g.State.CurrentTurn = 0
	g.State.Phase = model.PhaseActionSelection

	p1 := g.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 3
	p1.Crystal = 0
	g.State.RedGems = 1
	g.State.RedCrystals = 0

	mustHandleAction(t, g, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdExtract,
	})
	requireChoicePrompt(t, g, "p1", "extract")

	mustHandleAction(t, g, model.PlayerAction{
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
	err := g.handleExtract(p1)
	if err == nil || !strings.Contains(err.Error(), "能量已达上限") {
		t.Fatalf("expected extract blocked at cap=4, got err=%v", err)
	}
}
