// gameflow: 女武神角色选择流。

package valkyrie

import (
	"fmt"
	"strconv"

	"starcup-engine/internal/engine/core/runtimeutil"
	"starcup-engine/internal/engine/hook/promptfmt"
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

type choiceHandler struct{}

const (
	militaryGloryFlowID     = "valkyrie_military_glory"
	militaryGloryStepMode   = "mode"
	militaryGloryStepX      = "x"
	militaryGloryStepTarget = "target"
	militaryGloryStepPay    = "pay"
)

// militaryGloryFlowRuntime 军威神光多步流：第一步可取消（放弃发动），后续每步可返回上一步。
var militaryGloryFlowRuntime = model.MustNewPromptFlowRuntime(militaryGloryFlowID, []model.PromptFlowStepSpec{
	{ID: militaryGloryStepMode, ChoiceType: "valkyrie_military_glory_mode", CancelPolicy: model.CancelPolicyAbort},
	{ID: militaryGloryStepX, ChoiceType: "valkyrie_military_glory_x", CancelPolicy: model.CancelPolicyBack},
	{ID: militaryGloryStepTarget, ChoiceType: "valkyrie_military_glory_target", CancelPolicy: model.CancelPolicyBack},
	{ID: militaryGloryStepPay, ChoiceType: "valkyrie_military_glory_pay", CancelPolicy: model.CancelPolicyBack},
})

// militaryGloryCancelPolicy 取当前步骤的取消策略，供前端显示"取消/返回"按钮；没有流程状态时不可取消。
func militaryGloryCancelPolicy(data map[string]interface{}) string {
	flow := model.PromptFlowFromContext(data)
	if flow == nil || flow.FlowID != militaryGloryFlowID {
		return ""
	}
	if step, ok := militaryGloryFlowRuntime.CurrentStep(flow); ok {
		if step.CancelPolicy == model.CancelPolicyBack && len(flow.History) == 0 {
			return model.CancelPolicyAbort
		}
		return step.CancelPolicy
	}
	return ""
}

// advanceMilitaryGlory 在原上下文上推进到下一步（保留前面步骤的数据，返回上一步时直接复用）。
func advanceMilitaryGlory(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, stepID string) error {
	flow := model.PromptFlowFromContext(ctxData)
	if flow == nil || flow.FlowID != militaryGloryFlowID {
		return fmt.Errorf("军威神光缺少多步流状态")
	}
	return engineplayer.AdvancePromptFlowRuntimeChoice(rt, ctxData, militaryGloryFlowRuntime, flow, stepID)
}

func NewChoiceHandler() engineplayer.ChoiceHandler {
	return choiceHandler{}
}

func (choiceHandler) BuildPrompt(rt engineplayer.ChoiceRuntime, choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "valkyrie_military_glory_mode":
		maxX := runtimeutil.ToIntContextValue(data["max_x"])
		options := []model.PromptOption{
			{ID: "0", Label: "你+1治疗并脱离英灵形态"},
		}
		if maxX > 0 {
			options = append(options, model.PromptOption{
				ID:    "1",
				Label: fmt.Sprintf("移除我方战绩区星石（1~%d）并指定角色+X治疗", maxX),
			})
		}
		return &model.Prompt{
			Type:         model.PromptConfirm,
			PlayerID:     playerID,
			Message:      "【军威神光】请选择效果：",
			Options:      options,
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationBranchSelect, Layout: "overlay", CancelPolicy: militaryGloryCancelPolicy(data)},
		}

	case "valkyrie_military_glory_x":
		maxX := runtimeutil.ToIntContextValue(data["max_x"])
		options := make([]model.PromptOption, 0, maxX)
		for x := 1; x <= maxX; x++ {
			options = append(options, model.PromptOption{
				ID:    strconv.Itoa(x),
				Label: fmt.Sprintf("X=%d", x),
			})
		}
		return &model.Prompt{
			Type:         model.PromptConfirm,
			PlayerID:     playerID,
			Message:      "【军威神光】请选择X：",
			Options:      options,
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationNumeric, NumericBase: 0, CancelPolicy: militaryGloryCancelPolicy(data)},
		}

	case "valkyrie_military_glory_target":
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
			Message:      "【军威神光】请选择目标角色：",
			Options:      options,
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationTargetPicker, TargetFilter: "custom", CancelPolicy: militaryGloryCancelPolicy(data)},
		}

	case "valkyrie_military_glory_pay":
		x := runtimeutil.ToIntContextValue(data["x"])
		gemCounts := runtimeutil.ParseChoiceIntSlice(data["gem_counts"])
		options := make([]model.PromptOption, 0, len(gemCounts))
		for _, gems := range gemCounts {
			options = append(options, model.PromptOption{
				ID:    strconv.Itoa(gems),
				Label: formatStonePayLabel(gems, x-gems),
			})
		}
		return &model.Prompt{
			Type:         model.PromptConfirm,
			PlayerID:     playerID,
			Message:      fmt.Sprintf("【军威神光】请选择移除的%d个星石：", x),
			Options:      options,
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationBranchSelect, Layout: "overlay", CancelPolicy: militaryGloryCancelPolicy(data)},
		}

	case "valkyrie_heroic_heal_target":
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
			Message:      "【英灵召唤】请选择目标角色+1治疗：",
			Options:      options,
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationTargetPicker, TargetFilter: "custom"},
		}

	case "valkyrie_heroic_discard_card":
		if player == nil {
			return nil
		}
		magicIndices := runtimeutil.ParseChoiceIntSlice(data["magic_indices"])
		options := make([]model.PromptOption, 0, len(magicIndices)+1)
		for _, idx := range magicIndices {
			if idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{
				ID:     strconv.Itoa(idx),
				Label:  fmt.Sprintf("%d: %s", idx+1, promptfmt.FormatCardInfo(player.Hand[idx])),
				CardID: player.Hand[idx].ID,
			})
		}
		options = append(options, model.PromptOption{ID: "cancel", Label: "放弃额外效果"})
		return &model.Prompt{
			Type:         model.PromptChooseCards,
			PlayerID:     playerID,
			Message:      "【英灵召唤】可额外弃1张法术牌并令目标角色+1治疗（或点击取消放弃本次额外效果）：",
			Options:      options,
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationCardPicker, CardSource: "hand", CardFilter: "magic_only", HasDecline: true},
		}
	default:
		return nil
	}
}

