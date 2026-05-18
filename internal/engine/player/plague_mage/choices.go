// gameflow: 瘟疫术士角色选择流。

package plague_mage

import (
	"fmt"

	"starcup-engine/internal/engine/core/runtimeutil"
	"starcup-engine/internal/engine/hook/promptfmt"
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

type choiceHandler struct{}

func NewChoiceHandler() engineplayer.ChoiceHandler {
	return choiceHandler{}
}

func (choiceHandler) BuildPrompt(rt engineplayer.ChoiceRuntime, choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "plague_death_touch_element":
		elements := runtimeutil.ParseStringSliceContextValue(data["elements"])
		options := make([]model.PromptOption, 0, len(elements))
		for i, ele := range elements {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", i),
				Label: fmt.Sprintf("%s系", promptfmt.ElementName(ele)),
			})
		}
		return &model.Prompt{
			Type:         model.PromptConfirm,
			PlayerID:     playerID,
			Message:      "【死亡之触】请选择弃置同系牌的元素：",
			Options:      options,
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationBranchSelect, Layout: "overlay"},
		}
	case "plague_death_touch_x":
		maxHeal := runtimeutil.ToIntContextValue(data["max_heal"])
		options := make([]model.PromptOption, 0, maxHeal-1)
		for x := 2; x <= maxHeal; x++ {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", x),
				Label: fmt.Sprintf("X=%d（移除%d点治疗）", x, x),
			})
		}
		return &model.Prompt{
			Type:         model.PromptConfirm,
			PlayerID:     playerID,
			Message:      "【死亡之触】请选择X值：",
			Options:      options,
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationNumeric, NumericBase: 0},
		}
	case "plague_death_touch_cards":
		chosenElement, _ := data["chosen_element"].(string)
		cardIndices := engineplayer.GetCardIndicesByElement(player, model.Element(chosenElement))
		options := make([]model.PromptOption, 0, len(cardIndices))
		for _, idx := range cardIndices {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, promptfmt.FormatCardInfo(player.Hand[idx])),
			})
		}
		return &model.Prompt{
			Type:         model.PromptChooseCards,
			PlayerID:     playerID,
			Message:      "【死亡之触】请选择同系牌（选几张Y即为几）：",
			Options:      options,
			Min:          2,
			Max:          len(cardIndices),
			Presentation: &model.PromptPresentation{Kind: model.PresentationCardPicker, CardSource: "hand"},
		}
	case "plague_death_touch_target":
		targetIDs := runtimeutil.ParseStringSliceContextValue(data["target_ids"])
		options := make([]model.PromptOption, 0, len(targetIDs))
		for i, targetID := range targetIDs {
			if target := rt.GetPlayers()[targetID]; target != nil {
				options = append(options, model.PromptOption{
					ID:    fmt.Sprintf("%d", i),
					Label: target.Name,
				})
			}
		}
		return &model.Prompt{
			Type:         model.PromptConfirm,
			PlayerID:     playerID,
			Message:      "【死亡之触】请选择1名敌方角色承受法术伤害：",
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
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "plague_death_touch_element":
		return true, handlePlagueDeathTouchElementChoice(rt, ctxData, selectionIndex)
	case "plague_death_touch_x":
		return true, handlePlagueDeathTouchXChoice(rt, ctxData, selectionIndex)
	case "plague_death_touch_target":
		return true, handlePlagueDeathTouchTargetChoice(rt, ctxData, selectionIndex)
	default:
		return false, nil
	}
}

func (choiceHandler) HandleCancel(rt engineplayer.ChoiceRuntime, playerID string, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "plague_death_touch_element", "plague_death_touch_x",
		"plague_death_touch_cards", "plague_death_touch_target":
		return true, cancelPlagueDeathTouchChoice(rt, playerID, ctxData)
	default:
		return false, nil
	}
}

func handlePlagueDeathTouchElementChoice(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}

	elements := runtimeutil.ParseStringSliceContextValue(ctxData["elements"])
	if selectionIndex < 0 || selectionIndex >= len(elements) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}

	chosenElement := elements[selectionIndex]
	ctxData["chosen_element"] = chosenElement
	ctxData["choice_type"] = "plague_death_touch_x"
	if intr := rt.GetPendingInterrupt(); intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
	return nil
}

