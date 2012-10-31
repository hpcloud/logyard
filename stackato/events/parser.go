package main

import (
	"encoding/json"
	"fmt"
)

type EventParser struct {
	Substring string       // substring unique to this log record for efficient matching
	Re        string       // regex to use for matching
	Sample    string       // sample log record
	Handler   EventHandler // event handler
}

// EventParserGroup is a group of event parsers, which group is
// matched in a single attempt, independent of other groups.
type EventParserGroup map[string]*EventParser

type Parser struct {
	tree     map[string]EventParserGroup
	matchers map[string]*MultiRegexpMatcher
}

var parser Parser

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
		return nil, nil // this group not handled
	}
	if matcher, ok := parser.matchers[group_name]; ok {
		if event_type, results := matcher.Match(text); results != nil {
			event := Event{Type: event_type, Process: group_name}
			if event_parser := group[event_type]; ok {
				err := event_parser.Handler.HandleEvent(results, &event)
				if err != nil {
					return nil, err
				}
				return &event, nil
			}
			panic("not reachable")
		}
		return nil, nil // nothing matched
	}
	panic("not reachable")
}

func init() {
	s := NewSimpleEventHandler
	parser = NewParser(map[string]EventParserGroup{
		"supervisord": map[string]*EventParser{
			"process_start": &EventParser{
				Substring: "entered RUNNING state",
				Re:        `(\w) entered RUNNING`,
				Sample:    `INFO success: memcached_node entered RUNNING state, process has ...`,
				Handler:   s("Process '$1' started on a node")},
			"process_stop": &EventParser{
				Substring: "stopped",
				Re:        `stopped: (\w+) \((.+)\)`,
				Sample:    `INFO stopped: mysql_node (terminated by SIGKILL)`,
				Handler:   s("Process '$1' stopped on a node ($2)")},
			"process_exit": &EventParser{
				Substring: "exited",
				Re:        `exited: (\w+) \((.+)\)`,
				Sample:    `INFO exited: dea (exit status 1; not expected)`,
				Handler:   s("Process '$1' crashed on a node ($2)")},
		},
		"kato": map[string]*EventParser{
			"kato_action": &EventParser{
				Substring: "INVOKE",
				Re:        `INVOKE (.+)`,
				Sample:    `[info] (12339) INVOKE kato start`,
				Handler:   s("kato action taken on a node: $1"),
			},
		},
		"cloud_controller": map[string]*EventParser{
			"cc_start": &EventParser{
				Substring: "Sending start message",
				Re:        `Sending start message (.+) to DEA (\w+)$`,
				Sample:    `DEBUG -- Sending start message {"droplet":6,"name":"sinatra-env","uris":["sinatra-env.stackato-sf4r.local"],"runtime":"ruby18","framework":"sinatra","sha1":"4b89d4df0815603765b9e3c4864ca909c88564c4","executableFile":"/var/vcap/shared/droplets/droplet_6","executableUri":"http://172.16.145.180:9022/staged_droplets/6/4b89d4df0815603765b9e3c4864ca909c88564c4","version":"4b89d4df0815603765b9e3c4864ca909c88564c4-2","services":[],"limits":{"mem":128,"disk":2048,"fds":256,"sudo":false},"env":[],"group":"s@s.com","index":0,"repos":["deb mirror://mirrors.ubuntu.com/mirrors.txt precise main restricted universe multiverse","deb mirror://mirrors.ubuntu.com/mirrors.txt precise-updates main restricted universe multiverse","deb http://security.ubuntu.com/ubuntu precise-security main universe"]} to DEA 2c4b4d96d82f98f7d6d409ec49edbe44`,
				Handler: CustomEventHandler(func(results []string, event *Event) error {
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
			},
		},
	})

	parser.Build()
}
