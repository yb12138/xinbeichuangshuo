// gameflow: 英灵人形相关技能流。

package engine

import (
	"fmt"
	"starcup-engine/internal/engine/core/runtimeutil"

	"starcup-engine/internal/model"
)

func validateHomRuneCardSelection(user *model.Player, selected []int, attackElement string, glyph bool, mismatchErr, duplicateErr string) error {
	seen := map[model.Element]bool{}
	for _, idx := range selected {
		if idx < 0 || idx >= len(user.Hand) {
			return fmt.Errorf("无效的手牌索引: %d", idx)
		}
		elem := user.Hand[idx].Element
		if glyph {
			if attackElement != "" && string(elem) == attackElement {
				return fmt.Errorf(mismatchErr)
			}
			if duplicateErr != "" && seen[elem] {
				return fmt.Errorf(duplicateErr)
			}
			seen[elem] = true
			continue
		}
		if attackElement != "" && string(elem) != attackElement {
			return fmt.Errorf(mismatchErr)
		}
	}
	return nil
}

func applyHomRuneFlip(user *model.Player, glyph bool, flipCount int) error {
	if user.Tokens == nil {
		user.Tokens = map[string]int{}
	}
	if glyph {
		if user.Tokens["hom_magic_rune"] < flipCount {
			return fmt.Errorf("魔纹不足，至少需要%d个", flipCount)
		}
		user.Tokens["hom_magic_rune"] -= flipCount
		user.Tokens["hom_war_rune"] += flipCount
		return nil
	}
	if user.Tokens["hom_war_rune"] < flipCount {
		return fmt.Errorf("战纹不足，至少需要%d个", flipCount)
	}
	user.Tokens["hom_war_rune"] -= flipCount
	user.Tokens["hom_magic_rune"] += flipCount
	return nil
}

func filterHomRuneRemainingCandidates(user *model.Player, remaining []int, picked int, glyph bool) []int {
	nextRemaining := make([]int, 0, len(remaining))
	for _, idx := range remaining {
		if idx == picked {
			continue
		}
		if glyph && idx >= 0 && idx < len(user.Hand) && picked >= 0 && picked < len(user.Hand) && user.Hand[idx].Element == user.Hand[picked].Element {
			continue
		}
		nextRemaining = append(nextRemaining, idx)
	}
	return nextRemaining
}

func (e *GameEngine) updateHomRuneChoiceContext(ctxData map[string]interface{}) {
	e.State.PendingInterrupt.Context = ctxData
	e.notifyInterruptPrompt()
}

