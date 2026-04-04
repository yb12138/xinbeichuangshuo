# 主流程技能解耦改造 TodoList

## 扫描结论（`internal/engine/turn_fsm_dispatcher.go`）

主流程中存在 3 处与具体角色技能强耦合的调用点：

1. `driveBeforeActionPhase` 里直接调用 `maybeTriggerMoonGoddessMedusa(...)`
2. `driveCombatInteractionPhase` 里直接写 `canUseShadowRejectResponseMagic(...)` 的暗灭应战特判
3. `driveActionEndStage` 里直接调用 `maybeTriggerHolySwordDrawFromPhaseEndCtx(...)`，并直接读取 `HolySwordPhaseEndPending`

这些点会让 FSM 主流程承担角色细节，增加维护成本，也不利于后续新增角色技能时复用既有机制。

## 改造目标清单

- [x] 将“攻击开始时机”的角色中断改为统一 hook 注入
- [x] 将“战斗交互策略（暗灭/暗影抗拒）”改为统一 `combatInteractionHooks` 策略注入
- [x] 将“行动结束时机”的角色中断改为统一 hook 注入
- [x] 为主流程保留通用语义注释，说明角色逻辑通过框架扩展而非硬编码

## 每个改造点的实现映射

### 1) 攻击开始时机（美杜莎之眼）

- **原耦合点**：`turn_fsm_dispatcher.go` 在攻击开始阶段直接调用角色函数。
- **复用框架**：项目已有 `runtime_policy_hooks.go` 的策略聚合模式（如 `combatInteractionHooks`、`actionSelection...Policies`）。
- **改造方式**：
  - 新增 `attackStartInterruptHook` 与 `attackStartInterruptHooks`。
  - 新增统一入口 `runAttackStartInterruptHooks(...)`。
  - 将月神逻辑封装到 `attackStartMoonGoddessMedusaInterruptHook`。
  - 主流程改为只调用统一入口，不再感知角色名。

### 2) 战斗交互策略（暗灭+暗影抗拒）

- **原耦合点**：`turn_fsm_dispatcher.go` 里写死“暗灭不可应战 + 魔剑士例外”。
- **复用框架**：现有 `combatInteractionHooks` 已用于战斗交互阶段策略插拔。
- **改造方式**：
  - 新增 `combatInteractionDarkElementResponsePolicyHook` 并注册到 `combatInteractionHooks`。
  - 主流程中删除角色特判，仅保留“策略由 hooks 统一处理”的注释。

### 3) 行动结束时机（圣剑三连击中断）

- **原耦合点**：`turn_fsm_dispatcher.go` 直接调用圣剑函数并直接消费 `HolySwordPhaseEndPending`。
- **复用框架**：沿用策略聚合模式，新增行动结束中断 hook 入口。
- **改造方式**：
  - 新增 `actionEndInterruptHook` 与 `actionEndInterruptHooks`。
  - 新增统一入口 `runActionEndInterruptHooks(ctx)`。
  - 将圣剑逻辑封装到 `actionEndHolySwordInterruptHook`。
  - 新增 `consumeActionEndInterruptSkipFlag()`，对主流程屏蔽角色字段细节（内部暂复用现有状态位）。

## 结果

- `turn_fsm_dispatcher.go` 的主流程只保留阶段推进与通用决策。
- 角色技能效果通过现有 hook/policy 体系注入，降低与主流程的直接耦合。
- 后续新增同类技能时，可直接扩展对应 hooks 列表，避免继续污染 FSM。
