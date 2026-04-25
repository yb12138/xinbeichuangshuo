// gameflow: 圣枪骑士 Timing Hook 实现。

package holy_lancer

import (
	"fmt"

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

// responseSkillSkipHook 圣击：跳过[地枪]响应时触发治疗+1。
func responseSkillSkipHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	// 只在跳过地枪技能且是攻击命中阶段时触发
	if ctx.ResumePhase != "attack_hit" {
		return engineplayer.TimingHookResult{}
	}
	// 检查是否提供了地枪技能
	hasEarthSpear := false
	if ctx.OfferedSkillID == "holy_lancer_earth_spear" {
		hasEarthSpear = true
	}
	for _, skillID := range ctx.OfferedSkills {
		if skillID == "holy_lancer_earth_spear" {
			hasEarthSpear = true
			break
		}
	}
	if !hasEarthSpear {
		return engineplayer.TimingHookResult{}
	}
	player := rt.GetPlayer(ctx.TargetID)
	if player == nil || !rt.IsCharacter(player, "holy_lancer") {
		return engineplayer.TimingHookResult{}
	}
	// 如果已经阻止了圣击触发，不再触发
	if player.TurnState.UsedSkillCounts["holy_lancer_block_sacred_strike"] != 0 {
		return engineplayer.TimingHookResult{}
	}
	rt.Heal(player.ID, 1)
	rt.Log(fmt.Sprintf("%s 未发动 [地枪]，触发 [圣击]：+1治疗", player.Name))
	return engineplayer.TimingHookResult{}
}

// playerSetupHook 玩家设置时同步派生状态（天启最大治疗上限）。
func playerSetupHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	player := rt.GetPlayer(ctx.TargetID)
	if player == nil || !rt.IsCharacter(player, "holy_lancer") {
		return engineplayer.TimingHookResult{}
	}
	syncRevelationMaxHeal(rt, player)
	return engineplayer.TimingHookResult{}
}

// campCupChangedHook 阵营杯子变化时同步所有圣枪派生状态。
func campCupChangedHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	for _, player := range rt.GetPlayers() {
		if player == nil || !rt.IsCharacter(player, "holy_lancer") {
			continue
		}
		syncRevelationMaxHeal(rt, player)
	}
	return engineplayer.TimingHookResult{}
}

// syncRevelationMaxHeal 同步天启技能的最大治疗上限派生状态。
func syncRevelationMaxHeal(rt engineplayer.HookRuntime, player *model.Player) {
	if player == nil {
		return
	}
	enemyCamp := model.BlueCamp
	if player.Camp == model.BlueCamp {
		enemyCamp = model.RedCamp
	}
	maxHeal := 2
	if rt.GetCampCups(string(player.Camp)) >= rt.GetCampCups(string(enemyCamp)) {
		maxHeal = 3
	}
	player.MaxHeal = maxHeal
}
