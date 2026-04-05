package engine

import (
	"fmt"
	"strconv"

	"starcup-engine/internal/model"
)

func (e *GameEngine) consumeShieldForMagicMissileTake(target *model.Player, chain *model.MagicBulletChain) bool {
	if target == nil || chain == nil || !target.HasFieldEffect(model.EffectShield) {
		return false
	}
	removed := false
	for _, fc := range target.Field {
		if fc == nil || fc.Mode != model.FieldEffect || fc.Effect != model.EffectShield {
			continue
		}
		target.RemoveFieldCard(fc)
		e.State.DiscardPile = append(e.State.DiscardPile, fc.Card)
		removed = true
		break
	}
	if !removed {
		return false
	}

	e.addActionResponse(fmt.Sprintf("%s 的【圣盾】自动抵挡魔弹", target.Name))
	e.NotifyActionStep(fmt.Sprintf("%s 的【圣盾】触发，自动抵挡了魔弹", target.Name))
	e.Log(fmt.Sprintf("[Magic] %s 选择承受，触发【圣盾】自动抵挡魔弹", target.Name))
	e.State.MagicBulletChain = nil
	e.PopInterrupt()
	return true
}

// handleMagicMissileResponse 处理魔弹响应。
func (e *GameEngine) handleMagicMissileResponse(act model.PlayerAction) error {
	chain := e.State.MagicBulletChain
	if chain == nil {
		return fmt.Errorf("魔弹链条不存在")
	}
	if act.PlayerID != chain.TargetID {
		return fmt.Errorf("不是你的响应回合")
	}

	respType := ""
	if len(act.ExtraArgs) > 0 {
		respType = act.ExtraArgs[0]
	} else {
		return fmt.Errorf("缺少响应类型")
	}

	player := e.State.Players[act.PlayerID]
	if player == nil {
		return fmt.Errorf("玩家不存在")
	}

	switch respType {
	case "take":
		if e.consumeShieldForMagicMissileTake(player, chain) {
			return nil
		}

		damage := chain.CurrentDamage
		e.Log(fmt.Sprintf("[Magic] %s 选择承受魔弹伤害 (%d点)", player.Name, damage))
		magicCard := &model.Card{
			Name:        "魔弹",
			Type:        model.CardTypeMagic,
			Damage:      damage,
			Description: "魔弹伤害",
		}

		e.PopInterrupt()
		e.AddPendingDamage(model.PendingDamage{
			SourceID:   chain.SourcePlayerID,
			TargetID:   player.ID,
			Damage:     damage,
			DamageType: "magic",
			Card:       magicCard,
		})
		e.setReturnPoint(model.TurnStageExtraAction)
		e.enterDamageResolution(nil)
		e.State.MagicBulletChain = nil
		return nil

	case "counter":
		card, _, _, ok := getPlayableCardByIndex(player, act.CardIndex)
		if !ok {
			return fmt.Errorf("无效的卡牌索引")
		}
		if e.isMagicLancer(player) {
			return fmt.Errorf("魔枪受[黑暗束缚]影响，不能使用法术牌")
		}
		if card.Name != "魔弹" {
			return fmt.Errorf("必须使用【魔弹】进行传递")
		}

		hasParticipated := false
		for _, pid := range chain.InvolvedIDs {
			if pid == player.ID {
				hasParticipated = true
				break
			}
		}
		aliveCount := len(e.State.PlayerOrder)
		if hasParticipated {
			return fmt.Errorf("你在本轮传递中已参与过，无法再次传递")
		}

		if _, err := consumePlayableCardByIndex(player, act.CardIndex); err != nil {
			return err
		}
		e.State.DiscardPile = append(e.State.DiscardPile, card)
		e.Log(fmt.Sprintf("[Magic] %s 打出魔弹，将伤害传递给下一位！伤害+1", player.Name))

		chain.CurrentDamage += 1
		chain.SourcePlayerID = player.ID
		chain.InvolvedIDs = append(chain.InvolvedIDs, player.ID)
		if len(chain.InvolvedIDs) >= aliveCount {
			e.Log("[Magic] 本轮魔弹传递已覆盖所有角色，魔弹结算结束")
			e.State.MagicBulletChain = nil
			e.PopInterrupt()
			return nil
		}

		nextTargetID := e.findNextMagicBulletTarget(player.ID)
		if nextTargetID == "" {
			e.Log("[Magic] 没有下一个目标，魔弹失效")
			e.State.MagicBulletChain = nil
			e.PopInterrupt()
			return nil
		}

		nextTarget := e.State.Players[nextTargetID]
		chain.TargetID = nextTargetID
		e.State.PendingInterrupt.PlayerID = nextTargetID
		if ctx, ok := e.State.PendingInterrupt.Context.(map[string]interface{}); ok {
			ctx["damage"] = chain.CurrentDamage
			ctx["source_id"] = player.ID
		}
		e.notifyInterruptPrompt()
		if nextTarget != nil {
			e.Log(fmt.Sprintf("[Magic] 魔弹指向 %s (伤害: %d)，等待响应...", nextTarget.Name, chain.CurrentDamage))
		}
		return nil

	case "defend":
		if e.isMagicLancer(player) {
			return fmt.Errorf("魔枪受[黑暗束缚]影响，不能使用法术牌防御")
		}
		if card, _, _, ok := getPlayableCardByIndex(player, act.CardIndex); ok {
			if card.Name == "圣盾" {
				return fmt.Errorf("【圣盾】不能在防御时打出，请提前放置到场上触发")
			}
			if card.Name != "圣光" {
				return fmt.Errorf("必须使用【圣光】抵挡")
			}
			e.Log(fmt.Sprintf("[Magic] %s 使用【圣光】，抵挡了魔弹", player.Name))
			if _, err := consumePlayableCardByIndex(player, act.CardIndex); err != nil {
				return err
			}
			e.State.DiscardPile = append(e.State.DiscardPile, card)
		} else {
			holyIdx := -1
			for i := 0; i < playableCardCount(player); i++ {
				c, _, _, ok := getPlayableCardByIndex(player, i)
				if !ok {
					continue
				}
				if c.Name == "圣光" {
					holyIdx = i
					break
				}
			}
			if holyIdx < 0 {
				return fmt.Errorf("没有【圣光】可以抵挡（若有场上【圣盾】，可选择承受伤害来自动触发）")
			}
			card, _, _, _ := getPlayableCardByIndex(player, holyIdx)
			e.Log(fmt.Sprintf("[Magic] %s 使用【圣光】，抵挡了魔弹", player.Name))
			if _, err := consumePlayableCardByIndex(player, holyIdx); err != nil {
				return err
			}
			e.State.DiscardPile = append(e.State.DiscardPile, card)
		}

		e.State.MagicBulletChain = nil
		e.PopInterrupt()
		return nil

	default:
		return fmt.Errorf("未知的响应类型: %s", respType)
	}
}

