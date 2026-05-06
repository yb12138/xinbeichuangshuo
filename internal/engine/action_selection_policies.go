// gameflow: 行动选择界面：可选攻击/法术/技能等策略。

package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

// actionSelectionState 行动选择统一状态：同时承载「选项生成」与「输入校验」两条链路。
// 先由 policy 注入规则约束，再由 option-builder 生成 Prompt，最终由 validator 对玩家输入做一致性校验。
type actionSelectionState struct {
	currentPID string // 当前回合行动玩家 ID，供 Prompt 绑定

	// currentExtraAction 来自 TurnState：本段是否为「额外行动」及其类型（"" / Attack / Magic）。
	currentExtraAction string
	// isRestrictedExtraAction：当前处于额外行动且类型为攻击或法术时，行动枢纽只能选对应大类（不能买合提等混选）。
	isRestrictedExtraAction bool
	// canMagicAction：当前形态与规则下是否允许从手牌打出法术（如部分形态禁法术）。
	canMagicAction bool
	// canMagicSkillAction：是否存在可在行动阶段占用「法术」入口的行动技（如仲裁强制末日审判时的技能发动）。
	canMagicSkillAction bool
	// hasRestrictedExtraAction：在额外行动约束下，是否仍有满足属性/类型的可打出牌；若无则只能「跳过额外行动」。
	hasRestrictedExtraAction bool

	// actionRuleMode 表示当前行动枢纽被哪类规则“劫持”：
	// - force_skill_magic: 仅可通过法术入口发动指定行动技
	// - force_attack: 仅可主动攻击
	// - force_attack_or_skip: 仅可主动攻击，或显式跳过并清除约束
	actionRuleMode actionSelectionRuleMode
	// actionRulePriority 用于处理多规则并发时的优先级（值越大优先级越高）。
	actionRulePriority int
	// actionRuleSource 记录约束来源（用于日志/排障）。
	actionRuleSource string
	// actionRulePromptMessage 允许约束来源覆盖默认提示语（为空则走默认模板）。
	actionRulePromptMessage string
	// constrainedTargetID/Name：当约束要求攻击特定目标时用于文案展示（如挑衅）。
	constrainedTargetID   string
	constrainedTargetName string
	// ruleRequiresSkipOnly：约束要求攻击，但当前无可打出攻击牌时，行动枢纽只允许“跳过行动”。
	ruleRequiresSkipOnly bool

	// promptChoiceType / promptSkillID：前端特殊 UI 模式（如强制从枢纽进某技能选择）；可与 actionRuleMode 组合生效。
	promptChoiceType string
	promptSkillID    string

	validOptions   []model.PromptOption // 主选项：攻击 / 法术 / 特殊 / 无法行动等
	specialOptions []model.PromptOption // 选「特殊」后展开的购买 / 合成 / 提炼
	promptMessage  string               // 行动枢纽主提示语（含额外行动、挑衅、百龙、强制末日等说明）

	// ===== 以下字段供 validateActionSelectionPolicies 使用，与上方规则字段共享同一约束语义 =====
	requiredSkillID          string
	forceSkillMustUseMessage string
	forceSkillOnlyMessage    string
	forceAttackOnlyMessage   string

	onSkipChosen      func(e *GameEngine, player *model.Player, result *actionSelectionValidationResult) (handled bool, err error)
	onNonAttackChosen func(e *GameEngine, player *model.Player, act model.PlayerAction, result *actionSelectionValidationResult) error
	onAttackAccepted  func(e *GameEngine, player *model.Player, act model.PlayerAction, result *actionSelectionValidationResult) error
}

type actionSelectionRuleMode string

const (
	actionSelectionRuleNone              actionSelectionRuleMode = ""
	actionSelectionRuleForceSkillAsMagic actionSelectionRuleMode = "force_skill_magic"
	actionSelectionRuleForceAttack       actionSelectionRuleMode = "force_attack"
	actionSelectionRuleForceAttackOrSkip actionSelectionRuleMode = "force_attack_or_skip"
)

