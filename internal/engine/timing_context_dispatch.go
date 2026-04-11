// gameflow: Timing 与 Context Op 的总分派入口。

package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

type timingOnAttackDeclaredOp string

const (
	timingOnAttackDeclaredCardTransform     timingOnAttackDeclaredOp = "card_transform"
	timingOnAttackDeclaredTargetContext     timingOnAttackDeclaredOp = "target_context"
	timingOnAttackDeclaredStateReset        timingOnAttackDeclaredOp = "state_reset"
	timingOnAttackDeclaredPreCombat         timingOnAttackDeclaredOp = "pre_combat"
	timingOnAttackDeclaredPendingDamageInit timingOnAttackDeclaredOp = "pending_damage_init"
	timingOnAttackDeclaredInterrupt         timingOnAttackDeclaredOp = "interrupt"
)

type timingOnAttackDeclaredContext struct {
	Op            timingOnAttackDeclaredOp
	Player        *model.Player
	Target        *model.Player
	TargetID      string
	Card          model.Card
	CurrentAction *model.QueuedAction
	EventCtx      *model.EventContext
	PendingDamage *model.PendingDamage
	Attacker      *model.Player
	Victim        *model.Player
	UserCtx       *model.Context
}

type timingOnAttackDeclaredResult struct {
	Card model.Card
	Stop bool
}

type timingOnAttackDeclaredHandler func(e *GameEngine, ctx timingOnAttackDeclaredContext, result *timingOnAttackDeclaredResult)

// dispatchTimingOnAttackDeclared 统一处理 TimingOnAttackDeclared 的上下文分发。
func (e *GameEngine) dispatchTimingOnAttackDeclared(ctx timingOnAttackDeclaredContext) timingOnAttackDeclaredResult {
	// 游戏流程：攻击宣言时机会按 op 细分为“变牌 -> 目标上下文 -> 状态重置 -> 前置战斗规则 -> 伤害初始化 -> 中断判定”。
	// 这里不写角色分支，只做 op 分发；具体角色效果由重建阶段装配到 handler 内部的策略链决定。
	result := timingOnAttackDeclaredResult{Card: ctx.Card}
	handler, ok := e.attackDeclaredOpHandlers[ctx.Op]
	if !ok {
		panic(fmt.Sprintf("unregistered TimingOnAttackDeclared op: %s", ctx.Op))
	}
	handler(e, ctx, &result)
	return result
}

type timingOnHitCheckOp string

const (
	timingOnHitCheckPendingDamageAttackHit timingOnHitCheckOp = "pending_damage_attack_hit"
	timingOnHitCheckCombatInteraction      timingOnHitCheckOp = "combat_interaction"
	timingOnHitCheckCombatDefendValidation timingOnHitCheckOp = "combat_defend_validation"
	timingOnHitCheckCombatCounterCard      timingOnHitCheckOp = "combat_counter_card"
	timingOnHitCheckCombatCounterElement   timingOnHitCheckOp = "combat_counter_element"
	timingOnHitCheckCombatCounterResolve   timingOnHitCheckOp = "combat_counter_resolve"
	timingOnHitCheckMagicMissileDefend     timingOnHitCheckOp = "magic_missile_defend"
	timingOnHitCheckMagicMissileCounter    timingOnHitCheckOp = "magic_missile_counter"
	timingOnHitCheckResponseSkip           timingOnHitCheckOp = "response_skip"
)

type timingOnHitCheckContext struct {
	Op            timingOnHitCheckOp
	PendingDamage *model.PendingDamage
	Attacker      *model.Player
	Victim        *model.Player
	CombatReq     *model.CombatRequest
	Player        *model.Player
	Card          model.Card
	CardPtr       *model.Card
	UseFaction    bool
	MagicChain    *model.MagicBulletChain
	ResponseState *responseResumeState
}

