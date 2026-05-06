# 项目角色与技能状态机映射文档

> 目标：按照指定的“核心对局状态机 + Flow timing + Effect Nodes”结构，从当前项目代码中抽取全部角色及技能并归档。  
> 数据来源：`internal/model/*.go`、`internal/data/characters.go`。  
> 抽取口径：以当前代码实现为准（35 角色 / 206 技能）。

---

## 4. 核心对局状态机 (Game Phases)

引擎的核心齿轮。技能发动的合法性校验及事件总线的派发均严格依赖此时间轴。

说明：你提供的是“业务态命名”，项目代码里已有对应的 `GamePhase` 枚举。本文同时给出对照。

### 4.1 全局初试阶段

* `GameInit`：游戏初试时（初始化牌库、发放初始手牌、分配角色指示物）

项目实现对照：

- 初始化入口：`engine.NewGameEngine()` + `StartGame()` 流程
- 初始牌库：`rules.InitDeck()`（150 张）
- 角色初始指示物：`engine.applyRoleDefaults()`

### 4.2 玩家主回合 (8 阶段)

| 阶段标识名 | 中文名 | 阶段核心职能描述 | 项目代码Phase对照 |
|:---|:---|:---|:---|
| `TurnBeforeStart` | 回合开始前 | 无实际意义，仅作为极其罕见的技能触发点。 | `PhaseBuffResolve`（含开场/回合前置结算语义） |
| `TurnStart` | 回合开始时 | 宣言回合开始，触发对应的被动技能。 | `TimingOnTurnStart`（`SkillDispatcher.OnTiming`） |
| `BeforeAction` | 行动阶段开始前 | **核心控制点**：强制结算面前的【中毒】伤害或【虚弱】跳过/摸牌判定。 | `PhaseBeforeAction` |
| `ActionStart` | 行动阶段开始时 | **核心控制点**：玩家发动【启动技】的唯一合法窗口。 | `PhaseStartup` |
| `ActionExecution` | 行动阶段中 | 玩家选择执行三大行动之一（攻击、法术、特殊）。 | `PhaseActionSelection` + `PhaseActionExecution` |
| `ActionEnd` | 行动结束时 | 行动结算完毕后的收尾时点。 | `TimingOnActionEnd` |
| `ExtraAction` | 行动结束后追加行动 | 若本回合存在额外行动，按照额外行动类型继续执行，也可以选择跳过。 | `PhaseExtraAction` + `ActionQueue/PendingActions` |
| `TurnEnd` | 回合结束时 | 宣言回合结束，清理所有临时状态，移交当前回合归属。 | `PhaseTurnEnd` |

### 4.3 战斗/法术结算 (6 阶段)

当玩家在行动阶段发起攻击或法术时，触发此微型状态机闭环：

| 阶段标识名 | 中文名 | 阶段核心职能描述 | 项目代码对照 |
| :--- | :--- | :--- | :--- |
| `CombatDeclare` | ① 发动阶段 | 攻击/法术宣告，触发“发动时”被动。 | `TimingOnAttackDeclared`；法术宣告另见 `TimingOnMagicDeclared` / `EventCardUsed` |
| `CombatHitCheck` | ② 命中判定阶段 | 拦截点：等待【圣盾】【圣光】抵挡，或【应战】及响应技的打断。 | `PhaseCombatInteraction` |
| `CombatCalcDamage` | ③ 计算伤害阶段 | 增减伤结算点：处理狂化、撕裂等数值变化被动。 | `TimingOnDamageCalculated` + `applyPassiveAttackEffects` |
| `CombatHeal` | ④ 治疗响应阶段 | 询问遭受伤害者是否消耗【治疗】抵挡伤害。 | `PendingDamage` 中治疗处理 |
| `CombatApply` | ⑤ 实际产生伤害阶段 | 真实伤害落地，为攻击方结算阵营星石收入。 | `applyDamage` + 阵营资源发放 |
| `CombatDraw` | ⑥ 实际承受伤害阶段 | 承受伤害方摸牌，并执行爆牌检测及扣除士气逻辑。 | `ResolveDamage` + `checkHandLimit` |

### 4.4 事件触发时机钩子 (`FlowTiming`)

注意：Phase 是系统当前的“状态”，而 `FlowTiming` 是系统状态发生改变或动作发生时向外广播的“瞬间事件”。技能在静态配置里用 `SkillDefinition.Timings []FlowTiming` 声明自己监听哪些窗口，由 `SkillDispatcher.OnTiming` 收集并执行。

| 标识名 | 触发场景描述 | 逻辑与复用说明 | 技能侧配置要点 |
|:---|:---|:---|:---|
| **【主动与主回合钩子】** | | | |
| `TimingActive` | 玩家主动发动 | 普/独/大招的默认点击触发。 | `SkillTypeAction` 默认兜底；显式写在 `Timings` 中 |
| `TimingStartup` | 玩家主动发动（启动技专有） | 仅在`ActionStart`阶段合法，占用启动名额。 | `SkillTypeStartup` + 常见 `Timings` 含 `TimingOnTurnStart` 等 |
| `TimingOnTurnStart` | 玩家的回合开始时触发 | 对应各种回合初的被动/状态转换。 | `Timings` 包含 `TimingOnTurnStart` |
| **【动作与结算劫持钩子】** | | | |
| `TimingBeforeActionExecute`| 系统尝试执行某种行动前 | **高复用拦截点**：系统传入 `ActionType`，用于劫持/替换默认购买、提炼等规则。 | `Timings` 包含本常量；与回合 `BeforeAction` 阶段配合 |
| `TimingOnActionEnd` | 某项行动彻底结算完毕时 | **高复用结算点**：系统传入 `ActionType`，涵盖攻击、法术、特殊行动结束。 | `Timings` 包含 `TimingOnActionEnd` |
| `TimingOnSkillExecuted` | 某一特定技能完整执行完毕时 | 系统传入 `SkillID`。如监听真言术执行完毕。 | 部分链路由 Handler 内逻辑承载；可写入 `Timings` |
| **【战斗时间轴钩子】** | | | |
| `TimingOnAttackDeclared`| ① 任意攻击宣告发动时 | | `Timings` 包含 `TimingOnAttackDeclared` |
| `TimingOnMagicDeclared` | ① 任意法术宣告发动时 | | `Timings` 包含 `TimingOnMagicDeclared`（或法术路径上等价窗口） |
| `TimingOnHitCheck` | ② 命中判定时 | 拦截点：发效应战、圣盾、圣光、仪式中断。 | `Timings` 包含 `TimingOnHitCheck`；命中/未命中分支看 `AttackInfo` |
| `TimingOnDamageCalculated`| ③ 伤害计算完毕时 | 增减伤结算点：撕裂、剑魂等数值修饰（未扣治疗）。 | `Timings` 包含 `TimingOnDamageCalculated` |
| `TimingOnDamageApplied` | ⑤ 实际产生伤害时 | 伤害已定、扣除治疗后，未摸牌前（如蝶舞者【毒粉】）。 | `PendingDamage.Stage=DamageProcessed`（语义层） |
| `TimingOnDamageTaken` | ⑥ 实际承受伤害，准备摸牌前 | 摸牌和爆牌判定的前置点。 | `Timings` 包含 `TimingOnDamageTaken` |
| **【卡牌与状态流转钩子】** | | | |
| `TimingBeforeCardDrawn` | 摸牌动作发生前 | **拦截点**：用于修改摸牌数或劫持为弃牌（如暗杀者水影）。 | `Timings` 包含 `TimingBeforeCardDrawn` |
| `TimingOnCardDrawn` | 摸牌动作完成后 | 结算点：用于触发摸牌后的伴生效果。 | `Timings` 包含 `TimingOnCardDrawn` |
| `TimingOnCardDiscarded` | 弃牌动作完成后 | 系统传入弃掉的卡牌数组，用于判断是否触发额外技能。 | 由 `InterruptChoice + choice_type` 流程承载（统一选择主通路） |
| `TimingOnHealOverflow` | 获得治疗且超出自身上限时 | 专用于处理溢出转化机制。 | 在 `Heal`/上限校验逻辑中处理（无统一枚举） |
| `TimingOnCardPlayedOrRevealed` | 打出或展示卡牌时 | 如封印、冰霜祷言等。 | `Timings` 包含 `TimingOnCardPlayedOrRevealed` |
| `TimingOnFieldMarkChanged`| 基础效果/场上盖牌发生改变时| **高复用点**：系统传入行为是`Placed`还是`Removed`。 | `Timings` 包含 `TimingOnFieldMarkChanged` |
| `TimingOnOrientationChanged`| 角色发生横置/转正状态切换时| 触发对姿态敏感的技能（如兽灵武士）。 | `Timings` 包含 `TimingOnOrientationChanged` |
| `TimingBeforeMoraleLoss` | 士气下降结算前 | 神之庇护等。 | `Timings` 包含 `TimingBeforeMoraleLoss` |

### 4.5 技能执行序列配置 (Effect Nodes Sequence)

用来定义一个技能具体“干了什么”以及“执行的先后顺序”。

你给出的结构如下：

