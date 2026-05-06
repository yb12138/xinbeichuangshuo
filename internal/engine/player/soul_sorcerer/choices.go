// gameflow: 灵魂术士角色选择流。

package soul_sorcerer

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
	case "ss_convert_color":
		var modeOrder []string
		if arr, ok := data["mode_order"].([]string); ok {
			modeOrder = append(modeOrder, arr...)
		} else if arr, ok := data["mode_order"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					modeOrder = append(modeOrder, s)
				}
			}
		}
		var options []model.PromptOption
		for _, mode := range modeOrder {
			switch mode {
			case "y2b":
				options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "黄魂 -> 蓝魂（转换1点）"})
			case "b2y":
				options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "蓝魂 -> 黄魂（转换1点）"})
			}
		}
		return &model.Prompt{
			Type:       model.PromptConfirm,
			PlayerID:   playerID,
			ChoiceType: choiceType,
			Message:    "【灵魂转换】请选择转换方向：",
			Options:    options,
			Min:        1,
			Max:        1,
		}
	case "ss_link_target":
		var allyIDs []string
		if arr, ok := data["ally_ids"].([]string); ok {
			allyIDs = append(allyIDs, arr...)
		} else if arr, ok := data["ally_ids"].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					allyIDs = append(allyIDs, s)
				}
			}
		}
		var options []model.PromptOption
		for _, aid := range allyIDs {
			if p := rt.GetPlayers()[aid]; p != nil {
				options = append(options, model.PromptOption{ID: aid, Label: p.Name})
			}
		}
		return &model.Prompt{
			Type:       model.PromptConfirm,
			PlayerID:   playerID,
			ChoiceType: choiceType,
			Message:    "【灵魂链接】请选择要放置灵魂链接的队友：",
			Options:    options,
			Min:        1,
			Max:        1,
		}
	case "ss_link_transfer_x":
		maxX := runtimeutil.ToIntContextValue(data["max_x"])
		if maxX < 0 {
			maxX = 0
		}
		xOptions := make([]model.PromptOption, 0, maxX+1)
		for x := 0; x <= maxX; x++ {
			label := fmt.Sprintf("移除%d点蓝魂并转移%d点伤害", x, x)
			if x == 0 {
				label = "不转移伤害"
			}
			xOptions = append(xOptions, model.PromptOption{ID: fmt.Sprintf("%d", x), Label: label})
		}
		return &model.Prompt{
			Type:       model.PromptConfirm,
			PlayerID:   playerID,
			ChoiceType: choiceType,
			Message:    "【灵魂链接】请选择要转移的伤害点数X：",
			Options:    xOptions,
			Min:        1,
			Max:        1,
		}
	case "ss_recall_pick":
		var magicIndices []int
		if arr, ok := data["magic_indices"].([]int); ok {
			magicIndices = append(magicIndices, arr...)
		} else if arr, ok := data["magic_indices"].([]interface{}); ok {
			for _, v := range arr {
				if f, ok := v.(float64); ok {
					magicIndices = append(magicIndices, int(f))
				}
			}
		}
		if len(magicIndices) == 0 {
			if arr, ok := data["remaining_indices"].([]int); ok {
				magicIndices = append(magicIndices, arr...)
			} else if arr, ok := data["remaining_indices"].([]interface{}); ok {
				for _, v := range arr {
					if f, ok := v.(float64); ok {
						magicIndices = append(magicIndices, int(f))
					}
				}
			}
		}
		var options []model.PromptOption
		for _, idx := range magicIndices {
			if player == nil || idx < 0 || idx >= len(player.Hand) {
				continue
			}
			if player.Hand[idx].Type != model.CardTypeMagic {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    fmt.Sprintf("%d", idx),
				Label: fmt.Sprintf("%d: %s", idx+1, promptfmt.FormatCardInfo(player.Hand[idx])),
			})
		}
		maxSelect := len(options)
		if maxSelect < 1 {
			maxSelect = 1
		}
		return &model.Prompt{
			Type:       model.PromptChooseCards,
			PlayerID:   playerID,
			ChoiceType: choiceType,
			Message:    "【灵魂召还】请选择要弃置的法术牌（至少1张）：",
			Options:    options,
			Min:        1,
			Max:        maxSelect,
		}
	}
	return nil
}

