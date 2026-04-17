# Phase 4：拆分角色 helpers（new_roles_*_helpers.go）

## 目标

将 `engine/new_roles_*_helpers.go`（共 9 个文件）中的角色专属函数拆分到对应 `player/<role>/helpers.go`，多角色共用函数保留在 `player/shared_helpers.go`。

## 当前状况

9 个 helper 文件中混合了多个角色的函数，需要逐函数识别归属。

### 文件清单及函数归属分析

#### 1. `new_roles_shared_runtime_helpers.go` — 通用（不拆分）

所有角色共用的基础运行时函数，迁移到 `player/shared_helpers.go`：

| 函数 | 用途 | 被引用角色 |
|------|------|-----------|
| `tokenValueBounded()` | 读取并规范化 token 值 | 全部 |
| `addTokenValueBounded()` | 增减 token（带上限） | 全部 |
| `addTokenValueBoundedWithIgnoreCap()` | 增减 token（可跳过上限） | 全部 |
| `coverCardsByEffect()` | 按效果收集场上盖牌 | 多角色 |
| `coverCountByEffectAndElement()` | 统计指定效果盖牌数 | 多角色 |
| `removeFirstCoverByEffectAndElement()` | 移除首张匹配盖牌 | 多角色 |

**操作**：整体迁移到 `player/shared_helpers.go`，包名改为 `package player`。

#### 2. `new_roles_identity_forms_helpers.go` — 按角色拆分

此文件包含两大类：

**A. 角色判定函数** `isXxx(player) bool` — 全部角色共用的 `isCharacter()` 基础函数：

```go
func isCharacter(player *model.Player, charID string) bool {
    return player != nil && player.Character != nil && player.Character.ID == charID
}
```

以及 30+ 个角色判定包装函数如 `isElfArcher()`, `isSwordEmperor()` 等。

**这些判定函数应保留在 engine 中**，因为它们被 engine 的多个 hook 文件（`attack_role_hooks.go`, `damage_role_runtime_hooks.go` 等）广泛调用。或者，将 `isCharacter()` 移到 `player/` 包下，各角色在需要时直接调用 `player.IsCharacter()`。

**B. 角色形态函数** `hasXxxForm()`, `enterXxxForm()`, `leaveXxxForm()` — 每组属于特定角色：

| 形态函数 | 归属角色 |
|---------|---------|
| `hasPrayerMasterPrayerForm` | prayer_master |
| `hasValkyrieHeroicForm` / `enterValkyrieHeroicForm` / `leaveValkyrieHeroicForm` | valkyrie |
| `hasAssassinStealthForm` / `enterAssassinStealthForm` / `leaveAssassinStealthForm` | assassin |
| `hasCrimsonKnightHotBloodedForm` / `enterCrimsonKnightHotBloodedForm` / `leaveCrimsonKnightHotBloodedForm` | crimson_knight |
| `hasOnmyojiShikigamiForm` / `leaveOnmyojiShikigamiForm` | onmyoji |
| `hasBlazeWitchFlameForm` / `leaveBlazeWitchFlameForm` | blaze_witch |
| `hasHolyBowHolyGloryForm` / `enterHolyBowHolyGloryForm` / `leaveHolyBowHolyGloryForm` | holy_bow |
| `hasArbiterJudgmentForm` / `enterArbiterJudgmentForm` | arbiter |
| `hasElfArcherRitualForm` / `enterElfArcherRitualForm` / `leaveElfArcherRitualForm` | elf_archer |
| `hasMagicSwordsmanShadowForm` / `enterMagicSwordsmanShadowForm` / `leaveMagicSwordsmanShadowForm` | magic_swordsman |
| `hasWarHomunculusBurstForm` / `enterWarHomunculusBurstForm` / `leaveWarHomunculusBurstForm` | war_homunculus |
| `hasMagicLancerPhantomForm` / `leaveMagicLancerPhantomForm` | magic_lancer |
| `hasBardEternalPrisonerForm` / `enterBardEternalPrisonerForm` / `leaveBardEternalPrisonerForm` | bard |
| `hasHeroExhaustionForm` / `leaveHeroExhaustionForm` | hero |
| `hasFighterHundredDragonForm` / `leaveFighterHundredDragonForm` | fighter |
| `enterMoonGoddessDarkMoonForm` / `leaveMoonGoddessDarkMoonForm` | moon |
| `hasBloodPriestessBleedingForm` / `enterBloodPriestessBleedingFormState` / `leaveBloodPriestessBleedingFormState` | blood_priestess |

**通用形态基础设施**（保留在 shared）：

| 函数 | 说明 |
|------|------|
| `effectivePlayerOrientation()` | 读取有效朝向 |
| `effectivePlayerForm()` | 读取有效形态 |
| `playerHasForm()` | 判断是否处于某形态 |
| `setPlayerForm()` | 进入形态 |
| `clearPlayerForm()` | 退出形态 |

