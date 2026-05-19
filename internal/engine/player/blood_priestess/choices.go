// gameflow: 血祭司角色选择流。

package blood_priestess

import (
	"fmt"
	"sort"

	"starcup-engine/internal/engine/core/runtimeutil"
	"starcup-engine/internal/engine/hook/promptfmt"
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

type choiceHandler struct{}

const (
	bloodCurseDiscardFlowID    = "bp_blood_curse_discard"
	bloodCurseDiscardNeedStep  = "need"
	bloodCurseDiscardCardsStep = "cards"
)

func NewChoiceHandler() engineplayer.ChoiceHandler {
	return choiceHandler{}
}

// --------------- BuildPrompt ---------------

func (choiceHandler) BuildPrompt(rt engineplayer.ChoiceRuntime, choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "bp_shared_life_target":
		return buildSharedLifeTargetPrompt(rt, playerID, data)
	case "bp_blood_sorrow_mode":
		return buildBloodSorrowModePrompt(playerID)
	case "bp_blood_sorrow_target":
		return buildBloodSorrowTargetPrompt(rt, playerID, data)
	case "bp_blood_wail_x":
		return buildBloodWailXPrompt(playerID)
	case "bp_curse_discard":
		return buildCurseDiscardPrompt(playerID, player, data)
	default:
		return nil
	}
}

func buildSharedLifeTargetPrompt(rt engineplayer.ChoiceRuntime, playerID string, data map[string]interface{}) *model.Prompt {
	targetIDs := runtimeutil.ParseStringSliceContextValue(data["target_ids"])
	options := make([]model.PromptOption, 0, len(targetIDs))
	for _, tid := range targetIDs {
		if p := rt.GetPlayers()[tid]; p != nil {
			options = append(options, model.PromptOption{ID: tid, Label: p.Name})
		}
	}
	return &model.Prompt{
		Type:         model.PromptConfirm,
		PlayerID:     playerID,
		ChoiceType:   "bp_shared_life_target",
		Message:      "【同生共死】请选择放置目标：",
		Options:      options,
		Min:          1,
		Max:          1,
		Presentation: &model.PromptPresentation{Kind: model.PresentationTargetPicker, TargetFilter: "custom"},
	}
}

func buildBloodSorrowModePrompt(playerID string) *model.Prompt {
	return &model.Prompt{
		Type:       model.PromptConfirm,
		PlayerID:   playerID,
		ChoiceType: "bp_blood_sorrow_mode",
		Message:    "【血之哀伤】请选择后续效果（结算时会先对自己造成2点法术伤害）：",
		Options: []model.PromptOption{
			{ID: "0", Label: "移除同生共死"},
			{ID: "1", Label: "转移同生共死目标"},
		},
		Min:          1,
		Max:          1,
		Presentation: &model.PromptPresentation{Kind: model.PresentationBranchSelect, Layout: "overlay"},
	}
}

func buildBloodSorrowTargetPrompt(rt engineplayer.ChoiceRuntime, playerID string, data map[string]interface{}) *model.Prompt {
	targetIDs := runtimeutil.ParseStringSliceContextValue(data["target_ids"])
	options := make([]model.PromptOption, 0, len(targetIDs))
	for _, tid := range targetIDs {
		if p := rt.GetPlayers()[tid]; p != nil {
			options = append(options, model.PromptOption{ID: tid, Label: p.Name})
		}
	}
	return &model.Prompt{
		Type:         model.PromptConfirm,
		PlayerID:     playerID,
		ChoiceType:   "bp_blood_sorrow_target",
		Message:      "【血之哀伤】请选择新的同生共死目标：",
		Options:      options,
		Min:          1,
		Max:          1,
		Presentation: &model.PromptPresentation{Kind: model.PresentationTargetPicker, TargetFilter: "custom"},
	}
}

func buildBloodWailXPrompt(playerID string) *model.Prompt {
	return &model.Prompt{
		Type:       model.PromptConfirm,
		PlayerID:   playerID,
		ChoiceType: "bp_blood_wail_x",
		Message:    "【血之悲鸣】请选择X值：",
		Options: []model.PromptOption{
			{ID: "0", Label: "X=0（伤害=1）"},
			{ID: "1", Label: "X=1（伤害=2）"},
			{ID: "2", Label: "X=2（伤害=3）"},
		},
		Min:          1,
		Max:          1,
		Presentation: &model.PromptPresentation{Kind: model.PresentationNumeric, NumericBase: 0},
	}
}

