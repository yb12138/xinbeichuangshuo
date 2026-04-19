// gameflow: 冒险家：地下法则、欺诈攻击等辅助函数。

package engine

import (
	"fmt"
	"starcup-engine/internal/model"
)

func (e *GameEngine) isForcedAdventurerParadiseResponse(playerID string) bool {
	intr := e.State.PendingInterrupt
	if intr == nil || intr.Type != model.InterruptResponseSkill || intr.PlayerID != playerID {
		return false
	}
	player := e.State.Players[playerID]
	if player == nil || player.TurnState.SkillFlowState == nil || player.TurnState.SkillFlowState["adventurer_extract_requires_paradise"] <= 0 {
		return false
	}
	for _, sid := range intr.SkillIDs {
		if sid == "adventurer_paradise" {
			return true
		}
	}
	return false
}

func (e *GameEngine) resolveAdventurerUndergroundLaw(user *model.Player) {
	if user == nil {
		return
	}
	e.ModifyGem(string(user.Camp), 2)
	e.Log(fmt.Sprintf("%s 的 [地下法则] 生效，本次购买改为战绩区+2宝石", user.Name))
	e.Log(fmt.Sprintf("[Skill] %s 使用了技能: 地下法则", user.Name))
}
