// gameflow: 神圣弓手：射击与圣箭相关流程。

package engine

import (
	"fmt"
	"starcup-engine/internal/engine/core/runtimeutil"
	"strconv"
	"strings"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func (e *GameEngine) buildHolyBowChoicePrompt(choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	switch choiceType {
	case "hb_holy_shard_combo":
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
				Label: fmt.Sprintf("%s系：%d:%s + %d:%s", elementNameForPrompt(parts[0]), i+1, player.Hand[i].Name, j+1, player.Hand[j].Name),
			})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【圣屑飓暴】请选择要弃置的2张同系攻击牌：", Options: options, Min: 1, Max: 1}

	case "hb_holy_shard_target":
		targetIDs := runtimeutil.ParseStringSliceContextValue(data["target_ids"])
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【圣屑飓暴】请选择主动攻击目标：", Options: holyBowChoicePlayerOptions(e, targetIDs), Min: 1, Max: 1}

	case "hb_holy_shard_miss_confirm":
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【圣屑飓暴】未命中：是否移除治疗并令1名队友弃牌？", Options: []model.PromptOption{{ID: "0", Label: "是"}, {ID: "1", Label: "否"}}, Min: 1, Max: 1}

	case "hb_holy_shard_miss_x":
		validX := parseIntSliceContextValue(data["valid_x"])
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
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【圣屑飓暴】请选择移除治疗点数X：", Options: options, Min: 1, Max: 1}

	case "hb_holy_shard_miss_ally_target":
		allyIDs := runtimeutil.ParseStringSliceContextValue(data["ally_ids"])
		xValue := runtimeutil.ToIntContextValue(data["x_value"])
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: fmt.Sprintf("【圣屑飓暴】请选择1名队友弃置%d张手牌：", xValue), Options: holyBowChoicePlayerOptions(e, allyIDs), Min: 1, Max: 1}

	case "hb_radiant_descent_cost":
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
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【圣煌降临】请选择支付方式：", Options: options, Min: 1, Max: 1}

	case "hb_light_burst_mode":
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
					if enemy := e.State.Players[enemyID]; enemy != nil && len(enemy.Hand) <= limit {
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
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【圣光爆裂】请选择发动分支：", Options: options, Min: 1, Max: 1}

	case "hb_light_burst_mode_a_target", "hb_meteor_bullet_target":
		allyIDs := runtimeutil.ParseStringSliceContextValue(data["ally_ids"])
		msg := "请选择我方角色："
		if choiceType == "hb_light_burst_mode_a_target" {
			msg = "【圣光爆裂】分支①请选择获得治疗的我方角色："
		} else {
			msg = "【流星圣弹】请选择获得治疗的我方角色："
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: msg, Options: holyBowChoicePlayerOptions(e, allyIDs), Min: 1, Max: 1}

	case "hb_light_burst_mode_b_x":
		enemyIDs := runtimeutil.ParseStringSliceContextValue(data["enemy_ids"])
		maxX := runtimeutil.ToIntContextValue(data["max_x"])
		options := make([]model.PromptOption, 0, maxX)
		if player != nil {
			for x := 1; x <= maxX; x++ {
				limit := len(player.Hand) - x
				eligible := 0
				for _, enemyID := range enemyIDs {
					if enemy := e.State.Players[enemyID]; enemy != nil && len(enemy.Hand) <= limit {
						eligible++
					}
				}
				if eligible <= 0 {
					continue
				}
				options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", x), Label: fmt.Sprintf("X=%d（移除%d治疗并弃%d张牌）", x, x, x)})
			}
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【圣光爆裂】分支②请选择X值：", Options: options, Min: 1, Max: 1}

	case "hb_light_burst_mode_b_targets":
		candidates := runtimeutil.ParseStringSliceContextValue(data["candidate_target_ids"])
		selectedSet := runtimeutil.IDsToSet(runtimeutil.ParseStringSliceContextValue(data["selected_target_ids"]))
		options := make([]model.PromptOption, 0, len(candidates)+1)
		for _, targetID := range candidates {
			if selectedSet[targetID] {
				continue
			}
			if target := e.State.Players[targetID]; target != nil {
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
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: fmt.Sprintf("【圣光爆裂】分支②请点击角色立绘选择目标（已选%d/最多%d）：", len(selectedSet), maxTargets), Options: options, Min: 1, Max: 1}

	case "hb_light_burst_mode_b_discard":
		remaining := parseIntSliceContextValue(data["remaining_indices"])
		selectedCount := len(parseIntSliceContextValue(data["selected_indices"]))
		xValue := runtimeutil.ToIntContextValue(data["x_value"])
		options := make([]model.PromptOption, 0, len(remaining))
		for _, idx := range remaining {
			if player == nil || idx < 0 || idx >= len(player.Hand) {
				continue
			}
			options = append(options, model.PromptOption{ID: fmt.Sprintf("%d", idx), Label: fmt.Sprintf("%d: %s", idx+1, formatCardInfo(player.Hand[idx]))})
		}
		remainingPick := xValue - selectedCount
		if remainingPick < 1 {
			remainingPick = 1
		}
		if len(options) > 0 && remainingPick > len(options) {
			remainingPick = len(options)
		}
		return &model.Prompt{Type: model.PromptChooseCards, PlayerID: playerID, Message: fmt.Sprintf("【圣光爆裂】分支②请选择要弃置的%d张手牌：", remainingPick), Options: options, Min: remainingPick, Max: remainingPick}

	case "hb_meteor_bullet_cost":
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
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【流星圣弹】请选择要移除的资源：", Options: options, Min: 1, Max: 1}

	case "hb_radiant_cannon_side":
		requiredFaith := runtimeutil.ToIntContextValue(data["required_faith"])
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: fmt.Sprintf("【圣煌辉光炮】将消耗1辉光炮与%d点信仰。请选择士气对齐方向：", requiredFaith), Options: []model.PromptOption{{ID: "0", Label: "将红方士气调整为蓝方士气"}, {ID: "1", Label: "将蓝方士气调整为红方士气"}}, Min: 1, Max: 1}

	case "hb_auto_fill_resource":
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
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【自动填充】请选择要发动的分支：", Options: options, Min: 1, Max: 1}

	case "hb_auto_fill_gain":
		branch, _ := data["branch"].(string)
		options := make([]model.PromptOption, 0, 2)
		if branch == "gem" {
			options = append(options, model.PromptOption{ID: "0", Label: "+2信仰"})
			options = append(options, model.PromptOption{ID: "1", Label: "+2治疗"})
		} else {
			options = append(options, model.PromptOption{ID: "0", Label: "+1信仰"})
			options = append(options, model.PromptOption{ID: "1", Label: "+1治疗"})
		}
		return &model.Prompt{Type: model.PromptConfirm, PlayerID: playerID, Message: "【自动填充】请选择增益：", Options: options, Min: 1, Max: 1}
	}

	return nil
}

