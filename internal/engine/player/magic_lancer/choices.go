// gameflow: 魔枪士角色选择流。

package magic_lancer

import (
	"fmt"
	"strings"

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
	case "ml_black_spear_x":
		maxX := runtimeutil.ToIntContextValue(data["max_x"])
		if maxX < 1 {
			maxX = 1
		}
		options := make([]model.PromptOption, 0, maxX)
		for x := 1; x <= maxX; x++ {
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", x), Label: fmt.Sprintf("X=%d（消耗%d蓝水晶，伤害额外+%d）", x, x, x+2)})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【漆黑之枪】请选择X值：", Options: options, Min: 1, Max: 1, Presentation: &model.PromptPresentation{Kind: model.PresentationNumeric, NumericBase: 0}}

	case "ml_dark_barrier_mode":
		// 已废弃：改为直接进入 ml_dark_barrier_cards 单步选择
		return nil

	case "ml_dark_barrier_cards":
		options := make([]model.PromptOption, 0, len(player.Hand))
		for idx, card := range player.Hand {
			if card.Type == model.CardTypeMagic || card.Element == model.ElementThunder {
				options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", idx), Label: fmt.Sprintf("%d: %s", idx+1, promptfmt.FormatCardInfo(card))})
			}
		}
		return &model.Prompt{Type: model.PromptChooseCards, PlayerID: playerID, Message: "【暗之障壁】请选择要弃置的牌（法术牌或雷系牌）：", Options: options, Min: 1, Max: len(options), ChoiceType: "ml_dark_barrier_cards"}

	case "ml_fullness_cost_card":
		options := make([]model.PromptOption, 0, len(player.Hand))
		for idx, card := range player.Hand {
			if card.Type != model.CardTypeMagic && card.Element != model.ElementThunder {
				continue
			}
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", idx), Label: fmt.Sprintf("%d: %s", idx+1, promptfmt.FormatCardInfo(card))})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【充盈】请选择要弃置的1张法术牌或雷系牌：", Options: options, Min: 1, Max: 1}

	case "ml_fullness_discard_step":
		currentID, _ := data["current_player_id"].(string)
		target := rt.GetPlayers()[currentID]
		if target == nil {
			return nil
		}
		allowSkip, _ := data["allow_skip"].(bool)
		candidates := runtimeutil.ParseChoiceIntSlice(data["candidates"])
		options := make([]model.PromptOption, 0, len(candidates)+1)
		// 队友可选"不弃置"，放在第一位，索引=0 表示跳过
		if allowSkip {
			options = append(options, model.PromptOption{ID: "-1", Label: "不弃置"})
		}
		for _, idx := range candidates {
			if idx < 0 || idx >= len(target.Hand) {
				continue
			}
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", idx), Label: promptfmt.FormatCardInfo(target.Hand[idx])})
		}
		msg := "【充盈】请选择弃置1张手牌："
		if allowSkip {
			msg = "【充盈】请选择是否弃置1张手牌："
		}
		return &model.Prompt{Type: model.PromptChooseCards, PlayerID: playerID, Message: msg, Options: options, Min: 1, Max: 1, ChoiceType: "ml_fullness_discard_step"}

	case "ml_stardust_target":
		return engineplayer.BuildTargetChoicePrompt(rt, playerID, "【幻影星尘】请选择2点法术伤害目标：", data, false)
	}
	return nil
}

func (choiceHandler) HandleChoice(rt engineplayer.ChoiceRuntime, _ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "ml_black_spear_x":
		return true, handleMagicLancerBlackSpearXChoice(rt, selectionIndex, ctxData)
	case "ml_dark_barrier_mode":
		// 已废弃
		return false, fmt.Errorf("暗之障壁已改为单步选择")
	case "ml_dark_barrier_cards":
		handled, err := handleDarkBarrierCardsMultiSelect(rt, "", []int{selectionIndex}, ctxData)
		return handled, err
	case "ml_fullness_cost_card":
		return true, handleMagicLancerFullnessCostCardChoice(rt, selectionIndex, ctxData)
	case "ml_fullness_discard_step":
		return true, handleMagicLancerFullnessDiscardStepChoice(rt, selectionIndex, ctxData)
	case "ml_stardust_target":
		return true, handleMagicLancerStardustTargetChoice(rt, selectionIndex, ctxData)
	default:
		return false, nil
	}
}

