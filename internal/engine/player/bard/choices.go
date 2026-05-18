// gameflow: 吟游诗人角色选择流。

package bard

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

func (choiceHandler) BuildPrompt(rt engineplayer.ChoiceRuntime, choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "bd_descent_element":
		elemCounts := getSameElementCounts(player)
		elems := make([]model.Element, 0)
		for _, ele := range engineplayer.ElementOrderForPrompt() {
			if elemCounts[ele] >= 2 {
				elems = append(elems, ele)
			}
		}
		options := make([]model.PromptOption, 0, len(elems))
		for _, ele := range elems {
			options = append(options, model.PromptOption{ID: string(ele), Label: fmt.Sprintf("%s系", promptfmt.ElementName(string(ele)))})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【沉沦协奏曲】请选择要弃置的同系元素：", Options: options, Min: 1, Max: 1, Presentation: &model.PromptPresentation{Kind: model.PresentationBranchSelect, Layout: "overlay"}}
	case "bd_descent_cards":
		// 直接展示所有同系牌候选，无需前置元素选择步骤
		remaining := engineplayer.ParseIntSliceContextValue(data["remaining_indices"])
		selected := len(engineplayer.ParseIntSliceContextValue(data["selected_indices"]))
		options := make([]model.PromptOption, 0, len(remaining))
		for _, idx := range remaining {
			if idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", idx), Label: fmt.Sprintf("%d: %s", idx+1, promptfmt.FormatCardInfo(player.Hand[idx]))})
		}
		remainingPick := 2 - selected
		if remainingPick < 1 {
			remainingPick = 1
		}
		if len(options) > 0 && remainingPick > len(options) {
			remainingPick = len(options)
		}
		return &model.Prompt{Type: model.PromptChooseCards, PlayerID: playerID, Message: fmt.Sprintf("【沉沦协奏曲】请选择要弃置的%d张同系牌：", remainingPick), Options: options, Min: remainingPick, Max: remainingPick, Presentation: &model.PromptPresentation{Kind: model.PresentationCardPicker, CardSource: "hand"}}
	case "bd_dissonance_x":
		maxX := runtimeutil.ToIntContextValue(data["max_x"])
		if maxX < 2 {
			maxX = 2
		}
		options := make([]model.PromptOption, 0, maxX-1)
		for x := 2; x <= maxX; x++ {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", x), Label: fmt.Sprintf("X=%d", x)})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【不谐和弦】请选择X值：", Options: options, Min: 1, Max: 1, Presentation: &model.PromptPresentation{Kind: model.PresentationNumeric, NumericBase: 0}}
	case "bd_dissonance_mode":
		xValue := runtimeutil.ToIntContextValue(data["x_value"])
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: fmt.Sprintf("【不谐和弦】请选择分支（X=%d）：", xValue), Options: []model.PromptOption{{ID: "0", Label: fmt.Sprintf("你与目标各摸%d张牌", xValue-1)}, {ID: "1", Label: fmt.Sprintf("你与目标各弃%d张牌", xValue-1)}}, Min: 1, Max: 1, Presentation: &model.PromptPresentation{Kind: model.PresentationBranchSelect, Layout: "overlay"}}
	case "bd_dissonance_discard_step":
		currentActorID, _ := data["current_actor_id"].(string)
		actor := rt.GetPlayers()[currentActorID]
		if actor == nil {
			return nil
		}
		need := runtimeutil.ToIntContextValue(data["need_count"])
		selected := runtimeutil.ToIntContextValue(data["selected_count"])
		options := make([]model.PromptOption, 0, len(actor.Hand))
		for idx, c := range actor.Hand {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", idx), Label: fmt.Sprintf("%d: %s", idx+1, promptfmt.FormatCardInfo(c))})
		}
		remainingPick := need - selected
		if remainingPick < 1 {
			remainingPick = 1
		}
		if len(options) > 0 && remainingPick > len(options) {
			remainingPick = len(options)
		}
		return &model.Prompt{Type: model.PromptChooseCards, PlayerID: playerID, Message: fmt.Sprintf("【不谐和弦】请 %s 选择要弃置的%d张手牌：", actor.Name, remainingPick), Options: options, Min: remainingPick, Max: remainingPick, Presentation: &model.PromptPresentation{Kind: model.PresentationCardPicker, CardSource: "hand"}}
	case "bd_rousing_mode":
		opts := []model.PromptOption{
			{ID: "0", Label: "对2名对手各造成1点法术伤害"},
			{ID: "1", Label: "弃2张牌"},
			{ID: "2", Label: "跳过"},
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【激昂狂想曲】请选择效果：", Options: opts, Min: 1, Max: 1, Presentation: &model.PromptPresentation{Kind: model.PresentationBranchSelect, Layout: "overlay"}}
	case "bd_rousing_targets":
		targetIDs := runtimeutil.ParseStringSliceContextValue(data["target_ids"])
		selectedSet := runtimeutil.IDsToSet(runtimeutil.ParseStringSliceContextValue(data["selected_target_ids"]))
		options := make([]model.PromptOption, 0, len(targetIDs))
		for _, targetID := range targetIDs {
			if selectedSet[targetID] {
				continue
			}
			if target := rt.GetPlayers()[targetID]; target != nil {
				options = append(options, model.PromptOption{ID: targetID, Label: target.Name})
			}
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: fmt.Sprintf("【激昂狂想曲】请选择第 %d/2 名目标：", len(selectedSet)+1), Options: options, Min: 1, Max: 1, Presentation: &model.PromptPresentation{Kind: model.PresentationTargetPicker, TargetFilter: "custom"}}
	case "bd_rousing_discard_cards":
		selected := len(engineplayer.ParseIntSliceContextValue(data["selected_indices"]))
		remaining := engineplayer.ParseIntSliceContextValue(data["remaining_indices"])
		options := make([]model.PromptOption, 0, len(remaining))
		for _, idx := range remaining {
			if idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", idx), Label: fmt.Sprintf("%d: %s", idx+1, promptfmt.FormatCardInfo(player.Hand[idx]))})
		}
		remainingPick := 2 - selected
		if remainingPick < 1 {
			remainingPick = 1
		}
		if len(options) > 0 && remainingPick > len(options) {
			remainingPick = len(options)
		}
		return &model.Prompt{Type: model.PromptChooseCards, PlayerID: playerID, Message: fmt.Sprintf("【激昂狂想曲】请选择要弃置的%d张手牌：", remainingPick), Options: options, Min: remainingPick, Max: remainingPick, Presentation: &model.PromptPresentation{Kind: model.PresentationCardPicker, CardSource: "hand"}}
	case "bd_victory_confirm":
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【胜利交响诗】请选择效果：",
			Options: []model.PromptOption{
				{ID: "0", Label: "将我方战绩区1个星石提炼为你的能量"},
				{ID: "1", Label: "我方战绩区+1宝石，你+1治疗"},
				{ID: "2", Label: "取消"},
			},
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationBranchSelect, Layout: "overlay"},
		}
	case "bd_victory_mode":
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【胜利交响诗】请选择效果：", Options: []model.PromptOption{{ID: "0", Label: "将我方战绩区1个星石提炼为你的能量"}, {ID: "1", Label: "我方战绩区+1宝石，你+1治疗"}}, Min: 1, Max: 1, Presentation: &model.PromptPresentation{Kind: model.PresentationBranchSelect, Layout: "overlay"}}
	case "bd_victory_extract_stone":
		options := make([]model.PromptOption, 0, 2)
		if player != nil {
			camp := string(player.Camp)
			if rt.GetCampGems(camp) > 0 {
				options = append(options, model.PromptOption{ID: "0", Label: "提炼1个宝石"})
			}
			if rt.GetCampCrystals(camp) > 0 {
				options = append(options, model.PromptOption{ID: "1", Label: "提炼1个水晶"})
			}
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【胜利交响诗】请选择要提炼的星石：", Options: options, Min: 1, Max: 1, Presentation: &model.PromptPresentation{Kind: model.PresentationBranchSelect, Layout: "overlay"}}
	case "bd_hope_draw_confirm":
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【希望赋格曲】是否先摸1张牌？", Options: []model.PromptOption{{ID: "0", Label: "是"}, {ID: "1", Label: "否"}}, Min: 1, Max: 1, Presentation: &model.PromptPresentation{Kind: model.PresentationBranchSelect, Layout: "overlay"}}
	case "bd_hope_mode":
		opts := []model.PromptOption{{ID: "0", Label: "将永恒乐章放置于目标队友面前"}}
		if EternalHolderID(rt, player) != "" {
			opts = append(opts,
				model.PromptOption{ID: "1", Label: "转移永恒乐章，弃1张牌并+1治疗"},
				model.PromptOption{ID: "2", Label: "转移永恒乐章，弃1张牌并+1灵感"},
			)
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【希望赋格曲】请选择分支：", Options: opts, Min: 1, Max: 1, Presentation: &model.PromptPresentation{Kind: model.PresentationBranchSelect, Layout: "overlay"}}
	case "bd_hope_transfer_discard":
		options := make([]model.PromptOption, 0, len(player.Hand))
		for idx, c := range player.Hand {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", idx), Label: fmt.Sprintf("%d: %s", idx+1, promptfmt.FormatCardInfo(c))})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【希望赋格曲】请选择弃置1张手牌：", Options: options, Min: 1, Max: 1, Presentation: &model.PromptPresentation{Kind: model.PresentationCardPicker, CardSource: "hand"}}
	case "bd_descent_target":
		return engineplayer.BuildTargetChoicePrompt(rt, choiceType, playerID, "【沉沦协奏曲】请选择1点法术伤害目标：", data, false)
	case "bd_dissonance_target":
		return engineplayer.BuildTargetChoicePrompt(rt, choiceType, playerID, "【不谐和弦】请选择目标角色：", data, false)
	case "bd_hope_place_target":
		return engineplayer.BuildTargetChoicePrompt(rt, choiceType, playerID, "【希望赋格曲】请选择放置永恒乐章的目标队友：", data, false)
	case "bd_hope_transfer_target":
		return engineplayer.BuildTargetChoicePrompt(rt, choiceType, playerID, "【希望赋格曲】请选择转移永恒乐章的目标队友：", data, false)
	default:
		return nil
	}
}

