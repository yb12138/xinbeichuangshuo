# Engine 角色代码迁移到 player 子目录方案

## Context

当前 `internal/engine/` 目录下有约 180+ 个 .go 文件，其中大量文件是按角色组织的代码（技能流、配置测试、专属钩子等），混杂在引擎核心文件中。目标是将这些角色相关代码迁移到 `internal/engine/player/<角色名>/` 下对应目录，实现：

- 角色代码自治：每个角色的技能流、choices、defaults、helpers、专属测试 都在自己的目录下
- 引擎核心瘦身：`engine/` 只保留调度、注册、通用运行时等基础架构
- 与已有的 `player/<role>/module.go` + `choices.go` 模式对齐

## 迁移范围总览

### 当前 player 目录已有结构（保留）

每个 `player/<role>/` 已包含：
- `module.go` — 技能注册入口 (`SkillEntries()`, `ChoiceRouteSpecs()`)
- `choices.go` — 角色选择流 (`ChoiceHandler` 实现)

通用基础设施：
- `player/registry.go` — RoleRegistry
- `player/role_entry.go` — RoleEntry, ChoiceRuntime 接口
- `player/hand_limit_rule.go` — 手牌上限规则

### 需要迁移的文件分类

#### A. 技能流文件 → `player/<role>/skill_flow.go`

共 32 个 `skill_flow_*.go` 文件（不含 `skill_flow_system_choices.go`）：

| 角色 | 迁移文件 | 目标路径 |
|------|----------|----------|
| adventurer | `skill_flow_adventurer.go` | `player/adventurer/skill_flow.go` |
| angel | `skill_flow_angel.go` | `player/angel/skill_flow.go` |
| assassin | `skill_flow_assassin.go` | `player/assassin/skill_flow.go` |
| bard | `skill_flow_bard.go` + `skill_flow_bard_turn_hooks.go` | `player/bard/skill_flow.go` + `player/bard/turn_hooks.go` |
| beast_samurai | `skill_flow_beast_samurai.go` | `player/beast_samurai/skill_flow.go` |
| blaze_witch | `skill_flow_blaze_witch.go` | `player/blaze_witch/skill_flow.go` |
| blood_priestess | `skill_flow_blood_priestess.go` | `player/blood_priestess/skill_flow.go` |
| butterfly_dancer | `skill_flow_butterfly_dancer.go` | `player/butterfly_dancer/skill_flow.go` |
| crimson_knight | `skill_flow_crimson_knight.go` | `player/crimson_knight/skill_flow.go` |
| crimson_sword_spirit | `skill_flow_crimson_sword_spirit.go` | `player/crimson_sword_spirit/skill_flow.go` |
| elementalist | `skill_flow_elementalist.go` | `player/elementalist/skill_flow.go` |
| elf_archer | `skill_flow_elf_archer.go` | `player/elf_archer/skill_flow.go` |
| fighter | `skill_flow_fighter.go` | `player/fighter/skill_flow.go` |
| hero | `skill_flow_hero.go` | `player/hero/skill_flow.go` |
| holy_bow | `skill_flow_holy_bow.go` | `player/holy_bow/skill_flow.go` |
| holy_lancer | `skill_flow_holy_lancer.go` | `player/holy_lancer/skill_flow.go` |
| magic_bow | `skill_flow_magic_bow.go` | `player/magic_bow/skill_flow.go` |
| magic_lancer | `skill_flow_magic_lancer.go` | `player/magic_lancer/skill_flow.go` |
| magic_swordsman | `skill_flow_magic_swordsman.go` | `player/magic_swordsman/skill_flow.go` |
| moon_goddess | `skill_flow_moon_goddess.go` | `player/moon/skill_flow.go` |
| onmyoji | `skill_flow_onmyoji.go` | `player/onmyoji/skill_flow.go` |
| plague_mage | `skill_flow_plague_mage.go` | `player/plague_mage/skill_flow.go` |
| prayer_master | `skill_flow_prayer_master.go` | `player/prayer_master/skill_flow.go` |
| priest | `skill_flow_priest.go` | `player/priest/skill_flow.go` |
| sage | `skill_flow_sage.go` | `player/sage/skill_flow.go` |
| saintess | `skill_flow_saintess.go` | `player/saintess/skill_flow.go` |
| sealer | `skill_flow_sealer.go` | `player/sealer/skill_flow.go` |
| soul_sorcerer | `skill_flow_soul_sorcerer.go` | `player/soul_sorcerer/skill_flow.go` |
| spirit_caster | `skill_flow_spirit_caster.go` | `player/spirit_caster/skill_flow.go` |
| sword_emperor | `skill_flow_sword_emperor.go` | `player/sword_emperor/skill_flow.go` |
| valkyrie | `skill_flow_valkyrie.go` | `player/valkyrie/skill_flow.go` |
| war_homunculus | `skill_flow_war_homunculus.go` | `player/war_homunculus/skill_flow.go` |

