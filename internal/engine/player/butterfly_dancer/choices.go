// gameflow: 蝶舞者角色选择流。

package butterfly_dancer

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
	case "bt_dance_mode":
		return buildDanceModePrompt(playerID, data)
	case "bt_dance_discard":
		return buildDanceDiscardPrompt(playerID, player)
	case "bt_cocoon_overflow_discard":
		return buildCocoonOverflowDiscardPrompt(playerID, player, data)
	case "bt_reverse_mode":
		return buildReverseModePrompt(playerID, data)
	case "bt_reverse_target":
		return buildReverseTargetPrompt(rt, playerID, data)
	case "bt_reverse_branch2_cost":
		return buildReverseBranch2CostPrompt(playerID, data)
	case "bt_reverse_branch2_pick":
		return buildReverseBranch2PickPrompt(playerID, player)
	case "bt_pilgrimage_pick", "bt_poison_pick":
		return buildPilgrimageOrPoisonPickPrompt(playerID, player, data, choiceType)
	case "bt_mirror_pair":
		return buildMirrorPairPrompt(playerID, data)
	case "bt_wither_confirm":
		return buildWitherConfirmPrompt(playerID)
	case "bt_wither_target":
		return buildWitherTargetPrompt(rt, playerID, data)
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
	case "bt_dance_mode":
		return true, handleDanceMode(rt, ctxData, selectionIndex)
	case "bt_dance_discard":
		return true, handleDanceDiscard(rt, ctxData, selectionIndex)
	case "bt_cocoon_overflow_discard":
		return true, handleCocoonOverflowDiscard(rt, ctxData, selectionIndex)
	case "bt_reverse_mode":
		return true, handleReverseMode(rt, ctxData, selectionIndex)
	case "bt_reverse_target":
		return true, handleReverseTarget(rt, ctxData, selectionIndex)
	case "bt_reverse_branch2_cost":
		return true, handleReverseBranch2Cost(rt, ctxData, selectionIndex)
	case "bt_reverse_branch2_pick":
		return true, handleReverseBranch2Pick(rt, ctxData, selectionIndex)
	case "bt_pilgrimage_pick", "bt_poison_pick":
		return true, handlePilgrimageOrPoisonPick(rt, ctxData, selectionIndex)
	case "bt_mirror_pair":
		return true, handleMirrorPair(rt, ctxData, selectionIndex)
	case "bt_wither_confirm":
		return true, handleWitherConfirm(rt, ctxData, selectionIndex)
	case "bt_wither_target":
		return true, handleWitherTarget(rt, ctxData, selectionIndex)
	default:
		return false, nil
	}
}

// ===========================================================================
// BuildPrompt helpers
// ===========================================================================

func buildDanceModePrompt(playerID string, data map[string]interface{}) *model.Prompt {
	canDiscard := runtimeutil.ToBoolContextValue(data["can_discard"])
	options := []model.PromptOption{{ID: "0", Label: "摸1张牌"}}
	if canDiscard {
		options = append(options, model.PromptOption{ID: "1", Label: "弃1张牌"})
	}
	return &model.Prompt{
		Type:       model.PromptConfirm,
		PlayerID:   playerID,
		ChoiceType: "bt_dance_mode",
		Message:    "【舞动】请选择先执行的动作：",
		Options:    options,
		Min:        1,
		Max:        1,
	}
}

func buildDanceDiscardPrompt(playerID string, player *model.Player) *model.Prompt {
	var options []model.PromptOption
	for idx, c := range player.Hand {
		options = append(options, model.PromptOption{
			ID:    fmt.Sprintf("%d", idx),
			Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(c)),
		})
	}
	return &model.Prompt{
		Type:       model.PromptChooseCards,
		PlayerID:   playerID,
		ChoiceType: "bt_dance_discard",
		Message:    "【舞动】请选择要弃置的1张手牌：",
		Options:    options,
		Min:        1,
		Max:        1,
	}
}

