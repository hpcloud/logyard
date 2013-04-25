package state

// states
const (
	STOPPED = iota
	CRASHED
	FATAL // retry fail
	RETRYING
	RUNNING
)

// actions
const (
	START = iota
	STOP
)

func RetryingState(m *StateMachine, action int, rev int64) stateFn {
	switch action {
	case START:
		// future retry will automatically be cancelled as we are starting.
		start(m, rev)
		return RunningState
	case STOP:
		stop(m, rev)
		return StoppedState
	}
	panic("unreachable")
}

func StoppedState(m *StateMachine, action int, rev int64) stateFn {
	switch action {
	case START:
		start(m, rev)
		return RunningState
	case STOP:
		// ignore; already stopped
		return StoppedState
	}
	panic("unreachable")
}

func CrashedState(m *StateMachine, action int, rev int64) stateFn {
	return StoppedState(m, action, rev)
}

func RunningState(m *StateMachine, action int, rev int64) stateFn {
	switch action {
	case START:
		// ignore; already running
		return RunningState
	case STOP:
		stop(m, rev)
		return StoppedState
	}
	panic("unreachable")
}

func FatalState(m *StateMachine, action int, rev int64) stateFn {
	switch action {
	case START:
		start(m, rev)
		return RunningState
	case STOP:
		return StoppedState
	}
	panic("unreachable")
}
