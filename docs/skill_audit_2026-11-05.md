# 角色技能审计报告（2026-11-05）

## 说明

- 范围：代码里已实现的 37 个角色（`internal/engine/player/`），以 `docs/character.md` 的文案为准。文案里另外 64 个角色还没实现，不在范围内。
- 方法：8 个子 agent 分组读代码，覆盖后端角色包、通用流程和前端发动界面。各角色的 `go test` 全部通过，多数是缓存结果。少数结论用 `/tmp` 下的临时测试验证过，测试已删除。没跑 E2E，也没在浏览器里实测。
- 只读检查，没有改任何代码。
- 路径约定：角色章节标题里写了角色目录。条目里没带目录的文件名，都在 `internal/engine/player/<角色目录>/` 下；带目录的是从仓库根算起的路径。

置信度：

- 已核实：我自己复读代码确认过
- 已确认：子 agent 读代码或用临时测试确认，我没有复核
- 推测：需要人工确认

评审：每条末尾有一行「评审」，填写结论即可，比如"按文案改 / 保持现状 / 待定"。文案歧义的条目没有对错，只需要选一种理解。

## 建议修复顺序

1. 先修共用环节（第一部分）：通用-1、通用-2、通用-3、通用-5。修一次能覆盖一批角色。
2. 再修卡死和数值错误的高危条目：前端-1、魔弓-1、祈祷师-1、冒险家-1、魔枪-1、圣弓-1、英灵人形-1、兽灵武士-1、阴阳师-1、阴阳师-2、吟游诗人-1、吟游诗人-4、封印师-1。
3. 文案不符和文案歧义的条目，等你评审定下口径后再改。有几处现有测试把可疑行为固定成了预期（见第五部分），改实现时要一起改测试。

---

## 一、跨角色的共用问题

### 通用-1 回合限定技能在执行前就计数，取消后次数照样被扣
- Bug｜中（影响面大）｜已核实
- 证据：`internal/engine/runtime/skill/executor.go:37-39`
- 现象：带 `TagTurnLimit` 的技能在 `handler.Execute` 之前就 `UsedSkillCounts[skill.ID]++`。很多技能的 Execute 只是推出一个选择提示，玩家在提示里取消后，次数不会退回，本回合不能再用。
- 已知受影响：精灵射手·元素射击（`elf_archer/choices.go:84-102`）、英灵人形·双重回响（`war_homunculus/choices.go:209-214`）。其他在 Execute 里推提示的回合限定技能大概率也一样，没有逐个排查。
- 评审：

### 通用-2 先扣费再选择：取消不退费，或第一步取消报"当前步骤无法返回"
- Bug / 前端体验｜中｜已确认（英灵人形部分已核实）
- 证据：
  - `internal/model/prompt_flow.go:171-176, 263-266`：History 为空时 Back 失败
  - `internal/engine/skill_use.go:100-105`：宝石和回合计数在 Execute 之前扣除
  - `internal/engine/prompt_flow_cancel.go:39-41`：Abort 只 Pop 中断，不退费
- 现象：流程用 `MustBeginAt` 从中间某一步起步，这一步的取消策略是 Back，但前面没有上一步可退，取消就报错。另一类是取消策略为 Abort，只弹出中断，已经扣掉的宝石、水晶不退。
- 已知受影响（详见各角色条目）：
  - 贤者·魔道法典、圣洁法典
  - 苍炎魔女·魔能反转
  - 剑帝·剑气斩
  - 英灵人形·魔纹融合、战纹碎击
  - 冒险家·偷天换日
  - 月之女神·苍白之月
  - 阴阳师·生命结界、式神咒束
  - 血色剑灵·血染蔷薇
  - 灵魂术士·灵魂召还
  - 魔枪·漆黑之枪、充盈
  - 血之巫女·血之哀伤
  - 女武神·军威神光
  - 圣女·圣疗
  - 格斗家·百式幻龙拳
- 评审：

### 通用-3 回合结束钩子在每段额外行动之前都执行一次
- Bug｜中｜位置已核实，现象已确认（临时测试）
- 证据：`internal/engine/turn_fsm_dispatcher.go:609`，`runTimingTurnEndPreExtraHooks` 排在 PendingActions 出队之前
- 现象：每段额外行动结束后都会重新进入 TurnEnd，这个钩子就会再跑一次。
  - 兽灵武士：居合形态下身上挂着额外攻击时，一回合扣两次兽魂。临时测试里兽魂 2 → 1 → 0，还被强制脱离了形态。
  - 红莲骑士：额外行动开始前就脱离热血沸腾并 +2 治疗。额外行动里再次进入形态的话，同一回合会再 +2。
  - 阴阳师：式神降临给的额外攻击还没打，黑暗祭礼就先结算了。
- 文案写的是"回合结束"，按理应该在所有额外行动之后。
- 评审：

### 通用-4 回合限定的重置时点：应在每个玩家的回合开始时，而不是角色自己的回合
- Bug｜中｜已核实
- 证据：`internal/engine/turn_progression_runtime.go:72`（`nextPlayer.TurnState = model.NewPlayerTurnState()`）
- 现象：凡是"每回合一次"、又可能在别人回合触发的标记，都会一直留到角色自己的下一个回合。按裁定，应该在**每个玩家的回合开始时**重置。已知受影响：
  - 吟游诗人·沉沦协奏曲（见吟游诗人-4）
  - 英灵人形·双重回响（见英灵人形-5）
  - 月之女神·月渎（见月之女神-1）
- 评审：已裁定：回合限定按"任意一个回合"算，也就是每个玩家的回合各限一次。现在把标记记在 TurnState 里、只在角色自己的回合开始时重置，语义不对，要改成在每个玩家的回合开始时重置。受影响的有月渎、沉沦协奏曲、双重回响等。

### 通用-5 法术伤害类型的判断写法不统一
- Bug｜高（对蝶舞者）｜已核实
- 证据：
  - `internal/engine/core/runtimeutil/helpers.go:27-29`：`IsMagicDamageType` 的口径是"不是 Attack 就算法术"
  - `butterfly_dancer/timing_hooks.go:54`、`butterfly_dancer/runtime.go:126`：用 `!= model.MagicDamage` 严格比较
  - `archer/skill_handlers.go:137`：闪光陷阱传的是字面量 `"法术"`
  - `internal/engine/runtime_policy_turn_hooks.go:26`：中毒的类型是 `"poison"`
  - 元素师另有 `"magic_no_absorb"`
- 现象：用严格比较的地方识别不到中毒、闪光陷阱这类法术伤害。英灵人形·双重回响用 `.(string)` 断言 DamageType，断言永远失败（见英灵人形-4）。
- 评审：

### 通用-6 手牌上限混用基础值和修正值，部分加牌路径不检查上限
- Bug｜中｜判决天平已核实，其余已确认或推测
- 证据：`arbiter/choices.go:61`、`magic_bow/skill_handlers.go:166` 用的是 `MaxHand`（基础值），其他地方用的是 `GetMaxHand`（修正值）
- 现象：角色身上有同生共死、怜悯、仲裁仪式之类的修正时，这两处判断会出错。另外有几处手牌增加后没有调用 `CheckHandLimit`，都是推测：神官·水之神力、封印师·封印破碎、血之巫女·流血脱离。
- 评审：

### 通用-7 "目标角色"多处被实现成"只能选敌方"
- 文案不符｜中｜已确认
- 涉及：
  - 元素师·元素点燃、雷击、风刃、陨石、火球、月光（冰冻却是任意角色）
  - 仲裁者·末日审判
  - 精灵射手·精灵密仪的释放
  - 红莲骑士·腥红十字
  - 阴阳师·黑暗祭礼
  - 瘟疫法师·死亡之触
  - 魔剑士·暗影流星
- 证据：
  - `internal/data/characters.go:680, 692, 716, 728, 740, 752, 807, 1113, 1350`
  - `elf_archer/timing_hooks.go:88`
  - `crimson_knight/skill_handlers.go:180`
  - `onmyoji/timing_hooks.go:18`
  - `plague_mage/choices.go:347`
- 说明：文案里别的技能会明确写"目标对手"，所以"目标角色"按字面应该是任意角色。如果"只能选敌方"是有意的规则解释，统一确认一次即可。有测试固定了现状：`TestElementalistOffenseSkills_RejectAllyTargets`、腥红十字的 `OnlyTargetsEnemy`。
- 评审：元素师已裁定：独有技都是目标角色（任意角色）。其他技能按官网 FAQ 法术 Q1（"目标角色"可以指定包括自己在内的任何角色）应该同理，待确认要不要一并改。

### 通用-8 普通弃牌被当成"展示"公开，可能误触发展示类时点
- 文案不符 / Bug｜低～中｜已确认
- 涉及：
  - 神官·水之神力交给队友的牌：`internal/engine/skill_use_execution.go:14-19`
  - 蝶舞者·舞动：`butterfly_dancer/choices.go:563`
  - 圣弓·圣煌辉光炮调整手牌：`holy_bow/choices.go:1092-1100`
  - 兽灵武士·兽返的自弃：`beast_samurai/choices.go:451`（推测）
- 现象：文案没有［展示］的弃牌也调用了 `NotifyCardRevealed`，会触发冰霜祷言、五系封印等展示类时点。神官那张牌是交给队友的，画面上还会显示成"弃牌"。
- 评审：

### 通用-9 前端兜底规则写错（只在服务端没下发该技能时生效）
- 前端体验｜低｜已确认
- 证据：
  - `web/src/components/ActionPanel.vue:717-718`：战纹碎击写成要求魔纹≥1，应为战纹≥1；魔纹融合写成要求魔纹≥2，应为≥1
  - `web/src/components/ActionPanel.vue:731`：判决天平要求审判≥1，但这个技能是加审判的，不消耗审判
  - `web/src/components/ActionPanel.vue:598-601`：`isElementalistExclusivePromptFlowSkill` 永远返回 false
- 评审：

### 通用-10 死代码和残留文件
- 代码卫生｜低｜已确认
- `magic_lancer/runtime.go.tmp`：已被 git 追踪
- `hero/turn_hooks.go:13-49`：ExhaustionRelease、TauntStartup 从未被调用
- `sword_emperor/runtime.go:28-74`：ResolveAttackMiss、ResolveAttackHitAftermath 没有调用方
- `berserker/timing_hooks.go:13`：hitCheckHook 没有注册
- `sage/choices.go:156-180, 288-299`：`sage_holy_target_count` 这一步进不去
- `elf_archer/choices.go:43-47, 165-255`：旧的确认链路推不出来，但路由还注册着
- `adventurer/skill_handlers.go:26, 106-116`：如果被接回分派，地下法则会让战绩区重复 +2
- `blood_priestess/runtime.go:111`：ResolveBleedExitOnActionEnd 是没人调用的重复实现
- `moon_goddess/token.go:8, 18`：上限常量和实际生效的 10 矛盾
- 评审：

---

## 二、前端通用发动流程

### 前端-1 元素师 5 个独有技卡在"选独有技手牌"这一步，手牌点不动
- Bug｜高｜已核实
- 影响：雷击、冰冻、风刃、陨石、火球
- 证据：
  - `web/src/components/ActionPanel.vue:587-592`：`requiresExclusiveCardSelectionBeforeTarget` 对元素师放行，于是进入 `choosing_exclusive`
  - `web/src/components/GameBoard.vue:2258-2263`：同名函数只认 `cost_discards === 0`，返回 false，导致 `isCardSelectableForExclusiveSkillCard` 永远为 false；点牌时 `toggleExclusiveSkillHandCard` 直接 return
  - `web/src/components/__tests__/ActionPanel.spec.ts:464-530`：测试直接写 store，绕过了 GameBoard 的点牌逻辑，所以测不出来
