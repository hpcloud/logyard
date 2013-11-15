package pubchannel

import (
	"encoding/json"
	"github.com/ActiveState/zmqpubsub"
	"logyard"
	"logyard/clients/common"
	"time"
)

// PubChannel abstracts zmqpubsub.Publisher using Go's channels. Unlike
// zmqpubsub, senders can be any goroutine.
type PubChannel struct {
	Ch  chan interface{}
	key string
	pub *zmqpubsub.Publisher
}

func New(key string, stopCh chan bool) *PubChannel {
	pubch := &PubChannel{make(chan interface{}), key, nil}
	go pubch.loop(stopCh)
	return pubch
}

func (pubch *PubChannel) loop(stopCh chan bool) {
	if pubch.pub != nil {
		panic("loop called twice?")
	}
	pubch.pub = logyard.Broker.NewPublisherMust()

	select {
	// XXX: this delay is unfortunately required, else the publish calls
	// (instance.notify) below for warnings will get ignored.
	case <-time.After(100 * time.Millisecond):
	case <-stopCh:
		return
	}

	for data := range pubch.Ch {
		b, err := json.Marshal(data)
		if err != nil {
			common.Fatal("%v", err)
		}
		pubch.pub.MustPublish(pubch.key, string(b))
	}
}
