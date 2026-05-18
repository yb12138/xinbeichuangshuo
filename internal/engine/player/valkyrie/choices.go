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
			Presentation: &model.PromptPresentation{Kind: model.PresentationBranchSelect, Layout: "overlay"},
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
			Presentation: &model.PromptPresentation{Kind: model.PresentationNumeric, NumericBase: 0},
		}

	case "valkyrie_military_glory_target":
		targetIDs := runtimeutil.ParseStringSliceContextValue(data["target_ids"])
		options := make([]model.PromptOption, 0, len(targetIDs))
		for _, targetID := range targetIDs {
			if target := rt.GetPlayers()[targetID]; target != nil {
				options = append(options, model.PromptOption{
					ID:    targetID,
					Label: target.Name,
				})
			}
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【军威神光】请选择目标角色：",
			Options:  options,
			Min:      1,
			Max:      1,
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
				ID:    strconv.Itoa(idx),
				Label: fmt.Sprintf("%d: %s", idx+1, promptfmt.FormatCardInfo(player.Hand[idx])),
			})
		}
		options = append(options, model.PromptOption{ID: "cancel", Label: "放弃额外效果"})
		return &model.Prompt{
			Type:         model.PromptChooseCards,
			PlayerID:     playerID,
			Message:      "【英灵召唤】可额外弃1张法术牌并令当前战斗目标+1治疗（或点击取消放弃本次额外效果）：",
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
	case "valkyrie_heroic_discard_card":
		return true, handleHeroicDiscardCard(rt, selectionIndex, ctxData)
	default:
		return false, nil
	}
}

func (choiceHandler) HandleCancel(rt engineplayer.ChoiceRuntime, playerID string, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	if choiceType != "valkyrie_heroic_discard_card" {
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
	camp, _ := ctxData["camp"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	switch selectionIndex {
	case 0:
		rt.Heal(userID, 1)
		leaveHeroicForm(user)
		rt.Log(fmt.Sprintf("%s 选择军威神光选项1：+1治疗并脱离英灵形态", user.Name))
		rt.PopInterrupt()
		return nil
	case 1:
		maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
		if maxX <= 0 {
			return fmt.Errorf("当前阵营无可用能量")
		}
		if intr := rt.GetPendingInterrupt(); intr != nil {
			intr.Context = map[string]interface{}{
				"choice_type": "valkyrie_military_glory_x",
				"user_id":     userID,
				"camp":        camp,
				"max_x":       maxX,
			}
		}
		rt.NotifyInterruptPrompt()
		return nil
	default:
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
}

func handleMilitaryGloryX(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	camp, _ := ctxData["camp"].(string)
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
	if intr := rt.GetPendingInterrupt(); intr != nil {
		intr.Context = map[string]interface{}{
			"choice_type": "valkyrie_military_glory_target",
			"user_id":     userID,
			"camp":        camp,
			"x":           x,
			"target_ids":  targetIDs,
		}
	}
	rt.NotifyInterruptPrompt()
	return nil
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
	total := rt.GetCampCrystals(camp) + rt.GetCampGems(camp)
	if x > total {
		return fmt.Errorf("阵营能量不足")
	}
	useCrystal := x
	if crystals := rt.GetCampCrystals(camp); useCrystal > crystals {
		useCrystal = crystals
	}
	if useCrystal > 0 {
		rt.ModifyCrystal(camp, -useCrystal)
	}
	if remain := x - useCrystal; remain > 0 {
		rt.ModifyGem(camp, -remain)
	}
	rt.Heal(targetID, x)
	rt.Log(fmt.Sprintf("%s 选择军威神光选项2：移除%d星石并使 %s +%d治疗", user.Name, x, target.Name, x))
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

	rawCtx, _ := ctxData["user_ctx"].(*model.Context)
	targetID := ""
	if rawCtx != nil && rawCtx.EventCtx != nil {
		targetID = rawCtx.EventCtx.TargetID
	}
	if targetID == "" && rawCtx != nil && rawCtx.Target != nil {
		targetID = rawCtx.Target.ID
	}
	if targetID != "" {
		rt.Heal(targetID, 1)
		if target := rt.GetPlayers()[targetID]; target != nil {
			rt.Log(fmt.Sprintf("%s 因英灵召唤额外效果，获得1点治疗", target.Name))
		}
	}
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
