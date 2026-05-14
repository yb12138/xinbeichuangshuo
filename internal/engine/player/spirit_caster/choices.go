// gameflow: 灵能师角色选择流。

package spirit_caster

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
	case "sc_incant_confirm":
		return buildIncantConfirmPrompt(playerID)
	case "sc_incant_confirm_no_hand":
		return buildIncantConfirmNoHandPrompt(playerID)
	case "sc_incant_card":
		return buildIncantCardPrompt(playerID, player)
	case "sc_hundred_night_power":
		return buildHundredNightPowerPrompt(playerID, player)
	case "sc_hundred_night_fire_reveal":
		return buildHundredNightFireRevealPrompt(playerID)
	case "sc_hundred_night_exclude_pick":
		return buildHundredNightExcludePickPrompt(rt, playerID, data)
	case "sc_hundred_night_target":
		p := engineplayer.BuildTargetChoicePrompt(rt, playerID, "【百鬼夜行】请选择1点法术伤害目标：", data, false)
		if p != nil {
			p.ChoiceType = "sc_hundred_night_target"
		}
		return p
	case "sc_spiritual_collapse_confirm":
		return buildSpiritualCollapseConfirmPrompt(playerID)
	case "sc_talisman_wind_discard":
		return buildTalismanWindDiscardPrompt(rt, playerID, data)
	case "sc_talisman_pick":
		return buildTalismanPickPrompt(playerID, player)
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
	case "sc_incant_confirm":
		return true, handleIncantConfirm(rt, ctxData, selectionIndex)
	case "sc_incant_confirm_no_hand":
		return true, handleIncantConfirmNoHand(rt, ctxData, selectionIndex)
	case "sc_incant_card":
		return true, handleIncantCard(rt, ctxData, selectionIndex)
	case "sc_hundred_night_power":
		return true, handleHundredNightPower(rt, ctxData, selectionIndex)
	case "sc_hundred_night_fire_reveal":
		return true, handleHundredNightFireReveal(rt, ctxData, selectionIndex)
	case "sc_hundred_night_exclude_pick":
		return true, handleHundredNightExcludePick(rt, ctxData, selectionIndex)
	case "sc_hundred_night_target":
		return true, handleHundredNightTarget(rt, ctxData, selectionIndex)
	case "sc_spiritual_collapse_confirm":
		return true, handleSpiritualCollapseConfirm(rt, ctxData, selectionIndex)
	case "sc_talisman_wind_discard":
		return true, handleTalismanWindDiscard(rt, ctxData, selectionIndex)
	case "sc_talisman_pick":
		return true, handleTalismanPick(rt, ctxData, selectionIndex)
	default:
		return false, nil
	}
}

// ===========================================================================
// BuildPrompt helpers
// ===========================================================================

func buildIncantConfirmPrompt(playerID string) *model.Prompt {
	return &model.Prompt{
		Type:       model.PromptConfirm,
		PlayerID:   playerID,
		Message:    "【念咒】是否将1张手牌面朝下放置为妖力？",
		ChoiceType: "sc_incant_confirm",
		Options:    []model.PromptOption{{ID: "0", Label: "是"}, {ID: "1", Label: "否"}},
		Min:        1,
		Max:        1,
	}
}

func buildIncantConfirmNoHandPrompt(playerID string) *model.Prompt {
	return &model.Prompt{
		Type:       model.PromptConfirm,
		PlayerID:   playerID,
		Message:    "【念咒】无手牌可放置为妖力，是否跳过？",
		ChoiceType: "sc_incant_confirm_no_hand",
		Options:    []model.PromptOption{{ID: "0", Label: "跳过念咒"}, {ID: "1", Label: "取消"}},
		Min:        1,
		Max:        1,
	}
}

func buildIncantCardPrompt(playerID string, player *model.Player) *model.Prompt {
	options := make([]model.PromptOption, 0, len(player.Hand))
	for idx, c := range player.Hand {
		options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", idx), Label: fmt.Sprintf("%d: %s", idx+1, promptfmt.FormatCardInfo(c))})
	}
	return &model.Prompt{
		Type:       model.PromptConfirm,
		PlayerID:   playerID,
		Message:    "【念咒】请选择要作为妖力盖放的手牌：",
		ChoiceType: "sc_incant_card",
		Options:    options,
		Min:        1,
		Max:        1,
	}
}