// handleMagicBlastResponse 处理魔爆冲击弃牌响应。
func (e *GameEngine) handleMagicBlastResponse(act model.PlayerAction) error {
	interrupt := e.State.PendingInterrupt
	if interrupt == nil {
		return fmt.Errorf("没有待处理的中断")
	}

	data, ok := interrupt.Context.(map[string]interface{})
	if !ok {
		return fmt.Errorf("中断上下文格式错误")
	}
	stage, _ := data["stage"].(string)
	if stage == "" {
		stage = "target_discard"
	}

	player := e.State.Players[act.PlayerID]
	if player == nil {
		return fmt.Errorf("玩家不存在")
	}

	casterID, _ := data["caster_id"].(string)
	caster := e.State.Players[casterID]
	targetsRaw, _ := data["targets"].([]string)
	if targetsRaw == nil {
		if targetsIface, ok := data["targets"].([]interface{}); ok {
			targetsRaw = make([]string, len(targetsIface))
			for i, v := range targetsIface {
				targetsRaw[i], _ = v.(string)
			}
		}
	}

	currentTargetIdx := 0
	if ct, ok := data["current_target"].(int); ok {
		currentTargetIdx = ct
	} else if ctf, ok := data["current_target"].(float64); ok {
		currentTargetIdx = int(ctf)
	}

	prompt := e.buildMagicBlastPrompt()
	if prompt == nil {
		return fmt.Errorf("魔爆冲击提示构建失败")
	}

	if stage == "caster_forced_discard" {
		if act.Type != model.CmdSelect || len(act.Selections) == 0 {
			return fmt.Errorf("请选择1张牌弃置")
		}
		selection := act.Selections[0]
		if selection < 0 || selection >= len(prompt.Options) {
			return fmt.Errorf("无效的选择")
		}
		cardIdx, err := strconv.Atoi(prompt.Options[selection].ID)
		if err != nil {
			return fmt.Errorf("无效的卡牌索引")
		}
		if cardIdx < 0 || cardIdx >= len(player.Hand) {
			return fmt.Errorf("无效的卡牌索引")
		}

		card := player.Hand[cardIdx]
		player.Hand = append(player.Hand[:cardIdx], player.Hand[cardIdx+1:]...)
		e.State.DiscardPile = append(e.State.DiscardPile, card)
		e.Log(fmt.Sprintf("[Skill] %s 因【魔爆冲击】弃掉了 %s", player.Name, card.Name))

		if currentTargetIdx < len(targetsRaw) {
			data["stage"] = "target_discard"
			e.State.PendingInterrupt.PlayerID = targetsRaw[currentTargetIdx]
			e.State.PendingInterrupt.Context = data
			if nextTarget := e.State.Players[targetsRaw[currentTargetIdx]]; nextTarget != nil {
				e.Log(fmt.Sprintf("[Skill] %s 需要选择弃一张法术牌或受到2点伤害", nextTarget.Name))
			}
			e.notifyInterruptPrompt()
			return nil
		}

		e.PopInterrupt()
		return nil
	}

	discarded := false
	if act.Type == model.CmdCancel {
		discarded = false
	} else if act.Type == model.CmdSelect && len(act.Selections) > 0 {
		selection := act.Selections[0]
		if selection < 0 || selection >= len(prompt.Options) {
			return fmt.Errorf("无效的选择")
		}
		optionID := prompt.Options[selection].ID
		if optionID != "refuse" {
			cardIdx, err := strconv.Atoi(optionID)
			if err != nil {
				return fmt.Errorf("无效的卡牌索引")
			}
			if cardIdx < 0 || cardIdx >= len(player.Hand) {
				return fmt.Errorf("无效的卡牌索引")
			}

			card := player.Hand[cardIdx]
			if card.Type != model.CardTypeMagic {
				return fmt.Errorf("只能弃置法术牌")
			}
			e.NotifyCardRevealed(player.ID, []model.Card{card}, "discard")
			player.Hand = append(player.Hand[:cardIdx], player.Hand[cardIdx+1:]...)
			e.State.DiscardPile = append(e.State.DiscardPile, card)
			e.Log(fmt.Sprintf("[Skill] %s 弃掉了法术牌 %s", player.Name, card.Name))
			discarded = true
		}
	}

	currentTargetIdx++
	data["current_target"] = currentTargetIdx
	if discarded {
		if currentTargetIdx < len(targetsRaw) {
			e.State.PendingInterrupt.PlayerID = targetsRaw[currentTargetIdx]
			e.State.PendingInterrupt.Context = data
			if nextTarget := e.State.Players[targetsRaw[currentTargetIdx]]; nextTarget != nil {
				e.Log(fmt.Sprintf("[Skill] %s 需要选择弃一张法术牌或受到2点伤害", nextTarget.Name))
			}
			e.notifyInterruptPrompt()
			return nil
		}
		e.PopInterrupt()
		return nil
	}

	e.InflictDamage(casterID, player.ID, 2, "magic")
	e.Log(fmt.Sprintf("[Skill] %s 未弃法术牌，受到2点伤害", player.Name))
	if caster != nil && len(caster.Hand) > 0 {
		data["stage"] = "caster_forced_discard"
		e.State.PendingInterrupt.PlayerID = caster.ID
		e.State.PendingInterrupt.Context = data
		e.notifyInterruptPrompt()
		return nil
	}

	if currentTargetIdx < len(targetsRaw) {
		e.State.PendingInterrupt.PlayerID = targetsRaw[currentTargetIdx]
		e.State.PendingInterrupt.Context = data
		if nextTarget := e.State.Players[targetsRaw[currentTargetIdx]]; nextTarget != nil {
			e.Log(fmt.Sprintf("[Skill] %s 需要选择弃一张法术牌或受到2点伤害", nextTarget.Name))
		}
		e.notifyInterruptPrompt()
		return nil
	}

	e.PopInterrupt()
	return nil
}

