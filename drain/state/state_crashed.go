package state

type Crashed struct {
	*StateMachine
}

func (s Crashed) Transition(action int, rev int64) State {
	return Stopped{s.StateMachine}.Transition(action, rev)
}

func (s Crashed) String() string {
	return "CRASHED"
}
