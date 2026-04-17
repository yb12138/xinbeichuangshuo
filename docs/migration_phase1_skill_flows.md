# Phase 1：迁移技能流文件（skill_flow_*.go）

## 目标

将 `internal/engine/skill_flow_<role>.go`（共 32 个文件）迁移到 `internal/engine/player/<role>/skill_flow.go`。

迁移后每个角色目录结构：
```
player/<role>/
├── module.go        # 已有 — SkillEntries(), ChoiceRouteSpecs()
├── choices.go       # 已有 — ChoiceHandler 实现
└── skill_flow.go    # 新增 — 从 engine/ 迁入的技能流逻辑
```

## 核心问题：包循环引用

当前 `skill_flow_*.go` 的代码模式是 `func (e *GameEngine) method()`，即 `package engine` 下的 GameEngine 方法。迁移到 `player/<role>` 后变成 `package <rolename>`，不能再 import `engine` 包（否则循环引用）。

**解决方案**：所有 `(e *GameEngine)` 方法改为接受 `ChoiceRuntime` 接口的独立函数，与已有的 `choices.go` 中 `ChoiceHandler` 模式对齐。

## 详细步骤

### Step 1.1：分析原文件依赖

对每个 `skill_flow_<role>.go`，逐文件分析：

```bash
# 示例：分析 skill_flow_angel.go 的依赖
cd internal/engine
grep -n 'e\.State\.' skill_flow_angel.go          # 找出所有对 GameState 的直接访问
grep -n 'e\.PopInterrupt\|e\.Log\|e\.Heal' skill_flow_angel.go  # 找出所有引擎方法调用
grep -n 'func (e \*GameEngine)' skill_flow_angel.go             # 找出所有方法接收者
```

**常见依赖模式及对应转换**：

| 原写法 (package engine) | 转换后 (package <role>) | 说明 |
|---|---|---|
| `e.State.Players[id]` | `rt.LookupPlayer(id)` | 通过 ChoiceRuntime 接口 |
| `e.Log(msg)` | `rt.Log(msg)` | ChoiceRuntime 内嵌 IGameEngine |
| `e.PopInterrupt()` | `rt.PopInterrupt()` | ChoiceRuntime 方法 |
| `e.State.PendingInterrupt` | `rt.HasPendingInterrupt()` | 只检查不直接读 |
| `e.Heal(id, n)` | `rt.Heal(id, n)` | IGameEngine 方法 |
| `e.ConsumeCrystalCost(id, n)` | `rt.ConsumeCrystalCost(id, n)` | IGameEngine 方法 |
| `e.resumePendingMoraleLoss(ctx)` | `rt.ResumePendingAttackMiss(ctx)` | ChoiceRuntime 桥接 |
| `e.enterResponseWindow()` | `rt.ApplyChoiceResumePoint(phase)` | 通过 resume point |
| `e.applyChoiceResumePoint(data)` | `rt.ApplyChoiceResumePoint(data)` | ChoiceRuntime 方法 |

**需要但 ChoiceRuntime 尚未覆盖的方法**（需先扩展接口）：

部分 skill_flow 文件会用到 `ChoiceRuntime` 当前未定义的引擎方法。遇到这种情况：
1. 先在 `player/role_entry.go` 的 `ChoiceRuntime` 接口中添加新方法声明
2. 再在 `engine/role_choice_runtime.go` 的 `roleChoiceRuntime` 结构体上添加实现

```go
// player/role_entry.go — 新增方法签名
type ChoiceRuntime interface {
    // ... 已有方法 ...

    // 新增：判断是否为指定角色
    IsCharacter(player *model.Player, charID string) bool
    // 新增：进入弃牌流程
    EnterDiscardPhase()
}

// engine/role_choice_runtime.go — 新增实现
func (r roleChoiceRuntime) IsCharacter(player *model.Player, charID string) bool {
    return player != nil && player.Character != nil && player.Character.ID == charID
}

func (r roleChoiceRuntime) EnterDiscardPhase() {
    if r.GameEngine == nil { return }
    r.enterDiscardPhase()
}
```

### Step 1.2：创建目标文件

```bash
# 以 angel 为例
cp internal/engine/skill_flow_angel.go internal/engine/player/angel/skill_flow.go
```

### Step 1.3：修改包声明和 import

