# 遗留清理盘点（2026-05-19）

目标范围（仅盘点，不改代码）：
- `Prompt.Cancelable`
- `ctx["cancelable"]`
- 前端读取 `cancelable`
- target picker 基于 `label/name` 推断
- mock 缺 `target_id`
- 文档 checklist 文件

---

## 1) `Prompt.Cancelable` / `ctx["cancelable"]` 盘点

### 命中点

1. `/Users/yb/xinbeichuangshuo/internal/model/skill.go:267`
- 字段：`Cancelable bool 'json:"cancelable,omitempty"'`

2. `/Users/yb/xinbeichuangshuo/internal/engine/basic_effect_choice.go:91`
- 从上下文读取：`cancelable, _ := data["cancelable"].(bool)`

3. `/Users/yb/xinbeichuangshuo/internal/engine/basic_effect_choice.go:99`
- 写入 Prompt：`Cancelable: cancelable`

4. `/Users/yb/xinbeichuangshuo/internal/engine/choice_bindings_catalog.go:59`
- 取消路由判断：`cancelable, _ := ctx["cancelable"].(bool)`

5. `/Users/yb/xinbeichuangshuo/internal/engine/player/angel/skill_handlers.go:294`
- 唯一写入来源：`"cancelable": true`

### 结论
- 当前 `cancelable` 仅服务于 `basic_effect_pick`（风之洁净这类基础效果移除）。
- 该语义与现有 `presentation.cancel_policy` 并存，存在历史并行协议。

### 建议改法（机械）
- A. 后端统一到 `presentation.cancel_policy`：  
  将 `basic_effect_pick` 的可取消语义改为 `presentation.cancel_policy = "decline"`（或约定值），逐步移除 `Prompt.Cancelable`。
- B. 清理顺序：  
  1) 先让消费方不再依赖 `ctx["cancelable"]`；  
  2) 再删除 `Prompt.Cancelable` 字段；  
  3) 最后删除上游 `ctx["cancelable"]` 写入。

### 风险评估
- **中风险**（涉及后端 interrupt -> prompt -> cancel 路径；需要回归 `basic_effect_pick` 相关流程）。

---

## 2) 前端读取 `cancelable` 盘点

### 命中结果
- 在 `web/src` 与 `web/e2e` 中检索 `cancelable`，**无命中**。
- 现有前端取消能力走：`prompt.presentation?.cancel_policy`  
  参考：`/Users/yb/xinbeichuangshuo/web/src/components/PromptDialog.vue:385-393`

### 结论
- 前端已基本完成从 `cancelable` 到 `cancel_policy` 的迁移。

### 建议改法（机械）
- 无需前端改动。仅建议在类型层做“反向约束”：
  - 确认 DTO 不再透出 `cancelable`（由后端完成）。

### 风险评估
- **低风险**（仅文档/类型一致性确认）。

---

## 3) target picker 基于 `label/name` 推断 盘点

### 核心命中

1. `/Users/yb/xinbeichuangshuo/web/src/components/PromptDialog.vue:291-322`
- `resolveOptionPlayerId` 先用 `option.id` 命中玩家；
- 否则回退到 `option.label` 文本包含玩家 `id/name/role` 做推断。

2. `/Users/yb/xinbeichuangshuo/web/src/components/PromptDialog.vue:324-337`
- `playerOptionEntries` 依赖上述推断，影响后续目标确认映射。

3. `/Users/yb/xinbeichuangshuo/web/src/components/PromptDialog.vue:347-359`
- `selectedPromptTargetOptionIndexes` 依据玩家 ID 反查 option index。

### 结论
- target picker 仍存在“文案驱动映射”遗留 fallback（`label` 推断）。
- 这会把协议问题变成 UI 猜测，且多语言/重名时不稳定。

### 建议改法（机械）
- A. 优先使用结构化目标字段，不再解析 `label`：
  - 给目标类 `PromptOption` 增加 `target_id`（或复用 `id` 且明确 contract）。
  - `resolveOptionPlayerId` 改为：`option.target_id -> option.id(仅当是合法 playerId) -> null`。
