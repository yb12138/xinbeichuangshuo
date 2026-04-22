// gameflow: TimingOnAttackDeclared 动态钩子表重建。

package engine

// rebuildTimingOnAttackDeclaredRegistry 根据当前已上场角色，重建攻击宣言阶段的执行表。
// 同时刷新 TimingOnHitCheck / TimingOnDamageCalculated 的动态执行表，保证三段战斗时序一致。
func (e *GameEngine) rebuildTimingOnAttackDeclaredRegistry() {
	if e == nil {
		return
	}

	present := e.currentCharacterPresence()
	// 先重建“阶段分发器”的 op-handler 表，再重建各阶段内部的规则链。
	// 这样可保证同一轮角色变化后，分发入口与规则执行链使用同一份角色视图。
	e.rebuildTimingDispatchOpRegistryWithPresence(present)
	e.rebuildTimingOnAttackDeclaredRegistryWithPresence(present)
	e.rebuildTimingOnHitCheckRegistryWithPresence(present)
	e.rebuildTimingOnDamageCalculatedRegistryWithPresence(present)
	e.rebuildTurnTimingRegistryWithPresence(present)
	e.rebuildActionSelectionTimingRegistryWithPresence(present)
	e.rebuildSpecialActionTimingRegistryWithPresence(present)
	e.rebuildSkillTimingRegistryWithPresence(present)
	e.rebuildGameStartTimingRegistryWithPresence(present)
	e.rebuildCampChangedTimingRegistryWithPresence(present)
	e.rebuildInterruptTimingRegistryWithPresence(present)
}

// rebuildTimingOnAttackDeclaredRegistryWithPresence 装配 TimingOnAttackDeclared 的动态规则表。
func (e *GameEngine) rebuildTimingOnAttackDeclaredRegistryWithPresence(present map[string]bool) {
	e.attackDeclaredCardTransformHooks = buildPresenceHooks(present, []presenceHookEntry[attackCardRuntimeTransformHook]{
		{requireAny: requireAny("blaze_witch"), hook: applyBlazeWitchAttackCardRuntimeHook},
	})

	// 目标上下文已迁移到 TimingOnAttackTargetCtx TimingHookSpec。
	e.attackDeclaredTargetContextHooks = nil

	// 攻击状态重置已迁移到 TimingOnAttackStateReset TimingHookSpec。
	e.attackDeclaredStateResetHooks = nil

	// 攻击门控已迁移到 TimingOnAttackGating TimingHookSpec。
	e.attackDeclaredPreCombatHooks = nil

	// PD init 已迁移到 TimingOnAttackDeclared TimingHookSpec。
	e.attackDeclaredPendingDamageInitOps = nil
}

