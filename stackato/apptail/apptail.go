package main

import (
	"encoding/json"
	"fmt"
	"github.com/srid/tail"
	"log"
	"logyard"
	"path/filepath"
)

// AppInstance is the NATS message sent by dea/stager to notify of new
// instances.
type AppInstance struct {
	AppID    int
	AppName  string
	Type     string
	Index    int
	LogFiles []string
}

// The struct to be sent to logyard
type AppLogMessage struct {
	Text          string
	LogFilename   string
	UnixTime      int64
	HumanTime     string
	InstanceIndex int
	Source        string // possible values: app, staging, stackato.dea, stackato.stager
}

// AppInstanceStarted is invoked when dea/stager starts an application
// instance.
func AppInstanceStarted(c *logyard.Client, instance *AppInstance) {
	log.Printf("New instance was started: %v\n", instance)
	key := fmt.Sprintf("apptail.%d", instance.AppID)
	for _, filename := range instance.LogFiles {
		go func(filename string) {
			tail, err := tail.TailFile(filename, tail.Config{
				MaxLineSize: Config.MaxRecordSize,
				MustExist:   true,
				Follow:      true,
				Location:    -1,
				ReOpen:      false,
				Poll:        true})
			if err != nil {
				log.Printf("Error: cannot tail file (%s); %s\n", filename, err)
				return
			}
			for line := range tail.Lines {
				data, err := json.Marshal(AppLogMessage{
					Text:          line.Text,
					LogFilename:   filepath.Base(filename),
					UnixTime:      line.Time.Unix(),
					HumanTime:     line.Time.Format("2006-01-02T15:04:05-07:00"), // heroku-format
					InstanceIndex: instance.Index,
					Source:        instance.Type,
				})
				if err != nil {
					log.Fatal(err)
				}
				err = c.Send(key, string(data))
				if err != nil {
					log.Fatal("Failed to send: ", err)
				}
			}
			err = tail.Wait()
			if err != nil {
				log.Println(err)
			}
		}(filename)
	}
}
