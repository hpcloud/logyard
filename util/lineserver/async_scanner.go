package lineserver

import (
	"bufio"
)

// AsyncScanner is a bufio.Scanner with asynchronous `Scan` function
type AsyncScanner struct {
	ReadyCh chan bool
	*bufio.Scanner
}

func (s *AsyncScanner) Run() {
	println("Scanner start")
	for s.Scan() {
		s.ReadyCh <- true
	}
	close(s.ReadyCh)
	println("Scanner end")
}
