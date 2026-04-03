# 附录：`Player.Tokens` 代码 key 索引（自动生成）

本文件由 `scripts/gen_player_tokens_appendix.py` 生成，用于盘点 `map[string]int` 使用的字符串 key。 **动态 key**（如 `turnScopedResetKeys`）见「动态下标」一节。

**迁出 Tokens 的优化方案**（流程/回合/锁、UI 派生与可见性等）：见 [player_tokens_runtime_refactor_plan.md](player_tokens_runtime_refactor_plan.md)。

- **生成日期**：2026-04-01
- **扫描范围**：`internal/**/*.go`
- **匹配模式**：`Tokens["..."]`、`delete(...Tokens, "...")`、`getToken`/`setToken`/`addToken(..., "...")`
- **去重后条目数**：766 处引用；**不同 key 数**：113

**重新生成**：在仓库根目录执行 `python3 scripts/gen_player_tokens_appendix.py`

## 1. 按 key 字母序（含 `路径:行号`）

### `adventurer_extract_last_crystal`
- `internal/engine/skills/handlers_new_roles.go:693` — `transferCrystal := getToken(ctx.User, "adventurer_extract_last_crystal")`
- `internal/engine/special_action_runtime.go:285` — `p.Tokens["adventurer_extract_last_crystal"] = 0`
- `internal/engine/special_action_runtime.go:297` — `p.Tokens["adventurer_extract_last_crystal"] = crystal`
- `internal/server/state_view.go:91` — `delete(view.Tokens, "adventurer_extract_last_crystal")`

### `adventurer_extract_last_gem`
- `internal/engine/skills/handlers_new_roles.go:692` — `transferGem := getToken(ctx.User, "adventurer_extract_last_gem")`
- `internal/engine/special_action_runtime.go:284` — `p.Tokens["adventurer_extract_last_gem"] = 0`
- `internal/engine/special_action_runtime.go:296` — `p.Tokens["adventurer_extract_last_gem"] = gem`
- `internal/server/state_view.go:90` — `delete(view.Tokens, "adventurer_extract_last_gem")`

### `adventurer_extract_requires_paradise`
- `internal/engine/interrupt_runtime.go:27` — `if p := e.State.Players[act.PlayerID]; p != nil && p.Tokens != nil && p.Tokens["adventurer_extract_requires_paradise"] > 0 {`
- `internal/engine/skill_flow_adventurer.go:19` — `if player == nil || player.Tokens == nil || player.Tokens["adventurer_extract_requires_paradise"] <= 0 {`
- `internal/engine/skills/handlers_new_roles.go:696` — `setToken(ctx.User, "adventurer_extract_requires_paradise", 0)`
- `internal/engine/skills/handlers_new_roles.go:716` — `setToken(ctx.User, "adventurer_extract_requires_paradise", 0)`
- `internal/engine/skills/handlers_new_roles.go:720` — `forceTransfer := getToken(ctx.User, "adventurer_extract_requires_paradise") > 0`
- `internal/engine/special_action_runtime.go:286` — `p.Tokens["adventurer_extract_requires_paradise"] = 0`
- `internal/engine/special_action_runtime.go:299` — `p.Tokens["adventurer_extract_requires_paradise"] = 1`
- `internal/engine/special_action_runtime.go:301` — `p.Tokens["adventurer_extract_requires_paradise"] = 0`

### `arbiter_forced_doomsday_done_turn`
- `internal/engine/action_submission_runtime.go:62` — `player.Tokens["arbiter_forced_doomsday_done_turn"] = 1`
- `internal/engine/phase_hooks.go:117` — `player.Tokens["arbiter_forced_doomsday_done_turn"] = 0`

### `arbiter_forced_doomsday_pending`
- `internal/engine/action_selection_prompt_options_test.go:304` — `p1.Tokens["arbiter_forced_doomsday_pending"] = 1`
- `internal/engine/action_submission_runtime.go:60` — `if player.Tokens["arbiter_forced_doomsday_pending"] > 0 && act.SkillID == "arbiter_doomsday" {`
- `internal/engine/action_submission_runtime.go:61` — `player.Tokens["arbiter_forced_doomsday_pending"] = 0`
- `internal/engine/arbiter_law_regression_test.go:252` — `if got := p1.Tokens["arbiter_forced_doomsday_pending"]; got != 1 {`
- `internal/engine/arbiter_law_regression_test.go:279` — `if got := p1.Tokens["arbiter_forced_doomsday_pending"]; got != 0 {`
- `internal/engine/arbiter_law_regression_test.go:407` — `if got := p1.Tokens["arbiter_forced_doomsday_pending"]; got != 1 {`
- `internal/engine/runtime_policy_hooks.go:218` — `player.Tokens["arbiter_forced_doomsday_pending"] = 0`
- `internal/engine/runtime_policy_hooks.go:222` — `player.Tokens["arbiter_forced_doomsday_pending"] = 0`
- `internal/engine/runtime_policy_hooks.go:225` — `if player.Tokens["arbiter_forced_doomsday_pending"] == 0 {`
- `internal/engine/runtime_policy_hooks.go:228` — `player.Tokens["arbiter_forced_doomsday_pending"] = 1`
- `internal/engine/runtime_policy_hooks.go:238` — `if player.Tokens["arbiter_forced_doomsday_pending"] > 0 {`
- `internal/engine/runtime_policy_hooks.go:298` — `if player.Tokens["arbiter_forced_doomsday_pending"] <= 0 {`
- `internal/engine/runtime_policy_hooks.go:338` — `if e == nil || player == nil || player.Tokens["arbiter_forced_doomsday_pending"] <= 0 {`
- `internal/server/available_skill_adapter.go:16` — `forcedDoomsdayOnly := p.Tokens != nil && p.Tokens["arbiter_forced_doomsday_pending"] > 0`
- `internal/server/elementalist_available_skills_test.go:75` — `p1.Tokens["arbiter_forced_doomsday_pending"] = 1`

### `arbiter_form`
- `internal/server/state_view.go:104` — `delete(view.Tokens, "arbiter_form")`

### `arbiter_law_inited`
- `internal/engine/arbiter_law_regression_test.go:52` — `if got := p1.Tokens["arbiter_law_inited"]; got != 1 {`
- `internal/engine/skills/handlers_new_roles.go:482` — `return ctx != nil && ctx.User != nil && getToken(ctx.User, "arbiter_law_inited") == 0`
- `internal/engine/skills/handlers_new_roles.go:487` — `setToken(ctx.User, "arbiter_law_inited", 1)`

### `arbiter_skip_forced_doomsday`
- `internal/engine/phase_hooks.go:116` — `player.Tokens["arbiter_skip_forced_doomsday"] = 0`
- `internal/engine/skill_flow_sealer.go:71` — `player.Tokens["arbiter_skip_forced_doomsday"] = 1`
- `internal/engine/skill_flow_system_choices.go:119` — `player.Tokens["arbiter_skip_forced_doomsday"] = 1`

### `bd_descent_used_turn`
- `internal/engine/bard_regression_test.go:106` — `if got := bard.Tokens["bd_descent_used_turn"]; got != 1 {`
- `internal/engine/skill_flow_bard.go:205` — `user.Tokens["bd_descent_used_turn"] = 1`
- `internal/engine/skill_flow_bard_turn_hooks.go:99` — `if hasBardEternalPrisonerForm(source) || source.Tokens["bd_descent_used_turn"] > 0 {`

### `bd_inspiration`
- `internal/engine/action_skill_runtime_rules.go:104` — `inspiration = player.Tokens["bd_inspiration"]`
- `internal/engine/bard_regression_test.go:103` — `if got := bard.Tokens["bd_inspiration"]; got != 1 {`
- `internal/engine/bard_regression_test.go:168` — `bard.Tokens["bd_inspiration"] = 3`
- `internal/engine/bard_regression_test.go:190` — `if got := bard.Tokens["bd_inspiration"]; got != 1 {`
- `internal/engine/bard_regression_test.go:323` — `if got := bard.Tokens["bd_inspiration"]; got != 1 {`
- `internal/engine/bard_regression_test.go:376` — `if got := bard.Tokens["bd_inspiration"]; got != 1 {`
- `internal/engine/bard_regression_test.go:405` — `bard.Tokens["bd_inspiration"] = 3`
- `internal/engine/new_roles_helpers.go:1660` — `v := player.Tokens["bd_inspiration"]`
- `internal/engine/new_roles_helpers.go:1667` — `player.Tokens["bd_inspiration"] = v`
- `internal/engine/new_roles_helpers.go:1685` — `player.Tokens["bd_inspiration"] = v`
- `internal/engine/skills/handlers_bard.go:15` — `v := user.Tokens["bd_inspiration"]`
- `internal/engine/skills/handlers_bard.go:22` — `user.Tokens["bd_inspiration"] = v`
- `internal/server/available_skill_adapter.go:76` — `inspiration = p.Tokens["bd_inspiration"]`

### `bd_prisoner_form`
- `internal/server/state_view.go:99` — `delete(view.Tokens, "bd_prisoner_form")`

### `bd_rousing_prompted_turn`
- `internal/engine/skill_flow_bard_turn_hooks.go:15` — `if current.Tokens["bd_rousing_prompted_turn"] > 0 {`
- `internal/engine/skill_flow_bard_turn_hooks.go:18` — `current.Tokens["bd_rousing_prompted_turn"] = 1`

### `bd_victory_prompted_turn`
- `internal/engine/skill_flow_bard_turn_hooks.go:44` — `if current.Tokens["bd_victory_prompted_turn"] > 0 {`
- `internal/engine/skill_flow_bard_turn_hooks.go:47` — `current.Tokens["bd_victory_prompted_turn"] = 1`

### `bp_bleed_form`
- `internal/server/state_view.go:103` — `delete(view.Tokens, "bp_bleed_form")`

### `bp_bleed_tick_done_turn`
- `internal/engine/runtime_policy_hooks.go:68` — `if !hasBloodPriestessBleedingForm(player) || player.Tokens["bp_bleed_tick_done_turn"] > 0 {`
- `internal/engine/runtime_policy_hooks.go:71` — `player.Tokens["bp_bleed_tick_done_turn"] = 1`

### `bp_shared_life_active`
- `internal/server/state_view.go:147` — `view.Tokens["bp_shared_life_active"] = sharedLifeCount`
- `internal/server/state_view.go:149` — `delete(view.Tokens, "bp_shared_life_active")`

### `bp_shared_life_bound`
- `internal/server/state_view.go:153` — `view.Tokens["bp_shared_life_bound"] = sharedLifeBoundCount`
- `internal/server/state_view.go:155` — `delete(view.Tokens, "bp_shared_life_bound")`

### `bs_beast_soul`
- `internal/engine/beast_samurai_regression_test.go:104` — `if p1.Tokens["bs_beast_soul"] != 0 {`
- `internal/engine/beast_samurai_regression_test.go:105` — `t.Fatalf("expected initial beast soul=0, got %d", p1.Tokens["bs_beast_soul"])`
- `internal/engine/beast_samurai_regression_test.go:277` — `if got := p1.Tokens["bs_beast_soul"]; got != 1 {`
- `internal/engine/beast_samurai_regression_test.go:285` — `p1.Tokens["bs_beast_soul"] = 1`
- `internal/engine/beast_samurai_regression_test.go:313` — `if got := p1.Tokens["bs_beast_soul"]; got != 1 {`
- `internal/engine/beast_samurai_regression_test.go:323` — `p1.Tokens["bs_beast_soul"] = 1`
- `internal/engine/beast_samurai_regression_test.go:365` — `if got := p1.Tokens["bs_beast_soul"]; got != 1 {`
- `internal/engine/beast_samurai_regression_test.go:375` — `p1.Tokens["bs_beast_soul"] = 0`
- `internal/engine/beast_samurai_regression_test.go:427` — `p1.Tokens["bs_beast_soul"] = 1`
- `internal/engine/beast_samurai_regression_test.go:433` — `if got := p1.Tokens["bs_beast_soul"]; got != 0 {`
- `internal/engine/beast_samurai_regression_test.go:451` — `p1.Tokens["bs_beast_soul"] = 1`
- `internal/engine/beast_samurai_regression_test.go:514` — `p1.Tokens["bs_beast_soul"] = 2`
- `internal/engine/beast_samurai_regression_test.go:534` — `if got := p1.Tokens["bs_beast_soul"]; got != 3 {`
- `internal/engine/new_roles_helpers.go:1968` — `v := player.Tokens["bs_beast_soul"]`
- `internal/engine/new_roles_helpers.go:1975` — `player.Tokens["bs_beast_soul"] = v`
- `internal/engine/new_roles_helpers.go:1986` — `current := player.Tokens["bs_beast_soul"]`
- `internal/engine/new_roles_helpers.go:1997` — `player.Tokens["bs_beast_soul"] = current`
- `internal/engine/skills/handlers_beast_samurai.go:138` — `return getToken(ctx.User, "bs_beast_soul") >= 1`
- `internal/engine/skills/handlers_beast_samurai.go:158` — `if getToken(ctx.User, "bs_beast_soul") <= 0 {`
- `internal/engine/skills/handlers_beast_samurai.go:161` — `leftSoul := addToken(ctx.User, "bs_beast_soul", -1, 0, beastSamuraiBeastSoulCap)`
- `internal/engine/skills/handlers_beast_samurai.go:204` — `maxX := getToken(ctx.User, "bs_beast_soul")`
- `internal/engine/skills/handlers_beast_samurai.go:260` — `maxX := getToken(ctx.User, "bs_beast_soul")`

