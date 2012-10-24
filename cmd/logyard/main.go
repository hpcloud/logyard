package main

import (
	"log"
	"logyard"
)

func main() {
	f, err := logyard.NewForwarder()
	if err != nil {
		panic(err)
	}
	m := logyard.NewDrainManager()
	log.Println("Starting drain manager")
	m.Run()
	log.Println("Running zmq forwarder", f)
	f.Run()
}
