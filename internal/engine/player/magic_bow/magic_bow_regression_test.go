package magic_bow_test

import (
	"starcup-engine/internal/engine"
	"starcup-engine/internal/testutils"
	"strings"
	"testing"

	"starcup-engine/internal/data"
	magicbowplayer "starcup-engine/internal/engine/player/magic_bow"
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
	magicbowplayer.AddChargeCards(p, cards)
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

func chooseMagicBowChoice(t *testing.T, game *engine.GameEngine, playerID, choiceType string, selections ...int) {
	t.Helper()
	testutils.RequireChoicePrompt(t, game, playerID, choiceType)
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   playerID,
		Type:       model.CmdSelect,
		Selections: selections,
	})
}

func chooseMagicPierceAndFireCharge(t *testing.T, game *engine.GameEngine, playerID string) {
	t.Helper()
	testutils.ChooseResponseSkillByID(t, game, playerID, "mb_magic_pierce")
	chooseMagicBowChoice(t, game, playerID, "mb_magic_pierce_charge", 0)
}

func pendingAttackDamageTo(game *engine.GameEngine, sourceID, targetID string) int {
	total := 0
	for _, pd := range game.State.PendingDamageQueue {
		if pd.SourceID == sourceID && pd.TargetID == targetID && strings.EqualFold(string(pd.DamageType), string(model.AttackDamage)) {
			total += pd.Damage
		}
	}
	return total
}

func TestMagicBowMagicPierce_MissDealsMagicDamageAndLocksMultiShot(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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
	game.State.TurnStage = model.TurnStageActionExecution

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdAttack,
		TargetID: "p2",
		CardID:   testutils.PlayableCardID(t, game, "p1", 0),
	})
	chooseMagicPierceAndFireCharge(t, game, "p1")

	if err := game.HandleCombatResponse(model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		CardID:    testutils.PlayableCardID(t, game, "p2", 0),
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
	if got := p1.TurnState.SkillFlowState["mb_magic_pierce_pending"]; got != 0 {
		t.Fatalf("expected mb_magic_pierce_pending cleared, got %d", got)
	}
	if got := p1.TurnState.UsedSkillCounts["mb_magic_pierce_used_turn"]; got != 1 {
		t.Fatalf("expected magic pierce used mark=1, got %d", got)
	}

	multiShotCtx := game.BuildContext(p1, nil, model.TimingActionEnd, &model.EventContext{
		Type:       model.EventPhaseEnd,
		SourceID:   "p1",
		ActionType: model.ActionAttack,
		AttackInfo: &model.AttackEventInfo{
			ActionType:       string(model.ActionAttack),
			CounterInitiator: "",
		},
	})
	if (&magicbowplayer.MagicBowMultiShotHandler{}).CanUse(multiShotCtx) {
		t.Fatalf("expected multi-shot disabled after using magic pierce in same turn")
	}
}

func TestMagicBowMultiShot_TargetCannotRepeatPrevious(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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

	ctx := game.BuildContext(p1, nil, model.TimingActionEnd, &model.EventContext{
		Type:       model.EventPhaseEnd,
		SourceID:   "p1",
		ActionType: model.ActionAttack,
		AttackInfo: &model.AttackEventInfo{
			ActionType:       string(model.ActionAttack),
			CounterInitiator: "",
		},
	})
	h := &magicbowplayer.MagicBowMultiShotHandler{}
	if !h.CanUse(ctx) {
		t.Fatalf("expected multi-shot usable with wind charge and valid alternate target")
	}
	if err := h.Execute(ctx); err != nil {
		t.Fatalf("execute multi-shot failed: %v", err)
	}
	chooseMagicBowChoice(t, game, "p1", "mb_multi_shot_charge", 0)
	testutils.RequireChoicePrompt(t, game, "p1", "mb_multi_shot_target")

	targetIDs := pendingChoiceTargetIDs(game.State.PendingInterrupt)
	if len(targetIDs) != 1 || targetIDs[0] != "p3" {
		t.Fatalf("expected only p3 as valid target, got %v", targetIDs)
	}

	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
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
	if got := magicbowplayer.ChargeCount(p1, ""); got != 0 {
		t.Fatalf("expected wind charge consumed, remaining=%d", got)
	}
}

