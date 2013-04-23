package apptail

import (
	"confdis/go/confdis"
	"github.com/ActiveState/log"
	"logyard"
	"stackato/server"
)

type Config struct {
	MaxRecordSize int `json:"max_record_size"`
}

var c *confdis.ConfDis

func GetConfig() *Config {
	return c.Config.(*Config)
}

func LoadConfig() {
	conn, headRev, err := server.NewDoozerClient("apptail")
	if err != nil {
		log.Fatal(err)
	}
	server.Init(conn, headRev)
	if c, err = confdis.New(server.Config.CoreIP+":5454", "config:apptail", Config{}); err != nil {
		log.Fatal(err)
	}
	go logyard.MonitorConfig(c)
}
