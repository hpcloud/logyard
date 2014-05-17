package lineserver

import (
	"bufio"
)

// AsyncScanner is bufio.Scanner running in its own goroutine.
type AsyncScanner struct {
	ReadyCh chan bool
	*bufio.Scanner
}

func (s *AsyncScanner) Run() {
	for s.Scan() {
		s.ReadyCh <- true
	}
	close(s.ReadyCh)
}
