package events

import (
	"github.com/ActiveState/log"
)

// TODO: somehow merge this redundant struct with EventParser.
type EventParserSpec struct {
	Substring   string `json:"substring"`
	Re          string `json:"regex"`
	Sample      string `json:"sample"`
	Format      string `json:"format"`
	Severity    string `json:"severity"`
	HandlerType string `json:"handlertype"`
}

type EventParser struct {
	Substring string       // substring unique to this log record for efficient matching
	Re        string       // regex to use for matching
	Sample    string       // sample log record
	Handler   EventHandler // event handler
}

func (p *EventParserSpec) ToEventParser() *EventParser {
	if p.Severity == "" {
		p.Severity = "INFO" // default severity
	}
	e := &EventParser{p.Substring, p.Re, p.Sample, nil}
	switch p.HandlerType {
	case "", "simple":
		e.Handler = NewSimpleEventHandler(p.Severity, p.Format)
	case "json":
		e.Handler = NewJsonEventHandler(p.Severity, p.Format)
	}
	return e
}

// EventParserGroup is a group of event parsers, which group is
// matched in a single attempt, independent of other groups.
type EventParserGroup map[string]*EventParser

type Parser struct {
	tree     map[string]EventParserGroup
	matchers map[string]*MultiRegexpMatcher
}

func NewParser(tree map[string]EventParserGroup) Parser {
	return Parser{tree: tree, matchers: make(map[string]*MultiRegexpMatcher)}
}

func (parser Parser) Build() {
	for group_name, group := range parser.tree {
		matcher := NewMultiRegexpMatcher()
		for event_name, event_parser := range group {
			matcher.MustAdd(event_name, event_parser.Substring, event_parser.Re)
		}
		matcher.Build()
		parser.matchers[group_name] = matcher
	}
}

// DeleteSamples deletes the samples (EventParser.Sample) to free up
// some memory.
func (parser Parser) DeleteSamples() {
	for _, group := range parser.tree {
		for _, event_parser := range group {
			event_parser.Sample = ""
		}
	}
}

// Parser parses the given message under the given group and returns
// the matching event.
func (parser Parser) Parse(group_name string, text string) (*Event, error) {
	group, ok := parser.tree[group_name]
	if !ok {
		return parser.parseStarGroup(group_name, text)
	}
	if matcher, ok := parser.matchers[group_name]; ok {
		if event_type, results := matcher.Match(text); results != nil {
			event := Event{Type: event_type, Process: group_name, Severity: "INFO"}
			if event_parser, ok := group[event_type]; ok {
				err := event_parser.Handler.HandleEvent(results, &event)
				if err != nil {
					return nil, err
				}
				return &event, nil
			}
			panic("not reachable")
		}
		return parser.parseStarGroup(group_name, text)
	}
	panic("not reachable")
}

func (parser Parser) parseStarGroup(orig_group string, text string) (*Event, error) {
	group, ok := parser.tree["__all__"]
	if !ok {
		return nil, nil // no "*" group defined
	}
	matcher := parser.matchers["__all__"]
	if event_type, results := matcher.Match(text); results != nil {
		event := Event{Type: event_type, Process: orig_group, Severity: "INFO"}
		event_parser := group[event_type]
		err := event_parser.Handler.HandleEvent(results, &event)
		if err != nil {
			return nil, err
		}
		return &event, nil
	}
	return nil, nil
}

func NewStackatoParser(spec map[string]map[string]EventParserSpec) Parser {
	parserSpec := map[string]EventParserGroup{}
	for process, d := range spec {
		if _, ok := parserSpec[process]; !ok {
			parserSpec[process] = map[string]*EventParser{}
		}
		for eventName, evt := range d {
			log.Infof("Loading parse spec %s", eventName)
			parserSpec[process][eventName] = evt.ToEventParser()
		}
	}
	parser := NewParser(parserSpec)
	parser.Build()
	return parser
}
