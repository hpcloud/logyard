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

	f, err := logyard.NewForwarder()
	if err != nil {
		panic(err)
	}
	m, err := logyard.NewDrainManager(doozer, headRev)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Starting drain manager")
	m.Run()
	log.Println("Running zmq forwarder", f)
	f.Run()
}