func buildCurseDiscardPrompt(playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	flow, err := model.RequirePromptFlow(data, bloodCurseDiscardFlowID, "血之诅咒弃牌")
	if err != nil {
		return nil
	}
	discardCount := flow.Selection(bloodCurseDiscardNeedStep).Count
	if discardCount < 0 {
		discardCount = 0
	}
	if player != nil && discardCount > len(player.Hand) {
		discardCount = len(player.Hand)
	}
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
		ChoiceType:   "bp_curse_discard",
		Message:      fmt.Sprintf("【血之诅咒】请弃置%d张手牌：", discardCount),
		Options:      options,
		Min:          discardCount,
		Max:          discardCount,
		Presentation: &model.PromptPresentation{Kind: model.PresentationCardPicker, CardSource: "hand"},
	}
}

func initBloodCurseDiscardFlow(discardNeed int) *model.PromptFlowState {
	flow := model.NewPromptFlowState(bloodCurseDiscardFlowID, bloodCurseDiscardCardsStep)
	flow.PutSelection(bloodCurseDiscardNeedStep, model.PromptFlowSelection{Count: discardNeed})
	flow.PutSelection(bloodCurseDiscardCardsStep, model.PromptFlowSelection{})
	return flow
}

// --------------- HandleChoice ---------------

func (choiceHandler) HandleChoice(rt engineplayer.ChoiceRuntime, _ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "bp_shared_life_target":
		return true, handleSharedLifeTargetChoice(rt, ctxData, selectionIndex)
	case "bp_blood_sorrow_mode":
		return true, handleBloodSorrowModeChoice(rt, ctxData, selectionIndex)
	case "bp_blood_sorrow_target":
		return true, handleBloodSorrowTargetChoice(rt, ctxData, selectionIndex)
	case "bp_blood_wail_x":
		return true, handleBloodWailXChoice(rt, ctxData, selectionIndex)
	case "bp_curse_discard":
		return true, handleCurseDiscardChoice(rt, ctxData, selectionIndex)
	default:
		return false, nil
	}
}

// bp_shared_life_target: 玩家选择同生共死放置目标后，消耗专属卡、创建摸牌上下文并开始摸牌。
func handleSharedLifeTargetChoice(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}

	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
	if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	target := rt.GetPlayers()[targetIDs[selectionIndex]]
	if target == nil {
		return fmt.Errorf("同生共死目标不存在")
	}
	if user.Character == nil {
		return fmt.Errorf("角色信息缺失")
	}

	linkCard, ok := user.ConsumeExclusiveCard(user.Character.ID, "同生共死")
	if !ok {
		return fmt.Errorf("未找到【同生共死】专属技能卡")
	}

	drawCtx := rt.NewDrawContext(user, 2, "bp_shared_life_draw")
	if drawCtx == nil {
		user.RestoreExclusiveCard(linkCard)
		return fmt.Errorf("同生共死摸牌上下文创建失败")
	}
	drawCtx.Selections["draw_resume_phase"] = model.TurnStageActionExecution
	rt.AppendFlowContinuation(model.FlowContinuation{
		Kind:      model.FlowContinuationAfterDraw,
		RoleID:    "blood_priestess",
		PlayerID:  user.ID,
		TargetIDs: []string{target.ID},
		Data: map[string]any{
			"card": linkCard,
		},
	})
	rt.Log(fmt.Sprintf("[DEBUG] FlowContinuation appended: kind=%s, role=%s, player=%s", model.FlowContinuationAfterDraw, "blood_priestess", user.ID))
	rt.StartDraw(drawCtx)
	rt.Log(fmt.Sprintf("%s 发动 [同生共死]：先摸2张牌，待爆牌结算后放置于 %s 面前", user.Name, target.Name))

	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		if !rt.RoutePendingDamageWithReturn(model.TurnStageActionExecution) {
			rt.RestorePhaseAfterInterruptedDraw(drawCtx)
		}
	}
	return nil
}