- 现象：元素师 → 法术行动 → 点【雷击】，提示"请在手牌区选择对应的独有技手牌"，但手牌全灰，点了没反应，只能取消。
- 评审：

### 前端-2 提交前就清掉选择，后端一报错就要从头再来；等回包期间行动枢纽会重新出现
- Bug / 前端体验｜中｜清空和不恢复已确认，枢纽重新出现是推测
- 证据：
  - `web/src/composables/useSubmitAction.ts:116-121`：发 ws 之前就 `clearSkillMode()`
  - `web/src/network/gameplayMessageHandlers.ts:410-421`：error 分支只弹 toast
  - `web/src/components/ActionPanel.vue:123-129`：`isIdleMainTurnPanel` 的判断
- 现象：后端校验失败时，比如目标不合法、弃牌系别不符（`internal/engine/skill_use_validation.go:64-75`），已选的弃牌和目标全部丢失。提交到回包之间，以及 prompt gate 压住提示的最多 8 秒内，攻击 / 法术 / 特殊的行动枢纽会重新显示，可以再点一个行动。
- 评审：

### 前端-3 多目标技能：点已选目标是"发动"而不是"取消选中"；选满后再点会悄悄挤掉最早选的
- 前端体验｜中｜已确认
- 影响：min<max 的多目标技能，例如治愈之光、圣疗、天使祝福、圣洁法典
- 证据：`web/src/components/GameBoard.vue:1687-1698, 1709-1712`；提示文案在 `web/src/components/ActionPanel.vue:1419-1421`
- 现象：圣疗选了 A，想改成 B，于是去点 A 想取消，结果直接对 A 发动了。界面上没有"确认发动"按钮。
- 评审：

### 前端-4 弃牌选满后自动进入下一步，没有回退；无目标技能点完最后一张就直接发动
- 前端体验｜中｜已确认
- 证据：`web/src/components/ActionPanel.vue:927-938`（watch 选满即推进）、`:881-890`（target_type 0/1 直接提交）、`:1439-1440`（只有"取消"，没有"上一步"）、`:1366-1368`（无目标技能也显示"② 选目标"）
- 现象：风之洁净点一张风系牌就发动了，点错了没法撤。需要目标的技能如果弃错了牌，只能取消整个技能重来。
- 评审：

### 前端-5 单目标技能点头像就直接发动，没有确认
- 前端体验｜中｜已确认
- 影响：所有 max_targets=1 的主动技能，以及 min=max 的多目标技能（如魔爆冲击 2/2、灵符·雷鸣 2/2）
- 证据：`web/src/components/GameBoard.vue:1724-1735`
- 现象：移动端误触，或者只是想点头像看看信息，都会直接扣费发动。
- 评审：

### 前端-6 目标合法性只做了阵营粗筛，多数限制要等后端报错
- 前端体验｜中｜已确认
- 证据：`web/src/composables/useBattleInteractionState.ts:179-218`（type 5/6 直接返回全场）；后端规则在 `internal/engine/skill_target_rules.go:47-130`
- 现象：不合法的目标也会高亮、可点。点了直接提交，再叠加前端-2，就变成"报错 + 从头再来"。
- 评审：

### 前端-7 弃牌上限用的是 cost_discards，不是 requiredDiscardCount
- Bug｜低｜不一致已确认，实际体验影响是推测
- 证据：`web/src/components/GameBoard.vue:2209, 2321-2324` 用 `cost_discards`；`web/src/components/ActionPanel.vue:933, 1374, 1121-1132` 用 `requiredDiscardCount`
- 影响：神官·神圣领域、水之神力这类需求会降到手牌数的技能
- 评审：

### 前端-8 应战被后端拒绝后，目标看着还是选中状态，要点两次才能重新提交
- Bug｜低｜已确认
- 证据：`web/src/components/PromptDialog.vue:1667-1675`、`web/src/components/GameBoard.vue:1621-1625`、`web/src/network/gameplayMessageHandlers.ts:410-421`
- 影响：攻击响应里的应战，以及魔弹传递
- 评审：

### 前端-9 错误恢复靠匹配报错文案的子串
- Bug（脆弱）｜低｜已确认
- 证据：`web/src/network/gameplayMessageHandlers.ts:414-419` 匹配"技能发动失败""无效的卡牌索引"，而后者只出现在 `magical_girl/interrupt.go:574, 604`
- 现象：prompt 内提交失败时，选中的牌和目标原样保留，也没有"哪一步错了"的提示。后端改了文案，前端的清理逻辑就会静默失效。
- 评审：

### 前端-10 prompt gate 压住提示期间，看不出接下来该谁操作
- 前端体验｜低｜推测
- 证据：`web/src/network/gameplayMessageHandlers.ts:220-230`：gate 期间 prompt 为 null、waitingFor 为空，`ActionPanel.vue:1463` 的等待面板也不显示
- 现象：最长 8 秒内界面上没有任何提示，演出结束后提示才突然弹出。
- 评审：

检查过、没发现问题的地方：
- state_update 到达时，`syncAfterStateUpdate`（`web/src/stores/interrupt.store.ts:154-162`）会清掉全部选择状态。
- 新 prompt 到达时，`setPrompt` 会清掉选择状态。
- 手牌变化后，选中的下标会被清掉，没有跨技能残留。
- 不可发动的技能会置灰，并附带服务端给的原因。
- 只有一个响应技能可用时会自动预选。

---

## 三、按角色

### 祈祷师（prayer_master）

#### 祈祷师-1 威力赐福：+2 伤害完全没有实现
- Bug｜高｜已核实
- 证据：`prayer_master/choices.go:58-73`、`prayer_master/timing_hooks.go:62-82`
- 现象：攻击命中后选"是"，代码只移除赐福牌并打一行"本次攻击伤害+2"的日志。中断上下文里没有带伤害对象，也没有任何代码改 `pd.Damage`，结果赐福牌白白消耗，伤害不变。现有测试只覆盖放置，没覆盖响应。
- 评审：Bug 成立（已裁定）。+2 加在原有的攻击伤害上，也就是本次攻击伤害 +2。

#### 祈祷师-2 光辉信仰：把"弃 2 张牌"当成发动成本
- 文案不符｜低｜已确认
- 证据：`internal/data/characters.go`（prayer_radiant_faith 的 `CostDiscards: 2`）、`docs/rule.md:262-270`（部分发动规则）
- 现象：文案里的"你弃 2 张牌"是效果，不是发动条件。按部分发动规则，手牌不足 2 张时也应该能发动，现在按钮是灰的。
- 评审：

#### 祈祷师-3 祈祷符文上限 3（文案歧义）
- 证据：`prayer_master/timing_hooks.go:36`（`AddToken(...,2,3)`）
- 文案没写符文上限，需要确认 3 是否正确。
- 评审：已裁定：上限 3 正确，保持现状。

没发现问题：祈祷、黑暗诅咒、迅捷赐福、法力潮汐。

### 吟游诗人（bard）

#### 吟游诗人-1 不谐和弦：取消后再确认会重复扣灵感
- Bug｜高｜已确认
- 证据：`bard/choices.go:47-51`（mode 步骤的策略是 `CancelPolicyBack`）、`bard/choices.go:464-492`
- 现象：选完 X 就已经执行了 `addBardInspiration(-x)`，并脱离囚徒形态。在下一步选分支时点取消，会退回选 X，再确认一次就又扣一次灵感。
- 评审：

#### 吟游诗人-2 不谐和弦：被弃牌的对手可能可以取消
- Bug｜中｜推测（需要确认前端是否显示取消按钮）
- 证据：`bard/choices.go:51`（弃牌步骤是 `CancelPolicyAbort`）、`bard/choices.go:575-578`（中断的 PlayerID 换成了弃牌者）
- 现象：弃牌提示发给被弃牌的人。如果他那边能点取消，就能逃过弃牌，而灵感已经扣了。
- 评审：

#### 吟游诗人-3 不谐和弦：目标可以选自己（文案歧义）
- 证据：`bard/choices.go:510`
- 现在目标列表包含吟游诗人自己，选自己时"各摸 X-1"会让他摸两次。"目标角色"包不包括自己？
- 评审：

#### 吟游诗人-4 沉沦协奏曲：回合限定标记只在吟游诗人自己回合重置
- Bug｜中｜已核实
- 证据：`bard/timing_hooks.go:176-190`、`bard/choices.go:404`、`internal/engine/turn_progression_runtime.go:72`
- 现象：`bd_descent_prompted` 和 `bd_descent` 记在吟游诗人自己的 TurnState 里。在其他回合触发过一次之后，这一轮里后面的回合不会再弹出，要等吟游诗人自己的回合才重置。按"每个玩家的回合各限一次"的裁定，这是错的。
- 评审：已裁定：弹出询问但拒绝了，算本轮已经发动过。另按"每个玩家的回合各限一次"的裁定，沉沦协奏曲在任意回合结束时各能触发一次，所以标记要改成在每个回合开始时重置，现在的实现会在一轮里漏掉后面的触发。

#### 吟游诗人-5 沉沦协奏曲：触发时机（文案歧义）
- 证据：`bard/timing_hooks.go:166`（`TimingTurnEndPreExtra`）
- 文案是"……且结算完之后"，代码统一放在回合结束时检查，而不是第 2 名对手受伤结算完的那一刻。另外，这个时点本身也受通用-3 影响。
- 评审：

#### 吟游诗人-6 沉沦协奏曲：第二张牌不按系别过滤
- 前端体验｜中｜已确认
- 证据：`bard/choices.go:358-392`
- 现象：选完第 1 张后，候选牌不按它的系别过滤，第 2 张选了不同系的牌才报"需弃置同系牌"。
- 评审：

#### 吟游诗人-7 胜利交响诗：不能用的选项也照样显示
- 前端体验｜低｜已确认
- 证据：`bard/choices.go:188-201, 897-904`
- 现象："提炼星石"选项总是显示。能量已满或战绩区没有星石时，点了才报错。
- 评审：

#### 吟游诗人-8 希望赋格曲：转移目标的范围（文案歧义）
- 证据：`bard/choices.go:1038-1043`
- 文案是"我方另一名目标角色"，代码排除了吟游诗人自己。"另一名"可以理解为"当前持有者以外的人"（包括自己），也可以理解为"其他队友"。
- 评审：

#### 吟游诗人-9 希望赋格曲：放置分支可以免费换位
- Bug｜低｜推测（需要设计确认是否允许）
- 证据：`bard/runtime.go:58-64`
- 现象：永恒乐章已经在场时，选"放置"也会先收回再放出，相当于不弃牌、不加成就把它转给了另一名队友。
- 评审：

没发现问题：禁忌诗篇、激昂狂想曲。

### 天使（angel）

#### 天使-1 天使之歌：挂在"回合开始时"，文案是"回合开始前"
- 文案不符｜高｜时点已核实，和中毒结算的先后关系已确认
- 证据：`internal/data/characters.go`（angel_song 用的是 `TimingTurnStart`）、`internal/engine/turn_fsm_dispatcher.go:95-150`
- 现象：引擎的顺序是"回合开始前 → BeforeAction（中毒/虚弱在这里结算）→ 回合开始时"。现在没法用天使之歌先解掉自己身上的中毒或虚弱。
- 评审：

