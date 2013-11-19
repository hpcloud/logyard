package common

import (
	"time"
)

// MessageCommon contains common fields required for representing any message stream
type MessageCommon struct {
	Text      string        `json:"text"`       // Text content of the stream message
	NodeID    string        `json:"node_id"`    // IP address of the node from which this message originated
	UnixTime  int64         `json:"unix_time"`  // Unix timestamp
	HumanTime string        `json:"human_time"` // Human readable time (unspecified format)
	Syslog    syslogMessage `json:"syslog"`
}

func NewMessageCommon(text string, t time.Time, node string) MessageCommon {
	return MessageCommon{
		text,
		node,
		t.Unix(),
		toHerokuTime(t),
		newSyslogMessage(DEFAULT_SYSLOG_PRIORITY, t),
	}
}

// ToHerokuTime formats the given time using heroku's datetime format
// (heroku logs).
func toHerokuTime(t time.Time) string {
	return t.Format("2006-01-02T15:04:05-07:00")
}
