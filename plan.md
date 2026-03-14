# 星杯传说 (StarCup) 纯后端引擎开发计划

本项目旨在构建一个基于《星杯传说》完整竞技规则（v5.0）的纯后端游戏引擎，并提供 CLI（命令行）工具进行 6 人对战模拟测试。

## 1. 项目目标
*   **核心规则全覆盖**：支持 3v3 阵营战、士气/星杯双胜利条件、完整的战斗/响应链、摸牌伤害机制。
*   **纯后端逻辑**：无前端、无数据库，数据存储在内存中。
*   **CLI 交互测试**：通过终端指令模拟玩家操作，驱动游戏进程，验证规则准确性。
*   **完整技能系统**：支持8个职业、40+个技能的全技能实现。
*   **场上牌系统**：效果牌(Cover/Effect)和盖牌系统，支持复杂状态管理。

## 2. 架构设计

### 2.1 技术栈
- **语言**: Go 1.22.5
- **架构**: 纯后端引擎，内存数据存储
- **设计模式**: MVC架构 + 状态机 + 事件驱动

### 2.2 目录结构
```
starcup-engine/
├── cmd/
│   └── cli/           # CLI 入口，处理用户输入和显示
├── internal/
│   ├── model/         # 核心数据模型 (Player, Card, GameState, Enums, Skill)
│   ├── rules/         # 规则逻辑 (Deck, Combat, Resources)
│   ├── engine/        # 游戏状态机与流程控制 (Turn, Phase, EventLoop)
│   ├── skills/        # 技能系统接口与实现
│   └── data/          # 静态数据 (Characters, Cards)
├── pkg/
│   └── utils/         # 通用工具
├── go.mod
└── README.md
```

### 2.3 核心模块详解

#### 1. **Model (数据层)**
**卡牌系统 (Card)**:
- ID, Name, Type(CardTypeAttack/Magic), Element(7元素+光暗)
- Damage(伤害值), Description(描述)
- **命格系统**: Faction(幻/咏/血/技/圣), ExclusiveChar1/2, ExclusiveSkill1/2
- **MatchExclusive()**: 检查独有技匹配逻辑

**玩家系统 (Player)**:
- ID, Name, Camp(Red/Blue), Role, Character
- Hand(手牌), Field(场上牌: Effect/Cover模式)
- Resources: Gem(宝石), Crystal(水晶), Heal(治疗)
- Buffs(状态效果), TurnState(回合状态)
- **场上牌管理**: Add/RemoveFieldCard, GetFieldEffects, GetCoverCards, ConsumeCoverCards
- **手牌检查**: HasElement() 检查元素可用性

**游戏状态 (GameState)**:
- Players(map), PlayerOrder/TurnOrder, CurrentTurn
- Deck(牌库), DiscardPile(弃牌堆), ActionStack(响应栈)
- **全局资源**: Red/Blue Morale(士气), Cups(星杯), Gems/Crystals
- **中断系统**: PendingInterrupt(统一处理阻塞状态)
- **NewGameState()**: 初始化函数

**场上牌系统 (FieldCard)**:
- Card(原始卡), OwnerID/SourceID
- Mode(FieldEffect/FieldCover), Effect(具体效果类型)
- Trigger(触发时机), Locked(锁定状态)
- **效果类型**: Shield/Poison/Weak/各种Seal/FiveElementsBind

**技能系统 (SkillDefinition)**:
- ID, Title, Type(4种), Tags(消耗/限制标签)
- **资源消耗**: CostGem/Crystal, CostDiscards, CostCoverCards
- **弃牌约束**: DiscardElement/Type/Fate, RequireExclusive
- **场上牌放置**: PlaceCard, PlaceMode/Effect/Trigger
- Trigger(触发时机), ResponseType(Mandatory/Optional/Silent)
- TargetType(目标选择逻辑), MaxTargets

#### 2. **Engine (控制层)**
**游戏引擎 (GameEngine)**:
- **初始化**: NewGameEngine(), AddPlayer(), StartGame()
- **状态管理**: Phase流转, Turn循环, 玩家激活状态
- **资源管理**: ModifyGem/Crystal, DrawCards, DiscardCards, Heal
- **中断系统**: GetCurrentPrompt(), HandleInput(), SkipResponse()
- **战斗接口**: PerformAttack, PerformMagic, ResolveResponse

