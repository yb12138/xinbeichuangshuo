# Phase 4：拆分角色 helpers（new_roles_*_helpers.go）— 修订版

## 核心原则：按数据域分层

当前 helpers 混合了四种本质不同的操作，迁移前必须先按数据域分层：

| 数据域 | 底层数据 | 操作特点 | 迁移策略 |
|--------|---------|---------|---------|
| **Token（指示物）** | `player.Tokens map[string]int` | 纯数值读写，带 cap | 直接迁移到角色包 |
| **Form（形态）** | `player.Orientation` + `player.Form` | 状态切换 + 时序事件 | 基础操作迁移，快照/分发留在 engine |
| **FieldCard（场牌）** | `player.Field []*FieldCard` | 结构化对象管理 | 单角色操作迁移，跨角色操作留在 engine |
| **复杂业务** | `e.State.*`, `e.Log`, `e.PushInterrupt` | 多角色协作，引擎状态依赖 | 保留在 engine |

**关键约束：TokenAccessor 只管指示物读写，不碰形态、场牌、引擎状态。**

---

## 当前状况

9 个 helper 文件，共 2,707 行。约 60 个独立函数（仅依赖 `*model.Player`），约 55 个 GameEngine 方法。

### 按数据域分类

#### A. Token 读写类（~30 个函数）

纯 `player.Tokens` 操作，可直接迁移。

| Token 键 | 角色 | Cap | 函数 |
|----------|------|-----|------|
| `bd_inspiration` | bard | 3 | `bardInspiration`, `addBardInspiration` |
| `hb_faith` | holy_bow | 10 | `holyBowFaith`, `addHolyBowFaith` |
| `hb_cannon` | holy_bow | 1 | `holyBowCannon` |
| `se_sword_qi` | sword_emperor | 5 | `swordEmperorSwordQi`, `addSwordEmperorSwordQi` |
| `bs_zanshin` | beast_samurai | 4 | `beastSamuraiZanshin`, `addBeastSamuraiZanshin` |
| `bs_beast_soul` | beast_samurai | 2 | `beastSamuraiBeastSoul`, `addBeastSamuraiBeastSoul`, `consumeBeastSamuraiBeastSoul` |
| `ss_blue_soul` | soul_sorcerer | 6 | `soulSorcererBlue`, `addSoulSorcererBlue` |
| `ss_yellow_soul` | soul_sorcerer | 6 | `soulSorcererYellow`, `addSoulSorcererYellow` |
| `mg_new_moon` | moon | 2 | `moonGoddessNewMoon`, `addMoonGoddessNewMoon` |
| `mg_petrify` | moon | 3 | `moonGoddessPetrify`, `addMoonGoddessPetrify` |
| `bt_pupa` | butterfly_dancer | 无上限 | `butterflyPupa`, `addButterflyPupa` |
| `css_blood` | crimson_sword_spirit | 3 | `bloodLimit`, `addBlood` |

**共享基础设施**（保留在 `player/` 包）：

| 函数 | 用途 |
|------|------|
| `tokenValueBounded` | 读取 token，裁剪到 [0, cap] |
| `addTokenValueBounded` | 增减 token，带 cap |
| `addTokenValueBoundedWithIgnoreCap` | 可跳过上限 |

#### B. 形态管理类（17 组 has/enter/leave + 基础设施）

操作 `player.Orientation`（Normal/Tapped）和 `player.Form`（string 常量）。

**基础操作**（保留在 `player/` 包）：

| 函数 | 用途 |
|------|------|
| `effectivePlayerOrientation` | 读取有效朝向 |
| `effectivePlayerForm` | 读取有效形态 |
| `playerHasForm` | 判断是否处于某形态 |
| `setPlayerForm` | 设置 Orientation=Tapped + Form=xxx |
| `clearPlayerForm` | 恢复 Orientation=Normal + Form="" |

**角色专属形态三元组**（每组 1~3 个函数，迁移到对应角色包）：

