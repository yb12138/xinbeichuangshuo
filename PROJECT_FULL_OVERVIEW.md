# 星杯传说项目全景梳理（逻辑 / 内容 / 角色 / 技能）

> 统计时间：2026-03-19  
> 工作区：`/Users/yb/xinbeichuangshuo`

## 1. 项目定位与目标

本项目是《星杯传说》规则的工程化实现，当前由三层构成：

1. 核心引擎层（Go）
2. 联机服务层（WebSocket + 房间管理）
3. 可视化前端层（Vue3 + Pinia + Vite）

核心目标是支持多人对局（当前主流程为 6 人 3v3），并且以“规则准确 + 技能可扩展 + 回归可验证”为主线持续迭代。

## 2. 仓库结构总览

### 2.1 顶层目录

- `cmd/cli`：CLI 交互入口
- `cmd/server`：WebSocket/HTTP 服务入口
- `internal/model`：枚举、状态模型、协议模型
- `internal/rules`：基础牌库与洗牌等规则基础设施
- `internal/data`：角色配置与规则文档（规则/问答/行动顺序）
- `internal/engine`：状态机、攻击/法术结算、技能调度、中断系统
- `internal/engine/skills`：技能处理器实现与注册表
- `internal/server`：房间、客户端、重连、机器人自动操作
- `tests`：高层整体验证与技能专项测试
- `web`：前端代码与静态资源

### 2.2 技术栈

- 后端：Go 1.22.5（`module: starcup-engine`）
- 网络：`github.com/gorilla/websocket`
- 前端：Vue 3 + TypeScript + Pinia + Vite

## 3. 核心逻辑梳理（引擎视角）

## 3.1 状态机与回合阶段

核心阶段定义于 `internal/model/enums.go`，主要包括：

1. `PhaseBuffResolve`
2. `PhaseStartup`
3. `PhaseActionSelection`
4. `PhaseBeforeAction`
5. `PhaseActionExecution`
6. `PhaseCombatInteraction`
7. `PhaseDamageResolution`
8. `PhasePendingDamageResolution`
9. `PhaseExtraAction`
10. `PhaseTurnEnd`

兼容保留：`PhaseResponse`、`PhaseDiscardSelection`、`PhaseEnd`。

## 3.2 战斗链路（攻击）

核心由 `internal/engine/combat.go` + `internal/engine/game.go` 实现：

1. 发起攻击，构造 `CombatRequest` 压入 `CombatStack`
2. 进入战斗交互阶段（应战 / 防御 / 承受）
3. 计算伤害并应用被动修正
4. 分发 `TriggerOnDamageTaken` 以允许减伤/转移/响应
5. 处理中断（`PendingInterrupt`）
6. 执行最终伤害（摸牌伤害、爆牌、士气变化）
7. 触发攻击后事件（`TriggerOnPhaseEnd`）
8. 回到额外行动阶段，消费追加行动

## 3.3 法术链路

核心由 `internal/engine/magic.go` 实现，特征：

- 行动阶段和额外行动条件校验
- 法术牌类型校验 + 目标校验
- 支持法术转化/重定向链（如魔弹相关）
- 场上效果放置（中毒/虚弱/圣盾）
- 法术结束后分发阶段结束触发器

## 3.4 技能调度体系

核心由 `internal/engine/skill_dispatcher.go` + `internal/engine/skills/registry.go`：

- 事件驱动：按 Trigger 收集候选技能
- 身份驱动：区分 `Attacker` / `Defender` / `Any`
- 响应类型：`Mandatory` / `Optional` / `Silent`
- 执行前校验：费用、回合限制、独有卡匹配、CanUse
- 可处理角色技能与场上效果两种来源

## 3.5 中断与交互系统

`internal/model/types.go` 中定义统一中断模型 `Interrupt`，引擎通过 `PendingInterrupt` 阻塞推进，等待玩家输入后继续。

常见中断类型：

- `ResponseSkill`
- `StartupSkill`
- `Discard`
- `Choice`
- `MagicMissile`
- `MagicBulletFusion`
- `MagicBulletDirection`
- `HolySwordDraw`
- `SaintHeal`
- `MagicBlast`

