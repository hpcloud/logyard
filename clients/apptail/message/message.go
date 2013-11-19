package message

import (
	"encoding/json"
	"fmt"
	"github.com/ActiveState/log"
	"github.com/ActiveState/zmqpubsub"
	"logyard/clients/common"
	"unicode/utf8"
)

// Message corresponds to an entry in the app log stream.
type Message struct {
	LogFilename   string `json:"filename"`
	Source        string `json:"source"` // example: app, staging, stackato.dea_ng
	InstanceIndex int    `json:"instance_index"`
	AppGUID       string `json:"app_guid"`
	AppName       string `json:"app_name"`
	AppSpace      string `json:"app_space"`
	common.MessageCommon
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
