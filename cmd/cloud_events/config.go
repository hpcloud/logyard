package main

import (
	"github.com/ActiveState/log"
	"logyard/clients/events"
	"stackato/server"
)

type Config struct {
	Events map[string]map[string]events.EventParserSpec `json:"events"`
}

var c *server.Config

func getConfig() *Config {
	return c.GetConfig().(*Config)
}

func LoadConfig() {
	var err error
	c, err = server.NewConfig("cloud_events", Config{})
	if err != nil {
		log.Fatal(err)
	}
	log.Info(getConfig().Events)
}
