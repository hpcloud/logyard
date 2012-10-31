package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"
)

// A simple event handler filling in only the description field.
// Suitable for most events.
type SimpleEventHandler struct {
	*template.Template
}

func (tmpl SimpleEventHandler) HandleEvent(results []string, event *Event) error {
	var output bytes.Buffer
	err := tmpl.Execute(&output, results)
	if err != nil {
		return err
	}
	event.Desc = output.String()
	return nil
}

// Custom event handler. Usually necessary for parsing json data out
// of log records.
type CustomEventHandler func(results []string, event *Event) error

func (h CustomEventHandler) HandleEvent(results []string, event *Event) error {
	return h(results, event)
}

type EventHandler interface {
	HandleEvent(results []string, event *Event) error
}

var EventHandlers map[string]EventHandler

func init() {
	desc := func(tmpl string) SimpleEventHandler {
		t := template.Must(template.New("").Parse(tmpl))
		return SimpleEventHandler{t}
	}

	EventHandlers = map[string]EventHandler{
		// supervisord
		"process_start": desc("Process '{{index . 1}}' is now running"),
		"process_stop":  desc("Process '{{index . 1}}' is stopped ({{index . 2}})"),
		"process_exit":  desc("Process '{{index . 1}}' has crashed ({{index . 2}})"),

		// kato
		"kato_action": desc("kato action taken on a node: {{index . 1}}"),

		// cc
		"cc_start": CustomEventHandler(func(results []string, event *Event) error {
			event.Info, _ = results[1], results[2]
			var info map[string]interface{}
			// XXX: doing this merely to extract the appname ...maybe we shouldn't?
			err := json.Unmarshal([]byte(event.Info), &info)
			if err != nil {
				return fmt.Errorf("failed to parse the json in a cc_start record: %s", err)
			}
			event.Desc = fmt.Sprintf("Starting application '%v'", info["name"])
			return nil
		}),
	}
}
