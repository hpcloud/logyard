package logyard

import (
	"github.com/ActiveState/log"
	"logyard/zeroutine"
)

var Logyard zeroutine.Zeroutine

func init() {
	Logyard.PubAddr = "tcp://127.0.0.1:5559"
	Logyard.SubAddr = "tcp://127.0.0.1:5560"
	Logyard.BufferSize = 100
}

func RunBroker() {
	log.Fatal(Logyard.RunBroker())
}
