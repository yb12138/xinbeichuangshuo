// gameflow: 圣弓手角色选择流。

package holy_bow

import (
	"fmt"
	"strconv"
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

// ---------------------------------------------------------------------------
// BuildPrompt
// ---------------------------------------------------------------------------

func (choiceHandler) BuildPrompt(rt engineplayer.ChoiceRuntime, choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "hb_holy_shard_combo":
		return buildHolyShardComboPrompt(playerID, player, data)
	case "hb_holy_shard_target":
		return buildHolyShardTargetPrompt(rt, playerID, data)
	case "hb_holy_shard_miss_confirm":
		return buildHolyShardMissConfirmPrompt(playerID)
	case "hb_holy_shard_miss_x":
		return buildHolyShardMissXPrompt(playerID, player, data)
	case "hb_holy_shard_miss_ally_target":
		return buildHolyShardMissAllyTargetPrompt(rt, playerID, data)
	case "hb_radiant_descent_cost":
		return buildRadiantDescentCostPrompt(playerID, data)
	case "hb_light_burst_mode":
		return buildLightBurstModePrompt(rt, playerID, player, data)
	case "hb_light_burst_mode_a_target", "hb_meteor_bullet_target":
		return buildAllyTargetPrompt(rt, playerID, data, choiceType)
	case "hb_light_burst_mode_b_x":
		return buildLightBurstModeBXPrompt(rt, playerID, player, data)
	case "hb_light_burst_mode_b_targets":
		return buildLightBurstModeBTargetsPrompt(rt, playerID, data)
	case "hb_light_burst_mode_b_discard":
		return buildLightBurstModeBDiscardPrompt(playerID, player, data)
	case "hb_meteor_bullet_cost":
		return buildMeteorBulletCostPrompt(playerID, data)
	case "hb_radiant_cannon_side":
		return buildRadiantCannonSidePrompt(playerID, data)
	case "hb_auto_fill_resource":
		return buildAutoFillResourcePrompt(playerID, data)
	case "hb_auto_fill_gain":
		return buildAutoFillGainPrompt(playerID, data)
	default:
		return nil
	}
}

// ---------------------------------------------------------------------------
// HandleChoice
// ---------------------------------------------------------------------------

func (choiceHandler) HandleChoice(rt engineplayer.ChoiceRuntime, _ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "hb_holy_shard_combo":
		return true, handleHolyShardCombo(rt, selectionIndex, ctxData)
	case "hb_holy_shard_target":
		return true, handleHolyShardTarget(rt, selectionIndex, ctxData)
	case "hb_holy_shard_miss_confirm":
		return true, handleHolyShardMissConfirm(rt, selectionIndex, ctxData)
	case "hb_holy_shard_miss_x":
		return true, handleHolyShardMissX(rt, selectionIndex, ctxData)
	case "hb_holy_shard_miss_ally_target":
		return true, handleHolyShardMissAllyTarget(rt, selectionIndex, ctxData)
	case "hb_radiant_descent_cost":
		return true, handleRadiantDescentCost(rt, selectionIndex, ctxData)
	case "hb_light_burst_mode":
		return true, handleLightBurstMode(rt, selectionIndex, ctxData)
	case "hb_light_burst_mode_a_target":
		return true, handleLightBurstModeATarget(rt, selectionIndex, ctxData)
	case "hb_light_burst_mode_b_x":
		return true, handleLightBurstModeBX(rt, selectionIndex, ctxData)
	case "hb_light_burst_mode_b_targets":
		return true, handleLightBurstModeBTargets(rt, selectionIndex, ctxData)
	case "hb_light_burst_mode_b_discard":
		return true, handleLightBurstModeBDiscard(rt, selectionIndex, ctxData)
	case "hb_meteor_bullet_cost":
		return true, handleMeteorBulletCost(rt, selectionIndex, ctxData)
	case "hb_meteor_bullet_target":
		return true, handleMeteorBulletTarget(rt, selectionIndex, ctxData)
	case "hb_radiant_cannon_side":
		return true, handleRadiantCannonSide(rt, selectionIndex, ctxData)
	case "hb_auto_fill_resource":
		return true, handleAutoFillResource(rt, selectionIndex, ctxData)
	case "hb_auto_fill_gain":
		return true, handleAutoFillGain(rt, selectionIndex, ctxData)
	default:
		return false, nil
	}
}

// ===========================================================================
// BuildPrompt helpers
// ===========================================================================

func buildHolyShardComboPrompt(playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	combos := runtimeutil.ParseStringSliceContextValue(data["combos"])
	options := make([]model.PromptOption, 0, len(combos))
	for _, combo := range combos {
		parts := strings.Split(combo, ":")
		if len(parts) != 2 {
			continue
		}
		idxParts := strings.Split(parts[1], ",")
		if len(idxParts) != 2 {
			continue
		}
		i, err1 := strconv.Atoi(strings.TrimSpace(idxParts[0]))
		j, err2 := strconv.Atoi(strings.TrimSpace(idxParts[1]))
		if err1 != nil || err2 != nil || player == nil || i < 0 || j < 0 || i >= len(player.Hand) || j >= len(player.Hand) || i == j {
			continue
		}
		options = append(options, model.PromptOption{
			ID:    combo,
			Label: fmt.Sprintf("%s系：%d:%s + %d:%s", promptfmt.ElementName(parts[0]), i+1, player.Hand[i].Name, j+1, player.Hand[j].Name),
		})
	}
	return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, ChoiceType: "hb_holy_shard_combo", Message: "【圣屑飓暴】请选择要弃置的2张同系攻击牌：", Options: options, Min: 1, Max: 1}
}

