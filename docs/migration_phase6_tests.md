# Phase 6：迁移角色专属回归测试

## 目标

将 `engine/` 中明确归属单个角色的回归测试文件迁移到 `player/<role>/` 目录，通用/跨角色测试保留在 engine。

## 当前状况

`engine/` 下约有 70 个 `*_regression_test.go` 文件。其中约 50 个明确只测试单个角色，另外约 20 个是通用或跨角色测试。

### 分类原则

- **迁移**：文件名明确包含角色名（如 `angel_config_regression_test.go`）且测试代码只涉及该角色
- **不迁移**：通用测试（如 `basic_effect_*`）、跨角色测试（如 `berserker_sealer_damage_*`）、引擎机制测试（如 `choice_*_test.go`）

## 详细步骤

### Step 6.1：确认每个测试文件的角色归属

对每个待迁移的测试文件，验证其是否只测试单一角色：

```bash
# 以 angel 为例，检查测试中是否引用了其他角色
grep -n 'CharacterID\|"archer"\|"assassin"\|"berserker"' angel_config_regression_test.go
```

如果有交叉引用，需评估是否仍然迁移（可能需要保留在 engine 或拆分）。

### Step 6.2：创建目标目录和文件

```bash
# 以 angel 为例
cp internal/engine/angel_config_regression_test.go \
   internal/engine/player/angel/config_regression_test.go
```

### Step 6.3：修改包声明

```go
// 修改前
package engine

// 修改后
package angel
```

### Step 6.4：更新 import

迁移到角色包后，测试代码不能再直接创建 `GameEngine` 实例。需要根据情况调整：

**情况 A：测试通过 engine 公共 API 操作（推荐）**

```go
// 修改前 (package engine)
func TestAngelConfig(t *testing.T) {
    e := NewGameEngine()
    e.AddPlayer("p1", "Red", "angel")
    // ... 直接操作 e.State ...
}

// 修改后 (package angel) — 通过 engine 公共 API
import "starcup-engine/internal/engine"

func TestAngelConfig(t *testing.T) {
    e := engine.NewGameEngine()
    e.AddPlayer("p1", "Red", "angel")
    // ... 操作不变，只要通过公共 API ...
}
```

**情况 B：测试直接访问 engine 内部状态**

```go
// 修改前 (package engine)
func TestAngelConfig(t *testing.T) {
    e := NewGameEngine()
    e.AddPlayer("p1", "Red", "angel")
    player := e.State.Players["p1"]  // 直接访问内部状态
    assert.Equal(t, 6, player.MaxHand)
}

// 修改后 (package angel) — 使用白盒测试或导出 API
// 方案1：保持 engine 包的 whitebox 测试文件在 engine/ 下，只迁移黑盒测试
// 方案2：在 engine 中导出必要的查询函数
// 方案3：使用 engine_test 包（external test）
```

**推荐方案**：对于需要访问 `e.State` 内部状态的测试，使用 external test package：

```go
// player/angel/config_test.go — 使用 external test package
package angel_test

import (
    "testing"

    "starcup-engine/internal/engine"
)

func TestAngelConfig(t *testing.T) {
    e := engine.NewGameEngine()
    e.AddPlayer("p1", "Red", "angel")
    // 通过 engine 公共 API 验证
}
```

如果 engine 的公共 API 不够（无法访问 `e.State.Players`），则：
1. 在 engine 中导出必要的查询函数：`func (e *GameEngine) GetPlayer(id string) *model.Player`
2. 或者将该测试保留在 `engine/` 包内

### Step 6.5：处理跨文件依赖

部分测试文件中可能调用了 `engine` 包的 test helper 函数（如 `newTestEngine()`、`setupTestPlayers()` 等）：

```bash
# 查找测试 helper 函数
grep -rn 'func newTest\|func setupTest\|func makeTest' internal/engine/*_test.go | grep -v 'func Test'
```

这些 helper 函数的处理方式：
1. **通用 helper**：移到 `tests/` 或 `internal/testutils/` 包
2. **角色专属 helper**：随测试文件一起迁移

### Step 6.6：逐个迁移角色测试

**迁移清单**（按建议顺序）：

#### 批次 1（低复杂度，纯配置测试）

