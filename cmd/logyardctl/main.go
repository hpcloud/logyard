package main

import (
	"github.com/ActiveState/log"
	"logyard"
	"logyard/cmd/logyardctl/subcommand"
	"stackato/server"
)

func Init() {
	conn, headRev, err := server.NewDoozerClient("logyardctl")
	if err != nil {
		log.Fatal(err)
	}

	logyard.Init(conn, headRev, false)
}

func main() {
	subcommand.Parse(
		new(recv),
		new(stream),
		new(list),
		new(add),
		new(delete))
}
