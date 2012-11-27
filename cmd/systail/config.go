package main

import (
	"github.com/srid/doozerconfig"
	"github.com/srid/log2"
	"logyard/stackato"
)

var Config struct {
	MaxRecordSize int `doozer:"max_record_size"`
}

func LoadConfig() {
	conn, headRev, err := stackato.NewDoozerClient("systail")
	if err != nil {
		log2.Fatal(err)
	}

	key := "/proc/logyard/config/systail/"

	doozerCfg := doozerconfig.New(conn, headRev, &Config, key)
	err = doozerCfg.Load()
	if err != nil {
		log2.Fatal(err)
	}

	// Watch for config changes in doozer
	go doozerCfg.Monitor(key+"*", func(change *doozerconfig.Change, err error) {
		if err != nil {
			log2.Fatalf("Error processing config change in doozer: %s", err)
			return
		}
		log2.Infof("Config changed in doozer; %+v\n", change)
	})
}
