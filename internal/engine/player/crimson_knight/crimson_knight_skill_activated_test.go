package crimson_knight_test

import (
	"reflect"
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/testutils"
)

func requireHotBloodActivations(t *testing.T, obs *testutils.CaptureObserver, playerID string, want int) {
	t.Helper()
	got := obs.SkillActivated("crk_hot_blood")
	if len(got) != want {
		t.Fatalf("expected %d crk_hot_blood skill_activated, got %d: %+v", want, len(got), got)
	}
	for _, p := range got {
		if p.PlayerID != playerID || p.SkillName != "热血沸腾" || !reflect.DeepEqual(p.TargetIDs, []string{playerID}) {
			t.Fatalf("unexpected crk_hot_blood payload: %+v", p)
		}
	}
}

func TestCrimsonKnightSkillActivated_HotBloodEnterOnMoraleLoss(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	g := engine.NewGameEngine(obs)
	if err := g.AddPlayer("p1", "Crimson", "crimson_knight", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := g.State.Players["p1"]
	// 敌方回合行动中承伤爆牌，避免恢复主流程时走到红莲骑士自己的回合结束而脱离形态。
	g.State.CurrentTurn = 1
	g.State.TurnStage = model.TurnStageActionExecution
	g.State.Players["p2"].IsActive = true
	p1.MaxHand = 6
	p1.Hand = makeHandCards(8, model.ElementFire)
	overflowCtx := g.BuildContext(p1, nil, model.TimingActionDuring, nil)
	overflowCtx.Flags["FromDamageDraw"] = true
	g.CheckHandLimitCtx(p1, overflowCtx)
	if g.State.PendingInterrupt == nil || !engine.IsDiscardSelectionInterrupt(g.State.PendingInterrupt) {
		t.Fatalf("expected discard interrupt, got %+v", g.State.PendingInterrupt)
	}
	testutils.MustHandleAction(t, g, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0, 1}})
	if p1.Form != model.FormCrimsonKnightHotBlooded {
		t.Fatalf("expected hot blooded form, got %q", p1.Form)
	}
	requireHotBloodActivations(t, obs, "p1", 1)
}

func TestCrimsonKnightSkillActivated_HotBloodReleaseOnTurnEnd(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	g := engine.NewGameEngine(obs)
	if err := g.AddPlayer("p1", "Crimson", "crimson_knight", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := g.State.Players["p1"]
	p1.IsActive = true
	p1.Form = model.FormCrimsonKnightHotBlooded
	g.State.CurrentTurn = 0
	g.State.TurnStage = model.TurnStageTurnEnd
	g.Drive()

	if p1.Form != "" || p1.Heal != 2 {
		t.Fatalf("expected hot form released with heal 2, form=%q heal=%d", p1.Form, p1.Heal)
	}
	requireHotBloodActivations(t, obs, "p1", 1)
}

func TestCrimsonKnightSkillActivated_NoEmitAtTurnEndWithoutHotBlood(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	g := engine.NewGameEngine(obs)
	if err := g.AddPlayer("p1", "Crimson", "crimson_knight", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	g.State.Players["p1"].IsActive = true
	g.State.CurrentTurn = 0
	g.State.TurnStage = model.TurnStageTurnEnd
	g.Drive()
	requireHotBloodActivations(t, obs, "p1", 0)
}