**常量**（各角色 cap 值）：

| 常量 | 归属角色 |
|------|---------|
| `magicBowChargeCapEngine` | magic_bow |
| `bardInspirationCapEngine` | bard |
| `holyBowFaithCapEngine` | holy_bow |
| `holyBowCannonCapEngine` | holy_bow |
| `swordEmperorSwordQiCapEngine` | sword_emperor |
| `swordEmperorSwordSoulCapEngine` | sword_emperor |
| `beastSamuraiZanshinCapEngine` | beast_samurai |
| `beastSamuraiBeastSoulCapEngine` | beast_samurai |
| `soulSorcererBlueCapEngine` | soul_sorcerer |
| `soulSorcererYellowCapEngine` | soul_sorcerer |
| `moonGoddessNewMoonCapEngine` | moon |
| `moonGoddessPetrifyCapEngine` | moon |
| `butterflyCocoonCapEngine` | butterfly_dancer |

**跨角色方法**（需保留在 engine 或提取为共享接口）：

| 函数 | 说明 | 问题 |
|------|------|------|
| `snapshotPlayerPoses()` | 快照所有玩家形态 | 访问 e.State.Players |
| `dispatchOrientationChanges()` | 分发形态变更事件 | 访问 e.State.Players + dispatcher |

这两个函数被多个角色的形态切换逻辑调用，且直接操作 `e.State.Players` 和 `e.dispatcher`，需保留在 engine 或在 ChoiceRuntime 中暴露。

#### 3. `new_roles_combat_resource_helpers.go` — 按角色拆分

| 函数 | 归属角色 |
|------|---------|
| `bardInspiration()`, `addBardInspiration()` | bard |
| `holyBowFaith()`, `addHolyBowFaith()`, `holyBowCannon()` | holy_bow |
| `swordEmperorSwordQi()`, `addSwordEmperorSwordQi()`, `swordEmperorSwordSoulCards()`, `swordEmperorSwordSoulCount()`, `syncSwordEmperorSwordSoulToken()`, `placeSwordEmperorSwordSoul()`, `takeDiscardPileCardByID()`, `clearSwordEmperorCombatTokens()`, `resolveSwordEmperorAttackMiss()`, `resolveSwordEmperorAttackHitAftermath()` | sword_emperor |
| `beastSamuraiZanshin()`, `addBeastSamuraiZanshin()`, `beastSamuraiBeastSoul()`, `addBeastSamuraiBeastSoul()`, `consumeBeastSamuraiBeastSoul()`, `beastSamuraiInIaijutsuForm()`, `enterBeastSamuraiIaijutsuForm()`, `leaveBeastSamuraiIaijutsuForm()`, `clearBeastSamuraiAttackTokens()` | beast_samurai |
| `holyBowShardMissEligibleAllies()`, `holyBowShardMissValidXValues()` | holy_bow |
| `soulSorcererBlue()`, `addSoulSorcererBlue()`, `soulSorcererYellow()`, `addSoulSorcererYellow()`, `applySoulSorcererSoulDevour()` | soul_sorcerer |
| `moonGoddessNewMoon()`, `addMoonGoddessNewMoon()`, `moonGoddessPetrify()`, `addMoonGoddessPetrify()`, `moonGoddessDarkMoonCovers()`, `moonGoddessDarkMoonCount()`, `addMoonGoddessDarkMoonCards()`, `applyMoonGoddessDarkMoonCurse()`, `removeMoonGoddessDarkMoonByFieldIndex()`, `removeMoonGoddessDarkMoonAny()`, `moonGoddessEnemyIDs()`, `moonGoddessHasElementDarkMoon()` | moon |

#### 4. 其余 helper 文件

| 文件 | 主要归属 | 复杂度 |
|------|---------|--------|
| `new_roles_butterfly_damage_helpers.go` | butterfly_dancer（专属） | 中 |
| `new_roles_card_counter_helpers.go` | 多角色混合 | 高 |
| `new_roles_field_resource_morale_helpers.go` | magic_bow, holy_lancer 等 | 高 |
| `new_roles_link_status_helpers.go` | soul_sorcerer, blood_priestess | 中 |
| `new_roles_moon_followup_helpers.go` | moon（专属） | 中 |
| `new_roles_post_resolution_helpers.go` | magic_lancer, prayer_master 等 | 高 |

## 详细步骤

### Step 4.1：创建 player/shared_helpers.go

