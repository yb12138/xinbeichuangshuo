package engine

import (
	"fmt"
	"strings"

	"starcup-engine/internal/model"
)

type narrativeTraceState struct {
	windowID      string
	windowActor   string
	actionID      string
	actionActor   string
	actionType    string
	combatID      string
	actionSeq     int
	combatSeq     int
	sourceEventID string
}

func (e *GameEngine) ensureNarrativeWindow(playerID string) string {
	if e == nil || e.State == nil || playerID == "" {
		return ""
	}
	if e.narrativeTrace == nil {
		e.narrativeTrace = &narrativeTraceState{}
	}
	if e.narrativeTrace.windowID == "" || e.narrativeTrace.windowActor != playerID {
		e.narrativeTrace.windowActor = playerID
		e.narrativeTrace.windowID = fmt.Sprintf("nw-t%d-%s", e.State.CurrentTurn+1, playerID)
		e.narrativeTrace.actionID = ""
		e.narrativeTrace.actionActor = ""
		e.narrativeTrace.actionType = ""
		e.narrativeTrace.combatID = ""
		e.narrativeTrace.actionSeq = 0
		e.narrativeTrace.combatSeq = 0
		e.narrativeTrace.sourceEventID = ""
	}
	return e.narrativeTrace.windowID
}

func (e *GameEngine) beginNarrativeAction(actionType, actorID string) model.NarrativeTracePayload {
	if e == nil || e.State == nil || actorID == "" {
		return model.NarrativeTracePayload{}
	}
	windowID := e.ensureNarrativeWindow(actorID)
	e.narrativeTrace.actionSeq++
	e.narrativeTrace.actionActor = actorID
	e.narrativeTrace.actionType = strings.ToLower(actionType)
	e.narrativeTrace.actionID = fmt.Sprintf("%s-a%d-%s", windowID, e.narrativeTrace.actionSeq, strings.ToLower(actionType))
	e.narrativeTrace.combatID = ""
	e.narrativeTrace.sourceEventID = e.narrativeTrace.actionID
	return e.currentNarrativeTrace()
}

func (e *GameEngine) beginNarrativeCombat(attackerID string, isCounter bool) model.NarrativeTracePayload {
	if e == nil || e.State == nil || attackerID == "" {
		return model.NarrativeTracePayload{}
	}
	windowID := e.ensureNarrativeWindow(firstNonEmptyString(e.currentNarrativeWindowActor(), attackerID))
	if e.narrativeTrace.actionID == "" {
		e.beginNarrativeAction("attack", attackerID)
	}
	e.narrativeTrace.combatSeq++
	suffix := "combat"
	if isCounter {
		suffix = "counter"
	}
	e.narrativeTrace.combatID = fmt.Sprintf("%s-c%d-%s", windowID, e.narrativeTrace.combatSeq, suffix)
	return e.currentNarrativeTrace()
}

func (e *GameEngine) currentNarrativeWindowActor() string {
	if e == nil || e.narrativeTrace == nil {
		return ""
	}
	return e.narrativeTrace.windowActor
}

func (e *GameEngine) currentNarrativeTrace() model.NarrativeTracePayload {
	if e == nil || e.narrativeTrace == nil {
		return model.NarrativeTracePayload{}
	}
	return model.NarrativeTracePayload{
		NarrativeWindowID: e.narrativeTrace.windowID,
		ActionID:          e.narrativeTrace.actionID,
		CombatID:          e.narrativeTrace.combatID,
		SourceEventID:     e.narrativeTrace.sourceEventID,
	}
}

func (e *GameEngine) narrativeTraceWith(kind, visual string) *model.NarrativeTracePayload {
	trace := e.currentNarrativeTrace()
	trace.NarrativeKind = kind
	trace.VisualKind = visual
	return &trace
}

func (e *GameEngine) publishTimelineMarker(marker model.TimelineMarkerPayload) {
	if e == nil || e.observer == nil {
		return
	}
	trace := e.currentNarrativeTrace()
	if marker.NarrativeKind != "" {
		trace.NarrativeKind = marker.NarrativeKind
	}
	if marker.VisualKind != "" {
		trace.VisualKind = marker.VisualKind
	}
	trace.CardRole = marker.CardRole
	trace.SkillPhase = marker.SkillPhase
	trace.Timing = marker.Timing
	trace.EffectType = marker.EffectType
	trace.ExtraActionType = marker.ExtraActionType
	trace.ExtraActionElement = marker.ExtraActionElement
	trace.FieldCard = marker.FieldCard
	e.emitGameEvent(model.GameEvent{
		Type:           model.EventTimelineMarker,
		TimelineMarker: &marker,
		Narrative:      &trace,
	})
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func pendingActionSignature(action model.ActionContext) string {
	element := ""
	if len(action.MustElement) > 0 {
		parts := make([]string, 0, len(action.MustElement))
		for _, item := range action.MustElement {
			parts = append(parts, string(item))
		}
		element = strings.Join(parts, ",")
	}
	return strings.Join([]string{action.Source, action.MustType, element}, "|")
}

func snapshotPendingActions(p *model.Player) map[string]int {
	out := map[string]int{}
	if p == nil {
		return out
	}
	for _, action := range p.TurnState.PendingActions {
		out[pendingActionSignature(action)]++
	}
	return out
}

func (e *GameEngine) publishPendingActionDiff(player *model.Player, before map[string]int, source string) {
	if e == nil || player == nil {
		return
	}
	seen := map[string]int{}
	for _, action := range player.TurnState.PendingActions {
		sig := pendingActionSignature(action)
		seen[sig]++
		if seen[sig] <= before[sig] {
			continue
		}
		element := ""
		if len(action.MustElement) > 0 {
			element = string(action.MustElement[0])
		}
		e.publishTimelineMarker(model.TimelineMarkerPayload{
			PlayerID:           player.ID,
			PlayerName:         player.Name,
			ActionType:         action.MustType,
			Summary:            source,
			NarrativeKind:      "extra_action_granted",
			VisualKind:         "action_marker",
			ExtraActionType:    action.MustType,
			ExtraActionElement: element,
		})
	}
}
