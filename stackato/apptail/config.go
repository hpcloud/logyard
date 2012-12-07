package main

import (
	"github.com/srid/doozerconfig"
	"github.com/srid/log"
	"logyard/stackato"
)

var Config struct {
	MaxRecordSize int    `doozer:"logyard/config/apptail/max_record_size"`
	NatsUri       string `doozer:"cloud_controller/config/mbus"`
}

func LoadConfig() {
	conn, headRev, err := stackato.NewDoozerClient("apptail")
	if err != nil {
		log.Fatal(err)
	}

	key := "/proc/"

	doozerCfg := doozerconfig.New(conn, headRev, &Config, key)
	err = doozerCfg.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Watch for config changes in doozer
	go doozerCfg.Monitor(key+"*", func(change *doozerconfig.Change, err error) {
		if err != nil {
			log.Fatalf("Error processing config change in doozer: %s", err)
			return
		}
		log.Infof("Config changed in doozer; %+v\n", change)
	})
}
