# Phase 3：拆分开局专属牌（starter_role_cards.go）

## 目标

将 `engine/starter_role_cards.go` 中的角色专属开局牌逻辑拆分到 `player/<role>/starter_cards.go`，通过 `RoleEntry` 注册机制统一调度。

## 当前状况

`engine/starter_role_cards.go` 包含：

1. **通用辅助函数**：`ensureExclusiveStarterCard()` — 检查并添加专属牌
2. **角色专属牌构建函数**：6 个 `makeStarter*Card()` 函数
3. **统一入口**：`ensureStarterRoleCards()` — 按角色 ID 分发

```go
// engine/starter_role_cards.go 结构
func ensureExclusiveStarterCard(player *model.Player, skillTitle string, buildCard func() model.Card) bool { ... }
func makeStarterFiveElementsBindCard(player *model.Player) model.Card { ... }       // sealer
func makeStarterRoseCourtyardCard(player *model.Player) model.Card { ... }           // crimson_sword_spirit
func makeStarterHeroTauntCard(player *model.Player) model.Card { ... }               // hero
func makeStarterSoulLinkCard(player *model.Player) model.Card { ... }                // soul_sorcerer
func makeStarterBloodSharedLifeCard(player *model.Player) model.Card { ... }         // blood_priestess
func makeStarterBardRousingRhapsodyCard(player *model.Player) model.Card { ... }     // bard
func makeStarterBardVictorySymphonyCard(player *model.Player) model.Card { ... }     // bard
func makeStarterBardHopeFugueCard(player *model.Player) model.Card { ... }           // bard
func (e *GameEngine) ensureStarterRoleCards(player *model.Player) { ... }           // 分发入口
```

## 详细步骤

### Step 3.1：在 player 包中定义 StarterCard 注册机制

首先扩展 `RoleEntry`，新增 `StarterCards` 字段：

```go
// player/role_entry.go — 扩展
type RoleEntry struct {
    ID string
    Defaults func(player *model.Player)
    HandLimit HandLimitRule
    MaxHeal func(player *model.Player, current int) int
    Choices ChoiceHandler
    Skills []SkillEntry
    ChoiceRouteSpecs map[string]types.ChoiceRouteSpec
    FollowupSpecs map[string]FollowupSpec

    // 新增：开局专属牌构建器
    StarterCards []StarterCardBuilder
}

// StarterCardBuilder 定义角色开局专属牌的构建方式。
type StarterCardBuilder struct {
    SkillTitle string
    BuildCard  func(player *model.Player) model.Card
}

// ApplyStarterCards 为角色添加开局专属牌。
func (e RoleEntry) ApplyStarterCards(ensureFn func(player *model.Player, skillTitle string, buildCard func() model.Card) bool) {
    for _, sc := range e.StarterCards {
        ensureFn(nil, sc.SkillTitle, func() model.Card {
            return sc.BuildCard(nil) // player 由调用方传入
        })
    }
}
```

**简化方案**：也可以直接将专属牌逻辑合并到 `Defaults` 回调中，避免新增字段：

```go
// 方案B：在 Defaults 中直接添加专属牌
func Defaults(player *model.Player) {
    // ... 原有 defaults 逻辑 ...

    // 专属牌初始化
    ensureExclusiveStarterCard(player, "五系束缚", func() model.Card {
        return makeStarterFiveElementsBindCard(player)
    })
}
```

推荐方案B，更简洁，无需扩展 RoleEntry。

### Step 3.2：迁移通用辅助函数

将 `ensureExclusiveStarterCard` 移到 `player/` 包下：

```go
// player/starter_cards.go
package player

import "starcup-engine/internal/model"

// EnsureExclusiveStarterCard 确保角色拥有指定的专属技能牌（去重）。
func EnsureExclusiveStarterCard(player *model.Player, skillTitle string, buildCard func() model.Card) bool {
    charID := player.Character.ID
    for _, c := range player.ExclusiveCards {
        if c.MatchExclusive(charID, skillTitle) {
            return false
        }
    }
    player.ExclusiveCards = append(player.ExclusiveCards, buildCard())
    return true
}
```

### Step 3.3：创建各角色的 starter_cards.go

以 `sealer` 为例：

```go
// player/sealer/starter_cards.go
package sealer

import (
    "fmt"

    "starcup-engine/internal/engine/player"
    "starcup-engine/internal/model"
)

// EnsureStarterCards 补充封印师开局专属技能卡。
func EnsureStarterCards(p *model.Player, log func(string)) {
    if player.EnsureExclusiveStarterCard(p, "五系束缚", func() model.Card {
        return model.Card{
            ID:              fmt.Sprintf("starter-%s-five_elements_bind", p.ID),
            Name:            "五系束缚",
            Type:            model.CardTypeMagic,
            Element:         model.ElementLight,
            Faction:         p.Character.Faction,
            Damage:          0,
            Description:     "封印师开局自带专属技能卡",
            ExclusiveChar1:  p.Character.ID,
            ExclusiveSkill1: "五系束缚",
        }
    }) {
        log(fmt.Sprintf("[Setup] %s 获得开局专属技能卡【五系束缚】（专属卡区）", p.Name))
    }
}
```