func handlePlagueDeathTouchXChoice(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	xValue := selectionIndex + 2
	if maxHeal := runtimeutil.ToIntContextValue(ctxData["max_heal"]); xValue < 2 || xValue > maxHeal {
		return fmt.Errorf("无效的X值")
	}
	ctxData["x_value"] = xValue
	ctxData["choice_type"] = "plague_death_touch_cards"
	if intr := rt.GetPendingInterrupt(); intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
	return nil
}

// handleDeathTouchCardsMultiSelect 处理死亡之触同系牌多选。
func handleDeathTouchCardsMultiSelect(rt engineplayer.ChoiceRuntime, playerID string, selections []int, ctxData map[string]interface{}) (bool, error) {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return false, fmt.Errorf("玩家不存在")
	}
	if len(selections) < 2 {
		return false, fmt.Errorf("死亡之触至少需要选择2张同系牌")
	}
	chosenElement, _ := ctxData["chosen_element"].(string)
	for _, idx := range selections {
		if idx < 0 || idx >= len(user.Hand) {
			return false, fmt.Errorf("无效的选项索引: %d", idx)
		}
		card := user.Hand[idx]
		if string(card.Element) != chosenElement {
			return false, fmt.Errorf("死亡之触需弃置同系牌")
		}
	}
	ctxData["selected_indices"] = selections
	ctxData["y_value"] = len(selections)
	if targetID, _ := ctxData["target_id"].(string); targetID != "" {
		err := resolvePlagueDeathTouchFinal(rt, ctxData, targetID)
		return err == nil, err
	}
	ctxData["choice_type"] = "plague_death_touch_target"
	ctxData["target_ids"] = campEnemyIDs(rt, user)
	if intr := rt.GetPendingInterrupt(); intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
	return true, nil
}

func handlePlagueDeathTouchTargetChoice(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
	if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	return resolvePlagueDeathTouchFinal(rt, ctxData, targetIDs[selectionIndex])
}

func resolvePlagueDeathTouchFinal(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, targetID string) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	if rt.GetPlayers()[targetID] == nil {
		return fmt.Errorf("目标不存在")
	}

	selected := append([]int{}, runtimeutil.ParseChoiceIntSlice(ctxData["selected_indices"])...)
	xValue := runtimeutil.ToIntContextValue(ctxData["x_value"])
	yValue := runtimeutil.ToIntContextValue(ctxData["y_value"])

	removed, err := engineplayer.RemoveCardsByIndicesFromHand(user, selected)
	if err != nil {
		return err
	}
	rt.NotifyCardRevealed(user.ID, removed, "discard")
	rt.AppendToDiscard(removed)

	if user.Heal < xValue {
		return fmt.Errorf("治疗不足，无法移除X=%d", xValue)
	}
	user.Heal -= xValue

	damage := xValue + yValue - 3
	if damage < 0 {
		damage = 0
	}
	user.TurnState.UsedSkillCounts["plague_block_immortal"] = 1
	user.TurnState.HasActed = true
	user.TurnState.LastActionType = string(model.ActionMagic)
	user.TurnState.LastActionCard = nil
	rt.AddPendingDamage(model.PendingDamage{
		SourceID:           user.ID,
		TargetID:           targetID,
		Damage:             damage,
		DamageType:         model.MagicAttack,
		CapDrawToHandLimit: true,
	})

	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		if !rt.RoutePendingDamageWithReturn(model.TurnStageExtraAction) {
			rt.EnterExtraActionStage()
		}
	}
	return nil
}

func cancelPlagueDeathTouchChoice(rt engineplayer.ChoiceRuntime, playerID string, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	if userID == "" {
		userID = playerID
	}
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}

	user.TurnState.UsedSkillCounts["plague_block_immortal"] = 0
	user.TurnState.HasActed = false
	user.TurnState.LastActionType = ""
	user.TurnState.LastActionCard = nil

	rt.PopInterrupt()
	rt.Log(fmt.Sprintf("%s 取消了 [死亡之触] 的发动", user.Name))
	if rt.GetPendingInterrupt() == nil {
		rt.EnterActionExecutionStage()
	}
	return nil
}

// Helper functions for plague_mage

func campEnemyIDs(rt engineplayer.ChoiceRuntime, user *model.Player) []string {
	if user == nil {
		return nil
	}
	var ids []string
	for _, pid := range rt.GetPlayerOrder() {
		p := rt.GetPlayers()[pid]
		if p == nil || p.Camp == user.Camp {
			continue
		}
		ids = append(ids, p.ID)
	}
	return ids
}

var _ engineplayer.CancelChoiceHandler = choiceHandler{}