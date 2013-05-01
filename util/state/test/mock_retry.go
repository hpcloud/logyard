package test

import (
	"github.com/ActiveState/log"
)

type NoopRetryer struct{}

func (retry *NoopRetryer) Wait(msg string) bool {
	log.Errorf("%s -- never retrying.", msg)
	return false
}

// ThriceRetryer retries only three times.
type ThriceRetryer struct {
	count int
}

func (retry *ThriceRetryer) Wait(msg string) bool {
	if retry.count < 3 {
		retry.count += 1
		log.Infof("retry #%d -- %v.", retry.count, msg)
		return true
	}
	return false
}
