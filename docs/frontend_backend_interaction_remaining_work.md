# 前后端交互架构剩余工作清单

更新时间：2026-05-20

本文档用于承接当前前后端交互架构重构的剩余工作。已完成的 `target_id`、`cancel_policy`、`Prompt.Cancelable` 清理、目标选择结构化、旧路径扫描门禁等内容不在这里重复展开；本文件只记录接下来还需要落地的工作、推荐顺序、分工方式和验收标准。

## 当前基线

已经完成：

- `PromptOption.target_id` / `PromptOptionDTO.target_id` 已作为目标选择结构化身份字段落地。
- `target_picker` 前端目标识别已禁止从 `label`、玩家名、角色名推断，改为只通过 `target_id`。
- `Prompt.Cancelable` 和 `ctx["cancelable"]` 旧取消入口已清理。
- `basic_effect_pick` 的取消语义已改为 `presentation.cancel_policy`。
- `docs/character_skill_ui_config.md` 已补充强制协议约束。
- `scripts/check_no_legacy_choice.sh` 已阻止旧 UI / 协议路径回流。
- 已补充目标同名/重复角色名场景测试，以及 `cancel_policy` 识别测试。

仍需注意：

- `cancel_policy=back` 目前只完成策略识别，尚未实现统一 Flow Runtime 回退语义。
- `PromptFlowState` 已存在，但还不是完整 `PromptFlowRuntime`；技能多步流程仍主要靠各角色 `choices.go` 手写推进。
- `PromptDialog.vue` 仍是巨型组件，renderer registry 尚未真正落地。
- 前端交互状态仍散落在 `PromptDialog`、`GameBoard`、`ActionPanel`、`interrupt.store` 等处，尚未集中为 `InteractionController`。
- WS / gameplay event 前端类型仍有手写并行结构，尚未全部生成化。

## 工作方式约定

后续继续采用双 agent 分工：

- Godel：负责复杂分析、架构拆解、风险判断、验收边界，不修改文件。
- Noether：负责低风险实现、文档和脚本落地、测试补充，不做大范围架构判断。
- 主线程：负责派工、整合、冲突检查、全量验证和最终汇报。

执行顺序建议：

1. 每个较复杂阶段先让 Godel 给出边界和验收清单。
2. 再让 Noether 实现边界清晰、文件范围明确的任务。
3. 主线程做 review、必要小修、测试和进度文档更新。

## Phase 3：拆分 `PromptDialog` 为 Renderer Registry

目标：把 `PromptDialog.vue` 从巨型分支组件逐步拆成受控 renderer 组件。第一批只拆低耦合 overlay，不碰复杂交互路径。

### 3.1 第一刀：`DecisionOverlayRenderer`

优先级：P0

状态：已完成。普通 decision overlay 已迁出到受控展示组件，`PromptDialog` 仍保留业务判断、状态和提交入口。

已新增文件：

- `web/src/components/prompt/renderers/DecisionOverlayRenderer.vue`
- `web/src/components/prompt/renderers/__tests__/DecisionOverlayRenderer.spec.ts`

已完成改动：

- 修改 `web/src/components/PromptDialog.vue`，让普通 decision overlay 改由 `DecisionOverlayRenderer` 渲染。
- `PromptDialog` 继续保留业务判断和状态，不下沉 store 读取。

renderer 约束：

- 不 import store。
- 不调用 `useSubmitAction()`。
- 不理解完整 `Prompt`。
- 只通过 props 接收标题、模式、选项、取消按钮信息。
- 只通过 emits 返回 option id 或 cancel option id。

建议最小接口：

```ts
type DecisionOverlayMode = 'numeric' | 'text' | 'activation-cost' | 'yes-no'

type DecisionOverlayOption = {
  id: string
  label: string
  buttonLabel: string
  hint?: string
  disabled?: boolean
}

type Props = {
  visible: boolean
  title: string
  mode: DecisionOverlayMode
  options: DecisionOverlayOption[]
  activationHint?: string
  activationOptionId?: string
  activationDisabled?: boolean
  canCancel: boolean
  cancelLabel: string
  cancelOptionId: string
}

type Emits = {
  select: [optionId: string]
  cancel: [optionId: string]
}
```

