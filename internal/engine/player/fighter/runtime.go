package fighter

import (
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

func LockedTarget(rt engineplayer.ChoiceRuntime, p *model.Player) *model.Player {
	if p == nil {
		return nil
	}
	order := engineplayer.GetSkillFlowState(p, "fighter_hundred_dragon_target_order")
	if order <= 0 {
		return nil
	}
	orderIDs := rt.GetPlayerOrder()
	if order > len(orderIDs) {
		return nil
	}
	return rt.GetPlayers()[orderIDs[order-1]]
}

func ClearHundredDragon(rt engineplayer.ChoiceRuntime, p *model.Player, logLine string) bool {
	if p == nil || !engineplayer.IsCharacter(p, "fighter") {
		return false
	}
	defer rt.PoseChangeGuard()
	active := engineplayer.HasForm(p, model.FormFighterHundredDragon) || engineplayer.GetSkillFlowState(p, "fighter_hundred_dragon_target_order") > 0
	engineplayer.ClearForm(p, model.FormFighterHundredDragon)
	engineplayer.SetSkillFlowState(p, "fighter_hundred_dragon_target_order", 0)
	if active && logLine != "" {
		rt.Log(logLine)
	}
	return active
}
