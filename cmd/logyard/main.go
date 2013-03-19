package main

import (
	"github.com/ActiveState/log"
	"github.com/alecthomas/gozmq"
	"logyard"
	"logyard/drain"
	"stackato/server"
)

func main() {
	major, minor, patch := gozmq.Version()
	log.Infof("Starting logyard (zeromq %d.%d.%d)", major, minor, patch)

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
	m := drain.NewDrainManager()
	log.Info("Starting drain manager")
	go m.Run()
	log.Info("Running zmq forwarder", f)
	f.Run()
}
