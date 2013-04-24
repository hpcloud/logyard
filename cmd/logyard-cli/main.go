package main

import (
	"github.com/ActiveState/log"
	"logyard"
	"logyard/util/subcommand"
	"stackato/server"
)

func Init(name string) {
	conn, headRev, err := server.NewDoozerClient("logyard-cli:" + name)
	if err != nil {
		log.Fatal(err)
	}

	server.Init(conn, headRev)
	logyard.Init2(false)
}

func main() {
	subcommand.Parse(
		new(recv),
		new(stream),
		new(list),
		new(add),
		new(delete))
}
