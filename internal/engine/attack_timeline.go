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
	if e == nil || e.State == nil || attacker == nil || e.dispatcher == nil {
		return false
	}
	pendingBefore := e.State.PendingInterrupt
	queueLenBefore := len(e.State.InterruptQueue)
	ctx := e.buildAttackRulebookContext(timing, attacker, target, card, attackInfo, kind)
	e.dispatcher.OnTiming(timing, ctx)
	return e.State.PendingInterrupt != nil &&
		(e.State.PendingInterrupt != pendingBefore || len(e.State.InterruptQueue) != queueLenBefore)
}

func (e *GameEngine) buildAttackRulebookContext(timing model.Timing, attacker, target *model.Player, card *model.Card, attackInfo *model.AttackEventInfo, kind model.AttackKind) *model.Context {
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
	ctx.Selections["rulebook_timing"] = timing
	ctx.Selections["legacy_timing"] = model.LegacyTimingName(timing)
	ctx.Selections["attack_timeline"] = true
	ctx.Selections["attack_kind"] = kind
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
