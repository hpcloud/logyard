package logyard

import (
	zmq "github.com/alecthomas/gozmq"
)

type Client struct {
	ctx     zmq.Context
	pubSock zmq.Socket
}

func NewClient() *Client {
	return new(Client)
}

// StreamSend streams the message to local logyard instance.
func (c *Client) Send(key string, value string) error {
	err := c.init(true)
	if err != nil {
		return err
	}

	return c.pubSock.Send([]byte(key+" "+value), 0)
}

func (c *Client) Recv(filter string) *SubscribeStream {
	err := c.init(false)
	if err != nil {
		panic(err)
	}
	return NewSubscribeStream(c.ctx, SUBSCRIBER_ADDR, filter)
}

func (c *Client) init(send bool) error {
	var err error
	if c.ctx == nil {
		c.ctx, err = zmq.NewContext()
		if err != nil {
			return err
		}
	}

	if send && c.pubSock == nil {
		c.pubSock, err = c.ctx.NewSocket(zmq.PUB)
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