// resolveHomunculusRuneChoice 结算英灵人形“战纹碎击/魔纹融合”的X/Y交互结果。
func (e *GameEngine) resolveHomunculusRuneChoice(ctxData map[string]interface{}, glyph bool) error {
	userID, _ := ctxData["user_id"].(string)
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	rawCtx, _ := ctxData["user_ctx"].(*model.Context)
	if rawCtx == nil || rawCtx.EventCtx == nil {
		return fmt.Errorf("英灵人形技能上下文丢失")
	}
	xVal := runtimeutil.ToIntContextValue(ctxData["x_value"])
	yVal := runtimeutil.ToIntContextValue(ctxData["y_value"])
	if xVal <= 0 || yVal < 0 {
		return fmt.Errorf("X/Y 参数无效")
	}
	selected := runtimeutil.ParseChoiceIntSlice(ctxData["selected_indices"])
	if len(selected) != xVal {
		return fmt.Errorf("弃牌数量与X不一致")
	}

	attackElement, _ := ctxData["attack_element"].(string)
	mismatchErr := "战纹碎击需弃置同系牌"
	duplicateErr := ""
	if glyph {
		mismatchErr = "魔纹融合需弃置异系牌"
		duplicateErr = "魔纹融合需弃置元素互不相同的异系牌"
	}
	if err := validateHomRuneCardSelection(user, selected, attackElement, glyph, mismatchErr, duplicateErr); err != nil {
		return err
	}

	flipCount := 1 + yVal
	if err := applyHomRuneFlip(user, glyph, flipCount); err != nil {
		return err
	}

	removed, err := removeCardsByIndicesFromHand(user, append([]int{}, selected...))
	if err != nil {
		return err
	}
	e.NotifyCardRevealed(user.ID, removed, "discard")
	e.State.DiscardPile = append(e.State.DiscardPile, removed...)

	targetID := rawCtx.EventCtx.TargetID
	if glyph {
		damage := xVal - 1 + yVal
		if damage < 0 {
			damage = 0
		}
		if damage > 0 && targetID != "" {
			e.AddPendingDamage(model.PendingDamage{
				SourceID:   user.ID,
				TargetID:   targetID,
				Damage:     damage,
				DamageType: model.MagicAttack,
			})
		}
		e.Log(fmt.Sprintf("%s 发动 [魔纹融合]：弃%d张异系牌，翻转%d个魔纹为战纹，额外造成%d点法术伤害", user.Name, xVal, flipCount, damage))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil && rawCtx.ResumeAttackMissPhase() {
			if e.resumePendingAttackMiss(rawCtx) {
				return nil
			}
		}
		if e.State.PendingInterrupt == nil && len(e.State.PendingDamageQueue) > 0 {
			e.enterDamageResolution(nil)
		}
		return nil
	}

	bonusDamage := xVal - 1
	if bonusDamage < 0 {
		bonusDamage = 0
	}
	if rawCtx.EventCtx.DamageVal != nil && bonusDamage > 0 {
		*rawCtx.EventCtx.DamageVal += bonusDamage
	}
	if yVal > 0 && targetID != "" {
		e.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   targetID,
			Damage:     yVal,
			DamageType: model.MagicAttack,
		})
	}
	e.Log(fmt.Sprintf("%s 发动 [战纹碎击]：弃%d张同系牌，翻转%d个战纹为魔纹，本次攻击伤害+%d", user.Name, xVal, flipCount, bonusDamage))
	e.PopInterrupt()
	e.resumePendingAttackHit(ctxData)
	return nil
}

