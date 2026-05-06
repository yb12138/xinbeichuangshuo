// gameflow: 行动选择阶段可选项过滤规则（适配器）。

package engine

import (
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// actionSelectionModifierAdapter 适配 engine 内部的 actionSelectionState 到 player.ActionSelectionModifier。
type actionSelectionModifierAdapter struct {
	state *actionSelectionState
}

func (a actionSelectionModifierAdapter) SetActionRule(mode string, source string, priority int) {
	a.state.setActionRule(actionSelectionRuleMode(mode), source, priority)
}
func (a actionSelectionModifierAdapter) SetCanMagicAction(v bool) { a.state.canMagicAction = v }
func (a actionSelectionModifierAdapter) SetCanMagicSkillAction(v bool) {
	a.state.canMagicSkillAction = v
}
func (a actionSelectionModifierAdapter) SetPromptChoiceType(ct string) { a.state.promptChoiceType = ct }
func (a actionSelectionModifierAdapter) SetPromptSkillID(sid string)   { a.state.promptSkillID = sid }
func (a actionSelectionModifierAdapter) SetActionRulePromptMessage(msg string) {
	a.state.actionRulePromptMessage = msg
}
func (a actionSelectionModifierAdapter) SetConstrainedTarget(id, name string) {
	a.state.constrainedTargetID = id
	a.state.constrainedTargetName = name
}
func (a actionSelectionModifierAdapter) SetRuleRequiresSkipOnly(v bool) {
	a.state.ruleRequiresSkipOnly = v
}

// actionSelectionValidationModifierAdapter 扩展适配校验阶段的回调设置。
type actionSelectionValidationModifierAdapter struct {
	actionSelectionModifierAdapter
	result *actionSelectionValidationResult
	engine *GameEngine
}

func (a actionSelectionValidationModifierAdapter) SetRequiredSkillID(sid string) {
	a.state.requiredSkillID = sid
}
func (a actionSelectionValidationModifierAdapter) SetForceSkillMustUseMessage(msg string) {
	a.state.forceSkillMustUseMessage = msg
}
func (a actionSelectionValidationModifierAdapter) SetForceSkillOnlyMessage(msg string) {
	a.state.forceSkillOnlyMessage = msg
}
func (a actionSelectionValidationModifierAdapter) SetForceAttackOnlyMessage(msg string) {
	a.state.forceAttackOnlyMessage = msg
}
func (a actionSelectionValidationModifierAdapter) SetOnSkipChosen(callback func(rt engineplayer.ChoiceRuntime, player *model.Player) (bool, error)) {
	a.state.onSkipChosen = func(e *GameEngine, player *model.Player, result *actionSelectionValidationResult) (bool, error) {
		handled, err := callback(newRoleChoiceRuntime(e), player)
		if handled && result != nil {
			result.handled = true
		}
		return handled, err
	}
}
func (a actionSelectionValidationModifierAdapter) SetOnNonAttackChosen(callback func(rt engineplayer.ChoiceRuntime, player *model.Player, act model.PlayerAction) error) {
	a.state.onNonAttackChosen = func(e *GameEngine, player *model.Player, act model.PlayerAction, result *actionSelectionValidationResult) error {
		return callback(newRoleChoiceRuntime(e), player, act)
	}
}
func (a actionSelectionValidationModifierAdapter) SetOnAttackAccepted(callback func(rt engineplayer.ChoiceRuntime, player *model.Player, act model.PlayerAction) error) {
	a.state.onAttackAccepted = func(e *GameEngine, player *model.Player, act model.PlayerAction, result *actionSelectionValidationResult) error {
		return callback(newRoleChoiceRuntime(e), player, act)
	}
}
