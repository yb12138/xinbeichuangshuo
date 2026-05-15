package fighter_test

import (
	"starcup-engine/internal/engine"
	"starcup-engine/internal/testutils"
	"strings"
	"testing"

	"starcup-engine/internal/model"
)

func containsString(arr []string, target string) bool {
	for _, s := range arr {
		if s == target {
			return true
		}
	}
	return false
}

type actionPromptObserver struct {
	lastPrompt *model.Prompt
}

func (o *actionPromptObserver) OnGameEvent(event model.GameEvent) {
	if event.Type != model.EventAskInput {
		return
	}
	prompt, ok := event.Data.(*model.Prompt)
	if !ok || prompt == nil {
		return
	}
	copied := *prompt
	copied.Options = append([]model.PromptOption(nil), prompt.Options...)
	o.lastPrompt = &copied
}

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
	game := engine.NewGameEngine(testutils.NoopObserver{})
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
	game.AddPendingDamage(model.PendingDamage{
		SourceID:   "p2",
		TargetID:   "p1",
		Damage:     sourceCard.Damage,
		DamageType: model.MagicAttack,
		Card:       &sourceCard,
	})
	game.ProcessPendingDamages()
	game.Drive()
	if got := len(p1.Hand); got != 4 {
		t.Fatalf("expected psi field cap damage draw to 4 cards, got %d", got)
	}
}

func TestFighterChargeStrike_HitDamageBonus(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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
	game.State.TurnStage = model.TurnStageActionExecution

	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0})
	testutils.ChooseResponseSkillByID(t, game, "p1", "fighter_charge_strike")
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"take"}})

	if got := len(p2.Hand); got != 3 {
		t.Fatalf("expected charge strike hit damage=3, got target draw=%d", got)
	}
	if got := p1.Tokens["fighter_qi"]; got != 1 {
		t.Fatalf("expected qi=1 after charge strike, got %d", got)
	}
	if got := p1.TurnState.SkillFlowState["fighter_charge_pending"]; got != 0 {
		t.Fatalf("expected fighter_charge_pending cleared on hit, got %d", got)
	}
}

func TestFighterChargeStrike_MissSelfDamageByQi(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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
	game.State.TurnStage = model.TurnStageActionExecution

	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0})
	testutils.ChooseResponseSkillByID(t, game, "p1", "fighter_charge_strike")
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p2", Type: model.CmdRespond, CardIndex: 0, ExtraArgs: []string{"defend"}})

	if got := len(p1.Hand); got != 1 {
		t.Fatalf("expected fighter self-damage draw 1 card after miss, got hand=%d", got)
	}
	if got := p1.Tokens["fighter_qi"]; got != 1 {
		t.Fatalf("expected qi=1 after miss branch, got %d", got)
	}
	if got := p1.TurnState.SkillFlowState["fighter_charge_pending"]; got != 0 {
		t.Fatalf("expected fighter_charge_pending cleared on miss, got %d", got)
	}
}

func TestFighterChargeStrike_ShieldBlockAfterPendingDamageCountsAsMiss(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Fighter", "fighter", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Target", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.Hand = nil
	p1.Tokens["fighter_qi"] = 3
	p1.TurnState.SkillFlowState["fighter_charge_pending"] = 1
	game.ApplyNextAttackDamageRule(p1.ID, "fighter_charge_attack_bonus", "fighter_charge_strike", 1, model.RuleLifeThisEffectChain)
	p2.Field = []*model.FieldCard{
		{
			Card: model.Card{
				ID:      "shield_field_1",
				Name:    "圣盾",
				Type:    model.CardTypeMagic,
				Element: model.ElementLight,
			},
			OwnerID:  "p2",
			SourceID: "p2",
			Mode:     model.FieldEffect,
			Effect:   model.EffectShield,
			Hook:     model.FieldHookOnDamaged,
			Duration: 1,
		},
	}
	game.State.Deck = []model.Card{
		fighterTestCard("d1", "补牌1", model.CardTypeAttack, model.ElementFire, 2),
		fighterTestCard("d2", "补牌2", model.CardTypeAttack, model.ElementWater, 2),
		fighterTestCard("d3", "补牌3", model.CardTypeAttack, model.ElementThunder, 2),
	}

	attackCard := fighterTestCard("a1", "烈风斩", model.CardTypeAttack, model.ElementWind, 2)
	game.State.PendingDamageQueue = []model.PendingDamage{
		{
			SourceID:   "p1",
			TargetID:   "p2",
			Damage:     attackCard.Damage,
			DamageType: model.AttackDamage,
			Card:       &attackCard,
			IsCounter:  false,
		},
	}

	if interrupted := game.ProcessPendingDamages(); interrupted {
		t.Fatalf("expected pending damage pipeline to finish without interrupt")
	}

	if got := len(p2.Field); got != 0 {
		t.Fatalf("expected shield consumed after full block, got field=%d", got)
	}
	if got := len(p1.Hand); got != 3 {
		t.Fatalf("expected fighter backlash draw=3 after shield miss, got hand=%d", got)
	}
	if got := game.State.RedGems; got != 0 {
		t.Fatalf("expected hit gem rollback after shield full block, got red_gems=%d", got)
	}
	if got := p1.TurnState.SkillFlowState["fighter_charge_pending"]; got != 0 {
		t.Fatalf("expected fighter_charge_pending cleared after miss settle, got %d", got)
	}
}