```go
type SkillDefinition struct {
    // ... 前面的基础信息、Condition、Cost 等保持不变 ...

    // 【核心新增】技能的实际按序执行逻辑链！
    Effects []EffectNode
}

// 单个执行节点
type EffectNode struct {
    ActionType model.EffectActionType // 要执行的具体动作（如：造成伤害、加血、摸牌）
    Target     model.EffectTargetType // 这个动作作用于谁（如：自己、选中的敌人、全场）
    Value      int                    // 动作的数值（如：伤害值、摸牌数）

    // 可选：关联的具体实体（比如：如果要放置一个基础效果，放什么？）
    StatusRef  *model.StatusEffect
    TokenRef   *model.TokenType
}
```

当前项目实现说明：

1. 当前并未统一落为 `Effects []EffectNode` 执行器。
2. 目前采用“`SkillDefinition` 描述 + `LogicHandler` 执行”的策略。
3. 你给的 `EffectActionType` / `EffectTargetType` 在本文第 5 章里作为“建议映射”给出（便于后续演进到节点化执行引擎）。

---

## 5. 角色与技能映射抽取（共 35 角色 / 206 技能）

说明：以下映射来自 `internal/data/characters.go`，阶段/Timing/EffectNodes 为按你给定结构做的工程映射。

### 5.1 天使（`angel`）

| 技能 | 类型 | 主回合阶段映射 | 战斗阶段映射 | Timing钩子映射 | EffectNodes建议 | 规则摘要 |
|:---|:---|:---|:---|:---|:---|:---|
| 天使羁绊 | 被动技 | `BeforeAction/TurnEnd（依效果移除时点）` | `不固定/非战斗专属` | `TimingOnFieldMarkChanged` | `EffectHeal + EffectPlaceStatus` | （每当你移除一个基础效果或是使用［圣盾］时）目标角色+1［治疗］。 |
| 天使祝福 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDiscard` | （弃1张水系牌［展示］）指定目标玩家给你2张牌或指定2名角色各给你1张牌。 |
| 风之洁净 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDiscard` | （弃1张风系牌［展示］）移除场上任意1个基础效果。 |
| 天使之歌 | 启动技 | `ActionStart` | `不固定/非战斗专属` | `TimingStartup` | `需按 Handler 逻辑定制` | ［回合限定］［水晶］（在你的回合开始前发动）移除场上任意1个基础效果。 |
| 神之庇护 | 响应技 | `ActionExecution/CombatDraw（依事件来源）` | `CombatDraw（⑥）` | `TimingOnDamageTaken（士气变动前）` | `EffectDamage` | X个［水晶］为我方抵御X点因法术伤害而造成的士气下降。 |
| 天使之墙 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectPlaceStatus` | 此牌可以当作［圣盾］使用。 |

### 5.2 狂战士（`berserker`）

| 技能 | 类型 | 主回合阶段映射 | 战斗阶段映射 | Timing钩子映射 | EffectNodes建议 | 规则摘要 |
|:---|:---|:---|:---|:---|:---|:---|
| 狂化 | 被动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive（或引擎内部派发）` | `EffectAttackDamage / EffectDamage` | 你发动的所有攻击伤害额外+1。（攻击命中时②，若你的手牌>3）本次攻击伤害额外+1。 |
| 撕裂 | 响应技 | `ActionExecution（战斗中）` | `CombatHitCheck（②）` | `TimingOnHitCheck` | `EffectAttackDamage / EffectDamage` | ［宝石］攻击命中后发动②，本次攻击伤害额外+2。 |
| 血腥咆哮 | 响应技 | `ActionExecution（战斗中）` | `CombatDeclare（①）` | `TimingOnAttackDeclared` | `EffectHeal` | 作为主动攻击打出时发动，若攻击的目标拥有的［治疗］为2，则本次攻击强制命中。 |
| 血影狂刀 | 被动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive（或引擎内部派发）` | `EffectAttackDamage / EffectDamage` | 作为主动攻击打出时发动●若命中后②对手的手牌为2，本次攻击伤害额外+2。●若命中后②对手的手牌为3，本次攻击伤害额外+1。 |

### 5.3 封印师（`sealer`）

| 技能 | 类型 | 主回合阶段映射 | 战斗阶段映射 | Timing钩子映射 | EffectNodes建议 | 规则摘要 |
|:---|:---|:---|:---|:---|:---|:---|
| 法术激荡 | 响应技 | `ActionEnd` | `CombatHitCheck/CombatDraw（依技能插入点）` | `TimingOnActionEnd` | `model.AppendExtraAction` | （［法术行动］结束时发动）额外+1［攻击行动］。 |
| 封印破碎 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `需按 Handler 逻辑定制` | ［水晶］将场上任意一张基础效果牌收入自己手中。 |
| 五系束缚 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDrawCard + EffectPlaceStatus` | ［水晶］使用开局自带的五系束缚专属技能卡，将其放置于目标对手面前，该对手跳过其下个行动阶段。在其下个行动阶段开始前他可以选择摸（2+X）张牌来取消五系束缚的效果。X为场上封印的数量，X最高为2。不论效果是否发动，触发后移除此牌。 |
| 水之封印 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDamage + EffectPlaceStatus` | （将水之封印放置于目标对手面前）该对手获得（直到他从手中打出或展示出水系牌时强制触发）对他造成3点法术伤害③，触发后移除此牌。 |
| 火之封印 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDamage + EffectPlaceStatus` | （将火之封印放置于目标对手面前）该对手获得（直到他从手中打出或展示出火系牌时强制触发）对他造成3点法术伤害③，触发后移除此牌。 |
| 地之封印 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDamage + EffectPlaceStatus` | （将地之封印放置于目标对手面前）该对手获得（直到他从手中打出或展示出地系牌时强制触发）对他造成3点法术伤害③，触发后移除此牌。 |
| 风之封印 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDamage + EffectPlaceStatus` | （将风之封印放置于目标对手面前）该对手获得（直到他从手中打出或展示出风系牌时强制触发）对他造成3点法术伤害③，触发后移除此牌。 |
| 雷之封印 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDamage + EffectPlaceStatus` | （将雷之封印放置于目标对手面前）该对手获得（直到他从手中打出或展示出雷系牌时强制触发）对他造成3点法术伤害③，触发后移除此牌。 |

### 5.4 风之剑圣（`blade_master`）

| 技能 | 类型 | 主回合阶段映射 | 战斗阶段映射 | Timing钩子映射 | EffectNodes建议 | 规则摘要 |
|:---|:---|:---|:---|:---|:---|:---|
| 风怒追击 | 响应技 | `ActionEnd` | `CombatHitCheck/CombatDraw（依技能插入点）` | `TimingOnActionEnd` | `需按 Handler 逻辑定制` | ［回合限定］（［攻击行动］结束时发动）额外+1风系［攻击行动］。 |
| 圣剑 | 被动技 | `ActionExecution（战斗中）` | `CombatDeclare（①）` | `TimingOnAttackDeclared` | `EffectDrawCard + EffectDiscard` | 若你的主动攻击为你本次行动阶段的第三次［攻击行动］，则此攻击强制命中。本次［攻击行动］结束后，你摸X张牌，弃X张牌（X<4）。 |
| 剑影 | 响应技 | `ActionEnd` | `CombatHitCheck/CombatDraw（依技能插入点）` | `TimingOnActionEnd` | `model.AppendExtraAction` | ［回合限定］［蓝水晶］（［攻击行动］结束时发动）额外+1［攻击行动］。 |
| 疾风技 | 响应技 | `ActionExecution（战斗中）` | `CombatDeclare（①）` | `TimingOnAttackDeclared` | `model.AppendExtraAction` | （作为主动攻击打出时发动）额外+1［攻击行动］。 |
| 烈风技 | 响应技 | `ActionExecution（战斗中）` | `CombatDeclare（①）` | `TimingOnAttackDeclared` | `EffectPlaceStatus` | （攻击目标拥有圣盾时发动）无视对手圣盾的效果，且此攻击对手无法应战。 |

### 5.5 神箭手（`archer`）

| 技能 | 类型 | 主回合阶段映射 | 战斗阶段映射 | Timing钩子映射 | EffectNodes建议 | 规则摘要 |
|:---|:---|:---|:---|:---|:---|:---|
| 贯穿射击 | 响应技 | `ActionExecution（战斗中）` | `CombatHitCheck（②）` | `TimingOnHitCheck` | `EffectAttackDamage / EffectDamage + EffectDiscard` | （主动攻击未命中时发动②，弃1张法术牌［展示］）对你所攻击的目标造成2点法术伤害③。 |
| 闪电箭 | 被动技 | `ActionExecution（战斗中）` | `CombatDeclare（①）` | `TimingOnAttackDeclared` | `需按 Handler 逻辑定制` | 你的雷系攻击对手无法应战。 |
| 狙击 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `model.AppendExtraAction` | ［水晶］目标角色手牌补到5张［强制］，额外+1［攻击行动］。 |
| 精准射击 | 响应技 | `ActionExecution（战斗中）` | `CombatDeclare（①）` | `TimingOnAttackDeclared` | `EffectAttackDamage / EffectDamage` | 此攻击强制命中，但本次攻击伤害-1。 |
| 闪光陷阱 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDamage` | 对目标角色造成2点法术伤害③。 |

### 5.6 暗杀者（`assassin`）

