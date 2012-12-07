package main

import (
	"encoding/json"
	"fmt"
	"github.com/srid/log"
	"github.com/srid/tail"
	"logyard"
	"path/filepath"
	"unicode/utf8"
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
	Source        string // example: app, staging, stackato.dea, stackato.stager
	InstanceIndex int
	AppID         int
	AppName       string
}

// AppInstanceStarted is invoked when dea/stager starts an application
// instance.
func AppInstanceStarted(c *logyard.Client, instance *AppInstance) {
	log.Infof("New instance was started: %v\n", instance)
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
				log.Errorf("cannot tail file (%s); %s\n", filename, err)
				return
			}
			for line := range tail.Lines {
				// JSON must be a valid UTF-8 string
				if !utf8.ValidString(line.Text) {
					line.Text = string([]rune(line.Text))
				}
				data, err := json.Marshal(AppLogMessage{
					Text:          line.Text,
					LogFilename:   filepath.Base(filename),
					UnixTime:      line.Time.Unix(),
					HumanTime:     line.Time.Format("2006-01-02T15:04:05-07:00"), // heroku-format
					Source:        instance.Type,
					InstanceIndex: instance.Index,
					AppID:         instance.AppID,
					AppName:       instance.AppName,
				})
				if err != nil {
					log.Fatal("Failed to convert to JSON: ", err)
				}
				err = c.Send(key, string(data))
				if err != nil {
					log.Fatal("Failed to send to logyard: ", err)
				}
			}
			err = tail.Wait()
			if err != nil {
				log.Error(err)
			}
		}(filename)
	}
}
