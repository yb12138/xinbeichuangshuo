// gameflow: 手牌超限：弃牌中断与结算顺序。

package engine

import (
	"fmt"
	"sort"
	"starcup-engine/internal/engine/core/runtimeutil"

	playerpkg "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

type discardPhaseCandidate int

const (
	discardPhaseReturn discardPhaseCandidate = iota
	discardPhasePendingDamage
	discardPhaseActionQueue
)

func (e *GameEngine) handleDiscardSelection(playerID string, indices []int, data map[string]interface{}) error {
	discardCount := runtimeutil.ToIntContextValue(data["discard_count"])
	if len(indices) != discardCount {
		return fmt.Errorf("需要选择 %d 张牌丢弃，你选择了 %d 张", discardCount, len(indices))
	}
	if allowed := playerpkg.ParseIntSliceContextValue(data["remaining_indices"]); len(allowed) > 0 {
		allowedSet := make(map[int]struct{}, len(allowed))
		for _, idx := range allowed {
			allowedSet[idx] = struct{}{}
		}
		for _, idx := range indices {
			if _, ok := allowedSet[idx]; !ok {
				return fmt.Errorf("索引 %d 不是当前可弃置的牌", idx)
			}
		}
	}

	player := e.State.Players[playerID]
	if player == nil {
		return fmt.Errorf("玩家不存在")
	}

	discardedCards, err := e.discardCardsFromHand(player, indices)
	if err != nil {
		return err
	}
	e.notifyHiddenDiscard(playerID, discardedCards)

	finalLoss, pending, err := e.resolveDiscardSelectionMoraleLoss(player, discardedCards, data)
	if err != nil || pending {
		return err
	}

	e.Log(fmt.Sprintf("[System] %s 丢弃了 %d 张牌！士气 -%d", player.Name, len(discardedCards), finalLoss))

	if e.processAfterDiscardFlowContinuations(player.ID, discardedCards, data) {
		// 角色后续可能已插入新的中断；若无中断，继续恢复流程。
	}

	e.PopInterrupt()
	if e.State.PendingInterrupt == nil {
		if hasChoiceResumePoint(data["draw_resume_phase"]) && !e.hasReturnPoint() {
			e.setReturnPoint(data["draw_resume_phase"])
		}
		e.processFlowContinuations(model.FlowContinuationAfterDraw)
		e.restorePhaseAfterDiscardResolution(
			runtimeutil.ToBoolContextValue(data["stay_in_turn"]),
			runtimeutil.ToBoolContextValue(data["is_damage_resolution"]),
			[]discardPhaseCandidate{discardPhasePendingDamage, discardPhaseReturn},
			[]discardPhaseCandidate{discardPhasePendingDamage, discardPhaseReturn, discardPhaseActionQueue},
		)
	}

	e.checkGameEnd()
	return nil
}

func (e *GameEngine) discardCardsFromHand(player *model.Player, indices []int) ([]model.Card, error) {
	if player == nil {
		return nil, fmt.Errorf("玩家不存在")
	}

	seen := make(map[int]bool, len(indices))
	for _, idx := range indices {
		if idx < 0 || idx >= len(player.Hand) {
			return nil, fmt.Errorf("无效的牌索引: %d", idx)
		}
		if seen[idx] {
			return nil, fmt.Errorf("不能重复选择同一张牌")
		}
		seen[idx] = true
	}

	selected := append([]int(nil), indices...)
	sort.Sort(sort.Reverse(sort.IntSlice(selected)))

	discardedCards := make([]model.Card, 0, len(selected))
	for _, idx := range selected {
		discardedCards = append(discardedCards, player.Hand[idx])
		player.Hand = append(player.Hand[:idx], player.Hand[idx+1:]...)
	}
	return discardedCards, nil
}

func (e *GameEngine) notifyHiddenDiscard(playerID string, discardedCards []model.Card) {
	e.suppressSealOnDiscard = true
	e.NotifyCardHidden(playerID, discardedCards, "discard")
	e.suppressSealOnDiscard = false
}

func (e *GameEngine) resolveDiscardSelectionMoraleLoss(player *model.Player, discardedCards []model.Card, data map[string]interface{}) (int, bool, error) {
	moraleLoss := len(discardedCards)
	finalLoss := moraleLoss
	noMoraleLoss := runtimeutil.ToBoolContextValue(data["no_morale_loss"])
	overflowMoraleLossFixed := runtimeutil.ToIntContextValue(data["overflow_morale_loss_fixed"])
	stayInTurn := runtimeutil.ToBoolContextValue(data["stay_in_turn"])
	isDamageResolution := runtimeutil.ToBoolContextValue(data["is_damage_resolution"])

	if noMoraleLoss {
		moraleLoss = 0
		finalLoss = 0
	}
	if overflowMoraleLossFixed > 0 && moraleLoss > 0 {
		moraleLoss = overflowMoraleLossFixed
	}

	if moraleLoss <= 0 {
		e.State.DiscardPile = append(e.State.DiscardPile, discardedCards...)
		return finalLoss, false, nil
	}

	victimID, _ := data["victim_id"].(string)
	fromDamageDraw := runtimeutil.ToBoolContextValue(data["from_damage_draw"])
	victim := e.State.Players[victimID]
	if victim == nil {
		e.State.DiscardPile = append(e.State.DiscardPile, discardedCards...)
		return 0, false, nil
	}

	isMagic := runtimeutil.ToBoolContextValue(data["is_magic"])
	moraleLoss = e.capMoraleLoss(victim.Camp, moraleLoss, playerpkg.MoraleLossModifierExtra{
		Victim:             victim,
		FromDamageDraw:     fromDamageDraw,
		IsDamageResolution: isDamageResolution,
	})
	if moraleLoss <= 0 {
		finalLoss = e.applyMoraleLossAfterTimingWindow(victim, moraleLoss, isMagic, fromDamageDraw, overflowMoraleLossFixed, discardedCards, nil)
		return finalLoss, false, nil
	}

	lossCtx := e.buildDiscardMoraleLossContext(victim, player, discardedCards, moraleLoss, isMagic, fromDamageDraw, stayInTurn, isDamageResolution, data)
	e.dispatcher.OnTiming(lossCtx.Timing, lossCtx)
	if e.hasQueuedMoraleLossResponse() {
		lossCtx.Selections["morale_loss_pending"] = true
		lossCtx.Selections["morale_loss_value"] = moraleLoss
		lossCtx.Selections["is_magic"] = isMagic
		lossCtx.Selections["overflow_morale_loss_fixed"] = overflowMoraleLossFixed
		e.PopInterrupt()
		return 0, true, nil
	}

	finalLoss = e.applyMoraleLossAfterTimingWindow(victim, moraleLoss, isMagic, fromDamageDraw, overflowMoraleLossFixed, discardedCards, lossCtx)
	return finalLoss, false, nil
}

func (e *GameEngine) buildDiscardMoraleLossContext(victim *model.Player, player *model.Player, discardedCards []model.Card, moraleLoss int, isMagic bool, fromDamageDraw bool, stayInTurn bool, isDamageResolution bool, data map[string]interface{}) *model.Context {
	lossEventCtx := &model.EventContext{
		Type:      model.EventDamage,
		DamageVal: &moraleLoss,
	}
	lossCtx := e.buildContext(victim, nil, model.TimingBeforeMoraleLoss, lossEventCtx)
	lossCtx.Flags["IsMagicDamage"] = isMagic
	if lossCtx.Selections == nil {
		lossCtx.Selections = map[string]any{}
	}
	lossCtx.Selections["discarded_cards"] = append([]model.Card{}, discardedCards...)
	lossCtx.Selections["from_damage_draw"] = fromDamageDraw
	lossCtx.Selections["victim_id"] = victim.ID
	lossCtx.Selections["discard_player_id"] = player.ID
	lossCtx.Selections["morale_loss_stay_in_turn"] = stayInTurn
	lossCtx.Selections["morale_loss_is_damage_resolution"] = isDamageResolution
	lossCtx.Selections["damage_source_id"] = data["damage_source_id"]
	lossCtx.Selections["damage_source_skill"] = data["damage_source_skill"]
	return lossCtx
}

func (e *GameEngine) hasQueuedMoraleLossResponse() bool {
	for _, intr := range e.State.InterruptQueue {
		if intr != nil && intr.Type == model.InterruptResponseSkill {
			return true
		}
	}
	return false
}

func (e *GameEngine) processAfterDiscardFlowContinuations(playerID string, discardedCards []model.Card, data map[string]interface{}) bool {
	if e == nil || e.State == nil || playerID == "" {
		return false
	}
	if roleID, ok := data["flow_continuation_role_id"].(string); ok && roleID != "" {
		cont := e.buildAfterDiscardFlowContinuation(roleID, playerID, discardedCards, data)
		e.AppendFlowContinuation(cont)
		e.processFlowContinuations(model.FlowContinuationAfterDiscard)
		return true
	}
	return false
}

func (e *GameEngine) buildAfterDiscardFlowContinuation(roleID string, discardPlayerID string, discardedCards []model.Card, data map[string]interface{}) model.FlowContinuation {
	playerID := discardPlayerID
	if explicitPlayerID, _ := data["flow_continuation_player_id"].(string); explicitPlayerID != "" {
		playerID = explicitPlayerID
	}
	skillID, _ := data["flow_continuation_skill_id"].(string)
	if skillID == "" {
		skillID, _ = data["skill_id"].(string)
	}
	cont := model.FlowContinuation{
		Kind:     model.FlowContinuationAfterDiscard,
		RoleID:   roleID,
		PlayerID: playerID,
		SkillID:  skillID,
		Data: map[string]any{
			"discard_player_id": discardPlayerID,
			"discarded_cards":   append([]model.Card{}, discardedCards...),
		},
	}
	for key, value := range data {
		if _, exists := cont.Data[key]; exists {
			continue
		}
		cont.Data[key] = value
	}
	return cont
}

func (e *GameEngine) restorePhaseAfterDiscardResolution(stayInTurn bool, isDamageResolution bool, damageResolutionOrder []discardPhaseCandidate, stayInTurnOrder []discardPhaseCandidate) {
	if isDamageResolution {
		e.restoreDiscardResolutionPhaseByOrder(damageResolutionOrder, model.TurnStageExtraAction, false)
		return
	}
	if stayInTurn {
		e.restoreDiscardResolutionPhaseByOrder(stayInTurnOrder, model.TurnStageActionStart, true)
		return
	}
	e.enterTurnEndStage()
}

func (e *GameEngine) restoreDiscardResolutionPhaseByOrder(order []discardPhaseCandidate, fallback model.TurnStage, logStayInTurn bool) {
	if logStayInTurn {
		e.Log("[System] 弃牌完成，继续当前回合")
	}
	for _, candidate := range order {
		switch candidate {
		case discardPhaseReturn:
			if e.restoreReturnPoint() {
				return
			}
		case discardPhasePendingDamage:
			if len(e.State.PendingDamageQueue) > 0 {
				e.enterDamageResolution(nil)
				return
			}
		case discardPhaseActionQueue:
			if len(e.State.ActionQueue) > 0 {
				e.enterActionExecutionStage()
				return
			}
		}
	}
	e.setTurnStage(fallback)
	e.clearCombatStage()
	e.clearSubflow()
}

// resumePendingMoraleLoss 恢复被响应中断的士气损失结算。
func (e *GameEngine) resumePendingMoraleLoss(ctx *model.Context) bool {
	if ctx == nil || ctx.Selections == nil {
		return false
	}
	if !runtimeutil.ToBoolContextValue(ctx.Selections["morale_loss_pending"]) {
		return false
	}

	victimID, _ := ctx.Selections["victim_id"].(string)
	victim := e.State.Players[victimID]

	moraleLoss := runtimeutil.ToIntContextValue(ctx.Selections["morale_loss_value"])
	isMagic := runtimeutil.ToBoolContextValue(ctx.Selections["is_magic"])
	fromDamageDraw := runtimeutil.ToBoolContextValue(ctx.Selections["from_damage_draw"])
	overflowMoraleLossFixed := runtimeutil.ToIntContextValue(ctx.Selections["overflow_morale_loss_fixed"])
	discardedCards := discardedCardsFromContext(ctx.Selections["discarded_cards"])

	finalLoss := e.applyMoraleLossAfterTimingWindow(victim, moraleLoss, isMagic, fromDamageDraw, overflowMoraleLossFixed, discardedCards, ctx)
	discardPlayerID, _ := ctx.Selections["discard_player_id"].(string)
	discardPlayer := e.State.Players[discardPlayerID]
	if discardPlayer == nil {
		discardPlayer = victim
	}

	if discardPlayer != nil {
		e.Log(fmt.Sprintf("[System] %s 丢弃了 %d 张牌！士气 -%d", discardPlayer.Name, len(discardedCards), finalLoss))
		e.processAfterDiscardFlowContinuations(discardPlayer.ID, discardedCards, nil)
	}

	ctx.Selections["morale_loss_pending"] = false
	e.checkGameEnd()

	if e.State.PendingInterrupt == nil {
		e.processFlowContinuations(model.FlowContinuationAfterDraw)
		e.restorePhaseAfterDiscardResolution(
			runtimeutil.ToBoolContextValue(ctx.Selections["morale_loss_stay_in_turn"]),
			runtimeutil.ToBoolContextValue(ctx.Selections["morale_loss_is_damage_resolution"]),
			[]discardPhaseCandidate{discardPhaseReturn},
			[]discardPhaseCandidate{discardPhaseReturn, discardPhaseActionQueue, discardPhasePendingDamage},
		)
	}
	return true
}

func discardedCardsFromContext(raw interface{}) []model.Card {
	switch v := raw.(type) {
	case []model.Card:
		return append([]model.Card(nil), v...)
	case []interface{}:
		discardedCards := make([]model.Card, 0, len(v))
		for _, item := range v {
			if c, ok := item.(model.Card); ok {
				discardedCards = append(discardedCards, c)
				continue
			}
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			var c model.Card
			if name, _ := m["name"].(string); name != "" {
				c.Name = name
			}
			if element, _ := m["element"].(string); element != "" {
				c.Element = model.Element(element)
			}
			if c.Name != "" {
				discardedCards = append(discardedCards, c)
			}
		}
		return discardedCards
	default:
		return nil
	}
}