// rebuildTimingOnHitCheckRegistryWithPresence 装配 TimingOnHitCheck 的动态规则表。
func (e *GameEngine) rebuildTimingOnHitCheckRegistryWithPresence(present map[string]bool) {
	e.hitCheckCombatInteractionHooks = buildPresenceHooks(present, []presenceHookEntry[combatInteractionPolicyHook]{
		{requireAny: requireAny("onmyoji"), hook: combatInteractionOnmyojiBindingInterruptHook},
		{requireAny: requireAny("onmyoji"), hook: combatInteractionOnmyojiBindingCounterHook},
		{requireAny: requireAny("onmyoji"), hook: combatInteractionOnmyojiYinYangInterruptHook},
		// 暗系不可应战是命中判定公共规则，不依赖角色在场。
		{hook: combatInteractionDarkElementResponsePolicyHook},
	})

	e.hitCheckCombatDefendValidationPolicies = nil
	// 防御验证已迁移到 TimingOnDefendValidation TimingHookSpec。

	e.hitCheckCombatCounterCardPolicies = buildPresenceHooks(present, []presenceHookEntry[combatCounterCardPolicy]{
		{requireAny: requireAny("magic_swordsman"), hook: combatCounterShadowRejectMagicBulletPolicy},
	})

	e.hitCheckCombatCounterElementPolicies = buildPresenceHooks(present, []presenceHookEntry[combatCounterElementPolicy]{
		{requireAny: requireAny("onmyoji"), hook: combatCounterOnmyojiFactionElementPolicy},
	})

	e.hitCheckCombatCounterResolvePolicies = buildPresenceHooks(present, []presenceHookEntry[combatCounterResolvePolicy]{
		{requireAny: requireAny("onmyoji"), hook: combatCounterOnmyojiFactionResolvePolicy},
	})

	// 魔弹链防御验证已迁移到 TimingOnMagicMissileDefend TimingHookSpec。
	e.hitCheckMagicMissileDefendPolicies = nil
	// 魔弹链反击验证已迁移到 TimingOnMagicMissileCounter TimingHookSpec。
	e.hitCheckMagicMissileCounterPolicies = nil

	e.hitCheckResponseSkipHooks = nil
	// 响应跳过后效已迁移到 TimingOnResponseSkillSkip TimingHookSpec。

	e.hitCheckResponseSkillIDAugmenters = buildPresenceHooks(present, []presenceHookEntry[responseSkillIDAugmenter]{
		{requireAny: requireAny("beast_samurai"), hook: augmentBeastSamuraiResponseSkillIDs},
	})

	e.hitCheckResponseSkillIDNormalizers = buildPresenceHooks(present, []presenceHookEntry[responseSkillIDNormalizer]{
		{requireAny: requireAny("fighter"), hook: normalizeFighterResponseSkillIDs},
	})
}

// rebuildTimingOnDamageCalculatedRegistryWithPresence 装配 TimingOnDamageCalculated 的动态规则表。
// 所有伤害阶段规则已迁移到 TimingHookSpec，不再需要动态装配。
func (e *GameEngine) rebuildTimingOnDamageCalculatedRegistryWithPresence(present map[string]bool) {
	e.damageCalculatedAttackPassiveHooks = nil
	e.damageTakenAfterResolvedHooks = buildPresenceHooks(present, []presenceHookEntry[pendingDamageAfterResolvedHook]{
		{hook: pendingDamageRolePostResolvedHook},
	})
}

// rebuildTurnTimingRegistryWithPresence 装配 Turn 主流程中的 TimingOnTurnStart/OnBeforeAction/OnTurnEnd 阶段钩子。
func (e *GameEngine) rebuildTurnTimingRegistryWithPresence(present map[string]bool) {
	// 凋零过期已迁移到 TimingOnTurnBeforeStart TimingHookSpec。
	// arbiter turnStart hooks 已迁移到 TimingOnTurnStart TimingHookSpec。
	// holy_bow turnStart hooks 已迁移到 TimingOnTurnStart TimingHookSpec。
	// blaze_witch beforeAction 已迁移到 TimingBeforeAction TimingHookSpec。
	// assassin beforeAction 已迁移到 TimingBeforeAction TimingHookSpec。
	// crimson_knight turnEnd 已迁移到 TimingOnTurnEnd TimingHookSpec。

	e.turnStartMainHooks = buildPresenceHooks(present, []presenceHookEntry[turnTimingHook]{
		{requireAny: requireAny("bard"), hook: startupBardRousingHook},
		{requireAny: requireAny("blood_priestess"), hook: turnStartBloodPriestessBleedHook},
		{requireAny: requireAny("valkyrie"), hook: turnStartValkyrieMilitaryGloryHook},
	})

	// BeforeAction 场上基础效果属于通用流程，始终装配。
	e.beforeActionFieldHooks = []turnTimingHook{
		beforeActionPoisonHook,
		beforeActionFiveElementsBindHook,
		beforeActionWeakHook,
	}

	// blaze_witch/assassin hooks 已迁移到 TimingHookSpec，移除重复注册。
	e.beforeActionExecuteHooks = buildPresenceHooks(present, []presenceHookEntry[turnTimingHook]{
		{requireAny: requireAny("magic_swordsman"), hook: startupMagicSwordsmanShadowReleaseHook},
		{requireAny: requireAny("hero"), hook: startupHeroExhaustionReleaseHook},
		{requireAny: requireAny("arbiter"), hook: startupArbiterForcedDoomsdayHook},
		{requireAny: requireAny("hero"), hook: startupHeroTauntHook},
	})

	e.turnEndPreExtraHooks = buildPresenceHooks(present, []presenceHookEntry[turnTimingHook]{
		{requireAny: requireAny("moon_goddess"), hook: turnEndMoonGoddessHook},
		{requireAny: requireAny("bard"), hook: turnEndBardHook},
		{requireAny: requireAny("onmyoji"), hook: turnEndOnmyojiHook},
	})

	e.turnEndFinalHooks = nil // Migrated to TimingHookSpec
}