func (choiceHandler) HandleChoice(rt engineplayer.ChoiceRuntime, _ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "bd_descent_element":
		return true, handleDescentElement(rt, ctxData, selectionIndex)
	case "bd_descent_cards":
		return true, handleDescentCards(rt, ctxData, selectionIndex)
	case "bd_descent_target":
		return true, handleDescentTarget(rt, ctxData, selectionIndex)
	case "bd_dissonance_x":
		return true, handleDissonanceX(rt, ctxData, selectionIndex)
	case "bd_dissonance_mode":
		return true, handleDissonanceMode(rt, ctxData, selectionIndex)
	case "bd_dissonance_target":
		return true, handleDissonanceTarget(rt, ctxData, selectionIndex)
	case "bd_dissonance_discard_step":
		return true, handleDissonanceDiscardStep(rt, ctxData, selectionIndex)
	case "bd_rousing_mode":
		return true, handleRousingMode(rt, ctxData, selectionIndex)
	case "bd_rousing_targets":
		return true, handleRousingTargets(rt, ctxData, selectionIndex)
	case "bd_rousing_discard_cards":
		return true, handleRousingDiscardCards(rt, ctxData, selectionIndex)
	case "bd_victory_confirm":
		return true, handleVictoryConfirm(rt, ctxData, selectionIndex)
	case "bd_victory_mode":
		return true, handleVictoryMode(rt, ctxData, selectionIndex)
	case "bd_victory_extract_stone":
		return true, handleVictoryExtractStone(rt, ctxData, selectionIndex)
	case "bd_hope_draw_confirm":
		return true, handleHopeDrawConfirm(rt, ctxData, selectionIndex)
	case "bd_hope_mode":
		return true, handleHopeMode(rt, ctxData, selectionIndex)
	case "bd_hope_place_target":
		return true, handleHopePlaceTarget(rt, ctxData, selectionIndex)
	case "bd_hope_transfer_target":
		return true, handleHopeTransferTarget(rt, ctxData, selectionIndex)
	case "bd_hope_transfer_discard":
		return true, handleHopeTransferDiscard(rt, ctxData, selectionIndex)
	default:
		return false, nil
	}
}

