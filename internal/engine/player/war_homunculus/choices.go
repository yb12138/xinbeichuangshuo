// gameflow: 英灵人形角色选择流。

package war_homunculus

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

// ---------------------------------------------------------------------------
// BuildPrompt
// ---------------------------------------------------------------------------

func (choiceHandler) BuildPrompt(rt engineplayer.ChoiceRuntime, choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "hom_rune_reforge_distribution":
		return buildRuneReforgeDistributionPrompt(playerID, data)
	case "hom_rune_smash_x", "hom_glyph_fusion_x":
		return buildRuneXPrompt(playerID, data, choiceType == "hom_glyph_fusion_x")
	case "hom_rune_smash_cards", "hom_glyph_fusion_cards":
		return buildRuneCardsPrompt(playerID, player, data, choiceType == "hom_glyph_fusion_cards")
	case "hom_rune_smash_y", "hom_glyph_fusion_y":
		return buildRuneYPrompt(playerID, data, choiceType == "hom_glyph_fusion_y")
	case "hom_dual_echo_target":
		return buildDualEchoTargetPrompt(rt, playerID, data)
	default:
		return nil
	}
}

func buildRuneReforgeDistributionPrompt(playerID string, data map[string]interface{}) *model.Prompt {
	total := runtimeutil.ToIntContextValue(data["total_runes"])
	if total <= 0 {
		total = 3
	}
	options := make([]model.PromptOption, 0, total+1)
	for warRunes := 0; warRunes <= total; warRunes++ {
		magicRunes := total - warRunes
		options = append(options, model.PromptOption{
			ID:    fmt.Sprintf("%d", warRunes),
			Label: fmt.Sprintf("战纹 %d / 魔纹 %d", warRunes, magicRunes),
		})
	}
	return &model.Prompt{
		Type:     model.PromptConfirm,
		PlayerID: playerID,
		Message:  fmt.Sprintf("【符文改造】请选择战纹/魔纹分配（总计%d）：", total),
		Options:  options,
		Min:      1,
		Max:      1,
	}
}

func buildRuneXPrompt(playerID string, data map[string]interface{}, glyph bool) *model.Prompt {
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
}

func buildRuneCardsPrompt(playerID string, player *model.Player, data map[string]interface{}, glyph bool) *model.Prompt {
	remaining := ParseIntSliceContextValue(data["remaining_indices"])
	xValue := runtimeutil.ToIntContextValue(data["x_value"])
	selectedCount := len(ParseIntSliceContextValue(data["selected_indices"]))
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
}

func buildRuneYPrompt(playerID string, data map[string]interface{}, glyph bool) *model.Prompt {
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
}

func buildDualEchoTargetPrompt(rt engineplayer.ChoiceRuntime, playerID string, data map[string]interface{}) *model.Prompt {
	targetIDs := runtimeutil.ParseStringSliceContextValue(data["target_ids"])
	options := make([]model.PromptOption, 0, len(targetIDs)+1)
	for _, targetID := range targetIDs {
		if target := rt.LookupPlayer(targetID); target != nil {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: target.Name})
		}
	}
	options = append(options, model.PromptOption{ID: "cancel", Label: "取消"})
	damage := runtimeutil.ToIntContextValue(data["damage"])
	return &model.Prompt{
		Type:     model.PromptConfirm,
		PlayerID: playerID,
		Message:  fmt.Sprintf("【双重回响】请选择额外造成%d点法术伤害的目标：", damage),
		Options:  options,
		Min:      1,
		Max:      1,
	}
}

// ---------------------------------------------------------------------------
// HandleChoice
// ---------------------------------------------------------------------------

func (choiceHandler) HandleChoice(rt engineplayer.ChoiceRuntime, _ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)

	switch choiceType {
	case "hom_rune_reforge_distribution":
		return true, handleRuneReforgeDistribution(rt, ctxData, selectionIndex)
	case "hom_rune_smash_x", "hom_glyph_fusion_x":
		return true, handleRuneX(rt, ctxData, selectionIndex, choiceType == "hom_glyph_fusion_x")
	case "hom_rune_smash_cards", "hom_glyph_fusion_cards":
		return true, handleRuneCards(rt, ctxData, selectionIndex, choiceType == "hom_glyph_fusion_cards")
	case "hom_rune_smash_y", "hom_glyph_fusion_y":
		return true, handleRuneY(rt, ctxData, selectionIndex, choiceType == "hom_glyph_fusion_y")
	case "hom_dual_echo_target":
		return true, handleDualEchoTarget(rt, ctxData, selectionIndex)
	default:
		return false, nil
	}
}

