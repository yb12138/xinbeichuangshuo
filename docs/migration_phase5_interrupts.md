# Phase 5：迁移中断 prompt/response 文件

## 目标

将 `engine/` 中角色专属的中断 prompt 构建和 response 处理逻辑迁移到 `player/<role>/interrupt.go`，通用框架保留在 engine。

## 当前状况

### 角色专属文件

| 文件 | 角色 | 内容 |
|------|------|------|
| `interrupt_prompt_blademaster.go` | blade_master | `buildHolySwordDrawPrompt()` — 圣剑摸牌提示 |
| `interrupt_prompt_saintess.go` | saintess | 圣女系中断 Prompt 构建 |
| `interrupt_prompt_magic_bullet.go` | magical_girl | 魔弹相关 Prompt 与选项 |
| `interrupt_response_holy_saint.go` | saintess | 圣系响应（圣击等）分支 |
| `interrupt_response_magic_blast.go` | magical_girl | 魔爆冲击弃牌响应 |
| `interrupt_response_magic_missile.go` | magical_girl | 魔弹应战/圣盾抵挡响应 |

### 通用框架文件（不迁移）

| 文件 | 说明 |
|------|------|
| `interrupt_prompt_framework.go` | Prompt 构建公共框架与选项生成 |
| `interrupt_response_runtime.go` | 响应技能确认后的恢复与回落 |
| `interrupt_install.go` | 中断安装/注册 |
| `interrupt_runtime.go` | 中断运行时管理 |

## 核心问题

中断 prompt/response 文件中的函数都是 `(e *GameEngine)` 方法，直接访问 `e.State.PendingInterrupt`、`e.State.Players`、`e.Log()` 等引擎状态。

迁移到角色包后，需要通过接口解耦。

## 详细步骤

### Step 5.1：定义 InterruptRuntime 接口

在 `player/role_entry.go` 中新增中断运行时接口：

```go
// player/role_entry.go — 新增接口
type InterruptRuntime interface {
    ChoiceRuntime  // 内嵌已有的 ChoiceRuntime

    // 中断状态访问
    PendingInterrupt() *model.Interrupt
    SetPendingInterrupt(intr *model.Interrupt)

    // Prompt 构建
    NotifyInterruptPrompt()

    // 伤害与资源
    InflictDamage(sourceID, targetID string, damage int, damageType model.DamageType)
    NotifyCardRevealed(playerID string, cards []model.Card, reason string)

    // 弃牌
    DiscardFromHand(player *model.Player, cardIdx int) (model.Card, error)

    // 攻击流程
    PopInterrupt()
    HasPendingInterrupt() bool  // 已在 ChoiceRuntime 中
}
```

在 engine 中实现该接口：

```go
// engine/role_choice_runtime.go — 扩展实现
func (r roleChoiceRuntime) PendingInterrupt() *model.Interrupt {
    if r.GameEngine == nil || r.State == nil {
        return nil
    }
    return r.State.PendingInterrupt
}

func (r roleChoiceRuntime) SetPendingInterrupt(intr *model.Interrupt) {
    if r.GameEngine == nil || r.State == nil {
        return
    }
    r.State.PendingInterrupt = intr
}

func (r roleChoiceRuntime) InflictDamage(sourceID, targetID string, damage int, damageType model.DamageType) {
    if r.GameEngine == nil { return }
    r.InflictDamage(sourceID, targetID, damage, damageType)
}

func (r roleChoiceRuntime) DiscardFromHand(player *model.Player, cardIdx int) (model.Card, error) {
    if r.GameEngine == nil || player == nil { return model.Card{}, fmt.Errorf("invalid") }
    if cardIdx < 0 || cardIdx >= len(player.Hand) { return model.Card{}, fmt.Errorf("invalid index") }
    card := player.Hand[cardIdx]
    player.Hand = append(player.Hand[:cardIdx], player.Hand[cardIdx+1:]...)
    r.State.DiscardPile = append(r.State.DiscardPile, card)
    return card, nil
}
```

