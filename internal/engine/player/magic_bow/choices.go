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

const (
	magicBowChargePlaceFlowID   = "mb_charge_place"
	magicBowChargeNeedStep      = "need"
	magicBowChargeCardsStep     = "cards"
	magicBowDemonEyeFlowID      = "mb_demon_eye_charge"
	magicBowDemonEyeCardsStep   = "cards"
	magicBowSequentialFlowLabel = "魔弓选择"
)

var (
	magicBowChargePlaceFlowRuntime = model.MustNewPromptFlowRuntime(magicBowChargePlaceFlowID, []model.PromptFlowStepSpec{
		{ID: magicBowChargeCardsStep, ChoiceType: "mb_charge_place_cards", CancelPolicy: model.CancelPolicyAbort},
	})
	magicBowDemonEyeFlowRuntime = model.MustNewPromptFlowRuntime(magicBowDemonEyeFlowID, []model.PromptFlowStepSpec{
		{ID: magicBowDemonEyeCardsStep, ChoiceType: "mb_demon_eye_charge_card", CancelPolicy: model.CancelPolicyAbort},
	})
)

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
		return engineplayer.BuildTargetChoicePrompt(rt, choiceType, playerID, "【魔眼·分支①】请选择弃1张牌的目标角色：", data, false)
	case "mb_multi_shot_target":
		return engineplayer.BuildTargetChoicePrompt(rt, choiceType, playerID, "【多重射击】请选择暗系追加攻击目标：", data, false)
	case "mb_thunder_scatter_target":
		return engineplayer.BuildTargetChoicePrompt(rt, choiceType, playerID, fmt.Sprintf("【雷光散射】请选择额外受到%d点法术伤害的目标：", runtimeutil.ToIntContextValue(data["extra_x"])), data, false)
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
	for seq, fieldIndex := range fieldIndices {
		if fieldIndex < 0 || fieldIndex >= len(player.Field) {
			continue
		}
		fc := player.Field[fieldIndex]
		if fc == nil || fc.Card.ID == "" {
			continue
		}
		fi := fieldIndex
		// 选项 ID 与提交语义一致：是"同系充能列表中的序号"（removeSelectedMagicBowCharge 按序号解析），
		// 场上位置单独放在 FieldIndex 里供前端定位盖牌。
		options = append(options, model.PromptOption{
			ID:         fmt.Sprintf("%d", seq),
			Label:      fc.Card.Name,
			FieldIndex: &fi,
		})
	}
	return &model.Prompt{
		Type:     model.PromptChooseCards,
		PlayerID: playerID,
		Message:  message,
		Options:  options,
		Min:      1,
		Max:      1,
		Presentation: &model.PromptPresentation{
			Kind:       model.PresentationCardPicker,
			Layout:     "inline",
			CardSource: "field",
		},
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
		Min:          1,
		Max:          1,
		Presentation: &model.PromptPresentation{Kind: model.PresentationBranchSelect, Layout: "overlay"},
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
	if len(remaining) == 0 {
		remaining = engineplayer.AllHandIndices(player)
	}
	options := make([]model.PromptOption, 0, len(remaining))
	for _, idx := range remaining {
		if idx < 0 || idx >= len(player.Hand) {
			continue
		}
		options = append(options, model.PromptOption{
			ID:     fmt.Sprintf("%d", idx),
			Label:  fmt.Sprintf("%d: %s", idx+1, promptfmt.FormatCardInfo(player.Hand[idx])),
			CardID: player.Hand[idx].ID,
		})
	}

	maxPlace := runtimeutil.ToIntContextValue(data["max_place"])
	if maxPlace < 0 {
		maxPlace = 0
	}
	if maxPlace > len(options) {
		maxPlace = len(options)
	}

	switch choiceType {
	case "mb_demon_eye_charge_card":
		flow, err := model.RequirePromptFlow(data, magicBowDemonEyeFlowID, magicBowSequentialFlowLabel)
		if err != nil {
			return nil
		}
		needCount := flow.Selection(magicBowChargeNeedStep).Count
		if needCount <= 0 {
			needCount = 1
		}
		selectedCount := len(flow.Selection(magicBowChargeCardsStep).OptionIndexes)
		remainingPick := needCount - selectedCount
		if remainingPick < 1 {
			remainingPick = 1
		}
		if remainingPick > len(options) {
			remainingPick = len(options)
		}
		if remainingPick < 1 {
			remainingPick = 1
		}
		return &model.Prompt{
			Type:     model.PromptChooseCards,
			PlayerID: playerID,
			Message:  "【魔眼】请选择1张手牌作为充能：",
			Options:  options,
			Min:      remainingPick,
			Max:      remainingPick,
			Presentation: &model.PromptPresentation{
				Kind:       model.PresentationCardPicker,
				CardSource: "hand",
			},
		}
	case "mb_charge_place_cards":
		if flow := model.PromptFlowFromContext(data); flow != nil && flow.FlowID == magicBowChargePlaceFlowID {
			needCount := flow.Selection(magicBowChargeNeedStep).Count
			selectedCount := len(flow.Selection(magicBowChargeCardsStep).OptionIndexes)
			if needCount > 0 {
				remainingPick := needCount - selectedCount
				if remainingPick < 1 {
					remainingPick = 1
				}
				if remainingPick > len(options) {
					remainingPick = len(options)
				}
				return &model.Prompt{
					Type:     model.PromptChooseCards,
					PlayerID: playerID,
					Message:  fmt.Sprintf("【充能】请选择%d张作为充能的手牌：", remainingPick),
					Options:  options,
					Min:      remainingPick,
					Max:      remainingPick,
					Presentation: &model.PromptPresentation{
						Kind:       model.PresentationCardPicker,
						CardSource: "hand",
					},
				}
			}
		}
		return &model.Prompt{
			Type:     model.PromptChooseCards,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【充能】请选择要放置为充能的手牌（最多%d张，可不选）：", maxPlace),
			Options:  options,
			Min:      0,
			Max:      maxPlace,
			Presentation: &model.PromptPresentation{
				Kind:       model.PresentationCardPicker,
				CardSource: "hand",
			},
		}
	default:
		return nil
	}
}