func (choiceHandler) HandleCancel(rt engineplayer.ChoiceRuntime, _ string, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "hom_dual_echo_target":
		// Cancel dual echo: clear interrupt without consuming cost or adding damage
		rt.PopInterrupt()
		return true, nil
	default:
		return false, nil
	}
}

// ---------------------------------------------------------------------------
// Individual choice-type handlers
// ---------------------------------------------------------------------------

func handleRuneReforgeDistribution(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.LookupPlayer(userID)
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	total := runtimeutil.ToIntContextValue(ctxData["total_runes"])
	if total <= 0 {
		total = 3
	}
	warRunes := selectionIndex
	if warRunes < 0 || warRunes > total {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	if user.Tokens == nil {
		user.Tokens = map[string]int{}
	}
	user.Tokens["hom_war_rune"] = warRunes
	user.Tokens["hom_magic_rune"] = total - warRunes
	rt.Log(fmt.Sprintf("%s 的 [符文改造]：战纹=%d，魔纹=%d", user.Name, user.Tokens["hom_war_rune"], user.Tokens["hom_magic_rune"]))
	rt.PopInterrupt()
	if !rt.HasPendingInterrupt() {
		rt.ApplyChoiceResumePoint(model.TurnStageActionStart)
	}
	return nil
}

func handleRuneX(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int, glyph bool) error {
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

	candidates := ParseIntSliceContextValue(ctxData["candidate_indices"])
	if xValue > len(candidates) {
		return fmt.Errorf("可选牌数量不足")
	}

	ctxData["choice_type"] = nextChoice
	ctxData["x_value"] = xValue
	ctxData["selected_indices"] = []int{}
	ctxData["remaining_indices"] = append([]int{}, candidates...)
	updateRuneChoiceContext(rt, ctxData)
	return nil
}

func handleRuneCards(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int, glyph bool) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.LookupPlayer(userID)
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}

	remaining := ParseIntSliceContextValue(ctxData["remaining_indices"])
	selected := ParseIntSliceContextValue(ctxData["selected_indices"])
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
	if err := validateRuneCardSelection(user, nextSelected, attackElement, glyph, mismatchErr, duplicateErr); err != nil {
		return err
	}

	nextRemaining := filterRuneRemainingCandidates(user, remaining, cardIdx, glyph)
	if len(nextSelected) < xValue {
		ctxData["selected_indices"] = nextSelected
		ctxData["remaining_indices"] = nextRemaining
		updateRuneChoiceContext(rt, ctxData)
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
		updateRuneChoiceContext(rt, ctxData)
		return nil
	}

	ctxData["y_value"] = 0
	return resolveRuneChoice(rt, ctxData, glyph)
}

func handleRuneY(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int, glyph bool) error {
	maxY := runtimeutil.ToIntContextValue(ctxData["max_y"])
	yValue := selectionIndex
	if yValue < 0 || yValue > maxY {
		return fmt.Errorf("无效的Y值")
	}
	ctxData["y_value"] = yValue
	return resolveRuneChoice(rt, ctxData, glyph)
}

