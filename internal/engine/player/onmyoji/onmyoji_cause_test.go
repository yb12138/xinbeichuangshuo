package onmyoji_test

import (
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/testutils"
)

// 式神咒束：p1 攻击 p2，阴阳师 p3 代为应战并反弹给 p4。
// 期望发出 attack_taken_over 语义事件，归属式神咒束（持有者 p3），涉及原目标 p2 与新目标 p4。
func TestOnmyojiBinding_EmitsAttackTakenOverUnderBindingCause(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	_ = game.AddPlayer("p1", "Attacker", "berserker", model.RedCamp)
	_ = game.AddPlayer("p2", "TargetAlly", "angel", model.BlueCamp)
	_ = game.AddPlayer("p3", "Onmyoji", "onmyoji", model.BlueCamp)
	_ = game.AddPlayer("p4", "AttackerAlly", "angel", model.RedCamp)
	p3 := game.State.Players["p3"]
	p3.Form = model.FormOnmyojiShikigami
	p3.Hand = []model.Card{{ID: "c1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Faction: "咏", Damage: 2}}
	game.State.BlueGems = 1
	game.State.BlueCrystals = 1
	game.State.CombatStack = []model.CombatRequest{{
		AttackerID:     "p1",
		TargetID:       "p2",
		Card:           &model.Card{ID: "atk", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Faction: "咏", Damage: 2},
		CanBeResponded: true,
	}}
	req := game.State.CombatStack[0]
	if !game.RunAttackResponseCombatInteractionPolicies(&req) {
		t.Fatalf("binding should start")
	}
	for _, step := range []string{"onmyoji_binding_confirm", "onmyoji_binding_card", "onmyoji_binding_counter_target"} {
		testutils.RequireChoicePrompt(t, game, "p3", step)
		testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p3", Type: model.CmdSelect, Selections: []int{0}})
	}

	var takeover *model.GameEvent
	for i := range obs.Events {
		ev := &obs.Events[i]
		if ev.Type == model.EventTimelineMarker && ev.TimelineMarker != nil && ev.TimelineMarker.NarrativeKind == "attack_taken_over" {
			takeover = ev
		}
	}
	if takeover == nil {
		t.Fatalf("expected attack_taken_over marker")
	}
	if takeover.Cause == nil || takeover.Cause.SkillID != "onmyoji_binding" || takeover.Cause.OwnerID != "p3" {
		t.Fatalf("expected takeover under onmyoji_binding cause owned by p3, got %+v", takeover.Cause)
	}
	got := takeover.TimelineMarker.TargetIDs
	if len(got) != 2 || got[0] != "p2" || got[1] != "p4" {
		t.Fatalf("expected takeover targets [p2 p4], got %v", got)
	}
	if len(game.State.CombatStack) == 0 || game.State.CombatStack[len(game.State.CombatStack)-1].Cause == nil {
		t.Fatalf("expected reflected combat to carry the binding cause")
	}
}
