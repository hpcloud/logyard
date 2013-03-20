package logyard

import (
	"github.com/ActiveState/log"
	zmq "github.com/alecthomas/gozmq"
	"logyard/zeroutine"
)

// A logyard client must not be shared between threads (and thus
// goroutines).
type Client struct {
	ctx     zmq.Context
	pubSock zmq.Socket
}

// NewClient creates a new logyard Client. If `rw' is set, `Send' will
// be supported. FIXME: decouple recv and send.
func NewClient(ctx zmq.Context, rw bool) (*Client, error) {
	c := &Client{ctx, nil}
	if rw {
		var err error
		c.pubSock, err = zeroutine.NewPubSocket(c.ctx, MEMORY_BUFFER_SIZE)
		if err != nil {
			return nil, err
		}
		err = c.pubSock.Connect(PUBLISHER_ADDR)
		if err != nil {
			return nil, err
		}

	}
	return c, nil
}

func NewClientGlobal(rw bool) (*Client, error) {
	ctx, err := zeroutine.GetGlobalContext()
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, rw)
}

func (c *Client) Send(key string, value string) error {
	if c.pubSock == nil {
		log.Fatal("Client was created with rw=false")
	}
	return c.pubSock.Send([]byte(key+" "+value), 0)
}

func (c *Client) Recv(filters []string) (*zeroutine.SubChannel, error) {
	return zeroutine.NewSubChannel(SUBSCRIBER_ADDR, filters), nil
}

func (c *Client) Close() {
	if c.pubSock != nil {
		c.pubSock.Close()
	}
}