func initMagicBowChargePlaceFlow(needCount int) *model.PromptFlowState {
	flow := magicBowChargePlaceFlowRuntime.MustBeginAt(magicBowChargeCardsStep)
	flow.PutSelection(magicBowChargeNeedStep, model.PromptFlowSelection{Count: needCount})
	flow.PutSelection(magicBowChargeCardsStep, model.PromptFlowSelection{})
	return flow
}

func initMagicBowDemonEyeChargeFlow() *model.PromptFlowState {
	flow := magicBowDemonEyeFlowRuntime.MustBeginAt(magicBowChargeCardsStep)
	flow.PutSelection(magicBowChargeNeedStep, model.PromptFlowSelection{Count: 1})
	flow.PutSelection(magicBowChargeCardsStep, model.PromptFlowSelection{})
	return flow
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
	// 加伤记在本次攻击的规则修正上，由伤害计算时点消费；不能改实体攻击牌，
	// 否则这张牌进弃牌堆、重洗回牌库后会永久多 1 点伤害。
	rt.ApplyNextAttackDamageRule(user.ID, magicPierceDamageModifierID, "mb_magic_pierce", 1, model.RuleLifeUntilTurnEnd)
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

	// 摸牌后超出手牌上限：按通用爆牌规则先弃到上限（每弃 1 张士气 -1），弃完再选充能。
	// 摸牌本身跳过了自动爆牌检查（PreventOverflow），由这里接上，弃牌后经 after_discard 回到充能。
	if overflow := len(user.Hand) - rt.GetMaxHand(user); overflow > 0 {
		rt.PopInterrupt()
		rt.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: user.ID,
			Context: map[string]interface{}{
				"choice_type":                 "system_discard_cards",
				"discard_subflow":             true,
				"discard_count":               overflow,
				"discard_reason":              "hand_overflow",
				"victim_id":                   user.ID,
				"stay_in_turn":                true,
				"discard_forced":              true,
				"forced_reason":               "爆牌弃牌为强制步骤，不能取消",
				"prompt":                      fmt.Sprintf("【充能】手牌超出上限，请弃置%d张牌：", overflow),
				"flow_continuation_role_id":   "magic_bow",
				"flow_continuation_player_id": user.ID,
				"flow_continuation_skill_id":  "mb_charge_place",
				"charge_x":                    xValue,
			},
		})
		rt.Log(fmt.Sprintf("%s 的 [充能] 摸%d张后超出手牌上限%d张：先弃牌，再放置充能", user.Name, xValue, overflow))
		return nil
	}

	return enterChargePlaceCards(rt, user, xValue, ctxData)
}

