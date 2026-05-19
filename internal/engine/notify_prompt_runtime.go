// gameflow: 向观察者/前端通知当前 Prompt。

package engine

import (
	"fmt"
	"strings"

	"starcup-engine/internal/engine/core/runtimeutil"
	"starcup-engine/internal/engine/hook/promptfmt"
	"starcup-engine/internal/model"
)

func elementNameForPrompt(raw string) string {
	return promptfmt.ElementName(raw)
}

func (e *GameEngine) decoratePromptForClient(prompt *model.Prompt) *model.Prompt {
	if prompt == nil {
		return nil
	}
	cp := *prompt
	if prompt.Options != nil {
		cp.Options = append([]model.PromptOption{}, prompt.Options...)
	}
	if prompt.SpecialOptions != nil {
		cp.SpecialOptions = append([]model.PromptOption{}, prompt.SpecialOptions...)
	}
	if prompt.CounterTargetIDs != nil {
		cp.CounterTargetIDs = append([]string{}, prompt.CounterTargetIDs...)
	}
	if prompt.EffectHints != nil {
		cp.EffectHints = append([]string{}, prompt.EffectHints...)
	}
	return &cp
}

func (e *GameEngine) Notify(eventType model.GameEventType, msg string, data interface{}) {
	event := model.GameEvent{
		Type:    eventType,
		Message: msg,
	}
	switch eventType {
	case model.EventAskInput:
		switch p := data.(type) {
		case *model.Prompt:
			event.Prompt = e.decoratePromptForClient(p)
		case model.Prompt:
			cp := p
			event.Prompt = e.decoratePromptForClient(&cp)
		}
	case model.EventCardRevealed:
		switch payload := data.(type) {
		case model.CardRevealedPayload:
			cp := payload
			event.CardRevealed = &cp
		case *model.CardRevealedPayload:
			event.CardRevealed = payload
		}
	case model.EventDamageDealt:
		switch payload := data.(type) {
		case model.DamageDealtPayload:
			cp := payload
			event.DamageDealt = &cp
		case *model.DamageDealtPayload:
			event.DamageDealt = payload
		}
	case model.EventActionStep:
		switch payload := data.(type) {
		case model.ActionStepPayload:
			cp := payload
			event.ActionStep = &cp
		case *model.ActionStepPayload:
			event.ActionStep = payload
		}
	case model.EventCombatCue:
		switch payload := data.(type) {
		case model.CombatCuePayload:
			cp := payload
			event.CombatCue = &cp
		case *model.CombatCuePayload:
			event.CombatCue = payload
		}
	case model.EventDrawCards:
		switch payload := data.(type) {
		case model.DrawCardsPayload:
			cp := payload
			event.DrawCards = &cp
		case *model.DrawCardsPayload:
			event.DrawCards = payload
		}
	}
	if err := event.Validate(); err != nil {
		panic(err)
	}
	if e.observer != nil {
		e.observer.OnGameEvent(event)
	}
}

func (e *GameEngine) Log(msg string) {
	e.Notify(model.EventLog, msg, nil)
}

func (e *GameEngine) NotifyCardRevealed(playerID string, cards []model.Card, actionType model.DamageType) {
	e.notifyCards(playerID, cards, actionType, false)
}

func (e *GameEngine) NotifyCardHidden(playerID string, cards []model.Card, actionType model.DamageType) {
	e.notifyCards(playerID, cards, actionType, true)
}

func (e *GameEngine) dispatchCardTiming(player *model.Player, timing model.FlowTiming, targetID string, card model.Card) {
	if e == nil || e.dispatcher == nil || player == nil {
		return
	}
	cardCopy := card
	cardCtx := &model.EventContext{
		Type:     model.EventCardUsed,
		SourceID: player.ID,
		TargetID: targetID,
		Card:     &cardCopy,
	}
	skillCtx := e.BuildContext(player, nil, timing, cardCtx)
	e.dispatcher.OnTiming(skillCtx.Timing, skillCtx)
}

func (e *GameEngine) notifyCards(playerID string, cards []model.Card, actionType model.DamageType, hidden bool) {
	if e.observer == nil || len(cards) == 0 {
		return
	}
	switch actionType {
	case "discard":
		e.addActionDiscard(playerID, len(cards))
	case "defend":
		if p := e.State.Players[playerID]; p != nil {
			cardNames := make([]string, 0, len(cards))
			for _, c := range cards {
				cardNames = append(cardNames, c.Name)
			}
			if len(cardNames) > 0 {
				e.addActionResponse(fmt.Sprintf("%s 防御【%s】", p.Name, strings.Join(cardNames, "、")))
			}
		}
	case "counter":
		if p := e.State.Players[playerID]; p != nil {
			cardNames := make([]string, 0, len(cards))
			for _, c := range cards {
				cardNames = append(cardNames, c.Name)
			}
			if len(cardNames) > 0 {
				e.addActionResponse(fmt.Sprintf("%s 应战【%s】", p.Name, strings.Join(cardNames, "、")))
			}
		}
	}
	p := e.State.Players[playerID]
	if actionType == "discard" && !hidden && !e.suppressSealOnDiscard && p != nil {
		for i := range cards {
			e.dispatchCardTiming(p, model.TimingOnCardPlayedOrRevealed, "", cards[i])
		}
	}
	playerName := playerID
	if p != nil {
		playerName = p.Name
	}
	e.Notify(model.EventCardRevealed, "", model.CardRevealedPayload{
		PlayerID:   playerID,
		PlayerName: playerName,
		Cards:      cards,
		ActionType: string(actionType),
		Hidden:     hidden,
	})
}