func buildCocoonOverflowDiscardPrompt(playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	discardCount := runtimeutil.ToIntContextValue(data["discard_count"])
	if discardCount < 0 {
		discardCount = 0
	}
	cocoonIndices := CocoonFieldIndices(player)
	if discardCount > len(cocoonIndices) {
		discardCount = len(cocoonIndices)
	}
	var options []model.PromptOption
	for _, idx := range cocoonIndices {
		if idx < 0 || idx >= len(player.Field) || player.Field[idx] == nil {
			continue
		}
		fc := player.Field[idx]
		if fc.Mode != model.FieldCover || fc.Effect != model.EffectButterflyCocoon {
			continue
		}
		options = append(options, model.PromptOption{
			ID:    fmt.Sprintf("%d", idx),
			Label: fmt.Sprintf("茧[%d]: %s", idx, formatCardInfo(fc.Card)),
		})
	}
	return &model.Prompt{
		Type:       model.PromptChooseCards,
		PlayerID:   playerID,
		ChoiceType: "bt_cocoon_overflow_discard",
		Message:    fmt.Sprintf("【茧上限】请选择要舍弃的%d个茧：", discardCount),
		Options:    options,
		Min:        discardCount,
		Max:        discardCount,
	}
}

func buildReverseModePrompt(playerID string, data map[string]interface{}) *model.Prompt {
	canBranch2 := runtimeutil.ToBoolContextValue(data["can_branch2"])
	options := []model.PromptOption{{ID: "0", Label: "分支①：对目标造成1点不可治疗抵御的法术伤害"}}
	if canBranch2 {
		options = append(options, model.PromptOption{ID: "1", Label: "分支②：移除2个茧或自伤4，然后移除1个蛹"})
	}
	return &model.Prompt{
		Type:       model.PromptConfirm,
		PlayerID:   playerID,
		ChoiceType: "bt_reverse_mode",
		Message:    "【倒逆之蝶】请选择发动分支：",
		Options:    options,
		Min:        1,
		Max:        1,
	}
}

func buildReverseTargetPrompt(rt engineplayer.ChoiceRuntime, playerID string, data map[string]interface{}) *model.Prompt {
	targetIDs := parseStringSlice(data["target_ids"])
	var options []model.PromptOption
	for _, tid := range targetIDs {
		if p := rt.GetPlayers()[tid]; p != nil {
			options = append(options, model.PromptOption{ID: tid, Label: p.Name})
		}
	}
	return &model.Prompt{
		Type:       model.PromptConfirm,
		PlayerID:   playerID,
		ChoiceType: "bt_reverse_target",
		Message:    "【倒逆之蝶】请选择分支①伤害目标：",
		Options:    options,
		Min:        1,
		Max:        1,
	}
}

func buildReverseBranch2CostPrompt(playerID string, data map[string]interface{}) *model.Prompt {
	canRemove := runtimeutil.ToBoolContextValue(data["can_remove_cocoon"])
	options := []model.PromptOption{}
	if canRemove {
		options = append(options, model.PromptOption{ID: "0", Label: "移除2个茧"})
	}
	options = append(options, model.PromptOption{ID: "1", Label: "对自己造成4点法术伤害"})
	return &model.Prompt{
		Type:       model.PromptConfirm,
		PlayerID:   playerID,
		ChoiceType: "bt_reverse_branch2_cost",
		Message:    "【倒逆之蝶】请选择分支②代价：",
		Options:    options,
		Min:        1,
		Max:        1,
	}
}

