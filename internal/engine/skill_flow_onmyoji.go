// gameflow: 阴阳师：鬼火、式神、黑暗祭礼、阴阳转换、念咒等长流程与选项。

package engine

import (
	"fmt"
	"starcup-engine/internal/engine/core/runtimeutil"
	"strconv"
	"strings"

	"starcup-engine/internal/model"
)

type onmyojiCardOption struct {
	CardID     string
	UseFaction bool
	Label      string
}

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

func buildPromptOptionsForPlayerIDs(players map[string]*model.Player, targetIDs []string) []model.PromptOption {
	options := make([]model.PromptOption, 0, len(targetIDs))
	for _, targetID := range targetIDs {
		if player := players[targetID]; player != nil {
			options = append(options, model.PromptOption{
				ID:    targetID,
				Label: player.Name,
			})
		}
	}
	return options
}

func (e *GameEngine) buildOnmyojiChoicePrompt(choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
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
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【生命结界】当前鬼火=%d，请选择发动分支：", ghostFire),
			Options:  options,
			Min:      1,
			Max:      1,
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
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【生命结界·分支②】请选择要弃置的2张同命格手牌：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	case "onmyoji_dark_ritual_target":
		options := buildPromptOptionsForPlayerIDs(e.State.Players, runtimeutil.ParseStringSliceContextValue(data["target_ids"]))
		if len(options) == 0 {
			return nil
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【黑暗祭礼】请选择2点法术伤害目标：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	case "onmyoji_life_barrier_support_target":
		options := buildPromptOptionsForPlayerIDs(e.State.Players, runtimeutil.ParseStringSliceContextValue(data["target_ids"]))
		if len(options) == 0 {
			return nil
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【生命结界·分支①】请选择获得+1宝石/+1治疗的队友：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	case "onmyoji_life_barrier_release_target":
		options := buildPromptOptionsForPlayerIDs(e.State.Players, runtimeutil.ParseStringSliceContextValue(data["target_ids"]))
		if len(options) == 0 {
			return nil
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【生命结界·分支②】请选择弃1张手牌的队友：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	case "onmyoji_yinyang_confirm":
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【阴阳转换】你可使用同命格攻击牌应战，是否发动？",
			Options: []model.PromptOption{
				{ID: "0", Label: "是"},
				{ID: "1", Label: "否"},
			},
			Min: 1,
			Max: 1,
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
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【阴阳转换】请选择用于同命格应战的攻击牌：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	case "onmyoji_yinyang_counter_target":
		options := buildPromptOptionsForPlayerIDs(e.State.Players, runtimeutil.ParseStringSliceContextValue(data["counter_target_ids"]))
		if len(options) == 0 {
			return nil
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【阴阳转换】请选择应战反弹目标：",
			Options:  options,
			Min:      1,
			Max:      1,
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
			Min: 1,
			Max: 1,
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
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【式神咒束】请选择用于代应战的攻击牌：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	case "onmyoji_binding_counter_target":
		options := buildPromptOptionsForPlayerIDs(e.State.Players, runtimeutil.ParseStringSliceContextValue(data["counter_target_ids"]))
		if len(options) == 0 {
			return nil
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "【式神咒束】请选择应战反弹目标：",
			Options:  options,
			Min:      1,
			Max:      1,
		}
	default:
		return nil
	}
}

func stringSliceContains(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}

func (e *GameEngine) resolveOnmyojiLifeBarrierSupportTarget(ctxData map[string]interface{}, user *model.Player, targetID string) error {
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["support_target_ids"])
	if !stringSliceContains(targetIDs, targetID) {
		return fmt.Errorf("生命结界分支①目标不合法")
	}
	target := e.State.Players[targetID]
	if target == nil {
		return fmt.Errorf("目标不存在")
	}
	ghostFire := runtimeutil.ToIntContextValue(ctxData["ghost_fire"])
	target.Gem++
	e.Heal(targetID, 1)
	if ghostFire > 0 {
		damageType := model.MagicAttack
		if ghostFire >= 3 {
			damageType = model.DamageType("magic_no_morale")
		}
		e.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   user.ID,
			Damage:     ghostFire,
			DamageType: damageType,
		})
	}
	e.Log(fmt.Sprintf("%s 的 [生命结界] 分支①生效：%s +1宝石+1治疗，自身承受%d点法术伤害", user.Name, target.Name, ghostFire))
	e.PopInterrupt()
	if e.State.PendingInterrupt == nil {
		e.routePendingDamageOr(model.TurnStageTurnEnd, func() {
			e.enterTurnEndStage()
		})
	}
	return nil
}

