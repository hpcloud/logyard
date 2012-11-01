package main

import (
	"encoding/json"
	"fmt"
	"log"
	"logyard"
	"logyard/stackato/events"
	"time"
)

// Make relevant cloud events available in application logs. Heroku style.
func MonitorCloudEvents() {
	// TODO: add more events; will require modifying the log
	// invokation to include the required app id
	filters := []string{
		"event.dea_start",
		"event.stager_start",
	}

	c, err := logyard.NewClientGlobal()
	if err != nil {
		log.Fatal(err)
	}
	ss, err := c.Recv(filters)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Listening for app relevant cloud events...")
	for msg := range ss.Ch {
		fmt.Println(msg.Key, msg.Value)
		var event events.Event
		// TODO: refactor
		err := json.Unmarshal([]byte(msg.Value), &event)
		if err != nil {
			log.Fatal(err) // not expected at all
		}

		switch msg.Key {
		case "event.dea_start":
			// TODO: this could fail; handle it gracefully
			appid := int(event.Info["droplet"].(float64))
			index := int(event.Info["index"].(float64))
			typ := "stackato.dea"
			source := fmt.Sprintf("stackato.dea.%d", index)
			PublishAppLog(c, appid, index, typ, source, &event)
		case "event.stager_start":
			appid := int(event.Info["app_id"].(float64))
			PublishAppLog(c, appid, -1, "stackato.stager", "stackato.stage", &event)
		}

	}

	err = ss.Wait()
	if err != nil {
		log.Fatal(err)
	}
}

func PublishAppLog(client *logyard.Client, appid int, index int, typ string, source string, event *events.Event) {
	m := AppLogMessage{
		Text:          event.Desc,
		LogFilename:   "",
		UnixTime:      event.UnixTime,
		HumanTime:     time.Unix(event.UnixTime, 0).Format("2006-01-02T15:04:05-07:00"), // heroku-format
		InstanceIndex: index,
		InstanceType:  typ,
		Source:        source}
	data, err := json.Marshal(m)
	if err != nil {
		log.Printf("Error encoding %+v into JSON; %s. Skipping this message", m, err)
		return
	}
	key := fmt.Sprintf("apptail.%d", appid)
	err = client.Send(key, string(data))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("sent", key, "to", string(data))
}
