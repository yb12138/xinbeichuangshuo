# Engine 角色逻辑插件化迁移计划

## 背景与目标

当前旧 `engine` 已经具备 `RoleEntry`、`PolicySpecs`、`TimingHookSpecs`、`ChoiceHandler`、`FollowupSpecs` 等角色插件化入口，但 `internal/engine` 根包仍残留若干直接 import 具体角色包、直接识别角色 ID / 技能 ID / choice type / followup type 的逻辑。

本迁移计划的目标是：保留旧 `engine` 的状态机与外部协议不变，将剩余角色专属逻辑逐步迁回 `internal/engine/player/<role>`，让核心流程只保留通用调度、通用规则和抽象插件分发。

完成后的核心边界：

- `internal/engine` 根包允许依赖 `internal/engine/player` 抽象包。
- `internal/engine` 根包不直接 import `internal/engine/player/<具体角色>`。
- 具体角色逻辑通过 `RoleEntry` 暴露。
- 新增、删除、修改具体角色逻辑时，原则上只改对应 `player/<role>` 目录。
- 只有无法映射到现有插件点时，才记录为接口缺口并单独评审。

## 当前已有插件点

- `RoleEntry.Defaults`：角色初始化、token、默认形态、默认资源。
- `RoleEntry.HandLimit`：手牌上限硬限制、动态修正。
- `RoleEntry.MaxHeal`：治疗上限修正。
- `RoleEntry.Choices`：角色 choice prompt 构建和选择处理。
- `RoleEntry.ChoiceRouteSpecs`：choice type 路由到角色 choice handler。
- `RoleEntry.FollowupSpecs`：延迟后续 handler 注册。
- `RoleEntry.TimingHookSpecs`：回合、战斗、伤害、行动等 timing hook。
- `RoleEntry.PolicySpecs`：行动、响应、战斗、特殊行动、技能后置等 policy hook。
- `RoleEntry.AttackCardElementTransform`：角色攻击牌元素变换。
- `RoleEntry.CannotActChecker`：角色无法行动判断。

## 逐文件迁移清单

### `internal/engine/player_hand_rules.go`

**当前问题**

- 根包直接 import 几乎所有具体角色包。
- 文件同时承担角色注册、默认值、手牌上限、治疗上限和动态修正等职责。
- 注册职责导致核心流程包需要知道所有角色目录。

**目标插件点**

- 角色注册聚合：`internal/engine/rolecatalog.BuildRoleRegistry()`。
- 默认值：`RoleEntry.Defaults`。
- 手牌上限：`RoleEntry.HandLimit`。
- 治疗上限：`RoleEntry.MaxHeal`。

**迁移内容**

- 新增 `internal/engine/rolecatalog`，由该包导入具体角色包并构建 `*player.RoleRegistry`。
- `player_hand_rules.go` 改为只调用 `rolecatalog.BuildRoleRegistry()`。
- 删除 `player_hand_rules.go` 中的具体角色 import 和 `registerXRoleEntry` 函数。
- 保留通用手牌上限计算：`roleHandLimitRule`、`roleFixedMaxHandCapValue`、`fixedMaxHandCapValue`、`applyRoleMaxHandModifiers`。
- 将原本 `registerHandLimitRoleEntries` 中的 overlay 注册移到 `rolecatalog`。
- 将血之巫女同生共死手牌上限修正从具体包调用改为通用效果扫描，避免核心 import `blood_priestess`。

**接口缺口**

- 暂无。当前阶段可通过 `rolecatalog` 和现有 `HandLimit` 解决。
- 血之巫女同生共死动态修正短期以通用 `EffectBloodSharedLife` 场效扫描实现，后续可进一步迁入角色插件。

### `internal/engine/deferred_followup_router.go`

**当前问题**

- 核心 router 直接 import `blood_priestess`。
- 血之巫女 followup handler 写在核心 router 内。
- 核心知道具体 followup type。

**目标插件点**

- `RoleEntry.FollowupSpecs`。

**迁移内容**

- 将血之巫女 followup handler 移到 `player/blood_priestess/followups.go`。
- `blood_priestess.RoleEntry()` 注册 `FollowupSpecs()`。
- 核心只保留系统 followup、角色 followup 聚合、出队、查表、调用和日志。

