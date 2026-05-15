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
	case "mb_magic_pierce_charge":
		return buildChargeRemovalPrompt(playerID, player, model.ElementFire, "【魔贯冲击】请选择移除1个火系充能：")
	case "mb_magic_pierce_hit_bonus":
		return buildMagicPierceHitBonusPrompt(playerID)
	case "mb_magic_pierce_hit_charge":
		return buildChargeRemovalPrompt(playerID, player, model.ElementFire, "【魔贯冲击】请选择额外移除1个火系充能：")
	case "mb_thunder_scatter_base_charge":
		return buildChargeRemovalPrompt(playerID, player, model.ElementThunder, "【雷光散射】请选择移除1个雷系充能：")
	case "mb_multi_shot_charge":
		return buildChargeRemovalPrompt(playerID, player, model.ElementWind, "【多重射击】请选择移除1个风系充能：")
	case "mb_charge_draw_x":
		return buildChargeDrawXPrompt(playerID, data)
	case "mb_charge_place_count":
		return buildChargePlaceCountPrompt(playerID, data)
	case "mb_charge_place_cards", "mb_demon_eye_charge_card":
		return buildChargePlaceCardsPrompt(playerID, player, data, choiceType)
	case "mb_thunder_scatter_extra":
		return buildThunderScatterExtraPrompt(playerID, data)
	case "mb_demon_eye_mode":
		return &model.Prompt{
			Type:         model.PromptConfirm,
			PlayerID:     playerID,
			Message:      "【魔眼】请选择发动分支：",
			Options:      []model.PromptOption{{ID: "0", Label: "分支①：令1名角色弃1张牌"}, {ID: "1", Label: "分支②：你摸3张牌"}},
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationBranchSelect, Layout: "overlay"},
		}
	case "mb_demon_eye_target":
		return engineplayer.BuildTargetChoicePrompt(rt, playerID, "【魔眼·分支①】请选择弃1张牌的目标角色：", data, false)
	case "mb_multi_shot_target":
		p := engineplayer.BuildTargetChoicePrompt(rt, playerID, "【多重射击】请选择暗系追加攻击目标：", data, false)
		if p != nil {
			p.ChoiceType = "mb_multi_shot_target"
		}
		return p
	case "mb_thunder_scatter_target":
		p := engineplayer.BuildTargetChoicePrompt(rt, playerID, fmt.Sprintf("【雷光散射】请选择额外受到%d点法术伤害的目标：", runtimeutil.ToIntContextValue(data["extra_x"])), data, false)
		if p != nil {
			p.ChoiceType = "mb_thunder_scatter_target"
		}
		return p
	default:
		return nil
	}
}

func buildChargeRemovalPrompt(playerID string, player *model.Player, element model.Element, message string) *model.Prompt {
	if player == nil {
		return nil
	}
	fieldIndices := magicBowChargeFieldIndices(player, element)
	options := make([]model.PromptOption, 0, len(fieldIndices))
	for _, fieldIndex := range fieldIndices {
		if fieldIndex < 0 || fieldIndex >= len(player.Field) {
			continue
		}
		fc := player.Field[fieldIndex]
		if fc == nil || fc.Card.ID == "" {
			continue
		}
		eleZh := promptfmt.ElementName(string(fc.Card.Element))
		if eleZh == "" {
			eleZh = string(fc.Card.Element)
		}
		options = append(options, model.PromptOption{
			ID:    fmt.Sprintf("%d", fieldIndex),
			Label: fmt.Sprintf("%s（%s系）", fc.Card.Name, eleZh),
		})
	}
	return &model.Prompt{
		Type:     model.PromptConfirm,
		PlayerID: playerID,
		Message:  message,
		Options:  options,
		Min:      1,
		Max:      1,
	}
}

