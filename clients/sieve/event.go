package sieve

import (
	"encoding/json"
	"github.com/ActiveState/log"
	"github.com/ActiveState/zmqpubsub"
	"logyard/clients/common"
)

type Event struct {
	Type     string // what type of event?
	Desc     string // description of this event to be shown as-is to humans
	Severity string
	Info     map[string]interface{} // event-specific information as json
	Process  string                 // which process generated this event?
	common.MessageCommon
}

func (event *Event) MustPublish(pub *zmqpubsub.Publisher) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Fatal(err)
	}
	pub.MustPublish("event."+event.Type, string(data))
}