#### B. 角色默认配置 → `player/<role>/defaults.go`

当前在 `role_defaults.go` 中按角色 ID 映射。需拆分为：
- 每个角色目录下新建 `defaults.go`，导出 `Defaults()` 函数
- 通用基础设施保留在 `player/` 下

涉及角色：plague_mage, crimson_sword_spirit, prayer_master, crimson_knight, war_homunculus, priest, onmyoji, blaze_witch, magic_lancer, spirit_caster, bard, hero, fighter, holy_bow, sword_emperor, beast_samurai, holy_lancer, arbiter, soul_sorcerer, moon_goddess, butterfly_dancer, blood_priestess

#### C. 开局专属牌 → `player/<role>/starter_cards.go`

`starter_role_cards.go` 按角色拆分：
- sealer → `player/sealer/starter_cards.go`
- crimson_sword_spirit → `player/crimson_sword_spirit/starter_cards.go`
- hero → `player/hero/starter_cards.go`
- soul_sorcerer → `player/soul_sorcerer/starter_cards.go`
- blood_priestess → `player/blood_priestess/starter_cards.go`
- bard → `player/bard/starter_cards.go`
- 通用辅助 `ensureExclusiveStarterCard()` 保留在 `player/` 下

#### D. 角色专属 helpers → `player/<role>/helpers.go`

| 角色 | 来源文件 | 目标 |
|------|----------|------|
| butterfly_dancer | `new_roles_butterfly_damage_helpers.go` | `player/butterfly_dancer/helpers.go` |
| bard + hero 等 | `new_roles_card_counter_helpers.go` | 拆分到各角色目录 |
| bard + holy_bow 等 | `new_roles_combat_resource_helpers.go` | 拆分到各角色目录 |
| magic_bow + holy_lancer 等 | `new_roles_field_resource_morale_helpers.go` | 拆分到各角色目录 |
| elf_archer + magic_swordsman 等 | `new_roles_identity_forms_helpers.go` | 拆分到各角色目录 |
| soul_sorcerer + blood_priestess | `new_roles_link_status_helpers.go` | 拆分到各角色目录 |
| moon_goddess | `new_roles_moon_followup_helpers.go` | `player/moon/helpers.go` |
| magic_lancer + prayer_master 等 | `new_roles_post_resolution_helpers.go` | 拆分到各角色目录 |
| 通用 | `new_roles_shared_runtime_helpers.go` | 保留在 `player/shared_helpers.go` |

#### E. 中断 prompt/response 文件 → `player/<role>/interrupt.go`

| 角色 | 来源文件 | 目标 |
|------|----------|------|
| blademaster | `interrupt_prompt_blademaster.go` | `player/blade_master/interrupt.go` |
| saintess | `interrupt_prompt_saintess.go` | `player/saintess/interrupt.go` |
| magic_bullet | `interrupt_prompt_magic_bullet.go` | `player/magical_girl/interrupt.go` |
| holy/saint | `interrupt_response_holy_saint.go` | `player/saintess/interrupt_response.go` |
| magic_blast | `interrupt_response_magic_blast.go` | `player/magical_girl/interrupt_blast.go` |
| magic_missile | `interrupt_response_magic_missile.go` | `player/magical_girl/interrupt_missile.go` |

#### F. 角色专属回归测试 → `player/<role>/<test>.go`

明确属于单角色的回归测试：