func (e *GameEngine) handleHolyBowChoiceInput(_ string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	if holyBowChoiceFlow(choiceType) == "" {
		return false, nil
	}
	return e.handleHolyBowChoiceInputByTypeLegacy(selectionIndex, ctxData)
}

func holyBowChoiceFlow(choiceType string) string {
	switch choiceType {
	case "hb_holy_shard_combo", "hb_holy_shard_target", "hb_holy_shard_miss_confirm", "hb_holy_shard_miss_x", "hb_holy_shard_miss_ally_target":
		return "holy_shard"
	case "hb_radiant_descent_cost":
		return "radiant_descent"
	case "hb_light_burst_mode", "hb_light_burst_mode_a_target", "hb_light_burst_mode_b_x", "hb_light_burst_mode_b_targets", "hb_light_burst_mode_b_discard":
		return "light_burst"
	case "hb_meteor_bullet_cost", "hb_meteor_bullet_target":
		return "meteor_bullet"
	case "hb_radiant_cannon_side":
		return "radiant_cannon"
	case "hb_auto_fill_resource", "hb_auto_fill_gain":
		return "auto_fill"
	default:
		return ""
	}
}

func (e *GameEngine) handleHolyBowChoiceInputByTypeLegacy(selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)

	switch choiceType {
	case "hb_holy_shard_combo":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		combos := runtimeutil.ParseStringSliceContextValue(ctxData["combos"])
		if selectionIndex < 0 || selectionIndex >= len(combos) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		choice := combos[selectionIndex]
		parts := strings.Split(choice, ":")
		if len(parts) != 2 {
			return true, fmt.Errorf("同系组合格式错误")
		}
		element := strings.TrimSpace(parts[0])
		idxParts := strings.Split(parts[1], ",")
		if len(idxParts) != 2 {
			return true, fmt.Errorf("同系组合索引格式错误")
		}
		i, err1 := strconv.Atoi(strings.TrimSpace(idxParts[0]))
		j, err2 := strconv.Atoi(strings.TrimSpace(idxParts[1]))
		if err1 != nil || err2 != nil || i < 0 || j < 0 || i >= len(user.Hand) || j >= len(user.Hand) || i == j {
			return true, fmt.Errorf("无效的弃牌索引")
		}
		c1 := user.Hand[i]
		c2 := user.Hand[j]
		if c1.Type != model.CardTypeAttack || c2.Type != model.CardTypeAttack || c1.Element != c2.Element {
			return true, fmt.Errorf("圣屑飓暴需要弃置2张同系攻击牌")
		}
		removed, err := removeCardsByIndicesFromHand(user, []int{i, j})
		if err != nil {
			return true, err
		}
		e.NotifyCardRevealed(user.ID, removed, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, removed...)
		ctxData["selected_element"] = element
		ctxData["choice_type"] = "hb_holy_shard_target"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil

	case "hb_holy_shard_target":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["target_ids"])
		if selectionIndex < 0 || selectionIndex >= len(targetIDs) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		targetID := targetIDs[selectionIndex]
		target := e.State.Players[targetID]
		if target == nil {
			return true, fmt.Errorf("目标不存在")
		}
		if target.Camp == user.Camp {
			return true, fmt.Errorf("圣屑飓暴只能指定敌方目标")
		}
		eleStr, _ := ctxData["selected_element"].(string)
		ele := model.Element(eleStr)
		if ele == "" {
			return true, fmt.Errorf("圣屑飓暴攻击元素缺失")
		}
		virtualCard := model.Card{
			ID:          fmt.Sprintf("hb_holy_shard_%s_%d", user.ID, len(e.State.DiscardPile)+len(e.State.ActionQueue)+1),
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
		e.State.ActionQueue = append(e.State.ActionQueue, model.QueuedAction{SourceID: user.ID, TargetID: target.ID, Type: model.ActionAttack, Element: ele, Card: &virtualCard, CardIndex: -1, SourceSkill: "hb_holy_shard_storm", UsesVirtualCard: true})
		e.Log(fmt.Sprintf("%s 发动 [圣屑飓暴]：对 %s 发起1次%s系圣命格主动攻击", user.Name, target.Name, ele))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.enterActionExecutionStage()
		}
		return true, nil

	case "hb_holy_shard_miss_confirm":
		if selectionIndex == 1 {
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				if len(e.State.PendingDamageQueue) > 0 {
					e.enterDamageResolution(nil)
				} else if len(e.State.CombatStack) > 0 {
					if e.State.CombatStage == model.CombatStageNone {
						e.setCombatStage(model.CombatStageHitCheck)
					}
					e.clearSubflow()
				} else {
					e.enterExtraActionStage()
				}
			}
			return true, nil
		}
		if selectionIndex != 0 {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
		if maxX <= 0 {
			e.PopInterrupt()
			return true, nil
		}
		ctxData["choice_type"] = "hb_holy_shard_miss_x"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil

	case "hb_holy_shard_miss_x":
		validX := parseIntSliceContextValue(ctxData["valid_x"])
		if len(validX) == 0 {
			maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
			for x := 1; x <= maxX; x++ {
				validX = append(validX, x)
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(validX) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		xValue := validX[selectionIndex]
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		allyIDs := e.holyBowShardMissEligibleAllies(user, xValue)
		if len(allyIDs) == 0 {
			return true, fmt.Errorf("没有可弃置%d张牌的队友", xValue)
		}
		ctxData["x_value"] = xValue
		ctxData["ally_ids"] = allyIDs
		ctxData["choice_type"] = "hb_holy_shard_miss_ally_target"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil

	case "hb_holy_shard_miss_ally_target":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		allyIDs := runtimeutil.ParseStringSliceContextValue(ctxData["ally_ids"])
		if selectionIndex < 0 || selectionIndex >= len(allyIDs) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		targetID := allyIDs[selectionIndex]
		target := e.State.Players[targetID]
		if target == nil {
			return true, fmt.Errorf("目标队友不存在")
		}
		xValue := runtimeutil.ToIntContextValue(ctxData["x_value"])
		if xValue <= 0 {
			return true, fmt.Errorf("无效的X值")
		}
		if user.Heal < xValue {
			return true, fmt.Errorf("治疗不足，无法移除%d点治疗", xValue)
		}
		if len(target.Hand) < xValue {
			return true, fmt.Errorf("目标队友手牌不足%d张，无法作为圣屑飓暴的弃牌目标", xValue)
		}
		user.Heal -= xValue
		discardNeed := xValue
		e.Log(fmt.Sprintf("%s 的 [圣屑飓暴] 未命中分支生效：移除%d点治疗，指定 %s 弃置%d张手牌", user.Name, xValue, target.Name, discardNeed))
		e.PopInterrupt()
		e.PushInterrupt(newDiscardChoiceInterrupt(target.ID, map[string]interface{}{"discard_count": discardNeed, "prompt": fmt.Sprintf("【圣屑飓暴】请弃置%d张手牌：", discardNeed), "stay_in_turn": true, "is_damage_resolution": true}))
		return true, nil

	case "hb_radiant_descent_cost":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		modes := runtimeutil.ParseStringSliceContextValue(ctxData["cost_modes"])
		if selectionIndex < 0 || selectionIndex >= len(modes) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		switch modes[selectionIndex] {
		case "heal":
			if user.Heal < 2 {
				return true, fmt.Errorf("治疗不足2点")
			}
			user.Heal -= 2
		case "faith":
			if holyBowFaith(user) < 2 {
				return true, fmt.Errorf("信仰不足2点")
			}
			addHolyBowFaith(user, -2)
		default:
			return true, fmt.Errorf("无效的支付方式")
		}
		if user.Tokens == nil {
			user.Tokens = map[string]int{}
		}
		beforePoses := e.snapshotPlayerPoses()
		enterHolyBowHolyGloryForm(user)
		model.AppendMagicAction(user, "圣煌降临")
		e.Log(fmt.Sprintf("%s 发动 [圣煌降临]：进入圣煌形态并获得额外法术行动", user.Name))
		e.dispatchOrientationChanges(beforePoses)
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			// 规则：圣煌降临支付完成后，应继续其额外行动流程；恢复点必须和技能规则保持一致。
			e.applyChoiceResumePoint(mustChoiceResumePointFromMap(ctxData, "resume_phase"))
		}
		return true, nil

	case "hb_light_burst_mode":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
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
					if enemy := e.State.Players[enemyID]; enemy != nil && len(enemy.Hand) <= limit {
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
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		switch modeOrder[selectionIndex] {
		case "a":
			ctxData["choice_type"] = "hb_light_burst_mode_a_target"
		case "b":
			ctxData["choice_type"] = "hb_light_burst_mode_b_x"
		default:
			return true, fmt.Errorf("无效的分支")
		}
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil

	case "hb_light_burst_mode_a_target":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		allyIDs := runtimeutil.ParseStringSliceContextValue(ctxData["ally_ids"])
		if selectionIndex < 0 || selectionIndex >= len(allyIDs) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		targetID := allyIDs[selectionIndex]
		if user.Heal < 1 {
			return true, fmt.Errorf("治疗不足，无法发动分支①")
		}
		e.DrawCards(user.ID, 1)
		user.Heal--
		faith := addHolyBowFaith(user, 1)
		e.Heal(targetID, 1)
		targetName := targetID
		if target := e.State.Players[targetID]; target != nil {
			targetName = target.Name
		}
		e.Log(fmt.Sprintf("%s 的 [圣光爆裂] 分支①生效：摸1、移除1治疗、信仰+1（当前%d），%s +1治疗", user.Name, faith, targetName))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if !e.routePendingDamageWithReturn(model.TurnStageExtraAction) {
				e.enterExtraActionStage()
			}
		}
		return true, nil

	case "hb_light_burst_mode_b_x":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		enemyIDs := runtimeutil.ParseStringSliceContextValue(ctxData["enemy_ids"])
		maxX := runtimeutil.ToIntContextValue(ctxData["max_x"])
		validX := make([]int, 0, maxX)
		for x := 1; x <= maxX; x++ {
			limit := len(user.Hand) - x
			eligible := 0
			for _, enemyID := range enemyIDs {
				if enemy := e.State.Players[enemyID]; enemy != nil && len(enemy.Hand) <= limit {
					eligible++
				}
			}
			if eligible > 0 {
				validX = append(validX, x)
			}
		}
		if selectionIndex < 0 || selectionIndex >= len(validX) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		xValue := validX[selectionIndex]
		limit := len(user.Hand) - xValue
		candidateTargets := make([]string, 0, len(enemyIDs))
		for _, enemyID := range enemyIDs {
			if enemy := e.State.Players[enemyID]; enemy != nil && len(enemy.Hand) <= limit {
				candidateTargets = append(candidateTargets, enemyID)
			}
		}
		if len(candidateTargets) == 0 {
			return true, fmt.Errorf("没有满足手牌条件的目标")
		}
		ctxData["x_value"] = xValue
		ctxData["candidate_target_ids"] = candidateTargets
		ctxData["selected_target_ids"] = []string{}
		ctxData["choice_type"] = "hb_light_burst_mode_b_targets"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil

	case "hb_light_burst_mode_b_targets":
		candidates := runtimeutil.ParseStringSliceContextValue(ctxData["candidate_target_ids"])
		xValue := runtimeutil.ToIntContextValue(ctxData["x_value"])
		if xValue <= 0 {
			return true, fmt.Errorf("X值无效")
		}
		maxTargets := xValue
		if len(candidates) < maxTargets {
			maxTargets = len(candidates)
		}
		if maxTargets <= 0 {
			return true, fmt.Errorf("没有可选目标")
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
				return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
			}
			selected = append(selected, remaining[selectionIndex])
			ctxData["selected_target_ids"] = selected
			if len(selected) < maxTargets {
				e.State.PendingInterrupt.Context = ctxData
				e.notifyInterruptPrompt()
				return true, nil
			}
			proceedDiscard = true
		}
		if !proceedDiscard || len(selected) == 0 {
			return true, fmt.Errorf("至少需要选择1名目标")
		}
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		ctxData["selected_indices"] = []int{}
		ctxData["remaining_indices"] = allHandIndices(user)
		ctxData["choice_type"] = "hb_light_burst_mode_b_discard"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil

	case "hb_light_burst_mode_b_discard":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		xValue := runtimeutil.ToIntContextValue(ctxData["x_value"])
		if xValue <= 0 {
			return true, fmt.Errorf("X值无效")
		}
		remaining := parseIntSliceContextValue(ctxData["remaining_indices"])
		selected := parseIntSliceContextValue(ctxData["selected_indices"])
		cardIdx, ok := runtimeutil.ResolveSelectionToCandidate(selectionIndex, remaining)
		if !ok || cardIdx < 0 || cardIdx >= len(user.Hand) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
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
			e.State.PendingInterrupt.Context = ctxData
			e.notifyInterruptPrompt()
			return true, nil
		}
		if user.Heal < xValue {
			return true, fmt.Errorf("治疗不足，无法移除%d点治疗", xValue)
		}
		removed, err := removeCardsByIndicesFromHand(user, append([]int{}, selected...))
		if err != nil {
			return true, err
		}
		e.NotifyCardRevealed(user.ID, removed, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, removed...)
		user.Heal -= xValue
		targetIDs := runtimeutil.ParseStringSliceContextValue(ctxData["selected_target_ids"])
		yValue := 0
		for _, targetID := range targetIDs {
			if target := e.State.Players[targetID]; target != nil && target.Heal > 0 {
				yValue++
			}
		}
		damage := yValue + 2
		for _, targetID := range targetIDs {
			if e.State.Players[targetID] == nil {
				continue
			}
			e.AddPendingDamage(model.PendingDamage{SourceID: user.ID, TargetID: targetID, Damage: damage, DamageType: model.AttackDamage})
		}
		e.Log(fmt.Sprintf("%s 的 [圣光爆裂] 分支②生效：移除%d治疗并弃%d张牌，对%d名目标各造成%d点攻击伤害（Y=%d）", user.Name, xValue, xValue, len(targetIDs), damage, yValue))
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if !e.routePendingDamageWithReturn(model.TurnStageExtraAction) {
				e.enterExtraActionStage()
			}
		}
		return true, nil

	case "hb_meteor_bullet_cost":
		modes := runtimeutil.ParseStringSliceContextValue(ctxData["cost_modes"])
		if selectionIndex < 0 || selectionIndex >= len(modes) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		ctxData["chosen_cost_mode"] = modes[selectionIndex]
		ctxData["choice_type"] = "hb_meteor_bullet_target"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil

	case "hb_meteor_bullet_target":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		allyIDs := runtimeutil.ParseStringSliceContextValue(ctxData["ally_ids"])
		if selectionIndex < 0 || selectionIndex >= len(allyIDs) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		targetID := allyIDs[selectionIndex]
		mode, _ := ctxData["chosen_cost_mode"].(string)
		switch mode {
		case "heal":
			if user.Heal <= 0 {
				return true, fmt.Errorf("治疗不足，无法发动流星圣弹")
			}
			user.Heal--
		case "faith":
			if holyBowFaith(user) <= 0 {
				return true, fmt.Errorf("信仰不足，无法发动流星圣弹")
			}
			addHolyBowFaith(user, -1)
		default:
			return true, fmt.Errorf("流星圣弹资源选择无效")
		}
		e.Heal(targetID, 1)
		targetName := targetID
		if target := e.State.Players[targetID]; target != nil {
			targetName = target.Name
		}
		e.Log(fmt.Sprintf("%s 发动 [流星圣弹]：移除1点%s，令 %s +1治疗", user.Name, map[string]string{"heal": "治疗", "faith": "信仰"}[mode], targetName))
		rawCtx, _ := ctxData["user_ctx"].(*model.Context)
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if rawCtx != nil && rawCtx.AttackDeclaredPhase() {
				if len(e.State.ActionQueue) > 0 {
					e.enterActionExecutionStage()
				} else if len(e.State.CombatStack) > 0 {
					if e.State.CombatStage == model.CombatStageNone {
						e.setCombatStage(model.CombatStageHitCheck)
					}
					e.clearSubflow()
				} else {
					e.enterTurnEndStage()
				}
			} else {
				e.enterExtraActionStage()
			}
		}
		return true, nil

	case "hb_radiant_cannon_side":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		if !e.isHolyBow(user) {
			return true, fmt.Errorf("仅圣弓可发动圣煌辉光炮")
		}
		if user.Tokens == nil {
			user.Tokens = map[string]int{}
		}
		requiredFaith := runtimeutil.ToIntContextValue(ctxData["required_faith"])
		if requiredFaith <= 0 {
			requiredFaith = 4
		}
		if holyBowCannon(user) <= 0 {
			return true, fmt.Errorf("圣煌辉光炮指示物不足")
		}
		if holyBowFaith(user) < requiredFaith {
			return true, fmt.Errorf("信仰不足，需要%d点", requiredFaith)
		}
		if selectionIndex != 0 && selectionIndex != 1 {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		user.Tokens["hb_cannon"] = holyBowCannon(user) - 1
		addHolyBowFaith(user, -requiredFaith)
		for _, playerID := range e.State.PlayerOrder {
			player := e.State.Players[playerID]
			if player == nil {
				continue
			}
			if len(player.Hand) > 4 {
				discarded := append([]model.Card{}, player.Hand[4:]...)
				player.Hand = append([]model.Card{}, player.Hand[:4]...)
				e.NotifyCardRevealed(player.ID, discarded, "discard")
				e.State.DiscardPile = append(e.State.DiscardPile, discarded...)
			} else if len(player.Hand) < 4 {
				drawN := 4 - len(player.Hand)
				cards, newDeck, newDiscard := rules.DrawCards(e.State.Deck, e.State.DiscardPile, drawN)
				e.State.Deck = newDeck
				e.State.DiscardPile = newDiscard
				player.Hand = append(player.Hand, cards...)
				e.NotifyDrawCards(player.ID, drawN, "hb_radiant_cannon_adjust")
			}
		}
		e.addCampCup(user.Camp)
		if selectionIndex == 0 {
			e.State.RedMorale = e.State.BlueMorale
		} else {
			e.State.BlueMorale = e.State.RedMorale
		}
		e.Log(fmt.Sprintf("%s 发动 [圣煌辉光炮]：全员手牌调整至4，我方星杯+1，并完成士气对齐", user.Name))
		e.checkGameEnd()
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			if !e.routePendingDamageWithReturn(model.TurnStageExtraAction) {
				e.enterExtraActionStage()
			}
		}
		return true, nil

	case "hb_auto_fill_resource":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		modes := runtimeutil.ParseStringSliceContextValue(ctxData["resource_modes"])
		if selectionIndex < 0 || selectionIndex >= len(modes) {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		branch := modes[selectionIndex]
		switch branch {
		case "crystal":
			if !e.ConsumeCrystalCost(user.ID, 1) {
				return true, fmt.Errorf("自动填充分支①需要1点蓝水晶（红宝石可替代）")
			}
		case "gem":
			if user.Gem <= 0 {
				return true, fmt.Errorf("自动填充分支②需要1点红宝石")
			}
			user.Gem--
			maxEnergy := e.getPlayerEnergyCap(user)
			if user.Gem+user.Crystal < maxEnergy {
				user.Crystal++
				if user.Gem+user.Crystal > maxEnergy {
					user.Crystal -= user.Gem + user.Crystal - maxEnergy
				}
			}
		default:
			return true, fmt.Errorf("无效分支")
		}
		ctxData["branch"] = branch
		ctxData["choice_type"] = "hb_auto_fill_gain"
		e.State.PendingInterrupt.Context = ctxData
		e.notifyInterruptPrompt()
		return true, nil

	case "hb_auto_fill_gain":
		userID, _ := ctxData["user_id"].(string)
		user := e.State.Players[userID]
		if user == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		branch, _ := ctxData["branch"].(string)
		if selectionIndex != 0 && selectionIndex != 1 {
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}
		if branch == "gem" {
			if selectionIndex == 0 {
				now := addHolyBowFaith(user, 2)
				e.Log(fmt.Sprintf("%s 的 [自动填充] 分支②生效：+2信仰（当前%d）", user.Name, now))
			} else {
				e.Heal(user.ID, 2)
				e.Log(fmt.Sprintf("%s 的 [自动填充] 分支②生效：+2治疗", user.Name))
			}
		} else {
			if selectionIndex == 0 {
				now := addHolyBowFaith(user, 1)
				e.Log(fmt.Sprintf("%s 的 [自动填充] 分支①生效：+1信仰（当前%d）", user.Name, now))
			} else {
				e.Heal(user.ID, 1)
				e.Log(fmt.Sprintf("%s 的 [自动填充] 分支①生效：+1治疗", user.Name))
			}
		}
		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.enterTurnEndStage()
		}
		return true, nil
	}

	return false, nil
}

func holyBowChoicePlayerOptions(e *GameEngine, playerIDs []string) []model.PromptOption {
	options := make([]model.PromptOption, 0, len(playerIDs))
	for _, playerID := range playerIDs {
		if player := e.State.Players[playerID]; player != nil {
			options = append(options, model.PromptOption{ID: playerID, Label: player.Name})
		}
	}
	return options
}