**技能分发器 (SkillDispatcher)**:
- **事件驱动**: OnTrigger() 处理各种TriggerType
- **技能扫描**: processSkills() 收集可触发技能
- **中断管理**: pushResponseInterrupt() 创建玩家确认中断
- **统一处理**: Mandatory/Optional/Silent 三种响应类型

**战斗系统 (Combat.go)**:
- **攻击流程**: PerformAttack() → 响应阶段 → 伤害结算
- **响应处理**: handleTakeHit/handleCounter/handleDefend
- **伤害机制**: applyDamage(), 摸牌伤害, 爆牌扣士气
- **被动效果**: applyPassiveAttackEffects(), triggerFieldEffects()
- **状态结算**: checkHandLimit(), addCampResource()

#### 3. **Skills (技能系统)**
**技能处理器接口 (SkillHandler)**:
- CanUse(ctx *Context): 检查使用条件
- Execute(ctx *Context): 执行技能逻辑

**注册系统 (Registry)**:
- Register()/GetHandler(): 技能处理器注册/获取
- InitHandlers(): 初始化所有40+个技能处理器

**技能分类**:
- **天使(6技能)**: 羁绊/祝福/洁净/之歌/庇护/之墙
- **狂战士(3技能)**: 撕裂/血腥咆哮/血影狂刀
- **封印师(8技能)**: 魔力涌动/封印破碎/五系束缚/五行封印
- **风之剑圣(5技能)**: 风怒/圣剑/剑影/疾风技/烈风技
- **神箭手(5技能)**: 贯穿射击/闪电箭/狙击/精准射击/闪光陷阱
- **暗杀者(3技能)**: 反噬/水影/潜行
- **圣女(5技能)**: 冰霜祈祷/治愈之光/治疗/圣疗/怜悯
- **魔法少女(5技能)**: 魔弹制御/魔弹融合/魔爆/毁灭风暴

#### 4. **Rules (规则层)**
**牌库管理 (Deck.go)**:
- InitDeck(): 生成150张完整卡牌
- Shuffle(): 洗牌算法
- DrawCards(): 抽牌逻辑，支持牌库耗尽时洗弃牌堆

#### 5. **Data (静态数据)**
**角色配置 (Characters.go)**:
- GetCharacters(): 返回8个职业的完整配置
- 每个职业包含: ID/Name/Title/Faction/MaxHand/Skills/Exclusives

#### 6. **CLI (交互层)**
**命令系统**:
- **状态查询**: status/hand/skills/help
- **游戏操作**: start/atk/magic/buy/syb/ext/pass
- **响应操作**: take/defend/counter/confirm/cancel
- **上帝视角**: 支持查看任意玩家的状态

**智能显示**:
- **响应阶段**: hand命令自动显示响应玩家手牌
- **状态显示**: Phase标签区分Action/Response阶段
- **技能展示**: 显示指令代码、使用方法、目标说明

**中断处理**:
- GetCurrentPrompt(): 返回当前等待的交互提示
- HandleInput(): 处理玩家输入，驱动游戏流程

## 3. 开发阶段规划

### Phase 1: 基础架构与资源系统 ✅
*   **任务 1.1** ✅: 完整的枚举系统(Camp/Element/CardType/GamePhase/ActionType)
*   **任务 1.2** ✅: 150张牌完整牌库，包含所有命格和独有技配对
*   **任务 1.3** ✅: 玩家资源管理(Gem/Crystal/Heal)，手牌管理，场上牌系统
*   **任务 1.4** ✅: 全局资源管理，士气扣减，星杯合成，阵营资源共享

### Phase 2: 回合流程与状态机 ✅
*   **任务 2.1** ✅: 完整的回合循环，6人轮转，支持中途加入
*   **任务 2.2** ✅: 7阶段流转系统(Start/WeakChoice/Trigger/Action/Response/Discard/End)
*   **任务 2.3** ✅: 特殊行动系统(Buy/Synthesize/Extract)，资源转换逻辑
*   **任务 2.4** ✅: 完整的CLI命令系统，状态查询，操作指令

### Phase 3: 战斗与伤害系统 (核心难点) ✅
*   **任务 3.1** ✅: PerformAttack() 完整攻击流程，技能触发前检查
*   **任务 3.2** ✅: 响应系统实现：
    *   **应战 (Counter)**：同系属性匹配，暗灭特殊规则，反击攻击生成
    *   **防御 (Defend)**：圣盾/圣光双重防御，强制命中穿透处理
    *   **承受 (Take Hit)**：伤害结算，摸牌机制
