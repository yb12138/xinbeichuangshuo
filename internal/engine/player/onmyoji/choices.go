// gameflow: 鬼术师角色选择流。

package onmyoji

import (
	"fmt"
	"strconv"
	"strings"

	"starcup-engine/internal/engine/core/runtimeutil"
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

type choiceHandler struct{}

func NewChoiceHandler() engineplayer.ChoiceHandler {
	return choiceHandler{}
}

type onmyojiCardOption struct {
	CardID     string
	UseFaction bool
	Label      string
}

func (choiceHandler) BuildPrompt(rt engineplayer.ChoiceRuntime, choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "onmyoji_life_barrier_mode":
		ghostFire := runtimeutil.ToIntContextValue(data["ghost_fire"])
		options := []model.PromptOption{
			{ID: "0", Label: "分支①：1名队友+1宝石+1治疗，自己承受X点法伤"},
		}
		releaseCombos := len(runtimeutil.ParseStringSliceContextValue(data["release_card_combos"]))
		if releaseCombos > 0 {
			options = append(options, model.PromptOption{
				ID:    "1",
				Label: "分支②：弃2张同命格手牌并脱离式神形态，令1名队友弃1张手牌",
			})
		}
		return &model.Prompt{
			Type:          model.PromptConfirm,
			PlayerID:      playerID,
			Message:       fmt.Sprintf("【生命结界】当前鬼火=%d，请选择发动分支：", ghostFire),
			Options:       options,
			Min:           1,
			Max:           1,
			Presentation:  &model.PromptPresentation{Kind: model.PresentationBranchSelect, Layout: "overlay"},
		}

	case "onmyoji_life_barrier_release_combo":
		if player == nil {
			return nil
		}
		combos := runtimeutil.ParseStringSliceContextValue(data["release_card_combos"])
		options := make([]model.PromptOption, 0, len(combos))
		for _, combo := range combos {
			parts := strings.Split(combo, ",")
			if len(parts) != 2 {
				continue
			}
			i, err1 := strconv.Atoi(parts[0])
			j, err2 := strconv.Atoi(parts[1])
			if err1 != nil || err2 != nil || i < 0 || j < 0 || i >= len(player.Hand) || j >= len(player.Hand) || i == j {
				continue
			}
			options = append(options, model.PromptOption{
				ID:    combo,
				Label: fmt.Sprintf("%d:%s + %d:%s（%s命格）", i+1, player.Hand[i].Name, j+1, player.Hand[j].Name, player.Hand[i].Faction),
			})
		}
		if len(options) == 0 {
			return nil
		}
		return &model.Prompt{
			Type:         model.PromptConfirm,
			PlayerID:     playerID,
			Message:      "【生命结界·分支②】请选择要弃置的2张同命格手牌：",
			Options:      options,
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationBranchSelect, Layout: "overlay"},
		}

	case "onmyoji_dark_ritual_target":
		return engineplayer.BuildTargetChoicePrompt(rt, choiceType, playerID, "【黑暗祭礼】请选择2点法术伤害目标：", data, false)

	case "onmyoji_life_barrier_support_target":
		return engineplayer.BuildTargetChoicePrompt(rt, choiceType, playerID, "【生命结界·分支①】请选择获得+1宝石/+1治疗的队友：", data, false)

	case "onmyoji_life_barrier_release_target":
		return engineplayer.BuildTargetChoicePrompt(rt, choiceType, playerID, "【生命结界·分支②】请选择弃1张手牌的队友：", data, false)

	case "onmyoji_yinyang_confirm":
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【阴阳转换】你可使用同命格攻击牌应战，是否发动？",
			Options: []model.PromptOption{
				{ID: "0", Label: "是"},
				{ID: "1", Label: "否"},
			},
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationBranchSelect, Layout: "overlay"},
		}

	case "onmyoji_yinyang_card":
		rawOptions := parseOnmyojiCardOptions(data["card_options"])
		options := make([]model.PromptOption, 0, len(rawOptions))
		for _, option := range rawOptions {
			options = append(options, model.PromptOption{ID: option.CardID, Label: option.Label})
		}
		if len(options) == 0 {
			return nil
		}
		return &model.Prompt{
			Type:         model.PromptChooseCards,
			PlayerID:     playerID,
			Message:      "【阴阳转换】请选择用于同命格应战的攻击牌：",
			Options:      options,
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationCardPicker, CardSource: "hand"},
		}

	case "onmyoji_yinyang_counter_target":
		options := buildPromptOptionsForPlayerIDs(rt, runtimeutil.ParseStringSliceContextValue(data["counter_target_ids"]))
		if len(options) == 0 {
			return nil
		}
		return &model.Prompt{
			Type:         model.PromptConfirm,
			PlayerID:     playerID,
			Message:      "【阴阳转换】请选择应战反弹目标：",
			Options:      options,
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationTargetPicker, TargetFilter: "custom"},
		}

	case "onmyoji_binding_confirm":
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【式神咒束】是否代替队友执行应战？",
			Options: []model.PromptOption{
				{ID: "0", Label: "是"},
				{ID: "1", Label: "否"},
			},
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationBranchSelect, Layout: "overlay"},
		}

	case "onmyoji_binding_card":
		rawOptions := parseOnmyojiCardOptions(data["card_options"])
		options := make([]model.PromptOption, 0, len(rawOptions))
		for _, option := range rawOptions {
			options = append(options, model.PromptOption{ID: option.CardID, Label: option.Label})
		}
		if len(options) == 0 {
			return nil
		}
		return &model.Prompt{
			Type:         model.PromptChooseCards,
			PlayerID:     playerID,
			Message:      "【式神咒束】请选择用于代应战的攻击牌：",
			Options:      options,
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationCardPicker, CardSource: "hand"},
		}

	case "onmyoji_binding_counter_target":
		options := buildPromptOptionsForPlayerIDs(rt, runtimeutil.ParseStringSliceContextValue(data["counter_target_ids"]))
		if len(options) == 0 {
			return nil
		}
		return &model.Prompt{
			Type:         model.PromptConfirm,
			PlayerID:     playerID,
			Message:      "【式神咒束】请选择应战反弹目标：",
			Options:      options,
			Min:          1,
			Max:          1,
			Presentation: &model.PromptPresentation{Kind: model.PresentationTargetPicker, TargetFilter: "custom"},
		}
	}

	return nil
}

