package engine

import (
	"testing"

	"starcup-engine/internal/model"
)

func TestMagicSwordsmanAsuraCombo_TriggersWithoutFireAttackInHand(t *testing.T) {
	g := NewGameEngine(nil)
	if err := g.AddPlayer("p1", "MS", "magic_swordsman", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Dummy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{ID: "holy", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight},
	}

	g.State.CurrentTurn = 0
	g.State.TurnStage = model.TurnStageExtraAction
	p1.TurnState.LastActionType = string(model.ActionAttack)

	g.Drive()

	if g.State.PendingInterrupt == nil || g.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected asura combo response interrupt, got %+v", g.State.PendingInterrupt)
	}
	found := false
	for _, skillID := range g.State.PendingInterrupt.SkillIDs {
		if skillID == "ms_asura_combo" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected ms_asura_combo without fire attack in hand, got %+v", g.State.PendingInterrupt.SkillIDs)
	}
}

func TestMagicSwordsmanShadowGather_ReleasesBeforeNextActionSelectionPrompt(t *testing.T) {
	obs := &promptCaptureObserver{}
	g := NewGameEngine(obs)
	if err := g.AddPlayer("p1", "MS", "magic_swordsman", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Dummy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	enterMagicSwordsmanShadowForm(p1)
	p1.Hand = []model.Card{
		{ID: "holy", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight},
	}

	g.State.CurrentTurn = 0
	g.State.TurnStage = model.TurnStageActionStart

	g.Drive()

	if p1.Form != "" {
		t.Fatalf("expected shadow form released before action selection, got form=%q", p1.Form)
	}
	if g.State.PendingInterrupt == nil || g.State.PendingInterrupt.Type != model.InterruptStartupSkill {
		t.Fatalf("expected startup skill prompt after action-start release, got %+v", g.State.PendingInterrupt)
	}
	mustHandleAction(t, g, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{len(g.State.PendingInterrupt.SkillIDs)},
	})
	if obs.lastPrompt == nil {
		t.Fatalf("expected action selection prompt")
	}
	hasMagic := false
	for _, opt := range obs.lastPrompt.Options {
		if opt.ID == "magic" {
			hasMagic = true
			break
		}
	}
	if !hasMagic {
		t.Fatalf("expected magic option after shadow release, got %+v", obs.lastPrompt.Options)
	}
}

func TestMagicSwordsmanShadowMeteor_TargetsEnemyOnlyAndUsesUnifiedSkillFlow(t *testing.T) {
	g := NewGameEngine(nil)
	if err := g.AddPlayer("p1", "MS", "magic_swordsman", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p3", "Ally", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	enterMagicSwordsmanShadowForm(p1)
	p1.Hand = []model.Card{
		{ID: "m1", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight},
		{ID: "m2", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 2},
	}
	g.State.RedGems = 2

	g.State.CurrentTurn = 0
	g.State.TurnStage = model.TurnStageActionExecution

	err := g.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSkill,
		SkillID:    "ms_shadow_meteor",
		TargetIDs:  []string{"p3"},
		Selections: []int{0, 1},
	})
	if err == nil {
		t.Fatalf("expected ally target to be rejected for shadow meteor")
	}

	if err := g.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSkill,
		SkillID:    "ms_shadow_meteor",
		TargetIDs:  []string{"p2"},
		Selections: []int{0, 1},
	}); err != nil {
		t.Fatalf("shadow meteor failed: %v", err)
	}

	if g.State.PendingInterrupt == nil || g.State.PendingInterrupt.Type != model.InterruptChoice {
		t.Fatalf("expected release confirm interrupt after shadow meteor, got %+v", g.State.PendingInterrupt)
	}
	ctx, _ := g.State.PendingInterrupt.Context.(map[string]interface{})
	if ct, _ := ctx["choice_type"].(string); ct != "ms_shadow_meteor_release_confirm" {
		t.Fatalf("expected release confirm choice, got %+v", ctx)
	}
	if len(g.State.PendingDamageQueue) == 0 {
		t.Fatalf("expected pending damage queued by shadow meteor")
	}
	pd := g.State.PendingDamageQueue[0]
	if pd.TargetID != "p2" {
		t.Fatalf("expected shadow meteor target p2, got %+v", pd)
	}

	if err := g.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("confirm shadow meteor release failed: %v", err)
	}

	if p1.Form != "" {
		t.Fatalf("expected release confirm clear shadow form, got form=%q", p1.Form)
	}
	if p1.Gem != 1 {
		t.Fatalf("expected gain 1 gem after release confirm, got %d", p1.Gem)
	}
	if g.State.RedGems != 0 {
		t.Fatalf("expected spend 2 camp gems, got %d", g.State.RedGems)
	}
}
