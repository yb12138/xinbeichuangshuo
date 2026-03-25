package engine

import (
	"fmt"
	"strconv"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func (e *GameEngine) buildSystemChoicePrompt(choiceType, playerID string, player *model.Player, data map[string]interface{}) *model.Prompt {
	playerName := playerID
	if player != nil && player.Name != "" {
		playerName = player.Name
	}

	switch choiceType {
	case "weak":
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  fmt.Sprintf("【虚弱状态】%s，你需要做出选择：", playerName),
			Options: []model.PromptOption{
				{ID: "0", Label: "跳过行动阶段 (移除虚弱)"},
				{ID: "1", Label: "摸3张牌后继续行动阶段"},
			},
			Min: 1,
			Max: 1,
		}

	case "buy_resource":
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  "战绩区已有4个星石，选择添加宝石或水晶：",
			Options: []model.PromptOption{
				{ID: "0", Label: "添加宝石"},
				{ID: "1", Label: "添加水晶"},
			},
			Min: 1,
			Max: 1,
		}

	case "heal":
		maxHeal := toIntContextValue(data["max_heal"])
		if maxHeal < 0 {
			maxHeal = 0
		}
		options := make([]model.PromptOption, 0, maxHeal+1)
		for i := 0; i <= maxHeal; i++ {
			label := fmt.Sprintf("使用 %d 点治疗", i)
			if i == 0 {
				label = "不使用治疗"
			}
			options = append(options, model.PromptOption{
				ID:    strconv.Itoa(i),
				Label: label,
			})
		}
		return &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  fmt.Sprintf("%s 受到伤害，可选择使用治疗抵消：", playerName),
			Options:  options,
			Min:      1,
			Max:      1,
		}

	case "basic_effect_pick":
		return buildBasicEffectChoicePrompt(playerID, data)

	case "extract":
		options := buildExtractChoicePromptOptions(data["extract_options"])
		if len(options) == 0 {
			return nil
		}
		minSel := toIntContextValue(data["extract_min"])
		maxSel := toIntContextValue(data["extract_max"])
		if minSel < 1 {
			minSel = 1
		}
		if maxSel < 1 {
			maxSel = 2
		}
		message := fmt.Sprintf("战绩区可提炼的星石（共 %d 个）：请选择 %d-%d 个提炼到能量区：", len(options), minSel, maxSel)
		return &model.Prompt{
			Type:     model.PromptChooseExtract,
			PlayerID: playerID,
			Message:  message,
			Options:  options,
			Min:      minSel,
			Max:      maxSel,
		}
	}

	return nil
}