| 技能 | 类型 | 主回合阶段映射 | 战斗阶段映射 | Timing钩子映射 | EffectNodes建议 | 规则摘要 |
|:---|:---|:---|:---|:---|:---|:---|
| 反噬 | 被动技 | `ActionExecution/CombatDraw（依事件来源）` | `CombatDraw（⑥）` | `TimingOnDamageTaken` | `EffectAttackDamage / EffectDamage + EffectDrawCard` | （承受攻击伤害时发动⑥）攻击你的对手摸1张牌［强制］。 |
| 水影 | 被动技 | `ActionExecution/CombatDraw（依事件来源）` | `CombatDraw（⑥前置）` | `TimingBeforeCardDrawn` | `EffectDrawCard + EffectDiscard` | 摸牌前可弃X张水系牌；潜行状态下可额外弃1张法术牌。 |
| 潜行 | 启动技 | `ActionStart` | `不固定/非战斗专属` | `TimingStartup` | `EffectAttackDamage / EffectDamage + EffectDrawCard` | ［宝石］你可选择摸1张牌，［横置］持续到你的下个行动阶段开始，你的手牌上限-1；你不能成为主动攻击的目标；你的主动攻击对方无法应战且伤害额外+X，X为你剩余的能量数。潜行的效果结束时角色［转正］。 |

### 5.7 圣女（`saintess`）

| 技能 | 类型 | 主回合阶段映射 | 战斗阶段映射 | Timing钩子映射 | EffectNodes建议 | 规则摘要 |
|:---|:---|:---|:---|:---|:---|:---|
| 冰霜祷言 | 被动技 | `ActionExecution（战斗中）` | `不固定/非战斗专属` | `TimingOnSkillExecuted/TimingOnMagicDeclared（依上下文）` | `EffectHeal` | （每当你打出或展示水系牌或圣光时发动）目标角色+1［治疗］。 |
| 治愈之光 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectHeal` | 指定最多3名角色各+1［治疗］。 |
| 治疗术 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectHeal` | 目标角色+2［治疗］。 |
| 圣疗 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectHeal + model.AppendExtraAction` | ［回合限定］［水晶］任意分配3点［治疗］给1~3名角色，额外+1［攻击行动］或［法术行动］。 |
| 怜悯 | 启动技 | `ActionStart` | `不固定/非战斗专属` | `TimingStartup` | `需按 Handler 逻辑定制` | ［持续］［宝石］［横置］，你的手牌上限恒定为7［恒定］，你+1［水晶］。 |

### 5.8 魔法少女（`magical_girl`）

| 技能 | 类型 | 主回合阶段映射 | 战斗阶段映射 | Timing钩子映射 | EffectNodes建议 | 规则摘要 |
|:---|:---|:---|:---|:---|:---|:---|
| 魔弹掌控 | 被动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive（或引擎内部派发）` | `需按 Handler 逻辑定制` | 你主动使用魔弹时可以选择逆向传递。 |
| 魔弹融合 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDiscard` | （弃1张火系或地系牌［展示］）视为发动1次【魔弹】。 |
| 魔爆冲击 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDamage + EffectDiscard` | （弃1张法术牌［展示］）我方战绩区+1颗［宝石］；选择2名目标角色各选择弃1张法术牌，否则该角色受到2点法术伤害；之后你可选择弃1张牌。 |
| 毁灭风暴 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDamage` | ［宝石］对任2名目标对手各造成2点法术伤害③。 |

### 5.9 女武神（`valkyrie`）

| 技能 | 类型 | 主回合阶段映射 | 战斗阶段映射 | Timing钩子映射 | EffectNodes建议 | 规则摘要 |
|:---|:---|:---|:---|:---|:---|:---|
| 神圣追击 | 响应技 | `ActionEnd` | `CombatHitCheck/CombatDraw（依技能插入点）` | `TimingOnActionEnd` | `EffectHeal + model.AppendExtraAction` | 攻击/法术行动结束时，若你有治疗，可移除1点治疗，额外+1攻击行动。 |
| 秩序之印 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectHeal + EffectDrawCard` | 摸2张牌，然后自身+1治疗和+1蓝水晶。 |
| 和平行者 | 被动技 | `ActionExecution（战斗中）` | `CombatDeclare（①）` | `TimingOnAttackDeclared` | `需按 Handler 逻辑定制` | 主动攻击时若处于英灵形态，自动脱离英灵形态。 |
| 军神威光 | 启动技 | `ActionStart` | `不固定/非战斗专属` | `TimingStartup` | `EffectHeal` | （回合开始时，若你处于［英灵形态］）选择以下1项发动：●你+1［治疗］，［转正］脱离［英灵型态］。●（移除我方战绩区X个星石，X<3）目标角色+X［治疗］。 |
| 英灵召唤 | 响应技 | `ActionExecution（战斗中）` | `CombatHitCheck（②）` | `TimingOnHitCheck` | `EffectAttackDamage / EffectDamage + EffectHeal + EffectDiscard` | ［水晶］（攻击命中时发动②）本次攻击伤害额外+1，（若你额外弃1张法术牌［展示］）目标角色+1［治疗］。 |

### 5.10 元素师（`elementalist`）

| 技能 | 类型 | 主回合阶段映射 | 战斗阶段映射 | Timing钩子映射 | EffectNodes建议 | 规则摘要 |
|:---|:---|:---|:---|:---|:---|:---|
| 元素吸收 | 被动技 | `ActionExecution/CombatDraw（依事件来源）` | `CombatDraw（⑥）` | `TimingOnDamageTaken` | `EffectDamage + EffectAddToken` | 你造成法术伤害后，元素+1（上限3）。元素点燃造成的伤害不触发。 |
| 元素点燃 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDamage + EffectAddToken + model.AppendExtraAction` | 移除3点元素，对任意角色造成2点法术伤害，并额外+1法术行动。 |
| 雷击 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDamage + EffectDiscard` | 独有技法术：对任意角色造成1点法术伤害；可额外弃1张雷系牌使伤害+1；阵营+1宝石。 |
| 冰冻 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDamage + EffectHeal + EffectDiscard` | 独有技法术：选择1名角色受1点法术伤害，再选择1名角色+1治疗；可额外弃1张水系牌使伤害+1。 |
| 风刃 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectAttackDamage / EffectDamage + EffectDiscard + model.AppendExtraAction` | 独有技法术：对任意角色造成1点法术伤害，额外+1攻击行动；可额外弃1张风系牌使伤害+1。 |
| 陨石 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDamage + EffectDiscard + model.AppendExtraAction` | 独有技法术：对任意角色造成1点法术伤害，额外+1法术行动；可额外弃1张地系牌使伤害+1。 |
| 火球 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDamage + EffectDiscard` | 独有技法术：对任意角色造成2点法术伤害；可额外弃1张火系牌使伤害+1。 |
| 月光 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDamage` | 手上有红宝石时可发动，对任意角色造成（X+1）点法术伤害，X为发动后剩余能量数。 |

### 5.11 仲裁者（`arbiter`）

| 技能 | 类型 | 主回合阶段映射 | 战斗阶段映射 | Timing钩子映射 | EffectNodes建议 | 规则摘要 |
|:---|:---|:---|:---|:---|:---|:---|
| 仲裁法则 | 被动技 | `TurnStart` | `不固定/非战斗专属` | `TimingOnTurnStart` | `需按 Handler 逻辑定制` | 游戏开始时获得2个蓝水晶。 |
| 审判浪潮 | 被动技 | `ActionExecution/CombatDraw（依事件来源）` | `CombatDraw（⑥）` | `TimingOnDamageTaken` | `EffectDamage + EffectAddToken` | 每次承受伤害时，审判+1（上限4）。 |
| 仲裁仪式 | 启动技 | `ActionStart` | `不固定/非战斗专属` | `TimingStartup` | `EffectAddToken` | 回合开始时可消耗1宝石进入审判形态：手牌上限恒定为5，并+1审判。 |
| 仪式中断 | 启动技 | `ActionStart` | `不固定/非战斗专属` | `TimingStartup` | `EffectAddToken` | 回合开始时若处于审判形态，可脱离形态并使我方战绩区+1红宝石。 |
| 末日审判 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDamage + EffectAddToken` | 移除当前所有审判点数，对任意角色造成等量法术伤害。 |
| 判决天平 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDiscard + EffectAddToken` | ［水晶］你+1［审判］，再选择以下一项发动：●弃掉你的所有手牌。●将你的手牌补到上限［强制］，我方战绩区+1［宝石］。 |

### 5.12 冒险家（`adventurer`）

| 技能 | 类型 | 主回合阶段映射 | 战斗阶段映射 | Timing钩子映射 | EffectNodes建议 | 规则摘要 |
|:---|:---|:---|:---|:---|:---|:---|
| 欺诈 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDiscard` | 主动技能：选择1名敌方角色，弃同系牌将本次视为一次主动攻击（弃2张同系可选非暗灭系；弃3张同系视为暗灭）。 |
| 强运 | 被动技 | `ActionExecution（战斗中）` | `CombatDeclare（①）` | `TimingOnAttackDeclared` | `需按 Handler 逻辑定制` | 发动欺诈后，+1蓝水晶。 |
| 地下法则 | 响应技 | `ActionEnd` | `CombatHitCheck/CombatDraw（依技能插入点）` | `TimingOnActionEnd` | `需按 Handler 逻辑定制` | 购买行动后触发，战绩区+2红宝石。 |
| 冒险者天堂 | 响应技 | `ActionEnd` | `CombatHitCheck/CombatDraw（依技能插入点）` | `TimingOnActionEnd` | `需按 Handler 逻辑定制` | 你执行提炼时，可将本次提炼出的［宝石］和［水晶］全部交给1名队友（不能拆分），然后若你有能量则移除你的1［能量］。 |
| 偷天换日 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `model.AppendExtraAction` | ［回合限定］［水晶］二选一：转移对方战绩区1个［宝石］到我方；或将我方战绩区所有［水晶］转为［宝石］。随后额外+1［攻击行动］或［法术行动］。 |

