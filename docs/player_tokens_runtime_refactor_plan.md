# `Player.Tokens` 非指示物数据迁出方案

本文档针对 [player_tokens_keys_appendix.md](player_tokens_keys_appendix.md) 中与 **§1.7 专属指示物** 无关的 key，给出**不再写入 `Tokens`** 的优化方向：把「流程 / 回合 / 锁」「多步技能暂存」「UI 派生与可见性」分层到专门结构，避免 `map[string]int` 成为全局垃圾场。

---

## 1. 原则

| 保留在 `Tokens`（狭义「指示物」） | 不保留在 `Tokens` |
|:---|:---|
| 角色规则书上可计数的专属资源（与 `docs/data_model.md` §1.7 对齐） | 回合/行动/效果链生命周期内的布尔或计数 |
| 需持久跨回合、重放与存档一致的资源 | 仅控制 UI 或对手视野的派生值 |
| 客户端有权完整感知的「公开计数」 | 多步技能中间量（应在技能上下文或中断上下文关闭时一并回收） |

迁出后：`Player.Tokens` 仅承载 **§1.7 型** key，或重命名为 `ResourceCounters` 等以消除历史歧义（可选，需同步序列化与前端约定）。

---

## 2. 目标架构（模型分层）

### 2.1 回合级流程态 → `PlayerTurnState` 或并列结构体

现状已有 `model.PlayerTurnState`（`internal/model/skill.go`），负责本回合行动队列、额外行动约束等。**建议**：将 `turnScopedResetKeys`（`internal/engine/turn_progression_runtime.go`）所列 key 及同类 **「新回合清零」** 字段，统一迁入：

- **方案 A（优先）**：扩展 `PlayerTurnState`，增加具名字段或小型嵌套 `TurnScopedFlags`（`struct` + `json` tag），在 `prepareNextTurnRuntime` / `NewPlayerTurnState` 一处清零，**删除**对字符串 key 的循环重置。
- **方案 B**：新建 `PlayerFlowRuntime`，挂在 `Player` 上，与 `TurnState` 并列；仅当 TurnState 已过于臃肿时采用。

**典型迁入对象**（附录与 §1.7.1-B 已列举）：`mb_magic_pierce_pending`、`fighter_*_lock` / `*_pending`、`hb_shard_miss_pending`、`mg_blasphemy_*`、`arbiter_forced_doomsday_pending`、`hero_taunt_active_turn`、`bd_*_used_turn` / `*_prompted_turn`、`plague_*`、`bt_wither_pending`、`bp_bleed_tick_done_turn`、`valkyrie_military_glory_done_turn` 等。

`mg_extra_turn_pending` 等与「回合切换逻辑」强绑定，同样不宜留在泛型 map。

### 2.2 行动链 / FSM 级 → `GameEngine` 或 `PlayerTurnState.LastActionType` 周边显式字段

**不宜长期放在 `Player.Tokens`**：

- `post_action_end_effect_pending`、`post_action_end_effect_magic`
- `special_phase_end_dispatched`
- `holy_sword_phase_end_pending`

**建议**：在 `GameEngine` 上增加极小的 **行动收尾上下文**（例如 `ActionClosure struct { PendingPostEffects bool; MagicBranch bool; … }`），或挂在当前 `Subflow`/阶段枚举关联的结构体上；生命周期 = 单次行动提交到 turn FSM 收敛，避免与玩家资源混淆。

### 2.3 单次攻击 / 效果链内标志 → 已有 `RuleModifier` / 战斗上下文

部分 key 实为协议层修饰（如 `fighter_qiburst_force_no_counter`、`hero_calm_force_no_counter`、`mg_next_attack_no_counter`、`se_*_armed`、`holy_lancer_sky_spear_no_counter`）。

**建议优先级**：

1. 能表达为 **限时规则修饰物**（`ActiveRuleModifiers` + `RuleLifeUntilTurnEnd` / `RuleLifeThisEffectChain`）的，以修饰物为**唯一真源**，从 `Tokens` 删除镜像字段。
2. 必须随「单次攻击」清除的，放入 `AttackEventContext` / `lifecycle` 已存在路径上的结构体（见 `internal/engine/attack_lifecycle.go` 等），在攻击开始/结束时对称清零。

