package state

type Retrying struct {
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