func TestMagicBowCharge_FollowupPlaceCharges(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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

	ctx := game.BuildContext(p1, nil, model.TimingTurnStart, &model.EventContext{
		Type:     model.EventTurnStart,
		SourceID: "p1",
	})
	if err := (&magicbowplayer.MagicBowChargeHandler{}).Execute(ctx); err != nil {
		t.Fatalf("execute charge failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "mb_charge_draw_x")

	// 选择 X=2（摸2张）
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{2}}); err != nil {
		t.Fatalf("choose charge draw x failed: %v", err)
	}
	// 新流程：直接进入盖牌多选（跳过数量选择步骤）
	testutils.RequireChoicePrompt(t, game, "p1", "mb_charge_place_cards")

	// 多选：一次性选择2张手牌作为充能
	// 前端传来的选项索引（0=第一张手牌，1=第二张手牌）
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0, 1}}); err != nil {
		t.Fatalf("multi-select charge cards failed: %v", err)
	}

	if got := magicbowplayer.ChargeCount(p1, ""); got != 2 {
		t.Fatalf("expected 2 charges placed, got %d", got)
	}
	// mb_charge_count 由服务端 buildStateForPlayer 写入 PlayerView.indicators，引擎内不再同步到 Player.Tokens
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
	game := engine.NewGameEngine(testutils.NoopObserver{})
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

	ctx := game.BuildContext(p1, nil, model.TimingTurnStart, &model.EventContext{
		Type:     model.EventTurnStart,
		SourceID: "p1",
	})
	if err := (&magicbowplayer.MagicBowChargeHandler{}).Execute(ctx); err != nil {
		t.Fatalf("execute charge failed: %v", err)
	}

	// 新规则：先弃到4张，再让玩家选择X。
	if game.State.PendingInterrupt == nil || !engine.IsDiscardSelectionInterrupt(game.State.PendingInterrupt) {
		t.Fatalf("expected discard interrupt first, got %+v", game.State.PendingInterrupt)
	}
	data, _ := game.State.PendingInterrupt.Context.(map[string]interface{})
	downTo, _ := data["discard_down_to"].(int)
	if downTo != 4 {
		t.Fatalf("expected discard_down_to=4, got %v", data["discard_down_to"])
	}

	if err := game.ConfirmDiscard("p1", []int{4, 5}); err != nil {
		t.Fatalf("discard to 4 for charge failed: %v", err)
	}
	if got := len(p1.Hand); got != 4 {
		t.Fatalf("expected hand size=4 after forced discard, got %d", got)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "mb_charge_draw_x")
}

func TestMagicBowCharge_DrawOverflowMoraleLossWithoutDiscard(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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

	ctx := game.BuildContext(p1, nil, model.TimingTurnStart, &model.EventContext{
		Type:     model.EventTurnStart,
		SourceID: "p1",
	})
	if err := (&magicbowplayer.MagicBowChargeHandler{}).Execute(ctx); err != nil {
		t.Fatalf("execute charge failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "mb_charge_draw_x")

	// 选择X=4，4->8，默认上限6，爆士气2但不弃牌。
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{4}}); err != nil {
		t.Fatalf("choose charge draw x failed: %v", err)
	}
	if got := game.State.RedMorale; got != redMoraleBefore-2 {
		t.Fatalf("expected red morale -2 after overflow draw, before=%d after=%d", redMoraleBefore, got)
	}
	if got := len(p1.Hand); got != 8 {
		t.Fatalf("expected no discard after overflow draw, hand should stay 8, got %d", got)
	}
	if engine.IsDiscardSelectionInterrupt(game.State.PendingInterrupt) {
		t.Fatalf("should not open discard interrupt after charge overflow draw")
	}

	// 新流程：直接进入盖牌多选（跳过数量选择步骤）
	// maxPlace=4（X=4，上限8个充能，当前0，room=8）
	if game.State.PendingInterrupt == nil || testutils.ChoiceTypeOfInterrupt(game.State.PendingInterrupt) != "mb_charge_place_cards" {
		t.Fatalf("expected enter place-cards multi-select choice after draw, got %+v", game.State.PendingInterrupt)
	}

	// 多选：一次性选择4张手牌作为充能
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0, 1, 2, 3}}); err != nil {
		t.Fatalf("multi-select charge cards failed: %v", err)
	}
	if got := magicbowplayer.ChargeCount(p1, ""); got != 4 {
		t.Fatalf("expected 4 charges placed, got %d", got)
	}
}

