package commands

import (
	"flag"
	"fmt"
	"github.com/hpcloud/log"
	"github.com/hpcloud/stackato-go/server"
	"logyard"
	cli_stream "logyard/cli/stream"
	"logyard/drain"
	"logyard/util/lineserver"
	"math/rand"
	"os"
	"os/signal"
	"time"
)

const CLI_STREAM_PROTO = "tcp"

type stream struct {
	json    bool
	raw     bool
	time    bool
	nocolor bool
	nodeid  string
}

func (cmd *stream) Name() string {
	return "stream"
}

func (cmd *stream) DefineFlags(fs *flag.FlagSet) {
	fs.BoolVar(&cmd.json, "json", false, "(Unsupported)")
	fs.BoolVar(&cmd.raw, "raw", false,
		"Show unformatted logs, including logyard INFO records (skipped by default)")
	fs.BoolVar(&cmd.time, "time", false,
		"Show timestamp")
	fs.BoolVar(&cmd.nocolor, "nocolor", false,
		"Output with no colors")
	fs.StringVar(&cmd.nodeid, "nodeid", "",
		"Filter by this node IP address")
}

func (cmd *stream) Run(args []string) (string, error) {
	ipaddr, err := server.LocalIP()
	if err != nil {
		return "", err
	}

	rand.Seed(time.Now().UnixNano())
	port := 7000 + rand.Intn(1000)
	addr := fmt.Sprintf("%s:%d", ipaddr, port)

	srv, err := lineserver.NewLineServer(CLI_STREAM_PROTO, addr)
	if err != nil {
		return "", err
	}
	go srv.Start()

	// Debug mode allows one to debug just the logyard related logs,
	// without any magical stripping.
	debugMode := false
	if len(args) == 1 && args[0] == "debug" {
		debugMode = true
		args = []string{
			"systail.logyard",
			"systail.apptail",
			"systail.logyard_sieve",
			"systail.systail",
		}
	}

	name := fmt.Sprintf("tmp.logyard-cli.%s-%d", ipaddr, port)

	uri, err := drain.ConstructDrainURI(
		name,
		fmt.Sprintf("%s://%s", CLI_STREAM_PROTO, addr),
		args,
		map[string]string{"format": "raw"})
	if err != nil {
		return "", err
	}
	if err = logyard.AddDrain(name, uri); err != nil {
		return "", err
	}
	log.Infof("Added drain %s", uri)

	deleteDrain := func() {
		if err := logyard.DeleteDrain(name); err != nil {
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
		cmd.raw, cmd.raw || debugMode, cmd.time, cmd.nocolor, cmd.nodeid, cmd.json})

	return "", nil
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
