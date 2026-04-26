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

	if handled, err := e.handleDiscardSelectionFollowups(player, data); handled || err != nil {
		return err
	}

	e.PopInterrupt()
	if e.State.PendingInterrupt == nil {
		if hasChoiceResumePoint(data["draw_resume_phase"]) && !e.hasReturnPoint() {
			e.setReturnPoint(data["draw_resume_phase"])
		}
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

	if (fromDamageDraw || isDamageResolution) && playerpkg.IsCharacter(victim, "crimson_knight") && playerpkg.HasForm(victim, model.FormCrimsonKnightHotBlooded) {
		moraleLoss = 0
	}

	isMagic := runtimeutil.ToBoolContextValue(data["is_magic"])
	moraleLoss = e.capMoraleLoss(victim.Camp, moraleLoss)
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

func (e *GameEngine) handleDiscardSelectionFollowups(player *model.Player, data map[string]interface{}) (bool, error) {
	if player == nil {
		return false, fmt.Errorf("玩家不存在")
	}

	if demonEyeUserID, _ := data["mb_demon_eye_user_id"].(string); demonEyeUserID != "" {
		user := e.State.Players[demonEyeUserID]
		if user == nil {
			return false, fmt.Errorf("魔眼施法者不存在")
		}
		if len(user.Hand) == 0 {
			maxEnergy := e.getPlayerEnergyCap(user)
			if user.Gem+user.Crystal < maxEnergy {
				user.Crystal++
				if user.Gem+user.Crystal > maxEnergy {
					user.Crystal -= user.Gem + user.Crystal - maxEnergy
				}
			}
			e.Log(fmt.Sprintf("%s 的 [魔眼] 生效：已完成目标弃牌，但自己无手牌可充能，改为仅获得1点蓝水晶", user.Name))
			e.PopInterrupt()
			if e.State.PendingInterrupt == nil {
				e.setTurnStage(model.TurnStageActionStart)
			}
			return true, nil
		}
		e.State.PendingInterrupt.Type = model.InterruptChoice
		e.State.PendingInterrupt.PlayerID = demonEyeUserID
		e.State.PendingInterrupt.Context = map[string]interface{}{
			"choice_type":       "mb_demon_eye_charge_card",
			"user_id":           demonEyeUserID,
			"need_count":        1,
			"selected_indices":  []int{},
			"remaining_indices": allHandIndices(user),
		}
		e.syncGamePhaseWithInterrupt(e.State.PendingInterrupt)
		e.Log(fmt.Sprintf("%s 的 [魔眼] 生效：%s 已弃置1张手牌，请选择1张手牌作为充能", user.Name, player.Name))
		e.notifyInterruptPrompt()
		return true, nil
	}

	if playerpkg.IsCharacter(player, "magic_lancer") && player.TurnState.SkillFlowState != nil && player.TurnState.SkillFlowState["ml_stardust_wait_discard"] > 0 {
		for _, entry := range roleRegistry.Entries() {
			if entry.AfterDiscardFollowup != nil {
				entry.AfterDiscardFollowup(newRoleChoiceRuntime(e), player)
			}
		}
	}
	return false, nil
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
	mbChargeResume := runtimeutil.ToBoolContextValue(ctx.Selections["mb_charge_resume"])
	discardPlayerID, _ := ctx.Selections["discard_player_id"].(string)
	discardPlayer := e.State.Players[discardPlayerID]
	if discardPlayer == nil {
		discardPlayer = victim
	}

	if mbChargeResume {
		if victim != nil {
			e.Log(fmt.Sprintf("%s 的 [充能] 爆士气结算完成：士气-%d（本次不弃牌）", victim.Name, finalLoss))
		}
	} else if discardPlayer != nil {
		e.Log(fmt.Sprintf("[System] %s 丢弃了 %d 张牌！士气 -%d", discardPlayer.Name, len(discardedCards), finalLoss))
		if playerpkg.IsCharacter(discardPlayer, "magic_lancer") && discardPlayer.TurnState.SkillFlowState != nil && discardPlayer.TurnState.SkillFlowState["ml_stardust_wait_discard"] > 0 {
			for _, entry := range roleRegistry.Entries() {
				if entry.AfterDiscardFollowup != nil {
					entry.AfterDiscardFollowup(newRoleChoiceRuntime(e), discardPlayer)
				}
			}
		}
	}

	ctx.Selections["morale_loss_pending"] = false
	e.checkGameEnd()

	if mbChargeResume {
		if e.State.GameOver {
			return true
		}
		maxPlace := runtimeutil.ToIntContextValue(ctx.Selections["mb_charge_max_place"])
		userID, _ := ctx.Selections["mb_charge_user_id"].(string)
		user := e.State.Players[userID]
		if e.State.PendingInterrupt == nil {
			if user != nil && maxPlace > 0 {
				e.PushInterrupt(&model.Interrupt{
					Type:     model.InterruptChoice,
					PlayerID: user.ID,
					Context: map[string]interface{}{
						"choice_type": "mb_charge_place_count",
						"user_id":     user.ID,
						"max_place":   maxPlace,
					},
				})
			} else {
				e.setTurnStage(model.TurnStageActionStart)
			}
		}
		return true
	}

	if e.State.PendingInterrupt == nil {
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
