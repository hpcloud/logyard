package drain

import (
	"net"
	"time"
)

// NetDialer provides a channel-friendly wrapper for net.DialTimeout
// so that it can be used along with `select{}`.
type NetDialer struct {
	Ch    chan net.Conn
	Error error
}

func NewNetDialer(scheme, host string, timeout time.Duration) *NetDialer {
	d := NetDialer{make(chan net.Conn), nil}
	go d.dial(scheme, host, timeout)
	return &d
}

func (d *NetDialer) dial(scheme, host string, timeout time.Duration) {
	conn, err := net.DialTimeout(scheme, host, timeout)
	if err != nil {
		d.Error = err
	}
	d.Ch <- conn
}

// WaitAndClose waits for the connection to return and closes it
// immediately.
func (d *NetDialer) WaitAndClose() {
	conn := <-d.Ch
	if d.Error == nil {
		conn.Close()
	}
}