#### 天使-2 天使羁绊：放置中毒或虚弱时也会触发
- Bug｜中｜已确认
- 证据：`angel/skill_handlers.go:154-164`、`internal/engine/magic.go:160, 179`
- 现象：`CanUse` 只判断 BuffID 是不是基础效果，不区分是新增还是移除。天使用法术牌放置中毒或虚弱时，也会 +1 治疗。文案只包括"移除基础效果"和"使用圣盾"。
- 评审：

#### 天使-3 天使羁绊：自己身上的封印被动移除时也会触发
- Bug｜低｜推测
- 证据：`internal/engine/interface.go:325-326`、`internal/engine/damage_role_runtime_hooks.go:101`
- 现象：`RemoveFieldCard` 默认把 sourceID 设成持有者本人，封印在伤害结算后自动移除时，会被当作"天使移除了一个基础效果"。
- 评审：

#### 天使-4 风之洁净：场上没有基础效果也能发动
- 文案不符｜中｜已确认
- 证据：`angel/skill_handlers.go:236-245, 281-286`、`docs/rule.md:264`
- 现象：规则写明场上没有基础效果时不能发动。现在只检查手里有没有风系牌：弃牌、消耗一次行动，然后直接"跳过移除"。
- 评审：

#### 天使-5 天使之墙相关的 UI 文档过时
- 文档｜低｜已确认
- 证据：`docs/character_skill_ui_config.md:7`
- 现象：文档说羁绊"无需弹窗"，实际每次都会弹出选目标。文案里有"目标角色"，弹窗是合理的，要更新的是文档。
- 评审：

没发现问题：天使祝福、神之庇护。

### 狂战士（berserker）

只有死代码一项，见通用-10。

没发现问题：狂化、撕裂、血腥咆哮、血影狂刀。

### 神箭手（archer）

#### 神箭手-1 闪光陷阱：伤害类型写成字面量 "法术"
- Bug｜中｜已确认
- 证据：`archer/skill_handlers.go:137`
- 现象：传的是 `"法术"`，不是 `model.MagicAttack`。用 `IsMagicDamageType` 判断的地方不受影响，但蝶舞者的毒粉和镜花水月识别不到（见通用-5、蝶舞者-1）。
- 评审：

没发现问题：贯穿射击、闪电箭、狙击、精准射击。

### 神官（priest，没有任何单测）

#### 神官-1 水之神力：只剩 1 张牌时不能发动
- 文案不符｜中｜已确认
- 证据：`priest/module.go:46-62`、`priest/skill_handlers.go:69-74`、`web/src/components/ActionPanel.vue:1127-1130`
- 现象：前端在只剩 1 张水系牌时允许只弃 1 张，但后端 `CanUse` 要求手牌至少 2 张，`ValidateDiscardedCards` 也要求恰好 2 张。
- 评审：

#### 神官-2 水之神力：交给队友的牌被公开成"弃牌"
- Bug｜中｜已确认
- 证据：`internal/engine/skill_use_execution.go:14-19`（见通用-8）
- 评审：

#### 神官-3 水之神力：队友收牌后不检查手牌上限
- Bug｜中｜推测
- 证据：`priest/skill_handlers.go:93-94`（直接 `append`，没调 `CheckHandLimit`）
- 评审：

#### 神官-4 神圣领域：手牌不足 2 张时不能发动
- 文案不符｜中｜已确认
- 证据：`priest/module.go:69-75`、`internal/engine/action_skill_runtime_rules.go:127-160`、`docs/rule.md:268-270`
- 现象：规则明确写了神官没有手牌也能发动、只剩 1 张就弃 1 张，前端也按这个处理了。但后端 `ResolveDiscardCount` 固定返回 2。
- 评审：

#### 神官-5 神圣领域：伤害分支不能选自己（文案歧义）
- 证据：`priest/skill_handlers.go:196-200`
- 评审：按官网口径：伤害分支写的是"目标角色"，应该能选自己。现在不能选，要修。

#### 神官-6 神圣契约：X 的上限没考虑目标的治疗上限
- 文案歧义 + 前端体验｜低｜已确认
- 证据：`priest/choices.go:125-140`
- 现象：X 的上限是神官自己的治疗量。目标已有 3 点治疗时选 X=3，神官扣 3 点，目标只加 1 点；目标治疗超过 4 时，一点都加不上。多出来的治疗是否丢失，文案没写；选项上也没有提示。
- 评审：

没发现问题：神圣启示、神圣祈福、圣使守护。

### 蝶舞者（butterfly_dancer）

#### 蝶舞者-1 毒粉 / 镜花水月：中毒、闪光陷阱等法术伤害不触发
- Bug｜高｜已核实
- 证据：`butterfly_dancer/timing_hooks.go:54`、`butterfly_dancer/runtime.go:126`（见通用-5）
- 评审：

#### 蝶舞者-2 镜花水月：伤害对象（文案歧义）
- 推测
- 证据：`butterfly_dancer/choices.go:1117-1130`
- 文案是"你对他造成 2 次法术伤害"。代码里的"他"是原伤害的承受者，日志却写"改写本次伤害来源"。另外两次伤害是追加到队尾的，可能排在其他待结算伤害之后。
- 评审：

#### 蝶舞者-3 凋零：茧溢出被舍弃时不触发（文案歧义）
- 证据：`butterfly_dancer/choices.go:591-612`
- "你每次移除茧时"包不包括因超出上限而舍弃？
- 评审：

#### 蝶舞者-4 舞动：弃牌被公开
- 文案不符｜低｜已确认
- 证据：`butterfly_dancer/choices.go:563`（见通用-8）
- 评审：

#### 蝶舞者-5 舞动：茧的来源牌靠"手牌最后一张"推断
- Bug + 前端体验｜低｜推测
- 证据：`butterfly_dancer/choices.go:518-531`
- 现象：茧是先摸进手牌、再把最后一张移出来的。如果摸牌被响应打断，"最后一张"可能不是刚摸到的那张。前端还会看到手牌先 +2 再 -1。
- 评审：

#### 蝶舞者-6 朝圣：要两步操作
- 前端体验｜低｜已确认
- 证据：`butterfly_dancer/choices.go:1027-1040`
- 现象：先确认"是否发动"，再选具体哪个茧。
- 评审：

没发现问题：生命之火、蛹化、倒逆之蝶。

### 魔弓（magic_bow）

#### 魔弓-1 魔贯冲击：+1 伤害被永久写进实体攻击牌
- Bug｜高｜已核实
- 证据：`magic_bow/choices.go:407`（`rawCtx.EventCtx.Card.Damage++`）、`internal/engine/before_action_phase_runtime.go:172-180`（`card := *head.Card` 后原样放入弃牌堆）
- 现象：加伤改的是卡牌本身。这张牌进弃牌堆、重洗回牌库后，谁抽到都多 1 点伤害；每发动一次魔贯冲击，就再永久加 1。加伤应该记在本次攻击的事件上下文或待结算伤害上。
- 评审：

#### 魔弓-2 雷光散射：额外 +X 被拆成第二次独立伤害
- Bug / 文案不符｜中｜已确认
- 证据：`magic_bow/choices.go:696-703, 742-747, 836-844`
- 现象：指定的对手先受 1 点，再单独受 X 点。文案是"本次对其造成的伤害额外 +X"，应该是一次 1+X 点。拆成两次后，治疗抵挡、承伤触发、按次数计算的效果都会多算一次。
- 评审：

#### 魔弓-3 魔贯冲击：判断"手牌已满"用的是基础上限
- Bug｜低｜已确认
- 证据：`magic_bow/skill_handlers.go:166`（`ctx.Target.MaxHand`），见通用-6
- 评审：

#### 魔弓-4 魔眼：两处和文案不一致
- 文案不符｜低｜已确认
- 证据：`magic_bow/skill_handlers.go:372`、`magic_bow/choices.go:909-919`
- 现象：
  1. CanUse 要求手里有牌，但分支②会先摸 3 张，没手牌时也应该能用。
  2. 分支①的目标没手牌时，实现改成魔弓自己摸 3 张，文案里没有这个替代效果。
- 评审：

#### 魔弓-5 魔眼：摸 3 张后，爆牌和放充能的先后顺序
- Bug｜中｜推测
- 证据：`magic_bow/choices.go:790-812`；对照 `magic_bow/choices.go:553-576`，充能技能专门处理了"先弃到上限再放充能"
- 现象：玩家可能先把一张牌放成充能，从而躲掉爆牌、少掉士气。
- 评审：

没发现问题：多重射击、充能。

### 圣弓（holy_bow，以本体文案为准）

#### 圣弓-1 自动填充：普通形态下做过特殊行动，回合结束仍然会问
- Bug｜高｜已核实
- 证据：`holy_bow/policy_hooks.go:18-30`（只在圣煌形态下设置 `hb_special=1`）、`holy_bow/timing_hooks.go:33`
- 现象："是否执行过特殊行动"只看 `hb_special`，而全仓库只有一处在设置它。不在圣煌形态时做了购买、合成、提炼，回合结束时自动填充照样弹出。
- 评审：

#### 圣弓-2 自动填充：没有"不发动"选项（文案歧义）
- 证据：`holy_bow/choices.go:427-448`、`holy_bow/choices.go:144-151`
- 现在只要能付水晶或宝石，就必须选一个分支付费。本体文案是"你选择以下一项发动"，trick 版写的是"你可以"，数据描述是"可选择"。可选的话就应该能拒绝。
- 评审：已裁定：可以不发动，要加一个"不发动"选项。

#### 圣弓-3 流星圣弹：不能选自己
- 文案不符｜中｜已确认
- 证据：`holy_bow/skill_handlers.go:337, 357`（`skillAllies(..., false)`）；测试 `MeteorBullet_RequiresOtherAlly` 固定了这个行为
- 现象：文案是"我方目标角色"。没有其他队友时，这个技能直接不能用。
- 评审：

#### 圣弓-4 圣煌辉光炮：士气不落后时不能发动
- 文案不符｜中｜已确认（需要确认是否有意为之）
- 证据：`holy_bow/skill_handlers.go:388-391`、`holy_bow/choices.go:1078`；测试 `CannotUseWhenMoraleNotBehind` 固定了这个行为
- 现象：文案是"移除 4 点信仰，并额外移除等同我方落后士气的信仰数"，不落后时成本就是 4，并没有禁止发动。
- 评审：

#### 圣弓-5 圣煌辉光炮：调整手牌时系统替玩家决定弃哪几张
- Bug / 文案不符｜中｜已确认
- 证据：`holy_bow/choices.go:1092-1100`
- 现象：手牌超过 4 张时，直接丢掉 `Hand[4:]`，玩家没法选。弃牌还会被公开（见通用-8）。
- 评审：

#### 圣弓-6 圣屑飓暴：未命中分支要求队友手牌不少于 X（文案歧义）
- 证据：`holy_bow/runtime.go:8-24`、`holy_bow/choices.go:647-650`
- 文案只写"目标队友弃 X 张牌"，没要求队友手牌够 X 张。
- 评审：

没发现问题：天之弓、圣煌降临、圣光爆裂。

### 圣女（saintess）