func TestMagicBowThunderScatter_ExtraDamageSplit(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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
	game.State.TurnStage = model.TurnStageActionExecution

	if err := game.UseSkill("p1", "mb_thunder_scatter", nil, nil); err != nil {
		t.Fatalf("use thunder scatter failed: %v", err)
	}
	chooseMagicBowChoice(t, game, "p1", "mb_thunder_scatter_base_charge", 0)
	testutils.RequireChoicePrompt(t, game, "p1", "mb_thunder_scatter_extra")

	// 选择额外移除2个雷系充能。
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{2}}); err != nil {
		t.Fatalf("choose thunder scatter extra failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "mb_thunder_scatter_target")

	// 目标列表应为 [p2,p3]，选择第一个目标 p2。
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("choose thunder scatter target failed: %v", err)
	}

	if got := magicbowplayer.ChargeCount(p1, ""); got != 0 {
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
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "MagicBow", "magic_bow", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Defender", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		magicBowTestCard("atk_hit", "烈焰箭", model.CardTypeAttack, model.ElementFire),
	}
	game.State.Deck = []model.Card{
		magicBowTestCard("d1", "补牌1", model.CardTypeAttack, model.ElementFire),
		magicBowTestCard("d2", "补牌2", model.CardTypeAttack, model.ElementWater),
		magicBowTestCard("d3", "补牌3", model.CardTypeAttack, model.ElementWind),
		magicBowTestCard("d4", "补牌4", model.CardTypeAttack, model.ElementThunder),
		magicBowTestCard("d5", "补牌5", model.CardTypeAttack, model.ElementLight),
	}
	// 预置3个火系充能，命中追加只应消耗1个（总共最多+2伤害）。
	giveMagicBowCharges(p1, model.ElementFire, model.ElementFire, model.ElementFire)

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdAttack,
		TargetID: "p2",
		CardID:   testutils.PlayableCardID(t, game, "p1", 0),
	})
	chooseMagicPierceAndFireCharge(t, game, "p1")
	if err := game.HandleCombatResponse(model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	}); err != nil {
		t.Fatalf("combat take response failed: %v", err)
	}
	game.Drive()
	testutils.RequireChoicePrompt(t, game, "p1", "mb_magic_pierce_hit_bonus")
	queuedBeforeBonus := pendingAttackDamageTo(game, "p1", "p2")
	chooseMagicBowChoice(t, game, "p1", "mb_magic_pierce_hit_bonus", 0)
	chooseMagicBowChoice(t, game, "p1", "mb_magic_pierce_hit_charge", 0)

	if got := magicbowplayer.ChargeCount(p1, model.ElementFire); got != 1 {
		t.Fatalf("expected remain 1 fire charge after at-most-once hit bonus, got %d", got)
	}
	if got := p1.TurnState.SkillFlowState["mb_magic_pierce_pending"]; got != 0 {
		t.Fatalf("expected mb_magic_pierce_pending cleared, got %d", got)
	}
	if got := len(p2.Hand); got != queuedBeforeBonus+1 {
		t.Fatalf("expected hit bonus to raise final attack damage by 1, before=%d got hand=%d", queuedBeforeBonus, got)
	}
}

