package main

import (
	"github.com/ActiveState/log"
	"logyard"
	"logyard/stackato"
	"logyard/stackato/server"
)

func main() {
	doozer, headRev, err := stackato.NewDoozerClient("logyard")
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
