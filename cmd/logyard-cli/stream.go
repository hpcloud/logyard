package main

import (
	"flag"
	"fmt"
	"github.com/ActiveState/log"
	"logyard"
	"logyard/cli"
	"math/rand"
	"os"
	"os/signal"
	"stackato/server"
	"time"
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

	rand.Seed(time.Now().UnixNano())
	port := 7000 + rand.Intn(1000)
	addr := fmt.Sprintf("%s:%d", ipaddr, port)

	srv, err := cli.NewLineServer("tcp", addr)
	if err != nil {
		return err
	}

	go srv.Start()

	name := fmt.Sprintf("tmp.logyardctl.%s-%d", ipaddr, port)

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

	handleKeyboardInterrupt(func() {
		deleteDrain()
		os.Exit(1)
	})

	// Print incoming records
	for line := range srv.Ch {
		cli.PrintMessage(line)
	}

	return nil
}

func handleKeyboardInterrupt(cleanupFn func()) {
	// Handle Ctrl+C
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		for _ = range sigCh {
			cleanupFn()
		}
	}()
}
