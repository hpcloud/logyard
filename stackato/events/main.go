package main

import (
	"encoding/json"
	"log"
	"logyard"
	"strings"
)

type Event struct {
	Process     string // which process generated this event?
	NodeID      string // from which node did this event appear?
	Description string // description of this event to be shown as-is to humans
	Info        string // event-specific information as json
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
		record := map[string]string{}
		json.Unmarshal([]byte(message.Value), &record)
		prefix := ("event." + record["NodeID"])

		// XXX: get something working first; then clean up and optimize
		switch record["Name"] {
		case "supervisord":
			if strings.Contains(record["Text"], "entered RUNNING state") {
				event := Event{
					Process:     record["Name"],
					NodeID:      record["NodeID"],
					Description: "Something entered running state",
					Info:        ""}
				data, err := json.Marshal(event)
				if err != nil {
					log.Fatal(err)
				}
				log.Printf("Got event: %+v", event)
				logyardclient.Send(prefix, string(data))
			}
		default:
		}
	}
}