func (choiceHandler) HandleCancel(rt engineplayer.ChoiceRuntime, _ string, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	if choiceType != "bd_victory_confirm" {
		return false, nil
	}
	return true, cancelVictorySymphony(rt, ctxData)
}

// ---- 沉沦协奏曲 ----

func handleDescentElement(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	elemCounts := getSameElementCounts(user)
	elems := make([]model.Element, 0)
	for _, ele := range engineplayer.ElementOrderForPrompt() {
		if elemCounts[ele] >= 2 {
			elems = append(elems, ele)
		}
	}
	if selectionIndex < 0 || selectionIndex >= len(elems) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	chosen := elems[selectionIndex]
	ctxData["chosen_element"] = string(chosen)
	ctxData["selected_indices"] = []int{}
	ctxData["remaining_indices"] = engineplayer.GetCardIndicesByElement(user, chosen)
	ctxData["choice_type"] = "bd_descent_cards"
	intr := rt.GetPendingInterrupt()
	if intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
	return nil
}

func handleDescentCards(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	remaining := engineplayer.ParseIntSliceContextValue(ctxData["remaining_indices"])
	selected := engineplayer.ParseIntSliceContextValue(ctxData["selected_indices"])
	cardIdx, ok := runtimeutil.ResolveSelectionToCandidate(selectionIndex, remaining)
	if !ok || cardIdx < 0 || cardIdx >= len(user.Hand) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	selected = append(selected, cardIdx)
	nextRemaining := make([]int, 0, len(remaining))
	for _, idx := range remaining {
		if idx != cardIdx {
			nextRemaining = append(nextRemaining, idx)
		}
	}
	if len(selected) < 2 {
		ctxData["selected_indices"] = selected
		ctxData["remaining_indices"] = nextRemaining
		intr := rt.GetPendingInterrupt()
		if intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return nil
	}
	// 验证所选2张牌是否同系
	if len(selected) >= 2 && user.Hand[selected[0]].Element != user.Hand[selected[1]].Element {
		return fmt.Errorf("沉沦协奏曲需弃置同系牌")
	}
	removed, err := engineplayer.RemoveCardsByIndicesFromHand(user, append([]int{}, selected...))
	if err != nil {
		return err
	}
	rt.NotifyCardRevealed(user.ID, removed, "discard")
	rt.AppendToDiscard(removed)
	user.TurnState.UsedSkillCounts["bd_descent"] = 1
	now := addBardInspiration(user, 1)
	chosenEle := string(removed[0].Element)
	rt.Log(fmt.Sprintf("%s 发动 [沉沦协奏曲]：弃2张%s系牌，灵感+1（当前%d）", user.Name, chosenEle, now))

	hasMagic := false
	for _, card := range removed {
		if card.Type == model.CardTypeMagic {
			hasMagic = true
			break
		}
	}
	if hasMagic {
		ctxData["choice_type"] = "bd_descent_target"
		ctxData["target_ids"] = campEnemyIDs(rt, user)
		intr := rt.GetPendingInterrupt()
		if intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return nil
	}
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil && len(rt.GetPendingDamageQueue()) > 0 {
		rt.EnterDamageResolution(nil)
	}
	return nil
}