### 5.13 圣枪骑士（`holy_lancer`）

| 技能 | 类型 | 主回合阶段映射 | 战斗阶段映射 | Timing钩子映射 | EffectNodes建议 | 规则摘要 |
|:---|:---|:---|:---|:---|:---|:---|
| 神圣启示 | 被动技 | `TurnStart` | `不固定/非战斗专属` | `TimingOnTurnStart` | `EffectHeal` | 当我方星杯数不小于对方时，治疗上限+1。 |
| 辉耀 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectHeal + EffectDiscard + model.AppendExtraAction` | 弃1张水系牌，全场各+1治疗，同时额外+1攻击行动。 |
| 惩戒 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectHeal + EffectDiscard + model.AppendExtraAction` | 弃1张法术牌，将任意其他角色1点治疗转移给你，并额外+1攻击行动。 |
| 圣击 | 响应技 | `ActionExecution（战斗中）` | `CombatHitCheck（②）` | `TimingOnHitCheck` | `EffectHeal` | 攻击命中后，若本次未发动天枪/地枪，则+1治疗。 |
| 天枪 | 响应技 | `ActionExecution（战斗中）` | `CombatDeclare（①）` | `TimingOnAttackDeclared` | `EffectHeal` | 主动攻击前，若治疗≥2且本回合未发动圣光祈愈，可移除2治疗使本次攻击不可应战。 |
| 地枪 | 响应技 | `ActionExecution（战斗中）` | `CombatHitCheck（②）` | `TimingOnHitCheck` | `EffectAttackDamage / EffectDamage + EffectHeal` | （主动攻击命中后发动②）移除你的X点［治疗］，本次攻击伤害额外+X，X最高为4；不能和［圣击］同时发动。 |
| 圣光祈愈 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectHeal + model.AppendExtraAction` | 有宝石时可发动：+2治疗（不受上限但总治疗封顶5），并额外+1攻击行动。 |

### 5.14 精灵射手（`elf_archer`）

| 技能 | 类型 | 主回合阶段映射 | 战斗阶段映射 | Timing钩子映射 | EffectNodes建议 | 规则摘要 |
|:---|:---|:---|:---|:---|:---|:---|
| 元素射击 | 响应技 | `ActionExecution（战斗中）` | `CombatDeclare（①）` | `TimingOnAttackDeclared` | `EffectDiscard + EffectAddToken` | 每回合一次：主动攻击前（非暗系）可弃1法术牌或移除1祝福，按攻击系别附加元素箭效果。 |
| 动物伙伴 | 响应技 | `ActionEnd` | `CombatHitCheck/CombatDraw（依技能插入点）` | `TimingOnActionEnd` | `EffectDamage` | 你的回合内，当你造成的伤害结算后，可摸1弃1。 |
| 精灵密仪 | 启动技 | `ActionStart` | `不固定/非战斗专属` | `TimingStartup` | `需按 Handler 逻辑定制` | 启动技：消耗1红宝石，进入精灵祝福形态并将牌库顶3张作为祝福。 |
| 宠物强化 | 响应技 | `ActionExecution` | `CombatHitCheck/CombatDraw（依技能插入点）` | `TimingActive（或引擎内部派发）` | `需按 Handler 逻辑定制` | 触发动物伙伴时可消耗1蓝水晶，将效果改为任意角色摸1弃1。 |

### 5.15 瘟疫法师（`plague_mage`）

| 技能 | 类型 | 主回合阶段映射 | 战斗阶段映射 | Timing钩子映射 | EffectNodes建议 | 规则摘要 |
|:---|:---|:---|:---|:---|:---|:---|
| 不朽 | 响应技 | `ActionEnd` | `CombatHitCheck/CombatDraw（依技能插入点）` | `TimingOnActionEnd` | `EffectHeal` | 法术行动结束后自动+1治疗（圣光/魔弹传递不触发）。 |
| 圣渎 | 被动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive（或引擎内部派发）` | `EffectAttackDamage / EffectDamage + EffectHeal` | 你的治疗不能抵挡攻击伤害，但可抵挡法术伤害；治疗上限初始+3。 |
| 瘟疫 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDamage + EffectHeal + EffectDiscard` | 弃1张地系牌，对除自己外所有角色各造成1点法术伤害（逆序结算），自身+1治疗。 |
| 死亡之触 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDamage + EffectHeal + EffectDiscard` | 选择X点治疗与Y张同系牌（X,Y均≥2）并弃置，指定目标造成(X+Y-3)点法术伤害；本次不触发不朽。 |
| 剧毒新星 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDamage + EffectHeal` | 消耗1红宝石，对除自己外所有角色造成2点法术伤害（逆序结算），自身+1治疗。 |

### 5.16 魔剑士（`magic_swordsman`）

| 技能 | 类型 | 主回合阶段映射 | 战斗阶段映射 | Timing钩子映射 | EffectNodes建议 | 规则摘要 |
|:---|:---|:---|:---|:---|:---|:---|
| 修罗连斩 | 响应技 | `ActionEnd` | `CombatHitCheck/CombatDraw（依技能插入点）` | `TimingOnActionEnd` | `需按 Handler 逻辑定制` | 攻击行动结束后，若手牌存在火系攻击牌，可额外+1次火系攻击行动。 |
| 暗影凝聚 | 启动技 | `ActionStart` | `不固定/非战斗专属` | `TimingStartup` | `EffectDamage` | 启动阶段可对自己造成1点法术伤害并进入暗影形态，至下个己方回合开始转正。 |
| 暗影之力 | 被动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive（或引擎内部派发）` | `EffectAttackDamage / EffectDamage` | 暗影形态下，你发起的所有攻击伤害额外+1。 |
| 暗影抗拒 | 被动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive（或引擎内部派发）` | `需按 Handler 逻辑定制` | 行动阶段不能使用法术牌；非自己行动阶段可使用魔弹/圣光响应。 |
| 暗影流星 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDamage + EffectDiscard` | 暗影形态下弃2张法术牌，对任意角色造成2点法术伤害；可额外移除我方战绩区2星石转正并+1红宝石。 |
| 黄泉震颤 | 响应技 | `ActionExecution（战斗中）` | `CombatDeclare（①）` | `TimingOnAttackDeclared` | `EffectDiscard` | 每回合一次：主动攻击前若有红宝石可发动，本次攻击视为暗灭不可应战；命中后手牌补至上限并弃2。 |

### 5.17 血色剑灵（`crimson_sword_spirit`）

| 技能 | 类型 | 主回合阶段映射 | 战斗阶段映射 | Timing钩子映射 | EffectNodes建议 | 规则摘要 |
|:---|:---|:---|:---|:---|:---|:---|
| 血色荆棘 | 被动技 | `ActionExecution（战斗中）` | `CombatHitCheck（②）` | `TimingOnHitCheck` | `需按 Handler 逻辑定制` | 攻击命中时自动+1鲜血（上限3）。 |
| 赤色一闪 | 响应技 | `ActionEnd` | `CombatHitCheck/CombatDraw（依技能插入点）` | `TimingOnActionEnd` | `EffectAttackDamage / EffectDamage + model.AppendExtraAction` | 攻击行动结束后若有鲜血，可移除1鲜血并对自己造成2点法术伤害，额外+1攻击行动。 |
| 血染蔷薇 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDamage + EffectHeal` | 移除2鲜血，移除任意角色2点治疗，并将我方1蓝水晶翻为1红宝石；若庭院在场，额外对所有人各1点法术伤害。 |
| 血气屏障 | 响应技 | `ActionExecution/CombatDraw（依事件来源）` | `CombatDraw（⑥）` | `TimingOnDamageTaken` | `EffectDamage` | 受到法术伤害时可移除1鲜血使本次伤害-1，然后可对任意对手造成1点法术伤害。 |
| 血蔷薇庭院 | 被动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive（或引擎内部派发）` | `EffectDamage + EffectHeal + EffectPlaceStatus` | 专属卡在场时，所有人的治疗均不能用于抵挡伤害；你的回合结束时移回手牌区。 |
| 散华轮舞 | 启动技 | `ActionStart` | `不固定/非战斗专属` | `TimingStartup` | `EffectDiscard + EffectPlaceStatus` | 启动阶段二选一：1)耗蓝放置庭院并+2鲜血；2)耗红放置庭院并+2鲜血（上限可达4）且手牌弃至4。 |

### 5.18 祈祷师（`prayer_master`）

