***

# 《星杯传说》核心引擎枚举字典 (Enums Dictionary)

**文档说明**：本文档定义了星杯对战引擎（Starcup Engine）中所有的核心状态、类型与阶段枚举。它是前端UI渲染、后端状态机流转、以及事件总线派发的基石。

---

## 1. 基础战斗属性 (Attributes)

游戏中用于判定资源、卡牌归属及角色流派的基础枚举。

### 1.1 阵营类型 (Faction)
| 枚举值 | 标识名    | 描述    |
|:----|:-------|:------|
| `0` | `None` | 未选择阵营 |
| `1` | `Red`  | 红方阵营  |
| `2` | `Blue` | 蓝方阵营  |

### 1.2 星石类型 (Star Stone)
| 枚举值 | 标识名       | 描述                                  |
|:----|:----------|:------------------------------------|
| `0` | `None`    | 无星石/无消耗                             |
| `1` | `Gem`     | 宝石（主动攻击命中后获得，红色星石）                  |
| `2` | `Crystal` | 水晶（应战命中后获得，蓝色星石）                    |
| `3` | `Any`     | 能量/任意星石（通配符：作为技能消耗或发动条件时，指代不限颜色的星石） |

### 1.3 元素属性 (Element)
| 枚举值 | 标识名       | 描述  |
|:----|:----------|:----|
| `0` | `None`    | 无属性 |
| `1` | `Earth`   | 地系  |
| `2` | `Water`   | 水系  |
| `3` | `Fire`    | 火系  |
| `4` | `Wind`    | 风系  |
| `5` | `Thunder` | 雷系  |
| `6` | `Light`   | 光系  |
| `7` | `Dark`    | 暗系  |

### 1.4 命格类型 (Destiny)
| 枚举值 | 标识名     | 描述  |
|:----|:--------|:----|
| `0` | `None`  | 无命格 |
| `1` | `Huan`  | 幻   |
| `2` | `Yong`  | 咏   |
| `3` | `Xue`   | 血   |
| `4` | `Ji`    | 技   |
| `5` | `Sheng` | 圣   |

### 1.5 角色姿态 (Character Orientation)
| 枚举值 | 标识名      | 描述                           |
|:----|:---------|:-----------------------------|
| `0` | `Normal` | 转正（正常状态）                     |
| `1` | `Tapped` | 横置（通常伴随特定形态的开启或某些技能发动后的惩罚状态） |
角色横置后会进入横置形态，部分有名称、部分无。用**可选字符串**表示，由角色/技能配置定义，无需全局枚举。

### 1.7 专属指示物类型 (Token Type)
用于统一管理各角色特有的专属资源标记。每个角色实体需配置其对应的指示物上限。
| 枚举值 | 标识名 | 描述（对应角色举例） | 上限 |
|:---|:---|:---|:---|
| `0` | `None` | 无 |0|
| `1` | `BloodMark` |  血印（ 红莲骑士） | 2|
| `2` | `Rune` | 祈祷符文（祈祷师） | 3|
| `3` | `GhostFire` | 鬼火（阴阳师） | 3|
| `4` | `Rebirth` | 重生（苍炎魔女） | 4|
| `5` | `EnergyCharge` | 充能（魔弓） |8|
| `6` | `Inspiration` | 灵感（吟游诗人） |3|
| `7` | `Rage` | 怒气（勇者） |4 |
| `8` | `Intellect` | 知性（勇者） |4 |
| `9` | `BattleQi` | 斗气（格斗家） |6|
| `10` | `Faith` | 信仰（圣弓 / 原初之弓） |10|
| `11` | `SwordQi` | 剑气（剑帝） |5|
| `12` | `Zanshin` | 残心（兽灵武士） |4|
| `13` | `BeastSoul` | 兽魂（兽灵武士） |2|
| `14` | `SoulYellow` | 黄色灵魂（灵魂术士） |6|
| `15` | `SoulBlue` | 蓝色灵魂（灵魂术士） |6|
| `16` | `Petrifaction`| 石化（月之女神） |3|
| `17` | `Crescent` | 新月（月之女神） |无上限|
| `18` | `Pupa` | 蛹（蝶舞者） |无上限|
| `19` | `HolyMark` | 圣印（圣殿骑士） |2|
| `20` | `Judgment` | 裁决 / 审判（圣庭检察士 / 仲裁者） |
| `21` | `Law` | 律法（星坠女巫） | 3|
| `22` | `SpiritSurge` | 灵涌（灵熙之潮） | 4|
| `23` | `Silk` | 丝（女仆长） |4|
| `24` | `Sacrifice` | 祭（结界师） |3|
| `25` | `SecretArt` | 秘术（神秘学者） |4|
| `26` | `Hostility` | 戾气（染污者） |2|
| `27` | `SilverBullet`| 银制子弹（红衣主教 / 铸律者） |
| `28` | `HolyRelic` | 圣遗物（红衣主教 / 铸律者） |
| `29` | `History` | 史料（记录者） |
| `30` | `Piety` | 虔诚（传教士） |
| `31` | `Atom` | 元子（见习制片） |
| `32` | `Element` | 元素（元素师，用于元素点燃等技能消耗） |3|
| `33` | `WarRune` | 战纹（英灵人形） | 3|
| `34` | `MagicRune` | 魔纹（英灵人形） | 3|
| `35` | `BloodMark` | 鲜血（血色剑灵） | 3|

---

## 2. 卡牌与技能定义 (Cards & Skills)

决定技能的触发方式、UI展示层级以及消耗规则。

### 2.1 卡牌类型 (Card Type)
| 枚举值 | 标识名      | 描述                 |
|:----|:---------|:-------------------|
| `0` | `Attack` | 攻击牌（如：水涟斩、暗灭）      |
| `1` | `Magic`  | 法术牌（如：圣光、魔弹、中毒、圣盾） |

### 2.2 技能类别 (Skill Category) - 标明来源与成本
| 枚举值 | 标识名         | 描述                        |
|:----|:------------|:--------------------------|
| `0` | `Normal`    | 一般                        |
| `1` | `Unique`    | 独有技（必须抽到并打出对应的角色专属卡牌才能发动） |
| `2` | `Exclusive` | 专属技（角色专属技能，需搭配专属技能卡发动）    |
| `3` | `Ultimate`  | 大招                        |

### 2.3 技能类型 (Skill Type) - 标明发动机制
| 枚举值 | 标识名        | 描述                                      |
|:----|:-----------|:----------------------------------------|
| `0` | `Passive`  | 被动技（满足特定条件自动触发，无需玩家主动选择）                |
| `1` | `Startup`  | 启动技（仅能在“行动阶段开始时”主动宣告，执行启动技能后后续无法执行特殊行动） |
| `2` | `Magic`    | 法术技（仅能在“行动阶段中”主动宣告，视为执行法术行动）            |
| `3` | `Response` | 响应技（满足触发条件时，弹出窗口供玩家选择是否响应）              |

### 2.4 卡牌场上标记 (Card Field Mark)
当一张基础卡牌被扣置或明置于场上，作为角色的特殊资源时，其原本的 CardType 保持不变，但系统会为其附加 FieldMark 状态，用于技能判定。
| 枚举值 | 标识名 | 描述（对应角色举例） | 上限 |
|:---|:---|:---| :---|
| `0` | `None` | 无标记（正常手牌或弃牌） | 0 |
| `1` | `Cover` | 普通盖牌（通用资源） | 0 |
| `2` | `Shadow` | 影（魔剑士 / 女仆长） |3|
| `3` | `Cocoon` | 茧（蝶舞者） |8|
| `4` | `SwordSoul` | 剑魂 / 天使之魂 / 恶魔之魂（剑帝） | 3|
| `5` | `Bullet` | 子弹（游击士） |
| `6` | `Blessing` | 祝福（精灵射手） |3|
| `7` | `DemonForce` | 妖力（灵符师 / 咒符师） |2|
| `8` | `Barrier` | 结界（结界师） |
| `9` | `WordSpirit` | 言灵（神秘学者） |
| `10` | `Relic` | 遗迹（记录者） |
| `11` | `Prophecy` | 预言（异教徒） |
| `12` | `Tricky` | 捣蛋陷阱（捣蛋萝莉） |
| `13` | `RuneCard` | 卢恩（星坠女巫） |
| `14` | `MagicBottle`| 魔力瓶（猎巫人） |
| `15` | `WindHole` | 风穴（女仆长，放置于对手面前） |
| `16` | `AbsoluteBoundary` | 绝界（结界师，转移或放置于目标角色前） |
| `17` | `EternalMovement` | 永恒乐章（吟游诗人，放置于队友面前） |1|
| `18` | `SharedFate` | 同生共死（血之巫女，放置于目标角色面前） |1|
| `19` | `InquisitionCourt` | 异端裁决所（圣庭检察士，游戏初始拥有，可存储治疗） |
| `20` | `CrownOrb` | 王权宝珠（铸律者） |
| `21` | `Chronicle` | 史书（记录者，史料达上限时加入手牌） |
| `22` | `HolyGloryCannon` | 圣煌辉光炮（圣弓/原初之弓，游戏初始拥有） |
| `23` | `PowerBlessing` | 威力赐福（祈祷师，放置于队友面前，攻击命中可移除以发动伤害+2） |
| `24` | `SwiftBlessing` | 迅捷赐福（祈祷师，放置于队友面前，行动结束时移除以发动+1攻击行动） |
| `25` | `SoulLink` | 灵魂链接（灵魂术士，放置于队友面前，伤害转移机制） |
| `26` | `DarkMoon` | 暗月（月之女神，因爆牌而弃置的牌面朝下放置于角色旁） |无上限

### 2.5 目标选择规则 (Target Select Type)
决定技能或卡牌打出时，前端 UI 允许玩家选取的目标范围。
| 枚举值 | 标识名 | 描述 |
| :--- | :--- | :--- |
| `0` | `None` | 无需目标（如：合成、购买、提炼、或者作用于全局/自身的技能） |
| `1` | `Self` | 仅限自己 |
| `2` | `Teammate` | 己方阵营（己方任意角色） |
| `3` | `TeamOther` | 目标队友（除了自己以外的己方任意角色） |
| `4` | `Enemy` | 任意敌方对手 |
| `5` | `EnemyOther` | 敌方除某一目标外任意对手 |
| `6` | `Any` | 场上任意一名角色（敌我皆可） |

### 2.6 基础效果与场地状态 (Status Effects) type StatusEffect
放置于角色面前或公共区域，产生持续性限制、增益或判定劫持的对象。
| 枚举值 | 标识名 | 描述 |
|:---|:---|:---|
| `0` | `None` | 无 |
| `1` | `Poison` | 中毒（行动阶段开始前受法术伤害；同一名角色面前最多存在1个） |
| `2` | `Weakness` | 虚弱（跳过行动或摸3张牌） |
| `3` | `HolyShield` | 圣盾（抵挡一次攻击或魔弹） |
| `4` | `FiveElementsBind` | 五系束缚（封印师专属，跳过下个行动阶段或摸牌取消） |
| `5` | `ElementalSeal` | 元素封印（水/火/地/风/雷之封印，打出/展示对应系别牌时触发3点法术伤害） |
| `6` | `Taunt` | 挑衅（勇者专属：持有者下个行动阶段必须主动攻击状态来源，否则跳过该行动阶段） |


> 注：像 `FiveElementsBind`、`ElementalSeal` 这类“放置后延后触发”的状态，技能节点只负责 `EffectPlaceStatus`；后续结算时点和分支由 `StatusResolveConfig` 驱动。
> 注：`StatusResolveConfig.CanDecline=false` 表示触发条件满足后必须结算（如圣盾），不能让持有者放弃。
> 注：状态实体需记录 `StatusMeta.SourceUserID`（状态施加者），用于“挑衅”这类“必须攻击来源角色”的判定；元素封印继续使用 `StatusMeta.BoundElement`。
```go
type StatusStackingRule int
const (
    StackingNone     StatusStackingRule = 0
    
    StackingUnique   StatusStackingRule = 1 // 唯一标记：同一玩家面前同类状态只能存在一个（如：中毒、虚弱、圣盾）
	
    StackingMultiple StatusStackingRule = 2 // 可以叠加多个
)
```