## 3.6 资源与胜利条件

根据规则文档与状态模型，当前实现围绕：

- 阵营资源：宝石/水晶（战绩区）
- 阵营进度：星杯
- 阵营生命线：士气
- 角色资源：治疗、个人宝石、个人水晶、专属指示物

胜利条件：

1. 对方士气降至 0 或以下
2. 我方星杯达到 5

## 4. 内容资产梳理

## 4.1 规则/文档资产

- `internal/data/rule.md`：规则正文
- `internal/data/action.md`：阶段/时序清单
- `internal/data/qa.md`：FAQ 解释
- `star_cup_rules.md`：补充规则文档
- `plan.md`：开发阶段计划
- `role_skill_chain`：角色技能逻辑草案（早期）
- `role_skill_chain_followup_roles.md`：后续角色技能结构化梳理
- `xingbeichuangshuo.txt`：原始内容资料

## 4.2 卡牌内容（当前代码统计）

通过 `go run count_cards.go` 统计：

- 总牌数：150
- 攻击牌：111
- 法术牌：39

核心基础牌包含：火焰斩/水涟斩/地裂斩/风神斩/雷光斩/暗灭/中毒/虚弱/魔弹/圣盾/圣光。

## 4.3 角色立绘与前端素材

- `web/public/characters`：35 张角色立绘（与当前角色池一致）
- `web/public/assets/ui`：行动按钮、弹窗按钮、面板纹理等 UI 资产
- `web/public/image`：基础牌图像资源

## 5. 角色与技能全量清单（当前实现）

> 统计来源：`internal/data/characters.go`  
> 当前角色数：35  
> 角色技能定义总数：206  
> 技能注册表 `Register(...)` 数量：211（含基础效果处理器/跟进处理器等非角色直出项）

