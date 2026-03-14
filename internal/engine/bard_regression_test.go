package engine

import (
	"testing"

	"starcup-engine/internal/model"
)

func bardTestCard(id, name string, cardType model.CardType, ele model.Element) model.Card {
	return model.Card{
		ID:          id,
		Name:        name,
		Type:        cardType,
		Element:     ele,
		Damage:      1,
		Description: name,
	}
}

func TestBardDescentConcerto_TriggersAndResolves(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Bard", "bard", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "priest", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "EnemyA", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p4", "EnemyB", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	bard := game.State.Players["p1"]
	bard.Hand = []model.Card{
		bardTestCard("f_magic", "火法术", model.CardTypeMagic, model.ElementFire),
		bardTestCard("f_attack", "火攻击", model.CardTypeAttack, model.ElementFire),
		bardTestCard("w_attack", "水攻击", model.CardTypeAttack, model.ElementWater),
	}

	// 同一回合内我方先后对两名敌方造成法术伤害后，触发沉沦协奏曲可选中断。
	if paused := game.handlePostDamageResolved(&model.PendingDamage{
		SourceID: "p2", TargetID: "p3", Damage: 1, DamageType: "magic",
	}); paused {
		t.Fatalf("first magic damage should not trigger descent yet")
	}
	if paused := game.handlePostDamageResolved(&model.PendingDamage{
		SourceID: "p2", TargetID: "p4", Damage: 1, DamageType: "magic",
	}); !paused {
		t.Fatalf("second magic damage should trigger descent interrupt")
	}
	requireChoicePrompt(t, game, "p1", "bd_descent_confirm")

	if err := game.handleWeakChoiceInput("p1", 0); err != nil { // 发动
		t.Fatalf("confirm descent failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "bd_descent_element")

	if err := game.handleWeakChoiceInput("p1", 0); err != nil { // 选火系
		t.Fatalf("choose descent element failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "bd_descent_cards")

	if err := game.handleWeakChoiceInput("p1", 0); err != nil { // 第1张火牌
		t.Fatalf("choose first discard failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "bd_descent_cards")
	if err := game.handleWeakChoiceInput("p1", 0); err != nil { // 第2张火牌
		t.Fatalf("choose second discard failed: %v", err)
	}

	// 弃牌中包含法术牌，进入追加1点法术伤害目标选择。
	requireChoicePrompt(t, game, "p1", "bd_descent_target")
	if err := game.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("choose descent bonus target failed: %v", err)
	}

	if got := bard.Tokens["bd_inspiration"]; got != 1 {
		t.Fatalf("expected inspiration=1, got %d", got)
	}
	if got := bard.Tokens["bd_descent_used_turn"]; got != 1 {
		t.Fatalf("expected descent used flag=1, got %d", got)
	}
	if got := len(bard.Hand); got != 1 {
		t.Fatalf("expected bard hand reduced to 1, got %d", got)
	}
	if got := len(game.State.PendingDamageQueue); got != 1 {
		t.Fatalf("expected one bonus pending damage, got %d", got)
	}
	pd := game.State.PendingDamageQueue[0]
	if pd.SourceID != "p1" || pd.DamageType != "magic" || pd.Damage != 1 {
		t.Fatalf("unexpected bonus damage payload: %+v", pd)
	}
}

func TestBardDissonanceChord_DrawModeAndReleasePrisoner(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Bard", "bard", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	bard := game.State.Players["p1"]
	enemy := game.State.Players["p2"]
	bard.IsActive = true
	bard.TurnState = model.NewPlayerTurnState()
	bard.Tokens["bd_inspiration"] = 3
	bard.Tokens["bd_prisoner_form"] = 1
	bard.Hand = []model.Card{
		bardTestCard("h1", "手牌1", model.CardTypeAttack, model.ElementFire),
	}
	enemy.Hand = []model.Card{
		bardTestCard("e1", "敌方牌1", model.CardTypeAttack, model.ElementWater),
	}
	game.State.Deck = []model.Card{
		bardTestCard("d1", "牌堆1", model.CardTypeAttack, model.ElementWind),
		bardTestCard("d2", "牌堆2", model.CardTypeAttack, model.ElementThunder),
	}
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	if err := game.UseSkill("p1", "bd_dissonance_chord", nil, nil); err != nil {
		t.Fatalf("use dissonance failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "bd_dissonance_x")
	if err := game.handleWeakChoiceInput("p1", 0); err != nil { // X=2
		t.Fatalf("choose X failed: %v", err)
	}
	if got := bard.Tokens["bd_inspiration"]; got != 1 {
		t.Fatalf("expected inspiration consumed to 1, got %d", got)
	}
	if got := bard.Tokens["bd_prisoner_form"]; got != 0 {
		t.Fatalf("expected prisoner form released, got %d", got)
	}

	requireChoicePrompt(t, game, "p1", "bd_dissonance_mode")
	if err := game.handleWeakChoiceInput("p1", 0); err != nil { // 摸牌分支
		t.Fatalf("choose mode failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "bd_dissonance_target")
	if err := game.handleWeakChoiceInput("p1", 1); err != nil { // 目标选 p2
		t.Fatalf("choose target failed: %v", err)
	}

	if got := len(bard.Hand); got != 2 {
		t.Fatalf("expected bard drew 1 card, hand=%d", got)
	}
	if got := len(enemy.Hand); got != 2 {
		t.Fatalf("expected target drew 1 card, hand=%d", got)
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected dissonance flow completed, got pending interrupt %+v", game.State.PendingInterrupt)
	}
}

func TestBardHopeFugue_PlaceEternalThenRousingTriggersForbiddenVerse(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Bard", "bard", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "priest", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "EnemyA", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p4", "EnemyB", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	bard := game.State.Players["p1"]
	ally := game.State.Players["p2"]
	bard.IsActive = true
	bard.TurnState = model.NewPlayerTurnState()
	bard.Crystal = 1
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseStartup

	if err := game.UseSkill("p1", "bd_hope_fugue", nil, nil); err != nil {
		t.Fatalf("use hope fugue failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "bd_hope_draw_confirm")
	if err := game.handleWeakChoiceInput("p1", 1); err != nil { // 不摸牌
		t.Fatalf("choose draw confirm failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "bd_hope_mode")
	if err := game.handleWeakChoiceInput("p1", 0); err != nil { // 放置分支
		t.Fatalf("choose hope mode failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "bd_hope_place_target")
	if err := game.handleWeakChoiceInput("p1", 0); err != nil { // 目标队友 p2
		t.Fatalf("choose place target failed: %v", err)
	}

	if holder := game.bardEternalHolderID(bard); holder != "p2" {
		t.Fatalf("expected eternal movement holder p2, got %q", holder)
	}

	// 切到队友回合开始：应触发激昂狂想曲。
	bard.IsActive = false
	ally.IsActive = true
	ally.TurnState = model.NewPlayerTurnState()
	game.State.CurrentTurn = 1
	game.State.Phase = model.PhaseStartup
	game.State.PendingInterrupt = nil
	game.Drive()

	requireChoicePrompt(t, game, "p1", "bd_rousing_mode")
	if err := game.handleWeakChoiceInput("p1", 0); err != nil { // 选伤害分支
		t.Fatalf("choose rousing mode failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "bd_rousing_targets")
	if err := game.handleWeakChoiceInput("p1", 0); err != nil { // 先选 p3
		t.Fatalf("choose rousing first target failed: %v", err)
	}
	requireChoicePrompt(t, game, "p1", "bd_rousing_targets")
	if err := game.handleWeakChoiceInput("p1", 0); err != nil { // 再选 p4（重排后 index=0）
		t.Fatalf("choose rousing second target failed: %v", err)
	}

	if got := bard.Tokens["bd_inspiration"]; got != 1 {
		t.Fatalf("expected forbidden verse add inspiration to 1, got %d", got)
	}
	if holder := game.bardEternalHolderID(bard); holder != "" {
		t.Fatalf("expected eternal movement removed by forbidden verse, holder=%q", holder)
	}
	if got := len(game.State.PendingDamageQueue); got != 2 {
		t.Fatalf("expected rousing queued 2 magic damages, got %d", got)
	}
}

func TestBardVictorySymphony_AtInspirationCapEntersPrisonerAndSelfDamages(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Bard", "bard", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "priest", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	bard := game.State.Players["p1"]
	ally := game.State.Players["p2"]
	bard.Tokens["bd_inspiration"] = 3
	bard.Tokens["bd_prisoner_form"] = 0
	if err := game.placeBardEternalMovement(bard, ally); err != nil {
		t.Fatalf("place eternal movement failed: %v", err)
	}

	ally.IsActive = true
	ally.TurnState = model.NewPlayerTurnState()
	game.State.CurrentTurn = 1
	game.State.Phase = model.PhaseTurnEnd
	game.State.PendingInterrupt = nil
	game.Drive()

	requireChoicePrompt(t, game, "p1", "bd_victory_mode")
	if err := game.handleWeakChoiceInput("p1", 1); err != nil { // 分支②
		t.Fatalf("choose victory mode failed: %v", err)
	}

	if got := bard.Tokens["bd_prisoner_form"]; got != 1 {
		t.Fatalf("expected bard enter prisoner form at inspiration cap, got %d", got)
	}
	if got := len(game.State.PendingDamageQueue); got != 1 {
		t.Fatalf("expected one self magic damage from forbidden verse, got %d", got)
	}
	pd := game.State.PendingDamageQueue[0]
	if pd.TargetID != "p1" || pd.DamageType != "magic" || pd.Damage != 3 {
		t.Fatalf("unexpected self-damage payload: %+v", pd)
	}
}
