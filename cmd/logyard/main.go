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
	log.Println("Running forwarder ", f)
	f.Run()
}
