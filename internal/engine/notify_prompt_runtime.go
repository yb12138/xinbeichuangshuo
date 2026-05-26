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

func (e *GameEngine) emitGameEvent(event model.GameEvent) {
	if event.Narrative == nil && e != nil && e.narrativeTrace != nil && e.narrativeTrace.windowID != "" {
		trace := e.currentNarrativeTrace()
		event.Narrative = &trace
	}
	if err := event.Validate(); err != nil {
		panic(err)
	}
	if e.observer != nil {
		e.observer.OnGameEvent(event)
	}
}

func (e *GameEngine) Log(msg string) {
	e.emitGameEvent(model.GameEvent{Type: model.EventLog, Message: msg})
}

func (e *GameEngine) NotifyPrompt(prompt *model.Prompt) {
	e.emitGameEvent(model.GameEvent{
		Type:   model.EventAskInput,
		Prompt: e.decoratePromptForClient(prompt),
	})
}

func (e *GameEngine) NotifyGameEnd(msg string) {
	e.emitGameEvent(model.GameEvent{Type: model.EventGameEnd, Message: msg})
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
			e.dispatchCardTiming(p, model.TimingCardPlayedRevealed, "", cards[i])
		}
	}
	playerName := playerID
	if p != nil {
		playerName = p.Name
	}
	e.emitGameEvent(model.GameEvent{
		Type: model.EventCardRevealed,
		CardRevealed: &model.CardRevealedPayload{
			PlayerID:   playerID,
			PlayerName: playerName,
			Cards:      cards,
			ActionType: string(actionType),
			Hidden:     hidden,
		},
		Narrative: e.cardNarrativeTrace(actionType, hidden),
	})
}

func (e *GameEngine) cardNarrativeTrace(actionType model.DamageType, hidden bool) *model.NarrativeTracePayload {
	trace := e.currentNarrativeTrace()
	trace.NarrativeKind = "card_played"
	trace.VisualKind = "card"
	trace.CardRole = strings.ToLower(string(actionType))
	if hidden || trace.CardRole == "discard" {
		trace.VisualKind = "none"
	}
	if trace.CardRole == "magic" {
		trace.Timing = "magic.play_card"
	} else if trace.CardRole == "attack" {
		trace.Timing = "attack.play_card"
	}
	return &trace
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
	e.emitGameEvent(model.GameEvent{
		Type: model.EventDamageDealt,
		DamageDealt: &model.DamageDealtPayload{
			SourceID:   sourceID,
			SourceName: sourceName,
			TargetID:   targetID,
			TargetName: targetName,
			Damage:     damage,
			DamageType: string(damageType),
		},
		Narrative: e.narrativeTraceWith("damage_dealt", "damage"),
	})
}

func (e *GameEngine) NotifyActionStep(line string) {
	if e.observer == nil || line == "" {
		return
	}
	e.emitGameEvent(model.GameEvent{
		Type:       model.EventActionStep,
		ActionStep: &model.ActionStepPayload{Line: line, Kind: "detail"},
	})
}

func (e *GameEngine) NotifyActionSummaryNote(line string) {
	if e == nil || line == "" {
		return
	}
	e.addActionSummaryNote(line)
}

func (e *GameEngine) NotifyActionSummary(line string) {
	if e.observer == nil || line == "" {
		return
	}
	e.emitGameEvent(model.GameEvent{
		Type:       model.EventActionStep,
		ActionStep: &model.ActionStepPayload{Line: line, Kind: "summary"},
	})
}

func (e *GameEngine) NotifyCombatCue(attackerID, targetID, phase string) {
	if e.observer == nil || attackerID == "" || targetID == "" || phase == "" {
		return
	}
	e.emitGameEvent(model.GameEvent{
		Type: model.EventCombatCue,
		CombatCue: &model.CombatCuePayload{
			AttackerID: attackerID,
			TargetID:   targetID,
			Phase:      phase,
		},
		Narrative: e.combatCueNarrativeTrace(phase),
	})
}

func (e *GameEngine) combatCueNarrativeTrace(phase string) *model.NarrativeTracePayload {
	trace := e.currentNarrativeTrace()
	if phase == "attack" {
		trace.NarrativeKind = "combat_declared"
		trace.Timing = "attack.declare"
	} else {
		trace.NarrativeKind = "combat_response"
		trace.Timing = "combat.response"
	}
	trace.VisualKind = "none"
	trace.CardRole = phase
	return &trace
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
	e.emitGameEvent(model.GameEvent{
		Type: model.EventDrawCards,
		DrawCards: &model.DrawCardsPayload{
			PlayerID:   playerID,
			PlayerName: playerName,
			DrawCount:  count,
			Reason:     reason,
		},
	})
}

func (e *GameEngine) NotifySkillActivated(playerID, skillID, skillName, effectText string, targetIDs []string) {
	if e.observer == nil || playerID == "" || skillID == "" || skillName == "" {
		return
	}
	p := e.State.Players[playerID]
	playerName := playerID
	if p != nil {
		playerName = p.Name
	}
	e.emitGameEvent(model.GameEvent{
		Type: model.EventSkillActivated,
		SkillActivated: &model.SkillActivatedPayload{
			PlayerID:   playerID,
			PlayerName: playerName,
			SkillID:    skillID,
			SkillName:  skillName,
			EffectText: effectText,
			TargetIDs:  append([]string{}, targetIDs...),
		},
		Narrative: e.skillNarrativeTrace("skill_triggered", "triggered"),
	})
}

func (e *GameEngine) skillNarrativeTrace(kind, phase string) *model.NarrativeTracePayload {
	trace := e.currentNarrativeTrace()
	if e != nil && e.narrativeTrace != nil && e.narrativeTrace.actionType == "skill" {
		kind = "skill_resolved"
		phase = "resolved"
		trace.VisualKind = "none"
	} else {
		trace.VisualKind = "skill_token"
	}
	trace.NarrativeKind = kind
	trace.SkillPhase = phase
	trace.Timing = "skill." + phase
	return &trace
}

func (e *GameEngine) NotifySpecialAction(playerID string, actionType model.ActionType, summary string, targetIDs []string) {
	if e.observer == nil || playerID == "" || actionType == "" {
		return
	}
	p := e.State.Players[playerID]
	playerName := playerID
	if p != nil {
		playerName = p.Name
	}
	if summary == "" {
		summary = specialActionSummary(playerName, actionType)
	}
	e.emitGameEvent(model.GameEvent{
		Type: model.EventSpecialAction,
		SpecialAction: &model.SpecialActionPayload{
			PlayerID:   playerID,
			PlayerName: playerName,
			ActionType: string(actionType),
			TargetIDs:  append([]string{}, targetIDs...),
			Summary:    summary,
		},
	})
}

func specialActionSummary(playerName string, actionType model.ActionType) string {
	label := string(actionType)
	switch actionType {
	case model.ActionBuy:
		label = "购买"
	case model.ActionSynthesize:
		label = "合成"
	case model.ActionExtract:
		label = "提炼"
	}
	if playerName == "" {
		return label
	}
	return fmt.Sprintf("%s 执行特殊行动【%s】", playerName, label)
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
		e.NotifyPrompt(prompt)
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