func (e *GameEngine) resolveOnmyojiLifeBarrierReleaseTarget(ctxData map[string]interface{}, user *model.Player, targetID string) error {
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["release_target_ids"])
	if !stringSliceContains(targetIDs, targetID) {
		return fmt.Errorf("生命结界分支②目标不合法")
	}
	target := e.State.Players[targetID]
	if target == nil {
		return fmt.Errorf("目标不存在")
	}
	if target.Camp != user.Camp || target.ID == user.ID {
		return fmt.Errorf("分支②目标必须是其他队友")
	}
	if len(target.Hand) == 0 {
		return fmt.Errorf("目标队友没有手牌可弃置")
	}
	e.PushInterrupt(newDiscardChoiceInterrupt(target.ID, map[string]interface{}{
		"discard_count": 1,
		"prompt":        fmt.Sprintf("【生命结界】请弃置1张手牌（由 %s 指定）", user.Name),
	}))
	e.Log(fmt.Sprintf("%s 的 [生命结界] 分支②生效：指定 %s 弃置1张手牌", user.Name, target.Name))
	e.PopInterrupt()
	return nil
}

func (e *GameEngine) resolveOnmyojiDarkRitualTarget(ctxData map[string]interface{}, user *model.Player, targetID string) error {
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
	if !stringSliceContains(targetIDs, targetID) {
		return fmt.Errorf("黑暗祭礼目标不合法")
	}
	target := e.State.Players[targetID]
	if target == nil {
		return fmt.Errorf("目标不存在")
	}
	if target.Camp == user.Camp {
		return fmt.Errorf("黑暗祭礼只能选择敌方目标")
	}
	ghostFire := runtimeutil.ToIntContextValue(ctxData["ghost_fire"])
	user.Tokens["onmyoji_ghost_fire"] = 0
	e.AddPendingDamage(model.PendingDamage{
		SourceID:   user.ID,
		TargetID:   targetID,
		Damage:     2,
		DamageType: model.MagicAttack,
	})
	e.Log(fmt.Sprintf("%s 的 [黑暗祭礼] 生效：移除%d点鬼火，对 %s 造成2点法术伤害", user.Name, ghostFire, target.Name))
	e.PopInterrupt()
	if e.State.PendingInterrupt == nil {
		e.setReturnPoint(model.TurnStageTurnEnd)
		e.enterDamageResolution(nil)
	}
	return nil
}

func (e *GameEngine) handleOnmyojiChoiceInput(selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	if onmyojiChoiceFlow(choiceType) == "" {
		return false, nil
	}
	return e.handleOnmyojiChoiceInputByTypeLegacy(selectionIndex, ctxData)
}

func onmyojiChoiceFlow(choiceType string) string {
	switch choiceType {
	case "onmyoji_life_barrier_mode", "onmyoji_life_barrier_release_combo", "onmyoji_life_barrier_support_target", "onmyoji_life_barrier_release_target":
		return "life_barrier"
	case "onmyoji_dark_ritual_target":
		return "dark_ritual"
	case "onmyoji_binding_confirm", "onmyoji_binding_card", "onmyoji_binding_counter_target":
		return "binding"
	case "onmyoji_yinyang_confirm", "onmyoji_yinyang_card", "onmyoji_yinyang_counter_target":
		return "yinyang"
	default:
		return ""
	}
}

