package main

import (
	"confdis/go/confdis"
	"github.com/ActiveState/log"
	"logyard/stackato/events"
	"stackato/server"
)

var Config struct {
	Events map[string]map[string]events.EventParser `json:"events"`
}

func LoadConfig() {
	conn, headRev, err := server.NewDoozerClient("cloud_events")
	if err != nil {
		log.Fatal(err)
	}
	server.Init(conn, headRev)
	Config.Events = make(map[string]map[string]events.EventParser)
	if _, err = confdis.New(server.Config.CoreIP+":5454", "config:cloud_events", &Config); err != nil {
		log.Fatal(err)
	}
	log.Info(Config.Events)
}
