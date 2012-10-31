package main

import (
	"testing"
)

type Case struct {
	SampleLog string
	Event
}

// Cases are a map of process name to raw log record to match
var Cases map[string][]Case

func init() {
	Cases = map[string][]Case{
		"supervisord": []Case{Case{"INFO success: memcached_node entered RUNNING state, process has ...", Event{Type: "process_started"}}},
		"cloud_controller": []Case{
			Case{`DEBUG -- Sending start message {"droplet":6,"name":"sinatra-env","uris":["sinatra-env.stackato-sf4r.local"],"runtime":"ruby18","framework":"sinatra","sha1":"4b89d4df0815603765b9e3c4864ca909c88564c4","executableFile":"/var/vcap/shared/droplets/droplet_6","executableUri":"http://172.16.145.180:9022/staged_droplets/6/4b89d4df0815603765b9e3c4864ca909c88564c4","version":"4b89d4df0815603765b9e3c4864ca909c88564c4-2","services":[],"limits":{"mem":128,"disk":2048,"fds":256,"sudo":false},"env":[],"group":"s@s.com","index":0,"repos":["deb mirror://mirrors.ubuntu.com/mirrors.txt precise main restricted universe multiverse","deb mirror://mirrors.ubuntu.com/mirrors.txt precise-updates main restricted universe multiverse","deb http://security.ubuntu.com/ubuntu precise-security main universe"]} to DEA 2c4b4d96d82f98f7d6d409ec49edbe44`,
				Event{Type: "cc_start"}}},
	}
}

func TestSimple(t *testing.T) {
	for process, cases := range Cases {
		if detector, ok := eventDetectors[process]; ok {
			for _, cas := range cases {
				event := detector(cas.SampleLog)
				if event == nil {
					t.Fatalf("did not detect %s event for: %s", process, cas.SampleLog)
				}
				// we care only about the Type field; rest of the fields
				// (Description, Info) are not needed to be tested yet.
				if event.Type != cas.Event.Type {
					t.Fatalf("misdetection %s != %s -- for: %s", event.Type, cas.Event.Type, cas.SampleLog)
				}
			}
		} else {
			t.Fatalf("no detector registered for this event: %s", process)
		}
	}
}
