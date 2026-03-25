package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

func (e *GameEngine) buildBardChoicePrompt(choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "bd_descent_element":
		elemCounts := getSameElementCounts(player)
		elems := make([]model.Element, 0)
		for _, ele := range elementOrderForPrompt() {
			if elemCounts[ele] >= 2 {
				elems = append(elems, ele)
			}
		}
		options := make([]model.PromptOption, 0, len(elems))
		for _, ele := range elems {
			options = append(options, model.PromptOption{ID: string(ele), Label: fmt.Sprintf("%s系", elementNameForPrompt(string(ele)))})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【沉沦协奏曲】请选择要弃置的同系元素：", Options: options, Min: 1, Max: 1}
	case "bd_descent_cards":
		chosenEle, _ := data["chosen_element"].(string)
		chosenEleZh := elementNameForPrompt(chosenEle)
		remaining := parseIntSliceContextValue(data["remaining_indices"])
		selected := len(parseIntSliceContextValue(data["selected_indices"]))
		options := make([]model.PromptOption, 0, len(remaining))
		for _, idx := range remaining {
			if idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", idx), Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(player.Hand[idx]))})
		}
		remainingPick := 2 - selected
		if remainingPick < 1 {
			remainingPick = 1
		}
		if len(options) > 0 && remainingPick > len(options) {
			remainingPick = len(options)
		}
		return &model.Prompt{Type: model.PromptChooseCards, PlayerID: playerID, Message: fmt.Sprintf("【沉沦协奏曲】请选择要弃置的%d张%s系牌：", remainingPick, chosenEleZh), Options: options, Min: remainingPick, Max: remainingPick}
	case "bd_dissonance_x":
		maxX := toIntContextValue(data["max_x"])
		if maxX < 2 {
			maxX = 2
		}
		options := make([]model.PromptOption, 0, maxX-1)
		for x := 2; x <= maxX; x++ {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", x), Label: fmt.Sprintf("X=%d", x)})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【不谐和弦】请选择X值：", Options: options, Min: 1, Max: 1}
	case "bd_dissonance_mode":
		xValue := toIntContextValue(data["x_value"])
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: fmt.Sprintf("【不谐和弦】请选择分支（X=%d）：", xValue), Options: []model.PromptOption{{ID: "0", Label: fmt.Sprintf("你与目标各摸%d张牌", xValue-1)}, {ID: "1", Label: fmt.Sprintf("你与目标各弃%d张牌", xValue-1)}}, Min: 1, Max: 1}
	case "bd_dissonance_discard_step":
		currentActorID, _ := data["current_actor_id"].(string)
		actor := e.State.Players[currentActorID]
		if actor == nil {
			return nil
		}
		need := toIntContextValue(data["need_count"])
		selected := toIntContextValue(data["selected_count"])
		options := make([]model.PromptOption, 0, len(actor.Hand))
		for idx, c := range actor.Hand {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", idx), Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(c))})
		}
		remainingPick := need - selected
		if remainingPick < 1 {
			remainingPick = 1
		}
		if len(options) > 0 && remainingPick > len(options) {
			remainingPick = len(options)
		}
		return &model.Prompt{Type: model.PromptChooseCards, PlayerID: playerID, Message: fmt.Sprintf("【不谐和弦】请 %s 选择要弃置的%d张手牌：", actor.Name, remainingPick), Options: options, Min: remainingPick, Max: remainingPick}
	case "bd_rousing_mode":
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【激昂狂想曲】请选择效果：", Options: []model.PromptOption{{ID: "0", Label: "对2名对手各造成1点法术伤害"}, {ID: "1", Label: "弃2张牌"}}, Min: 1, Max: 1}
	case "bd_rousing_targets":
		targetIDs := parseStringSliceContextValue(data["target_ids"])
		selectedSet := idsToSet(parseStringSliceContextValue(data["selected_target_ids"]))
		options := make([]model.PromptOption, 0, len(targetIDs))
		for _, targetID := range targetIDs {
			if selectedSet[targetID] {
				continue
			}
			if target := e.State.Players[targetID]; target != nil {
				options = append(options, model.PromptOption{ID: targetID, Label: target.Name})
			}
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: fmt.Sprintf("【激昂狂想曲】请选择第 %d/2 名目标：", len(selectedSet)+1), Options: options, Min: 1, Max: 1}
	case "bd_rousing_discard_cards":
		selected := len(parseIntSliceContextValue(data["selected_indices"]))
		remaining := parseIntSliceContextValue(data["remaining_indices"])
		options := make([]model.PromptOption, 0, len(remaining))
		for _, idx := range remaining {
			if idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", idx), Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(player.Hand[idx]))})
		}
		remainingPick := 2 - selected
		if remainingPick < 1 {
			remainingPick = 1
		}
		if len(options) > 0 && remainingPick > len(options) {
			remainingPick = len(options)
		}
		return &model.Prompt{Type: model.PromptChooseCards, PlayerID: playerID, Message: fmt.Sprintf("【激昂狂想曲】请选择要弃置的%d张手牌：", remainingPick), Options: options, Min: remainingPick, Max: remainingPick}
	case "bd_victory_mode":
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【胜利交响诗】请选择效果：", Options: []model.PromptOption{{ID: "0", Label: "将我方战绩区1个星石提炼为你的能量"}, {ID: "1", Label: "我方战绩区+1宝石，你+1治疗"}}, Min: 1, Max: 1}
	case "bd_victory_extract_stone":
		options := make([]model.PromptOption, 0, 2)
		if player != nil {
			camp := string(player.Camp)
			if e.GetCampGems(camp) > 0 {
				options = append(options, model.PromptOption{ID: "0", Label: "提炼1个宝石"})
			}
			if e.GetCampCrystals(camp) > 0 {
				options = append(options, model.PromptOption{ID: "1", Label: "提炼1个水晶"})
			}
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【胜利交响诗】请选择要提炼的星石：", Options: options, Min: 1, Max: 1}
	case "bd_hope_draw_confirm":
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【希望赋格曲】是否先摸1张牌？", Options: []model.PromptOption{{ID: "0", Label: "是"}, {ID: "1", Label: "否"}}, Min: 1, Max: 1}
	case "bd_hope_mode":
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【希望赋格曲】请选择分支：", Options: []model.PromptOption{{ID: "0", Label: "将永恒乐章放置于目标队友面前"}, {ID: "1", Label: "转移永恒乐章，弃1张牌并+1治疗"}, {ID: "2", Label: "转移永恒乐章，弃1张牌并+1灵感"}}, Min: 1, Max: 1}
	case "bd_hope_transfer_discard":
		options := make([]model.PromptOption, 0, len(player.Hand))
		for idx, c := range player.Hand {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", idx), Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(c))})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【希望赋格曲】请选择弃置1张手牌：", Options: options, Min: 1, Max: 1}
	}
	return nil
}

