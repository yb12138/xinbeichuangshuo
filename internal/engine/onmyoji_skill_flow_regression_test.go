package engine

import (
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func choiceTypeOf(intr *model.Interrupt) string {
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

func TestOnmyojiDarkRitual_ChoosesTargetAtTurnEnd(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Onmyoji", "onmyoji", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["onmyoji_ghost_fire"] = 3

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseTurnEnd

	game.Drive()

	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptChoice {
		t.Fatalf("expected choice interrupt for dark ritual, got %+v", game.State.PendingInterrupt)
	}
	if got := choiceTypeOf(game.State.PendingInterrupt); got != "onmyoji_dark_ritual_target" {
		t.Fatalf("expected onmyoji_dark_ritual_target prompt, got %s", got)
	}
	// 选 p2（玩家顺序 p1,p2）
	if err := game.handleWeakChoiceInput("p1", 1); err != nil {
		t.Fatalf("choose dark ritual target failed: %v", err)
	}
	if got := p1.Tokens["onmyoji_ghost_fire"]; got != 0 {
		t.Fatalf("expected ghost fire reset to 0, got %d", got)
	}
	if len(game.State.PendingDamageQueue) == 0 {
		t.Fatalf("expected pending damage from dark ritual")
	}
	pd := game.State.PendingDamageQueue[0]
	if pd.SourceID != "p1" || pd.TargetID != "p2" || pd.Damage != 2 || pd.DamageType != "magic" {
		t.Fatalf("unexpected dark ritual pending damage: %+v", pd)
	}
}

func TestOnmyojiBinding_RequiresGemAndCrystal(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Attacker", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "TargetAlly", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Onmyoji", "onmyoji", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p4", "AttackerAlly", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	p3 := game.State.Players["p3"]
	p3.Tokens["onmyoji_form"] = 1
	p3.Hand = []model.Card{
		{ID: "c1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Faction: "咏", Damage: 2},
	}

	req := model.CombatRequest{
		AttackerID:     "p1",
		TargetID:       "p2",
		Card:           &model.Card{ID: "atk", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Faction: "咏", Damage: 2},
		CanBeResponded: true,
	}

	// 仅2宝石，0水晶：不满足式神咒束代应战成本
	game.State.BlueGems = 2
	game.State.BlueCrystals = 0
	if ok := game.tryStartOnmyojiBindingInterrupt(&req); ok {
		t.Fatalf("binding should not start without crystal")
	}

	// 1宝石+1水晶：可触发询问
	req2 := model.CombatRequest{
		AttackerID:     "p1",
		TargetID:       "p2",
		Card:           &model.Card{ID: "atk2", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Faction: "咏", Damage: 2},
		CanBeResponded: true,
	}
	game.State.BlueGems = 1
	game.State.BlueCrystals = 1
	if ok := game.tryStartOnmyojiBindingInterrupt(&req2); !ok {
		t.Fatalf("binding should start with 1 gem + 1 crystal")
	}
	if game.State.PendingInterrupt == nil || choiceTypeOf(game.State.PendingInterrupt) != "onmyoji_binding_confirm" {
		t.Fatalf("expected binding confirm interrupt, got %+v", game.State.PendingInterrupt)
	}
}

func TestOnmyojiLifeBarrier_Mode1_X3NoMoraleLoss(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Onmyoji", "onmyoji", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.Deck = rules.InitDeck()
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1
	p1.Tokens["onmyoji_ghost_fire"] = 2 // 技能后变3
	// 扣卡前手牌正好上限，确保受到3点伤害后爆牌并触发弃牌流程。
	p1.Hand = []model.Card{
		{ID: "h1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Faction: "咏", Damage: 2},
		{ID: "h2", Name: "水涟斩", Type: model.CardTypeAttack, Element: model.ElementWater, Faction: "咏", Damage: 2},
		{ID: "h3", Name: "风神斩", Type: model.CardTypeAttack, Element: model.ElementWind, Faction: "咏", Damage: 2},
		{ID: "h4", Name: "雷光斩", Type: model.CardTypeAttack, Element: model.ElementThunder, Faction: "咏", Damage: 2},
		{ID: "h5", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight, Faction: "圣", Damage: 0},
		{ID: "h6", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementDark, Faction: "幻", Damage: 2},
	}

	redMoraleBefore := game.State.RedMorale
	if err := game.UseSkill("p1", "onmyoji_life_barrier", nil, nil); err != nil {
		t.Fatalf("use life barrier failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || choiceTypeOf(game.State.PendingInterrupt) != "onmyoji_life_barrier_mode" {
		t.Fatalf("expected life barrier mode prompt, got %+v", game.State.PendingInterrupt)
	}
	// 选分支①
	if err := game.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("choose mode1 failed: %v", err)
	}
	// 选队友 p2（唯一候选）
	if err := game.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("choose support target failed: %v", err)
	}

	// 推进到爆牌弃牌中断
	game.Drive()
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptDiscard {
		t.Fatalf("expected discard interrupt from overflow, got %+v", game.State.PendingInterrupt)
	}
	if err := game.ConfirmDiscard("p1", []int{0, 1, 2}); err != nil {
		t.Fatalf("confirm discard failed: %v", err)
	}
	if game.State.RedMorale != redMoraleBefore {
		t.Fatalf("expected no morale loss when X=3 life barrier self-damage overflow, before=%d after=%d", redMoraleBefore, game.State.RedMorale)
	}
}

func TestOnmyojiLifeBarrier_Mode2_ReleaseFormAndForceAllyDiscard(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Onmyoji", "onmyoji", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1
	p1.Tokens["onmyoji_form"] = 1
	p1.Hand = []model.Card{
		{ID: "o1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Faction: "咏", Damage: 2},
		{ID: "o2", Name: "风神斩", Type: model.CardTypeAttack, Element: model.ElementWind, Faction: "咏", Damage: 2},
		{ID: "o3", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight, Faction: "圣", Damage: 0},
	}
	p2.Hand = []model.Card{
		{ID: "a1", Name: "水涟斩", Type: model.CardTypeAttack, Element: model.ElementWater, Faction: "咏", Damage: 2},
	}

	if err := game.UseSkill("p1", "onmyoji_life_barrier", nil, nil); err != nil {
		t.Fatalf("use life barrier failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || choiceTypeOf(game.State.PendingInterrupt) != "onmyoji_life_barrier_mode" {
		t.Fatalf("expected life barrier mode prompt, got %+v", game.State.PendingInterrupt)
	}
	// 选分支②
	if err := game.handleWeakChoiceInput("p1", 1); err != nil {
		t.Fatalf("choose mode2 failed: %v", err)
	}
	if got := choiceTypeOf(game.State.PendingInterrupt); got != "onmyoji_life_barrier_release_combo" {
		t.Fatalf("expected release combo prompt, got %s", got)
	}
	// 仅有一组同命格组合，索引0
	if err := game.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("choose release combo failed: %v", err)
	}
	if got := p1.Tokens["onmyoji_form"]; got != 0 {
		t.Fatalf("expected leave shikigami form, got onmyoji_form=%d", got)
	}
	if got := len(p1.Hand); got != 1 {
		t.Fatalf("expected p1 hand reduced by 2, got %d", got)
	}
	if got := choiceTypeOf(game.State.PendingInterrupt); got != "onmyoji_life_barrier_release_target" {
		t.Fatalf("expected release target prompt, got %s", got)
	}
	// 选择唯一队友 p2
	if err := game.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("choose release target failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptDiscard || game.State.PendingInterrupt.PlayerID != "p2" {
		t.Fatalf("expected ally discard interrupt for p2, got %+v", game.State.PendingInterrupt)
	}
	if err := game.ConfirmDiscard("p2", []int{0}); err != nil {
		t.Fatalf("ally confirm discard failed: %v", err)
	}
	if got := len(p2.Hand); got != 0 {
		t.Fatalf("expected p2 hand discarded to 0, got %d", got)
	}
}

func TestOnmyojiYinYangConfirm_PrioritizedBeforeNormalCombatPrompt(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Attacker", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Onmyoji", "onmyoji", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "AttackerAlly", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{ID: "atk-fire", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Faction: "咏", Damage: 2},
	}
	// 与来袭牌同命格但不同系：只能通过【阴阳转换】应战。
	p2.Hand = []model.Card{
		{ID: "ctr-water", Name: "水涟斩", Type: model.CardTypeAttack, Element: model.ElementWater, Faction: "咏", Damage: 2},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	}); err != nil {
		t.Fatalf("attack failed: %v", err)
	}

	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptChoice {
		t.Fatalf("expected yinyang confirm interrupt, got %+v", game.State.PendingInterrupt)
	}
	if got := choiceTypeOf(game.State.PendingInterrupt); got != "onmyoji_yinyang_confirm" {
		t.Fatalf("expected onmyoji_yinyang_confirm, got %s", got)
	}

	// 选择“否”：应回到常规战斗响应流程（承受/防御/应战）。
	if err := game.handleWeakChoiceInput("p2", 1); err != nil {
		t.Fatalf("decline yinyang confirm failed: %v", err)
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no pending interrupt after decline, got %+v", game.State.PendingInterrupt)
	}
	if game.State.Phase != model.PhaseCombatInteraction {
		t.Fatalf("expected phase stay in combat interaction, got %s", game.State.Phase)
	}
	if len(game.State.CombatStack) == 0 {
		t.Fatalf("expected combat stack still exists for normal response")
	}
}

func TestOnmyojiYinYangConfirm_YesBranchResolvesFactionCounterChain(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Attacker", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Onmyoji", "onmyoji", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "AttackerAlly", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	game.State.Deck = rules.InitDeck()
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p2.Tokens["onmyoji_form"] = 1
	p2.Tokens["onmyoji_ghost_fire"] = 1

	p1.Hand = []model.Card{
		{ID: "atk-fire", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Faction: "咏", Damage: 2},
	}
	p2.Hand = []model.Card{
		{ID: "ctr-water", Name: "水涟斩", Type: model.CardTypeAttack, Element: model.ElementWater, Faction: "咏", Damage: 2},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	}); err != nil {
		t.Fatalf("attack failed: %v", err)
	}
	if got := choiceTypeOf(game.State.PendingInterrupt); got != "onmyoji_yinyang_confirm" {
		t.Fatalf("expected onmyoji_yinyang_confirm, got %s", got)
	}
	if err := game.handleWeakChoiceInput("p2", 0); err != nil {
		t.Fatalf("confirm yinyang failed: %v", err)
	}
	if got := choiceTypeOf(game.State.PendingInterrupt); got != "onmyoji_yinyang_card" {
		t.Fatalf("expected onmyoji_yinyang_card, got %s", got)
	}
	// 仅1张可选应战牌
	if err := game.handleWeakChoiceInput("p2", 0); err != nil {
		t.Fatalf("choose yinyang counter card failed: %v", err)
	}
	if got := choiceTypeOf(game.State.PendingInterrupt); got != "onmyoji_yinyang_counter_target" {
		t.Fatalf("expected onmyoji_yinyang_counter_target, got %s", got)
	}
	// 仅1名可选反弹目标（p3）
	if err := game.handleWeakChoiceInput("p2", 0); err != nil {
		t.Fatalf("choose yinyang counter target failed: %v", err)
	}

	if got := p2.Tokens["onmyoji_ghost_fire"]; got != 3 {
		t.Fatalf("expected ghost fire=3 after yinyang+form chain, got %d", got)
	}
	if got := p2.Tokens["onmyoji_form"]; got != 0 {
		t.Fatalf("expected leave shikigami form, got onmyoji_form=%d", got)
	}
	if got := len(p2.Hand); got != 1 {
		t.Fatalf("expected hand size 1 (counter consume 1 then draw 1), got %d", got)
	}

	if len(game.State.CombatStack) != 1 {
		t.Fatalf("expected reflected combat stack size 1, got %d", len(game.State.CombatStack))
	}
	top := game.State.CombatStack[0]
	if top.AttackerID != "p2" || top.TargetID != "p3" {
		t.Fatalf("expected reflected combat p2->p3, got %+v", top)
	}
	if top.Card == nil {
		t.Fatalf("expected reflected combat card not nil")
	}
	if top.Card.Damage != 3 {
		t.Fatalf("expected reflected damage=3, got %d", top.Card.Damage)
	}
	if top.Card.Element != model.ElementWater {
		t.Fatalf("expected reflected element Water, got %s", top.Card.Element)
	}
}
