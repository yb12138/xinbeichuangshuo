// gameflow: 向观察者/前端通知当前 Prompt。

package engine

import (
	"fmt"
	"strconv"
	"strings"

	"starcup-engine/internal/engine/hook/promptfmt"
	"starcup-engine/internal/model"
)

func elementNameForPrompt(raw string) string {
	return promptfmt.ElementName(raw)
}

var promptButtonLabelByID = map[string]string{
	"confirm":    "发动",
	"yes":        "发动",
	"no":         "放弃",
	"cancel":     "取消",
	"skip":       "放弃",
	"take":       "承受",
	"counter":    "应战",
	"defend":     "防御",
	"normal":     "顺序",
	"reverse":    "反向",
	"attack":     "攻击",
	"magic":      "法术",
	"special":    "特殊",
	"buy":        "购买",
	"synthesize": "合成",
	"extract":    "提炼",
	"cannot_act": "放弃",
	"pass":       "放弃",
}

func parsePromptNonNegativeInt(raw string) (int, bool) {
	val, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || val < 0 {
		return 0, false
	}
	return val, true
}

func isPromptDeclineLabel(label string) bool {
	trimmed := strings.TrimSpace(label)
	if trimmed == "" {
		return false
	}
	return strings.Contains(trimmed, "不发动") ||
		strings.Contains(trimmed, "放弃") ||
		strings.Contains(trimmed, "跳过") ||
		strings.Contains(trimmed, "无法行动") ||
		strings.Contains(trimmed, "拒绝")
}

func shouldUseNumericPromptButtons(prompt *model.Prompt, options []model.PromptOption) (bool, bool) {
	if prompt == nil || len(options) < 2 {
		return false, false
	}
	if prompt.Type == model.PromptChooseCards {
		return false, false
	}
	numericIDs := make([]int, 0, len(options))
	hasLongLabel := false
	hasXHint := strings.ContainsAny(strings.ToLower(prompt.Message), "xｘ")
	for _, option := range options {
		if n, ok := parsePromptNonNegativeInt(option.ID); ok {
			numericIDs = append(numericIDs, n)
		}
		label := strings.TrimSpace(option.Label)
		if len([]rune(label)) >= 8 || strings.Contains(label, "分支") {
			hasLongLabel = true
		}
		lowLabel := strings.ToLower(label)
		if strings.Contains(lowLabel, "x=") || strings.Contains(label, "X=") || strings.Contains(lowLabel, "x值") || strings.ContainsAny(lowLabel, "xｘ") {
			hasXHint = true
		}
	}
	if len(numericIDs) < 2 || (!hasLongLabel && !hasXHint) {
		return false, false
	}
	minID := numericIDs[0]
	for _, id := range numericIDs[1:] {
		if id < minID {
			minID = id
		}
	}
	return true, minID == 0
}

func normalizePromptOptionForClient(option model.PromptOption, prompt *model.Prompt, useNumeric bool, plusOne bool) model.PromptOption {
	label := strings.TrimSpace(option.Label)
	button := strings.TrimSpace(option.ButtonLabel)
	hint := strings.TrimSpace(option.Hint)
	optionID := strings.ToLower(strings.TrimSpace(option.ID))

	if button == "" {
		if mapped, ok := promptButtonLabelByID[optionID]; ok {
			button = mapped
		}
	}
	if button == "" && prompt != nil && prompt.Type == model.PromptChooseSkill {
		button = "发动"
	}
	if button == "" && optionID == "-1" {
		if strings.Contains(label, "完成") || strings.Contains(label, "结束") {
			button = "完成"
		} else {
			button = "放弃"
		}
	}
	if button == "" && useNumeric {
		if n, ok := parsePromptNonNegativeInt(option.ID); ok {
			if plusOne {
				button = strconv.Itoa(n + 1)
			} else {
				button = strconv.Itoa(n)
			}
		}
	}
	if button == "" && isPromptDeclineLabel(label) {
		button = "放弃"
	}
	if button == "" {
		if label != "" && len([]rune(label)) <= 6 {
			button = label
		} else {
			button = "执行"
		}
	}

	isCombatResponseOption := optionID == "take" || optionID == "defend" || optionID == "counter" ||
		button == "承受" || button == "防御" || button == "应战"
	if isCombatResponseOption {
		hint = ""
	}

	if hint == "" && !isCombatResponseOption && label != "" && label != button {
		if !(button == "取消" && (label == "取消" || label == "取消/跳过")) &&
			!(button == "放弃" && isPromptDeclineLabel(label)) {
			hint = label
		}
	}

	option.ButtonLabel = button
	option.Hint = hint
	return option
}

