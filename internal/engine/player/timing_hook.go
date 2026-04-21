// gameflow: Timing Hook 声明式注册类型。

package player

import "starcup-engine/internal/model"

// TimingPoint 标识 Hook 触发时机。
type TimingPoint string

const (
	// 已有 - 行动后/攻击后/伤害后
	TimingPostActionEnd      TimingPoint = "post_action_end"
	TimingPostAttackHit      TimingPoint = "post_attack_hit"
	TimingPostDamageResolved TimingPoint = "post_damage_resolved"

	// 新增 - 回合阶段
	TimingOnTurnStart    TimingPoint = "on_turn_start"     // 回合开始（形态检查、状态清理）
	TimingOnTurnEnd      TimingPoint = "on_turn_end"       // 回合结束前置（形态释放，额外行动前）
	TimingOnTurnEndFinal TimingPoint = "on_turn_end_final" // 回合结束最终（额外行动耗尽后）
	TimingBeforeAction   TimingPoint = "before_action"     // 行动前（场上效果检查）
	TimingOnActionEnd    TimingPoint = "on_action_end"     // 行动结束（非 post_action_end，用于技能后）

	// 新增 - 攻击阶段
	TimingOnAttackDeclared   TimingPoint = "on_attack_declared"    // 攻击宣告时
	TimingOnAttackGating     TimingPoint = "on_attack_gating"      // 攻击门控检查
	TimingOnAttackCardHook   TimingPoint = "on_attack_card_hook"   // 攻击卡牌变换
	TimingOnAttackStateReset TimingPoint = "on_attack_state_reset" // 攻击状态重置
	TimingOnAttackTargetCtx  TimingPoint = "on_attack_target_ctx"  // 目标上下文记录
	TimingOnAttackMiss       TimingPoint = "on_attack_miss"        // 攻击未命中后

	// 新增 - 命中判定阶段
	TimingOnHitCheck         TimingPoint = "on_hit_check"          // 命中判定
	TimingOnCounterPolicy    TimingPoint = "on_counter_policy"     // 反击策略
	TimingOnDefendValidation TimingPoint = "on_defend_validation"  // 防御验证
	TimingOnResponseSkillAug TimingPoint = "on_response_skill_aug" // 响应技能增强/规范化

	// 新增 - 伤害阶段
	TimingOnDamageCalculate   TimingPoint = "on_damage_calculate"    // 伤害计算（被动增伤）
	TimingOnDamageBeforeApply TimingPoint = "on_damage_before_apply" // 伤害应用前（蝴蝶舞者等）
	TimingOnDamageAfterApply  TimingPoint = "on_damage_after_apply"  // 伤害应用后（灵魂链接等）
	TimingOnHealResist        TimingPoint = "on_heal_resist"         // 治愈抵抗规则

	// 新增 - 特殊阶段
	TimingOnGameStart   TimingPoint = "on_game_start"   // 游戏开始
	TimingOnPlayerAdded TimingPoint = "on_player_added" // 玩家加入
	TimingOnCampChanged TimingPoint = "on_camp_changed" // 阵营变化
)

// TimingHookSpec 角色贡献到全局 timing hook 链的条目。
type TimingHookSpec struct {
	Timing   TimingPoint
	Priority int // 数值越小越先执行
	Hook     TimingHookFunc
}

// TimingHookContext 传递给 Hook 的上下文。
type TimingHookContext struct {
	// 基础字段
	SourceID      string
	TargetID      string
	ActionType    model.ActionType     // post_action_end, attack hooks
	DamageType    model.DamageType     // post_damage_resolved / post_attack_hit
	Damage        int                  // post_damage_resolved
	IsCounter     bool                 // post_attack_hit
	Card          *model.Card          // attack/damage hooks
	PendingDamage *model.PendingDamage // 原始 PD（可选）

	// 攻击未命中标记
	ForceHeroRoarMiss      bool // 强制触发英雄怒吼未命中
	ForceFighterChargeMiss bool // 强制触发格斗家蓄力未命中

	// 新增 - 战斗上下文
	CombatStage      string                 // "attack_declared", "hit_check", "damage_calculated"
	CounterInitiator string                 // 反击发起者ID
	RespondingPlayer string                 // 响应玩家ID
	AttackAction     *model.PlayerAction    // 原始攻击行动（攻击门控需要）
	AttackInfo       *model.AttackEventInfo // 攻击劫持上下文（用于设置拦截标签）

	// 新增 - 技能上下文
	SkillID   string // 技能 ID
	SkillCost int    // 技能消耗（gem/crystal）

	// 新增 - 选择上下文
	ChoiceType     string // 选择类型
	SelectionIndex int    // 选择索引

	// 新增 - 回合上下文
	TurnPlayerID string          // 当前回合玩家ID
	TurnStage    model.TurnStage // 回合阶段
}