func (s *actionSelectionState) setActionRule(mode actionSelectionRuleMode, source string, priority int) {
	if s == nil {
		return
	}
	if priority < s.actionRulePriority {
		return
	}
	s.actionRuleMode = mode
	s.actionRuleSource = source
	s.actionRulePriority = priority
	s.actionRulePromptMessage = ""
	s.constrainedTargetID = ""
	s.constrainedTargetName = ""
	s.ruleRequiresSkipOnly = false
	s.requiredSkillID = ""
	s.forceSkillMustUseMessage = ""
	s.forceSkillOnlyMessage = ""
	s.forceAttackOnlyMessage = ""
	s.onSkipChosen = nil
	s.onNonAttackChosen = nil
	s.onAttackAccepted = nil
}

type actionSelectionValidationResult struct {
	handled             bool
	afterAttackAccepted func(e *GameEngine, player *model.Player, act model.PlayerAction) error
}

func (e *GameEngine) buildActionSelectionState(currentPID string, player *model.Player) actionSelectionState {
	state := actionSelectionState{currentPID: currentPID}
	if player == nil {
		return state
	}
	if player.Tokens == nil {
		player.Tokens = map[string]int{}
	}

	state.currentExtraAction = player.TurnState.CurrentExtraAction
	state.isRestrictedExtraAction = state.currentExtraAction == "Attack" || state.currentExtraAction == "Magic"
	state.canMagicAction = e.canCastMagicInAction(player)
	state.canMagicSkillAction = e.hasUsableActionSkillForExtraMagic(player)
	state.hasRestrictedExtraAction = true
	if state.isRestrictedExtraAction {
		state.hasRestrictedExtraAction = e.checkExtraActionCards(player, state.currentExtraAction, player.TurnState.CurrentExtraElement)
	}
	e.applyActionSelectionPolicies(player, &state)
	return state
}

func (e *GameEngine) buildActionSelectionOptions(currentPID string, player *model.Player) actionSelectionState {
	state := e.buildActionSelectionState(currentPID, player)
	e.appendBaseActionSelectionOptions(player, &state)
	e.finalizeActionSelectionPromptState(player, &state)
	return state
}

func (e *GameEngine) applyActionSelectionPolicies(player *model.Player, state *actionSelectionState) {
	e.applyTimingBeforeActionExecuteOptionPolicies(player, state)
	e.applyTimingBeforeActionExecuteValidationPolicies(player, state)
}

