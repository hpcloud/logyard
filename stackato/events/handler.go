package events

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"text/template"
)

type EventHandler interface {
	HandleEvent(results []string, event *Event) error
}

// A simple event handler filling in only the description field.
// Suitable for most events.
type SimpleEventHandler struct {
	*template.Template
	Severity string
}

func NewSimpleEventHandler(severity string, descTmpl string) SimpleEventHandler {
	return SimpleEventHandler{simpleTemplate(descTmpl), severity}
}

func (tmpl SimpleEventHandler) HandleEvent(results []string, event *Event) error {
	var output bytes.Buffer
	err := tmpl.Execute(&output, results)
	if err != nil {
		return fmt.Errorf("error in custom event handler %s; %s", event.Type, err)
	}
	event.Desc = output.String()
	event.Severity = tmpl.Severity
	return nil
}

// Handles EVENT level records from vcap components
type KnownEventHandler struct {
	*template.Template
}

func NewKnownEventHandler(descTmpl string) KnownEventHandler {
	return KnownEventHandler{template.Must(template.New("").Parse(descTmpl))}
}

func (tmpl KnownEventHandler) HandleEvent(results []string, event *Event) error {
	if len(results) != 2 {
		return fmt.Errorf("did not find a single JSON match; instead found %d", len(results))
	}
	err := json.Unmarshal([]byte(results[1]), &event.Info)
	if err != nil {
		return err
	}
	var output bytes.Buffer
	err = tmpl.Execute(&output, event.Info)
	if err != nil {
		return fmt.Errorf("error in known event handler %s; %s", event.Type, err)
	}
	event.Desc = output.String()
	event.Severity = "INFO"
	return nil
}

// Custom event handler. Usually necessary for parsing json data out
// of log records.
type CustomEventHandler func(results []string, event *Event) error

func (h CustomEventHandler) HandleEvent(results []string, event *Event) error {
	return h(results, event)
}

var templateIndexRe *regexp.Regexp

func simpleTemplate(tmpl string) *template.Template {
	// replace $1 with {{index . 1}} as understood by text/template
	tmpl = templateIndexRe.ReplaceAllString(tmpl, "{{index . $1}}")
	return template.Must(template.New("").Parse(tmpl))
}

func init() {
	templateIndexRe = regexp.MustCompile(`\$(\d+)`)
}