// bp_blood_sorrow_mode: 选择血之哀伤模式（移除 or 转移同生共死）。
func handleBloodSorrowModeChoice(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	if selectionIndex < 0 || selectionIndex > 1 {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}

	if !runtimeutil.ToBoolContextValue(ctxData["damage_queued"]) {
		rt.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   user.ID,
			Damage:     2,
			DamageType: model.MagicAttack,
		})
		ctxData["damage_queued"] = true
	}

	// mode 0: 移除同生共死
	if selectionIndex == 0 {
		rt.AppendFlowContinuation(model.FlowContinuation{
			Kind:     model.FlowContinuationAfterDamage,
			RoleID:   "blood_priestess",
			PlayerID: user.ID,
			SkillID:  "bp_blood_sorrow",
			Data: map[string]any{
				"mode": "remove",
			},
		})
		rt.Log(fmt.Sprintf("%s 发动 [血之哀伤]：先对自己造成2点法术伤害，伤害结算后移除【同生共死】", user.Name))
		rt.PopInterrupt()
		if rt.GetPendingInterrupt() == nil {
			rt.EnterDamageResolution(model.TurnStageActionStart)
		}
		return nil
	}

	// mode 1: 转移 — 切换到目标选择子流程
	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
	if len(targetIDs) == 0 {
		return fmt.Errorf("无可转移目标")
	}
	ctxData["choice_type"] = "bp_blood_sorrow_target"
	intr := rt.GetPendingInterrupt()
	if intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
	return nil
}

// bp_blood_sorrow_target: 选择血之哀伤转移目标。
func handleBloodSorrowTargetChoice(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}

	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
	if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}

	if !runtimeutil.ToBoolContextValue(ctxData["damage_queued"]) {
		rt.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   user.ID,
			Damage:     2,
			DamageType: model.MagicAttack,
		})
		ctxData["damage_queued"] = true
	}

	target := rt.GetPlayers()[targetIDs[selectionIndex]]
	if target == nil {
		return fmt.Errorf("转移目标不存在")
	}

	rt.AppendFlowContinuation(model.FlowContinuation{
		Kind:     model.FlowContinuationAfterDamage,
		RoleID:   "blood_priestess",
		PlayerID: user.ID,
		SkillID:  "bp_blood_sorrow",
		Data: map[string]any{
			"mode":      "transfer",
			"target_id": target.ID,
		},
	})
	rt.Log(fmt.Sprintf("%s 发动 [血之哀伤]：先对自己造成2点法术伤害，伤害结算后将【同生共死】转移至 %s", user.Name, target.Name))
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		rt.EnterDamageResolution(model.TurnStageActionStart)
	}
	return nil
}

// bp_blood_wail_x: 选择血之悲鸣 X 值，对目标和自己各造成 (X+1) 点法术伤害。
func handleBloodWailXChoice(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	if !InBleedingForm(user) {
		return fmt.Errorf("仅流血形态下可发动血之悲鸣")
	}

	targetID, _ := ctxData["target_id"].(string)
	target := rt.GetPlayers()[targetID]
	if target == nil {
		return fmt.Errorf("目标角色不存在")
	}
	if selectionIndex < 0 || selectionIndex > 2 {
		return fmt.Errorf("无效的X值")
	}

	damage := selectionIndex + 1
	rt.AddPendingDamage(model.PendingDamage{
		SourceID:   user.ID,
		TargetID:   target.ID,
		Damage:     damage,
		DamageType: model.MagicAttack,
	})
	rt.AddPendingDamage(model.PendingDamage{
		SourceID:   user.ID,
		TargetID:   user.ID,
		Damage:     damage,
		DamageType: model.MagicAttack,
	})
	rt.Log(fmt.Sprintf("%s 发动 [血之悲鸣]：对 %s 和自己各造成%d点法术伤害", user.Name, target.Name, damage))

	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		rt.RoutePendingDamageOr(model.TurnStageExtraAction, nil)
	}
	return nil
}