| 角色 | 形态 | 函数 |
|------|------|------|
| assassin | Stealth | has/enter/leave |
| arbiter | Judgment | has/enter |
| elf_archer | Ritual | has/enter/leave |
| magic_swordsman | Shadow | has/enter/leave |
| war_homunculus | Burst | has/enter/leave |
| prayer_master | Prayer | has |
| crimson_knight | HotBlooded | has/enter/leave |
| onmyoji | Shikigami | has/leave |
| blaze_witch | Flame | has/leave |
| beast_samurai | Iaijutsu | enter/leave (has 通过 playerHasForm) |
| holy_bow | HolyGlory | has/enter/leave |
| magic_lancer | Phantom | has/leave |
| bard | EternalPrisoner | has/enter/leave |
| hero | Exhaustion | has/leave |
| fighter | HundredDragon | has/leave |
| moon | DarkMoon | enter/leave |
| blood_priestess | Bleeding | has/enter/leave |
| valkyrie | Heroic | has/enter/leave |

**引擎级形态基础设施**（保留在 engine）：

| 函数 | 原因 |
|------|------|
| `snapshotPlayerPoses` | 遍历 e.State.Players |
| `dispatchOrientationChanges` | 触发 e.dispatcher 时序钩子 |

#### C. FieldCard 管理类（~20 个函数）

操作 `player.Field []*FieldCard`，分两个子类：

**C1. 单角色场牌资源**（Mode=FieldCover，仅操作自己的 Field）：

| 角色 | EffectType | 函数 |
|------|-----------|------|
| sword_emperor | SwordSoul | `swordEmperorSwordSoulCards`, `swordEmperorSwordSoulCount`, `syncSwordEmperorSwordSoulToken` |
| moon | MoonDarkMoon | `moonGoddessDarkMoonCovers`, `moonGoddessDarkMoonCount`, `addMoonGoddessDarkMoonCards` |
| magic_bow | MagicBowCharge | `magicBowChargeCount`, `syncMagicBowChargeToken`, `addMagicBowChargeCards`, `removeMagicBowChargeByElement` |
| spirit_caster | SpiritCasterPower | `spiritCasterPowerCovers`, `spiritCasterPowerCount`, `syncSpiritCasterPowerToken`, `addSpiritCasterPowerCard` |
| butterfly_dancer | ButterflyCocoon | `butterflyCocoonCovers`, `butterflyCocoonCount`, `syncButterflyCocoonToken`, `addButterflyCocoonCards`, `butterflyCocoonFieldIndices`, `removeButterflyCocoonByFieldIndex`, `removeButterflyCocoonByFieldIndices`, `butterflyMirrorPairDefs` |
| elf_archer | ElfBlessing | `markElfBlessings`, `syncElfBlessings`, `countElfBlessings`, `isElfBlessingCard`, `removeElfBlessingByCardID`, `elfBlessingHandIndices`, `elfBlessingCoverCards`, `elfBlessingCards` |

**共享场牌基础设施**（保留在 `player/` 包）：

| 函数 | 用途 |
|------|------|
| `coverCardsByEffect` | 按效果收集场上盖牌 |
| `coverCountByEffectAndElement` | 统计指定效果+元素盖牌数 |
| `removeFirstCoverByEffectAndElement` | 移除首张匹配盖牌 |

**C2. 跨角色场牌效果**（Mode=FieldEffect，操作别人的 Field，需 GameEngine）：

| 角色 | EffectType | 函数 | 保留在 engine 的原因 |
|------|-----------|------|---------------------|
| bard | BardEternalMovement | `findBardEternalMovement`, `bardEternalHolderID`, `removeBardEternalMovement`, `placeBardEternalMovement`, `placeBardEternalMovementWithCard`, `transferBardEternalMovement` | 遍历所有玩家 Field |
| soul_sorcerer | SoulLink | `findSoulLink`, `placeSoulLink` | 跨玩家绑定 |
| blood_priestess | BloodSharedLife | `findBloodPriestessSharedLife`, `detachBloodPriestessSharedLife`, `removeBloodPriestessSharedLife`, `placeBloodPriestessSharedLife`, `hasFixedMaxHandCap`, `bloodPriestessSharedLifeDeltaFor` | 跨玩家效果 |

#### D. 复杂业务类（~25 个 GameEngine 方法）

深度依赖引擎状态，暂时全部保留在 engine。