### `bs_one_strike_armed`
- `internal/engine/attack_role_hooks.go:170` — `if !e.isBeastSamurai(player) || player.Tokens == nil || player.Tokens["bs_one_strike_armed"] <= 0 || eventCtx == nil || eventCtx.AttackIn...`
- `internal/engine/attack_role_hooks.go:173` — `player.Tokens["bs_one_strike_armed"] = 0`
- `internal/engine/beast_samurai_regression_test.go:107` — `if p1.Tokens["bs_one_strike_armed"] != 0 {`
- `internal/engine/beast_samurai_regression_test.go:108` — `t.Fatalf("expected initial one-strike flag=0, got %d", p1.Tokens["bs_one_strike_armed"])`
- `internal/engine/beast_samurai_regression_test.go:156` — `p1.Tokens["bs_one_strike_armed"] = 1`
- `internal/engine/beast_samurai_regression_test.go:183` — `if p1.Tokens["bs_one_strike_armed"] != 0 {`
- `internal/engine/beast_samurai_regression_test.go:184` — `t.Fatalf("expected one-strike armed cleared, got %d", p1.Tokens["bs_one_strike_armed"])`
- `internal/engine/beast_samurai_regression_test.go:225` — `p1.Tokens["bs_one_strike_armed"] = 1`
- `internal/engine/phase_hooks.go:163` — `player.Tokens["bs_one_strike_armed"] = 0`
- `internal/engine/skills/handlers_beast_samurai.go:111` — `setToken(ctx.User, "bs_one_strike_armed", 1)`

### `bs_reversal_pending_x`
- `internal/engine/new_roles_helpers.go:2048` — `player.Tokens["bs_reversal_pending_x"] = 0`
- `internal/engine/skill_flow_beast_samurai.go:240` — `user.Tokens["bs_reversal_pending_x"] = x`

### `bs_zanshin`
- `internal/engine/beast_samurai_regression_test.go:101` — `if p1.Tokens["bs_zanshin"] != 0 {`
- `internal/engine/beast_samurai_regression_test.go:102` — `t.Fatalf("expected initial zanshin=0, got %d", p1.Tokens["bs_zanshin"])`
- `internal/engine/beast_samurai_regression_test.go:123` — `p1.Tokens["bs_zanshin"] = 3`
- `internal/engine/beast_samurai_regression_test.go:134` — `if got := p1.Tokens["bs_zanshin"]; got != 4 {`
- `internal/engine/beast_samurai_regression_test.go:310` — `if got := p1.Tokens["bs_zanshin"]; got != 1 {`
- `internal/engine/beast_samurai_regression_test.go:362` — `if got := p1.Tokens["bs_zanshin"]; got != 1 {`
- `internal/engine/beast_samurai_regression_test.go:428` — `p1.Tokens["bs_zanshin"] = 0`
- `internal/engine/beast_samurai_regression_test.go:436` — `if got := p1.Tokens["bs_zanshin"]; got != 1 {`
- `internal/engine/beast_samurai_regression_test.go:497` — `if got := p1.Tokens["bs_zanshin"]; got != 1 {`
- `internal/engine/new_roles_helpers.go:1932` — `v := player.Tokens["bs_zanshin"]`
- `internal/engine/new_roles_helpers.go:1939` — `player.Tokens["bs_zanshin"] = v`
- `internal/engine/new_roles_helpers.go:1957` — `player.Tokens["bs_zanshin"] = v`
- `internal/engine/skills/handlers_beast_samurai.go:82` — `now := addToken(ctx.User, "bs_zanshin", 1, 0, beastSamuraiZanshinCap)`
- `internal/engine/skills/handlers_beast_samurai.go:100` — `return getToken(ctx.User, "bs_zanshin") >= beastSamuraiZanshinCap`
- `internal/engine/skills/handlers_beast_samurai.go:107` — `if getToken(ctx.User, "bs_zanshin") < beastSamuraiZanshinCap {`
- `internal/engine/skills/handlers_beast_samurai.go:110` — `left := addToken(ctx.User, "bs_zanshin", -4, 0, beastSamuraiZanshinCap)`
- `internal/engine/skills/handlers_beast_samurai.go:162` — `nowZanshin := addToken(ctx.User, "bs_zanshin", 1, 0, beastSamuraiZanshinCap)`

### `bt_cocoon_count`
- `internal/engine/new_roles_helpers.go:1316` — `player.Tokens["bt_cocoon_count"] = count`
- `internal/server/state_view.go:160` — `view.Tokens["bt_cocoon_count"] = cocoonCount`
- `internal/server/state_view.go:162` — `delete(view.Tokens, "bt_cocoon_count")`

### `bt_pupa`
- `internal/engine/butterfly_dancer_regression_test.go:32` — `p1.Tokens["bt_pupa"] = 0`
- `internal/engine/butterfly_dancer_regression_test.go:36` — `p1.Tokens["bt_pupa"] = 2`
- `internal/engine/butterfly_dancer_regression_test.go:40` — `p1.Tokens["bt_pupa"] = 20`
- `internal/engine/butterfly_dancer_regression_test.go:115` — `if got := p1.Tokens["bt_pupa"]; got != 1 {`
- `internal/engine/new_roles_helpers.go:1273` — `v := player.Tokens["bt_pupa"]`
- `internal/engine/new_roles_helpers.go:1277` — `player.Tokens["bt_pupa"] = v`
- `internal/engine/new_roles_helpers.go:1292` — `player.Tokens["bt_pupa"] = v`

### `bt_wither_active`
- `internal/engine/butterfly_dancer_regression_test.go:346` — `p1.Tokens["bt_wither_active"] = 1`
- `internal/engine/butterfly_dancer_regression_test.go:412` — `if got := p1.Tokens["bt_wither_active"]; got != 1 {`
- `internal/engine/new_roles_helpers.go:1496` — `if p.Tokens["bt_wither_active"] > 0 {`
- `internal/engine/runtime_policy_hooks.go:55` — `if player.Tokens["bt_wither_active"] <= 0 {`
- `internal/engine/runtime_policy_hooks.go:58` — `player.Tokens["bt_wither_active"] = 0`
- `internal/engine/skill_flow_butterfly_dancer.go:612` — `user.Tokens["bt_wither_active"] = 1`

### `bt_wither_pending`
- `internal/engine/new_roles_helpers.go:1470` — `user.Tokens["bt_wither_pending"]++`
- `internal/engine/new_roles_helpers.go:1471` — `if user.Tokens["bt_wither_pending"] > 1 {`
- `internal/engine/skill_flow_butterfly_dancer.go:569` — `if user.Tokens["bt_wither_pending"] > 0 {`
- `internal/engine/skill_flow_butterfly_dancer.go:570` — `user.Tokens["bt_wither_pending"]--`
- `internal/engine/skill_flow_butterfly_dancer.go:572` — `if user.Tokens["bt_wither_pending"] > 0 {`
- `internal/engine/skill_flow_butterfly_dancer.go:613` — `if user.Tokens["bt_wither_pending"] > 0 {`
- `internal/engine/skill_flow_butterfly_dancer.go:614` — `user.Tokens["bt_wither_pending"]--`
- `internal/engine/skill_flow_butterfly_dancer.go:619` — `if user.Tokens["bt_wither_pending"] > 0 {`

### `bw_flame_form`
- `internal/server/state_view.go:96` — `delete(view.Tokens, "bw_flame_form")`

### `bw_flame_release_pending`
- `internal/engine/blaze_witch_skill_regression_test.go:176` — `p1.Tokens["bw_flame_release_pending"] = 1`
- `internal/engine/blaze_witch_skill_regression_test.go:187` — `if got := p1.Tokens["bw_flame_release_pending"]; got != 0 {`
- `internal/engine/phase_hooks.go:90` — `if player.Tokens["bw_flame_release_pending"] <= 0 {`
- `internal/engine/phase_hooks.go:95` — `player.Tokens["bw_flame_release_pending"] = 0`
- `internal/engine/skills/handlers_roles_18_22.go:1124` — `setToken(ctx.User, "bw_flame_release_pending", 1)`

### `bw_mana_inversion_lock`
- `internal/engine/damage_role_runtime_hooks.go:169` — `target.Tokens["bw_mana_inversion_lock"] = 0`
- `internal/engine/skills/handlers_roles_18_22.go:1242` — `if getToken(ctx.User, "bw_mana_inversion_lock") > 0 {`
- `internal/engine/skills/handlers_roles_18_22.go:1278` — `setToken(ctx.User, "bw_mana_inversion_lock", 1)`

### `bw_pain_link_pending_discard`
- `internal/engine/new_roles_helpers.go:3108` — `if e.isBlazeWitch(source) && source.Tokens != nil && source.Tokens["bw_pain_link_pending_discard"] > 0 {`
- `internal/engine/new_roles_helpers.go:3114` — `source.Tokens["bw_pain_link_pending_discard"] = 0`
- `internal/engine/skills/handlers_roles_18_22.go:1229` — `setToken(ctx.User, "bw_pain_link_pending_discard", 1)`

### `bw_pain_link_pending_hits`
- `internal/engine/new_roles_helpers.go:3109` — `if source.Tokens["bw_pain_link_pending_hits"] > 0 {`
- `internal/engine/new_roles_helpers.go:3110` — `source.Tokens["bw_pain_link_pending_hits"]--`
- `internal/engine/new_roles_helpers.go:3112` — `if source.Tokens["bw_pain_link_pending_hits"] <= 0 {`
- `internal/engine/new_roles_helpers.go:3113` — `source.Tokens["bw_pain_link_pending_hits"] = 0`
- `internal/engine/skills/handlers_roles_18_22.go:1230` — `setToken(ctx.User, "bw_pain_link_pending_hits", 2)`

### `bw_rebirth`
- `internal/engine/blaze_witch_skill_regression_test.go:103` — `p1.Tokens["bw_rebirth"] = 1`
- `internal/engine/blaze_witch_skill_regression_test.go:115` — `if got := p1.Tokens["bw_rebirth"]; got != 1 {`
- `internal/engine/blaze_witch_skill_regression_test.go:148` — `if got := p1.Tokens["bw_rebirth"]; got != 1 {`
- `internal/engine/blaze_witch_skill_regression_test.go:152` — `p1.Tokens["bw_rebirth"] = 4`
- `internal/engine/blaze_witch_skill_regression_test.go:158` — `if got := p1.Tokens["bw_rebirth"]; got != 4 {`
- `internal/engine/blaze_witch_skill_regression_test.go:204` — `p1.Tokens["bw_rebirth"] = 0`
- `internal/engine/blaze_witch_skill_regression_test.go:208` — `p1.Tokens["bw_rebirth"] = 1`
- `internal/engine/blaze_witch_skill_regression_test.go:212` — `p1.Tokens["bw_rebirth"] = 3`
- `internal/engine/blaze_witch_skill_regression_test.go:230` — `p1.Tokens["bw_rebirth"] = 1`
- `internal/engine/morale_loss_runtime.go:56` — `before := victim.Tokens["bw_rebirth"]`
- `internal/engine/morale_loss_runtime.go:57` — `victim.Tokens["bw_rebirth"]++`
- `internal/engine/morale_loss_runtime.go:58` — `if victim.Tokens["bw_rebirth"] > 4 {`
- `internal/engine/morale_loss_runtime.go:59` — `victim.Tokens["bw_rebirth"] = 4`
- `internal/engine/morale_loss_runtime.go:61` — `if victim.Tokens["bw_rebirth"] != before {`
- `internal/engine/morale_loss_runtime.go:62` — `e.Log(fmt.Sprintf("%s 的 [永生银时计] 触发，重生+1（当前%d）", victim.Name, victim.Tokens["bw_rebirth"]))`
- `internal/engine/player_hand_rules.go:42` — `maxHand += player.Tokens["bw_rebirth"] - 2`
- `internal/engine/skills/handlers_roles_18_22.go:1077` — `return getToken(ctx.User, "bw_rebirth") > 0`
- `internal/engine/skills/handlers_roles_18_22.go:1091` — `if getToken(ctx.User, "bw_rebirth") <= 0 {`
- `internal/engine/skills/handlers_roles_18_22.go:1094` — `addToken(ctx.User, "bw_rebirth", -1, 0, 4)`

### `bw_substitute_lock`
- `internal/engine/damage_role_runtime_hooks.go:168` — `target.Tokens["bw_substitute_lock"] = 0`
- `internal/engine/skills/handlers_roles_18_22.go:1145` — `if getToken(ctx.User, "bw_substitute_lock") > 0 {`
- `internal/engine/skills/handlers_roles_18_22.go:1175` — `setToken(ctx.User, "bw_substitute_lock", 1)`
- `internal/engine/skills/handlers_roles_18_22.go:1189` — `setToken(ctx.User, "bw_substitute_lock", 0)`

### `crk_blood_mark`
- `internal/engine/crimson_knight_bloody_prayer_regression_test.go:80` — `if got := p1.Tokens["crk_blood_mark"]; got != 1 {`
- `internal/engine/crimson_knight_killing_feast_regression_test.go:38` — `p1.Tokens["crk_blood_mark"] = 1`
- `internal/engine/crimson_knight_killing_feast_regression_test.go:88` — `if got := p1.Tokens["crk_blood_mark"]; got != 0 {`
- `internal/engine/crimson_knight_killing_feast_regression_test.go:116` — `p1.Tokens["crk_blood_mark"] = 1`
- `internal/engine/crimson_knight_killing_feast_regression_test.go:171` — `p1.Tokens["crk_blood_mark"] = 1`
- `internal/engine/crimson_knight_killing_feast_regression_test.go:217` — `p1.Tokens["crk_blood_mark"] = 1`
- `internal/engine/crimson_knight_killing_feast_regression_test.go:229` — `p1.Tokens["crk_blood_mark"] = 1`
- `internal/engine/skill_flow_crimson_knight.go:40` — `user.Tokens["crk_blood_mark"]++`
- `internal/engine/skill_flow_crimson_knight.go:41` — `if user.Tokens["crk_blood_mark"] > 3 {`
- `internal/engine/skill_flow_crimson_knight.go:42` — `user.Tokens["crk_blood_mark"] = 3`
- `internal/engine/skills/handlers_roles_18_22.go:265` — `return getToken(ctx.User, "crk_blood_mark") > 0`
- `internal/engine/skills/handlers_roles_18_22.go:269` — `if getToken(ctx.User, "crk_blood_mark") <= 0 {`
- `internal/engine/skills/handlers_roles_18_22.go:272` — `addToken(ctx.User, "crk_blood_mark", -1, 0, 3)`
- `internal/engine/skills/handlers_roles_18_22.go:341` — `if getToken(ctx.User, "crk_blood_mark") <= 0 {`
- `internal/engine/skills/handlers_roles_18_22.go:360` — `if getToken(ctx.User, "crk_blood_mark") <= 0 {`
- `internal/engine/skills/handlers_roles_18_22.go:363` — `addToken(ctx.User, "crk_blood_mark", -1, 0, 3)`

