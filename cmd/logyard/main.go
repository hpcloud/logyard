package main

import (
	"github.com/srid/log"
	"logyard"
	"stackato-go/server"
)

func main() {
	doozer, headRev, err := server.NewDoozerClient("logyard")
	if err != nil {
		log.Fatal(err)
	}

	logyard.Init(doozer, headRev, true)
	server.Init(doozer, headRev)

	f, err := logyard.NewForwarder()
	if err != nil {
		log.Fatal(err)
	}
	m := logyard.NewDrainManager()
	log.Info("Starting drain manager")
	go m.Run()
	log.Info("Running zmq forwarder", f)
	f.Run()
}
