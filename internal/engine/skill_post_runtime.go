package engine

// skillPostHook 在技能主动发动成功并完成基础收尾后执行。
// 用于处理“技能效果之外”的行动阶段约束收口（如强制行动标记清理）。
type skillPostHook func(e *GameEngine, use *skillUseRequest)

func (e *GameEngine) runTimingOnActionEndSkillPost(use *skillUseRequest) {
	if e == nil || use == nil {
		return
	}
	for _, hook := range e.skillPostHooks {
		hook(e, use)
	}
}

func skillPostArbiterForcedDoomsdayCleanupHook(e *GameEngine, use *skillUseRequest) {
	if e == nil || use == nil || use.player == nil {
		return
	}
	if use.skillID != "arbiter_doomsday" {
		return
	}
	player := use.player
	if player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] <= 0 {
		return
	}
	player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] = 0
	player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_done_turn"] = 1
	consumeHeroTauntRestriction(e, player)
}