// enterChargePlaceCards 进入"选最多 X 张手牌作为充能"；无可放置时直接结束充能。
// ctxData 为当前中断的上下文（爆牌弃完后由 after_discard 新开一个中断，传 nil）。
func enterChargePlaceCards(rt engineplayer.ChoiceRuntime, user *model.Player, xValue int, ctxData map[string]interface{}) error {
	room := ChargeCap - ChargeCount(user, "")
	maxPlace := xValue
	if maxPlace > len(user.Hand) {
		maxPlace = len(user.Hand)
	}
	if maxPlace > room {
		maxPlace = room
	}

	if maxPlace <= 0 {
		// 充能上限满了或其他原因导致不能盖牌
		if room <= 0 {
			rt.Log(fmt.Sprintf("%s 的 [充能] 生效：摸%d张，充能上限已满（%d/%d），不放置充能", user.Name, xValue, ChargeCount(user, ""), ChargeCap))
		} else {
			rt.Log(fmt.Sprintf("%s 的 [充能] 生效：摸%d张，不放置充能", user.Name, xValue))
		}
		if ctxData != nil {
			rt.PopInterrupt()
		}
		if rt.GetPendingInterrupt() == nil {
			rt.ApplyChoiceResumePoint(model.TurnStageActionStart)
		}
		return nil
	}

	// 直接进入盖牌选择（跳过数量选择步骤）
	// Min=0 允许不选，Max=maxPlace 最多可选maxPlace张
	if ctxData == nil {
		ctxData = map[string]interface{}{"user_id": user.ID}
	}
	ctxData["choice_type"] = "mb_charge_place_cards"
	ctxData["max_place"] = maxPlace
	model.SetPromptFlowContext(ctxData, initMagicBowChargePlaceFlow(0))
	ctxData["remaining_indices"] = engineplayer.AllHandIndices(user)
	if intr := rt.GetPendingInterrupt(); intr != nil && intr.PlayerID == user.ID && interruptChoiceType(intr) == "mb_charge_draw_x" {
		intr.Context = ctxData
		rt.NotifyInterruptPrompt()
	} else {
		rt.PushInterrupt(&model.Interrupt{Type: model.InterruptChoice, PlayerID: user.ID, Context: ctxData})
	}
	rt.Log(fmt.Sprintf("%s 的 [充能] 生效：摸%d张，请选择要放置为充能的手牌（最多%d张）", user.Name, xValue, maxPlace))
	return nil
}

func interruptChoiceType(intr *model.Interrupt) string {
	if intr == nil {
		return ""
	}
	data, _ := intr.Context.(map[string]interface{})
	choiceType, _ := data["choice_type"].(string)
	return choiceType
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
	model.SetPromptFlowContext(ctxData, initMagicBowChargePlaceFlow(needCount))
	ctxData["remaining_indices"] = engineplayer.AllHandIndices(user)
	if intr := rt.GetPendingInterrupt(); intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
	return nil
}

func handleChargePlaceCards(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	_, err := handleMagicBowChargeSelection(rt, ctxData, []int{selectionIndex}, true)
	return err
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

	actualExtra := 0
	for i := 0; i < extraX; i++ {
		if _, ok := RemoveChargeByElement(user, model.ElementThunder); !ok {
			break
		}
		actualExtra++
	}

	if actualExtra <= 0 {
		addThunderScatterDamages(rt, user.ID, targetIDs, "", 0)
		rt.Log(fmt.Sprintf("%s 的 [雷光散射] 生效：对所有对手各造成1点法术伤害", user.Name))
		finishThunderScatter(rt)
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
		addThunderScatterDamages(rt, user.ID, targetIDs, lockedTargetID, actualExtra)
		rt.Log(fmt.Sprintf("%s 的 [雷光散射] 生效：对 %s 造成%d点法术伤害，其余对手各1点", user.Name, target.Name, 1+actualExtra))
		finishThunderScatter(rt)
		return nil
	}

	// 尚未指定目标：先暂存所有对手列表，待选定后一次性按 1 / 1+X 入队。
	ctxData["choice_type"] = "mb_thunder_scatter_target"
	ctxData["extra_x"] = actualExtra
	if intr := rt.GetPendingInterrupt(); intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
	return nil
}

// addThunderScatterDamages 对所有对手各造成1点法术伤害；被指定的对手本次伤害额外+X，
// 合并为一次 1+X 点伤害（而不是再单独结算一次 X 点）。
func addThunderScatterDamages(rt engineplayer.ChoiceRuntime, userID string, targetIDs []string, boostedID string, extra int) {
	for _, targetID := range targetIDs {
		damage := 1
		if boostedID != "" && targetID == boostedID && extra > 0 {
			damage += extra
		}
		rt.AddPendingDamage(model.PendingDamage{
			SourceID:   userID,
			TargetID:   targetID,
			Damage:     damage,
			DamageType: model.MagicAttack,
		})
	}
}

