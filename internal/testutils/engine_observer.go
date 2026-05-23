package testutils

import (
	"starcup-engine/internal/model"
	"strings"
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
