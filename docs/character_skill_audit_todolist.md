# 角色技能一致性核对 TodoList

## 目标

依据 `docs/character_skills_config.md` 的角色顺序，从第一个条目开始，逐个核对每个角色（含基础法术牌）的技能配置与核心代码中的实际生效逻辑是否一致；若不一致，则以配置说明为准改造核心实现，并补齐回归测试。

当前清单总量：`38` 个检查条目，`265` 个技能/子技能节点。

## 执行规则

- [ ] 严格按 `character_skills_config.md` 的出现顺序推进，不跳角色、不跳技能。
- [ ] 每个角色至少完成以下 6 项检查：
  - [ ] 配置元数据：`Type / Timing / PhaseLimit / Target / Cost / Tags / StackingRule`
  - [ ] 主流程时机：`Startup / BeforeAction / ActionSelection / Response / CombatHitCheck / TurnEnd`
  - [ ] 生效链路：伤害、治疗、摸牌、弃牌、状态放置、状态移除、额外行动、形态切换
  - [ ] 分支与限制：强制发动、唯一叠放、锁行动、目标过滤、无法行动、跳过/取消
  - [ ] 跨阶段/跨回合子效果：命中后、未命中后、退场结算、持续状态、自动结算器
  - [ ] 测试回归：补/改针对该角色的回归测试，并跑定向 `go test`
- [ ] 若文档与代码不一致，优先确认文档是否为最新口径；若文档已更新，则按文档改代码。
- [ ] 即使某些角色之前已经做过局部修复，也要从文档重新完整复验，不视为已完成。

## 记录模板

每完成一个角色，在对应小节下补充：

- `结论`: 一致 / 部分不一致 / 大量不一致
- `差异`: 列出与配置不符的技能与原因
- `改动文件`: 记录核心代码与测试文件
- `验证`: 记录本次执行的 `go test` 命令

## 顺序清单

### [x] 00. 基础法术牌 (Common Base Magic Cards)
- `技能`: 圣光 / 中毒 / 虚弱 / 圣盾 / 魔弹
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 公共基础效果优先，后续大量角色与状态结算都会依赖这一层。
- `结论`: 部分不一致（已按配置修复）
- `差异`: `虚弱` 原实现会在行动开始前被旧 `triggerFieldEffects` 直接移除，未进入“跳过行动阶段 / 摸3张牌继续”的选择结算；`中毒` 与 `虚弱` 同时存在时，未保证先完成中毒伤害再进入虚弱选择；`虚弱` 选项顺序与文档配置相反；`中毒/虚弱` 放置触发仍沿用旧 `OnTurnStart` 标记，已改为 `OnBeforeAction`。
- `改动文件`: `internal/engine/game.go` / `internal/engine/magic.go` / `internal/engine/basic_effect_before_action_regression_test.go` / `internal/engine/config_alignment_regression_test.go` / `internal/engine/crk_hom_skill_regression_test.go`
- `验证`: `go test ./internal/engine -run 'Test(BuffResolve_PoisonResolvesBeforeWeaknessChoice|WeaknessPrompt_OrderMatchesConfig|WeaknessChoiceMappingMatchesConfig|PerformMagic_PoisonCannotStackOnSameTarget|PendingDamage_PoisonDoesNotConsumeHolyShield|CombatDefend_|MagicBullet_|CombatShield_|CrimsonKnightFaith_SelfPoisonCanUseHeal|AngelCleanse_)'`；`go test ./internal/engine -run 'Test(FiveElementsBind_|ElementalSeal_|AngelCleanse_|ConfigAlignment|BasicEffect|CrimsonKnightFaith_SelfPoisonCanUseHeal)'`

### [x] 01. 天使 (Angel)
- `技能`: 天使羁绊 / 天使祝福 / 风之洁净 / 天使之歌 / 神之庇护
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 按文档逐项复验。
- `结论`: 部分不一致（已按配置修复）
- `差异`: `天使之歌` 原元数据仍按 `Startup` 技能建模，未严格落在“回合开始前响应”链路；且移除基础效果后会回切到错误阶段，无法稳定续接回合开始流程。`神之庇护` 原实现会按可用水晶自动全额抵御士气下降，没有让玩家按文档选择 `X` 值进行部分抵御。其余 `天使羁绊 / 天使祝福 / 风之洁净` 与文档口径一致。
- `改动文件`: `internal/data/characters.go` / `internal/engine/game.go` / `internal/engine/skills/handlers_impl.go` / `internal/engine/angel_config_regression_test.go`
- `验证`: `go test ./internal/engine -run 'Test(AngelSong_|GodProtection_|AngelBond_|AngelCleanse_|CrystalSubstitute_|BasicEffect|BuffResolve_PoisonResolvesBeforeWeaknessChoice|WeaknessPrompt_OrderMatchesConfig|WeaknessChoiceMappingMatchesConfig)'`；`go test ./internal/engine -run 'Test(Angel|StartupSkill|StartupSkip|CrystalSubstitute_|MoonGoddess_|HolyBow_|Arbiter_)'`

### [x] 02. 狂战士 (Berserker)
- `技能`: 狂化 / 撕裂 / 血腥咆哮 / 血影狂刀
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 按文档逐项复验。
- `结论`: 部分不一致（已按配置修复）
- `差异`: `撕裂` 原实现错误限制为仅主动攻击命中后可发动，未覆盖应战攻击命中；`狂化` 原实现将“攻击声明时 +1”和“命中且手牌>3 再 +1”折叠在同一伤害修正钩子里，行为近似但时机建模不符合文档；`血影狂刀` 原实现挂在统一伤害修正阶段，未严格落在“主动独有攻击命中后”结算。`血腥咆哮` 本次复验未发现与当前文档口径冲突的实现差异。
- `改动文件`: `internal/data/characters.go` / `internal/engine/skills/handlers_impl.go` / `internal/engine/berserker_config_regression_test.go`
- `验证`: `go test ./internal/engine -run 'Test(Berserker|BloodRoar|BloodBlade|BerserkerTear|BerserkerAttackSealer|RoleSkillBugfix_)'`

### [x] 03. 封印师 (Sealer)
- `技能`: 法术激荡 / 封印破碎 / 五系束缚 / 水/火/地/风/雷之封印
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 已有部分修订记录，本次按最新文档全量复验。
- `结论`: 一致
- `差异`: 本次复验未发现与当前文档口径冲突的实现差异。`五系束缚` 已走独立通用状态结算器，正确在 `OnBeforeAction` 提示“摸 2+X 张牌取消 / 跳过行动阶段”，且 `X` 按全场封印数封顶 2；`水/火/地/风/雷之封印` 已统一走共享元素封印结算器，按放置时绑定元素在“打出或展示对应系别牌”时造成 3 点法术伤害并移除；`封印破碎` 已支持按文档从全场基础效果中选择并回收入手。
- `改动文件`: 无新增改动（沿用现有实现与回归）
- `验证`: `go test ./internal/engine -run 'Test(FiveElementsBind_|ElementalSeal_|SealBreak_|BerserkerAttackSealer|RoleSkillBugfix_SealBreak)'`

### [x] 04. 风之剑圣 (Wind Sword Saint)
- `技能`: 风怒追击 / 圣剑 / 剑影 / 疾风技 / 列风技
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 按文档逐项复验。
- `结论`: 部分不一致（已按配置修复）
- `差异`: `列风技` 在代码与牌库映射中仍沿用旧口径 `烈风技`，导致技能标题、独有技匹配名、测试数据与配置文档分叉；`风怒追击` 原实现额外要求“手牌里必须仍有风系攻击牌”才可发动，这一前置并不在文档配置中，且主流程本身已支持在没有合法额外行动时通过“无法行动”安全跳过。
- `改动文件`: `internal/data/characters.go` / `internal/rules/deck.go` / `internal/engine/skills/handlers_impl.go` / `internal/engine/game.go` / `internal/engine/blademaster_response_regression_test.go` / `internal/engine/exclusive_skill_card_regression_test.go` / `docs/card.md`
- `验证`: `go test ./internal/engine -run 'Test(BladeMaster_|ExclusiveSkillCard_)'`