func (choiceHandler) HandleChoice(rt engineplayer.ChoiceRuntime, _ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "ss_convert_color":
		return true, handleSoulConvertColorChoice(rt, selectionIndex, ctxData)
	case "ss_link_target":
		return true, handleSoulLinkTargetChoice(rt, selectionIndex, ctxData)
	case "ss_link_transfer_x":
		return true, handleSoulLinkTransferXChoice(rt, selectionIndex, ctxData)
	case "ss_recall_pick":
		return true, handleSoulRecallPickChoice(rt, selectionIndex, ctxData)
	default:
		return false, nil
	}
}

func handleSoulConvertColorChoice(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}

	modeOrder := runtimeutil.ParseStringSliceContextValue(ctxData["mode_order"])
	if selectionIndex < 0 || selectionIndex >= len(modeOrder) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}

	mode := modeOrder[selectionIndex]
	switch mode {
	case "y2b":
		if YellowSoul(user) <= 0 {
			return fmt.Errorf("黄色灵魂不足")
		}
		if BlueSoul(user) >= BlueSoulCap {
			return fmt.Errorf("蓝色灵魂已满")
		}
		AddYellowSoul(user, -1)
		AddBlueSoul(user, 1)
		rt.Log(fmt.Sprintf("%s 的 [灵魂转换] 生效：黄魂-1，蓝魂+1（黄:%d 蓝:%d）", user.Name, YellowSoul(user), BlueSoul(user)))
	case "b2y":
		if BlueSoul(user) <= 0 {
			return fmt.Errorf("蓝色灵魂不足")
		}
		if YellowSoul(user) >= YellowSoulCap {
			return fmt.Errorf("黄色灵魂已满")
		}
		AddBlueSoul(user, -1)
		AddYellowSoul(user, 1)
		rt.Log(fmt.Sprintf("%s 的 [灵魂转换] 生效：蓝魂-1，黄魂+1（黄:%d 蓝:%d）", user.Name, YellowSoul(user), BlueSoul(user)))
	default:
		return fmt.Errorf("无效的灵魂转换模式")
	}

	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		rt.RoutePendingDamageOr(model.TurnStageExtraAction, func() {
			if len(rt.GetActionQueue()) > 0 {
				rt.EnterActionExecutionStage()
			} else {
				rt.EnterExtraActionStage()
			}
		})
	}
	return nil
}

func handleSoulLinkTargetChoice(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}

	allyIDs := runtimeutil.ParseStringSliceContextValue(ctxData["ally_ids"])
	if selectionIndex < 0 || selectionIndex >= len(allyIDs) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}

	target := rt.GetPlayers()[allyIDs[selectionIndex]]
	if target == nil {
		return fmt.Errorf("目标队友不存在")
	}
	if target.Camp != user.Camp || target.ID == user.ID {
		return fmt.Errorf("灵魂链接只能指定队友")
	}
	if YellowSoul(user) < 1 || BlueSoul(user) < 1 {
		return fmt.Errorf("灵魂不足，无法放置灵魂链接")
	}
	if user.Character == nil {
		return fmt.Errorf("角色信息缺失")
	}

	linkCard, ok := user.ConsumeExclusiveCard(user.Character.ID, "灵魂链接")
	if !ok {
		return fmt.Errorf("未找到【灵魂链接】专属技能卡")
	}

	AddYellowSoul(user, -1)
	AddBlueSoul(user, -1)

	if err := rt.AttachEffectCard(user, target, model.EffectSoulLink, linkCard); err != nil {
		user.RestoreExclusiveCard(linkCard)
		AddYellowSoul(user, 1)
		AddBlueSoul(user, 1)
		return err
	}

	rt.Log(fmt.Sprintf("%s 发动 [灵魂链接]：移除1黄魂+1蓝魂，并将灵魂链接放置于 %s 面前", user.Name, target.Name))

	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		rt.ApplyChoiceResumePoint(model.TurnStageActionStart)
	}
	return nil
}

