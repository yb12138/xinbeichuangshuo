package model

import "testing"

func TestContextTimingHelpersPreferRulebookTimings(t *testing.T) {
	cases := []struct {
		name string
		ctx  *Context
		want func(*Context) bool
	}{
		{name: "attack declare", ctx: &Context{Timing: TimingAttackDeclare}, want: (*Context).AttackDeclaredPhase},
		{name: "attack committed", ctx: &Context{Timing: TimingAttackCommitted}, want: (*Context).AttackCommittedPhase},
		{name: "attack hit", ctx: &Context{Timing: TimingAttackHit}, want: (*Context).AttackHitPhase},
		{name: "attack miss", ctx: &Context{Timing: TimingAttackMiss}, want: (*Context).AttackMissPhase},
		{name: "magic resolve", ctx: &Context{Timing: TimingMagicResolve}, want: (*Context).MagicResolvePhase},
		{name: "damage taken", ctx: &Context{Timing: TimingDamageTaken}, want: (*Context).ResumeDamageTakenPhase},
		{name: "morale check", ctx: &Context{Timing: TimingMoraleLossCheck}, want: (*Context).ResumeBeforeMoraleLossPhase},
		{name: "action end", ctx: &Context{Timing: TimingActionEnd}, want: (*Context).ResumeActionEndPhase},
		{name: "turn start", ctx: &Context{Timing: TimingTurnStart}, want: (*Context).TurnStartOrStartupWindow},
		{name: "action start", ctx: &Context{Timing: TimingActionStart}, want: (*Context).TurnStartOrStartupWindow},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.want(tc.ctx) {
				t.Fatalf("helper returned false for timing %s", tc.ctx.Timing)
			}
		})
	}
}

func TestContextTimingHelpersKeepLegacyFallbacks(t *testing.T) {
	hitCtx := &Context{Timing: TimingOnHitCheck, EventCtx: &EventContext{AttackInfo: &AttackEventInfo{IsHit: true}}}
	if !hitCtx.AttackHitPhase() || !hitCtx.ResumeAttackHitPhase() {
		t.Fatalf("legacy hit-check hit context should remain compatible")
	}
	if hitCtx.AttackMissPhase() || hitCtx.ResumeAttackMissPhase() {
		t.Fatalf("legacy hit context should not be a miss")
	}

	missCtx := &Context{Timing: TimingOnHitCheck, EventCtx: &EventContext{AttackInfo: &AttackEventInfo{IsHit: false}}}
	if !missCtx.AttackMissPhase() || !missCtx.ResumeAttackMissPhase() {
		t.Fatalf("legacy hit-check miss context should remain compatible")
	}
	if missCtx.AttackHitPhase() || missCtx.ResumeAttackHitPhase() {
		t.Fatalf("legacy miss context should not be a hit")
	}

	if !(&Context{Timing: TimingOnAttackDeclared}).AttackDeclaredPhase() {
		t.Fatalf("legacy attack declared context should remain compatible")
	}
	if !(&Context{Timing: TimingOnDamageTaken}).ResumeDamageTakenPhase() {
		t.Fatalf("legacy damage taken context should remain compatible")
	}
	if !(&Context{Timing: TimingBeforeMoraleLoss}).ResumeBeforeMoraleLossPhase() {
		t.Fatalf("legacy morale loss context should remain compatible")
	}
	if !(&Context{Timing: TimingOnActionEnd}).ResumeActionEndPhase() {
		t.Fatalf("legacy action end context should remain compatible")
	}
	if !(&Context{Timing: TimingStartup}).TurnStartOrStartupWindow() {
		t.Fatalf("legacy startup context should remain compatible")
	}
}