func (e *GameEngine) buildWarHomunculusChoicePrompt(choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "hom_rune_reforge_distribution":
		total := runtimeutil.ToIntContextValue(data["total_runes"])
		if total <= 0 {
			total = 3
		}
		options := make([]model.PromptOption, 0, total+1)
		for warRunes := 0; warRunes <= total; warRunes++ {
			magicRunes := total - warRunes
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", warRunes), Label: fmt.Sprintf("战纹 %d / 魔纹 %d", warRunes, magicRunes)})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: fmt.Sprintf("【符文改造】请选择战纹/魔纹分配（总计%d）：", total), Options: options, Min: 1, Max: 1}

	case "hom_rune_smash_x", "hom_glyph_fusion_x":
		glyph := choiceType == "hom_glyph_fusion_x"
		maxX := runtimeutil.ToIntContextValue(data["max_x"])
		minX := 1
		message := "【战纹碎击】请选择X（弃置同系牌数量）："
		if glyph {
			minX = 2
			message = "【魔纹融合】请选择X（弃置异系牌数量）："
		}
		if maxX < minX {
			return nil
		}
		options := make([]model.PromptOption, 0, maxX-minX+1)
		for xValue := minX; xValue <= maxX; xValue++ {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", xValue), Label: fmt.Sprintf("X=%d", xValue)})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: message, Options: options, Min: 1, Max: 1}

	case "hom_rune_smash_cards", "hom_glyph_fusion_cards":
		glyph := choiceType == "hom_glyph_fusion_cards"
		remaining := parseIntSliceContextValue(data["remaining_indices"])
		xValue := runtimeutil.ToIntContextValue(data["x_value"])
		selectedCount := len(parseIntSliceContextValue(data["selected_indices"]))
		options := make([]model.PromptOption, 0, len(remaining))
		for _, idx := range remaining {
			if player == nil || idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", idx), Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(player.Hand[idx]))})
		}
		remainingPick := xValue - selectedCount
		if remainingPick < 1 {
			remainingPick = 1
		}
		if len(options) > 0 && remainingPick > len(options) {
			remainingPick = len(options)
		}
		message := fmt.Sprintf("【战纹碎击】请选择要弃置的%d张牌：", remainingPick)
		if glyph {
			message = fmt.Sprintf("【魔纹融合】请选择要弃置的%d张牌（元素不可重复）：", remainingPick)
		}
		return &model.Prompt{Type: model.PromptChooseCards, PlayerID: playerID, Message: message, Options: options, Min: remainingPick, Max: remainingPick}

	case "hom_rune_smash_y", "hom_glyph_fusion_y":
		glyph := choiceType == "hom_glyph_fusion_y"
		maxY := runtimeutil.ToIntContextValue(data["max_y"])
		options := make([]model.PromptOption, 0, maxY+1)
		for yValue := 0; yValue <= maxY; yValue++ {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", yValue), Label: fmt.Sprintf("Y=%d", yValue)})
		}
		message := "【战纹碎击】请选择Y（额外翻转战纹数）："
		if glyph {
			message = "【魔纹融合】请选择Y（额外翻转魔纹数）："
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: message, Options: options, Min: 1, Max: 1}

	case "hom_dual_echo_target":
		targetIDs := runtimeutil.ParseStringSliceContextValue(data["target_ids"])
		options := make([]model.PromptOption, 0, len(targetIDs)+1)
		for _, targetID := range targetIDs {
			if target := e.State.Players[targetID]; target != nil {
				options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: target.Name})
			}
		}
		options = append(options, model.PromptOption{ID: "cancel", Label: "取消"})
		damage := runtimeutil.ToIntContextValue(data["damage"])
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: fmt.Sprintf("【双重回响】请选择额外造成%d点法术伤害的目标：", damage), Options: options, Min: 1, Max: 1}
	}

	return nil
}

func (e *GameEngine) handleHomRuneXChoice(selectionIndex int, ctxData map[string]interface{}, glyph bool) error {
	maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
	minX := 1
	nextChoice := "hom_rune_smash_cards"
	if glyph {
		minX = 2
		nextChoice = "hom_glyph_fusion_cards"
	}

	xValue := selectionIndex
	if xValue < minX || xValue > maxX {
		xValue = selectionIndex + minX
	}
	if xValue < minX || xValue > maxX {
		return fmt.Errorf("无效的X值")
	}

	candidates := parseIntSliceContextValue(ctxData["candidate_indices"])
	if xValue > len(candidates) {
		return fmt.Errorf("可选牌数量不足")
	}

	ctxData["choice_type"] = nextChoice
	ctxData["x_value"] = xValue
	ctxData["selected_indices"] = []int{}
	ctxData["remaining_indices"] = append([]int{}, candidates...)
	e.updateHomRuneChoiceContext(ctxData)
	return nil
}