#### 圣女-1 圣疗：提示写"不超过 3"，后端要求正好等于 3
- Bug / 前端体验｜中｜已核实
- 证据：
  - 提示：`saintess/interrupt.go:127`
  - 校验：`saintess/interrupt.go:186`（`first+second != 3` 就报错）
  - 另一处用的是 `<=3`：`saintess/interrupt.go:70`
  - 前端允许总和 ≤3：`web/src/components/PromptDialog.vue:407-413`
- 现象：分配 1/1 或 2/0，前端允许提交，后端报错。这时水晶已经扣了，也不能取消。
- 评审：已裁定：分配上限是 3，可以小于 3。后端 `first+second != 3` 是 bug，要改成 ≤3；提示文案和前端是对的。

#### 圣女-2 冰霜祷言：展示牌也会触发（文案歧义）
- 证据：`saintess/module.go:23-30`、`internal/engine/notify_prompt_runtime.go:150-154`
- 文案是"使用水系牌或圣光时"，数据描述是"打出或展示"。现在公开弃牌（比如作为技能消耗正面弃掉的水系牌）也会触发。
- 评审：

没发现问题：治愈之光、治疗术、怜悯。

### 月之女神（moon_goddess）

#### 月之女神-1 月渎：多了"只在自己回合"的限制
- 文案不符｜中｜已确认
- 证据：`moon_goddess/timing_hooks.go:59-64`、`moon_goddess/runtime.go:119-124`；测试 `TestMoonGoddessBlasphemy_TargetLockedToDamagedEnemyAndSelfTurn` 固定了这个行为
- 现象：在对手回合造成的法术伤害（比如美杜莎之眼的那 1 点），永远接不上月渎。文案只写了"回合限定"（见通用-4）。
- 评审：Bug 成立（已裁定）：回合限定是每个回合各限一次，月渎在对手回合也能触发，要删掉"只在自己回合"的限制，同时改测试 `TestMoonGoddessBlasphemy_TargetLockedToDamagedEnemyAndSelfTurn`。

#### 月之女神-2 月渎：目标范围（文案歧义）
- 证据：`moon_goddess/timing_hooks.go:65-67, 85`
- 实现要求受伤的必须是敌方，追加伤害也锁定在同一个人身上。另一种理解是可以另选任意对手，而且受伤的"目标角色"也可以是队友。
- 评审：

#### 月之女神-3 美杜莎之眼："视为攻击"不触发（文案歧义）
- 证据：`moon_goddess/runtime.go:163`；测试 `ExcludesConvertedAttacks`
- 圣屑飓暴、欺诈这类"视为攻击"的情况被排除了。它们算不算"对手攻击时①"？
- 评审：

#### 月之女神-4 美杜莎之眼：没有手牌时照样造成伤害（文案歧义）
- 证据：`moon_goddess/choices.go:363-374`
- "你弃 1 张牌，对目标对手造成 1 点"：弃牌是不是前提？
- 评审：

#### 月之女神-5 闇月斩：加伤可能加错对象
- Bug｜低｜推测
- 证据：`moon_goddess/choices.go:848-856`
- 现象：只取队列里第一条 Damage>0 的伤害，不检查来源和类型。
- 评审：

#### 月之女神-6 新月和石化的上限常量互相矛盾
- 代码卫生｜低｜已确认
- 证据：`moon_goddess/token.go:8, 18`（2 和 3）对比 `moon_goddess/choices.go:742-744`（10，实际生效的是这个）
- 文案没写上限，需要确认正确值。
- 评审：

#### 月之女神-7 月之轮回、苍白之月：提示信息不全，付费后不能取消
- 前端体验｜低｜已确认
- 证据：`moon_goddess/choices.go:96`、`moon_goddess/skill_handlers.go:268-276`
- 现象：月之轮回分支①没有提示"移除闇月会让我方士气 -1"。苍白之月宝石扣完之后，后面四步都不能取消（见通用-2）。
- 评审：

没发现问题：新月庇护、闇月诅咒。

### 风之剑圣（blade_master）

#### 风之剑圣-1 疾风技：应战时也会触发
- Bug｜中｜触发路径已确认，后果是推测
- 证据：`blade_master/skill_handlers.go:108-114`（没有重写 CanUse）、`internal/engine/combat_response_runtime.go:488-496`、`internal/engine/runtime/skill/eligibility.go:162-169`
- 现象：用疾风技牌应战时，会静默追加一个攻击行动，挂在对手的回合里。文案限定"作为主动攻击打出时"。同一文件里列风技和圣剑都判断了 `CounterInitiator != ""`，只有疾风技漏了。
- 评审：

#### 风之剑圣-2 圣剑：摸 X 弃 X 可能多弹一次
- Bug｜低｜推测
- 证据：`blade_master/interrupt.go:20`、`internal/engine/runtime_role_post_resolution.go:36-43`
- 现象：只判断 `AttackCount == 3`，不看这次结束的是不是攻击行动。第 3 次攻击之后同回合再有一次法术行动时，可能会再弹一次。
- 评审：

没发现问题：风怒追击、剑影、列风技、圣剑的强制命中。

### 贤者（sage）

#### 贤者-1 法术反弹：自伤先于对目标的伤害结算
- Bug｜低～中｜已确认
- 证据：`sage/choices.go:354-371, 631-635`
- 现象：两笔伤害按"对目标、对自己"的顺序逐个插到队首，结果自伤排在前面，可能先触发我方士气下降或智慧法典。
- 评审：

#### 贤者-2 魔道法典、圣洁法典：取消时报错或白扣宝石
- Bug｜中｜前半部分已确认，流程是否卡死是推测
- 证据：`sage/choices.go:39-46`、`internal/model/prompt_flow.go:263-266`、`internal/engine/skill_use.go:100-105`、`internal/engine/prompt_flow_cancel.go:39-41`
- 现象：在选牌这一步取消，会报"当前步骤无法返回"；在选目标这一步取消，宝石已经扣了，牌没弃，效果也没有（见通用-2）。
- 评审：

#### 贤者-3 圣洁法典：前端不拦重复系别
- 前端体验｜中｜已确认
- 证据：`sage/choices.go:136`（`card_filter: "distinct_elements"`），web 里没有对应处理
- 现象：可以选两张同系牌，提交后才收到报错。
- 评审：

#### 贤者-4 圣洁法典：有一个进不去的步骤
- 死代码，见通用-10。

没发现问题：智慧法典。

### 冒险家（adventurer，以本体文案为准）

#### 冒险家-1 冒险者天堂：没扣能量，流程也和文案不一样
- 文案不符｜高｜已核实
- 证据：`adventurer/choices.go:126-141`（`handleParadiseAllyPick`）
- 现象：
  1. 文案里的"然后移除你的 1［能量］"完全没有实现。
  2. 文案是冒险家自己提炼，再把提炼结果交给队友。实现是 `StartExtractForPlayer(ally)`：提示弹给队友，由队友自己选，上限按队友的剩余能量算。
- 评审：Bug 成立（已裁定）。冒险家有能量时必须移除 1 能量；没有能量时不扣，官网冒险家页也是这么写的。流程部分（由队友能量上限决定能否提炼）和官网说明一致，不用改。

#### 冒险家-2 冒险者天堂：自己能量满时进不去
- Bug｜中｜选项被隐藏已确认，服务端是否拒绝是推测
- 证据：`internal/engine/action_selection_policies.go:192-194`（只有自身能量没满才给"提炼"选项）、`adventurer/policy_hooks.go:46-57`
- 现象：policy hook 专门写了"自身能量满时直接选队友"的分支，但这种情况下菜单里根本没有提炼。测试是直接发 CmdExtract，绕过了菜单。
- 评审：

#### 冒险家-3 偷天换日：条件不满足也能发动，资源白扣
- Bug｜中｜资源白扣已确认，流程卡死是推测
- 证据：`adventurer/choices.go:67-86`、`adventurer/skill_handlers.go:120-122`
- 现象：CanUse 只检查水晶。对方战绩区没有宝石、或我方没有水晶时，两个模式都能选；选了之后只记一条"失败"日志，水晶和回合限定次数都已经消耗，也没有调 EnterExtraActionStage。前端也没把这两个选项置灰。
- 评审：

#### 冒险家-4 偷天换日：我方战绩区已满时选"转移"（文案歧义）
- 现在对方少 1 宝石，我方因为上限截断加不上。是规定"转移必须能放进我方才可选"，还是"照常扣，溢出作废"？
- 评审：

没发现问题：欺诈主体、强运、地下法则的实际购买改写。

### 勇者（hero）

#### 勇者-1 怒吼：知性上限写成了 3
- Bug｜中｜已核实
- 证据：`hero/miss_hooks.go:23-26`（上限 3），其他路径都用 `heroTokenCap = 4`（`hero/skill_handlers.go:11`）
- 现象：只靠怒吼攒知性，永远到不了 4，明镜止水就发动不了。
- 评审：

#### 勇者-2 禁断之力：已经横置时仍然给额外攻击（需要规则裁定）
- 文案不符｜中｜推测
- 证据：`hero/skill_handlers.go:137-139`
- 现象：已经横置时再发动，照样无条件"额外 +1 攻击行动"，水晶够的话可以一直连续追加。
- 评审：

#### 勇者-3 禁断之力：上下文为空时的处理
- Bug｜低｜推测
- 证据：`hero/skill_handlers.go:116, 125`
- 现象：AttackInfo 为空时一律按未命中处理；DamageVal 为空时火系加伤被丢掉，但自伤照样结算。
- 评审：

#### 勇者-4 精疲力竭：解除可能被推迟
- Bug｜低｜推测
- 证据：`hero/timing_hooks.go:86`（额外要求 `!HasUsedActionSkill`）
- 评审：

#### 勇者-5 明镜止水：待发的 +1 水晶可能残留
- Bug｜低｜推测
- 证据：`hero/skill_handlers.go:175-176`、`hero/timing_hooks.go:54-57`
- 评审：

#### 勇者-6 挑衅："必须攻击勇者"不检查勇者能不能被选为目标
- Bug｜低｜推测
- 证据：`hero/policy.go:27, 45`
- 评审：

#### 勇者-7 用词和提示不全
- 前端体验｜低｜已确认
- 证据：`hero/miss_hooks.go:28`、`internal/engine/action_summary.go:537`、`web/src/constants/battleMarks.ts:88-89`
- 现象：日志和行动摘要把"知性"写成了"智慧"。怒气、知性的浮层提示漏了几个获得和消耗的来源。
- 评审：

没发现问题：勇者之心、死斗。

### 魔枪（magic_lancer）

#### 魔枪-1 幻影星尘：自伤后有待弃牌时，流程续不上
- Bug｜高｜已核实
- 证据：`magic_lancer/timing_hooks.go:51-54`、`internal/engine/hand_overflow_resolution.go:214`
- 现象：钩子记下 `ml_stardust_wait_discard` 后就返回，等弃牌结束再续接。但续接要求弃牌中断带有 `flow_continuation_role_id`，这个字段只有格斗家和魔弓设置过，魔枪没设。结果是角色不转正、手牌上限一直是 5，"未掉士气则对目标造成 2 点"也丢了。相关测试的手牌都是空的，覆盖不到这条路径。
- 评审：

