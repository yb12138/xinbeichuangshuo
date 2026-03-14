package engine

import (
	"strings"
	"testing"

	"starcup-engine/internal/engine/skills"
	"starcup-engine/internal/model"
)

func magicBowTestCard(id, name string, cardType model.CardType, ele model.Element) model.Card {
	return model.Card{
		ID:          id,
		Name:        name,
		Type:        cardType,
		Element:     ele,
		Damage:      2,
		Description: name,
	}
}

func giveMagicBowCharges(p *model.Player, elements ...model.Element) {
	cards := make([]model.Card, 0, len(elements))
	for i, ele := range elements {
		cards = append(cards, magicBowTestCard(
			"mb_charge_"+string(rune('a'+i)),
			"充能"+string(rune('A'+i)),
			model.CardTypeAttack,
			ele,
		))
	}
	addMagicBowChargeCards(p, cards)
}

func pendingChoiceTargetIDs(intr *model.Interrupt) []string {
	if intr == nil {
		return nil
	}
	data, ok := intr.Context.(map[string]interface{})
	if !ok {
		return nil
	}
	var out []string
	if arr, ok := data["target_ids"].([]string); ok {
		out = append(out, arr...)
		return out
	}
	if arr, ok := data["target_ids"].([]interface{}); ok {
		for _, v := range arr {
			if s, ok := v.(string); ok {
				out = append(out, s)
			}
		}
	}
	return out
}

func TestMagicBowMagicPierce_MissDealsMagicDamageAndLocksMultiShot(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "MagicBow", "magic_bow", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Defender", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy2", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		magicBowTestCard("atk1", "火焰斩", model.CardTypeAttack, model.ElementFire),
	}
	// 防御方准备【圣光】触发未命中分支。
	p2.Hand = []model.Card{
		magicBowTestCard("def1", "圣光", model.CardTypeMagic, model.ElementLight),
	}
	giveMagicBowCharges(p1, model.ElementFire, model.ElementWind)

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})
	chooseResponseSkillByID(t, game, "p1", "mb_magic_pierce")

	if err := game.handleCombatResponse(model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		CardIndex: 0,
		ExtraArgs: []string{"defend"},
	}); err != nil {
		t.Fatalf("combat defend response failed: %v", err)
	}

	if len(game.State.PendingDamageQueue) == 0 {
		t.Fatalf("expected pending magic damage from magic pierce miss")
	}
	pd := game.State.PendingDamageQueue[0]
	if pd.SourceID != "p1" || pd.TargetID != "p2" || pd.Damage != 3 || pd.DamageType != "magic" {
		t.Fatalf("unexpected pending damage: %+v", pd)
	}
	if got := p1.Tokens["mb_magic_pierce_pending"]; got != 0 {
		t.Fatalf("expected mb_magic_pierce_pending cleared, got %d", got)
	}
	if got := p1.TurnState.UsedSkillCounts["mb_magic_pierce_used_turn"]; got != 1 {
		t.Fatalf("expected magic pierce used mark=1, got %d", got)
	}

	multiShotCtx := game.buildContext(p1, nil, model.TriggerOnPhaseEnd, &model.EventContext{
		Type:       model.EventPhaseEnd,
		SourceID:   "p1",
		ActionType: model.ActionAttack,
		AttackInfo: &model.AttackEventInfo{
			ActionType:       string(model.ActionAttack),
			CounterInitiator: "",
		},
	})
	if (&skills.MagicBowMultiShotHandler{}).CanUse(multiShotCtx) {
		t.Fatalf("expected multi-shot disabled after using magic pierce in same turn")
	}
}

