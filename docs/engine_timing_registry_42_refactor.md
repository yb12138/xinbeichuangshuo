# Engine Hook/Runtime 42 Registry Timing Refactor Plan

## Goal
- Remove registry-array driven hook dispatch from `internal/engine`.
- Keep generic flow focused on game timeline stages.
- Move role effects to timing-aware rule entry points (no meaningless fallback compatibility).
- Reuse current framework (`dispatcher`, `policy`, existing handlers) without introducing new model/class types.

## Ordered Migration Matrix (42 Registries)
1. `actionFinalizeIdleHooks` -> `runTimingOnActionEndFinalizeEffects` TimingOnGameStart
2. `attackCardRuntimeTransformHooks` -> `applyTimingOnAttackDeclaredCardTransforms`  TimingOnAttackDeclared
3. `attackPassiveDamageHooks` -> `applyTimingOnDamageCalculatedAttackPassiveModifiers`  TimingOnDamageCalculated
4. `attackTargetContextHooks` -> `recordTimingOnAttackDeclaredTargetContext` TimingOnAttackDeclared
5. `attackStartStateResetHooks` -> `resetTimingOnAttackDeclaredState` TimingOnAttackDeclared
6. `attackPreCombatHooks` -> `applyTimingOnAttackDeclaredPreCombatRules` TimingOnAttackDeclared
7. `pendingDamageAttackHitHooks` -> `applyTimingOnHitCheckPendingDamageAttackHitRules` TimingOnHitCheck
8. `pendingDamageAttackInitHooks` -> `applyTimingOnAttackDeclaredPendingDamageInitRules` TimingOnAttackDeclared
9. `pendingDamageBeforeTakenHooks` -> `applyTimingOnDamageCalculatedBeforeTakenRules` TimingOnDamageCalculated
10. `pendingDamageHealResistRules` -> `applyTimingOnDamageCalculatedHealResistRules` TimingOnDamageCalculated
11. `pendingDamageHealCapHooks` -> `applyTimingOnDamageCalculatedHealCapRules` TimingOnDamageCalculated
12. `pendingDamageAfterTakenHooks` -> `applyTimingOnDamageTakenAfterTakenRules`  TimingOnDamageTaken
13. `pendingDamageBeforeApplyHooks` -> `applyTimingOnDamageAppliedBeforeApplyRules` TimingOnDamageApplied
14. `pendingDamageAfterApplyHooks` -> `applyTimingOnDamageTakenAfterApplyRules` TimingOnDamageTaken
15. `pendingDamageAfterResolvedHooks` -> `applyTimingOnDamageTakenAfterResolvedRules` TimingOnDamageTaken
16. `playerDerivedStateHooks` -> `refreshTimingDerivedStateOnPlayerSetup` TimingOnCampChanged
17. `campCupChangedHooks` -> `refreshTimingDerivedStateOnCampCupChanged` TimingOnCampChanged
18. `turnBeforeStartPhaseHooks` -> `runTimingOnTurnStartBeforeStartHooks`  TimingOnTurnStart
19. `turnStartPhaseHooks` -> `runTimingOnTurnStartHooks` TimingOnTurnStart
20. `actionStartPhaseHooks` -> `runTimingBeforeActionExecuteHooks` TimingOnBeforeAction
21. `turnEndPreExtraActionHooks` -> `runTimingOnTurnEndPreExtraHooks`  TimingOnTurnEnd
22. `turnEndFinalHooks` -> `runTimingOnTurnEndFinalHooks` TimingOnTurnEnd
23. `turnProgressionFallbackHooks` -> `runTimingOnTurnEndFallbackHooks` (remove fallback behavior, keep deterministic turn-end flow) TimingOnTurnEnd
24. `addPlayerBootstrapHooks` -> `runPlayerAddBootstrapTiming`  TimingOnGameStart
25. `gameStartBootstrapHooks` -> `runPlayerGameStartBootstrapTiming` TimingOnGameStart
26. `responseSkipHooks` -> `runTimingOnResponseSkipEffects` TimingOnHitCheck
27. `beforeActionPhaseHooks` -> `runTimingOnBeforeActionHooks` TimingOnBeforeAction
28. `actionSelectionOptionPolicies` -> `applyTimingBeforeActionExecuteOptionPolicies` TimingBeforeActionExecute
29. `actionSelectionValidationConstraintPolicies` -> `applyTimingBeforeActionExecuteValidationPolicies` TimingBeforeActionExecute
30. `combatInteractionHooks` -> `runTimingOnHitCheckCombatInteractionPolicies` TimingOnHitCheck
31. `attackStartInterruptHooks` -> `runTimingOnAttackDeclaredInterruptPolicies`  TimingOnAttackDeclared
32. `actionEndInterruptHooks` -> `runTimingOnActionEndInterruptPolicies`  TimingOnActionEnd
33. `combatDefendValidationPolicies` -> `applyTimingOnHitCheckCombatDefendValidation` TimingOnHitCheck
34. `combatCounterCardPolicies` -> `applyTimingOnHitCheckCombatCounterCardPolicy` TimingOnHitCheck
35. `combatCounterElementPolicies` -> `applyTimingOnHitCheckCombatCounterElementPolicy` TimingOnHitCheck
36. `combatCounterResolvePolicies` -> `applyTimingOnHitCheckCombatCounterResolvePolicy` TimingOnHitCheck
37. `magicMissileDefendValidationPolicies` -> `applyTimingOnHitCheckMagicMissileDefendValidation` TimingOnHitCheck
38. `magicMissileCounterValidationPolicies` -> `applyTimingOnHitCheckMagicMissileCounterValidation` TimingOnHitCheck
39. `responseSkillIDAugmenters` -> `applyTimingOnHitCheckResponseSkillAugment` TimingOnHitCheck
40. `responseSkillIDNormalizers` -> `applyTimingOnHitCheckResponseSkillNormalize` TimingOnHitCheck
41. `specialActionOverrides` -> `applyTimingBeforeActionExecuteSpecialActionOverride` TimingBeforeActionExecute
42. `specialActionPostHooks` -> `runTimingOnActionEndSpecialActionPost` TimingOnActionEnd