#### 魔枪-2 黑暗束缚：前端不知道魔枪不能出法术牌
- 前端体验｜中｜已确认
- 证据：`web/src/components/ActionPanel.vue:1188`、`web/src/components/GameBoard.vue:2250-2253`、`internal/server/state_view.go:204`（没有下发 CannotCastMagic）、`internal/engine/action_submission_runtime.go:155`
- 现象：法术牌照样能选，提交后才被拒。报错写的是"当前形态不能……"，但魔枪是始终不能用，和形态无关。
- 评审：

#### 魔枪-3 暗之解放：未命中时加成不消耗
- 文案不符｜中｜已确认
- 证据：`magic_lancer/timing_hooks.go:18-30`、`internal/engine/pending_damage_runtime.go:128-140`
- 现象：+1 要到伤害计算时才消耗。下一次主动攻击如果未命中，加成会留给本回合后面的攻击。文案写的是"下一次主动攻击"。
- 评审：

#### 魔枪-4 漆黑之枪："手牌为 1 或 2"在什么时点判断（文案歧义）
- 证据：`magic_lancer/skill_handlers.go:216-219`
- 实现是在命中后判断，另一种理解是在攻击宣言时判断。
- 评审：

#### 魔枪-5 漆黑之枪、充盈：点了发动就收不回
- 前端体验｜低｜已确认
- 证据：`magic_lancer/choices.go:32, 56, 119-129`（见通用-2）
- 评审：

#### 魔枪-6 暗之障壁：法术牌和雷系牌能不能混着弃（文案歧义）
- 证据：`magic_lancer/choices.go:186-203`
- 实现是不能混着弃。
- 评审：

#### 魔枪-7 充盈：遗留的目标规则
- Bug（遗留配置）｜低｜推测
- 证据：`magic_lancer/module.go:78-83`
- 现象：策略里还留着"至多 1 名其他队友"的目标规则，可能误导前端或 bot。
- 评审：

没发现问题：暗之解放同回合禁用充盈和漆黑之枪、幻影形态手牌上限、黑暗束缚的后端拦截、充盈的主体逻辑。

### 魔剑士（magic_swordsman）

#### 魔剑士-1 暗影抗拒：只在暗影形态下生效
- 文案不符｜高｜已核实
- 证据：`magic_swordsman/form.go:20-22`（`BlocksMagicCasting` 要求 `InShadowForm`）、`magic_swordsman/cannot_act.go:12`
- 现象：文案是"在你的行动阶段你始终不能使用法术牌"，现在普通形态下能正常打法术。`docs/skill_animation_map.md:189` 也记了这处不一致。`docs/character_skills_config.md:3014` 写成"禁 5 张基础法术牌"，和文案口径也不同。
- 评审：

#### 魔剑士-2 暗影流星：转正后角色仍然横置
- Bug｜中｜已核实
- 证据：`magic_swordsman/choices.go:63`（只写了 `user.Form = ""`）、`internal/engine/player/form_helpers.go:43-53`（`ClearForm` 会同时重置 Orientation）
- 现象：没调 `ClearForm`，Orientation 还是横置，也没有派发姿态变化。下个行动阶段的释放 hook（`magic_swordsman/timing_hooks.go:96`）要先 `HasForm` 才处理，而形态已经是空的，所以角色会一直横置。
- 评审：

#### 魔剑士-3 暗影流星：星石固定先扣水晶
- 文案不符｜低｜已核实
- 证据：`magic_swordsman/choices.go:55-62`
- 现象：2 个星石固定先扣水晶，不够再扣宝石，玩家不能自己选。女武神·军威神光也是一样（女武神-2）。
- 评审：

#### 魔剑士-4 暗影流星：目标只能选敌方
- 见通用-7，证据是 `internal/data/characters.go:1113`。
- 评审：

#### 魔剑士-5 暗影之力：只给主动攻击加伤
- 文案不符｜中｜已确认
- 证据：`magic_swordsman/timing_hooks.go:15`（`ctx.CounterInitiator != ""` 时直接返回）
- 现象：文案是"你发动的所有攻击伤害额外+1"。暗影形态会持续到下个行动阶段开始，期间在别人回合应战时拿不到 +1。
- 评审：

#### 魔剑士-6 黄泉震颤：未命中后标记可能残留
- Bug｜低｜推测
- 证据：`magic_swordsman/timing_hooks.go:47-52`
- 现象：没检查 `pd.IsCounter`，只靠 `ms_yellow_spring_pending` 标记，而这个标记要到下次攻击宣言时才清。目前没找到能导致误结算的路径。
- 评审：

没发现问题：修罗连斩、暗影凝聚。

### 暗杀者（assassin）

#### 暗杀者-1 反噬：在治疗抵挡之前就触发
- 文案不符｜中｜已确认
- 证据：`internal/engine/game_damage_pipeline.go:113`（`dispatchPendingDamageTaken`）先于 `internal/engine/game_damage_pipeline.go:122`（`resolvePendingDamageHealChoice`）执行；判定在 `assassin/skill_handlers.go:20-23`
- 现象：文案的时点是⑥，也就是实际承受伤害之后。暗杀者用治疗把攻击伤害完全挡掉时，攻击者照样会摸 1 张。
- 评审：

#### 暗杀者-2 水影：弃牌框不过滤，选错了静默失败
- 前端体验｜中｜不过滤已确认，失败后的流程是推测
- 证据：`internal/data/characters.go:446-452`（没配 DiscardElement 和 DiscardType，`MaxSelect` 是 99）、`internal/engine/runtime/skill/runtime.go:271-287`、`assassin/skill_handlers.go:107-118`、`internal/engine/runtime/skill/executor.go:42`
- 现象：非水系牌、非潜行状态下的法术牌都能选上，到 Execute 阶段才报错。错误只写进日志，玩家看不到。
- 评审：

没发现问题：潜行。

### 格斗家（fighter）

#### 格斗家-1 百式幻龙拳：回合结束没清锁定标记
- Bug｜中｜已核实
- 证据：`fighter/hooks.go:79-93`（`turnEndHook` 只清形态和 target_order）、`fighter/runtime.go:40, 62-70`（只有 `ClearHundredDragon` 会调用 `clearHundredDragonLock`）
- 现象：正常打完一回合后，目标身上的锁定专属效果卡 `EffectFighterHundredDragonLock` 一直不会被移除。
- 评审：

#### 格斗家-2 百式幻龙拳：［持续］持续多久（文案歧义）
- 证据：`fighter/hooks.go:79`
- 现在每回合结束都退场，相当于只覆盖当前这一个行动阶段。另一种理解是一直持续，直到违反条件才取消。
- 评审：

#### 格斗家-3 百式幻龙拳：多一步选择，误点会丢形态
- 前端体验｜低｜已确认
- 证据：`fighter/skill_handlers.go:176-185`、`fighter/policy.go:51-57`
- 现象：
  - 发动时要先选锁定目标，这一步不能取消，而斗气已经扣了。
  - 行动阶段点了法术或特殊行动，会报错并直接取消形态。如果这些按钮没有置灰，误点就会丢掉形态。
- 评审：

#### 格斗家-4 斗气上限 8（文案歧义）
- 证据：`fighter/skill_handlers.go:12`
- 文案和 docs 里都没写上限，需要对照原版规则确认。
- 评审：已裁定：上限 8 正确，保持现状。

#### 格斗家-5 气绝崩击：自伤和攻击伤害的结算顺序
- Bug｜低｜推测
- 证据：`fighter/skill_handlers.go:220-227`
- 现象：自伤在攻击宣言时就进了伤害队列，可能先于攻击伤害结算。文案写的是"然后"。
- 评审：

没发现问题：念气力场、念弹、斗神天驱。

### 苍炎魔女（blaze_witch）

#### 苍炎魔女-1 魔能反转：第一步取消会报错，水晶已扣
- Bug｜中｜已确认
- 证据：`blaze_witch/choices.go:25`、`internal/model/prompt_flow.go:172-174`、`blaze_witch/skill_handlers.go:271`
- 现象：见通用-2。
- 评审：

#### 苍炎魔女-2 魔女之怒：恢复点的 key 对不上
- Bug｜低｜推测
- 证据：`blaze_witch/skill_handlers.go:117` 写入的是 `waiting_phase`，`blaze_witch/choices.go:169` 读的是 `resume_phase`
- 现象：`ApplyChoiceResumePoint(nil)` 实际什么都不做，选完摸牌数之后只能靠 Drive 的默认推进。
- 评审：

#### 苍炎魔女-3 天火断空：+1 加在哪一段（文案歧义）
- 证据：`blaze_witch/skill_handlers.go:63-66`
- 实现是目标和自己两段都 +1。另一种理解是只加给目标那段。
- 评审：

没发现问题：苍炎法典、替身玩偶、永生银时计、痛苦链接、烈焰形态。

### 剑帝（sword_emperor）

#### 剑帝-1 剑气斩：第一步取消会报错
- 前端体验｜低｜已确认
- 证据：`sword_emperor/choices.go:26`
- 现象：见通用-2。
- 评审：

#### 剑帝-2 剑魂上限 3、剑气上限 5（文案歧义）
- 证据：`sword_emperor/field.go:10`、`sword_emperor/token.go:8`
- 文案没写这两个上限，需要确认。
- 评审：已裁定：剑魂上限 3、剑气上限 5 正确，保持现状。

没发现问题：天使之魂、恶魔之魂、不屈意志、剑魂守护、佯攻。

### 灵符师（spirit_caster）

#### 灵符师-1 风行：没有手牌时多弹一个确认，点"取消"效果就丢了
- Bug / 前端体验｜中｜已确认
- 证据：`spirit_caster/skill_handlers.go:103-113`、`spirit_caster/choices.go:100-111, 300-303`
- 现象：弃完风系牌后手里没牌，会多弹一个"无手牌可放置为妖力，是否跳过？"。选"取消"会直接弹出中断，风行效果丢失，但弃牌和这次法术行动已经消耗。雷鸣在同样情况下没有这一步。
- 评审：

#### 灵符师-2 雷鸣 + 灵力崩解：没资源时也会问要不要消耗水晶
- Bug / 前端体验｜低｜已确认
- 证据：`spirit_caster/skill_handlers.go:59-70`、`spirit_caster/choices.go:497-499`
- 现象：无手牌这条路径没有检查 `CanPayCrystalCost`，选"是"会报错。
- 评审：

#### 灵符师-3 雷鸣：伤害路由可能被重复设置
- Bug｜低｜推测
- 证据：`spirit_caster/choices.go:751-753` 与 `spirit_caster/choices.go:529-531`
- 现象：`RoutePendingDamageWithReturn` 已经设了一次，外层又调用 `EnterDamageResolution(nil)`，可能覆盖恢复点。
- 评审：

没发现问题：念咒、百鬼夜行。

### 英灵人形（war_homunculus，没有任何单测）

#### 英灵人形-1 魔纹融合：手牌全同系时卡死
- Bug｜高｜已核实
- 证据：`war_homunculus/skill_handlers.go:150-152`（CanUse 只要求手牌 ≥2）、`war_homunculus/skill_handlers.go:191`（`MustBeginAt(runeChoiceStepCards)`）、`war_homunculus/choices.go:30-33`（cards 步骤的策略是 Back）、`war_homunculus/choices.go:208-216`（HandleCancel 只处理双重回响）
- 现象：主动攻击未命中、有魔纹、手牌是 2 张以上同系（比如 2 张火），确认发动后选不出合法组合，也取消不掉。
- 评审：