```go
// player/shared_helpers.go
package player

import "starcup-engine/internal/model"

// EnsurePlayerTokensMap 确保玩家 tokens map 已初始化。
func EnsurePlayerTokensMap(player *model.Player) {
    if player != nil && player.Tokens == nil {
        player.Tokens = map[string]int{}
    }
}

// TokenValueBounded 读取并规范化 token 值。
func TokenValueBounded(player *model.Player, key string, cap int) int {
    if player == nil { return 0 }
    EnsurePlayerTokensMap(player)
    v := player.Tokens[key]
    if v < 0 { v = 0 }
    if cap >= 0 && v > cap { v = cap }
    player.Tokens[key] = v
    return v
}

// AddTokenValueBounded 增减 token（带上限）。
func AddTokenValueBounded(player *model.Player, key string, delta int, cap int) int {
    return AddTokenValueBoundedWithIgnoreCap(player, key, delta, cap, false)
}

// AddTokenValueBoundedWithIgnoreCap 允许跳过上限裁剪。
func AddTokenValueBoundedWithIgnoreCap(player *model.Player, key string, delta int, cap int, ignoreCap bool) int {
    if player == nil { return 0 }
    EnsurePlayerTokensMap(player)
    baseCap := cap
    if ignoreCap { baseCap = -1 }
    v := TokenValueBounded(player, key, baseCap) + delta
    if v < 0 { v = 0 }
    if !ignoreCap && cap >= 0 && v > cap { v = cap }
    player.Tokens[key] = v
    return v
}

// CoverCardsByEffect 收集指定效果的场上盖牌。
func CoverCardsByEffect(player *model.Player, effect model.EffectType) []*model.FieldCard {
    // ...
}

// IsCharacter 判断玩家是否为指定角色。
func IsCharacter(player *model.Player, charID string) bool {
    return player != nil && player.Character != nil && player.Character.ID == charID
}

// EffectivePlayerOrientation 读取有效朝向。
func EffectivePlayerOrientation(player *model.Player) model.CharacterOrientation { ... }

// EffectivePlayerForm 读取有效形态。
func EffectivePlayerForm(player *model.Player) string { ... }

// SetPlayerForm 设置形态（横置+形态名）。
func SetPlayerForm(player *model.Player, form string) bool { ... }

// ClearPlayerForm 清除形态。
func ClearPlayerForm(player *model.Player, form string) bool { ... }
```

注意：为避免循环引用，`player` 包内的 shared helpers 只能依赖 `model` 和 `types`，不能依赖 `engine`。

**重要**：由于这些函数目前被 `engine` 包大量调用，迁移后需要在 `engine` 中保留别名函数做过渡：

```go
// engine/compat_helpers.go — 过渡兼容层
package engine

import engineplayer "starcup-engine/internal/engine/player"

// deprecated: 使用 player.TokenValueBounded
func tokenValueBounded(player *model.Player, key string, cap int) int {
    return engineplayer.TokenValueBounded(player, key, cap)
}
```

### Step 4.2：创建各角色的 helpers.go

以 `bard` 为例：

```go
// player/bard/helpers.go
package bard

import (
    "starcup-engine/internal/engine/player"
    "starcup-engine/internal/model"
)

const inspirationCap = 3

// Inspiration 读取吟游诗人灵感值。
func Inspiration(p *model.Player) int {
    return player.TokenValueBounded(p, "bd_inspiration", inspirationCap)
}

// AddInspiration 增减灵感值。
func AddInspiration(p *model.Player, delta int) int {
    return player.AddTokenValueBounded(p, "bd_inspiration", delta, inspirationCap)
}

// HasEternalPrisonerForm 判断是否处于永恒囚徒形态。
func HasEternalPrisonerForm(p *model.Player) bool {
    return player.EffectivePlayerForm(p) == model.FormBardEternalPrisoner
}

// EnterEternalPrisonerForm 进入永恒囚徒形态。
func EnterEternalPrisonerForm(p *model.Player) bool {
    return player.SetPlayerForm(p, model.FormBardEternalPrisoner)
}

// LeaveEternalPrisonerForm 退出永恒囚徒形态。
func LeaveEternalPrisonerForm(p *model.Player) bool {
    return player.ClearPlayerForm(p, model.FormBardEternalPrisoner)
}
```

以 `sword_emperor` 为例（含 GameEngine 方法的函数需特殊处理）：

