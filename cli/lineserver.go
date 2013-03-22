package cli

// Funtionality to emulate line-based TCP server

import (
	"bufio"
	"github.com/ActiveState/log"
	"io"
	"net"
)

// LineServer is a line-based server à la `nc -l`. Ch channel will
// receive incoming lines from all clients.
type LineServer struct {
	listener net.Listener
	Ch       chan []byte
}

func NewLineServer(proto, laddr string) (*LineServer, error) {
	l, err := net.Listen(proto, laddr)
	if err != nil {
		return nil, err
	}
	return &LineServer{l, make(chan []byte)}, nil
}

// Start starts the server. Call as a goroutine.
func (srv *LineServer) Start() {
	for {
		conn, err := srv.listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go func(conn net.Conn) {
			for {
				reader := bufio.NewReader(conn)
				line, _, err := reader.ReadLine()

				if err == io.EOF {
					// Exit silently if a client disconnects.
					return
				}

				if err != nil {
					log.Fatal(err)
				}
				srv.Ch <- line
			}
		}(conn)
	}
}
