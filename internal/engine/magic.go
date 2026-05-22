// gameflow: 法术行动：消耗、目标、进入战斗或直伤等。

package engine

import (
	"errors"
	"fmt"

	"starcup-engine/internal/model"
)

// PerformMagic 发动法术。测试辅助可继续传 index；生产路径使用 PerformMagicByID。
func (e *GameEngine) PerformMagic(sourceID, targetID string, cardIdx int) error {
	player := e.State.Players[sourceID]
	if player == nil {
		return errors.New("玩家不存在")
	}
	card, _, _, ok := e.getPlayableCardByIndex(player, cardIdx)
	if !ok {
		return errors.New("无效的手牌索引")
	}
	return e.PerformMagicByID(sourceID, targetID, card.ID)
}

// PerformMagicByID 发动法术。
func (e *GameEngine) PerformMagicByID(sourceID, targetID, cardID string) error {
	// 1. 验证阶段
	if e.State.Subflow != model.SubflowNone ||
		e.State.CombatStage != model.CombatStageNone ||
		e.State.TurnStage != model.TurnStageActionExecution {
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
	card, _, _, ok := e.getPlayableCardByID(player, cardID)
	if !ok {
		return errors.New("无效的卡牌ID")
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

	if e.dispatchMagicRulebookTiming(model.TimingMagicDeclare, player, target, &card) {
		return nil
	}
	if e.dispatchMagicRulebookTiming(model.TimingMagicSelectTarget, player, target, &card) {
		return nil
	}
	if e.dispatchMagicRulebookTiming(model.TimingMagicValidate, player, target, &card) {
		return nil
	}

	if target != nil {
		e.Log(fmt.Sprintf("[Magic] %s 对 %s 使用了 %s", player.Name, target.Name, card.Name))
	} else {
		e.Log(fmt.Sprintf("[Magic] %s 使用了 %s，按传递顺序自动结算", player.Name, card.Name))
	}

	e.NotifyCardRevealed(sourceID, []model.Card{card}, "magic")

	// 3. 从可打出牌区移除卡牌 (注意：暂时不进弃牌堆，看是否放置到场上)
	if _, err := e.consumePlayableCardByID(player, cardID); err != nil {
		return err
	}

	// 4. 处理效果
	if e.dispatchMagicRulebookTiming(model.TimingMagicResolve, player, target, &card) {
		return nil
	}

	placedOnField := false // 标记卡牌是否留在了场上

	switch card.Name {
	case "魔弹":
		// 【魔弹掌控】检查：魔法少女使用魔弹时，询问是否逆向传递
		if roleRegistry.Entry(player.Character.ID).MagicBullet.CanDirect {
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
			Hook:     model.FieldHookOnBeforeAction,
		}
		target.AddFieldCard(fc)
		e.emitBuffAddedDispatch(player.ID, target.ID, fc.Effect)
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
			Hook:     model.FieldHookOnBeforeAction,
		}
		target.AddFieldCard(fc)
		e.emitBuffAddedDispatch(player.ID, target.ID, fc.Effect)
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
			Hook:     model.FieldHookOnDamaged,
		}
		target.AddFieldCard(fc)
		e.emitBuffAddedDispatch(player.ID, target.ID, fc.Effect)
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
		if target == nil {
			continue
		}

		// 必须是对手 (不同阵营)
		if target.Camp != currentPlayer.Camp {
			return pid
		}
	}

	return ""
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
	e.offerMagicMissileResponseSkills()

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

func (e *GameEngine) offerMagicMissileResponseSkills() {
	if e == nil || e.State == nil || e.State.PendingInterrupt == nil || e.State.PendingInterrupt.Type != model.InterruptMagicMissile {
		return
	}
	chain := e.State.MagicBulletChain
	if chain == nil || chain.TargetID == "" {
		return
	}
	player := e.State.Players[chain.TargetID]
	if player == nil {
		return
	}
	skillIDs := e.applyTimingOnMagicMissileResponseSkillAugment(nil, player, chain)
	if len(skillIDs) == 0 {
		return
	}

	missileInterrupt := cloneInterrupt(e.State.PendingInterrupt)
	ctx := e.BuildContext(player, player, model.TimingOnHitCheck, &model.EventContext{
		Type:     model.EventMagic,
		SourceID: chain.SourcePlayerID,
		TargetID: chain.TargetID,
	})
	ctx.Selections["magic_missile_interrupt"] = missileInterrupt
	ctx.Selections["magic_missile_response"] = true
	ctx.Selections["magic_missile_chain"] = chain
	e.PopInterrupt()
	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptResponseSkill,
		PlayerID: player.ID,
		SkillIDs: skillIDs,
		Context:  ctx,
	})
}

func cloneInterrupt(intr *model.Interrupt) *model.Interrupt {
	if intr == nil {
		return nil
	}
	copied := *intr
	copied.SkillIDs = append([]string{}, intr.SkillIDs...)
	return &copied
}

func (e *GameEngine) resumeMagicMissileAfterResponseSkill(ctx *model.Context, missileInterrupt *model.Interrupt) bool {
	if e == nil || ctx == nil || missileInterrupt == nil {
		return false
	}
	chain, _ := ctx.Selections["magic_missile_chain"].(*model.MagicBulletChain)
	if chain == nil || chain.TargetID == "" || ctx.User == nil {
		return false
	}
	resolved, _ := ctx.Selections["magic_missile_fusion_chain_resolved"].(bool)
	if !resolved {
		return false
	}
	aliveCount := len(e.State.PlayerOrder)
	if len(chain.InvolvedIDs) >= aliveCount {
		e.Log("[Magic] 本轮魔弹传递已覆盖所有角色，魔弹结算结束")
		e.State.MagicBulletChain = nil
		return true
	}
	nextTargetID := e.findNextMagicBulletTarget(ctx.User.ID)
	if nextTargetID == "" {
		e.Log("[Magic] 没有下一个目标，魔弹失效")
		e.State.MagicBulletChain = nil
		return true
	}
	nextTarget := e.State.Players[nextTargetID]
	chain.TargetID = nextTargetID
	missileInterrupt.PlayerID = nextTargetID
	missileInterrupt.Context = map[string]interface{}{
		"damage":    chain.CurrentDamage,
		"source_id": ctx.User.ID,
	}
	e.State.SetPendingInterrupt(missileInterrupt)
	e.syncGamePhaseWithInterrupt(missileInterrupt)
	e.offerMagicMissileResponseSkills()
	if e.State.PendingInterrupt == missileInterrupt {
		e.NotifyInterruptPrompt()
	}
	if nextTarget != nil {
		e.Log(fmt.Sprintf("[Magic] 魔弹指向 %s (伤害: %d)，等待响应...", nextTarget.Name, chain.CurrentDamage))
	}
	return true
}