func buildReverseBranch2PickPrompt(playerID string, player *model.Player) *model.Prompt {
	cocoonIndices := CocoonFieldIndices(player)
	pickCount := 2
	if pickCount > len(cocoonIndices) {
		pickCount = len(cocoonIndices)
	}
	var options []model.PromptOption
	for _, idx := range cocoonIndices {
		if idx < 0 || idx >= len(player.Field) || player.Field[idx] == nil {
			continue
		}
		fc := player.Field[idx]
		if fc.Mode != model.FieldCover || fc.Effect != model.EffectButterflyCocoon {
			continue
		}
		options = append(options, model.PromptOption{
			ID:    fmt.Sprintf("%d", idx),
			Label: fmt.Sprintf("茧[%d]: %s", idx, formatCardInfo(fc.Card)),
		})
	}
	return &model.Prompt{
		Type:       model.PromptChooseCards,
		PlayerID:   playerID,
		ChoiceType: "bt_reverse_branch2_pick",
		Message:    fmt.Sprintf("【倒逆之蝶】分支②请选择要移除的%d个茧：", pickCount),
		Options:    options,
		Min:        pickCount,
		Max:        pickCount,
	}
}

func buildPilgrimageOrPoisonPickPrompt(playerID string, player *model.Player, data map[string]interface{}, choiceType string) *model.Prompt {
	cocoonIndices := parseIntSlice(data["cocoon_indices"])
	options := []model.PromptOption{{ID: "-1", Label: "不发动"}}
	for _, idx := range cocoonIndices {
		if idx < 0 || idx >= len(player.Field) || player.Field[idx] == nil {
			continue
		}
		fc := player.Field[idx]
		if fc.Mode != model.FieldCover || fc.Effect != model.EffectButterflyCocoon {
			continue
		}
		options = append(options, model.PromptOption{
			ID:    fmt.Sprintf("%d", len(options)),
			Label: fmt.Sprintf("移除茧[%d]: %s", idx, formatCardInfo(fc.Card)),
		})
	}
	msg := "【朝圣】是否移除1个茧抵御1点伤害？"
	if choiceType == "bt_poison_pick" {
		msg = "【毒粉】是否移除1个茧使该次法术伤害+1？"
	}
	return &model.Prompt{
		Type:       model.PromptConfirm,
		PlayerID:   playerID,
		ChoiceType: choiceType,
		Message:    msg,
		Options:    options,
		Min:        1,
		Max:        1,
	}
}

func buildMirrorPairPrompt(playerID string, data map[string]interface{}) *model.Prompt {
	labels := parseStringSlice(data["pair_labels"])
	options := []model.PromptOption{{ID: "-1", Label: "不发动"}}
	for i, label := range labels {
		options = append(options, model.PromptOption{
			ID:    fmt.Sprintf("%d", i+1),
			Label: fmt.Sprintf("移除并展示：%s", label),
		})
	}
	return &model.Prompt{
		Type:       model.PromptConfirm,
		PlayerID:   playerID,
		ChoiceType: "bt_mirror_pair",
		Message:    "【镜花水月】是否发动并改写该次2点法术伤害？",
		Options:    options,
		Min:        1,
		Max:        1,
	}
}

func buildWitherConfirmPrompt(playerID string) *model.Prompt {
	return &model.Prompt{
		Type:       model.PromptConfirm,
		PlayerID:   playerID,
		ChoiceType: "bt_wither_confirm",
		Message:    "【凋零】可发动：是否对目标造成1点法术伤害并对自己造成2点法术伤害？",
		Options: []model.PromptOption{
			{ID: "0", Label: "发动凋零"},
			{ID: "1", Label: "不发动"},
		},
		Min: 1,
		Max: 1,
	}
}

func buildWitherTargetPrompt(rt engineplayer.ChoiceRuntime, playerID string, data map[string]interface{}) *model.Prompt {
	targetIDs := parseStringSlice(data["target_ids"])
	var options []model.PromptOption
	for _, tid := range targetIDs {
		if p := rt.GetPlayers()[tid]; p != nil {
			options = append(options, model.PromptOption{ID: tid, Label: p.Name})
		}
	}
	return &model.Prompt{
		Type:       model.PromptConfirm,
		PlayerID:   playerID,
		ChoiceType: "bt_wither_target",
		Message:    "【凋零】请选择1名目标角色：",
		Options:    options,
		Min:        1,
		Max:        1,
	}
}

// ===========================================================================
// HandleChoice handlers
// ===========================================================================