必须保留在 `PromptDialog` 容器里的内容：

- `showDecisionOverlay`
- `decisionOverlayMode`
- `decisionOverlayTitle`
- `inlinePrimaryButtons`
- `cancelDockButton`
- `singleActivationCostConfirmOption`
- `singleActivationCostConfirmHintText`
- `handleOptionClick`
- `canCancelPrompt`
- 自动提交相关 watcher / `autoResolveOptionId`

已覆盖测试：

- renderer 单测覆盖 text mode、numeric mode、activation-cost mode、yes-no mode、disabled option、cancel emit。
- 保留 `PromptDialog.spec.ts` 中至少一条 decision overlay 集成链路，确保 renderer emit 后仍调用原有 `handleOptionClick` 路径。
- 已运行：

```bash
cd web && npm test -- --run DecisionOverlayRenderer PromptDialog
cd web && npm run build
```

暂不做：

- 不拆 target/card/response/skill choice。
- 不抽 image button fallback composable。

### 3.2 第二刀：`FraudElementRenderer`

优先级：P1

状态：已完成。欺诈元素选择弹层已迁出到受控展示组件，`PromptDialog` 仍保留 prompt 判断、option 规整和提交入口。

已新增文件：

- `web/src/components/prompt/renderers/FraudElementRenderer.vue`
- `web/src/components/prompt/renderers/__tests__/FraudElementRenderer.spec.ts`

触发条件：

- `presentation.kind === 'branch_select'`
- `presentation.layout === 'fraud_attack_element'`

renderer 只负责展示欺诈元素选择 UI，并 emit 选中的 option id。

仍由 `PromptDialog` 保留：

- `isFraudElementCardPickerPrompt`
- `fraudElementCardOptions`
- `fraudAttackCardName`
- `handleOptionClick`

已完成改动：

- 修改 `web/src/components/PromptDialog.vue`，让 `fraud_attack_element` 弹层改由 `FraudElementRenderer` 渲染。
- 将 `prompt-fraud-*` 专属 DOM 和样式迁移到 renderer。
- 保留根节点 `data-testid="decision-overlay"` 与按钮 `data-testid="prompt-option-${option.id}"`。
- `PromptDialog` 接收 renderer emit 的 option id 后继续走 `handleOptionClick`，后端提交仍为 option index。

已覆盖测试：

- 点击元素按钮后 emit 对应 option id。
- 不影响普通 branch/numeric overlay。
- `PromptDialog.spec.ts` 中保留集成测试，确认点击 `fire` 后仍提交 `[1]`。
- 已运行：

```bash
cd web && npm test -- --run FraudElementRenderer PromptDialog DecisionOverlayRenderer
cd web && npm run build
cd web && npm test
go test ./internal/server/... ./internal/engine/...
```

### 3.3 第三刀：`AllocationOverlayRenderer`

优先级：P1

状态：已完成。圣疗治疗分配和符文改造分配的重复 overlay DOM 已合并到受控展示组件，`PromptDialog` 仍保留分配状态、校验、错误提示和最终提交。

已新增文件：

- `web/src/components/prompt/renderers/AllocationOverlayRenderer.vue`
- `web/src/components/prompt/renderers/__tests__/AllocationOverlayRenderer.spec.ts`

合并场景：

- 圣疗治疗分配：`layout === 'heal_allocate'`
- 符文改造分配：`layout === 'rune_allocate'`

renderer 约束：

- 以 controlled props 接收 `values`、`remaining`、`total`、`canSubmit`。
- 通过 `change(index, value)` 和 `submit()` emit 通知容器。
- 第一轮不把分配状态下沉到 renderer。

仍由 `PromptDialog` 保留：

- `saintHealAllocations`
- `runeReforgeAllocations`
- `setSaintHealAllocation`
- `setRuneReforgeAllocation`
- `canSubmitSaintHeal`
- `canSubmitRuneReforge`
- `submitSaintHealAllocation`
- `submitRuneReforgeAllocation`

已完成改动：