**接口缺口**

- 可能需要给 `FollowupHost` 补最小能力，如查玩家、推 choice interrupt、检查手牌上限、操作专属效果牌。
- 不新增新 spec 类型。

### `internal/engine/role_choice_runtime.go`

**当前问题**

- `roleChoiceRuntime` 直接 import `blood_priestess`、`holy_lancer`。
- 抽象 runtime 暴露了具体角色 helper。

**目标插件点**

- 固定手牌上限：`RoleEntry.HandLimit`。
- 治疗上限同步：`RoleEntry.MaxHeal` 或 `TimingHookSpecs`。

**迁移内容**

- 血之巫女固定手牌上限判断迁回 hand limit 规则。
- 圣枪治疗上限派生同步迁到 `MaxHeal` 或圣枪 timing hook。
- `roleChoiceRuntime` 只保留通用 host 能力。

**接口缺口**

- 可能需要补通用 `ChoiceRuntime.MaxHeal(playerID)` 或 `ChoiceRuntime.RefreshDerivedStates(playerID)`。
- 不新增角色专属 runtime 方法。

### `internal/engine/role_helpers_shared.go`

**当前问题**

- 核心 helper 直接 import 蝴蝶舞者、精灵弓手、格斗家、魔枪士、魔剑士。
- 通用手牌/卡牌 helper 与角色专属 helper 混在一起。

**目标插件点**

- 禁止施法/行动限制：`PolicyBeforeActionValidation`。
- 响应技能可用性：`PolicyResponseSkillAugment` 或角色技能 `CanUse`。
- 锁定目标/百龙状态：`PolicySpecs` 或 `TimingHookSpecs`。
- 手牌上限修正：`RoleEntry.HandLimit`。

**迁移内容**

- 魔枪士、格斗家、魔剑士禁止施法判断迁入各自 policy。
- 魔剑士暗影抗拒响应条件迁入魔剑士响应技能 handler/policy。
- 格斗家百龙锁定目标查询、清理迁入 fighter policy/timing hook。
- 蝴蝶舞者凋零影响迁入 butterfly_dancer timing/policy hook。
- 精灵弓手祝福牌计数和额外可用牌源从核心 helper 中拆出。

**接口缺口**

- 精灵弓手祝福牌属于额外可用牌源，可能需要 `PlayableCardProvider` 或先让相关技能/choice 由精灵弓手 handler 自行处理。

### `internal/engine/hand_overflow_resolution.go`

**当前问题**

- 直接 import `magic_lancer`。
- 核心在爆牌/伤害摸牌流程中特判赤血骑士热血形态和魔枪士星尘后续。

**目标插件点**

- `TimingHookSpecs`、`FollowupSpecs`、`PolicySpecs`。

**迁移内容**

- 赤血骑士热血形态逻辑迁到 crimson_knight timing hook。
- 魔枪士 `ml_stardust_wait_discard` 后续迁入 magic_lancer followup 或 timing hook。
- 核心爆牌流程只负责系统弃牌和通用 timing 广播。

**接口缺口**

- 如果没有“爆牌弃牌完成后”或“伤害摸牌完成后”timing，需要补最小 timing point。

### `internal/engine/morale_loss_runtime.go`

**当前问题**

- 直接 import `soul_sorcerer`。
- 士气损失流程中直接调用灵魂术士逻辑。

**目标插件点**

- `TimingHookSpecs`。

**迁移内容**

- 士气损失应用后广播 timing，携带受害者、最终损失值、是否来自伤害摸牌等上下文。
- 灵魂术士在本角色 timing hook 中处理灵魂吞噬。

**接口缺口**

- 如果缺少“士气损失应用后”timing，补最小 timing point。

### `internal/engine/phase_hooks.go`

**当前问题**

- 直接 import `bard`、`magic_swordsman`、`moon`。
- 回合阶段 hook 中直接调用具体角色 `Maybe...` 函数。

**目标插件点**

- `RoleEntry.TimingHookSpecs`。

**迁移内容**

- 魔剑士行动开始释放暗影迁入 magic_swordsman timing hook。
- 吟游诗人回合开始/结束触发迁入 bard timing hook。
- 月之女神回合结束月之轮回迁入 moon timing hook。
- 核心只广播 timing 并处理中断返回。

