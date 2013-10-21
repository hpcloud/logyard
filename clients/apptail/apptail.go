package apptail

import (
	"encoding/json"
	"fmt"
	"github.com/ActiveState/log"
	"github.com/ActiveState/tail"
	"github.com/ActiveState/zmqpubsub"
	"logyard"
	"os"
	"time"
	"unicode/utf8"
)

// AppInstance is the NATS message sent by dea/stager to notify of new
// instances.
type AppInstance struct {
	AppGUID  string
	AppName  string
	AppSpace string `json:"space"`
	Type     string
	Index    int
	LogFiles map[string]string
}

// AppLogMessage is a struct corresponding to an entry in the app log stream.
type AppLogMessage struct {
	Text          string
	LogFilename   string
	UnixTime      int64
	HumanTime     string
	Source        string // example: app, staging, stackato.dea, stackato.stager
	InstanceIndex int
	AppGUID       string
	AppName       string
	AppSpace      string
	NodeID        string // Host (DEA,stager) IP of this app instance
}

// Publish publishes an AppLogMessage to logyard after sanity checks.
func (line *AppLogMessage) Publish(pub *zmqpubsub.Publisher, allowInvalidJson bool) error {
	// JSON must be a UTF-8 encoded string.
	if !utf8.ValidString(line.Text) {
		line.Text = string([]rune(line.Text))
	}

	data, err := json.Marshal(line)
	if err != nil {
		if allowInvalidJson {
			log.Errorf("Cannot encode %+v into JSON -- %s. Skipping this message", line, err)
		} else {
			return fmt.Errorf("Failed to encode app log record to JSON: ", err)
		}
	}
	key := fmt.Sprintf("apptail.%v", line.AppGUID)
	pub.MustPublish(key, string(data))
	return nil
}

// AppInstanceStarted is a function to be invoked when dea/stager
// starts an application instance.
func AppInstanceStarted(instance *AppInstance, nodeid string) {
	log.Infof("Tailing %v logs for %v[%v] -- %+v",
		instance.Type, instance.AppName, instance.Index, instance)

	// convert MB to limit in bytes.
	filesize_limit := GetConfig().FileSizeLimit * 1024 * 1024

	if !(filesize_limit > 0) {
		panic("invalid value for `read_limit' in apptail config")
	}

	for name, filename := range instance.LogFiles {
		go func(name string, filename string) {
			pub := logyard.Broker.NewPublisherMust()
			defer pub.Stop()

			fi, err := os.Stat(filename)
			if err != nil {
				log.Errorf("Cannot stat file (%s); %s", filename, err)
				return
			}
			size := fi.Size()
			limit := filesize_limit
			if size > filesize_limit {
				err := fmt.Errorf("Skipping much of a large log file (%s); size (%v bytes) > read_limit (%v bytes)",
					name, size, filesize_limit)
				// Publish special error message.
				PublishLine(instance, nodeid, name, pub, &tail.Line{
					Text: err.Error(),
					Time: time.Now(),
					Err:  err})
			} else {
				limit = size
			}

			tail, err := tail.TailFile(filename, tail.Config{
				MaxLineSize: GetConfig().MaxRecordSize,
				MustExist:   true,
				Follow:      true,
				Location:    &tail.SeekInfo{-limit, os.SEEK_END},
				ReOpen:      false,
				Poll:        false,
				LimitRate:   GetConfig().RateLimit})
			if err != nil {
				log.Errorf("Cannot tail file (%s); %s", filename, err)
				return
			}

			for line := range tail.Lines {
				PublishLine(instance, nodeid, name, pub, line)
			}

			err = tail.Wait()
			if err != nil {
				log.Error(err)
			}

			log.Infof("Completed tailing %v for %v[%v]", name, instance.AppName, instance.Index)
		}(name, filename)
	}
}

func PublishLine(
	instance *AppInstance, nodeid string,
	name string, pub *zmqpubsub.Publisher,
	line *tail.Line) {

	msg := &AppLogMessage{
		Text:          line.Text,
		LogFilename:   name,
		UnixTime:      line.Time.Unix(),
		HumanTime:     ToHerokuTime(line.Time),
		Source:        instance.Type,
		InstanceIndex: instance.Index,
		AppGUID:       instance.AppGUID,
		AppName:       instance.AppName,
		AppSpace:      instance.AppSpace,
		NodeID:        nodeid,
	}

	if line.Err != nil {
		// Mark this as a special error record, as it is
		// coming from tail, not the app.
		msg.Source = "stackato.apptail"
		msg.LogFilename = ""
		log.Warnf("[%s] %s", instance.AppName, line.Text)
	}

	err := msg.Publish(pub, false)
	if err != nil {
		log.Fatal(err)
	}

}