| 函数 | 依赖的引擎状态 |
|------|--------------|
| `resolveSwordEmperorAttackMiss` | DiscardPile, Log, Heal, morale |
| `resolveSwordEmperorAttackHitAftermath` | Log, Heal |
| `handlePostAttackHitEffects` | 多角色形态检查, morale, dispatcher |
| `handlePostActionEndEffects` | 多角色状态, PushInterrupt |
| `handlePostDamageResolved` | 多角色, dispatcher, Log |
| `applySoulSorcererSoulDevour` | GetAllPlayers, Log |
| `applyMoonGoddessDarkMoonCurse` | snapshot, morale, dispatch, checkGameEnd |
| `maybeButterflyDamageResponses` | PushInterrupt |
| `maybeMoonGoddessMedusa` | PushInterrupt |
| `maybeMoonGoddessMoonCycleAtTurnEnd` | PushInterrupt |
| `tryQueueMoonGoddessBlasphemy` | PushInterrupt |
| `queueButterflyWitherFollowup` | PushInterrupt |
| `applyCampMoraleLoss` / `addCampMorale` | e.State.RedMorale/BlueMorale |
| `campMorale` | e.State.RedMorale/BlueMorale |
| `moraleFloorForCamp` | 遍历所有玩家 |
| `canCastMagicInAction` | 多角色形态检查 |
| `clearFighterHundredDragon` | snapshot, dispatch, Log |
| `queueElfAnimalResponse` | PushInterrupt |

---

## 迁移方案

### Step 4.1：创建共享基础设施

#### `player/shared_helpers.go` — Token 基础设施

```go
package player

import "starcup-engine/internal/model"

// EnsurePlayerTokensMap 确保 player.Tokens map 已初始化。
func EnsurePlayerTokensMap(p *model.Player) {
    if p != nil && p.Tokens == nil {
        p.Tokens = map[string]int{}
    }
}

// TokenValue 读取 token 值，裁剪到 [0, cap]（cap < 0 表示无上限）。
func TokenValue(p *model.Player, key string, cap int) int {
    if p == nil {
        return 0
    }
    EnsurePlayerTokensMap(p)
    v := p.Tokens[key]
    if v < 0 {
        v = 0
    }
    if cap >= 0 && v > cap {
        v = cap
        p.Tokens[key] = v
    }
    return v
}

// AddToken 增减 token 值，返回新值。
func AddToken(p *model.Player, key string, delta int, cap int) int {
    return AddTokenIgnoreCap(p, key, delta, cap, false)
}

// AddTokenIgnoreCap 增减 token 值，可跳过上限裁剪。
func AddTokenIgnoreCap(p *model.Player, key string, delta int, cap int, ignoreCap bool) int {
    if p == nil {
        return 0
    }
    EnsurePlayerTokensMap(p)
    effectiveCap := cap
    if ignoreCap {
        effectiveCap = -1
    }
    v := TokenValue(p, key, effectiveCap) + delta
    if v < 0 {
        v = 0
    }
    if !ignoreCap && cap >= 0 && v > cap {
        v = cap
    }
    p.Tokens[key] = v
    return v
}
```

#### `player/form_helpers.go` — 形态基础设施

```go
package player

import "starcup-engine/internal/model"

// EffectiveOrientation 读取有效朝向。
func EffectiveOrientation(p *model.Player) model.CharacterOrientation {
    if p == nil {
        return model.OrientationNormal
    }
    o := p.Orientation
    if o == "" {
        return model.OrientationNormal
    }
    return o
}

// EffectiveForm 读取有效形态。
func EffectiveForm(p *model.Player) string {
    if p == nil {
        return ""
    }
    return p.Form
}

// HasForm 判断是否处于指定形态。
func HasForm(p *model.Player, form string) bool {
    return EffectiveForm(p) == form
}

// SetForm 进入形态（横置 + 设置形态名）。
func SetForm(p *model.Player, form string) bool {
    if p == nil {
        return false
    }
    changed := EffectiveOrientation(p) != model.OrientationTapped || EffectiveForm(p) != form
    p.Orientation = model.OrientationTapped
    p.Form = form
    return changed
}

// ClearForm 退出形态（恢复直立 + 清空形态名）。
func ClearForm(p *model.Player) bool {
    if p == nil {
        return false
    }
    changed := EffectiveOrientation(p) != model.OrientationNormal || EffectiveForm(p) != ""
    p.Orientation = model.OrientationNormal
    p.Form = ""
    return changed
}
```

#### `player/field_helpers.go` — FieldCard 基础设施

