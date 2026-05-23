// gameflow: 规则书 Timing 统一分发入口。

package engine

import "starcup-engine/internal/model"

type ruleTimingDispatchInput struct {
	Timing   model.Timing
	User     *model.Player
	Target   *model.Player
	EventCtx *model.EventContext
	Context  *model.Context
	Markers  map[string]any
	Flags    map[string]bool
}

type ruleTimingDispatchResult struct {
	Dispatched      bool
	PendingChanged  bool
	Interrupted     bool
	QueuedInterrupt bool
	Context         *model.Context
}

func (e *GameEngine) dispatchRuleTiming(input ruleTimingDispatchInput) ruleTimingDispatchResult {
	if e == nil || e.State == nil || e.dispatcher == nil {
		return ruleTimingDispatchResult{}
	}
	if _, ok := model.TimingDescriptorOf(input.Timing); !ok {
		return ruleTimingDispatchResult{}
	}

	revisionBefore := e.State.InterruptRevision
	queueLenBefore := len(e.State.InterruptQueue)
	ctx := input.Context
	if ctx == nil {
		if input.User == nil {
			return ruleTimingDispatchResult{}
		}
		ctx = e.BuildContext(input.User, input.Target, input.Timing, input.EventCtx)
	} else {
		ctx.Timing = input.Timing
		if input.User != nil {
			ctx.User = input.User
		}
		if input.Target != nil {
			ctx.Target = input.Target
		}
		if input.EventCtx != nil {
			ctx.EventCtx = input.EventCtx
		}
	}
	if ctx.User == nil {
		return ruleTimingDispatchResult{}
	}
	if ctx.Selections == nil {
		ctx.Selections = map[string]any{}
	}
	if ctx.Flags == nil {
		ctx.Flags = map[string]bool{}
	}
	ctx.Selections["rulebook_timing"] = input.Timing
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
		Context:         ctx,
	}
}
