// gameflow: 运行时策略装配：combat rules。

package engine

import (
	"fmt"

	playerpkg "starcup-engine/internal/engine/player"
	magicswordsman "starcup-engine/internal/engine/player/magic_swordsman"
	onmyojiplayer "starcup-engine/internal/engine/player/onmyoji"
	"starcup-engine/internal/model"
)

func combatInteractionOnmyojiBindingInterruptHook(e *GameEngine, req *model.CombatRequest) bool {
	return e != nil && req != nil && e.tryStartOnmyojiBindingInterrupt(req)
}

func combatInteractionOnmyojiBindingCounterHook(e *GameEngine, req *model.CombatRequest) bool {
	return e != nil && req != nil && e.executeOnmyojiBindingCounter(req)
}

func combatInteractionOnmyojiYinYangInterruptHook(e *GameEngine, req *model.CombatRequest) bool {
	return e != nil && req != nil && e.tryStartOnmyojiYinYangInterrupt(req)
}

func combatInteractionDarkElementResponsePolicyHook(e *GameEngine, req *model.CombatRequest) bool {
	if e == nil || req == nil || req.Card == nil || req.Card.Element != model.ElementDark {
		return false
	}
	target := e.State.Players[req.TargetID]
	currentTurnPlayerID := ""
	if len(e.State.PlayerOrder) > 0 && e.State.CurrentTurn >= 0 && e.State.CurrentTurn < len(e.State.PlayerOrder) {
		currentTurnPlayerID = e.State.PlayerOrder[e.State.CurrentTurn]
	}
	if target == nil || magicswordsman.CanUseShadowRejectResponse(target, currentTurnPlayerID) {
		return false
	}
	req.SetInterceptTag(model.CombatInterceptUnrespondable)
	return false
}

func combatDefendMagicLancerDarkBindPolicy(e *GameEngine, player *model.Player, req *model.CombatRequest) error {
	if playerpkg.IsCharacter(player, "magic_lancer") {
		return fmt.Errorf("魔枪受[黑暗束缚]影响，不能使用法术牌防御")
	}
	return nil
}

func combatCounterShadowRejectMagicBulletPolicy(e *GameEngine, player *model.Player, req *model.CombatRequest, card model.Card) (bool, model.Card, error) {
	currentTurnPlayerID2 := ""
	if len(e.State.PlayerOrder) > 0 && e.State.CurrentTurn >= 0 && e.State.CurrentTurn < len(e.State.PlayerOrder) {
		currentTurnPlayerID2 = e.State.PlayerOrder[e.State.CurrentTurn]
	}
	if !magicswordsman.CanUseShadowRejectResponse(player, currentTurnPlayerID2) || card.Type != model.CardTypeMagic || card.Name != "魔弹" {
		return false, card, nil
	}
	e.Log(fmt.Sprintf("[Combat] %s 触发[暗影抗拒]：非自己行动阶段使用【魔弹】应战", player.Name))
	return true, card, nil
}

func combatCounterOnmyojiFactionElementPolicy(e *GameEngine, player *model.Player, req *model.CombatRequest, counterCard model.Card) (bool, bool) {
	if req.Card == nil {
		return false, false
	}
	if !playerpkg.IsCharacter(player, "onmyoji") || !onmyojiplayer.CanUseFactionCounter(req.Card) {
		return false, false
	}
	if counterCard.Faction == "" || counterCard.Faction != req.Card.Faction {
		return false, false
	}
	return true, true
}

func combatCounterOnmyojiFactionResolvePolicy(e *GameEngine, player *model.Player, req *model.CombatRequest, counterCard *model.Card, useFaction bool) {
	if !useFaction {
		return
	}
	e.applyOnmyojiFactionCounterBonuses(player, counterCard)
}

func magicMissileDefendMagicLancerDarkBindPolicy(e *GameEngine, player *model.Player, chain *model.MagicBulletChain) error {
	if playerpkg.IsCharacter(player, "magic_lancer") {
		return fmt.Errorf("魔枪受[黑暗束缚]影响，不能使用法术牌防御")
	}
	return nil
}

func magicMissileCounterMagicLancerDarkBindPolicy(e *GameEngine, player *model.Player, chain *model.MagicBulletChain, card model.Card) error {
	if playerpkg.IsCharacter(player, "magic_lancer") {
		return fmt.Errorf("魔枪受[黑暗束缚]影响，不能使用法术牌")
	}
	return nil
}

// ---- 阴阳师战斗交互 ----

func (e *GameEngine) maybeOnmyojiDarkRitual(player *model.Player) bool {
	return onmyojiplayer.MaybeDarkRitual(newRoleChoiceRuntime(e), player)
}