### `crk_hot_form`
- `internal/server/state_view.go:94` — `delete(view.Tokens, "crk_hot_form")`

### `css_blood`
- `internal/engine/counter_attack_action_gating_regression_test.go:79` — `if got := p2.Tokens["css_blood"]; got != 1 {`
- `internal/engine/crimson_sword_spirit_config_regression_test.go:24` — `p1.Tokens["css_blood"] = 2`
- `internal/engine/crimson_sword_spirit_config_regression_test.go:67` — `p1.Tokens["css_blood"] = 2`
- `internal/engine/crimson_sword_spirit_config_regression_test.go:84` — `if p1.Tokens["css_blood"] != 0 {`
- `internal/engine/crimson_sword_spirit_config_regression_test.go:85` — `t.Fatalf("expected blood rose spend 2 blood, got %d", p1.Tokens["css_blood"])`
- `internal/engine/crimson_sword_spirit_regression_test.go:20` — `p1.Tokens["css_blood"] = 1`
- `internal/engine/exclusive_skill_card_regression_test.go:169` — `p1.Tokens["css_blood"] = 4`
- `internal/engine/exclusive_skill_card_regression_test.go:189` — `if p1.Tokens["css_blood"] != 3 {`
- `internal/engine/exclusive_skill_card_regression_test.go:190` — `t.Fatalf("expected blood trimmed to 3, got %d", p1.Tokens["css_blood"])`
- `internal/engine/magic_swordsman_prayer_css_bugfix_regression_test.go:132` — `p1.Tokens["css_blood"] = 1`
- `internal/engine/magic_swordsman_prayer_css_bugfix_regression_test.go:181` — `p1.Tokens["css_blood"] = 1`
- `internal/engine/new_roles_helpers.go:1053` — `cur := player.Tokens["css_blood"]`
- `internal/engine/new_roles_helpers.go:1062` — `player.Tokens["css_blood"] = cur`
- `internal/engine/phase_hooks.go:236` — `if player.Tokens["css_blood"] > 3 {`
- `internal/engine/phase_hooks.go:237` — `player.Tokens["css_blood"] = 3`
- `internal/engine/skills/handlers_new_roles.go:1376` — `return getToken(ctx.User, "css_blood") > 0`
- `internal/engine/skills/handlers_new_roles.go:1380` — `if getToken(ctx.User, "css_blood") <= 0 {`
- `internal/engine/skills/handlers_new_roles.go:1396` — `return getToken(ctx.User, "css_blood") >= 2`
- `internal/engine/skills/handlers_new_roles.go:1403` — `if getToken(ctx.User, "css_blood") < 2 {`
- `internal/engine/skills/handlers_new_roles.go:1464` — `return getToken(ctx.User, "css_blood") > 0`
- `internal/engine/skills/handlers_new_roles.go:1468` — `if getToken(ctx.User, "css_blood") <= 0 {`
- `internal/engine/skills/handlers_new_roles.go:1615` — `cur := p.Tokens["css_blood"] + delta`
- `internal/engine/skills/handlers_new_roles.go:1622` — `p.Tokens["css_blood"] = cur`

### `css_blood_barrier_lock`
- `internal/engine/damage_role_runtime_hooks.go:161` — `target.Tokens["css_blood_barrier_lock"] = 0`
- `internal/engine/skills/handlers_new_roles.go:1461` — `if getToken(ctx.User, "css_blood_barrier_lock") > 0 {`
- `internal/engine/skills/handlers_new_roles.go:1471` — `setToken(ctx.User, "css_blood_barrier_lock", 1)`

### `css_blood_cap`
- `internal/engine/exclusive_skill_card_regression_test.go:168` — `p1.Tokens["css_blood_cap"] = 4`
- `internal/engine/exclusive_skill_card_regression_test.go:186` — `if p1.Tokens["css_blood_cap"] != 3 {`
- `internal/engine/exclusive_skill_card_regression_test.go:187` — `t.Fatalf("expected blood cap reset to 3, got %d", p1.Tokens["css_blood_cap"])`
- `internal/engine/new_roles_helpers.go:1040` — `if v := player.Tokens["css_blood_cap"]; v > 0 {`
- `internal/engine/phase_hooks.go:235` — `player.Tokens["css_blood_cap"] = 3`
- `internal/engine/skill_flow_crimson_sword_spirit.go:72` — `user.Tokens["css_blood_cap"] = 3`
- `internal/engine/skill_flow_crimson_sword_spirit.go:79` — `user.Tokens["css_blood_cap"] = 4`
- `internal/engine/skills/handlers_new_roles.go:1611` — `capV := p.Tokens["css_blood_cap"]`

### `element`
- `internal/engine/action_selection_prompt_options_test.go:92` — `p1.Tokens["element"] = 3`
- `internal/engine/action_selection_prompt_options_test.go:252` — `p1.Tokens["element"] = 3`
- `internal/engine/action_skill_runtime_rules.go:112` — `element = player.Tokens["element"]`
- `internal/engine/elementalist_regression_test.go:150` — `p1.Tokens["element"] = 2`
- `internal/engine/elementalist_regression_test.go:166` — `p1.Tokens["element"] = 3`
- `internal/engine/elementalist_regression_test.go:174` — `if got := p1.Tokens["element"]; got != 0 {`
- `internal/engine/elementalist_regression_test.go:200` — `p1.Tokens["element"] = 3`
- `internal/engine/skills/handlers_new_roles.go:272` — `return getToken(ctx.User, "element") < 3`
- `internal/engine/skills/handlers_new_roles.go:276` — `v := addToken(ctx.User, "element", 1, 0, 3)`
- `internal/engine/skills/handlers_new_roles.go:284` — `return getToken(ctx.User, "element") >= 3`
- `internal/engine/skills/handlers_new_roles.go:291` — `if getToken(ctx.User, "element") < 3 {`
- `internal/engine/skills/handlers_new_roles.go:294` — `addToken(ctx.User, "element", -3, 0, 3)`
- `internal/server/available_skill_adapter.go:85` — `element = p.Tokens["element"]`
- `internal/server/elementalist_available_skills_test.go:36` — `p1.Tokens["element"] = 2`
- `internal/server/elementalist_available_skills_test.go:47` — `p1.Tokens["element"] = 3`

### `elf_blessing_count`
- `internal/server/state_view.go:121` — `view.Tokens["elf_blessing_count"] = blessings`

### `elf_elemental_shot_earth_pending`
- `internal/engine/new_roles_helpers.go:864` — `player.Tokens["elf_elemental_shot_earth_pending"] = 0`
- `internal/engine/new_roles_helpers.go:1024` — `if attacker.Tokens["elf_elemental_shot_earth_pending"] > 0 {`
- `internal/engine/new_roles_helpers.go:2982` — `if attacker.Tokens["elf_elemental_shot_earth_pending"] > 0 {`
- `internal/engine/new_roles_helpers.go:2983` — `attacker.Tokens["elf_elemental_shot_earth_pending"] = 0`
- `internal/engine/skill_flow_elf_archer.go:132` — `user.Tokens["elf_elemental_shot_earth_pending"] = 0`
- `internal/engine/skill_flow_elf_archer.go:145` — `user.Tokens["elf_elemental_shot_earth_pending"] = 1`

### `elf_elemental_shot_thunder_pending`
- `internal/engine/elf_holy_lancer_bugfix_regression_test.go:90` — `if p1.Tokens["elf_elemental_shot_thunder_pending"] != 0 {`
- `internal/engine/elf_holy_lancer_bugfix_regression_test.go:91` — `t.Fatalf("expected thunder shot to stop using token state, got %d", p1.Tokens["elf_elemental_shot_thunder_pending"])`

### `elf_elemental_shot_water_pending`
- `internal/engine/new_roles_helpers.go:863` — `player.Tokens["elf_elemental_shot_water_pending"] = 0`
- `internal/engine/new_roles_helpers.go:1021` — `if attacker.Tokens["elf_elemental_shot_water_pending"] > 0 {`
- `internal/engine/new_roles_helpers.go:2975` — `if attacker.Tokens["elf_elemental_shot_water_pending"] > 0 {`
- `internal/engine/new_roles_helpers.go:2976` — `attacker.Tokens["elf_elemental_shot_water_pending"] = 0`
- `internal/engine/skill_flow_elf_archer.go:131` — `user.Tokens["elf_elemental_shot_water_pending"] = 0`
- `internal/engine/skill_flow_elf_archer.go:139` — `user.Tokens["elf_elemental_shot_water_pending"] = 1`

### `elf_elemental_shot_wind_pending`
- `internal/engine/elf_archer_skill_regression_test.go:64` — `if p1.Tokens["elf_elemental_shot_wind_pending"] != 1 {`
- `internal/engine/elf_archer_skill_regression_test.go:65` — `t.Fatalf("expected wind pending token before combat resolves, got %d", p1.Tokens["elf_elemental_shot_wind_pending"])`
- `internal/engine/elf_archer_skill_regression_test.go:82` — `if p1.Tokens["elf_elemental_shot_wind_pending"] != 0 {`
- `internal/engine/elf_archer_skill_regression_test.go:83` — `t.Fatalf("expected wind pending token cleared after action end, got %d", p1.Tokens["elf_elemental_shot_wind_pending"])`
- `internal/engine/new_roles_helpers.go:865` — `player.Tokens["elf_elemental_shot_wind_pending"] = 0`
- `internal/engine/new_roles_helpers.go:1027` — `if attacker.Tokens["elf_elemental_shot_wind_pending"] > 0 {`
- `internal/engine/new_roles_helpers.go:3052` — `if player.Tokens != nil && player.Tokens["elf_elemental_shot_wind_pending"] > 0 {`
- `internal/engine/skill_flow_elf_archer.go:133` — `user.Tokens["elf_elemental_shot_wind_pending"] = 0`
- `internal/engine/skill_flow_elf_archer.go:141` — `user.Tokens["elf_elemental_shot_wind_pending"] = 1`

### `elf_ritual_form`
- `internal/server/state_view.go:105` — `delete(view.Tokens, "elf_ritual_form")`

### `elf_ritual_release_waiting`
- `internal/engine/phase_hooks.go:180` — `if countElfBlessings(player) != 0 || player.Tokens["elf_ritual_release_waiting"] != 0 {`
- `internal/engine/phase_hooks.go:186` — `player.Tokens["elf_ritual_release_waiting"] = 0`
- `internal/engine/phase_hooks.go:190` — `player.Tokens["elf_ritual_release_waiting"] = 1`
- `internal/engine/skill_flow_elf_archer.go:237` — `user.Tokens["elf_ritual_release_waiting"] = 0`

### `fighter_attack_start_skill_lock`
- `internal/engine/attack_role_hooks.go:93` — `player.Tokens["fighter_attack_start_skill_lock"] = 0`
- `internal/engine/skills/handlers_fighter.go:55` — `if getToken(ctx.User, "fighter_attack_start_skill_lock") > 0 {`
- `internal/engine/skills/handlers_fighter.go:69` — `setToken(ctx.User, "fighter_attack_start_skill_lock", 1)`
- `internal/engine/skills/handlers_fighter.go:190` — `if getToken(ctx.User, "fighter_attack_start_skill_lock") > 0 {`
- `internal/engine/skills/handlers_fighter.go:204` — `setToken(ctx.User, "fighter_attack_start_skill_lock", 2)`

### `fighter_charge_pending`
- `internal/engine/attack_miss_aftermath.go:38` — `if !force && attacker.Tokens["fighter_charge_pending"] <= 0 {`
- `internal/engine/attack_miss_aftermath.go:41` — `attacker.Tokens["fighter_charge_pending"] = 0`
- `internal/engine/attack_passive_runtime_hooks.go:81` — `attacker.Tokens["fighter_charge_pending"] = 0`
- `internal/engine/damage_role_runtime_hooks.go:91` — `if attacker.Tokens["fighter_charge_pending"] > 0 {`
- `internal/engine/fighter_regression_test.go:86` — `if got := p1.Tokens["fighter_charge_pending"]; got != 0 {`
- `internal/engine/fighter_regression_test.go:127` — `if got := p1.Tokens["fighter_charge_pending"]; got != 0 {`
- `internal/engine/fighter_regression_test.go:145` — `p1.Tokens["fighter_charge_pending"] = 1`
- `internal/engine/fighter_regression_test.go:194` — `if got := p1.Tokens["fighter_charge_pending"]; got != 0 {`
- `internal/engine/skills/handlers_fighter.go:70` — `setToken(ctx.User, "fighter_charge_pending", 1)`

### `fighter_hundred_dragon_form`
- `internal/server/state_view.go:101` — `delete(view.Tokens, "fighter_hundred_dragon_form")`

