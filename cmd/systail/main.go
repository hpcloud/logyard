package main

import (
	"encoding/json"
	"fmt"
	"github.com/ActiveState/log"
	"github.com/ActiveState/tail"
	"github.com/alecthomas/gozmq"
	"logyard"
	"os"
	"stackato/server"
	"unicode/utf8"
)

func tailLogFile(name string, filepath string, nodeid string) (*tail.Tail, error) {
	if filepath == "" {
		filepath = fmt.Sprintf("/s/logs/%s.log", name)
	}

	log.Info("Tailing... ", filepath)

	t, err := tail.TailFile(filepath, tail.Config{
		MaxLineSize: Config.MaxRecordSize,
		MustExist:   false,
		Follow:      true,
		// ignore existing content, to support subsequent re-runs of systail
		Location: 0,
		ReOpen:   true,
		Poll:     false})

	if err != nil {
		return nil, err
	}

	go func(name string, tail *tail.Tail) {
		pub, err := logyard.Logyard.NewPublisher()
		if err != nil {
			log.Fatal(err)
		}
		defer pub.Stop()

		for line := range tail.Lines {
			// JSON must be a valid UTF-8 string
			if !utf8.ValidString(line.Text) {
				line.Text = string([]rune(line.Text))
			}
			data, err := json.Marshal(map[string]interface{}{
				"UnixTime": line.Time.Unix(),
				"Text":     line.Text,
				"Name":     name,
				"NodeID":   nodeid})
			if err != nil {
				tail.Killf("Failed to convert to JSON: %v", err)
				break
			}
			err = pub.Publish("systail."+name+"."+nodeid, string(data))
			if err != nil {
				log.Fatal("Failed to send to logyard: ", err)
			}
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

	fmt.Printf("%+v\n", Config.LogFiles)
	if len(Config.LogFiles) == 0 {
		log.Fatal("No log files configured in doozer")
	}

	for name, logfile := range Config.LogFiles {
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