### [x] 05. 神箭手 (Sharpshooter)
- `技能`: 贯穿射击 / 闪电箭 / 狙击 / 精准射击 / 闪光陷阱
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 按文档逐项复验。
- `结论`: 部分不一致（已按配置修复）
- `差异`: `精准射击` 原实现在设置“强制命中”之外，还额外把本次攻击标记为 `CanBeResponded=false`；而当前文档仅要求 `ForceHit + 伤害-1`，不额外声明独立的 `Unrespondable` 标记。主流程本身已会因 `ForceHit` 直接跳过响应阶段，因此这里属于多余且偏重的副作用，已移除。其余 `贯穿射击 / 闪电箭 / 狙击 / 闪光陷阱` 本次复验未发现与当前文档冲突的实现差异。
- `改动文件`: `internal/engine/skills/handlers_impl.go`
- `验证`: `go test ./internal/engine -run 'Test(Archer_|CounterAttackActionGating_)'`

### [x] 06. 暗杀者 (Assassin)
- `技能`: 反噬 / 水影 / 潜行
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 已有部分修订记录，本次按最新文档全量复验。
- `结论`: 一致
- `差异`: 本次复验未发现与当前文档口径冲突的实现差异。`反噬` 仅在承受攻击伤害后让攻击者强制摸 1 张牌，不再误判法术伤害；`水影` 会在摸牌前响应并正确恢复被中断的摸牌/受伤流程；`潜行` 的“可选摸 1 后进入形态”“主动攻击不可应战且按剩余能量增伤”“下个行动阶段开始时退场”链路均与文档一致。
- `改动文件`: 无新增改动（沿用现有实现与回归）
- `验证`: `go test ./internal/engine -run 'Test(Assassin_|AssassinWaterShadow|CombatMagicRoleFix_|StartupSkip_)'`

### [x] 07. 圣女 (Saintess)
- `技能`: 冰霜祷言 / 治愈之光 / 治疗术 / 圣疗 / 怜悯
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 按文档逐项复验。
- `结论`: 部分不一致（已按配置修复）
- `差异`: `圣疗` 原实现只支持固定分配治疗，并且始终只给予额外攻击行动，缺少文档要求的“1~3 名目标分配共 3 点治疗”与“额外攻击行动 / 额外法术行动二选一”完整交互；`怜悯` 原实现把 +1 水晶错误加到了阵营资源，并未建立持续态标记，导致“横置后手牌上限恒定为 7”的持续效果与后续启动阶段禁重复触发都不完整。其余 `冰霜祷言 / 治愈之光 / 治疗术` 本次复验未发现与当前文档冲突的实现差异。
- `改动文件`: `internal/engine/skill_use_policy.go` / `internal/engine/skills/handlers_impl.go` / `internal/engine/game.go` / `internal/engine/saintess_config_regression_test.go`
- `验证`: `go test ./internal/engine -run 'Test(Saintess_|CombatMagicRoleFix_)'`

### [x] 08. 魔法少女 (Magical Girl)
- `技能`: 魔弹掌控 / 魔弹融合 / 魔爆冲击 / 毁灭风暴
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 按文档逐项复验。
- `结论`: 部分不一致（已按配置修复）
- `差异`: `魔弹融合` 原实现仍保留为旧式主动技，并要求弃置火/地牌后直接进入魔弹方向选择，和文档要求的“使用火/地法术时可选择改写为魔弹链”不一致；`魔弹掌控` handler 仍停留在未完成占位，方向选择实际由主流程接管；`魔爆冲击` 原实现把文档内的“逐目标弃法术牌，否则受2点法术伤害且你弃1张牌”做成了目标全部处理后的可选一次性弃牌，且目标展示弃置时没有严格按提示选项映射；`毁灭风暴` 原实现除统一资源扣费外又额外手动扣了1次宝石，且目标下限未收紧到2名敌方。现已统一按文档口径修正。 
- `改动文件`: `internal/data/characters.go` / `internal/engine/skill_use_policy.go` / `internal/engine/skills/handlers_impl.go` / `internal/engine/game.go` / `internal/engine/combat_magic_role_fix_regression_test.go` / `internal/engine/magical_girl_config_regression_test.go` / `tests/magicalgirl_skills_test.go`
- `验证`: `go test ./internal/engine -run 'Test(MagicalGirl_|SaintessFrostPrayer_|MagicalGirlMagicBulletFusion_)'` / `go test ./tests -run 'TestMagicalGirl_Skills'`

### [x] 09. 女武神 (Valkyrie)
- `技能`: 神圣追击 / 秩序之印 / 和平行者 / 军威神光 / 英灵召唤
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 按文档逐项复验。
- `结论`: 部分不一致（已按配置修复）
- `差异`: `军威神光` 的配置标题和运行时提示仍沿用旧名“军神威光”；`英灵召唤` 原实现会在额外弃1张法术牌后错误地允许选择任意角色获得+1治疗，而文档要求的是“当前战斗目标 +1 治疗”；同一技能原实现还会在应战命中等非自己回合场景下直接进入英灵形态，但文档要求只有“你的回合内发动英灵召唤后”才进入形态；`军威神光` 的第二分支原实现会错误脱离英灵形态，而文档只有第一分支会转正退出。现已统一按文档口径修正，其余 `神圣追击 / 秩序之印 / 和平行者` 本轮复验未发现新的文档冲突。 
- `改动文件`: `internal/data/characters.go` / `internal/engine/skills/handlers_new_roles.go` / `internal/engine/game.go` / `internal/engine/valkyrie_combo_regression_test.go` / `internal/engine/valkyrie_config_regression_test.go`
- `验证`: `go test ./internal/engine -run 'Test(Valkyrie_|CounterHit_PhaseEndSkillsNotTriggeredForCounterAction)'`

### [x] 10. 元素师 (Elementalist)
- `技能`: 元素吸收 / 元素点燃 / 雷击 / 冰冻 / 风刃 / 陨石 / 火球 / 月光
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 按文档逐项复验。
- `结论`: 部分不一致（已按配置修复）
- `差异`: `元素点燃 / 雷击 / 风刃 / 陨石 / 火球 / 月光` 的配置目标域原先仍是“任意角色”，与文档要求的“敌方目标”不一致；`冰冻` 原实现虽然要求2名目标，但缺少“第1个目标必须是敌方伤害目标”的后端硬校验；另外 `元素吸收` 的技能类型元数据仍停留在被动技，`月光` 缺少文档对应的 `Ultimate` 标记。现已统一按文档口径修正；其余元素师技能的核心结算链本轮复验未发现新的文档冲突。 
- `改动文件`: `internal/data/characters.go` / `internal/engine/skill_use_policy.go` / `internal/engine/elementalist_regression_test.go`
- `验证`: `go test ./internal/engine -run 'TestElementalist'`

### [x] 11. 仲裁者 (Arbiter)
- `技能`: 仲裁法则 / 仪式中断 / 末日审判 / 审判浪潮 / 仲裁仪式 / 判决天平
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 已按最新文档全量复验。
- `结论`: 部分不一致（已按配置修复）
- `差异`: `仲裁法则` 原实现仍依赖 `TurnStart` 被动触发，和文档要求的“游戏初始时结算”不一致，且在旧状态/测试重建 token 映射时会被误触发；现已改为角色初始化直接结算，并移除常规回合触发入口。`仲裁仪式` 原实现会在启动时立即+1审判，且确认启动技后又会被主流程的“审判形态回合开始+1”偷跑一次，同样不符合文档“进入形态后，后续自己回合开始才+1审判”的口径；现已拆分为纯形态进入 + 独立的后续回合开始结算。`仲裁仪式 / 判决天平` 原配置缺少 `Ultimate` 标记；`仪式中断` 的运行时效果与文案没有严格体现“转正脱离审判形态并为战绩区+1宝石”。`判决天平` 本轮复验同时补测了“先+1审判再进分支”“弃光全部手牌”“补到手牌上限并+1宝石”的文档链路。
- `改动文件`: `internal/data/characters.go` / `internal/engine/game.go` / `internal/engine/skills/handlers_new_roles.go` / `internal/engine/arbiter_law_regression_test.go`
- `验证`: `go test ./internal/engine -run 'TestArbiter'`；`go test ./internal/engine -run 'Test(StartupSkill|CrystalSubstitute_ArbiterBalance|Arbiter)'`