| 技能 | 类型 | 主回合阶段映射 | 战斗阶段映射 | Timing钩子映射 | EffectNodes建议 | 规则摘要 |
|:---|:---|:---|:---|:---|:---|:---|
| 祈祷 | 启动技 | `ActionStart` | `不固定/非战斗专属` | `TimingStartup` | `需按 Handler 逻辑定制` | 启动阶段消耗1红宝石进入祈祷形态（本局不退出）；祈祷形态下主动攻击会累计祈祷符文。 |
| 祈祷符文 | 被动技 | `ActionExecution（战斗中）` | `CombatDeclare（①）` | `TimingOnAttackDeclared` | `需按 Handler 逻辑定制` | 祈祷形态下，每次主动攻击时+2祈祷符文（上限3）。 |
| 光辉信仰 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectHeal + EffectDiscard` | 祈祷形态下可发动：移除1祈祷符文，弃2张牌，我方战绩区+1宝石（若未满），并使1名队友+1治疗。 |
| 黑暗诅咒 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDamage` | 祈祷形态下可发动：移除1祈祷符文，对任意1名角色造成2点法术伤害，再对自己造成2点法术伤害。 |
| 威力赐福 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectAttackDamage / EffectDamage + EffectPlaceStatus` | 将独有牌当法术牌打出并放置于1名队友面前；该队友攻击命中后可移除此牌，本次伤害+2。 |
| 迅捷赐福 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `model.AppendExtraAction + EffectPlaceStatus` | 将独有牌当法术牌打出并放置于1名队友面前；该队友攻击/法术行动结束后可移除此牌，额外+1攻击行动。 |
| 法力潮汐 | 响应技 | `ActionEnd` | `CombatHitCheck/CombatDraw（依技能插入点）` | `TimingOnActionEnd` | `model.AppendExtraAction` | 回合限定：法术行动结束后可消耗1蓝水晶，额外+1法术行动。 |

### 5.19 红莲骑士（`crimson_knight`）

| 技能 | 类型 | 主回合阶段映射 | 战斗阶段映射 | Timing钩子映射 | EffectNodes建议 | 规则摘要 |
|:---|:---|:---|:---|:---|:---|:---|
| 腥红圣约 | 响应技 | `ActionExecution（战斗中）` | `CombatDeclare（①）` | `TimingOnAttackDeclared` | `EffectHeal` | 回合限定：主动攻击时可响应，+1治疗。 |
| 腥红信仰 | 被动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive（或引擎内部派发）` | `EffectDamage + EffectHeal` | 你的治疗只能抵御自己对自己造成的伤害；治疗上限初始+2。 |
| 血腥祷言 | 启动技 | `ActionStart` | `不固定/非战斗专属` | `TimingStartup` | `EffectDamage + EffectHeal` | 当你有治疗时可发动：移除你的X点［治疗］，对自己造成X点法术伤害③；选择1~2名队友并任意分配这X点［治疗］，你+1［血印］。 |
| 杀戮盛宴 | 响应技 | `ActionExecution（战斗中）` | `CombatHitCheck（②）` | `TimingOnHitCheck` | `EffectAttackDamage / EffectDamage` | 主动攻击命中且有血印时可响应：移除1血印并对自己造成4法术伤害，本次攻击伤害+2。 |
| 热血沸腾 | 被动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive（或引擎内部派发）` | `EffectDamage + EffectHeal` | 因伤害导致我方士气下降时进入热血沸腾形态；该形态下伤害导致的士气下降被免疫。回合结束时脱离并+2治疗。 |
| 戒骄戒躁 | 响应技 | `ActionEnd` | `CombatHitCheck/CombatDraw（依技能插入点）` | `TimingOnActionEnd` | `需按 Handler 逻辑定制` | 热血沸腾形态下，攻击/法术行动结束后可消耗1蓝水晶（红宝石可替代），转正脱离并额外+1行动（攻击或法术二选一）。 |
| 腥红十字 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDamage + EffectDiscard` | 有蓝水晶与血印且手牌中至少2法术牌时可发动：消耗1蓝水晶与1血印，弃2法术牌，对自己4法术伤害并对任意角色3法术伤害。 |

### 5.20 英灵人形（`war_homunculus`）

| 技能 | 类型 | 主回合阶段映射 | 战斗阶段映射 | Timing钩子映射 | EffectNodes建议 | 规则摘要 |
|:---|:---|:---|:---|:---|:---|:---|
| 战纹掌控 | 被动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive（或引擎内部派发）` | `EffectAddToken` | 开局获得3战纹（当前实现为战纹/魔纹指示物）。 |
| 怒火压制 | 响应技 | `ActionExecution（战斗中）` | `CombatHitCheck（②）` | `TimingOnHitCheck` | `需按 Handler 逻辑定制` | 主动攻击未命中时可响应：翻转1战纹为魔纹。 |
| 战纹碎击 | 响应技 | `ActionExecution（战斗中）` | `CombatHitCheck（②）` | `TimingOnHitCheck` | `EffectAttackDamage / EffectDamage + EffectDiscard` | 主动攻击命中时可响应：翻转1战纹为魔纹，弃X张同系牌，本次攻击伤害额外+(X-1)；若处于蓄势迸发形态，可额外翻转Y个战纹，本次额外法术伤害+Y。 |
| 魔纹融合 | 响应技 | `ActionExecution（战斗中）` | `CombatHitCheck（②）` | `TimingOnHitCheck` | `EffectAttackDamage / EffectDamage + EffectDiscard` | 主动攻击未命中时可响应：翻转1魔纹为战纹，弃X张异系牌（X>1），对本次攻击目标造成(X-1)点法术伤害；若处于蓄势迸发形态，可额外翻转Y个魔纹，本次法术伤害额外+Y。 |
| 符文改造 | 启动技 | `ActionStart` | `不固定/非战斗专属` | `TimingStartup` | `EffectDrawCard` | 启动阶段可消耗1红宝石进入蓄势迸发形态，手牌上限+1并强制摸1张牌；可重新分配战纹/魔纹（总数保持3）；回合结束时转正脱离该形态。 |
| 双重回响 | 响应技 | `ActionExecution/CombatDraw（依事件来源）` | `CombatDraw（⑥）` | `TimingOnDamageTaken` | `EffectAttackDamage / EffectDamage` | 回合限定：造成攻击/法术伤害时可消耗1蓝水晶，对另一名角色造成等量（最多3）法术伤害。 |

### 5.21 神官（`priest`）

| 技能 | 类型 | 主回合阶段映射 | 战斗阶段映射 | Timing钩子映射 | EffectNodes建议 | 规则摘要 |
|:---|:---|:---|:---|:---|:---|:---|
| 神圣启示 | 被动技 | `ActionEnd` | `不固定/非战斗专属` | `TimingOnActionEnd` | `EffectHeal` | 特殊行动结束时触发，+1治疗。 |
| 神圣祈福 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectHeal + EffectDiscard` | 弃2张法术牌，自己+2治疗。 |
| 水之神力 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectHeal + EffectDiscard` | 先弃1张水系牌；若仍有手牌，再选1张交给1名队友；双方各+1治疗。 |
| 圣使守护 | 被动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive（或引擎内部派发）` | `EffectDamage + EffectHeal` | 治疗上限+4；每次抵御伤害时最多使用1点治疗。 |
| 神圣契约 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectHeal` | 消耗1蓝水晶，将自身治疗转移给1名队友（目标治疗上限按4封顶）。 |
| 神圣领域 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDamage + EffectHeal + EffectDiscard` | ［水晶］弃2张牌（手牌不足2时弃全部），再选择以下1项：①移除你1点治疗并对任意目标造成2点法术伤害；②你+2治疗并令1名队友+1治疗。 |

### 5.22 阴阳师（`onmyoji`）

| 技能 | 类型 | 主回合阶段映射 | 战斗阶段映射 | Timing钩子映射 | EffectNodes建议 | 规则摘要 |
|:---|:---|:---|:---|:---|:---|:---|
| 式神降临 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDiscard + model.AppendExtraAction` | ［持续］（弃2张命格相同的手牌［展示］）［横置］转为［式神形态］，你+1［鬼火］，额外+1［攻击行动］。 |
| 阴阳转换 | 响应技 | `ActionExecution（战斗中）` | `CombatDeclare（①）` | `TimingOnAttackDeclared` | `EffectAttackDamage / EffectDamage` | 你应战时可展示1张与来袭攻击同命格的攻击牌，视为你应战此次攻击并将其系别转为该牌系别；你+1［鬼火］。若处于式神形态则转正脱离，本次攻击伤害=X（X为你的鬼火数）。 |
| 式神转换 | 响应技 | `ActionExecution（战斗中）` | `CombatDeclare（①）` | `TimingOnAttackDeclared` | `EffectDrawCard` | 当阴阳转换生效且你处于式神形态时：你强制摸1张牌并+1［鬼火］，随后脱离式神形态。 |
| 黑暗祭礼 | 被动技 | `ActionEnd` | `不固定/非战斗专属` | `TimingOnActionEnd` | `EffectDamage` | 回合结束时若鬼火达上限，强制发动：选择1名角色，移除全部鬼火并对其造成2点法术伤害。 |
| 式神咒束 | 响应技 | `ActionExecution（战斗中）` | `CombatDeclare（①）` | `TimingOnAttackDeclared` | `需按 Handler 逻辑定制` | （目标队友受到主动攻击时①，若此攻击可应战且你处于［式神形态］，打出1张合理的应战攻击牌［展示］，移除我方［战绩区］1［宝石］1［水晶］）将本次攻击目标变更为你，且视为你使用此牌执行应战攻击。 |
| 生命结界 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDamage + EffectHeal + EffectDiscard` | 消耗1蓝水晶并+1鬼火后选择：①队友+1宝石+1治疗，自己受X点法伤（X为鬼火，X=3时该伤害不导致士气下降）；②若处于式神形态，弃2张同命格手牌并脱离式神形态，令1名队友弃1张手牌。 |