### 2.6.1 状态延后结算枚举 (Status Resolve)
| 枚举值 | 标识名                                          | 描述                         |
|:----|:---------------------------------------------|:---------------------------|
| `0` | `ResolveNone`                                | 无延后结算                      |
| `1` | `ResolveOnHolderNextBeforeAction`            | 在状态持有者“下个行动阶段开始前”触发（如五系束缚） |
| `2` | `ResolveOnHolderCardPlayedOrRevealed`        | 在状态持有者打出/展示牌时触发（如元素封印）     |
| `3` | `ResolveOnHolderIncomingAttackOrMagicBullet` | 在状态持有者遭受攻击或魔弹时触发（如圣盾）      |

| 枚举值 | 标识名                 | 描述           |
|:----|:--------------------|:-------------|
| `0` | `ResolveModeAuto`   | 自动结算（无玩家选择）  |
| `1` | `ResolveModeChoice` | 结算时弹窗让玩家选择分支 |

| 枚举值 | 标识名                      | 描述       |
|:----|:-------------------------|:---------|
| `0` | `StatusSystemOpNone`     | 无系统动作    |
| `1` | `SkipCurrentActionPhase` | 跳过当前行动阶段 |
| `2` | `SystemOp_StartMagicBulletChain` | 启动魔弹传递链路结算 |

| 枚举值 | 标识名                                   | 描述 |
|:----|:--------------------------------------|:---|
| `0` | `MatchPlayedCardElementAny`           | 不校验打出/展示牌系别 |
| `1` | `MatchPlayedCardElementFixed`         | 与固定系别比较（由 `RequiredPlayedCardElement` 指定） |
| `2` | `MatchPlayedCardElementStatusBoundElement` | 与状态实例绑定系别比较（读取 `StatusMeta.BoundElement`） |

> 轻量动作锁（仅用于【挑衅】）：
> 当 `StatusResolveConfig.EnforceNextActionMustActiveAttackSource=true` 且 `ResolveTiming=TimingBeforeActionExecute` 时，
> 状态持有者在其下个行动阶段提交的首个行动必须满足：
> 1. 行动类型为主动攻击；
> 2. 攻击目标为该状态的 `StatusMeta.SourceUserID`。
> 否则系统直接执行 `SkipCurrentActionPhase`。无论满足与否，本次判定后都移除该状态牌。

### 2.7 战斗与结算劫持标记 (Combat Intercept Tag)
用于在流水线中作为 Flag 标记，一旦某技能附加了这些 Tag，底层状态机在运转时会自动应用相应的规则，无需为每个技能写死逻辑。

| 枚举值 | 标识名 (Tag)          | 描述        | 适用阶段/作用                                    |
|:----|:-------------------|:----------|:-------------------------------------------|
| `0` | `None`             | 无特殊标记     | -                                          |
| `1` | `Unrespondable`    | 无法应战      | 在 ②命中判定阶段 禁用目标打出应战牌的 UI。                   |
| `2` | `IgnoreHolyShield` | 无视圣盾      | 在 ②命中判定阶段 判定圣盾无效，且不消耗圣盾。                   |
| `3` | `ForceHit`         | 强制命中      | 跳过 ②命中判定阶段 的所有防御判定（包括应战和圣盾）。               |
| `4` | `NoMoraleDrop`     | 不造成士气下降   | 在 ⑥承受伤害阶段，若摸牌导致爆牌，忽略对士气的扣除。                |
| `5` | `ChangeElement`    | 改变攻击系别    | 在 ①发动或 ②应战时 动态覆写 `Combat.Element`。         |
| `6` | `UnhealableDamage` | 伤害无法用治疗抵御 | 在 ④治疗响应阶段 禁用遭受伤害者的治疗抵御操作。                  |
| `7` | `DamageLimitMax`   | 伤害上限限制    | 结算时强制进行 `Math.min(Damage, LimitValue)` 修饰。 |
| `8` | `ReverseDeliver`   | 逆向传递魔弹    | 在 ②判定魔弹时，改变传递方向指针。                         |
| `9` | `IgnoreTargetHoly` | 无视目标圣光抵挡  | 在 ②命中判定阶段，禁用目标打出圣光抵挡。                      |

### 2.8 卡牌可见性枚举(Card Visibility Type)
用于在流水线中作为 Flag 标记，一旦某技能附加了这些 Tag，底层状态机在运转时会自动应用相应的规则，无需为每个技能写死逻辑。

| 枚举值 | 标识名 (Tag)          | 描述    |
|:----|:-------------------|:------|
| `0` | `VisibilityHidden` | 暗置不可见 |
| `1` | `VisibilityPublic` | 可见    |

### 2.9 技能打牌载体枚举 (Card Play Cost Type)
用于区分“费用为 None 但发动是否必须打牌”。

| 枚举值 | 标识名 | 描述 |
|:----|:---|:---|
| `0` | `CardPlayNotRequired` | 发动不需要打出卡牌（如被动触发、纯按钮发动） |
| `1` | `CardPlayRequired` | 发动必须打出与该技能映射的卡牌（如基础法术牌、独有牌） |

### 2.10 玩家属性枚举 (Player Attribute Type)
用于 `RuleModifier` 的属性域（Attribute Domain），统一表达可被修饰的角色面板属性。

| 枚举值 | 标识名 | 描述 |
|:----|:---|:---|
| `0` | `PlayerAttributeMaxHand` | 手牌上限 |
| `1` | `PlayerAttributeMaxEnergy` | 能量上限 |
| `2` | `PlayerAttributeMaxHeal` | 治疗上限 |

### 2.11 属性修饰操作枚举 (Attribute Modify Op Type)
用于 `RuleModifier` 的属性域负载，定义如何作用到目标属性。

| 枚举值 | 标识名 | 描述 |
|:----|:---|:---|
| `0` | `AttributeModifyAdd` | 在当前值基础上叠加 `Value` |
| `1` | `AttributeModifySet` | 直接设为 `Value` |

### 2.11.1 属性值来源模式枚举 (Rule Attr Value Source Mode)
用于 `RuleModifier` 的属性域负载，声明属性修饰值来自固定值、表达式或 Token 联动。

| 枚举值 | 标识名 | 描述 |
|:----|:---|:---|
| `0` | `RuleAttrValueFromFixed` | 使用 `RuleAttrPayload.Value` 固定值 |
| `1` | `RuleAttrValueFromExpression` | 使用 `RuleAttrPayload.ValueExpression` 动态表达式 |
| `2` | `RuleAttrValueFromTokenLinear` | 使用 `RuleAttrPayload.TokenLink`（按 Token 数量线性计算） |

### 2.11.2 属性值 Token 读取域枚举 (Rule Attr Token Owner Scope)
用于 `RuleAttrPayload.TokenLink`，声明“按谁的 Token 数量”进行联动计算。

| 枚举值 | 标识名 | 描述 |
|:----|:---|:---|
| `0` | `RuleAttrTokenOwnerTarget` | 按规则作用目标（Target）的 Token 计算 |
| `1` | `RuleAttrTokenOwnerSource` | 按规则施加来源（Source）的 Token 计算 |

### 2.12 治疗应用策略枚举 (Heal Apply Mode)
用于 `RuleModifier` 的治疗策略域，控制治疗结算是否受当前治疗上限约束。

| 枚举值 | 标识名 | 描述 |
|:----|:---|:---|
| `0` | `HealApplyRespectMax` | 按目标当前有效治疗上限结算（默认） |
| `1` | `HealApplyIgnoreMax` | 忽略目标当前有效治疗上限 |

### 2.13 规则修饰器域枚举 (Rule Modifier Domain)
用于声明某条 `RuleModifier` 属于哪类规则空间。

| 枚举值 | 标识名 | 描述 |
|:----|:---|:---|
| `0` | `RuleModifierDomainAttribute` | 属性域（如 `MaxHeal +1`） |
| `1` | `RuleModifierDomainHealPolicy` | 治疗策略域（如“无视上限且封顶5”） |
| `2` | `RuleModifierDomainSkillGate` | 技能可用性域（如“禁用指定技能”） |
| `3` | `RuleModifierDomainCardSource` | 卡牌来源域（如“指定场标可视为手牌来源”） |
| `4` | `RuleModifierDomainTokenPolicy` | 指示物策略域（如“无视鲜血上限但绝对封顶4”） |
| `5` | `RuleModifierDomainHealResistPolicy` | 治疗抵伤策略域（如“单次受伤窗口最多用1点治疗抵伤”） |
| `6` | `RuleModifierDomainMoralePolicy` | 士气策略域（如“对方士气最低为1直到下回合开始前”） |

### 2.13.1 士气策略作用域枚举 (Morale Policy Apply Scope)
用于 `RuleMoralePolicyPayload` 声明“这条士气策略作用于哪一方阵营”。

| 枚举值 | 标识名 | 描述 |
|:----|:---|:---|
| `0` | `MoralePolicyApplySourceTeam` | 作用于规则施加者所属阵营 |
| `1` | `MoralePolicyApplyEnemyTeam` | 作用于规则施加者的敌方阵营 |
| `2` | `MoralePolicyApplyTargetTeam` | 作用于规则作用目标（Target）所属阵营 |
| `3` | `MoralePolicyApplyAllTeams` | 同时作用于双方阵营 |

### 2.14 规则叠加策略枚举 (Rule Modifier Stack Policy)
用于定义同类规则命中时是并存、刷新还是替换。

| 枚举值 | 标识名 | 描述 |
|:----|:---|:---|
| `0` | `RuleModifierStackAppend` | 直接叠加并存 |
| `1` | `RuleModifierStackRefreshByModifierID` | 同 `ModifierID` 命中时刷新时长 |
| `2` | `RuleModifierStackReplaceByDomainPriority` | 同域冲突时保留更高优先级规则 |

### 2.15 规则持续时长枚举 (Rule Modifier Lifetime Type)
用于 `EffectApplyRuleModifier` 指定实例何时自动失效。

| 枚举值 | 标识名 | 描述 |
|:----|:---|:---|
| `0` | `RuleLifeThisEffectChain` | 当前技能执行链结束后失效 |
| `1` | `RuleLifeUntilTurnEnd` | 到当前回合结束 |
| `2` | `RuleLifeUntilSourceNextTurnStart` | 到施加者下回合开始 |
| `3` | `RuleLifeUntilSourceNextTurnEnd` | 到施加者下回合结束 |
| `4` | `RuleLifePermanent` | 常驻，需显式移除 |
| `5` | `RuleLifeUntilCombatEnd` | 到当前战斗结算链结束（`CombatDraw` 收尾后） |

### 2.16 规则移除模式枚举 (Rule Modifier Remove Mode)
用于 `EffectRemoveRuleModifier` 指定移除筛选方式。

| 枚举值 | 标识名 | 描述 |
|:----|:---|:---|
| `0` | `RuleRemoveByModifierID` | 按模板 ID 移除 |
| `1` | `RuleRemoveByDomain` | 按规则域移除 |
| `2` | `RuleRemoveBySourceSkill` | 按来源技能 ID 移除 |
| `3` | `RuleRemoveAll` | 移除目标全部规则实例 |

### 2.17 技能门禁模式枚举 (Skill Gate Mode)
用于 `RuleModifier` 的技能可用性域负载。

| 枚举值 | 标识名 | 描述 |
|:----|:---|:---|
| `0` | `SkillGateDisallowList` | 禁用 `SkillIDs` 列表中的技能 |

### 2.18 场标手牌投影模式枚举 (Card Source Projection Mode)
用于 `RuleModifier` 的卡牌来源域负载。

| 枚举值 | 标识名 | 描述 |
|:----|:---|:---|
| `0` | `CardSourceProjectionAsHand` | 将目标面前指定 `FieldMark` 的场上牌投影为手牌候选来源（可用于打出/展示/弃置） |

### 2.19 指示物应用策略枚举 (Token Apply Mode)
用于 `RuleModifier` 的指示物策略域，控制“增加指示物”时是否受当前指示物上限约束。

