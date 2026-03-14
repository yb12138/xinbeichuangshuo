package engine

import (
	"testing"

	"starcup-engine/internal/engine/skills"
	"starcup-engine/internal/model"
)

func spiritCasterTestCard(id, name string, cardType model.CardType, ele model.Element) model.Card {
	return model.Card{
		ID:          id,
		Name:        name,
		Type:        cardType,
		Element:     ele,
		Damage:      1,
		Description: name,
	}
}

func addSpiritCasterPowerForTest(p *model.Player, card model.Card) {
	if p == nil {
		return
	}
	p.AddFieldCard(&model.FieldCard{
		Card:     card,
		OwnerID:  p.ID,
		SourceID: p.ID,
		Mode:     model.FieldCover,
		Effect:   model.EffectSpiritCasterPower,
	})
	syncSpiritCasterPowerToken(p)
}

func TestSpiritCasterTalismanThunder_SealThenIncantThenDamage(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "SpiritCaster", "spirit_caster", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "EnemyA", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "EnemyB", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		spiritCasterTestCard("t1", "雷符", model.CardTypeMagic, model.ElementThunder), // 发动成本
		spiritCasterTestCard("h1", "补牌", model.CardTypeAttack, model.ElementFire),   // 念咒盖放
	}
	// p1 身上存在雷之封印：发动雷鸣时先触发封印伤害。
	p1.AddFieldCard(&model.FieldCard{
		Card:     spiritCasterTestCard("seal_t", "雷封印", model.CardTypeMagic, model.ElementThunder),
		OwnerID:  p1.ID,
		SourceID: "p2",
		Mode:     model.FieldEffect,
		Effect:   model.EffectSealThunder,
		Trigger:  model.EffectTriggerManual,
	})

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	if err := game.UseSkill("p1", "sc_talisman_thunder", []string{"p2", "p3"}, []int{0}); err != nil {
		t.Fatalf("use talisman thunder failed: %v", err)
	}
	if len(game.State.PendingDamageQueue) != 1 {
		t.Fatalf("expected only seal damage pending first, got %d", len(game.State.PendingDamageQueue))
	}
	if len(game.State.DeferredFollowups) != 1 {
		t.Fatalf("expected deferred talisman followup, got %d", len(game.State.DeferredFollowups))
	}

	// 先结算封印伤害，再继续灵符后续。
	if paused := game.processPendingDamages(); paused {
		t.Fatalf("unexpected interrupt while resolving seal damage")
	}
	if len(game.State.PendingDamageQueue) != 0 {
		t.Fatalf("expected seal damage queue consumed")
	}
	game.processDeferredFollowups()
	requireChoicePrompt(t, game, "p1", "sc_incant_confirm")

	if err := game.handleWeakChoiceInput("p1", 0); err != nil { // 发动念咒
		t.Fatalf("confirm incantation failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "sc_incant_card")
	if err := game.handleWeakChoiceInput("p1", 0); err != nil { // 盖放补牌
		t.Fatalf("choose incantation card failed: %v", err)
	}

	if got := spiritCasterPowerCount(p1, ""); got != 1 {
		t.Fatalf("expected 1 spirit power after incantation, got %d", got)
	}
	if len(game.State.PendingDamageQueue) != 2 {
		t.Fatalf("expected thunder damage queued for 2 targets, got %d", len(game.State.PendingDamageQueue))
	}
	if game.State.PendingDamageQueue[0].TargetID != "p3" || game.State.PendingDamageQueue[1].TargetID != "p2" {
		t.Fatalf("expected reverse-order damage targets p3->p2, got %+v", game.State.PendingDamageQueue)
	}
}

func TestSpiritCasterIncantation_CapBlocksPromptAndResolvesWind(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "SpiritCaster", "spirit_caster", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "EnemyA", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "EnemyB", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p3 := game.State.Players["p3"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		spiritCasterTestCard("w1", "风符", model.CardTypeMagic, model.ElementWind),
	}
	p2.Hand = []model.Card{
		spiritCasterTestCard("a1", "攻击A1", model.CardTypeAttack, model.ElementFire),
		spiritCasterTestCard("a1x", "攻击A2", model.CardTypeAttack, model.ElementThunder),
	}
	p3.Hand = []model.Card{
		spiritCasterTestCard("b1", "攻击B1", model.CardTypeAttack, model.ElementWater),
		spiritCasterTestCard("b2", "攻击B2", model.CardTypeAttack, model.ElementWind),
	}
	addSpiritCasterPowerForTest(p1, spiritCasterTestCard("pow1", "妖力1", model.CardTypeMagic, model.ElementFire))
	addSpiritCasterPowerForTest(p1, spiritCasterTestCard("pow2", "妖力2", model.CardTypeMagic, model.ElementThunder))

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection
	if err := game.UseSkill("p1", "sc_talisman_wind", []string{"p2", "p3"}, []int{0}); err != nil {
		t.Fatalf("use talisman wind failed: %v", err)
	}
	game.processDeferredFollowups()

	// 念咒满层不会弹念咒提示，直接进入“由目标自行选择弃牌”流程。
	requireChoicePrompt(t, game, "p3", "sc_talisman_wind_discard")
	if err := game.handleWeakChoiceInput("p3", 1); err != nil { // p3 弃第2张
		t.Fatalf("p3 choose discard failed: %v", err)
	}
	requireChoicePrompt(t, game, "p2", "sc_talisman_wind_discard")
	if err := game.handleWeakChoiceInput("p2", 0); err != nil { // p2 弃第1张
		t.Fatalf("p2 choose discard failed: %v", err)
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected wind flow completed, got pending interrupt %+v", game.State.PendingInterrupt)
	}
	if got := spiritCasterPowerCount(p1, ""); got != 2 {
		t.Fatalf("expected power count keep at cap=2, got %d", got)
	}
	if got := len(p2.Hand); got != 1 {
		t.Fatalf("expected p2 discarded exactly 1 card, hand=%d", got)
	}
	if got := len(p3.Hand); got != 1 {
		t.Fatalf("expected p3 discarded exactly 1 card, hand=%d", got)
	}
}

func TestSpiritCasterHundredNight_FireRevealAOEWithCollapse(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "SpiritCaster", "spirit_caster", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "EnemyA", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "EnemyB", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p4", "EnemyC", "priest", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.Crystal = 1
	addSpiritCasterPowerForTest(p1, spiritCasterTestCard("pow_fire", "火妖力", model.CardTypeMagic, model.ElementFire))

	ctx := game.buildContext(p1, p2, model.TriggerOnAttackHit, &model.EventContext{
		Type:     model.EventAttack,
		SourceID: "p1",
		TargetID: "p2",
		AttackInfo: &model.AttackEventInfo{
			ActionType:       string(model.ActionAttack),
			CounterInitiator: "",
		},
	})
	h := &skills.SpiritCasterHundredNightHandler{}
	if !h.CanUse(ctx) {
		t.Fatalf("expected hundred-night available with fire power")
	}
	if err := h.Execute(ctx); err != nil {
		t.Fatalf("execute hundred-night failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "sc_hundred_night_power")

	if err := game.handleWeakChoiceInput("p1", 0); err != nil { // 选火妖力
		t.Fatalf("choose power failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "sc_hundred_night_fire_reveal")
	if err := game.handleWeakChoiceInput("p1", 0); err != nil { // 展示并走AOE
		t.Fatalf("choose reveal failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "sc_hundred_night_exclude_pick")
	if err := game.handleWeakChoiceInput("p1", 0); err != nil { // 排除 p1
		t.Fatalf("pick first excluded target failed: %v", err)
	}
	if err := game.handleWeakChoiceInput("p1", 0); err != nil { // 排除 p2（此时索引重排后仍是0）
		t.Fatalf("pick second excluded target failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "sc_spiritual_collapse_confirm")
	if err := game.handleWeakChoiceInput("p1", 0); err != nil { // 发动灵力崩解
		t.Fatalf("confirm spiritual collapse failed: %v", err)
	}

	if p1.Crystal != 0 {
		t.Fatalf("expected crystal consumed by spiritual collapse, got %d", p1.Crystal)
	}
	if len(game.State.PendingDamageQueue) != 2 {
		t.Fatalf("expected aoe damage to 2 remaining players, got %d", len(game.State.PendingDamageQueue))
	}
	for _, pd := range game.State.PendingDamageQueue {
		if pd.TargetID != "p3" && pd.TargetID != "p4" {
			t.Fatalf("unexpected aoe target: %+v", pd)
		}
		if pd.Damage != 2 {
			t.Fatalf("expected damage=2 with collapse bonus, got %+v", pd)
		}
	}
}

func TestSpiritCasterHundredNight_NonFireSingleTarget(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "SpiritCaster", "spirit_caster", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "EnemyA", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "EnemyB", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	addSpiritCasterPowerForTest(p1, spiritCasterTestCard("pow_w", "水妖力", model.CardTypeMagic, model.ElementWater))

	ctx := game.buildContext(p1, p2, model.TriggerOnAttackHit, &model.EventContext{
		Type:     model.EventAttack,
		SourceID: "p1",
		TargetID: "p2",
		AttackInfo: &model.AttackEventInfo{
			ActionType:       string(model.ActionAttack),
			CounterInitiator: "",
		},
	})
	h := &skills.SpiritCasterHundredNightHandler{}
	if !h.CanUse(ctx) {
		t.Fatalf("expected hundred-night available with non-fire power")
	}
	if err := h.Execute(ctx); err != nil {
		t.Fatalf("execute hundred-night failed: %v", err)
	}
	if err := game.handleWeakChoiceInput("p1", 0); err != nil { // 选水妖力
		t.Fatalf("choose power failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "sc_hundred_night_target")
	if err := game.handleWeakChoiceInput("p1", 2); err != nil { // 目标选 p3
		t.Fatalf("choose target failed: %v", err)
	}
	if len(game.State.PendingDamageQueue) != 1 {
		t.Fatalf("expected one pending damage, got %d", len(game.State.PendingDamageQueue))
	}
	pd := game.State.PendingDamageQueue[0]
	if pd.TargetID != "p3" || pd.Damage != 1 || pd.DamageType != "magic" {
		t.Fatalf("unexpected pending damage: %+v", pd)
	}
}
