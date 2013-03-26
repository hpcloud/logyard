package main

import (
	"flag"
	"fmt"
	"github.com/ActiveState/log"
	"logyard"
	"logyard/cli"
	cli_stream "logyard/cli/stream"
	"logyard/drain"
	"math/rand"
	"os"
	"os/signal"
	"stackato/server"
	"time"
)

type stream struct {
	raw  bool
	time bool
}

func (cmd *stream) Name() string {
	return "stream"
}

func (cmd *stream) DefineFlags(fs *flag.FlagSet) {
	fs.BoolVar(&cmd.raw, "raw", false,
		"Show unformatted logs, including logyard INFO records (skipped by default)")
	fs.BoolVar(&cmd.time, "time", false,
		"Show timestamp")
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

	Init("stream")
	uri, err := drain.ConstructDrainURI(name, "tcp://"+addr, args, map[string]string{
		"format": "raw",
	})
	if err != nil {
		return err
	}
	if err = logyard.Config.AddDrain(name, uri); err != nil {
		return err
	}
	log.Infof("Added drain %s", uri)

	deleteDrain := func() {
		if err := logyard.Config.DeleteDrain(name); err != nil {
			log.Fatal(err)
		}
		fmt.Println("")
		log.Infof("Deleted drain %s", name)
	}
	defer deleteDrain()

	handleKeyboardInterrupt(func() {
		deleteDrain()
		os.Exit(1)
	})

	cli_stream.Stream(srv.Ch, cli_stream.MessagePrinterOptions{
		cmd.raw, cmd.time})

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