func TestFighterChargeStrike_GrantsQiImmediatelyBeforeCombatResult(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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
		fighterTestCard("a1", "烈风斩", model.CardTypeAttack, model.ElementWind, 2),
	}
	p2.Hand = []model.Card{
		fighterTestCard("m1", "圣光", model.CardTypeMagic, model.ElementLight, 0),
	}
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})

	testutils.ChooseResponseSkillByID(t, game, "p1", "fighter_charge_strike")
	if got := p1.Tokens["fighter_qi"]; got != 1 {
		t.Fatalf("expected qi +1 immediately after confirming charge strike, got %d", got)
	}
	if intr := game.State.PendingInterrupt; intr != nil &&
		intr.Type == model.InterruptResponseSkill &&
		intr.PlayerID == "p1" {
		for _, sid := range intr.SkillIDs {
			if sid == "fighter_burst_crash" {
				t.Fatalf("expected burst crash blocked after choosing charge strike, got pending skills %+v", intr.SkillIDs)
			}
		}
	}

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		CardIndex: 0,
		ExtraArgs: []string{"defend"},
	})

	if got := p1.Tokens["fighter_qi"]; got != 1 {
		t.Fatalf("expected qi remains +1 regardless of hit/miss result, got %d", got)
	}
}

func TestFighterPsiBullet_TargetChoiceAndSelfDamage(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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
	game.State.TurnStage = model.TurnStageActionExecution

	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdMagic, TargetID: "p1", CardIndex: 0})
	testutils.ChooseResponseSkillByID(t, game, "p1", "fighter_psi_bullet")
	testutils.RequireChoicePrompt(t, game, "p1", "fighter_psi_bullet_target")
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0}})

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

func TestFighterHundredDragon_StartupLocksTargetImmediately(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["fighter_qi"] = 3
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionStart

	game.Drive()
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptStartupSkill {
		t.Fatalf("expected startup interrupt for hundred dragon, got %+v", game.State.PendingInterrupt)
	}
	if err := game.ConfirmStartupSkill("p1", "fighter_hundred_dragon"); err != nil {
		t.Fatalf("confirm fighter_hundred_dragon failed: %v", err)
	}
	testutils.RequireChoicePrompt(t, game, "p1", "fighter_hundred_dragon_target")
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0}})

	if got := p1.Tokens["fighter_qi"]; got != 0 {
		t.Fatalf("expected hundred dragon consume 3 qi, got %d", got)
	}
	if got := p1.Form; got != model.FormFighterHundredDragon {
		t.Fatalf("expected hundred dragon form active after startup, got %q", got)
	}
	if got := p1.TurnState.SkillFlowState["fighter_hundred_dragon_target_order"]; got != 2 {
		t.Fatalf("expected hundred dragon lock target p2(order=2), got %d", got)
	}
	if got := game.State.TurnStage; got != model.TurnStageActionExecution {
		t.Fatalf("expected action execution window after choosing hundred dragon target, got %s", got)
	}
}

