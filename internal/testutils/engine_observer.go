package testutils

import (
	"reflect"
	"starcup-engine/internal/model"
	"strings"
	"testing"
)

// NoopObserver implements model.GameObserver with no-op behavior.
type NoopObserver struct{}

func (NoopObserver) OnGameEvent(event model.GameEvent) {}

// CaptureObserver collects all game events for inspection.
type CaptureObserver struct {
	Events []model.GameEvent
}

func (o *CaptureObserver) OnGameEvent(event model.GameEvent) {
	o.Events = append(o.Events, event)
}

// CountLogContains counts EventLog entries whose message contains substr.
func (o *CaptureObserver) CountLogContains(substr string) int {
	n := 0
	for _, e := range o.Events {
		if e.Type != model.EventLog {
			continue
		}
		if strings.Contains(e.Message, substr) {
			n++
		}
	}
	return n
}

// SkillActivated returns captured skill_activated payloads for the given skill ID.
func (o *CaptureObserver) SkillActivated(skillID string) []model.SkillActivatedPayload {
	var out []model.SkillActivatedPayload
	for _, e := range o.Events {
		if e.Type != model.EventSkillActivated || e.SkillActivated == nil {
			continue
		}
		if e.SkillActivated.SkillID == skillID {
			out = append(out, *e.SkillActivated)
		}
	}
	return out
}

// Reset clears all captured events.
func (o *CaptureObserver) Reset() {
	o.Events = nil
}

// RequireSingleSkillActivation asserts exactly one skill_activated for skillID, emitted by
// playerID with exactly targetIDs (order-sensitive), carrying a skill name and effect text.
func RequireSingleSkillActivation(t testing.TB, o *CaptureObserver, skillID, playerID string, targetIDs ...string) {
	t.Helper()
	got := o.SkillActivated(skillID)
	if len(got) != 1 {
		t.Fatalf("expected exactly 1 skill_activated for %s, got %d: %+v", skillID, len(got), got)
	}
	p := got[0]
	if p.PlayerID != playerID {
		t.Fatalf("%s: expected player %s, got %s", skillID, playerID, p.PlayerID)
	}
	if len(targetIDs) == 0 {
		targetIDs = nil
	}
	if !reflect.DeepEqual(p.TargetIDs, targetIDs) && !(len(p.TargetIDs) == 0 && len(targetIDs) == 0) {
		t.Fatalf("%s: expected targets %v, got %v", skillID, targetIDs, p.TargetIDs)
	}
	if p.SkillName == "" || p.EffectText == "" {
		t.Fatalf("%s: expected skill name and effect text, got %+v", skillID, p)
	}
}
