package server

import (
	"starcup-engine/internal/model"
	"starcup-engine/internal/server/prompting"
	"starcup-engine/internal/server/timeline"
)

type scheduledBotPrompt struct {
	playerID string
	prompt   *model.Prompt
	epoch    uint64
}

// OnGameEvent implements model.GameObserver.
func (r *Room) OnGameEvent(event model.GameEvent) {
	scheduledBot := r.dispatchGameEvent(event)
	if event.Type != model.EventStateDelta {
		r.broadcastPublicStateDelta(string(event.Type))
	}
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
	case model.EventSkillActivated:
		r.handleSkillActivatedGameEvent(event)
	case model.EventSpecialAction:
		r.handleSpecialActionGameEvent(event)
	case model.EventStateDelta:
		r.handleStateDeltaGameEvent(event)
	}
	return scheduledBotPrompt{}
}

func (r *Room) handleLogGameEvent(event model.GameEvent) {
	r.broadcastTimeline(timeline.Payload{Type: "log", Message: event.Message})
}

func (r *Room) handleStateUpdateGameEvent() {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, client := range r.Clients {
		r.sendSyncStateToClient(client)
	}
}

func (r *Room) handleAskInputGameEvent(event model.GameEvent) scheduledBotPrompt {
	prompt := event.Prompt
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

	r.broadcastTimeline(timeline.Payload{Type: "prompt", PlayerID: prompt.PlayerID, Message: prompt.Message})

	return scheduledBot
}

func (r *Room) handleErrorGameEvent(event model.GameEvent) {
	r.broadcastTimeline(timeline.Payload{Type: "error", Message: event.Message})
}

func (r *Room) handleGameEndGameEvent(event model.GameEvent) {
	r.broadcastTimeline(timeline.Payload{Type: "game_end", Message: event.Message})
}

func (r *Room) handleCardRevealedGameEvent(event model.GameEvent) {
	payload := event.CardRevealed
	if payload == nil {
		return
	}
	r.botIntel.ObserveCardRevealed(*payload)
	r.broadcastTimeline(timeline.Payload{
		Type:       "card_revealed",
		PlayerID:   payload.PlayerID,
		PlayerName: payload.PlayerName,
		Cards:      payload.Cards,
		ActionType: payload.ActionType,
		Hidden:     payload.Hidden,
	})
}

func (r *Room) handleDamageDealtGameEvent(event model.GameEvent) {
	payload := event.DamageDealt
	if payload == nil {
		return
	}
	r.broadcastTimeline(timeline.Payload{
		Type:       "damage_dealt",
		SourceID:   payload.SourceID,
		SourceName: payload.SourceName,
		TargetID:   payload.TargetID,
		TargetName: payload.TargetName,
		Damage:     payload.Damage,
		DamageType: payload.DamageType,
	})
}

func (r *Room) handleActionStepGameEvent(event model.GameEvent) {
	payload := event.ActionStep
	if payload == nil {
		return
	}
	r.broadcastTimeline(timeline.Payload{Type: "action_step", Message: payload.Line, Kind: payload.Kind})
}

func (r *Room) handleCombatCueGameEvent(event model.GameEvent) {
	payload := event.CombatCue
	if payload == nil {
		return
	}
	r.broadcastTimeline(timeline.Payload{
		Type:       "combat_cue",
		AttackerID: payload.AttackerID,
		TargetID:   payload.TargetID,
		Phase:      payload.Phase,
	})
}

func (r *Room) handleDrawCardsGameEvent(event model.GameEvent) {
	payload := event.DrawCards
	if payload == nil {
		return
	}
	r.broadcastTimeline(timeline.Payload{
		Type:       "draw_cards",
		Message:    payload.Reason,
		PlayerID:   payload.PlayerID,
		PlayerName: payload.PlayerName,
		DrawCount:  payload.DrawCount,
		Reason:     payload.Reason,
	})
}

func (r *Room) handleSkillActivatedGameEvent(event model.GameEvent) {
	payload := event.SkillActivated
	if payload == nil {
		return
	}
	r.broadcastTimeline(timeline.Payload{
		Type:       "skill_activated",
		PlayerID:   payload.PlayerID,
		PlayerName: payload.PlayerName,
		SkillID:    payload.SkillID,
		SkillName:  payload.SkillName,
		EffectText: payload.EffectText,
		TargetIDs:  append([]string{}, payload.TargetIDs...),
	})
}

func (r *Room) handleSpecialActionGameEvent(event model.GameEvent) {
	payload := event.SpecialAction
	if payload == nil {
		return
	}
	r.broadcastTimeline(timeline.Payload{
		Type:       "special_action",
		PlayerID:   payload.PlayerID,
		PlayerName: payload.PlayerName,
		ActionType: payload.ActionType,
		TargetIDs:  append([]string{}, payload.TargetIDs...),
		Summary:    payload.Summary,
		Message:    payload.Summary,
	})
}

func (r *Room) handleStateDeltaGameEvent(event model.GameEvent) {
	payload := event.StateDelta
	if payload == nil || len(payload.Deltas) == 0 {
		return
	}
	r.broadcastTimeline(timeline.Payload{
		Type:   "state_delta",
		Reason: payload.Reason,
		Deltas: timelineDeltasFromModel(payload.Deltas),
	})
}

func (r *Room) broadcastTimeline(payload timeline.Payload) {
	r.broadcastHumans(CmdNotifyTimeline, r.buildTimelineNotify(payload))
}

func timelineDeltasFromModel(items []model.StateDeltaItem) []TimelineDelta {
	if len(items) == 0 {
		return nil
	}
	out := make([]TimelineDelta, 0, len(items))
	for _, item := range items {
		out = append(out, TimelineDelta{
			Type:          item.Type,
			Scope:         item.Scope,
			TargetUserID:  item.TargetUserID,
			Camp:          item.Camp,
			Field:         item.Field,
			Before:        item.Before,
			After:         item.After,
			Value:         item.Value,
			Reason:        item.Reason,
			SourceEventID: item.SourceEventID,
			BeforeText:    item.BeforeText,
			AfterText:     item.AfterText,
			FieldCard:     item.FieldCard,
		})
	}
	return out
}
