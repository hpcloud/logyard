package retry

import (
	"fmt"
	"github.com/ActiveState/log"
	"time"
)

// ProgressiveRetryer retries until a configurable limit (retryLimit),
// but with progressive delaysf (5secs -> 5mins) that resets itself
// after 20 minutes of no-retry.
type ProgressiveRetryer struct {
	firstRetry time.Time
	lastRetry  time.Time
	retryLimit time.Duration
}

// RESET_AFTER is the duration of no-retry period (`Wait` was not
// called) after which retry stats will be resetted. Calling `Wait`
// after reset happens is effectively the same as calling `Wait` the
// very first time.
var RESET_AFTER time.Duration

func init() {
	RESET_AFTER = time.Duration(20 * time.Minute)
}

func NewProgressiveRetryer(retryLimit time.Duration) Retryer {
	r := new(ProgressiveRetryer)
	r.retryLimit = retryLimit
	if r.hasRetryLimit() && r.retryLimit <= RESET_AFTER {
		log.Fatalf("retryLimit (%v) must be greater than RESET_AFTER (%v)",
			r.retryLimit, RESET_AFTER)
	}
	return r
}

func (retry *ProgressiveRetryer) Wait(msg string) bool {
	var delay time.Duration

	// how long is the retry happening?
	retryDuration := time.Now().Sub(retry.firstRetry)

	// how long since the last retry?
	silenceDuration := time.Now().Sub(retry.lastRetry)

	if retry.firstRetry.IsZero() {
		// first retry; just do it without waiting.
		retry.reset()
		delay = 0
	} else if silenceDuration > RESET_AFTER {
		// reset retry stats if Wait was not called in the last 20
		// minutes (implying sufficiently successful period).
		retry.reset()
		delay = 0
	} else if retry.hasRetryLimit() && retryDuration > retry.retryLimit {
		// respect retryLimit
		log.Errorf("%s -- giving up after retrying for %v.", msg, retry.retryLimit)
		retry.reset()
		return false
	} else {
		switch {
		case retryDuration < time.Minute:
			// once every 5 seconds for 1 minute
			delay = 5 * time.Second
		case retryDuration < (1+5)*time.Minute:
			// once every 30 seconds for next 5 minutes
			delay = 30 * time.Second
		case retryDuration < (1+5+10)*time.Minute:
			// once every 1 minute for next 10 minutes
			delay = time.Minute
		default:
			// once every 5 minutes therein
			delay = 5 * time.Minute
		}
	}

	// Log the retry action
	if delay == 0 {
		// Warn only for the first retry until next reset.
		log.Warnf(fmt.Sprintf("%s -- retrying now.", msg))
	} else {
		var msg string
		if retry.hasRetryLimit() {
			msg = fmt.Sprintf("%s -- retrying after %v (max %v).", msg, delay, retry.retryLimit)
		} else {
			msg = fmt.Sprintf("%s -- retrying after %v.", msg, delay)
		}
		log.Infof(msg)
	}

	time.Sleep(delay)
	retry.lastRetry = time.Now()
	return true
}

func (retry *ProgressiveRetryer) hasRetryLimit() bool {
	return retry.retryLimit.Seconds() > 0
}

func (retry *ProgressiveRetryer) reset() {
	now := time.Now()
	retry.firstRetry = now
	retry.lastRetry = now
}