func buildMagicPierceHitBonusPrompt(playerID string) *model.Prompt {
	return &model.Prompt{
		Type:     model.PromptConfirm,
		PlayerID: playerID,
		Message:  "【魔贯冲击】是否额外移除1个火系充能使伤害+1？",
		Options: []model.PromptOption{
			{ID: "0", Label: "是"},
			{ID: "1", Label: "否"},
		},
		Min: 1,
		Max: 1,
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
		Type:         model.PromptConfirm,
		PlayerID:     playerID,
		Message:      "【充能】请选择要放置为充能的手牌数量：",
		Options:      options,
		Min:          1,
		Max:          1,
		Presentation: &model.PromptPresentation{Kind: model.PresentationNumeric, NumericBase: 0},
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
	maxPlace := runtimeutil.ToIntContextValue(data["max_place"])
	if choiceType == "mb_demon_eye_charge_card" && needCount <= 0 {
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

	// 充能盖牌：允许选择 0~maxPlace 张（needCount=0 表示不强制数量）
	// 魔眼充能：必须选择 1 张（needCount=1）
	if choiceType == "mb_charge_place_cards" && needCount == 0 {
		// 多选模式：Min=0（可不选），Max=maxPlace
		minPick := 0
		maxPick := maxPlace
		if maxPick > len(options) {
			maxPick = len(options)
		}
		return &model.Prompt{
			Type:     model.PromptChooseCards,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【充能】请选择要放置为充能的手牌（最多%d张，可不选）：", maxPick),
			Options:  options,
			Min:      minPick,
			Max:      maxPick,
		}
	}

	// 原逻辑：逐张选择（needCount > 0）
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
		Type:         model.PromptConfirm,
		PlayerID:     playerID,
		Message:      "【雷光散射】请选择额外移除雷系充能数量X：",
		Options:      options,
		Min:          1,
		Max:          1,
		Presentation: &model.PromptPresentation{Kind: model.PresentationNumeric, NumericBase: 0},
	}
}

// ---------------------------------------------------------------------------
// HandleChoice
// ---------------------------------------------------------------------------

func (choiceHandler) HandleChoice(rt engineplayer.ChoiceRuntime, _ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)

	switch choiceType {
	case "mb_magic_pierce_charge":
		return true, handleMagicPierceCharge(rt, ctxData, selectionIndex)
	case "mb_magic_pierce_hit_bonus":
		return true, handleMagicPierceHitBonus(rt, ctxData, selectionIndex)
	case "mb_magic_pierce_hit_charge":
		return true, handleMagicPierceHitCharge(rt, ctxData, selectionIndex)
	case "mb_thunder_scatter_base_charge":
		return true, handleThunderScatterBaseCharge(rt, ctxData, selectionIndex)
	case "mb_multi_shot_charge":
		return true, handleMultiShotCharge(rt, ctxData, selectionIndex)
	case "mb_charge_draw_x":
		return true, handleChargeDrawX(rt, ctxData, selectionIndex)
	case "mb_charge_place_count":
		return true, handleChargePlaceCount(rt, ctxData, selectionIndex)
	case "mb_charge_place_cards", "mb_demon_eye_charge_card":
		return true, handleChargePlaceCards(rt, ctxData, selectionIndex)
	case "mb_demon_eye_mode":
		return true, handleDemonEyeMode(rt, ctxData, selectionIndex)
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

func removeSelectedMagicBowCharge(user *model.Player, element model.Element, selectionIndex int) (model.Card, error) {
	fieldIndices := magicBowChargeFieldIndices(user, element)
	if selectionIndex < 0 || selectionIndex >= len(fieldIndices) {
		return model.Card{}, fmt.Errorf("无效的充能选项: %d", selectionIndex)
	}
	card, ok := removeMagicBowChargeAtFieldIndex(user, fieldIndices[selectionIndex], element)
	if !ok {
		return model.Card{}, fmt.Errorf("所选充能不存在或元素不匹配")
	}
	return card, nil
}

func addMagicPierceAttackDamageBonus(rt engineplayer.ChoiceRuntime, userID string) bool {
	queue := rt.GetPendingDamageQueue()
	for i := range queue {
		queued, ok := rt.GetPendingDamageByIndex(i)
		if !ok || queued == nil {
			continue
		}
		if queued.SourceID != userID || queued.DamageType != model.AttackDamage {
			continue
		}
		queued.Damage++
		return true
	}
	return false
}

func handleMagicPierceCharge(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	rawCtx, _ := ctxData["user_ctx"].(*model.Context)
	userID, _ := ctxData["user_id"].(string)
	if userID == "" && rawCtx != nil && rawCtx.User != nil {
		userID = rawCtx.User.ID
	}
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	removed, err := removeSelectedMagicBowCharge(user, model.ElementFire, selectionIndex)
	if err != nil {
		return err
	}
	if rawCtx == nil || rawCtx.EventCtx == nil || rawCtx.EventCtx.Card == nil {
		return fmt.Errorf("魔贯冲击上下文无效")
	}
	user.TurnState.UsedSkillCounts["mb_magic_pierce_used_turn"]++
	engineplayer.SetSkillFlowState(user, "mb_magic_pierce_pending", 1)
	rawCtx.EventCtx.Card.Damage++
	rt.Log(fmt.Sprintf("%s 发动 [魔贯冲击]：移除火系充能 %s，本次攻击伤害+1", user.Name, removed.Name))
	rt.PopInterrupt()
	return nil
}

func handleMagicPierceHitBonus(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	switch selectionIndex {
	case 0:
		if magicBowChargeCount(user, model.ElementFire) <= 0 {
			user.TurnState.SkillFlowState["mb_magic_pierce_pending"] = 0
			rt.PopInterrupt()
			if rt.GetPendingInterrupt() == nil && len(rt.GetPendingDamageQueue()) > 0 {
				rt.EnterDamageResolution(nil)
			}
			return nil
		}
		ctxData["choice_type"] = "mb_magic_pierce_hit_charge"
		if intr := rt.GetPendingInterrupt(); intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return nil
	case 1:
		user.TurnState.SkillFlowState["mb_magic_pierce_pending"] = 0
		rt.Log(fmt.Sprintf("%s 放弃 [魔贯冲击] 命中追加充能", user.Name))
		rt.PopInterrupt()
		if rt.GetPendingInterrupt() == nil && len(rt.GetPendingDamageQueue()) > 0 {
			rt.EnterDamageResolution(nil)
		}
		return nil
	default:
		return fmt.Errorf("无效的魔贯冲击命中追加选项")
	}
}

func handleMagicPierceHitCharge(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	removed, err := removeSelectedMagicBowCharge(user, model.ElementFire, selectionIndex)
	if err != nil {
		return err
	}
	applied := addMagicPierceAttackDamageBonus(rt, user.ID)
	rt.Log(fmt.Sprintf("%s 的 [魔贯冲击] 命中追加生效：额外移除火系充能 %s，本次攻击伤害+1", user.Name, removed.Name))
	if !applied {
		rt.Log("[Warn] 魔弓冲击命中追加未找到对应伤害条目，未能叠加伤害")
	}
	user.TurnState.SkillFlowState["mb_magic_pierce_pending"] = 0
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil && len(rt.GetPendingDamageQueue()) > 0 {
		rt.EnterDamageResolution(nil)
	}
	return nil
}

func handleThunderScatterBaseCharge(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	removed, err := removeSelectedMagicBowCharge(user, model.ElementThunder, selectionIndex)
	if err != nil {
		return err
	}

	maxExtra := magicBowChargeCount(user, model.ElementThunder)
	if maxExtra > 0 {
		ctxData["choice_type"] = "mb_thunder_scatter_extra"
		ctxData["max_extra"] = maxExtra
		if intr := rt.GetPendingInterrupt(); intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		rt.Log(fmt.Sprintf("%s 的 [雷光散射]：已移除雷系充能 %s，可额外移除0~%d个雷系充能", user.Name, removed.Name, maxExtra))
		return nil
	}

	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
	for _, targetID := range targetIDs {
		rt.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   targetID,
			Damage:     1,
			DamageType: model.MagicAttack,
		})
	}
	rt.Log(fmt.Sprintf("%s 的 [雷光散射] 生效：移除雷系充能 %s，对所有对手各造成1点法术伤害", user.Name, removed.Name))
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		rt.ApplyChoiceResumePoint(model.TurnStageExtraAction)
		if len(rt.GetPendingDamageQueue()) > 0 {
			rt.EnterDamageResolution(nil)
		}
	}
	return nil
}