## Timing Implementation Detail (Current Code)
1. `runTimingOnActionEndFinalizeEffects`: `actionFinalizeBloodPriestessBleedHook`
2. `applyTimingOnAttackDeclaredCardTransforms`: `applyBlazeWitchAttackCardRuntimeHook`
3. `applyTimingOnDamageCalculatedAttackPassiveModifiers`: `ElfFireShot -> MagicSwordsmanShadow -> MagicLancer -> Fighter -> HeroRoar -> AssassinStealth -> HolyBowPenalty -> SwordEmperor -> BeastSamurai`
4. `recordTimingOnAttackDeclaredTargetContext`: `recordMagicBowAttackTargetOrder`
5. `resetTimingOnAttackDeclaredState`: `resetHolyLancerAttackFlags -> resetSwordEmperorAttackFlags -> resetBeastSamuraiAttackFlags -> resetMagicSwordsmanAttackFlags -> resetFighterAttackFlags`
6. `applyTimingOnAttackDeclaredPreCombatRules`: `applyCombatPolicyAttackGating -> applyHeroAttackGating -> applyFighterAttackGating -> applyMoonGoddessAttackGating -> applyAssassinAttackGating -> applyHolyLancerAttackGating -> applyMagicSwordsmanAttackGating -> applyDarkElementNoCounterRule -> applyBeastSamuraiAttackGating`
7. `applyTimingOnHitCheckPendingDamageAttackHitRules`: `pendingDamageBerserkerBloodRoarHook`
8. `applyTimingOnAttackDeclaredPendingDamageInitRules`: `pendingDamageHeroRoarMissArmHook -> pendingDamageFighterChargeMissArmHook`
9. `applyTimingOnDamageCalculatedBeforeTakenRules`: `pendingDamageHealResistGateHook -> pendingDamageSoulLinkTransferHook`
10. `applyTimingOnDamageCalculatedHealResistRules`: `pendingDamageRoseCourtyardHealResistRule -> pendingDamageCrimsonKnightHealResistRule -> pendingDamagePlagueMageHealResistRule`
11. `applyTimingOnDamageCalculatedHealCapRules`: `pendingDamagePriestHealCapHook`
12. `applyTimingOnDamageTakenAfterTakenRules`: `pendingDamageSwordEmperorAfterTakenHook`
13. `applyTimingOnDamageAppliedBeforeApplyRules`: `pendingDamageButterflyBeforeApplyHook`
14. `applyTimingOnDamageTakenAfterApplyRules`: `pendingDamageElementalSealCleanupHook -> pendingDamageResetCrimsonSwordSpiritLocksHook -> pendingDamageResetBlazeWitchLocksHook`
15. `applyTimingOnDamageTakenAfterResolvedRules`: `pendingDamageRolePostResolvedHook`
16. `refreshTimingDerivedStateOnPlayerSetup`: `syncHolyLancerDerivedStateOnPlayerSetup`
17. `refreshTimingDerivedStateOnCampCupChanged`: `syncHolyLancerDerivedStateOnCampCupChanged`
18. `runTimingOnTurnStartBeforeStartHooks`: `turnBeforeStartButterflyDancerWitherExpiryHook`
19. `runTimingOnTurnStartHooks`: `startupArbiterTurnResetHook -> startupHolyBowTurnResetHook -> startupBardRousingHook -> turnStartBloodPriestessBleedHook -> turnStartArbiterJudgmentUpkeepHook -> turnStartValkyrieMilitaryGloryHook`
20. `runTimingBeforeActionExecuteHooks`: `startupBlazeWitchFlameReleaseHook -> startupAssassinStealthReleaseHook -> startupMagicSwordsmanShadowReleaseHook -> startupHeroExhaustionReleaseHook -> startupArbiterForcedDoomsdayHook -> startupHeroTauntHook`
21. `runTimingOnTurnEndPreExtraHooks`: `turnEndBeastSamuraiHook -> turnEndFighterHook -> turnEndElfArcherHook -> turnEndPlagueMageHook -> turnEndMoonGoddessHook -> turnEndBardHook -> turnEndCrimsonSwordSpiritHook -> turnEndCrimsonKnightHook -> turnEndWarHomunculusHook -> turnEndOnmyojiHook`
22. `runTimingOnTurnEndFinalHooks`: `turnEndHolyBowHook -> turnEndHolyLancerHook`
23. `runTimingOnTurnEndFallbackHooks`: no-op (fallback path removed; only TurnEnd timing is authoritative)
24. `runPlayerAddBootstrapTiming`: `bootstrapApplyRoleDefaults`
25. `runPlayerGameStartBootstrapTiming`: `bootstrapEnsureStarterRoleCards`
26. `runTimingOnResponseSkipEffects`: `holyLancerEarthSkippedResponseHook`
27. `runTimingOnBeforeActionHooks`: `beforeActionPoisonHook -> beforeActionFiveElementsBindHook -> beforeActionWeakHook`
28. `applyTimingBeforeActionExecuteOptionPolicies`: `actionSelectionArbiterForcedDoomsdayOptionsPolicy -> actionSelectionHeroTauntOptionsPolicy -> actionSelectionFighterHundredDragonOptionsPolicy`
29. `applyTimingBeforeActionExecuteValidationPolicies`: `actionSelectionArbiterForcedDoomsdayValidationPolicy -> actionSelectionHeroTauntValidationPolicy -> actionSelectionFighterHundredDragonValidationPolicy`
30. `runTimingOnHitCheckCombatInteractionPolicies`: `combatInteractionOnmyojiBindingInterruptHook -> combatInteractionOnmyojiBindingCounterHook -> combatInteractionOnmyojiYinYangInterruptHook -> combatInteractionDarkElementResponsePolicyHook`
31. `runTimingOnAttackDeclaredInterruptPolicies`: `attackStartMoonGoddessMedusaInterruptHook`
32. `runTimingOnActionEndInterruptPolicies`: `actionEndHolySwordInterruptHook`
33. `applyTimingOnHitCheckCombatDefendValidation`: `combatDefendMagicLancerDarkBindPolicy`
34. `applyTimingOnHitCheckCombatCounterCardPolicy`: `combatCounterShadowRejectMagicBulletPolicy`
35. `applyTimingOnHitCheckCombatCounterElementPolicy`: `combatCounterOnmyojiFactionElementPolicy`
36. `applyTimingOnHitCheckCombatCounterResolvePolicy`: `combatCounterOnmyojiFactionResolvePolicy`
37. `applyTimingOnHitCheckMagicMissileDefendValidation`: `magicMissileDefendMagicLancerDarkBindPolicy`
38. `applyTimingOnHitCheckMagicMissileCounterValidation`: `magicMissileCounterMagicLancerDarkBindPolicy`
39. `applyTimingOnHitCheckResponseSkillAugment`: `augmentBeastSamuraiResponseSkillIDs`
40. `applyTimingOnHitCheckResponseSkillNormalize`: `normalizeFighterResponseSkillIDs`
41. `applyTimingBeforeActionExecuteSpecialActionOverride`: `specialActionAdventurerUndergroundLawOverride`
42. `runTimingOnActionEndSpecialActionPost`: `specialActionHolyBowHolyGloryExitHook`

