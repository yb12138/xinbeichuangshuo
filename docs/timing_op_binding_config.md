# Timing Op Binding Config

## 目标
- 将 `TimingOnAttackDeclared / TimingOnHitCheck / TimingOnDamageCalculated` 三大分发器的 `op -> handler` 映射外置为配置源。
- 保持主流程统一：分发器只负责按 `op` 路由，不写角色分支。
- 角色差异通过 `require_any_characters + priority` 在注册阶段做覆盖（overlay）。

## 配置文件
- 路径：`internal/engine/config/timing_op_bindings.json`
- 载入方式：`go:embed`（引擎启动时解析一次，角色变更时仅重建运行时表）。

## 字段说明
每条 binding 含以下字段：

1. `stage`
- 可选：`attack_declared` / `hit_check` / `damage_calculated` / `hit_check_skill`

2. `op`
- 对应阶段下的分发操作名（例如 `combat_counter_element`、`heal_resist`）。

3. `handler`
- 处理器 key，由引擎内 resolver 映射到具体函数。
- 命名建议：
  - 默认处理器：`default.<stage>.<op>`
  - 覆盖处理器：`overlay.<role_or_group>.<stage>.<op>`

4. `priority`
- 数值越大优先级越高；同一个 `stage+op` 在 presence 满足时选择优先级最高者。

5. `require_any_characters`（可选）
- 只要其中任一角色在场，此条 binding 即可参与竞争。

## 运行时流程
1. 引擎启动解析 JSON 为 binding registry（仅一次）。
2. 每次角色变化（加人/调试换角）重建运行时 `op-handler` 表。
3. 分发器在阶段内按 `op` 查表执行，查不到直接 panic（禁止静默兜底）。

## Overlay 示例
下面示例表示：当场上存在阴阳师时，覆盖 `hit_check.combat_counter_element` 的处理器。

```json
{
  "stage": "hit_check",
  "op": "combat_counter_element",
  "handler": "overlay.onmyoji.combat_counter_element",
  "priority": 100,
  "require_any_characters": ["onmyoji"]
}
```

## 新增一个 op 覆盖的步骤
1. 在 `timing_op_bindings.json` 添加一条 binding（设置 `stage/op/handler/priority/presence`）。
2. 在 `internal/engine/timing_op_registry.go` 的 resolver 映射里注册 `handler key -> handler func`。
3. 若是新 handler，补对应函数实现（可复用默认 handler，先保持行为等价）。
4. 运行测试：`go test ./internal/engine -count=1` 与 `go test ./...`。
