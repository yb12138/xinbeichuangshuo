# 重构无法行动判断逻辑

## Context

用户希望重构 `handleActionSelectionCannotAct` 函数，将无法行动的判断逻辑从核心流程中提取出来：

1. **通用无法行动判断** - 针对主行动回合，默认每个角色都会执行
2. **额外行动无法行动** - 直接跳过该额外行动，不进入核心流程
3. **角色自定义hook** - 特殊角色可以实现自己的无法行动判断函数

### 当前问题
- 无法行动判断逻辑硬编码在 `handleActionSelectionCannotAct` 中（约75行）
- 魔剑士特殊规则（全法术牌重摸）直接写在核心函数里
- 没有通用机制让其他角色自定义无法行动判断
- 额外行动的跳过逻辑和主行动的无法行动逻辑混在一起

### 通用无法行动判断场景
- 手牌全是无法使用的牌（如暗影抗拒下全是法术牌）
- 战绩区没有星石可提取
- 角色没有可发动的行动技能
- 其他角色特殊条件

---

## Design

### 1. 在 RoleEntry 中添加 CannotActChecker hook

在 `internal/engine/player/role_entry.go` 中：

```go
// CannotActChecker 无法行动判断函数
// 返回 (canCannotAct, reason)：
//   - canCannotAct=true: 该角色认为当前可以宣告无法行动
//   - canCannotAct=false: 该角色认为当前不能宣告无法行动（仍有可执行动作）
// player 参数已包含手牌、形态等信息，大多数角色只需要检查 player 即可
type CannotActChecker func(player *model.Player) (bool, string)

// RoleEntry 添加字段
type RoleEntry struct {
    // ... existing fields ...
    CannotActChecker CannotActChecker // 角色自定义无法行动判断hook（可选）
}
```

**说明**：
- `CannotActChecker` 只需要 `player` 参数，因为手牌、形态、Token 等信息都在 player 里
- 如果角色需要更复杂的引擎能力（如检查全场状态），可以后续扩展
- 大多数角色的自定义规则很简单，只需要检查 player 状态

### 2. 提取通用无法行动判断逻辑

创建新文件 `internal/engine/cannot_act_checker.go`：

```go
// DefaultCannotActChecker 默认无法行动判断
// 检查玩家是否有可用的攻击牌或法术牌
func DefaultCannotActChecker(e *GameEngine, player *model.Player) (bool, string) {
    // 有手牌时检查是否有可执行动作
    if len(player.Hand) > 0 {
        canUseMagic := e.canCastMagicInAction(player)
        for idx := 0; idx < playableCardCount(player); idx++ {
            card, _, _, ok := getPlayableCardByIndex(player, idx)
            if !ok {
                continue
            }
            // 有攻击牌可用 -> 不能宣告无法行动
            if card.Type == model.CardTypeAttack {
                return false, "有攻击牌可用"
            }
            // 有法术牌可用（且当前形态允许使用法术） -> 不能宣告无法行动
            if card.Type == model.CardTypeMagic && canUseMagic {
                return false, "有法术牌可用"
            }
        }
    }
    // 有可用的行动技能 -> 不能宣告无法行动
    if e.hasUsableActionSkillForExtraMagic(player) {
        return false, "有行动技能可用"
    }
    // 其他情况可以宣告无法行动
    return true, "没有可用的攻击牌、法术牌或行动技能"
}
```

### 3. 重构 handleActionSelectionCannotAct

修改 `internal/engine/action_submission_runtime.go`：