#### 英灵人形-2 战纹碎击、魔纹融合：发动后无法反悔
- 前端体验｜中｜已确认
- 证据：`war_homunculus/choices.go:26-33`
- 现象：战纹碎击总能选出 1 张，所以不会卡死，但点错了只能弃牌。
- 评审：

#### 英灵人形-3 双重回响：取消后回合次数已用掉
- Bug｜中｜已核实
- 证据：`internal/engine/runtime/skill/executor.go:37-39`、`war_homunculus/choices.go:209-214`（见通用-1）
- 评审：

#### 英灵人形-4 双重回响：伤害类型校验永远不生效
- Bug｜低｜已确认
- 证据：`war_homunculus/skill_handlers.go:239`、`internal/engine/settlement_timeline.go:53`
- 现象：用 `.(string)` 断言 `model.DamageType`，断言永远失败。
- 评审：已裁定：双重回响不限于攻击和法术伤害，官网写明中毒等伤害也能触发。这条校验应该删掉，而不是修好。它现在恰好不生效，所以行为是对的。

#### 英灵人形-5 双重回响："回合限定"按谁的回合算（文案歧义）
- 证据：`internal/engine/turn_progression_runtime.go:72`（见通用-4）
- 评审：已裁定：任意一个回合各限一次。现在的实现是一整个周期一次，要改成每个回合开始时重置。

#### 英灵人形-6 战纹碎击：+Y 加在哪里（文案歧义）
- 证据：`war_homunculus/choices.go:508-515`
- 文案是"本次法术伤害额外+Y"，但战纹碎击本身是攻击。现在的实现是另外推一笔 Y 点法术伤害。另一种理解是攻击伤害 +Y。
- 评审：已裁定：按技能文案，+Y 是"本次法术伤害额外+Y"。我理解为保持现状（另推一笔 Y 点法术伤害）。

#### 英灵人形-7 前端兜底规则写错
- 见通用-9。

没发现问题：战纹掌控、怒火压制、符文改造。

### 血色剑灵（crimson_sword_spirit）

#### 血色剑灵-1 散华轮舞：鲜血上限整回合都是 4，回合结束又被截回 3
- 文案不符｜中｜已确认
- 证据：`crimson_sword_spirit/choices.go:234-235`、`crimson_sword_spirit/timing_hooks.go:58-61`
- 现象：文案只说这一次 +2 可以无视上限、最高到 4。现在整个回合里血色荆棘也能加到 4；回合结束移除庭院时又强制截回 3，多出来的鲜血就丢了。
- 评审：

#### 血色剑灵-2 散华轮舞：先放置庭院，再扣水晶
- Bug｜低｜已确认（属于防御性问题，正常情况下 CanUse 已经挡住）
- 证据：`crimson_sword_spirit/choices.go:208-220`
- 评审：

#### 血色剑灵-3 血染蔷薇：两步选目标都不能取消
- 前端体验｜中｜已确认
- 证据：`crimson_sword_spirit/choices.go:22-24`
- 评审：

#### 血色剑灵-4 血染蔷薇："队友"是否包括自己（文案歧义）
- 证据：`crimson_sword_spirit/choices.go:89`（现在包括自己）
- 评审：按官网 FAQ 法术 Q1 的口径：文案写的是"任意 1 名队友"，队友不含自己。现在把自己也算进去了，要修。

#### 血色剑灵-5 血气屏障：反伤对象（文案歧义）
- 证据：`crimson_sword_spirit/skill_handlers.go:125-131`
- 现在自动反伤给伤害来源，不能另选；来源是队友时也会打队友。另一种理解是另选一名对手。
- 评审：已裁定："目标对手"是由使用者指定的任意一名对手。要改成让玩家选一名对手反伤，不再固定打伤害来源。

没发现问题：血色荆棘、赤色一闪、血蔷薇庭院。

### 女武神（valkyrie）

#### 女武神-1 英灵召唤：额外弃牌的 +1 治疗给了被攻击的敌人（文案歧义，建议优先确认）
- 证据：`valkyrie/choices.go:270-279`；测试 `TestValkyrie_HeroicSummon_ExtraDiscardHealsCurrentCombatTarget` 固定了这个行为
- 文案是"（若你额外弃 1 张法术牌［展示］）目标角色 +1［治疗］"。这点治疗在②加上，到④就能用来抵这次伤害，等于花一张牌把自己的 +1 伤害又抵消了，从设计上看大概率不对。另一种理解是任选一名角色，通常是队友。
- 评审：按官网口径："目标角色"由玩家任选（可以是自己或队友）。现在固定给被攻击的敌人，要修成让玩家选目标，同时改测试 `TestValkyrie_HeroicSummon_ExtraDiscardHealsCurrentCombatTarget`。

#### 女武神-2 军威神光：星石固定先扣水晶
- 文案不符 / 前端体验｜低｜已确认
- 证据：`valkyrie/choices.go:234-243`
- 评审：

#### 女武神-3 军威神光：三步都不能回到上一步
- 前端体验｜低｜已确认
- 证据：`valkyrie/choices.go:23-83`
- 评审：

没发现问题：神圣追击、秩序之印、和平行者。

### 元素师（elementalist）

#### 元素师-1 独有技卡死
- 见前端-1。

#### 元素师-2 攻击类技能只能选敌方
- 见通用-7。冰冻却可以选任意角色，口径不一致。
- 评审：已裁定：独有技的目标都是目标角色，要放开成任意角色，同时改测试 `TestElementalistOffenseSkills_RejectAllyTargets`。

#### 元素师-3 冰冻：两步选目标没有取消
- 前端体验｜低｜已确认
- 证据：`elementalist/choices.go:24-26`
- 评审：

没发现问题：元素吸收、元素点燃、雷击、风刃、陨石、火球、月光的结算本身。

### 圣枪骑士（holy_lancer）

#### 圣枪骑士-1 天枪：额外把攻击改成了暗系
- 文案不符｜低～中｜改写已确认，具体波及哪些效果是推测
- 证据：`holy_lancer/hooks.go:33`（`ElementOverride=Dark`）
- 现象：文案只说"对手无法应战"，没有"视为暗灭"。会影响按攻击元素判定的效果和提示显示。
- 评审：已裁定：天枪就是无法应战，视为暗灭可以接受；直接标记成无法应战也行。保持现状。

#### 圣枪骑士-2 圣击：没弹出地枪时 +1 会丢
- Bug｜低｜推测
- 证据：`holy_lancer/skill_handlers.go:69-71`、`holy_lancer/hooks.go:49-79`
- 现象：命中且自己有治疗时，圣击要等"跳过地枪响应"才补 +1。地枪没弹出时，圣击就丢了。
- 评审：

#### 圣枪骑士-3 惩戒：自己治疗已满时是否算转移（文案歧义）
- 证据：`holy_lancer/skill_handlers.go:47-50`
- 现在目标照样 -1，自己不 +1。
- 评审：

没发现问题：神圣启示、辉耀、地枪、圣光祈愈。

### 精灵射手（elf_archer）

#### 精灵射手-1 元素射击：取消后回合次数已用掉
- Bug｜中｜已核实
- 证据：`elf_archer/choices.go:84-102`（见通用-1）。测试 `elf_archer_skill_regression_test.go:40-78` 没有检查次数。
- 现象：在选牌界面取消后，同一回合风之矢给的额外攻击里就不能再用了。
- 评审：

#### 精灵射手-2 动物伙伴 / 宠物强化：只有攻击伤害能触发（文案歧义）
- 证据：`elf_archer/timing_hooks_post_damage.go:28`、`elf_archer/skill_handlers.go:102-107`
- 文案是"目标角色承受你造成的伤害后⑥"。地之矢的 1 点法术伤害、精灵密仪释放的 2 点伤害都不触发。characters.go 的描述写的是"主动攻击伤害"。
- 评审：已裁定：精灵射手在自己回合造成的任何伤害都能触发，包括法术伤害。

#### 精灵射手-3 动物伙伴：摸牌爆牌时，技能的弃 1 张被合并掉（文案歧义）
- 证据：`elf_archer/skill_handlers.go:189-193`、`elf_archer/choices.go:242-246`
- 现在一共只弃 1 张。另一种理解是爆牌弃牌和技能弃牌分开算，共弃 2 张。
- 评审：已裁定：先结算爆牌弃牌，再结算技能本身的"弃一张牌"，两者分开算。

#### 精灵射手-4 精灵密仪：释放目标只能选敌方
- 见通用-7，证据是 `elf_archer/timing_hooks.go:88`。
- 评审：

#### 精灵射手-5 精灵密仪：放置祝福失败时宝石已扣、形态已进入
- Bug｜低｜推测
- 证据：`elf_archer/skill_handlers.go:124-133`
- 现象：先进入形态，再"摸 3 张、挪成祝福"。摸牌被中断或摸不满 3 张时，Execute 报错，但费用和形态已经生效。另外，"牌库顶 3 张"走的是摸牌流程，可能误触发其他角色的摸牌钩子。
- 评审：

#### 精灵射手-6 精灵密仪：脱离形态时不派发姿态变化
- Bug｜低｜推测
- 证据：`elf_archer/choices.go:275` → `elf_archer/choices.go:359-368`（直接 `player.Form=""`）
- 现象：前端可能没有转正演出。和魔剑士-2 是同一类问题。
- 评审：

没发现问题：元素射击的五种元素箭效果、宠物强化的扣费。

### 魔法少女（magical_girl）

#### 魔法少女-1 魔弹掌控：每发魔弹都要选一次方向
- 前端体验｜低｜已确认
- 证据：`internal/engine/magic.go:128-137`
- 评审：

#### 魔法少女-2 魔弹融合：被魔弹指向时要多点 1～2 次
- 前端体验｜低｜已确认
- 证据：`internal/engine/magic.go:322-352`
- 现象：要先弹出响应技能窗口，选了融合才进入选牌。火系、地系牌不会出现在魔弹面板的"传递"里，玩家容易以为不能用。
- 评审：

#### 魔法少女-3 魔爆冲击：伤害和弃牌的先后顺序（文案歧义）
- 证据：`magical_girl/interrupt.go:550-556`
- 现在是先把伤害放进队列，马上让施法者弃 1 张，伤害要等中断结束才结算。
- 评审：

#### 魔法少女-4 魔爆冲击：目标没有法术牌时仍要手动点一次
- 前端体验｜低｜已确认
- 证据：`magical_girl/interrupt.go:149-166`
- 现象：会弹出一个只有"不弃牌（受到 2 点伤害）"这一个选项的选牌框。
- 评审：

没发现问题：毁灭风暴、魔弹融合（主动使用）。

### 灵魂术士（soul_sorcerer）

#### 灵魂术士-1 灵魂吞噬：只有承受伤害爆牌才加黄魂
- 文案不符｜中｜已确认
- 证据：`soul_sorcerer/runtime.go:33`（要求 `fromDamageDraw`）
- 现象：文案是"我方每有 1 点士气下降"。购买、合成、技能摸牌导致的爆牌，以及技能直接扣士气，现在都不加。characters.go 的描述已经改成"因承受伤害导致"，和原文案不一致。
- 评审：Bug 成立（已裁定）。只要我方士气下降就触发，去掉 `fromDamageDraw` 这个限制。

#### 灵魂术士-2 灵魂召还：不能取消
- 前端体验｜低｜已确认
- 证据：`soul_sorcerer/choices.go:113-162`
- 评审：

