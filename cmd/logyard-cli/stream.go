package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ActiveState/log"
	"github.com/wsxiaoys/terminal/color"
	"logyard"
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

	srv, err := NewLineServer("tcp", addr)
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
		handleLine(line)
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

func handleLine(line []byte) {
	var record map[string]interface{}

	if err := json.Unmarshal(line, &record); err != nil {
		log.Fatal(err)
	}

	// REFACTOR: use Go interfaces to moves these to the respective
	// packages (apptail, cloudevents, systail).
	if _, ok := record["Name"]; ok {
		// systail
		name := record["Name"].(string)
		nodeid := record["NodeID"].(string)
		text := record["Text"].(string)

		color.Printf("@y%s@|@@@c%s@|: %s\n", name, nodeid, text)
	} else if _, ok = record["Type"]; ok {
		// cloud event
		kind := record["Type"].(string)
		nodeid := record["NodeID"].(string)
		severity := record["Severity"].(string)
		process := record["Process"].(string)
		desc := record["Desc"].(string)

		color.Printf("@g%s[%s]@|::@y%s@!@@@c%s@|: %s\n",
			kind, severity, process, nodeid, desc)
	} else if _, ok = record["Source"]; ok {
		// app logs
		appname := record["AppName"].(string)
		nodeid := record["NodeID"].(string)
		source := record["Source"].(string)
		text := record["Text"].(string)

		color.Printf("@b%s[%s]@|@@@c%s@|: %s\n",
			appname, source, nodeid, text)

	}
}