### Step 5.2：迁移 blademaster 中断 prompt

```go
// player/blade_master/interrupt.go
package blade_master

import (
    "starcup-engine/internal/engine/player"
    "starcup-engine/internal/model"
)

// BuildHolySwordDrawPrompt 构建圣剑第3次攻击结束后的摸牌弃牌提示。
func BuildHolySwordDrawPrompt(rt player.InterruptRuntime) *model.Prompt {
    interrupt := rt.PendingInterrupt()
    if interrupt == nil {
        return nil
    }
    playerID := interrupt.PlayerID

    return &model.Prompt{
        Type:     model.PromptConfirm,
        PlayerID: playerID,
        Message:  "【圣剑】第3次攻击结束！选择摸X张牌然后弃X张牌 (X=0-3)：",
        Options: []model.PromptOption{
            {ID: "0", Label: "X=0"},
            {ID: "1", Label: "X=1"},
            {ID: "2", Label: "X=2"},
            {ID: "3", Label: "X=3"},
        },
        Min: 1,
        Max: 1,
    }
}
```

### Step 5.3：迁移 saintess 中断 prompt 和 response

```go
// player/saintess/interrupt.go
package saintess

import (
    "starcup-engine/internal/engine/player"
    "starcup-engine/internal/model"
)

// BuildSaintessPrompt 构建圣女系中断提示。
func BuildSaintessPrompt(rt player.InterruptRuntime, choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
    // 原 interrupt_prompt_saintess.go 的逻辑
    // 将 e.State.PendingInterrupt 替换为 rt.PendingInterrupt()
    // 将 e.xxx 替换为 rt.xxx
}

// HandleSaintessResponse 处理圣女系响应。
func HandleSaintessResponse(rt player.InterruptRuntime, act model.PlayerAction) error {
    // 原 interrupt_response_holy_saint.go 的逻辑
}
```

### Step 5.4：迁移 magical_girl 中断 prompt 和 response

```go
// player/magical_girl/interrupt.go
package magical_girl

import (
    "starcup-engine/internal/engine/player"
    "starcup-engine/internal/model"
)

// BuildMagicBulletPrompt 构建魔弹相关提示。
func BuildMagicBulletPrompt(rt player.InterruptRuntime, ...) *model.Prompt {
    // 原 interrupt_prompt_magic_bullet.go
}

// HandleMagicBlastResponse 处理魔爆冲击弃牌响应。
func HandleMagicBlastResponse(rt player.InterruptRuntime, act model.PlayerAction) error {
    // 原 interrupt_response_magic_blast.go
    // 关键转换：
    //   e.State.PendingInterrupt → rt.PendingInterrupt()
    //   e.State.Players[id] → rt.LookupPlayer(id)
    //   e.InflictDamage(...) → rt.InflictDamage(...)
    //   e.NotifyCardRevealed(...) → rt.NotifyCardRevealed(...)
    //   e.PopInterrupt() → rt.PopInterrupt()
    //   e.notifyInterruptPrompt() → rt.NotifyInterruptPrompt()
}

// HandleMagicMissileResponse 处理魔弹应战/圣盾抵挡响应。
func HandleMagicMissileResponse(rt player.InterruptRuntime, act model.PlayerAction) error {
    // 原 interrupt_response_magic_missile.go
}
```

### Step 5.5：通过 RoleEntry 注册中断处理器

扩展 `RoleEntry` 新增中断处理字段：

```go
// player/role_entry.go — 新增
type InterruptHandler interface {
    BuildPrompt(rt InterruptRuntime, choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt
    HandleResponse(rt InterruptRuntime, act model.PlayerAction) error
}

type RoleEntry struct {
    // ... 已有字段 ...

    // 新增：角色专属中断处理器
    Interrupt InterruptHandler
}
```

在 `module.go` 中注册：