func TestFighterHundredDragon_BonusesAndTargetLockReleaseStillContinuesAttack(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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
	p1.Form = model.FormFighterHundredDragon

	attackCard := fighterTestCard("atk", "烈风斩", model.CardTypeAttack, model.ElementWind, 2)
	if got := game.ApplyPassiveAttackEffects(p1, p2, 2, model.Action{
		SourceID: p1.ID,
		TargetID: p2.ID,
		Type:     model.ActionAttack,
		Card:     &attackCard,
	}); got != 4 {
		t.Fatalf("expected hundred_dragon active attack damage=4, got %d", got)
	}
	if got := game.ApplyPassiveAttackEffects(p1, p2, 2, model.Action{
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
	p1.Form = model.FormFighterHundredDragon
	p1.TurnState.SkillFlowState["fighter_hundred_dragon_target_order"] = 2 // 锁定 p2
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	err := game.HandleAction(model.PlayerAction{PlayerID: "p1", Type: model.CmdAttack, TargetID: "p3", CardIndex: 0})
	if err != nil {
		t.Fatalf("expected attack continue after releasing hundred dragon, got %v", err)
	}
	if got := p1.Form; got != "" {
		t.Fatalf("expected hundred dragon form cleared after violating lock, got %q", got)
	}
	if got := p1.TurnState.SkillFlowState["fighter_hundred_dragon_target_order"]; got != 0 {
		t.Fatalf("expected hundred dragon target lock cleared after violating lock, got %d", got)
	}
	if game.State.TurnStage == model.TurnStageActionExecution &&
		game.State.CombatStage == model.CombatStageNone &&
		game.State.Subflow == model.SubflowNone {
		t.Fatalf("expected attack submission continue instead of returning to idle action execution window")
	}
}

func TestFighterHundredDragon_CannotActEndsFormAtActionPhaseEnd(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Fighter", "fighter", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = nil
	p1.Form = model.FormFighterHundredDragon
	p1.TurnState.SkillFlowState["fighter_hundred_dragon_target_order"] = 2
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdCannotAct})

	if got := p1.Form; got != "" {
		t.Fatalf("expected hundred dragon form cleared when action phase ends, got %q", got)
	}
	if got := p1.TurnState.SkillFlowState["fighter_hundred_dragon_target_order"]; got != 0 {
		t.Fatalf("expected hundred dragon target lock cleared when action phase ends, got %d", got)
	}
}

func TestFighterHundredDragon_MagicAttemptCancelsFormAndAction(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Fighter", "fighter", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		fighterTestCard("m1", "圣光", model.CardTypeMagic, model.ElementLight, 0),
	}
	p1.Form = model.FormFighterHundredDragon
	p1.TurnState.SkillFlowState["fighter_hundred_dragon_target_order"] = 2
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdMagic,
		TargetID:  "p2",
		CardIndex: 0,
	})
	if err == nil || !strings.Contains(err.Error(), "不能执行法术行动") {
		t.Fatalf("expected hundred dragon magic attempt be canceled, got %v", err)
	}

	if got := p1.Form; got != "" {
		t.Fatalf("expected hundred dragon form cleared after canceled magic attempt, got %q", got)
	}
	if got := p1.TurnState.SkillFlowState["fighter_hundred_dragon_target_order"]; got != 0 {
		t.Fatalf("expected hundred dragon target lock cleared after canceled magic attempt, got %d", got)
	}
	if got := len(p1.Hand); got != 1 {
		t.Fatalf("expected canceled magic attempt not consume hand, got %d", got)
	}
	if got := game.State.TurnStage; got != model.TurnStageActionExecution {
		t.Fatalf("expected remain in action execution window after canceled magic attempt, got %s", got)
	}
}

func TestFighterHundredDragon_SpecialAttemptCancelsFormAndAction(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Fighter", "fighter", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = nil
	p1.Form = model.FormFighterHundredDragon
	p1.TurnState.SkillFlowState["fighter_hundred_dragon_target_order"] = 2
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	err := game.HandleAction(model.PlayerAction{PlayerID: "p1", Type: model.CmdBuy})
	if err == nil || !strings.Contains(err.Error(), "不能执行特殊行动") {
		t.Fatalf("expected hundred dragon special attempt be canceled, got %v", err)
	}

	if got := p1.Form; got != "" {
		t.Fatalf("expected hundred dragon form cleared after canceled special attempt, got %q", got)
	}
	if got := p1.TurnState.SkillFlowState["fighter_hundred_dragon_target_order"]; got != 0 {
		t.Fatalf("expected hundred dragon target lock cleared after canceled special attempt, got %d", got)
	}
}

func TestFighterHundredDragon_EndsWhenActionPhaseFinishes(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Fighter", "fighter", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Form = model.FormFighterHundredDragon
	p1.TurnState.SkillFlowState["fighter_hundred_dragon_target_order"] = 2
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageTurnEnd

	game.Drive()

	if got := p1.Form; got != "" {
		t.Fatalf("expected hundred dragon form end when action phase finishes, got %q", got)
	}
	if got := p1.TurnState.SkillFlowState["fighter_hundred_dragon_target_order"]; got != 0 {
		t.Fatalf("expected hundred dragon target lock cleared when action phase finishes, got %d", got)
	}
}

