package engine

import (
	"errors"
	"fmt"
	"starcup-engine/internal/model"
)

// PerformMagic 发动法术
func (e *GameEngine) PerformMagic(sourceID, targetID string, cardIdx int) error {
	// 1. 验证阶段
	if e.State.Phase != model.PhaseBeforeAction && e.State.Phase != model.PhaseActionExecution {
		return errors.New("当前不是行动阶段")
	}
	player := e.State.Players[sourceID]
	if !player.IsActive {
		return errors.New("不是你的回合")
	}

	// 验证额外行动类型限制
	if player.TurnState.CurrentExtraAction == "Attack" {
		return errors.New("当前是额外攻击行动，只能使用攻击行动")
	}
	if !e.canCastMagicInAction(player) {
		return errors.New("当前形态不能在行动阶段使用法术牌")
	}

	// 2. 验证卡牌
	card, _, _, ok := getPlayableCardByIndex(player, cardIdx)
	if !ok {
		return errors.New("无效的手牌索引")
	}
	if card.Type != model.CardTypeMagic {
		return errors.New("只能使用法术牌")
	}

	if len(player.TurnState.CurrentExtraElement) > 0 {
		isAllowed := false
		for _, allowedEle := range player.TurnState.CurrentExtraElement {
			if card.Element == allowedEle {
				isAllowed = true
				break
			}
		}
		if !isAllowed {
			return fmt.Errorf("属性不符合当前额外行动要求")
		}
	}

	var target *model.Player
	if targetID != "" {
		target = e.State.Players[targetID]
		if target == nil {
			return errors.New("目标不存在")
		}
	} else if card.Name != "魔弹" {
		return errors.New("该法术需要指定目标")
	}

	// 【魔弹融合】检查：魔法少女使用地系或火系非魔弹法术牌时，询问是否当魔弹使用
	// SkipFusionCheck=true 表示已经询问过了，玩家选择正常使用
	if !player.TurnState.SkipFusionCheck && e.isMagicalGirl(player) && card.Name != "魔弹" &&
		(card.Element == model.ElementEarth || card.Element == model.ElementFire) {
		// 先不移除手牌，等玩家确认后再处理
		e.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptMagicBulletFusion,
			PlayerID: player.ID,
			Context: map[string]interface{}{
				"card_idx":  cardIdx,
				"target_id": targetID,
			},
		})
		e.Log(fmt.Sprintf("[Skill] %s 可以发动【魔弹融合】将 %s 当魔弹使用", player.Name, card.Name))
		return nil
	}

	if target != nil {
		e.Log(fmt.Sprintf("[Magic] %s 对 %s 使用了 %s", player.Name, target.Name, card.Name))
	} else {
		e.Log(fmt.Sprintf("[Magic] %s 使用了 %s，按传递顺序自动结算", player.Name, card.Name))
	}

	e.NotifyCardRevealed(sourceID, []model.Card{card}, "magic")

	// 3. 从可打出牌区移除卡牌 (注意：暂时不进弃牌堆，看是否放置到场上)
	if _, err := consumePlayableCardByIndex(player, cardIdx); err != nil {
		return err
	}
	_ = e.maybeAutoReleaseBloodPriestessByHand(player, "手牌<3强制脱离流血形态")

	// 4. 处理效果
	placedOnField := false // 标记卡牌是否留在了场上

	switch card.Name {
	case "魔弹":
		// 【魔弹掌控】检查：魔法少女使用魔弹时，询问是否逆向传递
		if e.isMagicalGirl(player) {
			e.PushInterrupt(&model.Interrupt{
				Type:     model.InterruptMagicBulletDirection,
				PlayerID: player.ID,
				Context: map[string]interface{}{
					"source_id": player.ID,
				},
			})
			e.Log(fmt.Sprintf("[Skill] %s 可以发动【魔弹掌控】选择魔弹传递方向", player.Name))
			return nil
		}
		// 非魔法少女直接执行魔弹
		return e.executeMagicBullet(player, false, false, nil)

		// 此时函数返回 nil，但在 Game 循环中会检测到 PendingInterrupt 并暂停

	case "中毒":
		// 放置场上牌：中毒 (回合开始触发)
		// 规则：同名基础效果最多存在一个
		if target.HasFieldEffect(model.EffectPoison) {
			e.State.DiscardPile = append(e.State.DiscardPile, card)
			return fmt.Errorf("%s 面前已有中毒，不可重复放置", target.Name)
		}
		fc := &model.FieldCard{
			Card:     card,
			OwnerID:  target.ID,
			SourceID: player.ID,
			Mode:     model.FieldEffect,
			Effect:   model.EffectPoison,
			Trigger:  model.EffectTriggerOnTurnStart,
		}
		target.AddFieldCard(fc)
		placedOnField = true
		e.Log(fmt.Sprintf("[Magic] %s 面前放置了【中毒】", target.Name))

	case "虚弱":
		// 规则：每个角色面前同时只能有一个虚弱
		if target.HasFieldEffect(model.EffectWeak) {
			e.State.DiscardPile = append(e.State.DiscardPile, card)
			return fmt.Errorf("%s 面前已有虚弱，不可重复放置", target.Name)
		}
		fc := &model.FieldCard{
			Card:     card,
			OwnerID:  target.ID,
			SourceID: player.ID,
			Mode:     model.FieldEffect,
			Effect:   model.EffectWeak,
			Trigger:  model.EffectTriggerOnTurnStart,
		}
		target.AddFieldCard(fc)
		placedOnField = true
		e.Log(fmt.Sprintf("[Magic] %s 面前放置了【虚弱】", target.Name))

	case "圣盾":
		// 规则：同名基础效果最多存在一个
		if target.HasFieldEffect(model.EffectShield) {
			e.State.DiscardPile = append(e.State.DiscardPile, card)
			return fmt.Errorf("%s 面前已有圣盾，不可重复放置", target.Name)
		}
		fc := &model.FieldCard{
			Card:     card,
			OwnerID:  target.ID,
			SourceID: player.ID,
			Mode:     model.FieldEffect,
			Effect:   model.EffectShield,
			Trigger:  model.EffectTriggerOnDamaged,
		}
		target.AddFieldCard(fc)
		placedOnField = true
		e.Log(fmt.Sprintf("[Magic] %s 获得了【圣盾】保护", target.Name))

	case "圣光":
		// 即时效果：无法主动使用产生Buff，通常用于响应阶段的防御
		// 如果是在主动阶段打出（极为罕见，通常是误操作或特殊技能），这里仅做记录或视为空放
		e.Log(fmt.Sprintf("[Magic] %s 展示了圣光", player.Name))

	default:
		// 如果是未知的法术，默认进弃牌堆，防止卡牌消失
		e.Log(fmt.Sprintf("[Magic] 未知法术效果: %s", card.Name))
	}

	// 5. 如果卡牌没有放置在场上，则进入弃牌堆
	if !placedOnField {
		e.State.DiscardPile = append(e.State.DiscardPile, card)
	}

	// === 【新增】 5. 触发法术行动结束事件 (为了触发法术激荡等技能) ===
	phaseEventCtx := &model.EventContext{
		Type:       model.EventPhaseEnd,
		SourceID:   player.ID,
		Card:       &card,
		ActionType: model.ActionMagic,
	}
	phaseCtx := e.buildContext(player, nil, model.TriggerOnPhaseEnd, phaseEventCtx)
	e.dispatcher.OnTrigger(model.TriggerOnPhaseEnd, phaseCtx)

	return nil
}