### [x] 12. 冒险家 (Adventurer)
- `技能`: 欺诈 / 强运 / 地下法则 / 冒险者天堂 / 偷天换日
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 按文档逐项复验。
- `结论`: 部分不一致（已按配置修复）
- `差异`: `欺诈` 的 2 弃分支原实现把“可选除暗系外任意系攻击”误收窄成了“仅五基础系，不含光/暗”，与文档不符；现已恢复光系可选，仅继续排除暗系。`强运` 原实现挂在“欺诈虚拟攻击的 AttackStart”上蹭触发，语义上仍与主流程耦合；现已改为在 `欺诈` 技能完成时直接结算，贴合文档的“当你发动欺诈时”。`地下法则` 原实现是在默认购买结算完成后，于 `PhaseEnd` 再额外给我方战绩区+2宝石，导致错误保留了“摸3牌、战绩区+1宝石+1水晶”的默认购买收益；现已改为真正的购买改写。`冒险者天堂` 原实现依赖 `PhaseEnd` 自动响应链路；现已改成直接挂接在提炼结果分配流程上，提炼结果会在该链路中被强制/可选地转移给队友，更贴近文档的“提炼时”结算。`偷天换日` 原配置缺少 `Ultimate` 标记，本轮已补齐。
- `改动文件`: `internal/data/characters.go` / `internal/engine/game.go` / `internal/engine/adventurer_fraud_regression_test.go` / `internal/engine/adventurer_priest_rules_regression_test.go`
- `验证`: `go test ./internal/engine -run 'TestAdventurer'`；`go test ./internal/engine -run 'Test(Adventurer|PriestDivineDomain)'`

### [x] 13. 圣枪骑士 (Holy Lancer)
- `技能`: 神圣启示 / 辉耀 / 惩戒 / 圣击 / 天枪 / 地枪 / 圣光祈愈
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 按文档逐项复验。
- `结论`: 部分不一致（已按配置修复）
- `差异`: `神圣启示` 原实现主要靠角色被动在回合开始时刷新，未覆盖“游戏初始即生效 / 星杯变化后动态同步”的文档口径；现已改为常态同步治疗上限规则。`辉耀 / 惩戒` 原实现没有走技能主流程的弃牌成本，而是由 handler 直接吞掉手里第一张符合条件的牌，导致无法严格按玩家选择的弃牌结算；现已补齐配置成本并改回主流程处理。`圣光祈愈` 原 handler 还会在统一费用扣除之外再次手动扣 1 宝石，形成重复扣费；同时配置缺少 `Ultimate` 标记和“本回合禁用天枪”的完整文案，现已一并修正。其余 `圣击 / 天枪 / 地枪` 本轮复验未发现新的文档冲突。
- `改动文件`: `internal/data/characters.go` / `internal/engine/game.go` / `internal/engine/skills/handlers_new_roles.go` / `internal/engine/skill_use_policy.go` / `internal/engine/holy_lancer_earth_spear_regression_test.go`
- `验证`: `go test ./internal/engine -run 'Test(HolyLancer|ResponsePrompt_PrunesInvalidHolyLancerSkySpear)'`；`go test ./internal/engine -run 'Test(CrystalSubstitute_GemCostCannotUseCrystal|HolyLancer)'`

### [x] 14. 精灵射手 (Elf Archer)
- `技能`: 元素射击 / 元素射击·命中后结算 / 元素射击·行动后结算 / 动物伙伴 / 精灵密仪 / 精灵密仪·形态退场结算 / 宠物强化
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 按文档逐项复验。
- `结论`: 部分不一致（已按配置修复）
- `差异`: `元素射击` 原实现把 `水/地之矢` 做成了“命中后可选任意角色”的额外选目标分支，与文档要求的“当前战斗目标”不一致；`风之矢` 也被提前在发动时直接塞入额外攻击行动，而不是在“攻击行动结束后”结算，且 `火/水/地` 在攻击未命中时缺少统一的战斗结束清理。现已改为水/地自动作用于当前攻击目标，风之矢在行动结束时统一发放额外攻击，并在同一处清理本次战斗残留标记。`动物伙伴 / 宠物强化` 原实现绕开了通用响应技能窗口，改成了自定义 choice 链，导致不再符合文档里的“同一响应窗二选一、宠物强化替换动物伙伴”；现已收回统一 `ResponseSkill` 流程，并把 `宠物强化` 的目标严格收窄为“本次受伤敌方目标”。`精灵密仪` 原配置缺少 `Ultimate` 标记，退场分支也错误允许选择任意角色承受 2 点法术伤害；现已补齐标签并改成仅能选择敌方角色。
- `改动文件`: `internal/data/characters.go` / `internal/model/skill.go` / `internal/engine/new_roles_helpers.go` / `internal/engine/game.go` / `internal/engine/skill_dispatcher.go` / `internal/engine/skills/handlers_new_roles.go` / `internal/engine/elf_archer_skill_regression_test.go` / `internal/engine/elf_holy_lancer_bugfix_regression_test.go`
- `验证`: `go test ./internal/engine -run 'TestElf'`；`go test ./internal/engine -run 'Test(HolyLancer|Elf)'`；`go test ./internal/engine`

### [x] 15. 瘟疫法师 (Plague Mage)
- `技能`: 不朽 / 圣渎 / 瘟疫 / 瘟疫·回合结束奖励 / 死亡之触 / 剧毒新星
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 按文档逐项复验。
- `结论`: 部分不一致（已按配置修复）
- `差异`: `瘟疫` 原实现仍保留了“结算后立即自己 +1 治疗”，与文档要求的“仅当本回合该技能造成过士气下降时，回合结束才 +1 治疗”不一致；现已改为通过伤害溢出导致的士气下降链路追踪本回合奖励，并在自己的回合结束时统一结算。`死亡之触` 原配置和结算链都没有严格体现“敌方目标”，主流程里实际允许把最后一步目标选成任意角色；现已改为技能发动时就按 `Enemy` 目标域锁定目标，再进入 `X/Y/同系弃牌` 交互，并保持“本次不触发不朽”。`剧毒新星` 原 handler 会在统一技能费用之外再次手动扣 1 红宝石，形成重复扣费；现已移除额外扣费并补上 `Ultimate` 标记。`不朽` 的触发前置也收窄为“自己的法术行动结束时”，不再依赖 `圣光/魔弹` 名称硬编码。
- `改动文件`: `internal/data/characters.go` / `internal/model/types.go` / `internal/engine/combat.go` / `internal/engine/game.go` / `internal/engine/skills/handlers_new_roles.go` / `internal/engine/plague_mage_skill_regression_test.go`
- `验证`: `go test ./internal/engine -run 'TestPlague'`；`go test ./internal/engine`