// handleMagicBulletFusionResponse 处理魔弹融合响应。
func (e *GameEngine) handleMagicBulletFusionResponse(act model.PlayerAction) error {
	interrupt := e.State.PendingInterrupt
	if interrupt == nil {
		return fmt.Errorf("没有待处理的中断")
	}

	data, ok := interrupt.Context.(map[string]interface{})
	if !ok {
		return fmt.Errorf("中断上下文格式错误")
	}

	player := e.State.Players[act.PlayerID]
	if player == nil {
		return fmt.Errorf("玩家不存在")
	}

	cardIdx, _ := data["card_idx"].(int)
	targetID, _ := data["target_id"].(string)
	card, _, _, cardOK := getPlayableCardByIndex(player, cardIdx)
	if !cardOK {
		return fmt.Errorf("无效的卡牌索引")
	}

	choice := 1
	if len(act.Selections) > 0 {
		choice = act.Selections[0]
	}

	e.PopInterrupt()
	if choice == 0 {
		e.Log(fmt.Sprintf("[Skill] %s 发动【魔弹融合】，将 %s 当魔弹使用！", player.Name, card.Name))
		e.NotifyCardRevealed(player.ID, []model.Card{card}, "magic")
		if _, err := consumePlayableCardByIndex(player, cardIdx); err != nil {
			return err
		}
		e.State.DiscardPile = append(e.State.DiscardPile, card)
		e.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptMagicBulletDirection,
			PlayerID: player.ID,
			Context: map[string]interface{}{
				"source_id":   player.ID,
				"is_fusion":   true,
				"fusion_card": card,
			},
		})
		return nil
	}

	e.Log(fmt.Sprintf("[Magic] %s 选择正常使用 %s", player.Name, card.Name))
	return e.performMagic(act.PlayerID, targetID, cardIdx, true)
}