func (e *GameEngine) handleSystemChoiceInput(playerID string, selectionIndex int, ctxData map[string]interface{}) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)

	switch choiceType {
	case "weak":
		player := e.State.Players[playerID]
		if player == nil {
			return true, fmt.Errorf("玩家不存在")
		}
		if e.RemoveFieldCardBy(player.ID, model.EffectWeak, "") {
			e.Log(fmt.Sprintf("[System] %s 的虚弱状态已移除", player.Name))
		}

		switch selectionIndex {
		case 0:
			e.Log(fmt.Sprintf("[Weak] %s 选择跳过行动阶段", player.Name))
			if player.Tokens == nil {
				player.Tokens = map[string]int{}
			}
			player.Tokens["arbiter_skip_forced_doomsday"] = 1
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				e.enterTurnEndStage()
			}
			return true, nil

		case 1:
			e.Log(fmt.Sprintf("[Weak] %s 选择摸3张牌后继续行动阶段", player.Name))
			cards, newDeck, newDiscard := rules.DrawCards(e.State.Deck, e.State.DiscardPile, 3)
			e.State.Deck = newDeck
			e.State.DiscardPile = newDiscard
			player.Hand = append(player.Hand, cards...)
			e.NotifyDrawCards(player.ID, 3, "weak_choice")

			checkCtx := e.buildContext(player, nil, model.TriggerNone, nil)
			checkCtx.Flags["StayInTurn"] = true
			e.checkHandLimit(player, checkCtx)

			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				e.setTurnStage(model.TurnStageActionStart)
				e.clearCombatStage()
				e.clearSubflow()
			}
			return true, nil
		}

		return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)

	case "buy_resource":
		camp, _ := ctxData["camp"].(string)
		if camp == "" {
			if player := e.State.Players[playerID]; player != nil {
				camp = string(player.Camp)
			}
		}
		if camp == "" {
			return true, fmt.Errorf("购买资源选择缺少阵营信息")
		}

		switch selectionIndex {
		case 0:
			e.ModifyGem(camp, 1)
			e.Log(fmt.Sprintf("[Action] 购买结算：%s 阵营战绩区 +1 宝石", camp))
		case 1:
			e.ModifyCrystal(camp, 1)
			e.Log(fmt.Sprintf("[Action] 购买结算：%s 阵营战绩区 +1 水晶", camp))
		default:
			return true, fmt.Errorf("无效的选项索引: %d", selectionIndex)
		}

		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.enterExtraActionStage()
		}
		return true, nil

	case "heal":
		damageIdx := toIntContextValue(ctxData["damage_index"])
		if damageIdx < 0 || damageIdx >= len(e.State.PendingDamageQueue) {
			return true, fmt.Errorf("伤害上下文不存在")
		}
		pd := &e.State.PendingDamageQueue[damageIdx]
		target := e.State.Players[pd.TargetID]
		if target == nil {
			return true, fmt.Errorf("目标不存在")
		}
		healToUse := selectionIndex
		if healToUse < 0 {
			healToUse = 0
		}
		if healToUse > target.Heal {
			healToUse = target.Heal
		}
		if healToUse > pd.Damage {
			healToUse = pd.Damage
		}
		if healToUse > 0 {
			target.Heal -= healToUse
			pd.Damage -= healToUse
			e.Log(fmt.Sprintf("[Combat] %s 使用 %d 点治疗抵消伤害", target.Name, healToUse))
		} else {
			e.Log(fmt.Sprintf("[Combat] %s 选择不使用治疗", target.Name))
		}
		pd.HealResolved = true

		e.PopInterrupt()
		if e.State.PendingInterrupt == nil {
			e.enterDamageResolution(nil)
		}
		return true, nil

	case "basic_effect_pick":
		return true, e.handleBasicEffectChoiceInput(playerID, selectionIndex, ctxData)
	}

	return false, nil
}

func buildExtractChoicePromptOptions(raw interface{}) []model.PromptOption {
	options := make([]model.PromptOption, 0)
	optsIfaces, ok := raw.([]interface{})
	if !ok {
		return options
	}
	for i, item := range optsIfaces {
		om, _ := item.(map[string]interface{})
		if om == nil {
			continue
		}
		label := "红宝石"
		if typ, _ := om["type"].(string); typ == "crystal" {
			label = "蓝水晶"
		}
		options = append(options, model.PromptOption{
			ID:    strconv.Itoa(i),
			Label: label,
		})
	}
	return options
}

