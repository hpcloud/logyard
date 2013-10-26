package apptail

import (
	"encoding/json"
	"fmt"
	"github.com/ActiveState/log"
	"github.com/ActiveState/zmqpubsub"
	"logyard/clients/messagecommon"
	"unicode/utf8"
)

// Message corresponds to an entry in the app log stream.
type Message struct {
	LogFilename   string
	Source        string // example: app, staging, stackato.dea_ng
	InstanceIndex int
	AppGUID       string
	AppName       string
	AppSpace      string
	messagecommon.MessageCommon
}

// Publish publishes an AppLogMessage to logyard after sanity checks.
func (msg *Message) Publish(pub *zmqpubsub.Publisher, allowInvalidJson bool) error {
	// JSON must be a UTF-8 encoded string.
	if !utf8.ValidString(msg.Text) {
		msg.Text = string([]rune(msg.Text))
	}

	data, err := json.Marshal(msg)
	if err != nil {
		if allowInvalidJson {
			log.Errorf("Cannot encode %+v into JSON -- %s. Skipping this message", msg, err)
		} else {
			return fmt.Errorf("Failed to encode app log record to JSON: ", err)
		}
	}
	key := fmt.Sprintf("apptail.%v", msg.AppGUID)
	pub.MustPublish(key, string(data))
	return nil
}