func (e *GameEngine) applyOnmyojiFactionCounterBonuses(actor *model.Player, card *model.Card) {
	onmyojiplayer.ApplyFactionCounterBonuses(newRoleChoiceRuntime(e), actor, card)
}

func (e *GameEngine) canPayOnmyojiBindingCost(camp model.Camp) bool {
	return onmyojiplayer.CanPayBindingCost(newRoleChoiceRuntime(e), camp)
}

func collectOnmyojiCounterOptions(p *model.Player, incoming *model.Card) []map[string]interface{} {
	return onmyojiplayer.CollectCounterOptions(p, incoming)
}

func (e *GameEngine) tryStartOnmyojiBindingInterrupt(combatReq *model.CombatRequest) bool {
	if combatReq == nil {
		return false
	}
	if combatReq.OnmyojiBindingChecked {
		return false
	}
	combatReq.OnmyojiBindingChecked = true

	if combatReq.IsCounter || combatReq.IsForcedHit || !combatReq.CanBeResponded || combatReq.Card == nil {
		return false
	}
	if combatReq.Card.Element == model.ElementDark {
		return false
	}
	target := e.State.Players[combatReq.TargetID]
	attacker := e.State.Players[combatReq.AttackerID]
	if target == nil || attacker == nil {
		return false
	}
	if attacker.Camp == target.Camp {
		return false
	}
	if playerpkg.IsCharacter(target, "onmyoji") {
		return false
	}

	var counterTargetIDs []string
	for _, pid := range e.State.PlayerOrder {
		if pid == attacker.ID {
			continue
		}
		player := e.State.Players[pid]
		if player == nil || player.Camp != attacker.Camp {
			continue
		}
		counterTargetIDs = append(counterTargetIDs, pid)
	}
	if len(counterTargetIDs) == 0 {
		return false
	}

	for _, pid := range e.State.PlayerOrder {
		actor := e.State.Players[pid]
		if actor == nil || actor.ID == target.ID {
			continue
		}
		if !playerpkg.IsCharacter(actor, "onmyoji") || actor.Camp != target.Camp {
			continue
		}
		if !playerpkg.HasForm(actor, model.FormOnmyojiShikigami) {
			continue
		}
		if !e.canPayOnmyojiBindingCost(actor.Camp) {
			continue
		}
		cardOptions := collectOnmyojiCounterOptions(actor, combatReq.Card)
		if len(cardOptions) == 0 {
			continue
		}
		e.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: actor.ID,
			Context: map[string]interface{}{
				"choice_type":        "onmyoji_binding_confirm",
				"actor_id":           actor.ID,
				"attacker_id":        combatReq.AttackerID,
				"target_id":          combatReq.TargetID,
				"card_options":       cardOptions,
				"counter_target_ids": counterTargetIDs,
			},
		})
		e.Log(fmt.Sprintf("%s 可发动 [式神咒束] 代应战，等待其确认", actor.Name))
		return true
	}
	return false
}

func (e *GameEngine) tryStartOnmyojiYinYangInterrupt(combatReq *model.CombatRequest) bool {
	if combatReq == nil {
		return false
	}
	if combatReq.OnmyojiYinYangChecked {
		return false
	}
	combatReq.OnmyojiYinYangChecked = true

	if combatReq.IsCounter || combatReq.IsForcedHit || !combatReq.CanBeResponded || combatReq.Card == nil {
		return false
	}
	if combatReq.Card.Element == model.ElementDark {
		return false
	}

	target := e.State.Players[combatReq.TargetID]
	attacker := e.State.Players[combatReq.AttackerID]
	if target == nil || attacker == nil || !playerpkg.IsCharacter(target, "onmyoji") {
		return false
	}
	if !onmyojiplayer.CanUseFactionCounter(combatReq.Card) {
		return false
	}

	allOptions := collectOnmyojiCounterOptions(target, combatReq.Card)
	factionOptions := make([]map[string]interface{}, 0, len(allOptions))
	for _, option := range allOptions {
		useFaction, _ := option["use_faction"].(bool)
		if useFaction {
			factionOptions = append(factionOptions, option)
		}
	}
	if len(factionOptions) == 0 {
		return false
	}

	var counterTargetIDs []string
	for _, pid := range e.State.PlayerOrder {
		if pid == attacker.ID {
			continue
		}
		player := e.State.Players[pid]
		if player == nil || player.Camp != attacker.Camp {
			continue
		}
		counterTargetIDs = append(counterTargetIDs, pid)
	}
	if len(counterTargetIDs) == 0 {
		return false
	}

	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: target.ID,
		Context: map[string]interface{}{
			"choice_type":        "onmyoji_yinyang_confirm",
			"actor_id":           target.ID,
			"attacker_id":        combatReq.AttackerID,
			"target_id":          combatReq.TargetID,
			"card_options":       factionOptions,
			"counter_target_ids": counterTargetIDs,
		},
	})
	e.Log(fmt.Sprintf("%s 可发动 [阴阳转换]，等待其确认", target.Name))
	return true
}

