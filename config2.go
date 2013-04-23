package logyard

import (
	"confdis/go/confdis"
	"github.com/ActiveState/log"
	"stackato/server"
)

type logyardConfig2 struct {
	RetryLimits  map[string]string `json:"retrylimits"`
	DrainFormats map[string]string `json:"drainformats"`
}

var c *confdis.ConfDis

// GetConfig returns the latest logyard configuration.
func GetConfig() *logyardConfig2 {
	return c.Config.(*logyardConfig2)
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

func Init2() {
	var err error
	c, err = confdis.New(server.Config.CoreIP+":5454", "config:logyard", logyardConfig2{})
	if err != nil {
		log.Fatal(err)
	}
	go MonitorConfig(c)
}