| 角色 | 文件 | 目标 | 预期难度 |
|------|------|------|---------|
| angel | `angel_config_regression_test.go` | `player/angel/config_test.go` | 低 |
| archer | `archer_config_regression_test.go` | `player/archer/config_test.go` | 低 |
| assassin | `assassin_config_regression_test.go` | `player/assassin/config_test.go` | 低 |
| berserker | `berserker_config_regression_test.go` | `player/berserker/config_test.go` | 低 |
| magical_girl | `magical_girl_config_regression_test.go` | `player/magical_girl/config_test.go` | 低 |
| valkyrie | `valkyrie_config_regression_test.go` | `player/valkyrie/config_test.go` | 低 |
| saintess | `saintess_config_regression_test.go` | `player/saintess/config_test.go` | 低 |
| magic_swordsman | `magic_swordsman_config_regression_test.go` | `player/magic_swordsman/config_test.go` | 低 |
| crimson_sword_spirit | `crimson_sword_spirit_config_regression_test.go` | `player/crimson_sword_spirit/config_test.go` | 低 |
| prayer_master | `prayer_master_config_regression_test.go` | `player/prayer_master/config_test.go` | 低 |

#### 批次 2（中等复杂度，技能流测试）

| 角色 | 文件 | 目标 |
|------|------|------|
| assassin | `assassin_backlash_regression_test.go` | `player/assassin/backlash_test.go` |
| assassin | `assassin_water_shadow_skip_regression_test.go` | `player/assassin/water_shadow_test.go` |
| bard | `bard_regression_test.go` | `player/bard/skill_test.go` |
| beast_samurai | `beast_samurai_regression_test.go` | `player/beast_samurai/skill_test.go` |
| blaze_witch | `blaze_witch_skill_regression_test.go` | `player/blaze_witch/skill_test.go` |
| blood_priestess | `blood_priestess_regression_test.go` | `player/blood_priestess/skill_test.go` |
| butterfly_dancer | `butterfly_dancer_regression_test.go` | `player/butterfly_dancer/skill_test.go` |
| elementalist | `elementalist_regression_test.go` | `player/elementalist/skill_test.go` |
| elf_archer | `elf_archer_skill_regression_test.go` | `player/elf_archer/skill_test.go` |
| fighter | `fighter_regression_test.go` | `player/fighter/skill_test.go` |
| hero | `hero_regression_test.go` | `player/hero/skill_test.go` |
| holy_bow | `holy_bow_regression_test.go` | `player/holy_bow/skill_test.go` |
| magic_bow | `magic_bow_regression_test.go` | `player/magic_bow/skill_test.go` |
| magic_lancer | `magic_lancer_regression_test.go` | `player/magic_lancer/skill_test.go` |
| sage | `sage_skill_regression_test.go` | `player/sage/skill_test.go` |
| soul_sorcerer | `soul_sorcerer_regression_test.go` | `player/soul_sorcerer/skill_test.go` |
| spirit_caster | `spirit_caster_regression_test.go` | `player/spirit_caster/skill_test.go` |
| sword_emperor | `sword_emperor_regression_test.go` | `player/sword_emperor/skill_test.go` |
| valkyrie | `valkyrie_combo_regression_test.go` | `player/valkyrie/combo_test.go` |

#### 批次 3（高复杂度，多文件/多角色关联）

| 角色 | 文件 | 目标 | 注意事项 |
|------|------|------|---------|
| adventurer | `adventurer_fraud_regression_test.go` | `player/adventurer/fraud_test.go` | 欺诈机制测试 |
| adventurer | `adventurer_priest_rules_regression_test.go` | `player/adventurer/priest_rules_test.go` | 涉及 priest 交互 |
| arbiter | `arbiter_law_regression_test.go` | `player/arbiter/law_test.go` | 法则系统测试 |
| crimson_knight | `crimson_knight_bloody_prayer_regression_test.go` | `player/crimson_knight/bloody_prayer_test.go` | |
| crimson_knight | `crimson_knight_killing_feast_regression_test.go` | `player/crimson_knight/killing_feast_test.go` | |
| crimson_knight | `crk_hom_skill_regression_test.go` | `player/crimson_knight/hom_skill_test.go` | 涉及 war_homunculus？ |
| crimson_sword_spirit | `crimson_sword_spirit_regression_test.go` | `player/crimson_sword_spirit/skill_test.go` | |
| holy_lancer | `holy_lancer_earth_spear_regression_test.go` | `player/holy_lancer/earth_spear_test.go` | |
| magic_swordsman | `magic_swordsman_prayer_css_bugfix_regression_test.go` | `player/magic_swordsman/bugfix_test.go` | |
| magic_swordsman | `magic_swordsman_shadow_reject_response_regression_test.go` | `player/magic_swordsman/shadow_test.go` | |
| moon_goddess | `moon_goddess_regression_test.go` | `player/moon/skill_test.go` | |
| onmyoji | `onmyoji_rules_regression_test.go` | `player/onmyoji/rules_test.go` | |
| onmyoji | `onmyoji_skill_flow_regression_test.go` | `player/onmyoji/skill_flow_test.go` | |
| plague_mage | `plague_mage_skill_regression_test.go` | `player/plague_mage/skill_test.go` | |
| saintess | `saintess_frost_prayer_resume_regression_test.go` | `player/saintess/frost_prayer_test.go` | |
| sealer | `sealer_status_resolver_regression_test.go` | `player/sealer/status_test.go` | |