| 枚举值 | 标识名 | 描述 |
|:----|:---|:---|
| `0` | `TokenApplyRespectMax` | 按目标当前有效指示物上限结算（默认） |
| `1` | `TokenApplyIgnoreMax` | 忽略目标当前有效指示物上限 |

---

## 3. 联机大厅与网络状态 (Multiplayer & Network)

用于处理多玩家的房间管理、断线重连及AI托管。

### 3.1 房间状态 (Room State)
| 枚举值 | 标识名        | 描述               |
|:----|:-----------|:-----------------|
| `0` | `Waiting`  | 等待中（玩家进出房间，准备阶段） |
| `1` | `Drafting` | 选角中（BP阶段）        |
| `2` | `Playing`  | 游戏中（对局状态机正式运转）   |
| `3` | `Finished` | 已结束（展示战报与结算面板）   |

### 3.2 玩家网络状态 (Player State)
| 枚举值 | 标识名        | 描述                        |
|:----|:-----------|:--------------------------|
| `0` | `NotReady` | 未准备                       |
| `1` | `Ready`    | 已准备                       |
| `2` | `Playing`  | 游戏中（客户端WebSocket正常连接）     |
| `3` | `Offline`  | 已断线（挂起等待重连，可触发系统AI托管超时回合） |
| `4` | `Spectate` | 观战中                       |

---

## 4. 核心对局状态机 (Game Phases)

引擎的核心齿轮。技能发动的合法性校验及事件总线的派发均严格依赖此时间轴。

### 4.1 全局初试阶段
* `GameInit`：游戏初试时（初始化牌库、发放初始手牌、分配角色指示物）

### 4.2 玩家主回合 (8 阶段)
| 阶段标识名             | 中文名       | 阶段核心职能描述                             |
|:------------------|:----------|:-------------------------------------|
| `TurnBeforeStart` | 回合开始前     | 无实际意义，仅作为极其罕见的技能触发点。                 |
| `TurnStart`       | 回合开始时     | 宣言回合开始，触发对应的被动技能。                    |
| `BeforeAction`    | 行动阶段开始前   | **核心控制点**：强制结算面前的【中毒】伤害或【虚弱】跳过/摸牌判定。 |
| `ActionStart`     | 行动阶段开始时   | **核心控制点**：玩家发动【启动技】的唯一合法窗口。          |
| `ActionExecution` | 行动阶段中     | 玩家选择执行三大行动之一（攻击、法术、特殊）。              |
| `ActionEnd`       | 行动结束时     | 行动结算完毕后的收尾时点。                        |
| `ExtraAction`     | 行动结束后追加行动 | 若本回合存在额外行动，按照额外行动类型继续执行，也可以选择跳过。     |
| `TurnEnd`         | 回合结束时     | 宣言回合结束，清理所有临时状态，移交当前回合归属。            |

### 4.3 战斗/法术结算 (6 阶段)
当玩家在行动阶段发起攻击或法术时，触发此微型状态机闭环：
| 阶段标识名 | 中文名 | 阶段核心职能描述 |
| :--- | :--- | :--- |
| `CombatDeclare` | ① 发动阶段 | 攻击/法术宣告，触发“发动时”被动。 |
| `CombatHitCheck` | ② 命中判定阶段 | 拦截点：等待【圣盾】【圣光】抵挡，或【应战】及响应技的打断。 |
| `CombatCalcDamage` | ③ 计算伤害阶段 | 增减伤结算点：处理狂化、撕裂等数值变化被动。 |
| `CombatHeal` | ④ 治疗响应阶段 | 询问遭受伤害者是否消耗【治疗】抵挡伤害。 |
| `CombatApply` | ⑤ 实际产生伤害阶段 | 真实伤害落地，为攻击方结算阵营星石收入。 |
| `CombatDraw` | ⑥ 实际承受伤害阶段 | 承受伤害方摸牌，并执行爆牌检测及扣除士气逻辑。 |

### 4.4 事件触发时机钩子 (Trigger Timing)
注意：Phase 是系统当前的“状态”，而 Timing 是系统状态发生改变或动作发生时向外广播的“瞬间事件”。技能通过监听这些事件来触发。
| 标识名 | 触发场景描述 | 逻辑与复用说明 |
|:---|:---|:---|
| **【主动与主回合钩子】** | | |
| `TimingActive` | 玩家主动发动 | 普/独/大招的默认点击触发。 |
| `TimingStartup` | 玩家主动发动（启动技专有） | 仅在`ActionStart`阶段合法，占用启动名额。 |
| `TimingOnTurnStart` | 玩家的回合开始时触发 | 对应各种回合初的被动/状态转换。 |
| `TimingOnBeforeAction` | 回合玩家进入行动阶段开始前时触发 | 中毒、虚弱等状态结算。配合 `RequireHolderIsTurnPlayer` 筛出持有者=回合玩家的状态。 |
| **【动作与结算劫持钩子】** | | |
| `TimingBeforeActionExecute`| 系统尝试执行某种行动前 | **高复用拦截点**：系统传入 `ActionType`，用于劫持/替换默认购买、提炼等规则。 |
| `TimingOnActionEnd` | 某项行动彻底结算完毕时 | **高复用结算点**：系统传入 `ActionType`，涵盖攻击、法术、特殊行动结束。 |
| `TimingOnSkillExecuted` | 某一特定技能完整执行完毕时 | 系统传入 `SkillID`。如监听真言术执行完毕。 |
| **【战斗时间轴钩子】** | | |
| `TimingOnAttackDeclared`| ① 任意攻击宣告发动时 | |
| `TimingOnMagicDeclared` | ① 任意法术宣告发动时 | |
| `TimingOnHitCheck` | ② 命中判定时 | 拦截点：发效应战、圣盾、圣光、仪式中断。 |
| `TimingOnDamageCalculated`| ③ 伤害计算完毕时 | 增减伤结算点：撕裂、剑魂等数值修饰（未扣治疗）。 |
| `TimingOnDamageApplied` | ⑤ 实际产生伤害时 | 伤害已定、扣除治疗后，未摸牌前（如蝶舞者【毒粉】）。 |
| `TimingOnDamageTaken` | ⑥ 实际承受伤害，准备摸牌前 | 摸牌和爆牌判定的前置点。系统应在事件上下文提供 `PendingMoraleLoss`（待扣士气值）供 X 点抵御类技能读取。 |
| **【卡牌与状态流转钩子】** | | |
| `TimingBeforeCardDrawn` | 摸牌动作发生前 | **拦截点**：用于修改摸牌数或劫持为弃牌（如暗杀者水影）。 |
| `TimingOnCardDrawn` | 摸牌动作完成后 | 结算点：用于触发摸牌后的伴生效果。 |
| `TimingOnCardDiscarded` | 弃牌动作完成后 | 系统传入弃掉的卡牌数组，用于判断是否触发额外技能。 |
| `TimingOnCardPlayedOrRevealed` | 持有者打出或展示手牌时 | 元素封印等。配合 `PlayedCardElementMatchType` 做系别匹配（可固定系别，或与 `StatusMeta.BoundElement` 比较）。 |
| `TimingOnHealOverflow` | 获得治疗且超出自身上限时 | 专用于处理溢出转化机制（如圣殿骑士【神选者】）。 |
| `TimingOnFieldMarkChanged`| 基础效果/场上盖牌发生改变时| **高复用点**：系统传入行为`Placed`/`Removed`及变更类型。涵盖：主动移除基础效果、打出圣盾放置于目标（非抵挡触发）。如天使羁绊。 |
| `TimingOnOrientationChanged`| 角色发生横置/转正状态切换时| 触发对姿态敏感的技能（如兽灵武士）。 |

### 4.5 技能执行序列配置 (Effect Nodes Sequence)
用来定义一个技能具体“干了什么”以及“执行的先后顺序”。引擎会遍历 `Effects` 数组，按照 `[0], [1], [2]` 的顺序依次执行。

