// gameflow: 玩家角色入口定义。

package player

import (
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// CombatPolicyType 战斗策略类型（用于 TimingHook 分派）。
type CombatPolicyType string

const (
	CombatPolicyDefendValidation    CombatPolicyType = "defend_validation"
	CombatPolicyMagicMissileDefend  CombatPolicyType = "magic_missile_defend"
	CombatPolicyMagicMissileCounter CombatPolicyType = "magic_missile_counter"
)

// CombatPolicyRuntime 战斗策略运行时接口（用于 TimingHook 中需要战斗上下文的场景）。
type CombatPolicyRuntime interface {
	HookRuntime
	GetCombatRequest() *model.CombatRequest
	GetMagicBulletChain() *model.MagicBulletChain
}

// CombatPolicyContext 战斗策略上下文。
type CombatPolicyContext struct {
	Player        *model.Player
	CombatRequest *model.CombatRequest
	Chain         *model.MagicBulletChain
	Card          model.Card
	CounterCard   *model.Card
	UseFaction    bool
}

// InterruptSpec 定义角色包贡献的中断处理条目。
type InterruptSpec struct {
	Type                 model.InterruptType
	BuildPrompt          func(rt ChoiceRuntime) *model.Prompt
	HandleAction         func(rt ChoiceRuntime, act model.PlayerAction) error
	AllowedActionTypes   []model.PlayerActionType
	InvalidActionMessage string
}

// ChoiceSpec 定义声明式选择流程条目。
type ChoiceSpec struct {
	ChoiceType   string
	BuildPrompt  func(rt ChoiceRuntime, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt
	HandleChoice func(rt ChoiceRuntime, playerID string, selectionIndex int, data map[string]interface{}) (bool, error)
}

// ChoiceRuntime 抽象角色选择流运行时能力。
type ChoiceRuntime interface {
	model.IGameEngine
	LookupPlayer(playerID string) *model.Player
	AllPlayers() map[string]*model.Player
	PlayerOrder() []string
	PopInterrupt()
	HasPendingInterrupt() bool
	NotifyInterruptPrompt()
	DrawCardsDirect(playerID string, amount int, reason string)
	DrawRawCards(amount int) ([]model.Card, bool)
	EnsureCombatInteractionWindow()
	ReplacePendingInterruptContext(data map[string]interface{}) error
	ReplacePendingInterruptPlayerID(playerID string)
	ResumePendingAttackMiss(ctx *model.Context) bool
	ResumePendingAttackHit(ctxData map[string]interface{})
	ApplyChoiceResumePoint(raw interface{})
	RoutePendingDamageOr(defaultReturn interface{}, onNoPending func()) bool
	RoutePendingDamageWithReturn(returnTo interface{}) bool
	EnterExtraActionStage()
	EnterTurnEndStage()
	EnterDamageResolution(returnTo interface{})
	EnterActionExecutionStage()
	AllOtherPlayerIDs(userID string) []string
	PlayerOrderPosition(playerID string) int
	StartDraw(ctx *model.Context)
	NewDrawContext(player *model.Player, amount int, reason string) *model.Context
	RestorePhaseAfterInterruptedDraw(ctx *model.Context) bool
	PushInterrupt(intr *model.Interrupt)
	PendingDamageQueueLen() int
	GetPendingDamage(index int) (*model.PendingDamage, bool)
	ActionQueueLen() int
	AttachExclusiveEffectCard(sourceID, targetID string, effect model.EffectType, card model.Card) error
	ResumePendingMoraleLoss(ctx *model.Context) bool
	EnterResponseWindow()
	ApplyStealthEffect(player *model.Player)
	EnqueueVirtualAttack(sourceID, targetID string, card model.Card, sourceSkill string)
	ApplyCampMoraleLoss(camp model.Camp, wantLoss int) int
	ResolveCounterAttack(counterPlayerID, counterTargetID string, counterCard model.Card)
	NotifyCombatCue(attackerID, targetID, cueType string)
	ConsumePlayableCardByCardID(playerID, cardID string) (model.Card, bool)
	TopCombatRequest() *model.CombatRequest
	PopCombatRequest()
	PendingInterrupt() *model.Interrupt
	RoutePendingDamageWithDefaultReturn(defaultReturn interface{}) bool
	RestoreReturnPoint() bool
	PushDiscardChoiceInterrupt(playerID string, data map[string]interface{})
	EnterActionEndStage()
	MagicBulletChain() *model.MagicBulletChain
	SetMagicBulletChain(chain *model.MagicBulletChain)
	SetReturnPoint(returnTo interface{})
	GetPlayableCardByIndex(player *model.Player, idx int) (model.Card, bool)
	ConsumePlayableCardByIndex(player *model.Player, idx int) (model.Card, error)
	PerformMagic(playerID, targetID string, cardIdx int, isFusion bool) error
	ExecuteMagicBullet(player *model.Player, reverse, isFusion bool, fusionCard *model.Card) error
	FindNextMagicBulletTarget(playerID string) string
	DispatchHitCheckMagicMissileCounter(player *model.Player, chain *model.MagicBulletChain, card *model.Card) error
	DispatchHitCheckMagicMissileDefend(player *model.Player, chain *model.MagicBulletChain) error
	AddToDiscardPile(cards ...model.Card)
	CheckGameEnd()
	CampEnemyIDs(camp model.Camp) []string
	CampMorale(camp model.Camp) int
	AddCampMorale(camp model.Camp, amount int) int
	PendingDiscardVictimID() string
	NotifyCardHidden(playerID string, cards []model.Card, actionType model.DamageType)
	MarkPendingAttackDamageHitProcessed(ctx *model.Context) bool
	SyncGamePhaseWithInterrupt(intr *model.Interrupt)
	SnapshotPlayerPoses() map[string]PoseSnapshot
	DispatchOrientationChanges(before map[string]PoseSnapshot)
	FindSourceEffectCard(source *model.Player, effect model.EffectType) (*model.Player, *model.FieldCard)
	AttachSourceEffectCard(source, target *model.Player, effect model.EffectType, card model.Card) error
	DetachSourceEffectCard(source *model.Player, effect model.EffectType) (*model.Player, model.Card, bool)
	FindExclusiveEffectCard(source *model.Player, effect model.EffectType) (*model.Player, *model.FieldCard)
	DetachExclusiveEffectCard(source *model.Player, effect model.EffectType) (*model.Player, model.Card, bool)
	RemoveExclusiveEffectCard(source *model.Player, effect model.EffectType, restoreCard bool) bool
	EmitBuffRemovedDispatch(sourceID, targetID string, effect model.EffectType)
	InitCombat(attackerID, targetID string, card *model.Card, isForcedHit, canBeResponded, ignoreShield bool, interceptTags map[model.CombatInterceptTag]bool, isCounter ...bool)
	ResolveMagicBowPierceMiss(attackerID, targetID string, attackCard *model.Card, isCounter bool)
	HasFixedMaxHandCap(player *model.Player) bool
	HasMercyFixedMaxHandCap(player *model.Player) bool
	RoleFixedMaxHandCapValue(player *model.Player) (int, bool)
	RefreshAllPlayerDerivedStates()
	SyncHolyLancerRevelationMaxHeal(player *model.Player)
	BuildContext(user, target *model.Player, timing model.FlowTiming, eventCtx *model.EventContext) *model.Context
	TakeDiscardPileCardByID(cardID string) (model.Card, bool)
	IsSkillStillUsable(skillID string, user *model.Player, ctx *model.Context) bool
	CurrentTurnPlayerID() string
	RecordMagicDamageTarget(sourceID, targetID string)
	MagicDamageTargetCount(sourceID string) int
	RecordSkillUsage(playerID, title string, skillType model.SkillType)
	IsActionSkillUsableForExtraMagic(player *model.Player, skillDef model.SkillDefinition) bool
	FighterLockedTarget(player *model.Player) *model.Player
	ClearFighterHundredDragon(player *model.Player, logLine string) bool
	CanCastMagicInAction(player *model.Player) bool
}

// ChoiceHandler 抽象角色选择流入口。
type ChoiceHandler interface {
	BuildPrompt(rt ChoiceRuntime, choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt
	HandleChoice(rt ChoiceRuntime, playerID string, selectionIndex int, ctxData map[string]interface{}) (bool, error)
}

// CancelChoiceHandler 为支持取消的角色选择流提供扩展入口。
type CancelChoiceHandler interface {
	HandleCancel(rt ChoiceRuntime, playerID string, ctxData map[string]interface{}) (bool, error)
}

// SkillEntry 定义角色技能与策略绑定入口。
type SkillEntry struct {
	ID      string
	Handler model.SkillHandler
	Policy  types.SkillPolicy
}

// SkillUsabilityChecker 技能可用性检查器。
type SkillUsabilityChecker func(engine SkillUsabilityCheckerEngine, player *model.Player, skillDef model.SkillDefinition) bool

// SkillUsabilityCheckerEngine 抽象技能可用性检查所需的引擎能力。
type SkillUsabilityCheckerEngine interface {
	LookupPlayer(playerID string) *model.Player
	PlayerOrder() []string
	GetAllPlayers() []*model.Player
	HasForm(player *model.Player, form string) bool
	IsCharacter(player *model.Player, roleID string) bool
	GetToken(player *model.Player, key string) int
	CountCoverCardsByEffectAndElement(player *model.Player, effect model.EffectType, element model.Element) int
}

// FollowupHost 抽象延迟后续回调可用的最小引擎能力。
type FollowupHost interface {
	Log(message string)
	LookupPlayer(playerID string) *model.Player
	ResolveSkillFollowup(req ResolveSkillFollowupReq) error
}

// FollowupSpec 定义角色模块贡献到延迟后续执行表的条目。
type FollowupSpec struct {
	Label   string
	Resolve func(host FollowupHost, f model.DeferredFollowup) error
}

// ResolveSkillFollowupReq 是角色 followup resolver 发起引擎后续结算的通用请求。
type ResolveSkillFollowupReq struct {
	Kind     string
	Followup model.DeferredFollowup
}

// RoleEntry 表示单个角色在 player 子目录的统一入口定义。
type RoleEntry struct {
	ID                         string
	Defaults                   func(player *model.Player)
	StarterCards               func(player *model.Player) []model.Card
	HandLimit                  HandLimitRule
	MaxHeal                    func(player *model.Player, current int) int
	Choices                    ChoiceHandler
	ChoiceSpecs                []ChoiceSpec
	Skills                     []SkillEntry
	ChoiceRouteSpecs           map[string]types.ChoiceRouteSpec
	FollowupSpecs              map[string]FollowupSpec
	InterruptSpecs             []InterruptSpec
	TimingHookSpecs            []TimingHookSpec
	PolicySpecs                []PolicySpec // 角色策略声明
	SkillUsabilityCheckers     map[string]SkillUsabilityChecker
	AttackCardElementTransform func(player *model.Player, card model.Card) model.Element
}

// ApplyDefaults 应用默认角色属性。
func (e RoleEntry) ApplyDefaults(player *model.Player) {
	if e.Defaults == nil {
		return
	}
	e.Defaults(player)
}

// ApplyStarterCards 应用开局专属牌。
func (e RoleEntry) ApplyStarterCards(player *model.Player) []model.Card {
	if e.StarterCards == nil {
		return nil
	}
	return e.StarterCards(player)
}

// HandLimitRule 返回角色手牌规则。
func (e RoleEntry) HandLimitRule() HandLimitRule {
	if e.HandLimit == nil {
		return DefaultHandLimitRule
	}
	return e.HandLimit
}

// ApplyMaxHeal 应用治疗上限规则。
func (e RoleEntry) ApplyMaxHeal(player *model.Player, current int) int {
	if e.MaxHeal == nil {
		return current
	}
	return e.MaxHeal(player, current)
}

// BuildChoicePrompt 构建角色选择提示。
func (e RoleEntry) BuildChoicePrompt(rt ChoiceRuntime, choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	if e.Choices == nil {
		return nil
	}
	return e.Choices.BuildPrompt(rt, choiceType, playerID, player, data)
}

// HandleChoice 处理角色选择输入。
func (e RoleEntry) HandleChoice(rt ChoiceRuntime, playerID string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	if e.Choices == nil {
		return false, nil
	}
	return e.Choices.HandleChoice(rt, playerID, selectionIndex, ctxData)
}

// HandleChoiceCancel 处理角色选择流的取消事件。
func (e RoleEntry) HandleChoiceCancel(rt ChoiceRuntime, playerID string, ctxData map[string]interface{}) (bool, error) {
	if e.Choices == nil {
		return false, nil
	}
	handler, ok := e.Choices.(CancelChoiceHandler)
	if !ok {
		return false, nil
	}
	return handler.HandleCancel(rt, playerID, ctxData)
}

// ChoiceRoutes 返回角色 choice 路由声明。
func (e RoleEntry) ChoiceRoutes() map[string]types.ChoiceRouteSpec {
	return e.ChoiceRouteSpecs
}

// SkillPolicies 返回角色技能策略声明。
func (e RoleEntry) SkillPolicies() map[string]types.SkillPolicy {
	if len(e.Skills) == 0 {
		return nil
	}
	out := make(map[string]types.SkillPolicy, len(e.Skills))
	for _, skill := range e.Skills {
		if skill.ID != "" {
			out[skill.ID] = skill.Policy
		}
	}
	return out
}

// Followups 返回角色延迟后续声明。
func (e RoleEntry) Followups() map[string]FollowupSpec {
	return e.FollowupSpecs
}
