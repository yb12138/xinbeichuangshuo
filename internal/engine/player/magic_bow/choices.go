// gameflow: 魔弓手角色选择流。

package magic_bow

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

// ---------------------------------------------------------------------------
// BuildPrompt
// ---------------------------------------------------------------------------

func (choiceHandler) BuildPrompt(rt engineplayer.ChoiceRuntime, choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "mb_charge_draw_x":
		return buildChargeDrawXPrompt(playerID, data)
	case "mb_charge_place_count":
		return buildChargePlaceCountPrompt(playerID, data)
	case "mb_charge_place_cards", "mb_demon_eye_charge_card":
		return buildChargePlaceCardsPrompt(playerID, player, data, choiceType)
	case "mb_thunder_scatter_extra":
		return buildThunderScatterExtraPrompt(playerID, data)
	case "mb_demon_eye_target":
		return engineplayer.BuildTargetChoicePrompt(rt, playerID, "【魔眼】请选择弃1张牌的目标角色：", data, false)
	case "mb_multi_shot_target":
		return engineplayer.BuildTargetChoicePrompt(rt, playerID, "【多重射击】请选择暗系追加攻击目标：", data, false)
	case "mb_thunder_scatter_target":
		return engineplayer.BuildTargetChoicePrompt(rt, playerID, fmt.Sprintf("【雷光散射】请选择额外受到%d点法术伤害的目标：", runtimeutil.ToIntContextValue(data["extra_x"])), data, false)
	default:
		return nil
	}
}

func buildChargeDrawXPrompt(playerID string, data map[string]interface{}) *model.Prompt {
	maxDraw := runtimeutil.ToIntContextValue(data["max_draw"])
	if maxDraw <= 0 {
		maxDraw = 4
	}
	options := make([]model.PromptOption, 0, maxDraw+1)
	for x := 0; x <= maxDraw; x++ {
		options = append(options, model.PromptOption{
			ID:    fmt.Sprintf("%d", x),
			Label: fmt.Sprintf("X=%d（摸%d张）", x, x),
		})
	}
	return &model.Prompt{
		Type:         model.PromptConfirm,
		PlayerID:     playerID,
		Message:      "【充能】请选择摸牌数量X（0~4）：",
		Options:      options,
		Min:          1,
		Max:          1,
		Presentation: &model.PromptPresentation{Kind: model.PresentationNumeric, NumericBase: 0},
	}
}

func buildChargePlaceCountPrompt(playerID string, data map[string]interface{}) *model.Prompt {
	maxPlace := runtimeutil.ToIntContextValue(data["max_place"])
	if maxPlace < 0 {
		maxPlace = 0
	}
	options := make([]model.PromptOption, 0, maxPlace+1)
	for count := 0; count <= maxPlace; count++ {
		options = append(options, model.PromptOption{
			ID:    fmt.Sprintf("%d", count),
			Label: fmt.Sprintf("放置%d张充能", count),
		})
	}
	return &model.Prompt{
		Type:     model.PromptConfirm,
		PlayerID: playerID,
		Message:  "【充能】请选择要放置为充能的手牌数量：",
		Options:  options,
		Min:      1,
		Max:      1,
	}
}

func buildChargePlaceCardsPrompt(playerID string, player *model.Player, data map[string]interface{}, choiceType string) *model.Prompt {
	if player == nil {
		return nil
	}
	remaining := engineplayer.ParseIntSliceContextValue(data["remaining_indices"])
	if len(remaining) == 0 && choiceType == "mb_demon_eye_charge_card" {
		remaining = engineplayer.AllHandIndices(player)
	}
	selectedCount := len(engineplayer.ParseIntSliceContextValue(data["selected_indices"]))
	needCount := runtimeutil.ToIntContextValue(data["need_count"])
	if choiceType == "mb_demon_eye_charge_card" && needCount <= 0 {
		needCount = 1
	}
	if needCount <= 0 {
		needCount = 1
	}
	options := make([]model.PromptOption, 0, len(remaining))
	for _, idx := range remaining {
		if idx < 0 || idx >= len(player.Hand) {
			continue
		}
		options = append(options, model.PromptOption{
			ID:    fmt.Sprintf("%d", idx),
			Label: fmt.Sprintf("%d: %s", idx+1, promptfmt.FormatCardInfo(player.Hand[idx])),
		})
	}
	remainingPick := needCount - selectedCount
	if remainingPick < 1 {
		remainingPick = 1
	}
	if len(options) > 0 && remainingPick > len(options) {
		remainingPick = len(options)
	}
	message := fmt.Sprintf("【充能】请选择%d张作为充能的手牌：", remainingPick)
	if choiceType == "mb_demon_eye_charge_card" {
		message = "【魔眼】请选择1张手牌作为充能："
	}
	return &model.Prompt{
		Type:     model.PromptChooseCards,
		PlayerID: playerID,
		Message:  message,
		Options:  options,
		Min:      remainingPick,
		Max:      remainingPick,
	}
}