// handleMagicBulletDirectionResponse 处理魔弹掌控响应。
func (e *GameEngine) handleMagicBulletDirectionResponse(act model.PlayerAction) error {
	interrupt := e.State.PendingInterrupt
	if interrupt == nil {
		return fmt.Errorf("没有待处理的中断")
	}

	data, ok := interrupt.Context.(map[string]interface{})
	if !ok {
		return fmt.Errorf("中断上下文格式错误")
	}

	player := e.State.Players[act.PlayerID]
	if player == nil {
		return fmt.Errorf("玩家不存在")
	}

	reverse := false
	if len(act.Selections) > 0 && act.Selections[0] == 1 {
		reverse = true
	}

	isFusion, _ := data["is_fusion"].(bool)
	var fusionCard *model.Card
	if isFusion {
		if fc, ok := data["fusion_card"].(model.Card); ok {
			fusionCard = &fc
		}
	}

	e.PopInterrupt()
	direction := "顺时针"
	if reverse {
		direction = "逆时针"
		e.Log(fmt.Sprintf("[Skill] %s 发动【魔弹掌控】，魔弹将%s传递！", player.Name, direction))
	}
	return e.executeMagicBullet(player, reverse, isFusion, fusionCard)
}

// handleHolySwordDrawResponse 处理圣剑摸X弃X响应。
func (e *GameEngine) handleHolySwordDrawResponse(act model.PlayerAction) error {
	interrupt := e.State.PendingInterrupt
	if interrupt == nil {
		return fmt.Errorf("没有待处理的中断")
	}

	player := e.State.Players[act.PlayerID]
	if player == nil {
		return fmt.Errorf("玩家不存在")
	}

	x := 0
	if len(act.Selections) > 0 {
		x = act.Selections[0]
	}
	if x < 0 || x > 3 {
		x = 0
	}

	e.PopInterrupt()
	if x == 0 {
		e.Log(fmt.Sprintf("[Skill] %s 选择不摸不弃", player.Name))
		e.resumeHolySwordAftermath()
		return nil
	}

	e.DrawCards(player.ID, x)
	e.Log(fmt.Sprintf("[Skill] %s 摸了 %d 张牌", player.Name, x))
	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptDiscard,
		PlayerID: player.ID,
		Context: map[string]interface{}{
			"discard_count":        x,
			"is_holy_sword":        true,
			"stay_in_turn":         true,
			"is_damage_resolution": false,
		},
	})
	e.Log(fmt.Sprintf("[Skill] %s 需要弃 %d 张牌", player.Name, x))
	return nil
}