```go
type SkillDefinition struct {
    // ... 前面的基础信息、Condition、Cost 等保持不变 ...
    
    // 【可选】强制发动配置：命中后在当前行动阶段仅允许执行本技能
    Mandatory *SkillMandatoryConfig

    // 【可选】行动改写配置：在执行前将当前行动重解释到指定流水线
    ActionTransform *ActionTransformConfig

    // 【可选】响应分组配置：同一触发窗口内对候选响应技做组内仲裁（如二选一）
    ResponseGroup *ResponseGroupConfig

    // 【可选】替换技能ID列表：当本技能在响应分组中被选中时，取消这些候选技能的执行
    ReplacesSkillIDs []string

    // 【核心新增】技能的实际按序执行逻辑链！
    Effects []EffectNode 
}

// 技能强制发动配置（通用：解决“命中条件后，本行动阶段只能发动某技能”）
type SkillMandatoryConfig struct {
    MatchTiming model.TriggerTiming // 在哪个钩子上判定是否进入“强制发动”锁（通常为 TimingStartup）
    ConditionExpression string      // 命中强制锁的附加表达式（基于 Event/State/Player）
    LockMode    model.SkillMandatoryLockMode // 锁定模式
}

type SkillMandatoryLockMode int
const (
    SkillMandatoryLockNone                     SkillMandatoryLockMode = 0 // 不锁定（默认）
    SkillMandatoryLockActionPhaseToSelfSkill   SkillMandatoryLockMode = 1 // 本行动阶段仅允许执行该 SkillID
)

// 响应分组配置（通用：解决“同窗口只能二选一”的响应仲裁）
type ResponseGroupConfig struct {
    GroupID string // 响应分组ID（同一触发窗口内按 GroupID 聚合候选技能）
    Mode model.ResponseGroupMode // 分组仲裁模式
    OptionOrder int // 前端展示顺序（小者靠前）
}

type ResponseGroupMode int
const (
    ResponseGroupChooseOne ResponseGroupMode = 0 // 组内最多选择一个响应技能
)

// 单个执行节点
type EffectNode struct {
    ActionType model.EffectActionType // 要执行的具体动作（如：造成伤害、加血、摸牌）
    Target     model.EffectTargetType // 这个动作作用于谁（如：自己、选中的敌人、全场）
    Value      int                    // 动作的数值（如：伤害值、摸牌数）
    
    // 可选：关联的具体实体（比如：如果要放置一个基础效果，放什么？）
    StatusRef  *model.StatusEffect    
    TokenRef   *model.TokenType
    ActionRef  *model.ActionType      // 限制行动大类 (如 Attack, Magic，若无限制填 nil)
    ElementRef *model.ElementType     // 限制卡牌系别 (如 Wind, Fire，若无限制填 nil)；用于 EffectSetCurrentCombatElement 时表示改写后的战斗系别；用于 EffectRemoveFieldMark 时可按系别过滤要移除的场标牌
    StoneRef   *model.StarStoneType   // 用于 EffectAddTeamStone / EffectAddEnergyStone / EffectConvertTeamStone / EffectConvertEnergyStone：指定源颜色或星石类型（Gem/Crystal/Any）
    StoneToRef *model.StarStoneType   // 用于 EffectConvertTeamStone / EffectConvertEnergyStone：目标星石类型（Gem/Crystal）
    FromTargetRef *model.EffectTargetType // 用于 EffectTransferTeamStone / EffectTransferCard / EffectTransferFieldMark：源目标（队伍/玩家）
    FieldMarkRef *model.CardFieldMark // 用于 EffectRemoveFieldMark / EffectPlaceDeckTopAsFieldMark / EffectPlacePlayedCardAsFieldMark / EffectPlaceHandCardAsFieldMark / EffectTransferFieldMark：指定场标类型（如 Blessing）
    VisibilityRef *model.CardVisibilityType // 用于 EffectPlacePlayedCardAsFieldMark / EffectPlaceHandCardAsFieldMark：指定放置后明暗（VisibilityPublic/VisibilityHidden）
    OrientationRef *model.CharacterOrientation // 用于 EffectSetOrientation：设置姿态（Normal/Tapped）
    FormRef    *string                // 用于 EffectSetForm：设置命名形态（nil 表示清空）
    BranchRef  *model.PerTargetBranchConfig // 用于 EffectPerTargetBranch：逐目标响应分支
    RuleModifierRef *string           // 用于 EffectApplyRuleModifier：引用 RuleModifierTemplate.ModifierID
    RuleLifetimeRef *model.RuleModifierLifetimeType // 用于 EffectApplyRuleModifier：实例持续时长
    RuleRemoveRef *model.RuleModifierRemoveQuery // 用于 EffectRemoveRuleModifier：移除筛选条件
}

// —— 动作类型枚举 (Effect Action Type) ——
type EffectActionType int
const (
    EffectNone                 EffectActionType = 0
    EffectDamage               EffectActionType = 1  // 造成法术伤害
    EffectAttackDamage         EffectActionType = 2  // 造成攻击伤害 (附加物理命中判定)
    EffectHeal                 EffectActionType = 3  // 调整治疗（正数=增加；负数=移除；最小不低于0）
    EffectDrawCard             EffectActionType = 4  // 摸牌
    EffectDiscard              EffectActionType = 5  // 强制目标弃牌
    EffectAddToken             EffectActionType = 6  // 调整专属指示物（Value>0 增加；Value<0 移除；移除不足按可移除量结算且不阻断）
    EffectAddAction            EffectActionType = 7  // 增加额外行动次数 (如: +1 攻击行动)
    EffectPlaceStatus          EffectActionType = 8  // 放置基础效果 (如: 中毒、五系束缚)
    EffectRemoveStatus         EffectActionType = 9  // 移除基础效果
    EffectRemoveStatusToHand   EffectActionType = 10 // 将场上基础效果牌收入手牌 (如封印破碎)
    EffectAttackDamageModifier EffectActionType = 11 // 攻击伤害增减修饰 (如狂化、撕裂)
    EffectChangeMorale         EffectActionType = 12 // 直接扣除/增加已生效士气（立即生效）
    EffectTransferCard         EffectActionType = 13 // 转移卡牌（FromTargetRef -> Target；FromTargetRef 为空时兼容旧语义：Target -> Self）
    EffectAdjustHand           EffectActionType = 14 // 将手牌调整为X张 (如圣煌辉光炮)
    EffectSwapMorale           EffectActionType = 15 // 将一方士气调整与另一方相同
    EffectApplyCombatTag       EffectActionType = 16 // 将 Value 解释为 CombatInterceptTag 并写入 CombatContext
    EffectReducePendingMoraleLoss EffectActionType = 17 // 将当前事件窗口中的 PendingMoraleLoss 减少 Value（最小为 0）
    EffectCancelHit EffectActionType = 18 // 取消命中/抵挡：替代 SystemOp_BlockCurrentAttackOrMagicBullet
    EffectSkipPhase EffectActionType = 19 // 跳过阶段：替代 SystemOp: SkipCurrentActionPhase,配合 EffectNode 中的 Value 或 ActionRef 使用，指明跳过什么阶段
    EffectAddTeamStone EffectActionType = 20 // 调整目标阵营战绩区星石（StoneRef 指定 Gem/Crystal/Any；Value>0 增加，Value<0 移除；移除不足时按可移除量结算且不阻断）
    EffectPerTargetBranch EffectActionType = 21 // 对选中目标逐个发起响应，并按成功/失败分支结算
    EffectAddEnergyStone EffectActionType = 22 // 调整目标玩家能量区星石（StoneRef 指定 Gem/Crystal/Any；Value>0 为增加，Value<0 为移除；移除不足时按可移除量结算且不阻断）
    EffectSetOrientation EffectActionType = 23 // 设置目标姿态（OrientationRef 指定 Normal/Tapped）
    EffectSetForm EffectActionType = 24 // 设置/清空目标命名形态（FormRef 为空表示清空）
    EffectSetHandLimitFixed EffectActionType = 25 // 设置目标手牌上限固定值（Value；>0 生效）
    EffectTransferTeamStone EffectActionType = 26 // 在队伍战绩区间转移星石（FromTargetRef=源队伍，Target=目标队伍，StoneRef=类型，Value=数量）
    EffectConvertTeamStone EffectActionType = 27 // 转换队伍战绩区星石颜色（Target=队伍，StoneRef=源颜色，StoneToRef=目标颜色，Value<=0 表示全部）
    EffectRedirectCurrentExtractOutput EffectActionType = 28 // 将当前 Extract 行动产出的星石重定向给 TargetSelected（可结合 Action.TargetAllocations 分配）
    EffectApplyRuleModifier EffectActionType = 29 // 向目标施加规则修饰实例（RuleModifierRef + RuleLifetimeRef）
    EffectRemoveRuleModifier EffectActionType = 30 // 按 RuleRemoveRef 条件移除目标规则修饰实例
    EffectPlaceDeckTopAsFieldMark EffectActionType = 31 // 将目标牌库顶 Value 张牌放置到目标角色面前并赋予 FieldMarkRef（默认面朝下）
    EffectRemoveFieldMark EffectActionType = 32 // 移除目标角色面前指定 FieldMarkRef 的场标牌（Value 为数量；不足按可移除量结算；若 ElementRef 非空则仅移除该系别）；并刷新 Event.RemovedFieldCard* 为最近被移除牌快照
    EffectConvertEnergyStone EffectActionType = 33 // 转换目标玩家能量区星石颜色（StoneRef=源颜色，StoneToRef=目标颜色，Value<=0 表示全部）
    EffectModifyPendingDamage EffectActionType = 34 // 修改当前战斗结算链的待生效伤害值（TargetCurrentCombat；Value 可正可负；最终不低于 0）
    EffectPlacePlayedCardAsFieldMark EffectActionType = 35 // 将“本次打出的牌实体”放置到目标角色面前并赋予 FieldMarkRef（可结合 VisibilityRef 指定明暗）
    EffectSetCurrentCombatElement EffectActionType = 36 // 改写当前战斗上下文系别（通常 Target=TargetCurrentCombat，ElementRef 必填）
    EffectRedirectCurrentCombatTarget EffectActionType = 37 // 改写当前战斗承受者（Target=新承受者）
    EffectSetCurrentCounterExecutor EffectActionType = 38 // 改写当前战斗“应战执行者”（Target=新执行者）
    EffectPlaceHandCardAsFieldMark EffectActionType = 39 // 将目标手牌中的 Value 张牌放置到目标角色面前并赋予 FieldMarkRef（通常由 Action.UsedCardUUIDs 指定，支持 VisibilityRef）
    EffectRemoveSelectedFieldCard EffectActionType = 40 // 移除 Action.Targets[].SelectedFieldCards 指定的场上牌（最多 Value 张）；并刷新 Event.RemovedFieldCard* 为最近被移除牌快照
    EffectRevealRemovedFieldCard EffectActionType = 41 // 将 Event.RemovedFieldCard 公开展示（若该牌原为暗置场标）
    EffectAddTeamCup EffectActionType = 42 // 调整目标阵营星杯数（Target=TargetSelfTeam/TargetEnemyTeam；Value>0 增加，Value<0 移除；结果钳制在 [0, TargetStarCups]）
    EffectPlaceOverflowDiscardAsFieldMark EffectActionType = 43 // 将 Event.OverflowDiscardCardIDs 中的卡牌实体放置到目标角色面前并赋予 FieldMarkRef（通常用于“爆牌弃牌留场”）
    EffectGrantExtraTurn EffectActionType = 44 // 为目标角色追加额外回合（Value=追加回合数；Value<=0 按 1 处理）
    EffectTransferFieldMark EffectActionType = 45 // 将在场 FieldMark 实体从 FromTargetRef 指向范围迁移到 Target 指向角色（保持原实体与明暗）
)

// —— 动作目标枚举 (Effect Target Type) ——
type EffectTargetType int
const (
    TargetNone            EffectTargetType = 0
    TargetSelf            EffectTargetType = 1 // 自己
	TargetSelected        EffectTargetType = 2 // 玩家刚才选中的合法目标
    TargetAllEnemies      EffectTargetType = 3 // 所有敌人 (AOE)
    TargetAllTeammates    EffectTargetType = 4 // 所有队友 (含自己)
    TargetAllPlayers      EffectTargetType = 5 // 全场所有人
    TargetTriggerSource   EffectTargetType = 6 // 抽象指针：刚才触发这个被动事件的源头玩家 (比如反伤技)
    TargetSelfTeam        EffectTargetType = 7 // 己方阵营 (如神之庇护抵御士气下降)
    TargetCurrentCombat   EffectTargetType = 8 // 当前战斗上下文 (用于伤害修饰等)
    TargetCurrentEvent    EffectTargetType = 9 // 当前事件上下文 (用于修改 PendingMoraleLoss 等待结算值)
    TargetEnemyTeam       EffectTargetType = 10 // 敌方阵营（用于跨队星石转移）
    TargetAllOthers       EffectTargetType = 11 // 全体其他角色（排除技能发动者自己）
    TargetAllExceptSelected EffectTargetType = 12 // 全体角色中排除 SubmitAction 已选目标（用于“除所选目标外其余角色”）
)

// —— 规则修饰器模型（Rule Modifier） —— 
type PlayerAttributeType int
const (
    PlayerAttributeMaxHand   PlayerAttributeType = 0 // 手牌上限
    PlayerAttributeMaxEnergy PlayerAttributeType = 1 // 能量上限
    PlayerAttributeMaxHeal   PlayerAttributeType = 2 // 治疗上限
)

type AttributeModifyOpType int
const (
    AttributeModifyAdd AttributeModifyOpType = 0 // 在当前值基础上 +Value
    AttributeModifySet AttributeModifyOpType = 1 // 直接设为 Value
)

type RuleAttrValueSourceMode int
const (
    RuleAttrValueFromFixed RuleAttrValueSourceMode = 0 // 使用 RuleAttrPayload.Value 固定值（默认）
    RuleAttrValueFromExpression RuleAttrValueSourceMode = 1 // 使用 RuleAttrPayload.ValueExpression 动态表达式
    RuleAttrValueFromTokenLinear RuleAttrValueSourceMode = 2 // 使用 RuleAttrPayload.TokenLink（按 Token 数量线性计算）
)

type RuleAttrTokenOwnerScope int
const (
    RuleAttrTokenOwnerTarget RuleAttrTokenOwnerScope = 0 // 按规则作用目标（Target）的 Token 计算
    RuleAttrTokenOwnerSource RuleAttrTokenOwnerScope = 1 // 按规则施加来源（Source）的 Token 计算
)

type HealApplyMode int
const (
    HealApplyRespectMax HealApplyMode = 0 // 按目标当前有效治疗上限结算（默认）
    HealApplyIgnoreMax  HealApplyMode = 1 // 忽略目标当前有效治疗上限
)

type RuleModifierDomain int
const (
    RuleModifierDomainAttribute  RuleModifierDomain = 0 // 属性域（如 MaxHeal+1）
    RuleModifierDomainHealPolicy RuleModifierDomain = 1 // 治疗策略域（如 IgnoreMax+AbsoluteMax）
    RuleModifierDomainSkillGate  RuleModifierDomain = 2 // 技能门禁域（如禁用指定技能）
    RuleModifierDomainCardSource RuleModifierDomain = 3 // 卡牌来源域（如指定场标可视为手牌）
    RuleModifierDomainTokenPolicy RuleModifierDomain = 4 // 指示物策略域（如 IgnoreMax+AbsoluteMax，用于鲜血/战纹等）
    RuleModifierDomainHealResistPolicy RuleModifierDomain = 5 // 治疗抵伤策略域（如每次受伤窗口最多用1点治疗抵伤）
    RuleModifierDomainMoralePolicy RuleModifierDomain = 6 // 士气策略域（如阵营士气下限/上限）
)

type MoralePolicyApplyScope int
const (
    MoralePolicyApplySourceTeam MoralePolicyApplyScope = 0 // 作用于规则施加者所属阵营
    MoralePolicyApplyEnemyTeam  MoralePolicyApplyScope = 1 // 作用于规则施加者敌方阵营
    MoralePolicyApplyTargetTeam MoralePolicyApplyScope = 2 // 作用于规则目标所属阵营
    MoralePolicyApplyAllTeams   MoralePolicyApplyScope = 3 // 同时作用于双方阵营
)

type RuleModifierStackPolicy int
const (
    RuleModifierStackAppend RuleModifierStackPolicy = 0 // 直接叠加
    RuleModifierStackRefreshByModifierID RuleModifierStackPolicy = 1 // 相同 ModifierID 时刷新时长
    RuleModifierStackReplaceByDomainPriority RuleModifierStackPolicy = 2 // 同域按优先级替换
)

type RuleModifierLifetimeType int
const (
    RuleLifeThisEffectChain RuleModifierLifetimeType = 0 // 当前技能执行链结束后失效
    RuleLifeUntilTurnEnd RuleModifierLifetimeType = 1 // 到当前回合结束
    RuleLifeUntilSourceNextTurnStart RuleModifierLifetimeType = 2 // 到施加者下回合开始
    RuleLifeUntilSourceNextTurnEnd RuleModifierLifetimeType = 3 // 到施加者下回合结束
    RuleLifePermanent RuleModifierLifetimeType = 4 // 常驻，需显式移除
    RuleLifeUntilCombatEnd RuleModifierLifetimeType = 5 // 到当前战斗结算链结束（CombatDraw 收尾后）
)

type SkillGateMode int
const (
    SkillGateDisallowList SkillGateMode = 0 // 禁用 SkillIDs 列表中的技能
)

type CardSourceProjectionMode int
const (
    CardSourceProjectionAsHand CardSourceProjectionMode = 0 // 将目标场标牌投影为手牌候选来源
)

type TokenApplyMode int
const (
    TokenApplyRespectMax TokenApplyMode = 0 // 按目标当前有效指示物上限结算（默认）
    TokenApplyIgnoreMax  TokenApplyMode = 1 // 忽略目标当前有效指示物上限
)

type RuleModifierTemplate struct {
    ModifierID string                      // 模板ID（全局唯一）
    Domain model.RuleModifierDomain       // 规则域
    Priority int                          // 优先级（大者先应用）
    ConditionExpression string            // 命中条件，空字符串表示恒定生效
    StackPolicy model.RuleModifierStackPolicy // 叠加策略

    AttrPayload *model.RuleAttrPayload
    HealPolicyPayload *model.RuleHealPolicyPayload
    SkillGatePayload *model.RuleSkillGatePayload
    CardSourcePayload *model.RuleCardSourcePayload
    TokenPolicyPayload *model.RuleTokenPolicyPayload
    HealResistPolicyPayload *model.RuleHealResistPolicyPayload
    MoralePolicyPayload *model.RuleMoralePolicyPayload
}

type RuleAttrPayload struct {
    AttrType model.PlayerAttributeType
    Operation model.AttributeModifyOpType
    ValueSourceMode model.RuleAttrValueSourceMode // 取值来源模式；缺省按 RuleAttrValueFromFixed 处理
    Value int                                     // ValueSourceMode=RuleAttrValueFromFixed 时生效
    ValueExpression string                        // ValueSourceMode=RuleAttrValueFromExpression 时生效
    TokenLink *model.RuleAttrTokenLinkPayload     // ValueSourceMode=RuleAttrValueFromTokenLinear 时生效
}

type RuleAttrTokenLinkPayload struct {
    OwnerScope model.RuleAttrTokenOwnerScope // 按谁的 Token 计算（Target / Source）
    TokenType model.TokenType
    Coefficient int // 线性系数（可正可负）
    Offset int      // 线性偏移（可正可负）
    MinValue *int   // 可选：下限钳制
    MaxValue *int   // 可选：上限钳制
}

type RuleHealPolicyPayload struct {
    ApplyMode model.HealApplyMode
    AbsoluteMax *int // 可选绝对封顶（如 5）
}

type RuleSkillGatePayload struct {
    Mode model.SkillGateMode
    SkillIDs []string
}

type RuleCardSourcePayload struct {
    ProjectionMode model.CardSourceProjectionMode
    FieldMarks []model.CardFieldMark
}

type RuleTokenPolicyPayload struct {
    TokenType model.TokenType
    ApplyMode model.TokenApplyMode
    AbsoluteMax *int // 可选绝对封顶（如 4）
}

type RuleHealResistPolicyPayload struct {
    PerDamageWindowHealCap *int // 每次受伤窗口可使用的治疗抵伤上限（nil 表示不限制）
}

type RuleMoralePolicyPayload struct {
    ApplyScope model.MoralePolicyApplyScope
    MinMorale *int // 可选：士气下限（nil 表示不限制下限）
    MaxMorale *int // 可选：士气上限（nil 表示不限制上限）
}

type RuleModifierRemoveMode int
const (
    RuleRemoveByModifierID RuleModifierRemoveMode = 0
    RuleRemoveByDomain RuleModifierRemoveMode = 1
    RuleRemoveBySourceSkill RuleModifierRemoveMode = 2
    RuleRemoveAll RuleModifierRemoveMode = 3
)

type RuleModifierRemoveQuery struct {
    Mode model.RuleModifierRemoveMode
    ModifierID *string
    Domain *model.RuleModifierDomain
    SourceSkillID *string
    Limit int // <=0 表示不限制数量
}
```