func TestMagicBowMagicPierce_MissDealsExactlyThreeMagicDamage(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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
	game.State.TurnStage = model.TurnStageActionExecution

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdAttack,
		TargetID: "p2",
		CardID:   testutils.PlayableCardID(t, game, "p1", 0),
	})
	chooseMagicPierceAndFireCharge(t, game, "p1")
	if err := game.HandleCombatResponse(model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		CardID:    testutils.PlayableCardID(t, game, "p2", 0),
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
		if strings.EqualFold(string(pd.DamageType), string(model.MagicAttack)) {
			totalMagicToP2 += pd.Damage
		}
		if strings.EqualFold(string(pd.DamageType), string(model.AttackDamage)) {
			totalAttackToP2 += pd.Damage
		}
	}
	if totalMagicToP2 != 3 {
		t.Fatalf("expected miss fallback magic damage=3, got %d", totalMagicToP2)
	}
	if totalAttackToP2 != 0 {
		t.Fatalf("expected no pending attack damage on miss branch, got %d", totalAttackToP2)
	}
	if got := magicbowplayer.ChargeCount(p1, model.ElementFire); got != 1 {
		t.Fatalf("expected only first fire charge consumed on miss, remain=%d", got)
	}
	if got := p1.TurnState.SkillFlowState["mb_magic_pierce_pending"]; got != 0 {
		t.Fatalf("expected mb_magic_pierce_pending cleared after miss, got %d", got)
	}
}

func TestMagicBowThunderScatter_ExtraZeroSkipsTargetChoice(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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
	game.State.TurnStage = model.TurnStageActionExecution

	if err := game.UseSkill("p1", "mb_thunder_scatter", nil, nil); err != nil {
		t.Fatalf("use thunder scatter failed: %v", err)
	}
	chooseMagicBowChoice(t, game, "p1", "mb_thunder_scatter_base_charge", 0)
	testutils.RequireChoicePrompt(t, game, "p1", "mb_thunder_scatter_extra")
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("choose extra x=0 failed: %v", err)
	}

	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no target-choice interrupt when extra x=0, got %+v", game.State.PendingInterrupt)
	}
	if got := magicbowplayer.ChargeCount(p1, model.ElementThunder); got != 1 {
		t.Fatalf("expected only base thunder charge consumed, remain=%d", got)
	}
	if len(game.State.PendingDamageQueue) != 2 {
		t.Fatalf("expected base aoe damage to two enemies, got %d", len(game.State.PendingDamageQueue))
	}
	for _, pd := range game.State.PendingDamageQueue {
		if !strings.EqualFold(string(pd.DamageType), string(model.MagicAttack)) || pd.Damage != 1 {
			t.Fatalf("unexpected base thunder-scatter damage item: %+v", pd)
		}
	}
}