```go
func (e *GameEngine) handleActionSelectionCannotAct(player *model.Player) error {
    // 额外行动阶段的跳过逻辑（保持不变）
    if player.TurnState.CurrentExtraAction != "" {
        // ... existing extra action skip logic ...
    }

    // 主行动回合：使用统一的无法行动判断
    canCannotAct, reason := e.checkPlayerCannotAct(player)
    if !canCannotAct {
        return errors.New("你还有可用的攻击/法术牌或技能，无法宣告无法行动")
    }

    // 执行无法行动流程（展示、弃牌、重摸）
    e.executeCannotActFlow(player)
    return nil
}

// checkPlayerCannotAct 检查玩家是否可以宣告无法行动
func (e *GameEngine) checkPlayerCannotAct(player *model.Player) (bool, string) {
    roleID := ""
    if player.Character != nil {
        roleID = player.Character.ID
    }
    
    // 1. 先尝试角色自定义hook
    checker := roleRegistry.CannotActChecker(roleID)
    if checker != nil {
        canCannotAct, reason := checker(player)
        if canCannotAct {
            return true, "[角色规则] " + reason
        }
        // 角色hook返回false时，继续走默认判断
    }
    
    // 2. 无自定义hook或hook未拦截时，使用默认判断
    return DefaultCannotActChecker(e, player)
}
```

### 4. 魔剑士特殊规则示例

魔剑士暗影抗拒规则：手牌全是法术牌时也可以宣告无法行动。

```go
// MagicSwordsmanCannotActChecker 魔剑士无法行动判断
// 暗影抗拒：手牌全是法术牌时，即使有法术牌也不能使用，可以宣告无法行动
func MagicSwordsmanCannotActChecker(player *model.Player) (bool, string) {
    // 检查是否在暗影形态
    if !playerpkg.HasForm(player, model.FormMagicSwordsmanShadow) {
        return false, "" // 不在暗影形态，不拦截，走默认判断
    }
    
    // 检查手牌是否全是法术牌
    if len(player.Hand) == 0 {
        return false, "" // 无手牌，走默认判断
    }
    
    allMagic := true
    for _, c := range player.Hand {
        if c.Type != model.CardTypeMagic {
            allMagic = false
            break
        }
    }
    
    if allMagic {
        return true, "暗影抗拒：手牌全是法术牌，无法使用"
    }
    
    return false, "" // 有非法术牌，走默认判断
}

// 在 magic_swordsman module.go 中注册
func RoleEntry() player.RoleEntry {
    return player.RoleEntry{
        ID:               "magic_swordsman",
        CannotActChecker: MagicSwordsmanCannotActChecker,
        // ... other fields ...
    }
}
```

### 5. 在 RoleRegistry 中添加 CannotActChecker 查询方法

修改 `internal/engine/player/registry.go`：

```go
func mergeRoleEntry(base RoleEntry, overlay RoleEntry) RoleEntry {
    // ... existing merge logic ...
    if overlay.CannotActChecker != nil {
        merged.CannotActChecker = overlay.CannotActChecker
    }
    return merged
}

// CannotActChecker 返回指定角色的无法行动判断函数
func (r *RoleRegistry) CannotActChecker(roleID string) CannotActChecker {
    if r == nil || roleID == "" {
        return nil
    }
    entry := r.Entry(roleID)
    return entry.CannotActChecker
}
```

### 6. 额外行动无法行动的处理

额外行动无法行动时，当前逻辑是"跳过并进入TurnEnd"。这应该保持不变，因为：
- 额外行动是技能赋予的追加机会
- 无法执行时直接跳过，不涉及手牌展示/重摸

### 7. 关键修改文件

1. `internal/engine/player/role_entry.go` - 添加 `CannotActChecker` 类型定义和 RoleEntry 字段
2. `internal/engine/player/registry.go` - 添加 `CannotActChecker()` 查询方法和 merge 逻辑
3. `internal/engine/cannot_act_checker.go` - 新文件，通用无法行动判断逻辑
4. `internal/engine/action_submission_runtime.go` - 重构 `handleActionSelectionCannotAct`
5. `internal/engine/player/magic_swordsman/module.go` - 注册魔剑士 CannotActChecker

---

## Verification

1. 运行现有测试确保重构不破坏现有行为：
   ```bash
   go test -v ./internal/engine/action_selection_prompt_options_test.go
   go test -v -run "CannotAct" ./internal/engine/...
   ```

2. 验证魔剑士特殊规则仍然生效：
   ```bash
   go test -v -run "MagicSwordsman.*CannotAct" ./internal/engine/...
   ```

---

## 函数改造清单