### [x] 16. 魔剑士 (Magic Swordsman)
- `技能`: 修罗连斩 / 暗影凝聚 / 暗影凝聚·形态退场结算 / 暗影之力 / 暗影抗拒 / 暗影流星 / 黄泉震颤 / 黄泉震颤·命中后结算
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 按文档逐项复验。
- `结论`: 部分不一致（已按配置修复）
- `差异`: `修罗连斩` 原实现仍额外要求“手牌中已有火系攻击牌”才给额外攻击，不符合文档的纯行动追加口径；现已移除该前置，让额外火攻行动在没有合法牌时由统一“无法行动”兜底。`暗影凝聚` 原实现会在下一回合 `Startup` 阶段过早退场，早于文档要求的“下个行动阶段开始前”；现已改为进入行动阶段时再统一退场。`暗影流星` 原实现仍走旧式“发动后再选弃牌/再选目标”的主流程分支，目标域也放宽成任意角色；现已收回统一 `UseSkill(target + discard)` 动作协议，并把目标严格收窄为敌方角色。`黄泉震颤` 原实现仍把本次攻击强制改写成 `暗灭`，与文档当前口径不符；现已改为仅提供“不可应战 + 命中后补牌弃2”，并补齐 `Gem / Ultimate / CostGem` 配置与消耗。
- `改动文件`: `internal/data/characters.go` / `internal/engine/game.go` / `internal/engine/new_roles_helpers.go` / `internal/engine/skill_use.go` / `internal/engine/skills/handlers_new_roles.go` / `internal/engine/magic_swordsman_prayer_css_bugfix_regression_test.go` / `internal/engine/magic_swordsman_config_regression_test.go`
- `验证`: `go test ./internal/engine -run 'TestMagicSwordsman'`；`go test ./internal/engine`

### [x] 17. 血色剑灵 (Blood Sword Spirit)
- `技能`: 血色荆棘 / 赤色一闪 / 血染蔷薇 / 血气屏障 / 血蔷薇庭院 / 血蔷薇庭院·全场禁疗抵伤 / 散华轮舞
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 按文档逐项复验。
- `结论`: 部分不一致（已按配置修复）
- `差异`: `血色荆棘` 原实现把“攻击命中”错误收窄成了“仅主动攻击命中”，导致应战攻击命中不会获得鲜血；现已按文档恢复为所有攻击命中均可触发。`血染蔷薇` 原实现仍是单目标结算，并把“水晶翻宝石”固定做在自己身上，与文档要求的“恰好1敌方 + 1我方双目标”不一致；现已收回统一 `UseSkill` 双目标协议，并补上“不可重复选同一角色 / 必须一敌一我”的目标校验。`血气屏障` 原实现被做成了“减伤后再二次确认、再选任意敌人”的旧 choice 链，而文档要求的是直接反打伤害来源；现已改为自动对来源角色造成1点法术伤害，并退役相应主流程分支。`散华轮舞` 原配置缺少 `Ultimate` 标记，本轮已补齐。
- `改动文件`: `internal/data/characters.go` / `internal/engine/skill_use_policy.go` / `internal/engine/skills/handlers_new_roles.go` / `internal/engine/game.go` / `internal/engine/counter_attack_action_gating_regression_test.go` / `internal/engine/magic_swordsman_prayer_css_bugfix_regression_test.go` / `internal/engine/crimson_sword_spirit_config_regression_test.go` / `tests/ui_regression_prompts_test.go`
- `验证`: `go test ./internal/engine -run 'Test(Crimson|CounterHit_CrimsonBloodThornsAlsoTriggersOnCounterHit)'`；`go test ./internal/engine`；`go test ./tests -run 'TestUIRegression_(YellowSpring|CrimsonBloodBarrier_)'`

### [x] 18. 祈祷师 (Prayer Master)
- `技能`: 光辉信仰 / 黑暗诅咒 / 威力赐福 / 威力赐福·触发 / 迅捷赐福 / 迅捷赐福·触发 / 祈祷 / 祈祷·攻击增符 / 法力潮汐
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 按文档逐项复验。
- `结论`: 部分不一致（已按配置修复）
- `差异`: `祈祷` 原实现虽然会进入祈祷形态，但没有真正把“手牌上限固定为5”落到引擎固定上限规则里；现已补齐启动结算与 `GetMaxHand` 固定上限判定。`光辉信仰` 原 handler 在目标缺失或不合法时会偷偷回落到自己/同阵营默认目标，不符合文档要求的“明确指定其他队友”；现已改为严格要求有效队友目标。`威力赐福 / 迅捷赐福` 原配置把专属牌错误建模成“弃1张手牌”，导致统一 `UseSkill` 会去手牌区找专属牌，而不是直接消费专属卡区；现已改回“直接消耗专属卡区并放置场上”的统一协议。`法力潮汐` 与 `祈祷` 的配置元数据也补齐了文档对应的 `Gem/Crystal/Ultimate` 标记与基础费用展示。
- `改动文件`: `internal/data/characters.go` / `internal/engine/skills/handlers_roles_18_22.go` / `internal/engine/new_roles_helpers.go` / `internal/engine/game.go` / `internal/engine/prayer_form_persist_regression_test.go` / `internal/engine/prayer_master_config_regression_test.go`
- `验证`: `go test ./internal/engine -run 'TestPrayer'`；`go test ./internal/engine`

### [x] 19. 红莲骑士 (Crimson Knight)
- `技能`: 腥红圣约 / 腥红信仰 / 血腥祷言 / 杀戮盛宴 / 热血沸腾 / 热血沸腾·形态免士气 / 热血沸腾·回合结束退场 / 戒骄戒躁 / 腥红十字
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 按文档逐项复验。
- `结论`: 部分不一致（已按配置修复）
- `差异`: `戒骄戒躁` 原实现仍走 `game.go` 的额外行动类型二选一 prompt，与文档要求的“额外行动类型跟随刚结束的行动类型自动确定”不一致；现已改为在 handler 内直接按 `ActionType` 追加同类型额外行动，并删除旧 choice 分支。`腥红十字` 原配置与文案仍允许指定“任意角色”，不符合文档的敌方目标域；现已收紧为 `TargetEnemy`，同时补齐 `Crystal/Ultimate` 元数据，并在 handler 侧增加敌我校验。`血腥祷言` 现有多步交互流虽仍保留，但已复验其“仅其他队友”“1~2名目标”“X=1 时直接单目标”“分配总量等于 X”的文档边界，并补回归测试锁定行为。
- `改动文件`: `internal/data/characters.go` / `internal/engine/skills/handlers_roles_18_22.go` / `internal/engine/game.go` / `internal/engine/crk_hom_skill_regression_test.go` / `internal/engine/crimson_knight_bloody_prayer_regression_test.go` / `internal/engine/crimson_knight_killing_feast_regression_test.go`
- `验证`: `go test ./internal/engine -run 'TestCrimsonKnight(CalmMind|BloodyPrayer|CrimsonCross|KillingFeast|HotBlood)'`；`go test ./internal/engine`

### [x] 20. 英灵人形 (Heroic Puppet)
- `技能`: 战纹掌控 / 怒火压制 / 战纹碎击 / 魔纹融合 / 符文改造 / 符文改造·形态退场结算 / 双重回响
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 按文档逐项复验。
- `结论`: 部分不一致（已按配置修复）
- `差异`: `怒火压制` 与 `魔纹融合` 原实现虽然都会在攻击未命中时进入响应列表，但缺少文档要求的“二选一”互斥约束；现已补到响应互斥判定中，选择其一后不会继续保留另一个。`符文改造` 原配置缺少文档对应的 `Gem/Ultimate` 元数据与基础宝石费用；现已补齐。`双重回响` 原配置缺少 `Crystal/Ultimate`、目标域与费用元数据，且旧 handler 错把“另一个目标角色”实现成“不能选择自己”，同时把“不造成士气下降”实现成“摸牌上限截断”；现已改为允许选择除本次受伤目标外的任意角色，并统一使用 `magic_no_morale` 伤害语义来保留完整伤害摸牌但不扣士气。
- `改动文件`: `internal/data/characters.go` / `internal/engine/skills/handlers_roles_18_22.go` / `internal/engine/skill_dispatcher.go` / `internal/engine/game.go` / `internal/engine/crk_hom_skill_regression_test.go`
- `验证`: `go test ./internal/engine -run 'TestHom|TestHomunculus|TestCounterMiss_DoesNotTriggerActiveOnlyOnAttackMissSkills'`；`go test ./internal/engine`