func (e *GameEngine) decoratePromptForClient(prompt *model.Prompt) *model.Prompt {
	if prompt == nil {
		return nil
	}
	cp := *prompt
	if prompt.Options != nil {
		useNumeric, plusOne := shouldUseNumericPromptButtons(prompt, prompt.Options)
		cp.Options = make([]model.PromptOption, 0, len(prompt.Options))
		for _, option := range prompt.Options {
			cp.Options = append(cp.Options, normalizePromptOptionForClient(option, prompt, useNumeric, plusOne))
		}
	}
	if prompt.SpecialOptions != nil {
		useNumeric, plusOne := shouldUseNumericPromptButtons(prompt, prompt.SpecialOptions)
		cp.SpecialOptions = make([]model.PromptOption, 0, len(prompt.SpecialOptions))
		for _, option := range prompt.SpecialOptions {
			cp.SpecialOptions = append(cp.SpecialOptions, normalizePromptOptionForClient(option, prompt, useNumeric, plusOne))
		}
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
	if eventType == model.EventAskInput {
		switch p := data.(type) {
		case *model.Prompt:
			data = e.decoratePromptForClient(p)
		case model.Prompt:
			cp := p
			data = e.decoratePromptForClient(&cp)
		}
	}
	if e.observer != nil {
		e.observer.OnGameEvent(model.GameEvent{
			Type:    eventType,
			Message: msg,
			Data:    data,
		})
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
	skillCtx := e.buildContext(player, nil, timing, cardCtx)
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
	e.Notify(model.EventCardRevealed, "", map[string]interface{}{
		"player_id":   playerID,
		"player_name": playerName,
		"cards":       cards,
		"action_type": string(actionType),
		"hidden":      hidden,
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
	e.Notify(model.EventDamageDealt, "", map[string]interface{}{
		"source_id":   sourceID,
		"source_name": sourceName,
		"target_id":   targetID,
		"target_name": targetName,
		"damage":      damage,
		"damage_type": string(damageType),
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
	e.Notify(model.EventActionStep, "", map[string]interface{}{
		"line": line,
		"kind": "detail",
	})
}

func (e *GameEngine) NotifyActionSummary(line string) {
	if e.observer == nil || line == "" {
		return
	}
	e.Notify(model.EventActionStep, "", map[string]interface{}{
		"line": line,
		"kind": "summary",
	})
}

func (e *GameEngine) NotifyCombatCue(attackerID, targetID, phase string) {
	if e.observer == nil || attackerID == "" || targetID == "" || phase == "" {
		return
	}
	e.Notify(model.EventCombatCue, "", map[string]interface{}{
		"attacker_id": attackerID,
		"target_id":   targetID,
		"phase":       phase,
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
	e.Notify(model.EventDrawCards, "", map[string]interface{}{
		"player_id":   playerID,
		"player_name": playerName,
		"draw_count":  count,
		"reason":      reason,
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
	prompt := e.buildPendingInterruptPrompt()
	if prompt != nil {
		e.Notify(model.EventAskInput, "", prompt)
	}
}

func (e *GameEngine) buildChoicePrompt() *model.Prompt {
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
		e.Log(fmt.Sprintf("[System] buildChoicePrompt: %v", err))
		return nil
	}
	return p
}