func buildHolyShardTargetPrompt(rt engineplayer.ChoiceRuntime, playerID string, data map[string]interface{}) *model.Prompt {
	targetIDs := runtimeutil.ParseStringSliceContextValue(data["target_ids"])
	return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, ChoiceType: "hb_holy_shard_target", Message: "【圣屑飓暴】请选择主动攻击目标：", Options: playerOptions(rt, targetIDs), Min: 1, Max: 1}
}

func buildHolyShardMissConfirmPrompt(playerID string) *model.Prompt {
	return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, ChoiceType: "hb_holy_shard_miss_confirm", Message: "【圣屑飓暴】未命中：是否移除治疗并令1名队友弃牌？", Options: []model.PromptOption{{ID: "0", Label: "是"}, {ID: "1", Label: "否"}}, Min: 1, Max: 1}
}

func buildHolyShardMissXPrompt(playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	validX := engineplayer.ParseIntSliceContextValue(data["valid_x"])
	if len(validX) == 0 {
		maxX := runtimeutil.ToIntContextValue(data["max_x"])
		if maxX <= 0 {
			maxX = 1
		}
		for x := 1; x <= maxX; x++ {
			validX = append(validX, x)
		}
	}
	options := make([]model.PromptOption, 0, len(validX))
	for _, x := range validX {
		options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", x), Label: fmt.Sprintf("移除%d点治疗，并令队友弃%d张牌", x, x)})
	}
	return &model.Prompt{
		Type:         model.PromptConfirm,
		PlayerID:     playerID,
		ChoiceType:   "hb_holy_shard_miss_x",
		Message:      "【圣屑飓暴】请选择移除治疗点数X：",
		Options:      options,
		Min:          1,
		Max:          1,
		Presentation: &model.PromptPresentation{Kind: model.PresentationNumeric, NumericBase: 0},
	}
}

func buildHolyShardMissAllyTargetPrompt(rt engineplayer.ChoiceRuntime, playerID string, data map[string]interface{}) *model.Prompt {
	allyIDs := runtimeutil.ParseStringSliceContextValue(data["ally_ids"])
	xValue := runtimeutil.ToIntContextValue(data["x_value"])
	return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, ChoiceType: "hb_holy_shard_miss_ally_target", Message: fmt.Sprintf("【圣屑飓暴】请选择1名队友弃置%d张手牌：", xValue), Options: playerOptions(rt, allyIDs), Min: 1, Max: 1}
}

func buildRadiantDescentCostPrompt(playerID string, data map[string]interface{}) *model.Prompt {
	modes := runtimeutil.ParseStringSliceContextValue(data["cost_modes"])
	options := make([]model.PromptOption, 0, len(modes))
	for _, mode := range modes {
		switch mode {
		case "heal":
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "移除2点治疗"})
		case "faith":
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "移除2点信仰"})
		}
	}
	return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, ChoiceType: "hb_radiant_descent_cost", Message: "【圣煌降临】请选择支付方式：", Options: options, Min: 1, Max: 1}
}

func buildLightBurstModePrompt(rt engineplayer.ChoiceRuntime, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	allyIDs := runtimeutil.ParseStringSliceContextValue(data["ally_ids"])
	enemyIDs := runtimeutil.ParseStringSliceContextValue(data["enemy_ids"])
	maxX := runtimeutil.ToIntContextValue(data["max_x"])
	canModeA := player != nil && player.Heal >= 1 && len(allyIDs) > 0
	canModeB := false
	if player != nil && maxX > 0 && len(enemyIDs) > 0 {
		handCount := len(player.Hand)
		for x := 1; x <= maxX; x++ {
			limit := handCount - x
			eligible := 0
			for _, enemyID := range enemyIDs {
				if enemy := rt.GetPlayers()[enemyID]; enemy != nil && len(enemy.Hand) <= limit {
					eligible++
				}
			}
			if eligible > 0 {
				canModeB = true
				break
			}
		}
	}
	options := make([]model.PromptOption, 0, 2)
	if canModeA {
		options = append(options, model.PromptOption{ID: "0", Label: "分支①：摸1、移除1治疗、+1信仰、我方1人+1治疗"})
	}
	if canModeB {
		options = append(options, model.PromptOption{ID: "1", Label: "分支②：移除X治疗并弃X牌，至多X名对手各受攻击伤害"})
	}
	return &model.Prompt{
		Type:         model.PromptConfirm,
		PlayerID:     playerID,
		ChoiceType:   "hb_light_burst_mode",
		Message:      "【圣光爆裂】请选择发动分支：",
		Options:      options,
		Min:          1,
		Max:          1,
		Presentation: &model.PromptPresentation{Kind: model.PresentationBranchSelect, Layout: "overlay"},
	}
}