```go
// player/blade_master/module.go
func RoleEntry() player.RoleEntry {
    return player.RoleEntry{
        ID:        "blade_master",
        Defaults:  Defaults,
        Choices:   NewChoiceHandler(),
        Skills:    SkillEntries(),
        Interrupt: bladeMasterInterruptHandler{},  // 新增
    }
}

type bladeMasterInterruptHandler struct{}

func (bladeMasterInterruptHandler) BuildPrompt(rt player.InterruptRuntime, choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
    if choiceType == "holy_sword_draw" {
        return BuildHolySwordDrawPrompt(rt)
    }
    return nil
}

func (bladeMasterInterruptHandler) HandleResponse(rt player.InterruptRuntime, act model.PlayerAction) error {
    return nil // blade_master 的中断处理由 prompt 驱动
}
```

### Step 5.6：更新 engine 中的调用入口

在 engine 的中断处理流程中，优先通过 RoleEntry 调用角色专属处理器：

```go
// engine/interrupt_runtime.go — 修改
func (e *GameEngine) buildInterruptPrompt() *model.Prompt {
    // ... 通用逻辑 ...

    // 尝试角色专属处理器
    if interrupt := e.State.PendingInterrupt; interrupt != nil {
        if player := e.State.Players[interrupt.PlayerID]; player != nil && player.Character != nil {
            if entry := roleRegistry.Entry(player.Character.ID); entry.ID != "" && entry.Interrupt != nil {
                if prompt := entry.Interrupt.BuildPrompt(newRoleChoiceRuntime(e), ...); prompt != nil {
                    return prompt
                }
            }
        }
    }

    // 兜底：通用 prompt 构建
}
```

### Step 5.7：删除 engine 中的原文件

```bash
rm internal/engine/interrupt_prompt_blademaster.go
rm internal/engine/interrupt_prompt_saintess.go
rm internal/engine/interrupt_prompt_magic_bullet.go
rm internal/engine/interrupt_response_holy_saint.go
rm internal/engine/interrupt_response_magic_blast.go
rm internal/engine/interrupt_response_magic_missile.go
```

### Step 5.8：验证

```bash
go build ./...
go test ./internal/engine/... -run TestInterrupt -count=1
go test ./internal/engine/... -run TestMagicBlast -count=1
go test ./internal/engine/... -run TestSaintess -count=1
go test ./internal/engine/... -run TestBladeMaster -count=1
```

## 迁移对照表

| 原文件 | 目标文件 | 主要函数 | 接口需求 |
|--------|---------|---------|---------|
| `interrupt_prompt_blademaster.go` | `player/blade_master/interrupt.go` | `buildHolySwordDrawPrompt()` | PendingInterrupt() |
| `interrupt_prompt_saintess.go` | `player/saintess/interrupt.go` | 多个 prompt 构建函数 | PendingInterrupt(), LookupPlayer() |
| `interrupt_prompt_magic_bullet.go` | `player/magical_girl/interrupt.go` | 魔弹 prompt 构建 | PendingInterrupt(), LookupPlayer() |
| `interrupt_response_holy_saint.go` | `player/saintess/interrupt_response.go` | 圣系响应处理 | 全套 InterruptRuntime |
| `interrupt_response_magic_blast.go` | `player/magical_girl/interrupt_blast.go` | 魔爆冲击响应 | 全套 InterruptRuntime + InflictDamage + DiscardFromHand |
| `interrupt_response_magic_missile.go` | `player/magical_girl/interrupt_missile.go` | 魔弹应战响应 | 全套 InterruptRuntime |

## 风险

1. **InterruptRuntime 接口膨胀**：魔爆冲击的响应逻辑复杂（多阶段、多玩家切换），可能需要在接口中暴露较多方法
2. **中断上下文格式**：`interrupt.Context` 是 `map[string]interface{}`，迁移后需确保上下文读写格式一致
3. **notifyInterruptPrompt vs NotifyInterruptPrompt**：engine 中是小写私有方法，迁移后需通过接口暴露
4. **测试覆盖**：中断响应的回归测试可能依赖 `GameEngine` 的完整状态，迁移后需适配
