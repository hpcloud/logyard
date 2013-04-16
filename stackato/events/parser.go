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

func (p *EventParserSpec) Create() *EventParser {
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
// some memory..
func (parser Parser) DeleteSamples() {
	for _, group := range parser.tree {
		for _, event_parser := range group {
			event_parser.Sample = ""
		}
	}
}

// Parser parses the given message under the given group and returns
// the matching event
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
	group, ok := parser.tree["*"]
	if !ok {
		return nil, nil // no "*" group defined
	}
	matcher := parser.matchers["*"]
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
	parserSpec := builtinSpec()
	for process, d := range spec {
		if _, ok := parserSpec[process]; !ok {
			parserSpec[process] = map[string]*EventParser{}
		}
		for eventName, evt := range d {
			e := evt.Create()
			log.Infof("Loading parse spec [%s]: %+v", eventName, e)
			parserSpec[process][eventName] = e
		}
	}
	parser := NewParser(parserSpec)
	parser.Build()
	return parser
}

func builtinSpec() map[string]EventParserGroup {
	serviceNodeParserGroup := serviceParsers()
	return map[string]EventParserGroup{
		// Xxx: dynamic way to maintain this list?
		"filesystem_node": serviceNodeParserGroup,
		"mongodb_node":    serviceNodeParserGroup,
		"postgresql_node": serviceNodeParserGroup,
		"redis_node":      serviceNodeParserGroup,
		"memcached_node":  serviceNodeParserGroup,
		"mysql_node":      serviceNodeParserGroup,
		"rabbit_node":     serviceNodeParserGroup,
		// catch all matching
		"*": map[string]*EventParser{
			// Note: the "ERROR --" style of log prefix originates
			// from vcap logger. conventionally we try to use the same
			// prefix for non-vcap components. for eg., logyard itself
			// uses the same prefix style.
			"error": &EventParser{
				Substring: "ERROR",
				Re:        `ERROR -- (.+)$`,
				Sample:    `postgresql_gateway - pid=4340 tid=2e99 fid=bad6  ERROR -- Failed fetching handles: Errno::ETIMEDOUT`,
				Handler:   NewSimpleEventHandler("ERROR", "$1"),
			},
			"warning": &EventParser{
				Substring: "WARN",
				Re:        `WARN -- (.+)$`,
				Sample:    `WARN -- Took 18.09s to process ps and du stats`,
				Handler:   NewSimpleEventHandler("WARNING", "$1"),
			},
		},
	}
}

func serviceParsers() map[string]*EventParser {
	return map[string]*EventParser{
		"service_provision": &EventParser{
			Substring: "Successfully provisioned service",
			Re:        `^\[[^\]]+\] (\w+) .+ Successfully provisioned service for request`,
			Sample:    `[2012-11-01 07:30:51.290253] memcached_node_1 - pid=23282 tid=d0cf fid=5280 DEBUG -- MaaS-Node: Successfully provisioned service for request {"plan":"free"}: {:credentials=>{"hostname"=>"192.168.203.197", "host"=>"192.168.203.197", "port"=>11000, "user"=>"cc06b88a-aa63-45f2-82d8-e9ab06f6a3cf", "password"=>"7ce87b70-1ed8-4c12-86ab-0e1c237f6853", "name"=>"20017185-bfb3-4b5a-b9b1-3add745e6552", "node_id"=>"memcached_node_1"}}`,
			Handler:   NewSimpleEventHandler("INFO", "Provisioned a new service on $1"),
		},
	}
}
