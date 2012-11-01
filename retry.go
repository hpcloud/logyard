package logyard

import (
	"log"
	"time"
)

// Retryer is a simple retryer
type Retryer struct {
	recentAttempts int
	lastAttempt    time.Time
}

func NewRetryer() *Retryer {
	return new(Retryer)
}

// Wait appropriately waits until next try
func (retry *Retryer) Wait(msg string) {
	retry.recentAttempts += 1
	if retry.recentAttempts > 3 {
		log.Printf("%s; retrying after %d seconds...",
			msg, retry.recentAttempts)
		time.Sleep(time.Duration(retry.recentAttempts) * time.Second)
	} else {
		log.Printf("%s; retrying...", msg)
	}

	// if the retry happens after a long time, reset our stats
	if time.Now().Sub(retry.lastAttempt).Seconds() > 60 {
		log.Println("Resetting retry attempts; ", time.Now().Sub(retry.lastAttempt).Seconds())
		retry.recentAttempts = 0
	}

	retry.lastAttempt = time.Now()
}