### `fighter_hundred_dragon_target_order`
- `internal/engine/fighter_regression_test.go:330` — `if got := p1.Tokens["fighter_hundred_dragon_target_order"]; got != 2 {`
- `internal/engine/fighter_regression_test.go:377` — `p1.Tokens["fighter_hundred_dragon_target_order"] = 2 // 锁定 p2`
- `internal/engine/fighter_regression_test.go:388` — `if got := p1.Tokens["fighter_hundred_dragon_target_order"]; got != 0 {`
- `internal/engine/fighter_regression_test.go:412` — `p1.Tokens["fighter_hundred_dragon_target_order"] = 2`
- `internal/engine/fighter_regression_test.go:421` — `if got := p1.Tokens["fighter_hundred_dragon_target_order"]; got != 0 {`
- `internal/engine/fighter_regression_test.go:442` — `p1.Tokens["fighter_hundred_dragon_target_order"] = 2`
- `internal/engine/fighter_regression_test.go:459` — `if got := p1.Tokens["fighter_hundred_dragon_target_order"]; got != 0 {`
- `internal/engine/fighter_regression_test.go:484` — `p1.Tokens["fighter_hundred_dragon_target_order"] = 2`
- `internal/engine/fighter_regression_test.go:496` — `if got := p1.Tokens["fighter_hundred_dragon_target_order"]; got != 0 {`
- `internal/engine/fighter_regression_test.go:514` — `p1.Tokens["fighter_hundred_dragon_target_order"] = 2`
- `internal/engine/fighter_regression_test.go:523` — `if got := p1.Tokens["fighter_hundred_dragon_target_order"]; got != 0 {`
- `internal/engine/fighter_regression_test.go:545` — `p1.Tokens["fighter_hundred_dragon_target_order"] = 2`
- `internal/engine/new_roles_helpers.go:570` — `order := player.Tokens["fighter_hundred_dragon_target_order"]`
- `internal/engine/new_roles_helpers.go:585` — `active := hasFighterHundredDragonForm(player) || player.Tokens["fighter_hundred_dragon_target_order"] > 0`
- `internal/engine/new_roles_helpers.go:587` — `player.Tokens["fighter_hundred_dragon_target_order"] = 0`
- `internal/engine/runtime_policy_hooks.go:418` — `lockedOrder := player.Tokens["fighter_hundred_dragon_target_order"]`
- `internal/engine/skill_flow_fighter.go:72` — `user.Tokens["fighter_hundred_dragon_target_order"] = targetOrder`
- `internal/engine/skills/handlers_fighter.go:165` — `setToken(ctx.User, "fighter_hundred_dragon_target_order", 0)`

### `fighter_qi`
- `internal/engine/attack_miss_aftermath.go:42` — `damage := attacker.Tokens["fighter_qi"]`
- `internal/engine/fighter_regression_test.go:83` — `if got := p1.Tokens["fighter_qi"]; got != 1 {`
- `internal/engine/fighter_regression_test.go:124` — `if got := p1.Tokens["fighter_qi"]; got != 1 {`
- `internal/engine/fighter_regression_test.go:144` — `p1.Tokens["fighter_qi"] = 3`
- `internal/engine/fighter_regression_test.go:229` — `if got := p1.Tokens["fighter_qi"]; got != 1 {`
- `internal/engine/fighter_regression_test.go:249` — `if got := p1.Tokens["fighter_qi"]; got != 1 {`
- `internal/engine/fighter_regression_test.go:284` — `if got := p1.Tokens["fighter_qi"]; got != 1 {`
- `internal/engine/fighter_regression_test.go:310` — `p1.Tokens["fighter_qi"] = 3`
- `internal/engine/fighter_regression_test.go:324` — `if got := p1.Tokens["fighter_qi"]; got != 0 {`
- `internal/engine/fighter_regression_test.go:591` — `p1.Tokens["fighter_qi"] = 2`
- `internal/engine/fighter_regression_test.go:631` — `if got := p1.Tokens["fighter_qi"]; got != 1 {`
- `internal/engine/skill_flow_fighter.go:41` — `selfDamage = user.Tokens["fighter_qi"]`
- `internal/engine/skills/handlers_fighter.go:58` — `return getToken(ctx.User, "fighter_qi") < fighterQiCap`
- `internal/engine/skills/handlers_fighter.go:65` — `if getToken(ctx.User, "fighter_qi") >= fighterQiCap {`
- `internal/engine/skills/handlers_fighter.go:68` — `qi := addToken(ctx.User, "fighter_qi", 1, 0, fighterQiCap)`
- `internal/engine/skills/handlers_fighter.go:86` — `if getToken(ctx.User, "fighter_qi") >= fighterQiCap {`
- `internal/engine/skills/handlers_fighter.go:101` — `if getToken(ctx.User, "fighter_qi") >= fighterQiCap {`
- `internal/engine/skills/handlers_fighter.go:114` — `qi := addToken(ctx.User, "fighter_qi", 1, 0, fighterQiCap)`
- `internal/engine/skills/handlers_fighter.go:135` — `if getToken(ctx.User, "fighter_qi") < 3 || ctx.Game == nil {`
- `internal/engine/skills/handlers_fighter.go:150` — `if getToken(ctx.User, "fighter_qi") < 3 {`
- `internal/engine/skills/handlers_fighter.go:163` — `qi := addToken(ctx.User, "fighter_qi", -3, 0, fighterQiCap)`
- `internal/engine/skills/handlers_fighter.go:193` — `return getToken(ctx.User, "fighter_qi") > 0`
- `internal/engine/skills/handlers_fighter.go:200` — `if getToken(ctx.User, "fighter_qi") <= 0 {`
- `internal/engine/skills/handlers_fighter.go:203` — `qi := addToken(ctx.User, "fighter_qi", -1, 0, fighterQiCap)`

### `fighter_qiburst_force_no_counter`
- `internal/engine/attack_role_hooks.go:105` — `if player == nil || player.Tokens == nil || player.Tokens["fighter_qiburst_force_no_counter"] <= 0 || eventCtx == nil || eventCtx.AttackI...`
- `internal/engine/attack_role_hooks.go:109` — `player.Tokens["fighter_qiburst_force_no_counter"] = 0`
- `internal/engine/fighter_regression_test.go:634` — `if got := p1.Tokens["fighter_qiburst_force_no_counter"]; got != 0 {`
- `internal/engine/skills/handlers_fighter.go:205` — `setToken(ctx.User, "fighter_qiburst_force_no_counter", 1)`

### `hb_auto_fill_done_turn`
- `internal/engine/holy_bow_regression_test.go:270` — `p1.Tokens["hb_auto_fill_done_turn"] = 0`
- `internal/engine/phase_hooks.go:124` — `player.Tokens["hb_auto_fill_done_turn"] = 0`
- `internal/engine/phase_hooks.go:271` — `if player.Tokens["hb_auto_fill_done_turn"] > 0 {`
- `internal/engine/phase_hooks.go:283` — `player.Tokens["hb_auto_fill_done_turn"] = 1`
- `internal/engine/phase_hooks.go:297` — `player.Tokens["hb_auto_fill_done_turn"] = 1`

### `hb_cannon`
- `internal/engine/holy_bow_regression_test.go:42` — `if got := p1.Tokens["hb_cannon"]; got != 1 {`
- `internal/engine/holy_bow_regression_test.go:810` — `p1.Tokens["hb_cannon"] = 1`
- `internal/engine/holy_bow_regression_test.go:854` — `if got := p1.Tokens["hb_cannon"]; got != 0 {`
- `internal/engine/new_roles_helpers.go:1732` — `v := player.Tokens["hb_cannon"]`
- `internal/engine/new_roles_helpers.go:1739` — `player.Tokens["hb_cannon"] = v`
- `internal/engine/skill_flow_holy_bow.go:782` — `user.Tokens["hb_cannon"] = holyBowCannon(user) - 1`
- `internal/engine/skills/handlers_holy_bow.go:67` — `v := user.Tokens["hb_cannon"]`
- `internal/engine/skills/handlers_holy_bow.go:74` — `user.Tokens["hb_cannon"] = v`

### `hb_faith`
- `internal/engine/holy_bow_regression_test.go:45` — `if got := p1.Tokens["hb_faith"]; got != 0 {`
- `internal/engine/holy_bow_regression_test.go:108` — `if got := p1.Tokens["hb_faith"]; got != 1 {`
- `internal/engine/holy_bow_regression_test.go:122` — `if got := p1.Tokens["hb_faith"]; got != 1 {`
- `internal/engine/holy_bow_regression_test.go:134` — `if got := p1.Tokens["hb_faith"]; got != 1 {`
- `internal/engine/holy_bow_regression_test.go:152` — `p1.Tokens["hb_faith"] = 0`
- `internal/engine/holy_bow_regression_test.go:290` — `if got := p1.Tokens["hb_faith"]; got != 1 {`
- `internal/engine/holy_bow_regression_test.go:811` — `p1.Tokens["hb_faith"] = 6`
- `internal/engine/holy_bow_regression_test.go:857` — `if got := p1.Tokens["hb_faith"]; got != 2 {`
- `internal/engine/new_roles_helpers.go:1696` — `v := player.Tokens["hb_faith"]`
- `internal/engine/new_roles_helpers.go:1703` — `player.Tokens["hb_faith"] = v`
- `internal/engine/new_roles_helpers.go:1721` — `player.Tokens["hb_faith"] = v`
- `internal/engine/skills/handlers_holy_bow.go:42` — `v := user.Tokens["hb_faith"]`
- `internal/engine/skills/handlers_holy_bow.go:49` — `user.Tokens["hb_faith"] = v`
- `internal/engine/skills/handlers_holy_bow.go:57` — `return addToken(user, "hb_faith", delta, 0, holyBowFaithCap)`

### `hb_form`
- `internal/server/state_view.go:97` — `delete(view.Tokens, "hb_form")`

### `hb_shard_miss_pending`
- `internal/engine/attack_miss_aftermath.go:92` — `if attacker.Tokens == nil || attacker.Tokens["hb_shard_miss_pending"] <= 0 {`
- `internal/engine/attack_miss_aftermath.go:95` — `attacker.Tokens["hb_shard_miss_pending"] = 0`
- `internal/engine/holy_bow_regression_test.go:367` — `if got := p1.Tokens["hb_shard_miss_pending"]; got != 0 {`
- `internal/engine/holy_bow_regression_test.go:438` — `if got := p1.Tokens["hb_shard_miss_pending"]; got != 0 {`
- `internal/engine/holy_bow_regression_test.go:533` — `if got := p1.Tokens["hb_shard_miss_pending"]; got != 0 {`
- `internal/engine/new_roles_helpers.go:2937` — `if attacker.Tokens["hb_shard_miss_pending"] > 0 {`
- `internal/engine/new_roles_helpers.go:2938` — `attacker.Tokens["hb_shard_miss_pending"] = 0`
- `internal/engine/skill_flow_holy_bow.go:312` — `user.Tokens["hb_shard_miss_pending"] = 1`

### `hb_special_used_turn`
- `internal/engine/action_submission_runtime.go:109` — `player.Tokens["hb_special_used_turn"] = 1`
- `internal/engine/holy_bow_regression_test.go:269` — `p1.Tokens["hb_special_used_turn"] = 0`
- `internal/engine/phase_hooks.go:123` — `player.Tokens["hb_special_used_turn"] = 0`
- `internal/engine/phase_hooks.go:274` — `if player.Tokens["hb_special_used_turn"] <= 0 {`

### `hero_anger`
- `internal/engine/hero_regression_test.go:71` — `p1.Tokens["hero_anger"] = 1`
- `internal/engine/hero_regression_test.go:130` — `p1.Tokens["hero_anger"] = 1`
- `internal/engine/hero_regression_test.go:214` — `if got := p1.Tokens["hero_anger"]; got != 2 {`
- `internal/engine/hero_regression_test.go:273` — `if got := p1.Tokens["hero_anger"]; got != 2 {`
- `internal/engine/hero_regression_test.go:299` — `p1.Tokens["hero_anger"] = 0`
- `internal/engine/hero_regression_test.go:330` — `if got := p1.Tokens["hero_anger"]; got != 1 {`
- `internal/engine/hero_regression_test.go:355` — `p1.Tokens["hero_anger"] = 3`
- `internal/engine/hero_regression_test.go:398` — `if got := p1.Tokens["hero_anger"]; got != 4 {`
- `internal/engine/hero_regression_test.go:431` — `p1.Tokens["hero_anger"] = 1`
- `internal/engine/hero_regression_test.go:484` — `p1.Tokens["hero_anger"] = 1`
- `internal/engine/hero_regression_test.go:533` — `p1.Tokens["hero_anger"] = 1`
- `internal/engine/hero_regression_test.go:765` — `p1.Tokens["hero_anger"] = 1`
- `internal/engine/hero_regression_test.go:828` — `p1.Tokens["hero_anger"] = 1`
- `internal/engine/hero_regression_test.go:897` — `p1.Tokens["hero_anger"] = 0`
- `internal/engine/hero_regression_test.go:931` — `if got := p1.Tokens["hero_anger"]; got != 3 {`
- `internal/engine/skills/handlers_hero.go:42` — `return getToken(ctx.User, "hero_anger") > 0`
- `internal/engine/skills/handlers_hero.go:49` — `if getToken(ctx.User, "hero_anger") <= 0 {`
- `internal/engine/skills/handlers_hero.go:52` — `addToken(ctx.User, "hero_anger", -1, 0, heroTokenCap)`
- `internal/engine/skills/handlers_hero.go:111` — `anger := addToken(ctx.User, "hero_anger", magicCount, 0, heroTokenCap)`
- `internal/engine/skills/handlers_hero.go:178` — `return ctx != nil && ctx.User != nil && getToken(ctx.User, "hero_anger") > 0`
- `internal/engine/skills/handlers_hero.go:185` — `if getToken(ctx.User, "hero_anger") <= 0 {`
- `internal/engine/skills/handlers_hero.go:198` — `addToken(ctx.User, "hero_anger", -1, 0, heroTokenCap)`
- `internal/engine/skills/handlers_hero.go:228` — `anger := addToken(ctx.User, "hero_anger", 3, 0, heroTokenCap)`
- `internal/engine/soul_sorcerer_regression_test.go:745` — `if got := p3.Tokens["hero_anger"]; got != 3 {`

### `hero_calm_end_crystal_pending`
- `internal/engine/new_roles_helpers.go:3059` — `if e.isHero(player) && player.Tokens != nil && player.Tokens["hero_calm_end_crystal_pending"] > 0 && actionType == model.ActionAttack {`
- `internal/engine/new_roles_helpers.go:3060` — `player.Tokens["hero_calm_end_crystal_pending"]--`
- `internal/engine/new_roles_helpers.go:3061` — `if player.Tokens["hero_calm_end_crystal_pending"] < 0 {`
- `internal/engine/new_roles_helpers.go:3062` — `player.Tokens["hero_calm_end_crystal_pending"] = 0`
- `internal/engine/skills/handlers_hero.go:172` — `setToken(ctx.User, "hero_calm_end_crystal_pending", getToken(ctx.User, "hero_calm_end_crystal_pending")+1)`

