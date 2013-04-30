package state

import (
	"fmt"
	"github.com/ActiveState/log"
	"logyard/util/retry"
	"sync"
)

type StateMachine struct {
	running bool
	title   string
	process Process
	retryer retry.Retryer
	state   State
	rev     int64
	mux     sync.Mutex
}

func NewStateMachine(title string, process Process, retryer retry.Retryer) *StateMachine {
	m := &StateMachine{}
	m.title = title
	m.process = process
	m.retryer = retryer
	m.state = Stopped{m}
	m.rev = 1
	m.running = true
	return m
}

func (m *StateMachine) Log(msg string, v ...interface{}) {
	log.Infof(m.process.Logf(msg, v...))
}

func (m *StateMachine) SendAction(action int) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	if !m.running {
		return fmt.Errorf("StateMachine is stopped")
	}
	oldState := m.state
	m.Log("About to state change: %s -[%v]-> ?? (%d)\n",
		oldState, action, m.rev)
	m.state = m.state.Transition(action, m.rev)
	m.rev += 1
	m.Log("State change: %s => %s (%d)\n",
		oldState, m.state, m.rev)
	return nil
}

func (m *StateMachine) Stop() {
	m.Log("Stopping STM...")
	m.mux.Lock()
	defer m.mux.Unlock()

	m.running = false

	m.Log("Stopped STM.")

	// reset fields to prevent (buggy) future use
	m.process = nil
	m.state = nil
	m.rev = -10
}

func (m *StateMachine) SetStateCustom(rev int64, fn func() State) int64 {
	m.mux.Lock()
	defer m.mux.Unlock()
	if m.running && rev == m.rev {
		oldState := m.state
		m.state = fn()
		if m.state == nil {
			panic("nil state")
		}
		m.rev += 1
		m.Log("Custom state change: %s => %s (%d)\n",
			oldState, m.state, m.rev)
		return m.rev
	}
	m.Log("Skipping state change; rev changed (expected %d, have %d) or stopped (%v)\n",
		rev, m.rev, !m.running)
	return -1
}

func (m *StateMachine) SetState(rev int64, state State) int64 {
	return m.SetStateCustom(rev, func() State {
		return state
	})
}

func (s *StateMachine) stop(rev int64) State {
	err := s.process.Stop()
	if err != nil {
		// Error reporting?
		return Fatal{err, s}
	}
	return Stopped{s}
}

func (s *StateMachine) start(rev int64) State {
	// start it
	s.Log("STM starting process")
	err := s.process.Start()
	if err != nil {
		return Fatal{err, s}
	} else {
		rev = rev + 1 // account for settig of RunningState
		go s.monitor(rev)
		return Running{s}
	}
	panic("unreachable")
}

func (s *StateMachine) monitor(rev int64) {
	err := s.process.Wait()
	s.Log("%s exited with %v", s.title, err)
	if err == nil {
		// If a process exited cleanly (no errors), then just mark it
		// as STOPPED without retrying.
		s.SetState(rev, Stopped{s}) // rev confict here is normal.
	} else {
		s.SetStateCustom(rev, func() State {
			rev = rev + 1 // account for setting of RetryingState
			go s.doretry(rev, err)
			return Retrying{err, s}
		})
	}
}

func (s *StateMachine) doretry(rev int64, err error) {
	// This could block.
	if s.retryer.Wait(
		fmt.Sprintf(s.process.Logf(
			"%s exited abruptly -- %v", s.title, err))) {
		s.Log("Retrying now.")
		s.SetStateCustom(rev, func() State {
			return s.start(rev)
		})
	} else {
		err := fmt.Errorf("retried too long")
		s.Log("%v; marking as FATAL", err)
		s.SetState(rev, Fatal{err, s})
	}
}