func handleDanceMode(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	modes := []string{"draw"}
	if runtimeutil.ToBoolContextValue(ctxData["can_discard"]) {
		modes = append(modes, "discard")
	}
	if selectionIndex < 0 || selectionIndex >= len(modes) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	mode := modes[selectionIndex]
	if mode == "discard" {
		if len(user.Hand) <= 0 {
			return fmt.Errorf("手牌不足，无法弃牌")
		}
		ctxData["choice_type"] = "bt_dance_discard"
		intr := rt.GetPendingInterrupt()
		if intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return nil
	}
	// Draw mode: draw 1 card to hand, draw 1 cocoon
	rt.DrawCards(user.ID, 1)
	rt.DrawCardsWithOptions(user.ID, 1, model.DrawOptions{Reason: "bt_dance_cocoon"})
	// The DrawCardsWithOptions call draws into hand. For cocoon placement we
	// need to take the last card added and place it as cocoon instead.
	// However, the IGameEngine interface doesn't have a "draw to cocoon" method.
	// The original code uses rules.DrawCards to draw directly from deck state.
	// Since we can't access deck directly, we use an alternative approach:
	// draw to hand, then remove the last card and place as cocoon.
	if len(user.Hand) > 0 {
		lastCard := user.Hand[len(user.Hand)-1]
		user.Hand = user.Hand[:len(user.Hand)-1]
		AddCocoonCards(user, []model.Card{lastCard})
	}
	rt.Log(fmt.Sprintf("%s 发动 [舞动]：摸1张牌，并将牌库顶1张牌放置为茧", user.Name))
	rt.CheckHandLimit(user.ID, false)
	overflow := CocoonCount(user) - ButterflyCocoonCap
	if overflow > 0 {
		rt.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: user.ID,
			Context: map[string]interface{}{
				"choice_type":   "bt_cocoon_overflow_discard",
				"user_id":       user.ID,
				"discard_count": overflow,
			},
		})
	}
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		if !rt.RoutePendingDamageWithReturn(model.TurnStageExtraAction) {
			rt.EnterExtraActionStage()
		}
	}
	return nil
}

func handleDanceDiscard(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	if selectionIndex < 0 || selectionIndex >= len(user.Hand) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	card := user.Hand[selectionIndex]
	user.Hand = append(user.Hand[:selectionIndex], user.Hand[selectionIndex+1:]...)
	rt.NotifyCardRevealed(user.ID, []model.Card{card}, model.DamageType("discard"))
	rt.AppendToDiscard([]model.Card{card})

	// Draw 1 cocoon from deck
	rt.DrawCardsWithOptions(user.ID, 1, model.DrawOptions{Reason: "bt_dance_cocoon"})
	if len(user.Hand) > 0 {
		lastCard := user.Hand[len(user.Hand)-1]
		user.Hand = user.Hand[:len(user.Hand)-1]
		AddCocoonCards(user, []model.Card{lastCard})
	}

	rt.Log(fmt.Sprintf("%s 发动 [舞动]：弃1张牌，并将牌库顶1张牌放置为茧", user.Name))
	overflow := CocoonCount(user) - ButterflyCocoonCap
	if overflow > 0 {
		rt.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: user.ID,
			Context: map[string]interface{}{
				"choice_type":   "bt_cocoon_overflow_discard",
				"user_id":       user.ID,
				"discard_count": overflow,
			},
		})
	}
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		rt.EnterExtraActionStage()
	}
	return nil
}

