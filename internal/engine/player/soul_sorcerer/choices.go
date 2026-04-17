// gameflow: 灵魂术士角色选择流。

package soul_sorcerer

import (
	"fmt"
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

type choiceHandler struct{}

func NewChoiceHandler() engineplayer.ChoiceHandler {
	return choiceHandler{}
}

func (choiceHandler) BuildPrompt(rt engineplayer.ChoiceRuntime, choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "ss_convert_color":
		var modeOrder []string
		if arr, ok := data["mode_order"].([]string); ok {
			modeOrder = append(modeOrder, arr...)
		} else if arr, ok := data["mode_order"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					modeOrder = append(modeOrder, s)
				}
			}
		}
		var options []model.PromptOption
		for _, mode := range modeOrder {
			switch mode {
			case "y2b":
				options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "黄魂 -> 蓝魂（转换1点）"})
			case "b2y":
				options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "蓝魂 -> 黄魂（转换1点）"})
			}
		}
		return &model.Prompt{
			Type:       model.PromptConfirm,
			PlayerID:   playerID,
			ChoiceType: choiceType,
			Message:    "【灵魂转换】请选择转换方向：",
			Options:    options,
			Min:        1,
			Max:        1,
		}
	case "ss_link_target":
		var allyIDs []string
		if arr, ok := data["ally_ids"].([]string); ok {
			allyIDs = append(allyIDs, arr...)
		} else if arr, ok := data["ally_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					allyIDs = append(allyIDs, s)
				}
			}
		}
		var options []model.PromptOption
		for _, aid := range allyIDs {
			if p := rt.LookupPlayer(aid); p != nil {
				options = append(options, model.PromptOption{ID: aid, Label: p.Name})
			}
		}
		return &model.Prompt{
			Type:       model.PromptConfirm,
			PlayerID:   playerID,
			ChoiceType: choiceType,
			Message:    "【灵魂链接】请选择要放置灵魂链接的队友：",
			Options:    options,
			Min:        1,
			Max:        1,
		}
	case "ss_recall_pick":
		var magicIndices []int
		if arr, ok := data["magic_indices"].([]int); ok {
			magicIndices = append(magicIndices, arr...)
		} else if arr, ok := data["magic_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					magicIndices = append(magicIndices, int(f))
				}
			}
		}
		if len(magicIndices) == 0 {
			if arr, ok := data["remaining_indices"].([]int); ok {
				magicIndices = append(magicIndices, arr...)
			} else if arr, ok := data["remaining_indices"].([]interface{}); ok {
				for _, v := range arr {
					if f, ok := v.(float64); ok {
						magicIndices = append(magicIndices, int(f))
					}
				}
			}
		}
		var options []model.PromptOption
		for _, idx := range magicIndices {
			if player == nil || idx < 0 || idx >= len(player.Hand) {
				continue
			}
			if player.Hand[idx].Type != model.CardTypeMagic {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, player.Hand[idx].Name),
			})
		}
		maxSelect := len(options)
		if maxSelect < 1 {
			maxSelect = 1
		}
		return &model.Prompt{
			Type:       model.PromptChooseCards,
			PlayerID:   playerID,
			ChoiceType: choiceType,
			Message:    "【灵魂召还】请选择要弃置的法术牌（至少1张）：",
			Options:    options,
			Min:        1,
			Max:        maxSelect,
		}
	}
	return nil
}

func (choiceHandler) HandleChoice(rt engineplayer.ChoiceRuntime, _ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "ss_convert_color":
		return true, handleSoulConvertColorChoice(rt, selectionIndex, ctxData)
	case "ss_link_target":
		return true, handleSoulLinkTargetChoice(rt, selectionIndex, ctxData)
	case "ss_recall_pick":
		return true, handleSoulRecallPickChoice(rt, selectionIndex, ctxData)
	default:
		return false, nil
	}
}

func handleSoulConvertColorChoice(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	return fmt.Errorf("soul_sorcerer choice handler requires full engine access - temporarily disabled")
}

func handleSoulLinkTargetChoice(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	return fmt.Errorf("soul_sorcerer choice handler requires full engine access - temporarily disabled")
}

func handleSoulRecallPickChoice(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	return fmt.Errorf("soul_sorcerer choice handler requires full engine access - temporarily disabled")
}
