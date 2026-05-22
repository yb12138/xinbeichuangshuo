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
	manaInversionStepCards  = "cards"
	manaInversionStepTarget = "target"

	substituteDollFlowID     = "bw_substitute_doll"
	substituteDollStepCard   = "card"
	substituteDollStepTarget = "target"
)

var manaInversionFlowRuntime = model.MustNewPromptFlowRuntime(manaInversionFlowID, []model.PromptFlowStepSpec{
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
				ID:     fmt.Sprintf("%d", idx),
				Label:  fmt.Sprintf("%d: %s", idx+1, promptfmt.FormatCardInfo(player.Hand[idx])),
				CardID: player.Hand[idx].ID,
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
	case "bw_mana_inversion_cards":
		remaining := runtimeutil.ParseChoiceIntSlice(data["remaining_indices"])
		if _, err := model.RequirePromptFlow(data, manaInversionFlowID, "魔能反转"); err != nil {
			return nil
		}
		options := make([]model.PromptOption, 0, len(remaining))
		for _, idx := range remaining {
			if player == nil || idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{
				ID:     fmt.Sprintf("%d", idx),
				Label:  fmt.Sprintf("%d: %s", idx+1, promptfmt.FormatCardInfo(player.Hand[idx])),
				CardID: player.Hand[idx].ID,
			})
		}
		maxPick := len(options)
		minPick := 2
		if maxPick < minPick {
			minPick = maxPick
		}
		if minPick < 0 {
			minPick = 0
		}
		return &model.Prompt{
			Type:         model.PromptChooseCards,
			PlayerID:     playerID,
			Message:      "【魔能反转】请选择要弃置的法术牌（X=选择数量，至少2张）：",
			Options:      options,
			Min:          minPick,
			Max:          maxPick,
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
	case "bw_blazing_codex_target":
		return engineplayer.BuildTargetChoicePrompt(rt, choiceType, playerID, "【苍炎法典】请选择法术伤害目标：", data, false)
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
	case "bw_mana_inversion_cards":
		return handleBlazeWitchManaInversionCardsMultiSelect(rt, playerID, []int{selectionIndex}, ctxData)
	case "bw_blazing_codex_target", "bw_substitute_doll_target", "bw_mana_inversion_target":
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

func handleBlazeWitchManaInversionCardsMultiSelect(rt engineplayer.ChoiceRuntime, playerID string, selections []int, ctxData map[string]interface{}) (bool, error) {
	user := rt.GetPlayers()[playerID]
	if user == nil {
		return true, fmt.Errorf("玩家不存在")
	}
	flow, err := model.RequirePromptFlow(ctxData, manaInversionFlowID, "魔能反转")
	if err != nil {
		return true, err
	}
	if len(selections) < 2 {
		return true, fmt.Errorf("魔能反转至少需弃置2张法术牌")
	}
	remaining := runtimeutil.ParseChoiceIntSlice(ctxData["remaining_indices"])
	if len(selections) > len(remaining) {
		return true, fmt.Errorf("选择数量超过可弃置的法术牌数量")
	}

	selected := make([]int, 0, len(selections))
	seen := make(map[int]struct{}, len(selections))
	for _, selection := range selections {
		cardIdx, ok := runtimeutil.ResolveSelectionToCandidate(selection, remaining)
		if !ok || cardIdx < 0 || cardIdx >= len(user.Hand) {
			return true, fmt.Errorf("无效的选项索引: %d", selection)
		}
		if _, exists := seen[cardIdx]; exists {
			return true, fmt.Errorf("不能重复选择同一张牌")
		}
		if user.Hand[cardIdx].Type != model.CardTypeMagic {
			return true, fmt.Errorf("魔能反转需弃置法术牌")
		}
		seen[cardIdx] = struct{}{}
		selected = append(selected, cardIdx)
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
		return true, fmt.Errorf("无可选敌方目标")
	}
	flow.PutSelection(manaInversionStepCards, model.PromptFlowSelection{
		OptionIndexes: selected,
		Count:         len(selected),
	})
	ctxData["target_ids"] = enemyIDs
	return true, engineplayer.AdvancePromptFlowRuntimeChoice(rt, ctxData, manaInversionFlowRuntime, flow, manaInversionStepTarget)
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
	case "bw_blazing_codex_target":
		discardIndices := runtimeutil.ParseChoiceIntSlice(ctxData["discard_indices"])
		if len(discardIndices) != 1 {
			return fmt.Errorf("苍炎法典缺少弃牌参数")
		}
		rt.PopInterrupt()
		if rt.GetPendingInterrupt() != nil {
			return fmt.Errorf("当前仍有其他待处理的中断")
		}
		rt.ApplyChoiceResumePoint(ctxData["resume_phase"])
		skillRuntime, ok := rt.(interface {
			UseSkill(playerID, skillID string, targetIDs []string, discardIndices []int) error
		})
		if !ok {
			return fmt.Errorf("技能运行时不支持发动苍炎法典")
		}
		return skillRuntime.UseSkill(playerID, "bw_blazing_codex", []string{targetID}, discardIndices)
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
		xValue := flow.Selection(manaInversionStepCards).Count
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