func handleMultiShotCharge(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	removed, err := removeSelectedMagicBowCharge(user, model.ElementWind, selectionIndex)
	if err != nil {
		return err
	}
	user.TurnState.UsedSkillCounts["mb_multi_shot_used_turn"]++
	ctxData["choice_type"] = "mb_multi_shot_target"
	if intr := rt.GetPendingInterrupt(); intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
	rt.Log(fmt.Sprintf("%s 的 [多重射击]：移除风系充能 %s，请选择暗系追加攻击目标", user.Name, removed.Name))
	return nil
}

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
		rt.DrawCardsWithOptions(user.ID, xValue, model.DrawOptions{PreventOverflow: true})
	}

	room := ChargeCap - ChargeCount(user, "")
	maxPlace := xValue
	if maxPlace > len(user.Hand) {
		maxPlace = len(user.Hand)
	}
	if maxPlace > room {
		maxPlace = room
	}

	// 检查手牌溢出，无论是否进入盖牌流程
	overflow := len(user.Hand) - rt.GetMaxHand(user)
	if overflow > 0 {
		rt.ApplyCampMoraleLoss(user.Camp, overflow)
		rt.Log(fmt.Sprintf("%s 的 [充能] 摸牌后超出手牌上限%d：士气-%d（本次不弃牌）", user.Name, overflow, overflow))
	}

	if maxPlace <= 0 {
		// 充能上限满了或其他原因导致不能盖牌
		if room <= 0 {
			rt.Log(fmt.Sprintf("%s 的 [充能] 生效：摸%d张，充能上限已满（%d/%d），不放置充能", user.Name, xValue, ChargeCount(user, ""), ChargeCap))
		} else {
			rt.Log(fmt.Sprintf("%s 的 [充能] 生效：摸%d张，不放置充能", user.Name, xValue))
		}
		rt.PopInterrupt()
		if rt.GetPendingInterrupt() == nil {
			rt.ApplyChoiceResumePoint(model.TurnStageActionStart)
		}
		return nil
	}

	// 直接进入盖牌选择（跳过数量选择步骤）
	// Min=0 允许不选，Max=maxPlace 最多可选maxPlace张
	ctxData["choice_type"] = "mb_charge_place_cards"
	ctxData["max_place"] = maxPlace
	ctxData["need_count"] = 0 // 0表示不强制数量，由玩家决定
	ctxData["selected_indices"] = []int{}
	ctxData["remaining_indices"] = engineplayer.AllHandIndices(user)
	if intr := rt.GetPendingInterrupt(); intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
	rt.Log(fmt.Sprintf("%s 的 [充能] 生效：摸%d张，请选择要放置为充能的手牌（最多%d张）", user.Name, xValue, maxPlace))
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

