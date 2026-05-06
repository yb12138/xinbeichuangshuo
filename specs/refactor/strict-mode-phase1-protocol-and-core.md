# 严格模式重构 Spec（Phase 1：协议边界 + Engine 全量禁兜底）

## 背景/问题
- 当前症状:
  - `internal/engine` 中存在防御性兜底、静默 return、默认分支吞错、旧格式兼容解析等路径。
  - 核心规则层在异常输入场景下有“继续执行”而非“显式失败”的行为，导致问题定位困难。
  - 协议层与规则层存在“兼容式宽松处理”，削弱了“规则即唯一真相”。
- 典型例子:
  - `currentChoiceResumePoint()` 在空引擎/空状态时返回 `""`（隐式降级）。
  - `ParseResumePoint*` 同时接受带前缀与无前缀输入（旧格式兼容）。
  - 协议入口存在未知命令/非法 JSON 直接 return 的场景。
- 业务风险:
  - 错误输入被吞掉后，行为偏差会在后续阶段放大。
  - 回归失败难以复现和归因，重构成本持续升高。
  - 新角色或新流程接入时继续叠加兜底，形成长期技术债。

## 范围
- 允许改动的文件:
  - `internal/engine/**/*.go`（全量纳入）
  - `internal/model/**/*.go`（与规则核心相关部分）
  - `internal/server/message_dispatcher.go`
  - `internal/server/room_action_dispatcher.go`
  - `internal/server/protocol_adapter.go`
  - `internal/server/ws_protocol.go`（如需新增错误命令常量）
  - `internal/server/protocol/types.go`（如需新增错误载荷 DTO）
  - 相关测试文件（`internal/engine/**/*_test.go`、`internal/model/**/*_test.go`、`internal/server/**/*_test.go`、`tests/**/*`）
- 禁止改动的文件:
  - `web/**`
  - `cmd/**`
  - 规则文档本体（`docs/rule.md`、`docs/action.md`、`docs/data_model.md`）除注释性引用外不变更语义

## 目标
1. 将 `internal/engine` 全目录切换为严格执行语义: 不猜、不补、不吞。
2. 移除核心规则层中的防御性兜底与旧格式兼容路径（含 resume point 解析路径）。
3. 协议边界对非法输入统一显式失败并可观测。
4. 建立“严格模式”回归契约，确保后续新增代码不得引入同类兜底。

## 非目标
1. 不进行前端页面重构。
2. 不进行性能优化专项。
3. 不进行无关目录的大规模重命名。
4. 不改变合法输入下的规则结果与时序语义。

## 约束
1. 核心规则层约束（`internal/engine` + `internal/model`）:
   - 禁止防御性兜底继续执行。
   - 禁止兼容旧格式输入（除明确标注且经审批的迁移窗口）。
   - 禁止通过空字符串/零值/`nil` 表示“错误但继续流程”。
2. 分支约束:
   - `switch default` 仅允许返回显式错误，禁止静默成功。
3. 错误可观测约束:
   - 所有拒绝路径必须可被测试断言。
   - 协议层拒绝路径必须给出结构化错误反馈（含最小上下文）。
4. 变更治理约束:
   - 按子域分批提交（建议: resume_point -> choice/router -> interrupt -> action -> skill_flow）。
   - 每批均需通过对应测试后再进入下一批。

## 严格模式约束
1. 规则输入不合法即失败，不允许“自动修正”。
2. 旧格式兼容默认关闭:
   - 例如 resume point 仅接受规范格式，不再接受历史无前缀输入。
3. 边界可用性逻辑不得污染核心规则判定。
4. 新增代码必须默认 strict:
   - 若出现例外，必须在注释中标明“迁移窗口、下线条件、责任人”。

## 实施步骤
1. 建立全量清单（engine 全目录）:
   - 扫描并登记所有 fallback/静默 return/旧格式兼容点，形成改造清单。
2. Resume Point 严格化（优先）:
   - 改造 `currentChoiceResumePoint` 与 `ParseResumePoint*`。
   - 移除无前缀兼容分支，非法值显式失败。
3. Choice/Interrupt/Action 路由严格化:
   - 清理 default 隐式成功路径，统一错误返回。
4. Skill Flow 严格化（engine 全覆盖）:
   - 对 `internal/engine/skill_flow_*.go` 和相关 runtime 文件逐批移除兜底执行分支。
5. 协议入口严格化:
   - 未知 `Cmd`、非法 `RoomAction`、未知 `ActionType` 返回结构化错误，不再静默。
6. 测试与门禁:
   - 补齐严格模式用例，新增“禁止旧格式兼容/禁止静默吞错”断言。

## 验收
1. 编译:
   - `go test ./... -run '^$'`
2. Engine 全量回归:
   - `go test ./internal/engine/... -count=1`
3. Model 回归:
   - `go test ./internal/model/... -count=1`
4. Server 协议回归:
   - `go test ./internal/server/... -count=1`
5. 外围回归:
   - `go test ./tests/... -count=1`
6. 严格模式专项断言:
   - 无前缀 resume point 输入被拒绝。
   - `currentChoiceResumePoint` 不再通过空值静默降级掩盖非法状态。
   - 未知 `Cmd` / 非法 JSON / 未知 `RoomAction.Action` 返回结构化错误。
7. 静态扫描门禁（建议纳入 CI）:
   - 对 `internal/engine/**/*.go`、`internal/model/**/*.go` 进行关键字扫描，新增兜底路径需失败并解释。
8. 若存在历史失败:
   - 输出中明确区分“历史失败”与“本次引入失败”并附日志证据。

## 输出物
1. Engine 全量改造清单与完成状态（文件级）。
2. 严格模式契约清单（新增失败语义与禁止行为）。
3. 验收报告（命令、结果、失败归因）。
4. 后续卡片建议（Phase 2: 兼容字段收口与删除）。
