package sieve

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"testing"
)

var config struct {
	Events map[string]map[string]EventParserSpec `json:"events"`
}

func TestSampleLogs(t *testing.T) {
	parser := NewStackatoParser(config.Events)
	for process_pat, event_types := range parser.tree {
		for event_type, event_parser := range event_types {
			process := process_pat
			if process_pat == "*" {
				process = "foo"
			}
			event, err := parser.Parse(process, event_parser.Sample)
			if err != nil {
				t.Fatalf("parse error (%s) for %s", err, event_parser.Sample)
			}
			if event == nil {
				t.Fatalf("no event detected for: %s", event_parser.Sample)
			}
			if event.Process != process {
				t.Fatalf("misdetection process %s != %s -- for: %s", event.Process, process, event_parser.Sample)
			}
			if event.Type != event_type {
				t.Fatalf("misdetection type %s != %s -- for: %s", event.Type, event_type, event_parser.Sample)
			}
			// TODO: we should test the Desc field as well. meanwhile,
			// displaying the Desc field to the user.
			fmt.Printf("<< %19s | %16s >> -- [%8s] %s\n", event.Type, event.Process, event.Severity, event.Desc)
		}
	}
}

func init() {
	// XXX: I'm uncertain as to the simplest way to load
	// sieve.yml as JSON and json.Unmarshal into config.Events
	// .. all in Go. So, for now, I'll use ruby to do the JSON
	// converstion.
	fmt.Println("Loading etc/sieve.yml into test config struct")
	if output, err := exec.Command(
		"/usr/bin/ruby", "-ryaml", "-rjson", "-e",
		"puts YAML.load_file('../../etc/sieve.yml').to_json",
	).CombinedOutput(); err != nil {
		panic(fmt.Sprintf("Failed to run ruby: %v (output: %s)", err, output))
	} else {
		if err := json.Unmarshal(output, &config); err != nil {
			panic(err)
		}
	}

}
