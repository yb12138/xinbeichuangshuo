// gameflow: 仲裁者 Timing Hook 实现（回合开始/行动前）。

package arbiter

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// turnStartResetHook 回合开始重置仲裁者计数器。
func turnStartResetHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || !engineplayer.IsCharacter(p, "arbiter") {
		return engineplayer.TimingHookResult{}
	}
	p.TurnState.UsedSkillCounts["arbiter_skip_forced_doomsday"] = 0
	p.TurnState.UsedSkillCounts["arbiter_forced_doomsday_done_turn"] = 0
	return engineplayer.TimingHookResult{}
}

// turnStartJudgmentUpkeepHook 回合开始审判形态审判+1。
func turnStartJudgmentUpkeepHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || !engineplayer.IsCharacter(p, "arbiter") {
		return engineplayer.TimingHookResult{}
	}
	if !rt.HasForm(p, model.FormArbiterJudgment) || rt.HasUsedActionSkill(p) {
		return engineplayer.TimingHookResult{}
	}
	engineplayer.EnsurePlayerTokensMap(p)
	if p.Tokens["judgment"] >= 4 {
		return engineplayer.TimingHookResult{}
	}
	p.Tokens["judgment"]++
	rt.Log(fmt.Sprintf("%s 处于审判形态，回合开始审判+1（当前%d）", p.Name, p.Tokens["judgment"]))
	return engineplayer.TimingHookResult{}
}

// turnStartForcedDoomsdayHook 回合开始时末日审判强制发动检查。
func turnStartForcedDoomsdayHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || !engineplayer.IsCharacter(p, "arbiter") {
		return engineplayer.TimingHookResult{}
	}
	engineplayer.EnsurePlayerTokensMap(p)
	if p.Tokens["judgment"] < 4 || p.TurnState.UsedSkillCounts["arbiter_skip_forced_doomsday"] != 0 || p.TurnState.UsedSkillCounts["arbiter_forced_doomsday_done_turn"] != 0 {
		p.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] = 0
		return engineplayer.TimingHookResult{}
	}
	if len(rt.CampEnemyIDs(p.Camp)) == 0 {
		p.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] = 0
		return engineplayer.TimingHookResult{}
	}
	if p.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] == 0 {
		rt.Log(fmt.Sprintf("%s 的审判已达上限：本行动阶段必须发动 [末日审判]", p.Name))
	}
	p.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] = 1
	return engineplayer.TimingHookResult{}
}
