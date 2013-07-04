package state

import (
	"fmt"
)

type Retrying struct {
	Error error // Retrying on this error
	*StateMachine
}

func (s Retrying) Transition(action int, rev int64) State {
	switch action {
	case START:
		// future retry will automatically be cancelled as we are starting.
		return s.start(rev)
	case STOP:
		return s.stop(rev)
	}
	panic("unreachable")
}

func (s Retrying) String() string {
	return "RETRYING"
}

func (s Retrying) Info() map[string]string {
	return map[string]string{
		"name":  "RETRYING",
		"error": fmt.Sprintf("%v", s.Error)}
}
