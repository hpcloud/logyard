package state

import (
	"fmt"
)

type Fatal struct {
	Error error
	*StateMachine
}

func (s Fatal) Transition(action int, rev int64) State {
	switch action {
	case START:
		return s.start(rev)
	case STOP:
		return Stopped{s.StateMachine}
	}
	panic("unreachable")
}

func (s Fatal) String() string {
	return "FATAL"
}

func (s Fatal) Info() map[string]string {
	return map[string]string{
		"name":  "FATAL",
		"error": fmt.Sprintf("%v", s.Error)}
}