func buildThunderScatterExtraPrompt(playerID string, data map[string]interface{}) *model.Prompt {
	maxExtra := runtimeutil.ToIntContextValue(data["max_extra"])
	if maxExtra < 0 {
		maxExtra = 0
	}
	options := make([]model.PromptOption, 0, maxExtra+1)
	for x := 0; x <= maxExtra; x++ {
		options = append(options, model.PromptOption{
			ID:    fmt.Sprintf("%d", x),
			Label: fmt.Sprintf("额外移除%d个雷系充能", x),
		})
	}
	return &model.Prompt{
		Type:     model.PromptConfirm,
		PlayerID: playerID,
		Message:  "【雷光散射】请选择额外移除雷系充能数量X：",
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
	case "mb_charge_draw_x":
		return true, handleChargeDrawX(rt, ctxData, selectionIndex)
	case "mb_charge_place_count":
		return true, handleChargePlaceCount(rt, ctxData, selectionIndex)
	case "mb_charge_place_cards", "mb_demon_eye_charge_card":
		return true, handleChargePlaceCards(rt, ctxData, selectionIndex)
	case "mb_thunder_scatter_extra":
		return true, handleThunderScatterExtra(rt, ctxData, selectionIndex)
	case "mb_thunder_scatter_target", "mb_multi_shot_target", "mb_demon_eye_target":
		return true, handleTargetChoice(rt, ctxData, selectionIndex)
	default:
		return false, nil
	}
}

// ---------------------------------------------------------------------------
// Individual choice-type handlers
// ---------------------------------------------------------------------------

func handleChargeDrawX(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}

	maxDraw := runtimeutil.ToIntContextValue(ctxData["max_draw"])
	if maxDraw <= 0 {
		maxDraw = 4
	}
	xValue := selectionIndex
	if xValue < 0 || xValue > maxDraw {
		return fmt.Errorf("无效的X值")
	}

	if xValue > 0 {
		rt.DrawCards(user.ID, xValue)
	}

	room := ChargeCap - ChargeCount(user, "")
	maxPlace := xValue
	if maxPlace > len(user.Hand) {
		maxPlace = len(user.Hand)
	}
	if maxPlace > room {
		maxPlace = room
	}

	if maxPlace > 0 {
		overflow := len(user.Hand) - rt.GetMaxHand(user)
		if overflow > 0 {
			rt.ApplyCampMoraleLoss(user.Camp, overflow)
			rt.Log(fmt.Sprintf("%s 的 [充能] 摸牌后超出手牌上限%d：士气-%d（本次不弃牌）", user.Name, overflow, overflow))
		}
	}

	if maxPlace <= 0 {
		rt.Log(fmt.Sprintf("%s 的 [充能] 生效：摸%d张，不放置充能", user.Name, xValue))
		rt.PopInterrupt()
		if rt.GetPendingInterrupt() == nil {
			rt.ApplyChoiceResumePoint(model.TurnStageActionStart)
		}
		return nil
	}

	ctxData["choice_type"] = "mb_charge_place_count"
	ctxData["max_place"] = maxPlace
	if intr := rt.GetPendingInterrupt(); intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
	rt.Log(fmt.Sprintf("%s 的 [充能] 生效：摸%d张，可放置最多%d张充能", user.Name, xValue, maxPlace))
	return nil
}