func (e *GameEngine) handleExtractChoiceSelections(playerID string, selections []int) error {
	if e.State.PendingInterrupt == nil {
		return fmt.Errorf("没有待处理的中断")
	}
	data, ok := e.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		return fmt.Errorf("中断上下文格式错误")
	}
	player := e.State.Players[playerID]
	if player == nil {
		return fmt.Errorf("玩家不存在")
	}

	optsRaw, _ := data["extract_options"]
	optsIfaces, ok := optsRaw.([]interface{})
	if !ok || len(selections) == 0 {
		return fmt.Errorf("请选择要提炼的星石")
	}
	minSel := toIntContextValue(data["extract_min"])
	maxSel := toIntContextValue(data["extract_max"])
	if minSel < 1 {
		minSel = 1
	}
	if maxSel < 1 {
		maxSel = 2
	}
	if len(selections) < minSel || len(selections) > maxSel {
		return fmt.Errorf("请选择 %d-%d 个星石提炼", minSel, maxSel)
	}

	extractedGems := 0
	extractedCrystals := 0
	seen := make(map[int]bool, len(selections))
	for _, sel := range selections {
		if sel < 0 || sel >= len(optsIfaces) || seen[sel] {
			return fmt.Errorf("无效的提炼选择")
		}
		seen[sel] = true
		om, _ := optsIfaces[sel].(map[string]interface{})
		if om == nil {
			return fmt.Errorf("提炼选项格式错误")
		}
		switch typ, _ := om["type"].(string); typ {
		case "gem":
			extractedGems++
		case "crystal":
			extractedCrystals++
		}
	}

	selfRoom := toIntContextValue(data["extract_self_room"])
	if selfRoom < 0 {
		selfRoom = 0
	}
	maxAllyRoom := toIntContextValue(data["extract_max_ally_room"])
	allowParadise, _ := data["extract_allow_paradise"].(bool)
	totalExtracted := extractedGems + extractedCrystals
	requiresParadise := totalExtracted > selfRoom
	if requiresParadise && (!allowParadise || maxAllyRoom < totalExtracted) {
		return fmt.Errorf("本次提炼超出自身能量上限，且没有可承接的队友")
	}

	if player.Camp == model.RedCamp {
		if extractedGems > e.State.RedGems || extractedCrystals > e.State.RedCrystals {
			return fmt.Errorf("战绩区星石不足")
		}
		e.State.RedGems -= extractedGems
		e.State.RedCrystals -= extractedCrystals
	} else {
		if extractedGems > e.State.BlueGems || extractedCrystals > e.State.BlueCrystals {
			return fmt.Errorf("战绩区星石不足")
		}
		e.State.BlueGems -= extractedGems
		e.State.BlueCrystals -= extractedCrystals
	}
	e.recordAdventurerExtractResult(player, extractedGems, extractedCrystals, requiresParadise)
	if requiresParadise {
		e.State.PendingInterrupt = &model.Interrupt{
			Type:     model.InterruptResponseSkill,
			PlayerID: player.ID,
			SkillIDs: []string{"adventurer_paradise"},
			Context: e.buildContext(player, nil, model.TriggerNone, &model.EventContext{
				Type:       model.EventNone,
				SourceID:   player.ID,
				ActionType: model.ActionExtract,
			}),
		}
		e.Log(fmt.Sprintf("[Action] %s 提炼：获得 %d 宝石 %d 水晶，必须先通过[冒险者天堂]分配给队友",
			player.Name, extractedGems, extractedCrystals))
		return nil
	}

	player.Gem += extractedGems
	player.Crystal += extractedCrystals
	e.Log(fmt.Sprintf("[Action] %s 提炼：从战绩区获得 %d 宝石 %d 水晶（当前能量: %d）",
		player.Name, extractedGems, extractedCrystals, player.Gem+player.Crystal))

	if e.playerHasSkill(player, "adventurer_paradise") &&
		len(e.adventurerParadiseEligibleAllies(player, totalExtracted)) > 0 {
		e.State.PendingInterrupt = &model.Interrupt{
			Type:     model.InterruptResponseSkill,
			PlayerID: player.ID,
			SkillIDs: []string{"adventurer_paradise"},
			Context: e.buildContext(player, nil, model.TriggerNone, &model.EventContext{
				Type:       model.EventNone,
				SourceID:   player.ID,
				ActionType: model.ActionExtract,
			}),
		}
		e.Log(fmt.Sprintf("[Action] %s 提炼完成，可发动[冒险者天堂]转移本次提炼结果", player.Name))
		return nil
	}

	e.clearAdventurerExtractState(player)
	e.PopInterrupt()
	if e.State.PendingInterrupt == nil {
		e.enterTurnEndStage()
	}
	return nil
}
