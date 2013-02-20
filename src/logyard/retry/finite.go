package retry

import (
	"github.com/ActiveState/log"
	"time"
)

type FiniteRetryer struct {
	tracker        *Tracker
	recentAttempts int
	lastAttempt    time.Time
}

func NewFiniteRetryer() Retryer {
	return new(FiniteRetryer)
}

const MAX_WAIT_SECONDS = 60 * 5 // 5 minutes

func (retry *FiniteRetryer) Wait(msg string, shouldWarn bool) bool {
	if !shouldWarn {
		panic("FiniteRetryer does not support shouldWarn=false")
	}
	if retry.tracker == nil {
		retry.tracker = NewTracker(10) // keep track of the last 10 error events
	}
	retry.tracker.Event()

	// give up retrying if there were >= 10 errors in the last minute
	if retry.tracker.In(time.Minute) {
		log.Errorf("%s; giving up retrying (10 errors in last minute)", msg)
		return false
	}
	retry.recentAttempts += 1
	if retry.recentAttempts > 3 {
		waitSeconds := retry.recentAttempts
		if waitSeconds > MAX_WAIT_SECONDS {
			waitSeconds = MAX_WAIT_SECONDS
		}
		log.Warnf("%s; retrying after %d seconds...", msg, waitSeconds)
		time.Sleep(time.Duration(waitSeconds) * time.Second)
	} else {
		log.Warnf("%s; retrying...", msg)
	}

	// reset our stats if there weren't any retry attempts in the last
	// minute.
	if time.Now().Sub(retry.lastAttempt).Seconds() > 60 {
		log.Info("Resetting retry attempts; ", time.Now().Sub(retry.lastAttempt).Seconds())
		retry.recentAttempts = 0
	}

	retry.lastAttempt = time.Now()

	return true
}
