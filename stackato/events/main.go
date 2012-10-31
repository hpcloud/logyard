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
		// TODO: use apptail's struct to unmarshal
		record := map[string]interface{}{}
		json.Unmarshal([]byte(message.Value), &record)
		prefix := ("event." + record["NodeID"].(string))

		event := ParseEvent(record["Name"].(string), record["Text"].(string))
		if event != nil {
			event.NodeID = record["NodeID"].(string)
			event.UnixTime = int64(record["UnixTime"].(float64))
			data, err := json.Marshal(event)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Got event: %+v", event)
			logyardclient.Send(prefix, string(data))
		}

	}
}