### `hero_calm_force_no_counter`
- `internal/engine/attack_role_hooks.go:97` — `if player == nil || player.Tokens == nil || player.Tokens["hero_calm_force_no_counter"] <= 0 || eventCtx == nil || eventCtx.AttackInfo ==...`
- `internal/engine/attack_role_hooks.go:101` — `player.Tokens["hero_calm_force_no_counter"] = 0`
- `internal/engine/skills/handlers_hero.go:171` — `setToken(ctx.User, "hero_calm_force_no_counter", 1)`

### `hero_exhaustion_form`
- `internal/server/state_view.go:100` — `delete(view.Tokens, "hero_exhaustion_form")`

### `hero_exhaustion_release_pending`
- `internal/engine/hero_regression_test.go:223` — `if got := p1.Tokens["hero_exhaustion_release_pending"]; got != 0 {`
- `internal/engine/hero_regression_test.go:602` — `p1.Tokens["hero_exhaustion_release_pending"] = 1`
- `internal/engine/hero_regression_test.go:619` — `if got := p1.Tokens["hero_exhaustion_release_pending"]; got != 0 {`
- `internal/engine/hero_regression_test.go:660` — `p1.Tokens["hero_exhaustion_release_pending"] = 1`
- `internal/engine/runtime_policy_hooks.go:195` — `if player.Tokens["hero_exhaustion_release_pending"] <= 0 {`
- `internal/engine/runtime_policy_hooks.go:200` — `player.Tokens["hero_exhaustion_release_pending"] = 0`
- `internal/engine/skills/handlers_hero.go:133` — `setToken(ctx.User, "hero_exhaustion_release_pending", 1)`

### `hero_roar_active`
- `internal/engine/attack_miss_aftermath.go:18` — `if !force && attacker.Tokens["hero_roar_active"] <= 0 {`
- `internal/engine/attack_miss_aftermath.go:21` — `attacker.Tokens["hero_roar_active"] = 0`
- `internal/engine/attack_passive_runtime_hooks.go:108` — `attacker.Tokens["hero_roar_active"] = 0`
- `internal/engine/damage_role_runtime_hooks.go:82` — `if attacker.Tokens["hero_roar_active"] > 0 {`
- `internal/engine/hero_regression_test.go:109` — `if got := p1.Tokens["hero_roar_active"]; got != 0 {`
- `internal/engine/hero_regression_test.go:165` — `if got := p1.Tokens["hero_roar_active"]; got != 0 {`
- `internal/engine/skills/handlers_hero.go:53` — `setToken(ctx.User, "hero_roar_active", 1)`

### `hero_taunt_active_turn`
- `internal/engine/action_selection_prompt_options_test.go:342` — `p2.Tokens["hero_taunt_active_turn"] = 1`
- `internal/engine/runtime_policy_hooks.go:237` — `player.Tokens["hero_taunt_active_turn"] = 0`
- `internal/engine/runtime_policy_hooks.go:245` — `player.Tokens["hero_taunt_active_turn"] = 1`
- `internal/engine/runtime_policy_hooks.go:276` — `player.Tokens["hero_taunt_active_turn"] = 0`
- `internal/engine/runtime_policy_hooks.go:317` — `if player.Tokens["hero_taunt_active_turn"] <= 0 || player.Tokens["arbiter_forced_doomsday_pending"] > 0 {`
- `internal/engine/runtime_policy_hooks.go:355` — `if player.Tokens["hero_taunt_active_turn"] <= 0 || player.Tokens["arbiter_forced_doomsday_pending"] > 0 {`

### `hero_wisdom`
- `internal/engine/attack_miss_aftermath.go:22` — `wisdom := attacker.Tokens["hero_wisdom"] + 1`
- `internal/engine/attack_miss_aftermath.go:26` — `attacker.Tokens["hero_wisdom"] = wisdom`
- `internal/engine/hero_regression_test.go:112` — `if got := p1.Tokens["hero_wisdom"]; got != 0 {`
- `internal/engine/hero_regression_test.go:162` — `if got := p1.Tokens["hero_wisdom"]; got != 1 {`
- `internal/engine/hero_regression_test.go:276` — `if got := p1.Tokens["hero_wisdom"]; got != 2 {`
- `internal/engine/hero_regression_test.go:300` — `p1.Tokens["hero_wisdom"] = 0`
- `internal/engine/hero_regression_test.go:333` — `if got := p1.Tokens["hero_wisdom"]; got != 2 {`
- `internal/engine/hero_regression_test.go:356` — `p1.Tokens["hero_wisdom"] = 0`
- `internal/engine/hero_regression_test.go:715` — `p1.Tokens["hero_wisdom"] = 4`
- `internal/engine/hero_regression_test.go:744` — `if got := p1.Tokens["hero_wisdom"]; got != 0 {`
- `internal/engine/skills/handlers_hero.go:114` — `before := getToken(ctx.User, "hero_wisdom")`
- `internal/engine/skills/handlers_hero.go:115` — `after := addToken(ctx.User, "hero_wisdom", waterCount, 0, heroTokenCap)`
- `internal/engine/skills/handlers_hero.go:157` — `return getToken(ctx.User, "hero_wisdom") >= 4`
- `internal/engine/skills/handlers_hero.go:164` — `if getToken(ctx.User, "hero_wisdom") < 4 {`
- `internal/engine/skills/handlers_hero.go:167` — `addToken(ctx.User, "hero_wisdom", -4, 0, heroTokenCap)`
- `internal/engine/skills/handlers_hero.go:199` — `wisdom := addToken(ctx.User, "hero_wisdom", 1, 0, heroTokenCap)`

### `holy_lancer_block_sacred_strike`
- `internal/engine/attack_role_hooks.go:70` — `player.Tokens["holy_lancer_block_sacred_strike"] = 0`
- `internal/engine/response_resume_runtime.go:162` — `if user.Tokens["holy_lancer_block_sacred_strike"] != 0 {`
- `internal/engine/skill_flow_holy_lancer.go:77` — `user.Tokens["holy_lancer_block_sacred_strike"] = 1`
- `internal/engine/skills/handlers_new_roles.go:826` — `return getToken(ctx.User, "holy_lancer_block_sacred_strike") == 0`
- `internal/engine/skills/handlers_new_roles.go:854` — `setToken(ctx.User, "holy_lancer_block_sacred_strike", 1)`

### `holy_lancer_prayer_used_turn`
- `internal/engine/holy_lancer_earth_spear_regression_test.go:265` — `p1.Tokens["holy_lancer_prayer_used_turn"] = 1`
- `internal/engine/holy_lancer_earth_spear_regression_test.go:275` — `if got := p1.Tokens["holy_lancer_prayer_used_turn"]; got != 1 {`
- `internal/engine/holy_lancer_earth_spear_regression_test.go:287` — `if got := p1.Tokens["holy_lancer_prayer_used_turn"]; got != 0 {`
- `internal/engine/holy_lancer_earth_spear_regression_test.go:325` — `if got := p1.Tokens["holy_lancer_prayer_used_turn"]; got != 1 {`
- `internal/engine/phase_hooks.go:303` — `player.Tokens["holy_lancer_prayer_used_turn"] = 0`
- `internal/engine/skills/handlers_new_roles.go:838` — `if getToken(ctx.User, "holy_lancer_prayer_used_turn") > 0 {`
- `internal/engine/skills/handlers_new_roles.go:901` — `setToken(ctx.User, "holy_lancer_prayer_used_turn", 1)`

### `holy_lancer_sky_spear_no_counter`
- `internal/engine/attack_role_hooks.go:71` — `player.Tokens["holy_lancer_sky_spear_no_counter"] = 0`
- `internal/engine/attack_role_hooks.go:148` — `if !e.isHolyLancer(player) || player.Tokens == nil || player.Tokens["holy_lancer_sky_spear_no_counter"] <= 0 || eventCtx == nil || eventC...`
- `internal/engine/attack_role_hooks.go:152` — `player.Tokens["holy_lancer_sky_spear_no_counter"] = 0`
- `internal/engine/skills/handlers_new_roles.go:853` — `setToken(ctx.User, "holy_lancer_sky_spear_no_counter", 1)`

### `holy_sword_phase_end_pending`
- `internal/engine/combat.go:252` — `ctx.User.Tokens["holy_sword_phase_end_pending"] = 1`
- `internal/engine/turn_fsm_dispatcher.go:647` — `if player.Tokens != nil && player.Tokens["holy_sword_phase_end_pending"] > 0 {`
- `internal/engine/turn_fsm_dispatcher.go:649` — `player.Tokens["holy_sword_phase_end_pending"] = 0`

### `hom_burst_form`
- `internal/server/state_view.go:107` — `delete(view.Tokens, "hom_burst_form")`

### `hom_magic_rune`
- `internal/engine/crk_hom_skill_regression_test.go:229` — `p1.Tokens["hom_magic_rune"] = 0`
- `internal/engine/crk_hom_skill_regression_test.go:292` — `p1.Tokens["hom_magic_rune"] = 2`
- `internal/engine/crk_hom_skill_regression_test.go:359` — `p1.Tokens["hom_magic_rune"] = 1`
- `internal/engine/crk_hom_skill_regression_test.go:786` — `p1.Tokens["hom_magic_rune"] = 0`
- `internal/engine/skill_flow_war_homunculus.go:55` — `if user.Tokens["hom_magic_rune"] < flipCount {`
- `internal/engine/skill_flow_war_homunculus.go:58` — `user.Tokens["hom_magic_rune"] -= flipCount`
- `internal/engine/skill_flow_war_homunculus.go:65` — `user.Tokens["hom_magic_rune"] += flipCount`
- `internal/engine/skill_flow_war_homunculus.go:225` — `user.Tokens["hom_magic_rune"] = total - warRunes`
- `internal/engine/skills/handlers_roles_18_22.go:415` — `addToken(ctx.User, "hom_magic_rune", 1, 0, 99)`
- `internal/engine/skills/handlers_roles_18_22.go:498` — `if getToken(ctx.User, "hom_magic_rune") <= 0 {`
- `internal/engine/skills/handlers_roles_18_22.go:518` — `if getToken(ctx.User, "hom_magic_rune") <= 0 {`
- `internal/engine/skills/handlers_roles_18_22.go:541` — `magicRunes := getToken(ctx.User, "hom_magic_rune")`

### `hom_war_rune`
- `internal/engine/crk_hom_skill_regression_test.go:228` — `p1.Tokens["hom_war_rune"] = 3`
- `internal/engine/crk_hom_skill_regression_test.go:261` — `if p1.Tokens["hom_war_rune"] != 2 || p1.Tokens["hom_magic_rune"] != 1 {`
- `internal/engine/crk_hom_skill_regression_test.go:262` — `t.Fatalf("unexpected rune distribution: war=%d magic=%d", p1.Tokens["hom_war_rune"], p1.Tokens["hom_magic_rune"])`
- `internal/engine/crk_hom_skill_regression_test.go:358` — `p1.Tokens["hom_war_rune"] = 1`
- `internal/engine/crk_hom_skill_regression_test.go:785` — `p1.Tokens["hom_war_rune"] = 3`
- `internal/engine/crk_hom_skill_regression_test.go:840` — `if p1.Tokens["hom_war_rune"] != 1 || p1.Tokens["hom_magic_rune"] != 2 {`
- `internal/engine/crk_hom_skill_regression_test.go:841` — `t.Fatalf("unexpected rune flip result war=%d magic=%d", p1.Tokens["hom_war_rune"], p1.Tokens["hom_magic_rune"])`
- `internal/engine/skill_flow_war_homunculus.go:59` — `user.Tokens["hom_war_rune"] += flipCount`
- `internal/engine/skill_flow_war_homunculus.go:61` — `if user.Tokens["hom_war_rune"] < flipCount {`
- `internal/engine/skill_flow_war_homunculus.go:64` — `user.Tokens["hom_war_rune"] -= flipCount`
- `internal/engine/skill_flow_war_homunculus.go:224` — `user.Tokens["hom_war_rune"] = warRunes`
- `internal/engine/skill_flow_war_homunculus.go:226` — `e.Log(fmt.Sprintf("%s 的 [符文改造]：战纹=%d，魔纹=%d", user.Name, user.Tokens["hom_war_rune"], user.Tokens["hom_magic_rune"]))`
- `internal/engine/skills/handlers_roles_18_22.go:407` — `return getToken(ctx.User, "hom_war_rune") > 0`
- `internal/engine/skills/handlers_roles_18_22.go:411` — `if getToken(ctx.User, "hom_war_rune") <= 0 {`
- `internal/engine/skills/handlers_roles_18_22.go:414` — `addToken(ctx.User, "hom_war_rune", -1, 0, 99)`
- `internal/engine/skills/handlers_roles_18_22.go:430` — `if getToken(ctx.User, "hom_war_rune") <= 0 {`
- `internal/engine/skills/handlers_roles_18_22.go:450` — `if getToken(ctx.User, "hom_war_rune") <= 0 {`
- `internal/engine/skills/handlers_roles_18_22.go:465` — `warRunes := getToken(ctx.User, "hom_war_rune")`
- `internal/engine/skills/handlers_roles_18_22.go:581` — `totalRunes := getToken(ctx.User, "hom_war_rune") + getToken(ctx.User, "hom_magic_rune")`