```go
// —— 逐目标响应分支配置 —— 
type PerTargetBranchConfig struct {
    TargetSource       model.PerTargetSourceType  // 遍历来源
    InterruptType      model.InterruptType        // 对每个目标发起的中断（如 WaitDiscard / WaitChoice）
    TimeoutAsDeclined  bool                       // 超时是否按失败分支处理
    DiscardRequirement *model.DiscardRequirement  // WaitDiscard 模式下，目标需满足的弃牌要求
    DiscardVisibility  *model.CardVisibilityType  // 弃牌可见性（如 VisibilityPublic）
    OnSuccess          []EffectNode               // 目标响应成功时执行
    OnDeclined         []EffectNode               // 目标响应失败/拒绝时执行
}
```

#### 4.5.1 强制发动流程 (Mandatory Trigger Flow)
当某个技能配置了 `Mandatory` 且命中其条件时，统一按以下流程处理：
1. 在 `Mandatory.MatchTiming` 到达时，先校验 `Mandatory.ConditionExpression`。
2. 命中后根据 `Mandatory.LockMode` 写入行动阶段锁（如 `SkillMandatoryLockActionPhaseToSelfSkill`）。
3. 行动阶段内，玩家提交 `SubmitAction` 时仅允许 `ActionType=Skill` 且 `SkillID=锁定技能ID`。
4. 任何 `Attack / Magic / Buy / Synthesize / Extract / Deadlock / Pass` 均判定为非法输入。
5. 当玩家提交该技能时，仍需按常规流程完整校验该技能 `Condition/Cost/TargetRule`。
6. 锁在本行动阶段结束后自动清理（进入 `ActionEnd` 或 `TurnEnd` 时释放）。

#### 4.5.2 规则修饰器应用流程 (Rule Modifier Resolve Flow)
`RuleModifier` 采用“模板静态定义 + 实例运行时挂载”的统一流程：
1. `EffectApplyRuleModifier` 通过 `RuleModifierRef` 引用一条模板，并在 `Target` 上创建实例。
2. 实例记录来源 (`SourceUserID/SourceSkillID`) 与 `RuleLifetimeRef`，进入目标玩家的生效规则池。
3. 规则读取点（属性查询、治疗结算、治疗抵伤额度校验、技能可用性校验）统一走 `RuleResolver`：
   - 先取基础值；
   - 收集命中 `ConditionExpression` 的规则；
   - 按 `Priority` 与 `StackPolicy` 归并；
   - 对属性域规则按 `ValueSourceMode` 解析修饰值（固定值 / 表达式 / Token 联动）；
   - 输出最终结算结果。

#### 4.5.3 规则施加与移除 (EffectApplyRuleModifier / EffectRemoveRuleModifier)
1. `EffectApplyRuleModifier`：
   - 必填：`RuleModifierRef`；
   - 可选：`RuleLifetimeRef`（为空时默认 `RuleLifeThisEffectChain`）。
2. `EffectRemoveRuleModifier`：
   - 通过 `RuleRemoveRef` 指定移除条件（按 `ModifierID/Domain/SourceSkill/All`）。
3. 规则实例达到生命周期终点时自动过期；也可被 `EffectRemoveRuleModifier` 提前移除。

#### 4.5.4 关键规则域语义 (Attribute / HealPolicy / SkillGate / CardSource / TokenPolicy / HealResistPolicy / MoralePolicy)
1. 属性域（`RuleModifierDomainAttribute`）：用于修饰 `MaxHand/MaxEnergy/MaxHeal`。
   - `RuleAttrPayload.ValueSourceMode` 支持三种取值：
     - `RuleAttrValueFromFixed`：直接使用 `Value`；
     - `RuleAttrValueFromExpression`：运行时计算 `ValueExpression`；
     - `RuleAttrValueFromTokenLinear`：按 `TokenLink` 线性计算。
   - `TokenLink` 线性公式：`resolved = Coefficient * tokenCount + Offset`；
   - 若配置 `MinValue/MaxValue`，再对 `resolved` 进行上下限钳制；
   - `Operation=AttributeModifyAdd` 时在基础值上做增量；`Operation=AttributeModifySet` 时直接覆盖为 `resolved`。
2. 治疗策略域（`RuleModifierDomainHealPolicy`）：
   - `HealApplyRespectMax`: 按当前有效治疗上限结算；
   - `HealApplyIgnoreMax`: 忽略当前有效治疗上限；
   - `AbsoluteMax` 非空时再做绝对封顶。
3. 技能门禁域（`RuleModifierDomainSkillGate`）：
   - 当前支持 `SkillGateDisallowList`，命中 `SkillIDs` 则技能不可提交。
4. 卡牌来源域（`RuleModifierDomainCardSource`）：
   - 当前支持 `CardSourceProjectionAsHand`；
   - 命中规则后，`FieldMarks` 指定的场标牌会并入“手牌候选来源”，可参与打出/展示/弃置等手牌相关校验。
5. 指示物策略域（`RuleModifierDomainTokenPolicy`）：
   - `TokenApplyRespectMax`: 增加指示物时按当前有效上限结算；
   - `TokenApplyIgnoreMax`: 增加指示物时忽略当前有效上限；
   - `AbsoluteMax` 非空时再做绝对封顶（用于“无视上限但最高为N”）。
6. 治疗抵伤策略域（`RuleModifierDomainHealResistPolicy`）：
   - `PerDamageWindowHealCap`：约束“单次受伤窗口（WaitHeal）”可投入的治疗抵伤上限；
   - 多条命中时按规则优先级与叠加策略归并，最终得到当前窗口允许投入的最大治疗值。
7. 士气策略域（`RuleModifierDomainMoralePolicy`）：
   - `ApplyScope` 决定作用阵营（施加者本方 / 敌方 / 规则目标方 / 全体）；
   - `MinMorale`/`MaxMorale` 用于声明士气边界；
   - 引擎在“士气变动落地点”统一执行钳制（包括 `PendingMoraleLoss` 收尾与 `EffectChangeMorale` 直接改值）；
   - 多条命中时按规则优先级与叠加策略归并：下限取最大值、上限取最小值。

#### 4.5.5 指示物类型互换语义 (WarRune <-> MagicRune)
1. 战纹/魔纹不使用“朝向”建模，统一作为两个独立 `TokenType`：`WarRune` 与 `MagicRune`。
2. “翻转 N 个战纹为魔纹”的标准实现：
   - `EffectAddToken(TokenRef=WarRune, Value=-N)`；
   - `EffectAddToken(TokenRef=MagicRune, Value=+N)`。
3. “翻转 N 个魔纹为战纹”同理，交换两者即可。
4. `EffectAddToken` 的运行时约定：
   - `Value > 0`：增加；
   - `Value < 0`：移除；
   - 实际移除数量 = `min(abs(Value), 当前该 Token 数量)`，不足按可移除量结算，不阻断后续 Effect。
5. 若技能文本要求“必须翻转 N 个才能发动”，应在 `CustomExpression` 中显式校验源类型数量（如 `Self.Tokens[WarRune] >= N`）。