func buildAllyTargetPrompt(rt engineplayer.ChoiceRuntime, playerID string, data map[string]interface{}, choiceType string) *model.Prompt {
	allyIDs := runtimeutil.ParseStringSliceContextValue(data["ally_ids"])
	msg := "请选择我方角色："
	if choiceType == "hb_light_burst_mode_a_target" {
		msg = "【圣光爆裂】分支①请选择获得治疗的我方角色："
	} else if choiceType == "hb_meteor_bullet_target" {
		msg = "【流星圣弹】请选择获得治疗的我方角色："
	}
	return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, ChoiceType: choiceType, Message: msg, Options: playerOptions(rt, allyIDs), Min: 1, Max: 1}
}

func buildLightBurstModeBXPrompt(rt engineplayer.ChoiceRuntime, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	enemyIDs := runtimeutil.ParseStringSliceContextValue(data["enemy_ids"])
	maxX := runtimeutil.ToIntContextValue(data["max_x"])
	options := make([]model.PromptOption, 0, maxX)
	if player != nil {
		for x := 1; x <= maxX; x++ {
			limit := len(player.Hand) - x
			eligible := 0
			for _, enemyID := range enemyIDs {
				if enemy := rt.GetPlayers()[enemyID]; enemy != nil && len(enemy.Hand) <= limit {
					eligible++
				}
			}
			if eligible <= 0 {
				continue
			}
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", x), Label: fmt.Sprintf("X=%d（移除%d治疗并弃%d张牌）", x, x, x)})
		}
	}
	return &model.Prompt{
		Type:         model.PromptConfirm,
		PlayerID:     playerID,
		ChoiceType:   "hb_light_burst_mode_b_x",
		Message:      "【圣光爆裂】分支②请选择X值：",
		Options:      options,
		Min:          1,
		Max:          1,
		Presentation: &model.PromptPresentation{Kind: model.PresentationNumeric, NumericBase: 0},
	}
}

func buildLightBurstModeBTargetsPrompt(rt engineplayer.ChoiceRuntime, playerID string, data map[string]interface{}) *model.Prompt {
	candidates := runtimeutil.ParseStringSliceContextValue(data["candidate_target_ids"])
	selectedSet := runtimeutil.IDsToSet(runtimeutil.ParseStringSliceContextValue(data["selected_target_ids"]))
	options := make([]model.PromptOption, 0, len(candidates)+1)
	for _, targetID := range candidates {
		if selectedSet[targetID] {
			continue
		}
		if target := rt.GetPlayers()[targetID]; target != nil {
			options = append(options, model.PromptOption{ID: targetID, Label: target.Name})
		}
	}
	xValue := runtimeutil.ToIntContextValue(data["x_value"])
	maxTargets := xValue
	if len(candidates) < maxTargets {
		maxTargets = len(candidates)
	}
	if len(selectedSet) > 0 {
		options = append(options, model.PromptOption{ID: "finish", Label: "完成目标选择", ButtonLabel: "完成"})
	}
	return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, ChoiceType: "hb_light_burst_mode_b_targets", Message: fmt.Sprintf("【圣光爆裂】分支②请点击角色立绘选择目标（已选%d/最多%d）：", len(selectedSet), maxTargets), Options: options, Min: 1, Max: 1}
}

func buildLightBurstModeBDiscardPrompt(playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	remaining := engineplayer.ParseIntSliceContextValue(data["remaining_indices"])
	selectedCount := len(engineplayer.ParseIntSliceContextValue(data["selected_indices"]))
	xValue := runtimeutil.ToIntContextValue(data["x_value"])
	options := make([]model.PromptOption, 0, len(remaining))
	for _, idx := range remaining {
		if player == nil || idx < 0 || idx >= len(player.Hand) {
			continue
		}
		options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", idx), Label: fmt.Sprintf("%d: %s", idx+1, promptfmt.FormatCardInfo(player.Hand[idx]))})
	}
	remainingPick := xValue - selectedCount
	if remainingPick < 1 {
		remainingPick = 1
	}
	if len(options) > 0 && remainingPick > len(options) {
		remainingPick = len(options)
	}
	return &model.Prompt{Type: model.PromptChooseCards, PlayerID: playerID, ChoiceType: "hb_light_burst_mode_b_discard", Message: fmt.Sprintf("【圣光爆裂】分支②请选择要弃置的%d张手牌：", remainingPick), Options: options, Min: remainingPick, Max: remainingPick}
}

func buildMeteorBulletCostPrompt(playerID string, data map[string]interface{}) *model.Prompt {
	modes := runtimeutil.ParseStringSliceContextValue(data["cost_modes"])
	options := make([]model.PromptOption, 0, len(modes))
	for _, mode := range modes {
		switch mode {
		case "heal":
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "移除1点治疗"})
		case "faith":
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "移除1点信仰"})
		}
	}
	return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, ChoiceType: "hb_meteor_bullet_cost", Message: "【流星圣弹】请选择要移除的资源：", Options: options, Min: 1, Max: 1}
}

