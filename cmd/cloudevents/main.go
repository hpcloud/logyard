package main

import (
	"encoding/json"
	"github.com/ActiveState/log"
	"logyard"
	"logyard/stackato/events"
)

// TODO: share it with systail
type SystailRecord struct {
	UnixTime int64
	Text     string
	Name     string
	NodeID   string
}

func main() {
	parser := events.NewStackatoParser()
	parser.DeleteSamples()

	pub, err := logyard.Broker.NewPublisher()
	if err != nil {
		log.Fatal(err)
	}
	defer pub.Stop()
	sub := logyard.Broker.Subscribe("systail")
	defer sub.Stop()

	log.Info("Watching the systail stream on this node")
	for message := range sub.Ch {
		var record SystailRecord
		err := json.Unmarshal([]byte(message.Value), &record)
		if err != nil {
			log.Errorf("failed to parse json: %s; ignoring record: %s",
				err, message.Value)
			continue
		}

		event, err := parser.Parse(record.Name, record.Text)
		if err != nil {
			log.Errorf(
				"failed to parse event from %s: %s -- source: %s", record.Name, err, record.Text)
			continue
		}
		if event != nil {
			event.NodeID = record.NodeID
			event.UnixTime = record.UnixTime
			data, err := json.Marshal(event)
			if err != nil {
				log.Fatal(err)
			}
			err = pub.Publish("event."+event.Type, string(data))
			if err != nil {
				log.Fatal(err)
			}
		}

	}
}
