// gameflow: 策略钩子声明式注册类型（回合钩子、行动选择策略、战斗策略、响应技能策略等）。

package player

import (
	"starcup-engine/internal/model"
)

// ---- 回合时序钩子 ----

// TurnTimingPoint 标识回合阶段钩子触发点。
type TurnTimingPoint string

const (
	TurnBeforeStart      TurnTimingPoint = "before_start"
	TurnStart            TurnTimingPoint = "start"
	TurnBeforeActionExec TurnTimingPoint = "before_action_execute"
	TurnEndPreExtra      TurnTimingPoint = "end_pre_extra"
	TurnEndFinal         TurnTimingPoint = "end_final"
)

// TurnHookFunc 回合时序钩子签名。
type TurnHookFunc func(rt ChoiceRuntime, player *model.Player) bool

// TurnHookSpec 角色贡献到回合时序钩子链的条目。
type TurnHookSpec struct {
	Timing   TurnTimingPoint
	Priority int
	Hook     TurnHookFunc
}

// ---- 攻击宣言中断钩子 ----

// AttackDeclaredHookFunc 攻击宣言阶段中断钩子签名。
type AttackDeclaredHookFunc func(rt ChoiceRuntime, attacker *model.Player, target *model.Player, sourceSkill string, card *model.Card, userCtx *model.Context) bool

// AttackDeclaredHookSpec 攻击宣言中断条目。
type AttackDeclaredHookSpec struct {
	Priority int
	Hook     AttackDeclaredHookFunc
}

// ---- 行动结束中断钩子 ----

// ActionEndHookFunc 行动结束阶段中断钩子签名。
type ActionEndHookFunc func(rt ChoiceRuntime, ctx *model.Context) bool

// ActionEndHookSpec 行动结束中断条目。
type ActionEndHookSpec struct {
	Priority int
	Hook     ActionEndHookFunc
}

// ---- 响应技能增广/规范化 ----

// ResponseSkillAugmentFunc 响应技能列表增广函数签名。
type ResponseSkillAugmentFunc func(rt ChoiceRuntime, skillIDs []string, ctx *model.Context) []string

// ResponseSkillAugmentSpec 响应技能增广条目。
type ResponseSkillAugmentSpec struct {
	Priority int
	Augment  ResponseSkillAugmentFunc
}

// ResponseSkillNormalizeFunc 响应技能列表规范化函数签名。
type ResponseSkillNormalizeFunc func(rt ChoiceRuntime, skillIDs []string, ctx *model.Context) []string

// ResponseSkillNormalizeSpec 响应技能规范化条目。
type ResponseSkillNormalizeSpec struct {
	Priority  int
	Normalize ResponseSkillNormalizeFunc
}

// ---- 战斗交互策略 ----

// CombatInteractionFunc 战斗交互钩子签名。
type CombatInteractionFunc func(rt ChoiceRuntime, req *model.CombatRequest) bool

// CombatInteractionSpec 战斗交互策略条目。
type CombatInteractionSpec struct {
	Priority int
	Hook     CombatInteractionFunc
}

// CombatDefendValidationFunc 战斗防御校验策略签名。
type CombatDefendValidationFunc func(rt ChoiceRuntime, player *model.Player, req *model.CombatRequest) error

// CombatDefendValidationSpec 战斗防御校验条目。
type CombatDefendValidationSpec struct {
	Priority int
	Validate CombatDefendValidationFunc
}

// CombatCounterCardFunc 战斗应战出牌策略签名。
type CombatCounterCardFunc func(rt ChoiceRuntime, player *model.Player, req *model.CombatRequest, card model.Card) (bool, model.Card, error)

// CombatCounterCardSpec 战斗应战出牌策略条目。
type CombatCounterCardSpec struct {
	Priority int
	Policy   CombatCounterCardFunc
}

// CombatCounterElementFunc 战斗应战元素策略签名。
type CombatCounterElementFunc func(rt ChoiceRuntime, player *model.Player, req *model.CombatRequest, counterCard model.Card) (bool, bool)

// CombatCounterElementSpec 战斗应战元素策略条目。
type CombatCounterElementSpec struct {
	Priority int
	Policy   CombatCounterElementFunc
}

// CombatCounterResolveFunc 战斗应战结算策略签名。
type CombatCounterResolveFunc func(rt ChoiceRuntime, player *model.Player, req *model.CombatRequest, counterCard *model.Card, useFaction bool)

// CombatCounterResolveSpec 战斗应战结算策略条目。
type CombatCounterResolveSpec struct {
	Priority int
	Resolve  CombatCounterResolveFunc
}

// MagicMissileDefendFunc 魔弹防御校验签名。
type MagicMissileDefendFunc func(rt ChoiceRuntime, player *model.Player, chain *model.MagicBulletChain) error

// MagicMissileDefendSpec 魔弹防御校验条目。
type MagicMissileDefendSpec struct {
	Priority int
	Validate MagicMissileDefendFunc
}

// MagicMissileCounterFunc 魔弹传递校验签名。
type MagicMissileCounterFunc func(rt ChoiceRuntime, player *model.Player, chain *model.MagicBulletChain, card model.Card) error

// MagicMissileCounterSpec 魔弹传递校验条目。
type MagicMissileCounterSpec struct {
	Priority int
	Validate MagicMissileCounterFunc
}

// ---- 行动选择策略 ----

// ActionSelectionModifier 抽象行动选择状态的可变字段。
type ActionSelectionModifier interface {
	SetActionRule(mode string, source string, priority int)
	SetCanMagicAction(v bool)
	SetCanMagicSkillAction(v bool)
	SetPromptChoiceType(ct string)
	SetPromptSkillID(sid string)
	SetActionRulePromptMessage(msg string)
	SetConstrainedTarget(id, name string)
	SetRuleRequiresSkipOnly(v bool)
}

// ActionSelectionOptionFunc 行动选择选项策略签名。
type ActionSelectionOptionFunc func(rt ChoiceRuntime, player *model.Player, mod ActionSelectionModifier)

// ActionSelectionOptionSpec 行动选择选项策略条目。
type ActionSelectionOptionSpec struct {
	Priority int
	Policy   ActionSelectionOptionFunc
}

// ActionSelectionValidationModifier 扩展行动选择校验阶段可用的回调设置。
type ActionSelectionValidationModifier interface {
	ActionSelectionModifier
	SetRequiredSkillID(sid string)
	SetForceSkillMustUseMessage(msg string)
	SetForceSkillOnlyMessage(msg string)
	SetForceAttackOnlyMessage(msg string)
	SetOnSkipChosen(callback func(rt ChoiceRuntime, player *model.Player) (bool, error))
	SetOnNonAttackChosen(callback func(rt ChoiceRuntime, player *model.Player, act model.PlayerAction) error)
	SetOnAttackAccepted(callback func(rt ChoiceRuntime, player *model.Player, act model.PlayerAction) error)
}

// ActionSelectionValidationFunc 行动选择校验策略签名。
type ActionSelectionValidationFunc func(rt ChoiceRuntime, player *model.Player, mod ActionSelectionValidationModifier)

// ActionSelectionValidationSpec 行动选择校验策略条目。
type ActionSelectionValidationSpec struct {
	Priority int
	Policy   ActionSelectionValidationFunc
}
