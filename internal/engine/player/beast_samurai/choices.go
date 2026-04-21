// gameflow: 兽武士角色选择流。

package beast_samurai

import (
	"fmt"
	"strconv"

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
	case "bs_beast_return_x":
		return buildBeastReturnXPrompt(playerID, data)
	case "bs_reversal_x":
		return buildReversalXPrompt(playerID, data)
	case "bs_iaijutsu_style_mode":
		return buildIaijutsuStyleModePrompt(playerID, data)
	case "bs_alert_source_discard",
		"bs_beast_return_self_discard",
		"bs_beast_return_source_discard",
		"bs_iaijutsu_style_discard",
		"bs_reversal_target_discard":
		return buildDiscardPrompt(rt, playerID, player, data)
	default:
		return nil
	}
}

func buildBeastReturnXPrompt(playerID string, data map[string]interface{}) *model.Prompt {
	maxX := runtimeutil.ToIntContextValue(data["max_x"])
	options := make([]model.PromptOption, 0, maxX+1)
	for x := 0; x <= maxX; x++ {
		label := fmt.Sprintf("X=%d", x)
		if x == 0 {
			label = "X=0（不移除兽魂）"
		}
		options = append(options, model.PromptOption{
			ID:    fmt.Sprintf("%d", x),
			Label: label,
		})
	}
	return &model.Prompt{
		Type:       model.PromptConfirm,
		PlayerID:   playerID,
		ChoiceType: "bs_beast_return_x",
		Message:    fmt.Sprintf("【兽返】请选择要移除的兽魂数量（0-%d）：", maxX),
		Options:    options,
		Min:        1,
		Max:        1,
	}
}

func buildReversalXPrompt(playerID string, data map[string]interface{}) *model.Prompt {
	maxX := runtimeutil.ToIntContextValue(data["max_x"])
	options := make([]model.PromptOption, 0, maxX+1)
	for x := 0; x <= maxX; x++ {
		options = append(options, model.PromptOption{
			ID:    fmt.Sprintf("%d", x),
			Label: fmt.Sprintf("X=%d（目标将弃置%d张手牌）", x, x+2),
		})
	}
	return &model.Prompt{
		Type:       model.PromptConfirm,
		PlayerID:   playerID,
		ChoiceType: "bs_reversal_x",
		Message:    fmt.Sprintf("【逆反居合斩】请选择要移除的兽魂数量（0-%d）：", maxX),
		Options:    options,
		Min:        1,
		Max:        1,
	}
}

func buildIaijutsuStyleModePrompt(playerID string, data map[string]interface{}) *model.Prompt {
	modes := runtimeutil.ParseChoiceIntSlice(data["modes"])
	if len(modes) == 0 {
		modes = []int{0, 1}
	}
	options := make([]model.PromptOption, 0, len(modes))
	for _, mode := range modes {
		switch mode {
		case 0:
			options = append(options, model.PromptOption{ID: "0", Label: "摸1张牌"})
		case 1:
			options = append(options, model.PromptOption{ID: "1", Label: "弃1张牌"})
		}
	}
	if len(options) == 0 {
		return nil
	}
	return &model.Prompt{
		Type:       model.PromptConfirm,
		PlayerID:   playerID,
		ChoiceType: "bs_iaijutsu_style_mode",
		Message:    `【御魂流居合式】请选择"摸1张牌"或"弃1张牌"：`,
		Options:    options,
		Min:        1,
		Max:        1,
	}
}

