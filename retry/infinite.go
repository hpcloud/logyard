package retry

import (
	"github.com/ActiveState/log"
	"time"
)

// InfiniteRetryer retries infinitely, but with progressive delays
// (5secs -> 5mins) that resets itself after 20 minutes of no-retry.
type InfiniteRetryer struct {
	firstRetry time.Time
	lastRetry  time.Time
}

func NewInfiniteRetryer() Retryer {
	return new(InfiniteRetryer)
}

func (retry *InfiniteRetryer) Wait(msg string, shouldWarn bool) bool {
	var delay time.Duration

	if retry.firstRetry.IsZero() {
		retry.reset()
		delay = 0
	} else if time.Now().Sub(retry.lastRetry) > time.Duration(20*time.Minute) {
		// reset retry stats after 20 minutes of non-failures
		retry.reset()
		delay = 0
	} else {
		retryingDuration := time.Now().Sub(retry.firstRetry)
		switch {
		case retryingDuration < time.Minute:
			// once every 5 seconds for 1 minute
			delay = 5 * time.Second
		case retryingDuration < (1+5)*time.Minute:
			// once every 30 seconds for next 5 minutes
			delay = 30 * time.Second
		case retryingDuration < (1+5+10)*time.Minute:
			// once every 1 minute for next 10 minutes
			delay = time.Minute
		default:
			// once every 5 minutes therein
			delay = 5 * time.Minute
		}
	}

	if shouldWarn {
		log.Warnf("%s; retrying after %v.", msg, delay)
	} else {
		log.Infof("%s; retrying after %v.", msg, delay)
	}
	time.Sleep(delay)
	retry.lastRetry = time.Now()
	return true
}

func (retry *InfiniteRetryer) reset() {
	now := time.Now()
	retry.firstRetry = now
	retry.lastRetry = now
}
