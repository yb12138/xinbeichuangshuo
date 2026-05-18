// gameflow: 贤者角色选择流。

package sage

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
	case "sage_magic_rebound_confirm":
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【法术反弹】是否发动？",
			Options: []model.PromptOption{
				{ID: "0", Label: "是"},
				{ID: "1", Label: "否"},
			},
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationBranchSelect, Layout: "overlay"},
		}
	case "sage_magic_rebound_element":
		// 元素列表：只要有至少2张同系牌即可
		elements := availableElementsByMinCount(player, 2)
		options := make([]model.PromptOption, 0, len(elements))
		for idx, ele := range elements {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%s系", promptfmt.ElementName(ele)),
			})
		}
		return &model.Prompt{
			Type:         model.PromptConfirm,
			PlayerID:     playerID,
			Message:      "【法术反弹】请选择弃置同系牌的元素：",
			Options:      options,
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationBranchSelect, Layout: "overlay"},
		}
	case "sage_magic_rebound_cards":
		// 多选提示：显示所选元素的所有牌
		chosenElement, _ := data["chosen_element"].(string)
		cardIndices := engineplayer.GetCardIndicesByElement(player, model.Element(chosenElement))
		options := make([]model.PromptOption, 0, len(cardIndices))
		for _, idx := range cardIndices {
			card := player.Hand[idx]
			options = append(options, model.PromptOption{
				ID:     fmt.Sprintf("%d", idx),
				Label:  fmt.Sprintf("%d: %s", idx+1, promptfmt.FormatCardInfo(card)),
				CardID: card.ID,
			})
		}
		return &model.Prompt{
			Type:         model.PromptChooseCards,
			PlayerID:     playerID,
			Message:      "【法术反弹】请选择同系牌（选几张X即为几）：",
			Options:      options,
			Min:          2,
			Max:          len(cardIndices),
			Presentation: &model.PromptPresentation{Kind: model.PresentationCardPicker, CardSource: "hand", CardFilter: "same_element"},
		}
	case "sage_holy_cards":
		// 多选提示：显示所有手牌
		options := make([]model.PromptOption, 0, len(player.Hand))
		for idx, card := range player.Hand {
			options = append(options, model.PromptOption{
				ID:     fmt.Sprintf("%d", idx),
				Label:  fmt.Sprintf("%d: %s", idx+1, promptfmt.FormatCardInfo(card)),
				CardID: card.ID,
			})
		}
		return &model.Prompt{
			Type:         model.PromptChooseCards,
			PlayerID:     playerID,
			Message:      "【圣洁法典】请选择异系牌（选几张X即为几）：",
			Options:      options,
			Min:          3,
			Max:          len(player.Hand),
			Presentation: &model.PromptPresentation{Kind: model.PresentationCardPicker, CardSource: "hand", CardFilter: "distinct_elements"},
		}
	case "sage_arcane_cards":
		options := make([]model.PromptOption, 0, len(player.Hand))
		for idx, card := range player.Hand {
			options = append(options, model.PromptOption{
				ID:     fmt.Sprintf("%d", idx),
				Label:  fmt.Sprintf("%d: %s", idx+1, promptfmt.FormatCardInfo(card)),
				CardID: card.ID,
			})
		}
		return &model.Prompt{
			Type:         model.PromptChooseCards,
			PlayerID:     playerID,
			Message:      "【魔道法典】请选择异系牌（选几张X即为几）：",
			Options:      options,
			Min:          2,
			Max:          len(player.Hand),
			Presentation: &model.PromptPresentation{Kind: model.PresentationCardPicker, CardSource: "hand"},
		}
	case "sage_holy_target_count":
		maxCount := runtimeutil.ToIntContextValue(data["max_target_count"])
		if maxCount < 1 {
			maxCount = 1
		}
		options := make([]model.PromptOption, 0, maxCount)
		for count := 1; count <= maxCount; count++ {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", count-1),
				Label: fmt.Sprintf("选择%d名角色", count),
			})
		}
		return &model.Prompt{
			Type:         model.PromptConfirm,
			PlayerID:     playerID,
			Message:      "【圣洁法典】请选择要获得治疗的角色数量：",
			Options:      options,
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationNumeric, NumericBase: 0},
		}
	case "sage_holy_targets":
		allTargetIDs := runtimeutil.ParseStringSliceContextValue(data["target_ids"])
		maxTargetCount := runtimeutil.ToIntContextValue(data["max_target_count"])
		if maxTargetCount < 1 {
			maxTargetCount = runtimeutil.ToIntContextValue(data["target_count"])
		}
		if maxTargetCount < 1 {
			maxTargetCount = 1
		}
		options := make([]model.PromptOption, 0, len(allTargetIDs))
		for _, targetID := range allTargetIDs {
			if target := rt.GetPlayers()[targetID]; target != nil {
				options = append(options, model.PromptOption{ID: targetID, Label: target.Name})
			}
		}
		return &model.Prompt{
			Type:         model.PromptConfirm,
			PlayerID:     playerID,
			Message:      fmt.Sprintf("【圣洁法典】请选择治疗目标（1-%d名）：", maxTargetCount),
			Options:      options,
			Min:          1,
			Max:          maxTargetCount,
			Presentation: &model.PromptPresentation{Kind: model.PresentationTargetPicker, TargetFilter: "custom", MultiTarget: true},
		}
	case "sage_magic_rebound_target", "sage_arcane_target":
		targetIDs := runtimeutil.ParseStringSliceContextValue(data["target_ids"])
		options := make([]model.PromptOption, 0, len(targetIDs))
		for _, targetID := range targetIDs {
			if target := rt.GetPlayers()[targetID]; target != nil {
				options = append(options, model.PromptOption{ID: targetID, Label: target.Name})
			}
		}
		msg := "请选择目标角色："
		if choiceType == "sage_magic_rebound_target" {
			msg = "【法术反弹】请选择目标角色："
		} else {
			msg = "【魔道法典】请选择目标角色："
		}
		return &model.Prompt{
			Type:         model.PromptConfirm,
			PlayerID:     playerID,
			Message:      msg,
			Options:      options,
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationTargetPicker, TargetFilter: "custom"},
		}
	}

	return nil
}