### `judgment`
- `internal/engine/action_selection_prompt_options_test.go:303` — `p1.Tokens["judgment"] = 4`
- `internal/engine/arbiter_law_regression_test.go:96` — `if p1.Tokens["judgment"] != 4 {`
- `internal/engine/arbiter_law_regression_test.go:97` — `t.Fatalf("expected judgment to increase to 4 in startup, got %d", p1.Tokens["judgment"])`
- `internal/engine/arbiter_law_regression_test.go:123` — `p1.Tokens["judgment"] = 2`
- `internal/engine/arbiter_law_regression_test.go:153` — `if got := p1.Tokens["judgment"]; got != 2 {`
- `internal/engine/arbiter_law_regression_test.go:276` — `if got := p1.Tokens["judgment"]; got != 0 {`
- `internal/engine/arbiter_law_regression_test.go:318` — `if got := p1.Tokens["judgment"]; got != 1 {`
- `internal/engine/arbiter_law_regression_test.go:344` — `p1.Tokens["judgment"] = 2`
- `internal/engine/arbiter_law_regression_test.go:361` — `if got := p1.Tokens["judgment"]; got != 3 {`
- `internal/engine/arbiter_law_regression_test.go:425` — `if got := p1.Tokens["judgment"]; got != 0 {`
- `internal/engine/phase_hooks.go:137` — `if player.Tokens["judgment"] >= 4 {`
- `internal/engine/phase_hooks.go:140` — `player.Tokens["judgment"]++`
- `internal/engine/phase_hooks.go:141` — `e.Log(fmt.Sprintf("%s 处于审判形态，回合开始审判+1（当前%d）", player.Name, player.Tokens["judgment"]))`
- `internal/engine/runtime_policy_hooks.go:217` — `if player.Tokens["judgment"] < 4 || player.Tokens["arbiter_skip_forced_doomsday"] != 0 || player.Tokens["arbiter_forced_doomsday_done_tur...`
- `internal/engine/skills/handlers_new_roles.go:496` — `v := addToken(ctx.User, "judgment", 1, 0, 4)`
- `internal/engine/skills/handlers_new_roles.go:536` — `return getToken(ctx.User, "judgment") > 0`
- `internal/engine/skills/handlers_new_roles.go:543` — `dmg := getToken(ctx.User, "judgment")`
- `internal/engine/skills/handlers_new_roles.go:544` — `setToken(ctx.User, "judgment", 0)`
- `internal/engine/skills/handlers_new_roles.go:558` — `v := addToken(ctx.User, "judgment", 1, 0, 4)`
- `internal/server/elementalist_available_skills_test.go:74` — `p1.Tokens["judgment"] = 4`

### `mb_charge_count`
- `internal/engine/magic_bow_regression_test.go:244` — `if got := p1.Tokens["mb_charge_count"]; got != 2 {`
- `internal/engine/new_roles_helpers.go:1149` — `player.Tokens["mb_charge_count"] = magicBowChargeCount(player, "")`
- `internal/engine/skills/handlers_magic_bow.go:45` — `setToken(user, "mb_charge_count", magicBowChargeCount(user, ""))`
- `internal/server/state_view.go:126` — `view.Tokens["mb_charge_count"] = chargeCount`
- `internal/server/state_view.go:128` — `delete(view.Tokens, "mb_charge_count")`

### `mb_magic_pierce_pending`
- `internal/engine/attack_miss_aftermath.go:73` — `if attacker.Tokens == nil || attacker.Tokens["mb_magic_pierce_pending"] <= 0 {`
- `internal/engine/attack_miss_aftermath.go:76` — `attacker.Tokens["mb_magic_pierce_pending"] = 0`
- `internal/engine/magic_bow_regression_test.go:111` — `if got := p1.Tokens["mb_magic_pierce_pending"]; got != 0 {`
- `internal/engine/magic_bow_regression_test.go:489` — `if got := p1.Tokens["mb_magic_pierce_pending"]; got != 0 {`
- `internal/engine/magic_bow_regression_test.go:560` — `if got := p1.Tokens["mb_magic_pierce_pending"]; got != 0 {`
- `internal/engine/new_roles_helpers.go:1000` — `if attacker != nil && attacker.Tokens != nil && attacker.Tokens["mb_magic_pierce_pending"] > 0 {`
- `internal/engine/new_roles_helpers.go:3015` — `if attacker.Tokens["mb_magic_pierce_pending"] > 0 {`
- `internal/engine/new_roles_helpers.go:3017` — `attacker.Tokens["mb_magic_pierce_pending"] = 0`
- `internal/engine/new_roles_helpers.go:3035` — `attacker.Tokens["mb_magic_pierce_pending"] = 0`
- `internal/engine/skills/handlers_magic_bow.go:179` — `setToken(ctx.User, "mb_magic_pierce_pending", 1)`

### `mg_blasphemy_pending`
- `internal/engine/moon_goddess_regression_test.go:588` — `if got := moon.Tokens["mg_blasphemy_pending"]; got != 0 {`
- `internal/engine/moon_goddess_regression_test.go:653` — `moon.Tokens["mg_blasphemy_pending"] = 0`
- `internal/engine/new_roles_helpers.go:2785` — `if source.Tokens["mg_blasphemy_pending"] > 0 {`
- `internal/engine/new_roles_helpers.go:2802` — `source.Tokens["mg_blasphemy_pending"] = 1`
- `internal/engine/skill_flow_moon_goddess.go:496` — `user.Tokens["mg_blasphemy_pending"] = 0`
- `internal/engine/skill_flow_moon_goddess.go:520` — `user.Tokens["mg_blasphemy_pending"] = 0`

### `mg_blasphemy_used_turn`
- `internal/engine/moon_goddess_regression_test.go:585` — `if got := moon.Tokens["mg_blasphemy_used_turn"]; got != 1 {`
- `internal/engine/moon_goddess_regression_test.go:598` — `if got := moon.Tokens["mg_blasphemy_used_turn"]; got != 0 {`
- `internal/engine/moon_goddess_regression_test.go:652` — `moon.Tokens["mg_blasphemy_used_turn"] = 0`
- `internal/engine/new_roles_helpers.go:2782` — `if source.Tokens["mg_blasphemy_used_turn"] > 0 {`
- `internal/engine/skill_flow_moon_goddess.go:521` — `user.Tokens["mg_blasphemy_used_turn"] = 1`

### `mg_dark_form`
- `internal/server/state_view.go:102` — `delete(view.Tokens, "mg_dark_form")`

### `mg_dark_moon_count`
- `internal/engine/new_roles_helpers.go:2260` — `player.Tokens["mg_dark_moon_count"] = count`
- `internal/engine/skills/handlers_moon_goddess.go:55` — `user.Tokens["mg_dark_moon_count"] = count`
- `internal/server/state_view.go:140` — `view.Tokens["mg_dark_moon_count"] = darkMoonCount`
- `internal/server/state_view.go:142` — `delete(view.Tokens, "mg_dark_moon_count")`

### `mg_extra_turn_pending`
- `internal/engine/moon_goddess_regression_test.go:691` — `if got := moon.Tokens["mg_extra_turn_pending"]; got != 1 {`
- `internal/engine/moon_goddess_regression_test.go:702` — `if got := moon.Tokens["mg_extra_turn_pending"]; got != 0 {`
- `internal/engine/skill_flow_moon_goddess.go:599` — `user.Tokens["mg_extra_turn_pending"]++`
- `internal/engine/turn_progression_runtime.go:69` — `if player == nil || !e.isMoonGoddess(player) || player.Tokens == nil || player.Tokens["mg_extra_turn_pending"] <= 0 {`
- `internal/engine/turn_progression_runtime.go:72` — `player.Tokens["mg_extra_turn_pending"]--`
- `internal/engine/turn_progression_runtime.go:73` — `if player.Tokens["mg_extra_turn_pending"] < 0 {`
- `internal/engine/turn_progression_runtime.go:74` — `player.Tokens["mg_extra_turn_pending"] = 0`

### `mg_moon_cycle_used_turn`
- `internal/engine/moon_goddess_regression_test.go:211` — `if got := moon.Tokens["mg_moon_cycle_used_turn"]; got != 1 {`
- `internal/engine/moon_goddess_regression_test.go:317` — `moon.Tokens["mg_moon_cycle_used_turn"] = 0`
- `internal/engine/new_roles_helpers.go:2729` — `if player.Tokens["mg_moon_cycle_used_turn"] > 0 {`
- `internal/engine/new_roles_helpers.go:2744` — `player.Tokens["mg_moon_cycle_used_turn"] = 1`
- `internal/server/state_view.go:92` — `delete(view.Tokens, "mg_moon_cycle_used_turn")`

### `mg_new_moon`
- `internal/engine/moon_goddess_regression_test.go:734` — `moon.Tokens["mg_new_moon"] = 2`
- `internal/engine/moon_goddess_regression_test.go:760` — `if got := moon.Tokens["mg_new_moon"]; got != 1 {`
- `internal/engine/new_roles_helpers.go:2175` — `v := player.Tokens["mg_new_moon"]`
- `internal/engine/new_roles_helpers.go:2182` — `player.Tokens["mg_new_moon"] = v`
- `internal/engine/new_roles_helpers.go:2200` — `player.Tokens["mg_new_moon"] = v`
- `internal/engine/skills/handlers_moon_goddess.go:60` — `return addToken(user, "mg_new_moon", delta, 0, moonGoddessNewMoonCap)`
- `internal/engine/skills/handlers_moon_goddess.go:197` — `branch2 := getToken(ctx.User, "mg_new_moon") >= 1 &&`
- `internal/engine/skills/handlers_moon_goddess.go:211` — `if getToken(ctx.User, "mg_new_moon") >= 1 &&`

### `mg_next_attack_no_counter`
- `internal/engine/attack_role_hooks.go:128` — `if player == nil || player.Tokens == nil || player.Tokens["mg_next_attack_no_counter"] <= 0 || eventCtx == nil || eventCtx.AttackInfo == ...`
- `internal/engine/attack_role_hooks.go:132` — `player.Tokens["mg_next_attack_no_counter"]--`
- `internal/engine/attack_role_hooks.go:133` — `if player.Tokens["mg_next_attack_no_counter"] < 0 {`
- `internal/engine/attack_role_hooks.go:134` — `player.Tokens["mg_next_attack_no_counter"] = 0`
- `internal/engine/moon_goddess_regression_test.go:688` — `if got := moon.Tokens["mg_next_attack_no_counter"]; got != 1 {`
- `internal/engine/skill_flow_moon_goddess.go:598` — `user.Tokens["mg_next_attack_no_counter"]++`

### `mg_petrify`
- `internal/engine/moon_goddess_regression_test.go:675` — `moon.Tokens["mg_petrify"] = 3`
- `internal/engine/moon_goddess_regression_test.go:763` — `if got := moon.Tokens["mg_petrify"]; got != 1 {`
- `internal/engine/new_roles_helpers.go:2211` — `v := player.Tokens["mg_petrify"]`
- `internal/engine/new_roles_helpers.go:2218` — `player.Tokens["mg_petrify"] = v`
- `internal/engine/new_roles_helpers.go:2236` — `player.Tokens["mg_petrify"] = v`
- `internal/engine/skills/handlers_moon_goddess.go:64` — `return addToken(user, "mg_petrify", delta, 0, moonGoddessPetrifyCap)`
- `internal/engine/skills/handlers_moon_goddess.go:196` — `branch1 := getToken(ctx.User, "mg_petrify") >= 3`
- `internal/engine/skills/handlers_moon_goddess.go:208` — `if getToken(ctx.User, "mg_petrify") >= 3 {`

### `ml_dark_release_lock_turn`
- `internal/server/state_view.go:116` — `view.Tokens["ml_dark_release_lock_turn"] = 1`

### `ml_dark_release_next_attack_bonus`
- `internal/server/state_view.go:110` — `view.Tokens["ml_dark_release_next_attack_bonus"] = v`

### `ml_fullness_next_attack_bonus`
- `internal/server/state_view.go:113` — `view.Tokens["ml_fullness_next_attack_bonus"] = v`

### `ml_phantom_form`
- `internal/server/state_view.go:98` — `delete(view.Tokens, "ml_phantom_form")`

### `ml_stardust_locked_target_order`
- `internal/engine/new_roles_helpers.go:2886` — `lockedOrder := user.Tokens["ml_stardust_locked_target_order"]`
- `internal/engine/new_roles_helpers.go:2887` — `user.Tokens["ml_stardust_locked_target_order"] = 0`
- `internal/engine/skills/handlers_magic_lancer.go:99` — `setToken(ctx.User, "ml_stardust_locked_target_order", i+1)`
- `internal/engine/skills/handlers_magic_lancer.go:104` — `setToken(ctx.User, "ml_stardust_locked_target_order", 0)`

### `ml_stardust_morale_before`
- `internal/engine/new_roles_helpers.go:2862` — `before := user.Tokens["ml_stardust_morale_before"]`
- `internal/engine/new_roles_helpers.go:2866` — `user.Tokens["ml_stardust_morale_before"] = 0`
- `internal/engine/skills/handlers_magic_lancer.go:109` — `setToken(ctx.User, "ml_stardust_morale_before", before)`

### `ml_stardust_pending`
- `internal/engine/magic_lancer_regression_test.go:158` — `if got := p1.Tokens["ml_stardust_pending"]; got != 0 {`
- `internal/engine/new_roles_helpers.go:2852` — `if user.Tokens == nil || user.Tokens["ml_stardust_pending"] <= 0 {`
- `internal/engine/new_roles_helpers.go:2864` — `user.Tokens["ml_stardust_pending"] = 0`
- `internal/engine/new_roles_helpers.go:3132` — `source.Tokens["ml_stardust_pending"] > 0 &&`
- `internal/engine/skills/handlers_magic_lancer.go:107` — `setToken(ctx.User, "ml_stardust_pending", 1)`

### `ml_stardust_wait_discard`
- `internal/engine/hand_overflow_resolution.go:238` — `if e.isMagicLancer(player) && player.Tokens != nil && player.Tokens["ml_stardust_wait_discard"] > 0 {`
- `internal/engine/hand_overflow_resolution.go:315` — `if e.isMagicLancer(discardPlayer) && discardPlayer.Tokens != nil && discardPlayer.Tokens["ml_stardust_wait_discard"] > 0 {`
- `internal/engine/new_roles_helpers.go:2858` — `user.Tokens["ml_stardust_wait_discard"] = 1`
- `internal/engine/new_roles_helpers.go:2865` — `user.Tokens["ml_stardust_wait_discard"] = 0`
- `internal/engine/skills/handlers_magic_lancer.go:108` — `setToken(ctx.User, "ml_stardust_wait_discard", 0)`

### `ms_shadow_form`
- `internal/server/state_view.go:106` — `delete(view.Tokens, "ms_shadow_form")`