## Timing Stage Merged Dispatch (Code Simplification)
- `TimingOnGameStart`：统一入口 `runTimingOnGameStartHooks`，按 `AddPlayer / InitialDeal / FinalizeIdle` 分发。
- `TimingOnCampChanged`：统一入口 `runTimingOnCampChangedHooks`，按 `PlayerSetup / CampCupChanged` 分发。
- `TimingOnTurnStart`：统一入口 `runTimingOnTurnStartStageHooks`，按 `BeforeStart / Main` 分发。
- `TimingOnBeforeAction`：统一入口 `runTimingOnBeforeActionStageHooks`，按 `ResolveField / ResolveActionStart` 分发（把 #20 与 #27 放到同一机制）。
- `TimingOnTurnEnd`：统一入口 `runTimingOnTurnEndStageHooks`，按 `PreExtra / Final / Fallback` 分发。
- `TimingOnAttackDeclared`：统一入口 `dispatchTimingOnAttackDeclared`，按上下文 `Op` 分发（变牌、目标上下文、状态重置、前置规则、攻击伤害初始化、中断策略）。
- `TimingOnHitCheck`：统一入口 `dispatchTimingOnHitCheck`，按上下文 `Op` 分发（命中链、战斗交互、应战校验、魔弹校验、响应跳过等）。
- `TimingOnDamageCalculated`：统一入口 `dispatchTimingOnDamageCalculated`，按上下文 `Op` 分发（攻击被动修正、承伤前规则、治疗门禁、治疗上限）。
- 同类链式执行统一复用：
  - `runTimingBoolChain`：用于布尔短路规则链。
  - `runTimingErrorChain`：用于校验型错误短路规则链。

