// gameflow: 女武神 Timing Hook 实现。

package valkyrie

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
)

// turnStartMilitaryGloryHook 回合开始时军威神光触发检查。
func turnStartMilitaryGloryHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	player := rt.GetPlayer(ctx.SourceID)
	if player == nil || !engineplayer.IsCharacter(player, "valkyrie") {
		return engineplayer.TimingHookResult{}
	}
	if !InHeroicForm(player) || player.TurnState.UsedSkillCounts["valkyrie_military_glory"] > 0 {
		return engineplayer.TimingHookResult{}
	}
	crt := rt.AsChoiceRuntime()
	if crt == nil {
		return engineplayer.TimingHookResult{}
	}
	ctx2 := crt.BuildContext(player, nil, model.TimingOnTurnStart, &model.EventContext{
		Type:     model.EventTurnStart,
		SourceID: player.ID,
	})
	handler := skills.GetHandler("valkyrie_military_glory")
	if handler == nil || !handler.CanUse(ctx2) {
		return engineplayer.TimingHookResult{}
	}
	player.TurnState.UsedSkillCounts["valkyrie_military_glory"] = 1
	if err := handler.Execute(ctx2); err != nil {
		rt.Log(fmt.Sprintf("[Skill Error] 军威神光执行失败: %v", err))
		return engineplayer.TimingHookResult{}
	}
	crt.RecordSkillUsage(player.ID, "军威神光", model.SkillTypeStartup)
	if rt.GetPendingInterrupt() != nil {
		return engineplayer.TimingHookResult{Interrupted: true}
	}
	return engineplayer.TimingHookResult{}
}
