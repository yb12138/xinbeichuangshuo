package server

import (
	"fmt"
	"sync/atomic"

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