func TestMagicBowCharge_LockTurnDisablesPierceAndScatter(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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
	game.State.TurnStage = model.TurnStageActionStart

	ctx := game.BuildContext(p1, nil, model.TimingTurnStart, &model.EventContext{
		Type:     model.EventTurnStart,
		SourceID: p1.ID,
	})
	if err := (&magicbowplayer.MagicBowChargeHandler{}).Execute(ctx); err != nil {
		t.Fatalf("execute charge failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "mb_charge_draw_x")
	// 选择 X=0，快速完成本次启动。
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("finish charge with x=0 failed: %v", err)
	}

	if got := p1.TurnState.UsedSkillCounts["mb_charge_lock_turn"]; got != 1 {
		t.Fatalf("expected mb_charge_lock_turn=1 after charge, got %d", got)
	}
	game.State.TurnStage = model.TurnStageActionExecution
	if err := game.UseSkill("p1", "mb_thunder_scatter", nil, nil); err == nil {
		t.Fatalf("expected thunder scatter locked in same turn after charge")
	}

	// 魔贯冲击在锁回合内也应不可用（即使火充能与目标条件满足）。
	attackCard := magicBowTestCard("atk_lock", "火焰斩", model.CardTypeAttack, model.ElementFire)
	pierceCtx := game.BuildContext(p1, p2, model.TimingAttackDeclare, &model.EventContext{
		Type:     model.EventAttack,
		SourceID: p1.ID,
		TargetID: p2.ID,
		Card:     &attackCard,
		AttackInfo: &model.AttackEventInfo{
			ActionType: string(model.ActionAttack),
		},
	})
	if (&magicbowplayer.MagicBowMagicPierceHandler{}).CanUse(pierceCtx) {
		t.Fatalf("expected magic pierce disabled in charge-lock turn")
	}
}

func TestMagicBowMagicPierce_HitBonusDeclineKeepsSecondCharge(t *testing.T) {
	t.Helper()
	game := engine.NewGameEngine(testutils.NoopObserver{})
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
	game.State.Deck = []model.Card{
		magicBowTestCard("d1", "补牌1", model.CardTypeAttack, model.ElementFire),
		magicBowTestCard("d2", "补牌2", model.CardTypeAttack, model.ElementWater),
		magicBowTestCard("d3", "补牌3", model.CardTypeAttack, model.ElementWind),
		magicBowTestCard("d4", "补牌4", model.CardTypeAttack, model.ElementThunder),
		magicBowTestCard("d5", "补牌5", model.CardTypeAttack, model.ElementLight),
	}
	// 两个火充能：第一个用于发动，第二个用于命中追加分支验证。
	giveMagicBowCharges(p1, model.ElementFire, model.ElementFire)

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdAttack,
		TargetID: "p2",
		CardID:   testutils.PlayableCardID(t, game, "p1", 0),
	})
	chooseMagicPierceAndFireCharge(t, game, "p1")
	if err := game.HandleCombatResponse(model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	}); err != nil {
		t.Fatalf("combat take response failed: %v", err)
	}
	game.Drive()
	testutils.RequireChoicePrompt(t, game, "p1", "mb_magic_pierce_hit_bonus")
	queuedBeforeDecline := pendingAttackDamageTo(game, "p1", "p2")
	chooseMagicBowChoice(t, game, "p1", "mb_magic_pierce_hit_bonus", 1)

	if remainFire := magicbowplayer.ChargeCount(p1, model.ElementFire); remainFire != 1 {
		t.Fatalf("expected second fire charge kept after declining hit bonus, remain=%d", remainFire)
	}
	if got := len(game.State.Players["p2"].Hand); got != queuedBeforeDecline {
		t.Fatalf("expected final hit damage unchanged after decline, before=%d got hand=%d", queuedBeforeDecline, got)
	}
}

