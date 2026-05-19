# 前后端交互架构后续重构计划

## Summary

本轮不是继续改协议字段，而是把已经完成的硬切换成果沉淀成更稳定的交互架构。当前 `presentation`、`card_id/card_ids`、`indicators`、typed `GameEvent` 等主干已经基本落地，后续重点是：

- 前端 prompt 渲染从巨型分支组件拆成 renderer registry。
- 后端多步 prompt 从 `ctxData` 手写推进迁移到正式 Flow Runtime。
- 取消语义统一收口到 `cancel_policy`。
- 目标选择彻底结构化，避免 label/name heuristic。
- WebSocket/gameplay event 前端类型进一步生成化。
- 前端交互状态集中到 InteractionController，减少组件间隐式耦合。

## Progress

### Phase 0：当前状态确认

- [x] 确认 `card_id/card_ids` 已成为前后端卡牌动作主路径。
- [x] 确认 `presentation` 已成为前端 prompt 渲染主依据。
- [x] 确认 `PlayerView.indicators` 已替代旧顶层派生展示字段。
- [x] 确认后端 `GameEvent` 已从裸 `Data interface{}` 迁到 typed payload。
- [x] 确认前端仍存在 `PromptDialog` 巨型分支渲染问题。
- [x] 确认后端仍存在 `ctxData["choice_type"]`、`selected_*` 等多步流程残留。
- [x] 确认目标选择仍需要进一步结构化，避免玩家名、角色名、label 反推目标。

### Phase 1：取消语义收口到 `cancel_policy`

- [x] 删除或废弃 `Prompt.Cancelable`。
- [x] 删除 engine/context 中的 `cancelable` 语义入口。
- [x] 所有 prompt 的取消行为只由 `PromptPresentation.CancelPolicy` 表达。
- [ ] 后端取消处理统一识别 `deny`、`abort`、`decline`、`back`。
- [x] 前端确认按钮、取消按钮、跳过按钮不再读取任何旧取消字段。
- [x] 清理相关旧注释和 mock 数据。
- [x] 补充测试：每种 `cancel_policy` 都有明确行为覆盖。

### Phase 2：目标选择结构化

- [x] 为目标选择 prompt 明确结构化目标身份。
- [x] 推荐方案：在 prompt option DTO 中增加 `target_id`。
- [x] 当 `presentation.kind = target_picker` 时，前端只通过 `target_id` 识别玩家或目标。
- [x] 移除前端基于 label、玩家名、角色名推断目标的 fallback。
- [x] 保留 `option_indexes` 作为非卡牌选项提交协议。
- [x] 后端生成目标选择 prompt 时必须填充稳定 target id。
- [x] 更新 e2e mock，禁止目标 prompt 只提供中文 label。
- [x] 补充测试：同名玩家、改名玩家、角色名重复时目标选择仍正确。

### Phase 3：拆分 `PromptDialog` 为 Renderer Registry

- [ ] 新建 prompt renderer registry，根据 `presentation.kind` 和必要的 `layout` 选择 renderer。
- [x] 将分支选择拆为 `BranchPromptRenderer`。
- [x] 将数字选择拆为 `NumericPromptRenderer`。
- [ ] 将卡牌选择拆为 `CardPickerPromptRenderer`。
- [ ] 将目标选择拆为 `TargetPickerPromptRenderer`。
- [x] 将响应类 prompt 拆为 `ResponsePromptRenderer`。
- [x] 将治疗、符文等分配类 prompt 拆为 `AllocationPromptRenderer`。
- [ ] 将抽取、选择方向、特殊技能选择拆为独立 renderer。（`extract` 已完成；选择方向、特殊技能选择待拆）
- [ ] `PromptDialog` 只保留容器职责：标题、关闭、提交入口、renderer 挂载。
- [ ] 每个 renderer 只处理自己的选择状态和展示逻辑。
- [ ] 单测从“一个 PromptDialog 大测”改成按 renderer 覆盖。

### Phase 4：后端 Prompt Flow Runtime

- [ ] 定义统一 `PromptFlowRuntime`，管理 `flow_id`、`step_id`、history、accumulated selections。
- [ ] 技能多步流程不再手写 `ctxData["choice_type"]` 推进步骤。
- [ ] 技能中间选择不再散落保存为 `selected_card_id`、`selected_target_id`、`selected_element` 等临时字段。
- [ ] 每个多步技能声明 step spec、输入类型、下一步 transition、完成回调。
- [ ] `cancel_policy=back` 由 runtime 回退到上一 step。
- [ ] `cancel_policy=abort` 由 runtime 终止整个 flow。
- [ ] `cancel_policy=decline` 由 runtime 记录不发动或跳过。
- [ ] 优先迁移复杂流程：圣洁法典、毒粉、月神相关响应、阴阳师、剑帝目标选择。
- [ ] 迁移完成后删除对应技能里的手写推进 context 字段。
- [ ] 补充测试：多步流程前进、回退、取消、中断恢复、重复 prompt 都能稳定执行。

