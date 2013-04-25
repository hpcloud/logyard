package state

import (
	"fmt"
)

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

func stop(m *StateMachine, rev int64) State {
	err := m.process.Stop()
	if err != nil {
		return FatalState
	}
	return StoppedState
}

func start(m *StateMachine, rev int64) State {
	// start it
	err := m.process.Start()
	if err != nil {
		return FatalState
	} else {
		rev = rev + 1 // account for settig of RunningState
		go monitor(m, rev)
		return RunningState
	}
	panic("unreachable")
}

func monitor(m *StateMachine, rev int64) {
	err := m.process.Wait()
	if err == nil {
		m.SetState(rev, StoppedState)
	} else {
		m.SetStateCustom(rev, func() State {
			rev = rev + 1 // account for setting of RetryingState
			go doretry(m, rev, err)
			return RetryingState
		})
	}
}

func doretry(m *StateMachine, rev int64, err error) {
	// This could block.
	if m.retryer.Wait(
		fmt.Sprintf("[drain:???] Drain exited abruptly -- %v", err)) {
		// TODO: move 'drain' specific message (above) out of the
		// state package.
		m.SetStateCustom(rev, func() State {
			fmt.Println("starting ???")
			return start(m, rev)
		})
	} else {
		m.SetState(rev, FatalState)
	}
}