// bp_curse_discard: 血之诅咒弃牌选择（单张选择，多轮触发）。
func handleCurseDiscardChoice(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	if selectionIndex < 0 {
		return fmt.Errorf("请在手牌区选择需要弃置的手牌后确认")
	}

	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}

	discardNeed := runtimeutil.ToIntContextValue(ctxData["discard_count"])
	flow, err := model.RequirePromptFlow(ctxData, bloodCurseDiscardFlowID, "血之诅咒弃牌")
	if err != nil {
		return err
	}
	if flowNeed := flow.Selection(bloodCurseDiscardNeedStep).Count; flowNeed > 0 {
		discardNeed = flowNeed
	}
	if discardNeed < 0 {
		discardNeed = 0
	}
	if discardNeed > len(user.Hand) {
		discardNeed = len(user.Hand)
	}
	if discardNeed == 0 {
		rt.PopInterrupt()
		if rt.GetPendingInterrupt() == nil {
			rt.RoutePendingDamageOr(model.TurnStageExtraAction, func() {
				rt.EnterExtraActionStage()
			})
		}
		return nil
	}

	// Check if we're in batch multi-select flow (sequential_remaining key set by runSequentialChoiceSelections)
	_, inBatchFlow := ctxData["sequential_remaining"]
	sequentialRemaining := runtimeutil.ToIntContextValue(ctxData["sequential_remaining"])

	// Batch multi-select flow: accumulate indices and remove at the end
	if inBatchFlow {
		// Validate selection index against current hand (hand unchanged during accumulation)
		if selectionIndex < 0 || selectionIndex >= len(user.Hand) {
			return fmt.Errorf("无效的弃牌索引: %d", selectionIndex)
		}
		selectedIndices := append([]int{}, flow.Selection(bloodCurseDiscardCardsStep).OptionIndexes...)
		for _, idx := range selectedIndices {
			if idx == selectionIndex {
				return fmt.Errorf("不能重复选择同一张牌")
			}
		}
		selectedIndices = append(selectedIndices, selectionIndex)
		flow.PutSelection(bloodCurseDiscardCardsStep, model.PromptFlowSelection{OptionIndexes: selectedIndices})

		// Only remove cards when this is the last iteration
		if sequentialRemaining == 0 {
			// Remove all selected cards using sorted removal to handle index shifting
			removed := removeCardsByIndicesSorted(&user.Hand, selectedIndices)
			for _, c := range removed {
				rt.NotifyCardRevealed(user.ID, []model.Card{c}, "discard")
			}
			rt.AppendToDiscard(removed)
			rt.Log(fmt.Sprintf("%s 的 [血之诅咒] 后续：弃置%d张牌", user.Name, len(removed)))

			rt.PopInterrupt()
			if rt.GetPendingInterrupt() == nil {
				rt.RoutePendingDamageOr(model.TurnStageExtraAction, func() {
					rt.EnterExtraActionStage()
				})
			}
		} else {
			// Persist context for next iteration in batch flow
			intr := rt.GetPendingInterrupt()
			if intr != nil {
				intr.Context = ctxData
			}
		}
		return nil
	}

	// Single-pick flow: remove one card at a time
	if selectionIndex < 0 || selectionIndex >= len(user.Hand) {
		return fmt.Errorf("无效的弃牌索引: %d", selectionIndex)
	}

	discarded := user.Hand[selectionIndex]
	user.Hand = append(user.Hand[:selectionIndex], user.Hand[selectionIndex+1:]...)
	rt.NotifyCardRevealed(user.ID, []model.Card{discarded}, "discard")
	rt.AppendToDiscard([]model.Card{discarded})
	rt.Log(fmt.Sprintf("%s 的 [血之诅咒] 后续：弃置1张牌", user.Name))

	discardNeed--
	if discardNeed > 0 {
		// 还有牌需要弃置，更新上下文重新触发
		ctxData["discard_count"] = discardNeed
		flow.PutSelection(bloodCurseDiscardNeedStep, model.PromptFlowSelection{Count: discardNeed})
		flow.PutSelection(bloodCurseDiscardCardsStep, model.PromptFlowSelection{})
		rt.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: user.ID,
			Context:  ctxData,
		})
		return nil
	}

	// 弃牌完成
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		rt.RoutePendingDamageOr(model.TurnStageExtraAction, func() {
			rt.EnterExtraActionStage()
		})
	}
	return nil
}

// --------------- helpers ---------------

// removeCardsByIndicesSorted 从手牌中按索引移除卡牌（从大到小排序避免索引位移）。
func removeCardsByIndicesSorted(hand *[]model.Card, indices []int) []model.Card {
	if hand == nil || len(indices) == 0 {
		return nil
	}
	sorted := make([]int, len(indices))
	copy(sorted, indices)
	sort.Sort(sort.Reverse(sort.IntSlice(sorted)))
	var removed []model.Card
	for _, idx := range sorted {
		if idx < 0 || idx >= len(*hand) {
			continue
		}
		removed = append(removed, (*hand)[idx])
		*hand = append((*hand)[:idx], (*hand)[idx+1:]...)
	}
	return removed
}
