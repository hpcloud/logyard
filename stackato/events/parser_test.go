package main

import (
	"fmt"
	"testing"
)

func TestSampleLogs(t *testing.T) {
	for process, event_types := range parser.tree {
		for event_type, event_parser := range event_types {
			event, err := parser.Parse(process, event_parser.Sample)
			if err != nil {
				t.Fatal(err)
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
			// displaying the desc to the user.
			fmt.Printf("<< %19s >> -- %s\n", event.Type, event.Desc)
		}
	}
}
