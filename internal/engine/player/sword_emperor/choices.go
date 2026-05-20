// gameflow: 剑帝角色选择流。

package sword_emperor

import (
	"fmt"

	"starcup-engine/internal/engine/core/runtimeutil"
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

type choiceHandler struct{}

const (
	swordQiSlashFlowID     = "se_sword_qi_slash"
	swordQiSlashStepX      = "x"
	swordQiSlashStepTarget = "target"

	swordRainFlowID      = "se_sword_rain"
	swordRainStepTarget  = "target"
	swordRainStepDiscard = "discard"
)

var swordQiSlashFlowRuntime = model.MustNewPromptFlowRuntime(swordQiSlashFlowID, []model.PromptFlowStepSpec{
	{ID: swordQiSlashStepX, ChoiceType: "se_sword_qi_slash_x", CancelPolicy: model.CancelPolicyBack},
	{ID: swordQiSlashStepTarget, ChoiceType: "se_sword_qi_slash_target", CancelPolicy: model.CancelPolicyAbort},
})

var swordRainFlowRuntime = model.MustNewPromptFlowRuntime(swordRainFlowID, []model.PromptFlowStepSpec{
	{ID: swordRainStepTarget, ChoiceType: "se_sword_rain_target", CancelPolicy: model.CancelPolicyAbort},
	{ID: swordRainStepDiscard, ChoiceType: "se_sword_rain_discard", CancelPolicy: model.CancelPolicyBack},
})

func NewChoiceHandler() engineplayer.ChoiceHandler {
	return choiceHandler{}
}

func (choiceHandler) BuildPrompt(rt engineplayer.ChoiceRuntime, choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "se_sword_qi_slash_x":
		maxX := runtimeutil.ToIntContextValue(data["max_x"])
		if maxX < 1 {
			maxX = 1
		}
		options := make([]model.PromptOption, 0, maxX)
		for xValue := 1; xValue <= maxX; xValue++ {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", xValue), Label: fmt.Sprintf("移除%d点剑气，对另一名角色造成%d点法术伤害", xValue, xValue)})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【剑气斩】请选择X值：", Options: options, Min: 1, Max: 1, Presentation: &model.PromptPresentation{Kind: model.PresentationNumeric, NumericBase: 0}}
	case "se_sword_qi_slash_target":
		flow, err := model.RequirePromptFlow(data, swordQiSlashFlowID, "剑气斩")
		if err != nil {
			return nil
		}
		return engineplayer.BuildTargetChoicePrompt(rt, choiceType, playerID, fmt.Sprintf("【剑气斩】请选择承受%d点法术伤害的目标：", flow.Selection(swordQiSlashStepX).Count), data, false)
	case "se_sword_rain_target":
		targetIDs := runtimeutil.ParseStringSliceContextValue(data["target_ids"])
		options := make([]model.PromptOption, 0, len(targetIDs))
		for _, targetID := range targetIDs {
			if target := rt.GetPlayers()[targetID]; target != nil {
				options = append(options, model.PromptOption{
					ID:       targetID,
					Label:    target.Name,
					TargetID: targetID,
				})
			}
		}
		return &model.Prompt{
			Type:         model.PromptConfirm,
			PlayerID:     playerID,
			Message:      "【剑雨】请选择攻击目标：",
			Options:      options,
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationTargetPicker, TargetFilter: "custom"},
		}
	case "se_sword_rain_discard":
		if player == nil {
			return nil
		}
		indices := runtimeutil.ParseChoiceIntSlice(data["discard_indices"])
		options := make([]model.PromptOption, 0, len(indices))
		for _, idx := range indices {
			if idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, player.Hand[idx].Name),
			})
		}
		return &model.Prompt{
			Type:         model.PromptChooseCards,
			PlayerID:     playerID,
			Message:      "【剑雨】请选择要弃置的手牌：",
			Options:      options,
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationCardPicker, CardSource: "hand"},
		}
	default:
		return nil
	}
}

func (choiceHandler) HandleChoice(rt engineplayer.ChoiceRuntime, _ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "se_sword_qi_slash_x":
		return true, handleSwordEmperorSwordQiSlashXChoice(rt, selectionIndex, ctxData)
	case "se_sword_qi_slash_target":
		return true, handleSwordEmperorSwordQiSlashTargetChoice(rt, selectionIndex, ctxData)
	case "se_sword_rain_target":
		return true, handleSwordRainTarget(rt, selectionIndex, ctxData)
	case "se_sword_rain_discard":
		return true, handleSwordRainDiscard(rt, selectionIndex, ctxData)
	default:
		return false, nil
	}
}