```go
// player/sword_emperor/helpers.go
package sword_emperor

import (
    "starcup-engine/internal/engine/player"
    "starcup-engine/internal/model"
)

const swordQiCap = 5
const swordSoulCap = 3

// SwordQi 读取剑气值。
func SwordQi(p *model.Player) int {
    return player.TokenValueBounded(p, "se_sword_qi", swordQiCap)
}

// AddSwordQi 增减剑气值。
func AddSwordQi(p *model.Player, delta int) int {
    return player.AddTokenValueBounded(p, "se_sword_qi", delta, swordQiCap)
}

// SwordSoulCount 统计剑魂数量。
func SwordSoulCount(p *model.Player) int {
    return len(player.CoverCardsByEffect(p, model.EffectSwordSoul))
}

// ClearCombatTokens 清除战斗临时 token。
func ClearCombatTokens(p *model.Player) {
    if p == nil { return }
    player.EnsurePlayerTokensMap(p)
    // 注意：TurnState.UsedSkillCounts 的操作可能需要额外接口支持
    p.TurnState.UsedSkillCounts["se_guard_disabled_current_attack"] = 0
    p.TurnState.UsedSkillCounts["se_angel_soul_armed"] = 0
    p.TurnState.UsedSkillCounts["se_demon_soul_armed"] = 0
}
```

### Step 4.3：处理需要 GameEngine 方法的函数

部分 helper 函数是 `(e *GameEngine)` 方法（如 `resolveSwordEmperorAttackMiss`），它们直接操作 `e.State` 和 `e.Log` 等。

**策略**：
- **纯数据读写**的函数（如 `SwordQi()`, `AddSwordQi()`）→ 直接移到角色包
- **需要引擎状态**的函数（如 `resolveSwordEmperorAttackMiss()`）→ 保留在 engine，但改为调用角色包中的纯数据函数

```go
// engine 中保留的桥接代码
func (e *GameEngine) resolveSwordEmperorAttackMiss(attackerID string, attackCard *model.Card, isCounter bool) {
    attacker := e.State.Players[attackerID]
    if attacker == nil || !isCharacter(attacker, "sword_emperor") || isCounter {
        return
    }
    // 调用 player/sword_emperor 包中的纯数据函数
    if attacker.TurnState.UsedSkillCounts["se_guard_disabled_current_attack"] <= 0 &&
        sword_emperor.SwordSoulCount(attacker) < sword_emperor.SwordSoulCap() &&
        attackCard != nil {
        if card, ok := e.takeDiscardPileCardByID(attackCard.ID); ok {
            // 放置剑魂仍需引擎操作
            attacker.AddFieldCard(&model.FieldCard{...})
            sword_emperor.SyncSwordSoulToken(attacker)
            e.Log(...)
        }
    }
}
```

### Step 4.4：更新 import 链

迁移后，engine 中的代码需要 import 角色包来调用 helpers：

```go
// engine/new_roles_combat_resource_helpers.go — 修改后（过渡期）
package engine

import (
    "starcup-engine/internal/engine/player/bard"
    "starcup-engine/internal/engine/player/holy_bow"
    "starcup-engine/internal/engine/player/sword_emperor"
    // ...
)

// 旧函数改为调用新位置的别名
func bardInspiration(player *model.Player) int {
    return bard.Inspiration(player)
}
```

### Step 4.5：逐步删除旧文件

每个角色的 helpers 迁移完成后，从原文件中删除对应函数。当原文件中的所有函数都迁移完毕后，删除原文件。

最终 `engine/` 中不再有 `new_roles_*_helpers.go` 文件。

### Step 4.6：验证

```bash
go build ./...
go test ./internal/engine/... -count=1
go test ./internal/engine/player/... -count=1
```

## 拆分优先级

建议按以下顺序拆分（从独立到复杂）：

1. `new_roles_shared_runtime_helpers.go` → `player/shared_helpers.go`（最高优先级，被所有角色依赖）
2. `new_roles_butterfly_damage_helpers.go` → `player/butterfly_dancer/helpers.go`（独立文件，单角色）
3. `new_roles_moon_followup_helpers.go` → `player/moon/helpers.go`（独立文件，单角色）
4. `new_roles_identity_forms_helpers.go` → 各角色 `helpers.go` + `player/shared_helpers.go`
5. `new_roles_combat_resource_helpers.go` → 各角色 `helpers.go`
6. `new_roles_card_counter_helpers.go` → 各角色 `helpers.go`
7. `new_roles_field_resource_morale_helpers.go` → 各角色 `helpers.go`
8. `new_roles_link_status_helpers.go` → 各角色 `helpers.go`
9. `new_roles_post_resolution_helpers.go` → 各角色 `helpers.go`

## 风险

1. **循环引用**：角色包中的 helpers 不能调用 `(e *GameEngine)` 方法，只能操作纯数据或通过 `player` 包的共享函数
2. **engine 中的别名过渡**：迁移期间需在 engine 中保留别名函数，确保编译通过，全部迁移后再清理
3. **常量 cap 值**：当前以 `XxxCapEngine` 命名的常量需移到各角色包中并重命名（去掉 `Engine` 后缀）
4. **跨角色引用**：如 `coverCardsByEffect()` 被多个角色使用，必须保留在 `player/shared_helpers.go` 中