func TestMagicBowMultiShot_TargetCannotRepeatPrevious(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "MagicBow", "magic_bow", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "EnemyA", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "EnemyB", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.TurnState = model.NewPlayerTurnState()
	// 玩家顺序 p1,p2,p3；上一次攻击目标为 p2（序号2）。
	p1.TurnState.UsedSkillCounts["mb_last_attack_target_order"] = 2
	giveMagicBowCharges(p1, model.ElementWind)

	ctx := game.buildContext(p1, nil, model.TriggerOnPhaseEnd, &model.EventContext{
		Type:       model.EventPhaseEnd,
		SourceID:   "p1",
		ActionType: model.ActionAttack,
		AttackInfo: &model.AttackEventInfo{
			ActionType:       string(model.ActionAttack),
			CounterInitiator: "",
		},
	})
	h := &skills.MagicBowMultiShotHandler{}
	if !h.CanUse(ctx) {
		t.Fatalf("expected multi-shot usable with wind charge and valid alternate target")
	}
	if err := h.Execute(ctx); err != nil {
		t.Fatalf("execute multi-shot failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "mb_multi_shot_target")

	targetIDs := pendingChoiceTargetIDs(game.State.PendingInterrupt)
	if len(targetIDs) != 1 || targetIDs[0] != "p3" {
		t.Fatalf("expected only p3 as valid target, got %v", targetIDs)
	}

	if err := game.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("choose multi-shot target failed: %v", err)
	}
	if len(game.State.ActionQueue) != 1 {
		t.Fatalf("expected queued extra attack, got %d", len(game.State.ActionQueue))
	}
	qa := game.State.ActionQueue[0]
	if qa.SourceSkill != "mb_multi_shot" || qa.TargetID != "p3" {
		t.Fatalf("unexpected queued action: %+v", qa)
	}
	if qa.Card == nil || qa.Card.Element != model.ElementDark || qa.Card.Damage != 1 {
		t.Fatalf("expected virtual dark attack damage=1, got %+v", qa.Card)
	}
	if got := magicBowChargeCount(p1, ""); got != 0 {
		t.Fatalf("expected wind charge consumed, remaining=%d", got)
	}
}

func TestMagicBowCharge_FollowupPlaceCharges(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "MagicBow", "magic_bow", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1
	p1.Hand = []model.Card{
		magicBowTestCard("h1", "火焰斩", model.CardTypeAttack, model.ElementFire),
		magicBowTestCard("h2", "水涟斩", model.CardTypeAttack, model.ElementWater),
		magicBowTestCard("h3", "雷光斩", model.CardTypeAttack, model.ElementThunder),
		magicBowTestCard("h4", "风神斩", model.CardTypeAttack, model.ElementWind),
	}
	game.State.Deck = []model.Card{
		magicBowTestCard("d1", "补牌1", model.CardTypeAttack, model.ElementFire),
		magicBowTestCard("d2", "补牌2", model.CardTypeMagic, model.ElementThunder),
		magicBowTestCard("d3", "补牌3", model.CardTypeAttack, model.ElementWater),
	}

	ctx := game.buildContext(p1, nil, model.TriggerOnTurnStart, &model.EventContext{
		Type:     model.EventTurnStart,
		SourceID: "p1",
	})
	if err := (&skills.MagicBowChargeHandler{}).Execute(ctx); err != nil {
		t.Fatalf("execute charge failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "mb_charge_draw_x")

	if err := game.handleWeakChoiceInput("p1", 2); err != nil {
		t.Fatalf("choose charge draw x failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "mb_charge_place_count")

	if err := game.handleWeakChoiceInput("p1", 2); err != nil {
		t.Fatalf("choose charge place count failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "mb_charge_place_cards")
	if err := game.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("choose first charge card failed: %v", err)
	}
	if err := game.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("choose second charge card failed: %v", err)
	}

	if got := magicBowChargeCount(p1, ""); got != 2 {
		t.Fatalf("expected 2 charges placed, got %d", got)
	}
	if got := p1.Tokens["mb_charge_count"]; got != 2 {
		t.Fatalf("expected token mb_charge_count=2, got %d", got)
	}
	if got := len(p1.Hand); got != 4 {
		t.Fatalf("expected hand size back to 4 after draw2/place2, got %d", got)
	}
	if got := p1.Crystal; got != 0 {
		t.Fatalf("expected charge consumed 1 crystal, got %d", got)
	}
	if got := p1.TurnState.UsedSkillCounts["mb_charge_lock_turn"]; got != 1 {
		t.Fatalf("expected mb_charge_lock_turn=1, got %d", got)
	}
}

func TestMagicBowCharge_DiscardFirstThenChooseX(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "MagicBow", "magic_bow", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1
	p1.Hand = []model.Card{
		magicBowTestCard("h1", "火焰斩", model.CardTypeAttack, model.ElementFire),
		magicBowTestCard("h2", "水涟斩", model.CardTypeAttack, model.ElementWater),
		magicBowTestCard("h3", "雷光斩", model.CardTypeAttack, model.ElementThunder),
		magicBowTestCard("h4", "风神斩", model.CardTypeAttack, model.ElementWind),
		magicBowTestCard("h5", "圣光", model.CardTypeMagic, model.ElementLight),
		magicBowTestCard("h6", "魔弹", model.CardTypeMagic, model.ElementDark),
	}

	ctx := game.buildContext(p1, nil, model.TriggerOnTurnStart, &model.EventContext{
		Type:     model.EventTurnStart,
		SourceID: "p1",
	})
	if err := (&skills.MagicBowChargeHandler{}).Execute(ctx); err != nil {
		t.Fatalf("execute charge failed: %v", err)
	}

	// 新规则：先弃到4张，再让玩家选择X。
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptDiscard {
		t.Fatalf("expected discard interrupt first, got %+v", game.State.PendingInterrupt)
	}
	data, _ := game.State.PendingInterrupt.Context.(map[string]interface{})
	minNeed, _ := data["min"].(int)
	maxNeed, _ := data["max"].(int)
	if minNeed != 2 || maxNeed != 2 {
		t.Fatalf("expected forced discard min/max=2 before choosing x, got min=%v max=%v", data["min"], data["max"])
	}

	if err := game.ConfirmDiscard("p1", []int{4, 5}); err != nil {
		t.Fatalf("discard to 4 for charge failed: %v", err)
	}
	if got := len(p1.Hand); got != 4 {
		t.Fatalf("expected hand size=4 after forced discard, got %d", got)
	}
	requireChoicePrompt(t, game, "p1", "mb_charge_draw_x")
}

func TestMagicBowCharge_DrawOverflowMoraleLossWithoutDiscard(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "MagicBow", "magic_bow", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1
	p1.Hand = []model.Card{
		magicBowTestCard("h1", "火焰斩", model.CardTypeAttack, model.ElementFire),
		magicBowTestCard("h2", "水涟斩", model.CardTypeAttack, model.ElementWater),
		magicBowTestCard("h3", "雷光斩", model.CardTypeAttack, model.ElementThunder),
		magicBowTestCard("h4", "风神斩", model.CardTypeAttack, model.ElementWind),
	}
	game.State.Deck = []model.Card{
		magicBowTestCard("d1", "补牌1", model.CardTypeAttack, model.ElementFire),
		magicBowTestCard("d2", "补牌2", model.CardTypeMagic, model.ElementThunder),
		magicBowTestCard("d3", "补牌3", model.CardTypeAttack, model.ElementWater),
		magicBowTestCard("d4", "补牌4", model.CardTypeMagic, model.ElementEarth),
	}
	redMoraleBefore := game.State.RedMorale

	ctx := game.buildContext(p1, nil, model.TriggerOnTurnStart, &model.EventContext{
		Type:     model.EventTurnStart,
		SourceID: "p1",
	})
	if err := (&skills.MagicBowChargeHandler{}).Execute(ctx); err != nil {
		t.Fatalf("execute charge failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "mb_charge_draw_x")

	// 选择X=4，4->8，默认上限6，爆士气2但不弃牌。
	if err := game.handleWeakChoiceInput("p1", 4); err != nil {
		t.Fatalf("choose charge draw x failed: %v", err)
	}
	if got := game.State.RedMorale; got != redMoraleBefore-2 {
		t.Fatalf("expected red morale -2 after overflow draw, before=%d after=%d", redMoraleBefore, got)
	}
	if got := len(p1.Hand); got != 8 {
		t.Fatalf("expected no discard after overflow draw, hand should stay 8, got %d", got)
	}
	if game.State.PendingInterrupt == nil || choiceTypeOfInterrupt(game.State.PendingInterrupt) != "mb_charge_place_count" {
		t.Fatalf("expected enter place-count choice after draw, got %+v", game.State.PendingInterrupt)
	}
	if game.State.PendingInterrupt.Type == model.InterruptDiscard {
		t.Fatalf("should not open discard interrupt after charge overflow draw")
	}

	if err := game.handleWeakChoiceInput("p1", 4); err != nil {
		t.Fatalf("choose place count=4 failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "mb_charge_place_cards")
	for i := 0; i < 4; i++ {
		if err := game.handleWeakChoiceInput("p1", 0); err != nil {
			t.Fatalf("choose charge place card %d failed: %v", i+1, err)
		}
	}
	if got := magicBowChargeCount(p1, ""); got != 4 {
		t.Fatalf("expected 4 charges placed, got %d", got)
	}
}

func TestMagicBowThunderScatter_ExtraDamageSplit(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "MagicBow", "magic_bow", model.RedCamp); err != nil {
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
	giveMagicBowCharges(p1, model.ElementThunder, model.ElementThunder, model.ElementThunder)

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	if err := game.UseSkill("p1", "mb_thunder_scatter", nil, nil); err != nil {
		t.Fatalf("use thunder scatter failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "mb_thunder_scatter_extra")

	// 选择额外移除2个雷系充能。
	if err := game.handleWeakChoiceInput("p1", 2); err != nil {
		t.Fatalf("choose thunder scatter extra failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "mb_thunder_scatter_target")

	// 目标列表应为 [p2,p3]，选择第一个目标 p2。
	if err := game.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("choose thunder scatter target failed: %v", err)
	}

	if got := magicBowChargeCount(p1, ""); got != 0 {
		t.Fatalf("expected all thunder charges consumed, remaining=%d", got)
	}
	if len(game.State.PendingDamageQueue) != 3 {
		t.Fatalf("expected 3 pending magic damages, got %d", len(game.State.PendingDamageQueue))
	}

	totalToP2 := 0
	totalToP3 := 0
	for _, pd := range game.State.PendingDamageQueue {
		if pd.DamageType != "magic" || pd.SourceID != "p1" {
			t.Fatalf("unexpected pending damage item: %+v", pd)
		}
		switch pd.TargetID {
		case "p2":
			totalToP2 += pd.Damage
		case "p3":
			totalToP3 += pd.Damage
		default:
			t.Fatalf("unexpected target in pending damage: %+v", pd)
		}
	}
	if totalToP2 != 3 || totalToP3 != 1 {
		t.Fatalf("unexpected thunder scatter split damage, p2=%d p3=%d", totalToP2, totalToP3)
	}
}

func TestMagicBowMagicPierce_HitBonusCappedAtTwo(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "MagicBow", "magic_bow", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Defender", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		magicBowTestCard("atk_hit", "烈焰箭", model.CardTypeAttack, model.ElementFire),
	}
	// 预置3个火系充能，命中追加只应消耗1个（总共最多+2伤害）。
	giveMagicBowCharges(p1, model.ElementFire, model.ElementFire, model.ElementFire)

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})
	chooseResponseSkillByID(t, game, "p1", "mb_magic_pierce")
	if err := game.handleCombatResponse(model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	}); err != nil {
		t.Fatalf("combat take response failed: %v", err)
	}
	// 命中追加询问在“待结算伤害”阶段触发，需要推进主循环一次。
	game.Drive()
	requireChoicePrompt(t, game, "p1", "mb_magic_pierce_hit_confirm")
	if ctxData, ok := game.State.PendingInterrupt.Context.(map[string]interface{}); ok {
		src, _ := ctxData["source_id"].(string)
		dst, _ := ctxData["target_id"].(string)
		if src != "p1" || dst != "p2" {
			t.Fatalf("unexpected hit-confirm context source/target: src=%q dst=%q", src, dst)
		}
	}
	if err := game.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("confirm hit bonus failed: %v", err)
	}

	if got := magicBowChargeCount(p1, model.ElementFire); got != 1 {
		t.Fatalf("expected remain 1 fire charge after at-most-once hit bonus, got %d", got)
	}
	if got := p1.Tokens["mb_magic_pierce_pending"]; got != 0 {
		t.Fatalf("expected mb_magic_pierce_pending cleared, got %d", got)
	}

	totalAttackToP2 := 0
	for _, pd := range game.State.PendingDamageQueue {
		if pd.SourceID != "p1" || pd.TargetID != "p2" || !strings.EqualFold(pd.DamageType, "attack") {
			continue
		}
		totalAttackToP2 += pd.Damage
	}
	// 上限约束：本次攻击总伤害不应超过“基础伤害 + 2”。
	// 当前链路至少应包含发动前 +1，因此下界为3。
	if totalAttackToP2 < 3 || totalAttackToP2 > 4 {
		t.Fatalf("expected capped attack damage in [3,4], got %d, queue=%+v", totalAttackToP2, game.State.PendingDamageQueue)
	}
}

func TestMagicBowMagicPierce_MissDealsExactlyThreeMagicDamage(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "MagicBow", "magic_bow", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Defender", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		magicBowTestCard("atk_miss", "烈焰箭", model.CardTypeAttack, model.ElementFire),
	}
	// 防御方使用圣光使攻击未命中。
	p2.Hand = []model.Card{
		magicBowTestCard("def_light", "圣光", model.CardTypeMagic, model.ElementLight),
	}
	giveMagicBowCharges(p1, model.ElementFire, model.ElementFire)

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})
	chooseResponseSkillByID(t, game, "p1", "mb_magic_pierce")
	if err := game.handleCombatResponse(model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		CardIndex: 0,
		ExtraArgs: []string{"defend"},
	}); err != nil {
		t.Fatalf("combat defend response failed: %v", err)
	}

	totalMagicToP2 := 0
	totalAttackToP2 := 0
	for _, pd := range game.State.PendingDamageQueue {
		if pd.SourceID != "p1" || pd.TargetID != "p2" {
			continue
		}
		if strings.EqualFold(pd.DamageType, "magic") {
			totalMagicToP2 += pd.Damage
		}
		if strings.EqualFold(pd.DamageType, "attack") {
			totalAttackToP2 += pd.Damage
		}
	}
	if totalMagicToP2 != 3 {
		t.Fatalf("expected miss fallback magic damage=3, got %d", totalMagicToP2)
	}
	if totalAttackToP2 != 0 {
		t.Fatalf("expected no pending attack damage on miss branch, got %d", totalAttackToP2)
	}
	if got := magicBowChargeCount(p1, model.ElementFire); got != 1 {
		t.Fatalf("expected only first fire charge consumed on miss, remain=%d", got)
	}
	if got := p1.Tokens["mb_magic_pierce_pending"]; got != 0 {
		t.Fatalf("expected mb_magic_pierce_pending cleared after miss, got %d", got)
	}
}

func TestMagicBowThunderScatter_ExtraZeroSkipsTargetChoice(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "MagicBow", "magic_bow", model.RedCamp); err != nil {
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
	giveMagicBowCharges(p1, model.ElementThunder, model.ElementThunder)

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	if err := game.UseSkill("p1", "mb_thunder_scatter", nil, nil); err != nil {
		t.Fatalf("use thunder scatter failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "mb_thunder_scatter_extra")
	if err := game.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("choose extra x=0 failed: %v", err)
	}

	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no target-choice interrupt when extra x=0, got %+v", game.State.PendingInterrupt)
	}
	if got := magicBowChargeCount(p1, model.ElementThunder); got != 1 {
		t.Fatalf("expected only base thunder charge consumed, remain=%d", got)
	}
	if len(game.State.PendingDamageQueue) != 2 {
		t.Fatalf("expected base aoe damage to two enemies, got %d", len(game.State.PendingDamageQueue))
	}
	for _, pd := range game.State.PendingDamageQueue {
		if !strings.EqualFold(pd.DamageType, "magic") || pd.Damage != 1 {
			t.Fatalf("unexpected base thunder-scatter damage item: %+v", pd)
		}
	}
}

func TestMagicBowCharge_LockTurnDisablesPierceAndScatter(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "MagicBow", "magic_bow", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1
	giveMagicBowCharges(p1, model.ElementFire, model.ElementThunder)

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseStartup

	ctx := game.buildContext(p1, nil, model.TriggerOnTurnStart, &model.EventContext{
		Type:     model.EventTurnStart,
		SourceID: p1.ID,
	})
	if err := (&skills.MagicBowChargeHandler{}).Execute(ctx); err != nil {
		t.Fatalf("execute charge failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "mb_charge_draw_x")
	// 选择 X=0，快速完成本次启动。
	if err := game.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("finish charge with x=0 failed: %v", err)
	}

	if got := p1.TurnState.UsedSkillCounts["mb_charge_lock_turn"]; got != 1 {
		t.Fatalf("expected mb_charge_lock_turn=1 after charge, got %d", got)
	}
	game.State.Phase = model.PhaseActionSelection
	if err := game.UseSkill("p1", "mb_thunder_scatter", nil, nil); err == nil {
		t.Fatalf("expected thunder scatter locked in same turn after charge")
	}

	// 魔贯冲击在锁回合内也应不可用（即使火充能与目标条件满足）。
	attackCard := magicBowTestCard("atk_lock", "火焰斩", model.CardTypeAttack, model.ElementFire)
	pierceCtx := game.buildContext(p1, p2, model.TriggerOnAttackStart, &model.EventContext{
		Type:     model.EventAttack,
		SourceID: p1.ID,
		TargetID: p2.ID,
		Card:     &attackCard,
		AttackInfo: &model.AttackEventInfo{
			ActionType: string(model.ActionAttack),
		},
	})
	if (&skills.MagicBowMagicPierceHandler{}).CanUse(pierceCtx) {
		t.Fatalf("expected magic pierce disabled in charge-lock turn")
	}
}

func runMagicPierceHitConfirmScenario(t *testing.T, confirmSelection int) (remainFire int, attackDamage int) {
	t.Helper()
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "MagicBow", "magic_bow", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Defender", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		magicBowTestCard("atk_branch", "烈焰箭", model.CardTypeAttack, model.ElementFire),
	}
	// 两个火充能：第一个用于发动，第二个用于命中追加分支验证。
	giveMagicBowCharges(p1, model.ElementFire, model.ElementFire)

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})
	chooseResponseSkillByID(t, game, "p1", "mb_magic_pierce")
	if err := game.handleCombatResponse(model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	}); err != nil {
		t.Fatalf("combat take response failed: %v", err)
	}
	game.Drive()
	requireChoicePrompt(t, game, "p1", "mb_magic_pierce_hit_confirm")
	if err := game.handleWeakChoiceInput("p1", confirmSelection); err != nil {
		t.Fatalf("choose magic-pierce hit-confirm failed: %v", err)
	}

	remainFire = magicBowChargeCount(p1, model.ElementFire)
	for _, pd := range game.State.PendingDamageQueue {
		if pd.SourceID == "p1" && pd.TargetID == "p2" && strings.EqualFold(pd.DamageType, "attack") {
			attackDamage += pd.Damage
		}
	}
	return remainFire, attackDamage
}

func TestMagicBowMagicPierce_HitConfirmNo_DoesNotConsumeSecondFireCharge(t *testing.T) {
	remainFire, dmgNo := runMagicPierceHitConfirmScenario(t, 1)
	if remainFire != 1 {
		t.Fatalf("expected selecting NO keeps second fire charge, remain=%d", remainFire)
	}
	if dmgNo <= 0 {
		t.Fatalf("expected positive pending attack damage after NO branch, got %d", dmgNo)
	}
}

func TestMagicBowMagicPierce_HitConfirmYes_ConsumesSecondChargeAndAddsOne(t *testing.T) {
	remainNo, dmgNo := runMagicPierceHitConfirmScenario(t, 1)
	remainYes, dmgYes := runMagicPierceHitConfirmScenario(t, 0)

	if remainNo != 1 {
		t.Fatalf("precheck failed: NO branch should keep 1 fire charge, got %d", remainNo)
	}
	if remainYes != 0 {
		t.Fatalf("expected YES branch consumes second fire charge, remain=%d", remainYes)
	}
	if dmgYes != dmgNo+1 {
		t.Fatalf("expected YES branch damage = NO + 1, got yes=%d no=%d", dmgYes, dmgNo)
	}
}
