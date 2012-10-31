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

		event := ParseEvent(record["Name"], record["Text"])
		if event != nil {
			event.NodeID = record["NodeID"]
			data, err := json.Marshal(event)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Got event: %+v", event)
			logyardclient.Send(prefix, string(data))
		}

	}
}
