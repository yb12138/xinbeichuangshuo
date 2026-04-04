package engine

import "starcup-engine/internal/model"

const (
	turnFlagSpecialActionLocked   = "turn_special_action_locked"
	turnFlagActionEndHookResuming = "turn_action_end_hook_resuming"
)

func (e *GameEngine) hasPerformedStartupThisTurn(player *model.Player) bool {
	if player == nil {
		return false
	}
	if player.TurnState.HasUsedTriggerSkill {
		return true
	}
	return player.TurnState.UsedSkillCounts[turnFlagSpecialActionLocked] > 0
}

func (e *GameEngine) markSpecialActionLockedForTurn(player *model.Player) {
	if player == nil {
		return
	}
	if player.TurnState.UsedSkillCounts == nil {
		player.TurnState.UsedSkillCounts = map[string]int{}
	}
	player.TurnState.UsedSkillCounts[turnFlagSpecialActionLocked] = 1
}

func (e *GameEngine) markActionEndHookResuming(player *model.Player) {
	if player == nil {
		return
	}
	if player.TurnState.UsedSkillCounts == nil {
		player.TurnState.UsedSkillCounts = map[string]int{}
	}
	player.TurnState.UsedSkillCounts[turnFlagActionEndHookResuming] = 1
}

func (e *GameEngine) consumeActionEndHookResuming(player *model.Player) bool {
	if player == nil || player.TurnState.UsedSkillCounts == nil || player.TurnState.UsedSkillCounts[turnFlagActionEndHookResuming] <= 0 {
		return false
	}
	player.TurnState.UsedSkillCounts[turnFlagActionEndHookResuming] = 0
	return true
}

