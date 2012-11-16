package logyard

import (
	"logyard/log2"
	"time"
)

type Retryer struct {
	tracker        *Tracker
	recentAttempts int
	lastAttempt    time.Time
}

func NewRetryer() *Retryer {
	return new(Retryer)
}

const MAX_WAIT_SECONDS = 60 * 5 // 5 minutes

// Wait appropriately waits until next try (exponential backoff delay)
func (retry *Retryer) Wait(msg string) bool {
	if retry.tracker == nil {
		retry.tracker = NewTracker(10) // keep track of the last 10 error events
	}
	retry.tracker.Event()

	// if there were 10 errors in the last minute, give up trying
	if retry.tracker.In(time.Minute) {
		log2.Infof("%s; giving up retrying (10 errors in last minute)", msg)
		return false
	}
	retry.recentAttempts += 1
	if retry.recentAttempts > 3 {
		waitSeconds := retry.recentAttempts
		if waitSeconds > MAX_WAIT_SECONDS {
			waitSeconds = MAX_WAIT_SECONDS
		}
		log2.Infof("%s; retrying after %d seconds...", msg, waitSeconds)
		time.Sleep(time.Duration(waitSeconds) * time.Second)
	} else {
		log2.Infof("%s; retrying...", msg)
	}

	// reset our stats if there weren't any retry attempts in the last
	// minute.
	if time.Now().Sub(retry.lastAttempt).Seconds() > 60 {
		log2.Info("Resetting retry attempts; ", time.Now().Sub(retry.lastAttempt).Seconds())
		retry.recentAttempts = 0
	}

	retry.lastAttempt = time.Now()

	return true
}