func handleDescentTarget(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
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
	rt.AddPendingDamage(model.PendingDamage{SourceID: user.ID, TargetID: target.ID, Damage: 1, DamageType: model.MagicAttack})
	rt.Log(fmt.Sprintf("%s 的 [沉沦协奏曲] 追加效果：对 %s 造成1点法术伤害", user.Name, target.Name))
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		rt.EnterDamageResolution(nil)
	}
	return nil
}

// ---- 不谐和弦 ----

func handleDissonanceX(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
	xValue := selectionIndex + 2
	if xValue < 2 || xValue > maxX {
		return fmt.Errorf("无效的X值")
	}
	if bardInspiration(user) < xValue {
		return fmt.Errorf("灵感不足")
	}
	addBardInspiration(user, -xValue)
	if InEternalPrisonerForm(user) {
		LeaveEternalPrisonerForm(user)
		rt.Log(fmt.Sprintf("%s 发动 [不谐和弦]：脱离永恒囚徒形态", user.Name))
	}
	ctxData["x_value"] = xValue
	ctxData["choice_type"] = "bd_dissonance_mode"
	intr := rt.GetPendingInterrupt()
	if intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
	return nil
}

func handleDissonanceMode(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	if selectionIndex != 0 && selectionIndex != 1 {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	ctxData["mode"] = selectionIndex
	ctxData["choice_type"] = "bd_dissonance_target"
	ctxData["target_ids"] = append([]string{}, rt.GetPlayerOrder()...)
	intr := rt.GetPendingInterrupt()
	if intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
	return nil
}

func handleDissonanceTarget(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
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
	xValue := runtimeutil.ToIntContextValue(ctxData["x_value"])
	mode := runtimeutil.ToIntContextValue(ctxData["mode"])
	n := xValue - 1
	if n < 0 {
		n = 0
	}
	if mode == 0 {
		if n > 0 {
			rt.DrawCards(user.ID, n)
			rt.DrawCards(target.ID, n)
		}
		rt.Log(fmt.Sprintf("%s 发动 [不谐和弦]：与 %s 各摸%d张牌", user.Name, target.Name, n))
		rt.PopInterrupt()
		if rt.GetPendingInterrupt() == nil {
			rt.RoutePendingDamageOr(model.TurnStageExtraAction, func() {
				rt.EnterExtraActionStage()
			})
		}
		return nil
	}
	// mode == 1: 弃牌分支
	actors := []string{user.ID, target.ID}
	startCursor := 0
	for startCursor < len(actors) {
		actor := rt.GetPlayers()[actors[startCursor]]
		if actor != nil && len(actor.Hand) > 0 && n > 0 {
			break
		}
		startCursor++
	}
	if n <= 0 || startCursor >= len(actors) {
		rt.Log(fmt.Sprintf("%s 发动 [不谐和弦]：弃牌分支无可执行弃牌", user.Name))
		rt.PopInterrupt()
		if rt.GetPendingInterrupt() == nil {
			rt.EnterExtraActionStage()
		}
		return nil
	}
	currentActor := rt.GetPlayers()[actors[startCursor]]
	ctxData["choice_type"] = "bd_dissonance_discard_step"
	ctxData["actor_ids"] = actors
	ctxData["cursor"] = startCursor
	ctxData["current_actor_id"] = currentActor.ID
	ctxData["need_count"] = n
	ctxData["selected_count"] = 0
	ctxData["selected_indices"] = []int{}
	ctxData["remaining_indices"] = engineplayer.AllHandIndices(currentActor)
	intr := rt.GetPendingInterrupt()
	if intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
	return nil
}

func handleDissonanceDiscardStep(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	actorIDs := runtimeutil.ParseStringSliceContextValue(ctxData["actor_ids"])
	cursor := runtimeutil.ToIntContextValue(ctxData["cursor"])
	if cursor < 0 || cursor >= len(actorIDs) {
		return fmt.Errorf("弃牌游标无效")
	}
	currentActorID, _ := ctxData["current_actor_id"].(string)
	if currentActorID == "" {
		currentActorID = actorIDs[cursor]
	}
	actor := rt.GetPlayers()[currentActorID]
	if actor == nil {
		return fmt.Errorf("弃牌角色不存在")
	}
	needCount := runtimeutil.ToIntContextValue(ctxData["need_count"])
	selectedCount := runtimeutil.ToIntContextValue(ctxData["selected_count"])
	remaining := engineplayer.ParseIntSliceContextValue(ctxData["remaining_indices"])
	selected := engineplayer.ParseIntSliceContextValue(ctxData["selected_indices"])
	cardIdx, ok := runtimeutil.ResolveSelectionToCandidate(selectionIndex, remaining)
	if !ok || cardIdx < 0 || cardIdx >= len(actor.Hand) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	selected = append(selected, cardIdx)
	selectedCount++
	nextRemaining := make([]int, 0, len(remaining))
	for _, idx := range remaining {
		if idx != cardIdx {
			nextRemaining = append(nextRemaining, idx)
		}
	}
	if selectedCount < needCount && len(nextRemaining) > 0 {
		ctxData["selected_count"] = selectedCount
		ctxData["selected_indices"] = selected
		ctxData["remaining_indices"] = nextRemaining
		intr := rt.GetPendingInterrupt()
		if intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return nil
	}
	removed, err := engineplayer.RemoveCardsByIndicesFromHand(actor, append([]int{}, selected...))
	if err != nil {
		return err
	}
	rt.NotifyCardRevealed(actor.ID, removed, "discard")
	rt.AppendToDiscard(removed)
	rt.Log(fmt.Sprintf("%s 的 [不谐和弦]：%s 弃置了%d张手牌", user.Name, actor.Name, len(removed)))

	nextCursor := cursor + 1
	for nextCursor < len(actorIDs) {
		nextActor := rt.GetPlayers()[actorIDs[nextCursor]]
		if nextActor == nil || len(nextActor.Hand) == 0 {
			nextCursor++
			continue
		}
		ctxData["cursor"] = nextCursor
		ctxData["current_actor_id"] = nextActor.ID
		ctxData["selected_count"] = 0
		ctxData["selected_indices"] = []int{}
		ctxData["remaining_indices"] = engineplayer.AllHandIndices(nextActor)
		intr := rt.GetPendingInterrupt()
		if intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return nil
	}
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		rt.RoutePendingDamageOr(model.TurnStageExtraAction, func() {
			rt.EnterExtraActionStage()
		})
	}
	return nil
}

