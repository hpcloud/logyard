package messagecommon

import (
	"time"
)

// MessageCommon contains common fields required for representing any message stream
type MessageCommon struct {
	Text      string // Text content of the stream message
	NodeID    string // IP address of the node from which this message originated
	UnixTime  int64  // Unix timestamp
	HumanTime string // Human readable time (unspecified format)
	Syslog    syslogMessage
}

func New(text string, t time.Time, node string) MessageCommon {
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