### `ms_yellow_spring_pending`
- `internal/engine/attack_role_hooks.go:88` — `player.Tokens["ms_yellow_spring_pending"] = 0`
- `internal/engine/attack_role_hooks.go:156` — `if !e.isMagicSwordsman(player) || player.Tokens == nil || player.Tokens["ms_yellow_spring_pending"] <= 0 || eventCtx == nil || eventCtx.A...`
- `internal/engine/new_roles_helpers.go:2994` — `if attacker.Tokens["ms_yellow_spring_pending"] > 0 {`
- `internal/engine/new_roles_helpers.go:2995` — `attacker.Tokens["ms_yellow_spring_pending"] = 0`
- `internal/engine/skills/handlers_new_roles.go:1333` — `setToken(ctx.User, "ms_yellow_spring_pending", 1)`

### `onmyoji_form`
- `internal/server/state_view.go:95` — `delete(view.Tokens, "onmyoji_form")`

### `onmyoji_ghost_fire`
- `internal/engine/onmyoji_rules_regression_test.go:44` — `if got := p1.Tokens["onmyoji_ghost_fire"]; got != 1 {`
- `internal/engine/onmyoji_rules_regression_test.go:82` — `p2.Tokens["onmyoji_ghost_fire"] = 1`
- `internal/engine/onmyoji_rules_regression_test.go:119` — `if got := p2.Tokens["onmyoji_ghost_fire"]; got != 3 {`
- `internal/engine/onmyoji_skill_flow_regression_test.go:35` — `p1.Tokens["onmyoji_ghost_fire"] = 3`
- `internal/engine/onmyoji_skill_flow_regression_test.go:56` — `if got := p1.Tokens["onmyoji_ghost_fire"]; got != 0 {`
- `internal/engine/onmyoji_skill_flow_regression_test.go:140` — `p1.Tokens["onmyoji_ghost_fire"] = 2 // 技能后变3`
- `internal/engine/onmyoji_skill_flow_regression_test.go:273` — `p1.Tokens["onmyoji_ghost_fire"] = 1`
- `internal/engine/onmyoji_skill_flow_regression_test.go:376` — `p2.Tokens["onmyoji_ghost_fire"] = 1`
- `internal/engine/onmyoji_skill_flow_regression_test.go:414` — `if got := p2.Tokens["onmyoji_ghost_fire"]; got != 3 {`
- `internal/engine/skill_flow_onmyoji.go:347` — `user.Tokens["onmyoji_ghost_fire"] = 0`
- `internal/engine/skill_flow_onmyoji.go:641` — `if player == nil || !e.isOnmyoji(player) || player.Tokens == nil || player.Tokens["onmyoji_ghost_fire"] < 3 {`
- `internal/engine/skill_flow_onmyoji.go:655` — `"ghost_fire":  player.Tokens["onmyoji_ghost_fire"],`
- `internal/engine/skill_flow_onmyoji.go:670` — `actor.Tokens["onmyoji_ghost_fire"]++`
- `internal/engine/skill_flow_onmyoji.go:671` — `if actor.Tokens["onmyoji_ghost_fire"] > 3 {`
- `internal/engine/skill_flow_onmyoji.go:672` — `actor.Tokens["onmyoji_ghost_fire"] = 3`
- `internal/engine/skill_flow_onmyoji.go:677` — `actor.Tokens["onmyoji_ghost_fire"]++`
- `internal/engine/skill_flow_onmyoji.go:678` — `if actor.Tokens["onmyoji_ghost_fire"] > 3 {`
- `internal/engine/skill_flow_onmyoji.go:679` — `actor.Tokens["onmyoji_ghost_fire"] = 3`
- `internal/engine/skill_flow_onmyoji.go:684` — `card.Damage = actor.Tokens["onmyoji_ghost_fire"]`
- `internal/engine/skills/handlers_roles_18_22.go:928` — `addToken(ctx.User, "onmyoji_ghost_fire", 1, 0, 3)`
- `internal/engine/skills/handlers_roles_18_22.go:967` — `gf := addToken(ctx.User, "onmyoji_ghost_fire", 1, 0, 3)`

### `plague_block_immortal`
- `internal/engine/plague_mage_skill_regression_test.go:148` — `if got := p1.Tokens["plague_block_immortal"]; got != 0 {`
- `internal/engine/skill_flow_plague_mage.go:276` — `user.Tokens["plague_block_immortal"] = 1`
- `internal/engine/skills/handlers_new_roles.go:1100` — `if getToken(ctx.User, "plague_block_immortal") > 0 {`
- `internal/engine/skills/handlers_new_roles.go:1101` — `setToken(ctx.User, "plague_block_immortal", 0)`
- `internal/engine/skills/handlers_new_roles.go:1178` — `setToken(ctx.User, "plague_block_immortal", 1)`

### `plague_outbreak_morale_drop_turn`
- `internal/engine/morale_loss_runtime.go:85` — `source.Tokens["plague_outbreak_morale_drop_turn"] = 1`
- `internal/engine/phase_hooks.go:204` — `if !e.isPlagueMage(player) || player.Tokens["plague_outbreak_morale_drop_turn"] <= 0 {`
- `internal/engine/phase_hooks.go:207` — `player.Tokens["plague_outbreak_morale_drop_turn"] = 0`
- `internal/engine/plague_mage_skill_regression_test.go:77` — `if got := p1.Tokens["plague_outbreak_morale_drop_turn"]; got != 0 {`

### `post_action_end_effect_magic`
- `internal/engine/turn_fsm_dispatcher.go:634` — `if player.Tokens["post_action_end_effect_magic"] > 0 {`
- `internal/engine/turn_fsm_dispatcher.go:638` — `player.Tokens["post_action_end_effect_magic"] = 0`
- `internal/engine/turn_fsm_dispatcher.go:695` — `player.Tokens["post_action_end_effect_magic"] = 1`
- `internal/engine/turn_fsm_dispatcher.go:697` — `player.Tokens["post_action_end_effect_magic"] = 0`

### `post_action_end_effect_pending`
- `internal/engine/turn_fsm_dispatcher.go:72` — `(player.TurnState.LastActionType != "" || (player.Tokens != nil && player.Tokens["post_action_end_effect_pending"] > 0)) {`
- `internal/engine/turn_fsm_dispatcher.go:88` — `} else if player != nil && player.Tokens != nil && player.Tokens["post_action_end_effect_pending"] > 0 {`
- `internal/engine/turn_fsm_dispatcher.go:632` — `if player.TurnState.LastActionType == "" && player.Tokens != nil && player.Tokens["post_action_end_effect_pending"] > 0 {`
- `internal/engine/turn_fsm_dispatcher.go:637` — `player.Tokens["post_action_end_effect_pending"] = 0`
- `internal/engine/turn_fsm_dispatcher.go:693` — `player.Tokens["post_action_end_effect_pending"] = 1`

### `prayer_form`
- `internal/server/state_view.go:93` — `delete(view.Tokens, "prayer_form")`

### `prayer_rune`
- `internal/engine/prayer_form_persist_regression_test.go:27` — `p1.Tokens["prayer_rune"] = 3`
- `internal/engine/prayer_form_persist_regression_test.go:34` — `if got := p1.Tokens["prayer_rune"]; got != 3 {`
- `internal/engine/skills/handlers_roles_18_22.go:67` — `v := addToken(ctx.User, "prayer_rune", 2, 0, 3)`
- `internal/engine/skills/handlers_roles_18_22.go:76` — `return hasForm(ctx.User, model.FormPrayerMasterPrayer) && getToken(ctx.User, "prayer_rune") > 0`
- `internal/engine/skills/handlers_roles_18_22.go:86` — `if getToken(ctx.User, "prayer_rune") <= 0 {`
- `internal/engine/skills/handlers_roles_18_22.go:93` — `addToken(ctx.User, "prayer_rune", -1, 0, 3)`
- `internal/engine/skills/handlers_roles_18_22.go:104` — `return hasForm(ctx.User, model.FormPrayerMasterPrayer) && getToken(ctx.User, "prayer_rune") > 0`
- `internal/engine/skills/handlers_roles_18_22.go:114` — `if getToken(ctx.User, "prayer_rune") <= 0 {`
- `internal/engine/skills/handlers_roles_18_22.go:117` — `addToken(ctx.User, "prayer_rune", -1, 0, 3)`

### `sc_power_count`
- `internal/engine/new_roles_helpers.go:1229` — `player.Tokens["sc_power_count"] = spiritCasterPowerCount(player, "")`
- `internal/server/state_view.go:133` — `view.Tokens["sc_power_count"] = powerCount`
- `internal/server/state_view.go:135` — `delete(view.Tokens, "sc_power_count")`

### `se_angel_soul_armed`
- `internal/engine/attack_role_hooks.go:77` — `player.Tokens["se_angel_soul_armed"] = 0`
- `internal/engine/new_roles_helpers.go:1868` — `player.Tokens["se_angel_soul_armed"] = 0`
- `internal/engine/new_roles_helpers.go:1892` — `if attacker.Tokens["se_angel_soul_armed"] > 0 {`
- `internal/engine/new_roles_helpers.go:1918` — `if attacker.Tokens["se_angel_soul_armed"] > 0 {`
- `internal/engine/skills/handlers_sword_emperor.go:208` — `setToken(ctx.User, "se_angel_soul_armed", 1)`
- `internal/engine/sword_emperor_regression_test.go:192` — `if got := p1.Tokens["se_angel_soul_armed"]; got != 0 {`

### `se_demon_soul_armed`
- `internal/engine/attack_role_hooks.go:78` — `player.Tokens["se_demon_soul_armed"] = 0`
- `internal/engine/new_roles_helpers.go:1869` — `player.Tokens["se_demon_soul_armed"] = 0`
- `internal/engine/new_roles_helpers.go:1899` — `if attacker.Tokens["se_demon_soul_armed"] > 0 {`
- `internal/engine/skills/handlers_sword_emperor.go:238` — `setToken(ctx.User, "se_demon_soul_armed", 1)`

### `se_guard_disabled_current_attack`
- `internal/engine/attack_role_hooks.go:76` — `player.Tokens["se_guard_disabled_current_attack"] = 0`
- `internal/engine/new_roles_helpers.go:1867` — `player.Tokens["se_guard_disabled_current_attack"] = 0`
- `internal/engine/new_roles_helpers.go:1881` — `if attacker.Tokens["se_guard_disabled_current_attack"] <= 0 &&`
- `internal/engine/skills/handlers_sword_emperor.go:207` — `setToken(ctx.User, "se_guard_disabled_current_attack", 1)`
- `internal/engine/skills/handlers_sword_emperor.go:237` — `setToken(ctx.User, "se_guard_disabled_current_attack", 1)`
- `internal/engine/sword_emperor_regression_test.go:64` — `if got := p1.Tokens["se_guard_disabled_current_attack"]; got != 0 {`
- `internal/engine/sword_emperor_regression_test.go:189` — `if got := p1.Tokens["se_guard_disabled_current_attack"]; got != 0 {`

### `se_sword_qi`
- `internal/engine/new_roles_helpers.go:1750` — `v := player.Tokens["se_sword_qi"]`
- `internal/engine/new_roles_helpers.go:1757` — `player.Tokens["se_sword_qi"] = v`
- `internal/engine/new_roles_helpers.go:1775` — `player.Tokens["se_sword_qi"] = v`
- `internal/engine/skills/handlers_sword_emperor.go:62` — `v := user.Tokens["se_sword_qi"]`
- `internal/engine/skills/handlers_sword_emperor.go:69` — `user.Tokens["se_sword_qi"] = v`
- `internal/engine/skills/handlers_sword_emperor.go:77` — `return addToken(user, "se_sword_qi", delta, 0, swordEmperorSwordQiCap)`
- `internal/engine/sword_emperor_regression_test.go:58` — `if got := p1.Tokens["se_sword_qi"]; got != 0 {`
- `internal/engine/sword_emperor_regression_test.go:105` — `if got := p1.Tokens["se_sword_qi"]; got != 1 {`
- `internal/engine/sword_emperor_regression_test.go:137` — `if got := p1.Tokens["se_sword_qi"]; got != 1 {`
- `internal/engine/sword_emperor_regression_test.go:183` — `if got := p1.Tokens["se_sword_qi"]; got != 1 {`
- `internal/engine/sword_emperor_regression_test.go:316` — `if got := p1.Tokens["se_sword_qi"]; got != 3 {`
- `internal/engine/sword_emperor_regression_test.go:336` — `p1.Tokens["se_sword_qi"] = 3`
- `internal/engine/sword_emperor_regression_test.go:405` — `if got := p1.Tokens["se_sword_qi"]; got != 1 {`
- `internal/engine/sword_emperor_regression_test.go:451` — `if got := p1.Tokens["se_sword_qi"]; got != 1 {`

### `se_sword_soul_count`
- `internal/engine/new_roles_helpers.go:1811` — `player.Tokens["se_sword_soul_count"] = swordEmperorSwordSoulCount(player)`
- `internal/engine/skills/handlers_sword_emperor.go:119` — `ctx.User.Tokens["se_sword_soul_count"] = len(cards) - 1`
- `internal/engine/sword_emperor_regression_test.go:61` — `if got := p1.Tokens["se_sword_soul_count"]; got != 0 {`
- `internal/engine/sword_emperor_regression_test.go:102` — `if got := p1.Tokens["se_sword_soul_count"]; got != 1 {`
- `internal/engine/sword_emperor_regression_test.go:134` — `if got := p1.Tokens["se_sword_soul_count"]; got != 3 {`

### `special_phase_end_dispatched`
- `internal/engine/action_submission_runtime.go:123` — `player.Tokens["special_phase_end_dispatched"] = 1`
- `internal/engine/turn_fsm_dispatcher.go:653` — `player.Tokens["special_phase_end_dispatched"] > 0 &&`
- `internal/engine/turn_fsm_dispatcher.go:656` — `player.Tokens["special_phase_end_dispatched"] = 0`

