package state

import (
	"fmt"
)

// Process is a start-able/stop-able entity not unlike an OS process
// or thread.
type Process interface {
	Start() error
	Stop() error
	Wait() error
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
		go monitor(m, rev+1)
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
			go doretry(m, rev+1, err)
			return RetryingState
		})
	}
}

func doretry(m *StateMachine, rev int64, err error) {
	// This could block.
	if m.retryer.Wait(
		fmt.Sprintf("[drain:%s] Drain exited abruptly -- %v", "todo", err)) {
		// TODO: move 'drain' specific message (above) out of the
		// state package.
		m.SetStateCustom(rev, func() State {
			return start(m, rev)
		})
	} else {
		m.SetState(rev, FatalState)
	}
}