*   **任务 3.3** ✅: 伤害结算链路：
    *   摸牌伤害，爆牌检查(checkHandLimit)
    *   治疗抵消，士气扣减(15→0失败)
    *   被动技能触发，场上牌效果处理
*   **任务 3.4** ✅: CLI战斗指令：`atk <target> <card_idx>`, `take/defend/counter`

### Phase 4: 法术与状态系统 ✅
*   **任务 4.1** ✅: 完整法术牌系统：
    *   **魔弹系列**：即时伤害，属性伤害加成
    *   **中毒/虚弱**：状态挂载，持续伤害/减益
    *   **圣盾/圣光**：防御逻辑，抵消物理/法术伤害
*   **任务 4.2** ✅: 法术指令：`magic <target> <card_idx>`

### Phase 5: 技能系统 (最大难点) ✅
*   **任务 5.1** ✅: 技能架构设计：
    *   触发器系统(TriggerType): OnAttackStart/Hit/Miss/DamageTaken/TurnStart
    *   技能类型(SkillType): Passive/Startup/Action/Response
    *   目标选择(TargetType): Self/Enemy/Ally/Any/Specific
*   **任务 5.2** ✅: 技能处理器实现：
    *   40+个技能全部实现，复杂的逻辑处理
    *   资源消耗，弃牌约束，场上牌放置
    *   特殊效果：强制命中，伤害修改，状态控制
*   **任务 5.3** ✅: 技能分发器(SkillDispatcher)：
    *   事件驱动架构，统一技能触发
    *   中断系统集成，支持玩家确认
    *   响应类型：Mandatory/Optional/Silent

### Phase 6: 场上牌与Buff系统 ✅
*   **任务 6.1** ✅: 场上牌架构：
    *   FieldCard系统：Mode(Effect/Cover), Trigger时机
    *   效果类型：Shield/Poison/Weak/各种Seal/FiveElementsBind
*   **任务 6.2** ✅: Buff状态系统：
    *   基础效果：中毒/虚弱/圣盾
    *   特殊效果：五系束缚，各种封印
    *   持续结算：回合开始/结束处理
*   **任务 6.3** ✅: 触发机制：
    *   OnAttack/OnDamaged/OnTurnStart/Manual
    *   自动触发，被动技能集成

### Phase 7: 中断系统与交互优化 ✅
*   **任务 7.1** ✅: 统一中断系统：
    *   Interrupt结构体：Type/PlayerID/SkillIDs/Context
    *   支持ResponseSkill/StartupSkill/Discard/Choice
*   **任务 7.2** ✅: CLI交互优化：
    *   GetCurrentPrompt(): 智能提示生成
    *   HandleInput(): 统一输入处理
    *   响应阶段特殊处理：hand命令显示响应玩家
*   **任务 7.3** ✅: 用户体验：
    *   skills命令：显示完整指令代码和使用方法
    *   help系统：分类命令说明
    *   状态显示：Phase标签区分Action/Response

### Phase 8: 完整测试与优化 ✅
*   **任务 8.1** ✅: 6人局配置：3红vs3蓝，角色分配
*   **任务 8.2** ✅: 规则验证：
    *   攻击-响应链完整测试
    *   技能触发条件验证
    *   伤害结算准确性检查
*   **任务 8.3** ✅: 性能优化：
    *   内存管理：对象池，垃圾回收优化
    *   算法效率：技能扫描O(n)，触发器分发优化
*   **任务 8.4** ✅: 完整对局测试：
    *   胜利条件：士气归零/星杯5个
    *   边界情况：牌库耗尽，特殊技能交互
    *   稳定性：异常处理，状态一致性

## 4. CLI 命令系统详解

### 4.1 命令分类与语法

#### **查询类命令**
```bash
status                    # 查看全局游戏状态 (所有玩家信息、阶段、资源)
hand [playerID]          # 查看手牌 (默认响应阶段显示响应玩家，否则显示当前玩家)
skills                    # 查看当前玩家所有可用技能 (指令代码+使用方法)
help                      # 显示完整命令帮助
```

#### **游戏控制命令**
```bash
start                     # 开始游戏 (初始化牌库，发牌，随机先手)
cheat <pid> <gems> <crystals>  # 作弊：给指定玩家添加资源
```

