package logyard

import (
	"encoding/json"
	"github.com/ActiveState/doozerconfig"
	"github.com/ActiveState/log"
	"github.com/ha/doozer"
)

type logyardConfig struct {
	Drains       map[string]string `doozer:"drains"`
	RetryLimits  map[string]string `doozer:"retrylimits"`
	DrainFormats map[string]string `doozer:"drainformats"`
	Doozer       *doozer.Conn
	Rev          int64                     // Doozer revision this config struct corresponds to
	DrainChanges chan *doozerconfig.Change // Doozer changes to the Drains map
}

// Logyard configuration object tied to doozer config
var Config *logyardConfig

const DOOZER_PREFIX = "/proc/logyard/config/"

var doozerCfg *doozerconfig.DoozerConfig

func newLogyardConfig(conn *doozer.Conn, rev int64) *logyardConfig {
	c := new(logyardConfig)
	c.Drains = make(map[string]string)
	c.RetryLimits = make(map[string]string)
	c.DrainFormats = make(map[string]string)
	c.DrainChanges = make(chan *doozerconfig.Change)
	c.Rev = rev
	c.Doozer = conn
	return c
}

// DeleteDrain deletes the drain from doozer tree
func (Config *logyardConfig) DeleteDrain(name string) error {
	Config.mustBeInitialized()
	err := Config.Doozer.Del(DOOZER_PREFIX+"drains/"+name, -1)
	if err != nil {
		return err
	}
	return nil
}

// AddDrain adds a drain. URI should not contain a query fragment,
// which will be constructed from the `filters` and `params`
// arguments.
func (Config *logyardConfig) AddDrain(name, uri string) error {
	Config.mustBeInitialized()

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

func (Config *logyardConfig) mustBeInitialized() {
	// XXX: there should be a way to make the compiler do this job.
	if Config == nil {
		log.Fatal("Config object not initialized (`Init()` not called)")
	}
}

func doozerConfigChangedCallbackFn(change *doozerconfig.Change, err error) {
	if err != nil {
		// Do not crash logyard, because we want the existing drains
		// to continue to function despite an error in monitoring
		// config changes.
		log.Errorf("Unable to process drain config change in doozer: %s", err)
		return
	}
	log.Infof("Detected change in doozer config: %s (%s)", change.Key, change.FieldName)
	if change.FieldName == "Drains" {
		switch change.Type {
		case doozerconfig.DELETE, doozerconfig.SET:
			Config.DrainChanges <- change
		}
	}
}

// Init initializes logyard with the given config entry point in
// doozer.
func Init(conn *doozer.Conn, rev int64, monitor bool) {
	Config = newLogyardConfig(conn, rev)
	doozerCfg = doozerconfig.New(conn, rev, Config, DOOZER_PREFIX)

	if err := doozerCfg.Load(); err != nil {
		log.Fatal(err)
	}

	if monitor {
		// Monitor drain config changes in doozer
		go doozerCfg.Monitor(DOOZER_PREFIX+"**", doozerConfigChangedCallbackFn)
	}
}
