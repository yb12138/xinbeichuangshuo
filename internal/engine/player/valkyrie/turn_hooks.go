// gameflow: 女武神回合钩子。

package valkyrie

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
)

// MilitaryGlory 回合开始时军威神光触发检查。
func MilitaryGlory(rt engineplayer.ChoiceRuntime, player *model.Player) bool {
	if player == nil || !engineplayer.IsCharacter(player, "valkyrie") {
		return false
	}
	if !InHeroicForm(player) || player.TurnState.UsedSkillCounts["valkyrie_military_glory"] > 0 {
		return false
	}
	ctx := rt.BuildContext(player, nil, model.TimingTurnStart, &model.EventContext{
		Type:     model.EventTurnStart,
		SourceID: player.ID,
	})
	handler := skills.GetHandler("valkyrie_military_glory")
	if handler == nil || !handler.CanUse(ctx) {
		return false
	}
	player.TurnState.UsedSkillCounts["valkyrie_military_glory"] = 1
	if err := handler.Execute(ctx); err != nil {
		rt.Log(fmt.Sprintf("[Skill Error] 军威神光执行失败: %v", err))
		return false
	}
	rt.RecordSkillUsage(player.ID, "军威神光", model.SkillTypeStartup)
	return rt.GetPendingInterrupt() != nil
}
