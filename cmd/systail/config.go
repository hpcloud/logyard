package main

import (
	"github.com/ActiveState/log"
	"stackato/server"
)

type Config struct {
	MaxRecordSize int               `json:"max_record_size"`
	LogFiles      map[string]string `json:"log_files"`
}

var c *server.GroupConfig

func getConfig() *Config {
	return c.Config.(*Config)
}

func LoadConfig() {
	var err error
	c, err = server.NewGroupConfig("systail", Config{})
	if err != nil {
		log.Fatal(err)
	}
}