func (choiceHandler) HandleChoice(rt engineplayer.ChoiceRuntime, _ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)

	switch choiceType {
	case "onmyoji_life_barrier_mode":
		userID, _ := ctxData["user_id"].(string)
		user := rt.GetPlayers()[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		lockedTargetID, _ := ctxData["locked_target_id"].(string)
		switch selectionIndex {
		case 0:
			targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["support_target_ids"])
			if len(targetIDs) == 0 {
				return true, fmt.Errorf("生命结界分支①没有可选队友")
			}
			if lockedTargetID != "" {
				return true, resolveOnmyojiLifeBarrierSupportTarget(rt, ctxData, user, lockedTargetID)
			}
			ctxData["choice_type"] = "onmyoji_life_barrier_support_target"
			ctxData["target_ids"] = targetIDs
			if intr := rt.GetPendingInterrupt(); intr != nil {
				intr.Context = ctxData
			}
			rt.NotifyInterruptPrompt()
			return true, nil
		case 1:
			if !hasOnmyojiShikigamiForm(user) {
				return true, fmt.Errorf("不在式神形态，无法选择生命结界分支②")
			}
			combos := runtimeutil.ParseStringSliceContextValue(ctxData["release_card_combos"])
			if len(combos) == 0 {
				return true, fmt.Errorf("分支②需要弃2张同命格手牌")
			}
			targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["release_target_ids"])
			if len(targetIDs) == 0 {
				return true, fmt.Errorf("分支②没有可选队友目标")
			}
			ctxData["choice_type"] = "onmyoji_life_barrier_release_combo"
			ctxData["target_ids"] = targetIDs
			if intr := rt.GetPendingInterrupt(); intr != nil {
				intr.Context = ctxData
			}
			rt.NotifyInterruptPrompt()
			return true, nil
		default:
			return true, fmt.Errorf("无效的生命结界分支选择")
		}

	case "onmyoji_life_barrier_release_combo":
		userID, _ := ctxData["user_id"].(string)
		user := rt.GetPlayers()[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		combos := runtimeutil.ParseStringSliceContextValue(ctxData["release_card_combos"])
		if selectionIndex < 0 || selectionIndex >= len(combos) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		parts := strings.Split(combos[selectionIndex], ",")
		if len(parts) != 2 {
			return true, fmt.Errorf("无效的弃牌组合")
		}
		i, errI := strconv.Atoi(parts[0])
		j, errJ := strconv.Atoi(parts[1])
		if errI != nil || errJ != nil {
			return true, fmt.Errorf("无效的弃牌组合索引")
		}
		if i < 0 || j < 0 || i >= len(user.Hand) || j >= len(user.Hand) || i == j {
			return true, fmt.Errorf("弃牌组合越界")
		}
		c1 := user.Hand[i]
		c2 := user.Hand[j]
		if c1.Faction == "" || c2.Faction == "" || c1.Faction != c2.Faction {
			return true, fmt.Errorf("分支②需要弃2张同命格手牌")
		}
		rt.NotifyCardRevealed(user.ID, []model.Card{c1, c2}, "discard")
		if i < j {
			i, j = j, i
		}
		user.Hand = append(user.Hand[:i], user.Hand[i+1:]...)
		user.Hand = append(user.Hand[:j], user.Hand[j+1:]...)
		rt.AppendToDiscard([]model.Card{c1, c2})
		leaveOnmyojiShikigamiForm(user)
		rt.Log(fmt.Sprintf("%s 的 [生命结界] 分支②：弃2张同命格手牌并脱离式神形态", user.Name))

		lockedTargetID, _ := ctxData["locked_target_id"].(string)
		if lockedTargetID != "" {
			return true, resolveOnmyojiLifeBarrierReleaseTarget(rt, ctxData, user, lockedTargetID)
		}

		ctxData["choice_type"] = "onmyoji_life_barrier_release_target"
		if intr := rt.GetPendingInterrupt(); intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return true, nil

	case "onmyoji_life_barrier_support_target":
		userID, _ := ctxData["user_id"].(string)
		user := rt.GetPlayers()[userID]
		targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		return true, resolveOnmyojiLifeBarrierSupportTarget(rt, ctxData, user, targetIDs[selectionIndex])

	case "onmyoji_life_barrier_release_target":
		userID, _ := ctxData["user_id"].(string)
		user := rt.GetPlayers()[userID]
		targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		return true, resolveOnmyojiLifeBarrierReleaseTarget(rt, ctxData, user, targetIDs[selectionIndex])

	case "onmyoji_dark_ritual_target":
		userID, _ := ctxData["user_id"].(string)
		user := rt.GetPlayers()[userID]
		targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		return true, resolveOnmyojiDarkRitualTarget(rt, ctxData, user, targetIDs[selectionIndex])

	case "onmyoji_binding_confirm":
		if selectionIndex != 0 && selectionIndex != 1 {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		if selectionIndex == 1 {
			rt.PopInterrupt()
			if rt.GetPendingInterrupt() == nil {
				// 简化处理
			}
			return true, nil
		}
		cardOptions := parseOnmyojiCardOptions(ctxData["card_options"])
		if len(cardOptions) == 0 {
			rt.PopInterrupt()
			if rt.GetPendingInterrupt() == nil {
				// 简化处理
			}
			return true, nil
		}
		ctxData["choice_type"] = "onmyoji_binding_card"
		if intr := rt.GetPendingInterrupt(); intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return true, nil

	case "onmyoji_yinyang_confirm":
		if selectionIndex != 0 && selectionIndex != 1 {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		if selectionIndex == 1 {
			// Player declines yinyang: pop interrupt and stay in combat interaction window
			rt.PopInterrupt()
			// Ensure we stay in combat interaction window
			rt.EnsureCombatInteractionWindow()
			return true, nil
		}
		cardOptions := parseOnmyojiCardOptions(ctxData["card_options"])
		if len(cardOptions) == 0 {
			rt.PopInterrupt()
			if rt.GetPendingInterrupt() == nil {
				// 简化处理
			}
			return true, nil
		}
		ctxData["choice_type"] = "onmyoji_yinyang_card"
		if intr := rt.GetPendingInterrupt(); intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return true, nil

	case "onmyoji_yinyang_card":
		cardOptions := parseOnmyojiCardOptions(ctxData["card_options"])
		if selectionIndex < 0 || selectionIndex >= len(cardOptions) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		ctxData["selected_card_id"] = cardOptions[selectionIndex].CardID
		ctxData["choice_type"] = "onmyoji_yinyang_counter_target"
		if intr := rt.GetPendingInterrupt(); intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return true, nil

	case "onmyoji_yinyang_counter_target":
		counterTargetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["counter_target_ids"])
		if selectionIndex < 0 || selectionIndex >= len(counterTargetIDs) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		counterTargetID := counterTargetIDs[selectionIndex]
		actorID, _ := ctxData["actor_id"].(string)
		selectedCardID, _ := ctxData["selected_card_id"].(string)
		actor := rt.GetPlayers()[actorID]
		if actor == nil {
			return true, fmt.Errorf("阴阳师不存在")
		}

		// Consume the card from actor's hand
		card, ok := rt.ConsumePlayableCardByCardID(actorID, selectedCardID)
		if !ok {
			return true, fmt.Errorf("无法消耗选定的卡牌")
		}

		// Apply 阴阳转换 + 式神转换 bonuses (统一入口)
		ApplyFactionCounterBonuses(rt, actor, &card)

		// Add card to discard pile
		rt.AppendToDiscard([]model.Card{card})
		rt.NotifyCardRevealed(actorID, []model.Card{card}, "counter")

		// Get original combat info for cue
		var topCombat *model.CombatRequest
		if stack := rt.GetCombatStack(); len(stack) > 0 {
			topCombat = &stack[len(stack)-1]
		}
		if topCombat != nil {
			rt.NotifyCombatCue(topCombat.AttackerID, topCombat.TargetID, "counter")
		}

		// Resolve counter attack: pop original combat and create reflected one
		rt.PopInterrupt()
		rt.ResolveCounterAttack(actorID, counterTargetID, card)
		return true, nil

	case "onmyoji_binding_card":
		cardOptions := parseOnmyojiCardOptions(ctxData["card_options"])
		if selectionIndex < 0 || selectionIndex >= len(cardOptions) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		ctxData["selected_card_id"] = cardOptions[selectionIndex].CardID
		ctxData["selected_use_faction"] = cardOptions[selectionIndex].UseFaction
		ctxData["choice_type"] = "onmyoji_binding_counter_target"
		if intr := rt.GetPendingInterrupt(); intr != nil {
			intr.Context = ctxData
		}
		rt.NotifyInterruptPrompt()
		return true, nil

	case "onmyoji_binding_counter_target":
		counterTargetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["counter_target_ids"])
		if selectionIndex < 0 || selectionIndex >= len(counterTargetIDs) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		counterTargetID := counterTargetIDs[selectionIndex]
		actorID, _ := ctxData["actor_id"].(string)
		selectedCardID, _ := ctxData["selected_card_id"].(string)
		selectedUseFaction, _ := ctxData["selected_use_faction"].(bool)
		actor := rt.GetPlayers()[actorID]
		if actor == nil {
			return true, fmt.Errorf("阴阳师不存在")
		}

		// Consume the card from actor's hand
		card, ok := rt.ConsumePlayableCardByCardID(actorID, selectedCardID)
		if !ok {
			return true, fmt.Errorf("无法消耗选定的卡牌")
		}

		// Apply 阴阳转换 + 式神转换 bonuses (统一入口)
		// 注意：selectedUseFaction 表示是否使用同命格应战（阴阳转换）
		if selectedUseFaction {
			ApplyFactionCounterBonuses(rt, actor, &card)
		}

		// Add card to discard pile
		rt.AppendToDiscard([]model.Card{card})
		rt.NotifyCardRevealed(actorID, []model.Card{card}, "counter")

		// Get original combat info for cue
		var topCombat *model.CombatRequest
		if stack := rt.GetCombatStack(); len(stack) > 0 {
			topCombat = &stack[len(stack)-1]
		}
		if topCombat != nil {
			rt.NotifyCombatCue(topCombat.AttackerID, topCombat.TargetID, "counter")
		}

		// Resolve counter attack: pop original combat and create reflected one
		rt.PopInterrupt()
		rt.ResolveCounterAttack(actorID, counterTargetID, card)
		return true, nil
	}

	return false, nil
}

