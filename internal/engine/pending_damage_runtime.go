// gameflow: PendingDamage 队列：入队、结算、插入响应。

package engine

import (
	"fmt"
	"starcup-engine/internal/engine/core/runtimeutil"
	"strings"

	"starcup-engine/internal/model"
)

func (e *GameEngine) prependPendingDamages(pds []model.PendingDamage) {
	if len(pds) == 0 {
		return
	}
	reversed := make([]model.PendingDamage, 0, len(pds))
	for i := len(pds) - 1; i >= 0; i-- {
		reversed = append(reversed, pds[i])
	}
	e.State.PendingDamageQueue = append(reversed, e.State.PendingDamageQueue...)
	for _, pd := range reversed {
		e.Log(fmt.Sprintf("[System] 延迟伤害已前插: Source: %s, Target: %s, Damage: %d, Type: %s",
			pd.SourceID, pd.TargetID, pd.Damage, pd.DamageType))
	}
}

// syncPendingDamageRuntimeFromContext 将响应/被动技能在当前伤害上下文里写入的运行时元数据，
// 回填到正在处理的 PendingDamage 头结点，确保中断恢复后状态仍然存在。
func (e *GameEngine) syncPendingDamageRuntimeFromContext(ctx *model.Context) {
	if e == nil || ctx == nil || ctx.Timing != model.TimingOnDamageTaken || len(e.State.PendingDamageQueue) == 0 {
		return
	}

	pd := &e.State.PendingDamageQueue[0]
	if ctx.EventCtx != nil {
		if ctx.EventCtx.SourceID != "" && pd.SourceID != ctx.EventCtx.SourceID {
			return
		}
		if ctx.EventCtx.TargetID != "" && pd.TargetID != ctx.EventCtx.TargetID {
			return
		}
	}
	if ctx.Selections == nil {
		return
	}

	if raw, ok := ctx.Selections["overflow_morale_loss_fixed"]; ok {
		pd.OverflowMoraleLossFixed = runtimeutil.ToIntContextValue(raw)
	}
}

// checkAndProcessAttackHolyShield 检查并处理攻击圣盾（在加星石之前）。
// 返回 true 表示圣盾触发了，攻击被挡，按未命中处理。
func (e *GameEngine) checkAndProcessAttackHolyShield(pd *model.PendingDamage, attacker, victim *model.Player) bool {
	if pd == nil || victim == nil || attacker == nil {
		return false
	}
	if pd.IgnoreShield || pd.HasInterceptTag(model.CombatInterceptIgnoreHolyShield) {
		return false
	}

	var shieldToRemove *model.FieldCard
	for _, fc := range victim.Field {
		if fc != nil && fc.Mode == model.FieldEffect && fc.Effect == model.EffectShield {
			shieldToRemove = fc
			break
		}
	}
	if shieldToRemove == nil {
		return false
	}

	newField := make([]*model.FieldCard, 0, len(victim.Field))
	removed := false
	for _, fc := range victim.Field {
		if !removed && fc == shieldToRemove {
			removed = true
			if err := e.DiscardCard(fc); err != nil {
				e.Log(fmt.Sprintf("[Warning] 移除圣盾失败: %v", err))
			}
			continue
		}
		newField = append(newField, fc)
	}
	victim.Field = newField

	e.Log(fmt.Sprintf("[Shield] %s 的【圣盾】自动触发，抵消了攻击！", victim.Name))
	e.NotifyActionStep(fmt.Sprintf("%s 的【圣盾】触发，抵消了攻击", victim.Name))
	e.NotifyCombatCue(pd.SourceID, pd.TargetID, "shield")
	e.NotifyActionStep(fmt.Sprintf("%s 的【圣盾】抵消了本次攻击，判定为未命中", victim.Name))

	e.resolveMagicBowPierceMissWithOverride(
		pd.SourceID,
		pd.TargetID,
		pd.Card,
		pd.HasCheck(model.PendingDamageCheckHeroRoarMissArmed),
		pd.HasCheck(model.PendingDamageCheckFighterChargeMissArmed),
		pd.IsCounter,
	)
	pd.SetCheck(model.PendingDamageCheckAttackMissResolved, true)
	pd.AttackHitFlowDispatched = true
	return true
}