func (choiceHandler) HandleChoice(rt engineplayer.ChoiceRuntime, _ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)

	switch choiceType {
	case "sage_magic_rebound_confirm":
		userID, _ := ctxData["user_id"].(string)
		user := rt.GetPlayers()[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		if selectionIndex == 1 {
			rt.PopInterrupt()
			if rt.GetPendingInterrupt() == nil && len(rt.GetPendingDamageQueue()) > 0 {
				rt.EnterDamageResolution(nil)
			}
			return true, nil
		}
		if selectionIndex != 0 {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		// 检查是否有至少2张同系牌
		maxX := engineplayer.MaxSameElementCount(user)
		if maxX < 2 {
			return true, fmt.Errorf("同系手牌不足2张，无法发动法术反弹")
		}
		// 直接进入元素选择，不再选择X
		ctxData["choice_type"] = "sage_magic_rebound_element"
		if intr := rt.GetPendingInterrupt(); intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return true, nil

	case "sage_magic_rebound_element":
		userID, _ := ctxData["user_id"].(string)
		user := rt.GetPlayers()[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		elements := availableElementsByMinCount(user, 2)
		if selectionIndex < 0 || selectionIndex >= len(elements) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		chosenElement := model.Element(elements[selectionIndex])
		ctxData["chosen_element"] = string(chosenElement)
		ctxData["choice_type"] = "sage_magic_rebound_cards"
		if intr := rt.GetPendingInterrupt(); intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return true, nil

	case "sage_holy_target_count":
		targetCount := selectionIndex + 1
		maxCount := runtimeutil.ToIntContextValue(ctxData["max_target_count"])
		if targetCount < 1 || targetCount > maxCount {
			return true, fmt.Errorf("无效的治疗目标数量")
		}
		ctxData["target_count"] = targetCount
		ctxData["max_target_count"] = targetCount
		ctxData["choice_type"] = "sage_holy_targets"
		if intr := rt.GetPendingInterrupt(); intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return true, nil

	case "sage_holy_targets":
		return resolveHolyCodexTargets(rt, ctxData, []int{selectionIndex})

	case "sage_magic_rebound_target", "sage_arcane_target":
		userID, _ := ctxData["user_id"].(string)
		user := rt.GetPlayers()[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		selected := engineplayer.ParseIntSliceContextValue(ctxData["selected_indices"])
		xValue := runtimeutil.ToIntContextValue(ctxData["x_value"])
		if xValue <= 1 || len(selected) != xValue {
			return true, fmt.Errorf("弃牌参数无效")
		}
		removed, err := engineplayer.RemoveCardsByIndicesFromHand(user, append([]int{}, selected...))
		if err != nil {
			return true, err
		}
		rt.NotifyCardRevealed(user.ID, removed, "discard")
		rt.AppendToDiscard(removed)
		targetID := targetIDs[selectionIndex]
		target := rt.GetPlayers()[targetID]
		if target == nil {
			return true, fmt.Errorf("目标不存在")
		}

		switch choiceType {
		case "sage_magic_rebound_target":
			damageToTarget := xValue - 1
			damageToSelf := xValue
			pending := make([]model.PendingDamage, 0, 2)
			if damageToTarget > 0 {
				pending = append(pending, model.PendingDamage{
					SourceID:   user.ID,
					TargetID:   targetID,
					Damage:     damageToTarget,
					DamageType: model.MagicAttack,
				})
			}
			if damageToSelf > 0 {
				pending = append(pending, model.PendingDamage{
					SourceID:   user.ID,
					TargetID:   user.ID,
					Damage:     damageToSelf,
					DamageType: model.MagicAttack,
				})
			}
			prependPendingDamages(rt, pending)
			rt.Log(fmt.Sprintf("%s 发动 [法术反弹]：弃%d张同系牌，对 %s 造成%d点法术伤害，并对自己造成%d点法术伤害", user.Name, xValue, target.Name, damageToTarget, damageToSelf))
		case "sage_arcane_target":
			damage := xValue - 1
			if damage > 0 {
				rt.AddPendingDamage(model.PendingDamage{
					SourceID:   user.ID,
					TargetID:   targetID,
					Damage:     damage,
					DamageType: model.MagicAttack,
				})
				rt.AddPendingDamage(model.PendingDamage{
					SourceID:   user.ID,
					TargetID:   user.ID,
					Damage:     damage,
					DamageType: model.MagicAttack,
				})
			}
			if targetID == user.ID {
				rt.Log(fmt.Sprintf("%s 发动 [魔道法典]：弃%d张异系牌，对自己造成%d点法术伤害（目标为自己，两次各1点）", user.Name, xValue, damage))
			} else {
				rt.Log(fmt.Sprintf("%s 发动 [魔道法典]：弃%d张异系牌，对 %s 与自己各造成%d点法术伤害", user.Name, xValue, target.Name, damage))
			}
		}

		rt.PopInterrupt()
		if rt.GetPendingInterrupt() == nil {
			if !rt.RoutePendingDamageWithReturn(model.TurnStageExtraAction) {
				rt.EnterExtraActionStage()
			}
		}
		return true, nil
	}

	return false, nil
}

// handleArcaneCardsMultiSelect 处理魔道法典异系牌多选。
func handleArcaneCardsMultiSelect(rt engineplayer.ChoiceRuntime, playerID string, selections []int, ctxData map[string]interface{}) (bool, error) {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return false, fmt.Errorf("玩家不存在")
	}
	if len(selections) < 2 {
		return false, fmt.Errorf("魔道法典至少需要选择2张异系牌")
	}
	// 验证所选牌元素互不相同
	seenElements := map[model.Element]bool{}
	for _, idx := range selections {
		if idx < 0 || idx >= len(user.Hand) {
			return false, fmt.Errorf("无效的选项索引: %d", idx)
		}
		card := user.Hand[idx]
		if seenElements[card.Element] {
			return false, fmt.Errorf("魔道法典需弃置异系牌，不能重复选择同系")
		}
		seenElements[card.Element] = true
	}
	ctxData["selected_indices"] = selections
	ctxData["x_value"] = len(selections)
	ctxData["choice_type"] = "sage_arcane_target"
	ctxData["target_ids"] = append([]string{}, rt.GetPlayerOrder()...)
	if intr := rt.GetPendingInterrupt(); intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
	return true, nil
}

// handleReboundCardsMultiSelect 处理法术反弹同系牌多选。
func handleReboundCardsMultiSelect(rt engineplayer.ChoiceRuntime, playerID string, selections []int, ctxData map[string]interface{}) (bool, error) {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return false, fmt.Errorf("玩家不存在")
	}
	if len(selections) < 2 {
		return false, fmt.Errorf("法术反弹至少需要选择2张同系牌")
	}
	// 验证所选牌都是同一元素
	chosenElement, _ := ctxData["chosen_element"].(string)
	for _, idx := range selections {
		if idx < 0 || idx >= len(user.Hand) {
			return false, fmt.Errorf("无效的选项索引: %d", idx)
		}
		card := user.Hand[idx]
		if string(card.Element) != chosenElement {
			return false, fmt.Errorf("法术反弹需弃置同系牌")
		}
	}
	ctxData["selected_indices"] = selections
	ctxData["x_value"] = len(selections)
	ctxData["choice_type"] = "sage_magic_rebound_target"
	ctxData["target_ids"] = append([]string{}, rt.GetPlayerOrder()...)
	if intr := rt.GetPendingInterrupt(); intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
	return true, nil
}

// handleHolyCardsMultiSelect 处理圣洁法典异系牌多选。
func handleHolyCardsMultiSelect(rt engineplayer.ChoiceRuntime, playerID string, selections []int, ctxData map[string]interface{}) (bool, error) {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return false, fmt.Errorf("玩家不存在")
	}
	if len(selections) < 3 {
		return false, fmt.Errorf("圣洁法典至少需要选择3张异系牌")
	}
	// 验证所选牌元素互不相同
	seenElements := map[model.Element]bool{}
	for _, idx := range selections {
		if idx < 0 || idx >= len(user.Hand) {
			return false, fmt.Errorf("无效的选项索引: %d", idx)
		}
		card := user.Hand[idx]
		if seenElements[card.Element] {
			return false, fmt.Errorf("圣洁法典需弃置异系牌，不能重复选择同系")
		}
		seenElements[card.Element] = true
	}
	ctxData["selected_indices"] = selections
	ctxData["x_value"] = len(selections)
	maxTargetCount := len(selections) - 2
	if maxTargetCount < 1 {
		return false, fmt.Errorf("圣洁法典治疗目标数量无效")
	}
	ctxData["choice_type"] = "sage_holy_targets"
	ctxData["max_target_count"] = maxTargetCount
	ctxData["target_ids"] = append([]string{}, rt.GetPlayerOrder()...)
	if intr := rt.GetPendingInterrupt(); intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
	return true, nil
}

