package main

import (
	"fmt"
	"github.com/ActiveState/log"
	"github.com/ActiveState/zmqpubsub"
	"logyard"
	"logyard/clients/docker_events"
	"logyard/clients/sieve"
	"stackato/server"
)

var NodeID string

func SendToLogyard(pub *zmqpubsub.Publisher, event *docker_events.Event) {
	log.Infof("Event: %+v", event)
	(&sieve.Event{
		Type:     event.Status,
		Process:  "docker_events",
		Severity: "INFO",
		UnixTime: event.Time,
		NodeID:   NodeID,
		Desc: fmt.Sprintf("%v action for container %v (image: %v)",
			event.Status, event.Id, event.From),
	}).MustPublish(pub)
}

func main() {
	pub := logyard.Broker.NewPublisherMust()
	defer pub.Stop()

	if ch, err := docker_events.Stream(); err != nil {
		log.Fatal(err)
	} else {
		for event := range ch {
			SendToLogyard(pub, event)
		}
	}
}

func init() {
	var err error
	NodeID, err = server.LocalIP()
	if err != nil {
		log.Fatalf("Failed to determine IP addr: %v", err)
	}
}