// ---- 激昂狂想曲 ----

func handleRousingMode(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	holderID, _ := ctxData["user_id"].(string)
	holder := rt.GetPlayers()[holderID]
	if holder == nil {
		return fmt.Errorf("永恒乐章持有者不存在")
	}
	switch selectionIndex {
	case 0:
		bardID, _ := ctxData["bard_id"].(string)
		ctxData["choice_type"] = "bd_rousing_targets"
		ctxData["selected_target_ids"] = []string{}
		intr := rt.GetPendingInterrupt()
		if intr != nil {
			intr.Context = ctxData
			// 目标选择由吟游诗人执行（伤害来源是吟游诗人）
			intr.PlayerID = bardID
		}
		rt.NotifyInterruptPrompt()
		return nil
	case 1:
		if len(holder.Hand) < 2 {
			return fmt.Errorf("手牌不足2张，无法执行弃2张牌分支")
		}
		ctxData["choice_type"] = "bd_rousing_discard_cards"
		ctxData["selected_indices"] = []int{}
		ctxData["remaining_indices"] = engineplayer.AllHandIndices(holder)
		intr := rt.GetPendingInterrupt()
		if intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return nil
	case 2: // 跳过
		rt.Log(fmt.Sprintf("%s 选择跳过 [激昂狂想曲]", holder.Name))
		rt.PopInterrupt()
		if rt.GetPendingInterrupt() == nil {
			if !rt.RoutePendingDamageWithReturn(model.TurnStageActionStart) {
				rt.ApplyChoiceResumePoint(model.TurnStageActionStart)
			}
		}
		return nil
	default:
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
}

func handleRousingTargets(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	bardID, _ := ctxData["bard_id"].(string)
	bard := rt.GetPlayers()[bardID]
	holderID, _ := ctxData["user_id"].(string)
	holder := rt.GetPlayers()[holderID]
	if bard == nil || holder == nil {
		return fmt.Errorf("吟游诗人或持有者不存在")
	}
	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
	selected := runtimeutil.DedupeIDs(runtimeutil.ParseStringSliceContextValue(ctxData["selected_target_ids"]))
	selectedSet := runtimeutil.IDsToSet(selected)
	remaining := make([]string, 0, len(targetIDs))
	for _, targetID := range targetIDs {
		if !selectedSet[targetID] {
			remaining = append(remaining, targetID)
		}
	}
	if selectionIndex < 0 || selectionIndex >= len(remaining) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	selected = append(selected, remaining[selectionIndex])
	if len(selected) < 2 {
		ctxData["selected_target_ids"] = selected
		intr := rt.GetPendingInterrupt()
		if intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return nil
	}
	for _, targetID := range selected {
		rt.AddPendingDamage(model.PendingDamage{SourceID: bard.ID, TargetID: targetID, Damage: 1, DamageType: model.MagicAttack})
	}
	rt.Log(fmt.Sprintf("%s 发动 [激昂狂想曲]（%s 触发）：对2名目标各造成1点法术伤害", bard.Name, holder.Name))
	resolveBardForbiddenVerseAfterSong(rt, bard, "激昂狂想曲")
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		if !rt.RoutePendingDamageWithReturn(model.TurnStageActionStart) {
			rt.ApplyChoiceResumePoint(model.TurnStageActionStart)
		}
	}
	return nil
}

