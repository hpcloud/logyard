package main

import (
	"github.com/ActiveState/log"
	"logyard/stackato/events"
	"stackato/server"
)

type Config struct {
	Events map[string]map[string]events.EventParserSpec `json:"events"`
}

var c *server.GroupConfig

func getConfig() *Config {
	return c.Config.(*Config)
}

func LoadConfig() {
	var err error
	c, err = server.NewGroupConfig("cloud_events", Config{})
	if err != nil {
		log.Fatal(err)
	}
	log.Info(getConfig().Events)
}
