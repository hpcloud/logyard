package state

// States
const (
	STOPPED = iota
	CRASHED
	// TODO: FATAL state requires a reason attribute. implement
	// attributes.
	FATAL // retry fail, or other reasons.
	RETRYING
	RUNNING
)

// Actions
const (
	START = iota << 2
	STOP
)

func RetryingState(m *StateMachine, action int, rev int64) State {
	switch action {
	case START:
		// future retry will automatically be cancelled as we are starting.
		return start(m, rev)
	case STOP:
		return stop(m, rev)
	}
	panic("unreachable")
}

func StoppedState(m *StateMachine, action int, rev int64) State {
	switch action {
	case START:
		return start(m, rev)
	case STOP:
		// ignore; already stopped
		return StoppedState
	}
	panic("unreachable")
}

func CrashedState(m *StateMachine, action int, rev int64) State {
	return StoppedState(m, action, rev)
}

func RunningState(m *StateMachine, action int, rev int64) State {
	switch action {
	case START:
		// ignore; already running
		return RunningState
	case STOP:
		return stop(m, rev)
	}
	panic("unreachable")
}

func FatalState(m *StateMachine, action int, rev int64) State {
	switch action {
	case START:
		return start(m, rev)
	case STOP:
		return StoppedState
	}
	panic("unreachable")
}