func handleRousingDiscardCards(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	bardID, _ := ctxData["bard_id"].(string)
	bard := rt.GetPlayers()[bardID]
	holderID, _ := ctxData["user_id"].(string)
	holder := rt.GetPlayers()[holderID]
	if bard == nil || holder == nil {
		return fmt.Errorf("吟游诗人或持有者不存在")
	}
	remaining := engineplayer.ParseIntSliceContextValue(ctxData["remaining_indices"])
	selected := engineplayer.ParseIntSliceContextValue(ctxData["selected_indices"])
	cardIdx, ok := runtimeutil.ResolveSelectionToCandidate(selectionIndex, remaining)
	if !ok || cardIdx < 0 || cardIdx >= len(holder.Hand) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	selected = append(selected, cardIdx)
	nextRemaining := make([]int, 0, len(remaining))
	for _, idx := range remaining {
		if idx != cardIdx {
			nextRemaining = append(nextRemaining, idx)
		}
	}
	if len(selected) < 2 {
		ctxData["selected_indices"] = selected
		ctxData["remaining_indices"] = nextRemaining
		intr := rt.GetPendingInterrupt()
		if intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return nil
	}
	removed, err := engineplayer.RemoveCardsByIndicesFromHand(holder, append([]int{}, selected...))
	if err != nil {
		return err
	}
	rt.NotifyCardRevealed(holder.ID, removed, "discard")
	rt.AppendToDiscard(removed)
	rt.Log(fmt.Sprintf("%s 发动 [激昂狂想曲]（%s 触发）：选择弃2张牌", bard.Name, holder.Name))
	resolveBardForbiddenVerseAfterSong(rt, bard, "激昂狂想曲")
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		if !rt.RoutePendingDamageWithReturn(model.TurnStageActionStart) {
			rt.ApplyChoiceResumePoint(model.TurnStageActionStart)
		}
	}
	return nil
}

// ---- 胜利交响诗 ----

func handleVictoryConfirm(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	bardID, _ := ctxData["bard_id"].(string)
	bard := rt.GetPlayers()[bardID]
	holderID, _ := ctxData["user_id"].(string)
	holder := rt.GetPlayers()[holderID]
	if bard == nil || holder == nil {
		return fmt.Errorf("吟游诗人或持有者不存在")
	}
	switch selectionIndex {
	case 0, 1: // 选择任一分支即视为发动
		return handleVictoryMode(rt, ctxData, selectionIndex)
	case 2: // 不发动
		return cancelVictorySymphony(rt, ctxData)
	default:
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
}

func cancelVictorySymphony(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}) error {
	holderID, _ := ctxData["user_id"].(string)
	holder := rt.GetPlayers()[holderID]
	if holder == nil {
		return fmt.Errorf("永恒乐章持有者不存在")
	}
	rt.Log(fmt.Sprintf("%s 选择不发动 [胜利交响诗]", holder.Name))
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		rt.RoutePendingDamageOr(model.TurnStageTurnEnd, func() {
			rt.EnterTurnEndStage()
		})
	}
	return nil
}

func handleVictoryMode(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	bardID, _ := ctxData["bard_id"].(string)
	bard := rt.GetPlayers()[bardID]
	holderID, _ := ctxData["user_id"].(string)
	holder := rt.GetPlayers()[holderID]
	if bard == nil || holder == nil {
		return fmt.Errorf("吟游诗人或持有者不存在")
	}
	switch selectionIndex {
	case 0:
		camp := string(holder.Camp)
		if rt.GetCampGems(camp)+rt.GetCampCrystals(camp) <= 0 {
			return fmt.Errorf("我方战绩区没有可提炼的星石")
		}
		ctxData["choice_type"] = "bd_victory_extract_stone"
		intr := rt.GetPendingInterrupt()
		if intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return nil
	case 1:
		camp := string(holder.Camp)
		addCampResource(rt, camp, "gem")
		rt.Heal(holder.ID, 1)
		rt.Log(fmt.Sprintf("%s 发动 [胜利交响诗]（%s 触发）：我方战绩区+1宝石，%s+1治疗", bard.Name, holder.Name, holder.Name))
	default:
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	resolveBardForbiddenVerseAfterSong(rt, bard, "胜利交响诗")
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		rt.RoutePendingDamageOr(model.TurnStageTurnEnd, func() {
			rt.EnterTurnEndStage()
		})
	}
	return nil
}

