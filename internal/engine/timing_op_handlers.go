// gameflow: 各 Timing op 的具体 handler 实现。

package engine

import (
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// 以下 handler 是三大分发器在各 op 下的默认处理单元。
// 规则层（policy/hook）仍由对应 timing 阶段动态装配，不在这里写角色分支。
func timingOpAttackDeclaredCardTransform(e *GameEngine, ctx timingOnAttackDeclaredContext, result *timingOnAttackDeclaredResult) {
	result.Card = e.applyTimingOnAttackDeclaredCardTransforms(ctx.Player, ctx.Card)
}

func timingOpAttackDeclaredTargetContext(e *GameEngine, ctx timingOnAttackDeclaredContext, _ *timingOnAttackDeclaredResult) {
	e.recordTimingOnAttackDeclaredTargetContext(ctx.Player, ctx.TargetID)
}

func timingOpAttackDeclaredStateReset(e *GameEngine, ctx timingOnAttackDeclaredContext, _ *timingOnAttackDeclaredResult) {
	e.resetTimingOnAttackDeclaredState(ctx.Player)
}

func timingOpAttackDeclaredPreCombat(e *GameEngine, ctx timingOnAttackDeclaredContext, _ *timingOnAttackDeclaredResult) {
	e.applyTimingOnAttackDeclaredPreCombatRules(ctx.Player, ctx.Target, ctx.CurrentAction, ctx.EventCtx)
}

func timingOpAttackDeclaredPendingDamageInit(e *GameEngine, ctx timingOnAttackDeclaredContext, _ *timingOnAttackDeclaredResult) {
	e.applyTimingOnAttackDeclaredPendingDamageInitRules(ctx.PendingDamage, ctx.Attacker, ctx.Victim)
	// PD init 已迁移到 TimingOnAttackDeclared TimingHookSpec。
	if ctx.PendingDamage != nil && ctx.Attacker != nil {
		e.dispatchAllRoleTimingHooks(engineplayer.TimingOnAttackDeclared, engineplayer.TimingHookContext{
			SourceID:      ctx.Attacker.ID,
			TargetID:      ctx.TargetID,
			PendingDamage: ctx.PendingDamage,
			ActionType:    model.ActionAttack,
		})
	}
}

func timingOpAttackDeclaredInterrupt(e *GameEngine, ctx timingOnAttackDeclaredContext, result *timingOnAttackDeclaredResult) {
	result.Stop = e.runTimingOnAttackDeclaredInterruptPolicies(ctx.Player, ctx.Target, ctx.CurrentAction, ctx.UserCtx)
}

// 月神 overlay 示例：保留原流程，仅通过配置提升该 handler 的优先级。
func timingOpAttackDeclaredPreCombatMoonOverlay(e *GameEngine, ctx timingOnAttackDeclaredContext, result *timingOnAttackDeclaredResult) {
	timingOpAttackDeclaredPreCombat(e, ctx, result)
}

func timingOpHitCheckPendingDamageAttackHit(e *GameEngine, ctx timingOnHitCheckContext, _ *timingOnHitCheckResult) {
	e.applyTimingOnHitCheckPendingDamageAttackHitRules(ctx.PendingDamage, ctx.Attacker, ctx.Victim)
}

func timingOpHitCheckCombatInteraction(e *GameEngine, ctx timingOnHitCheckContext, result *timingOnHitCheckResult) {
	result.Stop = e.runTimingOnHitCheckCombatInteractionPolicies(ctx.CombatReq)
}

func timingOpHitCheckCombatDefendValidation(e *GameEngine, ctx timingOnHitCheckContext, result *timingOnHitCheckResult) {
	result.Err = e.applyTimingOnHitCheckCombatDefendValidation(ctx.Player, ctx.CombatReq)
}

func timingOpHitCheckCombatCounterCard(e *GameEngine, ctx timingOnHitCheckContext, result *timingOnHitCheckResult) {
	result.Handled, result.Card, result.Err = e.applyTimingOnHitCheckCombatCounterCardPolicy(ctx.Player, ctx.CombatReq, ctx.Card)
}

func timingOpHitCheckCombatCounterElement(e *GameEngine, ctx timingOnHitCheckContext, result *timingOnHitCheckResult) {
	result.Allowed, result.UseFaction = e.applyTimingOnHitCheckCombatCounterElementPolicy(ctx.Player, ctx.CombatReq, ctx.Card)
}

func timingOpHitCheckCombatCounterResolve(e *GameEngine, ctx timingOnHitCheckContext, _ *timingOnHitCheckResult) {
	e.applyTimingOnHitCheckCombatCounterResolvePolicy(ctx.Player, ctx.CombatReq, ctx.CardPtr, ctx.UseFaction)
}

func timingOpHitCheckMagicMissileDefend(e *GameEngine, ctx timingOnHitCheckContext, result *timingOnHitCheckResult) {
	result.Err = e.applyTimingOnHitCheckMagicMissileDefendValidation(ctx.Player, ctx.MagicChain)
}

func timingOpHitCheckMagicMissileCounter(e *GameEngine, ctx timingOnHitCheckContext, result *timingOnHitCheckResult) {
	result.Err = e.applyTimingOnHitCheckMagicMissileCounterValidation(ctx.Player, ctx.MagicChain, ctx.Card)
}

func timingOpHitCheckResponseSkip(e *GameEngine, ctx timingOnHitCheckContext, result *timingOnHitCheckResult) {
	e.runTimingOnResponseSkipEffects(ctx.ResponseState)
	result.Stop = e.State.PendingInterrupt != nil
}

// 阴阳师 overlay 示例：命中判定阶段沿用原流程，展示角色覆盖入口。
func timingOpHitCheckCombatCounterElementOnmyojiOverlay(e *GameEngine, ctx timingOnHitCheckContext, result *timingOnHitCheckResult) {
	timingOpHitCheckCombatCounterElement(e, ctx, result)
}

func timingOpHitCheckSkillAugment(sd *SkillDispatcher, skillIDs []string, ctx *model.Context) []string {
	return sd.applyTimingOnHitCheckResponseSkillAugment(skillIDs, ctx)
}

func timingOpHitCheckSkillNormalize(sd *SkillDispatcher, skillIDs []string, ctx *model.Context) []string {
	return sd.applyTimingOnHitCheckResponseSkillNormalize(skillIDs, ctx)
}

func timingOpDamageCalculatedAttackPassive(e *GameEngine, ctx timingOnDamageCalculatedContext, result *timingOnDamageCalculatedResult) {
	result.Damage = e.applyTimingOnDamageCalculatedAttackPassiveModifiers(ctx.Attacker, ctx.Target, ctx.Action, ctx.Damage)
}

func timingOpDamageCalculatedBeforeTaken(e *GameEngine, ctx timingOnDamageCalculatedContext, result *timingOnDamageCalculatedResult) {
	result.Stop = e.applyTimingOnDamageCalculatedBeforeTakenRules(ctx.PendingDamage)
}

func timingOpDamageCalculatedHealCap(e *GameEngine, ctx timingOnDamageCalculatedContext, result *timingOnDamageCalculatedResult) {
	result.MaxHeal = e.applyTimingOnDamageCalculatedHealCapRules(ctx.PendingDamage, ctx.Target, ctx.MaxHeal)
}

func timingOpDamageCalculatedHealResist(e *GameEngine, ctx timingOnDamageCalculatedContext, result *timingOnDamageCalculatedResult) {
	e.applyTimingOnDamageCalculatedHealResistRules(ctx.PendingDamage, ctx.Target)
	// 治疗抵抗规则已迁移到 TimingOnHealResist TimingHookSpec。
	if ctx.PendingDamage != nil && ctx.Target != nil {
		e.dispatchAllRoleTimingHooks(engineplayer.TimingOnHealResist, engineplayer.TimingHookContext{
			TargetID:      ctx.Target.ID,
			PendingDamage: ctx.PendingDamage,
		})
		result.IgnoreHeal = ctx.PendingDamage.IgnoreHeal
	}
}

// 猩红/瘟疫 overlay 示例：伤害计算阶段保持同一主流程，展示规则组替换能力。
func timingOpDamageCalculatedHealResistOverlay(e *GameEngine, ctx timingOnDamageCalculatedContext, result *timingOnDamageCalculatedResult) {
	timingOpDamageCalculatedHealResist(e, ctx, result)
}