// buildDiscardPrompt builds a card-discard prompt from the player's hand, using
// context data keys: discard_count, prompt (custom message), remaining_indices,
// discard_type, discard_element.
func buildDiscardPrompt(rt engineplayer.ChoiceRuntime, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	if player == nil || data == nil {
		return nil
	}
	promptChoiceType, _ := data["choice_type"].(string)

	var message string
	var min, max int

	if count, ok := data["discard_count"].(int); ok && count > 0 {
		min = count
		max = count
		message = fmt.Sprintf("请弃置 %d 张牌：", count)
		if customMsg, ok := data["prompt"].(string); ok && customMsg != "" {
			message = customMsg
		}
	} else {
		if v, ok := data["min"].(int); ok {
			min = v
		} else {
			min = 1
		}
		if v, ok := data["max"].(int); ok && v > 0 {
			max = v
		} else {
			max = len(player.Hand)
		}
		if customMsg, ok := data["prompt"].(string); ok && customMsg != "" {
			message = customMsg
		} else {
			message = fmt.Sprintf("请选择 %d-%d 张牌弃置：", min, max)
		}
	}

	remainingIndices := runtimeutil.ParseChoiceIntSlice(data["remaining_indices"])
	allowedSet := map[int]struct{}{}
	if len(remainingIndices) > 0 {
		for _, idx := range remainingIndices {
			allowedSet[idx] = struct{}{}
		}
	}

	discardType, _ := data["discard_type"].(model.CardType)
	discardElement, _ := data["discard_element"].(model.Element)

	var options []model.PromptOption
	for i, card := range player.Hand {
		if len(allowedSet) > 0 {
			if _, ok := allowedSet[i]; !ok {
				continue
			}
		}
		if discardType != "" && card.Type != discardType {
			continue
		}
		if discardElement != "" && card.Element != discardElement {
			continue
		}
		options = append(options, model.PromptOption{
			ID:    strconv.Itoa(i),
			Label: fmt.Sprintf("%d: %s", i+1, promptfmt.FormatCardInfo(card)),
		})
	}

	return &model.Prompt{
		Type:       model.PromptChooseCards,
		PlayerID:   playerID,
		ChoiceType: promptChoiceType,
		Message:    message,
		Options:    options,
		Min:        min,
		Max:        max,
	}
}

// ---------------------------------------------------------------------------
// HandleChoice
// ---------------------------------------------------------------------------

func (choiceHandler) HandleChoice(rt engineplayer.ChoiceRuntime, playerID string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "bs_beast_return_x":
		return true, handleBeastReturnX(rt, ctxData, selectionIndex)
	case "bs_reversal_x":
		return true, handleReversalX(rt, ctxData, selectionIndex)
	case "bs_iaijutsu_style_mode":
		return true, handleIaijutsuStyleMode(rt, ctxData, selectionIndex)
	case "bs_alert_source_discard":
		return true, handleAlertSourceDiscard(rt, ctxData, selectionIndex)
	case "bs_beast_return_self_discard":
		return true, handleBeastReturnSelfDiscard(rt, ctxData, selectionIndex)
	case "bs_beast_return_source_discard":
		return true, handleBeastReturnSourceDiscard(rt, ctxData, selectionIndex)
	case "bs_iaijutsu_style_discard":
		return true, handleIaijutsuStyleDiscard(rt, ctxData, selectionIndex)
	case "bs_reversal_target_discard":
		return true, handleReversalTargetDiscard(rt, ctxData, selectionIndex)
	default:
		return false, nil
	}
}

// ---------------------------------------------------------------------------
// Individual choice-type handlers
// ---------------------------------------------------------------------------

func handleBeastReturnX(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	sourceID, _ := ctxData["source_id"].(string)
	user := rt.LookupPlayer(userID)
	source := rt.LookupPlayer(sourceID)
	if user == nil {
		return fmt.Errorf("兽返执行者不存在")
	}

	maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
	if current := BeastSoul(user); maxX > current {
		maxX = current
	}
	if selectionIndex < 0 || selectionIndex > maxX {
		return fmt.Errorf("无效的X值: %d", selectionIndex)
	}

	x := selectionIndex
	consumed := consumeBeastSoul(user, x)
	resumePoint := resumePointFromCtx(ctxData, model.CombatStageCalcDamage)
	rt.Log(fmt.Sprintf("%s 的 [兽返] 结算：移除%d点兽魂，残心同步+%d", user.Name, consumed, consumed))

	selfDiscardCount := x
	if selfDiscardCount > len(user.Hand) {
		selfDiscardCount = len(user.Hand)
	}
	if selfDiscardCount > 0 {
		replaceDiscardInterrupt(rt, user.ID, map[string]interface{}{
			"choice_type":   "bs_beast_return_self_discard",
			"user_id":       user.ID,
			"source_id":     sourceID,
			"x_value":       x,
			"discard_count": selfDiscardCount,
			"prompt":        fmt.Sprintf("【兽返】请选择弃置%d张手牌：", selfDiscardCount),
			"resume_phase":  resumePoint,
		})
		return nil
	}
	if source != nil && len(source.Hand) > 0 {
		replaceDiscardInterrupt(rt, source.ID, map[string]interface{}{
			"choice_type":   "bs_beast_return_source_discard",
			"user_id":       user.ID,
			"source_id":     source.ID,
			"discard_count": 1,
			"prompt":        "【兽返】请选择弃置1张手牌：",
			"resume_phase":  resumePoint,
		})
		return nil
	}
	finishResume(rt, resumePoint)
	return nil
}

