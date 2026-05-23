// gameflow: 战斗响应窗口：承伤、防御牌、圣盾等与技能交互。

package engine

import (
	"errors"
	"fmt"
	"strings"

	"starcup-engine/internal/model"
)

type attackMissResumeMode string

const (
	attackMissResumeDefend  attackMissResumeMode = "defend"
	attackMissResumeShield  attackMissResumeMode = "shield"
	attackMissResumeCounter attackMissResumeMode = "counter"
)

type attackMissResumeState struct {
	Mode            attackMissResumeMode
	AttackerID      string
	TargetID        string
	CounterPlayerID string
	CounterTargetID string
	CounterCard     *model.Card
}

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
	resume, ok := attackMissResumeFromContext(ctx)
	if !ok || resume.Mode == "" {
		return false
	}
	top := e.State.CombatStack[len(e.State.CombatStack)-1]
	if resume.AttackerID != "" && top.AttackerID != resume.AttackerID {
		return false
	}
	if resume.TargetID != "" && top.TargetID != resume.TargetID {
		return false
	}

	switch resume.Mode {
	case attackMissResumeDefend:
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
	case attackMissResumeShield:
		e.resolveMagicBowPierceMiss(top.AttackerID, top.TargetID, top.Card, top.IsCounter)
		e.clearCombatStack()
		if !e.routePendingDamageWithReturn(model.TurnStageExtraAction) {
			e.enterExtraActionStage()
		}
		return true
	case attackMissResumeCounter:
		if resume.CounterPlayerID == "" || resume.CounterTargetID == "" || resume.CounterCard == nil || resume.CounterCard.Name == "" {
			return false
		}
		counterPlayer := e.State.Players[resume.CounterPlayerID]
		counterTarget := e.State.Players[resume.CounterTargetID]
		if counterPlayer != nil && counterTarget != nil {
			e.Log(fmt.Sprintf("[Combat] %s 使用 %s 应战成功！攻击反弹给 %s",
				counterPlayer.Name, resume.CounterCard.Name, counterTarget.Name))
		}
		e.resolveMagicBowPierceMiss(top.AttackerID, top.TargetID, top.Card, top.IsCounter)
		e.State.CombatStack = e.State.CombatStack[:len(e.State.CombatStack)-1]
		e.initCombat(resume.CounterPlayerID, resume.CounterTargetID, resume.CounterCard, false, true, false, nil, "", true)
		if counterPlayer != nil && counterTarget != nil {
			e.Log(fmt.Sprintf("[Combat] %s 应战成功！攻击转移向 %s", counterPlayer.Name, counterTarget.Name))
		}
		return true
	default:
		return false
	}
}

func (e *GameEngine) resolveCounterAttackAfterAttackMissTiming(counterPlayerID, counterTargetID string, counterCard model.Card) bool {
	if e == nil || e.State == nil || len(e.State.CombatStack) == 0 {
		return false
	}
	combatReq := e.State.CombatStack[len(e.State.CombatStack)-1]
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
	counterCardForResume := counterCard
	result := e.dispatchAttackRulebookEventTimingWithMarkers(model.TimingAttackMiss, e.State.Players[combatReq.AttackerID], e.State.Players[combatReq.TargetID], missCtx, attackKindFromCounter(combatReq.IsCounter), map[string]any{
		"attack_miss_resume": attackMissResumeState{
			Mode:            attackMissResumeCounter,
			AttackerID:      combatReq.AttackerID,
			TargetID:        combatReq.TargetID,
			CounterPlayerID: counterPlayerID,
			CounterTargetID: counterTargetID,
			CounterCard:     &counterCardForResume,
		},
	})
	if result.Interrupted {
		return true
	}

	counterPlayer := e.State.Players[counterPlayerID]
	counterTarget := e.State.Players[counterTargetID]
	if counterPlayer != nil && counterTarget != nil {
		e.Log(fmt.Sprintf("[Combat] %s 使用 %s 应战成功！攻击反弹给 %s", counterPlayer.Name, counterCard.Name, counterTarget.Name))
	}
	e.resolveMagicBowPierceMiss(combatReq.AttackerID, combatReq.TargetID, combatReq.Card, combatReq.IsCounter)
	e.State.CombatStack = e.State.CombatStack[:len(e.State.CombatStack)-1]
	e.initCombat(counterPlayerID, counterTargetID, &counterCard, false, true, false, nil, "", true)
	if counterPlayer != nil && counterTarget != nil {
		e.Log(fmt.Sprintf("[Combat] %s 应战成功！攻击转移向 %s", counterPlayer.Name, counterTarget.Name))
	}
	return true
}

func attackMissResumeFromContext(ctx *model.Context) (attackMissResumeState, bool) {
	if ctx == nil || ctx.Selections == nil {
		return attackMissResumeState{}, false
	}
	switch data := ctx.Selections["attack_miss_resume"].(type) {
	case attackMissResumeState:
		return data, true
	case *attackMissResumeState:
		if data == nil {
			return attackMissResumeState{}, false
		}
		return *data, true
	case map[string]interface{}:
		return legacyAttackMissResumeFromMap(data)
	default:
		return attackMissResumeState{}, false
	}
}

