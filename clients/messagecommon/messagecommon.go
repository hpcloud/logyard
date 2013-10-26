package messagecommon

import (
	"time"
)

// MessageCommon contains common fields required for representing any message stream
type MessageCommon struct {
	Text      string // Text content of the stream message
	UnixTime  int64  // Unix timestamp
	HumanTime string // Human readable time (unspecified format)
	NodeID    string // IP address of the node from which this message originated
}

func New(text string, t time.Time, node string) MessageCommon {
	return MessageCommon{
		text,
		t.Unix(),
		toHerokuTime(t),
		node}
}

// ToHerokuTime formats the given time using heroku's datetime format
// (heroku logs).
func toHerokuTime(t time.Time) string {
	return t.Format("2006-01-02T15:04:05-07:00")
}