func handleChargePlaceCount(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}

	maxPlace := runtimeutil.ToIntContextValue(ctxData["max_place"])
	if maxPlace < 0 {
		maxPlace = 0
	}
	needCount := selectionIndex
	if needCount < 0 || needCount > maxPlace {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	if needCount == 0 {
		rt.Log(fmt.Sprintf("%s 选择不放置充能", user.Name))
		rt.PopInterrupt()
		if rt.GetPendingInterrupt() == nil {
			rt.ApplyChoiceResumePoint(model.TurnStageActionStart)
		}
		return nil
	}

	ctxData["choice_type"] = "mb_charge_place_cards"
	ctxData["need_count"] = needCount
	ctxData["selected_indices"] = []int{}
	ctxData["remaining_indices"] = engineplayer.AllHandIndices(user)
	if intr := rt.GetPendingInterrupt(); intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
	return nil
}

func handleChargePlaceCards(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	choiceType, _ := ctxData["choice_type"].(string)
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}

	remaining := engineplayer.ParseIntSliceContextValue(ctxData["remaining_indices"])
	if len(remaining) == 0 {
		remaining = engineplayer.AllHandIndices(user)
	}
	selected := engineplayer.ParseIntSliceContextValue(ctxData["selected_indices"])
	needCount := runtimeutil.ToIntContextValue(ctxData["need_count"])
	if choiceType == "mb_demon_eye_charge_card" && needCount <= 0 {
		needCount = 1
	}
	if needCount <= 0 {
		needCount = 1
	}

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

	if len(selected) < needCount {
		ctxData["selected_indices"] = selected
		ctxData["remaining_indices"] = nextRemaining
		if intr := rt.GetPendingInterrupt(); intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return nil
	}

	removed, _ := engineplayer.RemoveCardsByIndicesFromHand(user, append([]int{}, selected...))

	// Calculate how many can actually be placed (cap = ChargeCap).
	room := ChargeCap - ChargeCount(user, "")
	toPlace := len(removed)
	if toPlace > room {
		toPlace = room
	}
	var toDiscard []model.Card
	if toPlace < len(removed) {
		toDiscard = removed[toPlace:]
	}
	AddChargeCards(user, removed[:toPlace])
	if len(toDiscard) > 0 {
		rt.AppendToDiscard(toDiscard)
	}

	if choiceType == "mb_demon_eye_charge_card" {
		maxEnergy := getPlayerEnergyCap(user)
		if user.Gem+user.Crystal < maxEnergy {
			user.Crystal++
			if user.Gem+user.Crystal > maxEnergy {
				user.Crystal -= user.Gem + user.Crystal - maxEnergy
			}
		}
		rt.Log(fmt.Sprintf("%s 的 [魔眼] 生效：放置1张充能并获得1点蓝水晶", user.Name))
	} else {
		rt.Log(fmt.Sprintf("%s 的 [充能] 生效：放置%d张充能", user.Name, toPlace))
	}

	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		rt.ApplyChoiceResumePoint(model.TurnStageActionStart)
	}
	return nil
}

func handleThunderScatterExtra(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}

	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
	if len(targetIDs) == 0 {
		return fmt.Errorf("雷光散射没有可选目标")
	}

	maxExtra := runtimeutil.ToIntContextValue(ctxData["max_extra"])
	extraX := selectionIndex
	if extraX < 0 || extraX > maxExtra {
		return fmt.Errorf("无效的X值")
	}

	for _, targetID := range targetIDs {
		rt.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   targetID,
			Damage:     1,
			DamageType: model.MagicAttack,
		})
	}

	actualExtra := 0
	for i := 0; i < extraX; i++ {
		if _, ok := RemoveChargeByElement(user, model.ElementThunder); !ok {
			break
		}
		actualExtra++
	}

	if actualExtra <= 0 {
		rt.Log(fmt.Sprintf("%s 的 [雷光散射] 生效：对所有对手各造成1点法术伤害", user.Name))
		rt.PopInterrupt()
		if rt.GetPendingInterrupt() == nil {
			rt.ApplyChoiceResumePoint(model.TurnStageExtraAction)
			if len(rt.GetPendingDamageQueue()) > 0 {
				rt.EnterDamageResolution(nil)
			}
		}
		return nil
	}

	lockedTargetID, _ := ctxData["locked_target_id"].(string)
	if lockedTargetID != "" {
		lockedValid := false
		for _, targetID := range targetIDs {
			if targetID == lockedTargetID {
				lockedValid = true
				break
			}
		}
		if !lockedValid {
			return fmt.Errorf("雷光散射预选目标无效")
		}
		target := rt.GetPlayers()[lockedTargetID]
		if target == nil {
			return fmt.Errorf("目标不存在")
		}
		rt.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   lockedTargetID,
			Damage:     actualExtra,
			DamageType: model.MagicAttack,
		})
		rt.Log(fmt.Sprintf("%s 的 [雷光散射] 生效：对所有对手各1点，并对 %s 额外造成%d点法术伤害", user.Name, target.Name, actualExtra))
		rt.PopInterrupt()
		if rt.GetPendingInterrupt() == nil {
			rt.ApplyChoiceResumePoint(model.TurnStageExtraAction)
			if len(rt.GetPendingDamageQueue()) > 0 {
				rt.EnterDamageResolution(nil)
			}
		}
		return nil
	}

	ctxData["choice_type"] = "mb_thunder_scatter_target"
	ctxData["extra_x"] = actualExtra
	if intr := rt.GetPendingInterrupt(); intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
	return nil
}

