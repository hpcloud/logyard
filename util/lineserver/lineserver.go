package lineserver

// Funtionality to emulate line-based TCP server

import (
	"bufio"
	"github.com/ActiveState/log"
	"net"
)

// LineServer is a line-based UDP server à la `nc -l`. Ch channel
// will receive incoming lines from all clients.
type LineServer struct {
	conn *net.UDPConn
	Ch   chan string
}

func NewLineServer(addr string) (*LineServer, error) {
	laddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return nil, err
	}
	return &LineServer{conn, make(chan string)}, nil
}

// Start starts the server. Call as a goroutine.
func (srv *LineServer) Start() {
	scanner := bufio.NewScanner(srv.conn)
	// Scanned tokens are limited in max size (64 * 1024); see
	// pkg/bufio/scan.go:MaxScanTokenSize in Go source tree.
	for scanner.Scan() {
		srv.Ch <- scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
