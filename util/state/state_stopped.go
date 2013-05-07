package state

type Stopped struct {
	*StateMachine
}

func (s Stopped) Transition(action int, rev int64) State {
	switch action {
	case START:
		return s.start(rev)
	case STOP:
		// ignore; already stopped
		return s
	}
	panic("unreachable")

}

func (s Stopped) String() string {
	return "STOPPED"
}

func (s Stopped) Info() map[string]string {
	return map[string]string{
		"name": "STOPPED"}
}
