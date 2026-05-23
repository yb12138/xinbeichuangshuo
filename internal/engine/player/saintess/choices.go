// gameflow: 圣女角色选择流。

package saintess

import (
	"fmt"

	"starcup-engine/internal/engine/core/runtimeutil"
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

type choiceHandler struct{}

// NewChoiceHandler 创建圣女角色选择流处理器。
func NewChoiceHandler() engineplayer.ChoiceHandler {
	return choiceHandler{}
}

func (choiceHandler) BuildPrompt(rt engineplayer.ChoiceRuntime, choiceType, playerID string, _ *model.Player, data map[string]interface{}) *model.Prompt {
	if rt == nil {
		return nil
	}
	switch choiceType {
	case "frost_prayer_target":
		options := buildPromptOptionsForPlayerIDs(rt.GetPlayers(), runtimeutil.ParseStringSliceContextValue(data["target_ids"]))
		if len(options) == 0 {
			return nil
		}
		return &model.Prompt{
			Type:         model.PromptConfirm,
			PlayerID:     playerID,
			Message:      "【冰霜祷言】请选择1名角色获得+1治疗：",
			Options:      options,
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationTargetPicker, TargetFilter: "custom"},
		}
	default:
		return nil
	}
}

func (choiceHandler) HandleChoice(rt engineplayer.ChoiceRuntime, _ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	if rt == nil {
		return false, fmt.Errorf("圣女选择流运行时未初始化")
	}
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "frost_prayer_target":
		return true, handleFrostPrayerChoice(rt, selectionIndex, ctxData)
	default:
		return false, nil
	}
}

func handleFrostPrayerChoice(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}

	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
	if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	targetID := targetIDs[selectionIndex]
	target := rt.GetPlayers()[targetID]
	if target == nil {
		return fmt.Errorf("目标不存在")
	}

	rt.Heal(targetID, 1)
	rt.Log(fmt.Sprintf("%s 的 [冰霜祷言] 生效：%s +1治疗", user.Name, target.Name))
	rt.PopInterrupt()

	if rt.GetPendingInterrupt() == nil {
		if rawCtx, _ := ctxData["user_ctx"].(*model.Context); rawCtx != nil && rawCtx.ResumeAttackMissPhase() {
			if rt.ResumePendingAttackMiss(rawCtx) {
				return nil
			}
		}
	}
	return nil
}

func buildPromptOptionsForPlayerIDs(players map[string]*model.Player, targetIDs []string) []model.PromptOption {
	if len(targetIDs) == 0 || len(players) == 0 {
		return nil
	}
	options := make([]model.PromptOption, 0, len(targetIDs))
	for _, id := range targetIDs {
		if id == "" {
			continue
		}
		p := players[id]
		if p == nil {
			continue
		}
		label := p.Name
		if label == "" {
			label = p.ID
		}
		options = append(options, model.PromptOption{
			ID:       p.ID,
			Label:    label,
			TargetID: p.ID,
		})
	}
	return options
}