func buildRadiantCannonSidePrompt(playerID string, data map[string]interface{}) *model.Prompt {
	requiredFaith := runtimeutil.ToIntContextValue(data["required_faith"])
	return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, ChoiceType: "hb_radiant_cannon_side", Message: fmt.Sprintf("【圣煌辉光炮】将消耗1辉光炮与%d点信仰。请选择士气对齐方向：", requiredFaith), Options: []model.PromptOption{{ID: "0", Label: "将红方士气调整为蓝方士气"}, {ID: "1", Label: "将蓝方士气调整为红方士气"}}, Min: 1, Max: 1}
}

func buildAutoFillResourcePrompt(playerID string, data map[string]interface{}) *model.Prompt {
	modes := runtimeutil.ParseStringSliceContextValue(data["resource_modes"])
	options := make([]model.PromptOption, 0, len(modes))
	for _, mode := range modes {
		switch mode {
		case "crystal":
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "分支①：消耗1蓝水晶（红宝石可替代）"})
		case "gem":
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", len(options)), Label: "分支②：消耗1红宝石并获得1蓝水晶"})
		}
	}
	return &model.Prompt{
		Type:         model.PromptConfirm,
		PlayerID:     playerID,
		ChoiceType:   "hb_auto_fill_resource",
		Message:      "【自动填充】请选择要发动的分支：",
		Options:      options,
		Min:          1,
		Max:          1,
		Presentation: &model.PromptPresentation{Kind: model.PresentationBranchSelect, Layout: "overlay"},
	}
}

func buildAutoFillGainPrompt(playerID string, data map[string]interface{}) *model.Prompt {
	branch, _ := data["branch"].(string)
	options := make([]model.PromptOption, 0, 2)
	if branch == "gem" {
		options = append(options, model.PromptOption{ID: "0", Label: "+2信仰"})
		options = append(options, model.PromptOption{ID: "1", Label: "+2治疗"})
	} else {
		options = append(options, model.PromptOption{ID: "0", Label: "+1信仰"})
		options = append(options, model.PromptOption{ID: "1", Label: "+1治疗"})
	}
	return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, ChoiceType: "hb_auto_fill_gain", Message: "【自动填充】请选择增益：", Options: options, Min: 1, Max: 1, Presentation: &model.PromptPresentation{Kind: model.PresentationBranchSelect, Layout: "overlay"}}
}

// ===========================================================================
// HandleChoice handlers
// ===========================================================================

func handleHolyShardCombo(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	combos := runtimeutil.ParseStringSliceContextValue(ctxData["combos"])
	if selectionIndex < 0 || selectionIndex >= len(combos) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	choice := combos[selectionIndex]
	parts := strings.Split(choice, ":")
	if len(parts) != 2 {
		return fmt.Errorf("同系组合格式错误")
	}
	element := strings.TrimSpace(parts[0])
	idxParts := strings.Split(parts[1], ",")
	if len(idxParts) != 2 {
		return fmt.Errorf("同系组合索引格式错误")
	}
	i, err1 := strconv.Atoi(strings.TrimSpace(idxParts[0]))
	j, err2 := strconv.Atoi(strings.TrimSpace(idxParts[1]))
	if err1 != nil || err2 != nil || i < 0 || j < 0 || i >= len(user.Hand) || j >= len(user.Hand) || i == j {
		return fmt.Errorf("无效的弃牌索引")
	}
	c1 := user.Hand[i]
	c2 := user.Hand[j]
	if c1.Type != model.CardTypeAttack || c2.Type != model.CardTypeAttack || c1.Element != c2.Element {
		return fmt.Errorf("圣屑飓暴需要弃置2张同系攻击牌")
	}
	removed, _ := engineplayer.RemoveCardsByIndicesFromHand(user, []int{i, j})
	rt.NotifyCardRevealed(user.ID, removed, "discard")
	rt.AppendToDiscard(removed)
	ctxData["selected_element"] = element
	ctxData["choice_type"] = "hb_holy_shard_target"
	if intr := rt.GetPendingInterrupt(); intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
	return nil
}

func handleHolyShardTarget(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
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
	if target.Camp == user.Camp {
		return fmt.Errorf("圣屑飓暴只能指定敌方目标")
	}
	eleStr, _ := ctxData["selected_element"].(string)
	ele := model.Element(eleStr)
	if ele == "" {
		return fmt.Errorf("圣屑飓暴攻击元素缺失")
	}
	virtualCard := model.Card{
		ID:          fmt.Sprintf("hb_holy_shard_%s_%d", user.ID, len(user.Hand)+1),
		Name:        "圣屑飓暴",
		Type:        model.CardTypeAttack,
		Element:     ele,
		Faction:     "圣",
		Damage:      2,
		Description: "由圣屑飓暴视为的圣命格主动攻击",
	}
	if user.Tokens == nil {
		user.Tokens = map[string]int{}
	}
	user.TurnState.SkillFlowState["hb_shard_miss_pending"] = 1
	rt.EnqueueVirtualAttack(user.ID, target.ID, virtualCard, "hb_holy_shard_storm")
	rt.Log(fmt.Sprintf("%s 发动 [圣屑飓暴]：对 %s 发起1次%s系圣命格主动攻击", user.Name, target.Name, ele))
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		rt.EnterActionExecutionStage()
	}
	return nil
}

