package main

import (
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
			// we care only about the Type/Process fields; rest of
			// the fields (Description, Info) are not needed to be
			// tested yet.
			if event.Process != process {
				t.Fatalf("misdetection process %s != %s -- for: %s", event.Process, process, event_parser.Sample)
			}
			if event.Type != event_type {
				t.Fatalf("misdetection type %s != %s -- for: %s", event.Type, event_type, event_parser.Sample)
			}
		}
	}
}
