package main

import (
	"github.com/ActiveState/log"
	"logyard/clients/sieve"
	"stackato/server"
)

type Config struct {
	Events map[string]map[string]sieve.EventParserSpec `json:"events"`
}

var c *server.Config

func getConfig() *Config {
	return c.GetConfig().(*Config)
}

func LoadConfig() {
	var err error
	c, err = server.NewConfig("logyard_sieve", Config{})
	if err != nil {
		log.Fatalf("Unable to load logyard_sieve config; %v", err)
	}
	log.Info(getConfig().Events)
}
