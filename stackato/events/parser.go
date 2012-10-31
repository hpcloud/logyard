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
		return nil, nil // this group not handled
	}
	if matcher, ok := parser.matchers[group_name]; ok {
		if event_type, results := matcher.Match(text); results != nil {
			event := Event{Type: event_type, Process: group_name, Severity: "INFO"}
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

func NewStackatoParser() Parser {
	s := NewSimpleEventHandler
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
					event.Info = results[1]
					var info map[string]interface{}
					err := json.Unmarshal([]byte(event.Info), &info)
					if err != nil {
						return err
					}
					event.Desc = fmt.Sprintf("No DEA can accept app '%v' of runtime '%v'; retrying...",
						info["name"], info["runtime"])
					event.Severity = "WARNING"
					return nil
				}),
			},
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
						return err
					}
					event.Desc = fmt.Sprintf("Starting application '%v'", info["name"])
					return nil
				}),
			},
		},
		"stager": map[string]*EventParser{
			"stager_start": &EventParser{
				Substring: "Decoding task",
				Re:        `Decoding task \'(.+)\'$`,
				Sample:    `DEBUG -- Decoding task '{"app_id":8,"properties":{"services":[{"label":"postgresql-9.1","tags":["postgresql","postgresql-9.1","relational"],"name":"postgresql-gtd","credentials":{"name":"d2ca64ddb68d0433b83d876f105659696","host":"172.16.145.180","hostname":"172.16.145.180","port":5432,"user":"u615ee9e7b64d40038e1d9131b3b7e924","username":"u615ee9e7b64d40038e1d9131b3b7e924","password":"pc871d5fef45f4ec29503caaff615b9fc"},"options":{},"plan":"free","plan_option":null}],"framework":"python","runtime":"python27","resources":{"memory":128,"disk":2048,"fds":256,"sudo":false},"environment":["DJANGO_SETTINGS_MODULE=settings"],"uris":["gtd2.stackato-sf4r.local"],"repos":["deb mirror://mirrors.ubuntu.com/mirrors.txt precise main restricted universe multiverse","deb mirror://mirrors.ubuntu.com/mirrors.txt precise-updates main restricted universe multiverse","deb http://security.ubuntu.com/ubuntu precise-security main universe"],"appname":"gtd2","meta":{"debug":null,"console":null}},"download_uri":"http://172.16.145.180:9022/staging/app/8","upload_uri":"http://172.16.145.180:9022/staging/droplet/8/d2fa46c59f6463842bae0480214541cf","notify_subj":"cc.staging.5542c73f6e2f3fed8abca220f94da9a1"}'`,
				Handler: CustomEventHandler(func(results []string, event *Event) error {
					event.Info = results[1]
					var info map[string]interface{}
					err := json.Unmarshal([]byte(event.Info), &info)
					if err != nil {
						return err
					}
					if props, ok := info["properties"].(map[string]interface{}); ok {
						event.Desc = fmt.Sprintf("Staging application '%v'", props["appname"])
						return nil
					}
					return fmt.Errorf("info['properties'] is not a map")
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
					event.Info = results[1]
					var info map[string]interface{}
					err := json.Unmarshal([]byte(event.Info), &info)
					if err != nil {
						return err
					}
					event.Desc = fmt.Sprintf("Starting application '%v' instance #%v", info["name"], info["index"])
					return nil
				}),
			},
			"dea_stop": &EventParser{
				Substring: "Stopping instance",
				Re:        `Stopping instance \(name=(\S+).+instance=(\w+)`,
				Sample:    `INFO -- Stopping instance (name=gtd app_id=5 instance=db82a00d5aa9ce968616b34e8f99109b index=0)`,
				Handler:   s("INFO", "Stopping an instance of '$1' ($2)"),
			},
			"dea_ready": &EventParser{
				Substring: "ready for connections",
				Re:        `Instance \(name=(\S+).+instance=(\w+).+is ready for connections`,
				Sample:    `INFO -- Instance (name=gtd2 app_id=8 instance=be34dd00d7a53801a38a87105dc332e6 index=0) is ready for connections, notifying system of statu`,
				Handler:   s("INFO", "Application '$1' instance '$2' is now running"),
			},
		},
	})

	parser.Build()

	return parser
}
