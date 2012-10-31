package main

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
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
var re map[string]*regexp.Regexp

func init() {
	re = map[string]*regexp.Regexp{
		"cc_start": regexp.MustCompile(`Sending start message (.+) to DEA (\w+)`),
	}
	eventDetectors = map[string]EventDetector{
		"supervisord": func(text string) *Event {
			if strings.Contains(text, "entered RUNNING state") {
				return &Event{Type: "process_started", Desc: "Something?? entered running state"}
			}
			return nil
		},
		"cloud_controller": func(text string) *Event {
			if strings.Contains(text, "Sending start message") {
				results := re["cc_start"].FindStringSubmatch(text)
				if results == nil {
					return nil // TODO: nope, just fall over
				}
				println(results)
				infoJson, _ := results[1], results[2]
				var info map[string]interface{}
				// XXX: doing this merely to extract the appname ...maybe we shouldn't?
				err := json.Unmarshal([]byte(infoJson), &info)
				if err != nil {
					log.Printf("Error: failed to parse the json in a cc_start record: %s", err)
					return nil
				}
				desc := fmt.Sprintf("Starting application %v", info["name"])
				return &Event{Type: "cc_start", Desc: desc, Info: infoJson}
			}
			return nil
		},
	}
}
