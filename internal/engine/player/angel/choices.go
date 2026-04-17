// gameflow: 天使角色选择流。

package angel

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
	switch choiceType {
	case "angel_bond_heal_target":
		options := buildPromptOptionsForPlayerIDs(rt.AllPlayers(), runtimeutil.ParseStringSliceContextValue(data["target_ids"]))
		if len(options) == 0 {
			return nil
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【天使羁绊】请选择1名角色获得+1治疗：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	case "god_protection_x":
		maxX := runtimeutil.ToIntContextValue(data["max_x"])
		if maxX < 1 {
			return nil
		}
		options := make([]model.PromptOption, 0, maxX)
		for x := 1; x <= maxX; x++ {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", x-1),
				Label: fmt.Sprintf("消耗%d点水晶，抵御%d点士气下降", x, x),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【神之庇护】请选择X值：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	default:
		return nil
	}
}

func (choiceHandler) HandleChoice(rt engineplayer.ChoiceRuntime, playerID string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "angel_bond_heal_target":
		return true, handleAngelBondHealChoice(rt, playerID, selectionIndex, ctxData)
	case "god_protection_x":
		return true, handleGodProtectionChoice(rt, selectionIndex, ctxData)
	default:
		return false, nil
	}
}

func handleGodProtectionChoice(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.LookupPlayer(userID)
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	userCtx, ok := ctxData["user_ctx"].(*model.Context)
	if !ok || userCtx == nil || userCtx.EventCtx == nil || userCtx.EventCtx.DamageVal == nil {
		return fmt.Errorf("神之庇护上下文丢失")
	}
	maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
	x := selectionIndex + 1
	if x < 1 || x > maxX {
		return fmt.Errorf("无效的X值")
	}
	if current := *userCtx.EventCtx.DamageVal; current > 0 && x > current {
		return fmt.Errorf("X值超过当前可抵御的士气下降")
	}
	if !rt.ConsumeCrystalCost(user.ID, x) {
		return fmt.Errorf("神之庇护需要%d点水晶（红宝石可替代）", x)
	}
	*userCtx.EventCtx.DamageVal -= x
	if *userCtx.EventCtx.DamageVal < 0 {
		*userCtx.EventCtx.DamageVal = 0
	}
	rt.Log(fmt.Sprintf("%s 发动 [神之庇护]，消耗%d点水晶（可由红宝石替代）抵御了%d点士气下降", user.Name, x, x))

	rt.PopInterrupt()
	if !rt.HasPendingInterrupt() && rt.ResumePendingMoraleLoss(userCtx) {
		return nil
	}
	if !rt.HasPendingInterrupt() {
		rt.EnterResponseWindow()
	}
	return nil
}

func handleAngelBondHealChoice(rt engineplayer.ChoiceRuntime, playerID string, selectionIndex int, ctxData map[string]interface{}) error {
	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
	if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	targetID := targetIDs[selectionIndex]
	target := rt.LookupPlayer(targetID)
	if target == nil {
		return fmt.Errorf("目标不存在")
	}
	rt.Heal(target.ID, 1)
	userName := playerID
	if user := rt.LookupPlayer(playerID); user != nil {
		userName = user.Name
	}
	rt.Log(fmt.Sprintf("%s 的 [天使羁绊] 生效：%s 获得 +1 治疗", userName, target.Name))
	rt.PopInterrupt()
	if !rt.HasPendingInterrupt() {
		rt.ApplyChoiceResumePoint(ctxData["resume_phase"])
	}
	return nil
}

func buildPromptOptionsForPlayerIDs(players map[string]*model.Player, ids []string) []model.PromptOption {
	options := make([]model.PromptOption, 0, len(ids))
	for _, id := range ids {
		if p := players[id]; p != nil {
			options = append(options, model.PromptOption{
				ID:    id,
				Label: p.Name,
			})
		}
	}
	return options
}
