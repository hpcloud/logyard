package zeroutine

import (
	zmq "github.com/alecthomas/gozmq"
)

type Zeroutine struct {
	PubAddr         string // Publisher Endpoint Address 
	SubAddr         string // Subscriber Endpoint Address
	BufferSize      int    // Memory buffer size
	SubscribeFilter string
}

// Run runs a broker for this zeroutine configuration.
func (z Zeroutine) Run() error {
	broker, err := NewBroker(z)
	if err == nil {
		err = broker.Run()
	}
	return err
}

// Subscribe returns a subscription (channel) for given filters.
func (z Zeroutine) Subscribe(filters ...string) *Subscription {
	return newSubscription(z.SubAddr, filters)
}

func (z Zeroutine) NewPublisher() (*Publisher, error) {
	sock, err := newPubSocket(z.BufferSize)
	if err != nil {
		return nil, err
	}
	if err = sock.Connect(z.PubAddr); err != nil {
		sock.Close()
		return nil, err
	}
	// Publisher.Close is responsible for closing `sock`.
	return newPublisher(sock), nil
}

func newPubSocket(bufferSize int) (zmq.Socket, error) {
	ctx, err := GetGlobalContext()
	if err != nil {
		return nil, err
	}

	sock, err := ctx.NewSocket(zmq.PUB)
	if err != nil {
		return nil, err
	}

	// prevent 0mq from infinitely buffering messages
	for _, hwm := range []zmq.IntSocketOption{zmq.SNDHWM, zmq.RCVHWM} {
		err = sock.SetSockOptInt(hwm, bufferSize)
		if err != nil {
			sock.Close()
			return nil, err
		}
	}

	return sock, nil
}
