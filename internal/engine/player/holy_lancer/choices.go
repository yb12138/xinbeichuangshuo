// gameflow: 圣枪骑士角色选择流。

package holy_lancer

import (
	"fmt"

	"starcup-engine/internal/engine/core/runtimeutil"
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

type choiceHandler struct{}

func NewChoiceHandler() engineplayer.ChoiceHandler {
	return choiceHandler{}
}

func (choiceHandler) BuildPrompt(rt engineplayer.ChoiceRuntime, choiceType, playerID string, _ *model.Player, data map[string]interface{}) *model.Prompt {
	if choiceType != "holy_lancer_earth_spear_x" {
		return nil
	}
	maxX := runtimeutil.ToIntContextValue(data["max_x"])
	options := make([]model.PromptOption, 0, maxX)
	for x := 1; x <= maxX; x++ {
		options = append(options, model.PromptOption{
			ID:    fmt.Sprintf("%d", x),
			Label: fmt.Sprintf("移除%d点治疗，本次伤害+%d", x, x),
		})
	}
	return &model.Prompt{
		Type:         model.PromptConfirm,
		PlayerID:     playerID,
		Message:      "【地枪】请选择X值：",
		Options:      options,
		Min:          1,
		Max:          1,
		Presentation: &model.PromptPresentation{Kind: model.PresentationNumeric, NumericBase: 0},
	}
}

func (choiceHandler) HandleChoice(rt engineplayer.ChoiceRuntime, _ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "holy_lancer_earth_spear_x":
		return true, handleHolyLancerEarthSpearXChoice(rt, selectionIndex, ctxData)
	default:
		return false, nil
	}
}

func handleHolyLancerEarthSpearXChoice(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
	x := selectionIndex + 1
	if x < 1 || x > maxX || x > user.Heal {
		return fmt.Errorf("无效的X值")
	}
	userCtx, ok := ctxData["user_ctx"].(*model.Context)
	if !ok || userCtx == nil || userCtx.EventCtx == nil || userCtx.EventCtx.DamageVal == nil {
		return fmt.Errorf("地枪上下文丢失")
	}
	user.Heal -= x
	*userCtx.EventCtx.DamageVal += x
	user.TurnState.UsedSkillCounts["holy_lancer_block_sacred_strike"] = 1
	rt.Log(fmt.Sprintf("%s 发动 [地枪]，移除%d治疗，本次伤害+%d", user.Name, x, x))
	rt.PopInterrupt()
	rt.ResumePendingAttackHit(ctxData)
	return nil
}
