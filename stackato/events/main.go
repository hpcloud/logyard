package main

import (
	"encoding/json"
	"log"
	"logyard"
)

// TODO: share it with systail
type SystailRecord struct {
	UnixTime int64
	Text     string
	Name     string
	NodeID   string
}

func main() {
	logyardclient, err := logyard.NewClientGlobal()
	if err != nil {
		log.Fatal(err)
	}

	sub, err := logyardclient.Recv([]string{"systail"})
	if err != nil {
		log.Fatal(err)
	}
	for message := range sub.Ch {
		var record SystailRecord
		err := json.Unmarshal([]byte(message.Value), &record)
		if err != nil {
			log.Printf("Error: failed to parse json: %s; ignoring record: %s",
				err, message.Value)
			continue
		}
		prefix := "event." + record.NodeID

		event, err := parser.Parse(record.Name, record.Text)
		if err != nil {
			log.Printf("Error parsing an event: %s", err)
			continue
		}
		if event != nil {
			event.NodeID = record.NodeID
			event.UnixTime = record.UnixTime
			data, err := json.Marshal(event)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Got event: %+v", event)
			logyardclient.Send(prefix, string(data))
		}

	}
}
