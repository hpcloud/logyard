package main

import (
	"encoding/json"
	"fmt"
	"github.com/ActiveState/log"
	"github.com/ActiveState/zmqpubsub"
	"io"
	"logyard"
	"logyard/clients/sieve"
	"net/http"
	"stackato/server"
)

var NodeID string

type Event struct {
	Id     string `json:"id"`
	Status string `json:"status"`
	From   string `json:"from"`
	Time   int64  `json:"time"`
}

func SendToLogyard(pub *zmqpubsub.Publisher, event *Event) {
	log.Infof("Event: %+v", event)
	(&sieve.Event{
		Type:     event.Status,
		Process:  "docker_events",
		Severity: "INFO",
		UnixTime: event.Time,
		NodeID:   NodeID,
		Desc: fmt.Sprintf("%v status for container %v (image: %v)",
			event.Status, event.Id, event.From),
	}).MustPublish(pub)
}

func main() {
	pub := logyard.Broker.NewPublisherMust()
	defer pub.Stop()

	c := http.Client{}
	res, err := c.Get("http://localhost:4243/events")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	// Read the streaming json from the events endpoint
	// http://docs.docker.io/en/latest/api/docker_remote_api_v1.3/#monitor-docker-s-events
	d := json.NewDecoder(res.Body)
	for {
		var event Event
		if err := d.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}
		SendToLogyard(pub, &event)
	}
}

func init() {
	var err error
	NodeID, err = server.LocalIP()
	if err != nil {
		log.Fatalf("Failed to determine IP addr: %v", err)
	}
}
