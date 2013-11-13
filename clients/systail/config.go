package main

import (
	"stackato/server"
)

type Config struct {
	MaxRecordSize int               `json:"max_record_size"`
	LogFiles      map[string]string `json:"log_files"`
}

var c *server.Config

func getConfig() *Config {
	return c.GetConfig().(*Config)
}

func LoadConfig() {
	var err error
	c, err = server.NewConfig("systail", Config{})
	if err != nil {
		fatal("Unable to load systail config; %v", err)
	}
}