func handleSwordEmperorSwordQiSlashXChoice(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
	if maxX <= 0 {
		return fmt.Errorf("剑气斩没有可选X值")
	}
	if selectionIndex < 0 || selectionIndex >= maxX {
		return fmt.Errorf("无效的X值选项: %d", selectionIndex)
	}
	xValue := selectionIndex + 1
	flow, err := model.RequirePromptFlow(ctxData, swordQiSlashFlowID, "剑气斩")
	if err != nil {
		return err
	}
	flow.PutSelection(swordQiSlashStepX, model.PromptFlowSelection{
		OptionIndexes: []int{selectionIndex},
		Count:         xValue,
	})
	return engineplayer.AdvancePromptFlowRuntimeChoice(rt, ctxData, swordQiSlashFlowRuntime, flow, swordQiSlashStepTarget)
}

func handleSwordEmperorSwordQiSlashTargetChoice(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
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
	flow, err := model.RequirePromptFlow(ctxData, swordQiSlashFlowID, "剑气斩")
	if err != nil {
		return err
	}
	xValue := flow.Selection(swordQiSlashStepX).Count
	if xValue <= 0 {
		return fmt.Errorf("剑气斩的X值无效")
	}
	rawCtx, _ := ctxData["user_ctx"].(*model.Context)
	if rawCtx != nil && rawCtx.EventCtx != nil && rawCtx.EventCtx.TargetID == targetID {
		return fmt.Errorf("剑气斩不能选择当前攻击目标")
	}
	flow.PutSelection(swordQiSlashStepTarget, model.PromptFlowSelection{
		OptionIndexes: []int{selectionIndex},
		TargetIDs:     []string{targetID},
	})

	nowQi := addSwordEmperorSwordQi(user, -xValue)
	rt.AddPendingDamage(model.PendingDamage{SourceID: user.ID, TargetID: targetID, Damage: xValue, DamageType: model.MagicAttack})
	rt.Log(fmt.Sprintf("%s 发动 [剑气斩]：移除%d点剑气（当前%d），对 %s 造成%d点法术伤害", user.Name, xValue, nowQi, target.Name, xValue))
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		rt.ResumePendingAttackHit(ctxData)
	}
	return nil
}

func handleSwordRainTarget(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
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

	// Proceed to discard phase
	discardIndices := engineplayer.AllHandIndices(user)
	if len(discardIndices) == 0 {
		// No cards to discard, proceed with attack
		performSwordRainAttack(rt, user, target, ctxData)
		rt.PopInterrupt()
		if rt.GetPendingInterrupt() == nil {
			rt.EnterActionExecutionStage()
		}
		return nil
	}

	flow := model.PromptFlowFromContext(ctxData)
	if flow == nil || flow.FlowID != swordRainFlowID {
		flow = swordRainFlowRuntime.MustBeginAt(swordRainStepTarget)
		model.SetPromptFlowContext(ctxData, flow)
	}
	flow.PutSelection(swordRainStepTarget, model.PromptFlowSelection{
		OptionIndexes: []int{selectionIndex},
		TargetIDs:     []string{targetID},
	})
	if err := swordRainFlowRuntime.MoveTo(flow, swordRainStepDiscard); err != nil {
		return err
	}
	ctxData["discard_indices"] = discardIndices
	ctxData["choice_type"] = "se_sword_rain_discard"
	engineplayer.NotifyChoiceContext(rt, ctxData)
	return nil
}

func handleSwordRainDiscard(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	discardIndices := runtimeutil.ParseChoiceIntSlice(ctxData["discard_indices"])
	cardIdx, ok := runtimeutil.ResolveSelectionToCandidate(selectionIndex, discardIndices)
	if !ok || cardIdx < 0 || cardIdx >= len(user.Hand) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}

	card := user.Hand[cardIdx]
	user.Hand = append(user.Hand[:cardIdx], user.Hand[cardIdx+1:]...)
	rt.NotifyCardRevealed(user.ID, []model.Card{card}, "discard")
	rt.AppendToDiscard([]model.Card{card})
	rt.Log(fmt.Sprintf("%s 发动 [剑雨]：弃置了1张手牌", user.Name))

	flow, err := model.RequirePromptFlow(ctxData, swordRainFlowID, "剑雨")
	if err != nil {
		return err
	}
	targetIDs := flow.Selection(swordRainStepTarget).TargetIDs
	if len(targetIDs) != 1 || targetIDs[0] == "" {
		return fmt.Errorf("剑雨缺少选定目标")
	}
	targetID := targetIDs[0]
	target := rt.GetPlayers()[targetID]
	if target == nil {
		return fmt.Errorf("目标不存在")
	}

	performSwordRainAttack(rt, user, target, ctxData)
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		rt.EnterActionExecutionStage()
	}
	return nil
}

func performSwordRainAttack(rt engineplayer.ChoiceRuntime, user *model.Player, target *model.Player, ctxData map[string]interface{}) {
	damage := 2
	rt.AddPendingDamage(model.PendingDamage{
		SourceID:   user.ID,
		TargetID:   target.ID,
		Damage:     damage,
		DamageType: model.AttackDamage,
	})
	rt.Log(fmt.Sprintf("%s 发动 [剑雨]：对 %s 造成%d点攻击伤害", user.Name, target.Name, damage))
}
