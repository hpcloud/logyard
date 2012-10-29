package logyard

import (
	zmq "github.com/alecthomas/gozmq"
	"logyard/zmqch"
	"strings"
)

// XXX: rewrite the client api to separate read/write. very confusing otherwise.
type Client struct {
	ctx     zmq.Context
	pubSock zmq.Socket
}

func NewClient(ctx zmq.Context) *Client {
	return &Client{ctx, nil}
}

func NewClientGlobal() (*Client, error) {
	ctx, err := zmqch.GetGlobalContext()
	if err != nil {
		return nil, err
	}
	return NewClient(ctx), nil
}

func (c *Client) Send(key string, value string) error {
	err := c.init(true)
	if err != nil {
		return err
	}

	return c.pubSock.Send([]byte(key+" "+value), 0)
}

func (c *Client) Recv(filters []string) (*zmqch.SubChannel, error) {
	err := c.init(false)
	if err != nil {
		return nil, err
	}
	addr := strings.Replace(SUBSCRIBER_ADDR, "*", "127.0.0.1", 1)
	return zmqch.NewSubChannel(addr, filters), nil
}

func (c *Client) Close() {
	if c.pubSock != nil {
		c.pubSock.Close()
	}
}

func (c *Client) init(send bool) error {
	if send && c.pubSock == nil {
		var err error
		c.pubSock, err = NewPubSocket(c.ctx)
		if err != nil {
			return err
		}
		err = c.pubSock.Connect(PUBLISHER_ADDR)
		if err != nil {
			return err
		}
	}
	return nil
}