func handleHolyShardMissConfirm(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	if selectionIndex == 1 {
		// Decline: skip the miss branch
		rt.PopInterrupt()
		if rt.GetPendingInterrupt() == nil {
			if len(rt.GetPendingDamageQueue()) > 0 {
				rt.EnterDamageResolution(nil)
			} else if len(rt.GetActionQueue()) > 0 {
				rt.EnterActionExecutionStage()
			} else {
				rt.EnterExtraActionStage()
			}
		}
		return nil
	}
	if selectionIndex != 0 {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
	if maxX <= 0 {
		rt.PopInterrupt()
		return nil
	}
	ctxData["choice_type"] = "hb_holy_shard_miss_x"
	if intr := rt.GetPendingInterrupt(); intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
	return nil
}

func handleHolyShardMissX(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	validX := engineplayer.ParseIntSliceContextValue(ctxData["valid_x"])
	if len(validX) == 0 {
		maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
		for x := 1; x <= maxX; x++ {
			validX = append(validX, x)
		}
	}
	if selectionIndex < 0 || selectionIndex >= len(validX) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	xValue := validX[selectionIndex]
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	allyIDs := holyBowShardMissEligibleAllies(rt, user, xValue)
	if len(allyIDs) == 0 {
		return fmt.Errorf("没有可弃置%d张牌的队友", xValue)
	}
	ctxData["x_value"] = xValue
	ctxData["ally_ids"] = allyIDs
	ctxData["choice_type"] = "hb_holy_shard_miss_ally_target"
	if intr := rt.GetPendingInterrupt(); intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
	return nil
}

func handleHolyShardMissAllyTarget(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	allyIDs := runtimeutil.ParseStringSliceContextValue(ctxData["ally_ids"])
	if selectionIndex < 0 || selectionIndex >= len(allyIDs) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	targetID := allyIDs[selectionIndex]
	target := rt.GetPlayers()[targetID]
	if target == nil {
		return fmt.Errorf("目标队友不存在")
	}
	xValue := runtimeutil.ToIntContextValue(ctxData["x_value"])
	if xValue <= 0 {
		return fmt.Errorf("无效的X值")
	}
	if user.Heal < xValue {
		return fmt.Errorf("治疗不足，无法移除%d点治疗", xValue)
	}
	if len(target.Hand) < xValue {
		return fmt.Errorf("目标队友手牌不足%d张，无法作为圣屑飓暴的弃牌目标", xValue)
	}
	user.Heal -= xValue
	discardNeed := xValue
	rt.Log(fmt.Sprintf("%s 的 [圣屑飓暴] 未命中分支生效：移除%d点治疗，指定 %s 弃置%d张手牌", user.Name, xValue, target.Name, discardNeed))
	rt.PopInterrupt()
	rt.PushInterrupt(newDiscardChoiceInterrupt(target.ID, map[string]interface{}{
		"discard_count":        discardNeed,
		"prompt":               fmt.Sprintf("【圣屑飓暴】请弃置%d张手牌：", discardNeed),
		"stay_in_turn":         true,
		"is_damage_resolution": true,
	}))
	return nil
}

func handleRadiantDescentCost(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	modes := runtimeutil.ParseStringSliceContextValue(ctxData["cost_modes"])
	if selectionIndex < 0 || selectionIndex >= len(modes) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	switch modes[selectionIndex] {
	case "heal":
		if user.Heal < 2 {
			return fmt.Errorf("治疗不足2点")
		}
		user.Heal -= 2
	case "faith":
		if Faith(user) < 2 {
			return fmt.Errorf("信仰不足2点")
		}
		AddFaith(user, -2)
	default:
		return fmt.Errorf("无效的支付方式")
	}
	if user.Tokens == nil {
		user.Tokens = map[string]int{}
	}
	EnterHolyGloryForm(user)
	model.AppendMagicAction(user, "圣煌降临")
	rt.Log(fmt.Sprintf("%s 发动 [圣煌降临]：进入圣煌形态并获得额外法术行动", user.Name))
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		resumePoint := engineplayer.MustChoiceResumePointFromMap(ctxData, "resume_phase")
		rt.ApplyChoiceResumePoint(resumePoint)
	}
	return nil
}

