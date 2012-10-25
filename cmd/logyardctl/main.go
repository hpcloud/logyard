package main

import (
	"log"
	"logyard"
	"logyard/cmd/logyardctl/subcommand"
	"logyard/stackato"
)

func Init() {
	conn, headRev, err := stackato.NewDoozerClient("logyard")
	if err != nil {
		log.Fatal(err)
	}

	logyard.Init(conn, headRev, false)
}

func main() {
	subcommand.Parse(
		new(recv),
		new(list),
		new(add),
		new(delete))
}
