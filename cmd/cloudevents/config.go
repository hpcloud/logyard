package main

import (
	"confdis/go/confdis"
	"github.com/ActiveState/log"
	"logyard"
	"logyard/stackato/events"
	"stackato/server"
)

type Config struct {
	Events map[string]map[string]events.EventParserSpec `json:"events"`
}

var c *confdis.ConfDis

func getConfig() *Config {
	return c.Config.(*Config)
}

func LoadConfig() {
	conn, headRev, err := server.NewDoozerClient("cloud_events")
	if err != nil {
		log.Fatal(err)
	}
	server.Init(conn, headRev)
	if c, err = confdis.New(server.Config.CoreIP+":5454", "config:cloud_events", Config{}); err != nil {
		log.Fatal(err)
	}
	go logyard.MonitorConfig(c)
	log.Info(getConfig().Events)
}