| # | 角色ID | 角色名 | 技能数 | 技能列表 |
|---|---|---|---:|---|
| 1 | `angel` | 天使 | 6 | 天使羁绊、天使祝福、风之洁净、天使之歌、神之庇护、天使之墙 |
| 2 | `berserker` | 狂战士 | 4 | 狂化、撕裂、血腥咆哮、血影狂刀 |
| 3 | `sealer` | 封印师 | 8 | 法术激荡、封印破碎、五系束缚、水之封印、火之封印、地之封印、风之封印、雷之封印 |
| 4 | `blade_master` | 风之剑圣 | 5 | 风怒追击、圣剑、剑影、疾风技、烈风技 |
| 5 | `archer` | 神箭手 | 5 | 贯穿射击、闪电箭、狙击、精准射击、闪光陷阱 |
| 6 | `assassin` | 暗杀者 | 3 | 反噬、水影、潜行 |
| 7 | `saintess` | 圣女 | 5 | 冰霜祷言、治愈之光、治疗术、圣疗、怜悯 |
| 8 | `magical_girl` | 魔法少女 | 4 | 魔弹掌控、魔弹融合、魔爆冲击、毁灭风暴 |
| 9 | `valkyrie` | 女武神 | 5 | 神圣追击、秩序之印、和平行者、军神威光、英灵召唤 |
| 10 | `elementalist` | 元素师 | 8 | 元素吸收、元素点燃、雷击、冰冻、风刃、陨石、火球、月光 |
| 11 | `arbiter` | 仲裁者 | 6 | 仲裁法则、审判浪潮、仲裁仪式、仪式中断、末日审判、判决天平 |
| 12 | `adventurer` | 冒险家 | 5 | 欺诈、强运、地下法则、冒险者天堂、偷天换日 |
| 13 | `holy_lancer` | 圣枪骑士 | 7 | 神圣启示、辉耀、惩戒、圣击、天枪、地枪、圣光祈愈 |
| 14 | `elf_archer` | 精灵射手 | 4 | 元素射击、动物伙伴、精灵密仪、宠物强化 |
| 15 | `plague_mage` | 瘟疫法师 | 5 | 不朽、圣渎、瘟疫、死亡之触、剧毒新星 |
| 16 | `magic_swordsman` | 魔剑士 | 6 | 修罗连斩、暗影凝聚、暗影之力、暗影抗拒、暗影流星、黄泉震颤 |
| 17 | `crimson_sword_spirit` | 血色剑灵 | 6 | 血色荆棘、赤色一闪、血染蔷薇、血气屏障、血蔷薇庭院、散华轮舞 |
| 18 | `prayer_master` | 祈祷师 | 7 | 祈祷、祈祷符文、光辉信仰、黑暗诅咒、威力赐福、迅捷赐福、法力潮汐 |
| 19 | `crimson_knight` | 红莲骑士 | 7 | 腥红圣约、腥红信仰、血腥祷言、杀戮盛宴、热血沸腾、戒骄戒躁、腥红十字 |
| 20 | `war_homunculus` | 英灵人形 | 6 | 战纹掌控、怒火压制、战纹碎击、魔纹融合、符文改造、双重回响 |
| 21 | `priest` | 神官 | 6 | 神圣启示、神圣祈福、水之神力、圣使守护、神圣契约、神圣领域 |
| 22 | `onmyoji` | 阴阳师 | 6 | 式神降临、阴阳转换、式神转换、黑暗祭礼、式神咒束、生命结界 |
| 23 | `blaze_witch` | 苍炎魔女 | 7 | 永生银时计、苍炎法典、天火断空、魔女之怒、替身玩偶、痛苦链接、魔能反转 |
| 24 | `sage` | 贤者 | 4 | 智慧法典、法术反弹、魔道法典、圣洁法典 |
| 25 | `magic_bow` | 魔弓 | 5 | 魔贯冲击、雷光散射、多重射击、充能、魔眼 |
| 26 | `magic_lancer` | 魔枪 | 6 | 暗之解放、幻影星尘、黑暗束缚、暗之障壁、充盈、漆黑之枪 |
| 27 | `spirit_caster` | 灵符师 | 5 | 灵符-雷鸣、灵符-风行、念咒、百鬼夜行、灵力崩解 |
| 28 | `bard` | 吟游诗人 | 6 | 沉沦协奏曲、不谐和弦、禁忌诗篇、激昂狂想曲、胜利交响诗、希望赋格曲 |
| 29 | `hero` | 勇者 | 7 | 勇者之心、怒吼、禁断之力、精疲力竭、明镜止水、挑衅、死斗 |
| 30 | `fighter` | 格斗家 | 6 | 念气力场、蓄力一击、念弹、百式幻龙拳、气绝崩击、斗神天驱 |
| 31 | `holy_bow` | 圣弓 | 7 | 天之弓、圣屑飓暴、圣煌降临、圣光爆裂、流星圣弹、圣煌辉光炮、自动填充 |
| 32 | `soul_sorcerer` | 灵魂术士 | 8 | 灵魂吞噬、灵魂召还、灵魂转换、灵魂镜像、灵魂震爆、灵魂赐予、灵魂链接、灵魂增幅 |
| 33 | `moon_goddess` | 月之女神 | 7 | 新月庇护、暗月诅咒、美杜莎之眼、月之轮回、月渎、暗月斩、苍白之月 |
| 34 | `blood_priestess` | 血之巫女 | 6 | 血之哀伤、流血、逆流、血之悲鸣、同生共死、血之诅咒 |
| 35 | `butterfly_dancer` | 蝶舞者 | 8 | 生命之火、舞动、毒粉、朝圣、镜花水月、凋零、蛹化、倒逆之蝶 |

## 6. 协议、交互与运行方式

## 6.1 CLI 侧

入口：`cmd/cli/main.go`

常用指令：

- `start` / `quit` / `pass`
- `atk <target> <idx>`
- `magic <target> <idx>`
- `skill <skill_id> [targets...] [discard_indices...]`
- `take` / `defend [idx]` / `counter [idx] [target]`
- `confirm` / `cancel|skip` / `choose`
- `buy` / `syb` / `ext`
- `cheat`

## 6.2 服务端与联机

入口：`cmd/server/main.go`

- WebSocket：`/ws`
- REST：
  - `/api/room/create`
  - `/api/room/info?room=XXXX`