func (e *GameEngine) processPendingAttackHit(pd *model.PendingDamage) bool {
	if pd == nil || !strings.EqualFold(string(pd.DamageType), string(model.AttackDamage)) || pd.AttackHitFlowDispatched {
		return false
	}
	if pd.Card == nil {
		pd.AttackHitFlowDispatched = true
		return false
	}
	attacker := e.State.Players[pd.SourceID]
	victim := e.State.Players[pd.TargetID]
	if attacker == nil || victim == nil {
		pd.AttackHitFlowDispatched = true
		return false
	}

	e.runPendingDamageAttackLifecycle(pd, attacker, victim)
	if e.checkAndProcessAttackHolyShield(pd, attacker, victim) {
		return false
	}

	action := model.Action{
		SourceID: pd.SourceID,
		TargetID: pd.TargetID,
		Type:     model.ActionAttack,
		Card:     pd.Card,
		CounterInitiator: func() string {
			if pd.IsCounter {
				return pd.SourceID
			}
			return ""
		}(),
	}
	pd.Damage = e.ApplyAttackDamageModifiers(attacker, victim, pd.Damage, action)

	resourceType := "gem"
	if pd.IsCounter {
		resourceType = "crystal"
	}
	pd.AttackHitResourceType = resourceType
	pd.AttackHitResourceGranted = e.addCampResource(attacker.Camp, resourceType)
	if pd.AttackHitResourceGranted {
		if resourceType == "crystal" {
			e.Log(fmt.Sprintf("[Combat] 应战攻击命中！%s 方战绩区+1水晶", attacker.Camp))
		} else {
			e.Log(fmt.Sprintf("[Combat] 主动攻击命中！%s 方战绩区+1宝石", attacker.Camp))
		}
	} else {
		e.Log(fmt.Sprintf("[Combat] 攻击命中，但 %s 方战绩区已满，本次不增加星石", attacker.Camp))
	}

	hitEventCtx := &model.EventContext{
		Type:      model.EventAttack,
		SourceID:  pd.SourceID,
		TargetID:  pd.TargetID,
		DamageVal: &pd.Damage,
		Card:      pd.Card,
		AttackInfo: &model.AttackEventInfo{
			ActionType: "Attack",
			IsHit:      true,
			CounterInitiator: func() string {
				if pd.IsCounter {
					return pd.SourceID
				}
				return ""
			}(),
		},
	}
	if e.dispatchAttackRulebookEventTimingWithMarkers(model.TimingAttackHit, attacker, victim, hitEventCtx, attackKindFromCounter(pd.IsCounter), nil).Interrupted {
		return true
	}
	if e.HandlePostAttackHitEffects(pd) {
		pd.AttackHitFlowDispatched = true
		return true
	}

	pd.AttackHitFlowDispatched = true
	return false
}

func (e *GameEngine) dispatchPendingDamageTaken(pd *model.PendingDamage) bool {
	if pd == nil || pd.DamageTakenFlowDispatched {
		return false
	}

	damageEventCtx := &model.EventContext{
		Type:      model.EventDamage,
		SourceID:  pd.SourceID,
		TargetID:  pd.TargetID,
		DamageVal: &pd.Damage,
		Card:      pd.Card,
	}
	damageCtx := e.BuildContext(e.State.Players[pd.TargetID], e.State.Players[pd.SourceID], model.TimingOnDamageTaken, damageEventCtx)
	damageCtx.Flags["IsMagicDamage"] = !strings.EqualFold(string(pd.DamageType), string(model.AttackDamage))
	damageCtx.Flags["holy_shield_eligible"] = strings.EqualFold(string(pd.DamageType), string(model.AttackDamage)) ||
		(pd.Card != nil && strings.TrimSpace(pd.Card.Name) == "魔弹")
	damageCtx.Flags["ignore_shield"] = pd.IgnoreShield || pd.HasInterceptTag(model.CombatInterceptIgnoreHolyShield)
	if damageCtx.Selections == nil {
		damageCtx.Selections = map[string]any{}
	}
	damageCtx.Selections["damage_type"] = pd.DamageType
	pd.DamageTakenFlowDispatched = true

	if e.dispatchDamageRulebookTiming(model.TimingDamageTaken, pd) {
		return true
	}
	e.dispatcher.OnTiming(damageCtx.Timing, damageCtx)
	return e.State.PendingInterrupt != nil
}

func (e *GameEngine) resolvePendingDamageHealChoice(pd *model.PendingDamage) bool {
	if pd == nil || pd.HealResolved {
		return false
	}

	target := e.State.Players[pd.TargetID]
	if target != nil && pd.Damage > 0 && e.canUseHealToResist(target, pd.SourceID, pd.DamageType, pd.IgnoreHeal, pd.AllowCrimsonFaithHeal) {
		maxHeal := target.Heal
		if pd.Damage < maxHeal {
			maxHeal = pd.Damage
		}
		e.dispatchSettlementRulebookTiming(model.TimingHealCap, target, e.State.Players[pd.SourceID], &model.EventContext{
			Type:      model.EventHeal,
			SourceID:  pd.SourceID,
			TargetID:  pd.TargetID,
			DamageVal: &maxHeal,
			Card:      pd.Card,
		})
		maxHeal = e.applyHealCapRules(pd, target, maxHeal)
		if maxHeal > 0 {
			pd.HealResolved = true // 设置标记防止重复推入中断
			e.PushInterrupt(&model.Interrupt{
				Type:     model.InterruptChoice,
				PlayerID: pd.TargetID,
				Context: map[string]interface{}{
					"choice_type": "heal",
					"max_heal":    maxHeal,
					"target_id":   pd.TargetID,
				},
			})
			return true
		}
	}

	pd.HealResolved = true
	return false
}
