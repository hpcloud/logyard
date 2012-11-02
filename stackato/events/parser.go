package events

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

func NewStackatoParser() Parser {
	s := NewSimpleEventHandler

	serviceNodeParserGroup := serviceParsers()

	parser := NewParser(map[string]EventParserGroup{
		"supervisord": map[string]*EventParser{
			"process_start": &EventParser{
				Substring: "entered RUNNING state",
				Re:        `(\w+) entered RUNNING`,
				Sample:    `INFO success: memcached_node entered RUNNING state, process has ...`,
				Handler:   s("INFO", "Process '$1' started on a node")},
			"process_stop": &EventParser{
				Substring: "stopped",
				Re:        `stopped: (\w+) \((.+)\)`,
				Sample:    `INFO stopped: mysql_node (terminated by SIGKILL)`,
				Handler:   s("WARNING", "Process '$1' stopped on a node ($2)")},
			"process_exit": &EventParser{
				Substring: "exited",
				Re:        `exited: (\w+) \((.+)\)`,
				Sample:    `INFO exited: dea (exit status 1; not expected)`,
				Handler:   s("FATAL", "Process '$1' crashed on a node ($2)")},
		},
		"kato": map[string]*EventParser{
			"kato_action": &EventParser{
				Substring: "INVOKE",
				Re:        `INVOKE (.+)`,
				Sample:    `[info] (12339) INVOKE kato start`,
				Handler:   s("INFO", "kato action taken on a node: $1"),
			},
		},
		"heath_manager": map[string]*EventParser{
			"hm_analyze": &EventParser{
				Substring: "Analyzed",
				Re:        `Analyzed (\d+) running and (\d+) down apps in (\S+)$`,
				Sample:    `INFO -- Analyzed 3 running and 0 down apps in 95.9ms`,
				Handler:   s("INFO", "HM analyzed $1 running apps and $2 down apps"),
			},
		},
		"cloud_controller": map[string]*EventParser{
			"cc_waiting_for_dea": &EventParser{
				Substring: "No resources available to",
				Re:        `No resources available to start instance (.+)$`,
				Sample:    `WARN -- No resources available to start instance {"droplet":6,"name":"sinatra-env","uris":["sinatra-env.stackato-sf4r.local"],"runtime":"ruby18","framework":"sinatra","sha1":"4b89d4df0815603765b9e3c4864ca909c88564c4","executableFile":"/var/vcap/shared/droplets/droplet_6","executableUri":"http://172.16.145.180:9022/staged_droplets/6/4b89d4df0815603765b9e3c4864ca909c88564c4","version":"4b89d4df0815603765b9e3c4864ca909c88564c4-2","services":[],"limits":{"mem":128,"disk":2048,"fds":256,"sudo":false},"env":[],"group":"s@s.com","index":0,"repos":["deb mirror://mirrors.ubuntu.com/mirrors.txt precise main restricted universe multiverse","deb mirror://mirrors.ubuntu.com/mirrors.txt precise-updates main restricted universe multiverse","deb http://security.ubuntu.com/ubuntu precise-security main universe"]}`,
				Handler: CustomEventHandler(func(results []string, event *Event) error {
					err := json.Unmarshal([]byte(results[1]), &event.Info)
					if err != nil {
						return err
					}
					event.Desc = fmt.Sprintf("No DEA can accept app '%v' of runtime '%v'; retrying...",
						event.Info["name"], event.Info["runtime"])
					event.Severity = "WARNING"
					return nil
				}),
			},
			"cc_start": &EventParser{
				Substring: "START_INSTANCE",
				Re:        `EVENT -- START_INSTANCE (.+)$`,
				Sample:    ` EVENT -- START_INSTANCE {"app_name":"env","app_id":6,"instance":0,"dea_id":"hash"}`,
				Handler: CustomEventHandler(func(results []string, event *Event) error {
					// XXX: doing this merely to extract the appname ...maybe we shouldn't?
					err := json.Unmarshal([]byte(results[1]), &event.Info)
					if err != nil {
						return err
					}
					event.Desc = fmt.Sprintf("Starting application '%v' on DEA %v",
						event.Info["app_name"], event.Info["dea_id"])
					return nil
				}),
			},
		},
		"stager": map[string]*EventParser{
			"stager_start": &EventParser{
				Substring: "START_STAGING",
				Re:        `EVENT -- START_STAGING (.+)$`,
				Sample:    `EVENT -- START_STAGING {"app_id":7,"app_name":"env"}`,
				Handler: CustomEventHandler(func(results []string, event *Event) error {
					err := json.Unmarshal([]byte(results[1]), &event.Info)
					if err != nil {
						return err
					}
					event.Desc = fmt.Sprintf("Staging application '%v'", event.Info["app_name"])
					return nil
				}),
			},
			"stager_end": &EventParser{
				Substring: "completed",
				Re:        `Task\, id\=(\w+) completed`,
				// NOTE: application name is not available in this
				// record. nor does it mention the result status
				// (success/failure).
				Sample:  `INFO -- Task, id=1e117625577284da3dc4f47bb780f0ae completed, result=`,
				Handler: s("INFO", "Completed staging an application; task $1"),
			},
		},
		"dea": map[string]*EventParser{
			"dea_start": &EventParser{
				Substring: "DEA received start message",
				Re:        `DEA received start message: (.+)$`,
				Sample:    `DEBUG -- DEA received start message: {"droplet":6,"name":"sinatra-env","uris":["sinatra-env.stackato-sf4r.local"],"runtime":"ruby18","framework":"sinatra","sha1":"e6d971d2863a5174647580b518b48b26dcf683a6","executableFile":"/var/vcap/shared/droplets/droplet_6","executableUri":"http://172.16.145.180:9022/staged_droplets/6/e6d971d2863a5174647580b518b48b26dcf683a6","version":"e6d971d2863a5174647580b518b48b26dcf683a6-7","services":[],"limits":{"mem":128,"disk":2048,"fds":256,"sudo":false},"env":[],"group":"s@s.com","debug":null,"console":null,"repos":["deb mirror://mirrors.ubuntu.com/mirrors.txt precise main restricted universe multiverse","deb mirror://mirrors.ubuntu.com/mirrors.txt precise-updates main restricted universe multiverse","deb http://security.ubuntu.com/ubuntu precise-security main universe"],"index":0}`,
				Handler: CustomEventHandler(func(results []string, event *Event) error {
					err := json.Unmarshal([]byte(results[1]), &event.Info)
					if err != nil {
						return err
					}
					event.Desc = fmt.Sprintf("Starting application '%v' instance #%v",
						event.Info["name"], event.Info["index"])
					return nil
				}),
			},
			"dea_stop": &EventParser{
				Substring: "STOPPING_INSTANCE",
				Re:        `EVENT -- STOPPING_INSTANCE (.+)$`,
				Sample:    `EVENT -- STOPPING_INSTANCE {"app_id":6,"app_name":"env","instance":0,"dea_id":"deahas"}`,
				Handler: CustomEventHandler(func(results []string, event *Event) error {
					err := json.Unmarshal([]byte(results[1]), &event.Info)
					if err != nil {
						return err
					}
					event.Desc = fmt.Sprintf("Stopping application '%v' on DEA %v",
						event.Info["app_name"], event.Info["dea_id"])
					return nil
				}),
			},
			"dea_ready": &EventParser{
				Substring: "INSTANCE_READY",
				Re:        `EVENT -- INSTANCE_READY (.+)$`,
				Sample:    `EVENT -- INSTANCE_READY {"app_id":6,"app_name":"env","instance":0}`,
				Handler: CustomEventHandler(func(results []string, event *Event) error {
					err := json.Unmarshal([]byte(results[1]), &event.Info)
					if err != nil {
						return err
					}
					event.Desc = fmt.Sprintf("Application '%v' is now running on DEA %v",
						event.Info["app_name"], event.Info["dea_id"])
					return nil
				}),
			},
		},
		// XXX: dynamic way to maintain this list?
		"filesystem_node": serviceNodeParserGroup,
		"mongodb_node":    serviceNodeParserGroup,
		"postgresql_node": serviceNodeParserGroup,
		"redis_node":      serviceNodeParserGroup,
		"memcached_node":  serviceNodeParserGroup,
		"mysql_node":      serviceNodeParserGroup,
		"rabbit_node":     serviceNodeParserGroup,
		// catch all matching
		"*": map[string]*EventParser{
			"vcap_error": &EventParser{
				Substring: "ERROR",
				Re:        `ERROR -- (.+)$`,
				Sample:    `postgresql_gateway - pid=4340 tid=2e99 fid=bad6  ERROR -- Failed fetching handles: Errno::ETIMEDOUT`,
				Handler:   s("ERROR", "$1"),
			},
			"vcap_warning": &EventParser{
				Substring: "WARN",
				Re:        `WARN -- (.+)$`,
				Sample:    `WARN -- Took 18.09s to process ps and du stats`,
				Handler:   s("WARNING", "$1"),
			},
		},
	})

	parser.Build()

	return parser
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
