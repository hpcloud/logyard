// lineserver emulates a line-based UDP server
package lineserver

import (
	"bufio"
	"net"

	"gopkg.in/tomb.v1"
)

// LineServer is a line-based UDP server à la `nc -u -l`. Ch channel
// will receive incoming lines from all clients.
type LineServer struct {
	Ch          chan string
	tcp         bool // True if using tcp
	udpConn     net.Conn
	tcpListener net.Listener
	tomb.Tomb   // provides: Done, Kill, Dying
}

func NewLineServer(proto, addr string) (*LineServer, error) {
	if proto == "tcp" {
		return NewLineServerTCP(addr)
	} else if proto == "udp" {
		return NewLineServerUDP(addr)
	} else {
		panic("unknown proto")
	}
}

func NewLineServerUDP(addr string) (*LineServer, error) {
	laddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return nil, err
	}
	var ls LineServer
	ls.Ch = make(chan string)
	ls.udpConn = conn
	ls.tcp = false
	return &ls, nil
}

func NewLineServerTCP(addr string) (*LineServer, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	var ls LineServer
	ls.Ch = make(chan string)
	ls.tcpListener = ln
	ls.tcp = true
	return &ls, nil
}

func (srv *LineServer) GetUDPAddr() (*net.UDPAddr, error) {
	return net.ResolveUDPAddr("udp", srv.udpConn.LocalAddr().String())
}

func (srv *LineServer) Start() {
	if srv.tcp {
		srv.serveTcp()
	} else {
		srv.serveConn(srv.udpConn)
	}
}

func (srv *LineServer) serveTcp() {
	for {
		conn, err := srv.tcpListener.Accept()
		if err != nil {
			// Probably not a good idea to kill, thereby ignoring
			// other connections.
			srv.Kill(err)
			return
		}

		// Handle this connection in a goroutine
		go srv.serveConn(conn)
	}
}

// Start starts the server. Call as a goroutine.
func (srv *LineServer) serveConn(conn net.Conn) {
	defer conn.Close()
	defer close(srv.Ch)
	defer srv.Done()

	lineReader := NewAsyncReadline(bufio.NewReader(conn))
	// Closing conn automatically ends lineReader goroutine.
	go lineReader.Run()

	// Scanned tokens are limited in max size (64 * 1024); see
	// pkg/bufio/scan.go:MaxScanTokenSize in Go source tree.
	for {
		select {
		case text := <-lineReader.LineCh:
			select {
			case srv.Ch <- text:
			case <-srv.Dying():
				return
			}
		case err := <-lineReader.ErrCh:
			srv.Kill(err)
		case <-srv.Dying():
			return
		}
	}
}