### 5.23 苍炎魔女（`blaze_witch`）

| 技能 | 类型 | 主回合阶段映射 | 战斗阶段映射 | Timing钩子映射 | EffectNodes建议 | 规则摘要 |
|:---|:---|:---|:---|:---|:---|:---|
| 永生银时计 | 被动技 | `ActionExecution/CombatDraw（依事件来源）` | `CombatDraw（⑥）` | `TimingOnDamageTaken（士气变动前）` | `EffectDamage + EffectAddToken` | ［重生］上限4；当你因承受法术伤害导致士气下降时，你+1［重生］。 |
| 苍炎法典 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDamage + EffectDiscard` | 弃1张火系牌［展示］，对目标角色和自己各造成2点法术伤害（目标先结算）。 |
| 天火断空 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDamage + EffectDiscard + EffectAddToken` | 弃2张火系牌并移除1点［重生］（烈焰形态下免移除），对目标角色和自己各造成3点法术伤害；若我方士气落后于目标方，本次伤害额外+1。 |
| 魔女之怒 | 启动技 | `ActionStart` | `不固定/非战斗专属` | `TimingStartup` | `EffectDrawCard + EffectAddToken` | 手牌<4时可发动：［横置］进入烈焰形态并选择摸0~2张牌；持续到下个行动阶段开始前。烈焰形态下：非水/暗攻击牌视为火系；发动天火断空无需消耗重生；手牌上限+(重生-2)。到时转正脱离。 |
| 替身玩偶 | 响应技 | `ActionExecution/CombatDraw（依事件来源）` | `CombatDraw（⑥）` | `TimingOnDamageTaken` | `EffectAttackDamage / EffectDamage + EffectDrawCard + EffectDiscard` | 任何人对你造成攻击伤害时可响应：弃1张法术牌［展示］，令1名队友摸1张牌。 |
| 痛苦链接 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDamage + EffectDiscard` | ［水晶］对目标对手和自己各造成1点法术伤害，然后你弃到3张手牌。 |
| 魔能反转 | 响应技 | `ActionExecution/CombatDraw（依事件来源）` | `CombatDraw（⑥）` | `TimingOnDamageTaken` | `EffectDamage + EffectDiscard` | ［水晶］任何人对你造成法术伤害时可响应：弃X张法术牌［展示］（X>1），对目标对手造成(X-1)点法术伤害。 |

### 5.24 贤者（`sage`）

| 技能 | 类型 | 主回合阶段映射 | 战斗阶段映射 | Timing钩子映射 | EffectNodes建议 | 规则摘要 |
|:---|:---|:---|:---|:---|:---|:---|
| 智慧法典 | 被动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive（或引擎内部派发）` | `EffectDamage + EffectDiscard` | 你的能量上限+1；你每次承受法术伤害时，若该伤害>3：你+2红宝石并可弃1张牌。 |
| 法术反弹 | 响应技 | `ActionExecution` | `CombatHitCheck/CombatDraw（依技能插入点）` | `TimingActive（或引擎内部派发）` | `EffectDamage + EffectDiscard` | 你每次承受法术伤害时，若该伤害仅为1点：可弃X张同系牌（X>1），对任意角色造成(X-1)点法术伤害，并对自己造成X点法术伤害。 |
| 魔道法典 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDamage + EffectDiscard` | ［宝石］弃X张异系牌（X>1），对目标角色与自己各造成(X-1)点法术伤害。 |
| 圣洁法典 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDamage + EffectHeal + EffectDiscard` | ［宝石］弃X张异系牌（X>2），最多(X-2)名角色各+2治疗，然后对自己造成(X-1)点法术伤害。 |

### 5.25 魔弓（`magic_bow`）

| 技能 | 类型 | 主回合阶段映射 | 战斗阶段映射 | Timing钩子映射 | EffectNodes建议 | 规则摘要 |
|:---|:---|:---|:---|:---|:---|:---|
| 魔贯冲击 | 响应技 | `ActionExecution（战斗中）` | `CombatDeclare（①）` | `TimingOnAttackDeclared` | `EffectAttackDamage / EffectDamage` | 主动攻击前可发动：移除1个火系充能，本次攻击伤害+1；若命中可再移除1个火系充能使伤害再+1；若未命中则对目标造成3点法术伤害。本回合与多重射击互斥。 |
| 雷光散射 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDamage` | 移除1个雷系充能：对所有对手各造成1点法术伤害；可额外移除X个雷系充能并指定1名对手，本次对其伤害额外+X。 |
| 多重射击 | 响应技 | `ActionEnd` | `CombatHitCheck/CombatDraw（依技能插入点）` | `TimingOnActionEnd` | `EffectAttackDamage / EffectDamage` | 攻击行动结束时可发动：移除1个风系充能，视为1次暗系主动攻击（不能攻击上次目标，且本次伤害-1）。本回合与魔贯冲击互斥。 |
| 充能 | 启动技 | `ActionStart` | `不固定/非战斗专属` | `TimingStartup` | `EffectDrawCard + EffectDiscard` | ［水晶］弃到4张牌后摸X张牌（X<5），可将最多X张手牌作为充能盖牌（上限8）。本回合不能发动魔贯冲击与雷光散射。 |
| 魔眼 | 启动技 | `ActionStart` | `不固定/非战斗专属` | `TimingStartup` | `EffectDrawCard + EffectDiscard` | ［宝石］选择：目标角色弃1张牌；或你摸3张牌。然后将1张手牌作为充能，并获得1蓝水晶。 |

### 5.26 魔枪（`magic_lancer`）

| 技能 | 类型 | 主回合阶段映射 | 战斗阶段映射 | Timing钩子映射 | EffectNodes建议 | 规则摘要 |
|:---|:---|:---|:---|:---|:---|:---|
| 暗之解放 | 启动技 | `ActionStart` | `不固定/非战斗专属` | `TimingStartup` | `EffectAttackDamage / EffectDamage` | ［横置］转为幻影形态，手牌上限恒定为5；本回合下一次主动攻击伤害+1，且本回合不能发动漆黑之枪与充盈。 |
| 幻影星尘 | 启动技 | `ActionStart` | `不固定/非战斗专属` | `TimingStartup` | `EffectDamage` | 仅幻影形态可发动：先对自己造成2点法术伤害并完全结算，随后转正脱离幻影形态；若未因此导致我方士气下降，则对目标角色造成2点法术伤害。 |
| 黑暗束缚 | 被动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive（或引擎内部派发）` | `需按 Handler 逻辑定制` | 你始终不能使用法术牌。 |
| 暗之障壁 | 响应技 | `ActionExecution/CombatDraw（依事件来源）` | `CombatDraw（⑥）` | `TimingOnDamageTaken` | `EffectDamage + EffectDiscard` | 任何人对你造成伤害时可发动：弃X张法术牌或X张雷系牌（同次发动不可混弃）。 |
| 充盈 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectAttackDamage / EffectDamage + EffectDiscard` | 弃1张法术牌或雷系牌：全体角色按逆时针各弃1张牌（我方可选择不弃）；除你外每有1名角色以此法弃置法术牌或雷系牌，你本回合下次主动攻击伤害+1；额外+1次攻击行动。 |
| 漆黑之枪 | 响应技 | `ActionExecution（战斗中）` | `CombatHitCheck（②）` | `TimingOnHitCheck` | `EffectAttackDamage / EffectDamage` | X［水晶］（仅幻影形态下，主动攻击手牌为1或2的对手并命中后）本次攻击伤害额外+（X+2）。 |

### 5.27 灵符师（`spirit_caster`）

| 技能 | 类型 | 主回合阶段映射 | 战斗阶段映射 | Timing钩子映射 | EffectNodes建议 | 规则摘要 |
|:---|:---|:---|:---|:---|:---|:---|
| 灵符-雷鸣 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDamage + EffectDiscard + EffectPlaceStatus` | 弃1张雷系牌［展示］，对任意2名角色各造成1点法术伤害。若触发封印，按“封印伤害→念咒→技能效果”顺序结算。 |
| 灵符-风行 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDamage + EffectDiscard + EffectPlaceStatus` | 弃1张风系牌［展示］，指定2名角色各弃1张牌。若触发封印，按“封印伤害→念咒→技能效果”顺序结算。 |
| 念咒 | 响应技 | `ActionExecution` | `CombatHitCheck/CombatDraw（依技能插入点）` | `TimingActive（或引擎内部派发）` | `EffectPlaceStatus` | 你每次发动灵符时，可将1张手牌面朝下放置在角色旁，作为［妖力］（上限2）。 |
| 百鬼夜行 | 响应技 | `ActionExecution（战斗中）` | `CombatHitCheck（②）` | `TimingOnHitCheck` | `EffectAttackDamage / EffectDamage` | 主动攻击命中后可发动：移除1个妖力。默认对1名角色造成1点法术伤害；若移除的是火系妖力，可展示并改为指定2名角色，对其余所有角色各造成1点法术伤害。 |
| 灵力崩解 | 响应技 | `ActionExecution` | `CombatHitCheck/CombatDraw（依技能插入点）` | `TimingActive（或引擎内部派发）` | `EffectDamage` | ［水晶］可与【灵符-雷鸣】或【百鬼夜行】同发动，使本次每段伤害额外+1。 |

### 5.28 吟游诗人（`bard`）

| 技能 | 类型 | 主回合阶段映射 | 战斗阶段映射 | Timing钩子映射 | EffectNodes建议 | 规则摘要 |
|:---|:---|:---|:---|:---|:---|:---|
| 沉沦协奏曲 | 响应技 | `ActionExecution` | `CombatHitCheck/CombatDraw（依技能插入点）` | `TimingActive（或引擎内部派发）` | `EffectDamage + EffectDiscard + EffectAddToken` | ［回合限定］仅普通形态：当本回合我方已对至少2名对手造成法术伤害并结算完后，可弃2张同系牌；你+1灵感。若弃牌中含法术牌，则再对1名对手造成1点法术伤害。 |
| 不谐和弦 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDrawCard + EffectDiscard + EffectAddToken` | 移除X点灵感（X>1）；若处于永恒囚徒形态则转正脱离。然后选择一项：你与目标角色各摸(X-1)张牌；或你与目标角色各弃(X-1)张牌。 |
| 禁忌诗篇 | 被动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive（或引擎内部派发）` | `EffectDamage + EffectAddToken` | 激昂狂想曲或胜利交响诗结算后：若灵感未满则你+1灵感并移除永恒乐章；若灵感已满则对自己造成3点法术伤害，且普通形态下转为永恒囚徒形态。 |
| 激昂狂想曲 | 响应技 | `ActionExecution` | `CombatHitCheck/CombatDraw（依技能插入点）` | `TimingActive（或引擎内部派发）` | `EffectDamage + EffectDiscard` | 回合开始时，若持有永恒乐章：选择一项——吟游诗人对2名目标对手各造成1点法术伤害；或你弃2张牌。 |
| 胜利交响诗 | 响应技 | `ActionExecution` | `CombatHitCheck/CombatDraw（依技能插入点）` | `TimingActive（或引擎内部派发）` | `EffectHeal` | 回合结束时，若持有永恒乐章：选择一项——将我方战绩区1个星石提炼为你的能量；或我方战绩区+1宝石且你+1治疗。 |
| 希望赋格曲 | 启动技 | `ActionStart` | `不固定/非战斗专属` | `TimingStartup` | `EffectHeal + EffectDrawCard + EffectDiscard + EffectAddToken + EffectPlaceStatus` | ［水晶］可先摸1张牌；然后选择：将永恒乐章放置于目标队友面前；或将永恒乐章转移给我方另一名目标角色，你弃1张牌并选择+1治疗或+1灵感。 |

### 5.29 勇者（`hero`）

| 技能 | 类型 | 主回合阶段映射 | 战斗阶段映射 | Timing钩子映射 | EffectNodes建议 | 规则摘要 |
|:---|:---|:---|:---|:---|:---|:---|
| 勇者之心 | 被动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive（或引擎内部派发）` | `需按 Handler 逻辑定制` | 游戏初始时，你+2［水晶］。 |
| 怒吼 | 响应技 | `ActionExecution（战斗中）` | `CombatDeclare（①）` | `TimingOnAttackDeclared` | `EffectAttackDamage / EffectDamage + EffectDrawCard + EffectAddToken` | 主动攻击前可发动：移除1点［怒气］，你摸0~1张牌，本次攻击伤害额外+2；若未命中，你+1［知性］。 |
| 禁断之力 | 响应技 | `ActionExecution（战斗中）` | `CombatHitCheck（②）` | `TimingOnHitCheck` | `EffectAttackDamage / EffectDamage + EffectDiscard + EffectAddToken` | ［水晶］主动攻击命中或未命中后可发动：展示并弃掉所有手牌；每有1张法术牌你+1怒气；未命中时每有1张水系牌你+1知性；命中时每有1张火系牌本次攻击伤害额外+1，并对自己造成等同火系牌数量的法术伤害。 |
| 精疲力竭 | 被动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive（或引擎内部派发）` | `EffectAttackDamage / EffectDamage + model.AppendExtraAction` | 发动禁断之力后强制触发：［横置］额外+1攻击行动；持续到你的下个行动阶段开始，手牌上限恒定为4。效果结束时转正并对自己造成3点法术伤害。 |
| 明镜止水 | 响应技 | `ActionExecution（战斗中）` | `CombatDeclare（①）` | `TimingOnAttackDeclared` | `EffectAddToken` | 主动攻击前可发动：移除4点［知性］，本次攻击对手无法应战；本次攻击结束时你+1［水晶］。 |
| 挑衅 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectAddToken + EffectPlaceStatus` | 移除1点［怒气］：将【挑衅】放置于目标对手面前，你+1［知性］；该对手在其下个行动阶段必须且只能主动攻击你，否则跳过该行动阶段。触发后移除此牌。 |
| 死斗 | 响应技 | `ActionExecution/CombatDraw（依事件来源）` | `CombatDraw（⑥）` | `TimingOnDamageTaken` | `EffectDamage + EffectAddToken` | ［宝石］每当你承受法术伤害时可发动：你+3［怒气］；若此伤害造成士气实际下降，本次士气下降值恒定为1。 |

