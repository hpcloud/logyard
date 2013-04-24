package logyard

import (
	"confdis/go/confdis"
	"github.com/ActiveState/log"
	"stackato/server"
)

type logyardConfig struct {
	RetryLimits  map[string]string `json:"retrylimits"`
	DrainFormats map[string]string `json:"drainformats"`
	Drains       map[string]string `json:"drains"`
}

var c *confdis.ConfDis

// GetConfig returns the latest logyard configuration.
func GetConfig() *logyardConfig {
	return c.Config.(*logyardConfig)
}

func GetConfigChanges() chan error {
	return c.Changes
}

// MonitorConfig monitors for configuration changes, and exits if
// there is an error.
func MonitorConfig(c *confdis.ConfDis) {
	for err := range c.Changes {
		if err != nil {
			log.Fatalf("Error re-loading config: %v", err)
		}
		log.Info("Config changed.")
	}
}

// DeleteDrain deletes the drain from config.
func DeleteDrain(name string) error {
	return c.AtomicSave(func(i interface{}) error {
		config := i.(*logyardConfig)
		delete(config.Drains, name)
		return nil
	})
}

// AddDrain adds a drain to the config.
func AddDrain(name, uri string) error {
	return c.AtomicSave(func(i interface{}) error {
		config := i.(*logyardConfig)
		config.Drains[name] = uri
		return nil
	})
}

// Initialize the logyard configuration system, optionally monitoring
// for future changes.
func Init(name string, monitor bool) {
	// Initialize doozer connection for reading the redis URI.
	if server.Config != nil {
		panic("stackato-go:server already initialized")
	}
	conn, headRev, err := server.NewDoozerClient("logyard-cli:" + name)
	if err != nil {
		log.Fatal(err)
	}
	server.Init(conn, headRev)

	// Setup confdis.
	c, err = confdis.New(server.Config.CoreIP+":5454", "config:logyard", logyardConfig{})
	if err != nil {
		log.Fatal(err)
	}

	// Monitor for future config changes if required.
	if monitor {
		go MonitorConfig(c)
	}
}