func TestMagicBowDemonEye_TargetPoolIncludesAll(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "MagicBow", "magic_bow", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 1
	p1.Hand = []model.Card{
		magicBowTestCard("h1", "火焰斩", model.CardTypeAttack, model.ElementFire),
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionStart

	if err := game.UseSkill("p1", "mb_demon_eye", nil, nil); err != nil {
		t.Fatalf("use demon eye failed: %v", err)
	}
	// Now shows branch selection first
	testutils.RequireChoicePrompt(t, game, "p1", "mb_demon_eye_mode")

	// Choose branch 1 (target discards)
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("choose branch 1 failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "mb_demon_eye_target")

	targetIDs := pendingChoiceTargetIDs(game.State.PendingInterrupt)
	// Now includes self (p1) + all others
	if len(targetIDs) != 3 || targetIDs[0] != "p1" || targetIDs[1] != "p2" || targetIDs[2] != "p3" {
		t.Fatalf("expected demon eye target pool include all roles, got %v", targetIDs)
	}
}

func TestMagicBowDemonEye_StartupPromptConsumesGemOnce(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "MagicBow", "magic_bow", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 1
	p1.Hand = []model.Card{
		magicBowTestCard("h1", "火焰斩", model.CardTypeAttack, model.ElementFire),
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionStart
	game.Drive()

	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptStartupSkill {
		t.Fatalf("expected startup interrupt, got %+v", game.State.PendingInterrupt)
	}
	demonEyeIdx := testutils.StartupSkillIndexByID(t, game, "p1", "mb_demon_eye")
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{demonEyeIdx},
	})

	testutils.RequireChoicePrompt(t, game, "p1", "mb_demon_eye_mode")
	if got := p1.Gem; got != 0 {
		t.Fatalf("expected demon eye startup cost to consume exactly 1 gem, got %d", got)
	}
	if got := p1.Crystal; got != 0 {
		t.Fatalf("expected demon eye not to consume or grant crystal before branch resolves, got %d", got)
	}
	if got := p1.TurnState.UsedSkillCounts["mb_demon_eye"]; got != 1 {
		t.Fatalf("expected mb_demon_eye usage recorded once, got %d", got)
	}
}

func TestMagicBowDemonEye_Branch2DrawThreeThenCharge(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "MagicBow", "magic_bow", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 1
	p1.Hand = []model.Card{
		magicBowTestCard("h1", "火焰斩", model.CardTypeAttack, model.ElementFire),
	}
	game.State.Deck = []model.Card{
		magicBowTestCard("d1", "补牌1", model.CardTypeAttack, model.ElementWater),
		magicBowTestCard("d2", "补牌2", model.CardTypeAttack, model.ElementWind),
		magicBowTestCard("d3", "补牌3", model.CardTypeAttack, model.ElementThunder),
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionStart

	if err := game.UseSkill("p1", "mb_demon_eye", nil, nil); err != nil {
		t.Fatalf("use demon eye failed: %v", err)
	}
	// Now shows branch selection first
	testutils.RequireChoicePrompt(t, game, "p1", "mb_demon_eye_mode")

	// Choose branch 2 (draw 3 cards)
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}}); err != nil {
		t.Fatalf("choose branch 2 failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "mb_demon_eye_charge_card")

	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("choose demon-eye charge card failed: %v", err)
	}
	if got := magicbowplayer.ChargeCount(p1, ""); got != 1 {
		t.Fatalf("expected demon eye to place 1 charge after draw3, got %d", got)
	}
	if got := len(p1.Hand); got != 3 {
		t.Fatalf("expected hand=3 after draw3 then place1 charge, got %d", got)
	}
	if got := p1.Crystal; got != 1 {
		t.Fatalf("expected demon eye grant 1 crystal after charge, got %d", got)
	}
	if got := p1.Gem; got != 0 {
		t.Fatalf("expected demon eye consume 1 gem, got %d", got)
	}
}