func (choiceHandler) HandleChoice(rt engineplayer.ChoiceRuntime, _ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "valkyrie_military_glory_mode":
		return true, handleMilitaryGloryMode(rt, selectionIndex, ctxData)
	case "valkyrie_military_glory_x":
		return true, handleMilitaryGloryX(rt, selectionIndex, ctxData)
	case "valkyrie_military_glory_target":
		return true, handleMilitaryGloryTarget(rt, selectionIndex, ctxData)
	case "valkyrie_military_glory_pay":
		return true, handleMilitaryGloryPay(rt, selectionIndex, ctxData)
	case "valkyrie_heroic_discard_card":
		return true, handleHeroicDiscardCard(rt, selectionIndex, ctxData)
	case "valkyrie_heroic_heal_target":
		return true, handleHeroicHealTarget(rt, selectionIndex, ctxData)
	default:
		return false, nil
	}
}

func (choiceHandler) HandleCancel(rt engineplayer.ChoiceRuntime, playerID string, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	if choiceType == "valkyrie_military_glory_mode" {
		// 第一步取消 = 本回合不发动军威神光（费用在最后一步才支付，无需退费）。
		rt.PopInterrupt()
		if user := rt.GetPlayers()[playerID]; user != nil {
			rt.Log(fmt.Sprintf("%s 放弃发动 [军威神光]", user.Name))
		}
		return true, nil
	}
	if choiceType != "valkyrie_heroic_discard_card" {
		// 军威神光后续步骤交给通用多步流处理（返回上一步）。
		return false, nil
	}
	rt.PopInterrupt()
	if user := rt.GetPlayers()[playerID]; user != nil {
		rt.Log(fmt.Sprintf("%s 放弃了 [英灵召唤] 的额外弃法术效果", user.Name))
	}
	rt.ResumePendingAttackHit(ctxData)
	return true, nil
}