func handleReversalX(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	targetID, _ := ctxData["target_id"].(string)
	user := rt.LookupPlayer(userID)
	target := rt.LookupPlayer(targetID)
	if user == nil || target == nil {
		return fmt.Errorf("逆反居合斩上下文目标不存在")
	}

	maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
	if current := BeastSoul(user); maxX > current {
		maxX = current
	}
	if selectionIndex < 0 || selectionIndex > maxX {
		return fmt.Errorf("无效的X值: %d", selectionIndex)
	}

	x := selectionIndex
	consumed := consumeBeastSoul(user, x)
	if user.TurnState.SkillFlowState == nil {
		user.TurnState.SkillFlowState = map[string]int{}
	}
	user.TurnState.SkillFlowState["bs_reversal_pending_x"] = x

	// Zero-out the pending attack damage (if still queued) by setting Damage to 0.
	zeroPendingAttackDamage(rt, ctxData)

	need := x + 2
	resumePoint := resumePointFromCtx(ctxData, model.CombatStageCalcDamage)
	rt.Log(fmt.Sprintf("%s 的 [逆反居合斩] 结算：移除%d点兽魂，残心同步+%d，改写本次攻击为弃牌效果", user.Name, consumed, consumed))

	discardCount := need
	if discardCount > len(target.Hand) {
		discardCount = len(target.Hand)
	}
	if discardCount > 0 {
		replaceDiscardInterrupt(rt, target.ID, map[string]interface{}{
			"choice_type":   "bs_reversal_target_discard",
			"user_id":       user.ID,
			"target_id":     target.ID,
			"x_value":       x,
			"need_count":    need,
			"discard_count": discardCount,
			"prompt":        fmt.Sprintf("【逆反居合斩】请选择弃置%d张手牌：", discardCount),
			"resume_phase":  resumePoint,
			"user_ctx":      ctxData["user_ctx"],
		})
		return nil
	}
	finishReversal(rt, ctxData, target, need, 0, resumePoint)
	return nil
}

func handleIaijutsuStyleMode(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.LookupPlayer(userID)
	if user == nil {
		return fmt.Errorf("御魂流居合式执行者不存在")
	}

	modes := runtimeutil.ParseChoiceIntSlice(ctxData["modes"])
	if len(modes) == 0 {
		modes = []int{0, 1}
	}
	if selectionIndex < 0 || selectionIndex >= len(modes) {
		return fmt.Errorf("无效的模式选择: %d", selectionIndex)
	}
	mode := modes[selectionIndex]
	resumePoint := resumePointFromCtx(ctxData, model.TurnStageActionStart)

	soulAfter := AddBeastSoul(user, 1, true)
	if InIaijutsuForm(user) {
		zanshinAfter := AddZanshin(user, 1)
		rt.Log(fmt.Sprintf("%s 的 [御魂流居合式] 生效：兽魂+1（当前%d），因已处于御魂流居合形态，残心+1（当前%d）", user.Name, soulAfter, zanshinAfter))
	} else {
		EnterIaijutsuForm(user)
		rt.Log(fmt.Sprintf("%s 的 [御魂流居合式] 生效：兽魂+1（当前%d），进入御魂流居合形态", user.Name, soulAfter))
	}

	switch mode {
	case 0:
		rt.DrawCards(user.ID, 1)
		finishResume(rt, resumePoint)
		return nil
	case 1:
		discardCount := 1
		if len(user.Hand) < discardCount {
			discardCount = len(user.Hand)
		}
		if discardCount > 0 {
			replaceDiscardInterrupt(rt, user.ID, map[string]interface{}{
				"choice_type":   "bs_iaijutsu_style_discard",
				"user_id":       user.ID,
				"discard_count": discardCount,
				"prompt":        "【御魂流居合式】请选择弃置1张手牌：",
				"resume_phase":  resumePoint,
			})
			return nil
		}
		finishResume(rt, resumePoint)
		return nil
	default:
		return fmt.Errorf("未知的御魂流居合式模式: %d", mode)
	}
}

