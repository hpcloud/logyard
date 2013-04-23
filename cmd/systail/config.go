package main

import (
	"confdis/go/confdis"
	"github.com/ActiveState/log"
	"logyard"
	"stackato/server"
)

type Config struct {
	MaxRecordSize int               `json:"max_record_size"`
	LogFiles      map[string]string `json:"log_files"`
}

var c *confdis.ConfDis

func getConfig() *Config {
	return c.Config.(*Config)
}

func LoadConfig() {
	conn, headRev, err := server.NewDoozerClient("systail")
	if err != nil {
		log.Fatal(err)
	}
	server.Init(conn, headRev)
	if c, err = confdis.New(server.Config.CoreIP+":5454", "config:systail", Config{}); err != nil {
		log.Fatal(err)
	}
	go logyard.MonitorConfig(c)
}
