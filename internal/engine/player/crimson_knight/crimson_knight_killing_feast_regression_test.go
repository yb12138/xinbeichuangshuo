package crimson_knight_test

import (
	"starcup-engine/internal/engine"
	"starcup-engine/internal/testutils"
	"testing"

	"starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func hasCrimsonKnightSkillID(skillIDs []string, want string) bool {
	for _, id := range skillIDs {
		if id == want {
			return true
		}
	}
	return false
}

// 回归：杀戮盛宴在“攻击命中”响应后应稳定让本次攻击伤害+2（常规 2 -> 4）。
func TestCrimsonKnightKillingFeast_BoostsCurrentHitDamage(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Crimson", "crimson_knight", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p1.Heal = 0
	p2.Heal = 0
	p1.Hand = nil
	p2.Hand = nil
	p1.Tokens["crk_blood_mark"] = 1

	game.State.Deck = rules.InitDeck()
	game.State.CombatStage = model.CombatStageCalcDamage

	attackCard := model.Card{
		ID:      "atk-fiery-cut",
		Name:    "火焰斩",
		Type:    model.CardTypeAttack,
		Element: model.ElementFire,
		Damage:  2,
	}

	// cap=1 用于稳定复现旧实现中 append 导致 DamageVal 指针失效的问题。
	game.State.PendingDamageQueue = make([]model.PendingDamage, 1, 1)
	game.State.PendingDamageQueue[0] = model.PendingDamage{
		SourceID:   "p1",
		TargetID:   "p2",
		Damage:     2,
		DamageType: model.AttackDamage,
		Card:       &attackCard,
	}

	paused := game.ProcessPendingDamages()
	if !paused {
		t.Fatalf("expected pause for OnAttackHit response, got paused=%v", paused)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response skill interrupt, got %+v", game.State.PendingInterrupt)
	}
	if game.State.PendingInterrupt.PlayerID != "p1" {
		t.Fatalf("expected p1 to respond killing feast, got player=%s", game.State.PendingInterrupt.PlayerID)
	}
	if !hasCrimsonKnightSkillID(game.State.PendingInterrupt.SkillIDs, "crk_killing_feast") {
		t.Fatalf("expected crk_killing_feast in response list, got %+v", game.State.PendingInterrupt.SkillIDs)
	}

	if err := game.ConfirmResponseSkill("p1", "crk_killing_feast"); err != nil {
		t.Fatalf("confirm killing feast failed: %v", err)
	}

	for i := 0; i < 8 && len(game.State.PendingDamageQueue) > 0; i++ {
		if game.ProcessPendingDamages() {
			t.Fatalf("unexpected interrupt while draining pending damage: %+v", game.State.PendingInterrupt)
		}
	}
	if len(game.State.PendingDamageQueue) != 0 {
		t.Fatalf("pending damage queue not drained, len=%d", len(game.State.PendingDamageQueue))
	}

	if got := p1.Tokens["crk_blood_mark"]; got != 0 {
		t.Fatalf("expected blood mark consumed to 0, got %d", got)
	}
	// p2 仅承受本次攻击：应为 4 点（基础 2 + 杀戮盛宴 +2）。
	if got := len(p2.Hand); got != 4 {
		t.Fatalf("expected p2 draw 4 cards from boosted hit, got %d", got)
	}
	// p1 承受杀戮盛宴自伤 4 点（摸 4）。
	if got := len(p1.Hand); got != 4 {
		t.Fatalf("expected p1 draw 4 cards from self magic damage, got %d", got)
	}
}