func handleTargetChoice(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	choiceType, _ := ctxData["choice_type"].(string)
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

	switch choiceType {
	case "mb_thunder_scatter_target":
		extraX := runtimeutil.ToIntContextValue(ctxData["extra_x"])
		if extraX > 0 {
			rt.AddPendingDamage(model.PendingDamage{
				SourceID:   user.ID,
				TargetID:   targetID,
				Damage:     extraX,
				DamageType: model.MagicAttack,
			})
		}
		rt.Log(fmt.Sprintf("%s 的 [雷光散射] 生效：对所有对手各1点，并对 %s 额外造成%d点法术伤害", user.Name, target.Name, extraX))
		rt.PopInterrupt()
		if rt.GetPendingInterrupt() == nil {
			rt.ApplyChoiceResumePoint(model.TurnStageExtraAction)
			if len(rt.GetPendingDamageQueue()) > 0 {
				rt.EnterDamageResolution(nil)
			}
		}
		return nil

	case "mb_multi_shot_target":
		prevOrder := 0
		if user.TurnState.SkillFlowState != nil {
			prevOrder = user.TurnState.SkillFlowState["mb_last_attack_target_order"]
		}
		if prevOrder > 0 {
			playerOrder := rt.GetPlayerOrder()
			for idx, pid := range playerOrder {
				if idx+1 == prevOrder && pid == targetID {
					return fmt.Errorf("多重射击不能选择上次攻击目标")
				}
			}
		}
		virtualCard := model.Card{
			ID:          fmt.Sprintf("mb_multi_shot_%s_%d", user.ID, len(user.Hand)+1),
			Name:        "多重射击",
			Type:        model.CardTypeAttack,
			Element:     model.ElementDark,
			Damage:      1,
			Description: "由多重射击视为的暗系主动攻击（伤害-1）",
		}
		rt.EnqueueVirtualAttack(user.ID, target.ID, virtualCard, "mb_multi_shot")
		RemoveChargeByElement(user, model.ElementWind)
		rt.Log(fmt.Sprintf("%s 的 [多重射击] 生效：对 %s 发起1次暗系追加攻击（伤害-1）", user.Name, target.Name))
		rt.PopInterrupt()
		if rt.GetPendingInterrupt() == nil {
			rt.EnterActionExecutionStage()
		}
		return nil

	case "mb_demon_eye_target":
		if targetID == user.ID {
			return fmt.Errorf("魔眼不能以自己为目标")
		}
		if len(target.Hand) > 0 {
			// Target must discard 1 card
			discardCtx := map[string]interface{}{
				"choice_type":            "system_discard_cards",
				"discard_subflow":        true,
				"discard_count":          1,
				"prompt":                 "【魔眼】请选择弃置1张手牌：",
				"mb_demon_eye_user_id":   user.ID,
				"mb_demon_eye_target_id": targetID,
			}
			rt.PushInterrupt(&model.Interrupt{
				Type:     model.InterruptChoice,
				PlayerID: targetID,
				Context:  discardCtx,
			})
			rt.Log(fmt.Sprintf("%s 的 [魔眼] 生效：请选择 %s 弃置1张手牌", user.Name, target.Name))
			rt.NotifyInterruptPrompt()
			return nil
		}
		// Target has no cards: user draws 3 and places 1 as charge
		rt.DrawCards(user.ID, 3)
		ctxData["choice_type"] = "mb_demon_eye_charge_card"
		ctxData["need_count"] = 1
		ctxData["selected_indices"] = []int{}
		ctxData["remaining_indices"] = engineplayer.AllHandIndices(user)
		if intr := rt.GetPendingInterrupt(); intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		rt.Log(fmt.Sprintf("%s 的 [魔眼] 生效：%s 无法弃牌，改为自己摸3张牌并选择1张作为充能", user.Name, target.Name))
		return nil
	}

	return nil
}

