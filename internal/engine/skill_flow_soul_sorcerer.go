// gameflow: 灵魂术士：灵魂链接、吞噬、魔魂等同调多步中断。

package engine

import (
	"fmt"
	"starcup-engine/internal/engine/core/runtimeutil"

	"starcup-engine/internal/model"
)

func (e *GameEngine) handleSoulRecallSelections(playerID string, selections []int) error {
	if e.State.PendingInterrupt == nil || e.State.PendingInterrupt.Type != model.InterruptChoice {
		return fmt.Errorf("当前不存在可处理的灵魂召还弃牌")
	}
	ctxData, ok := e.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		return fmt.Errorf("灵魂召还上下文错误")
	}
	choiceType, _ := ctxData["choice_type"].(string)
	if choiceType != "ss_recall_pick" {
		return fmt.Errorf("当前中断不是灵魂召还选牌")
	}

	userID, _ := ctxData["user_id"].(string)
	if userID == "" {
		userID = playerID
	}
	user := e.State.Players[userID]
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}

	magicIndices := runtimeutil.ParseChoiceIntSlice(ctxData["magic_indices"])
	if len(magicIndices) == 0 {
		magicIndices = runtimeutil.ParseChoiceIntSlice(ctxData["remaining_indices"])
	}
	allowed := make(map[int]struct{}, len(magicIndices))
	orderedCandidates := make([]int, 0, len(magicIndices))
	for _, idx := range magicIndices {
		if idx < 0 || idx >= len(user.Hand) {
			continue
		}
		if user.Hand[idx].Type != model.CardTypeMagic {
			continue
		}
		allowed[idx] = struct{}{}
		orderedCandidates = append(orderedCandidates, idx)
	}
	if len(allowed) == 0 {
		return fmt.Errorf("灵魂召还没有可弃置的法术牌")
	}
	if len(selections) == 0 {
		return fmt.Errorf("灵魂召还至少选择1张法术牌")
	}
	if len(selections) > len(allowed) {
		return fmt.Errorf("选择数量超过可选法术牌数量")
	}

	picked := make([]int, 0, len(selections))
	seen := make(map[int]struct{}, len(selections))
	for _, idx := range selections {
		resolvedIdx, ok := runtimeutil.ResolveSelectionToAllowedIndex(idx, orderedCandidates, allowed)
		if !ok {
			return fmt.Errorf("灵魂召还只能选择法术牌")
		}
		if _, dup := seen[resolvedIdx]; dup {
			return fmt.Errorf("不能重复选择同一张牌")
		}
		seen[resolvedIdx] = struct{}{}
		picked = append(picked, resolvedIdx)
	}

	removed, err := removeCardsByIndicesFromHand(user, picked)
	if err != nil {
		return err
	}
	e.NotifyCardRevealed(user.ID, removed, "discard")
	e.State.DiscardPile = append(e.State.DiscardPile, removed...)
	gain := len(removed)
	before := soulSorcererBlue(user)
	after := addSoulSorcererBlue(user, gain)
	e.Log(fmt.Sprintf("%s 发动 [灵魂召还]：弃置%d张法术牌，蓝色灵魂 +%d（%d→%d）", user.Name, gain, gain, before, after))

	e.PopInterrupt()
	if e.State.PendingInterrupt == nil {
		if !e.routePendingDamageWithReturn(model.TurnStageExtraAction) {
			e.enterExtraActionStage()
		}
	}
	return nil
}