func (e *GameEngine) handleOnmyojiChoiceInputByTypeLegacy(selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	switch choiceType {
	case "onmyoji_life_barrier_mode":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
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
				return true, e.resolveOnmyojiLifeBarrierSupportTarget(ctxData, user, lockedTargetID)
			}
			ctxData["choice_type"] = "onmyoji_life_barrier_support_target"
			ctxData["target_ids"] = targetIDs
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
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
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return true, nil
		default:
			return true, fmt.Errorf("无效的生命结界分支选择")
		}
	case "onmyoji_life_barrier_release_combo":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
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
		e.NotifyCardRevealed(user.ID, []model.Card{c1, c2}, "discard")
		if i < j {
			i, j = j, i
		}
		user.Hand = append(user.Hand[:i], user.Hand[i+1:]...)
		user.Hand = append(user.Hand[:j], user.Hand[j+1:]...)
		e.State.DiscardPile = append(e.State.DiscardPile, c1, c2)
		beforePoses := e.snapshotPlayerPoses()
		leaveOnmyojiShikigamiForm(user)
		e.Log(fmt.Sprintf("%s 的 [生命结界] 分支②：弃2张同命格手牌并脱离式神形态", user.Name))
		e.dispatchOrientationChanges(beforePoses)

		lockedTargetID, _ := ctxData["locked_target_id"].(string)
		if lockedTargetID != "" {
			return true, e.resolveOnmyojiLifeBarrierReleaseTarget(ctxData, user, lockedTargetID)
		}

		ctxData["choice_type"] = "onmyoji_life_barrier_release_target"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil
	case "onmyoji_life_barrier_support_target":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		return true, e.resolveOnmyojiLifeBarrierSupportTarget(ctxData, user, targetIDs[selectionIndex])
	case "onmyoji_life_barrier_release_target":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		return true, e.resolveOnmyojiLifeBarrierReleaseTarget(ctxData, user, targetIDs[selectionIndex])
	case "onmyoji_dark_ritual_target":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		return true, e.resolveOnmyojiDarkRitualTarget(ctxData, user, targetIDs[selectionIndex])
	case "onmyoji_binding_confirm":
		if selectionIndex != 0 && selectionIndex != 1 {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		if selectionIndex == 1 {
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				if e.State.CombatStage == model.CombatStageNone {
					e.setCombatStage(model.CombatStageHitCheck)
				}
				e.clearSubflow()
			}
			return true, nil
		}
		cardOptions := parseOnmyojiCardOptions(ctxData["card_options"])
		if len(cardOptions) == 0 {
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				if e.State.CombatStage == model.CombatStageNone {
					e.setCombatStage(model.CombatStageHitCheck)
				}
				e.clearSubflow()
			}
			return true, nil
		}
		ctxData["choice_type"] = "onmyoji_binding_card"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil
	case "onmyoji_yinyang_confirm":
		if selectionIndex != 0 && selectionIndex != 1 {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		if selectionIndex == 1 {
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				if e.State.CombatStage == model.CombatStageNone {
					e.setCombatStage(model.CombatStageHitCheck)
				}
				e.clearSubflow()
			}
			return true, nil
		}
		cardOptions := parseOnmyojiCardOptions(ctxData["card_options"])
		if len(cardOptions) == 0 {
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				if e.State.CombatStage == model.CombatStageNone {
					e.setCombatStage(model.CombatStageHitCheck)
				}
				e.clearSubflow()
			}
			return true, nil
		}
		ctxData["choice_type"] = "onmyoji_yinyang_card"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil
	case "onmyoji_yinyang_card":
		cardOptions := parseOnmyojiCardOptions(ctxData["card_options"])
		if selectionIndex < 0 || selectionIndex >= len(cardOptions) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		ctxData["selected_card_id"] = cardOptions[selectionIndex].CardID
		ctxData["choice_type"] = "onmyoji_yinyang_counter_target"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil
	case "onmyoji_yinyang_counter_target":
		counterTargets := runtimeutil.ParseStringSliceContextValue(ctxData["counter_target_ids"])
		if selectionIndex < 0 || selectionIndex >= len(counterTargets) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		actorID, _ := ctxData["actor_id"].(string)
		cardID, _ := ctxData["selected_card_id"].(string)
		if actorID == "" || cardID == "" {
			return true, fmt.Errorf("阴阳转换上下文缺失")
		}
		actor := e.State.Players[actorID]
		if actor == nil {
			return true, fmt.Errorf("阴阳师不存在")
		}
		cardIdx := findPlayableCardIndexByID(actor, cardID)
		if cardIdx < 0 {
			return true, fmt.Errorf("应战牌已不存在")
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if e.State.CombatStage == model.CombatStageNone {
				e.setCombatStage(model.CombatStageHitCheck)
			}
			e.clearSubflow()
		}
		return true, e.handleCombatResponse(model.PlayerAction{
			PlayerID:  actorID,
			Type:      model.CmdRespond,
			ExtraArgs: []string{"counter"},
			CardIndex: cardIdx,
			TargetID:  counterTargets[selectionIndex],
		})
	case "onmyoji_binding_card":
		cardOptions := parseOnmyojiCardOptions(ctxData["card_options"])
		if selectionIndex < 0 || selectionIndex >= len(cardOptions) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		ctxData["selected_card_id"] = cardOptions[selectionIndex].CardID
		ctxData["selected_use_faction"] = cardOptions[selectionIndex].UseFaction
		ctxData["choice_type"] = "onmyoji_binding_counter_target"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil
	case "onmyoji_binding_counter_target":
		counterTargets := runtimeutil.ParseStringSliceContextValue(ctxData["counter_target_ids"])
		if selectionIndex < 0 || selectionIndex >= len(counterTargets) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		actorID, _ := ctxData["actor_id"].(string)
		cardID, _ := ctxData["selected_card_id"].(string)
		if actorID == "" || cardID == "" {
			return true, fmt.Errorf("式神咒束上下文缺失")
		}
		actor := e.State.Players[actorID]
		if actor == nil {
			return true, fmt.Errorf("阴阳师不存在")
		}
		if len(e.State.CombatStack) == 0 {
			return true, fmt.Errorf("当前没有可代应战的战斗请求")
		}
		combatReq := &e.State.CombatStack[len(e.State.CombatStack)-1]
		if !e.payOnmyojiBindingCost(actor.Camp) {
			return true, fmt.Errorf("战绩区资源不足，无法发动式神咒束")
		}
		combatReq.OnmyojiBindingChecked = true
		combatReq.OnmyojiBindingActorID = actorID
		combatReq.OnmyojiBindingCounterID = cardID
		combatReq.OnmyojiBindingTargetID = counterTargets[selectionIndex]
		combatReq.OnmyojiBindingUseFaction = runtimeutil.ToBoolContextValue(ctxData["selected_use_faction"])
		combatReq.TargetID = actorID

		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if e.State.CombatStage == model.CombatStageNone {
				e.setCombatStage(model.CombatStageHitCheck)
			}
			e.clearSubflow()
		}
		return true, nil
	default:
		return false, nil
	}
}

