package sieve

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

// SimpleEventHandler assigns event description formatted based on regex groups
type SimpleEventHandler struct {
	*template.Template
	Severity string
}

func NewSimpleEventHandler(severity string, descTmpl string) SimpleEventHandler {
	return SimpleEventHandler{simpleTemplate(descTmpl), severity}
}

func (handler SimpleEventHandler) HandleEvent(results []string, event *Event) error {
	var output bytes.Buffer
	err := handler.Execute(&output, results)
	if err != nil {
		return fmt.Errorf("error in custom event handler %s; %s", event.Type, err)
	}
	event.Desc = output.String()
	event.Severity = handler.Severity
	return nil
}

// SimpleEventHandler assigns event description formatted based on
// fields of JSON extracted from the first and only regex match group
type JsonEventHandler struct {
	*template.Template
	Severity string
}

func NewJsonEventHandler(severity string, descTmpl string) JsonEventHandler {
	return JsonEventHandler{template.Must(template.New("").Parse(descTmpl)), severity}
}

func (handler JsonEventHandler) HandleEvent(results []string, event *Event) error {
	if len(results) < 1 {
		return fmt.Errorf("no regex matches found")
	}
	err := json.Unmarshal([]byte(results[1]), &event.Info)
	if err != nil {
		return err
	}
	var output bytes.Buffer
	err = handler.Execute(&output, event.Info)
	if err != nil {
		return fmt.Errorf("error in known event handler %s; %s", event.Type, err)
	}
	event.Desc = output.String()
	event.Severity = handler.Severity
	return nil
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