```go
package player

import "starcup-engine/internal/model"

// CoverCardsByEffect 按效果收集场上盖牌。
func CoverCardsByEffect(p *model.Player, effect model.EffectType) []*model.FieldCard {
    if p == nil {
        return nil
    }
    return p.GetCoverCardsByEffect(effect)
}

// CoverCountByEffect 统计指定效果盖牌数。
func CoverCountByEffect(p *model.Player, effect model.EffectType) int {
    return len(CoverCardsByEffect(p, effect))
}

// RemoveFirstCoverByEffect 移除首张匹配效果的盖牌。
func RemoveFirstCoverByEffect(p *model.Player, effect model.EffectType) (*model.FieldCard, bool) {
    cards := CoverCardsByEffect(p, effect)
    if len(cards) == 0 {
        return nil, false
    }
    fc := cards[0]
    p.RemoveFieldCard(fc)
    return fc, true
}
```

#### `player/identity_helpers.go` — 角色判定基础设施

```go
package player

import "starcup-engine/internal/model"

// IsCharacter 判断玩家是否为指定角色。
func IsCharacter(p *model.Player, charID string) bool {
    return p != nil && p.Character != nil && p.Character.ID == charID
}
```

### Step 4.2：迁移各角色的 Token helpers

每个角色的 Token 读写函数迁移到 `player/<role>/token.go`。

#### soul_sorcerer/token.go

```go
package soul_sorcerer

import (
    "starcup-engine/internal/engine/player"
    "starcup-engine/internal/model"
)

const (
    BlueSoulCap  = 6
    YellowSoulCap = 6
)

func BlueSoul(p *model.Player) int {
    return player.TokenValue(p, "ss_blue_soul", BlueSoulCap)
}

func AddBlueSoul(p *model.Player, delta int) int {
    return player.AddToken(p, "ss_blue_soul", delta, BlueSoulCap)
}

func YellowSoul(p *model.Player) int {
    return player.TokenValue(p, "ss_yellow_soul", YellowSoulCap)
}

func AddYellowSoul(p *model.Player, delta int) int {
    return player.AddToken(p, "ss_yellow_soul", delta, YellowSoulCap)
}
```

#### beast_samurai/token.go

```go
package beast_samurai

import (
    "starcup-engine/internal/engine/player"
    "starcup-engine/internal/model"
)

const (
    ZanshinCap   = 4
    BeastSoulCap = 2
)

func Zanshin(p *model.Player) int {
    return player.TokenValue(p, "bs_zanshin", ZanshinCap)
}

func AddZanshin(p *model.Player, delta int) int {
    return player.AddToken(p, "bs_zanshin", delta, ZanshinCap)
}

func BeastSoul(p *model.Player) int {
    return player.TokenValue(p, "bs_beast_soul", BeastSoulCap)
}

func AddBeastSoul(p *model.Player, delta int, ignoreCap bool) int {
    return player.AddTokenIgnoreCap(p, "bs_beast_soul", delta, BeastSoulCap, ignoreCap)
}

func ConsumeBeastSoul(p *model.Player) int {
    current := BeastSoul(p)
    if current <= 0 {
        return 0
    }
    AddBeastSoul(p, -1, false)
    AddZanshin(p, 1)
    return 1
}
```

其他角色同理（bard, holy_bow, sword_emperor, moon, butterfly_dancer, crimson_sword_spirit），每个角色一个 `token.go`。

### Step 4.3：迁移各角色的 Form helpers

每个角色的形态管理函数迁移到 `player/<role>/form.go`。

#### beast_samurai/form.go

```go
package beast_samurai

import (
    "starcup-engine/internal/engine/player"
    "starcup-engine/internal/model"
)

func InIaijutsuForm(p *model.Player) bool {
    return player.HasForm(p, model.FormBeastSamuraiIaijutsu)
}

func EnterIaijutsuForm(p *model.Player) bool {
    return player.SetForm(p, model.FormBeastSamuraiIaijutsu)
}

func LeaveIaijutsuForm(p *model.Player) bool {
    return player.ClearForm(p)
}
```

#### assassin/form.go

```go
package assassin

import (
    "starcup-engine/internal/engine/player"
    "starcup-engine/internal/model"
)

func InStealthForm(p *model.Player) bool {
    return player.HasForm(p, model.FormAssassinStealth)
}

func EnterStealthForm(p *model.Player) bool {
    return player.SetForm(p, model.FormAssassinStealth)
}

func LeaveStealthForm(p *model.Player) bool {
    return player.ClearForm(p)
}
```

17 个角色同理，每个角色一个 `form.go`。

