package skill

import (
	"reflect"
	"testing"

	"starcup-engine/internal/model"
)

type timingTestHost struct {
	state *model.GameState
}

func (h timingTestHost) GameState() *model.GameState { return h.state }

func TestTargetsForTimingUsesRulebookAttackRoles(t *testing.T) {
	host, attacker, defender := newTimingTestHost()
	ctx := &model.Context{User: attacker, Target: defender}

	for _, timing := range []model.FlowTiming{
		model.TimingAttackDeclare,
		model.TimingAttackCommitted,
		model.TimingAttackForceHitCheck,
		model.TimingAttackResponse,
		model.TimingAttackHit,
		model.TimingAttackMiss,
	} {
		got := targetIDs(targetsForTiming(host, timing, ctx))
		want := []string{"p1:Attacker", "p2:Defender"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%s targets = %v, want %v", timing, got, want)
		}
	}
}

func TestTargetsForTimingKeepsLegacyAttackRoles(t *testing.T) {
	host, attacker, defender := newTimingTestHost()
	ctx := &model.Context{User: attacker, Target: defender}

	got := targetIDs(targetsForTiming(host, model.TimingOnHitCheck, ctx))
	want := []string{"p1:Attacker", "p2:Defender"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("legacy hit-check targets = %v, want %v", got, want)
	}
}

func TestTargetsForTimingUsesRulebookDamageTakenRoles(t *testing.T) {
	host, source, target := newTimingTestHost()
	ctx := &model.Context{User: target, Target: source}

	got := targetIDs(targetsForTiming(host, model.TimingDamageTaken, ctx))
	want := []string{"p2:Defender", "p1:Attacker"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("damage taken targets = %v, want %v", got, want)
	}
}

func TestTargetsForTimingUsesSeatOrderForMoraleCamp(t *testing.T) {
	host, _, victim := newTimingTestHost()
	host.state.Players["p3"] = &model.Player{ID: "p3", Camp: model.BlueCamp, Character: &model.Character{}}
	host.state.PlayerOrder = []string{"p1", "p2", "p3"}
	ctx := &model.Context{User: victim}

	got := targetIDs(targetsForTiming(host, model.TimingMoraleLossCheck, ctx))
	want := []string{"p2:<any>", "p3:<any>"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("morale targets = %v, want %v", got, want)
	}
	if !timingPriorityOrdered(model.TimingMoraleLossCheck) || !timingPriorityOrdered(model.TimingBeforeMoraleLoss) {
		t.Fatalf("morale timing should be priority ordered")
	}
}

func newTimingTestHost() (timingTestHost, *model.Player, *model.Player) {
	state := model.NewGameState()
	state.PlayerOrder = []string{"p1", "p2"}
	state.CurrentTurn = 0
	p1 := &model.Player{ID: "p1", Camp: model.RedCamp, Character: &model.Character{}}
	p2 := &model.Player{ID: "p2", Camp: model.BlueCamp, Character: &model.Character{}}
	state.Players["p1"] = p1
	state.Players["p2"] = p2
	return timingTestHost{state: state}, p1, p2
}

func targetIDs(targets []checkTarget) []string {
	out := make([]string, 0, len(targets))
	for _, target := range targets {
		if target.Player == nil {
			continue
		}
		role := string(target.Role)
		if target.Role == model.RoleAny {
			role = "<any>"
		}
		out = append(out, target.Player.ID+":"+role)
	}
	return out
}
