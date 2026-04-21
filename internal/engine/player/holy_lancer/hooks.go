// gameflow: 圣枪骑士 Timing Hook 实现。

package holy_lancer

import (
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// attackStateResetHook resets holy lancer attack-related flags when a new attack is declared.
func attackStateResetHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	player := rt.GetPlayer(ctx.SourceID)
	if player == nil {
		return engineplayer.TimingHookResult{}
	}
	player.TurnState.UsedSkillCounts["holy_lancer_block_sacred_strike"] = 0
	player.TurnState.UsedSkillCounts["holy_lancer_sky_spear_no_counter"] = 0
	return engineplayer.TimingHookResult{}
}

// attackGatingHook applies holy lancer sky spear no-counter gating on attack.
func attackGatingHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || p.TurnState.UsedSkillCounts["holy_lancer_sky_spear_no_counter"] <= 0 {
		return engineplayer.TimingHookResult{}
	}
	if ctx.AttackInfo == nil {
		return engineplayer.TimingHookResult{}
	}
	ctx.AttackInfo.SetInterceptTag(model.CombatInterceptUnrespondable)
	p.TurnState.UsedSkillCounts["holy_lancer_sky_spear_no_counter"] = 0
	return engineplayer.TimingHookResult{}
}

// turnEndHook 回合结束：祈祷计数器重置。
func turnEndHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil {
		return engineplayer.TimingHookResult{}
	}
	p.TurnState.UsedSkillCounts["holy_lancer_prayer"] = 0
	return engineplayer.TimingHookResult{}
}