**接口缺口**

- 无。

### `internal/engine/runtime_policy_interrupt_response_rules.go`

**当前问题**

- 直接 import `beast_samurai`、`moon`。
- 核心处理兽武士响应技能增强、月之女神美杜莎之眼攻击宣言中断。

**目标插件点**

- `PolicyResponseSkillAugment`、`PolicyResponseSkillNormalize`、`PolicyAttackDeclaredInterrupt` 或 `TimingOnAttackDeclared`。

**迁移内容**

- 兽武士响应技能追加迁入 beast_samurai policy。
- 月之女神美杜莎之眼迁入 moon timing/policy hook。
- 核心只收集和规范化响应技能列表。

**接口缺口**

- 无。

### `internal/engine/runtime_policy_action_selection_rules.go`

**当前问题**

- 直接 import `arbiter`、`fighter`、`hero`。
- 核心处理行动选项、攻击限制、嘲讽消耗等角色逻辑。

**目标插件点**

- `PolicyBeforeActionOption`、`PolicyBeforeActionValidation`、`PolicyAttackCardTransform`、角色 post policy。

**迁移内容**

- 仲裁者行动选项/限制迁入 arbiter policy。
- 格斗家百龙目标锁定迁入 fighter policy。
- 英雄嘲讽限制与消费迁入 hero policy/timing。
- 核心 action selection 只聚合 policy result。

**接口缺口**

- 可能需要通用“行动成功提交后”post context，避免角色字段留在核心 result 中。

### `internal/engine/runtime_policy_turn_hooks.go`

**当前问题**

- 直接 import `butterfly_dancer`、`hero`。
- 回合状态处理仍有具体角色调用。

**目标插件点**

- `TimingHookSpecs`。

**迁移内容**

- 蝴蝶舞者回合状态处理迁入 butterfly_dancer timing hook。
- 英雄回合状态处理迁入 hero timing hook。

**接口缺口**

- 无。

### `internal/engine/runtime_policy_combat_rules.go`

**当前问题**

- 直接 import `onmyoji`。
- 核心调用阴阳师黑暗祭礼和同命格反击奖励。

**目标插件点**

- `TimingHookSpecs`、`ChoiceHandler`、`PolicyCombatCounterResolve`。

**迁移内容**

- 黑暗祭礼迁入 onmyoji timing hook。
- 同命格反击奖励迁入 onmyoji choice handler 或 combat counter resolve policy。

**接口缺口**

- 无。

### `internal/engine/policy_dispatch.go`

**当前问题**

- 通用 policy 分发器仍 import `blaze_witch`、`magic_swordsman`、`onmyoji`。

**目标插件点**

- `RoleEntry.PolicySpecs`。

**迁移内容**

- 烈焰魔女、魔剑士、阴阳师 policy 包装迁入各自 `policy_hooks.go`。
- 核心只保留装配、短路分发、聚合分发和 host adapter。

**接口缺口**

- 无。

### `internal/engine/skill_use_validation.go`

**当前问题**

- 直接 import `blaze_witch`，用于攻击牌元素计算。

**目标插件点**

- `RoleRegistry.AttackCardElementTransform` 或 `PolicyAttackCardTransform`。

**迁移内容**

- 烈焰魔女攻击元素转换注册到角色 entry。
- 核心通过 registry 或 policy 查询最终元素。

**接口缺口**

- 无。

### `internal/engine/skill_dispatcher.go`

**当前问题**

- 核心特判 `arbiter_doomsday`，并调用英雄嘲讽限制清理。

**目标插件点**

- `PolicySkillPost`。

**迁移内容**

- 仲裁者强制末日清理迁到 arbiter `PolicySkillPost`。
- 英雄嘲讽限制消费迁到 hero policy 或 timing。
- 核心只广播 `PolicySkillPost`。

**接口缺口**

- 可能需要补最小 `PolicyHost` 能力。

### `internal/engine/cannot_act_checker.go`

**当前问题**

- 核心直接判断魔剑士暗影抗拒无法行动特殊规则。

**目标插件点**

- `RoleEntry.CannotActChecker`。

