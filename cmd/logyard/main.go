package main

import (
	"github.com/srid/log2"
	"logyard"
	"logyard/stackato"
	"logyard/stackato/server"
)

func main() {
	doozer, headRev, err := stackato.NewDoozerClient("logyard")
	if err != nil {
		log2.Fatal(err)
	}

	logyard.Init(doozer, headRev, true)
	server.Init(doozer, headRev)

	f, err := logyard.NewForwarder()
	if err != nil {
		log2.Fatal(err)
	}
	m := logyard.NewDrainManager()
	log2.Info("Starting drain manager")
	go m.Run()
	log2.Info("Running zmq forwarder", f)
	f.Run()
}
