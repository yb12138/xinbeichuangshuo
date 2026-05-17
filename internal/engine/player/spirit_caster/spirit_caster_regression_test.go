package spirit_caster_test

import (
	"starcup-engine/internal/engine"
	"starcup-engine/internal/testutils"
	"strings"
	"testing"

	"starcup-engine/internal/data"
	spiritcasterplayer "starcup-engine/internal/engine/player/spirit_caster"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func countSpiritCasterDamageEvents(obs *testutils.CaptureObserver, sourceID, targetID string, damage int, damageType model.DamageType) int {
	if obs == nil {
		return 0
	}
	count := 0
	for _, ev := range obs.Events {
		if ev.Type != model.EventDamageDealt {
			continue
		}
		payload, ok := ev.Data.(model.DamageDealtPayload)
		if !ok {
			continue
		}
		if sourceID != "" && payload.SourceID != sourceID {
			continue
		}
		if targetID != "" && payload.TargetID != targetID {
			continue
		}
		if damage > 0 && payload.Damage != damage {
			continue
		}
		if damageType != "" && !strings.EqualFold(payload.DamageType, string(damageType)) {
			continue
		}
		count++
	}
	return count
}

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
	spiritcasterplayer.SyncPowerToken(p)
}

func TestSpiritCasterTalismanThunder_SealThenIncantThenDamage(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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
		Hook:     model.FieldHookManual,
	})

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	if err := game.UseSkill("p1", "sc_talisman_thunder", []string{"p2", "p3"}, []int{0}); err != nil {
		t.Fatalf("use talisman thunder failed: %v", err)
	}
	if len(game.State.PendingDamageQueue) != 1 {
		t.Fatalf("expected only seal damage pending first, got %d", len(game.State.PendingDamageQueue))
	}
	if !game.HasSkillResume() {
		t.Fatalf("expected pending talisman skill resume")
	}

	// 先结算封印伤害，再继续灵符后续。
	if paused := game.ProcessPendingDamages(); paused {
		t.Fatalf("unexpected interrupt while resolving seal damage")
	}
	if len(game.State.PendingDamageQueue) != 0 {
		t.Fatalf("expected seal damage queue consumed")
	}
	game.ProcessPendingSkillResume()
	testutils.RequireChoicePrompt(t, game, "p1", "sc_incant_confirm")

	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil { // 发动念咒
		t.Fatalf("confirm incantation failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "sc_incant_card")
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil { // 盖放补牌
		t.Fatalf("choose incantation card failed: %v", err)
	}

	if got := spiritcasterplayer.PowerCount(p1, ""); got != 1 {
		t.Fatalf("expected 1 spirit power after incantation, got %d", got)
	}
	if len(game.State.PendingDamageQueue) != 2 {
		t.Fatalf("expected thunder damage queued for 2 targets, got %d", len(game.State.PendingDamageQueue))
	}
	if game.State.PendingDamageQueue[0].TargetID != "p3" || game.State.PendingDamageQueue[1].TargetID != "p2" {
		t.Fatalf("expected reverse-order damage targets p3->p2, got %+v", game.State.PendingDamageQueue)
	}
}

func TestSpiritCasterIncantation_NoCapStillPromptsAndResolvesWind(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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
		spiritCasterTestCard("w1", "风符", model.CardTypeMagic, model.ElementWind),  // 发动成本
		spiritCasterTestCard("h1", "补牌", model.CardTypeAttack, model.ElementFire), // 念咒盖放
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
	game.State.TurnStage = model.TurnStageActionExecution
	if err := game.UseSkill("p1", "sc_talisman_wind", []string{"p2", "p3"}, []int{0}); err != nil {
		t.Fatalf("use talisman wind failed: %v", err)
	}
	game.ProcessPendingSkillResume()

	testutils.RequireChoicePrompt(t, game, "p1", "sc_incant_confirm")
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("confirm incantation failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "sc_incant_card")
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("choose incantation card failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p3", "sc_talisman_wind_discard")
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p3", Selections: []int{1}}); err != nil { // p3 弃第2张
		t.Fatalf("p3 choose discard failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p2", "sc_talisman_wind_discard")
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p2", Selections: []int{0}}); err != nil { // p2 弃第1张
		t.Fatalf("p2 choose discard failed: %v", err)
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected wind flow completed, got pending interrupt %+v", game.State.PendingInterrupt)
	}
	if got := spiritcasterplayer.PowerCount(p1, ""); got != 3 {
		t.Fatalf("expected incantation to add a third spirit power, got %d", got)
	}
	if got := len(p2.Hand); got != 1 {
		t.Fatalf("expected p2 discarded exactly 1 card, hand=%d", got)
	}
	if got := len(p3.Hand); got != 1 {
		t.Fatalf("expected p3 discarded exactly 1 card, hand=%d", got)
	}
}

func TestSpiritCasterHundredNight_FireRevealAOEWithCollapse(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
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

	game.State.Deck = rules.InitDeck()
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.Crystal = 1
	addSpiritCasterPowerForTest(p1, spiritCasterTestCard("pow_fire", "火妖力", model.CardTypeMagic, model.ElementFire))

	ctx := game.BuildContext(p1, p2, model.TimingOnHitCheck, &model.EventContext{
		Type:     model.EventAttack,
		SourceID: "p1",
		TargetID: "p2",
		AttackInfo: &model.AttackEventInfo{
			ActionType:       string(model.ActionAttack),
			IsHit:            true,
			CounterInitiator: "",
		},
	})
	h := &spiritcasterplayer.SpiritCasterHundredNightHandler{}
	if !h.CanUse(ctx) {
		t.Fatalf("expected hundred-night available with fire power")
	}
	if err := h.Execute(ctx); err != nil {
		t.Fatalf("execute hundred-night failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "sc_hundred_night_power")

	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil { // 选火妖力
		t.Fatalf("choose power failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "sc_hundred_night_fire_reveal")
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil { // 展示并走AOE
		t.Fatalf("choose reveal failed: %v", err)
	}
	if reveal := testutils.FindPublicDiscardReveal(obs, "p1"); reveal == nil {
		t.Fatalf("expected revealed fire spirit power to emit a public discard reveal event")
	}
	testutils.RequireChoicePrompt(t, game, "p1", "sc_hundred_night_exclude_pick")
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil { // 排除 p1
		t.Fatalf("pick first excluded target failed: %v", err)
	}
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil { // 排除 p2（此时索引重排后仍是0）
		t.Fatalf("pick second excluded target failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "sc_spiritual_collapse_confirm")
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil { // 发动灵力崩解
		t.Fatalf("confirm spiritual collapse failed: %v", err)
	}

	if p1.Crystal != 0 {
		t.Fatalf("expected crystal consumed by spiritual collapse, got %d", p1.Crystal)
	}
	if len(game.State.PendingDamageQueue) != 2 {
		t.Fatalf("expected aoe damage to 2 remaining players, got %d", len(game.State.PendingDamageQueue))
	}
	seenTargets := map[string]bool{}
	for _, pd := range game.State.PendingDamageQueue {
		if pd.TargetID != "p3" && pd.TargetID != "p4" {
			t.Fatalf("unexpected aoe target: %+v", pd)
		}
		if seenTargets[pd.TargetID] {
			t.Fatalf("aoe should not enqueue duplicate target damage, got %+v", game.State.PendingDamageQueue)
		}
		seenTargets[pd.TargetID] = true
		if pd.Damage != 2 {
			t.Fatalf("expected damage=2 with collapse bonus, got %+v", pd)
		}
	}
}

func TestSpiritCasterHundredNight_FireRevealAOEWithCollapseResolvesEachTargetOnce(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
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

	game.State.Deck = rules.InitDeck()
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p3 := game.State.Players["p3"]
	p4 := game.State.Players["p4"]
	p1.Crystal = 1
	p3.Heal = 0
	p4.Heal = 0
	addSpiritCasterPowerForTest(p1, spiritCasterTestCard("pow_fire", "火妖力", model.CardTypeMagic, model.ElementFire))

	ctx := game.BuildContext(p1, p2, model.TimingOnHitCheck, &model.EventContext{
		Type:     model.EventAttack,
		SourceID: "p1",
		TargetID: "p2",
		AttackInfo: &model.AttackEventInfo{
			ActionType: string(model.ActionAttack),
			IsHit:      true,
		},
	})
	h := &spiritcasterplayer.SpiritCasterHundredNightHandler{}
	if !h.CanUse(ctx) {
		t.Fatalf("expected hundred-night available with fire power")
	}
	if err := h.Execute(ctx); err != nil {
		t.Fatalf("execute hundred-night failed: %v", err)
	}
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("choose power failed: %v", err)
	}
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("choose fire reveal failed: %v", err)
	}
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("pick first excluded target failed: %v", err)
	}
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("pick second excluded target failed: %v", err)
	}
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("confirm spiritual collapse failed: %v", err)
	}

	if len(game.State.PendingDamageQueue) != 2 {
		t.Fatalf("expected two pending AOE damages, got %+v", game.State.PendingDamageQueue)
	}
	game.Drive()

	if got := countSpiritCasterDamageEvents(obs, "p1", "p3", 2, model.MagicAttack); got != 1 {
		t.Fatalf("expected exactly one 2-damage event for p3, got %d", got)
	}
	if got := countSpiritCasterDamageEvents(obs, "p1", "p4", 2, model.MagicAttack); got != 1 {
		t.Fatalf("expected exactly one 2-damage event for p4, got %d", got)
	}
	if got := countSpiritCasterDamageEvents(obs, "p1", "p2", 2, model.MagicAttack); got != 0 {
		t.Fatalf("excluded p2 should not take collapse AOE damage, got %d events", got)
	}
	if len(game.State.PendingDamageQueue) != 0 || game.State.PendingInterrupt != nil {
		t.Fatalf("expected damage queue fully resolved, queue=%d interrupt=%+v", len(game.State.PendingDamageQueue), game.State.PendingInterrupt)
	}
	if got := len(p3.Hand); got != 2 {
		t.Fatalf("expected p3 to draw exactly 2 cards from damage, got %d", got)
	}
	if got := len(p4.Hand); got != 2 {
		t.Fatalf("expected p4 to draw exactly 2 cards from damage, got %d", got)
	}
}

func TestSpiritCasterHundredNight_NonFireSingleTarget(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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

	ctx := game.BuildContext(p1, p2, model.TimingOnHitCheck, &model.EventContext{
		Type:     model.EventAttack,
		SourceID: "p1",
		TargetID: "p2",
		AttackInfo: &model.AttackEventInfo{
			ActionType:       string(model.ActionAttack),
			IsHit:            true,
			CounterInitiator: "",
		},
	})
	h := &spiritcasterplayer.SpiritCasterHundredNightHandler{}
	if !h.CanUse(ctx) {
		t.Fatalf("expected hundred-night available with non-fire power")
	}
	if err := h.Execute(ctx); err != nil {
		t.Fatalf("execute hundred-night failed: %v", err)
	}
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil { // 选水妖力
		t.Fatalf("choose power failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "sc_hundred_night_target")
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{2}}); err != nil { // 目标选 p3
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

func TestSpiritCasterConfig_MetadataAlignsWithDocument(t *testing.T) {
	characters := data.GetCharacters()
	var spiritCaster *model.Character
	for _, character := range characters {
		if character.ID == "spirit_caster" {
			copy := character
			spiritCaster = &copy
			break
		}
	}
	if spiritCaster == nil {
		t.Fatalf("spirit_caster character not found")
	}

	var incantation *model.SkillDefinition
	var hundredNight *model.SkillDefinition
	var collapse *model.SkillDefinition
	for i := range spiritCaster.Skills {
		switch spiritCaster.Skills[i].ID {
		case "sc_incantation":
			incantation = &spiritCaster.Skills[i]
		case "sc_hundred_night":
			hundredNight = &spiritCaster.Skills[i]
		case "sc_spiritual_collapse":
			collapse = &spiritCaster.Skills[i]
		}
	}
	if incantation == nil || hundredNight == nil || collapse == nil {
		t.Fatalf("expected incantation, hundred night, and spiritual collapse skills present")
	}
	if hundredNight.TargetType != model.TargetAny || hundredNight.MinTargets != 1 || hundredNight.MaxTargets != 2 {
		t.Fatalf("expected hundred night target metadata any(1..2), got type=%v min=%d max=%d", hundredNight.TargetType, hundredNight.MinTargets, hundredNight.MaxTargets)
	}
	if collapse.CostCrystal != 1 {
		t.Fatalf("expected spiritual collapse crystal cost=1, got %d", collapse.CostCrystal)
	}
	if incantation.Description == "" || containsSpiritCasterCapText(incantation.Description) {
		t.Fatalf("expected incantation description to omit legacy cap text, got %q", incantation.Description)
	}
}

func containsSpiritCasterCapText(desc string) bool {
	return len(desc) > 0 && (strings.Contains(desc, "上限2") || strings.Contains(desc, "上限 2"))
}

func TestSpiritCasterThunderCollapse_TwoTargetsBothDamaged(t *testing.T) {
	// Debugging: checking the heal interrupt flow issue
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
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
	p1.Crystal = 1 // 有水晶用于灵力崩解
	p1.Hand = []model.Card{
		spiritCasterTestCard("t1", "雷符", model.CardTypeMagic, model.ElementThunder), // 发动成本
	}
	p2.Heal = 3
	p3.Heal = 3

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	// 发动雷鸣技能，目标 p2 和 p3
	if err := game.UseSkill("p1", "sc_talisman_thunder", []string{"p2", "p3"}, []int{0}); err != nil {
		t.Fatalf("use talisman thunder failed: %v", err)
	}

	game.Drive()
	// 无手牌，应该直接弹出灵力崩解选择
	testutils.RequireChoicePrompt(t, game, "p1", "sc_spiritual_collapse_confirm")

	// 选择发动灵力崩解（选项0 = 是）
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("confirm spiritual collapse failed: %v", err)
	}

	// 伤害队列应该有2个伤害（逆序：p3先，p2后）
	if len(game.State.PendingDamageQueue) != 2 {
		t.Fatalf("expected 2 pending damages, got %d", len(game.State.PendingDamageQueue))
	}
	t.Logf("Pending damage queue: %+v", game.State.PendingDamageQueue)

	// 结算所有伤害
	game.Drive()
	t.Logf("Before first ProcessPendingDamages, queue length: %d", len(game.State.PendingDamageQueue))
	for i, pd := range game.State.PendingDamageQueue {
		t.Logf("  Queue[%d]: TargetID=%s, Damage=%d, HealResolved=%v", i, pd.TargetID, pd.Damage, pd.HealResolved)
	}

	if paused := game.ProcessPendingDamages(); paused {
		// 检查是什么中断
		if game.State.PendingInterrupt != nil {
			t.Logf("First interrupt generated: %+v", game.State.PendingInterrupt)
			ctxData, ok := game.State.PendingInterrupt.Context.(map[string]interface{})
			if !ok {
				t.Fatalf("unexpected interrupt context type")
			}
			choiceType, _ := ctxData["choice_type"].(string)
			if choiceType == "heal" {
				// 选择使用全部治疗抵消伤害（选择2表示使用2点治疗）
				if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: game.State.PendingInterrupt.PlayerID, Selections: []int{2}}); err != nil {
					t.Fatalf("handle first heal interrupt failed: %v", err)
				}
				t.Logf("After first heal choice handled, queue length: %d", len(game.State.PendingDamageQueue))
				for i, pd := range game.State.PendingDamageQueue {
					t.Logf("  Queue[%d]: TargetID=%s, Damage=%d, HealResolved=%v", i, pd.TargetID, pd.Damage, pd.HealResolved)
				}
				// 继续结算：Drive 会处理 p3 的伤害，然后为 p2 推入治疗中断
				game.Drive()
				// p2 应该有治疗中断
				if game.State.PendingInterrupt == nil {
					t.Fatalf("expected heal interrupt for p2 after Drive")
				}
				p2Ctx, ok := game.State.PendingInterrupt.Context.(map[string]interface{})
				if !ok {
					t.Fatalf("unexpected p2 interrupt context type")
				}
				p2ChoiceType, _ := p2Ctx["choice_type"].(string)
				if p2ChoiceType != "heal" {
					t.Fatalf("expected heal interrupt for p2, got %s", p2ChoiceType)
				}
				// 处理 p2 的治疗选择（使用2点治疗抵消伤害）
				if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p2", Selections: []int{2}}); err != nil {
					t.Fatalf("handle second heal interrupt failed: %v", err)
				}
				t.Logf("After second heal choice handled, queue length: %d", len(game.State.PendingDamageQueue))
				// 继续结算
				game.Drive()
			} else {
				t.Fatalf("unexpected interrupt type: %s", choiceType)
			}
		} else {
			t.Fatalf("paused but no pending interrupt")
		}
	}

	t.Logf("After all damages resolved, queue length: %d", len(game.State.PendingDamageQueue))

	// 验证两个目标都受到了伤害（灵力崩解+1，总共2点）
	if p2.Heal != 1 {
		t.Fatalf("expected p2 heal=1 (3-2), got %d", p2.Heal)
	}
	if p3.Heal != 1 {
		t.Fatalf("expected p3 heal=1 (3-2), got %d", p3.Heal)
	}
	t.Logf("After damage: p2.Heal=%d, p3.Heal=%d", p2.Heal, p3.Heal)

	// 验证水晶被消耗
	if p1.Crystal != 0 {
		t.Fatalf("expected crystal consumed, got %d", p1.Crystal)
	}
}