#### 灵魂术士-3 灵魂链接：转移的总是队列里的第一条伤害
- Bug｜低｜推测
- 证据：`soul_sorcerer/timing_hooks.go:115`、`soul_sorcerer/choices.go:309-342`（`damage_index` 固定为 0）
- 现象：毁灭风暴、血之悲鸣这类一次放进多条伤害的情况，可能减错伤害。
- 评审：

#### 灵魂术士-4 灵魂转换：每次主动攻击都会弹出三选一
- 前端体验｜低｜已确认（有取消按钮，只是打断节奏）
- 证据：`soul_sorcerer/timing_hooks.go:40-51`
- 评审：

#### 灵魂术士-5 灵魂链接："队友数>1"算不算自己（文案歧义）
- 证据：`soul_sorcerer/skill_handlers.go:222-225`
- 现在按"除自己以外的队友 ≥2"判断，所以 2v2 时不能发动。
- 评审：

没发现问题：灵魂镜像、灵魂震爆、灵魂赐予、灵魂增幅。

### 血之巫女（blood_priestess）

#### 血之巫女-1 流血：要等行动结束才检查是否脱离
- 文案不符｜中｜已确认
- 证据：`blood_priestess/turn_hooks.go:33-50`
- 现象：文案是"自身手牌<3 时强制脱离"。行动中途手牌降到 3 以下（比如逆流弃了 2 张），形态和同生共死的 +1 上限要到行动结束才失效。
- 评审：

#### 血之巫女-2 流血脱离后，同生共死持有者的手牌上限骤降，但不检查爆牌
- Bug｜中｜推测
- 证据：`blood_priestess/turn_hooks.go:46`、`blood_priestess/hand_limit.go:26-29`
- 现象：上限从 +1 变成 -2，没调 CheckHandLimit（见通用-6）。
- 评审：

#### 血之巫女-3 同生共死 / 血之哀伤：目标能不能是自己或现持有者（文案歧义）
- 证据：`blood_priestess/skill_handlers.go:22-34`、`blood_priestess/hand_limit.go:30-33`
- 现象：放在自己身上时，手牌上限只改一次；转给现持有者等于空转，还要白挨 2 点伤害。
- 评审：

#### 血之巫女-4 血之哀伤：两步都不能取消
- 前端体验｜低｜已确认
- 证据：`blood_priestess/choices.go:70-84`
- 评审：

没发现问题：逆流、血之悲鸣、血之诅咒。

### 仲裁者（arbiter）

#### 仲裁者-1 判决天平：按基础上限补牌
- Bug｜中｜已核实
- 证据：`arbiter/choices.go:61`（`maxHand := user.MaxHand`）
- 现象：身上有同生共死（-2）时，会补到 6 张，当场爆牌、扣士气；修正为正时又补不满。
- 评审：

#### 仲裁者-2 仲裁仪式："手牌上限恒定为 5"没有实现
- 文案不符｜中｜已确认
- 证据：`arbiter/skill_handlers.go:52`、`arbiter/module.go:13-27`（没有注册 `HandLimit.HardCap`）、`blood_priestess/hand_limit.go:35`
- 现象：只是把基础值设成了 5，同生共死等修正照样叠加上去。
- 评审：

#### 仲裁者-3 末日审判：只能选敌方
- 见通用-7，证据是 `internal/data/characters.go:807`。
- 评审：

#### 仲裁者-4 判决天平：前端兜底规则写反
- 见通用-9。

没发现问题：仲裁法则、仪式中断、审判浪潮、末日审判的强制发动。

### 红莲骑士（crimson_knight）

#### 红莲骑士-1 热血沸腾：在额外行动之前就结算
- 见通用-3。
- 评审：

#### 红莲骑士-2 腥红十字：只能选敌方
- 见通用-7。测试 `OnlyTargetsEnemy` 固定了这个行为。
- 评审：

#### 红莲骑士-3 血腥祷言：给某人分 0 点时，报错看不出原因
- 前端体验｜低｜已确认
- 证据：`crimson_knight/choices.go:264`
- 现象：报错是"无效的治疗分配"。建议改成"每名队友至少 1 点"。
- 评审：

没发现问题：腥红圣约、腥红信仰、杀戮盛宴、戒骄戒躁。

### 兽灵武士（beast_samurai）

#### 兽灵武士-1 兽返：X≥2 时流程卡死，或少弃牌
- Bug｜高｜已确认（临时测试）。已核实 ChoiceSpecs 里只注册了一个多选
- 证据：`beast_samurai/module.go:35`、`beast_samurai/choices.go:270`
- 现象：X=2 时，自弃提示是 min=max=2。提交时报 `unregistered or no OnMultiSelect for "bs_beast_return_self_discard"`。如果只交 1 张，后端能接受，但直接进入对方弃牌，结果只弃了 1 张。
- 评审：

#### 兽灵武士-2 兽魂超出上限的部分，第一次读取就被截回 2
- Bug｜中｜已确认（临时测试）
- 证据：`internal/engine/player/shared_helpers.go:190-193`（`TokenValue` 会把截断后的值写回去）、`beast_samurai/timing_hooks.go:121`
- 现象：兽魂溢出到 3，回合结束扣 1 后变成 1，应该是 2。"无视上限"多出来的那点等于丢了。
- 评审：

#### 兽灵武士-3 居合形态：一回合扣两次兽魂
- 见通用-3。
- 评审：

#### 兽灵武士-4 逆反居合斩：发动时机被改到命中后
- 文案不符｜中｜已确认
- 证据：`internal/data/characters.go:2327`（`TimingAttackHit`）、`beast_samurai/skill_handlers.go:226`
- 现象：文案是"攻击……前①发动"，命中时②改为弃牌。代码改成命中后才询问，玩家可以先看到是否命中再做决定。
- 评审：

#### 兽灵武士-5 逆反居合斩：X 能不能为 0（文案歧义）
- 证据：`beast_samurai/choices.go:84`（现在允许 X=0）
- 评审：

#### 兽灵武士-6 兽魂警戒：谁来弃牌（文案歧义）
- 证据：`beast_samurai/skill_handlers.go:157`
- 现在固定由被横置的角色弃牌，另一条可选目标的流程 `bs_alert_target` 没用上。
- 评审：

#### 兽灵武士-7 兽返：自弃也被"展示"
- 见通用-8。
- 评审：

没发现问题：武者残心、一击无念、兽魂意念。

### 阴阳师（onmyoji）

#### 阴阳师-1 式神咒束：不扣战绩区资源
- 文案不符｜高｜已确认（临时测试）。已核实 `CanPayBindingCost` 只检查、不扣除
- 证据：`onmyoji/runtime.go:100`、`onmyoji/choices.go:483-527`
- 现象：完整走完代应战后，战绩区的宝石和水晶都没少，可以无限发动。测试 `RequiresGemAndCrystal` 只测了能否触发。
- 评审：

#### 阴阳师-2 生命结界分支①：X=3 时照样掉士气
- 文案不符｜高｜已确认
- 证据：`onmyoji/choices.go:603-611`（自伤的 PendingDamage 没带防士气标记）
- 评审：

#### 阴阳师-3 生命结界：结算后的流程走向
- Bug｜中｜推测
- 证据：`onmyoji/choices.go:615`、`onmyoji/choices.go:633-645`
- 现象：分支①结算后直接进入 TurnEnd，跳过了行动结束时点和 ActionQueue。分支②在队友没有手牌时什么也不推，阶段也不推进，流程可能停住。
- 评审：

#### 阴阳师-4 生命结界、式神咒束：分支选择不能取消
- 前端体验｜低｜推测
- 证据：`onmyoji/choices.go:21`
- 评审：

#### 阴阳师-5 阴阳转换 / 式神转换：顺序和是否可选（文案歧义）
- 证据：`onmyoji/runtime.go:59-85`
- 两处歧义：伤害 X 要不要把式神转换给的那 1 点鬼火算进去；式神转换能不能选择不发动（文案里的［强制］只修饰摸牌）。
- 评审：

#### 阴阳师-6 阴阳转换：普通应战也被强制套上
- Bug｜低｜推测
- 证据：`onmyoji/policy_hooks.go:28`
- 现象：用一张同系、同命格的牌普通应战时，伤害会被强制改成鬼火数，玩家没法选择不发动。
- 评审：

#### 阴阳师-7 黑暗祭礼：时机和目标
- 见通用-3、通用-7。
- 评审：

没发现问题：式神降临。

### 封印师（sealer）

#### 封印师-1 五系封印：一次展示多张同系牌会造成多次伤害
- Bug｜高｜已确认会入队 2 笔伤害，是否两笔都按 3 点结算没有完全确认
- 证据：`internal/engine/notify_prompt_runtime.go:151-154`、`sealer/skill_handlers.go:271`
- 现象：一次展示 2 张火系牌（比如弃牌展示），会产生 2 笔 3 点法术伤害，但场上只有 1 张封印。封印要等伤害结算后才移除，挡不住重复触发。
- 评审：Bug 成立（已裁定）。一次打出或展示多张同系牌只算触发一次，触发后封印就移除，不会多次触发。

#### 封印师-2 封印破碎：拿牌后不检查手牌上限
- Bug｜低｜推测
- 证据：`sealer/skill_handlers.go:147`（见通用-6）
- 评审：

没发现问题：法术激荡、五系束缚、封印牌的放置。

### 瘟疫法师（plague_mage）

#### 瘟疫法师-1 死亡之触：伤害摸牌被限制在手牌上限以内
- 文案不符｜高｜已核实代码行为，有没有官方 FAQ 依据需要人工确认
- 证据：`plague_mage/choices.go:313`（`CapDrawToHandLimit: true`）
- 现象：目标最多摸到手牌上限，所以这个技能基本不可能让对方士气下降。文案里没有这条限制。
- 评审：

#### 瘟疫法师-2 死亡之触：校验顺序错误
- Bug｜低｜已确认
- 证据：`plague_mage/choices.go:288-298`
- 现象：先把牌从手牌移走，再检查治疗够不够。不够时牌已经进了弃牌堆，技能却报错失败。
- 评审：

#### 瘟疫法师-3 死亡之触：同一组牌要选两次
- 前端体验｜中｜已确认
- 证据：`plague_mage/choices.go:58`、`web/src/components/PromptDialog.vue:611`
- 现象：第一步点一张手牌来声明系别，中间选 X，第三步再重新选同系牌。可以合并成一步："直接选 ≥2 张同系牌"。
- 评审：

#### 瘟疫法师-4 死亡之触：只能选敌方
- 见通用-7。
- 评审：

没发现问题：不朽、圣渎、瘟疫、剧毒新星。

---

## 四、文案歧义汇总（只需选一种理解）