// handleHolyTargetsMultiSelect 处理圣洁法典的多目标治疗选择。
func handleHolyTargetsMultiSelect(rt engineplayer.ChoiceRuntime, _ string, selections []int, ctxData map[string]interface{}) (bool, error) {
	return resolveHolyCodexTargets(rt, ctxData, selections)
}

func resolveHolyCodexTargets(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selections []int) (bool, error) {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return true, fmt.Errorf("玩家不存在")
	}

	maxTargetCount := runtimeutil.ToIntContextValue(ctxData["max_target_count"])
	if maxTargetCount < 1 {
		maxTargetCount = runtimeutil.ToIntContextValue(ctxData["target_count"])
	}
	if maxTargetCount < 1 {
		return true, fmt.Errorf("圣洁法典治疗目标数量无效")
	}
	if len(selections) < 1 || len(selections) > maxTargetCount {
		return true, fmt.Errorf("圣洁法典治疗目标数量需为1-%d名", maxTargetCount)
	}

	allTargetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
	selectedTargets := make([]string, 0, len(selections))
	seenTargets := map[string]bool{}
	for _, selectionIndex := range selections {
		if selectionIndex < 0 || selectionIndex >= len(allTargetIDs) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		targetID := allTargetIDs[selectionIndex]
		if seenTargets[targetID] {
			return true, fmt.Errorf("圣洁法典治疗目标不能重复")
		}
		if rt.GetPlayers()[targetID] == nil {
			return true, fmt.Errorf("目标不存在")
		}
		seenTargets[targetID] = true
		selectedTargets = append(selectedTargets, targetID)
	}

	selectedCards := engineplayer.ParseIntSliceContextValue(ctxData["selected_indices"])
	xValue := runtimeutil.ToIntContextValue(ctxData["x_value"])
	if xValue <= 2 || len(selectedCards) != xValue {
		return true, fmt.Errorf("圣洁法典弃牌参数无效")
	}
	removed, err := engineplayer.RemoveCardsByIndicesFromHand(user, append([]int{}, selectedCards...))
	if err != nil {
		return true, err
	}
	rt.NotifyCardRevealed(user.ID, removed, "discard")
	rt.AppendToDiscard(removed)
	for _, targetID := range selectedTargets {
		rt.Heal(targetID, 2)
	}
	damage := xValue - 1
	if damage > 0 {
		rt.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   user.ID,
			Damage:     damage,
			DamageType: model.MagicAttack,
		})
	}
	rt.Log(fmt.Sprintf("%s 发动 [圣洁法典]：为%d名角色各+2治疗，并对自己造成%d点法术伤害", user.Name, len(selectedTargets), damage))
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		if !rt.RoutePendingDamageWithReturn(model.TurnStageExtraAction) {
			rt.EnterExtraActionStage()
		}
	}
	return true, nil
}

// Helper functions for sage

func buildElementCardIndexMap(player *model.Player) map[model.Element][]int {
	if player == nil {
		return nil
	}
	out := map[model.Element][]int{}
	for i, c := range player.Hand {
		if c.Element != "" {
			out[c.Element] = append(out[c.Element], i)
		}
	}
	return out
}

func availableElementsByMinCount(player *model.Player, minCount int) []string {
	if minCount <= 0 {
		minCount = 1
	}
	elemMap := buildElementCardIndexMap(player)
	var out []string
	for _, ele := range engineplayer.ElementOrderForPrompt() {
		if len(elemMap[ele]) >= minCount {
			out = append(out, string(ele))
		}
	}
	return out
}

func prependPendingDamages(rt engineplayer.ChoiceRuntime, pending []model.PendingDamage) {
	for _, pd := range pending {
		rt.AddPendingDamageFront(pd)
	}
}
