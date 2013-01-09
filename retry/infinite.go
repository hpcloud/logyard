package retry

import (
	"github.com/ActiveState/log"
	"time"
)

type InfiniteRetryer struct {
	started   time.Time
	lastDelay time.Duration
}

func NewInfiniteRetryer() Retryer {
	return &InfiniteRetryer{time.Now(), 0}
}

// Wait appropriately waits until next try. Wait delay is increased
// based on the length of failures, but Wait always returns true
// (hence InfiniteRetryer).
func (retry *InfiniteRetryer) Wait(msg string) bool {
	period := time.Now().Sub(retry.started)
	var delay time.Duration
	switch {
	case period < time.Minute:
		// once every 5 seconds for 1 minute
		delay = 5 * time.Second
	case period < (1+5)*time.Minute:
		// once every 30 seconds for next 5 minutes
		delay = 30 * time.Second
	case period < (1+5+10)*time.Minute:
		// once every 1 minute for next 10 minutes
		delay = time.Minute
	default:
		// once every 5 minutes therein
		delay = 5 * time.Minute
	}
	log.Warnf("%s; retrying after %v.", msg, delay)
	retry.lastDelay = delay
	time.Sleep(delay)
	return true
}