func (e *GameEngine) maybeOnmyojiDarkRitual(player *model.Player) bool {
	if player == nil || !e.isOnmyoji(player) || player.Tokens == nil || player.Tokens["onmyoji_ghost_fire"] < 3 {
		return false
	}
	targetIDs := e.campEnemyIDs(player.Camp)
	if len(targetIDs) == 0 {
		return false
	}
	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: player.ID,
		Context: map[string]interface{}{
			"choice_type": "onmyoji_dark_ritual_target",
			"user_id":     player.ID,
			"target_ids":  targetIDs,
			"ghost_fire":  player.Tokens["onmyoji_ghost_fire"],
		},
	})
	e.Log(fmt.Sprintf("%s 的 [黑暗祭礼] 触发，等待选择2点法术伤害目标", player.Name))
	return true
}

func (e *GameEngine) applyOnmyojiFactionCounterBonuses(actor *model.Player, card *model.Card) {
	if actor == nil || card == nil {
		return
	}
	beforePoses := e.snapshotPlayerPoses()
	if actor.Tokens == nil {
		actor.Tokens = map[string]int{}
	}
	actor.Tokens["onmyoji_ghost_fire"]++
	if actor.Tokens["onmyoji_ghost_fire"] > 3 {
		actor.Tokens["onmyoji_ghost_fire"] = 3
	}
	e.Log(fmt.Sprintf("%s 的 [阴阳转换] 触发，鬼火+1", actor.Name))
	if hasOnmyojiShikigamiForm(actor) {
		e.DrawCards(actor.ID, 1)
		actor.Tokens["onmyoji_ghost_fire"]++
		if actor.Tokens["onmyoji_ghost_fire"] > 3 {
			actor.Tokens["onmyoji_ghost_fire"] = 3
		}
		leaveOnmyojiShikigamiForm(actor)
		e.Log(fmt.Sprintf("%s 的 [式神转换] 触发：摸1并鬼火+1，然后脱离式神形态", actor.Name))
	}
	card.Damage = actor.Tokens["onmyoji_ghost_fire"]
	if card.Damage < 0 {
		card.Damage = 0
	}
	e.dispatchOrientationChanges(beforePoses)
}

