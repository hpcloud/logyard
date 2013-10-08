package sieve

import (
	"encoding/json"
	"github.com/ActiveState/log"
	"github.com/ActiveState/zmqpubsub"
)

type Event struct {
	Type     string // what type of event?
	Desc     string // description of this event to be shown as-is to humans
	Severity string
	Info     map[string]interface{} // event-specific information as json
	Process  string                 // which process generated this event?
	UnixTime int64
	NodeID   string // from which node did this event appear?
}

func (event *Event) MustPublish(pub *zmqpubsub.Publisher) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Fatal(err)
	}
	pub.MustPublish("event."+event.Type, string(data))
}