### Step 6.7：不迁移的测试文件

以下测试留在 `engine/`（通用机制或跨角色）：

| 文件 | 原因 |
|------|------|
| `basic_effect_before_action_regression_test.go` | 通用效果机制 |
| `basic_effect_stack_regression_test.go` | 通用效果堆叠 |
| `berserker_sealer_damage_regression_test.go` | 跨角色交互 |
| `choice_binding_completeness_test.go` | 选择系统核心 |
| `choice_unknown_choice_type_test.go` | 选择系统核心 |
| `combat_magic_role_fix_regression_test.go` | 通用战斗修复 |
| `config_alignment_regression_test.go` | 全角色配置对齐 |
| `counter_attack_action_gating_regression_test.go` | 通用反击机制 |
| `crystal_substitute_regression_test.go` | 通用水晶替代 |
| `dark_counter_visibility_regression_test.go` | 通用暗灭规则 |
| `discard_choice_runtime_test.go` | 弃牌系统核心 |
| `discard_subflow_invariant_test.go` | 弃牌流程不变量 |
| `elf_holy_lancer_bugfix_regression_test.go` | 跨角色 |
| `exclusive_skill_card_regression_test.go` | 通用专属牌 |
| `extract_cancel_regression_test.go` | 提炼取消机制 |
| `field_card_flow_regression_test.go` | 场上牌通用流程 |
| `followup_mount_regression_test.go` | followup 挂载机制 |
| `holy_sword_regression_test.go` | 圣剑（非角色专属道具） |
| `new_roles_regression_test.go` | 全新角色批量测试 |
| `prayer_form_persist_regression_test.go` | 祈祷形态持久化（跨角色） |
| `resume_point_test.go` | resume point 机制 |
| `role_skill_bugfix_regression_test.go` | 跨角色 bug 修复 |
| `shield_defend_rule_test.go` | 盾防御通用规则 |
| `skill_dispatcher_priority_regression_test.go` | 调度器优先级 |
| `skill_policy_mount_regression_test.go` | 策略挂载机制 |
| `stage_timing_regression_test.go` | 时序阶段测试 |
| `startup_skip_regression_test.go` | 启动跳过机制 |

### Step 6.8：删除 engine 中的已迁移测试

```bash
# 每迁移一个角色后
rm internal/engine/angel_config_regression_test.go
```

### Step 6.9：验证

```bash
# 编译
go build ./...

# 全量测试
go test ./internal/engine/... -count=1

# 按角色测试
go test ./internal/engine/player/angel/... -count=1
go test ./internal/engine/player/bard/... -count=1
go test ./internal/engine/player/sword_emperor/... -count=1

# 确认测试数量无遗漏
go test ./internal/engine/... -count=1 -v 2>&1 | grep -c '--- PASS'
```

## 风险

1. **白盒测试**：大量测试当前以 `package engine` 编写（白盒），可访问 `e.State` 等内部状态。迁移到角色包后变为黑盒，可能需要导出额外的查询 API
2. **测试 helper 共享**：如果多个角色测试共用相同的 setup 代码，需提取到 `internal/testutils/` 包
3. **`crk_hom_skill_regression_test.go`**：文件名暗示可能同时涉及 crimson_knight 和 war_homunculus，需确认归属
4. **编译速度**：将测试移到子包后，`go test ./...` 的并行度可能提升，但也可能增加编译开销
5. **CI 集成**：如果项目有 CI 流程，需确保测试路径更新