**注意**：形态切换需要触发 `TimingOnOrientationChanged` 时序钩子的情况，由 engine 层通过 `snapshotPlayerPoses` / `dispatchOrientationChanges` 包裹调用。角色包中的 `EnterForm` / `LeaveForm` 只做纯数据操作，不触发时序事件。

### Step 4.4：迁移各角色的 FieldCard helpers

每个角色的单角色场牌管理函数迁移到 `player/<role>/field.go`。

#### sword_emperor/field.go

```go
package sword_emperor

import (
    "starcup-engine/internal/engine/player"
    "starcup-engine/internal/model"
)

const SwordSoulCap = 3

func SwordSoulCards(p *model.Player) []*model.FieldCard {
    return player.CoverCardsByEffect(p, model.EffectSwordSoul)
}

func SwordSoulCount(p *model.Player) int {
    return player.CoverCountByEffect(p, model.EffectSwordSoul)
}

func SyncSwordSoulToken(p *model.Player) {
    // 同步 token 计数 = FieldCard 计数
    p.Tokens["se_sword_soul_count"] = SwordSoulCount(p)
}

func ClearCombatTokens(p *model.Player) {
    if p == nil {
        return
    }
    player.EnsurePlayerTokensMap(p)
    p.TurnState.UsedSkillCounts["se_guard_disabled_current_attack"] = 0
    p.TurnState.UsedSkillCounts["se_angel_soul_armed"] = 0
    p.TurnState.UsedSkillCounts["se_demon_soul_armed"] = 0
}
```

#### butterfly_dancer/field.go

```go
package butterfly_dancer

import (
    "starcup-engine/internal/engine/player"
    "starcup-engine/internal/model"
)

func CocoonCovers(p *model.Player) []*model.FieldCard {
    return player.CoverCardsByEffect(p, model.EffectButterflyCocoon)
}

func CocoonCount(p *model.Player) int {
    return player.CoverCountByEffect(p, model.EffectButterflyCocoon)
}

func SyncCocoonToken(p *model.Player) {
    p.Tokens["bt_cocoon_count"] = CocoonCount(p)
}

func AddCocoonCards(p *model.Player, cards []model.Card) {
    for _, c := range cards {
        p.AddFieldCard(&model.FieldCard{
            Card:    c,
            OwnerID: p.ID,
            SourceID: p.ID,
            Mode:    model.FieldCover,
            Effect:  model.EffectButterflyCocoon,
            Hook:    model.FieldHookManual,
        })
    }
    SyncCocoonToken(p)
}

func CocoonFieldIndices(p *model.Player) []int { /* ... */ }
func RemoveCocoonByFieldIndex(p *model.Player, idx int) (*model.FieldCard, bool) { /* ... */ }
func RemoveCocoonByFieldIndices(p *model.Player, indices []int) { /* ... */ }
func MirrorPairDefs(p *model.Player) [][2]int { /* ... */ }
```

其他角色（moon, magic_bow, spirit_caster, elf_archer）同理。

### Step 4.5：处理 engine 中的过渡层

迁移期间，engine 中需要保留别名函数确保编译通过。

#### `engine/compat_token_helpers.go` — Token 别名

```go
package engine

import (
    engineplayer "starcup-engine/internal/engine/player"
    ss "starcup-engine/internal/engine/player/soul_sorcerer"
    bs "starcup-engine/internal/engine/player/beast_samurai"
    // ...
)

// deprecated: 使用 player.TokenValue
func tokenValueBounded(p *model.Player, key string, cap int) int {
    return engineplayer.TokenValue(p, key, cap)
}

// deprecated: 使用 ss.BlueSoul
func soulSorcererBlue(p *model.Player) int {
    return ss.BlueSoul(p)
}

// deprecated: 使用 bs.Zanshin
func beastSamuraiZanshin(p *model.Player) int {
    return bs.Zanshin(p)
}
```

#### `engine/compat_form_helpers.go` — 形态别名

```go
package engine

import (
    engineplayer "starcup-engine/internal/engine/player"
    assassin "starcup-engine/internal/engine/player/assassin"
    // ...
)

// deprecated: 使用 player.HasForm
func playerHasForm(p *model.Player, form string) bool {
    return engineplayer.HasForm(p, form)
}

// deprecated: 使用 assassin.InStealthForm
func hasAssassinStealthForm(p *model.Player) bool {
    return assassin.InStealthForm(p)
}
```

### Step 4.6：保留在 engine 的复杂业务逻辑

这些函数深度依赖引擎状态，不做迁移。但在 engine 内部改为调用角色包的纯数据函数。