房间侧能力（`internal/server`）：

- 6人房间管理
- 座次与阵营/角色分配
- 断线重连（`player_id` + `reconnect_token`）
- 机器人托管与自动行动
- 广播状态、提示和日志事件

## 6.3 消息协议

`internal/model/protocol.go` 定义了统一上下行结构：

- 上行：`PlayerAction`（Attack/Magic/Skill/Respond/...）
- 下行：`GameEvent`（state_update/prompt/error/game_end/...）
- WS 包装：`action` / `event` / `room` / `chat`

## 7. 前端（web）内容梳理

## 7.1 组件结构（11 个核心组件）

- `RoomLobby.vue`
- `GameBoard.vue`
- `BattleZone.vue`
- `PlayerArea.vue`
- `CardComponent.vue`
- `ActionPanel.vue`
- `PromptDialog.vue`
- `ActionTimeline.vue`
- `SkillDetailModal.vue`
- `VfxLayer.vue`
- `HelloWorld.vue`（模板残留）

## 7.2 状态管理与交互能力

`web/src/stores/gameStore.ts` 持有：

- 房间态（房间码、角色、重连token）
- 游戏态（士气/星杯/战绩资源/阶段/玩家）
- 动效态（飞牌、战斗提示、摸牌演出、伤害特效）
- 引导态（Prompt、目标选择、技能选择、弃牌选择）
- 战报态（battle feed、士气变化记录、终局快照）

`web/src/composables/useWebSocket.ts`：

- 动态 WS 地址拼接
- 本地重连信息缓存
- 自动重连重试
- 房间事件与游戏事件解析分发

## 7.3 前后端角色一致性

- 前端 `roleNameMap.ts` 包含 35 个角色映射
- 后端 `availableRoles` 与 `characters.go` 角色池一致
- `web/public/characters` 角色立绘数量为 35，与角色池一致

## 8. 测试与质量现状

## 8.1 测试规模（文件数）

- 测试文件总数：70
- `internal/engine`：48
- `internal/server`：5
- `internal/rules`：1
- `tests`：16

其中 `internal/engine/*_regression_test.go` 回归类测试：43 个。

## 8.2 测试类型分层

1. 角色技能专项（例如 `tests/*_skills_test.go`）
2. 全局流程（`full_game_3v3_all_roles_test.go`、`full_game_regression_campaign_test.go`）
3. 提示与交互（`ui_regression_prompts_test.go`）
4. 引擎回归（大量 `*_regression_test.go`，覆盖新角色与复杂联动）
5. 服务端行为（重连、座次、角色合法性、bot 自动行动）

## 9. 代码中“内容层”的关键对象

从模型层看，当前内容系统主要由以下对象承载：

- `Character` / `SkillDefinition`：角色与技能静态配置
- `Card`：基础牌（含命格、独有技归属）
- `FieldCard`：场上牌（效果/盖牌）
- `Buff`：状态效果
- `Tokens`：角色专有指示物（审判、灵魂、形态计数等）

这套对象使“规则文本”能落到“可执行状态机 + 可测试动作序列”。

## 10. 当前项目的“逻辑 + 内容 + 角色 + 技能”结论

1. 逻辑层：已经形成完整回合状态机 + 技能触发调度 + 中断交互闭环。
2. 内容层：基础牌库 150 张、角色池 35 个、角色技能定义 206 条，且具备大量补充文档。
3. 角色技能层：`characters.go` 为事实来源，`role_skill_chain*` 文档可作为策划/补全参考。
4. 工程层：CLI、Server、Web 三端闭环可跑；测试体系以回归为主，覆盖面偏实战导向。
5. 可维护性层：角色和技能继续扩展时，建议继续沿用“数据定义 + handler + regression test”三件套。

---

## 附：建议你后续持续维护这份文档的方式

每次新增角色或技能时，建议同步维护以下四处：

1. `internal/data/characters.go`（配置）
2. `internal/engine/skills/registry.go` + 对应 handler 文件（逻辑）
3. 对应 `*_regression_test.go`（回归）
4. 本文档第 5 节角色技能表（全景索引）

