package state

type Running struct {
	*StateMachine
}

func (s Running) Transition(action int, rev int64) State {
	switch action {
	case START:
		// ignore; already running
		return s
	case STOP:
		return s.stop(rev)
	}
	panic("unreachable")
}

func (s Running) String() string {
	return "RUNNING"
}

func (s Running) Info() map[string]string {
	return map[string]string{
		"name": "RUNNING"}
}