- 修改 `web/src/components/PromptDialog.vue`，让 `heal_allocate` 与 `rune_allocate` 改由 `AllocationOverlayRenderer` 渲染。
- 保留圣疗当前前端语义：分配总和 `<= 3` 可提交。
- 保留符文改造当前前端语义：分配总和必须 `=== 3` 才可提交。
- renderer 只 emit `change(index, value)` / `submit()`，不调用 `handleOptionClick`，不直接拼出站协议。
- 新增 `allocation-overlay`、`allocation-option-${index}-${value}`、`allocation-submit` 测试锚点，避免和普通 numeric decision overlay 混用。

已覆盖测试：

- renderer 单测覆盖展示、active 数字、禁用超配数字、change emit、submit emit、隐藏态。
- `PromptDialog.spec.ts` 覆盖圣疗 `[1, 2]` 提交，以及符文改造从 disabled 到 `[2, 1]` 提交。
- 已运行：

```bash
cd web && npm test -- --run AllocationOverlayRenderer PromptDialog DecisionOverlayRenderer FraudElementRenderer
cd web && npm run build
```

### 3.4 第四刀：`TargetPickerPromptRenderer`

优先级：P1

状态：已完成。`PromptDialog` 中目标选择提示行和多目标手动确认按钮已迁出到受控展示组件；真实目标点击、`target_id` 解析、同名目标防漂移和提交协议仍由 `GameBoard` / `PromptDialog` 容器逻辑保留。

已新增文件：

- `web/src/components/prompt/renderers/TargetPickerPromptRenderer.vue`
- `web/src/components/prompt/renderers/__tests__/TargetPickerPromptRenderer.spec.ts`

合并场景：

- `target_picker` 的提示文案。
- `multi_target` 场景下的确认按钮。
- counter target 选择提示文案。

renderer 约束：

- 只接收 `visible`、`message`、`showConfirm`、`canConfirm`、确认按钮图片状态等受控 props。
- 只 emit `confirm()` 和 `confirmImageError()`。
- 不读取 store。
- 不调用 `useSubmitAction()`。
- 不接收完整 `Prompt`。
- 不解析 `target_id`，不维护 `selectedTargets`。

仍由 `PromptDialog` / `GameBoard` 保留：

- `showTargetSelectionHintRow`
- `targetSelectionPromptMessage`
- `promptRequiresManualTargetConfirm`
- `selectedPromptTargetOptionIndexes`
- `canConfirmPrompt`
- `confirmPromptAction`
- `GameBoard.onTargetClick`
- `promptOptionIndexForPlayer` / `target_id` 结构化目标识别

已完成改动：

- 修改 `web/src/components/PromptDialog.vue`，用 `TargetPickerPromptRenderer` 替换原本的目标选择提示行。
- 保留 `data-testid="prompt-confirm-btn"`，避免多目标确认测试锚点变化。
- 新增 `target-picker-prompt` 测试锚点，仅用于 renderer 单测。

已覆盖测试：

- renderer 单测覆盖提示文案、无确认按钮、确认 emit、disabled 不 emit、图片错误 emit、隐藏态。
- 现有 `PromptDialog.spec.ts` 保留 target picker 提示集成链路。
- 现有 `GameBoard.spec.ts` 保留同名玩家只按 `target_id` 选择的集成链路。
- 已运行：

```bash
cd web && npm test -- --run TargetPickerPromptRenderer PromptDialog GameBoard
```

### 3.5 第五刀：`ResponsePromptRenderer`

优先级：P1

状态：已完成。响应类 prompt 的 `take` / `defend` / `counter` 内联按钮展示已迁出到受控 renderer；响应校验、选牌要求、反弹目标要求和最终提交仍由 `PromptDialog` 保留。

已新增文件：

- `web/src/components/prompt/renderers/ResponsePromptRenderer.vue`
- `web/src/components/prompt/renderers/__tests__/ResponsePromptRenderer.spec.ts`

合并场景：

- `presentation.kind === 'response'`
- 存在 `take` / `defend` / `counter` 类响应选项

renderer 约束：

- 只接收 `hint` 和由容器预处理好的 response button view model。
- 只 emit `select(optionId)` 和 `imageError(optionId)`。
- 不读取 store。
- 不调用 `useSubmitAction()`。
- 不接收完整 `Prompt`。
- 不判断是否已选择应战/防御牌，不判断反弹目标。

