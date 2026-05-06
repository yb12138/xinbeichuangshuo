# 测试用例文档（按角色分类）

> 生成日期：2026-04-24
> 项目：新北创说（Starcup Engine）
> 测试框架：Go `testing` + Vitest
> 测试文件总数：约112个（不含 node_modules）

---

## 目录

- [一、角色技能测试](#一角色技能测试)
  - [1. 冒险者 (Adventurer)](#1-冒险者-adventurer)
  - [2. 天使 (Angel)](#2-天使-angel)
  - [3. 弓手 (Archer)](#3-弓手-archer)
  - [4. 刺客 (Assassin)](#4-刺客-assassin)
  - [5. 吟游诗人 (Bard)](#5-吟游诗人-bard)
  - [6. 兽武者 (Beast Samurai)](#6-兽武者-beast-samurai)
  - [7. 狂战士 (Berserker)](#7-狂战士-berserker)
  - [8. 剑圣 (BladeMaster)](#8-剑圣-blademaster)
  - [9. 烈焰女巫 (Blaze Witch)](#9-烈焰女巫-blaze-witch)
  - [10. 血巫 (Blood Priestess)](#10-血巫-blood-priestess)
  - [11. 蝶舞者 (Butterfly Dancer)](#11-蝶舞者-butterfly-dancer)
  - [12. 红莲骑士 / 人造人 (Crimson Knight / Hom)](#12-红莲骑士--人造人-crimson-knight--hom)
  - [13. 元素师 (Elementalist)](#13-元素师-elementalist)
  - [14. 精灵射手 (Elf Archer)](#14-精灵射手-elf-archer)
  - [15. 格斗家 (Fighter)](#15-格斗家-fighter)
  - [16. 勇者 (Hero)](#16-勇者-hero)
  - [17. 圣弓 (Holy Bow)](#17-圣弓-holy-bow)
  - [18. 圣枪 (Holy Lancer)](#18-圣枪-holy-lancer)
  - [19. 魔弓 (Magic Bow)](#19-魔弓-magic-bow)
  - [20. 魔枪 (Magic Lancer)](#20-魔枪-magic-lancer)
  - [21. 魔剑士 (Magic Swordsman)](#21-魔剑士-magic-swordsman)
  - [22. 魔法少女 (Magical Girl)](#22-魔法少女-magical-girl)
  - [23. 月女神 (Moon Goddess)](#23-月女神-moon-goddess)
  - [24. 阴阳师 (Onmyoji)](#24-阴阳师-onmyoji)
  - [25. 祭司 (Priest)](#25-祭司-priest)
  - [26. 瘟疫法师 (Plague Mage)](#26-瘟疫法师-plague-mage)
  - [27. 祈祷师 (Prayer Master)](#27-祈祷师-prayer-master)
  - [28. 贤者 (Sage)](#28-贤者-sage)
  - [29. 圣女 (Saintess)](#29-圣女-saintess)
  - [30. 封印师 (Sealer)](#30-封印师-sealer)
  - [31. 灵魂术士 (Soul Sorcerer)](#31-灵魂术士-soul-sorcerer)
  - [32. 灵符师 (Spirit Caster)](#32-灵符师-spirit-caster)
  - [33. 剑皇 (Sword Emperor)](#33-剑皇-sword-emperor)
  - [34. 女武神 (Valkyrie)](#34-女武神-valkyrie)
  - [35. 仲裁者 (Arbiter)](#35-仲裁者-arbiter)
- [二、引擎通用机制测试](#二引擎通用机制测试)
- [三、服务器层测试](#三服务器层测试)
- [四、集成测试](#四集成测试)
- [五、前端 Vitest 测试](#五前端-vitest-测试)

---

## 一、角色技能测试

### 1. 冒险家 (Adventurer)

**测试文件**：`internal/engine/adventurer_fraud_regression_test.go`、`internal/engine/adventurer_priest_rules_regression_test.go`

| 测试函数 | 技能/机制     | 场景摘要 |
|---|-----------|---|
| TestAdventurerFraud_PickTwoThenChooseAttackElement | 欺诈（选2张）   | 选2张同系牌，再选攻击元素(雷)，战斗栈元素=雷、伤害2、手牌-2、+1水晶 |
| TestAdventurerFraud_PickThreeAutoConvertsToDark | 欺诈（选3张）   | 选3张同系牌自动转为暗灭攻击（不可应战），元素=暗、CanBeResponded=false |
| TestAdventurerStealSky_ModeAndExtraActionChoice | 偷天换日      | 选宝石转移后蓝→红，进入额外行动 |
| TestAdventurerUndergroundLaw_RewritesBuyInsteadOfDefaultSettlement | 地下法则      | 摸3张、+2宝石（非1宝石1水晶） |
| TestAdventurerExtractFullEnergy_ForceParadiseTransfer | 满能量强制天堂转移 | 不可跳过，强制选天堂转给队友 |
| TestAdventurerParadise_TransferOnlyExtractedEnergy | 天堂仅转提炼能量  | 保留原有能量只转提炼部分 |
| TestAdventurerParadise_TargetsFilteredByExtractCapacity | 天堂目标按容量过滤 | 仅容量足够的队友出现在目标池 |
| TestAdventurerExtract_ParadisePromptRefreshAfterSelection | 提炼后天堂刷新提示 | 选择后立即下发新响应技能提示 |

---

### 2. 天使 (Angel)

**测试文件**：`internal/engine/angel_config_regression_test.go`、`tests/angel_skills_test.go`

| 测试函数 | 技能/机制 | 场景摘要 |
|---|---|---|
| TestAngelSong_RunsAsTurnStartResponseAndResumesActionSelection | 天使之歌 | 回合开始响应触发，消耗1水晶移除虚弱，羁绊补触发治疗选择，恢复正常行动窗口 |
| TestGodProtection_PromptsForXAndPartiallyMitigatesMoraleLoss | 神之庇护（部分减免） | 选X=2消耗2水晶将士气损失从3减至1 |
| TestGodProtection_TriggersWithGemFallbackWhenCrystalInsufficient | 神之庇护（宝石替代） | 水晶不足时红宝石替代消耗，完全减免士气损失 |
| TestGodProtection_DoesNotTriggerOnNonMagicMoraleLoss | 神之庇护（非法术） | 非法术来源士气下降不触发 |
| TestAngelBond_AfterShieldDoesNotReopenActionSelection | 天使羁绊（圣盾后） | 施放圣盾后触发羁绊治疗选择，行动结束后不再重开行动选择窗口 |
| TestAngelBond_OnlyRunsWhenAngelIsRemovalSource | 天使羁绊（触发者校验） | 只有天使本人移除基础效果才触发，其他人移除不触发 |
| TestAngelBond_RunsOnAngelWallShieldPlacement | 天使羁绊（圣盾放置） | 天使之墙放置圣盾后触发羁绊治疗选择 |
| TestAngelBlessing_ReceiverOverHandLimitTriggersOverflowDiscard | 天使祝福（溢出弃牌） | 接收者手牌超上限时触发爆牌弃1 |
| TestAngelBlessing_RejectsDuplicateTwoTargetSelection | 天使祝福（拒绝重复目标） | 选两个相同目标时被拒绝 |
| TestAngelBond_IgnoresSystemBuffRemoval | 天使羁绊（系统移除不触发） | 系统自动移除基础效果时不触发羁绊 |
| TestAngelCleanse_CanPickSpecificBasicEffect | 风之洁净（选择移除） | 目标同时有虚弱和中毒，可选择只移除中毒 |
| TestAngelCleanse_NoBasicEffectSkipsRemovalStep | 风之洁净（无效果跳过） | 目标无基础效果时跳过选择步骤 |
| TestAngel_Skills/AngelWall_PlaceShield | 天使之墙 | 放置圣盾场效 |
| TestAngel_Skills/AngelCleanse_RemoveBuff | 风之洁净 | 移除目标虚弱 |
| TestAngel_Skills/AngelSong_Startup | 天使之歌（启动链） | 启动技→移除基础效果→羁绊治疗→回到行动选择 |
| TestAngel_Skills/AngelBond_PassiveHeal | 天使羁绊（被动） | 移除Buff和使用圣盾时均触发治疗+1 |

---

### 3. 神箭手 (Archer)

**测试文件**：`internal/engine/archer_config_regression_test.go`、`tests/archer_skills_test.go`

| 测试函数 | 技能/机制 | 场景摘要 |
|---|---|---|
| TestArcher_PiercingShotDiscard_IsPublicReveal | 穿透射击（公开弃牌） | 弃牌为公开揭示事件 |
| TestArcher_PiercingShot_NotPromptedOnHit | 穿透射击（命中不触发） | 攻击命中时不在响应技能列表 |
| TestArcher_PiercingShot_PromptedWhenMissByShieldBlock | 穿透射击（圣盾未命中） | 圣盾抵挡导致未命中时触发 |
| TestArcher_LightningArrow_DisablesCounterButAllowsDefend | 雷矢 | 禁止应战但允许防御 |
| TestArcher_PreciseShot_AutoAppliesForceHit | 精准射击（强制命中） | 自动强制命中，无响应提示，目标摸1张 |
| TestArcher_PreciseShot_ForceHitSkipsShield | 精准射击（跳过圣盾） | 强制命中跳过圣盾（不消耗） |
| TestArcher_Skills/PiercingShot_OnMiss | 贯穿射击（完整流程） | 攻击未命中→响应中断→弃法术牌→2点法伤 |
| TestArcher_Skills/Snipe_RefillAndAction | 狙击 | 补满5张手牌+获得额外攻击行动 |
| TestArcher_Skills/FlashTrap_Damage | 闪光陷阱 | 独有牌直伤2点法术伤害 |

---

### 4. 刺客 (Assassin)

**测试文件**：`internal/engine/assassin_backlash_regression_test.go`、`assassin_water_shadow_skip_regression_test.go`、`assassin_config_regression_test.go`、`tests/assassin_skills_test.go`

| 测试函数 | 技能/机制 | 场景摘要 |
|---|---|---|
| TestAssassinBacklash_DoesNotTimingOnMagicDamage | 反噬（法术不触发） | 法术伤害不触发反噬 |
| TestAssassinBacklash_RunsOnAttackDamage | 反噬（攻击触发） | 攻击伤害触发，攻击者摸1张 |
| TestAssassinWaterShadowSkip_ResumesPendingDamageResolution | 水影跳过恢复 | 跳过后恢复伤害结算流程 |
| TestAssassinWaterShadowConfirm_PreservesRemainingDamageDraw | 水影确认保留剩余摸牌 | 弃1水系牌只抵扣1点摸牌，保留剩余 |
| TestAssassinWaterShadow_InterruptsNormalDrawAndPublicReveal | 水影中断摸牌 | 中断正常摸牌，弃牌为公开揭示事件 |
| TestAssassinStealth_DrawChoiceDelaysStealthUntilDrawResolves | 潜行（摸牌延迟） | 选摸牌分支后水影解决前不应用潜行 |
| TestAssassinStealth_HandLimitMinusOneAndReleasesNextStartup | 潜行（手牌上限与释放） | 上限-1(5)，溢出弃1，下回合自动释放恢复6 |
| TestAssassinStealth_DrawSkipResponse_ResumesStartupFlow | 潜行（跳过水影） | 跳过后正常摸牌并应用潜行，不错误进入战斗阶段 |
| TestAssassin_Skills/Backlash_AttackerDraws | 反噬（攻击者摸牌） | 刺客受攻击伤害后攻击者摸1张 |
| TestAssassin_Skills/Stealth_EnterState | 潜行 | 消耗1宝石进入形态，可选择摸牌 |

---

### 5. 吟游诗人 (Bard)

**测试文件**：`internal/engine/bard_regression_test.go`

| 测试函数 | 技能/机制 | 场景摘要 |
|---|---|---|
| TestBardDescentConcerto_RunsAndResolves | 降阶协奏 | 自身法术命中2次后触发，弃2火牌→1点额外法伤+1灵感 |
| TestBardDescentConcerto_DoesNotTimingOnAllyMagicDamage | 降阶协奏（队友不触发） | 队友法术伤害不触发 |
| TestBardDissonanceChord_DrawModeAndReleasePrisoner | 不协和弦 | 消耗灵感(3→1)，释放囚徒形态，选摸牌分支 |
| TestBardHopeFugue_PlaceUsesPlayedCardAsEternalMovement | 希望赋格曲（放置） | 打出的卡作为永恒乐章场卡放到队友身上 |
| TestBardHopeFugue_TransferMovesExistingEternalMovementAndGainsInspiration | 希望赋格曲（转移） | 永恒乐章从队友A转到队友B，弃1手牌，灵感+1 |
| TestBardRousingRhapsody_OnBardTurnStartRunsForbiddenVerse | 激昂狂想曲 | 回合开始触发禁忌诗篇，移除乐章，2名敌人各1法伤+1灵感 |
| TestBardVictorySymphony_AtInspirationCapEntersPrisonerAndSelfDamages | 胜利交响诗（灵感满） | 灵感满3时进囚徒形态，自伤3点 |
| TestBardVictorySymphony_ExtractStoneChoosesGemOrCrystal | 胜利交响诗（提石） | 可选择提取宝石或水晶 |
| TestBardConfig_MetadataAlignsWithDocument | 配置元数据 | 5个核心技能目标类型/数量/专属卡等与文档一致 |
| TestBardStarterExclusiveCards_NotInHand | 起手专属卡 | 手牌4张，3张专属卡在专属区 |

---

### 6. 兽武者 (Beast Samurai)

**测试文件**：`internal/engine/beast_samurai_regression_test.go`

| 测试函数 | 技能/机制 | 场景摘要 |
|---|---|---|
| TestBeastSamurai_InitTokens | 初始化 | 残心=0、兽魂=0、一击标记=0、形态为空 |
| TestBeastSamurai_WarriorZanshinThenOneStrikeBecomesAvailable | 残心/一击无念 | 攻击结束残心+1，一击无念可用 |
| TestBeastSamurai_OneStrike_NextAttackIgnoresShieldAndHoly | 一击无念（无视圣盾） | 下次攻击无视圣盾和圣击，防御选项被移除 |
| TestBeastSamurai_OneStrike_JiFactionForceHit | 一击无念（技系强制命中） | 技系卡强制命中，目标摸2张 |
| TestBeastSamurai_BeastSoulWill_NormalFormHitGainBeastSoul | 兽魂（命中获得） | 普通形态主动攻击命中+1兽魂 |
| TestBeastSamurai_BeastSoulAlert_RunsOnOtherPlayerTapped | 兽魂警报 | 其他角色tapped时触发警报，进入居合形态，残心+1 |
| TestBeastSamurai_BeastReturn_XFlowAndMagicDiscardGainSoul | 兽归（X流程） | 法术伤害触发，选X=1弃1牌，来源弃法术牌回复兽魂 |
| TestBeastSamurai_BeastReturnSkip_ResumesPendingDamageWithoutReprompt | 兽归（跳过） | 跳过后恢复伤害结算不再重复提示 |
| TestBeastSamurai_IaijutsuTurnEndDrainAndZeroExit | 居合形态（回合结束） | 兽魂归零，残心+1，退出居合 |
| TestBeastSamurai_ReversalIaijutsu_ReplacesDamageWithDiscard | 居合反转 | 命中后选X=1令目标弃2替代2伤害 |
| TestBeastSamurai_IaijutsuStyle_CanOverflowBeastSoulAndEnterForm | 居合剑法 | 消耗宝石溢出兽魂，进入居合形态 |

---

### 7. 狂战士 (Berserker)

**测试文件**：`internal/engine/berserker_config_regression_test.go`、`berserker_sealer_damage_regression_test.go`、`tests/berserker_skills_test.go`、`tests/berserker_tear_test.go`

| 测试函数 | 技能/机制 | 场景摘要 |
|---|---|---|
| TestBerserkerTear_CanTimingOnCounterAttackHit | 裂伤（反击命中） | 反击命中时触发，伤害+2(2→4)，消耗1宝石 |
| TestBloodBlade_RunsOnHitCheckForActiveUniqueAttack | 血影狂刀 | 主动独有攻击命中时根据目标手牌数增伤 |
| TestBerserkerAttackSealer_TakeDamageResolves | 狂战攻击封印师 | 雷光斩2伤，封印师承受后摸牌=2+狂化2=4张 |
| TestBerserker_Skills/Frenzy_DamageBonus | 狂化 | 手牌<=3时+1伤，手牌>3时+2伤 |
| TestBerserker_Skills/BloodRoar_ForcedHit | 血腥咆哮 | 目标治疗>=2时强制命中，不可应战 |
| TestBerserker_Skills/BloodBlade_DamageBonus | 血影狂刀 | 对手手牌2时+2伤，手牌3时+1伤 |
| TestBerserker_Tear | 撕裂 | 命中后消耗1宝石，伤害额外+2 |

---

### 8. 剑圣 (BladeMaster)

**测试文件**：`internal/engine/blademaster_response_regression_test.go`、`holy_sword_regression_test.go`、`tests/blademaster_skills_test.go`

| 测试函数 | 技能/机制 | 场景摘要 |
|---|---|---|
| TestBladeMaster_SwordShadow_ReAskOnEachAttackEnd | 剑影重复询问 | 跳过后第二次攻击结束再次询问 |
| TestBladeMaster_WindFury_ReAskOnEachAttackEnd | 风怒追击重复询问 | 同上 |
| TestBladeMaster_WindFury_StillRunsWithoutRemainingWindAttack | 风怒追击（无风牌） | 无剩余风系攻击仍触发，可跳过 |
| TestBladeMaster_GaleSkillExtraAction_PreservedWhenWindFuryCanceled | 疾风技（取消风怒保留额外行动） | 取消风怒不吞掉疾风技额外行动 |
| TestBladeMaster_GaleSkillExtraAction_PreservedWhenWindFuryCanceled_AfterShieldMiss | 疾风技（圣盾未命中） | 圣盾抵挡后取消风怒，保留疾风技额外行动 |
| TestBladeMaster_WindFuryCancel_WithStaleLastActionType_NoRepromptAndKeepExtraAction | 风怒取消（旧分支清理） | 旧LastActionType时取消不重复提示 |
| TestBladeMaster_DiscardContextResume_OnActionEnd_ClearsLastActionCatchup | 弃牌恢复清理 | ActionEnd弃牌恢复时清理LastActionType |
| TestBladeMaster_MultiResponse_ConfirmOneSettlesBeforeRemaining | 多响应确认顺序 | 确认剑影后先结算，风怒追击单独弹出 |
| TestBladeMaster_ResponseChain_SwordShadowThenWindFury_Integration | 响应链集成 | 剑影→风怒追击，共2次额外攻击 |
| TestBladeMaster_GaleSlash_DisablesCounterButAllowsDefend | 列风技 | 禁止应战但允许防御 |
| TestBladeMaster_HolySword_ThirdAttackForceHitIgnoresShield | 圣剑（第3次攻击） | 强制命中无视圣盾 |
| TestBladeMaster_HolySword_FullFlow_X0ResumesExtraAction | 圣剑X=0 | 不摸不弃，直接进额外行动 |
| TestBladeMaster_HolySword_FullFlow_DiscardResumesExtraAction | 圣剑X=1 | 摸1弃1后进额外攻击 |
| TestBladeMaster_Skills/HolySword_ThirdAttack | 圣剑 | 第3次攻击触发摸X弃X中断 |
| TestBladeMaster_Skills/SwordShadow_ExtraAction | 剑影 | 消耗1水晶获额外攻击 |
| TestBladeMaster_Skills/GaleSlash_IgnoreShield | 列风技 | 无视圣盾，不消耗圣盾 |

---

### 9. 烈焰女巫 (Blaze Witch)

**测试文件**：`internal/engine/blaze_witch_skill_regression_test.go`

| 测试函数 | 技能/机制 | 场景摘要 |
|---|---|---|
| TestBlazeWitchPainLink_ConsumesCrystalOnceAndQueuesDiscardToThree | 痛苦链接 | 消耗1水晶→2点法伤→弃牌至3 |
| TestBlazeWitchHeavenfireCleave_AllowsNonFireAttackDiscardInFlameForm | 天火斩（火焰形态） | 非火攻击牌可作为火系弃出 |
| TestBlazeWitchRebirthClock_IncreasesOnMagicMoraleLossWithCap | 轮回计时 | 法术士气损失时+1，上限4 |
| TestBlazeWitchFlameForm_ReleasesAtStartup | 火焰形态释放 | 回合开始自动释放 |
| TestBlazeWitchGetMaxHand_DynamicByRebirthInFlameForm | 手牌上限动态 | 基础6，火焰形态=4+轮回值 |
| TestBlazeWitchCodexAndHeavenfire_RejectSelfTarget | 焚典/天火斩 | 拒绝以自身为目标 |
| TestBlazeWitchFlameForm_AttackUsesPreparedTransformedCard | 火焰形态攻击 | 使用转化后的火系牌攻击 |
| TestBlazeWitchConfig_MetadataAlignsWithDocument | 配置元数据 | 替身人偶/法力逆转等与文档一致 |

---

### 10. 血巫 (Blood Priestess)

**测试文件**：`internal/engine/blood_priestess_regression_test.go`

| 测试函数 | 技能/机制 | 场景摘要 |
|---|---|---|
| TestBloodPriestessSharedLife_DrawBeforePlaceOverflowThenApply | 同生共死（溢出后放置） | 摸牌溢出弃2，再放置，士气-2，上限7 |
| TestBloodPriestessSharedLife_ChoiceAndDrawStayInActionExecution | 同生共死（留在行动执行） | 选择和摸牌均留在行动执行阶段 |
| TestBloodPriestessSharedLife_OverflowDiscardResumesActionExecution | 同生共死（溢出恢复） | 弃牌后恢复行动执行 |
| TestBloodPriestessBleeding_EnterOnMoraleLossAndReleaseOnActionEndLowHand | 流血形态（进入/释放） | 士气损失时进入+1治疗，手牌<3时释放 |
| TestBloodPriestessBleeding_TurnStartSelfDamageBeforeBuff | 流血形态（回合开始自伤） | 自伤摸1张 |
| TestBloodPriestessBloodSorrow_TransferThenRemove | 血之哀伤（转移+移除） | 先转移到p3，再移除恢复专属卡 |
| TestBloodPriestessBloodSorrow_Remove_ShouldEnterBleedWhenDamageCausesMoraleLoss | 血之哀伤（移除致流血） | 移除自伤导致士气损失后进流血 |
| TestBloodPriestessSharedLife_FixedHandCapTargetExempt | 固定上限豁免 | 固定手牌上限目标不受影响 |
| TestBloodPriestessBloodCurse_DiscardPromptAndConfirm | 血咒（弃牌） | 弃3张，手牌减3 |
| TestBloodPriestessBloodCurse_DiscardAllWhenHandInsufficient | 血咒（手牌不足） | 手牌<3时弃全部 |

---

### 11. 蝶舞者 (Butterfly Dancer)

**测试文件**：`internal/engine/butterfly_dancer_regression_test.go`

| 测试函数 | 技能/机制 | 场景摘要 |
|---|---|---|
| TestButterflyLifeFire_MaxHandFloor | 生命火（手牌上限） | 蛹值0→上限6，蛹值2→上限4，蛹值20→下限3 |
| TestButterflyDance_DrawAndGainCocoon | 蝶舞 | 摸1牌+1茧 |
| TestButterflyChrysalis_RunsOverflowDiscardWhenPupaLowersHandLimit | 蛹化（溢出弃牌） | 蛹值+1降低上限，6张手牌触发弃1 |
| TestButterflyReverse_RequiresTwoDiscardCardsByConfig | 逆蝶（弃牌需求） | 手牌仅1张时无法发动 |
| TestButterflyReverse_UsesUnifiedDiscardCostBeforeBranchChoice | 逆蝶（统一弃牌） | 先弃2再选分支 |
| TestButterflyPilgrimage_ResistOneDamage | 朝圣 | 消耗1茧抵御1伤害 |
| TestButterflyMirror_ReplaceTwoDamageToTwoHits | 镜 | 消耗2同系茧将2伤害替换为2次1点 |
| TestButterflyWither_MoraleFloorAtOne | 凋零（士气下限） | 士气为1时损失钳制为0 |
| TestButterflyWither_CanTargetAnyCharacter | 凋零（任意目标） | 可选包括队友在内任何角色 |

---

### 12. 红莲骑士 / 人造人 (Crimson Knight / Hom)

**测试文件**：`internal/engine/crimson_knight_bloody_prayer_regression_test.go`、`crimson_knight_killing_feast_regression_test.go`、`crimson_sword_spirit_config_regression_test.go`、`crimson_sword_spirit_regression_test.go`、`crk_hom_skill_regression_test.go`

| 测试函数 | 技能/机制 | 场景摘要 |
|---|---|---|
| TestCrimsonKnightBloodyPrayer_CanSplitHealToTwoAllies | 血祷（分配治疗） | X=3分配给2名队友(p2+2, p3+1)，获1血印，自伤3 |
| TestCrimsonKnightBloodyPrayer_XOneDirectlyChoosesOtherAlly | 血祷（X=1） | 直接选一名队友+1治疗 |
| TestCrimsonKnightBloodyPrayerXPrompt_NoZeroOption | 血祷X提示 | 不含X=0选项 |
| TestCrimsonKnightKillingFeast_BoostsCurrentHitDamage | 杀戮盛宴（命中增伤） | 伤害+2(2→4)，消耗血印 |
| TestCrimsonKnightKillingFeast_SelfDamageResolvesBeforeAttackDamage | 杀戮盛宴（自伤先结算） | 自伤3点法术伤害先于攻击伤害结算 |
| TestCrimsonKnightCrimsonCross_SelfDamageResolvesBeforeTargetDamage | 绯红十字（自伤先结算） | 自伤先于目标伤害结算 |
| TestCrimsonKnightCrimsonCross_OnlyTargetsEnemy | 绯红十字（仅敌方） | 拒绝以队友为目标 |
| TestCrimsonKnightCalmMind_AutoGrantsEndedActionType | 镇心 | 热血法术行动结束后退出形态+额外法术行动 |
| TestCrimsonKnightHotBlood_AutoReleaseOnTurnEnd | 热血形态（回合结束释放） | 回合结束自动释放+2治疗 |
| TestCrimsonKnightHotBlood_NextTurnNoFallbackRelease | 热血形态（无兜底释放） | NextTurn不做兜底释放 |
| TestCrimsonKnightHotForm_DamageOverflowNoMoraleLoss | 热血形态（溢出不扣士气） | 伤害溢出弃牌不扣除士气 |
| TestCrimsonFlash_PhaseEndDamageShouldNotStall | 绯红闪光（不停滞） | 自伤2点后不停留在Response子流程 |
| TestCrimsonFlash_CombatFlow_DealsExactlyTwoAndKeepsTurnProgressing | 绯红闪光（战斗流程） | 攻击命中后追加1次2点自伤，流程正常推进 |
| TestCrimsonKnightFaith_OnlyWhitelistedSelfDamageCanUseHeal | 绯红信仰（白名单） | 仅白名单自伤可用治疗抵御 |
| TestCrimsonKnightFaith_SelfPoisonCanUseHeal | 绯红信仰（自身中毒） | 中毒伤害允许使用治疗抵御 |
| TestHomRuneReforge_ReallocateAndOverflowCheckOnTurnEnd | 符文改造 | 消耗1宝石重分配战纹/魔纹，回合结束退出触发溢出弃1 |
| TestHomGlyphFusion_MaxXUsesDistinctElements | 符印融合 | X由不同元素数量决定，选1张后过滤同系 |
| TestHomAttackMissResponseGroup_ChooseOneOnly | 未命中响应互斥 | 选怒压后符印融合从列表移除 |
| TestHomDualEcho_TargetChoiceCanCancel | 双重回响（可取消） | 取消不消耗水晶不排伤害 |
| TestHomDualEcho_WhenDamagingEnemyInTwoPlayerGameCanTargetSelf | 双重回响（2人局） | 伤害敌人时可选择自身为替代目标 |
| TestHomDualEcho_TargetConfirmConsumesCostAndQueuesDamage | 双重回响（确认） | 消耗1水晶，排队1点法伤 |
| TestHomDualEcho_NoMoraleDamageStillOverflowsAndDoesNotDropMorale | 双重回响（无士气溢出） | 溢出弃牌触发但不扣士气 |
| TestHomRuneSmash_BurstAddsAttackAndMagicDamage | 符文猛击 | 爆发形态下X=2弃2增伤1点，Y=1翻战纹排队1法伤 |

---

### 13. 元素师 (Elementalist)

**测试文件**：`internal/engine/elementalist_regression_test.go`

| 测试函数 | 技能/机制 | 场景摘要 |
|---|---|---|
| TestElementalistFreeze_RequiresTwoTargets | 冰冻（双目标） | 单目标被拒绝，2名自身目标可发动 |
| TestElementalistMoonlight_ConsumesGemAndRequiresGem | 月光（需宝石） | 无宝石拒绝，有宝石消耗1→3点法伤 |
| TestElementalistIgnite_RequiresThreeElement | 点燃（需3元素） | 元素<3拒绝，=3消耗→2点伤害+额外法术行动 |
| TestElementalistOffenseSkills_RejectAllyTargets | 攻击技能拒绝队友 | 点燃/雷击/风刃/陨石/火球/月光均拒绝队友 |
| TestElementalistThunderStrike_BonusFlow_DirectCardPickWithCancel | 雷击（取消） | 额外弃牌可取消，仅基础1点法伤 |
| TestElementalistThunderStrike_BonusFlow_DirectCardPickWithConfirm | 雷击（确认） | 弃1雷系牌→2点法伤 |

---

### 14. 精灵射手 (Elf Archer)

**测试文件**：`internal/engine/elf_archer_skill_regression_test.go`、`elf_blessing_zone_test.go`、`elf_holy_lancer_bugfix_regression_test.go`

| 测试函数 | 技能/机制 | 场景摘要 |
|---|---|---|
| TestElfElementalShotCostChoice_CancelSupported | 元素射击（取消） | 费用选择可取消 |
| TestElfElementalShotWind_GrantsExtraAttackOnlyAfterActionEnd | 风之矢 | 行动结束后给予额外攻击 |
| TestElfElementalShotWater_AutoResolvesOnCurrentTarget | 水之矢 | 自动为战斗目标+1治疗 |
| TestElfElementalShotThunder_DisablesCounterResponse | 雷之矢 | 不可应战，CanBeResponded=false |
| TestElfRitualRelease_TargetsEnemyOnly | 仪式释放（仅敌方） | 仅包含敌方角色 |
| TestElfRitualStoresBlessingsOutsideHand | 仪式（祝福区） | 消耗1宝石，创建3祝福到手牌外，不触发溢出弃牌 |
| TestElfBlessingCanBePlayedAsMagic | 祝福法术使用 | 作为法术打出产生圣盾场效果 |
| TestElfBlessingCanBePlayedAsAttack | 祝福攻击使用 | 作为攻击打出入战斗栈 |
| TestElfRitualStartupConfirmShouldNotLeaveOverflowDiscard | 仪式启动确认 | 不触发溢出弃牌，3祝福在专属区 |
| TestElfPetEmpower_OverflowConsumesDiscardOnlyOnce | 宠物强化 | 溢出仅触发1次弃牌中断 |
| TestHolyLancer_EarthSpearAndHolyStrikeMutualExclusion | 地枪/圣击互斥 | 地枪可用时圣击不在响应列表 |

---

### 15. 格斗家 (Fighter)

**测试文件**：`internal/engine/fighter_regression_test.go`

| 测试函数 | 技能/机制 | 场景摘要 |
|---|---|---|
| TestFighterPsiField_CapsDamageAtFour | 念力场 | 6点法伤被钳制为4点 |
| TestFighterChargeStrike_HitDamageBonus | 蓄力一击（命中） | 命中后伤害+1(2→3)，气+1 |
| TestFighterChargeStrike_MissSelfDamageByQi | 蓄力一击（未命中） | 自伤摸1张，气+1 |
| TestFighterChargeStrike_ShieldBlockAfterPendingDamageCountsAsMiss | 蓄力一击（圣盾未命中） | 圣盾全挡视为未命中 |
| TestFighterChargeStrike_GrantsQiImmediatelyBeforeCombatResult | 蓄力一击（立即获气） | 确认后立即+1气 |
| TestFighterPsiBullet_TargetChoiceAndSelfDamage | 念力弹 | 法术后选目标→1法伤+1气，自伤摸1 |
| TestFighterHundredDragon_StartupLocksTargetImmediately | 百龙（启动锁定） | 消耗3气选目标，进入百龙形态 |
| TestFighterHundredDragon_BonusesAndTargetLockReleaseStillContinuesAttack | 百龙（加成与锁目标） | 主动攻击+2伤，反击+1伤，违反锁定则释放 |
| TestFighterHundredDragon_CannotActEndsFormAtActionPhaseEnd | 百龙（无法行动） | 清除形态和锁定 |
| TestFighterHundredDragon_MagicAttemptCancelsFormAndAction | 百龙（法术拒绝） | 法术被拒，形态清除 |
| TestFighterHundredDragon_SpecialAttemptCancelsFormAndAction | 百龙（特殊拒绝） | 特殊行动被拒，形态清除 |
| TestFighterHundredDragon_EndsWhenActionPhaseFinishes | 百龙（回合结束） | 自动清除 |
| TestFighterHundredDragon_ActionPromptOnlyKeepsAttackEntry | 百龙（仅攻击） | 隐藏法术/特殊选项 |
| TestFighterBurstCrash_NoCounterAndSelfDamage | 气绝崩击 | 不可应战，消耗1气，自伤摸1 |
| TestFighterWarGodDrive_DiscardToThreeAndHeal | 战神驱动 | 消耗1水晶，弃2至3，+2治疗 |

---

### 16. 勇者 (Hero)

**测试文件**：`internal/engine/hero_regression_test.go`

| 测试函数 | 技能/机制 | 场景摘要 |
|---|---|---|
| TestHeroHeart_InitialCrystalPlusTwo | 勇者之心 | 初始水晶=2 |
| TestHeroRoar_HitDamagePlusTwoAndCleared | 怒吼（命中） | 伤害+2(2→4)，怒吼标记清零 |
| TestHeroRoar_MissAddsWisdom | 怒吼（未命中） | 知性+1，怒吼标记清零 |
| TestHeroForbiddenPower_HitBranch | 禁断之力（命中） | 弃法术增怒+2，消耗1水晶，火牌增伤，额外攻击 |
| TestHeroForbiddenPower_MissBranchWaterToWisdom | 禁断之力（未命中） | 弃水系增知性+2，怒气+2 |
| TestHeroForbiddenPower_UserScenario_Miss_WaterAttackAndMagicToWisdom | 禁断之力（场景-未命中） | 弃[水攻,水法,地攻]：怒气+1，知性+2 |
| TestHeroForbiddenPower_UserScenario_Hit_FireCardsBonusAndSelfDamage | 禁断之力（场景-命中） | 弃[水法,火法,火攻]：火牌2→+2伤，法术2→怒气+1 |
| TestHeroRoar_AfterHitStillPromptsForbiddenPower | 怒吼后禁断之力（命中） | 怒吼结算完仍弹禁断之力 |
| TestHeroRoar_AfterMissStillPromptsForbiddenPower | 怒吼后禁断之力（未命中） | 同上 |
| TestHeroRoar_DrawOneWithOverflow_StillContinuesAttackAndPromptsForbiddenPower | 怒吼（爆牌后继续） | 爆牌后攻击流程继续 |
| TestHeroExhaustion_ReleaseAtActionStartAndSelfDamage_StillCanAct | 枯竭释放（自伤） | 行动开始释放，自伤摸3，回行动选择 |
| TestHeroExhaustion_ReleaseWithOverflow_StillStartsTurnNormally | 枯竭释放（溢出） | 满手牌释放触发爆牌弃3 |
| TestHeroCalmMind_DisablesCounterAndAttackEndGainCrystal | 平心 | 禁止应战，消耗4知性，攻击结束+1水晶 |
| TestHeroTaunt_NonAttackActionIsRejectedUntilSkipOrValidAttack | 挑衅（非攻击拒绝） | 非法术/特殊被拒，跳过后移除 |
| TestHeroTaunt_InvalidAttackDeclarationKeepsEffectUntilValidAttack | 挑衅（无效攻击） | 无效攻击被拒且效果保留 |
| TestHeroDeadDuel_MagicOverflowMoraleLossFlooredToOne | 死斗（士气下限1） | 法术爆牌弃3后士气仅损失1 |

---

### 17. 圣弓 (Holy Bow)

**测试文件**：`internal/engine/holy_bow_regression_test.go`

| 测试函数 | 技能/机制 | 场景摘要 |
|---|---|---|
| TestHolyBow_InitStatsAndTokens | 初始属性 | 水晶=2、治疗上限=3、炮令牌=1、信仰=0 |
| TestHolyBow_HeavenlyBowDamageAdjustments | 天弓被动 | 非圣命格伤害-1，圣命格不变 |
| TestHolyBow_HeavenlyBowHolyHitGainFaith | 天弓命中获信仰 | 圣命格主动命中+1信仰 |
| TestHolyBow_RadiantDescentAndSpecialExitForm | 圣辉降临 | 移除2治疗进圣煌形态+额外法术；特殊行动后脱离+1治疗 |
| TestHolyBow_LightBurstModeA_RequiresOtherAlly | 圣光爆裂A | 需其他队友 |
| TestHolyBow_MeteorBullet_RequiresOtherAlly | 流星弹 | 需其他队友 |
| TestHolyBow_AutoFillActivatedAtTurnEndWithoutSpecial | 自动填充 | 未用特殊行动时回合末弹出选择 |
| TestHolyBow_HolyShardStormMiss_NoBranch | 圣晶风暴（否） | 选否后治疗/队友手牌不变 |
| TestHolyBow_HolyShardStormMiss_NoEligibleAllySkipsPrompt | 无队友跳过 | 无可弃牌队友时跳过 |
| TestHolyBow_HolyShardStormMiss_YesBranch | 圣晶风暴（是） | X=2移除2治疗，队友弃2张 |
| TestHolyBow_HolyShardStormMiss_XChoicesRequireAllyEnoughCards | X受手牌限制 | 队友仅1张时X只允许1 |
| TestHolyBow_LightBurstModeB_XYBoundaries | 圣光爆裂B | X=2选1目标造成Y+2伤害 |
| TestHolyBow_LightBurst_NoAvailableModeCannotUse | 两分支均不可用 | 治疗不足时报错 |
| TestHolyBow_RadiantCannon_MoraleAlignBothSides | 圣光炮 | 选方向后双方士气统一、消耗炮令牌和4信仰 |

---

### 18. 圣枪 (Holy Lancer)

**测试文件**：`internal/engine/holy_lancer_earth_spear_regression_test.go`

| 测试函数 | 技能/机制 | 场景摘要 |
|---|---|---|
| TestHolyLancerRevelation_SyncsMaxHealWithCupState | 圣光启示 | 圣杯相同时MaxHeal=3，落后=2 |
| TestHolyLancerRevelation_UpdatesWhenCampCupChanges | 启示随圣杯变化 | 敌方+1杯后降2，追平恢复3 |
| TestHolyLancerEarthSpear_MaxXUsesCurrentHealValue | 地枪X上限 | 当前治疗4>MaxHeal3时maxX=4 |
| TestHolyLancerEarthSpear_SelectXResumesAttackFlow | 地枪选X | X=3后治疗-3，无残留中断 |
| TestHolyLancerPrayerToken_ClearedAtRealTurnEnd | 祈愈标记 | 有额外行动时不提前清理 |
| TestHolyLancerPrayer_ConsumesExactlyOneGemAndCapsHealAtFive | 祈愈 | 消耗1宝石治疗+1但不超过5 |
| TestHolyLancerRadiance_UsesSelectedWaterDiscard | 辉耀 | 弃水牌后双方各+1治疗 |
| TestHolyLancerPunishment_RejectsSelfAndUsesSelectedMagicDiscard | 惩戒 | 拒绝自身，弃法术后转移1治疗 |
| TestResponsePrompt_PrunesInvalidHolyLancerSkySpear | 响应弹窗修剪 | 治疗不足时天枪被剔除 |
| TestHolyLancer_SkySpearDisablesCounterResponse | 天枪 | 消耗2治疗，元素变暗，禁止应战 |

---

### 19. 魔弓 (Magic Bow)

**测试文件**：`internal/engine/magic_bow_regression_test.go`

| 测试函数 | 技能/机制 | 场景摘要 |
|---|---|---|
| TestMagicBowMagicPierce_MissDealsMagicDamageAndLocksMultiShot | 魔贯未命中 | 3点法伤，同回合连射被锁 |
| TestMagicBowMultiShot_TargetCannotRepeatPrevious | 连射不重复目标 | 上次攻击p2后目标池仅剩p3 |
| TestMagicBowCharge_FollowupPlaceCharges | 充能（后续放置） | 选摸X后选放置数量和牌，消耗水晶 |
| TestMagicBowCharge_DiscardFirstThenChooseX | 充能（先弃后选X） | 手牌超4时先强制弃到4 |
| TestMagicBowCharge_DrawOverflowMoraleLossWithoutDiscard | 充能（超摸士气） | 超摸仅扣士气不弹弃牌 |
| TestMagicBowThunderScatter_ExtraDamageSplit | 雷散额外分配 | 2雷充能对p2造成3点、p3造成1点 |
| TestMagicBowMagicPierce_HitBonusCappedAtTwo | 魔贯命中加伤上限 | 3火充能只消耗1个+2伤害 |
| TestMagicBowMagicPierce_MissDealsExactlyThreeMagicDamage | 魔贯未命中精确3点 | 仅消耗1火充能 |
| TestMagicBowThunderScatter_ExtraZeroSkipsTargetChoice | 雷散X=0 | 跳过目标选择，全体敌人各1点 |
| TestMagicBowCharge_LockTurnDisablesPierceAndScatter | 充能锁回合 | 同回合使用充能后禁用魔贯和雷散 |
| TestMagicBowMagicPierce_HitBonusAutoConsumesSecondCharge | 魔贯自动消耗第二充能 | 追加伤害至3 |
| TestMagicBowDemonEye_TargetPoolExcludesSelf | 魔眼 | 目标池排除自身 |
| TestMagicBowDemonEye_TargetNoHandFallsBackToDrawThreeThenCharge | 魔眼（无手牌） | 目标无手牌摸3张，使用者放1充能 |
| TestMagicBowDemonEye_TargetDiscardsThenUserCharges | 魔眼（弃牌后充能） | 目标弃1张后使用者放1充能+1水晶 |
| TestMagicBowConfig_MetadataAlignsWithDocument | 配置元数据 | 目标类型和数量范围与文档一致 |

---

### 20. 魔枪 (Magic Lancer)

**测试文件**：`internal/engine/magic_lancer_regression_test.go`

| 测试函数 | 技能/机制 | 场景摘要 |
|---|---|---|
| TestMagicLancerDarkRelease_HandCapAndAttackBonusAndLock | 暗影释放 | 上限5，锁充盈/黑枪，首次攻击+1 |
| TestMagicLancerPhantomStardust_LeaveFormAndPromptTarget | 幻影星尘（脱离形态） | 自伤后脱离幻影，弹目标选择 |
| TestMagicLancerPhantomStardust_PreselectedEnemySkipsTargetPrompt | 幻影星尘（预选目标） | 跳过弹窗直接法伤 |
| TestMagicLancerDarkBind_BlocksMagicUseAndDefend | 黑暗束缚 | 禁用法术和防御 |
| TestMagicLancerFullness_FlowBonusAndExtraAttack | 充盈 | 弃牌加攻+1，获额外攻击 |
| TestMagicLancerBlackSpear_ConsumesCrystalAndAddsDamage | 黑枪 | X=2消耗2水晶，2→6伤害 |
| TestMagicLancerConfig_MetadataAlignsWithDocument | 配置元数据 | 目标类型与文档一致 |

---

### 21. 魔剑士 (Magic Swordsman)

**测试文件**：`internal/engine/magic_swordsman_config_regression_test.go`、`magic_swordsman_prayer_css_bugfix_regression_test.go`、`magic_swordsman_shadow_reject_response_regression_test.go`

| 测试函数 | 技能/机制 | 场景摘要 |
|---|---|---|
| TestMagicSwordsmanAsuraCombo_RunsWithoutFireAttackInHand | 阿修罗连招 | 无火攻仍弹出 |
| TestMagicSwordsmanShadowGather_ReleasesBeforeNextActionSelectionPrompt | 暗影聚能 | 行动开始释放后弹启动技能 |
| TestMagicSwordsmanShadowMeteor_TargetsEnemyOnlyAndUsesUnifiedSkillFlow | 暗影流星 | 拒绝队友，选敌方后脱离形态 |
| TestMagicSwordsmanYellowSpring_NoCounterKeepsOriginalElementAndConsumesGem | 黄泉 | 保持火元素、不可应战、消耗1宝石 |
| TestCrimsonBloodBarrier_AutoDamagesSourceWithoutPrompt | 赤血壁垒 | 法伤减至1自动反弹1点给来源 |
| TestCrimsonBloodBarrier_DoesNotRetargetOtherEnemy | 赤血壁垒（不重定向） | 反弹锁定攻击来源 |
| TestPrayerManaTide_RunsAfterMagicActionEnd | 法力潮汐 | 法术行动结束后弹出 |
| TestPrayerSwiftBlessing_StillRunsAfterPhaseEndInterrupt | 迅捷赐福 | 跳过潮汐后继续弹出 |
| TestPrayerSwiftBlessing_AttackFollowupSurvivesPhaseEndResponseInterrupt | 迅捷赐福（攻击后延） | 确认后获额外攻击 |
| TestMagicSwordsmanShadowReject_AllowHolyLightDefendOutsideOwnTurn | 暗影抗拒（圣光防御） | 非己回合可用圣光防御 |
| TestMagicSwordsmanShadowReject_AllowMagicBulletCounterOutsideOwnTurn | 暗影抗拒（魔弹应战） | 非己回合可用魔弹传递 |
| TestMagicSwordsmanShadowGather_PersistsThisTurnAndReleasesNextTurn | 暗影凝聚 | 同回合保持暗影，下回合开始自动退出 |
| TestMagicSwordsmanAsuraCombo_OnlyOncePerTurn | 修罗连斩 | 每回合最多一次 |

---

### 22. 魔法少女 (Magical Girl)

**测试文件**：`internal/engine/magical_girl_config_regression_test.go`、`tests/magicalgirl_skills_test.go`

| 测试函数 | 技能/机制 | 场景摘要 |
|---|---|---|
| TestMagicalGirl_MagicBulletFusion_ReverseChainUsesConfiguredDirection | 魔弹融合（反向链） | 选反向后链目标为p3 |
| TestMagicalGirl_MagicBulletFusion_DeclineKeepsOriginalSpell | 魔弹融合（拒绝） | 拒绝后无链，原法术正常 |
| TestMagicalGirl_MagicBlast_DiscardsAfterEachFailedTarget | 爆裂（弃牌） | 目标拒绝后使用者弃1 |
| TestMagicalGirl_MagicBlast_TargetCanSelectMagicByHandIndex | 爆裂（手牌索引） | 目标选手牌索引3的法术牌弃掉 |
| TestMagicalGirl_DestructionStorm_RequiresTwoTargetsAndCostsOneGem | 毁灭风暴 | 少于2目标报错，消耗1宝石 |
| TestMagicalGirl_Skills/DestructionStorm_AOE | 毁灭风暴 | AOE 2点法伤 |
| TestMagicalGirl_Skills/MagicBlast_GainGem | 魔爆冲击 | 目标不弃则受2法伤，我方+1宝石 |

---

### 23. 月女神 (Moon Goddess)

**测试文件**：`internal/engine/moon_goddess_regression_test.go`

| 测试函数 | 技能/机制 | 场景摘要 |
|---|---|---|
| TestMoonGoddessNewMoonShelter_AbsorbsOverflowAndPreventsMoraleLoss | 新月庇护 | 吸收爆牌、进暗月、士气不变 |
| TestMoonGoddessNewMoonShelter_NoSoulDevourGainWhenMoraleLossPrevented | 庇护抵消后无黄魂 | 重复24次验证 |
| TestMoonGoddessNewMoonShelter_NotDispatchWhenActualMoraleWillNotDrop | 实际不降士气不触发 | 热血形态爆牌不触发 |
| TestMoonGoddessMoonCycle_Branch1AppliesCurseAndHeal | 月轮分支1 | 消耗暗月，脱离形态，士气-1，队友+1治疗 |
| TestMoonGoddessMoonCycle_OnlyOncePerTurn | 月轮仅1次 | 第二次被阻止 |
| TestMoonGoddessMoonCycle_Branch1NoRepromptBranch2InDriveFlow | 选分支1后不重复弹 | Drive流中不重复弹出 |
| TestMoonGoddessMoonCycle_TurnStateLatchPreventsRepromptWhenTokenResets | TurnState锁防重弹 | 完成分支1后状态锁阻止再次派发 |
| TestMoonGoddessDarkMoonSlash_AddsDamageAndConsumesDarkMoon | 暗月斩 | X=2攻击+2，消耗暗月，诅咒-2士气 |
| TestMoonGoddessDarkMoonSlash_XBoundaries_CurseAndDamage | 暗月斩X边界 | X=1伤害3/暗月剩1，X=2伤害4/暗月清0 |
| TestMoonGoddessMedusa_ExcludesConvertedAttacks | 美杜莎排除转换攻击 | 欺诈/圣晶风暴转换的攻击不触发 |
| TestMoonGoddessMedusa_OnlyAtAttackStart | 美杜莎仅攻击开始 | 非攻击开始不触发 |
| TestMoonGoddessMedusa_MagicDarkMoonExtraDamageTargetsAttackerOnly | 法术暗月额外伤 | 仅指向攻击者 |
| TestMoonGoddessBlasphemy_OncePerTurnAndResetNextTurn | 渎神每回合1次 | 同回合第2次被阻止 |
| TestMoonGoddessBlasphemy_TargetLockedToDamagedEnemyAndSelfTurn | 渎神锁定 | 仅弹受伤敌人，非己回合不触发 |
| TestMoonGoddessPaleMoon_Branch1GrantsExtraTurn | 苍月分支1 | 消耗1宝石，不可应战，额外攻击 |
| TestMoonGoddessPaleMoon_Branch2RequiresNewMoonAndXStartsAtOne | 苍月分支2 | 需新月，X=1消耗1新月，石化+1，2法伤 |

---

### 24. 阴阳师 (Onmyoji)

**测试文件**：`internal/engine/onmyoji_rules_regression_test.go`、`onmyoji_skill_flow_regression_test.go`

| 测试函数 | 技能/机制 | 场景摘要 |
|---|---|---|
| TestOnmyojiDarkRitual_ChoosesTargetAtTurnEnd | 暗夜仪式 | 回合末选敌方→鬼火清0→2点法伤 |
| TestOnmyojiBinding_RequiresGemAndCrystal | 式神咒束 | 需宝石+水晶 |
| TestOnmyojiLifeBarrier_Mode1_X3NoMoraleLoss | 生之结界分支1 | X=3自伤爆牌弃3，士气不变 |
| TestOnmyojiLifeBarrier_Mode2_ReleaseFormAndForceAllyDiscard | 分支2 | 弃同命格脱离形态，队友强制弃牌 |
| TestOnmyojiLifeBarrier_PreselectedTargetSkipsSecondTargetChoice | 预选目标 | 跳过二次选择，+1宝石+1治疗 |
| TestOnmyojiYinYangConfirm_PrioritizedBeforeNormalCombatPrompt | 阴阳确认 | 选否后回常规战斗响应 |
| TestOnmyojiYinYangConfirm_YesBranchResolvesFactionCounterChain | 阴阳"是" | 脱离形态、鬼火=3、反射战斗链 |
| TestOnmyojiShikigamiDescend_RequiresSameFactionDiscards | 式神降临 | 不同命格报错，同命格弃2→进形态+1鬼火+额外攻击 |
| TestOnmyojiYinYangShift_InShikigamiForm | 式神形态下阴阳转换 | 脱离形态、鬼火=3、反射战斗 |

---

### 25. 祭司 (Priest)

**测试文件**：`internal/engine/adventurer_priest_rules_regression_test.go`

| 测试函数 | 技能/机制 | 场景摘要 |
|---|---|---|
| TestPriestDivineDomain_HealBranchRequiresTwoDiscards | 神圣领域（治疗分支） | 弃2牌+1水晶，自身+2队友+1治疗 |
| TestPriestDivineDomain_RejectsPartialDiscard | 神圣领域（弃牌不足） | 手牌不足2张报错 |
| TestPriestDivineDomain_DamageBranchTargetsAnyPlayer | 神圣领域（伤害分支） | 排除自身，消耗1治疗 |
| TestPriestWaterPower_DiscardWaterThenGiveSelectedCard | 水之力量 | 弃水后送牌给队友，双方+1治疗 |
| TestPriestWaterPower_RequiresTransferCard | 水之力量（需送牌） | 仅1张水牌无法送牌时报错 |
| TestPriestDivineRevelation_RunsOnlyOncePerSpecialAction | 神圣启示 | 每次特殊行动仅1次 |
| TestPriestDivineContract_HasXChoiceAndCapsTargetAt4 | 神圣契约 | X=2后自身-2治疗，队友封顶4 |
| TestPriestDivineContract_TargetAlreadyAbove4KeepsUnchanged | 契约（超上限不变） | 治疗5时保持不变 |

---

### 26. 瘟疫法师 (Plague Mage)

**测试文件**：`internal/engine/plague_mage_skill_regression_test.go`、`role_skill_bugfix_regression_test.go`、`combat_magic_role_fix_regression_test.go`

| 测试函数 | 技能/机制 | 场景摘要 |
|---|---|---|
| TestPlagueOutbreak_UsesTurnEndRewardInsteadOfImmediateHeal | 瘟疫爆发 | 发动时+1治疗，回合末再+1 |
| TestPlagueDeathTouch_TargetsEnemyOnlyAndSuppressesImmortal | 死亡之触 | 仅敌方，治疗-2，弃2同系，压制不灭 |
| TestPlagueDeathTouch_CancelChoiceRestoresActionWindow | 死亡之触（首步取消） | 治疗/手牌/行动均不消耗 |
| TestPlagueDeathTouch_CancelAtCardPickRestoresActionWindow | 死亡之触（弃牌步取消） | 同上 |
| TestPlagueToxicNova_ConsumesExactlyOneGem | 毒之新星 | 消耗1宝石，+1治疗+不灭1次 |
| TestPlagueMageCannotUseHealAgainstAttackDamage | 治疗（攻击伤害不可用） | 治疗只能抵消法术伤害 |
| TestPlagueMageCanUseHealAgainstMagicDamage | 治疗（法术伤害可用） | 弹出治疗选择 |

---

### 27. 祈祷师 (Prayer Master)

**测试文件**：`internal/engine/prayer_master_config_regression_test.go`

| 测试函数 | 技能/机制 | 场景摘要 |
|---|---|---|
| TestPrayerEnterForm_FixesMaxHandAtFive | 祈祷入阵 | 消耗1宝石进形态，上限=5 |
| TestPrayerPowerBlessing_ConsumesExclusiveZoneCardDirectly | 威力赐福 | 从专属区消耗，队友获场效 |
| TestPrayerSwiftBlessing_ConsumesExclusiveZoneCardDirectly | 迅捷赐福 | 从专属区消耗，队友获场效 |

---

### 28. 贤者 (Sage)

**测试文件**：`internal/engine/sage_skill_regression_test.go`

| 测试函数 | 技能/机制 | 场景摘要 |
|---|---|---|
| TestSageMagicRebound_SameElementDiscardChain | 法术反弹 | 确认→选X=3→弃3同系→选目标，自己3点+目标2点 |
| TestSageMagicRebound_DispatchAfterDamageDraw | 反弹（承伤摸牌后） | 摸牌完成后再弹确认 |
| TestSageMagicRebound_TwoOneMagicDamagesPromptTwice | 连续2次1点法伤 | 逐条触发 |
| TestSageWisdomCodex_ForceDiscardAfterHeavyMagicDamage | 智慧法典 | 受4点法伤后强制弃1，获2宝石 |
| TestSageArcaneCodex_TargetPoolExcludesSelfAndSelfDamageStillRunsRebound | 魔道法典 | 排除自身，自伤1点弹出反弹确认 |
| TestSageHolyCodex_XAndTargetCountBoundaries | 圣道法典 | X越界报错，X=4弃4选2目标各+2治疗 |
| TestSageExtract_CanReachFourthEnergyAndStopsAtCap | 提炼 | 能量=4封顶 |
| TestSageConfig_MetadataAlignsWithDocument | 配置元数据 | 与文档一致 |

---

### 29. 圣女 (Saintess)

**测试文件**：`internal/engine/saintess_config_regression_test.go`、`saintess_frost_prayer_resume_regression_test.go`、`tests/saintess_skills_test.go`

| 测试函数 | 技能/机制 | 场景摘要 |
|---|---|---|
| TestSaintess_SaintHeal_TwoTargetSplitCanChooseMagicExtraAction | 圣疗（2目标） | 自身1队友2治疗，选额外法术 |
| TestSaintess_SaintHeal_ThreeTargetsApplyHealAfterExtraActionChoice | 圣疗（3目标） | 选额外攻击后3目标各+1治疗 |
| TestSaintess_Mercy_BecomesPersistentFixedHandCapState | 怜悯 | 消耗1宝石获1水晶，上限=7，持久不弹 |
| TestSaintessFrostPrayer_DefendChoiceDoesNotReopenActionSelection | 冰霜祈愈 | 防御后不重新打开行动选择 |
| TestSaintess_Skills/FrostPrayer_HealDispatch | 冰霜祷言 | 水系牌时目标+1治疗 |
| TestSaintess_Skills/HealingLight_MultiHeal | 治愈之光 | 多目标+1治疗 |
| TestSaintess_Skills/SaintHeal_Distribute | 圣疗 | 分配3点治疗+选额外行动 |

---

### 30. 封印师 (Sealer)

**测试文件**：`internal/engine/sealer_status_resolver_regression_test.go`、`exclusive_skill_card_regression_test.go`、`field_card_flow_regression_test.go`、`tests/sealer_skills_test.go`、`tests/seal_break_test.go`

| 测试函数 | 技能/机制 | 场景摘要 |
|---|---|---|
| TestFiveElementsBind_BuffPhaseChoiceUsesSealCountCapAtTwo | 五系束缚（封印数上限） | 3封印时摸牌4张（封顶） |
| TestFiveElementsBind_DrawCancelRemovesStatusAndResumesStartup | 五系束缚（取消） | 取消后摸3张、束缚移除、回行动开始 |
| TestElementalSeal_RevealedDiscardRunsButHiddenDiscardDoesNot | 元素封印（公开弃牌） | 公开弃牌触发法伤，隐藏不触发 |
| TestElementalSeal_UsesBoundElementMetaForMatching | 元素封印（元素匹配） | 绑定水后水牌触发、火牌不触发 |
| TestSealBreak_CanPickGlobalBasicEffectWithoutPreselectedTarget | 封印破除 | 无预选目标取全局效果，选火封印收入手 |
| TestSealerStarterExclusiveCard_NotInHand | 开局专属卡 | 专属牌（五系束缚）在专属区 |
| TestFiveElementsBind_UsesExclusiveZoneCard | 五系束缚消耗 | 从专属区消耗牌并放置场效 |
| TestSealer_SealBreak | 封印破碎 | 消耗1水晶，选圣盾移除后收入手 |
| TestSealer_Skills/MagicSurge_ExtraAction | 法术激荡 | 法术后确认获额外攻击 |
| TestSealer_Skills/FiveElementsBind_SkipAction | 五系束缚 | 消耗1水晶放置，下回合跳过行动 |

---

### 31. 灵魂术士 (Soul Sorcerer)

**测试文件**：`internal/engine/soul_sorcerer_regression_test.go`

| 测试函数 | 技能/机制 | 场景摘要 |
|---|---|---|
| TestSoulSorcerer_StartGameInitAndStarterCard | 初始化 | 蓝/黄魂=0，拥有灵魂链接 |
| TestSoulSorcererSoulDevour_OnlyRunsOnDamageOverflowMoraleLoss | 灵魂吞噬 | 仅爆牌士气损失触发 |
| TestSoulSorcererSoulDevour_UsesFinalAppliedMoraleLoss | 吞噬实际损失 | 实际损失被改写时按最终值 |
| TestSoulSorcererSoulRecall_MultiSelectSubmit | 灵魂追忆 | 弃1法术后蓝魂+1 |
| TestSoulSorcererSoulConvert_OnAttackStartChoice_DoesNotStallInResponsePhase | 灵魂转换 | 选颜色后不卡空响应阶段 |
| TestSoulSorcererSoulConvert_ChoiceFallback_NoUserCtxShouldNotStayResponse | 转换回退 | 缺user_ctx时不卡响应 |
| TestSoulSorcererSoulMirror_DrawUpToMaxHand | 灵魂之镜 | 弃2牌消耗2黄魂，目标补至6 |
| TestSoulSorcererSoulMirror_UsesDynamicMaxHand | 之镜（动态上限） | 勇者枯竭上限4，补至4 |
| TestSoulSorcererSoulBlast_ConditionalBonusDamage | 灵魂震爆 | 消耗3黄魂，上限>5时伤害5 |
| TestSoulSorcererSoulBlast_NoBonusWhenDynamicMaxHandIsNotGreaterThanFive | 震爆（无加伤） | 上限<=5时仅3 |
| TestSoulSorcererSoulGrant_RespectsEnergyCap | 灵魂赐予 | 目标能量封顶3 |
| TestSoulSorcererSoulAmp_ConsumesGemAndCapsSouls | 灵魂增幅 | 消耗1宝石，黄/蓝魂封顶6 |
| TestSoulSorcererSoulLink_TransferDamageBeforeResolve | 灵魂链接（转伤） | 队友承伤选转移2点 |
| TestSoulSorcererSoulLink_Replay_TransferSorcererToAlly_NoRecursiveLinkPrompt | 术士转队友不递归 | 转移法伤不再触发链接 |
| TestSoulSorcererSoulLink_Replay_TransferAllyToSorcerer_NoRecursiveLinkPrompt | 队友转术士不递归 | 同上 |
| TestSoulSorcererSoulLink_Replay_TransferDamageThenRunsResponseChain | 转伤后响应链 | 转伤给勇者触发死斗 |

---

### 32. 灵符师 (Spirit Caster)

**测试文件**：`internal/engine/spirit_caster_regression_test.go`

| 测试函数 | 技能/机制 | 场景摘要 |
|---|---|---|
| TestSpiritCasterTalismanThunder_SealThenIncantThenDamage | 灵符雷 | 封印→念咒→2目标逆序雷伤 |
| TestSpiritCasterIncantation_NoCapStillPromptsAndResolvesWind | 念咒（无上限） | 选风符+念咒盖放，2目标各弃1 |
| TestSpiritCasterHundredNight_FireRevealAOEWithCollapse | 百夜（火系AOE） | 火妖力+展示+崩解，2目标各2法伤 |
| TestSpiritCasterHundredNight_NonFireSingleTarget | 百夜（非火单目标） | 选水妖力后1点法伤 |
| TestSpiritCasterConfig_MetadataAlignsWithDocument | 配置元数据 | 与文档一致 |

---

### 33. 剑皇 (Sword Emperor)

**测试文件**：`internal/engine/sword_emperor_regression_test.go`

| 测试函数 | 技能/机制 | 场景摘要 |
|---|---|---|
| TestSwordEmperor_InitTokens | 初始化 | 剑气=0、剑魂=0、护卫禁用=0 |
| TestSwordEmperor_MissAddsSwordSoulAndSwordQi | 未命中 | +1剑魂+1剑气 |
| TestSwordEmperor_SwordSoulGuard_StopsAtCap | 剑魂守卫上限 | 3剑魂已达上限，未命中只+1剑气 |
| TestSwordEmperor_AngelSoul_MissDisablesGuardAndAddsMorale | 天使魂（未命中） | 禁护卫、+1剑气、红士气+1 |
| TestSwordEmperor_AngelSoul_HitHealsTwo | 天使魂（命中） | 治疗+2 |
| TestSwordEmperor_DemonSoul_HitAddsDamage | 恶魔魂（命中） | 伤害+1(共3) |
| TestSwordEmperor_DemonSoul_MissAddsTwoSwordQi | 恶魔魂（未命中） | +2剑气 |
| TestSwordEmperor_SwordQiSlash_ExcludeOriginalTargetAndDealMagicDamage | 剑气斩 | X=2对p3造成2法伤，剑气-2 |
| TestSwordEmperor_IndomitableWill_DrawQiAndExtraAttack | 不屈意志 | 消耗1水晶、摸1牌、剑气+1、额外攻击 |

---

### 34. 女武神 (Valkyrie)

**测试文件**：`internal/engine/valkyrie_combo_regression_test.go`、`valkyrie_config_regression_test.go`

| 测试函数 | 技能/机制 | 场景摘要 |
|---|---|---|
| TestValkyrie_ComboChain_FullFlow | 连招完整流程 | 秩序之印→神圣追击→额外攻击→英灵召唤→弃牌→目标+1治疗 |
| TestValkyrie_HeroicSummon_CancelDoesNotRepromptSameHit | 英灵召唤取消 | 取消不重复弹出 |
| TestValkyrie_HeroicSummon_ExtraDiscardHealsCurrentCombatTarget | 英灵召唤弃牌 | 弃法术后战斗目标+1治疗 |
| TestValkyrie_HeroicSummon_ExtraDiscardCanCancel | 英灵召唤（取消） | 手牌保留、目标不加治疗 |
| TestValkyrie_HeroicSummon_DoesNotEnterSpiritOnCounterHit | 应战不进英灵 | 应战命中后不进入形态 |
| TestValkyrie_MilitaryGlory_BranchTwoDoesNotExitSpirit | 军威分支2 | 保持英灵形态，队友+2治疗 |
| TestValkyrie_PeaceWalker_ReleasesSpiritOnActiveAttack | 和平行者 | 主动攻击脱离英灵 |

---

### 35. 仲裁者 (Arbiter)

**测试文件**：`internal/engine/arbiter_law_regression_test.go`

| 测试函数 | 技能/机制 | 场景摘要 |
|---|---|---|
| TestArbiterLaw_GrantsInitialCrystalAndDoesNotReactivateOnTurnStart | 仲裁法则 | 初始水晶=2，回合开始不重复 |
| TestArbiterForm_JudgmentAutoGainAtStartup | 审判形态 | 启动时审判+1 |
| TestArbiterRitual_EntersFormWithoutImmediateJudgment | 仪式进形态 | 消耗1宝石，上限5，审判不变 |
| TestArbiterRitualBreak_RestoresHandLimitAndAddsTeamGem | 仪式解除 | 退出形态，上限恢复6，红方+1宝石 |
| TestArbiterForcedDoomsday_HappensAfterStartupAndTargetsEnemiesOnly | 强制末日 | 仅指敌方，清空审判 |
| TestArbiterForcedDoomsday_IgnoresTauntAndClearsItAfterResolution | 强制末日（无视挑衅） | 末日期间覆盖挑衅，使用后清除 |
| TestArbiterBalance_BranchesFollowConfiguredEffects | 天平分支 | 分支0全弃；分支1补至上限+1宝石 |

---

## 二、引擎通用机制测试

**目录**：`internal/engine/`（非角色专属）

| 分类 | 测试文件 | 测试函数 | 验证内容 |
|---|---|---|---|
| 基础效果 | basic_effect_before_action | TestBuffResolve_PoisonResolvesBeforeWeaknessChoice | 中毒先于虚弱结算 |
| 基础效果 | basic_effect_before_action | TestBeforeActionHooks_PoisonEntersDamageResolutionBeforeWeakness | 中毒进入伤害结算后回到行动前 |
| 基础效果 | basic_effect_before_action | TestWeaknessPrompt_OrderMatchesConfig | 虚弱选项顺序与配置一致 |
| 基础效果 | basic_effect_before_action | TestWeaknessChoiceMappingMatchesConfig/skip_action_phase | 跳过行动：移除虚弱，不摸牌 |
| 基础效果 | basic_effect_before_action | TestWeaknessChoiceMappingMatchesConfig/draw_three_then_continue | 摸3张：移除虚弱，摸3，继续行动 |
| 基础效果 | basic_effect_stack | TestPerformMagic_PoisonCannotStackOnSameTarget | 同目标不可叠加中毒 |
| 基础效果 | basic_effect_stack | TestUseSkill_BasicEffectPlacementCannotStack | 天使之墙同目标不可叠加圣盾 |
| 基础效果 | basic_effect_stack | TestUseSkill_AngelWallCanTargetEnemy | 天使之墙可指定敌方 |
| 选择系统 | choice_binding_completeness | TestChoiceCatalogRouteSpecTableMatchesCatalogFile | 目录文件与路由表双向同步 |
| 选择系统 | choice_unknown_choice_type | TestChoiceEngineUnknownChoiceTypeErrors | 未注册choice_type返回错误 |
| 战斗规则 | counter_attack_action_gating | TestBloodRoar_ForcedHitIgnoresShield | 血腥咆哮强制命中无视圣盾 |
| 战斗规则 | counter_attack_action_gating | TestSealBreak_SelectSpecificBasicEffectAndTakeCard | 封印破碎选特定效果收入手 |
| 战斗规则 | dark_counter_visibility | TestDarkAttack_CombatRequestNotRespondable | 暗灭不可应战 |
| 战斗规则 | shield_defend_rule | TestCombatDefend_CannotPlayShieldFromHand | 防御阶段不能打手牌圣盾 |
| 战斗规则 | shield_defend_rule | TestCombatDefend_HolyLightStillValid | 圣光防御正常 |
| 战斗规则 | shield_defend_rule | TestMagicBulletDefend_CannotPlayShieldFromHand | 魔弹防御不能打手牌圣盾 |
| 战斗规则 | shield_defend_rule | TestMagicBullet_FieldShieldAutoBlocks | 圣盾延迟触发挡魔弹 |
| 战斗规则 | shield_defend_rule | TestCombatShield_WaitsForPlayerChoice | 不因圣盾跳过玩家选择 |
| 战斗规则 | shield_defend_rule | TestCombatShield_ConsumeOnTake | 承受时圣盾挡伤并消耗 |
| 战斗规则 | shield_defend_rule | TestCombatShield_CounterChoiceKeepsShield | 应战不消耗圣盾 |
| 资源替代 | crystal_substitute | TestCrystalSubstitute_SwordShadow_ResponseSkill | 红宝石替代蓝水晶（响应技能） |
| 资源替代 | crystal_substitute | TestCrystalSubstitute_ArbiterBalance_ActionSkill | 红宝石替代蓝水晶（主动技能） |
| 资源替代 | crystal_substitute | TestCrystalSubstitute_GemCostCannotUseCrystal | 蓝水晶不可替代红宝石（单向） |
| 魔弹 | magic_bullet_auto_target | TestMagicBullet_AllowsMagicWithoutExplicitTarget | 魔弹无需显式目标 |
| 魔弹 | magic_bullet_auto_target | TestMagicWithoutTarget_StillRequiresTargetForNonMagicBullet | 非魔弹法术仍需目标 |
| 弃牌子流程 | discard_subflow_invariant | TestEnterDiscardSelection_RequiresMatchingPendingInterrupt | 匹配中断才能进入弃牌子流程 |
| 弃牌子流程 | discard_subflow_invariant | TestIsDiscardSelectionActive_RequiresLifecycleConsistency | 子流程激活需标记+中断同时存在 |
| 弃牌子流程 | discard_subflow_invariant | TestPopInterrupt_ClearsDiscardSelectionSubflow | 弹出中断清除子流程 |
| 弃牌子流程 | discard_subflow_invariant | TestDriveDiscardSelectionPhase_DoesNotAutoRepairOnMismatch | 不匹配时不自动修复 |
| 行动选择 | action_selection_prompt_options | TestActionSelectionPrompt_ExtraAttackOnlyShowsAttack | 额外攻击仅显示攻击选项 |
| 行动选择 | action_selection_prompt_options | TestActionSelectionPrompt_ExtraMagicOnlyShowsMagic | 额外法术仅显示法术选项 |
| 行动选择 | action_selection_prompt_options | TestActionSelectionPrompt_ExtraAttackNoLegalActionShowsSkip | 无合法行动显示跳过 |
| 行动选择 | action_selection_prompt_options | TestActionSelection_ExtraActionCannotActSkipsWhenNoLegalAction | 无合法行动时跳过成功 |
| 行动选择 | action_selection_prompt_options | TestActionSelection_ExtraMagicAllowsSkill | 额外法术中可使用技能 |
| 行动选择 | action_selection_prompt_options | TestActionSelectionPrompt_ArbiterForcedDoomsdayOnlyShowsMagic | 强制末日仅显示法术 |
| 行动选择 | action_selection_prompt_options | TestActionSelectionPrompt_TauntWithoutAttackOnlyShowsSkip | 挑衅无攻击只能跳过 |
| 行动选择 | action_selection_prompt_options | TestActionSelectionPrompt_MagicSwordsmanShadowForm_StillShowsMagicWhenSkillUsable | 暗影形态技能可用时显示法术入口 |
| 提炼取消 | extract_cancel | TestExtractCancel_ReturnsToActionSelection | 取消后回到行动选择 |
| 时序 | stage_timing | TestStartupSkill_WindowSeparatedFromTurnStartTiming | 启动阶段与回合开始时序分离 |
| 时序 | startup_skip | TestStartupSkillSkip_OnlyPromptsOncePerTurn | 跳过后不再重复弹出 |
| 时序 | startup_skip | TestStartupSkillConfirm_EndsStartupPhaseAfterOneSkill | 确认一个后启动阶段立即结束 |
| 时序 | startup_skip | TestStartupSkillConfirm_DisablesSpecialActionsInSameTurn | 启动技能后锁定特殊行动 |
| 技能调度 | skill_dispatcher_priority | TestBeforeMoraleLoss_UsesSkillPriorityOrdering | 士气损失前按优先级排序 |
| 技能挂载 | skill_policy_mount | TestSkillPolicyMountedFromRoleEntriesForPilotRoles | 策略从角色注册表自动挂载 |
| 技能挂载 | skill_policy_mount | TestMountPlayerSkillPolicySpecsRoleEntryOverridesLegacy | 角色注册表策略覆盖旧策略 |
| 后续挂载 | followup_mount | TestMountPlayerDeferredFollowupSpecs_MountsAndExecutes | FollowupSpec自动挂载执行 |
| 后续挂载 | followup_mount | TestMountPlayerDeferredFollowupSpecs_SkipsEmptyResolver | 空Resolver不挂载 |
| 后续挂载 | followup_mount | TestSpiritCasterFollowupMountedFromRoleEntry | 灵符师旧机制已迁移验证 |
| 祈祷形态 | prayer_form_persist | TestPrayerForm_PersistsAfterTurnEnd | 祈祷形态回合结束不退出 |
| 恢复点 | resume_point | TestParseChoiceResumeTurnStage_BareBeforeActionStaysExactStage | 裸字符串panic |
| 恢复点 | resume_point | TestParseChoiceResumeTurnStage_ExplicitStageRoundTrips | 枚举值往返正确 |
| 恢复点 | resume_point | TestCurrentChoiceResumePoint_InvalidTurnStagePanics | 无效阶段panic |
| 恢复点 | resume_point | TestCurrentChoiceResumePoint_NilEnginePanics | 空引擎panic |
| 调试 | debug_cheat_controls | TestDebugCheat_EffectSetCount | 调试指令设效果数量 |
| 调试 | debug_cheat_controls | TestDebugCheat_CardFiltersAndExclusive | 调试卡牌筛选与发放 |
| 调试 | debug_cheat_controls | TestDebugCheat_DiscardByCount | 调试弃牌操作 |
| 调试 | debug_sword_interrupt | TestDebugHolySwordInterruptFiring | 调试圣剑触发 |

---

## 三、服务器层测试

**目录**：`internal/server/`

| 分类 | 测试文件 | 测试函数 | 验证内容 |
|---|---|---|---|
| 房间座位 | room_seating | TestBuildInterleavedLineup_AvoidsThreeSameCampInRow | 3v3交错无3同阵营连续 |
| 房间座位 | room_seating | TestBuildInterleavedLineup_ThreeVsOneStillAvoidsTriple | 3v1仍无3同阵营 |
| 房间座位 | room_seating | TestOrderedClientIDsLocked_SeatOrderFirst | 按座位顺序排客户端ID |
| 技能适配 | available_skill_adapter_placecard | TestBuildAvailableActionSkills_SealerPlaceCardSkillsNotBlocked | PlaceCard类不被延迟CanUse误挡 |
| 技能适配 | available_skill_adapter_placecard | TestBuildAvailableActionSkills_HeroTauntStillRequiresAngerToken | 挑衅需怒气 |
| 技能适配 | available_skill_adapter_placecard | TestBuildAvailableActionSkills_AngelCleanseExposedAsNoTargetSkill | 风之洁净无目标类型 |
| 机器人 | bot_auto_action | TestBotAutoRespondsInCombat | 机器人3秒内自动响应战斗 |
| 机器人 | bot_auto_action | TestRunBotTurnIgnoresStaleCombatPrompt | 忽略过期战斗提示 |
| 机器人 | bot_auto_action | TestPromptActionableForMagicMissileInterrupt | 魔弹中断对机器人可操作 |
| 元素师技能 | elementalist_available_skills | TestBuildAvailableActionSkills_ElementalistIgniteAndMoonlightGating | 点燃/月光可见性门控 |
| 元素师技能 | elementalist_available_skills | TestBuildAvailableActionSkills_ForcedDoomsdayOnlyShowsDoomsday | 强制末日仅显示末日审判 |
| 协议适配 | protocol_adapter | TestTranslateClientAction_AttackUsesUUIDAndTargets | UUID到索引转换 |
| 协议适配 | protocol_adapter | TestHandleAction_NotStartedUsesNotifyTimelineEnvelope | 未开始时返回错误信封 |
| 协议适配 | protocol_adapter | TestBuildTimelineNotify_DamageEvent | 伤害事件时间线结构 |
| 协议适配 | protocol_adapter | TestBuildRequireActionPayload_UsesStructuredPromptField | 结构化prompt字段 |
| 协议适配 | protocol_adapter | TestBuildSyncStatePayload_UsesStructuredFields | 结构化同步字段 |
| 角色校验 | roles_validation | TestAvailableRolesSyncWithCharacterData | 角色列表与数据同步 |
| 角色校验 | roles_validation | TestValidateLineupRejectsDuplicateRoles | 拒绝重复角色 |
| 重连接管 | room_reconnect_takeover | TestRoomUnregister_StartedHumanKeepsReconnectableSeat | 断线保留座位 |
| 重连接管 | room_reconnect_takeover | TestRoomHostCanTakeoverDisconnectedPlayerWithBot | 房主接管转机器人 |
| 重连接管 | room_reconnect_takeover | TestRoomReconnectByRoomAndNameWithoutToken | 姓名匹配重连 |
| 重连接管 | room_reconnect_takeover | TestRoomReconnectByPlayerIDWithoutToken_WhenSeatDisconnected | 玩家ID重连 |
| 重连接管 | room_reconnect_takeover | TestRoomReconnectByPlayerIDRejected_WhenSeatOnline | 在线时拒绝重连 |
| 严格协议 | strict_protocol | TestHandleMessage_UnknownCmdReturnsProtocolError | 未知命令返回协议错误 |
| 严格协议 | strict_protocol | TestHandleAction_InvalidJSONReturnsProtocolError | 无效JSON返回协议错误 |
| 严格协议 | strict_protocol | TestHandleAction_UnknownActionTypeReturnsProtocolError | 未知行动类型返回协议错误 |

---

## 四、集成测试

**目录**：`tests/`

| 分类 | 测试文件 | 测试函数 | 验证内容 |
|---|---|---|---|
| 核心引擎 | engine_test | TestBladeMaster_WindFury_ExtraAttack | 风怒追击额外攻击 |
| 核心引擎 | engine_test | TestSealer_FiveSeals | 5种封印表驱动覆盖 |
| 核心引擎 | engine_test | TestAngel_AngelBlessing | 天使祝福给牌中断 |
| 魔弹 | magic_bullet_test | TestMagicBullet_ChainAndDamage | 魔弹传递链+伤害递增 |
| 魔弹 | magic_bullet_test | TestMagicBullet_Defend | 圣光防御魔弹 |
| 魔弹 | magic_bullet_test | TestMagicBullet_CounterEndsWhenRoundCovered | 全轮覆盖后链条终止 |
| 神之庇护 | god_protection_test | TestGodProtectionMitigatesMoraleLossFromMagicDamage | X选择与士气减免 |
| 全角色覆盖 | full_game_3v3_all_roles_test | TestFullGame3v3_AllRolesCoverage | 3v3自动对局覆盖率>=30% |
| 回归战役 | full_game_regression_campaign_test | TestFullGame3v3_Regression_ActionSkillCoverage | 主动技能覆盖率>=45% |
| 回归战役 | full_game_regression_campaign_test | TestFullGame3v3_Regression_AllSkillCoverage | 全技能覆盖率>=54% |
| 回归战役 | full_game_regression_campaign_test | TestFullGame3v3_Regression_EachRoleRunsSkill | 每角色至少触发1技能 |
| 回归战役 | full_game_regression_campaign_test | TestFullGame3v3_DirectedScenario_EachScenarioHitsTarget | 20个定向场景命中目标 |
| 回归战役 | full_game_regression_campaign_test | TestFullGame3v3_DirectedScenario_TargetSkillCoverage | 目标技能覆盖率100% |
| UI回归 | ui_regression_prompts_test | TestUIRegression_YellowSpring_HidesCounterOptionInResponsePrompt | 黄泉隐藏应战按钮 |
| UI回归 | ui_regression_prompts_test | TestUIRegression_CrimsonBloodBarrier_ResolvesWithoutNestedChoicePrompt | 血气屏障无嵌套弹框 |

---

## 五、前端 Vitest 测试

**目录**：`web/src/`

### 网络层

| 测试文件 | 测试用例 | 验证内容 |
|---|---|---|
| actionRequestAdapter | 攻击载荷构建 | 攻击动作使用UUID和目标 |
| actionRequestAdapter | 响应动作拆分 | 拆分为response_mode和extra_args |
| actionRequestAdapter | 技能多目标 | 保留targets/selections，容忍过期索引 |
| gameplayMessageHandlers | SyncState | 应用到快照store，标记对局开始 |
| gameplayMessageHandlers | RequireAction路由 | 目标为自己→提示；目标为他人→等待 |
| gameplayMessageHandlers | NotifyTimeline | 重放到时间线条目和战斗特效 |
| gameplayTimeline | 事件格式转换 | 时间线事件转游戏事件格式 |
| gameplayTimeline | 错误/提示过滤 | 保留error事件，跳过prompt中断 |
| gameplayTimeline | 类型化重建 | 卡牌揭示/战斗提示/抽牌/行动步骤 |
| messageRouter | 已知命令分发 | 路由到对应处理器 |
| messageRouter | 未知命令 | 传递给回退处理器 |
| messageRouter | 旧版NotifyEvent | 视为未知命令 |
| messageRouter | ChatMessage | 无chat处理器时静默忽略 |
| roomMessageHandlers | assigned | 更新座位/角色/重连令牌 |
| roomMessageHandlers | player_list | 刷新房间玩家和角色 |
| roomMessageHandlers | started | 标记对局开始 |
| roomMessageHandlers | dissolved | 关闭传输、重置状态、展示错误 |
| syncState | SyncState转换 | 结构化载荷→战斗快照格式 |
| wsCommandClient | 断线阻止 | 传输断开时阻止提交 |
| wsCommandClient | 聊天断线 | 断线跳过，重连后发送 |
| wsCommandClient | 焦点特效 | 魔法/技能动作启动焦点 |
| wsCommandClient | 大厅命令 | 包装为RoomAction信封 |
| wsCommandClient | 作弊弃牌 | 发送为Cheat动作载荷 |
| wsConnectionClient | 重连续凭证 | 优先session凭证而非持久化存储 |
| wsConnectionClient | 消息路由 | 路由入站/序列化出站/展示错误 |
| wsConnectionClient | 自动重试 | 意外断开后自动重连 |
| wsConnectionClient | 手动断开 | 关闭socket不安排重连 |
| wsReconnect | 持久化 | 稳定存储键保存和加载 |
| wsReconnect | 格式校验 | 忽略错误/不匹配载荷 |
| wsReconnect | URL构建 | 含重连参数的创建/加入房间URL |

### 组合式函数 (Composables)

| 测试文件 | 测试用例 | 验证内容 |
|---|---|---|
| useBattleInteractionState | 可用技能来源 | 己方回合用后端available_skills |
| useBattleInteractionState | 空列表不回退 | 后端为空时不回退到静态技能 |
| useBattleInteractionState | 天使之墙目标 | target_type=any时可指定友方和敌方 |
| useBattleInteractionState | 祝福牌构建 | 从场Cover卡构建可打祝福牌 |
| useSubmitAction | 未选牌应战 | 报错且不发送请求 |
| useSubmitAction | 过期选择清除 | 越界索引发错误提示重新选择 |
| useSubmitAction | 攻击发送 | 选中牌正确发送到目标 |
| useWebSocket | 完整集成 | 连接→路由消息→更新store→发送信封 |
| useWebSocket | 断开重置 | 手动断开不触发重连 |

### Store

| 测试文件 | 测试用例 | 验证内容 |
|---|---|---|
| battleReview | 去重合并 | 短时间窗口内重复条目合并 |
| battleReview | 士气提示消费 | 消费最新匹配提示 |
| battleReview | 士气爆发排名 | 按损失量降序排列 |
| gameStore | 时间线复用 | 复用timelineStore载荷历史 |
| gameStore | 兼容副作用 | 同步状态和提示选择时保持兼容性 |
| gameStore | 动作模式 | 保留旧辅助行为 |
| matchLifecycle | 新对局清除 | 清除士气回顾和结束遮罩 |
| matchLifecycle | 终局快照 | 从最新状态构建快照并清除瞬态UI |
| matchLifecycle | 延迟同步刷新 | 终局状态同步后刷新快照 |

### 组件

| 测试文件 | 测试用例 | 验证内容 |
|---|---|---|
| HelloWorld | 按钮点击 | 点击后计数递增 |

---

## 统计汇总

| 类别 | 文件数 | 测试函数数 |
|---|---|---|
| 角色技能测试（Go） | ~55 | ~330 |
| 引擎通用机制测试（Go） | ~25 | ~80 |
| 服务器层测试（Go） | 7 | ~25 |
| 集成测试（Go） | 17 | ~30 |
| 模型/规则测试（Go） | 2 | ~7 |
| 前端 Vitest 测试 | 16 | ~45 |
| **合计** | **~122** | **~517** |
