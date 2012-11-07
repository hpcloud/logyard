package logyard

import (
	"time"
)

// A event count tracker (for last X duration) with a _worst_ case
// space complexity of O(N). If events are rare (as is typically the
// case with error events), then average space complexity is O(1).
// Time complexity is always O(1).
type Tracker struct {
	events  []time.Time
	left, n int
}

func NewTracker(n int) *Tracker {
	return &Tracker{events: []time.Time{}, n: n}
}

// Track a new event that happened now
func (t *Tracker) Event() {
	t.event(time.Now())
}

func (t *Tracker) event(w time.Time) {
	if len(t.events) < t.n {
		t.events = append(t.events, w)
	} else {
		// re-use like a ring buffer
		t.events[t.left] = w
		t.left = (t.left + 1) % len(t.events)
	}
}

// Were there N events in the last `d` duration?
func (t *Tracker) In(d time.Duration) bool {
	return len(t.events) == t.n && time.Since(t.events[t.left]) <= d
}
