package events

import (
	"bytes"
	"fmt"
	"regexp"
	"text/template"
)

type Event struct {
	Type     string // what type of event?
	Desc     string // description of this event to be shown as-is to humans
	Severity string
	Info     map[string]interface{} // event-specific information as json
	Process  string                 // which process generated this event?
	UnixTime int64
	NodeID   string // from which node did this event appear?
}

// A simple event handler filling in only the description field.
// Suitable for most events.
type SimpleEventHandler struct {
	*template.Template
	Severity string
}

var templateIndexRe *regexp.Regexp

func NewSimpleEventHandler(severity string, descTmpl string) SimpleEventHandler {
	// replace $1 with {{index . 1}} as understood by text/template
	descTmpl = templateIndexRe.ReplaceAllString(descTmpl, "{{index . $1}}")

	t := template.Must(template.New("").Parse(descTmpl))
	return SimpleEventHandler{t, severity}
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

// Custom event handler. Usually necessary for parsing json data out
// of log records.
type CustomEventHandler func(results []string, event *Event) error

func (h CustomEventHandler) HandleEvent(results []string, event *Event) error {
	return h(results, event)
}

type EventHandler interface {
	HandleEvent(results []string, event *Event) error
}

func init() {
	templateIndexRe = regexp.MustCompile(`\$(\d+)`)
}