func finishThunderScatter(rt engineplayer.ChoiceRuntime) {
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		rt.ApplyChoiceResumePoint(model.TurnStageExtraAction)
		if len(rt.GetPendingDamageQueue()) > 0 {
			rt.EnterDamageResolution(nil)
		}
	}
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
		// 与充能同口径：摸牌允许超过上限，摸完立刻结算爆牌（每弃 1 张士气 -1），再放置充能。
		rt.DrawCardsWithOptions(user.ID, 3, model.DrawOptions{PreventOverflow: true})
		rt.Log(fmt.Sprintf("%s 的 [魔眼] 分支②生效：摸3张牌", user.Name))
		if overflow := len(user.Hand) - rt.GetMaxHand(user); overflow > 0 {
			rt.PopInterrupt()
			rt.PushInterrupt(&model.Interrupt{
				Type:     model.InterruptChoice,
				PlayerID: user.ID,
				Context: map[string]interface{}{
					"choice_type":                 "system_discard_cards",
					"discard_subflow":             true,
					"discard_count":               overflow,
					"discard_reason":              "hand_overflow",
					"victim_id":                   user.ID,
					"stay_in_turn":                true,
					"discard_forced":              true,
					"forced_reason":               "爆牌弃牌为强制步骤，不能取消",
					"prompt":                      fmt.Sprintf("【魔眼】手牌超出上限，请弃置%d张牌：", overflow),
					"flow_continuation_role_id":   "magic_bow",
					"flow_continuation_player_id": user.ID,
					"flow_continuation_skill_id":  "mb_demon_eye_charge",
				},
			})
			rt.Log(fmt.Sprintf("%s 的 [魔眼] 摸3张后超出手牌上限%d张：先弃牌，再放置充能", user.Name, overflow))
			return nil
		}
		rt.PopInterrupt()
		demonEyeProceedToCharge(rt, user)
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
		if extraX < 0 {
			extraX = 0
		}
		addThunderScatterDamages(rt, user.ID, targetIDs, targetID, extraX)
		rt.Log(fmt.Sprintf("%s 的 [雷光散射] 生效：对 %s 造成%d点法术伤害，其余对手各1点", user.Name, target.Name, 1+extraX))
		finishThunderScatter(rt)
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
		// 目标没有手牌：弃牌这一步无事发生，没有替代效果；继续"将1张手牌作为充能，并获得1蓝水晶"。
		rt.Log(fmt.Sprintf("%s 的 [魔眼]：%s 没有手牌，无牌可弃", user.Name, target.Name))
		demonEyeProceedToCharge(rt, user)
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
	if cont.SkillID == "mb_charge_place" {
		user := rt.GetPlayers()[cont.PlayerID]
		if user == nil {
			return fmt.Errorf("充能后续执行者不存在: %s", cont.PlayerID)
		}
		return enterChargePlaceCards(rt, user, runtimeutil.ToIntContextValue(cont.Data["charge_x"]), nil)
	}
	if cont.SkillID == "mb_demon_eye_charge" {
		user := rt.GetPlayers()[cont.PlayerID]
		if user == nil {
			return fmt.Errorf("魔眼后续执行者不存在: %s", cont.PlayerID)
		}
		demonEyeProceedToCharge(rt, user)
		return nil
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

	rt.Log(fmt.Sprintf("%s 的 [魔眼]：%s 已弃置1张手牌", user.Name, discardPlayer.Name))
	demonEyeProceedToCharge(rt, user)
	return true
}

// demonEyeProceedToCharge 魔眼的后半段："然后将1张手牌作为充能，并获得1蓝水晶"。
//   - 魔弓无手牌 → 获得1点蓝水晶，恢复行动阶段
//   - 魔弓有手牌 → 推入 mb_demon_eye_charge_card 选择中断
func demonEyeProceedToCharge(rt engineplayer.ChoiceRuntime, user *model.Player) {
	if len(user.Hand) == 0 {
		gained := engineplayer.AddPlayerCrystalWithCap(rt, user, 1)
		rt.Log(fmt.Sprintf("%s 的 [魔眼] 生效：无手牌可充能，仅获得%d点蓝水晶", user.Name, gained))
		if rt.GetPendingInterrupt() == nil {
			rt.ApplyChoiceResumePoint(model.TurnStageActionStart)
		}
		return
	}

	// 有手牌：推入充能选择中断（需带上充能流程上下文，否则 BuildPrompt 返回 nil）
	chargeCtx := map[string]interface{}{
		"choice_type":       "mb_demon_eye_charge_card",
		"user_id":           user.ID,
		"remaining_indices": engineplayer.AllHandIndices(user),
	}
	model.SetPromptFlowContext(chargeCtx, initMagicBowDemonEyeChargeFlow())
	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: user.ID,
		Context:  chargeCtx,
	})
	rt.NotifyInterruptPrompt()
	rt.Log(fmt.Sprintf("%s 的 [魔眼]：请选择1张手牌作为充能", user.Name))
}