func buildHundredNightPowerPrompt(playerID string, player *model.Player) *model.Prompt {
	powers := PowerCovers(player)
	options := make([]model.PromptOption, 0, len(powers))
	for i, fc := range powers {
		if fc == nil {
			continue
		}
		eleZh := promptfmt.ElementName(string(fc.Card.Element))
		if eleZh == "" {
			eleZh = string(fc.Card.Element)
		}
		options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", i), Label: fmt.Sprintf("%s（%s系）", fc.Card.Name, eleZh)})
	}
	return &model.Prompt{
		Type:       model.PromptConfirm,
		PlayerID:   playerID,
		ChoiceType: "sc_hundred_night_power",
		Message:  "【百鬼夜行】请选择要移除的1个妖力：",
		Options:  options,
		Min:      1,
		Max:      1,
	}
}

func buildHundredNightFireRevealPrompt(playerID string) *model.Prompt {
	return &model.Prompt{
		Type:       model.PromptConfirm,
		PlayerID:   playerID,
		Message:    "【百鬼夜行】移除的是火系妖力，是否展示并改为范围伤害？",
		ChoiceType: "sc_hundred_night_fire_reveal",
		Options:    []model.PromptOption{{ID: "0", Label: "展示并改为范围伤害"}, {ID: "1", Label: "不展示，改为单体伤害"}},
		Min:        1,
		Max:        1,
	}
}

func buildHundredNightExcludePickPrompt(rt engineplayer.ChoiceRuntime, playerID string, data map[string]interface{}) *model.Prompt {
	targetIDs := runtimeutil.ParseStringSliceContextValue(data["target_ids"])
	selectedSet := runtimeutil.IDsToSet(runtimeutil.ParseStringSliceContextValue(data["selected_exclude_ids"]))
	options := make([]model.PromptOption, 0, len(targetIDs))
	for _, targetID := range targetIDs {
		if selectedSet[targetID] {
			continue
		}
		if target := rt.GetPlayers()[targetID]; target != nil {
			options = append(options, model.PromptOption{ID: targetID, Label: target.Name})
		}
	}
	return &model.Prompt{
		Type:       model.PromptConfirm,
		PlayerID:   playerID,
		Message:    fmt.Sprintf("【百鬼夜行】请选择第 %d/2 名排除目标：", len(selectedSet)+1),
		ChoiceType: "sc_hundred_night_exclude_pick",
		Options:  options,
		Min:      1,
		Max:      1,
	}
}

func buildSpiritualCollapseConfirmPrompt(playerID string) *model.Prompt {
	return &model.Prompt{
		Type:       model.PromptConfirm,
		PlayerID:   playerID,
		Message:    "【灵力崩解】是否消耗1点水晶（红宝石可替代），使本次每段伤害额外+1？",
		ChoiceType: "sc_spiritual_collapse_confirm",
		Options:    []model.PromptOption{{ID: "0", Label: "是"}, {ID: "1", Label: "否"}},
		Min:        1,
		Max:        1,
	}
}

func buildTalismanWindDiscardPrompt(rt engineplayer.ChoiceRuntime, playerID string, data map[string]interface{}) *model.Prompt {
	currentTargetID, _ := data["current_target_id"].(string)
	target := rt.GetPlayers()[currentTargetID]
	if target == nil {
		return nil
	}
	options := make([]model.PromptOption, 0, len(target.Hand))
	for idx, c := range target.Hand {
		options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", idx), Label: fmt.Sprintf("%d: %s", idx+1, promptfmt.FormatCardInfo(c))})
	}
	return &model.Prompt{
		Type:       model.PromptConfirm,
		PlayerID:   playerID,
		Message:    fmt.Sprintf("【灵符-风行】请 %s 选择1张手牌弃置：", target.Name),
		ChoiceType: "sc_talisman_wind_discard",
		Options:  options,
		Min:      1,
		Max:      1,
	}
}

func buildTalismanPickPrompt(playerID string, player *model.Player) *model.Prompt {
	options := make([]model.PromptOption, 0, 2)
	options = append(options, model.PromptOption{ID: "0", Label: "灵符-雷鸣"})
	options = append(options, model.PromptOption{ID: "1", Label: "灵符-风行"})
	return &model.Prompt{
		Type:       model.PromptConfirm,
		PlayerID:   playerID,
		Message:    "【灵符】请选择要发动的灵符类型：",
		ChoiceType: "sc_talisman_pick",
		Options:    options,
		Min:        1,
		Max:        1,
	}
}

