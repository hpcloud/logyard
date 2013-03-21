package apptail

import (
	"encoding/json"
	"github.com/ActiveState/log"
	"logyard"
	"logyard/stackato/events"
	"logyard/zmqsub"
	"time"
)

// Make relevant cloud events available in application logs. Heroku style.
func MonitorCloudEvents(nodeid string) {
	// TODO: add more events; will require modifying the log
	// invokation to include the required app id
	sub := logyard.Broker.Subscribe(
		"event.dea_start",
		"event.dea_ready",
		"event.dea_stop",
		"event.stager_start",
		"event.stager_end",
		"event.cc_app_update",
	)
	defer sub.Stop()

	pub, err := logyard.Broker.NewPublisher()
	if err != nil {
		log.Fatal(err)
	}
	defer pub.Stop()

	log.Info("Listening for app relevant cloud events...")
	for msg := range sub.Ch {
		var event events.Event
		err := json.Unmarshal([]byte(msg.Value), &event)
		if err != nil {
			log.Fatal(err) // not expected at all
		}

		switch msg.Key {
		case "event.dea_start", "event.dea_ready", "event.dea_stop":
			appid := int(event.Info["app_id"].(float64))
			name := event.Info["app_name"].(string)
			index := int(event.Info["instance"].(float64))
			source := "stackato.dea"
			PublishAppLog(pub, appid, name, index, source, nodeid, &event)
		case "event.stager_start", "event.stager_end":
			appid := int(event.Info["app_id"].(float64))
			name := event.Info["app_name"].(string)
			PublishAppLog(pub, appid, name, -1, "stackato.stager", nodeid, &event)
		case "event.cc_app_update":
			appid := int(event.Info["app_id"].(float64))
			name := event.Info["app_name"].(string)
			PublishAppLog(pub, appid, name, -1, "stackato.controller", nodeid, &event)
		}
	}
	log.Warn("Finished listening for app relevant cloud events.")

	err = sub.Wait()
	if err != nil {
		log.Fatal(err)
	}
}

func PublishAppLog(
	pub *zmqsub.Publisher, app_id int, app_name string,
	index int, source string, nodeid string, event *events.Event) {

	err := (&AppLogMessage{
		Text:          event.Desc,
		LogFilename:   "",
		UnixTime:      event.UnixTime,
		HumanTime:     time.Unix(event.UnixTime, 0).Format("2006-01-02T15:04:05-07:00"), // heroku-format
		Source:        source,
		InstanceIndex: index,
		AppID:         app_id,
		AppName:       app_name,
		NodeID:        nodeid,
	}).Publish(pub, true)

	if err != nil {
		log.Fatal(err)
	}
}