// rebuildActionSelectionTimingRegistryWithPresence 装配 TimingBeforeActionExecute 行动枢纽约束策略。
func (e *GameEngine) rebuildActionSelectionTimingRegistryWithPresence(present map[string]bool) {
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
}

// rebuildSpecialActionTimingRegistryWithPresence 装配特殊行动前后时机策略。
func (e *GameEngine) rebuildSpecialActionTimingRegistryWithPresence(present map[string]bool) {
	e.specialActionOverridePolicies = buildPresenceHooks(present, []presenceHookEntry[specialActionOverridePolicy]{
		{requireAny: requireAny("adventurer"), hook: specialActionAdventurerUndergroundLawOverride},
	})

	e.specialActionPostHooks = buildPresenceHooks(present, []presenceHookEntry[specialActionPostHook]{
		{requireAny: requireAny("holy_bow"), hook: specialActionHolyBowHolyGloryExitHook},
	})
}

// rebuildSkillTimingRegistryWithPresence 装配技能主动发动成功后的后置策略。
func (e *GameEngine) rebuildSkillTimingRegistryWithPresence(present map[string]bool) {
	e.skillPostHooks = buildPresenceHooks(present, []presenceHookEntry[skillPostHook]{
		{requireAny: requireAny("arbiter"), hook: skillPostArbiterForcedDoomsdayCleanupHook},
	})
}

// rebuildGameStartTimingRegistryWithPresence 装配 TimingOnGameStart 的三段规则。
func (e *GameEngine) rebuildGameStartTimingRegistryWithPresence(present map[string]bool) {
	e.gameStartAddPlayerHooks = buildPresenceHooks(present, []presenceHookEntry[gameStartPlayerHook]{
		{hook: bootstrapApplyRoleDefaults},
	})
	e.gameStartInitialDealHooks = buildPresenceHooks(present, []presenceHookEntry[gameStartPlayerHook]{
		{hook: bootstrapEnsureStarterRoleCards},
	})
	e.gameStartFinalizeHooks = nil
}

// rebuildCampChangedTimingRegistryWithPresence 装配 TimingOnCampChanged 派生状态同步规则。
// 已迁移到 TimingOnPlayerSetup / TimingOnCampCupChanged TimingHookSpec，不再需要动态注册。
func (e *GameEngine) rebuildCampChangedTimingRegistryWithPresence(present map[string]bool) {
}

// rebuildInterruptTimingRegistryWithPresence 装配战斗时序中的中断策略入口。
func (e *GameEngine) rebuildInterruptTimingRegistryWithPresence(present map[string]bool) {
	e.attackDeclaredInterrupts = buildPresenceHooks(present, []presenceHookEntry[attackDeclaredInterruptHook]{
		{requireAny: requireAny("moon_goddess"), hook: attackStartMoonGoddessMedusaInterruptHook},
	})

	// HolySword interrupt 已迁移到 blade_master TimingHookSpec
	e.actionEndInterrupts = nil
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
