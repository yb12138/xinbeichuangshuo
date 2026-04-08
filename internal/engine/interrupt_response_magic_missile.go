package engine

import (
	"fmt"

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
			DamageType: model.MagicDamage,
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
		if res := e.dispatchTimingOnHitCheck(timingOnHitCheckContext{
			Op:         timingOnHitCheckMagicMissileCounter,
			Player:     player,
			MagicChain: chain,
			Card:       card,
		}); res.Err != nil {
			return res.Err
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
		if res := e.dispatchTimingOnHitCheck(timingOnHitCheckContext{
			Op:         timingOnHitCheckMagicMissileDefend,
			Player:     player,
			MagicChain: chain,
		}); res.Err != nil {
			return res.Err
		}
		card, _, _, ok := getPlayableCardByIndex(player, act.CardIndex)
		if !ok {
			return fmt.Errorf("无效的卡牌索引")
		}
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

		e.State.MagicBulletChain = nil
		e.PopInterrupt()
		return nil

	default:
		return fmt.Errorf("未知的响应类型: %s", respType)
	}
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
		e.NotifyCardRevealed(player.ID, []model.Card{card}, model.MagicDamage)
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
