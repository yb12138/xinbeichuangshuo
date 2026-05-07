// gameflow: 技能恢复状态的调度与执行。

package engine

import (
	"fmt"

	"starcup-engine/internal/model"
	skillrt "starcup-engine/internal/engine/runtime/skill"
)

func (e *GameEngine) ProcessPendingSkillResume() bool {
	if e == nil || e.skillResume == nil {
		return false
	}
	if len(e.State.PendingDamageQueue) > 0 {
		return false
	}
	state := e.skillResume
	e.skillResume = nil
	return e.resumeSkillExecution(state)
}

func (e *GameEngine) resumeSkillExecution(state *skillResumeState) bool {
	if e == nil || e.State == nil || state == nil {
		return false
	}
	user := e.State.Players[state.playerID]
	if user == nil || user.Character == nil {
		e.Log(fmt.Sprintf("[Warn] 技能恢复失败：执行者不存在 %s", state.playerID))
		return true
	}
	use, err := e.buildSkillResumeUse(state, user)
	if err != nil {
		e.Log(fmt.Sprintf("[Warn] 技能恢复失败：%v", err))
		return true
	}

	if err := e.executeSkillFlow(use); err != nil {
		e.Log(fmt.Sprintf("[Error] 技能恢复执行失败: %v", err))
		return true
	}
	if err := e.finishSkillUse(use); err != nil {
		e.Log(fmt.Sprintf("[Error] 技能恢复收尾失败: %v", err))
	}
	return true
}

func (e *GameEngine) buildSkillResumeUse(state *skillResumeState, user *model.Player) (*skillUseRequest, error) {
	skillDef := skillrt.FindCharacterSkill(user.Character, state.skillID)
	if skillDef == nil {
		e.Log(fmt.Sprintf("[Warn] 技能恢复失败：技能不存在 %s", state.skillID))
		return nil, fmt.Errorf("技能不存在 %s", state.skillID)
	}

	actualTargets := make([]*model.Player, 0, len(state.targetIDs))
	if len(state.targetIDs) > 0 {
		for _, tid := range state.targetIDs {
			if p := e.State.Players[tid]; p != nil {
				actualTargets = append(actualTargets, p)
			}
		}
	}
	var target *model.Player
	if len(actualTargets) > 0 {
		target = actualTargets[0]
	}

	return &skillUseRequest{
		engine:         e,
		player:         user,
		skillDef:       skillDef,
		skillID:        state.skillID,
		policy:         resolveSkillUsePolicy(state.skillID),
		targetIDs:      append([]string{}, state.targetIDs...),
		discardedCards: append([]model.Card{}, state.discardedCards...),
		target:         target,
		actualTargets:  actualTargets,
	}, nil
}