// tryStartOnmyojiBindingInterrupt 检查并触发“式神咒束”代应战确认。
// 返回 true 表示已进入中断等待（应暂停当前 Drive）。
func (e *GameEngine) tryStartOnmyojiBindingInterrupt(combatReq *model.CombatRequest) bool {
	if combatReq == nil {
		return false
	}
	if combatReq.OnmyojiBindingChecked {
		return false
	}
	combatReq.OnmyojiBindingChecked = true

	if combatReq.IsCounter || combatReq.IsForcedHit || !combatReq.CanBeResponded || combatReq.Card == nil {
		return false
	}
	if combatReq.Card.Element == model.ElementDark {
		return false
	}
	target := e.State.Players[combatReq.TargetID]
	attacker := e.State.Players[combatReq.AttackerID]
	if target == nil || attacker == nil {
		return false
	}
	if attacker.Camp == target.Camp {
		return false
	}
	if e.isOnmyoji(target) {
		return false
	}

	var counterTargetIDs []string
	for _, pid := range e.State.PlayerOrder {
		if pid == attacker.ID {
			continue
		}
		player := e.State.Players[pid]
		if player == nil || player.Camp != attacker.Camp {
			continue
		}
		counterTargetIDs = append(counterTargetIDs, pid)
	}
	if len(counterTargetIDs) == 0 {
		return false
	}

	for _, pid := range e.State.PlayerOrder {
		actor := e.State.Players[pid]
		if actor == nil || actor.ID == target.ID {
			continue
		}
		if !e.isOnmyoji(actor) || actor.Camp != target.Camp {
			continue
		}
		if !hasOnmyojiShikigamiForm(actor) {
			continue
		}
		if !e.canPayOnmyojiBindingCost(actor.Camp) {
			continue
		}
		cardOptions := collectOnmyojiCounterOptions(actor, combatReq.Card)
		if len(cardOptions) == 0 {
			continue
		}
		e.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: actor.ID,
			Context: map[string]interface{}{
				"choice_type":        "onmyoji_binding_confirm",
				"actor_id":           actor.ID,
				"attacker_id":        combatReq.AttackerID,
				"target_id":          combatReq.TargetID,
				"card_options":       cardOptions,
				"counter_target_ids": counterTargetIDs,
			},
		})
		e.Log(fmt.Sprintf("%s 可发动 [式神咒束] 代应战，等待其确认", actor.Name))
		return true
	}
	return false
}

// tryStartOnmyojiYinYangInterrupt 检查并触发“阴阳转换”优先确认。
// 规则：目标阴阳师若手里有“与来袭攻击同命格”的攻击牌，则先询问是否发动；
// 若选择不发动，才进入常规 承受/防御/应战 弹框。
func (e *GameEngine) tryStartOnmyojiYinYangInterrupt(combatReq *model.CombatRequest) bool {
	if combatReq == nil {
		return false
	}
	if combatReq.OnmyojiYinYangChecked {
		return false
	}
	combatReq.OnmyojiYinYangChecked = true

	if combatReq.IsCounter || combatReq.IsForcedHit || !combatReq.CanBeResponded || combatReq.Card == nil {
		return false
	}
	if combatReq.Card.Element == model.ElementDark {
		return false
	}

	target := e.State.Players[combatReq.TargetID]
	attacker := e.State.Players[combatReq.AttackerID]
	if target == nil || attacker == nil || !e.isOnmyoji(target) {
		return false
	}
	if !onmyojiCanUseFactionCounter(combatReq.Card) {
		return false
	}

	allOptions := collectOnmyojiCounterOptions(target, combatReq.Card)
	factionOptions := make([]map[string]interface{}, 0, len(allOptions))
	for _, option := range allOptions {
		useFaction, _ := option["use_faction"].(bool)
		if useFaction {
			factionOptions = append(factionOptions, option)
		}
	}
	if len(factionOptions) == 0 {
		return false
	}

	var counterTargetIDs []string
	for _, pid := range e.State.PlayerOrder {
		if pid == attacker.ID {
			continue
		}
		player := e.State.Players[pid]
		if player == nil || player.Camp != attacker.Camp {
			continue
		}
		counterTargetIDs = append(counterTargetIDs, pid)
	}
	if len(counterTargetIDs) == 0 {
		return false
	}

	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: target.ID,
		Context: map[string]interface{}{
			"choice_type":        "onmyoji_yinyang_confirm",
			"actor_id":           target.ID,
			"attacker_id":        combatReq.AttackerID,
			"target_id":          combatReq.TargetID,
			"card_options":       factionOptions,
			"counter_target_ids": counterTargetIDs,
		},
	})
	e.Log(fmt.Sprintf("%s 可发动 [阴阳转换]，等待其确认", target.Name))
	return true
}

