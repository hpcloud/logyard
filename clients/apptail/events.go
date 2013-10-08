package apptail

import (
	"encoding/json"
	"github.com/ActiveState/log"
	"github.com/ActiveState/zmqpubsub"
	"logyard"
	"logyard/clients/sieve"
	"time"
)

type App struct {
	GUID  string `json:"guid"`
	Space string `json:"space_guid"`
	Name  string `json:"name"`
}

type TimelineEvent struct {
	App App `json:"app"`
}

// Make relevant cloud events available in application logs. Heroku style.
func MonitorCloudEvents(nodeid string) {
	sub := logyard.Broker.Subscribe("event.timeline")
	defer sub.Stop()

	pub := logyard.Broker.NewPublisherMust()
	defer pub.Stop()

	log.Info("Listening for app relevant cloud events...")
	for msg := range sub.Ch {
		var event sieve.Event

		err := json.Unmarshal([]byte(msg.Value), &event)
		if err != nil {
			log.Fatal(err) // not expected at all
		}

		// Re-parse the event json record into a TimelineEvent structure.
		var t TimelineEvent
		if data, err := json.Marshal(event.Info); err != nil {
			log.Fatal(err)
		} else {
			err = json.Unmarshal(data, &t)
			if err != nil {
				log.Fatalf("Invalid timeline event: %v", err)
			}
		}

		// TODO: add an instance_index properly to the timeline event
		index := -1 // FIXME: find the instance index; use -1 only for non-app instances (eg: staging)

		source := "stackato." + event.Process

		PublishAppLog(pub, t, index, source, nodeid, &event)
	}
	log.Warn("Finished listening for app relevant cloud events.")

	err := sub.Wait()
	if err != nil {
		log.Fatal(err)
	}
}

func PublishAppLog(
	pub *zmqpubsub.Publisher,
	t TimelineEvent,
	index int, source string, nodeid string, event *sieve.Event) {

	err := (&AppLogMessage{
		Text:          event.Desc,
		LogFilename:   "",
		UnixTime:      event.UnixTime,
		HumanTime:     time.Unix(event.UnixTime, 0).Format("2006-01-02T15:04:05-07:00"), // heroku-format
		Source:        source,
		InstanceIndex: index,
		AppGUID:       t.App.GUID,
		AppName:       t.App.Name,
		AppSpace:      t.App.Space,
		NodeID:        nodeid,
	}).Publish(pub, true)

	if err != nil {
		log.Fatal(err)
	}
}
