package main

import (
	"log"
	"logyard"
	"logyard/drain"
)

func main() {
	f, err := logyard.NewForwarder()
	if err != nil {
		panic(err)
	}
	log.Println("Starting drain manager")
	drain.Run()
	log.Println("Running forwarder ", f)
	f.Run()
}