func handleAlertSourceDiscard(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	actorID, _ := ctxData["actor_id"].(string)
	user := rt.LookupPlayer(userID)
	actor := rt.LookupPlayer(actorID)
	if user == nil || actor == nil {
		return fmt.Errorf("兽魂警戒弃牌上下文不存在")
	}

	removed := removeCardsByIndicesFromHand(actor, []int{selectionIndex})
	if len(removed) > 0 {
		rt.NotifyCardRevealed(actor.ID, removed, "discard")
		rt.AppendToDiscard(removed)
	}
	if discardedMagicCount(removed) > 0 {
		after := AddBeastSoul(user, 1, false)
		rt.Log(fmt.Sprintf("%s 的 [兽魂警戒] 生效：%s 展示弃牌中含法术牌，兽魂+1（当前%d）", user.Name, actor.Name, after))
	}
	finishResume(rt, resumePointFromCtx(ctxData, model.TurnStageActionExecution))
	return nil
}

func handleBeastReturnSelfDiscard(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	sourceID, _ := ctxData["source_id"].(string)
	user := rt.LookupPlayer(userID)
	source := rt.LookupPlayer(sourceID)
	if user == nil {
		return fmt.Errorf("兽返弃牌执行者不存在")
	}

	removed := removeCardsByIndicesFromHand(user, []int{selectionIndex})
	if len(removed) > 0 {
		rt.NotifyCardRevealed(user.ID, removed, "discard")
		rt.AppendToDiscard(removed)
	}

	resumePoint := resumePointFromCtx(ctxData, model.CombatStageCalcDamage)
	if source != nil && len(source.Hand) > 0 {
		replaceDiscardInterrupt(rt, source.ID, map[string]interface{}{
			"choice_type":   "bs_beast_return_source_discard",
			"user_id":       user.ID,
			"source_id":     source.ID,
			"discard_count": 1,
			"prompt":        "【兽返】请选择弃置1张手牌：",
			"resume_phase":  resumePoint,
		})
		return nil
	}
	finishResume(rt, resumePoint)
	return nil
}

func handleBeastReturnSourceDiscard(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	sourceID, _ := ctxData["source_id"].(string)
	user := rt.LookupPlayer(userID)
	source := rt.LookupPlayer(sourceID)
	if user == nil || source == nil {
		return fmt.Errorf("兽返来源弃牌上下文不存在")
	}

	removed := removeCardsByIndicesFromHand(source, []int{selectionIndex})
	if len(removed) > 0 {
		rt.NotifyCardRevealed(source.ID, removed, "discard")
		rt.AppendToDiscard(removed)
	}
	if discardedMagicCount(removed) > 0 {
		after := AddBeastSoul(user, 1, false)
		rt.Log(fmt.Sprintf("%s 的 [兽返] 生效：%s 弃牌中含法术牌，兽魂+1（当前%d）", user.Name, source.Name, after))
	}
	finishResume(rt, resumePointFromCtx(ctxData, model.CombatStageCalcDamage))
	return nil
}

