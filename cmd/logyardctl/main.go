package main

import (
	"logyard"
	"logyard/cmd/logyardctl/subcommand"
	"logyard/log2"
	"logyard/stackato"
)

func Init() {
	conn, headRev, err := stackato.NewDoozerClient("logyardctl")
	if err != nil {
		log2.Fatal(err)
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