func handleSoulLinkTransferXChoice(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	sorcererID, _ := ctxData["sorcerer_id"].(string)
	sorcerer := rt.GetPlayers()[sorcererID]
	if sorcerer == nil {
		return fmt.Errorf("灵魂术士不存在")
	}

	damageIdx := runtimeutil.ToIntContextValue(ctxData["damage_index"])
	pdQueue := rt.GetPendingDamageQueue()
	if damageIdx < 0 || damageIdx >= len(pdQueue) {
		return fmt.Errorf("伤害上下文不存在")
	}

	maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
	if maxX < 0 {
		maxX = 0
	}

	x := selectionIndex
	if x < 0 || x > maxX {
		return fmt.Errorf("无效的X值")
	}

	if x > BlueSoul(sorcerer) {
		x = BlueSoul(sorcerer)
	}

	counterpartID, _ := ctxData["counterpart_id"].(string)
	sourceID, _ := ctxData["source_id"].(string)
	counterpart := rt.GetPlayers()[counterpartID]

	// Reduce the original damage by x before adding the transferred damage.
	pd := &pdQueue[damageIdx]
	if x > pd.Damage {
		x = pd.Damage
	}

	if x > 0 && counterpart != nil {
		AddBlueSoul(sorcerer, -x)

		pd.Damage -= x
		if pd.Damage < 0 {
			pd.Damage = 0
		}

		rt.AddPendingDamage(model.PendingDamage{
			SourceID:   sourceID,
			TargetID:   counterpart.ID,
			Damage:     x,
			DamageType: model.MagicAttack,
			Checks: map[model.PendingDamageCheckKey]bool{
				model.PendingDamageCheckFromSoulLink: true,
			},
		})

		rt.Log(fmt.Sprintf("%s 的 [灵魂链接] 生效：移除%d点蓝魂，将%d点伤害转移给 %s（法术伤害）", sorcerer.Name, x, x, counterpart.Name))
	} else {
		rt.Log(fmt.Sprintf("%s 的 [灵魂链接] 选择不转移伤害", sorcerer.Name))
	}

	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		rt.EnterDamageResolution(nil)
	}
	return nil
}

func handleSoulRecallPickChoice(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	if selectionIndex < 0 {
		return fmt.Errorf("请从可选法术牌中至少选择1张")
	}

	userID, _ := ctxData["user_id"].(string)
	if userID == "" {
		return fmt.Errorf("玩家ID缺失")
	}
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}

	magicIndices := runtimeutil.ParseChoiceIntSlice(ctxData["magic_indices"])
	if len(magicIndices) == 0 {
		magicIndices = runtimeutil.ParseChoiceIntSlice(ctxData["remaining_indices"])
	}

	allowed := make(map[int]struct{}, len(magicIndices))
	orderedCandidates := make([]int, 0, len(magicIndices))
	for _, idx := range magicIndices {
		if idx < 0 || idx >= len(user.Hand) {
			continue
		}
		if user.Hand[idx].Type != model.CardTypeMagic {
			continue
		}
		allowed[idx] = struct{}{}
		orderedCandidates = append(orderedCandidates, idx)
	}
	if len(allowed) == 0 {
		return fmt.Errorf("灵魂召还没有可弃置的法术牌")
	}

	resolvedIdx, ok := runtimeutil.ResolveSelectionToAllowedIndex(selectionIndex, orderedCandidates, allowed)
	if !ok {
		return fmt.Errorf("灵魂召还只能选择法术牌")
	}

	picked := []int{resolvedIdx}

	removed, err := removeCardsByIndicesFromHand(user, picked)
	if err != nil {
		return err
	}
	rt.NotifyCardRevealed(user.ID, removed, "discard")
	rt.AppendToDiscard(removed)

	gain := len(removed)
	before := BlueSoul(user)
	after := AddBlueSoul(user, gain)
	rt.Log(fmt.Sprintf("%s 发动 [灵魂召还]：弃置%d张法术牌，蓝色灵魂 +%d（%d→%d）", user.Name, gain, gain, before, after))

	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		if !rt.RoutePendingDamageWithReturn(model.TurnStageExtraAction) {
			rt.EnterExtraActionStage()
		}
	}
	return nil
}

// removeCardsByIndicesFromHand removes cards at the given indices from a player's hand,
// returning the removed cards. Indices must be valid and unique.
func removeCardsByIndicesFromHand(player *model.Player, indices []int) ([]model.Card, error) {
	if player == nil {
		return nil, fmt.Errorf("玩家不存在")
	}
	for _, idx := range indices {
		if idx < 0 || idx >= len(player.Hand) {
			return nil, fmt.Errorf("无效的手牌索引: %d", idx)
		}
	}
	seen := map[int]bool{}
	for _, idx := range indices {
		if seen[idx] {
			return nil, fmt.Errorf("不能重复选择同一张牌")
		}
		seen[idx] = true
	}
	// Sort indices descending to avoid shift during removal.
	for i := 0; i < len(indices); i++ {
		for j := i + 1; j < len(indices); j++ {
			if indices[i] < indices[j] {
				indices[i], indices[j] = indices[j], indices[i]
			}
		}
	}
	var removed []model.Card
	for _, idx := range indices {
		removed = append(removed, player.Hand[idx])
		player.Hand = append(player.Hand[:idx], player.Hand[idx+1:]...)
	}
	return removed, nil
}
