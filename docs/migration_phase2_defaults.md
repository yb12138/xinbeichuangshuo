# Phase 2：拆分角色默认配置（role_defaults.go）

## 目标

将 `engine/role_defaults.go` 中的角色默认配置（token 初始值、治疗上限等）按角色拆分到 `player/<role>/defaults.go`，通过 `RoleEntry.Defaults` 字段统一注册。

## 当前状况

`engine/role_defaults.go` 中有一个大 map `roleDefaultConfigs`，按角色 ID 聚合了所有角色的默认配置：

```go
// engine/role_defaults.go
var roleDefaultConfigs = map[string]roleDefaultConfig{
    "plague_mage": { setMaxHeal: true, maxHeal: 5 },
    "crimson_sword_spirit": { tokens: map[string]int{"css_blood_cap": 3, "css_blood": 0} },
    "prayer_master": { tokens: map[string]int{"prayer_rune": 0} },
    // ... 22 个角色
}
```

每个条目包含以下字段：
- `setMaxHeal` / `maxHeal` — 设置治疗上限
- `addMaxHeal` — 增加治疗上限
- `addCrystal` — 增加初始水晶
- `tokens` — 角色专属指示物初始值

## 详细步骤

### Step 2.1：在每个角色目录下创建 defaults.go

以 `plague_mage` 为例：

```go
// player/plague_mage/defaults.go
package plague_mage

import "starcup-engine/internal/model"

// Defaults 初始化瘟疫法师的基础属性（治疗上限 5）。
func Defaults(player *model.Player) {
    if player == nil {
        return
    }
    player.MaxHeal = 5
}
```

以 `crimson_sword_spirit` 为例（含 tokens）：

```go
// player/crimson_sword_spirit/defaults.go
package crimson_sword_spirit

import "starcup-engine/internal/model"

// Defaults 初始化血色剑灵的基础指示物。
func Defaults(player *model.Player) {
    if player == nil {
        return
    }
    if player.Tokens == nil {
        player.Tokens = map[string]int{}
    }
    player.Tokens["css_blood_cap"] = 3
    player.Tokens["css_blood"] = 0
}
```

以 `hero` 为例（多字段组合）：

```go
// player/hero/defaults.go
package hero

import "starcup-engine/internal/model"

// Defaults 初始化勇者的基础属性（+2水晶、+1治疗上限、专属指示物）。
func Defaults(player *model.Player) {
    if player == nil {
        return
    }
    player.Crystal += 2
    player.MaxHeal += 1
    if player.Tokens == nil {
        player.Tokens = map[string]int{}
    }
    player.Tokens["hero_anger"] = 0
    player.Tokens["hero_wisdom"] = 0
    player.Tokens["hero_exhaustion_release_pending"] = 0
    player.Tokens["hero_calm_end_crystal_pending"] = 0
}
```

### Step 2.2：在 module.go 中注册 Defaults

每个角色的 `module.go` 需要导出一个 `NewRoleEntry()` 或扩展现有注册：

```go
// player/plague_mage/module.go（扩展现有文件）
package plague_mage

import (
    "starcup-engine/internal/engine/player"
    // ...
)

// RoleEntry 导出角色完整入口定义。
func RoleEntry() player.RoleEntry {
    return player.RoleEntry{
        ID:       "plague_mage",
        Defaults: Defaults,
        Choices:  NewChoiceHandler(),
        Skills:   SkillEntries(),
        // ... 其他字段
    }
}
```

**注意**：当前各角色的 `module.go` 只导出 `SkillEntries()` 和 `ChoiceRouteSpecs()`，没有统一的 `RoleEntry()` 构造函数。需新增此导出函数。

### Step 2.3：扩展 RoleEntry 注册机制

当前 `engine/role_defaults.go` 中的 `applyRoleDefaults` 方法按以下优先级初始化：

```go
func (e *GameEngine) applyRoleDefaults(player *model.Player) {
    // ...
    // 优先级1：player 子包注册的 Defaults
    if entry := roleRegistry.Entry(player.Character.ID); entry.ID != "" && entry.Defaults != nil {
        entry.ApplyDefaults(player)
        return
    }
    // 优先级2：roleDefaultConfigs map 兜底
    if cfg, ok := roleDefaultConfigs[player.Character.ID]; ok {
        cfg.apply(player)
    }
}
```

**迁移后**：随着每个角色的 `Defaults` 被移到 `player/<role>/defaults.go` 并通过 `RoleEntry.Defaults` 注册，优先级1 会覆盖优先级2，因此可以安全地逐步迁移。

### Step 2.4：更新全局注册

