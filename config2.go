package logyard

import (
	"confdis/go/confdis"
	"github.com/ActiveState/log"
	"stackato/server"
)

type logyardConfig2 struct {
	RetryLimits  map[string]string `json:"retrylimits"`
	DrainFormats map[string]string `json:"drainformats"`
	Drains       map[string]string `json:"drains"`
}

var c *confdis.ConfDis

// GetConfig returns the latest logyard configuration.
func GetConfig() *logyardConfig2 {
	return c.Config.(*logyardConfig2)
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
		config := i.(*logyardConfig2)
		delete(config.Drains, name)
		return nil
	})
}

// AddDrain adds a drain to the doozer tree.
func AddDrain(name, uri string) error {
	return c.AtomicSave(func(i interface{}) error {
		config := i.(*logyardConfig2)
		config.Drains[name] = uri
		return nil
	})
}

func Init2(monitor bool) {
	var err error
	if server.Config == nil {
		log.Fatal("stackato-go:server not initialized")
	}
	c, err = confdis.New(server.Config.CoreIP+":5454", "config:logyard", logyardConfig2{})
	if err != nil {
		log.Fatal(err)
	}
	if monitor {
		go MonitorConfig(c)
	}
}
