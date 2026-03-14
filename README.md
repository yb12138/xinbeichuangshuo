# StarCup Core Engine

基于《星杯传说》完整规则 (v5.0) 的纯后端游戏引擎。

## 功能特性

*   **纯后端逻辑**：Go 语言实现，无外部依赖。
*   **完整规则覆盖**：
    *   3v3 阵营对抗，6 人局支持。
    *   双胜利条件：士气归零 或 星杯 x5。
    *   完整的资源系统：星石、治疗剂、士气。
    *   战斗系统：攻击 -> 响应栈 -> 伤害结算（摸牌）。
    *   法术系统：支持完整法术牌库（150张牌）。
    *   角色技能系统：8个完整角色，包含被动、响应、主动技能。
    *   独有技系统：特定命格卡牌绑定独有技，只有对应角色才能使用。
*   **CLI 交互**：内置命令行工具，用于模拟对局和测试规则。

## 目录结构

```
starcup-core/
├── cmd/
│   └── cli/           # 命令行入口
├── internal/
│   ├── model/         # 数据模型 (Player, Card, GameState)
│   ├── rules/         # 基础规则 (Deck, Shuffle)
│   └── engine/        # 游戏引擎
│       ├── game.go    # 状态机与回合流程
│       ├── combat.go  # 战斗与响应系统
│       └── magic.go   # 法术系统
└── plan.md            # 开发大纲
```

## 快速开始

### 1. 启动 CLI

```bash
go run cmd/cli/main.go
```

默认启动 3v3 对局：
*   **红方 (Red)**:
    *   Player1 (狂战士)
    *   Player2 (魔法少女)
    *   Player3 (圣女)
*   **蓝方 (Blue)**:
    *   Player4 (神箭手)
    *   Player5 (暗杀者)
    *   Player6 (天使)

### 2. 常用命令

**游戏控制**
*   `start`: 开始一局新的 6 人游戏（随机先手）。
*   `status`: 查看当前全局状态（士气、星杯、当前玩家）。
*   `quit`: 退出。

**查询**
*   `hand`: 查看当前玩家手牌。
*   `hand <player_id>`: 查看指定玩家手牌 (e.g., `hand p2`).
*   `skills`: 显示当前玩家可用的主动技能。

**行动 (仅当前回合玩家)**
*   `atk <target> <card_index>`: 发起攻击 (e.g., `atk p2 1` 使用第1张牌攻击 p2)。
*   `magic <target> <card_index>`: 使用法术 (e.g., `magic p2 2`).
*   `buy`: 购买行动 (摸3牌，阵营资源池+1宝+1水)。
*   `syb`: 合成行动 (摸3牌，阵营资源池-3星石，+1星杯)。
*   `pass`: 结束回合。

**响应/交互**
*   `take`: 承受伤害。
*   `counter <card_index>`: 应战（需同系攻击牌）。
*   `confirm`: 确认操作（如确认技能发动）。
*   `skip`: 跳过/取消操作。
*   `choose <idx...>`: 选择/弃牌 (e.g., `choose 1 2` 选择第1和第2张牌)。

### 3. 3v3 对战示例

以下是一场模拟的战斗流程，展示了攻击、应战、技能响应和回合转换。

**场景**：游戏开始，轮到 Player1 (狂战士) 行动。

#### 1. 回合开始与查询
```bash
> start
[LOG] [Game] 游戏开始! 首发玩家: Player1 (Red)
[LOG] [Turn] 轮到 Player1 行动

> status
[STATE] --- Game Status (Phase: Action) ---
[STATE] Red Morale: 15 | Blue Morale: 15
[STATE] Current Turn: Player1 (Red)
[STATE] [*] Player1 (Red) [狂战士:血]: Hand 4/6, Gem 0, Cry 0, Heal 0 []
[STATE] [ ] Player2 (Red) [魔法少女:咏]: Hand 4/6, Gem 0, Cry 0, Heal 0 []
[STATE] [ ] Player3 (Red) [圣女:圣]: Hand 4/6, Gem 0, Cry 0, Heal 0 []
[STATE] [ ] Player4 (Blue) [神箭手:技]: Hand 4/6, Gem 0, Cry 0, Heal 0 []
[STATE] [ ] Player5 (Blue) [暗杀者:技]: Hand 4/6, Gem 0, Cry 0, Heal 0 []
[STATE] [ ] Player6 (Blue) [天使:圣]: Hand 4/6, Gem 0, Cry 0, Heal 0 []

> hand
[LOG] Player1's Hand:
[LOG] 1: [Attack] 火焰斩 (Fire) [Attack: 2]
[LOG] 2: [Attack] 雷光斩 (Thunder) [Attack: 2]
[LOG] 3: [Magic] 魔弹 (Magic)
[LOG] 4: [Attack] 地裂斩 (Earth) [Attack: 2]
```

#### 2. 发起攻击
Player1 决定使用第2张牌（雷光斩）攻击对方的 Player5 (暗杀者)。

```bash
> atk p5 2
[LOG] [Combat] Player1 对 Player5 发起了 雷光斩 (Thunder) 攻击！等待响应...
```

#### 3. 防守方响应
此时游戏进入响应阶段，轮到 Player5 操作。Player5 查看手牌，发现没有雷系牌可以应战，只能选择承受伤害。

```bash
> hand p5
[LOG] Player5's Hand:
[LOG] 1: [Attack] 火焰斩 (Fire) ...
[LOG] 2: [Magic] 圣盾 ...
...

> take
[LOG] [Combat] Player5 选择承受伤害
[LOG] [Combat] 攻击命中！红方阵营获得 1 宝石
[LOG] [Passive] Player1 的狂化发动！伤害 +1 (手牌: 3, 总伤害: 3)
[LOG] [Damage] Player5 受到 3 点伤害 (摸牌)
[LOG] [Passive] Player5 的反噬发动，攻击者 Player1 必须摸1张牌
[LOG] [Draw] Player1 摸了 1 张牌
```

#### 4. 技能互动 (水影)
Player5 在受到伤害摸牌前，触发了被动技能“水影”。系统自动提示是否使用。

```bash
>>> 是否发动 [水影] ?
>>> （除［特殊行动］外，当你摸牌前发动）弃X张水系牌（展示）。
选项:
  1: 确认发动 (ID: confirm)
  2: 跳过 (ID: skip)

> confirm
[LOG] Player5 确认发动 [水影]，请选择弃牌

>>> 请选择要弃置的卡牌发动水影技能
(请输入 'choose <index>' 选择手牌)

> choose 3
[LOG] [System] Player5 丢弃了 1 张牌！士气 -1
[LOG] [Draw] Player5 扣卡 3 张
[LOG] [System] Player5 手牌超出上限 1 张！需要选择 1 张牌丢弃
```

#### 5. 爆牌与回合结束
Player5 处理完手牌上限后，Player1 的攻击行动结束。Player1 没有其他额外行动，选择结束回合。

```bash
> choose 1
[LOG] [System] Player5 丢弃了 1 张牌！士气 -1

> pass
[LOG] [Turn] Player1 结束回合
[LOG] [Turn] 轮到 Player2 行动
```

此时轮到 Player2 (魔法少女) 行动，3v3 对战继续进行。
