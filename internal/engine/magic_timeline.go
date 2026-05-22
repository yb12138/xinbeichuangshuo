// gameflow: 主动法术规则书时间轴。

package engine

import "starcup-engine/internal/model"

func (e *GameEngine) dispatchMagicRulebookTiming(timing model.Timing, player, target *model.Player, card *model.Card) bool {
	if e == nil || e.State == nil || player == nil || e.dispatcher == nil {
		return false
	}
	pendingBefore := e.State.PendingInterrupt
	queueLenBefore := len(e.State.InterruptQueue)
	ctx := e.buildMagicRulebookContext(timing, player, target, card)
	e.dispatcher.OnTiming(timing, ctx)
	return e.State.PendingInterrupt != nil &&
		(e.State.PendingInterrupt != pendingBefore || len(e.State.InterruptQueue) != queueLenBefore)
}

func (e *GameEngine) buildMagicRulebookContext(timing model.Timing, player, target *model.Player, card *model.Card) *model.Context {
	targetID := ""
	if target != nil {
		targetID = target.ID
	}
	ctx := e.BuildContext(player, target, timing, &model.EventContext{
		Type:       model.EventMagic,
		SourceID:   player.ID,
		TargetID:   targetID,
		Card:       card,
		ActionType: model.ActionMagic,
	})
	ctx.Selections["rulebook_timing"] = timing
	ctx.Selections["legacy_timing"] = model.LegacyTimingName(timing)
	ctx.Selections["magic_timeline"] = true
	return ctx
}
