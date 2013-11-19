package common

import (
	"time"
)

const DEFAULT_SYSLOG_PRIORITY = "13"

// http://en.wikipedia.org/wiki/Syslog#Format_of_a_Syslog_Packet
type syslogMessage struct {
	Priority string `json:"priority"`
	Time     string `json:"time"`
}

func newSyslogMessage(pri string, t time.Time) syslogMessage {
	return syslogMessage{pri, t.Format(time.Stamp)}
}
