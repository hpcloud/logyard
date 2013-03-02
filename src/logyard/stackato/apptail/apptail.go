package apptail

import (
	"encoding/json"
	"fmt"
	"github.com/ActiveState/log"
	"github.com/ActiveState/tail"
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

// AppLogMessage is a struct corresponding to an entry in the app log stream.
type AppLogMessage struct {
	Text          string
	LogFilename   string
	UnixTime      int64
	HumanTime     string
	Source        string // example: app, staging, stackato.dea, stackato.stager
	InstanceIndex int
	AppID         int
	AppName       string
	NodeID        string // Host (DEA) IP of this app instance
}

// Publish publishes the receiver to logyard. Must be called once.
func (line *AppLogMessage) Publish(c *logyard.Client, allowInvalidJson bool) error {
	// JSON must be a UTF-8 encoded string.
	if !utf8.ValidString(line.Text) {
		line.Text = string([]rune(line.Text))
	}

	data, err := json.Marshal(line)
	if err != nil {
		if allowInvalidJson {
			log.Errorf("cannot encode %+v into JSON; %s. Skipping this message", line, err)
		} else {
			return fmt.Errorf("Failed to convert applogmsg to JSON: ", err)
		}
	}
	key := fmt.Sprintf("apptail.%d", line.AppID)
	err = c.Send(key, string(data))
	if err != nil {
		return fmt.Errorf("Failed to send applogmsg to logyard: ", err)
	}
	return nil
}

// AppInstanceStarted is a function to be invoked when dea/stager
// starts an application instance.
func AppInstanceStarted(c *logyard.Client, instance *AppInstance, nodeid string) {
	log.Infof("New app instance was started: %+v\n", instance)
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
				log.Errorf("Cannot tail file (%s); %s\n", filename, err)
				return
			}
			for line := range tail.Lines {
				// JSON must be a valid UTF-8 string
				if !utf8.ValidString(line.Text) {
					line.Text = string([]rune(line.Text))
				}
				err := (&AppLogMessage{
					Text:          line.Text,
					LogFilename:   filepath.Base(filename),
					UnixTime:      line.Time.Unix(),
					HumanTime:     line.Time.Format("2006-01-02T15:04:05-07:00"), // heroku-format
					Source:        instance.Type,
					InstanceIndex: instance.Index,
					AppID:         instance.AppID,
					AppName:       instance.AppName,
					NodeID:        nodeid,
				}).Publish(c, false)
				if err != nil {
					log.Fatal(err)
				}
			}
			err = tail.Wait()
			if err != nil {
				log.Error(err)
			}
		}(filename)
	}
}