func (e *GameEngine) handleBardChoiceInput(_ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "bd_descent_element":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		elemCounts := getSameElementCounts(user)
		elems := make([]model.Element, 0)
		for _, ele := range elementOrderForPrompt() {
			if elemCounts[ele] >= 2 {
				elems = append(elems, ele)
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(elems) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		chosen := elems[selectionIndex]
		ctxData["chosen_element"] = string(chosen)
		ctxData["selected_indices"] = []int{}
		ctxData["remaining_indices"] = getCardIndicesByElement(user, chosen)
		ctxData["choice_type"] = "bd_descent_cards"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil
	case "bd_descent_cards":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		chosenElement, _ := ctxData["chosen_element"].(string)
		remaining := parseIntSliceContextValue(ctxData["remaining_indices"])
		selected := parseIntSliceContextValue(ctxData["selected_indices"])
		cardIdx, ok := resolveSelectionToCandidate(selectionIndex, remaining)
		if !ok || cardIdx < 0 || cardIdx >= len(user.Hand) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		if string(user.Hand[cardIdx].Element) != chosenElement {
			return true, fmt.Errorf("沉沦协奏曲需弃置同系牌")
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
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return true, nil
		}
		removed, err := removeCardsByIndicesFromHand(user, append([]int{}, selected...))
		if err != nil {
			return true, err
		}
		e.NotifyCardRevealed(user.ID, removed, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, removed...)
		if user.Tokens == nil {
			user.Tokens = map[string]int{}
		}
		user.Tokens["bd_descent_used_turn"] = 1
		now := addBardInspiration(user, 1)
		e.Log(fmt.Sprintf("%s 发动 [沉沦协奏曲]：弃2张%s系牌，灵感+1（当前%d）", user.Name, chosenElement, now))

		hasMagic := false
		for _, card := range removed {
			if card.Type == model.CardTypeMagic {
				hasMagic = true
				break
			}
		}
		if hasMagic {
			ctxData["choice_type"] = "bd_descent_target"
			ctxData["target_ids"] = e.campEnemyIDs(user.Camp)
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return true, nil
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil && len(e.State.PendingDamageQueue) > 0 {
			e.enterDamageResolution(nil)
		}
		return true, nil
	case "bd_descent_target":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		targetIDs := parseStringSliceContextValue(ctxData["target_ids"])
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		targetID := targetIDs[selectionIndex]
		target := e.State.Players[targetID]
		if target == nil {
			return true, fmt.Errorf("目标不存在")
		}
		e.AddPendingDamage(model.PendingDamage{SourceID: user.ID, TargetID: target.ID, Damage: 1, DamageType: "magic"})
		e.Log(fmt.Sprintf("%s 的 [沉沦协奏曲] 追加效果：对 %s 造成1点法术伤害", user.Name, target.Name))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.enterDamageResolution(nil)
		}
		return true, nil
	case "bd_dissonance_x":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		maxX := toIntContextValue(ctxData["max_x"])
		xValue := selectionIndex + 2
		if xValue < 2 || xValue > maxX {
			return true, fmt.Errorf("无效的X值")
		}
		if bardInspiration(user) < xValue {
			return true, fmt.Errorf("灵感不足")
		}
		addBardInspiration(user, -xValue)
		if hasBardEternalPrisonerForm(user) {
			beforePoses := e.snapshotPlayerPoses()
			leaveBardEternalPrisonerForm(user)
			e.Log(fmt.Sprintf("%s 发动 [不谐和弦]：脱离永恒囚徒形态", user.Name))
			e.dispatchOrientationChanges(beforePoses)
		}
		ctxData["x_value"] = xValue
		ctxData["choice_type"] = "bd_dissonance_mode"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil
	case "bd_dissonance_mode":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		if selectionIndex != 0 && selectionIndex != 1 {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		ctxData["mode"] = selectionIndex
		ctxData["choice_type"] = "bd_dissonance_target"
		ctxData["target_ids"] = append([]string{}, e.State.PlayerOrder...)
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil
	case "bd_dissonance_target":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		targetIDs := parseStringSliceContextValue(ctxData["target_ids"])
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		targetID := targetIDs[selectionIndex]
		target := e.State.Players[targetID]
		if target == nil {
			return true, fmt.Errorf("目标不存在")
		}
		xValue := toIntContextValue(ctxData["x_value"])
		mode := toIntContextValue(ctxData["mode"])
		n := xValue - 1
		if n < 0 {
			n = 0
		}
		if mode == 0 {
			if n > 0 {
				e.DrawCards(user.ID, n)
				e.DrawCards(target.ID, n)
			}
			e.Log(fmt.Sprintf("%s 发动 [不谐和弦]：与 %s 各摸%d张牌", user.Name, target.Name, n))
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				e.routePendingDamageOr(model.TurnStageExtraAction, func() {
					e.enterExtraActionStage()
				})
			}
			return true, nil
		}
		actors := []string{user.ID, target.ID}
		startCursor := 0
		for startCursor < len(actors) {
			actor := e.State.Players[actors[startCursor]]
			if actor != nil && len(actor.Hand) > 0 && n > 0 {
				break
			}
			startCursor++
		}
		if n <= 0 || startCursor >= len(actors) {
			e.Log(fmt.Sprintf("%s 发动 [不谐和弦]：弃牌分支无可执行弃牌", user.Name))
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				e.enterExtraActionStage()
			}
			return true, nil
		}
		currentActor := e.State.Players[actors[startCursor]]
		ctxData["choice_type"] = "bd_dissonance_discard_step"
		ctxData["actor_ids"] = actors
		ctxData["cursor"] = startCursor
		ctxData["current_actor_id"] = currentActor.ID
		ctxData["need_count"] = n
		ctxData["selected_count"] = 0
		ctxData["selected_indices"] = []int{}
		ctxData["remaining_indices"] = allHandIndices(currentActor)
		e.State.PendingInterrupt.PlayerID = currentActor.ID
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil
	case "bd_dissonance_discard_step":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		actorIDs := parseStringSliceContextValue(ctxData["actor_ids"])
		cursor := toIntContextValue(ctxData["cursor"])
		if cursor < 0 || cursor >= len(actorIDs) {
			return true, fmt.Errorf("弃牌游标无效")
		}
		currentActorID, _ := ctxData["current_actor_id"].(string)
		if currentActorID == "" {
			currentActorID = actorIDs[cursor]
		}
		actor := e.State.Players[currentActorID]
		if actor == nil {
			return true, fmt.Errorf("弃牌角色不存在")
		}
		needCount := toIntContextValue(ctxData["need_count"])
		selectedCount := toIntContextValue(ctxData["selected_count"])
		remaining := parseIntSliceContextValue(ctxData["remaining_indices"])
		selected := parseIntSliceContextValue(ctxData["selected_indices"])
		cardIdx, ok := resolveSelectionToCandidate(selectionIndex, remaining)
		if !ok || cardIdx < 0 || cardIdx >= len(actor.Hand) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
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
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return true, nil
		}
		removed, err := removeCardsByIndicesFromHand(actor, append([]int{}, selected...))
		if err != nil {
			return true, err
		}
		e.NotifyCardHidden(actor.ID, removed, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, removed...)
		e.Log(fmt.Sprintf("%s 的 [不谐和弦]：%s 弃置了%d张手牌", user.Name, actor.Name, len(removed)))

		nextCursor := cursor + 1
		for nextCursor < len(actorIDs) {
			nextActor := e.State.Players[actorIDs[nextCursor]]
			if nextActor == nil || len(nextActor.Hand) == 0 || needCount <= 0 {
				nextCursor++
				continue
			}
			ctxData["cursor"] = nextCursor
			ctxData["current_actor_id"] = nextActor.ID
			ctxData["selected_count"] = 0
			ctxData["selected_indices"] = []int{}
			ctxData["remaining_indices"] = allHandIndices(nextActor)
			e.State.PendingInterrupt.PlayerID = nextActor.ID
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return true, nil
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.routePendingDamageOr(model.TurnStageExtraAction, func() {
				e.enterExtraActionStage()
			})
		}
		return true, nil
	case "bd_rousing_mode":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("吟游诗人不存在")
		}
		switch selectionIndex {
		case 0:
			ctxData["choice_type"] = "bd_rousing_targets"
			ctxData["selected_target_ids"] = []string{}
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return true, nil
		case 1:
			if len(user.Hand) < 2 {
				return true, fmt.Errorf("手牌不足2张，无法执行弃2张牌分支")
			}
			ctxData["choice_type"] = "bd_rousing_discard_cards"
			ctxData["selected_indices"] = []int{}
			ctxData["remaining_indices"] = allHandIndices(user)
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return true, nil
		default:
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
	case "bd_rousing_targets":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("吟游诗人不存在")
		}
		targetIDs := parseStringSliceContextValue(ctxData["target_ids"])
		selected := dedupeIDs(parseStringSliceContextValue(ctxData["selected_target_ids"]))
		selectedSet := idsToSet(selected)
		remaining := make([]string, 0, len(targetIDs))
		for _, targetID := range targetIDs {
			if !selectedSet[targetID] {
				remaining = append(remaining, targetID)
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(remaining) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		selected = append(selected, remaining[selectionIndex])
		if len(selected) < 2 {
			ctxData["selected_target_ids"] = selected
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return true, nil
		}
		for _, targetID := range selected {
			e.AddPendingDamage(model.PendingDamage{SourceID: user.ID, TargetID: targetID, Damage: 1, DamageType: "magic"})
		}
		e.Log(fmt.Sprintf("%s 发动 [激昂狂想曲]：对2名目标各造成1点法术伤害", user.Name))
		e.resolveBardForbiddenVerseAfterSong(user, "激昂狂想曲")
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.routePendingDamageOr(model.TurnStageActionStart, func() {
				e.setTurnStage(model.TurnStageActionStart)
				e.clearCombatStage()
				e.clearSubflow()
			})
		}
		return true, nil
	case "bd_rousing_discard_cards":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("吟游诗人不存在")
		}
		remaining := parseIntSliceContextValue(ctxData["remaining_indices"])
		selected := parseIntSliceContextValue(ctxData["selected_indices"])
		cardIdx, ok := resolveSelectionToCandidate(selectionIndex, remaining)
		if !ok || cardIdx < 0 || cardIdx >= len(user.Hand) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
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
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return true, nil
		}
		removed, err := removeCardsByIndicesFromHand(user, append([]int{}, selected...))
		if err != nil {
			return true, err
		}
		e.NotifyCardRevealed(user.ID, removed, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, removed...)
		e.Log(fmt.Sprintf("%s 发动 [激昂狂想曲]：选择弃2张牌", user.Name))
		e.resolveBardForbiddenVerseAfterSong(user, "激昂狂想曲")
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.routePendingDamageOr(model.TurnStageActionStart, func() {
				e.setTurnStage(model.TurnStageActionStart)
				e.clearCombatStage()
				e.clearSubflow()
			})
		}
		return true, nil
	case "bd_victory_mode":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("吟游诗人不存在")
		}
		switch selectionIndex {
		case 0:
			camp := string(user.Camp)
			if e.GetCampGems(camp)+e.GetCampCrystals(camp) <= 0 {
				return true, fmt.Errorf("我方战绩区没有可提炼的星石")
			}
			ctxData["choice_type"] = "bd_victory_extract_stone"
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return true, nil
		case 1:
			e.addCampResource(user.Camp, "gem")
			e.Heal(user.ID, 1)
			e.Log(fmt.Sprintf("%s 发动 [胜利交响诗]：我方战绩区+1宝石，自己+1治疗", user.Name))
		default:
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		e.resolveBardForbiddenVerseAfterSong(user, "胜利交响诗")
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.routePendingDamageOr(model.TurnStageTurnEnd, func() {
				e.enterTurnEndStage()
			})
		}
		return true, nil
	case "bd_victory_extract_stone":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("吟游诗人不存在")
		}
		camp := string(user.Camp)
		available := make([]string, 0, 2)
		if e.GetCampGems(camp) > 0 {
			available = append(available, "gem")
		}
		if e.GetCampCrystals(camp) > 0 {
			available = append(available, "crystal")
		}
		if selectionIndex < 0 || selectionIndex >= len(available) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		addGem, addCrystal := 0, 0
		switch available[selectionIndex] {
		case "gem":
			e.ModifyGem(camp, -1)
			addGem = 1
		case "crystal":
			e.ModifyCrystal(camp, -1)
			addCrystal = 1
		default:
			return true, fmt.Errorf("无效的星石类型")
		}
		maxEnergy := e.getPlayerEnergyCap(user)
		room := maxEnergy - (user.Gem + user.Crystal)
		if room <= 0 {
			e.Log(fmt.Sprintf("%s 的 [胜利交响诗]：提炼成功但能量已满，未增加个人能量", user.Name))
		} else {
			if addGem > room {
				addGem = room
				addCrystal = 0
			}
			if addCrystal > room {
				addCrystal = room
				addGem = 0
			}
			user.Gem += addGem
			user.Crystal += addCrystal
			e.Log(fmt.Sprintf("%s 发动 [胜利交响诗]：提炼1个星石为个人能量（+%d宝石 +%d水晶）", user.Name, addGem, addCrystal))
		}
		e.resolveBardForbiddenVerseAfterSong(user, "胜利交响诗")
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.routePendingDamageOr(model.TurnStageTurnEnd, func() {
				e.enterTurnEndStage()
			})
		}
		return true, nil
	case "bd_hope_draw_confirm":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("吟游诗人不存在")
		}
		if selectionIndex == 0 {
			e.DrawCards(user.ID, 1)
		} else if selectionIndex != 1 {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		ctxData["choice_type"] = "bd_hope_mode"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil
	case "bd_hope_mode":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("吟游诗人不存在")
		}
		switch selectionIndex {
		case 0:
			targetIDs := e.bardAlliesExcluding(user.Camp, user.ID)
			if len(targetIDs) == 0 {
				return true, fmt.Errorf("无可选队友目标")
			}
			ctxData["choice_type"] = "bd_hope_place_target"
			ctxData["target_ids"] = targetIDs
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return true, nil
		case 1, 2:
			holderID := e.bardEternalHolderID(user)
			if holderID == "" {
				return true, fmt.Errorf("当前没有永恒乐章可转移")
			}
			if len(user.Hand) == 0 {
				return true, fmt.Errorf("手牌不足，无法执行转移分支")
			}
			targetIDs := make([]string, 0)
			for _, targetID := range e.bardAlliesExcluding(user.Camp, user.ID) {
				if targetID == holderID {
					continue
				}
				targetIDs = append(targetIDs, targetID)
			}
			if len(targetIDs) == 0 {
				return true, fmt.Errorf("没有可转移的目标角色")
			}
			ctxData["mode"] = selectionIndex
			ctxData["choice_type"] = "bd_hope_transfer_target"
			ctxData["target_ids"] = targetIDs
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return true, nil
		default:
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
	case "bd_hope_place_target":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("吟游诗人不存在")
		}
		targetIDs := parseStringSliceContextValue(ctxData["target_ids"])
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		target := e.State.Players[targetIDs[selectionIndex]]
		if target == nil {
			return true, fmt.Errorf("目标不存在")
		}
		playedCard, ok := ctxData["played_card"].(model.Card)
		if !ok {
			return true, fmt.Errorf("希望赋格曲的专属牌上下文丢失")
		}
		if err := e.placeBardEternalMovementWithCard(user, target, playedCard); err != nil {
			return true, err
		}
		e.Log(fmt.Sprintf("%s 发动 [希望赋格曲]：将永恒乐章放置于 %s 面前", user.Name, target.Name))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.setTurnStage(model.TurnStageActionStart)
			e.clearCombatStage()
			e.clearSubflow()
		}
		return true, nil
	case "bd_hope_transfer_target":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("吟游诗人不存在")
		}
		targetIDs := parseStringSliceContextValue(ctxData["target_ids"])
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		target := e.State.Players[targetIDs[selectionIndex]]
		if target == nil {
			return true, fmt.Errorf("目标不存在")
		}
		if err := e.transferBardEternalMovement(user, target); err != nil {
			return true, err
		}
		ctxData["choice_type"] = "bd_hope_transfer_discard"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil
	case "bd_hope_transfer_discard":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("吟游诗人不存在")
		}
		if selectionIndex < 0 || selectionIndex >= len(user.Hand) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		card := user.Hand[selectionIndex]
		user.Hand = append(user.Hand[:selectionIndex], user.Hand[selectionIndex+1:]...)
		e.NotifyCardRevealed(user.ID, []model.Card{card}, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, card)
		playedCard, ok := ctxData["played_card"].(model.Card)
		if !ok {
			return true, fmt.Errorf("希望赋格曲的专属牌上下文丢失")
		}
		e.State.DiscardPile = append(e.State.DiscardPile, playedCard)
		mode := toIntContextValue(ctxData["mode"])
		switch mode {
		case 1:
			e.Heal(user.ID, 1)
			e.Log(fmt.Sprintf("%s 的 [希望赋格曲] 转移分支：+1治疗", user.Name))
		case 2:
			now := addBardInspiration(user, 1)
			e.Log(fmt.Sprintf("%s 的 [希望赋格曲] 转移分支：灵感+1（当前%d）", user.Name, now))
		default:
			return true, fmt.Errorf("希望赋格曲转移分支模式无效")
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.setTurnStage(model.TurnStageActionStart)
			e.clearCombatStage()
			e.clearSubflow()
		}
		return true, nil
	}
	return false, nil
}

func parseIntSliceContextValue(raw interface{}) []int {
	result := make([]int, 0)
	switch value := raw.(type) {
	case []int:
		result = append(result, value...)
	case []interface{}:
		for _, item := range value {
			switch v := item.(type) {
			case int:
				result = append(result, v)
			case float64:
				result = append(result, int(v))
			}
		}
	}
	return result
}