// Helper functions for onmyoji

func parseOnmyojiCardOptions(raw interface{}) []onmyojiCardOption {
	options := make([]onmyojiCardOption, 0)
	switch value := raw.(type) {
	case []map[string]interface{}:
		for _, item := range value {
			option, ok := parseOnmyojiCardOption(item)
			if ok {
				options = append(options, option)
			}
		}
	case []interface{}:
		for _, item := range value {
			m, ok := item.(map[string]interface{})
			if !ok || m == nil {
				continue
			}
			option, ok := parseOnmyojiCardOption(m)
			if ok {
				options = append(options, option)
			}
		}
	}
	return options
}

func parseOnmyojiCardOption(data map[string]interface{}) (onmyojiCardOption, bool) {
	if data == nil {
		return onmyojiCardOption{}, false
	}
	cardID, _ := data["card_id"].(string)
	label, _ := data["label"].(string)
	if cardID == "" || label == "" {
		return onmyojiCardOption{}, false
	}
	return onmyojiCardOption{
		CardID:     cardID,
		UseFaction: runtimeutil.ToBoolContextValue(data["use_faction"]),
		Label:      label,
	}, true
}

func buildPromptOptionsForPlayerIDs(rt engineplayer.ChoiceRuntime, targetIDs []string) []model.PromptOption {
	options := make([]model.PromptOption, 0, len(targetIDs))
	for _, targetID := range targetIDs {
		if player := rt.GetPlayers()[targetID]; player != nil {
			options = append(options, model.PromptOption{
				ID:    targetID,
				Label: player.Name,
			})
		}
	}
	return options
}

