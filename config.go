package logyard

import (
	"encoding/json"
	"fmt"
	"github.com/ActiveState/doozerconfig"
	"github.com/ActiveState/log"
	"github.com/ha/doozer"
	"net/url"
	"strings"
)

type logyardConfig struct {
	Drains       map[string]string `doozer:"drains"`
	RetryLimits  map[string]string `doozer:"retrylimits"`
	DrainFormats map[string]string `doozer:"drainformats"`
	Doozer       *doozer.Conn
	Rev          int64                     // Doozer revision this config struct corresponds to
	Ch           chan *doozerconfig.Change // Doozer changes to the Drains map
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
	c.Ch = make(chan *doozerconfig.Change)
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
func (Config *logyardConfig) AddDrain(
	name, uri string, filters []string, params map[string]string) (string, error) {
	Config.mustBeInitialized()
	if uri == "" {
		return "", fmt.Errorf("URI cannot be empty")
	}

	if !strings.Contains(uri, "://") {
		return "", fmt.Errorf("Not an URI: %s", uri)
	}

	// Build the query string
	query := url.Values{}
	for _, filter := range filters {
		query.Add("filter", filter)
	}
	for key, value := range params {
		if key == "filter" {
			return "", fmt.Errorf("params cannot have a key called 'filter'")
		}
		query.Set(key, value)
	}

	uri += "?" + query.Encode()

	err := Config.addDrain(name, uri)
	if err != nil {
		return "", err
	}
	return uri, err
}

func (Config *logyardConfig) addDrain(name string, uri string) error {
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

// Init initializes logyard with the given config entry point in
// doozer.
func Init(conn *doozer.Conn, rev int64, monitor bool) {
	Config = newLogyardConfig(conn, rev)
	doozerCfg = doozerconfig.New(conn, rev, Config, DOOZER_PREFIX)
	err := doozerCfg.Load()
	if err != nil {
		log.Fatal(err)
	}

	if !monitor {
		return
	}

	// Monitor drain config changes in doozer
	go doozerCfg.Monitor(DOOZER_PREFIX+"drains/*",
		func(change *doozerconfig.Change, err error) {
			if err != nil {
				// don't bring down the entire logyard, because we want
				// the existing drains to continue to function despite an
				// error in monitoring config changes.
				log.Errorf("Unable to process drain config change in doozer: %s", err)
				return
			}
			log.Infof("Detected change in doozer drains config: %s", change.Key)
			// Config.Rev = TODO
			if change.FieldName == "Drains" {
				switch change.Type {
				case doozerconfig.DELETE, doozerconfig.SET:
					Config.Ch <- change
				}
			}
		},
	)
}
