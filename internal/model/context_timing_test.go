package model

import "testing"

func TestContextTimingHelpersPreferRulebookTimings(t *testing.T) {
	cases := []struct {
		name string
		ctx  *Context
		want func(*Context) bool
	}{
		{name: "attack declare", ctx: &Context{Timing: TimingAttackDeclare}, want: (*Context).AttackDeclarePhase},
		{name: "attack declared compatibility", ctx: &Context{Timing: TimingAttackDeclare}, want: (*Context).AttackDeclaredPhase},
		{name: "attack select target", ctx: &Context{Timing: TimingAttackSelectTarget}, want: (*Context).AttackSelectTargetPhase},
		{name: "attack play card", ctx: &Context{Timing: TimingAttackPlayCard}, want: (*Context).AttackPlayCardPhase},
		{name: "attack modify card", ctx: &Context{Timing: TimingAttackModifyCard}, want: (*Context).AttackModifyCardPhase},
		{name: "attack committed", ctx: &Context{Timing: TimingAttackCommitted}, want: (*Context).AttackCommittedPhase},
		{name: "attack force hit check", ctx: &Context{Timing: TimingAttackForceHitCheck}, want: (*Context).AttackForceHitCheckPhase},
		{name: "attack no response check", ctx: &Context{Timing: TimingAttackNoResponseCheck}, want: (*Context).AttackNoResponseCheckPhase},
		{name: "attack response", ctx: &Context{Timing: TimingAttackResponse}, want: (*Context).AttackResponsePhase},
		{name: "attack hit", ctx: &Context{Timing: TimingAttackHit}, want: (*Context).AttackHitPhase},
		{name: "attack miss", ctx: &Context{Timing: TimingAttackMiss}, want: (*Context).AttackMissPhase},
		{name: "magic declare", ctx: &Context{Timing: TimingMagicDeclare}, want: (*Context).MagicDeclarePhase},
		{name: "magic resolve", ctx: &Context{Timing: TimingMagicResolve}, want: (*Context).MagicResolvePhase},
		{name: "magic missile response", ctx: &Context{Timing: TimingMagicMissileResponse}, want: (*Context).MagicMissileResponsePhase},
		{name: "magic missile defend", ctx: &Context{Timing: TimingMagicMissileDefend}, want: (*Context).MagicMissileDefendPhase},
		{name: "magic missile counter", ctx: &Context{Timing: TimingMagicMissileCounter}, want: (*Context).MagicMissileCounterPhase},
		{name: "magic missile response skill", ctx: &Context{Timing: TimingMagicMissileResponseSkill}, want: (*Context).MagicMissileResponseSkillPhase},
		{name: "damage source deal", ctx: &Context{Timing: TimingDamageSourceDeal}, want: (*Context).DamageSourceDealPhase},
		{name: "damage target before", ctx: &Context{Timing: TimingDamageTargetBefore}, want: (*Context).DamageTargetBeforePhase},
		{name: "damage applied", ctx: &Context{Timing: TimingDamageApplied}, want: (*Context).DamageAppliedPhase},
		{name: "damage taken", ctx: &Context{Timing: TimingDamageTaken}, want: (*Context).DamageTakenPhase},
		{name: "resume damage taken", ctx: &Context{Timing: TimingDamageTaken}, want: (*Context).ResumeDamageTakenPhase},
		{name: "damage resolved", ctx: &Context{Timing: TimingDamageResolved}, want: (*Context).DamageResolvedPhase},
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

func TestContextTimingHelpersHandleNilContext(t *testing.T) {
	var ctx *Context
	helpers := []struct {
		name string
		call func() bool
	}{
		{name: "attack declare", call: ctx.AttackDeclarePhase},
		{name: "attack declared compatibility", call: ctx.AttackDeclaredPhase},
		{name: "attack select target", call: ctx.AttackSelectTargetPhase},
		{name: "attack play card", call: ctx.AttackPlayCardPhase},
		{name: "attack modify card", call: ctx.AttackModifyCardPhase},
		{name: "attack committed", call: ctx.AttackCommittedPhase},
		{name: "attack force hit check", call: ctx.AttackForceHitCheckPhase},
		{name: "attack no response check", call: ctx.AttackNoResponseCheckPhase},
		{name: "attack response", call: ctx.AttackResponsePhase},
		{name: "attack hit", call: ctx.AttackHitPhase},
		{name: "attack miss", call: ctx.AttackMissPhase},
		{name: "magic declare", call: ctx.MagicDeclarePhase},
		{name: "magic resolve", call: ctx.MagicResolvePhase},
		{name: "magic missile response", call: ctx.MagicMissileResponsePhase},
		{name: "magic missile defend", call: ctx.MagicMissileDefendPhase},
		{name: "magic missile counter", call: ctx.MagicMissileCounterPhase},
		{name: "magic missile response skill", call: ctx.MagicMissileResponseSkillPhase},
		{name: "damage source deal", call: ctx.DamageSourceDealPhase},
		{name: "damage target before", call: ctx.DamageTargetBeforePhase},
		{name: "damage applied", call: ctx.DamageAppliedPhase},
		{name: "damage taken", call: ctx.DamageTakenPhase},
		{name: "damage resolved", call: ctx.DamageResolvedPhase},
	}
	for _, helper := range helpers {
		t.Run(helper.name, func(t *testing.T) {
			if helper.call() {
				t.Fatalf("helper returned true for nil context")
			}
		})
	}
}
