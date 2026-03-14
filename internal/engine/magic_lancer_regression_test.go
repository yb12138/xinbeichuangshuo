package engine

import (
	"strings"
	"testing"

	"starcup-engine/internal/engine/skills"
	"starcup-engine/internal/model"
)

func magicLancerTestCard(id, name string, cardType model.CardType, element model.Element, damage int) model.Card {
	if damage <= 0 {
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

func TestMagicLancerDarkRelease_HandCapAndAttackBonusAndLock(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Lancer", "magic_lancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Target", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseStartup
	if err := game.UseSkill("p1", "ml_dark_release", nil, nil); err != nil {
		t.Fatalf("use ml_dark_release failed: %v", err)
	}

	if got := p1.Tokens["ml_phantom_form"]; got != 1 {
		t.Fatalf("expected ml_phantom_form=1, got %d", got)
	}
	if got := game.GetMaxHand(p1); got != 5 {
		t.Fatalf("expected max hand=5 in phantom form, got %d", got)
	}

	fullnessHandler := skills.GetHandler("ml_fullness")
	if fullnessHandler == nil {
		t.Fatal("ml_fullness handler not found")
	}
	ctx := game.buildContext(p1, nil, model.TriggerNone, nil)
	if fullnessHandler.CanUse(ctx) {
		t.Fatal("ml_fullness should be locked in the same turn after dark release")
	}

	blackSpearHandler := skills.GetHandler("ml_black_spear")
	if blackSpearHandler == nil {
		t.Fatal("ml_black_spear handler not found")
	}
	p2.Hand = []model.Card{magicLancerTestCard("h1", "火焰斩", model.CardTypeAttack, model.ElementFire, 2)}
	hitCtx := game.buildContext(p1, p2, model.TriggerOnAttackHit, &model.EventContext{
		Type:     model.EventAttack,
		SourceID: p1.ID,
		TargetID: p2.ID,
		AttackInfo: &model.AttackEventInfo{
			ActionType:       "Attack",
			CounterInitiator: "",
		},
	})
	if blackSpearHandler.CanUse(hitCtx) {
		t.Fatal("ml_black_spear should be locked in the same turn after dark release")
	}

	attackCard := magicLancerTestCard("atk1", "雷斩", model.CardTypeAttack, model.ElementThunder, 2)
	dmg1 := game.applyPassiveAttackEffects(p1, p2, 2, model.Action{
		SourceID:         p1.ID,
		TargetID:         p2.ID,
		Type:             model.ActionAttack,
		CounterInitiator: "",
		Card:             &attackCard,
	})
	if dmg1 != 3 {
		t.Fatalf("expected first active attack damage=3, got %d", dmg1)
	}
	if got := p1.TurnState.UsedSkillCounts["ml_dark_release_next_attack_bonus"]; got != 0 {
		t.Fatalf("expected dark release bonus consumed, got %d", got)
	}
	dmg2 := game.applyPassiveAttackEffects(p1, p2, 2, model.Action{
		SourceID:         p1.ID,
		TargetID:         p2.ID,
		Type:             model.ActionAttack,
		CounterInitiator: "",
		Card:             &attackCard,
	})
	if dmg2 != 2 {
		t.Fatalf("expected subsequent active attack damage back to 2, got %d", dmg2)
	}
}

func TestMagicLancerPhantomStardust_LeaveFormAndPromptTarget(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Lancer", "magic_lancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["ml_phantom_form"] = 1
	p1.Hand = []model.Card{}

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseStartup
	if err := game.UseSkill("p1", "ml_phantom_stardust", nil, nil); err != nil {
		t.Fatalf("use ml_phantom_stardust failed: %v", err)
	}
	if len(game.State.PendingDamageQueue) == 0 {
		t.Fatalf("expected pending self magic damage from ml_phantom_stardust")
	}

	interrupted := game.processPendingDamages()
	if !interrupted {
		t.Fatalf("expected processPendingDamages to pause on stardust target prompt")
	}
	if got := p1.Tokens["ml_phantom_form"]; got != 0 {
		t.Fatalf("expected leave phantom form after stardust self damage, got %d", got)
	}
	if got := p1.Tokens["ml_stardust_pending"]; got != 0 {
		t.Fatalf("expected ml_stardust_pending cleared, got %d", got)
	}
	if game.State.PendingInterrupt == nil {
		t.Fatalf("expected pending interrupt for stardust target")
	}
	if got := choiceTypeOf(game.State.PendingInterrupt); got != "ml_stardust_target" {
		t.Fatalf("expected choice_type ml_stardust_target, got %q", got)
	}
}

func TestMagicLancerDarkBind_BlocksMagicUseAndDefend(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Lancer", "magic_lancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{magicLancerTestCard("m1", "圣光", model.CardTypeMagic, model.ElementLight, 0)}

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseBeforeAction
	if err := game.PerformMagic("p1", "p2", 0); err == nil || !strings.Contains(err.Error(), "法术牌") {
		t.Fatalf("expected dark bind to block PerformMagic, got err=%v", err)
	}

	game.State.CombatStack = []model.CombatRequest{{
		AttackerID:     "p2",
		TargetID:       "p1",
		Card:           &model.Card{ID: "a1", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
		CanBeResponded: true,
	}}
	game.State.Phase = model.PhaseCombatInteraction
	err := game.handleCombatResponse(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdRespond,
		CardIndex: 0,
		ExtraArgs: []string{"defend"},
	})
	if err == nil || !strings.Contains(err.Error(), "黑暗束缚") {
		t.Fatalf("expected dark bind to block defend, got err=%v", err)
	}
}

func TestMagicLancerFullness_FlowBonusAndExtraAttack(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Lancer", "magic_lancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p3 := game.State.Players["p3"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		magicLancerTestCard("cost", "圣光", model.CardTypeMagic, model.ElementLight, 0),
		magicLancerTestCard("atk", "雷斩", model.CardTypeAttack, model.ElementThunder, 2),
	}
	p2.Hand = []model.Card{magicLancerTestCard("ally", "雷击", model.CardTypeAttack, model.ElementThunder, 2)}
	p3.Hand = []model.Card{magicLancerTestCard("enemy", "圣光", model.CardTypeMagic, model.ElementLight, 0)}

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection
	if err := game.UseSkill("p1", "ml_fullness", nil, nil); err != nil {
		t.Fatalf("use ml_fullness failed: %v", err)
	}
	if got := choiceTypeOf(game.State.PendingInterrupt); got != "ml_fullness_cost_card" {
		t.Fatalf("expected ml_fullness_cost_card prompt, got %q", got)
	}

	if err := game.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("choose fullness cost card failed: %v", err)
	}
	if got := choiceTypeOf(game.State.PendingInterrupt); got != "ml_fullness_discard_step" {
		t.Fatalf("expected ml_fullness_discard_step prompt, got %q", got)
	}

	// 自己：可跳过
	if err := game.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("self skip failed: %v", err)
	}
	// 队友：可跳过
	if err := game.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("ally skip failed: %v", err)
	}
	// 敌方：必须弃牌，仅有1项可选
	if err := game.handleWeakChoiceInput("p1", 0); err != nil {
		t.Fatalf("enemy discard failed: %v", err)
	}

	if got := p1.TurnState.UsedSkillCounts["ml_fullness_next_attack_bonus"]; got != 1 {
		t.Fatalf("expected ml_fullness_next_attack_bonus=1, got %d", got)
	}
	if len(p1.TurnState.PendingActions) == 0 {
		t.Fatalf("expected extra attack action token from ml_fullness")
	}
	last := p1.TurnState.PendingActions[len(p1.TurnState.PendingActions)-1]
	if last.MustType != "Attack" {
		t.Fatalf("expected extra action type Attack, got %+v", last)
	}
	if len(p3.Hand) != 0 {
		t.Fatalf("expected enemy hand to be discarded, got %d cards", len(p3.Hand))
	}

	dmg := game.applyPassiveAttackEffects(p1, p2, 2, model.Action{
		SourceID:         p1.ID,
		TargetID:         p2.ID,
		Type:             model.ActionAttack,
		CounterInitiator: "",
		Card:             &model.Card{ID: "atk", Name: "雷斩", Type: model.CardTypeAttack, Element: model.ElementThunder, Damage: 2},
	})
	if dmg != 3 {
		t.Fatalf("expected fullness bonus damage to apply once (2+1), got %d", dmg)
	}
	if got := p1.TurnState.UsedSkillCounts["ml_fullness_next_attack_bonus"]; got != 0 {
		t.Fatalf("expected fullness bonus consumed, got %d", got)
	}
}