### [x] 21. 神官 (Priest)
- `技能`: 神圣启示 / 神圣祈福 / 水之神力 / 圣使守护 / 神圣契约 / 神圣领域
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 已有部分修订记录，现已按最新文档全量复验。
- `结论`: 部分不一致（已按配置修复）
- `差异`: `水之神力` 原实现把“弃水系牌 + 交1张牌给队友”放宽成“手牌只剩1张水牌时也能发动并跳过交牌”，不符合文档要求的“至少保留1张可转移手牌并强制交牌”；现已改回必须手牌数≥2，且第二张牌为必选转移牌。`神圣领域` 原实现允许“手牌不足2时弃全部”并继续发动，且伤害分支目标列表会包含自己；这两点都不符合文档口径，现已改为必须严格弃2张牌，且分支①只允许选择其他角色。另已补齐 `神圣契约 / 神圣领域` 的 `Ultimate` 元数据，以及 `神圣领域` 的基础水晶费用配置，避免协议展示与实际扣费口径不一致。
- `改动文件`: `internal/data/characters.go` / `internal/engine/skills/handlers_roles_18_22.go` / `internal/engine/skill_use_policy.go` / `internal/engine/game.go` / `internal/engine/adventurer_priest_rules_regression_test.go`
- `验证`: `go test ./internal/engine -run 'TestPriest|TestAdventurer.*Priest|TestAdventurerStealSky_ModeAndExtraActionChoice|TestAdventurerUndergroundLaw_RewritesBuyInsteadOfDefaultSettlement|TestAdventurerExtractFullEnergy_ForceParadiseTransfer|TestAdventurerParadise_'`；`go test ./internal/engine`

### [x] 22. 阴阳师 (Onmyoji)
- `技能`: 式神降临 / 阴阳转换 / 式神转换 / 黑暗祭礼 / 式神咒束 / 生命结界 / 生命结界·三鬼火免士气
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 已按最新文档全量复验，并把战斗响应编排从 `game.go` 主文件收回到独立 Onmyoji flow。
- `结论`: 部分不一致（已按配置修复）
- `差异`: `黑暗祭礼` 原实现会在回合结束时把全场角色都塞进目标池，和文档 `Enemy` 目标域不一致；现已收窄为仅敌方角色，并同步修正角色配置元数据。`生命结界` 原配置缺少 `Ultimate` 标记，且目标元数据仍停留在 `TargetNone`；现已补齐为 `Ultimate + Crystal + TeamOther`，并支持沿统一 `UseSkill` 协议预选队友目标。`生命结界` 原流程即使前端先选了目标，后端仍只能在中断里重复选目标；现已补上 `locked_target_id` 分支，预选目标时可直接跳过第二次目标确认。`阴阳转换 / 式神转换 / 式神咒束` 原先仍直接耦在 `game.go` 主战斗流程中，且普通应战与代应战各自复制了一份鬼火/退形态/改伤逻辑；现已抽到独立 Onmyoji flow，并统一复用同一套“同命格应战增益”结算 helper，避免后续继续分叉。
- `改动文件`: `internal/data/characters.go` / `internal/engine/game.go` / `internal/engine/skill_flow_onmyoji.go` / `internal/engine/skills/handlers_roles_18_22.go` / `internal/engine/skill_use_policy.go` / `internal/engine/onmyoji_skill_flow_regression_test.go`
- `验证`: `go test ./internal/engine -run 'TestOnmyoji'`；`go test ./internal/engine`

### [x] 23. 苍炎魔女 (Azure Flame Witch)
- `技能`: 苍炎法典 / 天火断空 / 魔女之怒 / 魔女之怒·烈焰改写 / 魔女之怒·形态退场结算 / 替身玩偶 / 永生银时计 / 痛苦链接 / 魔能反转
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 已按最新文档全量复验。
- `结论`: 部分不一致（已按配置修复）
- `差异`: `苍炎法典 / 天火断空` 的目标域虽然建成了 `Any`，但后端原先没有落实文档里的 `Target != Self` 过滤，理论上允许对自己发动；现已在统一 `skill_use_policy` 中补上“不能以自己为目标”的硬校验。`替身玩偶` 原配置缺少“弃1张法术牌”与“选择其他队友”元数据；`魔能反转` 原配置缺少 `CostCrystal: 1`，且目标元数据未声明为 `Enemy`。这些都已按文档补齐，避免协议展示与实际响应链路继续分叉。其余 `魔女之怒 / 烈焰改写 / 形态退场 / 永生银时计 / 痛苦链接` 本轮复验未发现新的文档冲突。
- `改动文件`: `internal/data/characters.go` / `internal/engine/skill_use_policy.go` / `internal/engine/blaze_witch_skill_regression_test.go`
- `验证`: `go test ./internal/engine -run 'TestBlazeWitch'`；`go test ./internal/engine`

### [x] 24. 贤者 (Sage)
- `技能`: 智慧法典 / 法术反弹 / 魔道法典 / 圣洁法典
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 已按最新文档全量复验。
- `结论`: 部分不一致（已按配置修复）
- `差异`: `智慧法典` 原实现把“弃1张牌”做成了可选确认，和文档的强制弃牌不一致；现已改为在承受 `>3` 点法术伤害后直接进入强制弃牌中断，并按伤害结算阶段恢复主流程。`法术反弹 / 魔道法典` 原实现会把自己也塞进目标池，允许对自己选目标；现已按文档收窄为“全场除自己”，并在最终选择时追加硬校验，防止后续流程重新放开。`圣洁法典` 原实现允许选择 `0` 名角色并只结算自伤，和文档 `MinCount: 1` 不一致；现已关闭 `0` 目标入口，目标数量选择改为仅允许 `1..X-2`。同时补齐了贤者技能配置元数据，并放通 `UseSkill` 对 `魔道法典 / 圣洁法典` 的“先起技能、后续流程再选目标”校验，避免元数据与中断式目标选择继续打架。
- `改动文件`: `internal/data/characters.go` / `internal/engine/new_roles_helpers.go` / `internal/engine/game.go` / `internal/engine/skill_use_policy.go` / `internal/engine/sage_skill_regression_test.go`
- `验证`: `go test ./internal/engine -run 'TestSage'`；`go test ./internal/engine`

### [x] 25. 魔弓 (Magic Bow)
- `技能`: 魔贯冲击 / 魔贯冲击·命中追加 / 魔贯冲击·未命中追伤 / 雷光散射 / 多重射击 / 多重射击·伤害修正 / 充能 / 魔眼
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 已按最新文档全量复验。
- `结论`: 部分不一致（已按配置修复）
- `差异`: `魔贯冲击·命中追加` 原实现会在命中后再次弹“是否发动”的确认框，而文档要求该分支为自动结算；现已改为命中后若仍有火系充能则强制再移除1个并自动为本次攻击伤害 `+1`。`魔眼` 原实现把“目标弃1张牌 / 你摸3张牌”做成了由施法者手动选分支，且目标池包含自己、目标无手牌时也不会回落到摸3；这些都和文档不一致。现已改成必须先指定其他角色，目标有手牌时由其强制弃1张，无法弃置时改为施法者摸3张，然后再选择1张手牌作为充能并获得1点蓝水晶。与此同时，补齐了 `雷光散射 / 魔眼` 的目标元数据，并让 `UseSkill` 与旧启动技 handler 在魔弓上统一兼容“公共扣费 + 后续中断选目标”的路径，避免统一协议下出现重复扣资源。
- `改动文件`: `internal/data/characters.go` / `internal/engine/skill_use_policy.go` / `internal/engine/skills/handlers_magic_bow.go` / `internal/engine/new_roles_helpers.go` / `internal/engine/game.go` / `internal/engine/magic_bow_regression_test.go`
- `验证`: `go test ./internal/engine -run 'TestMagicBow'`；`go test ./internal/engine`