func handleLightBurstMode(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	allyIDs := runtimeutil.ParseStringSliceContextValue(ctxData["ally_ids"])
	enemyIDs := runtimeutil.ParseStringSliceContextValue(ctxData["enemy_ids"])
	maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
	modeOrder := make([]string, 0, 2)
	if user.Heal >= 1 && len(allyIDs) > 0 {
		modeOrder = append(modeOrder, "a")
	}
	if maxX > 0 && len(enemyIDs) > 0 {
		handCount := len(user.Hand)
		canModeB := false
		for x := 1; x <= maxX; x++ {
			limit := handCount - x
			for _, enemyID := range enemyIDs {
				if enemy := rt.GetPlayers()[enemyID]; enemy != nil && len(enemy.Hand) <= limit {
					canModeB = true
					break
				}
			}
			if canModeB {
				break
			}
		}
		if canModeB {
			modeOrder = append(modeOrder, "b")
		}
	}
	if selectionIndex < 0 || selectionIndex >= len(modeOrder) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	switch modeOrder[selectionIndex] {
	case "a":
		ctxData["choice_type"] = "hb_light_burst_mode_a_target"
	case "b":
		ctxData["choice_type"] = "hb_light_burst_mode_b_x"
	default:
		return fmt.Errorf("无效的分支")
	}
	if intr := rt.GetPendingInterrupt(); intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
	return nil
}

func handleLightBurstModeATarget(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	allyIDs := runtimeutil.ParseStringSliceContextValue(ctxData["ally_ids"])
	if selectionIndex < 0 || selectionIndex >= len(allyIDs) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	targetID := allyIDs[selectionIndex]
	if user.Heal < 1 {
		return fmt.Errorf("治疗不足，无法发动分支①")
	}
	rt.DrawCards(user.ID, 1)
	user.Heal--
	faith := AddFaith(user, 1)
	rt.Heal(targetID, 1)
	targetName := targetID
	if target := rt.GetPlayers()[targetID]; target != nil {
		targetName = target.Name
	}
	rt.Log(fmt.Sprintf("%s 的 [圣光爆裂] 分支①生效：摸1、移除1治疗、信仰+1（当前%d），%s +1治疗", user.Name, faith, targetName))
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		if !rt.RoutePendingDamageWithReturn(model.TurnStageExtraAction) {
			rt.EnterExtraActionStage()
		}
	}
	return nil
}

func handleLightBurstModeBX(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	enemyIDs := runtimeutil.ParseStringSliceContextValue(ctxData["enemy_ids"])
	maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
	validX := make([]int, 0, maxX)
	for x := 1; x <= maxX; x++ {
		limit := len(user.Hand) - x
		eligible := 0
		for _, enemyID := range enemyIDs {
			if enemy := rt.GetPlayers()[enemyID]; enemy != nil && len(enemy.Hand) <= limit {
				eligible++
			}
		}
		if eligible > 0 {
			validX = append(validX, x)
		}
	}
	if selectionIndex < 0 || selectionIndex >= len(validX) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	xValue := validX[selectionIndex]
	limit := len(user.Hand) - xValue
	candidateTargets := make([]string, 0, len(enemyIDs))
	for _, enemyID := range enemyIDs {
		if enemy := rt.GetPlayers()[enemyID]; enemy != nil && len(enemy.Hand) <= limit {
			candidateTargets = append(candidateTargets, enemyID)
		}
	}
	if len(candidateTargets) == 0 {
		return fmt.Errorf("没有满足手牌条件的目标")
	}
	ctxData["x_value"] = xValue
	ctxData["candidate_target_ids"] = candidateTargets
	ctxData["selected_target_ids"] = []string{}
	ctxData["choice_type"] = "hb_light_burst_mode_b_targets"
	if intr := rt.GetPendingInterrupt(); intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
	return nil
}

func handleLightBurstModeBTargets(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	candidates := runtimeutil.ParseStringSliceContextValue(ctxData["candidate_target_ids"])
	xValue := runtimeutil.ToIntContextValue(ctxData["x_value"])
	if xValue <= 0 {
		return fmt.Errorf("X值无效")
	}
	maxTargets := xValue
	if len(candidates) < maxTargets {
		maxTargets = len(candidates)
	}
	if maxTargets <= 0 {
		return fmt.Errorf("没有可选目标")
	}
	selected := runtimeutil.ParseStringSliceContextValue(ctxData["selected_target_ids"])
	selectedSet := runtimeutil.IDsToSet(selected)
	remaining := make([]string, 0, len(candidates))
	for _, targetID := range candidates {
		if !selectedSet[targetID] {
			remaining = append(remaining, targetID)
		}
	}
	allowFinish := len(selected) > 0
	finishIndex := -1
	if allowFinish {
		finishIndex = len(remaining)
	}
	proceedDiscard := false
	if allowFinish && selectionIndex == finishIndex {
		proceedDiscard = true
	} else {
		if selectionIndex < 0 || selectionIndex >= len(remaining) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		selected = append(selected, remaining[selectionIndex])
		ctxData["selected_target_ids"] = selected
		if len(selected) < maxTargets {
			if intr := rt.GetPendingInterrupt(); intr != nil {
				intr.Context = ctxData
			}
			rt.NotifyInterruptPrompt()
			return nil
		}
		proceedDiscard = true
	}
	if !proceedDiscard || len(selected) == 0 {
		return fmt.Errorf("至少需要选择1名目标")
	}
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	ctxData["selected_indices"] = []int{}
	ctxData["remaining_indices"] = engineplayer.AllHandIndices(user)
	ctxData["choice_type"] = "hb_light_burst_mode_b_discard"
	if intr := rt.GetPendingInterrupt(); intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
	return nil
}

