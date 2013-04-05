package main

import (
	"github.com/ActiveState/log"
	"logyard/util/confdis"
	"stackato/server"
)

var Config struct {
	MaxRecordSize int               `json:"max_record_size"`
	LogFiles      map[string]string `json:"log_files"`
}

func LoadConfig() {
	conn, headRev, err := server.NewDoozerClient("systail")
	if err != nil {
		log.Fatal(err)
	}
	server.Init(conn, headRev)
	Config.LogFiles = make(map[string]string)
	if _, err = confdis.New(server.Config.CoreIP+":5454", "config:systail", &Config); err != nil {
		log.Fatal(err)
	}
}