func handleCocoonOverflowDiscard(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	if selectionIndex < 0 {
		return fmt.Errorf("请先选择要舍弃的茧后再确认")
	}
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	discardNeed := runtimeutil.ToIntContextValue(ctxData["discard_count"])
	if discardNeed < 0 {
		discardNeed = 0
	}
	cocoonIndices := CocoonFieldIndices(user)
	if discardNeed > len(cocoonIndices) {
		discardNeed = len(cocoonIndices)
	}
	if discardNeed != 1 {
		return fmt.Errorf("需要选择 %d 个茧舍弃", discardNeed)
	}
	fieldIdx, ok := runtimeutil.ResolveSelectionToCandidate(selectionIndex, cocoonIndices)
	if !ok {
		return fmt.Errorf("无效的茧索引: %d", selectionIndex)
	}

	// Remove cocoon at fieldIdx
	fc, ok := RemoveCocoonByFieldIndex(user, fieldIdx)
	if !ok {
		return fmt.Errorf("选择的茧无效")
	}
	removed := []model.Card{fc.Card}
	rt.AppendToDiscard(removed)

	rt.Log(fmt.Sprintf("%s 的 [茧上限] 结算：舍弃1个茧", user.Name))
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		rt.EnterExtraActionStage()
	}
	return nil
}

func handleReverseMode(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	canBranch2 := Pupa(user) > 0
	modes := []string{"branch1"}
	if canBranch2 {
		modes = append(modes, "branch2")
	}
	if selectionIndex < 0 || selectionIndex >= len(modes) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	if modes[selectionIndex] == "branch1" {
		ctxData["choice_type"] = "bt_reverse_target"
		intr := rt.GetPendingInterrupt()
		if intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return nil
	}
	// branch2
	if Pupa(user) <= 0 {
		return fmt.Errorf("蛹不足，无法发动分支②")
	}
	ctxData["choice_type"] = "bt_reverse_branch2_cost"
	ctxData["can_remove_cocoon"] = CocoonCount(user) >= 2
	intr := rt.GetPendingInterrupt()
	if intr != nil {
		intr.Context = ctxData
	}
	rt.NotifyInterruptPrompt()
	return nil
}

func handleReverseTarget(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	targetIDs := parseStringSlice(ctxData["target_ids"])
	if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	targetID := targetIDs[selectionIndex]
	rt.AddPendingDamage(model.PendingDamage{
		SourceID:   user.ID,
		TargetID:   targetID,
		Damage:     1,
		DamageType: model.MagicAttack,
		IgnoreHeal: true,
	})
	if target := rt.GetPlayers()[targetID]; target != nil {
		rt.Log(fmt.Sprintf("%s 的 [倒逆之蝶] 分支①：对 %s 造成1点不可治疗抵御的法术伤害", user.Name, target.Name))
	}
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		if !rt.RoutePendingDamageWithReturn(model.TurnStageExtraAction) {
			rt.EnterExtraActionStage()
		}
	}
	return nil
}

func handleReverseBranch2Cost(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	canRemove := runtimeutil.ToBoolContextValue(ctxData["can_remove_cocoon"])
	modes := []string{}
	if canRemove {
		modes = append(modes, "remove_cocoon")
	}
	modes = append(modes, "self_damage")
	if selectionIndex < 0 || selectionIndex >= len(modes) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	if modes[selectionIndex] == "remove_cocoon" {
		ctxData["choice_type"] = "bt_reverse_branch2_pick"
		delete(ctxData, "remaining_indices")
		delete(ctxData, "selected_indices")
		intr := rt.GetPendingInterrupt()
		if intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return nil
	}
	// self_damage branch
	rt.AddPendingDamage(model.PendingDamage{
		SourceID:   user.ID,
		TargetID:   user.ID,
		Damage:     4,
		DamageType: model.MagicAttack,
	})
	now := AddPupa(user, -1)
	rt.Log(fmt.Sprintf("%s 的 [倒逆之蝶] 分支②：对自己造成4点法术伤害并移除1个蛹（当前蛹=%d）", user.Name, now))
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		if !rt.RoutePendingDamageWithReturn(model.TurnStageExtraAction) {
			rt.EnterExtraActionStage()
		}
	}
	return nil
}