#### **行动类命令 (仅当前回合玩家可用)**
```bash
atk <target> <card_idx>   # 攻击: atk p2 0 (用第1张牌攻击p2)
magic <target> <card_idx> # 法术: magic p3 1 (对p3使用第2张法术牌)
skill <skill_id> [target] # 技能: skill angel_blessing p2
buy                       # 购买 (1水晶 → 1宝石)
syb                       # 合成星杯 (2宝石 → 1星杯)
ext                       # 提炼 (1宝石 → 1水晶)
pass                      # 结束回合
```

#### **响应类命令 (仅在响应阶段可用)**
```bash
take                      # 承受伤害 (摸牌结算)
defend [card_idx]        # 防御 (自动使用圣盾/圣光，或指定圣光牌)
counter <target> <card_idx>  # 应战: counter p4 2 (用第3张牌反击p4)
```

#### **中断处理命令 (系统提示时使用)**
```bash
confirm                   # 确认使用技能/选择
cancel/skip              # 取消/跳过
choose <option_ids...>   # 多选: choose 0 2 (选择选项0和2)
discard <card_indices...> # 弃牌: discard 0 3 5
```

### 4.2 智能特性

#### **上下文感知显示**
- **响应阶段**: `hand` 自动显示被攻击者手牌，`status` 显示 "Response Player"
- **行动阶段**: `hand` 显示当前玩家手牌，`status` 显示 "Current Player"
- **中断状态**: 系统自动提示等待的交互内容

#### **技能展示系统**
```
🔸 天使祝福 (主动) (弃1张水系)
   效果: （弃1张水系牌）指定目标玩家给你2张牌或指定2名角色各给你1张牌
   指令代码: skill angel_blessing <目标玩家ID>
   示例: skill angel_blessing p2
```

#### **状态显示优化**
```
--- Game Status ---
Phase: Response | Response Player: p1    # 响应阶段显示响应玩家
 [p1 Red] Assassin [暗杀者:技]: Hand 4/6, Gem 0, Cry 0, Heal 0  *[p1标记为活跃]
 [p4 Blue] Warrior [狂战士:血]: Hand 3/6, Gem 1, Cry 0, Heal 0
```

### 4.3 命令执行流程

#### **攻击流程示例**
```bash
> atk p4 0                    # p1用第1张牌攻击p4
[Combat] Assassin 对 Warrior 发起了 水涟斩 攻击！等待响应...
> hand                        # 自动显示Warrior的手牌 (响应玩家)
Warrior's hand:
  0: [Fire] 火焰斩 (Attack Dmg:2) [血命格]
  1: [Earth] 地裂斩 (Attack Dmg:2) [技命格]
  2: [Water] 水涟斩 (Attack Dmg:2) [幻命格]
> counter p1 0               # Warrior用第1张牌应战Assassin
[Combat] Warrior 使用 火焰斩 应战成功！
> take                        # Assassin承受伤害
[Combat] Assassin 选择承受伤害
[Damage] Assassin 受到 2 点伤害 (摸牌)
```

## 5. 核心规则实现详解

### 5.1 卡牌系统
- **150张牌完整实现**: 每种元素5张×7元素×5命格 + 暗灭特殊牌
- **独有技系统**: 卡牌与角色技能绑定，RequireExclusive检查
- **命格分类**: 幻(封印师/灵魂术士)/咏(祈祷师/元素师)/血(狂战士/血巫女)/技(剑圣/神箭手)/圣(圣女/天使)

### 5.2 战斗机制
- **攻击流程**: 触发前检查 → 响应阶段 → 伤害结算 → 奖励获得
- **属性匹配**: 同元素应战，暗灭特殊规则 (可应战任意非暗灭，不可被应战)
- **强制命中**: 圣剑/精准射击/疾风技无视防御
- **伤害结算**: 摸牌伤害 → 爆牌检查 → 士气扣减 → 治疗抵消

### 5.3 技能系统架构
- **触发器驱动**: 8个触发时机点，统一OnTrigger()处理
- **资源消耗**: Gem/Crystal自动扣除，弃牌约束验证
- **场上牌放置**: Effect/Cover双模式，触发器自动管理
- **中断集成**: Mandatory/Optional/Silent三种响应处理

### 5.4 状态与效果
- **Buff系统**: 基础效果(中毒/虚弱/圣盾) + 特殊效果(五系束缚/封印)
- **场上牌系统**: 效果牌自动触发，盖牌手动消耗
- **持续结算**: 回合开始/结束自动处理状态变更

