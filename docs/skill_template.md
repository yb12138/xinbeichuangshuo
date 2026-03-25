## 🔖 技能开发配置模板 (Strict Go-Struct Aligned Template)

```markdown
### 【技能名称】 (Skill Name)
* **技能描述**：[填入 character.md 中的原始文案，用于对照]

#### 1. 主干配置 (Main Definition)
* **SkillID**: `[例如: skill_saintess_heal]`
* **Category**: `[SkillCategory: Normal / Unique / Exclusive / Ultimate]`
* **Type**: `[SkillType: Passive / Startup / Magic / Response]`
* **Timing**: `[TriggerTiming Enum: 例如 TimingActive / TimingOnHitCheck，若为主动技通常为 TimingActive]`

#### 1.1 强制发动配置 (Mandatory Config, 可选)
* **Mandatory.MatchTiming**: `[TriggerTiming: 如 TimingStartup]`
* **Mandatory.ConditionExpression**: `[命中强制锁的表达式: 如 State.IsInSelfTurn(Self.UserID) == true && State.IsJudgmentAtCap(Self.UserID) == true]`
* **Mandatory.LockMode**: `[SkillMandatoryLockMode: SkillMandatoryLockNone / SkillMandatoryLockActionPhaseToSelfSkill]`
* **规则语义**: `[当命中 Mandatory 且 LockMode=SkillMandatoryLockActionPhaseToSelfSkill 时，该角色本行动阶段仅允许提交当前 SkillID]`
* **校验说明**: `[动作锁仅限制可提交动作；技能实际执行仍需通过 Condition/Cost/TargetRule 完整校验]`

#### 1.2 响应分组配置 (Response Group Config, 可选)
* **ResponseGroup.GroupID**: `[string: 响应分组ID；同一触发窗口里需二选一的技能填同一个ID]`
* **ResponseGroup.Mode**: `[ResponseGroupMode: ResponseGroupChooseOne]`
* **ResponseGroup.OptionOrder**: `[int: 弹窗选项排序，越小越靠前]`
* **ReplacesSkillIDs**: `[可选 string 数组: 当本技能被选中时，需要取消执行的技能ID列表]`
* **规则语义**: `[同组且 Mode=ResponseGroupChooseOne 时，引擎层最多允许执行一个；若命中 ReplacesSkillIDs，被替换技能不扣费不执行]`

#### 1.3 规则修饰器模板配置 (RuleModifierTemplate, 可选)
* **ModifierID**: `[string: 全局唯一模板ID]`
* **Domain**: `[RuleModifierDomain: RuleModifierDomainAttribute / RuleModifierDomainHealPolicy / RuleModifierDomainSkillGate / RuleModifierDomainCardSource / RuleModifierDomainTokenPolicy / RuleModifierDomainHealResistPolicy / RuleModifierDomainMoralePolicy]`
* **Priority**: `[int: 优先级，越大越先应用]`
* **ConditionExpression**: `[可选 string: 命中条件；为空表示常驻生效]`
* **StackPolicy**: `[RuleModifierStackPolicy: RuleModifierStackAppend / RuleModifierStackRefreshByModifierID / RuleModifierStackReplaceByDomainPriority]`
* **AttrPayload**: `[可选: AttrType + Operation + ValueSourceMode + (Value / ValueExpression / TokenLink)]`
* **AttrPayload.ValueSourceMode**: `[RuleAttrValueSourceMode: RuleAttrValueFromFixed / RuleAttrValueFromExpression / RuleAttrValueFromTokenLinear]`
* **AttrPayload.Value**: `[当 ValueSourceMode=RuleAttrValueFromFixed 时填写]`
* **AttrPayload.ValueExpression**: `[当 ValueSourceMode=RuleAttrValueFromExpression 时填写动态表达式]`
* **AttrPayload.TokenLink**: `[当 ValueSourceMode=RuleAttrValueFromTokenLinear 时填写 Token 联动：OwnerScope + TokenType + Coefficient + Offset + MinValue + MaxValue]`
* **HealPolicyPayload**: `[可选: ApplyMode + AbsoluteMax]`
* **SkillGatePayload**: `[可选: Mode + SkillIDs]`
* **CardSourcePayload**: `[可选: ProjectionMode + FieldMarks；用于把指定 FieldMark 投影为手牌候选来源]`
* **TokenPolicyPayload**: `[可选: TokenType + ApplyMode + AbsoluteMax；用于“无视指示物上限但绝对封顶N”]`
* **HealResistPolicyPayload**: `[可选: PerDamageWindowHealCap；用于“单次受伤窗口可投入治疗抵伤上限”]`
* **MoralePolicyPayload**: `[可选: ApplyScope + MinMorale + MaxMorale；用于“阵营士气上下限策略（可带生命周期）”]`

#### 2. 前置条件 (Condition Config)
* **PhaseLimit**: `[GamePhase 数组: 例如 [ActionStart, ActionExecution]，若无限制则填 None]`
* **RequireOrientation**: `[CharacterOrientation: Normal / Tapped，无要求填 None]`
* **HandLimit**: `[例如: MaxHandLimit: 3 / MinHandLimit: 2，无要求填 None]`
* **RequireFieldMark**: `[CardFieldMark: 例如 SwordSoul，无要求填 None]`
* **IsTurnLimited**: `[bool: true/false (是否为回合限定)]`
* **RequireSelfAct**: `[bool: 被动技专用，true 表示须由技能持有者本人操作触发]`
* **FieldMarkChangeFilter**: `[当 Timing=TimingOnFieldMarkChanged 时：AcceptBehaviors, AcceptTypesWhenPlaced, AcceptTypesWhenRemoved]`
* **CustomExpression**: `[基于 Player / CombatContext 等动态实体的自定义表达式]`
* **回合伤害去重查询（可选）**: `[可用 State.CountTurnDistinctDamageTargets(sourceUserID, onlyMagic, onlyEnemy) / State.HasTurnDistinctDamageTargetsAtLeast(...) 表达“本回合命中过至少N名不同目标”]`
* **令牌/场标辅助查询（可选）**: `[可用 State.IsTokenAtCap(userID, tokenType) 与 State.CountTeamFieldMark(sourceUserID, mark) 表达“是否达上限/队伍是否拥有某场标”]`
* **已选目标聚合查询（可选）**: `[可用 State.CountSelectedTargetsWithHealAtLeast(Action.Targets, 1) 表达“已选目标中治疗>0的人数”]`
* **回合行动轨迹查询（可选）**: `[可用 State.HasExecutedSpecialActionThisTurn(userID) 表达“本回合是否执行过特殊行动”]`
* **角色姿态查询（可选）**: `[可用 State.GetPlayerOrientation(userID) 判定任意角色当前是否横置（Tapped）]`
* **阵营/队友数量查询（可选）**: `[可用 State.IsSameTeam(userA, userB) 与 State.GetAliveTeammateCount(userID) 进行同阵营判定与“队友数>1”等限制]`

#### 3. 目标选择规则 (Target Rule Config)
* **SelectType**: `[TargetSelectType: None / Self / Teammate / TeamOther / Enemy / EnemyOther / Any]`
* **MinCount**: `[int]`
* **MaxCount**: `[int]`
* **RequireTargetAllocations**: `[bool: 是否需要客户端提交 Action.TargetAllocations]`
* **AllocationTotal**: `[可选 int: 分配总值，如 3]`
* **MinAllocationPerTarget**: `[可选 int: 单目标最小分配值]`
* **MaxAllocationPerTarget**: `[可选 int: 单目标最大分配值]`
* **AllowedActionRefs**: `[可选 ActionType 数组: 如 [Attack, Magic]]`
* **NamedValueConstraints**: `[可选 NamedValueConstraint 数组: 约束 Action.NamedValues（多变量输入，如 X/Y）]`
* **SelectedValueRule**: `[可选；用于 X 值技能，示例: 1 <= Action.SelectedValue <= Event.PendingMoraleLoss]`
* **批量弃牌结果上下文（可选）**: `[可读取 Event.DiscardedMagicCount 与 Event.DiscardedElementCount[Water/Fire/...]; 由同链路内的 EffectDiscard 批量弃牌后写入]`
* **爆牌弃牌上下文（可选）**: `[可读取 Event.OverflowDiscardOwnerID / Event.OverflowDiscardCardIDs / Event.OverflowDiscardCount；用于“因超手牌上限弃牌后再留场”类技能]`
* **Filters**: `[附加过滤条件: 如 RequireHeal >= 1, RequireStatus == "WindHole"]`
* **SubmitAction.NamedValues**: `[当使用 NamedValueConstraints 时，客户端需提交命名数值字典，如 {\"X\":3, \"Y\":2}]`

#### 3.1 行动改写配置 (Action Transform Config, 可选)
* **ActionTransform.Hook**: `[通常为 TimingBeforeActionExecute]`
* **ActionTransform.Optional**: `[bool: true=可选发动；false=满足条件自动改写]`
* **ActionTransform.Priority**: `[int: 多个改写命中时的优先级]`
* **ActionTransform.CancelCurrentAction**: `[bool: 是否取消当前行动默认结算（如替代 Buy 的原生流程）]`
* **ActionTransform.Match.RequireActionType**: `[可选 ActionType: 例如 Magic]`
* **ActionTransform.Match.RequirePlayedCardTypes**: `[可选 CardType 数组]`
* **ActionTransform.Match.RequirePlayedCardElements**: `[可选 ElementType 数组，例如 [Earth, Fire]]`
* **ActionTransform.Match.ExcludeTemplateIDs**: `[可选 string 数组，例如 [\"card_attack_dark_annihilation\"]]`
* **ActionTransform.Rewrite**: `[可选 ActionRewriteConfig；nil 表示仅取消当前行动并继续执行 Effects]`
* **ActionTransform.Rewrite.FlowRef**: `[ActionFlowType: ActionFlowNormalCombat / ActionFlowMagicBulletChain]`
* **ActionTransform.Rewrite.ActionTypeRef**: `[可选 ActionType: 需要同时改写 ActionType 时填写]`
* **ActionTransform.Rewrite.ExecuteImmediately**: `[bool: 是否立即按改写后的行动结算（不是追加一次行动）]`
* **ActionTransform.Rewrite.TreatAsActiveAttack**: `[bool: 改写为 Attack 时是否标记为主动攻击]`
* **ActionTransform.Rewrite.ElementPickMode**: `[RewriteElementPickMode: RewriteElementNone / RewriteElementFixed / RewriteElementFromActionRef]`
* **ActionTransform.Rewrite.FixedElementRef**: `[可选 ElementType: 当 ElementPickMode=RewriteElementFixed 时填写]`
* **SubmitAction.ElementRef**: `[当 ElementPickMode=RewriteElementFromActionRef 时，客户端必须提交 ElementRef]`

#### 4. 费用消耗 (Cost Config)
* **CardPlayCostType**: `[CardPlayCostType: CardPlayNotRequired / CardPlayRequired]`
* **Stones**: `[格式: {Type: StarStoneType, Amount: int}，例如 {Type: Gem, Amount: 1}]`
* **Tokens**: `[格式: {Type: TokenType, Amount: int}，例如 {Type: Rage, Amount: 1}]`
* **HealCost**: `[int: 需要移除的治疗点数]`
* **HPCost**: `[int: 对自己造成的法术伤害值]`
* **Discards**:
* `Count`: `[int: 需要弃牌的数量]`
* `Filter`: `[格式: ReqCardType / ReqElement / ReqDestiny / SameAttribute，例如: SameAttribute: MatchElement]`
* `Visibility`: `VisibilityPublic`  (卡牌是否可见)

#### 5. 战斗与结算劫持标记 (Combat Intercept Tags)
* **Tags**: `[CombatInterceptTag Enums, 用逗号分隔。例如: Unrespondable, IgnoreHolyShield，若无则填 None]`

#### 6. 执行效果序列 (Effects)
*(挂载到 SkillDefinition.Effects 数组)*
* **Effect[0]**:
* `EffectType`: `[EffectType Enum: 如 EffectDamage, EffectHeal, EffectAddAction]`
* `Target`: `[EffectTargetType Enum: 如 TargetSelected, TargetSelf, TargetAllOthers, TargetAllExceptSelected]`
* `Value`: `[数值或基于上下文的动态表达式，例如: 2 或 Player.Tokens[Rage]]`
* `Visibility`: `VisibilityPublic` (必须向全场展示)
* `Ref`: `[指向 TokenType 或 StatusEffect 的关联，例如 TokenRef: Rage，无则填 None]`
* `ElementRef`: `[可选 ElementType: 用于系别限制；当 EffectType=EffectSetCurrentCombatElement 时，表示改写后的战斗系别（必填）；当 EffectType=EffectRemoveFieldMark 时可按系别过滤移除]`
* `StoneRef`: `[可选 StarStoneType: 当 EffectType=EffectAddTeamStone / EffectAddEnergyStone / EffectConvertTeamStone / EffectConvertEnergyStone 时填写 Gem/Crystal/Any 或源颜色]`
* `StoneToRef`: `[可选 StarStoneType: 当 EffectType=EffectConvertTeamStone / EffectConvertEnergyStone 时填写目标颜色 Gem/Crystal]`
* `FromTargetRef`: `[可选 EffectTargetType: 当 EffectType=EffectTransferTeamStone / EffectTransferCard / EffectTransferFieldMark 时填写来源目标；EffectTransferCard 不填时兼容旧语义 Target->Self]`
* `FieldMarkRef`: `[可选 CardFieldMark: 当 EffectType=EffectRemoveFieldMark / EffectPlaceDeckTopAsFieldMark / EffectPlacePlayedCardAsFieldMark / EffectPlaceHandCardAsFieldMark / EffectTransferFieldMark 时填写，如 Blessing]`
* `VisibilityRef`: `[可选 CardVisibilityType: 当 EffectType=EffectPlacePlayedCardAsFieldMark / EffectPlaceHandCardAsFieldMark 时填写 VisibilityPublic / VisibilityHidden]`
* `OrientationRef`: `[可选 CharacterOrientation: 当 EffectType=EffectSetOrientation 时填写 Normal/Tapped]`
* `FormRef`: `[可选 string: 当 EffectType=EffectSetForm 时填写形态名；清空形态可填 nil]`
* `BranchRef`: `[可选 PerTargetBranchConfig: 当 EffectType=EffectPerTargetBranch 时填写逐目标响应分支]`
* `RuleModifierRef`: `[可选 string: 当 EffectType=EffectApplyRuleModifier 时，引用 RuleModifierTemplate.ModifierID]`
* `RuleLifetimeRef`: `[可选 RuleModifierLifetimeType: 当 EffectType=EffectApplyRuleModifier 时，填写 RuleLifeThisEffectChain / RuleLifeUntilTurnEnd / RuleLifeUntilSourceNextTurnStart / RuleLifeUntilSourceNextTurnEnd / RuleLifePermanent / RuleLifeUntilCombatEnd]`
* `RuleRemoveRef`: `[可选 RuleModifierRemoveQuery: 当 EffectType=EffectRemoveRuleModifier 时，填写移除条件]`
* `备注`: `[EffectHeal 的 Value 可为负数（用于移除治疗）；EffectAddEnergyStone 的 Value 可为负数（用于移除能量，移除不足时按可移除量结算且不阻断）；EffectAddTeamStone 的 Value 可为负数（用于移除战绩区星石，移除不足时按可移除量结算且不阻断）；EffectAddTeamCup 用于增减目标阵营星杯（结果钳制在 [0, TargetStarCups]）；EffectAddToken 的 Value 可为负数（用于移除指示物，移除不足时按可移除量结算且不阻断）；英灵人形“翻转战纹/魔纹”建议使用两条 EffectAddToken（源类型负值移除 + 目标类型正值增加）；EffectApplyCombatTag 的 Value 应填写 CombatInterceptTag 枚举标识名（如 Unrespondable、ForceHit），不建议使用裸数字；EffectRemoveFieldMark 用于移除面前场标资源（如 Blessing，若 ElementRef 非空则按系别过滤），并刷新 `Event.RemovedFieldCard*` 最近移除快照；EffectPlaceDeckTopAsFieldMark 用于把牌库顶 N 张牌放到角色面前作为场标资源；EffectPlacePlayedCardAsFieldMark 用于把“本次打出的牌实体”留场并标记为 FieldMark；EffectPlaceHandCardAsFieldMark 用于把目标手牌中的指定数量牌放置到角色面前并标记为 FieldMark（通常由 Action.UsedCardUUIDs 指定具体牌）；EffectTransferFieldMark 用于迁移“已在场”的场标实体（FromTargetRef -> Target，不是移除重放）；EffectRemoveSelectedFieldCard 用于按 SubmitAction 的 `SelectedFieldCards` 精确移除场上牌，同样会刷新 `Event.RemovedFieldCard*` 最近移除快照；EffectRevealRemovedFieldCard 用于公开展示最近一次被移除的场上牌；EffectPlaceOverflowDiscardAsFieldMark 用于将当前爆牌弃置列表（`Event.OverflowDiscardCardIDs`）转为目标角色面前场标（`Value<=0` 表示全部）；EffectGrantExtraTurn 用于在当前回合结算后为目标角色追加额外回合（`Value<=0` 按1处理）；TargetAllExceptSelected 用于“除所选目标外其他所有角色”；EffectConvertEnergyStone 用于角色能量区颜色转换；EffectModifyPendingDamage 用于统一修饰当前结算链待生效伤害（攻/法通用，最终不低于0）；EffectChangeMorale 用于直接改士气；EffectReducePendingMoraleLoss 用于修改本次结算窗口的待扣士气；EffectSetHandLimitFixed 的 Value 表示固定手牌上限；EffectTransferCard 支持显式方向（FromTargetRef->Target，不填则兼容 Target->Self）；EffectSetCurrentCombatElement 用于改写当前战斗系别（ElementRef 必填）；EffectRedirectCurrentCombatTarget 用于改写当前战斗承受者（Target=新目标）；EffectSetCurrentCounterExecutor 用于改写当前战斗应战执行者（Target=新执行者，通常与目标改写同链路）；EffectTransferTeamStone/EffectConvertTeamStone/EffectRedirectCurrentExtractOutput 用于跨队星石转移、颜色转换、提炼产出重定向；与“动态上限/无视上限治疗/禁用技能/场标视为手牌/指示物上限策略/单次受伤窗口治疗抵伤上限/阵营士气上下限策略”相关能力统一通过 RuleModifier + EffectApply/Remove 表达]`
* **Effect[1]**: ...

#### 6.1 逐目标响应分支 (PerTargetBranchConfig, 可选)
* **TargetSource**: `[PerTargetSourceType: 目前支持 PerTargetSelectedTargets]`
* **InterruptType**: `[InterruptType: WaitDiscard / WaitChoice]`
* **TimeoutAsDeclined**: `[bool: 超时是否按失败分支处理]`
* **DiscardRequirement**: `[可选: Count + Filter；用于 WaitDiscard]`
* **DiscardVisibility**: `[可选 CardVisibilityType: VisibilityPublic / VisibilityHidden]`
* **OnSuccess**: `[EffectNode 数组: 目标响应成功时执行]`
* **OnDeclined**: `[EffectNode 数组: 目标响应失败/拒绝时执行]`

#### 7. 延后结算配置 (Status Resolve Config, 可选)
* **StatusResolveConfigID**: `[可选；若技能只负责“放置状态牌”，真正效果在未来时点触发，则填写对应状态结算配置ID]`
* **CanDecline**: `[可选；bool。false 表示触发条件满足后必须结算，不能放弃]`
* **EnforceNextActionMustActiveAttackSource**: `[可选；bool。轻量动作锁（挑衅专用）：持有者下个行动阶段必须主动攻击状态来源，否则跳过该行动阶段]`
* **示例**: `status_resolve_five_elements_bind`
```