func (e *GameEngine) resumeHolySwordAftermath() {
	if e.State.PendingInterrupt != nil {
		return
	}
	if e.routePendingDamageWithDefaultReturn(model.TurnStageExtraAction) {
		return
	}
	if e.restoreReturnPoint() {
		return
	}
	e.enterExtraActionStage()
}

// handleSaintHealResponse 处理圣疗分配治疗响应。
func (e *GameEngine) handleSaintHealResponse(act model.PlayerAction) error {
	interrupt := e.State.PendingInterrupt
	if interrupt == nil {
		return fmt.Errorf("没有待处理的中断")
	}

	data, ok := interrupt.Context.(map[string]interface{})
	if !ok {
		return fmt.Errorf("中断上下文格式错误")
	}

	player := e.State.Players[act.PlayerID]
	if player == nil {
		return fmt.Errorf("玩家不存在")
	}

	targetIDs := saintHealTargetIDsFromContext(data)
	if len(targetIDs) == 0 {
		return fmt.Errorf("圣疗缺少目标")
	}

	stage, _ := data["stage"].(string)
	if stage == "allocate_heal" {
		if len(targetIDs) != 2 {
			return fmt.Errorf("圣疗双目标分配配置无效")
		}
		if len(act.Selections) != 1 {
			return fmt.Errorf("请选择一种治疗分配方式")
		}
		choice := act.Selections[0]
		if choice != 0 && choice != 1 {
			return fmt.Errorf("无效的圣疗分配选项: %d", choice)
		}
		allocations := map[string]int{}
		if choice == 0 {
			allocations[targetIDs[0]] = 2
			allocations[targetIDs[1]] = 1
		} else {
			allocations[targetIDs[0]] = 1
			allocations[targetIDs[1]] = 2
		}
		data["allocations"] = allocations
		data["stage"] = "choose_extra_action"
		e.State.PendingInterrupt.Context = data
		e.notifyInterruptPrompt()
		return nil
	}

	allocations := saintHealAllocationsFromContext(data, targetIDs)
	if len(act.Selections) != 1 {
		return fmt.Errorf("请选择额外行动类型")
	}

	extraActionType := "Attack"
	extraActionLabel := "攻击"
	if act.Selections[0] == 1 {
		extraActionType = "Magic"
		extraActionLabel = "法术"
	} else if act.Selections[0] != 0 {
		return fmt.Errorf("无效的额外行动类型选项: %d", act.Selections[0])
	}

	for _, targetID := range targetIDs {
		healAmount := allocations[targetID]
		if healAmount <= 0 {
			continue
		}
		e.Heal(targetID, healAmount)
		if target := e.State.Players[targetID]; target != nil {
			e.Log(fmt.Sprintf("[Skill] %s 获得 %d 点治疗", target.Name, healAmount))
		}
	}

	e.PopInterrupt()
	model.AppendExtraAction(player, "圣疗", extraActionType)
	e.Log(fmt.Sprintf("[Skill] %s 发动 [圣疗]，获得额外%s行动", player.Name, extraActionLabel))
	player.TurnState.HasActed = true
	player.TurnState.LastActionType = string(model.ActionMagic)
	player.TurnState.LastActionCard = nil
	if !e.routePendingDamageWithReturn(model.TurnStageActionEnd) {
		e.enterActionEndStage()
	}
	return nil
}