func legacyAttackMissResumeFromMap(data map[string]interface{}) (attackMissResumeState, bool) {
	if data == nil {
		return attackMissResumeState{}, false
	}
	mode, _ := data["mode"].(string)
	resume := attackMissResumeState{
		Mode:            attackMissResumeMode(mode),
		AttackerID:      stringContextValue(data["attacker_id"]),
		TargetID:        stringContextValue(data["target_id"]),
		CounterPlayerID: stringContextValue(data["counter_player_id"]),
		CounterTargetID: stringContextValue(data["counter_target_id"]),
	}
	switch v := data["counter_card"].(type) {
	case model.Card:
		card := v
		resume.CounterCard = &card
	case *model.Card:
		resume.CounterCard = v
	}
	return resume, resume.Mode != ""
}

func stringContextValue(value any) string {
	s, _ := value.(string)
	return s
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
	result := e.dispatchAttackRulebookEventTimingWithMarkers(model.TimingAttackMiss, e.State.Players[combatReq.AttackerID], e.State.Players[combatReq.TargetID], missCtx, attackKindFromCounter(combatReq.IsCounter), map[string]any{
		"attack_miss_resume": attackMissResumeState{
			Mode:       attackMissResumeShield,
			AttackerID: combatReq.AttackerID,
			TargetID:   combatReq.TargetID,
		},
	})
	if result.Interrupted {
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
	if res := e.applyAttackResponseDefendValidation(player, &combatReq); res != nil {
		return res
	}
	card, ok := e.cardForPlayerAction(player, act)
	if !ok {
		return errors.New("无效的卡牌ID")
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

	e.dispatchCardTiming(player, model.TimingCardPlayedRevealed, "", card)
	e.NotifyCardRevealed(act.PlayerID, []model.Card{card}, "defend")
	e.NotifyCombatCue(combatReq.AttackerID, combatReq.TargetID, "defend")
	if _, err := e.consumePlayableCardByID(player, act.CardID); err != nil {
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
	result := e.dispatchAttackRulebookEventTimingWithMarkers(model.TimingAttackMiss, e.State.Players[combatReq.AttackerID], e.State.Players[combatReq.TargetID], missCtx, attackKindFromCounter(combatReq.IsCounter), map[string]any{
		"attack_miss_resume": attackMissResumeState{
			Mode:       attackMissResumeDefend,
			AttackerID: combatReq.AttackerID,
			TargetID:   combatReq.TargetID,
		},
	})
	if result.Interrupted {
		e.bindPendingChoiceUserCtxIfMissing(result.Context)
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

	card, ok := e.cardForPlayerAction(player, act)
	if !ok {
		return errors.New("无效的卡牌ID")
	}
	useSpecialCounterCard, counterCard, err := e.applyAttackResponseCounterCardPolicy(player, &combatReq, card)
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
			allowedByPolicy, useFaction := e.applyAttackResponseCounterElementPolicy(player, &combatReq, card)
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

	counterInfo := &model.AttackEventInfo{
		ActionType:       string(model.ActionAttack),
		CanBeResponded:   true,
		CounterInitiator: player.ID,
		Element:          string(card.Element),
		InterceptTags:    map[model.CombatInterceptTag]bool{},
	}
	if e.dispatchAttackRulebookTiming(model.TimingAttackDeclare, player, target, &card, counterInfo, model.AttackKindCounter) {
		return nil
	}
	if e.dispatchAttackRulebookTiming(model.TimingAttackSelectTarget, player, target, &card, counterInfo, model.AttackKindCounter) {
		return nil
	}
	if e.dispatchAttackRulebookTiming(model.TimingAttackPlayCard, player, target, &card, counterInfo, model.AttackKindCounter) {
		return nil
	}
	if e.dispatchAttackRulebookTiming(model.TimingAttackModifyCard, player, target, &card, counterInfo, model.AttackKindCounter) {
		return nil
	}
	if e.dispatchAttackRulebookTiming(model.TimingAttackCommitted, player, target, &card, counterInfo, model.AttackKindCounter) {
		return nil
	}

	e.dispatchCardTiming(player, model.TimingCardPlayedRevealed, "", card)
	e.NotifyCardRevealed(act.PlayerID, []model.Card{card}, "counter")
	e.NotifyCombatCue(combatReq.AttackerID, combatReq.TargetID, "counter")
	if _, err := e.consumePlayableCardByID(player, act.CardID); err != nil {
		return err
	}
	e.State.DiscardPile = append(e.State.DiscardPile, card)
	e.applyAttackResponseCounterResolvePolicy(player, &combatReq, &card, useFactionCounter)

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
	counterCardForResume := card
	result := e.dispatchAttackRulebookEventTimingWithMarkers(model.TimingAttackMiss, e.State.Players[combatReq.AttackerID], e.State.Players[combatReq.TargetID], missCtx, attackKindFromCounter(combatReq.IsCounter), map[string]any{
		"attack_miss_resume": attackMissResumeState{
			Mode:            attackMissResumeCounter,
			AttackerID:      combatReq.AttackerID,
			TargetID:        combatReq.TargetID,
			CounterPlayerID: act.PlayerID,
			CounterTargetID: targetID,
			CounterCard:     &counterCardForResume,
		},
	})
	if result.Interrupted {
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