## Execution Sequence
1. Replace registry-array iteration with deterministic explicit call chains.
2. Keep behavior equivalent first, then move role checks into timing-aware handler/policy conditions.
3. Remove fallback-only paths where they hide flow ordering problems.
4. Run engine regression tests after each migration batch.

## Current Step
- Step A: Build deterministic call-chain functions for all 42 registries.
- Step B: Switch call sites to new function names in turn/action/combat/damage pipelines.
- Step C: Remove old registry arrays.

## Op-Handler Registry (Latest)
- 三大分发器（`TimingOnAttackDeclared / TimingOnHitCheck / TimingOnDamageCalculated`）已从静态 `switch/固定map` 迁移为“外部配置源 + 运行时重建表”。
- 配置文件：`internal/engine/config/timing_op_bindings.json`。
- 重建入口：`rebuildTimingDispatchOpRegistryWithPresence`，按当前在场角色选择最终 handler。
- 选择规则：
  - `require_any_characters` 命中才参与；
  - 同 `stage+op` 选 `priority` 最高项；
  - 若某必需 op 无处理器，直接 panic（不做兜底兼容）。

### Overlay 示例（已落地）
1. `attack_declared.pre_combat`：月神在场时走 overlay handler（行为等价，作为覆盖示例）。
2. `hit_check.combat_counter_element`：阴阳师在场时走 overlay handler（行为等价，作为覆盖示例）。
3. `damage_calculated.heal_resist`：猩红/瘟疫角色组在场时走 overlay handler（行为等价，作为覆盖示例）。

> 说明：当前 overlay 先保持行为等价，目的是先把“可配置覆盖链路”闭环；后续新增角色机制可直接在配置层替换对应 op-handler。