func TestMagicLancerBlackSpear_ConsumesCrystalAndAddsDamage(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Lancer", "magic_lancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["ml_phantom_form"] = 1
	p1.Crystal = 2
	p2.Hand = []model.Card{magicLancerTestCard("h1", "火斩", model.CardTypeAttack, model.ElementFire, 2)}

	handler := skills.GetHandler("ml_black_spear")
	if handler == nil {
		t.Fatal("ml_black_spear handler not found")
	}
	ctx := game.buildContext(p1, p2, model.TriggerOnAttackHit, &model.EventContext{
		Type:     model.EventAttack,
		SourceID: p1.ID,
		TargetID: p2.ID,
		AttackInfo: &model.AttackEventInfo{
			ActionType:       "Attack",
			CounterInitiator: "",
		},
	})
	if !handler.CanUse(ctx) {
		t.Fatal("expected ml_black_spear can use on active hit vs hand 1/2 target")
	}
	game.State.PendingDamageQueue = []model.PendingDamage{{
		SourceID:   p1.ID,
		TargetID:   p2.ID,
		Damage:     2,
		DamageType: "Attack",
		Stage:      1,
	}}
	if err := handler.Execute(ctx); err != nil {
		t.Fatalf("execute ml_black_spear failed: %v", err)
	}
	if got := choiceTypeOf(game.State.PendingInterrupt); got != "ml_black_spear_x" {
		t.Fatalf("expected ml_black_spear_x prompt, got %q", got)
	}

	// selection=1 -> X=2
	if err := game.handleWeakChoiceInput("p1", 1); err != nil {
		t.Fatalf("choose black spear x failed: %v", err)
	}
	if p1.Crystal != 0 || p1.Gem != 0 {
		t.Fatalf("expected consume 2 crystal-like resources, got gem=%d crystal=%d", p1.Gem, p1.Crystal)
	}
	if len(game.State.PendingDamageQueue) == 0 {
		t.Fatalf("expected pending attack damage entry")
	}
	if got := game.State.PendingDamageQueue[0].Damage; got != 6 {
		t.Fatalf("expected attack damage increased to 6 (2 + (2+2)), got %d", got)
	}
}