### [x] 26. 魔枪 (Magic Spear)
- `技能`: 暗之解放 / 暗之解放·下次主动攻击增伤 / 幻影星尘 / 黑暗束缚 / 暗之障壁 / 充盈 / 充盈·下次主动攻击增伤 / 漆黑之枪
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 已按最新文档全量复验。
- `结论`: 部分不一致（已按配置修复）
- `差异`: `幻影星尘` 原配置仍是 `TargetNone`，且自伤后续的二段法伤目标池会把全场都塞进去，和文档要求的 `Enemy(1)` 不一致；现已把元数据、统一目标校验与后续选择池一起收紧到“仅敌方1名角色”，并支持 `UseSkill` 预选敌方目标后直接跳过第二次提示。`充盈` 原实现把发动者自己也放进“所有人各弃1”链路里，等于把文档里“发动者自身由费用弃牌覆盖、额外可选弃牌仅面向队友”的约束打破了；现已改成“敌方强制各弃1 + 预选队友可选弃1”，并补齐 `TeamOther(0..1)` 元数据。其余 `暗之解放 / 暗之解放·下次主动攻击增伤 / 黑暗束缚 / 暗之障壁 / 充盈·下次主动攻击增伤 / 漆黑之枪` 本轮复验未发现新的文档冲突。
- `改动文件`: `internal/data/characters.go` / `internal/engine/skill_use_policy.go` / `internal/engine/skills/handlers_magic_lancer.go` / `internal/engine/new_roles_helpers.go` / `internal/engine/game.go` / `internal/engine/magic_lancer_regression_test.go`
- `验证`: `go test ./internal/engine -run 'TestMagicLancer'`；`go test ./internal/engine`

### [x] 27. 灵符师 (Talisman Master)
- `技能`: 灵符-雷鸣 / 灵符-风行 / 念咒 / 百鬼夜行 / 灵力崩解 / 灵力崩解·伤害增幅
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 已按最新文档全量复验。
- `结论`: 部分不一致（已按配置修复）
- `差异`: `念咒` 原实现和配置里都残留了“妖力上限2”的旧规则，会在已有2个妖力时直接跳过盖放，但文档当前版本并没有这个上限；现已移除上限拦截，让每次发动灵符且仍有手牌时都可以继续盖放1张手牌为妖力。`百鬼夜行` 的火妖力展示分支原来只有日志，没有真正发出公开展示事件，和文档“可展示之[展示]”不一致；现已在玩家选择展示时补发公开展示通知，再进入“指定2名角色排除，其余所有角色各受1点法伤”的范围分支。与此同时，补齐了 `百鬼夜行` 的 `Any(1..2)` 目标元数据、`灵力崩解` 的 `CostCrystal: 1`，并把 `念咒` 的技能描述回收为文档口径，避免协议展示继续停留在旧规则上。其余 `灵符-雷鸣 / 灵符-风行 / 灵力崩解·伤害增幅` 本轮复验未发现新的文档冲突。
- `改动文件`: `internal/data/characters.go` / `internal/engine/new_roles_helpers.go` / `internal/engine/skill_flow_spirit_caster.go` / `internal/engine/game.go` / `internal/engine/spirit_caster_regression_test.go`
- `验证`: `go test ./internal/engine -run 'TestSpiritCaster'`；`go test ./internal/engine`

### [x] 28. 吟游诗人 (Bard)
- `技能`: 沉沦协奏曲 / 不谐和弦 / 禁忌诗篇 / 激昂狂想曲 / 胜利交响诗 / 希望赋格曲
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 已按最新文档全量复验。
- `结论`: 部分不一致（已按配置修复）
- `差异`: `沉沦协奏曲` 原实现把“本回合我方任意角色造成的法术伤害”都算进触发条件，而且还额外弹了一个可选确认框；现已改为严格按文档只统计“吟游诗人自己在自己回合内对至少2名不同对手造成的法术伤害”，并在满足条件且确有两张同系手牌时直接强制进入弃2张同系牌流程。`激昂狂想曲 / 胜利交响诗 / 希望赋格曲` 原本没有真正接入专属牌驱动，`激昂/胜利` 还是直接从主流程塞 choice，`希望` 也没有把打出的专属牌作为永恒乐章的在场实体；现已补齐吟游诗人开局三张专属卡，改为先走 `InterruptResponseSkill` 响应入口，再在 handler 中消耗并公开专属卡，`希望赋格曲` 的放置分支会把实际打出的【希望赋格曲】作为 `永恒乐章` 场上实体，转移分支则改为迁移已有在场实体而不是“删旧建新”。同时修正了 `不谐和弦` / `希望赋格曲` / `激昂狂想曲` / `胜利交响诗` 的目标与 `RequireExclusive` 元数据，补了 `胜利交响诗` 的星石类型选择 prompt，收敛了 `希望赋格曲` 为文档的三态 mode（放置 / 转移后+1治疗 / 转移后+1灵感），并把 `激昂/胜利` 的触发窗口改回文档写明的“吟游诗人自己的回合开始 / 回合结束”，避免继续沿用旧代码里“永恒乐章持有者回合触发”的偏差。
- `改动文件`: `internal/data/characters.go` / `internal/engine/skill_use_policy.go` / `internal/engine/skill_use.go` / `internal/engine/new_roles_helpers.go` / `internal/engine/game.go` / `internal/engine/skill_dispatcher.go` / `internal/engine/skills/handlers_bard.go` / `internal/engine/bard_regression_test.go`
- `验证`: `go test ./internal/engine -run 'TestBard'`；`go test ./internal/engine -run 'TestMagicLancer|TestSpiritCaster'`；`go test ./internal/engine`

### [x] 29. 勇者 (Hero)
- `技能`: 勇者之心 / 怒吼 / 怒吼·未命中 / 精疲力竭 / 精疲力竭·结束 / 明镜止水 / 明镜止水·行动结束收益 / 挑衅 / 禁断之力 / 死斗
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 已按最新文档全量复验。
- `结论`: 部分不一致（已按配置修复）
- `差异`: `精疲力竭·结束` 原实现挂在 `Startup` 链路里，并且会额外“摸3张牌再自伤3”，与文档要求的“下个行动阶段开始时仅结束形态并对自己造成3点法术伤害”不一致；现已改到 `ActionSelection` 开始时结算，移除错误摸牌。`挑衅` 原实现会在攻击声明尚未通过卡牌/目标合法性校验前就提前移除效果，导致无效输入也会把“下个行动阶段必须主动攻击你”的约束吃掉；现已改为只有合法攻击声明成功进入主流程后才移除，错误输入不会白白消耗挑衅。其余 `勇者之心 / 怒吼 / 怒吼·未命中 / 精疲力竭 / 明镜止水 / 明镜止水·行动结束收益 / 禁断之力 / 死斗` 本轮复验未发现新的文档冲突。
- `改动文件`: `internal/engine/game.go` / `internal/engine/hero_regression_test.go`
- `验证`: `go test ./internal/engine -run 'TestHero'`；`go test ./internal/engine -run 'Test(Hero|Fighter)'`；`go test ./internal/engine`

### [x] 30. 格斗家 (Fighter)
- `技能`: 念气力场 / 蓄力一击 / 蓄力一击·未命中自伤 / 念弹 / 百式幻龙拳 / 百式幻龙拳·主动攻击增伤 / 百式幻龙拳·应战攻击增伤 / 百式幻龙拳·禁法术与特殊行动 / 百式幻龙拳·目标锁违例收束 / 百式幻龙拳·阶段结束 / 气绝崩击 / 斗神天驱
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 已按最新文档全量复验。
- `结论`: 部分不一致（已按配置修复）
- `差异`: `百式幻龙拳` 原实现没有按文档“启动时即选定并锁定目标”，而是拖到第一次主动攻击时才记录锁定对象；现已改为启动确认后立即弹出目标选择，并把锁定目标写入状态。该技能原本还错误允许在形态中直接执行法术/特殊行动，只是顺带退形态；文档要求的是“取消当前行动并立即终止百式”，现已改为取消动作、结束形态并返回重新选行动。`百式幻龙拳·目标锁违例收束` 原实现会在改打其他目标时直接报错并取消本次攻击提交，但文档要求“仅终止百式，不取消本次攻击”；现已改为释放形态后继续本次攻击。与此同时，原实现会让百式形态跨过整个回合甚至到下回合，和文档要求的“本行动阶段结束时自动退场”不一致；现已在行动阶段结束时统一转正退场，并同步收紧行动提示，不再向前端暴露百式期间的法术/特殊行动入口。另补齐了 `念弹 / 百式幻龙拳` 的目标元数据，保持与文档协议一致。其余 `念气力场 / 蓄力一击 / 蓄力一击·未命中自伤 / 气绝崩击 / 斗神天驱` 本轮复验未发现新的文档冲突。
- `改动文件`: `internal/data/characters.go` / `internal/engine/game.go` / `internal/engine/new_roles_helpers.go` / `internal/engine/skills/handlers_fighter.go` / `internal/engine/fighter_regression_test.go`
- `验证`: `go test ./internal/engine -run 'Test(Hero|Fighter)'`；`go test ./internal/engine`