### 文件 1: `internal/engine/player/role_entry.go`

#### 改造点 1: 新增 `CannotActChecker` 类型定义

**位置**: 在 `SkillUsabilityChecker` 定义之后

**新增内容**:
```go
// CannotActChecker 无法行动判断函数
// 返回 (canCannotAct, reason)：
//   - canCannotAct=true: 该角色认为当前可以宣告无法行动
//   - canCannotAct=false: 该角色认为当前不能宣告无法行动（仍有可执行动作）
// player 参数已包含手牌、形态等信息，大多数角色只需要检查 player 即可
type CannotActChecker func(player *model.Player) (bool, string)
```

#### 改造点 2: 在 `RoleEntry` 结构体中添加字段

**位置**: 在 `AttackCardElementTransform` 字段之后

**新增内容**:
```go
CannotActChecker CannotActChecker // 角色自定义无法行动判断hook（可选）
```

---

### 文件 2: `internal/engine/player/registry.go`

#### 改造点 1: 在 `mergeRoleEntry` 函数中添加 merge 逻辑

**位置**: 在现有 merge 逻辑末尾（第115行附近）

**新增内容**:
```go
if overlay.CannotActChecker != nil {
    merged.CannotActChecker = overlay.CannotActChecker
}
```

#### 改造点 2: 新增 `CannotActChecker()` 查询方法

**位置**: 在 `AttackCardElementTransform()` 方法之后

**新增内容**:
```go
// CannotActChecker 返回指定角色的无法行动判断函数
func (r *RoleRegistry) CannotActChecker(roleID string) CannotActChecker {
    if r == nil || roleID == "" {
        return nil
    }
    entry := r.Entry(roleID)
    return entry.CannotActChecker
}
```

---

### 文件 3: `internal/engine/cannot_act_checker.go` (新文件)

#### 新增函数 1: `DefaultCannotActChecker`

```go
// DefaultCannotActChecker 默认无法行动判断
// 检查玩家是否有可用的攻击牌或法术牌，或是否有可用的行动技能
func DefaultCannotActChecker(e *GameEngine, player *model.Player) (bool, string) {
    // 1. 检查手牌是否有可执行动作
    if len(player.Hand) > 0 {
        canUseMagic := e.canCastMagicInAction(player)
        total := playableCardCount(player)
        for idx := 0; idx < total; idx++ {
            card, _, _, ok := getPlayableCardByIndex(player, idx)
            if !ok {
                continue
            }
            // 有攻击牌可用 -> 不能宣告无法行动
            if card.Type == model.CardTypeAttack {
                return false, "有攻击牌可用"
            }
            // 有法术牌可用（且当前形态允许使用法术） -> 不能宣告无法行动
            if card.Type == model.CardTypeMagic && canUseMagic {
                return false, "有法术牌可用"
            }
        }
    }
    
    // 2. 检查是否有可用的行动技能
    if e.hasUsableActionSkillForExtraMagic(player) {
        return false, "有行动技能可用"
    }
    
    // 3. 其他情况：可以宣告无法行动
    return true, "没有可用的攻击牌、法术牌或行动技能"
}
```

#### 新增函数 2: `checkPlayerCannotAct`

```go
// checkPlayerCannotAct 检查玩家是否可以宣告无法行动
// 先调用角色自定义hook，若无hook则使用默认判断
func (e *GameEngine) checkPlayerCannotAct(player *model.Player) (bool, string) {
    roleID := ""
    if player.Character != nil {
        roleID = player.Character.ID
    }
    
    // 1. 先尝试角色自定义hook
    checker := roleRegistry.CannotActChecker(roleID)
    if checker != nil {
        canCannotAct, reason := checker(player)
        if canCannotAct {
            return true, "[角色规则] " + reason
        }
        // 角色hook返回false时，继续走默认判断
    }
    
    // 2. 无自定义hook或hook未拦截时，使用默认判断
    return DefaultCannotActChecker(e, player)
}
```

---

### 文件 4: `internal/engine/action_submission_runtime.go`

#### 改造点 1: 重构 `handleActionSelectionCannotAct` 函数