```go
// engine/new_roles_post_resolution_helpers.go — 保留

func (e *GameEngine) resolveSwordEmperorAttackMiss(attackerID string, attackCard *model.Card, isCounter bool) {
    attacker := e.State.Players[attackerID]
    if attacker == nil || !isCharacter(attacker, "sword_emperor") || isCounter {
        return
    }
    // 调用角色包的纯数据函数
    if attacker.TurnState.UsedSkillCounts["se_guard_disabled_current_attack"] <= 0 &&
        se.SwordSoulCount(attacker) < se.SwordSoulCap &&
        attackCard != nil {
        // ... 引擎操作（弃牌堆、日志、士气等）
    }
}
```

```go
// engine/new_roles_link_status_helpers.go — 跨角色效果保留

func (e *GameEngine) placeBardEternalMovement(bardPlayer, targetPlayer *model.Player, card model.Card) {
    // 跨玩家操作，保留在 engine
    e.attachSourceEffectCard(bardPlayer, targetPlayer, model.EffectBardEternalMovement, card)
    e.Log(...)
}
```

### Step 4.7：清理

当所有别名调用方都迁移完成后：
1. 删除 `engine/compat_token_helpers.go`
2. 删除 `engine/compat_form_helpers.go`
3. 删除已清空的 `engine/new_roles_*_helpers.go` 文件
4. 保留 `engine/shared_runtime_helpers.go`（含 `snapshotPlayerPoses`、`dispatchOrientationChanges`）

---

## 文件结构总览

迁移完成后的文件分布：

```
internal/engine/player/
├── shared_helpers.go      # Token 基础: TokenValue, AddToken
├── form_helpers.go         # 形态基础: HasForm, SetForm, ClearForm
├── field_helpers.go        # FieldCard 基础: CoverCardsByEffect, CoverCountByEffect
├── identity_helpers.go     # 角色判定: IsCharacter
├── registry.go
├── role_entry.go
├── hand_limit_rule.go
│
├── soul_sorcerer/
│   ├── module.go
│   ├── choices.go
│   ├── token.go            # BlueSoul, AddBlueSoul, YellowSoul, AddYellowSoul
│   └── field.go            # (无，FieldCard 操作在 link_status 保留在 engine)
│
├── beast_samurai/
│   ├── module.go
│   ├── choices.go
│   ├── token.go            # Zanshin, BeastSoul, ConsumeBeastSoul
│   └── form.go             # InIaijutsuForm, EnterIaijutsuForm, LeaveIaijutsuForm
│
├── sword_emperor/
│   ├── module.go
│   ├── choices.go
│   ├── token.go            # SwordQi, AddSwordQi
│   ├── form.go             # (无专属形态)
│   └── field.go            # SwordSoulCards, SwordSoulCount, SyncSwordSoulToken
│
├── moon/
│   ├── module.go
│   ├── choices.go
│   ├── token.go            # NewMoon, AddNewMoon, Petrify, AddPetrify
│   ├── form.go             # EnterDarkMoonForm, LeaveDarkMoonForm
│   └── field.go            # DarkMoonCovers, DarkMoonCount, AddDarkMoonCards
│
├── butterfly_dancer/
│   ├── module.go
│   ├── choices.go
│   ├── token.go            # Pupa, AddPupa
│   └── field.go            # CocoonCovers, CocoonCount, AddCocoonCards, MirrorPairDefs
│
├── bard/
│   ├── module.go
│   ├── choices.go
│   ├── token.go            # Inspiration, AddInspiration
│   └── form.go             # HasEternalPrisonerForm, EnterEternalPrisonerForm, LeaveEternalPrisonerForm
│
├── holy_bow/
│   ├── module.go
│   ├── choices.go
│   ├── token.go            # Faith, AddFaith, Cannon
│   └── form.go             # HasHolyGloryForm, EnterHolyGloryForm, LeaveHolyGloryForm
│
├── elf_archer/
│   ├── module.go
│   ├── choices.go
│   ├── form.go             # InRitualForm, EnterRitualForm, LeaveRitualForm
│   └── field.go            # ElfBlessing 系列函数
│
├── (其他 14 个角色)/
│   ├── module.go
│   ├── choices.go
│   └── form.go             # 各自的形态管理（如有）
│
internal/engine/                            # 保留
├── shared_runtime_helpers.go               # snapshotPlayerPoses, dispatchOrientationChanges
├── new_roles_link_status_helpers.go        # 跨角色效果（不迁移）
├── new_roles_post_resolution_helpers.go    # 复杂业务（不迁移）
├── new_roles_moon_followup_helpers.go      # 月神后续（不迁移）
├── new_roles_butterfly_damage_helpers.go   # 蝶舞者伤害（不迁移）
├── exclusive_effect_runtime.go             # 跨角色场牌框架（不迁移）
```