func (e *GameEngine) executeOnmyojiBindingCounter(combatReq *model.CombatRequest) bool {
	if combatReq == nil {
		return false
	}
	actorID := combatReq.OnmyojiBindingActorID
	cardID := combatReq.OnmyojiBindingCounterID
	counterTargetID := combatReq.OnmyojiBindingTargetID
	if actorID == "" || cardID == "" || counterTargetID == "" {
		return false
	}
	actor := e.State.Players[actorID]
	if actor == nil || combatReq.Card == nil {
		combatReq.OnmyojiBindingActorID = ""
		combatReq.OnmyojiBindingCounterID = ""
		combatReq.OnmyojiBindingTargetID = ""
		combatReq.OnmyojiBindingUseFaction = false
		return false
	}
	cardIdx := findPlayableCardIndexByID(actor, cardID)
	card, _, _, ok := getPlayableCardByIndex(actor, cardIdx)
	if !ok || card.Type != model.CardTypeAttack {
		combatReq.OnmyojiBindingActorID = ""
		combatReq.OnmyojiBindingCounterID = ""
		combatReq.OnmyojiBindingTargetID = ""
		combatReq.OnmyojiBindingUseFaction = false
		return false
	}
	useFaction := combatReq.OnmyojiBindingUseFaction
	canCounter := card.Element == combatReq.Card.Element || card.Element == model.ElementDark
	if !canCounter && useFaction {
		canCounter = onmyojiplayer.CanUseFactionCounter(combatReq.Card) &&
			card.Faction != "" && card.Faction == combatReq.Card.Faction
	}
	if !canCounter {
		combatReq.OnmyojiBindingActorID = ""
		combatReq.OnmyojiBindingCounterID = ""
		combatReq.OnmyojiBindingTargetID = ""
		combatReq.OnmyojiBindingUseFaction = false
		return false
	}

	e.NotifyCardRevealed(actor.ID, []model.Card{card}, "counter")
	e.NotifyCombatCue(combatReq.AttackerID, combatReq.TargetID, "counter")
	if _, err := consumePlayableCardByIndex(actor, cardIdx); err != nil {
		combatReq.OnmyojiBindingActorID = ""
		combatReq.OnmyojiBindingCounterID = ""
		combatReq.OnmyojiBindingTargetID = ""
		combatReq.OnmyojiBindingUseFaction = false
		return false
	}
	e.State.DiscardPile = append(e.State.DiscardPile, card)

	if useFaction {
		e.applyOnmyojiFactionCounterBonuses(actor, &card)
	}

	missCtx := &model.EventContext{
		Type:     model.EventAttack,
		SourceID: combatReq.AttackerID,
		TargetID: combatReq.TargetID,
		Card:     combatReq.Card,
		AttackInfo: &model.AttackEventInfo{
			ActionType: string(model.ActionAttack),
			CounterInitiator: func() string {
				if combatReq.IsCounter {
					return combatReq.AttackerID
				}
				return ""
			}(),
		},
	}
	skillCtx := e.buildContext(e.State.Players[combatReq.AttackerID], e.State.Players[combatReq.TargetID], model.TimingOnHitCheck, missCtx)
	skillCtx.Selections["attack_miss_resume"] = map[string]interface{}{
		"mode":              "counter",
		"attacker_id":       combatReq.AttackerID,
		"target_id":         combatReq.TargetID,
		"counter_player_id": actor.ID,
		"counter_target_id": counterTargetID,
		"counter_card":      card,
	}
	e.dispatcher.OnTiming(skillCtx.Timing, skillCtx)
	if e.State.PendingInterrupt != nil {
		return true
	}

	e.resolveMagicBowPierceMiss(combatReq.AttackerID, combatReq.TargetID, combatReq.Card, combatReq.IsCounter)
	e.Log(fmt.Sprintf("[Combat] %s 通过[式神咒束]代应战成功，攻击反弹给 %s", actor.Name, model.GetPlayerDisplayName(e.State.Players[counterTargetID])))
	e.State.CombatStack = e.State.CombatStack[:len(e.State.CombatStack)-1]
	e.initCombat(actor.ID, counterTargetID, &card, false, true, false, nil, true)
	combatReq.OnmyojiBindingActorID = ""
	combatReq.OnmyojiBindingCounterID = ""
	combatReq.OnmyojiBindingTargetID = ""
	combatReq.OnmyojiBindingUseFaction = false
	return true
}