func handleLightBurstModeBDiscard(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	xValue := runtimeutil.ToIntContextValue(ctxData["x_value"])
	if xValue <= 0 {
		return fmt.Errorf("X值无效")
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
	if len(selected) < xValue {
		ctxData["selected_indices"] = selected
		ctxData["remaining_indices"] = nextRemaining
		if intr := rt.GetPendingInterrupt(); intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return nil
	}
	if user.Heal < xValue {
		return fmt.Errorf("治疗不足，无法移除%d点治疗", xValue)
	}
	removed, _ := engineplayer.RemoveCardsByIndicesFromHand(user, append([]int{}, selected...))
	rt.NotifyCardRevealed(user.ID, removed, "discard")
	rt.AppendToDiscard(removed)
	user.Heal -= xValue
	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["selected_target_ids"])
	yValue := 0
	for _, tid := range targetIDs {
		if t := rt.GetPlayers()[tid]; t != nil && t.Heal > 0 {
			yValue++
		}
	}
	damage := yValue + 2
	for _, tid := range targetIDs {
		if rt.GetPlayers()[tid] == nil {
			continue
		}
		rt.AddPendingDamage(model.PendingDamage{SourceID: user.ID, TargetID: tid, Damage: damage, DamageType: model.AttackDamage})
	}
	rt.Log(fmt.Sprintf("%s 的 [圣光爆裂] 分支②生效：移除%d治疗并弃%d张牌，对%d名目标各造成%d点攻击伤害（Y=%d）", user.Name, xValue, xValue, len(targetIDs), damage, yValue))
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		if !rt.RoutePendingDamageWithReturn(model.TurnStageExtraAction) {
			rt.EnterExtraActionStage()
		}
	}
	return nil
}

func handleMeteorBulletCost(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	modes := runtimeutil.ParseStringSliceContextValue(ctxData["cost_modes"])
	if selectionIndex < 0 || selectionIndex >= len(modes) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	ctxData["chosen_cost_mode"] = modes[selectionIndex]
	ctxData["choice_type"] = "hb_meteor_bullet_target"
	if intr := rt.GetPendingInterrupt(); intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
	return nil
}

func handleMeteorBulletTarget(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	allyIDs := runtimeutil.ParseStringSliceContextValue(ctxData["ally_ids"])
	if selectionIndex < 0 || selectionIndex >= len(allyIDs) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	targetID := allyIDs[selectionIndex]
	mode, _ := ctxData["chosen_cost_mode"].(string)
	switch mode {
	case "heal":
		if user.Heal <= 0 {
			return fmt.Errorf("治疗不足，无法发动流星圣弹")
		}
		user.Heal--
	case "faith":
		if Faith(user) <= 0 {
			return fmt.Errorf("信仰不足，无法发动流星圣弹")
		}
		AddFaith(user, -1)
	default:
		return fmt.Errorf("流星圣弹资源选择无效")
	}
	rt.Heal(targetID, 1)
	targetName := targetID
	if target := rt.GetPlayers()[targetID]; target != nil {
		targetName = target.Name
	}
	rt.Log(fmt.Sprintf("%s 发动 [流星圣弹]：移除1点%s，令 %s +1治疗", user.Name, map[string]string{"heal": "治疗", "faith": "信仰"}[mode], targetName))
	rawCtx, _ := ctxData["user_ctx"].(*model.Context)
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		if rawCtx != nil && rawCtx.AttackDeclaredPhase() {
			if len(rt.GetActionQueue()) > 0 {
				rt.EnterActionExecutionStage()
			} else {
				rt.EnterTurnEndStage()
			}
		} else {
			rt.EnterExtraActionStage()
		}
	}
	return nil
}

func handleRadiantCannonSide(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	if !engineplayer.IsCharacter(user, "holy_bow") {
		return fmt.Errorf("仅圣弓可发动圣煌辉光炮")
	}
	if user.Tokens == nil {
		user.Tokens = map[string]int{}
	}
	requiredFaith := runtimeutil.ToIntContextValue(ctxData["required_faith"])
	if requiredFaith <= 0 {
		requiredFaith = 4
	}
	if Cannon(user) <= 0 {
		return fmt.Errorf("圣煌辉光炮指示物不足")
	}
	if Faith(user) < requiredFaith {
		return fmt.Errorf("信仰不足，需要%d点", requiredFaith)
	}
	if selectionIndex != 0 && selectionIndex != 1 {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	user.Tokens["hb_cannon"] = Cannon(user) - 1
	AddFaith(user, -requiredFaith)

	// Adjust all players' hands to 4 cards
	for _, pid := range rt.GetPlayerOrder() {
		p := rt.GetPlayers()[pid]
		if p == nil {
			continue
		}
		if len(p.Hand) > 4 {
			discarded := append([]model.Card{}, p.Hand[4:]...)
			p.Hand = append([]model.Card{}, p.Hand[:4]...)
			rt.NotifyCardRevealed(p.ID, discarded, "discard")
			rt.AppendToDiscard(discarded)
		} else if len(p.Hand) < 4 {
			drawN := 4 - len(p.Hand)
			rt.DrawCardsDirect(p.ID, drawN, "hb_radiant_cannon_adjust")
		}
	}

	// Add camp cup (+1)
	rt.AddCampCup(user.Camp)

	// Align morale
	if selectionIndex == 0 {
		rt.SetCampMorale(string(model.RedCamp), rt.GetCampMorale(string(model.BlueCamp)))
	} else {
		rt.SetCampMorale(string(model.BlueCamp), rt.GetCampMorale(string(model.RedCamp)))
	}
	rt.Log(fmt.Sprintf("%s 发动 [圣煌辉光炮]：全员手牌调整至4，我方星杯+1，并完成士气对齐", user.Name))
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		if !rt.RoutePendingDamageWithReturn(model.TurnStageExtraAction) {
			rt.EnterExtraActionStage()
		}
	}
	return nil
}