func TestCrimsonKnightKillingFeast_SelfDamageResolvesBeforeAttackDamage(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Crimson", "crimson_knight", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p1.Heal = 1
	p2.Heal = 1
	p1.Tokens["crk_blood_mark"] = 1
	game.State.Deck = rules.InitDeck()
	game.State.CombatStage = model.CombatStageCalcDamage

	attackCard := model.Card{
		ID:      "atk-fiery-cut-order",
		Name:    "火焰斩",
		Type:    model.CardTypeAttack,
		Element: model.ElementFire,
		Damage:  2,
	}
	game.State.PendingDamageQueue = []model.PendingDamage{
		{
			SourceID:   "p1",
			TargetID:   "p2",
			Damage:     2,
			DamageType: model.AttackDamage,
			Card:       &attackCard,
		},
	}

	if paused := game.ProcessPendingDamages(); !paused {
		t.Fatalf("expected response interrupt before resolving hit damage")
	}
	if err := game.ConfirmResponseSkill("p1", "crk_killing_feast"); err != nil {
		t.Fatalf("confirm killing feast failed: %v", err)
	}

	// 修复目标：杀戮盛宴的自伤必须先于本次攻击伤害结算。
	if paused := game.ProcessPendingDamages(); !paused {
		t.Fatalf("expected heal choice interrupt from self-damage first")
	}
	if game.State.PendingInterrupt == nil || choiceTypeOfInterrupt(game.State.PendingInterrupt) != "heal" {
		t.Fatalf("expected heal choice interrupt, got %+v", game.State.PendingInterrupt)
	}
	if game.State.PendingInterrupt.PlayerID != "p1" {
		t.Fatalf("expected self-damage heal choice for p1 first, got %s", game.State.PendingInterrupt.PlayerID)
	}
}

func TestCrimsonKnightCrimsonCross_SelfDamageResolvesBeforeTargetDamage(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Crimson", "crimson_knight", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p1.Heal = 1
	p2.Heal = 1
	p1.Tokens["crk_blood_mark"] = 1
	p1.Hand = []model.Card{
		{ID: "m1", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementFire},
		{ID: "m2", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementWater},
	}

	h := skills.GetHandler("crk_crimson_cross")
	if h == nil {
		t.Fatalf("crk_crimson_cross handler not found")
	}
	ctx := game.BuildContext(p1, p2, model.TimingActive, nil)
	if !h.CanUse(ctx) {
		t.Fatalf("expected crimson cross can use with blood mark and 2 magic cards")
	}
	if err := h.Execute(ctx); err != nil {
		t.Fatalf("execute crimson cross failed: %v", err)
	}
	game.State.CombatStage = model.CombatStageCalcDamage

	if paused := game.ProcessPendingDamages(); !paused {
		t.Fatalf("expected heal choice interrupt from crimson cross self-damage")
	}
	if game.State.PendingInterrupt == nil || choiceTypeOfInterrupt(game.State.PendingInterrupt) != "heal" {
		t.Fatalf("expected heal choice interrupt, got %+v", game.State.PendingInterrupt)
	}
	if game.State.PendingInterrupt.PlayerID != "p1" {
		t.Fatalf("expected self-damage heal choice for p1 first, got %s", game.State.PendingInterrupt.PlayerID)
	}
}

func TestCrimsonKnightCrimsonCross_OnlyTargetsEnemy(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Crimson", "crimson_knight", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1
	p1.Tokens["crk_blood_mark"] = 1
	p1.Hand = []model.Card{
		{ID: "m1", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementFire},
		{ID: "m2", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementWater},
	}
	game.State.TurnStage = model.TurnStageActionExecution

	if err := game.UseSkill("p1", "crk_crimson_cross", []string{"p2"}, []int{0, 1}); err == nil {
		t.Fatalf("expected crimson cross to reject ally target")
	}

	p1.Crystal = 1
	p1.Tokens["crk_blood_mark"] = 1
	p1.Hand = []model.Card{
		{ID: "m1", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementFire},
		{ID: "m2", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementWater},
	}
	if err := game.UseSkill("p1", "crk_crimson_cross", []string{"p3"}, []int{0, 1}); err != nil {
		t.Fatalf("expected crimson cross to accept enemy target: %v", err)
	}
}
