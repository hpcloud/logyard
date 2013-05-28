package logyard

import (
	"github.com/ActiveState/log"
	"stackato/server"
	"sync"
)

type logyardConfig struct {
	RetryLimits  map[string]string `json:"retrylimits"`
	DrainFormats map[string]string `json:"drainformats"`
	Drains       map[string]string `json:"drains"`
}

var config *server.Config

// GetConfig returns the latest logyard configuration.
func GetConfig() *logyardConfig {
	once.Do(createLogyardConfig)
	return config.GetConfig().(*logyardConfig)
}

func GetConfigChanges() chan error {
	once.Do(createLogyardConfig)
	return config.GetChangesChannel()
}

// DeleteDrain deletes the drain from config.
func DeleteDrain(name string) error {
	once.Do(createLogyardConfig)
	return config.AtomicSave(func(i interface{}) error {
		config := i.(*logyardConfig)
		delete(config.Drains, name)
		return nil
	})
}

// AddDrain adds a drain to the config.
func AddDrain(name, uri string) error {
	once.Do(createLogyardConfig)
	return config.AtomicSave(func(i interface{}) error {
		config := i.(*logyardConfig)
		config.Drains[name] = uri
		return nil
	})
}

var once sync.Once

func createLogyardConfig() {
	g, err := server.NewConfig("logyard", logyardConfig{})
	if err != nil {
		log.Fatal(err)
	}
	config = g
	if config.GetConfig().(*logyardConfig).Drains == nil {
		log.Fatal("Logyard configuration is missing")
	}
}
