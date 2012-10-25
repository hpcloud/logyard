package main

import (
	"flag"
	"fmt"
	"log"
	"logyard"
	"logyard/stackato"
)

type list struct {
}

func (cmd *list) Name() string {
	return "list"
}

func (cmd *list) DefineFlags(fs *flag.FlagSet) {
}

func (cmd *list) Run() {
	conn, headRev, err := stackato.NewDoozerClient("logyard")
	if err != nil {
		log.Fatal(err)
	}

	manager, err := logyard.NewDrainManager(conn, headRev)
	if err != nil {
		log.Fatal(err)
	}
	for name, uri := range manager.Config.Drains {
		fmt.Printf("%20s\t%s\n", name, uri)
	}
}
