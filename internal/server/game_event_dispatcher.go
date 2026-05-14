package server

import (
	"starcup-engine/internal/model"
	"starcup-engine/internal/server/prompting"
)

type scheduledBotPrompt struct {
	playerID string
	prompt   *model.Prompt
	epoch    uint64
}

// OnGameEvent implements model.GameObserver.
func (r *Room) OnGameEvent(event model.GameEvent) {
	scheduledBot := r.dispatchGameEvent(event)
	if scheduledBot.playerID != "" {
		go r.scheduleBotIfNeeded(scheduledBot.playerID, scheduledBot.prompt, scheduledBot.epoch)
	}
}

func (r *Room) dispatchGameEvent(event model.GameEvent) scheduledBotPrompt {
	switch event.Type {
	case model.EventLog:
		r.handleLogGameEvent(event)
	case model.EventStateUpdate:
		r.handleStateUpdateGameEvent()
	case model.EventAskInput:
		return r.handleAskInputGameEvent(event)
	case model.EventError:
		r.handleErrorGameEvent(event)
	case model.EventGameEnd:
		r.handleGameEndGameEvent(event)
	case model.EventCardRevealed:
		r.handleCardRevealedGameEvent(event)
	case model.EventDamageDealt:
		r.handleDamageDealtGameEvent(event)
	case model.EventActionStep:
		r.handleActionStepGameEvent(event)
	case model.EventCombatCue:
		r.handleCombatCueGameEvent(event)
	case model.EventDrawCards:
		r.handleDrawCardsGameEvent(event)
	}
	return scheduledBotPrompt{}
}

func (r *Room) handleLogGameEvent(event model.GameEvent) {
	r.broadcastTimeline("log", map[string]interface{}{
		"message": event.Message,
	}, event.Message)
}

func (r *Room) handleStateUpdateGameEvent() {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, client := range r.Clients {
		r.sendSyncStateToClient(client)
	}
}

func (r *Room) handleAskInputGameEvent(event model.GameEvent) scheduledBotPrompt {
	prompt := promptFromGameEventData(event.Data)
	if prompt == nil {
		return scheduledBotPrompt{}
	}

	// 先推送状态给所有人（含手牌），确保客户端有最新数据。
	r.mu.Lock()
	defer r.mu.Unlock()

	r.botPromptEpoch++
	scheduledBot := scheduledBotPrompt{epoch: r.botPromptEpoch}

	// 一次 AskInput 仅有一个有效提示，清空旧缓存避免旧定时器误动作。
	r.botPromptCache = map[string]*model.Prompt{
		prompt.PlayerID: prompting.ClonePrompt(prompt),
	}
	for _, client := range r.Clients {
		r.sendSyncStateToClient(client)
	}

	// Send prompt only to the target player.
	if client, exists := r.Clients[prompt.PlayerID]; exists {
		if client.IsBot {
			scheduledBot.playerID = client.PlayerID
			scheduledBot.prompt = prompting.ClonePrompt(prompt)
		} else if client.Disconnected {
			// 真人离线且未托管：暂停等待重连或房主手动托管。
		} else {
			r.sendRequireActionToClient(client, prompt)
		}
	}

	// Non-target players 也收到同一 RequireAction，用于显示等待态。
	for pid, client := range r.Clients {
		if pid != prompt.PlayerID {
			r.sendRequireActionToClient(client, prompt)
		}
	}

	r.broadcastTimeline("prompt", map[string]interface{}{
		"player_id": prompt.PlayerID,
		"message":   prompt.Message,
	}, prompt.Message)

	return scheduledBot
}

func (r *Room) handleErrorGameEvent(event model.GameEvent) {
	r.broadcastTimeline("error", map[string]interface{}{
		"message": event.Message,
	}, event.Message)
}

func (r *Room) handleGameEndGameEvent(event model.GameEvent) {
	r.broadcastTimeline("game_end", map[string]interface{}{
		"message": event.Message,
	}, event.Message)
}

func (r *Room) handleCardRevealedGameEvent(event model.GameEvent) {
	payload, ok := event.Data.(model.CardRevealedPayload)
	if !ok {
		return
	}
	data := map[string]interface{}{
		"player_id":   payload.PlayerID,
		"player_name": payload.PlayerName,
		"cards":       payload.Cards,
		"action_type": payload.ActionType,
		"hidden":      payload.Hidden,
	}
	r.botIntel.ObserveReveal(data)
	r.broadcastTimeline("card_revealed", data, "")
}

func (r *Room) handleDamageDealtGameEvent(event model.GameEvent) {
	payload, ok := event.Data.(model.DamageDealtPayload)
	if !ok {
		return
	}
	data := map[string]interface{}{
		"source_id":   payload.SourceID,
		"source_name": payload.SourceName,
		"target_id":   payload.TargetID,
		"target_name": payload.TargetName,
		"damage":      payload.Damage,
		"damage_type": payload.DamageType,
	}
	r.broadcastTimeline("damage_dealt", data, "")
}

func (r *Room) handleActionStepGameEvent(event model.GameEvent) {
	payload, ok := event.Data.(model.ActionStepPayload)
	if !ok {
		return
	}
	data := map[string]interface{}{
		"line": payload.Line,
		"kind": payload.Kind,
	}
	r.broadcastTimeline("action_step", data, payload.Line)
}

func (r *Room) handleCombatCueGameEvent(event model.GameEvent) {
	payload, ok := event.Data.(model.CombatCuePayload)
	if !ok {
		return
	}
	data := map[string]interface{}{
		"attacker_id": payload.AttackerID,
		"target_id":   payload.TargetID,
		"phase":       payload.Phase,
	}
	r.broadcastTimeline("combat_cue", data, "")
}

func (r *Room) handleDrawCardsGameEvent(event model.GameEvent) {
	payload, ok := event.Data.(model.DrawCardsPayload)
	if !ok {
		return
	}
	data := map[string]interface{}{
		"player_id":   payload.PlayerID,
		"player_name": payload.PlayerName,
		"draw_count":  payload.DrawCount,
		"reason":      payload.Reason,
	}
	r.broadcastTimeline("draw_cards", data, payload.Reason)
}

func (r *Room) broadcastTimeline(eventType string, data map[string]interface{}, message string) {
	r.broadcastHumans(CmdNotifyTimeline, r.buildTimelineNotify(eventType, data, message))
}

func promptFromGameEventData(data interface{}) *model.Prompt {
	switch prompt := data.(type) {
	case *model.Prompt:
		return prompt
	case model.Prompt:
		cp := prompt
		return &cp
	default:
		return nil
	}
}
