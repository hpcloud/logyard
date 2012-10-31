package main

import (
	"encoding/json"
	"log"
	"logyard"
)

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
		if detector, ok := eventDetectors[record["Name"]]; ok {
			event := detector(record["Text"])
			if event != nil {
				event.NodeID = record["NodeID"]
				event.Process = record["Name"]
				data, err := json.Marshal(event)
				if err != nil {
					log.Fatal(err)
				}
				log.Printf("Got event: %+v", event)
				logyardclient.Send(prefix, string(data))

			}
		}

	}
}