// executeOnmyojiBindingCounter 在战斗阶段自动执行已确认的“式神咒束应战”。
// 返回 true 表示已推进流程（可能进入中断），当前 Drive 应暂停。
func (e *GameEngine) executeOnmyojiBindingCounter(combatReq *model.CombatRequest) bool {
	if combatReq == nil {
		return false
	}
	actorID := combatReq.OnmyojiBindingActorID
	cardID := combatReq.OnmyojiBindingCounterID
	counterTargetID := combatReq.OnmyojiBindingTargetID
	if actorID == "" || cardID == "" || counterTargetID == "" {
		return false
	}
	actor := e.State.Players[actorID]
	if actor == nil || combatReq.Card == nil {
		combatReq.OnmyojiBindingActorID = ""
		combatReq.OnmyojiBindingCounterID = ""
		combatReq.OnmyojiBindingTargetID = ""
		combatReq.OnmyojiBindingUseFaction = false
		return false
	}
	cardIdx := findPlayableCardIndexByID(actor, cardID)
	card, _, _, ok := getPlayableCardByIndex(actor, cardIdx)
	if !ok || card.Type != model.CardTypeAttack {
		combatReq.OnmyojiBindingActorID = ""
		combatReq.OnmyojiBindingCounterID = ""
		combatReq.OnmyojiBindingTargetID = ""
		combatReq.OnmyojiBindingUseFaction = false
		return false
	}
	useFaction := combatReq.OnmyojiBindingUseFaction
	canCounter := card.Element == combatReq.Card.Element || card.Element == model.ElementDark
	if !canCounter && useFaction {
		canCounter = onmyojiCanUseFactionCounter(combatReq.Card) &&
			card.Faction != "" && card.Faction == combatReq.Card.Faction
	}
	if !canCounter {
		combatReq.OnmyojiBindingActorID = ""
		combatReq.OnmyojiBindingCounterID = ""
		combatReq.OnmyojiBindingTargetID = ""
		combatReq.OnmyojiBindingUseFaction = false
		return false
	}

	e.NotifyCardRevealed(actor.ID, []model.Card{card}, "counter")
	e.NotifyCombatCue(combatReq.AttackerID, combatReq.TargetID, "counter")
	if _, err := consumePlayableCardByIndex(actor, cardIdx); err != nil {
		combatReq.OnmyojiBindingActorID = ""
		combatReq.OnmyojiBindingCounterID = ""
		combatReq.OnmyojiBindingTargetID = ""
		combatReq.OnmyojiBindingUseFaction = false
		return false
	}
	e.State.DiscardPile = append(e.State.DiscardPile, card)

	if useFaction {
		e.applyOnmyojiFactionCounterBonuses(actor, &card)
	}

	missCtx := &model.EventContext{
		Type:     model.EventAttack,
		SourceID: combatReq.AttackerID,
		TargetID: combatReq.TargetID,
		Card:     combatReq.Card,
		AttackInfo: &model.AttackEventInfo{
			ActionType: string(model.ActionAttack),
			CounterInitiator: func() string {
				if combatReq.IsCounter {
					return combatReq.AttackerID
				}
				return ""
			}(),
		},
	}
	skillCtx := e.buildContext(e.State.Players[combatReq.AttackerID], e.State.Players[combatReq.TargetID], model.TimingOnHitCheck, missCtx)
	skillCtx.Selections["attack_miss_resume"] = map[string]interface{}{
		"mode":              "counter",
		"attacker_id":       combatReq.AttackerID,
		"target_id":         combatReq.TargetID,
		"counter_player_id": actor.ID,
		"counter_target_id": counterTargetID,
		"counter_card":      card,
	}
	e.dispatcher.OnTiming(skillCtx.Timing, skillCtx)
	if e.State.PendingInterrupt != nil {
		return true
	}

	e.resolveMagicBowPierceMiss(combatReq.AttackerID, combatReq.TargetID, combatReq.Card, combatReq.IsCounter)
	e.Log(fmt.Sprintf("[Combat] %s 通过[式神咒束]代应战成功，攻击反弹给 %s", actor.Name, model.GetPlayerDisplayName(e.State.Players[counterTargetID])))
	e.State.CombatStack = e.State.CombatStack[:len(e.State.CombatStack)-1]
	e.initCombat(actor.ID, counterTargetID, &card, false, true, false, nil, true)
	combatReq.OnmyojiBindingActorID = ""
	combatReq.OnmyojiBindingCounterID = ""
	combatReq.OnmyojiBindingTargetID = ""
	combatReq.OnmyojiBindingUseFaction = false
	return true
}
