package apptail

import (
	"confdis/go/confdis"
	"github.com/ActiveState/log"
	"stackato/server"
)

var Config struct {
	MaxRecordSize int `json:"max_record_size"`
}

func LoadConfig() {
	conn, headRev, err := server.NewDoozerClient("apptail")
	if err != nil {
		log.Fatal(err)
	}
	server.Init(conn, headRev)
	if _, err = confdis.New(server.Config.CoreIP+":5454", "config:apptail", &Config); err != nil {
		log.Fatal(err)
	}
}
