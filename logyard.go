package logyard

import (
	"logyard/zeroutine"
)

var Broker zeroutine.Zeroutine

func init() {
	Broker.PubAddr = "tcp://127.0.0.1:5559"
	Broker.SubAddr = "tcp://127.0.0.1:5560"
	Broker.BufferSize = 100
}
