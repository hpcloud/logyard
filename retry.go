package logyard

import (
	"log"
	"time"
)

type Retryer struct {
	recentAttempts int
	lastAttempt    time.Time
}

func NewRetryer() *Retryer {
	return new(Retryer)
}

const MAX_WAIT_SECONDS = 60 * 5 // 5 minutes

// Wait appropriately waits until next try (exponential backoff delay)
func (retry *Retryer) Wait(msg string) {
	retry.recentAttempts += 1
	if retry.recentAttempts > 3 {
		waitSeconds := retry.recentAttempts
		if waitSeconds > MAX_WAIT_SECONDS {
			waitSeconds = MAX_WAIT_SECONDS
		}
		log.Printf("%s; retrying after %d seconds...",
			msg, waitSeconds)
		time.Sleep(time.Duration(waitSeconds) * time.Second)
	} else {
		log.Printf("%s; retrying...", msg)
	}

	// reset our stats if there weren't any retry attempts in the last
	// minute.
	if time.Now().Sub(retry.lastAttempt).Seconds() > 60 {
		log.Println("Resetting retry attempts; ", time.Now().Sub(retry.lastAttempt).Seconds())
		retry.recentAttempts = 0
	}

	retry.lastAttempt = time.Now()
}
