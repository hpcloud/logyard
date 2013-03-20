package zeroutine

import (
	zmq "github.com/alecthomas/gozmq"
)

// Broker is a zeromq forwarder device acting as a broker between
// multiple publishers and multiple subscribes.
type Broker struct {
	ctx      zmq.Context
	frontend zmq.Socket
	backend  zmq.Socket
	options  Zeroutine
}

func NewBroker(options Zeroutine) (*Broker, error) {
	var err error
	b := new(Broker)
	b.options = options

	if b.ctx, err = GetGlobalContext(); err != nil {
		return nil, err
	}

	// Publishers speak to the frontend socket
	if b.frontend, err = b.ctx.NewSocket(zmq.SUB); err != nil {
		b.ctx.Close()
		return nil, err
	}
	if err = b.frontend.Bind(options.PubAddr); err != nil {
		b.ctx.Close()
		return nil, err
	}
	if err = b.frontend.SetSockOptString(
		zmq.SUBSCRIBE, options.SubscribeFilter); err != nil {
		b.ctx.Close()
		return nil, err
	}

	// Subscribers speak to the backend socket
	if b.backend, err = options.NewPubSocket(); err != nil {
		b.ctx.Close()
		return nil, err
	}
	if err = b.backend.Bind(options.SubAddr); err != nil {
		b.ctx.Close()
		return nil, err
	}

	return b, nil
}

func (b *Broker) Run() error {
	return zmq.Device(zmq.FORWARDER, b.frontend, b.backend)
}
