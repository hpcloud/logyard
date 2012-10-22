package logyard

import (
	zmq "github.com/alecthomas/gozmq"
	"strings"
	"sync"
)

type Client struct {
	ctx     zmq.Context
	pubSock zmq.Socket
}

func NewClient(ctx zmq.Context) *Client {
	return &Client{ctx, nil}
}

var globalContext zmq.Context
var globalContextErr error
var once sync.Once

func NewClientGlobal() (*Client, error) {
	once.Do(func() {
		globalContext, globalContextErr = zmq.NewContext()
	})
	if globalContextErr != nil {
		return nil, globalContextErr
	}
	return NewClient(globalContext), nil
}

func (c *Client) Send(key string, value string) error {
	err := c.init(true)
	if err != nil {
		return err
	}

	return c.pubSock.Send([]byte(key+" "+value), 0)
}

func (c *Client) Recv(filters []string) (*SubscribeStream, error) {
	err := c.init(false)
	if err != nil {
		return nil, err
	}
	addr := strings.Replace(SUBSCRIBER_ADDR, "*", "127.0.0.1", 1)
	return NewSubscribeStream(c.ctx, addr, filters), nil
}

func (c *Client) Close() {
	c.ctx.Close()
}

func (c *Client) init(send bool) error {
	if send && c.pubSock == nil {
		var err error
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