func handleMagicLancerBlackSpearXChoice(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
	xValue := selectionIndex + 1
	if xValue < 1 || xValue > maxX {
		return fmt.Errorf("无效的X值")
	}
	if !rt.ConsumeCrystalCost(user.ID, xValue) {
		return fmt.Errorf("漆黑之枪需要%d点蓝水晶（红宝石可替代）", xValue)
	}
	targetID, _ := ctxData["target_id"].(string)
	bonus := xValue + 2
	applied := false
	queue := rt.GetPendingDamageQueue()
	for i := 0; i < len(queue); i++ {
		pd := queue[i]
		if !strings.EqualFold(string(pd.DamageType), string(model.AttackDamage)) {
			continue
		}
		if pd.SourceID != user.ID {
			continue
		}
		if targetID != "" && pd.TargetID != targetID {
			continue
		}
		queue[i].Damage += bonus
		applied = true
		break
	}
	if !applied {
		rt.Log("[Warn] 漆黑之枪未找到可叠加的攻击伤害条目")
	}
	rt.Log(fmt.Sprintf("%s 的 [漆黑之枪] 生效：消耗%d点蓝水晶，本次攻击伤害额外+%d", user.Name, xValue, bonus))
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		rt.EnterDamageResolution(nil)
	}
	return nil
}

// handleDarkBarrierCardsMultiSelect 处理暗之障壁弃牌多选。
func handleDarkBarrierCardsMultiSelect(rt engineplayer.ChoiceRuntime, playerID string, selections []int, ctxData map[string]interface{}) (bool, error) {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return false, fmt.Errorf("玩家不存在")
	}
	if len(selections) < 1 {
		return false, fmt.Errorf("暗之障壁至少需要选择1张牌")
	}

	// 验证：所选牌必须全为法术牌或全为雷系牌（不能混选）
	var hasMagic, hasThunder bool
	for _, idx := range selections {
		if idx < 0 || idx >= len(user.Hand) {
			return false, fmt.Errorf("无效的选项索引: %d", idx)
		}
		card := user.Hand[idx]
		if card.Type == model.CardTypeMagic {
			hasMagic = true
		}
		if card.Element == model.ElementThunder {
			hasThunder = true
		}
	}
	if hasMagic && hasThunder {
		return false, fmt.Errorf("暗之障壁需选择相同类型的牌（全法术牌或全雷系牌）")
	}
	if !hasMagic && !hasThunder {
		return false, fmt.Errorf("暗之障壁需选择法术牌或雷系牌")
	}

	xValue := len(selections)
	removed, err := engineplayer.RemoveCardsByIndicesFromHand(user, append([]int{}, selections...))
	if err != nil {
		return false, err
	}
	rt.NotifyCardRevealed(user.ID, removed, "discard")
	rt.AppendToDiscard(removed)

	modeLabel := "法术"
	if hasThunder && !hasMagic {
		modeLabel = "雷系"
	}
	rt.Log(fmt.Sprintf("%s 的 [暗之障壁] 生效：弃置%d张%s牌", user.Name, xValue, modeLabel))
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		if len(rt.GetPendingDamageQueue()) > 0 {
			rt.EnterDamageResolution(nil)
		} else if len(rt.GetActionQueue()) > 0 {
			rt.EnterResponseWindow()
		}
	}
	return true, nil
}

func handleMagicLancerFullnessCostCardChoice(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	candidates := make([]int, 0)
	for idx, card := range user.Hand {
		if card.Type == model.CardTypeMagic || card.Element == model.ElementThunder {
			candidates = append(candidates, idx)
		}
	}
	cardIdx, ok := runtimeutil.ResolveSelectionToCandidate(selectionIndex, candidates)
	if !ok || cardIdx < 0 || cardIdx >= len(user.Hand) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	costCard := user.Hand[cardIdx]
	user.Hand = append(user.Hand[:cardIdx], user.Hand[cardIdx+1:]...)
	rt.NotifyCardRevealed(user.ID, []model.Card{costCard}, "discard")
	rt.AppendToDiscard([]model.Card{costCard})

	// Build order: all players in reverse order from the user (excluding user themselves).
	orderIDs := engineplayer.ReversePlayerIDsFromRuntime(rt, user.ID, engineplayer.ReverseOrderOption{})
	ctxData["order_ids"] = orderIDs
	ctxData["order_index"] = 0
	ctxData["bonus"] = 0
	ctxData["choice_type"] = "ml_fullness_discard_step"

	done := prepareFullnessStep(rt, ctxData, user)
	if done {
		if user.TurnState.UsedSkillCounts == nil {
			user.TurnState.UsedSkillCounts = map[string]int{}
		}
		model.AppendAttackAction(user, "充盈")
		rt.Log(fmt.Sprintf("%s 的 [充盈] 生效：无可处理弃牌目标，获得额外1次攻击行动", user.Name))
		rt.PopInterrupt()
		if rt.GetPendingInterrupt() == nil {
			rt.EnterExtraActionStage()
		}
		return nil
	}
	rt.PopInterrupt()
	currentID, _ := ctxData["current_player_id"].(string)
	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: currentID,
		Context:  ctxData,
	})
	rt.NotifyInterruptPrompt()
	return nil
}

