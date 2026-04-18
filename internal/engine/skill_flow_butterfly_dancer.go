// gameflow: 蝶舞者：蛹、毒粉、伤害传递等。

package engine

import (
	"fmt"
	"starcup-engine/internal/engine/core/runtimeutil"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func (e *GameEngine) butterflyActionTargetIDs() []string {
	if e == nil || e.State == nil {
		return nil
	}
	targetIDs := make([]string, 0, len(e.State.PlayerOrder))
	for _, pid := range e.State.PlayerOrder {
		if e.State.Players[pid] != nil {
			targetIDs = append(targetIDs, pid)
		}
	}
	return targetIDs
}

func (e *GameEngine) ResolveButterflyChrysalis(userID string) error {
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	now := addButterflyPupa(user, 1)
	cards, newDeck, newDiscard := rules.DrawCards(e.State.Deck, e.State.DiscardPile, 4)
	e.State.Deck = newDeck
	e.State.DiscardPile = newDiscard
	added := addButterflyCocoonCards(user, cards)
	e.Log(fmt.Sprintf("%s 发动 [蛹化]：蛹+1（当前%d），获得%d个茧", user.Name, now, added))
	e.checkHandLimit(user, nil)
	overflow := butterflyCocoonCount(user) - butterflyCocoonCapEngine
	if overflow > 0 {
		e.PushInterrupt(&model.Interrupt{Type: model.InterruptChoice, PlayerID: user.ID, Context: map[string]interface{}{"choice_type": "bt_cocoon_overflow_discard", "user_id": user.ID, "discard_count": overflow}})
	}
	return nil
}

func (e *GameEngine) StartButterflyReverse(userID string) error {
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: user.ID,
		Context: map[string]interface{}{
			"choice_type": "bt_reverse_mode",
			"user_id":     user.ID,
			"can_branch2": butterflyPupa(user) > 0,
			"target_ids":  e.butterflyActionTargetIDs(),
		},
	})
	e.Log(fmt.Sprintf("%s 发动 [倒逆之蝶]：已弃2张牌，请选择发动分支", user.Name))
	return nil
}

func (e *GameEngine) handleButterflyCocoonOverflowSelections(playerID string, selections []int) error {
	if e.State.PendingInterrupt == nil || e.State.PendingInterrupt.Type != model.InterruptChoice {
		return fmt.Errorf("当前不存在可处理的茧上限弃置")
	}
	ctxData, ok := e.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		return fmt.Errorf("茧上限上下文错误")
	}
	choiceType, _ := ctxData["choice_type"].(string)
	if choiceType != "bt_cocoon_overflow_discard" {
		return fmt.Errorf("当前中断不是茧上限弃置")
	}
	userID, _ := ctxData["user_id"].(string)
	if userID == "" {
		userID = playerID
	}
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	discardNeed := runtimeutil.ToIntContextValue(ctxData["discard_count"])
	if discardNeed < 0 {
		discardNeed = 0
	}
	cocoonIndices := butterflyCocoonFieldIndices(user)
	if discardNeed > len(cocoonIndices) {
		discardNeed = len(cocoonIndices)
	}
	if len(selections) != discardNeed {
		return fmt.Errorf("需要选择 %d 个茧舍弃", discardNeed)
	}
	picked := make([]int, 0, len(selections))
	seen := make(map[int]struct{}, len(selections))
	for _, selection := range selections {
		fieldIdx, ok := runtimeutil.ResolveSelectionToCandidate(selection, cocoonIndices)
		if !ok {
			return fmt.Errorf("无效的茧索引: %d", selection)
		}
		if _, dup := seen[fieldIdx]; dup {
			return fmt.Errorf("不能重复选择同一个茧")
		}
		seen[fieldIdx] = struct{}{}
		picked = append(picked, fieldIdx)
	}
	removed, err := removeButterflyCocoonByFieldIndices(user, picked)
	if err != nil {
		return err
	}
	if len(removed) > 0 {
		e.NotifyCardHidden(user.ID, removed, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, removed...)
	}
	e.Log(fmt.Sprintf("%s 的 [茧上限] 结算：舍弃%d个茧", user.Name, len(removed)))
	e.PopInterrupt()
	if e.State.PendingInterrupt == nil {
		e.enterExtraActionStage()
	}
	return nil
}

func (e *GameEngine) handleButterflyReverseBranch2PickSelections(playerID string, selections []int) error {
	if e.State.PendingInterrupt == nil || e.State.PendingInterrupt.Type != model.InterruptChoice {
		return fmt.Errorf("当前不存在可处理的倒逆之蝶分支②选茧")
	}
	ctxData, ok := e.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		return fmt.Errorf("倒逆之蝶分支②上下文错误")
	}
	choiceType, _ := ctxData["choice_type"].(string)
	if choiceType != "bt_reverse_branch2_pick" {
		return fmt.Errorf("当前中断不是倒逆之蝶分支②选茧")
	}
	userID, _ := ctxData["user_id"].(string)
	if userID == "" {
		userID = playerID
	}
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	const pickNeed = 2
	if len(selections) != pickNeed {
		return fmt.Errorf("需要选择 %d 个茧", pickNeed)
	}
	cocoonIndices := butterflyCocoonFieldIndices(user)
	picked := make([]int, 0, len(selections))
	seen := make(map[int]struct{}, len(selections))
	for _, selection := range selections {
		fieldIdx, ok := runtimeutil.ResolveSelectionToCandidate(selection, cocoonIndices)
		if !ok {
			return fmt.Errorf("无效的茧索引: %d", selection)
		}
		if _, dup := seen[fieldIdx]; dup {
			return fmt.Errorf("不能重复选择同一个茧")
		}
		seen[fieldIdx] = struct{}{}
		picked = append(picked, fieldIdx)
	}
	removed, err := removeButterflyCocoonByFieldIndices(user, picked)
	if err != nil {
		return err
	}
	if len(removed) > 0 {
		e.NotifyCardRevealed(user.ID, removed, "discard")
		e.State.DiscardPile = append(e.State.DiscardPile, removed...)
	}
	for _, c := range removed {
		if c.Type == model.CardTypeMagic {
			e.queueButterflyWitherFollowup(user)
		}
	}
	now := addButterflyPupa(user, -1)
	e.Log(fmt.Sprintf("%s 的 [倒逆之蝶] 分支②：移除2个茧并移除1个蛹（当前蛹=%d）", user.Name, now))
	e.PopInterrupt()
	if e.State.PendingInterrupt == nil {
		if !e.routePendingDamageWithReturn(model.TurnStageExtraAction) {
			e.enterExtraActionStage()
		}
	}
	return nil
}