func handleVictoryExtractStone(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	bardID, _ := ctxData["bard_id"].(string)
	bard := rt.GetPlayers()[bardID]
	holderID, _ := ctxData["user_id"].(string)
	holder := rt.GetPlayers()[holderID]
	if bard == nil || holder == nil {
		return fmt.Errorf("吟游诗人或持有者不存在")
	}
	camp := string(holder.Camp)
	available := make([]string, 0, 2)
	if rt.GetCampGems(camp) > 0 {
		available = append(available, "gem")
	}
	if rt.GetCampCrystals(camp) > 0 {
		available = append(available, "crystal")
	}
	if selectionIndex < 0 || selectionIndex >= len(available) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	addGem, addCrystal := 0, 0
	switch available[selectionIndex] {
	case "gem":
		rt.ModifyGem(camp, -1)
		addGem = 1
	case "crystal":
		rt.ModifyCrystal(camp, -1)
		addCrystal = 1
	default:
		return fmt.Errorf("无效的星石类型")
	}
	maxEnergy := getPlayerEnergyCap(holder)
	room := maxEnergy - (holder.Gem + holder.Crystal)
	if room <= 0 {
		rt.Log(fmt.Sprintf("%s 的 [胜利交响诗]：提炼成功但能量已满，未增加个人能量", holder.Name))
	} else {
		if addGem > room {
			addGem = room
			addCrystal = 0
		}
		if addCrystal > room {
			addCrystal = room
			addGem = 0
		}
		holder.Gem += addGem
		holder.Crystal += addCrystal
		rt.Log(fmt.Sprintf("%s 发动 [胜利交响诗]（%s 触发）：提炼1个星石为个人能量（+%d宝石 +%d水晶）", bard.Name, holder.Name, addGem, addCrystal))
	}
	resolveBardForbiddenVerseAfterSong(rt, bard, "胜利交响诗")
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		rt.RoutePendingDamageOr(model.TurnStageTurnEnd, func() {
			rt.EnterTurnEndStage()
		})
	}
	return nil
}

// ---- 希望赋格曲 ----

func handleHopeDrawConfirm(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("吟游诗人不存在")
	}
	if selectionIndex == 0 {
		rt.DrawCards(user.ID, 1)
	} else if selectionIndex != 1 {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	ctxData["choice_type"] = "bd_hope_mode"
	intr := rt.GetPendingInterrupt()
	if intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
	return nil
}

func handleHopeMode(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("吟游诗人不存在")
	}
	switch selectionIndex {
	case 0:
		targetIDs := bardAlliesExcluding(rt, user.Camp, user.ID)
		if len(targetIDs) == 0 {
			return fmt.Errorf("无可选队友目标")
		}
		ctxData["choice_type"] = "bd_hope_place_target"
		ctxData["target_ids"] = targetIDs
		intr := rt.GetPendingInterrupt()
		if intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return nil
	case 1, 2:
		holderID := EternalHolderID(rt, user)
		if holderID == "" {
			return fmt.Errorf("当前没有永恒乐章可转移")
		}
		if len(user.Hand) == 0 {
			return fmt.Errorf("手牌不足，无法执行转移分支")
		}
		targetIDs := make([]string, 0)
		for _, targetID := range bardAlliesExcluding(rt, user.Camp, user.ID) {
			if targetID == holderID {
				continue
			}
			targetIDs = append(targetIDs, targetID)
		}
		if len(targetIDs) == 0 {
			return fmt.Errorf("没有可转移的目标角色")
		}
		ctxData["mode"] = selectionIndex
		ctxData["choice_type"] = "bd_hope_transfer_target"
		ctxData["target_ids"] = targetIDs
		intr := rt.GetPendingInterrupt()
		if intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return nil
	default:
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
}

func handleHopePlaceTarget(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("吟游诗人不存在")
	}
	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
	if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	target := rt.GetPlayers()[targetIDs[selectionIndex]]
	if target == nil {
		return fmt.Errorf("目标不存在")
	}
	if err := PlaceEternalMovement(rt, user, target); err != nil {
		return err
	}
	rt.Log(fmt.Sprintf("%s 发动 [希望赋格曲]：将永恒乐章放置于 %s 面前", user.Name, target.Name))
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		rt.ApplyChoiceResumePoint(model.TurnStageActionStart)
	}
	return nil
}

func handleHopeTransferTarget(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("吟游诗人不存在")
	}
	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
	if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	target := rt.GetPlayers()[targetIDs[selectionIndex]]
	if target == nil {
		return fmt.Errorf("目标不存在")
	}
	if err := transferBardEternalMovement(rt, user, target); err != nil {
		return err
	}
	ctxData["choice_type"] = "bd_hope_transfer_discard"
	intr := rt.GetPendingInterrupt()
	if intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
	return nil
}

func handleHopeTransferDiscard(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("吟游诗人不存在")
	}
	if selectionIndex < 0 || selectionIndex >= len(user.Hand) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	card := user.Hand[selectionIndex]
	user.Hand = append(user.Hand[:selectionIndex], user.Hand[selectionIndex+1:]...)
	rt.NotifyCardRevealed(user.ID, []model.Card{card}, "discard")
	rt.AppendToDiscard([]model.Card{card})
	mode := runtimeutil.ToIntContextValue(ctxData["mode"])
	switch mode {
	case 1:
		rt.Heal(user.ID, 1)
		rt.Log(fmt.Sprintf("%s 的 [希望赋格曲] 转移分支：+1治疗", user.Name))
	case 2:
		now := addBardInspiration(user, 1)
		rt.Log(fmt.Sprintf("%s 的 [希望赋格曲] 转移分支：灵感+1（当前%d）", user.Name, now))
	default:
		return fmt.Errorf("希望赋格曲转移分支模式无效")
	}
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		rt.ApplyChoiceResumePoint(model.TurnStageActionStart)
	}
	return nil
}

// ---- Helper functions ----

const bardInspirationCap = 3

