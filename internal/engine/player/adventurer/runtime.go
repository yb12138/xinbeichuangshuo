package adventurer

import (
	"fmt"

	"starcup-engine/internal/model"
)

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
