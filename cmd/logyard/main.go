package main

import (
	"log"
	"logyard"
	"logyard/stackato"
)

func main() {
	doozer, headRev, err := stackato.NewDoozerClient("logyard")
	if err != nil {
		log.Fatal(err)
	}

	logyard.Init(doozer, headRev, true)

	f, err := logyard.NewForwarder()
	if err != nil {
		log.Fatal(err)
	}
	m := logyard.NewDrainManager()
	log.Println("Starting drain manager")
	go m.Run()
	log.Println("Running zmq forwarder", f)
	f.Run()
}
