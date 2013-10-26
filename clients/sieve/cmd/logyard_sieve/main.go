package main

import (
	"encoding/json"
	"github.com/ActiveState/log"
	"logyard"
	"logyard/clients/messagecommon"
	"logyard/clients/sieve"
	"time"
)

// TODO: share it with systail
type SystailRecord struct {
	UnixTime int64
	Text     string
	Name     string
	NodeID   string
}

func main() {
	LoadConfig()

	parser := sieve.NewStackatoParser(getConfig().Events)
	parser.DeleteSamples()

	pub := logyard.Broker.NewPublisherMust()
	defer pub.Stop()
	sub := logyard.Broker.Subscribe("systail")
	defer sub.Stop()

	log.Info("Watching the systail stream on this node")
	for message := range sub.Ch {
		var record SystailRecord
		err := json.Unmarshal([]byte(message.Value), &record)
		if err != nil {
			log.Warnf("failed to parse json: %s; ignoring record: %s",
				err, message.Value)
			continue
		}

		event, err := parser.Parse(record.Name, record.Text)
		if err != nil {
			log.Warnf(
				"failed to parse event from %s: %s -- source: %s", record.Name, err, record.Text)
			continue
		}
		if event != nil {
			event.MessageCommon = messagecommon.New(event.Desc, time.Unix(record.UnixTime, 0), record.NodeID)
			event.MustPublish(pub)
		}

	}
}