func (e *GameEngine) handleHomRuneCardsChoice(selectionIndex int, ctxData map[string]interface{}, glyph bool) error {
	userID, _ := ctxData["user_id"].(string)
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}

	remaining := parseIntSliceContextValue(ctxData["remaining_indices"])
	selected := parseIntSliceContextValue(ctxData["selected_indices"])
	xValue := runtimeutil.ToIntContextValue(ctxData["x_value"])
	cardIdx, ok := runtimeutil.ResolveSelectionToCandidate(selectionIndex, remaining)
	if !ok {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	if cardIdx < 0 || cardIdx >= len(user.Hand) {
		return fmt.Errorf("无效的手牌索引: %d", cardIdx)
	}

	attackElement, _ := ctxData["attack_element"].(string)
	nextSelected := append(append([]int{}, selected...), cardIdx)
	mismatchErr := "战纹碎击需选择与攻击同系的牌"
	duplicateErr := ""
	if glyph {
		mismatchErr = "魔纹融合需选择与攻击异系的牌"
		duplicateErr = "魔纹融合需选择元素互不相同的异系牌"
	}
	if err := validateHomRuneCardSelection(user, nextSelected, attackElement, glyph, mismatchErr, duplicateErr); err != nil {
		return err
	}

	nextRemaining := filterHomRuneRemainingCandidates(user, remaining, cardIdx, glyph)
	if len(nextSelected) < xValue {
		ctxData["selected_indices"] = nextSelected
		ctxData["remaining_indices"] = nextRemaining
		e.updateHomRuneChoiceContext(ctxData)
		return nil
	}

	ctxData["selected_indices"] = nextSelected
	maxY := runtimeutil.ToIntContextValue(ctxData["max_y"])
	if maxY > 0 {
		if glyph {
			ctxData["choice_type"] = "hom_glyph_fusion_y"
		} else {
			ctxData["choice_type"] = "hom_rune_smash_y"
		}
		e.updateHomRuneChoiceContext(ctxData)
		return nil
	}

	ctxData["y_value"] = 0
	return e.resolveHomunculusRuneChoice(ctxData, glyph)
}

func (e *GameEngine) handleWarHomunculusChoiceInput(_ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)

	switch choiceType {
	case "hom_rune_reforge_distribution":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		total := runtimeutil.ToIntContextValue(ctxData["total_runes"])
		if total <= 0 {
			total = 3
		}
		warRunes := selectionIndex
		if warRunes < 0 || warRunes > total {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		if user.Tokens == nil {
			user.Tokens = map[string]int{}
		}
		user.Tokens["hom_war_rune"] = warRunes
		user.Tokens["hom_magic_rune"] = total - warRunes
		e.Log(fmt.Sprintf("%s 的 [符文改造]：战纹=%d，魔纹=%d", user.Name, user.Tokens["hom_war_rune"], user.Tokens["hom_magic_rune"]))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.setTurnStage(model.TurnStageActionStart)
			e.clearCombatStage()
			e.clearSubflow()
		}
		return true, nil

	case "hom_rune_smash_x", "hom_glyph_fusion_x":
		return true, e.handleHomRuneXChoice(selectionIndex, ctxData, choiceType == "hom_glyph_fusion_x")

	case "hom_rune_smash_cards", "hom_glyph_fusion_cards":
		return true, e.handleHomRuneCardsChoice(selectionIndex, ctxData, choiceType == "hom_glyph_fusion_cards")

	case "hom_rune_smash_y", "hom_glyph_fusion_y":
		maxY := runtimeutil.ToIntContextValue(ctxData["max_y"])
		yValue := selectionIndex
		if yValue < 0 || yValue > maxY {
			return true, fmt.Errorf("无效的Y值")
		}
		ctxData["y_value"] = yValue
		return true, e.resolveHomunculusRuneChoice(ctxData, choiceType == "hom_glyph_fusion_y")

	case "hom_dual_echo_target":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		costPending := runtimeutil.ToIntContextValue(ctxData["cost_pending"])
		if costPending > 0 {
			if !e.ConsumeCrystalCost(user.ID, costPending) {
				return true, fmt.Errorf("双重回响需要1蓝水晶（红宝石可替代）")
			}
			ctxData["cost_pending"] = 0
		}
		damage := runtimeutil.ToIntContextValue(ctxData["damage"])
		if damage < 0 {
			damage = 0
		}
		targetID := targetIDs[selectionIndex]
		e.AddPendingDamage(model.PendingDamage{SourceID: user.ID, TargetID: targetID, Damage: damage, DamageType: "magic_no_morale"})
		if target := e.State.Players[targetID]; target != nil {
			e.Log(fmt.Sprintf("%s 的 [双重回响] 对 %s 造成%d点法术伤害", user.Name, target.Name, damage))
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil && len(e.State.PendingDamageQueue) > 0 {
			e.enterDamageResolution(nil)
		}
		return true, nil
	}

	return false, nil
}