### 5.30 格斗家（`fighter`）

| 技能 | 类型 | 主回合阶段映射 | 战斗阶段映射 | Timing钩子映射 | EffectNodes建议 | 规则摘要 |
|:---|:---|:---|:---|:---|:---|:---|
| 念气力场 | 被动技 | `ActionExecution/CombatDraw（依事件来源）` | `CombatDraw（⑥）` | `TimingOnDamageTaken` | `EffectDamage` | 所有对你造成的伤害每次最高为4点。 |
| 蓄力一击 | 响应技 | `ActionExecution（战斗中）` | `CombatDeclare（①）` | `TimingOnAttackDeclared` | `EffectAttackDamage / EffectDamage` | 主动攻击前可发动（斗气未满）：+1斗气，本次攻击伤害额外+1；若未命中，对自己造成X点法术伤害（X为当前斗气）。 |
| 念弹 | 响应技 | `ActionEnd` | `CombatHitCheck/CombatDraw（依技能插入点）` | `TimingOnActionEnd` | `EffectDamage + EffectHeal` | 法术行动结束时可发动（斗气未满）：+1斗气，对1名目标对手造成1点法术伤害；若其治疗为0，则你再承受X点法术伤害（X为当前斗气）。 |
| 百式幻龙拳 | 启动技 | `ActionStart` | `不固定/非战斗专属` | `TimingStartup` | `EffectAttackDamage / EffectDamage` | 持续：移除3斗气并横置。主动攻击伤害+2、应战攻击伤害+1；主动攻击需锁定同一目标，且不能发动蓄力一击。若改为执行法术行动或特殊行动，或主动攻击更换目标，则立即退出该状态。 |
| 气绝崩击 | 响应技 | `ActionExecution（战斗中）` | `CombatDeclare（①）` | `TimingOnAttackDeclared` | `EffectAttackDamage / EffectDamage` | 主动攻击前可发动：移除1斗气，本次攻击无法应战；然后对自己造成X点法术伤害（X为当前斗气）。不能与蓄力一击同时发动。 |
| 斗神天驱 | 启动技 | `ActionStart` | `不固定/非战斗专属` | `TimingStartup` | `EffectHeal + EffectDiscard` | ［水晶］你弃到3张牌，然后+2治疗。 |

### 5.31 圣弓（`holy_bow`）

| 技能 | 类型 | 主回合阶段映射 | 战斗阶段映射 | Timing钩子映射 | EffectNodes建议 | 规则摘要 |
|:---|:---|:---|:---|:---|:---|:---|
| 天之弓 | 被动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive（或引擎内部派发）` | `EffectAttackDamage / EffectDamage + EffectHeal + EffectAddToken` | 初始+1圣煌辉光炮、+2水晶、治疗上限+1；主动攻击若非圣命格伤害-1；主动攻击命中且为圣命格时+1信仰。 |
| 圣屑飓暴 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectHeal + EffectDiscard` | 弃2张同系攻击牌，视为一次圣命格同系主动攻击；若未命中，可移除最多2点治疗并令1名队友弃置等量手牌。 |
| 圣煌降临 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectHeal + EffectAddToken + model.AppendExtraAction` | 移除2点治疗或2点信仰，横置进入圣煌形态，并额外+1法术行动。 |
| 圣光爆裂 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectAttackDamage / EffectDamage + EffectHeal + EffectDiscard + EffectAddToken` | 仅圣煌形态可发动：①摸1，移除1治疗，+1信仰，1名我方+1治疗；②移除X治疗并弃X牌，至多选择X名对手各受(Y+2)点攻击伤害（Y为目标中有治疗者数量）。 |
| 流星圣弹 | 响应技 | `ActionExecution（战斗中）` | `CombatDeclare（①）` | `TimingOnAttackDeclared` | `EffectHeal + EffectAddToken` | 仅圣煌形态下，主动攻击前可发动：移除1点治疗或1点信仰，令1名我方角色+1治疗。 |
| 圣煌辉光炮 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectAddToken` | 仅圣煌形态可发动：移除1圣煌辉光炮与(4+士气落后值)信仰，所有角色手牌调整为4，我方星杯+1，然后选择将一方士气调整为与另一方相同。 |
| 自动填充 | 被动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive（或引擎内部派发）` | `EffectHeal + EffectAddToken` | 回合结束时，若你未执行特殊行动：可选择①消耗1水晶，+1信仰或+1治疗；②消耗1宝石，+1水晶并+2信仰或+2治疗。 |

