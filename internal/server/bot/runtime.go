package bot

import "starcup-engine/internal/model"

func CanActPrompt(state *model.GameState, playerID string, prompt *model.Prompt) bool {
	if state == nil || prompt == nil {
		return false
	}
	if prompt.PlayerID != "" && prompt.PlayerID != playerID {
		return false
	}

	// 中断提示优先：由 PendingInterrupt.PlayerID 统一判定，避免与选项ID模式冲突（如魔弹也有 take/defend/counter）。
	if state.PendingInterrupt != nil {
		return state.PendingInterrupt.PlayerID == playerID
	}

	// 战斗响应提示。
	if hasPromptOption(prompt, "take") || hasPromptOption(prompt, "defend") || hasPromptOption(prompt, "counter") {
		if len(state.CombatStack) == 0 || state.Subflow != model.SubflowNone ||
			(state.CombatStage != model.CombatStageDeclare && state.CombatStage != model.CombatStageHitCheck) {
			return false
		}
		combatReq := state.CombatStack[len(state.CombatStack)-1]
		return combatReq.TargetID == playerID
	}

	// 行动选择提示。
	if hasPromptOption(prompt, "attack") || hasPromptOption(prompt, "magic") || hasPromptOption(prompt, "special") ||
		hasPromptOption(prompt, "buy") || hasPromptOption(prompt, "extract") ||
		hasPromptOption(prompt, "synthesize") || hasPromptOption(prompt, "cannot_act") {
		if state.Subflow != model.SubflowNone || state.CombatStage != model.CombatStageNone ||
			state.TurnStage != model.TurnStageActionExecution || len(state.ActionQueue) > 0 || len(state.PlayerOrder) == 0 {
			return false
		}
		if state.CurrentTurn < 0 || state.CurrentTurn >= len(state.PlayerOrder) {
			return false
		}
		return state.PlayerOrder[state.CurrentTurn] == playerID
	}

	return false
}

func hasPromptOption(prompt *model.Prompt, optionID string) bool {
	if prompt == nil {
		return false
	}
	for _, o := range prompt.Options {
		if o.ID == optionID {
			return true
		}
	}
	for _, o := range prompt.SpecialOptions {
		if o.ID == optionID {
			return true
		}
	}
	return false
}
