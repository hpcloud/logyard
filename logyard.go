package logyard

import (
	"encoding/json"
	"github.com/ActiveState/doozer"
	"github.com/srid/doozerconfig"
	"github.com/srid/log2"
)

type logyardConfig struct {
	Drains map[string]string `doozer:"drains"`
	Doozer *doozer.Conn
	Rev    int64                     // Doozer revision this config struct corresponds to
	Ch     chan *doozerconfig.Change // Doozer changes to the Drains map
}

// Logyard configuration object tied to doozer config
var Config *logyardConfig

const DOOZER_PREFIX = "/proc/logyard/config/"

var doozerCfg *doozerconfig.DoozerConfig

// Init initializes logyard with the given config entry point in
// doozer.
func Init(conn *doozer.Conn, rev int64, monitor bool) {
	Config = new(logyardConfig)
	Config.Drains = make(map[string]string)
	Config.Ch = make(chan *doozerconfig.Change)
	Config.Rev = rev
	Config.Doozer = conn
	doozerCfg = doozerconfig.New(conn, rev, Config, DOOZER_PREFIX)
	err := doozerCfg.Load()
	if err != nil {
		log2.Fatal(err)
	}

	if !monitor {
		return
	}

	// Monitor drain config changes in doozer
	go doozerCfg.Monitor(DOOZER_PREFIX+"drains/*", func(change *doozerconfig.Change, err error) {
		if err != nil {
			// don't bring down the entire logyard, because we want
			// the existing drains to continue to function despite an
			// error in monitoring config changes.
			log2.Errorf("Unable to process drain config change in doozer: %s", err)
			return
		}
		log2.Infof("Detected change in doozer drains config: %s", change.Key)
		// Config.Rev = TODO
		if change.FieldName == "Drains" {
			switch change.Type {
			case doozerconfig.DELETE, doozerconfig.SET:
				Config.Ch <- change
			}
		}
	})
}

// DeleteDrain deletes the drain from doozer tree
func (Config *logyardConfig) DeleteDrain(name string) error {
	err := Config.Doozer.Del(DOOZER_PREFIX+"drains/"+name, Config.Rev)
	if err != nil {
		return err
	}
	return nil
}

// AddDrain adds the drain to the doozer tree
func (Config *logyardConfig) AddDrain(name string, uri string) error {
	data, err := json.Marshal(uri)
	if err != nil {
		return err
	}
	_, err = Config.Doozer.Set(DOOZER_PREFIX+"drains/"+name, Config.Rev, data)
	if err != nil {
		return err
	}
	return nil
}