- B. 删掉 `label` 包含匹配逻辑（`markersFor` + `includes` 分支）。

### 风险评估
- **中风险**（会影响所有 `target_picker` 交互；需 e2e 场景同步）。

---

## 4) mock 缺 `target_id` 盘点

### 协议/类型现状

1. `/Users/yb/xinbeichuangshuo/web/src/types/generated.ts:170-178`
- `PromptOptionDTO` 当前无 `target_id` 字段。

2. `/Users/yb/xinbeichuangshuo/web/src/types/game.ts:64-65`
- `PlayerAction` 有 `target_id/target_ids`，但这属于请求侧，不是 PromptOption。

### 典型 mock 缺失点（target picker 选项仅有 id/label）

1. `/Users/yb/xinbeichuangshuo/web/e2e/scenarios/angel.ts:231-234`
2. `/Users/yb/xinbeichuangshuo/web/e2e/scenarios/arbitrator.ts:245-249`
3. `/Users/yb/xinbeichuangshuo/web/e2e/scenarios/blazeWitch.ts:327-329`

> 上述都属于 `presentation.kind = "target_picker"`，但 option 没有 `target_id` 显式字段。

### 建议改法（机械）
- A. 扩展 `PromptOptionDTO`：新增 `target_id?: string`。
- B. 对 `target_picker` 场景统一补齐：
  - `target_id: <player_id>`
  - `id` 可保留兼容（建议与 `target_id` 同值，直到完全切换）。
- C. e2e 场景可机械替换模板：
  - `{ id: ENEMY_PLAYER_ID, label: 'Enemy E1', button_label: '选择' }`
  - -> `{ id: ENEMY_PLAYER_ID, target_id: ENEMY_PLAYER_ID, label: 'Enemy E1', button_label: '选择' }`

### 风险评估
- **低到中风险**  
  - “加字段 + 旧逻辑兼容读取”是低风险；  
  - “同时删除 label 推断”是中风险（建议分两步）。

---

## 5) 文档 checklist 文件盘点

### 命中结果

1. 仓库内未发现专门的 checklist 文档文件名（`*checklist*`）；
2. 发现脚本：`/Users/yb/xinbeichuangshuo/scripts/check_no_legacy_choice.sh:1-20`
   - 用于扫描并阻止旧 Choice 路由符号回流；
3. 发现带“待完成列表”的文档：`/Users/yb/xinbeichuangshuo/web/e2e/STATUS.md:110-115`
   - 包含 `- ⏳` 待办项，属于历史状态文档。

### 建议改法（机械）
- A. 若需要“清理 checklist 文件”：
  - 把 `web/e2e/STATUS.md` 标记为 `archive`（或迁移到 `specs/archive/`）。
- B. 把当前这份盘点文档纳入 `specs/refactor/`（已落库）。
- C. 如需自动化门禁，扩展 `scripts/check_no_legacy_choice.sh` 的 PAT（加入 `cancelable`/label 推断关键字）。

### 风险评估
- **低风险**（纯文档/脚本增强，不触发运行时逻辑）。

---

## 可直接做（低风险）清单

1. 在 `web/e2e` 的 `target_picker` mock 选项里批量补 `target_id`（保留原 `id` 不动）。
2. 扩展 `PromptOptionDTO`（前后端类型）增加 `target_id?: string`，暂不删旧字段。
3. 将 `web/e2e/STATUS.md` 归档或在文件头加“历史文档”标记。
4. 在 `scripts/check_no_legacy_choice.sh` 增加遗留扫描规则（只报警/阻断新增）。

## 建议分步做（中风险）

1. 删 `resolveOptionPlayerId` 的 `label/name/role` 推断 fallback。
2. 移除后端 `ctx["cancelable"]` -> `Prompt.Cancelable` 这条链路，统一到 `presentation.cancel_policy`。

---

## 备注
- 本文档为盘点结果，未修改业务代码，仅新增文档文件。
