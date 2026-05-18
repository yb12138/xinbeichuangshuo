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

const (
	runeSmashFlowID     = "hom_rune_smash"
	glyphFusionFlowID   = "hom_glyph_fusion"
	runeChoiceStepX     = "x"
	runeChoiceStepCards = "cards"
	runeChoiceStepY     = "y"
)

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
	// 使用类似圣疗的两行数字分配面板
	// 前端检测 choice_type="hom_rune_reforge_allocate" 后渲染战纹/魔纹两行数字选择器
	return &model.Prompt{
		Type:       model.PromptConfirm,
		ChoiceType: "hom_rune_reforge_allocate",
		PlayerID:   playerID,
		Message:    fmt.Sprintf("【符文改造】请分配战纹/魔纹（总计%d）：", total),
		Options: []model.PromptOption{
			{ID: "hom_war_rune", Label: "战纹", Hint: fmt.Sprintf("max:%d", total)},
			{ID: "hom_magic_rune", Label: "魔纹", Hint: fmt.Sprintf("max:%d", total)},
		},
		Min:          2,
		Max:          2,
		Presentation: &model.PromptPresentation{Kind: model.PresentationNumeric, Layout: "rune_allocate", NumericBase: 0},
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
	return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: message, Options: options, Min: 1, Max: 1, Presentation: &model.PromptPresentation{Kind: model.PresentationNumeric, NumericBase: 0}}
}

func buildRuneCardsPrompt(playerID string, player *model.Player, data map[string]interface{}, glyph bool) *model.Prompt {
	candidates := engineplayer.ParseIntSliceContextValue(data["candidate_indices"])
	remaining := engineplayer.ParseIntSliceContextValue(data["remaining_indices"])
	if len(remaining) == 0 {
		remaining = candidates
	}
	flow, err := model.RequirePromptFlow(data, runeChoiceFlowID(glyph), runeChoiceLabel(glyph))
	if err != nil {
		return nil
	}
	xValue := flow.Selection(runeChoiceStepX).Count
	minPick := runtimeutil.ToIntContextValue(data["min_pick"])
	if minPick < 1 {
		if glyph {
			minPick = 2 // 魔纹融合至少2张
		} else {
			minPick = 1 // 战纹碎击至少1张
		}
	}
	selectedCount := len(flow.Selection(runeChoiceStepCards).OptionIndexes)
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
	// 计算还需要选择多少张牌
	remainingPick := minPick
	if xValue > 0 {
		remainingPick = xValue - selectedCount
	}
	if remainingPick < 1 {
		remainingPick = 1
	}
	maxPick := len(options)
	if len(options) > 0 && remainingPick > len(options) {
		remainingPick = len(options)
	}
	message := fmt.Sprintf("【战纹碎击】请选择要弃置的同系牌（所选牌彼此同系，至少%d张）：", minPick)
	if glyph {
		message = fmt.Sprintf("【魔纹融合】请选择要弃置的异系牌（所选牌彼此异系，至少%d张）：", minPick)
	}
	return &model.Prompt{Type: model.PromptChooseCards, PlayerID: playerID, Message: message, Options: options, Min: remainingPick, Max: maxPick, Presentation: &model.PromptPresentation{Kind: model.PresentationCardPicker, CardSource: "hand"}}
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
	return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: message, Options: options, Min: 1, Max: 1, Presentation: &model.PromptPresentation{Kind: model.PresentationNumeric, NumericBase: 0}}
}

