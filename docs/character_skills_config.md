# 《星杯传说》角色技能配置说明文档 (Skill Configurations)

**文档说明**：本文档基于 `data_model.md` 与 `static_classes_dictionary.md` 定义的数据模型，将 `character.md` 中的自然语言技能文案拆解、提炼为面向后端开发（如 Go 结构体实例化）的强类型属性配置说明。

---

## 📦 角色技能配置列表 (第一批次)

### 00. 基础法术牌 (Common Base Magic Cards)

#### 【圣光】 (Holy Light)
* **技能描述**：抵挡一次攻击或【魔弹】。
* **1. 主干配置**：
  * **SkillID**: `skill_common_holy_light`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnHitCheck`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatHitCheck]`
  * **CustomExpression**: `Combat.TargetID == Self.UserID && (Combat.IsAttack == true || Combat.IsMagicBullet == true)`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectCancelHit`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `0`
    * `Visibility`: `VisibilityPublic`

#### 【中毒】 (Poison)
* **技能描述**：（将此牌放置于目标角色前，他的行动阶段开始前）对他造成1点法术伤害③。
* **1. 主干配置**：
  * **SkillID**: `skill_common_poison`
  * **Category**: `Normal`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `None`
  * **StackingRule**: StackingUnique *(同一名角色面前只允许存在一个中毒)*
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **Filters**: `Target.HasStatus(Poison) == false` *(同一名角色面前同时只能有一个中毒)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**
  * **Effect[0]**:
    * `EffectType`: `EffectPlaceStatus`
    * `Target`: `TargetSelected`
    * `Value`: `1`
    * `Ref`: `Status_Poison`
* **7. 状态结算行为**（挂载于 StatusEffect.Poison，用 Timing + Effects 表达）：
  * **ResolveTiming**: `TimingOnBeforeAction`
  * **RequireHolderIsTurnPlayer**: `true`
  * **ResolveMode**: `Auto`
  * **ResolveEffects**: `[{EffectType: EffectDamage, Target: TargetSelf, Value: 1, Ref: nil}]`
  * **RemoveAfterResolve**: `true`
  * **TriggerLimit**: `1`

#### 【虚弱】 (Weakness)
* **技能描述**：（将此牌放置于目标角色前，他的行动阶段开始前）目标选择：跳过行动阶段，或摸3张牌后继续行动阶段。
* **1. 主干配置**：
  * **SkillID**: `skill_common_weakness`
  * **Category**: `Normal`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `None`
  * **StackingRule**: StackingUnique *(同一名角色面前只允许存在一个中毒)*
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **Filters**: `Target.HasStatus(Weakness) == false` *(同一名角色面前同时只能有一个虚弱)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectPlaceStatus`
    * `Target`: `TargetSelected`
    * `Value`: `1`
    * `Ref`: `Status_Weakness`
* **7. 状态结算行为**（挂载于 StatusEffect.Weakness，用 Timing + Effects 表达）：
  * **ResolveTiming**: `TimingOnBeforeAction`
  * **RequireHolderIsTurnPlayer**: `true`
  * **ResolveMode**: `Choice`
  * **ResolveChoiceOptions**:
    * `[0]`: `{Effects: [{EffectType: EffectSkipPhase, Target: TargetSelf, Value: 0, ActionRef: nil}]}` *(跳过行动阶段)*
    * `[1]`: `{Effects: [{EffectType: EffectDrawCard, Target: TargetSelf, Value: 3, Ref: nil}]}` *(摸3张牌后继续)*
  * **ResolveChoiceTimeoutIndex**: `0`
  * **RemoveAfterResolve**: `true`
  * **TriggerLimit**: `1`

#### 【圣盾】 (Holy Shield)
* **技能描述**：（将此牌放置于目标角色前，他遭受攻击或【魔弹】时，移除此牌）视为未命中。
* **1. 主干配置**：
  * **SkillID**: `skill_common_holy_shield`
  * **Category**: `Normal`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `None`
  * **StackingRule**: StackingUnique *(同一名角色面前只允许存在一个圣盾)*
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **Filters**: `Target.HasStatus(HolyShield) == false` *(同一名角色面前同时只能有一个圣盾)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectPlaceStatus`
    * `Target`: `TargetSelected`
    * `Value`: `1`
    * `Ref`: `Status_HolyShield`
* **7. 状态结算行为**（挂载于 StatusEffect.HolyShield，用 Timing + Effects 表达）：
  * **ResolveTiming**: `TimingOnHitCheck`
  * **RequireHolderIsCombatTarget**: `true`
  * **ResolveMode**: `Auto`
  * **CanDecline**: `false`
  * **ResolveEffects**: `[{EffectType: EffectCancelHit, Target: TargetCurrentCombat, Value: 0, Ref: nil}]`
  * **RemoveAfterResolve**: `true`
  * **TriggerLimit**: `1`

#### 【魔弹】 (Magic Bullet)
* **技能描述**：（使用此牌）你右手边最近的一名对手选择：承受伤害，打出【魔弹】继续传递并伤害+1，或用【圣光】/【圣盾】抵挡。
* **1. 主干配置**：
  * **SkillID**: `skill_common_magic_bullet`
  * **Category**: `Normal`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `State.HasRightNearestEnemy(Self.UserID) == true`
* **3. 目标选择规则**：
  * **SelectType**: `None` *(首个目标由系统按“右手边最近对手”自动确定)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectNone`
    * `Target`: `TargetCurrentEvent`
    * `Value`: `0`
    * `Ref`: `SystemOp_StartMagicBulletChain(baseDamage=2, damageStep=1)` *(魔弹链路由系统专用结算器处理：可传递、可被圣光/圣盾终止、同轮去重参与者)*

### 01. 天使 (Angel)

#### 【天使羁绊】 (Angel's Bond)
* **技能描述**：（每当你移除一个基础效果或是使用［圣盾］时）目标角色+1［治疗］。
* **1. 主干配置**：
  * **SkillID**: `skill_angel_bond`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnFieldMarkChanged`
* **2. 前置条件**：
  * **PhaseLimit**: `None`
  * **RequireOrientation**: `None`
  * **HandLimit**: `None`
  * **RequireFieldMark**: `None`
  * **IsTurnLimited**: `false`
  * **RequireSourceType**: `SourcePlayer` (防 Bug 核心：必须是玩家主动造成的变动，过滤掉系统回合结束自然结算的移除)
  * **RequireOperator**: `OperatorSelf` (阵营溯源：这个主动行为的肇事者，必须是天使本人)
  * **CustomEventFilter**: `(MarkAction == 'Placed' && MarkType == 'HolyShield') OR (MarkAction == 'Removed')` (逻辑分支：只监听放置圣盾，或是移除任意基础效果)
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **Filters**: `None`
  * **【架构批注】**: 因为是 Passive 技能且 MinCount > 0，引擎在此处必定会触发 FSM 中断（WaitTarget），向玩家索要目标后，才会继续向下执行 Effect。
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * `None` (被动触发，无额外消耗)
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectHeal`
    * `Target`: `TargetSelected` (此处提取的是刚刚中断索要到的目标，而非触发事件时的原始目标)
    * `Value`: `1`
    * `Ref`: `None`

#### 【天使祝福】 (Angel's Blessing)
* **技能描述**：（弃1张水系牌［展示］）指定目标玩家给你2张牌或指定2名角色各给你1张牌。
* **1. 主干配置**：
  * **SkillID**: `skill_angel_blessing`
  * **Category**: `Normal`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **IsTurnLimited**: `false`
  * **CustomExpression**: `None`
  * **RequireOperator**: `None` (【架构批注】：主动发动的技能不需要溯源监听，因为操作者必定是玩家本人)
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `1`
  * **MaxCount**: `2`
  * **Filters**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Discards**:
    * `Filter`: {ReqElement: Water} (要求水系)
    * `Count`: 1
    * `Visibility`: `VisibilityPublic` (必须展示牌面)
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectTransferCard` *(注: 将选定目标的牌转移给自身)*
    * `Target`: `TargetSelected`
    * `Value`: `Action.SelectedTargetCount == 1 ? 2 : 1` *(如果选了1人则给2张，选了2人各给1张)*
    * `Ref`: `None`
    * `Visibility`: `VisibilityHidden` (玩家给你牌的时候是暗给的，不需要向全场展示具体给了什么牌)

#### 【风之洁净】 (Wind Cleansing)
* **技能描述**：（弃1张风系牌［展示］）移除场上任意1个基础效果。
* **1. 主干配置**：
  * **SkillID**: `skill_angel_wind_cleansing`
  * **Category**: `Normal`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `State.HasAnyStatusOnField() == true`
  * **RequireOperator**: `None` (【架构批注】：主动发动的技能不需要溯源监听，因为操作者必定是玩家本人)
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **Filters**: `RequireStatus != None`
  * **SubSelect **：
  * `SubType`: `FieldMark`
  * `SubMinCount`: `1`
  * `SubMaxCount`: `1`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Discards**: `Count: 1, Filter: {ReqElement: Wind}`
  * * `Visibility`: `VisibilityPublic` (指定为全场展示)
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectRemoveStatus`
    * `Target`: `TargetSelected`
    * `Value`: `1`
    * `StatusRef`: `Action.Targets[0].SelectedFieldCards[0]` （通过复合目标树的通用路径，精准提取玩家在二级菜单里选中的具体选择的基础效果对应的手牌)
    * `Visibility`: `VisibilityPublic`

#### 【天使之歌】 (Angel's Song)
* **技能描述**：［回合限定］［水晶］（在你的回合开始前发动）移除场上任意1个基础效果。
* **1. 主干配置**：
  * **SkillID**: `skill_angel_song`
  * **Category**: `Ultimate`
  * **Type**: `Response`
  * **Timing**: `TimingOnTurnStart` *(对应"回合开始前")*
* **2. 前置条件**：
  * **PhaseLimit**: `[TurnBeforeStart]`
  * **IsTurnLimited**: `true`
  * **CustomExpression**: `State.HasAnyStatusOnField() == true`
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **RequireStatus**: `null`
  * **SubSelect **：
  * `SubType`: `FieldCard`
  * `SubFilter`: `FieldCard.StatusMeta != nil && FieldCard.StatusMeta.Class == Basic`
  * `SubMinCount`: `1`
  * `SubMaxCount`: `1`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Crystal, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectRemoveStatus`
    * `Target`: `TargetSelected`
    * `Value`: `1`
    * `RefKind`: `FieldCardID`
    * `Ref`: `Action.Targets[0].SelectedFieldCards[0]`
    * `Visibility`: `VisibilityPublic` (法术大招明牌结算，全场可见)

#### 【神之庇护】 (Divine Protection)
* **技能描述**：X个［水晶］为我方抵御X点因法术伤害而造成的士气下降。
* **1. 主干配置**：
  * **SkillID**: `skill_angel_divine_protection`
  * **Category**: `Ultimate`
  * **Type**: `Response`
  * **Timing**: `TimingOnDamageTaken`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDraw]` *(法术伤害结算摸牌/爆牌判定士气时)*
  * **CustomExpression**: `Combat.IsMagic == true && Event.PendingMoraleLoss > 0`
* **3. 目标选择规则**：
  * **SelectType**: `None`
  * **SelectedValueRule**: `1 <= Action.SelectedValue <= Event.PendingMoraleLoss`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Crystal, Amount: Action.SelectedValue}]` *(动态消耗玩家选取的X值；由 SelectedValueRule 限制上限)*
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None` *(该技能使用数值抵消，不使用全免型 Tag，避免部分抵御被误实现为全部抵御)*
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectReducePendingMoraleLoss`
    * `Target`: `TargetCurrentEvent`
    * `Value`: `Action.SelectedValue` *(将待扣士气减少 X；引擎内部执行 `PendingMoraleLoss = Max(0, PendingMoraleLoss - Value)`)*
    * `Ref`: `None`

---

### 02. 狂战士 (Berserker)

#### 【狂化】 (Berserk)
* **技能描述**：你发动的所有攻击伤害额外+1。（攻击命中时②，若你的手牌>3）本次攻击伤害额外+1。
* **1. 主干配置**：
  * **SkillID**: `skill_berserker_berserk`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnAttackDeclared` / `TimingOnHitCheck`
* **2. 前置条件**：
  * **PhaseLimit**: `None`
  * **CustomExpression**: `Combat.SourceID == Self.UserID` *(攻击来源是自己即可，覆盖主动攻击与应战攻击)*
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(对应"所有攻击伤害额外+1")*:
    * `EffectType`: `EffectAttackDamageModifier`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `1`
    * `Condition`: `Event.TriggerHook == TimingOnAttackDeclared`
  * **Effect[1]** *(对应"命中且手牌>3额外+1")*:
    * `EffectType`: `EffectAttackDamageModifier`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `1`
    * `Condition`: `Event.TriggerHook == TimingOnHitCheck && Self.HandCount > 3 && Combat.IsHit == true`

#### 【撕裂】 (Tear)
* **技能描述**：［宝石］攻击命中后发动②，本次攻击伤害额外+2。
* **1. 主干配置**：
  * **SkillID**: `skill_berserker_tear`
  * **Category**: `Ultimate`
  * **Type**: `Response`
  * **Timing**: `TimingOnHitCheck`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatHitCheck]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsHit == true` *(攻击命中即可，覆盖主动攻击与应战攻击)*
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Gem, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectAttackDamageModifier`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `2`
    * `Ref`: `None`

#### 【血腥咆哮】 (Bloody Roar)
* **技能描述**：作为主动攻击打出时发动，若攻击的目标拥有的［治疗］为2，则本次攻击强制命中。
* **1. 主干配置**：
  * **SkillID**: `skill_berserker_bloody_roar`
  * **Category**: `Unique`
  * **Type**: `Response`
  * **Timing**: `TimingOnAttackDeclared`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDeclare]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsActiveAttack == true && Combat.AttackCard.CharacterSkillMap[Self.CharacterID] == "skill_berserker_bloody_roar" && Target.Heal == 2`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `ForceHit`
* **6. 执行效果序列**：
  * `EffectNone` (通过 Tag 劫持系统状态)

#### 【血影狂刀】 (Blood Shadow Blade)
* **技能描述**：作为主动攻击打出时发动●若命中后②对手的手牌为2，本次攻击伤害额外+2。●若命中后②对手的手牌为3，本次攻击伤害额外+1。
* **1. 主干配置**：
  * **SkillID**: `skill_berserker_blood_blade`
  * **Category**: `Unique`
  * **Type**: `Response`
  * **Timing**: `TimingOnHitCheck`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatHitCheck]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsActiveAttack == true && Combat.AttackCard.CharacterSkillMap[Self.CharacterID] == "skill_berserker_blood_blade" && Combat.IsHit == true`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectAttackDamageModifier`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `Target.HandCount == 2 ? 2 : (Target.HandCount == 3 ? 1 : 0)`
    * `Ref`: `None`

---

### 03. 封印师 (Sealer)

#### 【法术激荡】 (Spell Surge)
* **技能描述**：（［法术行动］结束时发动）额外+1［攻击行动］。
* **1. 主干配置**：
  * **SkillID**: `skill_sealer_spell_surge`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnActionEnd`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionEnd]`
  * **CustomExpression**: `Action.CurrentType == Magic`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `Implementation`: `model.AppendExtraAction`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `ActionRef`: `Attack`     // 限定追加的行动大类为：攻击行动
    * `ElementRef`: `None`  
    * `StatusRef`: `None`
    * `TokenRef`: `None`

#### 【封印破碎】 (Seal Break)
* **技能描述**：［水晶］将场上任意一张基础效果牌收入自己手中。
* **1. 主干配置**：
  * **SkillID**: `skill_sealer_seal_break`
  * **Category**: `Ultimate`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `State.HasAnyStatusOnField() == true`
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **Filters**: `RequireStatus != None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Crystal, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectRemoveStatusToHand` *(注: 将基础效果回收到手牌)*
    * `Target`: `TargetSelected`
    * `Value`: `1`
    * `Ref`: `Target.SelectedStatus`

#### 【五系束缚】 (Five Elements Bind)
* **技能描述**：［水晶］将五系束缚放置于目标对手面前，该对手跳过其下个行动阶段。在其下个行动阶段开始前他可以选择摸（2+X）张牌来取消五系束缚的效果。X为场上封印的数量，X最高为2。不论效果是否发动，触发后移除此牌。
* **1. 主干配置**：
  * **SkillID**: `skill_sealer_five_elements_bind`
  * **Category**: `Exclusive`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `None`
* **3. 目标选择规则**：
  * **SelectType**: `Enemy`
  * **MinCount**: `1`
  * **MaxCount**: `1`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayRequired`
  * **Stones**: `[{Type: Crystal, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectPlaceStatus`
    * `Target`: `TargetSelected`
    * `Value`: `1`
    * `Ref`: `Status_FiveElementsBind`
* **7. 状态结算行为**（挂载于 StatusEffect.FiveElementsBind，用 Timing + Effects 表达）：
  * **ResolveTiming**: `TimingOnBeforeAction`
  * **RequireHolderIsTurnPlayer**: `true`
  * **ResolveMode**: `Choice`
  * **ResolveChoiceOptions**:
    * `[0]`: `{ChoiceID: "draw_cancel", Label: "摸牌取消", Effects: [{EffectType: EffectDrawCard, Target: TargetSelf, Value: 2, Ref: nil}]}` *(Value 2+X，X=Min(2, 场上封印数)，由引擎计算)*
    * `[1]`: `{ChoiceID: "skip_action", Label: "跳过行动阶段", Effects: [{EffectType: EffectSkipPhase, Target: TargetSelf, Value: 0, Ref: nil}]}`
  * **ResolveChoiceTimeoutIndex**: `1`
  * **RemoveAfterResolve**: `true`
  * **TriggerLimit**: `1`

#### 【水/火/地/风/雷之封印】 (Elemental Seals)
* **技能描述**：（将对应封印放置于目标对手面前）该对手获得（直到他从手中打出或展示出对应系别牌时强制触发）对他造成3点法术伤害③，触发后移除此牌。
* **1. 主干配置**：
  * **SkillID**: `skill_sealer_elemental_seal` *(可统一处理或分系别配置)*
  * **Category**: `Unique`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Action.PlayedCard.CharacterSkillMap[Self.CharacterID] == "skill_sealer_elemental_seal"` *(由当前打出的封印牌映射到该技能)*
* **3. 目标选择规则**：
  * **SelectType**: `Enemy`
  * **MinCount**: `1`
  * **MaxCount**: `1`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectPlaceStatus`
    * `Target`: `TargetSelected`
    * `Value`: `1`
    * `StatusRef`: `ElementalSeal`
    * `ElementRef`: `Action.PlayedCard.Element` *(把当前打出的封印牌系别写入 StatusMeta.BoundElement；水/火/地/风/雷共享同一套结算配置)*
* **7. 状态结算行为**（挂载于 StatusEffect.ElementalSeal，用 Timing + Effects 表达）：
  * **ResolveTiming**: `TimingOnCardPlayedOrRevealed`
  * **PlayedCardElementMatchType**: `MatchPlayedCardElementStatusBoundElement`
  * **RequiredPlayedCardElement**: `nil` *(仅 MatchType=MatchPlayedCardElementFixed 时需要填写；本技能使用状态绑定系别，不用固定值)*
  * **ResolveMode**: `Auto`
  * **ResolveEffects**: `[{EffectType: EffectDamage, Target: TargetSelf, Value: 3, Ref: nil}]`
  * **RemoveAfterResolve**: `true`
  * **TriggerLimit**: `1`

### 04. 风之剑圣 (Wind Sword Saint)

#### 【风怒追击】 (Wind Fury Chase)
* **技能描述**：［回合限定］（［攻击行动］结束时发动）额外+1风系［攻击行动］。
* **1. 主干配置**：
  * **SkillID**: `skill_sword_saint_wind_fury_chase`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnActionEnd`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionEnd]`
  * **IsTurnLimited**: `true`
  * **CustomExpression**: `Action.CurrentType == Attack && Action.SourceID == Self.UserID`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `Implementation`: `model.AppendExtraAction`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `ActionRef`: `Attack`     // 限定追加的行动大类为：攻击行动
    * `ElementRef`: `Wind`      // 限定该行动必须打出的牌系别为：风系
    * `StatusRef`: `None`
    * `TokenRef`: `None`

#### 【圣剑】 (Holy Sword)
* **技能描述**：若你的主动攻击为你本次行动阶段的第三次［攻击行动］，则此攻击强制命中。本次［攻击行动］结束后，你摸X张牌，弃X张牌（X<4）。
* **1. 主干配置**：
  * **SkillID**: `skill_sword_saint_holy_sword`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnAttackDeclared` / `TimingOnActionEnd` *(前者做命中劫持，后者做延后结算)*
* **2. 前置条件**：
  * **PhaseLimit**: `None`
  * **CustomExpression**: `(Event.TriggerHook == TimingOnAttackDeclared && Combat.SourceID == Self.UserID && Combat.IsActiveAttack == true && Self.AttackCount == 3) || (Event.TriggerHook == TimingOnActionEnd && Action.SourceID == Self.UserID && Action.CurrentType == Attack && Self.AttackCount == 3)`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(第三次主动攻击宣告时强制命中)*:
    * `EffectType`: `EffectApplyCombatTag`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `ForceHit`
    * `Condition`: `Event.TriggerHook == TimingOnAttackDeclared`
* **7. 延后结算配置**：
  * **ResolveMoment**: `TimingOnActionEnd`
  * **InterruptType**: `WaitChoice` *(在攻击行动结束时弹出 X 值选择，而不是在宣告攻击时选择)*
  * **SelectedValueRule**: `0 <= Action.SelectedValue && Action.SelectedValue < 4`
  * **ResolveOrder**: `ChooseX -> ResolveEffects[0] -> ResolveEffects[1]` *(先摸后弃，保持原文案顺序)*
  * **ResolveEffects[0]**:
    * `EffectType`: `EffectDrawCard`
    * `Target`: `TargetSelf`
    * `Value`: `Action.SelectedValue`
  * **ResolveEffects[1]**:
    * `EffectType`: `EffectDiscard`
    * `Target`: `TargetSelf`
    * `Value`: `Action.SelectedValue`
  * **DataFlow**: `Action.SelectedValue` *(由该次延后结算窗口写入，供 ResolveEffects 读取)*

#### 【剑影】 (Sword Shadow)
* **技能描述**：［回合限定］［水晶］（［攻击行动］结束时发动）额外+1［攻击行动］。
* **1. 主干配置**：
  * **SkillID**: `skill_sword_saint_sword_shadow`
  * **Category**: `Ultimate`
  * **Type**: `Response`
  * **Timing**: `TimingOnActionEnd`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionEnd]`
  * **IsTurnLimited**: `true`
  * **CustomExpression**: `Action.CurrentType == Attack && Action.SourceID == Self.UserID`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Crystal, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `Implementation`: `model.AppendExtraAction`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `ActionRef`: `Attack`     // 限定追加的行动大类为：攻击行动
    * `ElementRef`: `None` 
    * `StatusRef`: `None`
    * `TokenRef`: `None`

#### 【疾风技】 (Gale Technique)
* **技能描述**：（作为主动攻击打出时发动）额外+1［攻击行动］。
* **1. 主干配置**：
  * **SkillID**: `skill_sword_saint_gale_technique`
  * **Category**: `Unique`
  * **Type**: `Response`
  * **Timing**: `TimingOnAttackDeclared`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDeclare]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsActiveAttack == true && Combat.AttackCard.CharacterSkillMap[Self.CharacterID] == "skill_sword_saint_gale_technique"`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `Implementation`: `model.AppendExtraAction`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `ActionRef`: `Attack`     // 限定追加的行动大类为：攻击行动
    * `ElementRef`: `None`
    * `StatusRef`: `None`
    * `TokenRef`: `None`

#### 【列风技】 (Raging Wind Technique)
* **技能描述**：（攻击目标拥有圣盾时发动）无视对手圣盾的效果，且此攻击对手无法应战。
* **1. 主干配置**：
  * **SkillID**: `skill_sword_saint_raging_wind_technique`
  * **Category**: `Unique`
  * **Type**: `Response`
  * **Timing**: `TimingOnAttackDeclared`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDeclare]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsActiveAttack == true && Combat.AttackCard.CharacterSkillMap[Self.CharacterID] == "skill_sword_saint_raging_wind_technique" && Target.HasStatus(HolyShield) == true`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `Unrespondable, IgnoreHolyShield`
* **6. 执行效果序列**：
  * `EffectNone` (通过 Tag 劫持本次 CombatContext)

### 05. 神箭手 (Sharpshooter)

#### 【贯穿射击】 (Piercing Shot)
* **技能描述**：（主动攻击未命中时发动②，弃1张法术牌［展示］）对你所攻击的目标造成2点法术伤害③。
* **1. 主干配置**：
  * **SkillID**: `skill_sharpshooter_piercing_shot`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnHitCheck`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatHitCheck]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsActiveAttack == true && Combat.IsHit == false`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Discards**: `Count: 1, Filter: {ReqCardType: Magic}`
  * **DiscardsVisibility**: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `2`
    * `Ref`: `Combat.TargetID` *(固定结算到本次攻击原目标)*

#### 【闪电箭】 (Lightning Arrow)
* **技能描述**：你的雷系攻击对手无法应战。
* **1. 主干配置**：
  * **SkillID**: `skill_sharpshooter_lightning_arrow`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnAttackDeclared`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDeclare]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.AttackCard.Element == Thunder`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `Unrespondable`
* **6. 执行效果序列**：
  * `EffectNone` (通过 Tag 劫持本次 CombatContext)

#### 【狙击】 (Snipe)
* **技能描述**：［水晶］目标角色手牌补到5张［强制］，额外+1［攻击行动］。
* **1. 主干配置**：
  * **SkillID**: `skill_sharpshooter_snipe`
  * **Category**: `Ultimate`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `None`
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `1`
  * **MaxCount**: `1`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Crystal, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(补到5张，不超抽)*:
    * `EffectType`: `EffectDrawCard`
    * `Target`: `TargetSelected`
    * `Value`: `Max(0, 5 - Target.HandCount)`
    * `Ref`: `None`
  * **Effect[1]**:
    * `Implementation`: `model.AppendExtraAction`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `ActionRef`: `Attack`     // 限定追加的行动大类为：攻击行动
    * `ElementRef`: `None`
    * `StatusRef`: `None`
    * `TokenRef`: `None`

#### 【精准射击】 (Precision Shot)
* **技能描述**：此攻击强制命中，但本次攻击伤害-1。
* **1. 主干配置**：
  * **SkillID**: `skill_sharpshooter_precision_shot`
  * **Category**: `Unique`
  * **Type**: `Response`
  * **Timing**: `TimingOnAttackDeclared`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDeclare]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsActiveAttack == true && Combat.AttackCard.CharacterSkillMap[Self.CharacterID] == "skill_sharpshooter_precision_shot"`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `ForceHit`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectAttackDamageModifier`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `-1`
    * `Ref`: `None`

#### 【闪光陷阱】 (Flash Trap)
* **技能描述**：对目标角色造成2点法术伤害③。
* **1. 主干配置**：
  * **SkillID**: `skill_sharpshooter_flash_trap`
  * **Category**: `Unique`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Action.PlayedCard.CharacterSkillMap[Self.CharacterID] == "skill_sharpshooter_flash_trap"`
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `1`
  * **MaxCount**: `1`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelected`
    * `Value`: `2`
    * `Ref`: `None`

### 06. 暗杀者 (Assassin)

#### 【反噬】 (Backlash)
* **技能描述**：（承受攻击伤害时发动⑥）攻击你的对手摸1张牌［强制］。
* **1. 主干配置**：
  * **SkillID**: `skill_assassin_backlash`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnDamageTaken`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDraw]`
  * **CustomExpression**: `Combat.TargetID == Self.UserID && Combat.IsAttack == true && Combat.FinalDamage > 0`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectDrawCard`
    * `Target`: `TargetTriggerSource` *(触发该次攻击伤害的来源玩家)*
    * `Value`: `1`
    * `Ref`: `None`

#### 【水影】 (Water Shadow)
* **技能描述**：（除［特殊行动］外，当你摸牌前发动）弃X张水系牌（展示）；（若你处于［潜行］效果下）你可额外弃1张法术牌（展示）。
* **1. 主干配置**：
  * **SkillID**: `skill_assassin_water_shadow`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingBeforeCardDrawn`
* **2. 前置条件**：
  * **PhaseLimit**: `None`
  * **CustomExpression**: `Event.DrawTargetID == Self.UserID && Action.CurrentType != Special`
* **3. 目标选择规则**：
  * **SelectType**: `None`
  * **SelectedValueRule**: `Action.SelectedValue >= 1`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Discards**: `Count: Action.SelectedValue, Filter: {ReqElement: Water}`
  * **DiscardsVisibility**: `VisibilityPublic`
  * **OptionalDiscards**: `Count: 1, Filter: {ReqCardType: Magic}, Condition: Self.Form == "shadow"`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectNone`
    * `Target`: `TargetCurrentEvent`
    * `Value`: `0`
    * `Ref`: `SystemOp_ReplaceCurrentDrawWithDiscard(Action.SelectedValue)` *(摸牌前拦截点：将本次摸牌劫持为弃牌逻辑)*

#### 【潜行】 (Stealth)
* **技能描述**：［宝石］你可选择摸1张牌，［横置］持续到你的下个行动阶段开始，你的手牌上限-1；你不能成为主动攻击的目标；你的主动攻击对方无法应战且伤害额外+X，X为你剩余的能量数。潜行的效果结束时角色［转正］。
* **1. 主干配置**：
  * **SkillID**: `skill_assassin_stealth`
  * **Category**: `Ultimate`
  * **Type**: `Startup`
  * **Timing**: `TimingStartup`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionStart]`
  * **CustomExpression**: `Self.Orientation == Normal`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Gem, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(可选摸1)*:
    * `EffectType`: `EffectDrawCard`
    * `Target`: `TargetSelf`
    * `Value`: `Action.ChooseDraw ? 1 : 0`
    * `Ref`: `None`
  * **Effect[1]** *(进入潜行形态并横置，持续到下个 ActionStart)*:
    * `EffectType`: `EffectNone`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `Ref`: `SystemOp_EnterForm(name=\"shadow\", orientation=Tapped, until=NextActionStart, handLimitDelta=-1, cannotBeTargetedByActiveAttack=true)`
  * **Effect[2]** *(潜行期间主动攻击强化：不可应战 + 伤害额外+剩余能量)*:
    * `EffectType`: `EffectNone`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `Ref`: `SystemOp_GrantWhileForm(\"shadow\", onActiveAttack=[Unrespondable, DamagePlus(RemainingEnergy)])`

---

## 暗杀者以下角色 TodoList (按 character.md 顺序)

- [ ] 07. 圣女（完成）
- [ ] 08. 魔法少女（完成）
- [ ] 09. 女武神（完成）
- [ ] 10. 元素师（完成）
- [ ] 11. 仲裁者（完成）
- [ ] 12. 冒险家（完成）
- [ ] 13. 圣枪骑士（完成）
- [ ] 14. 精灵射手（完成）
- [ ] 15. 瘟疫法师（完成）
- [ ] 16. 魔剑士（完成）
- [ ] 17. 血色剑灵（完成）
- [ ] 18. 祈祷师（完成）
- [ ] 19. 红莲骑士（完成）
- [ ] 20. 英灵人形（完成）
- [ ] 21. 神官（完成）
- [ ] 22. 阴阳师（完成）
- [ ] 23. 苍炎魔女（完成）
- [ ] 24. 贤者（完成）
- [ ] 25. 魔弓（完成）
- [ ] 26. 魔枪（完成）
- [ ] 27. 灵符师（完成）
- [ ] 28. 吟游诗人（完成）
- [ ] 29. 勇者（完成）
- [ ] 30. 格斗家（完成）
- [ ] 31. 圣弓（完成）
- [ ] 32. 剑帝（完成）
- [ ] 33. 兽灵武士（完成）
- [ ] 34. 灵魂术士（完成）
- [ ] 35. 月之女神（完成）
- [ ] 36. 血之巫女（完成）
- [ ] 37. 蝶舞者（完成）
- [ ] 38. 圣殿骑士
- [ ] 39. 圣庭检察士
- [ ] 40. 战斗法师
- [ ] 41. 星坠女巫
- [ ] 42. 猎巫人
- [ ] 43. 灵熙之潮
- [ ] 44. 剑之魔女
- [ ] 45. 节日魔导
- [ ] 46. 游击士
- [ ] 47. 暴食少女
- [ ] 48. 矜贵之女
- [ ] 49. 噬神者
- [ ] 50. 女仆长
- [ ] 51. 结界师
- [ ] 52. 神秘学者
- [ ] 53. 染污者
- [ ] 54. 咒符师
- [ ] 55. 原初之弓
- [ ] 56. 贪婪少女
- [ ] 57. 捣蛋萝莉
- [ ] 58. 圣弓（trick!）
- [ ] 59. 冒险家（trick!）
- [ ] 60. 红衣主教
- [ ] 61. 铸律者
- [ ] 62. 记录者
- [ ] 63. 传教士
- [ ] 64. 异教徒
- [ ] 65. 萝莉番长
- [ ] 66. 见习制片

### 07. 圣女 (Saintess)

#### 【冰霜祷言】 (Frost Prayer)
* **技能描述**：（每当你使用水系牌或圣光时发动）目标角色+1［治疗］。
* **1. 主干配置**：
  * **SkillID**: `skill_saintess_frost_prayer`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnCardPlayedOrRevealed`
* **2. 前置条件**：
  * **PhaseLimit**: `None`
  * **CustomExpression**: `Event.OperatorID == Self.UserID && (Action.PlayedCard.Element == Water || Action.PlayedCard.TemplateID == "card_common_holy_light")`
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `1`
  * **MaxCount**: `1`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectHeal`
    * `Target`: `TargetSelected`
    * `Value`: `1`
    * `Ref`: `None`

#### 【治愈之光】 (Healing Light)
* **技能描述**：指定最多3名角色各+1［治疗］。
* **1. 主干配置**：
  * **SkillID**: `skill_saintess_healing_light`
  * **Category**: `Unique`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Action.PlayedCard.CharacterSkillMap[Self.CharacterID] == "skill_saintess_healing_light"`
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `1`
  * **MaxCount**: `3`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectHeal`
    * `Target`: `TargetSelected`
    * `Value`: `1`
    * `Ref`: `None`

#### 【治疗术】 (Cure)
* **技能描述**：目标角色+2［治疗］。
* **1. 主干配置**：
  * **SkillID**: `skill_saintess_cure`
  * **Category**: `Unique`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Action.PlayedCard.CharacterSkillMap[Self.CharacterID] == "skill_saintess_cure"`
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `1`
  * **MaxCount**: `1`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectHeal`
    * `Target`: `TargetSelected`
    * `Value`: `2`
    * `Ref`: `None`

#### 【圣疗】 (Holy Therapy)
* **技能描述**：［回合限定］［水晶］任意分配3点［治疗］给1~3名角色，额外+1［攻击行动］或［法术行动］。
* **1. 主干配置**：
  * **SkillID**: `skill_saintess_holy_therapy`
  * **Category**: `Ultimate`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **IsTurnLimited**: `true`
  * **CustomExpression**: `Action.ActionRef == Attack || Action.ActionRef == Magic`
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `1`
  * **MaxCount**: `3`
  * **RequireTargetAllocations**: `true`
  * **AllocationTotal**: `3`
  * **MinAllocationPerTarget**: `1`
  * **MaxAllocationPerTarget**: `3`
  * **AllowedActionRefs**: `[Attack, Magic]`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Crystal, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(按 Action.TargetAllocations 对 1~3 个目标分配 3 点治疗)*:
    * `EffectType`: `EffectHeal`
    * `Target`: `TargetSelected`
    * `Value`: `Action.TargetAllocations[Target.UserID]`
    * `Ref`: `None`
  * **Effect[1]** *(额外+1攻击行动 或 +1法术行动，由 ActionRef 指定)*:
    * `Implementation`: `model.AppendExtraAction`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `ActionRef`: `Action.ActionRef`
    * `ElementRef`: `None`
    * `StatusRef`: `None`
    * `TokenRef`: `None`

#### 【怜悯】 (Mercy)
* **技能描述**：［持续］［宝石］［横置］，你的手牌上限恒定为7［恒定］，你+1［水晶］。
* **1. 主干配置**：
  * **SkillID**: `skill_saintess_mercy`
  * **Category**: `Ultimate`
  * **Type**: `Startup`
  * **Timing**: `TimingStartup`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionStart]`
  * **CustomExpression**: `Self.Orientation == Normal`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Gem, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(进入横置姿态)*:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Tapped`
    * `Ref`: `None`
  * **Effect[1]** *(手牌上限恒定为 7)*:
    * `EffectType`: `EffectSetHandLimitFixed`
    * `Target`: `TargetSelf`
    * `Value`: `7`
    * `Ref`: `None`
  * **Effect[2]** *(+1 水晶)*:
    * `EffectType`: `EffectAddEnergyStone`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `StoneRef`: `Crystal`
    * `Ref`: `None`

---

### 08. 魔法少女 (Magical Girl)

#### 【魔弹掌控】 (Magic Bullet Control)
* **技能描述**：你主动使用魔弹时可以选择逆向传递。
* **1. 主干配置**：
  * **SkillID**: `skill_magical_girl_magic_bullet_control`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnMagicDeclared`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDeclare]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.FlowType == ActionFlowMagicBulletChain`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectApplyCombatTag`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `ReverseDeliver`
    * `Ref`: `None`

#### 【魔弹融合】 (Magic Bullet Fusion)
* **技能描述**：你的地系或火系牌可以当魔弹使用。
* **1. 主干配置**：
  * **SkillID**: `skill_magical_girl_magic_bullet_fusion`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingBeforeActionExecute`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Action.SourceID == Self.UserID && (Action.PlayedCard.Element == Earth || Action.PlayedCard.Element == Fire)`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **3.1 行动改写配置**：
  * **ActionTransform.Hook**: `TimingBeforeActionExecute`
  * **ActionTransform.Optional**: `true` *(可选择是否将本次牌按魔弹链路处理)*
  * **ActionTransform.Priority**: `100`
  * **ActionTransform.Match.RequireActionType**: `Magic`
  * **ActionTransform.Match.RequirePlayedCardElements**: `[Earth, Fire]`
  * **ActionTransform.Match.ExcludeTemplateIDs**: `[]`
  * **ActionTransform.Rewrite.FlowRef**: `ActionFlowMagicBulletChain`
  * **ActionTransform.Rewrite.ActionTypeRef**: `Magic`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * `None` *(本技能的核心能力通过 ActionTransform 生效，不依赖 EffectNode)*

#### 【魔爆冲击】 (Magic Burst Impact)
* **技能描述**：（弃1张法术牌［展示］）我方战绩区+1颗［宝石］。2名目标对手各弃1张法术牌［展示］，每有人不如此做，你对他造成2点法术伤害③，然后你弃1张牌。
* **1. 主干配置**：
  * **SkillID**: `skill_magical_girl_magic_burst_impact`
  * **Category**: `Normal`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `None`
* **3. 目标选择规则**：
  * **SelectType**: `Enemy`
  * **MinCount**: `2`
  * **MaxCount**: `2`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Discards**: `[{Count: 1, Filter: {ReqCardType: Magic}}]`
  * **DiscardsVisibility**: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(我方战绩区 +1 宝石)*:
    * `EffectType`: `EffectAddTeamStone`
    * `Target`: `TargetSelfTeam`
    * `Value`: `1`
    * `StoneRef`: `Gem`
    * `Ref`: `None`
  * **Effect[1]** *(逐目标响应：目标可弃1张法术牌；若未弃则承受2点法术伤害，且你弃1张牌)*:
    * `EffectType`: `EffectPerTargetBranch`
    * `Target`: `TargetSelected`
    * `Value`: `0`
    * `BranchRef`:
      * `TargetSource`: `PerTargetSelectedTargets`
      * `InterruptType`: `WaitDiscard`
      * `TimeoutAsDeclined`: `true`
      * `DiscardRequirement`: `{Count: 1, Filter: {ReqCardType: Magic}}`
      * `DiscardVisibility`: `VisibilityPublic`
      * `OnSuccess`: `[]`
      * `OnDeclined`: `[{EffectType: EffectDamage, Target: TargetSelected, Value: 2, Ref: None}, {EffectType: EffectDiscard, Target: TargetSelf, Value: 1, Ref: None}]`

#### 【毁灭风暴】 (Destruction Storm)
* **技能描述**：［宝石］对任2名目标对手各造成2点法术伤害③。
* **1. 主干配置**：
  * **SkillID**: `skill_magical_girl_destruction_storm`
  * **Category**: `Ultimate`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `None`
* **3. 目标选择规则**：
  * **SelectType**: `Enemy`
  * **MinCount**: `2`
  * **MaxCount**: `2`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Gem, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelected`
    * `Value`: `2`
    * `Ref`: `None`

---

### 09. 女武神 (Valkyrie)

#### 【神圣追击】 (Holy Pursuit)
* **技能描述**：（［攻击行动］或［法术行动］结束后，移除你的1［治疗］）额外+1［攻击行动］。
* **1. 主干配置**：
  * **SkillID**: `skill_valkyrie_holy_pursuit`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnActionEnd`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionEnd]`
  * **CustomExpression**: `Action.SourceID == Self.UserID && (Action.CurrentType == Attack || Action.CurrentType == Magic)`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **HealCost**: `1`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `Implementation`: `model.AppendExtraAction`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `ActionRef`: `Attack`
    * `ElementRef`: `None`
    * `StatusRef`: `None`
    * `TokenRef`: `None`

#### 【秩序之印】 (Seal of Order)
* **技能描述**：（摸2张牌［强制］）你+1点［治疗］并+1［水晶］。
* **1. 主干配置**：
  * **SkillID**: `skill_valkyrie_seal_of_order`
  * **Category**: `Normal`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `None`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectDrawCard`
    * `Target`: `TargetSelf`
    * `Value`: `2`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectHeal`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `Ref`: `None`
  * **Effect[2]**:
    * `EffectType`: `EffectAddEnergyStone`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `StoneRef`: `Crystal`
    * `Ref`: `None`

#### 【和平行者】 (Peace Walker)
* **技能描述**：（你的回合内，发动［英灵召唤］后强制触发［强制］）［横置］，转为［英灵形态］，（当你执行主动［攻击行动］时）［转正］，脱离［英灵形态］。
* **1. 主干配置**：
  * **SkillID**: `skill_valkyrie_peace_walker`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnSkillExecuted` / `TimingBeforeActionExecute`
* **2. 前置条件**：
  * **PhaseLimit**: `None`
  * **CustomExpression**: `(Event.TriggerHook == TimingOnSkillExecuted && Event.SkillID == "skill_valkyrie_heroic_summon" && State.IsInSelfTurn(Self.UserID) == true) || (Event.TriggerHook == TimingBeforeActionExecute && Action.SourceID == Self.UserID && Action.CurrentType == Attack && Combat.IsActiveAttack == true && Self.Orientation == Tapped && Self.Form == "heroic_form")`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **规则约束**：`形态`是角色横置状态下的命名标签；横置进入形态，转正退出形态。
  * **Effect[0]** *(发动【英灵召唤】后：进入横置)*:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Tapped`
    * `Condition`: `Event.TriggerHook == TimingOnSkillExecuted && Event.SkillID == "skill_valkyrie_heroic_summon" && State.IsInSelfTurn(Self.UserID) == true`
    * `Ref`: `None`
  * **Effect[1]** *(横置时标记为英灵形态 heroic_form)*:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `"heroic_form"`
    * `Condition`: `Event.TriggerHook == TimingOnSkillExecuted && Event.SkillID == "skill_valkyrie_heroic_summon" && State.IsInSelfTurn(Self.UserID) == true`
    * `Ref`: `None`
* **7. 状态结算行为**（挂载于 `Self.Form == "heroic_form"`，用 Timing + Effects 表达）：
  * **ResolveTiming**: `TimingBeforeActionExecute`
  * **ResolveMode**: `Auto`
  * **ResolveOrder**: `ResolveEffects[0] -> ResolveEffects[1]` *(先转正，再退出英灵形态)*
  * **ResolveEffects[0]** *(英灵状态自动结算①：发起主动攻击时转正)*:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Normal`
    * `Condition`: `Event.TriggerHook == TimingBeforeActionExecute && Action.SourceID == Self.UserID && Action.CurrentType == Attack && Combat.IsActiveAttack == true && Self.Orientation == Tapped && Self.Form == "heroic_form"`
    * `Ref`: `None`
  * **ResolveEffects[1]** *(英灵状态自动结算②：转正后退出英灵形态)*:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `nil`
    * `Condition`: `Event.TriggerHook == TimingBeforeActionExecute && Action.SourceID == Self.UserID && Action.CurrentType == Attack && Combat.IsActiveAttack == true && Self.Orientation == Tapped && Self.Form == "heroic_form"`
    * `Ref`: `None`
  * **TriggerLimit**: `1`
  * **RemoveAfterResolve**: `true`

#### 【军威神光】 (Military Divine Light)
* **技能描述**：（回合开始时，若你处于［英灵形态］）选择以下1项发动：●你+1［治疗］，［转正］脱离［英灵型态］。●（移除我方战绩区X个星石，X<3）目标角色+X［治疗］。
* **1. 主干配置**：
  * **SkillID**: `skill_valkyrie_military_divine_light`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnTurnStart`
* **2. 前置条件**：
  * **PhaseLimit**: `[TurnBeforeStart]`
  * **CustomExpression**: `State.IsInSelfTurn(Self.UserID) == true && Self.Orientation == Tapped && Self.Form == "heroic_form"`
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `0` *(当选择“+1治疗并退出形态”时无需目标)*
  * **MaxCount**: `1` *(当选择“移除战绩区星石并治疗”时需要1名目标)*
  * **SelectedValueRule**: `Action.SelectedValue == 0 || Action.SelectedValue == 1 || Action.SelectedValue == 2` *(0=第一项；1/2=第二项中的X值，且 X<3)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列（分支）**：
  * **BranchSelector**: `Action.SelectedValue == 0 ? 0 : 1`
  * **Branch[0]** *(选项A：`Action.SelectedValue == 0`，无需目标)*:
    * **Effect[A0]**:
      * `EffectType`: `EffectHeal`
      * `Target`: `TargetSelf`
      * `Value`: `1`
      * `Ref`: `None`
    * **Effect[A1]**:
      * `EffectType`: `EffectSetOrientation`
      * `Target`: `TargetSelf`
      * `Value`: `0`
      * `OrientationRef`: `Normal`
      * `Ref`: `None`
    * **Effect[A2]**:
      * `EffectType`: `EffectSetForm`
      * `Target`: `TargetSelf`
      * `Value`: `0`
      * `FormRef`: `nil`
      * `Ref`: `None`
  * **Branch[1]** *(选项B：`Action.SelectedValue in {1,2}`，X=Action.SelectedValue，需1名目标)*:
    * **Effect[B0]** *(移除我方战绩区X个星石；负值表示移除)*:
      * `EffectType`: `EffectAddTeamStone`
      * `Target`: `TargetSelfTeam`
      * `Value`: `-Action.SelectedValue`
      * `StoneRef`: `Any`
      * `Ref`: `None`
    * **Effect[B1]** *(目标角色+X治疗)*:
      * `EffectType`: `EffectHeal`
      * `Target`: `TargetSelected`
      * `Value`: `Action.SelectedValue`
      * `Ref`: `Action.Targets[0].TargetUserID`

#### 【英灵召唤】 (Heroic Summon)
* **技能描述**：［水晶］（攻击命中时发动②）本次攻击伤害额外+1，（若你额外弃1张法术牌［展示］）目标角色+1［治疗］。
* **1. 主干配置**：
  * **SkillID**: `skill_valkyrie_heroic_summon`
  * **Category**: `Ultimate`
  * **Type**: `Response`
  * **Timing**: `TimingOnHitCheck`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatHitCheck]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsHit == true`
* **3. 目标选择规则**：
  * **SelectType**: `None`
  * **SelectedValueRule**: `Action.SelectedValue == 0 || Action.SelectedValue == 1` *(0=不额外弃牌；1=额外弃1张法术牌并治疗目标)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Crystal, Amount: 1}]`
  * **Discards**: `Count: Action.SelectedValue, Filter: {ReqCardType: Magic}`
  * **DiscardsVisibility**: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列（分支）**：
  * **Effect[0]** *(本次攻击伤害额外+1)*:
    * `EffectType`: `EffectAttackDamageModifier`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `1`
    * `Ref`: `None`
  * **BranchSelector**: `Action.SelectedValue`
  * **Branch[0]** *(不额外弃法术牌：`Action.SelectedValue == 0`)*:
    * `None` *(无附加效果)*
  * **Branch[1]** *(额外弃1张法术牌：`Action.SelectedValue == 1`)*:
    * **Effect[B0]** *(目标角色+1治疗)*:
      * `EffectType`: `EffectHeal`
      * `Target`: `TargetCurrentCombat`
      * `Value`: `1`
      * `Ref`: `Combat.TargetID`

---

### 10. 元素师 (Elementalist)

#### 【元素吸收】 (Elemental Absorption)
* **技能描述**：（对目标角色造成法术伤害时发动③）你+1［元素］。
* **1. 主干配置**：
  * **SkillID**: `skill_elementalist_elemental_absorption`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnDamageTaken`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDraw]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsMagic == true && Combat.FinalDamage > 0 && Combat.SkillID != "skill_elementalist_elemental_ignite"` *(元素点燃造成的伤害不触发元素吸收)*
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `TokenRef`: `Element`
    * `Ref`: `None`

#### 【元素点燃】 (Elemental Ignite)
* **技能描述**：（移除3点［元素］）对目标角色造成2点法术伤害③，额外+1［法术行动］；不能和［元素吸收］同时发动。
* **1. 主干配置**：
  * **SkillID**: `skill_elementalist_elemental_ignite`
  * **Category**: `Normal`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Self.Tokens[Element] >= 3`
* **3. 目标选择规则**：
  * **SelectType**: `Enemy`
  * **MinCount**: `1`
  * **MaxCount**: `1`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Tokens**: `[{Type: Element, Amount: 3}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelected`
    * `Value`: `2`
    * `Ref`: `None`
  * **Effect[1]**:
    * `Implementation`: `model.AppendExtraAction`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `ActionRef`: `Magic`
    * `ElementRef`: `None`
    * `StatusRef`: `None`
    * `TokenRef`: `None`

#### 【雷击】 (Thunder Strike)
* **技能描述**：对目标角色造成1点法术伤害③，我方战绩区+1宝石，（若你额外弃1张雷系牌［展示］）本次法术伤害额外+1。
* **1. 主干配置**：
  * **SkillID**: `skill_elementalist_thunder_strike`
  * **Category**: `Unique`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Action.PlayedCard.CharacterSkillMap[Self.CharacterID] == "skill_elementalist_thunder_strike"`
* **3. 目标选择规则**：
  * **SelectType**: `Enemy`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **SelectedValueRule**: `Action.SelectedValue == 0 || Action.SelectedValue == 1` *(0=不额外弃牌；1=额外弃1张雷系牌)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayRequired`
  * **Discards**: `Count: Action.SelectedValue, Filter: {ReqElement: Thunder}`
  * **DiscardsVisibility**: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **BranchSelector**: `Action.SelectedValue`
  * **Branch[0]** *(不额外弃雷系牌)*:
    * **Effect[B0]**:
      * `EffectType`: `EffectDamage`
      * `Target`: `TargetSelected`
      * `Value`: `1`
      * `Ref`: `None`
  * **Branch[1]** *(额外弃1张雷系牌［展示］)*:
    * **Effect[B1]**:
      * `EffectType`: `EffectDamage`
      * `Target`: `TargetSelected`
      * `Value`: `2`
      * `Ref`: `None`
  * **Effect[1]** *(我方战绩区+1宝石；与是否额外弃牌无关)*:
    * `EffectType`: `EffectAddTeamStone`
    * `Target`: `TargetSelfTeam`
    * `Value`: `1`
    * `StoneRef`: `Gem`
    * `Ref`: `None`

#### 【冰冻】 (Freeze)
* **技能描述**：对目标角色造成1点法术伤害③，并指定1名角色+1［治疗］，（若你额外弃1张水系牌［展示］）本次法术伤害额外+1。
* **1. 主干配置**：
  * **SkillID**: `skill_elementalist_freeze`
  * **Category**: `Unique`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Action.PlayedCard.CharacterSkillMap[Self.CharacterID] == "skill_elementalist_freeze"`
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `2` *(Targets[0]=伤害目标；Targets[1]=治疗目标)*
  * **MaxCount**: `2`
  * **Filters**: `Action.Targets[0] 为法术伤害目标（可为任意角色）；Action.Targets[1] 为治疗目标（可为任意角色）`
  * **SelectedValueRule**: `Action.SelectedValue == 0 || Action.SelectedValue == 1` *(0=不额外弃牌；1=额外弃1张水系牌)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayRequired`
  * **Discards**: `Count: Action.SelectedValue, Filter: {ReqElement: Water}`
  * **DiscardsVisibility**: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列（分支）**：
  * **BranchSelector**: `Action.SelectedValue == 0 ? 0 : 1`
  * **Branch[0]** *(不额外弃1张水系牌)*:
    * **Effect[B0]** *(对 Targets[0] 造成1点法术伤害)*:
      * `EffectType`: `EffectDamage`
      * `Target`: `TargetSelected`
      * `Value`: `1`
      * `Ref`: `Action.Targets[0].TargetUserID`
  * **Branch[1]** *(额外弃1张水系牌［展示］)*:
    * **Effect[B1]** *(对 Targets[0] 造成2点法术伤害)*:
      * `EffectType`: `EffectDamage`
      * `Target`: `TargetSelected`
      * `Value`: `2`
      * `Ref`: `Action.Targets[0].TargetUserID`
  * **Effect[1]** *(对 Targets[1] 增加1点治疗)*:
    * `EffectType`: `EffectHeal`
    * `Target`: `TargetSelected`
    * `Value`: `1`
    * `Ref`: `Action.Targets[1].TargetUserID`

#### 【风刃】 (Wind Blade)
* **技能描述**：对目标角色造成1点法术伤害③，额外+1［攻击行动］，（若你额外弃1张风系牌［展示］）本次法术伤害额外+1。
* **1. 主干配置**：
  * **SkillID**: `skill_elementalist_wind_blade`
  * **Category**: `Unique`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Action.PlayedCard.CharacterSkillMap[Self.CharacterID] == "skill_elementalist_wind_blade"`
* **3. 目标选择规则**：
  * **SelectType**: `Enemy`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **SelectedValueRule**: `Action.SelectedValue == 0 || Action.SelectedValue == 1` *(0=不额外弃牌；1=额外弃1张风系牌)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayRequired`
  * **Discards**: `Count: Action.SelectedValue, Filter: {ReqElement: Wind}`
  * **DiscardsVisibility**: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列（分支）**：
  * **BranchSelector**: `Action.SelectedValue`
  * **Branch[0]** *(不额外弃1张风系牌)*:
    * **Effect[B0]**:
      * `EffectType`: `EffectDamage`
      * `Target`: `TargetSelected`
      * `Value`: `1`
      * `Ref`: `None`
  * **Branch[1]** *(额外弃1张风系牌［展示］)*:
    * **Effect[B1]**:
      * `EffectType`: `EffectDamage`
      * `Target`: `TargetSelected`
      * `Value`: `2`
      * `Ref`: `None`
  * **Effect[1]** *(额外+1攻击行动)*:
    * `Implementation`: `model.AppendExtraAction`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `ActionRef`: `Attack`
    * `ElementRef`: `None`
    * `StatusRef`: `None`
    * `TokenRef`: `None`

#### 【陨石】 (Meteor)
* **技能描述**：对目标角色造成1点法术伤害③，额外+1［法术行动］，（若你额外弃1张地系牌［展示］）本次法术伤害额外+1。
* **1. 主干配置**：
  * **SkillID**: `skill_elementalist_meteor`
  * **Category**: `Unique`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Action.PlayedCard.CharacterSkillMap[Self.CharacterID] == "skill_elementalist_meteor"`
* **3. 目标选择规则**：
  * **SelectType**: `Enemy`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **SelectedValueRule**: `Action.SelectedValue == 0 || Action.SelectedValue == 1` *(0=不额外弃牌；1=额外弃1张地系牌)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayRequired`
  * **Discards**: `Count: Action.SelectedValue, Filter: {ReqElement: Earth}`
  * **DiscardsVisibility**: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列（分支）**：
  * **BranchSelector**: `Action.SelectedValue`
  * **Branch[0]** *(不额外弃1张地系牌)*:
    * **Effect[B0]**:
      * `EffectType`: `EffectDamage`
      * `Target`: `TargetSelected`
      * `Value`: `1`
      * `Ref`: `None`
  * **Branch[1]** *(额外弃1张地系牌［展示］)*:
    * **Effect[B1]**:
      * `EffectType`: `EffectDamage`
      * `Target`: `TargetSelected`
      * `Value`: `2`
      * `Ref`: `None`
  * **Effect[1]** *(额外+1法术行动)*:
    * `Implementation`: `model.AppendExtraAction`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `ActionRef`: `Magic`
    * `ElementRef`: `None`
    * `StatusRef`: `None`
    * `TokenRef`: `None`

#### 【火球】 (Fireball)
* **技能描述**：对目标角色造成2点法术伤害③，（若你额外弃1张火系牌［展示］）本次法术伤害额外+1。
* **1. 主干配置**：
  * **SkillID**: `skill_elementalist_fireball`
  * **Category**: `Unique`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Action.PlayedCard.CharacterSkillMap[Self.CharacterID] == "skill_elementalist_fireball"`
* **3. 目标选择规则**：
  * **SelectType**: `Enemy`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **SelectedValueRule**: `Action.SelectedValue == 0 || Action.SelectedValue == 1` *(0=不额外弃牌；1=额外弃1张火系牌)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayRequired`
  * **Discards**: `Count: Action.SelectedValue, Filter: {ReqElement: Fire}`
  * **DiscardsVisibility**: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列（分支）**：
  * **BranchSelector**: `Action.SelectedValue`
  * **Branch[0]** *(不额外弃1张火系牌)*:
    * **Effect[B0]**:
      * `EffectType`: `EffectDamage`
      * `Target`: `TargetSelected`
      * `Value`: `2`
      * `Ref`: `None`
  * **Branch[1]** *(额外弃1张火系牌［展示］)*:
    * **Effect[B1]**:
      * `EffectType`: `EffectDamage`
      * `Target`: `TargetSelected`
      * `Value`: `3`
      * `Ref`: `None`

#### 【月光】 (Moonlight)
* **技能描述**：［宝石］对目标角色造成（X+1）点法术伤害③，X为你剩余的能量数。
* **1. 主干配置**：
  * **SkillID**: `skill_elementalist_moonlight`
  * **Category**: `Ultimate`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `None`
* **3. 目标选择规则**：
  * **SelectType**: `Enemy`
  * **MinCount**: `1`
  * **MaxCount**: `1`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Gem, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(X=支付费用后的剩余能量数)*:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelected`
    * `Value`: `Self.EnergyCount + 1`
    * `Ref`: `None`

---

### 11. 仲裁者 (Arbiter)

#### 【仲裁法则】 (Arbitration Law)
* **技能描述**：游戏初始时，你+2［水晶］。
* **1. 主干配置**：
  * **SkillID**: `skill_arbiter_arbitration_law`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnTurnStart` *(仅用于承接系统 `GameStart` 事件，不用于常规回合开始结算)*
* **2. 前置条件**：
  * **PhaseLimit**: `[GameInit]`
  * **CustomExpression**: `Event.SourceType == SourceSystem && Event.CauseAction == "GameStart"`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectAddEnergyStone`
    * `Target`: `TargetSelf`
    * `Value`: `2`
    * `StoneRef`: `Crystal`
    * `Ref`: `None`

#### 【仪式中断】 (Ritual Interrupt)
* **技能描述**：（仅［审判形态］下发动）［转正］脱离［审判形态］，我方［战绩区］+1［宝石］。
* **1. 主干配置**：
  * **SkillID**: `skill_arbiter_ritual_interrupt`
  * **Category**: `Normal`
  * **Type**: `Startup`
  * **Timing**: `TimingStartup`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionStart]`
  * **CustomExpression**: `Self.Orientation == Tapped && Self.Form == "judgment_form"`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Normal`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `nil`
    * `Ref`: `None`
  * **Effect[2]**:
    * `EffectType`: `EffectAddTeamStone`
    * `Target`: `TargetSelfTeam`
    * `Value`: `1`
    * `StoneRef`: `Gem`
    * `Ref`: `None`

#### 【末日审判】 (Doomsday Judgment)
* **技能描述**：（移除所有［审判］）对目标角色造成等量的法术伤害③；在你的行动阶段开始时，若［审判］已达到上限，该行动阶段你必须发动［末日审判］。
* **1. 主干配置**：
  * **SkillID**: `skill_arbiter_doomsday_judgment`
  * **Category**: `Normal`
  * **Type**: `Magic`
  * **Timing**: `TimingActive` 
* **1.1 强制发动配置**：
  * **Mandatory.MatchTiming**: `TimingOnBeforeAction`
  * **Mandatory.ConditionExpression**: `State.IsInSelfTurn(Self.UserID) == true && State.IsJudgmentAtCap(Self.UserID) == true`
  * **Mandatory.LockMode**: `SkillMandatoryLockActionPhaseToSelfSkill` *(命中后该角色在本行动阶段只能执行【末日审判】)*
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Self.Tokens[Judgment] >= 1`
* **3. 目标选择规则**：
  * **SelectType**: `Enemy`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **SelectedValueRule**: `Action.SelectedValue == Self.Tokens[Judgment] && Action.SelectedValue >= 1` *(X=本次移除的审判数，按“移除所有审判”约束)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Tokens**: `[{Type: Judgment, Amount: Action.SelectedValue}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelected`
    * `Value`: `Action.SelectedValue`
    * `Ref`: `None`

#### 【审判浪潮】 (Judgment Tide)
* **技能描述**：（你每次承受伤害⑥）你+1［审判］。
* **1. 主干配置**：
  * **SkillID**: `skill_arbiter_judgment_tide`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnDamageTaken`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDraw]`
  * **CustomExpression**: `Combat.TargetID == Self.UserID && Combat.FinalDamage > 0`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `TokenRef`: `Judgment`
    * `Ref`: `None`

#### 【仲裁仪式】 (Arbitration Ritual)
* **技能描述**：［持续］［宝石］［横置］转为［审判形态］，你的手牌上限恒定为5；每次在你的回合开始时，你+1［审判］。
* **1. 主干配置**：
  * **SkillID**: `skill_arbiter_arbitration_ritual`
  * **Category**: `Ultimate`
  * **Type**: `Startup`
  * **Timing**: `TimingStartup`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionStart]`
  * **CustomExpression**: `Self.Orientation == Normal`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Gem, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(启动时横置)*:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Tapped`
    * `Condition`: `Event.TriggerHook == TimingStartup`
    * `Ref`: `None`
  * **Effect[1]** *(启动时进入审判形态)*:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `"judgment_form"`
    * `Condition`: `Event.TriggerHook == TimingStartup`
    * `Ref`: `None`
  * **Effect[2]** *(启动时手牌上限恒定为5)*:
    * `EffectType`: `EffectSetHandLimitFixed`
    * `Target`: `TargetSelf`
    * `Value`: `5`
    * `Condition`: `Event.TriggerHook == TimingStartup`
    * `Ref`: `None`
* **7. 状态结算行为**（挂载于 `Self.Form == "judgment_form"`）：
  * **ResolveTiming**: `TimingOnTurnStart`
  * **RequireHolderIsTurnPlayer**: `true`
  * **ResolveMode**: `Auto`
  * **ResolveEffects[0]** *(审判形态下每回合开始+1审判)*:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `TokenRef`: `Judgment`
    * `Condition`: `Self.Orientation == Tapped && Self.Form == "judgment_form"`
    * `Ref`: `None`

#### 【判决天平】 (Judgment Balance)
* **技能描述**：［水晶］你+1［审判］，再选择以下一项发动：●弃掉你的所有手牌。●将你的手牌补到上限［强制］，我方战绩区+1［宝石］。
* **1. 主干配置**：
  * **SkillID**: `skill_arbiter_judgment_balance`
  * **Category**: `Ultimate`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `None`
* **3. 目标选择规则**：
  * **SelectType**: `None`
  * **SelectedValueRule**: `Action.SelectedValue == 0 || Action.SelectedValue == 1` *(0=弃掉所有手牌；1=补到上限并+1宝石)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Crystal, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列（分支）**：
  * **Effect[0]** *(你+1审判)*:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `TokenRef`: `Judgment`
    * `Ref`: `None`
  * **BranchSelector**: `Action.SelectedValue`
  * **Branch[0]** *(弃掉你的所有手牌)*:
    * **Effect[B0]**:
      * `EffectType`: `EffectDiscard`
      * `Target`: `TargetSelf`
      * `Value`: `Self.HandCount`
      * `Ref`: `None`
  * **Branch[1]** *(手牌补到上限并我方战绩区+1宝石)*:
    * **Effect[B1]**:
      * `EffectType`: `EffectDrawCard`
      * `Target`: `TargetSelf`
      * `Value`: `Max(0, Self.HandLimit - Self.HandCount)`
      * `Ref`: `None`
    * **Effect[B2]**:
      * `EffectType`: `EffectAddTeamStone`
      * `Target`: `TargetSelfTeam`
      * `Value`: `1`
      * `StoneRef`: `Gem`
      * `Ref`: `None`

---

### 12. 冒险家 (Adventurer)

#### 【欺诈】 (Fraud)
* **技能描述**：弃2张同系牌［展示］，视为一次除暗系以外的任意系主动攻击，该系由你决定；或弃任意3张同系牌［展示］，视为一次暗系主动攻击。
* **1. 主干配置**：
  * **SkillID**: `skill_adventurer_fraud`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `None`
* **3. 目标选择规则**：
  * **SelectType**: `Enemy`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **SelectedValueRule**: `(Action.SelectedValue == 0 && Action.ElementRef != Dark) || (Action.SelectedValue == 1 && Action.ElementRef == Dark)` *(0=弃2张同系牌并选择非暗系；1=弃3张同系牌并固定暗系)*
* **3.1 行动改写配置**：
  * **ActionTransform.Hook**: `TimingBeforeActionExecute`
  * **ActionTransform.Optional**: `false`
  * **ActionTransform.Priority**: `100`
  * **ActionTransform.CancelCurrentAction**: `true`
  * **ActionTransform.Match.RequireActionType**: `None`
  * **ActionTransform.Match.RequirePlayedCardTypes**: `[]`
  * **ActionTransform.Match.RequirePlayedCardElements**: `[]`
  * **ActionTransform.Match.ExcludeTemplateIDs**: `[]`
  * **ActionTransform.Rewrite.FlowRef**: `ActionFlowNormalCombat`
  * **ActionTransform.Rewrite.ActionTypeRef**: `Attack`
  * **ActionTransform.Rewrite.ExecuteImmediately**: `true`
  * **ActionTransform.Rewrite.TreatAsActiveAttack**: `true`
  * **ActionTransform.Rewrite.ElementPickMode**: `RewriteElementFromActionRef`
  * **ActionTransform.Rewrite.FixedElementRef**: `None`
  * **SubmitAction.ElementRef**: `必填` *(由玩家指定本次“视为攻击”的系别；SelectedValueRule 负责约束是否允许暗系)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Discards**:
    * `Count`: `Action.SelectedValue == 0 ? 2 : 3`
    * `Filter`: `{SameAttribute: MatchElement}`
    * `Visibility`: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * `None` *(本技能通过 ActionTransform 将当前技能动作改写为“立即结算的主动攻击”)*  

#### 【强运】 (Good Fortune)
* **技能描述**：当你发动【欺诈】时，你+1［水晶］。
* **1. 主干配置**：
  * **SkillID**: `skill_adventurer_good_fortune`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnSkillExecuted`
* **2. 前置条件**：
  * **PhaseLimit**: `None`
  * **CustomExpression**: `Event.SkillID == "skill_adventurer_fraud" && Event.OperatorID == Self.UserID`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectAddEnergyStone`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `StoneRef`: `Crystal`
    * `Ref`: `None`

#### 【地下法则】 (Underground Rule)
* **技能描述**：你执行购买时，改为我方战绩区+2［宝石］。
* **1. 主干配置**：
  * **SkillID**: `skill_adventurer_underground_rule`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingBeforeActionExecute`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Action.SourceID == Self.UserID && Action.CurrentType == Buy`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **3.1 行动改写配置**：
  * **ActionTransform.Hook**: `TimingBeforeActionExecute`
  * **ActionTransform.Optional**: `false`
  * **ActionTransform.Priority**: `100`
  * **ActionTransform.CancelCurrentAction**: `true`
  * **ActionTransform.Match.RequireActionType**: `Buy`
  * **ActionTransform.Match.RequirePlayedCardTypes**: `[]`
  * **ActionTransform.Match.RequirePlayedCardElements**: `[]`
  * **ActionTransform.Match.ExcludeTemplateIDs**: `[]`
  * **ActionTransform.Rewrite**: `nil` *(只取消 Buy 默认结算，不改写到其他流水线)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectAddTeamStone`
    * `Target`: `TargetSelfTeam`
    * `Value`: `2`
    * `StoneRef`: `Gem`
    * `Ref`: `None`

#### 【冒险者天堂】 (Adventurer's Paradise)
* **技能描述**：你执行提炼时，将提炼出的［宝石］和［水晶］全部交给目标队友，然后移除你的1［能量］。
* **1. 主干配置**：
  * **SkillID**: `skill_adventurer_paradise`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingBeforeActionExecute`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Action.SourceID == Self.UserID && Action.CurrentType == Extract`
* **3. 目标选择规则**：
  * **SelectType**: `TeamOther`
  * **MinCount**: `1`
  * **MaxCount**: `1`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(将本次提炼产出全部重定向给目标队友)*:
    * `EffectType`: `EffectRedirectCurrentExtractOutput`
    * `Target`: `TargetSelected`
    * `Value`: `0`
    * `Ref`: `None`
  * **Effect[1]** *(移除自身1点能量；若当前能量不足则按可移除量结算，不阻断技能)*:
    * `EffectType`: `EffectAddEnergyStone`
    * `Target`: `TargetSelf`
    * `Value`: `-1`
    * `StoneRef`: `Any`
    * `Ref`: `None`

#### 【偷天换日】 (Steal the Sky)
* **技能描述**：［回合限定］［水晶］将对方战绩区的1［宝石］转移到我方战绩区，或将我方战绩区的［水晶］全部转换成［宝石］，额外+1［攻击行动］或［法术行动］。
* **1. 主干配置**：
  * **SkillID**: `skill_adventurer_steal_the_sky`
  * **Category**: `Ultimate`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **IsTurnLimited**: `true`
  * **CustomExpression**: `Action.ActionRef == Attack || Action.ActionRef == Magic`
* **3. 目标选择规则**：
  * **SelectType**: `None`
  * **SelectedValueRule**: `Action.SelectedValue == 0 || Action.SelectedValue == 1` *(0=转移对方1宝石；1=将我方水晶全部转宝石)*
  * **AllowedActionRefs**: `[Attack, Magic]` *(用于“额外+1攻击行动或法术行动”)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Crystal, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列（分支）**：
  * **BranchSelector**: `Action.SelectedValue`
  * **Branch[0]** *(将敌方战绩区1宝石转移到我方战绩区)*:
    * **Effect[B0]**:
      * `EffectType`: `EffectTransferTeamStone`
      * `Target`: `TargetSelfTeam`
      * `Value`: `1`
      * `StoneRef`: `Gem`
      * `FromTargetRef`: `TargetEnemyTeam`
      * `Ref`: `None`
  * **Branch[1]** *(将我方战绩区所有水晶转为宝石)*:
    * **Effect[B1]**:
      * `EffectType`: `EffectConvertTeamStone`
      * `Target`: `TargetSelfTeam`
      * `Value`: `0` *(<=0 表示全部)*
      * `StoneRef`: `Crystal`
      * `StoneToRef`: `Gem`
      * `Ref`: `None`
  * **Effect[1]** *(额外+1攻击行动或法术行动)*:
    * `Implementation`: `model.AppendExtraAction`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `ActionRef`: `Attack,Magic`
    * `ElementRef`: `None`
    * `StatusRef`: `None`
    * `TokenRef`: `None`

---

### 13. 圣枪骑士 (Holy Lancer)

#### 【神圣启示】 (Divine Revelation)
* **技能描述**：（我方［星杯区］的［星杯］数不小于对方时）你的［治疗］上限+1。
* **1. 主干配置**：
  * **SkillID**: `skill_holy_lancer_divine_revelation`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnTurnStart` *(仅用于承接系统 `GameStart` 事件，施加常驻规则模板实例)*
* **1.2 规则修饰器模板配置**：
  * **ModifierID**: `rm_holy_lancer_divine_revelation_max_heal_plus_1`
  * **Domain**: `RuleModifierDomainAttribute`
  * **Priority**: `100`
  * **ConditionExpression**: `State.GetTeamCups(Self.Team) >= State.GetTeamCups(State.GetEnemyTeam(Self.Team))`
  * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
  * **AttrPayload**: `{AttrType: PlayerAttributeMaxHeal, Operation: AttributeModifyAdd, Value: 1}`
* **2. 前置条件**：
  * **PhaseLimit**: `[GameInit]`
  * **CustomExpression**: `Event.SourceType == SourceSystem && Event.CauseAction == "GameStart"`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(施加“动态+1治疗上限”的常驻规则)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_holy_lancer_divine_revelation_max_heal_plus_1"`
    * `RuleLifetimeRef`: `RuleLifePermanent`
    * `Ref`: `None`

#### 【辉耀】 (Radiance)
* **技能描述**：（弃1张水系牌［展示］）所有人各+1［治疗］，额外+1［攻击行动］。
* **1. 主干配置**：
  * **SkillID**: `skill_holy_lancer_radiance`
  * **Category**: `Normal`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `None`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Discards**:
    * `Count`: `1`
    * `Filter`: `{ReqElement: Water}`
    * `Visibility`: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(所有人各+1治疗)*:
    * `EffectType`: `EffectHeal`
    * `Target`: `TargetAllPlayers`
    * `Value`: `1`
    * `Ref`: `None`
  * **Effect[1]** *(额外+1攻击行动)*:
    * `Implementation`: `model.AppendExtraAction`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `ActionRef`: `Attack`
    * `Ref`: `None`

#### 【惩戒】 (Punishment)
* **技能描述**：（弃1张法术牌［展示］）将其他角色的1点［治疗］转移给你，额外+1［攻击行动］。
* **1. 主干配置**：
  * **SkillID**: `skill_holy_lancer_punishment`
  * **Category**: `Normal`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `None`
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **Filters**: `Target.UserID != Self.UserID && RequireHeal >= 1`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Discards**:
    * `Count`: `1`
    * `Filter`: `{ReqCardType: Magic}`
    * `Visibility`: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(目标-1治疗)*:
    * `EffectType`: `EffectHeal`
    * `Target`: `TargetSelected`
    * `Value`: `-1`
    * `Ref`: `None`
  * **Effect[1]** *(你+1治疗)*:
    * `EffectType`: `EffectHeal`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `Ref`: `None`
  * **Effect[2]** *(额外+1攻击行动)*:
    * `Implementation`: `model.AppendExtraAction`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `ActionRef`: `Attack`
    * `Ref`: `None`

#### 【圣击】 (Holy Strike)
* **技能描述**：（攻击命中后发动②）你+1［治疗］。
* **1. 主干配置**：
  * **SkillID**: `skill_holy_lancer_holy_strike`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnHitCheck`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatHitCheck]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsHit == true && State.IsSkillDisabled(Self.UserID, "skill_holy_lancer_holy_strike") == false`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectHeal`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `Ref`: `None`

#### 【天枪】 (Sky Spear)
* **技能描述**：（主动攻击前发动①）移除你的2点［治疗］，本次攻击对手无法应战；不能和［圣击］同时发动。
* **1. 主干配置**：
  * **SkillID**: `skill_holy_lancer_sky_spear`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnAttackDeclared`
* **1.2 规则修饰器模板配置**：
  * **ModifierID**: `rm_holy_lancer_sky_spear_disable_holy_strike_combat`
  * **Domain**: `RuleModifierDomainSkillGate`
  * **Priority**: `200`
  * **ConditionExpression**: `None`
  * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
  * **SkillGatePayload**: `{Mode: SkillGateDisallowList, SkillIDs: ["skill_holy_lancer_holy_strike"]}`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDeclare]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && Self.Heal >= 2 && State.IsSkillDisabled(Self.UserID, "skill_holy_lancer_sky_spear") == false`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **HealCost**: `2`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `Unrespondable`
* **6. 执行效果序列**：
  * **Effect[0]** *(施加“禁用圣击”门禁规则，持续到本次攻击结算结束，确保不与圣击同时发动)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_holy_lancer_sky_spear_disable_holy_strike_combat"`
    * `RuleLifetimeRef`: `RuleLifeUntilCombatEnd`
    * `Ref`: `None`

#### 【地枪】 (Earth Spear)
* **技能描述**：（主动攻击命中后发动②）移除你的X点［治疗］，本次攻击伤害额外+X，X最高为4；不能和［圣击］同时发动。
* **1. 主干配置**：
  * **SkillID**: `skill_holy_lancer_earth_spear`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnHitCheck`
* **1.2 规则修饰器模板配置**：
  * **ModifierID**: `rm_holy_lancer_earth_spear_disable_holy_strike_combat`
  * **Domain**: `RuleModifierDomainSkillGate`
  * **Priority**: `200`
  * **ConditionExpression**: `None`
  * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
  * **SkillGatePayload**: `{Mode: SkillGateDisallowList, SkillIDs: ["skill_holy_lancer_holy_strike"]}`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatHitCheck]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && Combat.IsHit == true && Self.Heal >= 1`
* **3. 目标选择规则**：
  * **SelectType**: `None`
  * **SelectedValueRule**: `1 <= Action.SelectedValue && Action.SelectedValue <= Min(4, Self.Heal)`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **HealCost**: `Action.SelectedValue`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(施加“禁用圣击”门禁规则，持续到本次攻击结算结束，确保不与圣击同时发动)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_holy_lancer_earth_spear_disable_holy_strike_combat"`
    * `RuleLifetimeRef`: `RuleLifeUntilCombatEnd`
    * `Ref`: `None`
  * **Effect[1]** *(本次攻击伤害额外+X)*:
    * `EffectType`: `EffectAttackDamageModifier`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `Action.SelectedValue`
    * `Ref`: `None`

#### 【圣光祈愈】 (Holy Prayer)
* **技能描述**：［宝石］无视你的［治疗］上限为你+2［治疗］，但你的［治疗］数最高为5，额外+1［攻击行动］；本回合你不能再发动［天枪］。
* **1. 主干配置**：
  * **SkillID**: `skill_holy_lancer_holy_prayer`
  * **Category**: `Ultimate`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **1.2 规则修饰器模板配置**：
  * **Template[A]** *(本次治疗无视上限 + 绝对封顶5)*:
    * **ModifierID**: `rm_holy_lancer_prayer_heal_policy_ignore_cap_abs_5`
    * **Domain**: `RuleModifierDomainHealPolicy`
    * **Priority**: `300`
    * **ConditionExpression**: `None`
    * **StackPolicy**: `RuleModifierStackReplaceByDomainPriority`
    * **HealPolicyPayload**: `{ApplyMode: HealApplyIgnoreMax, AbsoluteMax: 5}`
  * **Template[B]** *(本回合禁用天枪)*:
    * **ModifierID**: `rm_holy_lancer_prayer_disable_sky_spear_turn`
    * **Domain**: `RuleModifierDomainSkillGate`
    * **Priority**: `300`
    * **ConditionExpression**: `None`
    * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
    * **SkillGatePayload**: `{Mode: SkillGateDisallowList, SkillIDs: ["skill_holy_lancer_sky_spear"]}`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `None`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Gem, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(施加“本次治疗无视上限+绝对封顶5”的治疗策略规则)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_holy_lancer_prayer_heal_policy_ignore_cap_abs_5"`
    * `RuleLifetimeRef`: `RuleLifeThisEffectChain`
    * `Ref`: `None`
  * **Effect[1]** *(在规则生效窗口内执行+2治疗)*:
    * `EffectType`: `EffectHeal`
    * `Target`: `TargetSelf`
    * `Value`: `2`
    * `Ref`: `None`
  * **Effect[2]** *(本回合禁用天枪)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_holy_lancer_prayer_disable_sky_spear_turn"`
    * `RuleLifetimeRef`: `RuleLifeUntilTurnEnd`
    * `Ref`: `None`
  * **Effect[3]** *(额外+1攻击行动)*:
    * `Implementation`: `model.AppendExtraAction`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `ActionRef`: `Attack`
    * `Ref`: `None`

---

### 14. 精灵射手 (Elf Archer)

#### 【元素射击】 (Elemental Shot)
* **技能描述**：［回合限定］（主动攻击时①，若攻击牌非暗系，弃1张法术牌［展示］或移除1个［祝福］）根据攻击牌系别附加以下［元素箭］效果：火之矢=本次攻击伤害额外+1；水之矢=命中后目标+1治疗；风之矢=攻击行动结束后+1攻击行动；雷之矢=本次攻击无法应战；地之矢=命中后对目标造成1点法术伤害③。
* **1. 主干配置**：
  * **SkillID**: `skill_elf_archer_elemental_shot`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnAttackDeclared`
* **1.2 规则修饰器模板配置**：
  * **Template[A]** *(元素射击已发动挂载标记；延后结算时直接按攻击牌系别分流)*:
    * **ModifierID**: `rm_elf_archer_elemental_shot_armed`
    * **Domain**: `RuleModifierDomainCardSource`
    * **Priority**: `50`
    * **ConditionExpression**: `None`
    * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
    * **CardSourcePayload**: `{ProjectionMode: CardSourceProjectionAsHand, FieldMarks: []}` *(空投影负载，仅作跨阶段挂载标记)*
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDeclare]`
  * **IsTurnLimited**: `true`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && Combat.AttackCard.Element != Dark`
* **3. 目标选择规则**：
  * **SelectType**: `None`
  * **SelectedValueRule**: `Action.SelectedValue == 0 || (Action.SelectedValue == 1 && State.CountFieldMark(Self.UserID, Blessing) >= 1)` *(0=弃1法术牌；1=移除1祝福)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Discards**:
    * `Count`: `Action.SelectedValue == 0 ? 1 : 0`
    * `Filter`: `{ReqCardType: Magic}`
    * `Visibility`: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(分支成本：选择“移除祝福”时执行)*:
    * `EffectType`: `EffectRemoveFieldMark`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `FieldMarkRef`: `Blessing`
    * `Condition`: `Action.SelectedValue == 1`
    * `Ref`: `None`
  * **BranchSelector**: `Combat.AttackCard.Element`
  * **Branch[Fire]** *(火之矢：本次攻击伤害额外+1)*:
    * **Effect[F0]**:
      * `EffectType`: `EffectAttackDamageModifier`
      * `Target`: `TargetCurrentCombat`
      * `Value`: `1`
      * `Ref`: `None`
  * **Branch[Thunder]** *(雷之矢：本次攻击无法应战)*:
    * **Effect[T0]**:
      * `EffectType`: `EffectApplyCombatTag`
      * `Target`: `TargetCurrentCombat`
      * `Value`: `Unrespondable` *(CombatInterceptTag 枚举；表示“无法应战”)*
      * `Ref`: `None`
  * **Branch[Water]** *(水之矢：挂载命中后延后结算标记)*:
    * **Effect[W0]**:
      * `EffectType`: `EffectApplyRuleModifier`
      * `Target`: `TargetSelf`
      * `Value`: `0`
      * `RuleModifierRef`: `"rm_elf_archer_elemental_shot_armed"`
      * `RuleLifetimeRef`: `RuleLifeUntilCombatEnd`
      * `Ref`: `None`
  * **Branch[Earth]** *(地之矢：挂载命中后延后结算标记)*:
    * **Effect[E0]**:
      * `EffectType`: `EffectApplyRuleModifier`
      * `Target`: `TargetSelf`
      * `Value`: `0`
      * `RuleModifierRef`: `"rm_elf_archer_elemental_shot_armed"`
      * `RuleLifetimeRef`: `RuleLifeUntilCombatEnd`
      * `Ref`: `None`
  * **Branch[Wind]** *(风之矢：挂载行动后延后结算标记)*:
    * **Effect[Wi0]**:
      * `EffectType`: `EffectApplyRuleModifier`
      * `Target`: `TargetSelf`
      * `Value`: `0`
      * `RuleModifierRef`: `"rm_elf_archer_elemental_shot_armed"`
      * `RuleLifetimeRef`: `RuleLifeUntilCombatEnd`
      * `Ref`: `None`

#### 【元素射击·命中后结算】 (Elemental Shot Hit Resolve)
* **说明**：由【元素射击】在①挂载的“已发动”标记驱动；在②命中判定时按攻击牌系别结算水/地箭。
* **1. 主干配置**：
  * **SkillID**: `skill_elf_archer_elemental_shot_hit_resolve`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnHitCheck`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatHitCheck]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && Combat.IsHit == true && State.HasRuleModifier(Self.UserID, "rm_elf_archer_elemental_shot_armed") == true && (Combat.AttackCard.Element == Water || Combat.AttackCard.Element == Earth)`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **BranchSelector**: `Combat.AttackCard.Element`
  * **Branch[Water]**:
    * **Effect[W0]** *(水之矢：目标+1治疗)*:
      * `EffectType`: `EffectHeal`
      * `Target`: `TargetCurrentCombat`
      * `Value`: `1`
      * `Ref`: `Combat.TargetID`
    * **Effect[W1]** *(清理“已发动”标记)*:
      * `EffectType`: `EffectRemoveRuleModifier`
      * `Target`: `TargetSelf`
      * `Value`: `0`
      * `RuleRemoveRef`: `{Mode: RuleRemoveByModifierID, ModifierID: "rm_elf_archer_elemental_shot_armed", Limit: 1}`
      * `Ref`: `None`
  * **Branch[Earth]**:
    * **Effect[E0]** *(地之矢：对目标造成1点法术伤害③)*:
      * `EffectType`: `EffectDamage`
      * `Target`: `TargetCurrentCombat`
      * `Value`: `1`
      * `Ref`: `Combat.TargetID`
    * **Effect[E1]** *(清理“已发动”标记)*:
      * `EffectType`: `EffectRemoveRuleModifier`
      * `Target`: `TargetSelf`
      * `Value`: `0`
      * `RuleRemoveRef`: `{Mode: RuleRemoveByModifierID, ModifierID: "rm_elf_archer_elemental_shot_armed", Limit: 1}`
      * `Ref`: `None`

#### 【元素射击·行动后结算】 (Elemental Shot Action End Resolve)
* **说明**：由【元素射击】在①挂载的“已发动”标记驱动；在攻击行动结束时按攻击牌系别结算风之矢并清理标记。
* **1. 主干配置**：
  * **SkillID**: `skill_elf_archer_elemental_shot_action_end_resolve`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnActionEnd`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionEnd]`
  * **CustomExpression**: `Action.SourceID == Self.UserID && Action.CurrentType == Attack && State.HasRuleModifier(Self.UserID, "rm_elf_archer_elemental_shot_armed") == true && Combat.AttackCard.Element == Wind`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(风之矢：+1攻击行动)*:
    * `Implementation`: `model.AppendExtraAction`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `ActionRef`: `Attack`
    * `Ref`: `None`
  * **Effect[1]** *(结算后清理“已发动”标记)*:
    * `EffectType`: `EffectRemoveRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleRemoveRef`: `{Mode: RuleRemoveByModifierID, ModifierID: "rm_elf_archer_elemental_shot_armed", Limit: 1}`
    * `Ref`: `None`

#### 【动物伙伴】 (Animal Partner)
* **技能描述**：（你的回合内，目标角色承受你造成的伤害后⑥）你摸一张牌［强制］，弃一张牌。
* **1. 主干配置**：
  * **SkillID**: `skill_elf_archer_animal_partner`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnDamageTaken`
* **1.1 响应分组配置**：
  * **ResponseGroup.GroupID**: `rg_elf_archer_animal_response`
  * **ResponseGroup.Mode**: `ResponseGroupChooseOne`
  * **ResponseGroup.OptionOrder**: `10`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDraw]`
  * **CustomExpression**: `State.IsInSelfTurn(Self.UserID) == true && Combat.SourceID == Self.UserID && Combat.TargetID != Self.UserID && Combat.FinalDamage > 0 && State.IsSkillDisabled(Self.UserID, "skill_elf_archer_animal_partner") == false`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(你摸1)*:
    * `EffectType`: `EffectDrawCard`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `Ref`: `None`
  * **Effect[1]** *(你弃1)*:
    * `EffectType`: `EffectDiscard`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `Ref`: `None`

#### 【精灵密仪】 (Elven Secret Rite)
* **技能描述**：［持续］［宝石］角色［横置］转为［精灵祝福形态］，将牌库顶的3张牌面朝下放置于角色旁作为［祝福］，此形态下你的［祝福］可以视为手牌使用或打出。（你的回合结束时，若你未拥有［祝福］）［转正］脱离［精灵祝福形态］，对目标角色造成2点法术伤害③。
* **1. 主干配置**：
  * **SkillID**: `skill_elf_archer_elven_secret_rite`
  * **Category**: `Ultimate`
  * **Type**: `Startup`
  * **Timing**: `TimingStartup`
* **1.2 规则修饰器模板配置**：
  * **ModifierID**: `rm_elf_archer_secret_rite_blessing_as_hand`
  * **Domain**: `RuleModifierDomainCardSource`
  * **Priority**: `200`
  * **ConditionExpression**: `Self.Form == "elf_blessing_form"`
  * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
  * **CardSourcePayload**: `{ProjectionMode: CardSourceProjectionAsHand, FieldMarks: [Blessing]}`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionStart]`
  * **CustomExpression**: `None`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Gem, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(横置)*:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Tapped`
    * `Ref`: `None`
  * **Effect[1]** *(进入精灵祝福形态)*:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `"elf_blessing_form"`
    * `Ref`: `None`
  * **Effect[2]** *(施加“祝福视为手牌来源”的规则)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_elf_archer_secret_rite_blessing_as_hand"`
    * `RuleLifetimeRef`: `RuleLifePermanent`
    * `Ref`: `None`
  * **Effect[3]** *(将牌库顶3张放置为祝福)*:
    * `EffectType`: `EffectPlaceDeckTopAsFieldMark`
    * `Target`: `TargetSelf`
    * `Value`: `3`
    * `FieldMarkRef`: `Blessing`
    * `Ref`: `None`

#### 【精灵密仪·形态退场结算】 (Elven Secret Rite Cleanup)
* **说明**：将原技能“回合结束时若无祝福则退场并造成2点法术伤害”的延后段拆为同源被动子配置，避免引入技能耦合型新字段。
* **1. 主干配置**：
  * **SkillID**: `skill_elf_archer_elven_secret_rite_cleanup`
  * **Category**: `Ultimate`
  * **Type**: `Passive`
  * **Timing**: `TimingOnTurnEnd`
* **2. 前置条件**：
  * **PhaseLimit**: `[TurnEnd]`
  * **CustomExpression**: `State.IsInSelfTurn(Self.UserID) == true && Self.Orientation == Tapped && Self.Form == "elf_blessing_form" && State.CountFieldMark(Self.UserID, Blessing) == 0`
* **3. 目标选择规则**：
  * **SelectType**: `Enemy`
  * **MinCount**: `1`
  * **MaxCount**: `1`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(转正)*:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Normal`
    * `Ref`: `None`
  * **Effect[1]** *(脱离精灵祝福形态)*:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `nil`
    * `Ref`: `None`
  * **Effect[2]** *(对目标造成2点法术伤害)*:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelected`
    * `Value`: `2`
    * `Ref`: `None`

#### 【宠物强化】 (Pet Enhancement)
* **技能描述**：［水晶］（触发动物伙伴时）效果改为“目标角色摸1张牌［强制］，弃1张牌”。
* **1. 主干配置**：
  * **SkillID**: `skill_elf_archer_pet_enhancement`
  * **Category**: `Ultimate`
  * **Type**: `Response`
  * **Timing**: `TimingOnDamageTaken`
* **1.1 响应分组配置**：
  * **ResponseGroup.GroupID**: `rg_elf_archer_animal_response`
  * **ResponseGroup.Mode**: `ResponseGroupChooseOne`
  * **ResponseGroup.OptionOrder**: `20`
  * **ReplacesSkillIDs**: `["skill_elf_archer_animal_partner"]` *(在同一响应窗口中选中【宠物强化】时，替换并取消【动物伙伴】)*
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDraw]`
  * **CustomExpression**: `State.IsInSelfTurn(Self.UserID) == true && Combat.SourceID == Self.UserID && Combat.TargetID != Self.UserID && Combat.FinalDamage > 0 && State.IsSkillDisabled(Self.UserID, "skill_elf_archer_animal_partner") == false` *(与【动物伙伴】共享触发窗口：仅当“动物伙伴可触发”时允许发动宠物强化)*
* **3. 目标选择规则**：
  * **SelectType**: `Enemy`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **Filters**: `Target.UserID == Combat.TargetID` *(只能选择本次承受伤害的目标角色)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Crystal, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(目标摸1)*:
    * `EffectType`: `EffectDrawCard`
    * `Target`: `TargetSelected`
    * `Value`: `1`
    * `Ref`: `None`
  * **Effect[1]** *(目标弃1)*:
    * `EffectType`: `EffectDiscard`
    * `Target`: `TargetSelected`
    * `Value`: `1`
    * `Ref`: `None`

---

### 15. 瘟疫法师 (Plague Mage)

#### 【不朽】 (Immortal)
* **技能描述**：（［法术行动］结束时发动）你+1［治疗］。
* **1. 主干配置**：
  * **SkillID**: `skill_plague_mage_immortal`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnActionEnd`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionEnd]`
  * **CustomExpression**: `Action.SourceID == Self.UserID && Action.CurrentType == Magic && Event.CauseAction != "skill_plague_mage_touch_of_death"` *(满足“不能和死亡之触同时发动”：当本次法术行动就是【死亡之触】时，不触发【不朽】)*
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectHeal`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `Ref`: `None`

#### 【圣渎】 (Blasphemy)
* **技能描述**：你的［治疗］不能抵御攻击伤害，你的［治疗］上限+3。
* **1. 主干配置**：
  * **SkillID**: `skill_plague_mage_blasphemy`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnAttackDeclared` / `TimingOnTurnStart`
* **1.2 规则修饰器模板配置**：
  * **ModifierID**: `rm_plague_mage_blasphemy_max_heal_plus_3`
  * **Domain**: `RuleModifierDomainAttribute`
  * **Priority**: `100`
  * **ConditionExpression**: `None`
  * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
  * **AttrPayload**: `{AttrType: PlayerAttributeMaxHeal, Operation: AttributeModifyAdd, Value: 3}`
* **2. 前置条件**：
  * **PhaseLimit**: `None`
  * **CustomExpression**: `(Event.TriggerHook == TimingOnAttackDeclared && Combat.IsAttack == true && Combat.TargetID == Self.UserID) || (Event.TriggerHook == TimingOnTurnStart && Event.SourceType == SourceSystem && Event.CauseAction == "GameStart")`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(当你成为攻击目标时，本次攻击伤害不可被治疗抵御)*:
    * `EffectType`: `EffectApplyCombatTag`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `UnhealableDamage`
    * `Condition`: `Event.TriggerHook == TimingOnAttackDeclared && Combat.IsAttack == true && Combat.TargetID == Self.UserID`
    * `Ref`: `None`
  * **Effect[1]** *(游戏初始化时施加“治疗上限+3”常驻规则)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_plague_mage_blasphemy_max_heal_plus_3"`
    * `RuleLifetimeRef`: `RuleLifePermanent`
    * `Condition`: `Event.TriggerHook == TimingOnTurnStart && Event.SourceType == SourceSystem && Event.CauseAction == "GameStart"`
    * `Ref`: `None`

#### 【瘟疫】 (Plague)
* **技能描述**：（弃1张地系牌［展示］）对其他所有角色各造成1点法术伤害③；（若因此造成士气下降）回合结束时，你+1［治疗］。
* **1. 主干配置**：
  * **SkillID**: `skill_plague_mage_plague`
  * **Category**: `Normal`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `None`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Discards**:
    * `Count`: `1`
    * `Filter`: `{ReqElement: Earth}`
    * `Visibility`: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(对除自己外所有角色各造成1点法术伤害)*:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetAllOthers`
    * `Value`: `1`
    * `Ref`: `None`

#### 【瘟疫·回合结束奖励】 (Plague Turn-End Reward)
* **说明**：由 `SkillScopedMoraleDropTrace` 驱动；若本回合内【瘟疫】造成过士气下降，则在你的回合结束时你+1治疗。
* **1. 主干配置**：
  * **SkillID**: `skill_plague_mage_plague_turn_end_reward`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnTurnEnd`
* **2. 前置条件**：
  * **PhaseLimit**: `[TurnEnd]`
  * **CustomExpression**: `State.IsInSelfTurn(Self.UserID) == true && State.HasSkillMoraleDropThisTurn(Self.UserID, "skill_plague_mage_plague") == true`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectHeal`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `Ref`: `None`

#### 【死亡之触】 (Touch of Death)
* **技能描述**：（移除你的X［治疗］并弃Y张同系牌［展示］，X，Y的数值由你决定，但每项最少为2）对目标角色造成（X+Y-3）点法术伤害③；不能和不朽同时发动。
* **1. 主干配置**：
  * **SkillID**: `skill_plague_mage_touch_of_death`
  * **Category**: `Normal`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Self.Heal >= 2 && Self.HandCount >= 2`
* **3. 目标选择规则**：
  * **SelectType**: `Enemy`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **NamedValueConstraints**:
    * `{Key: "X", Required: true, MinExpression: "2", MaxExpression: "Self.Heal"}`
    * `{Key: "Y", Required: true, MinExpression: "2", MaxExpression: "Self.HandCount"}`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **HealCost**: `Action.NamedValues["X"]`
  * **Discards**:
    * `Count`: `Action.NamedValues["Y"]`
    * `Filter`: `{SameAttribute: MatchElement}`
    * `Visibility`: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelected`
    * `Value`: `Action.NamedValues["X"] + Action.NamedValues["Y"] - 3`
    * `Ref`: `None`

#### 【剧毒新星】 (Toxic Nova)
* **技能描述**：［宝石］对其他所有角色各造成2点法术伤害③，你+1［治疗］。
* **1. 主干配置**：
  * **SkillID**: `skill_plague_mage_toxic_nova`
  * **Category**: `Ultimate`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `None`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Gem, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(对除自己外所有角色各造成2点法术伤害)*:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetAllOthers`
    * `Value`: `2`
    * `Ref`: `None`
  * **Effect[1]** *(你+1治疗)*:
    * `EffectType`: `EffectHeal`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `Ref`: `None`

---

### 16. 魔剑士 (Magic Swordsman)

#### 【修罗连斩】 (Asura Combo Slash)
* **技能描述**：［回合限定］（［攻击行动］结束时发动）额外+1火系［攻击行动］。
* **1. 主干配置**：
  * **SkillID**: `skill_magic_swordsman_asura_combo_slash`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnActionEnd`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionEnd]`
  * **IsTurnLimited**: `true`
  * **CustomExpression**: `Action.SourceID == Self.UserID && Action.CurrentType == Attack`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `Implementation`: `model.AppendExtraAction`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `ActionRef`: `Attack`
    * `ElementRef`: `Fire`
    * `Ref`: `None`

#### 【暗影凝聚】 (Shadow Convergence)
* **技能描述**：（对自己造成1点法术伤害③）［横置］持续到你的下个行动阶段开始，你都处于［暗影形态］，脱离［暗影形态］时角色［转正］。
* **1. 主干配置**：
  * **SkillID**: `skill_magic_swordsman_shadow_convergence`
  * **Category**: `Normal`
  * **Type**: `Startup`
  * **Timing**: `TimingStartup`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionStart]`
  * **CustomExpression**: `Self.Orientation == Normal`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(对自己造成1点法术伤害)*:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `Ref`: `None`
  * **Effect[1]** *(横置)*:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Tapped`
    * `Ref`: `None`
  * **Effect[2]** *(进入暗影形态)*:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `"magic_swordsman_shadow_form"`
    * `Ref`: `None`

#### 【暗影凝聚·形态退场结算】 (Shadow Convergence Cleanup)
* **说明**：在你的下个行动阶段开始前，结束暗影形态并转正。
* **1. 主干配置**：
  * **SkillID**: `skill_magic_swordsman_shadow_convergence_cleanup`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnBeforeAction`
* **2. 前置条件**：
  * **PhaseLimit**: `[BeforeAction]`
  * **CustomExpression**: `State.IsInSelfTurn(Self.UserID) == true && Self.Orientation == Tapped && Self.Form == "magic_swordsman_shadow_form"`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Normal`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `nil`
    * `Ref`: `None`

#### 【暗影之力】 (Shadow Power)
* **技能描述**：（仅暗影形态下）你发动的所有攻击伤害额外+1。
* **1. 主干配置**：
  * **SkillID**: `skill_magic_swordsman_shadow_power`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnAttackDeclared`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDeclare]`
  * **CustomExpression**: `Self.Form == "magic_swordsman_shadow_form" && Combat.SourceID == Self.UserID && Combat.IsAttack == true`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectAttackDamageModifier`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `1`
    * `Ref`: `None`

#### 【暗影抗拒】 (Shadow Rejection)
* **技能描述**：在你的行动阶段，你不能主动使用【圣光】【魔弹】【圣盾】【虚弱】【中毒】这五张基础法术牌；非你行动阶段的响应使用不受此限制。
* **1. 主干配置**：
  * **SkillID**: `skill_magic_swordsman_shadow_rejection`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingBeforeActionExecute`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `State.IsInSelfTurn(Self.UserID) == true && Action.SourceID == Self.UserID && Action.CurrentType == Magic && Action.PlayedCard != nil && (Action.PlayedCard.Name == "圣光" || Action.PlayedCard.Name == "魔弹" || Action.PlayedCard.Name == "圣盾" || Action.PlayedCard.Name == "虚弱" || Action.PlayedCard.Name == "中毒")` *(按 CardName 精确匹配，仅拦截自己行动阶段对这5张基础法术牌的主动使用；不影响非本回合响应)*
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **3.1 行动改写配置**：
  * **ActionTransform.Hook**: `TimingBeforeActionExecute`
  * **ActionTransform.Optional**: `false`
  * **ActionTransform.Priority**: `200`
  * **ActionTransform.CancelCurrentAction**: `true`
  * **ActionTransform.Match.RequireActionType**: `Magic`
  * **ActionTransform.Match.RequirePlayedCardTypes**: `[]`
  * **ActionTransform.Match.RequirePlayedCardElements**: `[]`
  * **ActionTransform.Match.ExcludeTemplateIDs**: `[]`
  * **ActionTransform.Rewrite**: `nil` *(命中时直接取消本次法术行动)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * `None` *(本技能通过 ActionTransform 生效)*

#### 【暗影流星】 (Shadow Meteor)
* **技能描述**：（仅暗影形态下发动，弃2张法术牌［展示］）对目标角色造成2点法术伤害③；（若你额外移除我方［战绩区］2星石）［转正］脱离［暗影形态］，你+1［宝石］。
* **1. 主干配置**：
  * **SkillID**: `skill_magic_swordsman_shadow_meteor`
  * **Category**: `Normal`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Self.Form == "magic_swordsman_shadow_form"`
* **3. 目标选择规则**：
  * **SelectType**: `Enemy`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **SelectedValueRule**: `Action.SelectedValue == 0 || Action.SelectedValue == 1` *(0=不额外移除战绩区星石；1=额外移除2星石并触发退场增益)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Discards**:
    * `Count`: `2`
    * `Filter`: `{ReqCardType: Magic}`
    * `Visibility`: `VisibilityPublic`
  * **Stones**: `[{Type: Any, Amount: Action.SelectedValue == 1 ? 2 : 0}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelected`
    * `Value`: `2`
    * `Ref`: `None`
  * **Effect[1]** *(额外移除战绩区2星石后：转正脱离暗影形态)*:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Normal`
    * `Condition`: `Action.SelectedValue == 1`
    * `Ref`: `None`
  * **Effect[2]**:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `nil`
    * `Condition`: `Action.SelectedValue == 1`
    * `Ref`: `None`
  * **Effect[3]**:
    * `EffectType`: `EffectAddEnergyStone`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `StoneRef`: `Gem`
    * `Condition`: `Action.SelectedValue == 1`
    * `Ref`: `None`

#### 【黄泉震颤】 (Yellow Spring Tremor)
* **技能描述**：［回合限定］［宝石］（主动攻击前发动①）本次攻击对手不能应战，（若命中②）你将手牌补至上限，然后弃2张牌。
* **1. 主干配置**：
  * **SkillID**: `skill_magic_swordsman_yellow_spring_tremor`
  * **Category**: `Ultimate`
  * **Type**: `Response`
  * **Timing**: `TimingOnAttackDeclared`
* **1.2 规则修饰器模板配置**：
  * **ModifierID**: `rm_magic_swordsman_yellow_spring_armed`
  * **Domain**: `RuleModifierDomainCardSource`
  * **Priority**: `60`
  * **ConditionExpression**: `None`
  * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
  * **CardSourcePayload**: `{ProjectionMode: CardSourceProjectionAsHand, FieldMarks: []}` *(空投影负载，仅作本次战斗的“已发动”标记)*
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDeclare]`
  * **IsTurnLimited**: `true`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Gem, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectApplyCombatTag`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `Unrespondable`
    * `Ref`: `None`
  * **Effect[1]** *(挂载“黄泉震颤已发动”标记，供命中后子技能读取)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_magic_swordsman_yellow_spring_armed"`
    * `RuleLifetimeRef`: `RuleLifeUntilCombatEnd`
    * `Ref`: `None`

#### 【黄泉震颤·命中后结算】 (Yellow Spring Tremor Hit Resolve)
* **说明**：由【黄泉震颤】在①挂载标记驱动；在②命中时执行“补到手牌上限后弃2”。
* **1. 主干配置**：
  * **SkillID**: `skill_magic_swordsman_yellow_spring_tremor_hit_resolve`
  * **Category**: `Ultimate`
  * **Type**: `Passive`
  * **Timing**: `TimingOnHitCheck`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatHitCheck]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && Combat.IsHit == true && State.HasRuleModifier(Self.UserID, "rm_magic_swordsman_yellow_spring_armed") == true`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(手牌补到上限)*:
    * `EffectType`: `EffectDrawCard`
    * `Target`: `TargetSelf`
    * `Value`: `Self.HandCount < State.GetEffectivePlayerAttribute(Self.UserID, PlayerAttributeMaxHand) ? (State.GetEffectivePlayerAttribute(Self.UserID, PlayerAttributeMaxHand) - Self.HandCount) : 0`
    * `Ref`: `None`
  * **Effect[1]** *(然后弃2张牌)*:
    * `EffectType`: `EffectDiscard`
    * `Target`: `TargetSelf`
    * `Value`: `2`
    * `Ref`: `None`

---

### 17. 血色剑灵 (Blood Sword Spirit)

#### 【血色荆棘】 (Blood Thorns)
* **技能描述**：（攻击命中时②）你+1［鲜血］。
* **1. 主干配置**：
  * **SkillID**: `skill_blood_sword_spirit_blood_thorns`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnHitCheck`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatHitCheck]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsHit == true`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `TokenRef`: `Blood`
    * `Ref`: `None`

#### 【赤色一闪】 (Crimson Flash)
* **技能描述**：（［攻击行动］结束后，移除1点［鲜血］）对自己造成2点法术伤害③，额外+1［攻击行动］。
* **1. 主干配置**：
  * **SkillID**: `skill_blood_sword_spirit_crimson_flash`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnActionEnd`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionEnd]`
  * **CustomExpression**: `Action.SourceID == Self.UserID && Action.CurrentType == Attack && Self.Tokens[Blood] >= 1`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Tokens**: `[{Type: Blood, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelf`
    * `Value`: `2`
    * `Ref`: `None`
  * **Effect[1]**:
    * `Implementation`: `model.AppendExtraAction`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `ActionRef`: `Attack`
    * `Ref`: `None`

#### 【血染蔷薇】 (Bloodstained Rose)
* **技能描述**：移除2点［鲜血］发动，移除目标角色2点［治疗］，将我方阵营的1［水晶］翻面为1［宝石］，再选择任意1名队友+1［治疗］。（若［血蔷薇庭院］在场）额外对所有角色各造成1点法术伤害③。
* **1. 主干配置**：
  * **SkillID**: `skill_blood_sword_spirit_bloodstained_rose`
  * **Category**: `Normal`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Self.Tokens[Blood] >= 2`
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `2`
  * **MaxCount**: `2`
  * **Filters**: `Action.Targets[0] 为移除治疗目标（任意角色）；Action.Targets[1] 为治疗目标（必须我方角色）`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Tokens**: `[{Type: Blood, Amount: 2}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(Targets[0] 移除2点治疗)*:
    * `EffectType`: `EffectHeal`
    * `Target`: `TargetSelected`
    * `Value`: `-2`
    * `Ref`: `Action.Targets[0].TargetUserID`
  * **Effect[1]** *(我方阵营战绩区1水晶转1宝石；无水晶时跳过)*:
    * `EffectType`: `EffectConvertTeamStone`
    * `Target`: `TargetSelfTeam`
    * `Value`: `1`
    * `StoneRef`: `Crystal`
    * `StoneToRef`: `Gem`
    * `Ref`: `None`
  * **Effect[2]** *(Targets[1] +1治疗)*:
    * `EffectType`: `EffectHeal`
    * `Target`: `TargetSelected`
    * `Value`: `1`
    * `Ref`: `Action.Targets[1].TargetUserID`
  * **Effect[3]** *(若血蔷薇庭院在场：全场各受1点法术伤害)*:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetAllPlayers`
    * `Value`: `1`
    * `Condition`: `State.HasRuleModifier(Self.UserID, "rm_blood_sword_spirit_blood_rose_courtyard_active") == true`
    * `Ref`: `None`

#### 【血气屏障】 (Blood Barrier)
* **技能描述**：（目标角色对你造成法术伤害时③，移除1点［鲜血］）本次法术伤害-1③，对目标对手造成1点法术伤害③。
* **1. 主干配置**：
  * **SkillID**: `skill_blood_sword_spirit_blood_barrier`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnDamageCalculated`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatCalcDamage]`
  * **CustomExpression**: `Combat.TargetID == Self.UserID && Combat.IsMagic == true && Combat.SourceID != Self.UserID && Self.Tokens[Blood] >= 1`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Tokens**: `[{Type: Blood, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(本次法术伤害-1，最低不低于0)*:
    * `EffectType`: `EffectModifyPendingDamage`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `-1`
    * `Ref`: `None`
  * **Effect[1]** *(对伤害来源角色造成1点法术伤害)*:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetTriggerSource`
    * `Value`: `1`
    * `Ref`: `None`

#### 【血蔷薇庭院】 (Blood Rose Courtyard)
* **技能描述**：此卡在场时，所有角色的［治疗］无法用于抵御伤害；（血色剑灵的回合结束时）移除此卡。
* **1. 主干配置**：
  * **SkillID**: `skill_blood_sword_spirit_blood_rose_courtyard`
  * **Category**: `Exclusive`
  * **Type**: `Passive`
  * **Timing**: `TimingOnCardPlayedOrRevealed`
* **1.2 规则修饰器模板配置**：
  * **ModifierID**: `rm_blood_sword_spirit_blood_rose_courtyard_active`
  * **Domain**: `RuleModifierDomainCardSource`
  * **Priority**: `80`
  * **ConditionExpression**: `None`
  * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
  * **CardSourcePayload**: `{ProjectionMode: CardSourceProjectionAsHand, FieldMarks: []}` *(空载荷：仅作为“庭院在场”的通用状态标记)*
* **2. 前置条件**：
  * **PhaseLimit**: `None`
  * **CustomExpression**: `Action.PlayedCard.CharacterSkillMap[Self.CharacterID] == "skill_blood_sword_spirit_blood_rose_courtyard"`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(挂载“庭院在场”标记；到血色剑灵回合结束移除)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_blood_sword_spirit_blood_rose_courtyard_active"`
    * `RuleLifetimeRef`: `RuleLifeUntilTurnEnd`
    * `Ref`: `None`

#### 【血蔷薇庭院·全场禁疗抵伤】 (Blood Rose Courtyard Aura)
* **说明**：当【血蔷薇庭院】在场时，为当前战斗写入 `UnhealableDamage`，使所有角色都无法以治疗抵御该次伤害。
* **1. 主干配置**：
  * **SkillID**: `skill_blood_sword_spirit_blood_rose_courtyard_aura`
  * **Category**: `Exclusive`
  * **Type**: `Passive`
  * **Timing**: `TimingOnAttackDeclared` / `TimingOnMagicDeclared`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDeclare]`
  * **CustomExpression**: `State.HasRuleModifier(Self.UserID, "rm_blood_sword_spirit_blood_rose_courtyard_active") == true`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectApplyCombatTag`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `UnhealableDamage`
    * `Ref`: `None`

#### 【散华轮舞】 (Scattered Blossom Waltz)
* **技能描述**：你选择以下一项发动：●［水晶］将［血蔷薇庭院］放置于场上，你+2［鲜血］。●［宝石］将［血蔷薇庭院］放置于场上，无视你的［鲜血］上限为你+2［鲜血］，但你的［鲜血］数最高为4，你弃到4张牌。
* **1. 主干配置**：
  * **SkillID**: `skill_blood_sword_spirit_scattered_blossom_waltz`
  * **Category**: `Ultimate`
  * **Type**: `Startup`
  * **Timing**: `TimingStartup`
* **1.2 规则修饰器模板配置**：
  * **ModifierID**: `rm_blood_sword_spirit_scattered_blossom_ignore_blood_cap_abs_4`
  * **Domain**: `RuleModifierDomainTokenPolicy`
  * **Priority**: `200`
  * **ConditionExpression**: `None`
  * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
  * **TokenPolicyPayload**: `{TokenType: Blood, ApplyMode: TokenApplyIgnoreMax, AbsoluteMax: 4}`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionStart]`
  * **CustomExpression**: `Action.SelectedValue == 0 || Action.SelectedValue == 1` *(0=水晶分支；1=宝石分支)*
* **3. 目标选择规则**：
  * **SelectType**: `None`
  * **SelectedValueRule**: `Action.SelectedValue == 0 || Action.SelectedValue == 1`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Crystal, Amount: Action.SelectedValue == 0 ? 1 : 0}, {Type: Gem, Amount: Action.SelectedValue == 1 ? 1 : 0}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(两分支共用：放置血蔷薇庭院到本回合结束)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_blood_sword_spirit_blood_rose_courtyard_active"`
    * `RuleLifetimeRef`: `RuleLifeUntilTurnEnd`
    * `Ref`: `None`
  * **Effect[1]** *(水晶分支：你+2鲜血)*:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `2`
    * `TokenRef`: `Blood`
    * `Condition`: `Action.SelectedValue == 0`
    * `Ref`: `None`
  * **Effect[2]** *(宝石分支：施加“鲜血无视上限但绝对封顶4”的策略，仅作用于本技能执行链)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_blood_sword_spirit_scattered_blossom_ignore_blood_cap_abs_4"`
    * `RuleLifetimeRef`: `RuleLifeThisEffectChain`
    * `Condition`: `Action.SelectedValue == 1`
    * `Ref`: `None`
  * **Effect[3]** *(宝石分支：你+2鲜血)*:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `2`
    * `TokenRef`: `Blood`
    * `Condition`: `Action.SelectedValue == 1`
    * `Ref`: `None`
  * **Effect[4]** *(宝石分支：弃到4张牌)*:
    * `EffectType`: `EffectAdjustHand`
    * `Target`: `TargetSelf`
    * `Value`: `4`
    * `Condition`: `Action.SelectedValue == 1`
    * `Ref`: `None`

---

### 18. 祈祷师 (Prayer Master)

#### 【光辉信仰】 (Radiant Faith)
* **技能描述**：（仅祈祷形态状态下发动，移除1点［祈祷符文］）你弃2张牌，我方战绩区+1［宝石］，目标队友+1［治疗］。
* **1. 主干配置**：
  * **SkillID**: `skill_prayer_master_radiant_faith`
  * **Category**: `Normal`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Self.Form == "prayer_master_prayer_form" && Self.Tokens[Rune] >= 1`
* **3. 目标选择规则**：
  * **SelectType**: `TeamOther`
  * **MinCount**: `1`
  * **MaxCount**: `1`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Tokens**: `[{Type: Rune, Amount: 1}]`
  * **Discards**:
    * `Count`: `2`
    * `Filter`: `{}`
    * `Visibility`: `VisibilityHidden`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectAddTeamStone`
    * `Target`: `TargetSelfTeam`
    * `Value`: `1`
    * `StoneRef`: `Gem`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectHeal`
    * `Target`: `TargetSelected`
    * `Value`: `1`
    * `Ref`: `None`

#### 【黑暗诅咒】 (Dark Curse)
* **技能描述**：（仅祈祷形态状态下发动，移除1点［祈祷符文］）对目标角色和自己各造成2点法术伤害③。
* **1. 主干配置**：
  * **SkillID**: `skill_prayer_master_dark_curse`
  * **Category**: `Normal`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Self.Form == "prayer_master_prayer_form" && Self.Tokens[Rune] >= 1`
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `1`
  * **MaxCount**: `1`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Tokens**: `[{Type: Rune, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelected`
    * `Value`: `2`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelf`
    * `Value`: `2`
    * `Ref`: `None`

#### 【威力赐福】 (Power Blessing)
* **技能描述**：（将威力赐福放置于目标队友面前）该队友获得（攻击命中后可以移除此牌发动②）本次攻击伤害额外+2。
* **1. 主干配置**：
  * **SkillID**: `skill_prayer_master_power_blessing`
  * **Category**: `Unique`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Action.PlayedCard.CharacterSkillMap[Self.CharacterID] == "skill_prayer_master_power_blessing"`
* **3. 目标选择规则**：
  * **SelectType**: `TeamOther`
  * **MinCount**: `1`
  * **MaxCount**: `1`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectPlacePlayedCardAsFieldMark`
    * `Target`: `TargetSelected`
    * `Value`: `1`
    * `FieldMarkRef`: `PowerBlessing`
    * `VisibilityRef`: `VisibilityPublic`
    * `Ref`: `None`

#### 【威力赐福·触发】 (Power Blessing Trigger)
* **说明**：当我方持有【威力赐福】标记的角色攻击命中后，可选择移除该标记以使本次攻击伤害额外+2。
* **实现约定**：此“可选择是否移除”由【威力赐福】持有者本人在其攻击命中窗口自行响应发动。
* **1. 主干配置**：
  * **SkillID**: `skill_prayer_master_power_blessing_trigger`
  * **Category**: `Unique`
  * **Type**: `Response`
  * **Timing**: `TimingOnHitCheck`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatHitCheck]`
  * **RequireSourceType**: `SourcePlayer`
  * **RequireOperator**: `OperatorSelf`
  * **CustomExpression**: `Combat.IsAttack == true && Combat.IsHit == true && State.CountFieldMark(Self.UserID, PowerBlessing) >= 1`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectRemoveFieldMark`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `FieldMarkRef`: `PowerBlessing`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectAttackDamageModifier`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `2`
    * `Ref`: `None`

#### 【迅捷赐福】 (Swift Blessing)
* **技能描述**：（将迅捷赐福放置于目标队友面前）该队友获得（［法术行动］或［攻击行动］结束时可以移除此牌发动）额外+1［攻击行动］。
* **1. 主干配置**：
  * **SkillID**: `skill_prayer_master_swift_blessing`
  * **Category**: `Unique`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Action.PlayedCard.CharacterSkillMap[Self.CharacterID] == "skill_prayer_master_swift_blessing"`
* **3. 目标选择规则**：
  * **SelectType**: `TeamOther`
  * **MinCount**: `1`
  * **MaxCount**: `1`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectPlacePlayedCardAsFieldMark`
    * `Target`: `TargetSelected`
    * `Value`: `1`
    * `FieldMarkRef`: `SwiftBlessing`
    * `VisibilityRef`: `VisibilityPublic`
    * `Ref`: `None`

#### 【迅捷赐福·触发】 (Swift Blessing Trigger)
* **说明**：当我方持有【迅捷赐福】标记的角色完成一次攻击行动或法术行动后，可选择移除该标记以额外+1攻击行动。
* **实现约定**：此“可选择是否移除”由【迅捷赐福】持有者本人在其行动结束窗口自行响应发动。
* **1. 主干配置**：
  * **SkillID**: `skill_prayer_master_swift_blessing_trigger`
  * **Category**: `Unique`
  * **Type**: `Response`
  * **Timing**: `TimingOnActionEnd`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionEnd]`
  * **RequireSourceType**: `SourcePlayer`
  * **RequireOperator**: `OperatorSelf`
  * **CustomExpression**: `(Action.CurrentType == Attack || Action.CurrentType == Magic) && State.CountFieldMark(Self.UserID, SwiftBlessing) >= 1`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectRemoveFieldMark`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `FieldMarkRef`: `SwiftBlessing`
    * `Ref`: `None`
  * **Effect[1]**:
    * `Implementation`: `model.AppendExtraAction`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `ActionRef`: `Attack`
    * `Ref`: `None`

#### 【祈祷】 (Prayer)
* **技能描述**：［持续］［宝石］［横置］转为［祈祷形态］，在此形态下，你每发动一次主动攻击①，你+2［祈祷符文］。
* **1. 主干配置**：
  * **SkillID**: `skill_prayer_master_prayer`
  * **Category**: `Ultimate`
  * **Type**: `Startup`
  * **Timing**: `TimingStartup`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionStart]`
  * **CustomExpression**: `Self.Orientation == Normal`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Gem, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Tapped`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `"prayer_master_prayer_form"`
    * `Ref`: `None`
  * **Effect[2]**:
    * `EffectType`: `EffectSetHandLimitFixed`
    * `Target`: `TargetSelf`
    * `Value`: `5`
    * `Ref`: `None`

#### 【祈祷·攻击增符】 (Prayer Attack Rune Gain)
* **说明**：祈祷形态下，每次你发动主动攻击时，你+2［祈祷符文］。
* **1. 主干配置**：
  * **SkillID**: `skill_prayer_master_prayer_attack_rune_gain`
  * **Category**: `Ultimate`
  * **Type**: `Passive`
  * **Timing**: `TimingOnAttackDeclared`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDeclare]`
  * **CustomExpression**: `Self.Form == "prayer_master_prayer_form" && Combat.SourceID == Self.UserID && Combat.IsActiveAttack == true`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `2`
    * `TokenRef`: `Rune`
    * `Ref`: `None`

#### 【法力潮汐】 (Mana Tide)
* **技能描述**：［回合限定］［水晶］（［法术行动］结束时发动）额外+1［法术行动］。
* **1. 主干配置**：
  * **SkillID**: `skill_prayer_master_mana_tide`
  * **Category**: `Ultimate`
  * **Type**: `Response`
  * **Timing**: `TimingOnActionEnd`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionEnd]`
  * **IsTurnLimited**: `true`
  * **CustomExpression**: `Action.SourceID == Self.UserID && Action.CurrentType == Magic`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Crystal, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `Implementation`: `model.AppendExtraAction`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `ActionRef`: `Magic`
    * `Ref`: `None`

---

### 19. 红莲骑士 (Crimson Knight)

#### 【腥红圣约】 (Crimson Covenant)
* **技能描述**：［回合限定］（主动攻击时发动①）你+1［治疗］。
* **1. 主干配置**：
  * **SkillID**: `skill_crimson_knight_crimson_covenant`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnAttackDeclared`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDeclare]`
  * **IsTurnLimited**: `true`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectHeal`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `Ref`: `None`

#### 【腥红信仰】 (Crimson Faith)
* **技能描述**：你的［治疗］只能抵御自己造成的伤害，你的［治疗］上限+2。
* **1. 主干配置**：
  * **SkillID**: `skill_crimson_knight_crimson_faith`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnAttackDeclared` / `TimingOnMagicDeclared` / `TimingOnTurnStart`
* **1.2 规则修饰器模板配置**：
  * **ModifierID**: `rm_crimson_knight_crimson_faith_max_heal_plus_2`
  * **Domain**: `RuleModifierDomainAttribute`
  * **Priority**: `100`
  * **ConditionExpression**: `None`
  * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
  * **AttrPayload**: `{AttrType: PlayerAttributeMaxHeal, Operation: AttributeModifyAdd, Value: 2}`
* **2. 前置条件**：
  * **PhaseLimit**: `None`
  * **CustomExpression**: `((Event.TriggerHook == TimingOnAttackDeclared || Event.TriggerHook == TimingOnMagicDeclared) && Combat.TargetID == Self.UserID && Combat.SourceID != Self.UserID) || (Event.TriggerHook == TimingOnTurnStart && Event.SourceType == SourceSystem && Event.CauseAction == "GameStart")`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(来自他人的伤害对你不可用治疗抵御)*:
    * `EffectType`: `EffectApplyCombatTag`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `UnhealableDamage`
    * `Condition`: `(Event.TriggerHook == TimingOnAttackDeclared || Event.TriggerHook == TimingOnMagicDeclared) && Combat.TargetID == Self.UserID && Combat.SourceID != Self.UserID`
    * `Ref`: `None`
  * **Effect[1]** *(游戏初始化时施加“治疗上限+2”常驻规则)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_crimson_knight_crimson_faith_max_heal_plus_2"`
    * `RuleLifetimeRef`: `RuleLifePermanent`
    * `Condition`: `Event.TriggerHook == TimingOnTurnStart && Event.SourceType == SourceSystem && Event.CauseAction == "GameStart"`
    * `Ref`: `None`

#### 【血腥祷言】 (Bloody Prayer)
* **技能描述**：（移除你的X点［治疗］，对自己造成X点法术伤害③）任意分配X点［治疗］给1~2名队友，你+1［血印］。
* **1. 主干配置**：
  * **SkillID**: `skill_crimson_knight_bloody_prayer`
  * **Category**: `Normal`
  * **Type**: `Startup`
  * **Timing**: `TimingStartup`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionStart]`
  * **CustomExpression**: `Self.Heal >= 1`
* **3. 目标选择规则**：
  * **SelectType**: `TeamOther`
  * **MinCount**: `1`
  * **MaxCount**: `2`
  * **RequireTargetAllocations**: `true`
  * **AllocationTotal**: `Action.NamedValues["X"]`
  * **MinAllocationPerTarget**: `1`
  * **MaxAllocationPerTarget**: `Action.NamedValues["X"]`
  * **NamedValueConstraints**:
    * `{Key: "X", Required: true, MinExpression: "1", MaxExpression: "Self.Heal"}`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **HealCost**: `Action.NamedValues["X"]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(对自己造成X点法术伤害)*:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelf`
    * `Value`: `Action.NamedValues["X"]`
    * `Ref`: `None`
  * **Effect[1]** *(任意分配X点治疗给1~2名队友)*:
    * `EffectType`: `EffectHeal`
    * `Target`: `TargetSelected`
    * `Value`: `Action.TargetAllocations[Target.UserID]`
    * `Ref`: `None`
  * **Effect[2]**:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `TokenRef`: `Blood`
    * `Ref`: `None`

#### 【杀戮盛宴】 (Slaughter Feast)
* **技能描述**：（主动攻击命中后发动②，移除1点［血印］，对自己造成4点法术伤害③）本次攻击伤害额外+2。
* **1. 主干配置**：
  * **SkillID**: `skill_crimson_knight_slaughter_feast`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnHitCheck`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatHitCheck]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && Combat.IsHit == true && Self.Tokens[BloodMark] >= 1`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Tokens**: `[{Type: Blood, Amount: 1}]`
  * **HPCost**: `4`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectAttackDamageModifier`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `2`
    * `Ref`: `None`

#### 【热血沸腾】 (Hot Blooded)
* **技能描述**：（当你因承受伤害而导致我方士气下降时强制发动）［横置］转为［热血沸腾形态］，该形态下你承受伤害不会导致我方士气下降。在你的回合结束阶段，若你处于此形态，［转正］并脱离此形态，你+2［治疗］。
* **1. 主干配置**：
  * **SkillID**: `skill_crimson_knight_hot_blooded`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnDamageTaken`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDraw]`
  * **CustomExpression**: `Combat.TargetID == Self.UserID && Event.PendingMoraleLoss > 0 && Self.Form != "crimson_knight_hot_blooded_form"`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Tapped`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `"crimson_knight_hot_blooded_form"`
    * `Ref`: `None`

#### 【热血沸腾·形态免士气】 (Hot Blooded Morale Shield)
* **说明**：处于热血沸腾形态时，你承受伤害不会导致我方士气下降。
* **1. 主干配置**：
  * **SkillID**: `skill_crimson_knight_hot_blooded_no_morale_drop`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnDamageTaken`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDraw]`
  * **CustomExpression**: `Self.Form == "crimson_knight_hot_blooded_form" && Combat.TargetID == Self.UserID && Event.PendingMoraleLoss > 0`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectReducePendingMoraleLoss`
    * `Target`: `TargetCurrentEvent`
    * `Value`: `Event.PendingMoraleLoss`
    * `Ref`: `None`

#### 【热血沸腾·回合结束退场】 (Hot Blooded Turn-End Cleanup)
* **说明**：你的回合结束时，若仍处于热血沸腾形态，则转正脱离并+2治疗。
* **1. 主干配置**：
  * **SkillID**: `skill_crimson_knight_hot_blooded_turn_end_cleanup`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnTurnEnd`
* **2. 前置条件**：
  * **PhaseLimit**: `[TurnEnd]`
  * **CustomExpression**: `State.IsInSelfTurn(Self.UserID) == true && Self.Form == "crimson_knight_hot_blooded_form"`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Normal`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `nil`
    * `Ref`: `None`
  * **Effect[2]**:
    * `EffectType`: `EffectHeal`
    * `Target`: `TargetSelf`
    * `Value`: `2`
    * `Ref`: `None`

#### 【戒骄戒躁】 (Stay Humble)
* **技能描述**：［水晶］（仅热血沸腾形态下，［攻击行动］或［法术行动］结束时发动）［转正］脱离［热血沸腾形态］，额外+1［攻击行动］或［法术行动］。
* **1. 主干配置**：
  * **SkillID**: `skill_crimson_knight_stay_humble`
  * **Category**: `Ultimate`
  * **Type**: `Response`
  * **Timing**: `TimingOnActionEnd`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionEnd]`
  * **CustomExpression**: `Self.Form == "crimson_knight_hot_blooded_form" && Action.SourceID == Self.UserID && (Action.CurrentType == Attack || Action.CurrentType == Magic) && (Action.ActionRef == Attack || Action.ActionRef == Magic)`
* **3. 目标选择规则**：
  * **SelectType**: `None`
  * **AllowedActionRefs**: `[Attack, Magic]`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Crystal, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Normal`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `nil`
    * `Ref`: `None`
  * **Effect[2]**:
    * `Implementation`: `model.AppendExtraAction`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `ActionRef`: `Action.ActionRef`
    * `Ref`: `None`

#### 【腥红十字】 (Crimson Cross)
* **技能描述**：［水晶］（移除1点［血印］，弃2张法术牌［展示］，对自己造成4点法术伤害③）对目标角色造成3点法术伤害③。
* **1. 主干配置**：
  * **SkillID**: `skill_crimson_knight_crimson_cross`
  * **Category**: `Ultimate`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Self.Tokens[BloodMark] >= 1`
* **3. 目标选择规则**：
  * **SelectType**: `Enemy`
  * **MinCount**: `1`
  * **MaxCount**: `1`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Crystal, Amount: 1}]`
  * **Tokens**: `[{Type: BloodMark, Amount: 1}]`
  * **Discards**:
    * `Count`: `2`
    * `Filter`: `{ReqCardType: Magic}`
    * `Visibility`: `VisibilityPublic`
  * **HPCost**: `4`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelected`
    * `Value`: `3`
    * `Ref`: `None`

---

### 20. 英灵人形 (Heroic Puppet)

#### 【战纹掌控】 (War Rune Mastery)
* **技能描述**：游戏初始时，你获得3个［战纹］。
* **1. 主干配置**：
  * **SkillID**: `skill_heroic_puppet_war_rune_mastery`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnTurnStart`
* **2. 前置条件**：
  * **PhaseLimit**: `[GameInit]`
  * **CustomExpression**: `Event.SourceType == SourceSystem && Event.CauseAction == "GameStart"`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `3`
    * `TokenRef`: `WarRune`
    * `Ref`: `None`

#### 【怒火压制】 (Rage Suppression)
* **技能描述**：（主动攻击未命中时②）翻转1个［战纹］，不能和［魔纹融合］同时发动。
* **1. 主干配置**：
  * **SkillID**: `skill_heroic_puppet_rage_suppression`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnHitCheck`
* **1.1 响应分组配置**：
  * **ResponseGroup.GroupID**: `rg_heroic_puppet_on_miss`
  * **ResponseGroup.Mode**: `ResponseGroupChooseOne`
  * **ResponseGroup.OptionOrder**: `10`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatHitCheck]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && Combat.IsHit == false && Self.Tokens[WarRune] >= 1`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(翻转1战纹=移除1战纹)*:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `-1`
    * `TokenRef`: `WarRune`
    * `Ref`: `None`
  * **Effect[1]** *(翻转1战纹=增加1魔纹)*:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `TokenRef`: `MagicRune`
    * `Ref`: `None`

#### 【战纹碎击】 (War Rune Smash)
* **技能描述**：（主动攻击命中时②，翻转1个［战纹］，弃X张同系牌［展示］）本次攻击伤害额外+（X-1）；（若你处于［蓄势迸发形态］下，额外翻转Y个战纹）本次法术伤害额外+Y。
* **1. 主干配置**：
  * **SkillID**: `skill_heroic_puppet_war_rune_smash`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnHitCheck`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatHitCheck]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && Combat.IsHit == true && Self.Tokens[WarRune] >= 1 && (Self.Form == "heroic_puppet_surge_form" || Action.NamedValues["Y"] == 0) && Self.Tokens[WarRune] >= 1 + Action.NamedValues["Y"] && Action.NamedValues["X"] >= 1`
* **3. 目标选择规则**：
  * **SelectType**: `None`
  * **NamedValueConstraints**:
    * `{Key: "X", Required: true, MinExpression: "1", MaxExpression: "Self.HandCount"}`
    * `{Key: "Y", Required: true, MinExpression: "0", MaxExpression: "Self.Tokens[WarRune]-1"}`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Discards**:
    * `Count`: `Action.NamedValues["X"]`
    * `Filter`: `{SameAttribute: MatchElement}`
    * `Visibility`: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(翻转战纹总量=1+Y：先移除战纹)*:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `-(1 + Action.NamedValues["Y"])`
    * `TokenRef`: `WarRune`
    * `Ref`: `None`
  * **Effect[1]** *(翻转战纹总量=1+Y：再增加魔纹)*:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `1 + Action.NamedValues["Y"]`
    * `TokenRef`: `MagicRune`
    * `Ref`: `None`
  * **Effect[2]** *(本次攻击伤害额外+(X-1))*:
    * `EffectType`: `EffectAttackDamageModifier`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `Action.NamedValues["X"] - 1`
    * `Ref`: `None`
  * **Effect[3]** *(蓄势迸发形态下额外法术伤害+Y；结算到本次攻击原目标)*:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `Action.NamedValues["Y"]`
    * `Condition`: `Action.NamedValues["Y"] > 0`
    * `Ref`: `Combat.TargetID`

#### 【魔纹融合】 (Magic Rune Fusion)
* **技能描述**：（主动攻击未命中时②，翻转1个［魔纹］，弃X张异系牌［展示］（X>1））对本次攻击的角色造成（X-1）点法术伤害③，（若你处于［蓄势迸发形态］下，额外翻转Y个魔纹）本次法术伤害额外+Y。
* **1. 主干配置**：
  * **SkillID**: `skill_heroic_puppet_magic_rune_fusion`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnHitCheck`
* **1.1 响应分组配置**：
  * **ResponseGroup.GroupID**: `rg_heroic_puppet_on_miss`
  * **ResponseGroup.Mode**: `ResponseGroupChooseOne`
  * **ResponseGroup.OptionOrder**: `20`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatHitCheck]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && Combat.IsHit == false && Self.Tokens[MagicRune] >= 1 && (Self.Form == "heroic_puppet_surge_form" || Action.NamedValues["Y"] == 0) && Self.Tokens[MagicRune] >= 1 + Action.NamedValues["Y"] && Action.NamedValues["X"] > 1`
* **3. 目标选择规则**：
  * **SelectType**: `None`
  * **NamedValueConstraints**:
    * `{Key: "X", Required: true, MinExpression: "2", MaxExpression: "Self.HandCount"}`
    * `{Key: "Y", Required: true, MinExpression: "0", MaxExpression: "Self.Tokens[MagicRune]-1"}`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Discards**:
    * `Count`: `Action.NamedValues["X"]`
    * `Filter`: `{SameAttribute: MatchOppositeElement}`
    * `Visibility`: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(翻转魔纹总量=1+Y：先移除魔纹)*:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `-(1 + Action.NamedValues["Y"])`
    * `TokenRef`: `MagicRune`
    * `Ref`: `None`
  * **Effect[1]** *(翻转魔纹总量=1+Y：再增加战纹)*:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `1 + Action.NamedValues["Y"]`
    * `TokenRef`: `WarRune`
    * `Ref`: `None`
  * **Effect[2]** *(对本次攻击目标造成法术伤害：X-1+Y)*:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `Action.NamedValues["X"] - 1 + Action.NamedValues["Y"]`
    * `Ref`: `Combat.TargetID`

#### 【符文改造】 (Rune Reforge)
* **技能描述**：［宝石］［横置］，转为［蓄势迸发形态］形态，此形态下你的手牌上限+1，摸1张牌［强制］，并任意调整你的［战纹］和［魔纹］，在你的回合结束阶段，［转正］并脱离此形态。
* **1. 主干配置**：
  * **SkillID**: `skill_heroic_puppet_rune_reforge`
  * **Category**: `Ultimate`
  * **Type**: `Startup`
  * **Timing**: `TimingStartup`
* **1.2 规则修饰器模板配置**：
  * **ModifierID**: `rm_heroic_puppet_rune_reforge_max_hand_plus_1`
  * **Domain**: `RuleModifierDomainAttribute`
  * **Priority**: `180`
  * **ConditionExpression**: `Self.Form == "heroic_puppet_surge_form"`
  * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
  * **AttrPayload**: `{AttrType: PlayerAttributeMaxHand, Operation: AttributeModifyAdd, Value: 1}`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionStart]`
  * **CustomExpression**: `Self.Orientation == Normal`
* **3. 目标选择规则**：
  * **SelectType**: `None`
  * **NamedValueConstraints**:
    * `{Key: "WarToMagic", Required: true, MinExpression: "0", MaxExpression: "Self.Tokens[WarRune]"}`
    * `{Key: "MagicToWar", Required: true, MinExpression: "0", MaxExpression: "Self.Tokens[MagicRune]"}`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Gem, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(横置)*:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Tapped`
    * `Ref`: `None`
  * **Effect[1]** *(进入蓄势迸发形态)*:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `"heroic_puppet_surge_form"`
    * `Ref`: `None`
  * **Effect[2]** *(形态内手牌上限+1，到回合结束)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_heroic_puppet_rune_reforge_max_hand_plus_1"`
    * `RuleLifetimeRef`: `RuleLifeUntilTurnEnd`
    * `Ref`: `None`
  * **Effect[3]** *(摸1张牌［强制］)*:
    * `EffectType`: `EffectDrawCard`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `Ref`: `None`
  * **Effect[4]** *(任意调整：战纹->魔纹，先移除战纹)*:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `-Action.NamedValues["WarToMagic"]`
    * `TokenRef`: `WarRune`
    * `Ref`: `None`
  * **Effect[5]** *(任意调整：战纹->魔纹，再增加魔纹)*:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `Action.NamedValues["WarToMagic"]`
    * `TokenRef`: `MagicRune`
    * `Ref`: `None`
  * **Effect[6]** *(任意调整：魔纹->战纹，先移除魔纹)*:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `-Action.NamedValues["MagicToWar"]`
    * `TokenRef`: `MagicRune`
    * `Ref`: `None`
  * **Effect[7]** *(任意调整：魔纹->战纹，再增加战纹)*:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `Action.NamedValues["MagicToWar"]`
    * `TokenRef`: `WarRune`
    * `Ref`: `None`

#### 【符文改造·形态退场结算】 (Rune Reforge Cleanup)
* **说明**：将原技能“回合结束阶段转正并脱离形态”拆为同源被动子配置，不新增耦合字段。
* **1. 主干配置**：
  * **SkillID**: `skill_heroic_puppet_rune_reforge_cleanup`
  * **Category**: `Ultimate`
  * **Type**: `Passive`
  * **Timing**: `TimingOnTurnEnd`
* **2. 前置条件**：
  * **PhaseLimit**: `[TurnEnd]`
  * **CustomExpression**: `State.IsInSelfTurn(Self.UserID) == true && Self.Orientation == Tapped && Self.Form == "heroic_puppet_surge_form"`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(转正)*:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Normal`
    * `Ref`: `None`
  * **Effect[1]** *(脱离蓄势迸发形态)*:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `nil`
    * `Ref`: `None`

#### 【双重回响】 (Dual Echo)
* **技能描述**：［回合限定］［水晶］（对目标角色造成攻击或法术伤害时发动③）对另一个目标角色造成X点法术伤害③，X与本次伤害相同但最高为3。双重回响的伤害不会造成士气下降。
* **1. 主干配置**：
  * **SkillID**: `skill_heroic_puppet_dual_echo`
  * **Category**: `Ultimate`
  * **Type**: `Response`
  * **Timing**: `TimingOnDamageTaken`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDraw]`
  * **IsTurnLimited**: `true`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && (Combat.IsAttack == true || Combat.IsMagic == true) && Combat.FinalDamage > 0`
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **Filters**: `Target.UserID != Combat.TargetID` *(“另一个目标角色”：不可与本次受伤目标相同)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Crystal, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `NoMoraleDrop` *(本技能造成的法术伤害不造成士气下降)*
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelected`
    * `Value`: `Min(3, Combat.FinalDamage)`
    * `Ref`: `None`

---

### 21. 神官 (Priest)

#### 【神圣启示】 (Holy Revelation)
* **技能描述**：（［特殊行动］结束时发动）你+1［治疗］。
* **1. 主干配置**：
  * **SkillID**: `skill_priest_holy_revelation`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnActionEnd`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionEnd]`
  * **CustomExpression**: `Action.SourceID == Self.UserID && (Action.CurrentType == Buy || Action.CurrentType == Synthesize || Action.CurrentType == Extract || Action.CurrentType == Deadlock)`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectHeal`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `Ref`: `None`

#### 【神圣祈福】 (Holy Blessing)
* **技能描述**：（弃2张法术牌［展示］）你+2［治疗］。
* **1. 主干配置**：
  * **SkillID**: `skill_priest_holy_blessing`
  * **Category**: `Normal`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `None`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Discards**:
    * `Count`: `2`
    * `Filter`: `{ReqCardType: Magic}`
    * `Visibility`: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectHeal`
    * `Target`: `TargetSelf`
    * `Value`: `2`
    * `Ref`: `None`

#### 【水之神力】 (Water Divine Power)
* **技能描述**：（弃1张水系牌［展示］）将手中的1张牌交给目标队友［强制］，你和他各+1［治疗］。
* **1. 主干配置**：
  * **SkillID**: `skill_priest_water_divine_power`
  * **Category**: `Normal`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Self.HandCount >= 2` *(至少保留1张可转移手牌；另1张用于支付弃牌费用)*
* **3. 目标选择规则**：
  * **SelectType**: `TeamOther`
  * **MinCount**: `1`
  * **MaxCount**: `1`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Discards**:
    * `Count`: `1`
    * `Filter`: `{ReqElement: Water}`
    * `Visibility`: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(强制交牌：从自己手牌转移1张给目标队友；若未预选则运行时补选)*:
    * `EffectType`: `EffectTransferCard`
    * `Target`: `TargetSelected`
    * `Value`: `1`
    * `FromTargetRef`: `TargetSelf`
    * `Ref`: `None`
  * **Effect[1]** *(你+1治疗)*:
    * `EffectType`: `EffectHeal`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `Ref`: `None`
  * **Effect[2]** *(目标队友+1治疗)*:
    * `EffectType`: `EffectHeal`
    * `Target`: `TargetSelected`
    * `Value`: `1`
    * `Ref`: `None`

#### 【圣使守护】 (Holy Envoy Guard)
* **技能描述**：你的［治疗］上限+4，你每次用［治疗］抵御伤害时，最多只能使用1点。
* **1. 主干配置**：
  * **SkillID**: `skill_priest_holy_envoy_guard`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnTurnStart`
* **1.2 规则修饰器模板配置**：
  * **Template[A]** *(治疗上限+4)*:
    * **ModifierID**: `rm_priest_holy_envoy_guard_max_heal_plus_4`
    * **Domain**: `RuleModifierDomainAttribute`
    * **Priority**: `120`
    * **ConditionExpression**: `None`
    * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
    * **AttrPayload**: `{AttrType: PlayerAttributeMaxHeal, Operation: AttributeModifyAdd, Value: 4}`
  * **Template[B]** *(单次受伤窗口治疗抵伤上限=1)*:
    * **ModifierID**: `rm_priest_holy_envoy_guard_heal_resist_cap_1`
    * **Domain**: `RuleModifierDomainHealResistPolicy`
    * **Priority**: `220`
    * **ConditionExpression**: `None`
    * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
    * **HealResistPolicyPayload**: `{PerDamageWindowHealCap: 1}`
* **2. 前置条件**：
  * **PhaseLimit**: `[GameInit]`
  * **CustomExpression**: `Event.SourceType == SourceSystem && Event.CauseAction == "GameStart"`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(施加“治疗上限+4”常驻规则)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_priest_holy_envoy_guard_max_heal_plus_4"`
    * `RuleLifetimeRef`: `RuleLifePermanent`
    * `Ref`: `None`
  * **Effect[1]** *(施加“单次受伤窗口治疗抵伤上限=1”常驻规则)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_priest_holy_envoy_guard_heal_resist_cap_1"`
    * `RuleLifetimeRef`: `RuleLifePermanent`
    * `Ref`: `None`

#### 【神圣契约】 (Holy Covenant)
* **技能描述**：［水晶］将你的X［治疗］转移给目标队友，以此法所转移的［治疗］无视他的［治疗］上限，但他的［治疗］最高为4。
* **1. 主干配置**：
  * **SkillID**: `skill_priest_holy_covenant`
  * **Category**: `Ultimate`
  * **Type**: `Startup`
  * **Timing**: `TimingStartup`
* **1.2 规则修饰器模板配置**：
  * **ModifierID**: `rm_priest_holy_covenant_heal_policy_ignore_cap_abs_4`
  * **Domain**: `RuleModifierDomainHealPolicy`
  * **Priority**: `260`
  * **ConditionExpression**: `None`
  * **StackPolicy**: `RuleModifierStackReplaceByDomainPriority`
  * **HealPolicyPayload**: `{ApplyMode: HealApplyIgnoreMax, AbsoluteMax: 4}`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionStart]`
  * **CustomExpression**: `Self.Heal >= 1`
* **3. 目标选择规则**：
  * **SelectType**: `TeamOther`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **NamedValueConstraints**:
    * `{Key: "X", Required: true, MinExpression: "1", MaxExpression: "Self.Heal"}`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Crystal, Amount: 1}]`
  * **HealCost**: `Action.NamedValues["X"]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(对目标施加“本次治疗无视上限+绝对封顶4”策略)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelected`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_priest_holy_covenant_heal_policy_ignore_cap_abs_4"`
    * `RuleLifetimeRef`: `RuleLifeThisEffectChain`
    * `Ref`: `None`
  * **Effect[1]** *(将X点治疗转移给目标队友)*:
    * `EffectType`: `EffectHeal`
    * `Target`: `TargetSelected`
    * `Value`: `Action.NamedValues["X"]`
    * `Ref`: `None`

#### 【神圣领域】 (Holy Domain)
* **技能描述**：［水晶］你弃2张牌，再选择以下1项发动：●（移除你的1［治疗］）对目标角色造成2点法术伤害③。●你+2［治疗］，目标队友+1［治疗］。
* **1. 主干配置**：
  * **SkillID**: `skill_priest_holy_domain`
  * **Category**: `Ultimate`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `None`
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **SelectedValueRule**: `(Action.SelectedValue == 0 && Self.Heal >= 1) || Action.SelectedValue == 1` *(0=伤害分支；1=治疗分支)*
  * **Filters**: `(Action.SelectedValue == 0 && Target.UserID != Self.UserID) || (Action.SelectedValue == 1 && Target.Team == Self.Team && Target.UserID != Self.UserID)`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Crystal, Amount: 1}]`
  * **Discards**:
    * `Count`: `2`
    * `Filter`: `{SameAttribute: MatchNone}`
    * `Visibility`: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **BranchSelector**: `Action.SelectedValue`
  * **Branch[0]** *(移除自身1治疗，对目标造成2点法术伤害)*:
    * **Effect[B0]**:
      * `EffectType`: `EffectHeal`
      * `Target`: `TargetSelf`
      * `Value`: `-1`
      * `Ref`: `None`
    * **Effect[B1]**:
      * `EffectType`: `EffectDamage`
      * `Target`: `TargetSelected`
      * `Value`: `2`
      * `Ref`: `None`
  * **Branch[1]** *(你+2治疗，目标队友+1治疗)*:
    * **Effect[C0]**:
      * `EffectType`: `EffectHeal`
      * `Target`: `TargetSelf`
      * `Value`: `2`
      * `Ref`: `None`
    * **Effect[C1]**:
      * `EffectType`: `EffectHeal`
      * `Target`: `TargetSelected`
      * `Value`: `1`
      * `Ref`: `None`

---

### 22. 阴阳师 (Onmyoji)

#### 【式神降临】 (Shikigami Descend)
* **技能描述**：［持续］（弃2张命格相同的手牌［展示］）［横置］转为［式神形态］，你+1［鬼火］，额外+1［攻击行动］。
* **1. 主干配置**：
  * **SkillID**: `skill_onmyoji_shikigami_descend`
  * **Category**: `Normal`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Self.Orientation == Normal`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Discards**:
    * `Count`: `2`
    * `Filter`: `{SameAttribute: MatchDestiny}`
    * `Visibility`: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Tapped`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `"onmyoji_shikigami_form"`
    * `Ref`: `None`
  * **Effect[2]**:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `TokenRef`: `GhostFire`
    * `Ref`: `None`
  * **Effect[3]**:
    * `Implementation`: `model.AppendExtraAction`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `ActionRef`: `Attack`
    * `Ref`: `None`

#### 【阴阳转换】 (Yin-Yang Conversion)
* **技能描述**：（应战攻击时①，打出1张与攻击牌命格相同的攻击牌［展示］）视为你应战此次攻击，并将本次攻击系别转为与此牌相同。你+1［鬼火］。（若处于式神形态，［转正］脱离［式神形态］）本次攻击伤害为X，X为你的［鬼火］数。
* **1. 主干配置**：
  * **SkillID**: `skill_onmyoji_yinyang_conversion`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnHitCheck`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatHitCheck]`
  * **CustomExpression**: `Combat.IsAttack == true && Combat.TargetID == Self.UserID && Action.PlayedCard != nil && Action.PlayedCard.CardType == Attack && Action.PlayedCard.Destiny == Combat.AttackCard.Destiny`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired` *(该响应以本次提交的应战攻击牌为载体，合法性由 CustomExpression 校验)*
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(声明“本次应战由你执行”)*:
    * `EffectType`: `EffectSetCurrentCounterExecutor`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `Ref`: `None`
  * **Effect[1]** *(将当前战斗系别改为应战牌系别)*:
    * `EffectType`: `EffectSetCurrentCombatElement`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `0`
    * `ElementRef`: `Action.PlayedCard.Element`
    * `Ref`: `None`
  * **Effect[2]** *(你+1鬼火)*:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `TokenRef`: `GhostFire`
    * `Ref`: `None`
  * **Effect[3]** *(若处于式神形态则转正)*:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Normal`
    * `Condition`: `Self.Form == "onmyoji_shikigami_form"`
    * `Ref`: `None`
  * **Effect[4]** *(若处于式神形态则脱离形态)*:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `nil`
    * `Condition`: `Self.Form == "onmyoji_shikigami_form"`
    * `Ref`: `None`
  * **Effect[5]** *(将本次攻击待生效伤害改写为当前鬼火数)*:
    * `EffectType`: `EffectModifyPendingDamage`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `Self.Tokens[GhostFire] - Combat.FinalDamage`
    * `Ref`: `None`

#### 【式神转换】 (Shikigami Conversion)
* **技能描述**：（与阴阳转换同时发动）你摸1张牌［强制］，+1［鬼火］。
* **1. 主干配置**：
  * **SkillID**: `skill_onmyoji_shikigami_conversion`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnHitCheck`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatHitCheck]`
  * **CustomExpression**: `Combat.IsAttack == true && Combat.TargetID == Self.UserID && Action.PlayedCard != nil && Action.PlayedCard.CardType == Attack && Action.PlayedCard.Destiny == Combat.AttackCard.Destiny`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectDrawCard`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `TokenRef`: `GhostFire`
    * `Ref`: `None`

#### 【黑暗祭礼】 (Dark Ritual)
* **技能描述**：（你的回合结束时，若［鬼火］达到上限）移除所有［鬼火］，对目标角色造成2点法术伤害③。
* **1. 主干配置**：
  * **SkillID**: `skill_onmyoji_dark_ritual`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnTurnEnd`
* **2. 前置条件**：
  * **PhaseLimit**: `[TurnEnd]`
  * **CustomExpression**: `State.IsInSelfTurn(Self.UserID) == true && Self.Tokens[GhostFire] >= 3`
* **3. 目标选择规则**：
  * **SelectType**: `Enemy`
  * **MinCount**: `1`
  * **MaxCount**: `1`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(移除所有鬼火)*:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `-Self.Tokens[GhostFire]`
    * `TokenRef`: `GhostFire`
    * `Ref`: `None`
  * **Effect[1]** *(对目标造成2点法术伤害)*:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelected`
    * `Value`: `2`
    * `Ref`: `None`

#### 【式神咒束】 (Shikigami Curse)
* **技能描述**：（目标队友受到主动攻击时①，若此攻击可应战且你处于［式神形态］，打出1张合理的应战攻击牌［展示］，移除我方［战绩区］1［宝石］1［水晶］）将本次攻击目标变更为你，且视为你使用此牌执行应战攻击。
* **1. 主干配置**：
  * **SkillID**: `skill_onmyoji_shikigami_curse`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnAttackDeclared`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDeclare]`
  * **CustomExpression**: `Combat.IsAttack == true && Combat.IsActiveAttack == true && Combat.TargetID != Self.UserID && Self.Form == "onmyoji_shikigami_form" && Action.PlayedCard != nil && Action.PlayedCard.CardType == Attack && Action.PlayedCard.IsCounter == true`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(移除我方战绩区1宝石)*:
    * `EffectType`: `EffectAddTeamStone`
    * `Target`: `TargetSelfTeam`
    * `Value`: `-1`
    * `StoneRef`: `Gem`
    * `Ref`: `None`
  * **Effect[1]** *(移除我方战绩区1水晶)*:
    * `EffectType`: `EffectAddTeamStone`
    * `Target`: `TargetSelfTeam`
    * `Value`: `-1`
    * `StoneRef`: `Crystal`
    * `Ref`: `None`
  * **Effect[2]** *(将当前攻击目标改为你)*:
    * `EffectType`: `EffectRedirectCurrentCombatTarget`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `Ref`: `None`
  * **Effect[3]** *(声明“本次应战由你执行”)*:
    * `EffectType`: `EffectSetCurrentCounterExecutor`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `Ref`: `None`
  * **Effect[4]** *(同步当前战斗系别为你打出的应战牌系别)*:
    * `EffectType`: `EffectSetCurrentCombatElement`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `0`
    * `ElementRef`: `Action.PlayedCard.Element`
    * `Ref`: `None`

#### 【生命结界】 (Life Barrier)
* **技能描述**：［水晶］你+1［鬼火］，选择以下一项发动：●目标队友+1［宝石］并+1［治疗］；然后你对自己造成X点法术伤害③，X为你的［鬼火］数。若X为3，本次法术伤害不会造成我方士气下降。●（仅［式神型态］下，弃2张相同命格的手牌［展示］）［转正］脱离［式神型态］，目标队友弃1张手牌。
* **1. 主干配置**：
  * **SkillID**: `skill_onmyoji_life_barrier`
  * **Category**: `Ultimate`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `None`
* **3. 目标选择规则**：
  * **SelectType**: `TeamOther`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **SelectedValueRule**: `Action.SelectedValue == 0 || (Action.SelectedValue == 1 && Self.Form == "onmyoji_shikigami_form" && Self.HandCount >= 2)` *(0=增益+自伤分支；1=退场弃牌分支)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Crystal, Amount: 1}]`
  * **Discards**:
    * `Count`: `Action.SelectedValue == 1 ? 2 : 0`
    * `Filter`: `{SameAttribute: MatchDestiny}`
    * `Visibility`: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(你+1鬼火)*:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `TokenRef`: `GhostFire`
    * `Ref`: `None`
  * **BranchSelector**: `Action.SelectedValue`
  * **Branch[0]** *(目标队友+1宝石并+1治疗；然后你承受X点法术伤害)*:
    * **Effect[B0]**:
      * `EffectType`: `EffectAddEnergyStone`
      * `Target`: `TargetSelected`
      * `Value`: `1`
      * `StoneRef`: `Gem`
      * `Ref`: `None`
    * **Effect[B1]**:
      * `EffectType`: `EffectHeal`
      * `Target`: `TargetSelected`
      * `Value`: `1`
      * `Ref`: `None`
    * **Effect[B2]**:
      * `EffectType`: `EffectDamage`
      * `Target`: `TargetSelf`
      * `Value`: `Self.Tokens[GhostFire]`
      * `Ref`: `None`
  * **Branch[1]** *(式神形态退场，目标队友弃1张手牌)*:
    * **Effect[C0]**:
      * `EffectType`: `EffectSetOrientation`
      * `Target`: `TargetSelf`
      * `Value`: `0`
      * `OrientationRef`: `Normal`
      * `Ref`: `None`
    * **Effect[C1]**:
      * `EffectType`: `EffectSetForm`
      * `Target`: `TargetSelf`
      * `Value`: `0`
      * `FormRef`: `nil`
      * `Ref`: `None`
    * **Effect[C2]**:
      * `EffectType`: `EffectDiscard`
      * `Target`: `TargetSelected`
      * `Value`: `1`
      * `Ref`: `None`

#### 【生命结界·三鬼火免士气】 (Life Barrier Morale Shield)
* **说明**：用于承接“生命结界分支0造成的自伤在 X=3 时不造成我方士气下降”。
* **1. 主干配置**：
  * **SkillID**: `skill_onmyoji_life_barrier_morale_shield`
  * **Category**: `Ultimate`
  * **Type**: `Passive`
  * **Timing**: `TimingOnDamageTaken`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDraw]`
  * **CustomExpression**: `Combat.TargetID == Self.UserID && Event.PendingMoraleLoss > 0 && Event.MoraleDropSourceSkillID == "skill_onmyoji_life_barrier" && Self.Tokens[GhostFire] == 3`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectReducePendingMoraleLoss`
    * `Target`: `TargetCurrentEvent`
    * `Value`: `Event.PendingMoraleLoss`
    * `Ref`: `None`

---

### 23. 苍炎魔女 (Azure Flame Witch)

#### 【苍炎法典】 (Azure Flame Codex)
* **技能描述**：（弃1张火系牌［展示］）对目标角色和自己各造成2点法术伤害③。
* **1. 主干配置**：
  * **SkillID**: `skill_azure_flame_witch_azure_flame_codex`
  * **Category**: `Normal`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `None`
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **Filters**: `Target.UserID != Self.UserID`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Discards**:
    * `Count`: `1`
    * `Filter`: `{ReqElement: Fire}`
    * `Visibility`: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(目标角色承受2点法术伤害)*:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelected`
    * `Value`: `2`
    * `Ref`: `None`
  * **Effect[1]** *(你承受2点法术伤害)*:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelf`
    * `Value`: `2`
    * `Ref`: `None`

#### 【天火断空】 (Heavenfire Severance)
* **技能描述**：（弃2张火系牌［展示］，移除1点［重生］）对目标角色和自己各造成3点法术伤害③，（若我方士气落后于该目标）本次法术伤害额外+1［强制］。
* **1. 主干配置**：
  * **SkillID**: `skill_azure_flame_witch_heavenfire_severance`
  * **Category**: `Normal`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Self.Form == "azure_flame_witch_blaze_form" || Self.Tokens[Rebirth] >= 1`
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **Filters**: `Target.UserID != Self.UserID`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Discards**:
    * `Count`: `2`
    * `Filter`: `{ReqElement: Fire}`
    * `Visibility`: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(非烈焰形态时移除1重生；烈焰形态下跳过，等价于“无需消耗重生”)*:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `-1`
    * `TokenRef`: `Rebirth`
    * `Condition`: `Self.Form != "azure_flame_witch_blaze_form"`
    * `Ref`: `None`
  * **Effect[1]** *(目标角色承受基础3点法术伤害)*:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelected`
    * `Value`: `3`
    * `Ref`: `None`
  * **Effect[2]** *(你承受基础3点法术伤害)*:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelf`
    * `Value`: `3`
    * `Ref`: `None`
  * **Effect[3]** *(若我方士气落后于该目标，则对目标伤害额外+1)*:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelected`
    * `Value`: `1`
    * `Condition`: `(Self.Team == Red && State.RedMorale < State.BlueMorale) || (Self.Team == Blue && State.BlueMorale < State.RedMorale)`
    * `Ref`: `None`
  * **Effect[4]** *(若我方士气落后于该目标，则对自己伤害额外+1)*:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `Condition`: `(Self.Team == Red && State.RedMorale < State.BlueMorale) || (Self.Team == Blue && State.BlueMorale < State.RedMorale)`
    * `Ref`: `None`

#### 【魔女之怒】 (Witch's Wrath)
* **技能描述**：（手牌<4张时）［横置］摸0-2张牌，数值由你决定，持续到你的下个行动阶段开始前，你都处于［烈焰形态］，在此形态下你的所有除水系和暗系外的攻击牌均视为火系［强制］，你释放天火断空时无需消耗［重生］，你的手牌上限+（Ｘ-2），X为你的［重生］数量；脱离［烈焰形态］时［转正］。
* **1. 主干配置**：
  * **SkillID**: `skill_azure_flame_witch_witch_wrath`
  * **Category**: `Normal`
  * **Type**: `Startup`
  * **Timing**: `TimingStartup`
* **1.2 规则修饰器模板配置**：
  * **ModifierID**: `rm_azure_flame_witch_blaze_form_max_hand_dynamic`
  * **Domain**: `RuleModifierDomainAttribute`
  * **Priority**: `190`
  * **ConditionExpression**: `Self.Form == "azure_flame_witch_blaze_form"`
  * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
  * **AttrPayload**: `{AttrType: PlayerAttributeMaxHand, Operation: AttributeModifyAdd, ValueSourceMode: RuleAttrValueFromTokenLinear, TokenLink: {OwnerScope: RuleAttrTokenOwnerTarget, TokenType: Rebirth, Coefficient: 1, Offset: -2}}`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionStart]`
  * **CustomExpression**: `Self.HandCount < 4 && Self.Orientation == Normal`
* **3. 目标选择规则**：
  * **SelectType**: `None`
  * **NamedValueConstraints**:
    * `{Key: "DrawN", Required: true, MinExpression: "0", MaxExpression: "2"}`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(横置进入形态)*:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Tapped`
    * `Ref`: `None`
  * **Effect[1]** *(进入烈焰形态)*:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `"azure_flame_witch_blaze_form"`
    * `Ref`: `None`
  * **Effect[2]** *(摸0-2张牌，由玩家选择)*:
    * `EffectType`: `EffectDrawCard`
    * `Target`: `TargetSelf`
    * `Value`: `Action.NamedValues["DrawN"]`
    * `Ref`: `None`
  * **Effect[3]** *(施加“手牌上限 + (重生-2)”动态规则，持续到下个行动阶段开始前)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_azure_flame_witch_blaze_form_max_hand_dynamic"`
    * `RuleLifetimeRef`: `RuleLifeUntilSourceNextTurnStart`
    * `Ref`: `None`

#### 【魔女之怒·烈焰改写】 (Witch's Wrath Blaze Rewrite)
* **说明**：烈焰形态下，你打出的非水/暗攻击牌在执行前强制改写为火系攻击。
* **1. 主干配置**：
  * **SkillID**: `skill_azure_flame_witch_witch_wrath_blaze_rewrite`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingBeforeActionExecute`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Self.Form == "azure_flame_witch_blaze_form" && Action.SourceID == Self.UserID && Action.CurrentType == Attack && Action.PlayedCard != nil && Action.PlayedCard.Element != Water && Action.PlayedCard.Element != Dark`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **3.1 行动改写配置**：
  * **ActionTransform.Hook**: `TimingBeforeActionExecute`
  * **ActionTransform.Optional**: `false`
  * **ActionTransform.Priority**: `220`
  * **ActionTransform.Match.RequireActionType**: `Attack`
  * **ActionTransform.Match.ExcludeTemplateIDs**: `[]`
  * **ActionTransform.Rewrite.FlowRef**: `ActionFlowNormalCombat`
  * **ActionTransform.Rewrite.ActionTypeRef**: `Attack`
  * **ActionTransform.Rewrite.ElementPickMode**: `RewriteElementFixed`
  * **ActionTransform.Rewrite.FixedElementRef**: `Fire`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * `None` *(本技能通过 ActionTransform 生效)*

#### 【魔女之怒·形态退场结算】 (Witch's Wrath Cleanup)
* **说明**：在你的下个行动阶段开始前，退出烈焰形态并转正。
* **1. 主干配置**：
  * **SkillID**: `skill_azure_flame_witch_witch_wrath_cleanup`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnBeforeAction`
* **2. 前置条件**：
  * **PhaseLimit**: `[BeforeAction]`
  * **CustomExpression**: `State.IsInSelfTurn(Self.UserID) == true && Self.Orientation == Tapped && Self.Form == "azure_flame_witch_blaze_form"`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Normal`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `nil`
    * `Ref`: `None`

#### 【替身玩偶】 (Substitute Doll)
* **技能描述**：（任何人对你造成攻击伤害时③，弃1张法术牌［展示］）目标队友摸1张牌［强制］。
* **1. 主干配置**：
  * **SkillID**: `skill_azure_flame_witch_substitute_doll`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnDamageTaken`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDraw]`
  * **CustomExpression**: `Combat.TargetID == Self.UserID && Combat.IsAttack == true && Combat.FinalDamage > 0`
* **3. 目标选择规则**：
  * **SelectType**: `TeamOther`
  * **MinCount**: `1`
  * **MaxCount**: `1`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Discards**:
    * `Count`: `1`
    * `Filter`: `{ReqCardType: Magic}`
    * `Visibility`: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectDrawCard`
    * `Target`: `TargetSelected`
    * `Value`: `1`
    * `Ref`: `None`

#### 【永生银时计】 (Eternal Silver Timepiece)
* **技能描述**：（当你因为承受法术伤害而造成士气下降时）你+1［重生］。
* **1. 主干配置**：
  * **SkillID**: `skill_azure_flame_witch_eternal_silver_timepiece`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnDamageTaken`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDraw]`
  * **CustomExpression**: `Combat.TargetID == Self.UserID && Combat.IsMagic == true && Event.PendingMoraleLoss > 0`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `TokenRef`: `Rebirth`
    * `Ref`: `None`

#### 【痛苦链接】 (Pain Link)
* **技能描述**：［水晶］对目标对手和自己各造成1点法术伤害③, 然后你弃到3张牌。
* **1. 主干配置**：
  * **SkillID**: `skill_azure_flame_witch_pain_link`
  * **Category**: `Ultimate`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `None`
* **3. 目标选择规则**：
  * **SelectType**: `Enemy`
  * **MinCount**: `1`
  * **MaxCount**: `1`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Crystal, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelected`
    * `Value`: `1`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `Ref`: `None`
  * **Effect[2]** *(弃到3张牌)*:
    * `EffectType`: `EffectDiscard`
    * `Target`: `TargetSelf`
    * `Value`: `Self.HandCount > 3 ? Self.HandCount - 3 : 0`
    * `Ref`: `None`

#### 【魔能反转】 (Mana Inversion)
* **技能描述**：［水晶］（任何人对你造成法术伤害时③，弃X张法术牌［展示］（X>1））对目标对手造成（Ｘ-1） 点法术伤害 ③。
* **1. 主干配置**：
  * **SkillID**: `skill_azure_flame_witch_mana_inversion`
  * **Category**: `Ultimate`
  * **Type**: `Response`
  * **Timing**: `TimingOnDamageTaken`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDraw]`
  * **CustomExpression**: `Combat.TargetID == Self.UserID && Combat.IsMagic == true && Combat.FinalDamage > 0`
* **3. 目标选择规则**：
  * **SelectType**: `Enemy`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **NamedValueConstraints**:
    * `{Key: "X", Required: true, MinExpression: "2", MaxExpression: "Self.HandCount"}`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Crystal, Amount: 1}]`
  * **Discards**:
    * `Count`: `Action.NamedValues["X"]`
    * `Filter`: `{ReqCardType: Magic}`
    * `Visibility`: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelected`
    * `Value`: `Action.NamedValues["X"] - 1`
    * `Ref`: `None`

---

### 24. 贤者 (Sage)

#### 【智慧法典】 (Wisdom Codex)
* **技能描述**：你的能量上限+1；（你每次承受法术伤害时⑥，若该伤害>3）你+2［宝石］并弃1张牌。
* **1. 主干配置**：
  * **SkillID**: `skill_sage_wisdom_codex`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnTurnStart` / `TimingOnDamageTaken`
* **1.2 规则修饰器模板配置**：
  * **ModifierID**: `rm_sage_wisdom_codex_max_energy_plus_1`
  * **Domain**: `RuleModifierDomainAttribute`
  * **Priority**: `100`
  * **ConditionExpression**: `None`
  * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
  * **AttrPayload**: `{AttrType: PlayerAttributeMaxEnergy, Operation: AttributeModifyAdd, ValueSourceMode: RuleAttrValueFromFixed, Value: 1}`
* **2. 前置条件**：
  * **PhaseLimit**: `None`
  * **CustomExpression**: `(Event.TriggerHook == TimingOnTurnStart && Event.SourceType == SourceSystem && Event.CauseAction == "GameStart") || (Event.TriggerHook == TimingOnDamageTaken && Combat.TargetID == Self.UserID && Combat.IsMagic == true && Combat.FinalDamage > 3)`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(游戏初始化时施加“能量上限+1”常驻规则)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_sage_wisdom_codex_max_energy_plus_1"`
    * `RuleLifetimeRef`: `RuleLifePermanent`
    * `Condition`: `Event.TriggerHook == TimingOnTurnStart && Event.SourceType == SourceSystem && Event.CauseAction == "GameStart"`
    * `Ref`: `None`
  * **Effect[1]** *(承受法术伤害>3时，+2宝石)*:
    * `EffectType`: `EffectAddEnergyStone`
    * `Target`: `TargetSelf`
    * `Value`: `2`
    * `StoneRef`: `Gem`
    * `Condition`: `Event.TriggerHook == TimingOnDamageTaken && Combat.TargetID == Self.UserID && Combat.IsMagic == true && Combat.FinalDamage > 3`
    * `Ref`: `None`
  * **Effect[2]** *(承受法术伤害>3时，弃1张牌)*:
    * `EffectType`: `EffectDiscard`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `Condition`: `Event.TriggerHook == TimingOnDamageTaken && Combat.TargetID == Self.UserID && Combat.IsMagic == true && Combat.FinalDamage > 3`
    * `Ref`: `None`

#### 【法术反弹】 (Spell Rebound)
* **技能描述**：（你每次承受法术伤害时⑥，若该伤害仅为1点，则可以弃X张同系牌［展示］（X>1））对目标角色造成（X-1）点法术伤害③，并对自己造成X点法术伤害③。
* **1. 主干配置**：
  * **SkillID**: `skill_sage_spell_rebound`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnDamageTaken`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDraw]`
  * **CustomExpression**: `Combat.TargetID == Self.UserID && Combat.IsMagic == true && Combat.FinalDamage == 1`
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **Filters**: `Target.UserID != Self.UserID`
  * **NamedValueConstraints**:
    * `{Key: "X", Required: true, MinExpression: "2", MaxExpression: "Self.HandCount"}`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Discards**:
    * `Count`: `Action.NamedValues["X"]`
    * `Filter`: `{SameAttribute: MatchElement}`
    * `Visibility`: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelected`
    * `Value`: `Action.NamedValues["X"] - 1`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelf`
    * `Value`: `Action.NamedValues["X"]`
    * `Ref`: `None`

#### 【魔道法典】 (Arcane Codex)
* **技能描述**：［宝石］（弃X张异系牌［展示］（X>1））对目标角色与自己各造成（X-1）点法术伤害③。
* **1. 主干配置**：
  * **SkillID**: `skill_sage_arcane_codex`
  * **Category**: `Ultimate`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `None`
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **Filters**: `Target.UserID != Self.UserID`
  * **NamedValueConstraints**:
    * `{Key: "X", Required: true, MinExpression: "2", MaxExpression: "Self.HandCount"}`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Gem, Amount: 1}]`
  * **Discards**:
    * `Count`: `Action.NamedValues["X"]`
    * `Filter`: `{SameAttribute: MatchOppositeElement}`
    * `Visibility`: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelected`
    * `Value`: `Action.NamedValues["X"] - 1`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelf`
    * `Value`: `Action.NamedValues["X"] - 1`
    * `Ref`: `None`

#### 【圣洁法典】 (Sacred Codex)
* **技能描述**：［宝石］（弃X张异系牌［展示］（X>2））最多（X-2）名角色各+2［治疗］，并对自己造成（X-1）点法术伤害③。
* **1. 主干配置**：
  * **SkillID**: `skill_sage_sacred_codex`
  * **Category**: `Ultimate`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `None`
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `1`
  * **MaxCount**: `6`
  * **Filters**: `Action.SelectedTargetCount <= Action.NamedValues["X"] - 2`
  * **NamedValueConstraints**:
    * `{Key: "X", Required: true, MinExpression: "3", MaxExpression: "Self.HandCount"}`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Gem, Amount: 1}]`
  * **Discards**:
    * `Count`: `Action.NamedValues["X"]`
    * `Filter`: `{SameAttribute: MatchOppositeElement}`
    * `Visibility`: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectHeal`
    * `Target`: `TargetSelected`
    * `Value`: `2`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelf`
    * `Value`: `Action.NamedValues["X"] - 1`
    * `Ref`: `None`

---

### 25. 魔弓 (Magic Bow)

#### 【魔贯冲击】 (Arc Piercing Impact)
* **技能描述**：（主动攻击前发动①，移除1个火系［充能］）本次攻击伤害额外+1，不能攻击手牌达到上限的对手；（若命中②，额外移除1个火系［充能］）本次攻击伤害额外+1；（若未命中②）对对手造成3点法术伤害③。本回合你不能发动多重射击。
* **1. 主干配置**：
  * **SkillID**: `skill_magic_bow_piercing_impact`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnAttackDeclared`
* **1.2 规则修饰器模板配置**：
  * **Template[A]** *(挂载“本次魔贯冲击已发动”战斗窗口标记，供命中/未命中延后分支读取)*:
    * **ModifierID**: `rm_magic_bow_piercing_impact_armed`
    * **Domain**: `RuleModifierDomainCardSource`
    * **Priority**: `80`
    * **ConditionExpression**: `None`
    * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
    * **CardSourcePayload**: `{ProjectionMode: CardSourceProjectionAsHand, FieldMarks: []}`
  * **Template[B]** *(本回合禁用【多重射击】)*:
    * **ModifierID**: `rm_magic_bow_piercing_impact_disable_multi_shot_turn`
    * **Domain**: `RuleModifierDomainSkillGate`
    * **Priority**: `220`
    * **ConditionExpression**: `None`
    * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
    * **SkillGatePayload**: `{Mode: SkillGateDisallowList, SkillIDs: ["skill_magic_bow_multi_shot"]}`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDeclare]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && State.CountFieldMarkByElement(Self.UserID, Cover, Fire) >= 1 && Combat.TargetHandCount < Combat.TargetHandLimit`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(移除1个火系充能；充能按 Cover 盖牌结算)*:
    * `EffectType`: `EffectRemoveFieldMark`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `FieldMarkRef`: `Cover`
    * `ElementRef`: `Fire`
    * `Ref`: `None`
  * **Effect[1]** *(本次攻击伤害+1)*:
    * `EffectType`: `EffectAttackDamageModifier`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `1`
    * `Ref`: `None`
  * **Effect[2]** *(挂载命中/未命中延后结算标记，持续到本次战斗结束)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_magic_bow_piercing_impact_armed"`
    * `RuleLifetimeRef`: `RuleLifeUntilCombatEnd`
    * `Ref`: `None`
  * **Effect[3]** *(本回合禁用【多重射击】)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_magic_bow_piercing_impact_disable_multi_shot_turn"`
    * `RuleLifetimeRef`: `RuleLifeUntilTurnEnd`
    * `Ref`: `None`

#### 【魔贯冲击·命中追加】 (Arc Piercing Impact Hit Bonus)
* **说明**：当【魔贯冲击】对应的本次攻击命中时，自动结算“额外移除1个火系充能并使本次攻击伤害额外+1”（若无火系充能则本分支不触发）。
* **1. 主干配置**：
  * **SkillID**: `skill_magic_bow_piercing_impact_hit_bonus`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnHitCheck`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatHitCheck]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsHit == true && State.HasRuleModifier(Self.UserID, "rm_magic_bow_piercing_impact_armed") == true && State.CountFieldMarkByElement(Self.UserID, Cover, Fire) >= 1`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectRemoveFieldMark`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `FieldMarkRef`: `Cover`
    * `ElementRef`: `Fire`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectAttackDamageModifier`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `1`
    * `Ref`: `None`

#### 【魔贯冲击·未命中追伤】 (Arc Piercing Impact Miss Follow-Up)
* **说明**：当【魔贯冲击】对应的本次攻击未命中时，对原目标造成3点法术伤害③。
* **1. 主干配置**：
  * **SkillID**: `skill_magic_bow_piercing_impact_miss_followup`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnHitCheck`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatHitCheck]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsHit == false && State.HasRuleModifier(Self.UserID, "rm_magic_bow_piercing_impact_armed") == true`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelected`
    * `Value`: `3`
    * `Ref`: `Combat.TargetID`

#### 【雷光散射】 (Thunder Scatter)
* **技能描述**：（移除1个雷系［充能］）对所有对手各造成1点法术伤害③，（若你额外移除X个雷系［充能］）指定一名对手，本次对其造成的伤害额外+X③。
* **1. 主干配置**：
  * **SkillID**: `skill_magic_bow_thunder_scatter`
  * **Category**: `Normal`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `State.CountFieldMarkByElement(Self.UserID, Cover, Thunder) >= 1`
* **3. 目标选择规则**：
  * **SelectType**: `Enemy`
  * **MinCount**: `0` *(当 X=0 时无需额外指定目标)*
  * **MaxCount**: `1`
  * **Filters**: `Action.NamedValues["X"] == 0 || Action.SelectedTargetCount == 1`
  * **NamedValueConstraints**:
    * `{Key: "X", Required: true, MinExpression: "0", MaxExpression: "State.CountFieldMarkByElement(Self.UserID, Cover, Thunder) - 1"}`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(移除 1+X 个雷系充能)*:
    * `EffectType`: `EffectRemoveFieldMark`
    * `Target`: `TargetSelf`
    * `Value`: `1 + Action.NamedValues["X"]`
    * `FieldMarkRef`: `Cover`
    * `ElementRef`: `Thunder`
    * `Ref`: `None`
  * **Effect[1]** *(所有对手各承受1点法术伤害)*:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetAllEnemies`
    * `Value`: `1`
    * `Ref`: `None`
  * **Effect[2]** *(若 X>0，对额外指定目标追加 X 点法术伤害)*:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelected`
    * `Value`: `Action.NamedValues["X"]`
    * `Condition`: `Action.NamedValues["X"] > 0`
    * `Ref`: `None`

#### 【多重射击】 (Multi Shot)
* **技能描述**：（［攻击行动］结束时发动，移除1个风系［充能］）视为1次暗系的主动攻击，但不能攻击上次的目标且本次攻击伤害-1；本回合你不能使用魔贯冲击。
* **1. 主干配置**：
  * **SkillID**: `skill_magic_bow_multi_shot`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnActionEnd`
* **1.2 规则修饰器模板配置**：
  * **Template[A]** *(本回合禁用【魔贯冲击】)*:
    * **ModifierID**: `rm_magic_bow_multi_shot_disable_piercing_impact_turn`
    * **Domain**: `RuleModifierDomainSkillGate`
    * **Priority**: `220`
    * **ConditionExpression**: `None`
    * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
    * **SkillGatePayload**: `{Mode: SkillGateDisallowList, SkillIDs: ["skill_magic_bow_piercing_impact"]}`
  * **Template[B]** *(挂载“本次多重射击改写攻击”标记，供伤害-1子技能读取)*:
    * **ModifierID**: `rm_magic_bow_multi_shot_armed`
    * **Domain**: `RuleModifierDomainCardSource`
    * **Priority**: `80`
    * **ConditionExpression**: `None`
    * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
    * **CardSourcePayload**: `{ProjectionMode: CardSourceProjectionAsHand, FieldMarks: []}`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionEnd]`
  * **CustomExpression**: `State.IsInSelfTurn(Self.UserID) == true && Event.OperatorID == Self.UserID && Action.CurrentType == Attack && State.CountFieldMarkByElement(Self.UserID, Cover, Wind) >= 1`
* **3. 目标选择规则**：
  * **SelectType**: `Enemy`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **Filters**: `Target.UserID != State.GetLastActiveAttackTarget(Self.UserID)`
* **3.1 行动改写配置**：
  * **ActionTransform.Hook**: `TimingBeforeActionExecute`
  * **ActionTransform.Optional**: `false`
  * **ActionTransform.Priority**: `130`
  * **ActionTransform.CancelCurrentAction**: `true`
  * **ActionTransform.Match.RequireActionType**: `None`
  * **ActionTransform.Match.RequirePlayedCardTypes**: `[]`
  * **ActionTransform.Match.RequirePlayedCardElements**: `[]`
  * **ActionTransform.Match.ExcludeTemplateIDs**: `[]`
  * **ActionTransform.Rewrite.FlowRef**: `ActionFlowNormalCombat`
  * **ActionTransform.Rewrite.ActionTypeRef**: `Attack`
  * **ActionTransform.Rewrite.ExecuteImmediately**: `true`
  * **ActionTransform.Rewrite.TreatAsActiveAttack**: `true`
  * **ActionTransform.Rewrite.ElementPickMode**: `RewriteElementFixed`
  * **ActionTransform.Rewrite.FixedElementRef**: `Dark`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(移除1个风系充能)*:
    * `EffectType`: `EffectRemoveFieldMark`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `FieldMarkRef`: `Cover`
    * `ElementRef`: `Wind`
    * `Ref`: `None`
  * **Effect[1]** *(本回合禁用【魔贯冲击】)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_magic_bow_multi_shot_disable_piercing_impact_turn"`
    * `RuleLifetimeRef`: `RuleLifeUntilTurnEnd`
    * `Ref`: `None`
  * **Effect[2]** *(挂载“多重射击改写攻击”标记，持续到本次战斗结束)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_magic_bow_multi_shot_armed"`
    * `RuleLifetimeRef`: `RuleLifeUntilCombatEnd`
    * `Ref`: `None`

#### 【多重射击·伤害修正】 (Multi Shot Damage Modifier)
* **说明**：对【多重射击】改写出的本次暗系主动攻击施加伤害-1。
* **1. 主干配置**：
  * **SkillID**: `skill_magic_bow_multi_shot_damage_modifier`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnAttackDeclared`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDeclare]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && State.HasRuleModifier(Self.UserID, "rm_magic_bow_multi_shot_armed") == true`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectAttackDamageModifier`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `-1`
    * `Ref`: `None`

#### 【充能】 (Charge)
* **技能描述**：［水晶］你弃到4张牌，摸X张牌［强制］，将自己最多X张手牌面朝下放置于你的角色旁，作为［充能］（X<5）；本回合不能发动魔贯冲击和雷光散射。
* **1. 主干配置**：
  * **SkillID**: `skill_magic_bow_charge`
  * **Category**: `Ultimate`
  * **Type**: `Startup`
  * **Timing**: `TimingStartup`
* **1.2 规则修饰器模板配置**：
  * **Template[A]** *(本回合禁用【魔贯冲击】与【雷光散射】)*:
    * **ModifierID**: `rm_magic_bow_charge_disable_piercing_and_thunder_turn`
    * **Domain**: `RuleModifierDomainSkillGate`
    * **Priority**: `220`
    * **ConditionExpression**: `None`
    * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
    * **SkillGatePayload**: `{Mode: SkillGateDisallowList, SkillIDs: ["skill_magic_bow_piercing_impact", "skill_magic_bow_thunder_scatter"]}`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionStart]`
  * **CustomExpression**: `None`
* **3. 目标选择规则**：
  * **SelectType**: `None`
  * **NamedValueConstraints**:
    * `{Key: "X", Required: true, MinExpression: "0", MaxExpression: "4"}`
    * `{Key: "Y", Required: true, MinExpression: "0", MaxExpression: "Action.NamedValues[\"X\"]"}`
  * **SubmitAction.UsedCardUUIDs**: `当 Y>0 时，提交 Y 张来自自身手牌的 CardUUID，作为要放置为 Cover 的牌。`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Crystal, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(弃到4张牌)*:
    * `EffectType`: `EffectAdjustHand`
    * `Target`: `TargetSelf`
    * `Value`: `4`
    * `Ref`: `None`
  * **Effect[1]** *(摸X张牌)*:
    * `EffectType`: `EffectDrawCard`
    * `Target`: `TargetSelf`
    * `Value`: `Action.NamedValues["X"]`
    * `Ref`: `None`
  * **Effect[2]** *(将最多Y张手牌面朝下放置为 Cover 盖牌/充能)*:
    * `EffectType`: `EffectPlaceHandCardAsFieldMark`
    * `Target`: `TargetSelf`
    * `Value`: `Action.NamedValues["Y"]`
    * `FieldMarkRef`: `Cover`
    * `VisibilityRef`: `VisibilityHidden`
    * `Ref`: `None`
  * **Effect[3]** *(本回合禁用【魔贯冲击】与【雷光散射】)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_magic_bow_charge_disable_piercing_and_thunder_turn"`
    * `RuleLifetimeRef`: `RuleLifeUntilTurnEnd`
    * `Ref`: `None`

#### 【魔眼】 (Magic Eye)
* **技能描述**：［宝石］目标角色弃1张牌或你摸3张牌［强制］，将自己1张手牌作为［充能］，你+1［水晶］。
* **1. 主干配置**：
  * **SkillID**: `skill_magic_bow_magic_eye`
  * **Category**: `Ultimate`
  * **Type**: `Startup`
  * **Timing**: `TimingStartup`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionStart]`
  * **CustomExpression**: `Self.HandCount >= 1`
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **Filters**: `Target.UserID != Self.UserID`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Gem, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(目标“弃1牌”分支中断；若拒绝/无法弃置，则改为你摸3张牌)*:
    * `EffectType`: `EffectPerTargetBranch`
    * `Target`: `TargetSelected`
    * `Value`: `0`
    * `BranchRef`:
      * `TargetSource`: `PerTargetSelectedTargets`
      * `InterruptType`: `WaitDiscard`
      * `TimeoutAsDeclined`: `true`
      * `DiscardRequirement`: `{Count: 1, Filter: {}}`
      * `DiscardVisibility`: `VisibilityHidden`
      * `OnSuccess`: `[]`
      * `OnDeclined`:
        * `{EffectType: EffectDrawCard, Target: TargetSelf, Value: 3}`
    * `Ref`: `None`
  * **Effect[1]** *(将自己1张手牌面朝下放置为 Cover 盖牌/充能)*:
    * `EffectType`: `EffectPlaceHandCardAsFieldMark`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `FieldMarkRef`: `Cover`
    * `VisibilityRef`: `VisibilityHidden`
    * `Ref`: `None`
  * **Effect[2]** *(你+1水晶)*:
    * `EffectType`: `EffectAddEnergyStone`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `StoneRef`: `Crystal`
    * `Ref`: `None`

---

### 26. 魔枪 (Magic Spear)

#### 【暗之解放】 (Dark Liberation)
* **技能描述**：［横置］转为［幻影形态］，你的手牌上限恒定为5；本回合你的下一次主动攻击伤害额外+1，但不能发动漆黑之枪和充盈。
* **1. 主干配置**：
  * **SkillID**: `skill_magic_lance_dark_liberation`
  * **Category**: `Normal`
  * **Type**: `Startup`
  * **Timing**: `TimingStartup`
* **1.2 规则修饰器模板配置**：
  * **Template[A]** *(幻影形态下手牌上限恒定=5)*:
    * **ModifierID**: `rm_magic_lance_phantom_form_hand_limit_fixed_5`
    * **Domain**: `RuleModifierDomainAttribute`
    * **Priority**: `130`
    * **ConditionExpression**: `Self.Form == "magic_lance_phantom_form"`
    * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
    * **AttrPayload**: `{AttrType: PlayerAttributeMaxHand, Operation: AttributeModifySet, ValueSourceMode: RuleAttrValueFromFixed, Value: 5}`
  * **Template[B]** *(本回合“下次主动攻击+1”武装标记)*:
    * **ModifierID**: `rm_magic_lance_next_active_attack_damage_plus_1_armed`
    * **Domain**: `RuleModifierDomainCardSource`
    * **Priority**: `80`
    * **ConditionExpression**: `None`
    * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
    * **CardSourcePayload**: `{ProjectionMode: CardSourceProjectionAsHand, FieldMarks: []}`
  * **Template[C]** *(本回合禁用【漆黑之枪】【充盈】)*:
    * **ModifierID**: `rm_magic_lance_dark_liberation_disable_dark_spear_and_recharge_turn`
    * **Domain**: `RuleModifierDomainSkillGate`
    * **Priority**: `220`
    * **ConditionExpression**: `None`
    * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
    * **SkillGatePayload**: `{Mode: SkillGateDisallowList, SkillIDs: ["skill_magic_lance_dark_spear", "skill_magic_lance_recharge"]}`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionStart]`
  * **CustomExpression**: `Self.Orientation == Normal`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(横置)*:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Tapped`
    * `Ref`: `None`
  * **Effect[1]** *(进入幻影形态)*:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `"magic_lance_phantom_form"`
    * `Ref`: `None`
  * **Effect[2]** *(挂载“幻影形态手牌上限=5”规则模板)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_magic_lance_phantom_form_hand_limit_fixed_5"`
    * `RuleLifetimeRef`: `RuleLifePermanent`
    * `Ref`: `None`
  * **Effect[3]** *(挂载“本回合下次主动攻击+1”标记)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_magic_lance_next_active_attack_damage_plus_1_armed"`
    * `RuleLifetimeRef`: `RuleLifeUntilTurnEnd`
    * `Ref`: `None`
  * **Effect[4]** *(本回合禁用【漆黑之枪】【充盈】)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_magic_lance_dark_liberation_disable_dark_spear_and_recharge_turn"`
    * `RuleLifetimeRef`: `RuleLifeUntilTurnEnd`
    * `Ref`: `None`

#### 【暗之解放·下次主动攻击增伤】 (Dark Liberation Next Attack Bonus)
* **说明**：消费【暗之解放】挂载的“本回合下次主动攻击+1”标记。
* **1. 主干配置**：
  * **SkillID**: `skill_magic_lance_dark_liberation_next_attack_bonus`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnAttackDeclared`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDeclare]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && State.HasRuleModifier(Self.UserID, "rm_magic_lance_next_active_attack_damage_plus_1_armed") == true`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(本次主动攻击伤害+1)*:
    * `EffectType`: `EffectAttackDamageModifier`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `1`
    * `Ref`: `None`
  * **Effect[1]** *(消耗标记，确保“下次”只生效1次)*:
    * `EffectType`: `EffectRemoveRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleRemoveRef`: `{Mode: RuleRemoveByModifierID, ModifierID: "rm_magic_lance_next_active_attack_damage_plus_1_armed", Limit: 1}`
    * `Ref`: `None`

#### 【幻影星尘】 (Phantom Stardust)
* **技能描述**：（仅［幻影形态］下发动，对自己造成2点法术伤害③）［转正］脱离［幻影形态］；若没有因此造成我方士气下降，则对目标角色造成2点法术伤害③。
* **1. 主干配置**：
  * **SkillID**: `skill_magic_lance_phantom_stardust`
  * **Category**: `Normal`
  * **Type**: `Startup`
  * **Timing**: `TimingStartup`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionStart]`
  * **CustomExpression**: `Self.Form == "magic_lance_phantom_form" && Self.Orientation == Tapped`
* **3. 目标选择规则**：
  * **SelectType**: `Enemy`
  * **MinCount**: `1`
  * **MaxCount**: `1`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **实现约定**：`Effect[3]` 的条件读取 `Effect[0]` 自伤后的结果（是否产生士气下降）。
  * **Effect[0]** *(对自己造成2点法术伤害③)*:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelf`
    * `Value`: `2`
    * `Ref`: `None`
  * **Effect[1]** *(转正)*:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Normal`
    * `Ref`: `None`
  * **Effect[2]** *(脱离幻影形态)*:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `nil`
    * `Ref`: `None`
  * **Effect[3]** *(若自伤未造成我方士气下降，则对目标造成2点法术伤害③)*:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelected`
    * `Value`: `2`
    * `Condition`: `State.GetLastDamagePendingMoraleLoss(Self.UserID) == 0`
    * `Ref`: `None`

#### 【黑暗束缚】 (Dark Binding)
* **技能描述**：你始终不能使用法术牌。
* **1. 主干配置**：
  * **SkillID**: `skill_magic_lance_dark_binding`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingBeforeActionExecute`
* **2. 前置条件**：
  * **PhaseLimit**: `None`
  * **CustomExpression**: `Action.SourceID == Self.UserID && Action.CurrentType == Magic`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **3.1 行动改写配置**：
  * **ActionTransform.Hook**: `TimingBeforeActionExecute`
  * **ActionTransform.Optional**: `false`
  * **ActionTransform.Priority**: `220`
  * **ActionTransform.CancelCurrentAction**: `true`
  * **ActionTransform.Match.RequireActionType**: `Magic`
  * **ActionTransform.Match.RequirePlayedCardTypes**: `[]`
  * **ActionTransform.Match.RequirePlayedCardElements**: `[]`
  * **ActionTransform.Match.ExcludeTemplateIDs**: `[]`
  * **ActionTransform.Rewrite**: `nil` *(命中时直接取消本次法术行动)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * `None` *(本技能通过 ActionTransform 生效)*

#### 【暗之障壁】 (Dark Barrier)
* **技能描述**：（任何人对你造成伤害时发动③）弃X张法术牌或雷系牌［展示］。
* **1. 主干配置**：
  * **SkillID**: `skill_magic_lance_dark_barrier`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnDamageCalculated`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDamage]`
  * **CustomExpression**: `Combat.TargetID == Self.UserID && Combat.FinalDamage > 0 && State.ValidateUsedCardUUIDs(Action.UsedCardUUIDs, "(CardType == Magic || Element == Thunder)", Action.NamedValues["X"]) == true`
* **3. 目标选择规则**：
  * **SelectType**: `None`
  * **NamedValueConstraints**:
    * `{Key: "X", Required: true, MinExpression: "1", MaxExpression: "Self.HandCount"}`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Discards**:
    * `Count`: `Action.NamedValues["X"]`
    * `Filter`: `{}`
    * `Visibility`: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * `None` *(该技能文案仅要求执行弃牌)*

#### 【充盈】 (Recharge)
* **技能描述**：（弃1张法术牌或雷系牌［展示］）所有人各弃一张牌［展示］。我方角色可以选择不如此做，除你以外每以此法弃1张法术牌或雷系牌，本回合你的下次主动攻击伤害额外+1；额外+1［攻击行动］。
* **1. 主干配置**：
  * **SkillID**: `skill_magic_lance_recharge`
  * **Category**: `Normal`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **1.2 规则修饰器模板配置**：
  * **Template[A]** *(本回合“下次主动攻击增伤”武装标记)*:
    * **ModifierID**: `rm_magic_lance_recharge_next_attack_armed`
    * **Domain**: `RuleModifierDomainCardSource`
    * **Priority**: `80`
    * **ConditionExpression**: `None`
    * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
    * **CardSourcePayload**: `{ProjectionMode: CardSourceProjectionAsHand, FieldMarks: []}`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `State.ValidateUsedCardUUIDs(Action.UsedCardUUIDs, "(CardType == Magic || Element == Thunder)", 1) == true`
* **3. 目标选择规则**：
  * **SelectType**: `TeamOther`
  * **MinCount**: `0` *(我方角色“可选择不弃牌”)*
  * **MaxCount**: `1`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Discards**:
    * `Count`: `1`
    * `Filter`: `{}`
    * `Visibility`: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **实现约定**：在当前模型下，将“所有人各弃1”拆为“敌方强制弃1 + 我方队友可选弃1”，发动者自身由 `Cost.Discards` 覆盖。
  * **Effect[0]** *(敌方全体强制各弃1张牌［展示］)*:
    * `EffectType`: `EffectDiscard`
    * `Target`: `TargetAllEnemies`
    * `Value`: `1`
    * `Visibility`: `VisibilityPublic`
    * `Ref`: `None`
  * **Effect[1]** *(我方队友可选是否弃1张牌［展示］)*:
    * `EffectType`: `EffectPerTargetBranch`
    * `Target`: `TargetSelected`
    * `Value`: `0`
    * `BranchRef`:
      * `TargetSource`: `PerTargetSelectedTargets`
      * `InterruptType`: `WaitDiscard`
      * `TimeoutAsDeclined`: `true`
      * `DiscardRequirement`: `{Count: 1, Filter: {}}`
      * `DiscardVisibility`: `VisibilityPublic`
      * `OnSuccess`: `[]`
      * `OnDeclined`: `[]`
    * `Ref`: `None`
  * **Effect[2]** *(挂载“本回合下次主动攻击增伤”标记)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_magic_lance_recharge_next_attack_armed"`
    * `RuleLifetimeRef`: `RuleLifeUntilTurnEnd`
    * `Ref`: `None`
  * **Effect[3]** *(额外+1攻击行动)*:
    * `Implementation`: `model.AppendExtraAction`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `ActionRef`: `Attack`
    * `Ref`: `None`

#### 【充盈·下次主动攻击增伤】 (Recharge Next Attack Bonus)
* **说明**：读取本回合【充盈】弃牌轨迹；除你以外每以此法弃1张“法术牌或雷系牌”，你的下次主动攻击伤害额外+1。
* **1. 主干配置**：
  * **SkillID**: `skill_magic_lance_recharge_next_attack_bonus`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnAttackDeclared`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDeclare]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && State.HasRuleModifier(Self.UserID, "rm_magic_lance_recharge_next_attack_armed") == true`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(按本回合【充盈】弃牌记录动态加伤)*:
    * `EffectType`: `EffectAttackDamageModifier`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `State.CountLastSkillDiscards("skill_magic_lance_recharge", "(OwnerID != Self.UserID) && (CardType == Magic || Element == Thunder)")`
    * `Ref`: `None`
  * **Effect[1]** *(消费“下次主动攻击增伤”标记)*:
    * `EffectType`: `EffectRemoveRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleRemoveRef`: `{Mode: RuleRemoveByModifierID, ModifierID: "rm_magic_lance_recharge_next_attack_armed", Limit: 1}`
    * `Ref`: `None`

#### 【漆黑之枪】 (Pitch-Black Spear)
* **技能描述**：X［水晶］（仅［幻影形态］下，主动攻击手牌为1或2的对手并命中后发动②）本次攻击伤害额外+（X+2）。
* **1. 主干配置**：
  * **SkillID**: `skill_magic_lance_dark_spear`
  * **Category**: `Ultimate`
  * **Type**: `Response`
  * **Timing**: `TimingOnHitCheck`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatHitCheck]`
  * **CustomExpression**: `Self.Form == "magic_lance_phantom_form" && Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && Combat.IsHit == true && (Combat.TargetHandCount == 1 || Combat.TargetHandCount == 2)`
* **3. 目标选择规则**：
  * **SelectType**: `None`
  * **NamedValueConstraints**:
    * `{Key: "X", Required: true, MinExpression: "1", MaxExpression: "State.GetTeamStoneCount(Self.Team, Crystal)"}`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Crystal, Amount: Action.NamedValues["X"]}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectAttackDamageModifier`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `Action.NamedValues["X"] + 2`
    * `Ref`: `None`

---

### 27. 灵符师 (Talisman Master)

#### 【灵符-雷鸣】 (Talisman - Thunder Roar)
* **技能描述**：（弃1张雷系牌［展示］）对任意2名角色各造成1点法术伤害③。
* **1. 主干配置**：
  * **SkillID**: `skill_talisman_master_thunder_roar`
  * **Category**: `Normal`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `None`
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `2`
  * **MaxCount**: `2`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Discards**:
    * `Count`: `1`
    * `Filter`: `{ReqElement: Thunder}`
    * `Visibility`: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelected`
    * `Value`: `1`
    * `Ref`: `None`

#### 【灵符-风行】 (Talisman - Wind Walk)
* **技能描述**：（弃1张风系牌［展示］）指定2名角色各弃一张牌。
* **1. 主干配置**：
  * **SkillID**: `skill_talisman_master_wind_walk`
  * **Category**: `Normal`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `None`
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `2`
  * **MaxCount**: `2`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Discards**:
    * `Count`: `1`
    * `Filter`: `{ReqElement: Wind}`
    * `Visibility`: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectDiscard`
    * `Target`: `TargetSelected`
    * `Value`: `1`
    * `Ref`: `None`

#### 【念咒】 (Chant)
* **技能描述**：你每次发动灵符，可将自己1张手牌面朝下放置在你的角色旁，作为［妖力］。
* **1. 主干配置**：
  * **SkillID**: `skill_talisman_master_chant`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnSkillExecuted`
* **2. 前置条件**：
  * **PhaseLimit**: `None`
  * **CustomExpression**: `Event.OperatorID == Self.UserID && (Event.SkillID == "skill_talisman_master_thunder_roar" || Event.SkillID == "skill_talisman_master_wind_walk") && Self.HandCount >= 1`
* **3. 目标选择规则**：
  * **SelectType**: `None`
  * **SelectedValueRule**: `Action.SelectedValue == 0 || Action.SelectedValue == 1` *(0=不放置；1=放置1张手牌为妖力)*
  * **SubmitAction.UsedCardUUIDs**: `当 SelectedValue=1 时，提交1张自身手牌 CardUUID。`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectPlaceHandCardAsFieldMark`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `FieldMarkRef`: `DemonForce`
    * `VisibilityRef`: `VisibilityHidden`
    * `Condition`: `Action.SelectedValue == 1`
    * `Ref`: `None`

#### 【百鬼夜行】 (Night Parade of One Hundred Demons)
* **技能描述**：（主动攻击命中后发动②，移除1个妖力）对目标角色造成1点法术伤害③；（若［妖力］为火系牌，可展示之［展示］）改为指定2名角色，对除他们以外的其他所有角色各造成1点法术伤害③。
* **1. 主干配置**：
  * **SkillID**: `skill_talisman_master_hundred_ghost_night`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnHitCheck`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatHitCheck]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && Combat.IsHit == true && State.CountFieldMark(Self.UserID, DemonForce) >= 1`
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `1`
  * **MaxCount**: `2`
  * **SelectedValueRule**: `(Action.SelectedValue == 0 && Action.SelectedTargetCount == 1) || (Action.SelectedValue == 1 && Action.SelectedTargetCount == 2 && State.GetSelectedFieldCardElement(Self.UserID, Action.Targets) == Fire)` *(0=普通分支；1=火妖力展示分支)*
  * **SubSelect**：
  * `SubType`: `FieldCard`
  * `SubFilter`: `FieldCard.HolderID == Self.UserID && FieldCard.FieldMark == DemonForce`
  * `SubMinCount`: `1`
  * `SubMaxCount`: `1`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列（分支）**：
  * **Effect[0]** *(按二级选择精确移除1张妖力，并写入 Event.RemovedFieldCard* 上下文)*:
    * `EffectType`: `EffectRemoveSelectedFieldCard`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `Ref`: `None`
  * **BranchSelector**: `Action.SelectedValue`
  * **Branch[0]** *(普通分支：对所选1名目标造成1点法术伤害)*:
    * **Effect[B0]**:
      * `EffectType`: `EffectDamage`
      * `Target`: `TargetSelected`
      * `Value`: `1`
      * `Ref`: `None`
  * **Branch[1]** *(火妖力展示分支：展示后，对“除所选2名外其余所有角色”各1点法伤)*:
    * **Effect[F0]** *(展示被移除妖力)*:
      * `EffectType`: `EffectRevealRemovedFieldCard`
      * `Target`: `TargetSelf`
      * `Value`: `0`
      * `Condition`: `Event.RemovedFieldCardElement == Fire`
      * `Ref`: `None`
    * **Effect[F1]** *(排除所选目标后的全体伤害)*:
      * `EffectType`: `EffectDamage`
      * `Target`: `TargetAllExceptSelected`
      * `Value`: `1`
      * `Condition`: `Event.RemovedFieldCardElement == Fire`
      * `Ref`: `None`

#### 【灵力崩解】 (Spiritual Collapse)
* **技能描述**：［水晶］（和［灵符-雷鸣］或［百鬼夜行］同时发动）你的本次［灵符-雷鸣］或［百鬼夜行］每次造成的伤害额外+1。
* **1. 主干配置**：
  * **SkillID**: `skill_talisman_master_spiritual_collapse`
  * **Category**: `Ultimate`
  * **Type**: `Response`
  * **Timing**: `TimingOnSkillExecuted`
* **1.2 规则修饰器模板配置**：
  * **ModifierID**: `rm_talisman_master_spiritual_collapse_armed`
  * **Domain**: `RuleModifierDomainCardSource`
  * **Priority**: `120`
  * **ConditionExpression**: `None`
  * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
  * **CardSourcePayload**: `{ProjectionMode: CardSourceProjectionAsHand, FieldMarks: []}`
* **2. 前置条件**：
  * **PhaseLimit**: `None`
  * **CustomExpression**: `Event.OperatorID == Self.UserID && (Event.SkillID == "skill_talisman_master_thunder_roar" || Event.SkillID == "skill_talisman_master_hundred_ghost_night")`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Crystal, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(挂载“本次灵符伤害+1”标记，仅作用当前技能结算链)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_talisman_master_spiritual_collapse_armed"`
    * `RuleLifetimeRef`: `RuleLifeThisEffectChain`
    * `Ref`: `None`

#### 【灵力崩解·伤害增幅】 (Spiritual Collapse Damage Boost)
* **说明**：当【灵力崩解】已挂载且当前伤害来源为【灵符-雷鸣】或【百鬼夜行】时，将该次待生效伤害+1。
* **1. 主干配置**：
  * **SkillID**: `skill_talisman_master_spiritual_collapse_damage_boost`
  * **Category**: `Ultimate`
  * **Type**: `Passive`
  * **Timing**: `TimingOnDamageCalculated`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDamage]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && State.HasRuleModifier(Self.UserID, "rm_talisman_master_spiritual_collapse_armed") == true && (Event.CauseAction == "skill_talisman_master_thunder_roar" || Event.CauseAction == "skill_talisman_master_hundred_ghost_night")`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectModifyPendingDamage`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `1`
    * `Ref`: `None`

---

### 28. 吟游诗人 (Bard)

#### 【沉沦协奏曲】 (Fallen Concerto)
* **技能描述**：［回合限定］（仅［普通形态］下，一回合内我方至少2名对手造成法术伤害且结算完之后，弃2张同系牌［展示］［强制］）你+1［灵感］。（若弃牌中有法术牌）对目标对手造成1点法术伤害③。
* **1. 主干配置**：
  * **SkillID**: `skill_bard_fallen_concerto`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnDamageApplied`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatApply]`
  * **IsTurnLimited**: `true`
  * **CustomExpression**: `State.IsInSelfTurn(Self.UserID) == true && Self.Orientation == Normal && Combat.SourceID == Self.UserID && Combat.IsMagic == true && State.CountTurnDistinctDamageTargets(Self.UserID, true, true) >= 2`
* **3. 目标选择规则**：
  * **SelectType**: `Enemy`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **SelectedValueRule**: `Action.SelectedValue == 0 || Action.SelectedValue == 1` *(0=弃牌中不含法术牌；1=弃牌中含至少1张法术牌)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Discards**: `Count: 2, Filter: {SameAttribute: MatchElement}`
  * **DiscardsVisibility**: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `TokenRef`: `Inspiration`
    * `Ref`: `None`
  * **Effect[1]** *(若本次弃牌中包含法术牌，则对目标对手造成1点法术伤害)*:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelected`
    * `Value`: `1`
    * `Condition`: `Action.SelectedValue == 1`
    * `Ref`: `None`

#### 【不谐和弦】 (Dissonant Chord)
* **技能描述**：（移除X点［灵感］，X>1。若你处于［永恒囚徒形态］，［转正］脱离［永恒囚徒形态］，你选择以下一项发动：●你和目标角色各摸（X-1）张牌［强制］。●你和目标角色各弃（X-1）张牌。
* **1. 主干配置**：
  * **SkillID**: `skill_bard_dissonant_chord`
  * **Category**: `Normal`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Self.Tokens[Inspiration] >= 2`
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **NamedValueConstraints**: `[{Key: "X", Required: true, MinExpression: "2", MaxExpression: "Self.Tokens[Inspiration]"}]`
  * **SelectedValueRule**: `Action.SelectedValue == 0 || Action.SelectedValue == 1` *(0=双方摸牌；1=双方弃牌)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Tokens**: `[{Type: Inspiration, Amount: Action.NamedValues["X"]}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列（分支）**：
  * **Effect[0]** *(若处于永恒囚徒形态：先转正)*:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Normal`
    * `Condition`: `Self.Orientation == Tapped && Self.Form == "bard_eternal_prisoner_form"`
    * `Ref`: `None`
  * **Effect[1]** *(若处于永恒囚徒形态：清空形态名)*:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `nil`
    * `Condition`: `Self.Form == "bard_eternal_prisoner_form"`
    * `Ref`: `None`
  * **BranchSelector**: `Action.SelectedValue`
  * **Branch[0]** *(你和目标各摸(X-1)张)*:
    * **Effect[B0]**:
      * `EffectType`: `EffectDrawCard`
      * `Target`: `TargetSelf`
      * `Value`: `Action.NamedValues["X"] - 1`
      * `Ref`: `None`
    * **Effect[B1]**:
      * `EffectType`: `EffectDrawCard`
      * `Target`: `TargetSelected`
      * `Value`: `Action.NamedValues["X"] - 1`
      * `Ref`: `None`
  * **Branch[1]** *(你和目标各弃(X-1)张)*:
    * **Effect[D0]**:
      * `EffectType`: `EffectDiscard`
      * `Target`: `TargetSelf`
      * `Value`: `Action.NamedValues["X"] - 1`
      * `Ref`: `None`
    * **Effect[D1]**:
      * `EffectType`: `EffectDiscard`
      * `Target`: `TargetSelected`
      * `Value`: `Action.NamedValues["X"] - 1`
      * `Ref`: `None`

#### 【禁忌诗篇】 (Forbidden Hymn)
* **技能描述**：（［激昂狂想曲］或［胜利交响诗］的效果结算完后）根据［灵感］数量：●（［灵感］未达上限）你+1［灵感］，移除永恒乐章。●（灵感已达上限）对自己造成3点法术伤害③。若你处于［普通形态］，［横置］转为［永恒囚徒形态］。
* **1. 主干配置**：
  * **SkillID**: `skill_bard_forbidden_hymn`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnSkillExecuted`
* **2. 前置条件**：
  * **PhaseLimit**: `None`
  * **CustomExpression**: `Event.OperatorID == Self.UserID && (Event.SkillID == "skill_bard_fervent_rhapsody" || Event.SkillID == "skill_bard_victory_symphony")`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(灵感未达上限：你+1灵感)*:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `TokenRef`: `Inspiration`
    * `Condition`: `State.IsTokenAtCap(Self.UserID, Inspiration) == false`
    * `Ref`: `None`
  * **Effect[1]** *(灵感未达上限：移除我方永恒乐章)*:
    * `EffectType`: `EffectRemoveFieldMark`
    * `Target`: `TargetAllTeammates`
    * `Value`: `1`
    * `FieldMarkRef`: `EternalMovement`
    * `Condition`: `State.IsTokenAtCap(Self.UserID, Inspiration) == false`
    * `Ref`: `None`
  * **Effect[2]** *(灵感已达上限：对自己造成3点法术伤害)*:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelf`
    * `Value`: `3`
    * `Condition`: `State.IsTokenAtCap(Self.UserID, Inspiration) == true`
    * `Ref`: `None`
  * **Effect[3]** *(灵感已达上限且当前为普通姿态：横置进入永恒囚徒形态)*:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Tapped`
    * `Condition`: `State.IsTokenAtCap(Self.UserID, Inspiration) == true && Self.Orientation == Normal`
    * `Ref`: `None`
  * **Effect[4]**:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `"bard_eternal_prisoner_form"`
    * `Condition`: `State.IsTokenAtCap(Self.UserID, Inspiration) == true && Self.Orientation == Normal`
    * `Ref`: `None`

#### 【激昂狂想曲】 (Fervent Rhapsody)
* **技能描述**：（回合开始时若你拥有永恒乐章）选择以下一项执行：●吟游诗人对2名目标对手各造成1点法术伤害③。●你弃2张牌。
* **1. 主干配置**：
  * **SkillID**: `skill_bard_fervent_rhapsody`
  * **Category**: `Exclusive`
  * **Type**: `Response`
  * **Timing**: `TimingOnTurnStart`
* **2. 前置条件**：
  * **PhaseLimit**: `[TurnStart]`
  * **CustomExpression**: `State.IsInSelfTurn(Self.UserID) == true && State.CountTeamFieldMark(Self.UserID, EternalMovement) >= 1 && Action.PlayedCard.CharacterSkillMap[Self.CharacterID] == "skill_bard_fervent_rhapsody"`
* **3. 目标选择规则**：
  * **SelectType**: `Enemy`
  * **MinCount**: `0`
  * **MaxCount**: `2`
  * **SelectedValueRule**: `(Action.SelectedValue == 0 && Action.SelectedTargetCount == 2) || (Action.SelectedValue == 1 && Action.SelectedTargetCount == 0)` *(0=伤害分支；1=弃牌分支)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列（分支）**：
  * **BranchSelector**: `Action.SelectedValue`
  * **Branch[0]** *(对2名目标对手各造成1点法术伤害)*:
    * **Effect[B0]**:
      * `EffectType`: `EffectDamage`
      * `Target`: `TargetSelected`
      * `Value`: `1`
      * `Ref`: `None`
  * **Branch[1]** *(你弃2张牌)*:
    * **Effect[D0]**:
      * `EffectType`: `EffectDiscard`
      * `Target`: `TargetSelf`
      * `Value`: `2`
      * `Ref`: `None`

#### 【胜利交响诗】 (Victory Symphony)
* **技能描述**：（回合结束时若你拥有永恒乐章）选择以下一项执行：●将我方战绩区的1个星石提炼成为你的能量。●为我方战绩区+1［宝石］，你+1［治疗］。
* **1. 主干配置**：
  * **SkillID**: `skill_bard_victory_symphony`
  * **Category**: `Exclusive`
  * **Type**: `Response`
  * **Timing**: `TimingOnTurnEnd`
* **2. 前置条件**：
  * **PhaseLimit**: `[TurnEnd]`
  * **CustomExpression**: `State.IsInSelfTurn(Self.UserID) == true && State.CountTeamFieldMark(Self.UserID, EternalMovement) >= 1 && Action.PlayedCard.CharacterSkillMap[Self.CharacterID] == "skill_bard_victory_symphony"`
* **3. 目标选择规则**：
  * **SelectType**: `None`
  * **SelectedValueRule**: `(Action.SelectedValue == 0 && (Action.StoneRef == Gem || Action.StoneRef == Crystal)) || Action.SelectedValue == 1` *(0=提炼1星石为自身能量；1=战绩区+1宝石并自疗+1)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列（分支）**：
  * **BranchSelector**: `Action.SelectedValue`
  * **Branch[0]** *(将我方战绩区1个指定颜色星石提炼为你的能量)*:
    * **Effect[E0]**:
      * `EffectType`: `EffectAddTeamStone`
      * `Target`: `TargetSelfTeam`
      * `Value`: `-1`
      * `StoneRef`: `Action.StoneRef`
      * `Ref`: `None`
    * **Effect[E1]**:
      * `EffectType`: `EffectAddEnergyStone`
      * `Target`: `TargetSelf`
      * `Value`: `1`
      * `StoneRef`: `Action.StoneRef`
      * `Ref`: `None`
  * **Branch[1]** *(我方战绩区+1宝石，你+1治疗)*:
    * **Effect[B0]**:
      * `EffectType`: `EffectAddTeamStone`
      * `Target`: `TargetSelfTeam`
      * `Value`: `1`
      * `StoneRef`: `Gem`
      * `Ref`: `None`
    * **Effect[B1]**:
      * `EffectType`: `EffectHeal`
      * `Target`: `TargetSelf`
      * `Value`: `1`
      * `Ref`: `None`

#### 【希望赋格曲】 (Hope Fugue)
* **技能描述**：［水晶］你可以选择摸1张牌，然后选择以下一项发动：●将［永恒乐章］放置于目标队友面前。●将［永恒乐章］转移给我方另一名目标角色，你弃1张牌，+1［治疗］或+1［灵感］。
* **1. 主干配置**：
  * **SkillID**: `skill_bard_hope_fugue`
  * **Category**: `Ultimate`
  * **Type**: `Startup`
  * **Timing**: `TimingStartup`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionStart]`
  * **CustomExpression**: `Action.PlayedCard.CharacterSkillMap[Self.CharacterID] == "skill_bard_hope_fugue" && (Action.SelectedValue == 0 || State.CountTeamFieldMark(Self.UserID, EternalMovement) >= 1)`
* **3. 目标选择规则**：
  * **SelectType**: `TeamOther`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **NamedValueConstraints**: `[{Key: "Draw", Required: false, MinExpression: "0", MaxExpression: "1"}]`
  * **SelectedValueRule**: `Action.SelectedValue == 0 || Action.SelectedValue == 1 || Action.SelectedValue == 2` *(0=放置；1=转移后+1治疗；2=转移后+1灵感)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayRequired`
  * **Stones**: `[{Type: Crystal, Amount: 1}]`
  * **Discards**: `Count: Action.SelectedValue == 0 ? 0 : 1`
  * **DiscardsVisibility**: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列（分支）**：
  * **Effect[0]** *(可选摸1张牌)*:
    * `EffectType`: `EffectDrawCard`
    * `Target`: `TargetSelf`
    * `Value`: `Action.NamedValues["Draw"]`
    * `Ref`: `None`
  * **BranchSelector**: `Action.SelectedValue`
  * **Branch[0]** *(将永恒乐章放置于目标队友面前)*:
    * **Effect[P0]** *(先移除我方现有永恒乐章，保证唯一性)*:
      * `EffectType`: `EffectRemoveFieldMark`
      * `Target`: `TargetAllTeammates`
      * `Value`: `1`
      * `FieldMarkRef`: `EternalMovement`
      * `Ref`: `None`
    * **Effect[P1]**:
      * `EffectType`: `EffectPlacePlayedCardAsFieldMark`
      * `Target`: `TargetSelected`
      * `Value`: `1`
      * `FieldMarkRef`: `EternalMovement`
      * `VisibilityRef`: `VisibilityPublic`
      * `Ref`: `None`
  * **Branch[1]** *(转移永恒乐章；弃1后+1治疗)*:
    * **Effect[T0]** *(迁移“在场实体”的永恒乐章到目标队友，不重放新实体)*:
      * `EffectType`: `EffectTransferFieldMark`
      * `Target`: `TargetSelected`
      * `FromTargetRef`: `TargetAllTeammates`
      * `Value`: `1`
      * `FieldMarkRef`: `EternalMovement`
      * `Ref`: `None`
    * **Effect[T1]**:
      * `EffectType`: `EffectHeal`
      * `Target`: `TargetSelf`
      * `Value`: `1`
      * `Ref`: `None`
  * **Branch[2]** *(转移永恒乐章；弃1后+1灵感)*:
    * **Effect[I0]** *(迁移“在场实体”的永恒乐章到目标队友，不重放新实体)*:
      * `EffectType`: `EffectTransferFieldMark`
      * `Target`: `TargetSelected`
      * `FromTargetRef`: `TargetAllTeammates`
      * `Value`: `1`
      * `FieldMarkRef`: `EternalMovement`
      * `Ref`: `None`
    * **Effect[I1]**:
      * `EffectType`: `EffectAddToken`
      * `Target`: `TargetSelf`
      * `Value`: `1`
      * `TokenRef`: `Inspiration`
      * `Ref`: `None`

---

### 29. 勇者 (Hero)

#### 【勇者之心】 (Heart of Hero)
* **技能描述**：游戏初始时，你+2［水晶］。
* **1. 主干配置**：
  * **SkillID**: `skill_hero_heart_of_hero`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnTurnStart`
* **2. 前置条件**：
  * **PhaseLimit**: `[GameInit]`
  * **CustomExpression**: `Event.SourceType == SourceSystem && Event.CauseAction == "GameStart"`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectAddEnergyStone`
    * `Target`: `TargetSelf`
    * `Value`: `2`
    * `StoneRef`: `Crystal`
    * `Ref`: `None`

#### 【怒吼】 (Roar)
* **技能描述**：（主动攻击时发动①，移除1点［怒气］）你摸0-1张牌，本次攻击伤害额外+2；（若未命中②）你+1［知性］。
* **1. 主干配置**：
  * **SkillID**: `skill_hero_roar`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnAttackDeclared`
* **1.2 规则修饰器模板配置**：
  * **ModifierID**: `rm_hero_roar_armed`
  * **Domain**: `RuleModifierDomainCardSource`
  * **Priority**: `120`
  * **ConditionExpression**: `None`
  * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
  * **CardSourcePayload**: `{ProjectionMode: CardSourceProjectionAsHand, FieldMarks: []}` *(空载荷：仅作“本次攻击已发动怒吼”的链路标记)*
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDeclare]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && Self.Tokens[Rage] >= 1`
* **3. 目标选择规则**：
  * **SelectType**: `None`
  * **SelectedValueRule**: `Action.SelectedValue == 0 || Action.SelectedValue == 1` *(0=不摸牌；1=摸1张)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Tokens**: `[{Type: Rage, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(挂载“怒吼已发动”标记，供未命中分支读取)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_hero_roar_armed"`
    * `RuleLifetimeRef`: `RuleLifeUntilCombatEnd`
    * `Ref`: `None`
  * **Effect[1]** *(摸0-1张牌)*:
    * `EffectType`: `EffectDrawCard`
    * `Target`: `TargetSelf`
    * `Value`: `Action.SelectedValue`
    * `Ref`: `None`
  * **Effect[2]** *(本次攻击伤害额外+2)*:
    * `EffectType`: `EffectAttackDamageModifier`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `2`
    * `Ref`: `None`

#### 【怒吼·未命中】 (Roar Miss Bonus)
* **说明**：若本次攻击发动过【怒吼】且未命中，则你+1［知性］。
* **1. 主干配置**：
  * **SkillID**: `skill_hero_roar_miss_bonus`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnHitCheck`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatHitCheck]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && Combat.IsHit == false && State.HasRuleModifier(Self.UserID, "rm_hero_roar_armed") == true`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `TokenRef`: `Intellect`
    * `Ref`: `None`

#### 【精疲力竭】 (Exhaustion)
* **技能描述**：（发动禁断之力后强制触发［强制］）［横置］额外+1［攻击行动］；持续到你的下个行动阶段开始，你的手牌上限恒定为4。［精疲力竭］的效果结束时角色［转正］，并对自己造成3点法术伤害③。
* **1. 主干配置**：
  * **SkillID**: `skill_hero_exhaustion`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnSkillExecuted`
* **1.2 规则修饰器模板配置**：
  * **ModifierID**: `rm_hero_exhaustion_hand_limit_fixed_4`
  * **Domain**: `RuleModifierDomainAttribute`
  * **Priority**: `220`
  * **ConditionExpression**: `None`
  * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
  * **AttrPayload**: `{AttrType: PlayerAttributeMaxHand, Operation: AttributeModifySet, ValueSourceMode: RuleAttrValueFromFixed, Value: 4}`
* **2. 前置条件**：
  * **PhaseLimit**: `None`
  * **CustomExpression**: `Event.OperatorID == Self.UserID && Event.SkillID == "skill_hero_forbidden_power"`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(横置进入精疲力竭形态)*:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Tapped`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `"hero_exhaustion_form"`
    * `Ref`: `None`
  * **Effect[2]** *(额外+1攻击行动)*:
    * `Implementation`: `model.AppendExtraAction`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `ActionRef`: `Attack`
    * `Ref`: `None`
  * **Effect[3]** *(持续到下个行动阶段开始：手牌上限固定为4)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_hero_exhaustion_hand_limit_fixed_4"`
    * `RuleLifetimeRef`: `RuleLifeUntilSourceNextTurnStart`
    * `Ref`: `None`

#### 【精疲力竭·结束】 (Exhaustion End)
* **说明**：在你的下个行动阶段开始时，结束【精疲力竭】形态并转正，然后对自己造成3点法术伤害。
* **1. 主干配置**：
  * **SkillID**: `skill_hero_exhaustion_end`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingStartup`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionStart]`
  * **CustomExpression**: `State.IsInSelfTurn(Self.UserID) == true && Self.Orientation == Tapped && Self.Form == "hero_exhaustion_form"`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(转正)*:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Normal`
    * `Ref`: `None`
  * **Effect[1]** *(退出形态)*:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `nil`
    * `Ref`: `None`
  * **Effect[2]**:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelf`
    * `Value`: `3`
    * `Ref`: `None`

#### 【明镜止水】 (Clear Mind)
* **技能描述**：（主动攻击前发动①，移除4点［知性］）本次攻击对手无法应战。（本次攻击结束时）你+1［水晶］。
* **1. 主干配置**：
  * **SkillID**: `skill_hero_clear_mind`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnAttackDeclared`
* **1.2 规则修饰器模板配置**：
  * **ModifierID**: `rm_hero_clear_mind_armed`
  * **Domain**: `RuleModifierDomainCardSource`
  * **Priority**: `120`
  * **ConditionExpression**: `None`
  * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
  * **CardSourcePayload**: `{ProjectionMode: CardSourceProjectionAsHand, FieldMarks: []}` *(空载荷：仅作“本次攻击已发动明镜止水”的链路标记)*
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDeclare]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && Self.Tokens[Intellect] >= 4`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Tokens**: `[{Type: Intellect, Amount: 4}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `Unrespondable`
* **6. 执行效果序列**：
  * **Effect[0]** *(挂载“明镜止水已发动”标记，供行动结束收益读取)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_hero_clear_mind_armed"`
    * `RuleLifetimeRef`: `RuleLifeUntilTurnEnd`
    * `Ref`: `None`

#### 【明镜止水·行动结束收益】 (Clear Mind End Gain)
* **说明**：若当前攻击行动发动过【明镜止水】，则在本次攻击行动结束时你+1［水晶］，并移除该标记。
* **1. 主干配置**：
  * **SkillID**: `skill_hero_clear_mind_end_gain`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnActionEnd`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionEnd]`
  * **CustomExpression**: `State.IsInSelfTurn(Self.UserID) == true && Action.SourceID == Self.UserID && Action.CurrentType == Attack && State.HasRuleModifier(Self.UserID, "rm_hero_clear_mind_armed") == true`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectAddEnergyStone`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `StoneRef`: `Crystal`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectRemoveRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleRemoveRef`: `{Mode: RuleRemoveByModifierID, ModifierID: "rm_hero_clear_mind_armed", Limit: 1}`
    * `Ref`: `None`

#### 【挑衅】 (Provocation)
* **技能描述**：（移除1点［怒气］）将挑衅放置于目标对手面前，你+1［知性］；该对手在其下个行动阶段必须且只能主动攻击你，否则他跳过该行动阶段，触发后移除此牌。
* **1. 主干配置**：
  * **SkillID**: `skill_hero_provocation`
  * **Category**: `Exclusive`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Action.PlayedCard.CharacterSkillMap[Self.CharacterID] == "skill_hero_provocation" && Self.Tokens[Rage] >= 1`
* **3. 目标选择规则**：
  * **SelectType**: `Enemy`
  * **MinCount**: `1`
  * **MaxCount**: `1`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayRequired`
  * **Tokens**: `[{Type: Rage, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(将挑衅状态放置于目标对手；状态来源自动记录为本技能发动者)*:
    * `EffectType`: `EffectPlaceStatus`
    * `Target`: `TargetSelected`
    * `Value`: `1`
    * `StatusRef`: `Taunt`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `TokenRef`: `Intellect`
    * `Ref`: `None`
* **7. 状态结算行为**（挂载于 `StatusEffect.Taunt`，轻量动作锁）：
  * **ResolveTiming**: `TimingBeforeActionExecute` *(状态持有者下个行动阶段尝试提交行动时判定)*
  * **RequireHolderIsTurnPlayer**: `true`
  * **ResolveMode**: `Auto`
  * **CanDecline**: `false`
  * **EnforceNextActionMustActiveAttackSource**: `true` *(必须“主动攻击且目标=StatusMeta.SourceUserID”，否则跳过当前行动阶段)*
  * **TriggerLimit**: `1`
  * **RemoveAfterResolve**: `true`

#### 【禁断之力】 (Forbidden Power)
* **技能描述**：［水晶］（主动攻击命中或未命中后发动②）弃掉你所有手牌［展示］，其中每有1张法术牌，你+1［怒气］；（若未命中②）其中每有1张水系牌，你+1［知性］；（若命中②）其中每有1张火系牌，本次攻击伤害额外+1，并对自己造成等同于火系牌数量的法术伤害③。
* **1. 主干配置**：
  * **SkillID**: `skill_hero_forbidden_power`
  * **Category**: `Ultimate`
  * **Type**: `Response`
  * **Timing**: `TimingOnHitCheck`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatHitCheck]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Crystal, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(弃掉你所有手牌［展示］，并将批量弃牌统计写入 `Event.Discarded*` 上下文)*:
    * `EffectType`: `EffectDiscard`
    * `Target`: `TargetSelf`
    * `Value`: `Self.HandCount`
    * `Visibility`: `VisibilityPublic`
    * `Ref`: `None`
  * **Effect[1]** *(每1张法术牌：你+1怒气)*:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `Event.DiscardedMagicCount`
    * `TokenRef`: `Rage`
    * `Ref`: `None`
  * **Effect[2]** *(若未命中：每1张水系牌，你+1知性)*:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `Event.DiscardedElementCount[Water]`
    * `TokenRef`: `Intellect`
    * `Condition`: `Combat.IsHit == false`
    * `Ref`: `None`
  * **Effect[3]** *(若命中：每1张火系牌，本次攻击伤害额外+1)*:
    * `EffectType`: `EffectAttackDamageModifier`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `Event.DiscardedElementCount[Fire]`
    * `Condition`: `Combat.IsHit == true`
    * `Ref`: `None`
  * **Effect[4]** *(若命中：对自己造成等同火系弃牌数量的法术伤害)*:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelf`
    * `Value`: `Event.DiscardedElementCount[Fire]`
    * `Condition`: `Combat.IsHit == true`
    * `Ref`: `None`

#### 【死斗】 (Deadlock)
* **技能描述**：［宝石］（每当你承受法术伤害时发动⑥）你+3［怒气］；（若此伤害造成士气实际下降）本次的士气下降值恒定为1。
* **1. 主干配置**：
  * **SkillID**: `skill_hero_deadlock`
  * **Category**: `Ultimate`
  * **Type**: `Response`
  * **Timing**: `TimingOnDamageTaken`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDraw]`
  * **CustomExpression**: `Combat.TargetID == Self.UserID && Combat.IsMagic == true && Combat.FinalDamage > 0`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Gem, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `3`
    * `TokenRef`: `Rage`
    * `Ref`: `None`
  * **Effect[1]** *(若本窗口待扣士气>1，则下调到1；若原本为0或1则不改)*:
    * `EffectType`: `EffectReducePendingMoraleLoss`
    * `Target`: `TargetCurrentEvent`
    * `Value`: `Event.PendingMoraleLoss - 1`
    * `Condition`: `Event.PendingMoraleLoss > 1`
    * `Ref`: `None`

---

### 30. 格斗家 (Fighter)

#### 【念气力场】 (Nen Barrier)
* **技能描述**：所有对你造成的伤害每次最高为4点③。
* **1. 主干配置**：
  * **SkillID**: `skill_fighter_nen_barrier`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnDamageCalculated`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDamage]`
  * **CustomExpression**: `Combat.TargetID == Self.UserID && Combat.FinalDamage > 4`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(将本次待生效伤害下调到4)*:
    * `EffectType`: `EffectModifyPendingDamage`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `4 - Combat.FinalDamage`
    * `Ref`: `None`

#### 【蓄力一击】 (Charged Strike)
* **技能描述**：（主动攻击前发动①，+1［斗气］）本次攻击伤害额外+1；（若未命中②）对自己造成X点法术伤害③，X为你所拥有的［斗气］数；（若［斗气］已经达到上限）你不能发动蓄力一击。
* **1. 主干配置**：
  * **SkillID**: `skill_fighter_charged_strike`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnAttackDeclared`
* **1.2 响应分组配置**：
  * **ResponseGroup.GroupID**: `rg_fighter_attack_prefix_choice`
  * **ResponseGroup.Mode**: `ResponseGroupChooseOne`
  * **ResponseGroup.OptionOrder**: `1`
* **1.3 规则修饰器模板配置**：
  * **ModifierID**: `rm_fighter_charged_strike_armed`
  * **Domain**: `RuleModifierDomainCardSource`
  * **Priority**: `120`
  * **ConditionExpression**: `None`
  * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
  * **CardSourcePayload**: `{ProjectionMode: CardSourceProjectionAsHand, FieldMarks: []}` *(空载荷，仅作“本次攻击已发动蓄力一击”的链路标记)*
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDeclare]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && State.IsTokenAtCap(Self.UserID, BattleQi) == false && State.IsSkillDisabled(Self.UserID, "skill_fighter_charged_strike") == false`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(挂载“已发动蓄力一击”标记，供未命中自伤分支读取)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_fighter_charged_strike_armed"`
    * `RuleLifetimeRef`: `RuleLifeUntilCombatEnd`
    * `Ref`: `None`
  * **Effect[1]** *(+1斗气)*:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `TokenRef`: `BattleQi`
    * `Ref`: `None`
  * **Effect[2]** *(本次攻击伤害额外+1)*:
    * `EffectType`: `EffectAttackDamageModifier`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `1`
    * `Ref`: `None`

#### 【蓄力一击·未命中自伤】 (Charged Strike Miss Self Damage)
* **说明**：若本次攻击发动过【蓄力一击】且未命中，则对自己造成X点法术伤害（X=当前斗气数）。
* **1. 主干配置**：
  * **SkillID**: `skill_fighter_charged_strike_miss_self_damage`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnHitCheck`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatHitCheck]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && Combat.IsHit == false && State.HasRuleModifier(Self.UserID, "rm_fighter_charged_strike_armed") == true`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelf`
    * `Value`: `Self.Tokens[BattleQi]`
    * `Ref`: `None`

#### 【念弹】 (Nen Bullet)
* **技能描述**：（［法术行动］结束时发动）+1［斗气］，对目标对手造成1点法术伤害③，（若发动前对方的［治疗］为0）对自己造成X点法术伤害③，X为你拥有的［斗气］数；（若［斗气］已达到上限）你不能发动［念弹］。
* **1. 主干配置**：
  * **SkillID**: `skill_fighter_nen_bullet`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnActionEnd`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionEnd]`
  * **CustomExpression**: `State.IsInSelfTurn(Self.UserID) == true && Action.SourceID == Self.UserID && Action.CurrentType == Magic && State.IsTokenAtCap(Self.UserID, BattleQi) == false`
* **3. 目标选择规则**：
  * **SelectType**: `Enemy`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **SelectedValueRule**: `(Action.SelectedValue == 0 && Target.Heal > 0) || (Action.SelectedValue == 1 && Target.Heal == 0)` *(0=目标发动前治疗>0，不触发自伤；1=目标发动前治疗=0，触发自伤)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(+1斗气)*:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `TokenRef`: `BattleQi`
    * `Ref`: `None`
  * **Effect[1]** *(对目标对手造成1点法术伤害)*:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelected`
    * `Value`: `1`
    * `Ref`: `None`
  * **Effect[2]** *(若发动前目标治疗为0：对自己造成X点法术伤害，X=当前斗气数)*:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelf`
    * `Value`: `Self.Tokens[BattleQi]`
    * `Condition`: `Action.SelectedValue == 1`
    * `Ref`: `None`

#### 【百式幻龙拳】 (Hundred Dragon Fist)
* **技能描述**：［持续］（移除3点［斗气］，［横置］）你的所有主动攻击伤害额外+2，所有应战攻击伤害额外+1；在你接下来的行动阶段，你不能执行［法术行动］和［特殊行动］；你的主动攻击必须以同一名角色为目标，并且不能发动［蓄力一击］；若不如此做，则取消［百式幻龙拳］的效果并［转正］。
* **1. 主干配置**：
  * **SkillID**: `skill_fighter_hundred_dragon_fist`
  * **Category**: `Normal`
  * **Type**: `Startup`
  * **Timing**: `TimingStartup`
* **1.2 规则修饰器模板配置**：
  * **Template[A]** *(百式期间禁用【蓄力一击】)*:
    * **ModifierID**: `rm_fighter_hundred_dragon_disable_charged_strike`
    * **Domain**: `RuleModifierDomainSkillGate`
    * **Priority**: `240`
    * **ConditionExpression**: `None`
    * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
    * **SkillGatePayload**: `{Mode: SkillGateDisallowList, SkillIDs: ["skill_fighter_charged_strike"]}`
  * **Template[B]** *(百式锁定目标标记，挂在被锁定目标身上)*:
    * **ModifierID**: `rm_fighter_hundred_dragon_locked_target_marker`
    * **Domain**: `RuleModifierDomainCardSource`
    * **Priority**: `80`
    * **ConditionExpression**: `None`
    * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
    * **CardSourcePayload**: `{ProjectionMode: CardSourceProjectionAsHand, FieldMarks: []}` *(空载荷，仅作“被百式锁定目标”标记)*
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionStart]`
  * **CustomExpression**: `State.IsInSelfTurn(Self.UserID) == true && Self.Tokens[BattleQi] >= 3`
* **3. 目标选择规则**：
  * **SelectType**: `Enemy`
  * **MinCount**: `1`
  * **MaxCount**: `1` *(百式启动时选定本行动阶段锁定目标)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Tokens**: `[{Type: BattleQi, Amount: 3}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(横置进入百式形态)*:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Tapped`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `"fighter_hundred_dragon_form"`
    * `Ref`: `None`
  * **Effect[2]** *(百式期间禁用蓄力一击)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_fighter_hundred_dragon_disable_charged_strike"`
    * `RuleLifetimeRef`: `RuleLifeUntilTurnEnd`
    * `Ref`: `None`
  * **Effect[3]** *(在选定目标上挂“被百式锁定”标记)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelected`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_fighter_hundred_dragon_locked_target_marker"`
    * `RuleLifetimeRef`: `RuleLifeUntilTurnEnd`
    * `Ref`: `None`

#### 【百式幻龙拳·主动攻击增伤】 (Hundred Dragon Active Attack Buff)
* **说明**：处于百式形态时，你的主动攻击伤害额外+2。
* **1. 主干配置**：
  * **SkillID**: `skill_fighter_hundred_dragon_active_attack_buff`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnAttackDeclared`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDeclare]`
  * **CustomExpression**: `Self.Form == "fighter_hundred_dragon_form" && Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectAttackDamageModifier`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `2`
    * `Ref`: `None`

#### 【百式幻龙拳·应战攻击增伤】 (Hundred Dragon Counter Attack Buff)
* **说明**：处于百式形态时，你的应战攻击伤害额外+1。
* **1. 主干配置**：
  * **SkillID**: `skill_fighter_hundred_dragon_counter_attack_buff`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnAttackDeclared`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDeclare]`
  * **CustomExpression**: `Self.Form == "fighter_hundred_dragon_form" && Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == false`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectAttackDamageModifier`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `1`
    * `Ref`: `None`

#### 【百式幻龙拳·禁法术与特殊行动】 (Hundred Dragon Action Lock)
* **说明**：百式形态下，若尝试执行法术或特殊行动，则取消该行动并立即终止百式（转正）。
* **1. 主干配置**：
  * **SkillID**: `skill_fighter_hundred_dragon_forbid_magic_and_special`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingBeforeActionExecute`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `State.IsInSelfTurn(Self.UserID) == true && Self.Form == "fighter_hundred_dragon_form" && Action.SourceID == Self.UserID && (Action.CurrentType == Magic || Action.CurrentType == Buy || Action.CurrentType == Synthesize || Action.CurrentType == Extract || Action.CurrentType == Deadlock)`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **3.1 行动改写配置**：
  * **ActionTransform.Hook**: `TimingBeforeActionExecute`
  * **ActionTransform.Optional**: `false`
  * **ActionTransform.Priority**: `260`
  * **ActionTransform.CancelCurrentAction**: `true`
  * **ActionTransform.Match.RequireActionType**: `None`
  * **ActionTransform.Match.RequirePlayedCardTypes**: `[]`
  * **ActionTransform.Match.RequirePlayedCardElements**: `[]`
  * **ActionTransform.Match.ExcludeTemplateIDs**: `[]`
  * **ActionTransform.Rewrite**: `nil` *(命中时仅取消当前行动，不改写到其他流水线)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(转正)*:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Normal`
    * `Ref`: `None`
  * **Effect[1]** *(退出百式形态)*:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `nil`
    * `Ref`: `None`
  * **Effect[2]** *(移除“禁用蓄力一击”门禁规则)*:
    * `EffectType`: `EffectRemoveRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleRemoveRef`: `{Mode: RuleRemoveByModifierID, ModifierID: "rm_fighter_hundred_dragon_disable_charged_strike", Limit: 1}`
    * `Ref`: `None`
  * **Effect[3]** *(清空场上“被百式锁定目标”标记)*:
    * `EffectType`: `EffectRemoveRuleModifier`
    * `Target`: `TargetAllPlayers`
    * `Value`: `0`
    * `RuleRemoveRef`: `{Mode: RuleRemoveByModifierID, ModifierID: "rm_fighter_hundred_dragon_locked_target_marker", Limit: 0}`
    * `Ref`: `None`

#### 【百式幻龙拳·目标锁违例收束】 (Hundred Dragon Target Lock Violation)
* **说明**：百式形态下，若主动攻击目标不是启动时锁定的角色，则立刻终止百式（不取消本次攻击提交）。
* **1. 主干配置**：
  * **SkillID**: `skill_fighter_hundred_dragon_target_lock_violation`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingBeforeActionExecute`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `State.IsInSelfTurn(Self.UserID) == true && Self.Form == "fighter_hundred_dragon_form" && Action.SourceID == Self.UserID && Action.CurrentType == Attack && Combat.IsActiveAttack == true && Action.SelectedTargetCount >= 1 && State.HasRuleModifier(Action.Targets[0].TargetUserID, "rm_fighter_hundred_dragon_locked_target_marker") == false`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(转正)*:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Normal`
    * `Ref`: `None`
  * **Effect[1]** *(退出百式形态)*:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `nil`
    * `Ref`: `None`
  * **Effect[2]** *(移除“禁用蓄力一击”门禁规则)*:
    * `EffectType`: `EffectRemoveRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleRemoveRef`: `{Mode: RuleRemoveByModifierID, ModifierID: "rm_fighter_hundred_dragon_disable_charged_strike", Limit: 1}`
    * `Ref`: `None`
  * **Effect[3]** *(清空场上“被百式锁定目标”标记)*:
    * `EffectType`: `EffectRemoveRuleModifier`
    * `Target`: `TargetAllPlayers`
    * `Value`: `0`
    * `RuleRemoveRef`: `{Mode: RuleRemoveByModifierID, ModifierID: "rm_fighter_hundred_dragon_locked_target_marker", Limit: 0}`
    * `Ref`: `None`

#### 【百式幻龙拳·阶段结束】 (Hundred Dragon End)
* **说明**：在本行动阶段结束时，若仍处于百式形态，则结束该形态并转正。
* **1. 主干配置**：
  * **SkillID**: `skill_fighter_hundred_dragon_end`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnActionEnd`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionEnd]`
  * **CustomExpression**: `State.IsInSelfTurn(Self.UserID) == true && Self.Form == "fighter_hundred_dragon_form"`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Normal`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `nil`
    * `Ref`: `None`
  * **Effect[2]**:
    * `EffectType`: `EffectRemoveRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleRemoveRef`: `{Mode: RuleRemoveByModifierID, ModifierID: "rm_fighter_hundred_dragon_disable_charged_strike", Limit: 1}`
    * `Ref`: `None`
  * **Effect[3]**:
    * `EffectType`: `EffectRemoveRuleModifier`
    * `Target`: `TargetAllPlayers`
    * `Value`: `0`
    * `RuleRemoveRef`: `{Mode: RuleRemoveByModifierID, ModifierID: "rm_fighter_hundred_dragon_locked_target_marker", Limit: 0}`
    * `Ref`: `None`

#### 【气绝崩击】 (Qi Breaker)
* **技能描述**：（主动攻击前发动①，移除1点［斗气］）本次攻击对方无法应战，然后对自己造成X点法术伤害③，X为你的［斗气］数；不能和蓄力一击同时发动。
* **1. 主干配置**：
  * **SkillID**: `skill_fighter_qi_breaker`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnAttackDeclared`
* **1.2 响应分组配置**：
  * **ResponseGroup.GroupID**: `rg_fighter_attack_prefix_choice`
  * **ResponseGroup.Mode**: `ResponseGroupChooseOne`
  * **ResponseGroup.OptionOrder**: `2`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDeclare]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && Self.Tokens[BattleQi] >= 1`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Tokens**: `[{Type: BattleQi, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `Unrespondable`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelf`
    * `Value`: `Self.Tokens[BattleQi]`
    * `Ref`: `None`

#### 【斗神天驱】 (War God Drive)
* **技能描述**：［水晶］你弃到3张牌，+2［治疗］。
* **1. 主干配置**：
  * **SkillID**: `skill_fighter_war_god_drive`
  * **Category**: `Ultimate`
  * **Type**: `Startup`
  * **Timing**: `TimingStartup`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionStart]`
  * **CustomExpression**: `State.IsInSelfTurn(Self.UserID) == true`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Crystal, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(若当前手牌超过3，则弃到3张)*:
    * `EffectType`: `EffectDiscard`
    * `Target`: `TargetSelf`
    * `Value`: `Self.HandCount > 3 ? Self.HandCount - 3 : 0`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectHeal`
    * `Target`: `TargetSelf`
    * `Value`: `2`
    * `Ref`: `None`

---

### 31. 圣弓 (Holy Bow)

#### 【天之弓】 (Heavenly Bow)
* **技能描述**：游戏初始时，你+1［圣煌辉光炮］，+2[水晶]。你的［治疗］上限+1。（主动攻击时，若该次攻击不为圣类命格）本次攻击伤害-1；（主动攻击命中时，若该次攻击为圣类命格）你+1［信仰］。
* **0. 角色静态配置补充（CharacterTemplate）**：
  * **InitialFieldMarks**: `{HolyGloryCannon: 1}` *(游戏初始化时发放 1 张【圣煌辉光炮】场标资源)*
* **1. 主干配置**：
  * **SkillID**: `skill_holy_bow_heavenly_bow`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnTurnStart`
* **1.2 规则修饰器模板配置**：
  * **ModifierID**: `rm_holy_bow_heavenly_bow_max_heal_plus_1`
  * **Domain**: `RuleModifierDomainAttribute`
  * **Priority**: `100`
  * **ConditionExpression**: `None`
  * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
  * **AttrPayload**: `{AttrType: PlayerAttributeMaxHeal, Operation: AttributeModifyAdd, ValueSourceMode: RuleAttrValueFromFixed, Value: 1}`
* **2. 前置条件**：
  * **PhaseLimit**: `[GameInit]`
  * **CustomExpression**: `Event.SourceType == SourceSystem && Event.CauseAction == "GameStart"`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(你+2水晶)*:
    * `EffectType`: `EffectAddEnergyStone`
    * `Target`: `TargetSelf`
    * `Value`: `2`
    * `StoneRef`: `Crystal`
    * `Ref`: `None`
  * **Effect[1]** *(施加“治疗上限+1”常驻规则)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_holy_bow_heavenly_bow_max_heal_plus_1"`
    * `RuleLifetimeRef`: `RuleLifePermanent`
    * `Ref`: `None`

#### 【天之弓·非圣主动减伤】 (Heavenly Bow Non-Holy Penalty)
* **说明**：主动攻击时，若该次攻击不为圣命格，则本次攻击伤害-1。
* **1. 主干配置**：
  * **SkillID**: `skill_holy_bow_heavenly_bow_non_sheng_attack_penalty`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnAttackDeclared`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDeclare]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && Combat.AttackCard.Destiny != Sheng`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectAttackDamageModifier`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `-1`
    * `Ref`: `None`

#### 【天之弓·圣命中增信仰】 (Heavenly Bow Sheng Hit Gain)
* **说明**：主动攻击命中时，若该次攻击为圣命格，你+1信仰。
* **1. 主干配置**：
  * **SkillID**: `skill_holy_bow_heavenly_bow_sheng_hit_faith`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnHitCheck`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatHitCheck]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && Combat.IsHit == true && Combat.AttackCard.Destiny == Sheng`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `TokenRef`: `Faith`
    * `Ref`: `None`

#### 【圣屑飓暴】 (Holy Debris Hurricane)
* **技能描述**：（弃2张同系攻击牌［展示］）视为一次圣类命格的该系主动攻击。（若攻击未命中②，移除X点［治疗］，X最高为2）目标队友弃X张牌。
* **1. 主干配置**：
  * **SkillID**: `skill_holy_bow_holy_debris_hurricane`
  * **Category**: `Normal`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **1.2 规则修饰器模板配置**：
  * **ModifierID**: `rm_holy_bow_holy_debris_hurricane_armed`
  * **Domain**: `RuleModifierDomainCardSource`
  * **Priority**: `120`
  * **ConditionExpression**: `None`
  * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
  * **CardSourcePayload**: `{ProjectionMode: CardSourceProjectionAsHand, FieldMarks: []}` *(空载荷，仅作“本次攻击由圣屑飓暴改写”链路标记)*
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `State.ValidateUsedCardUUIDs(Action.UsedCardUUIDs, "CardType == Attack", 2) == true`
* **3. 目标选择规则**：
  * **SelectType**: `Enemy`
  * **MinCount**: `1`
  * **MaxCount**: `1`
* **3.1 行动改写配置**：
  * **ActionTransform.Hook**: `TimingBeforeActionExecute`
  * **ActionTransform.Optional**: `false`
  * **ActionTransform.Priority**: `120`
  * **ActionTransform.CancelCurrentAction**: `true`
  * **ActionTransform.Match.RequireActionType**: `None`
  * **ActionTransform.Match.RequirePlayedCardTypes**: `[]`
  * **ActionTransform.Match.RequirePlayedCardElements**: `[]`
  * **ActionTransform.Match.ExcludeTemplateIDs**: `[]`
  * **ActionTransform.Rewrite.FlowRef**: `ActionFlowNormalCombat`
  * **ActionTransform.Rewrite.ActionTypeRef**: `Attack`
  * **ActionTransform.Rewrite.ExecuteImmediately**: `true`
  * **ActionTransform.Rewrite.TreatAsActiveAttack**: `true`
  * **ActionTransform.Rewrite.ElementPickMode**: `RewriteElementFromActionRef`
  * **ActionTransform.Rewrite.FixedElementRef**: `None`
  * **SubmitAction.ElementRef**: `必填` *(本次“视为主动攻击”的系别)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Discards**:
    * `Count`: `2`
    * `Filter`: `{ReqCardType: Attack, SameAttribute: MatchElement}`
    * `Visibility`: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(挂载“圣屑飓暴改写攻击”标记，供未命中分支读取)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_holy_bow_holy_debris_hurricane_armed"`
    * `RuleLifetimeRef`: `RuleLifeUntilCombatEnd`
    * `Ref`: `None`

#### 【圣屑飓暴·未命中惩罚】 (Holy Debris Hurricane Miss Penalty)
* **说明**：若本次由【圣屑飓暴】改写的主动攻击未命中，则移除X治疗（X<=2），目标队友弃X张牌。
* **1. 主干配置**：
  * **SkillID**: `skill_holy_bow_holy_debris_hurricane_miss_penalty`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnHitCheck`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatHitCheck]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && Combat.IsHit == false && State.HasRuleModifier(Self.UserID, "rm_holy_bow_holy_debris_hurricane_armed") == true`
* **3. 目标选择规则**：
  * **SelectType**: `Teammate`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **SelectedValueRule**: `0 <= Action.SelectedValue && Action.SelectedValue <= Min(2, Self.Heal)`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **HealCost**: `Action.SelectedValue`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectDiscard`
    * `Target`: `TargetSelected`
    * `Value`: `Action.SelectedValue`
    * `Ref`: `None`

#### 【圣煌降临】 (Holy Glory Descent)
* **技能描述**：［持续］（移除你的2个［治疗］或2点［信仰］）［横置］，转为［圣煌形态］，额外+1［法术行动］。此形态下，你若执行［特殊行动］，则［转正］脱离［圣煌形态］并+1［治疗］。
* **1. 主干配置**：
  * **SkillID**: `skill_holy_bow_holy_glory_descent`
  * **Category**: `Normal`
  * **Type**: `Startup`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `State.IsInSelfTurn(Self.UserID) == true && ((Action.SelectedValue == 0 && Self.Heal >= 2) || (Action.SelectedValue == 1 && Self.Tokens[Faith] >= 2))`
* **3. 目标选择规则**：
  * **SelectType**: `None`
  * **SelectedValueRule**: `Action.SelectedValue == 0 || Action.SelectedValue == 1` *(0=移除2治疗；1=移除2信仰)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **HealCost**: `Action.SelectedValue == 0 ? 2 : 0`
  * **Tokens**: `[{Type: Faith, Amount: Action.SelectedValue == 1 ? 2 : 0}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Tapped`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `"holy_bow_holy_glory_form"`
    * `Ref`: `None`
  * **Effect[2]**:
    * `Implementation`: `model.AppendExtraAction`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `ActionRef`: `Magic`
    * `Ref`: `None`

#### 【圣煌降临·特殊行动退场】 (Holy Glory Descent Special Exit)
* **说明**：处于圣煌形态时，若你执行特殊行动，则转正脱离形态并+1治疗。
* **1. 主干配置**：
  * **SkillID**: `skill_holy_bow_holy_glory_descent_special_exit`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnActionEnd`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionEnd]`
  * **CustomExpression**: `State.IsInSelfTurn(Self.UserID) == true && Self.Form == "holy_bow_holy_glory_form" && Action.SourceID == Self.UserID && (Action.CurrentType == Buy || Action.CurrentType == Synthesize || Action.CurrentType == Extract || Action.CurrentType == Deadlock)`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Normal`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `nil`
    * `Ref`: `None`
  * **Effect[2]**:
    * `EffectType`: `EffectHeal`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `Ref`: `None`

#### 【圣光爆裂】 (Holy Light Burst)
* **技能描述**：（仅［圣煌形态］下可发动）你选择以下一项发动：●摸一张牌［强制］，移除你的1点［治疗］，你+1［信仰］，目标队友+1［治疗］。●（移除你的X［治疗］，选择最多X名手牌数不大于你手牌数-X的对手）你弃X张牌，然后对他们各造成（Y+2）点攻击伤害。Y为目标数中拥有［治疗］的人数。
* **1. 主干配置**：
  * **SkillID**: `skill_holy_bow_holy_light_burst`
  * **Category**: `Normal`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Self.Form == "holy_bow_holy_glory_form" && ((Action.SelectedValue == 0 && Action.SelectedTargetCount == 1) || (Action.SelectedValue >= 1 && Action.SelectedValue <= Self.Heal && Action.SelectedTargetCount >= 1 && Action.SelectedTargetCount <= Action.SelectedValue))`
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `1`
  * **MaxCount**: `4`
  * **SelectedValueRule**: `Action.SelectedValue == 0 || (Action.SelectedValue >= 1 && Action.SelectedValue <= Self.Heal)`
  * **SubmitAction.Targets 约束**: `当 Action.SelectedValue=0 时，目标必须为队友；当 Action.SelectedValue>=1 时，所有目标必须为敌方且每个目标满足 HandCount <= Self.HandCount - Action.SelectedValue。`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列（分支）**：
  * **BranchSelector**: `Action.SelectedValue == 0 ? 0 : 1`
  * **Branch[0]** *(选项A：摸1［强制］，移除自身1治疗，你+1信仰，目标队友+1治疗)*:
    * **Effect[A0]**:
      * `EffectType`: `EffectDrawCard`
      * `Target`: `TargetSelf`
      * `Value`: `1`
      * `Ref`: `None`
    * **Effect[A1]**:
      * `EffectType`: `EffectHeal`
      * `Target`: `TargetSelf`
      * `Value`: `-1`
      * `Ref`: `None`
    * **Effect[A2]**:
      * `EffectType`: `EffectAddToken`
      * `Target`: `TargetSelf`
      * `Value`: `1`
      * `TokenRef`: `Faith`
      * `Ref`: `None`
    * **Effect[A3]**:
      * `EffectType`: `EffectHeal`
      * `Target`: `TargetSelected`
      * `Value`: `1`
      * `Ref`: `None`
  * **Branch[1]** *(选项B：X=Action.SelectedValue，移除自身X治疗，弃X，然后对已选目标各造成(Y+2)攻击伤害)*:
    * **Effect[B0]**:
      * `EffectType`: `EffectHeal`
      * `Target`: `TargetSelf`
      * `Value`: `-Action.SelectedValue`
      * `Ref`: `None`
    * **Effect[B1]**:
      * `EffectType`: `EffectDiscard`
      * `Target`: `TargetSelf`
      * `Value`: `Action.SelectedValue`
      * `Ref`: `None`
    * **Effect[B2]**:
      * `EffectType`: `EffectAttackDamage`
      * `Target`: `TargetSelected`
      * `Value`: `2 + State.CountSelectedTargetsWithHealAtLeast(Action.Targets, 1)`
      * `Ref`: `None`

#### 【流星圣弹】 (Meteor Holy Bullet)
* **技能描述**：（仅［圣煌形态］下，主动攻击前①，移除你的1点［治疗］或1点［信仰］）我方目标角色+1［治疗］。
* **1. 主干配置**：
  * **SkillID**: `skill_holy_bow_meteor_holy_bullet`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnAttackDeclared`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDeclare]`
  * **CustomExpression**: `Self.Form == "holy_bow_holy_glory_form" && Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && ((Action.SelectedValue == 0 && Self.Heal >= 1) || (Action.SelectedValue == 1 && Self.Tokens[Faith] >= 1))`
* **3. 目标选择规则**：
  * **SelectType**: `Teammate`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **SelectedValueRule**: `Action.SelectedValue == 0 || Action.SelectedValue == 1` *(0=移除1治疗；1=移除1信仰)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **HealCost**: `Action.SelectedValue == 0 ? 1 : 0`
  * **Tokens**: `[{Type: Faith, Amount: Action.SelectedValue == 1 ? 1 : 0}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectHeal`
    * `Target`: `TargetSelected`
    * `Value`: `1`
    * `Ref`: `None`

#### 【圣煌辉光炮】 (Holy Glory Cannon)
* **技能描述**：（仅［圣煌形态］下可发动，移除1点［圣煌辉光炮］，移除4点［信仰］，并额外移除等同我方落后［士气］的［信仰］数）所有角色将手牌调整为4张，我方［星杯区］+1［星杯］，然后将一方［士气］调整与另一方相同。
* **1. 主干配置**：
  * **SkillID**: `skill_holy_bow_holy_glory_cannon`
  * **Category**: `Normal`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Self.Form == "holy_bow_holy_glory_form" && State.CountFieldMark(Self.UserID, HolyGloryCannon) >= 1 && Self.Tokens[Faith] >= 4 + ((Self.Team == Red && State.RedMorale < State.BlueMorale) ? (State.BlueMorale - State.RedMorale) : ((Self.Team == Blue && State.BlueMorale < State.RedMorale) ? (State.RedMorale - State.BlueMorale) : 0))`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(移除1点圣煌辉光炮场标)*:
    * `EffectType`: `EffectRemoveFieldMark`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `FieldMarkRef`: `HolyGloryCannon`
    * `Ref`: `None`
  * **Effect[1]** *(移除 4 + 落后士气差 的信仰)*:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `-(4 + ((Self.Team == Red && State.RedMorale < State.BlueMorale) ? (State.BlueMorale - State.RedMorale) : ((Self.Team == Blue && State.BlueMorale < State.RedMorale) ? (State.RedMorale - State.BlueMorale) : 0)))`
    * `TokenRef`: `Faith`
    * `Ref`: `None`
  * **Effect[2]** *(所有角色手牌调整为4)*:
    * `EffectType`: `EffectAdjustHand`
    * `Target`: `TargetAllPlayers`
    * `Value`: `4`
    * `Ref`: `None`
  * **Effect[3]** *(我方星杯区+1)*:
    * `EffectType`: `EffectAddTeamCup`
    * `Target`: `TargetSelfTeam`
    * `Value`: `1`
    * `Ref`: `None`
  * **Effect[4]** *(将较低士气一方调整为较高士气一方)*:
    * `EffectType`: `EffectSwapMorale`
    * `Target`: `TargetNone`
    * `Value`: `0`
    * `Ref`: `None`

#### 【自动填充】 (Auto Refill)
* **技能描述**：（你的回合结束时，若你未执行［特殊行动］）你选择以下一项发动：●［水晶］你+1［信仰］或+1［治疗］。●［宝石］你+1水晶，+2［信仰］或+2［治疗］。
* **1. 主干配置**：
  * **SkillID**: `skill_holy_bow_auto_refill`
  * **Category**: `Ultimate`
  * **Type**: `Response`
  * **Timing**: `TimingOnTurnEnd`
* **2. 前置条件**：
  * **PhaseLimit**: `[TurnEnd]`
  * **CustomExpression**: `State.IsInSelfTurn(Self.UserID) == true && State.HasExecutedSpecialActionThisTurn(Self.UserID) == false && (Action.SelectedValue == 0 || Action.SelectedValue == 1 || Action.SelectedValue == 2 || Action.SelectedValue == 3)`
* **3. 目标选择规则**：
  * **SelectType**: `None`
  * **SelectedValueRule**: `Action.SelectedValue == 0 || Action.SelectedValue == 1 || Action.SelectedValue == 2 || Action.SelectedValue == 3` *(0=水晶+1信仰；1=水晶+1治疗；2=宝石+1水晶+2信仰；3=宝石+1水晶+2治疗)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Crystal, Amount: (Action.SelectedValue == 0 || Action.SelectedValue == 1) ? 1 : 0}, {Type: Gem, Amount: (Action.SelectedValue == 2 || Action.SelectedValue == 3) ? 1 : 0}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列（分支）**：
  * **BranchSelector**: `Action.SelectedValue`
  * **Branch[0]** *(支付1水晶，你+1信仰)*:
    * **Effect[B0]**:
      * `EffectType`: `EffectAddToken`
      * `Target`: `TargetSelf`
      * `Value`: `1`
      * `TokenRef`: `Faith`
      * `Ref`: `None`
  * **Branch[1]** *(支付1水晶，你+1治疗)*:
    * **Effect[B1]**:
      * `EffectType`: `EffectHeal`
      * `Target`: `TargetSelf`
      * `Value`: `1`
      * `Ref`: `None`
  * **Branch[2]** *(支付1宝石，你+1水晶并+2信仰)*:
    * **Effect[B2-0]**:
      * `EffectType`: `EffectAddEnergyStone`
      * `Target`: `TargetSelf`
      * `Value`: `1`
      * `StoneRef`: `Crystal`
      * `Ref`: `None`
    * **Effect[B2-1]**:
      * `EffectType`: `EffectAddToken`
      * `Target`: `TargetSelf`
      * `Value`: `2`
      * `TokenRef`: `Faith`
      * `Ref`: `None`
  * **Branch[3]** *(支付1宝石，你+1水晶并+2治疗)*:
    * **Effect[B3-0]**:
      * `EffectType`: `EffectAddEnergyStone`
      * `Target`: `TargetSelf`
      * `Value`: `1`
      * `StoneRef`: `Crystal`
      * `Ref`: `None`
    * **Effect[B3-1]**:
      * `EffectType`: `EffectHeal`
      * `Target`: `TargetSelf`
      * `Value`: `2`
      * `Ref`: `None`

---

### 32. 剑帝 (Sword Emperor)

#### 【剑魂守护】 (Sword Soul Guard)
* **技能描述**：（主动攻击未命中时发动②）将本次打出的攻击牌作为面朝下放置在你的角色旁，作为［剑魂］。若你现有能量为单数，你的所有［剑魂］视为［天使之魂］；若为双数，视为［恶魔之魂］；若没有能量，则不属于任何一种。（若剑魂达到上限）你不能发动［剑魂守护］。
* **1. 主干配置**：
  * **SkillID**: `skill_sword_emperor_sword_soul_guard`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnHitCheck`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatHitCheck]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && Combat.IsHit == false && State.IsSkillDisabled(Self.UserID, "skill_sword_emperor_sword_soul_guard") == false`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(将本次打出的攻击牌实体以暗置方式留场为剑魂)*:
    * `EffectType`: `EffectPlacePlayedCardAsFieldMark`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `FieldMarkRef`: `SwordSoul`
    * `VisibilityRef`: `VisibilityHidden`
    * `Ref`: `None`

#### 【佯攻】 (Feint)
* **技能描述**：（主动攻击未命中时发动②）你+1［剑气］。
* **1. 主干配置**：
  * **SkillID**: `skill_sword_emperor_feint`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnHitCheck`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatHitCheck]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && Combat.IsHit == false`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `TokenRef`: `SwordQi`
    * `Ref`: `None`

#### 【剑气斩】 (Sword Qi Slash)
* **技能描述**：（主动攻击命中后发动②，移除X点［剑气］，X最高为3）对除你所攻击的目标以外的任意一名角色造成X点法术伤害③。
* **1. 主干配置**：
  * **SkillID**: `skill_sword_emperor_sword_qi_slash`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnHitCheck`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatHitCheck]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && Combat.IsHit == true && Self.Tokens[SwordQi] >= 1 && Action.Targets[0].TargetUserID != Combat.TargetID`
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **SelectedValueRule**: `1 <= Action.SelectedValue && Action.SelectedValue <= Min(3, Self.Tokens[SwordQi])`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Tokens**: `[{Type: SwordQi, Amount: Action.SelectedValue}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelected`
    * `Value`: `Action.SelectedValue`
    * `Ref`: `None`

#### 【天使之魂】 (Angel Soul)
* **技能描述**：（主动攻击前发动①，移除1张天使之魂）本次攻击若命中②，你+2［治疗］；若未命中②，我方+1士气；不能和剑魂守护同时发动。
* **1. 主干配置**：
  * **SkillID**: `skill_sword_emperor_angel_soul`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnAttackDeclared`
* **1.2 规则修饰器模板配置**：
  * **ModifierID**: `rm_sword_emperor_disable_guard_current_combat_from_angel`
  * **Domain**: `RuleModifierDomainSkillGate`
  * **Priority**: `260`
  * **ConditionExpression**: `None`
  * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
  * **SkillGatePayload**: `{Mode: SkillGateDisallowList, SkillIDs: ["skill_sword_emperor_sword_soul_guard"]}`
* **1.3 规则修饰器模板配置**：
  * **ModifierID**: `rm_sword_emperor_angel_soul_armed`
  * **Domain**: `RuleModifierDomainCardSource`
  * **Priority**: `260`
  * **ConditionExpression**: `None`
  * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
  * **CardSourcePayload**: `{ProjectionMode: CardSourceProjectionAsHand, FieldMarks: []}` *(空载荷，仅作“天使之魂已挂载”的本战斗链路标记)*
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDeclare]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && Self.EnergyCount > 0 && Self.EnergyCount % 2 == 1 && State.CountFieldMark(Self.UserID, SwordSoul) >= 1`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(攻击前：移除1张剑魂（当前按奇数能量视为天使之魂）)*:
    * `EffectType`: `EffectRemoveFieldMark`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `FieldMarkRef`: `SwordSoul`
    * `Ref`: `None`
  * **Effect[1]** *(攻击前：本次战斗禁用【剑魂守护】)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_sword_emperor_disable_guard_current_combat_from_angel"`
    * `RuleLifetimeRef`: `RuleLifeUntilCombatEnd`
    * `Ref`: `None`
  * **Effect[2]** *(攻击前：挂载天使之魂战斗标记)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_sword_emperor_angel_soul_armed"`
    * `RuleLifetimeRef`: `RuleLifeUntilCombatEnd`
    * `Ref`: `None`

#### 【恶魔之魂】 (Demon Soul)
* **技能描述**：（主动攻击前发动①，移除1张恶魔之魂）本次攻击伤害额外+1；若未命中②，你+2［剑气］；不能和剑魂守护同时发动。
* **1. 主干配置**：
  * **SkillID**: `skill_sword_emperor_demon_soul`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnAttackDeclared`
* **1.2 规则修饰器模板配置**：
  * **ModifierID**: `rm_sword_emperor_disable_guard_current_combat_from_demon`
  * **Domain**: `RuleModifierDomainSkillGate`
  * **Priority**: `260`
  * **ConditionExpression**: `None`
  * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
  * **SkillGatePayload**: `{Mode: SkillGateDisallowList, SkillIDs: ["skill_sword_emperor_sword_soul_guard"]}`
* **1.3 规则修饰器模板配置**：
  * **ModifierID**: `rm_sword_emperor_demon_soul_armed`
  * **Domain**: `RuleModifierDomainCardSource`
  * **Priority**: `260`
  * **ConditionExpression**: `None`
  * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
  * **CardSourcePayload**: `{ProjectionMode: CardSourceProjectionAsHand, FieldMarks: []}` *(空载荷，仅作“恶魔之魂已挂载”的本战斗链路标记)*
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDeclare]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && Self.EnergyCount > 0 && Self.EnergyCount % 2 == 0 && State.CountFieldMark(Self.UserID, SwordSoul) >= 1`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(攻击前：移除1张剑魂（当前按偶数能量视为恶魔之魂）)*:
    * `EffectType`: `EffectRemoveFieldMark`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `FieldMarkRef`: `SwordSoul`
    * `Ref`: `None`
  * **Effect[1]** *(攻击前：本次攻击伤害额外+1)*:
    * `EffectType`: `EffectAttackDamageModifier`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `1`
    * `Ref`: `None`
  * **Effect[2]** *(攻击前：本次战斗禁用【剑魂守护】)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_sword_emperor_disable_guard_current_combat_from_demon"`
    * `RuleLifetimeRef`: `RuleLifeUntilCombatEnd`
    * `Ref`: `None`
  * **Effect[3]** *(攻击前：挂载恶魔之魂战斗标记)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_sword_emperor_demon_soul_armed"`
    * `RuleLifetimeRef`: `RuleLifeUntilCombatEnd`
    * `Ref`: `None`

#### 【天使之魂·命中结算】 (Angel Soul Hit Resolve)
* **说明**：若本次攻击由【天使之魂】挂载且命中，则你+2治疗。
* **1. 主干配置**：
  * **SkillID**: `skill_sword_emperor_angel_soul_hit_resolve`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnHitCheck`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatHitCheck]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && Combat.IsHit == true && State.HasRuleModifier(Self.UserID, "rm_sword_emperor_angel_soul_armed") == true`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectHeal`
    * `Target`: `TargetSelf`
    * `Value`: `2`
    * `Ref`: `None`

#### 【天使之魂·未命中结算】 (Angel Soul Miss Resolve)
* **说明**：若本次攻击由【天使之魂】挂载且未命中，则我方+1士气。
* **1. 主干配置**：
  * **SkillID**: `skill_sword_emperor_angel_soul_miss_resolve`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnHitCheck`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatHitCheck]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && Combat.IsHit == false && State.HasRuleModifier(Self.UserID, "rm_sword_emperor_angel_soul_armed") == true`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectChangeMorale`
    * `Target`: `TargetSelfTeam`
    * `Value`: `1`
    * `Ref`: `None`

#### 【恶魔之魂·未命中结算】 (Demon Soul Miss Resolve)
* **说明**：若本次攻击由【恶魔之魂】挂载且未命中，则你+2剑气。
* **1. 主干配置**：
  * **SkillID**: `skill_sword_emperor_demon_soul_miss_resolve`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnHitCheck`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatHitCheck]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && Combat.IsHit == false && State.HasRuleModifier(Self.UserID, "rm_sword_emperor_demon_soul_armed") == true`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `2`
    * `TokenRef`: `SwordQi`
    * `Ref`: `None`

#### 【不屈意志】 (Indomitable Will)
* **技能描述**：［水晶］（［攻击行动］结束时发动）你摸1张［强制］，+1［剑气］，额外+1［攻击行动］。
* **1. 主干配置**：
  * **SkillID**: `skill_sword_emperor_indomitable_will`
  * **Category**: `Ultimate`
  * **Type**: `Response`
  * **Timing**: `TimingOnActionEnd`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionEnd]`
  * **CustomExpression**: `State.IsInSelfTurn(Self.UserID) == true && Action.SourceID == Self.UserID && Action.CurrentType == Attack`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Crystal, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectDrawCard`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `TokenRef`: `SwordQi`
    * `Ref`: `None`
  * **Effect[2]**:
    * `Implementation`: `model.AppendExtraAction`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `ActionRef`: `Attack`
    * `Ref`: `None`

---

### 33. 兽灵武士 (Beast Spirit Samurai)

#### 【武者残心】 (Warrior Zanshin)
* **技能描述**：［回合限定］（［攻击行动］结束时）你+1［残心］。
* **1. 主干配置**：
  * **SkillID**: `skill_beast_samurai_warrior_zanshin`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnActionEnd`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionEnd]`
  * **IsTurnLimited**: `true`
  * **CustomExpression**: `State.IsInSelfTurn(Self.UserID) == true && Action.SourceID == Self.UserID && Action.CurrentType == Attack`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `TokenRef`: `Zanshin`
    * `Ref`: `None`

#### 【一击无念】 (One Strike No Thought)
* **技能描述**：（［攻击行动］结束后，移除4点［残心］）额外+1［攻击行动］，本次攻击无视圣盾且无法用［圣光］抵挡。（若攻击牌为技类命格）本次攻击强制命中。
* **1. 主干配置**：
  * **SkillID**: `skill_beast_samurai_one_strike_no_thought`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnActionEnd`
* **1.2 规则修饰器模板配置**：
  * **ModifierID**: `rm_beast_samurai_one_strike_no_thought_armed`
  * **Domain**: `RuleModifierDomainCardSource`
  * **Priority**: `250`
  * **ConditionExpression**: `None`
  * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
  * **CardSourcePayload**: `{ProjectionMode: CardSourceProjectionAsHand, FieldMarks: []}` *(空载荷，仅作“下次主动攻击套用一击无念劫持”标记)*
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionEnd]`
  * **CustomExpression**: `State.IsInSelfTurn(Self.UserID) == true && Action.SourceID == Self.UserID && Action.CurrentType == Attack && Self.Tokens[Zanshin] >= 4`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Tokens**: `[{Type: Zanshin, Amount: 4}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(额外+1攻击行动)*:
    * `Implementation`: `model.AppendExtraAction`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `ActionRef`: `Attack`
    * `Ref`: `None`
  * **Effect[1]** *(挂载“下次主动攻击劫持”标记，持续到本回合结束或被消耗)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_beast_samurai_one_strike_no_thought_armed"`
    * `RuleLifetimeRef`: `RuleLifeUntilTurnEnd`
    * `Ref`: `None`

#### 【一击无念·下次攻击劫持】 (One Strike No Thought Next-Attack Intercept)
* **说明**：当【一击无念】已挂载时，你的下一次主动攻击获得“无视圣盾+无视目标圣光”；若攻击牌命格为技，则强制命中。结算后移除该挂载标记。
* **1. 主干配置**：
  * **SkillID**: `skill_beast_samurai_one_strike_no_thought_next_attack_intercept`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnAttackDeclared`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDeclare]`
  * **CustomExpression**: `State.IsInSelfTurn(Self.UserID) == true && Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && State.HasRuleModifier(Self.UserID, "rm_beast_samurai_one_strike_no_thought_armed") == true`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectApplyCombatTag`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `IgnoreHolyShield`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectApplyCombatTag`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `IgnoreTargetHoly`
    * `Ref`: `None`
  * **Effect[2]** *(若攻击牌为技类命格，则强制命中)*:
    * `EffectType`: `EffectApplyCombatTag`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `ForceHit`
    * `Condition`: `Combat.AttackCard.Destiny == Ji`
    * `Ref`: `None`
  * **Effect[3]** *(消耗挂载标记，确保仅作用下一次主动攻击)*:
    * `EffectType`: `EffectRemoveRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleRemoveRef`: `{Mode: RuleRemoveByModifierID, ModifierID: "rm_beast_samurai_one_strike_no_thought_armed", Limit: 1}`
    * `Ref`: `None`

#### 【兽魂意念】 (Beast Soul Will)
* **技能描述**：（你每移除1点［兽魂］）你+1［残心］；（仅［普通形态］下，主动攻击命中时②）你+1［兽魂］。
* **实现约定**：第一条“每移除1点兽魂→+1残心”已在本角色所有“移除兽魂”的技能效果中内联落地（例如【兽魂警戒】【兽返】【逆反居合斩】与形态回合结算），不依赖未定义的 TokenChange 触发钩子。
* **1. 主干配置**：
  * **SkillID**: `skill_beast_samurai_beast_soul_will`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnHitCheck`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatHitCheck]`
  * **CustomExpression**: `Self.Form == nil && Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && Combat.IsHit == true`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `TokenRef`: `BeastSoul`
    * `Ref`: `None`

#### 【兽魂警戒】 (Beast Soul Alert)
* **技能描述**：［持续］（其他角色的［横置］效果结算完成后，移除1点［兽魂］，角色［横置］转为［御魂流居合形态］）目标角色弃1张牌［展示］；（若弃牌为法术牌）你+1［兽魂］。
* **1. 主干配置**：
  * **SkillID**: `skill_beast_samurai_beast_soul_alert`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnOrientationChanged`
* **2. 前置条件**：
  * **PhaseLimit**: `None`
  * **CustomExpression**: `Event.OperatorID != Self.UserID && Self.Tokens[BeastSoul] >= 1`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Tokens**: `[{Type: BeastSoul, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(移除兽魂触发兽魂意念：+1残心)*:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `TokenRef`: `Zanshin`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Tapped`
    * `Ref`: `None`
  * **Effect[2]**:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `"beast_samurai_iaijutsu_form"`
    * `Ref`: `None`
  * **Effect[3]** *(令触发该横置变化的角色弃1张牌)*:
    * `EffectType`: `EffectDiscard`
    * `Target`: `TargetTriggerSource`
    * `Value`: `1`
    * `Ref`: `None`
  * **Effect[4]** *(若其弃牌为法术牌，则你+1兽魂)*:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `TokenRef`: `BeastSoul`
    * `Condition`: `Event.DiscardedMagicCount >= 1`
    * `Ref`: `None`

#### 【兽返】 (Beast Return)
* **技能描述**：（目标角色对你造成法术伤害③时，移除X点［兽魂］）你弃X张牌，他弃1张牌；（若他的弃牌为法术牌）你+1［兽魂］。
* **1. 主干配置**：
  * **SkillID**: `skill_beast_samurai_beast_return`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnDamageCalculated`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDamage]`
  * **CustomExpression**: `Combat.TargetID == Self.UserID && Combat.IsMagic == true && Combat.SourceID != Self.UserID && Action.SelectedValue >= 0 && Action.SelectedValue <= Self.Tokens[BeastSoul]`
* **3. 目标选择规则**：
  * **SelectType**: `None`
  * **SelectedValueRule**: `0 <= Action.SelectedValue && Action.SelectedValue <= Self.Tokens[BeastSoul]`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Tokens**: `[{Type: BeastSoul, Amount: Action.SelectedValue}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(移除兽魂触发兽魂意念：+X残心)*:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `Action.SelectedValue`
    * `TokenRef`: `Zanshin`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectDiscard`
    * `Target`: `TargetSelf`
    * `Value`: `Action.SelectedValue`
    * `Ref`: `None`
  * **Effect[2]**:
    * `EffectType`: `EffectDiscard`
    * `Target`: `TargetTriggerSource`
    * `Value`: `1`
    * `Ref`: `None`
  * **Effect[3]** *(若对方弃牌为法术牌，则你+1兽魂)*:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `TokenRef`: `BeastSoul`
    * `Condition`: `Event.DiscardedMagicCount >= 1`
    * `Ref`: `None`

#### 【御魂流居合形态·回合结束扣魂】 (Iaijutsu Form Turn-End BeastSoul Drain)
* **说明**：处于御魂流居合形态时，你回合结束前-1兽魂；该次移除同步触发兽魂意念，+1残心。
* **1. 主干配置**：
  * **SkillID**: `skill_beast_samurai_iaijutsu_form_turn_end_drain`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnTurnEnd`
* **2. 前置条件**：
  * **PhaseLimit**: `[TurnEnd]`
  * **CustomExpression**: `State.IsInSelfTurn(Self.UserID) == true && Self.Form == "beast_samurai_iaijutsu_form" && Self.Tokens[BeastSoul] >= 1`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `-1`
    * `TokenRef`: `BeastSoul`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `TokenRef`: `Zanshin`
    * `Ref`: `None`

#### 【御魂流居合形态·造成伤害退场】 (Iaijutsu Form Exit On Deal Damage)
* **说明**：处于御魂流居合形态时，只要你造成过伤害（⑥窗口），立即转正并脱离该形态。
* **1. 主干配置**：
  * **SkillID**: `skill_beast_samurai_iaijutsu_form_exit_on_deal_damage`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnDamageTaken`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDraw]`
  * **CustomExpression**: `Self.Form == "beast_samurai_iaijutsu_form" && Combat.SourceID == Self.UserID && Combat.FinalDamage > 0`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Normal`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `nil`
    * `Ref`: `None`

#### 【御魂流居合形态·兽魂归零退场】 (Iaijutsu Form Exit On BeastSoul Zero)
* **说明**：你的回合结束时，若仍处于御魂流居合形态且兽魂为0，则转正并脱离该形态。
* **1. 主干配置**：
  * **SkillID**: `skill_beast_samurai_iaijutsu_form_exit_on_zero_beastsoul`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnTurnEnd`
* **2. 前置条件**：
  * **PhaseLimit**: `[TurnEnd]`
  * **CustomExpression**: `State.IsInSelfTurn(Self.UserID) == true && Self.Form == "beast_samurai_iaijutsu_form" && Self.Tokens[BeastSoul] == 0`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Normal`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `nil`
    * `Ref`: `None`

#### 【御魂流居合形态·横置目标增伤】 (Iaijutsu Form Tapped-Target Damage Boost)
* **说明**：处于御魂流居合形态时，你对横置目标角色的主动攻击伤害+1。
* **1. 主干配置**：
  * **SkillID**: `skill_beast_samurai_iaijutsu_form_tapped_target_boost`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnAttackDeclared`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDeclare]`
  * **CustomExpression**: `Self.Form == "beast_samurai_iaijutsu_form" && Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && State.GetPlayerOrientation(Combat.TargetID) == Tapped`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectAttackDamageModifier`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `1`
    * `Ref`: `None`

#### 【逆反居合斩】 (Reversal Iaijutsu Slash)
* **技能描述**：（仅［御魂流居合形态］下，攻击手牌<4的对手前①发动）移除X点［兽魂］。本次攻击命中时②，改为攻击目标弃置（X+2）张手牌。（若因此弃牌数小于X+2）对方士气-1。
* **实现约定**：为保证“命中后改为弃牌”可在同链路强类型落地，采用 `TimingOnHitCheck` 响应并在命中窗口提交 X 值，随后对当前攻击目标执行“伤害改写为0 + 弃牌替代”。
* **1. 主干配置**：
  * **SkillID**: `skill_beast_samurai_reversal_iaijutsu_slash`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnHitCheck`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatHitCheck]`
  * **CustomExpression**: `Self.Form == "beast_samurai_iaijutsu_form" && Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && Combat.IsHit == true && Combat.TargetHandCount < 4 && Action.SelectedValue >= 0 && Action.SelectedValue <= Self.Tokens[BeastSoul] && Action.Targets[0].TargetUserID == Combat.TargetID`
* **3. 目标选择规则**：
  * **SelectType**: `Enemy`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **SelectedValueRule**: `0 <= Action.SelectedValue && Action.SelectedValue <= Self.Tokens[BeastSoul]`
  * **Filters**: `Target.UserID == Combat.TargetID`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Tokens**: `[{Type: BeastSoul, Amount: Action.SelectedValue}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(移除兽魂触发兽魂意念：+X残心)*:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `Action.SelectedValue`
    * `TokenRef`: `Zanshin`
    * `Ref`: `None`
  * **Effect[1]** *(将本次攻击待生效伤害归零，改为弃牌结算)*:
    * `EffectType`: `EffectModifyPendingDamage`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `-Combat.FinalDamage`
    * `Ref`: `None`
  * **Effect[2]** *(攻击目标弃置 X+2 张手牌)*:
    * `EffectType`: `EffectDiscard`
    * `Target`: `TargetSelected`
    * `Value`: `Action.SelectedValue + 2`
    * `Ref`: `None`
  * **Effect[3]** *(若其发动前手牌数已不足 X+2，则视为“弃牌数小于X+2”，敌方士气-1)*:
    * `EffectType`: `EffectChangeMorale`
    * `Target`: `TargetEnemyTeam`
    * `Value`: `-1`
    * `Condition`: `Combat.TargetHandCount < Action.SelectedValue + 2`
    * `Ref`: `None`

#### 【御魂流居合式】 (Iaijutsu Style)
* **技能描述**：［持续］［宝石］无视你的［兽魂］上限+1［兽魂］，你可选择摸或弃1张牌；（若你处于［御魂流居合形态］）你+1［残心］；（若你处于［普通型态］）［横置］转为［御魂流居合形态］。
* **1. 主干配置**：
  * **SkillID**: `skill_beast_samurai_iaijutsu_style`
  * **Category**: `Ultimate`
  * **Type**: `Startup`
  * **Timing**: `TimingStartup`
* **1.2 规则修饰器模板配置**：
  * **ModifierID**: `rm_beast_samurai_iaijutsu_style_ignore_beastsoul_cap`
  * **Domain**: `RuleModifierDomainTokenPolicy`
  * **Priority**: `220`
  * **ConditionExpression**: `None`
  * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
  * **TokenPolicyPayload**: `{TokenType: BeastSoul, ApplyMode: TokenApplyIgnoreMax}`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionStart]`
  * **CustomExpression**: `Action.SelectedValue == 0 || Action.SelectedValue == 1`
* **3. 目标选择规则**：
  * **SelectType**: `None`
  * **SelectedValueRule**: `Action.SelectedValue == 0 || Action.SelectedValue == 1` *(0=摸1；1=弃1)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Gem, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(在本技能执行链内临时无视兽魂上限)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_beast_samurai_iaijutsu_style_ignore_beastsoul_cap"`
    * `RuleLifetimeRef`: `RuleLifeThisEffectChain`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `TokenRef`: `BeastSoul`
    * `Ref`: `None`
  * **Effect[2]** *(选择摸1)*:
    * `EffectType`: `EffectDrawCard`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `Condition`: `Action.SelectedValue == 0`
    * `Ref`: `None`
  * **Effect[3]** *(选择弃1)*:
    * `EffectType`: `EffectDiscard`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `Condition`: `Action.SelectedValue == 1`
    * `Ref`: `None`
  * **Effect[4]** *(若当前在御魂流居合形态，则+1残心)*:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `TokenRef`: `Zanshin`
    * `Condition`: `Self.Form == "beast_samurai_iaijutsu_form"`
    * `Ref`: `None`
  * **Effect[5]** *(若当前为普通形态，则横置入形态)*:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Tapped`
    * `Condition`: `Self.Form == nil`
    * `Ref`: `None`
  * **Effect[6]** *(若当前为普通形态，则命名进入御魂流居合形态)*:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `"beast_samurai_iaijutsu_form"`
    * `Condition`: `Self.Form == nil`
    * `Ref`: `None`

---

### 34. 灵魂术士 (Soul Sorcerer)

#### 【灵魂吞噬】 (Soul Devour)
* **技能描述**：（我方每有1点士气下降）你+1［黄色灵魂］。
* **1. 主干配置**：
  * **SkillID**: `skill_soul_sorcerer_soul_devour`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnDamageTaken`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDraw]`
  * **CustomExpression**: `Event.MoraleDropApplied > 0 && State.IsSameTeam(Self.UserID, Combat.TargetID) == true`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `Event.MoraleDropApplied`
    * `TokenRef`: `SoulYellow`
    * `Ref`: `None`

#### 【灵魂召还】 (Soul Recall)
* **技能描述**：（弃X张法术牌［展示］）你+X点［蓝色灵魂］。
* **1. 主干配置**：
  * **SkillID**: `skill_soul_sorcerer_soul_recall`
  * **Category**: `Normal`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Action.NamedValues["X"] >= 1 && State.ValidateUsedCardUUIDs(Action.UsedCardUUIDs, "CardType == Magic", Action.NamedValues["X"]) == true`
* **3. 目标选择规则**：
  * **SelectType**: `None`
  * **NamedValueConstraints**:
    * `{Key: "X", Required: true, MinExpression: "1", MaxExpression: "Self.HandCount"}`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Discards**:
    * `Count`: `Action.NamedValues["X"]`
    * `Filter`: `{ReqCardType: Magic}`
    * `Visibility`: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `Action.NamedValues["X"]`
    * `TokenRef`: `SoulBlue`
    * `Ref`: `None`

#### 【灵魂转换】 (Soul Conversion)
* **技能描述**：（你每发动1次主动攻击①）可转换1点你拥有的［灵魂］的颜色。
* **1. 主干配置**：
  * **SkillID**: `skill_soul_sorcerer_soul_conversion`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnAttackDeclared`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDeclare]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && ((Action.SelectedValue == 0 && Self.Tokens[SoulYellow] >= 1) || (Action.SelectedValue == 1 && Self.Tokens[SoulBlue] >= 1))`
* **3. 目标选择规则**：
  * **SelectType**: `None`
  * **SelectedValueRule**: `Action.SelectedValue == 0 || Action.SelectedValue == 1` *(0=黄转蓝；1=蓝转黄)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Tokens**: `[{Type: SoulYellow, Amount: Action.SelectedValue == 0 ? 1 : 0}, {Type: SoulBlue, Amount: Action.SelectedValue == 1 ? 1 : 0}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列（分支）**：
  * **BranchSelector**: `Action.SelectedValue`
  * **Branch[0]** *(黄转蓝)*:
    * **Effect[B0]**:
      * `EffectType`: `EffectAddToken`
      * `Target`: `TargetSelf`
      * `Value`: `1`
      * `TokenRef`: `SoulBlue`
      * `Ref`: `None`
  * **Branch[1]** *(蓝转黄)*:
    * **Effect[B1]**:
      * `EffectType`: `EffectAddToken`
      * `Target`: `TargetSelf`
      * `Value`: `1`
      * `TokenRef`: `SoulYellow`
      * `Ref`: `None`

#### 【灵魂镜像】 (Soul Mirror)
* **技能描述**：（移除2点［黄色灵魂］）你弃2张牌，目标角色摸2张牌［强制］，但最多补到其手牌上限。
* **1. 主干配置**：
  * **SkillID**: `skill_soul_sorcerer_soul_mirror`
  * **Category**: `Normal`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Self.Tokens[SoulYellow] >= 2`
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `1`
  * **MaxCount**: `1`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Tokens**: `[{Type: SoulYellow, Amount: 2}]`
  * **Discards**:
    * `Count`: `2`
    * `Filter`: `{}`
    * `Visibility`: `VisibilityHidden`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectDrawCard`
    * `Target`: `TargetSelected`
    * `Value`: `Min(2, Max(0, Target.HandLimit - Target.HandCount))`
    * `Ref`: `None`

#### 【灵魂震爆】 (Soul Blast)
* **技能描述**：（移除3点［黄色灵魂］）对目标角色造成3点法术伤害③，若他手牌<3且手牌上限>5，则本次伤害额外+2。
* **1. 主干配置**：
  * **SkillID**: `skill_soul_sorcerer_soul_blast`
  * **Category**: `Unique`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Self.Tokens[SoulYellow] >= 3 && Action.PlayedCard.CharacterSkillMap[Self.CharacterID] == "skill_soul_sorcerer_soul_blast"`
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `1`
  * **MaxCount**: `1`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayRequired`
  * **Tokens**: `[{Type: SoulYellow, Amount: 3}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelected`
    * `Value`: `3`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelected`
    * `Value`: `2`
    * `Condition`: `Target.HandCount < 3 && Target.HandLimit > 5`
    * `Ref`: `None`

#### 【灵魂赐予】 (Soul Gift)
* **技能描述**：（移除3点［蓝色灵魂］）目标角色+2［宝石］。
* **1. 主干配置**：
  * **SkillID**: `skill_soul_sorcerer_soul_gift`
  * **Category**: `Unique`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Self.Tokens[SoulBlue] >= 3 && Action.PlayedCard.CharacterSkillMap[Self.CharacterID] == "skill_soul_sorcerer_soul_gift"`
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `1`
  * **MaxCount**: `1`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayRequired`
  * **Tokens**: `[{Type: SoulBlue, Amount: 3}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectAddEnergyStone`
    * `Target`: `TargetSelected`
    * `Value`: `2`
    * `StoneRef`: `Gem`
    * `Ref`: `None`

#### 【灵魂链接】 (Soul Link)
* **技能描述**：（仅你队友数>1时可发动，移除1点【黄色灵魂】和1点【蓝色灵魂】）将此卡放置于目标队友面前。（若你拥有【灵魂链结】，每当你或灵魂术士将承受伤害前⑥，灵魂术士移除X点【蓝色灵魂】）将X点伤害转移给另一方，转移后的伤害为法术伤害⑥。
* **1. 主干配置**：
  * **SkillID**: `skill_soul_sorcerer_soul_link`
  * **Category**: `Exclusive`
  * **Type**: `Startup`
  * **Timing**: `TimingStartup`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionStart]`
  * **CustomExpression**: `State.GetAliveTeammateCount(Self.UserID) > 1 && Self.Tokens[SoulYellow] >= 1 && Self.Tokens[SoulBlue] >= 1 && Action.PlayedCard.CharacterSkillMap[Self.CharacterID] == "skill_soul_sorcerer_soul_link"`
* **3. 目标选择规则**：
  * **SelectType**: `TeamOther`
  * **MinCount**: `1`
  * **MaxCount**: `1`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayRequired`
  * **Tokens**: `[{Type: SoulYellow, Amount: 1}, {Type: SoulBlue, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectPlacePlayedCardAsFieldMark`
    * `Target`: `TargetSelected`
    * `Value`: `1`
    * `FieldMarkRef`: `SoulLink`
    * `VisibilityRef`: `VisibilityPublic`
    * `Ref`: `None`

#### 【灵魂链接·伤害转移】 (Soul Link Damage Transfer)
* **说明**：当【灵魂链接】在场时，若你或链接持有者将承受伤害，可移除X点蓝色灵魂：将当前伤害减少X，并对另一方造成X点法术伤害。
* **1. 主干配置**：
  * **SkillID**: `skill_soul_sorcerer_soul_link_damage_transfer`
  * **Category**: `Exclusive`
  * **Type**: `Response`
  * **Timing**: `TimingOnDamageCalculated`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDamage]`
  * **CustomExpression**: `Combat.FinalDamage > 0 && Self.Tokens[SoulBlue] >= 1 && ((Combat.TargetID == Self.UserID && Action.Targets[0].TargetUserID != Self.UserID && State.CountFieldMark(Action.Targets[0].TargetUserID, SoulLink) >= 1) || (Combat.TargetID != Self.UserID && State.CountFieldMark(Combat.TargetID, SoulLink) >= 1 && Action.Targets[0].TargetUserID == Self.UserID))`
* **3. 目标选择规则**：
  * **SelectType**: `Teammate`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **SelectedValueRule**: `1 <= Action.SelectedValue && Action.SelectedValue <= Min(Self.Tokens[SoulBlue], Combat.FinalDamage)`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Tokens**: `[{Type: SoulBlue, Amount: Action.SelectedValue}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(原目标本次待生效伤害-X)*:
    * `EffectType`: `EffectModifyPendingDamage`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `-Action.SelectedValue`
    * `Ref`: `None`
  * **Effect[1]** *(对另一方造成X点法术伤害⑥)*:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelected`
    * `Value`: `Action.SelectedValue`
    * `Ref`: `None`

#### 【灵魂增幅】 (Soul Amplify)
* **技能描述**：［宝石］你+2［黄色灵魂］和2［蓝色灵魂］。
* **1. 主干配置**：
  * **SkillID**: `skill_soul_sorcerer_soul_amplify`
  * **Category**: `Ultimate`
  * **Type**: `Startup`
  * **Timing**: `TimingStartup`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionStart]`
  * **CustomExpression**: `None`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Gem, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `2`
    * `TokenRef`: `SoulYellow`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `2`
    * `TokenRef`: `SoulBlue`
    * `Ref`: `None`

---

### 35. 月之女神 (Moon Goddess)

#### 【新月庇护】 (New Moon Shelter)
* **技能描述**：［持续］（我方角色因承受伤害造成手牌数超过手牌上限，导致士气下降时）［横置］转为［闇月形态］，将因此而造成的弃牌面朝下放置于角色旁，作为［闇月］。本次士气不会下降。
* **1. 主干配置**：
  * **SkillID**: `skill_moon_goddess_new_moon_shelter`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnDamageTaken`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDraw]`
  * **CustomExpression**: `Event.PendingMoraleLoss > 0 && Event.OverflowDiscardCount > 0 && State.IsSameTeam(Self.UserID, Event.OverflowDiscardOwnerID) == true`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(横置，进入闇月形态)*:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Tapped`
    * `Ref`: `None`
  * **Effect[1]** *(写入命名形态)*:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `"moon_goddess_dark_moon_form"`
    * `Ref`: `None`
  * **Effect[2]** *(将本次爆牌弃牌转为暗月场标)*:
    * `EffectType`: `EffectPlaceOverflowDiscardAsFieldMark`
    * `Target`: `TargetSelf`
    * `Value`: `0` *(<=0 表示搬运本次爆牌弃牌全部实体)*
    * `FieldMarkRef`: `DarkMoon`
    * `VisibilityRef`: `VisibilityHidden`
    * `Ref`: `None`
  * **Effect[3]** *(抵消本次待扣士气)*:
    * `EffectType`: `EffectReducePendingMoraleLoss`
    * `Target`: `TargetCurrentEvent`
    * `Value`: `Event.PendingMoraleLoss`
    * `Ref`: `None`

#### 【闇月诅咒】 (Dark Moon Curse)
* **技能描述**：（你每次移除［闇月］）我方士气-1；（你的［闇月］数为0时）［转正］，脱离［闇月形态］。
* **1. 主干配置**：
  * **SkillID**: `skill_moon_goddess_dark_moon_curse`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnFieldMarkChanged`
* **2. 前置条件**：
  * **PhaseLimit**: `None`
  * **FieldMarkChangeFilter**: `{AcceptBehaviors: [Removed], AcceptTypesWhenRemoved: [DarkMoon]}`
  * **CustomExpression**: `Event.MarkType == DarkMoon && Event.MarkAction == "Removed" && Event.OperatorID == Self.UserID`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(每次移除暗月，我方士气-1)*:
    * `EffectType`: `EffectChangeMorale`
    * `Target`: `TargetSelfTeam`
    * `Value`: `-1`
    * `Ref`: `None`
  * **Effect[1]** *(暗月归零后转正)*:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Normal`
    * `Condition`: `State.CountFieldMark(Self.UserID, DarkMoon) == 0 && Self.Form == "moon_goddess_dark_moon_form"`
    * `Ref`: `None`
  * **Effect[2]** *(清除闇月形态名)*:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `nil`
    * `Condition`: `State.CountFieldMark(Self.UserID, DarkMoon) == 0 && Self.Form == "moon_goddess_dark_moon_form"`
    * `Ref`: `None`

#### 【美杜莎之眼】 (Medusa's Eye)
* **技能描述**：（目标对手攻击时①，移除1个与攻击牌相应系别的［闇月］［展示］）你+1［治疗］，+1石化。（若该［闇月］为法术牌）你弃1张牌，对目标对手造成1点法术伤害③。
* **1. 主干配置**：
  * **SkillID**: `skill_moon_goddess_medusa_eye`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnAttackDeclared`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDeclare]`
  * **CustomExpression**: `Combat.IsAttack == true && State.IsSameTeam(Self.UserID, Combat.SourceID) == false && State.CountFieldMarkByElement(Self.UserID, DarkMoon, Combat.AttackElement) >= 1`
* **3. 目标选择规则**：
  * **SelectType**: `Self`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **SubSelect**：
  * `SubType`: `FieldCard`
  * `SubFilter`: `FieldCard.HolderID == Self.UserID && FieldCard.FieldMark == DarkMoon && FieldCard.Element == Combat.AttackElement`
  * `SubMinCount`: `1`
  * `SubMaxCount`: `1`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(精确移除1个匹配系别的闇月)*:
    * `EffectType`: `EffectRemoveSelectedFieldCard`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `Ref`: `None`
  * **Effect[1]** *(你+1治疗)*:
    * `EffectType`: `EffectHeal`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `Ref`: `None`
  * **Effect[2]** *(你+1石化)*:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `TokenRef`: `Petrifaction`
    * `Ref`: `None`
  * **Effect[3]** *(若移除的是法术牌：你弃1张牌)*:
    * `EffectType`: `EffectDiscard`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `Condition`: `Event.RemovedFieldCardType == Magic`
    * `Ref`: `None`
  * **Effect[4]** *(若移除的是法术牌：对攻击者造成1点法术伤害)*:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetTriggerSource`
    * `Value`: `1`
    * `Condition`: `Event.RemovedFieldCardType == Magic`
    * `Ref`: `None`

#### 【月之轮回】 (Moon Reincarnation)
* **技能描述**：（你的回合结束时）选择以下一项发动：●（移除1个［闇月］）目标角色+1［治疗］。●（移除你的1个［治疗］）你+1［新月］。
* **1. 主干配置**：
  * **SkillID**: `skill_moon_goddess_moon_reincarnation`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnTurnEnd`
* **2. 前置条件**：
  * **PhaseLimit**: `[TurnEnd]`
  * **CustomExpression**: `State.IsInSelfTurn(Self.UserID) == true && (State.CountFieldMark(Self.UserID, DarkMoon) >= 1 || Self.Heal >= 1)`
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `0`
  * **MaxCount**: `1`
  * **SelectedValueRule**: `(Action.SelectedValue == 0 && Action.SelectedTargetCount == 1 && State.CountFieldMark(Self.UserID, DarkMoon) >= 1) || (Action.SelectedValue == 1 && Action.SelectedTargetCount == 0 && Self.Heal >= 1)` *(0=移除暗月给目标+1治疗；1=移除自身1治疗换1新月)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列（分支）**：
  * **BranchSelector**: `Action.SelectedValue`
  * **Branch[0]** *(移除1暗月，目标+1治疗)*:
    * **Effect[A0]**:
      * `EffectType`: `EffectRemoveFieldMark`
      * `Target`: `TargetSelf`
      * `Value`: `1`
      * `FieldMarkRef`: `DarkMoon`
      * `Ref`: `None`
    * **Effect[A1]**:
      * `EffectType`: `EffectHeal`
      * `Target`: `TargetSelected`
      * `Value`: `1`
      * `Ref`: `None`
  * **Branch[1]** *(移除自身1治疗，自己+1新月)*:
    * **Effect[B0]**:
      * `EffectType`: `EffectHeal`
      * `Target`: `TargetSelf`
      * `Value`: `-1`
      * `Ref`: `None`
    * **Effect[B1]**:
      * `EffectType`: `EffectAddToken`
      * `Target`: `TargetSelf`
      * `Value`: `1`
      * `TokenRef`: `Crescent`
      * `Ref`: `None`

#### 【月渎】 (Moon Desecration)
* **技能描述**：［回合限定］（目标角色承受你造成的法术伤害后⑥，移除你的1个［治疗］）对目标对手造成1点法术伤害③。
* **1. 主干配置**：
  * **SkillID**: `skill_moon_goddess_moon_blight`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnDamageTaken`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDamage]`
  * **IsTurnLimited**: `true`
  * **CustomExpression**: `State.IsInSelfTurn(Self.UserID) == true && Combat.SourceID == Self.UserID && Combat.IsMagic == true && Combat.FinalDamage > 0 && State.IsSameTeam(Self.UserID, Combat.TargetID) == false && Self.Heal >= 1`
* **3. 目标选择规则**：
  * **SelectType**: `Enemy`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **SelectedValueRule**: `Action.Targets[0].TargetUserID == Combat.TargetID`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **HealCost**: `1`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelected`
    * `Value`: `1`
    * `Ref`: `None`

#### 【闇月斩】 (Dark Moon Slash)
* **技能描述**：［水晶］（仅［闇月形态］下可发动，主动攻击命中时②，移除X个［闇月］（X<3））本次攻击伤害额外+X。
* **1. 主干配置**：
  * **SkillID**: `skill_moon_goddess_dark_moon_slash`
  * **Category**: `Ultimate`
  * **Type**: `Response`
  * **Timing**: `TimingOnHitCheck`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatHitCheck]`
  * **CustomExpression**: `Self.Form == "moon_goddess_dark_moon_form" && Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && Combat.IsHit == true && State.CountFieldMark(Self.UserID, DarkMoon) >= 1`
* **3. 目标选择规则**：
  * **SelectType**: `None`
  * **SelectedValueRule**: `1 <= Action.SelectedValue && Action.SelectedValue < 3 && Action.SelectedValue <= State.CountFieldMark(Self.UserID, DarkMoon)`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Crystal, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(移除X个闇月)*:
    * `EffectType`: `EffectRemoveFieldMark`
    * `Target`: `TargetSelf`
    * `Value`: `Action.SelectedValue`
    * `FieldMarkRef`: `DarkMoon`
    * `Ref`: `None`
  * **Effect[1]** *(本次攻击伤害额外+X)*:
    * `EffectType`: `EffectAttackDamageModifier`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `Action.SelectedValue`
    * `Ref`: `None`

#### 【苍白之月】 (Pale Moon)
* **技能描述**：［宝石］选择以下一项发动：●（移除3点［石化］）你的下次主动攻击对手无法应战，额外+1［攻击行动］。你额外获得一个回合。●移除X点［新月］，你+1［石化］，弃1张牌，对目标对手造成（X+1）点法术伤害③。
* **1. 主干配置**：
  * **SkillID**: `skill_moon_goddess_pale_moon`
  * **Category**: `Ultimate`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **1.2 规则修饰器模板配置**：
  * **Template[A]** *(“下次主动攻击无法应战”武装标记)*:
    * **ModifierID**: `rm_moon_goddess_next_active_attack_unrespondable_armed`
    * **Domain**: `RuleModifierDomainCardSource`
    * **Priority**: `80`
    * **ConditionExpression**: `None`
    * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
    * **CardSourcePayload**: `{ProjectionMode: CardSourceProjectionAsHand, FieldMarks: []}` *(空载荷，仅作一次性武装标记)*
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Self.Tokens[Petrifaction] >= 3 || Self.Tokens[Crescent] >= 1`
* **3. 目标选择规则**：
  * **SelectType**: `Enemy`
  * **MinCount**: `0`
  * **MaxCount**: `1`
  * **SelectedValueRule**: `(Action.SelectedValue == 0 && Action.SelectedTargetCount == 0 && Self.Tokens[Petrifaction] >= 3) || (Action.SelectedValue >= 1 && Action.SelectedTargetCount == 1 && Action.SelectedValue <= Self.Tokens[Crescent])` *(0=石化分支；>=1=新月分支且 SelectedValue=X)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Gem, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列（分支）**：
  * **Effect[0]** *(石化分支：移除3石化)*:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `-3`
    * `TokenRef`: `Petrifaction`
    * `Condition`: `Action.SelectedValue == 0`
    * `Ref`: `None`
  * **Effect[1]** *(石化分支：挂载“下次主动攻击无法应战”标记)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_moon_goddess_next_active_attack_unrespondable_armed"`
    * `RuleLifetimeRef`: `RuleLifePermanent`
    * `Condition`: `Action.SelectedValue == 0`
    * `Ref`: `None`
  * **Effect[2]** *(石化分支：额外+1攻击行动)*:
    * `Implementation`: `model.AppendExtraAction`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `ActionRef`: `Attack`
    * `Condition`: `Action.SelectedValue == 0`
    * `Ref`: `None`
  * **Effect[3]** *(石化分支：额外获得1个回合)*:
    * `EffectType`: `EffectGrantExtraTurn`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `Condition`: `Action.SelectedValue == 0`
    * `Ref`: `None`
  * **Effect[4]** *(新月分支：移除X新月)*:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `-Action.SelectedValue`
    * `TokenRef`: `Crescent`
    * `Condition`: `Action.SelectedValue > 0`
    * `Ref`: `None`
  * **Effect[5]** *(新月分支：你+1石化)*:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `TokenRef`: `Petrifaction`
    * `Condition`: `Action.SelectedValue > 0`
    * `Ref`: `None`
  * **Effect[6]** *(新月分支：你弃1张牌)*:
    * `EffectType`: `EffectDiscard`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `Condition`: `Action.SelectedValue > 0`
    * `Ref`: `None`
  * **Effect[7]** *(新月分支：对目标造成(X+1)法术伤害)*:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelected`
    * `Value`: `Action.SelectedValue + 1`
    * `Condition`: `Action.SelectedValue > 0`
    * `Ref`: `None`

#### 【苍白之月·下次主动攻击无法应战】 (Pale Moon Next Active Attack Unrespondable)
* **说明**：消费【苍白之月】挂载的“下次主动攻击无法应战”标记，仅生效1次。
* **1. 主干配置**：
  * **SkillID**: `skill_moon_goddess_pale_moon_next_attack_unrespondable`
  * **Category**: `Ultimate`
  * **Type**: `Passive`
  * **Timing**: `TimingOnAttackDeclared`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDeclare]`
  * **CustomExpression**: `Combat.SourceID == Self.UserID && Combat.IsAttack == true && Combat.IsActiveAttack == true && State.HasRuleModifier(Self.UserID, "rm_moon_goddess_next_active_attack_unrespondable_armed") == true`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(本次主动攻击不可应战)*:
    * `EffectType`: `EffectApplyCombatTag`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `Unrespondable`
    * `Ref`: `None`
  * **Effect[1]** *(消耗一次性标记)*:
    * `EffectType`: `EffectRemoveRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleRemoveRef`: `{Mode: RuleRemoveByModifierID, ModifierID: "rm_moon_goddess_next_active_attack_unrespondable_armed", Limit: 1}`
    * `Ref`: `None`

---

### 36. 血之巫女 (Blood Witch)

#### 【血之哀伤】 (Blood Sorrow)
* **技能描述**：（对自己造成2点法术伤害③）转移同生共死的目标或是移除同生共死。
* **1. 主干配置**：
  * **SkillID**: `skill_blood_witch_blood_sorrow`
  * **Category**: `Normal`
  * **Type**: `Startup`
  * **Timing**: `TimingStartup`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionStart]`
  * **CustomExpression**: `None`
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `0`
  * **MaxCount**: `1`
  * **SelectedValueRule**: `(Action.SelectedValue == 0 && Action.SelectedTargetCount == 0) || (Action.SelectedValue == 1 && Action.SelectedTargetCount == 1)` *(0=移除同生共死；1=转移同生共死到所选目标)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **HPCost**: `2`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(对自己造成2点法术伤害)*:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelf`
    * `Value`: `2`
    * `Ref`: `None`
  * **Effect[1]** *(转移“在场实体”的同生共死到新目标)*:
    * `EffectType`: `EffectTransferFieldMark`
    * `Target`: `TargetSelected`
    * `FromTargetRef`: `TargetAllPlayers`
    * `Value`: `1`
    * `FieldMarkRef`: `SharedFate`
    * `Condition`: `Action.SelectedValue == 1`
    * `Ref`: `None`
  * **Effect[2]** *(移除场上同生共死)*:
    * `EffectType`: `EffectRemoveFieldMark`
    * `Target`: `TargetAllPlayers`
    * `Value`: `1`
    * `FieldMarkRef`: `SharedFate`
    * `Condition`: `Action.SelectedValue == 0`
    * `Ref`: `None`
  * **Effect[3]** *(若执行“移除同生共死”，清理自身-2手牌上限规则)*:
    * `EffectType`: `EffectRemoveRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleRemoveRef`: `{Mode: RuleRemoveByModifierID, ModifierID: "rm_blood_witch_shared_fate_self_normal_minus2", Limit: 1}`
    * `Condition`: `Action.SelectedValue == 0`
    * `Ref`: `None`
  * **Effect[4]** *(若执行“移除同生共死”，清理自身+1手牌上限规则)*:
    * `EffectType`: `EffectRemoveRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleRemoveRef`: `{Mode: RuleRemoveByModifierID, ModifierID: "rm_blood_witch_shared_fate_self_bleeding_plus1", Limit: 1}`
    * `Condition`: `Action.SelectedValue == 0`
    * `Ref`: `None`
  * **Effect[5]** *(若执行“移除同生共死”，清理持有者-2手牌上限规则)*:
    * `EffectType`: `EffectRemoveRuleModifier`
    * `Target`: `TargetAllPlayers`
    * `Value`: `0`
    * `RuleRemoveRef`: `{Mode: RuleRemoveByModifierID, ModifierID: "rm_blood_witch_shared_fate_holder_normal_minus2", Limit: 0}`
    * `Condition`: `Action.SelectedValue == 0`
    * `Ref`: `None`
  * **Effect[6]** *(若执行“移除同生共死”，清理持有者+1手牌上限规则)*:
    * `EffectType`: `EffectRemoveRuleModifier`
    * `Target`: `TargetAllPlayers`
    * `Value`: `0`
    * `RuleRemoveRef`: `{Mode: RuleRemoveByModifierID, ModifierID: "rm_blood_witch_shared_fate_holder_bleeding_plus1", Limit: 0}`
    * `Condition`: `Action.SelectedValue == 0`
    * `Ref`: `None`

#### 【流血】 (Bleeding)
* **技能描述**：［持续］（当你在普通形态下因承受伤害而导致我方士气减少时强制发动［强制］）［横置］转为［流血形态］，你+1［治疗］。
* **1. 主干配置**：
  * **SkillID**: `skill_blood_witch_bleeding_enter`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnDamageTaken`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDraw]`
  * **CustomExpression**: `Self.Form != "blood_witch_bleeding_form" && Combat.TargetID == Self.UserID && Event.PendingMoraleLoss > 0`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(横置进入流血形态)*:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Tapped`
    * `Ref`: `None`
  * **Effect[1]** *(写入流血形态)*:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `"blood_witch_bleeding_form"`
    * `Ref`: `None`
  * **Effect[2]** *(你+1治疗)*:
    * `EffectType`: `EffectHeal`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `Ref`: `None`
  * **Effect[3]** *(若同生共死生效中：移除自身-2手牌上限，切为+1)*:
    * `EffectType`: `EffectRemoveRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleRemoveRef`: `{Mode: RuleRemoveByModifierID, ModifierID: "rm_blood_witch_shared_fate_self_normal_minus2", Limit: 1}`
    * `Condition`: `State.HasRuleModifier(Self.UserID, "rm_blood_witch_shared_fate_self_normal_minus2") == true`
    * `Ref`: `None`
  * **Effect[4]** *(若同生共死生效中：移除持有者-2手牌上限规则)*:
    * `EffectType`: `EffectRemoveRuleModifier`
    * `Target`: `TargetAllPlayers`
    * `Value`: `0`
    * `RuleRemoveRef`: `{Mode: RuleRemoveByModifierID, ModifierID: "rm_blood_witch_shared_fate_holder_normal_minus2", Limit: 0}`
    * `Condition`: `State.HasRuleModifier(Self.UserID, "rm_blood_witch_shared_fate_self_normal_minus2") == true`
    * `Ref`: `None`
  * **Effect[5]** *(若同生共死生效中：施加自身+1手牌上限规则)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_blood_witch_shared_fate_self_bleeding_plus1"`
    * `RuleLifetimeRef`: `RuleLifePermanent`
    * `Condition`: `State.HasRuleModifier(Self.UserID, "rm_blood_witch_shared_fate_self_normal_minus2") == true`
    * `Ref`: `None`
  * **Effect[6]** *(若同生共死生效中：施加持有者+1手牌上限规则)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetAllPlayers`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_blood_witch_shared_fate_holder_bleeding_plus1"`
    * `RuleLifetimeRef`: `RuleLifePermanent`
    * `Condition`: `State.HasRuleModifier(Self.UserID, "rm_blood_witch_shared_fate_self_normal_minus2") == true`
    * `Ref`: `None`

#### 【流血·回合开始自损】 (Bleeding Turn Start Self Damage)
* **说明**：处于流血形态时，在你的每次回合开始时对自己造成1点法术伤害。
* **1. 主干配置**：
  * **SkillID**: `skill_blood_witch_bleeding_turn_start_self_damage`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnTurnStart`
* **2. 前置条件**：
  * **PhaseLimit**: `[TurnStart]`
  * **CustomExpression**: `State.IsInSelfTurn(Self.UserID) == true && Self.Form == "blood_witch_bleeding_form"`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `Ref`: `None`

#### 【流血·手牌不足脱离】 (Bleeding Exit On Low Hand)
* **说明**：当处于流血形态且自身手牌<3时强制转正，脱离流血形态。
* **1. 主干配置**：
  * **SkillID**: `skill_blood_witch_bleeding_exit_on_low_hand`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnActionEnd`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionEnd]`
  * **CustomExpression**: `Self.Form == "blood_witch_bleeding_form" && Self.HandCount < 3`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(转正)*:
    * `EffectType`: `EffectSetOrientation`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `OrientationRef`: `Normal`
    * `Ref`: `None`
  * **Effect[1]** *(清空流血形态)*:
    * `EffectType`: `EffectSetForm`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `FormRef`: `nil`
    * `Ref`: `None`
  * **Effect[2]** *(若同生共死生效中：移除自身+1手牌上限，切回-2)*:
    * `EffectType`: `EffectRemoveRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleRemoveRef`: `{Mode: RuleRemoveByModifierID, ModifierID: "rm_blood_witch_shared_fate_self_bleeding_plus1", Limit: 1}`
    * `Condition`: `State.HasRuleModifier(Self.UserID, "rm_blood_witch_shared_fate_self_bleeding_plus1") == true`
    * `Ref`: `None`
  * **Effect[3]** *(若同生共死生效中：移除持有者+1手牌上限规则)*:
    * `EffectType`: `EffectRemoveRuleModifier`
    * `Target`: `TargetAllPlayers`
    * `Value`: `0`
    * `RuleRemoveRef`: `{Mode: RuleRemoveByModifierID, ModifierID: "rm_blood_witch_shared_fate_holder_bleeding_plus1", Limit: 0}`
    * `Condition`: `State.HasRuleModifier(Self.UserID, "rm_blood_witch_shared_fate_self_bleeding_plus1") == true`
    * `Ref`: `None`
  * **Effect[4]** *(若同生共死生效中：施加自身-2手牌上限规则)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_blood_witch_shared_fate_self_normal_minus2"`
    * `RuleLifetimeRef`: `RuleLifePermanent`
    * `Condition`: `State.HasRuleModifier(Self.UserID, "rm_blood_witch_shared_fate_self_bleeding_plus1") == true`
    * `Ref`: `None`
  * **Effect[5]** *(若同生共死生效中：施加持有者-2手牌上限规则)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetAllPlayers`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_blood_witch_shared_fate_holder_normal_minus2"`
    * `RuleLifetimeRef`: `RuleLifePermanent`
    * `Condition`: `State.HasRuleModifier(Self.UserID, "rm_blood_witch_shared_fate_self_bleeding_plus1") == true`
    * `Ref`: `None`

#### 【逆流】 (Counterflow)
* **技能描述**：（仅［流血形态］下发动）你弃2张牌，你+1［治疗］。
* **1. 主干配置**：
  * **SkillID**: `skill_blood_witch_counterflow`
  * **Category**: `Normal`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Self.Form == "blood_witch_bleeding_form"`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Discards**:
    * `Count`: `2`
    * `Visibility`: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectHeal`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `Ref`: `None`

#### 【血之悲鸣】 (Blood Lament)
* **技能描述**：（仅［流血形态］下发动）对目标角色和自己各造成（X+1）点法术伤害③，X<3。
* **1. 主干配置**：
  * **SkillID**: `skill_blood_witch_blood_lament`
  * **Category**: `Unique`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Self.Form == "blood_witch_bleeding_form" && Action.PlayedCard.CharacterSkillMap[Self.CharacterID] == "skill_blood_witch_blood_lament"`
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **SelectedValueRule**: `0 <= Action.SelectedValue && Action.SelectedValue < 3`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(对目标造成(X+1)法术伤害)*:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelected`
    * `Value`: `Action.SelectedValue + 1`
    * `Ref`: `None`
  * **Effect[1]** *(对自己造成(X+1)法术伤害)*:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelf`
    * `Value`: `Action.SelectedValue + 1`
    * `Ref`: `None`

#### 【同生共死】 (Shared Fate)
* **技能描述**：（你摸2张牌［强制］）将同生共死放置于目标角色面前。（在［普通形态］下）你和他手牌上限各-2。（在［流血形态］下）你和他手牌上限各+1。
* **1. 主干配置**：
  * **SkillID**: `skill_blood_witch_shared_fate`
  * **Category**: `Exclusive`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **1.2 规则修饰器模板配置**：
  * **Template[A]** *(同生共死：血之巫女自身（普通形态）手牌上限-2)*:
    * **ModifierID**: `rm_blood_witch_shared_fate_self_normal_minus2`
    * **Domain**: `RuleModifierDomainAttribute`
    * **Priority**: `130`
    * **ConditionExpression**: `None`
    * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
    * **AttrPayload**: `{AttrType: PlayerAttributeMaxHand, Operation: AttributeModifyAdd, ValueSourceMode: RuleAttrValueFromFixed, Value: -2}`
  * **Template[B]** *(同生共死：血之巫女自身（流血形态）手牌上限+1)*:
    * **ModifierID**: `rm_blood_witch_shared_fate_self_bleeding_plus1`
    * **Domain**: `RuleModifierDomainAttribute`
    * **Priority**: `130`
    * **ConditionExpression**: `None`
    * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
    * **AttrPayload**: `{AttrType: PlayerAttributeMaxHand, Operation: AttributeModifyAdd, ValueSourceMode: RuleAttrValueFromFixed, Value: 1}`
  * **Template[C]** *(同生共死持有者（普通形态逻辑）手牌上限-2)*:
    * **ModifierID**: `rm_blood_witch_shared_fate_holder_normal_minus2`
    * **Domain**: `RuleModifierDomainAttribute`
    * **Priority**: `120`
    * **ConditionExpression**: `State.CountFieldMark(Self.UserID, SharedFate) >= 1`
    * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
    * **AttrPayload**: `{AttrType: PlayerAttributeMaxHand, Operation: AttributeModifyAdd, ValueSourceMode: RuleAttrValueFromFixed, Value: -2}`
  * **Template[D]** *(同生共死持有者（流血形态逻辑）手牌上限+1)*:
    * **ModifierID**: `rm_blood_witch_shared_fate_holder_bleeding_plus1`
    * **Domain**: `RuleModifierDomainAttribute`
    * **Priority**: `120`
    * **ConditionExpression**: `State.CountFieldMark(Self.UserID, SharedFate) >= 1`
    * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
    * **AttrPayload**: `{AttrType: PlayerAttributeMaxHand, Operation: AttributeModifyAdd, ValueSourceMode: RuleAttrValueFromFixed, Value: 1}`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Action.PlayedCard.CharacterSkillMap[Self.CharacterID] == "skill_blood_witch_shared_fate"`
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `1`
  * **MaxCount**: `1`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(你摸2张牌)*:
    * `EffectType`: `EffectDrawCard`
    * `Target`: `TargetSelf`
    * `Value`: `2`
    * `Ref`: `None`
  * **Effect[1]** *(将本次打出的同生共死牌实体放置于目标角色面前)*:
    * `EffectType`: `EffectPlacePlayedCardAsFieldMark`
    * `Target`: `TargetSelected`
    * `Value`: `1`
    * `FieldMarkRef`: `SharedFate`
    * `VisibilityRef`: `VisibilityPublic`
    * `Ref`: `None`
  * **Effect[2]** *(先清理自身旧规则：普通形态版本)*:
    * `EffectType`: `EffectRemoveRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleRemoveRef`: `{Mode: RuleRemoveByModifierID, ModifierID: "rm_blood_witch_shared_fate_self_normal_minus2", Limit: 1}`
    * `Ref`: `None`
  * **Effect[3]** *(先清理自身旧规则：流血形态版本)*:
    * `EffectType`: `EffectRemoveRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleRemoveRef`: `{Mode: RuleRemoveByModifierID, ModifierID: "rm_blood_witch_shared_fate_self_bleeding_plus1", Limit: 1}`
    * `Ref`: `None`
  * **Effect[4]** *(先清理持有者旧规则：普通形态版本)*:
    * `EffectType`: `EffectRemoveRuleModifier`
    * `Target`: `TargetAllPlayers`
    * `Value`: `0`
    * `RuleRemoveRef`: `{Mode: RuleRemoveByModifierID, ModifierID: "rm_blood_witch_shared_fate_holder_normal_minus2", Limit: 0}`
    * `Ref`: `None`
  * **Effect[5]** *(先清理持有者旧规则：流血形态版本)*:
    * `EffectType`: `EffectRemoveRuleModifier`
    * `Target`: `TargetAllPlayers`
    * `Value`: `0`
    * `RuleRemoveRef`: `{Mode: RuleRemoveByModifierID, ModifierID: "rm_blood_witch_shared_fate_holder_bleeding_plus1", Limit: 0}`
    * `Ref`: `None`
  * **Effect[6]** *(普通形态：施加血之巫女自身手牌上限-2)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_blood_witch_shared_fate_self_normal_minus2"`
    * `RuleLifetimeRef`: `RuleLifePermanent`
    * `Condition`: `Self.Form != "blood_witch_bleeding_form"`
    * `Ref`: `None`
  * **Effect[7]** *(普通形态：施加同生共死持有者手牌上限-2)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetAllPlayers`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_blood_witch_shared_fate_holder_normal_minus2"`
    * `RuleLifetimeRef`: `RuleLifePermanent`
    * `Condition`: `Self.Form != "blood_witch_bleeding_form"`
    * `Ref`: `None`
  * **Effect[8]** *(流血形态：施加血之巫女自身手牌上限+1)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_blood_witch_shared_fate_self_bleeding_plus1"`
    * `RuleLifetimeRef`: `RuleLifePermanent`
    * `Condition`: `Self.Form == "blood_witch_bleeding_form"`
    * `Ref`: `None`
  * **Effect[9]** *(流血形态：施加同生共死持有者手牌上限+1)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetAllPlayers`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_blood_witch_shared_fate_holder_bleeding_plus1"`
    * `RuleLifetimeRef`: `RuleLifePermanent`
    * `Condition`: `Self.Form == "blood_witch_bleeding_form"`
    * `Ref`: `None`

#### 【血之诅咒】 (Blood Curse)
* **技能描述**：［宝石］对目标角色造成2点法术伤害③，你弃3张牌。
* **1. 主干配置**：
  * **SkillID**: `skill_blood_witch_blood_curse`
  * **Category**: `Ultimate`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `None`
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `1`
  * **MaxCount**: `1`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Gem, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelected`
    * `Value`: `2`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectDiscard`
    * `Target`: `TargetSelf`
    * `Value`: `3`
    * `Ref`: `None`

---

### 37. 蝶舞者 (Butterfly Dancer)

#### 【生命之火】 (Flame of Life)
* **技能描述**：你的手牌上限-X，X为你拥有的［蛹］的数量，但你的手牌上限最少为3。
* **1. 主干配置**：
  * **SkillID**: `skill_butterfly_dancer_flame_of_life`
  * **Category**: `Normal`
  * **Type**: `Passive`
  * **Timing**: `TimingOnTurnStart` *(用于承接 GameStart，挂载常驻规则模板)*
* **1.2 规则修饰器模板配置**：
  * **ModifierID**: `rm_butterfly_dancer_flame_of_life_hand_limit_delta_by_pupa`
  * **Domain**: `RuleModifierDomainAttribute`
  * **Priority**: `120`
  * **ConditionExpression**: `None`
  * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
  * **AttrPayload**: `{AttrType: PlayerAttributeMaxHand, Operation: AttributeModifyAdd, ValueSourceMode: RuleAttrValueFromTokenLinear, TokenLink: {OwnerScope: RuleAttrTokenOwnerTarget, TokenType: Pupa, Coefficient: -1, Offset: 0, MinValue: -3, MaxValue: 0}}` *(按基础手牌上限6建模，确保不低于3)*
* **2. 前置条件**：
  * **PhaseLimit**: `[GameInit]`
  * **CustomExpression**: `Event.SourceType == SourceSystem && Event.CauseAction == "GameStart"`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_butterfly_dancer_flame_of_life_hand_limit_delta_by_pupa"`
    * `RuleLifetimeRef`: `RuleLifePermanent`
    * `Ref`: `None`

#### 【舞动】 (Dance)
* **技能描述**：（摸1张牌［强制］或弃1张牌［强制］）将牌库顶的1张牌面朝下放置在你角色旁，作为［茧］。
* **1. 主干配置**：
  * **SkillID**: `skill_butterfly_dancer_dance`
  * **Category**: `Normal`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Action.SelectedValue == 0 || Action.SelectedValue == 1`
* **3. 目标选择规则**：
  * **SelectType**: `None`
  * **SelectedValueRule**: `Action.SelectedValue == 0 || Action.SelectedValue == 1` *(0=摸1；1=弃1)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(选择摸1)*:
    * `EffectType`: `EffectDrawCard`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `Condition`: `Action.SelectedValue == 0`
    * `Ref`: `None`
  * **Effect[1]** *(选择弃1)*:
    * `EffectType`: `EffectDiscard`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `Condition`: `Action.SelectedValue == 1`
    * `Ref`: `None`
  * **Effect[2]** *(牌库顶1张变为茧)*:
    * `EffectType`: `EffectPlaceDeckTopAsFieldMark`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `FieldMarkRef`: `Cocoon`
    * `Ref`: `None`

#### 【毒粉】 (Poison Powder)
* **技能描述**：（每当有角色产生1点实际法术伤害时发动⑤，移除1个［茧］）该次伤害额外+1。
* **1. 主干配置**：
  * **SkillID**: `skill_butterfly_dancer_poison_powder`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnDamageCalculated`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDamage]`
  * **CustomExpression**: `Combat.IsMagic == true && Combat.FinalDamage == 1 && State.CountFieldMark(Self.UserID, Cocoon) >= 1`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectRemoveFieldMark`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `FieldMarkRef`: `Cocoon`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectModifyPendingDamage`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `1`
    * `Ref`: `None`

#### 【朝圣】 (Pilgrimage)
* **技能描述**：（每当你承受伤害时发动⑥，移除1个［茧］）抵御1点该来源的伤害。
* **1. 主干配置**：
  * **SkillID**: `skill_butterfly_dancer_pilgrimage`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnDamageCalculated`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDamage]`
  * **CustomExpression**: `Combat.TargetID == Self.UserID && Combat.FinalDamage > 0 && State.CountFieldMark(Self.UserID, Cocoon) >= 1`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectRemoveFieldMark`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `FieldMarkRef`: `Cocoon`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectModifyPendingDamage`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `-1`
    * `Ref`: `None`

#### 【镜花水月】 (Mirror Blossom Water Moon)
* **技能描述**：（每当有角色产生2点实际法术伤害时发动⑤，移除2张同系［茧］［展示］）抵御该次伤害，你对他造成2次法术伤害③，每次伤害为1点。
* **1. 主干配置**：
  * **SkillID**: `skill_butterfly_dancer_mirror_blossom_water_moon`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnDamageCalculated`
* **2. 前置条件**：
  * **PhaseLimit**: `[CombatDamage]`
  * **CustomExpression**: `Combat.IsMagic == true && Combat.FinalDamage == 2 && State.CountFieldMark(Self.UserID, Cocoon) >= 2 && Action.ElementRef != None`
* **3. 目标选择规则**：
  * **SelectType**: `Self`
  * **MinCount**: `1`
  * **MaxCount**: `1`
  * **SubSelect**：
  * `SubType`: `FieldCard`
  * `SubFilter`: `FieldCard.HolderID == Self.UserID && FieldCard.FieldMark == Cocoon && FieldCard.Element == Action.ElementRef`
  * `SubMinCount`: `2`
  * `SubMaxCount`: `2`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(精确移除2张同系茧)*:
    * `EffectType`: `EffectRemoveSelectedFieldCard`
    * `Target`: `TargetSelf`
    * `Value`: `2`
    * `Ref`: `None`
  * **Effect[1]** *(抵御该次伤害)*:
    * `EffectType`: `EffectModifyPendingDamage`
    * `Target`: `TargetCurrentCombat`
    * `Value`: `-Combat.FinalDamage`
    * `Ref`: `None`
  * **Effect[2]** *(对伤害来源造成第1次1点法术伤害)*:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetTriggerSource`
    * `Value`: `1`
    * `Ref`: `None`
  * **Effect[3]** *(对伤害来源造成第2次1点法术伤害)*:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetTriggerSource`
    * `Value`: `1`
    * `Ref`: `None`

#### 【凋零】 (Withering)
* **技能描述**：（你每次移除［茧］时，若为法术牌，可展示之［展示］）你对目标角色造成1点法术伤害③，再对自己造成2点法术伤害③；此技能发动后，直到你下个回合开始前，对方的士气最少为1［强制］。
* **1. 主干配置**：
  * **SkillID**: `skill_butterfly_dancer_withering`
  * **Category**: `Normal`
  * **Type**: `Response`
  * **Timing**: `TimingOnFieldMarkChanged`
* **1.2 规则修饰器模板配置**：
  * **ModifierID**: `rm_butterfly_dancer_withering_enemy_morale_floor_1`
  * **Domain**: `RuleModifierDomainMoralePolicy`
  * **Priority**: `240`
  * **ConditionExpression**: `None`
  * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
  * **MoralePolicyPayload**: `{ApplyScope: MoralePolicyApplyEnemyTeam, MinMorale: 1}`
* **2. 前置条件**：
  * **PhaseLimit**: `None`
  * **FieldMarkChangeFilter**: `{AcceptBehaviors: [Removed], AcceptTypesWhenRemoved: [Cocoon]}`
  * **CustomExpression**: `Event.OperatorID == Self.UserID`
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `1`
  * **MaxCount**: `1`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]** *(若本次被移除茧为法术牌，则公开展示)*:
    * `EffectType`: `EffectRevealRemovedFieldCard`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `Condition`: `Event.RemovedFieldCardType == Magic`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelected`
    * `Value`: `1`
    * `Ref`: `None`
  * **Effect[2]**:
    * `EffectType`: `EffectDamage`
    * `Target`: `TargetSelf`
    * `Value`: `2`
    * `Ref`: `None`
  * **Effect[3]** *(施加“敌方士气最低为1”的临时策略，持续到你下回合开始)*:
    * `EffectType`: `EffectApplyRuleModifier`
    * `Target`: `TargetSelf`
    * `Value`: `0`
    * `RuleModifierRef`: `"rm_butterfly_dancer_withering_enemy_morale_floor_1"`
    * `RuleLifetimeRef`: `RuleLifeUntilSourceNextTurnStart`
    * `Ref`: `None`

#### 【蛹化】 (Metamorphosis)
* **技能描述**：［宝石］（你+1［蛹］）将牌库顶的4张牌面朝下放置在你角色旁，作为［茧］。
* **1. 主干配置**：
  * **SkillID**: `skill_butterfly_dancer_metamorphosis`
  * **Category**: `Ultimate`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `None`
* **3. 目标选择规则**：
  * **SelectType**: `None`
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Gem, Amount: 1}]`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列**：
  * **Effect[0]**:
    * `EffectType`: `EffectAddToken`
    * `Target`: `TargetSelf`
    * `Value`: `1`
    * `TokenRef`: `Pupa`
    * `Ref`: `None`
  * **Effect[1]**:
    * `EffectType`: `EffectPlaceDeckTopAsFieldMark`
    * `Target`: `TargetSelf`
    * `Value`: `4`
    * `FieldMarkRef`: `Cocoon`
    * `Ref`: `None`

#### 【倒逆之蝶】 (Inverse Butterfly)
* **技能描述**：［水晶］你弃2张牌，再选择以下1项发动：●对目标角色造成1点法术伤害③，该伤害不能用［治疗］抵御。●（移除2个［茧］或对自己造成4点法术伤害③）移除1个［蛹］。
* **1. 主干配置**：
  * **SkillID**: `skill_butterfly_dancer_inverse_butterfly`
  * **Category**: `Ultimate`
  * **Type**: `Magic`
  * **Timing**: `TimingActive`
* **1.2 规则修饰器模板配置**：
  * **ModifierID**: `rm_butterfly_dancer_inverse_butterfly_target_unhealable_once`
  * **Domain**: `RuleModifierDomainHealResistPolicy`
  * **Priority**: `260`
  * **ConditionExpression**: `None`
  * **StackPolicy**: `RuleModifierStackRefreshByModifierID`
  * **HealResistPolicyPayload**: `{PerDamageWindowHealCap: 0}`
* **2. 前置条件**：
  * **PhaseLimit**: `[ActionExecution]`
  * **CustomExpression**: `Action.SelectedValue == 0 || Action.SelectedValue == 1 || Action.SelectedValue == 2`
* **3. 目标选择规则**：
  * **SelectType**: `Any`
  * **MinCount**: `0`
  * **MaxCount**: `1`
  * **SelectedValueRule**: `(Action.SelectedValue == 0 && Action.SelectedTargetCount == 1) || (Action.SelectedValue == 1 && Action.SelectedTargetCount == 0 && State.CountFieldMark(Self.UserID, Cocoon) >= 2 && Self.Tokens[Pupa] >= 1) || (Action.SelectedValue == 2 && Action.SelectedTargetCount == 0 && Self.Tokens[Pupa] >= 1)` *(0=点伤且不可治疗抵御；1=移除2茧后移除1蛹；2=自损4后移除1蛹)*
* **4. 费用消耗**：
  * **CardPlayCostType**: `CardPlayNotRequired`
  * **Stones**: `[{Type: Crystal, Amount: 1}]`
  * **Discards**:
    * `Count`: `2`
    * `Visibility`: `VisibilityPublic`
* **5. 战斗与结算劫持标记**：
  * **Tags**: `None`
* **6. 执行效果序列（分支）**：
  * **BranchSelector**: `Action.SelectedValue`
  * **Branch[0]** *(目标1点法伤，且该伤害不可被治疗抵御)*:
    * **Effect[A0]** *(给目标挂“本链路不可治疗抵伤”规则)*:
      * `EffectType`: `EffectApplyRuleModifier`
      * `Target`: `TargetSelected`
      * `Value`: `0`
      * `RuleModifierRef`: `"rm_butterfly_dancer_inverse_butterfly_target_unhealable_once"`
      * `RuleLifetimeRef`: `RuleLifeThisEffectChain`
      * `Ref`: `None`
    * **Effect[A1]**:
      * `EffectType`: `EffectDamage`
      * `Target`: `TargetSelected`
      * `Value`: `1`
      * `Ref`: `None`
  * **Branch[1]** *(移除2茧，移除1蛹)*:
    * **Effect[B0]**:
      * `EffectType`: `EffectRemoveFieldMark`
      * `Target`: `TargetSelf`
      * `Value`: `2`
      * `FieldMarkRef`: `Cocoon`
      * `Ref`: `None`
    * **Effect[B1]**:
      * `EffectType`: `EffectAddToken`
      * `Target`: `TargetSelf`
      * `Value`: `-1`
      * `TokenRef`: `Pupa`
      * `Ref`: `None`
  * **Branch[2]** *(对自己造成4点法伤，移除1蛹)*:
    * **Effect[C0]**:
      * `EffectType`: `EffectDamage`
      * `Target`: `TargetSelf`
      * `Value`: `4`
      * `Ref`: `None`
    * **Effect[C1]**:
      * `EffectType`: `EffectAddToken`
      * `Target`: `TargetSelf`
      * `Value`: `-1`
      * `TokenRef`: `Pupa`
      * `Ref`: `None`
