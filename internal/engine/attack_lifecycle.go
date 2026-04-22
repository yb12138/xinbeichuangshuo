// gameflow: 攻击生命周期对象：与 CombatRequest 绑定的步骤推进。
// 重构：删除 op 分发，直接调用 engine 方法。

package engine

import (
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// AttackLifecycle 聚合攻击相关的生命周期化机制，统一主流程调用入口。
// 目标是让 turn_fsm_dispatcher/pending_damage_runtime 只关心阶段推进，
// 而把攻击链的可扩展逻辑集中到生命周期接口中。
type AttackLifecycle interface {
	TransformAttackCard(player *model.Player, card model.Card) model.Card
	RecordAttackTargetContext(player *model.Player, targetID string)
	ResetAttackStartState(player *model.Player)
	ApplyPreCombatRules(player *model.Player, target *model.Player, currentAction *model.QueuedAction, eventCtx *model.EventContext)
	RunPendingDamageAttackInit(pd *model.PendingDamage, attacker *model.Player, victim *model.Player)
	RunPendingDamageAttackHit(pd *model.PendingDamage, attacker *model.Player, victim *model.Player)
}

type defaultAttackLifecycle struct {
	engine *GameEngine
}

func NewAttackLifecycle(engine *GameEngine) AttackLifecycle {
	return &defaultAttackLifecycle{engine: engine}
}

func (l *defaultAttackLifecycle) TransformAttackCard(player *model.Player, card model.Card) model.Card {
	if l == nil || l.engine == nil {
		return card
	}
	return l.engine.applyTimingOnAttackDeclaredCardTransforms(player, card)
}

func (l *defaultAttackLifecycle) RecordAttackTargetContext(player *model.Player, targetID string) {
	if l == nil || l.engine == nil {
		return
	}
	l.engine.recordTimingOnAttackDeclaredTargetContext(player, targetID)
}

func (l *defaultAttackLifecycle) ResetAttackStartState(player *model.Player) {
	if l == nil || l.engine == nil {
		return
	}
	l.engine.resetTimingOnAttackDeclaredState(player)
}

func (l *defaultAttackLifecycle) ApplyPreCombatRules(player *model.Player, target *model.Player, currentAction *model.QueuedAction, eventCtx *model.EventContext) {
	if l == nil || l.engine == nil {
		return
	}
	l.engine.applyTimingOnAttackDeclaredPreCombatRules(player, target, currentAction, eventCtx)
}

func (l *defaultAttackLifecycle) RunPendingDamageAttackInit(pd *model.PendingDamage, attacker *model.Player, victim *model.Player) {
	if l == nil || l.engine == nil {
		return
	}
	l.engine.applyTimingOnAttackDeclaredPendingDamageInitRules(pd, attacker, victim)
	// PD init 后触发角色 TimingHook
	if pd != nil && attacker != nil {
		var targetID string
		if victim != nil {
			targetID = victim.ID
		}
		l.engine.dispatchAllRoleTimingHooks(engineplayer.TimingOnAttackDeclared, engineplayer.TimingHookContext{
			SourceID:      attacker.ID,
			TargetID:      targetID,
			PendingDamage: pd,
			ActionType:    model.ActionAttack,
		})
	}
}

func (l *defaultAttackLifecycle) RunPendingDamageAttackHit(pd *model.PendingDamage, attacker *model.Player, victim *model.Player) {
	if l == nil || l.engine == nil {
		return
	}
	l.engine.applyTimingOnHitCheckPendingDamageAttackHitRules(pd, attacker, victim)
}

// ---------- 备用入口（无 lifecycle 时的直接调用） ----------

func (e *GameEngine) transformAttackCard(player *model.Player, card model.Card) model.Card {
	if e != nil && e.lifecycle != nil {
		return e.lifecycle.TransformAttackCard(player, card)
	}
	return e.applyTimingOnAttackDeclaredCardTransforms(player, card)
}

func (e *GameEngine) recordAttackTargetLifecycle(player *model.Player, targetID string) {
	if e != nil && e.lifecycle != nil {
		e.lifecycle.RecordAttackTargetContext(player, targetID)
		return
	}
	e.recordTimingOnAttackDeclaredTargetContext(player, targetID)
}

func (e *GameEngine) resetAttackStartLifecycle(player *model.Player) {
	if e != nil && e.lifecycle != nil {
		e.lifecycle.ResetAttackStartState(player)
		return
	}
	e.resetTimingOnAttackDeclaredState(player)
}

func (e *GameEngine) applyAttackPreCombatLifecycle(player *model.Player, target *model.Player, currentAction *model.QueuedAction, eventCtx *model.EventContext) {
	if e != nil && e.lifecycle != nil {
		e.lifecycle.ApplyPreCombatRules(player, target, currentAction, eventCtx)
		return
	}
	e.applyTimingOnAttackDeclaredPreCombatRules(player, target, currentAction, eventCtx)
}

func (e *GameEngine) runPendingDamageAttackLifecycle(pd *model.PendingDamage, attacker *model.Player, victim *model.Player) {
	if e != nil && e.lifecycle != nil {
		e.lifecycle.RunPendingDamageAttackInit(pd, attacker, victim)
		e.lifecycle.RunPendingDamageAttackHit(pd, attacker, victim)
		return
	}
	e.applyTimingOnAttackDeclaredPendingDamageInitRules(pd, attacker, victim)
	if pd != nil && attacker != nil {
		var targetID string
		if victim != nil {
			targetID = victim.ID
		}
		e.dispatchAllRoleTimingHooks(engineplayer.TimingOnAttackDeclared, engineplayer.TimingHookContext{
			SourceID:      attacker.ID,
			TargetID:      targetID,
			PendingDamage: pd,
			ActionType:    model.ActionAttack,
		})
	}
	e.applyTimingOnHitCheckPendingDamageAttackHitRules(pd, attacker, victim)
}
