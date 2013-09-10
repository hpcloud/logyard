package apptail

import (
	"encoding/json"
	"fmt"
	"github.com/ActiveState/log"
	"logyard"
	"logyard/stackato/events"
	"logyard/util/pubsub"
	"time"
)

// Make relevant cloud events available in application logs. Heroku style.
func MonitorCloudEvents(nodeid string) {
	// TODO: add more events; will require modifying the log
	// invocation to include the required app_id/app_name/group
	sub := logyard.Broker.Subscribe(
		"event.dea_start",
		"event.dea_ready",
		"event.dea_stop",
		"event.stager_start",
		"event.stager_end",
		"event.cc_app_update",
	)
	defer sub.Stop()

	pub := logyard.Broker.NewPublisherMust()
	defer pub.Stop()

	log.Info("Listening for app relevant cloud events...")
	for msg := range sub.Ch {
		var event events.Event
		err := json.Unmarshal([]byte(msg.Value), &event)
		if err != nil {
			log.Fatal(err) // not expected at all
		}

		guid, name, space, err := extractAppInfo(event)
		if err != nil {
			log.Warn(err)
			continue
		}

		switch msg.Key {
		case "event.dea_start", "event.dea_ready", "event.dea_stop":
			index := int(event.Info["instance"].(float64))
			source := "stackato.dea"
			PublishAppLog(pub, guid, name, space, index, source, nodeid, &event)
		case "event.stager_start", "event.stager_end":
			PublishAppLog(pub, guid, name, space, -1, "stackato.stager", nodeid, &event)
		case "event.cc_app_update":
			PublishAppLog(pub, guid, name, space, -1, "stackato.controller", nodeid, &event)
		}
	}
	log.Warn("Finished listening for app relevant cloud events.")

	err := sub.Wait()
	if err != nil {
		log.Fatal(err)
	}
}

func extractAppInfo(e events.Event) (guid, name, space string, err error) {
	var ok bool

	guid, ok = e.Info["app_guid"].(string)
	if !ok {
		err = fmt.Errorf("app_guid field missing in event '%s' from %s/%s'",
			e.Type, e.NodeID, e.Process)
		return
	}

	name, ok = e.Info["app_name"].(string)
	if !ok {
		err = fmt.Errorf("app_name field missing in event '%s' from %s/%s'",
			e.Type, e.NodeID, e.Process)
		return
	}

	space, ok = e.Info["space"].(string)
	if !ok {
		err = fmt.Errorf("space field missing in event '%s' from %s/%s",
			e.Type, e.NodeID, e.Process)
	}
	return
}

func PublishAppLog(
	pub *pubsub.Publisher,
	app_guid string, app_name string, space string,
	index int, source string, nodeid string, event *events.Event) {

	err := (&AppLogMessage{
		Text:          event.Desc,
		LogFilename:   "",
		UnixTime:      event.UnixTime,
		HumanTime:     time.Unix(event.UnixTime, 0).Format("2006-01-02T15:04:05-07:00"), // heroku-format
		Source:        source,
		InstanceIndex: index,
		AppGUID:       app_guid,
		AppName:       app_name,
		AppSpace:      space,
		NodeID:        nodeid,
	}).Publish(pub, true)

	if err != nil {
		log.Fatal(err)
	}
}
