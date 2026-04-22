// gameflow: 动态钩子表重建。

package engine

// rebuildTimingOnAttackDeclaredRegistry 根据当前已上场角色，重建必要的执行表。
func (e *GameEngine) rebuildTimingOnAttackDeclaredRegistry() {
	if e == nil {
		return
	}

	present := e.currentCharacterPresence()

	// 先重建 op-handler 分发表
	e.rebuildTimingDispatchOpRegistryWithPresence(present)

	// 通用流程钩子
	e.beforeActionFieldHooks = []turnTimingHook{
		beforeActionPoisonHook,
		beforeActionFiveElementsBindHook,
		beforeActionWeakHook,
	}

	// 游戏启动钩子
	e.gameStartAddPlayerHooks = []gameStartPlayerHook{bootstrapApplyRoleDefaults}
	e.gameStartInitialDealHooks = []gameStartPlayerHook{bootstrapEnsureStarterRoleCards}

	// 使用 buildPresenceHooks 装配策略（渐进式重构：逐步迁移到角色包 PolicySpecs）
	e.assembleRolePolicies(present)
}

// assembleRolePolicies 装配角色策略。
// 当前保持 buildPresenceHooks 模式，后续逐步迁移到角色包 PolicySpecs 注册。
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
