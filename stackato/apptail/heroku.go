package apptail

import (
	"time"
)

// ToHerokuTime formats the given time using heroku's datetime format
// (heroku logs).
func ToHerokuTime(t time.Time) string {
	return t.Format("2006-01-02T15:04:05-07:00")
}