func buildDualEchoTargetPrompt(rt engineplayer.ChoiceRuntime, playerID string, data map[string]interface{}) *model.Prompt {
	targetIDs := runtimeutil.ParseStringSliceContextValue(data["target_ids"])
	options := make([]model.PromptOption, 0, len(targetIDs)+1)
	for _, targetID := range targetIDs {
		if target := rt.GetPlayers()[targetID]; target != nil {
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
	case "hom_rune_reforge_distribution", "hom_rune_reforge_allocate":
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
	user := rt.GetPlayers()[userID]
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
	if rt.GetPendingInterrupt() == nil {
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

	candidates := engineplayer.ParseIntSliceContextValue(ctxData["candidate_indices"])
	if xValue > len(candidates) {
		return fmt.Errorf("可选牌数量不足")
	}

	flow, err := model.RequirePromptFlow(ctxData, runeChoiceFlowID(glyph), runeChoiceLabel(glyph))
	if err != nil {
		return err
	}
	flow.PutSelection(runeChoiceStepX, model.PromptFlowSelection{
		OptionIndexes: []int{selectionIndex},
		Count:         xValue,
	})
	flow.PutSelection(runeChoiceStepCards, model.PromptFlowSelection{Count: xValue})
	ctxData["remaining_indices"] = append([]int{}, candidates...)
	engineplayer.AdvancePromptFlowChoice(rt, ctxData, flow, runeChoiceStepCards, nextChoice)
	return nil
}

func handleRuneCards(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int, glyph bool) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}

	remaining := engineplayer.ParseIntSliceContextValue(ctxData["remaining_indices"])
	flow, err := model.RequirePromptFlow(ctxData, runeChoiceFlowID(glyph), runeChoiceLabel(glyph))
	if err != nil {
		return err
	}
	selected := append([]int{}, flow.Selection(runeChoiceStepCards).OptionIndexes...)
	xValue := flow.Selection(runeChoiceStepX).Count
	cardIdx, ok := runtimeutil.ResolveSelectionToCandidate(selectionIndex, remaining)
	if !ok {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	if cardIdx < 0 || cardIdx >= len(user.Hand) {
		return fmt.Errorf("无效的手牌索引: %d", cardIdx)
	}

	attackElement, _ := ctxData["attack_element"].(string)
	nextSelected := append(append([]int{}, selected...), cardIdx)
	mismatchErr := "战纹碎击需选择彼此同系的牌"
	duplicateErr := ""
	if glyph {
		mismatchErr = "魔纹融合需选择彼此异系的牌"
		duplicateErr = "魔纹融合需选择元素互不相同的异系牌"
	}
	if err := validateRuneCardSelection(user, nextSelected, attackElement, glyph, mismatchErr, duplicateErr); err != nil {
		return err
	}

	nextRemaining := filterRuneRemainingCandidates(user, remaining, cardIdx, glyph)
	// 如果还在直接选牌模式（xValue == 0），检查是否需要继续选牌
	if xValue == 0 {
		minPick := runtimeutil.ToIntContextValue(ctxData["min_pick"])
		if minPick < 1 {
			if glyph {
				minPick = 2
			} else {
				minPick = 1
			}
		}
		// 选择了足够数量的牌，进入 Y 选择或结算
		if len(nextSelected) >= minPick {
			flow.PutSelection(runeChoiceStepCards, model.PromptFlowSelection{
				OptionIndexes: nextSelected,
				Count:         len(nextSelected),
			})
			maxY := runtimeutil.ToIntContextValue(ctxData["max_y"])
			if maxY > 0 {
				nextStep := "hom_rune_smash_y"
				if glyph {
					nextStep = "hom_glyph_fusion_y"
				}
				engineplayer.AdvancePromptFlowChoice(rt, ctxData, flow, runeChoiceStepY, nextStep)
				return nil
			}
			return resolveRuneChoice(rt, ctxData, glyph)
		}
		// 还需要继续选牌
		flow.PutSelection(runeChoiceStepCards, model.PromptFlowSelection{OptionIndexes: nextSelected})
		ctxData["remaining_indices"] = nextRemaining
		engineplayer.NotifyChoiceContext(rt, ctxData)
		return nil
	}

	// 旧流程：逐张选牌直到达到 xValue
	if len(nextSelected) < xValue {
		flow.PutSelection(runeChoiceStepCards, model.PromptFlowSelection{
			OptionIndexes: nextSelected,
			Count:         xValue,
		})
		ctxData["remaining_indices"] = nextRemaining
		engineplayer.NotifyChoiceContext(rt, ctxData)
		return nil
	}

	flow.PutSelection(runeChoiceStepCards, model.PromptFlowSelection{
		OptionIndexes: nextSelected,
		Count:         xValue,
	})
	maxY := runtimeutil.ToIntContextValue(ctxData["max_y"])
	if maxY > 0 {
		nextStep := "hom_rune_smash_y"
		if glyph {
			nextStep = "hom_glyph_fusion_y"
		}
		engineplayer.AdvancePromptFlowChoice(rt, ctxData, flow, runeChoiceStepY, nextStep)
		return nil
	}

	return resolveRuneChoice(rt, ctxData, glyph)
}

// handleRuneCardsMultiSelect 返回处理多选的函数。
func handleRuneCardsMultiSelect(glyph bool) func(rt engineplayer.ChoiceRuntime, playerID string, selections []int, ctxData map[string]interface{}) (bool, error) {
	return func(rt engineplayer.ChoiceRuntime, playerID string, selections []int, ctxData map[string]interface{}) (bool, error) {
		userID, _ := ctxData["user_id"].(string)
		user := rt.GetPlayers()[userID]
		if user == nil {
			return false, fmt.Errorf("玩家不存在")
		}

		minPick := runtimeutil.ToIntContextValue(ctxData["min_pick"])
		if minPick < 1 {
			if glyph {
				minPick = 2 // 魔纹融合至少2张
			} else {
				minPick = 1 // 战纹碎击至少1张
			}
		}

		if len(selections) < minPick {
			return false, fmt.Errorf("至少需要选择%d张牌", minPick)
		}

		attackElement, _ := ctxData["attack_element"].(string)
		mismatchErr := "战纹碎击需选择彼此同系的牌"
		duplicateErr := ""
		if glyph {
			mismatchErr = "魔纹融合需选择彼此异系的牌"
			duplicateErr = "魔纹融合需选择元素互不相同的异系牌"
		}
		if err := validateRuneCardSelection(user, selections, attackElement, glyph, mismatchErr, duplicateErr); err != nil {
			return false, err
		}

		flow, err := model.RequirePromptFlow(ctxData, runeChoiceFlowID(glyph), runeChoiceLabel(glyph))
		if err != nil {
			return false, err
		}
		flow.PutSelection(runeChoiceStepCards, model.PromptFlowSelection{
			OptionIndexes: append([]int{}, selections...),
			Count:         len(selections),
		})
		maxY := runtimeutil.ToIntContextValue(ctxData["max_y"])
		if maxY > 0 {
			nextStep := "hom_rune_smash_y"
			if glyph {
				nextStep = "hom_glyph_fusion_y"
			}
			engineplayer.AdvancePromptFlowChoice(rt, ctxData, flow, runeChoiceStepY, nextStep)
			return true, nil // 消费当前选择步骤，继续弹出 Y 选择
		}

		return true, resolveRuneChoice(rt, ctxData, glyph)
	}
}

// handleRuneReforgeAllocate 处理符文改造的多选（战纹/魔纹分配）。
func handleRuneReforgeAllocate(rt engineplayer.ChoiceRuntime, playerID string, selections []int, ctxData map[string]interface{}) (bool, error) {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return false, fmt.Errorf("玩家不存在")
	}

	total := runtimeutil.ToIntContextValue(ctxData["total_runes"])
	if total <= 0 {
		total = 3
	}

	if len(selections) != 2 {
		return false, fmt.Errorf("请为战纹和魔纹分别选择数量")
	}

	warRunes := selections[0]
	magicRunes := selections[1]

	if warRunes < 0 || magicRunes < 0 {
		return false, fmt.Errorf("分配数量不能为负")
	}

	if warRunes+magicRunes != total {
		return false, fmt.Errorf("战纹和魔纹之和必须等于%d，当前：%d+%d=%d", total, warRunes, magicRunes, warRunes+magicRunes)
	}

	if user.Tokens == nil {
		user.Tokens = map[string]int{}
	}
	user.Tokens["hom_war_rune"] = warRunes
	user.Tokens["hom_magic_rune"] = magicRunes
	rt.Log(fmt.Sprintf("%s 的 [符文改造]：战纹=%d，魔纹=%d", user.Name, user.Tokens["hom_war_rune"], user.Tokens["hom_magic_rune"]))
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		rt.ApplyChoiceResumePoint(model.TurnStageActionStart)
	}
	return true, nil
}

