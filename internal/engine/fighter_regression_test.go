package engine

import (
	"strings"
	"testing"

	"starcup-engine/internal/model"
)

func fighterTestCard(id, name string, cardType model.CardType, element model.Element, damage int) model.Card {
	if damage <= 0 && cardType == model.CardTypeAttack {
		damage = 2
	}
	return model.Card{
		ID:          id,
		Name:        name,
		Type:        cardType,
		Element:     element,
		Damage:      damage,
		Description: name,
	}
}

func TestFighterPsiField_CapsDamageAtFour(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Fighter", "fighter", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	game.State.Deck = []model.Card{
		fighterTestCard("d1", "补牌1", model.CardTypeAttack, model.ElementFire, 2),
		fighterTestCard("d2", "补牌2", model.CardTypeAttack, model.ElementWater, 2),
		fighterTestCard("d3", "补牌3", model.CardTypeAttack, model.ElementThunder, 2),
		fighterTestCard("d4", "补牌4", model.CardTypeAttack, model.ElementWind, 2),
		fighterTestCard("d5", "补牌5", model.CardTypeAttack, model.ElementEarth, 2),
	}

	sourceCard := fighterTestCard("m1", "高伤法术", model.CardTypeMagic, model.ElementFire, 6)
	if err := game.ResolveDamage("p2", "p1", &sourceCard, "magic"); err != nil {
		t.Fatalf("resolve damage failed: %v", err)
	}
	if got := len(p1.Hand); got != 4 {
		t.Fatalf("expected psi field cap damage draw to 4 cards, got %d", got)
	}
}

func TestFighterChargeStrike_HitDamageBonus(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Fighter", "fighter", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Target", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		fighterTestCard("a1", "火焰斩", model.CardTypeAttack, model.ElementFire, 2),
	}
	game.State.Deck = []model.Card{
		fighterTestCard("d1", "补牌1", model.CardTypeAttack, model.ElementFire, 2),
		fighterTestCard("d2", "补牌2", model.CardTypeAttack, model.ElementWater, 2),
		fighterTestCard("d3", "补牌3", model.CardTypeAttack, model.ElementThunder, 2),
		fighterTestCard("d4", "补牌4", model.CardTypeAttack, model.ElementWind, 2),
	}
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	mustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0})
	chooseResponseSkillByID(t, game, "p1", "fighter_charge_strike")
	mustHandleAction(t, game, model.PlayerAction{PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"take"}})

	if got := len(p2.Hand); got != 3 {
		t.Fatalf("expected charge strike hit damage=3, got target draw=%d", got)
	}
	if got := p1.Tokens["fighter_qi"]; got != 1 {
		t.Fatalf("expected qi=1 after charge strike, got %d", got)
	}
	if got := p1.Tokens["fighter_charge_pending"]; got != 0 {
		t.Fatalf("expected fighter_charge_pending cleared on hit, got %d", got)
	}
	if got := p1.Tokens["fighter_charge_damage_pending"]; got != 0 {
		t.Fatalf("expected fighter_charge_damage_pending cleared on hit, got %d", got)
	}
}

func TestFighterChargeStrike_MissSelfDamageByQi(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Fighter", "fighter", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Target", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		fighterTestCard("a1", "雷斩", model.CardTypeAttack, model.ElementThunder, 2),
	}
	p2.Hand = []model.Card{
		fighterTestCard("m1", "圣光", model.CardTypeMagic, model.ElementLight, 0),
	}
	game.State.Deck = []model.Card{
		fighterTestCard("d1", "补牌1", model.CardTypeAttack, model.ElementFire, 2),
		fighterTestCard("d2", "补牌2", model.CardTypeAttack, model.ElementWater, 2),
	}
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	mustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0})
	chooseResponseSkillByID(t, game, "p1", "fighter_charge_strike")
	mustHandleAction(t, game, model.PlayerAction{PlayerID: "p2", Type: model.CmdRespond, CardIndex: 0, ExtraArgs: []string{"defend"}})

	if got := len(p1.Hand); got != 1 {
		t.Fatalf("expected fighter self-damage draw 1 card after miss, got hand=%d", got)
	}
	if got := p1.Tokens["fighter_qi"]; got != 1 {
		t.Fatalf("expected qi=1 after miss branch, got %d", got)
	}
	if got := p1.Tokens["fighter_charge_pending"]; got != 0 {
		t.Fatalf("expected fighter_charge_pending cleared on miss, got %d", got)
	}
	if got := p1.Tokens["fighter_charge_damage_pending"]; got != 0 {
		t.Fatalf("expected fighter_charge_damage_pending cleared on miss, got %d", got)
	}
}