// ---------------------------------------------------------------------------
// AfterDiscard FlowContinuation: 魔眼弃牌后续
// ---------------------------------------------------------------------------

func handleMagicBowAfterDiscard(rt engineplayer.ChoiceRuntime, cont model.FlowContinuation) error {
	if cont.SkillID == "mb_charge" {
		return continueChargeAfterDiscard(rt, cont.PlayerID)
	}
	if cont.SkillID == "mb_demon_eye" {
		discardPlayerID, _ := cont.Data["discard_player_id"].(string)
		if discardPlayerID == "" {
			discardPlayerID = cont.PlayerID
		}
		if discardPlayer := rt.GetPlayers()[discardPlayerID]; discardPlayer != nil {
			demonEyeAfterDiscardData(rt, discardPlayer, cont.Data)
		}
	}
	return nil
}

func continueChargeAfterDiscard(rt engineplayer.ChoiceRuntime, userID string) error {
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("充能后续执行者不存在: %s", userID)
	}
	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: user.ID,
		Context: map[string]interface{}{
			"choice_type": "mb_charge_draw_x",
			"user_id":     user.ID,
			"max_draw":    4,
		},
	})
	rt.Log(fmt.Sprintf("%s 的 [充能] 已弃至4张，请选择摸牌数量X（0~4）", user.Name))
	return nil
}

// demonEyeAfterDiscard 处理魔眼目标弃牌后的后续流程。
// 当弃牌中断的 context 中包含 mb_demon_eye_user_id 时：
//   - 魔弓无手牌 → 获得1点蓝水晶，恢复行动阶段
//   - 魔弓有手牌 → 推入 mb_demon_eye_charge_card 选择中断
func demonEyeAfterDiscard(rt engineplayer.ChoiceRuntime, discardPlayer *model.Player) bool {
	pending := rt.GetPendingInterrupt()
	if pending == nil {
		return false
	}
	data, ok := pending.Context.(map[string]interface{})
	if !ok {
		return false
	}
	return demonEyeAfterDiscardData(rt, discardPlayer, data)
}

func demonEyeAfterDiscardData(rt engineplayer.ChoiceRuntime, discardPlayer *model.Player, data map[string]any) bool {
	if data == nil {
		return false
	}
	demonEyeUserID, _ := data["mb_demon_eye_user_id"].(string)
	if demonEyeUserID == "" {
		return false
	}
	user := rt.GetPlayers()[demonEyeUserID]
	if user == nil {
		return false
	}

	if len(user.Hand) == 0 {
		// 无手牌：获得1点蓝水晶
		maxEnergy := rt.GetPlayerEnergyCap(user)
		if user.Gem+user.Crystal < maxEnergy {
			user.Crystal++
			if user.Gem+user.Crystal > maxEnergy {
				user.Crystal -= user.Gem + user.Crystal - maxEnergy
			}
		}
		rt.Log(fmt.Sprintf("%s 的 [魔眼] 生效：已完成目标弃牌，但自己无手牌可充能，改为仅获得1点蓝水晶", user.Name))
		if rt.GetPendingInterrupt() == nil {
			rt.ApplyChoiceResumePoint(model.TurnStageActionStart)
		}
		return true
	}

	// 有手牌：推入充能选择中断
	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: demonEyeUserID,
		Context: map[string]interface{}{
			"choice_type":       "mb_demon_eye_charge_card",
			"user_id":           demonEyeUserID,
			"need_count":        1,
			"selected_indices":  []int{},
			"remaining_indices": engineplayer.AllHandIndices(user),
		},
	})
	rt.Log(fmt.Sprintf("%s 的 [魔眼] 生效：%s 已弃置1张手牌，请选择1张手牌作为充能", user.Name, discardPlayer.Name))
	return true
}

// ---------------------------------------------------------------------------
// Local helpers
// ---------------------------------------------------------------------------


// getPlayerEnergyCap returns the energy capacity for a player (default 3).
func getPlayerEnergyCap(player *model.Player) int {
	if player == nil {
		return 3
	}
	return 3
}