### `ss_blue_soul`
- `internal/engine/new_roles_helpers.go:2089` — `v := player.Tokens["ss_blue_soul"]`
- `internal/engine/new_roles_helpers.go:2096` — `player.Tokens["ss_blue_soul"] = v`
- `internal/engine/new_roles_helpers.go:2114` — `player.Tokens["ss_blue_soul"] = v`
- `internal/engine/skills/handlers_soul_sorcerer.go:30` — `return addToken(user, "ss_blue_soul", 0, 0, soulSorcererBlueCap)`
- `internal/engine/skills/handlers_soul_sorcerer.go:38` — `return addToken(user, "ss_blue_soul", delta, 0, soulSorcererBlueCap)`
- `internal/engine/soul_sorcerer_regression_test.go:68` — `if got := p1.Tokens["ss_blue_soul"]; got != 0 {`
- `internal/engine/soul_sorcerer_regression_test.go:181` — `if got := p1.Tokens["ss_blue_soul"]; got != 1 {`
- `internal/engine/soul_sorcerer_regression_test.go:194` — `p1.Tokens["ss_blue_soul"] = 1`
- `internal/engine/soul_sorcerer_regression_test.go:219` — `if got := p1.Tokens["ss_blue_soul"]; got != 0 {`
- `internal/engine/soul_sorcerer_regression_test.go:232` — `p1.Tokens["ss_blue_soul"] = 1`
- `internal/engine/soul_sorcerer_regression_test.go:261` — `if got := p1.Tokens["ss_blue_soul"]; got != 0 {`
- `internal/engine/soul_sorcerer_regression_test.go:414` — `p1.Tokens["ss_blue_soul"] = 3`
- `internal/engine/soul_sorcerer_regression_test.go:423` — `if got := p1.Tokens["ss_blue_soul"]; got != 0 {`
- `internal/engine/soul_sorcerer_regression_test.go:445` — `p1.Tokens["ss_blue_soul"] = 5`
- `internal/engine/soul_sorcerer_regression_test.go:458` — `if got := p1.Tokens["ss_blue_soul"]; got != 6 {`
- `internal/engine/soul_sorcerer_regression_test.go:481` — `p1.Tokens["ss_blue_soul"] = 2`
- `internal/engine/soul_sorcerer_regression_test.go:516` — `if got := p1.Tokens["ss_blue_soul"]; got != 0 {`
- `internal/engine/soul_sorcerer_regression_test.go:552` — `p1.Tokens["ss_blue_soul"] = 3`
- `internal/engine/soul_sorcerer_regression_test.go:588` — `if got := p1.Tokens["ss_blue_soul"]; got != 1 {`
- `internal/engine/soul_sorcerer_regression_test.go:627` — `p1.Tokens["ss_blue_soul"] = 3`
- `internal/engine/soul_sorcerer_regression_test.go:662` — `if got := p1.Tokens["ss_blue_soul"]; got != 1 {`
- `internal/engine/soul_sorcerer_regression_test.go:700` — `p1.Tokens["ss_blue_soul"] = 3`

### `ss_yellow_soul`
- `internal/engine/moon_goddess_regression_test.go:117` — `if got := soul.Tokens["ss_yellow_soul"]; got != 0 {`
- `internal/engine/new_roles_helpers.go:2125` — `v := player.Tokens["ss_yellow_soul"]`
- `internal/engine/new_roles_helpers.go:2132` — `player.Tokens["ss_yellow_soul"] = v`
- `internal/engine/new_roles_helpers.go:2150` — `player.Tokens["ss_yellow_soul"] = v`
- `internal/engine/skills/handlers_soul_sorcerer.go:34` — `return addToken(user, "ss_yellow_soul", 0, 0, soulSorcererYellowCap)`
- `internal/engine/skills/handlers_soul_sorcerer.go:42` — `return addToken(user, "ss_yellow_soul", delta, 0, soulSorcererYellowCap)`
- `internal/engine/soul_sorcerer_regression_test.go:71` — `if got := p1.Tokens["ss_yellow_soul"]; got != 0 {`
- `internal/engine/soul_sorcerer_regression_test.go:113` — `if got := soul.Tokens["ss_yellow_soul"]; got != 0 {`
- `internal/engine/soul_sorcerer_regression_test.go:150` — `if got := soul.Tokens["ss_yellow_soul"]; got != 1 {`
- `internal/engine/soul_sorcerer_regression_test.go:195` — `p1.Tokens["ss_yellow_soul"] = 0`
- `internal/engine/soul_sorcerer_regression_test.go:222` — `if got := p1.Tokens["ss_yellow_soul"]; got != 1 {`
- `internal/engine/soul_sorcerer_regression_test.go:233` — `p1.Tokens["ss_yellow_soul"] = 0`
- `internal/engine/soul_sorcerer_regression_test.go:264` — `if got := p1.Tokens["ss_yellow_soul"]; got != 1 {`
- `internal/engine/soul_sorcerer_regression_test.go:271` — `p1.Tokens["ss_yellow_soul"] = 2`
- `internal/engine/soul_sorcerer_regression_test.go:294` — `if got := p1.Tokens["ss_yellow_soul"]; got != 0 {`
- `internal/engine/soul_sorcerer_regression_test.go:318` — `p1.Tokens["ss_yellow_soul"] = 2`
- `internal/engine/soul_sorcerer_regression_test.go:350` — `p1.Tokens["ss_yellow_soul"] = 3`
- `internal/engine/soul_sorcerer_regression_test.go:361` — `if got := p1.Tokens["ss_yellow_soul"]; got != 0 {`
- `internal/engine/soul_sorcerer_regression_test.go:386` — `p1.Tokens["ss_yellow_soul"] = 3`
- `internal/engine/soul_sorcerer_regression_test.go:444` — `p1.Tokens["ss_yellow_soul"] = 5`
- `internal/engine/soul_sorcerer_regression_test.go:455` — `if got := p1.Tokens["ss_yellow_soul"]; got != 6 {`
- `internal/engine/soul_sorcerer_regression_test.go:482` — `p1.Tokens["ss_yellow_soul"] = 1`
- `internal/engine/soul_sorcerer_regression_test.go:553` — `p1.Tokens["ss_yellow_soul"] = 1`
- `internal/engine/soul_sorcerer_regression_test.go:628` — `p1.Tokens["ss_yellow_soul"] = 1`
- `internal/engine/soul_sorcerer_regression_test.go:701` — `p1.Tokens["ss_yellow_soul"] = 1`

### `valkyrie_military_glory_done_turn`
- `internal/engine/runtime_policy_hooks.go:181` — `player.Tokens["valkyrie_military_glory_done_turn"] = 1`

### `valkyrie_spirit`
- `internal/engine/new_roles_helpers.go:163` — `case player.Tokens["valkyrie_spirit"] > 0:`
- `internal/engine/runtime_policy_hooks.go:170` — `if player.Tokens["valkyrie_spirit"] <= 0 || player.Tokens["valkyrie_military_glory_done_turn"] > 0 {`
- `internal/engine/skill_flow_valkyrie.go:128` — `user.Tokens["valkyrie_spirit"] = 0`
- `internal/engine/skills/handlers_new_roles.go:169` — `return getToken(ctx.User, "valkyrie_spirit") > 0`
- `internal/engine/skills/handlers_new_roles.go:173` — `if getToken(ctx.User, "valkyrie_spirit") <= 0 {`
- `internal/engine/skills/handlers_new_roles.go:176` — `setToken(ctx.User, "valkyrie_spirit", 0)`
- `internal/engine/skills/handlers_new_roles.go:184` — `return ctx != nil && ctx.Trigger == model.TriggerOnTurnStart && ctx.Timing == model.TimingOnTurnStart && getToken(ctx.User, "valkyrie_spi...`
- `internal/engine/skills/handlers_new_roles.go:239` — `setToken(ctx.User, "valkyrie_spirit", 1)`
- `internal/engine/valkyrie_combo_regression_test.go:154` — `if p1.Tokens["valkyrie_spirit"] != 1 {`
- `internal/engine/valkyrie_combo_regression_test.go:155` — `t.Fatalf("expected valkyrie spirit=1 after heroic summon on self turn, got %d", p1.Tokens["valkyrie_spirit"])`
- `internal/engine/valkyrie_config_regression_test.go:114` — `if got := p2.Tokens["valkyrie_spirit"]; got != 0 {`
- `internal/engine/valkyrie_config_regression_test.go:136` — `p1.Tokens["valkyrie_spirit"] = 1`
- `internal/engine/valkyrie_config_regression_test.go:162` — `if got := p1.Tokens["valkyrie_spirit"]; got != 1 {`
- `internal/engine/valkyrie_config_regression_test.go:185` — `p1.Tokens["valkyrie_spirit"] = 1`
- `internal/engine/valkyrie_config_regression_test.go:197` — `if got := p1.Tokens["valkyrie_spirit"]; got != 0 {`

## 2. 动态下标 `player.Tokens[key]`

以下位置通过变量写入 `Tokens`：`applyRoleDefaults` 中 `cfg.tokens` 的键，或 `resetTurnScopedPlayerTokens` 中 `turnScopedResetKeys` 的键。

- `internal/engine/role_defaults.go:33` — `player.Tokens[key] = value`
- `internal/engine/turn_progression_runtime.go:86` — `player.Tokens[key] = 0`

## 3. `role_defaults.go` 默认 token 初始化（map 字面量）

下列行在 `roleDefaultConfigs` 的 `tokens: map[string]int{...}` 中出现；多数 key 亦在 §1 中有引用，此处标出**初始化默认值**来源。

- `internal/engine/role_defaults.go:44` — `"css_blood_cap": 3,`
- `internal/engine/role_defaults.go:45` — `"css_blood":     0,`
- `internal/engine/role_defaults.go:50` — `"prayer_rune": 0,`
- `internal/engine/role_defaults.go:57` — `"crk_blood_mark": 0,`
- `internal/engine/role_defaults.go:62` — `"hom_war_rune":   3,`
- `internal/engine/role_defaults.go:63` — `"hom_magic_rune": 0,`
- `internal/engine/role_defaults.go:72` — `"onmyoji_ghost_fire": 0,`
- `internal/engine/role_defaults.go:77` — `"bw_rebirth":               0,`
- `internal/engine/role_defaults.go:78` — `"bw_flame_release_pending": 0,`
- `internal/engine/role_defaults.go:83` — `"ml_stardust_wait_discard":  0,`
- `internal/engine/role_defaults.go:84` — `"ml_stardust_morale_before": 0,`
- `internal/engine/role_defaults.go:89` — `"sc_power_count": 0,`
- `internal/engine/role_defaults.go:94` — `"bd_inspiration":           0,`
- `internal/engine/role_defaults.go:95` — `"bd_descent_used_turn":     0,`
- `internal/engine/role_defaults.go:96` — `"bd_rousing_prompted_turn": 0,`
- `internal/engine/role_defaults.go:97` — `"bd_victory_prompted_turn": 0,`
- `internal/engine/role_defaults.go:103` — `"hero_anger":                      0,`
- `internal/engine/role_defaults.go:104` — `"hero_wisdom":                     0,`
- `internal/engine/role_defaults.go:105` — `"hero_exhaustion_release_pending": 0,`
- `internal/engine/role_defaults.go:106` — `"hero_roar_active":                0,`
- `internal/engine/role_defaults.go:107` — `"hero_calm_end_crystal_pending":   0,`
- `internal/engine/role_defaults.go:112` — `"fighter_qi":                          0,`
- `internal/engine/role_defaults.go:113` — `"fighter_hundred_dragon_target_order": 0,`
- `internal/engine/role_defaults.go:114` — `"fighter_attack_start_skill_lock":     0,`
- `internal/engine/role_defaults.go:115` — `"fighter_charge_pending":              0,`
- `internal/engine/role_defaults.go:116` — `"fighter_qiburst_force_no_counter":    0,`
- `internal/engine/role_defaults.go:123` — `"hb_cannon":              1,`
- `internal/engine/role_defaults.go:124` — `"hb_faith":               0,`
- `internal/engine/role_defaults.go:125` — `"hb_special_used_turn":   0,`
- `internal/engine/role_defaults.go:126` — `"hb_auto_fill_done_turn": 0,`
- `internal/engine/role_defaults.go:127` — `"hb_shard_miss_pending":  0,`
- `internal/engine/role_defaults.go:132` — `"se_sword_qi":                      0,`
- `internal/engine/role_defaults.go:133` — `"se_sword_soul_count":              0,`
- `internal/engine/role_defaults.go:134` — `"se_guard_disabled_current_attack": 0,`
- `internal/engine/role_defaults.go:135` — `"se_angel_soul_armed":              0,`
- `internal/engine/role_defaults.go:136` — `"se_demon_soul_armed":              0,`
- `internal/engine/role_defaults.go:141` — `"bs_zanshin":            0,`
- `internal/engine/role_defaults.go:142` — `"bs_beast_soul":         0,`
- `internal/engine/role_defaults.go:143` — `"bs_one_strike_armed":   0,`
- `internal/engine/role_defaults.go:144` — `"bs_reversal_pending_x": 0,`
- `internal/engine/role_defaults.go:154` — `"arbiter_law_inited": 1,`
- `internal/engine/role_defaults.go:159` — `"ss_blue_soul":   0,`
- `internal/engine/role_defaults.go:160` — `"ss_yellow_soul": 0,`
- `internal/engine/role_defaults.go:165` — `"mg_new_moon":               0,`
- `internal/engine/role_defaults.go:166` — `"mg_petrify":                0,`
- `internal/engine/role_defaults.go:167` — `"mg_dark_moon_count":        0,`
- `internal/engine/role_defaults.go:168` — `"mg_moon_cycle_used_turn":   0,`
- `internal/engine/role_defaults.go:169` — `"mg_blasphemy_used_turn":    0,`
- `internal/engine/role_defaults.go:170` — `"mg_blasphemy_pending":      0,`
- `internal/engine/role_defaults.go:171` — `"mg_next_attack_no_counter": 0,`
- `internal/engine/role_defaults.go:172` — `"mg_extra_turn_pending":     0,`
- `internal/engine/role_defaults.go:177` — `"bt_pupa":           0,`
- `internal/engine/role_defaults.go:178` — `"bt_cocoon_count":   0,`
- `internal/engine/role_defaults.go:179` — `"bt_wither_active":  0,`
- `internal/engine/role_defaults.go:180` — `"bt_wither_pending": 0,`
- `internal/engine/role_defaults.go:185` — `"bp_bleed_tick_done_turn": 0,`
