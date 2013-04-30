package test

import (
	"github.com/ActiveState/log"
)

type NoopRetryer struct{}

func (retry *NoopRetryer) Wait(msg string) bool {
	log.Errorf("%s -- never retrying.", msg)
	return false
}
