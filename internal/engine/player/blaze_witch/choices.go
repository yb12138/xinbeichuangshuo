// gameflow: 烈焰魔女角色选择流。

package blaze_witch

import (
	"fmt"

	"starcup-engine/internal/engine/core/runtimeutil"
	"starcup-engine/internal/engine/hook/promptfmt"
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

type choiceHandler struct{}

const (
	manaInversionFlowID     = "bw_mana_inversion"
	manaInversionStepX      = "x"
	manaInversionStepCards  = "cards"
	manaInversionStepTarget = "target"

	substituteDollFlowID     = "bw_substitute_doll"
	substituteDollStepCard   = "card"
	substituteDollStepTarget = "target"
)

var manaInversionFlowRuntime = model.MustNewPromptFlowRuntime(manaInversionFlowID, []model.PromptFlowStepSpec{
	{ID: manaInversionStepX, ChoiceType: "bw_mana_inversion_x", CancelPolicy: model.CancelPolicyBack},
	{ID: manaInversionStepCards, ChoiceType: "bw_mana_inversion_cards", CancelPolicy: model.CancelPolicyBack},
	{ID: manaInversionStepTarget, ChoiceType: "bw_mana_inversion_target", CancelPolicy: model.CancelPolicyAbort},
})

var substituteDollFlowRuntime = model.MustNewPromptFlowRuntime(substituteDollFlowID, []model.PromptFlowStepSpec{
	{ID: substituteDollStepCard, ChoiceType: "bw_substitute_doll_card", CancelPolicy: model.CancelPolicyAbort},
	{ID: substituteDollStepTarget, ChoiceType: "bw_substitute_doll_target", CancelPolicy: model.CancelPolicyBack},
})

func NewChoiceHandler() engineplayer.ChoiceHandler {
	return choiceHandler{}
}

func (choiceHandler) BuildPrompt(rt engineplayer.ChoiceRuntime, choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "bw_witch_wrath_draw":
		return &model.Prompt{
			Type:       model.PromptConfirm,
			PlayerID:   playerID,
			Message:    "【魔女之怒】请选择摸牌数量：",
			ChoiceType: choiceType,
			Options: []model.PromptOption{
				{ID: "0", Label: "摸0张"},
				{ID: "1", Label: "摸1张"},
				{ID: "2", Label: "摸2张"},
			},
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationNumeric, NumericBase: 0},
		}
	case "bw_substitute_doll_card":
		magicIndices := runtimeutil.ParseChoiceIntSlice(data["magic_indices"])
		options := make([]model.PromptOption, 0, len(magicIndices))
		for _, idx := range magicIndices {
			if player == nil || idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, promptfmt.FormatCardInfo(player.Hand[idx])),
			})
		}
		return &model.Prompt{
			Type:         model.PromptConfirm,
			PlayerID:     playerID,
			Message:      "【替身玩偶】请选择弃置1张法术牌：",
			Options:      options,
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationCardPicker, CardSource: "hand"},
		}
	case "bw_mana_inversion_x":
		maxX := runtimeutil.ToIntContextValue(data["max_x"])
		options := make([]model.PromptOption, 0, maxX-1)
		for x := 2; x <= maxX; x++ {
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", x),
				Label: fmt.Sprintf("X=%d（弃%d张法术牌，造成%d点法术伤害）", x, x, x-1),
			})
		}
		return &model.Prompt{
			Type:         model.PromptConfirm,
			PlayerID:     playerID,
			Message:      "【魔能反转】请选择X值：",
			Options:      options,
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationNumeric, NumericBase: 0},
		}
	case "bw_mana_inversion_cards":
		remaining := runtimeutil.ParseChoiceIntSlice(data["remaining_indices"])
		flow, err := model.RequirePromptFlow(data, manaInversionFlowID, "魔能反转")
		if err != nil {
			return nil
		}
		selectedCount := len(flow.Selection(manaInversionStepCards).OptionIndexes)
		targetCount := flow.Selection(manaInversionStepX).Count
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
		remainingPick := targetCount - selectedCount
		if remainingPick < 1 {
			remainingPick = 1
		}
		if len(options) > 0 && remainingPick > len(options) {
			remainingPick = len(options)
		}
		return &model.Prompt{
			Type:         model.PromptChooseCards,
			PlayerID:     playerID,
			Message:      fmt.Sprintf("【魔能反转】请选择要弃置的%d张法术牌：", remainingPick),
			Options:      options,
			Min:          remainingPick,
			Max:          remainingPick,
			Presentation: &model.PromptPresentation{Kind: model.PresentationCardPicker, CardSource: "hand"},
		}
	case "bw_substitute_doll_target":
		return &model.Prompt{
			Type:         model.PromptConfirm,
			PlayerID:     playerID,
			Message:      "【替身玩偶】请选择摸1张牌的队友：",
			Options:      buildPromptOptionsForPlayerIDs(rt.GetPlayers(), runtimeutil.ParseStringSliceContextValue(data["target_ids"])),
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationTargetPicker, TargetFilter: "custom"},
		}
	case "bw_mana_inversion_target":
		return &model.Prompt{
			Type:         model.PromptConfirm,
			PlayerID:     playerID,
			Message:      "【魔能反转】请选择法术伤害目标：",
			Options:      buildPromptOptionsForPlayerIDs(rt.GetPlayers(), runtimeutil.ParseStringSliceContextValue(data["target_ids"])),
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationTargetPicker, TargetFilter: "custom"},
		}
	default:
		return nil
	}
}