以 `bard` 为例（3 张专属牌）：

```go
// player/bard/starter_cards.go
package bard

import (
    "fmt"

    "starcup-engine/internal/engine/player"
    "starcup-engine/internal/model"
)

// EnsureStarterCards 补充吟游诗人开局专属技能卡（激昂狂想曲、胜利交响诗、希望赋格曲）。
func EnsureStarterCards(p *model.Player, log func(string)) {
    cards := []struct {
        title  string
        name   string
        element model.Element
    }{
        {"激昂狂想曲", "激昂狂想曲", model.ElementDark},
        {"胜利交响诗", "胜利交响诗", model.ElementDark},
        {"希望赋格曲", "希望赋格曲", model.ElementDark},
    }
    for _, c := range cards {
        title := c.title // capture
        if player.EnsureExclusiveStarterCard(p, title, func() model.Card {
            return model.Card{
                ID:              fmt.Sprintf("starter-%s-bd_%s", p.ID, titleToSnakeCase(title)),
                Name:            c.name,
                Type:            model.CardTypeMagic,
                Element:         c.element,
                Faction:         p.Character.Faction,
                Damage:          0,
                Description:     "吟游诗人开局自带专属技能卡",
                ExclusiveChar1:  p.Character.ID,
                ExclusiveSkill1: title,
            }
        }) {
            log(fmt.Sprintf("[Setup] %s 获得开局专属技能卡【%s】（专属卡区）", p.Name, c.name))
        }
    }
}
```

### Step 3.4：扩展 RoleEntry 注册

在 `module.go` 或单独注册时，将 `EnsureStarterCards` 绑定到 `Defaults` 中（方案B）：

```go
// player/sealer/module.go
package sealer

import (
    "starcup-engine/internal/engine/player"
    // ...
)

func RoleEntry() player.RoleEntry {
    return player.RoleEntry{
        ID:       "sealer",
        Defaults: Defaults,  // 已有的 defaults
        Choices:  NewChoiceHandler(),
        Skills:   SkillEntries(),
        // StarterCards 不单独注册，而是通过 Defaults 回调触发
    }
}
```

或者，如果使用方案A（新增字段），在 `RoleEntry` 中添加：

```go
func RoleEntry() player.RoleEntry {
    return player.RoleEntry{
        ID:       "sealer",
        Defaults: Defaults,
        StarterCards: []player.StarterCardBuilder{
            {SkillTitle: "五系束缚", BuildCard: makeStarterFiveElementsBindCard},
        },
        // ...
    }
}
```

### Step 3.5：更新 engine 中的调用入口

修改 `engine/starter_role_cards.go` 中的 `ensureStarterRoleCards` 方法，优先走 RoleEntry 注册表：

```go
// engine/starter_role_cards.go — 修改后
func (e *GameEngine) ensureStarterRoleCards(player *model.Player) {
    if player == nil || player.Character == nil {
        return
    }
    // 通过 RoleEntry 注册表触发
    if entry := roleRegistry.Entry(player.Character.ID); entry.ID != "" {
        if ensureCards, ok := getStarterCardsEnsure(entry); ok {
            ensureCards(player, e.Log)
            return
        }
    }
    // 兜底：旧逻辑（逐步移除）
}
```

### Step 3.6：逐步删除旧代码

每迁移一个角色，删除 `makeStarter*Card()` 函数和 `ensureStarterRoleCards` 中对应 case 分支。

全部迁移完成后：
- 删除 `engine/starter_role_cards.go` 中所有 `makeStarter*Card()` 函数
- 删除 `ensureStarterRoleCards` 中的 switch 分支
- 将 `ensureExclusiveStarterCard` 从 engine 包移除（已在 player 包中）

### Step 3.7：验证

```bash
go build ./...
# 验证开局牌发放
go test ./internal/engine/... -run TestStarter -count=1
go test ./internal/engine/... -run TestStartup -count=1
go test ./internal/engine/... -run TestConfig -count=1
```

## 角色专属牌清单

| 角色 | 专属牌 | 数量 |
|------|--------|------|
| sealer | 五系束缚 | 1 |
| crimson_sword_spirit | 血蔷薇庭院 | 1 |
| hero | 挑衅 | 1 |
| soul_sorcerer | 灵魂链接 | 1 |
| blood_priestess | 同生共死 | 1 |
| bard | 激昂狂想曲 + 胜利交响诗 + 希望赋格曲 | 3 |

## 注意事项

1. **ID 格式**：专属牌的 ID 格式为 `starter-<playerID>-<skill_key>`，迁移时需保持一致
2. **log 回调**：迁移后的 `EnsureStarterCards` 无法直接调用 `e.Log()`，需通过参数传入 log 函数或在 Defaults 中通过接口获取
3. **Faction 来源**：卡牌的 Faction 从 `player.Character.Faction` 取值，迁移后需确保 `player.Character` 已初始化
