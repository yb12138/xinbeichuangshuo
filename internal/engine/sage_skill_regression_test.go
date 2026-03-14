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