type timingOnHitCheckResult struct {
	Stop       bool
	Err        error
	Handled    bool
	Card       model.Card
	Allowed    bool
	UseFaction bool
}

type timingOnHitCheckHandler func(e *GameEngine, ctx timingOnHitCheckContext, result *timingOnHitCheckResult)

// dispatchTimingOnHitCheck 统一处理 TimingOnHitCheck 的上下文分发。
func (e *GameEngine) dispatchTimingOnHitCheck(ctx timingOnHitCheckContext) timingOnHitCheckResult {
	// 游戏流程：命中判定阶段统一入口。无论是战斗响应、应战校验、还是魔弹响应，都先收敛成 op 再分发。
	result := timingOnHitCheckResult{Card: ctx.Card}
	handler, ok := e.hitCheckOpHandlers[ctx.Op]
	if !ok {
		panic(fmt.Sprintf("unregistered TimingOnHitCheck op: %s", ctx.Op))
	}
	handler(e, ctx, &result)
	return result
}

type timingOnHitCheckSkillOp string

const (
	timingOnHitCheckSkillAugment   timingOnHitCheckSkillOp = "augment"
	timingOnHitCheckSkillNormalize timingOnHitCheckSkillOp = "normalize"
)

type timingOnHitCheckSkillHandler func(sd *SkillDispatcher, skillIDs []string, ctx *model.Context) []string

// dispatchTimingOnHitCheckSkillIDs 统一处理响应技能列表在 TimingOnHitCheck 的变换。
func (sd *SkillDispatcher) dispatchTimingOnHitCheckSkillIDs(skillIDs []string, ctx *model.Context, op timingOnHitCheckSkillOp) []string {
	// 游戏流程：命中判定里“可响应技能列表”分两步处理：先 augment（补充），再 normalize（规范化顺序/互斥）。
	handler, ok := sd.engine.hitCheckSkillOpHandlers[op]
	if !ok {
		panic(fmt.Sprintf("unregistered TimingOnHitCheck skill op: %s", op))
	}
	return handler(sd, skillIDs, ctx)
}

type timingOnDamageCalculatedOp string

const (
	timingOnDamageCalculatedAttackPassive timingOnDamageCalculatedOp = "attack_passive"
	timingOnDamageCalculatedBeforeTaken   timingOnDamageCalculatedOp = "before_taken"
	timingOnDamageCalculatedHealCap       timingOnDamageCalculatedOp = "heal_cap"
	timingOnDamageCalculatedHealResist    timingOnDamageCalculatedOp = "heal_resist"
)

type timingOnDamageCalculatedContext struct {
	Op       timingOnDamageCalculatedOp
	Attacker *model.Player
	Target   *model.Player
	Action   model.Action
	Damage   int

	PendingDamage *model.PendingDamage
	MaxHeal       int
}

type timingOnDamageCalculatedResult struct {
	Damage     int
	Stop       bool
	MaxHeal    int
	IgnoreHeal bool
}

type timingOnDamageCalculatedHandler func(e *GameEngine, ctx timingOnDamageCalculatedContext, result *timingOnDamageCalculatedResult)

// dispatchTimingOnDamageCalculated 统一处理 TimingOnDamageCalculated 的上下文分发。
func (e *GameEngine) dispatchTimingOnDamageCalculated(ctx timingOnDamageCalculatedContext) timingOnDamageCalculatedResult {
	// 游戏流程：伤害计算统一入口，按 op 分发到“被动增减伤 -> 承伤前规则 -> 治疗上限 -> 治疗门禁”。
	result := timingOnDamageCalculatedResult{
		Damage:  ctx.Damage,
		MaxHeal: ctx.MaxHeal,
	}
	handler, ok := e.damageCalculatedOpHandlers[ctx.Op]
	if !ok {
		panic(fmt.Sprintf("unregistered TimingOnDamageCalculated op: %s", ctx.Op))
	}
	handler(e, ctx, &result)
	return result
}