func handleAutoFillResource(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	modes := runtimeutil.ParseStringSliceContextValue(ctxData["resource_modes"])
	if selectionIndex < 0 || selectionIndex >= len(modes) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	branch := modes[selectionIndex]
	switch branch {
	case "crystal":
		if !rt.ConsumeCrystalCost(user.ID, 1) {
			return fmt.Errorf("自动填充分支①需要1点蓝水晶（红宝石可替代）")
		}
	case "gem":
		if user.Gem <= 0 {
			return fmt.Errorf("自动填充分支②需要1点红宝石")
		}
		user.Gem--
		maxEnergy := rt.GetMaxHand(user) // Use GetMaxHand as proxy for energy cap
		if maxEnergy <= 0 {
			maxEnergy = 3
		}
		// Energy cap is typically 3 (gem + crystal total)
		// Replaced: original uses e.getPlayerEnergyCap(user) which returns 3
		energyCap := 3
		if user.Gem+user.Crystal < energyCap {
			user.Crystal++
			if user.Gem+user.Crystal > energyCap {
				user.Crystal -= user.Gem + user.Crystal - energyCap
			}
		}
	default:
		return fmt.Errorf("无效分支")
	}
	ctxData["branch"] = branch
	ctxData["choice_type"] = "hb_auto_fill_gain"
	if intr := rt.GetPendingInterrupt(); intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
	return nil
}

func handleAutoFillGain(rt engineplayer.ChoiceRuntime, selectionIndex int, ctxData map[string]interface{}) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	branch, _ := ctxData["branch"].(string)
	if selectionIndex != 0 && selectionIndex != 1 {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	if branch == "gem" {
		if selectionIndex == 0 {
			now := AddFaith(user, 2)
			rt.Log(fmt.Sprintf("%s 的 [自动填充] 分支②生效：+2信仰（当前%d）", user.Name, now))
		} else {
			rt.Heal(user.ID, 2)
			rt.Log(fmt.Sprintf("%s 的 [自动填充] 分支②生效：+2治疗", user.Name))
		}
	} else {
		if selectionIndex == 0 {
			now := AddFaith(user, 1)
			rt.Log(fmt.Sprintf("%s 的 [自动填充] 分支①生效：+1信仰（当前%d）", user.Name, now))
		} else {
			rt.Heal(user.ID, 1)
			rt.Log(fmt.Sprintf("%s 的 [自动填充] 分支①生效：+1治疗", user.Name))
		}
	}
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		rt.EnterTurnEndStage()
	}
	return nil
}

// ===========================================================================
// Local helpers
// ===========================================================================

// playerOptions builds PromptOption list from player IDs.
func playerOptions(rt engineplayer.ChoiceRuntime, playerIDs []string) []model.PromptOption {
	options := make([]model.PromptOption, 0, len(playerIDs))
	for _, pid := range playerIDs {
		if p := rt.GetPlayers()[pid]; p != nil {
			options = append(options, model.PromptOption{ID: pid, Label: p.Name})
		}
	}
	return options
}


// newDiscardChoiceInterrupt creates a normalized discard-choice interrupt.
func newDiscardChoiceInterrupt(playerID string, data map[string]interface{}) *model.Interrupt {
	if data == nil {
		data = map[string]interface{}{}
	}
	data["discard_subflow"] = true
	if _, ok := data["choice_type"].(string); !ok || data["choice_type"] == "" {
		data["choice_type"] = "system_discard_cards"
	}
	return &model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: playerID,
		Context:  data,
	}
}

// holyBowShardMissEligibleAllies returns ally IDs that have at least x cards in hand.
func holyBowShardMissEligibleAllies(rt engineplayer.ChoiceRuntime, user *model.Player, x int) []string {
	if user == nil || x <= 0 {
		return nil
	}
	allyIDs := make([]string, 0)
	for _, pid := range rt.GetPlayerOrder() {
		p := rt.GetPlayers()[pid]
		if p == nil || p.Camp != user.Camp || p.ID == user.ID {
			continue
		}
		if len(p.Hand) < x {
			continue
		}
		allyIDs = append(allyIDs, p.ID)
	}
	return allyIDs
}