func handleReverseBranch2Pick(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	if selectionIndex < 0 {
		return fmt.Errorf("请先选择要移除的茧后再确认")
	}
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	const pickNeed = 2
	cocoonIndices := CocoonFieldIndices(user)

	fieldIdx, ok := runtimeutil.ResolveSelectionToCandidate(selectionIndex, cocoonIndices)
	if !ok {
		return fmt.Errorf("无效的茧索引: %d", selectionIndex)
	}

	// Collect picked indices from context
	picked := append([]int{}, parseIntSlice(ctxData["picked_indices"])...)
	picked = append(picked, fieldIdx)

	if len(picked) < pickNeed {
		// Need more picks - update context and re-prompt
		ctxData["picked_indices"] = picked
		intr := rt.GetPendingInterrupt()
		if intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return nil
	}

	// Have enough picks - resolve
	RemoveCocoonByFieldIndices(user, picked)
	// Collect removed cards for notification
	var removed []model.Card
	for _, idx := range picked {
		// Cards were already removed by RemoveCocoonByFieldIndices, we log the effect
		_ = idx
	}
	rt.NotifyCardRevealed(user.ID, removed, model.DamageType("discard"))
	rt.AppendToDiscard(removed)

	// Check for magic cards among removed to queue wither
	for _, c := range removed {
		if c.Type == model.CardTypeMagic {
			queueWitherChoice(rt, user)
		}
	}
	now := AddPupa(user, -1)
	rt.Log(fmt.Sprintf("%s 的 [倒逆之蝶] 分支②：移除2个茧并移除1个蛹（当前蛹=%d）", user.Name, now))
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		if !rt.RoutePendingDamageWithReturn(model.TurnStageExtraAction) {
			rt.EnterExtraActionStage()
		}
	}
	return nil
}

func handlePilgrimageOrPoisonPick(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	choiceType, _ := ctxData["choice_type"].(string)
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	cocoonIndices := parseIntSlice(ctxData["cocoon_indices"])

	// selectionIndex == -1 or 0 means "skip"
	if selectionIndex == -1 || selectionIndex == 0 {
		rt.PopInterrupt()
		if rt.GetPendingInterrupt() == nil {
			if !rt.RoutePendingDamageOr(nil, nil) {
				rt.EnterExtraActionStage()
			}
		}
		return nil
	}

	pickIdx := -1
	if selectionIndex >= 1 && selectionIndex <= len(cocoonIndices) {
		pickIdx = cocoonIndices[selectionIndex-1]
	} else if idx, ok := runtimeutil.ResolveSelectionToCandidate(selectionIndex, cocoonIndices); ok {
		pickIdx = idx
	}
	if pickIdx < 0 {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}

	removedFC, ok := RemoveCocoonByFieldIndex(user, pickIdx)
	if !ok {
		return fmt.Errorf("选择的茧无效")
	}
	removedCard := removedFC.Card
	rt.NotifyCardRevealed(user.ID, []model.Card{removedCard}, model.DamageType("discard"))
	rt.AppendToDiscard([]model.Card{removedCard})

	damageIdx := runtimeutil.ToIntContextValue(ctxData["damage_index"])
	pd, ok := rt.GetPendingDamageByIndex(damageIdx)
	if !ok {
		return fmt.Errorf("伤害上下文不存在")
	}

	if choiceType == "bt_pilgrimage_pick" {
		if pd.Damage > 0 {
			pd.Damage--
		}
		rt.Log(fmt.Sprintf("%s 发动 [朝圣]：移除1个茧，抵御1点伤害（剩余伤害=%d）", user.Name, pd.Damage))
	} else {
		pd.Damage++
		rt.Log(fmt.Sprintf("%s 发动 [毒粉]：移除1个茧，本次法术伤害+1（当前伤害=%d）", user.Name, pd.Damage))
	}

	if removedCard.Type == model.CardTypeMagic {
		queueWitherChoice(rt, user)
	}

	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		rt.EnterDamageResolution(nil)
	}
	return nil
}