**迁移内容**

- 魔剑士无法行动判断迁入 magic_swordsman `CannotActChecker`。
- 核心默认无法行动检查只做通用规则。

**接口缺口**

- 无。当前已有 `CannotActChecker`。

### `internal/engine/special_action_runtime.go`

**当前问题**

- 直接 import `adventurer`。
- 核心特判冒险者、贤者、圣弓等特殊行动。

**目标插件点**

- `PolicySpecialActionOverride`、`PolicySpecialActionPost`、角色 `ChoiceHandler`。

**迁移内容**

- 冒险者地下法则/天堂转移迁入 adventurer policy/choice。
- 贤者特殊行动分支迁入 sage policy。
- 圣弓荣耀形态特殊行动后置迁入 holy_bow policy。
- 核心只保留默认购买、提炼、补牌等通用特殊行动。

**接口缺口**

- 可能需要扩展现有 `PolicyHookResult` 支持完整接管特殊行动执行。

### `internal/engine/interrupt_runtime.go`

**当前问题**

- 直接 import `adventurer`。
- 响应技能跳过/取消时特判冒险者天堂强制响应。

**目标插件点**

- `PolicyResponseSkillNormalize` 或响应 action guard。

**迁移内容**

- 冒险者强制天堂响应判断迁到 adventurer policy。
- 冒险者提炼状态清理迁到 adventurer policy 或 followup。
- 核心 response interrupt 只处理通用确认、跳过、取消、选择。

**接口缺口**

- 可能需要补“响应 skip/cancel guard”场景，优先扩展现有 response policy context。

### `internal/engine/discard_choice_runtime.go`

**当前问题**

- 直接 import `elf_archer`。
- 核心识别兽武士 `bs_*_discard` choice type。
- 通用弃牌 prompt 中包含精灵弓手祝福牌排除逻辑。

**目标插件点**

- `ChoiceHandler`，以及可选的通用弃牌过滤扩展。

**迁移内容**

- 兽武士弃牌选择迁回 beast_samurai `ChoiceHandler`。
- 核心只认识 `system_discard_cards`。
- 精灵弓手祝福牌过滤后续单独评审。

**接口缺口**

- 可能需要 `DiscardOptionFilter` 或额外可用牌源抽象。

### `internal/engine/discard_skill_runtime.go`

**当前问题**

- 直接 import `beast_samurai`。
- `ConfirmDiscard` 中对兽武士弃牌 choice 特判。

**目标插件点**

- `ChoiceHandler`。

**迁移内容**

- 兽武士弃牌处理迁入 `player/beast_samurai/choices.go`。
- 核心 `ConfirmDiscard` 只处理系统弃牌、技能通用弃牌成本、手牌上限溢出弃牌。

**接口缺口**

- 无。

## 分阶段实施顺序

1. 文档落地与 `player_hand_rules.go` 第一批迁移。
2. 无接口缺口项：`phase_hooks.go`、`policy_dispatch.go`、`runtime_policy_interrupt_response_rules.go`、`runtime_policy_combat_rules.go`、`skill_use_validation.go`。
3. Followup 与 Choice：`deferred_followup_router.go`、`discard_skill_runtime.go`、`discard_choice_runtime.go` 的角色分支。
4. Runtime helper 与复杂流程：`role_choice_runtime.go`、`role_helpers_shared.go`、`special_action_runtime.go`、`interrupt_runtime.go`、`hand_overflow_resolution.go`、`morale_loss_runtime.go`。
5. Guardrail test：禁止 `internal/engine` 根包非测试文件直接 import 具体角色包。

## 测试矩阵

- 全量：`go test ./internal/engine/...`
- 默认值/上限：所有 `*_config_regression_test.go`
- Followup：血之巫女、刺客、水影、伤害后续相关测试
- Choice：兽武士、精灵弓手、阴阳师、月之女神、圣弓、魔枪、吟游诗人
- Policy：冒险者、英雄、格斗家、仲裁者、烈焰魔女、魔剑士
- Timing：魔剑士、吟游诗人、月之女神、灵魂术士、蝴蝶舞者、赤血骑士
- 特殊行动：冒险者提炼/天堂/地下法则、贤者、圣弓荣耀形态
