// gameflow: 构建各 Timing 的 op→handler 注册表。

package engine

import "fmt"

// rebuildTimingDispatchOpRegistryWithPresence 在角色阵容变化后重建三段战斗时机的 op 分发表。
// 游戏流程顺序：先确定 op->handler，再进入各阶段规则链执行。
func (e *GameEngine) rebuildTimingDispatchOpRegistryWithPresence(present map[string]bool) {
	// 游戏流程入口重建：根据当前在场角色，把三大分发器与技能分发器的 op-handler 表一次性装配好。
	e.attackDeclaredOpHandlers = buildAttackDeclaredOpHandlers(present)
	e.hitCheckOpHandlers = buildHitCheckOpHandlers(present)
	e.damageCalculatedOpHandlers = buildDamageCalculatedOpHandlers(present)
	e.hitCheckSkillOpHandlers = buildHitCheckSkillOpHandlers(present)
}

func buildAttackDeclaredOpHandlers(present map[string]bool) map[timingOnAttackDeclaredOp]timingOnAttackDeclaredHandler {
	selected := map[timingOnAttackDeclaredOp]attackDeclaredOpBinding{}
	for _, binding := range timingOpBindings.attackDeclared {
		if !binding.presence.enabled(present) {
			continue
		}
		prev, exists := selected[binding.op]
		if !exists || binding.priority >= prev.priority {
			selected[binding.op] = binding
		}
	}
	handlers := map[timingOnAttackDeclaredOp]timingOnAttackDeclaredHandler{}
	for op, binding := range selected {
		handlers[op] = binding.handler
	}
	// 关键防线：攻击宣言阶段每个 op 都必须有唯一生效 handler，缺失即 panic，避免静默降级。
	for _, op := range []timingOnAttackDeclaredOp{
		timingOnAttackDeclaredCardTransform,
		timingOnAttackDeclaredTargetContext,
		timingOnAttackDeclaredStateReset,
		timingOnAttackDeclaredPreCombat,
		timingOnAttackDeclaredPendingDamageInit,
		timingOnAttackDeclaredInterrupt,
	} {
		if _, ok := handlers[op]; !ok {
			panic(fmt.Sprintf("missing TimingOnAttackDeclared op handler: %s", op))
		}
	}
	return handlers
}

func buildHitCheckOpHandlers(present map[string]bool) map[timingOnHitCheckOp]timingOnHitCheckHandler {
	selected := map[timingOnHitCheckOp]hitCheckOpBinding{}
	for _, binding := range timingOpBindings.hitCheck {
		if !binding.presence.enabled(present) {
			continue
		}
		prev, exists := selected[binding.op]
		if !exists || binding.priority >= prev.priority {
			selected[binding.op] = binding
		}
	}
	handlers := map[timingOnHitCheckOp]timingOnHitCheckHandler{}
	for op, binding := range selected {
		handlers[op] = binding.handler
	}
	// 关键防线：命中判定阶段各 op 必须完整可用，否则流程会出现“能进阶段但无处理器”的不一致。
	for _, op := range []timingOnHitCheckOp{
		timingOnHitCheckPendingDamageAttackHit,
		timingOnHitCheckCombatInteraction,
		timingOnHitCheckCombatDefendValidation,
		timingOnHitCheckCombatCounterCard,
		timingOnHitCheckCombatCounterElement,
		timingOnHitCheckCombatCounterResolve,
		timingOnHitCheckMagicMissileDefend,
		timingOnHitCheckMagicMissileCounter,
		timingOnHitCheckResponseSkip,
	} {
		if _, ok := handlers[op]; !ok {
			panic(fmt.Sprintf("missing TimingOnHitCheck op handler: %s", op))
		}
	}
	return handlers
}

func buildDamageCalculatedOpHandlers(present map[string]bool) map[timingOnDamageCalculatedOp]timingOnDamageCalculatedHandler {
	selected := map[timingOnDamageCalculatedOp]damageCalculatedOpBinding{}
	for _, binding := range timingOpBindings.damageCalc {
		if !binding.presence.enabled(present) {
			continue
		}
		prev, exists := selected[binding.op]
		if !exists || binding.priority >= prev.priority {
			selected[binding.op] = binding
		}
	}
	handlers := map[timingOnDamageCalculatedOp]timingOnDamageCalculatedHandler{}
	for op, binding := range selected {
		handlers[op] = binding.handler
	}
	// 关键防线：伤害计算阶段必须完整覆盖四个 op，确保伤害时间轴顺序固定。
	for _, op := range []timingOnDamageCalculatedOp{
		timingOnDamageCalculatedAttackPassive,
		timingOnDamageCalculatedBeforeTaken,
		timingOnDamageCalculatedHealCap,
		timingOnDamageCalculatedHealResist,
	} {
		if _, ok := handlers[op]; !ok {
			panic(fmt.Sprintf("missing TimingOnDamageCalculated op handler: %s", op))
		}
	}
	return handlers
}

func buildHitCheckSkillOpHandlers(present map[string]bool) map[timingOnHitCheckSkillOp]timingOnHitCheckSkillHandler {
	selected := map[timingOnHitCheckSkillOp]hitCheckSkillOpBinding{}
	for _, binding := range timingOpBindings.hitCheckSkill {
		if !binding.presence.enabled(present) {
			continue
		}
		prev, exists := selected[binding.op]
		if !exists || binding.priority >= prev.priority {
			selected[binding.op] = binding
		}
	}
	handlers := map[timingOnHitCheckSkillOp]timingOnHitCheckSkillHandler{}
	for op, binding := range selected {
		handlers[op] = binding.handler
	}
	// 关键防线：响应技能列表流程必须同时包含 augment 与 normalize 两步。
	for _, op := range []timingOnHitCheckSkillOp{timingOnHitCheckSkillAugment, timingOnHitCheckSkillNormalize} {
		if _, ok := handlers[op]; !ok {
			panic(fmt.Sprintf("missing TimingOnHitCheck skill op handler: %s", op))
		}
	}
	return handlers
}

// ---------- Timing 链式调用辅助（原 timing_chain_helpers.go） ----------

func runTimingErrorChain(steps ...func() error) error {
	for _, step := range steps {
		if step == nil {
			continue
		}
		if err := step(); err != nil {
			return err
		}
	}
	return nil
}

// ---------- 角色在场过滤（原 timing_presence_registry_helpers.go） ----------

// presenceHookEntry 描述"角色在场 -> 挂载规则函数"的一条声明式配置。
type presenceHookEntry[T any] struct {
	requireAny []string
	hook       T
}

func requireAny(roleIDs ...string) []string {
	return roleIDs
}

func buildPresenceHooks[T any](present map[string]bool, entries []presenceHookEntry[T]) []T {
	var hooks []T
	for _, entry := range entries {
		if !matchPresenceAny(present, entry.requireAny) {
			continue
		}
		hooks = append(hooks, entry.hook)
	}
	return hooks
}

func matchPresenceAny(present map[string]bool, roleIDs []string) bool {
	if len(roleIDs) == 0 {
		return true
	}
	for _, roleID := range roleIDs {
		if present[roleID] {
			return true
		}
	}
	return false
}