| 编号 | 问题 | 当前实现 |
|---|---|---|
| 女武神-1 | 英灵召唤 +1 治疗给谁 | 给被攻击的敌人 |
| 通用-4 / 英灵人形-5 | "回合限定"按谁的回合算 | 自己的回合周期 |
| 通用-7 | "目标角色"是否包含队友和自己 | 多数只能选敌方 |
| 吟游诗人-3、神官-5 | 目标能不能选自己 | 不谐和弦可以，神圣领域不行 |
| 血色剑灵-4、吟游诗人-8、灵魂术士-5、血之巫女-3 | "队友"是否包含自己 | 不统一 |
| 圣女-2 | 展示牌算不算"使用"，能否触发冰霜祷言 | 算 |
| 月之女神-2 | 月渎追加伤害的目标 | 锁定为受伤的那名敌人 |
| 月之女神-3 | "视为攻击"能否触发美杜莎之眼 | 不能 |
| 月之女神-4 | 美杜莎之眼没手牌时还造不造成伤害 | 造成 |
| 圣弓-2 | 自动填充能不能不发动 | 不能 |
| 圣弓-6 | 圣屑飓暴：队友手牌必须够 X 张吗 | 必须 |
| 格斗家-2 | 百式幻龙拳的［持续］持续多久 | 只持续一个行动阶段 |
| 苍炎魔女-3 | 天火断空 +1 加在哪一段 | 两段都加 |
| 英灵人形-6 | 战纹碎击 +Y 加在哪里 | 另推一笔法术伤害 |
| 血色剑灵-5 | 血气屏障反伤对象 | 固定为伤害来源 |
| 精灵射手-2 | 动物伙伴能被哪些伤害触发 | 只有攻击伤害 |
| 精灵射手-3 | 爆牌弃牌和技能弃牌是否合并 | 合并 |
| 魔法少女-3 | 魔爆冲击：伤害和弃牌谁先 | 先弃牌 |
| 魔枪-4 | 漆黑之枪在什么时点判断手牌数 | 命中后 |
| 魔枪-6 | 暗之障壁能不能混弃 | 不能 |
| 冒险家-4 | 偷天换日：我方战绩区满了时 | 照常扣对方，溢出作废 |
| 吟游诗人-5 | 沉沦协奏曲的触发时点 | 回合结束时 |
| 蝶舞者-2 | 镜花水月的"他"是谁 | 原伤害的承受者 |
| 蝶舞者-3 | 茧溢出舍弃算不算"移除" | 不算 |
| 神官-6 | 神圣契约溢出的治疗怎么处理 | 直接丢失 |
| 兽灵武士-5 | 逆反居合斩 X 能不能为 0 | 能 |
| 兽灵武士-6 | 兽魂警戒由谁弃牌 | 被横置的角色 |
| 阴阳师-5 | 阴阳转换的 X 算不算式神转换给的鬼火；式神转换能否不发动 | 算；不能 |
| 圣枪骑士-3 | 惩戒：自己治疗满了算不算转移 | 目标照样 -1 |
| 各专属资源上限 | 斗气 8、剑气 5、剑魂 3、祈祷符文 3、新月和石化 10 | 文案没写，需要对照原版规则 |

---

## 五、测试缺口

- 英灵人形、神官两个包完全没有单测。
- 下面这些测试把可疑的行为固定成了预期。如果评审结论是按文案改，要一起修改：
  - `TestElementalistOffenseSkills_RejectAllyTargets`（元素师）
  - 腥红十字 `OnlyTargetsEnemy`（红莲骑士）
  - `MeteorBullet_RequiresOtherAlly`（圣弓）
  - `CannotUseWhenMoraleNotBehind`（圣弓）
  - `TestMoonGoddessBlasphemy_TargetLockedToDamagedEnemyAndSelfTurn`（月之女神）
  - `ExcludesConvertedAttacks`（月之女神）
  - `TestValkyrie_HeroicSummon_ExtraDiscardHealsCurrentCombatTarget`（女武神）
- 下面这些路径没有测试：
  - 威力赐福的响应
  - 不谐和弦的取消
  - 沉沦协奏曲在别人回合的第二次触发
  - 蝶舞者遇到非 `"magic"` 类型的伤害
  - 幻影星尘在有待弃牌时的续接
  - 兽返 X≥2
  - 式神咒束扣费
  - 前端点牌选独有技（`ActionPanel.spec.ts` 是直接写 store 的）

---

## 六、评审结论（2026-11-05）

| # | 裁定 | 影响条目 | 结果 |
|---|---|---|---|
| 1 | 威力赐福 +2 加在原有攻击伤害上 | 祈祷师-1 | 要修 |
| 2 | 一次多张同系牌只触发一次，触发后封印移除 | 封印师-1 | 要修 |
| 3 | 冒险者天堂：有能量时必须扣 1 能量 | 冒险家-1 | 要修 |
| 4 | 回合限定：每个玩家的回合各限一次（重置应在每个回合开始时，而不是角色自己的回合） | 通用-4、月之女神-1、英灵人形-5、吟游诗人-4 | 要修（跨回合共用框架） |
| 5 | 圣疗分配上限 3，可以小于 3 | 圣女-1 | 改后端校验 |
| 6 | 元素师独有技都是目标角色 | 元素师-2、通用-7 | 元素师要修；其他"目标角色"待确认 |
| 7 | 双重回响不限于攻击/法术伤害 | 英灵人形-4 | 删掉校验；现在的行为是对的 |
| 8 | 战纹碎击 +Y 按技能文案 | 英灵人形-6 | 保持现状 |
| 9 | 只要士气下降就触发灵魂吞噬 | 灵魂术士-1 | 要修 |
| 10 | 天枪是无法应战，视为暗灭可以接受 | 圣枪骑士-1 | 保持现状 |
| 11 | 自动填充可以不发动 | 圣弓-2 | 要修 |
| 12 | 沉沦协奏曲：弹出询问但拒绝了，算本轮已经发动过 | 吟游诗人-4 | 保持现状，不用改 |
| 13 | 血气屏障：反伤由玩家任选一名对手 | 血色剑灵-5 | 要修 |
| 14 | 动物伙伴：自己回合造成的任何伤害都能触发 | 精灵射手-2 | 要修 |
| 15 | 动物伙伴：先结算爆牌，再弃技能的 1 张 | 精灵射手-3 | 要修 |
| 16 | 专属资源上限：斗气 8、剑气 5、剑魂 3、祈祷符文 3 | 格斗家-4、剑帝-2、祈祷师-3 | 保持现状 |

---

## 七、官网对照结论

来源：星杯传说百科（xingbeichuanshuobaike.webnode.tw）上的各角色"角色技能详解"、"常见问题"和"专有名词"。只列官网有明确说法的条目，没写到的条目仍按第四部分的歧义处理。

### 官网支持"是 bug，要修"

| 条目 | 官网原意 |
|---|---|
| 通用-7 | FAQ 法术 Q1：说明里写"目标角色"的，可以指定包括自己在内的任何角色；写"目标队友"的只能指定队友，不能指定自己；写"目标对手"的只能指定对手 |
| 天使-1 | 天使之歌和风之洁净的区别，就是可以**先**移除自己面前的负面基础效果 |
| 天使-2 | 天使羁绊只在两种情况下触发：风之洁净或天使之歌移除基础效果、使用圣盾。天使自己用掉而移除的基础效果不算 |
| 天使-4 | 风之洁净：场上没有基础效果时不能发动 |
| 冒险家-2 | 冒险家替队友提炼，只看队友的能量有没有到上限 |
| 暗杀者-1 | 反噬：用治疗完全抵掉攻击伤害，或者伤害全部被灵魂链接转走时，不会触发 |
| 神官-1 | 水之神力：只剩 1 张水系牌时，可以弃这一张发动 |
| 神官-3 | 水之神力：给队友牌也可能让队友爆牌 |
| 魔弓-2 | 雷光散射：额外伤害一次结算。移除 X 个雷系充能后，对指定的对手造成 X 点，其他对手各 1 点 |
| 魔枪-3 | 暗之解放：+1 只对进入幻影形态那个回合的第一次主动攻击有效 |
| 魔剑士-1 | 暗影抗拒禁的是中毒、虚弱、圣盾、魔弹、圣光；在自己回合外仍然可以用魔弹和圣光。和形态无关 |
| 仲裁者-2 | 仲裁仪式：手牌上限恒定为 5，不受其他改上限的技能影响 |
| 血之巫女-1 | 流血：手牌一少于 3 张，就**立即**重置并脱离流血形态 |
| 圣弓-4 | 圣煌辉光炮：差值为 0 时，X 可以取 0 |
| 圣女-2 | 冰霜祷言：只有使用水系牌（包括水系法术）才触发，单纯弃牌不触发 |
| 阴阳师-1 | 式神咒束的代价是 1 宝石 + 1 星石，星石用宝石或水晶都可以 |
| 格斗家-2 | 百式幻龙拳：被虚弱跳过行动阶段时不会取消；已经在百式状态下再次启动，只能指定上一个目标。这说明状态会**跨回合持续**，现在每回合结束就退场是错的 |
| 通用-8 | 五系封印：展示蓋牌、爆牌弃的牌都不触发封印。只有打出和展示才检查 |

### 官网支持"现状正确，不用改"

| 条目 | 官网原意 |
|---|---|
| 吟游诗人-3 | 不谐和弦可以对自己使用，结算方式是先弃 X-1 张，再弃 X-1 张 |
| 吟游诗人-5 | 沉沦协奏曲在回合结束时触发，前提是相关的法术伤害已经完全结算完 |
| 冒险家-4 | 己方星石满了也能用偷天换日，视为直接弃掉对方战绩区 1 个宝石 |
| 冒险家-3（部分） | 己方战绩区没有宝石或水晶时，依然可以选"水晶全部替换为宝石"这一项。所以"能选"是对的，白扣资源的问题要另看 |
| 蝶舞者-2 | 镜花水月拆分前后目标一样，只改伤害来源 |
| 蝶舞者-3 | 茧因为溢出被丢掉算"舍弃"，不算"移除"，所以不能触发镜花水月和凋零 |
| 月之女神-3 | 欺诈、圣屑飓暴都不触发美杜莎之眼 |
| 魔枪-4 | 漆黑之枪：命中后的手牌数不是 1 或 2，就不能发动 |
| 魔枪-6 | 暗之障壁：法术牌和雷系牌不能混着弃 |
| 灵魂术士-5 | 2v2 禁止使用灵魂链接 |
| 血之巫女-3 | 同生共死可以用在自己身上，效果不会翻倍 |
| 圣枪骑士-3 | 惩戒：自己治疗满了也能发动，只是自己不再加治疗 |
| 兽灵武士-5 | 逆反居合斩的 X 可以取 0（兽返的 X 不行） |
| 勇者-2 | FAQ 行动 Q3：勇者横置后还能再发动禁断之力，并保持横置 |
| 神官-6 | 神圣契约：目标治疗最多到 4，多出来的加不上 |
| 剑帝-2、格斗家-4 等上限 | 剑魂上限 3、妖力上限 2、充能上限 8、茧上限 8 |

### 官网补充的细节（修的时候要注意）

- 苍炎魔女-3 天火断空：对对手发动时，每段伤害都是 4 点；对我方角色发动时每段 3 点。也就是说只在目标是对手时才 +1，两段都加。现在要检查目标是队友时是否错加了。
- 阴阳师-5：两个技能都发动时，先结算阴阳转换的 +1 鬼火，再结算式神转换的"摸 1 张 +1 鬼火"，最后脱离形态。
- 祈祷师·黑暗诅咒、红莲骑士·杀戮盛宴、血之悲鸣：都是先结算对目标的伤害，再结算对自己的伤害。贤者-1（法术反弹自伤先结算）可以按这个口径对照。
- 魔弓·魔贯冲击：命中时最多额外移除 1 个火系充能。
- 风之剑圣：先做了法术行动或特殊行动，就不能再发动连续技、疾风技、剑影。
