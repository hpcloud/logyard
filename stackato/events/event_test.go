package main

import (
	"testing"
)

type Case struct {
	Event
	SampleLog string
}

// Cases are a map of process name to raw log record to match
var Cases []Case

func init() {
	Cases = []Case{
		Case{Event{Type: "process_started", Process: "supervisord"},
			"INFO success: memcached_node entered RUNNING state, process has ..."},
		Case{Event{Type: "cc_start", Process: "cloud_controller"},
			`DEBUG -- Sending start message {"droplet":6,"name":"sinatra-env","uris":["sinatra-env.stackato-sf4r.local"],"runtime":"ruby18","framework":"sinatra","sha1":"4b89d4df0815603765b9e3c4864ca909c88564c4","executableFile":"/var/vcap/shared/droplets/droplet_6","executableUri":"http://172.16.145.180:9022/staged_droplets/6/4b89d4df0815603765b9e3c4864ca909c88564c4","version":"4b89d4df0815603765b9e3c4864ca909c88564c4-2","services":[],"limits":{"mem":128,"disk":2048,"fds":256,"sudo":false},"env":[],"group":"s@s.com","index":0,"repos":["deb mirror://mirrors.ubuntu.com/mirrors.txt precise main restricted universe multiverse","deb mirror://mirrors.ubuntu.com/mirrors.txt precise-updates main restricted universe multiverse","deb http://security.ubuntu.com/ubuntu precise-security main universe"]} to DEA 2c4b4d96d82f98f7d6d409ec49edbe44`},
	}
}

func TestSimple(t *testing.T) {
	for _, cas := range Cases {
		event := ParseEvent(cas.Event.Process, cas.SampleLog)
		if event == nil {
			t.Fatalf("did detect event for: %s", cas.SampleLog)
		}
		// we care only about the Type/Process fields; rest of
		// the fields (Description, Info) are not needed to be
		// tested yet.
		if event.Process != cas.Event.Process {
			t.Fatalf("misdetection process %s != %s -- for: %s", event.Process, cas.Event.Process, cas.SampleLog)
		}
		if event.Type != cas.Event.Type {
			t.Fatalf("misdetection type %s != %s -- for: %s", event.Type, cas.Event.Type, cas.SampleLog)
		}

	}
}