func getSameElementCounts(player *model.Player) map[model.Element]int {
	out := map[model.Element]int{}
	if player == nil {
		return out
	}
	for _, c := range player.Hand {
		if c.Element == "" {
			continue
		}
		out[c.Element]++
	}
	return out
}

// Token helpers

func addTokenValueBounded(player *model.Player, key string, delta int, cap int) int {
	if player == nil {
		return 0
	}
	if player.Tokens == nil {
		player.Tokens = map[string]int{}
	}
	current := player.Tokens[key]
	newVal := current + delta
	if newVal < 0 {
		newVal = 0
	}
	if cap > 0 && newVal > cap {
		newVal = cap
	}
	player.Tokens[key] = newVal
	return newVal
}

func bardInspiration(player *model.Player) int {
	return engineplayer.TokenValue(player, "bd_inspiration", bardInspirationCap)
}

func addBardInspiration(player *model.Player, delta int) int {
	return addTokenValueBounded(player, "bd_inspiration", delta, bardInspirationCap)
}

// Camp helpers

func campEnemyIDs(rt engineplayer.ChoiceRuntime, user *model.Player) []string {
	if user == nil {
		return nil
	}
	var ids []string
	for _, pid := range rt.GetPlayerOrder() {
		p := rt.GetPlayers()[pid]
		if p == nil || p.Camp == user.Camp {
			continue
		}
		ids = append(ids, p.ID)
	}
	return ids
}

func bardAlliesExcluding(rt engineplayer.ChoiceRuntime, camp model.Camp, excludeID string) []string {
	var ids []string
	for _, pid := range rt.GetPlayerOrder() {
		p := rt.GetPlayers()[pid]
		if p == nil || p.Camp != camp || p.ID == excludeID {
			continue
		}
		ids = append(ids, p.ID)
	}
	return ids
}

// Bard eternal movement helpers

func transferBardEternalMovement(rt engineplayer.ChoiceRuntime, bard *model.Player, target *model.Player) error {
	if bard == nil || target == nil {
		return fmt.Errorf("转移永恒乐章时角色不存在")
	}
	if target.Camp != bard.Camp {
		return fmt.Errorf("永恒乐章只能转移给我方角色")
	}
	holder, _ := rt.FindEffectCard(bard, model.EffectBardEternalMovement)
	if holder == nil {
		return fmt.Errorf("当前没有永恒乐章可转移")
	}
	if holder.ID == target.ID {
		return fmt.Errorf("永恒乐章已在该角色面前")
	}
	// Detach from current holder and attach to new target.
	// We need to find the existing card first.
	_, existingCard := rt.FindEffectCard(bard, model.EffectBardEternalMovement)
	if existingCard == nil {
		return fmt.Errorf("当前没有永恒乐章可转移")
	}
	card := existingCard.Card
	// Remove from current holder by removing the field card.
	holder.RemoveFieldCard(existingCard)
	// Attach to new target.
	return rt.AttachEffectCard(bard, target, model.EffectBardEternalMovement, card)
}

// resolveBardForbiddenVerseAfterSong implements the forbidden verse after-song logic.
func resolveBardForbiddenVerseAfterSong(rt engineplayer.ChoiceRuntime, bard *model.Player, songName string) {
	if bard == nil {
		return
	}
	engineplayer.EnsurePlayerTokensMap(bard)
	if bardInspiration(bard) < bardInspirationCap {
		now := addBardInspiration(bard, 1)
		removed := RemoveEternalMovement(rt, bard)
		if removed {
			rt.Log(fmt.Sprintf("%s 的 [禁忌诗篇] 生效：灵感+1（当前%d），并移除永恒乐章", bard.Name, now))
		} else {
			rt.Log(fmt.Sprintf("%s 的 [禁忌诗篇] 生效：灵感+1（当前%d）", bard.Name, now))
		}
		return
	}

	if !InEternalPrisonerForm(bard) {
		EnterEternalPrisonerForm(bard)
		rt.Log(fmt.Sprintf("%s 的 [禁忌诗篇] 生效：转为永恒囚徒形态", bard.Name))
	}
	rt.AddPendingDamage(model.PendingDamage{
		SourceID:   bard.ID,
		TargetID:   bard.ID,
		Damage:     3,
		DamageType: model.MagicAttack,
	})
	rt.Log(fmt.Sprintf("%s 的 [禁忌诗篇] 生效：灵感已满，对自己造成3点法术伤害（来源：%s）", bard.Name, songName))
}

// addCampResource adds a camp resource (gem or crystal) respecting the max total of 5.
func addCampResource(rt engineplayer.ChoiceRuntime, camp string, resourceType string) bool {
	const maxTotalResources = 5
	currentTotal := rt.GetCampGems(camp) + rt.GetCampCrystals(camp)
	if currentTotal >= maxTotalResources {
		return false
	}
	switch resourceType {
	case "gem":
		rt.ModifyGem(camp, 1)
	default:
		rt.ModifyCrystal(camp, 1)
	}
	return true
}

// getPlayerEnergyCap returns the energy cap for a player (base 3).
func getPlayerEnergyCap(player *model.Player) int {
	if player == nil {
		return 3
	}
	return 3
}

var _ engineplayer.CancelChoiceHandler = choiceHandler{}