### [x] 31. 圣弓 (Holy Bow)
- `技能`: 天之弓 / 天之弓·非圣主动减伤 / 天之弓·圣命中增信仰 / 圣屑飓暴 / 圣屑飓暴·未命中惩罚 / 圣煌降临 / 圣煌降临·特殊行动退场 / 圣光爆裂 / 流星圣弹 / 圣煌辉光炮 / 自动填充
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 已按最新文档与用户更正口径全量复验，其中 `圣煌降临` 按“法术技能而非启动技能”处理。
- `结论`: 部分不一致（已按配置修复）
- `差异`: `圣光爆裂` 分支①与 `流星圣弹` 原实现都把自己错误纳入“队友/我方目标”池，现已收紧为只能选择其他我方角色。`圣屑飓暴·未命中惩罚` 原实现允许指定手牌不足 `X` 的队友，并按 `min(X, 手牌数)` 少弃，这与文档“目标队友弃 `X` 张牌”不一致；现已改为先按 `X` 过滤出能够弃满的队友，没有合法队友时整条未命中分支直接不再弹出。与此同时，补齐并收紧了 `流星圣弹 / 自动填充` 的部分协议元数据，并把 `圣煌降临 / 圣光爆裂` 的描述同步到当前文档口径，避免前端继续展示旧规则。
- `改动文件`: `internal/data/characters.go` / `internal/engine/game.go` / `internal/engine/new_roles_helpers.go` / `internal/engine/skills/handlers_holy_bow.go` / `internal/engine/holy_bow_regression_test.go`
- `验证`: `go test ./internal/engine -run 'TestHolyBow'`；`go test ./internal/engine`

### [x] 32. 剑帝 (Sword Emperor)
- `技能`: 剑魂守护 / 佯攻 / 剑气斩 / 天使之魂 / 恶魔之魂 / 天使之魂·命中结算 / 天使之魂·未命中结算 / 恶魔之魂·未命中结算 / 不屈意志
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 已按文档与本地数据模型补充口径全量复验；`SwordSoul` 上限按 `docs/data_model.md` 补齐为 `3`，`SwordQi` 上限为 `5`。
- `结论`: 原实现缺失（已按配置补齐）
- `差异`: `剑帝` 原本基本处于未落地状态，既没有 `剑魂/剑气` 的完整模型，也没有把 `剑魂守护 / 佯攻 / 天使之魂 / 恶魔之魂 / 剑气斩 / 不屈意志` 接进主战斗链。现已按文档补齐：主动攻击未命中时会先结算 `佯攻`，并在 `剑魂` 未满且未被 `天使/恶魔之魂` 禁用时，把本次攻击牌实体从弃牌堆转为隐藏 `SwordSoul` 场标；`天使之魂 / 恶魔之魂` 会在攻击前移除1张剑魂并挂载本次战斗状态，其中前者命中后给予自身 `+2治疗`、未命中时我方 `+1士气`，后者为本次主动攻击 `伤害+1` 且未命中时额外 `+2剑气`；`剑气斩` 现已按配置改为“命中后先选 `X`，再选除当前攻击目标外的任意角色”并造成 `X` 点法术伤害，`X` 严格限制为 `1..min(3, 当前剑气)`；`不屈意志` 现已在攻击行动结束时按 `[水晶]` 响应弹出，结算摸1、`+1剑气` 与额外1次攻击行动。另补了一处实现偏差：移除剑魂时原逻辑只把牌送入弃牌堆、没有真正离场，现已同步修正为正确移除场标并刷新计数。
- `改动文件`: `internal/model/types.go` / `internal/data/characters.go` / `internal/engine/combat.go` / `internal/engine/new_roles_helpers.go` / `internal/engine/game.go` / `internal/engine/skill_flow_onmyoji.go` / `internal/engine/skills/registry.go` / `internal/engine/skills/handlers_sword_emperor.go` / `internal/engine/sword_emperor_regression_test.go`
- `验证`: `go test ./internal/engine -run 'TestSwordEmperor'`；`go test ./internal/engine`

### [x] 33. 兽灵武士 (Beast Spirit Samurai)
- `技能`: 武者残心 / 一击无念 / 一击无念·下次攻击劫持 / 兽魂意念 / 兽魂警戒 / 兽返 / 御魂流居合形态·回合结束扣魂 / 御魂流居合形态·造成伤害退场 / 御魂流居合形态·兽魂归零退场 / 御魂流居合形态·横置目标增伤 / 逆反居合斩 / 御魂流居合式
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 已按文档与 `docs/data_model.md` 口径全量复验，并把姿态/形态事件与技能中断编排从主流程拆到独立 Beast Samurai flow。
- `结论`: 部分不一致（已按配置修复）
- `差异`: `兽灵武士` 原实现缺少完整的“姿态/形态”事件层，`兽魂警戒` 没有稳定的 `横置完成后` 落点；`一击无念` 仅停留在残心/额外攻击的局部实现，没有把“同窗口可接出”“下次主动攻击无视圣盾/禁圣光/技命格强制命中”完整接回主战斗链；`兽魂意念 / 兽返 / 逆反居合斩 / 御魂流居合式 / 居合形态回合结束扣魂与退场` 原先也缺少独立 choice/discard flow，部分效果仍停在 handler stub 或散落在主流程外。现已补齐统一的 `Orientation/Form` 模型与 `TriggerOnOrientationChanged`，新增独立 `skill_flow_beast_samurai.go` 承接 `兽返 / 逆反居合斩 / 御魂流居合式 / 兽魂警戒` 的中断编排，并把“普通形态主动攻击命中+兽魂”“横置目标增伤”“造成伤害退形态”“回合结束扣魂/兽魂归零退场”“miss/hit 后当前攻击 token 清理”完整接入主流程。
- `改动文件`: `internal/model/types.go` / `internal/model/skill.go` / `internal/data/characters.go` / `internal/engine/interface.go` / `internal/engine/new_roles_helpers.go` / `internal/engine/skill_dispatcher.go` / `internal/engine/skill_use.go` / `internal/engine/game.go` / `internal/engine/combat.go` / `internal/engine/draw_flow.go` / `internal/engine/skill_flow_beast_samurai.go` / `internal/engine/skills/handlers_beast_samurai.go` / `internal/engine/skills/registry.go` / `internal/engine/beast_samurai_regression_test.go` / `internal/engine/crimson_knight_killing_feast_regression_test.go`
- `验证`: `go test ./internal/engine -run 'TestBeastSamurai'`；`go test ./internal/engine`

