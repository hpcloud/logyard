package main

import (
	"flag"
	"fmt"
	"github.com/ActiveState/log"
	"logyard"
	"os"
	"os/signal"
	"stackato/server"
)

type stream struct {
}

func (cmd *stream) Name() string {
	return "stream"
}

func (cmd *stream) DefineFlags(fs *flag.FlagSet) {
}

func (cmd *stream) Run(args []string) error {
	ipaddr, err := server.LocalIP()
	if err != nil {
		return err
	}

	addr := ipaddr + ":7777"

	srv, err := NewLineServer("tcp", addr)
	if err != nil {
		return err
	}

	go srv.Start()

	name := "tmp.logyardctl.stream.7777"

	// REFACTOR: extract URI construction of add.go and then use
	// logyard.AddDrain directly.
	(&add{uri: "tcp://" + addr,
		filters: Filters(args)}).Run(
		[]string{name})

	deleteDrain := func() {
		if err := logyard.Config.DeleteDrain(name); err != nil {
			log.Fatal(err)
		}
		log.Infof("Deleted drain %s", name)
	}

	defer deleteDrain()

	// Handle Ctrl+C
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		for sig := range sigCh {
			log.Infof("captured %v", sig)
			deleteDrain()
			os.Exit(0)
		}
	}()

	// Print incoming records
	for line := range srv.Ch {
		fmt.Print(line)
	}

	return nil
}