```go
// 修改前 (engine/skill_flow_angel.go)
package engine

import (
    "fmt"
    "starcup-engine/internal/engine/core/runtimeutil"
    "starcup-engine/internal/model"
)

// 修改后 (player/angel/skill_flow.go)
package angel

import (
    "fmt"
    "starcup-engine/internal/engine/player"
    "starcup-engine/internal/engine/core/runtimeutil"
    "starcup-engine/internal/model"
)
```

### Step 1.4：转换方法接收者为独立函数

```go
// 修改前：(e *GameEngine) 方法
func (e *GameEngine) buildAngelChoicePrompt(choiceType, playerID string, _ *model.Player, data map[string]interface{}) *model.Prompt {
    // ...
    options := buildPromptOptionsForPlayerIDs(e.State.Players, ...)
    // ...
    e.PopInterrupt()
    if e.State.PendingInterrupt == nil && e.resumePendingMoraleLoss(userCtx) {
        return nil
    }
}

// 修改后：独立函数，接受 ChoiceRuntime
func buildAngelChoicePrompt(rt player.ChoiceRuntime, choiceType, playerID string, _ *model.Player, data map[string]interface{}) *model.Prompt {
    // e.State.Players → rt.AllPlayers()
    options := buildPromptOptionsForPlayerIDs(rt.AllPlayers(), ...)
    // ...
    rt.PopInterrupt()
    if !rt.HasPendingInterrupt() && rt.ResumePendingAttackMiss(userCtx) {
        return nil
    }
}
```

### Step 1.5：通过 module.go 注册到 RoleEntry

迁移后的函数不再通过 `engine` 包直接调用，而是通过 `module.go` 注册到 `RoleEntry` 中，由 engine 的桥接层统一调度。

```go
// player/angel/module.go — 扩展导出
package angel

import (
    "starcup-engine/internal/engine/player"
    // ...
)

// 新增：导出技能流构建器，供 ChoiceHandler 复用或直接注册
func RegisterSkillFlows(entry *player.RoleEntry) {
    // 如果 skill_flow 函数需要在 RoleEntry 层面暴露，
    // 可通过 ChoiceRoutes 或自定义回调注册
}
```

**注意**：大部分 skill_flow 文件实际上就是 choices.go 的逻辑补充（build prompt + handle input）。最优方案是将 skill_flow 中的 `build*ChoicePrompt` 和 `handle*ChoiceInput` 方法合并到已有的 `ChoiceHandler` 实现中：

```go
// player/angel/choices.go — 合并后的完整版本
type choiceHandler struct{}

func NewChoiceHandler() player.ChoiceHandler {
    return choiceHandler{}
}

func (choiceHandler) BuildPrompt(rt player.ChoiceRuntime, choiceType, playerID string, _ *model.Player, data map[string]interface{}) *model.Prompt {
    switch choiceType {
    case "angel_bond_heal_target":
        // 原 skill_flow_angel.go 的 buildAngelChoicePrompt 逻辑
        return buildAngelBondHealTargetPrompt(rt, playerID, data)
    case "god_protection_x":
        // 原 skill_flow_angel.go 的 god_protection_x 逻辑
        return buildGodProtectionXPrompt(rt, playerID, data)
    default:
        return nil
    }
}

func (choiceHandler) HandleChoice(rt player.ChoiceRuntime, playerID string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
    choiceType, _ := ctxData["choice_type"].(string)
    switch choiceType {
    case "angel_bond_heal_target":
        return true, handleAngelBondHealChoice(rt, playerID, selectionIndex, ctxData)
    case "god_protection_x":
        return true, handleGodProtectionChoice(rt, selectionIndex, ctxData)
    default:
        return false, nil
    }
}

// 私有辅助函数 — 从 skill_flow_angel.go 迁入
func buildAngelBondHealTargetPrompt(rt player.ChoiceRuntime, playerID string, data map[string]interface{}) *model.Prompt {
    options := buildPromptOptionsForPlayerIDs(rt.AllPlayers(), runtimeutil.ParseStringSliceContextValue(data["target_ids"]))
    if len(options) == 0 {
        return nil
    }
    return &model.Prompt{
        Type:     model.PromptConfirm,
        PlayerID: playerID,
        Message:  "【天使羁绊】请选择1名角色获得+1治疗：",
        Options:  options,
        Min:      1,
        Max:      1,
    }
}

func handleGodProtectionChoice(rt player.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
    userID, _ := ctxData["user_id"].(string)
    user := rt.LookupPlayer(userID)
    if user == nil {
        return fmt.Errorf("玩家不存在")
    }
    // ... 其余逻辑中所有 e.xxx 替换为 rt.xxx ...
    rt.PopInterrupt()
    if !rt.HasPendingInterrupt() && rt.ResumePendingAttackMiss(userCtx) {
        return nil
    }
    if !rt.HasPendingInterrupt() {
        rt.NotifyInterruptPrompt()
    }
    return nil
}
```