func handleRuneY(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int, glyph bool) error {
	maxY := runtimeutil.ToIntContextValue(ctxData["max_y"])
	yValue := selectionIndex
	if yValue < 0 || yValue > maxY {
		return fmt.Errorf("无效的Y值")
	}
	flow, err := model.RequirePromptFlow(ctxData, runeChoiceFlowID(glyph), runeChoiceLabel(glyph))
	if err != nil {
		return err
	}
	flow.PutSelection(runeChoiceStepY, model.PromptFlowSelection{
		OptionIndexes: []int{selectionIndex},
		Count:         yValue,
	})
	return resolveRuneChoice(rt, ctxData, glyph)
}

func handleDualEchoTarget(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
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
	if target := rt.GetPlayers()[targetID]; target != nil {
		rt.Log(fmt.Sprintf("%s 的 [双重回响] 对 %s 造成%d点法术伤害", user.Name, target.Name, damage))
	}
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil && len(rt.GetPendingDamageQueue()) > 0 {
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
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	rawCtx, _ := ctxData["user_ctx"].(*model.Context)
	if rawCtx == nil || rawCtx.EventCtx == nil {
		return fmt.Errorf("英灵人形技能上下文丢失")
	}
	flow, err := model.RequirePromptFlow(ctxData, runeChoiceFlowID(glyph), runeChoiceLabel(glyph))
	if err != nil {
		return err
	}
	cardSelection := flow.Selection(runeChoiceStepCards)
	xVal := cardSelection.Count
	yVal := flow.Selection(runeChoiceStepY).Count
	if xVal <= 0 || yVal < 0 {
		return fmt.Errorf("X/Y 参数无效")
	}
	selected := append([]int{}, cardSelection.OptionIndexes...)
	if len(selected) != xVal {
		return fmt.Errorf("弃牌数量与X不一致")
	}

	attackElement, _ := ctxData["attack_element"].(string)
	mismatchErr := "战纹碎击需弃置彼此同系的牌"
	duplicateErr := ""
	if glyph {
		mismatchErr = "魔纹融合需弃置彼此异系的牌"
		duplicateErr = "魔纹融合需弃置元素互不相同的异系牌"
	}
	if err := validateRuneCardSelection(user, selected, attackElement, glyph, mismatchErr, duplicateErr); err != nil {
		return err
	}

	flipCount := 1 + yVal
	if err := applyRuneFlip(user, glyph, flipCount); err != nil {
		return err
	}

	removed, _ := engineplayer.RemoveCardsByIndicesFromHand(user, append([]int{}, selected...))
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
		if rt.GetPendingInterrupt() == nil && rawCtx.ResumeAttackMissPhase() {
			if rt.ResumePendingAttackMiss(rawCtx) {
				return nil
			}
		}
		if rt.GetPendingInterrupt() == nil && len(rt.GetPendingDamageQueue()) > 0 {
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

// ---------------------------------------------------------------------------
// Rune validation & mutation helpers
// ---------------------------------------------------------------------------

func runeChoiceFlowID(glyph bool) string {
	if glyph {
		return glyphFusionFlowID
	}
	return runeSmashFlowID
}

func runeChoiceLabel(glyph bool) string {
	if glyph {
		return "魔纹融合"
	}
	return "战纹碎击"
}

// validateRuneCardSelection checks that the selected hand indices satisfy the
// element-matching rules for 战纹碎击 (selected cards must be same element with each other)
// or 魔纹融合 (selected cards must be different elements from each other).
func validateRuneCardSelection(user *model.Player, selected []int, attackElement string, glyph bool, mismatchErr, duplicateErr string) error {
	seen := map[model.Element]bool{}
	for _, idx := range selected {
		if idx < 0 || idx >= len(user.Hand) {
			return fmt.Errorf("无效的手牌索引: %d", idx)
		}
		elem := user.Hand[idx].Element
		if glyph {
			// 魔纹融合：选择的牌彼此异系（元素互不相同）
			if duplicateErr != "" && seen[elem] {
				return fmt.Errorf(duplicateErr)
			}
			seen[elem] = true
			continue
		}
		// 战纹碎击：选择的牌彼此同系
		// 第一张牌确定元素，后续牌必须与第一张牌元素相同
		if len(seen) == 0 {
			seen[elem] = true
			continue
		}
		if !seen[elem] || len(seen) > 1 {
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
// remaining candidate list.
// For glyph mode (魔纹融合), it removes cards with the same element as the picked card (彼此异系).
// For rune smash mode (战纹碎击), it removes cards with different element from the picked card (彼此同系).
func filterRuneRemainingCandidates(user *model.Player, remaining []int, picked int, glyph bool) []int {
	nextRemaining := make([]int, 0, len(remaining))
	for _, idx := range remaining {
		if idx == picked {
			continue
		}
		if idx >= 0 && idx < len(user.Hand) && picked >= 0 && picked < len(user.Hand) {
			pickedElem := user.Hand[picked].Element
			currentElem := user.Hand[idx].Element
			if glyph && currentElem == pickedElem {
				// 魔纹融合：过滤掉相同元素的牌（彼此异系）
				continue
			}
			if !glyph && currentElem != pickedElem {
				// 战纹碎击：过滤掉不同元素的牌（彼此同系）
				continue
			}
		}
		nextRemaining = append(nextRemaining, idx)
	}
	return nextRemaining
}

// ---------------------------------------------------------------------------
// Card helpers
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Context value helpers
// ---------------------------------------------------------------------------
