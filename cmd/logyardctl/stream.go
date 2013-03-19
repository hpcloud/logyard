package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ActiveState/log"
	"logyard/config"
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
		if err := config.Config.DeleteDrain(name); err != nil {
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
		handleLine(line)

	}

	return nil
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

		fmt.Println(FgYellow + name + Reset + "@" + FgCyan + nodeid + Reset +
			" : " + string(text))
	} else if _, ok = record["Type"]; ok {
		// cloud event
		kind := record["Type"].(string)
		nodeid := record["NodeID"].(string)
		severity := record["Severity"].(string)
		process := record["Process"].(string)
		desc := record["Desc"].(string)

		fmt.Println(FgGreen + kind + "[" + severity + "]" + Reset + " :: " +
			FgYellow + process + Reset + "@" + FgCyan + nodeid + Reset +
			" : " + string(desc))
	} else if _, ok = record["Source"]; ok {
		// app logs
		appname := record["AppName"].(string)
		nodeid := record["NodeID"].(string)
		source := record["Source"].(string)
		text := record["Text"].(string)
		fmt.Println(
			FgBlue + appname + "[" + source + "]" + Reset + "@" + FgCyan + nodeid + Reset +
				" : " + string(text))

	}
}
