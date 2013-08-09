package main

import (
	"encoding/json"
	"fmt"
	"github.com/ActiveState/log"
	"github.com/ActiveState/tail"
	"github.com/alecthomas/gozmq"
	"logyard"
	"logyard/stackato/apptail"
	"os"
	"stackato/server"
	"unicode/utf8"
)

// SystailLogMessage is a struct corresponding to an entry in the
// systail log stream. Generally a subset of the fields in
// `AppLogMessage`.
type SystailLogMessage struct {
	Text      string
	Name      string // Component name (eg: dea)
	UnixTime  int64
	HumanTime string
	NodeID    string
}

func tailLogFile(
	name string, filepath string, nodeid string) (*tail.Tail, error) {
	if filepath == "" {
		filepath = fmt.Sprintf("/s/logs/%s.log", name)
	}

	log.Info("Tailing... ", filepath)

	t, err := tail.TailFile(filepath, tail.Config{
		MaxLineSize: getConfig().MaxRecordSize,
		MustExist:   false,
		Follow:      true,
		// ignore existing content, to support subsequent re-runs of systail
		Location: &tail.SeekInfo{0, os.SEEK_END},
		ReOpen:   true,
		Poll:     false})

	if err != nil {
		return nil, err
	}

	go func(name string, tail *tail.Tail) {
		pub := logyard.Broker.NewPublisherMust()
		defer pub.Stop()

		for line := range tail.Lines {
			// JSON must be a valid UTF-8 string
			if !utf8.ValidString(line.Text) {
				line.Text = string([]rune(line.Text))
			}
			data, err := json.Marshal(SystailLogMessage{
				Text:      line.Text,
				Name:      name,
				UnixTime:  line.Time.Unix(),
				HumanTime: apptail.ToHerokuTime(line.Time),
				NodeID:    nodeid,
			})
			if err != nil {
				tail.Killf("Failed to encode to JSON: %v", err)
				break
			}
			pub.MustPublish("systail."+name+"."+nodeid, string(data))
		}
	}(name, t)

	return t, nil
}

func main() {
	major, minor, patch := gozmq.Version()
	log.Infof("Starting systail (zeromq %d.%d.%d)", major, minor, patch)

	LoadConfig()

	nodeid, err := server.LocalIP()
	if err != nil {
		log.Fatalf("Failed to determine IP addr: %v", err)
	}
	log.Info("Host IP: ", nodeid)

	tailers := []*tail.Tail{}

	logFiles := getConfig().LogFiles

	fmt.Printf("%+v\n", logFiles)
	if len(logFiles) == 0 {
		log.Fatal("No log files exist in configuration.")
	}

	for name, logfile := range logFiles {
		t, err := tailLogFile(name, logfile, nodeid)
		if err != nil {
			log.Fatal(err)
		}
		tailers = append(tailers, t)
	}

	for _, tail := range tailers {
		err := tail.Wait()
		if err != nil {
			log.Errorf("Cannot tail [%s]: %s", tail.Filename, err)
		}
	}

	// we don't expect any of the tailers to exit with or without
	// error.
	log.Error("No file left to tail; exiting.")
	os.Exit(1)
}
