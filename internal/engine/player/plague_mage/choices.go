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
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【死亡之触】请选择弃置同系牌的元素：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	case "plague_death_touch_x":
		maxHeal := runtimeutil.ToIntContextValue(data["max_heal"])
		options := make([]model.PromptOption, 0, maxHeal-1)
		for x := 2; x <= maxHeal; x++ {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", x-2),
				Label: fmt.Sprintf("X=%d（移除%d点治疗）", x, x),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【死亡之触】请选择X值：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	case "plague_death_touch_y":
		maxCards := runtimeutil.ToIntContextValue(data["max_cards"])
		options := make([]model.PromptOption, 0, maxCards-1)
		for y := 2; y <= maxCards; y++ {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", y-2),
				Label: fmt.Sprintf("Y=%d（弃%d张同系牌）", y, y),
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【死亡之触】请选择Y值：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	case "plague_death_touch_cards":
		remaining := runtimeutil.ParseChoiceIntSlice(data["remaining_indices"])
		yNeed := runtimeutil.ToIntContextValue(data["y_value"])
		selectedCount := len(runtimeutil.ParseChoiceIntSlice(data["selected_indices"]))
		options := make([]model.PromptOption, 0, len(remaining))
		for _, idx := range remaining {
			if player == nil || idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, promptfmt.FormatCardInfo(player.Hand[idx])),
			})
		}
		remainingPick := yNeed - selectedCount
		if remainingPick < 1 {
			remainingPick = 1
		}
		if len(options) > 0 && remainingPick > len(options) {
			remainingPick = len(options)
		}
		return &model.Prompt{
			Type:     model.PromptChooseCards,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【死亡之触】请选择要弃置的%d张牌：", remainingPick),
			Options:  options,
			Min:      remainingPick,
			Max:      remainingPick,
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
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【死亡之触】请选择1名敌方角色承受法术伤害：",
			Options:  options,
			Min:      1,
			Max:      1,
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
	case "plague_death_touch_y":
		return true, handlePlagueDeathTouchYChoice(rt, ctxData, selectionIndex)
	case "plague_death_touch_cards":
		return true, handlePlagueDeathTouchCardsChoice(rt, ctxData, selectionIndex)
	case "plague_death_touch_target":
		return true, handlePlagueDeathTouchTargetChoice(rt, ctxData, selectionIndex)
	default:
		return false, nil
	}
}

func (choiceHandler) HandleCancel(rt engineplayer.ChoiceRuntime, playerID string, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "plague_death_touch_element", "plague_death_touch_x", "plague_death_touch_y",
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
	if intr := rt.GetPendingInterrupt(); intr != nil {
		intr.Context = map[string]interface{}{
			"choice_type":      "plague_death_touch_x",
			"user_id":          userID,
			"target_id":        ctxData["target_id"],
			"chosen_element":   chosenElement,
			"max_heal":         user.Heal,
			"max_cards":        len(getCardIndicesByElement(user, model.Element(chosenElement))),
			"selected_indices": []int{},
		}
	}
	rt.NotifyInterruptPrompt()
	return nil
}

func handlePlagueDeathTouchXChoice(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	xValue := selectionIndex + 2
	if maxHeal := runtimeutil.ToIntContextValue(ctxData["max_heal"]); xValue < 2 || xValue > maxHeal {
		return fmt.Errorf("无效的X值")
	}
	ctxData["choice_type"] = "plague_death_touch_y"
	ctxData["x_value"] = xValue
	if intr := rt.GetPendingInterrupt(); intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
	return nil
}

func handlePlagueDeathTouchYChoice(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}

	yValue := selectionIndex + 2
	if maxCards := runtimeutil.ToIntContextValue(ctxData["max_cards"]); yValue < 2 || yValue > maxCards {
		return fmt.Errorf("无效的Y值")
	}

	chosenElement, _ := ctxData["chosen_element"].(string)
	ctxData["choice_type"] = "plague_death_touch_cards"
	ctxData["y_value"] = yValue
	ctxData["selected_indices"] = []int{}
	ctxData["remaining_indices"] = getCardIndicesByElement(user, model.Element(chosenElement))
	if intr := rt.GetPendingInterrupt(); intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
	return nil
}

func handlePlagueDeathTouchCardsChoice(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	remaining := runtimeutil.ParseChoiceIntSlice(ctxData["remaining_indices"])
	selected := append([]int{}, runtimeutil.ParseChoiceIntSlice(ctxData["selected_indices"])...)
	yValue := runtimeutil.ToIntContextValue(ctxData["y_value"])

	cardIdx, ok := runtimeutil.ResolveSelectionToCandidate(selectionIndex, remaining)
	if !ok {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	selected = append(selected, cardIdx)

	nextRemaining := make([]int, 0, len(remaining))
	for _, idx := range remaining {
		if idx != cardIdx {
			nextRemaining = append(nextRemaining, idx)
		}
	}

	if len(selected) < yValue {
		ctxData["selected_indices"] = selected
		ctxData["remaining_indices"] = nextRemaining
		if intr := rt.GetPendingInterrupt(); intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return nil
	}

	ctxData["selected_indices"] = selected
	if targetID, _ := ctxData["target_id"].(string); targetID != "" {
		return resolvePlagueDeathTouchFinal(rt, ctxData, targetID)
	}

	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	ctxData["choice_type"] = "plague_death_touch_target"
	ctxData["target_ids"] = campEnemyIDs(rt, user)
	if intr := rt.GetPendingInterrupt(); intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
	return nil
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

	removed, err := removeCardsByIndicesFromHand(user, selected)
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

func getCardIndicesByElement(player *model.Player, element model.Element) []int {
	if player == nil {
		return nil
	}
	var out []int
	for i, c := range player.Hand {
		if c.Element == element {
			out = append(out, i)
		}
	}
	return out
}

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

func removeCardsByIndicesFromHand(player *model.Player, indices []int) ([]model.Card, error) {
	if player == nil {
		return nil, fmt.Errorf("玩家不存在")
	}
	for _, idx := range indices {
		if idx < 0 || idx >= len(player.Hand) {
			return nil, fmt.Errorf("无效的手牌索引: %d", idx)
		}
	}
	seen := map[int]bool{}
	for _, idx := range indices {
		if seen[idx] {
			return nil, fmt.Errorf("不能重复选择同一张牌")
		}
		seen[idx] = true
	}
	// 从大到小删除，避免索引位移。
	for i := 0; i < len(indices); i++ {
		for j := i + 1; j < len(indices); j++ {
			if indices[i] < indices[j] {
				indices[i], indices[j] = indices[j], indices[i]
			}
		}
	}
	var removed []model.Card
	for _, idx := range indices {
		removed = append(removed, player.Hand[idx])
		player.Hand = append(player.Hand[:idx], player.Hand[idx+1:]...)
	}
	return removed, nil
}

var _ engineplayer.CancelChoiceHandler = choiceHandler{}
