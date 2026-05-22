// gameflow: 伤害、治疗、摸牌、弃牌、士气等共通结算规则书时间轴。

package engine

import "starcup-engine/internal/model"

const (
	pendingDamageCheckTimingDamageTargetBefore model.PendingDamageCheckKey = "timing.damage_target_before"
	pendingDamageCheckTimingHealBefore         model.PendingDamageCheckKey = "timing.heal_before"
	pendingDamageCheckTimingDamageApplied      model.PendingDamageCheckKey = "timing.damage_applied"
)

func (e *GameEngine) dispatchDamageRulebookTimingOnce(timing model.Timing, pd *model.PendingDamage, check model.PendingDamageCheckKey) bool {
	if pd == nil {
		return false
	}
	if pd.HasCheck(check) {
		return false
	}
	pd.SetCheck(check, true)
	return e.dispatchDamageRulebookTiming(timing, pd)
}

func (e *GameEngine) dispatchDamageRulebookTiming(timing model.Timing, pd *model.PendingDamage) bool {
	if e == nil || e.State == nil || pd == nil || e.dispatcher == nil {
		return false
	}
	source := e.State.Players[pd.SourceID]
	target := e.State.Players[pd.TargetID]
	user := target
	ctxTarget := source
	if timing == model.TimingDamageSourceDeal {
		user = source
		ctxTarget = target
	}
	if user == nil {
		return false
	}
	eventCtx := &model.EventContext{
		Type:      model.EventDamage,
		SourceID:  pd.SourceID,
		TargetID:  pd.TargetID,
		DamageVal: &pd.Damage,
		Card:      pd.Card,
	}
	ctx := e.BuildContext(user, ctxTarget, timing, eventCtx)
	ctx.Flags["IsMagicDamage"] = pd.DamageType != model.AttackDamage
	ctx.Selections["rulebook_timing"] = timing
	ctx.Selections["legacy_timing"] = model.LegacyTimingName(timing)
	ctx.Selections["settlement_timeline"] = true
	ctx.Selections["damage_type"] = pd.DamageType

	pendingBefore := e.State.PendingInterrupt
	queueLenBefore := len(e.State.InterruptQueue)
	e.dispatcher.OnTiming(timing, ctx)
	return e.State.PendingInterrupt != nil &&
		(e.State.PendingInterrupt != pendingBefore || len(e.State.InterruptQueue) != queueLenBefore)
}

func (e *GameEngine) dispatchSettlementRulebookTiming(timing model.Timing, user, target *model.Player, eventCtx *model.EventContext) bool {
	if e == nil || e.State == nil || user == nil || e.dispatcher == nil {
		return false
	}
	ctx := e.BuildContext(user, target, timing, eventCtx)
	ctx.Selections["rulebook_timing"] = timing
	ctx.Selections["legacy_timing"] = model.LegacyTimingName(timing)
	ctx.Selections["settlement_timeline"] = true

	pendingBefore := e.State.PendingInterrupt
	queueLenBefore := len(e.State.InterruptQueue)
	e.dispatcher.OnTiming(timing, ctx)
	return e.State.PendingInterrupt != nil &&
		(e.State.PendingInterrupt != pendingBefore || len(e.State.InterruptQueue) != queueLenBefore)
}