### Phase 5：前端 InteractionController

- [ ] 新建前端交互控制层，集中管理当前 prompt、当前选择、可提交状态。
- [ ] 统一暴露交互 intent：`selectCard`、`selectTarget`、`submitOption`、`submitCards`、`submitTargets`、`cancelPrompt`。
- [ ] `GameBoard`、`ActionPanel`、卡牌区、目标区、prompt renderer 不再各自推断当前交互模式。
- [ ] 所有出站动作仍统一经过 action request adapter。
- [ ] 组件只负责展示和调用 intent，不直接拼协议 payload。
- [ ] 补充测试：卡牌选择、目标选择、响应选择、行动面板确认不会互相污染状态。

### Phase 6：WebSocket / Gameplay Event 类型生成增强

- [ ] 扩展类型生成，把 WS message payload 生成成前端 discriminated union。
- [ ] 生成 gameplay event payload union。
- [ ] 前端 message handlers 不再维护平行手写事件结构。
- [ ] `Data?: unknown` 只保留在最低层 envelope 解码边界。
- [ ] 业务 handler 内部只接收 typed payload。
- [ ] 补充测试：后端新增或改动事件字段时，前端类型检查能发现不匹配。

### Phase 7：清理旧交互路径和文档同步

- [ ] 删除“兼容老客户端”“向后兼容”等已不符合实际方向的注释。
- [ ] 删除不再使用的 prompt helper、label fallback、choice_type UI 判断残留。
- [x] 更新 `docs/character_skill_ui_config.md` 中与新 prompt renderer / target picker / cancel policy 相关的说明。
- [x] 更新 e2e scenario builder 规范：所有 prompt mock 必须提供完整 presentation、button_label、结构化 card/target id。
- [x] 全量静态搜索确认不存在 `Prompt.Cancelable`、`ctx["cancelable"]`、target label/name fallback 残留。
- [x] 在本文件中标记所有已完成步骤。

## Notes

- 2026-05-19：新增 `PromptOption.target_id` / `PromptOptionDTO.target_id`，前端 `PromptDialog` 与 `GameBoard` 的 target picker 均只通过 `target_id` 解析目标。
- 2026-05-19：删除 `Prompt.Cancelable`，`basic_effect_pick` 取消语义改用 `cancel_policy=decline`。
- 2026-05-19：`mg_moon_cycle_heal_target` 等目标选择 prompt 补齐 `target_filter`，避免 DTO strict boundary panic。
- 2026-05-19：补充 `GameBoard` 同名玩家 target picker 测试，以及 `basic_effect_pick` 对 `deny`/`abort`/`decline`/`back` 的取消策略识别测试；`back` 的通用 flow 回退语义仍属于 Phase 4。
- 2026-05-19：Phase 3 第一刀完成，普通分支 / numeric / yes-no / activation-cost decision overlay 已迁到 `DecisionOverlayRenderer`；尚未建立完整 renderer registry，Fraud、Allocation、target/card/response/skill choice 仍保留在 `PromptDialog`。
- 2026-05-19：Phase 3 第二刀完成，`fraud_attack_element` 欺诈元素选择弹层已迁到 `FraudElementRenderer`；`PromptDialog` 仍保留 prompt 判断、option 规整和 `handleOptionClick` 提交路径。下一刀建议拆 `AllocationOverlayRenderer`。
- 2026-05-19：Phase 3 第三刀完成，`heal_allocate` 与 `rune_allocate` 分配弹层已迁到 `AllocationOverlayRenderer`；圣疗 `<=3`、符文 `==3` 的现有前端提交语义保持不变。下一刀建议在 `TargetPickerRenderer` / `CardPickerRenderer` 中择一继续。
- 2026-05-20：Phase 3 第四刀完成，`PromptDialog` 中目标选择提示行与多目标确认按钮已迁到 `TargetPickerPromptRenderer`；真实目标点击、`target_id` 解析和提交仍保留在 `GameBoard` / `PromptDialog` 容器侧。完整 `TargetPickerRenderer` 仍需等 InteractionController 或 GameBoard 目标区拆分后再勾选。
- 2026-05-20：Phase 3 第五刀完成，响应类 prompt 的 `take` / `defend` / `counter` 内联按钮展示已迁到 `ResponsePromptRenderer`；应战/防御选牌校验、反弹目标校验和 `submitRespond*` 提交仍保留在 `PromptDialog`。
- 2026-05-20：Phase 3 第六刀完成，`extract` 布局的提取选项 UI shell（两列按钮 + 确认按钮）已迁到 `ExtractPromptRenderer`；`selectedExtractIndices`、`toggleExtractOption`、`confirmExtractSelection` 与 `submitSelect(indexes)` 协议仍保留在 `PromptDialog` 容器侧。
