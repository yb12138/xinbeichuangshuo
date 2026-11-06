package sage_test

import (
	"reflect"
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/testutils"
)

func TestSageSkillActivated_WisdomCodexOnHeavyMagicDamage(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	g := engine.NewGameEngine(obs)
	if err := g.AddPlayer("p1", "Sage", "sage", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := g.State.Players["p1"]
	p1.Hand = []model.Card{sageTestCard("f1", "火焰斩", model.CardTypeAttack, model.ElementFire)}
	g.State.Deck = []model.Card{
		sageTestCard("d1", "补牌1", model.CardTypeAttack, model.ElementWater),
		sageTestCard("d2", "补牌2", model.CardTypeAttack, model.ElementEarth),
		sageTestCard("d3", "补牌3", model.CardTypeAttack, model.ElementWind),
		sageTestCard("d4", "补牌4", model.CardTypeAttack, model.ElementThunder),
	}
	g.AddPendingDamage(model.PendingDamage{SourceID: "p2", TargetID: "p1", Damage: 4, DamageType: model.MagicAttack})
	g.State.CombatStage = model.CombatStageCalcDamage
	runUntilChoiceInterrupt(g, 16)
	if !engine.IsDiscardSelectionInterrupt(g.State.PendingInterrupt) {
		t.Fatalf("expected wisdom codex discard interrupt, got %+v", g.State.PendingInterrupt)
	}
	testutils.MustHandleAction(t, g, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0}})

	got := obs.SkillActivated("sage_wisdom_codex")
	if len(got) != 1 {
		t.Fatalf("expected 1 sage_wisdom_codex skill_activated, got %d: %+v", len(got), got)
	}
	if got[0].PlayerID != "p1" || got[0].SkillName != "智慧法典" || !reflect.DeepEqual(got[0].TargetIDs, []string{"p1"}) {
		t.Fatalf("unexpected wisdom codex payload: %+v", got[0])
	}
}

func TestSageSkillActivated_WisdomCodexNotOnLightDamage(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	g := engine.NewGameEngine(obs)
	if err := g.AddPlayer("p1", "Sage", "sage", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	g.State.Deck = []model.Card{
		sageTestCard("d1", "补牌1", model.CardTypeAttack, model.ElementWater),
		sageTestCard("d2", "补牌2", model.CardTypeAttack, model.ElementEarth),
		sageTestCard("d3", "补牌3", model.CardTypeAttack, model.ElementWind),
	}
	g.AddPendingDamage(model.PendingDamage{SourceID: "p2", TargetID: "p1", Damage: 3, DamageType: model.MagicAttack})
	g.State.CombatStage = model.CombatStageCalcDamage
	runUntilChoiceInterrupt(g, 16)
	if got := obs.SkillActivated("sage_wisdom_codex"); len(got) != 0 {
		t.Fatalf("wisdom codex must not emit for damage<=3, got %+v", got)
	}
}

func TestSageSkillActivated_MagicReboundOnTargetConfirm(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	g := engine.NewGameEngine(obs)
	if err := g.AddPlayer("p1", "Sage", "sage", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := g.State.Players["p1"]
	p1.Hand = []model.Card{
		sageTestCard("f1", "火焰斩", model.CardTypeAttack, model.ElementFire),
		sageTestCard("f2", "烈焰击", model.CardTypeAttack, model.ElementFire),
	}
	g.State.Deck = []model.Card{sageTestCard("d1", "水涟斩", model.CardTypeAttack, model.ElementWater)}
	g.AddPendingDamage(model.PendingDamage{SourceID: "p2", TargetID: "p1", Damage: 1, DamageType: model.MagicAttack})
	g.State.CombatStage = model.CombatStageCalcDamage
	runUntilChoiceInterrupt(g, 12)

	// confirm -> element -> cards
	for _, sel := range [][]int{{0}, {0}, {0, 1}} {
		if err := g.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: sel}); err != nil {
			t.Fatalf("rebound step %v failed: %v", sel, err)
		}
		if got := obs.SkillActivated("sage_magic_rebound"); len(got) != 0 {
			t.Fatalf("magic rebound must not emit before target chosen, got %+v", got)
		}
	}
	if got := testutils.ChoiceTypeOfInterrupt(g.State.PendingInterrupt); got != "sage_magic_rebound_target" {
		t.Fatalf("expected rebound target prompt, got %q", got)
	}
	// 目标池 [p1, p2]，选 p2。
	if err := g.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{1}}); err != nil {
		t.Fatalf("choose rebound target failed: %v", err)
	}
	got := obs.SkillActivated("sage_magic_rebound")
	if len(got) != 1 {
		t.Fatalf("expected 1 sage_magic_rebound skill_activated, got %d: %+v", len(got), got)
	}
	if got[0].PlayerID != "p1" || got[0].SkillName != "法术反弹" || !reflect.DeepEqual(got[0].TargetIDs, []string{"p2", "p1"}) {
		t.Fatalf("unexpected magic rebound payload: %+v", got[0])
	}
}