func TestFighterHundredDragon_ActionPromptOnlyKeepsAttackEntry(t *testing.T) {
	obs := &actionPromptObserver{}
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Fighter", "fighter", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		fighterTestCard("m1", "圣光", model.CardTypeMagic, model.ElementLight, 0),
	}
	p1.Form = model.FormFighterHundredDragon
	p1.TurnState.SkillFlowState["fighter_hundred_dragon_target_order"] = 2
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	game.Drive()
	prompt := obs.lastPrompt
	if prompt == nil {
		t.Fatalf("expected action prompt")
	}
	if prompt.Type != model.PromptConfirm {
		t.Fatalf("expected confirm prompt, got %s", prompt.Type)
	}

	labels := map[string]string{}
	for _, option := range prompt.Options {
		labels[option.ID] = option.Label
	}
	if _, ok := labels["magic"]; ok {
		t.Fatalf("expected hundred dragon prompt hide magic option, got %+v", labels)
	}
	if _, ok := labels["special"]; ok {
		t.Fatalf("expected hundred dragon prompt hide special option, got %+v", labels)
	}
	if len(prompt.SpecialOptions) != 0 {
		t.Fatalf("expected hundred dragon prompt hide special action entries, got %+v", prompt.SpecialOptions)
	}
	if _, ok := labels["attack"]; !ok {
		t.Fatalf("expected hundred dragon prompt keep attack option, got %+v", labels)
	}
}

func TestFighterBurstCrash_NoCounterAndSelfDamage(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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
	game.State.TurnStage = model.TurnStageActionExecution

	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdAttack, TargetID: "p2", CardIndex: 0})
	testutils.RequireResponseSkillPrompt(t, game, "p1")
	// 蓄力一击与气绝崩击互斥应在同一面板内同时提供（含跳过），由玩家三选一。
	if got := game.State.PendingInterrupt.SkillIDs; len(got) != 2 ||
		!containsString(got, "fighter_charge_strike") ||
		!containsString(got, "fighter_burst_crash") {
		t.Fatalf("expected attack-start prompt to offer both charge strike and burst crash, got %+v", got)
	}
	testutils.ChooseResponseSkillByID(t, game, "p1", "fighter_burst_crash")

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

	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p2", Type: model.CmdRespond, ExtraArgs: []string{"take"}})

	if got := p1.Tokens["fighter_qi"]; got != 1 {
		t.Fatalf("expected qi reduced to 1 after burst crash, got %d", got)
	}
	if got := p1.TurnState.SkillFlowState["fighter_qiburst_force_no_counter"]; got != 0 {
		t.Fatalf("expected no-counter token consumed, got %d", got)
	}
	if got := len(p1.Hand); got != 1 {
		t.Fatalf("expected fighter self-damage draw 1 card after burst crash, got hand=%d", got)
	}
}

func TestFighterWarGodDrive_DiscardToThreeAndHeal(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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
	game.State.TurnStage = model.TurnStageActionStart

	game.Drive()
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptStartupSkill {
		t.Fatalf("expected startup interrupt before confirming fighter_war_god_drive")
	}
	if err := game.ConfirmStartupSkill("p1", "fighter_war_god_drive"); err != nil {
		t.Fatalf("confirm fighter_war_god_drive failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || !engine.IsDiscardSelectionInterrupt(game.State.PendingInterrupt) {
		t.Fatalf("expected discard interrupt for war_god_drive continuation")
	}

	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0, 1}})

	if got := len(p1.Hand); got != 3 {
		t.Fatalf("expected hand size 3 after war_god_drive discard, got %d", got)
	}
	if got := p1.Heal; got != 2 {
		t.Fatalf("expected heal +2 from war_god_drive, got %d", got)
	}
	if got := p1.Crystal; got != 0 {
		t.Fatalf("expected war_god_drive consume 1 crystal-like, crystal=%d", got)
	}
	if got := game.State.CurrentTurn; got != 0 {
		t.Fatalf("expected fighter turn continue after war_god_drive, current_turn=%d", got)
	}
	if got := game.State.TurnStage; got != model.TurnStageActionExecution {
		t.Fatalf("expected action execution window after war_god_drive continuation, got %s", got)
	}
}
