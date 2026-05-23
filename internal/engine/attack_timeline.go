// gameflow: 攻击/应战规则书时间轴。

package engine

import "starcup-engine/internal/model"

func attackKindFromCounter(isCounter bool) model.AttackKind {
	if isCounter {
		return model.AttackKindCounter
	}
	return model.AttackKindActive
}

func (e *GameEngine) dispatchAttackRulebookTiming(timing model.Timing, attacker, target *model.Player, card *model.Card, attackInfo *model.AttackEventInfo, kind model.AttackKind) bool {
	return e.dispatchAttackRulebookTimingWithMarkers(timing, attacker, target, card, attackInfo, kind, nil).Interrupted
}

func (e *GameEngine) dispatchAttackRulebookTimingWithMarkers(timing model.Timing, attacker, target *model.Player, card *model.Card, attackInfo *model.AttackEventInfo, kind model.AttackKind, markers map[string]any) ruleTimingDispatchResult {
	ctx := e.buildAttackRulebookContext(timing, attacker, target, card, attackInfo, kind)
	if ctx == nil {
		return ruleTimingDispatchResult{}
	}
	return e.dispatchAttackRulebookEventTimingWithMarkers(timing, ctx.User, ctx.Target, ctx.EventCtx, kind, markers)
}

func (e *GameEngine) dispatchAttackRulebookEventTimingWithMarkers(timing model.Timing, attacker, target *model.Player, eventCtx *model.EventContext, kind model.AttackKind, markers map[string]any) ruleTimingDispatchResult {
	if e == nil || e.State == nil || attacker == nil || eventCtx == nil {
		return ruleTimingDispatchResult{}
	}
	if eventCtx.Type == model.EventNone {
		eventCtx.Type = model.EventAttack
	}
	if eventCtx.SourceID == "" {
		eventCtx.SourceID = attacker.ID
	}
	if target != nil && eventCtx.TargetID == "" {
		eventCtx.TargetID = target.ID
	}
	if eventCtx.ActionType == "" {
		eventCtx.ActionType = model.ActionAttack
	}
	if eventCtx.AttackInfo == nil {
		eventCtx.AttackInfo = &model.AttackEventInfo{
			ActionType:       string(model.ActionAttack),
			CanBeResponded:   true,
			CounterInitiator: "",
			InterceptTags:    map[model.CombatInterceptTag]bool{},
		}
		if kind == model.AttackKindCounter {
			eventCtx.AttackInfo.CounterInitiator = attacker.ID
		}
	}
	allMarkers := map[string]any{
		"attack_timeline": true,
		"attack_kind":     kind,
	}
	for key, value := range markers {
		allMarkers[key] = value
	}
	result := e.dispatchRuleTiming(ruleTimingDispatchInput{
		Timing:   timing,
		User:     attacker,
		Target:   target,
		EventCtx: eventCtx,
		Markers:  allMarkers,
	})
	return result
}

func (e *GameEngine) buildAttackRulebookContext(timing model.Timing, attacker, target *model.Player, card *model.Card, attackInfo *model.AttackEventInfo, kind model.AttackKind) *model.Context {
	if e == nil || e.State == nil || attacker == nil {
		return nil
	}
	targetID := ""
	if target != nil {
		targetID = target.ID
	}
	if attackInfo == nil {
		attackInfo = &model.AttackEventInfo{
			ActionType:       string(model.ActionAttack),
			CanBeResponded:   true,
			CounterInitiator: "",
			InterceptTags:    map[model.CombatInterceptTag]bool{},
		}
		if kind == model.AttackKindCounter && attacker != nil {
			attackInfo.CounterInitiator = attacker.ID
		}
	}
	ctx := e.BuildContext(attacker, target, timing, &model.EventContext{
		Type:       model.EventAttack,
		SourceID:   attacker.ID,
		TargetID:   targetID,
		Card:       card,
		ActionType: model.ActionAttack,
		AttackInfo: attackInfo,
	})
	return ctx
}

func attackInfoFromCombatRequest(req *model.CombatRequest, isHit bool) *model.AttackEventInfo {
	if req == nil {
		return &model.AttackEventInfo{ActionType: string(model.ActionAttack), IsHit: isHit}
	}
	info := &model.AttackEventInfo{
		IsHit:           isHit,
		IsHitForced:     req.IsForcedHit,
		IgnoreShield:    req.IgnoreShield,
		CanBeResponded:  req.CanBeResponded,
		ActionType:      string(model.ActionAttack),
		InterceptTags:   model.CloneCombatInterceptTags(req.InterceptTags),
		ElementOverride: req.ElementOverride,
	}
	if req.Card != nil {
		info.Element = string(req.Card.Element)
	}
	if req.IsCounter {
		info.CounterInitiator = req.AttackerID
	}
	return info
}
