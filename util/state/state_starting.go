package state

type Starting struct {
	*StateMachine
}

func (s Starting) Transition(action int, rev int64) State {
	switch action {
	case START:
		// ignore; already running
		return s
	case STOP:
		return s.stop(rev)
	}
	panic("unreachable")
}

func (s Starting) String() string {
	return "STARTING"
}

func (s Starting) Info() map[string]string {
	return map[string]string{
		"name": "STARTING"}
}
