# Strict Mode Phase 1 分批任务清单（按文件包名拆卡）

本文是 `strict-mode-phase1-protocol-and-core.md` 的执行分解。目标是把“Engine 全量禁兜底”拆成可回滚的小卡片，避免一次性大改。

## 执行总则

1. 每张卡只改一个子域，不跨多个语义域并行。
2. 每张卡完成后必须执行最小验收，再进入下一张。
3. 若失败，先产出失败归因，不直接推进下一张。
4. 默认遵循 strict 规则: 不猜、不补、不吞，非法输入显式失败。

## 卡片 01：Resume Point 核心收口

- 文件范围:
  - `internal/model/resume_point.go`
  - `internal/engine/resume_point.go`
  - 对应测试文件
- 改造点:
  - 移除无前缀 resume point 兼容解析。
  - 收紧 `currentChoiceResumePoint` 的空值降级行为（禁止静默掩盖非法状态）。
- 完成标准:
  - 仅接受规范格式（`turn:` / `combat:` / `subflow:`）。
  - 非法输入显式失败并可测试断言。
- 最小验收:
  - `go test ./internal/model/... -count=1`
  - `go test ./internal/engine/... -run 'Test.*Resume.*' -count=1`

## 卡片 02：Choice Router 严格化

- 文件范围:
  - `internal/engine/choice_phase_router.go`
  - `internal/engine/basic_effect_choice.go`
  - `internal/engine/hand_overflow_*`
- 改造点:
  - 清理默认分支吞错和隐式回退。
  - 恢复点缺失时显式拒绝而非继续推进流程。
- 完成标准:
  - Choice 路由错误路径可观测。
- 最小验收:
  - `go test ./internal/engine/... -run 'Test.*Choice.*|Test.*Overflow.*' -count=1`

## 卡片 03：Interrupt 处理严格化

- 文件范围:
  - `internal/engine/interrupt_runtime.go`
  - `internal/engine/interrupt_prompt_*.go`
  - `internal/engine/interrupt_response_runtime.go`
- 改造点:
  - `Cmd` 与中断类型不匹配时统一显式错误。
  - 取消静默跳过和不透明 fallback。
- 完成标准:
  - 所有中断拒绝路径可复现可断言。
- 最小验收:
  - `go test ./internal/engine/... -run 'Test.*Interrupt.*|Test.*Prompt.*' -count=1`

## 卡片 04：Action Submission 收口

- 文件范围:
  - `internal/engine/action_submission_runtime.go`
  - `internal/engine/action_selection_policies.go`
  - `internal/engine/action_skill_runtime_rules.go`
- 改造点:
  - 动作类型、目标、上下文约束统一显式校验。
  - 去除“未知场景继续执行”的路径。
- 完成标准:
  - 非法 action 进入统一错误分支。
- 最小验收:
  - `go test ./internal/engine/... -run 'Test.*Action.*|Test.*Policy.*' -count=1`

## 卡片 05：Turn FSM 与主流程阶段严格化

- 文件范围:
  - `internal/engine/turn_fsm_dispatcher.go`
  - `internal/engine/flow_stage_runtime.go`
  - `internal/engine/response_resume_runtime.go`
- 改造点:
  - 阶段推进不再依赖兜底阶段猜测。
  - 非法阶段组合显式失败。
- 完成标准:
  - 阶段不变量有测试覆盖。
- 最小验收:
  - `go test ./internal/engine/... -run 'Test.*FSM.*|Test.*Stage.*|Test.*Resume.*' -count=1`

## 卡片 06：Draw / Damage / Combat 链路严格化

- 文件范围:
  - `internal/engine/draw_flow.go`
  - `internal/engine/pending_damage_runtime.go`
  - `internal/engine/combat*.go`
- 改造点:
  - 去除 draw/combat 的静默降级路径。
  - 上下文缺失时显式错误。
- 完成标准:
  - 伤害与摸牌链路在非法上下文下拒绝执行。
- 最小验收:
  - `go test ./internal/engine/... -run 'Test.*Combat.*|Test.*Draw.*|Test.*Damage.*' -count=1`

## 卡片 07：Special Action / Runtime 状态严格化

- 文件范围:
  - `internal/engine/special_action_runtime.go`
  - `internal/engine/discard_skill_runtime.go`
  - `internal/engine/exclusive_effect_runtime.go`
- 改造点:
  - 特殊行动条件不足时显式失败，不进行补偿执行。
  - 移除流程状态“默认值即合法”的假设。
- 完成标准:
  - 特殊行动与独有效果拒绝路径有测试断言。
- 最小验收:
  - `go test ./internal/engine/... -run 'Test.*Special.*|Test.*Exclusive.*|Test.*Discard.*' -count=1`

## 卡片 08：Skill Flow 第一批（角色技能流 1）

- 文件范围:
  - `internal/engine/skill_flow_assassin_hero.go`
  - `internal/engine/skill_flow_fighter.go`
  - `internal/engine/skill_flow_holy_bow.go`
  - `internal/engine/skill_flow_priest.go`
- 改造点:
  - 去除 waiting/resume 相关兜底。
  - 目标、形态、资源不足场景显式拒绝。
- 最小验收:
  - `go test ./internal/engine/... -run 'Test(Assassin|Hero|Fighter|HolyBow|Priest)' -count=1`

## 卡片 09：Skill Flow 第二批（角色技能流 2）

- 文件范围:
  - `internal/engine/skill_flow_beast_samurai.go`
  - `internal/engine/skill_flow_blood_priestess.go`
  - `internal/engine/skill_flow_butterfly_dancer.go`
  - `internal/engine/skill_flow_guardian_support.go`
- 改造点:
  - resume/context 兼容分支收口。
  - 取消空上下文自动跳过。
- 最小验收:
  - `go test ./internal/engine/... -run 'Test(Beast|BloodPriestess|Butterfly|Guardian)' -count=1`

## 卡片 10：Skill Flow 第三批（其余 skill_flow_*）

- 文件范围:
  - `internal/engine/skill_flow_*.go`（除卡片 08/09 已覆盖文件）
- 改造点:
  - 统一 strict 语义。
  - 对 default 分支进行显式错误化。
- 最小验收:
  - `go test ./internal/engine/... -count=1`

## 卡片 11：Engine 公共辅助与工具层收口

- 文件范围:
  - `internal/engine/runtimeutil/*.go`
  - `internal/engine/promptfmt/*.go`
  - `internal/engine/interface.go`
  - `internal/engine/action_summary.go`
- 改造点:
  - 公共 helper 不再隐式容错推进流程。
  - 错误向上抛出并由调用方处理。
- 最小验收:
  - `go test ./internal/engine/... -count=1`

## 卡片 12：协议边界严格化（与 engine 对齐）

- 文件范围:
  - `internal/server/message_dispatcher.go`
  - `internal/server/room_action_dispatcher.go`
  - `internal/server/protocol_adapter.go`
  - `internal/server/protocol/types.go`
- 改造点:
  - 未知 `Cmd`、非法 JSON、未知 room action、未知 action type 统一结构化错误。
- 最小验收:
  - `go test ./internal/server/... -count=1`

## 汇总验收（每 3 张卡后执行一次）

1. `go test ./... -run '^$'`
2. `go test ./internal/engine/... -count=1`
3. `go test ./internal/server/... -count=1`
4. `go test ./tests/... -count=1`

## 交付输出模板（每张卡）

1. 变更文件清单
2. 删除的兜底/兼容路径说明
3. 新增失败契约说明
4. 验收命令与结果
5. 风险与下一张卡建议