**当前状态**: 第236-311行，约75行代码

**改造目标**: 
1. 额外行动跳过逻辑保持不变（第237-248行）
2. 主行动回合无法行动判断改为调用 `checkPlayerCannotAct`
3. 魔剑士特殊重摸逻辑移到角色 hook 或单独函数

**改造后结构**:
```go
func (e *GameEngine) handleActionSelectionCannotAct(player *model.Player) error {
    // === 额外行动阶段：直接跳过 ===
    if player.TurnState.CurrentExtraAction != "" {
        if e.checkExtraActionCards(player, player.TurnState.CurrentExtraAction, player.TurnState.CurrentExtraElement) {
            return errors.New("当前额外行动仍有可执行动作，不能跳过")
        }
        e.skipExtraAction(player)
        return nil
    }

    // === 主行动回合：无法行动判断 ===
    canCannotAct, reason := e.checkPlayerCannotAct(player)
    if !canCannotAct {
        return errors.New("你还有可用的攻击/法术牌或技能，无法宣告无法行动")
    }
    
    // === 执行无法行动流程 ===
    e.executeCannotActFlow(player)
    return nil
}
```

#### 改造点 2: 新增 `skipExtraAction` 函数

**新增内容**:
```go
// skipExtraAction 跳过额外行动
func (e *GameEngine) skipExtraAction(player *model.Player) {
    constraintInfo := e.buildConstraintInfo(player.TurnState.CurrentExtraAction, player.TurnState.CurrentExtraElement)
    e.beginActionSummary("cannot_act", player.ID, "跳过额外行动", nil)
    e.Log(fmt.Sprintf("[Turn] %s 宣告【无法行动】，跳过本次额外行动%s", player.Name, constraintInfo))
    player.TurnState.CurrentExtraAction = ""
    player.TurnState.CurrentExtraElement = nil
    e.enterTurnEndStage()
}
```

#### 改造点 3: 新增 `executeCannotActFlow` 函数

**新增内容**:
```go
// executeCannotActFlow 执行无法行动流程（展示、弃牌、重摸）
func (e *GameEngine) executeCannotActFlow(player *model.Player) {
    e.beginActionSummary("cannot_act", player.ID, "无法行动", nil)
    handCount := len(player.Hand)
    
    if handCount == 0 {
        e.Log(fmt.Sprintf("[Action] %s 宣告【无法行动】（无手牌），结束本回合行动阶段", player.Name))
        player.TurnState.LockSpecialActionsForRemainderOfTurn()
        e.enterTurnEndStage()
        return
    }
    
    // 展示并弃掉全部手牌
    e.Log(fmt.Sprintf("[Action] %s 宣告【无法行动】，展示并弃掉全部手牌(%d张)", player.Name, handCount))
    e.NotifyCardRevealed(player.ID, append([]model.Card{}, player.Hand...), "discard")
    for _, c := range player.Hand {
        e.State.DiscardPile = append(e.State.DiscardPile, c)
    }
    player.Hand = player.Hand[:0]
    
    // 重摸相同数量的牌
    cards, newDeck, newDiscard := rules.DrawCards(e.State.Deck, e.State.DiscardPile, handCount)
    e.State.Deck = newDeck
    e.State.DiscardPile = newDiscard
    player.Hand = append(player.Hand, cards...)
    e.NotifyDrawCards(player.ID, handCount, "cannot_act_redraw")
    
    // 角色特定后续处理（如魔剑士全法术重摸）
    e.runCannotActFollowup(player)
    
    e.Log(fmt.Sprintf("[Action] %s 重新摸了%d张牌，且本回合不可执行特殊行动", player.Name, handCount))
    player.TurnState.LockSpecialActionsForRemainderOfTurn()
    e.enterActionExecutionStage()
}
```

#### 改造点 4: 新增 `runCannotActFollowup` 函数

