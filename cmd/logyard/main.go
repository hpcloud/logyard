package main

import (
	"github.com/ActiveState/tail"
	_ "logyard"
)

func main() {
	_, _ = tail.TailFile("/no/such/file", tail.Config{Follow: true})
}