func stringSliceContains(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}

func hasOnmyojiShikigamiForm(player *model.Player) bool {
	return engineplayer.HasForm(player, model.FormOnmyojiShikigami)
}

func leaveOnmyojiShikigamiForm(player *model.Player) bool {
	return engineplayer.ClearForm(player, model.FormOnmyojiShikigami)
}

func resolveOnmyojiLifeBarrierSupportTarget(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, user *model.Player, targetID string) error {
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["support_target_ids"])
	if !stringSliceContains(targetIDs, targetID) {
		return fmt.Errorf("生命结界分支①目标不合法")
	}
	target := rt.GetPlayers()[targetID]
	if target == nil {
		return fmt.Errorf("目标不存在")
	}
	ghostFire := runtimeutil.ToIntContextValue(ctxData["ghost_fire"])
	target.Gem++
	rt.Heal(targetID, 1)
	if ghostFire > 0 {
		damageType := model.MagicAttack
		rt.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   user.ID,
			Damage:     ghostFire,
			DamageType: damageType,
		})
	}
	rt.Log(fmt.Sprintf("%s 的 [生命结界] 分支①生效：%s +1宝石+1治疗，自身承受%d点法术伤害", user.Name, target.Name, ghostFire))
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		rt.RoutePendingDamageOr(model.TurnStageTurnEnd, func() {
			rt.EnterTurnEndStage()
		})
	}
	return nil
}