func (e *GameEngine) appendBaseActionSelectionOptions(player *model.Player, state *actionSelectionState) {
	if player == nil || state == nil {
		return
	}

	switch state.currentExtraAction {
	case "Attack":
		if state.hasRestrictedExtraAction {
			state.validOptions = append(state.validOptions, model.PromptOption{ID: "attack", Label: "攻击"})
		}
	case "Magic":
		if state.hasRestrictedExtraAction {
			state.validOptions = append(state.validOptions, model.PromptOption{ID: "magic", Label: "法术"})
		}
	default:
		if state.actionRuleMode == actionSelectionRuleForceSkillAsMagic {
			state.validOptions = append(state.validOptions, model.PromptOption{ID: "magic", Label: "法术（仅限末日审判）"})
		} else if state.actionRuleMode == actionSelectionRuleForceAttack {
			state.validOptions = append(state.validOptions, model.PromptOption{ID: "attack", Label: "攻击（百式幻龙拳）"})
		} else if state.actionRuleMode == actionSelectionRuleForceAttackOrSkip {
			if !state.ruleRequiresSkipOnly {
				state.validOptions = append(state.validOptions, model.PromptOption{ID: "attack", Label: "攻击（受挑衅约束）"})
			}
		} else {
			state.validOptions = append(state.validOptions, model.PromptOption{ID: "attack", Label: "攻击"})
			if state.canMagicAction || state.canMagicSkillAction {
				state.validOptions = append(state.validOptions, model.PromptOption{ID: "magic", Label: "法术"})
			}
		}
	}

	if state.actionRuleMode == actionSelectionRuleNone && !state.isRestrictedExtraAction && !player.TurnState.HasStartupSkillOrSpecialActionsLocked() {
		maxHand := e.GetMaxHand(player)
		canBuyOrSynth := len(player.Hand)+3 <= maxHand

		if canBuyOrSynth {
			state.specialOptions = append(state.specialOptions, model.PromptOption{ID: "buy", Label: "购买"})
		}

		var totalStones int
		if player.Camp == model.RedCamp {
			totalStones = e.State.RedGems + e.State.RedCrystals
		} else {
			totalStones = e.State.BlueGems + e.State.BlueCrystals
		}
		if canBuyOrSynth && totalStones >= 3 {
			state.specialOptions = append(state.specialOptions, model.PromptOption{ID: "synthesize", Label: "合成"})
		}

		currentEnergy := player.Gem + player.Crystal
		if totalStones > 0 && currentEnergy < e.getPlayerEnergyCap(player) {
			state.specialOptions = append(state.specialOptions, model.PromptOption{ID: "extract", Label: "提炼"})
		}

		if len(state.specialOptions) > 0 {
			state.validOptions = append(state.validOptions, model.PromptOption{ID: "special", Label: "特殊"})
		}
	}

	if !state.isRestrictedExtraAction {
		if state.actionRuleMode == actionSelectionRuleForceAttackOrSkip {
			state.validOptions = append(state.validOptions, model.PromptOption{ID: "cannot_act", Label: "跳过行动（移除挑衅）"})
			return
		}
		canCannotAct, _ := e.checkPlayerCannotAct(player)
		if canCannotAct {
			state.validOptions = append(state.validOptions, model.PromptOption{ID: "cannot_act", Label: "无法行动（展示手牌）"})
		}
	} else if !state.hasRestrictedExtraAction {
		state.validOptions = append(state.validOptions, model.PromptOption{ID: "cannot_act", Label: "跳过额外行动"})
	}
}

func (e *GameEngine) finalizeActionSelectionPromptState(player *model.Player, state *actionSelectionState) {
	if player == nil || state == nil {
		return
	}
	state.promptMessage = "请选择行动类型"
	if state.currentExtraAction == "Attack" {
		state.promptMessage = "当前为额外攻击行动，仅可执行攻击。请选择行动类型"
	} else if state.currentExtraAction == "Magic" {
		state.promptMessage = "当前为额外法术行动，仅可执行法术。请选择行动类型"
	} else if state.actionRuleMode == actionSelectionRuleForceSkillAsMagic {
		if state.actionRulePromptMessage != "" {
			state.promptMessage = state.actionRulePromptMessage
		} else {
			state.promptMessage = "当前规则要求：本行动阶段必须发动指定行动技。"
		}
	} else if state.actionRuleMode == actionSelectionRuleForceAttack {
		if state.actionRulePromptMessage != "" {
			state.promptMessage = state.actionRulePromptMessage
		} else {
			state.promptMessage = "当前规则要求：本行动阶段只能主动攻击。"
		}
	} else if state.actionRuleMode == actionSelectionRuleForceAttackOrSkip {
		targetName := state.constrainedTargetName
		if targetName == "" {
			targetName = "指定目标"
		}
		if state.ruleRequiresSkipOnly {
			state.promptMessage = fmt.Sprintf("你受到约束效果影响：本次行动阶段必须主动攻击 %s，但你没有攻击牌，只能选择跳过行动并移除此牌。", targetName)
		} else {
			state.promptMessage = fmt.Sprintf("你受到约束效果影响：本次行动阶段必须且只能主动攻击 %s，或选择跳过行动并移除此牌。", targetName)
		}
	}
	if state.isRestrictedExtraAction && !state.hasRestrictedExtraAction {
		state.promptMessage = "当前为额外行动阶段，但你没有满足约束的可执行动作。可选择跳过本次额外行动。"
	}
}

