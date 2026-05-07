// gameflow: 战斗响应窗口：承伤、防御牌、圣盾等与技能交互。

package engine

import (
	"errors"
	"fmt"
	"strings"

	"starcup-engine/internal/model"
)

// markPendingAttackDamageHitProcessed 将命中后响应结束的攻击伤害标记为已完成 OnAttackHit。
func (e *GameEngine) markPendingAttackDamageHitProcessed(ctx *model.Context) bool {
	if ctx == nil || ctx.EventCtx == nil || len(e.State.PendingDamageQueue) == 0 {
		return false
	}
	for i := range e.State.PendingDamageQueue {
		pd := &e.State.PendingDamageQueue[i]
		if !strings.EqualFold(string(pd.DamageType), string(model.AttackDamage)) {
			continue
		}
		if pd.SourceID != ctx.EventCtx.SourceID || pd.TargetID != ctx.EventCtx.TargetID {
			continue
		}
		pd.AttackHitFlowDispatched = true
		return true
	}
	return false
}

// resumePendingAttackHit 恢复被响应技能选择打断的“攻击命中后续结算”。
func (e *GameEngine) resumePendingAttackHit(ctxData map[string]interface{}) {
	rawCtx, ok := ctxData["user_ctx"].(*model.Context)
	if !ok || rawCtx == nil || !rawCtx.ResumeAttackHitPhase() || rawCtx.EventCtx == nil {
		return
	}
	if e.markPendingAttackDamageHitProcessed(rawCtx) {
		if e.State.PendingInterrupt == nil {
			e.enterDamageResolution(nil)
		}
	}
}

// resumePendingAttackMiss 恢复被响应中断打断的攻击未命中后续流程。
func (e *GameEngine) resumePendingAttackMiss(ctx *model.Context) bool {
	if ctx == nil || ctx.Selections == nil || len(e.State.CombatStack) == 0 {
		return false
	}
	raw := ctx.Selections["attack_miss_resume"]
	data, ok := raw.(map[string]interface{})
	if !ok || data == nil {
		return false
	}
	mode, _ := data["mode"].(string)
	if mode == "" {
		return false
	}
	attackerID, _ := data["attacker_id"].(string)
	targetID, _ := data["target_id"].(string)
	top := e.State.CombatStack[len(e.State.CombatStack)-1]
	if attackerID != "" && top.AttackerID != attackerID {
		return false
	}
	if targetID != "" && top.TargetID != targetID {
		return false
	}

	switch mode {
	case "defend":
		defender := e.State.Players[top.TargetID]
		if defender != nil {
			e.Log(fmt.Sprintf("[Combat] %s 防御成功，攻击未命中", defender.Name))
		}
		e.resolveMagicBowPierceMiss(top.AttackerID, top.TargetID, top.Card, top.IsCounter)
		e.clearCombatStack()
		if !e.routePendingDamageWithReturn(model.TurnStageActionEnd) {
			e.enterActionEndStage()
		}
		return true
	case "shield":
		e.resolveMagicBowPierceMiss(top.AttackerID, top.TargetID, top.Card, top.IsCounter)
		e.clearCombatStack()
		if !e.routePendingDamageWithReturn(model.TurnStageExtraAction) {
			e.enterExtraActionStage()
		}
		return true
	case "counter":
		counterPlayerID, _ := data["counter_player_id"].(string)
		counterTargetID, _ := data["counter_target_id"].(string)
		var counterCard model.Card
		switch v := data["counter_card"].(type) {
		case model.Card:
			counterCard = v
		case *model.Card:
			if v != nil {
				counterCard = *v
			}
		}
		if counterPlayerID == "" || counterTargetID == "" || counterCard.Name == "" {
			return false
		}
		counterPlayer := e.State.Players[counterPlayerID]
		counterTarget := e.State.Players[counterTargetID]
		if counterPlayer != nil && counterTarget != nil {
			e.Log(fmt.Sprintf("[Combat] %s 使用 %s 应战成功！攻击反弹给 %s",
				counterPlayer.Name, counterCard.Name, counterTarget.Name))
		}
		e.resolveMagicBowPierceMiss(top.AttackerID, top.TargetID, top.Card, top.IsCounter)
		e.State.CombatStack = e.State.CombatStack[:len(e.State.CombatStack)-1]
		e.initCombat(counterPlayerID, counterTargetID, &counterCard, false, true, false, nil, "", true)
		if counterPlayer != nil && counterTarget != nil {
			e.Log(fmt.Sprintf("[Combat] %s 应战成功！攻击转移向 %s", counterPlayer.Name, counterTarget.Name))
		}
		return true
	default:
		return false
	}
}

