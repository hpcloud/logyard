package logyard

import (
	"github.com/ActiveState/log"
	"github.com/ActiveState/zmqpubsub"
)

var Broker zmqpubsub.Broker

func init() {
	Broker.PubAddr = "ipc:///tmp/logyardpub.sock"
	Broker.SubAddr = "ipc:///tmp/logyardsub.sock"
	Broker.BufferSize = 100

	log.Infof("Loygard broker config: %+v\n", Broker)
}