func TestFighterPsiBullet_TargetChoiceAndSelfDamage(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Fighter", "fighter", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		fighterTestCard("m1", "圣光", model.CardTypeMagic, model.ElementLight, 0),
	}
	p2.Heal = 0
	game.State.Deck = []model.Card{
		fighterTestCard("d1", "补牌1", model.CardTypeAttack, model.ElementFire, 2),
		fighterTestCard("d2", "补牌2", model.CardTypeAttack, model.ElementWater, 2),
		fighterTestCard("d3", "补牌3", model.CardTypeAttack, model.ElementThunder, 2),
	}
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	mustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdMagic, TargetID: "p1", CardIndex: 0})
	chooseResponseSkillByID(t, game, "p1", "fighter_psi_bullet")
	requireChoicePrompt(t, game, "p1", "fighter_psi_bullet_target")
	mustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0}})

	if got := p1.Tokens["fighter_qi"]; got != 1 {
		t.Fatalf("expected qi=1 after psi bullet, got %d", got)
	}
	if got := len(p2.Hand); got != 1 {
		t.Fatalf("expected psi bullet target draw 1 card, got %d", got)
	}
	if got := len(p1.Hand); got != 1 {
		t.Fatalf("expected self-damage branch draw 1 card after spending magic card, got hand=%d", got)
	}
}

func TestFighterHundredDragon_BonusesAndTargetLockCancel(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Fighter", "fighter", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "EnemyA", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "EnemyB", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.Tokens["fighter_hundred_dragon_form"] = 1

	attackCard := fighterTestCard("atk", "烈风斩", model.CardTypeAttack, model.ElementWind, 2)
	if got := game.applyPassiveAttackEffects(p1, p2, 2, model.Action{
		SourceID: p1.ID,
		TargetID: p2.ID,
		Type:     model.ActionAttack,
		Card:     &attackCard,
	}); got != 4 {
		t.Fatalf("expected hundred_dragon active attack damage=4, got %d", got)
	}
	if got := game.applyPassiveAttackEffects(p1, p2, 2, model.Action{
		SourceID:         p1.ID,
		TargetID:         p2.ID,
		Type:             model.ActionAttack,
		CounterInitiator: "p2",
		Card:             &attackCard,
	}); got != 3 {
		t.Fatalf("expected hundred_dragon counter attack damage=3, got %d", got)
	}

	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{fighterTestCard("a1", "火斩", model.CardTypeAttack, model.ElementFire, 2)}
	p1.Tokens["fighter_hundred_dragon_form"] = 1
	p1.Tokens["fighter_hundred_dragon_target_order"] = 2 // 锁定 p2
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	err := game.HandleAction(model.PlayerAction{PlayerID: "p1", Type: model.CmdAttack, TargetID: "p3", CardIndex: 0})
	if err == nil || !strings.Contains(err.Error(), "同一目标") {
		t.Fatalf("expected target-lock violation error, got %v", err)
	}
	if got := p1.Tokens["fighter_hundred_dragon_form"]; got != 0 {
		t.Fatalf("expected hundred_dragon form cleared after violating lock, got %d", got)
	}
	if got := p1.Tokens["fighter_hundred_dragon_target_order"]; got != 0 {
		t.Fatalf("expected hundred_dragon target lock cleared after violating lock, got %d", got)
	}
}