func handleDemonEyeMode(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("魔弓不存在")
	}

	switch selectionIndex {
	case 0:
		// Branch 1: Target selection - select any character to discard 1 card
		targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
		if len(targetIDs) == 0 {
			return fmt.Errorf("魔眼分支①没有可选目标")
		}
		ctxData["choice_type"] = "mb_demon_eye_target"
		if intr := rt.GetPendingInterrupt(); intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return nil
	case 1:
		// Branch 2: Draw 3 cards
		rt.DrawCards(user.ID, 3)
		rt.Log(fmt.Sprintf("%s 的 [魔眼] 分支②生效：摸3张牌", user.Name))
		if len(user.Hand) == 0 {
			// No cards to charge: just grant +1 crystal
			maxEnergy := getPlayerEnergyCap(user)
			if user.Gem+user.Crystal < maxEnergy {
				user.Crystal++
			}
			rt.Log(fmt.Sprintf("%s 的 [魔眼]：无手牌可充能，改为仅获得1点蓝水晶", user.Name))
			rt.PopInterrupt()
			if rt.GetPendingInterrupt() == nil {
				rt.ApplyChoiceResumePoint(model.TurnStageActionStart)
			}
			return nil
		}
		// Push charge card selection
		ctxData["choice_type"] = "mb_demon_eye_charge_card"
		ctxData["need_count"] = 1
		ctxData["selected_indices"] = []int{}
		ctxData["remaining_indices"] = engineplayer.AllHandIndices(user)
		if intr := rt.GetPendingInterrupt(); intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		rt.Log(fmt.Sprintf("%s 的 [魔眼] 分支②：请选择1张手牌作为充能", user.Name))
		return nil
	default:
		return fmt.Errorf("无效的魔眼分支选择")
	}
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
		if user.TurnState.UsedSkillCounts != nil {
			prevOrder = user.TurnState.UsedSkillCounts["mb_last_attack_target_order"]
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
		rt.Log(fmt.Sprintf("%s 的 [多重射击] 生效：对 %s 发起1次暗系追加攻击（伤害-1）", user.Name, target.Name))
		rt.PopInterrupt()
		if rt.GetPendingInterrupt() == nil {
			rt.EnterActionExecutionStage()
		}
		return nil

	case "mb_demon_eye_target":
		rt.PopInterrupt()
		if len(target.Hand) > 0 {
			// Target must discard 1 card
			discardCtx := map[string]interface{}{
				"choice_type":                 "system_discard_cards",
				"discard_subflow":             true,
				"discard_count":               1,
				"prompt":                      "【魔眼】请选择弃置1张手牌：",
				"flow_continuation_role_id":   "magic_bow",
				"flow_continuation_player_id": user.ID,
				"flow_continuation_skill_id":  "mb_demon_eye",
				"mb_demon_eye_user_id":        user.ID,
				"mb_demon_eye_target_id":      targetID,
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
		rt.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: user.ID,
			Context:  ctxData,
		})
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

// handleChargePlaceCardsMultiSelect 批量多选盖牌（充能技能简化交互）。
// 玩家一次性选择要盖放的手牌（0~maxPlace张），无需先选择数量再逐张选择。
func handleChargePlaceCardsMultiSelect(rt engineplayer.ChoiceRuntime, playerID string, selections []int, ctxData map[string]interface{}) (bool, error) {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return false, fmt.Errorf("玩家不存在")
	}

	maxPlace := runtimeutil.ToIntContextValue(ctxData["max_place"])
	if maxPlace <= 0 {
		maxPlace = len(user.Hand)
	}

	// 验证选择数量不超过上限
	if len(selections) > maxPlace {
		return false, fmt.Errorf("选择数量超过上限: 最多%d张", maxPlace)
	}

	// 验证所有选择索引有效且不重复
	validIndices := engineplayer.AllHandIndices(user)
	seen := make(map[int]bool)
	cardIndices := make([]int, 0, len(selections))
	for _, sel := range selections {
		// sel 是前端传来的选项索引，需要映射到实际手牌索引
		cardIdx, ok := runtimeutil.ResolveSelectionToCandidate(sel, validIndices)
		if !ok || cardIdx < 0 || cardIdx >= len(user.Hand) {
			return false, fmt.Errorf("无效的选项索引: %d", sel)
		}
		if seen[cardIdx] {
			return false, fmt.Errorf("不能重复选择同一张牌")
		}
		seen[cardIdx] = true
		cardIndices = append(cardIndices, cardIdx)
	}

	// 如果没有选择任何牌，直接结束
	if len(cardIndices) == 0 {
		rt.Log(fmt.Sprintf("%s 选择不放置充能", user.Name))
		rt.PopInterrupt()
		if rt.GetPendingInterrupt() == nil {
			rt.ApplyChoiceResumePoint(model.TurnStageActionStart)
		}
		return true, nil
	}

	// 从手牌中移除选中的牌
	removed, err := engineplayer.RemoveCardsByIndicesFromHand(user, cardIndices)
	if err != nil {
		return false, fmt.Errorf("移除手牌失败: %v", err)
	}

	// 计算实际可以放置的充能数量（考虑上限）
	room := ChargeCap - ChargeCount(user, "")
	toPlace := len(removed)
	if toPlace > room {
		toPlace = room
	}

	// 放置充能
	var toDiscard []model.Card
	if toPlace < len(removed) {
		toDiscard = removed[toPlace:]
	}
	AddChargeCards(user, removed[:toPlace])
	if len(toDiscard) > 0 {
		rt.AppendToDiscard(toDiscard)
	}

	rt.Log(fmt.Sprintf("%s 的 [充能] 生效：放置%d张充能", user.Name, toPlace))
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		rt.ApplyChoiceResumePoint(model.TurnStageActionStart)
	}
	return true, nil
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