// TimingHookResult Hook 执行结果。
type TimingHookResult struct {
	Interrupted  bool   // true = 产生了中断，状态机应暂停
	Blocked      bool   // true = 攻击门控阻止（攻击不允许执行）
	BlockReason  string // 阻止原因（用于错误消息）
	SkipNextHook bool   // true = 跳过后续钩子（用于特殊阻断）
	DamageDelta  int    // 伤害修正值（正=增伤，负=减伤），用于被动增伤链
}

// TimingHookFunc 统一 Hook 签名。
type TimingHookFunc func(rt HookRuntime, ctx TimingHookContext) TimingHookResult

// HookRuntime 抽象 Timing Hook 运行时能力（窄接口）。
type HookRuntime interface {
	// 基础方法（已有）
	Log(message string)
	GetPlayer(playerID string) *model.Player
	PushInterrupt(intr *model.Interrupt)
	PushDiscardChoiceInterrupt(playerID string, data map[string]interface{})
	Heal(targetID string, amount int)
	AddPendingDamage(pd model.PendingDamage)
	GetMaxHand(player *model.Player) int
	GetPlayerEnergyCap(player *model.Player) int
	DrawCards(playerID string, amount int)
	GetPendingDamageQueue() []model.PendingDamage
	SetPendingDamageQueue(queue []model.PendingDamage)
	SnapshotPlayerPoses() map[string]PoseSnapshot
	DispatchOrientationChanges(before map[string]PoseSnapshot)
	CampMorale(camp model.Camp) int
	HasPendingDiscardFor(playerID string) bool
	PlayerOrder() []string
	RecordMagicDamageTarget(sourceID, targetID string)
	MagicDamageTargetCount(sourceID string) int
	LookupPlayer(playerID string) *model.Player
	CurrentTurnPlayerID() string
	BuildContext(user, target *model.Player, timing model.FlowTiming, eventCtx *model.EventContext) *model.Context
	IsSkillStillUsable(skillID string, user *model.Player, ctx *model.Context) bool
	IsMagicDamageType(damageType model.DamageType) bool

	// 新增 - 角色身份检查
	IsCharacter(player *model.Player, roleID string) bool

	// 新增 - 形态操作
	HasForm(player *model.Player, form string) bool
	SetForm(player *model.Player, form string) bool
	ClearForm(player *model.Player, form string) bool

	// 新增 - 指示物操作
	GetToken(player *model.Player, key string) int
	SetToken(player *model.Player, key string, value int)

	// 新增 - 阵营士气（扩展）
	AddCampMorale(camp model.Camp, amount int) int

	// 新增 - 战斗上下文
	GetPendingDamage() *model.PendingDamage
	GetPendingDamageByIndex(index int) (*model.PendingDamage, bool)
	SetPendingDamageDamage(pd *model.PendingDamage, damage int)
	GetCurrentCombat() *model.CombatRequest
	PopCombatRequest()

	// 新增 - 卡牌操作
	GetCardByIndex(player *model.Player, idx int) (model.Card, bool)
	ConsumeCardByIndex(player *model.Player, idx int) (model.Card, error)
	AddToDiscardPile(cards ...model.Card)
	TakeDiscardPileCardByID(cardID string) (model.Card, bool)

	// 新增 - 中断推送（扩展）
	PushInterruptForPlayer(playerID string, intr *model.Interrupt)

	// 新增 - 回合状态
	GetTurnStage() model.TurnStage
	SetTurnStage(stage model.TurnStage)
	IsPlayerActive(player *model.Player) bool

	// 新增 - 状态机控制
	EnterResponseWindow()
	EnterActionExecutionStage()
	EnterDamageResolution(returnTo interface{})

	// Phase C - 回合/攻击 hooks 扩展
	CampEnemyIDs(camp model.Camp) []string
	RemoveExclusiveEffectCard(source *model.Player, effect model.EffectType, restoreCard bool) bool
	CheckHandLimit(player *model.Player)
	HasPendingInterrupt() bool
	CanPayCrystalCost(playerID string, amount int) bool
	DrawCardsWithOptions(playerID string, count int, opts model.DrawOptions)
	ConsumeBeastSoul(player *model.Player, amount int) int
	ClearHundredDragonForm(player *model.Player, logLine string)
	ReleaseAssassinStealth(player *model.Player)
	ResolveCrimsonKnightHotFormTurnEnd(player *model.Player) bool
	NotifyCardRevealed(playerID string, cards []model.Card, actionType model.DamageType)
	GetAllPlayers() []*model.Player
	GetAllPlayersMap() map[string]*model.Player
	HasUsedActionSkill(player *model.Player) bool

	// 被动增伤/伤害计算扩展
	ConsumeAttackDamageRuleBonus(player *model.Player, modifierID string, action model.Action) int
	GetPlayerOrientation(player *model.Player) model.CharacterOrientation
}

// PoseSnapshot 记录玩家姿态快照（用于 orientation 变更前后对比）。
type PoseSnapshot struct {
	Orientation model.CharacterOrientation
	Form        string
}
