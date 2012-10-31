package main

import (
	"log"
)

type Event struct {
	Type     string // what type of event?
	Desc     string // description of this event to be shown as-is to humans
	Info     string // event-specific information as json
	Process  string // which process generated this event?
	UnixTime int64
	NodeID   string // from which node did this event appear?
}

// re is a re cache
var matchers map[string]*MultiRegexpMatcher

func init() {
	matchers = map[string]*MultiRegexpMatcher{}
	matchers["supervisord"] = NewMultiRegexpMatcher()
	matchers["supervisord"].MustAdd("process_start", "entered RUNNING state", `(\w) entered RUNNING`)
	matchers["supervisord"].MustAdd("process_stop", "stopped", `stopped: (\w+) \((.+)\)`)
	matchers["supervisord"].MustAdd("process_exit", "exited", `exited: (\w+) \((.+)\)`)
	matchers["kato"] = NewMultiRegexpMatcher()
	matchers["kato"].MustAdd("kato_action", "INVOKE", `INVOKE (.+)`)
	matchers["cloud_controller"] = NewMultiRegexpMatcher()
	matchers["cloud_controller"].MustAdd("cc_start", "Sending start message", `Sending start message (.+) to DEA (\w+)`)

	for _, matcher := range matchers {
		matcher.Build()
	}
}

func ParseEvent(process string, record string) *Event {
	if matcher, ok := matchers[process]; ok {
		if event_type, results := matcher.Match(record); results != nil {
			event := Event{Type: event_type, Process: process}
			if handler, ok := EventHandlers[event_type]; ok {
				err := handler.HandleEvent(results, &event)
				if err != nil {
					log.Println("Error handling %s; %s", event_type, err)
					return nil
				}
				return &event
			}
			log.Printf("Warning: no handler for event: %s", event_type)
		}
	}
	return nil
}