func (e *GameEngine) validateActionSelectionPolicies(player *model.Player, act model.PlayerAction) (actionSelectionValidationResult, error) {
	state := e.buildActionSelectionState("", player)
	var result actionSelectionValidationResult
	if err := e.validateActionSelectionRule(player, act, &state, &result); err != nil {
		return result, err
	}
	return result, nil
}

func (e *GameEngine) validateActionSelectionRule(player *model.Player, act model.PlayerAction, state *actionSelectionState, result *actionSelectionValidationResult) error {
	if e == nil || player == nil || state == nil {
		return nil
	}

	switch state.actionRuleMode {
	case actionSelectionRuleForceSkillAsMagic:
		if act.Type != model.CmdSkill {
			if state.forceSkillMustUseMessage != "" {
				return fmt.Errorf(state.forceSkillMustUseMessage)
			}
			return fmt.Errorf("当前规则限制：本行动阶段必须发动指定行动技")
		}
		if state.requiredSkillID != "" && act.SkillID != state.requiredSkillID {
			if state.forceSkillOnlyMessage != "" {
				return fmt.Errorf(state.forceSkillOnlyMessage)
			}
			return fmt.Errorf("当前规则限制：本行动阶段只能发动指定行动技")
		}
		return nil

	case actionSelectionRuleForceAttack, actionSelectionRuleForceAttackOrSkip:
		if state.actionRuleMode == actionSelectionRuleForceAttackOrSkip && act.Type == model.CmdCannotAct {
			if state.onSkipChosen != nil {
				handled, err := state.onSkipChosen(e, player, result)
				if err != nil {
					return err
				}
				if handled {
					return nil
				}
			}
			return nil
		}
		// 强制攻击模式仍允许走「无法行动」通用流程（由后续规则判断是否有牌可打）。
		if state.actionRuleMode == actionSelectionRuleForceAttack && act.Type == model.CmdCannotAct {
			return nil
		}
		if act.Type != model.CmdAttack {
			if state.onNonAttackChosen != nil {
				if err := state.onNonAttackChosen(e, player, act, result); err != nil {
					return err
				}
			}
			if state.ruleRequiresSkipOnly {
				targetName := state.constrainedTargetName
				if targetName == "" {
					targetName = "指定目标"
				}
				return fmt.Errorf("你受到约束效果影响：本次行动阶段必须主动攻击 %s，但你没有攻击牌，只能选择跳过行动", targetName)
			}
			if state.forceAttackOnlyMessage != "" {
				return fmt.Errorf(state.forceAttackOnlyMessage)
			}
			return fmt.Errorf("当前规则限制：本行动阶段只能主动攻击")
		}

		targetID := act.TargetID
		if targetID == "" && len(act.TargetIDs) > 0 {
			targetID = act.TargetIDs[0]
		}
		if targetID == "" {
			return fmt.Errorf("攻击必须指定目标")
		}
		targetPlayer := e.State.Players[targetID]
		if targetPlayer == nil {
			return fmt.Errorf("目标不存在")
		}
		if targetPlayer.Camp == player.Camp {
			return fmt.Errorf("攻击目标必须是敌方角色")
		}
		if state.constrainedTargetID != "" && targetID != state.constrainedTargetID {
			targetName := state.constrainedTargetName
			if targetName == "" {
				targetName = "指定目标"
			}
			return fmt.Errorf("你受到约束效果影响：本次行动阶段只能主动攻击 %s", targetName)
		}
		if state.onAttackAccepted != nil && result != nil {
			result.afterAttackAccepted = func(e *GameEngine, player *model.Player, act model.PlayerAction) error {
				return state.onAttackAccepted(e, player, act, result)
			}
		}
		return nil
	default:
		return nil
	}
}