func TestFighterBurstCrash_NoCounterAndSelfDamage(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Fighter", "fighter", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Defender", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["fighter_qi"] = 2
	p1.Hand = []model.Card{fighterTestCard("atk1", "火焰斩", model.CardTypeAttack, model.ElementFire, 2)}
	p2 := game.State.Players["p2"]
	p2.Hand = []model.Card{fighterTestCard("cnt1", "火焰斩", model.CardTypeAttack, model.ElementFire, 2)}
	game.State.Deck = []model.Card{
		fighterTestCard("d1", "补牌1", model.CardTypeAttack, model.ElementFire, 2),
		fighterTestCard("d2", "补牌2", model.CardTypeAttack, model.ElementWater, 2),
		fighterTestCard("d3", "补牌3", model.CardTypeAttack, model.ElementThunder, 2),
	}
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	mustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0})
	chooseResponseSkillByID(t, game, "p1", "fighter_burst_crash")

	if len(game.State.CombatStack) == 0 {
		t.Fatalf("expected combat stack after burst crash")
	}
	top := game.State.CombatStack[len(game.State.CombatStack)-1]
	if top.CanBeResponded {
		t.Fatalf("expected burst crash to force no-counter")
	}

	err := game.HandleAction(model.PlayerAction{PlayerID: "p2", Type: model.CmdRespond, CardIndex: 0, TargetID: "p3", ExtraArgs: []string{"counter"}})
	if err == nil || !strings.Contains(err.Error(), "无法被应战") {
		t.Fatalf("expected counter blocked by burst crash, got %v", err)
	}

	mustHandleAction(t, game, model.PlayerAction{PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"take"}})

	if got := p1.Tokens["fighter_qi"]; got != 1 {
		t.Fatalf("expected qi reduced to 1 after burst crash, got %d", got)
	}
	if got := p1.Tokens["fighter_qiburst_force_no_counter"]; got != 0 {
		t.Fatalf("expected no-counter token consumed, got %d", got)
	}
	if got := len(p1.Hand); got != 1 {
		t.Fatalf("expected fighter self-damage draw 1 card after burst crash, got hand=%d", got)
	}
}

func TestFighterWarGodDrive_DiscardToThreeAndHeal(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Fighter", "fighter", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1
	p1.Heal = 0
	p1.Hand = []model.Card{
		fighterTestCard("h1", "火斩", model.CardTypeAttack, model.ElementFire, 2),
		fighterTestCard("h2", "水斩", model.CardTypeAttack, model.ElementWater, 2),
		fighterTestCard("h3", "风斩", model.CardTypeAttack, model.ElementWind, 2),
		fighterTestCard("h4", "地斩", model.CardTypeAttack, model.ElementEarth, 2),
		fighterTestCard("h5", "雷斩", model.CardTypeAttack, model.ElementThunder, 2),
	}
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseStartup

	game.Drive()
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptStartupSkill {
		t.Fatalf("expected startup interrupt before confirming fighter_war_god_drive")
	}
	if err := game.ConfirmStartupSkill("p1", "fighter_war_god_drive"); err != nil {
		t.Fatalf("confirm fighter_war_god_drive failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptDiscard {
		t.Fatalf("expected discard interrupt for war_god_drive followup")
	}

	mustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0, 1}})

	if got := len(p1.Hand); got != 3 {
		t.Fatalf("expected hand size 3 after war_god_drive discard, got %d", got)
	}
	if got := p1.Heal; got != 2 {
		t.Fatalf("expected heal +2 from war_god_drive, got %d", got)
	}
	if got := p1.Crystal; got != 0 {
		t.Fatalf("expected war_god_drive consume 1 crystal-like, crystal=%d", got)
	}
}