#### 4.5.6 响应分组仲裁流程 (Response Group Arbitration Flow)
用于引擎层硬约束“同一触发窗口只能二选一”：
1. 在某个 `Timing` 触发时，先按常规流程收集所有可触发 `Type=Response` 的候选技能。
2. 将携带 `ResponseGroup` 的候选技能按 `GroupID` 分组；每组内按 `OptionOrder` 排序用于前端弹窗展示。
3. 当 `Mode=ResponseGroupChooseOne` 时，玩家在该组内最多选择一个技能提交；引擎拒绝“同组多选”。
4. 若被选技能配置了 `ReplacesSkillIDs`，则在同一触发窗口内将这些技能标记为 `CancelledByReplacement`，不进入执行队列。
5. 最终仅对被选中的技能执行 `Cost` 与 `Effects`；被取消技能不扣费、不执行效果。

#### 4.5.7 事件处理伪代码（候选收集 -> 分组 -> 选择 -> 替换 -> 入执行队列）
以下伪代码可直接映射到后端事件总线与中断系统实现：

```go
// 响应候选（运行时）
type ResponseCandidate struct {
    OwnerID          string
    SkillID          string
    Timing           model.TriggerTiming
    Group            *model.ResponseGroupConfig // nil 表示不参与分组仲裁
    ReplacesSkillIDs []string

    // 供前端展示与后续执行使用
    PromptPayload model.PromptPayload
}

// 收集某个触发窗口下的响应候选
func CollectResponseCandidates(
    state *GameState,
    evt *EventContext,
    timing model.TriggerTiming,
) []ResponseCandidate {
    out := make([]ResponseCandidate, 0)
    for _, user := range state.AllAliveUsers() {
        for _, skill := range state.GetUserSkills(user.UserID) {
            if skill.Type != model.Response || skill.Timing != timing {
                continue
            }
            // 轻校验：先过滤掉肯定不可能触发的技能
            if !EvalConditionSoft(skill, user.UserID, evt, state) {
                continue
            }
            if !CanBuildPrompt(skill, user.UserID, evt, state) {
                continue
            }
            out = append(out, ResponseCandidate{
                OwnerID:          user.UserID,
                SkillID:          skill.SkillID,
                Timing:           timing,
                Group:            skill.ResponseGroup,
                ReplacesSkillIDs: skill.ReplacesSkillIDs,
                PromptPayload:    BuildPromptPayload(skill, user.UserID, evt, state),
            })
        }
    }
    return out
}

// 将候选按“拥有者+分组ID”聚合；分组内按 OptionOrder 排序
func BuildResponseGroups(cands []ResponseCandidate) (
    grouped map[string][]ResponseCandidate, // key: ownerID + "#" + groupID
    ungrouped []ResponseCandidate,
) {
    grouped = map[string][]ResponseCandidate{}
    for _, c := range cands {
        if c.Group == nil || c.Group.GroupID == "" {
            ungrouped = append(ungrouped, c)
            continue
        }
        key := c.OwnerID + "#" + c.Group.GroupID
        grouped[key] = append(grouped[key], c)
    }
    for key := range grouped {
        sort.Slice(grouped[key], func(i, j int) bool {
            return grouped[key][i].Group.OptionOrder < grouped[key][j].Group.OptionOrder
        })
    }
    return grouped, ungrouped
}

// 触发窗口总流程：候选收集 -> 分组展示 -> 玩家选择 -> 替换 -> 入队执行
func ResolveResponseWindow(state *GameState, evt *EventContext, timing model.TriggerTiming) error {
    cands := CollectResponseCandidates(state, evt, timing)
    grouped, ungrouped := BuildResponseGroups(cands)

    // 1) 弹窗（含分组技能 + 非分组技能）
    choiceReq := BuildResponseChoiceRequest(grouped, ungrouped)
    choiceResp, err := WaitPlayerChoice(choiceReq) // 中断类型可复用 WaitChoice
    if err != nil {
        return err
    }

    // 2) 选择合法性校验（硬约束）
    // - 分组模式为 ResponseGroupChooseOne 时，每组最多一个
    // - 选择项必须在候选清单中
    if err := ValidateResponseChoice(choiceResp, grouped, ungrouped); err != nil {
        return err
    }

    // 3) 对已选择技能做“硬校验”（费用/目标/条件二次确认）
    selected := MaterializeSelectedCandidates(choiceResp, grouped, ungrouped)
    executable := make([]ResponseCandidate, 0, len(selected))
    for _, s := range selected {
        if !EvalConditionHard(s, evt, state) {
            continue
        }
        if !CheckCostAndTargetHard(s, choiceResp, evt, state) {
            continue
        }
        executable = append(executable, s)
    }

    // 4) 应用替换规则（ReplacesSkillIDs）
    // 语义：若 A 被选中且 A.ReplacesSkillIDs 包含 B，则 B 在同一触发窗口取消执行
    cancelled := map[string]bool{} // key: ownerID + "#" + skillID
    for _, s := range executable {
        for _, replacedID := range s.ReplacesSkillIDs {
            key := s.OwnerID + "#" + replacedID
            cancelled[key] = true
        }
    }

    // 5) 入执行队列（被替换技能不入队，不扣费，不执行）
    for _, s := range executable {
        key := s.OwnerID + "#" + s.SkillID
        if cancelled[key] {
            continue
        }
        state.EnqueueSkillExecution(SkillExecutionItem{
            OwnerID: s.OwnerID,
            SkillID: s.SkillID,
            Event:   evt,
            Timing:  timing,
            Source:  SkillExecutionSourceResponseWindow,
        })
    }
    return nil
}
```

实现约束建议：
1. `ValidateResponseChoice` 必须在服务端执行，不能只依赖前端禁用按钮。
2. `ResponseGroupChooseOne` 的“每组最多一个”属于协议硬错误，命中时应拒绝本次提交。
3. `ReplacesSkillIDs` 的替换作用域限定在“同一触发窗口”内，不跨窗口追溯。

#### 4.5.8 卡牌转移方向语义 (EffectTransferCard Direction)
1. `EffectTransferCard` 支持显式方向：
   - `FromTargetRef` 非空：从 `FromTargetRef` 指向的玩家转移到 `Target` 指向的玩家；
   - `FromTargetRef` 为空：兼容旧语义，按 `Target -> Self` 执行。
2. `Value` 表示转移张数；若来源不足，按可转移量结算，不阻断后续 Effect。
3. 若 `SubmitAction` 未预选具体卡牌（如 `SelectedHandCards` 为空），引擎在执行该 Effect 时对来源玩家发起补选中断，强制选满 `Value` 张。
4. 对应“动物伙伴/宠物强化”场景：两者同组后，玩家在同一弹窗内只能选一个；选中宠物强化时，动物伙伴在该窗口不会入队执行。

#### 4.5.9 当前战斗改写语义 (Combat Rewrite Effects)
用于表达“在当前战斗窗口内改写上下文”的通用能力，不绑定任意具体角色技能。
1. `EffectSetCurrentCombatElement`：
   - 语义：将当前 `CombatContext` 的生效系别改为 `ElementRef`；
   - 约束：`ElementRef` 必填，`Target` 建议固定为 `TargetCurrentCombat`；
   - 用途：支持“应战后改变本次攻击系别”这类规则。
2. `EffectRedirectCurrentCombatTarget`：
   - 语义：将当前 `CombatContext.TargetID` 改写为 `Target` 指向的角色；
   - 约束：`Target` 必须可解析为单一玩家目标；
   - 用途：支持“代替队友承受本次攻击”这类规则。
3. `EffectSetCurrentCounterExecutor`：
   - 语义：将当前 `CombatContext.CounterExecutorID` 改写为 `Target` 指向的角色；
   - 约束：`Target` 必须可解析为单一玩家目标；
   - 用途：支持“视为由某角色执行本次应战攻击”这类规则。
4. 结算顺序建议：
   - 若同一技能同时改写目标与应战执行者，建议先执行 `EffectRedirectCurrentCombatTarget`，再执行 `EffectSetCurrentCounterExecutor`，避免中间态歧义。
5. 与 `ActionTransform` 的边界：
   - `ActionTransform` 用于“把一个待执行行动改写成另一条流水线”；
   - 本节三个 Effect 用于“当前已进入的战斗流水线内，改写战斗上下文字段”。

#### 4.5.10 场标移除与展示语义 (FieldCard Remove & Reveal)
用于表达“移除场标实体，并基于最近被移除牌快照驱动后续分支”的通用能力。
1. `EffectRemoveFieldMark`：
   - 语义：按 `FieldMarkRef`（可叠加 `ElementRef`）批量移除场标实体；
   - 上下文：每移除 1 张，更新一次 `Event.RemovedFieldCard*`；Effect 结束后字段保留“最后一张被移除牌”快照。
2. `EffectRemoveSelectedFieldCard`：
   - 语义：移除 `SubmitAction.Targets[].SelectedFieldCards` 指定的场上牌实体（最多 `Value` 张）；
   - 约束：前置目标规则需确保客户端已提交合法 `SelectedFieldCards`；
   - 上下文：每移除 1 张，更新一次 `Event.RemovedFieldCard*`；Effect 结束后字段保留“最后一张被移除牌”快照。
3. `EffectRevealRemovedFieldCard`：
   - 语义：将 `Event.RemovedFieldCard*` 指向的牌面公开展示；
   - 约束：若上下文中不存在被移除牌，执行为空操作（不阻断后续 Effect）。
4. `TargetAllExceptSelected`：
   - 语义：作用目标为“全体角色中排除 SubmitAction 已选目标”；
   - 用途：支撑“指定若干排除对象后，对其余所有角色生效”的群体技能。
5. 表达式辅助函数（供 `Condition.CustomExpression` 使用）：
   - `State.GetSelectedFieldCardElement(selfUserID, actionTargets)`：读取本次提交中“由自身提交的 SelectedFieldCards 首张牌”对应原卡牌系别（未命中时返回 `None`）。

#### 4.5.11 批量弃牌结果统计语义 (Batch Discard Result Stats)
用于表达“先批量弃牌，再基于弃牌结果继续结算”的通用能力（如：弃全部手牌后按法术张数/系别张数增益）。
1. `EventContext` 增加以下运行时字段：
   - `DiscardedMagicCount int`：当前结算链最近一次批量弃牌中，`CardType == Magic` 的数量；
   - `DiscardedElementCount map[model.ElementType]int`：当前结算链最近一次批量弃牌中，按系别统计的数量（如 `Water/Fire/...`）。
2. 写入时机：
   - 当 `EffectDiscard` 执行批量弃牌（包括 `Value=Self.HandCount` 的“弃全部手牌”）后，立即刷新上述两个字段。
3. 读取方式：
   - 条件与效果表达式可直接读取：`Event.DiscardedMagicCount`、`Event.DiscardedElementCount[Water]`、`Event.DiscardedElementCount[Fire]` 等。
4. 作用域约束：
   - 字段作用域限定在“当前结算链”；下一条独立链路开始前应重置，避免跨链污染。

#### 4.5.12 阵营星杯增减语义 (Team Cup Delta)
用于表达“增加/移除阵营星杯”的通用能力（如“我方星杯区 +1 星杯”）。
1. 新增 `EffectAddTeamCup`：
   - `Target` 建议固定为 `TargetSelfTeam` 或 `TargetEnemyTeam`；
   - `Value > 0` 表示增加；`Value < 0` 表示移除（不足按可移除量结算）。
2. 结算钳制：
   - 目标星杯值始终钳制在 `[0, TargetStarCups]`。
3. 胜负衔接：
   - 当星杯增加后命中 `TargetStarCups`，按现有胜利判定流程在结算收尾触发。

#### 4.5.13 爆牌弃牌留场语义 (Overflow Discard To FieldMark)
用于表达“由超手牌上限导致的爆牌弃置，在同窗口被技能改写为场上资源”这类能力（如暗月/护盾型留场机制）。
1. `EventContext` 增加以下运行时字段：
   - `OverflowDiscardOwnerID string`：本次因超手牌上限发生爆牌弃置的角色；
   - `OverflowDiscardCardIDs []string`：该次爆牌弃置产生的卡牌实体列表（按弃置顺序）；
   - `OverflowDiscardCount int`：`OverflowDiscardCardIDs` 的快照数量（便于表达式直接比较，无需 `len`）。