func handleIaijutsuStyleDiscard(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.LookupPlayer(userID)
	if user == nil {
		return fmt.Errorf("御魂流居合式弃牌执行者不存在")
	}

	removed := removeCardsByIndicesFromHand(user, []int{selectionIndex})
	if len(removed) > 0 {
		rt.NotifyCardRevealed(user.ID, removed, "discard")
		rt.AppendToDiscard(removed)
	}
	finishResume(rt, resumePointFromCtx(ctxData, model.TurnStageActionStart))
	return nil
}

func handleReversalTargetDiscard(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	targetID, _ := ctxData["target_id"].(string)
	target := rt.LookupPlayer(targetID)
	if target == nil {
		return fmt.Errorf("逆反居合斩目标不存在")
	}

	need := runtimeutil.ToIntContextValue(ctxData["need_count"])
	removed := removeCardsByIndicesFromHand(target, []int{selectionIndex})
	if len(removed) > 0 {
		rt.NotifyCardRevealed(target.ID, removed, "discard")
		rt.AppendToDiscard(removed)
	}

	// Accumulate actually discarded count.
	actualSoFar := runtimeutil.ToIntContextValue(ctxData["actual_discarded"]) + len(removed)
	ctxData["actual_discarded"] = actualSoFar

	remainingNeeded := need - actualSoFar
	if remainingNeeded > 0 && len(target.Hand) > 0 {
		// More discards needed, update context and re-prompt.
		nextCount := remainingNeeded
		if nextCount > len(target.Hand) {
			nextCount = len(target.Hand)
		}
		ctxData["discard_count"] = nextCount
		ctxData["prompt"] = fmt.Sprintf("【逆反居合斩】请选择弃置%d张手牌：", nextCount)
		replaceDiscardInterrupt(rt, target.ID, ctxData)
		return nil
	}

	finishReversal(rt, ctxData, target, need, actualSoFar, resumePointFromCtx(ctxData, model.CombatStageCalcDamage))
	return nil
}

// ---------------------------------------------------------------------------
// Resume helpers
// ---------------------------------------------------------------------------

// resumePointFromCtx extracts the resume_phase from ctxData, falling back to
// the provided default if absent or invalid.
func resumePointFromCtx(ctxData map[string]interface{}, fallback interface{}) interface{} {
	if raw, ok := ctxData["resume_phase"]; ok {
		if _, valid := choiceResumePointValue(raw); valid {
			return raw
		}
	}
	return fallback
}

// choiceResumePointValue mirrors the engine-package function of the same name.
func choiceResumePointValue(raw interface{}) (interface{}, bool) {
	switch value := raw.(type) {
	case model.TurnStage:
		if value != "" && model.IsKnownTurnStage(value) {
			return value, true
		}
	case model.CombatStage:
		if value != model.CombatStageNone && model.IsKnownCombatStage(value) {
			return value, true
		}
	case model.Subflow:
		if value != model.SubflowNone && model.IsKnownSubflow(value) {
			return value, true
		}
	}
	return nil, false
}

// finishResume replicates beastSamuraiFinishResume: pop interrupt, then route
// to the appropriate next phase.
func finishResume(rt engineplayer.ChoiceRuntime, resumePoint interface{}) {
	rt.PopInterrupt()
	if rt.HasPendingInterrupt() {
		return
	}
	if _, ok := choiceResumePointValue(resumePoint); ok {
		rt.ApplyChoiceResumePoint(resumePoint)
		return
	}
	if rt.PendingDamageQueueLen() > 0 {
		rt.EnterDamageResolution(nil)
		return
	}
	if rt.ActionQueueLen() > 0 {
		rt.EnterActionExecutionStage()
		return
	}
	rt.ApplyChoiceResumePoint(model.TurnStageExtraAction)
}