func handleDualEchoTarget(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.LookupPlayer(userID)
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
	if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	costPending := runtimeutil.ToIntContextValue(ctxData["cost_pending"])
	if costPending > 0 {
		if !rt.ConsumeCrystalCost(user.ID, costPending) {
			return fmt.Errorf("双重回响需要1蓝水晶（红宝石可替代）")
		}
		ctxData["cost_pending"] = 0
	}
	damage := runtimeutil.ToIntContextValue(ctxData["damage"])
	if damage < 0 {
		damage = 0
	}
	targetID := targetIDs[selectionIndex]
	rt.AddPendingDamage(model.PendingDamage{SourceID: user.ID, TargetID: targetID, Damage: damage, DamageType: "magic_no_morale"})
	if target := rt.LookupPlayer(targetID); target != nil {
		rt.Log(fmt.Sprintf("%s 的 [双重回响] 对 %s 造成%d点法术伤害", user.Name, target.Name, damage))
	}
	rt.PopInterrupt()
	if !rt.HasPendingInterrupt() && rt.PendingDamageQueueLen() > 0 {
		rt.EnterDamageResolution(nil)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Rune choice resolution
// ---------------------------------------------------------------------------

// resolveRuneChoice resolves the final outcome of 战纹碎击/魔纹融合.
func resolveRuneChoice(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, glyph bool) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.LookupPlayer(userID)
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
	if err := validateRuneCardSelection(user, selected, attackElement, glyph, mismatchErr, duplicateErr); err != nil {
		return err
	}

	flipCount := 1 + yVal
	if err := applyRuneFlip(user, glyph, flipCount); err != nil {
		return err
	}

	removed := removeCardsByIndicesFromHand(user, append([]int{}, selected...))
	rt.NotifyCardRevealed(user.ID, removed, "discard")
	rt.AppendToDiscard(removed)

	targetID := rawCtx.EventCtx.TargetID
	if glyph {
		damage := xVal - 1 + yVal
		if damage < 0 {
			damage = 0
		}
		if damage > 0 && targetID != "" {
			rt.AddPendingDamage(model.PendingDamage{
				SourceID:   user.ID,
				TargetID:   targetID,
				Damage:     damage,
				DamageType: model.MagicAttack,
			})
		}
		rt.Log(fmt.Sprintf("%s 发动 [魔纹融合]：弃%d张异系牌，翻转%d个魔纹为战纹，额外造成%d点法术伤害", user.Name, xVal, flipCount, damage))
		rt.PopInterrupt()
		if !rt.HasPendingInterrupt() && rawCtx.ResumeAttackMissPhase() {
			if rt.ResumePendingAttackMiss(rawCtx) {
				return nil
			}
		}
		if !rt.HasPendingInterrupt() && rt.PendingDamageQueueLen() > 0 {
			rt.EnterDamageResolution(nil)
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
		rt.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   targetID,
			Damage:     yVal,
			DamageType: model.MagicAttack,
		})
	}
	rt.Log(fmt.Sprintf("%s 发动 [战纹碎击]：弃%d张同系牌，翻转%d个战纹为魔纹，本次攻击伤害+%d", user.Name, xVal, flipCount, bonusDamage))
	rt.PopInterrupt()
	rt.ResumePendingAttackHit(ctxData)
	return nil
}

// ---------------------------------------------------------------------------
// Interrupt context helpers
// ---------------------------------------------------------------------------

// updateRuneChoiceContext replaces the current pending interrupt context and
// re-notifies the prompt.
func updateRuneChoiceContext(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}) {
	_ = rt.ReplacePendingInterruptContext(ctxData)
	rt.NotifyInterruptPrompt()
}

// ---------------------------------------------------------------------------
// Rune validation & mutation helpers
// ---------------------------------------------------------------------------

// validateRuneCardSelection checks that the selected hand indices satisfy the
// element-matching rules for 战纹碎击 (same element) or 魔纹融合 (different
// element from attack element, no duplicates).
func validateRuneCardSelection(user *model.Player, selected []int, attackElement string, glyph bool, mismatchErr, duplicateErr string) error {
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

// applyRuneFlip converts rune tokens between war and magic types.
func applyRuneFlip(user *model.Player, glyph bool, flipCount int) error {
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

// filterRuneRemainingCandidates filters out the picked card index from the
// remaining candidate list. For glyph mode, it also removes cards with the
// same element as the picked card.
func filterRuneRemainingCandidates(user *model.Player, remaining []int, picked int, glyph bool) []int {
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

// ---------------------------------------------------------------------------
// Card helpers
// ---------------------------------------------------------------------------

// removeCardsByIndicesFromHand removes the cards at the given indices from the
// player's hand and returns the removed cards.
func removeCardsByIndicesFromHand(player *model.Player, indices []int) []model.Card {
	if player == nil || len(indices) == 0 {
		return nil
	}
	removed := make([]model.Card, 0, len(indices))
	for _, idx := range indices {
		if idx >= 0 && idx < len(player.Hand) {
			removed = append(removed, player.Hand[idx])
		}
	}
	newHand := make([]model.Card, 0, len(player.Hand)-len(indices))
	for i, card := range player.Hand {
		keep := true
		for _, idx := range indices {
			if i == idx {
				keep = false
				break
			}
		}
		if keep {
			newHand = append(newHand, card)
		}
	}
	player.Hand = newHand
	return removed
}

// ---------------------------------------------------------------------------
// Context value helpers
// ---------------------------------------------------------------------------

// parseIntSliceContextValue extracts a []int from an interface{}, handling both
// []int and []interface{} slices.
func ParseIntSliceContextValue(raw interface{}) []int {
	var out []int
	switch arr := raw.(type) {
	case []int:
		out = append(out, arr...)
	case []interface{}:
		for _, item := range arr {
			if f, ok := item.(float64); ok {
				out = append(out, int(f))
			}
		}
	}
	return out
}
