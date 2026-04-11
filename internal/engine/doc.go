// Package engine 实现星杯对局的核心状态机与规则运行时。
//
// # 对局如何推进（Review 主线）
//
// 客户端/CLI 提交 model.PlayerAction 后，由 GameEngine.HandleAction 解析并写入状态，
// 随后必须调用 Drive() 让引擎在「无待处理中断」的前提下自动推进阶段，直到再次需要玩家输入或回合结束。
//
// Drive（game_drive.go）循环大致顺序：
//  1. 若存在 PendingInterrupt → 停住，等待 HandleAction 处理选择/确认；
//  2. 若无延迟伤害且存在 DeferredFollowups → 先处理延迟后续（保证如封印伤害等顺序）；
//  3. 行动收尾钩子与 action 汇总（action_finalize / action_summary）；
//  4. 非回合阶段：延迟伤害结算、弃牌选择、响应恢复、战斗交互（turn_fsm_dispatcher.driveNonTurnPhase）；
//  5. 回合 FSM：TurnBeforeStart → BeforeAction → TurnStart → ActionStart → ActionExecution →
//     ActionEnd → ExtraAction → TurnEnd（turn_fsm_dispatcher.driveTurnFSM）。
//
// # 战斗与伤害
//
// 攻击/法术进入 CombatStack，经宣告、命中判定、应战、承伤等阶段（combat.go、attack_lifecycle.go、
// pending_damage_runtime.go）。伤害多经 PendingDamageQueue 排队，以便插入响应技能与圣盾等逻辑。
//
// # 技能系统
//
// SkillDispatcher.OnTiming 按 model.FlowTiming 窗口收集被动/响应/场上牌逻辑，并 execute 或 PushInterrupt。
// 具体技能效果在 skill 子包的 Handler（CanUse/Execute）；按角色分流的中断与选项多在 skill_flow_*.go。
//
// # 子目录
//
//   - skill/（package skills）：各技能 LogicHandler 与场上状态解析（field_status_resolver）。
//   - hook/promptfmt：Prompt 文案格式化辅助。
//   - core/runtimeutil：与引擎弱耦合的小工具（数值/上下文读取）。
//
// 测试文件 *_test.go 为回归场景，不参与运行时加载。
package engine