### 5.32 灵魂术士（`soul_sorcerer`）

| 技能 | 类型 | 主回合阶段映射 | 战斗阶段映射 | Timing钩子映射 | EffectNodes建议 | 规则摘要 |
|:---|:---|:---|:---|:---|:---|:---|
| 灵魂吞噬 | 被动技 | `ActionExecution/CombatDraw（依事件来源）` | `CombatDraw（⑥）` | `TimingOnDamageTaken（士气变动前）` | `EffectAddToken` | （我方每有1点士气下降）你+1［黄色灵魂］（上限6）。 |
| 灵魂召还 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDiscard + EffectAddToken` | 弃X张法术牌［展示］，你+X点［蓝色灵魂］（上限6）。 |
| 灵魂转换 | 响应技 | `ActionExecution（战斗中）` | `CombatDeclare（①）` | `TimingOnAttackDeclared` | `EffectAddToken` | （你每发动1次主动攻击①）可转换1点你拥有的［灵魂］颜色。 |
| 灵魂镜像 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDrawCard + EffectDiscard + EffectAddToken` | （移除2点［黄色灵魂］）你弃2张牌，目标角色摸2张牌［强制］，但最多补到其手牌上限。 |
| 灵魂震爆 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDamage + EffectAddToken` | （移除3点［黄色灵魂］）对目标角色造成3点法术伤害；若其手牌<3且手牌上限>5，本次伤害额外+2。 |
| 灵魂赐予 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectAddToken` | （移除3点［蓝色灵魂］）目标角色+2［宝石］。 |
| 灵魂链接 | 启动技 | `ActionStart` | `不固定/非战斗专属` | `TimingStartup` | `EffectDamage + EffectAddToken + EffectPlaceStatus` | （仅你队友数>1时可发动，移除1黄魂+1蓝魂）将灵魂链接放置于目标队友面前；你或其承受伤害前可移除X蓝魂，将X点伤害转移给另一方（转移伤害为法术伤害）。 |
| 灵魂增幅 | 启动技 | `ActionStart` | `不固定/非战斗专属` | `TimingStartup` | `EffectAddToken` | ［宝石］你+2［黄色灵魂］和+2［蓝色灵魂］。 |

### 5.33 月之女神（`moon_goddess`）

| 技能 | 类型 | 主回合阶段映射 | 战斗阶段映射 | Timing钩子映射 | EffectNodes建议 | 规则摘要 |
|:---|:---|:---|:---|:---|:---|:---|
| 新月庇护 | 响应技 | `ActionExecution/CombatDraw（依事件来源）` | `CombatDraw（⑥）` | `TimingOnDamageTaken（士气变动前）` | `EffectDamage` | （我方角色因承受伤害导致爆牌并将士气下降时）转为暗月形态，并将本次爆牌改为你的暗月；本次士气不下降。 |
| 暗月诅咒 | 被动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive（或引擎内部派发）` | `需按 Handler 逻辑定制` | 你每次移除1个暗月，我方士气-1；当暗月数为0时，脱离暗月形态。 |
| 美杜莎之眼 | 响应技 | `ActionExecution` | `CombatHitCheck/CombatDraw（依技能插入点）` | `TimingActive（或引擎内部派发）` | `EffectAttackDamage / EffectDamage + EffectHeal + EffectDiscard + EffectAddToken` | 目标对手攻击时，可移除1个同系暗月：你+1治疗、+1石化；若移除的是法术牌，再弃1张牌并对目标对手造成1点法术伤害。 |
| 月之轮回 | 响应技 | `ActionExecution` | `CombatHitCheck/CombatDraw（依技能插入点）` | `TimingActive（或引擎内部派发）` | `EffectHeal` | 你的回合结束时，选择其一：①移除1暗月，令任意角色+1治疗；②移除1治疗，你+1新月。 |
| 月渎 | 响应技 | `ActionExecution` | `CombatHitCheck/CombatDraw（依技能插入点）` | `TimingActive（或引擎内部派发）` | `EffectDamage + EffectHeal` | ［回合限定］目标角色承受你造成的法术伤害后，可移除1治疗，对目标对手造成1点法术伤害。 |
| 暗月斩 | 响应技 | `ActionExecution（战斗中）` | `CombatHitCheck（②）` | `TimingOnHitCheck` | `EffectAttackDamage / EffectDamage` | 仅暗月形态下，主动攻击命中时可发动：移除X个暗月（X<=2），本次攻击伤害额外+X。 |
| 苍白之月 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectAttackDamage / EffectDamage + EffectDiscard + EffectAddToken + model.AppendExtraAction` | ［宝石］选择其一：①移除3石化，下次主动攻击不可应战、额外+1攻击行动，并额外获得一个回合。②移除X新月、你+1石化、弃1张牌，对目标对手造成(X+1)点法术伤害（X可为0）。 |

### 5.34 血之巫女（`blood_priestess`）

| 技能 | 类型 | 主回合阶段映射 | 战斗阶段映射 | Timing钩子映射 | EffectNodes建议 | 规则摘要 |
|:---|:---|:---|:---|:---|:---|:---|
| 血之哀伤 | 启动技 | `ActionStart` | `不固定/非战斗专属` | `TimingStartup` | `EffectDamage` | （对自己造成2点法术伤害）选择：转移【同生共死】目标，或移除【同生共死】。 |
| 流血 | 被动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive（或引擎内部派发）` | `EffectDamage + EffectHeal` | 普通形态下因承伤导致我方士气下降时，强制进入流血形态并+1治疗；流血形态下回合开始对自己造成1点法术伤害；当手牌<3时强制脱离流血形态。 |
| 逆流 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectHeal + EffectDiscard` | 仅流血形态下可发动：你弃2张牌，你+1治疗。 |
| 血之悲鸣 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDamage` | 仅流血形态下可发动：选择X（X<3），对目标角色和自己各造成（X+1）点法术伤害（先目标后自己）。 |
| 同生共死 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDrawCard + EffectPlaceStatus` | 你摸2张牌（强制），将【同生共死】放置于目标角色面前：普通形态下你和其手牌上限各-2；流血形态下你和其手牌上限各+1。 |
| 血之诅咒 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDamage + EffectDiscard` | ［宝石］对目标角色造成2点法术伤害，然后你弃3张牌（手牌不足则全弃）。 |

### 5.35 蝶舞者（`butterfly_dancer`）

| 技能 | 类型 | 主回合阶段映射 | 战斗阶段映射 | Timing钩子映射 | EffectNodes建议 | 规则摘要 |
|:---|:---|:---|:---|:---|:---|:---|
| 生命之火 | 被动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive（或引擎内部派发）` | `EffectAddToken` | 你的手牌上限-X（X为蛹数量），但最低为3。 |
| 舞动 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDrawCard + EffectDiscard + EffectAddToken + EffectPlaceStatus` | 选择：摸1张牌（强制）或弃1张牌（强制）；然后将牌库顶1张牌面朝下放置为茧。 |
| 毒粉 | 响应技 | `ActionExecution` | `CombatHitCheck/CombatDraw（依技能插入点）` | `TimingActive（或引擎内部派发）` | `EffectDamage + EffectAddToken` | 每当有角色产生1点实际法术伤害时，可移除1个茧，使该次伤害额外+1。 |
| 朝圣 | 响应技 | `ActionExecution` | `CombatHitCheck/CombatDraw（依技能插入点）` | `TimingActive（或引擎内部派发）` | `EffectDamage + EffectAddToken` | 每当你承受伤害时，可移除1个茧，抵御1点该来源伤害（每次伤害最多发动1次）。 |
| 镜花水月 | 响应技 | `ActionExecution` | `CombatHitCheck/CombatDraw（依技能插入点）` | `TimingActive（或引擎内部派发）` | `EffectDamage + EffectAddToken` | 每当有角色产生2点实际法术伤害时，可移除2张同系茧并展示：抵御该次伤害，改为你对原目标造成2次1点法术伤害。 |
| 凋零 | 响应技 | `ActionExecution` | `CombatHitCheck/CombatDraw（依技能插入点）` | `TimingActive（或引擎内部派发）` | `EffectDamage + EffectAddToken` | 你每次移除茧时，若该茧为法术牌，可展示并发动：对目标造成1点法术伤害，再对自己造成2点法术伤害；直到你下回合开始前，对方士气最低为1。 |
| 蛹化 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectAddToken + EffectPlaceStatus` | ［宝石］你+1蛹，并将牌库顶4张牌面朝下放置为茧。 |
| 倒逆之蝶 | 主动技 | `ActionExecution` | `不固定/非战斗专属` | `TimingActive` | `EffectDamage + EffectHeal + EffectDiscard + EffectAddToken` | ［水晶］你弃2张牌（不足则全弃），再二选一：①对目标造成1点不可用治疗抵御的法术伤害；②移除2个茧或对自己造成4点法术伤害，然后移除1个蛹。 |
