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

var Config2 *logyardConfig2

func newLogyardConfig2() *logyardConfig2 {
	c := new(logyardConfig2)
	c.RetryLimits = make(map[string]string)
	c.DrainFormats = make(map[string]string)
	return c
}

func Init2() {
	Config2 = newLogyardConfig2()
	_, err := confdis.New(server.Config.CoreIP+":5454", "config:logyard", Config2)
	if err != nil {
		log.Fatal(err)
	}
}