2. 新增 `EffectPlaceOverflowDiscardAsFieldMark`：
   - 语义：将 `Event.OverflowDiscardCardIDs` 中的牌实体，放置到 `Target` 指向角色面前，并标注为 `FieldMarkRef`；
   - 明暗：放置后明暗由 `VisibilityRef` 决定（未填时默认 `VisibilityHidden`）；
   - 数量：`Value <= 0` 表示搬运全部；`Value > 0` 表示最多搬运前 `Value` 张（不足按可搬运量结算）。
3. 消费约束：
   - 被搬运成功的牌从 `Event.OverflowDiscardCardIDs` 中消费移除，避免同链路重复搬运。
4. 作用域：
   - `OverflowDiscard*` 字段仅在当前爆牌结算链有效；链路结束后应清空。

#### 4.5.14 额外回合调度语义 (Grant Extra Turn)
用于表达“在当前回合结算后追加额外回合”的通用能力。
1. 新增 `EffectGrantExtraTurn`：
   - `Target`：额外回合归属角色（常见为 `TargetSelf`）；
   - `Value`：追加回合数；`Value <= 0` 按 `1` 处理。
2. 调度规则：
   - 额外回合插入到“当前回合结束后、下一名常规回合玩家之前”；
   - 若同一窗口多次追加，同一角色的额外回合按入队顺序依次执行。
3. 归属与结算：
   - 额外回合完整执行“回合开始/行动/结束”流程；
   - 额外回合不改变常规轮转顺序，仅作为当前轮转链中的插队节点。

#### 4.5.15 场标迁移语义 (Transfer FieldMark)
用于表达“把已在场的场标实体从一个持有者迁移到另一个持有者”的通用能力（如同生共死/永恒乐章转移）。
1. 新增 `EffectTransferFieldMark`：
   - `FieldMarkRef` 必填：指定迁移哪一类场标；
   - `FromTargetRef` 必填：指定源范围（常见为 `TargetAllTeammates` / `TargetAllPlayers`）；
   - `Target` 必须可解析为单一目标角色（迁移目的地）。
2. 数量语义：
   - `Value <= 0` 表示迁移源范围内该 `FieldMarkRef` 的全部匹配实体；
   - `Value > 0` 表示最多迁移 `Value` 张，来源不足按可迁移量结算，不阻断后续 Effect。
3. 实体语义：
   - 迁移的是“原场标实体”，不是“移除后新建”；
   - 原始实体的 `CardID / Visibility / Element / Destiny` 等卡面信息保持不变，仅更新 `HolderID`。
4. 顺序约定：
   - 若源范围命中多名持有者，按“进入场上的时间顺序（旧到新）”选取迁移实体，保证可复现。
5. 与既有 Effect 边界：
   - `EffectRemoveFieldMark + EffectPlacePlayedCardAsFieldMark` 表示“移除旧场标 + 放置新实体”；
   - `EffectTransferFieldMark` 表示“同一实体换持有者”。

### 4.6 行动改写配置 (Action Transform Config)
用于表达“视为某行动/改写到另一条结算流水线（Flow）”的规则，避免为单个技能新增耦合型 `EffectType`。

```go
type ActionTransformConfig struct {
    Hook     model.TriggerTiming     // 建议固定为 TimingBeforeActionExecute
    Optional bool                    // true=可选发动；false=满足条件后强制改写
    Priority int                     // 多改写命中时优先级（大者优先）
    CancelCurrentAction bool         // 是否取消当前行动默认结算（如替代 Buy 的原生流程）

    Match   ActionTransformMatch     // 命中条件
    Rewrite *ActionRewriteConfig     // 改写结果；nil 表示仅取消当前行动并继续执行 Skill.Effects
}

type ActionTransformMatch struct {
    RequireActionType         *model.ActionType   // 当前待执行行动类型要求（如 Magic）
    RequirePlayedCardTypes    []model.CardType    // 打出的牌类型要求（可空）
    RequirePlayedCardElements []model.ElementType // 打出的牌系别要求（可空）
    ExcludeTemplateIDs        []string            // 排除模板（如“除暗灭外”）
}

type ActionRewriteConfig struct {
    FlowRef       model.ActionFlowType // 目标流水线（如 ActionFlowMagicBulletChain）
    ActionTypeRef *model.ActionType    // 可选：同时改写 ActionType（不需要时 nil）
    ExecuteImmediately bool            // true=立即按改写后的行动进入结算（不是“+1次行动”）
    TreatAsActiveAttack bool           // 当改写为 Attack 时，是否标记为主动攻击
    ElementPickMode model.RewriteElementPickMode // 改写后行动的元素来源策略
    FixedElementRef *model.ElementType // 当 ElementPickMode=RewriteElementFixed 时使用
}

type RewriteElementPickMode int
const (
    RewriteElementNone          RewriteElementPickMode = 0 // 不改写元素，沿用原行动/原牌元素
    RewriteElementFixed         RewriteElementPickMode = 1 // 使用 FixedElementRef
    RewriteElementFromActionRef RewriteElementPickMode = 2 // 使用 ClientActionRequest.ElementRef
)
```

### 4.7 操作者身份限制 (OperatorIdentity)
用来定义触发事件的“发动者（Operator）”与“技能监听者（Skill Owner）”之间的阵营关系。


```go
// 触发事件的操作者身份限制
type OperatorIdentity int

const (
    OperatorAny      OperatorIdentity = 0 // 任何人（默认，系统不限制）
    OperatorSelf     OperatorIdentity = 1 // 必须是自己
    OperatorTeammateOther OperatorIdentity = 2 // 必须是除自己以外的队友
    OperatorEnemy    OperatorIdentity = 3 // 必须是敌方角色
    OperatorOther    OperatorIdentity = 4 // 必须是除自己以外的任何人（含敌友）
    OperatorTeammate OperatorIdentity = 5 // 必须是己方阵营的人
)
```
---

## 5. 玩家交互指令与中断系统 (Actions & Interrupts)

定义前后端 WebSocket 通信时的标准动作和阻塞状态。

### 5.1 行动类型 (Action Type)
玩家在 `ActionExecution` (行动阶段中) 可以主动下发的合法指令包：
* `Attack`：攻击（一般行动）
* `Magic`：法术（一般行动）
* `Buy`：购买（特殊行动，要求玩家当前手牌数小于等于3）
* `Synthesize`：合成（特殊行动，要求战绩区星石数大于等于3，且玩家当前手牌数小于等于3）
* `Extract`：提炼（特殊行动，要求战绩区星石大于0，且角色能量数未达上限）
* `Deadlock`：宣告无法行动（展示手牌，通过系统校验后清空并重摸）

### 5.1.1 行动流水线类型 (Action Flow Type)
行动类型（`ActionType`）描述“玩家想做什么”，行动流水线（`ActionFlowType`）描述“引擎按哪条规则去结算它”。

| 枚举值 | 标识名 | 描述 |
|:----|:---|:---|
| `0` | `ActionFlowNormalCombat` | 常规攻击/法术结算流水线 |
| `1` | `ActionFlowMagicBulletChain` | 魔弹传递链路结算流水线 |

```go
type ActionFlowType int
const (
    ActionFlowNormalCombat     ActionFlowType = 0
    ActionFlowMagicBulletChain ActionFlowType = 1
)
```

> 运行时约定：`CombatContext` 应暴露 `FlowType model.ActionFlowType`。  
> 技能前置条件可用 `Combat.FlowType == ActionFlowMagicBulletChain` 来匹配“当前是否按魔弹链路结算”，从而兼容“被其他技能改写为魔弹”的场景。
> 若启用“当前战斗改写”能力，`CombatContext` 还应暴露：
> - `OverrideElement *model.ElementType`（由 `EffectSetCurrentCombatElement` 写入）；
> - `CounterExecutorID string`（由 `EffectSetCurrentCounterExecutor` 写入；为空时可回退为 `TargetID` 或既有默认规则）。

### 5.1.2 逐目标遍历来源 (Per Target Source Type)

| 枚举值 | 标识名 | 描述 |
|:----|:---|:---|
| `0` | `PerTargetSelectedTargets` | 使用本次技能在 `TargetRule` 中选中的目标列表 |

```go
type PerTargetSourceType int
const (
    PerTargetSelectedTargets PerTargetSourceType = 0
)
```

### 5.2 中断挂起类型 (Interrupt Type)
当引擎流转到特定节点，需挂起状态机并要求前端弹出 UI 等待玩家操作：
| 标识名 | 触发场景 | 前端交互预期 |
| :--- | :--- | :--- |
| `WaitAction` | 进入行动阶段 | 解锁行动按钮，等待玩家选牌并发出 `ActionType` 指令。 |
| `WaitResponse` | 遭受攻击/法术时 | 弹出防御面板，高亮手中的【圣光】、同系【攻击牌】或【响应技】。 |
| `WaitHeal` | 受到实际伤害时 | 弹出治疗面板，询问玩家是否消耗指示物抵挡伤害。 |
| `WaitDiscard` | 技能代价支付/爆牌时 | 弹出弃牌面板，根据下发的 `CardFilter` (卡牌过滤器) 限制玩家选中合法手牌弃除。 |
| `WaitChoice` | 流程抉择时 | 弹出多选一面板（如：虚弱效果下的“摸3张”或“跳过回合”；各种选角操作）。 |

### 5.3 技能提交扩展字段 (SubmitAction Extensions)
为支持“多目标点数分配”“追加行动二选一”以及“Any 星石选色”等技能，`SubmitAction` 负载增加以下字段：

```go
type ClientActionRequest struct {
    // ... 既有字段省略 ...

    // Key: TargetUserID, Value: 分配值
    // 例：圣疗分配3点治疗给3名角色时：{"u1":1, "u2":1, "u3":1}
    TargetAllocations map[string]int

    // 当技能要求在 Attack / Magic 中二选一时填写
    ActionRef *model.ActionType

    // 当某效果配置 StoneRef=Any 且需要玩家选色时填写（Gem / Crystal）
    StoneRef *model.StarStoneType

    // 当技能/行动改写需要玩家指定元素时填写（如“视为任意系主动攻击”）
    ElementRef *model.ElementType

    // 命名数值输入（支持同一技能同时输入多个变量，如 X/Y）
    // 例：{"X": 3, "Y": 2}
    NamedValues map[string]int
}
```

### 5.3.1 命名数值约束 (Named Value Constraint)
用于技能配置声明 `Action.NamedValues` 的合法输入范围（通用，不与具体技能耦合）。

```go
type NamedValueConstraint struct {
    Key string // 变量名（如 "X", "Y"）
    Required bool // 是否必须由客户端提交该变量
    MinExpression string // 下界表达式（空表示不限制）
    MaxExpression string // 上界表达式（空表示不限制）
}

type TargetRuleConfig struct {
    // ... 既有字段省略 ...
    NamedValueConstraints []NamedValueConstraint
}
```

> 运行时约定：服务端在 `SubmitAction` 校验阶段对每个 `NamedValueConstraint` 执行硬校验；不满足范围或缺失必填项时，直接拒绝本次提交。

### 5.3.2 已选目标聚合查询 (Selected Targets Aggregation)
用于表达“对本次已选目标集合做属性聚合计数”这类条件/数值（如 `Y=目标中治疗>0的人数`）。

```go
type GameStateAPI interface {
    // 统计本次 SubmitAction 已选目标中，当前治疗 >= minHeal 的人数
    // 常见用法：Y = CountSelectedTargetsWithHealAtLeast(Action.Targets, 1)
    CountSelectedTargetsWithHealAtLeast(actionTargets []TargetNode, minHeal int) int
}
```

> 运行时约定：该计数以“本次提交时的目标集合快照”为准；同一结算链内多次读取应保持一致。

### 5.3.3 角色姿态查询 (Player Orientation Query)
用于表达“读取任意角色当前姿态（Normal/Tapped）”这类条件判定（如“仅当当前攻击目标为横置时增伤”）。