仍由 `PromptDialog` 保留：

- `hasCounterOption` / `hasDefendOption` / `hasCounterOrDefend`
- `promptOptionResponseKind`
- `responseAttackElementHintText`
- `cardFooterOptions` 的 response 排序
- `isMagicMissilePrompt`
- `needsCounterTargetSelection`
- `handleOptionClick` 中 `take` / `counter` / `defend` 的校验和提交分支
- prompt image button fallback 状态

已完成改动：

- 修改 `web/src/components/PromptDialog.vue`，通过 `responsePromptOptions` view model 接入 `ResponsePromptRenderer`。
- 保留按钮 `data-testid="prompt-option-${option.id}"`。
- 保留 `take -> defend -> counter` 的现有排序来源。
- 图片加载失败仍回到 `PromptDialog` 的统一 fallback 状态。

已覆盖测试：

- renderer 单测覆盖提示文案、按钮 class、large 按钮、fallback 文本、disabled 不 emit、image error emit、隐藏态。
- `PromptDialog.spec.ts` 覆盖 response prompt 仍不进入 decision overlay，点击 `take` 仍调用 `submitRespondTake()`，不回退成 `submitSelect()`。
- 已运行：

```bash
cd web && npm test -- --run ResponsePromptRenderer PromptDialog
```

### 3.6 第六刀：`ExtractPromptRenderer`

优先级：P1

状态：已完成。`PromptDialog` 中提取选择的内联 UI（两列选项 + 确认按钮）已迁出到受控 renderer；提取选择状态、选择校验和最终提交协议仍由 `PromptDialog` 保留。

已新增文件：

- `web/src/components/prompt/renderers/ExtractPromptRenderer.vue`
- `web/src/components/prompt/renderers/__tests__/ExtractPromptRenderer.spec.ts`

合并场景：

- `presentation.layout === 'extract'`
- 提取选项按钮网格
- 提取确认按钮（含图片按钮 fallback）

renderer 约束：

- 只接收 `visible`、`options`、`selectedIndexes`、`min`、`max`、确认按钮图片状态等受控 props。
- 只 emit `toggle(index)`、`confirm()`、`confirmImageError()`。
- 不读取 store。
- 不调用 `useSubmitAction()`。
- 不接收完整 `Prompt`。
- 不拼接出站提交协议。

仍由 `PromptDialog` 保留：

- `isExtractPrompt`
- `selectedExtractIndices`
- `toggleExtractOption`
- `confirmExtractSelection`
- 提交 `actions.submitSelect(sel)` 的 index 协议
- prompt image button fallback 状态统一管理

已完成改动：

- 修改 `web/src/components/PromptDialog.vue`，用 `ExtractPromptRenderer` 替换 extract 内联模板。
- 选项按钮增加稳定锚点 `data-testid="extract-option-${idx}"`。
- 保留确认按钮 `data-testid="prompt-confirm-btn"`。
- 保留确认按钮 title / aria-label 语义：`确认提炼（selected/max）`。
- 图片加载失败仍回到 `PromptDialog` 的统一 fallback 路径。

已覆盖测试：

- renderer 单测覆盖红宝石/蓝水晶展示、option toggle emit、selected class、confirm disabled 不 emit、confirm emit、image error emit、hidden 不渲染。
- `PromptDialog.spec.ts` 覆盖 extract prompt 集成链路：选中两个 option 后确认仍提交 `submitSelect([0, 1])`。
- 已运行：

```bash
cd web && npm test -- --run ExtractPromptRenderer PromptDialog
```

## Phase 4：后端 Prompt Flow Runtime

目标：把多步技能从手写 `ctxData["choice_type"]` 推进，迁移到统一 `PromptFlowRuntime`。

优先级：P1，但必须先做设计分析，不建议直接开改。

### 4.1 Runtime 设计分析

交给 Godel。

需要输出：

- `PromptFlowRuntime` 与现有 `ChoiceEngine` 的边界。
- 是否放在 `internal/model`、`internal/engine/runtime/choice` 或新包。
- step spec 的最小结构。
- `cancel_policy=back/abort/decline` 的运行时语义。
- 如何兼容当前 `PromptFlowState`。
- 第一批迁移哪个技能风险最低。

