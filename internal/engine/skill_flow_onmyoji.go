// gameflow: 阴阳师：黑暗祭礼、阴阳转换、式神咒束等辅助函数。

package engine

import (
	"fmt"
	"starcup-engine/internal/model"
)

func (e *GameEngine) maybeOnmyojiDarkRitual(player *model.Player) bool {
	if player == nil || !e.isOnmyoji(player) || player.Tokens == nil || player.Tokens["onmyoji_ghost_fire"] < 3 {
		return false
	}
	targetIDs := e.campEnemyIDs(player.Camp)
	if len(targetIDs) == 0 {
		return false
	}
	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: player.ID,
		Context: map[string]interface{}{
			"choice_type": "onmyoji_dark_ritual_target",
			"user_id":     player.ID,
			"target_ids":  targetIDs,
			"ghost_fire":  player.Tokens["onmyoji_ghost_fire"],
		},
	})
	e.Log(fmt.Sprintf("%s 的 [黑暗祭礼] 触发，等待选择2点法术伤害目标", player.Name))
	return true
}

func (e *GameEngine) applyOnmyojiFactionCounterBonuses(actor *model.Player, card *model.Card) {
	if actor == nil || card == nil {
		return
	}
	beforePoses := e.snapshotPlayerPoses()
	if actor.Tokens == nil {
		actor.Tokens = map[string]int{}
	}
	actor.Tokens["onmyoji_ghost_fire"]++
	if actor.Tokens["onmyoji_ghost_fire"] > 3 {
		actor.Tokens["onmyoji_ghost_fire"] = 3
	}
	e.Log(fmt.Sprintf("%s 的 [阴阳转换] 触发，鬼火+1", actor.Name))
	if hasOnmyojiShikigamiForm(actor) {
		e.DrawCards(actor.ID, 1)
		actor.Tokens["onmyoji_ghost_fire"]++
		if actor.Tokens["onmyoji_ghost_fire"] > 3 {
			actor.Tokens["onmyoji_ghost_fire"] = 3
		}
		leaveOnmyojiShikigamiForm(actor)
		e.Log(fmt.Sprintf("%s 的 [式神转换] 触发：摸1并鬼火+1，然后脱离式神形态", actor.Name))
	}
	card.Damage = actor.Tokens["onmyoji_ghost_fire"]
	if card.Damage < 0 {
		card.Damage = 0
	}
	e.dispatchOrientationChanges(beforePoses)
}

// tryStartOnmyojiBindingInterrupt 检查并触发"式神咒束"代应战确认。
// 返回 true 表示已进入中断等待（应暂停当前 Drive）。
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
	if e.isOnmyoji(target) {
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
		if !e.isOnmyoji(actor) || actor.Camp != target.Camp {
			continue
		}
		if !hasOnmyojiShikigamiForm(actor) {
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

// tryStartOnmyojiYinYangInterrupt 检查并触发"阴阳转换"优先确认。
// 规则：目标阴阳师若手里有"与来袭攻击同命格"的攻击牌，则先询问是否发动；
// 若选择不发动，才进入常规 承受/防御/应战 弹框。
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
	if target == nil || attacker == nil || !e.isOnmyoji(target) {
		return false
	}
	if !onmyojiCanUseFactionCounter(combatReq.Card) {
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

// executeOnmyojiBindingCounter 在战斗阶段自动执行已确认的"式神咒束应战"。
// 返回 true 表示已推进流程（可能进入中断），当前 Drive 应暂停。
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
		canCounter = onmyojiCanUseFactionCounter(combatReq.Card) &&
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