### [x] 34. 灵魂术士 (Soul Sorcerer)
- `技能`: 灵魂吞噬 / 灵魂召还 / 灵魂转换 / 灵魂镜像 / 灵魂震爆 / 灵魂赐予 / 灵魂链接 / 灵魂链接·伤害转移 / 灵魂增幅
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 已按文档逐项复验，并把 `灵魂召还 / 灵魂转换 / 灵魂链接 / 灵魂链接·伤害转移` 的 choice flow 从 `game.go` 拆到独立 Soul Sorcerer flow。
- `结论`: 部分不一致（已按配置修复）
- `差异`: `灵魂吞噬` 原实现直接挂在 `TriggerBeforeMoraleLoss` 的静默 handler 上，会把非“承伤爆牌”造成的士气下降也算进去，而且在存在“士气下降前”响应时有机会先于最终改写值加黄魂；现已改为在 `applyMoraleLossAfterTrigger` 中按“最终实际士气下降值”统一结算，并严格限制为“我方角色因承受伤害导致的士气下降”。`灵魂镜像` 原实现按静态 `MaxHand` 补牌，没有尊重动态手牌上限；现已改为统一走 `GetMaxHand`。`灵魂增幅` 原 handler 还会在统一费用扣除之外再次手动扣 1 宝石，形成重复扣费；现已移除额外扣费。`灵魂链接` 的数据层元信息原先缺少 `RequireExclusive`，`灵魂召还 / 灵魂转换 / 灵魂链接` 相关选择分支也仍散落在 `game.go`；现已补齐专属卡元数据，并抽到独立 `skill_flow_soul_sorcerer.go` 承接。
- `改动文件`: `internal/data/characters.go` / `internal/engine/game.go` / `internal/engine/new_roles_helpers.go` / `internal/engine/skill_dispatcher.go` / `internal/engine/skill_flow_soul_sorcerer.go` / `internal/engine/skill_use_policy.go` / `internal/engine/skills/handlers_soul_sorcerer.go` / `internal/engine/soul_sorcerer_regression_test.go`
- `验证`: `go test ./internal/engine -run 'TestSoulSorcerer' -count=1`；`go test ./internal/engine`

### [x] 35. 月之女神 (Moon Goddess)
- `技能`: 新月庇护 / 闇月诅咒 / 美杜莎之眼 / 月之轮回 / 月渎 / 闇月斩 / 苍白之月 / 苍白之月·下次主动攻击无法应战
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 已按文档逐项复验，并把 `美杜莎之眼 / 月之轮回 / 月渎 / 闇月斩 / 苍白之月` 的 choice flow 从 `game.go` 拆到独立 Moon Goddess flow。
- `结论`: 部分不一致（已按配置修复）
- `差异`: `美杜莎之眼` 原实现的“移除法术闇月后追加1点法伤”会错误放宽成“可选任意对手”，与文档要求的“当前攻击者”不一致；现已锁定为攻击来源，并收回多余目标分支。`月渎` 原实现既没有严格限制“仅自己回合”，也错误允许在任意敌人之间重新选目标；现已改为只在月之女神自己的回合、且只对本次承受法术伤害的敌方目标追加1点法术伤害。`闇月斩` 原实现保留了 `X=0` 分支，提示与结算都允许“不移除闇月”；文档要求 `1<=X<3`，现已收紧为只能选择 `1..min(2, 当前闇月数)`。`苍白之月` 原实现的第二分支同样允许 `X=0`，甚至在没有 `新月` 时仍可进入分支②；现已改为必须至少拥有1点 `新月`，且 `X` 只能在 `1..当前新月数` 中选择。另补了一处架构层遗漏：移除最后1个闇月后，原 helper 会清 token 但不派发姿态/形态变更事件；现已在闇月移除通用 helper 中补上退形态的姿态事件派发。配置文案层也同步把 `暗月` 标题/描述统一校正为文档口径的 `闇月`。
- `改动文件`: `internal/data/characters.go` / `internal/engine/game.go` / `internal/engine/new_roles_helpers.go` / `internal/engine/skill_flow_moon_goddess.go` / `internal/engine/skills/handlers_moon_goddess.go` / `internal/engine/moon_goddess_regression_test.go`
- `验证`: `go test ./internal/engine -run 'TestMoonGoddess' -count=1`；`go test ./internal/engine`

### [x] 36. 血之巫女 (Blood Witch)
- `技能`: 血之哀伤 / 流血 / 流血·回合开始自损 / 流血·手牌不足脱离 / 逆流 / 血之悲鸣 / 同生共死 / 血之诅咒
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 已按文档逐项复验，并把 `血之哀伤 / 血之悲鸣 / 同生共死 / 血之诅咒` 的 choice flow 与 deferred followup 从 `game.go` 拆到独立 Blood Priestess flow。
- `结论`: 部分不一致（已按配置修复）
- `差异`: `血之哀伤` 原实现把分支映射做成了 `0=转移 / 1=移除`，与文档的 `0=移除 / 1=转移` 相反；现已按文档收口，并同步修正文案与后续结算。`血之诅咒` 原实现会在造成2点法术伤害前先弹出弃牌选择，导致“弃3张牌”跑到了伤害前；现已改为先入伤害结算，再在伤害完成后进入弃牌后续。`流血·手牌不足脱离` 原实现是手牌一旦 `<3` 就在多个主流程节点立即脱离，和文档的 `ActionEnd` 时机不一致；现已改为在动作完整结算并回到 idle 后统一检查，从而严格贴合文档时序。与此同时，流血形态的进入/脱离也统一收敛到 helper，补上姿态/形态变更派发，避免 silent state change。`血之悲鸣` 旧实现还保留了“先检查手牌并可能延迟伤害”的额外分支，实际是为了配合旧的“立即脱离流血形态”；现随 `ActionEnd` 口径一并移除，直接按文档结算目标与自身的 `(X+1)` 法术伤害。
- `改动文件`: `internal/engine/action_summary.go` / `internal/engine/game.go` / `internal/engine/interface.go` / `internal/engine/magic.go` / `internal/engine/new_roles_helpers.go` / `internal/engine/skill_flow_blood_priestess.go` / `internal/engine/skill_use.go` / `internal/engine/skills/handlers_blood_priestess.go` / `internal/engine/blood_priestess_regression_test.go` / `internal/model/skill.go`
- `验证`: `go test ./internal/engine -run 'TestBloodPriestess' -count=1`；`go test ./internal/engine`

### [x] 37. 蝶舞者 (Butterfly Dancer)
- `技能`: 生命之火 / 舞动 / 毒粉 / 朝圣 / 镜花水月 / 凋零 / 蛹化 / 倒逆之蝶
- `核对项`: 配置元数据、触发阶段、目标与费用、核心生效链、子效果/持续态/退场结算、回归测试。
- `状态`: 已完成
- `备注`: 已按文档逐项复验，并把 `舞动 / 蛹化 / 倒逆之蝶 / 镜花水月 / 凋零` 的 choice flow 从 `game.go` 收敛到独立 Butterfly flow。
- `结论`: 部分不一致（已按配置修复）
- `差异`: `镜花水月` 原实现把两次 1 点法术伤害错误打给“原受伤目标”，与文档要求的 `TargetTriggerSource` 不一致；现已改为严格打给伤害来源，并同步修正静态技能描述。`凋零` 原实现只允许选择敌方角色，文档配置为 `SelectType: Any`；现已放宽为可选任意角色，并补上对应回归测试。`蛹化` 原实现保留了额外的确认 prompt，导致发动后不会直接进入“+1 蛹并获得4个茧”的主效果；现已改为直接结算，后续仅在需要处理手牌上限/茧上限时进入相应中断。`倒逆之蝶` 原实现把“弃2张牌”写成了自定义 choice flow，既没有挂到统一的主动技弃牌成本元数据上，也会在手牌不足时退化成“能弃几张算几张”；现已改为严格走统一 `CostDiscards=2` 的发动成本，再进入独立分支编排，从而与文档的固定代价口径一致。
- `改动文件`: `internal/data/characters.go` / `internal/engine/game.go` / `internal/engine/new_roles_helpers.go` / `internal/engine/skill_flow_butterfly_dancer.go` / `internal/engine/skill_use.go` / `internal/engine/skills/handlers_butterfly_dancer.go` / `internal/engine/butterfly_dancer_regression_test.go`
- `验证`: `go test ./internal/engine -run 'TestButterfly' -count=1`；`go test ./internal/engine`