func handleMirrorPair(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}

	// selectionIndex == -1 or 0 means "skip"
	if selectionIndex == -1 || selectionIndex == 0 {
		rt.PopInterrupt()
		if rt.GetPendingInterrupt() == nil {
			if !rt.RoutePendingDamageOr(nil, nil) {
				rt.EnterExtraActionStage()
			}
		}
		return nil
	}

	pairDefs := parseStringSlice(ctxData["pair_defs"])
	pairChoice := -1
	if selectionIndex >= 1 && selectionIndex <= len(pairDefs) {
		pairChoice = selectionIndex - 1
	} else if selectionIndex >= 0 && selectionIndex < len(pairDefs) {
		pairChoice = selectionIndex
	}
	if pairChoice < 0 || pairChoice >= len(pairDefs) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}

	parts := strings.Split(pairDefs[pairChoice], ",")
	if len(parts) != 2 {
		return fmt.Errorf("镜花水月配对参数无效")
	}
	left, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	right, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil {
		return fmt.Errorf("镜花水月配对索引无效")
	}

	// Collect cocoon cards at those indices before removing.
	// Use pointer-based removal to avoid index shifting issues when
	// removing the first cocoon shifts the second one's field position.
	var removedCards []model.Card
	collected := make(map[int]*model.FieldCard)
	for _, idx := range []int{left, right} {
		if idx >= 0 && idx < len(user.Field) && user.Field[idx] != nil &&
			user.Field[idx].Mode == model.FieldCover && user.Field[idx].Effect == model.EffectButterflyCocoon {
			collected[idx] = user.Field[idx]
			removedCards = append(removedCards, user.Field[idx].Card)
		}
	}
	if len(collected) != 2 || removedCards[0].Element != removedCards[1].Element {
		return fmt.Errorf("镜花水月需要移除2张同系茧")
	}
	for _, fc := range collected {
		user.RemoveFieldCard(fc)
	}
	SyncCocoonToken(user)

	rt.NotifyCardRevealed(user.ID, removedCards, model.DamageType("discard"))
	rt.AppendToDiscard(removedCards)

	damageIdx := runtimeutil.ToIntContextValue(ctxData["damage_index"])
	pd, ok := rt.GetPendingDamageByIndex(damageIdx)
	if !ok {
		return fmt.Errorf("伤害上下文不存在")
	}
	originSourceID := pd.SourceID
	pd.Damage = 0

	rt.AddPendingDamage(model.PendingDamage{
		SourceID:   user.ID,
		TargetID:   originSourceID,
		Damage:     1,
		DamageType: model.MagicAttack,
	})
	rt.AddPendingDamage(model.PendingDamage{
		SourceID:   user.ID,
		TargetID:   originSourceID,
		Damage:     1,
		DamageType: model.MagicAttack,
	})

	for _, c := range removedCards {
		if c.Type == model.CardTypeMagic {
			queueWitherChoice(rt, user)
		}
	}

	if target := rt.GetPlayers()[originSourceID]; target != nil {
		rt.Log(fmt.Sprintf("%s 发动 [镜花水月]：抵御原伤害，并改为对 %s 造成2次1点法术伤害", user.Name, target.Name))
	}

	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		rt.EnterDamageResolution(nil)
	}
	return nil
}

func handleWitherConfirm(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	if user.Tokens == nil {
		user.Tokens = map[string]int{}
	}
	if selectionIndex == 0 {
		// Activate wither - ask for target
		ctxData["choice_type"] = "bt_wither_target"
		ctxData["target_ids"] = allPlayerIDs(rt)
		intr := rt.GetPendingInterrupt()
		if intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return nil
	}
	if selectionIndex != 1 {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	// Decline wither
	if user.TurnState.SkillFlowState == nil {
		user.TurnState.SkillFlowState = map[string]int{}
	}
	if user.TurnState.SkillFlowState["bt_wither_pending"] > 0 {
		user.TurnState.SkillFlowState["bt_wither_pending"]--
	}
	if user.TurnState.SkillFlowState["bt_wither_pending"] > 0 {
		ctxData["choice_type"] = "bt_wither_confirm"
		ctxData["target_ids"] = allPlayerIDs(rt)
		intr := rt.GetPendingInterrupt()
		if intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return nil
	}
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		if !rt.RoutePendingDamageOr(nil, nil) {
			rt.EnterExtraActionStage()
		}
	}
	return nil
}

