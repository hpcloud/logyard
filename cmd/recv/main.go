package main

import (
	"logyard"
)

func main() {
	c := logyard.NewClient()
	ss := c.Recv("")
	for msg := range ss.Ch {
		println(msg.Key, msg.Value)
	}
}