// handleChargePlaceCardsMultiSelect 批量多选盖牌（充能技能简化交互）。
// 玩家一次性选择要盖放的手牌（0~maxPlace张），无需先选择数量再逐张选择。
func handleChargePlaceCardsMultiSelect(rt engineplayer.ChoiceRuntime, playerID string, selections []int, ctxData map[string]interface{}) (bool, error) {
	return handleMagicBowChargeSelection(rt, ctxData, selections, false)
}

func handleMagicBowChargeSelection(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selections []int, allowPartial bool) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return false, fmt.Errorf("玩家不存在")
	}

	maxPlace := runtimeutil.ToIntContextValue(ctxData["max_place"])
	if maxPlace < 0 {
		maxPlace = 0
	}

	remaining := engineplayer.ParseIntSliceContextValue(ctxData["remaining_indices"])
	if len(remaining) == 0 {
		remaining = engineplayer.AllHandIndices(user)
	}

	flow := model.PromptFlowFromContext(ctxData)
	needCount := 0
	if flow != nil {
		switch flow.FlowID {
		case magicBowChargePlaceFlowID, magicBowDemonEyeFlowID:
			needCount = flow.Selection(magicBowChargeNeedStep).Count
		}
	}
	if choiceType == "mb_demon_eye_charge_card" && needCount <= 0 {
		needCount = 1
	}

	selected := make([]int, 0, len(selections))
	if flow != nil && needCount > 0 {
		selected = append(selected, flow.Selection(magicBowChargeCardsStep).OptionIndexes...)
	}

	resolveSelection := func(sel int) (int, error) {
		cardIdx, ok := runtimeutil.ResolveSelectionToCandidate(sel, remaining)
		if !ok || cardIdx < 0 || cardIdx >= len(user.Hand) {
			return 0, fmt.Errorf("无效的选项索引: %d", sel)
		}
		return cardIdx, nil
	}

	for _, sel := range selections {
		cardIdx, err := resolveSelection(sel)
		if err != nil {
			return false, err
		}
		for _, existing := range selected {
			if existing == cardIdx {
				return false, fmt.Errorf("不能重复选择同一张牌")
			}
		}
		selected = append(selected, cardIdx)
		nextRemaining := make([]int, 0, len(remaining))
		for _, idx := range remaining {
			if idx != cardIdx {
				nextRemaining = append(nextRemaining, idx)
			}
		}
		remaining = nextRemaining
	}

	if maxPlace > 0 && len(selected) > maxPlace {
		return false, fmt.Errorf("选择数量超过上限: 最多%d张", maxPlace)
	}

	if needCount > 0 {
		if len(selected) > needCount {
			return false, fmt.Errorf("选择数量超过上限: 需要%d张", needCount)
		}
		if len(selected) < needCount {
			if !allowPartial {
				return false, fmt.Errorf("需要选择%d张牌", needCount)
			}
			if flow == nil {
				return false, fmt.Errorf("缺少多步流状态")
			}
			flow.PutSelection(magicBowChargeCardsStep, model.PromptFlowSelection{
				OptionIndexes: append([]int{}, selected...),
				Count:         needCount,
			})
			ctxData["remaining_indices"] = remaining
			if intr := rt.GetPendingInterrupt(); intr != nil {
				intr.Context = ctxData
			}
			rt.NotifyInterruptPrompt()
			return true, nil
		}
	}

	if len(selected) == 0 {
		rt.Log(fmt.Sprintf("%s 选择不放置充能", user.Name))
		rt.PopInterrupt()
		if rt.GetPendingInterrupt() == nil {
			rt.ApplyChoiceResumePoint(model.TurnStageActionStart)
		}
		return true, nil
	}

	removed, err := engineplayer.RemoveCardsByIndicesFromHand(user, append([]int{}, selected...))
	if err != nil {
		return false, fmt.Errorf("移除手牌失败: %v", err)
	}

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
		gained := engineplayer.AddPlayerCrystalWithCap(rt, user, 1)
		rt.Log(fmt.Sprintf("%s 的 [魔眼] 生效：放置1张充能并获得%d点蓝水晶", user.Name, gained))
	} else {
		rt.Log(fmt.Sprintf("%s 的 [充能] 生效：放置%d张充能", user.Name, toPlace))
	}

	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		rt.ApplyChoiceResumePoint(model.TurnStageActionStart)
	}
	return true, nil
}
