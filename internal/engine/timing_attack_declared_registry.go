// gameflow: 动态钩子表重建。

package engine

// rebuildTimingOnAttackDeclaredRegistry 根据当前已上场角色，重建必要的执行表。
func (e *GameEngine) rebuildTimingOnAttackDeclaredRegistry() {
	if e == nil {
		return
	}

	present := e.currentCharacterPresence()

	// 通用流程钩子
	e.beforeActionFieldHooks = []turnTimingHook{
		beforeActionPoisonHook,
		beforeActionFiveElementsBindHook,
		beforeActionWeakHook,
	}

	// 游戏启动钩子
	e.gameStartAddPlayerHooks = []gameStartPlayerHook{bootstrapApplyRoleDefaults}
	e.gameStartInitialDealHooks = []gameStartPlayerHook{bootstrapEnsureStarterRoleCards}

	// 从角色注册表装配策略钩子
	e.assembleRolePolicies(present)
}

// assembleRolePolicies 从 roleRegistry 收集角色注册的策略钩子并装配。
// 当前使用 RoleEntry.StrategyHooks 字段，角色包在 RoleEntry 中注册策略。
func (e *GameEngine) assembleRolePolicies(present map[string]bool) {
	// 战斗交互策略
	e.hitCheckCombatInteractionHooks = buildPresenceHooks(present, []presenceHookEntry[combatInteractionPolicyHook]{
		{requireAny: requireAny("onmyoji"), hook: combatInteractionOnmyojiBindingInterruptHook},
		{requireAny: requireAny("onmyoji"), hook: combatInteractionOnmyojiBindingCounterHook},
		{requireAny: requireAny("onmyoji"), hook: combatInteractionOnmyojiYinYangInterruptHook},
		{hook: combatInteractionDarkElementResponsePolicyHook},
	})

	e.hitCheckCombatCounterCardPolicies = buildPresenceHooks(present, []presenceHookEntry[combatCounterCardPolicy]{
		{requireAny: requireAny("magic_swordsman"), hook: combatCounterShadowRejectMagicBulletPolicy},
	})

	e.hitCheckCombatCounterElementPolicies = buildPresenceHooks(present, []presenceHookEntry[combatCounterElementPolicy]{
		{requireAny: requireAny("onmyoji"), hook: combatCounterOnmyojiFactionElementPolicy},
	})

	e.hitCheckCombatCounterResolvePolicies = buildPresenceHooks(present, []presenceHookEntry[combatCounterResolvePolicy]{
		{requireAny: requireAny("onmyoji"), hook: combatCounterOnmyojiFactionResolvePolicy},
	})

	e.hitCheckResponseSkillIDAugmenters = buildPresenceHooks(present, []presenceHookEntry[responseSkillIDAugmenter]{
		{requireAny: requireAny("beast_samurai"), hook: augmentBeastSamuraiResponseSkillIDs},
	})

	e.hitCheckResponseSkillIDNormalizers = buildPresenceHooks(present, []presenceHookEntry[responseSkillIDNormalizer]{
		{requireAny: requireAny("fighter"), hook: normalizeFighterResponseSkillIDs},
	})

	e.damageTakenAfterResolvedHooks = buildPresenceHooks(present, []presenceHookEntry[pendingDamageAfterResolvedHook]{
		{hook: pendingDamageRolePostResolvedHook},
	})

	// 行动选择策略
	e.beforeActionOptionPolicies = buildPresenceHooks(present, []presenceHookEntry[actionSelectionOptionPolicy]{
		{requireAny: requireAny("arbiter"), hook: actionSelectionArbiterForcedDoomsdayOptionsPolicy},
		{requireAny: requireAny("hero"), hook: actionSelectionHeroTauntOptionsPolicy},
		{requireAny: requireAny("fighter"), hook: actionSelectionFighterHundredDragonOptionsPolicy},
	})

	e.beforeActionValidationPolicies = buildPresenceHooks(present, []presenceHookEntry[actionSelectionValidationPolicy]{
		{requireAny: requireAny("arbiter"), hook: actionSelectionArbiterForcedDoomsdayValidationPolicy},
		{requireAny: requireAny("hero"), hook: actionSelectionHeroTauntValidationPolicy},
		{requireAny: requireAny("fighter"), hook: actionSelectionFighterHundredDragonValidationPolicy},
	})

	e.skillPostHooks = buildPresenceHooks(present, []presenceHookEntry[skillPostHook]{
		{requireAny: requireAny("arbiter"), hook: skillPostArbiterForcedDoomsdayCleanupHook},
	})

	// 特殊行动策略
	e.specialActionOverridePolicies = buildPresenceHooks(present, []presenceHookEntry[specialActionOverridePolicy]{
		{requireAny: requireAny("adventurer"), hook: specialActionAdventurerUndergroundLawOverride},
	})

	e.specialActionPostHooks = buildPresenceHooks(present, []presenceHookEntry[specialActionPostHook]{
		{requireAny: requireAny("holy_bow"), hook: specialActionHolyBowHolyGloryExitHook},
	})

	// 中断策略
	e.attackDeclaredInterrupts = buildPresenceHooks(present, []presenceHookEntry[attackDeclaredInterruptHook]{
		{requireAny: requireAny("moon"), hook: attackStartMoonGoddessMedusaInterruptHook},
	})

	// 攻击牌变换
	e.attackDeclaredCardTransformHooks = buildPresenceHooks(present, []presenceHookEntry[attackCardRuntimeTransformHook]{
		{requireAny: requireAny("blaze_witch"), hook: applyBlazeWitchAttackCardRuntimeHook},
	})
}

func (e *GameEngine) currentCharacterPresence() map[string]bool {
	presence := map[string]bool{}
	if e == nil || e.State == nil {
		return presence
	}
	for _, player := range e.State.Players {
		if player == nil || player.Character == nil || player.Character.ID == "" {
			continue
		}
		presence[player.Character.ID] = true
	}
	return presence
}

// ---------- Presence Hook 辅助函数 ----------

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