func TestMagicBowDemonEye_Branch1TargetDiscardsThenUserCharges(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "MagicBow", "magic_bow", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 1
	p1.Hand = []model.Card{
		magicBowTestCard("h1", "火焰斩", model.CardTypeAttack, model.ElementFire),
	}
	p2.Hand = []model.Card{
		magicBowTestCard("e1", "水盾", model.CardTypeMagic, model.ElementWater),
		magicBowTestCard("e2", "雷刃", model.CardTypeAttack, model.ElementThunder),
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionStart

	if err := game.UseSkill("p1", "mb_demon_eye", nil, nil); err != nil {
		t.Fatalf("use demon eye failed: %v", err)
	}
	// Now shows branch selection first
	testutils.RequireChoicePrompt(t, game, "p1", "mb_demon_eye_mode")

	// Choose branch 1 (target discards)
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("choose branch 1 failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "mb_demon_eye_target")

	// Choose target p2 (index 1 since p1 is index 0 now)
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}}); err != nil {
		t.Fatalf("choose target failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || !engine.IsDiscardSelectionInterrupt(game.State.PendingInterrupt) || game.State.PendingInterrupt.PlayerID != "p2" {
		t.Fatalf("expected demon eye force target discard interrupt, got %+v", game.State.PendingInterrupt)
	}
	if err := game.ConfirmDiscard("p2", []int{1}); err != nil {
		t.Fatalf("confirm demon eye target discard failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "mb_demon_eye_charge_card")
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("choose demon-eye charge card after target discard failed: %v", err)
	}
	if got := len(p2.Hand); got != 1 || p2.Hand[0].ID != "e1" {
		t.Fatalf("expected target chosen discard applied, remaining hand=%+v", p2.Hand)
	}
	if got := magicbowplayer.ChargeCount(p1, ""); got != 1 {
		t.Fatalf("expected demon eye to place 1 charge after target discard branch, got %d", got)
	}
	if got := p1.Crystal; got != 1 {
		t.Fatalf("expected demon eye grant 1 crystal after charge, got %d", got)
	}
}

func TestMagicBowConfig_MetadataAlignsWithDocument(t *testing.T) {
	characters := data.GetCharacters()
	var magicBow *model.Character
	for _, character := range characters {
		if character.ID == "magic_bow" {
			copy := character
			magicBow = &copy
			break
		}
	}
	if magicBow == nil {
		t.Fatalf("magic_bow character not found")
	}

	var thunderScatter *model.SkillDefinition
	var demonEye *model.SkillDefinition
	for i := range magicBow.Skills {
		switch magicBow.Skills[i].ID {
		case "mb_thunder_scatter":
			thunderScatter = &magicBow.Skills[i]
		case "mb_demon_eye":
			demonEye = &magicBow.Skills[i]
		}
	}
	if thunderScatter == nil || demonEye == nil {
		t.Fatalf("expected thunder scatter and demon eye skills present")
	}
	if thunderScatter.TargetType != model.TargetEnemy || thunderScatter.MinTargets != 0 || thunderScatter.MaxTargets != 1 {
		t.Fatalf("expected thunder scatter target metadata enemy(0..1), got type=%v min=%d max=%d", thunderScatter.TargetType, thunderScatter.MinTargets, thunderScatter.MaxTargets)
	}
	if demonEye.TargetType != model.TargetNone || demonEye.MinTargets != 0 || demonEye.MaxTargets != 0 {
		t.Fatalf("expected demon eye target metadata none(0), got type=%v min=%d max=%d", demonEye.TargetType, demonEye.MinTargets, demonEye.MaxTargets)
	}
	if strings.Contains(demonEye.Description, "选择：") {
		t.Fatalf("expected demon eye description to remove old optional mode wording, got %q", demonEye.Description)
	}
}

func TestMagicBowCharge_FullCapSkipsPlaceChoice(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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
	}
	// 已有8个充能，上限满了
	giveMagicBowCharges(p1, model.ElementFire, model.ElementFire, model.ElementFire, model.ElementFire, model.ElementThunder, model.ElementThunder, model.ElementThunder, model.ElementThunder)
	game.State.Deck = []model.Card{
		magicBowTestCard("d1", "补牌1", model.CardTypeAttack, model.ElementFire),
		magicBowTestCard("d2", "补牌2", model.CardTypeMagic, model.ElementThunder),
	}

	ctx := game.BuildContext(p1, nil, model.TimingTurnStart, &model.EventContext{
		Type:     model.EventTurnStart,
		SourceID: "p1",
	})
	if err := (&magicbowplayer.MagicBowChargeHandler{}).Execute(ctx); err != nil {
		t.Fatalf("execute charge failed: %v", err)
	}
	// 即使手牌<=4，也应该进入 X 选择
	testutils.RequireChoicePrompt(t, game, "p1", "mb_charge_draw_x")

	// 选择 X=2，摸2张
	if err := game.HandleAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{2}}); err != nil {
		t.Fatalf("choose charge draw x failed: %v", err)
	}

	// 充能上限满了，不进入盖牌选择，直接结束
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no pending interrupt when charge cap is full, got %+v", game.State.PendingInterrupt)
	}
	if got := len(p1.Hand); got != 4 {
		t.Fatalf("expected hand=4 after draw 2, got %d", got)
	}
	// 充能数量不变
	if got := magicbowplayer.ChargeCount(p1, ""); got != 8 {
		t.Fatalf("expected charge count unchanged at cap, got %d", got)
	}
}