func (e *GameEngine) bindPendingChoiceUserCtxIfMissing(userCtx *model.Context) {
	if e == nil || e.State == nil || e.State.PendingInterrupt == nil || userCtx == nil {
		return
	}
	intr := e.State.PendingInterrupt
	if intr.Type != model.InterruptChoice {
		return
	}
	ctxData, ok := intr.Context.(map[string]interface{})
	if !ok || ctxData == nil {
		return
	}
	if _, exists := ctxData["user_ctx"]; exists {
		return
	}
	ctxData["user_ctx"] = userCtx
	intr.Context = ctxData
}

func (e *GameEngine) HasUsableShieldForCombat(target *model.Player, combatReq model.CombatRequest) bool {
	if target == nil {
		return false
	}
	if combatReq.HasInterceptTag(model.CombatInterceptIgnoreHolyShield) || combatReq.IgnoreShield {
		return false
	}
	for _, fc := range target.Field {
		if fc != nil && fc.Mode == model.FieldEffect && fc.Effect == model.EffectShield {
			return true
		}
	}
	return false
}

func (e *GameEngine) consumeShieldForCombatTake(target *model.Player, combatReq model.CombatRequest) bool {
	if !e.HasUsableShieldForCombat(target, combatReq) {
		return false
	}
	if target == nil {
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

	e.addActionResponse(fmt.Sprintf("%s 的【圣盾】自动抵挡本次攻击", target.Name))
	e.NotifyActionStep(fmt.Sprintf("%s 的【圣盾】触发，自动抵挡了本次攻击", target.Name))
	e.NotifyCombatCue(combatReq.AttackerID, combatReq.TargetID, "shield")
	e.Log(fmt.Sprintf("[Combat] %s 选择承受伤害，触发【圣盾】抵挡本次攻击！", target.Name))

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
	skillCtx := e.BuildContext(e.State.Players[combatReq.AttackerID], e.State.Players[combatReq.TargetID], model.TimingOnHitCheck, missCtx)
	skillCtx.Selections["attack_miss_resume"] = map[string]interface{}{
		"mode":        "shield",
		"attacker_id": combatReq.AttackerID,
		"target_id":   combatReq.TargetID,
	}
	e.dispatcher.OnTiming(skillCtx.Timing, skillCtx)
	if e.State.PendingInterrupt != nil {
		return true
	}

	e.resolveMagicBowPierceMiss(combatReq.AttackerID, combatReq.TargetID, combatReq.Card, combatReq.IsCounter)
	e.clearCombatStack()
	if !e.routePendingDamageWithReturn(model.TurnStageExtraAction) {
		e.enterExtraActionStage()
	}
	return true
}

// HandleCombatResponse 处理战斗交互阶段的响应。
func (e *GameEngine) HandleCombatResponse(act model.PlayerAction) error {
	if len(e.State.CombatStack) == 0 {
		return errors.New("响应时，战斗栈为空")
	}
	if len(act.ExtraArgs) == 0 {
		return errors.New("缺少响应类型")
	}

	respType := act.ExtraArgs[0]
	combatReq := e.State.CombatStack[len(e.State.CombatStack)-1]
	if act.PlayerID != combatReq.TargetID {
		return fmt.Errorf("不是 %s 的响应回合", e.State.Players[combatReq.TargetID].Name)
	}

	player := e.State.Players[act.PlayerID]
	if player == nil {
		return errors.New("玩家不存在")
	}

	handlers := map[string]func() error{
		"take": func() error {
			return e.handleCombatTakeResponse(player, combatReq)
		},
		"defend": func() error {
			return e.handleCombatDefendResponse(act, player, combatReq)
		},
		"counter": func() error {
			return e.handleCombatCounterResponse(act, player, combatReq)
		},
	}
	handler, ok := handlers[respType]
	if !ok {
		return fmt.Errorf("未知的响应类型: %s", respType)
	}
	return handler()
}

func (e *GameEngine) handleCombatTakeResponse(player *model.Player, combatReq model.CombatRequest) error {
	if e.consumeShieldForCombatTake(player, combatReq) {
		return nil
	}

	e.clearCombatStack()
	pd := model.PendingDamage{
		SourceID:      combatReq.AttackerID,
		TargetID:      combatReq.TargetID,
		Damage:        combatReq.Card.Damage,
		DamageType:    model.AttackDamage,
		Card:          combatReq.Card,
		IsCounter:     combatReq.IsCounter,
		IgnoreShield:  combatReq.IgnoreShield,
		InterceptTags: model.CloneCombatInterceptTags(combatReq.InterceptTags),
	}
	e.State.PendingDamageQueue = append([]model.PendingDamage{pd}, e.State.PendingDamageQueue...)
	e.addActionResponse(fmt.Sprintf("%s 承受伤害", player.Name))
	e.NotifyActionStep(fmt.Sprintf("%s承受伤害", model.GetPlayerDisplayName(player)))
	e.NotifyCombatCue(combatReq.AttackerID, combatReq.TargetID, "take")
	e.Log(fmt.Sprintf("[Combat] %s 选择承受伤害，进入伤害结算流程", player.Name))
	e.enterDamageResolution(model.TurnStageActionEnd)
	return nil
}

func (e *GameEngine) handleCombatDefendResponse(act model.PlayerAction, player *model.Player, combatReq model.CombatRequest) error {
	if !e.canUseHolyDefend(&combatReq) {
		return errors.New("本次攻击受【一击无念】影响，不能使用【圣光】防御")
	}
	if res := e.applyTimingOnHitCheckCombatDefendValidation(player, &combatReq); res != nil {
		return res
	}
	card, _, _, ok := e.getPlayableCardByIndex(player, act.CardIndex)
	if !ok {
		return errors.New("无效的卡牌索引")
	}
	if card.Type != model.CardTypeMagic {
		return errors.New("只能使用法术牌进行防御")
	}
	if card.Name == "圣盾" {
		return errors.New("【圣盾】不能在防御时打出，请提前放置到场上触发")
	}
	if card.Name != "圣光" {
		return errors.New("防御只能使用【圣光】；【圣盾】需提前放置到场上")
	}

	e.dispatchCardTiming(player, model.TimingOnCardPlayedOrRevealed, "", card)
	e.NotifyCardRevealed(act.PlayerID, []model.Card{card}, "defend")
	e.NotifyCombatCue(combatReq.AttackerID, combatReq.TargetID, "defend")
	if _, err := e.consumePlayableCardByIndex(player, act.CardIndex); err != nil {
		return err
	}
	e.State.DiscardPile = append(e.State.DiscardPile, card)

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
	skillCtx := e.BuildContext(e.State.Players[combatReq.AttackerID], e.State.Players[combatReq.TargetID], model.TimingOnHitCheck, missCtx)
	skillCtx.Selections["attack_miss_resume"] = map[string]interface{}{
		"mode":        "defend",
		"attacker_id": combatReq.AttackerID,
		"target_id":   combatReq.TargetID,
	}
	e.dispatcher.OnTiming(skillCtx.Timing, skillCtx)
	if e.State.PendingInterrupt != nil {
		e.bindPendingChoiceUserCtxIfMissing(skillCtx)
		return nil
	}

	e.Log(fmt.Sprintf("[Combat] %s 使用 %s 防御成功！", player.Name, card.Name))
	e.resolveMagicBowPierceMiss(combatReq.AttackerID, combatReq.TargetID, combatReq.Card, combatReq.IsCounter)
	e.clearCombatStack()
	if !e.routePendingDamageWithReturn(model.TurnStageActionEnd) {
		e.enterActionEndStage()
	}
	return nil
}

func (e *GameEngine) handleCombatCounterResponse(act model.PlayerAction, player *model.Player, combatReq model.CombatRequest) error {
	if !e.canUseCounter(&combatReq) {
		return errors.New("此攻击无法被应战")
	}

	card, _, _, ok := e.getPlayableCardByIndex(player, act.CardIndex)
	if !ok {
		return errors.New("无效的卡牌索引")
	}
	useSpecialCounterCard, counterCard, err := e.applyTimingOnHitCheckCombatCounterCardPolicy(player, &combatReq, card)
	if err != nil {
		return err
	}
	useFactionCounter := false

	if useSpecialCounterCard {
		card = counterCard
	} else {
		if card.Type != model.CardTypeAttack {
			return errors.New("只能使用攻击牌进行应战（暗影抗拒下可在非自己行动阶段使用【魔弹】）")
		}
		card = e.transformAttackCard(player, card)
		if combatReq.Card.Element == model.ElementDark {
			return errors.New("暗灭无法被应战，只能承受伤害或使用圣光抵挡（场上圣盾会自动生效）")
		}
		if card.Element != combatReq.Card.Element && card.Element != model.ElementDark {
			allowedByPolicy, useFaction := e.applyTimingOnHitCheckCombatCounterElementPolicy(player, &combatReq, card)
			if allowedByPolicy {
				useFactionCounter = useFaction
			} else {
				return fmt.Errorf("应战必须使用同系攻击牌或暗灭，对方为 %s 系", combatReq.Card.Element)
			}
		}
	}

	targetID := act.TargetID
	if targetID == "" {
		return errors.New("应战必须指定反弹目标（从攻击方队友中选择）")
	}
	if targetID == combatReq.AttackerID {
		return errors.New("不能选择攻击者本人，只能选择攻击方的队友进行反弹")
	}

	target := e.State.Players[targetID]
	if target == nil {
		return errors.New("目标不存在")
	}
	attacker := e.State.Players[combatReq.AttackerID]
	if attacker == nil {
		return errors.New("攻击者信息异常")
	}
	if target.Camp != attacker.Camp {
		return errors.New("应战反弹目标必须是攻击方的队友")
	}

	e.dispatchCardTiming(player, model.TimingOnCardPlayedOrRevealed, "", card)
	e.NotifyCardRevealed(act.PlayerID, []model.Card{card}, "counter")
	e.NotifyCombatCue(combatReq.AttackerID, combatReq.TargetID, "counter")
	if _, err := e.consumePlayableCardByIndex(player, act.CardIndex); err != nil {
		return err
	}
	e.State.DiscardPile = append(e.State.DiscardPile, card)
	e.applyTimingOnHitCheckCombatCounterResolvePolicy(player, &combatReq, &card, useFactionCounter)

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
	skillCtx := e.BuildContext(e.State.Players[combatReq.AttackerID], e.State.Players[combatReq.TargetID], model.TimingOnHitCheck, missCtx)
	skillCtx.Selections["attack_miss_resume"] = map[string]interface{}{
		"mode":              "counter",
		"attacker_id":       combatReq.AttackerID,
		"target_id":         combatReq.TargetID,
		"counter_player_id": act.PlayerID,
		"counter_target_id": targetID,
		"counter_card":      card,
	}
	e.dispatcher.OnTiming(skillCtx.Timing, skillCtx)
	if e.State.PendingInterrupt != nil {
		return nil
	}

	e.Log(fmt.Sprintf("[Combat] %s 使用 %s 应战成功！攻击反弹给 %s", player.Name, card.Name, target.Name))
	e.resolveMagicBowPierceMiss(combatReq.AttackerID, combatReq.TargetID, combatReq.Card, combatReq.IsCounter)
	e.State.CombatStack = e.State.CombatStack[:len(e.State.CombatStack)-1]
	e.initCombat(act.PlayerID, targetID, &card, false, true, false, nil, "", true)
	e.Log(fmt.Sprintf("[Combat] %s 应战成功！攻击转移向 %s", player.Name, target.Name))
	if len(e.State.PendingDamageQueue) > 0 {
		e.enterDamageResolution(model.CombatStageHitCheck)
	}
	return nil
}