**新增内容**:
```go
// runCannotActFollowup 无法行动后的角色特定后续处理
func (e *GameEngine) runCannotActFollowup(player *model.Player) {
    // 魔剑士特殊规则：全法术牌时继续重摸
    if playerpkg.IsCharacter(player, "magic_swordsman") {
        e.runMagicSwordsmanCannotActFollowup(player)
    }
    // 其他角色如有特殊规则，可在此扩展
}
```

#### 改造点 5: 新增 `runMagicSwordsmanCannotActFollowup` 函数

**将原有魔剑士重摸逻辑提取为独立函数**:
```go
// runMagicSwordsmanCannotActFollowup 魔剑士无法行动后续处理
// 如果新摸的手牌全是法术牌，则继续弃掉并重摸，直到有攻击牌或非纯法术牌
func (e *GameEngine) runMagicSwordsmanCannotActFollowup(player *model.Player) {
    for len(player.Hand) > 0 {
        hasAttack := false
        allMagic := true
        for _, c := range player.Hand {
            if c.Type == model.CardTypeAttack {
                hasAttack = true
                break
            }
            if c.Type != model.CardTypeMagic {
                allMagic = false
            }
        }
        if hasAttack || !allMagic {
            break
        }
        
        redrawCount := len(player.Hand)
        e.NotifyCardRevealed(player.ID, append([]model.Card{}, player.Hand...), "discard")
        e.State.DiscardPile = append(e.State.DiscardPile, player.Hand...)
        player.Hand = player.Hand[:0]
        nextCards, deck2, discard2 := rules.DrawCards(e.State.Deck, e.State.DiscardPile, redrawCount)
        e.State.Deck = deck2
        e.State.DiscardPile = discard2
        player.Hand = append(player.Hand, nextCards...)
        e.NotifyDrawCards(player.ID, redrawCount, "magic_swordsman_redraw")
        e.Log(fmt.Sprintf("[Action] %s 触发魔剑士重摸：全法术手牌已弃置并重摸%d张", player.Name, redrawCount))
    }
}
```

---

### 文件 5: `internal/engine/action_selection_policies.go`

#### 改造点: 在 `appendBaseActionSelectionOptions` 中使用统一判断

**位置**: 第194-224行

**当前逻辑**: 手动检查是否有攻击牌/法术牌

**改造目标**: 改为调用 `checkPlayerCannotAct` 判断是否应该显示"无法行动"选项

**改造后代码**:
```go
if !state.isRestrictedExtraAction {
    // 使用统一的无法行动判断
    canCannotAct, _ := e.checkPlayerCannotAct(player)
    
    if state.actionRuleMode == actionSelectionRuleForceAttackOrSkip {
        state.validOptions = append(state.validOptions, model.PromptOption{ID: "cannot_act", Label: "跳过行动（移除挑衅）"})
        return
    }
    if canCannotAct {
        state.validOptions = append(state.validOptions, model.PromptOption{ID: "cannot_act", Label: "无法行动（展示手牌）"})
    }
} else if !state.hasRestrictedExtraAction {
    state.validOptions = append(state.validOptions, model.PromptOption{ID: "cannot_act", Label: "跳过额外行动"})
}
```

---

## 改造执行顺序

### Phase 1: 基础设施
1. 在 `role_entry.go` 中新增 `CannotActChecker` 类型定义
2. 在 `RoleEntry` 结构体中添加 `CannotActChecker` 字段
3. 在 `registry.go` 中添加 merge 逻辑和查询方法

### Phase 2: 核心判断逻辑
4. 创建新文件 `cannot_act_checker.go`
5. 实现 `DefaultCannotActChecker`
6. 实现 `checkPlayerCannotAct`

### Phase 3: 流程重构
7. 新增 `skipExtraAction` 函数
8. 新增 `executeCannotActFlow` 函数
9. 新增 `runCannotActFollowup` 函数
10. 新增 `runMagicSwordsmanCannotActFollowup` 函数
11. 重构 `handleActionSelectionCannotAct`

### Phase 4: UI 层适配
12. 重构 `action_selection_policies.go` 中的判断逻辑

### Phase 5: 测试验证
14. 运行现有测试确保不破坏行为
15. 补充新测试覆盖重构后的逻辑