# Phase 5：迁移中断 prompt/response 文件

## 目标

将 `engine/` 中角色专属的中断 prompt 构建和 response 处理逻辑迁移到 `player/<role>/interrupt.go`，
通过 `InterruptSpec` 声明式注册机制实现角色与 engine 完全解耦。

## 架构设计

### 核心思路：InterruptSpec 声明式注册

每个角色包通过 `RoleEntry.InterruptSpecs` 声明自己处理的中断类型及其 prompt/action 函数。
engine 的 `interrupt_install.go` 在初始化时遍历所有角色的 InterruptSpecs 动态注册，
不再硬编码角色专属的 prompt builder / action handler。

### InterruptSpec 定义

```go
// player/role_entry.go — 新增

// InterruptSpec 定义角色包贡献的中断处理条目。
type InterruptSpec struct {
    // Type 该 spec 处理的中断类型。
    Type model.InterruptType
    // BuildPrompt 构建该中断类型的用户提示。返回 nil 表示无提示。
    BuildPrompt func(rt ChoiceRuntime) *model.Prompt
    // HandleAction 处理该中断类型的用户响应。
    HandleAction func(rt ChoiceRuntime, act model.PlayerAction) error
    // AllowedActionTypes 允许的玩家操作类型（空=不限制）。
    AllowedActionTypes []model.PlayerActionType
    // InvalidActionMessage 操作类型不匹配时的提示。
    InvalidActionMessage string
}
```

### ChoiceRuntime 新增方法

迁移需要以下新方法（在 `ChoiceRuntime` 接口和 `roleChoiceRuntime` 实现中添加）：

| 方法 | 用途 | 使用方 |
|------|------|--------|
| `PendingInterrupt() *model.Interrupt` | 获取当前中断 | blade_master, saintess, magical_girl |
| `SetPendingInterruptContext(data map[string]interface{})` | 已有 | — |
| `SetPendingInterruptPlayerID(playerID string)` | 修改中断响应者 | magical_girl (magic_blast, magic_missile) |
| `MagicBulletChain() *model.MagicBulletChain` | 获取魔弹链 | magical_girl |
| `SetMagicBulletChain(chain *model.MagicBulletChain)` | 设置魔弹链 | magical_girl |
| `RoutePendingDamageWithDefaultReturn(defaultReturn interface{}) bool` | 路由到伤害结算 | blade_master |
| `RestoreReturnPoint() bool` | 恢复返回点 | blade_master |
| `PushDiscardChoiceInterrupt(playerID string, data map[string]interface{})` | 推入弃牌选择中断 | blade_master |
| `SetReturnPoint(returnTo interface{})` | 设置返回点 | magical_girl |
| `EnterActionEndStage()` | 进入行动结束阶段 | saintess |
| `PerformMagic(playerID, targetID string, cardIdx int, isFusion bool) error` | 执行法术行动 | magical_girl |
| `ExecuteMagicBullet(player *model.Player, reverse, isFusion bool, fusionCard *model.Card) error` | 执行魔弹传递 | magical_girl |
| `FindNextMagicBulletTarget(playerID string) string` | 查找下一个魔弹目标 | magical_girl |
| `DispatchTimingOnHitCheck(ctx interface{}) interface{}` | 命中检查时序派发 | magical_girl |
| `ConsumePlayableCardByIndex(player *model.Player, cardIdx int) (model.Card, error)` | 消耗可出手牌 | magical_girl |
| `AddPendingDamage(pd model.PendingDamage)` | 已有 (IGameEngine) | magical_girl |
| `NotifyActionStep(line string)` | 已有 (IGameEngine) | magical_girl |

**精简策略**：部分方法已在 `IGameEngine` 或 `ChoiceRuntime` 中存在，只需补充缺失的。
对于 `PerformMagic`、`ExecuteMagicBullet`、`DispatchTimingOnHitCheck` 等复杂方法，
直接通过接口暴露，避免在角色包中重新实现。

### 角色包文件布局

```
player/blade_master/
  module.go          — RoleEntry() 新增 InterruptSpecs
  interrupt.go       — BuildHolySwordDrawPrompt, HandleHolySwordDrawAction

player/saintess/
  module.go          — RoleEntry() 新增 InterruptSpecs
  interrupt.go       — BuildSaintHealPrompt, HandleSaintHealAction + helpers

player/magical_girl/
  module.go          — RoleEntry() 新增 InterruptSpecs
  interrupt.go       — 4个 prompt builders + 3个 action handlers + helpers
```

### Engine 侧变更

1. **`interrupt_install.go`**：`registerInterruptActionRules` 和 `registerInterruptPromptRules`
   改为遍历 `roleRegistry` 的 `InterruptSpecs` 动态注册，删除硬编码的角色条目。

2. **`interrupt_response_runtime.go`**：stage 常量移到各自角色包中，
   engine 只保留通用常量（如有）。

3. **删除 6 个角色专属文件**：
   - `interrupt_prompt_blademaster.go`
   - `interrupt_prompt_saintess.go`
   - `interrupt_prompt_magic_bullet.go`
   - `interrupt_response_holy_saint.go`
   - `interrupt_response_magic_missile.go`
   - `interrupt_response_magic_blast.go`

## 实施步骤

### Step 1：扩展 ChoiceRuntime 接口 + roleChoiceRuntime 实现

在 `player/role_entry.go` 的 `ChoiceRuntime` 接口中新增方法，
在 `engine/role_choice_runtime.go` 中实现桥接。

### Step 2：定义 InterruptSpec + RoleEntry 扩展

在 `player/role_entry.go` 中新增 `InterruptSpec` 类型和 `RoleEntry.InterruptSpecs` 字段。

### Step 3：迁移 blade_master（最简单，1 prompt + 1 handler）

- 新建 `player/blade_master/interrupt.go`
- 在 `module.go` 的 RoleEntry 中注册 InterruptSpecs
- 验证编译和测试

### Step 4：迁移 saintess（1 prompt + 1 handler + helpers）

- 新建 `player/saintess/interrupt.go`
- 在 `module.go` 中注册
- 验证

### Step 5：迁移 magical_girl（4 prompts + 3 handlers + helpers）

- 新建 `player/magical_girl/interrupt.go`
- 在 `module.go` 中注册
- 验证

### Step 6：更新 interrupt_install.go 为动态注册

- 遍历 roleRegistry 的 InterruptSpecs 注册
- 删除硬编码的角色条目
- 删除 6 个原文件
- 验证

## 代码量估算

| 操作 | 行数变化 |
|------|---------|
| 删除 6 个 engine 文件 | -1023 行 |
| 新增 3 个角色 interrupt.go | ~+700 行（含接口调用转换） |
| ChoiceRuntime 接口扩展 | ~+15 行 |
| roleChoiceRuntime 桥接实现 | ~+80 行 |
| InterruptSpec 定义 | ~+10 行 |
| interrupt_install.go 简化 | ~-30 行 |
| **净减少** | **~-248 行** |

## 风险

1. **接口方法数量**：ChoiceRuntime 新增约 8 个方法，需权衡是否拆分子接口
2. **魔弹响应复杂度**：`handleMagicMissileResponse` 涉及 `dispatchTimingOnHitCheck`、
   `consumePlayableCardByIndex`、`findNextMagicBulletTarget` 等多个 engine 内部方法，
   需逐一暴露到接口
3. **上下文格式一致性**：`interrupt.Context` 是 `map[string]interface{}`，
   迁移后需确保读写格式与 engine 侧一致
4. **测试覆盖**：中断响应的回归测试在 `package engine` 中，
   迁移后需确保通过接口调用的行为等价
