package main

import (
	"github.com/ActiveState/doozerconfig"
	"github.com/ActiveState/log"
	"stackato/server"
)

var Config struct {
	MaxRecordSize int               `doozer:"max_record_size"`
	LogFiles      map[string]string `doozer:"log_files"`
}

func LoadConfig() {
	conn, headRev, err := server.NewDoozerClient("systail")
	if err != nil {
		Cleanup()
		log.Fatal(err)
	}

	key := "/proc/logyard/config/systail/"

	doozerCfg := doozerconfig.New(conn, headRev, &Config, key)
	Config.LogFiles = make(map[string]string)
	err = doozerCfg.Load()
	if err != nil {
		Cleanup()
		log.Fatal(err)
	}

	// Watch for config changes in doozer
	go doozerCfg.Monitor(key+"*", func(change *doozerconfig.Change, err error) {
		if err != nil {
			Cleanup()
			log.Fatalf("Error processing config change in doozer: %s", err)
			return
		}
		log.Infof("Config changed in doozer; %+v\n", change)
	})
}