func handleWitherTarget(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, selectionIndex int) error {
	userID, _ := ctxData["user_id"].(string)
	user := rt.GetPlayers()[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	targetIDs := parseStringSlice(ctxData["target_ids"])
	if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
		return fmt.Errorf("无效的选项索引: %d", selectionIndex)
	}
	targetID := targetIDs[selectionIndex]

	rt.AddPendingDamage(model.PendingDamage{
		SourceID:   user.ID,
		TargetID:   targetID,
		Damage:     1,
		DamageType: model.MagicAttack,
	})
	rt.AddPendingDamage(model.PendingDamage{
		SourceID:   user.ID,
		TargetID:   user.ID,
		Damage:     2,
		DamageType: model.MagicAttack,
	})
	engineplayer.EnsurePlayerSkillFlowState(user)
	user.TurnState.SkillFlowState["bt_wither_active"] = 1

	if user.TurnState.SkillFlowState == nil {
		user.TurnState.SkillFlowState = map[string]int{}
	}
	if user.TurnState.SkillFlowState["bt_wither_pending"] > 0 {
		user.TurnState.SkillFlowState["bt_wither_pending"]--
	}

	if target := rt.GetPlayers()[targetID]; target != nil {
		rt.Log(fmt.Sprintf("%s 发动 [凋零]：对 %s 造成1点法术伤害，并对自己造成2点法术伤害；对方士气最低为1直到其下回合开始前", user.Name, target.Name))
	}

	if user.TurnState.SkillFlowState["bt_wither_pending"] > 0 {
		ctxData["choice_type"] = "bt_wither_confirm"
		ctxData["target_ids"] = allPlayerIDs(rt)
		intr := rt.GetPendingInterrupt()
		if intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return nil
	}

	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		rt.EnterDamageResolution(nil)
	}
	return nil
}

// ===========================================================================
// Local helpers
// ===========================================================================

func formatCardInfo(card model.Card) string {
	return promptfmt.FormatCardInfo(card)
}

// allPlayerIDs returns all player IDs in order.
func allPlayerIDs(rt engineplayer.ChoiceRuntime) []string {
	var ids []string
	for _, pid := range rt.GetPlayerOrder() {
		if rt.GetPlayers()[pid] != nil {
			ids = append(ids, pid)
		}
	}
	return ids
}

// parseStringSlice extracts a []string from an interface{}, handling both
// []string and []interface{} slices.
func parseStringSlice(v interface{}) []string {
	var out []string
	switch arr := v.(type) {
	case []string:
		out = append(out, arr...)
	case []interface{}:
		for _, item := range arr {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
	}
	return out
}

// parseIntSlice extracts a []int from an interface{}, handling both
// []int and []interface{} slices.
func parseIntSlice(v interface{}) []int {
	var out []int
	switch arr := v.(type) {
	case []int:
		out = append(out, arr...)
	case []interface{}:
		for _, item := range arr {
			if f, ok := item.(float64); ok {
				out = append(out, int(f))
			}
		}
	}
	return out
}

// queueWitherChoice 在魔法茧移除后询问蝶舞者是否发动枯萎。
func queueWitherChoice(rt engineplayer.ChoiceRuntime, user *model.Player) {
	if user == nil {
		return
	}
	if user.TurnState.SkillFlowState == nil {
		user.TurnState.SkillFlowState = map[string]int{}
	}
	user.TurnState.SkillFlowState["bt_wither_pending"]++
	if user.TurnState.SkillFlowState["bt_wither_pending"] > 1 {
		return
	}
	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: user.ID,
		Context: map[string]interface{}{
			"choice_type": "bt_wither_confirm",
			"user_id":     user.ID,
			"target_ids":  allPlayerIDs(rt),
		},
	})
}