### 4.2 最小 Runtime 能力

建议先实现：

- `FlowID`
- `StepID`
- `History`
- `AccumulatedSelections`
- `Advance(stepID)`
- `Back()`
- `Abort()`
- `RecordDecline()`

验收：

- 单元测试覆盖前进、回退、终止、跳过。
- 不要求第一轮迁移所有技能。
- 不删除旧 `choice_type` 调试字段，直到至少一个复杂技能稳定迁完。

### 4.3 第一批技能迁移候选

候选顺序：

1. 剑帝目标选择：目标相对集中，适合验证 step spec。
2. 圣洁法典：已有 flow 状态，适合验证多目标和 accumulated selections。
3. 毒粉：适合验证 `abort`。
4. 月神相关响应：风险较高，建议等 runtime 基础稳定后再动。
5. 阴阳师：流程复杂，不作为第一批。

## Phase 5：前端 InteractionController

目标：集中管理当前 prompt、当前选择和提交 intent，降低 `GameBoard`、`ActionPanel`、`PromptDialog` 的隐式耦合。

优先级：P2，建议在 Phase 3 第一批 renderer 完成后开始。

建议新增：

- `web/src/composables/useInteractionController.ts`
- 或 `web/src/stores/interaction.controller.ts`

最小 intent：

- `selectCard(cardId)`
- `selectTarget(targetId)`
- `submitOption(optionIndex)`
- `submitCards(cardIds)`
- `submitTargets(targetIds)`
- `cancelPrompt()`
- `clearSelection()`

第一批只迁移：

- target picker 目标选择。
- card picker 确认提交。

暂不迁移：

- 战斗响应 counter/defend/take。
- 技能面板主动施放。
- action hub。

测试要求：

- 卡牌选择不会污染目标选择。
- target picker 切换 prompt 后清空旧目标。
- `submitTargets` 仍输出 `option_indexes`，不直接拼旧协议字段。

## Phase 6：WebSocket / Gameplay Event 类型生成增强

目标：减少前端手写协议漂移。

优先级：P2。

建议步骤：

1. 盘点 `web/src/types/game.ts` 中手写 `GameEvent` union。
2. 扩展 `scripts/generate_types.go`，生成 WS message payload discriminated union。
3. 生成 gameplay event payload union。
4. 将 `Data?: unknown` 限定在最低层 envelope decode。
5. 业务 handler 改收 typed payload。

验收：

- 后端新增事件字段或改字段时，前端类型检查能暴露不匹配。
- `npm run build` 能作为协议漂移门禁。

## Phase 7：清理旧路径与文档

目标：防止已经硬切换的旧路径回流。

已完成：

- 文档中写入 target/cancel 强约束。
- e2e scenario builder 自动为 target picker mock 补 `target_id`。
- `check_no_legacy_choice.sh` 增加旧 UI / 协议扫描。

剩余：

- 删除或改写文档和代码中不再成立的“兼容老客户端”“向后兼容”描述。
- 清理旧 prompt helper、无用注释、无用 choice_type UI 判断残留。
- 将 `scripts/check_no_legacy_choice.sh` 接入 `Makefile` 或 CI。

建议门禁：

```bash
bash scripts/check_no_legacy_choice.sh
go test ./internal/server/... ./internal/engine/...
cd web && npm test
cd web && npm run build
```

## 下一步推荐

最推荐下一步：

1. Godel 输出下一刀 `CardPickerRenderer` 或“选择方向 / 特殊技能选择” renderer 的最终验收边界。
2. Noether 实现单一 renderer 小刀拆分，避免同时拆多个交互路径。
3. 主线程运行前端相关测试和 build。
4. 更新 `specs/refactor/frontend-backend-interaction-followup-plan-2026-05-19.md` 的 Phase 3 进度。

暂时不要同时推进：

- `PromptFlowRuntime` 大迁移。
- `InteractionController` 大迁移。
- `PromptDialog` 多个 renderer 并行拆分。

原因：当前已经有较多协议收口改动待进入稳定期，同时打开多个架构切口会让回归定位变困难。
