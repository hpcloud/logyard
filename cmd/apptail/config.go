package main

import (
	"github.com/srid/doozerconfig"
	"log"
	"logyard/stackato"
)

var Config struct {
	MaxRecordSize int `doozer:"max_record_size"`
}

func LoadConfig() {
	conn, headRev, err := stackato.NewDoozerClient("apptail")
	if err != nil {
		log.Fatal(err)
	}

	key := "/proc/logyard/config/apptail/"

	doozerCfg := doozerconfig.New(conn, headRev, &Config, key)
	err = doozerCfg.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Watch for config changes in doozer
	go doozerCfg.Monitor(key+"*", func(name string, value interface{}, err error) {
		if err != nil {
			log.Fatalf("Error processing config change in doozer: %s", err)
			return
		}
		log.Printf("Config changed in doozer; %s=%v\n", name, value)
	})
}