### 2.4 多步技能 / 敏感暂存 → `PendingInterrupt`/专用 `SkillResume` 负载

**应从 `Tokens` 迁出**（避免泄漏到快照与错误 UI）：

- `ml_stardust_pending`、`ml_stardust_wait_discard`、`ml_stardust_morale_before`、`ml_stardust_locked_target_order`
- `bw_pain_link_pending_discard`、`bw_pain_link_pending_hits`
- `adventurer_extract_last_gem`、`adventurer_extract_last_crystal`、`adventurer_extract_requires_paradise`（已在 `state_view` 中对部分 key `delete`，说明本就不该当「公共指示物」）

**建议**：在 `PendingInterrupt` 或 `QueuedAction` / 选择阶段 `Context` 中增加 **小型 typed struct**（按 skill_id 分支或泛型 `map[string]any` 仅限中断存活期），技能完成或取消时统一 `defer` 式清理。**禁止**依赖 `Tokens` + 视图层 `delete` 做隐私隔离。

### 2.5 UI 派生与对手可见性 → `PlayerView` 显式字段 / `stateview` 纯函数

现状：`internal/server/state_view.go` 从 `p.Tokens` 拷贝后再 **delete** 一批 `*_form`，并对 `ml_*_bonus`、`elf_blessing_count`、`mb_charge_count` 等 **重写或推导**。说明这些值的真源本就在 **Form、Blessings、盖牌区、RuleModifier**，`Tokens` 只是历史债。

**建议**：

1. **真源**：盖牌数量、祝福张数、暗月张数等只从 `Player` 的切片/区域 + `stateview` 计数函数读取（已部分实现）。
2. **协议**：扩展 `PlayerView`（或并列 `DisplayCounters`），字段显式命名，例如 `ElfBlessingCount`、`MagicBowChargeCountPublic`、`MoonDarkMoonCountPublic`，由 `buildStateForPlayer` **单次写入**，不再先写入 `Tokens` 再覆盖。
3. **对手不可见**：不要写入引擎内 `Player` 的任意 map；仅在 **面向该 viewer** 的 `view` 上裁剪字段（现有的 Hand 掩码模式）。

这样可删除大段「复制 Tokens 再 delete」的逻辑，降低前后端对「幽灵 key」的耦合。

---

## 3. 迁移阶段（降低风险）

| 阶段 | 内容 |
|:---|:---|
| **0** | 约定：`Tokens` 仅增 §1.7 型 key；流程类新需求走 `TurnState`/Engine 上下文。可选：静态检查或 codegen 列表比对附录。 |
| **1** | 引入 `TurnScopedFlags`（或扩展 `PlayerTurnState`），对 `turnScopedResetKeys` 双写：读路径仍以 `Tokens` 为主。 |
| **2** | 切换读路径到结构体；回归测试覆盖仲裁者、圣弓、月女神、格斗家等关键角色。 |
| **3** | 移除 `Tokens` 中对应 key 的写入与 `resetTurnScopedPlayerTokens` 中的字符串重置。 |
| **4** | `state_view`：派生计数改为 `PlayerView` 显式字段；精简 `delete(view.Tokens, …)`。 |
| **5** | 多步技能上下文从中断/子流程迁入 typed struct，删除 `adventurer_extract_*` 等 Tokens。 |

每阶段保持 **存档 JSON** 兼容时可增加迁移钩子：旧存档若仍含历史 key，在 `LoadGame` 时填充到新结构体并清空旧 key。

---

## 4. 验收标准

- `Player.Tokens` 的 key 集合与 [player_tokens_keys_appendix.md](player_tokens_keys_appendix.md) 对比：**流程类 key 数量单调下降**。
- 无「视图层删除才能保密」的路径；保密数据不落地到可序列化的 `Player` 共享 map。
- 新回合/新行动边界上，状态清零位置 **可枚举**（打开一个文件即能看懂），而非散落字符串 key。

---

## 5. 与现有文档的关系

- 数据模型约定：[data_model.md](data_model.md) §1.7.1  
- 完整 key 索引：[player_tokens_keys_appendix.md](player_tokens_keys_appendix.md)  
- 重新生成附录：`python3 scripts/gen_player_tokens_appendix.py`