// finishReversal replicates beastSamuraiFinishReversal: apply morale loss if
// the target didn't discard enough, then resume attack-hit flow and finish.
func finishReversal(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, target *model.Player, need, actualDiscarded int, resumePoint interface{}) {
	if target != nil && actualDiscarded < need {
		userName := "兽灵武士"
		if userID, ok := ctxData["user_id"].(string); ok {
			if user := rt.LookupPlayer(userID); user != nil {
				userName = user.Name
			}
		}
		loss := applyCampMoraleLoss(rt, target.Camp, 1)
		if loss > 0 {
			rt.Log(fmt.Sprintf("%s 的 [逆反居合斩] 生效：%s 实际弃牌%d/%d，%s方士气-%d", userName, target.Name, actualDiscarded, need, target.Camp, loss))
		} else {
			rt.Log(fmt.Sprintf("%s 的 [逆反居合斩] 生效：%s 实际弃牌%d/%d，但%s方士气已触及下限", userName, target.Name, actualDiscarded, need, target.Camp))
		}
	}
	// Resume the pending attack-hit flow (marks attack-hit processed, enters
	// damage resolution if applicable).
	rt.ResumePendingAttackHit(ctxData)
	finishResume(rt, resumePoint)
}

// ---------------------------------------------------------------------------
// Token helpers
// ---------------------------------------------------------------------------

// consumeBeastSoul removes up to `amount` beast soul tokens and adds the same
// amount to zanshin.  Returns the actual amount consumed.
func consumeBeastSoul(player *model.Player, amount int) int {
	if player == nil || amount <= 0 {
		return 0
	}
	current := BeastSoul(player)
	if amount > current {
		amount = current
	}
	if amount <= 0 {
		return 0
	}
	AddBeastSoul(player, -amount, true)
	AddZanshin(player, amount)
	return amount
}

// ---------------------------------------------------------------------------
// Interrupt helpers
// ---------------------------------------------------------------------------

// replaceDiscardInterrupt replaces the current pending interrupt with a
// normalized discard-choice context and re-notifies.
func replaceDiscardInterrupt(rt engineplayer.ChoiceRuntime, playerID string, ctxData map[string]interface{}) {
	ctxData["discard_subflow"] = true
	if _, ok := ctxData["choice_type"].(string); !ok || ctxData["choice_type"] == "" {
		ctxData["choice_type"] = "system_discard_cards"
	}
	_ = rt.ReplacePendingInterruptContext(ctxData)
	rt.ReplacePendingInterruptPlayerID(playerID)
	rt.NotifyInterruptPrompt()
}

// ---------------------------------------------------------------------------
// Card helpers
// ---------------------------------------------------------------------------

// discardedMagicCount returns how many of the given cards are magic-type.
func discardedMagicCount(cards []model.Card) int {
	count := 0
	for _, card := range cards {
		if card.Type == model.CardTypeMagic {
			count++
		}
	}
	return count
}

// removeCardsByIndicesFromHand removes the cards at the given indices from the
// player's hand and returns the removed cards. Indices must be valid.
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
// Combat helpers
// ---------------------------------------------------------------------------

// zeroPendingAttackDamage walks the pending-damage queue and sets Damage to 0
// for any attack-type damage whose source/target match the context's
// user_ctx EventCtx (if present).
func zeroPendingAttackDamage(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}) {
	rawCtx, _ := ctxData["user_ctx"].(*model.Context)
	if rawCtx == nil || rawCtx.EventCtx == nil {
		return
	}
	// Walk the pending damage queue via ChoiceRuntime and zero-out the matching attack damage.
	for i := 0; i < rt.PendingDamageQueueLen(); i++ {
		pd, ok := rt.GetPendingDamage(i)
		if !ok || pd == nil {
			continue
		}
		if pd.SourceID == rawCtx.User.ID && pd.DamageType == model.AttackDamage && pd.Card != nil {
			pd.Damage = 0
		}
	}
}

// applyCampMoraleLoss reduces camp morale by the given amount (respecting
// floor). Returns the actual loss applied.
func applyCampMoraleLoss(rt engineplayer.ChoiceRuntime, camp model.Camp, wantLoss int) int {
	if wantLoss <= 0 {
		return 0
	}
	current := rt.GetCampMorale(string(camp))
	// Morale floor is 0 in the simplified version.
	maxLoss := current
	if maxLoss < 0 {
		maxLoss = 0
	}
	actual := wantLoss
	if actual > maxLoss {
		actual = maxLoss
	}
	if actual <= 0 {
		return 0
	}
	rt.ModifyGem(string(camp), -actual)
	return actual
}
