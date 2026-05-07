// gameflow: 伤害管线公共步骤（若存在）。

package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

// AddPendingDamage 将延迟伤害添加到队列
func (e *GameEngine) AddPendingDamage(pd model.PendingDamage) {
	e.State.PendingDamageQueue = append(e.State.PendingDamageQueue, pd)
	e.Log(fmt.Sprintf("[System] 延迟伤害已添加: Source: %s, Target: %s, Damage: %d, Type: %s",
		pd.SourceID, pd.TargetID, pd.Damage, pd.DamageType))

	if !e.isDamageResolutionActive() {
		if e.State.ReturnTurnStage == "" && e.State.ReturnCombatStage == model.CombatStageNone && e.State.ReturnSubflow == model.SubflowNone {
			if point := e.CurrentChoiceResumePoint(); hasChoiceResumePoint(point) {
				e.setReturnPoint(point)
			}
		}
		e.enterDamageResolution(nil)
	}
}

// AddPendingDamageFront 将延迟伤害插入队列头部（用于“必须先结算”的伤害）。
func (e *GameEngine) AddPendingDamageFront(pd model.PendingDamage) {
	e.State.PendingDamageQueue = append([]model.PendingDamage{pd}, e.State.PendingDamageQueue...)
	e.Log(fmt.Sprintf("[System] 延迟伤害已前插: Source: %s, Target: %s, Damage: %d, Type: %s",
		pd.SourceID, pd.TargetID, pd.Damage, pd.DamageType))

	if !e.isDamageResolutionActive() {
		if e.State.ReturnTurnStage == "" && e.State.ReturnCombatStage == model.CombatStageNone && e.State.ReturnSubflow == model.SubflowNone {
			if point := e.CurrentChoiceResumePoint(); hasChoiceResumePoint(point) {
				e.setReturnPoint(point)
			}
		}
		e.enterDamageResolution(nil)
	}
}

// ProcessPendingDamages 处理伤害队列中的所有伤害
// 返回 true 如果产生了中断需要暂停 Drive
func (e *GameEngine) ProcessPendingDamages() bool {
	for len(e.State.PendingDamageQueue) > 0 {
		// 伤害流水线固定顺序：
		// 1) 攻击命中链 -> 2) 承伤前规则 -> 3) 承伤触发 -> 4) 扣血前规则 -> 5) 扣血 -> 6) 结算后规则。
		pd := &e.State.PendingDamageQueue[0]

		if e.processPendingAttackHit(pd) {
			return true
		}
		if e.removePendingDamageIfAttackMissed(pd) {
			continue
		}
		if e.processPendingDamageBeforeApply(pd) {
			return true
		}

		resolved := e.applyAndPopPendingDamage(pd)
		if e.applyTimingOnDamageTakenAfterResolvedRules(&resolved) {
			return true
		}
		// 结算后若产生新中断（如爆牌/后续技能选择），暂停 Drive。
		if e.State.PendingInterrupt != nil {
			return true
		}
	}

	// 伤害队列清空后，触发 after_damage 流程边界恢复点
	e.processFlowContinuations(model.FlowContinuationAfterDamage)

	return false
}

func (e *GameEngine) removePendingDamageIfAttackMissed(pd *model.PendingDamage) bool {
	if pd == nil || !pd.HasCheck(model.PendingDamageCheckAttackMissResolved) {
		return false
	}
	e.State.PendingDamageQueue = e.State.PendingDamageQueue[1:]
	return true
}

func (e *GameEngine) processPendingDamageBeforeApply(pd *model.PendingDamage) bool {
	// 承伤前统一阶段：角色规则通过 hooks 注入，主流程不写角色特判。
	if e.applyTimingOnDamageCalculatedBeforeTakenRules(pd) {
		return true
	}
	if e.dispatchPendingDamageTaken(pd) {
		return true
	}
	if e.applyTimingOnDamageTakenAfterTakenRules(pd) {
		return true
	}
	if e.resolvePendingDamageHealChoice(pd) {
		return true
	}
	if e.applyTimingOnDamageAppliedBeforeApplyRules(pd) {
		return true
	}
	return false
}

func (e *GameEngine) applyAndPopPendingDamage(pd *model.PendingDamage) model.PendingDamage {
	if pd.Damage < 0 {
		pd.Damage = 0
	}
	target := e.State.Players[pd.TargetID]
	source := e.State.Players[pd.SourceID]
	if target != nil && pd.Damage > 0 {
		if pd.DamageType == model.AttackDamage && source != nil {
			e.NotifyActionStep(fmt.Sprintf("总共对%s造成%d点伤害", model.GetPlayerDisplayName(target), pd.Damage))
		}
		e.NotifyDamageDealt(pd.SourceID, pd.TargetID, pd.Damage, pd.DamageType)
	}
	if target != nil {
		// 扣血/摸牌等基础结算
		e.applyDamageWithOptions(target, pd.Damage, pd.DamageType, pd.CapDrawToHandLimit, pd.SourceID, pd.SourceSkillID, pd.OverflowMoraleLossFixed)
		// 扣血后角色/状态清理由 hooks 注入。
		e.applyTimingOnDamageTakenAfterApplyRules(pd, target)
	}

	resolved := *pd
	e.State.PendingDamageQueue = e.State.PendingDamageQueue[1:]
	return resolved
}
