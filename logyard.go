package logyard

import (
	"github.com/ActiveState/log"
	"logyard/zeroutine"
)

const (
	PUBLISHER_ADDR     = "tcp://127.0.0.1:5559"
	SUBSCRIBER_ADDR    = "tcp://127.0.0.1:5560"
	MEMORY_BUFFER_SIZE = 100
)

func RunBroker() {
	broker, err := zeroutine.NewBroker(
		zeroutine.BrokerOptions{
			PubAddr:    PUBLISHER_ADDR,
			SubAddr:    SUBSCRIBER_ADDR,
			BufferSize: MEMORY_BUFFER_SIZE})
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(broker.Run())
}