// TestMagicBowCharge_StartupSkillGemSubstitution 验证启动技充能时红宝石可替代蓝水晶
func TestMagicBowCharge_StartupSkillGemSubstitution(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "MagicBow", "magic_bow", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	// 只有红宝石，没有蓝水晶
	p1.Gem = 1
	p1.Crystal = 0
	p1.Hand = []model.Card{
		magicBowTestCard("h1", "火焰斩", model.CardTypeAttack, model.ElementFire),
		magicBowTestCard("h2", "水涟斩", model.CardTypeAttack, model.ElementWater),
		magicBowTestCard("h3", "雷光斩", model.CardTypeAttack, model.ElementThunder),
		magicBowTestCard("h4", "风神斩", model.CardTypeAttack, model.ElementWind),
	}
	game.State.Deck = []model.Card{
		magicBowTestCard("d1", "补牌1", model.CardTypeAttack, model.ElementFire),
		magicBowTestCard("d2", "补牌2", model.CardTypeMagic, model.ElementThunder),
	}

	// 构建启动技上下文（TimingActionStart）
	ctx := game.BuildContext(p1, nil, model.TimingActionStart, &model.EventContext{
		Type:     model.EventTurnStart,
		SourceID: "p1",
	})

	// CanUse 应返回 true（启动技由 runtime 检查资源，handler 不重复检查）
	h := &magicbowplayer.MagicBowChargeHandler{}
	if !h.CanUse(ctx) {
		t.Fatal("CanUse should return true for startup skill with gem substitution")
	}

	// Execute 应成功（启动技的能耗已由 runtime 在调用前扣减，handler 不再扣减）
	if err := h.Execute(ctx); err != nil {
		t.Fatalf("execute charge failed with gem substitution: %v", err)
	}

	// 验证中断被推入（应该收到摸牌选择提示）
	if game.State.PendingInterrupt == nil {
		t.Fatal("expected pending interrupt after charge execute")
	}

	// 验证是摸牌选择中断
	choiceType := ""
	if data, ok := game.State.PendingInterrupt.Context.(map[string]interface{}); ok {
		choiceType = data["choice_type"].(string)
	}
	if choiceType != "mb_charge_draw_x" {
		t.Fatalf("expected choice_type=mb_charge_draw_x, got %s", choiceType)
	}

	// 验证玩家红宝石已被扣除（由 runtime 执行，这里模拟 runtime 的扣减）
	// 注意：本测试只验证 handler 的 Execute 不会因"资源不足"报错
	// 实际扣减由 ConfirmStartupSkill 调用 ConsumeSkillEnergyCost 完成
}

func TestMagicBowCharge_StartupPromptPaysCrystalCostWithGem(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "MagicBow", "magic_bow", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 1
	p1.Crystal = 0
	p1.Hand = []model.Card{
		magicBowTestCard("h1", "火焰斩", model.CardTypeAttack, model.ElementFire),
		magicBowTestCard("h2", "水涟斩", model.CardTypeAttack, model.ElementWater),
		magicBowTestCard("h3", "雷光斩", model.CardTypeAttack, model.ElementThunder),
		magicBowTestCard("h4", "风神斩", model.CardTypeAttack, model.ElementWind),
	}
	game.State.Deck = []model.Card{
		magicBowTestCard("d1", "补牌1", model.CardTypeAttack, model.ElementFire),
		magicBowTestCard("d2", "补牌2", model.CardTypeMagic, model.ElementThunder),
	}
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionStart

	game.Drive()
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptStartupSkill {
		t.Fatalf("expected startup interrupt, got %+v", game.State.PendingInterrupt)
	}

	chargeIdx := testutils.StartupSkillIndexByID(t, game, "p1", "mb_charge")
	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{chargeIdx},
	})

	testutils.RequireChoicePrompt(t, game, "p1", "mb_charge_draw_x")
	if got := p1.Crystal; got != 0 {
		t.Fatalf("expected no blue crystal after paying charge cost, got %d", got)
	}
	if got := p1.Gem; got != 0 {
		t.Fatalf("expected red gem consumed as crystal substitute, got %d", got)
	}
	if got := p1.TurnState.UsedSkillCounts["mb_charge"]; got != 1 {
		t.Fatalf("expected mb_charge usage recorded once, got %d", got)
	}
	if got := p1.TurnState.UsedSkillCounts["mb_charge_lock_turn"]; got != 1 {
		t.Fatalf("expected mb_charge lock after startup confirm, got %d", got)
	}
}
