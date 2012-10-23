// Higher-level abstractions on top of gozmq
package logyard

import (
	zmq "github.com/alecthomas/gozmq"
	"launchpad.net/tomb"
	"strings"
	"sync"
)

var globalContext zmq.Context
var globalContextErr error
var once sync.Once

func initializeGlobalContext() {
	globalContext, globalContextErr = zmq.NewContext()
}

// GetGlobalContext returns a global zmq Context for this process.
func GetGlobalContext() (zmq.Context, error) {
	once.Do(initializeGlobalContext)
	return globalContext, globalContextErr
}

type Message struct {
	Key   string
	Value string
}

func NewMessage(data []byte) *Message {
	parts := strings.SplitN(string(data), " ", 2)
	return &Message{parts[0], parts[1]}
}

// SubChannel provides channel abstraction over zmq SUB sockets
type SubChannel struct {
	addr    string
	filters []string
	Ch      chan *Message
	tomb.Tomb
}

func NewSubChannel(addr string, filters []string) *SubChannel {
	sub := new(SubChannel)
	sub.addr = addr
	sub.filters = filters
	sub.Ch = make(chan *Message)
	go sub.run()
	return sub
}

func (sub *SubChannel) run() {
	defer sub.Done()

	ctx, err := GetGlobalContext()
	if err != nil {
		sub.Kill(err)
		close(sub.Ch)
		return
	}

	// Establish a connection and subscription filter
	socket, err := ctx.NewSocket(zmq.SUB)
	if err != nil {
		sub.Kill(err)
		close(sub.Ch)
		return
	}

	for _, filter := range sub.filters {
		err = socket.SetSockOptString(zmq.SUBSCRIBE, filter)
		if err != nil {
			sub.Kill(err)
			close(sub.Ch)
			return
		}
	}

	err = socket.Connect(sub.addr)
	if err != nil {
		sub.Killf("Couldn't connect to %s: %s", sub.addr, err)
		close(sub.Ch)
		return
	}

	// Read and stream the results in a channel
	go func() {
		pollItems := []zmq.PollItem{
			zmq.PollItem{socket, 0, zmq.POLLIN, 0}}
		for {
			// timeout in microseconds
			n, err := zmq.Poll(pollItems, 1000*1000)
			if err != nil {
				sub.Kill(err)
				close(sub.Ch)
				return
			}

			select {
			case <-sub.Dying():
				return
			default:
			}

			if n > 0 {
				data, err := socket.Recv(0)
				if err != nil {
					sub.Kill(err)
					close(sub.Ch)
					return
				}
				sub.Ch <- NewMessage(data)
			}
		}
	}()

	<-sub.Dying()
}

func (sub *SubChannel) Stop() error {
	sub.Kill(nil)
	return sub.Wait()
}