// ===========================================================================
// HandleChoice handlers
// ===========================================================================

func handleIncantConfirm(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	skillID, _ := ctxData["skill_id"].(string)
	targetIDs := runtimeutil.DedupeIDs(runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"]))
	switch selectionIndex {
	case 0:
		if len(user.Hand) == 0 {
			rt.Log(fmt.Sprintf("%s 的 [念咒] 未触发：无手牌可放置为妖力", user.Name))
			rt.PopInterrupt()
			return continueSpiritCasterResolution(rt, user, skillID, targetIDs)
		}
		ctxData["choice_type"] = "sc_incant_card"
		if intr := rt.GetPendingInterrupt(); intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return nil
	case 1:
		rt.PopInterrupt()
		return continueSpiritCasterResolution(rt, user, skillID, targetIDs)
	default:
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
}

func handleIncantConfirmNoHand(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	skillID, _ := ctxData["skill_id"].(string)
	targetIDs := runtimeutil.DedupeIDs(runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"]))
	switch selectionIndex {
	case 0:
		// Skip incant, go directly to the talisman flow
		rt.PopInterrupt()
		return continueSpiritCasterResolution(rt, user, skillID, targetIDs)
	case 1:
		// Cancel
		rt.PopInterrupt()
		return nil
	default:
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
}

func handleIncantCard(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	skillID, _ := ctxData["skill_id"].(string)
	targetIDs := runtimeutil.DedupeIDs(runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"]))
	if selectionIndex < 0 || selectionIndex >= len(user.Hand) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	card := user.Hand[selectionIndex]
	user.Hand = append(user.Hand[:selectionIndex], user.Hand[selectionIndex+1:]...)
	AddPowerCard(user, card)
	rt.Log(fmt.Sprintf("%s 发动 [念咒]：将1张手牌盖放为妖力（当前妖力%d）", user.Name, PowerCount(user, "")))
	rt.PopInterrupt()
	return continueSpiritCasterResolution(rt, user, skillID, targetIDs)
}

func handleHundredNightPower(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	powers := PowerCovers(user)
	if len(powers) == 0 {
		return fmt.Errorf("没有可移除的妖力")
	}
	powerIdx, ok := runtimeutil.ResolveSelectionToCandidate(selectionIndex, func() []int {
		idxs := make([]int, 0, len(powers))
		for i := range powers {
			idxs = append(idxs, i)
		}
		return idxs
	}())
	if !ok || powerIdx < 0 || powerIdx >= len(powers) || powers[powerIdx] == nil {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	selectedPower := powers[powerIdx]
	card := selectedPower.Card
	user.RemoveFieldCard(selectedPower)
	SyncPowerToken(user)
	rt.AppendToDiscard([]model.Card{card})
	rt.Log(fmt.Sprintf("%s 发动 [百鬼夜行]：移除1个妖力", user.Name))

	ctxData["removed_card"] = card
	if card.Element == model.ElementFire {
		ctxData["removed_element"] = string(card.Element)
		ctxData["removed_name"] = card.Name
		ctxData["choice_type"] = "sc_hundred_night_fire_reveal"
		ctxData["target_ids"] = append([]string{}, rt.GetPlayerOrder()...)
		if intr := rt.GetPendingInterrupt(); intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return nil
	}
	ctxData["choice_type"] = "sc_hundred_night_target"
	ctxData["target_ids"] = append([]string{}, rt.GetPlayerOrder()...)
	if intr := rt.GetPendingInterrupt(); intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
	return nil
}

func handleHundredNightFireReveal(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	switch selectionIndex {
	case 0:
		userID, _ := ctxData["user_id"].(string)
		user := rt.GetPlayers()[userID]
		if user != nil {
			if removedCard, ok := ctxData["removed_card"].(model.Card); ok {
				rt.NotifyCardRevealed(user.ID, []model.Card{removedCard}, "discard")
			}
			rt.Log(fmt.Sprintf("%s 展示了火系妖力，触发 [百鬼夜行] 范围分支", user.Name))
		}
		ctxData["choice_type"] = "sc_hundred_night_exclude_pick"
		ctxData["target_ids"] = append([]string{}, rt.GetPlayerOrder()...)
		ctxData["selected_exclude_ids"] = []string{}
		if intr := rt.GetPendingInterrupt(); intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return nil
	case 1:
		ctxData["choice_type"] = "sc_hundred_night_target"
		ctxData["target_ids"] = append([]string{}, rt.GetPlayerOrder()...)
		if intr := rt.GetPendingInterrupt(); intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return nil
	default:
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
}

func handleHundredNightExcludePick(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	allTargetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
	if len(allTargetIDs) < 2 {
		return fmt.Errorf("可选目标不足2名")
	}
	selected := runtimeutil.DedupeIDs(runtimeutil.ParseStringSliceContextValue(ctxData["selected_exclude_ids"]))
	selectedSet := runtimeutil.IDsToSet(selected)
	remaining := make([]string, 0, len(allTargetIDs))
	for _, targetID := range allTargetIDs {
		if !selectedSet[targetID] {
			remaining = append(remaining, targetID)
		}
	}
	if selectionIndex < 0 || selectionIndex >= len(remaining) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	selected = append(selected, remaining[selectionIndex])
	if len(selected) < 2 {
		ctxData["selected_exclude_ids"] = selected
		if intr := rt.GetPendingInterrupt(); intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return nil
	}
	if rt.CanPayCrystalCost(user.ID, 1) {
		ctxData["choice_type"] = "sc_spiritual_collapse_confirm"
		ctxData["mode"] = "sc_hundred_night_fire_aoe"
		ctxData["exclude_ids"] = selected
		if intr := rt.GetPendingInterrupt(); intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return nil
	}
	resolveHundredNightFireAOE(rt, user, selected, 0)
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		if len(rt.GetPendingDamageQueue()) > 0 {
			rt.EnterDamageResolution(nil)
		}
	}
	return nil
}

func handleHundredNightTarget(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
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
	if rt.CanPayCrystalCost(user.ID, 1) {
		ctxData["choice_type"] = "sc_spiritual_collapse_confirm"
		ctxData["mode"] = "sc_hundred_night_single"
		ctxData["target_id"] = targetID
		if intr := rt.GetPendingInterrupt(); intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return nil
	}
	resolveHundredNightSingle(rt, user, targetID, 0)
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		if len(rt.GetPendingDamageQueue()) > 0 {
			rt.EnterDamageResolution(nil)
		}
	}
	return nil
}

func handleSpiritualCollapseConfirm(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	bonus := 0
	if selectionIndex == 0 {
		if !rt.ConsumeCrystalCost(user.ID, 1) {
			return fmt.Errorf("灵力崩解需要1点水晶（红宝石可替代）")
		}
		bonus = 1
		rt.Log(fmt.Sprintf("%s 发动 [灵力崩解]：本次每段伤害额外+1", user.Name))
	} else if selectionIndex != 1 {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	mode, _ := ctxData["mode"].(string)
	switch mode {
	case "sc_talisman_thunder":
		targetIDs := runtimeutil.DedupeIDs(runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"]))
		rt.PopInterrupt()
		resolveThunderDamage(rt, user, targetIDs, bonus)
	case "sc_hundred_night_single":
		targetID, _ := ctxData["target_id"].(string)
		if targetID == "" {
			return fmt.Errorf("百鬼夜行目标缺失")
		}
		resolveHundredNightSingle(rt, user, targetID, bonus)
		rt.PopInterrupt()
	case "sc_hundred_night_fire_aoe":
		excludeIDs := runtimeutil.DedupeIDs(runtimeutil.ParseStringSliceContextValue(ctxData["exclude_ids"]))
		if len(excludeIDs) != 2 {
			return fmt.Errorf("百鬼夜行火系分支需要2名排除目标")
		}
		resolveHundredNightFireAOE(rt, user, excludeIDs, bonus)
		rt.PopInterrupt()
	default:
		return fmt.Errorf("灵力崩解上下文无效")
	}
	if rt.GetPendingInterrupt() == nil && len(rt.GetPendingDamageQueue()) > 0 {
		rt.EnterDamageResolution(nil)
	}
	return nil
}

func handleTalismanWindDiscard(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("灵符师不存在")
	}
	ordered := runtimeutil.ParseStringSliceContextValue(ctxData["ordered_target_ids"])
	if len(ordered) == 0 {
		return fmt.Errorf("灵符-风行上下文无效")
	}
	cursor := runtimeutil.ToIntContextValue(ctxData["cursor"])
	if cursor < 0 || cursor >= len(ordered) {
		return fmt.Errorf("灵符-风行游标无效")
	}
	currentTargetID, _ := ctxData["current_target_id"].(string)
	if currentTargetID == "" {
		currentTargetID = ordered[cursor]
	}
	target := rt.GetPlayers()[currentTargetID]
	if target == nil {
		return fmt.Errorf("弃牌目标不存在")
	}
	if len(target.Hand) == 0 {
		rt.Log(fmt.Sprintf("%s 的 [灵符-风行]：%s 已无手牌，跳过", user.Name, target.Name))
	} else {
		candidates := engineplayer.AllHandIndices(target)
		cardIdx, ok := runtimeutil.ResolveSelectionToCandidate(selectionIndex, candidates)
		if !ok || cardIdx < 0 || cardIdx >= len(target.Hand) {
			return fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		card := target.Hand[cardIdx]
		target.Hand = append(target.Hand[:cardIdx], target.Hand[cardIdx+1:]...)
		rt.NotifyCardRevealed(target.ID, []model.Card{card}, "discard")
		rt.AppendToDiscard([]model.Card{card})
		rt.Log(fmt.Sprintf("%s 的 [灵符-风行]：%s 选择弃置了1张手牌", user.Name, target.Name))
	}

	nextCursor := cursor + 1
	for nextCursor < len(ordered) {
		nextTarget := rt.GetPlayers()[ordered[nextCursor]]
		if nextTarget == nil {
			nextCursor++
			continue
		}
		if len(nextTarget.Hand) <= 0 {
			rt.Log(fmt.Sprintf("%s 的 [灵符-风行]：%s 无手牌可弃置", user.Name, nextTarget.Name))
			nextCursor++
			continue
		}
		ctxData["cursor"] = nextCursor
		ctxData["current_target_id"] = nextTarget.ID
		if intr := rt.GetPendingInterrupt(); intr != nil {
			intr.Context = ctxData
		}
		// Update the interrupt's PlayerID so the prompt goes to the right player
		if intr := rt.GetPendingInterrupt(); intr != nil {
			intr.PlayerID = nextTarget.ID
		}
		rt.NotifyInterruptPrompt()
		return nil
	}

	rt.Log(fmt.Sprintf("%s 的 [灵符-风行] 结算完成", user.Name))
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil && len(rt.GetPendingDamageQueue()) > 0 {
		rt.EnterDamageResolution(nil)
	}
	return nil
}

func handleTalismanPick(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	targetIDs := runtimeutil.DedupeIDs(runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"]))
	switch selectionIndex {
	case 0:
		// Thunder talisman
		rt.PopInterrupt()
		return continueSpiritCasterResolution(rt, user, "sc_talisman_thunder", targetIDs)
	case 1:
		// Wind talisman
		rt.PopInterrupt()
		return continueSpiritCasterResolution(rt, user, "sc_talisman_wind", targetIDs)
	default:
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
}

// ===========================================================================
// Talisman flow continuation
// ===========================================================================

// continueSpiritCasterResolution handles the post-incantation flow for both
// talisman types, pushing the appropriate next interrupt.
func continueSpiritCasterResolution(rt engineplayer.ChoiceRuntime, user *model.Player, skillID string, targetIDs []string) error {
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	switch skillID {
	case "sc_talisman_thunder":
		if rt.CanPayCrystalCost(user.ID, 1) {
			rt.PushInterrupt(&model.Interrupt{
				Type:     model.InterruptChoice,
				PlayerID: user.ID,
				Context: map[string]interface{}{
					"choice_type": "sc_spiritual_collapse_confirm",
					"user_id":     user.ID,
					"mode":        "sc_talisman_thunder",
					"target_ids":  append([]string{}, targetIDs...),
				},
			})
			return nil
		}
		resolveThunderDamage(rt, user, targetIDs, 0)
	case "sc_talisman_wind":
		return startWindDiscardFlow(rt, user, targetIDs)
	default:
		return fmt.Errorf("未知灵符技能: %s", skillID)
	}
	return nil
}

// startWindDiscardFlow sets up the iterative wind-discard interrupt chain.
func startWindDiscardFlow(rt engineplayer.ChoiceRuntime, user *model.Player, targetIDs []string) error {
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	targetSet := runtimeutil.IDsToSet(runtimeutil.DedupeIDs(targetIDs))
	orderedAll := engineplayer.ReversePlayerIDsFromRuntime(rt, user.ID, engineplayer.ReverseOrderOption{IncludeSelf: true})
	ordered := make([]string, 0, len(targetIDs))
	for _, playerID := range orderedAll {
		if !targetSet[playerID] {
			continue
		}
		ordered = append(ordered, playerID)
	}
	if len(ordered) == 0 {
		rt.Log(fmt.Sprintf("%s 的 [灵符-风行]：无有效目标", user.Name))
		return nil
	}

	cursor := 0
	for cursor < len(ordered) {
		target := rt.GetPlayers()[ordered[cursor]]
		if target == nil || len(target.Hand) == 0 {
			if target != nil {
				rt.Log(fmt.Sprintf("%s 的 [灵符-风行]：%s 无手牌可弃置", user.Name, target.Name))
			}
			cursor++
			continue
		}
		break
	}
	if cursor >= len(ordered) {
		rt.Log(fmt.Sprintf("%s 的 [灵符-风行]：所有目标均无手牌可弃置", user.Name))
		return nil
	}

	currentTargetID := ordered[cursor]
	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: currentTargetID,
		Context: map[string]interface{}{
			"choice_type":        "sc_talisman_wind_discard",
			"user_id":            user.ID,
			"ordered_target_ids": ordered,
			"cursor":             cursor,
			"current_target_id":  currentTargetID,
		},
	})
	return nil
}

// ===========================================================================
// Damage resolution helpers
// ===========================================================================

// resolveThunderDamage deals magic damage to the given targets in reverse
// order, then routes pending damage.
func resolveThunderDamage(rt engineplayer.ChoiceRuntime, user *model.Player, targetIDs []string, bonus int) {
	if user == nil {
		return
	}
	damage := 1 + bonus
	if damage < 0 {
		damage = 0
	}
	targetSet := runtimeutil.IDsToSet(runtimeutil.DedupeIDs(targetIDs))
	ordered := engineplayer.ReversePlayerIDsFromRuntime(rt, user.ID, engineplayer.ReverseOrderOption{IncludeSelf: true})
	hitCount := 0
	for _, targetID := range ordered {
		if !targetSet[targetID] {
			continue
		}
		rt.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   targetID,
			Damage:     damage,
			DamageType: model.MagicAttack,
		})
		hitCount++
	}
	rt.Log(fmt.Sprintf("%s 发动 [灵符-雷鸣]：对%d名角色各造成%d点法术伤害", user.Name, hitCount, damage))
	if len(rt.GetPendingDamageQueue()) > 0 {
		rt.RoutePendingDamageWithReturn(model.TurnStageExtraAction)
	}
}

// resolveHundredNightSingle deals magic damage to a single target.
func resolveHundredNightSingle(rt engineplayer.ChoiceRuntime, user *model.Player, targetID string, bonus int) error {
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	target := rt.GetPlayers()[targetID]
	if target == nil {
		return fmt.Errorf("目标不存在")
	}
	damage := 1 + bonus
	rt.AddPendingDamage(model.PendingDamage{
		SourceID:   user.ID,
		TargetID:   target.ID,
		Damage:     damage,
		DamageType: model.MagicAttack,
	})
	rt.Log(fmt.Sprintf("%s 发动 [百鬼夜行]：对 %s 造成%d点法术伤害", user.Name, target.Name, damage))
	return nil
}

// resolveHundredNightFireAOE deals magic damage to all players except the
// excluded ones.
func resolveHundredNightFireAOE(rt engineplayer.ChoiceRuntime, user *model.Player, excludeIDs []string, bonus int) error {
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	exclude := runtimeutil.IDsToSet(runtimeutil.DedupeIDs(excludeIDs))
	damage := 1 + bonus
	ordered := engineplayer.ReversePlayerIDsFromRuntime(rt, user.ID, engineplayer.ReverseOrderOption{IncludeSelf: true})
	hitCount := 0
	for _, playerID := range ordered {
		if exclude[playerID] {
			continue
		}
		target := rt.GetPlayers()[playerID]
		if target == nil {
			continue
		}
		rt.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   target.ID,
			Damage:     damage,
			DamageType: model.MagicAttack,
		})
		hitCount++
	}
	rt.Log(fmt.Sprintf("%s 发动 [百鬼夜行·火]：对除2名指定角色外的其他角色各造成%d点法术伤害（命中%d名）", user.Name, damage, hitCount))
	return nil
}

// ===========================================================================
// Utility helpers
// ===========================================================================