func (e *GameEngine) NotifyDamageDealt(sourceID, targetID string, damage int, damageType model.DamageType) {
	if e.observer == nil || damage <= 0 {
		return
	}
	e.addActionDamage(targetID, damage)
	source := e.State.Players[sourceID]
	target := e.State.Players[targetID]
	sourceName := sourceID
	targetName := targetID
	if source != nil {
		sourceName = source.Name
	}
	if target != nil {
		targetName = target.Name
	}
	e.Notify(model.EventDamageDealt, "", model.DamageDealtPayload{
		SourceID:   sourceID,
		SourceName: sourceName,
		TargetID:   targetID,
		TargetName: targetName,
		Damage:     damage,
		DamageType: string(damageType),
	})
}

func (e *GameEngine) NotifyActionStep(line string) {
	if e.observer == nil || line == "" {
		return
	}
	if e.actionSummary != nil && e.actionSummary.active {
		e.addActionNote(line)
		return
	}
	e.Notify(model.EventActionStep, "", model.ActionStepPayload{
		Line: line,
		Kind: "detail",
	})
}

func (e *GameEngine) NotifyActionSummary(line string) {
	if e.observer == nil || line == "" {
		return
	}
	e.Notify(model.EventActionStep, "", model.ActionStepPayload{
		Line: line,
		Kind: "summary",
	})
}

func (e *GameEngine) NotifyCombatCue(attackerID, targetID, phase string) {
	if e.observer == nil || attackerID == "" || targetID == "" || phase == "" {
		return
	}
	e.Notify(model.EventCombatCue, "", model.CombatCuePayload{
		AttackerID: attackerID,
		TargetID:   targetID,
		Phase:      phase,
	})
}

func (e *GameEngine) NotifyDrawCards(playerID string, count int, reason string) {
	if e.observer == nil || playerID == "" || count <= 0 {
		return
	}
	e.addActionDraw(playerID, count)
	p := e.State.Players[playerID]
	playerName := playerID
	if p != nil {
		playerName = p.Name
	}
	e.Notify(model.EventDrawCards, "", model.DrawCardsPayload{
		PlayerID:   playerID,
		PlayerName: playerName,
		DrawCount:  count,
		Reason:     reason,
	})
}

func (e *GameEngine) notifyInterruptPrompt() {
	if e.State.PendingInterrupt == nil {
		return
	}
	if e.State.PendingInterrupt.Type == model.InterruptResponseSkill && e.prunePendingResponseSkills() {
		if err := e.SkipResponse(); err != nil {
			e.Log(fmt.Sprintf("[System] 自动跳过无可用响应失败: %v", err))
		}
		return
	}
	prompt := e.BuildPendingInterruptPrompt()
	if prompt != nil {
		e.Notify(model.EventAskInput, "", prompt)
	} else if e.shouldAutoSkipDiscardDownTo() {
		e.autoSkipPendingDiscardDownTo()
	}
}

// shouldAutoSkipDiscardDownTo 判断当前中断是否为 discard_down_to 且手牌已不超目标，
// 无需弃牌操作可自动跳过。
func (e *GameEngine) shouldAutoSkipDiscardDownTo() bool {
	intr := e.State.PendingInterrupt
	if intr == nil || intr.Type != model.InterruptChoice {
		return false
	}
	data, ok := choiceCtxAsAnyMap(intr.Context)
	if !ok {
		return false
	}
	downTo := runtimeutil.ToIntContextValue(data["discard_down_to"])
	if downTo <= 0 {
		return false
	}
	player := e.State.Players[intr.PlayerID]
	if player == nil {
		return false
	}
	return len(player.Hand) <= downTo
}

// autoSkipPendingDiscardDownTo 自动跳过无需弃牌的中断，执行后续回调以继续游戏流程。
func (e *GameEngine) autoSkipPendingDiscardDownTo() {
	intr := e.State.PendingInterrupt
	data, _ := choiceCtxAsAnyMap(intr.Context)
	choiceType, _ := data["choice_type"].(string)
	player := e.State.Players[intr.PlayerID]
	playerName := intr.PlayerID
	if player != nil && player.Name != "" {
		playerName = player.Name
	}
	e.Log(fmt.Sprintf("[System] %s 手牌未超过目标，自动跳过弃牌: %s", playerName, choiceType))

	afterFn := systemChoiceAfterConsume(choiceType)
	e.PopInterrupt()
	if afterFn != nil {
		afterFn(e, data)
	}
}

func (e *GameEngine) BuildChoicePrompt() *model.Prompt {
	if e.State.PendingInterrupt == nil {
		return nil
	}
	data, ok := choiceCtxAsAnyMap(e.State.PendingInterrupt.Context)
	if !ok {
		return nil
	}
	if e.choiceEngine == nil {
		return nil
	}
	choiceType, _ := data["choice_type"].(string)
	playerID := e.State.PendingInterrupt.PlayerID
	player := e.State.Players[playerID]
	p, err := e.choiceEngine.BuildPrompt(choiceType, playerID, player, data)
	if err != nil {
		e.Log(fmt.Sprintf("[System] BuildChoicePrompt: %v", err))
		return nil
	}
	if p != nil && choiceType != "" && p.ChoiceType == "" {
		p.ChoiceType = choiceType
	}
	return p
}
