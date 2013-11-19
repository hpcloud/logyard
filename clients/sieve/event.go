package sieve

import (
	"encoding/json"
	"github.com/ActiveState/log"
	"github.com/ActiveState/zmqpubsub"
	"logyard/clients/common"
)

type Event struct {
	Type     string                 `json:"type"`     // Event identifier.
	Desc     string                 `json:"desc"`     // Event description
	Severity string                 `json:"severity"` // Event severity (INFO, WARN, ERROR)
	Info     map[string]interface{} `json:"info"`     // Aribtrary data specific to this event
	Process  string                 `json:"process"`  // The process that generated this event
	common.MessageCommon
}

func (event *Event) MustPublish(pub *zmqpubsub.Publisher) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Fatal(err)
	}
	pub.MustPublish("event."+event.Type, string(data))
}
