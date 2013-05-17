package logyard

import (
	"github.com/ActiveState/log"
	"stackato/server"
)

type logyardConfig struct {
	RetryLimits  map[string]string `json:"retrylimits"`
	DrainFormats map[string]string `json:"drainformats"`
	Drains       map[string]string `json:"drains"`
}

var config *server.GroupConfig

// GetConfig returns the latest logyard configuration.
func GetConfig() *logyardConfig {
	return config.Config.(*logyardConfig)
}

func GetConfigChanges() chan error {
	return config.GetChangesChannel()
}

// DeleteDrain deletes the drain from config.
func DeleteDrain(name string) error {
	return config.AtomicSave(func(i interface{}) error {
		config := i.(*logyardConfig)
		delete(config.Drains, name)
		return nil
	})
}

// AddDrain adds a drain to the config.
func AddDrain(name, uri string) error {
	return config.AtomicSave(func(i interface{}) error {
		config := i.(*logyardConfig)
		config.Drains[name] = uri
		return nil
	})
}

// Initialize the logyard configuration system.
func Init(name string) {
	server.Init()
	g, err := server.NewGroupConfig("logyard", logyardConfig{})
	if err != nil {
		log.Fatal(err)
	}
	config = g
}
