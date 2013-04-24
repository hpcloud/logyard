package main

import (
	"github.com/ActiveState/log"
	"github.com/alecthomas/gozmq"
	"logyard"
	"logyard/drain"
	"os"
	"os/signal"
	"stackato/server"
	"syscall"
)

func main() {
	major, minor, patch := gozmq.Version()
	log.Infof("Starting logyard (zeromq %d.%d.%d)", major, minor, patch)

	doozer, headRev, err := server.NewDoozerClient("logyard")
	if err != nil {
		log.Fatal(err)
	}

	// logyard.Init(doozer, headRev, true)
	server.Init(doozer, headRev)
	logyard.Init2(false) // config changes will be monitored by drain manager

	m := drain.NewDrainManager()
	log.Info("Starting drain manager")
	go m.Run()
	// SIGTERM handle for stopping running drains.
	go func() {
		sigchan := make(chan os.Signal)
		signal.Notify(sigchan, syscall.SIGTERM)
		<-sigchan
		log.Info("Stopping all drains before exiting")
		m.Stop()
		log.Info("Exiting now.")
		os.Exit(0)
	}()

	log.Info("Running pubsub broker")
	log.Fatal(logyard.Broker.Run())
}
