package lineserver

import (
	"bufio"
)

// AsyncReadline provides a non-blocking version of
// bufio.Reader.ReadString('\n') using goroutines.
type AsyncReadline struct {
	LineCh chan string // Lines read are written to this channel
	ErrCh  chan error  // Error returned by ReadString is written to this channel (once).
	*bufio.Reader
}

func NewAsyncReadline(r *bufio.Reader) *AsyncReadline {
	return &AsyncReadline{
		make(chan string),
		make(chan error),
		r}
}

func (r *AsyncReadline) Run() {
	for {
		if line, err := r.ReadString('\n'); err != nil {
			r.ErrCh <- err
			break
		} else {
			r.LineCh <- line
		}
	}
	close(r.ErrCh)
	close(r.LineCh)
}