func handleMilitaryGloryMode(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	switch selectionIndex {
	case 0:
		rt.Heal(userID, 1)
		leaveHeroicForm(user)
		rt.Log(fmt.Sprintf("%s 选择军威神光选项1：+1治疗并脱离英灵形态", user.Name))
		engineplayer.NotifySkillActivated(rt, user, user.ID, "valkyrie_military_glory", user.ID)
		rt.PopInterrupt()
		return nil
	case 1:
		maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
		if maxX <= 0 {
			return fmt.Errorf("当前阵营无可用能量")
		}
		return advanceMilitaryGlory(rt, ctxData, militaryGloryStepX)
	default:
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
}

func handleMilitaryGloryX(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
	if maxX <= 0 {
		return fmt.Errorf("当前阵营无可用能量")
	}
	x := selectionIndex + 1
	if x <= 0 || x > maxX || x >= 3 {
		return fmt.Errorf("无效的X值")
	}
	targetIDs := make([]string, 0, len(rt.GetPlayerOrder()))
	for _, pid := range rt.GetPlayerOrder() {
		if rt.GetPlayers()[pid] != nil {
			targetIDs = append(targetIDs, pid)
		}
	}
	ctxData["x"] = x
	ctxData["target_ids"] = targetIDs
	return advanceMilitaryGlory(rt, ctxData, militaryGloryStepTarget)
}

func handleMilitaryGloryTarget(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	camp, _ := ctxData["camp"].(string)
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
	x := runtimeutil.ToIntContextValue(ctxData["x"])
	if x <= 0 || x >= 3 {
		return fmt.Errorf("无效的X值")
	}
	gemCounts := militaryGloryGemOptions(rt.GetCampGems(camp), rt.GetCampCrystals(camp), x)
	if len(gemCounts) == 0 {
		return fmt.Errorf("阵营能量不足")
	}
	if len(gemCounts) == 1 {
		return finishMilitaryGlory(rt, user, camp, targetID, x, gemCounts[0])
	}
	// 宝石/水晶组合不唯一时，由玩家选择移除哪种星石。
	ctxData["target_id"] = targetID
	ctxData["gem_counts"] = gemCounts
	return advanceMilitaryGlory(rt, ctxData, militaryGloryStepPay)
}

func handleMilitaryGloryPay(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	camp, _ := ctxData["camp"].(string)
	targetID, _ := ctxData["target_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	x := runtimeutil.ToIntContextValue(ctxData["x"])
	gemCounts := runtimeutil.ParseChoiceIntSlice(ctxData["gem_counts"])
	if selectionIndex < 0 || selectionIndex >= len(gemCounts) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	return finishMilitaryGlory(rt, user, camp, targetID, x, gemCounts[selectionIndex])
}

// militaryGloryGemOptions 返回移除 x 个星石时所有可行的宝石数量（其余用水晶）。
func militaryGloryGemOptions(gems, crystals, x int) []int {
	out := make([]int, 0, x+1)
	for g := 0; g <= x; g++ {
		if g <= gems && x-g <= crystals {
			out = append(out, g)
		}
	}
	return out
}

func formatStonePayLabel(gems, crystals int) string {
	switch {
	case gems > 0 && crystals > 0:
		return fmt.Sprintf("移除%d宝石+%d水晶", gems, crystals)
	case gems > 0:
		return fmt.Sprintf("移除%d宝石", gems)
	default:
		return fmt.Sprintf("移除%d水晶", crystals)
	}
}

func finishMilitaryGlory(rt engineplayer.ChoiceRuntime, user *model.Player, camp, targetID string, x, gems int) error {
	target := rt.GetPlayers()[targetID]
	if target == nil {
		return fmt.Errorf("目标不存在")
	}
	crystals := x - gems
	if gems < 0 || crystals < 0 || gems > rt.GetCampGems(camp) || crystals > rt.GetCampCrystals(camp) {
		return fmt.Errorf("阵营能量不足")
	}
	if gems > 0 {
		rt.ModifyGem(camp, -gems)
	}
	if crystals > 0 {
		rt.ModifyCrystal(camp, -crystals)
	}
	rt.Heal(targetID, x)
	rt.Log(fmt.Sprintf("%s 选择军威神光选项2：移除%d星石（%d宝石/%d水晶）并使 %s +%d治疗", user.Name, x, gems, crystals, target.Name, x))
	engineplayer.NotifySkillActivated(rt, user, user.ID, "valkyrie_military_glory", target.ID)
	rt.PopInterrupt()
	return nil
}

func handleHeroicDiscardCard(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	magicIndices := runtimeutil.ParseChoiceIntSlice(ctxData["magic_indices"])
	cardIdx, ok := runtimeutil.ResolveSelectionToCandidate(selectionIndex, magicIndices)
	if !ok {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	if cardIdx < 0 || cardIdx >= len(user.Hand) || user.Hand[cardIdx].Type != model.CardTypeMagic {
		return fmt.Errorf("请选择法术牌")
	}
	card := user.Hand[cardIdx]
	rt.NotifyCardRevealed(userID, []model.Card{card}, "discard")
	user.Hand = append(user.Hand[:cardIdx], user.Hand[cardIdx+1:]...)
	rt.AppendToDiscard([]model.Card{card})

	// 额外效果的"目标角色"由玩家任选（包括自己、队友）。
	targetIDs := make([]string, 0, len(rt.GetPlayerOrder()))
	for _, pid := range rt.GetPlayerOrder() {
		if rt.GetPlayers()[pid] != nil {
			targetIDs = append(targetIDs, pid)
		}
	}
	if intr := rt.GetPendingInterrupt(); intr != nil {
		intr.Context = map[string]interface{}{
			"choice_type": "valkyrie_heroic_heal_target",
			"user_id":     userID,
			"user_ctx":    ctxData["user_ctx"],
			"target_ids":  targetIDs,
		}
	}
	rt.NotifyInterruptPrompt()
	return nil
}

func handleHeroicHealTarget(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
	if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	targetID := targetIDs[selectionIndex]
	target := rt.GetPlayers()[targetID]
	if target == nil {
		return fmt.Errorf("目标不存在")
	}
	rt.Heal(targetID, 1)
	rt.Log(fmt.Sprintf("%s 因英灵召唤额外效果，获得1点治疗", target.Name))
	rt.PopInterrupt()
	rt.ResumePendingAttackHit(ctxData)
	return nil
}

func leaveHeroicForm(player *model.Player) {
	if player == nil {
		return
	}
	if player.Form != "" && player.Form != model.FormValkyrieHeroic {
		return
	}
	player.Orientation = model.OrientationNormal
	player.Form = ""
}

var _ engineplayer.CancelChoiceHandler = choiceHandler{}
