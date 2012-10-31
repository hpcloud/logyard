package main

import (
	"encoding/json"
	"fmt"
	"log"
)

type Event struct {
	Type    string // what type of event?
	Desc    string // description of this event to be shown as-is to humans
	Info    string // event-specific information as json
	Process string // which process generated this event?
	NodeID  string // from which node did this event appear?
}

type EventDetector func(text string) *Event

var eventDetectors map[string]EventDetector

// re is a re cache
var matchers map[string]*MultiRegexpMatcher

func init() {
	matchers = map[string]*MultiRegexpMatcher{}
	matchers["supervisord"] = NewMultiRegexpMatcher()
	matchers["supervisord"].MustAdd("process_started", "entered RUNNING state", `entered`)
	matchers["cloud_controller"] = NewMultiRegexpMatcher()
	matchers["cloud_controller"].MustAdd("cc_start", "Sending start message", `Sending start message (.+) to DEA (\w+)`)

	for _, matcher := range matchers {
		matcher.Build()
	}
}

func ParseEvent(process string, record string) *Event {
	if matcher, ok := matchers[process]; ok {
		if event_type, results := matcher.Match(record); results != nil {
			switch event_type {
			case "process_started":
				return &Event{
					Type:    event_type,
					Process: process,
					Desc:    "Something?? was strted"}
			case "cc_start":
				infoJson, _ := results[1], results[2]
				var info map[string]interface{}
				// XXX: doing this merely to extract the appname ...maybe we shouldn't?
				err := json.Unmarshal([]byte(infoJson), &info)
				if err != nil {
					log.Printf("Error: failed to parse the json in a cc_start record: %s", err)
					return nil
				}
				return &Event{
					Type:    "cc_start",
					Process: process,
					Desc:    fmt.Sprintf("Starting application '%v'", info["name"]),
					Info:    infoJson}
			}
		}
	}
	return nil
}