func (choiceHandler) HandleChoice(rt engineplayer.ChoiceRuntime, playerID string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "bw_witch_wrath_draw":
		return true, handleBlazeWitchWrathDrawChoice(rt, playerID, selectionIndex, ctxData)
	case "bw_substitute_doll_card":
		return true, handleBlazeWitchSubstituteCardChoice(rt, playerID, selectionIndex, ctxData)
	case "bw_mana_inversion_x":
		return true, handleBlazeWitchManaInversionXChoice(rt, playerID, selectionIndex, ctxData)
	case "bw_mana_inversion_cards":
		return true, handleBlazeWitchManaInversionCardsChoice(rt, playerID, selectionIndex, ctxData)
	case "bw_substitute_doll_target", "bw_mana_inversion_target":
		return true, handleBlazeWitchTargetChoice(rt, playerID, selectionIndex, ctxData)
	default:
		return false, nil
	}
}

func handleBlazeWitchWrathDrawChoice(rt engineplayer.ChoiceRuntime, playerID string, selectionIndex int, ctxData map[string]interface{}) error {
	user := rt.GetPlayers()[playerID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	if selectionIndex < 0 || selectionIndex > 2 {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	if selectionIndex > 0 {
		rt.DrawCards(user.ID, selectionIndex)
	}
	rt.Log(fmt.Sprintf("%s 的 [魔女之怒]：选择摸%d张牌", user.Name, selectionIndex))
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		rt.ApplyChoiceResumePoint(ctxData["resume_phase"])
	}
	return nil
}

func handleBlazeWitchSubstituteCardChoice(rt engineplayer.ChoiceRuntime, playerID string, selectionIndex int, ctxData map[string]interface{}) error {
	user := rt.GetPlayers()[playerID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	magicIndices := runtimeutil.ParseChoiceIntSlice(ctxData["magic_indices"])
	cardIdx, ok := runtimeutil.ResolveSelectionToCandidate(selectionIndex, magicIndices)
	if !ok || cardIdx < 0 || cardIdx >= len(user.Hand) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	if user.Hand[cardIdx].Type != model.CardTypeMagic {
		return fmt.Errorf("替身玩偶需弃置法术牌")
	}
	flow, err := model.RequirePromptFlow(ctxData, substituteDollFlowID, "替身玩偶")
	if err != nil {
		return err
	}
	flow.PutSelection(substituteDollStepCard, model.PromptFlowSelection{
		OptionIndexes: []int{selectionIndex},
		CardIDs:       []string{user.Hand[cardIdx].ID},
	})
	ctxData["target_ids"] = runtimeutil.ParseStringSliceContextValue(ctxData["ally_ids"])
	return engineplayer.AdvancePromptFlowRuntimeChoice(rt, ctxData, substituteDollFlowRuntime, flow, substituteDollStepTarget)
}

func handleBlazeWitchManaInversionXChoice(rt engineplayer.ChoiceRuntime, playerID string, selectionIndex int, ctxData map[string]interface{}) error {
	user := rt.GetPlayers()[playerID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
	xValue := selectionIndex + 2
	if xValue < 2 || xValue > maxX {
		return fmt.Errorf("无效的X值")
	}
	magicIndices := make([]int, 0, len(user.Hand))
	for i, card := range user.Hand {
		if card.Type == model.CardTypeMagic {
			magicIndices = append(magicIndices, i)
		}
	}
	if len(magicIndices) < xValue {
		return fmt.Errorf("法术牌不足，无法弃置X=%d张", xValue)
	}
	flow, err := model.RequirePromptFlow(ctxData, manaInversionFlowID, "魔能反转")
	if err != nil {
		return err
	}
	flow.PutSelection(manaInversionStepX, model.PromptFlowSelection{
		OptionIndexes: []int{selectionIndex},
		Count:         xValue,
	})
	flow.PutSelection(manaInversionStepCards, model.PromptFlowSelection{})
	ctxData["remaining_indices"] = magicIndices
	return engineplayer.AdvancePromptFlowRuntimeChoice(rt, ctxData, manaInversionFlowRuntime, flow, manaInversionStepCards)
}

func handleBlazeWitchManaInversionCardsChoice(rt engineplayer.ChoiceRuntime, playerID string, selectionIndex int, ctxData map[string]interface{}) error {
	user := rt.GetPlayers()[playerID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	remaining := runtimeutil.ParseChoiceIntSlice(ctxData["remaining_indices"])
	flow, err := model.RequirePromptFlow(ctxData, manaInversionFlowID, "魔能反转")
	if err != nil {
		return err
	}
	selected := append([]int{}, flow.Selection(manaInversionStepCards).OptionIndexes...)
	xValue := flow.Selection(manaInversionStepX).Count

	cardIdx, ok := runtimeutil.ResolveSelectionToCandidate(selectionIndex, remaining)
	if !ok || cardIdx < 0 || cardIdx >= len(user.Hand) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	if user.Hand[cardIdx].Type != model.CardTypeMagic {
		return fmt.Errorf("魔能反转需弃置法术牌")
	}
	selected = append(selected, cardIdx)
	nextRemaining := make([]int, 0, len(remaining))
	for _, idx := range remaining {
		if idx != cardIdx {
			nextRemaining = append(nextRemaining, idx)
		}
	}
	if len(selected) < xValue {
		flow.PutSelection(manaInversionStepCards, model.PromptFlowSelection{OptionIndexes: selected})
		ctxData["remaining_indices"] = nextRemaining
		engineplayer.NotifyChoiceContext(rt, ctxData)
		return nil
	}

	enemyIDs := make([]string, 0)
	for _, pID := range rt.GetPlayerOrder() {
		target := rt.GetPlayers()[pID]
		if target == nil || target.Camp == user.Camp {
			continue
		}
		enemyIDs = append(enemyIDs, target.ID)
	}
	if len(enemyIDs) == 0 {
		return fmt.Errorf("无可选敌方目标")
	}
	flow.PutSelection(manaInversionStepCards, model.PromptFlowSelection{
		OptionIndexes: selected,
		Count:         len(selected),
	})
	ctxData["target_ids"] = enemyIDs
	return engineplayer.AdvancePromptFlowRuntimeChoice(rt, ctxData, manaInversionFlowRuntime, flow, manaInversionStepTarget)
}

func handleBlazeWitchTargetChoice(rt engineplayer.ChoiceRuntime, playerID string, selectionIndex int, ctxData map[string]interface{}) error {
	user := rt.GetPlayers()[playerID]
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

	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "bw_substitute_doll_target":
		flow, err := model.RequirePromptFlow(ctxData, substituteDollFlowID, "替身玩偶")
		if err != nil {
			return err
		}
		cardIDs := flow.Selection(substituteDollStepCard).CardIDs
		if len(cardIDs) != 1 || cardIDs[0] == "" {
			return fmt.Errorf("替身玩偶缺少选定卡牌")
		}
		cardID := cardIDs[0]
		cardIdx := -1
		for i, card := range user.Hand {
			if card.ID == cardID {
				cardIdx = i
				break
			}
		}
		if cardIdx < 0 || cardIdx >= len(user.Hand) {
			return fmt.Errorf("无效的弃牌ID")
		}
		if user.Hand[cardIdx].Type != model.CardTypeMagic {
			return fmt.Errorf("替身玩偶需弃置法术牌")
		}
		card := user.Hand[cardIdx]
		user.Hand = append(user.Hand[:cardIdx], user.Hand[cardIdx+1:]...)
		rt.NotifyCardRevealed(user.ID, []model.Card{card}, "discard")
		rt.AppendToDiscard([]model.Card{card})
		rt.DrawCards(targetID, 1)
		rt.Log(fmt.Sprintf("%s 的 [替身玩偶] 生效：%s 摸1张牌", user.Name, target.Name))
	case "bw_mana_inversion_target":
		flow, err := model.RequirePromptFlow(ctxData, manaInversionFlowID, "魔能反转")
		if err != nil {
			return err
		}
		selected := append([]int{}, flow.Selection(manaInversionStepCards).OptionIndexes...)
		xValue := flow.Selection(manaInversionStepX).Count
		if xValue < 2 || len(selected) != xValue {
			return fmt.Errorf("魔能反转弃牌参数错误")
		}
		flow.PutSelection(manaInversionStepTarget, model.PromptFlowSelection{
			OptionIndexes: []int{selectionIndex},
			TargetIDs:     []string{targetID},
		})
		for _, idx := range selected {
			if idx < 0 || idx >= len(user.Hand) || user.Hand[idx].Type != model.CardTypeMagic {
				return fmt.Errorf("魔能反转弃牌必须为法术牌")
			}
		}
		removed, _ := engineplayer.RemoveCardsByIndicesFromHand(user, selected)
		rt.NotifyCardRevealed(user.ID, removed, "discard")
		rt.AppendToDiscard(removed)
		damage := xValue - 1
		if damage > 0 {
			rt.AddPendingDamage(model.PendingDamage{
				SourceID:   user.ID,
				TargetID:   targetID,
				Damage:     damage,
				DamageType: model.MagicAttack,
			})
		}
		rt.Log(fmt.Sprintf("%s 的 [魔能反转] 生效：弃%d张法术牌，对 %s 造成%d点法术伤害", user.Name, xValue, target.Name, damage))
	}

	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil && len(rt.GetPendingDamageQueue()) > 0 {
		rt.EnterDamageResolution(nil)
	}
	return nil
}

func buildPromptOptionsForPlayerIDs(players map[string]*model.Player, ids []string) []model.PromptOption {
	options := make([]model.PromptOption, 0, len(ids))
	for _, id := range ids {
		if p := players[id]; p != nil {
			options = append(options, model.PromptOption{
				ID:       id,
				Label:    p.Name,
				TargetID: id,
			})
		}
	}
	return options
}
