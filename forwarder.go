package logyard

import (
	zmq "github.com/alecthomas/gozmq"
)

// Forwarder is a 0MQ forwarder device for transporting log streams.
type Forwarder struct {
	ctx      zmq.Context
	frontend zmq.Socket
	backend  zmq.Socket
}

const (
	PUBLISHER_ADDR     = "tcp://127.0.0.1:5559"
	SUBSCRIBER_ADDR    = "tcp://127.0.0.1:5560"
	MEMORY_BUFFER_SIZE = 100
)

func NewForwarder() (*Forwarder, error) {
	f := new(Forwarder)
	var err error
	f.ctx, err = zmq.NewContext()
	if err != nil {
		return nil, err
	}

	// frontend speaks to publishers
	f.frontend, err = f.ctx.NewSocket(zmq.SUB)
	if err != nil {
		f.ctx.Close()
		return nil, err
	}
	err = f.frontend.Bind(PUBLISHER_ADDR)
	if err != nil {
		f.ctx.Close()
		return nil, err
	}
	err = f.frontend.SetSockOptString(zmq.SUBSCRIBE, "")
	if err != nil {
		f.ctx.Close()
		return nil, err
	}

	// backend speaks to subscribers
	f.backend, err = NewPubSocket(f.ctx)
	if err != nil {
		f.ctx.Close()
		return nil, err
	}
	err = f.backend.Bind(SUBSCRIBER_ADDR)
	if err != nil {
		f.ctx.Close()
		return nil, err
	}

	return f, nil
}

func NewPubSocket(ctx zmq.Context) (zmq.Socket, error) {
	sock, err := ctx.NewSocket(zmq.PUB)
	if err != nil {
		return nil, err
	}
	// prevent 0mq from infinitely buffering messages
	err = sock.SetSockOptUInt64(zmq.HWM, MEMORY_BUFFER_SIZE)
	if err != nil {
		sock.Close()
		return nil, err
	}
	return sock, nil
}

func (f *Forwarder) Run() {
	panic(zmq.Device(zmq.FORWARDER, f.frontend, f.backend))
}