func handleMagicLancerFullnessDiscardStepChoice(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	currentID, _ := ctxData["current_player_id"].(string)
	target := rt.GetPlayers()[currentID]
	if target == nil {
		return fmt.Errorf("弃牌目标不存在")
	}
	allowSkip, _ := ctxData["allow_skip"].(bool)
	candidates := runtimeutil.ParseChoiceIntSlice(ctxData["candidates"])
	if len(candidates) == 0 {
		allowSkip = true
	}
	skipped := false
	var chosenCard model.Card
	if allowSkip && selectionIndex == 0 {
		skipped = true
	} else {
		optionIdx := selectionIndex
		if allowSkip {
			optionIdx--
		}
		cardIdx, ok := runtimeutil.ResolveSelectionToCandidate(optionIdx, candidates)
		if !ok || cardIdx < 0 || cardIdx >= len(target.Hand) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		chosenCard = target.Hand[cardIdx]
		target.Hand = append(target.Hand[:cardIdx], target.Hand[cardIdx+1:]...)
		rt.NotifyCardRevealed(target.ID, []model.Card{chosenCard}, "discard")
		rt.AppendToDiscard([]model.Card{chosenCard})
	}
	if skipped {
		rt.Log(fmt.Sprintf("%s 的 [充盈]：%s 选择不弃牌", user.Name, target.Name))
	} else {
		rt.Log(fmt.Sprintf("%s 的 [充盈]：%s 弃置了 %s", user.Name, target.Name, chosenCard.Name))
		if target.ID != user.ID && (chosenCard.Type == model.CardTypeMagic || chosenCard.Element == model.ElementThunder) {
			ctxData["bonus"] = runtimeutil.ToIntContextValue(ctxData["bonus"]) + 1
		}
	}

	ctxData["order_index"] = runtimeutil.ToIntContextValue(ctxData["order_index"]) + 1
	done := prepareFullnessStep(rt, ctxData, user)
	if !done {
		rt.PopInterrupt()
		nextID, _ := ctxData["current_player_id"].(string)
		rt.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: nextID,
			Context:  ctxData,
		})
		rt.NotifyInterruptPrompt()
		return nil
	}

	bonus := runtimeutil.ToIntContextValue(ctxData["bonus"])
	if bonus > 0 {
		rt.ApplyNextAttackDamageRule(user.ID, "ml_fullness_next_attack_bonus", "ml_fullness", bonus, model.RuleLifeUntilTurnEnd)
	}
	model.AppendAttackAction(user, "充盈")
	rt.Log(fmt.Sprintf("%s 的 [充盈] 结算完成：本回合下次主动攻击伤害额外+%d，额外获得1次攻击行动", user.Name, bonus))
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		if !rt.RoutePendingDamageWithReturn(model.TurnStageExtraAction) {
			rt.EnterExtraActionStage()
		}
	}
	return nil
}

func handleMagicLancerStardustTargetChoice(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
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
	rt.AddPendingDamage(model.PendingDamage{
		SourceID:   user.ID,
		TargetID:   targetID,
		Damage:     2,
		DamageType: model.MagicAttack,
	})
	if target := rt.GetPlayers()[targetID]; target != nil {
		rt.Log(fmt.Sprintf("%s 发动 [幻影星尘]：对 %s 造成2点法术伤害", user.Name, target.Name))
	}
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		if len(rt.GetPendingDamageQueue()) > 0 {
			rt.EnterDamageResolution(nil)
		} else {
			rt.EnterExtraActionStage()
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// prepareFullnessStep advances the fullness discard step to the next target.
// Returns true if all targets have been processed.
// Each player with cards gets their own interrupt (PlayerID = target), so they
// choose which card to discard on their own action panel.
// Allies get a "skip" option; enemies must discard.
func prepareFullnessStep(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, user *model.Player) bool {
	orderIDs := runtimeutil.ParseStringSliceContextValue(ctxData["order_ids"])
	if len(orderIDs) == 0 {
		return true
	}
	idx := runtimeutil.ToIntContextValue(ctxData["order_index"])
	if idx < 0 {
		idx = 0
	}
	for idx < len(orderIDs) {
		pid := orderIDs[idx]
		target := rt.GetPlayers()[pid]
		if target == nil {
			idx++
			continue
		}
		candidates := engineplayer.AllHandIndices(target)
		if len(candidates) == 0 {
			rt.Log(fmt.Sprintf("%s 的 [充盈]：%s 无手牌可弃，跳过", user.Name, target.Name))
			idx++
			continue
		}
		ctxData["order_index"] = idx
		ctxData["current_player_id"] = pid
		ctxData["allow_skip"] = target.Camp == user.Camp // 队友可跳过，敌方不可
		ctxData["candidates"] = candidates
		return false
	}
	ctxData["order_index"] = idx
	return true
}