当前各角色模块的注册散落在 engine 的初始化代码中。需找到注册点并添加 `Defaults` 字段：

```bash
# 搜索注册点
grep -rn 'roleRegistry.Register\|Register(' internal/engine/ | grep -v '_test.go'
```

在每个角色的注册调用中补上 `Defaults`：

```go
// 注册示例
roleRegistry.Register(player.RoleEntry{
    ID:       "plague_mage",
    Defaults: plague_mage.Defaults,  // 新增
    Choices:  plague_mage.NewChoiceHandler(),
    Skills:   plague_mage.SkillEntries(),
})
```

### Step 2.5：逐步删除 roleDefaultConfigs 中的条目

每迁移一个角色，从 `roleDefaultConfigs` map 中删除对应条目：

```go
// engine/role_defaults.go — 迁移 plague_mage 后
var roleDefaultConfigs = map[string]roleDefaultConfig{
    // "plague_mage": { ... },  ← 删除
    "crimson_sword_spirit": { ... },  // 待迁移
    // ...
}
```

### Step 2.6：最终清理

当所有 22 个角色的 defaults 都迁移完毕后：
- 删除 `engine/role_defaults.go` 中的 `roleDefaultConfigs` map 和 `roleDefaultConfig` 结构体
- 保留 `applyRoleDefaults` 方法（只走 RoleEntry.Defaults 路径）
- 或者将 `applyRoleDefaults` 移到 `player/` 包下

```go
// engine/role_defaults.go — 最终清理后
package engine

import "starcup-engine/internal/model"

// applyRoleDefaults 初始化角色的基础属性（统一走 RoleEntry 注册表）。
func (e *GameEngine) applyRoleDefaults(player *model.Player) {
    if player == nil || player.Character == nil {
        return
    }
    player.Orientation = model.OrientationNormal
    player.Form = ""
    if player.Tokens == nil {
        player.Tokens = map[string]int{}
    }
    if entry := roleRegistry.Entry(player.Character.ID); entry.ID != "" {
        entry.ApplyDefaults(player)
    }
}
```

### Step 2.7：验证

```bash
go build ./...
go test ./internal/engine/... -run TestConfig -count=1
go test ./internal/engine/... -run TestRoleDefaults -count=1
```

## 完整角色默认配置清单

| 角色 | setMaxHeal | maxHeal | addMaxHeal | addCrystal | tokens |
|------|-----------|---------|------------|------------|--------|
| plague_mage | true | 5 | | | |
| crimson_sword_spirit | | | | | css_blood_cap:3, css_blood:0 |
| prayer_master | | | | | prayer_rune:0 |
| crimson_knight | true | 4 | | | crk_blood_mark:0 |
| war_homunculus | | | | | hom_war_rune:3, hom_magic_rune:0 |
| priest | true | 6 | | | |
| onmyoji | | | | | onmyoji_ghost_fire:0 |
| blaze_witch | | | | | bw_rebirth:0, bw_flame_release_pending:0 |
| magic_lancer | | | | | (空) |
| spirit_caster | | | | | (空) |
| bard | | | | | bd_inspiration:0 |
| hero | | | 0 | 2 | hero_anger:0, hero_wisdom:0, hero_exhaustion_release_pending:0, hero_calm_end_crystal_pending:0 |
| fighter | | | | | fighter_qi:0 |
| holy_bow | | | 1 | 2 | hb_cannon:1, hb_faith:0 |
| sword_emperor | | | | | se_sword_qi:0, se_sword_soul_count:0 |
| beast_samurai | | | | | bs_zanshin:0, bs_beast_soul:0 |
| holy_lancer | true | 2 | | | |
| arbiter | | | | 2 | |
| soul_sorcerer | | | | | ss_blue_soul:0, ss_yellow_soul:0 |
| moon_goddess | | | | | mg_new_moon:0, mg_petrify:0 |
| butterfly_dancer | | | | | bt_pupa:0, bt_wither_active:0 |
| blood_priestess | | | | | (空) |

## 风险

1. **注册时机**：`Defaults` 必须在 `AddPlayer()` 时调用，确保 RoleEntry 已注册到 RoleRegistry
2. **空 tokens 角色**：magic_lancer、spirit_caster、blood_priestess 虽然配置为空 map，但迁移后应确保 `player.Tokens` 仍被初始化（由 `applyRoleDefaults` 的通用逻辑保证）
3. **hero 的 addMaxHeal**：hero 的配置中 `addMaxHeal` 实际是 0（不生效），但 `addCrystal: 2` 是关键，需确保 Defaults 函数中正确处理