### Step 1.6：更新 engine 桥接层

`engine/role_choice_runtime.go` 中的 `buildRoleChoicePrompt` 和 `handleRoleChoiceInput` 已通过 `roleRegistry` 桥接到 player 子包，无需额外修改。

但需确认 `engine/` 中原有的 `build*ChoicePrompt` / `handle*ChoiceInput` 方法入口已被移除或委托：

```go
// engine/ 中原有的方法调用点（如 choice_router.go 或 interrupt 流程中）
// 旧：e.buildAngelChoicePrompt(choiceType, playerID, player, data)
// 新：e.buildRoleChoicePrompt("angel", choiceType, playerID, player, data)
```

在 engine 中搜索所有 `build<Role>ChoicePrompt` 和 `handle<Role>ChoiceInput` 的调用点：

```bash
grep -rn 'buildAngel\|handleAngel\|buildBard\|handleBard\|buildSwordEmperor\|handleSwordEmperor' internal/engine/*.go \
    | grep -v 'skill_flow_' \
    | grep -v '_test.go'
```

将每个调用点替换为通过 `buildRoleChoicePrompt(roleID, ...)` 的统一路由。

### Step 1.7：删除 engine 中的原文件

确认所有调用点已迁移后，删除原文件：

```bash
rm internal/engine/skill_flow_angel.go
```

### Step 1.8：验证

```bash
go build ./...
go test ./internal/engine/... -count=1
go test ./internal/engine/player/angel/... -count=1
```

## 完整角色清单（按建议迁移顺序）

建议从简单角色开始，逐步积累模式经验：

| 批次 | 角色 | 文件 | 复杂度 |
|------|------|------|--------|
| 1 | angel | skill_flow_angel.go | 低 — 2 个 choiceType |
| 1 | archer | skill_flow_archer.go | 低 |
| 1 | assassin | skill_flow_assassin.go | 低 |
| 1 | berserker | — (无 skill_flow 文件) | 无需迁移 |
| 1 | sealer | skill_flow_sealer.go | 低 |
| 2 | saintess | skill_flow_saintess.go | 中 — 多个 choiceType |
| 2 | magical_girl | skill_flow_magical_girl.go | 中 |
| 2 | valkyrie | skill_flow_valkyrie.go | 中 |
| 2 | priest | skill_flow_priest.go | 中 |
| 3 | bard | skill_flow_bard.go + skill_flow_bard_turn_hooks.go | 高 — 双文件 |
| 3 | sword_emperor | skill_flow_sword_emperor.go | 高 — 多阶段流程 |
| 3 | hero | skill_flow_hero.go | 高 — 多 token 交互 |
| 3 | fighter | skill_flow_fighter.go | 高 |
| 4 | 其余 17 个角色 | — | 逐个评估 |

## 特殊情况处理

### 多文件角色

bard 有 2 个文件，需要一起迁移：
```
skill_flow_bard.go          → player/bard/skill_flow.go
skill_flow_bard_turn_hooks.go → player/bard/turn_hooks.go
```

`turn_hooks.go` 中的回合钩子可能需要额外的 `TurnHookRuntime` 接口支持。

### 需要扩展 ChoiceRuntime 的场景

以下操作在 skill_flow 中出现但 ChoiceRuntime 尚未支持，需按需扩展：

1. **弃牌流程**：`e.enterDiscardPhase()` → 新增 `ChoiceRuntime.EnterDiscardPhase()`
2. **摸牌**：`e.startDraw(ctx)` → 已有 `rt.StartDraw(ctx)`
3. **资源修改**：`e.ModifyGem(id, delta)` → 通过 IGameEngine 已有
4. **场上牌操作**：`player.AddFieldCard()` → 直接操作 Player 对象（通过 LookupPlayer 获取）
5. **阵营士气**：`e.addCampMorale(camp, delta)` → 需新增 `ChoiceRuntime.AddCampMorale()`

### skill_flow_system_choices.go

此文件是系统级选择（非角色专属），不迁移，保留在 engine/。
