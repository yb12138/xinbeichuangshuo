# Codex 重构自动化工作流

本文给出一条可执行的流水线: `Spec 输入 -> Codex 执行 -> Harness 验收 -> 报告沉淀`。

## 1. 目录约定

- Spec 模板: `specs/refactor/spec.template.md`
- Codex Prompt 模板: `specs/refactor/codex_prompt_template.md`
- 工作流脚本: `scripts/refactor/*.sh`
- 运行产物: `artifacts/refactor_runs/<run-id>/`
- 严格模式分批拆卡示例: `docs/strict_mode_phase1_batch_cards.md`

## 2. 一次完整执行

### Step A: 准备 Spec

1. 通过命令创建:
   ```bash
   make refactor-spec SPEC=<your-task-name>
   ```
2. 或手动复制模板:
   ```bash
   cp specs/refactor/spec.template.md specs/refactor/<your-task>.md
   ```
3. 填写以下内容:
   - 背景/问题
   - 允许改动范围
   - 非目标
   - 验收命令

可参考示例:
- `specs/refactor/example.mainflow-skill-decouple.md`

### Step B: 创建运行目录

```bash
make refactor-new SPEC=specs/refactor/<your-task>.md
```

这会创建:
- `artifacts/refactor_runs/<run-id>/spec.md`
- `artifacts/refactor_runs/<run-id>/README.md`

### Step C: 校验 Spec

```bash
make refactor-validate SPEC=artifacts/refactor_runs/<run-id>/spec.md
```

若校验失败，先补齐缺失章节再继续。

### Step D: 生成 Codex 输入

```bash
make refactor-prompt SPEC=artifacts/refactor_runs/<run-id>/spec.md
```

生成文件:
- `artifacts/refactor_runs/<run-id>/codex_prompt.md`

将其内容粘贴给 Codex 执行。

### Step D2: 自动调起 Agent（可选，推荐）

如果你不想手工粘贴 prompt，可直接自动调起 CLI:

```bash
make agent-run RUN=artifacts/refactor_runs/<run-id> AGENT=codex
# 或
make agent-run RUN=artifacts/refactor_runs/<run-id> AGENT=claude
```

其中:
- `AGENT=codex` 使用 `codex exec --full-auto`
- `AGENT=claude` 使用 `claude -p --permission-mode acceptEdits`

执行日志会写入:
- `artifacts/refactor_runs/<run-id>/logs/agent.codex.log` 或 `agent.claude.log`
- 最终输出副本: `agent_last_message.md`

### Step E: 运行验收 Harness

快速验收:
```bash
make harness PROFILE=smoke RUN=artifacts/refactor_runs/<run-id>
```

全量验收:
```bash
make harness PROFILE=full RUN=artifacts/refactor_runs/<run-id>
```

可用 profile:
- `compile`
- `smoke`
- `engine`
- `server`
- `full`

### Step F: 生成报告

```bash
make report RUN=artifacts/refactor_runs/<run-id>
```

报告文件:
- `artifacts/refactor_runs/<run-id>/report.md`

## 2.5 一键自动化（从 Spec 到报告）

你可以用一条命令跑完整流程:

```bash
make refactor-auto SPEC=specs/refactor/<your-task>.md AGENT=codex PROFILE=smoke
```

该命令会自动执行:
1. 创建 run 目录
2. 校验 spec
3. 生成 prompt
4. 调起 agent 执行改造
5. 跑 harness 验收
6. 生成 report

## 3. 推荐团队规范

1. 每个 Spec 只处理一种坏味道（主流程耦合、状态泛滥、机制重复三选一）。
2. 每次改造都要带验收命令，不接受“看起来没问题”。
3. 每次运行目录都保留，作为回归资产。
4. 失败日志必须和 Spec 绑定，便于后续修复卡片接力。
