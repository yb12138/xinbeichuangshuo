// gameflow: 主动法术规则书时间轴。

package engine

import "starcup-engine/internal/model"

func (e *GameEngine) dispatchMagicRulebookTiming(timing model.Timing, player, target *model.Player, card *model.Card) bool {
	ctx := e.buildMagicRulebookContext(timing, player, target, card)
	if ctx == nil {
		return false
	}
	result := e.dispatchRuleTiming(ruleTimingDispatchInput{
		Timing:   timing,
		User:     ctx.User,
		Target:   ctx.Target,
		EventCtx: ctx.EventCtx,
		Markers: map[string]any{
			"magic_timeline": true,
		},
	})
	return result.Interrupted
}

func (e *GameEngine) buildMagicRulebookContext(timing model.Timing, player, target *model.Player, card *model.Card) *model.Context {
	if e == nil || e.State == nil || player == nil {
		return nil
	}
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
	return ctx
}
