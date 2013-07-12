package apptail

import (
	"encoding/json"
	"fmt"
	"github.com/ActiveState/log"
	"github.com/ActiveState/tail"
	"logyard"
	"logyard/util/pubsub"
	"unicode/utf8"
)

// AppInstance is the NATS message sent by dea/stager to notify of new
// instances.
type AppInstance struct {
	AppID    int
	AppName  string
	AppGroup string `json:"group"`
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
	AppID         int
	AppName       string
	AppGroup      string
	NodeID        string // Host (DEA,stager) IP of this app instance
}

// Publish publishes an AppLogMessage to logyard after sanity checks.
func (line *AppLogMessage) Publish(pub *pubsub.Publisher, allowInvalidJson bool) error {
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
	key := fmt.Sprintf("apptail.%d", line.AppID)
	pub.MustPublish(key, string(data))
	return nil
}

// AppInstanceStarted is a function to be invoked when dea/stager
// starts an application instance.
func AppInstanceStarted(instance *AppInstance, nodeid string) {
	log.Infof("New app instance was started: %+v", instance)

	for name, filename := range instance.LogFiles {
		go func(name string, filename string) {
			pub := logyard.Broker.NewPublisherMust()
			defer pub.Stop()

			tail, err := tail.TailFile(filename, tail.Config{
				MaxLineSize: GetConfig().MaxRecordSize,
				MustExist:   true,
				Follow:      true,
				Location:    -1,
				ReOpen:      false,
				Poll:        true,
				LimitRate:   GetConfig().RateLimit})
			if err != nil {
				log.Errorf("Cannot tail file (%s); %s", filename, err)
				return
			}
			for line := range tail.Lines {
				// JSON must be a valid UTF-8 string
				if !utf8.ValidString(line.Text) {
					line.Text = string([]rune(line.Text))
				}
				msg := &AppLogMessage{
					Text:          line.Text,
					LogFilename:   name,
					UnixTime:      line.Time.Unix(),
					HumanTime:     ToHerokuTime(line.Time),
					Source:        instance.Type,
					InstanceIndex: instance.Index,
					AppID:         instance.AppID,
					AppName:       instance.AppName,
					AppGroup:      instance.AppGroup,
					NodeID:        nodeid,
				}

				if line.Err != nil {
					// Mark this as a special error record, as it is
					// coming from tail, not the app.
					msg.Source = "stackato.apptail"
				}

				err := msg.Publish(pub, false)
				if err != nil {
					log.Fatal(err)
				}
			}
			err = tail.Wait()
			if err != nil {
				log.Error(err)
			}
		}(name, filename)
	}
}