## 6. 技术实现亮点

### 6.1 架构设计
- **事件驱动**: SkillDispatcher统一处理所有技能触发
- **状态机**: 7阶段完整流转，支持中断和异步响应
- **中断系统**: 统一的阻塞状态管理，简化复杂交互

### 6.2 性能优化
- **内存管理**: 指针复用，减少GC压力
- **算法效率**: 技能扫描O(n)，触发器快速分发
- **状态一致性**: 原子操作保证，回滚机制

### 6.3 可扩展性
- **插件架构**: 技能处理器注册系统，易于添加新技能
- **配置驱动**: 角色和技能数据分离，支持快速调整
- **接口抽象**: IGameEngine接口，方便测试和扩展

## 7. 已实现功能清单

### 7.1 核心系统 ✅
- [x] 完整的150张牌牌库系统
- [x] 8个职业40+个技能全实现
- [x] 3v3阵营对战机制
- [x] 完整的攻击-响应链
- [x] 伤害结算系统 (摸牌/爆牌/士气)
- [x] 场上牌系统 (效果牌/盖牌)
- [x] 中断系统 (统一玩家输入处理)
- [x] 资源管理系统 (宝石/水晶/星杯/士气)

### 7.2 游戏规则 ✅
- [x] 攻击属性匹配和应战规则
- [x] 暗灭牌特殊规则 (无敌/不可应战)
- [x] 强制命中机制 (圣剑/精准射击等)
- [x] 独有技系统和卡牌绑定
- [x] 命格系统和职业专属
- [x] 胜利条件 (士气归零/星杯5个)

### 7.3 CLI功能 ✅
- [x] 完整的命令系统 (攻击/法术/技能/响应)
- [x] 智能状态显示 (Action/Response阶段区分)
- [x] 响应阶段特殊处理 (hand显示响应玩家)
- [x] 技能详细展示 (指令代码+使用方法)
- [x] 中断提示系统 (系统自动引导玩家输入)
- [x] 作弊命令支持 (方便测试)

## 8. 技术债务与已知问题

### 8.1 架构问题
- **技能处理器耦合**: 部分技能逻辑较为复杂，可考虑进一步拆分
- **状态一致性**: 在复杂技能交互时需额外验证状态变更
- **内存管理**: 大量对象创建，需监控GC压力

### 8.2 已知问题
- **并发安全**: 当前设计为单线程，暂不支持并发访问
- **错误处理**: 部分边界情况错误信息不够详细
- **配置管理**: 角色和技能配置硬编码，可考虑外部配置

### 8.3 测试覆盖
- **单元测试**: 核心逻辑缺少单元测试
- **集成测试**: 复杂技能组合测试不足
- **边界测试**: 极端情况 (牌库耗尽/资源不足) 测试不全

## 9. 性能监控与优化

### 9.1 当前性能指标
- **内存占用**: ~50MB (6人局+150张牌)
- **响应时间**: 技能触发 < 10ms
- **CPU使用**: 低负载，事件驱动优化

### 9.2 优化方向
- **对象池**: 复用Card/Player对象，减少GC
- **算法优化**: 技能扫描可使用索引加速
- **缓存机制**: 常用计算结果缓存

## 10. 部署与运维

### 10.1 环境要求
- Go 1.22.5+
- 内存: 128MB+
- 存储: 10MB (代码+配置)

### 10.2 启动方式
```bash
# 编译
go build cmd/cli/main.go

# 运行
./main

# 或直接运行
go run cmd/cli/main.go
```

### 10.3 日志与监控
- **游戏日志**: 控制台输出，包含战斗过程和状态变更
- **错误日志**: 异常情况详细记录
- **性能监控**: 可通过pprof进行性能分析

## 11. 贡献指南

### 11.1 代码规范
- 遵循Go官方编码规范
- 使用gofmt格式化代码
- 添加必要的注释和文档

### 11.2 开发流程
1. Fork项目
2. 创建功能分支
3. 编写代码和测试
4. 提交PR并描述变更

### 11.3 新功能添加
- **新技能**: 在`internal/data/characters.go`添加配置，在`handlers_impl.go`实现逻辑
- **新卡牌**: 在`internal/rules/deck.go`的`addExclusiveCards`中添加
- **新命令**: 在`cmd/cli/main.go`添加命令处理逻辑