```go
type GameStateAPI interface {
    // 查询指定角色当前姿态（Normal/Tapped）
    GetPlayerOrientation(userID string) model.CharacterOrientation
}
```

> 运行时约定：同一战斗结算链内读取应保持一致；当技能以 `Combat.TargetID` 为参数查询时，返回该目标在当前判定时点的实时姿态。

### 5.4 技能作用域士气下降轨迹 (Skill-Scoped Morale Drop Trace)
用于表达“若由某技能造成士气下降，则在后续时点触发额外效果”这类判定。

```go
type SkillMoraleDropTrace struct {
    TurnOwnerID string   // 该轨迹所属回合的回合玩家
    SourceUserID string  // 伤害来源玩家
    SourceSkillID string // 伤害来源技能（普攻/无技能时可为空）
    TargetTeam model.Faction // 实际士气下降的阵营
    Amount int           // 本次实际下降值（>0）
}

// 用于“本回合内对不同目标造成过伤害”这类去重统计判定
type TurnDamageTrace struct {
    TurnOwnerID  string // 该轨迹所属回合的回合玩家
    SourceUserID string // 伤害来源玩家
    SourceSkillID string // 伤害来源技能（普攻/无技能时可为空）
    TargetUserID string // 受伤目标玩家（去重维度）
    IsMagic      bool   // 是否为法术伤害（true=法术；false=攻击）
    Amount       int    // 本次实际生效伤害（>0）
}

type GameStateAPI interface {
    // 查询“本回合由指定玩家的指定技能造成的士气下降总量”
    GetSkillMoraleDropThisTurn(sourceUserID string, sourceSkillID string) int
    // 便捷布尔查询（等价于总量 > 0）
    HasSkillMoraleDropThisTurn(sourceUserID string, sourceSkillID string) bool

    // 查询“本回合由指定来源造成过伤害的去重目标数”
    // onlyMagic=true: 仅统计法术伤害；onlyEnemy=true: 仅统计敌方目标
    CountTurnDistinctDamageTargets(sourceUserID string, onlyMagic bool, onlyEnemy bool) int
    // 便捷布尔查询（等价于 CountTurnDistinctDamageTargets(...) >= threshold）
    HasTurnDistinctDamageTargetsAtLeast(sourceUserID string, onlyMagic bool, onlyEnemy bool, threshold int) bool

    // 令牌/场标辅助查询（用于形态分支与“拥有某场标”判定）
    IsTokenAtCap(userID string, tokenType model.TokenType) bool
    CountTeamFieldMark(sourceUserID string, mark model.CardFieldMark) int
    CountSelectedTargetsWithHealAtLeast(actionTargets []TargetNode, minHeal int) int
    HasExecutedSpecialActionThisTurn(userID string) bool // 本回合是否执行过任意特殊行动（Buy/Synthesize/Extract/Deadlock）
    GetPlayerOrientation(userID string) model.CharacterOrientation // 查询指定角色当前姿态（Normal/Tapped）
    IsSameTeam(userA string, userB string) bool // 判断两名角色是否同阵营
    GetAliveTeammateCount(userID string) int // 获取当前存活队友数（不含自身）
}
```

> 运行时约定（去重聚合）：
> 1. `TurnDamageTrace` 仅在“实际伤害结算后（`TimingOnDamageApplied`）且 `Amount > 0`”时写入；
> 2. 去重键为 `TargetUserID`（同一目标本回合被多次命中只计 1）；
> 3. 查询范围限定在“当前回合（`TurnOwnerID == State.TurnPlayerID`）”。
---

### 5.5 动画时间线事件模型 (Timeline Events)
用于前端按“发生顺序”播放对局动画（谁做了什么、谁响应了什么、结算结果如何），并与 `SyncState` 形成“演出/真值”双通道。

#### 5.5.1 时间线事件类型枚举 (Timeline Event Type)

| 枚举值 | 标识名 | 描述 |
|:----|:---|:---|
| `0` | `TimelinePhaseEntered` | 进入某阶段（Turn/Combat 阶段切换） |
| `1` | `TimelineActionDeclared` | 某玩家提交并通过校验的行动声明（Attack/Magic/Skill/Buy...） |
| `2` | `TimelineActionRejected` | 行动被拒绝（校验失败、强制锁冲突等） |
| `3` | `TimelineResponseWindowOpened` | 响应窗口打开（可响应人、可用选项） |
| `4` | `TimelineResponseSelected` | 某玩家选择了响应项（技能/卡牌/防御） |
| `5` | `TimelineResponseDeclined` | 某玩家主动放弃响应 |
| `6` | `TimelineSkillTriggered` | 技能触发并入执行队列（主动/被动/响应） |
| `7` | `TimelineEffectResolved` | 单个 Effect 节点完成结算（可携带多条 Delta） |
| `8` | `TimelineCombatResolved` | 一次战斗链路收尾（命中/未命中/最终伤害） |
| `9` | `TimelineStatusResolved` | 延后状态结算（中毒/虚弱/封印等） |
| `10` | `TimelineInterruptRaised` | 引擎挂起并请求玩家输入（WaitAction/WaitResponse...） |
| `11` | `TimelineInterruptCleared` | 中断被提交/超时处理后解除 |
| `12` | `TimelineChainClosed` | 当前结算链闭合（用于前端分段收束） |

#### 5.5.2 时间线结果枚举 (Timeline Event Outcome)

| 枚举值 | 标识名 | 描述 |
|:----|:---|:---|
| `0` | `TimelineOutcomeNone` | 无结果语义（仅阶段/窗口通知） |
| `1` | `TimelineOutcomeSuccess` | 成功执行 |
| `2` | `TimelineOutcomeDeclined` | 主动放弃 |
| `3` | `TimelineOutcomeBlocked` | 被阻断/被抵挡 |
| `4` | `TimelineOutcomeMiss` | 未命中 |
| `5` | `TimelineOutcomeTimeout` | 超时按兜底分支处理 |
| `6` | `TimelineOutcomeRejected` | 因合法性校验失败被拒绝 |

#### 5.5.3 时间线可见域枚举 (Timeline Event Visibility)

| 枚举值 | 标识名 | 描述 |
|:----|:---|:---|
| `0` | `TimelineVisibilityPublic` | 全场可见 |
| `1` | `TimelineVisibilityActorOnly` | 仅事件发起者可见 |
| `2` | `TimelineVisibilityActorTeam` | 仅发起者所在阵营可见 |
| `3` | `TimelineVisibilityTargetOnly` | 仅目标可见 |

#### 5.5.4 时间线 Delta 类型枚举 (Timeline Delta Type)

| 枚举值 | 标识名 | 描述 |
|:----|:---|:---|
| `0` | `TimelineDeltaDamage` | 伤害变化（通常为负向生命/士气轨迹） |
| `1` | `TimelineDeltaHeal` | 治疗变化 |
| `2` | `TimelineDeltaMorale` | 士气变化 |
| `3` | `TimelineDeltaHandCount` | 手牌数量变化 |
| `4` | `TimelineDeltaTeamStone` | 阵营战绩区星石变化 |
| `5` | `TimelineDeltaTeamCup` | 阵营星杯变化 |
| `6` | `TimelineDeltaEnergyStone` | 玩家能量区星石变化 |
| `7` | `TimelineDeltaToken` | 指示物变化 |
| `8` | `TimelineDeltaFieldMark` | 场标实体变化（放置/移除/迁移） |
| `9` | `TimelineDeltaStatus` | 状态变化（放置/移除/转移） |
| `10` | `TimelineDeltaOrientation` | 姿态变化（横置/转正） |
| `11` | `TimelineDeltaForm` | 形态变化（进入/退出命名形态） |

```go
type TimelineEventType int
const (
    TimelinePhaseEntered TimelineEventType = 0
    TimelineActionDeclared TimelineEventType = 1
    TimelineActionRejected TimelineEventType = 2
    TimelineResponseWindowOpened TimelineEventType = 3
    TimelineResponseSelected TimelineEventType = 4
    TimelineResponseDeclined TimelineEventType = 5
    TimelineSkillTriggered TimelineEventType = 6
    TimelineEffectResolved TimelineEventType = 7
    TimelineCombatResolved TimelineEventType = 8
    TimelineStatusResolved TimelineEventType = 9
    TimelineInterruptRaised TimelineEventType = 10
    TimelineInterruptCleared TimelineEventType = 11
    TimelineChainClosed TimelineEventType = 12
)

type TimelineEventOutcome int
const (
    TimelineOutcomeNone TimelineEventOutcome = 0
    TimelineOutcomeSuccess TimelineEventOutcome = 1
    TimelineOutcomeDeclined TimelineEventOutcome = 2
    TimelineOutcomeBlocked TimelineEventOutcome = 3
    TimelineOutcomeMiss TimelineEventOutcome = 4
    TimelineOutcomeTimeout TimelineEventOutcome = 5
    TimelineOutcomeRejected TimelineEventOutcome = 6
)

type TimelineEventVisibility int
const (
    TimelineVisibilityPublic TimelineEventVisibility = 0
    TimelineVisibilityActorOnly TimelineEventVisibility = 1
    TimelineVisibilityActorTeam TimelineEventVisibility = 2
    TimelineVisibilityTargetOnly TimelineEventVisibility = 3
)

type TimelineDeltaType int
const (
    TimelineDeltaDamage TimelineDeltaType = 0
    TimelineDeltaHeal TimelineDeltaType = 1
    TimelineDeltaMorale TimelineDeltaType = 2
    TimelineDeltaHandCount TimelineDeltaType = 3
    TimelineDeltaTeamStone TimelineDeltaType = 4
    TimelineDeltaTeamCup TimelineDeltaType = 5
    TimelineDeltaEnergyStone TimelineDeltaType = 6
    TimelineDeltaToken TimelineDeltaType = 7
    TimelineDeltaFieldMark TimelineDeltaType = 8
    TimelineDeltaStatus TimelineDeltaType = 9
    TimelineDeltaOrientation TimelineDeltaType = 10
    TimelineDeltaForm TimelineDeltaType = 11
)

type TimelineDelta struct {
    Type model.TimelineDeltaType
    TargetUserID string
    Value int
    StoneRef *model.StarStoneType
    TokenRef *model.TokenType
    FieldMarkRef *model.CardFieldMark
    StatusRef *model.StatusEffect
}

type TimelineEvent struct {
    EventID int64                     // 房间内单调递增序号
    TurnID int                        // 所属回合号（从 1 开始）
    Phase model.GamePhase             // 事件发生时阶段
    Timing model.TriggerTiming        // 事件对应的触发钩子
    ChainID string                    // 同一结算链路标识
    ParentEventID *int64              // 可选：父事件（如响应来源）

    Type model.TimelineEventType
    Outcome model.TimelineEventOutcome
    Visibility model.TimelineEventVisibility

    ActorUserID string
    TargetUserIDs []string
    ActionType *model.ActionType
    SkillID *string
    CardIDs []string

    Deltas []model.TimelineDelta
    Message string
}
```

#### 5.5.5 发射规范 (Emission Spec)
1. **顺序保证**：同一房间内 `TimelineEvent.EventID` 必须严格单调递增；前端按 `EventID` 排序播放。
2. **链路分段**：同一结算链路必须共享 `ChainID`，并以 `TimelineChainClosed` 结尾，方便前端做“整段动画收束”。
3. **窗口可视化**：每次出现交互挂起都应先发 `TimelineInterruptRaised`/`TimelineResponseWindowOpened`，结束时发 `TimelineInterruptCleared`。
4. **响应可追溯**：玩家选择响应时必须发 `TimelineResponseSelected`；若放弃/超时则发 `TimelineResponseDeclined` 或 `Outcome=TimelineOutcomeTimeout`。
5. **效果可解释**：`TimelineEffectResolved` 应附带 `Deltas`；若该节点无状态变化，也应发事件并用 `Outcome` 表明结果。
6. **战争迷雾**：私密信息（手牌 UUID、暗置牌信息）通过 `Visibility` 控制可见范围，禁止向非授权客户端广播。
7. **与真值同步关系**：Timeline 用于动画演出，不替代 `SyncState`；关键结算段结束后仍需下发状态真值快照用于纠偏。