---

## 执行顺序

### Round 1：基础设施（无风险，无破坏性）

1. 创建 `player/shared_helpers.go`（TokenValue, AddToken）
2. 创建 `player/form_helpers.go`（HasForm, SetForm, ClearForm）
3. 创建 `player/field_helpers.go`（CoverCardsByEffect, CoverCountByEffect）
4. 创建 `player/identity_helpers.go`（IsCharacter）
5. 在 engine 中创建别名过渡层 `compat_*.go`
6. 验证：`go build ./...`

### Round 2：Token 迁移（低风险）

每个角色独立进行，按角色逐个迁移：

1. `soul_sorcerer/token.go`
2. `beast_samurai/token.go`
3. `bard/token.go`
4. `sword_emperor/token.go`
5. `holy_bow/token.go`
6. `moon/token.go`
7. `butterfly_dancer/token.go`
8. `crimson_sword_spirit/token.go`

每迁移一个角色后：`go build ./...` + `go test ./internal/engine/... -count=1`

### Round 3：Form 迁移（低风险）

17 个角色的形态函数，每个独立进行：

1. `assassin/form.go`
2. `beast_samurai/form.go`
3. `bard/form.go`
4. `moon/form.go`
5. ... (其余 13 个角色)

每迁移一个角色后验证。

### Round 4：FieldCard 迁移（中等风险）

单角色场牌管理函数迁移：

1. `sword_emperor/field.go`
2. `moon/field.go`
3. `butterfly_dancer/field.go`
4. `magic_bow/field.go`
5. `spirit_caster/field.go`
6. `elf_archer/field.go`

每迁移一个角色后验证。注意 `sync*Token` 函数一并迁移。

### Round 5：清理过渡代码

1. 逐步删除 engine 中的别名函数
2. 删除已清空的 `new_roles_*_helpers.go`
3. 最终验证：全量测试通过

---

## 不迁移的文件（保留在 engine）

| 文件 | 原因 |
|------|------|
| `new_roles_link_status_helpers.go` | 跨角色效果，操作其他玩家的 Field |
| `new_roles_post_resolution_helpers.go` | 复杂业务，多角色协作 |
| `new_roles_moon_followup_helpers.go` | PushInterrupt 依赖 |
| `new_roles_butterfly_damage_helpers.go` | PushInterrupt 依赖 |
| `exclusive_effect_runtime.go` | 跨角色场牌框架 |

这些文件中的纯数据调用会逐步改为调用角色包函数，但文件本身留在 engine。

---

## 迁移进度追踪

### Round 1：基础设施
- [ ] `player/shared_helpers.go`
- [ ] `player/form_helpers.go`
- [ ] `player/field_helpers.go`
- [ ] `player/identity_helpers.go`
- [ ] engine 过渡别名
- [ ] 编译通过

### Round 2：Token 迁移
- [ ] soul_sorcerer
- [ ] beast_samurai
- [ ] bard
- [ ] sword_emperor
- [ ] holy_bow
- [ ] moon
- [ ] butterfly_dancer
- [ ] crimson_sword_spirit

### Round 3：Form 迁移
- [ ] assassin
- [ ] arbiter
- [ ] beast_samurai
- [ ] bard
- [ ] blaze_witch
- [ ] blood_priestess
- [ ] crimson_knight
- [ ] elf_archer
- [ ] fighter
- [ ] hero
- [ ] holy_bow
- [ ] magic_lancer
- [ ] magic_swordsman
- [ ] moon
- [ ] onmyoji
- [ ] prayer_master
- [ ] valkyrie
- [ ] war_homunculus

### Round 4：FieldCard 迁移
- [ ] sword_emperor
- [ ] moon
- [ ] butterfly_dancer
- [ ] magic_bow
- [ ] spirit_caster
- [ ] elf_archer

### Round 5：清理
- [ ] 删除 engine 别名函数
- [ ] 删除已清空的 helpers 文件
- [ ] 全量测试通过
