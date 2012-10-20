package logyard

import (
	zmq "github.com/alecthomas/gozmq"
	"launchpad.net/tomb"
	"strings"
)

type Message struct {
	Key   string
	Value string
}

func NewMessage(data []byte) *Message {
	parts := strings.SplitN(string(data), " ", 2)
	return &Message{parts[0], parts[1]}
}

type SubscribeStream struct {
	ctx     zmq.Context
	addr    string
	filters []string
	Ch      chan *Message
	tomb.Tomb
}

func NewSubscribeStream(ctx zmq.Context, addr string, filters []string) *SubscribeStream {
	ss := &SubscribeStream{}
	ss.ctx = ctx
	ss.addr = addr
	ss.filters = filters
	ss.Ch = make(chan *Message)
	go ss.run()
	return ss
}

func (ss *SubscribeStream) run() {
	defer ss.Done()

	// Establish a connection and subscription filter
	socket, err := ss.ctx.NewSocket(zmq.SUB)
	if err != nil {
		ss.Kill(err)
		close(ss.Ch)
		return
	}

	for _, filter := range ss.filters {
		err = socket.SetSockOptString(zmq.SUBSCRIBE, filter)
		if err != nil {
			ss.Kill(err)
			close(ss.Ch)
			return
		}
	}

	err = socket.Connect(ss.addr)
	if err != nil {
		ss.Killf("Couldn't connect to %s: %s", ss.addr, err)
		close(ss.Ch)
		return
	}

	// Read and stream the results in a channel
	go func() {
		for {
			data, err := socket.Recv(0)
			if err != nil {
				ss.Kill(err)
				close(ss.Ch)
				return
			}
			ss.Ch <- NewMessage(data)
		}
	}()

	<-ss.Dying()
}
