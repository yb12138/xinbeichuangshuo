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

	e.attackDeclaredTargetContextHooks = buildPresenceHooks(present, []presenceHookEntry[attackTargetContextHook]{
		{requireAny: requireAny("magic_bow"), hook: recordMagicBowAttackTargetOrder},
	})

	e.attackDeclaredStateResetHooks = buildPresenceHooks(present, []presenceHookEntry[attackStartStateResetHook]{
		{requireAny: requireAny("holy_lancer"), hook: resetHolyLancerAttackFlags},
		{requireAny: requireAny("sword_emperor"), hook: resetSwordEmperorAttackFlags},
		{requireAny: requireAny("beast_samurai"), hook: resetBeastSamuraiAttackFlags},
		{requireAny: requireAny("magic_swordsman"), hook: resetMagicSwordsmanAttackFlags},
		{requireAny: requireAny("fighter"), hook: resetFighterAttackFlags},
	})

	// 攻击宣言 -> 命中判定前：固定先跑战斗公共规则，再按角色挂载劫持规则。
	e.attackDeclaredPreCombatHooks = buildPresenceHooks(present, []presenceHookEntry[attackPreCombatHook]{
		{hook: applyCombatPolicyAttackGating},
		{requireAny: requireAny("hero"), hook: applyHeroAttackGating},
		{requireAny: requireAny("fighter"), hook: applyFighterAttackGating},
		{requireAny: requireAny("moon_goddess"), hook: applyMoonGoddessAttackGating},
		{requireAny: requireAny("assassin"), hook: applyAssassinAttackGating},
		{requireAny: requireAny("holy_lancer"), hook: applyHolyLancerAttackGating},
		{requireAny: requireAny("magic_swordsman"), hook: applyMagicSwordsmanAttackGating},
		{hook: applyDarkElementNoCounterRule},
		{requireAny: requireAny("beast_samurai"), hook: applyBeastSamuraiAttackGating},
	})

	e.attackDeclaredPendingDamageInitOps = buildPresenceHooks(present, []presenceHookEntry[pendingDamageAttackInitHook]{
		{requireAny: requireAny("hero"), hook: pendingDamageHeroRoarMissArmHook},
		{requireAny: requireAny("fighter"), hook: pendingDamageFighterChargeMissArmHook},
	})
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

	e.hitCheckCombatDefendValidationPolicies = buildPresenceHooks(present, []presenceHookEntry[combatDefendValidationPolicy]{
		{requireAny: requireAny("magic_lancer"), hook: combatDefendMagicLancerDarkBindPolicy},
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

	e.hitCheckMagicMissileDefendPolicies = buildPresenceHooks(present, []presenceHookEntry[magicMissileDefendValidationPolicy]{
		{requireAny: requireAny("magic_lancer"), hook: magicMissileDefendMagicLancerDarkBindPolicy},
	})
	e.hitCheckMagicMissileCounterPolicies = buildPresenceHooks(present, []presenceHookEntry[magicMissileCounterValidationPolicy]{
		{requireAny: requireAny("magic_lancer"), hook: magicMissileCounterMagicLancerDarkBindPolicy},
	})

	e.hitCheckResponseSkipHooks = buildPresenceHooks(present, []presenceHookEntry[responseSkipHook]{
		{requireAny: requireAny("holy_lancer"), hook: holyLancerEarthSkippedResponseHook},
	})

	e.hitCheckResponseSkillIDAugmenters = buildPresenceHooks(present, []presenceHookEntry[responseSkillIDAugmenter]{
		{requireAny: requireAny("beast_samurai"), hook: augmentBeastSamuraiResponseSkillIDs},
	})

	e.hitCheckResponseSkillIDNormalizers = buildPresenceHooks(present, []presenceHookEntry[responseSkillIDNormalizer]{
		{requireAny: requireAny("fighter"), hook: normalizeFighterResponseSkillIDs},
	})
}

// rebuildTimingOnDamageCalculatedRegistryWithPresence 装配 TimingOnDamageCalculated 的动态规则表。
func (e *GameEngine) rebuildTimingOnDamageCalculatedRegistryWithPresence(present map[string]bool) {
	e.damageCalculatedAttackPassiveHooks = buildPresenceHooks(present, []presenceHookEntry[attackPassiveDamageHook]{
		{requireAny: requireAny("elf_archer"), hook: attackPassiveElfFireShotHook},
		{requireAny: requireAny("magic_swordsman"), hook: attackPassiveMagicSwordsmanShadowHook},
		{requireAny: requireAny("magic_lancer"), hook: attackPassiveMagicLancerBonusHook},
		{requireAny: requireAny("fighter"), hook: attackPassiveFighterBonusHook},
		{requireAny: requireAny("hero"), hook: attackPassiveHeroRoarBonusHook},
		{requireAny: requireAny("assassin"), hook: attackPassiveAssassinStealthBonusHook},
		{requireAny: requireAny("holy_bow"), hook: attackPassiveHolyBowPenaltyHook},
		{requireAny: requireAny("sword_emperor"), hook: attackPassiveSwordEmperorBonusHook},
		{requireAny: requireAny("beast_samurai"), hook: attackPassiveBeastSamuraiBonusHook},
	})

	e.damageCalculatedBeforeTakenHooks = buildPresenceHooks(present, []presenceHookEntry[pendingDamageBeforeTakenHook]{
		{requireAny: requireAny("crimson_sword_spirit", "crimson_knight", "plague_mage"), hook: pendingDamageHealResistGateHook},
		{requireAny: requireAny("soul_sorcerer"), hook: pendingDamageSoulLinkTransferHook},
	})

	e.damageCalculatedHealCapHooks = buildPresenceHooks(present, []presenceHookEntry[pendingDamageHealCapHook]{
		{requireAny: requireAny("priest"), hook: pendingDamagePriestHealCapHook},
	})

	e.damageCalculatedHealResistRules = buildPresenceHooks(present, []presenceHookEntry[pendingDamageHealResistRule]{
		{requireAny: requireAny("crimson_sword_spirit"), hook: pendingDamageRoseCourtyardHealResistRule},
		{requireAny: requireAny("crimson_knight"), hook: pendingDamageCrimsonKnightHealResistRule},
		{requireAny: requireAny("plague_mage"), hook: pendingDamagePlagueMageHealResistRule},
	})

	e.damageTakenAfterTakenHooks = buildPresenceHooks(present, []presenceHookEntry[pendingDamageAfterTakenHook]{
		{requireAny: requireAny("sword_emperor"), hook: pendingDamageSwordEmperorAfterTakenHook},
	})

	e.damageAppliedBeforeApplyHooks = buildPresenceHooks(present, []presenceHookEntry[pendingDamageBeforeApplyHook]{
		{requireAny: requireAny("butterfly_dancer"), hook: pendingDamageButterflyBeforeApplyHook},
	})

	e.damageTakenAfterApplyHooks = buildPresenceHooks(present, []presenceHookEntry[pendingDamageAfterApplyHook]{
		{hook: pendingDamageElementalSealCleanupHook},
		{requireAny: requireAny("crimson_sword_spirit"), hook: pendingDamageResetCrimsonSwordSpiritLocksHook},
		{requireAny: requireAny("blaze_witch"), hook: pendingDamageResetBlazeWitchLocksHook},
	})

	e.damageTakenAfterResolvedHooks = buildPresenceHooks(present, []presenceHookEntry[pendingDamageAfterResolvedHook]{
		{hook: pendingDamageRolePostResolvedHook},
	})
}

// rebuildTurnTimingRegistryWithPresence 装配 Turn 主流程中的 TimingOnTurnStart/OnBeforeAction/OnTurnEnd 阶段钩子。
func (e *GameEngine) rebuildTurnTimingRegistryWithPresence(present map[string]bool) {
	e.turnStartBeforeStartHooks = buildPresenceHooks(present, []presenceHookEntry[turnTimingHook]{
		{requireAny: requireAny("butterfly_dancer"), hook: turnBeforeStartButterflyDancerWitherExpiryHook},
	})

	e.turnStartMainHooks = buildPresenceHooks(present, []presenceHookEntry[turnTimingHook]{
		{requireAny: requireAny("arbiter"), hook: startupArbiterTurnResetHook},
		{requireAny: requireAny("holy_bow"), hook: startupHolyBowTurnResetHook},
		{requireAny: requireAny("bard"), hook: startupBardRousingHook},
		{requireAny: requireAny("blood_priestess"), hook: turnStartBloodPriestessBleedHook},
		{requireAny: requireAny("arbiter"), hook: turnStartArbiterJudgmentUpkeepHook},
		{requireAny: requireAny("valkyrie"), hook: turnStartValkyrieMilitaryGloryHook},
	})

	// BeforeAction 场上基础效果属于通用流程，始终装配。
	e.beforeActionFieldHooks = []turnTimingHook{
		beforeActionPoisonHook,
		beforeActionFiveElementsBindHook,
		beforeActionWeakHook,
	}

	e.beforeActionExecuteHooks = buildPresenceHooks(present, []presenceHookEntry[turnTimingHook]{
		{requireAny: requireAny("blaze_witch"), hook: startupBlazeWitchFlameReleaseHook},
		{requireAny: requireAny("assassin"), hook: startupAssassinStealthReleaseHook},
		{requireAny: requireAny("magic_swordsman"), hook: startupMagicSwordsmanShadowReleaseHook},
		{requireAny: requireAny("hero"), hook: startupHeroExhaustionReleaseHook},
		{requireAny: requireAny("arbiter"), hook: startupArbiterForcedDoomsdayHook},
		{requireAny: requireAny("hero"), hook: startupHeroTauntHook},
	})

	e.turnEndPreExtraHooks = buildPresenceHooks(present, []presenceHookEntry[turnTimingHook]{
		{requireAny: requireAny("beast_samurai"), hook: turnEndBeastSamuraiHook},
		{requireAny: requireAny("fighter"), hook: turnEndFighterHook},
		{requireAny: requireAny("elf_archer"), hook: turnEndElfArcherHook},
		{requireAny: requireAny("plague_mage"), hook: turnEndPlagueMageHook},
		{requireAny: requireAny("moon_goddess"), hook: turnEndMoonGoddessHook},
		{requireAny: requireAny("bard"), hook: turnEndBardHook},
		{requireAny: requireAny("crimson_sword_spirit"), hook: turnEndCrimsonSwordSpiritHook},
		{requireAny: requireAny("crimson_knight"), hook: turnEndCrimsonKnightHook},
		{requireAny: requireAny("war_homunculus"), hook: turnEndWarHomunculusHook},
		{requireAny: requireAny("onmyoji"), hook: turnEndOnmyojiHook},
	})

	e.turnEndFinalHooks = buildPresenceHooks(present, []presenceHookEntry[turnTimingHook]{
		{requireAny: requireAny("holy_bow"), hook: turnEndHolyBowHook},
		{requireAny: requireAny("holy_lancer"), hook: turnEndHolyLancerHook},
	})
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
	e.gameStartFinalizeHooks = buildPresenceHooks(present, []presenceHookEntry[gameStartFinalizeHook]{
		{requireAny: requireAny("blood_priestess"), hook: actionFinalizeBloodPriestessBleedHook},
	})
}

// rebuildCampChangedTimingRegistryWithPresence 装配 TimingOnCampChanged 派生状态同步规则。
func (e *GameEngine) rebuildCampChangedTimingRegistryWithPresence(present map[string]bool) {
	e.campChangedPlayerSetupHooks = buildPresenceHooks(present, []presenceHookEntry[campChangedPlayerHook]{
		{requireAny: requireAny("holy_lancer"), hook: syncHolyLancerDerivedStateOnPlayerSetup},
	})
	e.campChangedCampCupHooks = buildPresenceHooks(present, []presenceHookEntry[campChangedCupHook]{
		{requireAny: requireAny("holy_lancer"), hook: syncHolyLancerDerivedStateOnCampCupChanged},
	})
}

// rebuildInterruptTimingRegistryWithPresence 装配战斗时序中的中断策略入口。
func (e *GameEngine) rebuildInterruptTimingRegistryWithPresence(present map[string]bool) {
	e.attackDeclaredInterrupts = buildPresenceHooks(present, []presenceHookEntry[attackDeclaredInterruptHook]{
		{requireAny: requireAny("moon_goddess"), hook: attackStartMoonGoddessMedusaInterruptHook},
	})

	e.actionEndInterrupts = buildPresenceHooks(present, []presenceHookEntry[actionEndInterruptHook]{
		{requireAny: requireAny("blade_master"), hook: actionEndHolySwordInterruptHook},
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
