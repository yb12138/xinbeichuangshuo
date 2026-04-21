package adventurer

import (
	"fmt"

	"starcup-engine/internal/model"
)

// IsForcedParadiseResponse returns true when the pending interrupt is a
// response-skill prompt for the given player that requires selecting
// "adventurer_paradise" and the player has the extract_requires_paradise flag.
func IsForcedParadiseResponse(pendingInterrupt *model.Interrupt, players map[string]*model.Player, playerID string) bool {
	intr := pendingInterrupt
	if intr == nil || intr.Type != model.InterruptResponseSkill || intr.PlayerID != playerID {
		return false
	}
	p := players[playerID]
	if p == nil || p.TurnState.SkillFlowState == nil || p.TurnState.SkillFlowState["adventurer_extract_requires_paradise"] <= 0 {
		return false
	}
	for _, sid := range intr.SkillIDs {
		if sid == "adventurer_paradise" {
			return true
		}
	}
	return false
}

// ResolveUndergroundLaw applies the Underground Law effect: instead of buying,
// the player's camp gains 2 gems.
func ResolveUndergroundLaw(rt model.IGameEngine, user *model.Player) {
	if user == nil {
		return
	}
	rt.ModifyGem(string(user.Camp), 2)
	rt.Log(fmt.Sprintf("%s 的 [地下法则] 生效，本次购买改为战绩区+2宝石", user.Name))
	rt.Log(fmt.Sprintf("[Skill] %s 使用了技能: 地下法则", user.Name))
}