| 角色 | 迁移文件 |
|------|----------|
| adventurer | `adventurer_fraud_regression_test.go`, `adventurer_priest_rules_regression_test.go` |
| angel | `angel_config_regression_test.go` |
| arbiter | `arbiter_law_regression_test.go` |
| archer | `archer_config_regression_test.go` |
| assassin | `assassin_backlash_regression_test.go`, `assassin_config_regression_test.go`, `assassin_water_shadow_skip_regression_test.go` |
| bard | `bard_regression_test.go` |
| beast_samurai | `beast_samurai_regression_test.go` |
| berserker | `berserker_config_regression_test.go` |
| blaze_witch | `blaze_witch_skill_regression_test.go` |
| blood_priestess | `blood_priestess_regression_test.go` |
| butterfly_dancer | `butterfly_dancer_regression_test.go` |
| crimson_knight | `crimson_knight_bloody_prayer_regression_test.go`, `crimson_knight_killing_feast_regression_test.go`, `crk_hom_skill_regression_test.go` |
| crimson_sword_spirit | `crimson_sword_spirit_config_regression_test.go`, `crimson_sword_spirit_regression_test.go` |
| elementalist | `elementalist_regression_test.go` |
| elf_archer | `elf_archer_skill_regression_test.go` |
| fighter | `fighter_regression_test.go` |
| hero | `hero_regression_test.go` |
| holy_bow | `holy_bow_regression_test.go` |
| holy_lancer | `holy_lancer_earth_spear_regression_test.go` |
| magic_bow | `magic_bow_regression_test.go` |
| magic_lancer | `magic_lancer_regression_test.go` |
| magic_swordsman | `magic_swordsman_config_regression_test.go`, `magic_swordsman_prayer_css_bugfix_regression_test.go`, `magic_swordsman_shadow_reject_response_regression_test.go` |
| magical_girl | `magical_girl_config_regression_test.go` |
| moon_goddess | `moon_goddess_regression_test.go` |
| onmyoji | `onmyoji_rules_regression_test.go`, `onmyoji_skill_flow_regression_test.go` |
| plague_mage | `plague_mage_skill_regression_test.go` |
| prayer_master | `prayer_master_config_regression_test.go` |
| priest | — (无专属测试) |
| sage | `sage_skill_regression_test.go` |
| saintess | `saintess_config_regression_test.go`, `saintess_frost_prayer_resume_regression_test.go` |
| sealer | `sealer_status_resolver_regression_test.go` |
| soul_sorcerer | `soul_sorcerer_regression_test.go` |
| spirit_caster | `spirit_caster_regression_test.go` |
| sword_emperor | `sword_emperor_regression_test.go` |
| valkyrie | `valkyrie_combo_regression_test.go`, `valkyrie_config_regression_test.go` |

**不迁移的通用测试**（留在 engine/）：
- `basic_effect_*`, `config_alignment_*`, `counter_attack_*`, `combat_magic_*`, `crystal_substitute_*`, `dark_counter_*`, `exclusive_skill_card_*`, `extract_cancel_*`, `field_card_flow_*`, `followup_mount_*`, `prayer_form_persist_*`, `role_skill_bugfix_*`, `shield_defend_*`, `skill_dispatcher_priority_*`, `skill_policy_mount_*`, `stage_timing_*`, `startup_skip_*`, `new_roles_regression_test.go`, `elf_holy_lancer_bugfix_regression_test.go`, `berserker_sealer_damage_regression_test.go`, `choice_*_test.go`, `discard_*_test.go`, `holy_sword_regression_test.go`

### 不迁移的文件（保留在 engine/）

引擎核心调度与基础设施：
- `game.go`, `game_drive.go`, `game_action_router.go`, `game_context_builder.go`, `game_damage_pipeline.go`, `game_player_lifecycle.go`
- `combat.go`, `magic.go`
- `skill_dispatcher.go`, `skill_registry.go`, `skill_use.go`, `skill_use_execution.go`, `skill_use_policy.go`, `skill_use_validation.go`, `skill_runtime_host.go`, `skill_post_runtime.go`, `skill_target_rules.go`
- `choice_*.go`（路由/引导核心，非角色专属）
- `interrupt_*.go`（框架/运行时核心，非角色专属部分）
- `timing_*` 通用时序注册
- `*_runtime.go` 通用运行时
- `player_role_mounts.go`, `role_choice_runtime.go`（桥接层）
- 所有 `*_policy_*.go`, `*_hooks.go`（通用钩子）
- `skill_flow_system_choices.go`（系统级选择）

## 迁移步骤

> 各 Phase 详细步骤见独立文档：
> - [Phase 1 详细步骤](migration_phase1_skill_flows.md) — 迁移技能流文件
> - [Phase 2 详细步骤](migration_phase2_defaults.md) — 拆分角色默认配置
> - [Phase 3 详细步骤](migration_phase3_starter_cards.md) — 拆分开局专属牌
> - [Phase 4 详细步骤](migration_phase4_helpers.md) — 拆分角色 helpers
> - [Phase 5 详细步骤](migration_phase5_interrupts.md) — 迁移中断 prompt/response
> - [Phase 6 详细步骤](migration_phase6_tests.md) — 迁移角色专属回归测试

## 验证方式

每个 Phase 完成后：
1. `go build ./...` — 编译通过
2. `go test ./internal/engine/...` — 全量引擎测试通过
3. `go test ./internal/engine/player/...` — player 目录测试通过
4. 抽样 CLI 对局验证关键角色技能流

## 风险与注意事项

1. **包循环引用**：角色子包不能 import engine 包，必须通过 `ChoiceRuntime` 接口解耦
2. **方法接收者转换**：`(e *GameEngine) method()` 需改为接受接口参数的函数
3. **共享 helper 归属**：`new_roles_*_helpers.go` 中混合多角色逻辑，需仔细拆分
4. **测试隔离**：迁移后的测试不能直接访问 GameEngine 内部状态，需通过接口
5. **渐进迁移**：建议按 Phase 逐步推进，每步保证编译和测试通过