// findNextMagicBulletTarget 寻找魔弹的下一个目标
// 规则修正：
// reverse=false: 右手方向（前一位对手，索引递减）
// reverse=true:  逆向（后一位对手，索引递增）
func (e *GameEngine) findNextMagicBulletTarget(currentPID string) string {
	chain := e.State.MagicBulletChain
	reverse := chain != nil && chain.Reverse

	currentPlayer := e.State.Players[currentPID]
	if currentPlayer == nil {
		return ""
	}

	startIdx := -1
	for i, pid := range e.State.PlayerOrder {
		if pid == currentPID {
			startIdx = i
			break
		}
	}

	if startIdx == -1 {
		return ""
	}

	n := len(e.State.PlayerOrder)
	for i := 1; i < n; i++ {
		var idx int
		if reverse {
			// 逆向：后一位（索引递增）
			idx = (startIdx + i) % n
		} else {
			// 默认右手方向：前一位（索引递减）
			idx = (startIdx - i + n) % n
		}
		pid := e.State.PlayerOrder[idx]
		target := e.State.Players[pid]

		// 必须是对手 (不同阵营)
		if target.Camp != currentPlayer.Camp {
			return pid
		}
	}

	return ""
}

// isMagicalGirl 检查玩家是否是魔法少女
func (e *GameEngine) isMagicalGirl(player *model.Player) bool {
	if player == nil || player.Character == nil {
		return false
	}
	return player.Character.ID == "magical_girl" ||
		player.Character.ID == "magic_bullet_girl" ||
		player.Character.Name == "魔法少女" ||
		player.Character.Name == "魔弹少女"
}

// executeMagicBullet 执行魔弹效果
// reverse: 是否逆向传递
// isFusion: 是否由魔弹融合触发
// fusionCard: 融合使用的原始卡牌（如果 isFusion=true）
func (e *GameEngine) executeMagicBullet(player *model.Player, reverse bool, isFusion bool, fusionCard *model.Card) error {
	// 初始化魔弹链条
	e.State.MagicBulletChain = &model.MagicBulletChain{
		CurrentDamage:  2,
		InvolvedIDs:    []string{player.ID}, // 发起者已参与
		SourcePlayerID: player.ID,
		Reverse:        reverse,
		IsFusion:       isFusion,
		FusionCard:     fusionCard,
	}

	// 寻找最近的对手（会根据 Reverse 自动选择方向）
	nextTargetID := e.findNextMagicBulletTarget(player.ID)
	if nextTargetID == "" {
		e.Log("[Magic] 魔弹没有有效目标，自动结束")
		e.State.MagicBulletChain = nil
		return nil
	}

	e.State.MagicBulletChain.TargetID = nextTargetID

	nextTarget := e.State.Players[nextTargetID]

	// 设置中断，等待目标响应
	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptMagicMissile,
		PlayerID: nextTargetID,
		Context: map[string]interface{}{
			"damage":    2,
			"source_id": player.ID,
		},
	})

	direction := "顺时针"
	if reverse {
		direction = "逆时针"
	}
	if isFusion {
		e.Log(fmt.Sprintf("[Magic] 【魔弹融合】%s 将 %s 当魔弹使用，%s传递，指向 %s (伤害: %d)",
			player.Name, fusionCard.Name, direction, nextTarget.Name, 2))
	} else {
		e.Log(fmt.Sprintf("[Magic] 魔弹%s传递，指向 %s (伤害: %d)...",
			direction, nextTarget.Name, 2))
	}

	return nil
}