func resolveOnmyojiLifeBarrierReleaseTarget(rt engineplayer.ChoiceRuntime, _ map[string]interface{}, user *model.Player, targetID string) error {
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	target := rt.GetPlayers()[targetID]
	if target == nil {
		return fmt.Errorf("目标不存在")
	}
	if target.Camp != user.Camp || target.ID == user.ID {
		return fmt.Errorf("分支②目标必须是其他队友")
	}
	rt.Log(fmt.Sprintf("%s 的 [生命结界] 分支②生效：指定 %s 弃置1张手牌", user.Name, target.Name))
	rt.PopInterrupt()
	// Create discard interrupt for the ally
	if len(target.Hand) > 0 {
		rt.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: target.ID,
			Context: map[string]interface{}{
				"choice_type":   "system_discard_cards",
				"discard_count": 1,
				"prompt":        "【生命结界】请弃置1张手牌：",
			},
		})
	}
	return nil
}

func resolveOnmyojiDarkRitualTarget(rt engineplayer.ChoiceRuntime, ctxData map[string]interface{}, user *model.Player, targetID string) error {
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
	if !stringSliceContains(targetIDs, targetID) {
		return fmt.Errorf("黑暗祭礼目标不合法")
	}
	target := rt.GetPlayers()[targetID]
	if target == nil {
		return fmt.Errorf("目标不存在")
	}
	if target.Camp == user.Camp {
		return fmt.Errorf("黑暗祭礼只能选择敌方目标")
	}
	if user.Tokens == nil {
		user.Tokens = map[string]int{}
	}
	user.Tokens["onmyoji_ghost_fire"] = 0
	rt.AddPendingDamage(model.PendingDamage{
		SourceID:   user.ID,
		TargetID:   targetID,
		Damage:     2,
		DamageType: model.MagicAttack,
	})
	rt.Log(fmt.Sprintf("%s 的 [黑暗祭礼] 生效：对 %s 造成2点法术伤害", user.Name, target.Name))
	rt.PopInterrupt()
	if rt.GetPendingInterrupt() == nil {
		rt.ApplyChoiceResumePoint(model.TurnStageTurnEnd)
		rt.EnterDamageResolution(nil)
	}
	return nil
}
