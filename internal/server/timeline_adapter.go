package server

import (
	"fmt"
	"sync/atomic"

	"starcup-engine/internal/model"
	"starcup-engine/internal/server/timeline"
)

func (r *Room) nextTimelineEventMeta() (int64, string, int, string, string, string) {
	eventID := atomic.AddInt64(&r.timelineSeq, 1)

	turnID := 0
	turnStage := ""
	combatStage := ""
	subflow := ""
	if r.Engine != nil && r.Engine.State != nil {
		turnID = r.Engine.State.CurrentTurn + 1
		turnStage = string(r.Engine.State.TurnStage)
		combatStage = string(r.Engine.State.CombatStage)
		subflow = string(r.Engine.State.Subflow)
	}

	return eventID, fmt.Sprintf("chain_%d", eventID), turnID, turnStage, combatStage, subflow
}

func (r *Room) buildTimelineNotify(payload timeline.Payload) TimelineNotifyPayload {
	if payload.Trace == nil {
		payload.Trace = r.inheritActiveNarrativeTrace()
	}
	eventID, chainID, turnID, turnStage, combatStage, subflow := r.nextTimelineEventMeta()
	event := timeline.BuildEvent(timeline.EventMeta{
		EventID:     eventID,
		TurnID:      turnID,
		TurnStage:   turnStage,
		CombatStage: combatStage,
		Subflow:     subflow,
		ChainID:     chainID,
	}, payload)

	return TimelineNotifyPayload{
		RoomID:   r.Code,
		SeqStart: eventID,
		SeqEnd:   eventID,
		IsReplay: false,
		Events:   []TimelineEvent{event},
	}
}

func (r *Room) inheritActiveNarrativeTrace() *model.NarrativeTracePayload {
	r.timelineMu.Lock()
	defer r.timelineMu.Unlock()
	for i := len(r.timelineHistory) - 1; i >= 0; i-- {
		event := r.timelineHistory[i]
		if r.activeNarrativeWindowID != "" && event.NarrativeWindowID != r.activeNarrativeWindowID {
			continue
		}
		if event.NarrativeWindowID == "" {
			continue
		}
		return &model.NarrativeTracePayload{
			NarrativeWindowID: event.NarrativeWindowID,
			ActionID:          event.ActionID,
			CombatID:          event.CombatID,
			SourceEventID:     event.SourceEventID,
		}
	}
	return nil
}

func (r *Room) recordTimelineHistory(events []TimelineEvent) {
	if len(events) == 0 {
		return
	}
	r.timelineMu.Lock()
	defer r.timelineMu.Unlock()
	for _, event := range events {
		if event.NarrativeKind == "action_closed" {
			r.activeNarrativeWindowID = ""
		} else if event.NarrativeWindowID != "" {
			r.activeNarrativeWindowID = event.NarrativeWindowID
		}
		r.timelineHistory = append(r.timelineHistory, event)
	}
	const maxTimelineHistory = 200
	if extra := len(r.timelineHistory) - maxTimelineHistory; extra > 0 {
		r.timelineHistory = append([]TimelineEvent{}, r.timelineHistory[extra:]...)
	}
}

func (r *Room) buildTimelineReplayNotify() *TimelineNotifyPayload {
	r.timelineMu.Lock()
	defer r.timelineMu.Unlock()
	if r.activeNarrativeWindowID == "" {
		return nil
	}
	events := make([]TimelineEvent, 0)
	for _, event := range r.timelineHistory {
		if event.NarrativeWindowID == r.activeNarrativeWindowID {
			events = append(events, event)
		}
	}
	if len(events) == 0 {
		return nil
	}
	return &TimelineNotifyPayload{
		RoomID:   r.Code,
		SeqStart: events[0].EventID,
		SeqEnd:   events[len(events)-1].EventID,
		IsReplay: true,
		Events:   events,
	}
}
