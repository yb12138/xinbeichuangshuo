// gameflow: 规则书 Timing 统一分发入口。

package engine

import "starcup-engine/internal/model"

type ruleTimingDispatchInput struct {
	Timing   model.Timing
	User     *model.Player
	Target   *model.Player
	EventCtx *model.EventContext
	Markers  map[string]any
	Flags    map[string]bool
}

type ruleTimingDispatchResult struct {
	Dispatched      bool
	PendingChanged  bool
	Interrupted     bool
	QueuedInterrupt bool
}

func (e *GameEngine) dispatchRuleTiming(input ruleTimingDispatchInput) ruleTimingDispatchResult {
	if e == nil || e.State == nil || e.dispatcher == nil || input.User == nil {
		return ruleTimingDispatchResult{}
	}
	if _, ok := model.TimingDescriptorOf(input.Timing); !ok {
		return ruleTimingDispatchResult{}
	}

	revisionBefore := e.State.InterruptRevision
	queueLenBefore := len(e.State.InterruptQueue)
	ctx := e.BuildContext(input.User, input.Target, input.Timing, input.EventCtx)
	ctx.Selections["rulebook_timing"] = input.Timing
	ctx.Selections["legacy_timing"] = model.LegacyTimingName(input.Timing)
	for key, value := range input.Markers {
		ctx.Selections[key] = value
	}
	for key, value := range input.Flags {
		ctx.Flags[key] = value
	}

	e.dispatcher.OnTiming(input.Timing, ctx)
	pendingChanged := e.State.InterruptRevision != revisionBefore
	return ruleTimingDispatchResult{
		Dispatched:      true,
		PendingChanged:  pendingChanged,
		Interrupted:     pendingChanged && e.State.PendingInterrupt != nil,
		QueuedInterrupt: pendingChanged && len(e.State.InterruptQueue) > queueLenBefore,
	}
}
